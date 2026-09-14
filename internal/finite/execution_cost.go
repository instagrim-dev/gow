package finite

import "fmt"

// ExecutionCost is the measured interpreter work for one expression over its
// declared finite domain. It is intentionally separate from an extractor's
// predicted objective: node count is a selection heuristic, while this counts
// the evaluator node visits actually performed over concrete assignments.
type ExecutionCost struct {
	Assignments int64
	NodeVisits  int64
}

// MeasureExecutionCost evaluates e over every assignment of d and counts each
// interpreted expression node visit. It refuses malformed or oversized inputs
// rather than treating a partial measurement as a cost result.
func MeasureExecutionCost(e Expr, d Domain) (ExecutionCost, error) {
	if defects := ValidateDomain(d); len(defects) > 0 {
		return ExecutionCost{}, fmt.Errorf("invalid execution-cost domain: %v", defects)
	}
	if defects := ValidateExpr(e, d); len(defects) > 0 {
		return ExecutionCost{}, fmt.Errorf("invalid execution-cost expression: %v", defects)
	}
	if d.Size() > ExhaustiveCap {
		return ExecutionCost{}, fmt.Errorf("execution-cost domain has %d assignments, above exhaustive cap %d", d.Size(), ExhaustiveCap)
	}
	var cost ExecutionCost
	enumerate(d, func(a Assignment) bool {
		_, visits := evalCounted(e, d, a)
		cost.Assignments++
		cost.NodeVisits += visits
		return true
	})
	return cost, nil
}

func evalCounted(e Expr, d Domain, a Assignment) (uint64, int64) {
	switch term := e.(type) {
	case Var:
		return a[term.Name], 1
	case Const:
		return term.Value & d.mask(), 1
	case Unary:
		x, visits := evalCounted(term.X, d, a)
		return Unary{Op: term.Op, X: Const{Value: x}}.eval(d, a), visits + 1
	case Binary:
		x, xVisits := evalCounted(term.X, d, a)
		y, yVisits := evalCounted(term.Y, d, a)
		return Binary{Op: term.Op, X: Const{Value: x}, Y: Const{Value: y}}.eval(d, a), xVisits + yVisits + 1
	default:
		return 0, 0 // ValidateExpr rejects foreign nodes before this helper runs.
	}
}
