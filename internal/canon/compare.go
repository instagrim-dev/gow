package canon

import (
	"fmt"
	"sort"

	"github.com/instagrim-dev/newf/internal/domain"
)

// ComparisonProfile makes "which axes are decisive" an explicit, versioned,
// caller-owned decision instead of a universal constant baked into the
// comparator. The comparator MEASURES every axis (see Comparison.Fields, which
// always contains all six set fields, plus Posture and OutcomeEqual); a profile
// decides what those measurements MEAN for a given experiment or domain.
//
// This is deliberate: newf's thesis is that mechanistic distance is not
// universal — a representation change or an auxiliary object can be the
// mechanism break in one domain and mere surface in another. Clustering (#11)
// and invariant mining must be able to choose the profile per experiment rather
// than inherit a hidden global definition of "mechanism".
type ComparisonProfile struct {
	// Version is the stable identifier persisted alongside a comparison so a
	// verdict is always reproducible against the exact axis selection used.
	Version string
	// DecisiveSetFields are the set-valued field kinds whose (dis)agreement
	// determines the mechanistic verdict. Any field kind NOT listed here is
	// still measured and reported, but is non-decisive (surface/diagnostic).
	DecisiveSetFields []domain.FieldKind
	// SurfaceField names the set field used only to qualify a verdict as
	// surface-near/surface-distinct. It is never decisive. Empty disables the
	// surface qualifier.
	SurfaceField domain.FieldKind
	// DecisivePosture, when true, treats a posture-axis disagreement as a
	// decisive mechanism difference. Default profile keeps posture non-decisive
	// (reported only) pending posture provenance (#issue hardening).
	DecisivePosture bool
	// DecisiveOutcome, when true, treats differing outcome classes as a decisive
	// difference. Default keeps it reported-only (outcome is an effect, not a
	// mechanism).
	DecisiveOutcome bool
}

// WeightsMechanismV1 is retained as the default profile version identifier for
// backward compatibility with persisted comparison rows.
const WeightsMechanismV1 = "weights/v1"

// ClassifyMechanismV1 is the versioned mechanistic-vs-surface classification
// rule (the default profile's classifier contract).
const ClassifyMechanismV1 = "classify/v1"

// ProfileMechanismV1 is the default comparison profile. It treats the
// structural moves, conserved/violated properties, and auxiliary constructions
// as decisive, and representation as a surface qualifier only. Unlike the prior
// hardcoded list it is a value a caller can inspect, override, or replace with a
// domain-specific profile.
//
// breaks and auxiliary_objects are decisive because a deliberately violated
// invariant or a newly introduced auxiliary object IS a mechanism change (the
// ES -> affine-lattice family is productive precisely because of that break),
// not a rewording. representation stays non-decisive: the same move written over
// an equivalent representation should not, by itself, read as a distinct
// mechanism — but a representation change that carries a real structural change
// surfaces through breaks/auxiliary_objects, which are decisive.
func ProfileMechanismV1() ComparisonProfile {
	return ComparisonProfile{
		Version: ClassifyMechanismV1,
		DecisiveSetFields: []domain.FieldKind{
			domain.FieldPreserves,
			domain.FieldOperator,
			domain.FieldAssumption,
			domain.FieldBreaks,
			domain.FieldAuxiliaryObject,
		},
		SurfaceField:    domain.FieldRepresentation,
		DecisivePosture: false,
		DecisiveOutcome: false,
	}
}

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

