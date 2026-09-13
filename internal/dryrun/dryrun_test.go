package dryrun

import (
	"fmt"
	"testing"

	"github.com/instagrim-dev/newf/internal/finite"
	"github.com/instagrim-dev/newf/internal/rewrite"
	"github.com/instagrim-dev/newf/internal/screen"
)

// Authorized ceilings (D3a, docs/plans/2026-09-12-015): 10,000 expansions
// per arm-episode cell; 24 development episodes; zero provider spend.
// This run uses far less than the ceiling and states what it used.
const expansionBudget = 500

// episode is one self-authored development episode: a start expression,
// an optimization target (cost threshold), and its population declaration.
// "History" is deliberately absent: no shaping selector exists to consume
// it (D6), and inventing one inside the dry run would fake the treatment.
type episode struct {
	decl   screen.Episode
	start  finite.Expr
	target int64
}

func v(n string) finite.Expr { return finite.Var{Name: n} }
func nn(e finite.Expr) finite.Expr {
	return finite.Unary{Op: finite.OpNot, X: finite.Unary{Op: finite.OpNot, X: e}}
}
func addZ(e finite.Expr) finite.Expr {
	return finite.Binary{Op: finite.OpAdd, X: e, Y: finite.Const{Value: 0}}
}
func xorSelf(e finite.Expr) finite.Expr {
	return finite.Binary{Op: finite.OpXor, X: e, Y: e}
}

// devEpisodes builds the 24-episode roadmap population deterministically:
// 12 informative (families A and B), 6 low-value (C), 6 misleading (D).
// Targets are chosen so some episodes are completable under budget and
// some are not — variety for the arithmetic, identical across arms.
func devEpisodes() []episode {
	var eps []episode
	base := []finite.Expr{v("x"), v("y"), finite.Binary{Op: finite.OpAnd, X: v("x"), Y: v("y")}}
	for i := 1; i <= 12; i++ {
		fam := "fam-A"
		if i > 6 {
			fam = "fam-B"
		}
		start := nn(addZ(base[i%3]))
		target := int64(3) // reachable: double-not + add-zero strip to the base
		if i%4 == 0 {
			target = 0 // unreachable: minimum cost is 1
		}
		eps = append(eps, episode{
			decl:   screen.Episode{ID: fmt.Sprintf("inf-%02d", i), Stratum: screen.StratumInformative, Family: fam},
			start:  start,
			target: target,
		})
	}
	for i := 1; i <= 6; i++ {
		eps = append(eps, episode{
			decl:   screen.Episode{ID: fmt.Sprintf("low-%02d", i), Stratum: screen.StratumLowValue, Family: "fam-C"},
			start:  xorSelf(nn(v("x"))),
			target: 1, // reachable via xor-self-zero
		})
	}
	for i := 1; i <= 6; i++ {
		target := int64(1)
		if i%2 == 0 {
			target = 0 // unreachable
		}
		eps = append(eps, episode{
			decl:   screen.Episode{ID: fmt.Sprintf("mis-%02d", i), Stratum: screen.StratumMisleading, Family: "fam-D"},
			start:  addZ(xorSelf(v("y"))),
			target: target,
		})
	}
	return eps
}

// admitDevRules admits the development rule set through the real warrant
// boundary — the dry run may not shortcut B→C admission.
func admitDevRules(t *testing.T) []rewrite.Rule {
	t.Helper()
	d1 := finite.Domain{Width: 4, Vars: []string{"a"}}
	adm := func(name string, lhs, rhs finite.Expr, d finite.Domain) rewrite.Rule {
		t.Helper()
		cert := finite.AssessEquivalence(finite.Binding{Sentence: name, Domain: d}, lhs, rhs)
		rule, defects := rewrite.AdmitRule(name, cert, lhs, rhs, d)
		if len(defects) > 0 {
			t.Fatalf("dev rule %q refused: %v", name, defects)
		}
		return rule
	}
	return []rewrite.Rule{
		adm("double-not", nn(v("a")), v("a"), d1),
		adm("add-zero", addZ(v("a")), v("a"), d1),
		adm("xor-self-zero", xorSelf(v("a")), finite.Const{Value: 0}, d1),
	}
}

