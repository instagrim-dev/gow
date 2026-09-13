package review

import (
	"sort"
	"strconv"
)

// ObligationRecords is the input view of one obligation's ledger records. The
// package takes plain values so the projection is testable without a store and
// cannot reach back into storage for a more convenient answer.
type ObligationRecords struct {
	// ID is the exact obligation record id, carried so the export can be traced
	// to the revision that was actually bound into the policy rather than to a
	// key that may have several revisions.
	ID               string
	Key              string
	SemanticRevision int
	Requirement      string
	PrimaryOwner     string
	Mandatory        bool

	// ApplicabilityDecisions are every recorded decision, oldest first. Zero or
	// conflicting values are unresolved.
	ApplicabilityDecisions []ApplicabilityRecord
	// Assessments are every assessment, oldest first, with contradictions
	// preserved.
	Assessments []AssessmentRecord
	// Checks are every check attempt, including blocked and inconclusive.
	Checks []CheckRecord
}

// ApplicabilityRecord is one applicability decision.
type ApplicabilityRecord struct {
	ID         string
	SubjectRef string
	Decision   string
	Rationale  string
	Authorizer string
	CreatedAt  string
}

// AssessmentRecord is one assessment plus the identity of what it depends on.
type AssessmentRecord struct {
	ID              string
	SubjectRef      string
	ContextRef      string
	Outcome         string
	Argument        string
	Assessor        string
	ManifestID      string
	CheckAttemptIDs []string
	CreatedAt       string
	// ReferenceScope is assigned by the persisted-record adapter. Its empty
	// value remains exact-scope-valid for pure unit inputs that predate this
	// read-side classification; persisted coverage always carries an explicit
	// classification.
	ReferenceScope       AssessmentReferenceScope
	ReferenceScopeReason string
	// Compatibility is the caller-supplied three-state judgment about whether
	// this assessment's declared dependencies match the current request's
	// context.
	//
	//   - CompatibilityCompatible: every declared dep matches (or no deps).
	//   - CompatibilityStale:      at least one declared dep's current ref
	//                              differs from the assessed value.
	//   - CompatibilityUnknown:    at least one declared dep is missing a
	//                              supplied current ref; current relevance
	//                              undetermined.
	//
	// The empty string is treated as `compatible` for backward compatibility
	// with callers that predate the three-state model. New callers should set
	// it explicitly.
	Compatibility Compatibility
	// CompatibilityReason is a short human-readable explanation naming the
	// declared dependency and (when known) the current ref that produced the
	// judgment. It is retained regardless of state so a reader can audit why
	// an assessment was treated as compatible / stale / unknown.
	CompatibilityReason string

	// ProjectRevision, ContractHash and RecipeHash are the manifest revisions
	// this assessment was bound to. They are carried onto the record so the
	// export can name the exact inputs a reader needs to judge compatibility,
	// instead of printing only a manifest id that forces a query back into the
	// store the export was meant to summarize.
	ProjectRevision string
	ContractHash    string
	RecipeHash      string

	// ManifestDependencies is the declared dependency manifest for this
	// assessment — every dep kind, ref, and why-relevant reason the assessment
	// itself cited. Carrying the manifest contents into the projection is what
	// lets the coverage export be an auditable evidence bundle rather than a
	// summary that forces the reader back into the source database.
	ManifestDependencies []ManifestDependency
	// EvidenceCutoff is the assessment's declared evidence-cutoff timestamp,
	// carried alongside the manifest so a reader can see the temporal boundary
	// the assessment was made against.
	EvidenceCutoff string
}

// ManifestDependency is one declared dependency of an assessment's manifest.
// The projection carries the manifest contents through so the coverage export
// can be independently audited; without them, a reader has to trust the manifest
// ID and query the source database.
type ManifestDependency struct {
	Ordinal        int
	DependencyKind string
	DependencyRef  string
	WhyRelevant    string
}

