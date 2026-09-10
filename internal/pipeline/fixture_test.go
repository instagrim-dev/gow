package pipeline

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/store"
)

// newRealStoreApp returns an App whose openStoreFn opens a fresh real store at
// the same temp DB path on every call (mirroring production), plus a helper to
// open the store directly for test seeding/assertions.
func newRealStoreApp(t *testing.T, now time.Time) (*App, string) {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "newf.db")
	app := &App{
		version: "test",
		now:     func() time.Time { return now },
		getwd:   func() (string, error) { return filepath.Dir(dbPath), nil },
	}
	app.openStoreFn = func(ctx context.Context, _ string) (string, problemStore, error) {
		repo, err := store.Open(dbPath)
		if err != nil {
			return "", nil, err
		}
		if err := repo.Migrate(ctx); err != nil {
			repo.Close()
			return "", nil, err
		}
		if err := app.seedVocabularies(ctx, repo); err != nil {
			repo.Close()
			return "", nil, err
		}
		return dbPath, repo, nil
	}
	return app, dbPath
}

func openTestStore(t *testing.T, ctx context.Context, dbPath string) *store.Store {
	t.Helper()
	repo, err := store.Open(dbPath)
	if err != nil {
		t.Fatalf("store.Open() error = %v", err)
	}
	if err := repo.Migrate(ctx); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}
	return repo
}

// seedProblemAndSnapshot creates a problem+run and one source snapshot, returning
// the IDs the fixture seeder needs.
func seedProblemAndSnapshot(t *testing.T, ctx context.Context, dbPath string, now time.Time) (problemID, runID, snapshotID string) {
	t.Helper()
	repo := openTestStore(t, ctx, dbPath)
	defer repo.Close()

	runID = domain.NewRunID(now)
	problem := domain.NewProblem{
		ID:             domain.NewProblemID(now),
		Slug:           "test-problem",
		Statement:      "Test problem",
		Status:         domain.ProblemStatusActive,
		CreatedAt:      now,
		CreatedByRunID: runID,
	}
	run := domain.NewRun{
		ID:          runID,
		ProblemID:   problem.ID,
		Operation:   "init",
		Status:      domain.RunStatusInitialized,
		InputRef:    "problem_slug:test-problem",
		ToolName:    "newf",
		ToolVersion: "test",
		StartedAt:   now,
		CompletedAt: now,
	}
	createdProblem, createdRun, err := repo.CreateProblemWithRun(ctx, problem, run)
	if err != nil {
		t.Fatalf("CreateProblemWithRun() error = %v", err)
	}

	admission, err := repo.CreateSourceSnapshot(ctx, store.SnapshotAdmission{
		ProblemID:   createdProblem.ID,
		Kind:        domain.SourceKindLocalPath,
		LogicalName: "approach.md",
		Origin:      "/tmp/approach.md",
		SHA256:      "deadbeef",
		ByteLength:  10,
		MediaType:   "text/markdown",
		ObjectPath:  "sha256/de/deadbeef",
		IngestRunID: createdRun.ID,
		ObservedAt:  now,
	})
	if err != nil {
		t.Fatalf("CreateSourceSnapshot() error = %v", err)
	}
	return createdProblem.ID, createdRun.ID, admission.Snapshot.ID
}

func TestSeedMechanismFixtureSingle(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	problemID, runID, snapshotID := seedProblemAndSnapshot(t, ctx, dbPath, now)

	result, err := app.SeedMechanismFixture(ctx, MechanismFixtureSeedInput{
		DBPath:     dbPath,
		ProblemID:  problemID,
		RunID:      runID,
		SnapshotID: snapshotID,
		Path:       filepath.Join("..", "..", "testdata", "fixtures", "mechanism", "single_modular_descent.json"),
	})
	if err != nil {
		t.Fatalf("SeedMechanismFixture() error = %v", err)
	}
	if len(result.MechanismIDs) != 1 {
		t.Fatalf("MechanismIDs = %d, want 1", len(result.MechanismIDs))
	}

	repo := openTestStore(t, ctx, dbPath)
	defer repo.Close()
	detail, err := repo.GetMechanismDetail(ctx, result.MechanismIDs[0])
	if err != nil {
		t.Fatalf("GetMechanismDetail() error = %v", err)
	}
	if detail.Mechanism.Locality != domain.LocalityLocal {
		t.Fatalf("locality = %q, want local", detail.Mechanism.Locality)
	}
	if detail.Outcome.Class != domain.OutcomePartialSuccess {
		t.Fatalf("outcome class = %q, want partial_success", detail.Outcome.Class)
	}
	// Surface labels preserved verbatim (no canonicalization at load time).
	var foundOperator bool
	for _, attr := range detail.Attributes {
		if attr.Kind == domain.AttrOperator && attr.Value == "modular decomposition" {
			foundOperator = true
		}
	}
	if !foundOperator {
		t.Fatalf("operator surface label not preserved; attributes = %+v", detail.Attributes)
	}
	if len(detail.Boundaries) != 1 || detail.Boundaries[0].Condition != "composite modulus" {
		t.Fatalf("boundaries not preserved; got %+v", detail.Boundaries)
	}
	if len(detail.Support) != 1 || detail.Support[0].SupportKind != domain.SupportExplicit {
		t.Fatalf("support not preserved; got %+v", detail.Support)
	}
}

func TestSeedMechanismFixtureMultiApproach(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	problemID, runID, snapshotID := seedProblemAndSnapshot(t, ctx, dbPath, now)

	result, err := app.SeedMechanismFixture(ctx, MechanismFixtureSeedInput{
		DBPath:     dbPath,
		ProblemID:  problemID,
		RunID:      runID,
		SnapshotID: snapshotID,
		Path:       filepath.Join("..", "..", "testdata", "fixtures", "mechanism", "multi_two_approaches.json"),
	})
	if err != nil {
		t.Fatalf("SeedMechanismFixture() error = %v", err)
	}
	if len(result.ApproachIDs) != 2 {
		t.Fatalf("ApproachIDs = %d, want 2", len(result.ApproachIDs))
	}
	if result.ApproachIDs[0] == result.ApproachIDs[1] {
		t.Fatalf("expected two distinct approaches under one revision, got duplicate %q", result.ApproachIDs[0])
	}
}

func TestSeedMechanismFixtureInvalidAborts(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	problemID, runID, snapshotID := seedProblemAndSnapshot(t, ctx, dbPath, now)

	// Invalid: bad locality enum should fail domain validation, leaving no rows.
	bad := MechanismFixture{Approaches: []MechanismFixtureApproach{{
		LogicalIdentity:  "bad",
		Label:            "Bad",
		Locality:         "not-a-locality",
		ConstructionMode: "constructive",
		UncertaintyMode:  "deterministic",
		Outcome:          MechanismFixtureOutcome{Class: "failure"},
	}}}
	_, err := app.SeedMechanismFixture(ctx, MechanismFixtureSeedInput{
		DBPath:     dbPath,
		ProblemID:  problemID,
		RunID:      runID,
		SnapshotID: snapshotID,
		Fixture:    &bad,
	})
	if err == nil {
		t.Fatal("SeedMechanismFixture() succeeded for invalid fixture, want error")
	}

	repo := openTestStore(t, ctx, dbPath)
	defer repo.Close()
	approaches, err := repo.ListApproaches(ctx, problemID)
	if err != nil {
		t.Fatalf("ListApproaches() error = %v", err)
	}
	if len(approaches) != 0 {
		t.Fatalf("expected no approaches after aborted seed, got %d", len(approaches))
	}
}
