package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/instagrim-dev/newf/internal/domain"
)

// FrontierProviderInvocation is the generation provider invocation recorded in
// the shared provider_invocations table with role='generate' (permitted after
// the v14 role generalization).
type FrontierProviderInvocation struct {
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

// FrontierTargetRow is one persisted target-invariant link + the code-verified
// per-target violation verdict.
type FrontierTargetRow struct {
	InvariantID string
	Verdict     string // 'satisfies'|'violates'|'unknown'
	Violated    bool
}

// FrontierNearestRow is one persisted nearest-family link with the code-computed
// classification + proximity ordinal.
type FrontierNearestRow struct {
	ClusterID      string
	Classification string
	Proximity      string
}

// FrontierProposalRow is one persisted, ranked frontier proposal. The `result`
// column is intentionally left NULL by this slice (M5.2 evaluation populates it).
// CanonicalFingerprint/SignatureJSON are the proposed mechanism's canonical
// content, persisted in the v17 sidecar (success compression must evaluate
// condition predicates against successful proposals; a hash is not evaluable).
type FrontierProposalRow struct {
	ID                        string
	ProposalHash              string
	CanonicalFingerprint      string
	SignatureJSON             string
	StructuralViolationClaim  string
	NoveltyArgument           string
	CheapestFalsificationPath string
	MechanisticDistance       string
	ExpectedInformationGain   string
	EvaluationCost            string
	ViolatesAnyTarget         bool
	Rank                      int
	Result                    sql.NullString
	Targets                   []FrontierTargetRow
	NearestClusters           []FrontierNearestRow
}

// FrontierGenerationRecord is the full persisted generation pass.
type FrontierGenerationRecord struct {
	ID                   string
	ProblemID            string
	ClusterRunID         string
	RunID                string
	ProviderInvocationID string
	GeneratorVersion     string
	RequestedCount       int
	ProposalCount        int
	Revision             int
	CreatedAt            string
	Invocation           FrontierProviderInvocation
	Proposals            []FrontierProposalRow
}

// PersistFrontierGenerationResult reports the persisted generation and newness.
type PersistFrontierGenerationResult struct {
	Record  FrontierGenerationRecord
	Created bool
}

// PersistFrontierGeneration writes a generation pass transactionally, assigning
// the next monotonic revision for the problem and recording the provider
// invocation with role='generate'. Proposals dedup on
// UNIQUE(problem_id, proposal_hash): a proposal whose hash already exists for the
// problem (from an earlier run) is skipped, so re-generation never rewrites a
// prior proposal (rows are immutable by trigger). A run that produces only
// already-seen proposals still persists as a new (empty) generation revision so
// the run lifecycle and provenance are recorded.
func (s *Store) PersistFrontierGeneration(ctx context.Context, record FrontierGenerationRecord) (PersistFrontierGenerationResult, error) {
	if err := domain.ValidateFrontierGenerationRunID(record.ID); err != nil {
		return PersistFrontierGenerationResult{}, err
	}
	if err := domain.ValidateProblemID(record.ProblemID); err != nil {
		return PersistFrontierGenerationResult{}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return PersistFrontierGenerationResult{}, err
	}
	defer tx.Rollback()

	var maxRev sql.NullInt64
	if err := tx.QueryRowContext(ctx, `SELECT MAX(revision) FROM frontier_generation_runs WHERE problem_id = ?`, record.ProblemID).Scan(&maxRev); err != nil {
		return PersistFrontierGenerationResult{}, err
	}
	record.Revision = int(maxRev.Int64) + 1

	inv := record.Invocation
	if _, err := tx.ExecContext(ctx, `
INSERT INTO provider_invocations(id, run_id, role, provider_name, provider_version, model_name, schema_version, request_hash, request_payload, response_payload, created_at)
VALUES(?, ?, 'generate', ?, ?, ?, ?, ?, ?, ?, ?)
`, inv.ID, inv.RunID, inv.ProviderName, inv.ProviderVersion, inv.ModelName, inv.SchemaVersion, inv.RequestHash, inv.RequestPayload, inv.ResponsePayload, inv.CreatedAt); err != nil {
		return PersistFrontierGenerationResult{}, err
	}

	// Filter out proposals whose hash already exists for the problem (dedup
	// across runs); the persisted proposal_count reflects only newly-written rows.
	// A deduped proposal is still ENRICHED with its canonical content when the
	// sidecar row is absent (pre-v17 proposals persisted only a hash): enrichment
	// is by INSERT OR IGNORE into the sidecar keyed by the EXISTING proposal id —
	// the immutable proposal row itself is never updated.
	persisted := make([]FrontierProposalRow, 0, len(record.Proposals))
	for _, p := range record.Proposals {
		var existingID string
		err := tx.QueryRowContext(ctx, `SELECT id FROM frontier_proposals WHERE problem_id = ? AND proposal_hash = ?`, record.ProblemID, p.ProposalHash).Scan(&existingID)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			persisted = append(persisted, p)
		case err != nil:
			return PersistFrontierGenerationResult{}, err
		default:
			if p.SignatureJSON != "" {
				if _, ierr := tx.ExecContext(ctx, `
INSERT OR IGNORE INTO frontier_proposal_signatures(proposal_id, canonical_fingerprint, signature_json, created_at)
VALUES(?, ?, ?, ?)
`, existingID, p.CanonicalFingerprint, p.SignatureJSON, record.CreatedAt); ierr != nil {
					return PersistFrontierGenerationResult{}, ierr
				}
			}
		}
	}
	record.Proposals = persisted
	record.ProposalCount = len(persisted)

	if _, err := tx.ExecContext(ctx, `
INSERT INTO frontier_generation_runs(id, problem_id, cluster_run_id, run_id, provider_invocation_id, generator_version, requested_count, proposal_count, revision, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, record.ID, record.ProblemID, record.ClusterRunID, record.RunID, inv.ID, record.GeneratorVersion, record.RequestedCount, record.ProposalCount, record.Revision, record.CreatedAt); err != nil {
		return PersistFrontierGenerationResult{}, err
	}

	for _, p := range record.Proposals {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO frontier_proposals(id, problem_id, frontier_generation_run_id, proposal_hash, structural_violation_claim, novelty_argument, cheapest_falsification_path, mechanistic_distance_ordinal, expected_information_gain_ordinal, evaluation_cost_ordinal, violates_any_target, rank_ordinal, result, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULL, ?)
`, p.ID, record.ProblemID, record.ID, p.ProposalHash, p.StructuralViolationClaim, p.NoveltyArgument, p.CheapestFalsificationPath, p.MechanisticDistance, p.ExpectedInformationGain, p.EvaluationCost, boolToInt(p.ViolatesAnyTarget), p.Rank, record.CreatedAt); err != nil {
			return PersistFrontierGenerationResult{}, err
		}
		if p.SignatureJSON != "" {
			if _, err := tx.ExecContext(ctx, `
INSERT INTO frontier_proposal_signatures(proposal_id, canonical_fingerprint, signature_json, created_at)
VALUES(?, ?, ?, ?)
`, p.ID, p.CanonicalFingerprint, p.SignatureJSON, record.CreatedAt); err != nil {
				return PersistFrontierGenerationResult{}, err
			}
		}
		for _, t := range p.Targets {
			if _, err := tx.ExecContext(ctx, `
INSERT INTO frontier_target_invariants(proposal_id, invariant_id, verdict, violated) VALUES(?, ?, ?, ?)
`, p.ID, t.InvariantID, t.Verdict, boolToInt(t.Violated)); err != nil {
				return PersistFrontierGenerationResult{}, err
			}
		}
		for _, n := range p.NearestClusters {
			if _, err := tx.ExecContext(ctx, `
INSERT INTO frontier_nearest_clusters(proposal_id, cluster_id, classification, proximity_ordinal) VALUES(?, ?, ?, ?)
`, p.ID, n.ClusterID, n.Classification, n.Proximity); err != nil {
				return PersistFrontierGenerationResult{}, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return PersistFrontierGenerationResult{}, err
	}
	return PersistFrontierGenerationResult{Record: record, Created: true}, nil
}

// GetFrontierGeneration loads a full generation pass by id.
func (s *Store) GetFrontierGeneration(ctx context.Context, id string) (FrontierGenerationRecord, error) {
	if err := domain.ValidateFrontierGenerationRunID(id); err != nil {
		return FrontierGenerationRecord{}, err
	}
	return s.loadFrontierGeneration(ctx, id)
}

func (s *Store) loadFrontierGeneration(ctx context.Context, id string) (FrontierGenerationRecord, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, problem_id, cluster_run_id, run_id, provider_invocation_id, generator_version, requested_count, proposal_count, revision, created_at
FROM frontier_generation_runs WHERE id = ?
`, id)
	var rec FrontierGenerationRecord
	if err := row.Scan(&rec.ID, &rec.ProblemID, &rec.ClusterRunID, &rec.RunID, &rec.ProviderInvocationID, &rec.GeneratorVersion, &rec.RequestedCount, &rec.ProposalCount, &rec.Revision, &rec.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return FrontierGenerationRecord{}, fmt.Errorf("%w: frontier generation %s", ErrNotFound, id)
		}
		return FrontierGenerationRecord{}, err
	}
	propRows, err := s.db.QueryContext(ctx, `
SELECT id, proposal_hash, structural_violation_claim, novelty_argument, cheapest_falsification_path, mechanistic_distance_ordinal, expected_information_gain_ordinal, evaluation_cost_ordinal, violates_any_target, rank_ordinal, result
FROM frontier_proposals WHERE frontier_generation_run_id = ? ORDER BY rank_ordinal
`, id)
	if err != nil {
		return FrontierGenerationRecord{}, err
	}
	defer propRows.Close()
	for propRows.Next() {
		var p FrontierProposalRow
		var violates int
		if err := propRows.Scan(&p.ID, &p.ProposalHash, &p.StructuralViolationClaim, &p.NoveltyArgument, &p.CheapestFalsificationPath, &p.MechanisticDistance, &p.ExpectedInformationGain, &p.EvaluationCost, &violates, &p.Rank, &p.Result); err != nil {
			return FrontierGenerationRecord{}, err
		}
		p.ViolatesAnyTarget = violates != 0
		rec.Proposals = append(rec.Proposals, p)
	}
	if err := propRows.Err(); err != nil {
		return FrontierGenerationRecord{}, err
	}
	for i := range rec.Proposals {
		if err := s.loadFrontierProposalDetail(ctx, &rec.Proposals[i]); err != nil {
			return FrontierGenerationRecord{}, err
		}
	}
	return rec, nil
}

func (s *Store) loadFrontierProposalDetail(ctx context.Context, p *FrontierProposalRow) error {
	tRows, err := s.db.QueryContext(ctx, `SELECT invariant_id, verdict, violated FROM frontier_target_invariants WHERE proposal_id = ? ORDER BY invariant_id`, p.ID)
	if err != nil {
		return err
	}
	defer tRows.Close()
	for tRows.Next() {
		var t FrontierTargetRow
		var violated int
		if err := tRows.Scan(&t.InvariantID, &t.Verdict, &violated); err != nil {
			return err
		}
		t.Violated = violated != 0
		p.Targets = append(p.Targets, t)
	}
	if err := tRows.Err(); err != nil {
		return err
	}
	nRows, err := s.db.QueryContext(ctx, `SELECT cluster_id, classification, proximity_ordinal FROM frontier_nearest_clusters WHERE proposal_id = ? ORDER BY cluster_id`, p.ID)
	if err != nil {
		return err
	}
	defer nRows.Close()
	for nRows.Next() {
		var n FrontierNearestRow
		if err := nRows.Scan(&n.ClusterID, &n.Classification, &n.Proximity); err != nil {
			return err
		}
		p.NearestClusters = append(p.NearestClusters, n)
	}
	return nRows.Err()
}

// ListFrontierGenerations returns the generation headers for a problem, newest
// first (highest revision).
func (s *Store) ListFrontierGenerations(ctx context.Context, problemID string) ([]FrontierGenerationRecord, error) {
	if err := domain.ValidateProblemID(problemID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT id, problem_id, cluster_run_id, run_id, provider_invocation_id, generator_version, requested_count, proposal_count, revision, created_at
FROM frontier_generation_runs WHERE problem_id = ? ORDER BY revision DESC, id DESC
`, problemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []FrontierGenerationRecord
	for rows.Next() {
		var rec FrontierGenerationRecord
		if err := rows.Scan(&rec.ID, &rec.ProblemID, &rec.ClusterRunID, &rec.RunID, &rec.ProviderInvocationID, &rec.GeneratorVersion, &rec.RequestedCount, &rec.ProposalCount, &rec.Revision, &rec.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

// LatestFrontierGeneration returns the highest-revision generation id for a
// problem.
func (s *Store) LatestFrontierGeneration(ctx context.Context, problemID string) (string, bool, error) {
	if err := domain.ValidateProblemID(problemID); err != nil {
		return "", false, err
	}
	row := s.db.QueryRowContext(ctx, `SELECT id FROM frontier_generation_runs WHERE problem_id = ? ORDER BY revision DESC, id DESC LIMIT 1`, problemID)
	var id string
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}
		return "", false, err
	}
	return id, true, nil
}
