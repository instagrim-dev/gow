# SQLite persistence design (`newf` v0)

> **Implementation note (slices 2–3).** The shipped corpus uses immutable,
> content-addressed `sources` + `source_snapshots` (see
> [`docs/normalization.md`](normalization.md) and the README) rather than the
> `source` + `evidence_record` split sketched below. Normalization therefore
> consumes **source snapshots** directly: `normalization_revisions` link a
> `snapshot_id` and a `provider_invocation`, and per-field provenance is stored
> in `source_supports` (`explicit | inferred | unsupported`) referencing the
> exact snapshot. The relational-first, revision-append-only, no-opaque-blob
> principles below still hold; the concrete evidence tables will be reconciled
> when evidence-unit extraction lands.

## Why SQLite

- Single-file reproducible experiments.
- Strong relational constraints for provenance-heavy graph.
- Transactional updates for command runs.
- Sufficient for v0 local CLI scale.
- Every opened SQLite connection enables `PRAGMA foreign_keys=ON`, and the
  migration/bootstrap path includes a failing orphan-insert self-test so
  relational guarantees do not silently disappear.

## Migrations advance; they are never rewritten

Migrations are an append-only, version-numbered list applied in a single
transaction; each unapplied version runs once and is recorded in
`schema_migrations`. **Editing an already-shipped migration in place is a bug**:
a database that already recorded that version will never re-run it, so a
corrected DDL silently never reaches existing databases (fresh installs look
fine; existing ones retain the old schema). To change a shipped schema, add a
*new* migration.

Most migrations are a static SQL blob. When a change must repair databases that
were created under an earlier (since-corrected) schema, a migration may instead
supply a Go `apply(ctx, tx)` func that inspects the live schema via `PRAGMA`
(e.g. `columnExists`, `columnInPrimaryKey`) and mutates only what is missing.
Such a repair must be **idempotent and a no-op on a fresh database** (which
already has the corrected schema). Migration `v9` is the reference example: it
adds absent signature-provenance columns, rebuilds `canonical_term_aliases` only
when its primary key still lacks `canonical_id`, and creates the durable
`canonical_rejected_terms` table — each guarded so a freshly-migrated database
is untouched. `TestMigrateV9RepairsPreFixV6Schema` reconstructs the actual
pre-fix v6 shapes and asserts the repair.

## ID strategy

- Stable textual IDs with prefixes (e.g., `prob_`, `src_`, `evd_`, `inv_`, `prop_`).
- ULID/UUIDv7 recommended for time-sort + uniqueness.
- Timestamps are stored at RFC3339Nano precision for deterministic event ordering.
- Unique constraints on natural dedupe keys where applicable.

## Core schema (typed relational first)

