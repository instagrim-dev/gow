package canon

import (
	"testing"

	"github.com/instagrim-dev/newf/internal/domain"
)

func baseInput(mechanismID string) MechanismInput {
	return MechanismInput{
		MechanismID: mechanismID,
		Posture: Posture{
			Locality:     domain.LocalityLocal,
			Construction: domain.ConstructionConstructive,
			Uncertainty:  domain.UncertaintyDeterministic,
		},
		OutcomeClass: domain.OutcomePartialSuccess,
		Claims: []MechanismClaimInput{
			{FieldKind: domain.FieldOperator, SurfaceLabel: "modular decomposition", Status: domain.ClaimExplicit},
			{FieldKind: domain.FieldPreserves, SurfaceLabel: "residue locality", Status: domain.ClaimExplicit},
		},
	}
}

func TestFingerprintOrderIndependence(t *testing.T) {
	t.Parallel()
	v := MechanismV1()
	a := baseInput("mech_a")
	b := baseInput("mech_b")
	// Reverse claim order in b.
	b.Claims[0], b.Claims[1] = b.Claims[1], b.Claims[0]
	fa := Fingerprint(BuildSignature(a, v))
	fb := Fingerprint(BuildSignature(b, v))
	if fa != fb {
		t.Fatalf("fingerprints differ by ordering: %s vs %s", fa, fb)
	}
}

func TestFingerprintEqualForDifferentWordingSameCanonical(t *testing.T) {
	t.Parallel()
	v := MechanismV1()
	a := baseInput("mech_a")
	b := baseInput("mech_b")
	// Different surface wording that canonicalizes identically.
	b.Claims = []MechanismClaimInput{
		{FieldKind: domain.FieldOperator, SurfaceLabel: "residue-by-residue reduction", Status: domain.ClaimExplicit},
		{FieldKind: domain.FieldPreserves, SurfaceLabel: "local congruence argument", Status: domain.ClaimInferred},
	}
	fa := Fingerprint(BuildSignature(a, v))
	fb := Fingerprint(BuildSignature(b, v))
	if fa != fb {
		t.Fatalf("different wording did not canonicalize equal: %s vs %s", fa, fb)
	}
}

func TestFingerprintVersionVisibility(t *testing.T) {
	t.Parallel()
	sig := BuildSignature(baseInput("mech_a"), MechanismV1())
	f1 := Fingerprint(sig)
	// Mutate vocabulary version; fingerprint must change.
	sig.VocabularyVersion = "mechanism/v2"
	f2 := Fingerprint(sig)
	if f1 == f2 {
		t.Fatal("changing vocabulary_version did not change fingerprint")
	}
	sig2 := BuildSignature(baseInput("mech_a"), MechanismV1())
	sig2.SchemaVersion = "mechanism/v9"
	if Fingerprint(sig2) == f1 {
		t.Fatal("changing schema_version did not change fingerprint")
	}
}

func TestFingerprintProvenanceInvariance(t *testing.T) {
	t.Parallel()
	v := MechanismV1()
	a := baseInput("mech_a")
	b := baseInput("mech_b")
	// b carries different provenance but same canonical content.
	for i := range b.Claims {
		b.Claims[i].SupportLocator = "p.99"
		b.Claims[i].Confidence = "high"
		b.Claims[i].SupportSnapshotID = "snap_different"
	}
	if Fingerprint(BuildSignature(a, v)) != Fingerprint(BuildSignature(b, v)) {
		t.Fatal("provenance changed the fingerprint")
	}
}

func TestBuildSignaturePreservesClaimStatus(t *testing.T) {
	t.Parallel()
	in := baseInput("mech_a")
	in.Claims[1].Status = domain.ClaimInferred
	sig := BuildSignature(in, MechanismV1())
	if len(sig.Preserves) != 1 {
		t.Fatalf("expected 1 preserves claim, got %d", len(sig.Preserves))
	}
	if sig.Preserves[0].Status != domain.ClaimInferred {
		t.Fatalf("claim status upgraded to %q, want inferred", sig.Preserves[0].Status)
	}
	if sig.Preserves[0].State != domain.ResolutionResolved {
		t.Fatalf("expected resolved, got %q", sig.Preserves[0].State)
	}
}

