package shape

import (
	"strings"
	"testing"

	"github.com/instagrim-dev/newf/internal/finite"
	"github.com/instagrim-dev/newf/internal/rewrite"
)

// v1 A/B and mechanism tests. The screen's run-2 defect (record 019):
// v0 admits lookalike lying histories at full weight. v1 grounds each
// history claim against the task via the strict-reduction probe.

func admitV1(t *testing.T, name string, lhs, rhs finite.Expr) rewrite.Rule {
	t.Helper()
	d := finite.Domain{Width: 4, Vars: []string{"a"}}
	cert := finite.AssessEquivalence(finite.Binding{Sentence: name, Domain: d}, lhs, rhs)
	rule, defects := rewrite.AdmitRule(name, cert, lhs, rhs, d)
	if len(defects) > 0 {
		t.Fatalf("rule %q refused: %v", name, defects)
	}
	return rule
}

// The measured mis-01 shape: lying history credits the explosive rule as
// the success and the true reducers as failures. v0 front-loads the lie;
// v1 must distrust it (not-intro can never reduce) AND distrust the
// failure claims (the reducers demonstrably reduce), restoring an order
// no worse than the catalog's.
func TestV1NeutralizesBlatantLie(t *testing.T) {
	task := nn(addZ(v("x")))
	catalog := []string{"double-not", "add-zero", "not-intro"}
	rules := []rewrite.Rule{
		admitV1(t, "double-not", nn(v("a")), v("a")),
		admitV1(t, "add-zero", addZ(v("a")), v("a")),
		admitV1(t, "not-intro", v("a"), nn(v("a"))),
	}
	in := InputV1{
		Input: Input{
			TaskStart: RenderExpr(task), Target: 1, Catalog: catalog,
			History: History{
				{Start: RenderExpr(nn(addZ(v("w")))), RulesApplied: []string{"not-intro"}, Completed: true},
				{Start: RenderExpr(nn(addZ(v("u")))), RulesApplied: []string{"double-not", "add-zero"}, Completed: false},
			},
		},
		Task: task, Domain: finite.Domain{Width: 4, Vars: []string{"x"}}, Rules: rules,
	}
	dec := SelectV1(in)
	if dec.EnabledRules[len(dec.EnabledRules)-1] != "not-intro" {
		t.Fatalf("the lying success claim must be demoted last: %v", dec.EnabledRules)
	}
	if dec.EnabledRules[0] == "not-intro" {
		t.Fatalf("v1 must not front-load the lie: %v", dec.EnabledRules)
	}
	distrusts := 0
	for _, p := range dec.Preferences {
		if strings.Contains(p.Rationale, "distrusted") {
			distrusts++
		}
	}
	if distrusts == 0 {
		t.Fatalf("the distrust must be recorded with its rationale: %+v", dec.Preferences)
	}
	if dec.ControllerVersion != ControllerVersionV1 {
		t.Fatalf("wrong version: %s", dec.ControllerVersion)
	}
}

// Informative history must still help: credited reducers that pass the
// probe are preferred, exactly as in v0's happy path.
func TestV1InformativeHistoryStillPrefersReducers(t *testing.T) {
	task := nn(addZ(nn(v("x"))))
	catalog := []string{"not-intro", "double-not", "add-zero"} // bad neutral order
	rules := []rewrite.Rule{
		admitV1(t, "double-not", nn(v("a")), v("a")),
		admitV1(t, "add-zero", addZ(v("a")), v("a")),
		admitV1(t, "not-intro", v("a"), nn(v("a"))),
	}
	in := InputV1{
		Input: Input{
			TaskStart: RenderExpr(task), Target: 1, Catalog: catalog,
			History: History{
				{Start: RenderExpr(nn(addZ(nn(v("y"))))), RulesApplied: []string{"double-not", "add-zero"}, Completed: true},
			},
		},
		Task: task, Domain: finite.Domain{Width: 4, Vars: []string{"x"}}, Rules: rules,
	}
	dec := SelectV1(in)
	if dec.EnabledRules[0] == "not-intro" {
		t.Fatalf("informative history must promote the reducers: %v", dec.EnabledRules)
	}
	if dec.EnabledRules[len(dec.EnabledRules)-1] != "not-intro" {
		t.Fatalf("the uncredited explosive stays behind the credited reducers: %v", dec.EnabledRules)
	}
}

// v1 must remain misleadable — the property that makes measuring it
// meaningful. The remaining lie surface is CREDIT ALLOCATION among
// probe-passing rules: a lying history can still decide which genuinely
// task-reducing rule leads the order and deny preference to the
// structurally dominant one. (Search-level harm from this narrower
// surface needs deep constructions; it is measured at the screen, not
// hand-built here.)
func TestV1StillMisleadableByCreditAllocation(t *testing.T) {
	// Task nn(addZ(nn(x))): double-not carries 3 of the 4 reduction
	// steps; add-zero carries 1. The lie credits add-zero as the sole
	// success and double-not as the failure. Both pass the probe, so
	// v1 prefers the lie's choice (add-zero first) and quietly
	// distrusts only the failure claim — the lie still steered the
	// ordering.
	task := nn(addZ(nn(v("x"))))
	catalog := []string{"double-not", "add-zero"}
	rules := []rewrite.Rule{
		admitV1(t, "double-not", nn(v("a")), v("a")),
		admitV1(t, "add-zero", addZ(v("a")), v("a")),
	}
	in := InputV1{
		Input: Input{
			TaskStart: RenderExpr(task), Target: 1, Catalog: catalog,
			History: History{
				{Start: RenderExpr(nn(addZ(nn(v("w"))))), RulesApplied: []string{"add-zero"}, Completed: true},
				{Start: RenderExpr(nn(addZ(nn(v("u"))))), RulesApplied: []string{"double-not"}, Completed: false},
			},
		},
		Task: task, Domain: finite.Domain{Width: 4, Vars: []string{"x"}}, Rules: rules,
	}
	dec := SelectV1(in)
	if dec.EnabledRules[0] != "add-zero" {
		t.Fatalf("a lie among probe-passing rules must still steer the ordering (the misleadability property): %v", dec.EnabledRules)
	}
	for _, p := range dec.Preferences {
		if p.Rule == "double-not" && p.Direction == "avoid" {
			t.Fatalf("a failure-credited rule that demonstrably reduces must not be avoided: %+v", p)
		}
	}
}

// Determinism and version identity.
func TestV1Deterministic(t *testing.T) {
	task := nn(v("x"))
	rules := []rewrite.Rule{admitV1(t, "double-not", nn(v("a")), v("a"))}
	in := InputV1{
		Input: Input{TaskStart: RenderExpr(task), Target: 1, Catalog: []string{"double-not"}},
		Task:  task, Domain: finite.Domain{Width: 4, Vars: []string{"x"}}, Rules: rules,
	}
	a, b := SelectV1(in), SelectV1(in)
	if a.InputHash != b.InputHash || a.SnapshotHash != b.SnapshotHash || len(a.EnabledRules) != len(b.EnabledRules) {
		t.Fatalf("nondeterministic: %+v vs %+v", a, b)
	}
}
