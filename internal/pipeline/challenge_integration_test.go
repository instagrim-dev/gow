package pipeline

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/invariant"
	"github.com/instagrim-dev/newf/internal/provider"
)

// biasOnlyChallenger proposes only a bias-critique, so a candidate whose
// support genuinely holds ends the campaign `surviving` deterministically.
type biasOnlyChallenger struct{}

func (biasOnlyChallenger) Challenge(_ context.Context, req provider.ChallengeRequest) (provider.ChallengeResponse, error) {
	return provider.ChallengeResponse{
		Proposals: []provider.ChallengeProposal{{
			Type:           invariant.ChallengeBiasCritique,
			Rationale:      "recount support under redundancy",
			ClaimedVerdict: "support_collapses",
		}},
		Metadata:        provider.Metadata{ProviderName: "fixture", ProviderVersion: "test", ModelName: "bias-only", SchemaVersion: invariant.PredicateSchemaV1},
		RequestPayload:  "req",
		ResponsePayload: "resp",
	}, nil
}

// minPreservesMiner proposes `preserves contains <id>` for the LEXICALLY
// SMALLEST preserves id across all families — insensitive to family order (and
// therefore to the randomness of freshly minted cluster ULIDs), so separate
// stores mine the same predicate.
type minPreservesMiner struct{}

func (minPreservesMiner) Identity() provider.MinerIdentity {
	return provider.MinerIdentity{
		ContractVersion: provider.InvariantMinerVersion,
		ProviderName:    "fixture", ProviderVersion: "v1", ModelName: "min-preserves",
	}
}

func (minPreservesMiner) Mine(_ context.Context, req provider.MiningRequest) (provider.MiningResponse, error) {
	var min string
	for _, f := range req.Families {
		for _, id := range f.Preserves {
			if min == "" || string(id) < min {
				min = string(id)
			}
		}
	}
	var proposals []provider.CandidateProposal
	if min != "" {
		proposals = append(proposals, provider.CandidateProposal{
			Predicate: invariant.Predicate{
				Schema: invariant.PredicateSchemaV1,
				Root:   invariant.Node{Op: invariant.OpContains, Field: invariant.FieldPreserves, CanonicalID: min},
			},
			Statement:        "failed methods preserve " + min,
			AbstractionLevel: "mechanism",
		})
	}
	return provider.MiningResponse{
		Proposals:       proposals,
		Metadata:        provider.Metadata{ProviderName: "fixture", ProviderVersion: "v1", ModelName: "min-preserves", SchemaVersion: invariant.PredicateSchemaV1},
		RequestPayload:  "req",
		ResponsePayload: "resp",
	}, nil
}

// mineOneCandidate seeds the corpus, clusters, builds the failure space, and
// mines one candidate, returning its id plus a snapshot usable as independent
// evidence.
func mineOneCandidate(t *testing.T, ctx context.Context, app *App, dbPath string) (problemID, invariantID, snapshotID string) {
	t.Helper()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	problemID, runID, snapshotID := seedProblemAndSnapshot(t, ctx, dbPath, now)
	seedClusterCorpus(t, ctx, app, dbPath, problemID, runID, snapshotID, canon.VocabularyMechanismV1)
	if _, err := app.BuildClustering(ctx, ClusterBuildInput{DBPath: dbPath, ProblemID: problemID}); err != nil {
		t.Fatalf("cluster build: %v", err)
	}
	if _, err := app.BuildFailureSpace(ctx, FailureSpaceBuildInput{DBPath: dbPath, ProblemID: problemID}); err != nil {
		t.Fatalf("failure-space build: %v", err)
	}
	mined, err := app.MineInvariants(ctx, InvariantMineInput{DBPath: dbPath, ProblemID: problemID, MinSupport: 1})
	if err != nil {
		t.Fatalf("mine: %v", err)
	}
	if mined.Revision.CandidateCount == 0 {
		t.Fatal("expected at least one mined candidate")
	}
	return problemID, mined.Revision.Candidates[0].ID, snapshotID
}

