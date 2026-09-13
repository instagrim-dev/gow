package sealedrun

import (
	"sort"
	"strings"
	"testing"

	"github.com/instagrim-dev/newf/internal/screen"
	"github.com/instagrim-dev/newf/internal/shape"
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
	mins, unreachable, err := CalibrateH0MinBudgets(pack, 500)
	if err != nil {
		t.Fatalf("calibration: %v", err)
	}
	var vals []int
	for _, v := range mins {
		if v > 0 {
			vals = append(vals, v)
		}
	}
	// Guard the empty case (2026-09-13 external review, corpus
	// correction): an empty set of positive minimums has no median, and
	// unreachable episodes carry no scarcity information.
	if len(vals) == 0 {
		t.Fatalf("no episode yields a positive H0 minimum budget (unreachable within cap: %v); the calibration rule is undefined on this pack", unreachable)
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

	// Run 3: failure-aware-selector diagnostic on the SAME pack —
	// adaptation reuse, labeled; development diagnosis of the repair,
	// never confirmation.
	//
	// IDENTITY NOTE: record 019's retained run-3 figures (14/12/16) were
	// produced by shape-selector/1. This assertion re-runs the pack under
	// shape-selector/2, whose ordering policy is unchanged, so the
	// figures coincide — but they are a /2 measurement and the emitted
	// label says so. The frozen /1 numbers stay attributed to /1.
	out3, traces3, err := RunDiagnosticV2(pack, budget)
	if err != nil {
		t.Fatalf("failure-aware-selector diagnostic: %v", err)
	}
	if out3.GateEligible {
		t.Fatal("gate eligibility must be unreachable regardless of outcome")
	}
	if !strings.Contains(out3.EvidenceLabel, shape.ControllerVersionV2) {
		t.Fatalf("the outcome label must name the controller that produced it: %q", out3.EvidenceLabel)
	}
	if !strings.Contains(out3.EvidenceLabel, "adaptation-reuse-diagnostic") {
		t.Fatalf("adaptation reuse must stay enforced in the label: %q", out3.EvidenceLabel)
	}
	logRun(t, "run-3 ("+shape.ControllerVersionV2+" diagnostic, adaptation reuse)", out3, traces3)
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
	// Cost ledger beside the completion counts: the charged task cost
	// (candidate rewrites materialized, search + selector probes) and the
	// probe component separately, so probe accounting is auditable from
	// the run log without re-running a selector (2026-09-13 external
	// review finding 1 asked for probe work to be visible, not merely
	// charged).
	for _, arm := range []screen.Arm{screen.ArmH0, screen.ArmH1, screen.ArmHG} {
		var generated int
		var probe int64
		for _, tr := range traces {
			generated += tr.Generated[arm]
			probe += tr.ProbeWork[arm]
		}
		rep := out.Arms[arm]
		t.Logf("  cost %-3s task=%d (search-generated=%d + probe=%d)  custodyKnown=%v", arm, rep.TaskCost, generated, probe, rep.CustodyKnown)
	}
	for _, tr := range traces {
		if tr.Completed[screen.ArmH0] != tr.Completed[screen.ArmH1] || tr.Completed[screen.ArmH1] != tr.Completed[screen.ArmHG] {
			t.Logf("  DIFFERS %-8s H0=%v H1=%v HG=%v", tr.EpisodeID, tr.Completed[screen.ArmH0], tr.Completed[screen.ArmH1], tr.Completed[screen.ArmHG])
		}
	}
}
