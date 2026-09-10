package pipeline

import (
	"context"
	"testing"
	"time"
)

// TestIntegrationEvaluateEndToEnd runs the full offline M5.2 path on top of the
// M5.1 substrate: seed -> cluster -> failure-space -> mine -> challenge
// (-> surviving) -> frontier generate -> evaluate -> evaluation show. It asserts
// the evaluation run completes, the proposal's result is populated, and every
// evaluation carries BOTH a verdict AND its verification strength (R1) produced
// by a deterministic tier (R3) — never a bare model opinion.
func TestIntegrationEvaluateEndToEnd(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = dataDrivenMiner{}
	app.challengerFn = biasOnlyChallenger{}

	problemID, invID, _ := mineOneCandidate(t, ctx, app, dbPath)

	if resp, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID}); err != nil {
		t.Fatalf("challenge: %v", err)
	} else if resp.Reports[0].StateAfter != "surviving" {
		t.Fatalf("state after bias-only campaign = %q, want surviving", resp.Reports[0].StateAfter)
	}

	gen, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("frontier generate: %v", err)
	}
	if len(gen.Generation.Proposals) == 0 {
		t.Fatal("expected at least one proposal to evaluate")
	}
	proposalID := gen.Generation.Proposals[0].ID

	// Evaluate the single proposal cheap-first.
	res, err := app.Evaluate(ctx, EvaluateInput{DBPath: dbPath, ProblemID: problemID, ProposalID: proposalID})
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if res.Run.Mode != "proposal" || res.Run.RoutingPolicy != "cheap-first" {
		t.Fatalf("unexpected run metadata: mode=%q policy=%q", res.Run.Mode, res.Run.RoutingPolicy)
	}
	if len(res.Run.Evaluations) != 1 {
		t.Fatalf("want 1 evaluation, got %d", len(res.Run.Evaluations))
	}
	ev := res.Run.Evaluations[0]
	// R1: an outcome is never seen without its strength.
	if ev.Verdict == "" || ev.VerificationStrength == "" || ev.VerifierKind == "" {
		t.Fatalf("evaluation missing verdict/kind/strength: %+v", ev)
	}
	// R3: a code-verified break (violates on the proposed signature, confirmed by
	// M5.1) is decided by a deterministic tier, not the model.
	if ev.VerifierKind == "model-judgment" {
		t.Fatalf("a code-verifiable proposal must not be decided by the model tier; got %q", ev.VerifierKind)
	}
	if ev.ProviderInvocationID != "" {
		t.Fatalf("a deterministic-tier evaluation must not carry a provider invocation; got %q", ev.ProviderInvocationID)
	}

	// Run lifecycle reflects success.
	run, err := app.ShowRun(ctx, LookupInput{DBPath: dbPath, ID: res.Run.RunID})
	if err != nil {
		t.Fatalf("run show: %v", err)
	}
	if run.Run.Status != "completed" {
		t.Fatalf("evaluate run status = %q, want completed", run.Run.Status)
	}

	// R5: the proposal result is now populated and mirrors the verdict.
	shown, err := app.ShowFrontier(ctx, FrontierShowInput{DBPath: dbPath, GenerationID: gen.Generation.ID})
	if err != nil {
		t.Fatalf("frontier show: %v", err)
	}
	var result string
	for _, p := range shown.Generation.Proposals {
		if p.ID == proposalID {
			result = p.Result
		}
	}
	if result != ev.Verdict {
		t.Fatalf("proposal result %q does not mirror evaluation verdict %q (R5)", result, ev.Verdict)
	}

	// evaluation show round-trips the strength-stamped verdict.
	es, err := app.ShowEvaluation(ctx, EvaluationShowInput{DBPath: dbPath, EvaluationID: res.Run.ID})
	if err != nil {
		t.Fatalf("evaluation show: %v", err)
	}
	if len(es.Run.Evaluations) != 1 || es.Run.Evaluations[0].VerificationStrength != ev.VerificationStrength {
		t.Fatalf("evaluation show lost the strength stamp: %+v", es.Run.Evaluations)
	}

	// R6: if the verdict was a failure/partial_failure it re-enters the atlas.
	if ev.Verdict == "failure" || ev.Verdict == "partial_failure" {
		fails, ferr := app.ListEvaluatedFailures(ctx, EvaluationListInput{DBPath: dbPath, ProblemID: problemID})
		if ferr != nil {
			t.Fatalf("list evaluated failures: %v", ferr)
		}
		found := false
		for _, f := range fails.Failures {
			if f.ProposalID == proposalID {
				found = true
			}
		}
		if !found {
			t.Fatal("a failed proposal must re-enter the failure atlas (R6)")
		}
	}

	// evaluation list surfaces the run.
	list, err := app.ListEvaluations(ctx, EvaluationListInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("evaluation list: %v", err)
	}
	if len(list.Runs) != 1 || list.Runs[0].ID != res.Run.ID {
		t.Fatalf("expected the evaluation run on the list surface, got %+v", list.Runs)
	}
}

// TestIntegrationEvaluateHoldoutRefused proves the holdout mode is refused at the
// service boundary with a deferred-to-M7 error (R9), not half-built.
func TestIntegrationEvaluateHoldoutRefused(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)

	_, err := app.Evaluate(ctx, EvaluateInput{DBPath: dbPath, ProblemID: "prb_x", Mode: "holdout"})
	if err == nil {
		t.Fatal("holdout mode must be refused")
	}
	if got := err.Error(); got == "" || !contains(got, "M7") {
		t.Fatalf("holdout refusal should mention M7 deferral; got %q", got)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
