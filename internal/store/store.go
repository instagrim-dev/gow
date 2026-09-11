package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/instagrim-dev/newf/internal/domain"
)

var (
	ErrNotFound      = errors.New("record not found")
	ErrDuplicateSlug = errors.New("duplicate problem slug")
	ErrMigration     = errors.New("migration failed")
	ErrCorruptStore  = errors.New("corrupt store")
)

type Option func(*Store)

type Store struct {
	db                 *sql.DB
	path               string
	afterProblemInsert func() error
}

func WithAfterProblemInsert(fn func() error) Option {
	return func(s *Store) {
		s.afterProblemInsert = fn
	}
}

func Open(path string, opts ...Option) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)

	store := &Store{db: db, path: path}
	for _, opt := range opts {
		opt(store)
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		_ = db.Close()
		return nil, err
	}

	return store, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Migrate(ctx context.Context) error {
	// SQLite requires foreign_keys to be toggled OUTSIDE a transaction. Disable it
	// for the duration of migration so that a table rebuild which must drop and
	// recreate a parent referenced by existing children (e.g. the legacy
	// cluster_runs identity change in v10) does not fail its deferred FK check at
	// commit. Children keep their textual parent ids across the rebuild, so a
	// PRAGMA foreign_key_check before commit still proves no reference was broken.
	// Re-enable on the way out regardless of outcome.
	if _, err := s.db.ExecContext(ctx, `PRAGMA foreign_keys = OFF`); err != nil {
		return fmt.Errorf("%w: %v", ErrMigration, err)
	}
	defer func() { _, _ = s.db.ExecContext(ctx, `PRAGMA foreign_keys = ON`) }()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL)`); err != nil {
		return fmt.Errorf("%w: %v", ErrMigration, err)
	}

	applied := map[int]struct{}{}
	rows, err := tx.QueryContext(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrCorruptStore, err)
	}
	defer rows.Close()

	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return fmt.Errorf("%w: %v", ErrCorruptStore, err)
		}
		if !isKnownMigrationVersion(version) {
			return fmt.Errorf("%w: unknown schema migration version %d", ErrCorruptStore, version)
		}
		applied[version] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("%w: %v", ErrCorruptStore, err)
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	for _, migration := range migrations {
		if _, ok := applied[migration.version]; ok {
			continue
		}
		if migration.apply != nil {
			if err := migration.apply(ctx, tx); err != nil {
				return fmt.Errorf("%w: %v", ErrMigration, err)
			}
		} else if _, err := tx.ExecContext(ctx, migration.sql); err != nil {
			return fmt.Errorf("%w: %v", ErrMigration, err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations(version, applied_at) VALUES(?, ?)`, migration.version, now); err != nil {
			return fmt.Errorf("%w: %v", ErrMigration, err)
		}
	}
	if err := validateSchemaTables(ctx, tx); err != nil {
		return err
	}

	// With foreign_keys disabled during migration, prove no rebuild broke a
	// reference before committing. Any dangling FK here is a migration bug.
	fkRows, err := tx.QueryContext(ctx, `PRAGMA foreign_key_check`)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrMigration, err)
	}
	if fkRows.Next() {
		var table, parent any
		var rowid, fkid any
		_ = fkRows.Scan(&table, &rowid, &parent, &fkid)
		_ = fkRows.Close()
		return fmt.Errorf("%w: foreign_key_check reported a dangling reference after migration (table=%v parent=%v)", ErrMigration, table, parent)
	}
	if err := fkRows.Err(); err != nil {
		_ = fkRows.Close()
		return fmt.Errorf("%w: %v", ErrMigration, err)
	}
	_ = fkRows.Close()

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%w: %v", ErrMigration, err)
	}

	return nil
}

func (s *Store) SchemaVersion(ctx context.Context) (int, error) {
	row := s.db.QueryRowContext(ctx, `SELECT COALESCE(MAX(version), 0) FROM schema_migrations`)
	var version int
	if err := row.Scan(&version); err != nil {
		return 0, fmt.Errorf("%w: %v", ErrCorruptStore, err)
	}
	return version, nil
}

