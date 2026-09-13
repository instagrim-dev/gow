package sealedrun

import (
	"fmt"

	"github.com/instagrim-dev/newf/internal/finite"
	"github.com/instagrim-dev/newf/internal/rewrite"
	"github.com/instagrim-dev/newf/internal/screen"
	"github.com/instagrim-dev/newf/internal/shape"
)

// RunDiagnosticV1 re-runs a pack with the v1 selector (shape-selector/1)
// in the HG arm. ADAPTATION-REUSE WARNING, enforced in the label: when
// the pack under test is the one whose failures motivated v1, this is a
// development diagnostic (the controller was fit to this pack's observed
// defects); it can never confirm v1's value. Confirmation requires a
// pack authored after v1 froze. The outcome label carries the suffix so
// no reader can mistake the grade.
func RunDiagnosticV1(p Pack, budget int) (screen.Outcome, []EpisodeTrace, error) {
	if len(p.Episodes) == 0 {
		return screen.Outcome{}, nil, fmt.Errorf("empty pack")
	}
	if p.Label == "" {
		return screen.Outcome{}, nil, fmt.Errorf("a pack must carry its evidence-tier label")
	}
	menu := Menu()

	var execs []screen.Execution
	var traces []EpisodeTrace
	decls := make([]screen.Episode, 0, len(p.Episodes))
	for _, ep := range p.Episodes {
		decls = append(decls, ep.Decl)
		domain := finite.Domain{Width: 4, Vars: ep.Vars}
		pool := map[string]rewrite.Rule{}
		ruleList := make([]rewrite.Rule, 0, len(ep.CatalogNames))
		for _, name := range ep.CatalogNames {
			def, ok := menu[name]
			if !ok {
				return screen.Outcome{}, nil, fmt.Errorf("episode %s: off-menu rule %q", ep.Decl.ID, name)
			}
			d := finite.Domain{Width: 4, Vars: def.domainVars}
			cert := finite.AssessEquivalence(finite.Binding{Sentence: name, Domain: d}, def.lhs, def.rhs)
			rule, defects := rewrite.AdmitRule(name, cert, def.lhs, def.rhs, d)
			if len(defects) > 0 {
				return screen.Outcome{}, nil, fmt.Errorf("episode %s: rule %q refused: %v", ep.Decl.ID, name, defects)
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
		hg := shape.SelectV1(shape.InputV1{Input: in, Task: ep.Start, Domain: domain, Rules: ruleList})
		h1 := shape.SelectUngatedFrequency(in)
		orders := map[screen.Arm][]string{
			screen.ArmH0: ep.CatalogNames,
			screen.ArmH1: h1.EnabledRules,
			screen.ArmHG: hg.EnabledRules,
		}
		trace := EpisodeTrace{EpisodeID: ep.Decl.ID, HG: hg, H1: h1, Completed: map[screen.Arm]bool{}, Explored: map[screen.Arm]int{}}
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
				TaskCost:  int64(res.Explored),
			})
		}
		traces = append(traces, trace)
	}

	out, err := screen.Evaluate(screen.Design{
		Episodes:      decls,
		RunsPerCell:   1,
		EvidenceLabel: p.Label + "+v1-adaptation-reuse-diagnostic",
	}, execs)
	if err != nil {
		return screen.Outcome{}, nil, err
	}
	return out, traces, nil
}
