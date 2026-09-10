package canon

import (
	"testing"

	"github.com/instagrim-dev/newf/internal/domain"
)

func sigFrom(t *testing.T, in MechanismInput, v *Vocabulary) MechanismSignature {
	t.Helper()
	return BuildSignature(in, v)
}

func TestCompareIdentical(t *testing.T) {
	t.Parallel()
	v := MechanismV1()
	sig := sigFrom(t, baseInput("mech_a"), v)
	cmp, err := Compare(sig, sig, "")
	if err != nil {
		t.Fatalf("Compare() error = %v", err)
	}
	if cmp.Classification != ClassMechanismNear {
		t.Fatalf("classification = %q, want mechanism-near", cmp.Classification)
	}
	for _, f := range cmp.Fields {
		if f.Incomparable {
			t.Fatalf("field %q incomparable on identical signatures", f.FieldKind)
		}
	}
}

func TestCompareSurfaceDistinctMechanismNear(t *testing.T) {
	t.Parallel()
	v := MechanismV1()
	a := baseInput("mech_a")
	b := baseInput("mech_b")
	b.Claims = []MechanismClaimInput{
		{FieldKind: domain.FieldOperator, SurfaceLabel: "residue-by-residue reduction", Status: domain.ClaimExplicit},
		{FieldKind: domain.FieldPreserves, SurfaceLabel: "local congruence argument", Status: domain.ClaimExplicit},
	}
	cmp, err := Compare(sigFrom(t, a, v), sigFrom(t, b, v), "")
	if err != nil {
		t.Fatalf("Compare() error = %v", err)
	}
	// Decisive fields identical -> mechanism-near (no representations differ here).
	if cmp.Classification != ClassMechanismNear {
		t.Fatalf("classification = %q, want mechanism-near", cmp.Classification)
	}
}

func TestCompareSurfaceNearMechanismDistinct(t *testing.T) {
	t.Parallel()
	v := MechanismV1()
	// Same representation (surface-near), different decisive operators/preserves.
	a := MechanismInput{
		MechanismID:  "mech_a",
		Posture:      Posture{Locality: domain.LocalityLocal, Construction: domain.ConstructionConstructive, Uncertainty: domain.UncertaintyDeterministic},
		OutcomeClass: domain.OutcomePartialSuccess,
		Claims: []MechanismClaimInput{
			{FieldKind: domain.FieldRepresentation, SurfaceLabel: "congruence classes", Status: domain.ClaimExplicit},
			{FieldKind: domain.FieldOperator, SurfaceLabel: "modular decomposition", Status: domain.ClaimExplicit},
			{FieldKind: domain.FieldPreserves, SurfaceLabel: "residue locality", Status: domain.ClaimExplicit},
		},
	}
	b := MechanismInput{
		MechanismID:  "mech_b",
		Posture:      Posture{Locality: domain.LocalityGlobal, Construction: domain.ConstructionExistential, Uncertainty: domain.UncertaintyProbabilistic},
		OutcomeClass: domain.OutcomeFailure,
		Claims: []MechanismClaimInput{
			{FieldKind: domain.FieldRepresentation, SurfaceLabel: "congruence classes", Status: domain.ClaimExplicit},
			{FieldKind: domain.FieldOperator, SurfaceLabel: "density averaging", Status: domain.ClaimExplicit},
			{FieldKind: domain.FieldPreserves, SurfaceLabel: "mean growth rate", Status: domain.ClaimExplicit},
		},
	}
	cmp, err := Compare(sigFrom(t, a, v), sigFrom(t, b, v), "")
	if err != nil {
		t.Fatalf("Compare() error = %v", err)
	}
	if cmp.Classification != ClassSurfaceNearMechDistinct {
		t.Fatalf("classification = %q, want surface-near+mechanism-distinct", cmp.Classification)
	}
}

