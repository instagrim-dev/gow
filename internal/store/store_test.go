package store

import (
	"context"
	"database/sql"
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
	// Stamp EVERY known migration as applied so none run, leaving the schema with
	// only schema_migrations. Migrate must then fail in validateSchemaTables
	// because the required tables are absent. (Stamping only the highest version
	// would let migrations 1..N-1 run and create the tables, defeating the test.)
	now := time.Now().UTC().Format(time.RFC3339Nano)
	for v := 1; v <= currentSchemaVersion; v++ {
		if _, err := store.db.ExecContext(ctx, `INSERT INTO schema_migrations(version, applied_at) VALUES(?, ?)`, v, now); err != nil {
			t.Fatalf("insert schema_migrations row error = %v", err)
		}
	}

	if err := store.Migrate(ctx); !errors.Is(err, ErrCorruptStore) {
		t.Fatalf("Migrate() error = %v, want ErrCorruptStore", err)
	}
}

// TestMigrateV9RepairsPreFixV6Schema is the review-blocker regression: a
// database created under the earlier v6 schema (where migrations 4/5 were later
// corrected in place) recorded migrations 1-6 as applied, so reopening it on
// current code must still upgrade the alias uniqueness key and add the signature
// provenance columns via migration 9 — not silently retain the incompatible
// schema. It reconstructs the actual pre-fix v6 shapes, stamps 1-6 applied, then
// runs Migrate and asserts the repair.
func TestMigrateV9RepairsPreFixV6Schema(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "newf.db")
	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer store.Close()

	// Build a REAL v6 database by running the actual migration SQL for versions
	// 1-6, then downgrade the two affected areas to their pre-fix shapes. This
	// faithfully reproduces a database created before 4/5 were corrected: all
	// v1-6 tables exist, but the alias PK omits canonical_id and the signature
	// provenance columns are absent.
	for _, m := range migrations {
		if m.version > 6 {
			break
		}
		if m.sql == "" {
			t.Fatalf("migration %d has no SQL to replay", m.version)
		}
		if _, err := store.db.ExecContext(ctx, m.sql); err != nil {
			t.Fatalf("replay migration %d: %v", m.version, err)
		}
	}
	// Downgrade alias table to the pre-fix 3-column PK (drop immutable triggers
	// first so we can rebuild it).
	downgrade := []string{
		`DROP TRIGGER IF EXISTS canonical_term_aliases_immutable_update`,
		`DROP TRIGGER IF EXISTS canonical_term_aliases_immutable_delete`,
		`DROP TABLE canonical_term_aliases`,
		`CREATE TABLE canonical_term_aliases (
  vocabulary_version TEXT NOT NULL,
  canonical_id TEXT NOT NULL,
  field_kind TEXT NOT NULL,
  alias_normalized TEXT NOT NULL,
  PRIMARY KEY(vocabulary_version, field_kind, alias_normalized)
)`,
		// Downgrade signature provenance tables: drop triggers, rebuild without
		// the claim_status / support columns.
		`DROP TRIGGER IF EXISTS signature_postures_immutable_update`,
		`DROP TRIGGER IF EXISTS signature_postures_immutable_delete`,
		`DROP TABLE signature_postures`,
		`CREATE TABLE signature_postures (signature_id TEXT NOT NULL, axis TEXT NOT NULL, value TEXT NOT NULL, PRIMARY KEY(signature_id, axis))`,
		`DROP TRIGGER IF EXISTS signature_boundaries_immutable_update`,
		`DROP TRIGGER IF EXISTS signature_boundaries_immutable_delete`,
		`DROP TABLE signature_boundaries`,
		`CREATE TABLE signature_boundaries (signature_id TEXT NOT NULL, surface_label TEXT NOT NULL, resolution_state TEXT NOT NULL, canonical_id TEXT NOT NULL DEFAULT '', relation TEXT NOT NULL DEFAULT '', ordinal INTEGER NOT NULL, PRIMARY KEY(signature_id, ordinal))`,
		`DROP TRIGGER IF EXISTS signature_outcomes_immutable_update`,
		`DROP TRIGGER IF EXISTS signature_outcomes_immutable_delete`,
		`DROP TABLE signature_outcomes`,
		`CREATE TABLE signature_outcomes (signature_id TEXT PRIMARY KEY, class TEXT NOT NULL)`,
	}
	for _, s := range downgrade {
		if _, err := store.db.ExecContext(ctx, s); err != nil {
			t.Fatalf("downgrade DDL error: %v\n%s", err, s)
		}
	}
	// Seed a vocabulary + an alias row so the rebuild must preserve data.
	if _, err := store.db.ExecContext(ctx, `INSERT INTO canonical_vocabulary(version, created_at, notes) VALUES('mechanism/v1','t','')`); err != nil {
		t.Fatalf("seed vocab: %v", err)
	}
	if _, err := store.db.ExecContext(ctx, `INSERT INTO canonical_term_aliases(vocabulary_version, canonical_id, field_kind, alias_normalized) VALUES('mechanism/v1','core.operator.alpha','operator','shared')`); err != nil {
		t.Fatalf("seed alias: %v", err)
	}
	// Stamp migrations 1-6 as applied (what a real pre-fix v6 DB recorded).
	for v := 1; v <= 6; v++ {
		if _, err := store.db.ExecContext(ctx, `INSERT INTO schema_migrations(version, applied_at) VALUES(?, 't')`, v); err != nil {
			t.Fatalf("stamp migration %d: %v", v, err)
		}
	}

	if err := store.Migrate(ctx); err != nil {
		t.Fatalf("Migrate() from pre-fix v6 error = %v", err)
	}

	// Repair 1: provenance columns now present.
	for _, c := range []struct{ table, column string }{
		{"signature_postures", "claim_status"},
		{"signature_boundaries", "claim_status"},
		{"signature_boundaries", "support_snapshot_id"},
		{"signature_boundaries", "support_locator"},
		{"signature_outcomes", "claim_status"},
	} {
		if !tableHasColumn(t, ctx, store.db, c.table, c.column) {
			t.Fatalf("migration 9 did not add %s.%s to pre-fix v6 schema", c.table, c.column)
		}
	}

	// Repair 2: alias PK now includes canonical_id — the same (version,
	// field_kind, alias) may bind to a SECOND canonical id without a UNIQUE
	// violation (intentional within-field-kind ambiguity). The preserved row
	// must still be present.
	if _, err := store.db.ExecContext(ctx, `INSERT INTO canonical_term_aliases(vocabulary_version, canonical_id, field_kind, alias_normalized) VALUES('mechanism/v1','core.operator.beta','operator','shared')`); err != nil {
		t.Fatalf("second binding rejected after repair (PK still 3-column?): %v", err)
	}
	var count int
	if err := store.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM canonical_term_aliases WHERE alias_normalized='shared'`).Scan(&count); err != nil {
		t.Fatalf("count aliases: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 preserved+new alias bindings, got %d", count)
	}

	// Repair 3: durable rejected-terms table exists.
	var exists bool
	if err := store.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type='table' AND name='canonical_rejected_terms')`).Scan(&exists); err != nil || !exists {
		t.Fatalf("canonical_rejected_terms missing after migration 9 (err=%v)", err)
	}

	// Idempotent: re-running Migrate is a no-op and does not error.
	if err := store.Migrate(ctx); err != nil {
		t.Fatalf("second Migrate() after repair error = %v", err)
	}
}

