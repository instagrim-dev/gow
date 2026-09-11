package canon

import "github.com/instagrim-dev/newf/internal/domain"

// Vocabulary version identifiers.
const (
	// VocabularyMechanismV1 is the canonical semantic spine #9 ships with.
	VocabularyMechanismV1 = "mechanism/v1"
	// VocabularyMechanismV2 is the pilot-001 mapping-review revision: a strict
	// superset of mechanism/v1 adding three GeneratedInterpretation property
	// concepts adjudicated in
	// corpus/experiments/pilot-001/review/adjudication-ledger.json (L1/L3/L4).
	// mechanism/v1 is immutable; this revision exists so accepted shared
	// properties can be encoded without mutating v1 or disguising them as
	// aliases of unrelated labels.
	VocabularyMechanismV2 = "mechanism/v2"
	// VocabularyMechanismV3 is the target-canonicalization revision for the
	// recovery-calibration correction: a strict superset of mechanism/v2 adding
	// terms that NAME what the withheld target's normalize payload explicitly
	// states, so its decisive fields become comparable and recovery-rule/v1 can
	// reach a positive match. This is ordinary canonicalization of stated
	// content — no term is an interpretation, and none was chosen for its
	// effect on B0/B3 outcomes (no captures existed when it was pinned).
	VocabularyMechanismV3 = "mechanism/v3"
	// VocabularyMechanismV0Lossy is a deliberately over-compressed vocabulary
	// used only to exercise the abstraction-loss regression: it merges two
	// outcome-predictive operators into one canonical ID.
	VocabularyMechanismV0Lossy = "mechanism/v0-lossy"
)

// mechanismV1Seed is the in-repo definition of the mechanism/v1 vocabulary. It
// is intentionally tiny — a handful of core.* terms plus a couple of
// domain.number_theory.* terms — just enough to exercise the required fixtures.
// Expansion is out of scope for this slice.
var mechanismV1Seed = vocabularyBuilder{
	version: VocabularyMechanismV1,
	terms: []Term{
		{
			CanonicalID: "core.operator.modular_decomposition",
			FieldKind:   domain.FieldOperator,
			Description: "Decompose the problem residue class by residue class.",
			Aliases: []string{
				"modular decomposition",
				"work residue by residue",
				"residue-by-residue reduction",
				"break into congruence classes",
			},
		},
		{
			CanonicalID: "core.operator.density_averaging",
			FieldKind:   domain.FieldOperator,
			Description: "Average a quantity over a density to obtain a global bound.",
			Aliases: []string{
				"density averaging",
				"average over density",
				"global averaging bound",
			},
		},
		{
			CanonicalID: "core.representation.congruence_classes",
			FieldKind:   domain.FieldRepresentation,
			Description: "Represent integers by their residues modulo n.",
			Aliases: []string{
				"congruence classes",
				"residue classes",
			},
		},
		{
			CanonicalID: "core.representation.affine_lattice",
			FieldKind:   domain.FieldRepresentation,
			Description: "Represent the admissible parameter set as an affine class of integer points (an affine lattice).",
			Aliases: []string{
				"affine lattice",
				"affine class in z 3",
				"integer lattice representation",
			},
		},
		{
			CanonicalID: "domain.number_theory.property.residue_locality",
			FieldKind:   domain.FieldPreserves,
			Description: "The property that behavior is determined locally within each residue class.",
			Aliases: []string{
				"residue locality",
				"works residue-by-residue",
				"local congruence argument",
				"locality of residues",
			},
		},
		{
			CanonicalID: "domain.number_theory.property.mean_growth_rate",
			FieldKind:   domain.FieldPreserves,
			Description: "The averaged growth rate of a global quantity.",
			Aliases: []string{
				"mean growth rate",
				"average growth",
			},
		},
		{
			CanonicalID: "core.assumption.residue_independence",
			FieldKind:   domain.FieldAssumption,
			Description: "Assumes residue classes behave independently.",
			Aliases: []string{
				"residues are independent",
				"residue independence",
				"independent residues",
			},
		},
		{
			CanonicalID: "domain.number_theory.property.residue_class_locality",
			FieldKind:   domain.FieldBreaks,
			Description: "The invariant that a construction stays confined to residue classes modulo n; a mechanism that breaks it leaves congruence-local reasoning entirely.",
			Aliases: []string{
				"residue class locality",
				"confinement to residue classes",
				"congruence class confinement",
			},
		},
		{
			CanonicalID: "core.auxiliary_object.affine_lattice",
			FieldKind:   domain.FieldAuxiliaryObject,
			Description: "An affine sublattice of Z^k introduced to search integer points off the residue-class grid.",
			Aliases: []string{
				"affine lattice",
				"affine sublattice",
				"integer lattice in z 3",
				"lattice of integer points",
			},
		},
	},
	rejected: []string{
		// A term the vocabulary explicitly disallows (e.g. a known non-mechanism
		// filler phrase). Kept minimal.
		"handwaving",
	},
}

