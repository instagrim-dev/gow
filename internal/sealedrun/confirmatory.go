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
		// pre-search probes were previously uncharged). The Decision
		// meters candidates in the same unit Search reports in
		// res.Generated, so one ledger covers both.
		probeCharge := map[screen.Arm]int64{
			screen.ArmH0: 0,
			screen.ArmH1: int64(h1Dec.ProbeCandidates),
			screen.ArmHG: int64(hgDec.ProbeCandidates),
		}
		trace := EpisodeTrace{EpisodeID: ep.Decl.ID, HG: hgDec, H1: h1Dec, Completed: map[screen.Arm]bool{}, Explored: map[screen.Arm]int{}}
		for _, arm := range []screen.Arm{screen.ArmH0, screen.ArmH1, screen.ArmHG} {
			ordered := make([]rewrite.Rule, 0, len(orders[arm]))
			for _, n := range orders[arm] {
				ordered = append(ordered, pool[n])
			}
			res, err := rewrite.Search(ep.Start, domain, ordered, rewrite.NodeCount, budget)
			if err != nil {
				return screen.Outcome{}, nil, fmt.Errorf("episode %s arm %s: %w", ep.Decl.ID, arm, err)
			}
			completed := res.BestCost <= ep.TargetCost && res.EndpointVerified
			trace.Completed[arm] = completed
			trace.Explored[arm] = res.Explored
			execs = append(execs, screen.Execution{
				Arm: arm, EpisodeID: ep.Decl.ID, Run: 1,
				Completed: completed,
				// TaskCost unit: candidate rewrites materialized —
				// search successors (res.Generated) plus selector probe
				// candidates. Expansion counts stay in the trace; they
				// are the budget unit, not the cost (finding 1:
				// "distinguish expansion counts from full costs").
				TaskCost: int64(res.Generated) + probeCharge[arm],
				// Custody costs are not measured by this runner; they
				// are recorded as unmeasured, never as zero.
				CustodyMeasured: false,
			})
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

func hgV0(in shape.Input, _ finite.Expr, _ finite.Domain, _ []rewrite.Rule) (shape.Decision, error) {
	return shape.Select(in), nil
}

func hgV1(in shape.Input, task finite.Expr, domain finite.Domain, rules []rewrite.Rule) (shape.Decision, error) {
	return shape.SelectV1(shape.InputV1{Input: in, Task: task, Domain: domain, Rules: rules})
}

// RunConfirmatoryV1 runs a POST-FREEZE pack with the v1 controller under
// the pack's own label (suffix "+v1" identifies the HG controller, not a
// grade downgrade). Only lawful for packs authored after shape-selector/1
// froze (f7554cb); running v1 on a pack that motivated its design is
// adaptation reuse and must go through RunDiagnosticV1 instead, which
// forces the taint suffix. The v0 counterpart needs no dedicated wrapper:
// RunWithBudget is already the v0 runner under the pack's own label.
func RunConfirmatoryV1(p Pack, budget int) (screen.Outcome, []EpisodeTrace, error) {
	if p.Label == "" {
		return screen.Outcome{}, nil, fmt.Errorf("a pack must carry its evidence-tier label")
	}
	return runWithSelector(p, budget, hgV1, p.Label+"+v1")
}
