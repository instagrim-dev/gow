// Development tests for the bounded reference rewriter. Everything here
// is synthetic development material.
package rewrite

import (
	"strings"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/finite"
)

func dom(width int, vars ...string) finite.Domain { return finite.Domain{Width: width, Vars: vars} }

// admit is a test helper that fails the test when admission is refused.
func admit(t *testing.T, name string, lhs, rhs finite.Expr, d finite.Domain) Rule {
	t.Helper()
	cert := finite.AssessEquivalence(finite.Binding{Sentence: name, Domain: d}, lhs, rhs)
	rule, defects := AdmitRule(name, cert, lhs, rhs, d)
	if len(defects) > 0 {
		t.Fatalf("admission of %q refused: %v", name, defects)
	}
	return rule
}

func doubleNot(d finite.Domain) (finite.Expr, finite.Expr) {
	return finite.Unary{Op: finite.OpNot, X: finite.Unary{Op: finite.OpNot, X: finite.Var{Name: "a"}}}, finite.Var{Name: "a"}
}

// The development loop end-to-end: warrant-admitted rules, bounded
// search, a cheaper equivalent found, and the endpoint independently
// replayed by the oracle. Residual (excessive cost) → applicable tool
// (admitted equalities + bounded search) → executed check (exhaustive
// replay) → attributed outcome (rule-named step path).
func TestSearchFindsCheaperEquivalentAndReplaysEndpoint(t *testing.T) {
	d4a := dom(4, "a")
	dnLHS, dnRHS := doubleNot(d4a)
	rules := []Rule{
		admit(t, "double-not", dnLHS, dnRHS, d4a),
		admit(t, "xor-self-zero",
			finite.Binary{Op: finite.OpXor, X: finite.Var{Name: "a"}, Y: finite.Var{Name: "a"}},
			finite.Const{Value: 0}, d4a),
	}

	searchDomain := dom(4, "x")
	nnx := finite.Unary{Op: finite.OpNot, X: finite.Unary{Op: finite.OpNot, X: finite.Var{Name: "x"}}}
	start := finite.Binary{Op: finite.OpXor, X: nnx, Y: nnx}

	res, err := Search(start, searchDomain, rules, NodeCount, 100)
	if err != nil {
		t.Fatal(err)
	}
	if res.Best != "0" || res.BestCost != 1 {
		t.Fatalf("expected the constant 0 (cost 1), got %q (cost %d)", res.Best, res.BestCost)
	}
	if res.OriginalCost != 7 {
		t.Fatalf("original cost should be 7 nodes, got %d", res.OriginalCost)
	}
	if len(res.Steps) == 0 {
		t.Fatal("the outcome must be attributed: a named rule path is required")
	}
	// Chain integrity (adversarial review finding 8): the steps form an
	// unbroken path from Original to Best.
	if res.Steps[0].Before != res.Original {
		t.Fatalf("path must start at the original: %q vs %q", res.Steps[0].Before, res.Original)
	}
	for i := 0; i+1 < len(res.Steps); i++ {
		if res.Steps[i].After != res.Steps[i+1].Before {
			t.Fatalf("broken chain at step %d: %q -> %q", i, res.Steps[i].After, res.Steps[i+1].Before)
		}
	}
	if res.Steps[len(res.Steps)-1].After != res.Best {
		t.Fatalf("path must end at the best: %q vs %q", res.Steps[len(res.Steps)-1].After, res.Best)
	}
	for _, s := range res.Steps {
		if s.Rule == "" {
			t.Fatalf("unattributed step: %+v", s)
		}
	}
	if !res.EndpointVerified || res.Endpoint.Verdict != finite.VerdictHoldsOnDomain {
		t.Fatalf("the endpoint must be independently replayed and hold: %s (%s)", res.Endpoint.Verdict, res.Endpoint.Reason)
	}
	if res.BudgetExhausted {
		t.Fatal("this small space should close under budget")
	}
}

// A false rule cannot become a Rule: admission requires a defect-free
// warrant, and the shift round-trip's certificate is REFUTED.
func TestFalseRuleCannotBeAdmitted(t *testing.T) {
	d := dom(4, "a")
	lhs := finite.Unary{Op: finite.OpShr, X: finite.Unary{Op: finite.OpShl, X: finite.Var{Name: "a"}}}
	rhs := finite.Var{Name: "a"}
	cert := finite.AssessEquivalence(finite.Binding{Sentence: "shift round-trip", Domain: d}, lhs, rhs)
	if _, defects := AdmitRule("shift-round-trip", cert, lhs, rhs, d); len(defects) == 0 {
		t.Fatal("a refuted identity must not be admissible as a rewrite rule")
	}
}

