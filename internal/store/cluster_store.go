package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/instagrim-dev/newf/internal/domain"
)

// ClusterMemberRow is one persisted cluster member.
type ClusterMemberRow struct {
	SignatureID string
	MechanismID string
	Redundant   bool
	Ordinal     int
}

// ClusterRow is one persisted mechanism cluster (family).
type ClusterRow struct {
	ID                        string
	Fingerprint               string
	RepresentativeSignatureID string
	MemberCount               int
	IntraVariation            string
	Isolate                   bool
	OutcomeClass              string
	OutcomeMixed              bool
	Ordinal                   int
	Members                   []ClusterMemberRow
}

// ClusterDistanceRow is one persisted representative-vs-representative distance.
type ClusterDistanceRow struct {
	ClusterAID     string
	ClusterBID     string
	Classification string
}

// ClusterCoverageAxisRow is one persisted coverage axis.
type ClusterCoverageAxisRow struct {
	Axis               string
	DistinctValueCount int
	UnderSampled       bool
}

// ClusterDiscriminationLossRow is one persisted discrimination-loss finding.
type ClusterDiscriminationLossRow struct {
	MechanismAID string
	MechanismBID string
	OutcomeA     string
	OutcomeB     string
}

// ClusterRunRecord is the full persisted clustering pass.
type ClusterRunRecord struct {
	ID                 string
	ProblemID          string
	RunID              string
	SchemaVersion      string
	VocabularyVersion  string
	ProfileVersion     string
	ClusterAlgoVersion string
	ThresholdsHash     string
	InputSetHash       string
	SignatureCount     int
	FamilyCount        int
	Status             string
	CreatedAt          string
	Clusters           []ClusterRow
	Distances          []ClusterDistanceRow
	CoverageAxes       []ClusterCoverageAxisRow
	DiscriminationLoss []ClusterDiscriminationLossRow
}

// PersistClusterRunResult reports the persisted run and whether it was new.
type PersistClusterRunResult struct {
	Record  ClusterRunRecord
	Created bool
}

// PersistClusterRun writes a clustering pass transactionally. It is idempotent
// on the full identity tuple (problem, schema, vocabulary, profile, algo,
// thresholds, input_set_hash): an existing run is returned unchanged with
// Created=false, never rewritten (the tables are immutable by trigger). The
// input_set_hash makes re-clustering a changed signature population a new run.
func (s *Store) PersistClusterRun(ctx context.Context, record ClusterRunRecord) (PersistClusterRunResult, error) {
	if err := domain.ValidateClusterRunID(record.ID); err != nil {
		return PersistClusterRunResult{}, err
	}
	if err := domain.ValidateProblemID(record.ProblemID); err != nil {
		return PersistClusterRunResult{}, err
	}

	existingID, found, err := s.findClusterRun(ctx, record)
	if err != nil {
		return PersistClusterRunResult{}, err
	}
	if found {
		full, err := s.loadClusterRun(ctx, existingID)
		if err != nil {
			return PersistClusterRunResult{}, err
		}
		return PersistClusterRunResult{Record: full, Created: false}, nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return PersistClusterRunResult{}, err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
INSERT INTO cluster_runs(id, problem_id, run_id, schema_version, vocabulary_version, profile_version, cluster_algo_version, thresholds_hash, input_set_hash, signature_count, family_count, status, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, record.ID, record.ProblemID, record.RunID, record.SchemaVersion, record.VocabularyVersion, record.ProfileVersion, record.ClusterAlgoVersion, record.ThresholdsHash, record.InputSetHash, record.SignatureCount, record.FamilyCount, record.Status, record.CreatedAt); err != nil {
		return PersistClusterRunResult{}, err
	}

	for _, c := range record.Clusters {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO mechanism_clusters(id, cluster_run_id, cluster_fingerprint, representative_signature_id, member_count, intra_variation, isolate, outcome_class, outcome_mixed, ordinal)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, c.ID, record.ID, c.Fingerprint, c.RepresentativeSignatureID, c.MemberCount, c.IntraVariation, boolToInt(c.Isolate), c.OutcomeClass, boolToInt(c.OutcomeMixed), c.Ordinal); err != nil {
			return PersistClusterRunResult{}, err
		}
		for _, m := range c.Members {
			if _, err := tx.ExecContext(ctx, `
INSERT INTO cluster_members(cluster_id, signature_id, mechanism_id, redundant, ordinal)
VALUES(?, ?, ?, ?, ?)
`, c.ID, m.SignatureID, m.MechanismID, boolToInt(m.Redundant), m.Ordinal); err != nil {
				return PersistClusterRunResult{}, err
			}
		}
	}
	for _, d := range record.Distances {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO cluster_distances(cluster_run_id, cluster_a_id, cluster_b_id, classification)
VALUES(?, ?, ?, ?)
`, record.ID, d.ClusterAID, d.ClusterBID, d.Classification); err != nil {
			return PersistClusterRunResult{}, err
		}
	}
	for _, ax := range record.CoverageAxes {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO cluster_coverage_axes(cluster_run_id, axis, distinct_value_count, under_sampled)
VALUES(?, ?, ?, ?)
`, record.ID, ax.Axis, ax.DistinctValueCount, boolToInt(ax.UnderSampled)); err != nil {
			return PersistClusterRunResult{}, err
		}
	}
	for i, dl := range record.DiscriminationLoss {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO cluster_discrimination_losses(cluster_run_id, mechanism_a_id, mechanism_b_id, outcome_a, outcome_b, ordinal)
VALUES(?, ?, ?, ?, ?, ?)
`, record.ID, dl.MechanismAID, dl.MechanismBID, dl.OutcomeA, dl.OutcomeB, i); err != nil {
			return PersistClusterRunResult{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return PersistClusterRunResult{}, err
	}
	return PersistClusterRunResult{Record: record, Created: true}, nil
}

func (s *Store) findClusterRun(ctx context.Context, record ClusterRunRecord) (string, bool, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id FROM cluster_runs
WHERE problem_id = ? AND schema_version = ? AND vocabulary_version = ? AND profile_version = ? AND cluster_algo_version = ? AND thresholds_hash = ? AND input_set_hash = ?
`, record.ProblemID, record.SchemaVersion, record.VocabularyVersion, record.ProfileVersion, record.ClusterAlgoVersion, record.ThresholdsHash, record.InputSetHash)
	var id string
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}
		return "", false, err
	}
	return id, true, nil
}

