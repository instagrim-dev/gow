// Package screen implements the G4-lite spending rule's decision
// arithmetic from the shaping roadmap (revision 0.3.0): the exact,
// inspectable evaluation of conditions (a)–(e) over a complete
// arm×episode×run execution grid, with task-directed budgets kept
// separate from custody costs.
//
// What this package is NOT: it is not the screen. Evaluating synthetic or
// development-labeled records exercises the decision procedure's
// arithmetic — it cannot satisfy the investment gate, and the Outcome says
// so on its face (GateEligible). Gate eligibility requires
// custodian-sealed protected episodes, frozen arms, and execution
// authorization, none of which this package can supply or verify; it
// echoes the evidence label it was handed and refuses eligibility for
// anything but the sealed-protected label, which only the (not yet built)
// custody machinery should ever apply.
//
// Arithmetic discipline (operator corrections to revision 0.3.0):
//   - exact integer comparisons on run-summed completions — condition (b)
//     compares sums against 3·r rather than rounding per-run averages;
//   - the execution grid must be complete: arms × episodes × runs, each
//     cell exactly once, or there is no decision at all (an incomplete
//     batch is malformed input, not a partial result);
//   - every stratum contributes to its declared endpoint: informative
//     episodes carry (b) and (d); low-value and misleading episodes carry
//     (c); every episode carries (a), (e), and the totals;
//   - costs are recorded per execution in two ledgers (task-directed and
//     custody) and reported per arm in both, plus their sum; a
//     capability pass with dominating custody stays a capability pass —
//     with the custody figure reported beside it, never inside it.
package screen

import (
	"fmt"
	"sort"
)

// Arm identifies one comparison arm from the roadmap's G4-lite design.
type Arm string

const (
	ArmH0 Arm = "H0" // no-history direct search
	ArmH1 Arm = "H1" // same-history direct search (primary comparator)
	ArmHG Arm = "HG" // history-conditioned shaping (proposed treatment)
)

var allArms = []Arm{ArmH0, ArmH1, ArmHG}

// Stratum names an episode's declared population stratum.
type Stratum string

const (
	StratumInformative Stratum = "history_informative"
	StratumLowValue    Stratum = "history_low_value"
	StratumMisleading  Stratum = "history_misleading"
)

// Episode is one protected task-and-history episode's declaration.
type Episode struct {
	ID      string
	Stratum Stratum
	Family  string // construction family, for condition (d)
}

// Design fixes the batch shape and evidence label before evaluation.
type Design struct {
	Episodes    []Episode
	RunsPerCell int
	// EvidenceLabel is applied by whoever supplied the records. Only
	// LabelProtectedSealed can make an outcome gate-eligible, and this
	// package has no way to verify the label — it echoes custody's
	// assertion and enforces only that everything else is ineligible.
	EvidenceLabel string
}

// LabelProtectedSealed is the one evidence label that custody machinery
// may apply to a sealed protected batch. Anything else is development.
const LabelProtectedSealed = "protected-sealed"

// Roadmap thresholds (proposed investment judgments, fixed in revision
// 0.3.0; changing them without a new roadmap revision is tuning).
const (
	MarginPerRun          = 3 // (b): HG completes at least 3 more tasks per run than H1
	MaxLowValueLossPerRun = 1 // (c): HG loses at most 1 completion per run on low-value/misleading strata
	MinFamilies           = 2 // (d): successful differences in at least this many construction families
)

// Execution is one arm×episode×run cell's recorded result.
type Execution struct {
	Arm              Arm
	EpisodeID        string
	Run              int // 1..RunsPerCell
	Completed        bool
	InvalidCertified bool // a result was certified valid that was not
	TaskCost         int64
	CustodyCost      int64
}

// Condition is one spending-rule clause with the exact integers compared.
type Condition struct {
	Name      string // "a".."e"
	Statement string
	Left      int64
	Op        string
	Right     int64
	Satisfied bool
	Detail    string
}

// ArmReport carries one arm's completions and its two cost ledgers.
type ArmReport struct {
	Completions int64
	TaskCost    int64
	CustodyCost int64
	FullCost    int64
}

