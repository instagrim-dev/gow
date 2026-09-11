package pipeline

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/store"
)

// seedExperimentSubstrate builds the full harness substrate: a TRAIN problem
// carried through mine -> challenge (surviving), and a QUARANTINED target
// problem with its own corpus-derived canonical signatures.
func seedExperimentSubstrate(t *testing.T, ctx context.Context, app *App, dbPath string) (trainProblem, targetProblem string) {
	t.Helper()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	trainProblem, invID, _ := mineOneCandidate(t, ctx, app, dbPath)
	if resp, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID}); err != nil {
		t.Fatalf("challenge: %v", err)
	} else if resp.Reports[0].StateAfter != "surviving" {
		t.Fatalf("state after campaign = %q, want surviving", resp.Reports[0].StateAfter)
	}

	targetProblem, targetRun, targetSnap := seedOtherProblem(t, ctx, dbPath, now.Add(time.Hour))
	seedClusterCorpus(t, ctx, app, dbPath, targetProblem, targetRun, targetSnap, canon.VocabularyMechanismV1)
	return trainProblem, targetProblem
}

// TestIntegrationExperimentEndToEnd runs the M7 v0 harness: define (blinded)
// -> run (leakage audit passes, B0+B3 under equal budgets, one recovery rule)
// -> show, asserting the mode stamp, the audited blinding, honest arm facts,
// run lifecycle, and idempotent replay.
func TestIntegrationExperimentEndToEnd(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = dataDrivenMiner{}
	app.challengerFn = biasOnlyChallenger{}

	trainProblem, targetProblem := seedExperimentSubstrate(t, ctx, app, dbPath)

	def, err := app.DefineExperiment(ctx, ExperimentDefineInput{
		DBPath: dbPath, ProblemID: trainProblem, TargetProblemID: targetProblem,
	})
	if err != nil {
		t.Fatalf("define: %v", err)
	}
	if !def.Created || def.HoldoutSet.Mode != "blinded" || len(def.HoldoutSet.WithheldSources) == 0 {
		t.Fatalf("unexpected holdout set: %+v", def.HoldoutSet)
	}

	res, err := app.RunExperiment(ctx, ExperimentRunInput{DBPath: dbPath, ProblemID: trainProblem})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	exp := res.Experiment
	if !res.Created {
		t.Fatal("first run must create the experiment")
	}
	if exp.Mode != "blinded" || !strings.Contains(exp.ModeDisclaimer, "NO chronological claim") {
		t.Fatalf("mode stamp missing/wrong: mode=%q disclaimer=%q", exp.Mode, exp.ModeDisclaimer)
	}
	if !exp.LeakageCheck.Passed {
		t.Fatalf("clean split must pass the leakage audit: %+v", exp.LeakageCheck)
	}
	if len(exp.Arms) != 2 {
		t.Fatalf("expected 2 arms (b0, b3), got %+v", exp.Arms)
	}
	var b0, b3 *ExperimentArmView
	for i := range exp.Arms {
		switch exp.Arms[i].Arm {
		case "b0_undirected":
			b0 = &exp.Arms[i]
		case "b3_invariant_guided":
			b3 = &exp.Arms[i]
		}
	}
	if b0 == nil || b3 == nil {
		t.Fatalf("arms missing: %+v", exp.Arms)
	}
	// Offline B0 honestly yields zero proposals (a fixture cannot brainstorm).
	if b0.ProposalCount != 0 || b0.Recovered {
		t.Fatalf("b0 must be honest-empty offline: %+v", b0)
	}
	// B3 generated against the surviving invariant.
	if b3.ProposalCount == 0 || b3.FrontierGenerationRun == "" {
		t.Fatalf("b3 must generate proposals: %+v", b3)
	}
	// The corpus target's completeness is unobserved, so under the corrected
	// missing-data contract (classify/v2) the recorded-set difference from the
	// break proposal is NOT decisive: unrecorded members could overturn it.
	// The honest conclusion is inconclusive — computed by code, not asserted.
	// (The decisive-negative pipeline control lives in the positive-control
	// matrix, where the target JUSTIFIES its fields complete.)
	if exp.Conclusion != "inconclusive" {
		t.Fatalf("conclusion = %q, want inconclusive for the unobserved-completeness corpus target", exp.Conclusion)
	}
	// Metrics carry exact counts for every arm.
	if len(exp.Metrics) < 6 {
		t.Fatalf("expected >=6 metric rows, got %d", len(exp.Metrics))
	}

	run, err := app.ShowRun(ctx, LookupInput{DBPath: dbPath, ID: exp.RunID})
	if err != nil {
		t.Fatalf("run show: %v", err)
	}
	if run.Run.Status != "completed" {
		t.Fatalf("experiment run status = %q, want completed", run.Run.Status)
	}

	// Idempotent replay: unchanged substrate returns the existing experiment.
	again, err := app.RunExperiment(ctx, ExperimentRunInput{DBPath: dbPath, ProblemID: trainProblem})
	if err != nil {
		t.Fatalf("re-run: %v", err)
	}
	if again.Created || again.Experiment.ID != exp.ID {
		t.Fatalf("replay must be idempotent: created=%v id=%s want %s", again.Created, again.Experiment.ID, exp.ID)
	}

	shown, err := app.ShowExperiment(ctx, ExperimentShowInput{DBPath: dbPath, ProblemID: trainProblem})
	if err != nil {
		t.Fatalf("show: %v", err)
	}
	if shown.Experiment.ID != exp.ID || shown.Experiment.ModeDisclaimer == "" {
		t.Fatalf("show mismatch: %+v", shown.Experiment)
	}

	// F1: explicit arm<->proposal membership persisted with rank + assessment;
	// only membership determines what the arm is scored on.
	if len(b3.Members) != b3.ProposalCount {
		t.Fatalf("b3 members = %d, want %d (one membership per scored proposal)", len(b3.Members), b3.ProposalCount)
	}
	for i, m := range b3.Members {
		if m.ProposalID == "" || m.Assessment == "" {
			t.Fatalf("member %d incomplete: %+v", i, m)
		}
	}
	// F2: the frozen target manifest is persisted alongside the experiment.
	if len(exp.Targets) == 0 {
		t.Fatalf("experiment must persist its frozen target manifest")
	}
	// F3/F5: budget consumption is persisted; under the corrected contract the
	// unobserved-completeness corpus target yields unknown (not decisive_no),
	// so the inconclusive conclusion above pairs with a fully-consumed,
	// non-decisive assessment — nothing is left unassessed.
	if b3.EvaluationsConsumed == 0 {
		t.Fatalf("b3 consumed no evaluations yet concluded: %+v", b3)
	}
	if b3.UnknownCount != b3.ProposalCount || b3.DecisiveCount != 0 || b3.UnassessedCount != 0 {
		t.Fatalf("inconclusive here means assessed-but-unknown, never unassessed: %+v", b3)
	}

	// F2 regression: target material added AFTER definition (outside the
	// registered withheld sources) is invisible to scoring AND to experiment
	// identity — the replay stays idempotent with an unchanged manifest.
	addPostDefinitionTargetSource(t, ctx, app, dbPath, targetProblem)
	after, err := app.RunExperiment(ctx, ExperimentRunInput{DBPath: dbPath, ProblemID: trainProblem})
	if err != nil {
		t.Fatalf("re-run after target growth: %v", err)
	}
	if after.Created || after.Experiment.ID != exp.ID {
		t.Fatalf("post-definition target growth must not change the experiment: created=%v id=%s", after.Created, after.Experiment.ID)
	}
	if len(after.Experiment.Targets) != len(exp.Targets) {
		t.Fatalf("frozen manifest changed: %d -> %d targets", len(exp.Targets), len(after.Experiment.Targets))
	}

	// F3 regression: an enforced evaluation budget too small to finish yields
	// unassessed proposals and an INCONCLUSIVE conclusion (a different
	// experiment identity: the budget is part of the contract).
	starved, err := app.RunExperiment(ctx, ExperimentRunInput{DBPath: dbPath, ProblemID: trainProblem, EvaluationBudget: 1})
	if err != nil {
		t.Fatalf("starved run: %v", err)
	}
	if !starved.Created {
		t.Fatal("a different evaluation budget is a different experiment")
	}
	var sb3 *ExperimentArmView
	for i := range starved.Experiment.Arms {
		if starved.Experiment.Arms[i].Arm == "b3_invariant_guided" {
			sb3 = &starved.Experiment.Arms[i]
		}
	}
	if sb3 == nil || sb3.ProposalCount == 0 {
		t.Fatalf("starved b3 needs proposals to exercise the budget: %+v", sb3)
	}
	if sb3.EvaluationsConsumed > 1 {
		t.Fatalf("evaluation budget not enforced: consumed=%d budget=1", sb3.EvaluationsConsumed)
	}
	if sb3.UnassessedCount == 0 {
		t.Fatalf("starved run must leave proposals unassessed: %+v", sb3)
	}
	if starved.Experiment.Conclusion != "inconclusive" {
		t.Fatalf("starved conclusion = %q, want inconclusive (unassessed proposals are an epistemic gap)", starved.Experiment.Conclusion)
	}
}