func TestUnresolvedExcludedFromFingerprintButRetained(t *testing.T) {
	t.Parallel()
	v := MechanismV1()
	base := baseInput("mech_a")
	withUnknown := baseInput("mech_b")
	withUnknown.Claims = append(withUnknown.Claims, MechanismClaimInput{
		FieldKind:    domain.FieldOperator,
		SurfaceLabel: "a totally unknown operator phrase",
		Status:       domain.ClaimInferred,
	})
	sigBase := BuildSignature(base, v)
	sigUnknown := BuildSignature(withUnknown, v)
	// The unknown claim is retained on the struct...
	if len(sigUnknown.Operators) != 2 {
		t.Fatalf("expected unknown claim retained, operators = %d", len(sigUnknown.Operators))
	}
	// ...but excluded from the hashed body, so fingerprints match.
	if Fingerprint(sigBase) != Fingerprint(sigUnknown) {
		t.Fatal("unknown claim leaked into fingerprint")
	}
}

func TestEmptyFieldsYieldStableFingerprint(t *testing.T) {
	t.Parallel()
	v := MechanismV1()
	in := MechanismInput{
		MechanismID:  "mech_empty",
		Posture:      Posture{Locality: domain.LocalityUnknown, Construction: domain.ConstructionUnknown, Uncertainty: domain.UncertaintyUnknown},
		OutcomeClass: domain.OutcomeUnknown,
	}
	sig := BuildSignature(in, v)
	f1 := Fingerprint(sig)
	f2 := Fingerprint(BuildSignature(in, v))
	if f1 != f2 {
		t.Fatal("empty signature fingerprint not stable")
	}
	if f1 == "" {
		t.Fatal("empty fingerprint")
	}
}

func TestOutcomeInFingerprint(t *testing.T) {
	t.Parallel()
	v := MechanismV1()
	a := baseInput("mech_a")
	b := baseInput("mech_b")
	b.OutcomeClass = domain.OutcomeFailure
	if Fingerprint(BuildSignature(a, v)) == Fingerprint(BuildSignature(b, v)) {
		t.Fatal("differing outcome.class produced equal fingerprints (outcome not hashed)")
	}
}

// TestBuildSignatureDeclaredCompletenessOverlay (v25): a justified declaration
// upgrades exactly the declared field; every undeclared field keeps the
// conservative unobserved default, and an invalid declared value is ignored.
func TestBuildSignatureDeclaredCompletenessOverlay(t *testing.T) {
	t.Parallel()
	v := MechanismV1()
	in := baseInput("mech_a")
	in.DeclaredCompleteness = map[domain.FieldKind]domain.FieldCompleteness{
		domain.FieldPreserves: domain.CompletenessComplete,
		domain.FieldOperator:  domain.FieldCompleteness("bogus"), // ignored
	}
	sig := BuildSignature(in, v)
	if got := sig.FieldCompleteness(domain.FieldPreserves); got != domain.CompletenessComplete {
		t.Fatalf("declared preserves completeness = %q, want complete", got)
	}
	if got := sig.FieldCompleteness(domain.FieldOperator); got != domain.CompletenessUnobserved {
		t.Fatalf("invalid declared value must not upgrade the field, got %q", got)
	}
	if got := sig.FieldCompleteness(domain.FieldBreaks); got != domain.CompletenessUnobserved {
		t.Fatalf("undeclared field must stay unobserved, got %q", got)
	}

	// The overlay must not change mechanism identity: fingerprint deliberately
	// excludes extraction completeness (evidence revision, not identity).
	plain := BuildSignature(baseInput("mech_a"), v)
	if Fingerprint(sig) != Fingerprint(plain) {
		t.Fatal("completeness overlay must not change the mechanism fingerprint")
	}
}
