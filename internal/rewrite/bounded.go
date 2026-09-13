package rewrite

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/instagrim-dev/newf/internal/finite"
)

// WorkBudget is an operational allowance shared by sequential probes and
// search. It is not concurrency-safe. Each rule traversal consumes one
// RuleApplications unit; each matched position consumes one Candidates
// unit before size checking or successor construction. Refused oversized
// candidates still consume a unit. A nonnil budget's zero limit allows
// no work of that kind; nil Limits.Work preserves the legacy allowance.
type WorkBudget struct {
	MaxRuleApplications int
	MaxCandidates       int
	RuleApplications    int
	Candidates          int
}

// Validate rejects malformed shared allowances before any work is performed.
func (lim Limits) Validate() error {
	if w := lim.Work; w != nil {
		if w.MaxRuleApplications < 0 || w.MaxCandidates < 0 || w.RuleApplications < 0 || w.Candidates < 0 ||
			w.RuleApplications > w.MaxRuleApplications || w.Candidates > w.MaxCandidates {
			return fmt.Errorf("invalid shared work budget: limits and used counts must be nonnegative, with used counts at or below their limits")
		}
	}
	return nil
}

func (lim Limits) defaults() Limits {
	if lim.MaxStates <= 0 {
		lim.MaxStates = DefaultMaxStates
	}
	if lim.MaxTermNodes <= 0 || lim.MaxTermNodes > DefaultMaxTermNodes {
		lim.MaxTermNodes = DefaultMaxTermNodes
	}
	return lim
}

func cancelled(cancel <-chan struct{}) bool {
	select {
	case <-cancel:
		return true
	default:
		return false
	}
}

// termInfo records each logical position once, so checking a replacement
// can account for its unchanged surroundings without repeatedly walking
// or reconstructing them. Shared subexpressions count per occurrence.
type termInfo struct {
	expr         finite.Expr
	nodes, depth int
	x, y         *termInfo
}

func indexTerm(e finite.Expr, cancel <-chan struct{}) *termInfo {
	if cancelled(cancel) {
		return nil
	}
	n := &termInfo{expr: e, nodes: 1}
	switch t := e.(type) {
	case finite.Unary:
		n.x = indexTerm(t.X, cancel)
		if n.x == nil {
			return nil
		}
		n.nodes += n.x.nodes
		n.depth = n.x.depth + 1
	case finite.Binary:
		n.x = indexTerm(t.X, cancel)
		if n.x == nil {
			return nil
		}
		n.y = indexTerm(t.Y, cancel)
		if n.y == nil {
			return nil
		}
		n.nodes += n.x.nodes + n.y.nodes
		n.depth = 1 + max(n.x.depth, n.y.depth)
	}
	return n
}

type successorStats struct {
	Candidates, RuleApplications                     int
	DepthBounded, Cancelled, ResourceBudgetExhausted bool
}

type ancestor struct {
	expr  finite.Expr
	right bool
}

// visitSuccessors yields one positional successor at a time. Size is
// calculated from substitution metadata before allocation, normalization,
// rendering, or enqueueing. The callback can stop traversal immediately.
func visitSuccessors(e finite.Expr, r Rule, d finite.Domain, lim Limits, yield func(finite.Expr, bool) bool) (stats successorStats) {
	if cancelled(lim.Cancel) {
		stats.Cancelled = true
		return
	}
	if w := lim.Work; w != nil {
		if w.RuleApplications >= w.MaxRuleApplications {
			stats.ResourceBudgetExhausted = true
			return
		}
		w.RuleApplications++
	}
	stats.RuleApplications++
	root := indexTerm(e, lim.Cancel)
	if root == nil {
		stats.Cancelled = true
		return
	}
	var walk func(*termInfo, []ancestor) bool
	walk = func(n *termInfo, path []ancestor) bool {
		if cancelled(lim.Cancel) {
			return false
		}
		bindings := map[string]*termInfo{}
		if matchTerm(r.lhs, n, d, bindings, lim.Cancel) {
			if cancelled(lim.Cancel) {
				return false
			}
			if w := lim.Work; w != nil {
				if w.Candidates >= w.MaxCandidates {
					stats.ResourceBudgetExhausted = true
					return false
				}
				w.Candidates++
			}
			stats.Candidates++
			remaining := lim.MaxTermNodes - (root.nodes - n.nodes)
			depthBounded := false
			fits := measureSubstitution(r.rhs, bindings, &remaining, len(path), &depthBounded, lim.Cancel)
			if cancelled(lim.Cancel) {
				return false
			}
			stats.DepthBounded = stats.DepthBounded || depthBounded
			if !fits {
				if !yield(nil, true) {
					return false
				}
			} else {
				next := instantiate(r.rhs, bindings, d, lim.Cancel)
				if next == nil {
					return false
				}
				for i := len(path) - 1; i >= 0; i-- {
					if cancelled(lim.Cancel) {
						return false
					}
					p := path[i]
					switch t := p.expr.(type) {
					case finite.Unary:
						next = finite.Unary{Op: t.Op, X: next}
					case finite.Binary:
						if p.right {
							next = finite.Binary{Op: t.Op, X: t.X, Y: next}
						} else {
							next = finite.Binary{Op: t.Op, X: next, Y: t.Y}
						}
					}
				}
				if !yield(next, false) {
					return false
				}
			}
		}
		if n.x != nil && !walk(n.x, append(path, ancestor{expr: n.expr})) {
			return false
		}
		return n.y == nil || walk(n.y, append(path, ancestor{expr: n.expr, right: true}))
	}
	walk(root, nil)
	stats.Cancelled = cancelled(lim.Cancel)
	return
}

