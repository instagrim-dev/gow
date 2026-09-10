package store

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/domain"
)

func TestMigrateFreshAndIdempotent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "state", "newf.db")
	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()

	if err := store.Migrate(ctx); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}
	if err := store.Migrate(ctx); err != nil {
		t.Fatalf("second Migrate() error = %v", err)
	}

	version, err := store.SchemaVersion(ctx)
	if err != nil {
		t.Fatalf("SchemaVersion() error = %v", err)
	}
	if version != currentSchemaVersion {
		t.Fatalf("SchemaVersion() = %d, want %d", version, currentSchemaVersion)
	}
}

func TestCreateAndFetchProblemAndRun(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := openTestStore(t)
	defer store.Close()

	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	runID := domain.NewRunID(now)
	problem := domain.NewProblem{
		ID:             domain.NewProblemID(now),
		Slug:           "erdos-straus-conjecture",
		Statement:      "Erdos-Straus conjecture",
		Status:         domain.ProblemStatusActive,
		CreatedAt:      now,
		CreatedByRunID: runID,
	}
	run := domain.NewRun{
		ID:          runID,
		ProblemID:   problem.ID,
		Operation:   "init",
		Status:      domain.RunStatusInitialized,
		InputRef:    "problem_slug:erdos-straus-conjecture",
		ToolName:    "newf",
		ToolVersion: "dev",
		StartedAt:   now,
		CompletedAt: now,
	}

	createdProblem, createdRun, err := store.CreateProblemWithRun(ctx, problem, run)
	if err != nil {
		t.Fatalf("CreateProblemWithRun() error = %v", err)
	}

	gotProblem, err := store.GetProblem(ctx, createdProblem.ID)
	if err != nil {
		t.Fatalf("GetProblem() error = %v", err)
	}
	if gotProblem.CreatedByRunID != createdRun.ID {
		t.Fatalf("GetProblem().CreatedByRunID = %q, want %q", gotProblem.CreatedByRunID, createdRun.ID)
	}

	gotRun, err := store.GetRun(ctx, createdRun.ID)
	if err != nil {
		t.Fatalf("GetRun() error = %v", err)
	}
	if gotRun.ProblemID != createdProblem.ID {
		t.Fatalf("GetRun().ProblemID = %q, want %q", gotRun.ProblemID, createdProblem.ID)
	}

	problems, err := store.ListProblems(ctx)
	if err != nil {
		t.Fatalf("ListProblems() error = %v", err)
	}
	if len(problems) != 1 {
		t.Fatalf("ListProblems() len = %d, want 1", len(problems))
	}
}

func TestCreateProblemWithRunRollsBackOnFailure(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := openTestStore(t, WithAfterProblemInsert(func() error {
		return errors.New("boom")
	}))
	defer store.Close()

	now := time.Now().UTC()
	problemID := domain.NewProblemID(now)
	runID := domain.NewRunID(now)
	_, _, err := store.CreateProblemWithRun(ctx, domain.NewProblem{
		ID:             problemID,
		Slug:           "rollback-case",
		Statement:      "Rollback case",
		Status:         domain.ProblemStatusActive,
		CreatedAt:      now,
		CreatedByRunID: runID,
	}, domain.NewRun{
		ID:          runID,
		ProblemID:   problemID,
		Operation:   "init",
		Status:      domain.RunStatusInitialized,
		InputRef:    "problem_slug:rollback-case",
		ToolName:    "newf",
		ToolVersion: "dev",
		StartedAt:   now,
		CompletedAt: now,
	})
	if err == nil {
		t.Fatal("CreateProblemWithRun() succeeded, want rollback error")
	}

	problems, err := store.ListProblems(ctx)
	if err != nil {
		t.Fatalf("ListProblems() error = %v", err)
	}
	if len(problems) != 0 {
		t.Fatalf("ListProblems() len = %d, want 0 after rollback", len(problems))
	}
}

func TestCreateProblemWithRunRejectsMismatchedLinkage(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := openTestStore(t)
	defer store.Close()

	now := time.Now().UTC()
	problemID := domain.NewProblemID(now)
	runID := domain.NewRunID(now)
	_, _, err := store.CreateProblemWithRun(ctx, domain.NewProblem{
		ID:             problemID,
		Slug:           "bad-linkage",
		Statement:      "Bad linkage",
		Status:         domain.ProblemStatusActive,
		CreatedAt:      now,
		CreatedByRunID: runID,
	}, domain.NewRun{
		ID:          runID,
		ProblemID:   domain.NewProblemID(now.Add(time.Second)),
		Operation:   "init",
		Status:      domain.RunStatusInitialized,
		InputRef:    "problem_slug:bad-linkage",
		ToolName:    "newf",
		ToolVersion: "dev",
		StartedAt:   now,
		CompletedAt: now,
	})
	if err == nil {
		t.Fatal("CreateProblemWithRun() succeeded for mismatched linkage")
	}
}

