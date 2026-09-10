package store

const currentSchemaVersion = 6

type migration struct {
	version int
	sql     string
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
}
