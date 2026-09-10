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
