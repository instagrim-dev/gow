package canon

import (
	"fmt"
	"sort"

	"github.com/instagrim-dev/newf/internal/domain"
)

// WeightsMechanismV1 is the default, versioned comparison weighting. Weights are
// explicit and versioned so no behavior hides inside an opaque distance. In v1
// the weights only tag which fields are decisive for the mechanistic
// classification; per-field results are always reported unweighted.
const WeightsMechanismV1 = "weights/v1"

// ClassifyMechanismV1 is the versioned mechanistic-vs-surface classification
// rule. It names which fields are decisive.
const ClassifyMechanismV1 = "classify/v1"

// Ordinal is the per-field similarity ordinal. No single scalar is emitted.
type Ordinal string

const (
	OrdinalIdentical    Ordinal = "identical"
	OrdinalHigh         Ordinal = "high"
	OrdinalLow          Ordinal = "low"
	OrdinalNone         Ordinal = "none"
	OrdinalIncomparable Ordinal = "incomparable"
)

// Classification is the categorical mechanistic-vs-surface verdict.
type Classification string

const (
	ClassMechanismNear           Classification = "mechanism-near"
	ClassMechanismDistinct       Classification = "mechanism-distinct"
	ClassSurfaceDistinctMechNear Classification = "surface-distinct+mechanism-near"
	ClassSurfaceNearMechDistinct Classification = "surface-near+mechanism-distinct"
	ClassUnknown                 Classification = "unknown"
)

// FieldResult is the per-field comparison outcome over resolved canonical IDs.
type FieldResult struct {
	FieldKind    domain.FieldKind
	OverlapCount int
	UnionCount   int
	Jaccard      float64
	Ordinal      Ordinal
	// Incomparable is true when either side had a non-resolved claim in the
	// field, so agreement cannot be asserted.
	Incomparable bool
}

// PostureResult reports enum equality per posture axis.
type PostureResult struct {
	LocalityEqual     bool
	ConstructionEqual bool
	UncertaintyEqual  bool
}

// Comparison is the full, per-field, non-scalar comparison result.
type Comparison struct {
	WeightsVersion  string
	ClassifyVersion string
	Fields          []FieldResult
	Posture         PostureResult
	OutcomeEqual    bool
	Classification  Classification
	// SurfaceSimilarity is auxiliary-only diagnostic data; it never drives the
	// mechanistic classification. Nil when no surface text was supplied.
	SurfaceSimilarity *SurfaceSimilarity
}

// SurfaceSimilarity is an auxiliary, non-identity signal derived from surface
// text. It is always tagged auxiliary-only.
type SurfaceSimilarity struct {
	Jaccard       float64
	AuxiliaryOnly bool
}

// decisiveFields are the fields that determine the mechanistic classification in
// classify/v1. Representations, boundaries, and posture are non-decisive and
// feed only the surface axis. The rationale: operators/assumptions/preserves are
// the structural moves and conserved properties that predict outcome; a
// different representation of the same move is a surface change.
var decisiveFields = []domain.FieldKind{
	domain.FieldPreserves,
	domain.FieldOperator,
	domain.FieldAssumption,
}

// Compare produces a deterministic, component-wise comparison of two signatures.
// Set fields are compared over resolved canonical IDs; a field with any
// non-resolved claim on either side is marked incomparable and cannot count as
// agreement. weightsVersion selects the (currently single) weighting; an unknown
// version is an error so callers cannot silently compare under an absent config.
func Compare(a, b MechanismSignature, weightsVersion string) (Comparison, error) {
	if weightsVersion == "" {
		weightsVersion = WeightsMechanismV1
	}
	if weightsVersion != WeightsMechanismV1 {
		return Comparison{}, fmt.Errorf("unknown weights version %q", weightsVersion)
	}

	cmp := Comparison{
		WeightsVersion:  weightsVersion,
		ClassifyVersion: ClassifyMechanismV1,
		OutcomeEqual:    a.OutcomeClass == b.OutcomeClass,
		Posture: PostureResult{
			LocalityEqual:     a.Posture.Locality == b.Posture.Locality,
			ConstructionEqual: a.Posture.Construction == b.Posture.Construction,
			UncertaintyEqual:  a.Posture.Uncertainty == b.Posture.Uncertainty,
		},
	}

	setFields := []struct {
		kind domain.FieldKind
		a, b []FieldClaim
	}{
		{domain.FieldRepresentation, a.Representations, b.Representations},
		{domain.FieldOperator, a.Operators, b.Operators},
		{domain.FieldAssumption, a.Assumptions, b.Assumptions},
		{domain.FieldPreserves, a.Preserves, b.Preserves},
		{domain.FieldBreaks, a.Breaks, b.Breaks},
		{domain.FieldAuxiliaryObject, a.AuxiliaryObjects, b.AuxiliaryObjects},
	}

	for _, sf := range setFields {
		cmp.Fields = append(cmp.Fields, compareSetField(sf.kind, sf.a, sf.b))
	}

	cmp.Classification = classifyV1(cmp)
	return cmp, nil
}

func compareSetField(kind domain.FieldKind, a, b []FieldClaim) FieldResult {
	res := FieldResult{FieldKind: kind}
	if hasUnresolved(a) || hasUnresolved(b) {
		res.Incomparable = true
		res.Ordinal = OrdinalIncomparable
		return res
	}
	setA := resolvedIDs(a)
	setB := resolvedIDs(b)
	overlap, union := intersectUnion(setA, setB)
	res.OverlapCount = overlap
	res.UnionCount = union
	if union == 0 {
		// Both empty: treat as identical (nothing to disagree on).
		res.Jaccard = 1
		res.Ordinal = OrdinalIdentical
		return res
	}
	res.Jaccard = float64(overlap) / float64(union)
	res.Ordinal = jaccardOrdinal(res.Jaccard)
	return res
}

