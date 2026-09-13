package pipeline

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/invariant"
	"github.com/instagrim-dev/newf/internal/provider"
	"github.com/instagrim-dev/newf/internal/store"
	"github.com/instagrim-dev/newf/internal/success"
)

// --- inputs ---

// SuccessCompressInput requests a compression pass over the problem's
// evaluated frontier outcomes. MinSupport <= 0 defaults to 1 (early corpora
// are sparse; the exact counts are persisted regardless).
type SuccessCompressInput struct {
	DBPath     string
	ProblemID  string
	MinSupport int
	JSONOutput bool
}

// SuccessListInput lists success-invariant revisions for a problem.
type SuccessListInput struct {
	DBPath     string
	ProblemID  string
	JSONOutput bool
}

// SuccessShowInput loads one revision (latest for the problem when id empty).
type SuccessShowInput struct {
	DBPath            string
	SuccessRevisionID string
	ProblemID         string
	JSONOutput        bool
}

// --- views ---

// SuccessCohortEvaluationView is one member verdict under a condition.
type SuccessCohortEvaluationView struct {
	ProposalID string `json:"proposal_id"`
	CohortRole string `json:"cohort_role"`
	Verdict    string `json:"verdict"`
	Strength   string `json:"verification_strength"`
}

// SuccessStrengthView is the verification-strength composition of support.
type SuccessStrengthView struct {
	Deterministic       int `json:"deterministic"`
	Reproducible        int `json:"reproducible"`
	IndependentEvidence int `json:"independent_evidence"`
	IndependentCritic   int `json:"independent_critic"`
	ModelJudgment       int `json:"single_model_judgment"`
}

// SuccessInvariantView is one persisted success invariant.
type SuccessInvariantView struct {
	ID                   string                        `json:"id"`
	PredicateFingerprint string                        `json:"predicate_fingerprint"`
	Predicate            string                        `json:"predicate"`
	Statement            string                        `json:"statement"`
	AbstractionLevel     string                        `json:"abstraction_level"`
	State                string                        `json:"state"`
	BrokenTargets        []string                      `json:"broken_targets"`
	CoverageNum          int                           `json:"progress_coverage_num"`
	CoverageDen          int                           `json:"progress_coverage_den"`
	ExclusionNum         int                           `json:"nonprogressor_exclusion_num"`
	ExclusionDen         int                           `json:"nonprogressor_exclusion_den"`
	CoverageOrdinal      string                        `json:"coverage_ordinal"`
	ExclusionOrdinal     string                        `json:"exclusion_ordinal"`
	DistinctSupport      int                           `json:"distinct_mechanism_support"`
	Strength             SuccessStrengthView           `json:"support_strength_composition"`
	CohortEvaluations    []SuccessCohortEvaluationView `json:"cohort_evaluations"`
}

// SuccessRevisionView is one full compression pass.
type SuccessRevisionView struct {
	ID                     string                 `json:"id"`
	ProblemID              string                 `json:"problem_id"`
	RunID                  string                 `json:"run_id"`
	CompressorVersion      string                 `json:"compressor_version"`
	PredicateSchema        string                 `json:"predicate_schema"`
	MinSupport             int                    `json:"min_support"`
	CohortHash             string                 `json:"cohort_hash"`
	IneligibleUnpersisted  int                    `json:"ineligible_unpersisted"`
	PendingReassessment    int                    `json:"pending_reassessment"`
	AmbiguousMembers       int                    `json:"ambiguous_members"`
	InadmissibleConditions int                    `json:"inadmissible_conditions"`
	Revision               int                    `json:"revision"`
	InvariantCount         int                    `json:"invariant_count"`
	CreatedAt              string                 `json:"created_at"`
	Invariants             []SuccessInvariantView `json:"invariants"`
}

// SuccessCompressResponse is returned by `newf successes compress`.
type SuccessCompressResponse struct {
	OK       bool                `json:"ok"`
	Command  string              `json:"command"`
	Store    string              `json:"store"`
	Created  bool                `json:"created"`
	Revision SuccessRevisionView `json:"revision"`
}

