package finite

import (
	"strings"
	"testing"
	"time"
)

func timeStamp() time.Time { return time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC) }

// dom4 is the roadmap's G4-lite seed domain: 4-bit words, three variables.
func dom4(vars ...string) Domain { return Domain{Width: 4, Vars: vars} }

func binding(sentence string, d Domain) Binding { return Binding{Sentence: sentence, Domain: d} }

// A known valid identity must earn the domain-scoped verdict with full
// enumeration accounting: x XOR y == (x OR y) AND NOT(x AND y).
func TestValidIdentityHoldsOnDomain(t *testing.T) {
	d := dom4("x", "y")
	left := Binary{Op: OpXor, X: Var{"x"}, Y: Var{"y"}}
	right := Binary{
		Op: OpAnd,
		X:  Binary{Op: OpOr, X: Var{"x"}, Y: Var{"y"}},
		Y:  Unary{Op: OpNot, X: Binary{Op: OpAnd, X: Var{"x"}, Y: Var{"y"}}},
	}
	cert := AssessEquivalence(binding("xor decomposes over or/and/not", d), left, right)
	if cert.Verdict != VerdictHoldsOnDomain {
		t.Fatalf("verdict %s: %s", cert.Verdict, cert.Reason)
	}
	if !cert.Exhaustive || cert.AssignmentsChecked != 256 || cert.DomainSize != 256 {
		t.Fatalf("expected exhaustive 256/256, got %d/%d exhaustive=%v", cert.AssignmentsChecked, cert.DomainSize, cert.Exhaustive)
	}
	// The strongest verdict must still name its scope, not general truth.
	if !strings.Contains(cert.Reason, "declared domain only") {
		t.Fatalf("domain-scoped reason missing: %s", cert.Reason)
	}
}

// The roadmap's canonical tempting invalid rewrite: shr1(shl1(x)) == x is a
// real-integer identity that fails in machine words (the high bit is
// discarded). The refutation must carry the exact first counterexample in
// canonical order: x=8 at width 4 (shl1(8)=0, shr1(0)=0 != 8).
func TestMachineWordRefutesRealIntegerIdentity(t *testing.T) {
	d := dom4("x")
	left := Unary{Op: OpShr, X: Unary{Op: OpShl, X: Var{"x"}}}
	right := Var{"x"}
	cert := AssessEquivalence(binding("shift left then right restores the value", d), left, right)
	if cert.Verdict != VerdictRefuted {
		t.Fatalf("verdict %s: %s", cert.Verdict, cert.Reason)
	}
	if cert.Counterexample == nil || cert.Counterexample.Assignment != "x=8" {
		t.Fatalf("expected canonical first counterexample x=8, got %+v", cert.Counterexample)
	}
	if cert.Counterexample.Left != 0 || cert.Counterexample.Right != 8 {
		t.Fatalf("counterexample evaluations wrong: %+v", cert.Counterexample)
	}
}

// De Morgan under NEG instead of NOT is a tempting near-identity:
// neg(x add y) == neg(x) add neg(y)? That one actually HOLDS mod 2^w, so
// use a genuinely false one: not(x add y) == not(x) add not(y) — refuted.
func TestNotDoesNotDistributeOverAdd(t *testing.T) {
	d := dom4("x", "y")
	left := Unary{Op: OpNot, X: Binary{Op: OpAdd, X: Var{"x"}, Y: Var{"y"}}}
	right := Binary{Op: OpAdd, X: Unary{Op: OpNot, X: Var{"x"}}, Y: Unary{Op: OpNot, X: Var{"y"}}}
	cert := AssessEquivalence(binding("not distributes over add", d), left, right)
	if cert.Verdict != VerdictRefuted {
		t.Fatalf("verdict %s: %s", cert.Verdict, cert.Reason)
	}
}

// Negation distributing over addition modulo 2^w is true and must pass:
// a control that the refuter is not refuse-everything.
func TestNegDistributesOverAddModulo(t *testing.T) {
	d := dom4("x", "y")
	left := Unary{Op: OpNeg, X: Binary{Op: OpAdd, X: Var{"x"}, Y: Var{"y"}}}
	right := Binary{Op: OpAdd, X: Unary{Op: OpNeg, X: Var{"x"}}, Y: Unary{Op: OpNeg, X: Var{"y"}}}
	cert := AssessEquivalence(binding("neg distributes over add modulo 2^w", d), left, right)
	if cert.Verdict != VerdictHoldsOnDomain {
		t.Fatalf("verdict %s: %s", cert.Verdict, cert.Reason)
	}
}