// addPostDefinitionTargetSource grows the target problem with a NEW source +
// canonical signatures after the holdout definition froze the withheld set.
func addPostDefinitionTargetSource(t *testing.T, ctx context.Context, app *App, dbPath, targetProblem string) {
	t.Helper()
	later := time.Date(2026, 9, 10, 15, 0, 0, 0, time.UTC)
	repo := openTestStore(t, ctx, dbPath)
	defer repo.Close()
	runID := domain.NewRunID(later)
	if _, err := repo.CreateRun(ctx, domain.NewRun{
		ID: runID, ProblemID: targetProblem, Operation: "ingest", Status: domain.RunStatusInitialized,
		InputRef: "late-target", ToolName: "newf", ToolVersion: "test", StartedAt: later, CompletedAt: later,
	}); err != nil {
		t.Fatalf("late run: %v", err)
	}
	admission, err := repo.CreateSourceSnapshot(ctx, store.SnapshotAdmission{
		ProblemID: targetProblem, Kind: domain.SourceKindLocalPath,
		LogicalName: "late-target.md", Origin: "/tmp/late-target.md", SHA256: "deadd00d",
		ByteLength: 9, MediaType: "text/markdown", ObjectPath: "sha256/de/deadd00d",
		IngestRunID: runID, ObservedAt: later,
	})
	if err != nil {
		t.Fatalf("late snapshot: %v", err)
	}
	seed, err := app.SeedMechanismFixture(ctx, MechanismFixtureSeedInput{
		DBPath: dbPath, ProblemID: targetProblem, RunID: runID,
		SnapshotID: admission.Snapshot.ID, Path: fixturePath("case4_novel_candidate.json"),
	})
	if err != nil {
		t.Fatalf("late seed: %v", err)
	}
	for _, mechID := range seed.MechanismIDs {
		if _, err := app.SignatureMechanism(ctx, SignatureInput{DBPath: dbPath, MechanismID: mechID, VocabVersion: canon.VocabularyMechanismV1}); err != nil {
			t.Fatalf("late signature: %v", err)
		}
	}
}