// SuccessListResponse is returned by `newf success-invariant list`.
type SuccessListResponse struct {
	OK        bool                  `json:"ok"`
	Command   string                `json:"command"`
	Store     string                `json:"store"`
	Revisions []SuccessRevisionView `json:"revisions"`
}

// SuccessShowResponse is returned by `newf success-invariant show`.
type SuccessShowResponse struct {
	OK       bool                `json:"ok"`
	Command  string              `json:"command"`
	Store    string              `json:"store"`
	Revision SuccessRevisionView `json:"revision"`
	// Selections is the compression execution history (v28, oldest first) when
	// the show resolved by problem: each row is one execution and the artifact
	// it selected (created OR reused). The LAST row is current guidance.
	Selections []CompressionSelectionView `json:"selections,omitempty"`
}

// CompressionSelectionView is one compression execution's selection.
type CompressionSelectionView struct {
	RunID             string `json:"run_id"`
	SuccessRevisionID string `json:"success_revision_id"`
	CreatedAt         string `json:"created_at"`
}

// --- cohort building (code-selected; the provider never nominates members) ---

// builtCohorts is the partitioned break-cohort material plus the named
// visibility gaps (pre-v17 unpersisted content; ambiguous verdicts).
type builtCohorts struct {
	Cohorts               []success.BreakCohort
	IneligibleUnpersisted int
	Ambiguous             int
	// PendingReassessment counts members whose NEWEST interpretation has no
	// compatible reassessment (round-2 F2): the selected evaluation assessed an
	// older revision, so the member is excluded from current guidance rather
	// than splicing an old outcome onto unassessed evidence.
	PendingReassessment int
	CohortHash          string
}

// buildBreakCohorts partitions the persisted (target, evaluated proposal)
// pairs into per-P cohorts (KTD-1). Progressors: partial_success/success;
// non-progressors: failure/partial_failure; unknown/verification_blocked
// counted ambiguous; rows without persisted canonical content counted
// ineligible — never silently included or fabricated.
func buildBreakCohorts(rows []store.BreakCohortRow) (builtCohorts, error) {
	out := builtCohorts{}
	byTarget := map[string]*success.BreakCohort{}
	var targetOrder []string
	var hashLines []string
	for _, r := range rows {
		// The identity line includes the CONTENT hash (F1): a revised
		// interpretation of the same mechanism (completeness/unresolved-claim
		// changes are fingerprint-invisible) changes what every condition C
		// evaluates against, so it must produce the next cohort revision.
		hashLines = append(hashLines, r.TargetInvariantID+"|"+r.ProposalID+"|"+r.EvaluationID+"|"+r.Result+"|"+r.Strength+"|"+r.ContentHash+"|"+r.LatestContentHash)
		// v27 finding 1: the selected evaluation has NO recorded content binding
		// on a multi-revision proposal. Which bytes it assessed is unknowable,
		// so its outcome must not become current support — the member is
		// pending until a BOUND reassessment of the current interpretation
		// exists. Its content fields are deliberately empty (never filled from
		// current bytes), so this check must precede the unpersisted-content
		// check below.
		if r.BindingUnknown {
			out.PendingReassessment++
			continue
		}
		if r.SignatureJSON == "" {
			out.IneligibleUnpersisted++
			continue
		}
		// Round-2 F2: the selected evaluation assessed an OLDER revision than
		// the proposal's newest interpretation. Carrying its outcome forward
		// would attribute the old assessment to unassessed evidence; the member
		// is pending until a compatible reassessment exists. (The latest hash is
		// part of the identity line above, so resolving the pending state
		// produces the next cohort revision.)
		if r.ContentHash != "" && r.LatestContentHash != "" && r.ContentHash != r.LatestContentHash {
			out.PendingReassessment++
			continue
		}
		var member success.Member
		var sig canon.MechanismSignature
		if err := json.Unmarshal([]byte(r.SignatureJSON), &sig); err != nil {
			return builtCohorts{}, fmt.Errorf("proposal %s: corrupt persisted signature: %w", r.ProposalID, err)
		}
		member = success.Member{ProposalID: r.ProposalID, EvaluationID: r.EvaluationID, Signature: sig, Result: domain.OutcomeClass(r.Result), Strength: r.Strength}
		cohort, ok := byTarget[r.TargetInvariantID]
		if !ok {
			cohort = &success.BreakCohort{TargetInvariantID: r.TargetInvariantID}
			byTarget[r.TargetInvariantID] = cohort
			targetOrder = append(targetOrder, r.TargetInvariantID)
		}
		switch domain.OutcomeClass(r.Result) {
		case domain.OutcomePartialSuccess, domain.OutcomeSuccess:
			cohort.Progressors = append(cohort.Progressors, member)
		case domain.OutcomeFailure, domain.OutcomePartialFailure:
			cohort.NonProgressors = append(cohort.NonProgressors, member)
		default:
			cohort.Ambiguous++
			out.Ambiguous++
		}
	}
	sort.Strings(targetOrder)
	for _, t := range targetOrder {
		out.Cohorts = append(out.Cohorts, *byTarget[t])
	}
	sort.Strings(hashLines)
	sum := sha256.Sum256([]byte(strings.Join(hashLines, "\n")))
	out.CohortHash = hex.EncodeToString(sum[:])
	return out, nil
}