func TestCompareUnresolvedIsIncomparable(t *testing.T) {
	t.Parallel()
	v := MechanismV1()
	a := baseInput("mech_a")
	b := baseInput("mech_b")
	// Add an unknown operator on b -> operator field incomparable.
	b.Claims = append(b.Claims, MechanismClaimInput{
		FieldKind: domain.FieldOperator, SurfaceLabel: "unseen phrase", Status: domain.ClaimInferred,
	})
	cmp, err := Compare(sigFrom(t, a, v), sigFrom(t, b, v), "")
	if err != nil {
		t.Fatalf("Compare() error = %v", err)
	}
	var opField FieldResult
	for _, f := range cmp.Fields {
		if f.FieldKind == domain.FieldOperator {
			opField = f
		}
	}
	if !opField.Incomparable {
		t.Fatal("operator field with unknown claim not marked incomparable")
	}
	// With a decisive field incomparable and the other decisive field near,
	// classification must not be mechanism-near on the strength of equal resolved sets.
	if cmp.Classification == ClassMechanismNear {
		t.Fatalf("classification = mechanism-near despite incomparable decisive field")
	}
}

func TestCompareDecisiveVsNonDecisiveFlip(t *testing.T) {
	t.Parallel()
	v := MechanismV1()
	// Agree on decisive (operator+preserves), differ on representation (non-decisive)
	// -> mechanism-near / surface-distinct.
	agreeDecisive := MechanismInput{
		MechanismID:  "mech_a",
		Posture:      Posture{Locality: domain.LocalityLocal, Construction: domain.ConstructionConstructive, Uncertainty: domain.UncertaintyDeterministic},
		OutcomeClass: domain.OutcomePartialSuccess,
		Claims: []MechanismClaimInput{
			{FieldKind: domain.FieldRepresentation, SurfaceLabel: "congruence classes", Status: domain.ClaimExplicit},
			{FieldKind: domain.FieldOperator, SurfaceLabel: "modular decomposition", Status: domain.ClaimExplicit},
			{FieldKind: domain.FieldPreserves, SurfaceLabel: "residue locality", Status: domain.ClaimExplicit},
		},
	}
	differRep := agreeDecisive
	differRep.MechanismID = "mech_b"
	differRep.Claims = []MechanismClaimInput{
		// No representation at all -> representation sets differ? Both empty vs one.
		{FieldKind: domain.FieldOperator, SurfaceLabel: "modular decomposition", Status: domain.ClaimExplicit},
		{FieldKind: domain.FieldPreserves, SurfaceLabel: "residue locality", Status: domain.ClaimExplicit},
	}
	cmp, err := Compare(sigFrom(t, agreeDecisive, v), sigFrom(t, differRep, v), "")
	if err != nil {
		t.Fatalf("Compare() error = %v", err)
	}
	// Decisive fields agree -> near. Representation differs (one has it, one doesn't).
	if cmp.Classification != ClassMechanismNear && cmp.Classification != ClassSurfaceDistinctMechNear {
		t.Fatalf("classification = %q, want a mechanism-near variant", cmp.Classification)
	}

	// Now flip: differ on decisive operator while representation agrees.
	differDecisive := agreeDecisive
	differDecisive.MechanismID = "mech_c"
	differDecisive.Claims = []MechanismClaimInput{
		{FieldKind: domain.FieldRepresentation, SurfaceLabel: "congruence classes", Status: domain.ClaimExplicit},
		{FieldKind: domain.FieldOperator, SurfaceLabel: "density averaging", Status: domain.ClaimExplicit},
		{FieldKind: domain.FieldPreserves, SurfaceLabel: "mean growth rate", Status: domain.ClaimExplicit},
	}
	cmp2, err := Compare(sigFrom(t, agreeDecisive, v), sigFrom(t, differDecisive, v), "")
	if err != nil {
		t.Fatalf("Compare() error = %v", err)
	}
	if cmp2.Classification != ClassMechanismDistinct && cmp2.Classification != ClassSurfaceNearMechDistinct {
		t.Fatalf("classification = %q, want a mechanism-distinct variant", cmp2.Classification)
	}
}

