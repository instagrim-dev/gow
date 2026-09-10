package pipeline

import (
	"context"
	"testing"
	"time"
)

// TestIntegrationFrontierGenerateEndToEnd runs the full offline path through the
// M5.1 slice: seed -> cluster -> failure-space -> mine -> challenge (bias-only,
// -> surviving) -> frontier generate, then asserts the generation run completes,
// at least one proposal is produced, it TARGETS the surviving invariant, and its
// claimed structural violation is CODE-VERIFIED (violates), never merely
// asserted by the provider.
func TestIntegrationFrontierGenerateEndToEnd(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = dataDrivenMiner{}
	app.challengerFn = biasOnlyChallenger{}

	problemID, invID, _ := mineOneCandidate(t, ctx, app, dbPath)

	// Drive the candidate to `surviving` (bias-only campaign whose support holds).
	resp, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID})
	if err != nil {
		t.Fatalf("challenge: %v", err)
	}
	if resp.Reports[0].StateAfter != "surviving" {
		t.Fatalf("state after bias-only campaign = %q, want surviving", resp.Reports[0].StateAfter)
	}

	// Generate frontier proposals against the surviving invariant.
	gen, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("frontier generate: %v", err)
	}
	if !gen.Created {
		t.Fatal("expected a newly-created generation")
	}
	if gen.Generation.ProposalCount == 0 || len(gen.Generation.Proposals) == 0 {
		t.Fatal("expected at least one frontier proposal against the surviving invariant")
	}
	p := gen.Generation.Proposals[0]
	// The proposal must TARGET the surviving invariant.
	foundTarget := false
	for _, tgt := range p.Targets {
		if tgt.InvariantID == invID {
			foundTarget = true
			// The break must be code-verified, not merely claimed.
			if tgt.Verdict != "violates" || !tgt.Violated {
				t.Fatalf("target verdict = %q violated=%v; want a code-verified violation", tgt.Verdict, tgt.Violated)
			}
		}
	}
	if !foundTarget {
		t.Fatalf("proposal does not target the surviving invariant %s; targets=%+v", invID, p.Targets)
	}
	if !p.ViolatesAnyTarget {
		t.Fatal("proposal should be flagged as violating a target")
	}
	if p.CheapestFalsificationPath == "" {
		t.Fatal("every proposal must carry a cheapest falsification path")
	}
	// result must be NULL until M5.2 evaluation.
	if p.Result != "" {
		t.Fatalf("result should be empty (NULL) until evaluation; got %q", p.Result)
	}

	// Run lifecycle reflects success.
	run, err := app.ShowRun(ctx, LookupInput{DBPath: dbPath, ID: gen.Generation.RunID})
	if err != nil {
		t.Fatalf("run show: %v", err)
	}
	if run.Run.Status != "completed" {
		t.Fatalf("frontier run status = %q, want completed", run.Run.Status)
	}

	// list/show round-trip.
	list, err := app.ListFrontier(ctx, FrontierListInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("frontier list: %v", err)
	}
	if len(list.Generations) != 1 || list.Generations[0].ID != gen.Generation.ID {
		t.Fatalf("expected 1 generation on the list surface, got %+v", list.Generations)
	}
	shown, err := app.ShowFrontier(ctx, FrontierShowInput{DBPath: dbPath, GenerationID: gen.Generation.ID})
	if err != nil {
		t.Fatalf("frontier show: %v", err)
	}
	if shown.Generation.ProposalCount != gen.Generation.ProposalCount {
		t.Fatalf("show proposal count %d != generate %d", shown.Generation.ProposalCount, gen.Generation.ProposalCount)
	}
}

// TestIntegrationFrontierNoSurvivorsIsEmpty proves that with NO surviving
// invariant (a freshly mined, unchallenged candidate is `proposed`, not
// surviving) the generator has nothing to attack: the run still completes and
// persists an empty generation rather than fabricating proposals.
func TestIntegrationFrontierNoSurvivorsIsEmpty(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = dataDrivenMiner{}

	problemID, _, _ := mineOneCandidate(t, ctx, app, dbPath)

	gen, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("frontier generate: %v", err)
	}
	if gen.Generation.ProposalCount != 0 {
		t.Fatalf("expected 0 proposals with no surviving invariant, got %d", gen.Generation.ProposalCount)
	}
}
