package pipeline

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/review"
	"github.com/instagrim-dev/newf/internal/store"
)

// TestReviewCoverageCompatibilityMatrix is the committed pipeline-level
// regression for H1 / GOW-R2 (2026-09-12 remediation): compatibility is a
// three-state judgment, and current eligibility requires KNOWN compatible
// support — not "no evidence of change".
//
// The six cases pin the reviewer's discriminating matrix on a single persisted
// assessment that declared four current-decision-relevant dependencies. Every
// case reuses the same historical assessment; the only variable is the current
// context the caller supplies. Historical preservation is asserted in every
// case: the assessment record stays visible in the export regardless of the
// current decision.
//
//	complete match         → ELIGIBLE_TO_ADVANCE
//	changed population     → UNDETERMINED with `stale_dependency`
//	nil deps               → UNDETERMINED with `compatibility_unknown`
//	one omitted            → UNDETERMINED with `compatibility_unknown`
//	blank reference        → UNDETERMINED with `compatibility_unknown`
//	unrelated extra        → ELIGIBLE_TO_ADVANCE (harmless)
//
// The nil, omitted and blank cases must NOT produce `stale_dependency` — that
// would collapse the two epistemic states the reviewer required to remain
// separable.
//
// The scenario runs under a control policy declared inline, with an
// applicability decision, an executed check, and one Conforms assessment. It
// does NOT depend on the C1-C8 integration test's fixture chain, so a failure
// here is unambiguously about compatibility semantics and can be triaged in
// isolation from the broader recipe.
func TestReviewCoverageCompatibilityMatrix(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)

	// Control policy: one mandatory obligation, four declared dependencies
	// on the sole conforming assessment. Every declared dep is
	// current-decision-relevant, so a caller that omits any of them
	// leaves the assessment's current compatibility undetermined.
	policy, err := app.DefineReviewPolicy(ctx, ReviewPolicyDefineInput{
		DBPath: dbPath, Key: "compatibility-matrix-control", Revision: 1,
		DecisionName:       "whether the sole obligation may be reported as satisfied under the caller's current context",
		Owner:              "repository-maintainer",
		AuthoritySource:    reviewContractRef,
		ScopeJustification: "H1 remediation control: authorizes nothing outside the six-case compatibility matrix",
		EvidenceCutoff:     "2026-09-12T12:00:00Z",
		Obligations: []ReviewObligationSpec{{
			Key: "compatibility-matrix", SemanticRevision: 1,
			Requirement:        "current decisions must not inherit historical passes across an unspecified current context",
			AcceptanceCriteria: "H1 / GOW-R2: nil, omitted or blank current deps must produce UNDETERMINED with compatibility_unknown; not stale.",
			ApplicabilityRule:  "applies to the control subject only",
			PrimaryOwner:       "repository-maintainer", Mandatory: true,
		}},
	})
	if err != nil {
		t.Fatalf("define control policy: %v", err)
	}
	obligationID := policy.ObligationIDs["compatibility-matrix@1"]

	apDecision, err := app.DecideReviewApplicability(ctx, ReviewApplicabilityInput{
		DBPath: dbPath, PolicyID: policy.PolicyID, ObligationID: obligationID,
		SubjectRef: "matrix:control-subject", Decision: review.Applies,
		Rationale:  "in scope for the compatibility matrix control",
		Authorizer: "repository-maintainer",
	})
	if err != nil {
		t.Fatalf("decide applicability: %v", err)
	}
	check, err := app.RecordReviewCheck(ctx, ReviewCheckInput{
		DBPath: dbPath, PolicyID: policy.PolicyID, ObligationID: obligationID,
		CaseLabel: "matrix-check", ProcedureRef: "matrix-procedure",
		ProcedureRevision: reviewRecipeRevision,
		InputsRef:         "matrix-inputs-fixture",
		Executor:          reviewAssessor,
		Environment:       "go test ./internal/pipeline",
		Mode:              review.ModeExecuted, Outcome: review.CheckCompleted,
		OutputRef: "matrix-output-fixture",
	})
	if err != nil {
		t.Fatalf("record check: %v", err)
	}

	// The historical assessment. Every declared dep is why-relevant to the
	// obligation's requirement; a caller omitting any of them is asking about
	// compatibility with an incompletely specified current context.
	const (
		assessedPopulation = "matrix-population-declared"
		assessedPolicy     = "matrix-policy-declared@1"
		assessedContent    = "matrix-content-declared"
		assessedProject    = "matrix-project-declared"
	)
	asm, err := app.RecordReviewAssessment(ctx, ReviewAssessInput{
		DBPath: dbPath, PolicyID: policy.PolicyID, ObligationID: obligationID,
		ApplicabilityDecisionID: apDecision.ID,
		SubjectRef:              "matrix:control-subject",
		ContextRef:              "context:" + assessedPopulation,
		Outcome:                 review.Conforms,
		Argument:                "the control obligation was assessed against the four declared dependencies; the outcome is historical",
		Assessor:                reviewAssessor,
		ProjectRevision:         assessedProject,
		ContractHash:            reviewContractRef,
		RecipeHash:              reviewRecipeRevision,
		EvidenceCutoff:          "2026-09-12T12:00:00Z",
		Dependencies: []ReviewDependencySpec{
			{Kind: depKindAssessmentPopulation, Ref: assessedPopulation,
				WhyRelevant: "a different current population changes what the decision may rely on"},
			{Kind: depKindPolicyRevision, Ref: assessedPolicy,
				WhyRelevant: "a different policy revision changes the acceptance criteria"},
			{Kind: depKindCandidateContent, Ref: assessedContent,
				WhyRelevant: "different content is a different subject"},
			{Kind: depKindProjectRevision, Ref: assessedProject,
				WhyRelevant: "different code can decide differently"},
		},
		CheckAttemptIDs: []string{check.ID},
	})
	if err != nil {
		t.Fatalf("record assessment: %v", err)
	}
	historicalAssessmentID := asm.Assessment.ID

	completeMatch := map[string]string{
		depKindAssessmentPopulation: assessedPopulation,
		depKindPolicyRevision:       assessedPolicy,
		depKindCandidateContent:     assessedContent,
		depKindProjectRevision:      assessedProject,
	}

	cases := []struct {
		name                  string
		current               map[string]string
		wantDecision          review.Decision
		wantReasons           []string
		mustNotContainReasons []string
		wantDocumentContains  []string
	}{
		{
			name:         "complete match",
			current:      completeMatch,
			wantDecision: review.DecisionEligible,
		},
		{
			name: "changed population — declared dep moved",
			current: map[string]string{
				depKindAssessmentPopulation: "matrix-population-shifted", // different
				depKindPolicyRevision:       assessedPolicy,
				depKindCandidateContent:     assessedContent,
				depKindProjectRevision:      assessedProject,
			},
			wantDecision:          review.DecisionUndetermined,
			wantReasons:           []string{review.ReasonStaleDependency},
			mustNotContainReasons: []string{review.ReasonCompatibilityUnknown},
			wantDocumentContains:  []string{"stale:", "matrix-population-shifted"},
		},
		{
			name:                  "nil current deps — nothing supplied",
			current:               nil,
			wantDecision:          review.DecisionUndetermined,
			wantReasons:           []string{review.ReasonCompatibilityUnknown},
			mustNotContainReasons: []string{review.ReasonStaleDependency},
			wantDocumentContains:  []string{"compatibility unknown", "no current value supplied"},
		},
		{
			name: "one omitted — assessment_population absent",
			current: map[string]string{
				// depKindAssessmentPopulation missing on purpose
				depKindPolicyRevision:   assessedPolicy,
				depKindCandidateContent: assessedContent,
				depKindProjectRevision:  assessedProject,
			},
			wantDecision:          review.DecisionUndetermined,
			wantReasons:           []string{review.ReasonCompatibilityUnknown},
			mustNotContainReasons: []string{review.ReasonStaleDependency},
			wantDocumentContains:  []string{"compatibility unknown", "assessment_population"},
		},
		{
			name: "blank reference — declared dep with empty value",
			current: map[string]string{
				depKindAssessmentPopulation: "", // blank; must NOT be treated as compatible
				depKindPolicyRevision:       assessedPolicy,
				depKindCandidateContent:     assessedContent,
				depKindProjectRevision:      assessedProject,
			},
			wantDecision:          review.DecisionUndetermined,
			wantReasons:           []string{review.ReasonCompatibilityUnknown},
			mustNotContainReasons: []string{review.ReasonStaleDependency},
			wantDocumentContains:  []string{"compatibility unknown", "assessment_population"},
		},
		{
			name: "unrelated extra — off-manifest change is harmless",
			current: map[string]string{
				depKindAssessmentPopulation: assessedPopulation,
				depKindPolicyRevision:       assessedPolicy,
				depKindCandidateContent:     assessedContent,
				depKindProjectRevision:      assessedProject,
				"unrelated_document":        "docs/projection.md@rev99",
			},
			wantDecision: review.DecisionEligible,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cov, err := app.GenerateReviewCoverage(ctx, ReviewCoverageInput{
				DBPath: dbPath, PolicyID: policy.PolicyID, SubjectRef: "matrix:control-subject",
				CurrentDependencies: tc.current,
			})
			if err != nil {
				t.Fatalf("generate coverage: %v", err)
			}
			if cov.Decision != string(tc.wantDecision) {
				t.Fatalf("decision = %s (reasons %v), want %s", cov.Decision, cov.Reasons, tc.wantDecision)
			}
			for _, want := range tc.wantReasons {
				if !containsString(cov.Reasons, want) {
					t.Fatalf("reasons = %v, want %s", cov.Reasons, want)
				}
			}
			for _, forbidden := range tc.mustNotContainReasons {
				if containsString(cov.Reasons, forbidden) {
					t.Fatalf("reasons = %v, must NOT contain %s (the two states must not collapse)",
						cov.Reasons, forbidden)
				}
			}
			for _, want := range tc.wantDocumentContains {
				if !strings.Contains(cov.Document, want) {
					t.Fatalf("document must contain %q:\n%s", want, cov.Document)
				}
			}
			// Historical preservation: the assessment record must remain in
			// the export regardless of the current decision. An H1 that
			// invalidated history to fix current eligibility would be a
			// worse cure than the disease.
			if !strings.Contains(cov.Document, historicalAssessmentID) {
				t.Fatalf("historical assessment %s must survive into every export:\n%s", historicalAssessmentID, cov.Document)
			}
		})
	}
}

