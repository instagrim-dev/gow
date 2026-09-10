package store

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/domain"
)

// openMigratedStore opens and migrates a fresh store in a temp dir.
func openMigratedStore(t *testing.T) *Store {
	t.Helper()
	st, err := Open(t.TempDir() + "/inv.sqlite")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := st.Migrate(context.Background()); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	return st
}

// seedInvariantPrereqs creates, via one transaction of raw inserts, the full FK
// chain an invariant revision references: problem, run, source/snapshot,
// provider invocation, normalization revision, approach + revision, mechanism,
// signature (+ vocabulary), cluster run, mechanism cluster, and failure space.
// It returns the ids the invariant layer needs.
func seedInvariantPrereqs(t *testing.T, st *Store) (problemID, runID, clusterRunID, clusterID, failureSpaceID string) {
	t.Helper()
	ctx := context.Background()
	tnow := time.Now().UTC()
	now := formatTime(tnow)
	problemID, runID = domain.NewProblemID(tnow), domain.NewRunID(tnow)
	clusterRunID, clusterID, failureSpaceID = domain.NewClusterRunID(tnow), domain.NewMechanismClusterID(tnow), domain.NewFailureSpaceID(tnow)
	sigID := domain.NewMechanismSignatureID(tnow)
	mechID := domain.NewMechanismID(tnow)

	tx, err := st.db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("seed tx: %v", err)
	}
	exec := func(q string, args ...any) {
		if _, err := tx.ExecContext(ctx, q, args...); err != nil {
			tx.Rollback()
			t.Fatalf("seed %q: %v", q, err)
		}
	}
	exec(`INSERT INTO problems(id, slug, statement, status, created_at, created_by_run_id) VALUES(?,?,?,?,?,?)`, problemID, "p-inv", "s", "open", now, runID)
	exec(`INSERT INTO runs(id, problem_id, operation, status, input_ref, tool_name, tool_version, started_at, completed_at) VALUES(?,?,?,?,?,?,?,?,?)`, runID, problemID, "op", "running", "", "t", "v", now, now)
	exec(`INSERT INTO sources(id, problem_id, kind, logical_name, origin, created_at) VALUES('src_s',?,?,?,?,?)`, problemID, "k", "ln", "o", now)
	exec(`INSERT INTO source_snapshots(id, source_id, sha256, byte_length, media_type, object_path, observed_at, ingest_run_id) VALUES('snap_s','src_s','h',1,'text/plain','p',?,?)`, now, runID)
	exec(`INSERT INTO provider_invocations(id, run_id, role, provider_name, schema_version, request_hash, created_at) VALUES('pinv_s',?,?,?,?,?,?)`, runID, "normalize", "p", "sv", "rh", now)
	exec(`INSERT INTO normalization_revisions(id, problem_id, run_id, snapshot_id, provider_invocation_id, schema_version, config_hash, status, created_at) VALUES('nrev_s',?,?,'snap_s','pinv_s','sv','ch','succeeded',?)`, problemID, runID, now)
	exec(`INSERT INTO approaches(id, problem_id, logical_identity, created_at) VALUES('app_s',?,'li',?)`, problemID, now)
	exec(`INSERT INTO approach_revisions(id, approach_id, normalization_revision_id, label, created_at) VALUES('apr_s','app_s','nrev_s','L',?)`, now)
	exec(`INSERT INTO mechanisms(id, approach_revision_id, locality, construction_mode, uncertainty_mode) VALUES(?,'apr_s','local','constructive','deterministic')`, mechID)
	exec(`INSERT INTO canonical_vocabulary(version, notes, created_at) VALUES('vocab/v1','',?)`, now)
	exec(`INSERT INTO mechanism_signatures(id, mechanism_id, schema_version, vocabulary_version, fingerprint, run_id, created_at) VALUES(?,?,'mechanism/v1','vocab/v1','fp',?,?)`, sigID, mechID, runID, now)
	exec(`INSERT INTO cluster_runs(id, problem_id, run_id, schema_version, vocabulary_version, profile_version, cluster_algo_version, thresholds_hash, input_set_hash, signature_count, family_count, status, created_at) VALUES(?,?,?,'mechanism/v1','vocab/v1','profile/v1','cluster/v1','th','ish',1,1,'clean',?)`, clusterRunID, problemID, runID, now)
	exec(`INSERT INTO mechanism_clusters(id, cluster_run_id, cluster_fingerprint, representative_signature_id, member_count, intra_variation, isolate, outcome_class, outcome_mixed, ordinal) VALUES(?,?,'cfp',?,1,'{}',0,'failure',0,0)`, clusterID, clusterRunID, sigID)
	exec(`INSERT INTO cluster_members(cluster_id, signature_id, mechanism_id, redundant, ordinal) VALUES(?,?,?,0,0)`, clusterID, sigID, mechID)
	exec(`INSERT INTO failure_spaces(id, problem_id, cluster_run_id, run_id, revision, distinct_family_count, redundant_member_count, created_at) VALUES(?,?,?,?,1,1,0,?)`, failureSpaceID, problemID, clusterRunID, runID, now)
	if err := tx.Commit(); err != nil {
		t.Fatalf("seed commit: %v", err)
	}
	return
}

