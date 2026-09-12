package store

import (
	"context"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/domain"
)

// v39 (#21): authored claim forms round-trip, are append-only with the latest
// form governing, are immutable at the SQL layer, and refuse unknown
// invariants and unknown quantifier/role vocabulary.
func TestInvariantClaimFormLatestWinsAndImmutable(t *testing.T) {
	st, _, _, invID := persistSampleInvariant(t)
	ctx := context.Background()

	if _, found, err := st.GetLatestInvariantClaimForm(ctx, invID); err != nil || found {
		t.Fatalf("no form authored yet: found=%v err=%v (absence must be explicit, never a default)", found, err)
	}

	t1 := time.Date(2026, 9, 12, 16, 0, 0, 0, time.UTC)
	first := InvariantClaimFormRow{
		ID: domain.NewInvariantClaimFormID(t1), InvariantID: invID,
		Quantifier: "recurrent", ClaimRole: "regularity",
		Scope: "failure families of the discovery run", AuthoredBy: "operator",
		Basis: "initial authoring", CreatedAt: formatTime(t1),
	}
	if err := st.PersistInvariantClaimForm(ctx, first); err != nil {
		t.Fatalf("persist first form: %v", err)
	}
	t2 := t1.Add(time.Minute)
	second := InvariantClaimFormRow{
		ID: domain.NewInvariantClaimFormID(t2), InvariantID: invID,
		Quantifier: "universal", ClaimRole: "obstruction",
		Scope: "failure families of the discovery run; no claim beyond the corpus", AuthoredBy: "operator",
		Basis: "re-authored after review", CreatedAt: formatTime(t2),
	}
	if err := st.PersistInvariantClaimForm(ctx, second); err != nil {
		t.Fatalf("persist second form: %v", err)
	}

	got, found, err := st.GetLatestInvariantClaimForm(ctx, invID)
	if err != nil || !found {
		t.Fatalf("get latest: found=%v err=%v", found, err)
	}
	if got.ID != second.ID || got.Quantifier != "universal" || got.ClaimRole != "obstruction" {
		t.Fatalf("latest form must govern: %+v", got)
	}

	if _, err := st.db.ExecContext(ctx, `UPDATE invariant_claim_forms SET quantifier = 'existential' WHERE id = ?`, first.ID); err == nil {
		t.Fatal("claim forms must be immutable")
	}
	if _, err := st.db.ExecContext(ctx, `DELETE FROM invariant_claim_forms WHERE id = ?`, first.ID); err == nil {
		t.Fatal("claim forms must be immutable (delete)")
	}

	bad := first
	bad.ID = domain.NewInvariantClaimFormID(t2.Add(time.Minute))
	bad.Quantifier = "most-of-the-time"
	if err := st.PersistInvariantClaimForm(ctx, bad); err == nil {
		t.Fatal("unknown quantifier must be refused")
	}
	missing := first
	missing.ID = domain.NewInvariantClaimFormID(t2.Add(2 * time.Minute))
	missing.InvariantID = "inv_00000000000000000000000000"
	if err := st.PersistInvariantClaimForm(ctx, missing); err == nil {
		t.Fatal("a claim form for an unknown invariant must be refused")
	}
}
