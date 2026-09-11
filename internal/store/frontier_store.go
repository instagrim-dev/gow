package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
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
	Role            string // defaults to 'generate' when empty (M7 arms set 'summarize-next'/'brainstorm')
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
	// Admission audit (v29): what the untrusted-proposal admission boundary
	// changed for THIS generation — corrected claims (provider resolution
	// disagreed with the pinned vocabulary), downgraded claims (unverifiable,
	// no surface label), proposals whose self-declared completeness was
	// stripped, and version-incompatible proposals rejected outright. All zero
	// for trusted (code-derived) generators, which bypass admission.
	AdmissionCorrected  int
	AdmissionDowngraded int
	AdmissionStripped   int
	AdmissionRejected   int
	Invocation          FrontierProviderInvocation
	Proposals           []FrontierProposalRow
}

// PersistFrontierGenerationResult reports the persisted generation and newness.
type PersistFrontierGenerationResult struct {
	Record  FrontierGenerationRecord
	Created bool
	// ProposalIDByHash maps EVERY candidate's proposal_hash to its persisted
	// proposal id — both proposals newly written by this pass AND proposals that
	// deduped onto a pre-existing row for the problem. The applied-bias log
	// (M6.2) keys on this so a policy-biased generation records the bias for the
	// whole ranked set, not only the newly-written subset (which is empty for a
	// deterministic re-generation).
	ProposalIDByHash map[string]string
	// RankedProposalIDs is the persisted proposal ids in this run's deterministic
	// rank order (populated by the pipeline after persist, not by the store).
	// M7 experiment arms consume this as the authoritative per-arm order so
	// assessment is reproducible even when an arm mixes newly-written and
	// cross-generation-deduped proposals (whose per-generation rank_ordinal
	// values are not unique across the mixed set).
	RankedProposalIDs []string
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
	role := inv.Role
	if role == "" {
		role = "generate"
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO provider_invocations(id, run_id, role, provider_name, provider_version, model_name, schema_version, request_hash, request_payload, response_payload, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, inv.ID, inv.RunID, role, inv.ProviderName, inv.ProviderVersion, inv.ModelName, inv.SchemaVersion, inv.RequestHash, inv.RequestPayload, inv.ResponsePayload, inv.CreatedAt); err != nil {
		return PersistFrontierGenerationResult{}, err
	}

	// Filter out proposals whose hash already exists for the problem (dedup
	// across runs); the persisted proposal_count reflects only newly-written rows.
	// A deduped proposal is still ENRICHED with its canonical content when the
	// sidecar row is absent (pre-v17 proposals persisted only a hash): enrichment
	// is by INSERT OR IGNORE into the sidecar keyed by the EXISTING proposal id —
	// the immutable proposal row itself is never updated.
	persisted := make([]FrontierProposalRow, 0, len(record.Proposals))
	idByHash := make(map[string]string, len(record.Proposals))
	// Occurrence bindings for DEDUPED proposals are collected here and written
	// after the generation-run row exists (FK ordering).
	type pendingOccurrence struct{ proposalID, contentHash string }
	var dedupOccurrences []pendingOccurrence
	for _, p := range record.Proposals {
		var existingID string
		err := tx.QueryRowContext(ctx, `SELECT id FROM frontier_proposals WHERE problem_id = ? AND proposal_hash = ?`, record.ProblemID, p.ProposalHash).Scan(&existingID)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			persisted = append(persisted, p)
			idByHash[p.ProposalHash] = p.ID
		case err != nil:
			return PersistFrontierGenerationResult{}, err
		default:
			// Deduped onto a pre-existing proposal row: record the existing id so
			// the applied-bias log can still attribute this run's policy bias to it.
			idByHash[p.ProposalHash] = existingID
			if p.SignatureJSON != "" {
				if _, ierr := tx.ExecContext(ctx, `
INSERT OR IGNORE INTO frontier_proposal_signatures(proposal_id, canonical_fingerprint, signature_json, created_at)
VALUES(?, ?, ?, ?)
`, existingID, p.CanonicalFingerprint, p.SignatureJSON, record.CreatedAt); ierr != nil {
					return PersistFrontierGenerationResult{}, ierr
				}
				// F1: the mechanism fingerprint (and thus the proposal hash)
				// deliberately excludes completeness/unresolved claims, so a
				// REVISED interpretation can dedup onto this proposal while
				// carrying evaluation-relevant changes. Append a content
				// revision so the revised evidence is never silently dropped,
				// and bind THIS generation to the content it emitted.
				contentHash, ierr := insertSignatureRevision(ctx, tx, existingID, p.CanonicalFingerprint, p.SignatureJSON, record.CreatedAt)
				if ierr != nil {
					return PersistFrontierGenerationResult{}, ierr
				}
				dedupOccurrences = append(dedupOccurrences, pendingOccurrence{existingID, contentHash})
			}
		}
	}
	record.Proposals = persisted
	record.ProposalCount = len(persisted)

	if _, err := tx.ExecContext(ctx, `
INSERT INTO frontier_generation_runs(id, problem_id, cluster_run_id, run_id, provider_invocation_id, generator_version, requested_count, proposal_count, revision, created_at, admission_corrected, admission_downgraded, admission_stripped, admission_rejected)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, record.ID, record.ProblemID, record.ClusterRunID, record.RunID, inv.ID, record.GeneratorVersion, record.RequestedCount, record.ProposalCount, record.Revision, record.CreatedAt, record.AdmissionCorrected, record.AdmissionDowngraded, record.AdmissionStripped, record.AdmissionRejected); err != nil {
		return PersistFrontierGenerationResult{}, err
	}
	for _, occ := range dedupOccurrences {
		if err := insertGenerationOccurrence(ctx, tx, record.ID, occ.proposalID, occ.contentHash, record.CreatedAt); err != nil {
			return PersistFrontierGenerationResult{}, err
		}
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
			contentHash, rerr := insertSignatureRevision(ctx, tx, p.ID, p.CanonicalFingerprint, p.SignatureJSON, record.CreatedAt)
			if rerr != nil {
				return PersistFrontierGenerationResult{}, rerr
			}
			if rerr := insertGenerationOccurrence(ctx, tx, record.ID, p.ID, contentHash, record.CreatedAt); rerr != nil {
				return PersistFrontierGenerationResult{}, rerr
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
	return PersistFrontierGenerationResult{Record: record, Created: true, ProposalIDByHash: idByHash}, nil
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
SELECT id, problem_id, cluster_run_id, run_id, provider_invocation_id, generator_version, requested_count, proposal_count, revision, created_at, admission_corrected, admission_downgraded, admission_stripped, admission_rejected
FROM frontier_generation_runs WHERE id = ?
`, id)
	var rec FrontierGenerationRecord
	if err := row.Scan(&rec.ID, &rec.ProblemID, &rec.ClusterRunID, &rec.RunID, &rec.ProviderInvocationID, &rec.GeneratorVersion, &rec.RequestedCount, &rec.ProposalCount, &rec.Revision, &rec.CreatedAt, &rec.AdmissionCorrected, &rec.AdmissionDowngraded, &rec.AdmissionStripped, &rec.AdmissionRejected); err != nil {
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

// FindProposalGeneration resolves the owning frontier_generation_run_id for a
// (problem, proposal). Cross-run dedup keeps a proposal under the generation
// that FIRST wrote it, so this is stable even when later generations dedup onto
// it and own zero rows. Returns found=false when the proposal does not exist for
// the problem (never fabricates a generation).
func (s *Store) FindProposalGeneration(ctx context.Context, problemID, proposalID string) (string, bool, error) {
	if err := domain.ValidateProblemID(problemID); err != nil {
		return "", false, err
	}
	row := s.db.QueryRowContext(ctx,
		`SELECT frontier_generation_run_id FROM frontier_proposals WHERE problem_id = ? AND id = ?`,
		problemID, proposalID)
	var genID string
	if err := row.Scan(&genID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}
		return "", false, err
	}
	return genID, true, nil
}

// LatestFrontierGenerationWithProposals returns the most recent generation that
// actually OWNS at least one proposal row for the problem. A fully-deduped
// latest generation owns zero rows; batch evaluate uses this so it does not go
// blind and miss existing un-evaluated artifacts (finding 2).
func (s *Store) LatestFrontierGenerationWithProposals(ctx context.Context, problemID string) (string, bool, error) {
	if err := domain.ValidateProblemID(problemID); err != nil {
		return "", false, err
	}
	row := s.db.QueryRowContext(ctx, `
SELECT g.id FROM frontier_generation_runs g
WHERE g.problem_id = ?
  AND EXISTS (SELECT 1 FROM frontier_proposals p WHERE p.frontier_generation_run_id = g.id)
ORDER BY g.revision DESC, g.id DESC LIMIT 1`, problemID)
	var id string
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}
		return "", false, err
	}
	return id, true, nil
}

