package pipeline

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/invariant"
	"github.com/instagrim-dev/newf/internal/provider"
	"github.com/instagrim-dev/newf/internal/verify"
)

// TestIntegrationCompressSuccessesEndToEnd runs the full offline M6.1 path on
// top of the whole chain: seed -> cluster -> failure-space -> mine ->
// challenge (-> surviving) -> frontier generate -> evaluate (model tier fixed
// to partial_success) -> successes compress -> success-invariant show.
//
// The deriving fixture generator's break-proposal preserves
// core.property.global_coupling, which is NOT in the pinned vocabulary — so
// the deriving compressor's derived condition is inadmissible. The honest
// outcome is an EMPTY revision with inadmissible_conditions = 1 and a
// completed run: the pipeline neither fabricates a success invariant nor
// bricks on an over-generating fixture. The revision still records the cohort
// facts (hash, counts) so the pass is reproducible and idempotent.
func TestIntegrationCompressSuccessesEndToEnd(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = dataDrivenMiner{}
	app.challengerFn = biasOnlyChallenger{}
	app.modelVerifierFn = provider.NewFixtureModelVerifier(verify.VerdictPartialSuccess, "medium")

	problemID, invID, _ := mineOneCandidate(t, ctx, app, dbPath)
	if resp, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID}); err != nil {
		t.Fatalf("challenge: %v", err)
	} else if resp.Reports[0].StateAfter != "surviving" {
		t.Fatalf("state after campaign = %q, want surviving", resp.Reports[0].StateAfter)
	}
	gen, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("frontier generate: %v", err)
	}
	if len(gen.Generation.Proposals) == 0 {
		t.Fatal("expected a proposal")
	}
	if _, err := app.Evaluate(ctx, EvaluateInput{DBPath: dbPath, ProblemID: problemID}); err != nil {
		t.Fatalf("evaluate: %v", err)
	}

	res, err := app.CompressSuccesses(ctx, SuccessCompressInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("compress: %v", err)
	}
	if !res.Created {
		t.Fatal("first compression must create a revision")
	}
	rev := res.Revision
	if rev.IneligibleUnpersisted != 0 {
		t.Fatalf("v17 generation persists content; ineligible = %d, want 0", rev.IneligibleUnpersisted)
	}
	if rev.InadmissibleConditions != 1 || rev.InvariantCount != 0 {
		t.Fatalf("expected the out-of-vocabulary derived condition to be counted inadmissible (1) with an empty revision, got inadmissible=%d invariants=%d", rev.InadmissibleConditions, rev.InvariantCount)
	}

	// Run lifecycle reflects success even for an honest empty outcome.
	run, err := app.ShowRun(ctx, LookupInput{DBPath: dbPath, ID: rev.RunID})
	if err != nil {
		t.Fatalf("run show: %v", err)
	}
	if run.Run.Status != "completed" {
		t.Fatalf("compress run status = %q, want completed", run.Run.Status)
	}

	// Idempotent: an unchanged cohort returns the existing revision.
	again, err := app.CompressSuccesses(ctx, SuccessCompressInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("re-compress: %v", err)
	}
	if again.Created || again.Revision.ID != rev.ID {
		t.Fatalf("re-compress must be idempotent; created=%v id=%s want %s", again.Created, again.Revision.ID, rev.ID)
	}

	// show + list round-trip.
	shown, err := app.ShowSuccess(ctx, SuccessShowInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("show: %v", err)
	}
	if shown.Revision.ID != rev.ID || shown.Revision.CohortHash != rev.CohortHash {
		t.Fatalf("show mismatch: %+v", shown.Revision)
	}
	listed, err := app.ListSuccesses(ctx, SuccessListInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(listed.Revisions) != 1 {
		t.Fatalf("expected 1 revision, got %d", len(listed.Revisions))
	}

	// Sidecar round-trip guard: the persisted canonical content re-fingerprints
	// to the stored fingerprint (falsify by corrupting the JSON and watching
	// this fail) — the content is genuinely evaluable, not decorative.
	repo := openTestStore(t, ctx, dbPath)
	defer repo.Close()
	rows, err := repo.ListBreakCohortRows(ctx, problemID)
	if err != nil {
		t.Fatalf("cohort rows: %v", err)
	}
	if len(rows) == 0 {
		t.Fatal("expected at least one break-cohort row")
	}
	for _, r := range rows {
		var sig canon.MechanismSignature
		if err := json.Unmarshal([]byte(r.SignatureJSON), &sig); err != nil {
			t.Fatalf("persisted signature must unmarshal: %v", err)
		}
		if got := canon.Fingerprint(sig); got != r.Fingerprint {
			t.Fatalf("persisted content re-fingerprints to %s, stored %s", got, r.Fingerprint)
		}
	}
}

