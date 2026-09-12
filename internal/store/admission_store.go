package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/instagrim-dev/newf/internal/domain"
)

// EvidenceAdmissionRow is one immutable evidence-admission decision (v35/S2):
// an evaluated failure either admitted into the atlas population (all four
// materialization ids set, pointing at rows that hold the exact assessed
// signature content) or withheld with the refusing rule in Basis. The two
// decisions for one evaluation that can legally coexist are one `withheld`
// followed by one `admitted` (operator attestation superseding a rule
// withholding) — both rows persist, so the supersession is visible.
type EvidenceAdmissionRow struct {
	ID           string
	ProblemID    string
	RunID        string
	ProposalID   string
	EvaluationID string
	// Decision is 'admitted' or 'withheld'.
	Decision string
	// ObservationKind is the typed observation taxonomy: a
	// 'structural-claim-failure' (the description failed its own claimed
	// break), a 'domain-checked-failure' (deterministic / reproducible /
	// independent-evidence strength failure of an attempt), or a
	// 'model-judged-failure' (model-judgment or independent-critic strength).
	ObservationKind string
	// AdmittedBy is 'rule' (typed admission rule decided) or 'operator'
	// (explicit attestation decided).
	AdmittedBy string
	// Basis is the admitting rule, withholding rule, or operator attestation
	// note. Never empty.
	Basis string
	// ContentHash is the assessed signature content hash the decision was
	// made over (empty only when the evaluation recorded no assessed content,
	// which itself forces a withholding).
	ContentHash string
	// Materialization ids: set exactly when Decision == 'admitted'.
	ApproachID         string
	ApproachRevisionID string
	MechanismID        string
	SignatureID        string
	CreatedAt          string
}

// PersistEvidenceAdmission appends one admission decision. The schema enforces
// decision/materialization coherence (admitted rows carry all four ids,
// withheld rows none) and at most one row per (evaluation, decision).
func (s *Store) PersistEvidenceAdmission(ctx context.Context, row EvidenceAdmissionRow) (EvidenceAdmissionRow, error) {
	if err := domain.ValidateEvidenceAdmissionID(row.ID); err != nil {
		return EvidenceAdmissionRow{}, err
	}
	if err := domain.ValidateProblemID(row.ProblemID); err != nil {
		return EvidenceAdmissionRow{}, err
	}
	if err := domain.ValidateEvaluationID(row.EvaluationID); err != nil {
		return EvidenceAdmissionRow{}, err
	}
	if err := domain.ValidateFrontierProposalID(row.ProposalID); err != nil {
		return EvidenceAdmissionRow{}, err
	}
	switch row.Decision {
	case "admitted":
		if row.ApproachID == "" || row.ApproachRevisionID == "" || row.MechanismID == "" || row.SignatureID == "" {
			return EvidenceAdmissionRow{}, fmt.Errorf("admitted evidence requires all four materialization ids (approach, approach revision, mechanism, signature)")
		}
	case "withheld":
		if row.ApproachID != "" || row.ApproachRevisionID != "" || row.MechanismID != "" || row.SignatureID != "" {
			return EvidenceAdmissionRow{}, fmt.Errorf("withheld evidence must not carry materialization ids")
		}
	default:
		return EvidenceAdmissionRow{}, fmt.Errorf("invalid admission decision %q", row.Decision)
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO evidence_admissions(id, problem_id, run_id, proposal_id, evaluation_id, decision, observation_kind, admitted_by, basis, content_hash, approach_id, approach_revision_id, mechanism_id, signature_id, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, row.ID, row.ProblemID, row.RunID, row.ProposalID, row.EvaluationID, row.Decision, row.ObservationKind, row.AdmittedBy, row.Basis, row.ContentHash,
		nullIfEmpty(row.ApproachID), nullIfEmpty(row.ApproachRevisionID), nullIfEmpty(row.MechanismID), nullIfEmpty(row.SignatureID), row.CreatedAt)
	if err != nil {
		return EvidenceAdmissionRow{}, fmt.Errorf("persist evidence admission: %w", err)
	}
	return row, nil
}

// ListEvidenceAdmissions returns every admission decision for a problem in
// creation order (then insertion order for same-instant rows).
func (s *Store) ListEvidenceAdmissions(ctx context.Context, problemID string) ([]EvidenceAdmissionRow, error) {
	if err := domain.ValidateProblemID(problemID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT id, problem_id, run_id, proposal_id, evaluation_id, decision, observation_kind, admitted_by, basis, content_hash,
       COALESCE(approach_id,''), COALESCE(approach_revision_id,''), COALESCE(mechanism_id,''), COALESCE(signature_id,''), created_at
FROM evidence_admissions WHERE problem_id = ? ORDER BY created_at, rowid
`, problemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EvidenceAdmissionRow
	for rows.Next() {
		var r EvidenceAdmissionRow
		if err := rows.Scan(&r.ID, &r.ProblemID, &r.RunID, &r.ProposalID, &r.EvaluationID, &r.Decision, &r.ObservationKind, &r.AdmittedBy, &r.Basis, &r.ContentHash,
			&r.ApproachID, &r.ApproachRevisionID, &r.MechanismID, &r.SignatureID, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ProposalSignatureContentRow is one persisted proposal-signature revision's
// content, addressed by the exact content hash an evaluation assessed.
type ProposalSignatureContentRow struct {
	ProposalID           string
	Revision             int
	ContentHash          string
	CanonicalFingerprint string
	SignatureJSON        string
}

// GetProposalSignatureContentByHash loads the persisted signature revision of
// a proposal whose content hash matches the hash an evaluation recorded as
// assessed. This is the ONLY content evidence admission may materialize: the
// bytes the verdict was actually computed over, never an independent "latest"
// lookup.
func (s *Store) GetProposalSignatureContentByHash(ctx context.Context, proposalID, contentHash string) (ProposalSignatureContentRow, bool, error) {
	if err := domain.ValidateFrontierProposalID(proposalID); err != nil {
		return ProposalSignatureContentRow{}, false, err
	}
	if contentHash == "" {
		return ProposalSignatureContentRow{}, false, nil
	}
	row := s.db.QueryRowContext(ctx, `
SELECT proposal_id, revision, content_hash, canonical_fingerprint, signature_json
FROM frontier_proposal_signature_revisions
WHERE proposal_id = ? AND content_hash = ?
ORDER BY revision DESC LIMIT 1
`, proposalID, contentHash)
	var r ProposalSignatureContentRow
	if err := row.Scan(&r.ProposalID, &r.Revision, &r.ContentHash, &r.CanonicalFingerprint, &r.SignatureJSON); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ProposalSignatureContentRow{}, false, nil
		}
		return ProposalSignatureContentRow{}, false, err
	}
	return r, true, nil
}
