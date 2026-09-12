package pipeline

import (
	"context"
	"fmt"
	"strings"

	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/review"
	"github.com/instagrim-dev/newf/internal/store"
)

// This file is the application surface for the normative review ledger (G1 of
// the 2026-09-12 review-flow run): define a policy and its obligations, decide
// applicability, record check attempts, record assessments with their dependency
// manifests, and GENERATE coverage.
//
// Deliberately absent: any way to write a coverage status. Coverage is derived
// on read from the records below. If a caller wants a different coverage
// document, it must change the records — which is the whole point, because a
// hand-patched status is indistinguishable from a real assessment once written.

// ReviewObligationSpec is one obligation revision to bind into a policy.
type ReviewObligationSpec struct {
	Key                string
	SemanticRevision   int
	Requirement        string
	AcceptanceCriteria string
	ApplicabilityRule  string
	PrimaryOwner       string
	Mandatory          bool
}

// ReviewPolicyDefineInput defines a decision policy revision and the obligation
// revisions it binds.
type ReviewPolicyDefineInput struct {
	DBPath             string
	Key                string
	Revision           int
	DecisionName       string
	Owner              string
	AuthoritySource    string
	ScopeJustification string
	EvidenceCutoff     string
	CaseBudget         int
	AttemptBudget      int
	ProviderCallBudget int
	SupersedesPolicyID string
	SupersedeRationale string
	Obligations        []ReviewObligationSpec
}

// ReviewPolicyDefineResponse reports the created policy and obligation ids.
type ReviewPolicyDefineResponse struct {
	PolicyID      string            `json:"policy_id"`
	ObligationIDs map[string]string `json:"obligation_ids"`
}

// DefineReviewPolicy writes a policy revision with its obligation revisions.
//
// The authority fields are required here rather than defaulted: a policy with no
// owner, no authority source, or no scope justification cannot authorize a
// decision, and accepting one would let an unauthorized policy silently produce
// an eligibility verdict later.
func (a *App) DefineReviewPolicy(ctx context.Context, in ReviewPolicyDefineInput) (ReviewPolicyDefineResponse, error) {
	switch {
	case strings.TrimSpace(in.Key) == "":
		return ReviewPolicyDefineResponse{}, fmt.Errorf("policy key is required")
	case in.Revision <= 0:
		return ReviewPolicyDefineResponse{}, fmt.Errorf("policy revision must be positive")
	case strings.TrimSpace(in.DecisionName) == "":
		return ReviewPolicyDefineResponse{}, fmt.Errorf("the concrete named decision is required: a policy that does not say what it decides cannot scope one")
	case strings.TrimSpace(in.Owner) == "", strings.TrimSpace(in.AuthoritySource) == "", strings.TrimSpace(in.ScopeJustification) == "":
		return ReviewPolicyDefineResponse{}, fmt.Errorf("owner, authority source and scope justification are required: an unauthorized policy must not be able to grant eligibility")
	case len(in.Obligations) == 0:
		return ReviewPolicyDefineResponse{}, fmt.Errorf("a policy with no obligations is vacuous")
	}

	_, repoStore, err := a.openStoreFn(ctx, in.DBPath)
	if err != nil {
		return ReviewPolicyDefineResponse{}, err
	}
	defer repoStore.Close()

	now := a.now()
	ts := now.Format(timeLayout)
	out := ReviewPolicyDefineResponse{
		PolicyID:      domain.NewReviewPolicyID(now),
		ObligationIDs: map[string]string{},
	}
	rec := store.ReviewPolicyRecord{Policy: store.ReviewPolicyRow{
		ID: out.PolicyID, PolicyKey: in.Key, Revision: in.Revision,
		DecisionName: in.DecisionName, Owner: in.Owner, AuthoritySource: in.AuthoritySource,
		ScopeJustification: in.ScopeJustification, EvidenceCutoff: in.EvidenceCutoff,
		CaseBudget: in.CaseBudget, AttemptBudget: in.AttemptBudget, ProviderCallBudget: in.ProviderCallBudget,
		SupersedesPolicyID: in.SupersedesPolicyID, SupersedeRationale: in.SupersedeRationale,
		CreatedAt: ts,
	}}
	for i, spec := range in.Obligations {
		if strings.TrimSpace(spec.Key) == "" || spec.SemanticRevision <= 0 {
			return ReviewPolicyDefineResponse{}, fmt.Errorf("obligation %d: key and positive semantic revision are required", i)
		}
		if strings.TrimSpace(spec.Requirement) == "" || strings.TrimSpace(spec.AcceptanceCriteria) == "" {
			return ReviewPolicyDefineResponse{}, fmt.Errorf("obligation %s: requirement and acceptance criteria are required", spec.Key)
		}
		if strings.TrimSpace(spec.ApplicabilityRule) == "" {
			// Without a rule, "does this apply?" has no answer to appeal to, so
			// every later applicability decision would be unreviewable assertion.
			return ReviewPolicyDefineResponse{}, fmt.Errorf("obligation %s: an applicability rule is required", spec.Key)
		}
		if strings.TrimSpace(spec.PrimaryOwner) == "" {
			// An unowned obligation has nobody to answer for a nonconformance,
			// which is how a finding becomes nobody's problem.
			return ReviewPolicyDefineResponse{}, fmt.Errorf("obligation %s: a primary owner is required", spec.Key)
		}
		obligationID := domain.NewReviewObligationID(now)
		if _, err := repoStore.PersistReviewObligation(ctx, store.ReviewObligationRow{
			ID: obligationID, ObligationKey: spec.Key, SemanticRevision: spec.SemanticRevision,
			Requirement: spec.Requirement, AcceptanceCriteria: spec.AcceptanceCriteria,
			ApplicabilityRule: spec.ApplicabilityRule, PrimaryOwner: spec.PrimaryOwner, CreatedAt: ts,
		}); err != nil {
			return ReviewPolicyDefineResponse{}, err
		}
		out.ObligationIDs[obligationKey(spec.Key, spec.SemanticRevision)] = obligationID
		rec.Obligations = append(rec.Obligations, store.ReviewPolicyObligationRow{
			ObligationID: obligationID, Mandatory: spec.Mandatory,
		})
	}
	if _, err := repoStore.PersistReviewPolicy(ctx, rec); err != nil {
		return ReviewPolicyDefineResponse{}, err
	}
	return out, nil
}