// TestIntegrationExperimentLeakageFails seeds withheld CONTENT into the train
// problem (same sha256 under a different source) and asserts the audit fails,
// the run fails, and no experiment is persisted — the blinding claim is void.
func TestIntegrationExperimentLeakageFails(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = dataDrivenMiner{}
	app.challengerFn = biasOnlyChallenger{}

	trainProblem, targetProblem := seedExperimentSubstrate(t, ctx, app, dbPath)
	if _, err := app.DefineExperiment(ctx, ExperimentDefineInput{DBPath: dbPath, ProblemID: trainProblem, TargetProblemID: targetProblem}); err != nil {
		t.Fatalf("define: %v", err)
	}

	// Leak: admit the withheld bytes (sha cafebabe) into the TRAIN problem.
	repo := openTestStore(t, ctx, dbPath)
	defer repo.Close()
	leakRun := domain.NewRunID(now.Add(2 * time.Hour))
	if _, err := repo.CreateRun(ctx, domain.NewRun{
		ID: leakRun, ProblemID: trainProblem, Operation: "ingest", Status: domain.RunStatusInitialized,
		InputRef: "leak", ToolName: "newf", ToolVersion: "test", StartedAt: now, CompletedAt: now,
	}); err != nil {
		t.Fatalf("leak run: %v", err)
	}
	if _, err := repo.CreateSourceSnapshot(ctx, store.SnapshotAdmission{
		ProblemID: trainProblem, Kind: domain.SourceKindLocalPath,
		LogicalName: "leaked.md", Origin: "/tmp/leaked.md", SHA256: "cafebabe",
		ByteLength: 8, MediaType: "text/markdown", ObjectPath: "sha256/ca/cafebabe",
		IngestRunID: leakRun, ObservedAt: now,
	}); err != nil {
		t.Fatalf("seed leak: %v", err)
	}

	_, err := app.RunExperiment(ctx, ExperimentRunInput{DBPath: dbPath, ProblemID: trainProblem})
	if err == nil || !strings.Contains(err.Error(), "leakage check FAILED") {
		t.Fatalf("expected leakage failure, got %v", err)
	}
	if _, serr := app.ShowExperiment(ctx, ExperimentShowInput{DBPath: dbPath, ProblemID: trainProblem}); serr == nil {
		t.Fatal("a failed audit must persist no experiment")
	}
}

