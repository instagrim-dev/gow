# SQLite persistence design (`newf` v0)

## Why SQLite

- Single-file reproducible experiments.
- Strong relational constraints for provenance-heavy graph.
- Transactional updates for command runs.
- Sufficient for v0 local CLI scale.

## ID strategy

- Stable textual IDs with prefixes (e.g., `prob_`, `src_`, `evd_`, `inv_`, `prop_`).
- ULID/UUIDv7 recommended for time-sort + uniqueness.
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
  created_at TEXT NOT NULL
);

CREATE TABLE run (
  id TEXT PRIMARY KEY,
  problem_id TEXT NOT NULL REFERENCES problem(id),
  experiment_id TEXT REFERENCES experiment(id),
  command TEXT NOT NULL,
  status TEXT NOT NULL, -- initialized|running|completed|failed
  config_hash TEXT,
  started_at TEXT NOT NULL,
  completed_at TEXT
);

CREATE TABLE provider_call (
  id TEXT PRIMARY KEY,
  run_id TEXT NOT NULL REFERENCES run(id),
  role TEXT NOT NULL, -- normalize|cluster|mine|critic|generate|judge|compress
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
  kind TEXT NOT NULL, -- literature|human|experiment
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
  strength TEXT,
  created_at TEXT NOT NULL
);

-- Revisioned normalization artifacts
CREATE TABLE normalization_revision (
  id TEXT PRIMARY KEY,
  problem_id TEXT NOT NULL REFERENCES problem(id),
  run_id TEXT NOT NULL REFERENCES run(id),
  parent_revision_id TEXT REFERENCES normalization_revision(id),
  config_hash TEXT NOT NULL,
  status TEXT NOT NULL,
  created_at TEXT NOT NULL
);

CREATE TABLE approach (
  id TEXT PRIMARY KEY,
  problem_id TEXT NOT NULL REFERENCES problem(id),
  normalization_revision_id TEXT NOT NULL REFERENCES normalization_revision(id),
  source_id TEXT NOT NULL REFERENCES source(id),
  surface_summary TEXT
);

CREATE TABLE mechanism (
  id TEXT PRIMARY KEY,
  approach_id TEXT NOT NULL REFERENCES approach(id),
  notes TEXT,
  UNIQUE(approach_id)
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
  PRIMARY KEY(mechanism_id, axis_key)
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
  label TEXT NOT NULL,
  UNIQUE(problem_id, label)
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

CREATE TABLE mechanism_cluster (
  id TEXT PRIMARY KEY,
  cluster_revision_id TEXT NOT NULL REFERENCES cluster_revision(id),
  label TEXT,
  rationale TEXT
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
  state TEXT NOT NULL,
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
  relation TEXT NOT NULL, -- split|merge|weaken
  PRIMARY KEY(parent_invariant_id, child_invariant_id, relation)
);

CREATE TABLE invariant_challenge (
  id TEXT PRIMARY KEY,
  invariant_id TEXT NOT NULL REFERENCES candidate_invariant(id),
  run_id TEXT NOT NULL REFERENCES run(id),
  challenge_type TEXT NOT NULL,
  result_state TEXT NOT NULL,
  result_summary TEXT,
  created_at TEXT NOT NULL
);

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

CREATE TABLE holdout_set (
  id TEXT PRIMARY KEY,
  problem_id TEXT NOT NULL REFERENCES problem(id),
  name TEXT NOT NULL,
  cutoff_time TEXT NOT NULL,
  held_out_family_label TEXT,
  leakage_check_status TEXT NOT NULL,
  created_at TEXT NOT NULL
);

CREATE TABLE holdout_set_source (
  holdout_set_id TEXT NOT NULL REFERENCES holdout_set(id),
  source_id TEXT NOT NULL REFERENCES source(id),
  PRIMARY KEY(holdout_set_id, source_id)
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
  mode TEXT NOT NULL, -- proposal|holdout
  cutoff_time TEXT,
  baseline_type TEXT,
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

CREATE TABLE evaluation_metric (
  id TEXT PRIMARY KEY,
  evaluation_run_id TEXT NOT NULL REFERENCES evaluation_run(id),
  metric_name TEXT NOT NULL,
  metric_value TEXT NOT NULL,
  metric_scale TEXT NOT NULL, -- numeric|ordinal|categorical
  comparator TEXT, -- baseline id/name
  created_at TEXT NOT NULL
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

- Immutable: `source`, `evidence_record`.
- Revisioned append-only views: normalization, clustering, invariant mining, success compression.
- Mutable-by-transition (not overwrite): invariant state via appended `invariant_challenge` + optional lineage rows.

## Re-normalization / re-clustering semantics

- Never rewrite old derived rows.
- Create new revision row with the relevant parent pointer (`parent_revision_id`, `parent_cluster_revision_id`, `parent_invariant_revision_id`, `parent_success_invariant_revision_id`).
- All downstream commands must declare which upstream revision they read (explicit or “latest” resolution logged in `run_event`).
- Enables A/B comparisons across revisions and providers.

## Provider reproducibility

`provider_call` stores role-specific metadata (provider/model/schema/prompt hashes), allowing reproduction and audit per artifact-producing run.

## Experiment comparison

- Compare runs by `problem_id`, revision lineage, provider role/model, and baseline type.
- Store metrics in normalized `evaluation_metric` rows rather than opaque JSON.
- Open-ended model raw payloads can be retained in optional side tables (`provider_payload_json`) only for forensic replay, never as primary query surface.
