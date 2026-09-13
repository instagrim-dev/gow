package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
)

const currentSchemaVersion = 47

// reviewAssessmentReferenceScopeSQL closes the review-ledger subject-binding
// gap for existing stores. Fresh stores also get this trigger from v44's
// reviewLedgerSQL definition below.
const reviewAssessmentReferenceScopeSQL = `
CREATE TRIGGER IF NOT EXISTS review_assessments_reference_scope
BEFORE INSERT ON review_assessments
BEGIN
  SELECT CASE
    WHEN NOT EXISTS (
      SELECT 1
      FROM review_applicability_decisions d
      WHERE d.id = NEW.applicability_decision_id
        AND d.policy_id = NEW.policy_id
        AND d.obligation_id = NEW.obligation_id
        AND d.subject_ref = NEW.subject_ref
    )
    THEN RAISE(ABORT, 'review assessment applicability decision scope does not match assessment')
  END;
  SELECT CASE
    WHEN NOT EXISTS (
      SELECT 1
      FROM review_dependency_manifests m
      WHERE m.id = NEW.manifest_id
        AND m.policy_id = NEW.policy_id
        AND m.obligation_id = NEW.obligation_id
    )
    THEN RAISE(ABORT, 'review assessment dependency manifest scope does not match assessment')
  END;
END;
`

// witnessAttemptBindingSQL is the additive DDL for migration v46: one
// checkable attempt→output binding per witness-backed evaluation (see the v46
// migration comment for doctrine).
const witnessAttemptBindingSQL = `
CREATE TABLE IF NOT EXISTS witness_attempt_bindings (
  evaluation_id TEXT PRIMARY KEY REFERENCES evaluations(id),
  proposal_id TEXT NOT NULL REFERENCES frontier_proposals(id),
  procedure TEXT NOT NULL,
  executor_version TEXT NOT NULL,
  params_canonical TEXT NOT NULL,
  tuple_canonical TEXT NOT NULL,
  created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_witness_attempt_bindings_proposal
  ON witness_attempt_bindings(proposal_id);
CREATE TRIGGER IF NOT EXISTS witness_attempt_bindings_immutable_update
BEFORE UPDATE ON witness_attempt_bindings
BEGIN
  SELECT RAISE(ABORT, 'witness attempt bindings are immutable');
END;
CREATE TRIGGER IF NOT EXISTS witness_attempt_bindings_immutable_delete
BEFORE DELETE ON witness_attempt_bindings
BEGIN
  SELECT RAISE(ABORT, 'witness attempt bindings are immutable');
END;
`

// migration is one ordered schema step. Most steps are a static SQL blob run as
// one statement batch. A step may instead supply an `apply` func when the change
// must inspect the live schema (e.g. an in-place repair that must be a no-op on
// databases already created with the corrected schema). Exactly one of sql or
// apply is set.
type migration struct {
	version int
	sql     string
	apply   func(ctx context.Context, tx *sql.Tx) error
}