// format frontier.RedundancyKey produces: "t:<target ids>,|n:<cluster ids>,")
// that appear on at least minCount DISTINCT persisted proposals for the problem.
// Search policy uses these to derive penalize/redundant_attack directives from
// accumulated evidence, so a mechanism repeatedly attacked the same way is
// down-ranked in future generations (AGENTS.md: penalize mechanisms repeatedly
// shown redundant). Reconstructed deterministically from the child tables; the
// target/cluster orderings match the insert-time ORDER BY that RedundancyKey
// also relies on.
func (s *Store) RedundantAttackKeys(ctx context.Context, problemID string, minCount int) ([]string, error) {
	if err := domain.ValidateProblemID(problemID); err != nil {
		return nil, err
	}
	if minCount < 1 {
		minCount = 1
	}
	// Build the per-proposal target segment ("t:a,b,") and cluster segment
	// ("n:x,y,") via ordered aggregation, then combine and count distinct
	// proposals per full key. group_concat preserves the ORDER BY within each
	// aggregate group on SQLite.
	rows, err := s.db.QueryContext(ctx, `
WITH t AS (
  SELECT fp.id AS pid,
         COALESCE((SELECT group_concat(ti.invariant_id || ',', '')
                   FROM (SELECT invariant_id FROM frontier_target_invariants
                         WHERE proposal_id = fp.id ORDER BY invariant_id) ti), '') AS tseg,
         COALESCE((SELECT group_concat(nc.cluster_id || ',', '')
                   FROM (SELECT cluster_id FROM frontier_nearest_clusters
                         WHERE proposal_id = fp.id ORDER BY cluster_id) nc), '') AS nseg
  FROM frontier_proposals fp
  WHERE fp.problem_id = ?
)
SELECT 't:' || tseg || '|n:' || nseg AS key, COUNT(*) AS n
FROM t GROUP BY key HAVING n >= ? ORDER BY key
`, problemID, minCount)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var keys []string
	for rows.Next() {
		var key string
		var n int
		if err := rows.Scan(&key, &n); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, rows.Err()
}

// FindGenerationForProposal resolves a proposal id to its containing
// generation run (E3: the proposal-level read surface).
func (s *Store) FindGenerationForProposal(ctx context.Context, proposalID string) (string, bool, error) {
	if err := domain.ValidateFrontierProposalID(proposalID); err != nil {
		return "", false, err
	}
	var genID string
	err := s.db.QueryRowContext(ctx, `SELECT frontier_generation_run_id FROM frontier_proposals WHERE id = ?`, proposalID).Scan(&genID)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return genID, true, nil
}

// insertSignatureRevision appends an immutable content revision for a
// proposal's signature JSON, keyed on sha256 of the persisted bytes, and
// returns that content hash. Identical content is a revision no-op (dedup);
// changed content (e.g. revised extraction completeness or unresolved claims —
// evaluation-relevant but fingerprint-invisible) gets the next revision number.
func insertSignatureRevision(ctx context.Context, tx *sql.Tx, proposalID, fingerprint, signatureJSON, createdAt string) (string, error) {
	sum := sha256.Sum256([]byte(signatureJSON))
	contentHash := hex.EncodeToString(sum[:])
	var exists int
	err := tx.QueryRowContext(ctx, `SELECT 1 FROM frontier_proposal_signature_revisions WHERE proposal_id = ? AND content_hash = ?`, proposalID, contentHash).Scan(&exists)
	if err == nil {
		return contentHash, nil // identical content already revisioned
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	_, err = tx.ExecContext(ctx, `
INSERT INTO frontier_proposal_signature_revisions(proposal_id, revision, content_hash, canonical_fingerprint, signature_json, created_at)
SELECT ?, COALESCE(MAX(revision), 0) + 1, ?, ?, ?, ?
FROM frontier_proposal_signature_revisions WHERE proposal_id = ?
`, proposalID, contentHash, fingerprint, signatureJSON, createdAt, proposalID)
	return contentHash, err
}

// insertGenerationOccurrence records which content THIS generation emitted for
// a proposal — the occurrence-level binding that keeps A -> B -> A honest:
// re-emitting earlier content references A's existing immutable revision, and
// downstream readers score what the generator actually produced, never a
// "latest" reconstruction.
func insertGenerationOccurrence(ctx context.Context, tx *sql.Tx, generationRunID, proposalID, contentHash, createdAt string) error {
	_, err := tx.ExecContext(ctx, `
INSERT OR IGNORE INTO frontier_generation_contents(generation_run_id, proposal_id, content_hash, created_at)
VALUES(?, ?, ?, ?)
`, generationRunID, proposalID, contentHash, createdAt)
	return err
}

// LatestProposalOccurrenceGeneration resolves the most recent generation that
// EMITTED the proposal (occurrence binding) — the assessment-context selector
// (review of 87759d9, finding 2). Artifact ownership answers "where does the
// row live"; occurrence membership answers "which generation last produced an
// interpretation of it", which is what re-evaluation must reach: a fully
// deduped later generation owns zero rows but binds the REVISED content.
// Falls back to the owning generation for pre-v24 history without bindings.
func (s *Store) LatestProposalOccurrenceGeneration(ctx context.Context, problemID, proposalID string) (string, bool, error) {
	if err := domain.ValidateProblemID(problemID); err != nil {
		return "", false, err
	}
	row := s.db.QueryRowContext(ctx, `
SELECT gc.generation_run_id
FROM frontier_generation_contents gc
JOIN frontier_generation_runs g ON g.id = gc.generation_run_id
WHERE g.problem_id = ? AND gc.proposal_id = ?
ORDER BY g.revision DESC, g.id DESC LIMIT 1
`, problemID, proposalID)
	var id string
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return s.FindProposalGeneration(ctx, problemID, proposalID)
		}
		return "", false, err
	}
	return id, true, nil
}

