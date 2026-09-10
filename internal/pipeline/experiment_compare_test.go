package pipeline

import (
	"sort"
	"strings"
	"testing"
)

// TestArmRecoveryStatus locks the compare-time F5 gate: a non-recovery is a
// decisive negative ("no_recovery") only when EVERY membership proposal was
// decisively assessed; anything less (unknown/unassessed, or an empty arm) is
// "inconclusive" and must never be reported as a negative.
func TestArmRecoveryStatus(t *testing.T) {
	cases := []struct {
		name string
		arm  ExperimentArmView
		want string
	}{
		{"recovered wins over everything", ExperimentArmView{Recovered: true, ProposalCount: 3, DecisiveCount: 1, UnknownCount: 2}, "recovered"},
		{"fully decisive no-recovery is negative", ExperimentArmView{Recovered: false, ProposalCount: 3, DecisiveCount: 3}, "no_recovery"},
		{"unknown proposals => inconclusive", ExperimentArmView{Recovered: false, ProposalCount: 3, DecisiveCount: 2, UnknownCount: 1}, "inconclusive"},
		{"unassessed proposals => inconclusive", ExperimentArmView{Recovered: false, ProposalCount: 3, DecisiveCount: 2, UnassessedCount: 1}, "inconclusive"},
		{"empty arm => inconclusive, not a negative", ExperimentArmView{Recovered: false, ProposalCount: 0}, "inconclusive"},
	}
	for _, c := range cases {
		if got := armRecoveryStatus(c.arm); got != c.want {
			t.Errorf("%s: armRecoveryStatus = %q, want %q", c.name, got, c.want)
		}
	}
}

// TestRecoveryDeltaFromNeverCoercesInconclusive is the Leak-2 truth table: an
// inconclusive arm never contributes a recovery negative. A "*-only" / "neither"
// direction is emitted ONLY when the counterpart arm reached a decisive
// no_recovery.
func TestRecoveryDeltaFromNeverCoercesInconclusive(t *testing.T) {
	cases := []struct {
		base, treat, want string
	}{
		{"recovered", "recovered", "both"},
		{"no_recovery", "recovered", "treatment-only"},
		{"recovered", "no_recovery", "baseline-only"},
		{"no_recovery", "no_recovery", "neither"},
		// Inconclusive counterpart must NOT become a decisive negative.
		{"inconclusive", "recovered", "treatment-only-baseline-inconclusive"},
		{"recovered", "inconclusive", "baseline-only-treatment-inconclusive"},
		{"inconclusive", "no_recovery", "inconclusive"},
		{"no_recovery", "inconclusive", "inconclusive"},
		{"inconclusive", "inconclusive", "inconclusive"},
	}
	for _, c := range cases {
		got := recoveryDeltaFrom(c.base, c.treat)
		if got != c.want {
			t.Errorf("recoveryDeltaFrom(%q,%q) = %q, want %q", c.base, c.treat, got, c.want)
		}
		// Hard invariant: the negative directions require a decisive no_recovery
		// on the arm being called out; they may never fire against inconclusive.
		switch got {
		case "neither":
			if c.base != "no_recovery" || c.treat != "no_recovery" {
				t.Errorf("'neither' emitted without both arms decisive: base=%q treat=%q", c.base, c.treat)
			}
		case "baseline-only":
			if c.treat != "no_recovery" {
				t.Errorf("'baseline-only' emitted while treatment not decisively no_recovery: %q", c.treat)
			}
		case "treatment-only":
			if c.base != "no_recovery" {
				t.Errorf("'treatment-only' emitted while baseline not decisively no_recovery: %q", c.base)
			}
		}
	}
}

// TestArmAssessmentIdentityIsOrderSensitive locks finding 1: because a finite
// evaluation budget makes proposal order consequential, the experiment-identity
// manifest must distinguish two orderings that change the assessment, while an
// exact replay (same order, same assessments) still collides for idempotency.
func TestArmAssessmentIdentityIsOrderSensitive(t *testing.T) {
	// Reviewer's scenario: one target, two proposals, budget of one comparison.
	// Run 1: A (rank 0) does not recover; B (rank 1) recovers but is unassessed.
	run1 := []string{
		armAssessmentIdentityPart("b3", 0, "hashA", "decisive_no"),
		armAssessmentIdentityPart("b3", 1, "hashB", "unassessed"),
	}
	// Run 2: the SAME artifacts reranked — B (rank 0) recovers; A (rank 1)
	// unassessed. Different assessment outcome => must be a different identity.
	run2 := []string{
		armAssessmentIdentityPart("b3", 0, "hashB", "recovered"),
		armAssessmentIdentityPart("b3", 1, "hashA", "unassessed"),
	}
	if strings.Join(run1, "\n") == strings.Join(run2, "\n") {
		t.Fatal("reordering that changes the assessment must not collide in identity (finding 1)")
	}

	// Exact replay: identical order and assessments => identical manifest.
	replay := []string{
		armAssessmentIdentityPart("b3", 0, "hashA", "decisive_no"),
		armAssessmentIdentityPart("b3", 1, "hashB", "unassessed"),
	}
	if strings.Join(run1, "\n") != strings.Join(replay, "\n") {
		t.Fatal("exact replay must produce an identical identity manifest (idempotency)")
	}

	// The pre-fix set projection (sort + hash-only) would collide on these runs;
	// assert the two runs really do share the same proposal SET so the only thing
	// distinguishing them is the order-and-assessment the fix now encodes.
	setOf := func(parts []string) string {
		hashes := []string{}
		for _, p := range parts {
			hashes = append(hashes, strings.Split(p, "|")[2]) // proposal_hash field
		}
		sort.Strings(hashes)
		return strings.Join(hashes, ",")
	}
	if setOf(run1) != setOf(run2) {
		t.Fatal("test premise: the two runs share the same proposal SET (only order/assessment differ)")
	}
}

// TestCompareInterpretationDoesNotPromoteInconclusive locks finding 3 at the
// prose boundary: the interpretation for a delta where a counterpart arm is
// inconclusive must say "not decisively assessed / inconclusive", and must NOT
// assert the counterpart "did not recover" (a demonstrated negative). Only a
// decisive no_recovery earns "did not".
func TestCompareInterpretationDoesNotPromoteInconclusive(t *testing.T) {
	// Treatment recovered, baseline inconclusive: must be inconclusive-worded.
	msg := compareInterpretation("treatment-only-baseline-inconclusive", "b0", "b3")
	if !strings.Contains(msg, "inconclusive") || !strings.Contains(msg, "not decisively assessed") {
		t.Errorf("inconclusive baseline must be worded as inconclusive, got: %q", msg)
	}
	if strings.Contains(msg, "b0 did not") {
		t.Errorf("must not promote inconclusive baseline to a demonstrated negative, got: %q", msg)
	}
	// Only a decisive no_recovery earns the demonstrated "did not".
	neg := compareInterpretation("treatment-only", "b0", "b3")
	if !strings.Contains(neg, "b0 did not") {
		t.Errorf("decisive no_recovery baseline should read as a negative, got: %q", neg)
	}
	// Both inconclusive: no negative may be claimed.
	both := compareInterpretation("inconclusive", "b0", "b3")
	if !strings.Contains(both, "inconclusive") || strings.Contains(both, "did not") {
		t.Errorf("inconclusive/inconclusive must claim no negative, got: %q", both)
	}
}
