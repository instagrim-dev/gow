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
  "artifacts": [{"type": "normalization_revision", "id": "normrev_..."}],
  "warnings": [],
  "errors": []
}
```

## Command contracts

### `newf init <problem>`

- Purpose: create problem, workspace metadata, and first run shell.
- Required input: problem statement string (or `--problem-file`).
- Optional flags: `--slug`, `--tags`, `--description`, `--assumption-set`, `--experiment-name`.
- Persists:
  - `problem`
  - `experiment` (created when `--experiment-name` is provided)
  - `run` (state `initialized`)
- Idempotency: same explicit `--slug` returns existing problem unless `--new-problem`.
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
- Failure semantics: unreadable source, unsupported scheme, parse failure (partial ingest allowed; failures become artifacts).
- Human output: ingested count, duplicate count, failed inputs with reasons.
- Depends on: `init`.
- Downstream: `normalize`.

### `newf normalize`

- Purpose: map evidence-derived approaches into typed mechanisms/outcomes; separate interpretation from source.
- Required input: existing ingested evidence for problem.
- Optional flags: `--source-filter`, `--provider`, `--revision-note`, `--from-cutoff`.
- Persists:
  - `normalization_revision`
  - `approach`, `mechanism`, `outcome`, `failure_boundary` (all linked to revision)
  - extraction provider call metadata
- Idempotency: deterministic mode can reuse existing revision fingerprint; otherwise creates new revision.
- Failure semantics: provider/schema validation failures stored as failed normalization items.
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
- Optional flags: `--mode known-counterexample|synthetic-counterexample|success-preserving|split|merge|bias-critique|all`, `--provider`, `--budget`.
- Persists:
  - `invariant_challenge`
  - state transition in challenge result (`result_state`) + optional `invariant_lineage`
  - optional child/sibling invariants for split/merge
- Idempotency: always appends new challenge record; invariant state reflects latest accepted transition.
- Failure semantics: unknown invariant, invalid transition, provider/evidence fetch failure.
- Human output: challenge result, old/new state, cited evidence IDs.
- Depends on: `invariants`.
- Downstream: `generate`, `compress`, `evaluate`.

### `newf generate --against <invariant-id> --count <n>`

- Purpose: produce frontier proposals explicitly violating surviving failure invariants.
- Required input: target invariant ID(s), count.
- Optional flags: `--cluster-rev`, `--provider`, `--max-cost`, `--novelty-threshold`.
- Persists:
  - `frontier_generation_run`
  - `frontier_proposal`
  - targeted invariant links + nearest cluster links
- Idempotency: new run each invocation; dedupe identical structural proposal hash within problem.
- Failure semantics: non-surviving invariant target, generation schema violation.
- Human output: ranked proposals with component scores (ordinal/component, not fake scalar precision).
- Depends on: `challenge`/`invariants` output.
- Downstream: `evaluate`, `compress`.

### `newf evaluate`

- Purpose: evaluate frontier proposals and/or run historical holdout experiment.
- Required input: proposals or `--experiment-id`.
- Optional flags: `--mode proposal|holdout`, `--baseline undirected|semantic-summary`, `--cutoff`, `--holdout-set-id`, `--judge-provider`.
- Persists:
  - `evaluation_run`
  - `evaluation`
  - `evaluation_metric`
  - `holdout_set` (when creating from flags/sources for holdout mode)
  - baseline comparison rows
- Idempotency: new evaluation run per invocation.
- Failure semantics: missing holdout partition, unevaluable proposals, judge disagreement (recorded as unresolved).
- Human output: metric summary + per-proposal/per-experiment outcomes.
- Depends on: `generate` (proposal mode) or `ingest..generate` pipeline artifacts (holdout mode).
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
proposed -> challenged -> surviving | weakened | split | merged | falsified | established
```

- `established` requires external independent evidence (e.g., theorem/proof-check or independently replicated experiment), never model consensus alone.
- split/merge produce lineage edges so history remains queryable.

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
