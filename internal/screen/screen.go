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
//   - every stratum contributes to its declared endpoint: conditions (b)
//     and (e) sum completions over ALL episodes (matching the roadmap's
//     "3 more of the 24 tasks"); condition (d)'s family-diversity guard
//     counts informative episodes only; low-value and misleading episodes
//     additionally carry (c); every episode carries (a) and the totals;
//   - costs are recorded per execution in two ledgers (task-directed and
//     custody) and reported per arm in both, plus their sum; a
//     capability pass with dominating custody stays a capability pass —
//     with the custody figure reported beside it, never inside it.
package screen

import (
	"fmt"
	"math"
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
	// EvidenceLabel describes where the records came from. It is echoed
	// on the Outcome for the reader; it grants nothing. This package has
	// no way to verify custody, and therefore never produces an eligible
	// outcome regardless of label.
	EvidenceLabel string
}

// Roadmap population shape (revision 0.3.0 §4 G4-lite). The spending rule
// is DEFINED over this population; arithmetic on any other shape is a
// development calculation, not the screen.
const (
	RequiredInformative     = 12
	RequiredLowValue        = 6
	RequiredMisleading      = 6
	RequiredInformativeFams = 2 // condition (d) is unsatisfiable below this
)

// Roadmap thresholds (proposed investment judgments, fixed in revision
// 0.3.0; changing them without a new roadmap revision is tuning).
const (
	MarginPerRun          = 3 // (b): HG completes at least 3 more tasks per run than H1
	MaxLowValueLossPerRun = 1 // (c): HG loses at most 1 completion per run on low-value/misleading strata
	MinFamilies           = 2 // (d): successful differences in at least this many construction families
)

// Execution is one arm×episode×run cell's recorded result. Costs are
// unit-agnostic nonnegative integer units, consistent within a batch; a
// future runtime adapter must represent an unavailable measurement as
// unavailable, never as a measured zero.
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