```sql
-- Problem and run provenance
CREATE TABLE problem (
  id TEXT PRIMARY KEY,
  slug TEXT UNIQUE NOT NULL,
  statement TEXT NOT NULL,
  description TEXT,
  created_at TEXT NOT NULL
);

CREATE TABLE experiment (
  id TEXT PRIMARY KEY,
  problem_id TEXT NOT NULL REFERENCES problem(id),
  name TEXT,
  created_at TEXT NOT NULL,
  UNIQUE(problem_id, name)
);

CREATE TABLE problem_tag (
  problem_id TEXT NOT NULL REFERENCES problem(id),
  tag TEXT NOT NULL,
  PRIMARY KEY(problem_id, tag)
);

CREATE TABLE experiment_assumption (
  experiment_id TEXT NOT NULL REFERENCES experiment(id),
  value TEXT NOT NULL,
  PRIMARY KEY(experiment_id, value)
);

CREATE TABLE run (
  id TEXT PRIMARY KEY,
  problem_id TEXT NOT NULL REFERENCES problem(id),
  experiment_id TEXT REFERENCES experiment(id),
  command TEXT NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('initialized', 'running', 'completed', 'failed')),
  config_hash TEXT,
  started_at TEXT NOT NULL,
  completed_at TEXT
);

CREATE TABLE provider_call (
  id TEXT PRIMARY KEY,
  run_id TEXT NOT NULL REFERENCES run(id),
  role TEXT NOT NULL CHECK (role IN ('normalize', 'cluster', 'mine', 'critic', 'generate', 'judge', 'compress')),
  provider_name TEXT NOT NULL,
  provider_version TEXT,
  model_name TEXT,
  schema_version TEXT,
  prompt_hash TEXT,
  request_fingerprint TEXT,
  created_at TEXT NOT NULL
);

CREATE TABLE run_event (
  id TEXT PRIMARY KEY,
  run_id TEXT NOT NULL REFERENCES run(id),
  phase TEXT NOT NULL,
  status TEXT NOT NULL,
  message TEXT,
  created_at TEXT NOT NULL
);

-- Immutable sources/evidence
CREATE TABLE source (
  id TEXT PRIMARY KEY,
  problem_id TEXT NOT NULL REFERENCES problem(id),
  kind TEXT NOT NULL CHECK (kind IN ('literature', 'human', 'experiment')),
  canonical_ref TEXT NOT NULL,
  content_hash TEXT NOT NULL,
  title TEXT,
  published_at TEXT,
  metadata_json TEXT,
  created_at TEXT NOT NULL,
  UNIQUE(problem_id, canonical_ref, content_hash)
);

CREATE TABLE evidence_record (
  id TEXT PRIMARY KEY,
  problem_id TEXT NOT NULL REFERENCES problem(id),
  source_id TEXT NOT NULL REFERENCES source(id),
  evidence_type TEXT NOT NULL,
  snippet TEXT NOT NULL,
  locator TEXT,
  dedupe_key TEXT NOT NULL,
  strength TEXT,
  created_at TEXT NOT NULL,
  UNIQUE(problem_id, source_id, dedupe_key)
);

CREATE TRIGGER source_immutable_update
BEFORE UPDATE ON source
BEGIN
  SELECT RAISE(ABORT, 'source is immutable');
END;

CREATE TRIGGER source_immutable_delete
BEFORE DELETE ON source
BEGIN
  SELECT RAISE(ABORT, 'source is immutable');
END;

CREATE TRIGGER evidence_record_immutable_update
BEFORE UPDATE ON evidence_record
BEGIN
  SELECT RAISE(ABORT, 'evidence_record is immutable');
END;

CREATE TRIGGER evidence_record_immutable_delete
BEFORE DELETE ON evidence_record
BEGIN
  SELECT RAISE(ABORT, 'evidence_record is immutable');
END;

-- Holdout metadata can be defined before later revisions reference it.
CREATE TABLE holdout_set (
  id TEXT PRIMARY KEY,
  problem_id TEXT NOT NULL REFERENCES problem(id),
  name TEXT NOT NULL,
  cutoff_time TEXT NOT NULL,
  created_at TEXT NOT NULL
);

CREATE TABLE holdout_set_source (
  holdout_set_id TEXT NOT NULL REFERENCES holdout_set(id),
  source_id TEXT NOT NULL REFERENCES source(id),
  PRIMARY KEY(holdout_set_id, source_id)
);

CREATE TRIGGER holdout_set_source_problem_guard_insert
BEFORE INSERT ON holdout_set_source
BEGIN
  SELECT CASE
    WHEN (
      SELECT hs.problem_id
      FROM holdout_set hs
      WHERE hs.id = NEW.holdout_set_id
    ) <> (
      SELECT s.problem_id
      FROM source s
      WHERE s.id = NEW.source_id
    ) THEN RAISE(ABORT, 'holdout_set_source problem_id mismatch')
  END;
END;

CREATE TRIGGER holdout_set_source_problem_guard_update
BEFORE UPDATE ON holdout_set_source
BEGIN
  SELECT CASE
    WHEN (
      SELECT hs.problem_id
      FROM holdout_set hs
      WHERE hs.id = NEW.holdout_set_id
    ) <> (
      SELECT s.problem_id
      FROM source s
      WHERE s.id = NEW.source_id
    ) THEN RAISE(ABORT, 'holdout_set_source problem_id mismatch')
  END;
END;

CREATE TABLE holdout_set_family_label (
  holdout_set_id TEXT NOT NULL REFERENCES holdout_set(id),
  family_label TEXT NOT NULL,
  PRIMARY KEY(holdout_set_id, family_label)
);

-- Revisioned normalization artifacts
CREATE TABLE normalization_revision (
  id TEXT PRIMARY KEY,
  problem_id TEXT NOT NULL REFERENCES problem(id),
  run_id TEXT NOT NULL REFERENCES run(id),
  parent_revision_id TEXT REFERENCES normalization_revision(id),
  excluded_holdout_set_id TEXT REFERENCES holdout_set(id),
  source_filter_json TEXT NOT NULL,
  source_manifest_hash TEXT NOT NULL,
  config_hash TEXT NOT NULL,
  status TEXT NOT NULL,
  created_at TEXT NOT NULL
);

CREATE TABLE normalization_revision_evidence (
  normalization_revision_id TEXT NOT NULL REFERENCES normalization_revision(id),
  evidence_id TEXT NOT NULL REFERENCES evidence_record(id),
  PRIMARY KEY(normalization_revision_id, evidence_id)
);

CREATE TABLE approach (
  id TEXT PRIMARY KEY,
  problem_id TEXT NOT NULL REFERENCES problem(id),
  normalization_revision_id TEXT NOT NULL REFERENCES normalization_revision(id),
  source_id TEXT NOT NULL REFERENCES source(id),
  surface_summary TEXT
);

CREATE TABLE approach_evidence (
  approach_id TEXT NOT NULL REFERENCES approach(id),
  evidence_id TEXT NOT NULL REFERENCES evidence_record(id),
  relation TEXT NOT NULL CHECK (relation IN ('primary', 'supporting', 'counterexample')),
  locator TEXT NOT NULL DEFAULT '',
  PRIMARY KEY(approach_id, evidence_id, relation, locator)
);

CREATE TABLE normalization_item (
  id TEXT PRIMARY KEY,
  normalization_revision_id TEXT NOT NULL REFERENCES normalization_revision(id),
  evidence_id TEXT NOT NULL REFERENCES evidence_record(id),
  provider_call_id TEXT REFERENCES provider_call(id),
  approach_id TEXT REFERENCES approach(id),
  status TEXT NOT NULL CHECK (status IN ('succeeded', 'failed', 'skipped')),
  error_code TEXT,
  error_message TEXT,
  created_at TEXT NOT NULL,
  UNIQUE(normalization_revision_id, evidence_id)
);

CREATE TABLE mechanism (
  id TEXT PRIMARY KEY,
  approach_id TEXT NOT NULL REFERENCES approach(id),
  notes TEXT,
  UNIQUE(approach_id)
);

CREATE TABLE mechanism_axis_vocabulary (
  version TEXT PRIMARY KEY,
  created_at TEXT NOT NULL,
  notes TEXT
);

CREATE TABLE mechanism_axis_definition (
  vocabulary_version TEXT NOT NULL REFERENCES mechanism_axis_vocabulary(version),
  axis_key TEXT NOT NULL,
  description TEXT,
  PRIMARY KEY(vocabulary_version, axis_key)
);

CREATE TABLE mechanism_axis_allowed_value (
  vocabulary_version TEXT NOT NULL,
  axis_key TEXT NOT NULL,
  value_key TEXT NOT NULL,
  description TEXT,
  PRIMARY KEY(vocabulary_version, axis_key, value_key),
  FOREIGN KEY (vocabulary_version, axis_key)
    REFERENCES mechanism_axis_definition(vocabulary_version, axis_key)
);

CREATE TABLE mechanism_representation (
  mechanism_id TEXT NOT NULL REFERENCES mechanism(id),
  value TEXT NOT NULL,
  PRIMARY KEY(mechanism_id, value)
);

CREATE TABLE mechanism_assumption (
  mechanism_id TEXT NOT NULL REFERENCES mechanism(id),
  value TEXT NOT NULL,
  PRIMARY KEY(mechanism_id, value)
);

CREATE TABLE mechanism_operator (
  mechanism_id TEXT NOT NULL REFERENCES mechanism(id),
  value TEXT NOT NULL,
  PRIMARY KEY(mechanism_id, value)
);

CREATE TABLE mechanism_preserves (
  mechanism_id TEXT NOT NULL REFERENCES mechanism(id),
  value TEXT NOT NULL,
  PRIMARY KEY(mechanism_id, value)
);

CREATE TABLE mechanism_axis_value (
  mechanism_id TEXT NOT NULL REFERENCES mechanism(id),
  axis_key TEXT NOT NULL,
  value_key TEXT NOT NULL,
  vocabulary_version TEXT NOT NULL,
  PRIMARY KEY(mechanism_id, axis_key, vocabulary_version), -- one value per axis per mechanism per vocabulary version
  FOREIGN KEY (vocabulary_version, axis_key, value_key)
    REFERENCES mechanism_axis_allowed_value(vocabulary_version, axis_key, value_key)
);

CREATE TABLE outcome (
  id TEXT PRIMARY KEY,
  approach_id TEXT NOT NULL REFERENCES approach(id),
  class TEXT NOT NULL, -- failure|partial_failure|partial_success|success
  notes TEXT
);

CREATE TABLE failure_boundary (
  id TEXT PRIMARY KEY,
  problem_id TEXT NOT NULL REFERENCES problem(id),
  normalization_revision_id TEXT NOT NULL REFERENCES normalization_revision(id),
  label TEXT NOT NULL,
  UNIQUE(normalization_revision_id, label)
);

CREATE TABLE outcome_boundary (
  outcome_id TEXT NOT NULL REFERENCES outcome(id),
  boundary_id TEXT NOT NULL REFERENCES failure_boundary(id),
  PRIMARY KEY(outcome_id, boundary_id)
);

-- Clustering revisions
CREATE TABLE cluster_revision (
  id TEXT PRIMARY KEY,
  problem_id TEXT NOT NULL REFERENCES problem(id),
  normalization_revision_id TEXT NOT NULL REFERENCES normalization_revision(id),
  run_id TEXT NOT NULL REFERENCES run(id),
  parent_cluster_revision_id TEXT REFERENCES cluster_revision(id),
  config_hash TEXT NOT NULL,
  created_at TEXT NOT NULL
);
-- Invariant: when parent_cluster_revision_id is set, parent and child must share normalization_revision_id (enforced in application logic in v0).

CREATE TABLE mechanism_cluster (
  id TEXT PRIMARY KEY,
  cluster_revision_id TEXT NOT NULL REFERENCES cluster_revision(id),
  label TEXT NOT NULL,
  rationale TEXT,
  UNIQUE(cluster_revision_id, label)
);

CREATE TABLE cluster_evidence (
  cluster_id TEXT NOT NULL REFERENCES mechanism_cluster(id),
  evidence_id TEXT NOT NULL REFERENCES evidence_record(id),
  relation TEXT NOT NULL, -- representative|boundary|counterexample
  PRIMARY KEY(cluster_id, evidence_id, relation)
);

CREATE TABLE cluster_membership (
  cluster_id TEXT NOT NULL REFERENCES mechanism_cluster(id),
  approach_id TEXT NOT NULL REFERENCES approach(id),
  membership_confidence TEXT,
  PRIMARY KEY(cluster_id, approach_id)
);

-- Invariants and lifecycle
CREATE TABLE invariant_revision (
  id TEXT PRIMARY KEY,
  problem_id TEXT NOT NULL REFERENCES problem(id),
  cluster_revision_id TEXT NOT NULL REFERENCES cluster_revision(id),
  run_id TEXT NOT NULL REFERENCES run(id),
  parent_invariant_revision_id TEXT REFERENCES invariant_revision(id),
  config_hash TEXT NOT NULL,
  created_at TEXT NOT NULL
);

CREATE TABLE candidate_invariant (
  id TEXT PRIMARY KEY,
  invariant_revision_id TEXT NOT NULL REFERENCES invariant_revision(id),
  statement TEXT NOT NULL,
  abstraction_level TEXT NOT NULL,
  initial_state TEXT NOT NULL DEFAULT 'proposed' CHECK (initial_state = 'proposed'),
  confidence_ordinal TEXT
);

CREATE TABLE invariant_transition_counter (
  invariant_id TEXT PRIMARY KEY REFERENCES candidate_invariant(id),
  last_transition_seq INTEGER NOT NULL DEFAULT 0 CHECK (last_transition_seq >= 0)
);

CREATE TABLE invariant_support_cluster (
  invariant_id TEXT NOT NULL REFERENCES candidate_invariant(id),
  cluster_id TEXT NOT NULL REFERENCES mechanism_cluster(id),
  PRIMARY KEY(invariant_id, cluster_id)
);

CREATE TABLE invariant_support_evidence (
  invariant_id TEXT NOT NULL REFERENCES candidate_invariant(id),
  evidence_id TEXT NOT NULL REFERENCES evidence_record(id),
  relation TEXT NOT NULL, -- supports|challenges
  PRIMARY KEY(invariant_id, evidence_id, relation)
);

CREATE TABLE invariant_lineage (
  parent_invariant_id TEXT NOT NULL REFERENCES candidate_invariant(id),
  child_invariant_id TEXT NOT NULL REFERENCES candidate_invariant(id),
  relation TEXT NOT NULL CHECK (relation IN ('split', 'merge', 'weaken')),
  PRIMARY KEY(parent_invariant_id, child_invariant_id, relation)
);

CREATE TABLE invariant_challenge (
  id TEXT PRIMARY KEY,
  invariant_id TEXT NOT NULL REFERENCES candidate_invariant(id),
  run_id TEXT NOT NULL REFERENCES run(id),
  challenge_type TEXT NOT NULL CHECK (challenge_type IN ('known-counterexample', 'synthetic-counterexample', 'success-preserving', 'split', 'merge', 'bias-critique')),
  result_summary TEXT,
  created_at TEXT NOT NULL
);

CREATE TABLE invariant_state_transition (
  id TEXT PRIMARY KEY,
  invariant_id TEXT NOT NULL REFERENCES candidate_invariant(id),
  challenge_id TEXT NOT NULL REFERENCES invariant_challenge(id),
  transition_seq INTEGER NOT NULL,
  from_state TEXT NOT NULL CHECK (from_state IN ('proposed', 'challenged', 'surviving', 'weaken')),
  to_state TEXT NOT NULL CHECK (to_state IN ('proposed', 'challenged', 'surviving', 'weaken', 'falsified', 'established')),
  created_at TEXT NOT NULL,
  UNIQUE(invariant_id, transition_seq)
);

CREATE TRIGGER invariant_state_transition_validate_insert
BEFORE INSERT ON invariant_state_transition
BEGIN
  SELECT CASE
    WHEN COALESCE((
      SELECT itc.last_transition_seq
      FROM invariant_transition_counter itc
      WHERE itc.invariant_id = NEW.invariant_id
    ), 0) = 0 THEN RAISE(ABORT, 'transition_seq requires a prior counter allocation')
    WHEN NEW.transition_seq <> (
      SELECT itc.last_transition_seq
      FROM invariant_transition_counter itc
      WHERE itc.invariant_id = NEW.invariant_id
    ) THEN RAISE(ABORT, 'transition_seq must match the atomically allocated invariant counter')
    WHEN NEW.from_state <> COALESCE((
      SELECT t.to_state
      FROM invariant_state_transition t
      WHERE t.invariant_id = NEW.invariant_id
      ORDER BY t.transition_seq DESC
      LIMIT 1
    ), (
      SELECT ci.initial_state
      FROM candidate_invariant ci
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

CREATE VIEW invariant_current_state AS
WITH ranked AS (
  SELECT
    t.invariant_id,
    t.to_state,
    t.created_at,
    ROW_NUMBER() OVER (
      PARTITION BY t.invariant_id
      ORDER BY t.transition_seq DESC
    ) AS rn
  FROM invariant_state_transition t
),
latest AS (
  SELECT invariant_id, to_state, created_at
  FROM ranked
  WHERE rn = 1
)
SELECT ci.id AS invariant_id,
       COALESCE(latest.to_state, ci.initial_state) AS state,
       COALESCE(latest.created_at, ir.created_at) AS as_of
FROM candidate_invariant ci
JOIN invariant_revision ir ON ir.id = ci.invariant_revision_id
LEFT JOIN latest ON latest.invariant_id = ci.id;

CREATE TABLE invariant_challenge_source_evidence (
  challenge_id TEXT NOT NULL REFERENCES invariant_challenge(id),
  evidence_id TEXT NOT NULL REFERENCES evidence_record(id),
  PRIMARY KEY(challenge_id, evidence_id)
);

CREATE TABLE synthetic_artifact (
  id TEXT PRIMARY KEY,
  problem_id TEXT NOT NULL REFERENCES problem(id),
  run_id TEXT NOT NULL REFERENCES run(id),
  artifact_type TEXT NOT NULL, -- synthetic_attempt|synthetic_counterexample
  content TEXT NOT NULL,
  created_at TEXT NOT NULL
);

CREATE TABLE invariant_challenge_synthetic_artifact (
  challenge_id TEXT NOT NULL REFERENCES invariant_challenge(id),
  synthetic_artifact_id TEXT NOT NULL REFERENCES synthetic_artifact(id),
  PRIMARY KEY(challenge_id, synthetic_artifact_id)
);

-- Frontier proposals
CREATE TABLE frontier_generation_run (
  id TEXT PRIMARY KEY,
  problem_id TEXT NOT NULL REFERENCES problem(id),
  cluster_revision_id TEXT NOT NULL REFERENCES cluster_revision(id),
  run_id TEXT NOT NULL REFERENCES run(id),
  holdout_leakage_check_id TEXT REFERENCES holdout_leakage_check(id),
  created_at TEXT NOT NULL
);

CREATE TABLE frontier_proposal (
  id TEXT PRIMARY KEY,
  problem_id TEXT NOT NULL REFERENCES problem(id),
  frontier_generation_run_id TEXT NOT NULL REFERENCES frontier_generation_run(id),
  structural_violation_claim TEXT NOT NULL,
  novelty_argument TEXT NOT NULL,
  cheapest_falsification_path TEXT NOT NULL,
  expected_information_gain_ordinal TEXT,
  evaluation_cost_ordinal TEXT,
  proposal_hash TEXT NOT NULL,
  result TEXT,
  UNIQUE(problem_id, proposal_hash)
);

CREATE TABLE frontier_target_invariant (
  proposal_id TEXT NOT NULL REFERENCES frontier_proposal(id),
  invariant_id TEXT NOT NULL REFERENCES candidate_invariant(id),
  PRIMARY KEY(proposal_id, invariant_id)
);

CREATE TABLE frontier_nearest_cluster (
  proposal_id TEXT NOT NULL REFERENCES frontier_proposal(id),
  cluster_id TEXT NOT NULL REFERENCES mechanism_cluster(id),
  proximity_ordinal TEXT,
  PRIMARY KEY(proposal_id, cluster_id)
);

CREATE TABLE holdout_leakage_check (
  id TEXT PRIMARY KEY,
  holdout_set_id TEXT NOT NULL REFERENCES holdout_set(id),
  normalization_revision_id TEXT NOT NULL REFERENCES normalization_revision(id),
  checked_scope TEXT NOT NULL CHECK (checked_scope IN ('training_sources', 'training_evidence')),
  failure_basis TEXT NOT NULL CHECK (failure_basis IN ('no_overlap', 'source_overlap', 'evidence_overlap')),
  overlap_count INTEGER NOT NULL DEFAULT 0,
  status TEXT NOT NULL CHECK (status IN ('passed', 'failed')),
  checked_at TEXT NOT NULL,
  CHECK (
    (checked_scope = 'training_sources' AND failure_basis IN ('no_overlap', 'source_overlap')) OR
    (checked_scope = 'training_evidence' AND failure_basis IN ('no_overlap', 'evidence_overlap'))
  ),
  CHECK (
    (status = 'passed' AND failure_basis = 'no_overlap' AND overlap_count = 0) OR
    (status = 'failed' AND failure_basis IN ('source_overlap', 'evidence_overlap') AND overlap_count > 0)
  ),
  UNIQUE(holdout_set_id, normalization_revision_id, checked_scope)
);

CREATE TRIGGER holdout_leakage_check_lineage_guard_insert
BEFORE INSERT ON holdout_leakage_check
BEGIN
  SELECT CASE
    WHEN (
      SELECT hs.problem_id
      FROM holdout_set hs
      WHERE hs.id = NEW.holdout_set_id
    ) <> (
      SELECT nr.problem_id
      FROM normalization_revision nr
      WHERE nr.id = NEW.normalization_revision_id
    ) THEN RAISE(ABORT, 'holdout_leakage_check problem_id mismatch')
    WHEN COALESCE((
      SELECT nr.excluded_holdout_set_id
      FROM normalization_revision nr
      WHERE nr.id = NEW.normalization_revision_id
    ), '') <> NEW.holdout_set_id THEN RAISE(ABORT, 'normalization_revision must exclude the same holdout_set')
  END;
END;

CREATE TRIGGER holdout_leakage_check_lineage_guard_update
BEFORE UPDATE ON holdout_leakage_check
BEGIN
  SELECT CASE
    WHEN (
      SELECT hs.problem_id
      FROM holdout_set hs
      WHERE hs.id = NEW.holdout_set_id
    ) <> (
      SELECT nr.problem_id
      FROM normalization_revision nr
      WHERE nr.id = NEW.normalization_revision_id
    ) THEN RAISE(ABORT, 'holdout_leakage_check problem_id mismatch')
    WHEN COALESCE((
      SELECT nr.excluded_holdout_set_id
      FROM normalization_revision nr
      WHERE nr.id = NEW.normalization_revision_id
    ), '') <> NEW.holdout_set_id THEN RAISE(ABORT, 'normalization_revision must exclude the same holdout_set')
  END;
END;

CREATE TABLE holdout_leakage_check_overlap (
  id TEXT PRIMARY KEY,
  holdout_leakage_check_id TEXT NOT NULL REFERENCES holdout_leakage_check(id),
  overlapping_source_id TEXT REFERENCES source(id),
  overlapping_evidence_id TEXT REFERENCES evidence_record(id),
  overlap_kind TEXT NOT NULL CHECK (overlap_kind IN ('source', 'evidence')),
  CHECK (
    (overlap_kind = 'source' AND overlapping_source_id IS NOT NULL AND overlapping_evidence_id IS NULL) OR
    (overlap_kind = 'evidence' AND overlapping_source_id IS NULL AND overlapping_evidence_id IS NOT NULL)
  ),
  UNIQUE(holdout_leakage_check_id, overlap_kind, overlapping_source_id),
  UNIQUE(holdout_leakage_check_id, overlap_kind, overlapping_evidence_id)
);

CREATE TRIGGER holdout_leakage_check_overlap_insert_guard
BEFORE INSERT ON holdout_leakage_check_overlap
BEGIN
  SELECT CASE
    WHEN (
      SELECT status
      FROM holdout_leakage_check
      WHERE id = NEW.holdout_leakage_check_id
    ) = 'passed' THEN RAISE(ABORT, 'passed leakage checks cannot record overlaps')
  END;
END;

CREATE TRIGGER holdout_leakage_check_overlap_immutable_update
BEFORE UPDATE ON holdout_leakage_check_overlap
BEGIN
  SELECT RAISE(ABORT, 'holdout_leakage_check_overlap rows are immutable');
END;

CREATE TRIGGER holdout_leakage_check_overlap_immutable_delete
BEFORE DELETE ON holdout_leakage_check_overlap
BEGIN
  SELECT RAISE(ABORT, 'holdout_leakage_check_overlap rows are immutable');
END;

-- Evaluation and baselines
CREATE TABLE evaluation_run (
  id TEXT PRIMARY KEY,
  problem_id TEXT NOT NULL REFERENCES problem(id),
  run_id TEXT NOT NULL REFERENCES run(id),
  holdout_set_id TEXT REFERENCES holdout_set(id),
  normalization_revision_id TEXT REFERENCES normalization_revision(id),
  cluster_revision_id TEXT REFERENCES cluster_revision(id),
  invariant_revision_id TEXT REFERENCES invariant_revision(id),
  frontier_generation_run_id TEXT REFERENCES frontier_generation_run(id),
  holdout_leakage_check_id TEXT REFERENCES holdout_leakage_check(id),
  mode TEXT NOT NULL CHECK (mode IN ('proposal', 'holdout')),
  cutoff_time TEXT,
  baseline_type TEXT CHECK (baseline_type IN ('undirected', 'semantic-summary')),
  proposal_budget_count INTEGER,
  evaluation_budget_count INTEGER,
  judge_provider_name TEXT,
  judge_provider_version TEXT,
  judge_model_name TEXT,
  judge_config_hash TEXT,
  created_at TEXT NOT NULL
);

CREATE TRIGGER evaluation_run_holdout_gate_insert
BEFORE INSERT ON evaluation_run
BEGIN
  SELECT CASE
    WHEN NEW.mode = 'holdout' AND (
      NEW.holdout_set_id IS NULL OR
      NEW.normalization_revision_id IS NULL OR
      NEW.holdout_leakage_check_id IS NULL
    ) THEN RAISE(ABORT, 'holdout mode requires holdout_set_id, normalization_revision_id, and holdout_leakage_check_id')
    WHEN NEW.mode = 'holdout' AND COALESCE((
      SELECT hlc.holdout_set_id
      FROM holdout_leakage_check hlc
      WHERE hlc.id = NEW.holdout_leakage_check_id
    ), '') <> COALESCE(NEW.holdout_set_id, '') THEN RAISE(ABORT, 'evaluation_run holdout_set_id must match holdout_leakage_check')
    WHEN NEW.mode = 'holdout' AND COALESCE((
      SELECT hlc.normalization_revision_id
      FROM holdout_leakage_check hlc
      WHERE hlc.id = NEW.holdout_leakage_check_id
    ), '') <> COALESCE(NEW.normalization_revision_id, '') THEN RAISE(ABORT, 'evaluation_run normalization_revision_id must match holdout_leakage_check')
    WHEN NEW.mode = 'holdout' AND (
      SELECT hlc.status
      FROM holdout_leakage_check hlc
      WHERE hlc.id = NEW.holdout_leakage_check_id
    ) <> 'passed' THEN RAISE(ABORT, 'holdout mode requires a passed holdout_leakage_check')
  END;
END;

CREATE TRIGGER evaluation_run_holdout_gate_update
BEFORE UPDATE ON evaluation_run
BEGIN
  SELECT CASE
    WHEN NEW.mode = 'holdout' AND (
      NEW.holdout_set_id IS NULL OR
      NEW.normalization_revision_id IS NULL OR
      NEW.holdout_leakage_check_id IS NULL
    ) THEN RAISE(ABORT, 'holdout mode requires holdout_set_id, normalization_revision_id, and holdout_leakage_check_id')
    WHEN NEW.mode = 'holdout' AND COALESCE((
      SELECT hlc.holdout_set_id
      FROM holdout_leakage_check hlc
      WHERE hlc.id = NEW.holdout_leakage_check_id
    ), '') <> COALESCE(NEW.holdout_set_id, '') THEN RAISE(ABORT, 'evaluation_run holdout_set_id must match holdout_leakage_check')
    WHEN NEW.mode = 'holdout' AND COALESCE((
      SELECT hlc.normalization_revision_id
      FROM holdout_leakage_check hlc
      WHERE hlc.id = NEW.holdout_leakage_check_id
    ), '') <> COALESCE(NEW.normalization_revision_id, '') THEN RAISE(ABORT, 'evaluation_run normalization_revision_id must match holdout_leakage_check')
    WHEN NEW.mode = 'holdout' AND (
      SELECT hlc.status
      FROM holdout_leakage_check hlc
      WHERE hlc.id = NEW.holdout_leakage_check_id
    ) <> 'passed' THEN RAISE(ABORT, 'holdout mode requires a passed holdout_leakage_check')
  END;
END;

CREATE TABLE evaluation (
  id TEXT PRIMARY KEY,
  evaluation_run_id TEXT NOT NULL REFERENCES evaluation_run(id),
  proposal_id TEXT REFERENCES frontier_proposal(id),
  verdict TEXT NOT NULL,
  confidence_ordinal TEXT,
  notes TEXT
);

CREATE TABLE evaluation_holdout_match (
  id TEXT PRIMARY KEY,
  evaluation_id TEXT NOT NULL REFERENCES evaluation(id),
  holdout_set_id TEXT NOT NULL REFERENCES holdout_set(id),
  target_key TEXT NOT NULL,
  holdout_source_id TEXT REFERENCES source(id),
  holdout_family_label TEXT,
  match_kind TEXT NOT NULL CHECK (match_kind IN ('source_recovery', 'family_recovery', 'structural_break')),
  match_verdict TEXT NOT NULL CHECK (match_verdict IN ('exact', 'equivalent', 'miss')),
  notes TEXT,
  CHECK (
    (match_kind = 'source_recovery' AND holdout_source_id IS NOT NULL AND holdout_family_label IS NULL) OR
    (match_kind = 'family_recovery' AND holdout_source_id IS NULL AND holdout_family_label IS NOT NULL) OR
    (match_kind = 'structural_break' AND holdout_source_id IS NULL AND holdout_family_label IS NOT NULL)
  ),
  UNIQUE(evaluation_id, target_key),
  FOREIGN KEY (holdout_set_id, holdout_source_id)
    REFERENCES holdout_set_source(holdout_set_id, source_id),
  FOREIGN KEY (holdout_set_id, holdout_family_label)
    REFERENCES holdout_set_family_label(holdout_set_id, family_label)
);

CREATE TRIGGER evaluation_holdout_match_holdout_set_guard
BEFORE INSERT ON evaluation_holdout_match
BEGIN
  SELECT CASE
    WHEN COALESCE((
      SELECT er.mode
      FROM evaluation e
      JOIN evaluation_run er ON er.id = e.evaluation_run_id
      WHERE e.id = NEW.evaluation_id
    ), '') <> 'holdout' THEN RAISE(ABORT, 'evaluation_holdout_match rows require a holdout-mode evaluation_run')
    WHEN COALESCE(NEW.holdout_set_id, '') <> COALESCE((
      SELECT er.holdout_set_id
      FROM evaluation e
      JOIN evaluation_run er ON er.id = e.evaluation_run_id
      WHERE e.id = NEW.evaluation_id
    ), '') THEN RAISE(ABORT, 'evaluation_holdout_match holdout_set_id must match parent evaluation_run')
  END;
END;

CREATE TRIGGER evaluation_holdout_match_holdout_set_guard_update
BEFORE UPDATE ON evaluation_holdout_match
BEGIN
  SELECT CASE
    WHEN COALESCE((
      SELECT er.mode
      FROM evaluation e
      JOIN evaluation_run er ON er.id = e.evaluation_run_id
      WHERE e.id = NEW.evaluation_id
    ), '') <> 'holdout' THEN RAISE(ABORT, 'evaluation_holdout_match rows require a holdout-mode evaluation_run')
    WHEN COALESCE(NEW.holdout_set_id, '') <> COALESCE((
      SELECT er.holdout_set_id
      FROM evaluation e
      JOIN evaluation_run er ON er.id = e.evaluation_run_id
      WHERE e.id = NEW.evaluation_id
    ), '') THEN RAISE(ABORT, 'evaluation_holdout_match holdout_set_id must match parent evaluation_run')
  END;
END;

CREATE TABLE evaluation_metric (
  id TEXT PRIMARY KEY,
  evaluation_id TEXT NOT NULL REFERENCES evaluation(id),
  metric_name TEXT NOT NULL,
  metric_scale TEXT NOT NULL CHECK (metric_scale IN ('numeric', 'ordinal', 'categorical')),
  numeric_value REAL,
  ordinal_value TEXT,
  ordinal_scale_key TEXT,
  ordinal_scale_version TEXT,
  categorical_value TEXT,
  comparator TEXT, -- baseline id/name
  created_at TEXT NOT NULL,
  CHECK (
    (metric_scale = 'numeric' AND numeric_value IS NOT NULL AND ordinal_value IS NULL AND ordinal_scale_key IS NULL AND ordinal_scale_version IS NULL AND categorical_value IS NULL) OR
    (metric_scale = 'ordinal' AND numeric_value IS NULL AND ordinal_value IS NOT NULL AND ordinal_scale_key IS NOT NULL AND ordinal_scale_version IS NOT NULL AND categorical_value IS NULL) OR
    (metric_scale = 'categorical' AND numeric_value IS NULL AND ordinal_value IS NULL AND ordinal_scale_key IS NULL AND ordinal_scale_version IS NULL AND categorical_value IS NOT NULL)
  )
);

CREATE TABLE evaluation_run_metric (
  id TEXT PRIMARY KEY,
  evaluation_run_id TEXT NOT NULL REFERENCES evaluation_run(id),
  metric_name TEXT NOT NULL,
  metric_scale TEXT NOT NULL CHECK (metric_scale IN ('numeric', 'ordinal', 'categorical')),
  numeric_value REAL,
  ordinal_value TEXT,
  ordinal_scale_key TEXT,
  ordinal_scale_version TEXT,
  categorical_value TEXT,
  comparator TEXT, -- baseline id/name
  created_at TEXT NOT NULL,
  CHECK (
    (metric_scale = 'numeric' AND numeric_value IS NOT NULL AND ordinal_value IS NULL AND ordinal_scale_key IS NULL AND ordinal_scale_version IS NULL AND categorical_value IS NULL) OR
    (metric_scale = 'ordinal' AND numeric_value IS NULL AND ordinal_value IS NOT NULL AND ordinal_scale_key IS NOT NULL AND ordinal_scale_version IS NOT NULL AND categorical_value IS NULL) OR
    (metric_scale = 'categorical' AND numeric_value IS NULL AND ordinal_value IS NULL AND ordinal_scale_key IS NULL AND ordinal_scale_version IS NULL AND categorical_value IS NOT NULL)
  )
);

-- Success invariants/compression
CREATE TABLE success_invariant_revision (
  id TEXT PRIMARY KEY,
  problem_id TEXT NOT NULL REFERENCES problem(id),
  evaluation_run_id TEXT NOT NULL REFERENCES evaluation_run(id),
  run_id TEXT NOT NULL REFERENCES run(id),
  parent_success_invariant_revision_id TEXT REFERENCES success_invariant_revision(id),
  created_at TEXT NOT NULL
);

CREATE TABLE success_invariant (
  id TEXT PRIMARY KEY,
  success_invariant_revision_id TEXT NOT NULL REFERENCES success_invariant_revision(id),
  statement TEXT NOT NULL,
  boundary_crossing TEXT NOT NULL,
  confidence_ordinal TEXT
);

CREATE TABLE success_invariant_boundary (
  success_invariant_id TEXT NOT NULL REFERENCES success_invariant(id),
  boundary_id TEXT NOT NULL REFERENCES failure_boundary(id),
  relation TEXT NOT NULL, -- crossed|depends_on
  PRIMARY KEY(success_invariant_id, boundary_id, relation)
);

CREATE TABLE success_invariant_failure_invariant (
  success_invariant_id TEXT NOT NULL REFERENCES success_invariant(id),
  candidate_invariant_id TEXT NOT NULL REFERENCES candidate_invariant(id),
  relation TEXT NOT NULL, -- breaks|refines|coexists_with
  PRIMARY KEY(success_invariant_id, candidate_invariant_id, relation)
);
```

