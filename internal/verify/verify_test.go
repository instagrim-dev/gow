package verify

import (
	"context"
	"testing"

	"github.com/instagrim-dev/newf/internal/invariant"
)

// alwaysVerifier is a test model tier returning a fixed decisive verdict.
type alwaysVerifier struct {
	verdict Verdict
	cost    int
}

func (alwaysVerifier) Kind() VerifierKind { return KindModelJudgment }
func (a alwaysVerifier) Cost() int        { return a.cost }
func (a alwaysVerifier) Verify(_ context.Context, _ VerificationContext) (Decision, error) {
	return Decision{Verdict: a.verdict, Kind: KindModelJudgment, Strength: StrengthSingleModelJudgment, ConfidenceOrdinal: "medium"}, nil
}

func TestDeterministicCheckFailsWhenStillPreserved(t *testing.T) {
	// A target the proposal claims to break is still SATISFIED -> deterministic
	// failure (the break did not happen).
	vc := VerificationContext{
		TargetVerdicts: map[string]invariant.Verdict{"inv_1": invariant.VerdictSatisfies},
	}
	d, err := DeterministicCheck{}.Verify(context.Background(), vc)
	if err != nil {
		t.Fatal(err)
	}
	if d.Verdict != VerdictFailure || d.Strength != StrengthDeterministic {
		t.Fatalf("got %q/%q, want failure/deterministic", d.Verdict, d.Strength)
	}
}

func TestDeterministicCheckDefersOnConfirmedBreak(t *testing.T) {
	vc := VerificationContext{
		TargetVerdicts: map[string]invariant.Verdict{"inv_1": invariant.VerdictViolates},
	}
	d, _ := DeterministicCheck{}.Verify(context.Background(), vc)
	if d.Verdict.Decisive() {
		t.Fatalf("a confirmed break must be non-decisive at the deterministic tier; got %q", d.Verdict)
	}
}

func TestCounterexampleSearchSharedBreakIsNonDecisive(t *testing.T) {
	// H3: a known failure family that ALSO violates the target merely shares a
	// PREDICATE bit with the proposal. A shared predicate verdict is not a
	// mechanism-level refutation (two different constructions can both violate the
	// same predicate; one failing does not make the other fail). This branch must
	// therefore be NON-DECISIVE — at most a "break previously observed" novelty
	// signal — never a decisive proposal failure.
	vc := VerificationContext{
		TargetVerdicts:  map[string]invariant.Verdict{"inv_1": invariant.VerdictViolates},
		NearestVerdicts: map[string][]invariant.Verdict{"inv_1": {invariant.VerdictViolates}},
	}
	d, _ := CounterexampleSearch{}.Verify(context.Background(), vc)
	if d.Verdict != VerdictUnknown || d.Kind != KindCounterexampleSearch {
		t.Fatalf("got %q/%q, want unknown/counterexample-search", d.Verdict, d.Kind)
	}
	if d.Verdict.Decisive() {
		t.Fatalf("a shared predicate verdict must not decide a proposal failure; got decisive %q", d.Verdict)
	}
}

func TestCounterexampleSearchDoesNotRefuteIntendedStructuralDifference(t *testing.T) {
	// G1 (the reversed test, corrected): the proposal VIOLATES a target that an
	// old failure family SATISFIES. That is exactly the structural difference
	// frontier generation is trying to produce — NOT a refutation. The search
	// finds no refuter and must return a NON-DECISIVE bounded-search result, never
	// `failure` and never a free `partial_success`.
	vc := VerificationContext{
		TargetVerdicts:  map[string]invariant.Verdict{"inv_1": invariant.VerdictViolates},
		NearestVerdicts: map[string][]invariant.Verdict{"inv_1": {invariant.VerdictSatisfies}},
	}
	d, _ := CounterexampleSearch{}.Verify(context.Background(), vc)
	if d.Verdict.Decisive() {
		t.Fatalf("an old family preserving a property must not refute a new mechanism that breaks it; got decisive %q", d.Verdict)
	}
	if d.Verdict != VerdictUnknown {
		t.Fatalf("got %q, want unknown (bounded-search negative)", d.Verdict)
	}
}

func TestCounterexampleSearchNoComparisonEvidenceIsNotProgress(t *testing.T) {
	// G1: a confirmed break with NO comparison evidence (or only unknown verdicts)
	// establishes neither refutation nor realizability. The bounded search must
	// NOT reward missing evidence with partial_success; it stays non-decisive so
	// the model tier judges realizability.
	for name, nearest := range map[string]map[string][]invariant.Verdict{
		"no evidence":   {},
		"unknown only":  {"inv_1": {invariant.VerdictUnknown}},
		"empty for tgt": {"inv_1": {}},
	} {
		vc := VerificationContext{
			TargetVerdicts:  map[string]invariant.Verdict{"inv_1": invariant.VerdictViolates},
			NearestVerdicts: nearest,
		}
		d, _ := CounterexampleSearch{}.Verify(context.Background(), vc)
		if d.Verdict != VerdictUnknown {
			t.Fatalf("%s: got %q, want unknown (missing/unknown comparison evidence is not progress)", name, d.Verdict)
		}
	}
}