var migrations = []migration{
	{
		version: 1,
		sql: `
CREATE TABLE IF NOT EXISTS schema_migrations (
  version INTEGER PRIMARY KEY,
  applied_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS problems (
  id TEXT PRIMARY KEY,
  slug TEXT NOT NULL UNIQUE,
  statement TEXT NOT NULL,
  status TEXT NOT NULL,
  created_at TEXT NOT NULL,
  created_by_run_id TEXT NOT NULL REFERENCES runs(id) DEFERRABLE INITIALLY DEFERRED
);

CREATE TABLE IF NOT EXISTS runs (
  id TEXT PRIMARY KEY,
  problem_id TEXT NOT NULL REFERENCES problems(id),
  parent_run_id TEXT REFERENCES runs(id),
  operation TEXT NOT NULL,
  status TEXT NOT NULL,
  input_ref TEXT NOT NULL,
  tool_name TEXT NOT NULL,
  tool_version TEXT NOT NULL,
  started_at TEXT NOT NULL,
  completed_at TEXT NOT NULL,
  error_summary TEXT
);

CREATE INDEX IF NOT EXISTS idx_problems_created_at ON problems(created_at, id);
CREATE INDEX IF NOT EXISTS idx_runs_problem_id ON runs(problem_id, started_at, id);
`,
	},
	{
		version: 2,
		sql: `
CREATE TABLE IF NOT EXISTS sources (
  id TEXT PRIMARY KEY,
  problem_id TEXT NOT NULL REFERENCES problems(id),
  kind TEXT NOT NULL,
  logical_name TEXT NOT NULL,
  origin TEXT NOT NULL,
  created_at TEXT NOT NULL,
  UNIQUE(problem_id, origin)
);

CREATE TABLE IF NOT EXISTS source_snapshots (
  id TEXT PRIMARY KEY,
  source_id TEXT NOT NULL REFERENCES sources(id),
  sha256 TEXT NOT NULL,
  byte_length INTEGER NOT NULL,
  media_type TEXT NOT NULL,
  object_path TEXT NOT NULL,
  observed_at TEXT NOT NULL,
  ingest_run_id TEXT NOT NULL REFERENCES runs(id),
  supersedes_snapshot_id TEXT REFERENCES source_snapshots(id),
  UNIQUE(source_id, sha256)
);

CREATE INDEX IF NOT EXISTS idx_sources_problem_origin ON sources(problem_id, origin);
CREATE INDEX IF NOT EXISTS idx_source_snapshots_source_observed ON source_snapshots(source_id, observed_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_source_snapshots_source_hash ON source_snapshots(source_id, sha256);

CREATE TRIGGER IF NOT EXISTS sources_immutable_update
BEFORE UPDATE ON sources
BEGIN
  SELECT RAISE(ABORT, 'sources are immutable');
END;

CREATE TRIGGER IF NOT EXISTS sources_immutable_delete
BEFORE DELETE ON sources
BEGIN
  SELECT RAISE(ABORT, 'sources are immutable');
END;

CREATE TRIGGER IF NOT EXISTS source_snapshots_immutable_update
BEFORE UPDATE ON source_snapshots
BEGIN
  SELECT RAISE(ABORT, 'source snapshots are immutable');
END;

CREATE TRIGGER IF NOT EXISTS source_snapshots_immutable_delete
BEFORE DELETE ON source_snapshots
BEGIN
  SELECT RAISE(ABORT, 'source snapshots are immutable');
END;
`,
	},
	{
		version: 3,
		sql: `
CREATE TABLE IF NOT EXISTS provider_invocations (
  id TEXT PRIMARY KEY,
  run_id TEXT NOT NULL REFERENCES runs(id),
  role TEXT NOT NULL CHECK (role IN ('normalize')),
  provider_name TEXT NOT NULL,
  provider_version TEXT NOT NULL DEFAULT '',
  model_name TEXT NOT NULL DEFAULT '',
  schema_version TEXT NOT NULL,
  request_hash TEXT NOT NULL,
  request_payload TEXT NOT NULL DEFAULT '',
  response_payload TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS approaches (
  id TEXT PRIMARY KEY,
  problem_id TEXT NOT NULL REFERENCES problems(id),
  logical_identity TEXT NOT NULL,
  created_at TEXT NOT NULL,
  UNIQUE(problem_id, logical_identity)
);

CREATE TABLE IF NOT EXISTS normalization_revisions (
  id TEXT PRIMARY KEY,
  problem_id TEXT NOT NULL REFERENCES problems(id),
  run_id TEXT NOT NULL REFERENCES runs(id),
  snapshot_id TEXT NOT NULL REFERENCES source_snapshots(id),
  provider_invocation_id TEXT NOT NULL REFERENCES provider_invocations(id),
  schema_version TEXT NOT NULL,
  config_hash TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('succeeded', 'failed', 'skipped')),
  supersedes_revision_id TEXT REFERENCES normalization_revisions(id),
  skip_reason TEXT,
  created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS approach_revisions (
  id TEXT PRIMARY KEY,
  approach_id TEXT NOT NULL REFERENCES approaches(id),
  normalization_revision_id TEXT NOT NULL REFERENCES normalization_revisions(id),
  label TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  supersedes_revision_id TEXT REFERENCES approach_revisions(id),
  created_at TEXT NOT NULL,
  UNIQUE(approach_id, normalization_revision_id)
);

CREATE TABLE IF NOT EXISTS mechanisms (
  id TEXT PRIMARY KEY,
  approach_revision_id TEXT NOT NULL REFERENCES approach_revisions(id),
  locality TEXT NOT NULL,
  construction_mode TEXT NOT NULL,
  uncertainty_mode TEXT NOT NULL,
  notes TEXT NOT NULL DEFAULT '',
  UNIQUE(approach_revision_id)
);

CREATE TABLE IF NOT EXISTS mechanism_attributes (
  mechanism_id TEXT NOT NULL REFERENCES mechanisms(id),
  kind TEXT NOT NULL CHECK (kind IN ('representation', 'assumption', 'operator', 'preserves', 'breaks', 'auxiliary_object')),
  value TEXT NOT NULL,
  ordinal INTEGER NOT NULL,
  PRIMARY KEY(mechanism_id, kind, value)
);

CREATE TABLE IF NOT EXISTS outcomes (
  id TEXT PRIMARY KEY,
  approach_revision_id TEXT NOT NULL REFERENCES approach_revisions(id),
  class TEXT NOT NULL CHECK (class IN ('unknown', 'failure', 'partial_failure', 'partial_success', 'success')),
  boundary_statement TEXT NOT NULL DEFAULT '',
  notes TEXT NOT NULL DEFAULT '',
  UNIQUE(approach_revision_id)
);

CREATE TABLE IF NOT EXISTS failure_boundaries (
  id TEXT PRIMARY KEY,
  approach_revision_id TEXT NOT NULL REFERENCES approach_revisions(id),
  condition TEXT NOT NULL,
  ordinal INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS source_supports (
  approach_revision_id TEXT NOT NULL REFERENCES approach_revisions(id),
  snapshot_id TEXT NOT NULL REFERENCES source_snapshots(id),
  field_path TEXT NOT NULL,
  support_kind TEXT NOT NULL CHECK (support_kind IN ('explicit', 'inferred', 'unsupported')),
  locator TEXT NOT NULL DEFAULT '',
  confidence TEXT NOT NULL DEFAULT '',
  PRIMARY KEY(approach_revision_id, snapshot_id, field_path)
);

CREATE INDEX IF NOT EXISTS idx_approaches_problem ON approaches(problem_id, created_at, id);
CREATE INDEX IF NOT EXISTS idx_normalization_revisions_snapshot ON normalization_revisions(snapshot_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_normalization_revisions_problem ON normalization_revisions(problem_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_approach_revisions_approach ON approach_revisions(approach_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_approach_revisions_norm ON approach_revisions(normalization_revision_id);
CREATE INDEX IF NOT EXISTS idx_source_supports_revision ON source_supports(approach_revision_id);

-- Derived normalization artifacts are append-only revisions: never mutate or
-- delete a persisted interpretation. Re-normalization creates a new revision
-- with lineage instead of overwriting history.
CREATE TRIGGER IF NOT EXISTS provider_invocations_immutable_update
BEFORE UPDATE ON provider_invocations
BEGIN
  SELECT RAISE(ABORT, 'provider invocations are immutable');
END;
CREATE TRIGGER IF NOT EXISTS provider_invocations_immutable_delete
BEFORE DELETE ON provider_invocations
BEGIN
  SELECT RAISE(ABORT, 'provider invocations are immutable');
END;

CREATE TRIGGER IF NOT EXISTS normalization_revisions_immutable_update
BEFORE UPDATE ON normalization_revisions
BEGIN
  SELECT RAISE(ABORT, 'normalization revisions are immutable');
END;
CREATE TRIGGER IF NOT EXISTS normalization_revisions_immutable_delete
BEFORE DELETE ON normalization_revisions
BEGIN
  SELECT RAISE(ABORT, 'normalization revisions are immutable');
END;

CREATE TRIGGER IF NOT EXISTS approach_revisions_immutable_update
BEFORE UPDATE ON approach_revisions
BEGIN
  SELECT RAISE(ABORT, 'approach revisions are immutable');
END;
CREATE TRIGGER IF NOT EXISTS approach_revisions_immutable_delete
BEFORE DELETE ON approach_revisions
BEGIN
  SELECT RAISE(ABORT, 'approach revisions are immutable');
END;

CREATE TRIGGER IF NOT EXISTS mechanisms_immutable_update
BEFORE UPDATE ON mechanisms
BEGIN
  SELECT RAISE(ABORT, 'mechanisms are immutable');
END;
CREATE TRIGGER IF NOT EXISTS mechanisms_immutable_delete
BEFORE DELETE ON mechanisms
BEGIN
  SELECT RAISE(ABORT, 'mechanisms are immutable');
END;
`,
	},
	{
		version: 4,
		sql: `
-- Canonicalization vocabulary layer (#9). Versioned, immutable-per-version
-- semantic spine: models discover candidate labels; software owns identity.
CREATE TABLE IF NOT EXISTS canonical_vocabulary (
  version TEXT PRIMARY KEY,
  created_at TEXT NOT NULL,
  notes TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS canonical_terms (
  vocabulary_version TEXT NOT NULL REFERENCES canonical_vocabulary(version),
  canonical_id TEXT NOT NULL,
  field_kind TEXT NOT NULL CHECK (field_kind IN ('representation', 'assumption', 'operator', 'preserves', 'breaks', 'auxiliary_object', 'outcome', 'boundary', 'posture')),
  description TEXT NOT NULL DEFAULT '',
  parent_canonical_id TEXT,
  PRIMARY KEY(vocabulary_version, canonical_id)
);

CREATE TABLE IF NOT EXISTS canonical_term_aliases (
  vocabulary_version TEXT NOT NULL REFERENCES canonical_vocabulary(version),
  canonical_id TEXT NOT NULL,
  field_kind TEXT NOT NULL CHECK (field_kind IN ('representation', 'assumption', 'operator', 'preserves', 'breaks', 'auxiliary_object', 'outcome', 'boundary', 'posture')),
  alias_normalized TEXT NOT NULL,
  -- Uniqueness is scoped by (version, field_kind, alias_normalized, canonical_id):
  --   * the same phrase MAY map to different canonical ids in DIFFERENT field
  --     kinds (the resolver namespaces aliases by field kind on purpose), and
  --   * the same phrase MAY map to more than one canonical id WITHIN a field
  --     kind, which the resolver reports as an explicit ambiguous state.
  -- Including canonical_id in the key forbids only exact-duplicate rows, never
  -- a legitimate second binding, so no binding is ever silently dropped.
  PRIMARY KEY(vocabulary_version, field_kind, alias_normalized, canonical_id)
);

CREATE TABLE IF NOT EXISTS classification_rubrics (
  contract_version TEXT PRIMARY KEY,
  field_kind TEXT NOT NULL,
  definition_json TEXT NOT NULL,
  created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_canonical_terms_field ON canonical_terms(vocabulary_version, field_kind, canonical_id);
CREATE INDEX IF NOT EXISTS idx_canonical_term_aliases_id ON canonical_term_aliases(vocabulary_version, canonical_id);

-- Vocabulary rows are immutable per version: evolution appends a new version,
-- never rewrites history, so historical signatures stay reproducible.
CREATE TRIGGER IF NOT EXISTS canonical_vocabulary_immutable_update
BEFORE UPDATE ON canonical_vocabulary
BEGIN
  SELECT RAISE(ABORT, 'canonical vocabulary is immutable');
END;
CREATE TRIGGER IF NOT EXISTS canonical_vocabulary_immutable_delete
BEFORE DELETE ON canonical_vocabulary
BEGIN
  SELECT RAISE(ABORT, 'canonical vocabulary is immutable');
END;
CREATE TRIGGER IF NOT EXISTS canonical_terms_immutable_update
BEFORE UPDATE ON canonical_terms
BEGIN
  SELECT RAISE(ABORT, 'canonical terms are immutable');
END;
CREATE TRIGGER IF NOT EXISTS canonical_terms_immutable_delete
BEFORE DELETE ON canonical_terms
BEGIN
  SELECT RAISE(ABORT, 'canonical terms are immutable');
END;
CREATE TRIGGER IF NOT EXISTS canonical_term_aliases_immutable_update
BEFORE UPDATE ON canonical_term_aliases
BEGIN
  SELECT RAISE(ABORT, 'canonical term aliases are immutable');
END;
CREATE TRIGGER IF NOT EXISTS canonical_term_aliases_immutable_delete
BEFORE DELETE ON canonical_term_aliases
BEGIN
  SELECT RAISE(ABORT, 'canonical term aliases are immutable');
END;
`,
	},
	{
		version: 5,
		sql: `
-- Mechanism signatures (#9): derived, version-keyed, immutable projections of a
-- mechanism under a specific schema + vocabulary version. Re-canonicalizing
-- under a newer vocabulary creates a new row; the old one stays reproducible.
CREATE TABLE IF NOT EXISTS mechanism_signatures (
  id TEXT PRIMARY KEY,
  mechanism_id TEXT NOT NULL REFERENCES mechanisms(id),
  schema_version TEXT NOT NULL,
  vocabulary_version TEXT NOT NULL REFERENCES canonical_vocabulary(version),
  fingerprint TEXT NOT NULL,
  run_id TEXT NOT NULL REFERENCES runs(id),
  created_at TEXT NOT NULL,
  UNIQUE(mechanism_id, schema_version, vocabulary_version)
);

CREATE TABLE IF NOT EXISTS signature_field_claims (
  signature_id TEXT NOT NULL REFERENCES mechanism_signatures(id),
  field_kind TEXT NOT NULL,
  surface_label TEXT NOT NULL,
  resolution_state TEXT NOT NULL CHECK (resolution_state IN ('resolved', 'ambiguous', 'novel_candidate', 'unknown', 'rejected')),
  canonical_id TEXT NOT NULL DEFAULT '',
  claim_status TEXT NOT NULL,
  support_snapshot_id TEXT NOT NULL DEFAULT '',
  support_locator TEXT NOT NULL DEFAULT '',
  confidence TEXT NOT NULL DEFAULT '',
  classifier_contract TEXT NOT NULL DEFAULT '',
  ordinal INTEGER NOT NULL,
  PRIMARY KEY(signature_id, field_kind, surface_label, ordinal)
);

CREATE TABLE IF NOT EXISTS signature_postures (
  signature_id TEXT NOT NULL REFERENCES mechanism_signatures(id),
  axis TEXT NOT NULL,
  value TEXT NOT NULL,
  claim_status TEXT NOT NULL DEFAULT 'unknown',
  PRIMARY KEY(signature_id, axis)
);

CREATE TABLE IF NOT EXISTS signature_boundaries (
  signature_id TEXT NOT NULL REFERENCES mechanism_signatures(id),
  surface_label TEXT NOT NULL,
  resolution_state TEXT NOT NULL,
  canonical_id TEXT NOT NULL DEFAULT '',
  relation TEXT NOT NULL DEFAULT '',
  claim_status TEXT NOT NULL DEFAULT 'unknown',
  support_snapshot_id TEXT NOT NULL DEFAULT '',
  support_locator TEXT NOT NULL DEFAULT '',
  ordinal INTEGER NOT NULL,
  PRIMARY KEY(signature_id, ordinal)
);

CREATE TABLE IF NOT EXISTS signature_outcomes (
  signature_id TEXT PRIMARY KEY REFERENCES mechanism_signatures(id),
  class TEXT NOT NULL,
  claim_status TEXT NOT NULL DEFAULT 'unknown'
);

CREATE INDEX IF NOT EXISTS idx_mechanism_signatures_mechanism ON mechanism_signatures(mechanism_id);

CREATE TRIGGER IF NOT EXISTS mechanism_signatures_immutable_update
BEFORE UPDATE ON mechanism_signatures
BEGIN
  SELECT RAISE(ABORT, 'mechanism signatures are immutable');
END;
CREATE TRIGGER IF NOT EXISTS mechanism_signatures_immutable_delete
BEFORE DELETE ON mechanism_signatures
BEGIN
  SELECT RAISE(ABORT, 'mechanism signatures are immutable');
END;
CREATE TRIGGER IF NOT EXISTS signature_field_claims_immutable_update
BEFORE UPDATE ON signature_field_claims
BEGIN
  SELECT RAISE(ABORT, 'signature field claims are immutable');
END;
CREATE TRIGGER IF NOT EXISTS signature_field_claims_immutable_delete
BEFORE DELETE ON signature_field_claims
BEGIN
  SELECT RAISE(ABORT, 'signature field claims are immutable');
END;
CREATE TRIGGER IF NOT EXISTS signature_postures_immutable_update
BEFORE UPDATE ON signature_postures
BEGIN
  SELECT RAISE(ABORT, 'signature postures are immutable');
END;
CREATE TRIGGER IF NOT EXISTS signature_postures_immutable_delete
BEFORE DELETE ON signature_postures
BEGIN
  SELECT RAISE(ABORT, 'signature postures are immutable');
END;
CREATE TRIGGER IF NOT EXISTS signature_boundaries_immutable_update
BEFORE UPDATE ON signature_boundaries
BEGIN
  SELECT RAISE(ABORT, 'signature boundaries are immutable');
END;
CREATE TRIGGER IF NOT EXISTS signature_boundaries_immutable_delete
BEFORE DELETE ON signature_boundaries
BEGIN
  SELECT RAISE(ABORT, 'signature boundaries are immutable');
END;
CREATE TRIGGER IF NOT EXISTS signature_outcomes_immutable_update
BEFORE UPDATE ON signature_outcomes
BEGIN
  SELECT RAISE(ABORT, 'signature outcomes are immutable');
END;
CREATE TRIGGER IF NOT EXISTS signature_outcomes_immutable_delete
BEFORE DELETE ON signature_outcomes
BEGIN
  SELECT RAISE(ABORT, 'signature outcomes are immutable');
END;
`,
	},
	{
		version: 6,
		sql: `
-- Comparison runs (#9): component-wise, versioned comparison results. No single
-- scalar; per-field results are stored explicitly.
CREATE TABLE IF NOT EXISTS comparison_runs (
  id TEXT PRIMARY KEY,
  signature_a_id TEXT NOT NULL REFERENCES mechanism_signatures(id),
  signature_b_id TEXT NOT NULL REFERENCES mechanism_signatures(id),
  weights_version TEXT NOT NULL,
  classify_version TEXT NOT NULL,
  classification TEXT NOT NULL,
  run_id TEXT NOT NULL REFERENCES runs(id),
  created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS comparison_field_results (
  comparison_run_id TEXT NOT NULL REFERENCES comparison_runs(id),
  field_kind TEXT NOT NULL,
  overlap_count INTEGER NOT NULL,
  union_count INTEGER NOT NULL,
  jaccard REAL NOT NULL,
  ordinal TEXT NOT NULL,
  incomparable INTEGER NOT NULL,
  PRIMARY KEY(comparison_run_id, field_kind)
);

CREATE INDEX IF NOT EXISTS idx_comparison_runs_sigs ON comparison_runs(signature_a_id, signature_b_id);

CREATE TRIGGER IF NOT EXISTS comparison_runs_immutable_update
BEFORE UPDATE ON comparison_runs
BEGIN
  SELECT RAISE(ABORT, 'comparison runs are immutable');
END;
CREATE TRIGGER IF NOT EXISTS comparison_runs_immutable_delete
BEFORE DELETE ON comparison_runs
BEGIN
  SELECT RAISE(ABORT, 'comparison runs are immutable');
END;
CREATE TRIGGER IF NOT EXISTS comparison_field_results_immutable_update
BEFORE UPDATE ON comparison_field_results
BEGIN
  SELECT RAISE(ABORT, 'comparison field results are immutable');
END;
CREATE TRIGGER IF NOT EXISTS comparison_field_results_immutable_delete
BEFORE DELETE ON comparison_field_results
BEGIN
  SELECT RAISE(ABORT, 'comparison field results are immutable');
END;
`,
	},
	{
		version: 7,
		sql: `
-- Mechanism clustering (#11): a deterministic, version-keyed grouping pass over
-- the signatures of one problem under one (schema_version, vocabulary_version).
-- Immutable + idempotent on the full version tuple: re-running the same pass
-- returns the existing run rather than rewriting history. No embeddings: the
-- grouping is reproducible from comparison verdicts + the recorded profile.
CREATE TABLE IF NOT EXISTS cluster_runs (
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
  created_at TEXT NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_cluster_runs_identity
  ON cluster_runs(problem_id, schema_version, vocabulary_version, profile_version, cluster_algo_version, thresholds_hash);

CREATE TABLE IF NOT EXISTS mechanism_clusters (
  id TEXT PRIMARY KEY,
  cluster_run_id TEXT NOT NULL REFERENCES cluster_runs(id),
  cluster_fingerprint TEXT NOT NULL,
  representative_signature_id TEXT NOT NULL REFERENCES mechanism_signatures(id),
  member_count INTEGER NOT NULL,
  intra_variation TEXT NOT NULL,
  isolate INTEGER NOT NULL DEFAULT 0,
  ordinal INTEGER NOT NULL,
  UNIQUE(cluster_run_id, cluster_fingerprint)
);

CREATE TABLE IF NOT EXISTS cluster_members (
  cluster_id TEXT NOT NULL REFERENCES mechanism_clusters(id),
  signature_id TEXT NOT NULL REFERENCES mechanism_signatures(id),
  mechanism_id TEXT NOT NULL REFERENCES mechanisms(id),
  redundant INTEGER NOT NULL DEFAULT 0,
  ordinal INTEGER NOT NULL,
  PRIMARY KEY(cluster_id, signature_id)
);

CREATE TABLE IF NOT EXISTS cluster_distances (
  cluster_run_id TEXT NOT NULL REFERENCES cluster_runs(id),
  cluster_a_id TEXT NOT NULL REFERENCES mechanism_clusters(id),
  cluster_b_id TEXT NOT NULL REFERENCES mechanism_clusters(id),
  classification TEXT NOT NULL,
  PRIMARY KEY(cluster_run_id, cluster_a_id, cluster_b_id)
);

CREATE TABLE IF NOT EXISTS cluster_coverage_axes (
  cluster_run_id TEXT NOT NULL REFERENCES cluster_runs(id),
  axis TEXT NOT NULL,
  distinct_value_count INTEGER NOT NULL,
  under_sampled INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY(cluster_run_id, axis)
);

CREATE TABLE IF NOT EXISTS cluster_discrimination_losses (
  cluster_run_id TEXT NOT NULL REFERENCES cluster_runs(id),
  mechanism_a_id TEXT NOT NULL,
  mechanism_b_id TEXT NOT NULL,
  outcome_a TEXT NOT NULL,
  outcome_b TEXT NOT NULL,
  ordinal INTEGER NOT NULL,
  PRIMARY KEY(cluster_run_id, ordinal)
);

CREATE INDEX IF NOT EXISTS idx_cluster_runs_problem ON cluster_runs(problem_id);
CREATE INDEX IF NOT EXISTS idx_mechanism_clusters_run ON mechanism_clusters(cluster_run_id);

CREATE TRIGGER IF NOT EXISTS cluster_runs_immutable_update
BEFORE UPDATE ON cluster_runs
BEGIN
  SELECT RAISE(ABORT, 'cluster runs are immutable');
END;
CREATE TRIGGER IF NOT EXISTS cluster_runs_immutable_delete
BEFORE DELETE ON cluster_runs
BEGIN
  SELECT RAISE(ABORT, 'cluster runs are immutable');
END;
CREATE TRIGGER IF NOT EXISTS mechanism_clusters_immutable_update
BEFORE UPDATE ON mechanism_clusters
BEGIN
  SELECT RAISE(ABORT, 'mechanism clusters are immutable');
END;
CREATE TRIGGER IF NOT EXISTS mechanism_clusters_immutable_delete
BEFORE DELETE ON mechanism_clusters
BEGIN
  SELECT RAISE(ABORT, 'mechanism clusters are immutable');
END;
CREATE TRIGGER IF NOT EXISTS cluster_members_immutable_update
BEFORE UPDATE ON cluster_members
BEGIN
  SELECT RAISE(ABORT, 'cluster members are immutable');
END;
CREATE TRIGGER IF NOT EXISTS cluster_members_immutable_delete
BEFORE DELETE ON cluster_members
BEGIN
  SELECT RAISE(ABORT, 'cluster members are immutable');
END;
`,
	},
	{
		version: 8,
		sql: `
-- FailureSpace (#11 / EPIC M3): the first explicit, versioned failure-space
-- artifact materialized from a cluster run. Partitioned by the outcome class
-- already carried on signatures; preserves cluster coverage and provenance.
-- Immutable, revisioned per problem: a new cluster run yields a new revision,
-- never a rewrite.
CREATE TABLE IF NOT EXISTS failure_spaces (
  id TEXT PRIMARY KEY,
  problem_id TEXT NOT NULL REFERENCES problems(id),
  cluster_run_id TEXT NOT NULL REFERENCES cluster_runs(id),
  run_id TEXT NOT NULL REFERENCES runs(id),
  revision INTEGER NOT NULL,
  distinct_family_count INTEGER NOT NULL,
  redundant_member_count INTEGER NOT NULL,
  created_at TEXT NOT NULL,
  UNIQUE(problem_id, cluster_run_id),
  UNIQUE(problem_id, revision)
);

CREATE TABLE IF NOT EXISTS failure_space_outcomes (
  failure_space_id TEXT NOT NULL REFERENCES failure_spaces(id),
  outcome_class TEXT NOT NULL,
  family_count INTEGER NOT NULL,
  PRIMARY KEY(failure_space_id, outcome_class)
);

CREATE TABLE IF NOT EXISTS failure_space_axes (
  failure_space_id TEXT NOT NULL REFERENCES failure_spaces(id),
  axis_kind TEXT NOT NULL,
  distinct_value_count INTEGER NOT NULL,
  under_sampled INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY(failure_space_id, axis_kind)
);

CREATE INDEX IF NOT EXISTS idx_failure_spaces_problem ON failure_spaces(problem_id);

CREATE TRIGGER IF NOT EXISTS failure_spaces_immutable_update
BEFORE UPDATE ON failure_spaces
BEGIN
  SELECT RAISE(ABORT, 'failure spaces are immutable');
END;
CREATE TRIGGER IF NOT EXISTS failure_spaces_immutable_delete
BEFORE DELETE ON failure_spaces
BEGIN
  SELECT RAISE(ABORT, 'failure spaces are immutable');
END;
CREATE TRIGGER IF NOT EXISTS failure_space_outcomes_immutable_update
BEFORE UPDATE ON failure_space_outcomes
BEGIN
  SELECT RAISE(ABORT, 'failure space outcomes are immutable');
END;
CREATE TRIGGER IF NOT EXISTS failure_space_outcomes_immutable_delete
BEFORE DELETE ON failure_space_outcomes
BEGIN
  SELECT RAISE(ABORT, 'failure space outcomes are immutable');
END;
CREATE TRIGGER IF NOT EXISTS failure_space_axes_immutable_update
BEFORE UPDATE ON failure_space_axes
BEGIN
  SELECT RAISE(ABORT, 'failure space axes are immutable');
END;
CREATE TRIGGER IF NOT EXISTS failure_space_axes_immutable_delete
BEFORE DELETE ON failure_space_axes
BEGIN
  SELECT RAISE(ABORT, 'failure space axes are immutable');
END;
`,
	},
	{
		version: 9,
		// In-place repair for databases created under the earlier v6 schema.
		// Migrations 4 and 5 were corrected in place (alias PK gained
		// canonical_id; signature posture/boundary/outcome tables gained
		// claim_status/support columns). A database that already recorded
		// migrations 1-6 as applied will never re-run 4/5, so it retains the
		// pre-fix schema. This step upgrades such a database and also adds the
		// durable rejected-terms table. It is introspective and idempotent: on a
		// fresh database (already correct via 4/5) every check finds the target
		// state already present and does nothing.
		apply: migrateV9RepairAndRejected,
	},
	{
		version: 10,
		// Cluster-run identity hardening (#11 review): (1) add input_set_hash to
		// cluster_runs and fold it into the idempotency UNIQUE so re-clustering a
		// changed signature population produces a new run instead of colliding
		// with a stale run (KTD-1); (2) add per-family outcome_class + outcome_mixed
		// to mechanism_clusters so a family's outcome is reported as its member
		// distribution ("mixed" when heterogeneous) rather than compressed through
		// the representative signature (KTD-9). Introspective + idempotent: a fresh
		// database created at this version already has the target shape and the
		// checks are no-ops.
		apply: migrateV10ClusterIdentityAndOutcome,
	},
	{
		version: 11,
		// Candidate-invariant mining (#12 / M4.2). Additive new tables with
		// immutability triggers PLUS one guarded generalization of the
		// provider_invocations.role CHECK from ('normalize') to
		// ('normalize','invariant') so a mining invocation can be recorded in the
		// same auditable invocation table (KTD-8a). The generalization edits the
		// table CHECK in place via writable_schema (NOT a DROP+RENAME rebuild,
		// which is not FK-safe here because normalization_revisions references
		// provider_invocations and the deferred child FK fails at commit after
		// the parent is dropped). Introspective + idempotent: a fresh v11 database
		// already permits 'invariant' and the edit is skipped.
		apply: migrateV11CandidateInvariants,
	},
	{
		version: 12,
		// M4.2 semantic hardening (F2). Two additive changes to candidate_invariants:
		// (1) add contrast_violating_num + contrast_eligible_den so contrast is
		// recorded as complete counts rather than collapsed into the boolean the
		// association_status carried; (2) rename the association_status enum value
		// 'discriminative' -> 'contrast_observed' in the CHECK, because a single
		// violating contrast family does not establish directional discrimination.
		// The columns are plain ADD COLUMN (FK-safe). The CHECK is edited in place
		// via writable_schema (never a DROP+RENAME parent rebuild, which is not
		// FK-safe: invariant_predicates/_family_evaluations/_counterexamples all
		// reference candidate_invariants). Introspective + idempotent: a fresh v12
		// database already has the target shape and the checks are no-ops.
		apply: migrateV12ContrastCountsAndRename,
	},
	{
		version: 13,
		// M4.3 challenge/falsify lifecycle (#13). Additive challenge, synthetic-
		// artifact, lineage, and append-only state-transition tables with the
		// blueprint's transition-validating trigger and the invariant_current_state
		// read view (docs/persistence.md, names pluralized to the shipped
		// convention; challenge evidence links reference the real persisted rows —
		// clusters/signatures/snapshots — because the sketched evidence_record
		// table was never shipped). Also widens provider_invocations.role to admit
		// 'challenge' via the same guarded in-place writable_schema CHECK edit v11
		// used (FK-safe by construction; normalization_revisions holds live FKs
		// into the table). Introspective + idempotent: a fresh v13 database already
		// has the target shape and every step is a no-op.
		apply: migrateV13ChallengeLifecycle,
	},
	{
		version: 14,
		// M5.1 frontier generation (#14). Additive frontier_generation_runs,
		// frontier_proposals, frontier_target_invariants, and
		// frontier_nearest_clusters tables (docs/persistence.md 534-569, names
		// pluralized to the shipped convention; the nullable holdout_leakage_check
		// link is deferred to M7 and not created here). frontier_proposals dedups
		// on UNIQUE(problem_id, proposal_hash) and leaves `result` NULL for M5.2 to
		// populate. Also widens provider_invocations.role to admit 'generate' via
		// the same guarded in-place writable_schema CHECK edit v11/v13 used
		// (FK-safe by construction). Introspective + idempotent: a fresh v14
		// database already has the target shape and every step is a no-op.
		apply: migrateV14FrontierGeneration,
	},
	{
		version: 15,
		// M5.2 evaluation + verifier routing (#15). Additive evaluation_runs,
		// evaluations (+ verifier_kind + verification_strength, KTD-1), the
		// evaluation_metrics / evaluation_run_metrics metric-scale triads, and an
		// evaluated_failures re-entry marker (R6/KTD-5). Transcribed from
		// docs/persistence.md 670-853 (names pluralized to the shipped convention),
		// mode='proposal' only: the holdout-only tables (holdout_set*,
		// evaluation_holdout_match) and baseline arms are deferred to M7, so this
		// migration ships the nullable holdout columns plus a holdout-REFUSAL gate
		// trigger (a verbatim copy of the blueprint gate is impossible without the
		// M7 holdout_leakage_check/holdout_set tables it queries; M7 replaces the
		// refusal trigger with the full gate when it adds those tables). Also widens
		// provider_invocations.role to admit 'evaluate' via the same guarded in-place
		// writable_schema CHECK edit v11/v13/v14 used. Introspective + idempotent.
		apply: migrateV15Evaluation,
	},
	{
		version: 16,
		// F1 epistemic honesty: rename the invariant lifecycle terminal state
		// `established` to `operator_attested`. The prior name asserted a
		// machine-confirmed, independently-verified claim, but the establish path
		// only records an operator-supplied snapshot + locator string — it does
		// not inspect the snapshot content, validate the locator, establish
		// independence, or check a proof against the predicate. Renaming keeps the
		// useful operator-attestation provenance while removing the false
		// "verified" claim (AGENTS.md: ModelJudgment != Verification; no silent
		// status promotion). Rebuilds the transition-validating trigger and the
		// invariant_state_transitions.to_state CHECK, and rewrites any existing
		// `established` transition rows to `operator_attested`. Introspective +
		// idempotent: a fresh v16 database already uses the new name.
		apply: migrateV16RenameEstablishedToOperatorAttested,
	},
	{
		version: 17,
		// M6.1 success compression (#16). Additive: (1) frontier_proposal_signatures
		// closes the verified substrate gap that a proposal's canonical CONTENT was
		// never persisted (only its hash) — success compression must evaluate
		// condition predicates against successful proposals; (2) the immutable,
		// revisioned success-invariant layer (revisions, invariants, predicates,
		// broken-target links, cohort evaluations). One guarded non-additive step:
		// widens provider_invocations.role to admit 'success-compress' via the
		// established in-place writable_schema CHECK edit (FK-safe; v11/v13/v14/v15
		// precedent). Introspective + idempotent: a fresh v17 database already has
		// the target shape.
		apply: migrateV17SuccessCompression,
	},
	{
		version: 18,
		// M4.3 lifecycle hardening (G2/G4). Widens the transition-validating
		// trigger's legal-edge set so undecided campaigns are resumable and
		// operator attestation stays challengeable:
		//   challenged        -> challenged   (re-open an undecided campaign)
		//   operator_attested -> challenged   (attestation is not immune to challenge)
		//   operator_attested -> weaken|falsified (a challenge can overturn it)
		// DROP + CREATE the trigger (immutable-by-convention). Introspective +
		// idempotent: recreating with the wider rule is safe on a fresh v18 DB.
		apply: migrateV18ChallengeableAttestationAndResume,
	},
	{
		version: 19,
		// M6.2 search-policy mutation (#18). Additive, immutable, revisioned
		// search_policy_revisions / search_policy_directives / search_policy_provenance:
		// a per-problem SearchPolicy derived by code from persisted evidence
		// (success invariants, surviving failure invariants, coverage gaps,
		// redundancy / repeated-failure) that biases the NEXT frontier generation
		// as a bounded ordinal transform (never crossing the code-verified
		// violation gate). Also a frontier_generation_policy child recording which
		// policy revision biased a generation and the per-proposal applied bias
		// (the reproducible "why favored/suppressed" surface). One guarded
		// non-additive step: widens provider_invocations.role to admit
		// 'policy-mutate' via the established in-place writable_schema CHECK edit
		// (FK-safe; v11/v13/v14/v15/v17 precedent). Introspective + idempotent.
		apply: migrateV19SearchPolicy,
	},
	{
		version: 20,
		// M6.1 hardening (H1): success-compression cohort evaluations record the
		// EVALUATION_ID their (verdict, strength) pair came from, so provenance is a
		// single coherent evaluation record rather than a result-cache spliced onto
		// a later evaluation's strength. Additive nullable column; INSERT is allowed
		// by the immutability triggers (they guard UPDATE/DELETE only). Idempotent:
		// skip when the column already exists.
		apply: migrateV20CohortEvaluationProvenance,
	},
	{
		version: 21,
		// M7 v0 blinded-benchmark experiment (#19). Additive: holdout_sets
		// (mode blinded|historical; historical requires a cutoff and stays
		// execution-refused without per-source dating evidence), withheld-source
		// links with problem-guard triggers, holdout_source_dating,
		// leakage_checks (the code-computed quarantine audit), and the
		// experiment_runs / experiment_arms / experiment_metrics layer with
		// mode-disjoint conclusion vocabularies (BlindedRecovery !=
		// HistoricalPrediction, enforced by CHECK). Also REPLACES the M5.2
		// blanket holdout-REFUSAL gate on evaluation_runs with the reserved
		// leakage-keyed condition (holdout mode requires a passing leakage check
		// bound to the same holdout set), and widens provider_invocations.role
		// to admit the baseline roles 'summarize-next' and 'brainstorm'
		// (guarded in-place writable_schema edit). Introspective + idempotent.
		apply: migrateV21BlindedExperiment,
	},
	{
		version: 22,
		// M7 measurement-contract hardening (review findings 1-3, 5). (1)
		// experiment_arm_proposals: explicit arm<->proposal MEMBERSHIP with rank
		// and per-proposal assessment — artifact deduplication and experiment
		// participation are different identities; an arm is scored ONLY on what
		// it derived this run (dedup may map a derived proposal to an existing
		// artifact id, recorded explicitly). (2) experiment_targets: the FROZEN
		// target manifest (signature ids + content fingerprints) that audit,
		// scoring, and experiment identity all consume — scored targets are
		// restricted to signatures derived from the REGISTERED withheld sources.
		// (3) experiment_arms gains decisive/unknown/unassessed counts and
		// evaluations_consumed so the evaluation budget is enforced and its
		// consumption persisted, and unknown-only assessments can conclude
		// inconclusive rather than being coerced to no_recovery. Additive +
		// guarded; idempotent on fresh v22.
		apply: migrateV22ExperimentMeasurementContract,
	},
	{
		version: 23,
		// Evidence-revision separation (review F1). The mechanism FINGERPRINT
		// deliberately excludes extraction completeness and unresolved claims —
		// correct for mechanism identity/comparison, but those fields change
		// predicate evaluation (unobserved field => unknown; exhaustively
		// complete => violates). Proposal dedup keys on that fingerprint, and
		// the v17 sidecar is single-row INSERT OR IGNORE, so a REVISED
		// interpretation with identical resolved content was silently dropped.
		// v23 adds frontier_proposal_signature_revisions: an immutable,
		// append-only CONTENT revision per (proposal, sha256(signature_json)),
		// backfilled from the v17 sidecar as revision 1. Readers consume the
		// LATEST revision; verdicts and experiment arm assessments reference
		// the exact content hash they evaluated (new guarded columns on
		// evaluations + experiment_arm_proposals). The v17 sidecar remains as
		// the frozen first-observed content.
		apply: migrateV23SignatureContentRevisions,
	},
	{
		version: 24,
		// Occurrence-level revision binding (review round 2). v23 preserved
		// revised content but read it back through MAX(revision) lookups,
		// which (1) let an evaluation stamp a hash its verifier never
		// assessed, (2) let compression splice an old evaluation onto a
		// newer unassessed revision, and (3) mis-attributed content when a
		// generation re-emitted earlier content (A -> B -> A returns B).
		// v24 adds frontier_generation_contents: an immutable binding
		// generation -> (proposal, EMITTED content hash), backfilled for each
		// proposal's ORIGIN generation from revision 1 (dedup occurrences on
		// other generations are historically unrecoverable — readers fall
		// back to the latest revision and say so). success_invariant_revisions
		// gains pending_reassessment: members whose newest interpretation has
		// no compatible reassessment are excluded from current guidance and
		// counted, never silently carried forward.
		apply: migrateV24OccurrenceBindings,
	},
	{
		version: 25,
		// Justified field-completeness declarations (persisted-input positive
		// control). The signature builder's conservative default marks every
		// set field `unobserved`, so `contains`-absence evaluates unknown. An
		// extractor that can HONESTLY declare a field exhaustively extracted
		// (e.g. "all entries of the declared payload's preserves list were
		// parsed") now persists that declaration per (mechanism, field) with a
		// REQUIRED basis, and the signature build/rehydrate path carries it —
		// making absence-based verified negatives reachable from persisted
		// inputs without letting an unqualified provider assert completeness.
		apply: migrateV25FieldCompleteness,
	},
	{
		version: 26,
		// Declared vs ACCEPTED completeness + assessment-context break verdicts
		// (review of 87759d9, findings 1 and 3).
		// (1) mechanism_field_completeness gains scope / admission /
		// admission_basis: the provider supplies a typed scope + basis (a
		// CLAIM); code decides admission. Only accepted declarations influence
		// predicate evaluation; declared_only rows are auditable claims with
		// no evaluation authority. Existing v25 rows are backfilled
		// declared_only (preserve the weaker type; re-normalize to re-admit).
		// (2) evaluation_target_verdicts: the per-target break verdicts an
		// evaluation ACTUALLY computed against its assessed content revision,
		// so success-cohort admission can follow the selected assessment
		// context instead of the origin-time frontier_target_invariants flag.
		// Existing evaluations are backfilled from the origin rows they
		// historically consumed.
		apply: migrateV26CompletenessAdmissionAndTargetVerdicts,
	},
	{
		version: 27,
		// Legacy-binding honesty (review of a47dd24, finding 1).
		// evaluation_target_verdicts gains a provenance column and existing
		// rows whose evaluation binding is NOT establishable — no recorded
		// assessed hash on a proposal with multiple retained revisions — are
		// reclassified 'unverified_legacy': retained for inspection, excluded
		// from cohort admission. Establishable bindings (assessed hash equals
		// a retained revision, or hash-less on a single-revision proposal)
		// stay 'recomputed'. Pre-v26 evaluations could not reach revised
		// occurrences (selection resolved the owning generation), so their
		// origin-flag backfills with establishable bindings remain valid.
		apply: migrateV27TargetVerdictProvenance,
	},
	{
		version: 28,
		// Compression execution/selection log (review of 040b8c9, finding 1).
		// PersistSuccessRevision deduplicates by cohort/configuration identity
		// (correct), but "current guidance" was read as MAX(revision) — so
		// after support appeared (R2) and then disappeared (reusing empty R1),
		// policy still consumed R2's stale support. v28 separates the
		// immutable artifact from the operation that SELECTED it: every
		// compression execution appends a selection row (also when the
		// artifact is reused), and policy consumes the latest selection.
		// Existing revisions are backfilled as their own creating selection.
		apply: migrateV28CompressionSelections,
	},
	{
		version: 29,
		// Admission audit persistence (follow-on P3 from the 040b8c9 review
		// series). The proposal-admission boundary computed corrected /
		// downgraded / stripped / rejected counts and dropped them; operators
		// could only reconstruct them by diffing raw provider payloads. v29
		// persists the per-generation aggregate on frontier_generation_runs
		// (zero for trusted code-derived generators, which bypass admission).
		apply: migrateV29AdmissionAudit,
	},
	{
		version: 30,
		// Policy mutation selection log (P5 audit finding 1): search-policy
		// revisions dedup-reuse by evidence cohort identity exactly like
		// success revisions, and generation-time application read
		// MAX(revision) — so evidence that REVERTED to an earlier cohort
		// reused the older revision while generation kept applying the
		// higher-numbered stale one. Same fix as v28: every mutation
		// execution appends a selection row (also on reuse), and current
		// policy follows the latest selection. Backfilled from existing
		// revisions (their creating executions).
		apply: migrateV30PolicySelections,
	},
	{
		version: 31,
		// External-proposal import hardening (512bc54 review, finding 2):
		// frontier_generation_runs gains admission_overflow — proposals an
		// untrusted provider supplied beyond the requested count, which the
		// pipeline deterministically truncates (wire order) BEFORE scoring
		// rather than leaving the bound to provider cooperation. The raw
		// response retains the full set.
		apply: migrateV31AdmissionOverflow,
	},
	{
		version: 32,
		// Execution-vs-assessment attribution (c860720 review, finding 2).
		// Experiment identity keys on the ASSESSED structure, so a changed
		// external capture whose assessed structure is unchanged (e.g. only
		// prose that ProposalHash excludes) correctly REUSES the experiment
		// artifact — but it is still a different execution of a different
		// capture. experiment_executions records, per run and arm, which
		// experiment that execution selected, which frontier generation it
		// actually produced, and the sha256 of the supplied proposals file —
		// so reuse never obscures which capture was assessed.
		apply: migrateV32ExperimentExecutions,
	},
	{
		version: 33,
		// v33 (pilot-001 mapping review): interpretation_claims stores
		// operator-adjudicated, ledger-provenanced GeneratedInterpretation
		// claims attached to a specific mechanism. They exist so an accepted
		// shared-property hypothesis (e.g. adjudication-ledger L1/L3/L4) can
		// enter signature construction as an explicit ClaimInferred claim
		// WITHOUT editing frozen source notes, altering per-field support
		// rows, or being disguised as a label alias. Rows are immutable;
		// status is never stored because code fixes it to `inferred` at
		// signature-build time — an interpretation can never carry or acquire
		// explicit source-backed status.
		apply: migrateV33InterpretationClaims,
	},
	{
		version: 34,
		// v34 (2026-09-12 structural review, finding S1): a challenge campaign
		// must identify BOTH the discovery population its claim was mined over
		// and the assessment population its evidence searches actually ran
		// against. Before this, challengeOne froze every re-challenge to the
		// original cluster run, so newly ingested and reclustered evidence
		// never entered the known-counterexample check while frontier
		// generation advanced on the latest map. The table records the two
		// cluster-run identities plus the requested population policy per
		// campaign (run_id, invariant_id); rows are immutable.
		apply: migrateV34ChallengeAssessmentPopulations,
	},
	{
		version: 35,
		// v35 (2026-09-12 structural review, finding S2): evidence admission.
		// `evaluated_failures` is a re-entry MARKER, not an admitted
		// observation — before this migration no code path carried an
		// evaluated failure back into the atlas population that clustering
		// and mining consume. The `evidence_admissions` ledger records, per
		// evaluated failure, an explicit typed admission decision:
		// admitted (materialized into the atlas as an approach revision +
		// mechanism + signature holding the EXACT assessed signature content)
		// or withheld (with the rule that refused it). Decisions distinguish
		// observation kinds — a description that failed its own structural
		// claim, a domain-checked failed attempt, and a model-judged failure
		// are different observations with different admission rules — and
		// record whether the rule or an operator admitted the row. Rows are
		// immutable; an evaluation may be withheld once and later admitted by
		// operator attestation (both rows persist, showing supersession), but
		// never admitted twice.
		apply: migrateV35EvidenceAdmissions,
	},
	{
		version: 36,
		// v36 (2026-09-12 structural review, finding S5): typed boundary
		// deltas. The glossary mapped `boundary_delta` to a ChallengeResult
		// field that did not exist — a refinement described in a transcript
		// was not a typed object subsequent policy could consume. Each
		// CONFIRMED challenge now derives, in code, the minimal separating
		// condition it observed (counterexample separation, contrast
		// collapse, constructibility, support recount, split partition,
		// merge union) and persists it here, one row per challenge,
		// immutable. Disposition stays on the challenge's transitions and
		// derived candidate ids on its derived-children rows.
		apply: migrateV36ChallengeBoundaryDeltas,
	},
	{
		version: 37,
		// v37 (2026-09-12 structural review, finding S5 part B): the typed
		// projection chain. A structural description plus prose is not a
		// concrete construction; the four records are now separate:
		// frontier_proposals (proposed structural change, existing) ->
		// projection_artifacts (the authored concrete plan: typed steps with
		// requires/provides tokens) -> projection_obligations (what must be
		// verified: a code-decided steps-compose check; an open
		// domain-realization obligation for an external checker) ->
		// projection_obligation_decisions (append-once verdicts whose
		// evidence_ref points at the domain observation, e.g. an evaluation).
		// A plan whose steps cannot compose is refuted deterministically at
		// projection time, before any domain work. All rows immutable.
		apply: migrateV37ProjectionChain,
	},
	{
		version: 38,
		// v38 (2026-09-12 semantic review, issue #21): the verification
		// SUBJECT axis. A deterministic check of a predicate over a
		// normalized signature establishes something about that signature —
		// not that it faithfully describes a realizable mechanism, nor that
		// the mechanism satisfies the domain goal. Evaluations and challenge
		// evidence now record WHAT OBJECT the verdict is about (annotation /
		// realization / domain-goal) alongside checker kind and strength, so
		// a truthful `deterministic` label cannot be read as certifying the
		// wrong object. Nullable: NULL/'' = pre-v38 history whose subject was
		// never recorded (the reader reports it as unrecorded rather than the
		// migration fabricating one retroactively).
		apply: migrateV38VerificationSubject,
	},
	{
		version: 39,
		// v39 (2026-09-12 semantic review, issue #21): the AUTHORED claim
		// form. Sample recurrence, transformation invariance, and obstruction
		// are distinct propositions; full coverage of a finite sample must
		// not supply an unstated universal domain. A candidate's universal-
		// counterexample treatment (falsifiable-by-one) is now granted ONLY
		// by an operator-authored claim form (quantifier + scope + role +
		// basis), never inferred from measured coverage. Authoring a
		// universal quantifier does not strengthen evidence — it makes the
		// claim MORE falsifiable and records who fixed its refutation
		// semantics. Append-only; latest form wins deterministically.
		apply: migrateV39InvariantClaimForms,
	},
	{
		version: 40,
		// v40 (2026-09-12 review F6): typed epistemic strength on projection
		// obligation discharge decisions. An operator discharge is backed by
		// a domain observation (an evaluation), but the observation's verdict
		// and verification strength were previously embedded only in the
		// prose basis — a policy consumer could not distinguish a discharge
		// backed by a reproducible computation from one backed by a single
		// model judgment without parsing prose. The verdict/strength pair is
		// now a typed column pair, copied verbatim from the backing
		// evaluation at decision time. Nullable: NULL/'' = code decisions
		// (steps-compose has no backing evaluation) and pre-v40 history.
		apply: migrateV40ObligationDecisionStrength,
	},
	{
		version: 41,
		// v41 (2026-09-12 review, epistemic finding; issue #23 consumer): the
		// prospective two-observation episode protocol. The reviewer's
		// minimum persuasive demonstration is a closed loop — freeze map,
		// action, prediction, and scoring rule; obtain an EXTERNALLY CHECKED
		// outcome; revise the map; commit to a DIFFERENT next action; obtain
		// and score the next outcome. The first miss remains a miss; the
		// revision earns credit only on later evidence. Four records enforce
		// that order by construction: episodes (frozen preregistration),
		// episode_commitments (append-only, one per step, committed BEFORE
		// outcomes), episode_observations (append-once per commitment,
		// domain-goal subject, reproducible strength, code-scored hit/miss),
		// episode_revisions (the map change between steps, recorded before
		// step 2 may be committed).
		apply: migrateV41Episodes,
	},
	{
		version: 42,
		// v42 (issue #22, decision D3): the projection/v2 semantic-preservation
		// contract. Widens the projection_obligations.kind CHECK to admit
		// 'semantic-preservation' — the obligation a v2 artifact owes for its
		// claimed preserved properties and correspondence class, decided only
		// by an external observation (checker_kind vocabulary unchanged:
		// 'external' already covers it). The widening uses the established
		// in-place writable_schema CHECK edit (FK-safe by construction;
		// projection_obligation_decisions holds live FKs into the table).
		// Introspective + idempotent: a DDL already admitting the kind is
		// left untouched. Artifact schema_version has no CHECK, so
		// 'projection/v2' rows need no DDL change.
		apply: migrateV42SemanticPreservationObligations,
	},
	{
		version: 43,
		// v43 (decision D5): boundary deltas become the FIRST persisted
		// next-decision edge consumed by search-policy derivation. Widens the
		// search_policy_directives.target_kind CHECK to admit
		// 'refuted_boundary' — an expand directive at a predicate fingerprint
		// whose believed failure structure a confirmed challenge PROVED
		// violable (separation-class deltas only; doctrine in
		// policy.ExpansionBearingDelta). Same FK-safe in-place CHECK edit as
		// v42; introspective + idempotent. The other two candidate edges
		// (projection obligations, episode outcomes) stay recorded-but-
		// unconsumed with their own triggers.
		apply: migrateV43RefutedBoundaryDirectives,
	},
	{
		version: 44,
		// v44 (G1, 2026-09-12 review-flow run): the NORMATIVE REVIEW LEDGER —
		// the review contract's four record responsibilities, plus the
		// versioned inputs they cite. Additive and fully separate from the
		// scientific tables: an obligation REQUIRES a property, whereas a
		// candidate invariant CLAIMS a regularity over a conditioned
		// population, so no scientific row is reused, relabeled, or promoted.
		// The four records are applicability decisions, assessments, check
		// attempts and dependency manifests; every one is immutable, and there
		// is deliberately NO editable coverage-status column anywhere —
		// coverage is generated from these records.
		sql: reviewLedgerSQL,
	},
	{
		version: 45,
		// v45 (F-1, 2026-09-12 review run remediation handoff 2): the
		// evaluated_failures re-entry marker becomes PER-EVALUATION. The
		// original PK was proposal_id with INSERT OR IGNORE, so a proposal
		// whose first failure was model-judged (withheld) kept that first
		// marker forever: a later witness-checked (reproducible-strength)
		// failure of the SAME proposal never reached admission — stronger
		// evidence silently shadowed by a weaker earlier marker. Every
		// applicable evaluation must remain independently visible to
		// admission under its exact context; visibility is NOT strength-
		// ranked replacement, and earlier decisions are preserved. The
		// companion no-inflation constraint is already structural: admission
		// materializes under the stable logical identity
		// "frontier-proposal:<id>", so a re-admitted proposal produces a new
		// REVISION of the same approach and the current-heads population
		// keeps exactly one interpretation current. Introspective +
		// idempotent; fresh DBs already carry the per-evaluation shape.
		apply: migrateV45EvaluatedFailuresPerEvaluation,
	},
	{
		version: 46,
		// v46 (attribution slice, 2026-09-12 review run remediation handoff 3):
		// the ATTEMPT→OUTPUT BINDING for witness-backed evaluations. A supplied
		// tuple proves only "this tuple fails the identity"; it does not
		// establish that the proposal's executed bounded attempt PRODUCED the
		// tuple. This table records, in the same transaction as the evaluation,
		// which registered deterministic procedure over which declared params
		// emitted the checked tuple — so admission can RECOMPUTE the procedure
		// and verify the binding instead of trusting a declared fixture
		// relationship. One binding per evaluation (PK); immutable; absence is
		// the recorded attribution gap for operator-supplied tuples, never
		// faked.
		sql: witnessAttemptBindingSQL,
	},
	{
		version: 47,
		// v47 (2026-09-13 review remediation): assessment scope is enforced at
		// the consuming write boundary. Foreign keys prove cited rows exist, but
		// not that the applicability decision and manifest belong to the same
		// policy/obligation/subject the assessment claims. The trigger refuses a
		// cross-subject applicability citation and a cross-policy/obligation
		// manifest citation; the Go store performs the same check to return a
		// clear error before SQLite aborts the insert.
		sql: reviewAssessmentReferenceScopeSQL,
	},
}

