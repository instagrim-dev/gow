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

func TestCounterexampleSearchRefutesBreak(t *testing.T) {
	vc := VerificationContext{
		TargetVerdicts:  map[string]invariant.Verdict{"inv_1": invariant.VerdictViolates},
		NearestVerdicts: map[string][]invariant.Verdict{"inv_1": {invariant.VerdictSatisfies}},
	}
	d, _ := CounterexampleSearch{}.Verify(context.Background(), vc)
	if d.Verdict != VerdictFailure || d.Kind != KindCounterexampleSearch {
		t.Fatalf("got %q/%q, want failure/counterexample-search", d.Verdict, d.Kind)
	}
}

func TestCounterexampleSearchPartialSuccessWhenNoRefuter(t *testing.T) {
	vc := VerificationContext{
		TargetVerdicts:  map[string]invariant.Verdict{"inv_1": invariant.VerdictViolates},
		NearestVerdicts: map[string][]invariant.Verdict{"inv_1": {invariant.VerdictViolates}},
	}
	d, _ := CounterexampleSearch{}.Verify(context.Background(), vc)
	if d.Verdict != VerdictPartialSuccess || d.Strength != StrengthReproducible {
		t.Fatalf("got %q/%q, want partial_success/reproducible", d.Verdict, d.Strength)
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

func TestRouteCheapestFirstStrongestDecisive(t *testing.T) {
	vc := VerificationContext{
		TargetVerdicts:  map[string]invariant.Verdict{"inv_1": invariant.VerdictViolates},
		NearestVerdicts: map[string][]invariant.Verdict{"inv_1": {invariant.VerdictViolates}},
	}
	verifiers := []Verifier{DeterministicCheck{}, CounterexampleSearch{}, alwaysVerifier{verdict: VerdictSuccess, cost: 99}}
	d, _ := Route(context.Background(), verifiers, vc)
	if d.Verdict != VerdictPartialSuccess || d.Kind != KindCounterexampleSearch {
		t.Fatalf("got %q/%q, want partial_success/counterexample-search", d.Verdict, d.Kind)
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
