package pipeline

import (
	"context"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/review"
)

// This file holds the INSTANTIATION of the review contract's four record
// responsibilities for the single obligation `current-assessment-authority@1`
// (G1 of the 2026-09-12 review-flow run). The integrated C1-C8 scenario lives in
// review_obligation_integration_test.go and consumes these helpers.
//
// The obligation is a NORM, not a CandidateInvariant. The scientific claims the
// scenario exercises (mined predicates, their lifecycle states) are its
// SUBJECTS; they are never relabeled as obligations, and nothing here can
// promote one into the other.

const (
	// reviewPolicyKey is the pinned decision policy P1.
	reviewPolicyKey = "assessment-admission-decision"
	// reviewObligationKey is the single obligation under test.
	reviewObligationKey = "current-assessment-authority"
	// reviewObligationRevision is its semantic revision.
	reviewObligationRevision = 1
	// reviewRecipeRevision pins the recipe the cases come from.
	reviewRecipeRevision = "docs/reviews/prompts/recipes/assessment-admission-decision.md@1"
	// reviewContractRef pins the contract that defines the record mapping.
	reviewContractRef = "docs/reviews/prompts/review-contract.md"
	// reviewAssessor names who reached the assessment.
	reviewAssessor = "repository-gate:integration-test"

	// Dependency kinds. These are the ONLY things that can make an assessment
	// stale, which is what keeps C5 (unrelated change) decidable.
	depKindAssessmentPopulation = "assessment_population"
	depKindPolicyRevision       = "policy_revision"
	depKindCandidateContent     = "candidate_content"
)

// reviewLedger is the instantiated normative context for one scenario run.
type reviewLedger struct {
	policyID      string
	obligationID  string
	applicability string
}

// instantiateReviewObligation defines policy P1 with the single obligation and
// records the applicability decision for an exact subject.
//
// Every authority field is supplied concretely: an authorized owner, the source
// of that authority, and the scope justification. A policy missing any of them
// is refused by DefineReviewPolicy, so this helper cannot accidentally create a
// policy that would later hand out eligibility it was never entitled to grant.
func instantiateReviewObligation(t *testing.T, ctx context.Context, app *App, dbPath, subjectRef string) reviewLedger {
	t.Helper()

	policy, err := app.DefineReviewPolicy(ctx, ReviewPolicyDefineInput{
		DBPath:   dbPath,
		Key:      reviewPolicyKey,
		Revision: 1,
		// The concrete named decision — NOT "is the conjecture true".
		DecisionName:       "whether the selected candidate may guide the next search action",
		Owner:              "repository-maintainer",
		AuthoritySource:    reviewContractRef + " (four-record mapping) + " + reviewRecipeRevision,
		ScopeJustification: "bounded to one obligation over the pinned checkout's assessment/admission/decision path; no claim about domain conjectures",
		EvidenceCutoff:     "2026-09-12T12:00:00Z",
		CaseBudget:         8,
		AttemptBudget:      2,
		ProviderCallBudget: 0,
		Obligations: []ReviewObligationSpec{{
			Key:              reviewObligationKey,
			SemanticRevision: reviewObligationRevision,
			Requirement: "A current decision uses assessments compatible with its declared evidence " +
				"population and policy. Historical replay remains reproducible without restoring " +
				"obsolete current authority.",
			AcceptanceCriteria: "C1-C8 of " + reviewRecipeRevision + ": exact record ids explain the selected action; " +
				"a relevant population change makes the affected assessment stale; an unrelated change does not; " +
				"reassessment is the only path back to current eligibility; replay stays reproducible without " +
				"restoring current authority; coverage is deterministically derived from records.",
			ApplicabilityRule: "applies to any decision that selects a candidate to guide the next search action " +
				"from a persisted invariant lifecycle state",
			PrimaryOwner: "repository-maintainer",
			Mandatory:    true,
		}},
	})
	if err != nil {
		t.Fatalf("define review policy: %v", err)
	}
	obligationID := policy.ObligationIDs[obligationKey(reviewObligationKey, reviewObligationRevision)]
	if obligationID == "" {
		t.Fatalf("obligation id missing from %+v", policy.ObligationIDs)
	}

	decision, err := app.DecideReviewApplicability(ctx, ReviewApplicabilityInput{
		DBPath: dbPath, PolicyID: policy.PolicyID, ObligationID: obligationID,
		SubjectRef: subjectRef, Decision: review.Applies,
		Rationale:  "the subject is a persisted candidate whose lifecycle state is read to select the next frontier target, which is exactly the decision this policy names",
		Authorizer: "repository-maintainer",
	})
	if err != nil {
		t.Fatalf("decide applicability: %v", err)
	}
	return reviewLedger{policyID: policy.PolicyID, obligationID: obligationID, applicability: decision.ID}
}

// recordCase records one executed C-case check attempt and returns its id.
//
// `mode=executed` is used only where the scenario really ran the path through
// the migrated store. Anything merely read is recorded as `inspected`, which the
// projection refuses to accept as support for conformance.
func (l reviewLedger) recordCase(t *testing.T, ctx context.Context, app *App, dbPath, caseLabel, procedureRef, inputsRef, outcome, outputRef, blocker string) string {
	t.Helper()
	at := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC).Format(timeLayout)
	row, err := app.RecordReviewCheck(ctx, ReviewCheckInput{
		DBPath: dbPath, PolicyID: l.policyID, ObligationID: l.obligationID,
		CaseLabel: caseLabel, ProcedureRef: procedureRef, ProcedureRevision: reviewRecipeRevision,
		InputsRef: inputsRef, Executor: reviewAssessor, Environment: "go test ./internal/pipeline",
		Mode: review.ModeExecuted, Outcome: outcome, OutputRef: outputRef, Blocker: blocker,
		StartedAt: at, EndedAt: at,
		ResourceNote: "zero paid-provider calls; fixture providers only; disposable store",
	})
	if err != nil {
		t.Fatalf("record check %s: %v", caseLabel, err)
	}
	return row.ID
}
