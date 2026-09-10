package pipeline

import (
	"context"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/invariant"
	"github.com/instagrim-dev/newf/internal/provider"
)

// dataDrivenMiner proposes one predicate: `preserves contains <id>` for the
// canonical id shared across the request's failure families. It is a
// deterministic, offline stand-in for a model, and it deliberately does NOT
// assert any support — the engine computes support against persisted signatures.
type dataDrivenMiner struct {
	obstruction bool
}

func (m dataDrivenMiner) Mine(_ context.Context, req provider.MiningRequest) (provider.MiningResponse, error) {
	var id domain.CanonicalID
	for _, f := range req.Families {
		if len(f.Preserves) > 0 {
			id = f.Preserves[0]
			break
		}
	}
	var proposals []provider.CandidateProposal
	if id != "" {
		proposals = append(proposals, provider.CandidateProposal{
			Predicate: invariant.Predicate{
				Schema: invariant.PredicateSchemaV1,
				Root:   invariant.Node{Op: invariant.OpContains, Field: invariant.FieldPreserves, CanonicalID: string(id)},
			},
			Statement:             "failed methods preserve " + string(id),
			AbstractionLevel:      "mechanism",
			ObstructionHypothesis: m.obstruction,
		})
	}
	return provider.MiningResponse{
		Proposals:       proposals,
		Metadata:        provider.Metadata{ProviderName: "fixture", ProviderVersion: "v1", ModelName: "deterministic-fixture", SchemaVersion: invariant.PredicateSchemaV1},
		RequestPayload:  "req",
		ResponsePayload: "resp",
	}, nil
}

// TestIntegrationInvariantMineEndToEnd runs the full offline path:
// seed -> signatures -> cluster build -> failure-space build -> invariants mine
// -> invariant show, asserting the mining run is `completed`, at least one
// candidate is produced with code-computed support, and the show/list surfaces
// round-trip.
func TestIntegrationInvariantMineEndToEnd(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = dataDrivenMiner{}

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
		t.Fatalf("MineInvariants: %v", err)
	}
	if !mined.Created {
		t.Fatal("first mine should be created")
	}
	if mined.Revision.CandidateCount == 0 {
		t.Fatal("expected at least one candidate")
	}
	cand := mined.Revision.Candidates[0]
	if cand.State != "proposed" {
		t.Fatalf("candidate state = %q, want proposed", cand.State)
	}
	if cand.DistinctFamilySupport == 0 {
		t.Fatal("expected code-computed support > 0")
	}
	if cand.PredicateFingerprint == "" || cand.Predicate == "" {
		t.Fatal("expected a persisted predicate + fingerprint")
	}

	// The mining run reflects `completed` via `run show`.
	shownRun, err := app.ShowRun(ctx, LookupInput{DBPath: dbPath, ID: mined.Revision.RunID})
	if err != nil {
		t.Fatalf("run show: %v", err)
	}
	if shownRun.Run.Status != "completed" {
		t.Fatalf("mining run status = %q, want completed", shownRun.Run.Status)
	}

	// show + list round-trip.
	shown, err := app.ShowInvariant(ctx, InvariantShowInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("invariant show: %v", err)
	}
	if shown.Revision.ID != mined.Revision.ID {
		t.Fatalf("show returned %s, want %s", shown.Revision.ID, mined.Revision.ID)
	}
	listed, err := app.ListInvariants(ctx, InvariantListInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("invariant list: %v", err)
	}
	if len(listed.Revisions) != 1 {
		t.Fatalf("expected 1 revision, got %d", len(listed.Revisions))
	}
}

// TestIntegrationInvariantMineIdempotentReMine asserts re-mining the same
// failure space under identical versions returns the existing revision (never a
// rewrite), and identical predicate fingerprints.
func TestIntegrationInvariantMineIdempotentReMine(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = dataDrivenMiner{}

	problemID, runID, snapshotID := seedProblemAndSnapshot(t, ctx, dbPath, now)
	seedClusterCorpus(t, ctx, app, dbPath, problemID, runID, snapshotID, canon.VocabularyMechanismV1)
	if _, err := app.BuildClustering(ctx, ClusterBuildInput{DBPath: dbPath, ProblemID: problemID}); err != nil {
		t.Fatalf("cluster build: %v", err)
	}
	if _, err := app.BuildFailureSpace(ctx, FailureSpaceBuildInput{DBPath: dbPath, ProblemID: problemID}); err != nil {
		t.Fatalf("failure-space build: %v", err)
	}

	first, err := app.MineInvariants(ctx, InvariantMineInput{DBPath: dbPath, ProblemID: problemID, MinSupport: 1})
	if err != nil {
		t.Fatalf("first mine: %v", err)
	}
	second, err := app.MineInvariants(ctx, InvariantMineInput{DBPath: dbPath, ProblemID: problemID, MinSupport: 1})
	if err != nil {
		t.Fatalf("second mine: %v", err)
	}
	if !first.Created || second.Created {
		t.Fatalf("expected first created, second idempotent; got %v/%v", first.Created, second.Created)
	}
	if first.Revision.ID != second.Revision.ID {
		t.Fatalf("idempotent mine produced different revisions: %s vs %s", first.Revision.ID, second.Revision.ID)
	}
	if first.Revision.Candidates[0].PredicateFingerprint != second.Revision.Candidates[0].PredicateFingerprint {
		t.Fatal("predicate fingerprint changed across identical mines")
	}

	// A different min-support is a different identity tuple -> new revision.
	third, err := app.MineInvariants(ctx, InvariantMineInput{DBPath: dbPath, ProblemID: problemID, MinSupport: 3})
	if err != nil {
		t.Fatalf("third mine: %v", err)
	}
	if !third.Created {
		t.Fatal("changing min-support must create a new revision")
	}
}
