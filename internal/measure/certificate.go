package measure

import (
	"fmt"
	"math/big"
	"sort"
	"strings"
)

// Verdict states for the single question a certificate answers.
const (
	VerdictRefuted      = "REFUTED"
	VerdictHolds        = "HOLDS_AT_COMPARED_POINTS"
	VerdictUnresolved   = "UNRESOLVED"
	VerdictInapplicable = "INAPPLICABLE"
	VerdictNotAssessed  = "NOT_ASSESSED"
)

// Observation is one recorded execution under declared conditions.
type Observation struct {
	Conditions Conditions
	Traces     Traces
}

// Certificate is the machine-readable outcome plus a fixed-template
// explanation. It answers exactly the bound claim; the NotAssessed lines
// are scope guards, not disclaimers — they prevent a certificate answering
// one question from being consumed as an answer to another.
type Certificate struct {
	Binding Binding

	// Applicability
	ConditionsComparable bool
	ConditionNotes       []string
	TraceExtension       string // "verified" | "not_nested" | "not_checked"
	ExtensionViolations  []string

	// Quantities, kept separate so accumulation cannot masquerade as yield:
	// total achievement (solved instances), cumulative yield (rate), and
	// marginal yield (added batch's rate) are different quantities.
	Rows []CertificateRow

	// The one answered question.
	Question string
	Verdict  string
	Reason   string

	// Scope guards.
	NotAssessed []string
}

// CertificateRow is one exact transition accounting line.
type CertificateRow struct {
	Label         string
	Before        Counts
	Added         Counts
	After         Counts
	Direction     RateDirection
	SolvedBefore  int64
	SolvedAfter   int64
	SolvedLost    int64
	MarginalYield string // rate of the added batch alone
	Accounting    string // fixed-template explanation of the change
}

var standardNotAssessed = []string{
	"Statistical anomaly: NOT ASSESSED",
	"Change in underlying success probability: NOT ASSESSED (two different observed rates do not, by themselves, disprove a constant underlying probability)",
	"Change in uncertainty: NOT ASSESSED (uncertainty is a different quantity requiring its own definition and model)",
}