func TestRouteDeterministicFailureOverridesModelSuccess(t *testing.T) {
	// R3: a deterministic failure must override a confident model "success", even
	// when the model tier declares a cheaper cost.
	vc := VerificationContext{
		TargetVerdicts: map[string]invariant.Verdict{"inv_1": invariant.VerdictSatisfies},
	}
	verifiers := []Verifier{
		alwaysVerifier{verdict: VerdictSuccess, cost: 5},
		DeterministicCheck{},
	}
	d, err := Route(context.Background(), verifiers, vc)
	if err != nil {
		t.Fatal(err)
	}
	if d.Verdict != VerdictFailure || d.Kind != KindDeterministicCheck || d.Strength != StrengthDeterministic {
		t.Fatalf("got %q/%q/%q; a deterministic failure must override a model success", d.Verdict, d.Kind, d.Strength)
	}
}

func TestRouteFallsThroughNonDecisiveCounterexampleToModelTier(t *testing.T) {
	// H3: the counterexample-search tier is non-decisive on a shared predicate
	// break, so the router must fall through to the model tier for the
	// realizability judgment rather than laundering a deterministic `failure` out
	// of a shared predicate bit. The model tier's verdict/kind is what survives.
	vc := VerificationContext{
		TargetVerdicts:  map[string]invariant.Verdict{"inv_1": invariant.VerdictViolates},
		NearestVerdicts: map[string][]invariant.Verdict{"inv_1": {invariant.VerdictViolates}},
	}
	verifiers := []Verifier{DeterministicCheck{}, CounterexampleSearch{}, alwaysVerifier{verdict: VerdictSuccess, cost: 99}}
	d, _ := Route(context.Background(), verifiers, vc)
	if d.Verdict != VerdictSuccess || d.Kind != KindModelJudgment {
		t.Fatalf("got %q/%q, want success/model-judgment (counterexample search must not decide a shared-break failure)", d.Verdict, d.Kind)
	}
}

// overreportingVerifier is a model tier that returns a VALID but too-strong
// strength, to exercise the router's ceiling clamp (G5).
type overreportingVerifier struct{ verdict Verdict }

func (overreportingVerifier) Kind() VerifierKind { return KindModelJudgment }
func (overreportingVerifier) Cost() int          { return 1 }
func (o overreportingVerifier) Verify(_ context.Context, _ VerificationContext) (Decision, error) {
	// A model-judgment verifier claiming deterministic strength: the router must
	// clamp it to the verifier's registered tier, never store the laundered value.
	return Decision{Verdict: o.verdict, Kind: KindModelJudgment, Strength: StrengthDeterministic}, nil
}

func TestRouteClampsStrengthToRegisteredCeiling(t *testing.T) {
	// G5: a model-kind verifier returning the VALID enum `deterministic` must be
	// recorded as single-model-judgment; a verifier's self-reported strength can
	// never exceed its registered tier.
	vc := VerificationContext{}
	d, err := Route(context.Background(), []Verifier{overreportingVerifier{verdict: VerdictSuccess}}, vc)
	if err != nil {
		t.Fatal(err)
	}
	if d.Kind != KindModelJudgment || d.Strength != StrengthSingleModelJudgment {
		t.Fatalf("got %q/%q, want model-judgment/single-model-judgment (strength must not exceed registered ceiling)", d.Kind, d.Strength)
	}
}

func TestRouteFallsToModelTier(t *testing.T) {
	// No deterministic verdicts -> deterministic tiers abstain -> model tier
	// decides -> single-model-judgment.
	vc := VerificationContext{}
	verifiers := []Verifier{DeterministicCheck{}, alwaysVerifier{verdict: VerdictSuccess, cost: 50}}
	d, _ := Route(context.Background(), verifiers, vc)
	if d.Verdict != VerdictSuccess || d.Strength != StrengthSingleModelJudgment {
		t.Fatalf("got %q/%q, want success/single-model-judgment", d.Verdict, d.Strength)
	}
}

func TestRouteNothingDecidesIsBlocked(t *testing.T) {
	vc := VerificationContext{}
	verifiers := []Verifier{DeterministicCheck{}, CounterexampleSearch{}}
	d, _ := Route(context.Background(), verifiers, vc)
	if d.Verdict != VerdictVerificationBlocked {
		t.Fatalf("got %q, want verification_blocked", d.Verdict)
	}
}
