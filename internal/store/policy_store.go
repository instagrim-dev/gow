package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/instagrim-dev/newf/internal/domain"
)

// PolicyDirectiveRow is one persisted, resolved policy directive.
type PolicyDirectiveRow struct {
	ID              string
	Kind            string
	TargetKind      string
	TargetID        string
	Weight          string
	EpistemicSource string
	Ordinal         int
	// Provenance links this directive to the justifying evidence rows.
	Provenance []PolicyProvenanceRow
}

// PolicyProvenanceRow is one justifying-evidence reference for a directive.
type PolicyProvenanceRow struct {
	EvidenceKind string
	EvidenceRef  string
}

// PolicyRevisionRecord is one full mutation pass. Idempotent on
// (problem, evidence_cohort_hash, mutator_version, policy_schema).
type PolicyRevisionRecord struct {
	ID                 string // spr_
	ProblemID          string
	RunID              string
	MutatorVersion     string
	PolicySchema       string
	EvidenceCohortHash string
	InertProposals     int
	Revision           int
	DirectiveCount     int
	CreatedAt          string
	Invocation         *InvariantProviderInvocation // nil when purely code-derived (no provider fold-in)
	Directives         []PolicyDirectiveRow
}

// PersistPolicyRevision writes a mutation pass transactionally: the optional
// provider invocation (role='policy-mutate'), the revision, and every directive
// with its provenance. Idempotent on the identity tuple; new evidence changes
// the cohort hash and yields the next revision — never a rewrite.
func (s *Store) PersistPolicyRevision(ctx context.Context, record PolicyRevisionRecord) (PolicyRevisionRecord, bool, error) {
	if err := domain.ValidateSearchPolicyRevisionID(record.ID); err != nil {
		return PolicyRevisionRecord{}, false, err
	}
	if err := domain.ValidateProblemID(record.ProblemID); err != nil {
		return PolicyRevisionRecord{}, false, err
	}

	var existingID string
	err := s.db.QueryRowContext(ctx, `
SELECT id FROM search_policy_revisions
WHERE problem_id = ? AND evidence_cohort_hash = ? AND mutator_version = ? AND policy_schema = ?
`, record.ProblemID, record.EvidenceCohortHash, record.MutatorVersion, record.PolicySchema).Scan(&existingID)
	switch {
	case err == nil:
		full, lerr := s.loadPolicyRevision(ctx, existingID)
		if lerr != nil {
			return PolicyRevisionRecord{}, false, lerr
		}
		// v30: dedup preserves the artifact; THIS execution still selected it.
		if _, serr := s.db.ExecContext(ctx, `
INSERT OR IGNORE INTO policy_mutation_selections(run_id, problem_id, policy_revision_id, created_at)
VALUES(?, ?, ?, ?)
`, record.RunID, record.ProblemID, existingID, record.CreatedAt); serr != nil {
			return PolicyRevisionRecord{}, false, serr
		}
		return full, false, nil
	case !errors.Is(err, sql.ErrNoRows):
		return PolicyRevisionRecord{}, false, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return PolicyRevisionRecord{}, false, err
	}
	defer tx.Rollback()

	var maxRev sql.NullInt64
	if err := tx.QueryRowContext(ctx, `SELECT MAX(revision) FROM search_policy_revisions WHERE problem_id = ?`, record.ProblemID).Scan(&maxRev); err != nil {
		return PolicyRevisionRecord{}, false, err
	}
	record.Revision = int(maxRev.Int64) + 1

	var invID sql.NullString
	if record.Invocation != nil {
		inv := record.Invocation
		if _, err := tx.ExecContext(ctx, `
INSERT INTO provider_invocations(id, run_id, role, provider_name, provider_version, model_name, schema_version, request_hash, request_payload, response_payload, created_at)
VALUES(?, ?, 'policy-mutate', ?, ?, ?, ?, ?, ?, ?, ?)
`, inv.ID, inv.RunID, inv.ProviderName, inv.ProviderVersion, inv.ModelName, inv.SchemaVersion, inv.RequestHash, inv.RequestPayload, inv.ResponsePayload, inv.CreatedAt); err != nil {
			return PolicyRevisionRecord{}, false, err
		}
		invID = sql.NullString{String: inv.ID, Valid: true}
	}

	if _, err := tx.ExecContext(ctx, `
INSERT INTO search_policy_revisions(id, problem_id, run_id, provider_invocation_id, mutator_version, policy_schema, evidence_cohort_hash, inert_proposals, revision, directive_count, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, record.ID, record.ProblemID, record.RunID, invID, record.MutatorVersion, record.PolicySchema, record.EvidenceCohortHash, record.InertProposals, record.Revision, record.DirectiveCount, record.CreatedAt); err != nil {
		return PolicyRevisionRecord{}, false, err
	}
	// v30: the creating execution selects its own artifact.
	if _, err := tx.ExecContext(ctx, `
INSERT OR IGNORE INTO policy_mutation_selections(run_id, problem_id, policy_revision_id, created_at)
VALUES(?, ?, ?, ?)
`, record.RunID, record.ProblemID, record.ID, record.CreatedAt); err != nil {
		return PolicyRevisionRecord{}, false, err
	}

	for _, d := range record.Directives {
		if err := domain.ValidateSearchPolicyDirectiveID(d.ID); err != nil {
			return PolicyRevisionRecord{}, false, err
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO search_policy_directives(id, policy_revision_id, kind, target_kind, target_id, weight, epistemic_source, ordinal)
VALUES(?, ?, ?, ?, ?, ?, ?, ?)
`, d.ID, record.ID, d.Kind, d.TargetKind, d.TargetID, d.Weight, d.EpistemicSource, d.Ordinal); err != nil {
			return PolicyRevisionRecord{}, false, err
		}
		for _, pv := range d.Provenance {
			if _, err := tx.ExecContext(ctx, `
INSERT INTO search_policy_provenance(policy_directive_id, evidence_kind, evidence_ref) VALUES(?, ?, ?)
`, d.ID, pv.EvidenceKind, pv.EvidenceRef); err != nil {
				return PolicyRevisionRecord{}, false, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return PolicyRevisionRecord{}, false, err
	}
	return record, true, nil
}

// GetPolicyRevision loads one full mutation pass by id.
func (s *Store) GetPolicyRevision(ctx context.Context, id string) (PolicyRevisionRecord, error) {
	if err := domain.ValidateSearchPolicyRevisionID(id); err != nil {
		return PolicyRevisionRecord{}, err
	}
	return s.loadPolicyRevision(ctx, id)
}

func (s *Store) loadPolicyRevision(ctx context.Context, id string) (PolicyRevisionRecord, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, problem_id, run_id, mutator_version, policy_schema, evidence_cohort_hash, inert_proposals, revision, directive_count, created_at
FROM search_policy_revisions WHERE id = ?
`, id)
	var rec PolicyRevisionRecord
	if err := row.Scan(&rec.ID, &rec.ProblemID, &rec.RunID, &rec.MutatorVersion, &rec.PolicySchema, &rec.EvidenceCohortHash, &rec.InertProposals, &rec.Revision, &rec.DirectiveCount, &rec.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return PolicyRevisionRecord{}, fmt.Errorf("%w: search policy revision %s", ErrNotFound, id)
		}
		return PolicyRevisionRecord{}, err
	}
	dRows, err := s.db.QueryContext(ctx, `
SELECT id, kind, target_kind, target_id, weight, epistemic_source, ordinal
FROM search_policy_directives WHERE policy_revision_id = ? ORDER BY ordinal
`, id)
	if err != nil {
		return PolicyRevisionRecord{}, err
	}
	defer dRows.Close()
	for dRows.Next() {
		var d PolicyDirectiveRow
		if err := dRows.Scan(&d.ID, &d.Kind, &d.TargetKind, &d.TargetID, &d.Weight, &d.EpistemicSource, &d.Ordinal); err != nil {
			return PolicyRevisionRecord{}, err
		}
		rec.Directives = append(rec.Directives, d)
	}
	if err := dRows.Err(); err != nil {
		return PolicyRevisionRecord{}, err
	}
	for i := range rec.Directives {
		pRows, err := s.db.QueryContext(ctx, `
SELECT evidence_kind, evidence_ref FROM search_policy_provenance
WHERE policy_directive_id = ? ORDER BY evidence_kind, evidence_ref
`, rec.Directives[i].ID)
		if err != nil {
			return PolicyRevisionRecord{}, err
		}
		for pRows.Next() {
			var pv PolicyProvenanceRow
			if err := pRows.Scan(&pv.EvidenceKind, &pv.EvidenceRef); err != nil {
				pRows.Close()
				return PolicyRevisionRecord{}, err
			}
			rec.Directives[i].Provenance = append(rec.Directives[i].Provenance, pv)
		}
		if err := pRows.Err(); err != nil {
			pRows.Close()
			return PolicyRevisionRecord{}, err
		}
		pRows.Close()
	}
	return rec, nil
}

// ListPolicyRevisions returns the revision headers for a problem, newest first.
func (s *Store) ListPolicyRevisions(ctx context.Context, problemID string) ([]PolicyRevisionRecord, error) {
	if err := domain.ValidateProblemID(problemID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT id, problem_id, run_id, mutator_version, policy_schema, evidence_cohort_hash, inert_proposals, revision, directive_count, created_at
FROM search_policy_revisions WHERE problem_id = ? ORDER BY revision DESC, id DESC
`, problemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []PolicyRevisionRecord
	for rows.Next() {
		var rec PolicyRevisionRecord
		if err := rows.Scan(&rec.ID, &rec.ProblemID, &rec.RunID, &rec.MutatorVersion, &rec.PolicySchema, &rec.EvidenceCohortHash, &rec.InertProposals, &rec.Revision, &rec.DirectiveCount, &rec.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	return out, rows.Err()
}

// LatestPolicyRevision returns the highest-revision mutation pass id.
func (s *Store) LatestPolicyRevision(ctx context.Context, problemID string) (string, bool, error) {
	if err := domain.ValidateProblemID(problemID); err != nil {
		return "", false, err
	}
	row := s.db.QueryRowContext(ctx, `SELECT id FROM search_policy_revisions WHERE problem_id = ? ORDER BY revision DESC, id DESC LIMIT 1`, problemID)
	var id string
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}
		return "", false, err
	}
	return id, true, nil
}

// LatestSelectedPolicyRevision resolves the artifact chosen by the MOST RECENT
// mutation execution (v30) — the current-policy selector. Dedup means the
// largest revision number is not current state: evidence that reverts to an
// earlier cohort reuses the earlier revision, and generation must apply THAT,
// not the higher-numbered stale one. Falls back to the highest revision for
// pre-v30 history without selection rows.
func (s *Store) LatestSelectedPolicyRevision(ctx context.Context, problemID string) (string, bool, error) {
	if err := domain.ValidateProblemID(problemID); err != nil {
		return "", false, err
	}
	row := s.db.QueryRowContext(ctx, `
SELECT policy_revision_id FROM policy_mutation_selections
WHERE problem_id = ? ORDER BY created_at DESC, rowid DESC LIMIT 1
`, problemID)
	var id string
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return s.LatestPolicyRevision(ctx, problemID)
		}
		return "", false, err
	}
	return id, true, nil
}