// AssessObservedRateInvariance decides an observed-rate-invariance claim
// against two or more recorded observations, exactly. It refuses to
// attribute a change to budget when any other declared condition differs,
// and it never extends equality at tested points into invariance over an
// untested range.
func AssessObservedRateInvariance(b Binding, obs []Observation) Certificate {
	cert := Certificate{
		Binding:        b,
		Question:       "Is the observed rate unchanged across the compared budgets, under otherwise identical declared conditions?",
		TraceExtension: "not_checked",
		NotAssessed:    append([]string(nil), standardNotAssessed...),
	}
	if b.Kind != ClaimObservedRateInvariance {
		cert.Verdict = VerdictInapplicable
		cert.Reason = fmt.Sprintf("binding kind is %q; this assessor decides %q only", b.Kind, ClaimObservedRateInvariance)
		return cert
	}
	if len(obs) < 2 {
		cert.Verdict = VerdictUnresolved
		cert.Reason = "fewer than two observations: an invariance comparison needs at least two recorded conditions"
		return cert
	}

	ordered := append([]Observation(nil), obs...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].Conditions.Budget < ordered[j].Conditions.Budget })

	// Applicability: everything except budget must match, or the change is
	// not attributable to budget alone.
	cert.ConditionsComparable = true
	base := ordered[0].Conditions
	for _, o := range ordered[1:] {
		c := o.Conditions
		if c.Ordering != base.Ordering {
			cert.ConditionsComparable = false
			cert.ConditionNotes = append(cert.ConditionNotes, fmt.Sprintf("ordering differs (%q vs %q): change not attributable to budget alone", base.Ordering, c.Ordering))
		}
		if c.StoppingRule != base.StoppingRule {
			cert.ConditionsComparable = false
			cert.ConditionNotes = append(cert.ConditionNotes, fmt.Sprintf("stopping rule differs (%q vs %q): change not attributable to budget alone", base.StoppingRule, c.StoppingRule))
		}
		if c.Population != base.Population {
			cert.ConditionsComparable = false
			cert.ConditionNotes = append(cert.ConditionNotes, fmt.Sprintf("population differs (%q vs %q): not the same measured quantity", base.Population, c.Population))
		}
	}
	if !cert.ConditionsComparable {
		cert.Verdict = VerdictInapplicable
		cert.Reason = "compared records do not share the declared non-budget conditions; refusing attribution to budget alone"
		return cert
	}

	// Range applicability: compared budgets must lie inside the claimed range
	// (an unrestricted claim admits every budget).
	if !b.Budgets.Unrestricted {
		for _, o := range ordered {
			if o.Conditions.Budget < b.Budgets.Lo || o.Conditions.Budget > b.Budgets.Hi {
				cert.Verdict = VerdictInapplicable
				cert.Reason = fmt.Sprintf("budget %d lies outside the claim's declared range [%d,%d]; a range-scoped claim is not decided by out-of-range points", o.Conditions.Budget, b.Budgets.Lo, b.Budgets.Hi)
				return cert
			}
		}
	}

	// Exact transitions between successive budgets. Trace extension is
	// checked per pair; the incremental decomposition is only used where it
	// holds, and the fraction comparison is exact regardless.
	anyRefutation := false
	anyUndefined := false
	allNested := true
	for i := 1; i < len(ordered); i++ {
		lo, hi := ordered[i-1], ordered[i]
		loCounts, hiCounts := CountsFromTraces(lo.Traces), CountsFromTraces(hi.Traces)
		ext := ValidateExtension(lo.Traces, hi.Traces)
		row := CertificateRow{
			Label:        fmt.Sprintf("budget %d -> %d", lo.Conditions.Budget, hi.Conditions.Budget),
			Before:       loCounts,
			After:        hiCounts,
			SolvedBefore: SolvedInstances(lo.Traces),
			SolvedAfter:  SolvedInstances(hi.Traces),
		}
		if ext.Nested {
			row.Added = Counts{
				Successes:   hiCounts.Successes - loCounts.Successes,
				Submissions: hiCounts.Submissions - loCounts.Submissions,
			}
			row.MarginalYield = row.Added.Rate()
			loSolved, hiSolved := solvedSet(lo.Traces), solvedSet(hi.Traces)
			for id := range loSolved {
				if !hiSolved[id] {
					row.SolvedLost++
				}
			}
		} else {
			allNested = false
			cert.ExtensionViolations = append(cert.ExtensionViolations, ext.Violations...)
			row.MarginalYield = "undefined (executions not nested; incremental decomposition rejected)"
		}

		switch {
		case loCounts.Submissions == 0 || hiCounts.Submissions == 0:
			anyUndefined = true
			row.Accounting = "a zero-denominator rate is undefined; no rate comparison is invented for it"
		default:
			t := Transition{Before: loCounts, Added: Counts{Successes: hiCounts.Successes - loCounts.Successes, Submissions: hiCounts.Submissions - loCounts.Submissions}}
			if ext.Nested && t.Added.Submissions > 0 {
				row.Direction = t.Sign()
				switch row.Direction {
				case RateFalls:
					anyRefutation = true
					row.Accounting = fmt.Sprintf("added batch (%s) underperforms the prior average (%s); the added observations fully account for the fall — accumulated achievement is unchanged (solved %d -> %d, lost %d)", t.Added.Rate(), loCounts.Rate(), row.SolvedBefore, row.SolvedAfter, row.SolvedLost)
				case RateRises:
					anyRefutation = true
					row.Accounting = fmt.Sprintf("added batch (%s) outperforms the prior average (%s); the added observations fully account for the rise", t.Added.Rate(), loCounts.Rate())
				case RateEqual:
					row.Accounting = fmt.Sprintf("added batch arrives at exactly the prior proportion (%s); rate unchanged despite a larger denominator", t.Added.Rate())
				}
			} else {
				// Not nested (or nothing added): compare the recorded
				// fractions directly, exactly.
				cmp := exactRateCompare(loCounts, hiCounts)
				row.Direction = cmp
				if cmp != RateEqual {
					anyRefutation = true
					row.Accounting = "recorded fractions differ under the declared comparison conditions (no incremental decomposition applied)"
				} else {
					row.Accounting = "recorded fractions are exactly equal (no incremental decomposition applied)"
				}
			}
		}
		cert.Rows = append(cert.Rows, row)
	}
	if allNested {
		cert.TraceExtension = "verified"
	} else {
		cert.TraceExtension = "not_nested"
	}

	switch {
	case anyRefutation:
		cert.Verdict = VerdictRefuted
		if b.Budgets.Unrestricted {
			cert.Reason = "the observed rate differs between compared budgets; one admissible counterexample refutes the unrestricted invariance claim"
		} else {
			cert.Reason = "the observed rate differs between compared budgets inside the declared range"
		}
	case anyUndefined:
		cert.Verdict = VerdictUnresolved
		cert.Reason = "at least one compared rate is undefined (zero denominator); equality is not established and failure is not invented"
	default:
		cert.Verdict = VerdictHolds
		cert.Reason = "recorded fractions are exactly equal at every compared point; this supports equality AT THESE POINTS only and does not establish invariance over any untested budget"
	}
	return cert
}

