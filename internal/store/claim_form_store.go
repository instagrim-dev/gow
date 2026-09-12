package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/instagrim-dev/newf/internal/domain"
)

// InvariantClaimFormRow is one operator-authored claim form (v39, issue #21):
// the proposition shape of a candidate invariant — quantifier + scope + claim
// role — authored explicitly, never inferred from measured coverage. Forms are
// append-only and immutable; the latest form governs refutation semantics at
// challenge time.
type InvariantClaimFormRow struct {
	ID          string
	InvariantID string
	Quantifier  string // universal|recurrent|existential
	ClaimRole   string // regularity|obstruction|enabling_condition|boundary_hypothesis
	Scope       string // authored statement of the population/domain claimed
	AuthoredBy  string // operator
	Basis       string
	CreatedAt   string
}

// PersistInvariantClaimForm appends one authored claim form. The invariant
// must exist; the row is immutable once written.
func (s *Store) PersistInvariantClaimForm(ctx context.Context, row InvariantClaimFormRow) error {
	if err := domain.ValidateInvariantClaimFormID(row.ID); err != nil {
		return err
	}
	if err := domain.ValidateCandidateInvariantID(row.InvariantID); err != nil {
		return err
	}
	var exists int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM candidate_invariants WHERE id = ?`, row.InvariantID).Scan(&exists); err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("%w: candidate invariant %s", ErrNotFound, row.InvariantID)
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO invariant_claim_forms(id, invariant_id, quantifier, claim_role, scope, authored_by, basis, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?)
`, row.ID, row.InvariantID, row.Quantifier, row.ClaimRole, row.Scope, row.AuthoredBy, row.Basis, row.CreatedAt)
	if err != nil {
		return fmt.Errorf("persist invariant claim form: %w", err)
	}
	return nil
}

// GetLatestInvariantClaimForm returns the governing (latest) authored claim
// form for a candidate, reporting absence explicitly: no form means no
// authored quantifier — code must not invent one.
func (s *Store) GetLatestInvariantClaimForm(ctx context.Context, invariantID string) (InvariantClaimFormRow, bool, error) {
	if err := domain.ValidateCandidateInvariantID(invariantID); err != nil {
		return InvariantClaimFormRow{}, false, err
	}
	row := s.db.QueryRowContext(ctx, `
SELECT id, invariant_id, quantifier, claim_role, scope, authored_by, basis, created_at
FROM invariant_claim_forms WHERE invariant_id = ?
ORDER BY created_at DESC, id DESC LIMIT 1
`, invariantID)
	var out InvariantClaimFormRow
	if err := row.Scan(&out.ID, &out.InvariantID, &out.Quantifier, &out.ClaimRole, &out.Scope, &out.AuthoredBy, &out.Basis, &out.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return InvariantClaimFormRow{}, false, nil
		}
		return InvariantClaimFormRow{}, false, err
	}
	return out, true, nil
}