func (s *Store) FindProblemBySlug(ctx context.Context, slug string) (domain.Problem, bool, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, slug, statement, status, created_at, created_by_run_id
FROM problems
WHERE slug = ?
`, slug)

	problem, err := scanProblem(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Problem{}, false, nil
		}
		return domain.Problem{}, false, err
	}

	return problem, true, nil
}

func (s *Store) NextProblemSlug(ctx context.Context, base string) (string, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT slug
FROM problems
WHERE slug = ? OR slug GLOB ?
`, base, base+"-*")
	if err != nil {
		return "", err
	}
	defer rows.Close()

	baseUsed := false
	maxSuffix := 1
	prefix := base + "-"
	for rows.Next() {
		var slug string
		if err := rows.Scan(&slug); err != nil {
			return "", err
		}
		if slug == base {
			baseUsed = true
			continue
		}
		if !strings.HasPrefix(slug, prefix) {
			continue
		}
		suffix := strings.TrimPrefix(slug, prefix)
		if suffix == "" {
			continue
		}
		number, convErr := strconv.Atoi(suffix)
		if convErr != nil {
			continue
		}
		if number > maxSuffix {
			maxSuffix = number
		}
	}
	if err := rows.Err(); err != nil {
		return "", err
	}
	if !baseUsed {
		return base, nil
	}

	return fmt.Sprintf("%s-%d", base, maxSuffix+1), nil
}

