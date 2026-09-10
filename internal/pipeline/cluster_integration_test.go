package pipeline

import (
	"context"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/canon"
)

// seedClusterCorpus seeds the four-family clustering fixture and computes a
// signature for every mechanism, returning the problem id. It is the shared
// setup for the clustering + failure-space integration tests.
func seedClusterCorpus(t *testing.T, ctx context.Context, app *App, dbPath, problemID, runID, snapshotID, vocab string) []string {
	t.Helper()
	seed, err := app.SeedMechanismFixture(ctx, MechanismFixtureSeedInput{
		DBPath:     dbPath,
		ProblemID:  problemID,
		RunID:      runID,
		SnapshotID: snapshotID,
		Path:       fixturePath("cluster_four_families.json"),
	})
	if err != nil {
		t.Fatalf("seed cluster corpus: %v", err)
	}
	if len(seed.MechanismIDs) != 4 {
		t.Fatalf("fixture produced %d mechanisms, want 4", len(seed.MechanismIDs))
	}
	for _, mechID := range seed.MechanismIDs {
		if _, err := app.SignatureMechanism(ctx, SignatureInput{DBPath: dbPath, MechanismID: mechID, VocabVersion: vocab}); err != nil {
			t.Fatalf("signature %s: %v", mechID, err)
		}
	}
	return seed.MechanismIDs
}

// TestIntegrationClusterBuildFamiliesAndCoverage exercises the full offline
// path: seed -> signature -> cluster build. It asserts the deterministic family
// structure (a redundant surface-variant family, a distinct family, and an
// incomparable isolate), redundancy accounting, and coverage reporting.
func TestIntegrationClusterBuildFamiliesAndCoverage(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	problemID, runID, snapshotID := seedProblemAndSnapshot(t, ctx, dbPath, now)
	seedClusterCorpus(t, ctx, app, dbPath, problemID, runID, snapshotID, canon.VocabularyMechanismV1)

	built, err := app.BuildClustering(ctx, ClusterBuildInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("BuildClustering: %v", err)
	}
	if !built.Created {
		t.Fatal("first cluster build should be created")
	}
	run := built.ClusterRun
	if run.SignatureCount != 4 {
		t.Fatalf("signature_count = %d, want 4", run.SignatureCount)
	}
	// Expected families:
	//   1) modular-descent {congruence, affine} -> one redundant family (2 members)
	//   2) global-averaging -> distinct family (1 member)
	//   3) novel-operator-isolate -> incomparable isolate (1 member)
	if run.FamilyCount != 3 {
		t.Fatalf("family_count = %d, want 3", run.FamilyCount)
	}

	var redundantFamilies, isolates, redundantMembers int
	for _, c := range run.Clusters {
		if c.MemberCount == 2 {
			redundantFamilies++
		}
		if c.Isolate {
			isolates++
		}
		for _, m := range c.Members {
			if m.Redundant {
				redundantMembers++
			}
		}
	}
	if redundantFamilies != 1 {
		t.Fatalf("redundant (2-member) families = %d, want 1", redundantFamilies)
	}
	if isolates != 1 {
		t.Fatalf("isolate families = %d, want 1", isolates)
	}
	if redundantMembers != 2 {
		t.Fatalf("redundant members = %d, want 2 (surface-distinct pair)", redundantMembers)
	}

	// Coverage must report the operator axis with >=2 distinct canonical values.
	var sawOperator bool
	for _, ax := range run.CoverageAxes {
		if ax.Axis == "operator" {
			sawOperator = true
			if ax.DistinctValueCount < 2 {
				t.Fatalf("operator axis distinct = %d, want >=2", ax.DistinctValueCount)
			}
		}
	}
	if !sawOperator {
		t.Fatal("operator axis missing from coverage")
	}

	if run.Status != "clean" {
		t.Fatalf("status = %q, want clean under mechanism/v1", run.Status)
	}
}

// TestIntegrationClusterBuildIdempotent verifies a re-run under the identical
// version tuple returns the existing run (never rewrites) with a stable id.
func TestIntegrationClusterBuildIdempotent(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	problemID, runID, snapshotID := seedProblemAndSnapshot(t, ctx, dbPath, now)
	seedClusterCorpus(t, ctx, app, dbPath, problemID, runID, snapshotID, canon.VocabularyMechanismV1)

	first, err := app.BuildClustering(ctx, ClusterBuildInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("first build: %v", err)
	}
	second, err := app.BuildClustering(ctx, ClusterBuildInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("second build: %v", err)
	}
	if !first.Created {
		t.Fatal("first build should be created")
	}
	if second.Created {
		t.Fatal("second identical build should NOT be created (idempotent)")
	}
	if first.ClusterRun.ID != second.ClusterRun.ID {
		t.Fatalf("idempotent build produced different ids: %s vs %s", first.ClusterRun.ID, second.ClusterRun.ID)
	}
	if first.ClusterRun.ThresholdsHash != second.ClusterRun.ThresholdsHash {
		t.Fatal("thresholds hash changed across identical builds")
	}

	// Show round-trips the persisted run.
	shown, err := app.ShowClustering(ctx, ClusterShowInput{DBPath: dbPath, ClusterRunID: first.ClusterRun.ID})
	if err != nil {
		t.Fatalf("show: %v", err)
	}
	if shown.ClusterRun.FamilyCount != first.ClusterRun.FamilyCount {
		t.Fatalf("show family_count = %d, want %d", shown.ClusterRun.FamilyCount, first.ClusterRun.FamilyCount)
	}
}

