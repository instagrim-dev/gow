// Regression tests for MaxStates refusal semantics (2026-09-13
// adversarial review): the cap stops a search only when an unseen
// successor is actually refused admission because the visited set is
// full. A reachable set that fits the cap exactly completes with
// StateBounded=false.
package rewrite

import (
	"testing"

	"github.com/instagrim-dev/newf/internal/finite"
)

// exactFitFixture is the double-not rule not(not(a)) -> a searched from
// not(not(x)). The reachable set is exactly {not(not(x)), x}, so the
// deterministic visited-set size is N = 2: the rule matches only at the
// root of not(not(x)), yields x, and x admits no further rewrites.
func exactFitFixture(t *testing.T) (finite.Expr, finite.Domain, []Rule) {
	t.Helper()
	d4a := dom(4, "a")
	lhs, rhs := doubleNot(d4a)
	r := admit(t, "double-not", lhs, rhs, d4a)
	start := finite.Unary{Op: finite.OpNot, X: finite.Unary{Op: finite.OpNot, X: finite.Var{Name: "x"}}}
	return start, dom(4, "x"), []Rule{r}
}

func TestReachableSetExactlyAtStateCapCompletesUnbounded(t *testing.T) {
	start, d, rules := exactFitFixture(t)

	// Establish N with a generous cap first: the search must exhaust the
	// space naturally, visiting exactly the two reachable states.
	free, err := SearchBounded(start, d, rules, nil, 10, Limits{MaxStates: 100})
	if err != nil {
		t.Fatal(err)
	}
	if free.StateBounded || free.BudgetExhausted || free.Explored != 2 {
		t.Fatalf("setup: the space must close naturally at two states: %+v", free)
	}

	// The reachable set fits MaxStates exactly: nothing was refused, so
	// the run is a completion, not a resource stop.
	res, err := SearchBounded(start, d, rules, nil, 10, Limits{MaxStates: 2})
	if err != nil {
		t.Fatal(err)
	}
	if res.StateBounded {
		t.Fatalf("an exact-fit reachable set refused no successor and must not report StateBounded: %+v", res)
	}
	if res.BudgetExhausted || res.Explored != 2 {
		t.Fatalf("the exact-fit search must complete by natural exhaustion under the budget: %+v", res)
	}
	if res.Best != "x" || res.BestCost != 1 || !res.EndpointVerified {
		t.Fatalf("the exact-fit search must retain the verified best-found endpoint: %+v", res)
	}
}

func TestStateCapBelowReachableSetSetsStateBounded(t *testing.T) {
	start, d, rules := exactFitFixture(t)

	// One state fewer than the reachable set: the unseen successor x is
	// refused admission, and only that refusal reports StateBounded.
	res, err := SearchBounded(start, d, rules, nil, 10, Limits{MaxStates: 1})
	if err != nil {
		t.Fatal(err)
	}
	if !res.StateBounded {
		t.Fatalf("refusing an unseen successor must report StateBounded: %+v", res)
	}
	if res.Best != res.Original {
		t.Fatalf("the refused successor must not have been admitted as best: %+v", res)
	}
}