func (s *Store) CreateProblemWithRun(ctx context.Context, problem domain.NewProblem, run domain.NewRun) (domain.Problem, domain.Run, error) {
	if err := problem.Validate(); err != nil {
		return domain.Problem{}, domain.Run{}, err
	}
	if err := run.Validate(); err != nil {
		return domain.Problem{}, domain.Run{}, err
	}
	if run.ProblemID != problem.ID {
		return domain.Problem{}, domain.Run{}, errors.New("run problem_id must match problem id")
	}
	if problem.CreatedByRunID != run.ID {
		return domain.Problem{}, domain.Run{}, errors.New("problem created_by_run_id must match run id")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Problem{}, domain.Run{}, err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
INSERT INTO problems(id, slug, statement, status, created_at, created_by_run_id)
VALUES(?, ?, ?, ?, ?, ?)
`, problem.ID, problem.Slug, strings.TrimSpace(problem.Statement), problem.Status, formatTime(problem.CreatedAt), problem.CreatedByRunID); err != nil {
		if isDuplicateSlugError(err) {
			return domain.Problem{}, domain.Run{}, fmt.Errorf("%w: %v", ErrDuplicateSlug, err)
		}
		return domain.Problem{}, domain.Run{}, err
	}

	if s.afterProblemInsert != nil {
		if err := s.afterProblemInsert(); err != nil {
			return domain.Problem{}, domain.Run{}, err
		}
	}

	if _, err := tx.ExecContext(ctx, `
INSERT INTO runs(id, problem_id, parent_run_id, operation, status, input_ref, tool_name, tool_version, started_at, completed_at, error_summary)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, run.ID, run.ProblemID, run.ParentRunID, run.Operation, run.Status, run.InputRef, run.ToolName, run.ToolVersion, formatTime(run.StartedAt), formatTime(run.CompletedAt), run.ErrorSummary); err != nil {
		return domain.Problem{}, domain.Run{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.Problem{}, domain.Run{}, err
	}

	return domain.Problem{
			ID:             problem.ID,
			Slug:           problem.Slug,
			Statement:      strings.TrimSpace(problem.Statement),
			Status:         problem.Status,
			CreatedAt:      problem.CreatedAt.UTC(),
			CreatedByRunID: problem.CreatedByRunID,
		}, domain.Run{
			ID:           run.ID,
			ProblemID:    run.ProblemID,
			ParentRunID:  run.ParentRunID,
			Operation:    run.Operation,
			Status:       run.Status,
			InputRef:     run.InputRef,
			ToolName:     run.ToolName,
			ToolVersion:  run.ToolVersion,
			StartedAt:    run.StartedAt.UTC(),
			CompletedAt:  run.CompletedAt.UTC(),
			ErrorSummary: run.ErrorSummary,
		}, nil
}

func (s *Store) CreateRun(ctx context.Context, run domain.NewRun) (domain.Run, error) {
	if err := run.Validate(); err != nil {
		return domain.Run{}, err
	}

	if _, err := s.db.ExecContext(ctx, `
INSERT INTO runs(id, problem_id, parent_run_id, operation, status, input_ref, tool_name, tool_version, started_at, completed_at, error_summary)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, run.ID, run.ProblemID, run.ParentRunID, run.Operation, run.Status, run.InputRef, run.ToolName, run.ToolVersion, formatTime(run.StartedAt), formatTime(run.CompletedAt), run.ErrorSummary); err != nil {
		return domain.Run{}, err
	}

	return domain.Run{
		ID:           run.ID,
		ProblemID:    run.ProblemID,
		ParentRunID:  run.ParentRunID,
		Operation:    run.Operation,
		Status:       run.Status,
		InputRef:     run.InputRef,
		ToolName:     run.ToolName,
		ToolVersion:  run.ToolVersion,
		StartedAt:    run.StartedAt.UTC(),
		CompletedAt:  run.CompletedAt.UTC(),
		ErrorSummary: run.ErrorSummary,
	}, nil
}

func (s *Store) GetProblem(ctx context.Context, id string) (domain.Problem, error) {
	if err := domain.ValidateProblemID(id); err != nil {
		return domain.Problem{}, err
	}

	row := s.db.QueryRowContext(ctx, `
SELECT id, slug, statement, status, created_at, created_by_run_id
FROM problems
WHERE id = ?
`, id)

	problem, err := scanProblem(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Problem{}, fmt.Errorf("%w: problem %s", ErrNotFound, id)
		}
		return domain.Problem{}, err
	}
	return problem, nil
}

func (s *Store) ListProblems(ctx context.Context) ([]domain.Problem, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, slug, statement, status, created_at, created_by_run_id
FROM problems
ORDER BY created_at ASC, id ASC
`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var problems []domain.Problem
	for rows.Next() {
		problem, err := scanProblem(rows)
		if err != nil {
			return nil, err
		}
		problems = append(problems, problem)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return problems, nil
}

func (s *Store) GetRun(ctx context.Context, id string) (domain.Run, error) {
	if err := domain.ValidateRunID(id); err != nil {
		return domain.Run{}, err
	}

	row := s.db.QueryRowContext(ctx, `
SELECT id, problem_id, parent_run_id, operation, status, input_ref, tool_name, tool_version, started_at, completed_at, error_summary
FROM runs
WHERE id = ?
`, id)

	run, err := scanRun(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Run{}, fmt.Errorf("%w: run %s", ErrNotFound, id)
		}
		return domain.Run{}, err
	}

	return run, nil
}

// UpdateRunStatus transitions a run to a terminal (or running) lifecycle state
// after execution, recording completed_at and an optional error_summary. Run
// status is otherwise write-once at creation; this is the only mutation path,
// so run telemetry (`run show`) can model `running` -> `completed`/`failed`
// instead of being frozen at the value chosen before work happened.
func (s *Store) UpdateRunStatus(ctx context.Context, id string, status domain.RunStatus, completedAt time.Time, errorSummary *string) error {
	if err := domain.ValidateRunID(id); err != nil {
		return err
	}
	switch status {
	case domain.RunStatusInitialized, domain.RunStatusRunning, domain.RunStatusCompleted, domain.RunStatusFailed:
	default:
		return fmt.Errorf("invalid run status %q", status)
	}

	result, err := s.db.ExecContext(ctx, `
UPDATE runs SET status = ?, completed_at = ?, error_summary = ?
WHERE id = ?
`, status, formatTime(completedAt.UTC()), errorSummary, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("%w: run %s", ErrNotFound, id)
	}
	return nil
}

type SnapshotAdmission struct {
	ProblemID   string
	Kind        domain.SourceKind
	LogicalName string
	Origin      string
	SHA256      string
	ByteLength  int64
	MediaType   string
	ObjectPath  string
	IngestRunID string
	ObservedAt  time.Time
}

type SnapshotAdmissionResult struct {
	Source        domain.Source
	Snapshot      domain.SourceSnapshot
	Status        string
	CreatedSource bool
}

type SourceSummary struct {
	Source           domain.Source
	SnapshotCount    int
	LatestSnapshotID *string
}

func (s *Store) CreateSourceSnapshot(ctx context.Context, input SnapshotAdmission) (SnapshotAdmissionResult, error) {
	if err := domain.ValidateProblemID(input.ProblemID); err != nil {
		return SnapshotAdmissionResult{}, err
	}
	if strings.TrimSpace(input.Origin) == "" {
		return SnapshotAdmissionResult{}, errors.New("source origin is required")
	}
	if strings.TrimSpace(input.LogicalName) == "" {
		return SnapshotAdmissionResult{}, errors.New("source logical_name is required")
	}
	if strings.TrimSpace(input.SHA256) == "" {
		return SnapshotAdmissionResult{}, errors.New("source snapshot sha256 is required")
	}
	if strings.TrimSpace(input.MediaType) == "" {
		return SnapshotAdmissionResult{}, errors.New("source snapshot media_type is required")
	}
	if strings.TrimSpace(input.ObjectPath) == "" {
		return SnapshotAdmissionResult{}, errors.New("source snapshot object_path is required")
	}
	if err := domain.ValidateRunID(input.IngestRunID); err != nil {
		return SnapshotAdmissionResult{}, err
	}
	if input.ObservedAt.IsZero() {
		return SnapshotAdmissionResult{}, errors.New("source snapshot observed_at is required")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return SnapshotAdmissionResult{}, err
	}
	defer tx.Rollback()

	source, createdSource, err := getOrCreateSourceTx(ctx, tx, input)
	if err != nil {
		return SnapshotAdmissionResult{}, err
	}

	existing, found, err := findSnapshotByHashTx(ctx, tx, source.ID, input.SHA256)
	if err != nil {
		return SnapshotAdmissionResult{}, err
	}
	if found {
		if err := tx.Commit(); err != nil {
			return SnapshotAdmissionResult{}, err
		}
		return SnapshotAdmissionResult{
			Source:        source,
			Snapshot:      existing,
			Status:        "existing_snapshot",
			CreatedSource: createdSource,
		}, nil
	}

	latest, hasLatest, err := getLatestSnapshotTx(ctx, tx, source.ID)
	if err != nil {
		return SnapshotAdmissionResult{}, err
	}

	supersedesID := (*string)(nil)
	status := "created_snapshot"
	if hasLatest && latest.SHA256 != input.SHA256 {
		supersedesID = &latest.ID
		status = "new_revision"
	}

	newSnapshot := domain.NewSourceSnapshot{
		ID:                   domain.NewSnapshotID(input.ObservedAt),
		SourceID:             source.ID,
		SHA256:               input.SHA256,
		ByteLength:           input.ByteLength,
		MediaType:            input.MediaType,
		ObjectPath:           input.ObjectPath,
		ObservedAt:           input.ObservedAt,
		IngestRunID:          input.IngestRunID,
		SupersedesSnapshotID: supersedesID,
	}
	if err := newSnapshot.Validate(); err != nil {
		return SnapshotAdmissionResult{}, err
	}

	if _, err := tx.ExecContext(ctx, `
INSERT INTO source_snapshots(id, source_id, sha256, byte_length, media_type, object_path, observed_at, ingest_run_id, supersedes_snapshot_id)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)
`, newSnapshot.ID, newSnapshot.SourceID, newSnapshot.SHA256, newSnapshot.ByteLength, newSnapshot.MediaType, newSnapshot.ObjectPath, formatTime(newSnapshot.ObservedAt), newSnapshot.IngestRunID, newSnapshot.SupersedesSnapshotID); err != nil {
		if isDuplicateSnapshotError(err) {
			existing, found, findErr := findSnapshotByHashTx(ctx, tx, source.ID, input.SHA256)
			if findErr != nil {
				return SnapshotAdmissionResult{}, findErr
			}
			if found {
				if err := tx.Commit(); err != nil {
					return SnapshotAdmissionResult{}, err
				}
				return SnapshotAdmissionResult{
					Source:        source,
					Snapshot:      existing,
					Status:        "existing_snapshot",
					CreatedSource: createdSource,
				}, nil
			}
		}
		return SnapshotAdmissionResult{}, err
	}

	createdSnapshot := domain.SourceSnapshot{
		ID:                   newSnapshot.ID,
		SourceID:             newSnapshot.SourceID,
		SHA256:               newSnapshot.SHA256,
		ByteLength:           newSnapshot.ByteLength,
		MediaType:            newSnapshot.MediaType,
		ObjectPath:           newSnapshot.ObjectPath,
		ObservedAt:           newSnapshot.ObservedAt.UTC(),
		IngestRunID:          newSnapshot.IngestRunID,
		SupersedesSnapshotID: newSnapshot.SupersedesSnapshotID,
	}
	if err := tx.Commit(); err != nil {
		return SnapshotAdmissionResult{}, err
	}
	return SnapshotAdmissionResult{
		Source:        source,
		Snapshot:      createdSnapshot,
		Status:        status,
		CreatedSource: createdSource,
	}, nil
}

func (s *Store) ListSourcesByProblem(ctx context.Context, problemID string) ([]domain.Source, error) {
	if err := domain.ValidateProblemID(problemID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT id, problem_id, kind, logical_name, origin, created_at
FROM sources
WHERE problem_id = ?
ORDER BY created_at ASC, id ASC
`, problemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Source
	for rows.Next() {
		source, err := scanSource(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, source)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Store) GetSource(ctx context.Context, sourceID string) (domain.Source, error) {
	if err := domain.ValidateSourceID(sourceID); err != nil {
		return domain.Source{}, err
	}
	row := s.db.QueryRowContext(ctx, `
SELECT id, problem_id, kind, logical_name, origin, created_at
FROM sources
WHERE id = ?
`, sourceID)
	source, err := scanSource(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Source{}, fmt.Errorf("%w: source %s", ErrNotFound, sourceID)
		}
		return domain.Source{}, err
	}
	return source, nil
}

func (s *Store) ListSourcesWithSnapshotStats(ctx context.Context, problemID string) ([]SourceSummary, error) {
	if err := domain.ValidateProblemID(problemID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT
  s.id,
  s.problem_id,
  s.kind,
  s.logical_name,
  s.origin,
  s.created_at,
  (
    SELECT COUNT(1)
    FROM source_snapshots ss_count
    WHERE ss_count.source_id = s.id
  ) AS snapshot_count,
  (
    SELECT ss_latest.id
    FROM source_snapshots ss_latest
    WHERE ss_latest.source_id = s.id
    ORDER BY ss_latest.observed_at DESC, ss_latest.id DESC
    LIMIT 1
  ) AS latest_snapshot_id
FROM sources s
WHERE s.problem_id = ?
ORDER BY s.created_at ASC, s.id ASC
`, problemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []SourceSummary
	for rows.Next() {
		var source domain.Source
		var kind string
		var createdAt string
		var latestSnapshotID sql.NullString
		var snapshotCount int
		if err := rows.Scan(&source.ID, &source.ProblemID, &kind, &source.LogicalName, &source.Origin, &createdAt, &snapshotCount, &latestSnapshotID); err != nil {
			return nil, err
		}
		parsedCreatedAt, err := parseTime(createdAt)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrCorruptStore, err)
		}
		source.Kind = domain.SourceKind(kind)
		source.CreatedAt = parsedCreatedAt
		summary := SourceSummary{
			Source:        source,
			SnapshotCount: snapshotCount,
		}
		if latestSnapshotID.Valid {
			summary.LatestSnapshotID = &latestSnapshotID.String
		}
		summaries = append(summaries, summary)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return summaries, nil
}

func (s *Store) GetSourceSnapshot(ctx context.Context, snapshotID string) (domain.SourceSnapshot, error) {
	if err := domain.ValidateSnapshotID(snapshotID); err != nil {
		return domain.SourceSnapshot{}, err
	}
	row := s.db.QueryRowContext(ctx, `
SELECT id, source_id, sha256, byte_length, media_type, object_path, observed_at, ingest_run_id, supersedes_snapshot_id
FROM source_snapshots
WHERE id = ?
`, snapshotID)
	snapshot, err := scanSourceSnapshot(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.SourceSnapshot{}, fmt.Errorf("%w: source snapshot %s", ErrNotFound, snapshotID)
		}
		return domain.SourceSnapshot{}, err
	}
	return snapshot, nil
}

func (s *Store) ListSourceSnapshots(ctx context.Context, sourceID string) ([]domain.SourceSnapshot, error) {
	if err := domain.ValidateSourceID(sourceID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT id, source_id, sha256, byte_length, media_type, object_path, observed_at, ingest_run_id, supersedes_snapshot_id
FROM source_snapshots
WHERE source_id = ?
ORDER BY observed_at DESC, id DESC
`, sourceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snapshots []domain.SourceSnapshot
	for rows.Next() {
		snapshot, err := scanSourceSnapshot(rows)
		if err != nil {
			return nil, err
		}
		snapshots = append(snapshots, snapshot)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return snapshots, nil
}

func getOrCreateSourceTx(ctx context.Context, tx *sql.Tx, input SnapshotAdmission) (domain.Source, bool, error) {
	row := tx.QueryRowContext(ctx, `
SELECT id, problem_id, kind, logical_name, origin, created_at
FROM sources
WHERE problem_id = ? AND origin = ?
`, input.ProblemID, input.Origin)
	source, err := scanSource(row)
	if err == nil {
		return source, false, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return domain.Source{}, false, err
	}

	newSource := domain.NewSource{
		ID:          domain.NewSourceID(input.ObservedAt),
		ProblemID:   input.ProblemID,
		Kind:        input.Kind,
		LogicalName: input.LogicalName,
		Origin:      input.Origin,
		CreatedAt:   input.ObservedAt,
	}
	if err := newSource.Validate(); err != nil {
		return domain.Source{}, false, err
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO sources(id, problem_id, kind, logical_name, origin, created_at)
VALUES(?, ?, ?, ?, ?, ?)
`, newSource.ID, newSource.ProblemID, newSource.Kind, newSource.LogicalName, newSource.Origin, formatTime(newSource.CreatedAt)); err != nil {
		if !isDuplicateSourceOriginError(err) {
			return domain.Source{}, false, err
		}
		row := tx.QueryRowContext(ctx, `
SELECT id, problem_id, kind, logical_name, origin, created_at
FROM sources
WHERE problem_id = ? AND origin = ?
`, input.ProblemID, input.Origin)
		source, scanErr := scanSource(row)
		return source, false, scanErr
	}

	return domain.Source{
		ID:          newSource.ID,
		ProblemID:   newSource.ProblemID,
		Kind:        newSource.Kind,
		LogicalName: newSource.LogicalName,
		Origin:      newSource.Origin,
		CreatedAt:   newSource.CreatedAt.UTC(),
	}, true, nil
}

func getLatestSnapshotTx(ctx context.Context, tx *sql.Tx, sourceID string) (domain.SourceSnapshot, bool, error) {
	row := tx.QueryRowContext(ctx, `
SELECT id, source_id, sha256, byte_length, media_type, object_path, observed_at, ingest_run_id, supersedes_snapshot_id
FROM source_snapshots
WHERE source_id = ?
ORDER BY observed_at DESC, id DESC
LIMIT 1
`, sourceID)
	snapshot, err := scanSourceSnapshot(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.SourceSnapshot{}, false, nil
		}
		return domain.SourceSnapshot{}, false, err
	}
	return snapshot, true, nil
}

