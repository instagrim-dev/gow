package pipeline

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/provider"
	"github.com/instagrim-dev/newf/internal/review"
	"github.com/instagrim-dev/newf/internal/verify"
)

// gateLocalSignature is pcSignature with LOCAL posture: it PRESERVES the mined
// claim `locality == local`, so admitting its member does not weaken the
// candidate and C6's counterexample search over A1 can still reach `surviving`.
func gateLocalSignature(preservesID string) canon.MechanismSignature {
	sig := pcSignature(preservesID, "", true)
	sig.Posture.Locality = domain.LocalityLocal
	return sig
}

// gateUnknownLocalitySignature is pcSignature with UNKNOWN locality, which is
// what C2's withholding control requires: the structural break claim against
// `locality == local` is UNDECIDABLE, so the evaluation escalates to the model
// tier and produces the single-model-judgment failure that must be withheld. A
// LOCAL-posture proposal would instead be refuted deterministically and never
// reach the model tier at all.
func gateUnknownLocalitySignature(preservesID string) canon.MechanismSignature {
	sig := pcSignature(preservesID, "", true)
	sig.Posture.Locality = domain.LocalityUnknown
	return sig
}

// TestIntegrationCurrentAssessmentAuthorityObligation is the integrated gate G1
// of the 2026-09-12 review-flow run: ONE normative obligation
// (`current-assessment-authority@1`), instantiated through the real migrated
// store with the contract's four record responsibilities, exercised across cases
// C1-C8, with coverage GENERATED from those records.
//
// The earlier run could not establish this gate because the mapping did not
// exist: there was nowhere to put a policy, an applicability decision, a check
// attempt or an assessment, so coverage could only have been hand-written. This
// test is the thing whose absence made the run's disposition UNDETERMINED.
//
// What each case must observably do:
//
//	C1 baseline    exact record ids explain the selected action
//	C2 withheld    a model-only failure stays auditable but out of the population
//	C3 admission   only INDEPENDENTLY CHECKED exact content enters A1
//	C4 relevant    a declared-dependency change makes the assessment stale;
//	               eligibility is NOT inherited from A0
//	C5 unrelated   an off-manifest change does NOT make it stale
//	C6 reassess    a distinct assessment drives the next decision; A0 reproducible
//	C7 replay      history replays without restoring obsolete current authority,
//	               without displacing COMPATIBLE current authority, and without
//	               depending on clock order to tell those apart
//	C8 projection  coverage is deterministic; unexamined/blocked stay distinct
//
// The obligation is a NORM. The mined candidate is its SUBJECT. Nothing here
// promotes the candidate's lifecycle state into a normative conclusion, and
// nothing turns the norm into a CandidateInvariant.
func TestIntegrationCurrentAssessmentAuthorityObligation(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = dataDrivenMiner{}
	app.challengerFn = biasOnlyChallenger{}
	app.modelVerifierFn = provider.NewFixtureModelVerifier(verify.VerdictFailure, "high")

	problemID, invID, _ := mineOneCandidate(t, ctx, app, dbPath)
	subjectRef := "invariant:" + invID
	ledger := instantiateReviewObligation(t, ctx, app, dbPath, subjectRef)
	// ---- C1: baseline. Assess the candidate under P1 and the discovery
	// population D0, then derive the decision from records only.
	first, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID})
	if err != nil {
		t.Fatalf("baseline challenge: %v", err)
	}
	baseline := first.Reports[0]
	if baseline.StateAfter != "surviving" {
		t.Fatalf("baseline campaign must earn surviving, got %q", baseline.StateAfter)
	}
	d0 := baseline.AssessmentClusterRunID
	if d0 == "" {
		t.Fatal("baseline campaign must record its assessment population")
	}

	c1 := ledger.recordCase(t, ctx, app, dbPath, "C1", "app.ChallengeInvariants + app.GenerateFrontier",
		"invariant="+invID+"; population_policy="+baseline.PopulationPolicy,
		review.CheckCompleted, "challenge_run="+baseline.RunID+"; assessment_cluster_run="+d0, "")

	// The baseline generation is the actual decision this policy names: may the
	// candidate guide the next search action? Record what really happened.
	genBase, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("baseline generate: %v", err)
	}
	if len(genBase.ExcludedStaleAuthority) != 0 {
		t.Fatalf("a candidate assessed against the current population must be targetable: %+v", genBase.ExcludedStaleAuthority)
	}

	a0, err := app.RecordReviewAssessment(ctx, ReviewAssessInput{
		DBPath: dbPath, PolicyID: ledger.policyID, ObligationID: ledger.obligationID,
		ApplicabilityDecisionID: ledger.applicability,
		SubjectRef:              subjectRef,
		ContextRef:              "cluster_run:" + d0 + "; generation:" + genBase.Generation.ID,
		Outcome:                 review.Conforms,
		Argument: "the campaign that produced the current state assessed cluster run " + d0 +
			", which is the current compatible population; the generation therefore selected the candidate " +
			"without a stale-authority exclusion",
		Assessor:        reviewAssessor,
		ProjectRevision: gateProjectRevision,
		ContractHash:    reviewContractRef,
		RecipeHash:      reviewRecipeRevision,
		EvidenceCutoff:  "2026-09-12T12:00:00Z",
		Dependencies: []ReviewDependencySpec{
			{Kind: depKindAssessmentPopulation, Ref: d0,
				WhyRelevant: "the obligation is about compatibility between the decision and its evidence population; a different current population changes what the decision may rely on"},
			{Kind: depKindPolicyRevision, Ref: reviewPolicyKey + "@1",
				WhyRelevant: "a semantically relevant policy revision changes the applicable acceptance criteria"},
			{Kind: depKindCandidateContent, Ref: invID,
				WhyRelevant: "the assessment is about this exact candidate; different content is a different subject"},
			{Kind: depKindProjectRevision, Ref: gateProjectRevision,
				WhyRelevant: "the obligation is about behavior of the authority-selection code; a different checkout can decide the next action differently"},
		},
		CheckAttemptIDs: []string{c1},
	})
	if err != nil {
		t.Fatalf("record baseline assessment: %v", err)
	}

	// Coverage is DERIVED. Under the baseline population the mandatory
	// obligation has current, check-supported conformance.
	cov1, err := app.GenerateReviewCoverage(ctx, ReviewCoverageInput{
		DBPath: dbPath, PolicyID: ledger.policyID, SubjectRef: subjectRef,
		CurrentDependencies: gateCurrentContext(d0, invID, gateProjectRevision),
	})
	if err != nil {
		t.Fatalf("generate coverage (C1): %v", err)
	}
	if cov1.Decision != string(review.DecisionEligible) {
		t.Fatalf("C1 decision = %s (reasons %v), want %s", cov1.Decision, cov1.Reasons, review.DecisionEligible)
	}
	// C1's requirement: EXACT ids explain the selected action — no global
	// artifact-result shortcut.
	for _, want := range []string{ledger.policyID, ledger.obligationID, ledger.applicability, a0.Assessment.ID, c1, d0} {
		if !strings.Contains(cov1.Document, want) {
			t.Fatalf("C1 coverage must name the exact record id %s:\n%s", want, cov1.Document)
		}
	}

	// ---- C2: withheld control. A model-only failure is recorded and remains
	// auditable, but it must not enter the policy's required observed population
	// and must not change this decision's evidence basis.
	if len(genBase.Generation.Proposals) == 0 {
		t.Fatal("expected a proposal to evaluate")
	}
	evalRes, err := app.Evaluate(ctx, EvaluateInput{DBPath: dbPath, ProblemID: problemID, ProposalID: genBase.Generation.Proposals[0].ID})
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	ev := evalRes.Run.Evaluations[0]
	if ev.VerificationStrength != "single-model-judgment" {
		t.Fatalf("C2 needs a model-judged failure, got %q", ev.VerificationStrength)
	}
	popBefore, err := app.BuildClustering(ctx, ClusterBuildInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("cluster build before withholding: %v", err)
	}
	withheld, err := app.AdmitEvidence(ctx, AdmitEvidenceInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("admit (rule pass): %v", err)
	}
	if len(withheld.Admitted) != 0 || len(withheld.Withheld) != 1 {
		t.Fatalf("C2: the rule pass must withhold model judgment: %+v", withheld)
	}
	popAfterWithhold, err := app.BuildClustering(ctx, ClusterBuildInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("cluster build after withholding: %v", err)
	}
	if popAfterWithhold.ClusterRun.SignatureCount != popBefore.ClusterRun.SignatureCount {
		t.Fatalf("C2: a withheld observation must not enter the population: %d -> %d",
			popBefore.ClusterRun.SignatureCount, popAfterWithhold.ClusterRun.SignatureCount)
	}
	c2 := ledger.recordCase(t, ctx, app, dbPath, "C2", "app.Evaluate + app.AdmitEvidence (rule pass)",
		"evaluation="+ev.ID, review.CheckCompleted,
		"withheld_basis="+withheld.Withheld[0].Basis+"; population_unchanged="+popBefore.ClusterRun.ID, "")

	// The withheld observation changed nothing about C1's evidence basis: same
	// records, same derived decision.
	cov2, err := app.GenerateReviewCoverage(ctx, ReviewCoverageInput{
		DBPath: dbPath, PolicyID: ledger.policyID, SubjectRef: subjectRef,
		CurrentDependencies: gateCurrentContext(d0, invID, gateProjectRevision),
	})
	if err != nil {
		t.Fatalf("generate coverage (C2): %v", err)
	}
	if cov2.Decision != cov1.Decision {
		t.Fatalf("C2: a withheld model judgment must not change the derived decision: %s -> %s", cov1.Decision, cov2.Decision)
	}

	// ---- C3: INDEPENDENTLY CHECKED admission. The recipe requires a checked,
	// scope-matched, candidate-specific observation whose checker inputs and
	// outputs are retained — and explicitly forbids substituting an
	// operator-attested model judgment when the adapter cannot reach it.
	//
	// So the observation admitted here is produced by an EXECUTED bounded
	// attempt: `equal-denominator` at n=7 emits the tuple that IS the checked
	// claim, and admission re-derives the procedure over the recorded params
	// before naming the binding in its basis. A fixture checker still only
	// exercises the software contract — it establishes no research claim — but
	// the attribution from attempt to output is machine-rechecked rather than
	// asserted.
	//
	// The proposal minted here carries LOCAL posture, required by a LATER case
	// rather than this one: an unknown-locality member in A1 leaves C6's
	// counterexample search inconclusive (`challenged` instead of `surviving`),
	// as run-3 attempt 3.2 demonstrated. C2's proposal stays on the default
	// deriving generator, whose structural claim is undecidable and therefore
	// escalates to the model tier that the withholding control needs.
	app.generatorFn = pcGenerator{signatures: []canon.MechanismSignature{
		gateLocalSignature(pcResidueLocality),
	}}
	genMint, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil || len(genMint.Generation.Proposals) == 0 {
		t.Fatalf("C3 mint proposal: %v (proposals=%d)", err, len(genMint.Generation.Proposals))
	}
	app.generatorFn = nil
	c3Proposal := genMint.Generation.Proposals[0].ID

	wit, err := app.WitnessCheck(ctx, WitnessCheckInput{
		DBPath: dbPath, ProposalID: c3Proposal,
		Procedure: "equal-denominator", Params: map[string]string{"n": "7"},
	})
	if err != nil {
		t.Fatalf("C3 witness check: %v", err)
	}
	if wit.WitnessVerdict != "witness-invalid" {
		t.Fatalf("C3: equal-denominator at n=7 must fail its own check, got %q", wit.WitnessVerdict)
	}
	if wit.AttemptBinding == nil || wit.AttemptBinding.TupleCanonical != wit.CanonicalClaim {
		t.Fatalf("C3: the bound tuple must BE the checked claim: %+v vs %q", wit.AttemptBinding, wit.CanonicalClaim)
	}
	admitted, err := app.AdmitEvidence(ctx, AdmitEvidenceInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("admit (checked): %v", err)
	}
	if len(admitted.Admitted) != 1 {
		t.Fatalf("C3: the rule pass must admit exactly the checked observation: %+v", admitted)
	}
	adm := admitted.Admitted[0]
	// The recipe's distinction: this is a DOMAIN-CHECKED failure admitted by
	// RULE, not a model judgment admitted by an operator. Asserting
	// `AdmittedBy != "operator"` is what keeps the forbidden substitution from
	// silently returning.
	if adm.ObservationKind != ObservationDomainCheckedFailure || adm.AdmittedBy != "rule" {
		t.Fatalf("C3: admission must be an independently checked domain observation admitted by rule, got kind=%q by=%q",
			adm.ObservationKind, adm.AdmittedBy)
	}
	if adm.EvaluationID != wit.Evaluation.ID || adm.ProposalID != c3Proposal {
		t.Fatalf("C3: admission must reference the witness evaluation of the bound proposal: %+v", adm)
	}
	if !strings.Contains(adm.Basis, "attempt→output binding verified by recomputation: equal-denominator@") {
		t.Fatalf("C3: the admission basis must name the recomputation-verified binding: %q", adm.Basis)
	}
	if adm.ContentHash == "" || adm.SignatureID == "" {
		t.Fatalf("C3: admission must materialize the exact assessed content: %+v", adm)
	}
	a1Run, err := app.BuildClustering(ctx, ClusterBuildInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("cluster build after admission: %v", err)
	}
	a1 := a1Run.ClusterRun.ID
	if a1 == d0 {
		t.Fatal("C3: the admitted observation must produce a NEW population A1")
	}
	beforeMembers := clusterSignatureSet(popBefore.ClusterRun)
	if beforeMembers[adm.SignatureID] {
		t.Fatalf("C3: admitted signature %s was already in the population before admission", adm.SignatureID)
	}
	wantMembers := copyStringSet(beforeMembers)
	wantMembers[adm.SignatureID] = true
	assertStringSetEqual(t, clusterSignatureSet(a1Run.ClusterRun), wantMembers)
	c3 := ledger.recordCase(t, ctx, app, dbPath, "C3",
		"app.WitnessCheck(--procedure equal-denominator) + app.AdmitEvidence (rule) + app.BuildClustering",
		"proposal="+c3Proposal+"; procedure=equal-denominator; params=n=7", review.CheckCompleted,
		"admitted_signature="+adm.SignatureID+"; content_hash="+adm.ContentHash+
			"; observation_kind="+adm.ObservationKind+"; admitted_by="+adm.AdmittedBy+
			"; attempt_binding_verified_by_recomputation=true; population_a1="+a1, "")

	// ---- C4: relevant change. A1 is now current while the candidate is
	// unchanged. The assessment declared the population as a why-relevant
	// dependency, so it is stale for a CURRENT request — and eligibility is not
	// inherited from A0.
	cov4, err := app.GenerateReviewCoverage(ctx, ReviewCoverageInput{
		DBPath: dbPath, PolicyID: ledger.policyID, SubjectRef: subjectRef,
		CurrentDependencies: gateCurrentContext(a1, invID, gateProjectRevision),
	})
	if err != nil {
		t.Fatalf("generate coverage (C4): %v", err)
	}
	if cov4.Decision != string(review.DecisionUndetermined) {
		t.Fatalf("C4 decision = %s, want %s (eligibility must not be inherited across a relevant population change)",
			cov4.Decision, review.DecisionUndetermined)
	}
	if !containsString(cov4.Reasons, review.ReasonStaleDependency) {
		t.Fatalf("C4 reasons = %v, want %s", cov4.Reasons, review.ReasonStaleDependency)
	}
	if !strings.Contains(cov4.Document, "stale:") || !strings.Contains(cov4.Document, a1) {
		t.Fatalf("C4 coverage must state which declared dependency moved and to what:\n%s", cov4.Document)
	}

	// ---- C5: unrelated control. An artifact OUTSIDE the declared dependency
	// graph changes. A still-compatible assessment must not become stale merely
	// because something moved in the repository.
	cov5, err := app.GenerateReviewCoverage(ctx, ReviewCoverageInput{
		DBPath: dbPath, PolicyID: ledger.policyID, SubjectRef: subjectRef,
		CurrentDependencies: withUnrelatedChange(gateCurrentContext(d0, invID, gateProjectRevision),
			// Declared by nobody in this manifest: an unrelated document moved.
			"unrelated_document", "docs/projection.md@rev99"),
	})
	if err != nil {
		t.Fatalf("generate coverage (C5): %v", err)
	}
	if cov5.Decision != string(review.DecisionEligible) {
		t.Fatalf("C5 decision = %s (reasons %v), want %s: an off-manifest change is not staleness",
			cov5.Decision, cov5.Reasons, review.DecisionEligible)
	}

	// ---- C6: reassessment. The unchanged candidate is assessed against A1 under
	// the pinned policy. A DISTINCT assessment drives the next decision, and A0
	// stays reproducible — newly admitted evidence cannot rewrite the bounded
	// historical claim.
	reassess, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID, Population: PopulationLatest})
	if err != nil {
		t.Fatalf("C6 reassessment: %v", err)
	}
	rr := reassess.Reports[0]
	if rr.RunID == baseline.RunID {
		t.Fatal("C6: reassessment must be a distinct campaign, not a mutation of A0's")
	}
	if rr.AssessmentClusterRunID != a1 {
		t.Fatalf("C6: reassessment must assess A1 (%s), got %s", a1, rr.AssessmentClusterRunID)
	}
	if rr.DiscoveryClusterRunID != d0 {
		t.Fatalf("C6: the discovery population must remain %s, got %s", d0, rr.DiscoveryClusterRunID)
	}
	genC6, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("C6 generate: %v", err)
	}
	c6 := ledger.recordCase(t, ctx, app, dbPath, "C6", "app.ChallengeInvariants(population=latest) + app.GenerateFrontier",
		"invariant="+invID+"; population="+a1, review.CheckCompleted,
		"challenge_run="+rr.RunID+"; state_after="+rr.StateAfter+"; generation="+genC6.Generation.ID, "")

	a1Assessment, err := app.RecordReviewAssessment(ctx, ReviewAssessInput{
		DBPath: dbPath, PolicyID: ledger.policyID, ObligationID: ledger.obligationID,
		ApplicabilityDecisionID: ledger.applicability,
		SubjectRef:              subjectRef,
		ContextRef:              "cluster_run:" + a1 + "; generation:" + genC6.Generation.ID,
		Outcome:                 review.Conforms,
		Argument: "reassessment under the current population " + a1 + " produced campaign " + rr.RunID +
			", so the state that governs the next decision was earned against the population that decision uses; " +
			"the A0 assessment is retained unchanged as history",
		Assessor:        reviewAssessor,
		ProjectRevision: gateProjectRevision,
		ContractHash:    reviewContractRef,
		RecipeHash:      reviewRecipeRevision,
		EvidenceCutoff:  "2026-09-12T12:00:00Z",
		Dependencies: []ReviewDependencySpec{
			{Kind: depKindAssessmentPopulation, Ref: a1,
				WhyRelevant: "this assessment's authority is bounded to the population it examined"},
			{Kind: depKindPolicyRevision, Ref: reviewPolicyKey + "@1",
				WhyRelevant: "a semantically relevant policy revision changes the applicable acceptance criteria"},
			{Kind: depKindCandidateContent, Ref: invID,
				WhyRelevant: "the assessment is about this exact candidate"},
			{Kind: depKindProjectRevision, Ref: gateProjectRevision,
				WhyRelevant: "the obligation is about behavior of the authority-selection code; a different checkout can decide the next action differently"},
		},
		CheckAttemptIDs: []string{c6},
	})
	if err != nil {
		t.Fatalf("C6 record assessment: %v", err)
	}
	if a1Assessment.Assessment.ID == a0.Assessment.ID {
		t.Fatal("C6: the reassessment must be a distinct assessment record")
	}

	// Under A1 the mandatory obligation is supported again — and the path back
	// was reassessment, not inheritance. The C4 gate is therefore not a dead end.
	cov6, err := app.GenerateReviewCoverage(ctx, ReviewCoverageInput{
		DBPath: dbPath, PolicyID: ledger.policyID, SubjectRef: subjectRef,
		CurrentDependencies: gateCurrentContext(a1, invID, gateProjectRevision),
	})
	if err != nil {
		t.Fatalf("generate coverage (C6): %v", err)
	}
	if cov6.Decision != string(review.DecisionEligible) {
		t.Fatalf("C6 decision = %s (reasons %v), want %s: reassessment must restore current support",
			cov6.Decision, cov6.Reasons, review.DecisionEligible)
	}
	if len(cov6.Obligations) != 1 || cov6.Obligations[0].GoverningAssessmentID != a1Assessment.Assessment.ID {
		t.Fatalf("C6: the A1 assessment must govern the current decision: %+v", cov6.Obligations)
	}
	// A0 is still in the document verbatim: history is retained, not replaced.
	if !strings.Contains(cov6.Document, a0.Assessment.ID) {
		t.Fatalf("C6: the superseded A0 assessment must remain in the record export:\n%s", cov6.Document)
	}

	// ---- C7: historical replay. Replaying the discovery population reproduces
	// the historical result WITHOUT restoring current authority. This is the
	// exact hole the earlier run reported: a replay could re-earn survival
	// against an obsolete population.
	replay, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID, Population: PopulationDiscovery})
	if err != nil {
		t.Fatalf("C7 replay: %v", err)
	}
	pr := replay.Reports[0]
	if pr.AssessmentClusterRunID != d0 {
		t.Fatalf("C7: replay must assess the historical population %s, got %s", d0, pr.AssessmentClusterRunID)
	}
	if pr.StateAfter != baseline.StateAfter {
		t.Fatalf("C7: the historical result must be reproducible (%q), got %q", baseline.StateAfter, pr.StateAfter)
	}
	genC7, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("C7 generate: %v", err)
	}
	// F-2 (2026-09-12 review run 2, remediation handoff 1): the next CURRENT
	// decision selects the compatible current context, not the last execution.
	// The A1 campaign (C6) assessed the CURRENT population and earned this
	// exact targetable state, so it remains the operative authority: the
	// replay reproduces its bounded historical claim WITHOUT displacing that
	// authority, and the candidate stays selectable. (The negative control —
	// a replay whose only surviving authority is obsolete — lives in
	// TestIntegrationChallengeAssessmentPopulation 5b, where the compatible
	// campaign earned `weaken`, a different state, and exclusion stands.)
	for _, e := range genC7.ExcludedStaleAuthority {
		if e.InvariantID == invID {
			t.Fatalf("C7: replay must not displace compatible current authority (F-2): %+v", e)
		}
	}
	repoC7 := openTestStore(t, ctx, dbPath)
	genRow, err := repoC7.GetFrontierGeneration(ctx, genC7.Generation.ID)
	if err != nil {
		t.Fatalf("C7: load post-replay generation: %v", err)
	}
	invocs, err := repoC7.ListProviderInvocationsForRun(ctx, genRow.RunID)
	repoC7.Close()
	if err != nil || len(invocs) == 0 {
		t.Fatalf("C7: load post-replay invocation: %v (n=%d)", err, len(invocs))
	}
	var request provider.GenerationRequest
	if err := json.Unmarshal([]byte(invocs[0].RequestPayload), &request); err != nil {
		t.Fatalf("C7: decode post-replay generation request: %v\n%s", err, invocs[0].RequestPayload)
	}
	if !generationTargetsContain(request.Targets, invID) {
		t.Fatalf("C7: the post-replay generation request targets %v, want %s via the compatible current authority",
			request.Targets, invID)
	}
	c7 := ledger.recordCase(t, ctx, app, dbPath, "C7", "app.ChallengeInvariants(population=discovery) + app.GenerateFrontier",
		"invariant="+invID+"; population="+d0, review.CheckCompleted,
		"replay_run="+pr.RunID+"; replay_state="+pr.StateAfter+"; compatible_current_authority_preserved=true; historical_row_durable=true", "")

	// ---- C7 stress: equal clock, order-dependent outcome.
	//
	// Every row above was written by a FIXED clock, so C6's reassessment and
	// C7's replay carry identical `created_at` values. That makes this scenario
	// the equal-clock case the recipe asks for — but only if the assertion is
	// made explicitly, because an implementation that silently ordered by
	// timestamp would pass everything above by luck of insertion order while
	// being one `ORDER BY created_at` away from resolving a tie arbitrarily.
	//
	// The requirement: authority selection must be decided by a total order the
	// store controls (`transition_seq`), not by a wall clock that can tie. So
	// with A1's reassessment and the D0 replay recorded at the SAME instant, the
	// compatible campaign must still be identified, and the replay's transition
	// must still be the strictly later one.
	assertEqualClockAuthorityOrdering(t, ctx, dbPath, invID, a1, d0, rr.RunID, pr.RunID)

	// ---- C8: derived projection. Generate twice from identical records, then
	// regenerate after the relevant update. Substantive output must be
	// deterministic, only affected assessments may change, and
	// unexamined/blocked must stay distinguishable from each other and from a
	// pass. No manual status patch exists to reach for.
	currentA1 := gateCurrentContext(a1, invID, gateProjectRevision)
	genA, err := app.GenerateReviewCoverage(ctx, ReviewCoverageInput{DBPath: dbPath, PolicyID: ledger.policyID, SubjectRef: subjectRef, CurrentDependencies: currentA1})
	if err != nil {
		t.Fatalf("C8 first generation: %v", err)
	}
	genB, err := app.GenerateReviewCoverage(ctx, ReviewCoverageInput{DBPath: dbPath, PolicyID: ledger.policyID, SubjectRef: subjectRef, CurrentDependencies: currentA1})
	if err != nil {
		t.Fatalf("C8 second generation: %v", err)
	}
	// C8's determinism requirement now applies to SUBSTANTIVE content. The
	// export names its generation time (the contract requires it) and classifies
	// that line as non-semantic (the contract says so explicitly), so equality
	// is checked with that line stripped rather than by deleting the field.
	if review.StripNonSemantic(genA.Document) != review.StripNonSemantic(genB.Document) {
		t.Fatalf("C8: repeated generation from identical records must preserve substantive content")
	}
	// Provenance must actually be present: an export that cannot name its
	// generator or its inputs cannot be checked for compatibility later.
	if !strings.Contains(genA.Document, review.GeneratorVersion) {
		t.Fatalf("C8: the export must name its generator version:\n%s", genA.Document)
	}
	if !strings.Contains(genA.Document, "**generated at**") {
		t.Fatalf("C8: the export must name its generation time:\n%s", genA.Document)
	}
	if !strings.Contains(genA.Document, gateProjectRevision) {
		t.Fatalf("C8: the export must name the governing assessment's project revision %q:\n%s",
			gateProjectRevision, genA.Document)
	}
	if genA.Decision != genB.Decision {
		t.Fatalf("C8: repeated generation changed the decision: %s -> %s", genA.Decision, genB.Decision)
	}
	// Regeneration after the relevant update changes exactly the affected
	// current assessment — and nothing else. Under D0 the A1 assessment is the
	// stale one, so the decision flips while every record still appears.
	genShifted, err := app.GenerateReviewCoverage(ctx, ReviewCoverageInput{
		DBPath: dbPath, PolicyID: ledger.policyID, SubjectRef: subjectRef,
		CurrentDependencies: gateCurrentContext(d0, invID, gateProjectRevision),
	})
	if err != nil {
		t.Fatalf("C8 shifted generation: %v", err)
	}
	if genShifted.Document == genA.Document {
		t.Fatal("C8: a relevant population change must change the derived document")
	}
	for _, id := range []string{a0.Assessment.ID, a1Assessment.Assessment.ID, c1, c2, c3, c6, c7} {
		if !strings.Contains(genShifted.Document, id) {
			t.Fatalf("C8: regeneration must retain every record (%s missing):\n%s", id, genShifted.Document)
		}
	}

	// C8's distinguishability requirement, checked against a SECOND policy whose
	// obligations are deliberately unexamined and blocked. Collapsing these into
	// one "not passing" bucket is the failure mode being excluded.
	assertUnexaminedAndBlockedRemainDistinct(t, ctx, app, dbPath)

	// Writing the document is a file export of the same bytes, never a second
	// source of truth.
	outPath := t.TempDir() + "/COVERAGE.md"
	written, err := app.GenerateReviewCoverage(ctx, ReviewCoverageInput{
		DBPath: dbPath, PolicyID: ledger.policyID, SubjectRef: subjectRef, CurrentDependencies: currentA1, OutPath: outPath,
	})
	if err != nil {
		t.Fatalf("C8 write: %v", err)
	}
	if written.WrittenPath != outPath {
		t.Fatalf("C8: written path = %q, want %q", written.WrittenPath, outPath)
	}
	onDisk, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("C8 read back: %v", err)
	}
	if string(onDisk) != genA.Document {
		t.Fatal("C8: the exported file must be exactly the derived document")
	}

	// Export the run's coverage for the evidence bundle when asked. This is
	// opt-in via an env var rather than an unconditional repository write,
	// because a test that rewrites tracked files on every run turns a derived
	// snapshot into a source of merge conflicts.
	if dest := os.Getenv("NEWF_REVIEW_COVERAGE_OUT"); dest != "" {
		if err := os.WriteFile(dest, []byte(genA.Document), 0o644); err != nil {
			t.Fatalf("C8 bundle export: %v", err)
		}
		t.Logf("wrote generated coverage to %s", dest)
	}
}

