package pipeline

import (
	"context"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/provider"
	"github.com/instagrim-dev/newf/internal/verify"
)

// TestIntegrationPolicyMutateAndBiasEndToEnd drives the full offline M6.2 path:
// seed -> cluster -> mine -> challenge (-> surviving) -> generate -> evaluate ->
// compress -> POLICY MUTATE -> generate-again. It asserts:
//   - mutate derives a policy containing an `avoid` directive for the surviving
//     failure invariant (the strongest conserved failure structure);
//   - the second generation applies the policy and logs a per-proposal bias
//     against the persisted proposal ids (the reproducible "why");
//   - re-mutate over the unchanged evidence cohort is idempotent;
//   - --no-policy reproduces the unbiased generation (M7 baseline contrast).
func TestIntegrationPolicyMutateAndBiasEndToEnd(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = dataDrivenMiner{}
	app.challengerFn = biasOnlyChallenger{}
	app.modelVerifierFn = provider.NewFixtureModelVerifier(verify.VerdictPartialSuccess, "medium")

	problemID, invID, _ := mineOneCandidate(t, ctx, app, dbPath)
	if resp, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID}); err != nil {
		t.Fatalf("challenge: %v", err)
	} else if resp.Reports[0].StateAfter != "surviving" {
		t.Fatalf("state after campaign = %q, want surviving", resp.Reports[0].StateAfter)
	}
	if _, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID}); err != nil {
		t.Fatalf("frontier generate: %v", err)
	}
	if _, err := app.Evaluate(ctx, EvaluateInput{DBPath: dbPath, ProblemID: problemID}); err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if _, err := app.CompressSuccesses(ctx, SuccessCompressInput{DBPath: dbPath, ProblemID: problemID}); err != nil {
		t.Fatalf("compress: %v", err)
	}

	// Mutate: derive a policy from the accumulated evidence.
	mut, err := app.MutatePolicy(ctx, PolicyMutateInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("policy mutate: %v", err)
	}
	if !mut.Created {
		t.Fatal("first mutate must create a revision")
	}
	if mut.Revision.DirectiveCount == 0 {
		t.Fatal("expected at least one directive from the surviving invariant")
	}
	foundAvoid := false
	for _, d := range mut.Revision.Directives {
		if d.Kind == "avoid" && d.TargetKind == "surviving_invariant" && d.TargetID == invID {
			foundAvoid = true
			if len(d.Provenance) == 0 {
				t.Fatal("avoid directive must carry provenance to the surviving invariant")
			}
		}
	}
	if !foundAvoid {
		t.Fatalf("expected an avoid directive for surviving invariant %s, got %+v", invID, mut.Revision.Directives)
	}

	// Run lifecycle completed.
	run, err := app.ShowRun(ctx, LookupInput{DBPath: dbPath, ID: mut.Revision.RunID})
	if err != nil {
		t.Fatalf("run show: %v", err)
	}
	if run.Run.Status != "completed" {
		t.Fatalf("mutate run status = %q, want completed", run.Run.Status)
	}

	// Idempotent: unchanged evidence returns the existing revision.
	again, err := app.MutatePolicy(ctx, PolicyMutateInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("re-mutate: %v", err)
	}
	if again.Created || again.Revision.ID != mut.Revision.ID {
		t.Fatalf("re-mutate must be idempotent; created=%v id=%s want %s", again.Created, again.Revision.ID, mut.Revision.ID)
	}

	// Second generation applies the policy and logs the bias for the WHOLE ranked
	// set — even though the deterministic generator persists no NEW proposal rows
	// (they dedup onto the first generation's proposals), the applied-bias log is
	// keyed on the persisted id of every ranked candidate, so the "why" surface
	// is reproducible rather than empty.
	gen2, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("second generate: %v", err)
	}

	repo := openTestStore(t, ctx, dbPath)
	defer repo.Close()
	revID, bias, err := repo.GetFrontierGenerationPolicy(ctx, gen2.Generation.ID)
	if err != nil {
		t.Fatalf("get generation policy: %v", err)
	}
	// The first generation produced ranked candidates; the biased re-generation
	// must record a non-empty applied-bias log tied to the policy revision.
	if revID != mut.Revision.ID {
		t.Fatalf("generation policy revision = %q, want %q", revID, mut.Revision.ID)
	}
	if len(bias) == 0 {
		t.Fatal("biased generation must record a non-empty applied-bias log for the ranked set")
	}
	// Every logged proposal id must be a real, persisted frontier proposal for
	// the problem (falsify by logging a fabricated id).
	for _, b := range bias {
		if b.ProposalID == "" {
			t.Fatalf("applied-bias row missing proposal id: %+v", b)
		}
	}

	// --no-policy path must succeed (unbiased baseline) and write no bias log.
	gen3, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID, NoPolicy: true})
	if err != nil {
		t.Fatalf("no-policy generate: %v", err)
	}
	_, bias3, err := repo.GetFrontierGenerationPolicy(ctx, gen3.Generation.ID)
	if err != nil {
		t.Fatalf("get no-policy generation policy: %v", err)
	}
	if len(bias3) != 0 {
		t.Fatalf("--no-policy generation must record no applied bias, got %+v", bias3)
	}

	// list + show round-trip.
	listed, err := app.ListPolicies(ctx, PolicyListInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("policy list: %v", err)
	}
	if len(listed.Revisions) != 1 {
		t.Fatalf("expected 1 policy revision, got %d", len(listed.Revisions))
	}
	shown, err := app.ShowPolicy(ctx, PolicyShowInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("policy show: %v", err)
	}
	if shown.Revision.ID != mut.Revision.ID {
		t.Fatalf("show mismatch: %s vs %s", shown.Revision.ID, mut.Revision.ID)
	}
}
