// Two-arm comparison runner — see PROTOCOL.md (preregistered at d75b4e0).
// Deterministic, exact-integer, no model calls. Run: go run .
package main

import (
	"fmt"
	"math/big"

	"github.com/instagrim-dev/newf/internal/witness"
)

var one = big.NewInt(1)

// instances returns the first k primes p >= 10^6 with p ≡ 1 (mod 24).
func instances(k int) []*big.Int {
	var out []*big.Int
	p := big.NewInt(1000000)
	for len(out) < k {
		if new(big.Int).Mod(p, big.NewInt(24)).Cmp(one) == 0 && p.ProbablyPrime(64) {
			out = append(out, new(big.Int).Set(p))
		}
		p = new(big.Int).Add(p, one)
	}
	return out
}

// ceilDiv returns ceil(a/b) for positive a, b.
func ceilDiv(a, b *big.Int) *big.Int {
	q, r := new(big.Int).QuoRem(a, b, new(big.Int))
	if r.Sign() != 0 {
		q.Add(q, one)
	}
	return q
}

// greedyFrom builds the 3-term tuple with first denominator x, then greedy.
// Abstains (returns nil) if a remainder hits zero early or goes nonpositive.
func greedyFrom(p, x *big.Int) []*big.Int {
	num := new(big.Int).Sub(new(big.Int).Mul(big.NewInt(4), x), p) // 4x - p
	den := new(big.Int).Mul(p, x)
	if num.Sign() <= 0 {
		return nil
	}
	terms := []*big.Int{new(big.Int).Set(x)}
	for len(terms) < 3 {
		u := ceilDiv(den, num)
		terms = append(terms, u)
		// r <- r - 1/u = (num*u - den) / (den*u)
		num = new(big.Int).Sub(new(big.Int).Mul(num, u), den)
		den = new(big.Int).Mul(den, u)
		if num.Sign() == 0 && len(terms) < 3 {
			return nil // exact in <3 terms; move abstains per protocol
		}
	}
	if num.Sign() != 0 {
		// inexact tuple: submitted anyway — the checker adjudicates.
		_ = num
	}
	return terms
}

// factorTrial returns the prime factorization of n by trial division.
func factorTrial(n *big.Int) map[string]int {
	f := map[string]int{}
	m := new(big.Int).Set(n)
	d := big.NewInt(2)
	for new(big.Int).Mul(d, d).Cmp(m) <= 0 {
		for new(big.Int).Mod(m, d).Sign() == 0 {
			f[d.String()]++
			m.Div(m, d)
		}
		d = new(big.Int).Add(d, one)
	}
	if m.Cmp(one) > 0 {
		f[m.String()]++
	}
	return f
}

// divisorSearch implements G1: bounded global divisor-lattice search.
func divisorSearch(p *big.Int) []*big.Int {
	x0 := ceilDiv(p, big.NewInt(4))
	for i := 0; i <= 10000; i++ {
		x := new(big.Int).Add(x0, big.NewInt(int64(i)))
		a := new(big.Int).Sub(new(big.Int).Mul(big.NewInt(4), x), p)
		b := new(big.Int).Mul(p, x)
		g := new(big.Int).GCD(nil, nil, a, b)
		ap := new(big.Int).Div(a, g)
		bp := new(big.Int).Div(b, g)
		// divisors of bp^2 from factorization of bp
		f := factorTrial(bp)
		divs := []*big.Int{big.NewInt(1)}
		for q, e := range f {
			qi, _ := new(big.Int).SetString(q, 10)
			cur := divs
			divs = nil
			for _, d := range cur {
				dd := new(big.Int).Set(d)
				for k := 0; k <= 2*e; k++ {
					divs = append(divs, new(big.Int).Set(dd))
					dd = new(big.Int).Mul(dd, qi)
				}
			}
		}
		target := new(big.Int).Mod(new(big.Int).Neg(bp), ap)
		var best *big.Int
		for _, d := range divs {
			if d.Cmp(bp) <= 0 && new(big.Int).Mod(d, ap).Cmp(target) == 0 {
				if best == nil || d.Cmp(best) < 0 {
					best = d
				}
			}
		}
		if best != nil {
			y := new(big.Int).Div(new(big.Int).Add(bp, best), ap)
			bp2 := new(big.Int).Mul(bp, bp)
			z := new(big.Int).Div(new(big.Int).Add(bp, new(big.Int).Div(bp2, best)), ap)
			return []*big.Int{x, y, z}
		}
	}
	return nil
}

type move struct {
	name string
	run  func(p *big.Int) []*big.Int
}

func check(p *big.Int, t []*big.Int) bool {
	c := witness.ErdosStrausClaim{N: p, X: t[0], Y: t[1], Z: t[2]}
	return c.Check() == nil
}

func main() {
	x0plus := func(off int64) func(p *big.Int) []*big.Int {
		return func(p *big.Int) []*big.Int {
			x := new(big.Int).Add(ceilDiv(p, big.NewInt(4)), big.NewInt(off))
			return greedyFrom(p, x)
		}
	}
	moves := map[string]move{
		"L1": {"L1", x0plus(0)},
		"L2": {"L2", x0plus(1)},
		"L3": {"L3", x0plus(2)},
		"G1": {"G1", divisorSearch},
	}
	arms := []struct {
		name  string
		order []string
	}{
		{"U", []string{"L1", "L2", "L3", "G1"}},
		{"M", []string{"G1", "L1", "L2"}},
	}
	const budget = 3
	ps := instances(20)
	fmt.Println("instance        arm  submissions  hit-move  per-move")
	totals := map[string]*struct{ subs, hits int }{
		"U": {}, "M": {},
	}
	localHits := 0
	localSubs := 0
	for _, p := range ps {
		for _, arm := range arms {
			subs := 0
			hitMove := "-"
			detail := ""
			for _, mn := range arm.order {
				if subs >= budget {
					break
				}
				t := moves[mn].run(p)
				if t == nil {
					detail += mn + ":abstain "
					continue // abstention emits no tuple, costs no submission
				}
				subs++
				ok := check(p, t)
				if mn != "G1" && arm.name == "U" {
					localSubs++
					if ok {
						localHits++
					}
				}
				if ok {
					hitMove = mn
					detail += mn + ":HIT "
					break
				}
				detail += mn + ":miss "
			}
			totals[arm.name].subs += subs
			if hitMove != "-" {
				totals[arm.name].hits++
			}
			fmt.Printf("%-14s  %-3s  %-11d  %-8s  %s\n", p, arm.name, subs, hitMove, detail)
		}
	}
	fmt.Println()
	for _, a := range []string{"U", "M"} {
		t := totals[a]
		cph := "inf"
		if t.hits > 0 {
			cph = fmt.Sprintf("%.2f", float64(t.subs)/float64(t.hits))
		}
		fmt.Printf("arm %s: hits=%d/20 submissions=%d cost-per-hit=%s\n", a, t.hits, t.subs, cph)
	}
	fmt.Printf("local-move submissions in arm U: %d, of which hits: %d\n", localSubs, localHits)
	u, m := totals["U"], totals["M"]
	earned := m.hits >= u.hits && m.subs < u.subs
	fmt.Printf("frozen decision rule — map earns its cost: %v\n", earned)
}
