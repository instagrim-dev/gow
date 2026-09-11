package pipeline

import (
	"context"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/invariant"
	"github.com/instagrim-dev/newf/internal/store"
)

// TestIntegrationEvaluateEndToEnd runs the full offline M5.2 path on top of the
// M5.1 substrate: seed -> cluster -> failure-space -> mine -> challenge
// (-> surviving) -> frontier generate -> evaluate -> evaluation show. It asserts
// the evaluation run completes, the proposal's result is populated, and every
// evaluation carries BOTH a verdict AND its verification strength (R1).
//
// The seeded proposal VIOLATES its target (`preserves contains <id>`) while the
// nearest known failure families still SATISFY it — the intended structural
// difference, not a refutation (G1). The bounded counterexample search therefore
// finds no refuter and is NON-DECISIVE; realizability is a model judgment, so
// the proposal is honestly routed to the model tier rather than being awarded a
// deterministic `partial_success` for missing comparison evidence.
func TestIntegrationEvaluateEndToEnd(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = dataDrivenMiner{}
	app.challengerFn = biasOnlyChallenger{}

	problemID, invID, _ := mineOneCandidate(t, ctx, app, dbPath)

	if resp, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID}); err != nil {
		t.Fatalf("challenge: %v", err)
	} else if resp.Reports[0].StateAfter != "surviving" {
		t.Fatalf("state after bias-only campaign = %q, want surviving", resp.Reports[0].StateAfter)
	}

	gen, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("frontier generate: %v", err)
	}
	if len(gen.Generation.Proposals) == 0 {
		t.Fatal("expected at least one proposal to evaluate")
	}
	proposalID := gen.Generation.Proposals[0].ID

	// Evaluate the single proposal cheap-first.
	res, err := app.Evaluate(ctx, EvaluateInput{DBPath: dbPath, ProblemID: problemID, ProposalID: proposalID})
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if res.Run.Mode != "proposal" || res.Run.RoutingPolicy != "cheap-first" {
		t.Fatalf("unexpected run metadata: mode=%q policy=%q", res.Run.Mode, res.Run.RoutingPolicy)
	}
	if len(res.Run.Evaluations) != 1 {
		t.Fatalf("want 1 evaluation, got %d", len(res.Run.Evaluations))
	}
	ev := res.Run.Evaluations[0]
	// R1: an outcome is never seen without its strength.
	if ev.Verdict == "" || ev.VerificationStrength == "" || ev.VerifierKind == "" {
		t.Fatalf("evaluation missing verdict/kind/strength: %+v", ev)
	}
	// G1: a confirmed break whose nearest known failures merely SATISFY the target
	// (the intended structural difference) must NOT be awarded a deterministic
	// partial_success for missing comparison evidence. The two deterministic tiers
	// abstain and the decision is the model tier's — recorded honestly as a
	// single-model judgment (here the conservative fixture abstains -> blocked).
	if ev.VerifierKind == "counterexample-search" && ev.Verdict == "partial_success" {
		t.Fatalf("bounded search must not award partial_success without a reproducing family (G1); got %+v", ev)
	}
	if ev.VerifierKind == "model-judgment" && ev.VerificationStrength != "single-model-judgment" {
		t.Fatalf("a model-tier decision must record single-model-judgment strength; got %q", ev.VerificationStrength)
	}

	// Run lifecycle reflects success.
	run, err := app.ShowRun(ctx, LookupInput{DBPath: dbPath, ID: res.Run.RunID})
	if err != nil {
		t.Fatalf("run show: %v", err)
	}
	if run.Run.Status != "completed" {
		t.Fatalf("evaluate run status = %q, want completed", run.Run.Status)
	}

	// R5: the proposal result is now populated and mirrors the verdict.
	shown, err := app.ShowFrontier(ctx, FrontierShowInput{DBPath: dbPath, GenerationID: gen.Generation.ID})
	if err != nil {
		t.Fatalf("frontier show: %v", err)
	}
	var result string
	for _, p := range shown.Generation.Proposals {
		if p.ID == proposalID {
			result = p.Result
		}
	}
	if result != ev.Verdict {
		t.Fatalf("proposal result %q does not mirror evaluation verdict %q (R5)", result, ev.Verdict)
	}

	// evaluation show round-trips the strength-stamped verdict.
	es, err := app.ShowEvaluation(ctx, EvaluationShowInput{DBPath: dbPath, EvaluationID: res.Run.ID})
	if err != nil {
		t.Fatalf("evaluation show: %v", err)
	}
	if len(es.Run.Evaluations) != 1 || es.Run.Evaluations[0].VerificationStrength != ev.VerificationStrength {
		t.Fatalf("evaluation show lost the strength stamp: %+v", es.Run.Evaluations)
	}

	// R6: if the verdict was a failure/partial_failure it re-enters the atlas.
	if ev.Verdict == "failure" || ev.Verdict == "partial_failure" {
		fails, ferr := app.ListEvaluatedFailures(ctx, EvaluationListInput{DBPath: dbPath, ProblemID: problemID})
		if ferr != nil {
			t.Fatalf("list evaluated failures: %v", ferr)
		}
		found := false
		for _, f := range fails.Failures {
			if f.ProposalID == proposalID {
				found = true
			}
		}
		if !found {
			t.Fatal("a failed proposal must re-enter the failure atlas (R6)")
		}
	}

	// evaluation list surfaces the run.
	list, err := app.ListEvaluations(ctx, EvaluationListInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("evaluation list: %v", err)
	}
	if len(list.Runs) != 1 || list.Runs[0].ID != res.Run.ID {
		t.Fatalf("expected the evaluation run on the list surface, got %+v", list.Runs)
	}
}