func sampleRevision(t *testing.T, st *Store) InvariantRevisionRecord {
	t.Helper()
	problemID, runID, clusterRunID, clusterID, fsID := seedInvariantPrereqs(t, st)
	now := time.Now().UTC()
	return InvariantRevisionRecord{
		ID:              domain.NewInvariantRevisionID(now),
		ProblemID:       problemID,
		FailureSpaceID:  fsID,
		ClusterRunID:    clusterRunID,
		RunID:           runID,
		MinerVersion:    "invariant/v1",
		PredicateSchema: "invariant-predicate/v1",
		MinSupport:      2,
		CandidateCount:  1,
		CreatedAt:       formatTime(now),
		Invocation: InvariantProviderInvocation{
			ID: domain.NewProviderInvocationID(now), RunID: runID,
			ProviderName: "fixture", SchemaVersion: "invariant-predicate/v1",
			RequestHash: "rh", CreatedAt: formatTime(now),
		},
		Candidates: []CandidateInvariantRow{{
			ID:                    domain.NewCandidateInvariantID(now),
			PredicateFingerprint:  "fp1",
			PredicateJSON:         `{"schema":"invariant-predicate/v1"}`,
			Statement:             "preserves residue locality",
			AbstractionLevel:      "mechanism",
			AssociationStatus:     "recurring",
			DistinctFamilySupport: 2,
			FailureCoverageNum:    2, FailureCoverageDen: 2,
			SupportInferredCount: 1,
			Ordinal:              0,
			FamilyEvaluations: []InvariantFamilyEvaluationRow{{
				ClusterID: clusterID, OutcomeClass: "failure", Role: "support", Verdict: "satisfies", InferredCount: 1,
			}},
		}},
	}
}

