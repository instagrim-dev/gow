package pipeline

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"time"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/invariant"
	"github.com/instagrim-dev/newf/internal/provider"
	"github.com/instagrim-dev/newf/internal/store"
	"github.com/instagrim-dev/newf/internal/verify"
)

// EvaluateInput requests an evaluation pass. Exactly one selector applies:
// ProposalID evaluates a single proposal; otherwise the latest generation for
// ProblemID is evaluated (All => every un-evaluated proposal; else the same).
type EvaluateInput struct {
	DBPath     string
	ProblemID  string
	ProposalID string
	All        bool
	Mode       string // 'proposal' (default); 'holdout' is refused (R9)
	JSONOutput bool
}

// EvaluationListInput lists evaluation runs for a problem.
type EvaluationListInput struct {
	DBPath     string
	ProblemID  string
	JSONOutput bool
}

// EvaluationShowInput loads one evaluation run by id.
type EvaluationShowInput struct {
	DBPath       string
	EvaluationID string
	JSONOutput   bool
}

// Evaluate runs the M5.2 evaluate operator under the real run lifecycle. For
// each selected frontier proposal it builds a code-owned VerificationContext
// from persisted state (the per-target violation verdicts M5.1 computed, plus
// the nearest families re-checked against each target predicate), routes it
// cheap-first / strongest-decisive through the verifier hierarchy, and persists
// a durable Evaluation stamped with the deciding verifier's kind + strength.
// The proposal's result is populated in the same transaction; failures re-enter
// the atlas. Holdout mode is refused pending M7 (R9).
func (a *App) Evaluate(ctx context.Context, input EvaluateInput) (EvaluateResponse, error) {
	if input.Mode == "holdout" {
		return EvaluateResponse{}, fmt.Errorf("holdout evaluation mode is deferred to M7; only mode=proposal is supported")
	}

	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return EvaluateResponse{}, err
	}
	defer repoStore.Close()

	if _, err := repoStore.GetProblem(ctx, input.ProblemID); err != nil {
		return EvaluateResponse{}, err
	}

	// Resolve the frontier generation to evaluate. By-id evaluation must reach
	// the proposal's OWNING generation (the one that first wrote it), not the
	// latest — cross-run dedup keeps the proposal under its original generation,
	// so a later fully-deduped generation owns zero rows and would hide it
	// (finding 2). Batch evaluation resolves the latest generation that actually
	// owns proposal rows for the same reason.
	var genID string
	var found bool
	if input.ProposalID != "" {
		genID, found, err = repoStore.FindProposalGeneration(ctx, input.ProblemID, input.ProposalID)
		if err != nil {
			return EvaluateResponse{}, err
		}
		if !found {
			return EvaluateResponse{}, fmt.Errorf("proposal %s not found for problem %s", input.ProposalID, input.ProblemID)
		}
	} else {
		genID, found, err = repoStore.LatestFrontierGenerationWithProposals(ctx, input.ProblemID)
		if err != nil {
			return EvaluateResponse{}, err
		}
		if !found {
			return EvaluateResponse{}, fmt.Errorf("no frontier generation with proposals for problem %s; run `frontier generate` first", input.ProblemID)
		}
	}
	gen, err := repoStore.GetFrontierGeneration(ctx, genID)
	if err != nil {
		return EvaluateResponse{}, err
	}

	// Select proposals: a single id, or all un-evaluated proposals in the gen.
	selected := make([]store.FrontierProposalRow, 0, len(gen.Proposals))
	for _, p := range gen.Proposals {
		if input.ProposalID != "" {
			if p.ID == input.ProposalID {
				selected = append(selected, p)
			}
			continue
		}
		if p.Result.Valid {
			continue // already evaluated; re-evaluation is an explicit single-id op
		}
		selected = append(selected, p)
	}
	if input.ProposalID != "" && len(selected) == 0 {
		// The proposal resolved to this generation but is absent from its owned
		// rows — a store inconsistency, not a missing artifact.
		return EvaluateResponse{}, fmt.Errorf("proposal %s resolved to generation %s but is absent from its owned proposals", input.ProposalID, gen.ID)
	}

	// Target predicates are the parsed predicates of the CURRENTLY targetable
	// invariants, keyed by id. A proposal's cached target verdict for a target
	// that is no longer here (it became weaken/falsified since generation) is
	// dropped consistently from BOTH the deterministic check input and the
	// comparison population (G4): a proposal must not gain a better evaluation
	// merely because a hypothesis it targeted became less credible, and its
	// comparison evidence must not silently disappear while its claimed break
	// stays. Stale targets are surfaced as an inapplicable/unknown result, not a
	// free partial_success.
	predicates, err := a.targetPredicates(ctx, repoStore, input.ProblemID)
	if err != nil {
		return EvaluateResponse{}, err
	}
	// Representative signature per cluster in the generation's population, so each
	// proposal's RECORDED nearest families (not every representative, G1) can be
	// re-checked against the target predicates.
	reps, err := a.clusterRepresentatives(ctx, repoStore, gen.ClusterRunID)
	if err != nil {
		return EvaluateResponse{}, err
	}

	model := a.modelVerifier()
	verifiers := a.verifiers(model)
	now := a.now()
	run, err := repoStore.CreateRun(ctx, domain.NewRun{
		ID:          domain.NewRunID(now),
		ProblemID:   input.ProblemID,
		Operation:   "evaluate",
		Status:      domain.RunStatusRunning,
		InputRef:    "frontier_generation_run:" + gen.ID,
		ToolName:    "newf",
		ToolVersion: a.version,
		StartedAt:   now,
		CompletedAt: now,
	})
	if err != nil {
		return EvaluateResponse{}, err
	}

	record := store.EvaluationRunRecord{
		ID:                      domain.NewEvaluationRunID(now),
		ProblemID:               input.ProblemID,
		RunID:                   run.ID,
		FrontierGenerationRunID: gen.ID,
		ClusterRunID:            gen.ClusterRunID,
		Mode:                    "proposal",
		RoutingPolicy:           "cheap-first",
		CreatedAt:               now.Format(timeLayout),
	}

	for _, p := range selected {
		vc, staleTarget := verificationContextForProposal(p, predicates, reps)
		var decision verify.Decision
		if staleTarget {
			// H5: a claimed target is no longer targetable (weaken/falsified). We do
			// NOT erase it and route the reduced context — that would silently change
			// the question and let a decisive model adapter award a better verdict to
			// an unchanged proposal. The proposal requires explicit reassessment
			// against a fresh context; record a non-decisive, non-routed result.
			decision = verify.Decision{
				Verdict:  verify.VerdictVerificationBlocked,
				Kind:     verify.KindDeterministicCheck,
				Strength: verify.StrengthForKind(verify.KindDeterministicCheck),
				Notes:    "a targeted invariant is no longer targetable (weaken/falsified); evaluation requires reassessment against a fresh context (H5)",
			}
		} else {
			d, derr := verify.Route(ctx, verifiers, vc)
			if derr != nil {
				a.failRun(ctx, repoStore, run.ID, derr)
				return EvaluateResponse{}, derr
			}
			decision = d
		}
		row := store.EvaluationRow{
			ID:                   domain.NewEvaluationID(now),
			ProposalID:           p.ID,
			Verdict:              string(decision.Verdict),
			VerifierKind:         string(decision.Kind),
			VerificationStrength: string(decision.Strength),
			ConfidenceOrdinal:    decision.ConfidenceOrdinal,
			Notes:                decision.Notes,
		}
		if decision.Kind == verify.KindModelJudgment {
			row.Invocation = a.evaluationInvocation(model, run.ID, now)
		} else {
			row.ToolName = "newf-" + string(decision.Kind)
			row.ToolVersion = provider.VerifierVersion
		}
		record.Evaluations = append(record.Evaluations, row)
	}

	persisted, err := repoStore.PersistEvaluationRun(ctx, record)
	if err != nil {
		a.failRun(ctx, repoStore, run.ID, err)
		return EvaluateResponse{}, err
	}
	if err := a.finalizeRun(ctx, repoStore, run.ID, nil); err != nil {
		return EvaluateResponse{}, err
	}

	return EvaluateResponse{
		OK:      true,
		Command: "evaluate",
		Store:   dbPath,
		Run:     evaluationRunView(persisted),
	}, nil
}