// Comparison is the full, per-field, non-scalar comparison result. Every axis
// is measured and present regardless of which axes the profile treats as
// decisive, so no measurement is hidden by axis selection.
type Comparison struct {
	WeightsVersion  string
	ClassifyVersion string
	Fields          []FieldResult
	// Boundary is the measured (non-decisive by default) boundary-set result.
	Boundary       FieldResult
	Posture        PostureResult
	OutcomeEqual   bool
	Classification Classification
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

// Compare produces a deterministic, component-wise comparison of two signatures
// under the default profile (ProfileMechanismV1). It is a thin wrapper over
// CompareWithProfile kept for backward compatibility; weightsVersion continues
// to select the default profile version. An unknown version is an error so
// callers cannot silently compare under an absent config.
func Compare(a, b MechanismSignature, weightsVersion string) (Comparison, error) {
	if weightsVersion == "" {
		weightsVersion = WeightsMechanismV1
	}
	if weightsVersion != WeightsMechanismV1 {
		return Comparison{}, fmt.Errorf("unknown weights version %q", weightsVersion)
	}
	return CompareWithProfile(a, b, ProfileMechanismV1()), nil
}

// CompareWithProfile measures every axis and then applies the given profile to
// decide the verdict. The measurement (Comparison.Fields for all six set
// fields, Posture, OutcomeEqual, and per-field boundary results) is independent
// of the profile: the comparator MEASURES, the profile DECIDES. Two callers
// with different profiles see identical component data and only differ in the
// resulting Classification, so no measurement is ever hidden by axis selection.
func CompareWithProfile(a, b MechanismSignature, profile ComparisonProfile) Comparison {
	cmp := Comparison{
		WeightsVersion:  profile.Version,
		ClassifyVersion: profile.Version,
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

	// Boundaries are measured too, as a set over resolved canonical IDs, so a
	// profile MAY treat them as decisive without the comparator having to
	// re-run. They default to non-decisive in ProfileMechanismV1.
	cmp.Boundary = compareSetField(domain.FieldBoundary, boundaryClaims(a.Boundaries), boundaryClaims(b.Boundaries))

	cmp.Classification = classify(cmp, profile)
	return cmp
}

// boundaryClaims adapts a signature's boundaries into the FieldClaim shape so
// they can be compared with the same set machinery as the other fields.
func boundaryClaims(bs []Boundary) []FieldClaim {
	out := make([]FieldClaim, 0, len(bs))
	for _, b := range bs {
		out = append(out, FieldClaim{
			FieldKind:   domain.FieldBoundary,
			State:       b.State,
			CanonicalID: b.CanonicalID,
		})
	}
	return out
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

// classify applies a ComparisonProfile to already-measured components. The
// profile names which set fields (and optionally posture/outcome) are decisive:
//
//   - mechanism-near: every decisive field is identical-or-high and none is
//     low/none/incomparable;
//   - mechanism-distinct: any decisive field is low/none (so differing in what
//     an approach *breaks* or in its auxiliary construction alone is enough to
//     be distinct under the default profile);
//   - unknown: otherwise (e.g. all decisive fields incomparable).
//
// The profile's SurfaceField (non-decisive) only adds the
// surface-distinct/surface-near qualifier to an already-decided verdict; it can
// never flip the verdict, and no measured axis is discarded — a non-decisive
// axis is simply not consulted for THIS profile's verdict while remaining fully
// present in the Comparison for other profiles to use.
func classify(cmp Comparison, profile ComparisonProfile) Classification {
	byKind := map[domain.FieldKind]FieldResult{}
	for _, f := range cmp.Fields {
		byKind[f.FieldKind] = f
	}
	byKind[domain.FieldBoundary] = cmp.Boundary

	anyDistinct := false
	allNearOrBetter := true
	anyComparable := false
	consider := func(f FieldResult, present bool) {
		if !present {
			return
		}
		if f.Incomparable {
			allNearOrBetter = false
			return
		}
		anyComparable = true
		switch f.Ordinal {
		case OrdinalNone, OrdinalLow:
			anyDistinct = true
		case OrdinalHigh, OrdinalIdentical:
			// still near
		}
	}
	for _, k := range profile.DecisiveSetFields {
		f, ok := byKind[k]
		consider(f, ok)
	}
	// Posture / outcome may be promoted to decisive by the profile. When
	// decisive, a disagreement counts as a distinct-making difference; posture
	// and outcome are always comparable enums, so they never mark incomparable.
	if profile.DecisivePosture {
		anyComparable = true
		if !cmp.Posture.LocalityEqual || !cmp.Posture.ConstructionEqual || !cmp.Posture.UncertaintyEqual {
			anyDistinct = true
		}
	}
	if profile.DecisiveOutcome {
		anyComparable = true
		if !cmp.OutcomeEqual {
			anyDistinct = true
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

	// Surface qualifier from the profile's non-decisive surface field, if any.
	surfaceDistinct, surfaceNear := false, false
	if profile.SurfaceField != "" {
		if rep, ok := byKind[profile.SurfaceField]; ok && !rep.Incomparable {
			surfaceDistinct = rep.Ordinal == OrdinalLow || rep.Ordinal == OrdinalNone
			surfaceNear = rep.Ordinal == OrdinalHigh || rep.Ordinal == OrdinalIdentical
		}
	}

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
