package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/instagrim-dev/newf/internal/domain"
)

// FailureSpaceOutcomeRow is the family count for one outcome class.
type FailureSpaceOutcomeRow struct {
	OutcomeClass string
	FamilyCount  int
}

// FailureSpaceAxisRow is one coverage axis of a failure space.
type FailureSpaceAxisRow struct {
	AxisKind           string
	DistinctValueCount int
	UnderSampled       bool
}

// FailureSpaceRecord is a materialized failure-space revision.
type FailureSpaceRecord struct {
	ID                   string
	ProblemID            string
	ClusterRunID         string
	RunID                string
	Revision             int
	DistinctFamilyCount  int
	RedundantMemberCount int
	CreatedAt            string
	Outcomes             []FailureSpaceOutcomeRow
	Axes                 []FailureSpaceAxisRow
}

// PersistFailureSpaceResult reports the persisted record and whether it was new.
type PersistFailureSpaceResult struct {
	Record  FailureSpaceRecord
	Created bool
}

// PersistFailureSpace materializes a failure space for a cluster run. It is
// idempotent on (problem, cluster_run): re-materializing an existing cluster
// run returns the existing record. A new cluster run yields the next revision
// for the problem, never a rewrite (rows are immutable by trigger).
func (s *Store) PersistFailureSpace(ctx context.Context, record FailureSpaceRecord) (PersistFailureSpaceResult, error) {
	if err := domain.ValidateFailureSpaceID(record.ID); err != nil {
		return PersistFailureSpaceResult{}, err
	}
	if err := domain.ValidateProblemID(record.ProblemID); err != nil {
		return PersistFailureSpaceResult{}, err
	}
	if err := domain.ValidateClusterRunID(record.ClusterRunID); err != nil {
		return PersistFailureSpaceResult{}, err
	}

	if existingID, found, err := s.findFailureSpace(ctx, record.ProblemID, record.ClusterRunID); err != nil {
		return PersistFailureSpaceResult{}, err
	} else if found {
		full, lerr := s.loadFailureSpace(ctx, existingID)
		if lerr != nil {
			return PersistFailureSpaceResult{}, lerr
		}
		return PersistFailureSpaceResult{Record: full, Created: false}, nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return PersistFailureSpaceResult{}, err
	}
	defer tx.Rollback()

	// Next revision for the problem.
	var maxRev sql.NullInt64
	if err := tx.QueryRowContext(ctx, `SELECT MAX(revision) FROM failure_spaces WHERE problem_id = ?`, record.ProblemID).Scan(&maxRev); err != nil {
		return PersistFailureSpaceResult{}, err
	}
	record.Revision = int(maxRev.Int64) + 1

	if _, err := tx.ExecContext(ctx, `
INSERT INTO failure_spaces(id, problem_id, cluster_run_id, run_id, revision, distinct_family_count, redundant_member_count, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?)
`, record.ID, record.ProblemID, record.ClusterRunID, record.RunID, record.Revision, record.DistinctFamilyCount, record.RedundantMemberCount, record.CreatedAt); err != nil {
		return PersistFailureSpaceResult{}, err
	}
	for _, o := range record.Outcomes {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO failure_space_outcomes(failure_space_id, outcome_class, family_count) VALUES(?, ?, ?)
`, record.ID, o.OutcomeClass, o.FamilyCount); err != nil {
			return PersistFailureSpaceResult{}, err
		}
	}
	for _, ax := range record.Axes {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO failure_space_axes(failure_space_id, axis_kind, distinct_value_count, under_sampled) VALUES(?, ?, ?, ?)
`, record.ID, ax.AxisKind, ax.DistinctValueCount, boolToInt(ax.UnderSampled)); err != nil {
			return PersistFailureSpaceResult{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return PersistFailureSpaceResult{}, err
	}
	return PersistFailureSpaceResult{Record: record, Created: true}, nil
}

func (s *Store) findFailureSpace(ctx context.Context, problemID, clusterRunID string) (string, bool, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id FROM failure_spaces WHERE problem_id = ? AND cluster_run_id = ?`, problemID, clusterRunID)
	var id string
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}
		return "", false, err
	}
	return id, true, nil
}

// GetFailureSpace loads a full failure space by id.
func (s *Store) GetFailureSpace(ctx context.Context, id string) (FailureSpaceRecord, error) {
	if err := domain.ValidateFailureSpaceID(id); err != nil {
		return FailureSpaceRecord{}, err
	}
	return s.loadFailureSpace(ctx, id)
}

func (s *Store) loadFailureSpace(ctx context.Context, id string) (FailureSpaceRecord, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, problem_id, cluster_run_id, run_id, revision, distinct_family_count, redundant_member_count, created_at
FROM failure_spaces WHERE id = ?
`, id)
	var rec FailureSpaceRecord
	if err := row.Scan(&rec.ID, &rec.ProblemID, &rec.ClusterRunID, &rec.RunID, &rec.Revision, &rec.DistinctFamilyCount, &rec.RedundantMemberCount, &rec.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return FailureSpaceRecord{}, fmt.Errorf("%w: failure space %s", ErrNotFound, id)
		}
		return FailureSpaceRecord{}, err
	}

	outRows, err := s.db.QueryContext(ctx, `SELECT outcome_class, family_count FROM failure_space_outcomes WHERE failure_space_id = ? ORDER BY outcome_class`, id)
	if err != nil {
		return FailureSpaceRecord{}, err
	}
	defer outRows.Close()
	for outRows.Next() {
		var o FailureSpaceOutcomeRow
		if err := outRows.Scan(&o.OutcomeClass, &o.FamilyCount); err != nil {
			return FailureSpaceRecord{}, err
		}
		rec.Outcomes = append(rec.Outcomes, o)
	}
	if err := outRows.Err(); err != nil {
		return FailureSpaceRecord{}, err
	}

	axRows, err := s.db.QueryContext(ctx, `SELECT axis_kind, distinct_value_count, under_sampled FROM failure_space_axes WHERE failure_space_id = ? ORDER BY axis_kind`, id)
	if err != nil {
		return FailureSpaceRecord{}, err
	}
	defer axRows.Close()
	for axRows.Next() {
		var ax FailureSpaceAxisRow
		var under int
		if err := axRows.Scan(&ax.AxisKind, &ax.DistinctValueCount, &under); err != nil {
			return FailureSpaceRecord{}, err
		}
		ax.UnderSampled = under != 0
		rec.Axes = append(rec.Axes, ax)
	}
	if err := axRows.Err(); err != nil {
		return FailureSpaceRecord{}, err
	}
	return rec, nil
}

// LatestFailureSpace returns the highest-revision failure space for a problem.
func (s *Store) LatestFailureSpace(ctx context.Context, problemID string) (FailureSpaceRecord, bool, error) {
	if err := domain.ValidateProblemID(problemID); err != nil {
		return FailureSpaceRecord{}, false, err
	}
	row := s.db.QueryRowContext(ctx, `SELECT id FROM failure_spaces WHERE problem_id = ? ORDER BY revision DESC LIMIT 1`, problemID)
	var id string
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return FailureSpaceRecord{}, false, nil
		}
		return FailureSpaceRecord{}, false, err
	}
	rec, err := s.loadFailureSpace(ctx, id)
	if err != nil {
		return FailureSpaceRecord{}, false, err
	}
	return rec, true, nil
}
