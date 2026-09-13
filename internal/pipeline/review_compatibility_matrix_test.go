package pipeline

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/review"
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
				DBPath: dbPath, PolicyID: policy.PolicyID, CurrentDependencies: tc.current,
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
		DBPath: dbPath, PolicyID: policy.PolicyID,
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
