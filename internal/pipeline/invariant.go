package pipeline

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/invariant"
	"github.com/instagrim-dev/newf/internal/provider"
	"github.com/instagrim-dev/newf/internal/store"
)

// signatureFromRecordWithProvenance rebuilds a canon.MechanismSignature from
// persisted rows INCLUDING full epistemic provenance. It deliberately does not
// reuse signatureFromRecord: that helper is lossy for epistemic purposes (it
// drops boundary Status and never sets Posture/Outcome provenance), which would
// make every boundary/posture/outcome predicate's epistemic composition
// structurally empty — an unearned status R6 forbids. This reader maps
// SignatureRecord.ClaimStatus (set fields), boundary ClaimStatus, PostureStatus
// per axis, and OutcomeStatus so the explicit/inferred/... of a matched claim is
// available for the specific axis a predicate reads.
func signatureFromRecordWithProvenance(rec store.SignatureRecord) canon.MechanismSignature {
	sig := signatureFromRecord(rec)
	for i, b := range rec.Boundaries {
		if i < len(sig.Boundaries) {
			sig.Boundaries[i].Status = domain.ClaimStatus(defaultStatus(b.ClaimStatus))
			sig.Boundaries[i].SupportSnapshotID = b.SupportSnapshotID
			sig.Boundaries[i].SupportLocator = b.SupportLocator
		}
	}
	sig.PostureProvenance = canon.PostureProvenance{
		Locality:     domain.ClaimStatus(defaultStatus(rec.PostureStatus["locality"])),
		Construction: domain.ClaimStatus(defaultStatus(rec.PostureStatus["construction"])),
		Uncertainty:  domain.ClaimStatus(defaultStatus(rec.PostureStatus["uncertainty"])),
	}
	sig.OutcomeProvenance = domain.ClaimStatus(defaultStatus(rec.OutcomeStatus))
	return sig
}

// defaultStatus preserves a missing persisted status as unknown, never explicit.
func defaultStatus(s string) string {
	if s == "" {
		return string(domain.ClaimUnknown)
	}
	return s
}

// familiesForFailureSpace rehydrates the distinct mechanism families behind a
// failure space: for each cluster, every non-redundant member signature is
// loaded (per-member outcome comes from signature_outcomes.class via
// GetSignature — it is not on the member row) with full epistemic provenance,
// so mixed families can split member-wise and support carries real per-claim
// strength.
func familiesForFailureSpace(ctx context.Context, repoStore problemStore, clusterRun store.ClusterRunRecord) ([]invariant.Family, error) {
	families := make([]invariant.Family, 0, len(clusterRun.Clusters))
	for _, c := range clusterRun.Clusters {
		fam := invariant.Family{
			ClusterID:    c.ID,
			OutcomeClass: domain.OutcomeClass(c.OutcomeClass),
			OutcomeMixed: c.OutcomeMixed,
		}
		redundantMembers := 0
		for _, m := range c.Members {
			if m.Redundant {
				redundantMembers++
				continue
			}
			rec, err := repoStore.GetSignature(ctx, m.SignatureID)
			if err != nil {
				return nil, fmt.Errorf("load member signature %s: %w", m.SignatureID, err)
			}
			sig := signatureFromRecordWithProvenance(rec)
			fam.Members = append(fam.Members, invariant.Member{
				SignatureID:  m.SignatureID,
				Signature:    sig,
				OutcomeClass: domain.OutcomeClass(rec.OutcomeClass),
			})
		}
		// A family whose members are all #11-redundant contributes no
		// non-redundant support unit.
		fam.Redundant = len(c.Members) > 0 && redundantMembers == len(c.Members)
		if fam.Redundant {
			// Still evaluate it (visibility), via its representative.
			rec, err := repoStore.GetSignature(ctx, c.RepresentativeSignatureID)
			if err != nil {
				return nil, fmt.Errorf("load representative signature %s: %w", c.RepresentativeSignatureID, err)
			}
			fam.Members = append(fam.Members, invariant.Member{
				SignatureID:  c.RepresentativeSignatureID,
				Signature:    signatureFromRecordWithProvenance(rec),
				OutcomeClass: domain.OutcomeClass(rec.OutcomeClass),
			})
		}
		families = append(families, fam)
	}
	return families, nil
}

