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

func admitRule(t *testing.T, name string, lhs, rhs finite.Expr) rewrite.Rule {
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
func TestSelectorNeutralizesBlatantLie(t *testing.T) {
	task := nn(addZ(v("x")))
	catalog := []string{"double-not", "add-zero", "not-intro"}
	rules := []rewrite.Rule{
		admitRule(t, "double-not", nn(v("a")), v("a")),
		admitRule(t, "add-zero", addZ(v("a")), v("a")),
		admitRule(t, "not-intro", v("a"), nn(v("a"))),
	}
	in := InputV2{
		Input: Input{
			TaskStart: RenderExpr(task), Target: 1, Catalog: catalog,
			History: History{
				{Start: RenderExpr(nn(addZ(v("w")))), RulesApplied: []string{"not-intro"}, Completed: true},
				{Start: RenderExpr(nn(addZ(v("u")))), RulesApplied: []string{"double-not", "add-zero"}, Completed: false},
			},
		},
		Task: task, Domain: finite.Domain{Width: 4, Vars: []string{"x"}}, Rules: rules,
	}
	dec, err := SelectV2(in)
	if err != nil {
		t.Fatalf("SelectV2: %v", err)
	}
	if dec.EnabledRules[len(dec.EnabledRules)-1] != "not-intro" {
		t.Fatalf("the lying success claim must be demoted last: %v", dec.EnabledRules)
	}
	if dec.EnabledRules[0] == "not-intro" {
		t.Fatalf("v1 must not front-load the lie: %v", dec.EnabledRules)
	}
	// The demotion must be recorded with a rationale stating the probe's
	// actual claim scope (2026-09-13 external review finding 2): a
	// one-step negative from the current start — never "can never reduce
	// this task" and never a verdict that the history itself is false.
	withheld := 0
	for _, p := range dec.Preferences {
		if strings.Contains(p.Rationale, "can never reduce") {
			t.Fatalf("rationale exceeds the one-step probe's claim scope: %q", p.Rationale)
		}
		if strings.Contains(p.Rationale, "no one-step strict NodeCount decrease from the current start") {
			withheld++
		}
	}
	if withheld == 0 {
		t.Fatalf("the one-step demotion must be recorded with its scoped rationale: %+v", dec.Preferences)
	}
	if dec.ControllerVersion != ControllerVersionV2 {
		t.Fatalf("wrong version: %s", dec.ControllerVersion)
	}
}

// Informative history must still help: credited reducers that pass the
// probe are preferred, exactly as in v0's happy path.
func TestSelectorInformativeHistoryStillPrefersReducers(t *testing.T) {
	task := nn(addZ(nn(v("x"))))
	catalog := []string{"not-intro", "double-not", "add-zero"} // bad neutral order
	rules := []rewrite.Rule{
		admitRule(t, "double-not", nn(v("a")), v("a")),
		admitRule(t, "add-zero", addZ(v("a")), v("a")),
		admitRule(t, "not-intro", v("a"), nn(v("a"))),
	}
	in := InputV2{
		Input: Input{
			TaskStart: RenderExpr(task), Target: 1, Catalog: catalog,
			History: History{
				{Start: RenderExpr(nn(addZ(nn(v("y"))))), RulesApplied: []string{"double-not", "add-zero"}, Completed: true},
			},
		},
		Task: task, Domain: finite.Domain{Width: 4, Vars: []string{"x"}}, Rules: rules,
	}
	dec, err := SelectV2(in)
	if err != nil {
		t.Fatalf("SelectV2: %v", err)
	}
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
func TestSelectorStillMisleadableByCreditAllocation(t *testing.T) {
	// Task nn(addZ(nn(x))): double-not carries 3 of the 4 reduction
	// steps; add-zero carries 1. The lie credits add-zero as the sole
	// success and double-not as the failure. Both pass the probe, so
	// v1 prefers the lie's choice (add-zero first) and quietly
	// distrusts only the failure claim — the lie still steered the
	// ordering.
	task := nn(addZ(nn(v("x"))))
	catalog := []string{"double-not", "add-zero"}
	rules := []rewrite.Rule{
		admitRule(t, "double-not", nn(v("a")), v("a")),
		admitRule(t, "add-zero", addZ(v("a")), v("a")),
	}
	in := InputV2{
		Input: Input{
			TaskStart: RenderExpr(task), Target: 1, Catalog: catalog,
			History: History{
				{Start: RenderExpr(nn(addZ(nn(v("w"))))), RulesApplied: []string{"add-zero"}, Completed: true},
				{Start: RenderExpr(nn(addZ(nn(v("u"))))), RulesApplied: []string{"double-not"}, Completed: false},
			},
		},
		Task: task, Domain: finite.Domain{Width: 4, Vars: []string{"x"}}, Rules: rules,
	}
	dec, err := SelectV2(in)
	if err != nil {
		t.Fatalf("SelectV2: %v", err)
	}
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
func TestSelectorDeterministic(t *testing.T) {
	task := nn(v("x"))
	rules := []rewrite.Rule{admitRule(t, "double-not", nn(v("a")), v("a"))}
	in := InputV2{
		Input: Input{TaskStart: RenderExpr(task), Target: 1, Catalog: []string{"double-not"}},
		Task:  task, Domain: finite.Domain{Width: 4, Vars: []string{"x"}}, Rules: rules,
	}
	a, errA := SelectV2(in)
	b, errB := SelectV2(in)
	if errA != nil || errB != nil {
		t.Fatalf("SelectV2: %v / %v", errA, errB)
	}
	if a.InputHash != b.InputHash || a.SnapshotHash != b.SnapshotHash || len(a.EnabledRules) != len(b.EnabledRules) {
		t.Fatalf("nondeterministic: %+v vs %+v", a, b)
	}
}

// Identity contract (2026-09-13 external review finding 3): the decision
// hash must move when decision-relevant inputs move, and inputs whose
// declared rendering disagrees with the actual task must be refused —
// never silently probed against a task other than the one hashed.
func TestSelectorIdentityContract(t *testing.T) {
	task := nn(v("x"))
	domain := finite.Domain{Width: 4, Vars: []string{"x"}}
	doubleNot := admitRule(t, "double-not", nn(v("a")), v("a"))
	base := InputV2{
		Input: Input{TaskStart: RenderExpr(task), Target: 1, Catalog: []string{"double-not"}},
		Task:  task, Domain: domain, Rules: []rewrite.Rule{doubleNot},
	}
	baseDec, err := SelectV2(base)
	if err != nil {
		t.Fatalf("SelectV2(base): %v", err)
	}

	// (i) Task/rendering mismatch is refused: keep TaskStart, swap the
	// actual expression (the review's counterexample: not(not(x)) -> x
	// flips the double-not probe while a name-only hash is unchanged).
	mismatch := base
	mismatch.Task = v("x")
	if _, err := SelectV2(mismatch); err == nil {
		t.Fatal("a task expression disagreeing with its declared rendering must be refused")
	}

	// (ii) Rule content is bound: replace the rule body under the same
	// name (a valid admitted rule, different rewrite). The hash must
	// move, because the probe result may move.
	swapped := base
	swapped.Rules = []rewrite.Rule{admitRule(t, "double-not", addZ(v("a")), v("a"))}
	swappedDec, err := SelectV2(swapped)
	if err != nil {
		t.Fatalf("SelectV2(swapped): %v", err)
	}
	if swappedDec.InputHash == baseDec.InputHash {
		t.Fatal("replacing a rule's content under the same name must change the input hash; names alone are not an identity")
	}

	// (iii) The task is bound by REFUSAL, not by a second hashed copy of
	// its rendering. This is the discriminating form: the review's
	// counterexample (hold TaskStart, swap the actual task) is defeated
	// by (i); a subtest that changes the task AND its rendering, then
	// asserts the hash moved, passes on the unfixed code too — Input
	// already contained TaskStart before the repair — so it protected
	// nothing (2026-09-13 self-review). What is actually guaranteed is
	// the invariant below: every ACCEPTED input has Render(Task) equal to
	// TaskStart, so the hashed Input pins the probed expression.
	accepted := base
	if got := RenderExpr(accepted.Task); got != accepted.Input.TaskStart {
		t.Fatalf("accepted input must satisfy Render(Task) == TaskStart; got %q vs %q", got, accepted.Input.TaskStart)
	}
	// And the pinning is load-bearing: a different task under its own
	// correct rendering is a different hashed Input.
	other := base
	other.Task = nn(nn(v("x")))
	other.TaskStart = RenderExpr(other.Task)
	otherDec, err := SelectV2(other)
	if err != nil {
		t.Fatalf("SelectV2(other): %v", err)
	}
	if otherDec.InputHash == baseDec.InputHash {
		t.Fatal("changing the task must change the input hash")
	}

	// (iv) Conflicting duplicate rule names are refused; exact
	// duplicates are tolerated.
	conflicting := base
	conflicting.Rules = []rewrite.Rule{doubleNot, admitRule(t, "double-not", addZ(v("a")), v("a"))}
	if _, err := SelectV2(conflicting); err == nil {
		t.Fatal("two rules sharing a name with different content must be refused")
	}
	duplicated := base
	duplicated.Rules = []rewrite.Rule{doubleNot, doubleNot}
	if _, err := SelectV2(duplicated); err != nil {
		t.Fatalf("an exact duplicate rule is harmless and must not be refused: %v", err)
	}
}

// Probe metering (2026-09-13 external review finding 1): the probes v1
// runs before the budgeted search are work, and the decision must carry
// their meter so a runner can charge them.
func TestSelectorMetersProbeWork(t *testing.T) {
	task := nn(addZ(v("x")))
	in := InputV2{
		Input: Input{
			TaskStart: RenderExpr(task), Target: 1,
			Catalog: []string{"double-not", "add-zero"},
		},
		Task: task, Domain: finite.Domain{Width: 4, Vars: []string{"x"}},
		Rules: []rewrite.Rule{
			admitRule(t, "double-not", nn(v("a")), v("a")),
			admitRule(t, "add-zero", addZ(v("a")), v("a")),
		},
	}
	dec, err := SelectV2(in)
	if err != nil {
		t.Fatalf("SelectV2: %v", err)
	}
	if dec.ProbeRuleApplications != 2 {
		t.Fatalf("two catalog rules probed, meter says %d", dec.ProbeRuleApplications)
	}
	if dec.ProbeCandidates < 1 {
		t.Fatalf("the probes materialized candidate rewrites; the meter must not report %d", dec.ProbeCandidates)
	}
	// The comparator runs no probe and must meter zero.
	h1 := SelectUngatedFrequency(in.Input)
	if h1.ProbeRuleApplications != 0 || h1.ProbeCandidates != 0 {
		t.Fatalf("H1 runs no probe; meter must be zero: %+v", h1)
	}
}