// compressionRequestForCohorts projects the cohorts into the compact provider
// request (structured facts only; resolved canonical content).
func compressionRequestForCohorts(problemID string, minSupport int, cohorts []success.BreakCohort) provider.CompressionRequest {
	req := provider.CompressionRequest{ProblemID: problemID, MinSupport: minSupport}
	memberFacts := func(m success.Member) provider.CohortMemberFacts {
		return provider.CohortMemberFacts{
			ProposalID: m.ProposalID,
			Result:     m.Result,
			Strength:   m.Strength,
			Preserves:  resolvedFieldIDs(m.Signature.Preserves),
			Operators:  resolvedFieldIDs(m.Signature.Operators),
		}
	}
	for _, c := range cohorts {
		facts := provider.BreakCohortFacts{TargetInvariantID: c.TargetInvariantID}
		for _, m := range c.Progressors {
			facts.Progressors = append(facts.Progressors, memberFacts(m))
		}
		for _, m := range c.NonProgressors {
			facts.NonProgressors = append(facts.NonProgressors, memberFacts(m))
		}
		req.Cohorts = append(req.Cohorts, facts)
	}
	return req
}

func successRevisionView(rec store.SuccessRevisionRecord) SuccessRevisionView {
	view := SuccessRevisionView{
		ID:                     rec.ID,
		ProblemID:              rec.ProblemID,
		RunID:                  rec.RunID,
		CompressorVersion:      rec.CompressorVersion,
		PredicateSchema:        rec.PredicateSchema,
		MinSupport:             rec.MinSupport,
		CohortHash:             rec.CohortHash,
		IneligibleUnpersisted:  rec.IneligibleUnpersisted,
		PendingReassessment:    rec.PendingReassessment,
		AmbiguousMembers:       rec.AmbiguousMembers,
		InadmissibleConditions: rec.InadmissibleConditions,
		Revision:               rec.Revision,
		InvariantCount:         rec.InvariantCount,
		CreatedAt:              rec.CreatedAt,
	}
	for _, si := range rec.Invariants {
		sv := SuccessInvariantView{
			ID:                   si.ID,
			PredicateFingerprint: si.PredicateFingerprint,
			Predicate:            si.PredicateJSON,
			Statement:            si.Statement,
			AbstractionLevel:     si.AbstractionLevel,
			State:                "proposed",
			BrokenTargets:        si.BrokenTargets,
			CoverageNum:          si.CoverageNum,
			CoverageDen:          si.CoverageDen,
			ExclusionNum:         si.ExclusionNum,
			ExclusionDen:         si.ExclusionDen,
			CoverageOrdinal:      si.CoverageOrdinal,
			ExclusionOrdinal:     si.ExclusionOrdinal,
			DistinctSupport:      si.DistinctSupport,
			Strength: SuccessStrengthView{
				Deterministic:       si.StrengthDeterministic,
				Reproducible:        si.StrengthReproducible,
				IndependentEvidence: si.StrengthIndependentEvidence,
				IndependentCritic:   si.StrengthIndependentCritic,
				ModelJudgment:       si.StrengthModelJudgment,
			},
		}
		for _, ce := range si.CohortEvaluations {
			sv.CohortEvaluations = append(sv.CohortEvaluations, SuccessCohortEvaluationView{
				ProposalID: ce.ProposalID, CohortRole: ce.CohortRole, Verdict: ce.Verdict, Strength: ce.Strength,
			})
		}
		view.Invariants = append(view.Invariants, sv)
	}
	return view
}