// A true identity whose RHS uses a variable the LHS never binds is
// unsubstitutable and refused at admission, even with a valid warrant:
// mul(a,0) == mul(b,0) holds everywhere, but rewriting with it would
// have no value for b.
func TestUnboundRightVariableRefused(t *testing.T) {
	d := dom(4, "a", "b")
	lhs := finite.Binary{Op: finite.OpMul, X: finite.Var{Name: "a"}, Y: finite.Const{Value: 0}}
	rhs := finite.Binary{Op: finite.OpMul, X: finite.Var{Name: "b"}, Y: finite.Const{Value: 0}}
	cert := finite.AssessEquivalence(finite.Binding{Sentence: "both sides are zero", Domain: d}, lhs, rhs)
	if cert.Verdict != finite.VerdictHoldsOnDomain {
		t.Fatalf("setup: identity should hold, got %s", cert.Verdict)
	}
	_, defects := AdmitRule("zero-swap", cert, lhs, rhs, d)
	if len(defects) == 0 {
		t.Fatal("an unbound right-hand variable must refuse admission")
	}
	if !strings.Contains(strings.Join(defects, " | "), `"b"`) {
		t.Fatalf("the unbound variable must be named: %v", defects)
	}
}

// A rule admitted at one width cannot be applied in a search at another:
// the certificate warrants exactly its domain.
func TestWidthScopeEnforcedAtSearch(t *testing.T) {
	d8 := dom(8, "a")
	lhs, rhs := doubleNot(d8)
	rule := admit(t, "double-not-w8", lhs, rhs, d8)
	if _, err := Search(finite.Var{Name: "x"}, dom(4, "x"), []Rule{rule}, NodeCount, 10); err == nil {
		t.Fatal("a width-8 rule must be refused in a width-4 search")
	} else if !strings.Contains(err.Error(), "width") {
		t.Fatalf("refusal must name the width scope: %v", err)
	}
}

// A zero-value Rule (never admitted) is refused: rules exist only via
// AdmitRule.
func TestUnadmittedRuleRefused(t *testing.T) {
	if _, err := Search(finite.Var{Name: "x"}, dom(4, "x"), []Rule{{}}, NodeCount, 10); err == nil {
		t.Fatal("a zero-value rule must be refused")
	}
}

