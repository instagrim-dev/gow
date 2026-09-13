package finite

import (
	"fmt"
	"strings"
)

// Verdict states for the single question a certificate answers. The
// vocabulary deliberately has no state readable as "equivalent in
// general": the strongest verdict names the domain it exhausted.
const (
	// VerdictHoldsOnDomain: every assignment of the declared finite domain
	// was enumerated and the two expressions agreed on all of them. This
	// supports exactly the declared domain.
	VerdictHoldsOnDomain = "HOLDS_ON_DECLARED_DOMAIN"
	// VerdictRefuted: an exact counterexample assignment is recorded.
	VerdictRefuted = "REFUTED"
	// VerdictInstanceOnly: the supplied assignments agreed, and that is
	// all this outcome states. It is not an equivalence over any domain
	// and must not be consumed as one (T0: instance success remains
	// instance evidence).
	VerdictInstanceOnly = "INSTANCE_EVIDENCE_ONLY"
	// VerdictInapplicable: a premise failed (undeclared variable, unknown
	// operator, invalid width); the identified failure is recorded.
	VerdictInapplicable = "INAPPLICABLE"
	// VerdictUnresolved: the declared domain exceeds the exhaustiveness
	// cap; this tool refuses to substitute sampling for enumeration.
	VerdictUnresolved = "UNRESOLVED"
)

// Binding records the exact equivalence claim under assessment, so the
// certificate cannot drift from the sentence it decides.
type Binding struct {
	Sentence string // exact sentence or rule statement from the source record
	Domain   Domain // the declared finite semantics the claim quantifies over
}

// Counterexample is one exact refuting assignment with both evaluations.
type Counterexample struct {
	Assignment string // canonical rendering, e.g. "x=8, y=0"
	Left       uint64
	Right      uint64
}

// Certificate is the machine-readable outcome plus a fixed-template
// explanation. It answers exactly the bound claim; the NotAssessed lines
// are scope guards preventing a certificate answering one question from
// being consumed as an answer to another.
type Certificate struct {
	Binding Binding
	Left    string // rendered expressions, fixed at assessment time
	Right   string

	// Applicability
	PremiseFailures []string

	// Coverage accounting.
	DomainSize         int64
	AssignmentsChecked int64
	Exhaustive         bool

	// The one answered question.
	Question string
	Verdict  string
	Reason   string

	Counterexample *Counterexample

	// Scope guards.
	NotAssessed []string
}

func scopeGuards(d Domain) []string {
	return []string{
		fmt.Sprintf("Equivalence outside the declared domain (%s): NOT ASSESSED (an exhaustive finite certificate supports exactly its domain)", d),
		"Equivalence at other word widths: NOT ASSESSED (a width-4 identity need not hold at width 8, and conversely)",
		"Real or unbounded-integer identity: NOT ASSESSED (machine-word semantics do not inherit real-number identities)",
		"Cost, performance, or preferability: NOT ASSESSED (semantic equality is not improvement)",
		"General rewrite-rule admission: NOT ASSESSED (admission is a review decision citing this certificate, not a property of it)",
	}
}

// AssessEquivalence decides whether left and right agree on every
// assignment of the declared finite domain, by enumeration. The first
// counterexample in canonical order is recorded exactly. Structure is a
// checked premise: a nil or malformed expression is an applicability
// refusal, never a panic.
func AssessEquivalence(b Binding, left, right Expr) Certificate {
	return AssessEquivalenceCancelled(b, left, right, nil)
}

