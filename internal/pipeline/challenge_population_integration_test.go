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

// localityMiner proposes the enum-axis predicate `locality equals local`. Axis
// predicates evaluate decisively on every signature (no completeness gate), so
// a later-ingested global-locality failure is a DECISIVE violator — exactly
// the contradictory evidence the v34/S1 regression needs.
type localityMiner struct{}

func (localityMiner) Identity() provider.MinerIdentity {
	return provider.MinerIdentity{
		ContractVersion: provider.InvariantMinerVersion,
		ProviderName:    "fixture", ProviderVersion: "v1", ModelName: "locality-equals-local",
	}
}

func (localityMiner) Mine(_ context.Context, _ provider.MiningRequest) (provider.MiningResponse, error) {
	return provider.MiningResponse{
		Proposals: []provider.CandidateProposal{{
			Predicate: invariant.Predicate{
				Schema: invariant.PredicateSchemaV1,
				Root:   invariant.Node{Op: invariant.OpEquals, Field: invariant.FieldLocality, Values: []string{"local"}},
			},
			Statement:        "failed methods are local",
			AbstractionLevel: "mechanism",
		}},
		Metadata:        provider.Metadata{ProviderName: "fixture", ProviderVersion: "v1", ModelName: "locality-equals-local", SchemaVersion: invariant.PredicateSchemaV1},
		RequestPayload:  "req",
		ResponsePayload: "resp",
	}, nil
}

// counterexampleOnlyChallenger proposes a single known-counterexample search.
// Verification is code-owned over the routed population, so this is the
// minimal attack that exposes WHICH evidence the campaign actually searched.
type counterexampleOnlyChallenger struct{}

func (counterexampleOnlyChallenger) Challenge(_ context.Context, _ provider.ChallengeRequest) (provider.ChallengeResponse, error) {
	return provider.ChallengeResponse{
		Proposals: []provider.ChallengeProposal{{
			Type:           invariant.ChallengeKnownCounterexample,
			Rationale:      "search the atlas for a violating failed approach",
			ClaimedVerdict: "violates",
		}},
		Metadata:        provider.Metadata{ProviderName: "fixture", ProviderVersion: "test", ModelName: "counterexample-only", SchemaVersion: invariant.PredicateSchemaV1},
		RequestPayload:  "req",
		ResponsePayload: "resp",
	}, nil
}

// seedAndSign loads a mechanism fixture and signs every mechanism it produces.
func seedAndSign(t *testing.T, ctx context.Context, app *App, dbPath, problemID, runID, snapshotID, fixture string) {
	t.Helper()
	seed, err := app.SeedMechanismFixture(ctx, MechanismFixtureSeedInput{
		DBPath: dbPath, ProblemID: problemID, RunID: runID, SnapshotID: snapshotID,
		Path: fixturePath(fixture),
	})
	if err != nil {
		t.Fatalf("seed %s: %v", fixture, err)
	}
	for _, mechID := range seed.MechanismIDs {
		if _, err := app.SignatureMechanism(ctx, SignatureInput{DBPath: dbPath, MechanismID: mechID, VocabVersion: canon.VocabularyMechanismV1}); err != nil {
			t.Fatalf("signature %s: %v", mechID, err)
		}
	}
}

