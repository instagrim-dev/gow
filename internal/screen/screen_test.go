// Synthetic decision-arithmetic tests: score records sitting exactly on
// and just below each spending-rule threshold. Synthetic outcomes validate
// the procedure; none of them satisfies, or could satisfy, the investment
// gate (every design here is development-labeled, and the Outcome says so).
package screen

import (
	"fmt"
	"math"
	"strings"
	"testing"
)

const r = 3

// episodes24 is the roadmap population: 12 informative (two families),
// 6 low-value, 6 misleading.
func episodes24() []Episode {
	var eps []Episode
	for i := 1; i <= 12; i++ {
		fam := "fam-A"
		if i > 6 {
			fam = "fam-B"
		}
		eps = append(eps, Episode{ID: fmt.Sprintf("inf-%02d", i), Stratum: StratumInformative, Family: fam})
	}
	for i := 1; i <= 6; i++ {
		eps = append(eps, Episode{ID: fmt.Sprintf("low-%02d", i), Stratum: StratumLowValue, Family: "fam-C"})
	}
	for i := 1; i <= 6; i++ {
		eps = append(eps, Episode{ID: fmt.Sprintf("mis-%02d", i), Stratum: StratumMisleading, Family: "fam-D"})
	}
	return eps
}

// grid builds a complete arm×episode×run grid from per-arm completion
// counts (episode → completions in 0..r). Task cost 1 per execution;
// custody cost as supplied.
func grid(eps []Episode, counts map[Arm]map[string]int64, custodyPerExec int64) []Execution {
	var execs []Execution
	for _, arm := range []Arm{ArmH0, ArmH1, ArmHG} {
		for _, ep := range eps {
			c := counts[arm][ep.ID]
			for run := 1; run <= r; run++ {
				execs = append(execs, Execution{
					Arm: arm, EpisodeID: ep.ID, Run: run,
					Completed:   int64(run) <= c,
					TaskCost:    1,
					CustodyCost: custodyPerExec,
				})
			}
		}
	}
	return execs
}

// baseCounts puts every condition exactly ON its threshold:
//
//	H0 = 0 total; H1 = 9 (inf-01:3, low-01..06:1); HG = 18
//	(b) 18 >= 9 + 3·3 = 18   exactly
//	(c) H1 lowmis 6 − HG lowmis 3 = 3 <= 1·3   exactly
//	(d) HG>H1 in fam-A (inf-02,03) and fam-B (inf-07,08) = 2 families, exactly
//	(e) 18 >= 0
func baseCounts() map[Arm]map[string]int64 {
	return map[Arm]map[string]int64{
		ArmH0: {},
		ArmH1: {"inf-01": 3, "low-01": 1, "low-02": 1, "low-03": 1, "low-04": 1, "low-05": 1, "low-06": 1},
		ArmHG: {"inf-01": 3, "inf-02": 3, "inf-03": 3, "inf-07": 3, "inf-08": 3, "low-01": 1, "low-02": 1, "low-03": 1},
	}
}

func devDesign() Design {
	return Design{Episodes: episodes24(), RunsPerCell: r, EvidenceLabel: "development"}
}