// AssessEquivalenceCancelled is independent exhaustive replay with a
// cancellation check inside each expression evaluation. Cancellation
// retains completed-assignment counts and never certifies exhaustiveness.
func AssessEquivalenceCancelled(b Binding, left, right Expr, cancel <-chan struct{}) Certificate {
	cert := Certificate{
		Binding:  b,
		Question: "Do the two expressions evaluate identically on every assignment of the declared finite domain?",
	}
	if defects := ValidateDomain(b.Domain); len(defects) > 0 {
		cert.PremiseFailures = defects
		cert.Verdict = VerdictInapplicable
		cert.Reason = "a domain premise failed before rendering: " + strings.Join(defects, "; ")
		return cert
	}
	if isCancelled(cancel) {
		cert.Verdict = VerdictUnresolved
		cert.Reason = "independent replay cancelled before evaluation; equivalence remains unverified"
		return cert
	}
	cert.NotAssessed = scopeGuards(b.Domain)
	if defects := append(structureDefects("left", left), structureDefects("right", right)...); len(defects) > 0 {
		cert.PremiseFailures = defects
		cert.Verdict = VerdictInapplicable
		cert.Reason = "a premise failed before evaluation: " + strings.Join(defects, "; ")
		return cert
	}
	cert.Left = left.render()
	cert.Right = right.render()
	if b.Domain.Width < MinWidth || b.Domain.Width > MaxWidth {
		cert.Verdict = VerdictInapplicable
		cert.Reason = fmt.Sprintf("declared width %d is outside the supported range [%d,%d]", b.Domain.Width, MinWidth, MaxWidth)
		return cert
	}
	cert.PremiseFailures = append(validate(left, b.Domain), validate(right, b.Domain)...)
	if len(cert.PremiseFailures) > 0 {
		cert.Verdict = VerdictInapplicable
		cert.Reason = "a premise failed before evaluation: " + strings.Join(cert.PremiseFailures, "; ")
		return cert
	}
	cert.DomainSize = b.Domain.Size()
	if cert.DomainSize > ExhaustiveCap {
		cert.Verdict = VerdictUnresolved
		cert.Reason = fmt.Sprintf("declared domain has at least %d assignments, above the exhaustiveness cap %d; this tool does not substitute sampling for enumeration, so the claim stays undecided here", cert.DomainSize, int64(ExhaustiveCap))
		return cert
	}
	enumerate(b.Domain, func(a Assignment) bool {
		l, ok := evalCancelled(left, b.Domain, a, cancel)
		if !ok {
			return false
		}
		r, ok := evalCancelled(right, b.Domain, a, cancel)
		if !ok {
			return false
		}
		cert.AssignmentsChecked++
		if l != r {
			cert.Counterexample = &Counterexample{Assignment: a.render(), Left: l, Right: r}
			return false
		}
		return true
	})
	if cert.Counterexample != nil {
		cert.Verdict = VerdictRefuted
		cert.Reason = fmt.Sprintf("exact counterexample at {%s}: left evaluates to %d, right to %d", cert.Counterexample.Assignment, cert.Counterexample.Left, cert.Counterexample.Right)
		return cert
	}
	if isCancelled(cancel) {
		cert.Verdict = VerdictUnresolved
		cert.Reason = fmt.Sprintf("independent replay cancelled after %d complete assignments; equivalence remains unverified", cert.AssignmentsChecked)
		return cert
	}
	cert.Exhaustive = true
	cert.Verdict = VerdictHoldsOnDomain
	cert.Reason = fmt.Sprintf("all %d assignments of the declared domain were enumerated and agree; this supports the declared domain only", cert.AssignmentsChecked)
	return cert
}

