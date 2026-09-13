package shape

import (
	"testing"

	"github.com/instagrim-dev/newf/internal/finite"
	"github.com/instagrim-dev/newf/internal/rewrite"
)

// The D6 cheapest-falsification A/B (docs/plans/2026-09-12-017 §6),
// run through the REAL bounded search: under expansion scarcity, an
// informative history must make completion cheaper than the neutral
// order, and a misleading history must make it more expensive. If v0
// cannot show both signs on development episodes, the mechanism is dead
// before anything sealed. Development material only.

func v(n string) finite.Expr { return finite.Var{Name: n} }
func nn(e finite.Expr) finite.Expr {
	return finite.Unary{Op: finite.OpNot, X: finite.Unary{Op: finite.OpNot, X: e}}
}
func addZ(e finite.Expr) finite.Expr {
	return finite.Binary{Op: finite.OpAdd, X: e, Y: finite.Const{Value: 0}}
}

// admitAB admits the A/B rule set: three shrinking rules plus one true
// but explosive distractor (not-intro grows every subexpression), all
// through the real warrant boundary.
func admitAB(t *testing.T) map[string]rewrite.Rule {
	t.Helper()
	d1 := finite.Domain{Width: 4, Vars: []string{"a"}}
	adm := func(name string, lhs, rhs finite.Expr) rewrite.Rule {
		t.Helper()
		cert := finite.AssessEquivalence(finite.Binding{Sentence: name, Domain: d1}, lhs, rhs)
		rule, defects := rewrite.AdmitRule(name, cert, lhs, rhs, d1)
		if len(defects) > 0 {
			t.Fatalf("rule %q refused: %v", name, defects)
		}
		return rule
	}
	return map[string]rewrite.Rule{
		"double-not": adm("double-not", nn(v("a")), v("a")),
		"add-zero":   adm("add-zero", addZ(v("a")), v("a")),
		"not-intro":  adm("not-intro", v("a"), nn(v("a"))), // true, explosive
	}
}

func ordered(rules map[string]rewrite.Rule, names []string) []rewrite.Rule {
	out := make([]rewrite.Rule, 0, len(names))
	for _, n := range names {
		out = append(out, rules[n])
	}
	return out
}

// minBudget finds the smallest expansion budget that completes the task
// (BestCost <= target with a verified endpoint) under the given rule
// order. Completion is monotone in budget (more expansions only add
// exploration), so binary search is sound. Returns cap+1 when even cap
// fails.
func minBudget(t *testing.T, start finite.Expr, d finite.Domain, rules []rewrite.Rule, target int64, cap int) int {
	t.Helper()
	completes := func(budget int) bool {
		res, err := rewrite.Search(start, d, rules, rewrite.NodeCount, budget)
		if err != nil {
			t.Fatalf("search: %v", err)
		}
		return res.BestCost <= target && res.EndpointVerified
	}
	if !completes(cap) {
		return cap + 1
	}
	lo, hi := 0, cap // completes(hi) true
	for lo < hi {
		mid := (lo + hi) / 2
		if completes(mid) {
			hi = mid
		} else {
			lo = mid + 1
		}
	}
	return hi
}

func TestInformativeHistoryHelpsUnderScarcity(t *testing.T) {
	rules := admitAB(t)
	d := finite.Domain{Width: 4, Vars: []string{"x"}}
	task := nn(addZ(nn(v("x")))) // needs double-not, add-zero, double-not
	const target = 1

	// Neutral catalog order front-loads the explosive distractor.
	neutral := []string{"not-intro", "double-not", "add-zero"}

	// Informative history: relevant completed attempts used the
	// shrinking rules; the distractor appears only in a relevant failure.
	dec := Select(Input{
		TaskStart: RenderExpr(task),
		Target:    target,
		Catalog:   neutral,
		History: History{
			{Start: RenderExpr(nn(addZ(nn(v("y"))))), RulesApplied: []string{"double-not", "add-zero"}, Completed: true},
			{Start: RenderExpr(nn(addZ(v("y")))), RulesApplied: []string{"not-intro"}, Completed: false},
		},
	})
	if dec.EnabledRules[0] == "not-intro" {
		t.Fatalf("informative history should demote the distractor: %v", dec.EnabledRules)
	}

	const budgetCap = 128
	neutralMin := minBudget(t, task, d, ordered(rules, neutral), target, budgetCap)
	shapedMin := minBudget(t, task, d, ordered(rules, dec.EnabledRules), target, budgetCap)
	if shapedMin >= neutralMin {
		t.Fatalf("informative shaping must make completion cheaper under scarcity: shaped needs %d expansions, neutral needs %d", shapedMin, neutralMin)
	}
	t.Logf("informative: shaped min budget %d < neutral %d", shapedMin, neutralMin)
}

func TestMisleadingHistoryHurtsUnderScarcity(t *testing.T) {
	rules := admitAB(t)
	d := finite.Domain{Width: 4, Vars: []string{"x"}}
	task := nn(addZ(nn(v("x"))))
	const target = 1

	// Neutral catalog order is GOOD here (shrinking rules first).
	neutral := []string{"double-not", "add-zero", "not-intro"}

	// Misleading history: relevant "successes" used the distractor;
	// the shrinking rules appear only in relevant failures.
	dec := Select(Input{
		TaskStart: RenderExpr(task),
		Target:    target,
		Catalog:   neutral,
		History: History{
			{Start: RenderExpr(nn(addZ(nn(v("y"))))), RulesApplied: []string{"not-intro"}, Completed: true},
			{Start: RenderExpr(nn(addZ(nn(v("z"))))), RulesApplied: []string{"double-not", "add-zero"}, Completed: false},
		},
	})
	if dec.EnabledRules[0] != "not-intro" {
		t.Fatalf("misleading history should promote the distractor: %v", dec.EnabledRules)
	}

	const budgetCap = 128
	neutralMin := minBudget(t, task, d, ordered(rules, neutral), target, budgetCap)
	misledMin := minBudget(t, task, d, ordered(rules, dec.EnabledRules), target, budgetCap)
	if misledMin <= neutralMin {
		t.Fatalf("misleading shaping must make completion more expensive: misled needs %d expansions, neutral needs %d — a selector that cannot be hurt is not consuming history", misledMin, neutralMin)
	}
	t.Logf("misleading: misled min budget %d > neutral %d", misledMin, neutralMin)
}
