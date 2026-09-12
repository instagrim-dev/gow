package pipeline

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/domain"
)

// TestIntegrationReplayPreservesCompatibleCurrentAuthority is the committed
// regression for F-2 (2026-09-12 review run 2, remediation handoff 1): a
// historical replay over an obsolete population must NOT displace authority
// earned against the current compatible population.
//
// Sequence (exactly the review run's C7 positive branch):
//  1. mine a candidate over discovery population A0; author the claim;
//     challenge -> surviving over A0; the candidate is targeted.
//  2. admit one witness-checked (domain-checked) failure whose signature has
//     LOCAL posture, so it PRESERVES the mined claim `locality == local`;
//     recluster -> current population A1.
//  3. re-challenge under `latest` -> surviving over A1 (the compatible
//     current authority); the candidate is targeted again.
//  4. replay under `discovery` -> the bounded historical claim over A0 is
//     reproducible (surviving) and takes the latest state transition.
//  5. the next generation must STILL target the candidate: the compatible A1
//     campaign is the operative current authority; the replay must appear
//     nowhere in ExcludedStaleAuthority for this claim.
//
// The negative control — a replay whose only compatible-population campaign
// earned a DIFFERENT state (weaken) — is asserted by
// TestIntegrationChallengeAssessmentPopulation (5b): there the exclusion must
// stand. Together they pin both halves of the recipe's C7 requirement.
func TestIntegrationReplayPreservesCompatibleCurrentAuthority(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2026, 9, 12, 14, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = localityMiner{}
	app.challengerFn = counterexampleOnlyChallenger{}

	problemID, runID, snapshotID := seedProblemAndSnapshot(t, ctx, dbPath, now)
	seedAndSign(t, ctx, app, dbPath, problemID, runID, snapshotID, "challenge_population_a.json")
	discovery, err := app.BuildClustering(ctx, ClusterBuildInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("cluster build (A0): %v", err)
	}
	if _, err := app.BuildFailureSpace(ctx, FailureSpaceBuildInput{DBPath: dbPath, ProblemID: problemID}); err != nil {
		t.Fatalf("failure space: %v", err)
	}
	mined, err := app.MineInvariants(ctx, InvariantMineInput{DBPath: dbPath, ProblemID: problemID, MinSupport: 2})
	if err != nil || mined.Revision.CandidateCount == 0 {
		t.Fatalf("mine: %v (candidates=%d)", err, mined.Revision.CandidateCount)
	}
	invID := mined.Revision.Candidates[0].ID
	if _, err := app.AuthorInvariantClaim(ctx, AuthorClaimInput{
		DBPath: dbPath, InvariantID: invID, Quantifier: "universal", ClaimRole: "regularity",
		Scope: "failure families of discovery cluster run " + discovery.ClusterRun.ID,
		Note:  "F-2 regression: universal treatment over A0",
	}); err != nil {
		t.Fatalf("author claim: %v", err)
	}
	baseline, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID})
	if err != nil {
		t.Fatalf("baseline challenge: %v", err)
	}
	if baseline.Reports[0].StateAfter != "surviving" {
		t.Fatalf("baseline over A0 must survive, got %q", baseline.Reports[0].StateAfter)
	}

	// One admitted domain-checked failure with LOCAL posture -> population A1.
	local := pcSignature(pcResidueLocality, "", true)
	local.Posture.Locality = domain.LocalityLocal
	app.generatorFn = pcGenerator{signatures: []canon.MechanismSignature{local}}
	gen, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil || len(gen.Generation.Proposals) == 0 {
		t.Fatalf("mint proposal: %v (n=%d)", err, len(gen.Generation.Proposals))
	}
	app.generatorFn = nil
	if _, err := app.WitnessCheck(ctx, WitnessCheckInput{
		DBPath: dbPath, ProposalID: gen.Generation.Proposals[0].ID, Tuple: "7,5,5,5",
		Note: "F-2 regression: equal-denominator probe x=y=z=round(3n/4), n=7; deliberately failing bounded attempt",
	}); err != nil {
		t.Fatalf("witness: %v", err)
	}
	admitted, err := app.AdmitEvidence(ctx, AdmitEvidenceInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil || len(admitted.Admitted) != 1 {
		t.Fatalf("admit: %+v err=%v", admitted, err)
	}
	recluster, err := app.BuildClustering(ctx, ClusterBuildInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil || !recluster.Created {
		t.Fatalf("recluster (A1): created=%v err=%v", recluster.Created, err)
	}

	// Compatible current authority: surviving over A1.
	second, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID, Population: PopulationLatest})
	if err != nil {
		t.Fatalf("reassessment: %v", err)
	}
	if got := second.Reports[0]; got.StateAfter != "surviving" || got.AssessmentClusterRunID != recluster.ClusterRun.ID {
		t.Fatalf("reassessment must survive over A1 %s, got %q over %s", recluster.ClusterRun.ID, got.StateAfter, got.AssessmentClusterRunID)
	}

	// Historical replay over A0 takes the latest transition...
	replay, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID, Population: PopulationDiscovery})
	if err != nil {
		t.Fatalf("replay: %v", err)
	}
	if got := replay.Reports[0]; got.StateAfter != "surviving" || got.AssessmentClusterRunID != discovery.ClusterRun.ID {
		t.Fatalf("replay must reproduce surviving over A0 %s, got %q over %s", discovery.ClusterRun.ID, got.StateAfter, got.AssessmentClusterRunID)
	}

	// ...and must NOT displace the compatible current authority.
	after, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("generate after replay: %v", err)
	}
	for _, e := range after.ExcludedStaleAuthority {
		if e.InvariantID == invID {
			t.Fatalf("F-2 regression: replay displaced compatible current authority: %+v", e)
		}
	}
	repo := openTestStore(t, ctx, dbPath)
	defer repo.Close()
	genRow, err := repo.GetFrontierGeneration(ctx, after.Generation.ID)
	if err != nil {
		t.Fatalf("load generation: %v", err)
	}
	invocs, err := repo.ListProviderInvocationsForRun(ctx, genRow.RunID)
	if err != nil || len(invocs) == 0 {
		t.Fatalf("load invocation: %v (n=%d)", err, len(invocs))
	}
	if !strings.Contains(invocs[0].RequestPayload, invID) {
		t.Fatalf("post-replay generation request must still target %s", invID)
	}

	// The compatible authority the gate consulted is durable and names the
	// reassessment campaign, not the replay.
	compat, found, err := repo.GetLatestCompatibleAuthority(ctx, invID, recluster.ClusterRun.ID)
	if err != nil || !found {
		t.Fatalf("compatible authority lookup: found=%v err=%v", found, err)
	}
	if compat.RunID != second.Reports[0].RunID || compat.ToState != "surviving" {
		t.Fatalf("compatible authority must be the A1 reassessment %s (surviving), got %+v", second.Reports[0].RunID, compat)
	}
}
