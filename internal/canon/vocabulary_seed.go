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
	// VocabularyMechanismV4 is the pilot-004 N2a challenge revision: a strict
	// superset of mechanism/v3 adding the density_averaging_ceiling
	// GeneratedInterpretation property. A seven-probe challenge campaign was
	// recorded (model-assisted, operator-recorded; see
	// corpus/experiments/pilot-004-discovery/records/N2a-challenge.md and its
	// attributed reassessment): the property survived at model-judgment
	// strength with one required refinement (rate-ceiling vs reach-ceiling
	// boundary delta at C4); C2 produced no completed counterexample search
	// and C7 remains a proposed explanation, not a checked mathematical
	// argument.
	// Scope: es-05 (averaging/Bombieri–Vinogradov) and es-10 (Vaughan
	// congruence density), both partial_failure. The claim is that each
	// method's almost-all statement has an intrinsic ceiling that cannot be
	// closed within the method's own machinery: es-05 because sieve-average
	// growth is polylogarithmic (subthreshold for existence-forcing); es-10
	// because QR-survivor classes provide an absolute floor (the QR
	// obstruction). Both sub-mechanisms are recorded as scope notes; the
	// single canonical ID covers both.
	VocabularyMechanismV4 = "mechanism/v4"
	// VocabularyMechanismV5 is the pilot-004 N5 challenge revision: a strict
	// superset of mechanism/v4 adding the reorganisation_without_qr_existence
	// GeneratedInterpretation property. A seven-probe challenge campaign was
	// recorded (model-assisted, operator-recorded; see
	// corpus/experiments/pilot-004-discovery/records/N5-challenge.md and its
	// attributed reassessment): six probes were consistent with the property
	// at model-judgment strength; C3 (success discrimination) was
	// inapplicable by corpus construction and remains an open obligation.
	// Scope: es-09 (higher-dimensional variety lift) and es-11
	// (Monks–Velingker structural analysis), both partial_success.
	// The claim is that reorganisation/re-representation operators deliver
	// structural improvements but add no new existence at the QR survivor
	// classes; the proposed explanation attributes the absence to the
	// operator type (reorganisation cannot generate positivity) — a
	// source-grounded proposal for the two in-scope members, not an
	// established general impossibility. C4 found that the property holds
	// independently at two structural levels (parameterisation-family
	// unification for es-09; solution-constraint geometry for es-11),
	// which strengthens rather than splits the claim. C3 caveat: the
	// property is non-discriminating within partial_success by construction;
	// success-contrast is untestable until a success outcome appears.
	VocabularyMechanismV5 = "mechanism/v5"
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

// mechanismV4Terms are the pilot-004 N2a challenge additions layered on top of
// mechanism/v3 to form mechanism/v4. A single GeneratedInterpretation property
// covers the intrinsic density/averaging ceiling shared by es-05 and es-10.
// The two sub-mechanisms differ concretely (rate-subthreshold for es-05;
// QR-absolute-floor for es-10); the canonical ID covers both, and the scope
// notes in the description record that distinction. Source occurrences:
// pilot-004 entries E13/E16/E20/E23, all unanimous (3-0). Challenge record:
// corpus/experiments/pilot-004-discovery/records/N2a-challenge.md.
var mechanismV4Terms = []Term{
	{
		CanonicalID: "domain.number_theory.property.density_averaging_ceiling",
		FieldKind:   domain.FieldPreserves,
		Description: "GeneratedInterpretation (pilot-004 N2a challenge, admitted): each of these methods produces a monotone-improving almost-all statement whose gap to universality cannot be closed within the method's own machinery. Scope: es-05 (averaging/Bombieri–Vinogradov — ceiling because sieve-average growth is polylogarithmic, below the polynomial threshold required to force existence); es-10 (Vaughan congruence density — ceiling because QR-survivor classes provide an absolute floor that no finite congruence battery can remove). Note on es-10 and L1: es-10 also carries confined_to_quadratic_nonresidues (L1), which identifies the same QR wall from the reach direction (congruence relations cannot touch QR classes). density_averaging_ceiling adds the rate/density framing: the exceptional-set density converges toward zero but has a positive lower bound — a statement about rate of improvement, not just confinement of reach. The two properties are informationally distinct in framing; for es-10 both apply. The two sub-mechanisms are distinct; the property covers both.",
		Aliases: []string{
			"density averaging ceiling",
			"intrinsic density ceiling",
			"averaging ceiling",
			"almost all ceiling",
		},
	},
}