// AssessInstances checks agreement on the supplied assignments only. The
// strongest available outcome is INSTANCE_EVIDENCE_ONLY: by construction
// this function cannot produce a domain-level equivalence, which is the T0
// instance-to-universal rejection enforced as a type of outcome. The
// asymmetry is preserved exactly: agreeing instances decide nothing about
// the domain, but one admissible counterexample DOES refute the universal
// claim over that domain.
func AssessInstances(b Binding, left, right Expr, instances []Assignment) Certificate {
	if defects := ValidateDomain(b.Domain); len(defects) > 0 {
		return Certificate{Binding: b, PremiseFailures: defects, Verdict: VerdictInapplicable,
			Reason: "a domain premise failed before rendering: " + strings.Join(defects, "; ")}
	}
	cert := Certificate{
		Binding:     b,
		Question:    "Do the two expressions evaluate identically on the supplied assignments (and only those)?",
		NotAssessed: scopeGuards(b.Domain),
	}
	if defects := append(structureDefects("left", left), structureDefects("right", right)...); len(defects) > 0 {
		cert.PremiseFailures = defects
		cert.Verdict = VerdictInapplicable
		cert.Reason = "a premise failed before evaluation: " + strings.Join(defects, "; ")
		return cert
	}
	cert.Left = left.render()
	cert.Right = right.render()
	if b.Domain.Width < MinWidth || b.Domain.Width > MaxWidth {
		cert.Verdict = VerdictInapplicable
		cert.Reason = fmt.Sprintf("declared width %d is outside the supported range [%d,%d]", b.Domain.Width, MinWidth, MaxWidth)
		return cert
	}
	cert.DomainSize = b.Domain.Size()
	cert.PremiseFailures = append(validate(left, b.Domain), validate(right, b.Domain)...)
	if len(cert.PremiseFailures) > 0 {
		cert.Verdict = VerdictInapplicable
		cert.Reason = "a premise failed before evaluation: " + strings.Join(cert.PremiseFailures, "; ")
		return cert
	}
	if len(instances) == 0 {
		cert.Verdict = VerdictUnresolved
		cert.Reason = "no instances supplied; nothing was checked"
		return cert
	}
	mask := b.Domain.mask()
	for _, a := range instances {
		for _, v := range b.Domain.Vars {
			val, ok := a[v]
			if !ok {
				cert.Verdict = VerdictInapplicable
				cert.Reason = fmt.Sprintf("instance {%s} does not assign declared variable %q", a.render(), v)
				return cert
			}
			if val > mask {
				cert.Verdict = VerdictInapplicable
				cert.Reason = fmt.Sprintf("instance {%s} assigns %q a value outside the %d-bit domain", a.render(), v, b.Domain.Width)
				return cert
			}
		}
		cert.AssignmentsChecked++
		l, r := left.eval(b.Domain, a), right.eval(b.Domain, a)
		if l != r {
			cert.Counterexample = &Counterexample{Assignment: a.render(), Left: l, Right: r}
			cert.Verdict = VerdictRefuted
			// A counterexample decides the universal claim over the
			// domain negatively; no "domain NOT ASSESSED" guard is
			// attached here, because the domain claim WAS assessed —
			// and refuted.
			cert.Reason = fmt.Sprintf("exact counterexample at {%s}: left evaluates to %d, right to %d (a single admissible counterexample refutes the universal claim over the declared domain)", a.render(), l, r)
			return cert
		}
	}
	cert.Verdict = VerdictInstanceOnly
	cert.NotAssessed = append(cert.NotAssessed, "Equivalence over the declared domain: NOT ASSESSED (agreeing instances are instance evidence, never a domain certificate)")
	cert.Reason = fmt.Sprintf("the %d supplied assignments agree; this is instance evidence and decides nothing about the %d-assignment domain", cert.AssignmentsChecked, cert.DomainSize)
	return cert
}

// Render produces the fixed-template human-readable explanation.
func (c Certificate) Render() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Claim: %s\n", c.Binding.Sentence)
	fmt.Fprintf(&sb, "Domain: %s\n", c.Binding.Domain)
	fmt.Fprintf(&sb, "Left:  %s\n", c.Left)
	fmt.Fprintf(&sb, "Right: %s\n", c.Right)
	for _, p := range c.PremiseFailures {
		fmt.Fprintf(&sb, "Premise failure: %s\n", p)
	}
	fmt.Fprintf(&sb, "Assignments checked: %d of %d (exhaustive: %v)\n", c.AssignmentsChecked, c.DomainSize, c.Exhaustive)
	if c.Counterexample != nil {
		fmt.Fprintf(&sb, "Counterexample: {%s} -> left %d, right %d\n", c.Counterexample.Assignment, c.Counterexample.Left, c.Counterexample.Right)
	}
	fmt.Fprintf(&sb, "\nQuestion: %s\n", c.Question)
	fmt.Fprintf(&sb, "Verdict: %s\n", c.Verdict)
	fmt.Fprintf(&sb, "Reason: %s\n\n", c.Reason)
	for _, n := range c.NotAssessed {
		fmt.Fprintf(&sb, "%s\n", n)
	}
	return sb.String()
}