// assertUnexaminedAndBlockedRemainDistinct pins C8's distinguishability
// requirement: never examined, examined-and-blocked, and conforming are three
// different states, and none of them is a pass.
//
// It uses its own policy so the main scenario's records stay untouched — the
// alternative (mutating the scenario's ledger) would destroy the very history
// C6/C7 assert is preserved.
func assertUnexaminedAndBlockedRemainDistinct(t *testing.T, ctx context.Context, app *App, dbPath string) {
	t.Helper()

	policy, err := app.DefineReviewPolicy(ctx, ReviewPolicyDefineInput{
		DBPath: dbPath, Key: "coverage-distinguishability-control", Revision: 1,
		DecisionName:       "whether the control obligations may be reported as satisfied",
		Owner:              "repository-maintainer",
		AuthoritySource:    reviewContractRef,
		ScopeJustification: "control policy for C8 distinguishability only; authorizes no product decision",
		Obligations: []ReviewObligationSpec{
			{Key: "never-examined", SemanticRevision: 1, Requirement: "an obligation nobody assessed",
				AcceptanceCriteria: "must report unexamined",
				ApplicabilityRule:  "applies to the control subject only",
				PrimaryOwner:       "repository-maintainer", Mandatory: true},
			{Key: "examined-but-blocked", SemanticRevision: 1, Requirement: "an obligation whose check could not run",
				AcceptanceCriteria: "must report blocked, not unexamined and not a pass",
				ApplicabilityRule:  "applies to the control subject only",
				PrimaryOwner:       "repository-maintainer", Mandatory: true},
		},
	})
	if err != nil {
		t.Fatalf("control policy: %v", err)
	}
	unexaminedID := policy.ObligationIDs["never-examined@1"]
	blockedID := policy.ObligationIDs["examined-but-blocked@1"]

	// The unexamined obligation gets an applicability decision and NOTHING else.
	// It must not be possible to make it look assessed by leaving it alone.
	if _, err := app.DecideReviewApplicability(ctx, ReviewApplicabilityInput{
		DBPath: dbPath, PolicyID: policy.PolicyID, ObligationID: unexaminedID,
		SubjectRef: "control:unexamined", Decision: review.Applies,
		Rationale:  "in scope for the control; deliberately left unassessed",
		Authorizer: "repository-maintainer",
	}); err != nil {
		t.Fatalf("control applicability (unexamined): %v", err)
	}

	blockedApplicability, err := app.DecideReviewApplicability(ctx, ReviewApplicabilityInput{
		DBPath: dbPath, PolicyID: policy.PolicyID, ObligationID: blockedID,
		SubjectRef: "control:blocked", Decision: review.Applies,
		Rationale:  "in scope for the control; its check is blocked",
		Authorizer: "repository-maintainer",
	})
	if err != nil {
		t.Fatalf("control applicability (blocked): %v", err)
	}
	blockedCheck, err := app.RecordReviewCheck(ctx, ReviewCheckInput{
		DBPath: dbPath, PolicyID: policy.PolicyID, ObligationID: blockedID,
		CaseLabel: "control-blocked", ProcedureRef: "unavailable-verifier",
		ProcedureRevision: reviewRecipeRevision,
		Mode:              review.ModeExecuted, Outcome: review.CheckBlocked,
		Blocker:     "the required independent verifier is unavailable in this environment",
		Executor:    reviewAssessor,
		Environment: "go test ./internal/pipeline",
	})
	if err != nil {
		t.Fatalf("control blocked check: %v", err)
	}
	// The assessment cites the blocked attempt and claims INCONCLUSIVE, which is
	// the honest outcome. Claiming `conforms` on this basis is separately
	// refused by the store trigger.
	if _, err := app.RecordReviewAssessment(ctx, ReviewAssessInput{
		DBPath: dbPath, PolicyID: policy.PolicyID, ObligationID: blockedID,
		ApplicabilityDecisionID: blockedApplicability.ID,
		SubjectRef:              "control:blocked", ContextRef: "control-context",
		Outcome: review.Inconclusive,
		Argument: "the check could not execute, so no conclusion about the requirement is available; " +
			"the blocker is retained rather than converted into support",
		Assessor:        reviewAssessor,
		ProjectRevision: gateProjectRevision, ContractHash: reviewContractRef, RecipeHash: reviewRecipeRevision,
		Dependencies: []ReviewDependencySpec{{Kind: "verifier_availability", Ref: "unavailable-verifier",
			WhyRelevant: "the requirement can only be established by executing this verifier"}},
		CheckAttemptIDs: []string{blockedCheck.ID},
	}); err != nil {
		t.Fatalf("control blocked assessment: %v", err)
	}

	// The control supplies its declared dependency's CURRENT value so the
	// examination states it is testing are what the projection reports. Omitting
	// it would make both obligations `compatibility_unknown` — a true statement
	// about an underspecified request, but it would mask the very distinction
	// this control exists to pin (never-examined versus examined-and-blocked).
	cov, err := app.GenerateReviewCoverage(ctx, ReviewCoverageInput{
		DBPath: dbPath, PolicyID: policy.PolicyID,
		CurrentDependencies: map[string]string{"verifier_availability": "unavailable-verifier"},
	})
	if err != nil {
		t.Fatalf("control coverage: %v", err)
	}
	if cov.Decision != string(review.DecisionUndetermined) {
		t.Fatalf("control decision = %s, want %s", cov.Decision, review.DecisionUndetermined)
	}
	states := map[string]string{}
	for _, o := range cov.Obligations {
		states[o.Key] = o.State
	}
	if states["never-examined"] != string(review.StateUnexamined) {
		t.Fatalf("an obligation with no assessment must report unexamined, got %q", states["never-examined"])
	}
	// The assessment honestly said `inconclusive` and cited a BLOCKED attempt, so
	// the derived state is inconclusive AND the execution blocker is reported as
	// its own reason. Accepting either state alone would let the two collapse.
	if states["examined-but-blocked"] != string(review.StateInconclusive) {
		t.Fatalf("an examined-but-blocked obligation must report inconclusive, got %q", states["examined-but-blocked"])
	}
	if states["never-examined"] == states["examined-but-blocked"] {
		t.Fatal("unexamined and blocked must remain distinguishable in the derived projection")
	}
	if !containsString(cov.Reasons, review.ReasonUnexamined) || !containsString(cov.Reasons, review.ReasonExecutionBlocked) {
		t.Fatalf("control reasons must carry both unexamined and execution_blocked: %v", cov.Reasons)
	}
	if !strings.Contains(cov.Document, "the required independent verifier is unavailable") {
		t.Fatalf("the blocker text must survive into the export:\n%s", cov.Document)
	}
}