func TestPersistInvariantRevisionIdempotentAndRoundTrip(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()
	rec := sampleRevision(t, st)

	res, err := st.PersistInvariantRevision(ctx, rec)
	if err != nil {
		t.Fatalf("persist: %v", err)
	}
	if !res.Created || res.Record.Revision != 1 {
		t.Fatalf("expected created revision 1, got created=%v rev=%d", res.Created, res.Record.Revision)
	}
	// Re-persisting the same identity tuple returns the existing revision.
	res2, err := st.PersistInvariantRevision(ctx, rec)
	if err != nil {
		t.Fatalf("re-persist: %v", err)
	}
	if res2.Created {
		t.Fatal("re-persist must be idempotent (Created=false)")
	}
	got, err := st.GetInvariantRevision(ctx, res.Record.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(got.Candidates) != 1 || got.Candidates[0].SupportInferredCount != 1 {
		t.Fatalf("round-trip lost data: %+v", got.Candidates)
	}
	if len(got.Candidates[0].FamilyEvaluations) != 1 {
		t.Fatalf("expected 1 family evaluation, got %d", len(got.Candidates[0].FamilyEvaluations))
	}
}

func TestCandidateInvariantImmutable(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()
	rec := sampleRevision(t, st)
	if _, err := st.PersistInvariantRevision(ctx, rec); err != nil {
		t.Fatalf("persist: %v", err)
	}
	if _, err := st.db.ExecContext(ctx, `UPDATE candidate_invariants SET statement='x'`); err == nil {
		t.Fatal("expected immutability trigger to abort UPDATE")
	}
	if _, err := st.db.ExecContext(ctx, `DELETE FROM invariant_revisions`); err == nil {
		t.Fatal("expected immutability trigger to abort DELETE")
	}
}

func TestCandidateInvariantRejectsNonProposedState(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()
	_, runID, _, _, _ := seedInvariantPrereqs(t, st)
	_ = runID
	// A direct insert with a non-proposed initial_state must be rejected by CHECK.
	_, err := st.db.ExecContext(ctx, `INSERT INTO candidate_invariants(id, invariant_revision_id, predicate_fingerprint, statement, abstraction_level, initial_state, association_status, distinct_family_support, failure_coverage_num, failure_coverage_den, support_explicit_count, support_inferred_count, support_other_count, ordinal) VALUES('inv_x','ivr_x','fp','s','m','established','recurring',0,0,0,0,0,0,0)`)
	if err == nil || !strings.Contains(err.Error(), "CHECK") {
		t.Fatalf("expected CHECK rejection of non-proposed state, got %v", err)
	}
}

// TestV11ProviderRoleGeneralizationFromPreV11 constructs the pre-v11
// normalize-only provider_invocations shape with a child normalization_revisions
// FK, runs the v11 role generalization, and asserts (i) 'invariant' is now
// permitted and (ii) the pre-existing child FK still resolves (the parent
// rebuild did not orphan it).
func TestV11ProviderRoleGeneralizationFromPreV11(t *testing.T) {
	ctx := context.Background()
	st, err := Open(t.TempDir() + "/prev11.sqlite")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { st.Close() })

	// Apply migrations 1..10 only (stop before v11) to reach the normalize-only
	// role CHECK with the child normalization_revisions table present.
	if err := applyMigrationsThrough(ctx, st, 10); err != nil {
		t.Fatalf("apply through v10: %v", err)
	}
	if allows, _ := providerRoleAllowsInvariantDB(ctx, st); allows {
		t.Fatal("pre-v11 must NOT permit invariant role")
	}

	now := formatTime(time.Now().UTC())
	// Seed problem/run/snapshot/source so the FK chain for a normalization_revision
	// resolves. problems.created_by_run_id <-> runs.problem_id is a deferred
	// circular FK, so the seed must run in one transaction.
	seedTx, err := st.db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("seed tx: %v", err)
	}
	seed := func(q string, args ...any) {
		if _, err := seedTx.ExecContext(ctx, q, args...); err != nil {
			seedTx.Rollback()
			t.Fatalf("seed exec %q: %v", q, err)
		}
	}
	seed(`INSERT INTO problems(id, slug, statement, status, created_at, created_by_run_id) VALUES('prb_1','s','st','open',?,'run_1')`, now)
	seed(`INSERT INTO runs(id, problem_id, operation, status, input_ref, tool_name, tool_version, started_at, completed_at) VALUES('run_1','prb_1','op','running','','t','v',?,?)`, now, now)
	seed(`INSERT INTO sources(id, problem_id, kind, logical_name, origin, created_at) VALUES('src_1','prb_1','k','ln','o',?)`, now)
	seed(`INSERT INTO source_snapshots(id, source_id, sha256, byte_length, media_type, object_path, observed_at, ingest_run_id) VALUES('snap_1','src_1','h',1,'text/plain','path',?,'run_1')`, now)
	seed(`INSERT INTO provider_invocations(id, run_id, role, provider_name, schema_version, request_hash, created_at) VALUES('pinv_1','run_1','normalize','p','sv','rh',?)`, now)
	seed(`INSERT INTO normalization_revisions(id, problem_id, run_id, snapshot_id, provider_invocation_id, schema_version, config_hash, status, created_at) VALUES('nrev_1','prb_1','run_1','snap_1','pinv_1','sv','ch','succeeded',?)`, now)
	if err := seedTx.Commit(); err != nil {
		t.Fatalf("seed commit: %v", err)
	}

	// Run v11.
	if err := applyMigrationsThrough(ctx, st, 11); err != nil {
		t.Fatalf("apply v11: %v", err)
	}
	if allows, _ := providerRoleAllowsInvariantDB(ctx, st); !allows {
		t.Fatal("v11 must permit invariant role")
	}
	// The pre-existing child FK still resolves after the parent rebuild.
	var cnt int
	if err := st.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM normalization_revisions nr JOIN provider_invocations pi ON nr.provider_invocation_id = pi.id WHERE nr.id='nrev_1'`).Scan(&cnt); err != nil {
		t.Fatalf("join: %v", err)
	}
	if cnt != 1 {
		t.Fatalf("child FK orphaned by parent rebuild, join count=%d", cnt)
	}
	// An invariant-role insert now succeeds.
	mustExec(t, st, `INSERT INTO provider_invocations(id, run_id, role, provider_name, schema_version, request_hash, created_at) VALUES('pinv_inv','run_1','invariant','p','sv','rh',?)`, now)
}

func applyMigrationsThrough(ctx context.Context, st *Store, maxVersion int) error {
	tx, err := st.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL)`); err != nil {
		return err
	}
	applied := map[int]struct{}{}
	rows, err := tx.QueryContext(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return err
	}
	for rows.Next() {
		var v int
		rows.Scan(&v)
		applied[v] = struct{}{}
	}
	rows.Close()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	for _, m := range migrations {
		if m.version > maxVersion {
			break
		}
		if _, ok := applied[m.version]; ok {
			continue
		}
		if m.apply != nil {
			if err := m.apply(ctx, tx); err != nil {
				return err
			}
		} else if _, err := tx.ExecContext(ctx, m.sql); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations(version, applied_at) VALUES(?, ?)`, m.version, now); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func providerRoleAllowsInvariantDB(ctx context.Context, st *Store) (bool, error) {
	var ddl string
	if err := st.db.QueryRowContext(ctx, `SELECT sql FROM sqlite_master WHERE type='table' AND name='provider_invocations'`).Scan(&ddl); err != nil {
		return false, err
	}
	return strings.Contains(ddl, "'invariant'"), nil
}

func mustExec(t *testing.T, st *Store, q string, args ...any) {
	t.Helper()
	if _, err := st.db.ExecContext(context.Background(), q, args...); err != nil {
		t.Fatalf("exec %q: %v", q, err)
	}
}