// TestIntegrationEvaluateReachesDedupedProposal is the finding-2 regression:
// generate a frontier twice with an unchanged substrate so the second
// generation FULLY dedups (it owns zero proposal rows). A by-id evaluate of an
// original proposal must still resolve it through its OWNING generation — not
// the latest, which is empty — and must not fabricate another copy. A batch
// evaluate must likewise reach the un-evaluated originals rather than going
// blind on the empty latest generation.
func TestIntegrationEvaluateReachesDedupedProposal(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = dataDrivenMiner{}
	app.challengerFn = biasOnlyChallenger{}

	problemID, invID, _ := mineOneCandidate(t, ctx, app, dbPath)
	if resp, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID}); err != nil {
		t.Fatalf("challenge: %v", err)
	} else if resp.Reports[0].StateAfter != "surviving" {
		t.Fatalf("state after bias-only campaign = %q, want surviving", resp.Reports[0].StateAfter)
	}

	gen1, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("first generate: %v", err)
	}
	if len(gen1.Generation.Proposals) == 0 {
		t.Fatal("first generation must own proposals")
	}
	originalID := gen1.Generation.Proposals[0].ID

	// Second generation over the unchanged substrate: deterministic candidates
	// dedup onto the existing rows, so this generation owns ZERO proposal rows.
	gen2, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("second generate: %v", err)
	}
	if len(gen2.Generation.Proposals) != 0 {
		t.Fatalf("second generation should fully dedup (own zero rows), got %d", len(gen2.Generation.Proposals))
	}

	// By-id evaluate must reach the original through its owning generation even
	// though the LATEST generation (gen2) owns nothing.
	res, err := app.Evaluate(ctx, EvaluateInput{DBPath: dbPath, ProblemID: problemID, ProposalID: originalID})
	if err != nil {
		t.Fatalf("by-id evaluate of a deduped-away proposal must still resolve it: %v", err)
	}
	if len(res.Run.Evaluations) != 1 {
		t.Fatalf("want exactly one evaluation (no duplicate proposal created), got %d", len(res.Run.Evaluations))
	}
	if res.Run.FrontierGenerationRunID != gen1.Generation.ID {
		t.Fatalf("by-id evaluate must run in the OWNING generation %s, got %s", gen1.Generation.ID, res.Run.FrontierGenerationRunID)
	}

	// The proposal count is unchanged: no copy was fabricated.
	repo, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer repo.Close()
	owning, err := repo.GetFrontierGeneration(ctx, gen1.Generation.ID)
	if err != nil {
		t.Fatalf("get owning gen: %v", err)
	}
	if len(owning.Proposals) != len(gen1.Generation.Proposals) {
		t.Fatalf("owning generation proposal count changed: was %d now %d", len(gen1.Generation.Proposals), len(owning.Proposals))
	}
}