func TestGettersRejectCrossClassIDs(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := openTestStore(t)
	defer store.Close()

	runID := domain.NewRunID(time.Now().UTC())
	if _, err := store.GetProblem(ctx, runID); err == nil {
		t.Fatal("GetProblem() succeeded for run ID")
	}

	problemID := domain.NewProblemID(time.Now().UTC())
	if _, err := store.GetRun(ctx, problemID); err == nil {
		t.Fatal("GetRun() succeeded for problem ID")
	}
}

func TestMigrateRejectsUnknownSchemaVersion(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "newf.db")
	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()

	if err := store.Migrate(ctx); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}
	if _, err := store.db.ExecContext(ctx, `INSERT INTO schema_migrations(version, applied_at) VALUES(?, ?)`, currentSchemaVersion+1, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
		t.Fatalf("insert schema_migrations row error = %v", err)
	}

	if err := store.Migrate(ctx); !errors.Is(err, ErrCorruptStore) {
		t.Fatalf("second Migrate() error = %v, want ErrCorruptStore", err)
	}
}

func TestMigrateRejectsMissingSchemaTables(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "newf.db")
	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()

	if _, err := store.db.ExecContext(ctx, `CREATE TABLE schema_migrations (version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL)`); err != nil {
		t.Fatalf("create schema_migrations error = %v", err)
	}
	if _, err := store.db.ExecContext(ctx, `INSERT INTO schema_migrations(version, applied_at) VALUES(?, ?)`, currentSchemaVersion, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
		t.Fatalf("insert schema_migrations row error = %v", err)
	}

	if err := store.Migrate(ctx); !errors.Is(err, ErrCorruptStore) {
		t.Fatalf("Migrate() error = %v, want ErrCorruptStore", err)
	}
}

func TestNextProblemSlugIgnoresNonNumericSuffixes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := openTestStore(t)
	defer store.Close()

	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	for i, slug := range []string{"topic", "topic-2abc"} {
		start := now.Add(time.Duration(i) * time.Second)
		runID := domain.NewRunID(start)
		problemID := domain.NewProblemID(start)
		_, _, err := store.CreateProblemWithRun(ctx, domain.NewProblem{
			ID:             problemID,
			Slug:           slug,
			Statement:      slug,
			Status:         domain.ProblemStatusActive,
			CreatedAt:      start,
			CreatedByRunID: runID,
		}, domain.NewRun{
			ID:          runID,
			ProblemID:   problemID,
			Operation:   "init",
			Status:      domain.RunStatusInitialized,
			InputRef:    "problem_slug:" + slug,
			ToolName:    "newf",
			ToolVersion: "dev",
			StartedAt:   start,
			CompletedAt: start,
		})
		if err != nil {
			t.Fatalf("CreateProblemWithRun(%q) error = %v", slug, err)
		}
	}

	next, err := store.NextProblemSlug(ctx, "topic")
	if err != nil {
		t.Fatalf("NextProblemSlug() error = %v", err)
	}
	if next != "topic-2" {
		t.Fatalf("NextProblemSlug() = %q, want %q", next, "topic-2")
	}
}

func TestCreateSourceSnapshotIdempotentAndRevisioned(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := openTestStore(t)
	defer repo.Close()

	problemID, runID := seedProblemForSourceTests(t, ctx, repo)
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	first, err := repo.CreateSourceSnapshot(ctx, SnapshotAdmission{
		ProblemID:   problemID,
		Kind:        domain.SourceKindLocalPath,
		LogicalName: "paper.md",
		Origin:      "/tmp/paper.md",
		SHA256:      "aaa",
		ByteLength:  3,
		MediaType:   "text/markdown",
		ObjectPath:  "sha256/aa/aaa",
		IngestRunID: runID,
		ObservedAt:  now,
	})
	if err != nil {
		t.Fatalf("CreateSourceSnapshot(first) error = %v", err)
	}
	if first.Status != "created_snapshot" || !first.CreatedSource {
		t.Fatalf("first admission = %+v", first)
	}

	second, err := repo.CreateSourceSnapshot(ctx, SnapshotAdmission{
		ProblemID:   problemID,
		Kind:        domain.SourceKindLocalPath,
		LogicalName: "paper.md",
		Origin:      "/tmp/paper.md",
		SHA256:      "aaa",
		ByteLength:  3,
		MediaType:   "text/markdown",
		ObjectPath:  "sha256/aa/aaa",
		IngestRunID: runID,
		ObservedAt:  now.Add(time.Second),
	})
	if err != nil {
		t.Fatalf("CreateSourceSnapshot(second) error = %v", err)
	}
	if second.Status != "existing_snapshot" || second.Snapshot.ID != first.Snapshot.ID {
		t.Fatalf("second admission = %+v", second)
	}

	third, err := repo.CreateSourceSnapshot(ctx, SnapshotAdmission{
		ProblemID:   problemID,
		Kind:        domain.SourceKindLocalPath,
		LogicalName: "paper.md",
		Origin:      "/tmp/paper.md",
		SHA256:      "bbb",
		ByteLength:  4,
		MediaType:   "text/markdown",
		ObjectPath:  "sha256/bb/bbb",
		IngestRunID: runID,
		ObservedAt:  now.Add(2 * time.Second),
	})
	if err != nil {
		t.Fatalf("CreateSourceSnapshot(third) error = %v", err)
	}
	if third.Status != "new_revision" {
		t.Fatalf("third status = %q, want new_revision", third.Status)
	}
	if third.Snapshot.SupersedesSnapshotID == nil || *third.Snapshot.SupersedesSnapshotID != first.Snapshot.ID {
		t.Fatalf("third supersedes = %+v, want %s", third.Snapshot.SupersedesSnapshotID, first.Snapshot.ID)
	}

	snapshots, err := repo.ListSourceSnapshots(ctx, first.Source.ID)
	if err != nil {
		t.Fatalf("ListSourceSnapshots() error = %v", err)
	}
	if len(snapshots) != 2 {
		t.Fatalf("len(snapshots) = %d, want 2", len(snapshots))
	}
}

