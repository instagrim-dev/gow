package finite

import (
	"strings"
	"testing"
)

func TestValidateExprChecksDomainDeclaration(t *testing.T) {
	for _, tt := range []struct {
		name   string
		domain Domain
		want   string
	}{
		{"zero width", Domain{Width: 0}, "outside the supported range"},
		{"negative width", Domain{Width: -1}, "outside the supported range"},
		{"unsupported width", Domain{Width: MaxWidth + 1}, "outside the supported range"},
		{"empty variable", Domain{Width: 4, Vars: []string{""}}, "not a plain identifier"},
		{"duplicate variable", Domain{Width: 4, Vars: []string{"x", "x"}}, "duplicated"},
		{"nonidentifier variable", Domain{Width: 4, Vars: []string{"not(x)"}}, "not a plain identifier"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			defects := ValidateExpr(Const{Value: 0}, tt.domain)
			if !strings.Contains(strings.Join(defects, "; "), tt.want) {
				t.Fatalf("want %q, got %v", tt.want, defects)
			}
		})
	}
	// A valid domain above the independent oracle's cap remains usable
	// for candidate search; domain validation does not grant verification.
	domain := Domain{Width: 8, Vars: []string{"a", "b", "c"}}
	if defects := ValidateExpr(Var{Name: "a"}, domain); len(defects) != 0 {
		t.Fatalf("valid nonexhaustive domain refused: %v", defects)
	}
}