// TestBreakIsDecisive_ESAffineLattice is the falsifiable guard for the
// classify/v1 decisive-field contract (issue #9 finding). Two mechanisms share
// *identical* operators, assumptions, and preserves, but one breaks residue-
// class locality (and introduces an affine-lattice auxiliary object) — exactly
// the ES -> affine-lattice move. The project thesis is that such a break IS the
// mechanism change, so the verdict must be mechanism-distinct. Under the old
// three-field classifier (operator/assumption/preserves only) this case
// wrongly returned mechanism-near; this test fails if that regression returns.
func TestBreakIsDecisive_ESAffineLattice(t *testing.T) {
	t.Parallel()
	v := MechanismV1()

	// Congruence-local family: works residue-by-residue, preserves locality,
	// breaks nothing, introduces no auxiliary object.
	local := MechanismInput{
		MechanismID:  "mech_local",
		Posture:      Posture{Locality: domain.LocalityLocal, Construction: domain.ConstructionConstructive, Uncertainty: domain.UncertaintyDeterministic},
		OutcomeClass: domain.OutcomeFailure,
		Claims: []MechanismClaimInput{
			{FieldKind: domain.FieldOperator, SurfaceLabel: "modular decomposition", Status: domain.ClaimExplicit},
			{FieldKind: domain.FieldAssumption, SurfaceLabel: "residue independence", Status: domain.ClaimExplicit},
			{FieldKind: domain.FieldPreserves, SurfaceLabel: "residue locality", Status: domain.ClaimExplicit},
		},
	}
	// Affine-lattice family: SAME operator/assumption/preserves surface labels,
	// but it *breaks* residue-class locality and introduces an affine lattice.
	affine := MechanismInput{
		MechanismID:  "mech_affine",
		Posture:      Posture{Locality: domain.LocalityLocal, Construction: domain.ConstructionConstructive, Uncertainty: domain.UncertaintyDeterministic},
		OutcomeClass: domain.OutcomePartialSuccess,
		Claims: []MechanismClaimInput{
			{FieldKind: domain.FieldOperator, SurfaceLabel: "modular decomposition", Status: domain.ClaimExplicit},
			{FieldKind: domain.FieldAssumption, SurfaceLabel: "residue independence", Status: domain.ClaimExplicit},
			{FieldKind: domain.FieldPreserves, SurfaceLabel: "residue locality", Status: domain.ClaimExplicit},
			{FieldKind: domain.FieldBreaks, SurfaceLabel: "residue-class locality", Status: domain.ClaimExplicit},
			{FieldKind: domain.FieldAuxiliaryObject, SurfaceLabel: "affine lattice", Status: domain.ClaimExplicit},
		},
	}

	cmp, err := Compare(sigFrom(t, local, v), sigFrom(t, affine, v), "")
	if err != nil {
		t.Fatalf("Compare() error = %v", err)
	}
	switch cmp.Classification {
	case ClassMechanismDistinct, ClassSurfaceNearMechDistinct:
		// correct: the break/auxiliary difference is decisive
	default:
		t.Fatalf("classification = %q, want a mechanism-distinct variant; a mechanism break must be decisive, not surface", cmp.Classification)
	}

	// The decisive difference must be visible on the breaks/auxiliary fields.
	byKind := map[domain.FieldKind]FieldResult{}
	for _, f := range cmp.Fields {
		byKind[f.FieldKind] = f
	}
	if b := byKind[domain.FieldBreaks]; b.Ordinal == OrdinalIdentical {
		t.Fatal("breaks field reported identical despite one side breaking residue-class locality")
	}
	if a := byKind[domain.FieldAuxiliaryObject]; a.Ordinal == OrdinalIdentical {
		t.Fatal("auxiliary_object field reported identical despite one side introducing an affine lattice")
	}
}