func intersectUnion(a, b []domain.CanonicalID) (overlap, union int) {
	set := map[domain.CanonicalID]int{}
	for _, id := range a {
		set[id] |= 1
	}
	for _, id := range b {
		set[id] |= 2
	}
	for _, bits := range set {
		union++
		if bits == 3 {
			overlap++
		}
	}
	return overlap, union
}

func jaccardOrdinal(j float64) Ordinal {
	switch {
	case j >= 1:
		return OrdinalIdentical
	case j >= 0.5:
		return OrdinalHigh
	case j > 0:
		return OrdinalLow
	default:
		return OrdinalNone
	}
}

// classifyV1 applies the versioned decisive-field rule:
//
//   - mechanism-near: every decisive field is identical-or-high and none is
//     low/none/incomparable;
//   - mechanism-distinct: any decisive field is low/none;
//   - unknown: otherwise (e.g. all decisive fields incomparable).
//
// Surface axis: representations similarity (non-decisive) produces the
// surface-distinct/surface-near qualifier when a mechanistic verdict exists.
func classifyV1(cmp Comparison) Classification {
	byKind := map[domain.FieldKind]FieldResult{}
	for _, f := range cmp.Fields {
		byKind[f.FieldKind] = f
	}

	anyDistinct := false
	allNearOrBetter := true
	anyComparable := false
	for _, k := range decisiveFields {
		f, ok := byKind[k]
		if !ok {
			continue
		}
		if f.Incomparable {
			allNearOrBetter = false
			continue
		}
		anyComparable = true
		switch f.Ordinal {
		case OrdinalNone, OrdinalLow:
			anyDistinct = true
		case OrdinalHigh, OrdinalIdentical:
			// still near
		}
	}

	var mech Classification
	switch {
	case anyDistinct:
		mech = ClassMechanismDistinct
	case anyComparable && allNearOrBetter:
		mech = ClassMechanismNear
	default:
		return ClassUnknown
	}

	// Surface qualifier from the non-decisive representation field, when present.
	rep, hasRep := byKind[domain.FieldRepresentation]
	surfaceDistinct := hasRep && !rep.Incomparable && (rep.Ordinal == OrdinalLow || rep.Ordinal == OrdinalNone)
	surfaceNear := hasRep && !rep.Incomparable && (rep.Ordinal == OrdinalHigh || rep.Ordinal == OrdinalIdentical)

	switch mech {
	case ClassMechanismNear:
		if surfaceDistinct {
			return ClassSurfaceDistinctMechNear
		}
		return ClassMechanismNear
	case ClassMechanismDistinct:
		if surfaceNear {
			return ClassSurfaceNearMechDistinct
		}
		return ClassMechanismDistinct
	default:
		return ClassUnknown
	}
}

// SurfaceStrings computes an auxiliary surface-text Jaccard over normalized
// tokens. It is diagnostic only and never influences classification.
func SurfaceStrings(a, b []string) SurfaceSimilarity {
	tokenize := func(ss []string) map[string]struct{} {
		out := map[string]struct{}{}
		for _, s := range ss {
			for _, tok := range splitTokens(Normalize(s)) {
				out[tok] = struct{}{}
			}
		}
		return out
	}
	ta, tb := tokenize(a), tokenize(b)
	overlap, union := 0, 0
	seen := map[string]struct{}{}
	for tok := range ta {
		seen[tok] = struct{}{}
		if _, ok := tb[tok]; ok {
			overlap++
		}
	}
	for tok := range tb {
		seen[tok] = struct{}{}
	}
	union = len(seen)
	j := 0.0
	if union > 0 {
		j = float64(overlap) / float64(union)
	}
	return SurfaceSimilarity{Jaccard: j, AuxiliaryOnly: true}
}

func splitTokens(normalized string) []string {
	if normalized == "" {
		return nil
	}
	toks := []string{}
	cur := ""
	for _, r := range normalized {
		if r == ' ' {
			if cur != "" {
				toks = append(toks, cur)
				cur = ""
			}
			continue
		}
		cur += string(r)
	}
	if cur != "" {
		toks = append(toks, cur)
	}
	sort.Strings(toks)
	return toks
}

// AssertDiscriminationPreserved is the system-owned abstraction-loss invariant
// (R9). For every pair of signatures whose outcome.class differs, it recomputes
// an outcome-excluded fingerprint; if those are equal, the vocabulary has
// erased every non-outcome distinction between two mechanisms known to differ
// in outcome — a discrimination-loss defect. It returns the offending pairs.
//
// This is not a hand-checked equality: any lossy vocabulary that collapses
// outcome-predictive structure is caught, including ones introduced later.
func AssertDiscriminationPreserved(signatures []MechanismSignature) []DiscriminationLoss {
	var losses []DiscriminationLoss
	for i := 0; i < len(signatures); i++ {
		for j := i + 1; j < len(signatures); j++ {
			a, b := signatures[i], signatures[j]
			if a.OutcomeClass == b.OutcomeClass {
				continue
			}
			if outcomeExcludedFingerprint(a) == outcomeExcludedFingerprint(b) {
				losses = append(losses, DiscriminationLoss{
					MechanismA: a.MechanismID,
					MechanismB: b.MechanismID,
					OutcomeA:   a.OutcomeClass,
					OutcomeB:   b.OutcomeClass,
					Vocabulary: a.VocabularyVersion,
				})
			}
		}
	}
	return losses
}

// DiscriminationLoss records one detected abstraction-loss defect.
type DiscriminationLoss struct {
	MechanismA string
	MechanismB string
	OutcomeA   domain.OutcomeClass
	OutcomeB   domain.OutcomeClass
	Vocabulary string
}
