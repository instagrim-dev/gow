# SQLite persistence design (`newf` v0)

## Why SQLite

- Single-file reproducible experiments.
- Strong relational constraints for provenance-heavy graph.
- Transactional updates for command runs.
- Sufficient for v0 local CLI scale.
- Every opened SQLite connection enables `PRAGMA foreign_keys=ON`, and the
  migration/bootstrap path includes a failing orphan-insert self-test so
  relational guarantees do not silently disappear.

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
  leakage_check_status TEXT NOT NULL CHECK (leakage_check_status IN ('pending', 'passed', 'failed')),
  created_at TEXT NOT NULL
);

CREATE TABLE holdout_set_source (
  holdout_set_id TEXT NOT NULL REFERENCES holdout_set(id),
  source_id TEXT NOT NULL REFERENCES source(id),
  PRIMARY KEY(holdout_set_id, source_id)
);

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
  initial_state TEXT NOT NULL CHECK (initial_state IN ('proposed', 'challenged', 'surviving', 'weakened', 'split', 'merged', 'falsified', 'established')), -- normally proposed
  confidence_ordinal TEXT
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
  relation TEXT NOT NULL CHECK (relation IN ('split', 'merged', 'weakened')),
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
  from_state TEXT NOT NULL CHECK (from_state IN ('proposed', 'challenged', 'surviving', 'weakened', 'split', 'merged', 'falsified', 'established')),
  to_state TEXT NOT NULL CHECK (to_state IN ('proposed', 'challenged', 'surviving', 'weakened', 'split', 'merged', 'falsified', 'established')),
  created_at TEXT NOT NULL,
  UNIQUE(invariant_id, transition_seq)
);

CREATE TRIGGER invariant_state_transition_validate_insert
BEFORE INSERT ON invariant_state_transition
BEGIN
  SELECT CASE
    WHEN NEW.transition_seq <> COALESCE((
      SELECT MAX(t.transition_seq) + 1
      FROM invariant_state_transition t
      WHERE t.invariant_id = NEW.invariant_id
    ), 1) THEN RAISE(ABORT, 'transition_seq must append exactly once per invariant')
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
      (NEW.from_state = 'challenged' AND NEW.to_state IN ('surviving', 'weakened', 'split', 'merged', 'falsified')) OR
      (NEW.from_state = 'surviving' AND NEW.to_state IN ('challenged', 'weakened', 'split', 'merged', 'falsified', 'established')) OR
      (NEW.from_state = 'weakened' AND NEW.to_state IN ('challenged', 'surviving', 'split', 'merged', 'falsified'))
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
  status TEXT NOT NULL CHECK (status IN ('passed', 'failed')),
  checked_at TEXT NOT NULL,
  UNIQUE(holdout_set_id, normalization_revision_id)
);

CREATE TABLE holdout_leakage_check_overlap (
  holdout_leakage_check_id TEXT NOT NULL REFERENCES holdout_leakage_check(id),
  overlapping_source_id TEXT REFERENCES source(id),
  overlapping_evidence_id TEXT REFERENCES evidence_record(id),
  overlap_kind TEXT NOT NULL CHECK (overlap_kind IN ('source', 'evidence')),
  CHECK (
    (overlap_kind = 'source' AND overlapping_source_id IS NOT NULL AND overlapping_evidence_id IS NULL) OR
    (overlap_kind = 'evidence' AND overlapping_source_id IS NULL AND overlapping_evidence_id IS NOT NULL)
  ),
  PRIMARY KEY(holdout_leakage_check_id, overlap_kind, overlapping_source_id, overlapping_evidence_id)
);

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
  mode TEXT NOT NULL, -- proposal|holdout
  cutoff_time TEXT,
  baseline_type TEXT,
  proposal_budget_count INTEGER,
  evaluation_budget_count INTEGER,
  judge_provider_name TEXT,
  judge_provider_version TEXT,
  judge_model_name TEXT,
  judge_config_hash TEXT,
  created_at TEXT NOT NULL
);

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
  holdout_source_id TEXT REFERENCES source(id),
  holdout_family_label TEXT,
  match_kind TEXT NOT NULL CHECK (match_kind IN ('source_recovery', 'family_recovery', 'structural_break')),
  match_verdict TEXT NOT NULL CHECK (match_verdict IN ('exact', 'equivalent', 'miss')),
  notes TEXT,
  CHECK (
    (match_kind = 'source_recovery' AND holdout_source_id IS NOT NULL AND holdout_family_label IS NULL) OR
    (match_kind = 'family_recovery' AND holdout_source_id IS NULL AND holdout_family_label IS NOT NULL) OR
    (match_kind = 'structural_break' AND holdout_source_id IS NULL AND holdout_family_label IS NOT NULL)
  )
);

CREATE TABLE evaluation_metric (
  id TEXT PRIMARY KEY,
  evaluation_run_id TEXT NOT NULL REFERENCES evaluation_run(id),
  metric_name TEXT NOT NULL,
  metric_scale TEXT NOT NULL CHECK (metric_scale IN ('numeric', 'ordinal', 'categorical')),
  numeric_value REAL,
  ordinal_value TEXT,
  categorical_value TEXT,
  comparator TEXT, -- baseline id/name
  created_at TEXT NOT NULL,
  CHECK (
    (metric_scale = 'numeric' AND numeric_value IS NOT NULL AND ordinal_value IS NULL AND categorical_value IS NULL) OR
    (metric_scale = 'ordinal' AND numeric_value IS NULL AND ordinal_value IS NOT NULL AND categorical_value IS NULL) OR
    (metric_scale = 'categorical' AND numeric_value IS NULL AND ordinal_value IS NULL AND categorical_value IS NOT NULL)
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
- `holdout_leakage_check_overlap` stores the concrete overlapping source/evidence
  rows when a leakage check fails, so audits can show exactly what violated the
  split.
- `invariant_state_transition` inserts are validated by trigger, and repositories
  allocate `transition_seq` with `INSERT ... SELECT COALESCE(MAX(...)+1, 1)` in
  the same `BEGIN IMMEDIATE` transaction so competing writers cannot create
  gaps, replays, or impossible transitions.

## Experiment comparison

- Compare runs by `problem_id`, revision lineage, provider role/model, and baseline type.
- Store metrics in normalized `evaluation_metric` rows rather than opaque JSON.
- Open-ended model raw payloads can be retained in optional side tables (`provider_payload_json`) only for forensic replay, never as primary query surface.
