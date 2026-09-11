package pipeline

import (
	"context"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/provider"
	"github.com/instagrim-dev/newf/internal/verify"
)

// TestIntegrationReassessRevisedOccurrence is the finding-2/finding-3
// regression (review of 87759d9): generate interpretation A (preserves
// COMPLETE), regenerate the SAME proposal with changed completeness B
// (unobserved — same fingerprint, same hash, dedups; the new generation binds
// B), then:
//
//  1. default by-id evaluation reaches the REVISED occurrence: the persisted
//     evaluation names B's content hash and its recomputed break verdict is
//     unknown (absence undecidable without accepted completeness);
//  2. generation-pinned replay of the ORIGINAL occurrence names A's hash with
//     the verified violates verdict;
//  3. success-cohort admission ranks CONTENT COMPATIBILITY before
//     decisiveness/strength (selection-policy/v3): B's completed
//     reassessment (break unknown) governs current guidance, and the
//     stronger, decisive, but STALE A-evaluation cannot override it —
//     the proposal is excluded from the break cohort, with A retained as
//     history/replay.
func TestIntegrationReassessRevisedOccurrence(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = minPreservesMiner{}
	app.challengerFn = biasOnlyChallenger{}
	app.generatorFn = pcGenerator{signatures: []canon.MechanismSignature{pcSignature(pcResidueLocality, "", true)}} // A: complete

	problemID, invID, _ := mineOneCandidate(t, ctx, app, dbPath)
	if resp, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID}); err != nil {
		t.Fatalf("challenge: %v", err)
	} else if resp.Reports[0].StateAfter != "surviving" {
		t.Fatalf("state = %q, want surviving", resp.Reports[0].StateAfter)
	}

	gen1, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("generate A: %v", err)
	}
	if len(gen1.Generation.Proposals) != 1 {
		t.Fatalf("want one owned proposal, got %d", len(gen1.Generation.Proposals))
	}
	proposalID := gen1.Generation.Proposals[0].ID

	// B: identical resolved content, completeness dropped -> same proposal
	// hash (fingerprint excludes completeness) -> dedup + revised binding.
	app.generatorFn = pcGenerator{signatures: []canon.MechanismSignature{pcSignature(pcResidueLocality, "", false)}}
	gen2, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("generate B: %v", err)
	}
	if len(gen2.Generation.Proposals) != 0 {
		t.Fatalf("B must fully dedup (own zero rows), got %d", len(gen2.Generation.Proposals))
	}

	repo := openTestStore(t, ctx, dbPath)
	defer repo.Close()
	occ1, err := repo.ListGenerationOccurrenceContents(ctx, gen1.Generation.ID)
	if err != nil {
		t.Fatalf("occurrences gen1: %v", err)
	}
	occ2, err := repo.ListGenerationOccurrenceContents(ctx, gen2.Generation.ID)
	if err != nil {
		t.Fatalf("occurrences gen2: %v", err)
	}
	hashA, hashB := occ1[proposalID].ContentHash, occ2[proposalID].ContentHash
	if hashA == "" || hashB == "" || hashA == hashB {
		t.Fatalf("A and B must be distinct persisted revisions: %q vs %q", hashA, hashB)
	}

	// (1) Default by-id evaluation reaches the REVISED occurrence.
	resB, err := app.Evaluate(ctx, EvaluateInput{DBPath: dbPath, ProblemID: problemID, ProposalID: proposalID})
	if err != nil {
		t.Fatalf("evaluate B: %v", err)
	}
	if resB.Run.FrontierGenerationRunID != gen2.Generation.ID {
		t.Fatalf("default by-id evaluation must run in the revised occurrence generation %s, got %s", gen2.Generation.ID, resB.Run.FrontierGenerationRunID)
	}
	runB, err := repo.GetEvaluationRun(ctx, resB.Run.ID)
	if err != nil {
		t.Fatalf("read back run B: %v", err)
	}
	eB := runB.Evaluations[0]
	if eB.SignatureContentHash != hashB {
		t.Fatalf("evaluation must name B's content hash %s, got %s", hashB, eB.SignatureContentHash)
	}
	if len(eB.TargetVerdicts) != 1 || eB.TargetVerdicts[0].InvariantID != invID ||
		eB.TargetVerdicts[0].Verdict != "unknown" || eB.TargetVerdicts[0].Violated {
		t.Fatalf("B's recomputed break verdict must be unknown (unobserved completeness): %+v", eB.TargetVerdicts)
	}

	// (2) Generation-pinned replay of the ORIGINAL occurrence, with a decisive
	// model tier so cohort selection below is deterministic.
	app.modelVerifierFn = provider.NewFixtureModelVerifier(verify.VerdictPartialSuccess, "medium")
	resA, err := app.Evaluate(ctx, EvaluateInput{DBPath: dbPath, ProblemID: problemID, ProposalID: proposalID, GenerationID: gen1.Generation.ID})
	if err != nil {
		t.Fatalf("replay A: %v", err)
	}
	if resA.Run.FrontierGenerationRunID != gen1.Generation.ID {
		t.Fatalf("pinned replay must run in %s, got %s", gen1.Generation.ID, resA.Run.FrontierGenerationRunID)
	}
	runA, err := repo.GetEvaluationRun(ctx, resA.Run.ID)
	if err != nil {
		t.Fatalf("read back run A: %v", err)
	}
	eA := runA.Evaluations[0]
	if eA.SignatureContentHash != hashA {
		t.Fatalf("replay must name A's content hash %s, got %s", hashA, eA.SignatureContentHash)
	}
	if len(eA.TargetVerdicts) != 1 || eA.TargetVerdicts[0].Verdict != "violates" || !eA.TargetVerdicts[0].Violated {
		t.Fatalf("A's recomputed break verdict must be violates: %+v", eA.TargetVerdicts)
	}

	// (3) Cohort admission under selection-policy/v3: content compatibility
	// ranks BEFORE decisiveness/strength. B is the current interpretation and
	// its completed reassessment (EB) says the break is unknown — so the
	// STRONGER, DECISIVE, but STALE A-evaluation must NOT override it and
	// resurrect cohort support. The proposal is excluded from the break
	// cohort entirely: stale evidence is history/replay, never current
	// guidance.
	rows, err := repo.ListBreakCohortRows(ctx, problemID)
	if err != nil {
		t.Fatalf("cohort rows: %v", err)
	}
	if len(rows) != 0 {
		t.Fatalf("stronger stale evidence must not override the current reassessment; want 0 cohort rows, got %+v", rows)
	}
}