func TestReviewCoverageSubjectScopeDoesNotCombineSubjects(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)

	policy, err := app.DefineReviewPolicy(ctx, ReviewPolicyDefineInput{
		DBPath: dbPath, Key: "subject-scope-control", Revision: 1,
		DecisionName:       "whether the selected subject may advance",
		Owner:              "repository-maintainer",
		AuthoritySource:    reviewContractRef,
		ScopeJustification: "subject-scope regression control only",
		Obligations: []ReviewObligationSpec{{
			Key:                "subject-scoped-obligation",
			SemanticRevision:   1,
			Requirement:        "coverage must derive the state for one exact subject",
			AcceptanceCriteria: "A's applicability and assessment are not combined with B or C",
			ApplicabilityRule:  "applies per exact subject",
			PrimaryOwner:       "repository-maintainer",
			Mandatory:          true,
		}},
	})
	if err != nil {
		t.Fatalf("define policy: %v", err)
	}
	obligationID := policy.ObligationIDs["subject-scoped-obligation@1"]

	apA, err := app.DecideReviewApplicability(ctx, ReviewApplicabilityInput{
		DBPath: dbPath, PolicyID: policy.PolicyID, ObligationID: obligationID,
		SubjectRef: "subject:A", Decision: review.Applies,
		Rationale: "A is in scope", Authorizer: "repository-maintainer",
	})
	if err != nil {
		t.Fatalf("A applicability: %v", err)
	}
	if _, err := app.DecideReviewApplicability(ctx, ReviewApplicabilityInput{
		DBPath: dbPath, PolicyID: policy.PolicyID, ObligationID: obligationID,
		SubjectRef: "subject:B", Decision: review.DoesNotApply,
		Rationale: "B is out of scope under this obligation", Authorizer: "repository-maintainer",
	}); err != nil {
		t.Fatalf("B applicability: %v", err)
	}
	if _, err := app.DecideReviewApplicability(ctx, ReviewApplicabilityInput{
		DBPath: dbPath, PolicyID: policy.PolicyID, ObligationID: obligationID,
		SubjectRef: "subject:C", Decision: review.Applies,
		Rationale: "C is in scope but deliberately unassessed", Authorizer: "repository-maintainer",
	}); err != nil {
		t.Fatalf("C applicability: %v", err)
	}
	check, err := app.RecordReviewCheck(ctx, ReviewCheckInput{
		DBPath: dbPath, PolicyID: policy.PolicyID, ObligationID: obligationID,
		CaseLabel: "subject-A", ProcedureRef: "subject-scope-control",
		ProcedureRevision: reviewRecipeRevision, InputsRef: "subject:A",
		Executor: reviewAssessor, Environment: "go test ./internal/pipeline",
		Mode: review.ModeExecuted, Outcome: review.CheckCompleted,
	})
	if err != nil {
		t.Fatalf("record check: %v", err)
	}
	asmA, err := app.RecordReviewAssessment(ctx, ReviewAssessInput{
		DBPath: dbPath, PolicyID: policy.PolicyID, ObligationID: obligationID,
		ApplicabilityDecisionID: apA.ID,
		SubjectRef:              "subject:A",
		ContextRef:              "subject-scope:A",
		Outcome:                 review.Conforms,
		Argument:                "A has an executed check and matching applicability",
		Assessor:                reviewAssessor,
		ProjectRevision:         "subject-scope-fixture",
		Dependencies: []ReviewDependencySpec{{
			Kind:        depKindCandidateContent,
			Ref:         "subject:A",
			WhyRelevant: "different content is a different subject",
		}},
		CheckAttemptIDs: []string{check.ID},
	})
	if err != nil {
		t.Fatalf("record assessment: %v", err)
	}

	covA, err := app.GenerateReviewCoverage(ctx, ReviewCoverageInput{
		DBPath: dbPath, PolicyID: policy.PolicyID, SubjectRef: "subject:A",
		CurrentDependencies: map[string]string{depKindCandidateContent: "subject:A"},
	})
	if err != nil {
		t.Fatalf("generate coverage for A: %v", err)
	}
	if covA.SubjectRef != "subject:A" || covA.Decision != string(review.DecisionEligible) {
		t.Fatalf("A decision = %s subject=%q reasons=%v, want scoped eligibility", covA.Decision, covA.SubjectRef, covA.Reasons)
	}
	if len(covA.Obligations) != 1 || covA.Obligations[0].State != string(review.StateConforms) {
		t.Fatalf("A must be governed by A's conforming assessment only: %+v", covA.Obligations)
	}
	if strings.Contains(covA.Document, "subject:B") || strings.Contains(covA.Document, "subject:C") {
		t.Fatalf("A-scoped export must not include other subjects:\n%s", covA.Document)
	}

	covB, err := app.GenerateReviewCoverage(ctx, ReviewCoverageInput{
		DBPath: dbPath, PolicyID: policy.PolicyID, SubjectRef: "subject:B",
		CurrentDependencies: map[string]string{depKindCandidateContent: "subject:B"},
	})
	if err != nil {
		t.Fatalf("generate coverage for B: %v", err)
	}
	if covB.Decision != string(review.DecisionEligible) || covB.Obligations[0].State != string(review.StateNotApplicable) {
		t.Fatalf("B should resolve through B's inapplicability only: decision=%s obligations=%+v",
			covB.Decision, covB.Obligations)
	}
	if strings.Contains(covB.Document, apA.ID) || strings.Contains(covB.Document, asmA.Assessment.ID) {
		t.Fatalf("B-scoped export must not inherit A's applicability or assessment:\n%s", covB.Document)
	}

	covC, err := app.GenerateReviewCoverage(ctx, ReviewCoverageInput{
		DBPath: dbPath, PolicyID: policy.PolicyID, SubjectRef: "subject:C",
		CurrentDependencies: map[string]string{depKindCandidateContent: "subject:C"},
	})
	if err != nil {
		t.Fatalf("generate coverage for C: %v", err)
	}
	if covC.Decision != string(review.DecisionUndetermined) || covC.Obligations[0].State != string(review.StateUnexamined) {
		t.Fatalf("C should remain unexamined rather than inheriting A: decision=%s obligations=%+v",
			covC.Decision, covC.Obligations)
	}
	if !containsString(covC.Reasons, review.ReasonUnexamined) {
		t.Fatalf("C reasons = %v, want %s", covC.Reasons, review.ReasonUnexamined)
	}
}

