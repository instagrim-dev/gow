package finite

import (
	"strings"
	"testing"
)

func TestIdentifierByteBoundPrecedesValidationRendering(t *testing.T) {
	oversize := strings.Repeat("x", MaxIdentifierBytes+1)
	for _, d := range []Domain{{Width: 4, Vars: []string{oversize}}, {Width: 4, Vars: []string{"x", oversize}}} {
		defects := ValidateExpr(Var{Name: "x"}, d)
		if len(defects) == 0 || !strings.Contains(strings.Join(defects, " "), "resource refusal") {
			t.Fatalf("oversized declared name must be refused: %v", defects)
		}
		cert := AssessEquivalence(Binding{Sentence: "bounded rendering", Domain: d}, Var{Name: "x"}, Var{Name: "x"})
		if cert.Verdict != VerdictInapplicable || strings.Contains(cert.Reason, oversize) || len(cert.NotAssessed) != 0 {
			t.Fatalf("domain must be refused before rendering its oversized name: %+v", cert)
		}
	}
	if defects := ValidateExpr(Var{Name: oversize}, Domain{Width: 4, Vars: []string{"x"}}); len(defects) == 0 {
		t.Fatal("undeclared expression name still requires the byte bound")
	}
	atLimit := strings.Repeat("x", MaxIdentifierBytes)
	if defects := ValidateExpr(Var{Name: atLimit}, Domain{Width: 4, Vars: []string{atLimit}}); len(defects) > 0 {
		t.Fatalf("identifier at the documented bound must be admitted: %v", defects)
	}
}

func TestPreCancelledIndependentReplayDoesNotClaimExhaustiveness(t *testing.T) {
	cancel := make(chan struct{})
	close(cancel)
	x := Var{Name: "x"}
	cert := AssessEquivalenceCancelled(Binding{Sentence: "cancelled replay", Domain: dom4("x")}, x, x, cancel)
	if cert.Verdict != VerdictUnresolved || cert.Exhaustive || cert.AssignmentsChecked != 0 {
		t.Fatalf("cancelled replay must preserve an unverified receipt: %+v", cert)
	}
	if _, ok := evalCancelled(Binary{Op: OpAdd, X: x, Y: x}, dom4("x"), Assignment{"x": 1}, cancel); ok {
		t.Fatal("recursive evaluation ignored cancellation")
	}
}
