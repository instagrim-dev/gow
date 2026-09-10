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

func (dataDrivenMiner) Identity() provider.MinerIdentity {
	return provider.MinerIdentity{
		ContractVersion: provider.InvariantMinerVersion,
		ProviderName:    "fixture", ProviderVersion: "v1", ModelName: "deterministic-fixture",
	}
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

// outcomeReadingMiner proposes a predicate that reads the outcome axis. Such a
// predicate matches the support cohort's own labels by definition and must be
// rejected at mine-time (F2), failing the run rather than being stored.
type outcomeReadingMiner struct{}

func (outcomeReadingMiner) Identity() provider.MinerIdentity {
	return provider.MinerIdentity{ContractVersion: provider.InvariantMinerVersion, ProviderName: "fixture", ProviderVersion: "v1", ModelName: "deterministic-fixture"}
}

func (outcomeReadingMiner) Mine(_ context.Context, _ provider.MiningRequest) (provider.MiningResponse, error) {
	return provider.MiningResponse{
		Proposals: []provider.CandidateProposal{{
			Predicate: invariant.Predicate{
				Schema: invariant.PredicateSchemaV1,
				Root:   invariant.Node{Op: invariant.OpIn, Field: invariant.FieldOutcome, Values: []string{"failure", "partial_failure"}},
			},
			Statement:        "failures have a failure outcome",
			AbstractionLevel: "mechanism",
		}},
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

// unknownTermMiner proposes a predicate whose canonical_id is syntactically
// valid but absent from the pinned vocabulary. ValidateReferences must reject
// it at mine-time so it cannot evaluate to a permanent violates/unknown (F3).
type unknownTermMiner struct{}

func (unknownTermMiner) Identity() provider.MinerIdentity {
	return provider.MinerIdentity{ContractVersion: provider.InvariantMinerVersion, ProviderName: "fixture", ProviderVersion: "v1", ModelName: "deterministic-fixture"}
}

func (unknownTermMiner) Mine(_ context.Context, _ provider.MiningRequest) (provider.MiningResponse, error) {
	return provider.MiningResponse{
		Proposals: []provider.CandidateProposal{{
			Predicate: invariant.Predicate{
				Schema: invariant.PredicateSchemaV1,
				Root:   invariant.Node{Op: invariant.OpContains, Field: invariant.FieldPreserves, CanonicalID: "domain.number_theory.property.nonexistent_term"},
			},
			Statement:        "failures preserve a term that does not exist",
			AbstractionLevel: "mechanism",
		}},
		Metadata:        provider.Metadata{ProviderName: "fixture", ProviderVersion: "v1", ModelName: "deterministic-fixture", SchemaVersion: invariant.PredicateSchemaV1},
		RequestPayload:  "req",
		ResponsePayload: "resp",
	}, nil
}

// countingMiner wraps dataDrivenMiner with a call counter and a configurable
// identity, so tests can assert reuse-check-first (identical request calls the
// provider once) and config-sensitivity (a changed identity is a new revision).
type countingMiner struct {
	inner    dataDrivenMiner
	identity provider.MinerIdentity
	calls    *int
}

func (m countingMiner) Identity() provider.MinerIdentity { return m.identity }

func (m countingMiner) Mine(ctx context.Context, req provider.MiningRequest) (provider.MiningResponse, error) {
	*m.calls++
	return m.inner.Mine(ctx, req)
}

// TestIntegrationInvariantMineReuseChecksBeforeInvoking asserts the F5 contract:
// an identical re-mine reuses the prior revision WITHOUT calling the provider a
// second time (reuse-check-first), while changing the miner identity
// (provider/model/config) produces a new revision and does call the provider.
func TestIntegrationInvariantMineReuseChecksBeforeInvoking(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)

	problemID, runID, snapshotID := seedProblemAndSnapshot(t, ctx, dbPath, now)
	seedClusterCorpus(t, ctx, app, dbPath, problemID, runID, snapshotID, canon.VocabularyMechanismV1)
	if _, err := app.BuildClustering(ctx, ClusterBuildInput{DBPath: dbPath, ProblemID: problemID}); err != nil {
		t.Fatalf("cluster build: %v", err)
	}
	if _, err := app.BuildFailureSpace(ctx, FailureSpaceBuildInput{DBPath: dbPath, ProblemID: problemID}); err != nil {
		t.Fatalf("failure-space build: %v", err)
	}

	calls := 0
	idA := provider.MinerIdentity{ContractVersion: provider.InvariantMinerVersion, ProviderName: "prov-a", ProviderVersion: "v1", ModelName: "model-a"}
	app.invariantMinerFn = countingMiner{identity: idA, calls: &calls}

	first, err := app.MineInvariants(ctx, InvariantMineInput{DBPath: dbPath, ProblemID: problemID, MinSupport: 1})
	if err != nil {
		t.Fatalf("first mine: %v", err)
	}
	if !first.Created {
		t.Fatal("first mine should create a revision")
	}
	if calls != 1 {
		t.Fatalf("first mine should call provider once, got %d", calls)
	}

	// Identical request + identical identity: reuse WITHOUT invoking the provider.
	second, err := app.MineInvariants(ctx, InvariantMineInput{DBPath: dbPath, ProblemID: problemID, MinSupport: 1})
	if err != nil {
		t.Fatalf("second mine: %v", err)
	}
	if second.Created {
		t.Fatal("identical re-mine must reuse (Created=false)")
	}
	if second.Revision.ID != first.Revision.ID {
		t.Fatalf("reuse returned a different revision: %s vs %s", second.Revision.ID, first.Revision.ID)
	}
	if calls != 1 {
		t.Fatalf("reuse-check-first must NOT re-invoke the provider, got %d calls", calls)
	}

	// Changed model (a different identity) is a different reuse key: new revision,
	// and the provider IS invoked.
	idB := idA
	idB.ModelName = "model-b"
	app.invariantMinerFn = countingMiner{identity: idB, calls: &calls}
	third, err := app.MineInvariants(ctx, InvariantMineInput{DBPath: dbPath, ProblemID: problemID, MinSupport: 1})
	if err != nil {
		t.Fatalf("third mine: %v", err)
	}
	if !third.Created {
		t.Fatal("changed miner identity must create a new revision")
	}
	if third.Revision.ID == first.Revision.ID {
		t.Fatal("changed identity must not reuse the prior revision")
	}
	if calls != 2 {
		t.Fatalf("changed identity must invoke the provider, got %d calls", calls)
	}
}
func TestIntegrationInvariantMineRejectsOutcomeReadingProposal(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = outcomeReadingMiner{}

	problemID, runID, snapshotID := seedProblemAndSnapshot(t, ctx, dbPath, now)
	seedClusterCorpus(t, ctx, app, dbPath, problemID, runID, snapshotID, canon.VocabularyMechanismV1)
	if _, err := app.BuildClustering(ctx, ClusterBuildInput{DBPath: dbPath, ProblemID: problemID}); err != nil {
		t.Fatalf("cluster build: %v", err)
	}
	if _, err := app.BuildFailureSpace(ctx, FailureSpaceBuildInput{DBPath: dbPath, ProblemID: problemID}); err != nil {
		t.Fatalf("failure-space build: %v", err)
	}

	_, err := app.MineInvariants(ctx, InvariantMineInput{DBPath: dbPath, ProblemID: problemID, MinSupport: 1})
	if err == nil {
		t.Fatal("expected MineInvariants to reject an outcome-reading proposal")
	}
	// No revision should have been persisted.
	listed, lerr := app.ListInvariants(ctx, InvariantListInput{DBPath: dbPath, ProblemID: problemID})
	if lerr != nil {
		t.Fatalf("invariant list: %v", lerr)
	}
	if len(listed.Revisions) != 0 {
		t.Fatalf("rejected mine must persist no revision, got %d", len(listed.Revisions))
	}
}

// TestIntegrationInvariantMineRejectsUnknownVocabularyTerm proves the F3
// reference-resolution contract end to end: a proposal referencing a term not in
// the pinned vocabulary fails the run, and nothing is persisted.
func TestIntegrationInvariantMineRejectsUnknownVocabularyTerm(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = unknownTermMiner{}

	problemID, runID, snapshotID := seedProblemAndSnapshot(t, ctx, dbPath, now)
	seedClusterCorpus(t, ctx, app, dbPath, problemID, runID, snapshotID, canon.VocabularyMechanismV1)
	if _, err := app.BuildClustering(ctx, ClusterBuildInput{DBPath: dbPath, ProblemID: problemID}); err != nil {
		t.Fatalf("cluster build: %v", err)
	}
	if _, err := app.BuildFailureSpace(ctx, FailureSpaceBuildInput{DBPath: dbPath, ProblemID: problemID}); err != nil {
		t.Fatalf("failure-space build: %v", err)
	}

	_, err := app.MineInvariants(ctx, InvariantMineInput{DBPath: dbPath, ProblemID: problemID, MinSupport: 1})
	if err == nil {
		t.Fatal("expected MineInvariants to reject a proposal referencing an unknown vocabulary term")
	}
	listed, lerr := app.ListInvariants(ctx, InvariantListInput{DBPath: dbPath, ProblemID: problemID})
	if lerr != nil {
		t.Fatalf("invariant list: %v", lerr)
	}
	if len(listed.Revisions) != 0 {
		t.Fatalf("rejected mine must persist no revision, got %d", len(listed.Revisions))
	}
}