// TestAuxiliaryObjectAloneIsDecisive proves introducing an auxiliary object,
// with everything else identical, is a mechanism change (not surface).
func TestAuxiliaryObjectAloneIsDecisive(t *testing.T) {
	t.Parallel()
	v := MechanismV1()
	base := MechanismInput{
		MechanismID:  "mech_base",
		Posture:      Posture{Locality: domain.LocalityLocal, Construction: domain.ConstructionConstructive, Uncertainty: domain.UncertaintyDeterministic},
		OutcomeClass: domain.OutcomeFailure,
		Claims: []MechanismClaimInput{
			{FieldKind: domain.FieldOperator, SurfaceLabel: "modular decomposition", Status: domain.ClaimExplicit},
			{FieldKind: domain.FieldPreserves, SurfaceLabel: "residue locality", Status: domain.ClaimExplicit},
		},
	}
	withAux := base
	withAux.MechanismID = "mech_aux"
	withAux.Claims = append(append([]MechanismClaimInput{}, base.Claims...),
		MechanismClaimInput{FieldKind: domain.FieldAuxiliaryObject, SurfaceLabel: "affine lattice", Status: domain.ClaimExplicit},
	)
	cmp, err := Compare(sigFrom(t, base, v), sigFrom(t, withAux, v), "")
	if err != nil {
		t.Fatalf("Compare() error = %v", err)
	}
	switch cmp.Classification {
	case ClassMechanismDistinct, ClassSurfaceNearMechDistinct:
		// correct
	default:
		t.Fatalf("classification = %q, want mechanism-distinct: an auxiliary-object introduction is a structural move", cmp.Classification)
	}
}

// TestComparisonProfileIsCallerControlled proves the core item-2 contract: the
// comparator MEASURES every axis and the PROFILE decides what is decisive. The
// same two signatures — identical structural fields, differing only in
// representation — classify as mechanism-near under the default profile (where
// representation is surface) but mechanism-distinct under a caller profile that
// promotes representation to decisive. No re-measurement, only reinterpretation.
func TestComparisonProfileIsCallerControlled(t *testing.T) {
	t.Parallel()
	v := MechanismV1()
	a := MechanismInput{
		MechanismID:  "mech_a",
		Posture:      Posture{Locality: domain.LocalityLocal, Construction: domain.ConstructionConstructive, Uncertainty: domain.UncertaintyDeterministic},
		OutcomeClass: domain.OutcomeFailure,
		Claims: []MechanismClaimInput{
			{FieldKind: domain.FieldOperator, SurfaceLabel: "modular decomposition", Status: domain.ClaimExplicit},
			{FieldKind: domain.FieldRepresentation, SurfaceLabel: "congruence classes", Status: domain.ClaimExplicit},
		},
	}
	b := MechanismInput{
		MechanismID:  "mech_b",
		Posture:      Posture{Locality: domain.LocalityLocal, Construction: domain.ConstructionConstructive, Uncertainty: domain.UncertaintyDeterministic},
		OutcomeClass: domain.OutcomeFailure,
		Claims: []MechanismClaimInput{
			{FieldKind: domain.FieldOperator, SurfaceLabel: "modular decomposition", Status: domain.ClaimExplicit},
			// Different representation (congruence classes vs affine lattice),
			// same operator. Under default profile: surface-only difference.
			{FieldKind: domain.FieldRepresentation, SurfaceLabel: "affine lattice", Status: domain.ClaimExplicit},
		},
	}
	sigA, sigB := sigFrom(t, a, v), sigFrom(t, b, v)

	// Default profile: representation is surface -> mechanism-near variant.
	def := CompareWithProfile(sigA, sigB, ProfileMechanismV1())
	switch def.Classification {
	case ClassMechanismNear, ClassSurfaceDistinctMechNear:
		// correct
	default:
		t.Fatalf("default profile classification = %q, want a mechanism-near variant", def.Classification)
	}

	// Caller profile that promotes representation to decisive -> distinct.
	repDecisive := ComparisonProfile{
		Version:           "classify/rep-decisive-test",
		DecisiveSetFields: []domain.FieldKind{domain.FieldOperator, domain.FieldRepresentation},
	}
	got := CompareWithProfile(sigA, sigB, repDecisive)
	if got.Classification != ClassMechanismDistinct {
		t.Fatalf("rep-decisive profile classification = %q, want mechanism-distinct", got.Classification)
	}

	// The MEASUREMENT must be identical under both profiles: only the verdict
	// and the recorded profile version differ. No axis is hidden by selection.
	repDef := fieldByKind(def.Fields, domain.FieldRepresentation)
	repGot := fieldByKind(got.Fields, domain.FieldRepresentation)
	if repDef.Ordinal != repGot.Ordinal || repDef.Jaccard != repGot.Jaccard {
		t.Fatalf("representation measurement differed across profiles: %+v vs %+v", repDef, repGot)
	}
	if def.ClassifyVersion == got.ClassifyVersion {
		t.Fatal("profiles must record distinct classify versions")
	}
}

