package pipeline

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/canon"
)

// TestIntegrationWitnessAttemptBinding is the committed regression for the
// attribution slice (2026-09-12 C1–C8 review, remediation handoff 3).
//
// The review's C3 finding: a supplied tuple establishes only "this submitted
// tuple does not satisfy the equation" — not "this proposal's executed
// mechanism produced this tuple and failed its specified obligation".
// Acceptance: witness admission carries a CHECKABLE attempt→output link.
//
// Exercised here:
//  1. `witness check --procedure` executes a registered deterministic bounded
//     attempt whose OUTPUT IS the checked tuple, and records the binding in
//     the same transaction as the evaluation;
//  2. admission RECOMPUTES the recorded procedure over the recorded params
//     and names the verified binding in the rule basis — attribution is
//     checked, not declared;
//  3. the supplied-tuple path keeps its explicitly weaker provenance: still
//     rule-admissible (unchanged behavior), with the attribution gap recorded
//     in the evaluation notes and no binding claim in the basis;
//  4. abstention is an input-level refusal — no tuple, no claim, nothing
//     persisted — and --procedure/--tuple are mutually exclusive.
func TestIntegrationWitnessAttemptBinding(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 12, 15, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = dataDrivenMiner{}
	app.challengerFn = biasOnlyChallenger{}
	// Two distinct proposals: one for the executed-attempt path, one for the
	// supplied-tuple contrast (distinct proposals keep the two provenance
	// regimes on separate admission subjects).
	app.generatorFn = pcGenerator{signatures: []canon.MechanismSignature{
		pcSignature(pcResidueLocality, "", true),
		pcSignature(pcResidueLocality, "core.operator.decompose", true),
	}}

	problemID, invID, _ := mineOneCandidate(t, ctx, app, dbPath)
	if _, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID}); err != nil {
		t.Fatalf("challenge: %v", err)
	}
	gen, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil || len(gen.Generation.Proposals) < 2 {
		t.Fatalf("generate: %v (proposals=%d, need 2)", err, len(gen.Generation.Proposals))
	}
	boundProposal := gen.Generation.Proposals[0].ID
	suppliedProposal := gen.Generation.Proposals[1].ID

	// 1) Executed bounded attempt: equal-denominator n=7 computes (7,6,6,6),
	// a deliberately failing probe. The tuple is NOT supplied.
	res, err := app.WitnessCheck(ctx, WitnessCheckInput{
		DBPath: dbPath, ProposalID: boundProposal,
		Procedure: "equal-denominator", Params: map[string]string{"n": "7"},
	})
	if err != nil {
		t.Fatalf("witness (procedure): %v", err)
	}
	if res.WitnessVerdict != "witness-invalid" {
		t.Fatalf("want witness-invalid, got %q", res.WitnessVerdict)
	}
	if res.AttemptBinding == nil || res.AttemptBinding.Procedure != "equal-denominator" || res.AttemptBinding.ParamsCanonical != "n=7" {
		t.Fatalf("response must carry the recorded binding: %+v", res.AttemptBinding)
	}
	if res.AttemptBinding.TupleCanonical != res.CanonicalClaim {
		t.Fatalf("the bound tuple must BE the checked claim: %q vs %q", res.AttemptBinding.TupleCanonical, res.CanonicalClaim)
	}
	if !strings.Contains(res.Evaluation.Notes, "attempt-bound: equal-denominator@") {
		t.Fatalf("evaluation notes must name the binding: %q", res.Evaluation.Notes)
	}

	// 2) Admission rechecks the binding by recomputation and names it.
	admit, err := app.AdmitEvidence(ctx, AdmitEvidenceInput{DBPath: dbPath, ProblemID: problemID, EvaluationID: res.Evaluation.ID})
	if err != nil {
		t.Fatalf("admit (bound): %v", err)
	}
	if len(admit.Admitted) != 1 || admit.Admitted[0].AdmittedBy != "rule" {
		t.Fatalf("bound witness failure must be rule-admitted: %+v", admit)
	}
	if !strings.Contains(admit.Admitted[0].Basis, "attempt→output binding verified by recomputation: equal-denominator@") ||
		!strings.Contains(admit.Admitted[0].Basis, "(n=7)") {
		t.Fatalf("admission basis must carry the verified checkable link: %q", admit.Admitted[0].Basis)
	}

	// 3) Supplied-tuple contrast on a second proposal: same arithmetic,
	// operator provenance. Still rule-admissible; the attribution gap is
	// explicit and no binding verification is claimed.
	res2, err := app.WitnessCheck(ctx, WitnessCheckInput{
		DBPath: dbPath, ProposalID: suppliedProposal, Tuple: "7,6,6,6",
		Note: "attribution-slice regression: operator-supplied copy of the probe tuple",
	})
	if err != nil {
		t.Fatalf("witness (supplied): %v", err)
	}
	if res2.AttemptBinding != nil {
		t.Fatalf("a supplied tuple must not fabricate a binding: %+v", res2.AttemptBinding)
	}
	if !strings.Contains(res2.Evaluation.Notes, "supplied tuple; no executed-attempt binding") {
		t.Fatalf("supplied-tuple notes must record the attribution gap: %q", res2.Evaluation.Notes)
	}
	admit2, err := app.AdmitEvidence(ctx, AdmitEvidenceInput{DBPath: dbPath, ProblemID: problemID, EvaluationID: res2.Evaluation.ID})
	if err != nil {
		t.Fatalf("admit (supplied): %v", err)
	}
	if len(admit2.Admitted) != 1 || admit2.Admitted[0].AdmittedBy != "rule" {
		t.Fatalf("supplied-tuple path must keep its rule admissibility: %+v", admit2)
	}
	if strings.Contains(admit2.Admitted[0].Basis, "binding verified") {
		t.Fatalf("no binding exists; the basis must not claim one: %q", admit2.Admitted[0].Basis)
	}

	// 4) Abstention refuses at the input level: greedy n=8 has zero remainder
	// after x=2 (4/8 = 1/2 exactly), so there is no two-term expansion.
	if _, err := app.WitnessCheck(ctx, WitnessCheckInput{
		DBPath: dbPath, ProposalID: boundProposal,
		Procedure: "greedy", Params: map[string]string{"n": "8"},
	}); err == nil || !strings.Contains(err.Error(), "abstains") {
		t.Fatalf("abstention must refuse with a typed reason, got %v", err)
	}
	// Mutual exclusion: an executed attempt computes its own tuple.
	if _, err := app.WitnessCheck(ctx, WitnessCheckInput{
		DBPath: dbPath, ProposalID: boundProposal,
		Procedure: "equal-denominator", Params: map[string]string{"n": "7"}, Tuple: "7,6,6,6",
	}); err == nil || !strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("procedure+tuple must be refused, got %v", err)
	}

	// 5) A procedure can also produce a VALID witness: greedy n=7 yields
	// (7,2,15,210) — verdict success, binding recorded all the same.
	res3, err := app.WitnessCheck(ctx, WitnessCheckInput{
		DBPath: dbPath, ProposalID: boundProposal,
		Procedure: "greedy", Params: map[string]string{"n": "7"},
	})
	if err != nil {
		t.Fatalf("witness (greedy valid): %v", err)
	}
	if res3.WitnessVerdict != "witness-valid" || res3.AttemptBinding == nil {
		t.Fatalf("greedy n=7 must be a bound valid witness: verdict=%q binding=%+v", res3.WitnessVerdict, res3.AttemptBinding)
	}
}