func findSnapshotByHashTx(ctx context.Context, tx *sql.Tx, sourceID, hash string) (domain.SourceSnapshot, bool, error) {
	row := tx.QueryRowContext(ctx, `
SELECT id, source_id, sha256, byte_length, media_type, object_path, observed_at, ingest_run_id, supersedes_snapshot_id
FROM source_snapshots
WHERE source_id = ? AND sha256 = ?
`, sourceID, hash)
	snapshot, err := scanSourceSnapshot(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.SourceSnapshot{}, false, nil
		}
		return domain.SourceSnapshot{}, false, err
	}
	return snapshot, true, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanProblem(row scanner) (domain.Problem, error) {
	var problem domain.Problem
	var createdAt string
	var status string
	if err := row.Scan(&problem.ID, &problem.Slug, &problem.Statement, &status, &createdAt, &problem.CreatedByRunID); err != nil {
		return domain.Problem{}, err
	}

	parsedCreatedAt, err := parseTime(createdAt)
	if err != nil {
		return domain.Problem{}, fmt.Errorf("%w: %v", ErrCorruptStore, err)
	}

	problem.Status = domain.ProblemStatus(status)
	problem.CreatedAt = parsedCreatedAt
	return problem, nil
}

func scanRun(row scanner) (domain.Run, error) {
	var run domain.Run
	var parentRunID sql.NullString
	var errorSummary sql.NullString
	var startedAt string
	var completedAt string
	var status string
	if err := row.Scan(&run.ID, &run.ProblemID, &parentRunID, &run.Operation, &status, &run.InputRef, &run.ToolName, &run.ToolVersion, &startedAt, &completedAt, &errorSummary); err != nil {
		return domain.Run{}, err
	}

	parsedStartedAt, err := parseTime(startedAt)
	if err != nil {
		return domain.Run{}, fmt.Errorf("%w: %v", ErrCorruptStore, err)
	}
	parsedCompletedAt, err := parseTime(completedAt)
	if err != nil {
		return domain.Run{}, fmt.Errorf("%w: %v", ErrCorruptStore, err)
	}

	run.Status = domain.RunStatus(status)
	run.StartedAt = parsedStartedAt
	run.CompletedAt = parsedCompletedAt
	if parentRunID.Valid {
		run.ParentRunID = &parentRunID.String
	}
	if errorSummary.Valid {
		run.ErrorSummary = &errorSummary.String
	}
	return run, nil
}

