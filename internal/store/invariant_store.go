package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/instagrim-dev/newf/internal/domain"
)

// InvariantProviderInvocation is the mining provider invocation recorded in the
// shared provider_invocations table with role='invariant' (permitted after the
// v11 role generalization).
type InvariantProviderInvocation struct {
	ID              string
	RunID           string
	ProviderName    string
	ProviderVersion string
	ModelName       string
	SchemaVersion   string
	RequestHash     string
	RequestPayload  string
	ResponsePayload string
	CreatedAt       string
}

// InvariantFamilyEvaluationRow is one persisted per-family evaluation.
type InvariantFamilyEvaluationRow struct {
	ClusterID     string
	OutcomeClass  string
	Role          string // 'support' | 'contrast'
	Verdict       string // 'satisfies'|'violates'|'unknown'|'member_mixed'
	ExplicitCount int
	InferredCount int
	OtherCount    int
}

// InvariantCounterexampleRow is one persisted counterexample.
type InvariantCounterexampleRow struct {
	ClusterID string
	Reason    string
}

// CandidateInvariantRow is one persisted candidate invariant (proposed only).
type CandidateInvariantRow struct {
	ID                           string
	PredicateFingerprint         string
	PredicateJSON                string
	Statement                    string
	AbstractionLevel             string
	AssociationStatus            string
	ObstructionIsModelHypothesis bool
	DistinctFamilySupport        int
	FailureCoverageNum           int
	FailureCoverageDen           int
	ContrastViolatingNum         int
	ContrastEligibleDen          int
	SupportExplicitCount         int
	SupportInferredCount         int
	SupportOtherCount            int
	ConfidenceOrdinal            string
	Ordinal                      int
	FamilyEvaluations            []InvariantFamilyEvaluationRow
	Counterexamples              []InvariantCounterexampleRow
}

// InvariantRevisionRecord is the full persisted mining pass.
type InvariantRevisionRecord struct {
	ID                   string
	ProblemID            string
	FailureSpaceID       string
	ClusterRunID         string
	RunID                string
	ProviderInvocationID string
	MinerVersion         string
	PredicateSchema      string
	MinSupport           int
	Revision             int
	CandidateCount       int
	CreatedAt            string
	Invocation           InvariantProviderInvocation
	Candidates           []CandidateInvariantRow
}

// PersistInvariantRevisionResult reports the persisted revision and newness.
type PersistInvariantRevisionResult struct {
	Record  InvariantRevisionRecord
	Created bool
}

// PersistInvariantRevision writes a mining pass transactionally. It is
// idempotent on (problem, failure_space, miner_version, predicate_schema,
// min_support): an existing revision is returned unchanged with Created=false,
// never rewritten (rows are immutable by trigger). A new tuple is assigned the
// next monotonic revision for the problem and records the provider invocation
// with role='invariant'.
func (s *Store) PersistInvariantRevision(ctx context.Context, record InvariantRevisionRecord) (PersistInvariantRevisionResult, error) {
	if err := domain.ValidateInvariantRevisionID(record.ID); err != nil {
		return PersistInvariantRevisionResult{}, err
	}
	if err := domain.ValidateProblemID(record.ProblemID); err != nil {
		return PersistInvariantRevisionResult{}, err
	}
	if err := domain.ValidateFailureSpaceID(record.FailureSpaceID); err != nil {
		return PersistInvariantRevisionResult{}, err
	}

	if existingID, found, err := s.findInvariantRevision(ctx, record); err != nil {
		return PersistInvariantRevisionResult{}, err
	} else if found {
		full, lerr := s.loadInvariantRevision(ctx, existingID)
		if lerr != nil {
			return PersistInvariantRevisionResult{}, lerr
		}
		return PersistInvariantRevisionResult{Record: full, Created: false}, nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return PersistInvariantRevisionResult{}, err
	}
	defer tx.Rollback()

	record, err = writeInvariantRevisionTx(ctx, tx, record)
	if err != nil {
		return PersistInvariantRevisionResult{}, err
	}

	if err := tx.Commit(); err != nil {
		return PersistInvariantRevisionResult{}, err
	}
	return PersistInvariantRevisionResult{Record: record, Created: true}, nil
}

