// Package sealedrun: pack authored 2026-09-13.
//
// PROVENANCE AND EVIDENCE GRADE — READ BEFORE CITING:
//
// This pack is IMPLEMENTER-AUTHORED (label "implementer-authored/v1").
// The intended agent-sealed/v1 tier was NOT achieved: three consecutive
// clean-room custodian agent launches (2026-09-13, transcripts retained
// under the session's subagent records) died on host infrastructure
// before authoring anything. Per the operator's D19 directive to continue
// research progress, the implementer authored this pack directly, with
// the honest downgrade recorded here and in the run record.
//
// What this grade means: the author has full knowledge of both frozen
// arms (shape-selector/0, h1-frequency/0) and of the development
// episodes; registry- and arm-shaping of episode construction cannot be
// excluded. Mitigations actually in force, rather than claimed:
//   - PRE-COMMITMENT: this file's content hash is committed to git
//     BEFORE the first screen execution; no edit after the first run is
//     admissible (any such edit is visible in history and voids the run).
//   - The author did NOT hand-simulate arm outcomes for the nontrivial
//     episodes; targets were derived from intended rewrite paths only,
//     and several episodes' completability under budget is genuinely
//     unknown to the author before the run.
//   - Variety disciplines from the custodian brief were followed:
//     operator families beyond not/add, mixed depths and variable
//     counts, churn rules (add-comm/xor-comm/not-intro) mixed into
//     catalogs, partial-help and subtly-lying histories.
//
// Upgrade path: a clean-room custodian re-authors a fresh pack when
// agent infrastructure recovers; this pack then becomes development
// material, per the standing tier ordering.
package sealedrun

import (
	"github.com/instagrim-dev/newf/internal/finite"
	"github.com/instagrim-dev/newf/internal/screen"
	"github.com/instagrim-dev/newf/internal/shape"
)

// Expression shorthands local to the pack.
func xv(n string) finite.Expr                         { return finite.Var{Name: n} }
func cn(v uint64) finite.Expr                         { return finite.Const{Value: v} }
func un(op finite.UnaryOp, x finite.Expr) finite.Expr { return finite.Unary{Op: op, X: x} }
func bn(op finite.BinaryOp, x, y finite.Expr) finite.Expr {
	return finite.Binary{Op: op, X: x, Y: y}
}

func att(start string, completed bool, rules ...string) shape.Attempt {
	return shape.Attempt{Start: start, RulesApplied: rules, Completed: completed}
}

func spec(id string, st screen.Stratum, fam string, start finite.Expr, vars, catalog []string, hist shape.History, target int64) EpisodeSpec {
	return EpisodeSpec{
		Decl:         screen.Episode{ID: id, Stratum: st, Family: fam},
		Start:        start,
		Vars:         vars,
		CatalogNames: catalog,
		History:      hist,
		TargetCost:   target,
	}
}

