package witness

import (
	"errors"
	"strings"
	"testing"
)

func TestExecuteAttemptEqualDenominator(t *testing.T) {
	// n=7: x = ceil(21/4) = 6 — the deliberately failing probe shape.
	claim, err := ExecuteAttempt("equal-denominator", map[string]string{"n": "7"})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if got := claim.Canonical(); !strings.Contains(got, `"n":"7","x":"6","y":"6","z":"6"`) {
		t.Fatalf("wrong tuple: %s", got)
	}
	if claim.Check() == nil {
		t.Fatal("n=7 equal-denominator probe must fail the identity (4 does not divide 21)")
	}
	// n=8: x = 6, 3/6 = 1/2 = 4/8 — exact.
	claim, err = ExecuteAttempt("equal-denominator", map[string]string{"n": "8"})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if err := claim.Check(); err != nil {
		t.Fatalf("n=8 equal-denominator must be witness-valid: %v", err)
	}
}

func TestExecuteAttemptGreedy(t *testing.T) {
	// n=7, offset 0: x=2, r=1/14, y0=14 leaves zero, y=15 leaves 1/210.
	claim, err := ExecuteAttempt("greedy", map[string]string{"n": "7"})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if got := claim.Canonical(); !strings.Contains(got, `"n":"7","x":"2","y":"15","z":"210"`) {
		t.Fatalf("wrong tuple: %s", got)
	}
	if err := claim.Check(); err != nil {
		t.Fatalf("greedy n=7 must be witness-valid: %v", err)
	}
	// n=8, offset 0: x=2 and 4/8 − 1/2 = 0 — nothing to expand: abstains.
	_, err = ExecuteAttempt("greedy", map[string]string{"n": "8"})
	var abst AbstentionError
	if !errors.As(err, &abst) {
		t.Fatalf("want AbstentionError, got %v", err)
	}
}

func TestExecuteAttemptDeterministicAndStrict(t *testing.T) {
	a, err := ExecuteAttempt("greedy", map[string]string{"n": "7", "x0_offset": "1"})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	b, err := ExecuteAttempt("greedy", map[string]string{"n": "7", "x0_offset": "1"})
	if err != nil {
		t.Fatalf("re-execute: %v", err)
	}
	if a.Canonical() != b.Canonical() {
		t.Fatalf("nondeterministic: %s vs %s", a.Canonical(), b.Canonical())
	}
	if _, err := ExecuteAttempt("nope", map[string]string{"n": "7"}); err == nil {
		t.Fatal("unknown procedure must be an input error")
	}
	if _, err := ExecuteAttempt("equal-denominator", map[string]string{"n": "7", "junk": "1"}); err == nil {
		t.Fatal("unknown param must be an input error")
	}
	if _, err := ExecuteAttempt("equal-denominator", map[string]string{"n": "-3"}); err == nil {
		t.Fatal("non-positive n must be an input error")
	}
}

func TestCanonicalAttemptParamsRoundTrip(t *testing.T) {
	canonical := CanonicalAttemptParams(map[string]string{"x0_offset": "1", "n": "7"})
	if canonical != "n=7,x0_offset=1" {
		t.Fatalf("canonical params not sorted/stable: %q", canonical)
	}
	parsed, err := ParseAttemptParams(canonical)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if CanonicalAttemptParams(parsed) != canonical {
		t.Fatalf("round trip failed: %q", CanonicalAttemptParams(parsed))
	}
	if _, err := ParseAttemptParams("n=7,n=8"); err == nil {
		t.Fatal("duplicate keys must be rejected")
	}
}

func TestVerifyAttemptBinding(t *testing.T) {
	claim, err := ExecuteAttempt("equal-denominator", map[string]string{"n": "7"})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	tuple := claim.Canonical()
	if err := VerifyAttemptBinding("equal-denominator", AttemptExecutorVersion, "n=7", tuple); err != nil {
		t.Fatalf("true binding must recheck: %v", err)
	}
	// A tuple the procedure did not produce fails the recheck.
	other, _ := ExecuteAttempt("equal-denominator", map[string]string{"n": "11"})
	if err := VerifyAttemptBinding("equal-denominator", AttemptExecutorVersion, "n=7", other.Canonical()); err == nil {
		t.Fatal("mismatched tuple must fail the recheck")
	}
	// An unknown executor version is a refusal, not a mismatch.
	if err := VerifyAttemptBinding("equal-denominator", "v0", "n=7", tuple); err == nil || !strings.Contains(err.Error(), "cannot recompute") {
		t.Fatalf("version refusal expected, got %v", err)
	}
}
