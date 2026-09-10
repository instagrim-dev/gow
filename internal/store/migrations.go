package store

const currentSchemaVersion = 3

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
}