// vocabCompressor proposes a condition over an IN-VOCABULARY id, so the full
// admit -> evaluate -> persist path is exercised with a real success invariant.
type vocabCompressor struct{ id string }

func containsPreservesPredicate(id string) invariant.Predicate {
	return invariant.Predicate{
		Schema: invariant.PredicateSchemaV1,
		Root:   invariant.Node{Op: invariant.OpContains, Field: invariant.FieldPreserves, CanonicalID: id},
	}
}

func (c vocabCompressor) Compress(_ context.Context, req provider.CompressionRequest) (provider.CompressionResponse, error) {
	var proposals []provider.ConditionProposal
	for _, cohort := range req.Cohorts {
		proposals = append(proposals, provider.ConditionProposal{
			TargetInvariantID: cohort.TargetInvariantID,
			Predicate:         containsPreservesPredicate(c.id),
			Statement:         "progress retains " + c.id,
			AbstractionLevel:  "mechanism",
		})
	}
	return provider.CompressionResponse{
		Proposals:       proposals,
		Metadata:        provider.Metadata{ProviderName: "fixture", ProviderVersion: "test", ModelName: "vocab-compressor", SchemaVersion: "invariant-predicate/v1"},
		RequestPayload:  "req",
		ResponsePayload: "resp",
	}, nil
}

// TestIntegrationCompressPersistsDiscriminatingCondition drives the same chain
// but with a compressor proposing an in-vocabulary condition. The break
// proposal's signature does NOT preserve residue_locality (its preserves set is
// exhaustively {global_coupling}), so the condition is VIOLATED by the single
// progressor: coverage 0/1 — persisted honestly with the strength composition
// empty (no supporting verdicts). Code computed it; the provider could not
// inflate it.
func TestIntegrationCompressPersistsDiscriminatingCondition(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = dataDrivenMiner{}
	app.challengerFn = biasOnlyChallenger{}
	app.modelVerifierFn = provider.NewFixtureModelVerifier(verify.VerdictPartialSuccess, "medium")
	app.successCompressorFn = vocabCompressor{id: "domain.number_theory.property.residue_locality"}

	problemID, invID, _ := mineOneCandidate(t, ctx, app, dbPath)
	if _, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID}); err != nil {
		t.Fatalf("challenge: %v", err)
	}
	if _, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID}); err != nil {
		t.Fatalf("generate: %v", err)
	}
	if _, err := app.Evaluate(ctx, EvaluateInput{DBPath: dbPath, ProblemID: problemID}); err != nil {
		t.Fatalf("evaluate: %v", err)
	}

	res, err := app.CompressSuccesses(ctx, SuccessCompressInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("compress: %v", err)
	}
	rev := res.Revision
	if rev.InvariantCount != 1 || rev.InadmissibleConditions != 0 {
		t.Fatalf("expected 1 admitted condition, got invariants=%d inadmissible=%d", rev.InvariantCount, rev.InadmissibleConditions)
	}
	si := rev.Invariants[0]
	if si.State != "proposed" {
		t.Fatalf("state = %q, want proposed", si.State)
	}
	if len(si.BrokenTargets) != 1 || si.BrokenTargets[0] != invID {
		t.Fatalf("broken targets = %v, want [%s]", si.BrokenTargets, invID)
	}
	// Code-computed discrimination: the progressor does NOT retain residue
	// locality (it preserves only global_coupling, exhaustively) -> violated.
	if si.CoverageNum != 0 || si.CoverageDen != 1 {
		t.Fatalf("coverage = %d/%d, want 0/1 (code refutes the claimed condition)", si.CoverageNum, si.CoverageDen)
	}
	if si.DistinctSupport != 0 {
		t.Fatalf("distinct support = %d, want 0", si.DistinctSupport)
	}
	total := si.Strength.Deterministic + si.Strength.Reproducible + si.Strength.IndependentEvidence + si.Strength.IndependentCritic + si.Strength.ModelJudgment
	if total != 0 {
		t.Fatalf("no supporting verdicts -> empty strength composition, got %+v", si.Strength)
	}
	if len(si.CohortEvaluations) != 1 || si.CohortEvaluations[0].Verdict != "violates" || si.CohortEvaluations[0].CohortRole != "progressor" {
		t.Fatalf("cohort evaluations = %+v", si.CohortEvaluations)
	}
	if si.PredicateFingerprint == "" || si.Predicate == "" {
		t.Fatal("expected persisted predicate + fingerprint")
	}
}
