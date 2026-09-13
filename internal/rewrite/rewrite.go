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
	// Generated counts candidate rewrites the search materialized
	// (successors produced by rule application, before dedup). It is
	// the same unit of work a selector probe performs per candidate, so
	// callers can charge search and pre-search probe work on one ledger
	// (2026-09-13 external review finding 1). Explored (expansions)
	// remains the budget unit; the two are different measures and must
	// not be conflated.
	Generated int
	// BudgetExhausted: the expansion budget stopped the search. Best is
	// best-found under the budget — not a claim of optimality and not
	// saturation.
	BudgetExhausted bool
	// StateBounded: the generated-state ceiling stopped the search
	// (2026-09-13 external review finding 5: one counted expansion can
	// materialize many successors, so an expansion budget alone does not
	// bound memory). Best is best-found under the ceiling — a resource
	// stop, never a semantic judgment.
	StateBounded bool
	// TermSizeBounded: at least one successor exceeded the rendered-size
	// ceiling and was not enqueued. The reachable set was truncated by a
	// resource bound, not by rule semantics.
	TermSizeBounded bool
	// Cancelled: the caller's cancellation channel closed mid-search.
	// Best is best-found at the stop; nothing stronger is claimed.
	Cancelled bool
	// Endpoint is the independent oracle replay of Original == Best.
	// EndpointVerified is true only when the oracle exhaustively holds;
	// any refusal leaves the best expression an unverified candidate.
	Endpoint         finite.Certificate
	EndpointVerified bool
}

// Limits bounds search work beyond the expansion budget (2026-09-13
// external review finding 5). Zero values take the package defaults; the
// defaults are set far above anything a legitimate pack episode reaches,
// so bounded and unbounded runs coincide on all retained results.
type Limits struct {
	// MaxStates caps the visited-set size (generated, deduplicated
	// states). Exceeding it stops the search with StateBounded.
	MaxStates int
	// MaxTermNodes caps a successor's TREE SIZE, checked before the
	// successor is rendered. Larger successors are skipped and
	// TermSizeBounded is set: growth rules (e.g. not-intro) must not
	// materialize unbounded terms, and the check must not itself cost a
	// full rendering of the term it rejects (self-review of the first
	// repair, which rendered first and measured after).
	MaxTermNodes int
	// Cancel, when non-nil and closed, stops the search with Cancelled.
	Cancel <-chan struct{}
}

// Default work ceilings. MaxTermNodes matches finite.MaxExprNodes: a
// search may not construct a term the domain would refuse to admit, so
// the ceiling is the admission bound rather than an independent number
// that could drift from it.
const (
	DefaultMaxStates    = 1 << 20 // ~1M visited states
	DefaultMaxTermNodes = finite.MaxExprNodes
)

// Search runs deterministic breadth-first rewriting from start under the
// admitted rules, within maxExpansions and the default work ceilings. It
// refuses malformed inputs and out-of-scope rules rather than proceeding
// on unstated premises.
func Search(start finite.Expr, d finite.Domain, rules []Rule, cost CostModel, maxExpansions int) (Result, error) {
	return SearchBounded(start, d, rules, cost, maxExpansions, Limits{})
}