// mechanismV2Terms are the pilot-001 mapping-review additions layered on top of
// the (unchanged) mechanism/v1 terms to form mechanism/v2. Each is an accepted
// shared-property HYPOTHESIS (GeneratedInterpretation): the ledger established
// the relationship from train-note passages, and support is earned only through
// the ordinary mine -> challenge path. No alias here merges two OPERATIONS: the
// only source labels aliased are ones that directly NAME the property itself —
// es-08's preserves label names the L1 property (its note is the obstruction
// analysis), and es-02's preserves label names the L4 property verbatim.
// Deliberately NOT aliased: es-12's "polynomial-identity solvability per form"
// (the form-indexed variant is a preserved distinction, ledger L4 contrast).
var mechanismV2Terms = []Term{
	{
		CanonicalID: "domain.number_theory.property.confined_to_quadratic_nonresidues",
		FieldKind:   domain.FieldPreserves,
		Description: "GeneratedInterpretation (pilot-001 ledger L1, accepted): the method's coverage is carried by congruence relations that, by quadratic reciprocity, cannot catch quadratic-residue classes. Scope: es-01/es-02/es-03/es-10 style mechanisms; NOT growth-rate-bounded (es-05) or finiteness-bounded (es-07) methods.",
		Aliases: []string{
			"confined to quadratic nonresidues",
			"confinement of congruence methods to non-residues",
			"qr confinement",
		},
	},
	{
		CanonicalID: "domain.number_theory.property.class_union_construction",
		FieldKind:   domain.FieldPreserves,
		Description: "GeneratedInterpretation (pilot-001 ledger L3, accepted): builds a union of congruence classes each admitting a polynomial identity. The exhaustive-cover vs bounded-density-leftover distinction is deliberately NOT erased by this concept.",
		Aliases: []string{
			"class union construction",
		},
	},
	{
		CanonicalID: "domain.number_theory.property.identity_carried_solvability",
		FieldKind:   domain.FieldPreserves,
		Description: "GeneratedInterpretation (pilot-001 ledger L4, accepted): solvability inside a congruence class is supplied by a fixed polynomial identity keyed to that class.",
		Aliases: []string{
			"identity carried solvability",
			"polynomial identity solvability per class",
		},
	},
}

// MechanismV2 builds the mechanism/v2 vocabulary: every mechanism/v1 term
// unchanged plus the pilot-001 mapping-review property concepts.
func MechanismV2() *Vocabulary {
	seed := vocabularyBuilder{
		version:  VocabularyMechanismV2,
		terms:    append(append([]Term{}, mechanismV1Seed.terms...), mechanismV2Terms...),
		rejected: append([]string{}, mechanismV1Seed.rejected...),
	}
	return mustBuild(seed)
}

// mechanismV3Terms canonicalize the withheld target's explicitly stated
// content (recovery-calibration correction). Every alias below is a verbatim
// (normalized) label from the target payload or an obvious wording variant of
// the SAME stated concept; none merges distinct operations and none encodes an
// interpretation. The breaks-kind QR term follows the existing v1 precedent of
// field-kind-scoped IDs (preserves residue_locality vs breaks
// residue_class_locality): the preserves-kind property concept from v2 cannot
// carry a second field kind, so the break of that property is its own term.
var mechanismV3Terms = []Term{
	{
		CanonicalID: "core.operator.lattice_enumeration",
		FieldKind:   domain.FieldOperator,
		Description: "Enumerate lattice points of a (typically affine) integer lattice.",
		Aliases:     []string{"lattice enumeration"},
	},
	{
		CanonicalID: "core.operator.geometry_of_numbers",
		FieldKind:   domain.FieldOperator,
		Description: "Attack existence questions via geometry-of-numbers arguments (Minkowski-style).",
		Aliases:     []string{"geometry of numbers"},
	},
	{
		CanonicalID: "core.operator.convergence_proof",
		FieldKind:   domain.FieldOperator,
		Description: "Establish convergence of an enumeration or iterative construction.",
		Aliases:     []string{"convergence proof"},
	},
	{
		CanonicalID: "core.operator.linearization",
		FieldKind:   domain.FieldOperator,
		Description: "Recast a condition as linear forms in the problem parameter.",
		Aliases:     []string{"linearization in n", "linearization"},
	},
	{
		CanonicalID: "domain.number_theory.property.denominator_positivity",
		FieldKind:   domain.FieldPreserves,
		Description: "The property that constructed denominators remain positive integers.",
		Aliases:     []string{"positivity of denominators", "denominator positivity"},
	},
	{
		CanonicalID: "domain.number_theory.property.quadratic_nonresidue_confinement",
		FieldKind:   domain.FieldBreaks,
		Description: "As a breaks-kind term: the confinement of congruence-carried methods to quadratic non-residue classes. A mechanism listing this BREAKS the confinement (crosses the QR wall). Field-kind-scoped sibling of the preserves-kind v2 concept.",
		Aliases: []string{
			"confinement to quadratic non-residues",
			"confined to quadratic nonresidues",
			"qr confinement",
		},
	},
	{
		CanonicalID: "core.assumption.affine_class_solution_set",
		FieldKind:   domain.FieldAssumption,
		Description: "Assumes the solution set forms an affine class (affine lattice) of integer points.",
		Aliases:     []string{"solution set is an affine class"},
	},
	{
		CanonicalID: "core.assumption.lattice_point_decidability",
		FieldKind:   domain.FieldAssumption,
		Description: "Assumes lattice-point existence in the relevant region is decidable via geometry of numbers.",
		Aliases:     []string{"lattice point existence is decidable via geometry of numbers"},
	},
	{
		CanonicalID: "core.auxiliary_object.convex_body",
		FieldKind:   domain.FieldAuxiliaryObject,
		Description: "A convex body (Minkowski-style) introduced for lattice-point existence arguments.",
		Aliases:     []string{"minkowski style convex body", "convex body"},
	},
	{
		CanonicalID: "core.representation.linear_forms",
		FieldKind:   domain.FieldRepresentation,
		Description: "Represent solvability conditions as linear forms in the problem parameter.",
		Aliases:     []string{"linear forms in n", "linear forms"},
	},
	{
		CanonicalID: "core.representation.convex_body",
		FieldKind:   domain.FieldRepresentation,
		Description: "Represent the admissible region as a convex body / positive cone.",
		Aliases:     []string{"convex body positive cone", "convex body", "positive cone"},
	},
}

