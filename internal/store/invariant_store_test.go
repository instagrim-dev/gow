package store

import (
	"context"
	"fmt"
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
			AssociationStatus:     "contrast_observed",
			DistinctFamilySupport: 2,
			FailureCoverageNum:    2, FailureCoverageDen: 2,
			ContrastViolatingNum: 1, ContrastEligibleDen: 3,
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
	// Contrast counts and the renamed association_status survive the round-trip (F2).
	if got.Candidates[0].ContrastViolatingNum != 1 || got.Candidates[0].ContrastEligibleDen != 3 {
		t.Fatalf("contrast counts lost: got %d/%d", got.Candidates[0].ContrastViolatingNum, got.Candidates[0].ContrastEligibleDen)
	}
	if got.Candidates[0].AssociationStatus != "contrast_observed" {
		t.Fatalf("association_status = %q, want contrast_observed", got.Candidates[0].AssociationStatus)
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

// TestCandidateInvariantAssociationStatusEnum proves the v12 rename took effect:
// the legacy 'discriminative' value is rejected by CHECK and the new
// 'contrast_observed' value is accepted (F2).
func TestCandidateInvariantAssociationStatusEnum(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()
	rec := sampleRevision(t, st)
	// Persist the revision so a valid invariant_revision_id exists to reference.
	if _, err := st.PersistInvariantRevision(ctx, rec); err != nil {
		t.Fatalf("persist: %v", err)
	}
	revID := rec.ID

	// Legacy 'discriminative' must now be rejected by the widened CHECK.
	_, err := st.db.ExecContext(ctx, `INSERT INTO candidate_invariants(id, invariant_revision_id, predicate_fingerprint, statement, abstraction_level, initial_state, association_status, distinct_family_support, failure_coverage_num, failure_coverage_den, contrast_violating_num, contrast_eligible_den, support_explicit_count, support_inferred_count, support_other_count, ordinal) VALUES('inv_disc',?,'fpd','s','m','proposed','discriminative',0,0,0,0,0,0,0,0,9)`, revID)
	if err == nil || !strings.Contains(err.Error(), "CHECK") {
		t.Fatalf("expected CHECK rejection of legacy 'discriminative', got %v", err)
	}

	// New 'contrast_observed' must be accepted.
	if _, err := st.db.ExecContext(ctx, `INSERT INTO candidate_invariants(id, invariant_revision_id, predicate_fingerprint, statement, abstraction_level, initial_state, association_status, distinct_family_support, failure_coverage_num, failure_coverage_den, contrast_violating_num, contrast_eligible_den, support_explicit_count, support_inferred_count, support_other_count, ordinal) VALUES('inv_co',?,'fpc','s','m','proposed','contrast_observed',0,0,0,0,0,0,0,0,10)`, revID); err != nil {
		t.Fatalf("contrast_observed must be accepted, got %v", err)
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

	// The widened CHECK still constrains: a role outside the enum is rejected,
	// proving the writable_schema edit widened the constraint rather than
	// dropping it. This is the parseability/semantic proof the migration's
	// integrity_check cannot give (integrity_check validates storage structure,
	// not DDL text).
	if _, err := st.db.ExecContext(ctx, `INSERT INTO provider_invocations(id, run_id, role, provider_name, schema_version, request_hash, created_at) VALUES('pinv_bad','run_1','frontier','p','sv','rh',?)`, now); err == nil {
		t.Fatal("insert with out-of-enum role 'frontier' succeeded, want CHECK rejection")
	}
}

// TestV10ClusterRunsInputSetHashFromPopulatedPreV10 is the F4 regression: a
// POPULATED pre-v10 database (cluster_runs with child mechanism_clusters +
// cluster_members + failure_spaces rows, and NO input_set_hash column) must
// migrate through v10 without orphaning any FK child. The original v10 code
// dropped and recreated cluster_runs, which is not FK-safe under foreign_keys=ON
// with existing children; the repaired path adds the column + a UNIQUE index in
// place, preserving every child by construction.
func TestV10ClusterRunsInputSetHashFromPopulatedPreV10(t *testing.T) {
	ctx := context.Background()
	st, err := Open(t.TempDir() + "/prev10.sqlite")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { st.Close() })

	// Apply migrations 1..9 only: reach the pre-v10 cluster_runs shape (no
	// input_set_hash) with the cluster child tables present.
	if err := applyMigrationsThrough(ctx, st, 9); err != nil {
		t.Fatalf("apply through v9: %v", err)
	}
	if has := dbColumnExists(t, st, "cluster_runs", "input_set_hash"); has {
		t.Fatal("pre-v10 cluster_runs must NOT have input_set_hash")
	}

	now := formatTime(time.Now().UTC())
	// Seed a full populated cluster chain in one transaction (circular
	// problems<->runs FK is deferred). cluster_runs is inserted WITHOUT
	// input_set_hash (the pre-v10 column list).
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
	seed(`INSERT INTO approaches(id, problem_id, logical_identity, created_at) VALUES('app_1','prb_1','li',?)`, now)
	seed(`INSERT INTO sources(id, problem_id, kind, logical_name, origin, created_at) VALUES('src_1','prb_1','k','ln','o',?)`, now)
	seed(`INSERT INTO source_snapshots(id, source_id, sha256, byte_length, media_type, object_path, observed_at, ingest_run_id) VALUES('snap_1','src_1','h',1,'text/plain','path',?,'run_1')`, now)
	seed(`INSERT INTO provider_invocations(id, run_id, role, provider_name, schema_version, request_hash, created_at) VALUES('pinv_1','run_1','normalize','p','sv','rh',?)`, now)
	seed(`INSERT INTO normalization_revisions(id, problem_id, run_id, snapshot_id, provider_invocation_id, schema_version, config_hash, status, created_at) VALUES('nrev_1','prb_1','run_1','snap_1','pinv_1','sv','ch','succeeded',?)`, now)
	seed(`INSERT INTO approach_revisions(id, approach_id, normalization_revision_id, label, created_at) VALUES('apr_1','app_1','nrev_1','L',?)`, now)
	seed(`INSERT INTO mechanisms(id, approach_revision_id, locality, construction_mode, uncertainty_mode) VALUES('mech_1','apr_1','local','constructive','deterministic')`)
	seed(`INSERT INTO canonical_vocabulary(version, notes, created_at) VALUES('vocab/v1','',?)`, now)
	seed(`INSERT INTO mechanism_signatures(id, mechanism_id, schema_version, vocabulary_version, fingerprint, run_id, created_at) VALUES('msig_1','mech_1','mechanism/v1','vocab/v1','fp','run_1',?)`, now)
	// Pre-v10 cluster_runs INSERT: note the column list has NO input_set_hash.
	seed(`INSERT INTO cluster_runs(id, problem_id, run_id, schema_version, vocabulary_version, profile_version, cluster_algo_version, thresholds_hash, signature_count, family_count, status, created_at) VALUES('crun_1','prb_1','run_1','mechanism/v1','vocab/v1','profile/v1','cluster/v1','th',1,1,'clean',?)`, now)
	seed(`INSERT INTO mechanism_clusters(id, cluster_run_id, cluster_fingerprint, representative_signature_id, member_count, intra_variation, isolate, ordinal) VALUES('mcl_1','crun_1','cfp','msig_1',1,'{}',0,0)`)
	seed(`INSERT INTO cluster_members(cluster_id, signature_id, mechanism_id, redundant, ordinal) VALUES('mcl_1','msig_1','mech_1',0,0)`)
	seed(`INSERT INTO failure_spaces(id, problem_id, cluster_run_id, run_id, revision, distinct_family_count, redundant_member_count, created_at) VALUES('fs_1','prb_1','crun_1','run_1',1,1,0,?)`, now)
	if err := seedTx.Commit(); err != nil {
		t.Fatalf("seed commit: %v", err)
	}

	// Migrate forward through the latest version (runs v10, v11, v12).
	if err := applyMigrationsThrough(ctx, st, currentSchemaVersion); err != nil {
		t.Fatalf("apply forward from populated v9: %v", err)
	}

	// The column now exists and the parent row survived with an empty hash.
	if has := dbColumnExists(t, st, "cluster_runs", "input_set_hash"); !has {
		t.Fatal("post-v10 cluster_runs must have input_set_hash")
	}
	var hash string
	if err := st.db.QueryRowContext(ctx, `SELECT input_set_hash FROM cluster_runs WHERE id='crun_1'`).Scan(&hash); err != nil {
		t.Fatalf("cluster_runs row lost by migration: %v", err)
	}
	if hash != "" {
		t.Fatalf("migrated pre-v10 row must get empty input_set_hash, got %q", hash)
	}

	// Every FK child still resolves to its parent (not orphaned by a parent drop).
	for _, c := range []struct {
		name, query string
	}{
		{"mechanism_clusters", `SELECT COUNT(*) FROM mechanism_clusters mc JOIN cluster_runs cr ON mc.cluster_run_id = cr.id WHERE mc.id='mcl_1'`},
		{"cluster_members", `SELECT COUNT(*) FROM cluster_members cm JOIN mechanism_clusters mc ON cm.cluster_id = mc.id WHERE cm.signature_id='msig_1'`},
		{"failure_spaces", `SELECT COUNT(*) FROM failure_spaces fs JOIN cluster_runs cr ON fs.cluster_run_id = cr.id WHERE fs.id='fs_1'`},
	} {
		var cnt int
		if err := st.db.QueryRowContext(ctx, c.query).Scan(&cnt); err != nil {
			t.Fatalf("%s join: %v", c.name, err)
		}
		if cnt != 1 {
			t.Fatalf("%s child orphaned by cluster_runs migration, join count=%d", c.name, cnt)
		}
	}

	// A global FK integrity check must be clean: no dangling references anywhere.
	rows, err := st.db.QueryContext(ctx, `PRAGMA foreign_key_check`)
	if err != nil {
		t.Fatalf("foreign_key_check: %v", err)
	}
	defer rows.Close()
	if rows.Next() {
		t.Fatal("PRAGMA foreign_key_check reported a dangling FK after migration")
	}

	// The widened idempotency key is enforced: a second cluster_runs row with the
	// same identity tuple (including the empty input_set_hash) is rejected.
	_, dupErr := st.db.ExecContext(ctx, `INSERT INTO cluster_runs(id, problem_id, run_id, schema_version, vocabulary_version, profile_version, cluster_algo_version, thresholds_hash, input_set_hash, signature_count, family_count, status, created_at) VALUES('crun_dup','prb_1','run_1','mechanism/v1','vocab/v1','profile/v1','cluster/v1','th','',1,1,'clean',?)`, now)
	if dupErr == nil {
		t.Fatal("duplicate cluster_runs identity tuple should violate the UNIQUE index")
	}
}

// TestV10RemovesLegacyClusterRunsInlineUnique covers the legacy path of the F4
// fix: a database whose cluster_runs was created under the ORIGINAL v7 DDL (an
// inline table-level UNIQUE backed by an sqlite_autoindex, no named identity
// index) must have that constraint replaced by the 7-column UNIQUE index without
// corrupting the schema or orphaning the autoindex.
func TestV10RemovesLegacyClusterRunsInlineUnique(t *testing.T) {
	ctx := context.Background()
	st, err := Open(t.TempDir() + "/legacy_v7.sqlite")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { st.Close() })

	// Build the minimal legacy schema by hand: the prerequisite tables plus a
	// cluster_runs with the ORIGINAL inline UNIQUE (the exact pre-index DDL), then
	// stamp migrations 1..9 as applied so Migrate() runs v10+ against it.
	now := formatTime(time.Now().UTC())
	tx, err := st.db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("tx: %v", err)
	}
	exec := func(q string, args ...any) {
		if _, err := tx.ExecContext(ctx, q, args...); err != nil {
			tx.Rollback()
			t.Fatalf("exec %q: %v", q, err)
		}
	}
	exec(`CREATE TABLE schema_migrations (version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL)`)
	exec(`CREATE TABLE problems (id TEXT PRIMARY KEY, slug TEXT, statement TEXT, status TEXT, created_at TEXT, created_by_run_id TEXT)`)
	exec(`CREATE TABLE runs (id TEXT PRIMARY KEY, problem_id TEXT, operation TEXT, status TEXT, input_ref TEXT, tool_name TEXT, tool_version TEXT, started_at TEXT, completed_at TEXT)`)
	exec(`CREATE TABLE canonical_vocabulary (version TEXT PRIMARY KEY, notes TEXT, created_at TEXT)`)
	// The ORIGINAL v7 cluster_runs DDL: inline table-level UNIQUE, no input_set_hash.
	exec(`CREATE TABLE cluster_runs (
  id TEXT PRIMARY KEY,
  problem_id TEXT NOT NULL REFERENCES problems(id),
  run_id TEXT NOT NULL REFERENCES runs(id),
  schema_version TEXT NOT NULL,
  vocabulary_version TEXT NOT NULL REFERENCES canonical_vocabulary(version),
  profile_version TEXT NOT NULL,
  cluster_algo_version TEXT NOT NULL,
  thresholds_hash TEXT NOT NULL,
  signature_count INTEGER NOT NULL,
  family_count INTEGER NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('clean', 'degraded')),
  created_at TEXT NOT NULL,
  UNIQUE(problem_id, schema_version, vocabulary_version, profile_version, cluster_algo_version, thresholds_hash)
)`)
	// A populated row so the constraint's autoindex is exercised.
	exec(`INSERT INTO problems(id, slug, statement, status, created_at, created_by_run_id) VALUES('prb_1','s','st','open',?,'run_1')`, now)
	exec(`INSERT INTO runs(id, problem_id, operation, status, input_ref, tool_name, tool_version, started_at, completed_at) VALUES('run_1','prb_1','op','running','','t','v',?,?)`, now, now)
	exec(`INSERT INTO canonical_vocabulary(version, notes, created_at) VALUES('vocab/v1','',?)`, now)
	exec(`INSERT INTO cluster_runs(id, problem_id, run_id, schema_version, vocabulary_version, profile_version, cluster_algo_version, thresholds_hash, signature_count, family_count, status, created_at) VALUES('crun_1','prb_1','run_1','mechanism/v1','vocab/v1','profile/v1','cluster/v1','th',1,1,'clean',?)`, now)
	for v := 1; v <= 9; v++ {
		exec(`INSERT INTO schema_migrations(version, applied_at) VALUES(?, ?)`, v, now)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}

	// Run only the cluster_runs v10 migration step against this hand-built schema
	// (the full Migrate would require the entire v1..v9 table set). Migrate runs
	// with foreign_keys OFF so a parent rebuild is FK-safe; mirror that here.
	if _, err := st.db.ExecContext(ctx, `PRAGMA foreign_keys = OFF`); err != nil {
		t.Fatalf("fk off: %v", err)
	}
	mtx, err := st.db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("mtx: %v", err)
	}
	if err := addClusterRunsInputSetHashFKSafe(ctx, mtx); err != nil {
		mtx.Rollback()
		t.Fatalf("addClusterRunsInputSetHashFKSafe on legacy schema: %v", err)
	}
	// No dangling references after the rebuild.
	fkRows, err := mtx.QueryContext(ctx, `PRAGMA foreign_key_check`)
	if err != nil {
		mtx.Rollback()
		t.Fatalf("foreign_key_check: %v", err)
	}
	if fkRows.Next() {
		fkRows.Close()
		mtx.Rollback()
		t.Fatal("foreign_key_check reported a dangling reference after legacy rebuild")
	}
	fkRows.Close()
	if err := mtx.Commit(); err != nil {
		t.Fatalf("mtx commit: %v", err)
	}
	if _, err := st.db.ExecContext(ctx, `PRAGMA foreign_keys = ON`); err != nil {
		t.Fatalf("fk on: %v", err)
	}

	var ddl string
	if err := st.db.QueryRowContext(ctx, `SELECT sql FROM sqlite_master WHERE type='table' AND name='cluster_runs'`).Scan(&ddl); err != nil {
		t.Fatalf("read ddl: %v", err)
	}
	if strings.Contains(ddl, "UNIQUE(problem_id") {
		t.Fatalf("legacy inline UNIQUE not removed: %s", ddl)
	}
	// integrity_check must be clean (no orphaned autoindex).
	var res string
	if err := st.db.QueryRowContext(ctx, `PRAGMA integrity_check`).Scan(&res); err != nil {
		t.Fatalf("integrity_check: %v", err)
	}
	if res != "ok" {
		t.Fatalf("integrity_check = %q, want ok", res)
	}
	// The migrated row survived with an empty hash.
	var hash string
	if err := st.db.QueryRowContext(ctx, `SELECT input_set_hash FROM cluster_runs WHERE id='crun_1'`).Scan(&hash); err != nil {
		t.Fatalf("row lost: %v", err)
	}
	if hash != "" {
		t.Fatalf("migrated legacy row must get empty input_set_hash, got %q", hash)
	}
	// The widened 7-column index now enforces identity: a same-6-tuple row that
	// differs only by input_set_hash is now ALLOWED (the old 6-col UNIQUE would
	// have rejected it), while an exact duplicate is rejected.
	if _, err := st.db.ExecContext(ctx, `INSERT INTO cluster_runs(id, problem_id, run_id, schema_version, vocabulary_version, profile_version, cluster_algo_version, thresholds_hash, input_set_hash, signature_count, family_count, status, created_at) VALUES('crun_2','prb_1','run_1','mechanism/v1','vocab/v1','profile/v1','cluster/v1','th','different-hash',1,1,'clean',?)`, now); err != nil {
		t.Fatalf("row differing only by input_set_hash must be allowed after widening: %v", err)
	}
	if _, err := st.db.ExecContext(ctx, `INSERT INTO cluster_runs(id, problem_id, run_id, schema_version, vocabulary_version, profile_version, cluster_algo_version, thresholds_hash, input_set_hash, signature_count, family_count, status, created_at) VALUES('crun_dup','prb_1','run_1','mechanism/v1','vocab/v1','profile/v1','cluster/v1','th','',1,1,'clean',?)`, now); err == nil {
		t.Fatal("exact identity duplicate should violate the widened UNIQUE index")
	}
}

func applyMigrationsThrough(ctx context.Context, st *Store, maxVersion int) error {
	// Mirror Store.Migrate: run with foreign_keys OFF so parent rebuilds are
	// FK-safe, and re-enable afterward.
	if _, err := st.db.ExecContext(ctx, `PRAGMA foreign_keys = OFF`); err != nil {
		return err
	}
	defer func() { _, _ = st.db.ExecContext(ctx, `PRAGMA foreign_keys = ON`) }()
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

// dbColumnExists reports whether a column exists, probing the live DB directly
// (no transaction) via PRAGMA table_info.
func dbColumnExists(t *testing.T, st *Store, table, column string) bool {
	t.Helper()
	rows, err := st.db.QueryContext(context.Background(), fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		t.Fatalf("table_info(%s): %v", table, err)
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt any
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			t.Fatalf("scan table_info: %v", err)
		}
		if name == column {
			return true
		}
	}
	return false
}
