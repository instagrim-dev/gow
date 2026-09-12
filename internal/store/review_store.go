package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/instagrim-dev/newf/internal/domain"
)

// This file is the storage mapping for the review contract's four record
// responsibilities (v44, G1 of the 2026-09-12 review-flow run).
//
// It is deliberately NOT reused scientific storage. A normative obligation
// REQUIRES a property; a CandidateInvariant CLAIMS a regularity over a
// conditioned population. Coercing one into the other is exactly what the
// contract forbids, so these rows have their own tables, own id kinds, and no
// promotion path into the scientific lifecycle.
//
// There is no coverage-status column. Coverage is generated from these records.

// ReviewPolicyRow is a versioned decision policy: the authority under which a
// review may reach a decision at all. Absent or unauthorized policy blocks
// eligibility rather than defaulting to a pass.
type ReviewPolicyRow struct {
	ID                 string
	PolicyKey          string
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
	CreatedAt          string
}

// ReviewObligationRow is one individually versioned normative obligation.
type ReviewObligationRow struct {
	ID                 string
	ObligationKey      string
	SemanticRevision   int
	Requirement        string
	AcceptanceCriteria string
	ApplicabilityRule  string
	PrimaryOwner       string
	CreatedAt          string
}

// ReviewPolicyObligationRow binds an obligation revision into a policy, with
// its mandatory flag. A policy whose mandatory set is empty is vacuous and must
// not grant eligibility; storing the flag makes that detectable.
type ReviewPolicyObligationRow struct {
	ObligationID string
	Mandatory    bool
}

// ReviewPolicyRecord is a policy plus the obligation revisions it pins.
type ReviewPolicyRecord struct {
	Policy      ReviewPolicyRow
	Obligations []ReviewPolicyObligationRow
}

// ReviewApplicabilityDecisionRow records whether an obligation applies to a
// named subject under a policy. Missing or conflicting applicability is
// unresolved, never a pass.
type ReviewApplicabilityDecisionRow struct {
	ID           string
	ObligationID string
	PolicyID     string
	SubjectRef   string
	Decision     string // applies | does_not_apply
	Rationale    string
	Authorizer   string
	CreatedAt    string
}

// ReviewManifestDependencyRow is one dependency an assessment's meaning rests
// on, with the reason it is relevant. `why_relevant` is required: an unexplained
// dependency cannot support a staleness judgment.
type ReviewManifestDependencyRow struct {
	Ordinal        int
	DependencyKind string
	DependencyRef  string
	WhyRelevant    string
}

// ReviewDependencyManifestRow is the dependency manifest of one assessment.
type ReviewDependencyManifestRow struct {
	ID              string
	PolicyID        string
	ObligationID    string
	ProjectRevision string
	ContractHash    string
	RecipeHash      string
	EvidenceCutoff  string
	CreatedAt       string
	Dependencies    []ReviewManifestDependencyRow
}

// ReviewCheckAttemptRow is one check attempt. Mode separates INSPECTION from
// EXECUTION; outcome separates a completed check, an executed-but-inconclusive
// check, and an execution blocker. The contract forbids collapsing these.
type ReviewCheckAttemptRow struct {
	ID                string
	ObligationID      string
	PolicyID          string
	CaseLabel         string
	ProcedureRef      string
	ProcedureRevision string
	InputsRef         string
	Executor          string
	Environment       string
	Mode              string // executed | inspected
	Outcome           string // completed | inconclusive | blocked
	OutputRef         string
	Blocker           string
	StartedAt         string
	EndedAt           string
	ResourceNote      string
	CreatedAt         string
}

// ReviewAssessmentRow is one assessment of one obligation revision against an
// exact subject and context, with its argument, assessor, applicability
// decision, dependency manifest and check references.
type ReviewAssessmentRow struct {
	ID                      string
	ObligationID            string
	PolicyID                string
	ApplicabilityDecisionID string
	ManifestID              string
	SubjectRef              string
	ContextRef              string
	Outcome                 string // conforms | nonconforms | inconclusive
	Argument                string
	Assessor                string
	CreatedAt               string
	CheckAttemptIDs         []string
}

