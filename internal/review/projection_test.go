package review

import (
	"strings"
	"testing"
)

// These tests pin the projection's PRECEDENCE and its refusals. They exist
// because every rule here is a place where a review could quietly become more
// favorable than its records: a blocker treated as support, an absence treated
// as a pass, a contradiction resolved by recency, or an empty policy granting
// eligibility.

// baseObligation is a conforming mandatory obligation with executed support.
func baseObligation() ObligationRecords {
	return ObligationRecords{
		ID: "robl_1", Key: "o", SemanticRevision: 1, Requirement: "r",
		PrimaryOwner: "owner", Mandatory: true,
		ApplicabilityDecisions: []ApplicabilityRecord{{
			ID: "rapp_1", SubjectRef: "s", Decision: Applies, Rationale: "why", Authorizer: "who", CreatedAt: "t1",
		}},
		Assessments: []AssessmentRecord{{
			ID: "rasm_1", SubjectRef: "s", ContextRef: "c", Outcome: Conforms,
			Argument: "a", Assessor: "who", ManifestID: "rdep_1",
			CheckAttemptIDs: []string{"rchk_1"}, CreatedAt: "t2",
		}},
		Checks: []CheckRecord{{
			ID: "rchk_1", CaseLabel: "C1", ProcedureRef: "p",
			Mode: ModeExecuted, Outcome: CheckCompleted, Executor: "who",
		}},
	}
}

func basePolicy(obligations ...ObligationRecords) PolicyRecords {
	return PolicyRecords{
		ID: "rpol_1", Key: "p", Revision: 1, DecisionName: "d",
		Owner: "owner", AuthoritySource: "contract", ScopeJustification: "scoped",
		Obligations: obligations,
	}
}

func TestProjectEligibleRequiresExecutedSupport(t *testing.T) {
	t.Parallel()
	got := Project(basePolicy(baseObligation()))
	if got.Decision != DecisionEligible {
		t.Fatalf("decision = %s (reasons %v), want %s", got.Decision, got.Reasons, DecisionEligible)
	}
	if got.Obligations[0].State != StateConforms {
		t.Fatalf("state = %s, want %s", got.Obligations[0].State, StateConforms)
	}
}

func TestProjectInspectionCannotSupportConformance(t *testing.T) {
	t.Parallel()
	o := baseObligation()
	// The cited attempt was READ, not run. Reading a verifier's source is not
	// evidence that it passed.
	o.Checks[0].Mode = ModeInspected
	o.Checks[0].Outcome = CheckInconclusive
	got := Project(basePolicy(o))
	if got.Obligations[0].State != StateInconclusive {
		t.Fatalf("state = %s, want %s: an inspection-only basis cannot certify", got.Obligations[0].State, StateInconclusive)
	}
	if got.Decision != DecisionUndetermined {
		t.Fatalf("decision = %s, want %s", got.Decision, DecisionUndetermined)
	}
}

func TestProjectBlockedCheckIsNotSupport(t *testing.T) {
	t.Parallel()
	o := baseObligation()
	o.Checks[0].Outcome = CheckBlocked
	o.Checks[0].Blocker = "verifier unavailable"
	got := Project(basePolicy(o))
	if got.Obligations[0].State != StateBlocked {
		t.Fatalf("state = %s, want %s: a blocker must not be laundered into conformance", got.Obligations[0].State, StateBlocked)
	}
	if !contains(got.Reasons, ReasonExecutionBlocked) {
		t.Fatalf("reasons = %v, want %s", got.Reasons, ReasonExecutionBlocked)
	}
}

func TestProjectAbsenceIsUnexaminedNotPass(t *testing.T) {
	t.Parallel()
	o := baseObligation()
	o.Assessments = nil
	got := Project(basePolicy(o))
	if got.Obligations[0].State != StateUnexamined {
		t.Fatalf("state = %s, want %s", got.Obligations[0].State, StateUnexamined)
	}
	if got.Decision != DecisionUndetermined || !contains(got.Reasons, ReasonUnexamined) {
		t.Fatalf("decision = %s reasons = %v, want %s with %s", got.Decision, got.Reasons, DecisionUndetermined, ReasonUnexamined)
	}
}

