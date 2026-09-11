package canon

import (
	"testing"

	"github.com/instagrim-dev/newf/internal/domain"
)

// TestMechanismV2IsStrictSupersetOfV1 guards the immutability contract from the
// pilot-001 mapping review: every mechanism/v1 term must appear in mechanism/v2
// with an identical field kind, alias set, and parent. v2 may only ADD terms.
func TestMechanismV2IsStrictSupersetOfV1(t *testing.T) {
	v1 := MechanismV1()
	v2 := MechanismV2()

	for _, t1 := range v1.Terms("") {
		t2, ok := v2.Term(t1.CanonicalID)
		if !ok {
			t.Fatalf("v1 term %s missing from v2", t1.CanonicalID)
		}
		if t2.FieldKind != t1.FieldKind || t2.Parent != t1.Parent || t2.Description != t1.Description {
			t.Fatalf("v1 term %s changed in v2", t1.CanonicalID)
		}
		if len(t2.Aliases) != len(t1.Aliases) {
			t.Fatalf("v1 term %s alias set changed in v2", t1.CanonicalID)
		}
		for i := range t1.Aliases {
			if t1.Aliases[i] != t2.Aliases[i] {
				t.Fatalf("v1 term %s alias %d changed in v2", t1.CanonicalID, i)
			}
		}
	}
	if len(v2.Terms("")) != len(v1.Terms(""))+3 {
		t.Fatalf("v2 must add exactly the 3 adjudicated property concepts; v1=%d v2=%d",
			len(v1.Terms("")), len(v2.Terms("")))
	}
}

// TestMechanismV2ResolvesAdjudicatedProperties confirms the three accepted
// ledger properties resolve under v2 — including the two source labels that
// NAME properties directly (es-08 for L1, es-02 for L4) — while es-12's
// form-indexed variant stays unresolved (a preserved distinction, L4 contrast).
func TestMechanismV2ResolvesAdjudicatedProperties(t *testing.T) {
	v2 := MechanismV2()

	cases := []struct {
		label string
		want  domain.CanonicalID
	}{
		{"confined to quadratic nonresidues", "domain.number_theory.property.confined_to_quadratic_nonresidues"},
		{"confinement of congruence methods to non-residues", "domain.number_theory.property.confined_to_quadratic_nonresidues"},
		{"class union construction", "domain.number_theory.property.class_union_construction"},
		{"identity carried solvability", "domain.number_theory.property.identity_carried_solvability"},
		{"polynomial-identity solvability per class", "domain.number_theory.property.identity_carried_solvability"},
	}
	for _, c := range cases {
		res := v2.Resolve(domain.FieldPreserves, c.label, false)
		if res.State != domain.ResolutionResolved || res.CanonicalID != c.want {
			t.Fatalf("label %q: got state=%s id=%s, want resolved %s", c.label, res.State, res.CanonicalID, c.want)
		}
	}

	// The form-indexed es-12 variant must NOT resolve: merging it would erase
	// the by-form vs by-producing-congruence distinction the ledger preserved.
	res := v2.Resolve(domain.FieldPreserves, "polynomial-identity solvability per form", false)
	if res.State != domain.ResolutionUnknown {
		t.Fatalf("es-12 form-indexed label must stay unknown under v2, got %s (%s)", res.State, res.CanonicalID)
	}

	// None of the property labels may resolve under immutable v1.
	v1 := MechanismV1()
	for _, c := range cases {
		if r := v1.Resolve(domain.FieldPreserves, c.label, false); r.State == domain.ResolutionResolved {
			t.Fatalf("label %q must not resolve under v1", c.label)
		}
	}
}