// CheckRecord is one check attempt.
type CheckRecord struct {
	ID           string
	CaseLabel    string
	ProcedureRef string
	Mode         string
	Outcome      string
	Executor     string
	Blocker      string
	OutputRef    string
	StartedAt    string
	EndedAt      string
	ResourceNote string

	// ProcedureRevision, InputsRef, and Environment carry checker provenance
	// through to the coverage export. Without them a reader with only the
	// export cannot reconstruct which version of a procedure produced the
	// outcome, what inputs it was given, or in which environment it ran —
	// which the recipe requires to be independently auditable evidence rather
	// than a rendered summary.
	ProcedureRevision string
	InputsRef         string
	Environment       string
}

// PolicyRecords is the input view of the decision policy.
type PolicyRecords struct {
	// ID is the exact policy record id. It is carried into the export because a
	// decision that cites only a policy KEY cannot be traced to the revision
	// that actually authorized it.
	ID                 string
	Key                string
	Revision           int
	DecisionName       string
	Owner              string
	AuthoritySource    string
	ScopeJustification string
	EvidenceCutoff     string
	Obligations        []ObligationRecords
}

// ObligationProjection is the derived view of one obligation.
type ObligationProjection struct {
	Obligation ObligationRecords
	State      ExaminationState
	// Reasons are the UNDETERMINED reason codes this obligation contributes.
	Reasons []string
	// GoverningAssessmentID is the assessment the state came from; empty when
	// unexamined or applicability-unresolved.
	GoverningAssessmentID string
	// Contradiction is set when assessments disagree in outcome. The
	// disagreement is retained rather than resolved by recency.
	Contradiction bool
	// Notes records derivation details worth reading in the export (for example
	// which conflicting values were seen).
	Notes []string
}

// Projection is the whole decision projection for one policy revision.
type Projection struct {
	Policy      PolicyRecords
	Obligations []ObligationProjection
	Decision    Decision
	// Reasons are the deduplicated, sorted UNDETERMINED reason codes. Empty for
	// WITHHOLD and ELIGIBLE_TO_ADVANCE.
	Reasons []string
	// Blockers lists the demonstrated blocking nonconformances behind a
	// WITHHOLD, so the decision is never a bare label.
	Blockers []string
	// Vacuous reports a policy with no mandatory obligation. A vacuous policy
	// can never yield eligibility: an empty mandatory set is not success.
	Vacuous bool
}

// Project derives the decision for a policy revision from its records.
//
// Precedence is deliberate and follows the contract:
//
//  1. WITHHOLD if any mandatory obligation has a demonstrated, unresolved
//     nonconformance. A blocker outranks every uncertainty; the uncertainties
//     are still listed.
//  2. Otherwise UNDETERMINED if any mandatory obligation's applicability,
//     authority, or evidentiary support is unresolved — including a vacuous or
//     unauthorized policy.
//  3. Otherwise ELIGIBLE_TO_ADVANCE.
func Project(policy PolicyRecords) Projection {
	out := Projection{Policy: policy}

	reasons := map[string]bool{}
	for _, o := range policy.Obligations {
		p := projectObligation(o)
		out.Obligations = append(out.Obligations, p)
		if !o.Mandatory {
			// A non-mandatory obligation's state is reported but does not drive
			// the decision. Its findings are not silently dropped: they remain
			// on the projection for the export.
			continue
		}
		if p.State == StateNonconforms {
			out.Blockers = append(out.Blockers, o.Key+"@"+strconv.Itoa(o.SemanticRevision)+": "+p.GoverningAssessmentID)
		}
		for _, r := range p.Reasons {
			reasons[r] = true
		}
	}

	// Authority: an unauthorized or unscoped policy cannot grant eligibility,
	// regardless of how favorable the assessments look.
	if policy.Owner == "" || policy.AuthoritySource == "" || policy.ScopeJustification == "" {
		reasons[ReasonPolicyMissingOrUnauthorized] = true
	}
	mandatory := 0
	for _, o := range policy.Obligations {
		if o.Mandatory {
			mandatory++
		}
	}
	if mandatory == 0 {
		// Vacuous success is refused: an empty mandatory set is an absent scope
		// justification in practice, not a satisfied one.
		out.Vacuous = true
		reasons[ReasonPolicyMissingOrUnauthorized] = true
	}

	switch {
	case len(out.Blockers) > 0:
		out.Decision = DecisionWithhold
		// Remaining uncertainties stay visible alongside the blocker.
		out.Reasons = sortedKeys(reasons)
	case len(reasons) > 0:
		out.Decision = DecisionUndetermined
		out.Reasons = sortedKeys(reasons)
	default:
		out.Decision = DecisionEligible
	}
	sort.Strings(out.Blockers)
	return out
}

