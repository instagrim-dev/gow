package store

import (
	"context"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/domain"
)

// seedFrontierPrereqs persists an invariant revision (yielding a real candidate
// invariant id for the target FK) and returns the ids a frontier generation
// references.
func seedFrontierPrereqs(t *testing.T, st *Store) (problemID, runID, clusterRunID, clusterID, invariantID string) {
	t.Helper()
	ctx := context.Background()
	rec := sampleRevision(t, st)
	res, err := st.PersistInvariantRevision(ctx, rec)
	if err != nil {
		t.Fatalf("persist invariant revision: %v", err)
	}
	// Recover the cluster id from the seeded family evaluation.
	clusterID = res.Record.Candidates[0].FamilyEvaluations[0].ClusterID
	return res.Record.ProblemID, res.Record.RunID, res.Record.ClusterRunID, clusterID, res.Record.Candidates[0].ID
}

func sampleFrontier(t *testing.T, st *Store) FrontierGenerationRecord {
	t.Helper()
	problemID, runID, clusterRunID, clusterID, invariantID := seedFrontierPrereqs(t, st)
	now := time.Now().UTC()
	return FrontierGenerationRecord{
		ID:               domain.NewFrontierGenerationRunID(now),
		ProblemID:        problemID,
		ClusterRunID:     clusterRunID,
		RunID:            runID,
		GeneratorVersion: "frontier/v1",
		RequestedCount:   4,
		CreatedAt:        formatTime(now),
		Invocation: FrontierProviderInvocation{
			ID: domain.NewProviderInvocationID(now), RunID: runID,
			ProviderName: "fixture", SchemaVersion: "mechanism/v1",
			RequestHash: "rh", CreatedAt: formatTime(now),
		},
		Proposals: []FrontierProposalRow{{
			ID:                        domain.NewFrontierProposalID(now),
			ProposalHash:              "hash-1",
			StructuralViolationClaim:  "introduces a global coupling object",
			NoveltyArgument:           "global auxiliary object absent from every failure family",
			CheapestFalsificationPath: "check whether it degenerates to a residue cover",
			MechanisticDistance:       string(domain.OrdinalHigh),
			ExpectedInformationGain:   string(domain.OrdinalMedium),
			EvaluationCost:            string(domain.OrdinalLow),
			ViolatesAnyTarget:         true,
			Rank:                      0,
			Targets:                   []FrontierTargetRow{{InvariantID: invariantID, Verdict: "violates", Violated: true}},
			NearestClusters:           []FrontierNearestRow{{ClusterID: clusterID, Classification: "mechanism-distinct", Proximity: string(domain.OrdinalLow)}},
		}},
	}
}

