// Package review is the normative review-record layer: the decision projection
// and the deterministic coverage generator for the review contract's four
// record responsibilities (G1 of the 2026-09-12 review-flow run).
//
// It is pure. It holds no SQL and no CLI concepts, and it computes nothing that
// is not derivable from the records handed to it. The point of the package is
// that coverage is an OUTPUT: given applicability decisions, assessments, check
// attempts and dependency manifests, exactly one decision and one document
// follow. There is no status to edit.
//
// What this package deliberately does NOT do:
//
//   - It does not invent assessments. An obligation with no assessment is
//     `unexamined`; absence is a reported state, not a gap to fill.
//   - It does not resolve contradictions by recency. Conflicting evidence is
//     surfaced, because discarding the inconvenient record by timestamp is how a
//     review launders a blocker into a pass.
//   - It does not aggregate legacy per-report verdicts (READY/BLOCKED/etc.) into
//     eligibility. Those are not a lattice.
//   - It does not treat scientific claims as obligations. A CandidateInvariant
//     claims a regularity over a population; an obligation requires a property.
package review

// Decision is the projection of a policy's mandatory set onto one outcome.
type Decision string

const (
	// DecisionWithhold: at least one demonstrated, unresolved blocking
	// nonconformance applies.
	DecisionWithhold Decision = "WITHHOLD"
	// DecisionUndetermined: no demonstrated blocker, but mandatory
	// applicability, authority or evidentiary support is unresolved.
	DecisionUndetermined Decision = "UNDETERMINED"
	// DecisionEligible: every mandatory obligation has current sufficient
	// conformance support under an authorized, non-vacuous policy. Scoped
	// permission to advance — never a claim of scientific success.
	DecisionEligible Decision = "ELIGIBLE_TO_ADVANCE"
)

// Reason codes for DecisionUndetermined. Multiple may apply. `Unexamined` and
// `Inconclusive` are kept distinct on purpose: never examined and examined
// without a conclusion are different epistemic states.
const (
	ReasonPolicyMissingOrUnauthorized = "policy_missing_or_unauthorized"
	ReasonApplicabilityUnresolved     = "applicability_unresolved"
	ReasonUnexamined                  = "unexamined"
	ReasonInconclusive                = "inconclusive"
	ReasonExecutionBlocked            = "execution_blocked"
	ReasonStaleDependency             = "stale_dependency"
)

// Applicability outcomes.
const (
	Applies      = "applies"
	DoesNotApply = "does_not_apply"
)

// Assessment outcomes.
const (
	Conforms     = "conforms"
	Nonconforms  = "nonconforms"
	Inconclusive = "inconclusive"
)

// Check attempt outcomes and modes.
const (
	CheckCompleted    = "completed"
	CheckInconclusive = "inconclusive"
	CheckBlocked      = "blocked"

	ModeExecuted  = "executed"
	ModeInspected = "inspected"
)

// ExaminationState is the derived examination status of one obligation. It is
// computed, never stored, so it cannot drift from the records.
type ExaminationState string

const (
	// StateUnexamined: no assessment exists. Absence of an assessment is
	// unexamined — not a pass, and not something an empty assessment may fake.
	StateUnexamined ExaminationState = "unexamined"
	// StateApplicabilityUnresolved: no applicability decision, or conflicting
	// decisions.
	StateApplicabilityUnresolved ExaminationState = "applicability_unresolved"
	// StateNotApplicable: an authorized decision says the obligation does not
	// apply. Distinct from unexamined and from missing implementation.
	StateNotApplicable ExaminationState = "not_applicable"
	// StateConforms: assessed as conforming, with supporting checks.
	StateConforms ExaminationState = "conforms"
	// StateNonconforms: assessed as nonconforming — a demonstrated blocker when
	// the obligation is mandatory.
	StateNonconforms ExaminationState = "nonconforms"
	// StateInconclusive: examined, no conclusion reached.
	StateInconclusive ExaminationState = "inconclusive"
	// StateBlocked: examination could not complete (an execution blocker).
	StateBlocked ExaminationState = "blocked"
)