// An undeclared free variable is a premise failure identified by name —
// INAPPLICABLE, not a guess and not a runtime surprise.
func TestUndeclaredVariableIsInapplicable(t *testing.T) {
	d := dom4("x")
	left := Binary{Op: OpAdd, X: Var{"x"}, Y: Var{"w"}}
	cert := AssessEquivalence(binding("uses an undeclared variable", d), left, Var{"x"})
	if cert.Verdict != VerdictInapplicable {
		t.Fatalf("verdict %s: %s", cert.Verdict, cert.Reason)
	}
	if !strings.Contains(cert.Reason, `"w"`) {
		t.Fatalf("failed premise not identified: %s", cert.Reason)
	}
}

// A domain above the exhaustiveness cap is UNRESOLVED: the tool refuses to
// substitute sampling for enumeration rather than inventing confidence.
func TestOversizedDomainIsUnresolved(t *testing.T) {
	d := Domain{Width: 8, Vars: []string{"a", "b", "c"}} // 256^3 = 16,777,216
	cert := AssessEquivalence(binding("too large to exhaust", d), Var{"a"}, Var{"a"})
	if cert.Verdict != VerdictUnresolved {
		t.Fatalf("verdict %s: %s", cert.Verdict, cert.Reason)
	}
	if cert.AssignmentsChecked != 0 {
		t.Fatalf("no enumeration should have run, checked %d", cert.AssignmentsChecked)
	}
}

// T0 instance-to-universal rejection, enforced as a type of outcome:
// agreeing instances yield INSTANCE_EVIDENCE_ONLY with an explicit scope
// guard, never the domain verdict — even when the identity is in fact
// false elsewhere in the domain.
func TestInstanceAgreementNeverBecomesDomainEquivalence(t *testing.T) {
	d := dom4("x")
	left := Unary{Op: OpShr, X: Unary{Op: OpShl, X: Var{"x"}}} // false at x=8..15
	right := Var{"x"}
	instances := []Assignment{{"x": 0}, {"x": 1}, {"x": 7}} // all agree
	cert := AssessInstances(binding("shift round-trip, sampled favorably", d), left, right, instances)
	if cert.Verdict != VerdictInstanceOnly {
		t.Fatalf("verdict %s: %s", cert.Verdict, cert.Reason)
	}
	if cert.Exhaustive {
		t.Fatal("instance check must never claim exhaustiveness")
	}
	guardFound := false
	for _, n := range cert.NotAssessed {
		if strings.Contains(n, "never a domain certificate") {
			guardFound = true
		}
	}
	if !guardFound {
		t.Fatalf("instance certificate missing the domain scope guard: %v", cert.NotAssessed)
	}
}

// One refuting instance still refutes: a single admissible counterexample
// decides the universal claim negatively.
func TestSingleInstanceCounterexampleRefutes(t *testing.T) {
	d := dom4("x")
	left := Unary{Op: OpShr, X: Unary{Op: OpShl, X: Var{"x"}}}
	right := Var{"x"}
	cert := AssessInstances(binding("shift round-trip at the failing point", d), left, right, []Assignment{{"x": 12}})
	if cert.Verdict != VerdictRefuted {
		t.Fatalf("verdict %s: %s", cert.Verdict, cert.Reason)
	}
	if cert.Counterexample == nil || cert.Counterexample.Assignment != "x=12" {
		t.Fatalf("expected counterexample x=12, got %+v", cert.Counterexample)
	}
}

// An instance outside the declared width is a premise failure, not data.
func TestOutOfDomainInstanceIsInapplicable(t *testing.T) {
	d := dom4("x")
	cert := AssessInstances(binding("out-of-range instance", d), Var{"x"}, Var{"x"}, []Assignment{{"x": 16}})
	if cert.Verdict != VerdictInapplicable {
		t.Fatalf("verdict %s: %s", cert.Verdict, cert.Reason)
	}
}

// Structurally invalid expressions are applicability refusals, never
// panics: nil expressions, nil children, foreign node types, and
// unbounded nesting are all rejected before rendering or traversal.
func TestMalformedExpressionsRefusedNotPanicked(t *testing.T) {
	d := dom4("x")

	if cert := AssessEquivalence(binding("nil left", d), nil, Var{"x"}); cert.Verdict != VerdictInapplicable {
		t.Fatalf("nil expression must be INAPPLICABLE, got %s", cert.Verdict)
	}
	if cert := AssessEquivalence(binding("nil unary child", d), Unary{Op: OpNot, X: nil}, Var{"x"}); cert.Verdict != VerdictInapplicable {
		t.Fatalf("nil unary child must be INAPPLICABLE, got %s", cert.Verdict)
	}
	if cert := AssessEquivalence(binding("nil binary child", d), Binary{Op: OpAdd, X: Var{"x"}, Y: nil}, Var{"x"}); cert.Verdict != VerdictInapplicable {
		t.Fatalf("nil binary child must be INAPPLICABLE, got %s", cert.Verdict)
	}
	if cert := AssessInstances(binding("nil in instances", d), nil, Var{"x"}, []Assignment{{"x": 1}}); cert.Verdict != VerdictInapplicable {
		t.Fatalf("nil expression must be INAPPLICABLE in instance checks, got %s", cert.Verdict)
	}

	deep := Expr(Var{"x"})
	for i := 0; i <= MaxExprDepth; i++ {
		deep = Unary{Op: OpNot, X: deep}
	}
	cert := AssessEquivalence(binding("too deep", d), deep, Var{"x"})
	if cert.Verdict != VerdictInapplicable {
		t.Fatalf("over-deep expression must be INAPPLICABLE, got %s", cert.Verdict)
	}
	if !strings.Contains(cert.Reason, "maximum depth") {
		t.Fatalf("refusal must name the depth bound: %s", cert.Reason)
	}

	// The warrant boundary is guarded the same way.
	good := AssessEquivalence(binding("reflexivity", d), Var{"x"}, Var{"x"})
	if defects := VerifyRuleWarrant(good, Unary{Op: OpNot, X: nil}, Var{"x"}, d); len(defects) == 0 {
		t.Fatal("a malformed rule expression must be a warrant defect, not a panic")
	}
}

