package sealedrun

import (
	"strings"
	"testing"

	"github.com/instagrim-dev/newf/internal/finite"
	"github.com/instagrim-dev/newf/internal/rewrite"
	"github.com/instagrim-dev/newf/internal/screen"
	"github.com/instagrim-dev/newf/internal/shape"
)

// Runner boundary tests. The pack itself arrives separately (clean-room
// authored); these pin the runner's refusals and the menu's admissibility.
func TestRunnerRefusesEmptyAndUnlabeledPacks(t *testing.T) {
	if _, _, err := Run(Pack{Label: "x"}); err == nil {
		t.Fatal("empty pack must be refused")
	}
	if _, _, err := Run(Pack{Episodes: []EpisodeSpec{{}}}); err == nil {
		t.Fatal("unlabeled pack must be refused")
	}
}

func TestRunnerRefusesOffMenuCatalog(t *testing.T) {
	_, _, err := Run(Pack{Label: "dev", Episodes: []EpisodeSpec{{
		Decl:         screen.Episode{ID: "e1", Stratum: screen.StratumInformative, Family: "f"},
		Start:        finite.Var{Name: "x"},
		Vars:         []string{"x"},
		CatalogNames: []string{"made-up-rule"},
		TargetCost:   1,
	}}})
	if err == nil || !strings.Contains(err.Error(), "not on the admissible menu") {
		t.Fatalf("off-menu rule must be refused by name: %v", err)
	}
}

// Every menu rule must survive its own warrant admission — a menu entry
// that cannot be admitted is a defect here, not at pack time.
func TestEveryMenuRuleIsAdmissible(t *testing.T) {
	for name, def := range Menu() {
		d := finite.Domain{Width: 4, Vars: def.domainVars}
		cert := finite.AssessEquivalence(finite.Binding{Sentence: name, Domain: d}, def.lhs, def.rhs)
		if cert.Verdict != finite.VerdictHoldsOnDomain {
			t.Fatalf("menu rule %q does not hold: %s", name, cert.Reason)
		}
	}
}

// A resource-truncated search must abort the run, not be recorded as an
// arm failing the episode (2026-09-13 self-review of the finding-5
// repair). searchBlockedReason is the seam; BudgetExhausted is
// deliberately excluded because the budget is the declared measurement
// parameter, not a resource accident.
func TestSearchBlockedReasonSeparatesResourceStopsFromBudget(t *testing.T) {
	cases := []struct {
		name string
		res  rewrite.Result
		want string
	}{
		{"clean", rewrite.Result{}, ""},
		{"budget is not a block", rewrite.Result{BudgetExhausted: true, Explored: 2}, ""},
		{"state ceiling blocks", rewrite.Result{StateBounded: true, Explored: 7}, "generated-state ceiling"},
		{"term size blocks", rewrite.Result{TermSizeBounded: true}, "term-size ceiling"},
		{"cancellation blocks", rewrite.Result{Cancelled: true}, "cancelled"},
	}
	for _, c := range cases {
		got := searchBlockedReason(c.res)
		if c.want == "" {
			if got != "" {
				t.Fatalf("%s: expected no block, got %q", c.name, got)
			}
			continue
		}
		if !strings.Contains(got, c.want) {
			t.Fatalf("%s: reason %q must name %q", c.name, got, c.want)
		}
	}
}

// Probe work charges BOTH meters: a probe that matches nothing still
// traverses the task, so charging candidates alone would leave
// non-matching probes free (2026-09-13 self-review).
func TestProbeWorkChargesApplicationsAndCandidates(t *testing.T) {
	if got := probeWork(shape.Decision{ProbeRuleApplications: 5, ProbeCandidates: 0}); got != 5 {
		t.Fatalf("a probe that materialized no candidate still traversed the task; charge %d, got %d", 5, got)
	}
	if got := probeWork(shape.Decision{ProbeRuleApplications: 3, ProbeCandidates: 4}); got != 7 {
		t.Fatalf("both meters are charged: want 7, got %d", got)
	}
	if got := probeWork(shape.Decision{}); got != 0 {
		t.Fatalf("a selector that runs no probe charges nothing, got %d", got)
	}
}