// defaultSuccessMinSupport is the distinct-mechanism threshold provenance
// default (exact counts are persisted regardless; sparse early corpora start
// at 1).
const defaultSuccessMinSupport = 1

// CompressSuccesses runs the M6.1 compress operator under the run lifecycle:
// code selects the break cohorts from persisted verdicts, the compressor
// proposes candidate conditions C, every C is admitted (grammar + pinned
// vocabulary + no outcome reads) and evaluated by code against every cohort
// member, and the resulting success invariants are persisted as an immutable,
// revisioned artifact. Idempotent on the cohort-hash identity tuple.
func (a *App) CompressSuccesses(ctx context.Context, input SuccessCompressInput) (SuccessCompressResponse, error) {
	minSupport := input.MinSupport
	if minSupport <= 0 {
		minSupport = defaultSuccessMinSupport
	}

	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return SuccessCompressResponse{}, err
	}
	defer repoStore.Close()

	if _, err := repoStore.GetProblem(ctx, input.ProblemID); err != nil {
		return SuccessCompressResponse{}, err
	}
	rows, err := repoStore.ListBreakCohortRows(ctx, input.ProblemID)
	if err != nil {
		return SuccessCompressResponse{}, err
	}
	built, err := buildBreakCohorts(rows)
	if err != nil {
		return SuccessCompressResponse{}, err
	}

	// The vocabulary the corpus was canonicalized under gates condition
	// references (same pinned-vocab discipline as mining). A store failure
	// must not silently fall back to the default vocabulary — that would
	// change compression semantics with no recorded cause; only a genuinely
	// absent cluster run may use the default.
	vocabVersion := canon.VocabularyMechanismV1
	latest, found, err := repoStore.LatestClusterRun(ctx, input.ProblemID)
	if err != nil {
		return SuccessCompressResponse{}, err
	}
	if found {
		clusterRun, gerr := repoStore.GetClusterRun(ctx, latest)
		if gerr != nil {
			return SuccessCompressResponse{}, gerr
		}
		vocabVersion = clusterRun.VocabularyVersion
	}
	vocab, err := a.loadVocabulary(ctx, repoStore, vocabVersion)
	if err != nil {
		return SuccessCompressResponse{}, err
	}

	now := a.now()
	run, err := repoStore.CreateRun(ctx, domain.NewRun{
		ID:          domain.NewRunID(now),
		ProblemID:   input.ProblemID,
		Operation:   "successes compress",
		Status:      domain.RunStatusRunning,
		InputRef:    "problem:" + input.ProblemID,
		ToolName:    "newf",
		ToolVersion: a.version,
		StartedAt:   now,
		CompletedAt: now,
	})
	if err != nil {
		return SuccessCompressResponse{}, err
	}

	compressor := a.successCompressorFn
	if compressor == nil {
		compressor = provider.NewDerivingFixtureCompressor()
	}
	req := compressionRequestForCohorts(input.ProblemID, minSupport, built.Cohorts)
	resp, err := compressor.Compress(ctx, req)
	if err != nil {
		a.failRun(ctx, repoStore, run.ID, err)
		return SuccessCompressResponse{}, err
	}

	// Every proposed condition passes the shared admissibility gate. An
	// inadmissible condition (bad grammar, out-of-vocabulary reference, outcome
	// read) is SKIPPED AND COUNTED — recorded on the revision as
	// inadmissible_conditions — never silently dropped and never stored
	// unevaluated. This keeps a fixture or model that over-generates from
	// bricking the whole pass while preserving full visibility.
	conditions := make([]success.Condition, 0, len(resp.Proposals))
	inadmissible := 0
	for _, p := range resp.Proposals {
		if verr := invariant.AdmitCandidate(p.Predicate, vocab); verr != nil {
			inadmissible++
			continue
		}
		conditions = append(conditions, success.Condition{
			TargetInvariantID: p.TargetInvariantID,
			Predicate:         p.Predicate,
			Statement:         p.Statement,
			AbstractionLevel:  p.AbstractionLevel,
		})
	}
	candidates := success.Compress(conditions, built.Cohorts)

	record, err := successRevisionRecord(input.ProblemID, run.ID, minSupport, built, resp, req.Fingerprint(), candidates, now)
	if err != nil {
		a.failRun(ctx, repoStore, run.ID, err)
		return SuccessCompressResponse{}, err
	}
	record.InadmissibleConditions = inadmissible
	result, err := repoStore.PersistSuccessRevision(ctx, record)
	if err != nil {
		a.failRun(ctx, repoStore, run.ID, err)
		return SuccessCompressResponse{}, err
	}
	if err := a.finalizeRun(ctx, repoStore, run.ID, nil); err != nil {
		return SuccessCompressResponse{}, err
	}

	return SuccessCompressResponse{
		OK:       true,
		Command:  "successes compress",
		Store:    dbPath,
		Created:  result.Created,
		Revision: successRevisionView(result.Record),
	}, nil
}

