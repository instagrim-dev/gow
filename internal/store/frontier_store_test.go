package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
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

// F3 boundary: ListProposalContentsByIDs batches its IN-list so an id set larger
// than SQLite's bound-parameter ceiling (historically 999) does not error. Only
// the persisted proposal is returned; the >900 non-existent ids are ignored.
func TestListProposalContentsByIDsBatchesPastParameterLimit(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()
	rec := sampleFrontier(t, st)
	res, err := st.PersistFrontierGeneration(ctx, rec)
	if err != nil {
		t.Fatalf("persist: %v", err)
	}
	realID := res.Record.Proposals[0].ID
	insertTestSignatureContent(t, st, realID, `{"schema_version":"mechanism/v1"}`, "cfp-batch")

	ids := []string{realID}
	for i := 0; i < 2500; i++ { // well past both the 900 batch size and the 999 ceiling
		ids = append(ids, domain.NewFrontierProposalID(time.Now().UTC()))
	}
	rows, err := st.ListProposalContentsByIDs(ctx, ids)
	if err != nil {
		t.Fatalf("ListProposalContentsByIDs must batch past the parameter limit: %v", err)
	}
	if len(rows) != 1 || rows[0].ProposalID != realID || rows[0].SignatureJSON == "" {
		t.Fatalf("want exactly the one persisted proposal with content, got %+v", rows)
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

// TestPersistFrontierGenerationIDByHashCoversDedup is the Finding-1 regression:
// the applied-bias log must be able to attribute a policy-biased re-generation
// to persisted proposal ids EVEN when every proposal deduped onto a pre-existing
// row (ProposalCount == 0). The returned ProposalIDByHash must therefore map the
// re-derived hash to the ORIGINAL persisted proposal id, not be empty.
func TestPersistFrontierGenerationIDByHashCoversDedup(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()
	rec := sampleFrontier(t, st)
	first, err := st.PersistFrontierGeneration(ctx, rec)
	if err != nil {
		t.Fatalf("first persist: %v", err)
	}
	originalID := first.Record.Proposals[0].ID
	hash := first.Record.Proposals[0].ProposalHash

	rec2 := sampleFrontierReusingProblem(t, st, rec)
	second, err := st.PersistFrontierGeneration(ctx, rec2)
	if err != nil {
		t.Fatalf("second persist: %v", err)
	}
	if second.Record.ProposalCount != 0 {
		t.Fatalf("expected all proposals deduped, got %d new", second.Record.ProposalCount)
	}
	// The map must still resolve the deduped hash to the original persisted id.
	got, ok := second.ProposalIDByHash[hash]
	if !ok {
		t.Fatal("ProposalIDByHash must cover deduped proposals (applied-bias log would be empty otherwise)")
	}
	if got != originalID {
		t.Fatalf("deduped hash maps to %s, want original %s", got, originalID)
	}
}

// TestRedundantAttackKeysMatchesRedundancyKeyFormat is the Finding-4 regression:
// two proposals sharing the same (targets, nearest clusters) are the SAME
// directed attack; RedundantAttackKeys must return that key (count >= 2) in the
// exact "t:<ids>,|n:<ids>," format frontier.RedundancyKey produces, so a
// penalize/redundant_attack directive keyed by Derive fires in policy.Apply.
func TestRedundantAttackKeysMatchesRedundancyKeyFormat(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()
	rec := sampleFrontier(t, st)
	if _, err := st.PersistFrontierGeneration(ctx, rec); err != nil {
		t.Fatalf("first persist: %v", err)
	}
	// A second, DISTINCT proposal (different hash so it persists) but the SAME
	// target set + nearest cluster set: same directed attack.
	now := time.Now().UTC()
	second := rec
	second.ID = domain.NewFrontierGenerationRunID(now)
	second.Invocation.ID = domain.NewProviderInvocationID(now)
	prop := rec.Proposals[0]
	prop.ID = domain.NewFrontierProposalID(now)
	prop.ProposalHash = "hash-2" // distinct → actually persists
	second.Proposals = []FrontierProposalRow{prop}
	if _, err := st.PersistFrontierGeneration(ctx, second); err != nil {
		t.Fatalf("second persist: %v", err)
	}

	problemID := rec.ProblemID
	invID := rec.Proposals[0].Targets[0].InvariantID
	clusterID := rec.Proposals[0].NearestClusters[0].ClusterID
	wantKey := "t:" + invID + ",|n:" + clusterID + ","

	keys, err := st.RedundantAttackKeys(ctx, problemID, 2)
	if err != nil {
		t.Fatalf("RedundantAttackKeys: %v", err)
	}
	found := false
	for _, k := range keys {
		if k == wantKey {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected redundant-attack key %q (>=2 proposals), got %v", wantKey, keys)
	}

	// A single proposal must NOT be flagged redundant (count < 2 filtered out).
	single := openMigratedStore(t)
	solo := sampleFrontier(t, single)
	if _, err := single.PersistFrontierGeneration(ctx, solo); err != nil {
		t.Fatalf("solo persist: %v", err)
	}
	soloKeys, err := single.RedundantAttackKeys(ctx, solo.ProblemID, 2)
	if err != nil {
		t.Fatalf("solo RedundantAttackKeys: %v", err)
	}
	if len(soloKeys) != 0 {
		t.Fatalf("a single attack must not be redundant, got %v", soloKeys)
	}
}

// F1 regression: the mechanism fingerprint (and thus the proposal hash)
// excludes evaluation-relevant fields, so a REVISED interpretation dedups onto
// the existing proposal row. The revised content must NOT be silently dropped:
// both interpretations are preserved as immutable content revisions, readers
// consume the LATEST revision, and the content hash distinguishes them.
func TestDedupPreservesRevisedSignatureContent(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()

	first := sampleFrontier(t, st)
	first.Proposals[0].SignatureJSON = `{"schema_version":"mechanism/v1","completeness":"unobserved"}`
	first.Proposals[0].CanonicalFingerprint = "cfp-same"
	res1, err := st.PersistFrontierGeneration(ctx, first)
	if err != nil {
		t.Fatalf("persist first: %v", err)
	}
	proposalID := res1.Record.Proposals[0].ID

	// Same proposal hash (fingerprint-identical), REVISED evaluation-relevant
	// content: extraction completeness changed unobserved -> complete.
	second := sampleFrontierReusingProblem(t, st, first)
	second.Proposals[0].SignatureJSON = `{"schema_version":"mechanism/v1","completeness":"complete"}`
	second.Proposals[0].CanonicalFingerprint = "cfp-same"
	if _, err := st.PersistFrontierGeneration(ctx, second); err != nil {
		t.Fatalf("persist revised: %v", err)
	}

	// Both interpretations preserved as immutable revisions.
	var count int
	if err := st.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM frontier_proposal_signature_revisions WHERE proposal_id = ?`, proposalID).Scan(&count); err != nil {
		t.Fatalf("count revisions: %v", err)
	}
	if count != 2 {
		t.Fatalf("want 2 content revisions (original + revised), got %d", count)
	}

	// Readers consume the LATEST revision (the revised interpretation).
	rows, err := st.ListProposalContentsByIDs(ctx, []string{proposalID})
	if err != nil {
		t.Fatalf("read contents: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("want 1 content row, got %d", len(rows))
	}
	if !strings.Contains(rows[0].SignatureJSON, `"completeness":"complete"`) {
		t.Fatalf("reader must see the REVISED interpretation, got %s", rows[0].SignatureJSON)
	}
	if rows[0].ContentHash == "" {
		t.Fatal("content hash must identify the revision consumed")
	}

	// Idempotent replay of identical content adds NO new revision.
	if _, err := st.PersistFrontierGeneration(ctx, sampleFrontierReusingProblem(t, st, second)); err != nil {
		t.Fatalf("replay: %v", err)
	}
	if err := st.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM frontier_proposal_signature_revisions WHERE proposal_id = ?`, proposalID).Scan(&count); err != nil {
		t.Fatalf("recount: %v", err)
	}
	if count != 2 {
		t.Fatalf("identical replay must not mint a revision; got %d", count)
	}
}

// insertTestSignatureContent writes a proposal's canonical content the way the
// real persist path does post-v23: legacy first-observed sidecar + content
// revision (readers consume the latest revision).
func insertTestSignatureContent(t *testing.T, st *Store, proposalID, signatureJSON, fingerprint string) {
	t.Helper()
	ctx := context.Background()
	at := formatTime(time.Now().UTC())
	if _, err := st.db.ExecContext(ctx, `
INSERT OR IGNORE INTO frontier_proposal_signatures(proposal_id, signature_json, canonical_fingerprint, created_at)
VALUES(?, ?, ?, ?)`, proposalID, signatureJSON, fingerprint, at); err != nil {
		t.Fatalf("insert sidecar: %v", err)
	}
	sum := sha256.Sum256([]byte(signatureJSON))
	if _, err := st.db.ExecContext(ctx, `
INSERT OR IGNORE INTO frontier_proposal_signature_revisions(proposal_id, revision, content_hash, canonical_fingerprint, signature_json, created_at)
SELECT ?, COALESCE(MAX(revision), 0) + 1, ?, ?, ?, ? FROM frontier_proposal_signature_revisions WHERE proposal_id = ?`,
		proposalID, hex.EncodeToString(sum[:]), fingerprint, signatureJSON, at, proposalID); err != nil {
		t.Fatalf("insert revision: %v", err)
	}
}
