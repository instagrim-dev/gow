// Package rewrite is the roadmap's "tiny bounded reference
// rewriter/enumerator" (revision 0.3.0 §4 G4-lite): a deterministic,
// budget-bounded search over equivalent finite expressions using only
// warrant-admitted equality rules. It exists so the development loop can
// actually search — and so G4-lite sensitivity calibration has a reference
// procedure — without waiting on an equality-saturation engine.
//
// It is deliberately NOT an e-graph and must not grow into one (the
// roadmap's first tranche explicitly excludes a custom e-graph
// implementation). It is plain breadth-first term rewriting with a visited
// set and an expansion budget.
//
// T0 discipline:
//
//   - A Rule can only be constructed through AdmitRule, which requires a
//     defect-free finite.VerifyRuleWarrant (structural checks plus
//     independent replay). Rule fields are unexported: a rule without a
//     warrant is unrepresentable here, not merely discouraged.
//   - A rule is scoped to its admitted width. Search refuses rules whose
//     width differs from the search domain's: a width-4 certificate
//     warrants nothing at width 8.
//   - A budget stop yields best-found, never a claim of optimality or
//     saturation, and the result says which.
//   - The endpoint is independently replayed: the search's claimed
//     equivalence between the original and the best-found expression is
//     re-decided by the exhaustive oracle (finite.AssessEquivalence),
//     which shares no search state. An oracle refusal (for example a
//     domain above the exhaustiveness cap) leaves the result an
//     UNVERIFIED candidate, stated as such.
package rewrite

import (
	"fmt"
	"sort"

	"github.com/instagrim-dev/newf/internal/finite"
)

// Rule is one admitted, width-scoped equality rewrite. Fields are
// unexported: construction goes through AdmitRule's warrant check only.
type Rule struct {
	name     string
	domain   finite.Domain
	lhs, rhs finite.Expr
}

// Name returns the rule's admission name.
func (r Rule) Name() string { return r.name }

// AdmitRule constructs a Rule from a certificate presented as warrant for
// "lhs == rhs on domain d". It returns the complete defect list when the
// warrant is insufficient (finite.VerifyRuleWarrant's checks plus the
// substitution-safety requirement that every right-hand variable is bound
// by the left-hand side).
func AdmitRule(name string, cert finite.Certificate, lhs, rhs finite.Expr, d finite.Domain) (Rule, []string) {
	defects := finite.VerifyRuleWarrant(cert, lhs, rhs, d)
	if name == "" {
		defects = append(defects, "rule name is empty")
	}
	if len(defects) > 0 {
		return Rule{}, defects
	}
	// Substitution safety: rewriting binds metavariables by matching the
	// LHS; an RHS variable the LHS never binds would have no value.
	lhsVars := freeVars(lhs)
	for _, v := range sortedVars(freeVars(rhs)) {
		if !lhsVars[v] {
			defects = append(defects, fmt.Sprintf("right-hand variable %q is not bound by the left-hand side; the rewrite would be unsubstitutable", v))
		}
	}
	if len(defects) > 0 {
		return Rule{}, defects
	}
	return Rule{name: name, domain: d, lhs: lhs, rhs: rhs}, nil
}

// CostModel assigns a nonnegative cost to an expression. Ties are broken
// by canonical rendering, so search results are deterministic.
type CostModel func(finite.Expr) int64

// NodeCount is the default cost model: one unit per node.
func NodeCount(e finite.Expr) int64 {
	switch t := e.(type) {
	case finite.Var, finite.Const:
		return 1
	case finite.Unary:
		return 1 + NodeCount(t.X)
	case finite.Binary:
		return 1 + NodeCount(t.X) + NodeCount(t.Y)
	}
	return 1
}

// Step is one applied rewrite on the path to the best-found expression.
type Step struct {
	Rule   string
	Before string
	After  string
}