// writeInvariantRevisionTx writes a mining revision inside an existing
// transaction (numbering + invocation + revision + candidates). It is shared
// by PersistInvariantRevision and the challenge layer, which persists confirmed
// split/merge children INSIDE the campaign transaction so a failing campaign
// leaves no orphaned derived candidates.
func writeInvariantRevisionTx(ctx context.Context, tx *sql.Tx, record InvariantRevisionRecord) (InvariantRevisionRecord, error) {
	var maxRev sql.NullInt64
	if err := tx.QueryRowContext(ctx, `SELECT MAX(revision) FROM invariant_revisions WHERE problem_id = ?`, record.ProblemID).Scan(&maxRev); err != nil {
		return InvariantRevisionRecord{}, err
	}
	record.Revision = int(maxRev.Int64) + 1

	inv := record.Invocation
	if _, err := tx.ExecContext(ctx, `
INSERT INTO provider_invocations(id, run_id, role, provider_name, provider_version, model_name, schema_version, request_hash, request_payload, response_payload, created_at)
VALUES(?, ?, 'invariant', ?, ?, ?, ?, ?, ?, ?, ?)
`, inv.ID, inv.RunID, inv.ProviderName, inv.ProviderVersion, inv.ModelName, inv.SchemaVersion, inv.RequestHash, inv.RequestPayload, inv.ResponsePayload, inv.CreatedAt); err != nil {
		return InvariantRevisionRecord{}, err
	}

	if _, err := tx.ExecContext(ctx, `
INSERT INTO invariant_revisions(id, problem_id, failure_space_id, cluster_run_id, run_id, provider_invocation_id, miner_version, predicate_schema, min_support, revision, candidate_count, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, record.ID, record.ProblemID, record.FailureSpaceID, record.ClusterRunID, record.RunID, inv.ID, record.MinerVersion, record.PredicateSchema, record.MinSupport, record.Revision, record.CandidateCount, record.CreatedAt); err != nil {
		return InvariantRevisionRecord{}, err
	}

	for _, c := range record.Candidates {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO candidate_invariants(id, invariant_revision_id, predicate_fingerprint, statement, abstraction_level, association_status, obstruction_is_model_hypothesis, distinct_family_support, failure_coverage_num, failure_coverage_den, contrast_violating_num, contrast_eligible_den, support_explicit_count, support_inferred_count, support_other_count, confidence_ordinal, ordinal)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, c.ID, record.ID, c.PredicateFingerprint, c.Statement, c.AbstractionLevel, c.AssociationStatus, boolToInt(c.ObstructionIsModelHypothesis), c.DistinctFamilySupport, c.FailureCoverageNum, c.FailureCoverageDen, c.ContrastViolatingNum, c.ContrastEligibleDen, c.SupportExplicitCount, c.SupportInferredCount, c.SupportOtherCount, c.ConfidenceOrdinal, c.Ordinal); err != nil {
			return InvariantRevisionRecord{}, err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO invariant_predicates(invariant_id, predicate_json) VALUES(?, ?)`, c.ID, c.PredicateJSON); err != nil {
			return InvariantRevisionRecord{}, err
		}
		for _, fe := range c.FamilyEvaluations {
			if _, err := tx.ExecContext(ctx, `
INSERT INTO invariant_family_evaluations(invariant_id, cluster_id, outcome_class, role, verdict, explicit_count, inferred_count, other_count)
VALUES(?, ?, ?, ?, ?, ?, ?, ?)
`, c.ID, fe.ClusterID, fe.OutcomeClass, fe.Role, fe.Verdict, fe.ExplicitCount, fe.InferredCount, fe.OtherCount); err != nil {
				return InvariantRevisionRecord{}, err
			}
		}
		for _, ce := range c.Counterexamples {
			if _, err := tx.ExecContext(ctx, `
INSERT INTO invariant_counterexamples(invariant_id, cluster_id, reason) VALUES(?, ?, ?)
`, c.ID, ce.ClusterID, ce.Reason); err != nil {
				return InvariantRevisionRecord{}, err
			}
		}
	}
	return record, nil
}

