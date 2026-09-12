package pipeline

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/provider"
	"github.com/instagrim-dev/newf/internal/verify"
	"github.com/instagrim-dev/newf/internal/witness"
)

// TestIntegrationWitnessCheckCarriesToRuleAdmission is the cemented D2-C
// acceptance check (issue #23, decision recorded 2026-09-12):
//
//	one decisive witness-invalid evaluation rule-admits as
//	domain-checked-failure WITHOUT attestation, and its admission basis
//	names the canonical witness; the proposal wire vocabulary is unchanged.
//
// The carrier is `newf witness check`: the tuple is checked exactly BEFORE
// any write, and the verdict is persisted as a reproducible-computation
// evaluation (subject domain-goal) through the same transactional choke
// point every evaluation uses — including the evaluated_failures re-entry
// marker the admission pass consumes.
func TestIntegrationWitnessCheckCarriesToRuleAdmission(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
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
		t.Fatalf("frontier generate: %v (proposals=%d)", err, len(gen.Generation.Proposals))
	}
	proposalID := gen.Generation.Proposals[0].ID

	// A malformed tuple is an input error, never a domain verdict.
	if _, err := app.WitnessCheck(ctx, WitnessCheckInput{
		DBPath: dbPath, ProposalID: proposalID, Tuple: "2,4,not-a-number", Note: "n",
	}); err == nil {
		t.Fatal("malformed tuple must be refused before any write")
	}
	// A witness claim without provenance is refused.
	if _, err := app.WitnessCheck(ctx, WitnessCheckInput{
		DBPath: dbPath, ProposalID: proposalID, Tuple: "2,4,21,84",
	}); err == nil || !strings.Contains(err.Error(), "provenance") {
		t.Fatalf("missing note must be refused, got %v", err)
	}

	// The decisive witness-invalid outcome: (4,21,84) misses for n=2 —
	// the issue's near-miss vector, rejected exactly.
	res, err := app.WitnessCheck(ctx, WitnessCheckInput{
		DBPath: dbPath, ProposalID: proposalID, Tuple: "2,4,21,84",
		Note: "produced by the attempted mechanism's step 3 (test fixture)",
	})
	if err != nil {
		t.Fatalf("witness check: %v", err)
	}
	if res.WitnessVerdict != "witness-invalid" || res.Detail == "" {
		t.Fatalf("near-miss tuple must be witness-invalid with detail: %+v", res)
	}
	ev := res.Evaluation
	if ev.Verdict != string(verify.VerdictFailure) ||
		ev.VerifierKind != string(verify.KindReproducibleComputation) ||
		ev.VerificationStrength != string(verify.StrengthReproducible) ||
		ev.VerificationSubject != string(verify.SubjectDomainGoal) {
		t.Fatalf("witness evaluation must be a reproducible-computation domain-goal failure: %+v", ev)
	}
	if ev.ToolName != witness.CheckerName || ev.ToolVersion != witness.CheckerVersion {
		t.Fatalf("evaluation must attribute the checker: %+v", ev)
	}
	if !strings.Contains(ev.Notes, res.CanonicalClaim) {
		t.Fatalf("evaluation notes must carry the canonical claim %q: %q", res.CanonicalClaim, ev.Notes)
	}

	// The failure re-entered as a marker consumable by the admission pass.
	fl, err := app.ListEvaluatedFailures(ctx, EvaluationListInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil || len(fl.Failures) != 1 || fl.Failures[0].EvaluationID != ev.ID {
		t.Fatalf("witness failure must re-enter as an evaluated-failure marker: %v %+v", err, fl.Failures)
	}

	// The acceptance check proper: the batch RULE pass (no attestation)
	// admits it as a domain-checked failure, basis naming the canonical
	// witness via the checker attribution.
	adm, err := app.AdmitEvidence(ctx, AdmitEvidenceInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("evidence admit: %v", err)
	}
	if len(adm.Admitted) != 1 || len(adm.Withheld) != 0 {
		t.Fatalf("witness-invalid evaluation must rule-admit: %+v", adm)
	}
	got := adm.Admitted[0]
	if got.ObservationKind != ObservationDomainCheckedFailure || got.AdmittedBy != "rule" {
		t.Fatalf("admission must be a rule-admitted domain-checked failure: %+v", got)
	}
	if !strings.Contains(got.Basis, witness.CheckerName) || !strings.Contains(got.Basis, res.CanonicalClaim) {
		t.Fatalf("admission basis must name the checker and canonical witness: %q", got.Basis)
	}

	// A witness-VALID check on the same proposal records success and does NOT
	// add a failure marker (append-only re-evaluation; the proposal's result
	// stays at its first-set verdict by the R5 one-time rule).
	ok, err := app.WitnessCheck(ctx, WitnessCheckInput{
		DBPath: dbPath, ProposalID: proposalID, Tuple: "2,1,2,2",
		Note: "corrected tuple (test fixture)",
	})
	if err != nil {
		t.Fatalf("witness check valid: %v", err)
	}
	if ok.WitnessVerdict != "witness-valid" || ok.Evaluation.Verdict != string(verify.VerdictSuccess) {
		t.Fatalf("valid witness must record success: %+v", ok)
	}
	fl2, err := app.ListEvaluatedFailures(ctx, EvaluationListInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil || len(fl2.Failures) != 1 {
		t.Fatalf("a success must not add a failure marker: %v %+v", err, fl2.Failures)
	}
}