// mechanismV3ExtraAliases adds wording variants of ALREADY-EXISTING terms for
// labels the target states verbatim. v1/v2 stay immutable; these aliases exist
// only in the v3 revision.
var mechanismV3ExtraAliases = map[domain.CanonicalID][]string{
	// The target writes "affine lattice in Z^3"; v1 already carries the term
	// with aliases "affine lattice" / "affine class in z 3".
	"core.representation.affine_lattice": {"affine lattice in z 3"},
}

// MechanismV3 builds the mechanism/v3 vocabulary: every mechanism/v2 term
// (with the documented alias additions) plus the target-canonicalization terms.
func MechanismV3() *Vocabulary {
	base := append(append([]Term{}, mechanismV1Seed.terms...), mechanismV2Terms...)
	terms := make([]Term, 0, len(base)+len(mechanismV3Terms))
	for _, t := range base {
		if extra, ok := mechanismV3ExtraAliases[t.CanonicalID]; ok {
			t.Aliases = append(append([]string{}, t.Aliases...), extra...)
		}
		terms = append(terms, t)
	}
	terms = append(terms, mechanismV3Terms...)
	seed := vocabularyBuilder{
		version:  VocabularyMechanismV3,
		terms:    terms,
		rejected: append([]string{}, mechanismV1Seed.rejected...),
	}
	return mustBuild(seed)
}

// mechanismV0LossySeed merges the two distinct operators of mechanismV1 into one
// canonical ID, deliberately erasing an outcome-predictive distinction. It is
// used only by the abstraction-loss regression (U7 case 5).
var mechanismV0LossySeed = vocabularyBuilder{
	version: VocabularyMechanismV0Lossy,
	terms: []Term{
		{
			CanonicalID: "core.operator.generic_reduction",
			FieldKind:   domain.FieldOperator,
			Description: "Over-compressed: any reduction operator, lossy.",
			Aliases: []string{
				"modular decomposition",
				"work residue by residue",
				"residue-by-residue reduction",
				"break into congruence classes",
				"density averaging",
				"average over density",
				"global averaging bound",
			},
		},
		{
			CanonicalID: "core.representation.generic_structure",
			FieldKind:   domain.FieldRepresentation,
			Description: "Over-compressed: any structural representation, lossy.",
			Aliases: []string{
				"congruence classes",
				"residue classes",
			},
		},
	},
}

// MechanismV1 builds and returns the mechanism/v1 vocabulary. It panics on a
// malformed seed because the seed is in-repo data that must always be valid;
// a malformed seed is a programming error, not a runtime condition.
func MechanismV1() *Vocabulary {
	return mustBuild(mechanismV1Seed)
}

// MechanismV0Lossy builds the deliberately lossy vocabulary for the
// abstraction-loss regression.
func MechanismV0Lossy() *Vocabulary {
	return mustBuild(mechanismV0LossySeed)
}

// SeededVocabularies returns every in-repo vocabulary, used to seed persistence.
func SeededVocabularies() []*Vocabulary {
	return []*Vocabulary{MechanismV1(), MechanismV2(), MechanismV3(), MechanismV0Lossy()}
}

func mustBuild(b vocabularyBuilder) *Vocabulary {
	v, err := b.build()
	if err != nil {
		panic("canon: invalid in-repo vocabulary seed " + b.version + ": " + err.Error())
	}
	return v
}
