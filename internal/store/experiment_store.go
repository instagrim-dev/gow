package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/instagrim-dev/newf/internal/domain"
)

// HoldoutSetRecord binds a TRAIN problem to a QUARANTINED target problem plus
// the withheld sources. Mode is epistemically load-bearing: `blinded` makes no
// chronological claim; `historical` additionally requires a cutoff and dated
// evidence per withheld source before any execution path accepts it.
type HoldoutSetRecord struct {
	ID              string // hset_
	ProblemID       string // train problem
	TargetProblemID string // quarantined target problem
	Name            string
	Mode            string // blinded|historical
	CutoffTime      string // required when historical
	CreatedAt       string
	SourceIDs       []string // withheld sources (must belong to the target problem)
}

// PersistHoldoutSet writes a holdout set + its withheld-source links. It is
// idempotent on (problem, name): an existing set is returned unchanged.
func (s *Store) PersistHoldoutSet(ctx context.Context, rec HoldoutSetRecord) (HoldoutSetRecord, bool, error) {
	if err := domain.ValidateHoldoutSetID(rec.ID); err != nil {
		return HoldoutSetRecord{}, false, err
	}
	if err := domain.ValidateProblemID(rec.ProblemID); err != nil {
		return HoldoutSetRecord{}, false, err
	}
	if err := domain.ValidateProblemID(rec.TargetProblemID); err != nil {
		return HoldoutSetRecord{}, false, err
	}

	var existingID string
	err := s.db.QueryRowContext(ctx, `SELECT id FROM holdout_sets WHERE problem_id = ? AND name = ?`, rec.ProblemID, rec.Name).Scan(&existingID)
	switch {
	case err == nil:
		full, lerr := s.GetHoldoutSet(ctx, existingID)
		return full, false, lerr
	case !errors.Is(err, sql.ErrNoRows):
		return HoldoutSetRecord{}, false, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return HoldoutSetRecord{}, false, err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
INSERT INTO holdout_sets(id, problem_id, target_problem_id, name, mode, cutoff_time, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?)
`, rec.ID, rec.ProblemID, rec.TargetProblemID, rec.Name, rec.Mode, nullable(rec.CutoffTime), rec.CreatedAt); err != nil {
		return HoldoutSetRecord{}, false, err
	}
	for _, src := range rec.SourceIDs {
		if _, err := tx.ExecContext(ctx, `INSERT INTO holdout_set_sources(holdout_set_id, source_id) VALUES(?, ?)`, rec.ID, src); err != nil {
			return HoldoutSetRecord{}, false, err
		}
	}
	if err := tx.Commit(); err != nil {
		return HoldoutSetRecord{}, false, err
	}
	return rec, true, nil
}

// GetHoldoutSet loads a holdout set + withheld sources.
func (s *Store) GetHoldoutSet(ctx context.Context, id string) (HoldoutSetRecord, error) {
	if err := domain.ValidateHoldoutSetID(id); err != nil {
		return HoldoutSetRecord{}, err
	}
	row := s.db.QueryRowContext(ctx, `
SELECT id, problem_id, target_problem_id, name, mode, COALESCE(cutoff_time,''), created_at
FROM holdout_sets WHERE id = ?
`, id)
	var rec HoldoutSetRecord
	if err := row.Scan(&rec.ID, &rec.ProblemID, &rec.TargetProblemID, &rec.Name, &rec.Mode, &rec.CutoffTime, &rec.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return HoldoutSetRecord{}, fmt.Errorf("%w: holdout set %s", ErrNotFound, id)
		}
		return HoldoutSetRecord{}, err
	}
	rows, err := s.db.QueryContext(ctx, `SELECT source_id FROM holdout_set_sources WHERE holdout_set_id = ? ORDER BY source_id`, id)
	if err != nil {
		return HoldoutSetRecord{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var src string
		if err := rows.Scan(&src); err != nil {
			return HoldoutSetRecord{}, err
		}
		rec.SourceIDs = append(rec.SourceIDs, src)
	}
	return rec, rows.Err()
}

// CountHoldoutSourceDating reports how many withheld sources carry dated
// evidence — the historical-mode execution gate reads this.
func (s *Store) CountHoldoutSourceDating(ctx context.Context, holdoutSetID string) (dated, total int, err error) {
	if err := domain.ValidateHoldoutSetID(holdoutSetID); err != nil {
		return 0, 0, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM holdout_set_sources WHERE holdout_set_id = ?`, holdoutSetID).Scan(&total); err != nil {
		return 0, 0, err
	}
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM holdout_source_dating WHERE holdout_set_id = ?`, holdoutSetID).Scan(&dated); err != nil {
		return 0, 0, err
	}
	return dated, total, nil
}

// LeakageCheckRecord is the persisted quarantine audit. The audit is by
// CONTENT identity (sha256): a withheld snapshot's bytes appearing anywhere in
// the train problem is a leak, regardless of which source carried them.
type LeakageCheckRecord struct {
	ID                 string // lkc_
	HoldoutSetID       string
	RunID              string
	CheckerVersion     string
	SnapshotLeaks      int
	NormalizationLeaks int
	SignatureLeaks     int
	Passed             bool
	CreatedAt          string
}

// leakageCheckerVersion identifies the audit implementation.
const leakageCheckerVersion = "leakage-check/v1"

// RunLeakageCheck computes and persists the quarantine audit for a holdout
// set: no train-problem snapshot, normalization revision, or mechanism
// signature may derive from withheld CONTENT. The result row is immutable; a
// re-check is a new row.
func (s *Store) RunLeakageCheck(ctx context.Context, holdoutSetID, runID, checkID, createdAt string) (LeakageCheckRecord, error) {
	if err := domain.ValidateLeakageCheckID(checkID); err != nil {
		return LeakageCheckRecord{}, err
	}
	hs, err := s.GetHoldoutSet(ctx, holdoutSetID)
	if err != nil {
		return LeakageCheckRecord{}, err
	}

	// Withheld content identity: every sha256 of every snapshot of a withheld source.
	shaRows, err := s.db.QueryContext(ctx, `
SELECT DISTINCT ss.sha256
FROM holdout_set_sources hss
JOIN source_snapshots ss ON ss.source_id = hss.source_id
WHERE hss.holdout_set_id = ?
`, holdoutSetID)
	if err != nil {
		return LeakageCheckRecord{}, err
	}
	var shas []string
	for shaRows.Next() {
		var sha string
		if err := shaRows.Scan(&sha); err != nil {
			shaRows.Close()
			return LeakageCheckRecord{}, err
		}
		shas = append(shas, sha)
	}
	if err := shaRows.Err(); err != nil {
		shaRows.Close()
		return LeakageCheckRecord{}, err
	}
	shaRows.Close()

	rec := LeakageCheckRecord{
		ID: checkID, HoldoutSetID: holdoutSetID, RunID: runID,
		CheckerVersion: leakageCheckerVersion, CreatedAt: createdAt,
	}
	if len(shas) > 0 {
		placeholders := strings.TrimSuffix(strings.Repeat("?,", len(shas)), ",")
		args := func(extra ...any) []any {
			out := append([]any{}, extra...)
			for _, sha := range shas {
				out = append(out, sha)
			}
			return out
		}
		if err := s.db.QueryRowContext(ctx, `
SELECT COUNT(*) FROM source_snapshots ss
JOIN sources s ON s.id = ss.source_id
WHERE s.problem_id = ? AND ss.sha256 IN (`+placeholders+`)
`, args(hs.ProblemID)...).Scan(&rec.SnapshotLeaks); err != nil {
			return LeakageCheckRecord{}, err
		}
		if err := s.db.QueryRowContext(ctx, `
SELECT COUNT(*) FROM normalization_revisions nr
JOIN source_snapshots ss ON ss.id = nr.snapshot_id
WHERE nr.problem_id = ? AND ss.sha256 IN (`+placeholders+`)
`, args(hs.ProblemID)...).Scan(&rec.NormalizationLeaks); err != nil {
			return LeakageCheckRecord{}, err
		}
		if err := s.db.QueryRowContext(ctx, `
SELECT COUNT(*) FROM mechanism_signatures ms
JOIN mechanisms m ON m.id = ms.mechanism_id
JOIN approach_revisions ar ON ar.id = m.approach_revision_id
JOIN normalization_revisions nr ON nr.id = ar.normalization_revision_id
JOIN source_snapshots ss ON ss.id = nr.snapshot_id
WHERE nr.problem_id = ? AND ss.sha256 IN (`+placeholders+`)
`, args(hs.ProblemID)...).Scan(&rec.SignatureLeaks); err != nil {
			return LeakageCheckRecord{}, err
		}
	}
	rec.Passed = rec.SnapshotLeaks == 0 && rec.NormalizationLeaks == 0 && rec.SignatureLeaks == 0

	if _, err := s.db.ExecContext(ctx, `
INSERT INTO leakage_checks(id, holdout_set_id, run_id, checker_version, snapshot_leaks, normalization_leaks, signature_leaks, passed, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)
`, rec.ID, rec.HoldoutSetID, rec.RunID, rec.CheckerVersion, rec.SnapshotLeaks, rec.NormalizationLeaks, rec.SignatureLeaks, boolToInt(rec.Passed), rec.CreatedAt); err != nil {
		return LeakageCheckRecord{}, err
	}
	return rec, nil
}

// TargetSignatureRow is one FROZEN-manifest member: a canonical signature of
// the quarantined target problem that derives from a REGISTERED withheld
// source. Audit, scoring, and experiment identity all consume this same set —
// target material outside the registered withheld sources is invisible to
// scoring, exactly as it is invisible to the leakage audit.
type TargetSignatureRow struct {
	SignatureID          string
	CanonicalFingerprint string
}

// ListTargetSignaturesForHoldout returns the target signatures derived (via
// normalization provenance) from the holdout set's registered withheld
// sources, ordered by signature id.
func (s *Store) ListTargetSignaturesForHoldout(ctx context.Context, holdoutSetID string) ([]TargetSignatureRow, error) {
	if err := domain.ValidateHoldoutSetID(holdoutSetID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT DISTINCT ms.id, ms.fingerprint
FROM holdout_set_sources hss
JOIN source_snapshots ss ON ss.source_id = hss.source_id
JOIN normalization_revisions nr ON nr.snapshot_id = ss.id
JOIN approach_revisions ar ON ar.normalization_revision_id = nr.id
JOIN mechanisms m ON m.approach_revision_id = ar.id
JOIN mechanism_signatures ms ON ms.mechanism_id = m.id
WHERE hss.holdout_set_id = ?
ORDER BY ms.id
`, holdoutSetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []TargetSignatureRow
	for rows.Next() {
		var r TargetSignatureRow
		if err := rows.Scan(&r.SignatureID, &r.CanonicalFingerprint); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ExperimentArmRow is one baseline/treatment arm's persisted facts.
type ExperimentArmRow struct {
	Arm                   string // b0_undirected|b1_semantic_summary|b2_brainstorm|b3_invariant_guided
	FrontierGenerationRun string // nullable link
	ProposalCount         int
	Recovered             bool
	FirstRecoveryRank     sql.NullInt64
	NearestClassification string
	DistinctFamilyCount   int
	RedundantCount        int
	StoppingCondition     string
	// v22 measurement contract: assessment counts + actual budget consumption.
	DecisiveCount       int
	UnknownCount        int
	UnassessedCount     int
	EvaluationsConsumed int
	// Members are the EXPLICIT arm<->proposal memberships (rank + assessment):
	// what this arm derived this run, dedup mapped to persisted artifact ids.
	Members []ExperimentArmProposalRow
}

// ExperimentArmProposalRow is one explicit membership record: artifact
// deduplication and experiment participation are different identities.
type ExperimentArmProposalRow struct {
	ProposalID string
	MemberRank int
	Assessment string // recovered|decisive_no|unknown|unassessed
	// ContentHash is the exact signature content revision this assessment
	// consumed (F1): a revised interpretation is a different assessment input.
	ContentHash string
}

// ExperimentTargetRow is one frozen-manifest member persisted on the
// experiment.
type ExperimentTargetRow struct {
	SignatureID          string
	CanonicalFingerprint string
}

// ExperimentMetricRow is one exact-count metric with its derived ordinal.
type ExperimentMetricRow struct {
	Arm         string
	Metric      string
	Numerator   int
	Denominator int
	Ordinal     string
}

// ExperimentRecord is one full, immutable experiment revision.
type ExperimentRecord struct {
	ID                    string // exp_
	ProblemID             string
	HoldoutSetID          string
	LeakageCheckID        string
	RunID                 string
	Mode                  string
	RecoveryRuleVersion   string
	ProfileVersion        string
	ProposalBudgetCount   int
	EvaluationBudgetCount int
	Conclusion            string
	IdentityHash          string
	Revision              int
	CreatedAt             string
	Arms                  []ExperimentArmRow
	Metrics               []ExperimentMetricRow
	// Targets is the frozen target manifest (withheld-source-scoped signatures)
	// that scoring and identity consumed.
	Targets []ExperimentTargetRow
}

// PersistExperimentResult reports the persisted experiment and newness.
type PersistExperimentResult struct {
	Record  ExperimentRecord
	Created bool
}

// PersistExperiment writes an experiment transactionally, idempotent on
// (problem, identity_hash). A completed experiment is never rewritten.
func (s *Store) PersistExperiment(ctx context.Context, rec ExperimentRecord) (PersistExperimentResult, error) {
	if err := domain.ValidateExperimentID(rec.ID); err != nil {
		return PersistExperimentResult{}, err
	}
	if err := domain.ValidateProblemID(rec.ProblemID); err != nil {
		return PersistExperimentResult{}, err
	}
	var existingID string
	err := s.db.QueryRowContext(ctx, `SELECT id FROM experiment_runs WHERE problem_id = ? AND identity_hash = ?`, rec.ProblemID, rec.IdentityHash).Scan(&existingID)
	switch {
	case err == nil:
		full, lerr := s.GetExperiment(ctx, existingID)
		if lerr != nil {
			return PersistExperimentResult{}, lerr
		}
		return PersistExperimentResult{Record: full, Created: false}, nil
	case !errors.Is(err, sql.ErrNoRows):
		return PersistExperimentResult{}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return PersistExperimentResult{}, err
	}
	defer tx.Rollback()

	var maxRev sql.NullInt64
	if err := tx.QueryRowContext(ctx, `SELECT MAX(revision) FROM experiment_runs WHERE problem_id = ?`, rec.ProblemID).Scan(&maxRev); err != nil {
		return PersistExperimentResult{}, err
	}
	rec.Revision = int(maxRev.Int64) + 1

	if _, err := tx.ExecContext(ctx, `
INSERT INTO experiment_runs(id, problem_id, holdout_set_id, leakage_check_id, run_id, mode, recovery_rule_version, profile_version, proposal_budget_count, evaluation_budget_count, conclusion, identity_hash, revision, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, rec.ID, rec.ProblemID, rec.HoldoutSetID, rec.LeakageCheckID, rec.RunID, rec.Mode, rec.RecoveryRuleVersion, rec.ProfileVersion, rec.ProposalBudgetCount, rec.EvaluationBudgetCount, rec.Conclusion, rec.IdentityHash, rec.Revision, rec.CreatedAt); err != nil {
		return PersistExperimentResult{}, err
	}
	for _, arm := range rec.Arms {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO experiment_arms(experiment_id, arm, frontier_generation_run_id, proposal_count, recovered, first_recovery_rank, nearest_classification, distinct_family_count, redundant_count, stopping_condition, decisive_count, unknown_count, unassessed_count, evaluations_consumed)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, rec.ID, arm.Arm, nullable(arm.FrontierGenerationRun), arm.ProposalCount, boolToInt(arm.Recovered), arm.FirstRecoveryRank, arm.NearestClassification, arm.DistinctFamilyCount, arm.RedundantCount, arm.StoppingCondition, arm.DecisiveCount, arm.UnknownCount, arm.UnassessedCount, arm.EvaluationsConsumed); err != nil {
			return PersistExperimentResult{}, err
		}
		for _, m := range arm.Members {
			if _, err := tx.ExecContext(ctx, `
INSERT INTO experiment_arm_proposals(experiment_id, arm, proposal_id, member_rank, assessment, signature_content_hash)
VALUES(?, ?, ?, ?, ?, ?)
`, rec.ID, arm.Arm, m.ProposalID, m.MemberRank, m.Assessment, m.ContentHash); err != nil {
				return PersistExperimentResult{}, err
			}
		}
	}
	for _, target := range rec.Targets {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO experiment_targets(experiment_id, signature_id, canonical_fingerprint) VALUES(?, ?, ?)
`, rec.ID, target.SignatureID, target.CanonicalFingerprint); err != nil {
			return PersistExperimentResult{}, err
		}
	}
	for _, m := range rec.Metrics {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO experiment_metrics(experiment_id, arm, metric, numerator, denominator, ordinal)
VALUES(?, ?, ?, ?, ?, ?)
`, rec.ID, m.Arm, m.Metric, m.Numerator, m.Denominator, m.Ordinal); err != nil {
			return PersistExperimentResult{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return PersistExperimentResult{}, err
	}
	return PersistExperimentResult{Record: rec, Created: true}, nil
}

// GetExperiment loads one full experiment by id.
func (s *Store) GetExperiment(ctx context.Context, id string) (ExperimentRecord, error) {
	if err := domain.ValidateExperimentID(id); err != nil {
		return ExperimentRecord{}, err
	}
	row := s.db.QueryRowContext(ctx, `
SELECT id, problem_id, holdout_set_id, leakage_check_id, run_id, mode, recovery_rule_version, profile_version, proposal_budget_count, evaluation_budget_count, conclusion, identity_hash, revision, created_at
FROM experiment_runs WHERE id = ?
`, id)
	var rec ExperimentRecord
	if err := row.Scan(&rec.ID, &rec.ProblemID, &rec.HoldoutSetID, &rec.LeakageCheckID, &rec.RunID, &rec.Mode, &rec.RecoveryRuleVersion, &rec.ProfileVersion, &rec.ProposalBudgetCount, &rec.EvaluationBudgetCount, &rec.Conclusion, &rec.IdentityHash, &rec.Revision, &rec.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ExperimentRecord{}, fmt.Errorf("%w: experiment %s", ErrNotFound, id)
		}
		return ExperimentRecord{}, err
	}
	armRows, err := s.db.QueryContext(ctx, `
SELECT arm, COALESCE(frontier_generation_run_id,''), proposal_count, recovered, first_recovery_rank, nearest_classification, distinct_family_count, redundant_count, stopping_condition, decisive_count, unknown_count, unassessed_count, evaluations_consumed
FROM experiment_arms WHERE experiment_id = ? ORDER BY arm
`, id)
	if err != nil {
		return ExperimentRecord{}, err
	}
	defer armRows.Close()
	for armRows.Next() {
		var a ExperimentArmRow
		var recovered int
		if err := armRows.Scan(&a.Arm, &a.FrontierGenerationRun, &a.ProposalCount, &recovered, &a.FirstRecoveryRank, &a.NearestClassification, &a.DistinctFamilyCount, &a.RedundantCount, &a.StoppingCondition, &a.DecisiveCount, &a.UnknownCount, &a.UnassessedCount, &a.EvaluationsConsumed); err != nil {
			return ExperimentRecord{}, err
		}
		a.Recovered = recovered != 0
		rec.Arms = append(rec.Arms, a)
	}
	if err := armRows.Err(); err != nil {
		return ExperimentRecord{}, err
	}
	for i := range rec.Arms {
		memberRows, err := s.db.QueryContext(ctx, `
SELECT proposal_id, member_rank, assessment, COALESCE(signature_content_hash, '') FROM experiment_arm_proposals
WHERE experiment_id = ? AND arm = ? ORDER BY member_rank
`, id, rec.Arms[i].Arm)
		if err != nil {
			return ExperimentRecord{}, err
		}
		for memberRows.Next() {
			var m ExperimentArmProposalRow
			if err := memberRows.Scan(&m.ProposalID, &m.MemberRank, &m.Assessment, &m.ContentHash); err != nil {
				memberRows.Close()
				return ExperimentRecord{}, err
			}
			rec.Arms[i].Members = append(rec.Arms[i].Members, m)
		}
		if err := memberRows.Err(); err != nil {
			memberRows.Close()
			return ExperimentRecord{}, err
		}
		memberRows.Close()
	}
	targetRows, err := s.db.QueryContext(ctx, `SELECT signature_id, canonical_fingerprint FROM experiment_targets WHERE experiment_id = ? ORDER BY signature_id`, id)
	if err != nil {
		return ExperimentRecord{}, err
	}
	defer targetRows.Close()
	for targetRows.Next() {
		var tr ExperimentTargetRow
		if err := targetRows.Scan(&tr.SignatureID, &tr.CanonicalFingerprint); err != nil {
			return ExperimentRecord{}, err
		}
		rec.Targets = append(rec.Targets, tr)
	}
	if err := targetRows.Err(); err != nil {
		return ExperimentRecord{}, err
	}
	metricRows, err := s.db.QueryContext(ctx, `
SELECT arm, metric, numerator, denominator, ordinal FROM experiment_metrics WHERE experiment_id = ? ORDER BY arm, metric
`, id)
	if err != nil {
		return ExperimentRecord{}, err
	}
	defer metricRows.Close()
	for metricRows.Next() {
		var m ExperimentMetricRow
		if err := metricRows.Scan(&m.Arm, &m.Metric, &m.Numerator, &m.Denominator, &m.Ordinal); err != nil {
			return ExperimentRecord{}, err
		}
		rec.Metrics = append(rec.Metrics, m)
	}
	return rec, metricRows.Err()
}

// LatestHoldoutSet returns the most recent holdout set id for a problem.
func (s *Store) LatestHoldoutSet(ctx context.Context, problemID string) (string, bool, error) {
	if err := domain.ValidateProblemID(problemID); err != nil {
		return "", false, err
	}
	row := s.db.QueryRowContext(ctx, `SELECT id FROM holdout_sets WHERE problem_id = ? ORDER BY created_at DESC, id DESC LIMIT 1`, problemID)
	var id string
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}
		return "", false, err
	}
	return id, true, nil
}

// GetLeakageCheck loads one persisted quarantine audit by id.
func (s *Store) GetLeakageCheck(ctx context.Context, id string) (LeakageCheckRecord, error) {
	if err := domain.ValidateLeakageCheckID(id); err != nil {
		return LeakageCheckRecord{}, err
	}
	row := s.db.QueryRowContext(ctx, `
SELECT id, holdout_set_id, run_id, checker_version, snapshot_leaks, normalization_leaks, signature_leaks, passed, created_at
FROM leakage_checks WHERE id = ?
`, id)
	var rec LeakageCheckRecord
	var passed int
	if err := row.Scan(&rec.ID, &rec.HoldoutSetID, &rec.RunID, &rec.CheckerVersion, &rec.SnapshotLeaks, &rec.NormalizationLeaks, &rec.SignatureLeaks, &passed, &rec.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return LeakageCheckRecord{}, fmt.Errorf("%w: leakage check %s", ErrNotFound, id)
		}
		return LeakageCheckRecord{}, err
	}
	rec.Passed = passed != 0
	return rec, nil
}

// ProposalContentRow is one generated proposal's persisted canonical content
// (v17 sidecar) + rank, for recovery detection.
type ProposalContentRow struct {
	ProposalID    string
	Rank          int
	SignatureJSON string
	// ContentHash identifies the exact signature content revision read (F1):
	// evaluation-relevant fields (completeness, unresolved claims) revise
	// content without changing the mechanism fingerprint.
	ContentHash string
}

// ListProposalContents returns the proposals of one generation with their
// persisted content, rank-ordered. Proposals without sidecar content (pre-v17)
// are returned with empty JSON so callers can COUNT them as ineligible.
func (s *Store) ListProposalContents(ctx context.Context, generationRunID string) ([]ProposalContentRow, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT p.id, p.rank_ordinal, COALESCE(fps.signature_json, ''), COALESCE(fps.content_hash, '')
FROM frontier_proposals p
LEFT JOIN (
  SELECT r.proposal_id, r.signature_json, r.canonical_fingerprint, r.content_hash
  FROM frontier_proposal_signature_revisions r
  JOIN (SELECT proposal_id, MAX(revision) AS mr FROM frontier_proposal_signature_revisions GROUP BY proposal_id) lr
    ON lr.proposal_id = r.proposal_id AND lr.mr = r.revision
) fps ON fps.proposal_id = p.id
WHERE p.frontier_generation_run_id = ?
ORDER BY p.rank_ordinal
`, generationRunID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ProposalContentRow
	for rows.Next() {
		var r ProposalContentRow
		if err := rows.Scan(&r.ProposalID, &r.Rank, &r.SignatureJSON, &r.ContentHash); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ListProposalContentsForProblem returns every persisted proposal of a problem
// with its content, rank-ordered within generation insertion order. Experiment
// arms key on this problem-wide set so an unchanged atlas replays idempotently
// (a fresh generation whose proposals all dedup adds nothing to the set).
func (s *Store) ListProposalContentsForProblem(ctx context.Context, problemID string) ([]ProposalContentRow, error) {
	if err := domain.ValidateProblemID(problemID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT p.id, p.rank_ordinal, COALESCE(fps.signature_json, ''), COALESCE(fps.content_hash, '')
FROM frontier_proposals p
LEFT JOIN (
  SELECT r.proposal_id, r.signature_json, r.canonical_fingerprint, r.content_hash
  FROM frontier_proposal_signature_revisions r
  JOIN (SELECT proposal_id, MAX(revision) AS mr FROM frontier_proposal_signature_revisions GROUP BY proposal_id) lr
    ON lr.proposal_id = r.proposal_id AND lr.mr = r.revision
) fps ON fps.proposal_id = p.id
WHERE p.problem_id = ?
ORDER BY p.created_at, p.rank_ordinal, p.id
`, problemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ProposalContentRow
	for rows.Next() {
		var r ProposalContentRow
		if err := rows.Scan(&r.ProposalID, &r.Rank, &r.SignatureJSON, &r.ContentHash); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ListProposalContentsByIDs returns the persisted content + per-generation rank
// for a set of proposal ids. Ordering is NOT imposed here (the returned Rank is
// the per-generation rank_ordinal, which is not unique across a set spanning
// multiple generations) — the caller supplies the id list in its authoritative
// order and re-associates content by id. Batched to stay under SQLite's
// bound-parameter ceiling.
func (s *Store) ListProposalContentsByIDs(ctx context.Context, ids []string) ([]ProposalContentRow, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	// Batch to stay under SQLite's bound-parameter ceiling (SQLITE_MAX_VARIABLE_NUMBER,
	// historically 999). Callers impose ordering from the id list, so per-batch
	// order does not matter here.
	const batch = 900
	var out []ProposalContentRow
	for start := 0; start < len(ids); start += batch {
		end := start + batch
		if end > len(ids) {
			end = len(ids)
		}
		chunk := ids[start:end]
		placeholders := strings.TrimSuffix(strings.Repeat("?,", len(chunk)), ",")
		args := make([]any, 0, len(chunk))
		for _, id := range chunk {
			args = append(args, id)
		}
		rows, err := s.db.QueryContext(ctx, `
SELECT p.id, p.rank_ordinal, COALESCE(fps.signature_json, ''), COALESCE(fps.content_hash, '')
FROM frontier_proposals p
LEFT JOIN (
  SELECT r.proposal_id, r.signature_json, r.canonical_fingerprint, r.content_hash
  FROM frontier_proposal_signature_revisions r
  JOIN (SELECT proposal_id, MAX(revision) AS mr FROM frontier_proposal_signature_revisions GROUP BY proposal_id) lr
    ON lr.proposal_id = r.proposal_id AND lr.mr = r.revision
) fps ON fps.proposal_id = p.id
WHERE p.id IN (`+placeholders+`)
`, args...)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var r ProposalContentRow
			if err := rows.Scan(&r.ProposalID, &r.Rank, &r.SignatureJSON, &r.ContentHash); err != nil {
				rows.Close()
				return nil, err
			}
			out = append(out, r)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, err
		}
		rows.Close()
	}
	return out, nil
}

// ListExperiments returns experiment headers for a problem, newest first.
func (s *Store) ListExperiments(ctx context.Context, problemID string) ([]ExperimentRecord, error) {
	if err := domain.ValidateProblemID(problemID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT id, problem_id, holdout_set_id, leakage_check_id, run_id, mode, recovery_rule_version, profile_version, proposal_budget_count, evaluation_budget_count, conclusion, identity_hash, revision, created_at
FROM experiment_runs WHERE problem_id = ? ORDER BY revision DESC, id DESC
`, problemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ExperimentRecord
	for rows.Next() {
		var rec ExperimentRecord
		if err := rows.Scan(&rec.ID, &rec.ProblemID, &rec.HoldoutSetID, &rec.LeakageCheckID, &rec.RunID, &rec.Mode, &rec.RecoveryRuleVersion, &rec.ProfileVersion, &rec.ProposalBudgetCount, &rec.EvaluationBudgetCount, &rec.Conclusion, &rec.IdentityHash, &rec.Revision, &rec.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

// ListGenerationOccurrenceContents returns, for one generation, the content
// each proposal ACTUALLY emitted in that generation (occurrence binding), with
// the exact revision's bytes. Proposals without an occurrence row (pre-v24
// history) fall back to their LATEST revision, flagged via FallbackLatest so
// callers can report the weaker attribution honestly.
func (s *Store) ListGenerationOccurrenceContents(ctx context.Context, generationRunID string) (map[string]OccurrenceContent, error) {
	if err := domain.ValidateFrontierGenerationRunID(generationRunID); err != nil {
		return nil, err
	}
	out := make(map[string]OccurrenceContent)
	rows, err := s.db.QueryContext(ctx, `
SELECT o.proposal_id, o.content_hash, r.signature_json
FROM frontier_generation_contents o
JOIN frontier_proposal_signature_revisions r
  ON r.proposal_id = o.proposal_id AND r.content_hash = o.content_hash
WHERE o.generation_run_id = ?
`, generationRunID)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var id string
		var oc OccurrenceContent
		if err := rows.Scan(&id, &oc.ContentHash, &oc.SignatureJSON); err != nil {
			rows.Close()
			return nil, err
		}
		out[id] = oc
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	// Pre-v24 fallback: proposals of this generation without occurrence rows.
	fallback, err := s.db.QueryContext(ctx, `
SELECT p.id, COALESCE(fps.content_hash, ''), COALESCE(fps.signature_json, '')
FROM frontier_proposals p
LEFT JOIN (
  SELECT r.proposal_id, r.signature_json, r.content_hash
  FROM frontier_proposal_signature_revisions r
  JOIN (SELECT proposal_id, MAX(revision) AS mr FROM frontier_proposal_signature_revisions GROUP BY proposal_id) lr
    ON lr.proposal_id = r.proposal_id AND lr.mr = r.revision
) fps ON fps.proposal_id = p.id
WHERE p.frontier_generation_run_id = ?
`, generationRunID)
	if err != nil {
		return nil, err
	}
	defer fallback.Close()
	for fallback.Next() {
		var id string
		var oc OccurrenceContent
		if err := fallback.Scan(&id, &oc.ContentHash, &oc.SignatureJSON); err != nil {
			return nil, err
		}
		if _, bound := out[id]; !bound && oc.SignatureJSON != "" {
			oc.FallbackLatest = true
			out[id] = oc
		}
	}
	return out, fallback.Err()
}

// OccurrenceContent is one occurrence-bound signature revision.
type OccurrenceContent struct {
	ContentHash    string
	SignatureJSON  string
	FallbackLatest bool // pre-v24 history: no occurrence binding recorded
}