// invariantTablesSQL is the additive DDL for the candidate-invariant layer
// (migration v11). It creates the revisioned, immutable tables and their
// immutability triggers. It creates NONE of the M4.3 challenge-lifecycle tables
// (invariant_challenge, invariant_state_transition, the transition-counter
// trigger, the current-state view, invariant_lineage): candidates ship only in
// the `proposed` state, enforced by CHECK.
const invariantTablesSQL = `
CREATE TABLE IF NOT EXISTS invariant_revisions (
  id TEXT PRIMARY KEY,
  problem_id TEXT NOT NULL REFERENCES problems(id),
  failure_space_id TEXT NOT NULL REFERENCES failure_spaces(id),
  cluster_run_id TEXT NOT NULL REFERENCES cluster_runs(id),
  run_id TEXT NOT NULL REFERENCES runs(id),
  provider_invocation_id TEXT NOT NULL REFERENCES provider_invocations(id),
  miner_version TEXT NOT NULL,
  predicate_schema TEXT NOT NULL,
  min_support INTEGER NOT NULL,
  revision INTEGER NOT NULL,
  candidate_count INTEGER NOT NULL,
  created_at TEXT NOT NULL,
  UNIQUE(problem_id, failure_space_id, miner_version, predicate_schema, min_support),
  UNIQUE(problem_id, revision)
);

CREATE TABLE IF NOT EXISTS candidate_invariants (
  id TEXT PRIMARY KEY,
  invariant_revision_id TEXT NOT NULL REFERENCES invariant_revisions(id),
  predicate_fingerprint TEXT NOT NULL,
  statement TEXT NOT NULL,
  abstraction_level TEXT NOT NULL,
  initial_state TEXT NOT NULL DEFAULT 'proposed' CHECK (initial_state = 'proposed'),
  association_status TEXT NOT NULL CHECK (association_status IN ('recurring','contrast_observed','candidate_obstruction','unknown')),
  obstruction_is_model_hypothesis INTEGER NOT NULL DEFAULT 0,
  distinct_family_support INTEGER NOT NULL,
  failure_coverage_num INTEGER NOT NULL,
  failure_coverage_den INTEGER NOT NULL,
  contrast_violating_num INTEGER NOT NULL DEFAULT 0,
  contrast_eligible_den INTEGER NOT NULL DEFAULT 0,
  support_explicit_count INTEGER NOT NULL,
  support_inferred_count INTEGER NOT NULL,
  support_other_count INTEGER NOT NULL,
  confidence_ordinal TEXT NOT NULL DEFAULT '',
  ordinal INTEGER NOT NULL,
  UNIQUE(invariant_revision_id, predicate_fingerprint)
);

CREATE TABLE IF NOT EXISTS invariant_predicates (
  invariant_id TEXT PRIMARY KEY REFERENCES candidate_invariants(id),
  predicate_json TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS invariant_family_evaluations (
  invariant_id TEXT NOT NULL REFERENCES candidate_invariants(id),
  cluster_id TEXT NOT NULL REFERENCES mechanism_clusters(id),
  outcome_class TEXT NOT NULL,
  role TEXT NOT NULL CHECK (role IN ('support','contrast')),
  verdict TEXT NOT NULL CHECK (verdict IN ('satisfies','violates','unknown','member_mixed')),
  explicit_count INTEGER NOT NULL,
  inferred_count INTEGER NOT NULL,
  other_count INTEGER NOT NULL,
  PRIMARY KEY(invariant_id, cluster_id, role)
);

CREATE TABLE IF NOT EXISTS invariant_counterexamples (
  invariant_id TEXT NOT NULL REFERENCES candidate_invariants(id),
  cluster_id TEXT NOT NULL REFERENCES mechanism_clusters(id),
  reason TEXT NOT NULL,
  PRIMARY KEY(invariant_id, cluster_id)
);

CREATE INDEX IF NOT EXISTS idx_invariant_revisions_problem ON invariant_revisions(problem_id, revision DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_candidate_invariants_revision ON candidate_invariants(invariant_revision_id, ordinal);

CREATE TRIGGER IF NOT EXISTS invariant_revisions_immutable_update
BEFORE UPDATE ON invariant_revisions
BEGIN
  SELECT RAISE(ABORT, 'invariant revisions are immutable');
END;
CREATE TRIGGER IF NOT EXISTS invariant_revisions_immutable_delete
BEFORE DELETE ON invariant_revisions
BEGIN
  SELECT RAISE(ABORT, 'invariant revisions are immutable');
END;
CREATE TRIGGER IF NOT EXISTS candidate_invariants_immutable_update
BEFORE UPDATE ON candidate_invariants
BEGIN
  SELECT RAISE(ABORT, 'candidate invariants are immutable');
END;
CREATE TRIGGER IF NOT EXISTS candidate_invariants_immutable_delete
BEFORE DELETE ON candidate_invariants
BEGIN
  SELECT RAISE(ABORT, 'candidate invariants are immutable');
END;
CREATE TRIGGER IF NOT EXISTS invariant_predicates_immutable_update
BEFORE UPDATE ON invariant_predicates
BEGIN
  SELECT RAISE(ABORT, 'invariant predicates are immutable');
END;
CREATE TRIGGER IF NOT EXISTS invariant_predicates_immutable_delete
BEFORE DELETE ON invariant_predicates
BEGIN
  SELECT RAISE(ABORT, 'invariant predicates are immutable');
END;
CREATE TRIGGER IF NOT EXISTS invariant_family_evaluations_immutable_update
BEFORE UPDATE ON invariant_family_evaluations
BEGIN
  SELECT RAISE(ABORT, 'invariant family evaluations are immutable');
END;
CREATE TRIGGER IF NOT EXISTS invariant_family_evaluations_immutable_delete
BEFORE DELETE ON invariant_family_evaluations
BEGIN
  SELECT RAISE(ABORT, 'invariant family evaluations are immutable');
END;
CREATE TRIGGER IF NOT EXISTS invariant_counterexamples_immutable_update
BEFORE UPDATE ON invariant_counterexamples
BEGIN
  SELECT RAISE(ABORT, 'invariant counterexamples are immutable');
END;
CREATE TRIGGER IF NOT EXISTS invariant_counterexamples_immutable_delete
BEFORE DELETE ON invariant_counterexamples
BEGIN
  SELECT RAISE(ABORT, 'invariant counterexamples are immutable');
END;
`

// migrateV11CandidateInvariants creates the candidate-invariant tables and, when
// necessary, generalizes provider_invocations.role. Both parts are guarded so a
// fresh v11 database is untouched by the rebuild.
func migrateV11CandidateInvariants(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, invariantTablesSQL); err != nil {
		return err
	}
	allowsInvariant, err := providerRoleAllowsInvariant(ctx, tx)
	if err != nil {
		return err
	}
	if !allowsInvariant {
		if err := generalizeProviderInvocationsRole(ctx, tx); err != nil {
			return err
		}
	}
	return nil
}

// providerRoleAllowsInvariant reports whether the provider_invocations.role
// CHECK already admits 'invariant'. It reads the table's DDL text from
// sqlite_master and looks for 'invariant' in the role CHECK, which is
// deterministic and side-effect-free (no synthetic FK seeding needed). A fresh
// v11 database created with the generalized shape reports true and the rebuild
// is skipped; a pre-v11 database still on ('normalize') reports false.
func providerRoleAllowsInvariant(ctx context.Context, tx *sql.Tx) (bool, error) {
	var ddl string
	row := tx.QueryRowContext(ctx, `SELECT sql FROM sqlite_master WHERE type='table' AND name='provider_invocations'`)
	if err := row.Scan(&ddl); err != nil {
		return false, err
	}
	return strings.Contains(ddl, "'invariant'"), nil
}

// generalizeProviderInvocationsRole widens the provider_invocations.role CHECK
// from ('normalize') to ('normalize','invariant').
//
// KTD-8a proposed a v10-style DROP+RENAME parent rebuild, but that is NOT
// FK-safe under this SQLite build when child rows already reference
// provider_invocations (normalization_revisions.provider_invocation_id): the
// deferred FK still fails at commit after the parent is dropped and recreated
// (verified empirically). Instead we edit the table's CHECK in place via
// PRAGMA writable_schema, which never drops the parent, so every child FK is
// preserved by construction. This is a targeted, single-table DDL-text edit
// (widening one CHECK enum), guarded by providerRoleAllowsInvariant so it is a
// no-op on a fresh v11 DB, and followed by an integrity_check to reject a
// schema whose btree/index structure was corrupted by the write. The
// immutability triggers are untouched.
func generalizeProviderInvocationsRole(ctx context.Context, tx *sql.Tx) error {
	var ddl string
	if err := tx.QueryRowContext(ctx, `SELECT sql FROM sqlite_master WHERE type='table' AND name='provider_invocations'`).Scan(&ddl); err != nil {
		return err
	}
	updated := strings.Replace(ddl, "role IN ('normalize')", "role IN ('normalize','invariant')", 1)
	if updated == ddl {
		return fmt.Errorf("v11: could not locate provider_invocations.role normalize-only CHECK to generalize")
	}
	var schemaVersion int
	if err := tx.QueryRowContext(ctx, `PRAGMA schema_version`).Scan(&schemaVersion); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `PRAGMA writable_schema = ON`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE sqlite_master SET sql = ? WHERE type='table' AND name='provider_invocations'`, updated); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `PRAGMA writable_schema = OFF`); err != nil {
		return err
	}
	// Force SQLite to reparse the edited schema so the widened CHECK takes effect
	// immediately (both later in this transaction and after commit).
	if _, err := tx.ExecContext(ctx, fmt.Sprintf(`PRAGMA schema_version = %d`, schemaVersion+1)); err != nil {
		return err
	}
	// integrity_check validates btree/index/data consistency, not DDL
	// parseability, so it confirms the writable_schema write did not corrupt
	// storage structures. The edited CHECK's parseability is proven separately:
	// the new sqlite_master text is re-parsed on the next connection, and
	// TestV11ProviderRoleGeneralizationFromPreV11 asserts an 'invariant' role
	// row inserts (accepted) while an out-of-enum role is still rejected.
	var res string
	if err := tx.QueryRowContext(ctx, `PRAGMA integrity_check`).Scan(&res); err != nil {
		return err
	}
	if res != "ok" {
		return fmt.Errorf("v11: integrity_check after role generalization: %s", res)
	}
	return nil
}

// editTableCheckInPlace performs an FK-safe, in-place edit of a single table's
// DDL text (typically widening or renaming a CHECK enum) via PRAGMA
// writable_schema, then forces a schema reparse and runs integrity_check. It
// never drops the parent table, so child FKs are preserved by construction —
// the pattern v11 established after the v10-style DROP+RENAME rebuild proved not
// FK-safe with existing child rows. old must occur exactly once in the current
// DDL; a no-op (old absent) is reported as an error so callers can guard with an
// introspective idempotency check first.
func editTableCheckInPlace(ctx context.Context, tx *sql.Tx, table, old, new string) error {
	var ddl string
	if err := tx.QueryRowContext(ctx, `SELECT sql FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&ddl); err != nil {
		return err
	}
	updated := strings.Replace(ddl, old, new, 1)
	if updated == ddl {
		return fmt.Errorf("in-place CHECK edit: could not locate %q in %s DDL", old, table)
	}
	var schemaVersion int
	if err := tx.QueryRowContext(ctx, `PRAGMA schema_version`).Scan(&schemaVersion); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `PRAGMA writable_schema = ON`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE sqlite_master SET sql = ? WHERE type='table' AND name=?`, updated, table); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `PRAGMA writable_schema = OFF`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, fmt.Sprintf(`PRAGMA schema_version = %d`, schemaVersion+1)); err != nil {
		return err
	}
	var res string
	if err := tx.QueryRowContext(ctx, `PRAGMA integrity_check`).Scan(&res); err != nil {
		return err
	}
	if res != "ok" {
		return fmt.Errorf("integrity_check after %s CHECK edit: %s", table, res)
	}
	return nil
}

// migrateV12ContrastCountsAndRename applies the M4.2 semantic-hardening schema
// changes (F2): additive contrast count columns and the association_status enum
// rename discriminative -> contrast_observed. Introspective + idempotent.
func migrateV12ContrastCountsAndRename(ctx context.Context, tx *sql.Tx) error {
	// (1) Additive contrast count columns (FK-safe ADD COLUMN, guarded).
	for _, col := range []struct{ name, ddl string }{
		{"contrast_violating_num", "ALTER TABLE candidate_invariants ADD COLUMN contrast_violating_num INTEGER NOT NULL DEFAULT 0"},
		{"contrast_eligible_den", "ALTER TABLE candidate_invariants ADD COLUMN contrast_eligible_den INTEGER NOT NULL DEFAULT 0"},
	} {
		has, err := columnExists(ctx, tx, "candidate_invariants", col.name)
		if err != nil {
			return err
		}
		if !has {
			if _, err := tx.ExecContext(ctx, col.ddl); err != nil {
				return err
			}
		}
	}
	// (2) Rename the association_status enum value in the CHECK, in place. Guard
	// on whether the legacy value is still present so a fresh v12 DB is a no-op.
	var ddl string
	if err := tx.QueryRowContext(ctx, `SELECT sql FROM sqlite_master WHERE type='table' AND name='candidate_invariants'`).Scan(&ddl); err != nil {
		return err
	}
	if strings.Contains(ddl, "'discriminative'") {
		if err := editTableCheckInPlace(ctx, tx, "candidate_invariants",
			"'recurring','discriminative','candidate_obstruction','unknown'",
			"'recurring','contrast_observed','candidate_obstruction','unknown'"); err != nil {
			return err
		}
	}
	return nil
}

// migrateV10ClusterIdentityAndOutcome adds input_set_hash to cluster_runs (and
// the corresponding idempotency UNIQUE) plus per-family outcome columns on
// mechanism_clusters. Both changes are guarded so a fresh v10 schema is
// untouched.
func migrateV10ClusterIdentityAndOutcome(ctx context.Context, tx *sql.Tx) error {
	// 1. mechanism_clusters per-family outcome columns (plain ADD COLUMN; no PK
	//    or UNIQUE change, so no rebuild needed).
	clusterCols := []struct {
		table, column, ddl string
	}{
		{"mechanism_clusters", "outcome_class", "ALTER TABLE mechanism_clusters ADD COLUMN outcome_class TEXT NOT NULL DEFAULT 'unknown'"},
		{"mechanism_clusters", "outcome_mixed", "ALTER TABLE mechanism_clusters ADD COLUMN outcome_mixed INTEGER NOT NULL DEFAULT 0"},
	}
	for _, c := range clusterCols {
		has, err := columnExists(ctx, tx, c.table, c.column)
		if err != nil {
			return err
		}
		if has {
			continue
		}
		if _, err := tx.ExecContext(ctx, c.ddl); err != nil {
			return err
		}
	}

	// 2. cluster_runs.input_set_hash + UNIQUE. The idempotency UNIQUE must gain
	//    input_set_hash. The original v10 code rebuilt the table
	//    (create-copy-DROP-rename) to widen the UNIQUE, but dropping cluster_runs
	//    is NOT FK-safe: mechanism_clusters, cluster_distances,
	//    cluster_coverage_axes, cluster_discrimination_losses, failure_spaces and
	//    invariant_revisions all reference it, and with foreign_keys=ON a
	//    populated pre-v10 database fails at commit after the parent is dropped
	//    (the same hazard v11 documents for provider_invocations). Instead we ADD
	//    COLUMN (FK-safe, never drops the parent) and enforce the widened key with
	//    a UNIQUE INDEX — a UNIQUE index is equivalent to a table-level UNIQUE
	//    constraint for conflict detection (F4).
	has, err := columnExists(ctx, tx, "cluster_runs", "input_set_hash")
	if err != nil {
		return err
	}
	if !has {
		if err := addClusterRunsInputSetHashFKSafe(ctx, tx); err != nil {
			return err
		}
	}
	return nil
}

// addClusterRunsInputSetHashFKSafe adds cluster_runs.input_set_hash and widens
// the idempotency UNIQUE without dropping the parent table, so every child FK is
// preserved by construction (F4). It handles both fresh v7+ databases (identity
// enforced by a named UNIQUE INDEX ux_cluster_runs_identity) and legacy
// databases created under the earlier v7 that used an inline table-level UNIQUE
// (backed by an sqlite_autoindex). In both cases the 6-column identity is
// replaced by a 7-column UNIQUE INDEX that includes input_set_hash. The
// immutability triggers are untouched because the table is never recreated.
func addClusterRunsInputSetHashFKSafe(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, `ALTER TABLE cluster_runs ADD COLUMN input_set_hash TEXT NOT NULL DEFAULT ''`); err != nil {
		return err
	}
	// Fresh v7+ path: identity was enforced by a droppable named index. Drop the
	// 6-column index and replace it with the 7-column one. No inline UNIQUE exists
	// in the DDL, so no rebuild is required.
	if _, err := tx.ExecContext(ctx, `DROP INDEX IF EXISTS ux_cluster_runs_identity`); err != nil {
		return err
	}
	// Legacy path: databases created under the earliest v7 carry an inline
	// table-level UNIQUE(6 columns) backed by an sqlite_autoindex. That constraint
	// is stricter than the widened key and would reject a re-cluster that differs
	// only by input_set_hash, so it must be removed. An autoindex cannot be
	// dropped directly, and editing it out of the schema in place corrupts the
	// freelist. The only robust removal is a table rebuild — which is FK-safe here
	// because Migrate runs with foreign_keys=OFF and performs a foreign_key_check
	// before commit. rebuildLegacyClusterRuns is a no-op on a fresh database whose
	// DDL has no inline UNIQUE.
	if err := rebuildLegacyClusterRuns(ctx, tx); err != nil {
		return err
	}
	// Enforce the widened 7-column identity with a UNIQUE index.
	if _, err := tx.ExecContext(ctx, `CREATE UNIQUE INDEX IF NOT EXISTS ux_cluster_runs_identity