// measureSubstitution does no expression allocation and skips an entire
// bound subtree using its measured logical size. A duplicated 4K-node
// binding is rejected before normalization can expand its references.
func measureSubstitution(e finite.Expr, bindings map[string]*termInfo, remaining *int, depth int, depthBounded *bool, cancel <-chan struct{}) bool {
	if cancelled(cancel) {
		return false
	}
	if depth > finite.MaxExprDepth {
		*depthBounded = true
		return false
	}
	if t, ok := e.(finite.Var); ok {
		n := bindings[t.Name]
		if depth+n.depth > finite.MaxExprDepth {
			*depthBounded = true
			return false
		}
		*remaining -= n.nodes
		return *remaining >= 0
	}
	*remaining--
	if *remaining < 0 {
		return false
	}
	switch t := e.(type) {
	case finite.Unary:
		return measureSubstitution(t.X, bindings, remaining, depth+1, depthBounded, cancel)
	case finite.Binary:
		return measureSubstitution(t.X, bindings, remaining, depth+1, depthBounded, cancel) &&
			measureSubstitution(t.Y, bindings, remaining, depth+1, depthBounded, cancel)
	}
	return true
}

func matchTerm(pattern finite.Expr, subject *termInfo, d finite.Domain, bindings map[string]*termInfo, cancel <-chan struct{}) bool {
	if cancelled(cancel) {
		return false
	}
	switch p := pattern.(type) {
	case finite.Var:
		if prev, ok := bindings[p.Name]; ok {
			return equalTerms(prev, subject, cancel)
		}
		bindings[p.Name] = subject
		return true
	case finite.Const:
		s, ok := subject.expr.(finite.Const)
		mask := uint64(1)<<uint(d.Width) - 1
		return ok && p.Value&mask == s.Value&mask
	case finite.Unary:
		s, ok := subject.expr.(finite.Unary)
		return ok && s.Op == p.Op && matchTerm(p.X, subject.x, d, bindings, cancel)
	case finite.Binary:
		s, ok := subject.expr.(finite.Binary)
		return ok && s.Op == p.Op && matchTerm(p.X, subject.x, d, bindings, cancel) && matchTerm(p.Y, subject.y, d, bindings, cancel)
	}
	return false
}

// Repeated metavariables use structural equality, avoiding two temporary
// rendered keys for every comparison. Constants retain the same raw-value
// equality as the original canonical-rendering comparison.
func equalTerms(a, b *termInfo, cancel <-chan struct{}) bool {
	if cancelled(cancel) || a.nodes != b.nodes {
		return false
	}
	switch x := a.expr.(type) {
	case finite.Var:
		y, ok := b.expr.(finite.Var)
		return ok && x.Name == y.Name
	case finite.Const:
		y, ok := b.expr.(finite.Const)
		return ok && x.Value == y.Value
	case finite.Unary:
		y, ok := b.expr.(finite.Unary)
		return ok && x.Op == y.Op && equalTerms(a.x, b.x, cancel)
	case finite.Binary:
		y, ok := b.expr.(finite.Binary)
		return ok && x.Op == y.Op && equalTerms(a.x, b.x, cancel) && equalTerms(a.y, b.y, cancel)
	}
	return false
}

func instantiate(e finite.Expr, bindings map[string]*termInfo, d finite.Domain, cancel <-chan struct{}) finite.Expr {
	if cancelled(cancel) {
		return nil
	}
	switch t := e.(type) {
	case finite.Var:
		if bindings != nil {
			return instantiate(bindings[t.Name].expr, nil, d, cancel)
		}
		return t
	case finite.Const:
		return finite.Const{Value: t.Value & (uint64(1)<<uint(d.Width) - 1)}
	case finite.Unary:
		x := instantiate(t.X, bindings, d, cancel)
		if x == nil {
			return nil
		}
		return finite.Unary{Op: t.Op, X: x}
	case finite.Binary:
		x := instantiate(t.X, bindings, d, cancel)
		if x == nil {
			return nil
		}
		y := instantiate(t.Y, bindings, d, cancel)
		if y == nil {
			return nil
		}
		return finite.Binary{Op: t.Op, X: x, Y: y}
	}
	return nil
}

// render uses one bounded builder and checks cancellation at each node.
func render(e finite.Expr, cancel <-chan struct{}) (string, bool) {
	var b strings.Builder
	var walk func(finite.Expr) bool
	walk = func(e finite.Expr) bool {
		if cancelled(cancel) {
			return false
		}
		switch t := e.(type) {
		case finite.Var:
			b.WriteString(t.Name)
		case finite.Const:
			b.WriteString(strconv.FormatUint(t.Value, 10))
		case finite.Unary:
			b.WriteString(string(t.Op))
			b.WriteByte('(')
			if !walk(t.X) {
				return false
			}
			b.WriteByte(')')
		case finite.Binary:
			b.WriteString(string(t.Op))
			b.WriteByte('(')
			if !walk(t.X) {
				return false
			}
			b.WriteString(", ")
			if !walk(t.Y) {
				return false
			}
			b.WriteByte(')')
		}
		return true
	}
	ok := walk(e)
	return b.String(), ok
}