// TestIntegrationHistoricalModeRefused: historical execution requires dated
// evidence for EVERY withheld source; the shipped corpus has none, so the run
// is refused with the missing evidence named.
func TestIntegrationHistoricalModeRefused(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = dataDrivenMiner{}
	app.challengerFn = biasOnlyChallenger{}

	trainProblem, targetProblem := seedExperimentSubstrate(t, ctx, app, dbPath)

	// Historical define requires a cutoff.
	if _, err := app.DefineExperiment(ctx, ExperimentDefineInput{
		DBPath: dbPath, ProblemID: trainProblem, TargetProblemID: targetProblem, Mode: "historical",
	}); err == nil || !strings.Contains(err.Error(), "cutoff") {
		t.Fatalf("historical define without cutoff must be refused, got %v", err)
	}
	def, err := app.DefineExperiment(ctx, ExperimentDefineInput{
		DBPath: dbPath, ProblemID: trainProblem, TargetProblemID: targetProblem,
		Mode: "historical", CutoffTime: "2020-01-01T00:00:00Z", Name: "hist",
	})
	if err != nil {
		t.Fatalf("historical define: %v", err)
	}
	_, err = app.RunExperiment(ctx, ExperimentRunInput{DBPath: dbPath, HoldoutSetID: def.HoldoutSet.ID})
	if err == nil || !strings.Contains(err.Error(), "historical mode refused") {
		t.Fatalf("historical run without dating evidence must be refused, got %v", err)
	}
}

