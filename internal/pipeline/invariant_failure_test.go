package pipeline

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/provider"
)

// failingMiner simulates a provider transport failure so the run-lifecycle
// failure path is observable.
type failingMiner struct{}

func (failingMiner) Identity() provider.MinerIdentity {
	return provider.MinerIdentity{ContractVersion: provider.InvariantMinerVersion, ProviderName: "fixture", ProviderVersion: "v1", ModelName: "failing-fixture"}
}

func (failingMiner) Mine(context.Context, provider.MiningRequest) (provider.MiningResponse, error) {
	return provider.MiningResponse{}, errors.New("simulated miner transport failure")
}

// runRecordingStore wraps a problemStore to capture created run ids so the
// failure-path test can inspect the run the command minted internally.
type runRecordingStore struct {
	problemStore
	runIDs *[]string
}

func (s runRecordingStore) CreateRun(ctx context.Context, run domain.NewRun) (domain.Run, error) {
	created, err := s.problemStore.CreateRun(ctx, run)
	if err == nil {
		*s.runIDs = append(*s.runIDs, created.ID)
	}
	return created, err
}

// TestIntegrationInvariantMineFailureMarksRunFailed forces a miner error and
// asserts the mining run is `failed` with the cause as its summary via
// `run show` — the run lifecycle must reflect the real command outcome, not the
// status chosen before work happened.
func TestIntegrationInvariantMineFailureMarksRunFailed(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = failingMiner{}

	problemID, runID, snapshotID := seedProblemAndSnapshot(t, ctx, dbPath, now)
	seedClusterCorpus(t, ctx, app, dbPath, problemID, runID, snapshotID, canon.VocabularyMechanismV1)
	if _, err := app.BuildClustering(ctx, ClusterBuildInput{DBPath: dbPath, ProblemID: problemID}); err != nil {
		t.Fatalf("cluster build: %v", err)
	}
	if _, err := app.BuildFailureSpace(ctx, FailureSpaceBuildInput{DBPath: dbPath, ProblemID: problemID}); err != nil {
		t.Fatalf("failure-space build: %v", err)
	}

	// Record run ids created during the failing mine.
	var minedRunIDs []string
	baseOpen := app.openStoreFn
	app.openStoreFn = func(ctx context.Context, dbPath string) (string, problemStore, error) {
		path, st, err := baseOpen(ctx, dbPath)
		if err != nil {
			return path, st, err
		}
		return path, runRecordingStore{problemStore: st, runIDs: &minedRunIDs}, nil
	}

	if _, err := app.MineInvariants(ctx, InvariantMineInput{DBPath: dbPath, ProblemID: problemID}); err == nil {
		t.Fatal("expected miner failure to surface as an error")
	}
	if len(minedRunIDs) != 1 {
		t.Fatalf("expected exactly 1 mining run created, got %d", len(minedRunIDs))
	}

	// The freshly-created mining run must be `failed` with the cause recorded.
	app.openStoreFn = baseOpen
	shown, err := app.ShowRun(ctx, LookupInput{DBPath: dbPath, ID: minedRunIDs[0]})
	if err != nil {
		t.Fatalf("run show: %v", err)
	}
	if shown.Run.Status != "failed" {
		t.Fatalf("mining run status = %q, want failed", shown.Run.Status)
	}
	if shown.Run.ErrorSummary == nil || *shown.Run.ErrorSummary == "" {
		t.Fatal("failed mining run must record an error summary")
	}
}