// assertEqualClockAuthorityOrdering pins C7's equal-clock/order stress
// requirement.
//
// Runs 1-3 never executed this branch: run 2 deliberately used a strictly
// monotonic clock, which is the one configuration where a timestamp-ordered
// implementation and a sequence-ordered one behave identically. Under the gate's
// FIXED clock the two campaigns tie, so this is where the distinction is
// observable.
//
// What must hold: the compatible campaign is found by population match (not by
// being newest), the replay's transition is strictly later in the store's own
// total order, and the compatible authority is NOT the replay — the exact F-2
// displacement, now checked at the ordering layer rather than only through the
// frontier decision.
func assertEqualClockAuthorityOrdering(t *testing.T, ctx context.Context, dbPath, invID, a1, d0, reassessRun, replayRun string) {
	t.Helper()
	repo := openTestStore(t, ctx, dbPath)
	defer repo.Close()

	reassessPop, found, err := repo.GetChallengeAssessmentPopulation(ctx, reassessRun, invID)
	if err != nil || !found {
		t.Fatalf("C7 stress: load reassessment population: %v (found=%v)", err, found)
	}
	replayPop, found, err := repo.GetChallengeAssessmentPopulation(ctx, replayRun, invID)
	if err != nil || !found {
		t.Fatalf("C7 stress: load replay population: %v (found=%v)", err, found)
	}
	// The equal-clock precondition. If this ever stops holding, the branch below
	// silently degenerates into the monotonic case runs 1-3 already covered, so
	// the precondition is asserted rather than assumed.
	if reassessPop.CreatedAt != replayPop.CreatedAt {
		t.Fatalf("C7 stress: this branch requires an equal clock, got reassess=%q replay=%q",
			reassessPop.CreatedAt, replayPop.CreatedAt)
	}

	current, found, err := repo.GetLatestCompatibleAuthority(ctx, invID, a1)
	if err != nil || !found {
		t.Fatalf("C7 stress: the campaign that assessed current population %s must remain findable: %v (found=%v)", a1, err, found)
	}
	historical, found, err := repo.GetLatestCompatibleAuthority(ctx, invID, d0)
	if err != nil || !found {
		t.Fatalf("C7 stress: the historical population %s must stay reproducible: %v (found=%v)", d0, err, found)
	}

	if current.RunID != reassessRun {
		t.Fatalf("C7 stress: current authority for %s must be the reassessment campaign %s, got %s",
			a1, reassessRun, current.RunID)
	}
	// The replay really is the later execution: without this, "the compatible
	// campaign was selected" could be true merely because nothing newer existed,
	// and the test would prove nothing about displacement.
	if historical.TransitionSeq <= current.TransitionSeq {
		t.Fatalf("C7 stress: the replay must be strictly later in the store's total order "+
			"(replay seq=%d, compatible seq=%d); otherwise this case cannot demonstrate that "+
			"recency was rejected in favor of compatibility",
			historical.TransitionSeq, current.TransitionSeq)
	}
	if current.RunID == replayRun {
		t.Fatal("C7 stress: the replay must not become the compatible current authority (F-2)")
	}
}

