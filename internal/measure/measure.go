// Package measure is a claim-aware measurement checker: a deterministic
// verifier for claims that attach properties to recorded observations
// (rates, budgets, traces). It is NOT an anomaly detector. It determines
// what changed in the recorded observations, applies the relevant exact
// mathematical rule, and emits an explanatory certificate that answers
// exactly one question — so a certificate answering one question cannot be
// silently consumed as an answer to another (a falling observed rate is not
// "the strategy got worse"; it is not "uncertainty increased").
//
// The package distinguishes three kinds of claims and refuses to conflate
// them:
//
//   - Observed-rate invariance across conditions: decided by exact
//     comparison of recorded fractions under declared comparison
//     conditions (cross-multiplication in big.Int; no floats).
//   - Solved-set monotonicity under budget extension: decided by trace
//     prefix verification plus success retention — a conditional theorem
//     whose precondition (nested executions) is checked, never assumed.
//   - Probabilistic or uncertainty claims: NOT ASSESSED here. Two
//     different observed rates do not by themselves disprove a constant
//     underlying probability; deciding that requires a declared
//     probabilistic target, sampling assumptions, and a statistical
//     procedure this package does not own.
//
// It is pure: no SQL, no CLI, no provider concepts. Certificates are
// evidence for the review machinery (internal/review); this package grants
// itself no authority over publication or experiment policy.
package measure

import (
	"fmt"
	"math/big"
	"sort"
)

// Submission is one counted submission in an instance's trace.
type Submission struct {
	Move    string // optional move label; compared as part of the prefix
	Success bool
}

// Traces maps instance ID to that instance's ordered submissions.
type Traces map[string][]Submission

// Conditions are the declared conditions a set of traces was produced
// under. Everything except Budget must be identical between two compared
// records for a change to be attributable to budget alone.
type Conditions struct {
	Population   string
	Ordering     string
	StoppingRule string
	Budget       int64
}

// ClaimKind selects the check a binding is entitled to.
type ClaimKind string

const (
	// ClaimObservedRateInvariance: "this observed rate is unchanged across
	// these budgets" — exact comparison of recorded fractions.
	ClaimObservedRateInvariance ClaimKind = "observed_rate_invariance"
	// ClaimSolvedMonotonicity: "increasing this execution's budget cannot
	// lose already-solved instances" — trace extension + success retention.
	ClaimSolvedMonotonicity ClaimKind = "solved_monotonicity"
	// ClaimProbabilisticProperty: claims about an underlying probability or
	// uncertainty. This package refuses to assess them.
	ClaimProbabilisticProperty ClaimKind = "probabilistic_property"
)

// BudgetRange is the range over which an invariance claim quantifies.
type BudgetRange struct {
	Lo, Hi       int64
	Unrestricted bool // an unqualified label quantifies over all budgets
}

// Binding records the exact claim under assessment. Certificates carry the
// binding so the certificate cannot drift from the sentence it decides.
type Binding struct {
	Sentence        string // exact sentence from the manuscript/record
	MetricNumerator string // e.g. "verified local-move successes"
	MetricDenom     string // e.g. "local-move submissions"
	Population      string // instance population description
	Budgets         BudgetRange
	Ordering        string
	StoppingRule    string
	Kind            ClaimKind
}

// Counts is an exact success/submission tally.
type Counts struct {
	Successes   int64
	Submissions int64
}

// Rate renders the exact fraction, or "undefined" for a zero denominator.
func (c Counts) Rate() string {
	if c.Submissions == 0 {
		return "undefined (0 submissions)"
	}
	r := new(big.Rat).SetFrac64(c.Successes, c.Submissions)
	return fmt.Sprintf("%d/%d (= %s)", c.Successes, c.Submissions, r.FloatString(5))
}

// ExtensionCheck is the applicability record for the incremental
// decomposition. Aggregate counts alone cannot establish nesting; this is
// computed per instance from the traces.
type ExtensionCheck struct {
	Nested     bool
	Violations []string // per-instance reasons when not nested
}