// TestIntegrationFailureSpaceFromClusterRun materializes a failure space from a
// cluster run and asserts the outcome partition preserves the per-family
// outcome classes and inherits the coverage report.
func TestIntegrationFailureSpaceFromClusterRun(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	problemID, runID, snapshotID := seedProblemAndSnapshot(t, ctx, dbPath, now)
	seedClusterCorpus(t, ctx, app, dbPath, problemID, runID, snapshotID, canon.VocabularyMechanismV1)

	if _, err := app.BuildClustering(ctx, ClusterBuildInput{DBPath: dbPath, ProblemID: problemID}); err != nil {
		t.Fatalf("cluster build: %v", err)
	}

	built, err := app.BuildFailureSpace(ctx, FailureSpaceBuildInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("BuildFailureSpace: %v", err)
	}
	if !built.Created {
		t.Fatal("first failure-space build should be created")
	}
	fs := built.FailureSpace
	if fs.Revision != 1 {
		t.Fatalf("first revision = %d, want 1", fs.Revision)
	}
	if fs.DistinctFamilyCount != 3 {
		t.Fatalf("distinct_family_count = %d, want 3", fs.DistinctFamilyCount)
	}
	if fs.RedundantMemberCount != 2 {
		t.Fatalf("redundant_member_count = %d, want 2", fs.RedundantMemberCount)
	}

	// Outcome partition: the three representatives carry partial_success,
	// failure, and partial_failure -> one family each.
	byClass := map[string]int{}
	total := 0
	for _, o := range fs.Outcomes {
		byClass[o.OutcomeClass] = o.FamilyCount
		total += o.FamilyCount
	}
	if total != 3 {
		t.Fatalf("outcome family total = %d, want 3", total)
	}
	for _, want := range []string{"partial_success", "failure", "partial_failure"} {
		if byClass[want] != 1 {
			t.Fatalf("outcome %q family count = %d, want 1 (partition: %+v)", want, byClass[want], byClass)
		}
	}

	// Coverage axes inherited from the cluster run.
	if len(fs.Axes) == 0 {
		t.Fatal("failure space inherited no coverage axes")
	}

	// Idempotent per cluster run.
	second, err := app.BuildFailureSpace(ctx, FailureSpaceBuildInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("second failure-space build: %v", err)
	}
	if second.Created {
		t.Fatal("second failure-space build for same cluster run should not be created")
	}
	if second.FailureSpace.ID != fs.ID {
		t.Fatal("idempotent failure space produced a different id")
	}
}

// TestIntegrationClusterDegradedUnderLossyVocabulary is the R9/KTD-8 regression
// at the population level: the same corpus clustered under the lossy vocabulary
// collapses outcome-predictive structure and the run is marked degraded with a
// recorded discrimination-loss finding.
func TestIntegrationClusterDegradedUnderLossyVocabulary(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	problemID, runID, snapshotID := seedProblemAndSnapshot(t, ctx, dbPath, now)

	// Seed the two lossy-collapse fixtures (different outcomes, structurally
	// identical under the lossy vocabulary) and sign them under v0-lossy.
	seedA, err := app.SeedMechanismFixture(ctx, MechanismFixtureSeedInput{DBPath: dbPath, ProblemID: problemID, RunID: runID, SnapshotID: snapshotID, Path: fixturePath("case5_lossy_a.json")})
	if err != nil {
		t.Fatalf("seed a: %v", err)
	}
	seedB, err := app.SeedMechanismFixture(ctx, MechanismFixtureSeedInput{DBPath: dbPath, ProblemID: problemID, RunID: runID, SnapshotID: snapshotID, Path: fixturePath("case5_lossy_b.json")})
	if err != nil {
		t.Fatalf("seed b: %v", err)
	}
	for _, id := range []string{seedA.MechanismIDs[0], seedB.MechanismIDs[0]} {
		if _, err := app.SignatureMechanism(ctx, SignatureInput{DBPath: dbPath, MechanismID: id, VocabVersion: canon.VocabularyMechanismV0Lossy}); err != nil {
			t.Fatalf("signature %s: %v", id, err)
		}
	}

	built, err := app.BuildClustering(ctx, ClusterBuildInput{DBPath: dbPath, ProblemID: problemID, VocabVersion: canon.VocabularyMechanismV0Lossy})
	if err != nil {
		t.Fatalf("BuildClustering (lossy): %v", err)
	}
	if built.ClusterRun.Status != "degraded" {
		t.Fatalf("status = %q, want degraded under lossy vocabulary", built.ClusterRun.Status)
	}
	if len(built.ClusterRun.DiscriminationLoss) == 0 {
		t.Fatal("expected a recorded discrimination-loss finding under lossy vocabulary")
	}
}