// TestIntegrationEvaluateHoldoutRefused proves the holdout mode is refused at the
// service boundary with a deferred-to-M7 error (R9), not half-built.
func TestIntegrationEvaluateHoldoutRefused(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)

	_, err := app.Evaluate(ctx, EvaluateInput{DBPath: dbPath, ProblemID: "prb_x", Mode: "holdout"})
	if err == nil {
		t.Fatal("holdout mode must be refused")
	}
	if got := err.Error(); got == "" || !contains(got, "M7") {
		t.Fatalf("holdout refusal should mention M7 deferral; got %q", got)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// TestVerificationContextDropsStaleTarget is the G4 regression: a proposal's
// cached target verdict must NOT survive its target becoming stale. When a
// targeted invariant is no longer currently targetable (it became
// weaken/falsified since generation, so it is absent from `predicates`), its
// cached `violates` verdict AND its comparison evidence are dropped together —
// an unchanged proposal cannot gain a better evaluation merely because the
// hypothesis it targeted became less credible. Here the ONLY target is stale, so
// the resulting context is empty and routes to a non-decisive result rather than
// a free partial_success.
func TestVerificationContextDropsStaleTarget(t *testing.T) {
	p := store.FrontierProposalRow{
		ID: "prop_1",
		Targets: []store.FrontierTargetRow{
			{InvariantID: "inv_live", Verdict: "violates"},
			{InvariantID: "inv_stale", Verdict: "violates"},
		},
		NearestClusters: []store.FrontierNearestRow{{ClusterID: "clu_1"}},
	}
	// inv_stale is absent from predicates (it became weaken/falsified). inv_live
	// remains, with a predicate no representative satisfies -> no false refuter.
	livePred := invariant.Predicate{Schema: invariant.PredicateSchemaV1, Root: invariant.Node{
		Op: invariant.OpContains, Field: invariant.FieldPreserves, CanonicalID: "core.operator.absent",
	}}
	predicates := map[string]invariant.Predicate{"inv_live": livePred}
	reps := map[string]canon.MechanismSignature{"clu_1": {
		SchemaVersion: canon.SchemaMechanismV1, VocabularyVersion: "mechanism/v1",
		OutcomeClass: domain.OutcomeFailure, Preserves: []canon.FieldClaim{},
		SetFieldCompleteness: map[domain.FieldKind]domain.FieldCompleteness{domain.FieldPreserves: domain.CompletenessComplete},
	}}

	vc, staleTarget := verificationContextForProposal(p, predicates, reps, nil)
	if !staleTarget {
		t.Fatal("a proposal with a no-longer-targetable target must be flagged stale (H5)")
	}
	if _, ok := vc.TargetVerdicts["inv_stale"]; ok {
		t.Fatal("a stale target's cached verdict must be dropped (G4)")
	}
	if _, ok := vc.NearestVerdicts["inv_stale"]; ok {
		t.Fatal("a stale target's comparison evidence must be dropped with its verdict (G4)")
	}
	if _, ok := vc.TargetVerdicts["inv_live"]; !ok {
		t.Fatal("a live target must be retained")
	}
}

// F3 regression: the DEFAULT pipeline (no injected model verifier) must retain
// the executed model-tier request/response payloads and a request hash on the
// persisted provider invocation. Routing and provenance retention share ONE
// verifier instance; a second construction would return empty payloads.
func TestIntegrationEvaluateDefaultVerifierRetainsPayloads(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = dataDrivenMiner{}
	app.challengerFn = biasOnlyChallenger{}
	app.modelVerifierFn = nil // the seam under test: bare deployment default

	problemID, invID, _ := mineOneCandidate(t, ctx, app, dbPath)
	if _, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID}); err != nil {
		t.Fatalf("challenge: %v", err)
	}
	gen, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("frontier generate: %v", err)
	}
	res, err := app.Evaluate(ctx, EvaluateInput{DBPath: dbPath, ProblemID: problemID, ProposalID: gen.Generation.Proposals[0].ID})
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	ev := res.Run.Evaluations[0]
	if ev.VerifierKind != "model-judgment" {
		t.Skipf("deterministic tier decided (%s); model tier not exercised on this corpus", ev.VerifierKind)
	}

	repo := openTestStore(t, ctx, dbPath)
	defer repo.Close()
	var invocationID string
	full, err := repo.GetEvaluationRun(ctx, res.Run.ID)
	if err != nil {
		t.Fatalf("load run: %v", err)
	}
	for _, e := range full.Evaluations {
		if e.ProviderInvocationID != "" {
			invocationID = e.ProviderInvocationID
		}
	}
	if invocationID == "" {
		t.Fatal("model-tier evaluation persisted no provider invocation")
	}
	inv, err := repo.GetProviderInvocation(ctx, invocationID)
	if err != nil {
		t.Fatalf("load invocation: %v", err)
	}
	if inv.RequestPayload == "" || inv.ResponsePayload == "" {
		t.Fatalf("executed payloads lost: request=%q response=%q", inv.RequestPayload, inv.ResponsePayload)
	}
	if inv.RequestHash == "" {
		t.Fatal("request hash must be stored for replay/audit")
	}
}