// SearchBounded is Search with explicit work ceilings and cancellation.
func SearchBounded(start finite.Expr, d finite.Domain, rules []Rule, cost CostModel, maxExpansions int, lim Limits) (Result, error) {
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
	if lim.MaxStates <= 0 {
		lim.MaxStates = DefaultMaxStates
	}
	if lim.MaxTermNodes <= 0 {
		lim.MaxTermNodes = DefaultMaxTermNodes
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
search:
	for len(queue) > 0 {
		if lim.Cancel != nil {
			select {
			case <-lim.Cancel:
				res.Cancelled = true
				break search
			default:
			}
		}
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
			successors, dropped := applyEverywhereBounded(current, r, d, lim.MaxTermNodes)
			res.Generated += len(successors) + dropped
			if dropped > 0 {
				// A resource bound truncated the reachable set; the
				// result says so rather than silently narrowing.
				res.TermSizeBounded = true
			}
			for _, next := range successors {
				nextKey := finite.Render(next)
				if _, seen := visited[nextKey]; seen {
					continue
				}
				if len(visited) >= lim.MaxStates {
					res.StateBounded = true
					break search
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

// CanStrictlyReduce reports whether a single application of the rule at
// any position of e strictly reduces the cost. See ProbeStrictReduction
// for the metered form; this convenience discards the work count.
//
// CLAIM SCOPE: a false answer means only "no ONE-STEP strict cost
// decrease from THIS start expression". It is not a proof that the rule
// can never contribute — e.g. add-zero cannot one-step reduce add(0,x),
// yet reduces it after the cost-neutral add-comm step (2026-09-13
// external review finding 2).
func CanStrictlyReduce(e finite.Expr, r Rule, d finite.Domain, cost CostModel) bool {
	reduces, _ := ProbeStrictReduction(e, r, d, cost)
	return reduces
}

// ProbeStrictReduction is the metered task-grounding probe for history
// claims (shape.ControllerVersionV2): deterministic, computable by any arm from
// the task and admitted rules alone, and independent of any history
// content. It returns whether one application of the rule at any
// position of e strictly reduces the cost, and the number of candidate
// rewrites the probe materialized — the probe's work, which the caller
// must charge to whatever cost ledger covers the arm that ran it
// (2026-09-13 external review finding 1: uncharged pre-search probes).
func ProbeStrictReduction(e finite.Expr, r Rule, d finite.Domain, cost CostModel) (reduces bool, candidates int) {
	if cost == nil {
		cost = NodeCount
	}
	base := cost(e)
	// The probe path carries the same successor ceiling the search path
	// does (2026-09-13 validator finding D8: the probe used the unbounded
	// form, so a growth rule could materialize terms the search would
	// refuse). Dropped successors cannot change the answer — the probe
	// asks whether some successor is CHEAPER than the base, and an
	// oversize successor is never cheaper — so bounding costs no
	// discrimination. Dropped candidates are still charged: the work of
	// constructing them was performed.
	successors, dropped := applyEverywhereBounded(e, r, d, DefaultMaxTermNodes)
	candidates = len(successors) + dropped
	for _, next := range successors {
		if cost(next) < base {
			reduces = true
		}
	}
	return reduces, candidates
}

// Identity returns a canonical content identity for the admitted rule:
// name, admitted domain, and both sides' canonical renderings. Two rules
// with equal Identity are the same rewrite; a name alone is not an
// identity (2026-09-13 external review finding 3: a decision hash bound
// to names only misses rule-content changes).
func (r Rule) Identity() string {
	return fmt.Sprintf("%s|w%d|vars=%v|%s=>%s", r.name, r.domain.Width, r.domain.Vars, finite.Render(r.lhs), finite.Render(r.rhs))
}

// applyEverywhere returns every expression obtained by applying the rule
// at exactly one position of e, in deterministic order. Results have
// their constants normalized to the search width so semantically
// identical states share one visited-set key and one rendering
// (adversarial review finding 10).
func applyEverywhere(e finite.Expr, r Rule, d finite.Domain) []finite.Expr {
	out, _ := applyEverywhereBounded(e, r, d, 0)
	return out
}

// applyEverywhereBounded is applyEverywhere with an optional successor
// tree-size ceiling (maxNodes <= 0 disables it). Oversize successors are
// dropped at construction so the caller never renders, hashes, or
// enqueues a term the domain would refuse to admit; dropped reports how
// many were refused, so the caller can say the reachable set was
// truncated instead of silently narrowing it.
func applyEverywhereBounded(e finite.Expr, r Rule, d finite.Domain, maxNodes int) (out []finite.Expr, dropped int) {
	keep := func(x finite.Expr) {
		if maxNodes > 0 && treeSize(x) > maxNodes {
			dropped++
			return
		}
		out = append(out, x)
	}
	if b, ok := match(r.lhs, e, d, map[string]finite.Expr{}); ok {
		keep(normalizeConsts(subst(r.rhs, b), d))
	}
	switch t := e.(type) {
	case finite.Unary:
		sub, subDropped := applyEverywhereBounded(t.X, r, d, maxNodes)
		dropped += subDropped
		for _, v := range sub {
			keep(finite.Unary{Op: t.Op, X: v})
		}
	case finite.Binary:
		left, leftDropped := applyEverywhereBounded(t.X, r, d, maxNodes)
		dropped += leftDropped
		for _, v := range left {
			keep(finite.Binary{Op: t.Op, X: v, Y: t.Y})
		}
		right, rightDropped := applyEverywhereBounded(t.Y, r, d, maxNodes)
		dropped += rightDropped
		for _, v := range right {
			keep(finite.Binary{Op: t.Op, X: t.X, Y: v})
		}
	}
	return out, dropped
}

// treeSize counts logical nodes — the measure finite.MaxExprNodes bounds
// at admission, and the size a canonical rendering expands to. Shared
// subexpressions are counted once per reference, deliberately: that is
// the cost a traversal or rendering actually pays.
func treeSize(e finite.Expr) int {
	switch t := e.(type) {
	case finite.Unary:
		return 1 + treeSize(t.X)
	case finite.Binary:
		return 1 + treeSize(t.X) + treeSize(t.Y)
	}
	return 1
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
