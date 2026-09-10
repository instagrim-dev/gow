package canon

import "github.com/instagrim-dev/newf/internal/domain"

// Vocabulary version identifiers.
const (
	// VocabularyMechanismV1 is the canonical semantic spine #9 ships with.
	VocabularyMechanismV1 = "mechanism/v1"
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
	return []*Vocabulary{MechanismV1(), MechanismV0Lossy()}
}

func mustBuild(b vocabularyBuilder) *Vocabulary {
	v, err := b.build()
	if err != nil {
		panic("canon: invalid in-repo vocabulary seed " + b.version + ": " + err.Error())
	}
	return v
}