ON cluster_runs(problem_id, schema_version, vocabulary_version, profile_version, cluster_algo_version, thresholds_hash, input_set_hash)`); err != nil {
		return err
	}
	return nil
}

// rebuildLegacyClusterRuns rebuilds cluster_runs to drop the legacy inline
// table-level UNIQUE (present only in databases created under the earliest v7
// DDL) by creating a constraint-free replacement, copying every row, dropping
// the old table, and renaming. It is FK-safe only under foreign_keys=OFF (the
// mode Migrate establishes); children keep their textual parent ids across the
// swap and Migrate's pre-commit foreign_key_check proves none dangled. The
// cluster_runs immutability triggers are dropped by DROP TABLE and recreated
// against the new table. On a fresh database (no inline UNIQUE) it is a no-op.
func rebuildLegacyClusterRuns(ctx context.Context, tx *sql.Tx) error {
	var ddl string
	if err := tx.QueryRowContext(ctx, `SELECT sql FROM sqlite_master WHERE type='table' AND name='cluster_runs'`).Scan(&ddl); err != nil {
		return err
	}
	if !strings.Contains(ddl, "UNIQUE(problem_id") {
		return nil // fresh DB: no inline UNIQUE to remove.
	}
	// Determine the current column list so the copy is exact regardless of which
	// pre-v10 additive columns are present. input_set_hash has already been added
	// by the caller, so it is included and copied through.
	cols, err := tableColumnNames(ctx, tx, "cluster_runs")
	if err != nil {
		return err
	}
	colList := strings.Join(cols, ", ")
	// Constraint-free replacement (no inline UNIQUE; the widened UNIQUE INDEX is
	// created by the caller afterward). Columns mirror the seed v7 DDL plus any
	// additively-migrated columns copied verbatim.
	if _, err := tx.ExecContext(ctx, `CREATE TABLE cluster_runs__rebuild (
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
  input_set_hash TEXT NOT NULL DEFAULT ''
)`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, fmt.Sprintf(`INSERT INTO cluster_runs__rebuild (%s) SELECT %s FROM cluster_runs`, colList, colList)); err != nil {
		return err
	}
	// Drop the immutability triggers before dropping the table so the rename does
	// not collide, then drop and rename.
	if _, err := tx.ExecContext(ctx, `DROP TRIGGER IF EXISTS cluster_runs_immutable_update`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DROP TRIGGER IF EXISTS cluster_runs_immutable_delete`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DROP TABLE cluster_runs`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `ALTER TABLE cluster_runs__rebuild RENAME TO cluster_runs`); err != nil {
		return err
	}
	// Recreate the immutability triggers and the problem index on the new table.
	if _, err := tx.ExecContext(ctx, `CREATE TRIGGER IF NOT EXISTS cluster_runs_immutable_update
BEFORE UPDATE ON cluster_runs
BEGIN
  SELECT RAISE(ABORT, 'cluster runs are immutable');
END`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `CREATE TRIGGER IF NOT EXISTS cluster_runs_immutable_delete
BEFORE DELETE ON cluster_runs
BEGIN
  SELECT RAISE(ABORT, 'cluster runs are immutable');
END`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_cluster_runs_problem ON cluster_runs(problem_id)`); err != nil {
		return err
	}
	return nil
}

// tableColumnNames returns the column names of table in schema (declaration)
// order via PRAGMA table_info.
func tableColumnNames(ctx context.Context, tx *sql.Tx, table string) ([]string, error) {
	rows, err := tx.QueryContext(ctx, fmt.Sprintf(`PRAGMA table_info(%s)`, table))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cols []string
	for rows.Next() {
		var (
			cid        int
			name, ctyp string
			notnull    int
			dflt       any
			pk         int
		)
		if err := rows.Scan(&cid, &name, &ctyp, &notnull, &dflt, &pk); err != nil {
			return nil, err
		}
		cols = append(cols, name)
	}
	return cols, rows.Err()
}

