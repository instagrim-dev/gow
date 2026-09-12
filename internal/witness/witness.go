// Package witness implements pure, exact-integer domain witness checkers —
// the strongest tier of the verification hierarchy applied to DOMAIN claims
// (verification_subject = domain-goal), not to annotations.
//
// The first checker decides single-instance Erdős–Straus witness claims: for
// a positive integer n, the tuple (x, y, z) of positive integers witnesses
// 4/n = 1/x + 1/y + 1/z exactly when
//
//	4·x·y·z == n·(y·z + x·z + x·y)
//
// over exact big integers — no floats, no tolerance, no judgment. A verdict
// from this package is a reproducible computation about the domain object
// itself; it is the external outcome mechanism the two-observation episode
// protocol consumes (issue #23; 2026-09-12 review, epistemic finding).
package witness

import (
	"fmt"
	"math/big"
	"strings"
)

// CheckerName and CheckerVersion identify this checker in provenance records.
const (
	CheckerName    = "erdos-straus-witness"
	CheckerVersion = "v1"
)

// ErdosStrausClaim is one single-instance witness claim: (X, Y, Z) witnesses
// 4/N = 1/X + 1/Y + 1/Z. All four values must be positive integers.
type ErdosStrausClaim struct {
	N *big.Int
	X *big.Int
	Y *big.Int
	Z *big.Int
}

// Check decides the claim exactly. nil means the identity holds; a non-nil
// error names precisely why it does not (missing value, non-positive value,
// or identity mismatch with both sides shown). The check is deterministic and
// reproducible from the recorded tuple alone.
func (c ErdosStrausClaim) Check() error {
	for _, v := range []struct {
		name string
		val  *big.Int
	}{{"n", c.N}, {"x", c.X}, {"y", c.Y}, {"z", c.Z}} {
		if v.val == nil {
			return fmt.Errorf("witness claim missing %s", v.name)
		}
		if v.val.Sign() <= 0 {
			return fmt.Errorf("witness claim requires positive %s; got %s", v.name, v.val.String())
		}
	}
	// lhs = 4·x·y·z
	lhs := new(big.Int).Mul(c.X, c.Y)
	lhs.Mul(lhs, c.Z)
	lhs.Mul(lhs, big.NewInt(4))
	// rhs = n·(y·z + x·z + x·y)
	yz := new(big.Int).Mul(c.Y, c.Z)
	xz := new(big.Int).Mul(c.X, c.Z)
	xy := new(big.Int).Mul(c.X, c.Y)
	sum := new(big.Int).Add(yz, xz)
	sum.Add(sum, xy)
	rhs := new(big.Int).Mul(c.N, sum)
	if lhs.Cmp(rhs) != 0 {
		return fmt.Errorf("identity fails: 4·x·y·z = %s but n·(yz+xz+xy) = %s for n=%s x=%s y=%s z=%s",
			lhs.String(), rhs.String(), c.N.String(), c.X.String(), c.Y.String(), c.Z.String())
	}
	return nil
}

// Canonical returns the deterministic serialization recorded with an
// observation, sufficient to re-run the check.
func (c ErdosStrausClaim) Canonical() string {
	s := func(v *big.Int) string {
		if v == nil {
			return ""
		}
		return v.String()
	}
	return fmt.Sprintf(`{"checker":"%s/%s","n":"%s","x":"%s","y":"%s","z":"%s"}`,
		CheckerName, CheckerVersion, s(c.N), s(c.X), s(c.Y), s(c.Z))
}

// ParseErdosStraus parses an operator-supplied tuple "n,x,y,z" (decimal,
// arbitrary precision). It rejects malformed input loudly rather than
// guessing.
func ParseErdosStraus(tuple string) (ErdosStrausClaim, error) {
	parts := strings.Split(tuple, ",")
	if len(parts) != 4 {
		return ErdosStrausClaim{}, fmt.Errorf("witness tuple must be n,x,y,z; got %d values", len(parts))
	}
	vals := make([]*big.Int, 4)
	names := []string{"n", "x", "y", "z"}
	for i, p := range parts {
		v, ok := new(big.Int).SetString(strings.TrimSpace(p), 10)
		if !ok {
			return ErdosStrausClaim{}, fmt.Errorf("witness tuple %s is not a decimal integer: %q", names[i], strings.TrimSpace(p))
		}
		vals[i] = v
	}
	return ErdosStrausClaim{N: vals[0], X: vals[1], Y: vals[2], Z: vals[3]}, nil
}