// modelVerifier resolves the model-tier verifier ONCE per evaluation pass.
// Routing and provenance retention must share the SAME instance: constructing
// a second fixture for invocation recording returns empty LastPayloads() and
// silently discards the executed request/response (F3).
func (a *App) modelVerifier() provider.ModelVerifier {
	if a.modelVerifierFn != nil {
		return a.modelVerifierFn
	}
	// A conservative default model tier: it abstains (unknown) rather than
	// manufacturing a verdict, so a bare deployment never launders judgment.
	return provider.NewFixtureModelVerifier(verify.VerdictUnknown, "low")
}

// verifiers returns the routing verifier set: the two deterministic tiers plus
// the model tier last. The model tier is a fixture in CI; a live adapter would
// replace modelVerifierFn.
func (a *App) verifiers(model provider.ModelVerifier) []verify.Verifier {
	return []verify.Verifier{verify.DeterministicCheck{}, verify.CounterexampleSearch{}, model}
}

// evaluationInvocation builds the model-tier provider invocation for provenance
// retention, pulling the payloads recorded by the EXECUTED verifier instance
// and hashing the retained request for replay/audit.
func (a *App) evaluationInvocation(model provider.ModelVerifier, runID string, now time.Time) *store.EvaluationProviderInvocation {
	meta := model.Metadata()
	req, resp := model.LastPayloads()
	reqHash := ""
	if req != "" {
		sum := sha256.Sum256([]byte(req))
		reqHash = hex.EncodeToString(sum[:])
	}
	return &store.EvaluationProviderInvocation{
		ID:              domain.NewProviderInvocationID(now),
		RunID:           runID,
		ProviderName:    meta.ProviderName,
		ProviderVersion: meta.ProviderVersion,
		ModelName:       meta.ModelName,
		SchemaVersion:   meta.SchemaVersion,
		RequestHash:     reqHash,
		RequestPayload:  req,
		ResponsePayload: resp,
		CreatedAt:       now.Format(timeLayout),
	}
}

