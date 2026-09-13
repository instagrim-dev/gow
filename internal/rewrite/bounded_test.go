package rewrite

import (
	"strings"
	"testing"

	"github.com/instagrim-dev/newf/internal/finite"
)

func growingRule(t *testing.T) Rule {
	return admit(t, "not-intro", finite.Var{Name: "a"},
		finite.Unary{Op: finite.OpNot, X: finite.Unary{Op: finite.OpNot, X: finite.Var{Name: "a"}}}, dom(4, "a"))
}

func TestStateAllowanceStopsBeforeGeneratingFurtherSuccessors(t *testing.T) {
	r := growingRule(t)
	// Each position matches (three candidate positions). With room for
	// one new state, the first candidate is admitted and the second is
	// refused: generation stops there, not at the end of the batch.
	start := finite.Binary{Op: finite.OpAdd, X: finite.Var{Name: "x"}, Y: finite.Var{Name: "x"}}
	res, err := SearchBounded(start, dom(4, "x"), []Rule{r}, nil, 10, Limits{MaxStates: 2})
	if err != nil {
		t.Fatal(err)
	}
	if !res.StateBounded || res.Generated != 2 || res.RuleApplications != 1 {
		t.Fatalf("state allowance must stop incremental generation at the first refused candidate: %+v", res)
	}
	// A cap already met by the start state refuses the very first
	// unseen successor: StateBounded reports that refusal — the set
	// merely being full performs no further admission.
	res, err = SearchBounded(start, dom(4, "x"), []Rule{r}, nil, 10, Limits{MaxStates: 1})
	if err != nil {
		t.Fatal(err)
	}
	if !res.StateBounded || res.Generated != 1 || res.RuleApplications != 1 {
		t.Fatalf("a full allowance must refuse the first unseen successor and stop: %+v", res)
	}
}

func TestOversizeSubstitutionRejectedBeforeInstantiation(t *testing.T) {
	// A compact DAG defines 4095 RHS nodes, and an admitted subject has
	// 4095 logical nodes. Substituting first would normalize over eight
	// million nodes for the very first candidate.
	var rhs finite.Expr = finite.Var{Name: "a"}
	var subject finite.Expr = finite.Var{Name: "x"}
	for i := 0; i < 11; i++ {
		rhs = finite.Binary{Op: finite.OpAnd, X: rhs, Y: rhs}
		subject = finite.Binary{Op: finite.OpAnd, X: subject, Y: subject}
	}
	r := admit(t, "duplicate-idempotently", finite.Var{Name: "a"}, rhs, dom(1, "a"))
	work := &WorkBudget{MaxRuleApplications: 1, MaxCandidates: 1}
	materialized, dropped := 0, 0
	stats := visitSuccessors(subject, r, dom(1, "x"), Limits{MaxTermNodes: finite.MaxExprNodes, Work: work}, func(next finite.Expr, refused bool) bool {
		if refused {
			dropped++
		} else {
			materialized++
		}
		return true
	})
	if materialized != 0 || dropped != 1 || stats.Candidates != 1 || !stats.ResourceBudgetExhausted {
		t.Fatalf("one oversized matched position must be refused before construction: materialized=%d dropped=%d stats=%+v", materialized, dropped, stats)
	}
	if work.Candidates != 1 || work.RuleApplications != 1 {
		t.Fatalf("refusal must preserve exact incurred candidate work: %+v", work)
	}
}

func TestProbeAndSearchSpendTheSameAllowance(t *testing.T) {
	lhs, rhs := doubleNot(dom(4, "a"))
	r := admit(t, "double-not", lhs, rhs, dom(4, "a"))
	x := finite.Var{Name: "x"}
	nnx := finite.Unary{Op: finite.OpNot, X: finite.Unary{Op: finite.OpNot, X: x}}
	work := &WorkBudget{MaxRuleApplications: 2, MaxCandidates: 1}
	lim := Limits{Work: work}
	probe, err := ProbeStrictReductionBounded(nnx, r, dom(4, "x"), nil, lim)
	if err != nil {
		t.Fatal(err)
	}
	if !probe.Reduces || !probe.Complete || probe.Candidates != 1 || probe.RuleApplications != 1 {
		t.Fatalf("unexpected probe: %+v", probe)
	}
	res, err := SearchBounded(nnx, dom(4, "x"), []Rule{r}, nil, 10, lim)
	if err != nil {
		t.Fatal(err)
	}
	if !res.ResourceBudgetExhausted || res.Generated != 0 || res.Best != res.Original || !res.EndpointVerified {
		t.Fatalf("search must retain its independently checked start after the probe spent the candidate allowance: %+v", res)
	}
	if work.Candidates != 1 || work.RuleApplications != 2 {
		t.Fatalf("shared allowance overrun: %+v", work)
	}
}

