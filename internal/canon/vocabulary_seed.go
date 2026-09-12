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
	// VocabularyMechanismV6 is the pvnp-holdout corpus-canonicalization
	// revision: a strict superset of mechanism/v5 adding VERBATIM
	// canonicalizations of every label the pvnp-holdout train and target
	// normalize payloads explicitly state (corpus/experiments/pvnp-holdout,
	// train/pnp-01..12 + target/pnp-target-01..02), one term per distinct
	// (field, label) pair, no cross-label aliasing. This is the same ordinary
	// stated-content canonicalization discipline as mechanism/v3: no term is
	// an interpretation, no two source labels are merged (the corpus's
	// authored distinctions — e.g. KI03's algorithm-to-hardness IMPLICATION
	// vs the target's algorithm-to-hardness CONVERSION — survive exactly),
	// and none was chosen for its effect on arm outcomes (no captures existed
	// when it was pinned; see the pvnp-holdout SCOPE.md freeze record).
	VocabularyMechanismV6 = "mechanism/v6"
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

// mechanismV6Terms are the pvnp-holdout corpus-canonicalization additions
// layered on top of mechanism/v5 to form mechanism/v6. Generated mechanically
// from the embedded newf-normalize payloads: one canonical term per distinct
// (field kind, stated label) pair, alias = the stated label verbatim
// (lowercased), no merging. Field namespaces: operator/assumption/property
// (preserves)/broken (breaks)/auxiliary_object/representation under
// domain.complexity_theory.*.
var mechanismV6Terms = []Term{
	{
		CanonicalID: "domain.complexity_theory.operator.algebraic_geometric_degeneration",
		FieldKind:   domain.FieldOperator,
		Description: "pvnp-holdout stated label (verbatim canonicalization): algebraic-geometric degeneration.",
		Aliases:     []string{"algebraic-geometric degeneration"},
	},
	{
		CanonicalID: "domain.complexity_theory.operator.algorithm_to_hardness_conversion",
		FieldKind:   domain.FieldOperator,
		Description: "pvnp-holdout stated label (verbatim canonicalization): algorithm-to-hardness conversion.",
		Aliases:     []string{"algorithm-to-hardness conversion"},
	},
	{
		CanonicalID: "domain.complexity_theory.operator.algorithm_to_hardness_implication",
		FieldKind:   domain.FieldOperator,
		Description: "pvnp-holdout stated label (verbatim canonicalization): algorithm-to-hardness implication.",
		Aliases:     []string{"algorithm-to-hardness implication"},
	},
	{
		CanonicalID: "domain.complexity_theory.operator.alternation_speedup",
		FieldKind:   domain.FieldOperator,
		Description: "pvnp-holdout stated label (verbatim canonicalization): alternation speedup.",
		Aliases:     []string{"alternation speedup"},
	},
	{
		CanonicalID: "domain.complexity_theory.operator.bottleneck_counting",
		FieldKind:   domain.FieldOperator,
		Description: "pvnp-holdout stated label (verbatim canonicalization): bottleneck counting.",
		Aliases:     []string{"bottleneck counting"},
	},
	{
		CanonicalID: "domain.complexity_theory.operator.circuit_normal_form_transformation",
		FieldKind:   domain.FieldOperator,
		Description: "pvnp-holdout stated label (verbatim canonicalization): circuit normal-form transformation.",
		Aliases:     []string{"circuit normal-form transformation"},
	},
	{
		CanonicalID: "domain.complexity_theory.operator.clocked_simulation",
		FieldKind:   domain.FieldOperator,
		Description: "pvnp-holdout stated label (verbatim canonicalization): clocked simulation.",
		Aliases:     []string{"clocked simulation"},
	},
	{
		CanonicalID: "domain.complexity_theory.operator.completeness_leverage",
		FieldKind:   domain.FieldOperator,
		Description: "pvnp-holdout stated label (verbatim canonicalization): completeness leverage.",
		Aliases:     []string{"completeness leverage"},
	},
	{
		CanonicalID: "domain.complexity_theory.operator.completeness_translation",
		FieldKind:   domain.FieldOperator,
		Description: "pvnp-holdout stated label (verbatim canonicalization): completeness translation.",
		Aliases:     []string{"completeness translation"},
	},
	{
		CanonicalID: "domain.complexity_theory.operator.complexity_transfer",
		FieldKind:   domain.FieldOperator,
		Description: "pvnp-holdout stated label (verbatim canonicalization): complexity transfer.",
		Aliases:     []string{"complexity transfer"},
	},
	{
		CanonicalID: "domain.complexity_theory.operator.degree_counting",
		FieldKind:   domain.FieldOperator,
		Description: "pvnp-holdout stated label (verbatim canonicalization): degree counting.",
		Aliases:     []string{"degree counting"},
	},
	{
		CanonicalID: "domain.complexity_theory.operator.depth_reduction",
		FieldKind:   domain.FieldOperator,
		Description: "pvnp-holdout stated label (verbatim canonicalization): depth reduction.",
		Aliases:     []string{"depth reduction"},
	},
	{
		CanonicalID: "domain.complexity_theory.operator.diagonalization",
		FieldKind:   domain.FieldOperator,
		Description: "pvnp-holdout stated label (verbatim canonicalization): diagonalization.",
		Aliases:     []string{"diagonalization"},
	},
	{
		CanonicalID: "domain.complexity_theory.operator.distance_from_polynomials_argument",
		FieldKind:   domain.FieldOperator,
		Description: "pvnp-holdout stated label (verbatim canonicalization): distance-from-polynomials argument.",
		Aliases:     []string{"distance-from-polynomials argument"},
	},
	{
		CanonicalID: "domain.complexity_theory.operator.dynamic_programming_evaluation",
		FieldKind:   domain.FieldOperator,
		Description: "pvnp-holdout stated label (verbatim canonicalization): dynamic-programming evaluation.",
		Aliases:     []string{"dynamic-programming evaluation"},
	},
	{
		CanonicalID: "domain.complexity_theory.operator.error_counting_on_test_inputs",
		FieldKind:   domain.FieldOperator,
		Description: "pvnp-holdout stated label (verbatim canonicalization): error counting on test inputs.",
		Aliases:     []string{"error counting on test inputs"},
	},
	{
		CanonicalID: "domain.complexity_theory.operator.fast_rectangular_matrix_multiplication",
		FieldKind:   domain.FieldOperator,
		Description: "pvnp-holdout stated label (verbatim canonicalization): fast rectangular matrix multiplication.",
		Aliases:     []string{"fast rectangular matrix multiplication"},
	},
	{
		CanonicalID: "domain.complexity_theory.operator.gate_by_gate_approximation",
		FieldKind:   domain.FieldOperator,
		Description: "pvnp-holdout stated label (verbatim canonicalization): gate-by-gate approximation.",
		Aliases:     []string{"gate-by-gate approximation"},
	},
	{
		CanonicalID: "domain.complexity_theory.operator.hierarchy_contradiction",
		FieldKind:   domain.FieldOperator,
		Description: "pvnp-holdout stated label (verbatim canonicalization): hierarchy contradiction.",
		Aliases:     []string{"hierarchy contradiction"},
	},
	{
		CanonicalID: "domain.complexity_theory.operator.hierarchy_theorem_coupling",
		FieldKind:   domain.FieldOperator,
		Description: "pvnp-holdout stated label (verbatim canonicalization): hierarchy-theorem coupling.",
		Aliases:     []string{"hierarchy-theorem coupling"},
	},
	{
		CanonicalID: "domain.complexity_theory.operator.independence_proof",
		FieldKind:   domain.FieldOperator,
		Description: "pvnp-holdout stated label (verbatim canonicalization): independence proof.",
		Aliases:     []string{"independence proof"},
	},
	{
		CanonicalID: "domain.complexity_theory.operator.indirect_diagonalization",
		FieldKind:   domain.FieldOperator,
		Description: "pvnp-holdout stated label (verbatim canonicalization): indirect diagonalization.",
		Aliases:     []string{"indirect diagonalization"},
	},
	{
		CanonicalID: "domain.complexity_theory.operator.machine_enumeration",
		FieldKind:   domain.FieldOperator,
		Description: "pvnp-holdout stated label (verbatim canonicalization): machine enumeration.",
		Aliases:     []string{"machine enumeration"},
	},
	{
		CanonicalID: "domain.complexity_theory.operator.multiplicity_comparison",
		FieldKind:   domain.FieldOperator,
		Description: "pvnp-holdout stated label (verbatim canonicalization): multiplicity comparison.",
		Aliases:     []string{"multiplicity comparison"},
	},
	{
		CanonicalID: "domain.complexity_theory.operator.oracle_construction",
		FieldKind:   domain.FieldOperator,
		Description: "pvnp-holdout stated label (verbatim canonicalization): oracle construction.",
		Aliases:     []string{"oracle construction"},
	},
	{
		CanonicalID: "domain.complexity_theory.operator.polynomial_approximation",
		FieldKind:   domain.FieldOperator,
		Description: "pvnp-holdout stated label (verbatim canonicalization): polynomial approximation.",
		Aliases:     []string{"polynomial approximation"},
	},
	{
		CanonicalID: "domain.complexity_theory.operator.polynomial_method_extension",
		FieldKind:   domain.FieldOperator,
		Description: "pvnp-holdout stated label (verbatim canonicalization): polynomial-method extension.",
		Aliases:     []string{"polynomial-method extension"},
	},
	{
		CanonicalID: "domain.complexity_theory.operator.property_search",
		FieldKind:   domain.FieldOperator,
		Description: "pvnp-holdout stated label (verbatim canonicalization): property search.",
		Aliases:     []string{"property search"},
	},
	{
		CanonicalID: "domain.complexity_theory.operator.property_to_distinguisher_reduction",
		FieldKind:   domain.FieldOperator,
		Description: "pvnp-holdout stated label (verbatim canonicalization): property-to-distinguisher reduction.",
		Aliases:     []string{"property-to-distinguisher reduction"},
	},
	{
		CanonicalID: "domain.complexity_theory.operator.random_restriction",
		FieldKind:   domain.FieldOperator,
		Description: "pvnp-holdout stated label (verbatim canonicalization): random restriction.",
		Aliases:     []string{"random restriction"},
	},
	{
		CanonicalID: "domain.complexity_theory.operator.restriction_lifting",
		FieldKind:   domain.FieldOperator,
		Description: "pvnp-holdout stated label (verbatim canonicalization): restriction lifting.",
		Aliases:     []string{"restriction lifting"},
	},
	{
		CanonicalID: "domain.complexity_theory.operator.restriction_on_proofs",
		FieldKind:   domain.FieldOperator,
		Description: "pvnp-holdout stated label (verbatim canonicalization): restriction on proofs.",
		Aliases:     []string{"restriction on proofs"},
	},
	{
		CanonicalID: "domain.complexity_theory.operator.stage_wise_diagonalization_against_query_behavior",
		FieldKind:   domain.FieldOperator,
		Description: "pvnp-holdout stated label (verbatim canonicalization): stage-wise diagonalization against query behavior.",
		Aliases:     []string{"stage-wise diagonalization against query behavior"},
	},
	{
		CanonicalID: "domain.complexity_theory.operator.succinct_witness_compression",
		FieldKind:   domain.FieldOperator,
		Description: "pvnp-holdout stated label (verbatim canonicalization): succinct witness compression.",
		Aliases:     []string{"succinct witness compression"},
	},
	{
		CanonicalID: "domain.complexity_theory.operator.sunflower_lemma",
		FieldKind:   domain.FieldOperator,
		Description: "pvnp-holdout stated label (verbatim canonicalization): sunflower lemma.",
		Aliases:     []string{"sunflower lemma"},
	},
	{
		CanonicalID: "domain.complexity_theory.operator.switching_lemma",
		FieldKind:   domain.FieldOperator,
		Description: "pvnp-holdout stated label (verbatim canonicalization): switching lemma.",
		Aliases:     []string{"switching lemma"},
	},
	{
		CanonicalID: "domain.complexity_theory.operator.symmetry_classification",
		FieldKind:   domain.FieldOperator,
		Description: "pvnp-holdout stated label (verbatim canonicalization): symmetry classification.",
		Aliases:     []string{"symmetry classification"},
	},
	{
		CanonicalID: "domain.complexity_theory.operator.width_lower_bounds",
		FieldKind:   domain.FieldOperator,
		Description: "pvnp-holdout stated label (verbatim canonicalization): width lower bounds.",
		Aliases:     []string{"width lower bounds"},
	},
	{
		CanonicalID: "domain.complexity_theory.assumption.a_property_based_route_exists_that_evades_largeness_or_constructivity",
		FieldKind:   domain.FieldAssumption,
		Description: "pvnp-holdout stated label (verbatim canonicalization): a property-based route exists that evades largeness or constructivity.",
		Aliases:     []string{"a property-based route exists that evades largeness or constructivity"},
	},
	{
		CanonicalID: "domain.complexity_theory.assumption.a_supplied_sat_algorithm_with_superpolynomial_savings_over_brute_force",
		FieldKind:   domain.FieldAssumption,
		Description: "pvnp-holdout stated label (verbatim canonicalization): a supplied SAT algorithm with superpolynomial savings over brute force.",
		Aliases:     []string{"a supplied sat algorithm with superpolynomial savings over brute force"},
	},
	{
		CanonicalID: "domain.complexity_theory.assumption.algebraic_proxy_transfers_to_boolean_separation",
		FieldKind:   domain.FieldAssumption,
		Description: "pvnp-holdout stated label (verbatim canonicalization): algebraic proxy transfers to Boolean separation.",
		Aliases:     []string{"algebraic proxy transfers to boolean separation"},
	},
	{
		CanonicalID: "domain.complexity_theory.assumption.black_box_simulation_suffices_for_separation",
		FieldKind:   domain.FieldAssumption,
		Description: "pvnp-holdout stated label (verbatim canonicalization): black-box simulation suffices for separation.",
		Aliases:     []string{"black-box simulation suffices for separation"},
	},
	{
		CanonicalID: "domain.complexity_theory.assumption.deterministic_subexponential_pit_exists_hypothesis_not_theorem",
		FieldKind:   domain.FieldAssumption,
		Description: "pvnp-holdout stated label (verbatim canonicalization): deterministic subexponential PIT exists (hypothesis, not theorem).",
		Aliases:     []string{"deterministic subexponential pit exists (hypothesis, not theorem)"},
	},
	{
		CanonicalID: "domain.complexity_theory.assumption.gate_wise_approximation_errors_stay_controllable",
		FieldKind:   domain.FieldAssumption,
		Description: "pvnp-holdout stated label (verbatim canonicalization): gate-wise approximation errors stay controllable.",
		Aliases:     []string{"gate-wise approximation errors stay controllable"},
	},
	{
		CanonicalID: "domain.complexity_theory.assumption.gate_wise_polynomial_approximation_composes_along_constant_depth",
		FieldKind:   domain.FieldAssumption,
		Description: "pvnp-holdout stated label (verbatim canonicalization): gate-wise polynomial approximation composes along constant depth.",
		Aliases:     []string{"gate-wise polynomial approximation composes along constant depth"},
	},
	{
		CanonicalID: "domain.complexity_theory.assumption.hard_function_is_far_from_all_low_degree_polynomials",
		FieldKind:   domain.FieldAssumption,
		Description: "pvnp-holdout stated label (verbatim canonicalization): hard function is far from all low-degree polynomials.",
		Aliases:     []string{"hard function is far from all low-degree polynomials"},
	},
	{
		CanonicalID: "domain.complexity_theory.assumption.hard_function_survives_restriction",
		FieldKind:   domain.FieldAssumption,
		Description: "pvnp-holdout stated label (verbatim canonicalization): hard function survives restriction.",
		Aliases:     []string{"hard function survives restriction"},
	},
	{
		CanonicalID: "domain.complexity_theory.assumption.hierarchy_theorem_engine_transfers_to_p_vs_np",
		FieldKind:   domain.FieldAssumption,
		Description: "pvnp-holdout stated label (verbatim canonicalization): hierarchy-theorem engine transfers to P vs NP.",
		Aliases:     []string{"hierarchy-theorem engine transfers to p vs np"},
	},
	{
		CanonicalID: "domain.complexity_theory.assumption.hypothetical_fast_low_space_sat_algorithm_for_contradiction",
		FieldKind:   domain.FieldAssumption,
		Description: "pvnp-holdout stated label (verbatim canonicalization): hypothetical fast low-space SAT algorithm (for contradiction).",
		Aliases:     []string{"hypothetical fast low-space sat algorithm (for contradiction)"},
	},
	{
		CanonicalID: "domain.complexity_theory.assumption.monotone_complexity_approximates_general_complexity_on_monotone_functions",
		FieldKind:   domain.FieldAssumption,
		Description: "pvnp-holdout stated label (verbatim canonicalization): monotone complexity approximates general complexity on monotone functions.",
		Aliases:     []string{"monotone complexity approximates general complexity on monotone functions"},
	},
	{
		CanonicalID: "domain.complexity_theory.assumption.obstruction_multiplicities_are_computable_or_boundable",
		FieldKind:   domain.FieldAssumption,
		Description: "pvnp-holdout stated label (verbatim canonicalization): obstruction multiplicities are computable or boundable.",
		Aliases:     []string{"obstruction multiplicities are computable or boundable"},
	},
	{
		CanonicalID: "domain.complexity_theory.assumption.per_system_lower_bounds_accumulate_toward_all_systems_hardness",
		FieldKind:   domain.FieldAssumption,
		Description: "pvnp-holdout stated label (verbatim canonicalization): per-system lower bounds accumulate toward all-systems hardness.",
		Aliases:     []string{"per-system lower bounds accumulate toward all-systems hardness"},
	},
	{
		CanonicalID: "domain.complexity_theory.assumption.strong_prfs_exist_in_the_target_class",
		FieldKind:   domain.FieldAssumption,
		Description: "pvnp-holdout stated label (verbatim canonicalization): strong PRFs exist in the target class.",
		Aliases:     []string{"strong prfs exist in the target class"},
	},
	{
		CanonicalID: "domain.complexity_theory.assumption.structural_normal_form_for_acc0_holds_at_quasipolynomial_size",
		FieldKind:   domain.FieldAssumption,
		Description: "pvnp-holdout stated label (verbatim canonicalization): structural normal form for ACC0 holds at quasipolynomial size.",
		Aliases:     []string{"structural normal form for acc0 holds at quasipolynomial size"},
	},
	{
		CanonicalID: "domain.complexity_theory.assumption.succinct_completeness_for_nexp_verification",
		FieldKind:   domain.FieldAssumption,
		Description: "pvnp-holdout stated label (verbatim canonicalization): succinct completeness for NEXP verification.",
		Aliases:     []string{"succinct completeness for nexp verification"},
	},
	{
		CanonicalID: "domain.complexity_theory.assumption.technique_validity_is_preserved_under_a_shared_oracle",
		FieldKind:   domain.FieldAssumption,
		Description: "pvnp-holdout stated label (verbatim canonicalization): technique validity is preserved under a shared oracle.",
		Aliases:     []string{"technique validity is preserved under a shared oracle"},
	},
	{
		CanonicalID: "domain.complexity_theory.assumption.test_distributions_separate_hard_function_from_approximators",
		FieldKind:   domain.FieldAssumption,
		Description: "pvnp-holdout stated label (verbatim canonicalization): test distributions separate hard function from approximators.",
		Aliases:     []string{"test distributions separate hard function from approximators"},
	},
	{
		CanonicalID: "domain.complexity_theory.assumption.weak_class_internal_structure_is_analyzable",
		FieldKind:   domain.FieldAssumption,
		Description: "pvnp-holdout stated label (verbatim canonicalization): weak-class internal structure is analyzable.",
		Aliases:     []string{"weak-class internal structure is analyzable"},
	},
	{
		CanonicalID: "domain.complexity_theory.assumption.width_bottleneck_properties_certify_proof_size",
		FieldKind:   domain.FieldAssumption,
		Description: "pvnp-holdout stated label (verbatim canonicalization): width/bottleneck properties certify proof size.",
		Aliases:     []string{"width/bottleneck properties certify proof size"},
	},
	{
		CanonicalID: "domain.complexity_theory.property.black_box_relativizing_simulation",
		FieldKind:   domain.FieldPreserves,
		Description: "pvnp-holdout stated label (verbatim canonicalization): black-box relativizing simulation.",
		Aliases:     []string{"black-box relativizing simulation"},
	},
	{
		CanonicalID: "domain.complexity_theory.property.large_constructive_distinguishing_property",
		FieldKind:   domain.FieldPreserves,
		Description: "pvnp-holdout stated label (verbatim canonicalization): large constructive distinguishing property.",
		Aliases:     []string{"large constructive distinguishing property"},
	},
	{
		CanonicalID: "domain.complexity_theory.property.model_analysis_first_direction",
		FieldKind:   domain.FieldPreserves,
		Description: "pvnp-holdout stated label (verbatim canonicalization): model-analysis-first direction.",
		Aliases:     []string{"model-analysis-first direction"},
	},
	{
		CanonicalID: "domain.complexity_theory.broken.black_box_relativizing_simulation",
		FieldKind:   domain.FieldBreaks,
		Description: "pvnp-holdout stated label (verbatim canonicalization): black-box relativizing simulation.",
		Aliases:     []string{"black-box relativizing simulation"},
	},
	{
		CanonicalID: "domain.complexity_theory.broken.large_constructive_distinguishing_property",
		FieldKind:   domain.FieldBreaks,
		Description: "pvnp-holdout stated label (verbatim canonicalization): large constructive distinguishing property.",
		Aliases:     []string{"large constructive distinguishing property"},
	},
	{
		CanonicalID: "domain.complexity_theory.broken.model_analysis_first_direction",
		FieldKind:   domain.FieldBreaks,
		Description: "pvnp-holdout stated label (verbatim canonicalization): model-analysis-first direction.",
		Aliases:     []string{"model-analysis-first direction"},
	},
	{
		CanonicalID: "domain.complexity_theory.auxiliary_object.kronecker_plethysm_coefficients",
		FieldKind:   domain.FieldAuxiliaryObject,
		Description: "pvnp-holdout stated label (verbatim canonicalization): Kronecker/plethysm coefficients.",
		Aliases:     []string{"kronecker/plethysm coefficients"},
	},
	{
		CanonicalID: "domain.complexity_theory.auxiliary_object.tardos_function",
		FieldKind:   domain.FieldAuxiliaryObject,
		Description: "pvnp-holdout stated label (verbatim canonicalization): Tardos function.",
		Aliases:     []string{"tardos function"},
	},
	{
		CanonicalID: "domain.complexity_theory.auxiliary_object.alternation_hierarchy",
		FieldKind:   domain.FieldAuxiliaryObject,
		Description: "pvnp-holdout stated label (verbatim canonicalization): alternation hierarchy.",
		Aliases:     []string{"alternation hierarchy"},
	},
	{
		CanonicalID: "domain.complexity_theory.auxiliary_object.approximating_polynomial_ensemble",
		FieldKind:   domain.FieldAuxiliaryObject,
		Description: "pvnp-holdout stated label (verbatim canonicalization): approximating polynomial ensemble.",
		Aliases:     []string{"approximating polynomial ensemble"},
	},
	{
		CanonicalID: "domain.complexity_theory.auxiliary_object.approximator_lattice",
		FieldKind:   domain.FieldAuxiliaryObject,
		Description: "pvnp-holdout stated label (verbatim canonicalization): approximator lattice.",
		Aliases:     []string{"approximator lattice"},
	},
	{
		CanonicalID: "domain.complexity_theory.auxiliary_object.bounded_arithmetic_fragments",
		FieldKind:   domain.FieldAuxiliaryObject,
		Description: "pvnp-holdout stated label (verbatim canonicalization): bounded-arithmetic fragments.",
		Aliases:     []string{"bounded-arithmetic fragments"},
	},
	{
		CanonicalID: "domain.complexity_theory.auxiliary_object.clocked_simulations",
		FieldKind:   domain.FieldAuxiliaryObject,
		Description: "pvnp-holdout stated label (verbatim canonicalization): clocked simulations.",
		Aliases:     []string{"clocked simulations"},
	},
	{
		CanonicalID: "domain.complexity_theory.auxiliary_object.collapsed_decision_tree",
		FieldKind:   domain.FieldAuxiliaryObject,
		Description: "pvnp-holdout stated label (verbatim canonicalization): collapsed decision tree.",
		Aliases:     []string{"collapsed decision tree"},
	},
	{
		CanonicalID: "domain.complexity_theory.auxiliary_object.enumeration_of_clocked_machines",
		FieldKind:   domain.FieldAuxiliaryObject,
		Description: "pvnp-holdout stated label (verbatim canonicalization): enumeration of clocked machines.",
		Aliases:     []string{"enumeration of clocked machines"},
	},
	{
		CanonicalID: "domain.complexity_theory.auxiliary_object.field_f_p",
		FieldKind:   domain.FieldAuxiliaryObject,
		Description: "pvnp-holdout stated label (verbatim canonicalization): field F_p.",
		Aliases:     []string{"field f_p"},
	},
	{
		CanonicalID: "domain.complexity_theory.auxiliary_object.mild_savings_circuit_analysis_algorithm",
		FieldKind:   domain.FieldAuxiliaryObject,
		Description: "pvnp-holdout stated label (verbatim canonicalization): mild-savings circuit-analysis algorithm.",
		Aliases:     []string{"mild-savings circuit-analysis algorithm"},
	},
	{
		CanonicalID: "domain.complexity_theory.auxiliary_object.oracle_a_with_p_a_np_a",
		FieldKind:   domain.FieldAuxiliaryObject,
		Description: "pvnp-holdout stated label (verbatim canonicalization): oracle A with P^A = NP^A.",
		Aliases:     []string{"oracle a with p^a = np^a"},
	},
	{
		CanonicalID: "domain.complexity_theory.auxiliary_object.oracle_b_with_p_b_np_b",
		FieldKind:   domain.FieldAuxiliaryObject,
		Description: "pvnp-holdout stated label (verbatim canonicalization): oracle B with P^B != NP^B.",
		Aliases:     []string{"oracle b with p^b != np^b"},
	},
	{
		CanonicalID: "domain.complexity_theory.auxiliary_object.perfect_matching_function",
		FieldKind:   domain.FieldAuxiliaryObject,
		Description: "pvnp-holdout stated label (verbatim canonicalization): perfect matching function.",
		Aliases:     []string{"perfect matching function"},
	},
	{
		CanonicalID: "domain.complexity_theory.auxiliary_object.permanent_and_determinant_orbit_closures",
		FieldKind:   domain.FieldAuxiliaryObject,
		Description: "pvnp-holdout stated label (verbatim canonicalization): permanent and determinant orbit closures.",
		Aliases:     []string{"permanent and determinant orbit closures"},
	},
	{
		CanonicalID: "domain.complexity_theory.auxiliary_object.permanent_as_hard_candidate",
		FieldKind:   domain.FieldAuxiliaryObject,
		Description: "pvnp-holdout stated label (verbatim canonicalization): permanent as hard candidate.",
		Aliases:     []string{"permanent as hard candidate"},
	},
	{
		CanonicalID: "domain.complexity_theory.auxiliary_object.pigeonhole_tautologies",
		FieldKind:   domain.FieldAuxiliaryObject,
		Description: "pvnp-holdout stated label (verbatim canonicalization): pigeonhole tautologies.",
		Aliases:     []string{"pigeonhole tautologies"},
	},
	{
		CanonicalID: "domain.complexity_theory.auxiliary_object.polynomial_identity_tests",
		FieldKind:   domain.FieldAuxiliaryObject,
		Description: "pvnp-holdout stated label (verbatim canonicalization): polynomial identity tests.",
		Aliases:     []string{"polynomial identity tests"},
	},
	{
		CanonicalID: "domain.complexity_theory.auxiliary_object.positive_negative_test_distributions",
		FieldKind:   domain.FieldAuxiliaryObject,
		Description: "pvnp-holdout stated label (verbatim canonicalization): positive/negative test distributions.",
		Aliases:     []string{"positive/negative test distributions"},
	},
	{
		CanonicalID: "domain.complexity_theory.auxiliary_object.proof_width_measure",
		FieldKind:   domain.FieldAuxiliaryObject,
		Description: "pvnp-holdout stated label (verbatim canonicalization): proof-width measure.",
		Aliases:     []string{"proof-width measure"},
	},
	{
		CanonicalID: "domain.complexity_theory.auxiliary_object.pseudorandom_function_family",
		FieldKind:   domain.FieldAuxiliaryObject,
		Description: "pvnp-holdout stated label (verbatim canonicalization): pseudorandom function family.",
		Aliases:     []string{"pseudorandom function family"},
	},
	{
		CanonicalID: "domain.complexity_theory.auxiliary_object.restriction_distribution",
		FieldKind:   domain.FieldAuxiliaryObject,
		Description: "pvnp-holdout stated label (verbatim canonicalization): restriction distribution.",
		Aliases:     []string{"restriction distribution"},
	},
	{
		CanonicalID: "domain.complexity_theory.auxiliary_object.succinct_sat_instances",
		FieldKind:   domain.FieldAuxiliaryObject,
		Description: "pvnp-holdout stated label (verbatim canonicalization): succinct SAT instances.",
		Aliases:     []string{"succinct sat instances"},
	},
	{
		CanonicalID: "domain.complexity_theory.auxiliary_object.supplied_acc0_evaluation_algorithm",
		FieldKind:   domain.FieldAuxiliaryObject,
		Description: "pvnp-holdout stated label (verbatim canonicalization): supplied ACC0 evaluation algorithm.",
		Aliases:     []string{"supplied acc0 evaluation algorithm"},
	},
	{
		CanonicalID: "domain.complexity_theory.auxiliary_object.symmetric_function_top_gate",
		FieldKind:   domain.FieldAuxiliaryObject,
		Description: "pvnp-holdout stated label (verbatim canonicalization): symmetric-function top gate.",
		Aliases:     []string{"symmetric-function top gate"},
	},
	{
		CanonicalID: "domain.complexity_theory.auxiliary_object.truth_table_property_tester",
		FieldKind:   domain.FieldAuxiliaryObject,
		Description: "pvnp-holdout stated label (verbatim canonicalization): truth-table property tester.",
		Aliases:     []string{"truth-table property tester"},
	},
	{
		CanonicalID: "domain.complexity_theory.auxiliary_object.universal_machine",
		FieldKind:   domain.FieldAuxiliaryObject,
		Description: "pvnp-holdout stated label (verbatim canonicalization): universal machine.",
		Aliases:     []string{"universal machine"},
	},
	{
		CanonicalID: "domain.complexity_theory.representation.ac0_p_circuits",
		FieldKind:   domain.FieldRepresentation,
		Description: "pvnp-holdout stated label (verbatim canonicalization): AC0[p] circuits.",
		Aliases:     []string{"ac0[p] circuits"},
	},
	{
		CanonicalID: "domain.complexity_theory.representation.acc0_circuits",
		FieldKind:   domain.FieldRepresentation,
		Description: "pvnp-holdout stated label (verbatim canonicalization): ACC0 circuits.",
		Aliases:     []string{"acc0 circuits"},
	},
	{
		CanonicalID: "domain.complexity_theory.representation.turing_machine_enumerations",
		FieldKind:   domain.FieldRepresentation,
		Description: "pvnp-holdout stated label (verbatim canonicalization): Turing machine enumerations.",
		Aliases:     []string{"turing machine enumerations"},
	},
	{
		CanonicalID: "domain.complexity_theory.representation.all_inputs_evaluation_tables",
		FieldKind:   domain.FieldRepresentation,
		Description: "pvnp-holdout stated label (verbatim canonicalization): all-inputs evaluation tables.",
		Aliases:     []string{"all-inputs evaluation tables"},
	},
	{
		CanonicalID: "domain.complexity_theory.representation.alternating_time_classes",
		FieldKind:   domain.FieldRepresentation,
		Description: "pvnp-holdout stated label (verbatim canonicalization): alternating time classes.",
		Aliases:     []string{"alternating time classes"},
	},
	{
		CanonicalID: "domain.complexity_theory.representation.arithmetic_circuits",
		FieldKind:   domain.FieldRepresentation,
		Description: "pvnp-holdout stated label (verbatim canonicalization): arithmetic circuits.",
		Aliases:     []string{"arithmetic circuits"},
	},
	{
		CanonicalID: "domain.complexity_theory.representation.bounded_arithmetic_systems",
		FieldKind:   domain.FieldRepresentation,
		Description: "pvnp-holdout stated label (verbatim canonicalization): bounded-arithmetic systems.",
		Aliases:     []string{"bounded-arithmetic systems"},
	},
	{
		CanonicalID: "domain.complexity_theory.representation.circuit_sat_instances",
		FieldKind:   domain.FieldRepresentation,
		Description: "pvnp-holdout stated label (verbatim canonicalization): circuit-SAT instances.",
		Aliases:     []string{"circuit-sat instances"},
	},
	{
		CanonicalID: "domain.complexity_theory.representation.completeness_translations",
		FieldKind:   domain.FieldRepresentation,
		Description: "pvnp-holdout stated label (verbatim canonicalization): completeness translations.",
		Aliases:     []string{"completeness translations"},
	},
	{
		CanonicalID: "domain.complexity_theory.representation.conditional_implications",
		FieldKind:   domain.FieldRepresentation,
		Description: "pvnp-holdout stated label (verbatim canonicalization): conditional implications.",
		Aliases:     []string{"conditional implications"},
	},
	{
		CanonicalID: "domain.complexity_theory.representation.constant_depth_circuits",
		FieldKind:   domain.FieldRepresentation,
		Description: "pvnp-holdout stated label (verbatim canonicalization): constant-depth circuits.",
		Aliases:     []string{"constant-depth circuits"},
	},
	{
		CanonicalID: "domain.complexity_theory.representation.contrary_oracle_constructions",
		FieldKind:   domain.FieldRepresentation,
		Description: "pvnp-holdout stated label (verbatim canonicalization): contrary oracle constructions.",
		Aliases:     []string{"contrary oracle constructions"},
	},
	{
		CanonicalID: "domain.complexity_theory.representation.decision_trees",
		FieldKind:   domain.FieldRepresentation,
		Description: "pvnp-holdout stated label (verbatim canonicalization): decision trees.",
		Aliases:     []string{"decision trees"},
	},
	{
		CanonicalID: "domain.complexity_theory.representation.depth_two_symmetric_top_gate_normal_form",
		FieldKind:   domain.FieldRepresentation,
		Description: "pvnp-holdout stated label (verbatim canonicalization): depth-two symmetric-top-gate normal form.",
		Aliases:     []string{"depth-two symmetric-top-gate normal form"},
	},
	{
		CanonicalID: "domain.complexity_theory.representation.general_circuits",
		FieldKind:   domain.FieldRepresentation,
		Description: "pvnp-holdout stated label (verbatim canonicalization): general circuits.",
		Aliases:     []string{"general circuits"},
	},
	{
		CanonicalID: "domain.complexity_theory.representation.identity_testing_algorithms",
		FieldKind:   domain.FieldRepresentation,
		Description: "pvnp-holdout stated label (verbatim canonicalization): identity-testing algorithms.",
		Aliases:     []string{"identity-testing algorithms"},
	},
	{
		CanonicalID: "domain.complexity_theory.representation.irreducible_representations",
		FieldKind:   domain.FieldRepresentation,
		Description: "pvnp-holdout stated label (verbatim canonicalization): irreducible representations.",
		Aliases:     []string{"irreducible representations"},
	},
	{
		CanonicalID: "domain.complexity_theory.representation.lattice_of_approximator_functions",
		FieldKind:   domain.FieldRepresentation,
		Description: "pvnp-holdout stated label (verbatim canonicalization): lattice of approximator functions.",
		Aliases:     []string{"lattice of approximator functions"},
	},
	{
		CanonicalID: "domain.complexity_theory.representation.low_degree_polynomials_over_f_p",
		FieldKind:   domain.FieldRepresentation,
		Description: "pvnp-holdout stated label (verbatim canonicalization): low-degree polynomials over F_p.",
		Aliases:     []string{"low-degree polynomials over f_p"},
	},
	{
		CanonicalID: "domain.complexity_theory.representation.monotone_circuits",
		FieldKind:   domain.FieldRepresentation,
		Description: "pvnp-holdout stated label (verbatim canonicalization): monotone circuits.",
		Aliases:     []string{"monotone circuits"},
	},
	{
		CanonicalID: "domain.complexity_theory.representation.multiplicity_obstructions",
		FieldKind:   domain.FieldRepresentation,
		Description: "pvnp-holdout stated label (verbatim canonicalization): multiplicity obstructions.",
		Aliases:     []string{"multiplicity obstructions"},
	},
	{
		CanonicalID: "domain.complexity_theory.representation.natural_properties_over_truth_tables",
		FieldKind:   domain.FieldRepresentation,
		Description: "pvnp-holdout stated label (verbatim canonicalization): natural properties over truth tables.",
		Aliases:     []string{"natural properties over truth tables"},
	},
	{
		CanonicalID: "domain.complexity_theory.representation.nondeterministic_time_hierarchy",
		FieldKind:   domain.FieldRepresentation,
		Description: "pvnp-holdout stated label (verbatim canonicalization): nondeterministic time hierarchy.",
		Aliases:     []string{"nondeterministic time hierarchy"},
	},
	{
		CanonicalID: "domain.complexity_theory.representation.oracle_machines",
		FieldKind:   domain.FieldRepresentation,
		Description: "pvnp-holdout stated label (verbatim canonicalization): oracle machines.",
		Aliases:     []string{"oracle machines"},
	},
	{
		CanonicalID: "domain.complexity_theory.representation.orbit_closures_under_gl_action",
		FieldKind:   domain.FieldRepresentation,
		Description: "pvnp-holdout stated label (verbatim canonicalization): orbit closures under GL-action.",
		Aliases:     []string{"orbit closures under gl-action"},
	},
	{
		CanonicalID: "domain.complexity_theory.representation.propositional_proof_systems",
		FieldKind:   domain.FieldRepresentation,
		Description: "pvnp-holdout stated label (verbatim canonicalization): propositional proof systems.",
		Aliases:     []string{"propositional proof systems"},
	},
	{
		CanonicalID: "domain.complexity_theory.representation.pseudorandom_function_families",
		FieldKind:   domain.FieldRepresentation,
		Description: "pvnp-holdout stated label (verbatim canonicalization): pseudorandom function families.",
		Aliases:     []string{"pseudorandom function families"},
	},
	{
		CanonicalID: "domain.complexity_theory.representation.random_restrictions",
		FieldKind:   domain.FieldRepresentation,
		Description: "pvnp-holdout stated label (verbatim canonicalization): random restrictions.",
		Aliases:     []string{"random restrictions"},
	},
	{
		CanonicalID: "domain.complexity_theory.representation.resolution_refutations",
		FieldKind:   domain.FieldRepresentation,
		Description: "pvnp-holdout stated label (verbatim canonicalization): resolution refutations.",
		Aliases:     []string{"resolution refutations"},
	},
	{
		CanonicalID: "domain.complexity_theory.representation.resource_bounded_simulations",
		FieldKind:   domain.FieldRepresentation,
		Description: "pvnp-holdout stated label (verbatim canonicalization): resource-bounded simulations.",
		Aliases:     []string{"resource-bounded simulations"},
	},
	{
		CanonicalID: "domain.complexity_theory.representation.restricted_circuit_models",
		FieldKind:   domain.FieldRepresentation,
		Description: "pvnp-holdout stated label (verbatim canonicalization): restricted circuit models.",
		Aliases:     []string{"restricted circuit models"},
	},
	{
		CanonicalID: "domain.complexity_theory.representation.simultaneous_time_space_machines",
		FieldKind:   domain.FieldRepresentation,
		Description: "pvnp-holdout stated label (verbatim canonicalization): simultaneous time-space machines.",
		Aliases:     []string{"simultaneous time-space machines"},
	},
	{
		CanonicalID: "domain.complexity_theory.representation.slice_transfer_arguments",
		FieldKind:   domain.FieldRepresentation,
		Description: "pvnp-holdout stated label (verbatim canonicalization): slice/transfer arguments.",
		Aliases:     []string{"slice/transfer arguments"},
	},
	{
		CanonicalID: "domain.complexity_theory.representation.succinct_witness_encodings",
		FieldKind:   domain.FieldRepresentation,
		Description: "pvnp-holdout stated label (verbatim canonicalization): succinct witness encodings.",
		Aliases:     []string{"succinct witness encodings"},
	},
	{
		CanonicalID: "domain.complexity_theory.representation.sunflower_systems",
		FieldKind:   domain.FieldRepresentation,
		Description: "pvnp-holdout stated label (verbatim canonicalization): sunflower systems.",
		Aliases:     []string{"sunflower systems"},
	},
	{
		CanonicalID: "domain.complexity_theory.representation.tautology_families",
		FieldKind:   domain.FieldRepresentation,
		Description: "pvnp-holdout stated label (verbatim canonicalization): tautology families.",
		Aliases:     []string{"tautology families"},
	},
	{
		CanonicalID: "domain.complexity_theory.representation.truth_table_properties",
		FieldKind:   domain.FieldRepresentation,
		Description: "pvnp-holdout stated label (verbatim canonicalization): truth-table properties.",
		Aliases:     []string{"truth-table properties"},
	},
}

