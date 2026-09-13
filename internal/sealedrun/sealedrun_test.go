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