// targetPredicates loads the parsed predicate for every surviving invariant of
// the problem, keyed by invariant id, for counterexample re-checks.
func (a *App) targetPredicates(ctx context.Context, repoStore problemStore, problemID string) (map[string]invariant.Predicate, error) {
	survivors, err := a.survivingInvariants(ctx, repoStore, problemID)
	if err != nil {
		return nil, err
	}
	out := make(map[string]invariant.Predicate, len(survivors))
	for _, s := range survivors {
		out[s.InvariantID] = s.Predicate
	}
	return out, nil
}

// clusterRepresentatives loads the representative signature (with full epistemic
// provenance) of every cluster in the generation's cluster run, keyed by cluster
// id. Per-proposal nearest-family re-checks (G1) index into this map using the
// proposal's OWN recorded nearest clusters, rather than scanning every
// representative for every target.
func (a *App) clusterRepresentatives(ctx context.Context, repoStore problemStore, clusterRunID string) (map[string]canon.MechanismSignature, error) {
	out := map[string]canon.MechanismSignature{}
	if clusterRunID == "" {
		return out, nil
	}
	clusterRun, err := repoStore.GetClusterRun(ctx, clusterRunID)
	if err != nil {
		return nil, err
	}
	families, err := frontierFamilies(ctx, repoStore, clusterRun)
	if err != nil {
		return nil, err
	}
	for _, fam := range families {
		out[fam.ClusterID] = fam.Representative
	}
	return out, nil
}

// verificationContextForProposal assembles the code-owned context for one
// proposal and reports whether any of its claimed targets is STALE (no longer
// currently targetable — it became weaken/falsified since generation).
//
//   - Target pinning (G4): a cached per-target verdict is included only when its
//     target is still in `predicates`. A stale target is NOT silently dropped so
//     the remainder can seek a favorable fallback (H5): the caller treats a
//     proposal with any stale target as requiring reassessment and does not route
//     it, because erasing a target changes the question being evaluated and an
//     unchanged proposal must not gain a better verdict merely because a
//     hypothesis it targeted lost credibility.
//   - Recorded nearest families (G1): comparison verdicts come from re-evaluating
//     the live target predicate against the proposal's OWN recorded nearest
//     cluster representatives, not every representative in the run.
func verificationContextForProposal(p store.FrontierProposalRow, predicates map[string]invariant.Predicate, reps map[string]canon.MechanismSignature) (verify.VerificationContext, bool) {
	targets := make(map[string]invariant.Verdict)
	nearest := make(map[string][]invariant.Verdict)
	staleTarget := false
	for _, t := range p.Targets {
		pred, live := predicates[t.InvariantID]
		if !live {
			staleTarget = true
			continue
		}
		targets[t.InvariantID] = invariant.Verdict(t.Verdict)
		for _, nc := range p.NearestClusters {
			rep, ok := reps[nc.ClusterID]
			if !ok {
				continue
			}
			nearest[t.InvariantID] = append(nearest[t.InvariantID], invariant.Evaluate(pred, rep))
		}
	}
	return verify.VerificationContext{
		ProposalID:       p.ID,
		TargetVerdicts:   targets,
		NearestVerdicts:  nearest,
		ClaimedViolation: p.StructuralViolationClaim,
	}, staleTarget
}