func TestProjectUnexaminedDiffersFromInconclusive(t *testing.T) {
	t.Parallel()
	unexamined := baseObligation()
	unexamined.Assessments = nil
	inconclusive := baseObligation()
	inconclusive.Assessments[0].Outcome = Inconclusive

	a := Project(basePolicy(unexamined)).Obligations[0].State
	b := Project(basePolicy(inconclusive)).Obligations[0].State
	if a == b {
		t.Fatalf("never examined and examined-without-conclusion must stay distinct, both = %s", a)
	}
}

func TestProjectMissingApplicabilityIsUnresolved(t *testing.T) {
	t.Parallel()
	o := baseObligation()
	o.ApplicabilityDecisions = nil
	got := Project(basePolicy(o))
	if got.Obligations[0].State != StateApplicabilityUnresolved {
		t.Fatalf("state = %s, want %s", got.Obligations[0].State, StateApplicabilityUnresolved)
	}
	if !contains(got.Reasons, ReasonApplicabilityUnresolved) {
		t.Fatalf("reasons = %v, want %s", got.Reasons, ReasonApplicabilityUnresolved)
	}
}

func TestProjectConflictingApplicabilityIsUnresolvedNotChosen(t *testing.T) {
	t.Parallel()
	o := baseObligation()
	o.ApplicabilityDecisions = append(o.ApplicabilityDecisions, ApplicabilityRecord{
		ID: "rapp_2", SubjectRef: "s", Decision: DoesNotApply,
		Rationale: "later claim of inapplicability", Authorizer: "who", CreatedAt: "t9",
	})
	got := Project(basePolicy(o))
	if got.Obligations[0].State != StateApplicabilityUnresolved {
		t.Fatalf("state = %s, want %s: the newer decision must not win by recency", got.Obligations[0].State, StateApplicabilityUnresolved)
	}
}

func TestProjectNonconformanceOutranksLaterFavorableAssessment(t *testing.T) {
	t.Parallel()
	o := baseObligation()
	o.Assessments = []AssessmentRecord{
		{ID: "rasm_bad", SubjectRef: "s", ContextRef: "c", Outcome: Nonconforms, Argument: "a", Assessor: "w", ManifestID: "m", CreatedAt: "t2"},
		{ID: "rasm_good", SubjectRef: "s", ContextRef: "c", Outcome: Conforms, Argument: "a", Assessor: "w", ManifestID: "m", CheckAttemptIDs: []string{"rchk_1"}, CreatedAt: "t3"},
	}
	got := Project(basePolicy(o))
	if got.Obligations[0].State != StateNonconforms {
		t.Fatalf("state = %s, want %s: a demonstrated blocker is not erased by a later pass", got.Obligations[0].State, StateNonconforms)
	}
	if !got.Obligations[0].Contradiction {
		t.Fatal("disagreeing assessments must be reported as a contradiction")
	}
	if got.Decision != DecisionWithhold || len(got.Blockers) != 1 {
		t.Fatalf("decision = %s blockers = %v, want %s with one blocker", got.Decision, got.Blockers, DecisionWithhold)
	}
}

func TestProjectStaleAssessmentLosesCurrentAuthority(t *testing.T) {
	t.Parallel()
	o := baseObligation()
	o.Assessments[0].StaleDependency = true
	o.Assessments[0].StaleReason = "declared population moved"
	got := Project(basePolicy(o))
	if got.Decision != DecisionUndetermined || !contains(got.Reasons, ReasonStaleDependency) {
		t.Fatalf("decision = %s reasons = %v, want %s with %s", got.Decision, got.Reasons, DecisionUndetermined, ReasonStaleDependency)
	}
}

