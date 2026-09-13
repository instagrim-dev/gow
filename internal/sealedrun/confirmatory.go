package sealedrun

import (
	"fmt"

	"github.com/instagrim-dev/newf/internal/finite"
	"github.com/instagrim-dev/newf/internal/rewrite"
	"github.com/instagrim-dev/newf/internal/screen"
	"github.com/instagrim-dev/newf/internal/shape"
)

// hgSelector abstracts which frozen controller drives the HG arm. A
// selector may refuse its input (v1's identity contract); a refusal
// aborts the run rather than substituting a silent decision.
type hgSelector func(in shape.Input, task finite.Expr, domain finite.Domain, rules []rewrite.Rule) (shape.Decision, error)

// runWithSelector is the shared three-arm execution core. Labeling is
// the caller's responsibility and is where evidence-grade honesty lives;
// this function refuses only structural defects.
func runWithSelector(p Pack, budget int, hg hgSelector, label string) (screen.Outcome, []EpisodeTrace, error) {
	if len(p.Episodes) == 0 {
		return screen.Outcome{}, nil, fmt.Errorf("empty pack")
	}
	if label == "" {
		return screen.Outcome{}, nil, fmt.Errorf("an outcome must carry its evidence-tier label")
	}
	menu := Menu()

	var execs []screen.Execution
	var traces []EpisodeTrace
	decls := make([]screen.Episode, 0, len(p.Episodes))
	for _, ep := range p.Episodes {
		decls = append(decls, ep.Decl)
		if len(ep.Vars) == 0 || len(ep.Vars) > 3 {
			return screen.Outcome{}, nil, fmt.Errorf("episode %s: search domain needs 1..3 variables, got %d", ep.Decl.ID, len(ep.Vars))
		}
		domain := finite.Domain{Width: 4, Vars: ep.Vars}
		pool := map[string]rewrite.Rule{}
		ruleList := make([]rewrite.Rule, 0, len(ep.CatalogNames))
		for _, name := range ep.CatalogNames {
			def, ok := menu[name]
			if !ok {
				return screen.Outcome{}, nil, fmt.Errorf("episode %s: catalog names %q, which is not on the admissible menu", ep.Decl.ID, name)
			}
			d := finite.Domain{Width: 4, Vars: def.domainVars}
			cert := finite.AssessEquivalence(finite.Binding{Sentence: name, Domain: d}, def.lhs, def.rhs)
			rule, defects := rewrite.AdmitRule(name, cert, def.lhs, def.rhs, d)
			if len(defects) > 0 {
				return screen.Outcome{}, nil, fmt.Errorf("episode %s: rule %q refused admission: %v", ep.Decl.ID, name, defects)
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
		if err != nil {
			return screen.Outcome{}, nil, fmt.Errorf("episode %s: HG selector refused: %w", ep.Decl.ID, err)
		}
		h1Dec := shape.SelectUngatedFrequency(in)
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
		trace := EpisodeTrace{
			EpisodeID: ep.Decl.ID,
			HG:        hgDec,
			H1:        h1Dec,
			Completed: map[screen.Arm]bool{},
			Explored:  map[screen.Arm]int{},
			Generated: map[screen.Arm]int{},
			ProbeWork: map[screen.Arm]int64{},
			Blocked:   map[screen.Arm]string{},
		}
		for _, arm := range []screen.Arm{screen.ArmH0, screen.ArmH1, screen.ArmHG} {
			ordered := make([]rewrite.Rule, 0, len(orders[arm]))
			for _, n := range orders[arm] {
				ordered = append(ordered, pool[n])
			}
			res, err := rewrite.Search(ep.Start, domain, ordered, rewrite.NodeCount, budget)
			if err != nil {
				return screen.Outcome{}, nil, fmt.Errorf("episode %s arm %s: %w", ep.Decl.ID, arm, err)
			}
			cell := measurementFor(arm, ep.Decl.ID, res, ep.TargetCost, probeCharge[arm])
			trace.Completed[arm] = cell.Completed
			trace.Explored[arm] = res.Explored
			trace.Generated[arm] = res.Generated
			trace.ProbeWork[arm] = probeCharge[arm]
			if cell.MeasurementBlocked {
				trace.Blocked[arm] = cell.BlockedReason
			}
			execs = append(execs, cell)
		}
		traces = append(traces, trace)
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
		return screen.Outcome{}, nil, err
	}
	return out, traces, nil
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
// TaskCost unit: candidate rewrites materialized — search successors
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
// materialized. Selectors that run no probe charge zero.
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

// RunConfirmatoryV2 runs a POST-FREEZE pack with the current
// failure-aware controller under the pack's own label (the version suffix
// identifies the HG controller, not a grade downgrade). Running it on a
// pack that motivated the controller's design is adaptation reuse and
// must go through RunDiagnosticV2 instead, which forces the taint suffix.
// The v0 counterpart needs no dedicated wrapper: RunWithBudget is already
// the v0 runner under the pack's own label.
//
// FREEZE ANCHOR: lawful only for packs authored after
// shape.ControllerVersionV2 froze — the commit that introduced THIS
// version string, not the commit that introduced its predecessor. The
// anchor was previously a hard-coded hash (`f7554cb`) naming
// shape-selector/1's freeze; when /1's frozen parameters changed, that
// hash silently became an anchor to a superseded procedure, and a pack
// authored in between would have been admitted as confirmatory evidence
// for a controller frozen after it (2026-09-13 self-review, finding 2).
// The anchor is therefore expressed as the version identity plus the
// label the outcome carries: a reader comparing a pack's authorship date
// against `git log -S'shape-selector/2'` resolves the true freeze point,
// and the emitted label records which procedure ran. A hash written here
// cannot stay correct across a version bump; the version string can.
func RunConfirmatoryV2(p Pack, budget int) (screen.Outcome, []EpisodeTrace, error) {
	if p.Label == "" {
		return screen.Outcome{}, nil, fmt.Errorf("a pack must carry its evidence-tier label")
	}
	return runWithSelector(p, budget, hgV2, p.Label+"+"+shape.ControllerVersionV2)
}