func TestPersistFrontierGenerationRoundTrip(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()
	rec := sampleFrontier(t, st)

	res, err := st.PersistFrontierGeneration(ctx, rec)
	if err != nil {
		t.Fatalf("persist: %v", err)
	}
	if !res.Created || res.Record.Revision != 1 || res.Record.ProposalCount != 1 {
		t.Fatalf("expected created rev1 with 1 proposal; got created=%v rev=%d n=%d", res.Created, res.Record.Revision, res.Record.ProposalCount)
	}
	got, err := st.GetFrontierGeneration(ctx, res.Record.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(got.Proposals) != 1 {
		t.Fatalf("want 1 proposal, got %d", len(got.Proposals))
	}
	p := got.Proposals[0]
	if !p.ViolatesAnyTarget || len(p.Targets) != 1 || p.Targets[0].Verdict != "violates" {
		t.Fatalf("target verdict lost: %+v", p.Targets)
	}
	if len(p.NearestClusters) != 1 || p.NearestClusters[0].Classification != "mechanism-distinct" {
		t.Fatalf("nearest cluster lost: %+v", p.NearestClusters)
	}
	if p.MechanisticDistance != string(domain.OrdinalHigh) {
		t.Fatalf("distance = %q, want high", p.MechanisticDistance)
	}
	if p.Result.Valid {
		t.Fatal("result must be NULL until M5.2 evaluation")
	}
}

func TestPersistFrontierGenerationDedupAcrossRuns(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()
	rec := sampleFrontier(t, st)
	if _, err := st.PersistFrontierGeneration(ctx, rec); err != nil {
		t.Fatalf("first persist: %v", err)
	}
	// A second generation for the same problem proposing the same hash: the
	// proposal is deduped (already exists), so the new run persists with 0 new
	// proposals rather than colliding on UNIQUE(problem_id, proposal_hash).
	rec2 := sampleFrontierReusingProblem(t, st, rec)
	res2, err := st.PersistFrontierGeneration(ctx, rec2)
	if err != nil {
		t.Fatalf("second persist: %v", err)
	}
	if !res2.Created {
		t.Fatal("second run should still persist (append-only)")
	}
	if res2.Record.ProposalCount != 0 {
		t.Fatalf("duplicate proposal hash should be deduped; got %d new", res2.Record.ProposalCount)
	}
	if res2.Record.Revision != 2 {
		t.Fatalf("second run revision = %d, want 2", res2.Record.Revision)
	}
}

// sampleFrontierReusingProblem builds a second generation for the SAME problem
// (reusing its ids + a colliding proposal hash) to exercise cross-run dedup.
func sampleFrontierReusingProblem(t *testing.T, st *Store, first FrontierGenerationRecord) FrontierGenerationRecord {
	t.Helper()
	now := time.Now().UTC()
	second := first
	second.ID = domain.NewFrontierGenerationRunID(now)
	second.Invocation.ID = domain.NewProviderInvocationID(now)
	// New proposal row id but the SAME proposal_hash (a re-derived identical attack).
	prop := first.Proposals[0]
	prop.ID = domain.NewFrontierProposalID(now)
	second.Proposals = []FrontierProposalRow{prop}
	return second
}

func TestFrontierProposalOneTimeResultSet(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()
	res, err := st.PersistFrontierGeneration(ctx, sampleFrontier(t, st))
	if err != nil {
		t.Fatalf("persist: %v", err)
	}
	propID := res.Record.Proposals[0].ID

	// M5.2 will set result once: NULL -> a verdict is allowed by the trigger.
	if _, err := st.db.ExecContext(ctx, `UPDATE frontier_proposals SET result = 'partial_success' WHERE id = ?`, propID); err != nil {
		t.Fatalf("one-time result set should be allowed: %v", err)
	}
	// Overwriting a non-null result must abort.
	if _, err := st.db.ExecContext(ctx, `UPDATE frontier_proposals SET result = 'success' WHERE id = ?`, propID); err == nil {
		t.Fatal("overwriting a non-null result must abort")
	}
	// Changing any other column must abort.
	if _, err := st.db.ExecContext(ctx, `UPDATE frontier_proposals SET novelty_argument = 'tampered' WHERE id = ?`, propID); err == nil {
		t.Fatal("mutating a frontier proposal column must abort")
	}
	// Deletes must abort.
	if _, err := st.db.ExecContext(ctx, `DELETE FROM frontier_proposals WHERE id = ?`, propID); err == nil {
		t.Fatal("deleting a frontier proposal must abort")
	}
}

func TestFrontierGenerationRunImmutable(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()
	res, err := st.PersistFrontierGeneration(ctx, sampleFrontier(t, st))
	if err != nil {
		t.Fatalf("persist: %v", err)
	}
	if _, err := st.db.ExecContext(ctx, `UPDATE frontier_generation_runs SET proposal_count = 99 WHERE id = ?`, res.Record.ID); err == nil {
		t.Fatal("frontier generation runs must be immutable")
	}
}

func TestV14ProviderRoleAllowsGenerate(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()
	tx, err := st.db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("tx: %v", err)
	}
	defer tx.Rollback()
	allows, err := providerRoleAllows(ctx, tx, "generate")
	if err != nil {
		t.Fatalf("providerRoleAllows: %v", err)
	}
	if !allows {
		t.Fatal("v14 must widen provider_invocations.role to permit 'generate'")
	}
}