## Implemented clustering + failure-space schema (#11)

The `## Core schema` blueprint above (`cluster_revision`, `mechanism_cluster`,
`cluster_membership`) is the forward-looking design. The tables that actually
ship for issue #11 are the migration `v7`/`v8` schema below, hardened by `v10`
(see `internal/store/migrations.go` and
[`mechanism-clustering.md`](mechanism-clustering.md)). They are all immutable by
trigger.

- `cluster_runs` — one deterministic clustering pass. Unique on
  `(problem_id, schema_version, vocabulary_version, profile_version,
  cluster_algo_version, thresholds_hash, input_set_hash)`, so a re-run under the
  identical version tuple **and identical signature population** is idempotent,
  a run under a different decisive set (profile) is a distinct row, and a re-run
  after the population changes (a newly ingested/signed mechanism) is a **new
  run** rather than a stale replay. `input_set_hash` is a `sha256` over the
  sorted `signature_id | fingerprint` of the clustered signatures (added in
  `v10`, KTD-1). Carries `signature_count`, `family_count`, and
  `status IN ('clean','degraded')`.
- `mechanism_clusters` — one family per row, with `cluster_fingerprint`,
  `representative_signature_id`, `member_count`, `isolate`, a summarized
  `intra_variation` (now including a `mechanism_distinct` pair count), and the
  per-family `outcome_class` + `outcome_mixed` flag (added in `v10`, KTD-9): a
  family spanning more than one member outcome is `mixed`, never compressed to
  the representative's outcome. Unique on `(cluster_run_id, cluster_fingerprint)`
  — the fingerprint now folds in member signature IDs so ambiguity-separated
  singletons cannot collide.