func TestProjectStaleNonconformanceDoesNotBlockCurrentDecision(t *testing.T) {
	t.Parallel()
	// A nonconformance about an obsolete context is history, not a current
	// blocker. It must not be silently dropped either: the current state comes
	// from the compatible assessment, and the stale record stays in the records.
	o := baseObligation()
	o.Assessments = []AssessmentRecord{
		{ID: "rasm_old", SubjectRef: "s", ContextRef: "old", Outcome: Nonconforms, Argument: "a", Assessor: "w",
			ManifestID: "m", CreatedAt: "t1", StaleDependency: true, StaleReason: "assessed an obsolete population"},
		{ID: "rasm_new", SubjectRef: "s", ContextRef: "new", Outcome: Conforms, Argument: "a", Assessor: "w",
			ManifestID: "m", CheckAttemptIDs: []string{"rchk_1"}, CreatedAt: "t2"},
	}
	got := Project(basePolicy(o))
	if got.Decision != DecisionEligible {
		t.Fatalf("decision = %s (reasons %v), want %s", got.Decision, got.Reasons, DecisionEligible)
	}
	if got.Obligations[0].GoverningAssessmentID != "rasm_new" {
		t.Fatalf("governing = %s, want rasm_new", got.Obligations[0].GoverningAssessmentID)
	}
	if !got.Obligations[0].Contradiction {
		t.Fatal("the superseded nonconformance must still be reported as a disagreement")
	}
}

func TestProjectUnauthorizedPolicyCannotGrantEligibility(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		mutate func(*PolicyRecords)
	}{
		{"no owner", func(p *PolicyRecords) { p.Owner = "" }},
		{"no authority source", func(p *PolicyRecords) { p.AuthoritySource = "" }},
		{"no scope justification", func(p *PolicyRecords) { p.ScopeJustification = "" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := basePolicy(baseObligation())
			tc.mutate(&p)
			got := Project(p)
			if got.Decision != DecisionUndetermined || !contains(got.Reasons, ReasonPolicyMissingOrUnauthorized) {
				t.Fatalf("decision = %s reasons = %v, want %s with %s",
					got.Decision, got.Reasons, DecisionUndetermined, ReasonPolicyMissingOrUnauthorized)
			}
		})
	}
}

func TestProjectVacuousPolicyIsNotSuccess(t *testing.T) {
	t.Parallel()
	o := baseObligation()
	o.Mandatory = false
	got := Project(basePolicy(o))
	if !got.Vacuous {
		t.Fatal("a policy binding no mandatory obligation must be reported vacuous")
	}
	if got.Decision != DecisionUndetermined {
		t.Fatalf("decision = %s, want %s: an empty mandatory set is not a pass", got.Decision, DecisionUndetermined)
	}
}

func TestProjectRecordedInapplicabilityIsNotUnexamined(t *testing.T) {
	t.Parallel()
	o := baseObligation()
	o.ApplicabilityDecisions[0].Decision = DoesNotApply
	o.Assessments = nil
	got := Project(basePolicy(o))
	if got.Obligations[0].State != StateNotApplicable {
		t.Fatalf("state = %s, want %s", got.Obligations[0].State, StateNotApplicable)
	}
	if got.Decision != DecisionEligible {
		t.Fatalf("decision = %s, want %s: an authorized exclusion resolves the obligation", got.Decision, DecisionEligible)
	}
}

func TestProjectWithholdStillReportsRemainingUncertainty(t *testing.T) {
	t.Parallel()
	blocker := baseObligation()
	blocker.ID, blocker.Key = "robl_bad", "bad"
	blocker.Assessments[0].Outcome = Nonconforms
	unexamined := baseObligation()
	unexamined.ID, unexamined.Key = "robl_unk", "unknown"
	unexamined.Assessments = nil

	got := Project(basePolicy(blocker, unexamined))
	if got.Decision != DecisionWithhold {
		t.Fatalf("decision = %s, want %s", got.Decision, DecisionWithhold)
	}
	// A blocker must not hide the fact that something else was never examined.
	if !contains(got.Reasons, ReasonUnexamined) {
		t.Fatalf("reasons = %v, want the unexamined obligation to remain visible", got.Reasons)
	}
}

func TestRenderCoverageIsDeterministicAndCitesRecords(t *testing.T) {
	t.Parallel()
	p := Project(basePolicy(baseObligation()))
	first := RenderCoverage(p)
	if first != RenderCoverage(p) {
		t.Fatal("rendering must be a pure function of the projection")
	}
	for _, want := range []string{"rpol_1", "robl_1", "rapp_1", "rasm_1", "rchk_1"} {
		if !strings.Contains(first, want) {
			t.Fatalf("document must cite record %s:\n%s", want, first)
		}
	}
}

func contains(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}