// gateCurrentContext builds a COMPLETE current-dependency context for the
// scenario's assessments, varying only the population.
//
// Completeness matters: a declared dependency with no supplied current value is
// `compatibility_unknown`, not compatible. That is deliberate — an incompletely
// specified request must not inherit a historical pass — so a case that means to
// test staleness has to supply every other kind, or it would measure the
// caller's omission instead of the dependency's movement.
func gateCurrentContext(population, invID, projectRevision string) map[string]string {
	return map[string]string{
		depKindAssessmentPopulation: population,
		depKindPolicyRevision:       reviewPolicyKey + "@1",
		depKindCandidateContent:     invID,
		depKindProjectRevision:      projectRevision,
	}
}

// withUnrelatedChange copies a current context and adds one dependency kind the
// assessments never declared. Copying keeps C5 from mutating the shared context
// other cases rely on.
func withUnrelatedChange(current map[string]string, kind, ref string) map[string]string {
	out := make(map[string]string, len(current)+1)
	for k, v := range current {
		out[k] = v
	}
	out[kind] = ref
	return out
}

func clusterSignatureSet(run ClusterRunView) map[string]bool {
	out := map[string]bool{}
	for _, c := range run.Clusters {
		for _, m := range c.Members {
			out[m.SignatureID] = true
		}
	}
	return out
}

func copyStringSet(in map[string]bool) map[string]bool {
	out := make(map[string]bool, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func assertStringSetEqual(t *testing.T, got, want map[string]bool) {
	t.Helper()
	for k := range want {
		if !got[k] {
			t.Fatalf("C3: population missing signature %s; got=%v want=%v", k, got, want)
		}
	}
	for k := range got {
		if !want[k] {
			t.Fatalf("C3: population has unexpected signature %s; got=%v want=%v", k, got, want)
		}
	}
}

func generationTargetsContain(targets []provider.GenerationTarget, invariantID string) bool {
	for _, target := range targets {
		if target.InvariantID == invariantID {
			return true
		}
	}
	return false
}

// containsString reports membership without pulling in a helper dependency.
func containsString(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}