// TestIntegrationReemittedOccurrenceRecompresses is the a47dd24 finding-2
// regression (A -> B -> A): re-emitting A makes A the CURRENT VIEW again even
// though B holds the higher retained revision number. Reassessing G3/A and
// recompressing must produce support from that assessment — the continued
// existence of revision B is history, not an instruction to use it.
func TestIntegrationReemittedOccurrenceRecompresses(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = minPreservesMiner{}
	app.challengerFn = biasOnlyChallenger{}
	sigA := pcSignature(pcResidueLocality, "", true)
	sigB := pcSignature(pcResidueLocality, "", false)

	problemID, invID, _ := mineOneCandidate(t, ctx, app, dbPath)
	if resp, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID}); err != nil || resp.Reports[0].StateAfter != "surviving" {
		t.Fatalf("challenge: %v / %+v", err, resp.Reports)
	}

	// G1 emits A, G2 emits B, G3 re-emits A (reuses revision 1's content).
	app.generatorFn = pcGenerator{signatures: []canon.MechanismSignature{sigA}}
	gen1, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("G1: %v", err)
	}
	proposalID := gen1.Generation.Proposals[0].ID
	app.generatorFn = pcGenerator{signatures: []canon.MechanismSignature{sigB}}
	if _, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID}); err != nil {
		t.Fatalf("G2: %v", err)
	}
	app.generatorFn = pcGenerator{signatures: []canon.MechanismSignature{sigA}}
	gen3, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("G3: %v", err)
	}

	repo := openTestStore(t, ctx, dbPath)
	defer repo.Close()
	occ1, err := repo.ListGenerationOccurrenceContents(ctx, gen1.Generation.ID)
	if err != nil {
		t.Fatalf("occ1: %v", err)
	}
	occ3, err := repo.ListGenerationOccurrenceContents(ctx, gen3.Generation.ID)
	if err != nil {
		t.Fatalf("occ3: %v", err)
	}
	hashA := occ1[proposalID].ContentHash
	if occ3[proposalID].ContentHash != hashA {
		t.Fatalf("G3 must reuse A's revision: %s vs %s", occ3[proposalID].ContentHash, hashA)
	}

	// Reassess (by-id default = latest occurrence = G3/A) with a decisive tier.
	app.modelVerifierFn = provider.NewFixtureModelVerifier(verify.VerdictPartialSuccess, "medium")
	res, err := app.Evaluate(ctx, EvaluateInput{DBPath: dbPath, ProblemID: problemID, ProposalID: proposalID})
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if res.Run.FrontierGenerationRunID != gen3.Generation.ID {
		t.Fatalf("assessment context must be G3, got %s", res.Run.FrontierGenerationRunID)
	}
	evalRun, err := repo.GetEvaluationRun(ctx, res.Run.ID)
	if err != nil {
		t.Fatalf("read run: %v", err)
	}
	if evalRun.Evaluations[0].SignatureContentHash != hashA {
		t.Fatalf("assessment must name A's bytes, got %s", evalRun.Evaluations[0].SignatureContentHash)
	}

	// The cohort's current view is the latest EMITTED occurrence (A), so the
	// completed G3/A assessment is current-compatible — not pending.
	rows, err := repo.ListBreakCohortRows(ctx, problemID)
	if err != nil {
		t.Fatalf("cohort: %v", err)
	}
	if len(rows) != 1 || rows[0].ContentHash != hashA || rows[0].LatestContentHash != hashA {
		t.Fatalf("current view must be the re-emitted A occurrence: %+v", rows)
	}

	// Recompression consumes that selection: support, not pending.
	comp, err := app.CompressSuccesses(ctx, SuccessCompressInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("compress: %v", err)
	}
	if comp.Revision.PendingReassessment != 0 {
		t.Fatalf("an assessed re-emitted occurrence must not be pending: %+v", comp.Revision)
	}
	if comp.Revision.InvariantCount == 0 {
		t.Fatalf("the assessed current occurrence must yield progressor support: %+v", comp.Revision)
	}
}

