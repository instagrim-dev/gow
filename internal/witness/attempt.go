// Bounded-attempt executor: the attribution slice of the 2026-09-12 C1–C8
// review (remediation handoff 3).
//
// The review's C3 finding: a supplied tuple proves only "this submitted tuple
// does not satisfy the equation" — it does not, without another link,
// establish "this proposal's executed mechanism produced this tuple and
// failed its specified obligation". Fixture data is legitimate; unstated
// fixture relationships are not.
//
// This file supplies the link as a tiny, explicitly defined candidate
// procedure whose OUTPUT IS the tuple being checked: a named, versioned,
// deterministic procedure over declared integer parameters. The executed
// attempt records (procedure, version, canonical params, canonical tuple),
// and anyone — in particular admission — can RECOMPUTE the procedure over the
// recorded params and compare with the recorded tuple. The binding is
// checkable, not declared.
//
// An attempt's failure remains a failure of THAT bounded attempt only, never
// evidence against the broader conjecture. A procedure that cannot produce a
// tuple for its params ABSTAINS with a typed error: abstention is an input-
// level refusal — no tuple, no claim, nothing to persist.
package witness

import (
	"fmt"
	"math/big"
	"sort"
	"strings"
)

// AttemptExecutorVersion identifies the executor in provenance records. Bump
// on ANY change to a procedure's computation: the recorded version is part of
// what makes an old binding recomputable.
const AttemptExecutorVersion = "v1"

// AttemptProcedures names the registered bounded-attempt procedures.
//
//	equal-denominator  x = y = z = ceil(3n/4): 4/n = 3/x exactly when 4 | 3n,
//	                   so the tuple is exact only for n ≡ 0 (mod 4); for other
//	                   n it is a deliberately failing bounded probe (the shape
//	                   used throughout the review runs). Params: n.
//	greedy             x = ceil(n/4) + x0_offset, then greedy unit-fraction
//	                   expansion of the remainder trying y ∈ {ceil(1/r),
//	                   ceil(1/r)+1}; abstains when neither leaves an exact
//	                   unit-fraction remainder. Params: n, optional x0_offset.
func AttemptProcedures() []string { return []string{"equal-denominator", "greedy"} }

// AbstentionError reports that a procedure executed but produced no tuple for
// its params. Abstention is not a verdict and not a failure of the domain
// claim: there is nothing to check and nothing to persist.
type AbstentionError struct{ Reason string }

func (e AbstentionError) Error() string { return "attempt abstains: " + e.Reason }

// CanonicalAttemptParams is the deterministic serialization of attempt
// params recorded with a binding: keys sorted, "k=v" joined by ",". It is the
// exact string ExecuteAttempt re-parses, so recorded params round-trip.
func CanonicalAttemptParams(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+params[k])
	}
	return strings.Join(parts, ",")
}