- `cluster_members` — signatures assigned to a family, with a `redundant` flag.
- `cluster_distances` — representative-vs-representative comparison verdicts.
- `cluster_coverage_axes` — per-axis distinct-value counts + `under_sampled`.
- `cluster_discrimination_losses` — recorded abstraction-loss pairs when the run
  is `degraded`.
- `failure_spaces` — a materialized failure-space revision. Unique on
  `(problem_id, cluster_run_id)` and `(problem_id, revision)`.
- `failure_space_outcomes` — family count per family-outcome class, including the
  aggregate `mixed` class for heterogeneous families.
- `failure_space_axes` — coverage axes inherited from the cluster run.

Migration `v10` is introspective and idempotent: it adds `input_set_hash` to
`cluster_runs` (rebuilding the table to widen the idempotency UNIQUE, since
SQLite cannot alter a UNIQUE in place) and `outcome_class`/`outcome_mixed` to
`mechanism_clusters`. A fresh database already at `v10` finds the target shape
present and does nothing. The post-migration integrity check now asserts these
columns exist, so an incomplete upgrade (tables present but the `v10` columns
absent) is reported as a corrupt store rather than silently accepted.

## Immutable vs mutable/revisioned

- Immutable: `source`, `evidence_record` (enforced with update/delete-rejecting
  triggers; `CHECK` constraints on enum-like fields are separate validity rules).