// migrateV9RepairAndRejected upgrades a pre-fix v6-era schema and adds durable
// rejected-term storage, doing nothing on an already-correct (fresh) database.
func migrateV9RepairAndRejected(ctx context.Context, tx *sql.Tx) error {
	// 1. Provenance columns added when migrations 4/5 were corrected in place.
	//    Add only if absent so a fresh v5 table (which already has them) is
	//    untouched.
	provenanceCols := []struct {
		table, column, ddl string
	}{
		{"signature_postures", "claim_status", "ALTER TABLE signature_postures ADD COLUMN claim_status TEXT NOT NULL DEFAULT 'unknown'"},
		{"signature_boundaries", "claim_status", "ALTER TABLE signature_boundaries ADD COLUMN claim_status TEXT NOT NULL DEFAULT 'unknown'"},
		{"signature_boundaries", "support_snapshot_id", "ALTER TABLE signature_boundaries ADD COLUMN support_snapshot_id TEXT NOT NULL DEFAULT ''"},
		{"signature_boundaries", "support_locator", "ALTER TABLE signature_boundaries ADD COLUMN support_locator TEXT NOT NULL DEFAULT ''"},
		{"signature_outcomes", "claim_status", "ALTER TABLE signature_outcomes ADD COLUMN claim_status TEXT NOT NULL DEFAULT 'unknown'"},
	}
	for _, c := range provenanceCols {
		has, err := columnExists(ctx, tx, c.table, c.column)
		if err != nil {
			return err
		}
		if has {
			continue
		}
		if _, err := tx.ExecContext(ctx, c.ddl); err != nil {
			return err
		}
	}

	// 2. Alias uniqueness key. The corrected PK includes canonical_id so a phrase
	//    may legitimately bind to more than one canonical id within a field kind.
	//    Rebuild only when canonical_id is absent from the primary key (i.e. the
	//    pre-fix 3-column PK). SQLite cannot ALTER a PK, so rebuild the table.
	aliasHasCanonicalPK, err := columnInPrimaryKey(ctx, tx, "canonical_term_aliases", "canonical_id")
	if err != nil {
		return err
	}
	if !aliasHasCanonicalPK {
		if err := rebuildAliasTableWithCanonicalPK(ctx, tx); err != nil {
			return err
		}
	}

	// 3. Durable rejected terms (previously in-memory only, silently lost on
	//    reload). CREATE IF NOT EXISTS + immutable triggers; idempotent.
	if _, err := tx.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS canonical_rejected_terms (
  vocabulary_version TEXT NOT NULL REFERENCES canonical_vocabulary(version),
  rejected_normalized TEXT NOT NULL,
  PRIMARY KEY(vocabulary_version, rejected_normalized)
);
CREATE TRIGGER IF NOT EXISTS canonical_rejected_terms_immutable_update
BEFORE UPDATE ON canonical_rejected_terms
BEGIN
  SELECT RAISE(ABORT, 'canonical rejected terms are immutable');
END;
CREATE TRIGGER IF NOT EXISTS canonical_rejected_terms_immutable_delete
BEFORE DELETE ON canonical_rejected_terms
BEGIN
  SELECT RAISE(ABORT, 'canonical rejected terms are immutable');
END;
`); err != nil {
		return err
	}
	return nil
}

// columnExists reports whether table has a column of the given name.
func columnExists(ctx context.Context, tx *sql.Tx, table, column string) (bool, error) {
	rows, err := tx.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return false, err
		}
		if name == column {
			return true, nil
		}
	}
	return false, rows.Err()
}

// columnInPrimaryKey reports whether the named column participates in table's
// primary key (pk > 0 in PRAGMA table_info).
func columnInPrimaryKey(ctx context.Context, tx *sql.Tx, table, column string) (bool, error) {
	rows, err := tx.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return false, err
		}
		if name == column && pk > 0 {
			return true, nil
		}
	}
	return false, rows.Err()
}

// rebuildAliasTableWithCanonicalPK migrates canonical_term_aliases from the
// pre-fix 3-column PK to the corrected 4-column PK that includes canonical_id.
// It drops the immutability triggers, rebuilds via a temp table preserving all
// rows, and reinstates the triggers. Because SQLite cannot ALTER a primary key
// this is the standard create-copy-drop-rename rebuild.
func rebuildAliasTableWithCanonicalPK(ctx context.Context, tx *sql.Tx) error {
	stmts := []string{
		`DROP TRIGGER IF EXISTS canonical_term_aliases_immutable_update`,
		`DROP TRIGGER IF EXISTS canonical_term_aliases_immutable_delete`,
		`CREATE TABLE canonical_term_aliases__v9 (
  vocabulary_version TEXT NOT NULL REFERENCES canonical_vocabulary(version),
  canonical_id TEXT NOT NULL,
  field_kind TEXT NOT NULL CHECK (field_kind IN ('representation', 'assumption', 'operator', 'preserves', 'breaks', 'auxiliary_object', 'outcome', 'boundary', 'posture')),
  alias_normalized TEXT NOT NULL,
  PRIMARY KEY(vocabulary_version, field_kind, alias_normalized, canonical_id)
)`,
		`INSERT OR IGNORE INTO canonical_term_aliases__v9(vocabulary_version, canonical_id, field_kind, alias_normalized)
  SELECT vocabulary_version, canonical_id, field_kind, alias_normalized FROM canonical_term_aliases`,
		`DROP TABLE canonical_term_aliases`,
		`ALTER TABLE canonical_term_aliases__v9 RENAME TO canonical_term_aliases`,
		`CREATE INDEX IF NOT EXISTS idx_canonical_term_aliases_id ON canonical_term_aliases(vocabulary_version, canonical_id)`,
		`CREATE TRIGGER IF NOT EXISTS canonical_term_aliases_immutable_update
BEFORE UPDATE ON canonical_term_aliases
BEGIN
  SELECT RAISE(ABORT, 'canonical term aliases are immutable');
END`,
		`CREATE TRIGGER IF NOT EXISTS canonical_term_aliases_immutable_delete
BEFORE DELETE ON canonical_term_aliases
BEGIN
  SELECT RAISE(ABORT, 'canonical term aliases are immutable');
END`,
	}
	for _, s := range stmts {
		if _, err := tx.ExecContext(ctx, s); err != nil {
			return err
		}
	}
	return nil
}

// challengeTablesSQL is the additive DDL for the M4.3 challenge lifecycle
// (migration v13): immutable challenge rows, the mutable per-invariant
// transition counter (the ONLY mutation surface), append-only state transitions
// guarded by the blueprint's validating trigger, the invariant_current_state
// read view, evidence links to real persisted rows, synthetic artifacts, and
// split/merge/weaken lineage. State lives only in transitions (KTD-1).
const challengeTablesSQL = `
CREATE TABLE IF NOT EXISTS invariant_challenges (
  id TEXT PRIMARY KEY,
  invariant_id TEXT NOT NULL REFERENCES candidate_invariants(id),
  run_id TEXT NOT NULL REFERENCES runs(id),
  provider_invocation_id TEXT NOT NULL REFERENCES provider_invocations(id),
  challenge_type TEXT NOT NULL CHECK (challenge_type IN (
    'known-counterexample', 'synthetic-counterexample', 'success-preserving',
    'split', 'merge', 'bias-critique', 'independent-verification')),
  claimed_verdict TEXT NOT NULL DEFAULT '',
  result_summary TEXT NOT NULL CHECK (result_summary IN ('confirmed', 'unconfirmed')),
  detail TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS invariant_transition_counters (
  invariant_id TEXT PRIMARY KEY REFERENCES candidate_invariants(id),
  last_transition_seq INTEGER NOT NULL DEFAULT 0 CHECK (last_transition_seq >= 0)
);

CREATE TABLE IF NOT EXISTS invariant_state_transitions (
  invariant_id TEXT NOT NULL REFERENCES candidate_invariants(id),
  transition_seq INTEGER NOT NULL,
  challenge_id TEXT NOT NULL REFERENCES invariant_challenges(id),
  from_state TEXT NOT NULL CHECK (from_state IN ('proposed', 'challenged', 'surviving', 'weaken')),
  to_state TEXT NOT NULL CHECK (to_state IN ('challenged', 'surviving', 'weaken', 'falsified', 'established')),
  created_at TEXT NOT NULL,
  PRIMARY KEY(invariant_id, transition_seq)
);

CREATE TRIGGER IF NOT EXISTS invariant_state_transitions_validate_insert
BEFORE INSERT ON invariant_state_transitions
BEGIN
  SELECT CASE
    WHEN COALESCE((
      SELECT itc.last_transition_seq
      FROM invariant_transition_counters itc
      WHERE itc.invariant_id = NEW.invariant_id
    ), 0) = 0 THEN RAISE(ABORT, 'transition_seq requires a prior counter allocation')
    WHEN NEW.transition_seq <> (
      SELECT itc.last_transition_seq
      FROM invariant_transition_counters itc
      WHERE itc.invariant_id = NEW.invariant_id
    ) THEN RAISE(ABORT, 'transition_seq must match the atomically allocated invariant counter')
    WHEN NEW.from_state <> COALESCE((
      SELECT t.to_state
      FROM invariant_state_transitions t
      WHERE t.invariant_id = NEW.invariant_id
      ORDER BY t.transition_seq DESC
      LIMIT 1
    ), (
      SELECT ci.initial_state
      FROM candidate_invariants ci
      WHERE ci.id = NEW.invariant_id
    )) THEN RAISE(ABORT, 'from_state must match current invariant state')
    WHEN NOT (
      (NEW.from_state = 'proposed' AND NEW.to_state = 'challenged') OR
      (NEW.from_state = 'challenged' AND NEW.to_state IN ('surviving', 'weaken', 'falsified')) OR
      (NEW.from_state = 'surviving' AND NEW.to_state IN ('challenged', 'weaken', 'falsified', 'established')) OR
      (NEW.from_state = 'weaken' AND NEW.to_state IN ('challenged', 'surviving', 'falsified'))
    ) THEN RAISE(ABORT, 'invalid invariant state transition')
  END;
END;

CREATE VIEW IF NOT EXISTS invariant_current_state AS
WITH ranked AS (
  SELECT
    t.invariant_id,
    t.to_state,
    t.created_at,
    ROW_NUMBER() OVER (
      PARTITION BY t.invariant_id
      ORDER BY t.transition_seq DESC
    ) AS rn
  FROM invariant_state_transitions t
),
latest AS (
  SELECT invariant_id, to_state, created_at
  FROM ranked
  WHERE rn = 1
)
SELECT ci.id AS invariant_id,
       COALESCE(latest.to_state, ci.initial_state) AS state,
       COALESCE(latest.created_at, ir.created_at) AS as_of
FROM candidate_invariants ci
JOIN invariant_revisions ir ON ir.id = ci.invariant_revision_id
LEFT JOIN latest ON latest.invariant_id = ci.id;

CREATE TABLE IF NOT EXISTS invariant_challenge_evidence (
  challenge_id TEXT NOT NULL REFERENCES invariant_challenges(id),
  kind TEXT NOT NULL CHECK (kind IN (
    'counterexample_member', 'success_family', 'support_recount',
    'grounding', 'independent_source')),
  cluster_id TEXT REFERENCES mechanism_clusters(id),
  signature_id TEXT REFERENCES mechanism_signatures(id),
  snapshot_id TEXT REFERENCES source_snapshots(id),
  detail TEXT NOT NULL DEFAULT '',
  ordinal INTEGER NOT NULL,
  PRIMARY KEY(challenge_id, ordinal)
);

CREATE TABLE IF NOT EXISTS synthetic_artifacts (
  id TEXT PRIMARY KEY,
  problem_id TEXT NOT NULL REFERENCES problems(id),
  run_id TEXT NOT NULL REFERENCES runs(id),
  artifact_type TEXT NOT NULL CHECK (artifact_type IN ('synthetic_attempt', 'synthetic_counterexample')),
  content TEXT NOT NULL,
  created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS invariant_challenge_synthetic_artifacts (
  challenge_id TEXT NOT NULL REFERENCES invariant_challenges(id),
  synthetic_artifact_id TEXT NOT NULL REFERENCES synthetic_artifacts(id),
  PRIMARY KEY(challenge_id, synthetic_artifact_id)
);

CREATE TABLE IF NOT EXISTS invariant_lineage (
  parent_invariant_id TEXT NOT NULL REFERENCES candidate_invariants(id),
  child_invariant_id TEXT NOT NULL REFERENCES candidate_invariants(id),
  relation TEXT NOT NULL CHECK (relation IN ('split', 'merge', 'weaken')),
  PRIMARY KEY(parent_invariant_id, child_invariant_id, relation)
);

CREATE INDEX IF NOT EXISTS idx_invariant_challenges_invariant ON invariant_challenges(invariant_id);
CREATE INDEX IF NOT EXISTS idx_invariant_state_transitions_invariant ON invariant_state_transitions(invariant_id);

CREATE TRIGGER IF NOT EXISTS invariant_challenges_immutable_update
BEFORE UPDATE ON invariant_challenges
BEGIN
  SELECT RAISE(ABORT, 'invariant challenges are immutable');
END;
CREATE TRIGGER IF NOT EXISTS invariant_challenges_immutable_delete
BEFORE DELETE ON invariant_challenges
BEGIN
  SELECT RAISE(ABORT, 'invariant challenges are immutable');
END;
CREATE TRIGGER IF NOT EXISTS invariant_state_transitions_immutable_update
BEFORE UPDATE ON invariant_state_transitions
BEGIN
  SELECT RAISE(ABORT, 'invariant state transitions are immutable');
END;
CREATE TRIGGER IF NOT EXISTS invariant_state_transitions_immutable_delete
BEFORE DELETE ON invariant_state_transitions
BEGIN
  SELECT RAISE(ABORT, 'invariant state transitions are immutable');
END;
CREATE TRIGGER IF NOT EXISTS invariant_challenge_evidence_immutable_update
BEFORE UPDATE ON invariant_challenge_evidence
BEGIN
  SELECT RAISE(ABORT, 'invariant challenge evidence is immutable');
END;
CREATE TRIGGER IF NOT EXISTS invariant_challenge_evidence_immutable_delete
BEFORE DELETE ON invariant_challenge_evidence
BEGIN
  SELECT RAISE(ABORT, 'invariant challenge evidence is immutable');
END;
CREATE TRIGGER IF NOT EXISTS synthetic_artifacts_immutable_update
BEFORE UPDATE ON synthetic_artifacts
BEGIN
  SELECT RAISE(ABORT, 'synthetic artifacts are immutable');
END;
CREATE TRIGGER IF NOT EXISTS synthetic_artifacts_immutable_delete
BEFORE DELETE ON synthetic_artifacts
BEGIN
  SELECT RAISE(ABORT, 'synthetic artifacts are immutable');
END;
CREATE TRIGGER IF NOT EXISTS invariant_challenge_synthetic_artifacts_immutable_update
BEFORE UPDATE ON invariant_challenge_synthetic_artifacts
BEGIN
  SELECT RAISE(ABORT, 'invariant challenge synthetic links are immutable');
END;
CREATE TRIGGER IF NOT EXISTS invariant_challenge_synthetic_artifacts_immutable_delete
BEFORE DELETE ON invariant_challenge_synthetic_artifacts
BEGIN
  SELECT RAISE(ABORT, 'invariant challenge synthetic links are immutable');
END;
CREATE TRIGGER IF NOT EXISTS invariant_lineage_immutable_update
BEFORE UPDATE ON invariant_lineage
BEGIN
  SELECT RAISE(ABORT, 'invariant lineage is immutable');
END;
CREATE TRIGGER IF NOT EXISTS invariant_lineage_immutable_delete
BEFORE DELETE ON invariant_lineage
BEGIN
  SELECT RAISE(ABORT, 'invariant lineage is immutable');
END;
`

// migrateV13ChallengeLifecycle creates the M4.3 challenge tables and widens the
// provider_invocations.role CHECK to admit 'challenge'. Both steps are guarded
// and idempotent: the DDL is IF NOT EXISTS throughout, and the CHECK edit runs
// only when 'challenge' is not already permitted (a fresh v13 database is
// untouched).
func migrateV13ChallengeLifecycle(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, challengeTablesSQL); err != nil {
		return err
	}
	allows, err := providerRoleAllows(ctx, tx, "challenge")
	if err != nil {
		return err
	}
	if !allows {
		if err := editTableCheckInPlace(ctx, tx, "provider_invocations",
			"role IN ('normalize','invariant')",
			"role IN ('normalize','invariant','challenge')"); err != nil {
			return err
		}
	}
	return nil
}

// providerRoleAllows reports whether the provider_invocations.role CHECK
// already admits the given role, by reading the table DDL from sqlite_master
// (deterministic and side-effect-free).
func providerRoleAllows(ctx context.Context, tx *sql.Tx, role string) (bool, error) {
	var ddl string
	row := tx.QueryRowContext(ctx, `SELECT sql FROM sqlite_master WHERE type='table' AND name='provider_invocations'`)
	if err := row.Scan(&ddl); err != nil {
		return false, err
	}
	return strings.Contains(ddl, "'"+role+"'"), nil
}

// frontierTablesSQL is the additive DDL for the M5.1 frontier layer. Every table
// is immutable by trigger and append-only; a proposal's `result` is left NULL
// here (M5.2 evaluation populates it). The holdout_leakage_check link from the
// persistence blueprint is deferred to M7 and intentionally omitted.
const frontierTablesSQL = `
CREATE TABLE IF NOT EXISTS frontier_generation_runs (
  id TEXT PRIMARY KEY,
  problem_id TEXT NOT NULL REFERENCES problems(id),
  cluster_run_id TEXT NOT NULL REFERENCES cluster_runs(id),
  run_id TEXT NOT NULL REFERENCES runs(id),
  provider_invocation_id TEXT NOT NULL REFERENCES provider_invocations(id),
  generator_version TEXT NOT NULL,
  requested_count INTEGER NOT NULL,
  proposal_count INTEGER NOT NULL,
  revision INTEGER NOT NULL,
  created_at TEXT NOT NULL,
  UNIQUE(problem_id, revision)
);

CREATE TABLE IF NOT EXISTS frontier_proposals (
  id TEXT PRIMARY KEY,
  problem_id TEXT NOT NULL REFERENCES problems(id),
  frontier_generation_run_id TEXT NOT NULL REFERENCES frontier_generation_runs(id),
  proposal_hash TEXT NOT NULL,
  structural_violation_claim TEXT NOT NULL,
  novelty_argument TEXT NOT NULL,
  cheapest_falsification_path TEXT NOT NULL,
  mechanistic_distance_ordinal TEXT NOT NULL,
  expected_information_gain_ordinal TEXT NOT NULL,
  evaluation_cost_ordinal TEXT NOT NULL,
  violates_any_target INTEGER NOT NULL DEFAULT 0,
  rank_ordinal INTEGER NOT NULL,
  result TEXT,
  created_at TEXT NOT NULL,
  UNIQUE(problem_id, proposal_hash)
);

CREATE TABLE IF NOT EXISTS frontier_target_invariants (
  proposal_id TEXT NOT NULL REFERENCES frontier_proposals(id),
  invariant_id TEXT NOT NULL REFERENCES candidate_invariants(id),
  verdict TEXT NOT NULL CHECK (verdict IN ('satisfies','violates','unknown')),
  violated INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY(proposal_id, invariant_id)
);

CREATE TABLE IF NOT EXISTS frontier_nearest_clusters (
  proposal_id TEXT NOT NULL REFERENCES frontier_proposals(id),
  cluster_id TEXT NOT NULL REFERENCES mechanism_clusters(id),
  classification TEXT NOT NULL,
  proximity_ordinal TEXT NOT NULL,
  PRIMARY KEY(proposal_id, cluster_id)
);

CREATE INDEX IF NOT EXISTS idx_frontier_proposals_run
  ON frontier_proposals(frontier_generation_run_id);
CREATE INDEX IF NOT EXISTS idx_frontier_generation_runs_problem
  ON frontier_generation_runs(problem_id);

CREATE TRIGGER IF NOT EXISTS frontier_generation_runs_immutable_update
BEFORE UPDATE ON frontier_generation_runs
BEGIN
  SELECT RAISE(ABORT, 'frontier generation runs are immutable');
END;
CREATE TRIGGER IF NOT EXISTS frontier_generation_runs_immutable_delete
BEFORE DELETE ON frontier_generation_runs
BEGIN
  SELECT RAISE(ABORT, 'frontier generation runs are immutable');
END;

-- frontier_proposals is append-only, EXCEPT the M5.2 result column: an
-- evaluation may set result exactly once (NULL -> a verdict). Any other column
-- change, or overwriting a non-null result, aborts. Deletes always abort.
CREATE TRIGGER IF NOT EXISTS frontier_proposals_immutable_update
BEFORE UPDATE ON frontier_proposals
BEGIN
  SELECT CASE
    WHEN NEW.id <> OLD.id
      OR NEW.problem_id <> OLD.problem_id
      OR NEW.frontier_generation_run_id <> OLD.frontier_generation_run_id
      OR NEW.proposal_hash <> OLD.proposal_hash
      OR NEW.structural_violation_claim <> OLD.structural_violation_claim
      OR NEW.novelty_argument <> OLD.novelty_argument
      OR NEW.cheapest_falsification_path <> OLD.cheapest_falsification_path
      OR NEW.mechanistic_distance_ordinal <> OLD.mechanistic_distance_ordinal
      OR NEW.expected_information_gain_ordinal <> OLD.expected_information_gain_ordinal
      OR NEW.evaluation_cost_ordinal <> OLD.evaluation_cost_ordinal
      OR NEW.violates_any_target <> OLD.violates_any_target
      OR NEW.rank_ordinal <> OLD.rank_ordinal
      OR NEW.created_at <> OLD.created_at
      OR OLD.result IS NOT NULL
    THEN RAISE(ABORT, 'frontier proposals are immutable except a one-time result set')
  END;
END;
CREATE TRIGGER IF NOT EXISTS frontier_proposals_immutable_delete
BEFORE DELETE ON frontier_proposals
BEGIN
  SELECT RAISE(ABORT, 'frontier proposals are immutable');
END;

CREATE TRIGGER IF NOT EXISTS frontier_target_invariants_immutable_update
BEFORE UPDATE ON frontier_target_invariants
BEGIN
  SELECT RAISE(ABORT, 'frontier target invariants are immutable');
END;
CREATE TRIGGER IF NOT EXISTS frontier_target_invariants_immutable_delete
BEFORE DELETE ON frontier_target_invariants
BEGIN
  SELECT RAISE(ABORT, 'frontier target invariants are immutable');
END;

CREATE TRIGGER IF NOT EXISTS frontier_nearest_clusters_immutable_update
BEFORE UPDATE ON frontier_nearest_clusters
BEGIN
  SELECT RAISE(ABORT, 'frontier nearest clusters are immutable');
END;
CREATE TRIGGER IF NOT EXISTS frontier_nearest_clusters_immutable_delete
BEFORE DELETE ON frontier_nearest_clusters
BEGIN
  SELECT RAISE(ABORT, 'frontier nearest clusters are immutable');
END;
`

// migrateV14FrontierGeneration creates the M5.1 frontier tables and widens the
// provider_invocations.role CHECK to admit 'generate'. Both steps are guarded
// and idempotent: the DDL is IF NOT EXISTS throughout, and the CHECK edit runs
// only when 'generate' is not already permitted (a fresh v14 database is
// untouched).
func migrateV14FrontierGeneration(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, frontierTablesSQL); err != nil {
		return err
	}
	allows, err := providerRoleAllows(ctx, tx, "generate")
	if err != nil {
		return err
	}
	if !allows {
		if err := editTableCheckInPlace(ctx, tx, "provider_invocations",
			"role IN ('normalize','invariant','challenge')",
			"role IN ('normalize','invariant','challenge','generate')"); err != nil {
			return err
		}
	}
	return nil
}

// evaluationTablesSQL is the additive DDL for the M5.2 evaluation layer. Every
// table is immutable by trigger and append-only. The KTD-1 columns
// verifier_kind + verification_strength make the verification hierarchy
// structural (a model verdict can never be stored as if deterministic). Holdout
// mode is present in the CHECK vocabulary and its nullable columns are retained
// (M7-ready), but a holdout-REFUSAL gate trigger aborts any holdout-mode row:
// the blueprint's full gate queries holdout_leakage_check/holdout_set, which do
// not exist until M7, so a verbatim copy cannot be created now. M7 replaces this
// refusal trigger with the full gate when it lands those tables.
const evaluationTablesSQL = `
CREATE TABLE IF NOT EXISTS evaluation_runs (
  id TEXT PRIMARY KEY,
  problem_id TEXT NOT NULL REFERENCES problems(id),
  run_id TEXT NOT NULL REFERENCES runs(id),
  frontier_generation_run_id TEXT REFERENCES frontier_generation_runs(id),
  invariant_revision_id TEXT REFERENCES invariant_revisions(id),
  cluster_run_id TEXT REFERENCES cluster_runs(id),
  normalization_revision_id TEXT REFERENCES normalization_revisions(id),
  -- holdout hooks retained for M7 (nullable; holdout mode refused by gate below):
  holdout_set_id TEXT,
  holdout_leakage_check_id TEXT,
  mode TEXT NOT NULL CHECK (mode IN ('proposal','holdout')),
  cutoff_time TEXT,
  baseline_type TEXT CHECK (baseline_type IN ('undirected','semantic-summary')),
  proposal_budget_count INTEGER,
  evaluation_budget_count INTEGER,
  routing_policy TEXT NOT NULL DEFAULT 'cheap-first' CHECK (routing_policy = 'cheap-first'),
  evaluation_count INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL
);

-- Holdout-REFUSAL gate (M5.2): holdout mode is deferred to M7. Until the holdout
-- machinery (holdout_set*, holdout_leakage_check) ships, any holdout-mode row is
-- rejected here so no half-populated holdout path can exist. M7 replaces this
-- with the blueprint's full leakage-checking gate.
CREATE TRIGGER IF NOT EXISTS evaluation_runs_holdout_gate_insert
BEFORE INSERT ON evaluation_runs
BEGIN
  SELECT CASE WHEN NEW.mode = 'holdout'
    THEN RAISE(ABORT, 'holdout mode is deferred to M7; only mode=proposal is supported')
  END;
END;
CREATE TRIGGER IF NOT EXISTS evaluation_runs_holdout_gate_update
BEFORE UPDATE ON evaluation_runs
BEGIN
  SELECT CASE WHEN NEW.mode = 'holdout'
    THEN RAISE(ABORT, 'holdout mode is deferred to M7; only mode=proposal is supported')
  END;
END;

CREATE TABLE IF NOT EXISTS evaluations (
  id TEXT PRIMARY KEY,
  evaluation_run_id TEXT NOT NULL REFERENCES evaluation_runs(id),
  proposal_id TEXT REFERENCES frontier_proposals(id),
  verdict TEXT NOT NULL CHECK (verdict IN ('failure','partial_failure','partial_success','success','unknown','verification_blocked')),
  verifier_kind TEXT NOT NULL CHECK (verifier_kind IN ('deterministic-check','counterexample-search','reproducible-computation','independent-evidence','independent-critic','model-judgment')),
  verification_strength TEXT NOT NULL CHECK (verification_strength IN ('deterministic','reproducible','independent-evidence','independent-critic','single-model-judgment')),
  confidence_ordinal TEXT,
  tool_name TEXT,
  tool_version TEXT,
  provider_invocation_id TEXT REFERENCES provider_invocations(id),
  notes TEXT,
  created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS evaluation_metrics (
  id TEXT PRIMARY KEY,
  evaluation_id TEXT NOT NULL REFERENCES evaluations(id),
  metric_name TEXT NOT NULL,
  metric_scale TEXT NOT NULL CHECK (metric_scale IN ('numeric','ordinal','categorical')),
  numeric_value REAL,
  ordinal_value TEXT,
  ordinal_scale_key TEXT,
  ordinal_scale_version TEXT,
  categorical_value TEXT,
  comparator TEXT,
  created_at TEXT NOT NULL,
  CHECK (
    (metric_scale = 'numeric' AND numeric_value IS NOT NULL AND ordinal_value IS NULL AND ordinal_scale_key IS NULL AND ordinal_scale_version IS NULL AND categorical_value IS NULL) OR
    (metric_scale = 'ordinal' AND numeric_value IS NULL AND ordinal_value IS NOT NULL AND ordinal_scale_key IS NOT NULL AND ordinal_scale_version IS NOT NULL AND categorical_value IS NULL) OR
    (metric_scale = 'categorical' AND numeric_value IS NULL AND ordinal_value IS NULL AND ordinal_scale_key IS NULL AND ordinal_scale_version IS NULL AND categorical_value IS NOT NULL)
  )
);

CREATE TABLE IF NOT EXISTS evaluation_run_metrics (
  id TEXT PRIMARY KEY,
  evaluation_run_id TEXT NOT NULL REFERENCES evaluation_runs(id),
  metric_name TEXT NOT NULL,
  metric_scale TEXT NOT NULL CHECK (metric_scale IN ('numeric','ordinal','categorical')),
  numeric_value REAL,
  ordinal_value TEXT,
  ordinal_scale_key TEXT,
  ordinal_scale_version TEXT,
  categorical_value TEXT,
  comparator TEXT,
  created_at TEXT NOT NULL,
  CHECK (
    (metric_scale = 'numeric' AND numeric_value IS NOT NULL AND ordinal_value IS NULL AND ordinal_scale_key IS NULL AND ordinal_scale_version IS NULL AND categorical_value IS NULL) OR
    (metric_scale = 'ordinal' AND numeric_value IS NULL AND ordinal_value IS NOT NULL AND ordinal_scale_key IS NOT NULL AND ordinal_scale_version IS NOT NULL AND categorical_value IS NULL) OR
    (metric_scale = 'categorical' AND numeric_value IS NULL AND ordinal_value IS NULL AND ordinal_scale_key IS NULL AND ordinal_scale_version IS NULL AND categorical_value IS NOT NULL)
  )
);

-- Failure-atlas re-entry marker (R6/KTD-5): a failure/partial_failure evaluation
-- records the proposal as a newly-evaluated failure so a later cluster build can
-- include it. It is a persisted, queryable flag, not an auto-rerun.
CREATE TABLE IF NOT EXISTS evaluated_failures (
  evaluation_id TEXT PRIMARY KEY REFERENCES evaluations(id),
  proposal_id TEXT NOT NULL REFERENCES frontier_proposals(id),
  problem_id TEXT NOT NULL REFERENCES problems(id),
  verdict TEXT NOT NULL CHECK (verdict IN ('failure','partial_failure')),
  created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_evaluations_run ON evaluations(evaluation_run_id);
CREATE INDEX IF NOT EXISTS idx_evaluation_runs_problem ON evaluation_runs(problem_id);
CREATE INDEX IF NOT EXISTS idx_evaluated_failures_problem ON evaluated_failures(problem_id);
CREATE INDEX IF NOT EXISTS idx_evaluated_failures_proposal ON evaluated_failures(proposal_id);

CREATE TRIGGER IF NOT EXISTS evaluation_runs_immutable_update
BEFORE UPDATE ON evaluation_runs
BEGIN
  SELECT RAISE(ABORT, 'evaluation runs are immutable');
END;
CREATE TRIGGER IF NOT EXISTS evaluation_runs_immutable_delete
BEFORE DELETE ON evaluation_runs
BEGIN
  SELECT RAISE(ABORT, 'evaluation runs are immutable');
END;

CREATE TRIGGER IF NOT EXISTS evaluations_immutable_update
BEFORE UPDATE ON evaluations
BEGIN
  SELECT RAISE(ABORT, 'evaluations are immutable');
END;
CREATE TRIGGER IF NOT EXISTS evaluations_immutable_delete
BEFORE DELETE ON evaluations
BEGIN
  SELECT RAISE(ABORT, 'evaluations are immutable');
END;

CREATE TRIGGER IF NOT EXISTS evaluation_metrics_immutable_update
BEFORE UPDATE ON evaluation_metrics
BEGIN
  SELECT RAISE(ABORT, 'evaluation metrics are immutable');
END;
CREATE TRIGGER IF NOT EXISTS evaluation_metrics_immutable_delete
BEFORE DELETE ON evaluation_metrics
BEGIN
  SELECT RAISE(ABORT, 'evaluation metrics are immutable');
END;

CREATE TRIGGER IF NOT EXISTS evaluation_run_metrics_immutable_update
BEFORE UPDATE ON evaluation_run_metrics
BEGIN
  SELECT RAISE(ABORT, 'evaluation run metrics are immutable');
END;
CREATE TRIGGER IF NOT EXISTS evaluation_run_metrics_immutable_delete
BEFORE DELETE ON evaluation_run_metrics
BEGIN
  SELECT RAISE(ABORT, 'evaluation run metrics are immutable');
END;

CREATE TRIGGER IF NOT EXISTS evaluated_failures_immutable_update
BEFORE UPDATE ON evaluated_failures
BEGIN
  SELECT RAISE(ABORT, 'evaluated failures are immutable');
END;
CREATE TRIGGER IF NOT EXISTS evaluated_failures_immutable_delete
BEFORE DELETE ON evaluated_failures
BEGIN
  SELECT RAISE(ABORT, 'evaluated failures are immutable');
END;
`

// migrateV15Evaluation creates the M5.2 evaluation tables and widens the
// provider_invocations.role CHECK to admit 'evaluate'. Both steps are guarded
// and idempotent.
func migrateV15Evaluation(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, evaluationTablesSQL); err != nil {
		return err
	}
	allows, err := providerRoleAllows(ctx, tx, "evaluate")
	if err != nil {
		return err
	}
	if !allows {
		if err := editTableCheckInPlace(ctx, tx, "provider_invocations",
			"role IN ('normalize','invariant','challenge','generate')",
			"role IN ('normalize','invariant','challenge','generate','evaluate')"); err != nil {
			return err
		}
	}
	return nil
}

// migrateV16RenameEstablishedToOperatorAttested renames the terminal invariant
// lifecycle state `established` to `operator_attested` (F1). It rewrites the
// transition-table to_state CHECK, the transition-validating trigger's legality
// rule, and any existing `established` transition rows. Every step is guarded so
// a fresh v16 database (which never created an `established` row and whose v13
// DDL still spells the CHECK/trigger with `established`) is migrated exactly
// once and re-running is a no-op.
func migrateV16RenameEstablishedToOperatorAttested(ctx context.Context, tx *sql.Tx) error {
	// 1) to_state CHECK on invariant_state_transitions: 'established' -> 'operator_attested'.
	var ddl string
	if err := tx.QueryRowContext(ctx, `SELECT sql FROM sqlite_master WHERE type='table' AND name='invariant_state_transitions'`).Scan(&ddl); err != nil {
		return err
	}
	if strings.Contains(ddl, "'established'") {
		if err := editTableCheckInPlace(ctx, tx, "invariant_state_transitions",
			"'weaken', 'falsified', 'established'",
			"'weaken', 'falsified', 'operator_attested'"); err != nil {
			return err
		}
	}

	// 2) Rebuild the transition-validating trigger with the renamed target. The
	//    trigger body is immutable-by-convention (not by trigger), so DROP + CREATE
	//    is the supported reshape. Recreated verbatim except the surviving ->
	//    operator_attested legality edge.
	if _, err := tx.ExecContext(ctx, `DROP TRIGGER IF EXISTS invariant_state_transitions_validate_insert`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
CREATE TRIGGER invariant_state_transitions_validate_insert
BEFORE INSERT ON invariant_state_transitions
BEGIN
  SELECT CASE
    WHEN COALESCE((
      SELECT itc.last_transition_seq
      FROM invariant_transition_counters itc
      WHERE itc.invariant_id = NEW.invariant_id
    ), 0) = 0 THEN RAISE(ABORT, 'transition_seq requires a prior counter allocation')
    WHEN NEW.transition_seq <> (
      SELECT itc.last_transition_seq
      FROM invariant_transition_counters itc
      WHERE itc.invariant_id = NEW.invariant_id
    ) THEN RAISE(ABORT, 'transition_seq must match the atomically allocated invariant counter')
    WHEN NEW.from_state <> COALESCE((
      SELECT t.to_state
      FROM invariant_state_transitions t
      WHERE t.invariant_id = NEW.invariant_id
      ORDER BY t.transition_seq DESC
      LIMIT 1
    ), (
      SELECT ci.initial_state
      FROM candidate_invariants ci
      WHERE ci.id = NEW.invariant_id
    )) THEN RAISE(ABORT, 'from_state must match current invariant state')
    WHEN NOT (
      (NEW.from_state = 'proposed' AND NEW.to_state = 'challenged') OR
      (NEW.from_state = 'challenged' AND NEW.to_state IN ('surviving', 'weaken', 'falsified')) OR
      (NEW.from_state = 'surviving' AND NEW.to_state IN ('challenged', 'weaken', 'falsified', 'operator_attested')) OR
      (NEW.from_state = 'weaken' AND NEW.to_state IN ('challenged', 'surviving', 'falsified'))
    ) THEN RAISE(ABORT, 'invalid invariant state transition')
  END;
END;`); err != nil {
		return err
	}

	// 3) Rewrite any existing 'established' transition rows. The immutability
	//    update trigger blocks UPDATE, so drop it for the rewrite and recreate it.
	//    A fresh DB has zero such rows (no establishment has run) and this is a
	//    harmless no-op UPDATE.
	if _, err := tx.ExecContext(ctx, `DROP TRIGGER IF EXISTS invariant_state_transitions_immutable_update`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE invariant_state_transitions SET to_state = 'operator_attested' WHERE to_state = 'established'`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
CREATE TRIGGER invariant_state_transitions_immutable_update
BEFORE UPDATE ON invariant_state_transitions
BEGIN
  SELECT RAISE(ABORT, 'invariant state transitions are immutable');
END;`); err != nil {
		return err
	}
	return nil
}

// migrateV18ChallengeableAttestationAndResume widens the transition-validating
// trigger's legal-edge set (G2/G4). Two lifecycle defects are closed:
//
//   - Retry dead end (G2): a campaign that reaches no decisive outcome parks the
//     invariant at `challenged`. Without a `challenged -> challenged` edge the
//     opening transition of a resuming campaign is rejected, so `challenged`
//     candidates could never be re-attacked. The edge makes undecided campaigns
//     resumable.
//   - Attestation immunity (G4): `operator_attested` had NO outgoing edge, so an
//     operator assertion (which is NOT machine verification) became permanently
//     immune to challenge. It now admits `-> challenged` (open a campaign) and
//     `-> weaken | falsified` (a decisive challenge can overturn it).
//
// The trigger is immutable-by-convention (not by trigger), so DROP + CREATE is
// the supported reshape. Recreated verbatim from v16 except the widened
// legality block. Idempotent: a fresh v18 DB simply gets the wider rule.
func migrateV18ChallengeableAttestationAndResume(ctx context.Context, tx *sql.Tx) error {
	// The from_state CHECK never admitted operator_attested (establishment was
	// terminal). Attestation is now challengeable, so operator_attested must be a
	// legal from_state. Guarded in-place CHECK edit (v11/v13/v14/v16 precedent),
	// idempotent: skip when already widened.
	var ddl string
	if err := tx.QueryRowContext(ctx, `SELECT sql FROM sqlite_master WHERE type='table' AND name='invariant_state_transitions'`).Scan(&ddl); err != nil {
		return err
	}
	if !strings.Contains(ddl, "from_state IN ('proposed', 'challenged', 'surviving', 'weaken', 'operator_attested')") {
		if err := editTableCheckInPlace(ctx, tx, "invariant_state_transitions",
			"from_state IN ('proposed', 'challenged', 'surviving', 'weaken')",
			"from_state IN ('proposed', 'challenged', 'surviving', 'weaken', 'operator_attested')"); err != nil {
			return err
		}
	}

	if _, err := tx.ExecContext(ctx, `DROP TRIGGER IF EXISTS invariant_state_transitions_validate_insert`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
CREATE TRIGGER invariant_state_transitions_validate_insert
BEFORE INSERT ON invariant_state_transitions
BEGIN
  SELECT CASE
    WHEN COALESCE((
      SELECT itc.last_transition_seq
      FROM invariant_transition_counters itc
      WHERE itc.invariant_id = NEW.invariant_id
    ), 0) = 0 THEN RAISE(ABORT, 'transition_seq requires a prior counter allocation')
    WHEN NEW.transition_seq <> (
      SELECT itc.last_transition_seq
      FROM invariant_transition_counters itc
      WHERE itc.invariant_id = NEW.invariant_id
    ) THEN RAISE(ABORT, 'transition_seq must match the atomically allocated invariant counter')
    WHEN NEW.from_state <> COALESCE((
      SELECT t.to_state
      FROM invariant_state_transitions t
      WHERE t.invariant_id = NEW.invariant_id
      ORDER BY t.transition_seq DESC
      LIMIT 1
    ), (
      SELECT ci.initial_state
      FROM candidate_invariants ci
      WHERE ci.id = NEW.invariant_id
    )) THEN RAISE(ABORT, 'from_state must match current invariant state')
    WHEN NOT (
      (NEW.from_state = 'proposed' AND NEW.to_state = 'challenged') OR
      (NEW.from_state = 'challenged' AND NEW.to_state IN ('challenged', 'surviving', 'weaken', 'falsified')) OR
      (NEW.from_state = 'surviving' AND NEW.to_state IN ('challenged', 'weaken', 'falsified', 'operator_attested')) OR
      (NEW.from_state = 'weaken' AND NEW.to_state IN ('challenged', 'surviving', 'falsified')) OR
      (NEW.from_state = 'operator_attested' AND NEW.to_state IN ('challenged', 'weaken', 'falsified'))
    ) THEN RAISE(ABORT, 'invalid invariant state transition')
  END;
END;`); err != nil {
		return err
	}
	return nil
}

// successCompressionSQL is the additive DDL for M6.1 (migration v17): the
// proposed-signature content sidecar plus the immutable, revisioned
// success-invariant layer. Counts are exact numerators/denominators; ordinal
// bands are derived, never stored in place of the counts.
const successCompressionSQL = `
CREATE TABLE IF NOT EXISTS frontier_proposal_signatures (
  proposal_id TEXT PRIMARY KEY REFERENCES frontier_proposals(id),
  canonical_fingerprint TEXT NOT NULL,
  signature_json TEXT NOT NULL,
  created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS success_invariant_revisions (
  id TEXT PRIMARY KEY,
  problem_id TEXT NOT NULL REFERENCES problems(id),
  run_id TEXT NOT NULL REFERENCES runs(id),
  provider_invocation_id TEXT NOT NULL REFERENCES provider_invocations(id),
  compressor_version TEXT NOT NULL,
  predicate_schema TEXT NOT NULL,
  min_support INTEGER NOT NULL,
  cohort_hash TEXT NOT NULL,
  ineligible_unpersisted INTEGER NOT NULL DEFAULT 0,
  ambiguous_members INTEGER NOT NULL DEFAULT 0,
  inadmissible_conditions INTEGER NOT NULL DEFAULT 0,
  revision INTEGER NOT NULL,
  invariant_count INTEGER NOT NULL,
  created_at TEXT NOT NULL,
  UNIQUE(problem_id, cohort_hash, compressor_version, predicate_schema, min_support),
  UNIQUE(problem_id, revision)
);

CREATE TABLE IF NOT EXISTS success_invariants (
  id TEXT PRIMARY KEY,
  success_revision_id TEXT NOT NULL REFERENCES success_invariant_revisions(id),
  predicate_fingerprint TEXT NOT NULL,
  statement TEXT NOT NULL,
  abstraction_level TEXT NOT NULL,
  initial_state TEXT NOT NULL DEFAULT 'proposed' CHECK (initial_state = 'proposed'),
  progress_coverage_num INTEGER NOT NULL,
  progress_coverage_den INTEGER NOT NULL,
  nonprogressor_exclusion_num INTEGER NOT NULL,
  nonprogressor_exclusion_den INTEGER NOT NULL,
  coverage_ordinal TEXT NOT NULL CHECK (coverage_ordinal IN ('low','medium','high','unknown')),
  exclusion_ordinal TEXT NOT NULL CHECK (exclusion_ordinal IN ('low','medium','high','unknown')),
  distinct_mechanism_support INTEGER NOT NULL,
  strength_deterministic INTEGER NOT NULL DEFAULT 0,
  strength_reproducible INTEGER NOT NULL DEFAULT 0,
  strength_independent_evidence INTEGER NOT NULL DEFAULT 0,
  strength_independent_critic INTEGER NOT NULL DEFAULT 0,
  strength_model_judgment INTEGER NOT NULL DEFAULT 0,
  ordinal INTEGER NOT NULL,
  UNIQUE(success_revision_id, predicate_fingerprint)
);

CREATE TABLE IF NOT EXISTS success_invariant_predicates (
  success_invariant_id TEXT PRIMARY KEY REFERENCES success_invariants(id),
  predicate_json TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS success_invariant_broken_targets (
  success_invariant_id TEXT NOT NULL REFERENCES success_invariants(id),
  invariant_id TEXT NOT NULL REFERENCES candidate_invariants(id),
  PRIMARY KEY(success_invariant_id, invariant_id)
);

CREATE TABLE IF NOT EXISTS success_invariant_cohort_evaluations (
  success_invariant_id TEXT NOT NULL REFERENCES success_invariants(id),
  proposal_id TEXT NOT NULL REFERENCES frontier_proposals(id),
  cohort_role TEXT NOT NULL CHECK (cohort_role IN ('progressor','non_progressor')),
  verdict TEXT NOT NULL CHECK (verdict IN ('satisfies','violates','unknown')),
  verification_strength TEXT NOT NULL CHECK (verification_strength IN ('deterministic','reproducible','independent-evidence','independent-critic','single-model-judgment')),
  PRIMARY KEY(success_invariant_id, proposal_id)
);

CREATE INDEX IF NOT EXISTS idx_success_revisions_problem ON success_invariant_revisions(problem_id);
CREATE INDEX IF NOT EXISTS idx_success_invariants_revision ON success_invariants(success_revision_id);

CREATE TRIGGER IF NOT EXISTS frontier_proposal_signatures_immutable_update
BEFORE UPDATE ON frontier_proposal_signatures
BEGIN
  SELECT RAISE(ABORT, 'frontier proposal signatures are immutable');
END;
CREATE TRIGGER IF NOT EXISTS frontier_proposal_signatures_immutable_delete
BEFORE DELETE ON frontier_proposal_signatures
BEGIN
  SELECT RAISE(ABORT, 'frontier proposal signatures are immutable');
END;
CREATE TRIGGER IF NOT EXISTS success_invariant_revisions_immutable_update
BEFORE UPDATE ON success_invariant_revisions
BEGIN
  SELECT RAISE(ABORT, 'success invariant revisions are immutable');
END;
CREATE TRIGGER IF NOT EXISTS success_invariant_revisions_immutable_delete
BEFORE DELETE ON success_invariant_revisions
BEGIN
  SELECT RAISE(ABORT, 'success invariant revisions are immutable');
END;
CREATE TRIGGER IF NOT EXISTS success_invariants_immutable_update
BEFORE UPDATE ON success_invariants
BEGIN
  SELECT RAISE(ABORT, 'success invariants are immutable');
END;
CREATE TRIGGER IF NOT EXISTS success_invariants_immutable_delete
BEFORE DELETE ON success_invariants
BEGIN
  SELECT RAISE(ABORT, 'success invariants are immutable');
END;
CREATE TRIGGER IF NOT EXISTS success_invariant_predicates_immutable_update
BEFORE UPDATE ON success_invariant_predicates
BEGIN
  SELECT RAISE(ABORT, 'success invariant predicates are immutable');
END;
CREATE TRIGGER IF NOT EXISTS success_invariant_predicates_immutable_delete
BEFORE DELETE ON success_invariant_predicates
BEGIN
  SELECT RAISE(ABORT, 'success invariant predicates are immutable');
END;
CREATE TRIGGER IF NOT EXISTS success_invariant_broken_targets_immutable_update
BEFORE UPDATE ON success_invariant_broken_targets
BEGIN
  SELECT RAISE(ABORT, 'success invariant broken targets are immutable');
END;
CREATE TRIGGER IF NOT EXISTS success_invariant_broken_targets_immutable_delete
BEFORE DELETE ON success_invariant_broken_targets
BEGIN
  SELECT RAISE(ABORT, 'success invariant broken targets are immutable');
END;
CREATE TRIGGER IF NOT EXISTS success_invariant_cohort_evaluations_immutable_update
BEFORE UPDATE ON success_invariant_cohort_evaluations
BEGIN
  SELECT RAISE(ABORT, 'success invariant cohort evaluations are immutable');
END;
CREATE TRIGGER IF NOT EXISTS success_invariant_cohort_evaluations_immutable_delete
BEFORE DELETE ON success_invariant_cohort_evaluations
BEGIN
  SELECT RAISE(ABORT, 'success invariant cohort evaluations are immutable');
END;
`

// migrateV17SuccessCompression creates the M6.1 tables and widens the
// provider_invocations.role CHECK to admit 'success-compress'. Guarded and
// idempotent throughout.
func migrateV17SuccessCompression(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, successCompressionSQL); err != nil {
		return err
	}
	allows, err := providerRoleAllows(ctx, tx, "success-compress")
	if err != nil {
		return err
	}
	if !allows {
		if err := editTableCheckInPlace(ctx, tx, "provider_invocations",
			"role IN ('normalize','invariant','challenge','generate','evaluate')",
			"role IN ('normalize','invariant','challenge','generate','evaluate','success-compress')"); err != nil {
			return err
		}
	}
	return nil
}

const searchPolicySQL = `
CREATE TABLE IF NOT EXISTS search_policy_revisions (
  id TEXT PRIMARY KEY,
  problem_id TEXT NOT NULL REFERENCES problems(id),
  run_id TEXT NOT NULL REFERENCES runs(id),
  provider_invocation_id TEXT REFERENCES provider_invocations(id),
  mutator_version TEXT NOT NULL,
  policy_schema TEXT NOT NULL,
  evidence_cohort_hash TEXT NOT NULL,
  inert_proposals INTEGER NOT NULL DEFAULT 0,
  revision INTEGER NOT NULL,
  directive_count INTEGER NOT NULL,
  created_at TEXT NOT NULL,
  UNIQUE(problem_id, evidence_cohort_hash, mutator_version, policy_schema),
  UNIQUE(problem_id, revision)
);

CREATE TABLE IF NOT EXISTS search_policy_directives (
  id TEXT PRIMARY KEY,
  policy_revision_id TEXT NOT NULL REFERENCES search_policy_revisions(id),
  kind TEXT NOT NULL CHECK (kind IN ('prefer','avoid','expand','penalize')),
  target_kind TEXT NOT NULL CHECK (target_kind IN ('success_invariant','surviving_invariant','mechanism_family','redundant_attack','repeated_failure')),
  target_id TEXT NOT NULL,
  weight TEXT NOT NULL CHECK (weight IN ('low','medium','high','unknown')),
  epistemic_source TEXT NOT NULL DEFAULT '',
  ordinal INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS search_policy_provenance (
  policy_directive_id TEXT NOT NULL REFERENCES search_policy_directives(id),
  evidence_kind TEXT NOT NULL,
  evidence_ref TEXT NOT NULL,
  PRIMARY KEY(policy_directive_id, evidence_kind, evidence_ref)
);

-- Records which policy revision biased a frontier generation and the net
-- per-proposal bias applied (the reproducible "why favored/suppressed"). A
-- generation with no policy writes no rows here (unbiased == absence).
CREATE TABLE IF NOT EXISTS frontier_generation_policy (
  frontier_generation_run_id TEXT NOT NULL REFERENCES frontier_generation_runs(id),
  proposal_id TEXT NOT NULL REFERENCES frontier_proposals(id),
  policy_revision_id TEXT NOT NULL REFERENCES search_policy_revisions(id),
  net_bias INTEGER NOT NULL,
  preferred INTEGER NOT NULL DEFAULT 0,
  avoided INTEGER NOT NULL DEFAULT 0,
  penalized INTEGER NOT NULL DEFAULT 0,
  floor_protected INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY(frontier_generation_run_id, proposal_id)
);

CREATE INDEX IF NOT EXISTS idx_search_policy_revisions_problem ON search_policy_revisions(problem_id);
CREATE INDEX IF NOT EXISTS idx_search_policy_directives_revision ON search_policy_directives(policy_revision_id);

CREATE TRIGGER IF NOT EXISTS search_policy_revisions_immutable_update
BEFORE UPDATE ON search_policy_revisions
BEGIN
  SELECT RAISE(ABORT, 'search policy revisions are immutable');
END;
CREATE TRIGGER IF NOT EXISTS search_policy_revisions_immutable_delete
BEFORE DELETE ON search_policy_revisions
BEGIN
  SELECT RAISE(ABORT, 'search policy revisions are immutable');
END;
CREATE TRIGGER IF NOT EXISTS search_policy_directives_immutable_update
BEFORE UPDATE ON search_policy_directives
BEGIN
  SELECT RAISE(ABORT, 'search policy directives are immutable');
END;
CREATE TRIGGER IF NOT EXISTS search_policy_directives_immutable_delete
BEFORE DELETE ON search_policy_directives
BEGIN
  SELECT RAISE(ABORT, 'search policy directives are immutable');
END;
CREATE TRIGGER IF NOT EXISTS search_policy_provenance_immutable_update
BEFORE UPDATE ON search_policy_provenance
BEGIN
  SELECT RAISE(ABORT, 'search policy provenance is immutable');
END;
CREATE TRIGGER IF NOT EXISTS search_policy_provenance_immutable_delete
BEFORE DELETE ON search_policy_provenance
BEGIN
  SELECT RAISE(ABORT, 'search policy provenance is immutable');
END;
CREATE TRIGGER IF NOT EXISTS frontier_generation_policy_immutable_update
BEFORE UPDATE ON frontier_generation_policy
BEGIN
  SELECT RAISE(ABORT, 'frontier generation policy log is immutable');
END;
CREATE TRIGGER IF NOT EXISTS frontier_generation_policy_immutable_delete
BEFORE DELETE ON frontier_generation_policy
BEGIN
  SELECT RAISE(ABORT, 'frontier generation policy log is immutable');
END;
`

// migrateV19SearchPolicy creates the M6.2 search-policy tables and widens the
// provider_invocations.role CHECK to admit 'policy-mutate'. Guarded and
// idempotent throughout.
func migrateV19SearchPolicy(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, searchPolicySQL); err != nil {
		return err
	}
	allows, err := providerRoleAllows(ctx, tx, "policy-mutate")
	if err != nil {
		return err
	}
	if !allows {
		if err := editTableCheckInPlace(ctx, tx, "provider_invocations",
			"role IN ('normalize','invariant','challenge','generate','evaluate','success-compress')",
			"role IN ('normalize','invariant','challenge','generate','evaluate','success-compress','policy-mutate')"); err != nil {
			return err
		}
	}
	return nil
}

// migrateV20CohortEvaluationProvenance adds the evaluation_id provenance column
// to success_invariant_cohort_evaluations (H1). Additive nullable column;
// idempotent (skip when present). ADD COLUMN does not trip the table's
// UPDATE/DELETE immutability triggers.
func migrateV20CohortEvaluationProvenance(ctx context.Context, tx *sql.Tx) error {
	has, err := columnExists(ctx, tx, "success_invariant_cohort_evaluations", "evaluation_id")
	if err != nil {
		return err
	}
	if has {
		return nil
	}
	_, err = tx.ExecContext(ctx, `ALTER TABLE success_invariant_cohort_evaluations ADD COLUMN evaluation_id TEXT`)
	return err
}

// blindedExperimentSQL is the additive DDL for M7 v0 (migration v21): the
// holdout-set + leakage-audit layer and the experiment layer with
// mode-disjoint conclusion vocabularies. All rows immutable by trigger.
const blindedExperimentSQL = `
CREATE TABLE IF NOT EXISTS holdout_sets (
  id TEXT PRIMARY KEY,
  problem_id TEXT NOT NULL REFERENCES problems(id),
  target_problem_id TEXT NOT NULL REFERENCES problems(id),
  name TEXT NOT NULL,
  mode TEXT NOT NULL CHECK (mode IN ('blinded','historical')),
  cutoff_time TEXT,
  created_at TEXT NOT NULL,
  UNIQUE(problem_id, name),
  CHECK (mode <> 'historical' OR cutoff_time IS NOT NULL),
  CHECK (problem_id <> target_problem_id)
);

CREATE TABLE IF NOT EXISTS holdout_set_sources (
  holdout_set_id TEXT NOT NULL REFERENCES holdout_sets(id),
  source_id TEXT NOT NULL REFERENCES sources(id),
  PRIMARY KEY(holdout_set_id, source_id)
);

CREATE TRIGGER IF NOT EXISTS holdout_set_sources_problem_guard_insert
BEFORE INSERT ON holdout_set_sources
BEGIN
  SELECT CASE
    WHEN (SELECT hs.target_problem_id FROM holdout_sets hs WHERE hs.id = NEW.holdout_set_id)
      <> (SELECT s.problem_id FROM sources s WHERE s.id = NEW.source_id)
    THEN RAISE(ABORT, 'withheld source must belong to the holdout set target problem')
  END;
END;

CREATE TABLE IF NOT EXISTS holdout_source_dating (
  holdout_set_id TEXT NOT NULL REFERENCES holdout_sets(id),
  source_id TEXT NOT NULL REFERENCES sources(id),
  dated_at TEXT NOT NULL,
  evidence_locator TEXT NOT NULL,
  provenance TEXT NOT NULL DEFAULT '',
  PRIMARY KEY(holdout_set_id, source_id)
);

CREATE TABLE IF NOT EXISTS leakage_checks (
  id TEXT PRIMARY KEY,
  holdout_set_id TEXT NOT NULL REFERENCES holdout_sets(id),
  run_id TEXT NOT NULL REFERENCES runs(id),
  checker_version TEXT NOT NULL,
  snapshot_leaks INTEGER NOT NULL,
  normalization_leaks INTEGER NOT NULL,
  signature_leaks INTEGER NOT NULL,
  passed INTEGER NOT NULL,
  created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS experiment_runs (
  id TEXT PRIMARY KEY,
  problem_id TEXT NOT NULL REFERENCES problems(id),
  holdout_set_id TEXT NOT NULL REFERENCES holdout_sets(id),
  leakage_check_id TEXT NOT NULL REFERENCES leakage_checks(id),
  run_id TEXT NOT NULL REFERENCES runs(id),
  mode TEXT NOT NULL CHECK (mode IN ('blinded','historical')),
  recovery_rule_version TEXT NOT NULL,
  profile_version TEXT NOT NULL,
  proposal_budget_count INTEGER NOT NULL,
  evaluation_budget_count INTEGER NOT NULL,
  conclusion TEXT NOT NULL,
  identity_hash TEXT NOT NULL,
  revision INTEGER NOT NULL,
  created_at TEXT NOT NULL,
  UNIQUE(problem_id, identity_hash),
  UNIQUE(problem_id, revision),
  CHECK (
    (mode = 'blinded' AND conclusion IN ('structural_recovery','no_recovery','inconclusive')) OR
    (mode = 'historical' AND conclusion IN ('predicts_later_advance','fails_to_predict','inconclusive'))
  )
);

CREATE TABLE IF NOT EXISTS experiment_arms (
  experiment_id TEXT NOT NULL REFERENCES experiment_runs(id),
  arm TEXT NOT NULL CHECK (arm IN ('b0_undirected','b1_semantic_summary','b2_brainstorm','b3_invariant_guided')),
  frontier_generation_run_id TEXT REFERENCES frontier_generation_runs(id),
  proposal_count INTEGER NOT NULL,
  recovered INTEGER NOT NULL DEFAULT 0,
  first_recovery_rank INTEGER,
  nearest_classification TEXT NOT NULL DEFAULT '',
  distinct_family_count INTEGER NOT NULL DEFAULT 0,
  redundant_count INTEGER NOT NULL DEFAULT 0,
  stopping_condition TEXT NOT NULL CHECK (stopping_condition IN ('completed','budget_exhausted','no_information_gain','verification_blocked')),
  PRIMARY KEY(experiment_id, arm)
);

CREATE TABLE IF NOT EXISTS experiment_metrics (
  experiment_id TEXT NOT NULL REFERENCES experiment_runs(id),
  arm TEXT NOT NULL,
  metric TEXT NOT NULL,
  numerator INTEGER NOT NULL,
  denominator INTEGER NOT NULL,
  ordinal TEXT NOT NULL CHECK (ordinal IN ('low','medium','high','unknown')),
  PRIMARY KEY(experiment_id, arm, metric)
);

CREATE INDEX IF NOT EXISTS idx_holdout_sets_problem ON holdout_sets(problem_id);
CREATE INDEX IF NOT EXISTS idx_experiment_runs_problem ON experiment_runs(problem_id);

CREATE TRIGGER IF NOT EXISTS holdout_sets_immutable_update
BEFORE UPDATE ON holdout_sets
BEGIN
  SELECT RAISE(ABORT, 'holdout sets are immutable');
END;
CREATE TRIGGER IF NOT EXISTS holdout_sets_immutable_delete
BEFORE DELETE ON holdout_sets
BEGIN
  SELECT RAISE(ABORT, 'holdout sets are immutable');
END;
CREATE TRIGGER IF NOT EXISTS holdout_set_sources_immutable_update
BEFORE UPDATE ON holdout_set_sources
BEGIN
  SELECT RAISE(ABORT, 'holdout set sources are immutable');
END;
CREATE TRIGGER IF NOT EXISTS holdout_set_sources_immutable_delete
BEFORE DELETE ON holdout_set_sources
BEGIN
  SELECT RAISE(ABORT, 'holdout set sources are immutable');
END;
CREATE TRIGGER IF NOT EXISTS holdout_source_dating_immutable_update
BEFORE UPDATE ON holdout_source_dating
BEGIN
  SELECT RAISE(ABORT, 'holdout source dating are immutable');
END;
CREATE TRIGGER IF NOT EXISTS holdout_source_dating_immutable_delete
BEFORE DELETE ON holdout_source_dating
BEGIN
  SELECT RAISE(ABORT, 'holdout source dating are immutable');
END;
CREATE TRIGGER IF NOT EXISTS leakage_checks_immutable_update
BEFORE UPDATE ON leakage_checks
BEGIN
  SELECT RAISE(ABORT, 'leakage checks are immutable');
END;
CREATE TRIGGER IF NOT EXISTS leakage_checks_immutable_delete
BEFORE DELETE ON leakage_checks
BEGIN
  SELECT RAISE(ABORT, 'leakage checks are immutable');
END;
CREATE TRIGGER IF NOT EXISTS experiment_runs_immutable_update
BEFORE UPDATE ON experiment_runs
BEGIN
  SELECT RAISE(ABORT, 'experiment runs are immutable');
END;
CREATE TRIGGER IF NOT EXISTS experiment_runs_immutable_delete
BEFORE DELETE ON experiment_runs
BEGIN
  SELECT RAISE(ABORT, 'experiment runs are immutable');
END;
CREATE TRIGGER IF NOT EXISTS experiment_arms_immutable_update
BEFORE UPDATE ON experiment_arms
BEGIN
  SELECT RAISE(ABORT, 'experiment arms are immutable');
END;
CREATE TRIGGER IF NOT EXISTS experiment_arms_immutable_delete
BEFORE DELETE ON experiment_arms
BEGIN
  SELECT RAISE(ABORT, 'experiment arms are immutable');
END;
CREATE TRIGGER IF NOT EXISTS experiment_metrics_immutable_update
BEFORE UPDATE ON experiment_metrics
BEGIN
  SELECT RAISE(ABORT, 'experiment metrics are immutable');
END;
CREATE TRIGGER IF NOT EXISTS experiment_metrics_immutable_delete
BEFORE DELETE ON experiment_metrics
BEGIN
  SELECT RAISE(ABORT, 'experiment metrics are immutable');
END;
`

// migrateV21BlindedExperiment creates the M7 v0 layer, replaces the M5.2
// blanket holdout-refusal gate with the reserved leakage-keyed condition, and
// widens the role CHECK for the baseline provider roles. Guarded + idempotent.
func migrateV21BlindedExperiment(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, blindedExperimentSQL); err != nil {
		return err
	}
	// Replace the blanket refusal with the leakage-keyed gate. Recreating the
	// trigger is idempotent (drop + create), and the new condition still refuses
	// every pre-M7 caller (a NULL holdout_leakage_check_id never matches a
	// passing check), so nothing weakens for existing paths.
	gate := []string{
		`DROP TRIGGER IF EXISTS evaluation_runs_holdout_gate_insert`,
		`DROP TRIGGER IF EXISTS evaluation_runs_holdout_gate_update`,
		`CREATE TRIGGER IF NOT EXISTS evaluation_runs_holdout_gate_insert
BEFORE INSERT ON evaluation_runs
BEGIN
  SELECT CASE WHEN NEW.mode = 'holdout' AND NOT EXISTS (
    SELECT 1 FROM leakage_checks lc
    WHERE lc.id = NEW.holdout_leakage_check_id
      AND lc.holdout_set_id = NEW.holdout_set_id
      AND lc.passed = 1
  ) THEN RAISE(ABORT, 'holdout mode requires a passing leakage check bound to the holdout set')
  END;
END`,
		`CREATE TRIGGER IF NOT EXISTS evaluation_runs_holdout_gate_update
BEFORE UPDATE ON evaluation_runs
BEGIN
  SELECT CASE WHEN NEW.mode = 'holdout' AND NOT EXISTS (
    SELECT 1 FROM leakage_checks lc
    WHERE lc.id = NEW.holdout_leakage_check_id
      AND lc.holdout_set_id = NEW.holdout_set_id
      AND lc.passed = 1
  ) THEN RAISE(ABORT, 'holdout mode requires a passing leakage check bound to the holdout set')
  END;
END`,
	}
	for _, s := range gate {
		if _, err := tx.ExecContext(ctx, s); err != nil {
			return err
		}
	}
	for _, role := range []string{"summarize-next", "brainstorm"} {
		allows, err := providerRoleAllows(ctx, tx, role)
		if err != nil {
			return err
		}
		if allows {
			continue
		}
		var ddlText string
		if err := tx.QueryRowContext(ctx, `SELECT sql FROM sqlite_master WHERE type='table' AND name='provider_invocations'`).Scan(&ddlText); err != nil {
			return err
		}
		// Widen by replacing the closing paren of the role CHECK enum.
		const marker = "'policy-mutate'"
		if err := editTableCheckInPlace(ctx, tx, "provider_invocations", marker, marker+",'"+role+"'"); err != nil {
			return err
		}
	}
	return nil
}

// experimentMeasurementSQL is the additive DDL for the v22 measurement-contract
// hardening: explicit arm membership + the frozen target manifest.
const experimentMeasurementSQL = `
CREATE TABLE IF NOT EXISTS experiment_arm_proposals (
  experiment_id TEXT NOT NULL REFERENCES experiment_runs(id),
  arm TEXT NOT NULL,
  proposal_id TEXT NOT NULL REFERENCES frontier_proposals(id),
  member_rank INTEGER NOT NULL,
  assessment TEXT NOT NULL CHECK (assessment IN ('recovered','decisive_no','unknown','unassessed')),
  PRIMARY KEY(experiment_id, arm, proposal_id)
);

CREATE TABLE IF NOT EXISTS experiment_targets (
  experiment_id TEXT NOT NULL REFERENCES experiment_runs(id),
  signature_id TEXT NOT NULL REFERENCES mechanism_signatures(id),
  canonical_fingerprint TEXT NOT NULL,
  PRIMARY KEY(experiment_id, signature_id)
);

CREATE TRIGGER IF NOT EXISTS experiment_arm_proposals_immutable_update
BEFORE UPDATE ON experiment_arm_proposals
BEGIN
  SELECT RAISE(ABORT, 'experiment arm proposals are immutable');
END;
CREATE TRIGGER IF NOT EXISTS experiment_arm_proposals_immutable_delete
BEFORE DELETE ON experiment_arm_proposals
BEGIN
  SELECT RAISE(ABORT, 'experiment arm proposals are immutable');
END;
CREATE TRIGGER IF NOT EXISTS experiment_targets_immutable_update
BEFORE UPDATE ON experiment_targets
BEGIN
  SELECT RAISE(ABORT, 'experiment targets are immutable');
END;
CREATE TRIGGER IF NOT EXISTS experiment_targets_immutable_delete
BEFORE DELETE ON experiment_targets
BEGIN
  SELECT RAISE(ABORT, 'experiment targets are immutable');
END;
`

// migrateV22ExperimentMeasurementContract creates the membership + manifest
// tables and adds the assessment-count/consumption columns to experiment_arms.
// Guarded + idempotent.
func migrateV22ExperimentMeasurementContract(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, experimentMeasurementSQL); err != nil {
		return err
	}
	cols := []struct{ column, ddl string }{
		{"decisive_count", "ALTER TABLE experiment_arms ADD COLUMN decisive_count INTEGER NOT NULL DEFAULT 0"},
		{"unknown_count", "ALTER TABLE experiment_arms ADD COLUMN unknown_count INTEGER NOT NULL DEFAULT 0"},
		{"unassessed_count", "ALTER TABLE experiment_arms ADD COLUMN unassessed_count INTEGER NOT NULL DEFAULT 0"},
		{"evaluations_consumed", "ALTER TABLE experiment_arms ADD COLUMN evaluations_consumed INTEGER NOT NULL DEFAULT 0"},
	}
	for _, c := range cols {
		has, err := columnExists(ctx, tx, "experiment_arms", c.column)
		if err != nil {
			return err
		}
		if has {
			continue
		}
		if _, err := tx.ExecContext(ctx, c.ddl); err != nil {
			return err
		}
	}
	return nil
}

// signatureRevisionsSQL is the v23 DDL: append-only signature content
// revisions with immutability triggers.
const signatureRevisionsSQL = `
CREATE TABLE IF NOT EXISTS frontier_proposal_signature_revisions (
  proposal_id TEXT NOT NULL REFERENCES frontier_proposals(id),
  revision INTEGER NOT NULL,
  content_hash TEXT NOT NULL,
  canonical_fingerprint TEXT NOT NULL,
  signature_json TEXT NOT NULL,
  created_at TEXT NOT NULL,
  PRIMARY KEY(proposal_id, revision),
  UNIQUE(proposal_id, content_hash)
);

CREATE TRIGGER IF NOT EXISTS fps_revisions_immutable_update
BEFORE UPDATE ON frontier_proposal_signature_revisions
BEGIN
  SELECT RAISE(ABORT, 'signature content revisions are immutable');
END;
CREATE TRIGGER IF NOT EXISTS fps_revisions_immutable_delete
BEFORE DELETE ON frontier_proposal_signature_revisions
BEGIN
  SELECT RAISE(ABORT, 'signature content revisions are immutable');
END;
`

// migrateV23SignatureContentRevisions creates the revisions table, backfills
// revision 1 from the v17 sidecar (content hash computed over the persisted
// JSON bytes), and adds the content-hash reference columns.
func migrateV23SignatureContentRevisions(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, signatureRevisionsSQL); err != nil {
		return err
	}
	rows, err := tx.QueryContext(ctx, `SELECT proposal_id, canonical_fingerprint, signature_json, created_at FROM frontier_proposal_signatures`)
	if err != nil {
		return err
	}
	type sidecar struct{ id, fp, js, at string }
	var backfill []sidecar
	for rows.Next() {
		var sc sidecar
		if err := rows.Scan(&sc.id, &sc.fp, &sc.js, &sc.at); err != nil {
			rows.Close()
			return err
		}
		backfill = append(backfill, sc)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()
	for _, sc := range backfill {
		sum := sha256.Sum256([]byte(sc.js))
		if _, err := tx.ExecContext(ctx, `
INSERT OR IGNORE INTO frontier_proposal_signature_revisions(proposal_id, revision, content_hash, canonical_fingerprint, signature_json, created_at)
VALUES(?, 1, ?, ?, ?, ?)
`, sc.id, hex.EncodeToString(sum[:]), sc.fp, sc.js, sc.at); err != nil {
			return err
		}
	}
	cols := []struct{ table, column, ddl string }{
		{"evaluations", "signature_content_hash", "ALTER TABLE evaluations ADD COLUMN signature_content_hash TEXT NOT NULL DEFAULT ''"},
		{"experiment_arm_proposals", "signature_content_hash", "ALTER TABLE experiment_arm_proposals ADD COLUMN signature_content_hash TEXT NOT NULL DEFAULT ''"},
	}
	for _, c := range cols {
		has, err := columnExists(ctx, tx, c.table, c.column)
		if err != nil {
			return err
		}
		if has {
			continue
		}
		if _, err := tx.ExecContext(ctx, c.ddl); err != nil {
			return err
		}
	}
	return nil
}

// occurrenceBindingsSQL is the v24 DDL: the immutable generation-occurrence
// binding (which content a generation actually emitted per proposal).
const occurrenceBindingsSQL = `
CREATE TABLE IF NOT EXISTS frontier_generation_contents (
  generation_run_id TEXT NOT NULL REFERENCES frontier_generation_runs(id),
  proposal_id TEXT NOT NULL REFERENCES frontier_proposals(id),
  content_hash TEXT NOT NULL,
  created_at TEXT NOT NULL,
  PRIMARY KEY(generation_run_id, proposal_id)
);

CREATE TRIGGER IF NOT EXISTS frontier_generation_contents_immutable_update
BEFORE UPDATE ON frontier_generation_contents
BEGIN
  SELECT RAISE(ABORT, 'generation content occurrences are immutable');
END;
CREATE TRIGGER IF NOT EXISTS frontier_generation_contents_immutable_delete
BEFORE DELETE ON frontier_generation_contents
BEGIN
  SELECT RAISE(ABORT, 'generation content occurrences are immutable');
END;
`

// migrateV24OccurrenceBindings creates the occurrence table, backfills each
// proposal's ORIGIN generation from its revision-1 content (the only
// occurrence recoverable from history), and adds the pending-reassessment
// count to success revisions.
func migrateV24OccurrenceBindings(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, occurrenceBindingsSQL); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
INSERT OR IGNORE INTO frontier_generation_contents(generation_run_id, proposal_id, content_hash, created_at)
SELECT p.frontier_generation_run_id, p.id, r.content_hash, r.created_at
FROM frontier_proposals p
JOIN frontier_proposal_signature_revisions r ON r.proposal_id = p.id AND r.revision = 1
`); err != nil {
		return err
	}
	has, err := columnExists(ctx, tx, "success_invariant_revisions", "pending_reassessment")
	if err != nil {
		return err
	}
	if !has {
		if _, err := tx.ExecContext(ctx, `ALTER TABLE success_invariant_revisions ADD COLUMN pending_reassessment INTEGER NOT NULL DEFAULT 0`); err != nil {
			return err
		}
	}
	return nil
}

// fieldCompletenessSQL is the additive DDL for justified field-completeness
// declarations (migration v25): one row per (mechanism, set field) an extractor
// HONESTLY declared exhaustively extracted, with the required audit basis.
// Immutable: a declaration is provenance, never edited in place.
const fieldCompletenessSQL = `
CREATE TABLE IF NOT EXISTS mechanism_field_completeness (
  mechanism_id TEXT NOT NULL REFERENCES mechanisms(id),
  field_kind TEXT NOT NULL CHECK (field_kind IN ('representation', 'assumption', 'operator', 'preserves', 'breaks', 'auxiliary_object')),
  completeness TEXT NOT NULL CHECK (completeness IN ('complete', 'partial')),
  basis TEXT NOT NULL CHECK (length(basis) > 0),
  PRIMARY KEY(mechanism_id, field_kind)
);

CREATE TRIGGER IF NOT EXISTS mechanism_field_completeness_immutable_update
BEFORE UPDATE ON mechanism_field_completeness
BEGIN
  SELECT RAISE(ABORT, 'field completeness declarations are immutable');
END;

CREATE TRIGGER IF NOT EXISTS mechanism_field_completeness_immutable_delete
BEFORE DELETE ON mechanism_field_completeness
BEGIN
  SELECT RAISE(ABORT, 'field completeness declarations are immutable');
END;
`

func migrateV25FieldCompleteness(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, fieldCompletenessSQL)
	return err
}

// evaluationTargetVerdictsSQL is the additive DDL for assessment-context break
// verdicts (migration v26): one immutable row per (evaluation, targeted
// invariant) recording the verdict the evaluation ACTUALLY computed against
// its assessed signature content revision.
const evaluationTargetVerdictsSQL = `
CREATE TABLE IF NOT EXISTS evaluation_target_verdicts (
  evaluation_id TEXT NOT NULL REFERENCES evaluations(id),
  invariant_id TEXT NOT NULL,
  verdict TEXT NOT NULL CHECK (verdict IN ('satisfies', 'violates', 'unknown')),
  violated INTEGER NOT NULL CHECK (violated IN (0, 1)),
  PRIMARY KEY(evaluation_id, invariant_id)
);

CREATE TRIGGER IF NOT EXISTS evaluation_target_verdicts_immutable_update
BEFORE UPDATE ON evaluation_target_verdicts
BEGIN
  SELECT RAISE(ABORT, 'evaluation target verdicts are immutable');
END;

CREATE TRIGGER IF NOT EXISTS evaluation_target_verdicts_immutable_delete
BEFORE DELETE ON evaluation_target_verdicts
BEGIN
  SELECT RAISE(ABORT, 'evaluation target verdicts are immutable');
END;
`

func migrateV26CompletenessAdmissionAndTargetVerdicts(ctx context.Context, tx *sql.Tx) error {
	// (1) Admission split on mechanism_field_completeness. ALTER ... ADD COLUMN
	// is DDL and does not fire the immutability triggers. Existing v25 rows
	// carried no scope/admission, so they are backfilled DECLARED_ONLY —
	// preserving the weaker epistemic type rather than granting retroactive
	// evaluation authority; re-normalizing re-admits under the new contract.
	for _, col := range []struct{ name, ddl string }{
		{"scope", `ALTER TABLE mechanism_field_completeness ADD COLUMN scope TEXT NOT NULL DEFAULT 'declared_payload' CHECK (scope IN ('declared_payload', 'mechanism_exhaustive'))`},
		{"admission", `ALTER TABLE mechanism_field_completeness ADD COLUMN admission TEXT NOT NULL DEFAULT 'declared_only' CHECK (admission IN ('accepted', 'declared_only'))`},
		{"admission_basis", `ALTER TABLE mechanism_field_completeness ADD COLUMN admission_basis TEXT NOT NULL DEFAULT 'v26 backfill: pre-admission declaration; re-normalize to re-admit'`},
	} {
		has, err := columnExists(ctx, tx, "mechanism_field_completeness", col.name)
		if err != nil {
			return err
		}
		if !has {
			if _, err := tx.ExecContext(ctx, col.ddl); err != nil {
				return err
			}
		}
	}

	// (2) Assessment-context break verdicts + backfill: existing evaluations
	// historically consumed the origin-time frontier_target_invariants
	// verdicts (recomputation against occurrence bytes reproduced them for
	// origin content), so those rows are copied per evaluation. Imperfect for
	// any pre-v26 evaluation of a revised occurrence — the binding is
	// unknowable there and the origin verdict is what the pre-v26 cohort
	// query consumed anyway.
	if _, err := tx.ExecContext(ctx, evaluationTargetVerdictsSQL); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `
INSERT OR IGNORE INTO evaluation_target_verdicts(evaluation_id, invariant_id, verdict, violated)
SELECT e.id, t.invariant_id, t.verdict, t.violated
FROM evaluations e
JOIN frontier_target_invariants t ON t.proposal_id = e.proposal_id
`)
	return err
}

func migrateV27TargetVerdictProvenance(ctx context.Context, tx *sql.Tx) error {
	has, err := columnExists(ctx, tx, "evaluation_target_verdicts", "provenance")
	if err != nil {
		return err
	}
	if !has {
		if _, err := tx.ExecContext(ctx, `ALTER TABLE evaluation_target_verdicts ADD COLUMN provenance TEXT NOT NULL DEFAULT 'recomputed' CHECK (provenance IN ('recomputed', 'unverified_legacy'))`); err != nil {
			return err
		}
	}
	// Reclassify unestablishable bindings. The rows are immutable by trigger;
	// a migration legitimately owns this one-time reclassification, so the
	// triggers are dropped around the UPDATE and recreated. A binding is
	// unestablishable when the evaluation recorded NO assessed hash and the
	// proposal retains MULTIPLE content revisions — which bytes it assessed is
	// unknowable, so its verdict rows must not read as assessment-context
	// facts. (Hash-less on a single-revision proposal is establishable: only
	// one content ever existed. A recorded hash pins the binding directly.)
	for _, ddl := range []string{
		`DROP TRIGGER IF EXISTS evaluation_target_verdicts_immutable_update`,
		`UPDATE evaluation_target_verdicts SET provenance = 'unverified_legacy'
WHERE provenance = 'recomputed' AND evaluation_id IN (
  SELECT e.id FROM evaluations e
  JOIN (SELECT proposal_id, COUNT(*) AS c FROM frontier_proposal_signature_revisions GROUP BY proposal_id) rc
    ON rc.proposal_id = e.proposal_id
  WHERE COALESCE(e.signature_content_hash, '') = '' AND rc.c > 1
)`,
		`CREATE TRIGGER IF NOT EXISTS evaluation_target_verdicts_immutable_update
BEFORE UPDATE ON evaluation_target_verdicts
BEGIN
  SELECT RAISE(ABORT, 'evaluation target verdicts are immutable');
END`,
	} {
		if _, err := tx.ExecContext(ctx, ddl); err != nil {
			return err
		}
	}
	return nil
}

// compressionSelectionsSQL is the additive DDL for the compression
// execution/selection log (migration v28): one immutable row per compression
// execution recording which artifact that execution selected — created OR
// reused. "Current guidance" is the latest selection, never MAX(revision).
const compressionSelectionsSQL = `
CREATE TABLE IF NOT EXISTS success_compression_selections (
  run_id TEXT PRIMARY KEY REFERENCES runs(id),
  problem_id TEXT NOT NULL REFERENCES problems(id),
  success_revision_id TEXT NOT NULL REFERENCES success_invariant_revisions(id),
  created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_success_compression_selections_problem
ON success_compression_selections(problem_id, created_at);

CREATE TRIGGER IF NOT EXISTS success_compression_selections_immutable_update
BEFORE UPDATE ON success_compression_selections
BEGIN
  SELECT RAISE(ABORT, 'compression selections are immutable');
END;

CREATE TRIGGER IF NOT EXISTS success_compression_selections_immutable_delete
BEFORE DELETE ON success_compression_selections
BEGIN
  SELECT RAISE(ABORT, 'compression selections are immutable');
END;
`

func migrateV28CompressionSelections(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, compressionSelectionsSQL); err != nil {
		return err
	}
	// Backfill: each existing revision was selected by the execution that
	// created it. Pre-v28 reuse executions left no durable trace, so their
	// selections are unrecoverable; the creating selection is the honest floor.
	_, err := tx.ExecContext(ctx, `
INSERT OR IGNORE INTO success_compression_selections(run_id, problem_id, success_revision_id, created_at)
SELECT run_id, problem_id, id, created_at FROM success_invariant_revisions
`)
	return err
}

func migrateV29AdmissionAudit(ctx context.Context, tx *sql.Tx) error {
	for _, col := range []string{"admission_corrected", "admission_downgraded", "admission_stripped", "admission_rejected"} {
		has, err := columnExists(ctx, tx, "frontier_generation_runs", col)
		if err != nil {
			return err
		}
		if !has {
			if _, err := tx.ExecContext(ctx, `ALTER TABLE frontier_generation_runs ADD COLUMN `+col+` INTEGER NOT NULL DEFAULT 0`); err != nil {
				return err
			}
		}
	}
	return nil
}

// policySelectionsSQL mirrors success_compression_selections (v28) for the
// search-policy layer (v30): one immutable row per mutation execution.
const policySelectionsSQL = `
CREATE TABLE IF NOT EXISTS policy_mutation_selections (
  run_id TEXT PRIMARY KEY REFERENCES runs(id),
  problem_id TEXT NOT NULL REFERENCES problems(id),
  policy_revision_id TEXT NOT NULL REFERENCES search_policy_revisions(id),
  created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_policy_mutation_selections_problem
ON policy_mutation_selections(problem_id, created_at);

CREATE TRIGGER IF NOT EXISTS policy_mutation_selections_immutable_update
BEFORE UPDATE ON policy_mutation_selections
BEGIN
  SELECT RAISE(ABORT, 'policy selections are immutable');
END;

CREATE TRIGGER IF NOT EXISTS policy_mutation_selections_immutable_delete
BEFORE DELETE ON policy_mutation_selections
BEGIN
  SELECT RAISE(ABORT, 'policy selections are immutable');
END;
`

func migrateV30PolicySelections(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(ctx, policySelectionsSQL); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `
INSERT OR IGNORE INTO policy_mutation_selections(run_id, problem_id, policy_revision_id, created_at)
SELECT run_id, problem_id, id, created_at FROM search_policy_revisions
`)
	return err
}

func migrateV31AdmissionOverflow(ctx context.Context, tx *sql.Tx) error {
	has, err := columnExists(ctx, tx, "frontier_generation_runs", "admission_overflow")
	if err != nil {
		return err
	}
	if !has {
		if _, err := tx.ExecContext(ctx, `ALTER TABLE frontier_generation_runs ADD COLUMN admission_overflow INTEGER NOT NULL DEFAULT 0`); err != nil {
			return err
		}
	}
	return nil
}

// experimentExecutionsSQL is the additive DDL for execution attribution
// (migration v32): one immutable row per (experiment execution run, arm).
const experimentExecutionsSQL = `
CREATE TABLE IF NOT EXISTS experiment_executions (
  run_id TEXT NOT NULL REFERENCES runs(id),
  arm TEXT NOT NULL,
  problem_id TEXT NOT NULL REFERENCES problems(id),
  experiment_id TEXT NOT NULL REFERENCES experiment_runs(id),
  frontier_generation_run_id TEXT NOT NULL REFERENCES frontier_generation_runs(id),
  proposals_file_sha256 TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  PRIMARY KEY(run_id, arm)
);

CREATE TRIGGER IF NOT EXISTS experiment_executions_immutable_update
BEFORE UPDATE ON experiment_executions
BEGIN
  SELECT RAISE(ABORT, 'experiment executions are immutable');
END;

CREATE TRIGGER IF NOT EXISTS experiment_executions_immutable_delete
BEFORE DELETE ON experiment_executions
BEGIN
  SELECT RAISE(ABORT, 'experiment executions are immutable');
END;
`

func migrateV32ExperimentExecutions(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, experimentExecutionsSQL)
	return err
}

// interpretationClaimsSQL is the additive DDL for operator-adjudicated
// interpretation claims (migration v33). One immutable row per
// (mechanism, field_kind, normalized label); provenance names the adjudication
// artifact so a claim is always traceable to its accepted ledger entry.
const interpretationClaimsSQL = `
CREATE TABLE IF NOT EXISTS interpretation_claims (
  id TEXT PRIMARY KEY,
  problem_id TEXT NOT NULL REFERENCES problems(id),
  mechanism_id TEXT NOT NULL REFERENCES mechanisms(id),
  field_kind TEXT NOT NULL CHECK (field_kind IN ('representation','operator','assumption','preserves','breaks','auxiliary_object')),
  surface_label TEXT NOT NULL,
  label_normalized TEXT NOT NULL,
  provenance_ref TEXT NOT NULL,
  basis TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  UNIQUE(mechanism_id, field_kind, label_normalized)
);

CREATE INDEX IF NOT EXISTS idx_interpretation_claims_mechanism
  ON interpretation_claims(mechanism_id);

CREATE TRIGGER IF NOT EXISTS interpretation_claims_immutable_update
BEFORE UPDATE ON interpretation_claims
BEGIN
  SELECT RAISE(ABORT, 'interpretation claims are immutable');
END;

CREATE TRIGGER IF NOT EXISTS interpretation_claims_immutable_delete
BEFORE DELETE ON interpretation_claims
BEGIN
  SELECT RAISE(ABORT, 'interpretation claims are immutable');
END;
`

func migrateV33InterpretationClaims(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, interpretationClaimsSQL)
	return err
}

// challengeAssessmentPopulationsSQL is the additive DDL for migration v34.
// One row per challenge campaign (run_id, invariant_id). The DISCOVERY run is
// the cluster run the candidate's mining revision was derived over (claim
// scope); the ASSESSMENT run is the population the campaign's evidence
// searches (known-counterexample, success-preserving) actually ran against.
// Claim-scope attacks (bias-critique recount, split, merge, synthetic
// grounding) always run over the discovery population, so a support recount
// never conflates two populations. population_policy records what the
// operator requested; equal run ids mean no newer compatible population
// existed (or replay was requested).
const challengeAssessmentPopulationsSQL = `
CREATE TABLE IF NOT EXISTS challenge_assessment_populations (
  run_id TEXT NOT NULL REFERENCES runs(id),
  invariant_id TEXT NOT NULL REFERENCES candidate_invariants(id),
  discovery_cluster_run_id TEXT NOT NULL REFERENCES cluster_runs(id),
  assessment_cluster_run_id TEXT NOT NULL REFERENCES cluster_runs(id),
  population_policy TEXT NOT NULL CHECK (population_policy IN ('discovery', 'latest')),
  created_at TEXT NOT NULL,
  PRIMARY KEY(run_id, invariant_id)
);

CREATE INDEX IF NOT EXISTS idx_challenge_assessment_populations_invariant
  ON challenge_assessment_populations(invariant_id);

CREATE TRIGGER IF NOT EXISTS challenge_assessment_populations_immutable_update
BEFORE UPDATE ON challenge_assessment_populations
BEGIN
  SELECT RAISE(ABORT, 'challenge assessment populations are immutable');
END;

CREATE TRIGGER IF NOT EXISTS challenge_assessment_populations_immutable_delete
BEFORE DELETE ON challenge_assessment_populations
BEGIN
  SELECT RAISE(ABORT, 'challenge assessment populations are immutable');
END;
`

func migrateV34ChallengeAssessmentPopulations(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, challengeAssessmentPopulationsSQL)
	return err
}

