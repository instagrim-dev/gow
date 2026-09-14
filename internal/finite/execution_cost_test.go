package finite

import "testing"

func TestMeasureExecutionCostCountsActualEvaluatorVisits(t *testing.T) {
	d := Domain{Width: 2, Vars: []string{"x"}}
	expr := Binary{Op: OpAdd, X: Var{Name: "x"}, Y: Const{Value: 0}}
	cost, err := MeasureExecutionCost(expr, d)
	if err != nil {
		t.Fatal(err)
	}
	if cost.Assignments != 4 || cost.NodeVisits != 12 {
		t.Fatalf("cost = %+v, want 4 assignments and 12 visits", cost)
	}
}

func TestMeasureExecutionCostRefusesAnUnboundedDomain(t *testing.T) {
	if _, err := MeasureExecutionCost(Var{Name: "x"}, Domain{Width: 8, Vars: []string{"x", "y", "z"}}); err == nil {
		t.Fatal("over-cap execution measurement must be refused")
	}
}