- Revisioned append-only views: normalization, clustering, invariant mining, success compression.
- Mutable-by-transition (not overwrite): invariant state via appended `invariant_state_transition` (+ `invariant_current_state` view) and optional lineage rows.

## Re-normalization / re-clustering semantics

- Never rewrite old derived rows.
- Create new revision row with the relevant parent pointer (`parent_revision_id`, `parent_cluster_revision_id`, `parent_invariant_revision_id`, `parent_success_invariant_revision_id`).
- All downstream commands must declare which upstream revision they read (explicit or “latest” resolution logged in `run_event`).
- Enables A/B comparisons across revisions and providers.

## Provider reproducibility

`provider_call` stores role-specific metadata (provider/model/schema/prompt hashes), allowing reproduction and audit per artifact-producing run.

## Holdout leakage and lifecycle enforcement

- `normalization_revision` persists both the declared filter and the exact
  evidence snapshot used to build the training split.
- `holdout_leakage_check` records the pass/fail result for a
  `holdout_set`/`normalization_revision` pair; generation and holdout evaluation
  reference that row rather than relying on narrative notes.
- Leakage checks are persisted as finalized results only: passed checks write the
  parent row alone with `overlap_count = 0`, while failed checks write the
  parent row plus immutable overlap-detail rows in the same transaction.
- `holdout_leakage_check_overlap` stores the concrete overlapping source/evidence
  rows when a leakage check fails, so audits can show exactly what violated the
  split.
- `invariant_state_transition` inserts are validated by trigger, and repositories
  allocate `transition_seq` from `invariant_transition_counter` by atomically
  incrementing `last_transition_seq` with `UPDATE ... RETURNING` before inserting
  the transition row, so competing writers cannot derive the same sequence
  number or rely on `MAX(...)+1`.

## Experiment comparison

- Compare runs by `problem_id`, revision lineage, provider role/model, and baseline type.
- Store proposal-level metrics in `evaluation_metric`, run aggregates in
  `evaluation_run_metric`, rather than opaque JSON.
- Open-ended model raw payloads can be retained in optional side tables (`provider_payload_json`) only for forensic replay, never as primary query surface.