// TestIntegrationChallengeAssessmentPopulation is the v34/S1 regression the
// 2026-09-12 structural review demanded: derive a claim under population A,
// add contradictory evidence B, and assess the existing claim under A+B.
//
//  1. Population A: two local-locality failures. Mining `locality equals
//     local` yields a candidate with FULL failure coverage over A — the
//     universal-counterexample treatment applies.
//  2. A first campaign under the default `latest` policy finds no newer
//     cluster run: assessment == discovery (recorded as such), the
//     counterexample search completes negative, and the candidate survives.
//  3. Evidence B: a global-locality failure is ingested, signed, and
//     reclustered into a NEW cluster run (A+B).
//  4. A re-challenge under `latest` runs the counterexample search against
//     A+B and finds the violator — but the violator is OUTSIDE the discovery
//     population, so the historical claim about A is preserved: the candidate
//     transitions to `weaken` (generalization refuted, search eligibility
//     dropped), NOT `falsified`. The campaign records both population ids.
//  5. A `discovery` replay still sees only A: the violator is invisible and
//     the search completes negative, reproducing the historical assessment.
func TestIntegrationChallengeAssessmentPopulation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = localityMiner{}
	app.challengerFn = counterexampleOnlyChallenger{}

	problemID, runID, snapshotID := seedProblemAndSnapshot(t, ctx, dbPath, now)

	// 1. Population A -> cluster -> failure space -> mine.
	seedAndSign(t, ctx, app, dbPath, problemID, runID, snapshotID, "challenge_population_a.json")
	discovery, err := app.BuildClustering(ctx, ClusterBuildInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("cluster build (A): %v", err)
	}
	if _, err := app.BuildFailureSpace(ctx, FailureSpaceBuildInput{DBPath: dbPath, ProblemID: problemID}); err != nil {
		t.Fatalf("failure-space build: %v", err)
	}
	mined, err := app.MineInvariants(ctx, InvariantMineInput{DBPath: dbPath, ProblemID: problemID, MinSupport: 2})
	if err != nil {
		t.Fatalf("mine: %v", err)
	}
	if mined.Revision.CandidateCount == 0 {
		t.Fatal("expected a mined candidate")
	}
	cand := mined.Revision.Candidates[0]
	invID := cand.ID
	if cand.FailureCoverageDen == 0 || cand.FailureCoverageNum != cand.FailureCoverageDen {
		t.Fatalf("regression setup requires FULL failure coverage over A, got %d/%d", cand.FailureCoverageNum, cand.FailureCoverageDen)
	}

	// 2. No newer population exists: latest degrades to the discovery run and
	// says so; the negative search over A earns survival.
	first, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID})
	if err != nil {
		t.Fatalf("first challenge: %v", err)
	}
	r := first.Reports[0]
	if r.PopulationPolicy != PopulationLatest {
		t.Fatalf("default population policy = %q, want %q", r.PopulationPolicy, PopulationLatest)
	}
	if r.DiscoveryClusterRunID != discovery.ClusterRun.ID || r.AssessmentClusterRunID != discovery.ClusterRun.ID {
		t.Fatalf("with no newer run, assessment must equal discovery (%s): got discovery=%s assessment=%s",
			discovery.ClusterRun.ID, r.DiscoveryClusterRunID, r.AssessmentClusterRunID)
	}
	if r.StateAfter != "surviving" {
		t.Fatalf("negative search over A should earn surviving, got %q", r.StateAfter)
	}

	// 3. Contradictory evidence B enters the atlas and a NEW cluster run.
	seedAndSign(t, ctx, app, dbPath, problemID, runID, snapshotID, "challenge_population_b.json")
	recluster, err := app.BuildClustering(ctx, ClusterBuildInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("cluster build (A+B): %v", err)
	}
	if !recluster.Created || recluster.ClusterRun.ID == discovery.ClusterRun.ID {
		t.Fatal("regression setup requires a NEW cluster run for A+B")
	}

	// 4. Re-challenge under `latest`: the violator from B is found, but it is
	// outside the discovery scope — weaken, never a retroactive falsification.
	second, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID, Population: PopulationLatest})
	if err != nil {
		t.Fatalf("second challenge: %v", err)
	}
	r = second.Reports[0]
	if r.DiscoveryClusterRunID != discovery.ClusterRun.ID {
		t.Fatalf("discovery run = %s, want %s", r.DiscoveryClusterRunID, discovery.ClusterRun.ID)
	}
	if r.AssessmentClusterRunID != recluster.ClusterRun.ID {
		t.Fatalf("assessment run = %s, want the A+B run %s", r.AssessmentClusterRunID, recluster.ClusterRun.ID)
	}
	if r.StateAfter != "weaken" {
		t.Fatalf("out-of-scope counterexample must WEAKEN (bound generalization), got %q", r.StateAfter)
	}
	if len(r.Challenges) != 1 || r.Challenges[0].ResultSummary != "confirmed" {
		t.Fatalf("expected one confirmed counterexample search, got %+v", r.Challenges)
	}
	if !strings.Contains(r.Challenges[0].Detail, "outside the discovery population") {
		t.Fatalf("challenge detail must record the scope classification, got %q", r.Challenges[0].Detail)
	}
	// v36/S5: the confirmed counterexample carries a TYPED boundary delta — the
	// separating condition in canonical predicate form — end-to-end through the
	// report view, not only in the transcript detail.
	if d := r.Challenges[0].BoundaryDelta; d == nil || d.Kind != "counterexample-separation" ||
		d.Condition == "" || d.PredicateFingerprint == "" {
		t.Fatalf("confirmed counterexample must surface a typed boundary delta: %+v", r.Challenges[0].BoundaryDelta)
	}

	// The campaign's population identity is durable, not just reported.
	repo := openTestStore(t, ctx, dbPath)
	defer repo.Close()
	row, found, err := repo.GetChallengeAssessmentPopulation(ctx, r.RunID, invID)
	if err != nil || !found {
		t.Fatalf("assessment population row missing (found=%v, err=%v)", found, err)
	}
	if row.DiscoveryClusterRunID != discovery.ClusterRun.ID || row.AssessmentClusterRunID != recluster.ClusterRun.ID || row.PopulationPolicy != PopulationLatest {
		t.Fatalf("persisted population identity wrong: %+v", row)
	}

	// 5. Historical replay: under `discovery` the violator is invisible and
	// the search over A completes negative — the historical statement about A
	// is reproducible, not rewritten.
	replay, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID, Population: PopulationDiscovery})
	if err != nil {
		t.Fatalf("discovery replay: %v", err)
	}
	r = replay.Reports[0]
	if r.AssessmentClusterRunID != discovery.ClusterRun.ID {
		t.Fatalf("discovery replay must assess against %s, got %s", discovery.ClusterRun.ID, r.AssessmentClusterRunID)
	}
	if got := r.Challenges[0].ResultSummary; got != "unconfirmed" {
		t.Fatalf("replay over A must find no counterexample, got %q", got)
	}
}