// Instance certificates account for coverage consistently: the domain size
// is populated, exhaustiveness stays false, and rendering shows N of the
// real domain, never "N of 0".
func TestInstanceCertificateCoverageAccounting(t *testing.T) {
	d := dom4("x")
	cert := AssessInstances(binding("one instance", d), Var{"x"}, Var{"x"}, []Assignment{{"x": 5}})
	if cert.Verdict != VerdictInstanceOnly {
		t.Fatalf("verdict %s: %s", cert.Verdict, cert.Reason)
	}
	if cert.DomainSize != 16 || cert.AssignmentsChecked != 1 || cert.Exhaustive {
		t.Fatalf("coverage accounting wrong: %d of %d, exhaustive=%v", cert.AssignmentsChecked, cert.DomainSize, cert.Exhaustive)
	}
	if !strings.Contains(cert.Render(), "Assignments checked: 1 of 16") {
		t.Fatalf("render must show real coverage: %s", cert.Render())
	}
}

// A refuting instance decides the universal claim over the domain
// negatively — so the "domain NOT ASSESSED" guard must not appear on a
// REFUTED instance certificate, while it must appear on an agreeing one.
func TestInstanceScopeGuardFollowsOutcome(t *testing.T) {
	d := dom4("x")
	left := Unary{Op: OpShr, X: Unary{Op: OpShl, X: Var{"x"}}}
	right := Var{"x"}

	refuted := AssessInstances(binding("refuting point", d), left, right, []Assignment{{"x": 12}})
	if refuted.Verdict != VerdictRefuted {
		t.Fatalf("verdict %s", refuted.Verdict)
	}
	for _, n := range refuted.NotAssessed {
		if strings.Contains(n, "never a domain certificate") {
			t.Fatalf("a refutation DID assess the domain claim; the contradictory guard must be absent: %v", refuted.NotAssessed)
		}
	}

	agreeing := AssessInstances(binding("agreeing points", d), left, right, []Assignment{{"x": 1}})
	if agreeing.Verdict != VerdictInstanceOnly {
		t.Fatalf("verdict %s", agreeing.Verdict)
	}
	guard := false
	for _, n := range agreeing.NotAssessed {
		if strings.Contains(n, "never a domain certificate") {
			guard = true
		}
	}
	if !guard {
		t.Fatalf("agreeing instances must carry the domain scope guard: %v", agreeing.NotAssessed)
	}
}

// Refusals adapt to blocked check records; decisions adapt to completed
// ones — a refusal to assess must not read as an executed check.
func TestCheckRecordOutcomeSeparatesRefusalFromDecision(t *testing.T) {
	d := dom4("x")
	refuted := AssessEquivalence(binding("shift round-trip", d), Unary{Op: OpShr, X: Unary{Op: OpShl, X: Var{"x"}}}, Var{"x"})
	blockedCert := AssessEquivalence(binding("too big", Domain{Width: 8, Vars: []string{"a", "b", "c"}}), Var{"a"}, Var{"a"})

	recDone, err := refuted.ToCheckRecord("chk-1", "refuted", "test", "", "", timeStamp(), timeStamp())
	if err != nil {
		t.Fatal(err)
	}
	if recDone.Outcome != "completed" {
		t.Fatalf("a refutation is an executed decision; outcome %q", recDone.Outcome)
	}
	recBlocked, err := blockedCert.ToCheckRecord("chk-2", "oversized", "test", "", "", timeStamp(), timeStamp())
	if err != nil {
		t.Fatal(err)
	}
	if recBlocked.Outcome != "blocked" {
		t.Fatalf("an unresolved refusal must be blocked, not completed; outcome %q", recBlocked.Outcome)
	}
}