// evidenceAdmissionsSQL is the additive DDL for migration v35 (structural
// review finding S2): the typed evidence-admission ledger bridging evaluated
// failures into the atlas population.
//
// Decision semantics:
//   - `admitted`: the evaluated failure was materialized into the atlas — the
//     four materialization ids are REQUIRED (approach, approach revision,
//     mechanism, signature) and point at rows holding the exact assessed
//     signature content.
//   - `withheld`: the evaluated failure was refused entry — the four ids are
//     REQUIRED to be absent and `basis` records the refusing rule.
//
// Observation kinds (the reviewer's taxonomy, mapped onto what the verifier
// hierarchy actually records):
//   - `structural-claim-failure`: a deterministic-check failure — the proposed
//     DESCRIPTION failed its own structural claim ("claimed break did not
//     occur"). It narrows the description space, not the observed-mechanism
//     space, and is never atlas-admissible.
//   - `domain-checked-failure`: a deterministic / reproducible /
//     independent-evidence strength failure of an actual attempt.
//   - `model-judged-failure`: a model-judgment or independent-critic failure;
//     admissible only by explicit operator attestation, and permanently
//     labeled as model-judged (ModelJudgment != Verification).
//
// UNIQUE(evaluation_id, decision): an evaluation is withheld at most once and
// admitted at most once; a withheld-then-attested evaluation keeps BOTH rows,
// so supersession is visible instead of rewritten. Rows are immutable.
const evidenceAdmissionsSQL = `
CREATE TABLE IF NOT EXISTS evidence_admissions (
  id TEXT PRIMARY KEY,
  problem_id TEXT NOT NULL REFERENCES problems(id),
  run_id TEXT NOT NULL REFERENCES runs(id),
  proposal_id TEXT NOT NULL REFERENCES frontier_proposals(id),
  evaluation_id TEXT NOT NULL REFERENCES evaluations(id),
  decision TEXT NOT NULL CHECK (decision IN ('admitted','withheld')),
  observation_kind TEXT NOT NULL CHECK (observation_kind IN ('domain-checked-failure','structural-claim-failure','model-judged-failure')),
  admitted_by TEXT NOT NULL CHECK (admitted_by IN ('rule','operator')),
  basis TEXT NOT NULL CHECK (length(basis) > 0),
  content_hash TEXT NOT NULL DEFAULT '',
  approach_id TEXT REFERENCES approaches(id),
  approach_revision_id TEXT REFERENCES approach_revisions(id),
  mechanism_id TEXT REFERENCES mechanisms(id),
  signature_id TEXT REFERENCES mechanism_signatures(id),
  created_at TEXT NOT NULL,
  UNIQUE(evaluation_id, decision),
  CHECK (
    (decision = 'admitted' AND approach_id IS NOT NULL AND approach_revision_id IS NOT NULL AND mechanism_id IS NOT NULL AND signature_id IS NOT NULL)
    OR
    (decision = 'withheld' AND approach_id IS NULL AND approach_revision_id IS NULL AND mechanism_id IS NULL AND signature_id IS NULL)
  )
);

CREATE INDEX IF NOT EXISTS idx_evidence_admissions_problem ON evidence_admissions(problem_id);
CREATE INDEX IF NOT EXISTS idx_evidence_admissions_evaluation ON evidence_admissions(evaluation_id);

CREATE TRIGGER IF NOT EXISTS evidence_admissions_immutable_update
BEFORE UPDATE ON evidence_admissions
BEGIN
  SELECT RAISE(ABORT, 'evidence admissions are immutable');
END;

CREATE TRIGGER IF NOT EXISTS evidence_admissions_immutable_delete
BEFORE DELETE ON evidence_admissions
BEGIN
  SELECT RAISE(ABORT, 'evidence admissions are immutable');
END;
`