func TestMalformedAllowanceRefusedBeforeCostCallback(t *testing.T) {
	r := growingRule(t)
	for _, work := range []WorkBudget{
		{MaxRuleApplications: -1}, {MaxCandidates: -1},
		{RuleApplications: 1}, {Candidates: 1}, {Candidates: -1},
	} {
		cost := func(finite.Expr) int64 { t.Fatal("invalid allowance reached cost callback"); return 0 }
		if _, err := SearchBounded(finite.Var{Name: "x"}, dom(4, "x"), []Rule{r}, cost, 10, Limits{Work: &work}); err == nil {
			t.Fatalf("invalid allowance accepted by search: %+v", work)
		}
		if _, err := ProbeStrictReductionBounded(finite.Var{Name: "x"}, r, dom(4, "x"), cost, Limits{Work: &work}); err == nil {
			t.Fatalf("invalid allowance accepted by probe: %+v", work)
		}
	}
}

func TestCustomProbeCostReportsIncompletePruning(t *testing.T) {
	r := growingRule(t)
	cost := func(e finite.Expr) int64 {
		if _, ok := e.(finite.Unary); ok {
			return 0
		}
		return 1
	}
	res, err := ProbeStrictReductionBounded(finite.Var{Name: "x"}, r, dom(4, "x"), cost, Limits{MaxTermNodes: 1})
	if err != nil {
		t.Fatal(err)
	}
	if res.Reduces || res.Complete || !res.TermSizeBounded || res.Candidates != 1 {
		t.Fatalf("a pruned cheaper custom-cost successor cannot support a negative assessment: %+v", res)
	}
	full, err := ProbeStrictReductionBounded(finite.Var{Name: "x"}, r, dom(4, "x"), cost, Limits{})
	if err != nil {
		t.Fatal(err)
	}
	if !full.Reduces || !full.Complete {
		t.Fatalf("the omitted successor really is a custom-cost improvement: %+v", full)
	}
	nodes, err := ProbeStrictReductionBounded(finite.Var{Name: "x"}, r, dom(4, "x"), NodeCount, Limits{MaxTermNodes: 1})
	if err != nil {
		t.Fatal(err)
	}
	if nodes.Reduces || !nodes.Complete || !nodes.TermSizeBounded {
		t.Fatalf("larger terms cannot hide a NodeCount improvement when the bound covers the start: %+v", nodes)
	}
}

func TestLegacyProbeRejectsUnsupportedCostWithoutExecutingIt(t *testing.T) {
	r := growingRule(t)
	defer func() {
		got := recover()
		if got == nil || !strings.Contains(got.(string), "ProbeStrictReductionBounded") {
			t.Fatalf("legacy probe must direct custom costs to explicit assessment API, got %v", got)
		}
	}()
	ProbeStrictReduction(finite.Var{Name: "x"}, r, dom(4, "x"), func(finite.Expr) int64 {
		t.Fatal("unsupported cost must not execute")
		return 0
	})
}

func TestCancellationDuringYieldStopsGenerationAndIndependentReplay(t *testing.T) {
	r := growingRule(t)
	cancel := make(chan struct{})
	calls := 0
	cost := func(e finite.Expr) int64 {
		calls++
		if calls == 2 {
			close(cancel)
		}
		return NodeCount(e)
	}
	start := finite.Binary{Op: finite.OpAdd, X: finite.Var{Name: "x"}, Y: finite.Var{Name: "x"}}
	res, err := SearchBounded(start, dom(4, "x"), []Rule{r}, cost, 100, Limits{Cancel: cancel})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Cancelled || res.Generated != 1 || calls != 2 {
		t.Fatalf("cancellation must stop inside the first expansion: generated=%d calls=%d cancelled=%v", res.Generated, calls, res.Cancelled)
	}
	if res.EndpointVerified || res.Endpoint.Exhaustive || res.Endpoint.AssignmentsChecked != 0 || res.Endpoint.Verdict != finite.VerdictUnresolved {
		t.Fatalf("cancelled endpoint must remain unverified: %+v", res.Endpoint)
	}
}

func TestGeneratedDepthRemainsAnAdmissionBoundary(t *testing.T) {
	r := growingRule(t)
	var start finite.Expr = finite.Var{Name: "x"}
	for i := 0; i < finite.MaxExprDepth; i++ {
		start = finite.Unary{Op: finite.OpNot, X: start}
	}
	res, err := SearchBounded(start, dom(4, "x"), []Rule{r}, nil, 1, Limits{})
	if err != nil {
		t.Fatal(err)
	}
	if !res.TermSizeBounded || res.Generated != finite.MaxExprDepth+1 || res.Best != res.Original {
		t.Fatalf("every too-deep positional growth must be counted and refused: %+v", res)
	}
}
