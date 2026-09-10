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
	// The corpus target is mechanistically distinct from the break proposal, so
	// the honest conclusion is no_recovery — computed by code, not asserted.
	if exp.Conclusion != "no_recovery" {
		t.Fatalf("conclusion = %q, want no_recovery for the distinct corpus target", exp.Conclusion)
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