// TestDevelopmentDryRunOfTheG4LiteLoop is the authorized D3a execution.
func TestDevelopmentDryRunOfTheG4LiteLoop(t *testing.T) {
	rules := admitDevRules(t)
	eps := devEpisodes()
	searchDomain := finite.Domain{Width: 4, Vars: []string{"x", "y"}}

	decls := make([]screen.Episode, 0, len(eps))
	for _, ep := range eps {
		decls = append(decls, ep.decl)
	}

	// All three arms run the identical deterministic procedure: no
	// shaping selector exists, and the dry run must not fake one.
	arms := []screen.Arm{screen.ArmH0, screen.ArmH1, screen.ArmHG}
	var execs []screen.Execution
	completedEpisodes := 0
	for _, ep := range eps {
		res, err := rewrite.Search(ep.start, searchDomain, rules, rewrite.NodeCount, expansionBudget)
		if err != nil {
			t.Fatalf("episode %s: %v", ep.decl.ID, err)
		}
		if res.BudgetExhausted {
			t.Fatalf("episode %s exhausted the %d-expansion budget; the dev space should close", ep.decl.ID, expansionBudget)
		}
		// A completion requires BOTH the target and the independently
		// replayed endpoint: an unverified candidate completes nothing.
		completed := res.BestCost <= ep.target && res.EndpointVerified
		if !res.EndpointVerified {
			t.Fatalf("episode %s endpoint must replay on this small domain: %s", ep.decl.ID, res.Endpoint.Reason)
		}
		if completed {
			completedEpisodes++
		}
		for _, arm := range arms {
			execs = append(execs, screen.Execution{
				Arm:       arm,
				EpisodeID: ep.decl.ID,
				Run:       1,
				Completed: completed,
				TaskCost:  int64(res.Explored), // search effort, per the cost ledger
				// CustodyCost 0: the dry run meters no custody; the
				// sealed screen must (docs/plans/2026-09-12-015).
			})
		}
	}

	// Variety check: the arithmetic must see both completions and
	// non-completions, or the dry run validates less than it claims.
	if completedEpisodes == 0 || completedEpisodes == len(eps) {
		t.Fatalf("degenerate dry-run population: %d/%d episodes completed; targets must produce both outcomes", completedEpisodes, len(eps))
	}

	out, err := screen.Evaluate(screen.Design{
		Episodes:      decls,
		RunsPerCell:   1, // deterministic procedure; run variance is structurally zero and disclosed
		EvidenceLabel: "development-dry-run",
	}, execs)
	if err != nil {
		t.Fatalf("the 24x3x1 grid must evaluate: %v", err)
	}

	// Plumbing assertions: population and per-condition arithmetic all
	// computed on a conforming grid.
	if !out.PopulationConforms {
		t.Fatalf("the dev population mirrors the roadmap shape: %v", out.PopulationDefects)
	}
	if len(out.Conditions) != 5 {
		t.Fatalf("all five spending-rule conditions must be computed, got %d", len(out.Conditions))
	}

	// Honesty assertions: identical arms mean margin zero — the spending
	// rule MUST fail, and a passing dry run would be a dry-run bug.
	byName := map[string]screen.Condition{}
	for _, c := range out.Conditions {
		byName[c.Name] = c
	}
	if byName["a"].Satisfied != true {
		t.Fatalf("(a) no invalid certifications in the dry run: %+v", byName["a"])
	}
	if byName["b"].Satisfied {
		t.Fatalf("(b) must fail at margin zero — no shaping selector exists: %+v", byName["b"])
	}
	if !byName["c"].Satisfied {
		t.Fatalf("(c) identical arms lose nothing on control strata: %+v", byName["c"])
	}
	if byName["d"].Satisfied {
		t.Fatalf("(d) identical arms produce no differing families: %+v", byName["d"])
	}
	if !byName["e"].Satisfied {
		t.Fatalf("(e) HG ties H0 at margin zero: %+v", byName["e"])
	}
	if out.ArithmeticSatisfied || out.RuleSatisfied {
		t.Fatal("the dry run must not satisfy the spending rule")
	}
	if out.GateEligible {
		t.Fatal("gate eligibility must be unreachable from a dry run")
	}

	// Cost ledger sanity: real nonzero task-directed effort was metered.
	for _, arm := range arms {
		if out.Arms[arm].TaskCost <= 0 {
			t.Fatalf("arm %s metered no search effort", arm)
		}
	}
}
