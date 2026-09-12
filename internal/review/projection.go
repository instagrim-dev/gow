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
	// StaleDependency marks an assessment whose manifest names a dependency
	// that changed RELEVANTLY since it was made. Staleness is supplied by the
	// caller that knows current dependency values; this package does not guess
	// at repository state, and an unrelated edit must never set it.
	StaleDependency bool
	StaleReason     string
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

	// Contradiction detection runs over ALL assessments before any selection,
	// so a disagreement can never be hidden by picking one.
	outcomes := map[string]bool{}
	for _, a := range o.Assessments {
		outcomes[a.Outcome] = true
	}
	p.Contradiction = len(outcomes) > 1
	if p.Contradiction {
		p.Notes = append(p.Notes, "assessments disagree; every outcome is retained rather than resolved by recency")
	}

	// A demonstrated nonconformance governs whenever one exists and is not
	// stale: the weaker/negative result is preserved rather than being
	// overwritten by a later favorable assessment.
	for _, a := range o.Assessments {
		if a.Outcome == Nonconforms && !a.StaleDependency {
			p.State = StateNonconforms
			p.GoverningAssessmentID = a.ID
			return p
		}
	}

	// Otherwise the newest COMPATIBLE assessment governs. A stale assessment is
	// not usable for a current decision, but it keeps its historical result.
	var governing *AssessmentRecord
	for i := len(o.Assessments) - 1; i >= 0; i-- {
		if !o.Assessments[i].StaleDependency {
			governing = &o.Assessments[i]
			break
		}
	}
	if governing == nil {
		p.State = StateInconclusive
		p.Reasons = append(p.Reasons, ReasonStaleDependency)
		for _, a := range o.Assessments {
			if a.StaleReason != "" {
				p.Notes = append(p.Notes, "stale: "+a.ID+": "+a.StaleReason)
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