// ValidateExtension checks that `after` is a genuine extension of `before`:
// identical instance population, and for every instance the earlier trace
// is a prefix of the later trace with unchanged verdicts and moves. A
// budget-aware algorithm may legitimately produce non-nested executions;
// then the incremental theorem simply does not apply.
func ValidateExtension(before, after Traces) ExtensionCheck {
	check := ExtensionCheck{Nested: true}
	ids := make([]string, 0, len(before))
	for id := range before {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		b := before[id]
		a, ok := after[id]
		if !ok {
			check.Nested = false
			check.Violations = append(check.Violations, fmt.Sprintf("instance %s present before, absent after: earlier observation removed", id))
			continue
		}
		if len(a) < len(b) {
			check.Nested = false
			check.Violations = append(check.Violations, fmt.Sprintf("instance %s trace shrank from %d to %d submissions", id, len(b), len(a)))
			continue
		}
		for i := range b {
			if a[i].Success != b[i].Success {
				check.Nested = false
				check.Violations = append(check.Violations, fmt.Sprintf("instance %s submission %d verdict changed (%v -> %v)", id, i, b[i].Success, a[i].Success))
			}
			if a[i].Move != b[i].Move {
				check.Nested = false
				check.Violations = append(check.Violations, fmt.Sprintf("instance %s submission %d move changed (%q -> %q)", id, i, b[i].Move, a[i].Move))
			}
		}
	}
	for id := range after {
		if _, ok := before[id]; !ok {
			check.Nested = false
			check.Violations = append(check.Violations, fmt.Sprintf("instance %s appears only after: population changed, not extended", id))
		}
	}
	sort.Strings(check.Violations)
	return check
}

// RateDirection is the exact sign consequence of the weighted-average
// identity: for S successes among N>0 submissions extended by s among m>0,
//
//	(S+s)/(N+m) − S/N  has the sign of  N·s − S·m.
//
// The new cumulative rate is the weighted average of the old rate and the
// added batch's rate; it rises, stays equal, or falls according as the
// added batch outperforms, matches, or underperforms the old average.
type RateDirection int

const (
	RateFalls RateDirection = -1
	RateEqual RateDirection = 0
	RateRises RateDirection = 1
)

// Transition is one exact before/added comparison.
type Transition struct {
	Before Counts
	Added  Counts
}

// After is the cumulative tally.
func (t Transition) After() Counts {
	return Counts{
		Successes:   t.Before.Successes + t.Added.Successes,
		Submissions: t.Before.Submissions + t.Added.Submissions,
	}
}

// Sign computes sign(N·s − S·m) exactly in big.Int. It is only meaningful
// when both denominators are positive; callers must gate on that.
func (t Transition) Sign() RateDirection {
	ns := new(big.Int).Mul(big.NewInt(t.Before.Submissions), big.NewInt(t.Added.Successes))
	sm := new(big.Int).Mul(big.NewInt(t.Before.Successes), big.NewInt(t.Added.Submissions))
	return RateDirection(ns.Cmp(sm))
}

// CountsFromTraces derives the tally from traces; counts are never accepted
// on faith when traces are available.
func CountsFromTraces(t Traces) Counts {
	var c Counts
	for _, subs := range t {
		for _, s := range subs {
			c.Submissions++
			if s.Success {
				c.Successes++
			}
		}
	}
	return c
}

// SolvedInstances counts instances with at least one successful submission.
func SolvedInstances(t Traces) int64 {
	var n int64
	for _, subs := range t {
		for _, s := range subs {
			if s.Success {
				n++
				break
			}
		}
	}
	return n
}

// solvedSet returns the set of solved instance IDs.
func solvedSet(t Traces) map[string]bool {
	out := map[string]bool{}
	for id, subs := range t {
		for _, s := range subs {
			if s.Success {
				out[id] = true
				break
			}
		}
	}
	return out
}
