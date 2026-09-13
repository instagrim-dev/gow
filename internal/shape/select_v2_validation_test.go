package shape

import (
	"reflect"
	"strings"
	"testing"

	"github.com/instagrim-dev/newf/internal/finite"
	"github.com/instagrim-dev/newf/internal/rewrite"
)

func TestSelectorRejectsIncompleteOrIncompatibleInputs(t *testing.T) {
	task := nn(v("x"))
	oneBitDomain := finite.Domain{Width: 1, Vars: []string{"a"}}
	lhs, rhs := nn(v("a")), v("a")
	cert := finite.AssessEquivalence(finite.Binding{Sentence: "one-bit double-not", Domain: oneBitDomain}, lhs, rhs)
	oneBitRule, defects := rewrite.AdmitRule("double-not", cert, lhs, rhs, oneBitDomain)
	if len(defects) > 0 {
		t.Fatalf("admit one-bit fixture: %v", defects)
	}
	rule := admitRule(t, "double-not", lhs, rhs)
	base := InputV2{
		Input: Input{
			TaskStart: RenderExpr(task), Target: 1, Catalog: []string{"double-not"},
			History: History{{Start: RenderExpr(task), RulesApplied: []string{"double-not"}, Completed: true}},
		},
		Task: task, Domain: finite.Domain{Width: 4, Vars: []string{"x"}}, Rules: []rewrite.Rule{rule},
	}
	tests := []struct {
		name string
		edit func(*InputV2)
		want string
	}{
		{"missing completed-history rule", func(in *InputV2) { in.Rules = nil }, "no admitted rule"},
		{"missing failed-history rule", func(in *InputV2) {
			in.Rules = nil
			in.History = History{{Start: RenderExpr(task), RulesApplied: []string{"double-not"}}}
		}, "no admitted rule"},
		{"missing uncredited rule", func(in *InputV2) {
			in.Rules = nil
			in.History = nil
		}, "no admitted rule"},
		{"wrong admitted width", func(in *InputV2) { in.Rules = []rewrite.Rule{oneBitRule} }, "admitted at width 1"},
		{"zero rule", func(in *InputV2) { in.Rules = []rewrite.Rule{{}} }, "unadmitted zero-value rule"},
		{"zero width", func(in *InputV2) { in.Domain.Width = 0 }, "outside the supported range"},
		{"excessive width", func(in *InputV2) { in.Domain.Width = finite.MaxWidth + 1 }, "outside the supported range"},
		{"negative width", func(in *InputV2) { in.Domain.Width = -1 }, "outside the supported range"},
		{"nil task", func(in *InputV2) { in.Task = nil }, "expression node is nil"},
		{"unlisted rule", func(in *InputV2) { in.Catalog = nil }, "absent from the catalog"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := base
			tt.edit(&in)
			defer func() {
				if got := recover(); got != nil {
					t.Fatalf("invalid selector input must return an error before rendering or probing, panicked: %v", got)
				}
			}()
			decision, err := SelectV2(in)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("want refusal containing %q, got error %v and decision %+v", tt.want, err, decision)
			}
			if !reflect.DeepEqual(decision, Decision{}) {
				t.Fatalf("refused input must yield no hashed decision, probe evidence, or rationale: %+v", decision)
			}
		})
	}
}

func TestSelectorValidationPreservesAcceptedInputIdentity(t *testing.T) {
	task := nn(v("x"))
	rule := admitRule(t, "double-not", nn(v("a")), v("a"))
	in := InputV2{
		Input: Input{TaskStart: RenderExpr(task), Target: 1, Catalog: []string{"double-not"}},
		Task:  task, Domain: finite.Domain{Width: 4, Vars: []string{"x"}}, Rules: []rewrite.Rule{rule},
	}
	decision, err := SelectV2(in)
	if err != nil {
		t.Fatal(err)
	}
	if decision.ControllerVersion != "shape-selector/2" || decision.SnapshotHash != frozenSnapshotV2 ||
		decision.InputHash != "b0ce5ccbef18955aeeabc17731c2cb6fc3161ffb4374c7b159435d1feea7d2f9" {
		t.Fatalf("validation must preserve the pre-repair identity of valid inputs: %+v", decision)
	}
	duplicate := in
	duplicate.Catalog = []string{"double-not", "double-not"}
	duplicate.Rules = []rewrite.Rule{rule, rule}
	duplicateDecision, err := SelectV2(duplicate)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decision, duplicateDecision) {
		t.Fatalf("exact duplicates must preserve the existing normalized decision: %+v vs %+v", decision, duplicateDecision)
	}
}