func (s *Store) findInvariantRevision(ctx context.Context, record InvariantRevisionRecord) (string, bool, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id FROM invariant_revisions
WHERE problem_id = ? AND failure_space_id = ? AND miner_version = ? AND predicate_schema = ? AND min_support = ?
`, record.ProblemID, record.FailureSpaceID, record.MinerVersion, record.PredicateSchema, record.MinSupport)
	var id string
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}
		return "", false, err
	}
	return id, true, nil
}

// InvariantReuseKey is the complete identity that keys revision reuse: the
// problem/failure-space plus the full miner identity (folded into miner_version),
// the predicate schema, and the support threshold. Callers compute it BEFORE
// invoking a provider so an identical request reuses the prior revision without
// re-running (F5).
type InvariantReuseKey struct {
	ProblemID       string
	FailureSpaceID  string
	MinerVersion    string
	PredicateSchema string
	MinSupport      int
}

// LookupInvariantRevision returns the existing revision for a reuse key, if any,
// without side effects. It lets the pipeline perform a reuse-check-first before
// any (potentially costly) provider invocation.
func (s *Store) LookupInvariantRevision(ctx context.Context, key InvariantReuseKey) (InvariantRevisionRecord, bool, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id FROM invariant_revisions
WHERE problem_id = ? AND failure_space_id = ? AND miner_version = ? AND predicate_schema = ? AND min_support = ?
`, key.ProblemID, key.FailureSpaceID, key.MinerVersion, key.PredicateSchema, key.MinSupport)
	var id string
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return InvariantRevisionRecord{}, false, nil
		}
		return InvariantRevisionRecord{}, false, err
	}
	rec, err := s.loadInvariantRevision(ctx, id)
	if err != nil {
		return InvariantRevisionRecord{}, false, err
	}
	return rec, true, nil
}

// GetInvariantRevision loads a full mining pass by id.
func (s *Store) GetInvariantRevision(ctx context.Context, id string) (InvariantRevisionRecord, error) {
	if err := domain.ValidateInvariantRevisionID(id); err != nil {
		return InvariantRevisionRecord{}, err
	}
	return s.loadInvariantRevision(ctx, id)
}