// Outcome is the decision-arithmetic result. ArithmeticSatisfied is a
// statement about the supplied records only; RuleSatisfied is the spending
// rule as defined, which additionally requires the roadmap population.
type Outcome struct {
	EvidenceLabel string
	// GateEligible is ALWAYS false from this package. Eligibility for the
	// investment gate requires custodian-sealed protected episodes, a
	// validated custody chain, frozen arm snapshots, and execution
	// authorization — bindings this package cannot verify and that do not
	// yet exist. When they do, eligibility will be established by the
	// custody machinery that owns them, not by relabeling this
	// calculator's output.
	GateEligible bool
	Conditions   []Condition
	// ArithmeticSatisfied: conditions (a)–(e) all hold on the supplied
	// grid, whatever its shape. A development calculation can earn this.
	ArithmeticSatisfied bool
	// PopulationConforms: the episode population matches the roadmap
	// shape the spending rule is defined over (12/6/6, at least 2
	// construction families among informative episodes).
	PopulationConforms bool
	PopulationDefects  []string
	// RuleSatisfied = ArithmeticSatisfied && PopulationConforms: the
	// spending rule as written. It is still not the gate (GateEligible).
	RuleSatisfied    bool
	Arms             map[Arm]ArmReport
	CustodyDominates bool // total custody exceeds total task-directed cost; a diagnostic, not an affordability decision
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

	// Tallies. Costs are validated nonnegative and accumulated with
	// overflow checks: a negative charge would reduce reported
	// expenditure, and a wrapped total would misstate it silently.
	r := int64(d.RunsPerCell)
	sums := map[Arm]int64{}
	lowMisSums := map[Arm]int64{}
	perEpisode := map[Arm]map[string]int64{ArmH0: {}, ArmH1: {}, ArmHG: {}}
	reports := map[Arm]ArmReport{}
	invalidCount := int64(0)
	for _, e := range execs {
		if e.TaskCost < 0 || e.CustodyCost < 0 {
			return Outcome{}, fmt.Errorf("execution %s/%s run %d carries a negative cost (task %d, custody %d); charges are nonnegative and an unavailable measurement must be represented as unavailable, not negative or zero", e.Arm, e.EpisodeID, e.Run, e.TaskCost, e.CustodyCost)
		}
		rep := reports[e.Arm]
		var err error
		if rep.TaskCost, err = addChecked(rep.TaskCost, e.TaskCost); err != nil {
			return Outcome{}, fmt.Errorf("arm %s task-cost ledger: %w", e.Arm, err)
		}
		if rep.CustodyCost, err = addChecked(rep.CustodyCost, e.CustodyCost); err != nil {
			return Outcome{}, fmt.Errorf("arm %s custody-cost ledger: %w", e.Arm, err)
		}
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
		var err error
		if rep.FullCost, err = addChecked(rep.TaskCost, rep.CustodyCost); err != nil {
			return Outcome{}, fmt.Errorf("arm %s full-cost ledger: %w", a, err)
		}
		reports[a] = rep
		if taskTotal, err = addChecked(taskTotal, rep.TaskCost); err != nil {
			return Outcome{}, fmt.Errorf("cross-arm task total: %w", err)
		}
		if custodyTotal, err = addChecked(custodyTotal, rep.CustodyCost); err != nil {
			return Outcome{}, fmt.Errorf("cross-arm custody total: %w", err)
		}
	}

	// Population conformance: the spending rule is defined over the
	// roadmap population. Arithmetic on any other shape stays a
	// development calculation and cannot satisfy the rule.
	var popDefects []string
	stratumCounts := map[Stratum]int{}
	infFams := map[string]bool{}
	for _, ep := range episodes {
		stratumCounts[ep.Stratum]++
		if ep.Stratum == StratumInformative {
			infFams[ep.Family] = true
		}
	}
	if got := stratumCounts[StratumInformative]; got != RequiredInformative {
		popDefects = append(popDefects, fmt.Sprintf("%d informative episodes, roadmap population requires %d", got, RequiredInformative))
	}
	if got := stratumCounts[StratumLowValue]; got != RequiredLowValue {
		popDefects = append(popDefects, fmt.Sprintf("%d low-value episodes, roadmap population requires %d", got, RequiredLowValue))
	}
	if got := stratumCounts[StratumMisleading]; got != RequiredMisleading {
		popDefects = append(popDefects, fmt.Sprintf("%d misleading episodes, roadmap population requires %d", got, RequiredMisleading))
	}
	if len(infFams) < RequiredInformativeFams {
		popDefects = append(popDefects, fmt.Sprintf("%d construction families among informative episodes, roadmap population requires at least %d (condition (d) is unsatisfiable below that)", len(infFams), RequiredInformativeFams))
	}
	populationConforms := len(popDefects) == 0

	// (d): construction families containing an INFORMATIVE episode where
	// HG's summed completions strictly exceed H1's. Control strata
	// (low-value, misleading) are deliberately excluded: (d) is the
	// diversification guard on where shaping value comes from, and a
	// control-stratum fluke must not substitute for a second informative
	// family (adversarial review finding 2).
	famSet := map[string]bool{}
	for id, ep := range episodes {
		if ep.Stratum != StratumInformative {
			continue
		}
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
			Detail:    fmt.Sprintf("informative-stratum families with an episode where HG > H1: %v", families),
		},
		{
			Name:      "e",
			Statement: "HG completes at least as many protected tasks as H0 overall",
			Left:      sums[ArmHG], Op: ">=", Right: sums[ArmH0],
			Satisfied: sums[ArmHG] >= sums[ArmH0],
			Detail:    fmt.Sprintf("HG %d vs H0 %d", sums[ArmHG], sums[ArmH0]),
		},
	}
	arithmeticSatisfied := true
	for _, c := range conds {
		if !c.Satisfied {
			arithmeticSatisfied = false
		}
	}

	out := Outcome{
		EvidenceLabel:       d.EvidenceLabel,
		GateEligible:        false, // constitutionally: see Outcome.GateEligible
		Conditions:          conds,
		ArithmeticSatisfied: arithmeticSatisfied,
		PopulationConforms:  populationConforms,
		PopulationDefects:   popDefects,
		RuleSatisfied:       arithmeticSatisfied && populationConforms,
		Arms:                reports,
		CustodyDominates:    custodyTotal > taskTotal,
	}
	out.Notes = append(out.Notes, "gate eligibility cannot be produced here: it requires custodian-sealed protected episodes, a validated custody chain, frozen arm snapshots, and execution authorization — none of which this arithmetic can verify or supply")
	if !populationConforms {
		out.Notes = append(out.Notes, "the episode population does not match the roadmap shape the spending rule is defined over; this evaluation is a development calculation, not the screen")
	}
	if out.CustodyDominates {
		out.Notes = append(out.Notes, fmt.Sprintf("custody cost (%d) exceeds task-directed cost (%d): a capability pass does not make the discipline affordable; report both figures to the spending decision", custodyTotal, taskTotal))
	}
	return out, nil
}

// addChecked adds two nonnegative int64 values, refusing overflow rather
// than wrapping a cost total.
func addChecked(a, b int64) (int64, error) {
	if a > math.MaxInt64-b {
		return 0, fmt.Errorf("cost accumulation overflows int64 (%d + %d)", a, b)
	}
	return a + b, nil
}