func tableHasColumn(t *testing.T, ctx context.Context, db *sql.DB, table, column string) bool {
	t.Helper()
	rows, err := db.QueryContext(ctx, "PRAGMA table_info("+table+")")
	if err != nil {
		t.Fatalf("PRAGMA table_info(%s): %v", table, err)
	}
	defer rows.Close()
	for rows.Next() {
		var cid, notnull, pk int
		var name, ctype string
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			t.Fatalf("scan table_info: %v", err)
		}
		if name == column {
			return true
		}
	}
	return false
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

func TestUpdateRunStatusTransitionsLifecycle(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := openTestStore(t)
	defer store.Close()

	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	runID := domain.NewRunID(now)
	problem := domain.NewProblem{
		ID:             domain.NewProblemID(now),
		Slug:           "run-status-lifecycle",
		Statement:      "run status lifecycle",
		Status:         domain.ProblemStatusActive,
		CreatedAt:      now,
		CreatedByRunID: runID,
	}
	run := domain.NewRun{
		ID:          runID,
		ProblemID:   problem.ID,
		Operation:   "ingest",
		Status:      domain.RunStatusRunning,
		InputRef:    "ingest",
		ToolName:    "newf",
		ToolVersion: "dev",
		StartedAt:   now,
		CompletedAt: now,
	}
	if _, _, err := store.CreateProblemWithRun(ctx, problem, run); err != nil {
		t.Fatalf("CreateProblemWithRun() error = %v", err)
	}

	summary := "1 item(s) failed: a: boom"
	later := now.Add(time.Second)
	if err := store.UpdateRunStatus(ctx, runID, domain.RunStatusFailed, later, &summary); err != nil {
		t.Fatalf("UpdateRunStatus() error = %v", err)
	}

	got, err := store.GetRun(ctx, runID)
	if err != nil {
		t.Fatalf("GetRun() error = %v", err)
	}
	if got.Status != domain.RunStatusFailed {
		t.Fatalf("GetRun().Status = %q, want %q", got.Status, domain.RunStatusFailed)
	}
	if got.ErrorSummary == nil || *got.ErrorSummary != summary {
		t.Fatalf("GetRun().ErrorSummary = %v, want %q", got.ErrorSummary, summary)
	}
	if !got.CompletedAt.Equal(later) {
		t.Fatalf("GetRun().CompletedAt = %v, want %v", got.CompletedAt, later)
	}

	// Clearing error_summary on a subsequent completed transition must persist nil.
	if err := store.UpdateRunStatus(ctx, runID, domain.RunStatusCompleted, later, nil); err != nil {
		t.Fatalf("UpdateRunStatus(completed) error = %v", err)
	}
	got, err = store.GetRun(ctx, runID)
	if err != nil {
		t.Fatalf("GetRun() error = %v", err)
	}
	if got.Status != domain.RunStatusCompleted {
		t.Fatalf("GetRun().Status = %q, want completed", got.Status)
	}
	if got.ErrorSummary != nil {
		t.Fatalf("GetRun().ErrorSummary = %v, want nil", got.ErrorSummary)
	}
}

func TestUpdateRunStatusRejectsUnknownRunAndStatus(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	store := openTestStore(t)
	defer store.Close()

	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	if err := store.UpdateRunStatus(ctx, domain.NewRunID(now), domain.RunStatusCompleted, now, nil); !errors.Is(err, ErrNotFound) {
		t.Fatalf("UpdateRunStatus(missing run) error = %v, want ErrNotFound", err)
	}
	if err := store.UpdateRunStatus(ctx, domain.NewRunID(now), domain.RunStatus("bogus"), now, nil); err == nil {
		t.Fatal("UpdateRunStatus(bogus status) succeeded, want error")
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