func (s *Store) loadInvariantRevision(ctx context.Context, id string) (InvariantRevisionRecord, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, problem_id, failure_space_id, cluster_run_id, run_id, provider_invocation_id, miner_version, predicate_schema, min_support, revision, candidate_count, created_at
FROM invariant_revisions WHERE id = ?
`, id)
	var rec InvariantRevisionRecord
	if err := row.Scan(&rec.ID, &rec.ProblemID, &rec.FailureSpaceID, &rec.ClusterRunID, &rec.RunID, &rec.ProviderInvocationID, &rec.MinerVersion, &rec.PredicateSchema, &rec.MinSupport, &rec.Revision, &rec.CandidateCount, &rec.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return InvariantRevisionRecord{}, fmt.Errorf("%w: invariant revision %s", ErrNotFound, id)
		}
		return InvariantRevisionRecord{}, err
	}
	candRows, err := s.db.QueryContext(ctx, `
SELECT id, predicate_fingerprint, statement, abstraction_level, association_status, obstruction_is_model_hypothesis, distinct_family_support, failure_coverage_num, failure_coverage_den, contrast_violating_num, contrast_eligible_den, support_explicit_count, support_inferred_count, support_other_count, confidence_ordinal, ordinal
FROM candidate_invariants WHERE invariant_revision_id = ? ORDER BY ordinal
`, id)
	if err != nil {
		return InvariantRevisionRecord{}, err
	}
	defer candRows.Close()
	for candRows.Next() {
		var c CandidateInvariantRow
		var obstruction int
		if err := candRows.Scan(&c.ID, &c.PredicateFingerprint, &c.Statement, &c.AbstractionLevel, &c.AssociationStatus, &obstruction, &c.DistinctFamilySupport, &c.FailureCoverageNum, &c.FailureCoverageDen, &c.ContrastViolatingNum, &c.ContrastEligibleDen, &c.SupportExplicitCount, &c.SupportInferredCount, &c.SupportOtherCount, &c.ConfidenceOrdinal, &c.Ordinal); err != nil {
			return InvariantRevisionRecord{}, err
		}
		c.ObstructionIsModelHypothesis = obstruction != 0
		rec.Candidates = append(rec.Candidates, c)
	}
	if err := candRows.Err(); err != nil {
		return InvariantRevisionRecord{}, err
	}
	for i := range rec.Candidates {
		if err := s.loadCandidateDetail(ctx, &rec.Candidates[i]); err != nil {
			return InvariantRevisionRecord{}, err
		}
	}
	return rec, nil
}

func (s *Store) loadCandidateDetail(ctx context.Context, c *CandidateInvariantRow) error {
	if err := s.db.QueryRowContext(ctx, `SELECT predicate_json FROM invariant_predicates WHERE invariant_id = ?`, c.ID).Scan(&c.PredicateJSON); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}
	}
	feRows, err := s.db.QueryContext(ctx, `
SELECT cluster_id, outcome_class, role, verdict, explicit_count, inferred_count, other_count
FROM invariant_family_evaluations WHERE invariant_id = ? ORDER BY cluster_id, role
`, c.ID)
	if err != nil {
		return err
	}
	defer feRows.Close()
	for feRows.Next() {
		var fe InvariantFamilyEvaluationRow
		if err := feRows.Scan(&fe.ClusterID, &fe.OutcomeClass, &fe.Role, &fe.Verdict, &fe.ExplicitCount, &fe.InferredCount, &fe.OtherCount); err != nil {
			return err
		}
		c.FamilyEvaluations = append(c.FamilyEvaluations, fe)
	}
	if err := feRows.Err(); err != nil {
		return err
	}
	ceRows, err := s.db.QueryContext(ctx, `SELECT cluster_id, reason FROM invariant_counterexamples WHERE invariant_id = ? ORDER BY cluster_id`, c.ID)
	if err != nil {
		return err
	}
	defer ceRows.Close()
	for ceRows.Next() {
		var ce InvariantCounterexampleRow
		if err := ceRows.Scan(&ce.ClusterID, &ce.Reason); err != nil {
			return err
		}
		c.Counterexamples = append(c.Counterexamples, ce)
	}
	return ceRows.Err()
}

// ListInvariantRevisions returns the revision headers for a problem, newest
// first (highest revision).
func (s *Store) ListInvariantRevisions(ctx context.Context, problemID string) ([]InvariantRevisionRecord, error) {
	if err := domain.ValidateProblemID(problemID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT id, problem_id, failure_space_id, cluster_run_id, run_id, provider_invocation_id, miner_version, predicate_schema, min_support, revision, candidate_count, created_at
FROM invariant_revisions WHERE problem_id = ? ORDER BY revision DESC, id DESC
`, problemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []InvariantRevisionRecord
	for rows.Next() {
		var rec InvariantRevisionRecord
		if err := rows.Scan(&rec.ID, &rec.ProblemID, &rec.FailureSpaceID, &rec.ClusterRunID, &rec.RunID, &rec.ProviderInvocationID, &rec.MinerVersion, &rec.PredicateSchema, &rec.MinSupport, &rec.Revision, &rec.CandidateCount, &rec.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

// LatestInvariantRevision returns the highest-revision mining pass id for a
// problem.
func (s *Store) LatestInvariantRevision(ctx context.Context, problemID string) (string, bool, error) {
	if err := domain.ValidateProblemID(problemID); err != nil {
		return "", false, err
	}
	row := s.db.QueryRowContext(ctx, `SELECT id FROM invariant_revisions WHERE problem_id = ? ORDER BY revision DESC, id DESC LIMIT 1`, problemID)
	var id string
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}
		return "", false, err
	}
	return id, true, nil
}
