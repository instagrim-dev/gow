# `newf` v0 CLI Design

## Scope

v0 exposes the research stages directly (not a hidden monolithic loop) so users and agents can inspect and rerun each artifact-producing step:

```text
newf init <problem>
newf ingest <sources...>
newf normalize
newf cluster
newf invariants
newf challenge <invariant-id>
newf generate --against <invariant-id> --count <n>
newf evaluate
newf compress
```

## Global CLI contracts

- `--db <path>`: SQLite path (default `.newf/newf.db`).
- each command invocation creates a new `run` provenance record.
- `--experiment-id <id>`: optional link to an existing experiment for cross-command grouping.
- `--json`: stable machine-readable output.
- `--format table|json|yaml` (human defaults to table; `--json` wins).
- `--quiet`: suppress narrative text, keep IDs/summaries.
- `--no-write`: dry run where supported.

Common machine envelope:

```json
{
  "ok": true,
  "command": "normalize",
  "problem_id": "prob_...",
  "run_id": "run_...",
  "artifacts": [{"type": "normalization_revision", "id": "normrev_...", "status": "created"}],
  "warnings": [],
  "errors": []
}
```

Duplicate/skip cases are reported in `artifacts` or `warnings` with stable statuses such as `duplicate_existing` and include the existing artifact ID.

## Command contracts

### `newf init <problem>`

- Purpose: create problem, workspace metadata, and first run shell.
- Required input: problem statement string (or `--problem-file`).
- Optional flags: `--slug`, `--tags`, `--description`, `--assumption-set`, `--experiment-name`.
- Persists:
  - `problem`
  - `problem_tag` rows for `--tags`
  - `experiment` (created when `--experiment-name` is provided)
  - `experiment_assumption` rows for `--assumption-set` values attached to the created/selected experiment
  - `run` (state `initialized`)
- Idempotency: same explicit `--slug` returns existing problem unless `--new-problem`; experiment creation reuses the same `(problem_id, experiment_name)` shell and merges duplicate tag/assumption values idempotently.
- Failure semantics: invalid slug, DB unavailable, uniqueness conflict.
- Human output: created IDs, db path, next-command hints.
- Depends on: none.
- Downstream dependencies: all later commands require `problem_id`.

### `newf ingest <sources...>`

- Purpose: register immutable source metadata and immutable evidence units.
- Required input: one or more source locators (file path, URL, DOI/ref key).
- Optional flags: `--kind literature|human|experiment`, `--title`, `--authors`, `--published-at`, `--cutoff-tag`, `--allow-duplicates`.
- Persists:
  - `source`
  - `evidence_record` (immutable raw claims/snippets/results)
  - `run_event` provenance
- Notes: synthetic/generated attempts are persisted as `synthetic_artifact` via challenge/generation/evaluation flows, not as source evidence.
- Idempotency: content hash + canonical source ref dedupe by default.
- Failure semantics: unreadable source, unsupported scheme, parse failure, or generated/synthetic material sent to ingest (must use synthetic-artifact flow).
- Human output: ingested count, duplicate count, failed inputs with reasons.
- Depends on: `init`.
- Downstream: `normalize`.

### `newf normalize`

- Purpose: map evidence-derived approaches into typed mechanisms/outcomes; separate interpretation from source.
- Required input: existing ingested evidence for problem.
- Optional flags: `--source-filter`, `--exclude-holdout-set-id`, `--provider`, `--revision-note`.
- Persists:
  - `normalization_revision`
  - `normalization_revision_evidence` snapshot rows + manifest hash for the exact training set
  - `normalization_item` rows for per-evidence success/failure/skipped outcomes
  - `approach`, `mechanism`, `outcome`, `failure_boundary` (all linked to revision)
  - extraction provider call metadata
- Idempotency: deterministic mode can reuse an existing `(source_filter_json, excluded_holdout_set_id, source_manifest_hash, config_hash)` fingerprint; otherwise creates new revision.
- Failure semantics: provider/schema validation failures are stored as failed `normalization_item` rows linked to the specific evidence record.
- Human output: revision ID, created/failed counts, axis coverage summary.
- Depends on: `ingest`.
- Downstream: `cluster`, `invariants`, `evaluate`.

### `newf cluster`

- Purpose: create mechanistic (not surface) clusters from normalized mechanisms.
- Required input: normalization revision.
- Optional flags: `--normalization-rev`, `--provider`, `--k`, `--strategy`.
- Persists:
  - `cluster_revision`
  - `mechanism_cluster`
  - `cluster_membership`
  - cluster rationale/evidence refs
- Idempotency: same input set + config hash can be reused; otherwise new revision.
- Failure semantics: insufficient normalized mechanisms, provider failure.
- Human output: cluster IDs, sizes, mechanistic-axis dispersion.
- Depends on: `normalize`.
- Downstream: `invariants`, `generate`, `evaluate`.

### `newf invariants`

- Purpose: infer candidate failure invariants from independent failure clusters.
- Required input: cluster revision with failure outcomes.
- Optional flags: `--cluster-rev`, `--provider`, `--min-support-clusters`, `--max-candidates`.
- Persists:
  - `invariant_revision`
  - `candidate_invariant` (state `proposed`)
  - supporting evidence and support-cluster links
- Idempotency: same cluster revision + config hash can reuse; otherwise new revision.
- Failure semantics: no qualifying failure support set, schema violations.
- Human output: invariant table (`id`, statement, abstraction level, support, confidence).
- Depends on: `cluster`.
- Downstream: `challenge`, `generate`, `evaluate`, `compress`.

### `newf challenge <invariant-id>`

