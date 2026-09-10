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
		Status:      domain.RunStatusSucceeded,
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
		Status:      domain.RunStatusSucceeded,
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
		Status:      domain.RunStatusSucceeded,
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
