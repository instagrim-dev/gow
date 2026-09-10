package policy

import (
	"testing"

	"github.com/instagrim-dev/newf/internal/domain"
)

func admitEvidence() Evidence {
	return Evidence{
		Successes: []SuccessEvidence{
			{PredicateFingerprint: "fp-supported", Strength: "deterministic", DistinctSupport: 2},
			{PredicateFingerprint: "fp-zero", Strength: "deterministic", DistinctSupport: 0},
		},
		Surviving: []SurvivingEvidence{
			{InvariantID: "inv_surv", Attested: false},
			{InvariantID: "inv_att", Attested: true},
		},
		UncoveredFamilies: []string{"locality"},
		RedundantAttacks:  []string{"attack-key"},
		RepeatedFailures:  []string{"fail-fp"},
	}
}

// The central regression: a provider preference targeting an existing
// ZERO-SUPPORT success invariant must remain inert. A resolvable reference is
// not sufficient evidence for the requested action — the admission gate
// enforces the SAME support requirement Derive applies to its own output.
func TestAdmitRejectsZeroSupportPreference(t *testing.T) {
	ev := admitEvidence()
	if _, admitted := AdmitProposedDirective(KindPrefer, TargetSuccessInvariant, "fp-zero", ev); admitted {
		t.Fatal("zero-support success invariant must earn no preference, even via a provider")
	}
	// Sanity: Derive itself also refuses it.
	for _, d := range Derive(ev).Directives {
		if d.TargetID == "fp-zero" {
			t.Fatalf("Derive emitted a zero-support directive: %+v", d)
		}
	}
}

func TestAdmitSupportedPreferenceCappedAtMedium(t *testing.T) {
	ev := admitEvidence()
	d, admitted := AdmitProposedDirective(KindPrefer, TargetSuccessInvariant, "fp-supported", ev)
	if !admitted {
		t.Fatal("supported preference must be admitted")
	}
	// Deterministic strength earns HIGH in Derive; a provider proposal is capped
	// at medium — it can confirm evidence-backed bias, never amplify it.
	if d.Weight != domain.OrdinalMedium {
		t.Fatalf("provider weight = %s, want medium (evidence-derived, capped)", d.Weight)
	}
}

// Kind/target pairings are the ones Derive emits; a prefer targeting a
// SURVIVING invariant (which Derive only ever avoids) is inert.
func TestAdmitRejectsForeignKindTargetPairings(t *testing.T) {
	ev := admitEvidence()
	cases := []struct {
		kind Kind
		tk   TargetKind
		id   string
	}{
		{KindPrefer, TargetSurvivingInvariant, "inv_surv"},
		{KindAvoid, TargetSuccessInvariant, "fp-supported"},
		{KindPenalize, TargetMechanismFamily, "locality"},
		{KindExpand, TargetRedundantAttack, "attack-key"},
	}
	for _, c := range cases {
		if _, admitted := AdmitProposedDirective(c.kind, c.tk, c.id, ev); admitted {
			t.Fatalf("pairing %s/%s must be inert", c.kind, c.tk)
		}
	}
}

func TestAdmitRejectsUnresolvableTargets(t *testing.T) {
	ev := admitEvidence()
	if _, admitted := AdmitProposedDirective(KindPrefer, TargetSuccessInvariant, "fp-missing", ev); admitted {
		t.Fatal("unresolvable target must be inert")
	}
	if _, admitted := AdmitProposedDirective(KindAvoid, TargetSurvivingInvariant, "inv_missing", ev); admitted {
		t.Fatal("unresolvable surviving target must be inert")
	}
}

func TestAdmitAttestedAvoidStaysCapped(t *testing.T) {
	ev := admitEvidence()
	d, admitted := AdmitProposedDirective(KindAvoid, TargetSurvivingInvariant, "inv_att", ev)
	if !admitted {
		t.Fatal("attested avoid must be admitted")
	}
	if d.Weight != domain.OrdinalMedium {
		t.Fatalf("provider avoid weight = %s, want medium cap (Derive itself grants high)", d.Weight)
	}
}
