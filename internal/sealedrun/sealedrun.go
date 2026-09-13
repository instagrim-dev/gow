// Package sealedrun executes an agent-sealed G4-lite screen: a custodian-
// authored episode pack against the three frozen arms, under fixed
// ceilings, scored by the screen's decision arithmetic.
//
// Evidence tier (decision D19, docs/plans/2026-09-12-015): packs at this
// tier are labeled "agent-sealed/v1" — authored by a clean-room agent
// that received only the custodian handoff and public interfaces, AFTER
// both arms froze (HG shape-selector/0, H1 h1-frequency/0), with the pack
// committed verbatim for replay. The tier is conversation-isolated and
// post-freeze but model-family-dependent: stronger than development,
// weaker than human-sealed. screen.Outcome.GateEligible remains false by
// construction; the spending decision is the operator's act reading the
// labeled outcome.
//
// Arms (all deterministic, zero provider spend):
//
//	H0: catalog order, history withheld
//	H1: shape.SelectUngatedFrequency (frozen comparator)
//	HG: shape.Select (frozen v0 controller)
package sealedrun

import (
	"fmt"

	"github.com/instagrim-dev/newf/internal/finite"
	"github.com/instagrim-dev/newf/internal/rewrite"
	"github.com/instagrim-dev/newf/internal/screen"
	"github.com/instagrim-dev/newf/internal/shape"
)

// ExpansionBudget is the per-cell search ceiling (within the D3a-class
// zero-spend authorization).
const ExpansionBudget = 500

// EpisodeSpec is one custodian-authored episode.
type EpisodeSpec struct {
	Decl         screen.Episode
	Start        finite.Expr
	Vars         []string // search-domain variables (width fixed at 4)
	CatalogNames []string // drawn from the admissible rule menu
	History      shape.History
	TargetCost   int64 // NodeCount threshold; completion also requires a verified endpoint
}

// Pack is a sealed episode collection with its provenance.
type Pack struct {
	Label      string // evidence tier label, e.g. "agent-sealed/v1"
	Provenance string // who authored it, from what inputs, when
	Episodes   []EpisodeSpec
}

// ruleDef is one admissible rule the runner can admit through the
// warrant boundary at execution time.
type ruleDef struct {
	domainVars []string
	lhs, rhs   finite.Expr
}

func va() finite.Expr { return finite.Var{Name: "a"} }
func vb() finite.Expr { return finite.Var{Name: "b"} }

// Menu returns the admissible rule menu: every rule a pack catalog may
// name. Each is admitted through finite.VerifyRuleWarrant at run time —
// the menu offers candidates, it does not bypass admission.
func Menu() map[string]ruleDef {
	u := func(op finite.UnaryOp, x finite.Expr) finite.Expr { return finite.Unary{Op: op, X: x} }
	b := func(op finite.BinaryOp, x, y finite.Expr) finite.Expr { return finite.Binary{Op: op, X: x, Y: y} }
	c := func(v uint64) finite.Expr { return finite.Const{Value: v} }
	one := []string{"a"}
	two := []string{"a", "b"}
	return map[string]ruleDef{
		"double-not":    {one, u(finite.OpNot, u(finite.OpNot, va())), va()},
		"neg-neg":       {one, u(finite.OpNeg, u(finite.OpNeg, va())), va()},
		"add-zero":      {one, b(finite.OpAdd, va(), c(0)), va()},
		"sub-zero":      {one, b(finite.OpSub, va(), c(0)), va()},
		"xor-zero":      {one, b(finite.OpXor, va(), c(0)), va()},
		"xor-self-zero": {one, b(finite.OpXor, va(), va()), c(0)},
		"or-self":       {one, b(finite.OpOr, va(), va()), va()},
		"and-self":      {one, b(finite.OpAnd, va(), va()), va()},
		"mul-one":       {one, b(finite.OpMul, va(), c(1)), va()},
		"mul-zero":      {one, b(finite.OpMul, va(), c(0)), c(0)},
		"not-intro":     {one, va(), u(finite.OpNot, u(finite.OpNot, va()))},
		"add-comm":      {two, b(finite.OpAdd, va(), vb()), b(finite.OpAdd, vb(), va())},
		"xor-comm":      {two, b(finite.OpXor, va(), vb()), b(finite.OpXor, vb(), va())},
	}
}

// EpisodeTrace is the per-episode attribution record.
type EpisodeTrace struct {
	EpisodeID string
	HG        shape.Decision
	H1        shape.Decision
	Completed map[screen.Arm]bool
	Explored  map[screen.Arm]int
}

// Run executes the pack against the three frozen arms and scores it.
// Deterministic; r=1 is disclosed via the runner's design (deterministic
// arms have structurally zero run variance).
func Run(p Pack) (screen.Outcome, []EpisodeTrace, error) {
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
		if len(ep.Vars) == 0 || len(ep.Vars) > 3 {
			return screen.Outcome{}, nil, fmt.Errorf("episode %s: search domain needs 1..3 variables, got %d", ep.Decl.ID, len(ep.Vars))
		}
		domain := finite.Domain{Width: 4, Vars: ep.Vars}

		// Admit this episode's catalog through the warrant boundary.
		pool := map[string]rewrite.Rule{}
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
		}

		in := shape.Input{
			TaskStart: shape.RenderExpr(ep.Start),
			Target:    ep.TargetCost,
			Catalog:   ep.CatalogNames,
			History:   ep.History,
		}
		hg := shape.Select(in)
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
			res, err := rewrite.Search(ep.Start, domain, ordered, rewrite.NodeCount, ExpansionBudget)
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
		EvidenceLabel: p.Label,
	}, execs)
	if err != nil {
		return screen.Outcome{}, nil, err
	}
	return out, traces, nil
}
