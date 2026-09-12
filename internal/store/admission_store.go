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
	if err := validateEvidenceAdmissionRow(row); err != nil {
		return EvidenceAdmissionRow{}, err
	}
	if err := insertEvidenceAdmission(ctx, s.db, row); err != nil {
		return EvidenceAdmissionRow{}, err
	}
	return row, nil
}

func validateEvidenceAdmissionRow(row EvidenceAdmissionRow) error {
	if err := domain.ValidateEvidenceAdmissionID(row.ID); err != nil {
		return err
	}
	if err := domain.ValidateProblemID(row.ProblemID); err != nil {
		return err
	}
	if err := domain.ValidateEvaluationID(row.EvaluationID); err != nil {
		return err
	}
	if err := domain.ValidateFrontierProposalID(row.ProposalID); err != nil {
		return err
	}
	switch row.Decision {
	case "admitted":
		if row.ApproachID == "" || row.ApproachRevisionID == "" || row.MechanismID == "" || row.SignatureID == "" {
			return fmt.Errorf("admitted evidence requires all four materialization ids (approach, approach revision, mechanism, signature)")
		}
	case "withheld":
		if row.ApproachID != "" || row.ApproachRevisionID != "" || row.MechanismID != "" || row.SignatureID != "" {
			return fmt.Errorf("withheld evidence must not carry materialization ids")
		}
	default:
		return fmt.Errorf("invalid admission decision %q", row.Decision)
	}
	return nil
}

// admissionExecer is the write surface shared by *sql.DB and *sql.Tx, so the
// admission insert runs standalone or inside a composite transaction.
type admissionExecer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

func insertEvidenceAdmission(ctx context.Context, db admissionExecer, row EvidenceAdmissionRow) error {
	_, err := db.ExecContext(ctx, `
INSERT INTO evidence_admissions(id, problem_id, run_id, proposal_id, evaluation_id, decision, observation_kind, admitted_by, basis, content_hash, approach_id, approach_revision_id, mechanism_id, signature_id, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, row.ID, row.ProblemID, row.RunID, row.ProposalID, row.EvaluationID, row.Decision, row.ObservationKind, row.AdmittedBy, row.Basis, row.ContentHash,
		nullIfEmpty(row.ApproachID), nullIfEmpty(row.ApproachRevisionID), nullIfEmpty(row.MechanismID), nullIfEmpty(row.SignatureID), row.CreatedAt)
	if err != nil {
		return fmt.Errorf("persist evidence admission: %w", err)
	}
	return nil
}

// AdmittedFailureInput is one evaluated failure's full atlas materialization:
// the content-addressed snapshot of the assessed bytes, the normalization
// revision deriving one approach from them, the assessed signature persisted
// verbatim against the new mechanism, and the admission decision row. The
// Normalization's snapshot id, the Signature's mechanism id, and the
// Admission's four materialization ids are resolved DURING the write (snapshot
// dedup and approach-identity reuse make them unknowable beforehand), so the
// caller leaves them empty.
type AdmittedFailureInput struct {
	Snapshot      SnapshotAdmission
	Normalization NormalizationInput
	Signature     SignatureRecord
	Admission     EvidenceAdmissionRow
}

// AdmittedFailureResult reports the resolved materialization identities and
// the persisted admission row.
type AdmittedFailureResult struct {
	SnapshotID         string
	ApproachID         string
	ApproachRevisionID string
	MechanismID        string
	SignatureID        string
	Admission          EvidenceAdmissionRow
}

// PersistAdmittedFailure materializes one admitted evaluated failure into the
// atlas population in a SINGLE transaction: snapshot, normalization revision,
// signature, and admission decision row commit together or not at all. A crash
// or failure mid-sequence must never leave atlas population rows without the
// admission decision that justifies them (or vice versa).
func (s *Store) PersistAdmittedFailure(ctx context.Context, input AdmittedFailureInput) (AdmittedFailureResult, error) {
	if err := validateSnapshotAdmission(input.Snapshot); err != nil {
		return AdmittedFailureResult{}, err
	}
	if input.Admission.Decision != "admitted" {
		return AdmittedFailureResult{}, fmt.Errorf("PersistAdmittedFailure requires decision 'admitted', got %q", input.Admission.Decision)
	}
	if err := domain.ValidateMechanismSignatureID(input.Signature.ID); err != nil {
		return AdmittedFailureResult{}, err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return AdmittedFailureResult{}, err
	}
	defer tx.Rollback()

	snapshot, err := createSourceSnapshotTx(ctx, tx, input.Snapshot)
	if err != nil {
		return AdmittedFailureResult{}, fmt.Errorf("admit evidence snapshot: %w", err)
	}

	norm := input.Normalization
	norm.Revision.SnapshotID = snapshot.Snapshot.ID
	if err := validateNormalizationInput(norm); err != nil {
		return AdmittedFailureResult{}, err
	}
	normResult, err := persistNormalizationTx(ctx, tx, norm)
	if err != nil {
		return AdmittedFailureResult{}, fmt.Errorf("admit evidence normalization: %w", err)
	}
	if len(normResult.Approaches) != 1 {
		return AdmittedFailureResult{}, fmt.Errorf("admit evidence normalization wrote %d approaches, want 1", len(normResult.Approaches))
	}
	ref := normResult.Approaches[0]

	sig := input.Signature
	sig.MechanismID = ref.MechanismID
	if err := domain.ValidateMechanismID(sig.MechanismID); err != nil {
		return AdmittedFailureResult{}, err
	}
	if err := insertSignatureTx(ctx, tx, sig); err != nil {
		return AdmittedFailureResult{}, fmt.Errorf("admit evidence signature: %w", err)
	}

	row := input.Admission
	row.ApproachID = ref.ApproachID
	row.ApproachRevisionID = ref.ApproachRevisionID
	row.MechanismID = ref.MechanismID
	row.SignatureID = sig.ID
	if err := validateEvidenceAdmissionRow(row); err != nil {
		return AdmittedFailureResult{}, err
	}
	if err := insertEvidenceAdmission(ctx, tx, row); err != nil {
		return AdmittedFailureResult{}, err
	}

	if err := tx.Commit(); err != nil {
		return AdmittedFailureResult{}, err
	}
	return AdmittedFailureResult{
		SnapshotID:         snapshot.Snapshot.ID,
		ApproachID:         ref.ApproachID,
		ApproachRevisionID: ref.ApproachRevisionID,
		MechanismID:        ref.MechanismID,
		SignatureID:        sig.ID,
		Admission:          row,
	}, nil
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