// Result is the search outcome with its honesty flags.
type Result struct {
	Original     string
	Best         string
	OriginalCost int64
	BestCost     int64
	Steps        []Step // rewrite path from Original to Best
	Explored     int    // expressions expanded
	// BudgetExhausted: the expansion budget stopped the search. Best is
	// best-found under the budget — not a claim of optimality and not
	// saturation.
	BudgetExhausted bool
	// Endpoint is the independent oracle replay of Original == Best.
	// EndpointVerified is true only when the oracle exhaustively holds;
	// any refusal leaves the best expression an unverified candidate.
	Endpoint         finite.Certificate
	EndpointVerified bool
}

// Search runs deterministic breadth-first rewriting from start under the
// admitted rules, within maxExpansions. It refuses malformed inputs and
// out-of-scope rules rather than proceeding on unstated premises.
func Search(start finite.Expr, d finite.Domain, rules []Rule, cost CostModel, maxExpansions int) (Result, error) {
	if defects := finite.ValidateExpr(start, d); len(defects) > 0 {
		return Result{}, fmt.Errorf("start expression is not valid in the search domain: %v", defects)
	}
	if maxExpansions < 0 {
		return Result{}, fmt.Errorf("expansion budget must be nonnegative, got %d", maxExpansions)
	}
	if cost == nil {
		cost = NodeCount
	}
	for _, r := range rules {
		if r.name == "" {
			return Result{}, fmt.Errorf("an unadmitted zero-value rule was supplied; rules exist only via AdmitRule")
		}
		if r.domain.Width != d.Width {
			return Result{}, fmt.Errorf("rule %q was admitted at width %d and cannot be applied in a width-%d search; a certificate warrants exactly its domain", r.name, r.domain.Width, d.Width)
		}
	}

	type parentEdge struct {
		parent string
		step   Step
	}
	startKey := finite.Render(start)
	visited := map[string]finite.Expr{startKey: start}
	parents := map[string]parentEdge{}
	queue := []string{startKey}

	res := Result{
		Original:     startKey,
		Best:         startKey,
		OriginalCost: cost(start),
		BestCost:     cost(start),
	}

	expansions := 0
	for len(queue) > 0 {
		if expansions >= maxExpansions {
			res.BudgetExhausted = true
			break
		}
		key := queue[0]
		queue = queue[1:]
		current := visited[key]
		expansions++
		res.Explored++

		for _, r := range rules {
			for _, next := range applyEverywhere(current, r, d) {
				nextKey := finite.Render(next)
				if _, seen := visited[nextKey]; seen {
					continue
				}
				visited[nextKey] = next
				parents[nextKey] = parentEdge{parent: key, step: Step{Rule: r.name, Before: key, After: nextKey}}
				queue = append(queue, nextKey)
				c := cost(next)
				if c < res.BestCost || (c == res.BestCost && nextKey < res.Best) {
					res.Best = nextKey
					res.BestCost = c
				}
			}
		}
	}

	// Reconstruct the path to the best expression.
	if res.Best != res.Original {
		var steps []Step
		for at := res.Best; at != res.Original; {
			edge := parents[at]
			steps = append(steps, edge.step)
			at = edge.parent
		}
		for i, j := 0, len(steps)-1; i < j; i, j = i+1, j-1 {
			steps[i], steps[j] = steps[j], steps[i]
		}
		res.Steps = steps
	}

	// Independent endpoint replay: the oracle re-decides Original == Best
	// with no search state. A refusal is preserved, not upgraded.
	best := visited[res.Best]
	res.Endpoint = finite.AssessEquivalence(finite.Binding{
		Sentence: fmt.Sprintf("search endpoint: %s == %s", res.Original, res.Best),
		Domain:   d,
	}, start, best)
	res.EndpointVerified = res.Endpoint.Verdict == finite.VerdictHoldsOnDomain
	return res, nil
}