// TestIntegrationMixedPopulationAssociationNotFalsified pins the association
// discipline under the new default assessment population (v34/S1): when the
// discovery population itself contains the violator, coverage is PARTIAL, the
// claim is an association (not a universal), and an isolated in-atlas violator
// is recorded as evidence without falsifying — the G3 rule is unchanged by
// population routing. (A universal claim with an in-scope violator remains the
// falsification path; with deterministic predicates and full mining-time
// coverage such a violator cannot arise within the same recorded population,
// which is exactly what the scope classification preserves.)
func TestIntegrationMixedPopulationAssociationNotFalsified(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = localityMiner{}
	app.challengerFn = counterexampleOnlyChallenger{}

	problemID, runID, snapshotID := seedProblemAndSnapshot(t, ctx, dbPath, now)

	// Discovery population contains BOTH local failures and the global
	// violator: `locality equals local` is mined WITHOUT full failure
	// coverage, so the universal-counterexample treatment must not apply.
	seedAndSign(t, ctx, app, dbPath, problemID, runID, snapshotID, "challenge_population_a.json")
	seedAndSign(t, ctx, app, dbPath, problemID, runID, snapshotID, "challenge_population_b.json")
	if _, err := app.BuildClustering(ctx, ClusterBuildInput{DBPath: dbPath, ProblemID: problemID}); err != nil {
		t.Fatalf("cluster build: %v", err)
	}
	if _, err := app.BuildFailureSpace(ctx, FailureSpaceBuildInput{DBPath: dbPath, ProblemID: problemID}); err != nil {
		t.Fatalf("failure-space build: %v", err)
	}
	mined, err := app.MineInvariants(ctx, InvariantMineInput{DBPath: dbPath, ProblemID: problemID, MinSupport: 2})
	if err != nil {
		t.Fatalf("mine: %v", err)
	}
	if mined.Revision.CandidateCount == 0 {
		t.Skip("miner produced no candidate over the mixed population (coverage gate)")
	}
	cand := mined.Revision.Candidates[0]
	if cand.FailureCoverageNum == cand.FailureCoverageDen {
		t.Fatalf("setup expects PARTIAL coverage over the mixed population, got %d/%d", cand.FailureCoverageNum, cand.FailureCoverageDen)
	}

	// Partial coverage means the claim is an association, not a universal:
	// the in-atlas violator is recorded as evidence but must NOT falsify.
	resp, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: cand.ID})
	if err != nil {
		t.Fatalf("challenge: %v", err)
	}
	if got := resp.Reports[0].StateAfter; got == "falsified" {
		t.Fatalf("a non-universal claim must not be falsified by an isolated in-scope violator, got %q", got)
	}
}