// PersistReviewPolicy writes a decision policy and its pinned obligation
// revisions in one transaction. A policy without its obligation bindings would
// be an authority claim with no scope, so the two are never separable writes.
func (s *Store) PersistReviewPolicy(ctx context.Context, rec ReviewPolicyRecord) (ReviewPolicyRow, error) {
	if err := domain.ValidateReviewPolicyID(rec.Policy.ID); err != nil {
		return ReviewPolicyRow{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ReviewPolicyRow{}, err
	}
	defer tx.Rollback()

	p := rec.Policy
	if _, err := tx.ExecContext(ctx, `
INSERT INTO review_policies(id, policy_key, revision, decision_name, owner, authority_source, scope_justification, evidence_cutoff, case_budget, attempt_budget, provider_call_budget, supersedes_policy_id, supersede_rationale, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, p.ID, p.PolicyKey, p.Revision, p.DecisionName, p.Owner, p.AuthoritySource, p.ScopeJustification, p.EvidenceCutoff,
		p.CaseBudget, p.AttemptBudget, p.ProviderCallBudget, nullIfEmpty(p.SupersedesPolicyID), p.SupersedeRationale, p.CreatedAt); err != nil {
		return ReviewPolicyRow{}, err
	}
	for _, o := range rec.Obligations {
		if err := domain.ValidateReviewObligationID(o.ObligationID); err != nil {
			return ReviewPolicyRow{}, err
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO review_policy_obligations(policy_id, obligation_id, mandatory) VALUES(?, ?, ?)
`, p.ID, o.ObligationID, boolToInt(o.Mandatory)); err != nil {
			return ReviewPolicyRow{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return ReviewPolicyRow{}, err
	}
	return p, nil
}

// PersistReviewObligation writes one versioned obligation revision.
func (s *Store) PersistReviewObligation(ctx context.Context, o ReviewObligationRow) (ReviewObligationRow, error) {
	if err := domain.ValidateReviewObligationID(o.ID); err != nil {
		return ReviewObligationRow{}, err
	}
	if _, err := s.db.ExecContext(ctx, `
INSERT INTO review_obligations(id, obligation_key, semantic_revision, requirement, acceptance_criteria, applicability_rule, primary_owner, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?)
`, o.ID, o.ObligationKey, o.SemanticRevision, o.Requirement, o.AcceptanceCriteria, o.ApplicabilityRule, o.PrimaryOwner, o.CreatedAt); err != nil {
		return ReviewObligationRow{}, err
	}
	return o, nil
}

// PersistReviewApplicabilityDecision writes one applicability decision.
func (s *Store) PersistReviewApplicabilityDecision(ctx context.Context, d ReviewApplicabilityDecisionRow) (ReviewApplicabilityDecisionRow, error) {
	if err := domain.ValidateReviewApplicabilityDecisionID(d.ID); err != nil {
		return ReviewApplicabilityDecisionRow{}, err
	}
	if _, err := s.db.ExecContext(ctx, `
INSERT INTO review_applicability_decisions(id, obligation_id, policy_id, subject_ref, decision, rationale, authorizer, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?)
`, d.ID, d.ObligationID, d.PolicyID, d.SubjectRef, d.Decision, d.Rationale, d.Authorizer, d.CreatedAt); err != nil {
		return ReviewApplicabilityDecisionRow{}, err
	}
	return d, nil
}

// PersistReviewDependencyManifest writes a manifest and its dependencies in one
// transaction: a manifest with a partially written dependency list would
// understate what the assessment depends on.
func (s *Store) PersistReviewDependencyManifest(ctx context.Context, m ReviewDependencyManifestRow) (ReviewDependencyManifestRow, error) {
	if err := domain.ValidateReviewDependencyManifestID(m.ID); err != nil {
		return ReviewDependencyManifestRow{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ReviewDependencyManifestRow{}, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `
INSERT INTO review_dependency_manifests(id, policy_id, obligation_id, project_revision, contract_hash, recipe_hash, evidence_cutoff, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?)
`, m.ID, m.PolicyID, m.ObligationID, m.ProjectRevision, m.ContractHash, m.RecipeHash, m.EvidenceCutoff, m.CreatedAt); err != nil {
		return ReviewDependencyManifestRow{}, err
	}
	for i, d := range m.Dependencies {
		ordinal := d.Ordinal
		if ordinal == 0 {
			ordinal = i + 1
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO review_manifest_dependencies(manifest_id, ordinal, dependency_kind, dependency_ref, why_relevant)
VALUES(?, ?, ?, ?, ?)
`, m.ID, ordinal, d.DependencyKind, d.DependencyRef, d.WhyRelevant); err != nil {
			return ReviewDependencyManifestRow{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return ReviewDependencyManifestRow{}, err
	}
	return m, nil
}

// PersistReviewCheckAttempt writes one check attempt. A blocked attempt is
// retained exactly like a completed one — a blocker must stay visible.
func (s *Store) PersistReviewCheckAttempt(ctx context.Context, c ReviewCheckAttemptRow) (ReviewCheckAttemptRow, error) {
	if err := domain.ValidateReviewCheckAttemptID(c.ID); err != nil {
		return ReviewCheckAttemptRow{}, err
	}
	if c.Outcome == "blocked" && c.Blocker == "" {
		return ReviewCheckAttemptRow{}, fmt.Errorf("check attempt %s: a blocked outcome requires the blocker to be recorded", c.ID)
	}
	if _, err := s.db.ExecContext(ctx, `
INSERT INTO review_check_attempts(id, obligation_id, policy_id, case_label, procedure_ref, procedure_revision, inputs_ref, executor, environment, mode, outcome, output_ref, blocker, started_at, ended_at, resource_note, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, c.ID, c.ObligationID, c.PolicyID, c.CaseLabel, c.ProcedureRef, c.ProcedureRevision, c.InputsRef, c.Executor, c.Environment,
		c.Mode, c.Outcome, c.OutputRef, c.Blocker, c.StartedAt, c.EndedAt, c.ResourceNote, c.CreatedAt); err != nil {
		return ReviewCheckAttemptRow{}, err
	}
	return c, nil
}

// PersistReviewAssessment writes one assessment and its check references in one
// transaction. The schema trigger refuses to link a blocked attempt to a
// `conforms` assessment, so a blocker cannot be laundered into support.
func (s *Store) PersistReviewAssessment(ctx context.Context, a ReviewAssessmentRow) (ReviewAssessmentRow, error) {
	if err := domain.ValidateReviewAssessmentID(a.ID); err != nil {
		return ReviewAssessmentRow{}, err
	}
	if a.Outcome == "conforms" && len(a.CheckAttemptIDs) == 0 {
		return ReviewAssessmentRow{}, fmt.Errorf("assessment %s: a conforms outcome requires at least one check reference", a.ID)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ReviewAssessmentRow{}, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `
INSERT INTO review_assessments(id, obligation_id, policy_id, applicability_decision_id, manifest_id, subject_ref, context_ref, outcome, argument, assessor, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, a.ID, a.ObligationID, a.PolicyID, a.ApplicabilityDecisionID, a.ManifestID, a.SubjectRef, a.ContextRef, a.Outcome, a.Argument, a.Assessor, a.CreatedAt); err != nil {
		return ReviewAssessmentRow{}, err
	}
	for _, id := range a.CheckAttemptIDs {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO review_assessment_checks(assessment_id, check_attempt_id) VALUES(?, ?)
`, a.ID, id); err != nil {
			return ReviewAssessmentRow{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return ReviewAssessmentRow{}, err
	}
	return a, nil
}

// LatestReviewPolicy resolves the highest revision of a policy key.
func (s *Store) LatestReviewPolicy(ctx context.Context, policyKey string) (ReviewPolicyRow, bool, error) {
	row := s.db.QueryRowContext(ctx, reviewPolicySelect+`
WHERE policy_key = ? ORDER BY revision DESC LIMIT 1`, policyKey)
	p, err := scanReviewPolicy(row)
	if errors.Is(err, sql.ErrNoRows) {
		return ReviewPolicyRow{}, false, nil
	}
	if err != nil {
		return ReviewPolicyRow{}, false, err
	}
	return p, true, nil
}

// GetReviewPolicy loads one policy revision by id.
func (s *Store) GetReviewPolicy(ctx context.Context, id string) (ReviewPolicyRow, error) {
	if err := domain.ValidateReviewPolicyID(id); err != nil {
		return ReviewPolicyRow{}, err
	}
	p, err := scanReviewPolicy(s.db.QueryRowContext(ctx, reviewPolicySelect+` WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return ReviewPolicyRow{}, fmt.Errorf("%w: review policy %s", ErrNotFound, id)
	}
	return p, err
}

const reviewPolicySelect = `
SELECT id, policy_key, revision, decision_name, owner, authority_source, scope_justification, evidence_cutoff,
       case_budget, attempt_budget, provider_call_budget, COALESCE(supersedes_policy_id,''), supersede_rationale, created_at
FROM review_policies`

func scanReviewPolicy(row *sql.Row) (ReviewPolicyRow, error) {
	var p ReviewPolicyRow
	err := row.Scan(&p.ID, &p.PolicyKey, &p.Revision, &p.DecisionName, &p.Owner, &p.AuthoritySource, &p.ScopeJustification,
		&p.EvidenceCutoff, &p.CaseBudget, &p.AttemptBudget, &p.ProviderCallBudget, &p.SupersedesPolicyID, &p.SupersedeRationale, &p.CreatedAt)
	return p, err
}