// TestIntegrationChallengeCampaign runs the deriving challenger against a mined
// candidate and asserts: the run completes, every executed challenge is
// persisted with its verified result, the first challenge opens the campaign
// (-> challenged), and the final state is one the trigger-guarded ledger
// reached (never proposed, never established via a provider).
func TestIntegrationChallengeCampaign(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = dataDrivenMiner{}

	problemID, invID, _ := mineOneCandidate(t, ctx, app, dbPath)

	resp, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID})
	if err != nil {
		t.Fatalf("challenge: %v", err)
	}
	if len(resp.Reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(resp.Reports))
	}
	report := resp.Reports[0]
	if report.StateBefore != "proposed" {
		t.Fatalf("state before = %q, want proposed", report.StateBefore)
	}
	if report.StateAfter == "proposed" || report.StateAfter == "established" {
		t.Fatalf("state after = %q: a campaign must move past proposed and can never reach established", report.StateAfter)
	}
	if len(report.Challenges) < 3 {
		t.Fatalf("expected >=3 executed challenges, got %d", len(report.Challenges))
	}
	if len(report.Challenges[0].Transitions) == 0 || report.Challenges[0].Transitions[0] != "challenged" {
		t.Fatalf("first challenge must open the campaign, got %+v", report.Challenges[0].Transitions)
	}
	// Run lifecycle reflects the real outcome.
	run, err := app.ShowRun(ctx, LookupInput{DBPath: dbPath, ID: report.RunID})
	if err != nil {
		t.Fatalf("run show: %v", err)
	}
	if run.Run.Status != "completed" {
		t.Fatalf("challenge run status = %q, want completed", run.Run.Status)
	}
	// State is queryable and history round-trips.
	state, err := app.ShowInvariantState(ctx, InvariantStateInput{DBPath: dbPath, InvariantID: invID})
	if err != nil {
		t.Fatalf("invariant state: %v", err)
	}
	if state.Invariant.State != report.StateAfter {
		t.Fatalf("state view %q != report %q", state.Invariant.State, report.StateAfter)
	}
	if len(state.Challenges) != len(report.Challenges) {
		t.Fatalf("history %d != campaign %d", len(state.Challenges), len(report.Challenges))
	}
	// A falsified candidate is terminal: a second campaign is refused.
	if report.StateAfter == "falsified" {
		if _, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID}); err == nil {
			t.Fatal("falsified must be terminal")
		}
	}
	_ = problemID
}

// TestIntegrationEstablishedIsCodeGated proves KTD-2 end to end: a candidate
// that SURVIVES its campaign cannot be established by model judgment (no
// snapshot -> refused), only by operator-supplied independent evidence; and a
// non-surviving candidate is refused regardless.
func TestIntegrationEstablishedIsCodeGated(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = dataDrivenMiner{}
	app.challengerFn = biasOnlyChallenger{}

	problemID, invID, snapshotID := mineOneCandidate(t, ctx, app, dbPath)

	// A proposed (unchallenged) candidate cannot be established at all.
	if _, err := app.EstablishInvariant(ctx, EstablishInput{DBPath: dbPath, InvariantID: invID, SnapshotID: "snap_x", Locator: "l"}); err == nil {
		t.Fatal("only a surviving invariant may be established")
	}

	// bias-only campaign whose support holds -> surviving.
	resp, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID})
	if err != nil {
		t.Fatalf("challenge: %v", err)
	}
	if resp.Reports[0].StateAfter != "surviving" {
		t.Fatalf("state after bias-only campaign = %q, want surviving", resp.Reports[0].StateAfter)
	}

	// Surviving + NO independent evidence -> refused (the gate).
	_, err = app.EstablishInvariant(ctx, EstablishInput{DBPath: dbPath, InvariantID: invID})
	if err == nil || !strings.Contains(err.Error(), "independent evidence") {
		t.Fatalf("expected independent-evidence refusal, got %v", err)
	}

	// Surviving + a real persisted snapshot -> established.
	est, err := app.EstablishInvariant(ctx, EstablishInput{
		DBPath: dbPath, InvariantID: invID,
		SnapshotID: snapshotID, Locator: "sec. 3, theorem 2", Note: "independent proof",
	})
	if err != nil {
		t.Fatalf("establish: %v", err)
	}
	if est.Invariant.State != "established" {
		t.Fatalf("state = %q, want established", est.Invariant.State)
	}

	// The frontier read surface exposes it under --state established.
	list, err := app.ListInvariantStates(ctx, InvariantStatesInput{DBPath: dbPath, ProblemID: problemID, State: "established"})
	if err != nil {
		t.Fatalf("list established: %v", err)
	}
	if len(list.Invariants) != 1 || list.Invariants[0].InvariantID != invID {
		t.Fatalf("expected the established candidate on the read surface, got %+v", list.Invariants)
	}
}

// TestIntegrationChallengeAllIsDeterministic runs --all twice on identical
// stores and asserts identical final states in identical candidate order
// (KTD-4: campaign determinism / replayability).
func TestIntegrationChallengeAllIsDeterministic(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)

	runCampaign := func() []string {
		app, dbPath := newRealStoreApp(t, now)
		app.invariantMinerFn = minPreservesMiner{}
		problemID, _, _ := mineOneCandidate(t, ctx, app, dbPath)
		resp, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, ProblemID: problemID, All: true})
		if err != nil {
			t.Fatalf("challenge --all: %v", err)
		}
		var out []string
		for _, r := range resp.Reports {
			out = append(out, r.StateAfter)
			for _, ch := range r.Challenges {
				out = append(out, ch.ChallengeType+"="+ch.ResultSummary)
			}
		}
		return out
	}

	a, b := runCampaign(), runCampaign()
	if len(a) != len(b) {
		t.Fatalf("campaign shapes differ: %v vs %v", a, b)
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("campaign not deterministic at %d: %q vs %q", i, a[i], b[i])
		}
	}
}
