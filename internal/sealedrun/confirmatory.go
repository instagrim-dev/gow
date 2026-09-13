package sealedrun

import (
	"fmt"
	"reflect"

	"github.com/instagrim-dev/newf/internal/finite"
	"github.com/instagrim-dev/newf/internal/rewrite"
	"github.com/instagrim-dev/newf/internal/screen"
	"github.com/instagrim-dev/newf/internal/shape"
)

// hgSelector abstracts which frozen controller drives the HG arm. A
// selector may refuse its input (v1's identity contract); a refusal
// aborts the run rather than substituting a silent decision.
type hgSelector func(in shape.Input, task finite.Expr, domain finite.Domain, rules []rewrite.Rule) (shape.Decision, error)

// runWithSelector is the shared three-arm execution core. The route supplies
// the presentation label; this function establishes execution and arithmetic,
// never custody or freshness. Receipts survive errors and unscored batches.
func runWithSelector(p Pack, budget int, hg hgSelector, label string) (screen.Outcome, []EpisodeTrace, error) {
	return runWithSelectorBounded(p, budget, hg, label, rewrite.Limits{})
}

func runWithSelectorBounded(p Pack, budget int, hg hgSelector, label string, limits rewrite.Limits) (screen.Outcome, []EpisodeTrace, error) {
	if len(p.Episodes) == 0 {
		return screen.Outcome{}, nil, fmt.Errorf("empty pack")
	}
	if label == "" {
		return screen.Outcome{}, nil, fmt.Errorf("an outcome must carry its evidence-tier label")
	}
	menu := Menu()

	var execs []screen.Execution
	traces := make([]EpisodeTrace, len(p.Episodes))
	decls := make([]screen.Episode, 0, len(p.Episodes))
	for i, ep := range p.Episodes {
		traces[i] = newEpisodeTrace(ep.Decl.ID, p.Label, label)
		decls = append(decls, ep.Decl)
	}
	for i, ep := range p.Episodes {
		trace := &traces[i]
		refuse := func(err error) (screen.Outcome, []EpisodeTrace, error) {
			trace.Error = err.Error()
			return unscoredOutcome(label, err), traces, err
		}
		if len(ep.Vars) == 0 || len(ep.Vars) > 3 {
			return refuse(fmt.Errorf("episode %s: search domain needs 1..3 variables, got %d", ep.Decl.ID, len(ep.Vars)))
		}
		domain := finite.Domain{Width: 4, Vars: ep.Vars}
		if defects := finite.ValidateExpr(ep.Start, domain); len(defects) > 0 {
			return refuse(fmt.Errorf("episode %s: start expression is invalid: %v", ep.Decl.ID, defects))
		}
		pool := map[string]rewrite.Rule{}
		ruleList := make([]rewrite.Rule, 0, len(ep.CatalogNames))
		for _, name := range ep.CatalogNames {
			def, ok := menu[name]
			if !ok {
				return refuse(fmt.Errorf("episode %s: catalog names %q, which is not on the admissible menu", ep.Decl.ID, name))
			}
			d := finite.Domain{Width: 4, Vars: def.domainVars}
			cert := finite.AssessEquivalence(finite.Binding{Sentence: name, Domain: d}, def.lhs, def.rhs)
			rule, defects := rewrite.AdmitRule(name, cert, def.lhs, def.rhs, d)
			if len(defects) > 0 {
				return refuse(fmt.Errorf("episode %s: rule %q refused admission: %v", ep.Decl.ID, name, defects))
			}
			pool[name] = rule
			ruleList = append(ruleList, rule)
		}

		in := shape.Input{
			TaskStart: shape.RenderExpr(ep.Start),
			Target:    ep.TargetCost,
			Catalog:   ep.CatalogNames,
			History:   ep.History,
		}
		hgDec, err := hg(in, ep.Start, domain, ruleList)
		trace.HG = hgDec
		if err != nil {
			trace.recordProbeWork(screen.ArmHG, probeWork(hgDec))
			return refuse(fmt.Errorf("episode %s: HG selector refused: %w", ep.Decl.ID, err))
		}
		h1Dec := shape.SelectUngatedFrequency(in)
		trace.H1 = h1Dec
		orders := map[screen.Arm][]string{
			screen.ArmH0: ep.CatalogNames,
			screen.ArmH1: h1Dec.EnabledRules,
			screen.ArmHG: hgDec.EnabledRules,
		}
		// Selector probe work is charged to the arm whose selector
		// performed it (2026-09-13 external review finding 1: v1's
		// pre-search probes were previously uncharged). BOTH meters are
		// charged: a probe that matches nothing still traverses every
		// position of the task, so charging candidates alone would leave
		// non-matching probes free — reopening the uncharged-work hole
		// at a smaller scale (2026-09-13 self-review).
		probeCharge := map[screen.Arm]int64{
			screen.ArmH0: 0,
			screen.ArmH1: probeWork(h1Dec),
			screen.ArmHG: probeWork(hgDec),
		}
		for arm, charge := range probeCharge {
			trace.recordProbeWork(arm, charge)
		}
		for _, arm := range []screen.Arm{screen.ArmH0, screen.ArmH1, screen.ArmHG} {
			ordered := make([]rewrite.Rule, 0, len(orders[arm]))
			for _, n := range orders[arm] {
				ordered = append(ordered, pool[n])
			}
			res, err := rewrite.SearchBounded(ep.Start, domain, ordered, rewrite.NodeCount, budget, limits)
			cell := measurementFor(arm, ep.Decl.ID, res, ep.TargetCost, probeCharge[arm])
			if err != nil {
				cell.Completed = false
				cell.MeasurementBlocked = true
				cell.BlockedReason = err.Error()
			}
			trace.Cells[arm] = CellTrace{Execution: cell, SearchStarted: true, Result: res}
			trace.Completed[arm] = cell.Completed
			trace.Explored[arm] = res.Explored
			trace.Generated[arm] = res.Generated
			trace.ProbeWork[arm] = probeCharge[arm]
			if cell.MeasurementBlocked {
				trace.Blocked[arm] = cell.BlockedReason
			} else {
				delete(trace.Blocked, arm)
			}
			if err != nil {
				ct := trace.Cells[arm]
				ct.Error = err.Error()
				trace.Cells[arm] = ct
				return refuse(fmt.Errorf("episode %s arm %s: %w", ep.Decl.ID, arm, err))
			}
			execs = append(execs, cell)
		}
	}

	return finishEvaluate(decls, execs, traces, label)
}