func scanSource(row scanner) (domain.Source, error) {
	var source domain.Source
	var kind string
	var createdAt string
	if err := row.Scan(&source.ID, &source.ProblemID, &kind, &source.LogicalName, &source.Origin, &createdAt); err != nil {
		return domain.Source{}, err
	}
	parsedCreatedAt, err := parseTime(createdAt)
	if err != nil {
		return domain.Source{}, fmt.Errorf("%w: %v", ErrCorruptStore, err)
	}
	source.Kind = domain.SourceKind(kind)
	source.CreatedAt = parsedCreatedAt
	return source, nil
}

func scanSourceSnapshot(row scanner) (domain.SourceSnapshot, error) {
	var snapshot domain.SourceSnapshot
	var observedAt string
	var supersedesID sql.NullString
	if err := row.Scan(&snapshot.ID, &snapshot.SourceID, &snapshot.SHA256, &snapshot.ByteLength, &snapshot.MediaType, &snapshot.ObjectPath, &observedAt, &snapshot.IngestRunID, &supersedesID); err != nil {
		return domain.SourceSnapshot{}, err
	}
	parsedObservedAt, err := parseTime(observedAt)
	if err != nil {
		return domain.SourceSnapshot{}, fmt.Errorf("%w: %v", ErrCorruptStore, err)
	}
	snapshot.ObservedAt = parsedObservedAt
	if supersedesID.Valid {
		snapshot.SupersedesSnapshotID = &supersedesID.String
	}
	return snapshot, nil
}

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func parseTime(raw string) (time.Time, error) {
	return time.Parse(time.RFC3339Nano, raw)
}

