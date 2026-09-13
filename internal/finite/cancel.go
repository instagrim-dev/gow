package finite

func isCancelled(cancel <-chan struct{}) bool {
	select {
	case <-cancel:
		return true
	default:
		return false
	}
}

// evalCancelled preserves the language's evaluator as the operator
// authority while checking cancellation at every recursive node visit.
func evalCancelled(e Expr, d Domain, a Assignment, cancel <-chan struct{}) (uint64, bool) {
	if cancel == nil {
		return e.eval(d, a), true
	}
	if isCancelled(cancel) {
		return 0, false
	}
	switch t := e.(type) {
	case Unary:
		x, ok := evalCancelled(t.X, d, a, cancel)
		if !ok {
			return 0, false
		}
		return (Unary{Op: t.Op, X: Const{Value: x}}).eval(d, a), true
	case Binary:
		x, ok := evalCancelled(t.X, d, a, cancel)
		if !ok {
			return 0, false
		}
		y, ok := evalCancelled(t.Y, d, a, cancel)
		if !ok {
			return 0, false
		}
		return (Binary{Op: t.Op, X: Const{Value: x}, Y: Const{Value: y}}).eval(d, a), true
	default:
		return e.eval(d, a), true
	}
}