func finishEvaluate(decls []screen.Episode, execs []screen.Execution, traces []EpisodeTrace, label string) (screen.Outcome, []EpisodeTrace, error) {
	out, err := screen.Evaluate(screen.Design{
		Episodes:      decls,
		RunsPerCell:   1,
		EvidenceLabel: label,
	}, execs)
	if err != nil {
		out = unscoredOutcome(label, err)
	}
	return out, traces, err
}

func unscoredOutcome(label string, err error) screen.Outcome {
	return screen.Outcome{EvidenceLabel: label, Notes: []string{"comparison not scored: " + err.Error()}}
}

func newEpisodeTrace(episodeID, sourceLabel, evidenceLabel string) EpisodeTrace {
	trace := EpisodeTrace{
		EpisodeID: episodeID, SourcePackLabel: sourceLabel, EvidenceLabel: evidenceLabel,
		Cells: map[screen.Arm]CellTrace{}, Completed: map[screen.Arm]bool{},
		Explored: map[screen.Arm]int{}, Generated: map[screen.Arm]int{},
		ProbeWork: map[screen.Arm]int64{}, Blocked: map[screen.Arm]string{},
	}
	for _, arm := range []screen.Arm{screen.ArmH0, screen.ArmH1, screen.ArmHG} {
		const reason = "search not executed"
		trace.Cells[arm] = CellTrace{Execution: screen.Execution{
			Arm: arm, EpisodeID: episodeID, Run: 1,
			MeasurementBlocked: true, BlockedReason: reason,
		}}
		trace.Blocked[arm] = reason
	}
	return trace
}

func (trace *EpisodeTrace) recordProbeWork(arm screen.Arm, charge int64) {
	trace.ProbeWork[arm] = charge
	cell := trace.Cells[arm]
	cell.Execution.TaskCost = charge
	trace.Cells[arm] = cell
}