// obligationKey is the stable map key for an obligation revision.
func obligationKey(key string, revision int) string {
	return fmt.Sprintf("%s@%d", key, revision)
}

// ReviewApplicabilityInput records whether an obligation applies to a subject.
type ReviewApplicabilityInput struct {
	DBPath       string
	PolicyID     string
	ObligationID string
	SubjectRef   string
	Decision     string
	Rationale    string
	Authorizer   string
}

// DecideReviewApplicability records one applicability decision.
//
// A rationale and an authorizer are mandatory, including for `does_not_apply`:
// an unexplained exclusion is how missing implementation gets recorded as
// out-of-scope.
func (a *App) DecideReviewApplicability(ctx context.Context, in ReviewApplicabilityInput) (store.ReviewApplicabilityDecisionRow, error) {
	if in.Decision != review.Applies && in.Decision != review.DoesNotApply {
		return store.ReviewApplicabilityDecisionRow{}, fmt.Errorf("applicability decision must be %q or %q, got %q", review.Applies, review.DoesNotApply, in.Decision)
	}
	if strings.TrimSpace(in.Rationale) == "" || strings.TrimSpace(in.Authorizer) == "" {
		return store.ReviewApplicabilityDecisionRow{}, fmt.Errorf("applicability requires a rationale and an authorizer")
	}
	if strings.TrimSpace(in.SubjectRef) == "" {
		return store.ReviewApplicabilityDecisionRow{}, fmt.Errorf("applicability requires an exact subject reference")
	}
	_, repoStore, err := a.openStoreFn(ctx, in.DBPath)
	if err != nil {
		return store.ReviewApplicabilityDecisionRow{}, err
	}
	defer repoStore.Close()
	now := a.now()
	return repoStore.PersistReviewApplicabilityDecision(ctx, store.ReviewApplicabilityDecisionRow{
		ID: domain.NewReviewApplicabilityDecisionID(now), ObligationID: in.ObligationID, PolicyID: in.PolicyID,
		SubjectRef: in.SubjectRef, Decision: in.Decision, Rationale: in.Rationale, Authorizer: in.Authorizer,
		CreatedAt: now.Format(timeLayout),
	})
}