- Purpose: adversarially test one candidate invariant and transition lifecycle state.
- Required input: invariant ID.
- Optional flags: `--mode known-counterexample|synthetic-counterexample|success-preserving|split|merge|bias-critique|all`, `--parent-invariant-id <id>` (repeatable for `--mode merge`), `--provider`, `--budget`.
- Persists:
  - `invariant_challenge`
  - `invariant_state_transition` (with current state exposed through derived `invariant_current_state` view)
  - optional `invariant_lineage`
  - optional child/sibling invariants for split/merge lineage cases
- Idempotency: always appends new challenge record; invariant state reflects latest accepted transition.
- Failure semantics: unknown invariant, invalid transition, provider/evidence fetch failure.
- Human output: challenge result, old/new state, cited evidence IDs.
- Depends on: `invariants`.
- Downstream: `generate`, `compress`, `evaluate`.

### `newf generate --against <invariant-id> --count <n>`

- Purpose: produce frontier proposals explicitly violating surviving failure invariants.
- Required input: target invariant ID, count.
- Optional flags: `--cluster-rev`, `--holdout-set-id`, `--provider`, `--max-cost`, `--novelty-threshold`.
- Persists:
  - `frontier_generation_run`
  - `frontier_proposal`
  - targeted invariant links + nearest cluster links
- Idempotency: new run each invocation; duplicate proposal hashes are deduped globally per problem (reported as `duplicate_existing` with existing `proposal_id`), while new hashes create new proposals.
- Failure semantics: unknown invariant ID, target invariant not in `surviving`, generation schema violation, or failed/missing holdout leakage check when `--holdout-set-id` is supplied.
- Human output: ranked proposals with component scores (ordinal/component, not fake scalar precision).
- Depends on: `challenge`/`invariants` output.
- Downstream: `evaluate`, `compress`.

### `newf evaluate`

- Purpose: evaluate frontier proposals and/or run historical holdout experiment.
- Required input:
  - proposal mode: `--proposal-id <prop-id>` (repeatable) or `--frontier-generation-run-id <run-id>`
  - holdout mode: either `--holdout-set-id <set-id>` or raw holdout definition inputs (`--cutoff`, `--holdout-source-id`, `--holdout-family-label`), plus a holdout-filtered upstream chain (`--normalization-rev`, `--cluster-rev`, `--invariant-rev`, `--frontier-generation-run-id`) derived from the same excluded holdout set
- Optional flags:
  - shared: `--mode proposal|holdout`, `--baseline undirected|semantic-summary`, `--judge-provider`
  - holdout-set creation (required only when `--holdout-set-id` is omitted in holdout mode): `--cutoff`, `--holdout-source-id <src-id>` (repeatable), `--holdout-family-label <label>` (repeatable)
  - upstream revision selectors (required in holdout mode): `--normalization-rev`, `--cluster-rev`, `--invariant-rev`, `--frontier-generation-run-id`
- Persists:
  - `evaluation_run`
  - `evaluation`
  - `evaluation_holdout_match` rows when scoring holdout-family/source recovery
  - `evaluation_metric`
  - `holdout_set` (when creating from `--holdout-source-id`/`--holdout-family-label` inputs for holdout mode)
  - baseline comparison rows
- Idempotency: new evaluation run per invocation.
- Failure semantics: missing holdout partition, unevaluable proposals, judge disagreement (recorded as unresolved).
- Human output: metric summary + per-proposal/per-experiment outcomes, selected proposal/generation scope, and leakage-check status.
- Depends on:
  - proposal mode: `generate`
  - holdout mode: `holdout_set` + `normalize --exclude-holdout-set-id` + `cluster` + `invariants` + `challenge` + `generate --holdout-set-id`, all referencing the same upstream filtered revision chain and a passed leakage check
- Downstream: `compress`.

### `newf compress`

- Purpose: derive candidate success invariants from partial-success/success outcomes and update generalized frontier guidance.
- Required input: evaluated outcomes including partial successes.
- Optional flags: `--provider`, `--evaluation-run`, `--min-support`.
- Persists:
  - `success_invariant_revision`
  - `success_invariant`
  - links to boundary crossings and prior failure invariants
- Idempotency: revisioned output; config hash can reuse.
- Failure semantics: insufficient success evidence; contradiction across outcomes.
- Human output: success invariants + boundary-crossing summary and next frontier hints.
- Depends on: `evaluate`.
- Downstream: next research iteration.

## Invariant lifecycle state machine

```text
proposed -> challenged -> surviving | weaken | split | merge | falsified
surviving -> challenged | weaken | split | merge | falsified | established
```

- `established` requires external independent evidence (e.g., theorem/proof-check or independently replicated experiment), never model consensus alone.
- split/merge cases produce lineage edges so history remains queryable.

## Human + machine UX expectations

- IDs are stable and copy/paste ready (`inv_...`, `cluster_...`, `prop_...`).
- Long operations persist intermediate states in DB and expose status via `run_event`.
- Informative failures are retained as artifacts and shown in summaries.
- Confidence is shown with provenance (`provider`, `prompt/schema version`, `evidence IDs`), not prose-heavy explanations.

## Suggested package boundaries

```text
cmd/newf/                 # Cobra wiring only
internal/domain/          # entities, lifecycle/state rules, scoring components
internal/store/           # SQLite schema + repositories + migrations
internal/provider/        # role interfaces + adapters
internal/pipeline/        # command orchestration/application services
internal/normalize/       # extraction + mechanism axis mapping
internal/cluster/         # mechanistic clustering logic
internal/invariant/       # mining + challenge + lifecycle transitions
internal/frontier/        # generation + novelty/ranking components
internal/eval/            # holdout workflow + metrics + baselines
```

Rule: add package only where a real dependency boundary exists.