// MechanismV4 builds the mechanism/v4 vocabulary: every mechanism/v3 term
// unchanged plus the pilot-004 N2a density_averaging_ceiling property.
func MechanismV4() *Vocabulary {
	base := append(append([]Term{}, mechanismV1Seed.terms...), mechanismV2Terms...)
	base = append(base, mechanismV3Terms...)
	// Apply v3 extra aliases (same as MechanismV3 does).
	terms := make([]Term, 0, len(base)+len(mechanismV4Terms))
	for _, t := range base {
		if extra, ok := mechanismV3ExtraAliases[t.CanonicalID]; ok {
			t.Aliases = append(append([]string{}, t.Aliases...), extra...)
		}
		terms = append(terms, t)
	}
	terms = append(terms, mechanismV4Terms...)
	seed := vocabularyBuilder{
		version:  VocabularyMechanismV4,
		terms:    terms,
		rejected: append([]string{}, mechanismV1Seed.rejected...),
	}
	return mustBuild(seed)
}

// mechanismV5Terms are the pilot-004 N5 challenge additions layered on top of
// mechanism/v4 to form mechanism/v5. A single GeneratedInterpretation property
// covers the reorganisation-without-QR-existence shared feature of es-09 and
// es-11. The two approaches reorganise at different structural levels
// (parameterisation-family unification vs solution-constraint geometry), but
// both independently fail to generate QR existence; the C4 boundary_delta
// strengthens rather than splits the claim. See challenge record:
// corpus/experiments/pilot-004-discovery/records/N5-challenge.md.
var mechanismV5Terms = []Term{
	{
		CanonicalID: "domain.number_theory.property.reorganisation_without_qr_existence",
		FieldKind:   domain.FieldPreserves,
		Description: "GeneratedInterpretation (pilot-004 N5 challenge, admitted): reorganisation/re-representation operators deliver structural improvements (unification, constraint geometry) but add no new existence at the QR survivor classes. The absence is causal: reorganisation cannot generate positivity. Scope: es-09 (parameterisation-family unification via algebraic variety lift) and es-11 (solution-constraint analysis), both partial_success. C4 boundary_delta: the property holds independently at parameterisation-family level (es-09) and solution-constraint level (es-11). C3 caveat: non-discriminating within partial_success by construction; success-contrast untestable until a success outcome appears in corpus.",
		Aliases: []string{
			"reorganisation without qr existence",
			"reorganisation without existence",
			"structural reorganisation adds no positivity",
		},
	},
}

// MechanismV5 builds the mechanism/v5 vocabulary: every mechanism/v4 term
// unchanged plus the pilot-004 N5 reorganisation_without_qr_existence property.
func MechanismV5() *Vocabulary {
	base := append(append([]Term{}, mechanismV1Seed.terms...), mechanismV2Terms...)
	base = append(base, mechanismV3Terms...)
	base = append(base, mechanismV4Terms...)
	terms := make([]Term, 0, len(base)+len(mechanismV5Terms))
	for _, t := range base {
		if extra, ok := mechanismV3ExtraAliases[t.CanonicalID]; ok {
			t.Aliases = append(append([]string{}, t.Aliases...), extra...)
		}
		terms = append(terms, t)
	}
	terms = append(terms, mechanismV5Terms...)
	seed := vocabularyBuilder{
		version:  VocabularyMechanismV5,
		terms:    terms,
		rejected: append([]string{}, mechanismV1Seed.rejected...),
	}
	return mustBuild(seed)
}

// SeededVocabularies returns every in-repo vocabulary, used to seed persistence.
func SeededVocabularies() []*Vocabulary {
	return []*Vocabulary{MechanismV1(), MechanismV2(), MechanismV3(), MechanismV4(), MechanismV5(), MechanismV0Lossy()}
}

func mustBuild(b vocabularyBuilder) *Vocabulary {
	v, err := b.build()
	if err != nil {
		panic("canon: invalid in-repo vocabulary seed " + b.version + ": " + err.Error())
	}
	return v
}