// measurementFor converts one search result into the cell the screen will
// score. It is the SINGLE place a search outcome becomes a measurement,
// so the resource-stop rule cannot be bypassed by a producer that forgets
// it (2026-09-13 validator findings D2/D3: the previous inline abort had
// no regression guard, and screen.Execution.MeasurementBlocked had no
// production producer, leaving the guarantee to caller discipline).
//
// A resource stop is NOT a non-completion: reaching the state ceiling,
// refusing an oversize term, or being cancelled truncates the reachable
// set, so the cell measures nothing about the arm's capability and is
// marked blocked — screen.Evaluate then refuses the batch rather than
// scoring truncation as a miss. BudgetExhausted is deliberately not in
// that set: the budget is the declared measurement parameter, and
// stopping on it is the measurement working as designed.
//
// TaskCost unit: candidates admitted to size preflight — search candidates
// (res.Generated) plus the selector's charged probe work (rule
// applications AND candidates; see probeWork). Expansion counts stay in
// the trace; they are the budget unit, not the cost (external review
// finding 1: "distinguish expansion counts from full costs").
func measurementFor(arm screen.Arm, episodeID string, res rewrite.Result, target int64, probeCharge int64) screen.Execution {
	cell := screen.Execution{
		Arm: arm, EpisodeID: episodeID, Run: 1,
		TaskCost: int64(res.Generated) + probeCharge,
		// Custody costs are not measured by this runner; they are
		// recorded as unmeasured, never as zero.
		CustodyMeasured: false,
	}
	if blocked := searchBlockedReason(res); blocked != "" {
		cell.MeasurementBlocked = true
		cell.BlockedReason = blocked
		return cell // Completed stays false and means "not measured"
	}
	cell.Completed = res.BestCost <= target && res.EndpointVerified
	return cell
}

func hgV0(in shape.Input, _ finite.Expr, _ finite.Domain, _ []rewrite.Rule) (shape.Decision, error) {
	return shape.Select(in), nil
}

// probeWork is a selector's total charged pre-search work: one unit per
// rule probed against the task (the positional traversal every probe
// performs, match or not) plus one unit per candidate rewrite the probes
// admitted to size preflight. Selectors that run no probe charge zero.
func probeWork(dec shape.Decision) int64 {
	return int64(dec.ProbeRuleApplications) + int64(dec.ProbeCandidates)
}

// searchBlockedReason names the resource bound that truncated a search,
// or "" when the search stopped on its declared budget or ran to
// completion. BudgetExhausted is not a block: the budget is the declared
// measurement parameter.
func searchBlockedReason(res rewrite.Result) string {
	switch {
	case res.Cancelled:
		return "search cancelled before it completed"
	case res.StateBounded:
		return fmt.Sprintf("generated-state ceiling reached after %d expansions", res.Explored)
	case res.TermSizeBounded:
		return "at least one successor exceeded the term-size ceiling, truncating the reachable set"
	default:
		return ""
	}
}

func hgV2(in shape.Input, task finite.Expr, domain finite.Domain, rules []rewrite.Rule) (shape.Decision, error) {
	return shape.SelectV2(shape.InputV2{Input: in, Task: task, Domain: domain, Rules: rules})
}

// RunConfirmatoryV2 preserves the legacy API but supplies execution-only
// evidence. Pack carries no validated freeze, exposure, or custody manifest,
// so neither this function's name nor the caller's label can establish
// freshness. SourcePackLabel remains in each trace as an unverified claim;
// the outcome explicitly withholds freshness and custody validation.
//
// Deprecated: the name overstates the evidence this API can establish.
// Use RunDiagnosticV2 for known adaptation reuse.
func RunConfirmatoryV2(p Pack, budget int) (screen.Outcome, []EpisodeTrace, error) {
	if p.Label == "" {
		return screen.Outcome{}, nil, fmt.Errorf("a pack must carry its evidence-tier label")
	}
	label := "execution-only/freshness-unverified+" + shape.ControllerVersionV2
	if reusesExposedTask(p) {
		label += "-adaptation-reuse-diagnostic"
	}
	return runWithSelector(p, budget, hgV2, label)
}

// reusesExposedTask recognizes exact task reuse from the already exposed
// development pack, including renamed/relabelled packs or changed histories.
// Failure to recognize an exposed task proves nothing: the default label
// still withholds freshness and custody validation.
func reusesExposedTask(p Pack) bool {
	exposedEpisodes := AgentSealedV1().Episodes
	for _, ep := range p.Episodes {
		for _, exposed := range exposedEpisodes {
			if ep.TargetCost == exposed.TargetCost && reflect.DeepEqual(ep.Start, exposed.Start) &&
				reflect.DeepEqual(ep.Vars, exposed.Vars) && reflect.DeepEqual(ep.CatalogNames, exposed.CatalogNames) {
				return true
			}
		}
	}
	return false
}