// Budget exhaustion is reported as best-found, never as optimality; the
// endpoint replay still runs on whatever was found.
func TestBudgetStopIsBestFoundNotOptimal(t *testing.T) {
	d4a := dom(4, "a")
	lhs, rhs := doubleNot(d4a)
	rule := admit(t, "double-not", lhs, rhs, d4a)

	searchDomain := dom(4, "x")
	nnx := finite.Unary{Op: finite.OpNot, X: finite.Unary{Op: finite.OpNot, X: finite.Var{Name: "x"}}}
	start := finite.Binary{Op: finite.OpXor, X: nnx, Y: nnx}

	res, err := Search(start, searchDomain, []Rule{rule}, NodeCount, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !res.BudgetExhausted {
		t.Fatal("a zero budget with pending work must report exhaustion")
	}
	if res.Best != res.Original {
		t.Fatalf("nothing was explored; best must be the original, got %q", res.Best)
	}
	if !res.EndpointVerified {
		t.Fatalf("original == original must still verify: %s", res.Endpoint.Reason)
	}
}

// When the oracle cannot exhaust the domain, the result is an UNVERIFIED
// candidate — stated, not upgraded.
func TestOversizedDomainLeavesEndpointUnverified(t *testing.T) {
	d8 := dom(8, "a")
	lhs, rhs := doubleNot(d8)
	rule := admit(t, "double-not-w8", lhs, rhs, d8)

	big := dom(8, "a", "b", "c") // 256^3 assignments: above the cap
	start := finite.Unary{Op: finite.OpNot, X: finite.Unary{Op: finite.OpNot, X: finite.Binary{Op: finite.OpAdd, X: finite.Var{Name: "a"}, Y: finite.Binary{Op: finite.OpMul, X: finite.Var{Name: "b"}, Y: finite.Var{Name: "c"}}}}}
	res, err := Search(start, big, []Rule{rule}, NodeCount, 5)
	if err != nil {
		t.Fatal(err)
	}
	if res.BestCost >= res.OriginalCost {
		t.Fatalf("the rule should still strip the double negation, got cost %d vs %d", res.BestCost, res.OriginalCost)
	}
	if res.EndpointVerified {
		t.Fatal("an oracle refusal must leave the endpoint unverified")
	}
	if res.Endpoint.Verdict != finite.VerdictUnresolved {
		t.Fatalf("expected the oracle's UNRESOLVED refusal, got %s", res.Endpoint.Verdict)
	}
}

// Search is deterministic: identical inputs yield identical results.
func TestSearchIsDeterministic(t *testing.T) {
	d4a := dom(4, "a")
	dnLHS, dnRHS := doubleNot(d4a)
	rules := []Rule{
		admit(t, "double-not", dnLHS, dnRHS, d4a),
		admit(t, "xor-self-zero",
			finite.Binary{Op: finite.OpXor, X: finite.Var{Name: "a"}, Y: finite.Var{Name: "a"}},
			finite.Const{Value: 0}, d4a),
	}
	searchDomain := dom(4, "x", "y")
	start := finite.Binary{Op: finite.OpXor,
		X: finite.Unary{Op: finite.OpNot, X: finite.Unary{Op: finite.OpNot, X: finite.Binary{Op: finite.OpAdd, X: finite.Var{Name: "x"}, Y: finite.Var{Name: "y"}}}},
		Y: finite.Binary{Op: finite.OpAdd, X: finite.Var{Name: "x"}, Y: finite.Var{Name: "y"}},
	}
	a, err := Search(start, searchDomain, rules, NodeCount, 200)
	if err != nil {
		t.Fatal(err)
	}
	b, err := Search(start, searchDomain, rules, NodeCount, 200)
	if err != nil {
		t.Fatal(err)
	}
	if a.Best != b.Best || a.BestCost != b.BestCost || a.Explored != b.Explored || len(a.Steps) != len(b.Steps) {
		t.Fatalf("nondeterministic search: %+v vs %+v", a, b)
	}
	if a.Best != "0" {
		t.Fatalf("xor of identical subexpressions should collapse to 0, got %q", a.Best)
	}
}

// 2026-09-13 external review finding 5: an expansion budget alone does
// not bound generated state — one counted expansion can materialize many
// successors, and a growth rule (not-intro) admits unbounded terms.
// SearchBounded's ceilings must convert resource exhaustion into a
// bounded result, never a semantic claim.
func TestStateCeilingStopsGrowthAsBoundedResult(t *testing.T) {
	d4a := dom(4, "a")
	notIntro := admit(t, "not-intro",
		finite.Var{Name: "a"},
		finite.Unary{Op: finite.OpNot, X: finite.Unary{Op: finite.OpNot, X: finite.Var{Name: "a"}}}, d4a)
	start := finite.Var{Name: "x"}
	res, err := SearchBounded(start, dom(4, "x"), []Rule{notIntro}, NodeCount, 1_000_000, Limits{MaxStates: 8})
	if err != nil {
		t.Fatal(err)
	}
	if !res.StateBounded {
		t.Fatalf("the state ceiling must be reported as the stop reason: %+v", res)
	}
	if res.Best != "x" || res.BestCost != 1 {
		t.Fatalf("best-found must survive a resource stop: %+v", res)
	}
	if !res.EndpointVerified {
		t.Fatal("the trivial endpoint x == x must still replay")
	}
}

func TestTermSizeCeilingTruncatesAndSaysSo(t *testing.T) {
	d4a := dom(4, "a")
	notIntro := admit(t, "not-intro",
		finite.Var{Name: "a"},
		finite.Unary{Op: finite.OpNot, X: finite.Unary{Op: finite.OpNot, X: finite.Var{Name: "a"}}}, d4a)
	start := finite.Var{Name: "x"}
	// Tree sizes: x is 1 node, not(not(x)) is 3, the next doubling is 5.
	// A 3-node ceiling admits the first successor and refuses the next.
	res, err := SearchBounded(start, dom(4, "x"), []Rule{notIntro}, NodeCount, 100, Limits{MaxTermNodes: 3})
	if err != nil {
		t.Fatal(err)
	}
	if !res.TermSizeBounded {
		t.Fatalf("a skipped oversize successor must be flagged: %+v", res)
	}
	if res.StateBounded || res.Cancelled {
		t.Fatalf("only the term-size bound fired here: %+v", res)
	}
}

// The size ceiling must be paid BEFORE rendering, not after (2026-09-13
// self-review): a bound that renders every candidate and measures the
// string has already spent the cost it exists to refuse. A start
// expression with shared subexpressions makes the difference measurable —
// under a render-then-measure bound this search rendered 16k multi-hundred-KB
// terms and kept none.
func TestOversizeSuccessorsAreRefusedWithoutRendering(t *testing.T) {
	d4a := dom(4, "a")
	addZero := admit(t, "add-zero",
		finite.Binary{Op: finite.OpAdd, X: finite.Var{Name: "a"}, Y: finite.Const{Value: 0}},
		finite.Var{Name: "a"}, d4a)
	var e finite.Expr = finite.Binary{Op: finite.OpAdd, X: finite.Var{Name: "x"}, Y: finite.Const{Value: 0}}
	for i := 0; i < 10; i++ { // shared doubling: ~3k nodes, valid for admission
		e = finite.Binary{Op: finite.OpAdd, X: e, Y: e}
	}
	if defects := finite.ValidateExpr(e, dom(4, "x")); len(defects) > 0 {
		t.Fatalf("the start must be admissible for this test to say anything: %v", defects)
	}
	done := make(chan Result, 1)
	go func() {
		// A ceiling below the start's own size refuses every successor.
		res, err := SearchBounded(e, dom(4, "x"), []Rule{addZero}, NodeCount, 500, Limits{MaxTermNodes: 8})
		if err != nil {
			t.Error(err)
			close(done)
			return
		}
		done <- res
	}()
	select {
	case res, ok := <-done:
		if !ok {
			t.Fatal("search failed")
		}
		if !res.TermSizeBounded {
			t.Fatalf("every successor exceeded the ceiling and must be reported as truncation: %+v", res)
		}
		if res.Best != res.Original {
			t.Fatalf("no successor was retained, so the best must remain the original: %+v", res)
		}
	case <-time.After(20 * time.Second):
		t.Fatal("the size bound is being paid after rendering; refusal must not cost what it refuses")
	}
}

func TestCancellationStopsSearchAsBoundedResult(t *testing.T) {
	d4a := dom(4, "a")
	notIntro := admit(t, "not-intro",
		finite.Var{Name: "a"},
		finite.Unary{Op: finite.OpNot, X: finite.Unary{Op: finite.OpNot, X: finite.Var{Name: "a"}}}, d4a)
	cancel := make(chan struct{})
	close(cancel) // already cancelled: the search must stop immediately
	res, err := SearchBounded(finite.Var{Name: "x"}, dom(4, "x"), []Rule{notIntro}, NodeCount, 1_000_000, Limits{Cancel: cancel})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Cancelled {
		t.Fatalf("cancellation must be reported: %+v", res)
	}
	if res.Explored != 0 {
		t.Fatalf("a pre-cancelled search performs no expansions, got %d", res.Explored)
	}
}

// The metered probe reports the same reduces verdict as the boolean
// convenience, plus the candidates it materialized — the work a caller
// must charge (finding 1).
func TestProbeStrictReductionMetersWork(t *testing.T) {
	d4a := dom(4, "a")
	dnLHS, dnRHS := doubleNot(d4a)
	dn := admit(t, "double-not", dnLHS, dnRHS, d4a)
	e := finite.Unary{Op: finite.OpNot, X: finite.Unary{Op: finite.OpNot, X: finite.Var{Name: "x"}}}
	reduces, candidates := ProbeStrictReduction(e, dn, dom(4, "x"), NodeCount)
	if !reduces {
		t.Fatal("double-not reduces not(not(x))")
	}
	if candidates < 1 {
		t.Fatalf("the probe materialized at least one candidate, meter says %d", candidates)
	}
	if got := CanStrictlyReduce(e, dn, dom(4, "x"), NodeCount); got != reduces {
		t.Fatalf("convenience and metered probe disagree: %v vs %v", got, reduces)
	}
}

// Rule identity binds content, not just the name (finding 3).
func TestRuleIdentityBindsContent(t *testing.T) {
	d4a := dom(4, "a")
	dnLHS, dnRHS := doubleNot(d4a)
	a := admit(t, "same-name", dnLHS, dnRHS, d4a)
	b := admit(t, "same-name",
		finite.Binary{Op: finite.OpAdd, X: finite.Var{Name: "a"}, Y: finite.Const{Value: 0}},
		finite.Var{Name: "a"}, d4a)
	if a.Identity() == b.Identity() {
		t.Fatal("two different rewrites under one name must have distinct identities")
	}
	if a.Identity() != admit(t, "same-name", dnLHS, dnRHS, d4a).Identity() {
		t.Fatal("identical rules must share an identity")
	}
}

// Search still counts generated candidates for the cost ledger.
func TestSearchCountsGeneratedCandidates(t *testing.T) {
	d4a := dom(4, "a")
	dnLHS, dnRHS := doubleNot(d4a)
	dn := admit(t, "double-not", dnLHS, dnRHS, d4a)
	e := finite.Unary{Op: finite.OpNot, X: finite.Unary{Op: finite.OpNot, X: finite.Var{Name: "x"}}}
	res, err := Search(e, dom(4, "x"), []Rule{dn}, NodeCount, 10)
	if err != nil {
		t.Fatal(err)
	}
	if res.Generated < 1 {
		t.Fatalf("the search materialized candidates; the meter must not report %d", res.Generated)
	}
}