func isDuplicateSlugError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed: problems.slug")
}

func isDuplicateSourceOriginError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed: sources.problem_id, sources.origin")
}

func isDuplicateSnapshotError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed: source_snapshots.source_id, source_snapshots.sha256")
}

func isKnownMigrationVersion(version int) bool {
	for _, migration := range migrations {
		if migration.version == version {
			return true
		}
	}
	return false
}

func validateSchemaTables(ctx context.Context, tx *sql.Tx) error {
	for _, table := range []string{
		"problems", "runs", "sources", "source_snapshots",
		"provider_invocations", "approaches", "normalization_revisions",
		"approach_revisions", "mechanisms", "mechanism_attributes",
		"mechanism_field_completeness",
		"outcomes", "failure_boundaries", "source_supports",
		"canonical_vocabulary", "canonical_terms", "canonical_term_aliases",
		"canonical_rejected_terms",
		"classification_rubrics", "mechanism_signatures", "signature_field_claims",
		"signature_postures", "signature_boundaries", "signature_outcomes",
		"comparison_runs", "comparison_field_results",
		"cluster_runs", "mechanism_clusters", "cluster_members",
		"cluster_distances", "cluster_coverage_axes", "cluster_discrimination_losses",
		"failure_spaces", "failure_space_outcomes", "failure_space_axes",
		"invariant_revisions", "candidate_invariants", "invariant_predicates",
		"invariant_family_evaluations", "invariant_counterexamples",
		"invariant_challenges", "invariant_transition_counters",
		"invariant_state_transitions", "invariant_challenge_evidence",
		"synthetic_artifacts", "invariant_challenge_synthetic_artifacts",
		"invariant_lineage",
		"frontier_generation_runs", "frontier_proposals",
		"frontier_target_invariants", "frontier_nearest_clusters",
		"evaluation_runs", "evaluations", "evaluation_metrics",
		"evaluation_run_metrics", "evaluated_failures",
		"holdout_sets", "holdout_set_sources", "holdout_source_dating",
		"leakage_checks", "experiment_runs", "experiment_arms", "experiment_metrics",
		"experiment_arm_proposals", "experiment_targets",
		"frontier_proposal_signatures", "frontier_proposal_signature_revisions", "frontier_generation_contents", "success_invariant_revisions",
		"success_invariants", "success_invariant_predicates",
		"success_invariant_broken_targets", "success_invariant_cohort_evaluations",
		"search_policy_revisions", "search_policy_directives",
		"search_policy_provenance", "frontier_generation_policy",
	} {
		row := tx.QueryRowContext(ctx, `
SELECT EXISTS(
  SELECT 1
  FROM sqlite_master
  WHERE type = 'table' AND name = ?
)
`, table)
		var exists bool
		if err := row.Scan(&exists); err != nil {
			return fmt.Errorf("%w: %v", ErrCorruptStore, err)
		}
		if !exists {
			return fmt.Errorf("%w: missing table %q for current schema", ErrCorruptStore, table)
		}
	}
	// Column-level checks for the newest schema shape. A table can exist while a
	// later migration that only ADDs columns (rather than a new table) was
	// skipped, leaving an incomplete upgrade that a table-existence check alone
	// would miss. Assert the columns owned by the most recent migration so a
	// partial upgrade is reported as corrupt rather than silently accepted.
	for _, c := range []struct{ table, column string }{
		{"cluster_runs", "input_set_hash"},      // migration v10 (KTD-1)
		{"mechanism_clusters", "outcome_class"}, // migration v10 (KTD-9)
		{"mechanism_clusters", "outcome_mixed"}, // migration v10 (KTD-9)
	} {
		has, err := columnExists(ctx, tx, c.table, c.column)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrCorruptStore, err)
		}
		if !has {
			return fmt.Errorf("%w: missing column %q on %q for current schema", ErrCorruptStore, c.column, c.table)
		}
	}
	// A column probe cannot see a CHECK constraint, so a database left on the
	// v3 normalize-only provider_invocations.role CHECK would pass the checks
	// above while silently rejecting invariant-mining or challenge invocations.
	// Assert the v11 ('invariant'), v13 ('challenge'), v14 ('generate'), v15
	// ('evaluate'), v17 ('success-compress'), and v19 ('policy-mutate') role
	// generalizations by reading the table DDL directly.
	for _, role := range []string{"invariant", "challenge", "generate", "evaluate", "success-compress", "policy-mutate"} {
		allows, err := providerRoleAllows(ctx, tx, role)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrCorruptStore, err)
		}
		if !allows {
			return fmt.Errorf("%w: provider_invocations.role CHECK does not permit %q for current schema", ErrCorruptStore, role)
		}
	}
	return nil
}
