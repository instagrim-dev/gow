package store

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/domain"
)

func TestReviewAssessmentRequiresReferenceScopeMatch(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st, err := Open(t.TempDir() + "/review-scope.sqlite")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer st.Close()
	if err := st.Migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	now := time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC)
	created := now.Format(time.RFC3339)
	obligation := ReviewObligationRow{
		ID:                 domain.NewReviewObligationID(now),
		ObligationKey:      "subject-binding",
		SemanticRevision:   1,
		Requirement:        "assessment references match their subject",
		AcceptanceCriteria: "mismatched applicability cannot be cited",
		ApplicabilityRule:  "applies to exact subjects only",
		PrimaryOwner:       "test",
		CreatedAt:          created,
	}
	if _, err := st.PersistReviewObligation(ctx, obligation); err != nil {
		t.Fatalf("obligation: %v", err)
	}
	policy := ReviewPolicyRecord{
		Policy: ReviewPolicyRow{
			ID:                 domain.NewReviewPolicyID(now),
			PolicyKey:          "subject-binding",
			Revision:           1,
			DecisionName:       "whether an assessment is scoped to its applicability",
			Owner:              "test",
			AuthoritySource:    "test",
			ScopeJustification: "store boundary regression",
			CreatedAt:          created,
		},
		Obligations: []ReviewPolicyObligationRow{{ObligationID: obligation.ID, Mandatory: true}},
	}
	if _, err := st.PersistReviewPolicy(ctx, policy); err != nil {
		t.Fatalf("policy: %v", err)
	}
	applicability, err := st.PersistReviewApplicabilityDecision(ctx, ReviewApplicabilityDecisionRow{
		ID:           domain.NewReviewApplicabilityDecisionID(now),
		PolicyID:     policy.Policy.ID,
		ObligationID: obligation.ID,
		SubjectRef:   "subject:A",
		Decision:     "applies",
		Rationale:    "subject A is in scope",
		Authorizer:   "test",
		CreatedAt:    created,
	})
	if err != nil {
		t.Fatalf("applicability: %v", err)
	}
	manifest, err := st.PersistReviewDependencyManifest(ctx, ReviewDependencyManifestRow{
		ID:              domain.NewReviewDependencyManifestID(now),
		PolicyID:        policy.Policy.ID,
		ObligationID:    obligation.ID,
		ProjectRevision: "test-revision",
		CreatedAt:       created,
		Dependencies: []ReviewManifestDependencyRow{{
			Ordinal:        1,
			DependencyKind: "candidate_content",
			DependencyRef:  "subject:A",
			WhyRelevant:    "different content is a different subject",
		}},
	})
	if err != nil {
		t.Fatalf("manifest: %v", err)
	}

	bad := ReviewAssessmentRow{
		ID:                      domain.NewReviewAssessmentID(now),
		PolicyID:                policy.Policy.ID,
		ObligationID:            obligation.ID,
		ApplicabilityDecisionID: applicability.ID,
		ManifestID:              manifest.ID,
		SubjectRef:              "subject:B",
		ContextRef:              "ctx",
		Outcome:                 "inconclusive",
		Argument:                "this tries to cite A's applicability for B",
		Assessor:                "test",
		CreatedAt:               created,
	}
	if _, err := st.PersistReviewAssessment(ctx, bad); err == nil || !strings.Contains(err.Error(), "not assessment policy") {
		t.Fatalf("mismatched applicability must be refused by the store, got %v", err)
	}

	matching := bad
	matching.ID = domain.NewReviewAssessmentID(now.Add(time.Second))
	matching.SubjectRef = "subject:A"
	if _, err := st.PersistReviewAssessment(ctx, matching); err != nil {
		t.Fatalf("matching applicability should persist: %v", err)
	}

	raw := bad
	raw.ID = domain.NewReviewAssessmentID(now.Add(2 * time.Second))
	_, err = st.db.ExecContext(ctx, `
INSERT INTO review_assessments(id, obligation_id, policy_id, applicability_decision_id, manifest_id, subject_ref, context_ref, outcome, argument, assessor, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, raw.ID, raw.ObligationID, raw.PolicyID, raw.ApplicabilityDecisionID, raw.ManifestID, raw.SubjectRef, raw.ContextRef, raw.Outcome, raw.Argument, raw.Assessor, raw.CreatedAt)
	if err == nil || !strings.Contains(err.Error(), "applicability decision scope does not match assessment") {
		t.Fatalf("mismatched raw SQL insert must be refused by the trigger, got %v", err)
	}
}

func TestMigrateV47AddsReviewAssessmentScopeTrigger(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st, err := Open(t.TempDir() + "/review-v47.sqlite")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer st.Close()
	if err := st.Migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if _, err := st.db.ExecContext(ctx, `DROP TRIGGER review_assessments_reference_scope`); err != nil {
		t.Fatalf("drop trigger: %v", err)
	}
	if _, err := st.db.ExecContext(ctx, `DELETE FROM schema_migrations WHERE version = 47`); err != nil {
		t.Fatalf("unstamp v47: %v", err)
	}
	if err := st.Migrate(ctx); err != nil {
		t.Fatalf("rerun v47: %v", err)
	}
	var count int
	if err := st.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM sqlite_master
WHERE type = 'trigger' AND name = 'review_assessments_reference_scope'
`).Scan(&count); err != nil {
		t.Fatalf("query trigger: %v", err)
	}
	if count != 1 {
		t.Fatalf("review_assessments_reference_scope trigger count = %d, want 1", count)
	}
}