// AgentSealedV1 returns the pack. The name is kept for the runner wiring;
// the Label field, not the identifier, carries the evidence grade.
func AgentSealedV1() Pack {
	var eps []EpisodeSpec

	// ---- Informative, family neg-sub (4): reducers neg-neg/sub-zero
	// buried behind churn; histories credit the reducers on similar
	// starts, with irrelevant bool-heavy noise crediting add-comm.
	catD := []string{"add-comm", "not-intro", "neg-neg", "sub-zero"}
	negsubHist := shape.History{
		att("neg(neg(sub(y, 0)))", true, "neg-neg", "sub-zero"),
		att("sub(neg(neg(y)), 0)", true, "sub-zero", "neg-neg"),
		att("or(and(p, q), and(p, q))", true, "add-comm"),
		att("and(or(p, q), or(p, q))", true, "add-comm"),
	}
	eps = append(eps,
		spec("inf-01", screen.StratumInformative, "fam-negsub", un(finite.OpNeg, un(finite.OpNeg, bn(finite.OpSub, xv("x"), cn(0)))), []string{"x"}, catD, negsubHist, 1),
		spec("inf-02", screen.StratumInformative, "fam-negsub", bn(finite.OpSub, bn(finite.OpSub, un(finite.OpNeg, un(finite.OpNeg, xv("y"))), cn(0)), cn(0)), []string{"y"}, catD, negsubHist, 1),
		spec("inf-03", screen.StratumInformative, "fam-negsub", un(finite.OpNeg, un(finite.OpNeg, un(finite.OpNeg, un(finite.OpNeg, bn(finite.OpSub, xv("x"), cn(0)))))), []string{"x"}, catD, negsubHist, 1),
		spec("inf-04", screen.StratumInformative, "fam-negsub", bn(finite.OpSub, bn(finite.OpAdd, un(finite.OpNeg, un(finite.OpNeg, xv("z"))), bn(finite.OpSub, xv("z"), xv("z"))), cn(0)), []string{"z"}, catD, negsubHist, 5),
	)

	// ---- Informative, family mul-chain (4): mul-zero/add-zero/mul-one
	// reductions; not-intro is the buried explosive; noise credits
	// not-intro on not/xor-heavy irrelevant starts.
	catC := []string{"not-intro", "mul-zero", "mul-one", "add-zero"}
	mulHist := shape.History{
		att("mul(add(u, mul(v, 0)), 1)", true, "mul-zero", "add-zero", "mul-one"),
		att("add(mul(u, 0), mul(v, 1))", true, "mul-zero", "mul-one"),
		att("not(xor(p, not(q)))", true, "not-intro"),
		att("xor(not(p), not(not(q)))", true, "not-intro"),
	}
	eps = append(eps,
		spec("inf-05", screen.StratumInformative, "fam-mulchain", bn(finite.OpMul, bn(finite.OpAdd, xv("x"), bn(finite.OpMul, xv("y"), cn(0))), cn(1)), []string{"x", "y"}, catC, mulHist, 1),
		// inf-06 ANNOTATION (2026-09-13 external review, corpus
		// correction; the pre-committed pack bytes are retained
		// unchanged): the declared one-node target (cost 1) is
		// semantically valid — the expression equals y — but UNREACHABLE
		// through this catalog: mul-zero and mul-one reach add(0, y),
		// and catC has neither left-zero elimination (add(0,a)→a) nor
		// add-comm to expose the right-zero rule. The reachability
		// argument is structural, not a budget observation. All arms
		// fail it identically, so recorded counts stand; the episode
		// measures nothing informative about history use. Future packs:
		// validate a reference rewrite path for every intended-attainable
		// target before sealing.
		spec("inf-06", screen.StratumInformative, "fam-mulchain", bn(finite.OpAdd, bn(finite.OpMul, xv("x"), cn(0)), bn(finite.OpMul, xv("y"), cn(1))), []string{"x", "y"}, catC, mulHist, 1),
		spec("inf-07", screen.StratumInformative, "fam-mulchain", bn(finite.OpMul, bn(finite.OpMul, bn(finite.OpAdd, xv("z"), cn(0)), cn(1)), cn(1)), []string{"z"}, catC, mulHist, 1),
		spec("inf-08", screen.StratumInformative, "fam-mulchain", bn(finite.OpAdd, bn(finite.OpAdd, bn(finite.OpMul, xv("x"), cn(0)), xv("y")), bn(finite.OpMul, xv("x"), cn(0))), []string{"x", "y"}, catC, mulHist, 3),
	)

	// ---- Informative, family bool-collapse (4): or-self/and-self/
	// xor-self-zero; histories are only PARTIALLY helpful (they also
	// credit mul-one, present in the catalog but useless here).
	catB := []string{"xor-comm", "or-self", "and-self", "not-intro", "mul-one"}
	catE := []string{"xor-self-zero", "xor-comm", "not-intro", "add-zero"}
	boolHist := shape.History{
		att("or(and(p, q), and(p, q))", true, "or-self", "mul-one"),
		att("and(or(p, q), or(p, q))", true, "and-self", "mul-one"),
	}
	xorHist := shape.History{
		att("xor(or(p, q), or(p, q))", true, "xor-self-zero"),
		att("xor(add(p, q), add(p, q))", true, "xor-self-zero", "add-zero"),
	}
	eps = append(eps,
		spec("inf-09", screen.StratumInformative, "fam-boolcollapse", bn(finite.OpOr, bn(finite.OpAnd, xv("x"), xv("y")), bn(finite.OpAnd, xv("x"), xv("y"))), []string{"x", "y"}, catB, boolHist, 3),
		spec("inf-10", screen.StratumInformative, "fam-boolcollapse", bn(finite.OpAnd, bn(finite.OpOr, xv("x"), xv("z")), bn(finite.OpOr, xv("x"), xv("z"))), []string{"x", "z"}, catB, boolHist, 3),
		spec("inf-11", screen.StratumInformative, "fam-boolcollapse", bn(finite.OpXor, bn(finite.OpOr, xv("x"), xv("y")), bn(finite.OpOr, xv("x"), xv("y"))), []string{"x", "y"}, catE, xorHist, 1),
		spec("inf-12", screen.StratumInformative, "fam-boolcollapse", bn(finite.OpXor, bn(finite.OpAdd, xv("y"), cn(0)), bn(finite.OpAdd, xv("y"), cn(0))), []string{"y"}, catE, xorHist, 1),
	)

	return Pack{
		Label:      "implementer-authored/v1",
		Provenance: "Implementer-authored 2026-09-13 after three clean-room custodian agent launches died on host infrastructure before authoring anything (transcripts retained). Arms froze before authoring (shape-selector/0 at d68b961, h1-frequency/0 at d21ad00). Pre-commitment: content hash committed to git before first execution; no post-run edits admissible. Agent-sealed/v1 remains the upgrade path.",
		Episodes:   appendControlStrata(eps),
	}
}
