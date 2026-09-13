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
	// ReasonInvalidAssessmentReference means historical assessments exist, but
	// none cites applicability and a manifest for its exact scope. Their rows
	// remain exported as immutable history while current authority is withheld.
	ReasonInvalidAssessmentReference = "invalid_assessment_reference"
	// ReasonCompatibilityUnknown: a declared current dependency was not
	// supplied when the coverage was requested. The historical assessment is
	// retained; current permission is undetermined because compatibility with
	// the current context has not been established. Preserving historical
	// validity does not grant current permission to advance.
	ReasonCompatibilityUnknown = "compatibility_unknown"
)

// AssessmentReferenceScope is the replay-time classification of an
// assessment's cited applicability decision and dependency manifest. Fresh
// writes are blocked structurally by v47; historical rows can still be invalid
// and must not be selected as authority or contradiction evidence.
type AssessmentReferenceScope string

const (
	AssessmentReferenceScopeExactValid AssessmentReferenceScope = "exact_scope_valid"
	AssessmentReferenceScopeInvalid    AssessmentReferenceScope = "invalid"
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

// Well-known dependency kinds.
//
// A kind is only meaningful because an assessment DECLARED it with a stated
// reason; these constants exist so the CLI, the repository gates and the
// projection agree on spelling rather than each inventing a near-synonym that
// silently never matches the current value it was supposed to track.
//
// Declaring a kind is opt-in on purpose. `DepKindProjectRevision` in particular
// must not be implied for every assessment: if any code change invalidated every
// assessment, the relevant-versus-unrelated distinction would collapse into
// "HEAD moved" and staleness would stop carrying information.
const (
	// DepKindProjectRevision is the checkout revision whose behavior was
	// assessed. Declare it when the conclusion depends on code that could change.
	DepKindProjectRevision = "project_revision"
	// DepKindAssessmentPopulation is the evidence population the assessment
	// examined.
	DepKindAssessmentPopulation = "assessment_population"
	// DepKindPolicyRevision is the decision policy revision whose acceptance
	// criteria applied.
	DepKindPolicyRevision = "policy_revision"
	// DepKindCandidateContent is the exact subject content; different content is
	// a different subject, not a stale assessment.
	DepKindCandidateContent = "candidate_content"
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

// Compatibility is the three-state judgment about whether one assessment's
// declared dependencies match the current context supplied for a coverage
// request. It replaces the earlier boolean staleness because "compatibility
// unknown" is not the same as "compatible": an incompletely specified current
// context must not be permitted to grant current eligibility on the strength
// of an assessment whose current relevance was never established.
//
//   - `compatible`   — every declared dependency the assessment cited has a
//     current ref that matches, or the assessment declared no
//     dependencies at all.
//   - `stale`        — a declared dependency's current ref differs from the
//     value the assessment was made against. The historical
//     assessment stands as history; it does not carry current
//     authority.
//   - `unknown`      — one or more declared dependencies have no supplied
//     current ref. The historical assessment stands as
//     history; current authority is undetermined until the
//     caller supplies those current values.
//
// Compatibility is a CALLER-SUPPLIED judgment against the current request. The
// projection reads it; it does not compute compatibility itself, because
// deciding whether two refs describe the same thing requires knowledge of the
// current repository/database state that a record reader does not have.
type Compatibility string

const (
	// CompatibilityCompatible is the default when the caller did not supply a
	// compatibility function: with no current context declared, an assessment
	// is exported as-is for historical purposes and treated as compatible for
	// the decision the caller has said they want to make against no current
	// context. Callers seeking current-decision guarantees must supply a
	// compatibility function that reports unknown for missing values.
	CompatibilityCompatible Compatibility = "compatible"
	// CompatibilityStale is a definite mismatch: a declared dependency's
	// current ref differs from the value the assessment was made against.
	CompatibilityStale Compatibility = "stale"
	// CompatibilityUnknown is the missing state: the caller has not supplied
	// a current ref for a declared dependency, so we cannot say whether the
	// historical assessment applies to the current context.
	CompatibilityUnknown Compatibility = "unknown"
)