// TestIntegrationExperimentBaselineArmsAndCompare runs all four arms (B0-B3),
// asserts the now-executable B1/B2 baselines produce proposals with their own
// provenance role, verifies the shared recovery rule classifies every arm, and
// exercises the within-experiment compare service (default b0 vs b3 plus an
// explicit b2-vs-b3 delta). Idempotent replay must not create a second
// experiment and must not mutate the research state.
func TestIntegrationExperimentBaselineArmsAndCompare(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = dataDrivenMiner{}
	app.challengerFn = biasOnlyChallenger{}

	trainProblem, targetProblem := seedExperimentSubstrate(t, ctx, app, dbPath)
	if _, err := app.DefineExperiment(ctx, ExperimentDefineInput{DBPath: dbPath, ProblemID: trainProblem, TargetProblemID: targetProblem}); err != nil {
		t.Fatalf("define: %v", err)
	}

	allArms := []string{"b0_undirected", "b1_semantic_summary", "b2_brainstorm", "b3_invariant_guided"}
	res, err := app.RunExperiment(ctx, ExperimentRunInput{DBPath: dbPath, ProblemID: trainProblem, Arms: allArms, ProposalBudget: 4})
	if err != nil {
		t.Fatalf("run all arms: %v", err)
	}
	if !res.Created {
		t.Fatal("first run must create the experiment")
	}
	if len(res.Experiment.Arms) != 4 {
		t.Fatalf("expected 4 arms, got %d: %+v", len(res.Experiment.Arms), res.Experiment.Arms)
	}
	arms := map[string]ExperimentArmView{}
	for _, a := range res.Experiment.Arms {
		arms[a.Arm] = a
	}
	// B1/B2 are now executable: their deriving fixtures produce proposals from
	// the family projection (no invariant targets), so they are NOT empty.
	if arms["b1_semantic_summary"].ProposalCount == 0 {
		t.Fatalf("b1 must generate restatement proposals: %+v", arms["b1_semantic_summary"])
	}
	if arms["b2_brainstorm"].ProposalCount == 0 {
		t.Fatalf("b2 must generate brainstorm proposals: %+v", arms["b2_brainstorm"])
	}
	// B2's proposals are canonically redundant (surface-only variation), so its
	// distinct-mechanism count must be 1 and redundancy must be > 0 when >1.
	if b2 := arms["b2_brainstorm"]; b2.ProposalCount > 1 {
		if b2.DistinctFamilyCount != 1 {
			t.Fatalf("b2 distinct mechanisms = %d, want 1 (surface-only variation): %+v", b2.DistinctFamilyCount, b2)
		}
		if b2.RedundantCount == 0 {
			t.Fatalf("b2 must record redundant proposals: %+v", b2)
		}
	}
	// Every arm carries a stopping condition from the AGENTS.md vocabulary.
	for name, a := range arms {
		switch a.StoppingCondition {
		case "completed", "budget_exhausted", "no_information_gain", "verification_blocked":
		default:
			t.Fatalf("arm %s has invalid stopping condition %q", name, a.StoppingCondition)
		}
	}

	// Leak-1 regression: arms share the train problem and frontier_proposals
	// dedups on (problem_id, proposal_hash), so two arms that derive the same
	// mechanism legitimately reference the SAME persisted proposal_id. Arm
	// isolation must therefore NOT come from proposal_id uniqueness; it must come
	// from each arm's member_rank being its OWN arm-local position (dense, unique
	// 0..n-1) — a rank is never reused from another arm's ordering.
	for name, a := range arms {
		if len(a.Members) != a.ProposalCount {
			t.Fatalf("arm %s: %d members != proposal_count %d", name, len(a.Members), a.ProposalCount)
		}
		seenRank := map[int]bool{}
		for _, m := range a.Members {
			if m.MemberRank < 0 || m.MemberRank >= len(a.Members) {
				t.Fatalf("arm %s member_rank %d out of arm-local range [0,%d)", name, m.MemberRank, len(a.Members))
			}
			if seenRank[m.MemberRank] {
				t.Fatalf("arm %s reuses member_rank %d (ranks must be arm-local and unique)", name, m.MemberRank)
			}
			seenRank[m.MemberRank] = true
		}
	}

	// Default compare (b0 vs b3): apples-to-apples within one experiment.
	cmp, err := app.CompareExperiment(ctx, ExperimentCompareInput{DBPath: dbPath, ProblemID: trainProblem})
	if err != nil {
		t.Fatalf("compare: %v", err)
	}
	if cmp.Comparison.BaselineArm.Arm != "b0_undirected" || cmp.Comparison.TreatmentArm.Arm != "b3_invariant_guided" {
		t.Fatalf("default compare arms wrong: %+v", cmp.Comparison)
	}
	if cmp.Comparison.ModeDisclaimer == "" || !strings.Contains(cmp.Comparison.Interpretation, "split") {
		t.Fatalf("compare must carry mode disclaimer + non-inflated interpretation: %+v", cmp.Comparison)
	}
	if len(cmp.Comparison.Metrics) == 0 {
		t.Fatal("compare must report metric deltas")
	}
	// Leak-2 regression: b0 is offline-empty (ProposalCount 0), so it was NOT
	// decisively assessed — its status must be "inconclusive", never coerced to a
	// recovery negative, and the delta must not read "baseline-only"/"neither"
	// (which would credit b0 with a decisive no_recovery it never earned).
	if cmp.Comparison.BaselineArm.ProposalCount == 0 {
		if cmp.Comparison.BaselineRecoveryStatus != "inconclusive" {
			t.Fatalf("empty baseline arm must be inconclusive, got %q", cmp.Comparison.BaselineRecoveryStatus)
		}
		switch cmp.Comparison.RecoveryDelta {
		case "baseline-only", "neither":
			t.Fatalf("inconclusive baseline must not be coerced into a negative delta: %q", cmp.Comparison.RecoveryDelta)
		}
	}
	for _, s := range []string{cmp.Comparison.BaselineRecoveryStatus, cmp.Comparison.TreatmentRecoveryStatus} {
		switch s {
		case "recovered", "no_recovery", "inconclusive":
		default:
			t.Fatalf("invalid recovery status %q", s)
		}
	}
	for _, d := range cmp.Comparison.Metrics {
		switch d.Direction {
		case "higher", "lower", "same", "incomparable":
		default:
			t.Fatalf("metric %s has invalid direction %q", d.Metric, d.Direction)
		}
		// F4: the recovery metric encodes OBSERVED detections; with an
		// inconclusive arm in the pair its direction must be incomparable —
		// 0 observed hits is not a demonstrated negative.
		if d.Metric == "held_out_family_recovery" &&
			(cmp.Comparison.BaselineRecoveryStatus == "inconclusive" || cmp.Comparison.TreatmentRecoveryStatus == "inconclusive") &&
			d.Direction != "incomparable" {
			t.Fatalf("recovery metric direction = %q with an inconclusive arm, want incomparable", d.Direction)
		}
	}

	// Explicit b2-vs-b3 compare pairs the same metric names across two arms.
	cmp2, err := app.CompareExperiment(ctx, ExperimentCompareInput{DBPath: dbPath, ProblemID: trainProblem, BaselineArm: "b2_brainstorm", TreatmentArm: "b3_invariant_guided"})
	if err != nil {
		t.Fatalf("compare b2 vs b3: %v", err)
	}
	if cmp2.Comparison.BaselineArm.Arm != "b2_brainstorm" {
		t.Fatalf("explicit baseline arm wrong: %+v", cmp2.Comparison)
	}

	// Same baseline == treatment is refused.
	if _, err := app.CompareExperiment(ctx, ExperimentCompareInput{DBPath: dbPath, ProblemID: trainProblem, BaselineArm: "b3_invariant_guided", TreatmentArm: "b3_invariant_guided"}); err == nil {
		t.Fatal("compare must refuse identical arms")
	}

	// Idempotent replay: same arms + budget returns the SAME experiment, and the
	// research state (surviving invariants) is not mutated by measurement.
	statesBefore, err := listSurvivingCount(ctx, dbPath, trainProblem)
	if err != nil {
		t.Fatalf("states before: %v", err)
	}
	again, err := app.RunExperiment(ctx, ExperimentRunInput{DBPath: dbPath, ProblemID: trainProblem, Arms: allArms, ProposalBudget: 4})
	if err != nil {
		t.Fatalf("replay: %v", err)
	}
	if again.Created || again.Experiment.ID != res.Experiment.ID {
		t.Fatalf("replay must be idempotent: created=%v id=%s want %s", again.Created, again.Experiment.ID, res.Experiment.ID)
	}
	statesAfter, err := listSurvivingCount(ctx, dbPath, trainProblem)
	if err != nil {
		t.Fatalf("states after: %v", err)
	}
	if statesBefore != statesAfter {
		t.Fatalf("experiment must not mutate invariant state: surviving before=%d after=%d", statesBefore, statesAfter)
	}
}

// listSurvivingCount counts surviving invariants for a problem — a proxy for
// "the experiment did not mutate the research state."
func listSurvivingCount(ctx context.Context, dbPath, problemID string) (int, error) {
	repo, err := store.Open(dbPath)
	if err != nil {
		return 0, err
	}
	defer repo.Close()
	rows, err := repo.ListInvariantStates(ctx, problemID, "surviving")
	if err != nil {
		return 0, err
	}
	return len(rows), nil
}
