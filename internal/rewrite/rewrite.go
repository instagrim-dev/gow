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
	"reflect"
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

// EngineRule is the immutable, admitted rule material an external equality
// engine may receive. It is intentionally produced only by Rule.Export: an
// engine adapter never accepts an arbitrary rule-shaped payload as authority
// to union terms.
type EngineRule struct {
	Name     string
	Identity string
	Domain   finite.Domain
	Left     finite.Expr
	Right    finite.Expr
}

// StepBinding is one pattern metavariable assignment cited by an external
// engine explanation. It stays expression-typed so callers cannot make an
// unchecked string look like a replayed substitution.
type StepBinding struct {
	Variable string
	Term     finite.Expr
}

// Name returns the rule's admission name.
func (r Rule) Name() string { return r.name }

// Export returns the rule's admitted definition for a bounded external engine.
// The returned domain slice is copied so callers cannot mutate the Rule's
// identity through the exported view.
func (r Rule) Export() EngineRule {
	return EngineRule{
		Name:     r.name,
		Identity: r.Identity(),
		Domain: finite.Domain{
			Width: r.domain.Width,
			Vars:  append([]string(nil), r.domain.Vars...),
		},
		Left:  r.lhs,
		Right: r.rhs,
	}
}

// ValidateForDomain checks the admission boundary before rendering,
// matching, or hashing a rule. Pattern variables are metavariables and
// need not share the subject's names; the admitted word width must match.
func (r Rule) ValidateForDomain(d finite.Domain) error {
	if r.name == "" {
		return fmt.Errorf("an unadmitted zero-value rule was supplied; rules exist only via AdmitRule")
	}
	if defects := finite.ValidateDomain(d); len(defects) > 0 {
		return fmt.Errorf("invalid rule application domain: %v", defects)
	}
	if r.domain.Width != d.Width {
		return fmt.Errorf("rule %q was admitted at width %d and cannot be applied at width %d; a certificate warrants exactly its domain", r.name, r.domain.Width, d.Width)
	}
	for _, side := range []struct {
		name string
		expr finite.Expr
	}{{"left", r.lhs}, {"right", r.rhs}} {
		if defects := finite.ValidateExpr(side.expr, r.domain); len(defects) > 0 {
			return fmt.Errorf("rule %q has an invalid %s side in its admitted domain: %v", r.name, side.name, defects)
		}
	}
	return nil
}

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
	// Generated counts matched candidate positions whose size preflight
	// began, including rejected terms and before deduplication. It is
	// the same unit of work a selector probe performs per candidate, so
	// callers can charge search and pre-search probe work on one ledger
	// (2026-09-13 external review finding 1). Explored (expansions)
	// remains the budget unit; the two are different measures and must
	// not be conflated.
	Generated int
	// RuleApplications counts rule traversals begun, including no-match traversals.
	RuleApplications int
	// ResourceBudgetExhausted is a stop under the caller's shared operational allowance.
	ResourceBudgetExhausted bool
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
	// TermSizeBounded: at least one successor exceeded the node/depth
	// ceiling and was not constructed. The reachable set was truncated by a
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
	// states). A reachable set that fits the cap exactly completes
	// normally; exceeding it — refusing a new unseen successor because
	// the set is full — stops the search with StateBounded.
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
	// Work, when supplied, is shared by probes and search. A zero cap is a
	// zero allowance; nil retains the legacy expansion-only allowance.
	Work *WorkBudget
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
	if err := lim.Validate(); err != nil {
		return Result{}, err
	}
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
		if err := r.ValidateForDomain(d); err != nil {
			return Result{}, err
		}
	}
	lim = lim.defaults()

	type parentEdge struct {
		parent string
		step   Step
	}
	// Even the initial key obeys cancellation while rendering. A cancelled
	// partial key is never exposed as a canonical expression identity.
	startKey, rendered := render(start, lim.Cancel)
	if !rendered {
		return Result{Cancelled: true}, nil
	}
	visited := map[string]finite.Expr{startKey: start}
	parents := map[string]parentEdge{}
	queue := []string{startKey}

	if cancelled(lim.Cancel) && !nodeCountCost(cost) {
		return Result{Original: startKey, Best: startKey, Cancelled: true}, fmt.Errorf("search cancelled before custom cost assessment; costs and endpoint were not assessed")
	}
	initialCost := NodeCount(start)
	if !cancelled(lim.Cancel) {
		initialCost = cost(start)
	}
	res := Result{
		Original:     startKey,
		Best:         startKey,
		OriginalCost: initialCost,
		BestCost:     initialCost,
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
			stats := visitSuccessors(current, r, d, lim, func(next finite.Expr, dropped bool) bool {
				if dropped {
					res.TermSizeBounded = true
					return true
				}
				if cancelled(lim.Cancel) {
					res.Cancelled = true
					return false
				}
				nextKey, ok := render(next, lim.Cancel)
				if !ok {
					res.Cancelled = true
					return false
				}
				if _, seen := visited[nextKey]; seen {
					return true
				}
				// Refusal semantics: StateBounded is set only when an
				// UNSEEN successor is denied admission because the set
				// is full. A reachable set that fits MaxStates exactly
				// completes with StateBounded=false.
				if len(visited) >= lim.MaxStates {
					res.StateBounded = true
					return false
				}
				visited[nextKey] = next
				parents[nextKey] = parentEdge{parent: key, step: Step{Rule: r.name, Before: key, After: nextKey}}
				queue = append(queue, nextKey)
				c := cost(next)
				if c < res.BestCost || (c == res.BestCost && nextKey < res.Best) {
					res.Best = nextKey
					res.BestCost = c
				}
				return !cancelled(lim.Cancel)
			})
			res.Generated += stats.Candidates
			res.RuleApplications += stats.RuleApplications
			res.Cancelled = res.Cancelled || stats.Cancelled
			res.ResourceBudgetExhausted = stats.ResourceBudgetExhausted
			if res.StateBounded || res.Cancelled || res.ResourceBudgetExhausted {
				break search
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
	res.Endpoint = finite.AssessEquivalenceCancelled(finite.Binding{
		Sentence: fmt.Sprintf("search endpoint: %s == %s", res.Original, res.Best),
		Domain:   d,
	}, start, best, lim.Cancel)
	res.Cancelled = res.Cancelled || cancelled(lim.Cancel)
	res.EndpointVerified = res.Endpoint.Verdict == finite.VerdictHoldsOnDomain
	return res, nil
}

// CanStrictlyReduce is the NodeCount-only convenience for the metered probe.
// A false answer excludes a one-step NodeCount improvement only; it says
// nothing about later rewrites or the truth of a history claim.
func CanStrictlyReduce(e finite.Expr, r Rule, d finite.Domain, cost CostModel) bool {
	reduces, _ := ProbeStrictReduction(e, r, d, cost)
	return reduces
}

// ProbeStrictReduction retains the original NodeCount probe API. Custom
// costs must use ProbeStrictReductionBounded and inspect Complete: a
// pruned larger expression may still be cheaper under a custom cost.
// Unsupported costs and invalid premises panic explicitly rather than
// turning an unperformed or incomplete assessment into a negative answer.
func ProbeStrictReduction(e finite.Expr, r Rule, d finite.Domain, cost CostModel) (bool, int) {
	if !nodeCountCost(cost) {
		panic("custom probe cost requires ProbeStrictReductionBounded and its Complete assessment")
	}
	res, err := ProbeStrictReductionBounded(e, r, d, cost, Limits{})
	if err != nil {
		panic(err)
	}
	if !res.Complete {
		panic("incomplete reduction probe requires ProbeStrictReductionBounded and its Complete assessment")
	}
	return res.Reduces, res.Candidates
}

// ProbeResult distinguishes a witnessed improvement from an exhaustive
// one-step negative. Reduces remains a valid witness when Complete is
// false; only Complete && !Reduces supports a negative assessment.
type ProbeResult struct {
	Reduces, Complete                                   bool
	Candidates, RuleApplications                        int
	TermSizeBounded, Cancelled, ResourceBudgetExhausted bool
}

// ProbeStrictReductionBounded consumes the same operational allowance as
// SearchBounded. Candidates counts matched positions whose size preflight
// began, including refused terms; no successor batch is constructed.
// A custom cost callback must return promptly: cancellation is checked
// before and after it, but Go callbacks cannot be forcibly preempted.
func ProbeStrictReductionBounded(e finite.Expr, r Rule, d finite.Domain, cost CostModel, lim Limits) (ProbeResult, error) {
	if err := lim.Validate(); err != nil {
		return ProbeResult{}, err
	}
	if defects := finite.ValidateExpr(e, d); len(defects) > 0 {
		return ProbeResult{}, fmt.Errorf("probe expression is invalid: %v", defects)
	}
	if err := r.ValidateForDomain(d); err != nil {
		return ProbeResult{}, err
	}
	isNodeCount := nodeCountCost(cost)
	if cost == nil {
		cost = NodeCount
	}
	lim = lim.defaults()
	res := ProbeResult{Complete: true}
	if cancelled(lim.Cancel) {
		res.Complete, res.Cancelled = false, true
		return res, nil
	}
	base := cost(e)
	stats := visitSuccessors(e, r, d, lim, func(next finite.Expr, dropped bool) bool {
		if dropped {
			res.TermSizeBounded = true
			// A node bound at or above the starting node count cannot
			// conceal a NodeCount improvement. Depth pruning still can.
			res.Complete = false
			return true
		}
		if cost(next) < base {
			res.Reduces = true
		}
		return !cancelled(lim.Cancel)
	})
	res.Candidates, res.RuleApplications = stats.Candidates, stats.RuleApplications
	res.Cancelled, res.ResourceBudgetExhausted = stats.Cancelled, stats.ResourceBudgetExhausted
	if isNodeCount && !stats.DepthBounded && int64(lim.MaxTermNodes) >= base {
		res.Complete = true
	}
	res.Complete = res.Complete && !res.Cancelled && !res.ResourceBudgetExhausted
	return res, nil
}

func nodeCountCost(cost CostModel) bool {
	return cost == nil || reflect.ValueOf(cost).Pointer() == reflect.ValueOf(NodeCount).Pointer()
}

// Identity returns a canonical content identity for the admitted rule:
// name, admitted domain, and both sides' canonical renderings. Two rules
// with equal Identity are the same rewrite; a name alone is not an
// identity (2026-09-13 external review finding 3: a decision hash bound
// to names only misses rule-content changes).
func (r Rule) Identity() string {
	return fmt.Sprintf("%s|w%d|vars=%v|%s=>%s", r.name, r.domain.Width, r.domain.Vars, finite.Render(r.lhs), finite.Render(r.rhs))
}

// ReplaysOneStep reports whether after is reachable from before by exactly one
// positional application of this admitted rule. Reverse is safe only because
// Rule exists after an equality warrant; it asks the same scoped equality in
// the opposite direction. This is an explanation checker, not a search API:
// it preserves all possible matching positions and stops after finding the
// claimed target.
func (r Rule) ReplaysOneStep(before, after finite.Expr, d finite.Domain, reverse bool) (bool, error) {
	return r.replaysOneStep(before, after, d, reverse, nil)
}

// ReplaysOneStepWithBindings additionally requires the reported pattern
// substitution to be exactly the one used by a positional rewrite. It checks
// source, target, direction, and substitutions as one proof step; no wire
// metadata becomes trusted merely because the endpoint later holds.
func (r Rule) ReplaysOneStepWithBindings(before, after finite.Expr, d finite.Domain, reverse bool, bindings []StepBinding) (bool, error) {
	return r.replaysOneStep(before, after, d, reverse, &bindings)
}

func (r Rule) replaysOneStep(before, after finite.Expr, d finite.Domain, reverse bool, expected *[]StepBinding) (bool, error) {
	if defects := finite.ValidateExpr(before, d); len(defects) > 0 {
		return false, fmt.Errorf("invalid replay source: %v", defects)
	}
	if defects := finite.ValidateExpr(after, d); len(defects) > 0 {
		return false, fmt.Errorf("invalid replay target: %v", defects)
	}
	if err := r.ValidateForDomain(d); err != nil {
		return false, err
	}
	oriented := r
	if reverse {
		oriented.lhs, oriented.rhs = r.rhs, r.lhs
	}
	target := finite.Render(after)
	found := false
	root := indexTerm(before, nil)
	var walk func(*termInfo, []ancestor) bool
	walk = func(at *termInfo, path []ancestor) bool {
		actual := map[string]*termInfo{}
		if matchTerm(oriented.lhs, at, d, actual, nil) {
			remaining := finite.MaxExprNodes - (root.nodes - at.nodes)
			depthBounded := false
			if measureSubstitution(oriented.rhs, actual, &remaining, len(path), &depthBounded, nil) {
				next := instantiate(oriented.rhs, actual, d, nil)
				for i := len(path) - 1; i >= 0; i-- {
					parent := path[i]
					switch p := parent.expr.(type) {
					case finite.Unary:
						next = finite.Unary{Op: p.Op, X: next}
					case finite.Binary:
						if parent.right {
							next = finite.Binary{Op: p.Op, X: p.X, Y: next}
						} else {
							next = finite.Binary{Op: p.Op, X: next, Y: p.Y}
						}
					}
				}
				if finite.Render(next) == target && (expected == nil || bindingsMatch(actual, *expected)) {
					found = true
					return false
				}
			}
		}
		if at.x != nil && !walk(at.x, append(path, ancestor{expr: at.expr})) {
			return false
		}
		return at.y == nil || walk(at.y, append(path, ancestor{expr: at.expr, right: true}))
	}
	walk(root, nil)
	return found, nil
}

func bindingsMatch(actual map[string]*termInfo, expected []StepBinding) bool {
	if len(actual) != len(expected) {
		return false
	}
	seen := make(map[string]bool, len(expected))
	for _, binding := range expected {
		if binding.Variable == "" || seen[binding.Variable] {
			return false
		}
		seen[binding.Variable] = true
		actualTerm, ok := actual[binding.Variable]
		if !ok || binding.Term == nil {
			return false
		}
		expectedTerm := indexTerm(binding.Term, nil)
		if expectedTerm == nil || !equalTerms(actualTerm, expectedTerm, nil) {
			return false
		}
	}
	return true
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
