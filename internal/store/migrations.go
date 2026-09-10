package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

const currentSchemaVersion = 11

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
  created_at TEXT NOT NULL,
  UNIQUE(problem_id, schema_version, vocabulary_version, profile_version, cluster_algo_version, thresholds_hash)
);

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
  association_status TEXT NOT NULL CHECK (association_status IN ('recurring','discriminative','candidate_obstruction','unknown')),
  obstruction_is_model_hypothesis INTEGER NOT NULL DEFAULT 0,
  distinct_family_support INTEGER NOT NULL,
  failure_coverage_num INTEGER NOT NULL,
  failure_coverage_den INTEGER NOT NULL,
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
// corrupted schema edit. The immutability triggers are untouched.
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
	// A writable_schema edit bypasses parser validation, so verify the schema is
	// still coherent before trusting it.
	var res string
	if err := tx.QueryRowContext(ctx, `PRAGMA integrity_check`).Scan(&res); err != nil {
		return err
	}
	if res != "ok" {
		return fmt.Errorf("v11: integrity_check after role generalization: %s", res)
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

	// 2. cluster_runs.input_set_hash + UNIQUE rebuild. The UNIQUE key must gain
	//    input_set_hash; SQLite cannot ALTER a UNIQUE, so when the column is
	//    absent we rebuild via create-copy-drop-rename (preserving any rows and
	//    the immutable triggers).
	has, err := columnExists(ctx, tx, "cluster_runs", "input_set_hash")
	if err != nil {
		return err
	}
	if !has {
		if err := rebuildClusterRunsWithInputSetHash(ctx, tx); err != nil {
			return err
		}
	}
	return nil
}

// rebuildClusterRunsWithInputSetHash rebuilds cluster_runs to add input_set_hash
// and include it in the idempotency UNIQUE. Existing rows (if any) receive an
// empty input_set_hash, which is a distinct identity from any real population.
func rebuildClusterRunsWithInputSetHash(ctx context.Context, tx *sql.Tx) error {
	stmts := []string{
		`DROP TRIGGER IF EXISTS cluster_runs_immutable_update`,
		`DROP TRIGGER IF EXISTS cluster_runs_immutable_delete`,
		`CREATE TABLE cluster_runs__v10 (
  id TEXT PRIMARY KEY,
  problem_id TEXT NOT NULL REFERENCES problems(id),
  run_id TEXT NOT NULL REFERENCES runs(id),
  schema_version TEXT NOT NULL,
  vocabulary_version TEXT NOT NULL REFERENCES canonical_vocabulary(version),
  profile_version TEXT NOT NULL,
  cluster_algo_version TEXT NOT NULL,
  thresholds_hash TEXT NOT NULL,
  input_set_hash TEXT NOT NULL DEFAULT '',
  signature_count INTEGER NOT NULL,
  family_count INTEGER NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('clean', 'degraded')),
  created_at TEXT NOT NULL,
  UNIQUE(problem_id, schema_version, vocabulary_version, profile_version, cluster_algo_version, thresholds_hash, input_set_hash)
)`,
		`INSERT INTO cluster_runs__v10(id, problem_id, run_id, schema_version, vocabulary_version, profile_version, cluster_algo_version, thresholds_hash, input_set_hash, signature_count, family_count, status, created_at)
  SELECT id, problem_id, run_id, schema_version, vocabulary_version, profile_version, cluster_algo_version, thresholds_hash, '', signature_count, family_count, status, created_at FROM cluster_runs`,
		`DROP TABLE cluster_runs`,
		`ALTER TABLE cluster_runs__v10 RENAME TO cluster_runs`,
		`CREATE INDEX IF NOT EXISTS idx_cluster_runs_problem ON cluster_runs(problem_id)`,
		`CREATE TRIGGER IF NOT EXISTS cluster_runs_immutable_update
BEFORE UPDATE ON cluster_runs
BEGIN
  SELECT RAISE(ABORT, 'cluster runs are immutable');
END`,
		`CREATE TRIGGER IF NOT EXISTS cluster_runs_immutable_delete
BEFORE DELETE ON cluster_runs
BEGIN
  SELECT RAISE(ABORT, 'cluster runs are immutable');
END`,
	}
	for _, s := range stmts {
		if _, err := tx.ExecContext(ctx, s); err != nil {
			return err
		}
	}
	return nil
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
