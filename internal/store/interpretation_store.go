package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/instagrim-dev/newf/internal/domain"
)

// InterpretationClaimRow is one operator-adjudicated interpretation claim
// (migration v33): a ledger-provenanced GeneratedInterpretation attached to a
// specific mechanism. It carries NO claim status column by design — code fixes
// the status to `inferred` at signature-build time, so an interpretation can
// never be stored as (or drift into) an explicit source-backed claim.
type InterpretationClaimRow struct {
	ID            string
	ProblemID     string
	MechanismID   string
	FieldKind     string
	SurfaceLabel  string
	ProvenanceRef string
	Basis         string
	CreatedAt     string
}

// AddInterpretationClaim persists one immutable interpretation claim. The
// normalized-label uniqueness key makes re-adding the same claim to the same
// mechanism idempotent (the existing row is returned with Created=false).
func (s *Store) AddInterpretationClaim(ctx context.Context, row InterpretationClaimRow, normalizedLabel string) (InterpretationClaimRow, bool, error) {
	if err := domain.ValidateMechanismID(row.MechanismID); err != nil {
		return InterpretationClaimRow{}, false, err
	}
	if !domain.FieldKind(row.FieldKind).Valid() {
		return InterpretationClaimRow{}, false, fmt.Errorf("invalid field kind %q", row.FieldKind)
	}
	if strings.TrimSpace(row.SurfaceLabel) == "" {
		return InterpretationClaimRow{}, false, fmt.Errorf("interpretation claim requires a surface label")
	}
	if strings.TrimSpace(row.ProvenanceRef) == "" {
		return InterpretationClaimRow{}, false, fmt.Errorf("interpretation claim requires a provenance ref (adjudication ledger entry)")
	}
	if strings.TrimSpace(normalizedLabel) == "" {
		return InterpretationClaimRow{}, false, fmt.Errorf("interpretation claim requires a normalized label")
	}

	res, err := s.db.ExecContext(ctx, `
INSERT OR IGNORE INTO interpretation_claims
  (id, problem_id, mechanism_id, field_kind, surface_label, label_normalized, provenance_ref, basis, created_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
`, row.ID, row.ProblemID, row.MechanismID, row.FieldKind, row.SurfaceLabel, normalizedLabel, row.ProvenanceRef, row.Basis, row.CreatedAt)
	if err != nil {
		return InterpretationClaimRow{}, false, err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return InterpretationClaimRow{}, false, err
	}
	existing := s.db.QueryRowContext(ctx, `
SELECT id, problem_id, mechanism_id, field_kind, surface_label, provenance_ref, basis, created_at
FROM interpretation_claims
WHERE mechanism_id = ? AND field_kind = ? AND label_normalized = ?
`, row.MechanismID, row.FieldKind, normalizedLabel)
	var out InterpretationClaimRow
	if err := existing.Scan(&out.ID, &out.ProblemID, &out.MechanismID, &out.FieldKind, &out.SurfaceLabel, &out.ProvenanceRef, &out.Basis, &out.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return InterpretationClaimRow{}, false, fmt.Errorf("interpretation claim vanished after insert")
		}
		return InterpretationClaimRow{}, false, err
	}
	return out, affected > 0, nil
}

// ListInterpretationClaims returns the interpretation claims attached to one
// mechanism, in insertion order.
func (s *Store) ListInterpretationClaims(ctx context.Context, mechanismID string) ([]InterpretationClaimRow, error) {
	if err := domain.ValidateMechanismID(mechanismID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT id, problem_id, mechanism_id, field_kind, surface_label, provenance_ref, basis, created_at
FROM interpretation_claims
WHERE mechanism_id = ?
ORDER BY created_at, id
`, mechanismID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []InterpretationClaimRow
	for rows.Next() {
		var r InterpretationClaimRow
		if err := rows.Scan(&r.ID, &r.ProblemID, &r.MechanismID, &r.FieldKind, &r.SurfaceLabel, &r.ProvenanceRef, &r.Basis, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
