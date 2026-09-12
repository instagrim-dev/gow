package invariant

// Predeclared in
// corpus/experiments/pilot-004-discovery/records/l-coh-mechanism-test-predeclaration.md
// on 2026-09-12. This test operationalizes the reviewer's
// correction-7 refinement from the (l-coh) record:
//
//   Among emitted, support-valid candidates with comparable resolved
//   populations, absence of a preserving success prevents this
//   particular weakening route. Overall survival additionally
//   depends on the remaining challenge outcomes and claim semantics.
//
// Two fixtures share identical discrim (fPrev - sPrev = 0.5) but
// differ only in sPrev. If VerifySuccessPreserving returns
// OutcomeCompletedNegative on the zero-preserving-success fixture
// and OutcomeConfirmed on the three-preserving-success fixture, the
// probe's binary threshold on `∃ preserving success` is code-verified
// as the mechanism that distinguishes them at the challenge boundary.

import (
	"testing"

	"github.com/instagrim-dev/newf/internal/domain"
)

// buildDiscriminatingFixture assembles a resolved singleton-family
// contrast fixture:
//   - `failWithX` failure-side singleton families whose one member
//     preserves X (satisfies `contains(preserves, X)`),
//   - `failWithY` failure-side singleton families whose one member
//     preserves Y only (violates the predicate),
//   - `succWithX` success-side (partial_success) singleton families
//     preserving X (satisfy),
//   - `succWithY` success-side singleton families preserving Y only
//     (violate).
//
// All members have SetFieldCompleteness[FieldPreserves] = complete
// via chSignature, so Evaluate never returns VerdictUnknown and the
// "fully resolved singleton success families" scope from the
// reviewer's correction 7 is satisfied.
func buildDiscriminatingFixture(x, y string, failWithX, failWithY, succWithX, succWithY int) []Family {
	families := make([]Family, 0, failWithX+failWithY+succWithX+succWithY)
	idx := 0
	addFamily := func(outcome domain.OutcomeClass, preserves string) {
		clusterID := "cl_" + string(outcome) + "_" + preserves + "_" + itoa(idx)
		sigID := "sig_" + clusterID
		sig := chSignature(sigID, outcome, preserves)
		families = append(families, chFamily(clusterID, outcome, sig))
		idx++
	}
	for i := 0; i < failWithX; i++ {
		addFamily(domain.OutcomeFailure, x)
	}
	for i := 0; i < failWithY; i++ {
		addFamily(domain.OutcomeFailure, y)
	}
	for i := 0; i < succWithX; i++ {
		addFamily(domain.OutcomePartialSuccess, x)
	}
	for i := 0; i < succWithY; i++ {
		addFamily(domain.OutcomePartialSuccess, y)
	}
	return families
}

// itoa is a tiny integer→ascii helper (avoids importing strconv for
// a two-digit index range).
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [4]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

// TestSuccessPreservingDiscriminatesBySPrev is the controlled
// policy-mechanism test predeclared for (l-coh-mechanism-test). The
// predicate under test is `contains(preserves, residue_locality)`,
// used only as a discrimination target; nothing here depends on the
// canonical id's real-world meaning.
func TestSuccessPreservingDiscriminatesBySPrev(t *testing.T) {
	_ = chVocab(t) // ensure the pinned test vocabulary is registered
	pred := chPredicate(chIDResidue)
	x := chIDResidue
	y := chIDSieve

	t.Run("FixtureA_ZeroPreservingSuccess_discrim=0.5", func(t *testing.T) {
		// fPrev = 5/10 = 0.5; sPrev = 0/10 = 0.0; discrim = +0.5.
		families := buildDiscriminatingFixture(x, y, 5, 5, 0, 10)

		got := VerifySuccessPreserving(pred, families)

		if got.Outcome != OutcomeCompletedNegative {
			t.Errorf("MECH-DISTINGUISH violated for Fixture A: got Outcome %q, want %q; detail=%q",
				got.Outcome, OutcomeCompletedNegative, got.Detail)
		}
		if got.Confirmed {
			t.Errorf("MECH-DISTINGUISH violated for Fixture A: got Confirmed=true, want false")
		}
		if got.Delta != nil {
			t.Errorf("MECH-DISTINGUISH violated for Fixture A: got non-nil Delta %+v, want nil", got.Delta)
		}
		if len(got.Evidence) != 0 {
			t.Errorf("MECH-DISTINGUISH violated for Fixture A: got %d Evidence entries, want 0", len(got.Evidence))
		}
	})

	t.Run("FixtureB_ThreePreservingSuccesses_discrim=0.5", func(t *testing.T) {
		// fPrev = 8/10 = 0.8; sPrev = 3/10 = 0.3; discrim = +0.5.
		families := buildDiscriminatingFixture(x, y, 8, 2, 3, 7)

		got := VerifySuccessPreserving(pred, families)

		if got.Outcome != OutcomeConfirmed {
			t.Errorf("MECH-DISTINGUISH violated for Fixture B: got Outcome %q, want %q; detail=%q",
				got.Outcome, OutcomeConfirmed, got.Detail)
		}
		if !got.Confirmed {
			t.Errorf("MECH-DISTINGUISH violated for Fixture B: got Confirmed=false, want true")
		}
		if got.Delta == nil {
			t.Fatalf("MECH-DISTINGUISH violated for Fixture B: got nil Delta, want non-nil DeltaContrastCollapse")
		}
		if got.Delta.Kind != DeltaContrastCollapse {
			t.Errorf("MECH-DISTINGUISH violated for Fixture B: got Delta.Kind %q, want %q",
				got.Delta.Kind, DeltaContrastCollapse)
		}
		if len(got.Evidence) < 3 {
			t.Errorf("MECH-DISTINGUISH violated for Fixture B: got %d Evidence entries, want >= 3", len(got.Evidence))
		}
		for _, ev := range got.Evidence {
			if ev.Kind != EvidenceSuccessFamily {
				t.Errorf("MECH-DISTINGUISH violated for Fixture B: got Evidence.Kind %q, want %q",
					ev.Kind, EvidenceSuccessFamily)
			}
		}
	})
}
