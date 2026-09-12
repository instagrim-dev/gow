package store

import (
	"context"
	"fmt"
)

// This file is the READ side of the normative review ledger: the single
// authoritative projection coverage generation is allowed to consume.
//
// It reads records only. It computes no status, invents no assessment, and has
// no writer — so a generated view can never be a second editable status field.
// Absence is preserved as absence: an obligation with no assessment comes back
// with none, which the generator must report as `unexamined`.

// ReviewObligationCoverage is everything the ledger holds about ONE obligation
// revision under ONE policy.
type ReviewObligationCoverage struct {
	Obligation ReviewObligationRow
	Mandatory  bool
	// Applicability holds every applicability decision recorded for this
	// obligation under the policy, oldest first. Zero decisions means
	// applicability is UNRESOLVED; more than one distinct decision value means
	// CONFLICTING. Both are unresolved, not a pass — so the generator needs the
	// whole list, not a "current" pick.
	Applicability []ReviewApplicabilityDecisionRow
	// Assessments are every assessment, oldest first. Contradictory evidence is
	// retained rather than resolved by a latest-timestamp rule.
	Assessments []ReviewAssessmentRow
	// Manifests are the dependency manifests cited by those assessments, keyed
	// by manifest id.
	Manifests map[string]ReviewDependencyManifestRow
	// Checks are every check attempt recorded for this obligation, oldest
	// first — including blocked and inconclusive ones.
	Checks []ReviewCheckAttemptRow
}

// ReviewCoverage is the ledger projection for one decision policy.
type ReviewCoverage struct {
	Policy      ReviewPolicyRow
	Obligations []ReviewObligationCoverage
}

// LoadReviewCoverage assembles the coverage projection for a policy revision.
// Ordering is fully deterministic (obligation key/revision, then created_at with
// an id tiebreak) so repeated generation from identical inputs is byte-identical
// apart from explicitly non-semantic metadata.
func (s *Store) LoadReviewCoverage(ctx context.Context, policyID string) (ReviewCoverage, error) {
	policy, err := s.GetReviewPolicy(ctx, policyID)
	if err != nil {
		return ReviewCoverage{}, err
	}
	out := ReviewCoverage{Policy: policy}

	rows, err := s.db.QueryContext(ctx, `
SELECT o.id, o.obligation_key, o.semantic_revision, o.requirement, o.acceptance_criteria, o.applicability_rule, o.primary_owner, o.created_at, po.mandatory
FROM review_policy_obligations po
JOIN review_obligations o ON o.id = po.obligation_id
WHERE po.policy_id = ?
ORDER BY o.obligation_key, o.semantic_revision, o.id
`, policyID)
	if err != nil {
		return ReviewCoverage{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var c ReviewObligationCoverage
		var mandatory int
		if err := rows.Scan(&c.Obligation.ID, &c.Obligation.ObligationKey, &c.Obligation.SemanticRevision,
			&c.Obligation.Requirement, &c.Obligation.AcceptanceCriteria, &c.Obligation.ApplicabilityRule,
			&c.Obligation.PrimaryOwner, &c.Obligation.CreatedAt, &mandatory); err != nil {
			return ReviewCoverage{}, err
		}
		c.Mandatory = mandatory != 0
		c.Manifests = map[string]ReviewDependencyManifestRow{}
		out.Obligations = append(out.Obligations, c)
	}
	if err := rows.Err(); err != nil {
		return ReviewCoverage{}, err
	}

	for i := range out.Obligations {
		obligationID := out.Obligations[i].Obligation.ID
		if err := s.loadReviewApplicability(ctx, policyID, obligationID, &out.Obligations[i]); err != nil {
			return ReviewCoverage{}, err
		}
		if err := s.loadReviewChecks(ctx, policyID, obligationID, &out.Obligations[i]); err != nil {
			return ReviewCoverage{}, err
		}
		if err := s.loadReviewAssessments(ctx, policyID, obligationID, &out.Obligations[i]); err != nil {
			return ReviewCoverage{}, err
		}
	}
	return out, nil
}

func (s *Store) loadReviewApplicability(ctx context.Context, policyID, obligationID string, c *ReviewObligationCoverage) error {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, obligation_id, policy_id, subject_ref, decision, rationale, authorizer, created_at
FROM review_applicability_decisions
WHERE policy_id = ? AND obligation_id = ?
ORDER BY created_at, id
`, policyID, obligationID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var d ReviewApplicabilityDecisionRow
		if err := rows.Scan(&d.ID, &d.ObligationID, &d.PolicyID, &d.SubjectRef, &d.Decision, &d.Rationale, &d.Authorizer, &d.CreatedAt); err != nil {
			return err
		}
		c.Applicability = append(c.Applicability, d)
	}
	return rows.Err()
}

func (s *Store) loadReviewChecks(ctx context.Context, policyID, obligationID string, c *ReviewObligationCoverage) error {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, obligation_id, policy_id, case_label, procedure_ref, procedure_revision, inputs_ref, executor, environment,
       mode, outcome, output_ref, blocker, started_at, ended_at, resource_note, created_at
FROM review_check_attempts
WHERE policy_id = ? AND obligation_id = ?
ORDER BY case_label, created_at, id
`, policyID, obligationID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var a ReviewCheckAttemptRow
		if err := rows.Scan(&a.ID, &a.ObligationID, &a.PolicyID, &a.CaseLabel, &a.ProcedureRef, &a.ProcedureRevision,
			&a.InputsRef, &a.Executor, &a.Environment, &a.Mode, &a.Outcome, &a.OutputRef, &a.Blocker,
			&a.StartedAt, &a.EndedAt, &a.ResourceNote, &a.CreatedAt); err != nil {
			return err
		}
		c.Checks = append(c.Checks, a)
	}
	return rows.Err()
}