// FrontierGenerationPolicyRow is one persisted applied-bias entry linking a
// frontier generation, a proposal, and the policy revision that biased it.
type FrontierGenerationPolicyRow struct {
	ProposalID     string
	NetBias        int
	Preferred      bool
	Avoided        bool
	Penalized      bool
	FloorProtected bool
}

// PersistFrontierGenerationPolicy records the applied-bias log for a generation
// that was biased by a policy revision. It is a separate immutable-insert pass
// keyed on the already-persisted generation + proposal ids; a generation with
// no policy writes nothing (unbiased == absence). Idempotent: re-writing the
// same (generation, proposal) is ignored.
func (s *Store) PersistFrontierGenerationPolicy(ctx context.Context, generationRunID, policyRevisionID string, rows []FrontierGenerationPolicyRow) error {
	if err := domain.ValidateFrontierGenerationRunID(generationRunID); err != nil {
		return err
	}
	if err := domain.ValidateSearchPolicyRevisionID(policyRevisionID); err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, r := range rows {
		if _, err := tx.ExecContext(ctx, `
INSERT OR IGNORE INTO frontier_generation_policy(frontier_generation_run_id, proposal_id, policy_revision_id, net_bias, preferred, avoided, penalized, floor_protected)
VALUES(?, ?, ?, ?, ?, ?, ?, ?)
`, generationRunID, r.ProposalID, policyRevisionID, r.NetBias, boolToInt(r.Preferred), boolToInt(r.Avoided), boolToInt(r.Penalized), boolToInt(r.FloorProtected)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// GetFrontierGenerationPolicy returns the applied-bias log for a generation
// (empty when no policy was applied), plus the policy revision id that biased
// it (empty when unbiased).
func (s *Store) GetFrontierGenerationPolicy(ctx context.Context, generationRunID string) (string, []FrontierGenerationPolicyRow, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT policy_revision_id, proposal_id, net_bias, preferred, avoided, penalized, floor_protected
FROM frontier_generation_policy WHERE frontier_generation_run_id = ? ORDER BY proposal_id
`, generationRunID)
	if err != nil {
		return "", nil, err
	}
	defer rows.Close()
	var revID string
	var out []FrontierGenerationPolicyRow
	for rows.Next() {
		var r FrontierGenerationPolicyRow
		if err := rows.Scan(&revID, &r.ProposalID, &r.NetBias, &r.Preferred, &r.Avoided, &r.Penalized, &r.FloorProtected); err != nil {
			return "", nil, err
		}
		out = append(out, r)
	}
	return revID, out, rows.Err()
}