func migrateV35EvidenceAdmissions(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, evidenceAdmissionsSQL)
	return err
}

// challengeBoundaryDeltasSQL is the additive DDL for migration v36 (structural
// review finding S5, part A): the typed boundary refinement a confirmed
// challenge derived. One row per confirmed challenge with a code-derivable
// delta; unconfirmed/inert challenges have none. `condition` is the canonical
// (paraphrase-stable) serialization of the separating predicate;
// `child_fingerprints` is a JSON array of derived child predicate fingerprints
// (split: >=2, merge: exactly 1, else empty). measured_support /
// support_threshold are set only for support-recount deltas. Rows are
// immutable.
const challengeBoundaryDeltasSQL = `
CREATE TABLE IF NOT EXISTS challenge_boundary_deltas (
  challenge_id TEXT PRIMARY KEY REFERENCES invariant_challenges(id),
  kind TEXT NOT NULL CHECK (kind IN ('counterexample-separation','contrast-collapse','constructibility','support-recount','split-partition','merge-union')),
  predicate_fingerprint TEXT NOT NULL CHECK (length(predicate_fingerprint) > 0),
  condition TEXT NOT NULL CHECK (length(condition) > 0),
  measured_support INTEGER,
  support_threshold INTEGER,
  child_fingerprints TEXT NOT NULL DEFAULT '[]',
  created_at TEXT NOT NULL,
  CHECK (
    (kind = 'support-recount' AND measured_support IS NOT NULL AND support_threshold IS NOT NULL)
    OR (kind != 'support-recount' AND measured_support IS NULL AND support_threshold IS NULL)
  )
);

CREATE TRIGGER IF NOT EXISTS challenge_boundary_deltas_immutable_update
BEFORE UPDATE ON challenge_boundary_deltas
BEGIN
  SELECT RAISE(ABORT, 'challenge boundary deltas are immutable');
END;

CREATE TRIGGER IF NOT EXISTS challenge_boundary_deltas_immutable_delete
BEFORE DELETE ON challenge_boundary_deltas
BEGIN
  SELECT RAISE(ABORT, 'challenge boundary deltas are immutable');
END;
`

func migrateV36ChallengeBoundaryDeltas(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, challengeBoundaryDeltasSQL)
	return err
}

// projectionChainSQL is the additive DDL for migration v37 (S5 part B): the
// typed projection chain separating a proposed structural change from its
// concrete plan, its verification obligations, and the domain observations
// that decide them.
//
//   - projection_artifacts: one authored concrete plan revision per proposal
//     (append-only revisions; identical content refused per proposal). The
//     typed steps live in content_json (projection.Artifact, schema-versioned).
//   - projection_obligations: what must be verified for this artifact.
//     `steps-compose` is decided by code at projection time; a
//     `domain-realization` obligation is created ONLY for composing plans and
//     stays open until an external observation decides it. Recording the open
//     obligation is the honest form of "the domain checker is missing".
//   - projection_obligation_decisions: append-once (PK = obligation) verdicts.
//     evidence_kind/evidence_ref point at the DOMAIN OBSERVATION backing an
//     operator decision (an evaluation row) or record 'code-check' for the
//     deterministic composition verdict; both set or neither.
const projectionChainSQL = `
CREATE TABLE IF NOT EXISTS projection_artifacts (
  id TEXT PRIMARY KEY,
  problem_id TEXT NOT NULL REFERENCES problems(id),
  proposal_id TEXT NOT NULL REFERENCES frontier_proposals(id),
  run_id TEXT NOT NULL REFERENCES runs(id),
  revision INTEGER NOT NULL CHECK (revision >= 1),
  author_kind TEXT NOT NULL CHECK (author_kind IN ('operator','tool')),
  schema_version TEXT NOT NULL,
  content_json TEXT NOT NULL,
  content_hash TEXT NOT NULL,
  created_at TEXT NOT NULL,
  UNIQUE(proposal_id, revision),
  UNIQUE(proposal_id, content_hash)
);

CREATE INDEX IF NOT EXISTS idx_projection_artifacts_problem ON projection_artifacts(problem_id);

CREATE TABLE IF NOT EXISTS projection_obligations (
  id TEXT PRIMARY KEY,
  artifact_id TEXT NOT NULL REFERENCES projection_artifacts(id),
  ordinal INTEGER NOT NULL,
  kind TEXT NOT NULL CHECK (kind IN ('steps-compose','domain-realization')),
  statement TEXT NOT NULL CHECK (length(statement) > 0),
  checker_kind TEXT NOT NULL CHECK (checker_kind IN ('deterministic-check','external')),
  created_at TEXT NOT NULL,
  UNIQUE(artifact_id, ordinal)
);

CREATE TABLE IF NOT EXISTS projection_obligation_decisions (
  obligation_id TEXT PRIMARY KEY REFERENCES projection_obligations(id),
  run_id TEXT NOT NULL REFERENCES runs(id),
  status TEXT NOT NULL CHECK (status IN ('discharged','failed')),
  decided_by TEXT NOT NULL CHECK (decided_by IN ('code','operator')),
  basis TEXT NOT NULL CHECK (length(basis) > 0),
  evidence_kind TEXT CHECK (evidence_kind IN ('code-check','evaluation')),
  evidence_ref TEXT,
  created_at TEXT NOT NULL,
  CHECK ((evidence_kind IS NULL) = (evidence_ref IS NULL))
);

CREATE TRIGGER IF NOT EXISTS projection_artifacts_immutable_update
BEFORE UPDATE ON projection_artifacts
BEGIN
  SELECT RAISE(ABORT, 'projection artifacts are immutable');
END;

CREATE TRIGGER IF NOT EXISTS projection_artifacts_immutable_delete
BEFORE DELETE ON projection_artifacts
BEGIN
  SELECT RAISE(ABORT, 'projection artifacts are immutable');
END;

CREATE TRIGGER IF NOT EXISTS projection_obligations_immutable_update
BEFORE UPDATE ON projection_obligations
BEGIN
  SELECT RAISE(ABORT, 'projection obligations are immutable');
END;

CREATE TRIGGER IF NOT EXISTS projection_obligations_immutable_delete
BEFORE DELETE ON projection_obligations
BEGIN
  SELECT RAISE(ABORT, 'projection obligations are immutable');
END;

CREATE TRIGGER IF NOT EXISTS projection_obligation_decisions_immutable_update
BEFORE UPDATE ON projection_obligation_decisions
BEGIN
  SELECT RAISE(ABORT, 'projection obligation decisions are immutable');
END;

CREATE TRIGGER IF NOT EXISTS projection_obligation_decisions_immutable_delete
BEFORE DELETE ON projection_obligation_decisions
BEGIN
  SELECT RAISE(ABORT, 'projection obligation decisions are immutable');
END;
`