// ParseAttemptParams inverts CanonicalAttemptParams.
func ParseAttemptParams(canonical string) (map[string]string, error) {
	out := map[string]string{}
	if strings.TrimSpace(canonical) == "" {
		return out, nil
	}
	for _, part := range strings.Split(canonical, ",") {
		k, v, ok := strings.Cut(part, "=")
		if !ok || strings.TrimSpace(k) == "" {
			return nil, fmt.Errorf("malformed attempt param %q (want k=v)", part)
		}
		if _, dup := out[strings.TrimSpace(k)]; dup {
			return nil, fmt.Errorf("duplicate attempt param %q", k)
		}
		out[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	return out, nil
}

// ExecuteAttempt deterministically runs one registered bounded-attempt
// procedure over declared params and returns the claim whose tuple the
// procedure produced. Unknown procedures and malformed params are input
// errors; a procedure that cannot produce a tuple returns AbstentionError.
// The computation is pure and exact — same procedure + params always yields
// the same tuple, which is what makes the recorded binding recheckable.
func ExecuteAttempt(procedure string, params map[string]string) (ErdosStrausClaim, error) {
	n, err := requiredPositiveParam(params, "n")
	if err != nil {
		return ErdosStrausClaim{}, err
	}
	switch procedure {
	case "equal-denominator":
		if err := rejectUnknownParams(params, "n"); err != nil {
			return ErdosStrausClaim{}, err
		}
		// x = ceil(3n/4)
		x := new(big.Int).Mul(n, big.NewInt(3))
		x = ceilDivInt(x, big.NewInt(4))
		return ErdosStrausClaim{N: n, X: x, Y: new(big.Int).Set(x), Z: new(big.Int).Set(x)}, nil
	case "greedy":
		if err := rejectUnknownParams(params, "n", "x0_offset"); err != nil {
			return ErdosStrausClaim{}, err
		}
		offset := big.NewInt(0)
		if raw, ok := params["x0_offset"]; ok {
			v, valid := new(big.Int).SetString(raw, 10)
			if !valid || v.Sign() < 0 {
				return ErdosStrausClaim{}, fmt.Errorf("attempt param x0_offset must be a non-negative decimal integer; got %q", raw)
			}
			offset = v
		}
		return greedyAttempt(n, offset)
	default:
		return ErdosStrausClaim{}, fmt.Errorf("unknown attempt procedure %q (registered: %s)", procedure, strings.Join(AttemptProcedures(), ", "))
	}
}

// VerifyAttemptBinding rechecks a recorded attempt→output binding: it re-runs
// the recorded procedure over the recorded canonical params and compares the
// canonical tuple. nil means the binding holds — the tuple IS the procedure's
// output. A non-nil error names exactly what diverged. Version discipline: a
// recorded executorVersion other than the running AttemptExecutorVersion is a
// verification refusal, not a mismatch — the old computation is not available
// to recompute.
func VerifyAttemptBinding(procedure, executorVersion, paramsCanonical, tupleCanonical string) error {
	if executorVersion != AttemptExecutorVersion {
		return fmt.Errorf("binding recorded under executor %s; running executor %s cannot recompute it", executorVersion, AttemptExecutorVersion)
	}
	params, err := ParseAttemptParams(paramsCanonical)
	if err != nil {
		return err
	}
	claim, err := ExecuteAttempt(procedure, params)
	if err != nil {
		return fmt.Errorf("recorded attempt no longer executes: %w", err)
	}
	if got := claim.Canonical(); got != tupleCanonical {
		return fmt.Errorf("attempt→output binding fails recheck: %s(%s) recomputes to %s, but the recorded tuple is %s", procedure, paramsCanonical, got, tupleCanonical)
	}
	return nil
}

// greedyAttempt: x = ceil(n/4) + offset, r = 4/n − 1/x; try y ∈ {ceil(1/r),
// ceil(1/r)+1}; accept the first leaving an exact positive unit-fraction
// remainder 1/z. Deterministic; abstains otherwise.
func greedyAttempt(n, offset *big.Int) (ErdosStrausClaim, error) {
	x := ceilDivInt(n, big.NewInt(4))
	x.Add(x, offset)
	r := new(big.Rat).SetFrac(big.NewInt(4), n)
	r.Sub(r, new(big.Rat).SetFrac(big.NewInt(1), x))
	if r.Sign() <= 0 {
		return ErdosStrausClaim{}, AbstentionError{Reason: fmt.Sprintf("4/%s − 1/%s is not positive; no two-term remainder to expand", n, x)}
	}
	y0 := ceilDivInt(r.Denom(), r.Num())
	for _, dy := range []int64{0, 1} {
		y := new(big.Int).Add(y0, big.NewInt(dy))
		if y.Sign() <= 0 {
			continue
		}
		r2 := new(big.Rat).Sub(r, new(big.Rat).SetFrac(big.NewInt(1), y))
		if r2.Sign() <= 0 {
			continue
		}
		if r2.Num().Cmp(big.NewInt(1)) == 0 { // big.Rat normalizes: unit fraction iff numerator is 1
			return ErdosStrausClaim{N: new(big.Int).Set(n), X: x, Y: y, Z: new(big.Int).Set(r2.Denom())}, nil
		}
	}
	return ErdosStrausClaim{}, AbstentionError{Reason: fmt.Sprintf("no exact unit-fraction remainder at x=%s for y in {%s, %s+1}", x, y0, y0)}
}

func requiredPositiveParam(params map[string]string, key string) (*big.Int, error) {
	raw, ok := params[key]
	if !ok {
		return nil, fmt.Errorf("attempt param %q is required", key)
	}
	v, valid := new(big.Int).SetString(raw, 10)
	if !valid || v.Sign() <= 0 {
		return nil, fmt.Errorf("attempt param %q must be a positive decimal integer; got %q", key, raw)
	}
	return v, nil
}

func rejectUnknownParams(params map[string]string, allowed ...string) error {
	ok := make(map[string]bool, len(allowed))
	for _, k := range allowed {
		ok[k] = true
	}
	for k := range params {
		if !ok[k] {
			return fmt.Errorf("unknown attempt param %q (allowed: %s)", k, strings.Join(allowed, ", "))
		}
	}
	return nil
}

// ceilDivInt returns ceil(a/b) for positive b.
func ceilDivInt(a, b *big.Int) *big.Int {
	q, m := new(big.Int).QuoRem(a, b, new(big.Int))
	if m.Sign() > 0 {
		q.Add(q, big.NewInt(1))
	}
	return q
}