// successRevisionRecord flattens evaluated candidates into store rows.
func successRevisionRecord(problemID, runID string, minSupport int, built builtCohorts, resp provider.CompressionResponse, requestHash string, candidates []success.Candidate, now time.Time) (store.SuccessRevisionRecord, error) {
	rec := store.SuccessRevisionRecord{
		ID:                    domain.NewSuccessRevisionID(now),
		ProblemID:             problemID,
		RunID:                 runID,
		CompressorVersion:     provider.SuccessCompressorVersion,
		PredicateSchema:       invariant.PredicateSchemaV1,
		MinSupport:            minSupport,
		CohortHash:            built.CohortHash,
		IneligibleUnpersisted: built.IneligibleUnpersisted,
		PendingReassessment:   built.PendingReassessment,
		AmbiguousMembers:      built.Ambiguous,
		InvariantCount:        len(candidates),
		CreatedAt:             now.Format(timeLayout),
		Invocation: store.InvariantProviderInvocation{
			ID:              domain.NewProviderInvocationID(now),
			RunID:           runID,
			ProviderName:    resp.Metadata.ProviderName,
			ProviderVersion: resp.Metadata.ProviderVersion,
			ModelName:       resp.Metadata.ModelName,
			SchemaVersion:   resp.Metadata.SchemaVersion,
			RequestHash:     requestHash,
			RequestPayload:  resp.RequestPayload,
			ResponsePayload: resp.ResponsePayload,
			CreatedAt:       now.Format(timeLayout),
		},
	}
	for i, c := range candidates {
		predJSON, err := invariant.MarshalCanonical(c.Predicate)
		if err != nil {
			return store.SuccessRevisionRecord{}, fmt.Errorf("canonicalize condition: %w", err)
		}
		row := store.SuccessInvariantRow{
			ID:                          domain.NewSuccessInvariantID(now),
			PredicateFingerprint:        c.PredicateFingerprint,
			PredicateJSON:               predJSON,
			Statement:                   c.Statement,
			AbstractionLevel:            c.AbstractionLevel,
			CoverageNum:                 c.CoverageNum,
			CoverageDen:                 c.CoverageDen,
			ExclusionNum:                c.ExclusionNum,
			ExclusionDen:                c.ExclusionDen,
			CoverageOrdinal:             string(c.CoverageOrdinal),
			ExclusionOrdinal:            string(c.ExclusionOrdinal),
			DistinctSupport:             c.DistinctMechanismSupport,
			StrengthDeterministic:       c.Support.Deterministic,
			StrengthReproducible:        c.Support.Reproducible,
			StrengthIndependentEvidence: c.Support.IndependentEvidence,
			StrengthIndependentCritic:   c.Support.IndependentCritic,
			StrengthModelJudgment:       c.Support.ModelJudgment,
			Ordinal:                     i,
			BrokenTargets:               c.TargetInvariantIDs,
		}
		for _, ce := range c.CohortEvaluations {
			row.CohortEvaluations = append(row.CohortEvaluations, store.SuccessCohortEvaluationRow{
				ProposalID:   ce.ProposalID,
				EvaluationID: ce.EvaluationID,
				CohortRole:   ce.Role,
				Verdict:      string(ce.Verdict),
				Strength:     normalizeStrength(ce.Strength),
			})
		}
		rec.Invariants = append(rec.Invariants, row)
	}
	return rec, nil
}

