package measure

import (
	"strings"
	"testing"
	"time"
)

// tracesWith builds n instances; solvedAt maps instance index -> the
// submission index (0-based) that succeeds; each instance gets subs
// submissions unless stopOnHit truncates after the success.
func tracesWith(n int, subs int, solvedAt map[int]int, stopOnHit bool) Traces {
	t := Traces{}
	for i := 0; i < n; i++ {
		id := instanceID(i)
		var trace []Submission
		hitAt, solved := solvedAt[i]
		for j := 0; j < subs; j++ {
			success := solved && j == hitAt
			trace = append(trace, Submission{Move: moveName(j), Success: success})
			if success && stopOnHit {
				break
			}
		}
		t[id] = trace
	}
	return t
}

func instanceID(i int) string { return "p" + string(rune('a'+i%26)) + string(rune('0'+i/26)) }
func moveName(j int) string   { return "L" + string(rune('1'+j)) }

func conds(budget int64) Conditions {
	return Conditions{Population: "20 fixed instances", Ordering: "L1,L2,L3", StoppingRule: "stop-on-hit", Budget: budget}
}

func invarianceBinding(unrestricted bool, lo, hi int64) Binding {
	return Binding{
		Sentence:        "the local hit rate is the budget-independent measurement",
		MetricNumerator: "verified local successes",
		MetricDenom:     "local submissions",
		Population:      "20 fixed instances",
		Budgets:         BudgetRange{Lo: lo, Hi: hi, Unrestricted: unrestricted},
		Ordering:        "L1,L2,L3",
		StoppingRule:    "stop-on-hit",
		Kind:            ClaimObservedRateInvariance,
	}
}

// The motivating shape: three instances solve at their second submission;
// budget 1 sees 0/20, budget 2 sees 3/40, budget 3 adds 17 failures for
// 3/57 (solved instances stop at their hit, so only unsolved instances
// contribute a third submission).
func m1Observations() []Observation {
	solved := map[int]int{0: 1, 1: 1, 2: 1} // instances 0..2 succeed at submission index 1
	return []Observation{
		{Conditions: conds(1), Traces: tracesWith(20, 1, solved, true)},
		{Conditions: conds(2), Traces: tracesWith(20, 2, solved, true)},
		{Conditions: conds(3), Traces: tracesWith(20, 3, solved, true)},
	}
}

func TestMotivatingCaseRefutesUnrestrictedInvariance(t *testing.T) {
	obs := m1Observations()
	// Shape sanity: exactly the audited tallies.
	for i, want := range []Counts{{0, 20}, {3, 40}, {3, 57}} {
		got := CountsFromTraces(obs[i].Traces)
		if got != want {
			t.Fatalf("budget %d tally = %+v, want %+v", i+1, got, want)
		}
	}
	cert := AssessObservedRateInvariance(invarianceBinding(true, 0, 0), obs)
	if cert.Verdict != VerdictRefuted {
		t.Fatalf("verdict = %s (%s), want REFUTED", cert.Verdict, cert.Reason)
	}
	if cert.TraceExtension != "verified" {
		t.Fatalf("trace extension = %s, want verified", cert.TraceExtension)
	}
	// Direction accounting: rise then fall, exactly as the sign rule says.
	if cert.Rows[0].Direction != RateRises || cert.Rows[1].Direction != RateFalls {
		t.Fatalf("directions = %d,%d, want rise then fall", cert.Rows[0].Direction, cert.Rows[1].Direction)
	}
	// Marginal yield of the second transition is 0/17.
	if !strings.Contains(cert.Rows[1].MarginalYield, "0/17") {
		t.Fatalf("marginal yield = %q, want 0/17", cert.Rows[1].MarginalYield)
	}
}

func TestAddOnlyFailuresExplainsDecreaseWithoutLostSuccesses(t *testing.T) {
	solved := map[int]int{0: 0, 1: 0, 2: 0}
	before := Observation{Conditions: conds(1), Traces: tracesWith(20, 1, solved, true)} // 3/20
	after := Observation{Conditions: conds(2), Traces: tracesWith(20, 2, solved, true)}  // 3/37
	cert := AssessObservedRateInvariance(invarianceBinding(true, 0, 0), []Observation{before, after})
	if cert.Verdict != VerdictRefuted {
		t.Fatalf("verdict = %s, want REFUTED", cert.Verdict)
	}
	row := cert.Rows[0]
	if row.Direction != RateFalls {
		t.Fatalf("direction = %d, want falls", row.Direction)
	}
	if row.SolvedLost != 0 {
		t.Fatalf("solved lost = %d, want 0 — a falling rate must not be reported as lost successes", row.SolvedLost)
	}
	if !strings.Contains(row.Accounting, "accumulated achievement is unchanged") {
		t.Fatalf("accounting %q does not separate accumulation from yield", row.Accounting)
	}
}