// projectObligation derives one obligation's examination state and reasons.
func projectObligation(o ObligationRecords) ObligationProjection {
	p := ObligationProjection{Obligation: o}

	// Applicability first: an obligation whose applicability is unknown or
	// contested cannot be assessed into a decision.
	switch decision, conflict := applicability(o.ApplicabilityDecisions); {
	case conflict:
		p.State = StateApplicabilityUnresolved
		p.Reasons = append(p.Reasons, ReasonApplicabilityUnresolved)
		p.Notes = append(p.Notes, "conflicting applicability decisions recorded; conflict is unresolved, not a pass")
		return p
	case decision == "":
		p.State = StateApplicabilityUnresolved
		p.Reasons = append(p.Reasons, ReasonApplicabilityUnresolved)
		p.Notes = append(p.Notes, "no applicability decision recorded")
		return p
	case decision == DoesNotApply:
		// Recorded inapplicability is a real outcome. It is NOT the same as
		// unexamined, and missing implementation must never arrive here.
		p.State = StateNotApplicable
		return p
	}

	if len(o.Assessments) == 0 {
		// Absence of an assessment is unexamined. No empty assessment is
		// created to claim coverage.
		p.State = StateUnexamined
		p.Reasons = append(p.Reasons, ReasonUnexamined)
		return p
	}

	validAssessments := make([]AssessmentRecord, 0, len(o.Assessments))
	for _, a := range o.Assessments {
		if referenceScopeOf(a) == AssessmentReferenceScopeInvalid {
			p.Notes = append(p.Notes, "invalid assessment reference: "+a.ID+": "+a.ReferenceScopeReason)
			continue
		}
		validAssessments = append(validAssessments, a)
	}
	if len(validAssessments) == 0 {
		// The subject was examined, but every historical assessment cited the
		// wrong applicability decision or manifest. Keep the rows visible; do
		// not let malformed history become either an unexamined gap or current
		// authority.
		p.State = StateInconclusive
		p.Reasons = append(p.Reasons, ReasonInvalidAssessmentReference)
		return p
	}

	// Contradiction detection runs only over exact-scope-valid assessments
	// before selection. A malformed historical row is retained for export but
	// cannot create a current contradiction or hide a valid governing record.
	outcomes := map[string]bool{}
	for _, a := range validAssessments {
		outcomes[a.Outcome] = true
	}
	p.Contradiction = len(outcomes) > 1
	if p.Contradiction {
		p.Notes = append(p.Notes, "assessments disagree; every outcome is retained rather than resolved by recency")
	}

	// A demonstrated nonconformance governs whenever one exists and is not
	// stale-or-unknown: the weaker/negative result is preserved rather than
	// being overwritten by a later favorable assessment. A nonconformance
	// under an unknown-compatibility context is preserved as history but does
	// not become a current blocker, matching the symmetric treatment of a
	// stale nonconformance — an assessment whose current relevance is not
	// established cannot supply a current-decision outcome in either
	// direction.
	for _, a := range validAssessments {
		if a.Outcome == Nonconforms && compatibilityOf(a) == CompatibilityCompatible {
			p.State = StateNonconforms
			p.GoverningAssessmentID = a.ID
			return p
		}
	}

	// Otherwise the newest COMPATIBLE assessment governs. A stale or
	// unknown-compatibility assessment is not usable for a current decision,
	// but it keeps its historical result. The two non-compatible states are
	// distinguished in the reasons so a reader can tell "we know the dep
	// moved" from "the caller didn't say what the current dep is".
	var governing *AssessmentRecord
	hadStale := false
	hadUnknown := false
	for i := len(validAssessments) - 1; i >= 0; i-- {
		switch compatibilityOf(validAssessments[i]) {
		case CompatibilityCompatible:
			governing = &validAssessments[i]
		case CompatibilityStale:
			hadStale = true
		case CompatibilityUnknown:
			hadUnknown = true
		}
		if governing != nil {
			break
		}
	}
	if governing == nil {
		p.State = StateInconclusive
		if hadStale {
			p.Reasons = append(p.Reasons, ReasonStaleDependency)
		}
		if hadUnknown {
			p.Reasons = append(p.Reasons, ReasonCompatibilityUnknown)
		}
		if !hadStale && !hadUnknown {
			// Every assessment must have fallen into a state we don't
			// enumerate — surface it as inconclusive rather than pretending
			// the obligation was assessed and passed.
			p.Reasons = append(p.Reasons, ReasonInconclusive)
		}
		for _, a := range validAssessments {
			if a.CompatibilityReason != "" {
				p.Notes = append(p.Notes, string(compatibilityOf(a))+": "+a.ID+": "+a.CompatibilityReason)
			}
		}
		return p
	}
	p.GoverningAssessmentID = governing.ID

	switch governing.Outcome {
	case Conforms:
		// Conformance requires supporting checks that actually ran. A blocked
		// attempt cannot support conformance (the store trigger refuses the
		// link too); an inspection-only basis is reported as inconclusive
		// rather than certified, because no inspection-only review may certify
		// an unexecuted check.
		support, blocked := checkSupport(o.Checks, governing.CheckAttemptIDs)
		switch {
		case blocked:
			p.State = StateBlocked
			p.Reasons = append(p.Reasons, ReasonExecutionBlocked)
			p.Notes = append(p.Notes, "a cited check attempt was blocked; conformance is not established")
		case !support:
			p.State = StateInconclusive
			p.Reasons = append(p.Reasons, ReasonInconclusive)
			p.Notes = append(p.Notes, "no completed EXECUTED check supports this conformance assessment")
		default:
			p.State = StateConforms
		}
	case Inconclusive:
		p.State = StateInconclusive
		p.Reasons = append(p.Reasons, ReasonInconclusive)
		if anyBlocked(o.Checks, governing.CheckAttemptIDs) {
			p.Reasons = append(p.Reasons, ReasonExecutionBlocked)
		}
	case Nonconforms:
		p.State = StateNonconforms
	default:
		p.State = StateInconclusive
		p.Reasons = append(p.Reasons, ReasonInconclusive)
	}
	return p
}