// miningRequestForFamilies projects families into the compact, deterministic
// MiningRequest (KTD-1): per family, the representative's resolved canonical
// content plus outcome/redundancy facts. The provider sees structured facts,
// never free text; support is computed later against every member, not this
// projection.
func miningRequestForFamilies(problemID, failureSpaceID string, minSupport int, families []invariant.Family) provider.MiningRequest {
	req := provider.MiningRequest{
		ProblemID:      problemID,
		FailureSpaceID: failureSpaceID,
		MinSupport:     minSupport,
	}
	for _, fam := range families {
		if len(fam.Members) == 0 {
			continue
		}
		rep := fam.Members[0].Signature
		mf := provider.MiningFamily{
			ClusterID:    fam.ClusterID,
			OutcomeClass: fam.OutcomeClass,
			OutcomeMixed: fam.OutcomeMixed,
			Redundant:    fam.Redundant,
			Preserves:    resolvedFieldIDs(rep.Preserves),
			Operators:    resolvedFieldIDs(rep.Operators),
			Assumptions:  resolvedFieldIDs(rep.Assumptions),
			Locality:     rep.Posture.Locality,
			Construction: rep.Posture.Construction,
			Uncertainty:  rep.Posture.Uncertainty,
		}
		req.Families = append(req.Families, mf)
	}
	sort.Slice(req.Families, func(i, j int) bool { return req.Families[i].ClusterID < req.Families[j].ClusterID })
	return req
}

func resolvedFieldIDs(claims []canon.FieldClaim) []domain.CanonicalID {
	seen := map[domain.CanonicalID]struct{}{}
	var out []domain.CanonicalID
	for _, c := range claims {
		if c.State != domain.ResolutionResolved || c.CanonicalID == "" {
			continue
		}
		if _, dup := seen[c.CanonicalID]; dup {
			continue
		}
		seen[c.CanonicalID] = struct{}{}
		out = append(out, c.CanonicalID)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

// --- U6 service inputs ---

// InvariantMineInput requests a mining pass over a failure space (latest for
// the problem when FailureSpaceID is empty). MinSupport <= 0 defaults to 2.
type InvariantMineInput struct {
	DBPath         string
	ProblemID      string
	FailureSpaceID string
	MinSupport     int
	JSONOutput     bool
}

// InvariantListInput lists invariant revisions for a problem.
type InvariantListInput struct {
	DBPath     string
	ProblemID  string
	JSONOutput bool
}

// InvariantShowInput loads one invariant revision (latest for the problem when
// InvariantRevisionID is empty).
type InvariantShowInput struct {
	DBPath              string
	InvariantRevisionID string
	ProblemID           string
	JSONOutput          bool
}

// defaultMinSupport is the distinct-family threshold for `recurring` when the
// caller does not set one.
const defaultMinSupport = 2

// MineInvariants runs the M4.2 compress operator under the real run lifecycle:
// it rehydrates the failure space's families (per-member, with epistemic
// provenance), asks the miner for typed predicate proposals, deterministically
// evaluates every proposal against persisted signatures (code computes support;
// the model cannot self-certify), and persists an immutable, revisioned
// InvariantRevision. Idempotent on (problem, failure_space, miner_version,
// predicate_schema, min_support).
func (a *App) MineInvariants(ctx context.Context, input InvariantMineInput) (InvariantMineResponse, error) {
	minSupport := input.MinSupport
	if minSupport <= 0 {
		minSupport = defaultMinSupport
	}

	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return InvariantMineResponse{}, err
	}
	defer repoStore.Close()

	if _, err := repoStore.GetProblem(ctx, input.ProblemID); err != nil {
		return InvariantMineResponse{}, err
	}
	fs, err := a.resolveFailureSpace(ctx, repoStore, input.FailureSpaceID, input.ProblemID)
	if err != nil {
		return InvariantMineResponse{}, err
	}
	if fs.ProblemID != input.ProblemID {
		return InvariantMineResponse{}, fmt.Errorf("failure space %s does not belong to problem %s", fs.ID, input.ProblemID)
	}
	clusterRun, err := repoStore.GetClusterRun(ctx, fs.ClusterRunID)
	if err != nil {
		return InvariantMineResponse{}, err
	}

	families, err := familiesForFailureSpace(ctx, repoStore, clusterRun)
	if err != nil {
		return InvariantMineResponse{}, err
	}
	req := miningRequestForFamilies(input.ProblemID, fs.ID, minSupport, families)

	now := a.now()
	run, err := repoStore.CreateRun(ctx, domain.NewRun{
		ID:          domain.NewRunID(now),
		ProblemID:   input.ProblemID,
		Operation:   "invariants mine",
		Status:      domain.RunStatusRunning,
		InputRef:    "failure_space:" + fs.ID,
		ToolName:    "newf",
		ToolVersion: a.version,
		StartedAt:   now,
		CompletedAt: now,
	})
	if err != nil {
		return InvariantMineResponse{}, err
	}

	miner := a.invariantMinerFn
	if miner == nil {
		miner = provider.NewDerivingFixtureInvariantMiner()
	}
	resp, err := miner.Mine(ctx, req)
	if err != nil {
		a.failRun(ctx, repoStore, run.ID, err)
		return InvariantMineResponse{}, err
	}

	// Every proposal is validated and re-evaluated by code; an invalid predicate
	// fails the run rather than being silently stored.
	proposals := make([]invariant.Proposal, 0, len(resp.Proposals))
	for _, p := range resp.Proposals {
		if verr := p.Predicate.Validate(); verr != nil {
			a.failRun(ctx, repoStore, run.ID, verr)
			return InvariantMineResponse{}, verr
		}
		proposals = append(proposals, invariant.Proposal{
			Predicate:             p.Predicate,
			Statement:             p.Statement,
			AbstractionLevel:      p.AbstractionLevel,
			ObstructionHypothesis: p.ObstructionHypothesis,
		})
	}
	cands := invariant.EvaluateCandidates(proposals, families, minSupport)

	record, err := invariantRevisionRecord(input.ProblemID, fs, run.ID, minSupport, resp, req.Fingerprint(), cands, now)
	if err != nil {
		a.failRun(ctx, repoStore, run.ID, err)
		return InvariantMineResponse{}, err
	}
	result, err := repoStore.PersistInvariantRevision(ctx, record)
	if err != nil {
		a.failRun(ctx, repoStore, run.ID, err)
		return InvariantMineResponse{}, err
	}
	if err := a.finalizeRun(ctx, repoStore, run.ID, nil); err != nil {
		return InvariantMineResponse{}, err
	}

	return InvariantMineResponse{
		OK:       true,
		Command:  "invariants mine",
		Store:    dbPath,
		Created:  result.Created,
		Revision: invariantRevisionView(result.Record),
	}, nil
}

// ListInvariants lists invariant-revision headers for a problem.
func (a *App) ListInvariants(ctx context.Context, input InvariantListInput) (InvariantListResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return InvariantListResponse{}, err
	}
	defer repoStore.Close()

	recs, err := repoStore.ListInvariantRevisions(ctx, input.ProblemID)
	if err != nil {
		return InvariantListResponse{}, err
	}
	resp := InvariantListResponse{OK: true, Command: "invariant list", Store: dbPath}
	for _, r := range recs {
		resp.Revisions = append(resp.Revisions, InvariantRevisionSummaryView{
			ID:             r.ID,
			FailureSpaceID: r.FailureSpaceID,
			MinerVersion:   r.MinerVersion,
			MinSupport:     r.MinSupport,
			Revision:       r.Revision,
			CandidateCount: r.CandidateCount,
			CreatedAt:      r.CreatedAt,
		})
	}
	return resp, nil
}