func condition(t *testing.T, out Outcome, name string) Condition {
	t.Helper()
	for _, c := range out.Conditions {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("condition %q missing from outcome", name)
	return Condition{}
}

// Exactly on every threshold: the rule is satisfied, the grid is 24×3×3 =
// 216 executions, and the outcome still refuses gate eligibility.
func TestExactlyOnThresholdsSatisfiesRuleButNotGate(t *testing.T) {
	execs := grid(episodes24(), baseCounts(), 0)
	if len(execs) != 24*3*r {
		t.Fatalf("grid size %d, want %d", len(execs), 24*3*r)
	}
	out, err := Evaluate(devDesign(), execs)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range out.Conditions {
		if !c.Satisfied {
			t.Fatalf("condition %s should sit exactly on its threshold: %+v", c.Name, c)
		}
	}
	if !out.RuleSatisfied {
		t.Fatal("rule should be satisfied exactly on thresholds")
	}
	if !out.PopulationConforms {
		t.Fatalf("the 12/6/6 population conforms: %v", out.PopulationDefects)
	}
	if out.GateEligible {
		t.Fatal("no outcome from this package is ever gate-eligible")
	}
	found := false
	for _, n := range out.Notes {
		if strings.Contains(n, "gate eligibility cannot be produced here") {
			found = true
		}
	}
	if !found {
		t.Fatalf("outcome must state its own ineligibility: %v", out.Notes)
	}
	if b := condition(t, out, "b"); b.Left != 18 || b.Right != 18 {
		t.Fatalf("(b) exact integers wrong: %+v", b)
	}
}

// One completion below the (b) margin: 17 < 9 + 3·3.
func TestOneBelowMarginFailsB(t *testing.T) {
	counts := baseCounts()
	counts[ArmHG]["inf-08"] = 2 // 18 → 17
	out, err := Evaluate(devDesign(), grid(episodes24(), counts, 0))
	if err != nil {
		t.Fatal(err)
	}
	if b := condition(t, out, "b"); b.Satisfied || b.Left != 17 {
		t.Fatalf("(b) must fail at 17 vs 18: %+v", b)
	}
	if out.RuleSatisfied {
		t.Fatal("rule must fail when (b) fails")
	}
}

// One extra low-value loss fails (c) even when totals still clear (b):
// HG drops low-03 (lowmis diff 4 > 3) and compensates on inf-04.
func TestOneExtraLowValueLossFailsCIsolated(t *testing.T) {
	counts := baseCounts()
	delete(counts[ArmHG], "low-03") // lowmis HG 3 → 2
	counts[ArmHG]["inf-04"] = 1     // total stays 18
	out, err := Evaluate(devDesign(), grid(episodes24(), counts, 0))
	if err != nil {
		t.Fatal(err)
	}
	if b := condition(t, out, "b"); !b.Satisfied {
		t.Fatalf("(b) should still pass at 18: %+v", b)
	}
	if c := condition(t, out, "c"); c.Satisfied || c.Left != 4 || c.Right != 3 {
		t.Fatalf("(c) must fail at 4 vs 3: %+v", c)
	}
	if out.RuleSatisfied {
		t.Fatal("rule must fail when (c) fails")
	}
}

// Wins concentrated in one construction family fail (d) despite an intact
// margin: same totals, all HG advantages moved into fam-A.
func TestSingleFamilyConcentrationFailsD(t *testing.T) {
	counts := baseCounts()
	delete(counts[ArmHG], "inf-07")
	delete(counts[ArmHG], "inf-08")
	counts[ArmHG]["inf-04"] = 3
	counts[ArmHG]["inf-05"] = 3 // total back to 18, all wins in fam-A
	out, err := Evaluate(devDesign(), grid(episodes24(), counts, 0))
	if err != nil {
		t.Fatal(err)
	}
	if b := condition(t, out, "b"); !b.Satisfied {
		t.Fatalf("(b) should pass: %+v", b)
	}
	if d := condition(t, out, "d"); d.Satisfied || d.Left != 1 {
		t.Fatalf("(d) must fail with one family: %+v", d)
	}
}

// H0 beating HG fails (e) — the 0.3.0 H0 guard — even though HG still
// out-margins H1.
func TestH0BeatingHGFailsE(t *testing.T) {
	counts := baseCounts()
	counts[ArmH0] = map[string]int64{
		"inf-01": 3, "inf-02": 3, "inf-03": 3, "inf-04": 3, "inf-05": 3, "inf-06": 3, // 18
		"inf-07": 1, // 19 > HG's 18
	}
	out, err := Evaluate(devDesign(), grid(episodes24(), counts, 0))
	if err != nil {
		t.Fatal(err)
	}
	if b := condition(t, out, "b"); !b.Satisfied {
		t.Fatalf("(b) compares HG to H1 and should still pass: %+v", b)
	}
	if e := condition(t, out, "e"); e.Satisfied || e.Left != 18 || e.Right != 19 {
		t.Fatalf("(e) must fail at 18 vs 19: %+v", e)
	}
	if out.RuleSatisfied {
		t.Fatal("the H0 guard must fail the rule")
	}
}

// A single certified-invalid result fails (a) regardless of completions.
func TestSingleInvalidCertificationFailsA(t *testing.T) {
	execs := grid(episodes24(), baseCounts(), 0)
	execs[0].InvalidCertified = true
	out, err := Evaluate(devDesign(), execs)
	if err != nil {
		t.Fatal(err)
	}
	if a := condition(t, out, "a"); a.Satisfied {
		t.Fatalf("(a) must fail with one invalid certification: %+v", a)
	}
	if out.RuleSatisfied {
		t.Fatal("rule must fail when (a) fails")
	}
}

// An incomplete grid yields no decision at all — not a partial one.
func TestIncompleteGridYieldsNoDecision(t *testing.T) {
	execs := grid(episodes24(), baseCounts(), 0)
	if _, err := Evaluate(devDesign(), execs[:len(execs)-1]); err == nil {
		t.Fatal("a missing cell must be an error, not a decision")
	} else if !strings.Contains(err.Error(), "incomplete grid") {
		t.Fatalf("error must name the incompleteness: %v", err)
	}
}

// A duplicate cell is malformed input.
func TestDuplicateCellYieldsNoDecision(t *testing.T) {
	execs := grid(episodes24(), baseCounts(), 0)
	execs[1] = execs[0]
	if _, err := Evaluate(devDesign(), execs); err == nil {
		t.Fatal("a duplicate cell must be an error")
	}
}

// Capability pass with unaffordable custody: the rule is satisfied, the
// custody figure dominates, and BOTH facts are on the outcome — the
// capability answer is not netted against the cost answer.
func TestCapabilityPassWithUnaffordableCustody(t *testing.T) {
	execs := grid(episodes24(), baseCounts(), 50) // custody 50× task cost per execution
	out, err := Evaluate(devDesign(), execs)
	if err != nil {
		t.Fatal(err)
	}
	if !out.RuleSatisfied {
		t.Fatal("capability arithmetic must still pass")
	}
	if !out.CustodyDominates {
		t.Fatal("dominating custody must be flagged")
	}
	noted := false
	for _, n := range out.Notes {
		if strings.Contains(n, "does not make the discipline affordable") {
			noted = true
		}
	}
	if !noted {
		t.Fatalf("custody domination must be stated, not hidden: %v", out.Notes)
	}
	hg := out.Arms[ArmHG]
	if hg.TaskCost != 72 || hg.CustodyCost != 72*50 || hg.FullCost != 72+72*50 {
		t.Fatalf("per-arm ledgers wrong: %+v", hg)
	}
}

// The review's source-derived counterexample: three informative episodes
// spanning two families, one repetition, HG completing everything, a
// favorable label. The arithmetic passes — and the outcome must still
// refuse both the spending rule (population nonconforming) and gate
// eligibility (unconditionally).
func TestReviewCounterexampleThreeEpisodesCannotSatisfyRule(t *testing.T) {
	eps := []Episode{
		{ID: "inf-01", Stratum: StratumInformative, Family: "fam-A"},
		{ID: "inf-02", Stratum: StratumInformative, Family: "fam-A"},
		{ID: "inf-03", Stratum: StratumInformative, Family: "fam-B"},
	}
	var execs []Execution
	for _, arm := range []Arm{ArmH0, ArmH1, ArmHG} {
		for _, ep := range eps {
			execs = append(execs, Execution{
				Arm: arm, EpisodeID: ep.ID, Run: 1,
				Completed: arm == ArmHG, TaskCost: 1,
			})
		}
	}
	out, err := Evaluate(Design{Episodes: eps, RunsPerCell: 1, EvidenceLabel: "protected-sealed"}, execs)
	if err != nil {
		t.Fatal(err)
	}
	if !out.ArithmeticSatisfied {
		t.Fatalf("the arithmetic itself passes in this construction: %+v", out.Conditions)
	}
	if out.PopulationConforms {
		t.Fatal("three episodes with no control strata must not conform to the roadmap population")
	}
	if out.RuleSatisfied {
		t.Fatal("the spending rule is defined over the roadmap population; a nonconforming batch cannot satisfy it")
	}
	if out.GateEligible {
		t.Fatal("no label can mint gate eligibility from this package")
	}
	if len(out.PopulationDefects) == 0 {
		t.Fatal("population defects must be named")
	}
}

// A negative charge is rejected: it would reduce reported expenditure.
func TestNegativeCostIsRejected(t *testing.T) {
	execs := grid(episodes24(), baseCounts(), 0)
	execs[0].TaskCost = -1
	if _, err := Evaluate(devDesign(), execs); err == nil {
		t.Fatal("a negative cost must be an error")
	} else if !strings.Contains(err.Error(), "negative cost") {
		t.Fatalf("error must name the negative charge: %v", err)
	}
}

// Cost accumulation refuses int64 overflow instead of wrapping.
func TestCostOverflowIsRejected(t *testing.T) {
	execs := grid(episodes24(), baseCounts(), 0)
	execs[0].TaskCost = math.MaxInt64
	execs[1].TaskCost = math.MaxInt64
	if _, err := Evaluate(devDesign(), execs); err == nil {
		t.Fatal("overflowing cost totals must be an error")
	} else if !strings.Contains(err.Error(), "overflow") {
		t.Fatalf("error must name the overflow: %v", err)
	}
}

// No evidence label — including the sealed-sounding one — flips
// eligibility: the field is constitutionally false here.
func TestGateEligibilityIsUnconditionallyFalse(t *testing.T) {
	execs := grid(episodes24(), baseCounts(), 0)
	for _, label := range []string{"development", "protected", "protected-sealed"} {
		d := devDesign()
		d.EvidenceLabel = label
		out, err := Evaluate(d, execs)
		if err != nil {
			t.Fatal(err)
		}
		if out.GateEligible {
			t.Fatalf("label %q must not produce gate eligibility", label)
		}
	}
}