// AssessSolvedMonotonicity decides "increasing this execution's budget
// cannot lose already-solved instances" for one before/after pair. The
// conditional theorem requires nested executions; a non-nested pair is
// INAPPLICABLE, not a failure.
func AssessSolvedMonotonicity(b Binding, before, after Observation) Certificate {
	cert := Certificate{
		Binding:     b,
		Question:    "Did the larger-budget execution retain every instance the smaller-budget execution had solved?",
		NotAssessed: append([]string(nil), standardNotAssessed...),
	}
	if b.Kind != ClaimSolvedMonotonicity {
		cert.Verdict = VerdictInapplicable
		cert.Reason = fmt.Sprintf("binding kind is %q; this assessor decides %q only", b.Kind, ClaimSolvedMonotonicity)
		return cert
	}
	ext := ValidateExtension(before.Traces, after.Traces)
	if !ext.Nested {
		cert.TraceExtension = "not_nested"
		cert.ExtensionViolations = ext.Violations
		cert.Verdict = VerdictInapplicable
		cert.Reason = "executions are not nested (a budget-aware algorithm may choose differently from the start); the conditional monotonicity theorem does not apply automatically"
		return cert
	}
	cert.TraceExtension = "verified"
	loSolved, hiSolved := solvedSet(before.Traces), solvedSet(after.Traces)
	var lost []string
	for id := range loSolved {
		if !hiSolved[id] {
			lost = append(lost, id)
		}
	}
	sort.Strings(lost)
	row := CertificateRow{
		Label:        fmt.Sprintf("budget %d -> %d", before.Conditions.Budget, after.Conditions.Budget),
		Before:       CountsFromTraces(before.Traces),
		After:        CountsFromTraces(after.Traces),
		SolvedBefore: SolvedInstances(before.Traces),
		SolvedAfter:  SolvedInstances(after.Traces),
		SolvedLost:   int64(len(lost)),
	}
	cert.Rows = append(cert.Rows, row)
	if len(lost) == 0 {
		cert.Verdict = VerdictHolds
		cert.Reason = "trace extension and success retention verified; no previously solved instance was lost (successes per submission may still fall — that is a different quantity)"
	} else {
		cert.Verdict = VerdictRefuted
		cert.Reason = fmt.Sprintf("previously solved instances lost under a verified extension: %s", strings.Join(lost, ", "))
	}
	return cert
}

// AssessProbabilistic refuses, with the reason recorded. Deciding a claim
// about an underlying probability or uncertainty requires a declared
// probabilistic target, sampling assumptions, and a statistical procedure;
// inequality of observed proportions alone is not statistical refutation.
func AssessProbabilistic(b Binding) Certificate {
	return Certificate{
		Binding:     b,
		Question:    "Does the underlying success probability (or uncertainty) change with budget?",
		Verdict:     VerdictNotAssessed,
		Reason:      "this deterministic checker decides observed statistics under specified procedures only; a probabilistic claim requires a declared target quantity, sampling assumptions, and an appropriate statistical procedure — ordinary random samples from one probability can produce different proportions",
		NotAssessed: append([]string(nil), standardNotAssessed...),
	}
}

// exactRateCompare compares S1/N1 vs S2/N2 by cross-multiplication in
// big.Int, so the verdict can never depend on rounding. Callers gate on
// positive denominators.
func exactRateCompare(a, b Counts) RateDirection {
	left := new(big.Int).Mul(big.NewInt(a.Successes), big.NewInt(b.Submissions))
	right := new(big.Int).Mul(big.NewInt(b.Successes), big.NewInt(a.Submissions))
	return RateDirection(right.Cmp(left))
}

// Render produces the fixed-template human-readable explanation.
func (c Certificate) Render() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Claim: %s\n", c.Binding.Sentence)
	fmt.Fprintf(&sb, "Metric: %s / %s over %s\n", c.Binding.MetricNumerator, c.Binding.MetricDenom, c.Binding.Population)
	if c.Binding.Budgets.Unrestricted {
		fmt.Fprintf(&sb, "Claimed range: unrestricted (unqualified label)\n")
	} else {
		fmt.Fprintf(&sb, "Claimed range: budgets [%d,%d]\n", c.Binding.Budgets.Lo, c.Binding.Budgets.Hi)
	}
	fmt.Fprintf(&sb, "Trace extension: %s\n", c.TraceExtension)
	for _, n := range c.ConditionNotes {
		fmt.Fprintf(&sb, "Condition note: %s\n", n)
	}
	for _, v := range c.ExtensionViolations {
		fmt.Fprintf(&sb, "Extension violation: %s\n", v)
	}
	sb.WriteString("\n")
	for _, r := range c.Rows {
		fmt.Fprintf(&sb, "%s\n", r.Label)
		fmt.Fprintf(&sb, "  Before: %s   solved %d\n", r.Before.Rate(), r.SolvedBefore)
		if r.MarginalYield != "" {
			fmt.Fprintf(&sb, "  Added:  %s (marginal yield)\n", r.MarginalYield)
		}
		fmt.Fprintf(&sb, "  After:  %s   solved %d   previously solved lost: %d\n", r.After.Rate(), r.SolvedAfter, r.SolvedLost)
		if r.Accounting != "" {
			fmt.Fprintf(&sb, "  Accounting: %s\n", r.Accounting)
		}
	}
	fmt.Fprintf(&sb, "\nQuestion: %s\n", c.Question)
	fmt.Fprintf(&sb, "Verdict: %s\n", c.Verdict)
	fmt.Fprintf(&sb, "Reason: %s\n\n", c.Reason)
	for _, n := range c.NotAssessed {
		fmt.Fprintf(&sb, "%s\n", n)
	}
	return sb.String()
}