func (s *Store) loadReviewAssessments(ctx context.Context, policyID, obligationID string, c *ReviewObligationCoverage) error {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, obligation_id, policy_id, applicability_decision_id, manifest_id, subject_ref, context_ref, outcome, argument, assessor, created_at
FROM review_assessments
WHERE policy_id = ? AND obligation_id = ?
ORDER BY created_at, id
`, policyID, obligationID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var a ReviewAssessmentRow
		if err := rows.Scan(&a.ID, &a.ObligationID, &a.PolicyID, &a.ApplicabilityDecisionID, &a.ManifestID,
			&a.SubjectRef, &a.ContextRef, &a.Outcome, &a.Argument, &a.Assessor, &a.CreatedAt); err != nil {
			return err
		}
		c.Assessments = append(c.Assessments, a)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	rows.Close()

	for i := range c.Assessments {
		ids, err := s.reviewAssessmentCheckIDs(ctx, c.Assessments[i].ID)
		if err != nil {
			return err
		}
		c.Assessments[i].CheckAttemptIDs = ids
		manifest, err := s.getReviewDependencyManifest(ctx, c.Assessments[i].ManifestID)
		if err != nil {
			return err
		}
		c.Manifests[manifest.ID] = manifest
	}
	return nil
}

func (s *Store) reviewAssessmentCheckIDs(ctx context.Context, assessmentID string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT check_attempt_id FROM review_assessment_checks WHERE assessment_id = ? ORDER BY check_attempt_id
`, assessmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func (s *Store) getReviewDependencyManifest(ctx context.Context, id string) (ReviewDependencyManifestRow, error) {
	var m ReviewDependencyManifestRow
	if err := s.db.QueryRowContext(ctx, `
SELECT id, policy_id, obligation_id, project_revision, contract_hash, recipe_hash, evidence_cutoff, created_at
FROM review_dependency_manifests WHERE id = ?
`, id).Scan(&m.ID, &m.PolicyID, &m.ObligationID, &m.ProjectRevision, &m.ContractHash, &m.RecipeHash, &m.EvidenceCutoff, &m.CreatedAt); err != nil {
		return ReviewDependencyManifestRow{}, fmt.Errorf("dependency manifest %s: %w", id, err)
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT ordinal, dependency_kind, dependency_ref, why_relevant
FROM review_manifest_dependencies WHERE manifest_id = ? ORDER BY ordinal
`, id)
	if err != nil {
		return ReviewDependencyManifestRow{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var d ReviewManifestDependencyRow
		if err := rows.Scan(&d.Ordinal, &d.DependencyKind, &d.DependencyRef, &d.WhyRelevant); err != nil {
			return ReviewDependencyManifestRow{}, err
		}
		m.Dependencies = append(m.Dependencies, d)
	}
	return m, rows.Err()
}
