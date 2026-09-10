package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

const currentSchemaVersion = 18

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
  proposal_id TEXT PRIMARY KEY REFERENCES frontier_proposals(id),
  evaluation_id TEXT NOT NULL REFERENCES evaluations(id),
  problem_id TEXT NOT NULL REFERENCES problems(id),
  verdict TEXT NOT NULL CHECK (verdict IN ('failure','partial_failure')),
  created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_evaluations_run ON evaluations(evaluation_run_id);
CREATE INDEX IF NOT EXISTS idx_evaluation_runs_problem ON evaluation_runs(problem_id);
CREATE INDEX IF NOT EXISTS idx_evaluated_failures_problem ON evaluated_failures(problem_id);

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
