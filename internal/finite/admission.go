package finite

import (
	"fmt"
	"sort"
)

// VerifyRuleWarrant decides whether a presented certificate is sufficient
// warrant for the stated equality rule: "left == right for every
// assignment of domain d". It is the certificate-consumption boundary for
// the T0 B→C transition — the place where a distinct verdict's MEANING is
// enforced, so that no consumer can treat instance agreement, a foreign
// binding, or an incomplete enumeration as exhaustive evidence.
//
// It returns the complete list of warrant defects (empty means the
// certificate warrants exactly this rule). It grants nothing: admission
// remains a review decision that may cite a defect-free warrant check.
//
// Two layers, both required:
//
//  1. Structural checks name each way a certificate can fail to mean
//     "exhaustively verified equivalence of THIS rule on THIS domain":
//     wrong verdict type, non-exhaustive coverage, count inconsistency,
//     domain mismatch, expression mismatch.
//  2. Independent replay re-enumerates the domain and re-decides the
//     claim. A certificate is a record of a past execution; replay is
//     what makes a forged or stale record insufficient. (The roadmap
//     makes replay a G2 exit requirement for engine explanations; for
//     this finite language it is affordable at admission time, so it is
//     simply always done.)
func VerifyRuleWarrant(cert Certificate, left, right Expr, d Domain) []string {
	var defects []string

	// Structure first: the presented rule's expressions are external
	// input and must be traversable before anything is rendered.
	if sd := append(structureDefects("rule left", left), structureDefects("rule right", right)...); len(sd) > 0 {
		return append(defects, sd...)
	}

	// Layer 1: structural meaning of the presented certificate.
	if cert.Verdict != VerdictHoldsOnDomain {
		defects = append(defects, fmt.Sprintf("verdict %q does not warrant a domain equality; only %q does (instance agreement, refutation, refusal, and unresolved outcomes admit nothing)", cert.Verdict, VerdictHoldsOnDomain))
	}
	if !cert.Exhaustive {
		defects = append(defects, "certificate does not claim exhaustive enumeration; a partial check warrants no domain equality")
	}
	if cert.DomainSize <= 0 || cert.AssignmentsChecked != cert.DomainSize {
		defects = append(defects, fmt.Sprintf("coverage accounting is inconsistent: %d assignments checked of a domain of %d; an exhaustive claim must account for every assignment", cert.AssignmentsChecked, cert.DomainSize))
	}
	if cert.Counterexample != nil {
		defects = append(defects, "certificate carries a counterexample; it cannot simultaneously warrant the equality")
	}

	// Binding drift: the certificate must be about THIS rule on THIS
	// domain. Width, variable set, and both expression renderings must
	// match exactly — a certificate for a different width, a subset
	// domain, or a syntactically different expression warrants nothing
	// here.
	if cert.Binding.Domain.Width != d.Width {
		defects = append(defects, fmt.Sprintf("domain width mismatch: certificate is for %d-bit words, rule admission requests %d-bit", cert.Binding.Domain.Width, d.Width))
	}
	if !sameVars(cert.Binding.Domain.Vars, d.Vars) {
		// Structural comparison, not formatted-string comparison: a
		// pathological variable name could make two different domains
		// render identically (adversarial review finding 6).
		defects = append(defects, fmt.Sprintf("domain mismatch: certificate covers (%s), rule admission requests (%s)", cert.Binding.Domain, d))
	}
	if cert.Left != left.render() {
		defects = append(defects, fmt.Sprintf("left expression mismatch: certificate decided %q, rule admission presents %q", cert.Left, left.render()))
	}
	if cert.Right != right.render() {
		defects = append(defects, fmt.Sprintf("right expression mismatch: certificate decided %q, rule admission presents %q", cert.Right, right.render()))
	}
	if len(defects) > 0 {
		return defects
	}

	// Layer 2: independent replay. Re-decide the claim from the
	// presented rule and domain; the certificate's record must agree
	// with a fresh enumeration.
	replay := AssessEquivalence(Binding{Sentence: cert.Binding.Sentence, Domain: d}, left, right)
	if replay.Verdict != VerdictHoldsOnDomain {
		defects = append(defects, fmt.Sprintf("replay does not reproduce the equality: fresh enumeration returned %s (%s)", replay.Verdict, replay.Reason))
	}
	if replay.AssignmentsChecked != cert.AssignmentsChecked {
		defects = append(defects, fmt.Sprintf("replay coverage differs: %d assignments on replay vs %d recorded", replay.AssignmentsChecked, cert.AssignmentsChecked))
	}
	return defects
}

// sameVars compares variable sets structurally, order-insensitively.
func sameVars(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	as := append([]string(nil), a...)
	bs := append([]string(nil), b...)
	sort.Strings(as)
	sort.Strings(bs)
	for i := range as {
		if as[i] != bs[i] {
			return false
		}
	}
	return true
}
