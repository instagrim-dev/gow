package finite

import (
	"strings"
	"testing"
)

// A defect-free warrant: exhaustive certificate for exactly the rule and
// domain presented.
func TestWarrantAcceptsExactExhaustiveCertificate(t *testing.T) {
	d := dom4("x", "y")
	left := Unary{Op: OpNeg, X: Binary{Op: OpAdd, X: Var{"x"}, Y: Var{"y"}}}
	right := Binary{Op: OpAdd, X: Unary{Op: OpNeg, X: Var{"x"}}, Y: Unary{Op: OpNeg, X: Var{"y"}}}
	cert := AssessEquivalence(binding("neg distributes over add modulo 2^w", d), left, right)
	if defects := VerifyRuleWarrant(cert, left, right, d); len(defects) != 0 {
		t.Fatalf("expected no defects, got %v", defects)
	}
}

// The central rejection: instance agreement presented as exhaustive
// evidence. The INSTANCE_EVIDENCE_ONLY verdict must be named as
// insufficient, whatever the consumer hoped it meant.
func TestWarrantRejectsInstanceAgreementPresentedAsExhaustive(t *testing.T) {
	d := dom4("x")
	left := Unary{Op: OpShr, X: Unary{Op: OpShl, X: Var{"x"}}}
	right := Var{"x"}
	cert := AssessInstances(binding("shift round-trip, sampled favorably", d), left, right,
		[]Assignment{{"x": 0}, {"x": 1}, {"x": 7}})
	defects := VerifyRuleWarrant(cert, left, right, d)
	if len(defects) == 0 {
		t.Fatal("instance evidence must not warrant a domain equality")
	}
	joined := strings.Join(defects, " | ")
	if !strings.Contains(joined, "INSTANCE_EVIDENCE_ONLY") || !strings.Contains(joined, "does not claim exhaustive enumeration") {
		t.Fatalf("defects must name the verdict and the missing exhaustiveness: %s", joined)
	}
}

// A tampered certificate: the verdict and exhaustive flag claim domain
// equality, but the coverage accounting does not add up. Structural
// consistency must catch it before replay is even consulted.
func TestWarrantRejectsIncompleteEnumeration(t *testing.T) {
	d := dom4("x", "y")
	left := Binary{Op: OpXor, X: Var{"x"}, Y: Var{"y"}}
	right := Binary{Op: OpXor, X: Var{"y"}, Y: Var{"x"}}
	cert := AssessEquivalence(binding("xor commutes", d), left, right)
	cert.AssignmentsChecked = 100 // forged: fewer than the 256-assignment domain
	defects := VerifyRuleWarrant(cert, left, right, d)
	if len(defects) == 0 {
		t.Fatal("inconsistent coverage accounting must be a warrant defect")
	}
	if !strings.Contains(strings.Join(defects, " | "), "100 assignments checked of a domain of 256") {
		t.Fatalf("defect must state the exact inconsistency: %v", defects)
	}
}

// Binding drift, width: a certificate earned at width 4 presented for a
// width-8 rule. The identity may even be true at width 8 — the point is
// that THIS certificate does not warrant it.
func TestWarrantRejectsChangedWidth(t *testing.T) {
	d4 := dom4("x", "y")
	d8 := Domain{Width: 8, Vars: []string{"x", "y"}}
	left := Binary{Op: OpXor, X: Var{"x"}, Y: Var{"y"}}
	right := Binary{Op: OpXor, X: Var{"y"}, Y: Var{"x"}}
	cert := AssessEquivalence(binding("xor commutes", d4), left, right)
	defects := VerifyRuleWarrant(cert, left, right, d8)
	if len(defects) == 0 {
		t.Fatal("a width-4 certificate must not warrant a width-8 rule")
	}
	if !strings.Contains(strings.Join(defects, " | "), "width mismatch") {
		t.Fatalf("defect must name the width mismatch: %v", defects)
	}
}

// Binding drift, domain variables: a certificate for {x} presented for a
// {x, y} rule (a larger quantification than was checked).
func TestWarrantRejectsChangedDomainVariables(t *testing.T) {
	d1 := dom4("x")
	d2 := dom4("x", "y")
	left, right := Var{"x"}, Var{"x"}
	cert := AssessEquivalence(binding("reflexivity", d1), left, right)
	defects := VerifyRuleWarrant(cert, left, right, d2)
	if len(defects) == 0 {
		t.Fatal("a certificate over {x} must not warrant a rule quantified over {x, y}")
	}
}

// Binding drift, expressions: a certificate for one identity presented as
// warrant for a syntactically different rule.
func TestWarrantRejectsChangedExpressions(t *testing.T) {
	d := dom4("x", "y")
	certLeft := Binary{Op: OpXor, X: Var{"x"}, Y: Var{"y"}}
	certRight := Binary{Op: OpXor, X: Var{"y"}, Y: Var{"x"}}
	cert := AssessEquivalence(binding("xor commutes", d), certLeft, certRight)
	otherLeft := Binary{Op: OpAdd, X: Var{"x"}, Y: Var{"y"}}
	otherRight := Binary{Op: OpAdd, X: Var{"y"}, Y: Var{"x"}}
	defects := VerifyRuleWarrant(cert, otherLeft, otherRight, d)
	if len(defects) == 0 {
		t.Fatal("a certificate about xor must not warrant a rule about add")
	}
	if !strings.Contains(strings.Join(defects, " | "), "expression mismatch") {
		t.Fatalf("defect must name the expression mismatch: %v", defects)
	}
}

// Forged verdict: a certificate whose fields all claim success for a rule
// that is actually false must fall to independent replay.
func TestWarrantReplayCatchesForgedVerdict(t *testing.T) {
	d := dom4("x")
	left := Unary{Op: OpShr, X: Unary{Op: OpShl, X: Var{"x"}}}
	right := Var{"x"}
	forged := Certificate{
		Binding:            binding("shift round-trip (forged)", d),
		Left:               left.render(),
		Right:              right.render(),
		Verdict:            VerdictHoldsOnDomain,
		Exhaustive:         true,
		DomainSize:         16,
		AssignmentsChecked: 16,
	}
	defects := VerifyRuleWarrant(forged, left, right, d)
	if len(defects) == 0 {
		t.Fatal("replay must refuse a forged certificate for a false identity")
	}
	if !strings.Contains(strings.Join(defects, " | "), "replay does not reproduce the equality") {
		t.Fatalf("defect must come from replay: %v", defects)
	}
}

// A refuted certificate presented as warrant is rejected on verdict type
// and on the carried counterexample, independently.
func TestWarrantRejectsRefutedCertificate(t *testing.T) {
	d := dom4("x")
	left := Unary{Op: OpShr, X: Unary{Op: OpShl, X: Var{"x"}}}
	right := Var{"x"}
	cert := AssessEquivalence(binding("shift round-trip", d), left, right)
	defects := VerifyRuleWarrant(cert, left, right, d)
	if len(defects) < 2 {
		t.Fatalf("expected verdict-type and counterexample defects, got %v", defects)
	}
}
