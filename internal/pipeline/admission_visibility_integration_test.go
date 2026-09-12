package pipeline

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/provider"
	"github.com/instagrim-dev/newf/internal/verify"
)

// TestIntegrationLaterStrongerEvaluationReachesAdmission is the committed
// regression for F-1 (2026-09-12 review run, remediation handoff 2): every
// applicable evaluation remains independently visible to admission under its
// exact context, even when the proposal already carries an earlier failure
// marker and admission decision.
//
// Before v45 the evaluated_failures marker was keyed per PROPOSAL with
// INSERT OR IGNORE: after a model-judged failure was recorded (and withheld),
// a later witness-checked (reproducible-strength) failure of the SAME proposal
// never appeared to AdmitEvidence — stronger evidence silently shadowed by a
// weaker earlier marker (reproduced in the review run's retained
// f3-shadowing log).
//
// Acceptance (from the review bundle):
//  1. the same-proposal witness-after-model-judged case is rule-admissible
//     exactly once — a rerun of the rule pass repeats nothing;
//  2. the earlier withheld decision is preserved, not replaced (visibility is
//     not strength-ranked replacement);
//  3. no-inflation companion: a SECOND admitted evaluation of the same
//     proposal revises the same approach (stable logical identity
//     "frontier-proposal:<id>"), so the current-heads population does not
//     grow a second member for the same underlying attempt.
func TestIntegrationLaterStrongerEvaluationReachesAdmission(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 12, 14, 30, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = dataDrivenMiner{}
	app.challengerFn = biasOnlyChallenger{}
	app.modelVerifierFn = provider.NewFixtureModelVerifier(verify.VerdictFailure, "high")

	problemID, invID, _ := mineOneCandidate(t, ctx, app, dbPath)
	if _, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID}); err != nil {
		t.Fatalf("challenge: %v", err)
	}
	gen, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil || len(gen.Generation.Proposals) == 0 {
		t.Fatalf("generate: %v (proposals=%d)", err, len(gen.Generation.Proposals))
	}
	proposalID := gen.Generation.Proposals[0].ID

	base, err := app.BuildClustering(ctx, ClusterBuildInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("baseline cluster: %v", err)
	}
	basePop := base.ClusterRun.SignatureCount

	// 1) Model-judged failure -> withheld by rule.
	res, err := app.Evaluate(ctx, EvaluateInput{DBPath: dbPath, ProblemID: problemID, ProposalID: proposalID})
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	modelEval := res.Run.Evaluations[0]
	if modelEval.VerificationStrength != "single-model-judgment" {
		t.Fatalf("fixture must be model-judged, got %q", modelEval.VerificationStrength)
	}
	admit1, err := app.AdmitEvidence(ctx, AdmitEvidenceInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil || len(admit1.Withheld) != 1 || len(admit1.Admitted) != 0 {
		t.Fatalf("rule pass must withhold the model judgment: %+v err=%v", admit1, err)
	}

	// 2) Later witness-checked failure of the SAME proposal. Before v45 this
	// evaluation never reached admission at all.
	wit, err := app.WitnessCheck(ctx, WitnessCheckInput{
		DBPath: dbPath, ProposalID: proposalID, Tuple: "7,5,5,5",
		Note: "F-1 regression: equal-denominator probe x=y=z=round(3n/4), n=7; deliberately failing bounded attempt",
	})
	if err != nil {
		t.Fatalf("witness: %v", err)
	}
	if wit.WitnessVerdict != "witness-invalid" {
		t.Fatalf("want witness-invalid, got %q", wit.WitnessVerdict)
	}
	admit2, err := app.AdmitEvidence(ctx, AdmitEvidenceInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("admit (witness): %v", err)
	}
	if len(admit2.Admitted) != 1 {
		t.Fatalf("the later stronger evaluation must be rule-admitted: %+v", admit2)
	}
	adm := admit2.Admitted[0]
	if adm.ProposalID != proposalID || adm.ObservationKind != ObservationDomainCheckedFailure || adm.AdmittedBy != "rule" {
		t.Fatalf("admission must be the same proposal's domain-checked failure by rule: %+v", adm)
	}
	if adm.EvaluationID == modelEval.ID {
		t.Fatal("the admitted row must reference the witness evaluation, not the model judgment")
	}
	// The earlier withheld decision is preserved, not replaced.
	list, err := app.ListEvidenceAdmissions(ctx, AdmitEvidenceInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("list admissions: %v", err)
	}
	decisions := map[string]bool{}
	for _, a := range list.Admissions {
		decisions[a.Decision+"/"+a.EvaluationID] = true
	}
	if !decisions["withheld/"+modelEval.ID] || !decisions["admitted/"+adm.EvaluationID] {
		t.Fatalf("ledger must keep the withheld model judgment AND the admitted witness failure: %+v", list.Admissions)
	}

	// Exactly once: a rerun of the rule pass repeats nothing.
	admit3, err := app.AdmitEvidence(ctx, AdmitEvidenceInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil || len(admit3.Admitted) != 0 || len(admit3.Withheld) != 0 {
		t.Fatalf("rerun must decide nothing new: %+v err=%v", admit3, err)
	}
	if len(admit3.Skipped) != 2 {
		t.Fatalf("rerun must skip both decided evaluations: %+v", admit3.Skipped)
	}

	// The admitted observation enters the population exactly once.
	after, err := app.BuildClustering(ctx, ClusterBuildInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("recluster: %v", err)
	}
	if after.ClusterRun.SignatureCount != basePop+1 {
		t.Fatalf("population must grow by exactly one: %d -> %d", basePop, after.ClusterRun.SignatureCount)
	}

	// 3) No-inflation: another witness-checked failure of the SAME proposal is
	// visible and admissible under its own context, but it revises the same
	// approach — the current-heads population must NOT grow a second member.
	wit2, err := app.WitnessCheck(ctx, WitnessCheckInput{
		DBPath: dbPath, ProposalID: proposalID, Tuple: "11,8,8,8",
		Note: "F-1 regression: equal-denominator probe n=11 (second bounded attempt, same proposal)",
	})
	if err != nil {
		t.Fatalf("witness 2: %v", err)
	}
	if wit2.WitnessVerdict != "witness-invalid" {
		t.Fatalf("want witness-invalid, got %q", wit2.WitnessVerdict)
	}
	admit4, err := app.AdmitEvidence(ctx, AdmitEvidenceInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("admit (witness 2): %v", err)
	}
	if len(admit4.Admitted) != 1 {
		t.Fatalf("the second evaluation must be independently visible and admitted: %+v", admit4)
	}
	final, err := app.BuildClustering(ctx, ClusterBuildInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("final cluster: %v", err)
	}
	if final.ClusterRun.SignatureCount != basePop+1 {
		t.Fatalf("per-evaluation visibility must not inflate the population: %d -> %d (want %d)",
			basePop, final.ClusterRun.SignatureCount, basePop+1)
	}
	if !strings.Contains(admit4.Admitted[0].Basis, "reproducible") {
		t.Fatalf("second admission must carry its own reproducible-strength basis: %q", admit4.Admitted[0].Basis)
	}
}