// applyEverywhere returns every expression obtained by applying the rule
// at exactly one position of e, in deterministic order. Results have
// their constants normalized to the search width so semantically
// identical states share one visited-set key and one rendering
// (adversarial review finding 10).
func applyEverywhere(e finite.Expr, r Rule, d finite.Domain) []finite.Expr {
	var out []finite.Expr
	if b, ok := match(r.lhs, e, d, map[string]finite.Expr{}); ok {
		out = append(out, normalizeConsts(subst(r.rhs, b), d))
	}
	switch t := e.(type) {
	case finite.Unary:
		for _, v := range applyEverywhere(t.X, r, d) {
			out = append(out, finite.Unary{Op: t.Op, X: v})
		}
	case finite.Binary:
		for _, v := range applyEverywhere(t.X, r, d) {
			out = append(out, finite.Binary{Op: t.Op, X: v, Y: t.Y})
		}
		for _, v := range applyEverywhere(t.Y, r, d) {
			out = append(out, finite.Binary{Op: t.Op, X: t.X, Y: v})
		}
	}
	return out
}

// match binds the pattern's variables (metavariables, quantified by the
// rule's admitted domain) to subexpressions of the subject. A repeated
// metavariable must bind to structurally identical subexpressions.
func match(pattern, subject finite.Expr, d finite.Domain, bindings map[string]finite.Expr) (map[string]finite.Expr, bool) {
	switch p := pattern.(type) {
	case finite.Var:
		if prev, ok := bindings[p.Name]; ok {
			if finite.Render(prev) != finite.Render(subject) {
				return nil, false
			}
			return bindings, true
		}
		bindings[p.Name] = subject
		return bindings, true
	case finite.Const:
		s, ok := subject.(finite.Const)
		if !ok {
			return nil, false
		}
		mask := uint64(1)<<uint(d.Width) - 1
		if p.Value&mask != s.Value&mask {
			return nil, false
		}
		return bindings, true
	case finite.Unary:
		s, ok := subject.(finite.Unary)
		if !ok || s.Op != p.Op {
			return nil, false
		}
		return match(p.X, s.X, d, bindings)
	case finite.Binary:
		s, ok := subject.(finite.Binary)
		if !ok || s.Op != p.Op {
			return nil, false
		}
		b, ok := match(p.X, s.X, d, bindings)
		if !ok {
			return nil, false
		}
		return match(p.Y, s.Y, d, b)
	}
	return nil, false
}

// subst instantiates the rule's right-hand side under the bindings. Every
// variable is bound: AdmitRule rejected rules with unbound RHS variables.
func subst(rhs finite.Expr, bindings map[string]finite.Expr) finite.Expr {
	switch t := rhs.(type) {
	case finite.Var:
		return bindings[t.Name]
	case finite.Const:
		return t
	case finite.Unary:
		return finite.Unary{Op: t.Op, X: subst(t.X, bindings)}
	case finite.Binary:
		return finite.Binary{Op: t.Op, X: subst(t.X, bindings), Y: subst(t.Y, bindings)}
	}
	return rhs
}

// normalizeConsts masks constant values to the search width, matching
// evaluation semantics, so renderings of semantically identical states
// coincide.
func normalizeConsts(e finite.Expr, d finite.Domain) finite.Expr {
	mask := uint64(1)<<uint(d.Width) - 1
	switch t := e.(type) {
	case finite.Const:
		return finite.Const{Value: t.Value & mask}
	case finite.Unary:
		return finite.Unary{Op: t.Op, X: normalizeConsts(t.X, d)}
	case finite.Binary:
		return finite.Binary{Op: t.Op, X: normalizeConsts(t.X, d), Y: normalizeConsts(t.Y, d)}
	}
	return e
}

func freeVars(e finite.Expr) map[string]bool {
	out := map[string]bool{}
	var walk func(finite.Expr)
	walk = func(x finite.Expr) {
		switch t := x.(type) {
		case finite.Var:
			out[t.Name] = true
		case finite.Unary:
			walk(t.X)
		case finite.Binary:
			walk(t.X)
			walk(t.Y)
		}
	}
	walk(e)
	return out
}

func sortedVars(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for v := range m {
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}