// ListEvaluations lists a problem's evaluation runs.
func (a *App) ListEvaluations(ctx context.Context, input EvaluationListInput) (EvaluationListResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return EvaluationListResponse{}, err
	}
	defer repoStore.Close()

	recs, err := repoStore.ListEvaluations(ctx, input.ProblemID)
	if err != nil {
		return EvaluationListResponse{}, err
	}
	resp := EvaluationListResponse{OK: true, Command: "evaluation list", Store: dbPath}
	for _, r := range recs {
		resp.Runs = append(resp.Runs, evaluationRunView(r))
	}
	return resp, nil
}

// ShowEvaluation loads one evaluation run.
func (a *App) ShowEvaluation(ctx context.Context, input EvaluationShowInput) (EvaluationShowResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return EvaluationShowResponse{}, err
	}
	defer repoStore.Close()

	rec, err := repoStore.GetEvaluationRun(ctx, input.EvaluationID)
	if err != nil {
		return EvaluationShowResponse{}, err
	}
	return EvaluationShowResponse{
		OK:      true,
		Command: "evaluation show",
		Store:   dbPath,
		Run:     evaluationRunView(rec),
	}, nil
}

// ListEvaluatedFailures returns the proposals a prior evaluation marked as
// re-entered failures for a problem (R6/KTD-5). This is the queryable
// eligibility path a subsequent `cluster build` (or an M6.2 search policy) can
// consult to include newly-evaluated failure structure in the next pass:
// FailureSpace(t+1) ⊇ FailureSpace(t) + these failures. It does NOT trigger any
// re-clustering; re-entry is an explicit operator/policy decision.
func (a *App) ListEvaluatedFailures(ctx context.Context, input EvaluationListInput) (EvaluatedFailureListResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return EvaluatedFailureListResponse{}, err
	}
	defer repoStore.Close()

	rows, err := repoStore.ListEvaluatedFailures(ctx, input.ProblemID)
	if err != nil {
		return EvaluatedFailureListResponse{}, err
	}
	resp := EvaluatedFailureListResponse{OK: true, Command: "evaluation failures", Store: dbPath}
	for _, r := range rows {
		resp.Failures = append(resp.Failures, EvaluatedFailureView{
			ProposalID:   r.ProposalID,
			EvaluationID: r.EvaluationID,
			Verdict:      r.Verdict,
			CreatedAt:    r.CreatedAt,
		})
	}
	return resp, nil
}

// evaluationRunView maps a persisted evaluation run to its stable view.
func evaluationRunView(rec store.EvaluationRunRecord) EvaluationRunView {
	view := EvaluationRunView{
		ID:                      rec.ID,
		ProblemID:               rec.ProblemID,
		RunID:                   rec.RunID,
		FrontierGenerationRunID: rec.FrontierGenerationRunID,
		ClusterRunID:            rec.ClusterRunID,
		Mode:                    rec.Mode,
		RoutingPolicy:           rec.RoutingPolicy,
		EvaluationCount:         rec.EvaluationCount,
		CreatedAt:               rec.CreatedAt,
	}
	for _, e := range rec.Evaluations {
		ev := EvaluationView{
			ID:                   e.ID,
			ProposalID:           e.ProposalID,
			Verdict:              e.Verdict,
			VerifierKind:         e.VerifierKind,
			VerificationStrength: e.VerificationStrength,
			ConfidenceOrdinal:    e.ConfidenceOrdinal,
			ToolName:             e.ToolName,
			ToolVersion:          e.ToolVersion,
			ProviderInvocationID: e.ProviderInvocationID,
			Notes:                e.Notes,
		}
		for _, m := range e.Metrics {
			mv := EvaluationMetricView{Name: m.MetricName, Scale: m.MetricScale}
			switch m.MetricScale {
			case "numeric":
				if m.NumericValue.Valid {
					mv.Value = fmt.Sprintf("%v", m.NumericValue.Float64)
				}
			case "ordinal":
				mv.Value = m.OrdinalValue
			case "categorical":
				mv.Value = m.CategoricalValue
			}
			ev.Metrics = append(ev.Metrics, mv)
		}
		view.Evaluations = append(view.Evaluations, ev)
	}
	sort.Slice(view.Evaluations, func(i, j int) bool { return view.Evaluations[i].ID < view.Evaluations[j].ID })
	return view
}
