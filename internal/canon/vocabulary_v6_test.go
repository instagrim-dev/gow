package canon

import (
	"strings"
	"testing"

	"github.com/instagrim-dev/newf/internal/domain"
)

// TestMechanismV6IsSupersetOfV5 guards the revision discipline: every v5 term
// appears in v6 with identical field kind, parent, description, and aliases
// (v6 adds only the pvnp-holdout verbatim corpus canonicalizations).
func TestMechanismV6IsSupersetOfV5(t *testing.T) {
	v5 := MechanismV5()
	v6 := MechanismV6()

	for _, t5 := range v5.Terms("") {
		t6, ok := v6.Term(t5.CanonicalID)
		if !ok {
			t.Fatalf("v5 term %s missing from v6", t5.CanonicalID)
		}
		if t6.FieldKind != t5.FieldKind || t6.Parent != t5.Parent || t6.Description != t5.Description {
			t.Fatalf("v5 term %s changed in v6", t5.CanonicalID)
		}
		kept := map[string]bool{}
		for _, a := range t6.Aliases {
			kept[a] = true
		}
		for _, a := range t5.Aliases {
			if !kept[a] {
				t.Fatalf("v5 term %s lost alias %q in v6", t5.CanonicalID, a)
			}
		}
		if len(t6.Aliases) != len(t5.Aliases) {
			t.Fatalf("v5 term %s gained aliases in v6 (only new terms are allowed)", t5.CanonicalID)
		}
	}
	if len(v6.Terms("")) != len(v5.Terms(""))+len(mechanismV6Terms) {
		t.Fatalf("v6 must add exactly the pvnp corpus-canonicalization terms; v5=%d v6=%d add=%d",
			len(v5.Terms("")), len(v6.Terms("")), len(mechanismV6Terms))
	}
}

// TestMechanismV6VerbatimDiscipline pins the no-merging contract: every v6
// addition carries exactly one alias (the stated label, lowercased), lives in
// the domain.complexity_theory.* namespace, and no two additions of the same
// field kind share an alias — so the corpus's authored distinctions (e.g.
// KI03's algorithm-to-hardness IMPLICATION vs the target's CONVERSION) cannot
// be collapsed by the vocabulary.
func TestMechanismV6VerbatimDiscipline(t *testing.T) {
	seen := map[string]domain.CanonicalID{}
	for _, term := range mechanismV6Terms {
		if !strings.HasPrefix(string(term.CanonicalID), "domain.complexity_theory.") {
			t.Fatalf("v6 addition %s outside the complexity_theory namespace", term.CanonicalID)
		}
		if len(term.Aliases) != 1 {
			t.Fatalf("v6 addition %s must carry exactly one verbatim alias, has %d", term.CanonicalID, len(term.Aliases))
		}
		key := string(term.FieldKind) + "\x00" + Normalize(term.Aliases[0])
		if prev, dup := seen[key]; dup {
			t.Fatalf("v6 additions %s and %s share alias %q on field %s (merging is forbidden)",
				prev, term.CanonicalID, term.Aliases[0], term.FieldKind)
		}
		seen[key] = term.CanonicalID
	}
}

// TestMechanismV6ResolvesPvnpStatedLabels pins representative decisive-field
// labels from the pvnp-holdout payloads: each resolves under v6 and none
// resolved under v5, so every prior pinned substrate is untouched by
// construction. The KI03-vs-target operator distinction (honesty constraint 1
// of the pvnp SCOPE) is asserted explicitly: implication and conversion
// resolve to distinct canonical IDs.
func TestMechanismV6ResolvesPvnpStatedLabels(t *testing.T) {
	v6 := MechanismV6()
	v5 := MechanismV5()

	cases := []struct {
		kind  domain.FieldKind
		label string
		want  domain.CanonicalID
	}{
		{domain.FieldOperator, "diagonalization", "domain.complexity_theory.operator.diagonalization"},
		{domain.FieldOperator, "algorithm-to-hardness implication", "domain.complexity_theory.operator.algorithm_to_hardness_implication"},
		{domain.FieldOperator, "algorithm-to-hardness conversion", "domain.complexity_theory.operator.algorithm_to_hardness_conversion"},
		{domain.FieldOperator, "succinct witness compression", "domain.complexity_theory.operator.succinct_witness_compression"},
		{domain.FieldOperator, "hierarchy contradiction", "domain.complexity_theory.operator.hierarchy_contradiction"},
		{domain.FieldAssumption, "a supplied SAT algorithm with superpolynomial savings over brute force", "domain.complexity_theory.assumption.a_supplied_sat_algorithm_with_superpolynomial_savings_over_brute_force"},
		{domain.FieldPreserves, "black-box relativizing simulation", "domain.complexity_theory.property.black_box_relativizing_simulation"},
		{domain.FieldPreserves, "model-analysis-first direction", "domain.complexity_theory.property.model_analysis_first_direction"},
		{domain.FieldBreaks, "model-analysis-first direction", "domain.complexity_theory.broken.model_analysis_first_direction"},
		{domain.FieldBreaks, "large constructive distinguishing property", "domain.complexity_theory.broken.large_constructive_distinguishing_property"},
		{domain.FieldAuxiliaryObject, "mild-savings circuit-analysis algorithm", "domain.complexity_theory.auxiliary_object.mild_savings_circuit_analysis_algorithm"},
		{domain.FieldAuxiliaryObject, "oracle A with P^A = NP^A", "domain.complexity_theory.auxiliary_object.oracle_a_with_p_a_np_a"},
		{domain.FieldRepresentation, "nondeterministic time hierarchy", "domain.complexity_theory.representation.nondeterministic_time_hierarchy"},
	}
	for _, c := range cases {
		got := v6.Resolve(c.kind, c.label, false)
		if got.State != domain.ResolutionResolved || got.CanonicalID != c.want {
			t.Fatalf("v6 resolve(%s, %q) = %+v, want resolved %s", c.kind, c.label, got, c.want)
		}
		prior := v5.Resolve(c.kind, c.label, false)
		if prior.State == domain.ResolutionResolved {
			t.Fatalf("label %q already resolved under v5 (%s); v6 must add only new resolutions", c.label, prior.CanonicalID)
		}
	}

	imp := v6.Resolve(domain.FieldOperator, "algorithm-to-hardness implication", false)
	conv := v6.Resolve(domain.FieldOperator, "algorithm-to-hardness conversion", false)
	if imp.CanonicalID == conv.CanonicalID {
		t.Fatal("KI03 implication and target conversion collapsed to one canonical ID; honesty constraint 1 violated")
	}
}