// ReviewCheckInput records one check attempt.
type ReviewCheckInput struct {
	DBPath            string
	PolicyID          string
	ObligationID      string
	CaseLabel         string
	ProcedureRef      string
	ProcedureRevision string
	InputsRef         string
	Executor          string
	Environment       string
	Mode              string
	Outcome           string
	OutputRef         string
	Blocker           string
	StartedAt         string
	EndedAt           string
	ResourceNote      string
}

// RecordReviewCheck persists one check attempt.
//
// Mode and outcome are separate axes on purpose. An INSPECTED procedure did not
// run, so it can never carry the weight of an executed one; and `blocked` is
// retained as a distinct outcome instead of being folded into `inconclusive`,
// because "we could not run it" and "we ran it and learned nothing" call for
// different follow-up.
func (a *App) RecordReviewCheck(ctx context.Context, in ReviewCheckInput) (store.ReviewCheckAttemptRow, error) {
	switch in.Mode {
	case review.ModeExecuted, review.ModeInspected:
	default:
		return store.ReviewCheckAttemptRow{}, fmt.Errorf("check mode must be %q or %q, got %q", review.ModeExecuted, review.ModeInspected, in.Mode)
	}
	switch in.Outcome {
	case review.CheckCompleted, review.CheckInconclusive, review.CheckBlocked:
	default:
		return store.ReviewCheckAttemptRow{}, fmt.Errorf("check outcome must be completed|inconclusive|blocked, got %q", in.Outcome)
	}
	if in.Mode == review.ModeInspected && in.Outcome == review.CheckCompleted {
		// An inspection cannot complete a check: reading a procedure is not
		// running it. Allowing this would let source inspection certify
		// unexecuted verification.
		return store.ReviewCheckAttemptRow{}, fmt.Errorf("an inspected procedure cannot be recorded as completed: inspection is not execution")
	}
	if strings.TrimSpace(in.CaseLabel) == "" || strings.TrimSpace(in.ProcedureRef) == "" {
		return store.ReviewCheckAttemptRow{}, fmt.Errorf("a check attempt requires a case label and a procedure reference")
	}
	if strings.TrimSpace(in.ProcedureRevision) == "" {
		// A procedure reference without a revision cannot be replayed: "we ran
		// the checker" is not a reproducible claim about which checker.
		return store.ReviewCheckAttemptRow{}, fmt.Errorf("a check attempt requires the procedure revision it executed")
	}
	if strings.TrimSpace(in.Executor) == "" || strings.TrimSpace(in.Environment) == "" {
		return store.ReviewCheckAttemptRow{}, fmt.Errorf("a check attempt requires an executor and an environment")
	}
	_, repoStore, err := a.openStoreFn(ctx, in.DBPath)
	if err != nil {
		return store.ReviewCheckAttemptRow{}, err
	}
	defer repoStore.Close()
	now := a.now()
	return repoStore.PersistReviewCheckAttempt(ctx, store.ReviewCheckAttemptRow{
		ID: domain.NewReviewCheckAttemptID(now), ObligationID: in.ObligationID, PolicyID: in.PolicyID,
		CaseLabel: in.CaseLabel, ProcedureRef: in.ProcedureRef, ProcedureRevision: in.ProcedureRevision,
		InputsRef: in.InputsRef, Executor: in.Executor, Environment: in.Environment,
		Mode: in.Mode, Outcome: in.Outcome, OutputRef: in.OutputRef, Blocker: in.Blocker,
		StartedAt: in.StartedAt, EndedAt: in.EndedAt, ResourceNote: in.ResourceNote,
		CreatedAt: now.Format(timeLayout),
	})
}
