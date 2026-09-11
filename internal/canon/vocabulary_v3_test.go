package canon

import (
	"testing"

	"github.com/instagrim-dev/newf/internal/domain"
)

// TestMechanismV3IsSupersetOfV2 guards the revision discipline: every v2 term
// appears in v3 with identical field kind, parent, and description, and every
// v2 alias is retained (v3 may add aliases only where documented in
// mechanismV3ExtraAliases, and may add new terms).
func TestMechanismV3IsSupersetOfV2(t *testing.T) {
	v2 := MechanismV2()
	v3 := MechanismV3()

	for _, t2 := range v2.Terms("") {
		t3, ok := v3.Term(t2.CanonicalID)
		if !ok {
			t.Fatalf("v2 term %s missing from v3", t2.CanonicalID)
		}
		if t3.FieldKind != t2.FieldKind || t3.Parent != t2.Parent || t3.Description != t2.Description {
			t.Fatalf("v2 term %s changed in v3", t2.CanonicalID)
		}
		kept := map[string]bool{}
		for _, a := range t3.Aliases {
			kept[a] = true
		}
		for _, a := range t2.Aliases {
			if !kept[a] {
				t.Fatalf("v2 term %s lost alias %q in v3", t2.CanonicalID, a)
			}
		}
		added := len(t3.Aliases) - len(t2.Aliases)
		if added > 0 && len(mechanismV3ExtraAliases[t2.CanonicalID]) != added {
			t.Fatalf("v2 term %s gained undocumented aliases in v3", t2.CanonicalID)
		}
	}
	if len(v3.Terms("")) != len(v2.Terms(""))+len(mechanismV3Terms) {
		t.Fatalf("v3 must add exactly the target-canonicalization terms; v2=%d v3=%d add=%d",
			len(v2.Terms("")), len(v3.Terms("")), len(mechanismV3Terms))
	}
}

// TestMechanismV3ResolvesTargetStatedLabels pins the recovery-calibration
// canonicalization: every decisive-field label the withheld target's payload
// states verbatim resolves under v3 (and none resolved under v2, so the train
// substrate that was mined/challenged under v2 is untouched by construction).
func TestMechanismV3ResolvesTargetStatedLabels(t *testing.T) {
	v3 := MechanismV3()
	v2 := MechanismV2()

	cases := []struct {
		kind  domain.FieldKind
		label string
		want  domain.CanonicalID
	}{
		{domain.FieldOperator, "lattice enumeration", "core.operator.lattice_enumeration"},
		{domain.FieldOperator, "geometry of numbers", "core.operator.geometry_of_numbers"},
		{domain.FieldOperator, "convergence proof", "core.operator.convergence_proof"},
		{domain.FieldOperator, "linearization in n", "core.operator.linearization"},
		{domain.FieldPreserves, "positivity of denominators", "domain.number_theory.property.denominator_positivity"},
		{domain.FieldBreaks, "confinement to quadratic non-residues", "domain.number_theory.property.quadratic_nonresidue_confinement"},
		{domain.FieldAssumption, "solution set is an affine class", "core.assumption.affine_class_solution_set"},
		{domain.FieldAssumption, "lattice-point existence is decidable via geometry of numbers", "core.assumption.lattice_point_decidability"},
		{domain.FieldAuxiliaryObject, "Minkowski-style convex body", "core.auxiliary_object.convex_body"},
		{domain.FieldRepresentation, "linear forms in n", "core.representation.linear_forms"},
		{domain.FieldRepresentation, "convex body / positive cone", "core.representation.convex_body"},
		{domain.FieldRepresentation, "affine lattice in Z^3", "core.representation.affine_lattice"},
	}
	for _, c := range cases {
		res := v3.Resolve(c.kind, c.label, false)
		if res.State != domain.ResolutionResolved || res.CanonicalID != c.want {
			t.Fatalf("v3 %s %q: got state=%s id=%s, want resolved %s", c.kind, c.label, res.State, res.CanonicalID, c.want)
		}
		if r2 := v2.Resolve(c.kind, c.label, false); r2.State == domain.ResolutionResolved {
			t.Fatalf("label %q must not have resolved under v2 (train substrate must be unaffected)", c.label)
		}
	}

	// Field-kind scoping: the breaks-kind QR term must not leak into preserves
	// (where the v2 property concept owns the label) or vice versa.
	if r := v3.Resolve(domain.FieldPreserves, "confined to quadratic nonresidues", false); r.CanonicalID != "domain.number_theory.property.confined_to_quadratic_nonresidues" {
		t.Fatalf("preserves label must stay with the v2 property concept, got %s", r.CanonicalID)
	}
	if r := v3.Resolve(domain.FieldBreaks, "confined to quadratic nonresidues", false); r.CanonicalID != "domain.number_theory.property.quadratic_nonresidue_confinement" {
		t.Fatalf("breaks label must resolve to the breaks-kind term, got %s (%s)", r.CanonicalID, r.State)
	}
}
