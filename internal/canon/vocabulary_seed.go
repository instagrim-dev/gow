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
	return []*Vocabulary{MechanismV1(), MechanismV2(), MechanismV0Lossy()}
}

func mustBuild(b vocabularyBuilder) *Vocabulary {
	v, err := b.build()
	if err != nil {
		panic("canon: invalid in-repo vocabulary seed " + b.version + ": " + err.Error())
	}
	return v
}
