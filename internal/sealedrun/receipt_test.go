package sealedrun

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/instagrim-dev/newf/internal/finite"
	"github.com/instagrim-dev/newf/internal/rewrite"
	"github.com/instagrim-dev/newf/internal/screen"
	"github.com/instagrim-dev/newf/internal/shape"
)

func receiptEpisode(id string) EpisodeSpec {
	return EpisodeSpec{
		Decl:  screen.Episode{ID: id, Stratum: screen.StratumInformative, Family: "receipt"},
		Start: finite.Unary{Op: finite.OpNot, X: finite.Unary{Op: finite.OpNot, X: finite.Var{Name: "x"}}},
		Vars:  []string{"x"}, CatalogNames: []string{"double-not"}, TargetCost: 1,
	}
}

func TestLegacyConfirmatoryCannotGrantFreshnessFromPackLabel(t *testing.T) {
	pack := AgentSealedV1()        // Already exposed; no custody or freshness manifest.
	pack.Label = "agent-sealed/v1" // A stronger caller label proves nothing.
	pack.Provenance = "caller asserts clean-room freshness"
	pack.Episodes[0].Decl.ID = "renamed-exposed-task"
	pack.Episodes[0].History = shape.History{}
	out, traces, err := RunConfirmatoryV2(pack, 2)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.EvidenceLabel, "freshness-unverified") {
		t.Fatalf("legacy routing must explicitly withhold freshness, got %q", out.EvidenceLabel)
	}
	if strings.Contains(out.EvidenceLabel, "agent-sealed") || out.GateEligible {
		t.Fatalf("an unchecked pack label must not grant an evidence tier: %+v", out)
	}
	if !strings.Contains(out.EvidenceLabel, "adaptation-reuse-diagnostic") {
		t.Fatalf("renaming an exposed task must not remove the reuse marker: %q", out.EvidenceLabel)
	}
	if traces[0].SourcePackLabel != pack.Label || traces[0].EvidenceLabel != out.EvidenceLabel {
		t.Fatalf("source claim must survive separately from evidence presentation: %+v", traces[0])
	}
}

func TestRunnerRetainsTracesWhenScoringRefusesDeclaration(t *testing.T) {
	ep := receiptEpisode("bad-stratum")
	ep.Decl.Stratum = "undeclared"
	out, traces, err := RunWithBudget(Pack{Label: "development", Episodes: []EpisodeSpec{ep}}, 2)
	if err == nil || !strings.Contains(err.Error(), "undeclared stratum") {
		t.Fatalf("scoring must refuse the invalid declaration: %v", err)
	}
	if len(traces) != 1 || len(traces[0].Completed) != 3 {
		t.Fatalf("scoring refusal erased the executed cells: %+v", traces)
	}
	if out.EvidenceLabel != "development" || out.Arms != nil || out.GateEligible {
		t.Fatalf("failed scoring must retain attribution but produce no score: %+v", out)
	}
	assertCheckedReceipt(t, traces[0], screen.ArmH0)
}

func TestRunnerRetainsEarlierEpisodeWhenLaterEpisodeIsInvalid(t *testing.T) {
	bad := receiptEpisode("bad-domain")
	bad.Vars = nil
	_, traces, err := RunWithBudget(Pack{Label: "development", Episodes: []EpisodeSpec{receiptEpisode("done"), bad}}, 2)
	if err == nil {
		t.Fatal("invalid second episode must refuse execution")
	}
	if len(traces) != 2 || len(traces[0].Completed) != 3 {
		t.Fatalf("later refusal erased earlier completed cells: %+v", traces)
	}
	assertCheckedReceipt(t, traces[0], screen.ArmH0)
	if traces[1].Error == "" || len(traces[1].Cells) != 3 {
		t.Fatalf("invalid episode needs an explicit receipt: %+v", traces[1])
	}
	for _, cell := range traces[1].Cells {
		if cell.SearchStarted || !cell.Execution.MeasurementBlocked || cell.Execution.Completed {
			t.Fatalf("unstarted cell must never read as a measured miss: %+v", cell)
		}
	}
}

func TestRunnerRefusesMalformedExpressionBeforeRendering(t *testing.T) {
	bad := receiptEpisode("bad-expression")
	bad.Start = finite.Unary{Op: finite.OpNot} // Missing child cannot be rendered.
	_, traces, err := RunWithBudget(Pack{Label: "development", Episodes: []EpisodeSpec{receiptEpisode("done"), bad}}, 2)
	if err == nil || !strings.Contains(err.Error(), "start expression is invalid") || len(traces) != 2 {
		t.Fatalf("invalid expression must refuse before rendering and retain the receipt: %v %+v", err, traces)
	}
	assertCheckedReceipt(t, traces[0], screen.ArmH0)
}