// The WIRING is what needed a guard, not just the helper: the validator
// proved (2026-09-13, finding D2) that deleting the runner's entire
// resource-block abort passed the whole suite, because the only test
// exercised searchBlockedReason in isolation. measurementFor is now the
// single place a search result becomes a scored cell, and these pin its
// behavior; deleting the block inside it fails here.
func TestMeasurementForBlocksResourceStopsAndScoresBudgetStops(t *testing.T) {
	// A resource-truncated search: no completion measurement, even though
	// the cost threshold is nominally met.
	truncated := rewrite.Result{StateBounded: true, Explored: 4, Generated: 9, BestCost: 1, EndpointVerified: true}
	cell := measurementFor(screen.ArmHG, "inf-01", truncated, 1, 5)
	if !cell.MeasurementBlocked {
		t.Fatal("a resource-truncated search must produce a BLOCKED cell, never a scored one")
	}
	if cell.Completed {
		t.Fatal("a blocked cell must not claim completion")
	}
	if cell.BlockedReason == "" {
		t.Fatal("a blocked cell must carry its reason")
	}
	// The blocked cell must be rejected by the scorer it is handed to:
	// this is the end-to-end contract, not a local flag.
	if _, err := screen.Evaluate(
		screen.Design{Episodes: []screen.Episode{{ID: "inf-01", Stratum: screen.StratumInformative, Family: "f"}}, RunsPerCell: 1, EvidenceLabel: "development"},
		[]screen.Execution{cell},
	); err == nil {
		t.Fatal("the screen must refuse a blocked cell")
	}

	// A budget stop is a MEASUREMENT: the budget is the declared
	// parameter, so this cell is scored normally.
	budgetStop := rewrite.Result{BudgetExhausted: true, Explored: 2, Generated: 7, BestCost: 1, EndpointVerified: true}
	scored := measurementFor(screen.ArmHG, "inf-01", budgetStop, 1, 3)
	if scored.MeasurementBlocked {
		t.Fatal("a budget stop is the measurement working as designed, not a block")
	}
	if !scored.Completed {
		t.Fatal("a budget-stopped search meeting the target is a completion")
	}
	if scored.TaskCost != 7+3 {
		t.Fatalf("task cost is search-generated plus probe work: want 10, got %d", scored.TaskCost)
	}
	if scored.CustodyMeasured {
		t.Fatal("this runner meters no custody; it must stay unmeasured, never zero")
	}
}

// Calibration is production code that reads a search result the same way
// the runner does, and the first repair left it unguarded (2026-09-13
// validator finding D1): a truncated search read as "does not complete"
// inverts the binary search and silently shifts the calibrated budget the
// whole screen is measured at. Completion must be monotone in budget;
// truncation breaks that premise, so calibration must refuse.
func TestCalibrationRefusesTruncatedSearchesRatherThanGuessing(t *testing.T) {
	// The guard shares one seam with the runner, so pinning the seam's
	// polarity pins both consumers.
	if searchBlockedReason(rewrite.Result{BudgetExhausted: true}) != "" {
		t.Fatal("a budget stop must not block calibration; it is the parameter being searched over")
	}
	for _, res := range []rewrite.Result{
		{StateBounded: true},
		{TermSizeBounded: true},
		{Cancelled: true},
	} {
		if searchBlockedReason(res) == "" {
			t.Fatalf("a resource stop must block calibration: %+v", res)
		}
	}
	// And the real calibration path still resolves the frozen pack,
	// proving the guard does not fire on legitimate runs.
	mins, unreachable, err := CalibrateH0MinBudgets(AgentSealedV1(), 500)
	if err != nil {
		t.Fatalf("calibration must not refuse a legitimate pack: %v", err)
	}
	if len(mins) != 21 || len(unreachable) != 3 {
		t.Fatalf("frozen calibration shape changed: %d measured, %d unreachable (want 21 and 3)", len(mins), len(unreachable))
	}

	// The guard itself, induced: squeeze the state ceiling so the same
	// pack truncates. Without the guard, truncation reads as "does not
	// complete", the binary search inverts, and a wrong budget is
	// returned SILENTLY — so the required behavior is an error, not a
	// number. This is the assertion that fails when the guard is deleted;
	// pinning the shared seam's polarity above does not cover calibration
	// USING it (2026-09-13 validator finding D1, and a mutation of the
	// first repair to this file showed the seam-only test passing).
	if _, _, err := calibrateH0MinBudgets(AgentSealedV1(), 500, rewrite.Limits{MaxStates: 4}); err == nil {
		t.Fatal("a resource-truncated calibration must refuse; returning a minimum derived from truncated searches silently shifts the budget the whole screen is measured at")
	} else if !strings.Contains(err.Error(), "no longer monotone in budget") {
		t.Fatalf("the refusal must name the broken premise: %v", err)
	}
}