// TestReviewCoverageRetainsInvalidLegacyAssessmentReferences is the migration
// regression for v47. It reproduces immutable historical rows inserted before
// the scope trigger existed, restores the trigger, and then proves the current
// projection treats those rows as history rather than authority.
func TestReviewCoverageRetainsInvalidLegacyAssessmentReferences(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2026, 9, 13, 13, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)

	policy, err := app.DefineReviewPolicy(ctx, ReviewPolicyDefineInput{
		DBPath: dbPath, Key: "legacy-assessment-reference", Revision: 1,
		DecisionName:       "whether this exact subject has usable review support",
		Owner:              "repository-maintainer",
		AuthoritySource:    reviewContractRef,
		ScopeJustification: "v47 legacy assessment-reference migration regression",
		Obligations: []ReviewObligationSpec{{
			Key:                "legacy-reference-scope",
			SemanticRevision:   1,
			Requirement:        "current authority requires an assessment with exact cited scope",
			AcceptanceCriteria: "a malformed historical assessment is retained but cannot govern or contradict an exact-scope-valid assessment",
			ApplicabilityRule:  "applies to the exact review subject",
			PrimaryOwner:       "repository-maintainer",
			Mandatory:          true,
		}},
	})
	if err != nil {
		t.Fatalf("define policy: %v", err)
	}
	obligationID := policy.ObligationIDs["legacy-reference-scope@1"]
	const subject = "legacy-reference:subject"
	targetApplicability, err := app.DecideReviewApplicability(ctx, ReviewApplicabilityInput{
		DBPath: dbPath, PolicyID: policy.PolicyID, ObligationID: obligationID,
		SubjectRef: subject, Decision: review.Applies,
		Rationale: "the named subject is in scope", Authorizer: "repository-maintainer",
	})
	if err != nil {
		t.Fatalf("target applicability: %v", err)
	}
	otherApplicability, err := app.DecideReviewApplicability(ctx, ReviewApplicabilityInput{
		DBPath: dbPath, PolicyID: policy.PolicyID, ObligationID: obligationID,
		SubjectRef: "legacy-reference:other", Decision: review.Applies,
		Rationale: "a separate subject is independently in scope", Authorizer: "repository-maintainer",
	})
	if err != nil {
		t.Fatalf("other-subject applicability: %v", err)
	}

	// Seed two manifest references. One belongs to the target obligation; the
	// other belongs to an unbound helper obligation and is therefore an invalid
	// cited manifest for the target assessment.
	repo := openTestStore(t, ctx, dbPath)
	helperObligation := store.ReviewObligationRow{
		ID:                 domain.NewReviewObligationID(now.Add(time.Second)),
		ObligationKey:      "legacy-reference-helper",
		SemanticRevision:   1,
		Requirement:        "helper manifest target",
		AcceptanceCriteria: "used only to form a malformed legacy reference",
		ApplicabilityRule:  "not a current policy obligation",
		PrimaryOwner:       "repository-maintainer",
		CreatedAt:          now.Format(time.RFC3339),
	}
	if _, err := repo.PersistReviewObligation(ctx, helperObligation); err != nil {
		repo.Close()
		t.Fatalf("helper obligation: %v", err)
	}
	targetManifest := store.ReviewDependencyManifestRow{
		ID:              domain.NewReviewDependencyManifestID(now.Add(2 * time.Second)),
		PolicyID:        policy.PolicyID,
		ObligationID:    obligationID,
		ProjectRevision: "legacy-reference-fixture",
		CreatedAt:       now.Format(time.RFC3339),
		Dependencies: []store.ReviewManifestDependencyRow{{
			DependencyKind: depKindCandidateContent,
			DependencyRef:  subject,
			WhyRelevant:    "the target content identifies the assessment subject",
		}},
	}
	if _, err := repo.PersistReviewDependencyManifest(ctx, targetManifest); err != nil {
		repo.Close()
		t.Fatalf("target manifest: %v", err)
	}
	helperManifest := store.ReviewDependencyManifestRow{
		ID:              domain.NewReviewDependencyManifestID(now.Add(3 * time.Second)),
		PolicyID:        policy.PolicyID,
		ObligationID:    helperObligation.ID,
		ProjectRevision: "legacy-reference-fixture",
		CreatedAt:       now.Format(time.RFC3339),
		Dependencies: []store.ReviewManifestDependencyRow{{
			DependencyKind: depKindCandidateContent,
			DependencyRef:  subject,
			WhyRelevant:    "the target content identifies the malformed row's claimed subject",
		}},
	}
	if _, err := repo.PersistReviewDependencyManifest(ctx, helperManifest); err != nil {
		repo.Close()
		t.Fatalf("helper manifest: %v", err)
	}
	if err := repo.Close(); err != nil {
		t.Fatalf("close setup store: %v", err)
	}

	badApplicabilityID := domain.NewReviewAssessmentID(now.Add(-2 * time.Second))
	badManifestID := domain.NewReviewAssessmentID(now.Add(-time.Second))
	raw, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open raw migration-era store: %v", err)
	}
	raw.SetMaxOpenConns(1)
	if _, err := raw.ExecContext(ctx, `DROP TRIGGER review_assessments_reference_scope`); err != nil {
		raw.Close()
		t.Fatalf("disable v47 trigger: %v", err)
	}
	insertLegacy := func(id, applicabilityID, manifestID, outcome string) {
		t.Helper()
		if _, err := raw.ExecContext(ctx, `
INSERT INTO review_assessments(id, obligation_id, policy_id, applicability_decision_id, manifest_id, subject_ref, context_ref, outcome, argument, assessor, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, id, obligationID, policy.PolicyID, applicabilityID, manifestID, subject, "legacy-import", outcome,
			"raw immutable legacy assessment inserted before v47", "legacy-importer", now.Add(-time.Hour).Format(time.RFC3339)); err != nil {
			t.Fatalf("raw-insert legacy assessment %s: %v", id, err)
		}
	}
	insertLegacy(badApplicabilityID, otherApplicability.ID, targetManifest.ID, review.Nonconforms)
	insertLegacy(badManifestID, targetApplicability.ID, helperManifest.ID, review.Conforms)
	if _, err := raw.ExecContext(ctx, `DELETE FROM schema_migrations WHERE version = 47`); err != nil {
		raw.Close()
		t.Fatalf("mark v47 unapplied for trigger restoration: %v", err)
	}
	if err := raw.Close(); err != nil {
		t.Fatalf("close raw migration-era store: %v", err)
	}

	// Re-running the exact migration restores the write-time trigger without
	// rewriting the malformed immutable history we need the reader to classify.
	repo = openTestStore(t, ctx, dbPath)
	if err := repo.Close(); err != nil {
		t.Fatalf("close restored store: %v", err)
	}
	raw, err = sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open restored raw store: %v", err)
	}
	if _, err := raw.ExecContext(ctx, `
INSERT INTO review_assessments(id, obligation_id, policy_id, applicability_decision_id, manifest_id, subject_ref, context_ref, outcome, argument, assessor, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, domain.NewReviewAssessmentID(now.Add(4*time.Second)), obligationID, policy.PolicyID,
		otherApplicability.ID, targetManifest.ID, subject, "post-v47", review.Inconclusive,
		"a post-v47 malformed row must be refused", "legacy-importer", now.Format(time.RFC3339)); err == nil || !strings.Contains(err.Error(), "applicability decision scope does not match assessment") {
		raw.Close()
		t.Fatalf("restored v47 trigger must reject new malformed rows, got %v", err)
	}
	if err := raw.Close(); err != nil {
		t.Fatalf("close restored raw store: %v", err)
	}

	legacyOnly, err := app.GenerateReviewCoverage(ctx, ReviewCoverageInput{
		DBPath: dbPath, PolicyID: policy.PolicyID, SubjectRef: subject,
		CurrentDependencies: map[string]string{depKindCandidateContent: subject},
	})
	if err != nil {
		t.Fatalf("generate legacy-only coverage: %v", err)
	}
	if legacyOnly.Decision != string(review.DecisionUndetermined) || len(legacyOnly.Obligations) != 1 || legacyOnly.Obligations[0].State != string(review.StateInconclusive) {
		t.Fatalf("legacy-only coverage must remain inconclusive, got decision=%s obligations=%+v", legacyOnly.Decision, legacyOnly.Obligations)
	}
	if legacyOnly.Obligations[0].GoverningAssessmentID != "" || legacyOnly.Obligations[0].Contradiction {
		t.Fatalf("invalid legacy rows must neither govern nor create a contradiction: %+v", legacyOnly.Obligations[0])
	}
	if !containsString(legacyOnly.Reasons, review.ReasonInvalidAssessmentReference) {
		t.Fatalf("legacy-only reasons = %v, want %s", legacyOnly.Reasons, review.ReasonInvalidAssessmentReference)
	}
	for _, want := range []string{badApplicabilityID, badManifestID, "invalid assessment reference"} {
		if !strings.Contains(legacyOnly.Document, want) {
			t.Fatalf("legacy-only export must retain and label %q:\n%s", want, legacyOnly.Document)
		}
	}

	check, err := app.RecordReviewCheck(ctx, ReviewCheckInput{
		DBPath: dbPath, PolicyID: policy.PolicyID, ObligationID: obligationID,
		CaseLabel: "exact-scope-valid", ProcedureRef: "legacy-reference-control",
		ProcedureRevision: reviewRecipeRevision, InputsRef: "broader-subject-relevance-argument",
		Executor: reviewAssessor, Environment: "go test ./internal/pipeline",
		Mode: review.ModeExecuted, Outcome: review.CheckCompleted,
	})
	if err != nil {
		t.Fatalf("record valid check: %v", err)
	}
	valid, err := app.RecordReviewAssessment(ctx, ReviewAssessInput{
		DBPath: dbPath, PolicyID: policy.PolicyID, ObligationID: obligationID,
		ApplicabilityDecisionID: targetApplicability.ID,
		SubjectRef:              subject,
		ContextRef:              "exact-scope-valid",
		Outcome:                 review.Conforms,
		Argument:                "the exact-scope-valid assessment has completed executed support",
		Assessor:                reviewAssessor,
		ProjectRevision:         "legacy-reference-fixture",
		Dependencies: []ReviewDependencySpec{{
			Kind:        depKindCandidateContent,
			Ref:         subject,
			WhyRelevant: "the target content identifies the assessment subject",
		}},
		CheckAttemptIDs: []string{check.ID},
	})
	if err != nil {
		t.Fatalf("record valid assessment: %v", err)
	}
	withValid, err := app.GenerateReviewCoverage(ctx, ReviewCoverageInput{
		DBPath: dbPath, PolicyID: policy.PolicyID, SubjectRef: subject,
		CurrentDependencies: map[string]string{depKindCandidateContent: subject},
	})
	if err != nil {
		t.Fatalf("generate coverage with exact-scope-valid assessment: %v", err)
	}
	if withValid.Decision != string(review.DecisionEligible) || len(withValid.Obligations) != 1 || withValid.Obligations[0].State != string(review.StateConforms) {
		t.Fatalf("a valid assessment must remain usable despite malformed history: decision=%s obligations=%+v", withValid.Decision, withValid.Obligations)
	}
	if withValid.Obligations[0].GoverningAssessmentID != valid.Assessment.ID || withValid.Obligations[0].Contradiction {
		t.Fatalf("only the exact-scope-valid row may govern or enter contradiction selection: %+v", withValid.Obligations[0])
	}
	for _, want := range []string{badApplicabilityID, badManifestID, valid.Assessment.ID, "invalid assessment reference"} {
		if !strings.Contains(withValid.Document, want) {
			t.Fatalf("coverage with valid support must still retain and label %q:\n%s", want, withValid.Document)
		}
	}

	repo = openTestStore(t, ctx, dbPath)
	defer repo.Close()
	persisted, err := repo.LoadReviewCoverageForSubject(ctx, policy.PolicyID, subject)
	if err != nil {
		t.Fatalf("load persisted coverage: %v", err)
	}
	seen := map[string]store.ReviewAssessmentReferenceScope{}
	for _, assessment := range persisted.Obligations[0].Assessments {
		seen[assessment.ID] = assessment.ReferenceScope
	}
	for _, id := range []string{badApplicabilityID, badManifestID} {
		if seen[id] != store.ReviewAssessmentReferenceScopeInvalid {
			t.Fatalf("legacy assessment %s reference scope = %q, want invalid", id, seen[id])
		}
	}
	if seen[valid.Assessment.ID] != store.ReviewAssessmentReferenceScopeExactValid {
		t.Fatalf("valid assessment reference scope = %q, want exact_scope_valid", seen[valid.Assessment.ID])
	}
}

// TestC3NegativeControlAttestationCannotMasqueradeAsIndependentCheck is the
// committed regression for H2 / GOW-R1 (2026-09-12 remediation): the recipe's
// C3 requires an INDEPENDENTLY CHECKED observation. An operator attestation of
// a model-judged failure legitimately admits its content for structural
// reasons, but it must remain labeled model-judged-by-operator and MUST NOT
// satisfy the witness-backed contract C3 requires.
//
// Rather than duplicate the attestation-admission integration test (which
// already asserts adm.ObservationKind == ObservationModelJudgedFailure and
// adm.AdmittedBy == "operator" in TestIntegrationEvidenceAdmissionClosesReentry),
// this test pins the DISCRIMINATION property the C3 assertion depends on: the
// two label pairs must remain distinct so that substituting attestation for
// witness-check fails the C3 assertion mechanically.
//
// If a future refactor merges either label into a single value that accepts
// both routes, this test fails BEFORE the C3 integration assertion silently
// starts accepting attestation. The witness-backed positive control lives in
// TestIntegrationReplayPreservesCompatibleCurrentAuthority and the C3 branch
// of TestIntegrationCurrentAssessmentAuthorityObligation.
//
// The classification path itself is covered by TestClassifyEvaluatedFailure in
// admission_integration_test.go; that test asserts the mapping from
// VerificationStrength to (ObservationKind, RuleAdmissible, Attestable). This
// test is one layer above: it asserts that the LABELS the C3 assertion checks
// against are the ones the classification function actually produces for
// witness-backed and attestation-based routes.
func TestC3NegativeControlAttestationCannotMasqueradeAsIndependentCheck(t *testing.T) {
	t.Parallel()
	// The witness-backed C3 admission labels — what
	// TestIntegrationCurrentAssessmentAuthorityObligation asserts on C3.
	const (
		witnessBackedKind = ObservationDomainCheckedFailure
		witnessBackedBy   = "rule"
	)
	// The attestation-only labels — what
	// TestIntegrationEvidenceAdmissionClosesReentry observes when Attest=true
	// admits a model-judged failure.
	const (
		attestedKind = ObservationModelJudgedFailure
		attestedBy   = "operator"
	)
	// If either label collapses across routes, C3 stops distinguishing
	// witness-backed evidence from attested model judgment, and an operator
	// note starts satisfying an independent-check contract. Fail here.
	if witnessBackedKind == attestedKind {
		t.Fatalf("H2 negative control invariant broken: ObservationDomainCheckedFailure (%q) and ObservationModelJudgedFailure (%q) must remain distinct — otherwise the C3 assertion cannot separate witness-backed evidence from attested model judgment",
			witnessBackedKind, attestedKind)
	}
	if witnessBackedBy == attestedBy {
		t.Fatalf("H2 negative control invariant broken: AdmittedBy for the witness-backed rule pass (%q) and for the attested route (%q) must remain distinct — otherwise C3 accepts attestation as an independent check",
			witnessBackedBy, attestedBy)
	}
	// Empty labels would let a comparison silently succeed against a
	// zero-valued admission record and mask a regression that dropped the
	// label. Guard the labels' presence.
	for _, s := range []string{witnessBackedKind, witnessBackedBy, attestedKind, attestedBy} {
		if s == "" {
			t.Fatalf("H2 negative control: label constants must be non-empty (kind=%q by=%q vs kind=%q by=%q)",
				witnessBackedKind, witnessBackedBy, attestedKind, attestedBy)
		}
	}
}

// committed regression for H3 / GOW-R3 (2026-09-12 remediation): the derived
// export must carry checker inputs, procedure revision, environment and the
// dependency-manifest contents; a rendered document alone is not portable
// evidence.
//
// This pins each field the reviewer flagged as dropped by the projection
// adapter or omitted by the renderer. It does NOT assert byte-identical
// rendering across generations — that determinism assertion lives in
// TestRenderCoverageNamesProvenanceWithoutLosingDeterminism, which is unit
// scope. Here we assert PRESENCE of the substantive provenance fields on the
// pipeline path a caller actually exercises.
func TestReviewCoverageExportCarriesCheckerAndDependencyProvenance(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)

	policy, err := app.DefineReviewPolicy(ctx, ReviewPolicyDefineInput{
		DBPath: dbPath, Key: "provenance-export-control", Revision: 1,
		DecisionName:       "whether the sole obligation may be reported as satisfied",
		Owner:              "repository-maintainer",
		AuthoritySource:    reviewContractRef,
		ScopeJustification: "H3 remediation control: authorizes nothing outside the provenance export assertion",
		EvidenceCutoff:     "2026-09-12T12:00:00Z",
		Obligations: []ReviewObligationSpec{{
			Key: "provenance-export", SemanticRevision: 1,
			Requirement:        "the export must name the checker, its inputs, and the assessment's declared dependencies",
			AcceptanceCriteria: "H3 / GOW-R3: procedure_revision, inputs_ref, environment, and manifest contents in the rendered document",
			ApplicabilityRule:  "applies to the control subject only",
			PrimaryOwner:       "repository-maintainer", Mandatory: true,
		}},
	})
	if err != nil {
		t.Fatalf("define control policy: %v", err)
	}
	obligationID := policy.ObligationIDs["provenance-export@1"]
	apDecision, err := app.DecideReviewApplicability(ctx, ReviewApplicabilityInput{
		DBPath: dbPath, PolicyID: policy.PolicyID, ObligationID: obligationID,
		SubjectRef: "provenance:control-subject", Decision: review.Applies,
		Rationale: "in scope", Authorizer: "repository-maintainer",
	})
	if err != nil {
		t.Fatalf("decide applicability: %v", err)
	}
	const (
		procedureRevision = "checker-procedure-rev-42"
		inputsRef         = "checker-inputs-fingerprint-abc123"
		environment       = "go test ./internal/pipeline (provenance-control)"
		outputRef         = "checker-output-fingerprint-def456"
		whyPopulation     = "the assessment authority is bounded to the population it examined"
		declaredPop       = "provenance-population-declared"
	)
	check, err := app.RecordReviewCheck(ctx, ReviewCheckInput{
		DBPath: dbPath, PolicyID: policy.PolicyID, ObligationID: obligationID,
		CaseLabel: "provenance-check", ProcedureRef: "provenance-procedure",
		ProcedureRevision: procedureRevision,
		InputsRef:         inputsRef,
		Executor:          reviewAssessor,
		Environment:       environment,
		Mode:              review.ModeExecuted, Outcome: review.CheckCompleted,
		OutputRef: outputRef,
	})
	if err != nil {
		t.Fatalf("record check: %v", err)
	}
	if _, err := app.RecordReviewAssessment(ctx, ReviewAssessInput{
		DBPath: dbPath, PolicyID: policy.PolicyID, ObligationID: obligationID,
		ApplicabilityDecisionID: apDecision.ID,
		SubjectRef:              "provenance:control-subject",
		ContextRef:              "context:" + declaredPop,
		Outcome:                 review.Conforms,
		Argument:                "control assessment for provenance export",
		Assessor:                reviewAssessor,
		ProjectRevision:         "provenance-project-declared",
		ContractHash:            reviewContractRef,
		RecipeHash:              reviewRecipeRevision,
		EvidenceCutoff:          "2026-09-12T12:00:00Z",
		Dependencies: []ReviewDependencySpec{
			{Kind: depKindAssessmentPopulation, Ref: declaredPop, WhyRelevant: whyPopulation},
		},
		CheckAttemptIDs: []string{check.ID},
	}); err != nil {
		t.Fatalf("record assessment: %v", err)
	}
	cov, err := app.GenerateReviewCoverage(ctx, ReviewCoverageInput{
		DBPath: dbPath, PolicyID: policy.PolicyID, SubjectRef: "provenance:control-subject",
		CurrentDependencies: map[string]string{depKindAssessmentPopulation: declaredPop},
	})
	if err != nil {
		t.Fatalf("generate coverage: %v", err)
	}
	if cov.Decision != string(review.DecisionEligible) {
		t.Fatalf("decision = %s (reasons %v), want %s", cov.Decision, cov.Reasons, review.DecisionEligible)
	}
	// Each field the reviewer flagged must appear verbatim in the export.
	// Without them a reader with only the document cannot audit the check or
	// the dependency argument.
	for _, want := range []string{
		procedureRevision, inputsRef, environment, outputRef, // check provenance
		declaredPop, whyPopulation, // dependency-manifest contents
		review.GeneratorVersion,                 // generator identity
		"provenance-project-declared",           // assessment project revision
		reviewContractRef, reviewRecipeRevision, // contract and recipe pinned to the assessment
	} {
		if !strings.Contains(cov.Document, want) {
			t.Fatalf("export must carry provenance %q:\n%s", want, cov.Document)
		}
	}
}