func fieldByKind(fields []FieldResult, kind domain.FieldKind) FieldResult {
	for _, f := range fields {
		if f.FieldKind == kind {
			return f
		}
	}
	return FieldResult{}
}

func TestBoundaryRelationIsDecisiveInComparison(t *testing.T) {
	t.Parallel()
	// Two signatures whose boundaries share a canonical ID but differ in the
	// typed relation must NOT compare as identical boundaries: stops_at(X) is a
	// different failure structure than requires(X). This guards the prior defect
	// where boundaryClaims reduced boundaries to canonical ID only, dropping the
	// relation the fingerprint already encodes as "canonicalID|relation".
	base := MechanismSignature{
		SchemaVersion:     "mechanism/v1",
		VocabularyVersion: "mechanism/v1",
	}
	stopsAt := base
	stopsAt.MechanismID = "mech_stops"
	stopsAt.Boundaries = []Boundary{{
		SurfaceLabel: "composite modulus",
		State:        domain.ResolutionResolved,
		CanonicalID:  domain.CanonicalID("core.boundary.composite_modulus"),
		Relation:     "stops_at",
	}}
	requires := base
	requires.MechanismID = "mech_requires"
	requires.Boundaries = []Boundary{{
		SurfaceLabel: "composite modulus",
		State:        domain.ResolutionResolved,
		CanonicalID:  domain.CanonicalID("core.boundary.composite_modulus"),
		Relation:     "requires",
	}}

	cmp := CompareWithProfile(stopsAt, requires, ProfileMechanismV1())
	if cmp.Boundary.Incomparable {
		t.Fatalf("boundary comparison incomparable, want a resolved-set result: %+v", cmp.Boundary)
	}
	if cmp.Boundary.Ordinal == OrdinalIdentical || cmp.Boundary.OverlapCount != 0 {
		t.Fatalf("stops_at(X) vs requires(X) compared identical (overlap=%d, ordinal=%q); the typed relation must be decisive", cmp.Boundary.OverlapCount, cmp.Boundary.Ordinal)
	}

	// Same relation + same canonical ID must still compare identical (the
	// composite key does not spuriously distinguish equal boundaries).
	same := CompareWithProfile(stopsAt, stopsAt, ProfileMechanismV1())
	if same.Boundary.Ordinal != OrdinalIdentical {
		t.Fatalf("identical boundaries compared non-identical: %+v", same.Boundary)
	}
}

func TestComparisonProfileHashStableAndContentSensitive(t *testing.T) {
	t.Parallel()
	p := ProfileMechanismV1()

	// Stable: two calls to the default profile hash equal, and independent of the
	// declared decisive-field ordering.
	if p.Hash() != ProfileMechanismV1().Hash() {
		t.Fatal("default profile hash is not stable across calls")
	}
	reordered := ProfileMechanismV1()
	reordered.DecisiveSetFields = []domain.FieldKind{
		domain.FieldAuxiliaryObject, domain.FieldBreaks, domain.FieldAssumption,
		domain.FieldOperator, domain.FieldPreserves,
	}
	if reordered.Hash() != p.Hash() {
		t.Fatal("profile hash changed under decisive-field reordering; must be order-independent")
	}

	// Content-sensitive: promoting representation to decisive changes the hash,
	// even if the human-authored Version string is left identical (the exact
	// silent-drift the hash exists to catch).
	promoted := ProfileMechanismV1()
	promoted.DecisiveSetFields = append(promoted.DecisiveSetFields, domain.FieldRepresentation)
	if promoted.Hash() == p.Hash() {
		t.Fatal("adding a decisive field did not change the profile hash")
	}
	posture := ProfileMechanismV1()
	posture.DecisivePosture = true
	if posture.Hash() == p.Hash() {
		t.Fatal("flipping DecisivePosture did not change the profile hash")
	}

	// The hash flows onto the comparison verdict for #11 to persist.
	cmp := CompareWithProfile(MechanismSignature{}, MechanismSignature{}, p)
	if cmp.ProfileHash != p.Hash() {
		t.Fatalf("comparison profile hash = %q, want %q", cmp.ProfileHash, p.Hash())
	}
}