func TestRunnerRetainsBlockedCellsAndCheckedCandidate(t *testing.T) {
	ep := receiptEpisode("truncated")
	ep.CatalogNames = []string{"double-not", "not-intro"}
	pack := Pack{Label: "development", Episodes: []EpisodeSpec{receiptEpisode("completed"), ep}}
	// MaxStates 2 admits {not(not(x)), x} and then REFUSES the unseen
	// not-intro successor: a real truncation under refusal semantics
	// (the reachable set exceeds the cap; merely filling it would not).
	out, traces, err := runWithSelectorBounded(pack, 2, hgV0, pack.Label, rewrite.Limits{MaxStates: 2})
	if err == nil || !strings.Contains(err.Error(), "BLOCKED measurement") {
		t.Fatalf("real resource-truncated search must reach scorer refusal: %v", err)
	}
	if len(traces) != 2 || out.Arms != nil || out.GateEligible {
		t.Fatalf("comparison must be unscored with both receipts retained: out=%+v traces=%+v", out, traces)
	}
	assertCheckedReceipt(t, traces[0], screen.ArmH0)
	for _, arm := range []screen.Arm{screen.ArmH0, screen.ArmH1, screen.ArmHG} {
		cell := traces[1].Cells[arm]
		if !cell.SearchStarted || !cell.Execution.MeasurementBlocked || cell.Execution.Completed || !cell.Result.StateBounded {
			t.Fatalf("truncated cell lost its stop semantics: %+v", cell)
		}
		if cell.Execution.TaskCost <= 0 || cell.Execution.TaskCost != int64(cell.Result.Generated)+traces[1].ProbeWork[arm] {
			t.Fatalf("blocked cell lost incurred work: %+v", cell)
		}
		if !cell.Result.EndpointVerified || cell.Result.Best != "x" || len(cell.Result.Steps) == 0 || cell.Result.Endpoint.Verdict != finite.VerdictHoldsOnDomain {
			t.Fatalf("checked candidate must survive even though truncation blocks scoring: %+v", cell.Result)
		}
	}
	// The receipt is a durable record, including the independent certificate.
	data, err := json.Marshal(traces)
	if err != nil {
		t.Fatal(err)
	}
	var restored []EpisodeTrace
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatal(err)
	}
	if restored[1].Cells[screen.ArmH0].Result.Endpoint.Verdict != finite.VerdictHoldsOnDomain || restored[1].Cells[screen.ArmH0].Execution.TaskCost == 0 {
		t.Fatal("serializing the receipt lost the checked candidate or cost")
	}
}

func TestRunnerRetainsSelectorWorkWhenSearchRefuses(t *testing.T) {
	selector := func(shape.Input, finite.Expr, finite.Domain, []rewrite.Rule) (shape.Decision, error) {
		return shape.Decision{EnabledRules: []string{"double-not"}, ProbeRuleApplications: 3, ProbeCandidates: 4}, nil
	}
	_, traces, err := runWithSelector(Pack{Label: "development", Episodes: []EpisodeSpec{receiptEpisode("refused")}}, -1, selector, "development")
	if err == nil || len(traces) != 1 {
		t.Fatalf("invalid search budget must refuse but return receipts: %v %+v", err, traces)
	}
	h0, hg := traces[0].Cells[screen.ArmH0], traces[0].Cells[screen.ArmHG]
	if !h0.SearchStarted || h0.Error == "" || !h0.Execution.MeasurementBlocked {
		t.Fatalf("refused search must have an explicit cell: %+v", h0)
	}
	if hg.SearchStarted || hg.Execution.TaskCost != 7 || traces[0].ProbeWork[screen.ArmHG] != 7 || !hg.Execution.MeasurementBlocked {
		t.Fatalf("selector work was incurred before search refusal and must survive: %+v", hg)
	}
}

func assertCheckedReceipt(t *testing.T, trace EpisodeTrace, arm screen.Arm) {
	t.Helper()
	cell := trace.Cells[arm]
	if !cell.SearchStarted || !cell.Execution.Completed || cell.Execution.MeasurementBlocked || cell.Execution.TaskCost <= 0 || !cell.Result.EndpointVerified || cell.Result.Best != "x" {
		t.Fatalf("completed execution receipt lost its checked candidate or cost: %+v", cell)
	}
}
