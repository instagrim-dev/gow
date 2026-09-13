package sealedrun

import (
	"github.com/instagrim-dev/newf/internal/finite"
	"github.com/instagrim-dev/newf/internal/screen"
	"github.com/instagrim-dev/newf/internal/shape"
)

// appendControlStrata adds the 6 low-value and 6 misleading episodes.
func appendControlStrata(eps []EpisodeSpec) []EpisodeSpec {
	// ---- Low-value (6, fam-lowvalue): history empty or gated-out
	// irrelevant. Catalogs place reducers first, churn last, so a blind
	// arm is not handicapped; tasks span reducer families.
	catL1 := []string{"double-not", "xor-zero", "not-intro", "add-comm"}
	catL2 := []string{"or-self", "mul-one", "xor-comm"}
	irrelevant := shape.History{
		att("shr1(shl1(mul(w, w)))", true, "mul-one"),
		att("mul(sub(w, v), sub(w, v))", false, "add-comm"),
	}
	lows := []EpisodeSpec{
		spec("low-01", screen.StratumLowValue, "fam-lowvalue", un(finite.OpNot, un(finite.OpNot, bn(finite.OpXor, xv("x"), cn(0)))), []string{"x"}, catL1, nil, 1),
		spec("low-02", screen.StratumLowValue, "fam-lowvalue", bn(finite.OpXor, un(finite.OpNot, un(finite.OpNot, xv("y"))), cn(0)), []string{"y"}, catL1, nil, 1),
		spec("low-03", screen.StratumLowValue, "fam-lowvalue", bn(finite.OpOr, bn(finite.OpMul, xv("x"), cn(1)), bn(finite.OpMul, xv("x"), cn(1))), []string{"x"}, catL2, irrelevant, 1),
		spec("low-04", screen.StratumLowValue, "fam-lowvalue", bn(finite.OpMul, bn(finite.OpOr, xv("z"), xv("z")), cn(1)), []string{"z"}, catL2, irrelevant, 1),
		// Two deliberately uncompletable controls: target 1 needs a
		// reduction the catalog cannot express.
		spec("low-05", screen.StratumLowValue, "fam-lowvalue", bn(finite.OpAdd, xv("x"), xv("y")), []string{"x", "y"}, catL1, nil, 1),
		spec("low-06", screen.StratumLowValue, "fam-lowvalue", bn(finite.OpAnd, xv("y"), bn(finite.OpXor, xv("z"), cn(3))), []string{"y", "z"}, catL2, irrelevant, 2),
	}

	// ---- Misleading (6, fam-mislead): histories PASS the similarity
	// gate (same operator families as the task) but credit wrong rules.
	// mis-01/02: blatant — churn credited as the "success" on lookalike
	// starts, real reducers credited only in failures.
	catM1 := []string{"double-not", "add-zero", "not-intro"}
	blatant := shape.History{
		att("not(not(add(v, 0)))", true, "not-intro"),
		att("not(not(add(w, 0)))", false, "double-not", "add-zero"),
	}
	// mis-03/04: SUBTLE lie — xor-comm genuinely finishes lookalike
	// history tasks (xor(a,b) vs xor(b,a) framing) but on THIS task only
	// xor-self-zero reduces; the history promotes xor-comm, useful
	// elsewhere, useless here.
	catM2 := []string{"xor-self-zero", "xor-comm", "not-intro"}
	subtleXor := shape.History{
		att("xor(and(p, q), and(q, p))", true, "xor-comm"),
		att("xor(or(p, q), or(q, p))", true, "xor-comm"),
		att("xor(and(p, q), and(p, q))", false, "xor-self-zero"),
	}
	// mis-05/06: SUBTLE lie — mul-one credited on similar mul-heavy
	// starts, but this task reduces via mul-zero only.
	catM3 := []string{"mul-zero", "mul-one", "add-comm", "not-intro"}
	subtleMul := shape.History{
		att("mul(add(u, v), 1)", true, "mul-one"),
		att("mul(mul(u, 1), 1)", true, "mul-one"),
		att("mul(add(u, v), 0)", false, "mul-zero"),
	}
	mis := []EpisodeSpec{
		spec("mis-01", screen.StratumMisleading, "fam-mislead", un(finite.OpNot, un(finite.OpNot, bn(finite.OpAdd, xv("x"), cn(0)))), []string{"x"}, catM1, blatant, 1),
		spec("mis-02", screen.StratumMisleading, "fam-mislead", bn(finite.OpAdd, un(finite.OpNot, un(finite.OpNot, xv("y"))), cn(0)), []string{"y"}, catM1, blatant, 1),
		spec("mis-03", screen.StratumMisleading, "fam-mislead", bn(finite.OpXor, bn(finite.OpAnd, xv("x"), xv("y")), bn(finite.OpAnd, xv("x"), xv("y"))), []string{"x", "y"}, catM2, subtleXor, 1),
		spec("mis-04", screen.StratumMisleading, "fam-mislead", bn(finite.OpXor, bn(finite.OpOr, xv("y"), xv("z")), bn(finite.OpOr, xv("y"), xv("z"))), []string{"y", "z"}, catM2, subtleXor, 1),
		spec("mis-05", screen.StratumMisleading, "fam-mislead", bn(finite.OpMul, bn(finite.OpAdd, xv("x"), xv("y")), cn(0)), []string{"x", "y"}, catM3, subtleMul, 1),
		spec("mis-06", screen.StratumMisleading, "fam-mislead", bn(finite.OpAdd, bn(finite.OpMul, xv("z"), cn(0)), bn(finite.OpMul, xv("z"), cn(0))), []string{"z"}, catM3, subtleMul, 3),
	}

	return append(append(eps, lows...), mis...)
}
