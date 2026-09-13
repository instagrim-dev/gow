package dryrun

import (
	"testing"

	"github.com/instagrim-dev/newf/internal/finite"
	"github.com/instagrim-dev/newf/internal/rewrite"
	"github.com/instagrim-dev/newf/internal/screen"
	"github.com/instagrim-dev/newf/internal/shape"
)

func admitShapedRules(t *testing.T) map[string]rewrite.Rule {
	t.Helper()
	d1 := finite.Domain{Width: 4, Vars: []string{"a"}}
	adm := func(name string, lhs, rhs finite.Expr) rewrite.Rule {
		t.Helper()
		cert := finite.AssessEquivalence(finite.Binding{Sentence: name, Domain: d1}, lhs, rhs)
		rule, defects := rewrite.AdmitRule(name, cert, lhs, rhs, d1)
		if len(defects) > 0 {
			t.Fatalf("rule %q refused: %v", name, defects)
		}
		return rule
	}
	return map[string]rewrite.Rule{
		"double-not": adm("double-not", nn(v("a")), v("a")),
		"add-zero":   adm("add-zero", addZ(v("a")), v("a")),
		"not-intro":  adm("not-intro", v("a"), nn(v("a"))),
	}
}

func orderRules(pool map[string]rewrite.Rule, names []string) []rewrite.Rule {
	out := make([]rewrite.Rule, 0, len(names))
	for _, n := range names {
		out = append(out, pool[n])
	}
	return out
}

// TestShapedDevelopmentDryRun runs the three-arm loop with the real v0
// selector and asserts (1) the P0 identity discipline on every HG
// decision, (2) arm-level outcomes consistent with the designed
// construction, and (3) the screen's positive arithmetic path on a
// conforming population — while the outcome remains gate-ineligible and
// the designed margin licenses nothing.
func TestShapedDevelopmentDryRun(t *testing.T) {
	pool := admitShapedRules(t)
	eps := shapedEpisodes()
	searchDomain := finite.Domain{Width: 4, Vars: []string{"x", "y", "z"}}

	decls := make([]screen.Episode, 0, len(eps))
	for _, ep := range eps {
		decls = append(decls, ep.decl)
	}

	var execs []screen.Execution
	sums := map[screen.Arm]int{}
	snapshot := shape.SnapshotHash()
	for _, ep := range eps {
		// Arm orderings: H0 blind, H1 ungated frequency, HG selector.
		orders := map[screen.Arm][]string{
			screen.ArmH0: ep.catalog,
			screen.ArmH1: h1Order(ep.catalog, ep.history),
		}
		dec := shape.Select(shape.Input{
			TaskStart: shape.RenderExpr(ep.start),
			Target:    ep.target,
			Catalog:   ep.catalog,
			History:   ep.history,
		})
		// P0 identity discipline on every decision.
		if dec.ControllerVersion != shape.ControllerVersion || dec.SnapshotHash != snapshot {
			t.Fatalf("episode %s: controller identity drifted: %+v", ep.decl.ID, dec)
		}
		if dec.InputHash == "" || len(dec.EnabledRules) != len(ep.catalog) {
			t.Fatalf("episode %s: decision trace incomplete: %+v", ep.decl.ID, dec)
		}
		orders[screen.ArmHG] = dec.EnabledRules

		for _, arm := range []screen.Arm{screen.ArmH0, screen.ArmH1, screen.ArmHG} {
			res, err := rewrite.Search(ep.start, searchDomain, orderRules(pool, orders[arm]), rewrite.NodeCount, shapedBudget)
			if err != nil {
				t.Fatalf("episode %s arm %s: %v", ep.decl.ID, arm, err)
			}
			completed := res.BestCost <= ep.target && res.EndpointVerified
			if completed {
				sums[arm]++
			}
			execs = append(execs, screen.Execution{
				Arm: arm, EpisodeID: ep.decl.ID, Run: 1,
				Completed: completed,
				TaskCost:  int64(res.Explored),
			})
		}
	}

	// Designed arm-level expectations (constructed, not discovered):
	// informative episodes separate HG (gated) from H1 (ungated) and H0
	// (blind); misleading episodes hurt both history arms; low-value
	// episodes equalize everyone.
	if sums[screen.ArmHG] <= sums[screen.ArmH1] {
		t.Fatalf("the designed construction must separate HG (%d) from H1 (%d); if it cannot, the margin machinery is broken", sums[screen.ArmHG], sums[screen.ArmH1])
	}
	if sums[screen.ArmHG] < sums[screen.ArmH0] {
		t.Fatalf("designed construction violated: HG %d < H0 %d", sums[screen.ArmHG], sums[screen.ArmH0])
	}

	out, err := screen.Evaluate(screen.Design{
		Episodes:      decls,
		RunsPerCell:   1, // deterministic arms; disclosed
		EvidenceLabel: "development-dry-run-shaped",
	}, execs)
	if err != nil {
		t.Fatalf("grid must evaluate: %v", err)
	}
	if !out.PopulationConforms {
		t.Fatalf("population must conform: %v", out.PopulationDefects)
	}
	// The positive arithmetic path: on this DESIGNED margin the rule's
	// conditions all compute and (given the construction) hold — which
	// demonstrates the machinery, not shaping value.
	if !out.ArithmeticSatisfied || !out.RuleSatisfied {
		t.Fatalf("the designed margin should satisfy the arithmetic; conditions: %+v", out.Conditions)
	}
	// And the boundary that matters: a satisfied rule on development
	// data still cannot touch the gate.
	if out.GateEligible {
		t.Fatal("gate eligibility must remain unreachable from a development dry run, even with the rule satisfied")
	}
}