// ShowInvariant loads a full invariant revision (latest for the problem when
// no id is given).
func (a *App) ShowInvariant(ctx context.Context, input InvariantShowInput) (InvariantShowResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return InvariantShowResponse{}, err
	}
	defer repoStore.Close()

	id := input.InvariantRevisionID
	if id == "" {
		latest, found, lerr := repoStore.LatestInvariantRevision(ctx, input.ProblemID)
		if lerr != nil {
			return InvariantShowResponse{}, lerr
		}
		if !found {
			return InvariantShowResponse{}, fmt.Errorf("no invariant revision for problem %s; run `invariants mine` first", input.ProblemID)
		}
		id = latest
	}
	rec, err := repoStore.GetInvariantRevision(ctx, id)
	if err != nil {
		return InvariantShowResponse{}, err
	}
	return InvariantShowResponse{
		OK:       true,
		Command:  "invariant show",
		Store:    dbPath,
		Revision: invariantRevisionView(rec),
	}, nil
}

// invariantRevisionView maps a persisted revision to its stable view.
func invariantRevisionView(rec store.InvariantRevisionRecord) InvariantRevisionView {
	view := InvariantRevisionView{
		ID:              rec.ID,
		ProblemID:       rec.ProblemID,
		FailureSpaceID:  rec.FailureSpaceID,
		ClusterRunID:    rec.ClusterRunID,
		RunID:           rec.RunID,
		MinerVersion:    rec.MinerVersion,
		PredicateSchema: rec.PredicateSchema,
		MinSupport:      rec.MinSupport,
		Revision:        rec.Revision,
		CandidateCount:  rec.CandidateCount,
		CreatedAt:       rec.CreatedAt,
	}
	for _, c := range rec.Candidates {
		cv := CandidateInvariantView{
			ID:                           c.ID,
			PredicateFingerprint:         c.PredicateFingerprint,
			Predicate:                    c.PredicateJSON,
			Statement:                    c.Statement,
			AbstractionLevel:             c.AbstractionLevel,
			State:                        "proposed",
			AssociationStatus:            c.AssociationStatus,
			ObstructionIsModelHypothesis: c.ObstructionIsModelHypothesis,
			DistinctFamilySupport:        c.DistinctFamilySupport,
			FailureCoverageNum:           c.FailureCoverageNum,
			FailureCoverageDen:           c.FailureCoverageDen,
			SupportExplicitCount:         c.SupportExplicitCount,
			SupportInferredCount:         c.SupportInferredCount,
			SupportOtherCount:            c.SupportOtherCount,
		}
		for _, fe := range c.FamilyEvaluations {
			cv.FamilyEvaluations = append(cv.FamilyEvaluations, InvariantFamilyEvaluationView{
				ClusterID:     fe.ClusterID,
				OutcomeClass:  fe.OutcomeClass,
				Role:          fe.Role,
				Verdict:       fe.Verdict,
				ExplicitCount: fe.ExplicitCount,
				InferredCount: fe.InferredCount,
				OtherCount:    fe.OtherCount,
			})
		}
		for _, ce := range c.Counterexamples {
			cv.Counterexamples = append(cv.Counterexamples, InvariantCounterexampleView{
				ClusterID: ce.ClusterID,
				Reason:    ce.Reason,
			})
		}
		view.Candidates = append(view.Candidates, cv)
	}
	return view
}