func TestProportionalBatchReportsUnchangedRate(t *testing.T) {
	// before: 2 successes / 4 submissions. added: 1 success / 2 submissions
	// (same proportion) => rate unchanged with a larger denominator.
	before := Traces{
		"a": {{Move: "L1", Success: true}, {Move: "L2", Success: false}},
		"b": {{Move: "L1", Success: true}, {Move: "L2", Success: false}},
	}
	after := Traces{
		"a": {{Move: "L1", Success: true}, {Move: "L2", Success: false}, {Move: "L3", Success: true}},
		"b": {{Move: "L1", Success: true}, {Move: "L2", Success: false}, {Move: "L3", Success: false}},
	}
	obs := []Observation{
		{Conditions: conds(2), Traces: before},
		{Conditions: conds(3), Traces: after},
	}
	cert := AssessObservedRateInvariance(invarianceBinding(true, 0, 0), obs)
	if cert.Verdict != VerdictHolds {
		t.Fatalf("verdict = %s (%s), want HOLDS_AT_COMPARED_POINTS", cert.Verdict, cert.Reason)
	}
	if !strings.Contains(cert.Rows[0].Accounting, "exactly the prior proportion") {
		t.Fatalf("accounting %q, want proportional-batch explanation", cert.Rows[0].Accounting)
	}
	// The scope guard: equality at tested points must not extend.
	if !strings.Contains(cert.Reason, "does not establish invariance over any untested budget") {
		t.Fatalf("reason %q lacks the tested-points scope guard", cert.Reason)
	}
}

func TestChangedOrderingRefusesBudgetAttribution(t *testing.T) {
	solved := map[int]int{0: 1}
	a := Observation{Conditions: conds(2), Traces: tracesWith(5, 2, solved, true)}
	b := Observation{Conditions: Conditions{Population: "20 fixed instances", Ordering: "L2,L1,L3", StoppingRule: "stop-on-hit", Budget: 3}, Traces: tracesWith(5, 3, solved, true)}
	cert := AssessObservedRateInvariance(invarianceBinding(true, 0, 0), []Observation{a, b})
	if cert.Verdict != VerdictInapplicable {
		t.Fatalf("verdict = %s, want INAPPLICABLE", cert.Verdict)
	}
	if !strings.Contains(cert.Reason, "budget alone") {
		t.Fatalf("reason %q must refuse attribution to budget alone", cert.Reason)
	}
}

func TestMutatedHistoryRejectsAppendOnlyDecomposition(t *testing.T) {
	before := Traces{"a": {{Move: "L1", Success: false}, {Move: "L2", Success: true}}}
	// Same instance, but the earlier verdict changed.
	afterMutated := Traces{"a": {{Move: "L1", Success: true}, {Move: "L2", Success: true}, {Move: "L3", Success: false}}}
	ext := ValidateExtension(before, afterMutated)
	if ext.Nested {
		t.Fatal("mutated earlier verdict must reject nesting")
	}
	// Removed observation likewise.
	afterShrunk := Traces{"a": {{Move: "L1", Success: false}}}
	if ValidateExtension(before, afterShrunk).Nested {
		t.Fatal("shrunk trace must reject nesting")
	}
	// The invariance assessor still compares fractions exactly, but reports
	// the rejected decomposition.
	obs := []Observation{
		{Conditions: conds(2), Traces: before},
		{Conditions: conds(3), Traces: afterMutated},
	}
	cert := AssessObservedRateInvariance(invarianceBinding(true, 0, 0), obs)
	if cert.TraceExtension != "not_nested" {
		t.Fatalf("trace extension = %s, want not_nested", cert.TraceExtension)
	}
	if len(cert.ExtensionViolations) == 0 {
		t.Fatal("expected recorded extension violations")
	}
	if !strings.Contains(cert.Rows[0].MarginalYield, "incremental decomposition rejected") {
		t.Fatalf("marginal yield %q must mark the rejected decomposition", cert.Rows[0].MarginalYield)
	}
}

func TestProbabilisticClaimIsNotAssessed(t *testing.T) {
	b := invarianceBinding(true, 0, 0)
	b.Kind = ClaimProbabilisticProperty
	b.Sentence = "the underlying success probability is budget-independent"
	cert := AssessProbabilistic(b)
	if cert.Verdict != VerdictNotAssessed {
		t.Fatalf("verdict = %s, want NOT_ASSESSED", cert.Verdict)
	}
	if !strings.Contains(cert.Reason, "different proportions") {
		t.Fatalf("reason %q must state that unequal observed proportions alone are not statistical refutation", cert.Reason)
	}
	// And the observed-rate assessor refuses the kind outright.
	got := AssessObservedRateInvariance(b, m1Observations())
	if got.Verdict != VerdictInapplicable {
		t.Fatalf("observed-rate assessor on probabilistic binding = %s, want INAPPLICABLE", got.Verdict)
	}
}