// Round-2 F1 regression (unit): when the occurrence-bound signature revision is
// supplied, target verdicts are RECOMPUTED against those exact bytes — the
// cached generation-time verdict must not describe content the verifier never
// saw. Here the cached verdict says "violates" but the revised content
// SATISFIES the predicate.
func TestVerificationContextRecomputesAgainstRevision(t *testing.T) {
	pred := invariant.Predicate{
		Schema: invariant.PredicateSchemaV1,
		Root:   invariant.Node{Op: invariant.OpContains, Field: invariant.FieldPreserves, CanonicalID: "domain.number_theory.property.residue_locality"},
	}
	predicates := map[string]invariant.Predicate{"inv_x": pred}
	p := store.FrontierProposalRow{
		ID:      "fpr_x",
		Targets: []store.FrontierTargetRow{{InvariantID: "inv_x", Verdict: "violates", Violated: true}},
	}

	// Revised content that SATISFIES the predicate (complete + resolved claim).
	sig := canon.MechanismSignature{
		SchemaVersion:     canon.SchemaMechanismV1,
		VocabularyVersion: "mechanism/v1",
		Preserves: []canon.FieldClaim{{
			FieldKind: domain.FieldPreserves, State: domain.ResolutionResolved,
			CanonicalID: "domain.number_theory.property.residue_locality", Status: domain.ClaimExplicit,
		}},
		SetFieldCompleteness: map[domain.FieldKind]domain.FieldCompleteness{
			domain.FieldPreserves: domain.CompletenessComplete,
		},
	}
	vc, stale := verificationContextForProposal(p, predicates, map[string]canon.MechanismSignature{}, &sig)
	if stale {
		t.Fatal("live predicate must not be stale")
	}
	if got := vc.TargetVerdicts["inv_x"]; got != invariant.VerdictSatisfies {
		t.Fatalf("verdict must be recomputed against the revision (satisfies), got %q (cached was violates)", got)
	}

	// Without content (pre-v17 gap) the cached verdict is the honest fallback.
	vc2, _ := verificationContextForProposal(p, predicates, map[string]canon.MechanismSignature{}, nil)
	if got := vc2.TargetVerdicts["inv_x"]; got != invariant.VerdictViolates {
		t.Fatalf("no-content fallback must keep the cached verdict, got %q", got)
	}
}

// Round-2 F1 regression (integration): the persisted evaluation references the
// EXACT occurrence-bound revision it assessed — the supplied hash, not an
// independent latest-revision lookup at persistence time.
func TestIntegrationEvaluationStampsAssessedRevision(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = dataDrivenMiner{}
	app.challengerFn = biasOnlyChallenger{}

	problemID, invID, _ := mineOneCandidate(t, ctx, app, dbPath)
	if _, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID}); err != nil {
		t.Fatalf("challenge: %v", err)
	}
	gen, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	proposalID := gen.Generation.Proposals[0].ID
	res, err := app.Evaluate(ctx, EvaluateInput{DBPath: dbPath, ProblemID: problemID, ProposalID: proposalID})
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}

	repo := openTestStore(t, ctx, dbPath)
	defer repo.Close()
	occ, err := repo.ListGenerationOccurrenceContents(ctx, gen.Generation.ID)
	if err != nil {
		t.Fatalf("occurrences: %v", err)
	}
	wantHash := occ[proposalID].ContentHash
	if wantHash == "" {
		t.Fatal("generation must have an occurrence binding for its proposal")
	}
	full, err := repo.GetEvaluationRun(ctx, res.Run.ID)
	if err != nil {
		t.Fatalf("load run: %v", err)
	}
	if got := full.Evaluations[0].SignatureContentHash; got != wantHash {
		t.Fatalf("evaluation must reference the assessed occurrence revision %s, got %q", wantHash, got)
	}
}