// LatestFrontierGenerationWithOccurrences resolves the most recent generation
// with >=1 occurrence binding — batch evaluation consumes occurrence
// MEMBERSHIP, not artifact ownership, so a fully-deduped latest generation
// (zero owned rows, revised bindings) is reachable (finding 2). Falls back to
// the latest generation owning proposal rows for pre-v24 history.
func (s *Store) LatestFrontierGenerationWithOccurrences(ctx context.Context, problemID string) (string, bool, error) {
	if err := domain.ValidateProblemID(problemID); err != nil {
		return "", false, err
	}
	row := s.db.QueryRowContext(ctx, `
SELECT g.id FROM frontier_generation_runs g
WHERE g.problem_id = ?
  AND EXISTS (SELECT 1 FROM frontier_generation_contents gc WHERE gc.generation_run_id = g.id)
ORDER BY g.revision DESC, g.id DESC LIMIT 1`, problemID)
	var id string
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return s.LatestFrontierGenerationWithProposals(ctx, problemID)
		}
		return "", false, err
	}
	return id, true, nil
}

// ListOccurrenceProposalRows returns the FULL proposal rows (targets, nearest
// clusters, result) for every proposal the generation EMITTED — its occurrence
// membership — regardless of which generation owns the artifact row. This is
// the evaluation membership view (finding 2): owned-row enumeration hides
// cross-generation-deduped proposals whose revised content this generation
// bound. Deterministic order by proposal id.
func (s *Store) ListOccurrenceProposalRows(ctx context.Context, generationRunID string) ([]FrontierProposalRow, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT p.id, p.proposal_hash, p.structural_violation_claim, p.novelty_argument, p.cheapest_falsification_path, p.mechanistic_distance_ordinal, p.expected_information_gain_ordinal, p.evaluation_cost_ordinal, p.violates_any_target, p.rank_ordinal, p.result
FROM frontier_generation_contents gc
JOIN frontier_proposals p ON p.id = gc.proposal_id
WHERE gc.generation_run_id = ?
ORDER BY p.id
`, generationRunID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []FrontierProposalRow
	for rows.Next() {
		var p FrontierProposalRow
		var violates int
		if err := rows.Scan(&p.ID, &p.ProposalHash, &p.StructuralViolationClaim, &p.NoveltyArgument, &p.CheapestFalsificationPath, &p.MechanisticDistance, &p.ExpectedInformationGain, &p.EvaluationCost, &violates, &p.Rank, &p.Result); err != nil {
			return nil, err
		}
		p.ViolatesAnyTarget = violates != 0
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range out {
		if err := s.loadFrontierProposalDetail(ctx, &out[i]); err != nil {
			return nil, err
		}
	}
	return out, nil
}