// normalizeStrength maps an absent/unknown persisted strength to the weakest
// class rather than inventing a stronger one.
func normalizeStrength(s string) string {
	switch s {
	case "deterministic", "reproducible", "independent-evidence", "independent-critic", "single-model-judgment":
		return s
	default:
		return "single-model-judgment"
	}
}

// ListSuccesses lists success-invariant revision headers for a problem.
func (a *App) ListSuccesses(ctx context.Context, input SuccessListInput) (SuccessListResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return SuccessListResponse{}, err
	}
	defer repoStore.Close()

	recs, err := repoStore.ListSuccessRevisions(ctx, input.ProblemID)
	if err != nil {
		return SuccessListResponse{}, err
	}
	resp := SuccessListResponse{OK: true, Command: "success-invariant list", Store: dbPath}
	for _, r := range recs {
		resp.Revisions = append(resp.Revisions, successRevisionView(r))
	}
	return resp, nil
}

// ShowSuccess loads one full revision (latest when no id given).
func (a *App) ShowSuccess(ctx context.Context, input SuccessShowInput) (SuccessShowResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return SuccessShowResponse{}, err
	}
	defer repoStore.Close()

	id := input.SuccessRevisionID
	if id == "" {
		latest, found, lerr := repoStore.LatestSelectedSuccessRevision(ctx, input.ProblemID)
		if lerr != nil {
			return SuccessShowResponse{}, lerr
		}
		if !found {
			return SuccessShowResponse{}, fmt.Errorf("no success-invariant revision for problem %s; run `successes compress` first", input.ProblemID)
		}
		id = latest
	}
	rec, err := repoStore.GetSuccessRevision(ctx, id)
	if err != nil {
		return SuccessShowResponse{}, err
	}
	resp := SuccessShowResponse{OK: true, Command: "success-invariant show", Store: dbPath, Revision: successRevisionView(rec)}
	// Selection history (v28): the audit trail behind current guidance.
	if input.ProblemID != "" {
		selections, serr := repoStore.ListCompressionSelections(ctx, input.ProblemID)
		if serr != nil {
			return SuccessShowResponse{}, serr
		}
		for _, s := range selections {
			resp.Selections = append(resp.Selections, CompressionSelectionView{
				RunID: s.RunID, SuccessRevisionID: s.SuccessRevisionID, CreatedAt: s.CreatedAt,
			})
		}
	}
	return resp, nil
}
