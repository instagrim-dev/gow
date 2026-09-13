package shape

import (
	"reflect"
	"strings"
	"testing"

	"github.com/instagrim-dev/newf/internal/finite"
	"github.com/instagrim-dev/newf/internal/rewrite"
)

func boundedSelectorInput(t *testing.T) InputV2 {
	t.Helper()
	task := nn(v("x"))
	return InputV2{
		Input: Input{
			TaskStart: RenderExpr(task), Target: 1, Catalog: []string{"not-intro", "double-not"},
			History: History{{Start: RenderExpr(task), RulesApplied: []string{"double-not"}, Completed: true}},
		},
		Task: task, Domain: finite.Domain{Width: 4, Vars: []string{"x"}},
		Rules: []rewrite.Rule{
			admitRule(t, "not-intro", v("a"), nn(v("a"))),
			admitRule(t, "double-not", nn(v("a")), v("a")),
		},
	}
}

func TestBoundedSelectorChargesSharedAllowance(t *testing.T) {
	in := boundedSelectorInput(t)
	legacy, err := SelectV2(in)
	if err != nil {
		t.Fatal(err)
	}
	work := &rewrite.WorkBudget{MaxRuleApplications: 10, MaxCandidates: 20}
	decision, err := SelectV2Bounded(in, rewrite.Limits{Work: work})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(decision.EnabledRules, legacy.EnabledRules) || !reflect.DeepEqual(decision.Preferences, legacy.Preferences) {
		t.Fatalf("ample budget changed the /2 policy: %+v vs %+v", decision, legacy)
	}
	if decision.ControllerVersion != ControllerVersionV2Bounded || decision.SnapshotHash == legacy.SnapshotHash {
		t.Fatalf("budgeted policy needs its own frozen identity: %+v", decision)
	}
	if work.RuleApplications != decision.ProbeRuleApplications || work.Candidates != decision.ProbeCandidates || work.Candidates == 0 {
		t.Fatalf("probe receipt and shared ledger disagree: %+v vs %+v", decision, work)
	}
	// Search receives the same pointer, so consumed probe work cannot be
	// spent a second time under an unchanged allowance.
	beforeRules, beforeCandidates := work.RuleApplications, work.Candidates
	search, err := rewrite.SearchBounded(in.Task, in.Domain, in.Rules, rewrite.NodeCount, 1, rewrite.Limits{Work: work})
	if err != nil {
		t.Fatal(err)
	}
	if work.RuleApplications != beforeRules+search.RuleApplications || work.Candidates != beforeCandidates+search.Generated {
		t.Fatalf("search failed to accumulate on probe ledger: %+v search %+v", work, search)
	}
}

func TestBoundedSelectorRetainsPartialProbeReceipt(t *testing.T) {
	in := boundedSelectorInput(t)
	work := &rewrite.WorkBudget{MaxRuleApplications: 1, MaxCandidates: 20}
	decision, err := SelectV2Bounded(in, rewrite.Limits{Work: work})
	if err == nil || !strings.Contains(err.Error(), "incomplete") {
		t.Fatalf("exhausted shared allowance must block the selector: %+v, %v", decision, err)
	}
	if decision.InputHash == "" || decision.SnapshotHash == "" || decision.ProbeRuleApplications != 1 || decision.ProbeCandidates == 0 {
		t.Fatalf("partial execution lost identity or incurred costs: %+v", decision)
	}
	if len(decision.EnabledRules) != 0 || len(decision.Preferences) != 0 {
		t.Fatalf("incomplete evidence must not issue ordering or negative rationale: %+v", decision)
	}
	if decision.ProbeCandidates != work.Candidates || decision.ProbeRuleApplications != work.RuleApplications {
		t.Fatalf("receipt does not match consumed allowance: %+v vs %+v", decision, work)
	}
}

func TestBoundedSelectorDoesNotInferNegativesFromPruningOrCancellation(t *testing.T) {
	closed := make(chan struct{})
	close(closed)
	for _, tt := range []struct {
		name   string
		limits rewrite.Limits
		want   string
	}{
		{"cancelled", rewrite.Limits{Cancel: closed}, "cancelled=true"},
		{"pruned", rewrite.Limits{MaxTermNodes: 1}, "term-size-bounded=true"},
		{"no candidate allowance", rewrite.Limits{Work: &rewrite.WorkBudget{MaxRuleApplications: 10}}, "resource-budget-exhausted=true"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			decision, err := SelectV2Bounded(boundedSelectorInput(t), tt.limits)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("want scoped incomplete receipt %q, got %+v, %v", tt.want, decision, err)
			}
			if len(decision.Preferences) != 0 || len(decision.EnabledRules) != 0 {
				t.Fatalf("resource stop became a semantic decision: %+v", decision)
			}
		})
	}
}

func TestTaskProbeIsHistoryIndependentAndBudgetIdentityIsBound(t *testing.T) {
	in := boundedSelectorInput(t)
	first, err := SelectTaskProbeBounded(in, rewrite.Limits{Work: &rewrite.WorkBudget{MaxRuleApplications: 10, MaxCandidates: 20}})
	if err != nil {
		t.Fatal(err)
	}
	in.History = History{{Start: "not(not(z))", RulesApplied: []string{"not-intro"}, Completed: true}}
	second, err := SelectTaskProbeBounded(in, rewrite.Limits{Work: &rewrite.WorkBudget{MaxRuleApplications: 10, MaxCandidates: 20}})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("task-only comparator consumed history: %+v vs %+v", first, second)
	}
	if first.ControllerVersion != TaskProbeControllerVersion || !reflect.DeepEqual(first.EnabledRules, []string{"double-not", "not-intro"}) {
		t.Fatalf("task-only comparator must prefer checked reducers: %+v", first)
	}
	for _, pref := range first.Preferences {
		if len(pref.Evidence) != 0 || pref.Support != 0 {
			t.Fatalf("task-only probe cannot manufacture historical support: %+v", pref)
		}
	}
	changed, err := SelectTaskProbeBounded(in, rewrite.Limits{Work: &rewrite.WorkBudget{MaxRuleApplications: 11, MaxCandidates: 20}})
	if err != nil {
		t.Fatal(err)
	}
	if changed.SnapshotHash == first.SnapshotHash || changed.InputHash == first.InputHash {
		t.Fatal("changed resource allowance must change frozen and input identity")
	}
}

func TestBoundedSelectorRejectsInvalidAllowanceBeforeDecision(t *testing.T) {
	in := boundedSelectorInput(t)
	in.Catalog, in.Rules = nil, nil
	decision, err := SelectV2Bounded(in, rewrite.Limits{Work: &rewrite.WorkBudget{MaxCandidates: -1}})
	if err == nil || !strings.Contains(err.Error(), "invalid shared work budget") {
		t.Fatalf("invalid empty-catalog allowance must be refused: %+v, %v", decision, err)
	}
	if !reflect.DeepEqual(decision, Decision{}) {
		t.Fatalf("invalid allowance is not an executed partial decision: %+v", decision)
	}
}
