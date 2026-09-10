package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/instagrim-dev/newf/internal/domain"
)

// BreakCohortRow is one (broken target, evaluated proposal) pair: the raw
// material of success compression. SignatureJSON is empty for pre-v17
// proposals whose canonical content was never persisted (they are ineligible
// for compression and must be COUNTED, never silently dropped).
type BreakCohortRow struct {
	TargetInvariantID string
	ProposalID        string
	Result            string // frontier_proposals.result (one-time set by M5.2)
	Strength          string // latest evaluation's verification_strength
	SignatureJSON     string
	Fingerprint       string
}

// ListBreakCohortRows returns every code-verified break (violated=1) whose
// proposal has been evaluated (result set), joined with the persisted
// canonical content and the latest evaluation's verification strength.
func (s *Store) ListBreakCohortRows(ctx context.Context, problemID string) ([]BreakCohortRow, error) {
	if err := domain.ValidateProblemID(problemID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT t.invariant_id, p.id, p.result,
       COALESCE((SELECT e.verification_strength FROM evaluations e WHERE e.proposal_id = p.id ORDER BY e.created_at DESC, e.id DESC LIMIT 1), ''),
       COALESCE(fps.signature_json, ''), COALESCE(fps.canonical_fingerprint, '')
FROM frontier_target_invariants t
JOIN frontier_proposals p ON p.id = t.proposal_id
LEFT JOIN frontier_proposal_signatures fps ON fps.proposal_id = p.id
WHERE p.problem_id = ? AND t.violated = 1 AND p.result IS NOT NULL
ORDER BY t.invariant_id, p.id
`, problemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []BreakCohortRow
	for rows.Next() {
		var r BreakCohortRow
		if err := rows.Scan(&r.TargetInvariantID, &r.ProposalID, &r.Result, &r.Strength, &r.SignatureJSON, &r.Fingerprint); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// SuccessCohortEvaluationRow is one persisted member verdict under a condition.
type SuccessCohortEvaluationRow struct {
	ProposalID string
	CohortRole string // progressor|non_progressor
	Verdict    string // satisfies|violates|unknown
	Strength   string
}

// SuccessInvariantRow is one persisted success invariant (proposed only).
type SuccessInvariantRow struct {
	ID                          string // sinv_
	PredicateFingerprint        string
	PredicateJSON               string
	Statement                   string
	AbstractionLevel            string
	CoverageNum                 int
	CoverageDen                 int
	ExclusionNum                int
	ExclusionDen                int
	CoverageOrdinal             string
	ExclusionOrdinal            string
	DistinctSupport             int
	StrengthDeterministic       int
	StrengthReproducible        int
	StrengthIndependentEvidence int
	StrengthIndependentCritic   int
	StrengthModelJudgment       int
	Ordinal                     int
	BrokenTargets               []string // candidate_invariants ids (the P links)
	CohortEvaluations           []SuccessCohortEvaluationRow
}

// SuccessRevisionRecord is one full compression pass. Idempotent on
// (problem, cohort_hash, compressor_version, predicate_schema, min_support).
type SuccessRevisionRecord struct {
	ID                    string // svr_
	ProblemID             string
	RunID                 string
	CompressorVersion     string
	PredicateSchema       string
	MinSupport            int
	CohortHash            string
	IneligibleUnpersisted int
	AmbiguousMembers      int
	// InadmissibleConditions counts provider-proposed conditions rejected by the
	// AdmitCandidate gate (grammar / pinned-vocabulary / outcome-read). They are
	// skipped and COUNTED, never silently dropped or stored unevaluated.
	InadmissibleConditions int
	Revision               int
	InvariantCount         int
	CreatedAt              string
	Invocation             InvariantProviderInvocation
	Invariants             []SuccessInvariantRow
}

// PersistSuccessRevisionResult reports the persisted revision and newness.
type PersistSuccessRevisionResult struct {
	Record  SuccessRevisionRecord
	Created bool
}

// PersistSuccessRevision writes a compression pass transactionally: the
// provider invocation (role='success-compress'), the revision, and every
// success invariant with predicate, broken-target links, and cohort
// evaluations. Idempotent on the identity tuple; a new evaluation changes the
// cohort hash and yields the next revision — never a rewrite.
func (s *Store) PersistSuccessRevision(ctx context.Context, record SuccessRevisionRecord) (PersistSuccessRevisionResult, error) {
	if err := domain.ValidateSuccessRevisionID(record.ID); err != nil {
		return PersistSuccessRevisionResult{}, err
	}
	if err := domain.ValidateProblemID(record.ProblemID); err != nil {
		return PersistSuccessRevisionResult{}, err
	}

	var existingID string
	err := s.db.QueryRowContext(ctx, `
SELECT id FROM success_invariant_revisions
WHERE problem_id = ? AND cohort_hash = ? AND compressor_version = ? AND predicate_schema = ? AND min_support = ?
`, record.ProblemID, record.CohortHash, record.CompressorVersion, record.PredicateSchema, record.MinSupport).Scan(&existingID)
	switch {
	case err == nil:
		full, lerr := s.loadSuccessRevision(ctx, existingID)
		if lerr != nil {
			return PersistSuccessRevisionResult{}, lerr
		}
		return PersistSuccessRevisionResult{Record: full, Created: false}, nil
	case !errors.Is(err, sql.ErrNoRows):
		return PersistSuccessRevisionResult{}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return PersistSuccessRevisionResult{}, err
	}
	defer tx.Rollback()

	var maxRev sql.NullInt64
	if err := tx.QueryRowContext(ctx, `SELECT MAX(revision) FROM success_invariant_revisions WHERE problem_id = ?`, record.ProblemID).Scan(&maxRev); err != nil {
		return PersistSuccessRevisionResult{}, err
	}
	record.Revision = int(maxRev.Int64) + 1

	inv := record.Invocation
	if _, err := tx.ExecContext(ctx, `
INSERT INTO provider_invocations(id, run_id, role, provider_name, provider_version, model_name, schema_version, request_hash, request_payload, response_payload, created_at)
VALUES(?, ?, 'success-compress', ?, ?, ?, ?, ?, ?, ?, ?)
`, inv.ID, inv.RunID, inv.ProviderName, inv.ProviderVersion, inv.ModelName, inv.SchemaVersion, inv.RequestHash, inv.RequestPayload, inv.ResponsePayload, inv.CreatedAt); err != nil {
		return PersistSuccessRevisionResult{}, err
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO success_invariant_revisions(id, problem_id, run_id, provider_invocation_id, compressor_version, predicate_schema, min_support, cohort_hash, ineligible_unpersisted, ambiguous_members, inadmissible_conditions, revision, invariant_count, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, record.ID, record.ProblemID, record.RunID, inv.ID, record.CompressorVersion, record.PredicateSchema, record.MinSupport, record.CohortHash, record.IneligibleUnpersisted, record.AmbiguousMembers, record.InadmissibleConditions, record.Revision, record.InvariantCount, record.CreatedAt); err != nil {
		return PersistSuccessRevisionResult{}, err
	}

	for _, si := range record.Invariants {
		if err := domain.ValidateSuccessInvariantID(si.ID); err != nil {
			return PersistSuccessRevisionResult{}, err
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO success_invariants(id, success_revision_id, predicate_fingerprint, statement, abstraction_level, progress_coverage_num, progress_coverage_den, nonprogressor_exclusion_num, nonprogressor_exclusion_den, coverage_ordinal, exclusion_ordinal, distinct_mechanism_support, strength_deterministic, strength_reproducible, strength_independent_evidence, strength_independent_critic, strength_model_judgment, ordinal)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, si.ID, record.ID, si.PredicateFingerprint, si.Statement, si.AbstractionLevel, si.CoverageNum, si.CoverageDen, si.ExclusionNum, si.ExclusionDen, si.CoverageOrdinal, si.ExclusionOrdinal, si.DistinctSupport, si.StrengthDeterministic, si.StrengthReproducible, si.StrengthIndependentEvidence, si.StrengthIndependentCritic, si.StrengthModelJudgment, si.Ordinal); err != nil {
			return PersistSuccessRevisionResult{}, err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO success_invariant_predicates(success_invariant_id, predicate_json) VALUES(?, ?)`, si.ID, si.PredicateJSON); err != nil {
			return PersistSuccessRevisionResult{}, err
		}
		for _, target := range si.BrokenTargets {
			if _, err := tx.ExecContext(ctx, `INSERT INTO success_invariant_broken_targets(success_invariant_id, invariant_id) VALUES(?, ?)`, si.ID, target); err != nil {
				return PersistSuccessRevisionResult{}, err
			}
		}
		for _, ce := range si.CohortEvaluations {
			if _, err := tx.ExecContext(ctx, `
INSERT INTO success_invariant_cohort_evaluations(success_invariant_id, proposal_id, cohort_role, verdict, verification_strength)
VALUES(?, ?, ?, ?, ?)
`, si.ID, ce.ProposalID, ce.CohortRole, ce.Verdict, ce.Strength); err != nil {
				return PersistSuccessRevisionResult{}, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return PersistSuccessRevisionResult{}, err
	}
	return PersistSuccessRevisionResult{Record: record, Created: true}, nil
}

// GetSuccessRevision loads one full compression pass by id.
func (s *Store) GetSuccessRevision(ctx context.Context, id string) (SuccessRevisionRecord, error) {
	if err := domain.ValidateSuccessRevisionID(id); err != nil {
		return SuccessRevisionRecord{}, err
	}
	return s.loadSuccessRevision(ctx, id)
}

func (s *Store) loadSuccessRevision(ctx context.Context, id string) (SuccessRevisionRecord, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, problem_id, run_id, compressor_version, predicate_schema, min_support, cohort_hash, ineligible_unpersisted, ambiguous_members, inadmissible_conditions, revision, invariant_count, created_at
FROM success_invariant_revisions WHERE id = ?
`, id)
	var rec SuccessRevisionRecord
	if err := row.Scan(&rec.ID, &rec.ProblemID, &rec.RunID, &rec.CompressorVersion, &rec.PredicateSchema, &rec.MinSupport, &rec.CohortHash, &rec.IneligibleUnpersisted, &rec.AmbiguousMembers, &rec.InadmissibleConditions, &rec.Revision, &rec.InvariantCount, &rec.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SuccessRevisionRecord{}, fmt.Errorf("%w: success revision %s", ErrNotFound, id)
		}
		return SuccessRevisionRecord{}, err
	}
	siRows, err := s.db.QueryContext(ctx, `
SELECT si.id, si.predicate_fingerprint, COALESCE(sp.predicate_json,''), si.statement, si.abstraction_level, si.progress_coverage_num, si.progress_coverage_den, si.nonprogressor_exclusion_num, si.nonprogressor_exclusion_den, si.coverage_ordinal, si.exclusion_ordinal, si.distinct_mechanism_support, si.strength_deterministic, si.strength_reproducible, si.strength_independent_evidence, si.strength_independent_critic, si.strength_model_judgment, si.ordinal
FROM success_invariants si
LEFT JOIN success_invariant_predicates sp ON sp.success_invariant_id = si.id
WHERE si.success_revision_id = ? ORDER BY si.ordinal
`, id)
	if err != nil {
		return SuccessRevisionRecord{}, err
	}
	defer siRows.Close()
	for siRows.Next() {
		var si SuccessInvariantRow
		if err := siRows.Scan(&si.ID, &si.PredicateFingerprint, &si.PredicateJSON, &si.Statement, &si.AbstractionLevel, &si.CoverageNum, &si.CoverageDen, &si.ExclusionNum, &si.ExclusionDen, &si.CoverageOrdinal, &si.ExclusionOrdinal, &si.DistinctSupport, &si.StrengthDeterministic, &si.StrengthReproducible, &si.StrengthIndependentEvidence, &si.StrengthIndependentCritic, &si.StrengthModelJudgment, &si.Ordinal); err != nil {
			return SuccessRevisionRecord{}, err
		}
		rec.Invariants = append(rec.Invariants, si)
	}
	if err := siRows.Err(); err != nil {
		return SuccessRevisionRecord{}, err
	}
	for i := range rec.Invariants {
		tRows, err := s.db.QueryContext(ctx, `SELECT invariant_id FROM success_invariant_broken_targets WHERE success_invariant_id = ? ORDER BY invariant_id`, rec.Invariants[i].ID)
		if err != nil {
			return SuccessRevisionRecord{}, err
		}
		for tRows.Next() {
			var t string
			if err := tRows.Scan(&t); err != nil {
				tRows.Close()
				return SuccessRevisionRecord{}, err
			}
			rec.Invariants[i].BrokenTargets = append(rec.Invariants[i].BrokenTargets, t)
		}
		if err := tRows.Err(); err != nil {
			tRows.Close()
			return SuccessRevisionRecord{}, err
		}
		tRows.Close()
		ceRows, err := s.db.QueryContext(ctx, `
SELECT proposal_id, cohort_role, verdict, verification_strength FROM success_invariant_cohort_evaluations
WHERE success_invariant_id = ? ORDER BY proposal_id
`, rec.Invariants[i].ID)
		if err != nil {
			return SuccessRevisionRecord{}, err
		}
		for ceRows.Next() {
			var ce SuccessCohortEvaluationRow
			if err := ceRows.Scan(&ce.ProposalID, &ce.CohortRole, &ce.Verdict, &ce.Strength); err != nil {
				ceRows.Close()
				return SuccessRevisionRecord{}, err
			}
			rec.Invariants[i].CohortEvaluations = append(rec.Invariants[i].CohortEvaluations, ce)
		}
		if err := ceRows.Err(); err != nil {
			ceRows.Close()
			return SuccessRevisionRecord{}, err
		}
		ceRows.Close()
	}
	return rec, nil
}

// ListSuccessRevisions returns the revision headers for a problem, newest first.
func (s *Store) ListSuccessRevisions(ctx context.Context, problemID string) ([]SuccessRevisionRecord, error) {
	if err := domain.ValidateProblemID(problemID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT id, problem_id, run_id, compressor_version, predicate_schema, min_support, cohort_hash, ineligible_unpersisted, ambiguous_members, inadmissible_conditions, revision, invariant_count, created_at
FROM success_invariant_revisions WHERE problem_id = ? ORDER BY revision DESC, id DESC
`, problemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []SuccessRevisionRecord
	for rows.Next() {
		var rec SuccessRevisionRecord
		if err := rows.Scan(&rec.ID, &rec.ProblemID, &rec.RunID, &rec.CompressorVersion, &rec.PredicateSchema, &rec.MinSupport, &rec.CohortHash, &rec.IneligibleUnpersisted, &rec.AmbiguousMembers, &rec.InadmissibleConditions, &rec.Revision, &rec.InvariantCount, &rec.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

// LatestSuccessRevision returns the highest-revision compression pass id.
func (s *Store) LatestSuccessRevision(ctx context.Context, problemID string) (string, bool, error) {
	if err := domain.ValidateProblemID(problemID); err != nil {
		return "", false, err
	}
	row := s.db.QueryRowContext(ctx, `SELECT id FROM success_invariant_revisions WHERE problem_id = ? ORDER BY revision DESC, id DESC LIMIT 1`, problemID)
	var id string
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}
		return "", false, err
	}
	return id, true, nil
}
