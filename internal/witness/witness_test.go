package witness

import (
	"math/big"
	"strings"
	"testing"
)

func claim(n, x, y, z int64) ErdosStrausClaim {
	return ErdosStrausClaim{N: big.NewInt(n), X: big.NewInt(x), Y: big.NewInt(y), Z: big.NewInt(z)}
}

func TestCheckAcceptsKnownWitnesses(t *testing.T) {
	// 4/5 = 1/2 + 1/4 + 1/20 (the probe case from issue #23).
	for _, c := range []ErdosStrausClaim{
		claim(5, 2, 4, 20),    // 4/5 = 1/2 + 1/4 + 1/20 (probe case from issue #23)
		claim(7, 2, 15, 210),  // 4/7 = 1/2 + 1/15 + 1/210
		claim(4, 2, 4, 4),     // 4/4 = 1/2 + 1/4 + 1/4
		claim(24, 12, 18, 36), // 4/24 = 1/6 = 1/12 + 1/18 + 1/36
	} {
		if err := c.Check(); err != nil {
			t.Fatalf("valid witness rejected: %v", err)
		}
	}
}

func TestCheckRejectsNearMiss(t *testing.T) {
	// The near-miss from issue #23: (2,4,21) for n=5 — 1/2+1/4+1/21 != 4/5.
	err := claim(5, 2, 4, 21).Check()
	if err == nil {
		t.Fatal("near-miss accepted")
	}
	if !strings.Contains(err.Error(), "identity fails") {
		t.Fatalf("rejection must name the exact failure: %v", err)
	}
}

func TestCheckRejectsNonPositiveAndMissing(t *testing.T) {
	if err := claim(5, 0, 4, 20).Check(); err == nil {
		t.Fatal("zero x accepted")
	}
	if err := claim(-5, 2, 4, 20).Check(); err == nil {
		t.Fatal("negative n accepted")
	}
	if err := (ErdosStrausClaim{N: big.NewInt(5), X: big.NewInt(2), Y: big.NewInt(4)}).Check(); err == nil {
		t.Fatal("missing z accepted")
	}
}

func TestCheckOverflowScaleExact(t *testing.T) {
	// Exactness far beyond int64: scale a known witness by 10^30. If
	// 4/n = 1/x+1/y+1/z then 4/(kn) = 1/(kx)+1/(ky)+1/(kz).
	k, _ := new(big.Int).SetString("1000000000000000000000000000000", 10)
	base := claim(5, 2, 4, 20)
	scaled := ErdosStrausClaim{
		N: new(big.Int).Mul(base.N, k), X: new(big.Int).Mul(base.X, k),
		Y: new(big.Int).Mul(base.Y, k), Z: new(big.Int).Mul(base.Z, k),
	}
	if err := scaled.Check(); err != nil {
		t.Fatalf("scaled witness rejected (must be exact at any magnitude): %v", err)
	}
	// And a scaled near-miss must still fail: perturb z by 1.
	scaled.Z.Add(scaled.Z, big.NewInt(1))
	if err := scaled.Check(); err == nil {
		t.Fatal("perturbed large witness accepted (tolerance has crept in)")
	}
}

func TestParseErdosStraus(t *testing.T) {
	c, err := ParseErdosStraus(" 5, 2, 4, 20 ")
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if err := c.Check(); err != nil {
		t.Fatalf("parsed witness must check: %v", err)
	}
	if _, err := ParseErdosStraus("5,2,4"); err == nil {
		t.Fatal("three values accepted")
	}
	if _, err := ParseErdosStraus("5,2,4,twenty"); err == nil {
		t.Fatal("non-integer accepted")
	}
	if got := c.Canonical(); !strings.Contains(got, `"n":"5"`) || !strings.Contains(got, CheckerName) {
		t.Fatalf("canonical form must carry checker identity and values: %s", got)
	}
}
