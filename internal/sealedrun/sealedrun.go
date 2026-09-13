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

// Run executes the pack under the default ceiling. See RunWithBudget.
func Run(p Pack) (screen.Outcome, []EpisodeTrace, error) {
	return RunWithBudget(p, ExpansionBudget)
}

// RunWithBudget executes the pack against the three frozen arms under
// the given per-cell expansion budget and scores it. Deterministic; r=1
// is disclosed via the runner's design (deterministic arms have
// structurally zero run variance). The budget is a measurement
// parameter: choosing it requires an arm-blind calibration rule stated
// before computation (see CalibrateH0MinBudgets), because a budget tuned
// after observing inter-arm outcomes would be tuning the screen.
func RunWithBudget(p Pack, budget int) (screen.Outcome, []EpisodeTrace, error) {
	if p.Label == "" {
		return screen.Outcome{}, nil, fmt.Errorf("a pack must carry its evidence-tier label")
	}
	return runWithSelector(p, budget, hgV0, p.Label)
}
