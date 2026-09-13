package sealedrun

import (
	"sort"
	"testing"

	"github.com/instagrim-dev/newf/internal/screen"
)

// TestImplementerAuthoredV1ScreenRun executes the pre-committed pack
// (commit 8182e13 precedes this run; post-run pack edits are
// inadmissible). It asserts integrity properties only — population
// shape, grid evaluability, gate ineligibility — and LOGS the measured
// conditions. It deliberately does not assert any margin direction: the
// author did not know the outcome when the pack was committed, and this
// test must never be edited to expect one.
func TestImplementerAuthoredV1ScreenRun(t *testing.T) {
	pack := AgentSealedV1()
	if pack.Label != "implementer-authored/v1" {
		t.Fatalf("the label must state the actual evidence grade, got %q", pack.Label)
	}
	if len(pack.Episodes) != 24 {
		t.Fatalf("24 episodes required, got %d", len(pack.Episodes))
	}

	out, traces, err := Run(pack)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if !out.PopulationConforms {
		t.Fatalf("population must conform: %v", out.PopulationDefects)
	}
	if out.GateEligible {
		t.Fatal("gate eligibility must be unreachable regardless of outcome")
	}
	if len(traces) != 24 {
		t.Fatalf("24 attribution traces required, got %d", len(traces))
	}
	for _, tr := range traces {
		if tr.HG.ControllerVersion == "" || tr.H1.ControllerVersion == "" {
			t.Fatalf("episode %s: arm identity missing from trace", tr.EpisodeID)
		}
	}
	logRun(t, "run-1 (budget 500, recorded ceiling effect — inconclusive)", out, traces)

	// Run 2: the calibrated budget. The calibration rule was stated
	// before computation and is arm-blind (CalibrateH0MinBudgets):
	// budget = median of per-episode minimum H0-completing budgets.
	mins, err := CalibrateH0MinBudgets(pack, 500)
	if err != nil {
		t.Fatalf("calibration: %v", err)
	}
	var vals []int
	for _, v := range mins {
		if v > 0 {
			vals = append(vals, v)
		}
	}
	sort.Ints(vals)
	budget := vals[len(vals)/2]
	out2, traces2, err := RunWithBudget(pack, budget)
	if err != nil {
		t.Fatalf("calibrated run: %v", err)
	}
	if out2.GateEligible {
		t.Fatal("gate eligibility must be unreachable regardless of outcome")
	}
	t.Logf("calibrated budget (median of blind-H0 minimums) = %d", budget)
	logRun(t, "run-2 (calibrated)", out2, traces2)
}

func logRun(t *testing.T, name string, out screen.Outcome, traces []EpisodeTrace) {
	t.Helper()
	sums := map[screen.Arm]int{}
	for _, tr := range traces {
		for arm, done := range tr.Completed {
			if done {
				sums[arm]++
			}
		}
	}
	t.Logf("%s — completions: H0=%d H1=%d HG=%d", name, sums[screen.ArmH0], sums[screen.ArmH1], sums[screen.ArmHG])
	for _, c := range out.Conditions {
		t.Logf("  condition (%s): satisfied=%v  %d %s %d  — %s", c.Name, c.Satisfied, c.Left, c.Op, c.Right, c.Detail)
	}
	t.Logf("  ArithmeticSatisfied=%v RuleSatisfied=%v label=%s", out.ArithmeticSatisfied, out.RuleSatisfied, out.EvidenceLabel)
	for _, tr := range traces {
		if tr.Completed[screen.ArmH0] != tr.Completed[screen.ArmH1] || tr.Completed[screen.ArmH1] != tr.Completed[screen.ArmHG] {
			t.Logf("  DIFFERS %-8s H0=%v H1=%v HG=%v", tr.EpisodeID, tr.Completed[screen.ArmH0], tr.Completed[screen.ArmH1], tr.Completed[screen.ArmHG])
		}
	}
}
