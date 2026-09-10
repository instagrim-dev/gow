package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/instagrim-dev/newf/internal/domain"
)

var (
	ErrNotFound     = errors.New("record not found")
	ErrMigration    = errors.New("migration failed")
	ErrCorruptStore = errors.New("corrupt store")
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
		if _, err := tx.ExecContext(ctx, migration.sql); err != nil {
			return fmt.Errorf("%w: %v", ErrMigration, err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations(version, applied_at) VALUES(?, ?)`, migration.version, now); err != nil {
			return fmt.Errorf("%w: %v", ErrMigration, err)
		}
	}

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
	slug := base
	for i := 2; ; i++ {
		_, found, err := s.FindProblemBySlug(ctx, slug)
		if err != nil {
			return "", err
		}
		if !found {
			return slug, nil
		}
		slug = fmt.Sprintf("%s-%d", base, i)
	}
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

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}

func parseTime(raw string) (time.Time, error) {
	return time.Parse(time.RFC3339Nano, raw)
}
