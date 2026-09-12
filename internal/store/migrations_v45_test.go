package store

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
)

// TestMigrateV45RebuildsPerProposalMarkerTable proves the F-1 repair on a REAL
// pre-v45 database: a store whose evaluated_failures marker is keyed per
// proposal (the shadowing shape) is rebuilt to the per-evaluation key, existing
// marker rows survive, a second evaluation of the same proposal becomes
// insertable, and the immutability triggers are restored against the rebuilt
// table.
func TestMigrateV45RebuildsPerProposalMarkerTable(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "newf.db")
	s, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer s.Close()
	if err := s.Migrate(ctx); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}

	// Downgrade to the pre-v45 shape: per-proposal PK, no proposal index, and
	// un-stamp v45 so Migrate replays it. Triggers first (they abort deletes).
	downgrade := []string{
		`DROP TRIGGER IF EXISTS evaluated_failures_immutable_update`,
		`DROP TRIGGER IF EXISTS evaluated_failures_immutable_delete`,
		`DROP INDEX IF EXISTS idx_evaluated_failures_proposal`,
		`DROP TABLE evaluated_failures`,
		`CREATE TABLE evaluated_failures (
  proposal_id TEXT PRIMARY KEY REFERENCES frontier_proposals(id),
  evaluation_id TEXT NOT NULL REFERENCES evaluations(id),
  problem_id TEXT NOT NULL REFERENCES problems(id),
  verdict TEXT NOT NULL CHECK (verdict IN ('failure','partial_failure')),
  created_at TEXT NOT NULL
)`,
		`DELETE FROM schema_migrations WHERE version >= 45`,
	}
	for _, ddl := range downgrade {
		if _, err := s.db.ExecContext(ctx, ddl); err != nil {
			t.Fatalf("downgrade DDL error: %v\n%s", err, ddl)
		}
	}
	// Seed one legacy marker row with a REAL consistent parent chain (Migrate's
	// pre-commit foreign_key_check verifies every reference), then the marker.
	if _, err := s.db.ExecContext(ctx, `PRAGMA foreign_keys = OFF`); err != nil {
		t.Fatalf("pragma: %v", err)
	}
	seed := []string{
		`INSERT INTO runs(id, problem_id, parent_run_id, operation, status, input_ref, tool_name, tool_version, started_at, completed_at)
VALUES('run_legacy','prob_legacy',NULL,'test','succeeded','x','newf','test','t0','t0')`,
		`INSERT INTO problems(id, slug, statement, status, created_at, created_by_run_id)
VALUES('prob_legacy','legacy-1','s','open','t0','run_legacy')`,
		`INSERT INTO canonical_vocabulary(version, created_at, notes) VALUES('mechanism/v1','t0','')`,
		`INSERT INTO cluster_runs(id, problem_id, run_id, schema_version, vocabulary_version, profile_version, cluster_algo_version, thresholds_hash, signature_count, family_count, status, created_at)
VALUES('clr_legacy','prob_legacy','run_legacy','1','mechanism/v1','1','1','h',0,0,'clean','t0')`,
		`INSERT INTO provider_invocations(id, run_id, role, provider_name, schema_version, request_hash, created_at)
VALUES('pin_legacy','run_legacy','normalize','fixture','1','h','t0')`,
		`INSERT INTO frontier_generation_runs(id, problem_id, cluster_run_id, run_id, provider_invocation_id, generator_version, requested_count, proposal_count, revision, created_at)
VALUES('fgr_legacy','prob_legacy','clr_legacy','run_legacy','pin_legacy','1',1,1,1,'t0')`,
		`INSERT INTO frontier_proposals(id, problem_id, frontier_generation_run_id, proposal_hash, structural_violation_claim, novelty_argument, cheapest_falsification_path, mechanistic_distance_ordinal, expected_information_gain_ordinal, evaluation_cost_ordinal, violates_any_target, rank_ordinal, result, created_at)
VALUES('fp_legacy','prob_legacy','fgr_legacy','h','c','n','f','low','low','low',0,1,'failure','t0')`,
		`INSERT INTO evaluation_runs(id, problem_id, run_id, mode, created_at)
VALUES('evr_legacy','prob_legacy','run_legacy','proposal','t0')`,
		`INSERT INTO evaluations(id, evaluation_run_id, proposal_id, verdict, verifier_kind, verification_strength, created_at)
VALUES('ev_first','evr_legacy','fp_legacy','failure','model-judgment','single-model-judgment','t0')`,
		`INSERT INTO evaluations(id, evaluation_run_id, proposal_id, verdict, verifier_kind, verification_strength, created_at)
VALUES('ev_second','evr_legacy','fp_legacy','failure','deterministic-check','deterministic','t1')`,
		`INSERT INTO evaluated_failures(proposal_id, evaluation_id, problem_id, verdict, created_at)
VALUES('fp_legacy','ev_first','prob_legacy','failure','t0')`,
	}
	for _, stmt := range seed {
		if _, err := s.db.ExecContext(ctx, stmt); err != nil {
			t.Fatalf("seed error: %v\n%s", err, stmt)
		}
	}

	if err := s.Migrate(ctx); err != nil {
		t.Fatalf("Migrate() replay error = %v", err)
	}

	// The PK is now evaluation_id: the seeded row survived and a SECOND
	// evaluation of the SAME proposal inserts without conflict — the exact
	// visibility the pre-v45 key made impossible.
	var ddl string
	if err := s.db.QueryRowContext(ctx, `SELECT sql FROM sqlite_master WHERE type='table' AND name='evaluated_failures'`).Scan(&ddl); err != nil {
		t.Fatalf("read DDL: %v", err)
	}
	if !strings.Contains(ddl, "evaluation_id TEXT PRIMARY KEY") {
		t.Fatalf("evaluated_failures PK not rebuilt per-evaluation:\n%s", ddl)
	}
	if _, err := s.db.ExecContext(ctx, `INSERT INTO evaluated_failures(evaluation_id, proposal_id, problem_id, verdict, created_at)
VALUES('ev_second','fp_legacy','prob_legacy','failure','t1')`); err != nil {
		t.Fatalf("second same-proposal marker must insert after v45: %v", err)
	}
	var n int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM evaluated_failures WHERE proposal_id='fp_legacy'`).Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != 2 {
		t.Fatalf("want preserved + new marker (2), got %d", n)
	}

	// Immutability triggers were recreated against the rebuilt table.
	if _, err := s.db.ExecContext(ctx, `UPDATE evaluated_failures SET verdict='partial_failure' WHERE evaluation_id='ev_first'`); err == nil {
		t.Fatal("update must be rejected by the recreated immutability trigger")
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM evaluated_failures WHERE evaluation_id='ev_first'`); err == nil {
		t.Fatal("delete must be rejected by the recreated immutability trigger")
	}

	// Idempotent: a second Migrate is a no-op.
	if err := s.Migrate(ctx); err != nil {
		t.Fatalf("second Migrate() error = %v", err)
	}
}
