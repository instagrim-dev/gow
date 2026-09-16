package finite

import (
	"strings"
	"testing"
)

func TestParseTermRoundTripsCanonicalExpression(t *testing.T) {
	want := "not(add(x, 0))"
	expr, err := ParseTerm(want)
	if err != nil {
		t.Fatal(err)
	}
	if defects := ValidateExpr(expr, Domain{Width: 4, Vars: []string{"x"}}); len(defects) != 0 {
		t.Fatalf("parsed expression invalid: %v", defects)
	}
	if got := Render(expr); got != want {
		t.Fatalf("render = %q, want %q", got, want)
	}
}

func TestParseTermSeparatesSyntaxFromSemantics(t *testing.T) {
	expr, err := ParseTerm("invented(x)")
	if err != nil {
		t.Fatalf("well-formed unknown operator should parse: %v", err)
	}
	if defects := ValidateExpr(expr, Domain{Width: 4, Vars: []string{"x"}}); len(defects) == 0 || !strings.Contains(defects[0], "unknown unary operator") {
		t.Fatalf("semantic defect missing: %v", defects)
	}
	for _, input := range []string{"", "add(x,)", "add(x,0", "add(x,0,1)", "x trailing"} {
		if _, err := ParseTerm(input); err == nil {
			t.Fatalf("malformed expression %q accepted", input)
		}
	}
}