func TestCreateSourceSnapshotDedupesBytesButKeepsDistinctSources(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := openTestStore(t)
	defer repo.Close()

	problemID, runID := seedProblemForSourceTests(t, ctx, repo)
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	first, err := repo.CreateSourceSnapshot(ctx, SnapshotAdmission{
		ProblemID:   problemID,
		Kind:        domain.SourceKindLocalPath,
		LogicalName: "paper-a.md",
		Origin:      "/tmp/a.md",
		SHA256:      "same",
		ByteLength:  42,
		MediaType:   "text/markdown",
		ObjectPath:  "sha256/sa/same",
		IngestRunID: runID,
		ObservedAt:  now,
	})
	if err != nil {
		t.Fatalf("CreateSourceSnapshot(first) error = %v", err)
	}
	second, err := repo.CreateSourceSnapshot(ctx, SnapshotAdmission{
		ProblemID:   problemID,
		Kind:        domain.SourceKindLocalPath,
		LogicalName: "paper-b.md",
		Origin:      "/tmp/b.md",
		SHA256:      "same",
		ByteLength:  42,
		MediaType:   "text/markdown",
		ObjectPath:  "sha256/sa/same",
		IngestRunID: runID,
		ObservedAt:  now.Add(time.Second),
	})
	if err != nil {
		t.Fatalf("CreateSourceSnapshot(second) error = %v", err)
	}
	if first.Source.ID == second.Source.ID {
		t.Fatalf("source IDs unexpectedly match: %s", first.Source.ID)
	}
	if first.Snapshot.ObjectPath != second.Snapshot.ObjectPath {
		t.Fatalf("object paths differ: %s vs %s", first.Snapshot.ObjectPath, second.Snapshot.ObjectPath)
	}

	summaries, err := repo.ListSourcesWithSnapshotStats(ctx, problemID)
	if err != nil {
		t.Fatalf("ListSourcesWithSnapshotStats() error = %v", err)
	}
	if len(summaries) != 2 {
		t.Fatalf("len(summaries) = %d, want 2", len(summaries))
	}
	for _, summary := range summaries {
		if summary.SnapshotCount != 1 || summary.LatestSnapshotID == nil {
			t.Fatalf("unexpected summary = %+v", summary)
		}
	}
}

func seedProblemForSourceTests(t *testing.T, ctx context.Context, repo *Store) (string, string) {
	t.Helper()

	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	problemID := domain.NewProblemID(now)
	runID := domain.NewRunID(now)
	_, _, err := repo.CreateProblemWithRun(ctx, domain.NewProblem{
		ID:             problemID,
		Slug:           "source-test-problem",
		Statement:      "Source test problem",
		Status:         domain.ProblemStatusActive,
		CreatedAt:      now,
		CreatedByRunID: runID,
	}, domain.NewRun{
		ID:          runID,
		ProblemID:   problemID,
		Operation:   "init",
		Status:      domain.RunStatusInitialized,
		InputRef:    "problem_slug:source-test-problem",
		ToolName:    "newf",
		ToolVersion: "dev",
		StartedAt:   now,
		CompletedAt: now,
	})
	if err != nil {
		t.Fatalf("CreateProblemWithRun() error = %v", err)
	}

	ingestRunNow := now.Add(time.Second)
	ingestRunID := domain.NewRunID(ingestRunNow)
	if _, err := repo.CreateRun(ctx, domain.NewRun{
		ID:          ingestRunID,
		ProblemID:   problemID,
		Operation:   "ingest",
		Status:      domain.RunStatusCompleted,
		InputRef:    "ingest",
		ToolName:    "newf",
		ToolVersion: "dev",
		StartedAt:   ingestRunNow,
		CompletedAt: ingestRunNow,
	}); err != nil {
		t.Fatalf("CreateRun() error = %v", err)
	}
	return problemID, ingestRunID
}

func openTestStore(t *testing.T, opts ...Option) *Store {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "newf.db")
	store, err := Open(dbPath, opts...)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if err := store.Migrate(context.Background()); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}
	return store
}