// Outcome is the decision-arithmetic result. RuleSatisfied is a statement
// about the arithmetic of the supplied records only.
type Outcome struct {
	EvidenceLabel string
	// GateEligible is false for every label except LabelProtectedSealed.
	// A synthetic or development outcome that satisfies the rule is a
	// test of the procedure, not an earned tranche.
	GateEligible     bool
	Conditions       []Condition
	RuleSatisfied    bool
	Arms             map[Arm]ArmReport
	CustodyDominates bool // total custody exceeds total task-directed cost; reported, never netted away
	Notes            []string
}

// Evaluate checks the grid for completeness and computes conditions
// (a)–(e) exactly. A malformed grid returns an error and no Outcome:
// an incomplete batch yields no decision rather than a generous one.
func Evaluate(d Design, execs []Execution) (Outcome, error) {
	if d.RunsPerCell < 1 {
		return Outcome{}, fmt.Errorf("runs per cell must be at least 1, got %d", d.RunsPerCell)
	}
	if len(d.Episodes) == 0 {
		return Outcome{}, fmt.Errorf("no episodes declared")
	}
	episodes := make(map[string]Episode, len(d.Episodes))
	for _, ep := range d.Episodes {
		if ep.ID == "" || ep.Family == "" {
			return Outcome{}, fmt.Errorf("episode %+v is missing an ID or construction family", ep)
		}
		switch ep.Stratum {
		case StratumInformative, StratumLowValue, StratumMisleading:
		default:
			return Outcome{}, fmt.Errorf("episode %q has undeclared stratum %q", ep.ID, ep.Stratum)
		}
		if _, dup := episodes[ep.ID]; dup {
			return Outcome{}, fmt.Errorf("episode %q declared twice", ep.ID)
		}
		episodes[ep.ID] = ep
	}

	// Grid completeness: every arm×episode×run exactly once.
	type cell struct {
		arm Arm
		ep  string
		run int
	}
	seen := make(map[cell]bool, len(execs))
	for _, e := range execs {
		if _, ok := episodes[e.EpisodeID]; !ok {
			return Outcome{}, fmt.Errorf("execution references undeclared episode %q", e.EpisodeID)
		}
		validArm := false
		for _, a := range allArms {
			if e.Arm == a {
				validArm = true
			}
		}
		if !validArm {
			return Outcome{}, fmt.Errorf("execution references unknown arm %q", e.Arm)
		}
		if e.Run < 1 || e.Run > d.RunsPerCell {
			return Outcome{}, fmt.Errorf("execution %s/%s has run %d outside 1..%d", e.Arm, e.EpisodeID, e.Run, d.RunsPerCell)
		}
		c := cell{e.Arm, e.EpisodeID, e.Run}
		if seen[c] {
			return Outcome{}, fmt.Errorf("duplicate execution for %s/%s run %d", e.Arm, e.EpisodeID, e.Run)
		}
		seen[c] = true
	}
	want := len(allArms) * len(episodes) * d.RunsPerCell
	if len(execs) != want {
		return Outcome{}, fmt.Errorf("incomplete grid: %d executions supplied, %d arms × %d episodes × %d runs = %d required; an incomplete batch yields no decision", len(execs), len(allArms), len(episodes), d.RunsPerCell, want)
	}

	// Tallies.
	r := int64(d.RunsPerCell)
	sums := map[Arm]int64{}
	lowMisSums := map[Arm]int64{}
	perEpisode := map[Arm]map[string]int64{ArmH0: {}, ArmH1: {}, ArmHG: {}}
	reports := map[Arm]ArmReport{}
	invalidCount := int64(0)
	for _, e := range execs {
		rep := reports[e.Arm]
		rep.TaskCost += e.TaskCost
		rep.CustodyCost += e.CustodyCost
		if e.Completed {
			rep.Completions++
			sums[e.Arm]++
			perEpisode[e.Arm][e.EpisodeID]++
			st := episodes[e.EpisodeID].Stratum
			if st == StratumLowValue || st == StratumMisleading {
				lowMisSums[e.Arm]++
			}
		}
		if e.InvalidCertified {
			invalidCount++
		}
		reports[e.Arm] = rep
	}
	var taskTotal, custodyTotal int64
	for a, rep := range reports {
		rep.FullCost = rep.TaskCost + rep.CustodyCost
		reports[a] = rep
		taskTotal += rep.TaskCost
		custodyTotal += rep.CustodyCost
	}

	// (d): construction families containing an episode where HG's summed
	// completions strictly exceed H1's.
	famSet := map[string]bool{}
	for id, ep := range episodes {
		if perEpisode[ArmHG][id] > perEpisode[ArmH1][id] {
			famSet[ep.Family] = true
		}
	}
	families := make([]string, 0, len(famSet))
	for f := range famSet {
		families = append(families, f)
	}
	sort.Strings(families)

	conds := []Condition{
		{
			Name:      "a",
			Statement: "no invalid result is certified",
			Left:      invalidCount, Op: "==", Right: 0,
			Satisfied: invalidCount == 0,
			Detail:    fmt.Sprintf("%d executions certified an invalid result", invalidCount),
		},
		{
			Name:      "b",
			Statement: fmt.Sprintf("HG completes at least %d more tasks per run than H1 (sum comparison: HG_sum >= H1_sum + %d·r)", MarginPerRun, MarginPerRun),
			Left:      sums[ArmHG], Op: ">=", Right: sums[ArmH1] + MarginPerRun*r,
			Satisfied: sums[ArmHG] >= sums[ArmH1]+MarginPerRun*r,
			Detail:    fmt.Sprintf("HG %d vs H1 %d + %d·%d", sums[ArmHG], sums[ArmH1], MarginPerRun, r),
		},
		{
			Name:      "c",
			Statement: fmt.Sprintf("on low-value/misleading strata HG loses at most %d completion per run to H1 (H1_lowmis_sum − HG_lowmis_sum <= %d·r)", MaxLowValueLossPerRun, MaxLowValueLossPerRun),
			Left:      lowMisSums[ArmH1] - lowMisSums[ArmHG], Op: "<=", Right: MaxLowValueLossPerRun * r,
			Satisfied: lowMisSums[ArmH1]-lowMisSums[ArmHG] <= MaxLowValueLossPerRun*r,
			Detail:    fmt.Sprintf("H1 low/mis %d vs HG low/mis %d", lowMisSums[ArmH1], lowMisSums[ArmHG]),
		},
		{
			Name:      "d",
			Statement: fmt.Sprintf("successful differences occur in at least %d construction families", MinFamilies),
			Left:      int64(len(families)), Op: ">=", Right: MinFamilies,
			Satisfied: len(families) >= MinFamilies,
			Detail:    fmt.Sprintf("families with an episode where HG > H1: %v", families),
		},
		{
			Name:      "e",
			Statement: "HG completes at least as many protected tasks as H0 overall",
			Left:      sums[ArmHG], Op: ">=", Right: sums[ArmH0],
			Satisfied: sums[ArmHG] >= sums[ArmH0],
			Detail:    fmt.Sprintf("HG %d vs H0 %d", sums[ArmHG], sums[ArmH0]),
		},
	}
	ruleSatisfied := true
	for _, c := range conds {
		if !c.Satisfied {
			ruleSatisfied = false
		}
	}

	out := Outcome{
		EvidenceLabel:    d.EvidenceLabel,
		GateEligible:     d.EvidenceLabel == LabelProtectedSealed,
		Conditions:       conds,
		RuleSatisfied:    ruleSatisfied,
		Arms:             reports,
		CustodyDominates: custodyTotal > taskTotal,
	}
	if !out.GateEligible {
		out.Notes = append(out.Notes, fmt.Sprintf("evidence label %q is not %q: this outcome exercises the decision arithmetic and cannot satisfy the investment gate", d.EvidenceLabel, LabelProtectedSealed))
	}
	if out.CustodyDominates {
		out.Notes = append(out.Notes, fmt.Sprintf("custody cost (%d) exceeds task-directed cost (%d): a capability pass does not make the discipline affordable; report both figures to the spending decision", custodyTotal, taskTotal))
	}
	return out, nil
}