// TestIntegrationGenerationBatchReachesRevisedContent is the a47dd24
// finding-3 regression: after A is assessed (artifact result populated),
// emitting B and batch-evaluating B's GENERATION (no proposal id) must still
// assess the never-assessed B occurrence — batch eligibility is
// content-scoped, not artifact-scoped.
func TestIntegrationGenerationBatchReachesRevisedContent(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = minPreservesMiner{}
	app.challengerFn = biasOnlyChallenger{}
	app.generatorFn = pcGenerator{signatures: []canon.MechanismSignature{pcSignature(pcResidueLocality, "", true)}}

	problemID, invID, _ := mineOneCandidate(t, ctx, app, dbPath)
	if resp, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID}); err != nil || resp.Reports[0].StateAfter != "surviving" {
		t.Fatalf("challenge: %v / %+v", err, resp.Reports)
	}
	gen1, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("G1: %v", err)
	}
	proposalID := gen1.Generation.Proposals[0].ID

	// Assess A -> the artifact result is populated.
	if _, err := app.Evaluate(ctx, EvaluateInput{DBPath: dbPath, ProblemID: problemID}); err != nil {
		t.Fatalf("evaluate A: %v", err)
	}

	// Emit B (revised interpretation; fully dedups onto the artifact row).
	app.generatorFn = pcGenerator{signatures: []canon.MechanismSignature{pcSignature(pcResidueLocality, "", false)}}
	gen2, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("G2: %v", err)
	}

	// Generation-pinned BATCH evaluation must reach the unassessed B.
	res, err := app.Evaluate(ctx, EvaluateInput{DBPath: dbPath, ProblemID: problemID, GenerationID: gen2.Generation.ID})
	if err != nil {
		t.Fatalf("batch evaluate G2: %v", err)
	}
	if len(res.Run.Evaluations) != 1 {
		t.Fatalf("the never-assessed B occurrence must be selected, got %d evaluations", len(res.Run.Evaluations))
	}
	repo := openTestStore(t, ctx, dbPath)
	defer repo.Close()
	occ2, err := repo.ListGenerationOccurrenceContents(ctx, gen2.Generation.ID)
	if err != nil {
		t.Fatalf("occ2: %v", err)
	}
	evalRun, err := repo.GetEvaluationRun(ctx, res.Run.ID)
	if err != nil {
		t.Fatalf("read run: %v", err)
	}
	if evalRun.Evaluations[0].SignatureContentHash != occ2[proposalID].ContentHash {
		t.Fatalf("batch must assess B's bytes, got %s", evalRun.Evaluations[0].SignatureContentHash)
	}

	// Idempotence of the content-scoped filter: re-running the same batch now
	// selects nothing (this interpretation is assessed).
	again, err := app.Evaluate(ctx, EvaluateInput{DBPath: dbPath, ProblemID: problemID, GenerationID: gen2.Generation.ID})
	if err != nil {
		t.Fatalf("re-batch: %v", err)
	}
	if len(again.Run.Evaluations) != 0 {
		t.Fatalf("an assessed occurrence must not be re-selected by batch, got %d", len(again.Run.Evaluations))
	}
}