// invariantRevisionRecord flattens evaluated candidates into store rows.
func invariantRevisionRecord(problemID string, fs store.FailureSpaceRecord, runID string, minSupport int, resp provider.MiningResponse, requestHash string, cands []invariant.Candidate, now time.Time) (store.InvariantRevisionRecord, error) {
	rec := store.InvariantRevisionRecord{
		ID:              domain.NewInvariantRevisionID(now),
		ProblemID:       problemID,
		FailureSpaceID:  fs.ID,
		ClusterRunID:    fs.ClusterRunID,
		RunID:           runID,
		MinerVersion:    provider.InvariantMinerVersion,
		PredicateSchema: invariant.PredicateSchemaV1,
		MinSupport:      minSupport,
		CandidateCount:  len(cands),
		CreatedAt:       now.Format(timeLayout),
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
	for i, c := range cands {
		predJSON, err := invariant.MarshalCanonical(c.Predicate)
		if err != nil {
			return store.InvariantRevisionRecord{}, fmt.Errorf("canonicalize predicate: %w", err)
		}
		row := store.CandidateInvariantRow{
			ID:                           domain.NewCandidateInvariantID(now),
			PredicateFingerprint:         c.PredicateFingerprint,
			PredicateJSON:                predJSON,
			Statement:                    c.Statement,
			AbstractionLevel:             c.AbstractionLevel,
			AssociationStatus:            string(c.AssociationStatus),
			ObstructionIsModelHypothesis: c.ObstructionIsHypothesis,
			DistinctFamilySupport:        c.DistinctFamilySupport,
			FailureCoverageNum:           c.FailureCoverageNum,
			FailureCoverageDen:           c.FailureCoverageDen,
			SupportExplicitCount:         c.SupportEpistemic.Explicit,
			SupportInferredCount:         c.SupportEpistemic.Inferred,
			SupportOtherCount:            c.SupportEpistemic.Other,
			Ordinal:                      i,
		}
		for _, fe := range c.FamilyEvaluations {
			row.FamilyEvaluations = append(row.FamilyEvaluations, store.InvariantFamilyEvaluationRow{
				ClusterID:     fe.ClusterID,
				OutcomeClass:  fe.OutcomeClass,
				Role:          string(fe.Role),
				Verdict:       string(fe.Verdict),
				ExplicitCount: fe.Epistemic.Explicit,
				InferredCount: fe.Epistemic.Inferred,
				OtherCount:    fe.Epistemic.Other,
			})
		}
		for _, ce := range c.Counterexamples {
			row.Counterexamples = append(row.Counterexamples, store.InvariantCounterexampleRow{
				ClusterID: ce.ClusterID,
				Reason:    ce.Reason,
			})
		}
		rec.Candidates = append(rec.Candidates, row)
	}
	return rec, nil
}