func TestCompareUnknownWeightsVersionErrors(t *testing.T) {
	t.Parallel()
	v := MechanismV1()
	sig := sigFrom(t, baseInput("mech_a"), v)
	if _, err := Compare(sig, sig, "weights/bogus"); err == nil {
		t.Fatal("Compare() accepted unknown weights version")
	}
}

func TestJaccardMath(t *testing.T) {
	t.Parallel()
	a := []domain.CanonicalID{"core.operator.a", "core.operator.b"}
	b := []domain.CanonicalID{"core.operator.b", "core.operator.c"}
	overlap, union := intersectUnion(a, b)
	if overlap != 1 || union != 3 {
		t.Fatalf("overlap=%d union=%d, want 1,3", overlap, union)
	}
}

func TestAssertDiscriminationPreservedFiresUnderLossyVocab(t *testing.T) {
	t.Parallel()
	// Two mechanisms with different outcomes whose only distinguishing structure
	// is the operator (modular decomposition vs density averaging).
	mkInput := func(id, op string, outcome domain.OutcomeClass) MechanismInput {
		return MechanismInput{
			MechanismID:  id,
			Posture:      Posture{Locality: domain.LocalityLocal, Construction: domain.ConstructionConstructive, Uncertainty: domain.UncertaintyDeterministic},
			OutcomeClass: outcome,
			Claims: []MechanismClaimInput{
				{FieldKind: domain.FieldOperator, SurfaceLabel: op, Status: domain.ClaimExplicit},
			},
		}
	}
	a := mkInput("mech_a", "modular decomposition", domain.OutcomePartialSuccess)
	b := mkInput("mech_b", "density averaging", domain.OutcomeFailure)

	// Under v1 the operators canonicalize distinctly -> no discrimination loss.
	v1 := MechanismV1()
	v1sigs := []MechanismSignature{BuildSignature(a, v1), BuildSignature(b, v1)}
	if losses := AssertDiscriminationPreserved(v1sigs); len(losses) != 0 {
		t.Fatalf("v1 reported discrimination loss: %+v", losses)
	}

	// Under the lossy vocab both operators collapse to one canonical ID, so the
	// two different-outcome mechanisms become indistinguishable (outcome-excluded)
	// -> the invariant must fire.
	v0 := MechanismV0Lossy()
	v0sigs := []MechanismSignature{BuildSignature(a, v0), BuildSignature(b, v0)}
	losses := AssertDiscriminationPreserved(v0sigs)
	if len(losses) != 1 {
		t.Fatalf("lossy vocab reported %d losses, want 1: %+v", len(losses), losses)
	}
	if losses[0].OutcomeA == losses[0].OutcomeB {
		t.Fatal("discrimination loss recorded equal outcomes")
	}
}

func TestSurfaceStringsAuxiliaryOnly(t *testing.T) {
	t.Parallel()
	sim := SurfaceStrings([]string{"modular decomposition"}, []string{"modular reduction"})
	if !sim.AuxiliaryOnly {
		t.Fatal("surface similarity not tagged auxiliary-only")
	}
	if sim.Jaccard <= 0 || sim.Jaccard >= 1 {
		t.Fatalf("expected partial overlap, got jaccard=%v", sim.Jaccard)
	}
}