// MechanismV6 builds the mechanism/v6 vocabulary: every mechanism/v5 term
// unchanged plus the pvnp-holdout verbatim corpus canonicalizations.
func MechanismV6() *Vocabulary {
	base := append(append([]Term{}, mechanismV1Seed.terms...), mechanismV2Terms...)
	base = append(base, mechanismV3Terms...)
	base = append(base, mechanismV4Terms...)
	base = append(base, mechanismV5Terms...)
	terms := make([]Term, 0, len(base)+len(mechanismV6Terms))
	for _, t := range base {
		if extra, ok := mechanismV3ExtraAliases[t.CanonicalID]; ok {
			t.Aliases = append(append([]string{}, t.Aliases...), extra...)
		}
		terms = append(terms, t)
	}
	terms = append(terms, mechanismV6Terms...)
	seed := vocabularyBuilder{
		version:  VocabularyMechanismV6,
		terms:    terms,
		rejected: append([]string{}, mechanismV1Seed.rejected...),
	}
	return mustBuild(seed)
}

// SeededVocabularies returns every in-repo vocabulary, used to seed persistence.
func SeededVocabularies() []*Vocabulary {
	return []*Vocabulary{MechanismV1(), MechanismV2(), MechanismV3(), MechanismV4(), MechanismV5(), MechanismV6(), MechanismV0Lossy()}
}

func mustBuild(b vocabularyBuilder) *Vocabulary {
	v, err := b.build()
	if err != nil {
		panic("canon: invalid in-repo vocabulary seed " + b.version + ": " + err.Error())
	}
	return v
}