func referenceScopeOf(a AssessmentRecord) AssessmentReferenceScope {
	if a.ReferenceScope == "" {
		return AssessmentReferenceScopeExactValid
	}
	return a.ReferenceScope
}

// compatibilityOf reads the caller-supplied compatibility state, defaulting to
// `compatible` when unset. The empty string is preserved for backward
// compatibility with callers that do not know about the three-state model:
// they see the projection they would have seen under the old boolean when the
// bool was false.
func compatibilityOf(a AssessmentRecord) Compatibility {
	if a.Compatibility == "" {
		return CompatibilityCompatible
	}
	return a.Compatibility
}

// applicability collapses the recorded decisions into one value, reporting
// conflict rather than choosing.
func applicability(decisions []ApplicabilityRecord) (string, bool) {
	seen := map[string]bool{}
	for _, d := range decisions {
		seen[d.Decision] = true
	}
	switch len(seen) {
	case 0:
		return "", false
	case 1:
		for k := range seen {
			return k, false
		}
	}
	return "", true
}

// checkSupport reports whether the cited attempts include at least one
// completed EXECUTED check, and whether any cited attempt was blocked.
func checkSupport(checks []CheckRecord, citedIDs []string) (support bool, blocked bool) {
	cited := map[string]bool{}
	for _, id := range citedIDs {
		cited[id] = true
	}
	for _, c := range checks {
		if !cited[c.ID] {
			continue
		}
		if c.Outcome == CheckBlocked {
			blocked = true
		}
		if c.Outcome == CheckCompleted && c.Mode == ModeExecuted {
			support = true
		}
	}
	return support, blocked
}

func anyBlocked(checks []CheckRecord, citedIDs []string) bool {
	_, blocked := checkSupport(checks, citedIDs)
	return blocked
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