func TestZeroDenominatorStaysUnresolved(t *testing.T) {
	empty := Observation{Conditions: conds(1), Traces: Traces{"a": {}}}
	solved := map[int]int{0: 0}
	nonEmpty := Observation{Conditions: conds(2), Traces: tracesWith(1, 1, solved, true)}
	// The empty trace is a prefix of anything with the same instance? Here
	// populations differ ("a" vs generated id), so make them match:
	nonEmpty.Traces = Traces{"a": {{Move: "L1", Success: true}}}
	cert := AssessObservedRateInvariance(invarianceBinding(true, 0, 0), []Observation{empty, nonEmpty})
	if cert.Verdict != VerdictUnresolved {
		t.Fatalf("verdict = %s (%s), want UNRESOLVED — a zero-denominator rate must not become a failure or a rate", cert.Verdict, cert.Reason)
	}
}

func TestRangeScopedClaimRejectsOutOfRangePoints(t *testing.T) {
	obs := m1Observations() // budgets 1,2,3
	cert := AssessObservedRateInvariance(invarianceBinding(false, 3, 10), obs)
	if cert.Verdict != VerdictInapplicable {
		t.Fatalf("verdict = %s, want INAPPLICABLE — budgets 1,2 lie outside the claimed range [3,10]", cert.Verdict)
	}
	// Inside the range with a plateau: budgets 3->4 with no local additions
	// hold at the compared points only.
	solved := map[int]int{0: 1, 1: 1, 2: 1}
	b3 := Observation{Conditions: conds(3), Traces: tracesWith(20, 3, solved, true)}
	b4 := Observation{Conditions: conds(4), Traces: tracesWith(20, 3, solved, true)} // plateau: no 4th local submission
	plateau := AssessObservedRateInvariance(invarianceBinding(false, 3, 10), []Observation{b3, b4})
	if plateau.Verdict != VerdictHolds {
		t.Fatalf("plateau verdict = %s (%s), want HOLDS_AT_COMPARED_POINTS", plateau.Verdict, plateau.Reason)
	}
	if !strings.Contains(plateau.Reason, "AT THESE POINTS") {
		t.Fatalf("plateau reason %q must scope equality to the compared points", plateau.Reason)
	}
}

func TestSolvedMonotonicityHoldsAndRefutes(t *testing.T) {
	solved := map[int]int{0: 0, 1: 1}
	before := Observation{Conditions: conds(2), Traces: tracesWith(4, 2, solved, true)}
	after := Observation{Conditions: conds(3), Traces: tracesWith(4, 3, solved, true)}
	b := invarianceBinding(true, 0, 0)
	b.Kind = ClaimSolvedMonotonicity
	b.Sentence = "increasing the budget cannot lose already-solved instances"
	cert := AssessSolvedMonotonicity(b, before, after)
	if cert.Verdict != VerdictHolds {
		t.Fatalf("verdict = %s (%s), want HOLDS", cert.Verdict, cert.Reason)
	}
	// Non-nested pair: theorem inapplicable, not failed.
	otherStrategy := Observation{Conditions: conds(3), Traces: tracesWith(4, 3, map[int]int{2: 0}, true)}
	inapp := AssessSolvedMonotonicity(b, before, otherStrategy)
	if inapp.Verdict != VerdictInapplicable {
		t.Fatalf("non-nested verdict = %s, want INAPPLICABLE", inapp.Verdict)
	}
}

func TestCertificateRenderAndCheckRecord(t *testing.T) {
	cert := AssessObservedRateInvariance(invarianceBinding(true, 0, 0), m1Observations())
	out := cert.Render()
	for _, want := range []string{
		"Verdict: REFUTED",
		"Statistical anomaly: NOT ASSESSED",
		"Change in underlying success probability: NOT ASSESSED",
		"Change in uncertainty: NOT ASSESSED",
		"previously solved lost: 0",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("rendered certificate missing %q:\n%s", want, out)
		}
	}
	rec, err := cert.ToCheckRecord("chk-1", "m1-budget-sweep", "test", "traces: synthetic m1 shape", "go test", time.Now(), time.Now())
	if err != nil {
		t.Fatalf("ToCheckRecord: %v", err)
	}
	if rec.Outcome != "completed" || rec.ProcedureRevision != CheckerVersion {
		t.Fatalf("check record outcome=%q revision=%q", rec.Outcome, rec.ProcedureRevision)
	}
	// A refusal renders as blocked, not completed.
	prob := AssessProbabilistic(Binding{Kind: ClaimProbabilisticProperty, Sentence: "p is budget-independent"})
	rec2, err := prob.ToCheckRecord("chk-2", "prob", "test", "", "", time.Now(), time.Now())
	if err != nil {
		t.Fatalf("ToCheckRecord: %v", err)
	}
	if rec2.Outcome != "blocked" {
		t.Fatalf("NOT_ASSESSED must map to blocked, got %q", rec2.Outcome)
	}
}