func migrateV37ProjectionChain(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, projectionChainSQL)
	return err
}

func migrateV38VerificationSubject(ctx context.Context, tx *sql.Tx) error {
	for _, stmt := range []string{
		`ALTER TABLE evaluations ADD COLUMN verification_subject TEXT CHECK (verification_subject IS NULL OR verification_subject IN ('annotation','realization','domain-goal'))`,
		`ALTER TABLE invariant_challenge_evidence ADD COLUMN verification_subject TEXT CHECK (verification_subject IS NULL OR verification_subject IN ('annotation','realization','domain-goal'))`,
	} {
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

// invariantClaimFormsSQL is the additive DDL for migration v39 (issue #21):
// the operator-authored claim form fixing a candidate's proposition shape —
// predicate (existing) + quantifier + scope + claim role + assessment context
// (v34 populations). One row per authoring act, append-only and immutable; the
// latest form (created_at, id order) governs. `universal` is the only
// quantifier that grants falsifiable-by-one-counterexample treatment at
// challenge time; `recurrent` asserts recurrence across the recorded scope;
// `existential` asserts at least one instance. claim_role records the
// interpretive standing (docs/theory/glossary.md).
const invariantClaimFormsSQL = `
CREATE TABLE IF NOT EXISTS invariant_claim_forms (
  id TEXT PRIMARY KEY,
  invariant_id TEXT NOT NULL REFERENCES candidate_invariants(id),
  quantifier TEXT NOT NULL CHECK (quantifier IN ('universal','recurrent','existential')),
  claim_role TEXT NOT NULL CHECK (claim_role IN ('regularity','obstruction','enabling_condition','boundary_hypothesis')),
  scope TEXT NOT NULL CHECK (length(scope) > 0),
  authored_by TEXT NOT NULL CHECK (authored_by = 'operator'),
  basis TEXT NOT NULL CHECK (length(basis) > 0),
  created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_invariant_claim_forms_invariant ON invariant_claim_forms(invariant_id);

CREATE TRIGGER IF NOT EXISTS invariant_claim_forms_immutable_update
BEFORE UPDATE ON invariant_claim_forms
BEGIN
  SELECT RAISE(ABORT, 'invariant claim forms are immutable');
END;

CREATE TRIGGER IF NOT EXISTS invariant_claim_forms_immutable_delete
BEFORE DELETE ON invariant_claim_forms
BEGIN
  SELECT RAISE(ABORT, 'invariant claim forms are immutable');
END;
`

func migrateV39InvariantClaimForms(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, invariantClaimFormsSQL)
	return err
}

func migrateV40ObligationDecisionStrength(ctx context.Context, tx *sql.Tx) error {
	for _, stmt := range []string{
		`ALTER TABLE projection_obligation_decisions ADD COLUMN evaluation_verdict TEXT`,
		`ALTER TABLE projection_obligation_decisions ADD COLUMN evaluation_strength TEXT`,
	} {
		if _, err := tx.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

// episodesSQL is the additive DDL for migration v41: the prospective
// two-observation episode protocol (2026-09-12 review, epistemic finding).
// Preregistration fields are frozen by trigger (only status/completed_at may
// change); commitments, observations, and revisions are append-only and
// immutable. Observations pin the v38 axes: subject is always the DOMAIN GOAL
// and strength is `reproducible` (an in-repo exact checker is a reproducible
// computation, not a formal proof of the surrounding claim). hit/miss is
// code-derived from predicted vs observed verdict — never operator-scored.
const episodesSQL = `
CREATE TABLE IF NOT EXISTS episodes (
  id TEXT PRIMARY KEY,
  problem_id TEXT NOT NULL REFERENCES problems(id),
  title TEXT NOT NULL CHECK (length(title) > 0),
  map_ref TEXT NOT NULL CHECK (length(map_ref) > 0),
  scoring_rule TEXT NOT NULL CHECK (length(scoring_rule) > 0),
  status TEXT NOT NULL CHECK (status IN ('open','completed','abandoned')),
  created_at TEXT NOT NULL,
  completed_at TEXT
);

CREATE INDEX IF NOT EXISTS idx_episodes_problem ON episodes(problem_id);

CREATE TRIGGER IF NOT EXISTS episodes_frozen_preregistration
BEFORE UPDATE ON episodes
WHEN NEW.id != OLD.id OR NEW.problem_id != OLD.problem_id
  OR NEW.title != OLD.title OR NEW.map_ref != OLD.map_ref
  OR NEW.scoring_rule != OLD.scoring_rule OR NEW.created_at != OLD.created_at
BEGIN
  SELECT RAISE(ABORT, 'episode preregistration fields are frozen');
END;

CREATE TRIGGER IF NOT EXISTS episodes_immutable_delete
BEFORE DELETE ON episodes
BEGIN
  SELECT RAISE(ABORT, 'episodes are immutable');
END;

CREATE TABLE IF NOT EXISTS episode_commitments (
  id TEXT PRIMARY KEY,
  episode_id TEXT NOT NULL REFERENCES episodes(id),
  step INTEGER NOT NULL CHECK (step IN (1,2)),
  action TEXT NOT NULL CHECK (length(action) > 0),
  prediction TEXT NOT NULL CHECK (length(prediction) > 0),
  predicted_verdict TEXT NOT NULL CHECK (predicted_verdict IN ('witness-valid','witness-invalid')),
  map_ref TEXT NOT NULL CHECK (length(map_ref) > 0),
  basis TEXT NOT NULL CHECK (length(basis) > 0),
  created_at TEXT NOT NULL,
  UNIQUE(episode_id, step)
);

CREATE TRIGGER IF NOT EXISTS episode_commitments_immutable_update
BEFORE UPDATE ON episode_commitments
BEGIN
  SELECT RAISE(ABORT, 'episode commitments are immutable');
END;
CREATE TRIGGER IF NOT EXISTS episode_commitments_immutable_delete
BEFORE DELETE ON episode_commitments
BEGIN
  SELECT RAISE(ABORT, 'episode commitments are immutable');
END;

CREATE TABLE IF NOT EXISTS episode_observations (
  id TEXT PRIMARY KEY,
  commitment_id TEXT NOT NULL UNIQUE REFERENCES episode_commitments(id),
  checker TEXT NOT NULL CHECK (length(checker) > 0),
  payload TEXT NOT NULL CHECK (length(payload) > 0),
  verdict TEXT NOT NULL CHECK (verdict IN ('witness-valid','witness-invalid')),
  score TEXT NOT NULL CHECK (score IN ('hit','miss')),
  verification_subject TEXT NOT NULL CHECK (verification_subject = 'domain-goal'),
  verification_strength TEXT NOT NULL CHECK (verification_strength = 'reproducible'),
  detail TEXT NOT NULL,
  created_at TEXT NOT NULL
);

CREATE TRIGGER IF NOT EXISTS episode_observations_immutable_update
BEFORE UPDATE ON episode_observations
BEGIN
  SELECT RAISE(ABORT, 'episode observations are immutable');
END;
CREATE TRIGGER IF NOT EXISTS episode_observations_immutable_delete
BEFORE DELETE ON episode_observations
BEGIN
  SELECT RAISE(ABORT, 'episode observations are immutable');
END;

CREATE TABLE IF NOT EXISTS episode_revisions (
  id TEXT PRIMARY KEY,
  episode_id TEXT NOT NULL REFERENCES episodes(id),
  after_step INTEGER NOT NULL CHECK (after_step = 1),
  map_ref_before TEXT NOT NULL CHECK (length(map_ref_before) > 0),
  map_ref_after TEXT NOT NULL CHECK (length(map_ref_after) > 0),
  what_changed TEXT NOT NULL CHECK (length(what_changed) > 0),
  basis TEXT NOT NULL CHECK (length(basis) > 0),
  created_at TEXT NOT NULL,
  UNIQUE(episode_id, after_step)
);

CREATE TRIGGER IF NOT EXISTS episode_revisions_immutable_update
BEFORE UPDATE ON episode_revisions
BEGIN
  SELECT RAISE(ABORT, 'episode revisions are immutable');
END;
CREATE TRIGGER IF NOT EXISTS episode_revisions_immutable_delete
BEFORE DELETE ON episode_revisions
BEGIN
  SELECT RAISE(ABORT, 'episode revisions are immutable');
END;
`

func migrateV41Episodes(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, episodesSQL)
	return err
}

// migrateV42SemanticPreservationObligations widens the
// projection_obligations.kind CHECK vocabulary to admit
// 'semantic-preservation' (issue #22, D3). Introspective + idempotent.
func migrateV42SemanticPreservationObligations(ctx context.Context, tx *sql.Tx) error {
	var ddl string
	if err := tx.QueryRowContext(ctx, `SELECT sql FROM sqlite_master WHERE type='table' AND name='projection_obligations'`).Scan(&ddl); err != nil {
		return err
	}
	if strings.Contains(ddl, "'semantic-preservation'") {
		return nil
	}
	return editTableCheckInPlace(ctx, tx, "projection_obligations",
		"kind IN ('steps-compose','domain-realization')",
		"kind IN ('steps-compose','domain-realization','semantic-preservation')")
}

// migrateV43RefutedBoundaryDirectives widens the
// search_policy_directives.target_kind CHECK vocabulary to admit
// 'refuted_boundary' (decision D5). Introspective + idempotent.
// migrateV45EvaluatedFailuresPerEvaluation rebuilds evaluated_failures with a
// per-evaluation primary key (see the v45 migration comment). FK-safe only
// under foreign_keys=OFF (the mode Migrate establishes); Migrate's pre-commit
// foreign_key_check proves no child dangled. No-op when the PK is already
// evaluation_id (fresh databases, or an already-migrated store).
func migrateV45EvaluatedFailuresPerEvaluation(ctx context.Context, tx *sql.Tx) error {
	var ddl string
	if err := tx.QueryRowContext(ctx, `SELECT sql FROM sqlite_master WHERE type='table' AND name='evaluated_failures'`).Scan(&ddl); err != nil {
		return err
	}
	if !strings.Contains(ddl, "proposal_id TEXT PRIMARY KEY") {
		return nil // already per-evaluation.
	}
	if _, err := tx.ExecContext(ctx, `CREATE TABLE evaluated_failures__rebuild (
  evaluation_id TEXT PRIMARY KEY REFERENCES evaluations(id),
  proposal_id TEXT NOT NULL REFERENCES frontier_proposals(id),
  problem_id TEXT NOT NULL REFERENCES problems(id),
  verdict TEXT NOT NULL CHECK (verdict IN ('failure','partial_failure')),
  created_at TEXT NOT NULL
)`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO evaluated_failures__rebuild (evaluation_id, proposal_id, problem_id, verdict, created_at)
SELECT evaluation_id, proposal_id, problem_id, verdict, created_at FROM evaluated_failures`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DROP TABLE evaluated_failures`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `ALTER TABLE evaluated_failures__rebuild RENAME TO evaluated_failures`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_evaluated_failures_problem ON evaluated_failures(problem_id)`); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_evaluated_failures_proposal ON evaluated_failures(proposal_id)`); err != nil {
		return err
	}
	// DROP TABLE dropped the immutability triggers; recreate them against the
	// rebuilt table (same DDL as the v14 originals).
	_, err := tx.ExecContext(ctx, `
CREATE TRIGGER IF NOT EXISTS evaluated_failures_immutable_update
BEFORE UPDATE ON evaluated_failures
BEGIN
  SELECT RAISE(ABORT, 'evaluated failures are immutable');
END;
CREATE TRIGGER IF NOT EXISTS evaluated_failures_immutable_delete
BEFORE DELETE ON evaluated_failures
BEGIN
  SELECT RAISE(ABORT, 'evaluated failures are immutable');
END;
`)
	return err
}

func migrateV43RefutedBoundaryDirectives(ctx context.Context, tx *sql.Tx) error {
	var ddl string
	if err := tx.QueryRowContext(ctx, `SELECT sql FROM sqlite_master WHERE type='table' AND name='search_policy_directives'`).Scan(&ddl); err != nil {
		return err
	}
	if strings.Contains(ddl, "'refuted_boundary'") {
		return nil
	}
	return editTableCheckInPlace(ctx, tx, "search_policy_directives",
		"target_kind IN ('success_invariant','surviving_invariant','mechanism_family','redundant_attack','repeated_failure')",
		"target_kind IN ('success_invariant','surviving_invariant','mechanism_family','redundant_attack','repeated_failure','refuted_boundary')")
}

// reviewLedgerSQL is the additive DDL for migration v44: the normative review
// ledger (G1 of the 2026-09-12 review-flow run).
//
// The review contract names four record responsibilities and forbids a second
// independently editable coverage status. This mapping preserves those
// distinctions structurally:
//
//	review_policies              versioned decision policy (input)
//	review_obligations           versioned normative obligation (input)
//	review_applicability_decisions   applies | does_not_apply, with authority
//	review_dependency_manifests      what the assessment's meaning depends on
//	review_assessments               conforms | nonconforms | inconclusive
//	review_check_attempts            procedure, inputs, executor, outcome, cost
//
// Deliberate design choices, each traceable to a contract requirement:
//
//   - NOTHING here is a CandidateInvariant. An obligation REQUIRES a property;
//     that scientific type CLAIMS a regularity over a conditioned population.
//     Separate tables, separate id kinds, no cross-promotion path.
//   - No `status` or `coverage` column exists. Coverage is DERIVED from these
//     rows by the generator; there is no field to hand-patch.
//   - Absence is not a pass. An obligation with no assessment is `unexamined`
//     by absence, so empty assessments are never needed and never created.
//   - `inconclusive` (an executed check that decided nothing) is a distinct
//     assessment outcome from a blocked check attempt, and both are distinct
//     from never having been examined. The contract forbids collapsing these.
//   - Every assessment REQUIRES a dependency manifest (NOT NULL FK): an
//     assessment whose staleness cannot be evaluated is not a usable record.
//   - Rows are immutable. A new policy revision, a reassessment, or a re-run is
//     a NEW row; the superseded one keeps its authorizer, rationale and timing.
//     Removing an unfavorable obligation therefore cannot be hidden.
//   - review_policy_obligations carries `mandatory`, so a vacuous (empty
//     mandatory set) policy is detectable rather than silently granting
//     eligibility, and `scope_justification` on the policy is NOT NULL/non-empty
//     so an authorized affirmative scope must be stated.
const reviewLedgerSQL = `
CREATE TABLE IF NOT EXISTS review_policies (
  id TEXT PRIMARY KEY,
  policy_key TEXT NOT NULL,
  revision INTEGER NOT NULL CHECK (revision >= 1),
  decision_name TEXT NOT NULL CHECK (length(decision_name) > 0),
  owner TEXT NOT NULL CHECK (length(owner) > 0),
  authority_source TEXT NOT NULL CHECK (length(authority_source) > 0),
  scope_justification TEXT NOT NULL CHECK (length(scope_justification) > 0),
  evidence_cutoff TEXT NOT NULL,
  case_budget INTEGER NOT NULL CHECK (case_budget >= 0),
  attempt_budget INTEGER NOT NULL CHECK (attempt_budget >= 0),
  provider_call_budget INTEGER NOT NULL CHECK (provider_call_budget >= 0),
  supersedes_policy_id TEXT REFERENCES review_policies(id),
  supersede_rationale TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL,
  UNIQUE(policy_key, revision)
);

CREATE TABLE IF NOT EXISTS review_obligations (
  id TEXT PRIMARY KEY,
  obligation_key TEXT NOT NULL,
  semantic_revision INTEGER NOT NULL CHECK (semantic_revision >= 1),
  requirement TEXT NOT NULL CHECK (length(requirement) > 0),
  acceptance_criteria TEXT NOT NULL CHECK (length(acceptance_criteria) > 0),
  applicability_rule TEXT NOT NULL CHECK (length(applicability_rule) > 0),
  primary_owner TEXT NOT NULL CHECK (length(primary_owner) > 0),
  created_at TEXT NOT NULL,
  UNIQUE(obligation_key, semantic_revision)
);

CREATE TABLE IF NOT EXISTS review_policy_obligations (
  policy_id TEXT NOT NULL REFERENCES review_policies(id),
  obligation_id TEXT NOT NULL REFERENCES review_obligations(id),
  mandatory INTEGER NOT NULL CHECK (mandatory IN (0, 1)),
  PRIMARY KEY(policy_id, obligation_id)
);

CREATE TABLE IF NOT EXISTS review_applicability_decisions (
  id TEXT PRIMARY KEY,
  obligation_id TEXT NOT NULL REFERENCES review_obligations(id),
  policy_id TEXT NOT NULL REFERENCES review_policies(id),
  subject_ref TEXT NOT NULL CHECK (length(subject_ref) > 0),
  decision TEXT NOT NULL CHECK (decision IN ('applies', 'does_not_apply')),
  rationale TEXT NOT NULL CHECK (length(rationale) > 0),
  authorizer TEXT NOT NULL CHECK (length(authorizer) > 0),
  created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_review_applicability_obligation
  ON review_applicability_decisions(obligation_id, policy_id);

CREATE TABLE IF NOT EXISTS review_dependency_manifests (
  id TEXT PRIMARY KEY,
  policy_id TEXT NOT NULL REFERENCES review_policies(id),
  obligation_id TEXT NOT NULL REFERENCES review_obligations(id),
  project_revision TEXT NOT NULL CHECK (length(project_revision) > 0),
  contract_hash TEXT NOT NULL DEFAULT '',
  recipe_hash TEXT NOT NULL DEFAULT '',
  evidence_cutoff TEXT NOT NULL,
  created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS review_manifest_dependencies (
  manifest_id TEXT NOT NULL REFERENCES review_dependency_manifests(id),
  ordinal INTEGER NOT NULL,
  dependency_kind TEXT NOT NULL CHECK (length(dependency_kind) > 0),
  dependency_ref TEXT NOT NULL CHECK (length(dependency_ref) > 0),
  why_relevant TEXT NOT NULL CHECK (length(why_relevant) > 0),
  PRIMARY KEY(manifest_id, ordinal)
);

CREATE TABLE IF NOT EXISTS review_check_attempts (
  id TEXT PRIMARY KEY,
  obligation_id TEXT NOT NULL REFERENCES review_obligations(id),
  policy_id TEXT NOT NULL REFERENCES review_policies(id),
  case_label TEXT NOT NULL CHECK (length(case_label) > 0),
  procedure_ref TEXT NOT NULL CHECK (length(procedure_ref) > 0),
  procedure_revision TEXT NOT NULL CHECK (length(procedure_revision) > 0),
  inputs_ref TEXT NOT NULL DEFAULT '',
  executor TEXT NOT NULL CHECK (length(executor) > 0),
  environment TEXT NOT NULL DEFAULT '',
  mode TEXT NOT NULL CHECK (mode IN ('executed', 'inspected')),
  outcome TEXT NOT NULL CHECK (outcome IN ('completed', 'inconclusive', 'blocked')),
  output_ref TEXT NOT NULL DEFAULT '',
  blocker TEXT NOT NULL DEFAULT '',
  started_at TEXT NOT NULL,
  ended_at TEXT NOT NULL,
  resource_note TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_review_check_attempts_obligation
  ON review_check_attempts(obligation_id, case_label);

CREATE TABLE IF NOT EXISTS review_assessments (
  id TEXT PRIMARY KEY,
  obligation_id TEXT NOT NULL REFERENCES review_obligations(id),
  policy_id TEXT NOT NULL REFERENCES review_policies(id),
  applicability_decision_id TEXT NOT NULL REFERENCES review_applicability_decisions(id),
  manifest_id TEXT NOT NULL REFERENCES review_dependency_manifests(id),
  subject_ref TEXT NOT NULL CHECK (length(subject_ref) > 0),
  context_ref TEXT NOT NULL DEFAULT '',
  outcome TEXT NOT NULL CHECK (outcome IN ('conforms', 'nonconforms', 'inconclusive')),
  argument TEXT NOT NULL CHECK (length(argument) > 0),
  assessor TEXT NOT NULL CHECK (length(assessor) > 0),
  created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_review_assessments_obligation
  ON review_assessments(obligation_id, policy_id);

CREATE TABLE IF NOT EXISTS review_assessment_checks (
  assessment_id TEXT NOT NULL REFERENCES review_assessments(id),
  check_attempt_id TEXT NOT NULL REFERENCES review_check_attempts(id),
  PRIMARY KEY(assessment_id, check_attempt_id)
);

CREATE TRIGGER IF NOT EXISTS review_policies_immutable_update
BEFORE UPDATE ON review_policies
BEGIN
  SELECT RAISE(ABORT, 'review policies are immutable; supersede with a new revision');
END;
CREATE TRIGGER IF NOT EXISTS review_policies_immutable_delete
BEFORE DELETE ON review_policies
BEGIN
  SELECT RAISE(ABORT, 'review policies are immutable; supersede with a new revision');
END;

CREATE TRIGGER IF NOT EXISTS review_obligations_immutable_update
BEFORE UPDATE ON review_obligations
BEGIN
  SELECT RAISE(ABORT, 'review obligations are immutable; version them individually');
END;
CREATE TRIGGER IF NOT EXISTS review_obligations_immutable_delete
BEFORE DELETE ON review_obligations
BEGIN
  SELECT RAISE(ABORT, 'review obligations are immutable; version them individually');
END;

CREATE TRIGGER IF NOT EXISTS review_applicability_decisions_immutable_update
BEFORE UPDATE ON review_applicability_decisions
BEGIN
  SELECT RAISE(ABORT, 'review applicability decisions are immutable');
END;
CREATE TRIGGER IF NOT EXISTS review_applicability_decisions_immutable_delete
BEFORE DELETE ON review_applicability_decisions
BEGIN
  SELECT RAISE(ABORT, 'review applicability decisions are immutable');
END;

CREATE TRIGGER IF NOT EXISTS review_dependency_manifests_immutable_update
BEFORE UPDATE ON review_dependency_manifests
BEGIN
  SELECT RAISE(ABORT, 'review dependency manifests are immutable');
END;
CREATE TRIGGER IF NOT EXISTS review_dependency_manifests_immutable_delete
BEFORE DELETE ON review_dependency_manifests
BEGIN
  SELECT RAISE(ABORT, 'review dependency manifests are immutable');
END;

CREATE TRIGGER IF NOT EXISTS review_check_attempts_immutable_update
BEFORE UPDATE ON review_check_attempts
BEGIN
  SELECT RAISE(ABORT, 'review check attempts are immutable; re-running is a new attempt');
END;
CREATE TRIGGER IF NOT EXISTS review_check_attempts_immutable_delete
BEFORE DELETE ON review_check_attempts
BEGIN
  SELECT RAISE(ABORT, 'review check attempts are immutable; re-running is a new attempt');
END;

CREATE TRIGGER IF NOT EXISTS review_assessments_immutable_update
BEFORE UPDATE ON review_assessments
BEGIN
  SELECT RAISE(ABORT, 'review assessments are immutable; reassessment is a new assessment');
END;
CREATE TRIGGER IF NOT EXISTS review_assessments_immutable_delete
BEFORE DELETE ON review_assessments
BEGIN
  SELECT RAISE(ABORT, 'review assessments are immutable; reassessment is a new assessment');
END;

` + reviewAssessmentReferenceScopeSQL + `

-- A blocked check attempt cannot be the basis of a CONFORMS assessment: an
-- execution blocker leaves examination unresolved, which is inconclusive at
-- best. Enforced in the schema so no writer can launder a blocker into support.
CREATE TRIGGER IF NOT EXISTS review_assessment_checks_no_blocked_conformance
BEFORE INSERT ON review_assessment_checks
BEGIN
  SELECT CASE
    WHEN (SELECT outcome FROM review_assessments WHERE id = NEW.assessment_id) = 'conforms'
     AND (SELECT outcome FROM review_check_attempts WHERE id = NEW.check_attempt_id) = 'blocked'
    THEN RAISE(ABORT, 'a blocked check attempt cannot support a conforms assessment')
  END;
END;
`