// GetClusterRun loads a full cluster run by id.
func (s *Store) GetClusterRun(ctx context.Context, id string) (ClusterRunRecord, error) {
	if err := domain.ValidateClusterRunID(id); err != nil {
		return ClusterRunRecord{}, err
	}
	return s.loadClusterRun(ctx, id)
}

func (s *Store) loadClusterRun(ctx context.Context, id string) (ClusterRunRecord, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, problem_id, run_id, schema_version, vocabulary_version, profile_version, cluster_algo_version, thresholds_hash, input_set_hash, signature_count, family_count, status, created_at
FROM cluster_runs WHERE id = ?
`, id)
	var rec ClusterRunRecord
	if err := row.Scan(&rec.ID, &rec.ProblemID, &rec.RunID, &rec.SchemaVersion, &rec.VocabularyVersion, &rec.ProfileVersion, &rec.ClusterAlgoVersion, &rec.ThresholdsHash, &rec.InputSetHash, &rec.SignatureCount, &rec.FamilyCount, &rec.Status, &rec.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ClusterRunRecord{}, fmt.Errorf("%w: cluster run %s", ErrNotFound, id)
		}
		return ClusterRunRecord{}, err
	}

	clusterRows, err := s.db.QueryContext(ctx, `
SELECT id, cluster_fingerprint, representative_signature_id, member_count, intra_variation, isolate, outcome_class, outcome_mixed, ordinal
FROM mechanism_clusters WHERE cluster_run_id = ? ORDER BY ordinal
`, id)
	if err != nil {
		return ClusterRunRecord{}, err
	}
	defer clusterRows.Close()
	for clusterRows.Next() {
		var c ClusterRow
		var isolate, mixed int
		if err := clusterRows.Scan(&c.ID, &c.Fingerprint, &c.RepresentativeSignatureID, &c.MemberCount, &c.IntraVariation, &isolate, &c.OutcomeClass, &mixed, &c.Ordinal); err != nil {
			return ClusterRunRecord{}, err
		}
		c.Isolate = isolate != 0
		c.OutcomeMixed = mixed != 0
		rec.Clusters = append(rec.Clusters, c)
	}
	if err := clusterRows.Err(); err != nil {
		return ClusterRunRecord{}, err
	}
	for i := range rec.Clusters {
		memberRows, err := s.db.QueryContext(ctx, `
SELECT signature_id, mechanism_id, redundant, ordinal FROM cluster_members WHERE cluster_id = ? ORDER BY ordinal
`, rec.Clusters[i].ID)
		if err != nil {
			return ClusterRunRecord{}, err
		}
		for memberRows.Next() {
			var m ClusterMemberRow
			var redundant int
			if err := memberRows.Scan(&m.SignatureID, &m.MechanismID, &redundant, &m.Ordinal); err != nil {
				memberRows.Close()
				return ClusterRunRecord{}, err
			}
			m.Redundant = redundant != 0
			rec.Clusters[i].Members = append(rec.Clusters[i].Members, m)
		}
		if err := memberRows.Err(); err != nil {
			memberRows.Close()
			return ClusterRunRecord{}, err
		}
		memberRows.Close()
	}

	distRows, err := s.db.QueryContext(ctx, `
SELECT cluster_a_id, cluster_b_id, classification FROM cluster_distances WHERE cluster_run_id = ? ORDER BY cluster_a_id, cluster_b_id
`, id)
	if err != nil {
		return ClusterRunRecord{}, err
	}
	defer distRows.Close()
	for distRows.Next() {
		var d ClusterDistanceRow
		if err := distRows.Scan(&d.ClusterAID, &d.ClusterBID, &d.Classification); err != nil {
			return ClusterRunRecord{}, err
		}
		rec.Distances = append(rec.Distances, d)
	}
	if err := distRows.Err(); err != nil {
		return ClusterRunRecord{}, err
	}

	axisRows, err := s.db.QueryContext(ctx, `
SELECT axis, distinct_value_count, under_sampled FROM cluster_coverage_axes WHERE cluster_run_id = ? ORDER BY axis
`, id)
	if err != nil {
		return ClusterRunRecord{}, err
	}
	defer axisRows.Close()
	for axisRows.Next() {
		var ax ClusterCoverageAxisRow
		var under int
		if err := axisRows.Scan(&ax.Axis, &ax.DistinctValueCount, &under); err != nil {
			return ClusterRunRecord{}, err
		}
		ax.UnderSampled = under != 0
		rec.CoverageAxes = append(rec.CoverageAxes, ax)
	}
	if err := axisRows.Err(); err != nil {
		return ClusterRunRecord{}, err
	}

	lossRows, err := s.db.QueryContext(ctx, `
SELECT mechanism_a_id, mechanism_b_id, outcome_a, outcome_b FROM cluster_discrimination_losses WHERE cluster_run_id = ? ORDER BY ordinal
`, id)
	if err != nil {
		return ClusterRunRecord{}, err
	}
	defer lossRows.Close()
	for lossRows.Next() {
		var dl ClusterDiscriminationLossRow
		if err := lossRows.Scan(&dl.MechanismAID, &dl.MechanismBID, &dl.OutcomeA, &dl.OutcomeB); err != nil {
			return ClusterRunRecord{}, err
		}
		rec.DiscriminationLoss = append(rec.DiscriminationLoss, dl)
	}
	if err := lossRows.Err(); err != nil {
		return ClusterRunRecord{}, err
	}

	return rec, nil
}

// ListClusterRuns returns the cluster-run headers for a problem, newest first.
func (s *Store) ListClusterRuns(ctx context.Context, problemID string) ([]ClusterRunRecord, error) {
	if err := domain.ValidateProblemID(problemID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT id, problem_id, run_id, schema_version, vocabulary_version, profile_version, cluster_algo_version, thresholds_hash, input_set_hash, signature_count, family_count, status, created_at
FROM cluster_runs WHERE problem_id = ? ORDER BY created_at DESC, id DESC
`, problemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ClusterRunRecord
	for rows.Next() {
		var rec ClusterRunRecord
		if err := rows.Scan(&rec.ID, &rec.ProblemID, &rec.RunID, &rec.SchemaVersion, &rec.VocabularyVersion, &rec.ProfileVersion, &rec.ClusterAlgoVersion, &rec.ThresholdsHash, &rec.InputSetHash, &rec.SignatureCount, &rec.FamilyCount, &rec.Status, &rec.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

// LatestClusterRun returns the most recent cluster run id for a problem.
func (s *Store) LatestClusterRun(ctx context.Context, problemID string) (string, bool, error) {
	if err := domain.ValidateProblemID(problemID); err != nil {
		return "", false, err
	}
	row := s.db.QueryRowContext(ctx, `
SELECT id FROM cluster_runs WHERE problem_id = ? ORDER BY created_at DESC, id DESC LIMIT 1
`, problemID)
	var id string
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}
		return "", false, err
	}
	return id, true, nil
}

// LatestClusterRunForVersions returns the most recent cluster run id for a
// problem under an EXACT schema/vocabulary tuple. This is the assessment-
// population selector for challenge campaigns (v34/S1): a re-challenge may
// only widen its evidence to a population whose signatures are comparable with
// the claim's predicate — same signature schema, same vocabulary. A newer run
// under a different vocabulary is not silently substituted.
func (s *Store) LatestClusterRunForVersions(ctx context.Context, problemID, schemaVersion, vocabularyVersion string) (string, bool, error) {
	if err := domain.ValidateProblemID(problemID); err != nil {
		return "", false, err
	}
	row := s.db.QueryRowContext(ctx, `
SELECT id FROM cluster_runs
WHERE problem_id = ? AND schema_version = ? AND vocabulary_version = ?
ORDER BY created_at DESC, id DESC LIMIT 1
`, problemID, schemaVersion, vocabularyVersion)
	var id string
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}
		return "", false, err
	}
	return id, true, nil
}
