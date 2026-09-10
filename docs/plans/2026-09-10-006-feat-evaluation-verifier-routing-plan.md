---
artifact_contract: ce-unified-plan/v1
artifact_readiness: implementation-ready
execution: code
product_contract_source: ce-plan-bootstrap
origin: GitHub issue #15 (instagrim-dev/newf) — "Route frontier proposals to the strongest available verifier and record epistemic strength" (to be filed)
epic: EPIC.md milestone M5.2 — Evaluation and verifier routing
depends_on: >-
  GitHub issue #14 (M5.1 — frontier generation), planned in
  docs/plans/2026-09-10-005-feat-frontier-generation-plan.md, which lands the
  `frontier_generation_run` / `frontier_proposal` / `frontier_target_invariant` /
  `frontier_nearest_cluster` tables and leaves `frontier_proposal.result` NULL for
  THIS slice to populate. Upstream, in the SHIPPED tree: #13 (M4.3 challenge
  lifecycle, schema v13, `invariant_current_state`), #12/F2 (M4.2 candidate
  invariants, v11/v12), #11 (clustering + FailureSpace, v7/v8/v10), #9
  (canonicalization/comparison, v4/v5/v6). NOTE ON MIGRATION NUMBERING: the shipped
  tree is at `currentSchemaVersion = 13`; M5.1 frontier is expected to land at v14,
  so THIS slice lands at **v15**. Re-cement the exact number against
  `internal/store/migrations.go` at implementation time and take `+1` over the
  landed frontier migration.
created: 2026-09-10
plan_type: feat
---

# feat: Route frontier proposals to the strongest available verifier and record epistemic strength

## Summary

Implement the next epic slice (EPIC.md **M5.2 — Evaluation and verifier
routing**): take the typed **`FrontierProposal`** records M5.1 produces (whose
`result` column is deliberately left NULL) and **route each proposal to the
strongest verifier actually available for it**, recording a durable,
evidence-bearing **`Evaluation`** whose **verification strength is explicit** and
whose failed proposals can **re-enter the failure atlas**.

This is the toolbox `evaluate` operator (`docs/toolbox-dsl.md`), run
**cheap-first**:

```text
results := evaluate frontier {
  proof_check
  symbolic_check
  computation
  counterexample_search
}
```

```text
evaluate  Proposal -> Outcome
> evaluate F-21 --cheap-first
partial_success:
  eliminates survivor family S4
  fails on S7
```

The load-bearing idea is the **verification hierarchy** (`AGENTS.md`, EPIC.md
M5.2). An outcome is worth exactly as much as the mechanism that produced it, and
that strength must be **stored, not implied**:

```text
formal proof / deterministic check
> reproducible computation or experiment
> independently sourced evidence
> independent critic / cross-model agreement
> single-model judgment
```

> A model may propose a verification procedure. It should not silently certify
> its own output merely because it sounds convincing. (`AGENTS.md`)
> `ModelJudgment != Verification`.

The exit condition (EPIC.md M5.2): **evaluations are durable evidence-bearing
artifacts and failed proposals can re-enter the atlas.** An "evaluation" that
records a verdict without recording *which tier of verifier produced it* is not
an evaluation in this system — it is a model opinion wearing a lab coat, and this
slice must make that distinction structural.

This slice **consumes** the M5.1 frontier substrate (only real
`frontier_proposal` rows may be evaluated; it populates their `result`), the
`run` lifecycle (`running` → `completed`/`failed`), and the existing
`provider_invocations` provenance table (role `'evaluate'`). It adds a
**`Verifier` routing layer** (deterministic verifiers first, a fixture
model-judge last), the evaluation persistence layer (transcribed from
`docs/persistence.md`'s `evaluation_run` / `evaluation` / `evaluation_metric`
blueprint, **plus an explicit `verifier_kind` + `verification_strength`** the
blueprint does not yet carry), the **failure-atlas re-entry** wiring, pipeline
services, and CLI surface. It runs entirely offline in CI against project-authored
fixtures. It deliberately **defers** the holdout experiment (M7) and success
compression (M6): it persists the nullable holdout hooks the blueprint already
defines but does not implement leakage checking or success-invariant mining.

---

## Problem Frame

After M5.1, `newf` can turn a surviving failure invariant into a ranked set of
directed, falsifiable frontier proposals — but every proposal's `result` is NULL.
Nothing decides *whether a proposal actually did what it claimed*, and nothing
records *how strongly we know it*. Two failure modes are equally fatal to the
research loop:

1. **No evaluation at all** — proposals accumulate as untested speculation; the
   loop `frontier -> evaluate -> success_space -> compress` (`docs/toolbox-dsl.md`)
   has a hole where its evidence should be, and the M7 holdout experiment has no
   scored arm to compare.
2. **Evaluation that launders judgment as proof** — a single model says "this
   works," it is stored as `success`, and the system now believes a theorem
   because a language model was fluent. This is precisely the failure `AGENTS.md`
   forbids: *"It should not silently certify its own output merely because it
   sounds convincing."*

The value of this slice is therefore **not** producing verdicts. A bare model
produces verdicts. The value is producing verdicts that are (a) **routed to the
strongest verifier that can actually run for this proposal**, (b) **stamped with
the epistemic strength of that verifier**, and (c) **reversible into search
pressure** — a failed proposal is not discarded but returned to the failure
atlas as a newly-evaluated failure, exactly as `docs/toolbox-dsl.md` requires:

```text
An evaluated proposal does not disappear when it fails.
FailureSpace(t+1) = FailureSpace(t) + newly_evaluated_failures
```

Today the repository has the full upstream chain (sources → normalization →
signatures → clusters → FailureSpace → candidate invariants → challenge
lifecycle → [frontier proposals, once M5.1 lands]) but **no evaluation artifact**:
nothing routes a proposal to a verifier, nothing records verification strength,
and a failed proposal cannot feed the next clustering pass. Without it, M6
(success compression) has no `success_space` to compress and M7 (holdout) has no
evaluated outcomes to score.

### Governing constraints (from `AGENTS.md`, `docs/persistence.md`, `docs/toolbox-dsl.md`, EPIC.md M5.2)

- **The verification hierarchy is recorded, not implied.** Every `evaluation`
  stores a `verifier_kind` (which tier ran) and a `verification_strength`
  (ordinal position in the hierarchy). A `single-model-judgment` verdict and a
  `deterministic-check` verdict may share a `verdict` value but MUST NOT share a
  strength. This is the structural expression of `ModelJudgment != Verification`.
- **Prefer the strongest verifier that can actually run; never fabricate a
  stronger one.** Routing is cheap-first by cost but records the *strength* of the
  verifier that produced the stored verdict. A proposal a deterministic checker
  cannot evaluate falls back to a weaker tier, and the stored strength drops
  accordingly — it is never silently upgraded. Higher layers do not
  retroactively invalidate lower ones, but the stored strength stays explicit
  (`AGENTS.md`, "Verification hierarchy").
- **A model may propose a procedure; it may not certify its own output.** Where a
  deterministic or independently checkable verifier exists (arithmetic, symbolic,
  code execution, counterexample search), truth-sensitive verdicts route there;
  the model-judge tier is the *last* resort and is stamped `single-model-judgment`.
- **Failure is a first-class artifact.** A `failure`/`partial_failure` evaluation
  writes the proposal's result AND makes the proposal's mechanism eligible for
  re-entry into the failure atlas (a subsequent clustering pass sees it), so the
  loop can learn. `Failure != UselessOutput`.
- **No fake precision.** Per `docs/persistence.md`, evaluation metrics are stored
  in `evaluation_metric` / `evaluation_run_metric` with an explicit
  `metric_scale IN ('numeric','ordinal','categorical')`; confidence is an
  `confidence_ordinal`, not an invented float. Component-wise/ordinal judgments
  are acceptable in v0 (`AGENTS.md`).
- **No epistemic promotion of the invariant.** Evaluating a proposal that targets
  a surviving invariant does NOT change that invariant's state; success-invariant
  mining and boundary compression are M6. This slice records proposal outcomes
  only.
- **Holdout discipline is wired but deferred.** The `evaluation_run` blueprint
  carries a `mode IN ('proposal','holdout')` and holdout-gate triggers. This
  slice ships **`mode='proposal'` only**, persists the nullable holdout columns
  and the blueprint's holdout-gate triggers verbatim (so M7 does not retrofit),
  and refuses `mode='holdout'` at the service boundary with a "deferred to M7"
  error rather than a half-built path.
- **Provenance + replay.** Each evaluation run records the run lifecycle plus, for
  the model-judge tier only, provider/model/config + raw payloads
  (`provider_invocations`, role `'evaluate'`); deterministic verifiers record
  their tool identity/version instead of a provider. Re-running is append-only.
- **Additive, immutable persistence.** New tables in the migration **after M5.1's
  frontier migration**, with the repo's `RAISE(ABORT, …)` immutability triggers.
  No edits to prior migrations except the guarded `provider_invocations.role`
  widening to admit `'evaluate'` (same in-place `writable_schema` mechanism v11/v13
  used).
- **Deterministic fixtures / offline CI.** All verifiers used in tests are
  deterministic (the model-judge tier is a fixture); no network call in CI.

---

## Scope Boundaries

### In scope

- A **`Verifier` abstraction** (`internal/verify`, pure of SQL/Cobra) with a small
  tiered set:
  - `deterministic-check` — a pure, code-owned checker over the proposal's typed
    content and the target invariant's canonical predicate (e.g. does the proposed
    mechanism's signature actually violate the invariant's predicate axis, via
    `internal/invariant.Evaluate` — reusing the M4.2 evaluator as a real check),
    and any arithmetic/structural assertion expressible over canonical fields;
  - `counterexample-search` — a bounded deterministic search over persisted
    signatures / known families for a concrete counterexample to the proposal's
    structural-violation claim;
  - `model-judgment` — a `Verifier`-role provider (fixture in CI) that returns a
    verdict + `confidence_ordinal` + rationale, used **only** when no stronger
    tier can decide, and stamped `single-model-judgment`.
- A **cheap-first router** that, per proposal, tries verifiers in cost order,
  takes the verdict from the **strongest tier that returned a decisive result**
  (not `unknown`/`verification_blocked`), and records that tier's
  `verifier_kind` + `verification_strength`. `--cheap-first` is the default and
  only v0 policy; the router is deterministic and documented.
- **Evaluation persistence** (new migration): `evaluation_run`, `evaluation`
  (extended with `verifier_kind` + `verification_strength`), `evaluation_metric`,
  `evaluation_run_metric`, transcribed from `docs/persistence.md` 670–853 with the
  holdout-gate triggers verbatim and immutability triggers; **populate
  `frontier_proposal.result`** transactionally with the evaluation's verdict.
- **Failure-atlas re-entry**: a `failure`/`partial_failure` evaluation marks the
  proposal's mechanism eligible for the next `cluster build` (the mechanism enters
  the clustered population), so `FailureSpace(t+1)` grows. The concrete wiring is
  a persisted, queryable link (an `evaluated_failure` view or flag) that
  `cluster build` already-existing readers can include; no re-clustering is
  triggered automatically here.
- **Run lifecycle**: `evaluate` creates its run `running`, then
  `finalizeRun`→`completed` / `failRun`→`failed`.
- **Verdict vocabulary** exactly matching EPIC.md M5.2 expected outcomes:
  `failure | partial_failure | partial_success | success | unknown |
  verification_blocked`.
- **CLI** (all `--json`): `newf evaluate <proposal-id> [--cheap-first]`,
  `newf evaluate --problem <id> [--all]` (evaluate all un-evaluated proposals),
  `newf evaluation list --problem <id>`, `newf evaluation show <evaluation-id>`.
- **Deterministic fixtures + regressions + one offline end-to-end** test
  (`seed → mine → challenge→surviving → frontier generate → evaluate → show`).
- **Docs**: `docs/evaluation.md` update (or new `docs/verifier-routing.md`),
  EPIC.md **M5.2** status note, `docs/persistence.md` v-note for the shipped
  evaluation subset + the `verifier_kind`/`verification_strength` addition.

### Deferred to follow-up work

- **Holdout experiment (M7):** `mode='holdout'`, `holdout_set` /
  `holdout_leakage_check` population, `evaluation_holdout_match`, and baseline arms
  (`undirected` / `semantic-summary`). This slice persists the nullable columns +
  the holdout-gate triggers but refuses `mode='holdout'` with a deferred error.
- **Success compression (M6.1):** `success_space` mining, `success_invariant*`
  tables, boundary compression. This slice only records proposal outcomes; it does
  not compress the partial-successes it finds.
- **Search-policy mutation (M6.2):** evaluated failures become *eligible* for
  re-clustering but do not yet mutate a persisted search policy.
- **Real (non-fixture) verifier adapters** for external proof assistants / code
  execution: the tiered `Verifier` interface + deterministic in-process checkers
  ship here; wiring an external prover/executor is a separate, adapter-only change
  behind the same interface + strength stamp.

### Out of scope (do not implement)

- Automatic re-runs of clustering/mining/challenge after evaluation.
- Any change to invariant state as a result of evaluation (that is M6).
- A single fabricated scalar "proposal quality"; metrics stay typed/ordinal.
- Network/model calls in CI.

---

## Assumptions

1. **M5.1 frontier landed** (its migration + `frontier_proposal` with a NULL
   `result`, `frontier_target_invariant`, `frontier_nearest_cluster`, and the
   `Generator` provider role). Re-cement the frontier table/column names against
   the landed migration; this slice's migration is `+1` over it (expected **v15**;
   verify against `internal/store/migrations.go`, currently at `v13`).
2. **`provider_invocations.role`** currently admits at least
   `('normalize','invariant','challenge')` and, once M5.1 lands, `'generate'`.
   This slice widens it to also admit `'evaluate'` via the same guarded in-place
   `writable_schema` CHECK edit v11/v13 used (FK-safe by construction; idempotent
   on a fresh DB).
3. **Run lifecycle** is the create-`running` → `finalizeRun`/`failRun` contract
   used by mining/clustering; `run show` reflects the real outcome.
4. **The M4.2 evaluator is reusable as a real deterministic verifier.**
   `internal/invariant.Evaluate(predicate, signature)` already returns
   `satisfies|violates|unknown`; the `deterministic-check` verifier uses it to
   decide whether a proposal's mechanism signature actually violates its target
   invariant's predicate — this is a genuine code-owned check, not model judgment.
5. **New ID classes** follow the ULID+prefix convention in
   `internal/domain/id.go`: add `EvaluationRunIDPrefix` (e.g. `evr_`),
   `EvaluationIDPrefix` (e.g. `evl_`) with `New*`/`Validate*` and `id_test.go`
   round-trip + cross-class rejection tests.
6. **Offline CI:** the model-judge tier is a `FixtureVerifier`; deterministic
   verifiers are pure. No network.

If any surface differs at implementation time (renamed frontier tables, a
since-widened role CHECK, a changed run lifecycle), re-cement against the live API
before U2/U5; do not fabricate a shape.

---

## Requirements

- **R1. Strength is structural.** Every `evaluation` records `verifier_kind` and
  an ordinal `verification_strength`; a verdict from `model-judgment` and one from
  `deterministic-check` are distinguishable by strength even when the `verdict`
  string matches. (Falsify: attempt to store an evaluation without a strength → CHECK
  rejects it.)
- **R2. Cheapest-first, strongest-decisive.** The router tries verifiers in cost
  order and records the verdict + strength of the **strongest tier that returned a
  decisive verdict**; a tier returning `unknown`/`verification_blocked` does not
  set the result and routing falls back. The stored strength is never higher than
  the deciding verifier's tier.
- **R3. Deterministic verifiers own truth-sensitive verdicts.** Where a
  deterministic check can decide (e.g. the invariant-violation check via
  `invariant.Evaluate`, or counterexample search), the stored verdict comes from
  it, not the model tier. (Falsify: a fixture where the deterministic check
  decides `failure` but the model tier "claims" `success` — the stored result is
  `failure` / `deterministic-check`.)
- **R4. Verdict vocabulary is exactly the EPIC set.** `failure | partial_failure |
  partial_success | success | unknown | verification_blocked`, enforced by CHECK.
- **R5. `frontier_proposal.result` is populated transactionally** with the
  evaluation verdict in the same transaction that writes the evaluation; an
  evaluation and its proposal result never diverge.
- **R6. Failed proposals re-enter the atlas.** A `failure`/`partial_failure`
  evaluation makes the proposal's mechanism eligible for inclusion in a subsequent
  clustering pass (a persisted, queryable link/flag), so `FailureSpace(t+1) ⊇
  FailureSpace(t) + this failure`. (Verify: after an evaluated failure, the atlas
  reader reports the new failure as eligible.)
- **R7. No fake precision; metrics typed.** Confidence is `confidence_ordinal`;
  metrics use `evaluation_metric.metric_scale`. No invented float verdict score.
- **R8. Provenance / replay / lifecycle.** Model-tier evaluations record
  provider/model/config + raw payloads (role `'evaluate'`); deterministic tiers
  record tool identity/version; the run uses `running → completed/failed`;
  re-evaluation is append-only (a new `evaluation_run`), never a rewrite. Immutable
  rows (raw UPDATE/DELETE aborts).
- **R9. Holdout is refused, not half-built.** `mode='holdout'` is rejected at the
  service boundary with a deferred-to-M7 error; the nullable columns + gate
  triggers are present so M7 needs no schema retrofit.
- **R10. CLI + `--json`; offline determinism.** `evaluate`/`evaluation list/show`
  emit stable human + JSON; all tests use fixtures (no network/model).

Requirements (R1–R10) and implementation units (U0–U8) are separate axes.

---

## Key Technical Decisions

- **KTD-1 — Add `verifier_kind` + `verification_strength` to `evaluation`.** The
  `docs/persistence.md` `evaluation` blueprint stores `verdict` +
  `confidence_ordinal` + `notes` but NOT which tier produced the verdict — the one
  thing EPIC.md M5.2 makes central. This slice adds two columns:
  - `verifier_kind TEXT NOT NULL CHECK (verifier_kind IN
    ('deterministic-check','counterexample-search','reproducible-computation',
    'independent-evidence','independent-critic','model-judgment'))`;
  - `verification_strength TEXT NOT NULL CHECK (verification_strength IN
    ('deterministic','reproducible','independent-evidence','independent-critic',
    'single-model-judgment'))`.
  The two are related but distinct: `verifier_kind` is *what ran*,
  `verification_strength` is *its position in the hierarchy* (so multiple kinds can
  map to one strength band, and the ordering is explicit and documented). This is
  the structural encoding of `AGENTS.md`'s verification hierarchy and the single
  most important decision in the slice.
- **KTD-2 — Transcribe the evaluation schema from `docs/persistence.md`.** The
  `evaluation_run` / `evaluation` / `evaluation_metric` / `evaluation_run_metric`
  tables, their `metric_scale` CHECK triads, and the `evaluation_run_holdout_gate_*`
  triggers are already the contract; the migration copies them verbatim (fixing FK
  names to the shipped `run`/`problem`/`frontier_*`/`invariant_revision`/
  `cluster_run` tables and pluralized conventions) and adds KTD-1's columns. The
  holdout-only tables (`evaluation_holdout_match`, `holdout_set*`) and the baseline
  arms are **not** created here (M7).
- **KTD-3 — Cheap-first routing, strongest-decisive verdict.** The router is a
  pure, deterministic function `Route(proposal, context, verifiers) -> Decision`.
  Verifiers are ordered by (cost asc); the router executes them until one returns a
  **decisive** verdict (not `unknown`/`verification_blocked`), then records that
  verdict with the verifier's kind+strength. If a cheaper deterministic verifier
  decides, stronger tiers are not consulted (they cannot *raise* strength above a
  deterministic check — the deterministic check already sits at the top of the
  hierarchy). If only the model tier decides, strength is `single-model-judgment`.
  If nothing decides, the stored verdict is `verification_blocked` /
  (kind of the last tier tried) — an honest "we could not verify," not a guess.
- **KTD-4 — Deterministic verifiers reuse existing code-owned checks.** The
  `deterministic-check` verifier is not new truth machinery: it reuses
  `internal/invariant.Evaluate` (the M4.2 predicate evaluator) to decide whether the
  proposal's proposed mechanism signature actually *violates* the target
  invariant's predicate (the proposal's central claim). `satisfies` (invariant
  still holds) → the proposal did NOT break it → `failure`; `violates` (invariant
  broken as claimed) → at least `partial_success` pending counterexample search;
  `unknown` → fall through to the next tier. `counterexample-search` then looks for
  a known/synthetic family that refutes the claimed break. This makes "code owns
  truth-sensitive verdicts" literal, reusing shipped code rather than inventing a
  parallel checker.
- **KTD-5 — Failure re-entry is a persisted flag, not an auto-rerun.** A
  `failure`/`partial_failure` evaluation writes an `evaluated_failure` marker
  (either a column on the proposal or a thin link table) that the existing
  clustering population reader can include on the *next* `cluster build`. This
  honors `FailureSpace(t+1) = FailureSpace(t) + newly_evaluated_failures` without
  hidden orchestration (`AGENTS.md`: "do not add orchestration abstractions before
  the underlying typed operations exist"). The operator (or a later M6.2 policy)
  triggers the re-cluster explicitly.
- **KTD-6 — `mode='proposal'` only; holdout refused, not stubbed.** The gate
  triggers are copied verbatim so the schema is M7-ready, but the service rejects
  `mode='holdout'` with a clear deferred error. This keeps the FK-heavy holdout
  machinery honest (no half-populated `holdout_leakage_check`).
- **KTD-7 — Fixture-first verifiers** enable a fully offline end-to-end test and
  keep the model-judge tier from ever being the *silent* decider: in the fixtures,
  the deterministic tier decides the truth-sensitive cases and the model tier is
  exercised only where no deterministic check applies, stamped
  `single-model-judgment`.

---

## High-Level Technical Design

### Data flow

```text
frontier_proposal (result IS NULL, from M5.1)
  -> Evaluate(problem, [proposal_ids] | --all, policy=cheap-first)
       -> for each proposal, build a VerificationContext:
            target invariant predicate (from candidate_invariants/invariant_predicates),
            proposed mechanism signature, nearest-family signatures (from frontier_nearest_cluster)
       -> Router (pure, KTD-3): verifiers in cost order
            [deterministic-check] invariant.Evaluate(predicate, proposed_sig)  ---- decisive? --> record
            [counterexample-search] bounded search over persisted families ---- decisive? --> record
            [model-judgment] FixtureVerifier verdict+confidence_ordinal --------- decisive? --> record
            (none decisive) -> verdict=verification_blocked
       -> write evaluation_run (mode='proposal'), evaluation (+verifier_kind,
          +verification_strength, confidence_ordinal), evaluation_metric(s)
       -> UPDATE frontier_proposal.result = verdict  (SAME transaction, R5)
       -> if verdict in (failure, partial_failure): mark evaluated_failure (R6/KTD-5)
       -> run lifecycle: running -> completed/failed
  -> evaluation list/show reflect verdict + STRENGTH + provenance
  -> next `cluster build` MAY include evaluated failures (atlas growth)
```

### Persistence shape (new tables — additive, on top of M5.1)

Transcribed from `docs/persistence.md` 670–853 (adapted names), **plus KTD-1**:

```sql
CREATE TABLE evaluation_runs (
  id TEXT PRIMARY KEY,                              -- evr_<ulid>
  problem_id TEXT NOT NULL REFERENCES problems(id),
  run_id TEXT NOT NULL REFERENCES runs(id),
  frontier_generation_run_id TEXT REFERENCES frontier_generation_runs(id),
  invariant_revision_id TEXT REFERENCES invariant_revisions(id),
  cluster_run_id TEXT REFERENCES cluster_runs(id),
  -- holdout hooks retained for M7 (nullable; mode='holdout' refused by service):
  holdout_set_id TEXT,                              -- FK added with holdout tables in M7
  normalization_revision_id TEXT REFERENCES normalization_revisions(id),
  holdout_leakage_check_id TEXT,                    -- FK added in M7
  mode TEXT NOT NULL CHECK (mode IN ('proposal','holdout')),
  cutoff_time TEXT,
  baseline_type TEXT CHECK (baseline_type IN ('undirected','semantic-summary')),
  proposal_budget_count INTEGER,
  evaluation_budget_count INTEGER,
  routing_policy TEXT NOT NULL DEFAULT 'cheap-first'
                 CHECK (routing_policy = 'cheap-first'),   -- v0 only
  created_at TEXT NOT NULL
);
-- + the evaluation_run_holdout_gate_insert/update triggers VERBATIM (M7-ready).

CREATE TABLE evaluations (
  id TEXT PRIMARY KEY,                              -- evl_<ulid>
  evaluation_run_id TEXT NOT NULL REFERENCES evaluation_runs(id),
  proposal_id TEXT REFERENCES frontier_proposals(id),
  verdict TEXT NOT NULL
          CHECK (verdict IN ('failure','partial_failure','partial_success',
                             'success','unknown','verification_blocked')),   -- R4
  verifier_kind TEXT NOT NULL
          CHECK (verifier_kind IN ('deterministic-check','counterexample-search',
                                   'reproducible-computation','independent-evidence',
                                   'independent-critic','model-judgment')),  -- KTD-1
  verification_strength TEXT NOT NULL
          CHECK (verification_strength IN ('deterministic','reproducible',
                                           'independent-evidence','independent-critic',
                                           'single-model-judgment')),        -- KTD-1/R1
  confidence_ordinal TEXT,
  tool_name TEXT,                                   -- deterministic-tier identity
  tool_version TEXT,
  provider_invocation_id TEXT REFERENCES provider_invocations(id),  -- model tier only
  notes TEXT,
  created_at TEXT NOT NULL
);

CREATE TABLE evaluation_metrics (   -- transcribed metric_scale triad, unchanged
  id TEXT PRIMARY KEY,
  evaluation_id TEXT NOT NULL REFERENCES evaluations(id),
  metric_name TEXT NOT NULL,
  metric_scale TEXT NOT NULL CHECK (metric_scale IN ('numeric','ordinal','categorical')),
  numeric_value REAL, ordinal_value TEXT, ordinal_scale_key TEXT,
  ordinal_scale_version TEXT, categorical_value TEXT, comparator TEXT,
  created_at TEXT NOT NULL,
  CHECK ( /* the persistence.md numeric/ordinal/categorical exclusivity triad */ )
);
-- evaluation_run_metrics: identical shape keyed to evaluation_runs (run aggregates).
-- + immutability triggers (UPDATE/DELETE RAISE(ABORT,...)) for every new table.
-- + guarded provider_invocations.role widening to admit 'evaluate' (writable_schema, idempotent).
-- frontier_proposals.result populated in the SAME tx (R5); evaluated-failure marker (R6/KTD-5).
```

> **Blueprint reconciliation.** `docs/persistence.md` sketches the full evaluation
> layer including holdout mode, `evaluation_holdout_match`, and baseline arms. This
> slice ships the **`mode='proposal'` subset** + the KTD-1 strength columns and
> **intentionally does not create** `evaluation_holdout_match`, `holdout_set*`, or
> the baseline-comparison surfaces (all M7). A reviewer can confirm no holdout
> machinery leaked by checking those names are absent from the migration (only the
> nullable columns + the two holdout-gate triggers are present, M7-ready).

### Output structure

New pipeline views (`internal/pipeline/output.go`), all `--json`-stable:

- `EvaluationRunView` — id, problem, run, frontier_generation_run, mode,
  routing_policy, evaluation_count, created_at.
- `EvaluationView` — id, proposal_id, verdict, **verifier_kind**,
  **verification_strength**, confidence_ordinal, tool identity or
  provider_invocation, metrics (name/scale/value), notes.

Human output leads with verdict AND strength, e.g.
`E-07  proposal F-21  →  partial_success  [deterministic-check / deterministic]  eliminates S4, fails on S7`,
so a reader can never see an outcome without its epistemic strength.

---

## Implementation Units

- **U0. Dependency gate.** Confirm M5.1 frontier landed green (its migration +
  `frontier_proposals` with NULL `result`, `frontier_target_invariants`,
  `frontier_nearest_clusters`, `Generator` role). Record its migration version;
  this slice's migration is `+1` (expected **v15**; verify against
  `internal/store/migrations.go`, currently `v13`). Confirm/plan the
  `provider_invocations.role` widening to `'evaluate'`.
- **U1. Domain + verifier contract (`internal/verify`, pure — no SQL/Cobra).**
  ID classes (`EvaluationRunID`, `EvaluationID`) + tests; typed `Verdict`,
  `VerifierKind`, `VerificationStrength` enums with the documented hierarchy
  ordering; the `Verifier` interface (`Verify(ctx, VerificationContext) ->
  (Decision, error)` where `Decision{Verdict, Kind, Strength, ConfidenceOrdinal,
  Metrics, Notes}`); and the pure **`Route(ctx, verifiers, policy)`** function
  (cheap-first, strongest-decisive; KTD-3). Unit tests: routing picks the
  strongest decisive tier; a deterministic `failure` overrides a model `success`
  (R3); all-`unknown` → `verification_blocked` (KTD-3).
- **U2. Deterministic verifiers (`internal/verify`, pure).** `deterministic-check`
  reusing `internal/invariant.Evaluate` against the target predicate (KTD-4), and
  `counterexample-search` over supplied family signatures. Pure over already-loaded
  context structs; the DB loads happen in the U5 reader. Unit tests: violates →
  ≥partial_success; satisfies → failure; ambiguous read → unknown (fall-through).
- **U3. `Verifier` model tier + deterministic fixture (`internal/provider`).**
  `provider.Verifier` role (`Verify(ctx, request) (VerifyResponse, error)`) with a
  `FixtureVerifier` returning canned verdict+confidence keyed on request content;
  metadata records role `'evaluate'` + model/provider/version; raw payloads
  retained. No domain/SQL coupling.
- **U4. Evaluation persistence (migration vNN) + writer/reader.** New tables +
  immutability triggers + the two holdout-gate triggers verbatim + the guarded
  `provider_invocations.role` widening. `validateSchemaTables` extension
  (mandatory): append `evaluation_runs`, `evaluations`, `evaluation_metrics`,
  `evaluation_run_metrics`; probe the `verifier_kind`/`verification_strength` and
  `role='evaluate'` CHECKs. `store.PersistEvaluationRun(...)` transactional,
  populating `frontier_proposals.result` in the SAME tx (R5) and writing the
  evaluated-failure marker (R6). Readers: `GetEvaluationRun`,
  `ListEvaluations(problem)`, `GetEvaluation(id)`. Store tests: immutability;
  verdict/kind/strength CHECKs reject bad values; `result` mirrors the verdict;
  `mode='holdout'` insert is blocked by the gate trigger; role widening + child-FK
  survival from the pre-migration shape; evaluated-failure marker round-trips.
- **U5. Pipeline services + CLI + run lifecycle.** `Evaluate` (single proposal,
  `--problem --all`, cheap-first), building the `VerificationContext` from the
  store (target predicate, proposed signature, nearest-family signatures),
  invoking `Route`, persisting, and running `running → completed/failed`. Refuse
  `mode='holdout'` with a deferred-to-M7 error (R9). `evaluation list/show`
  readers. CLI (`cmd/newf`, thin) registered in `root.go`; `--json` contract in
  `output_test.go`.
- **U6. Failure-atlas re-entry wiring.** The persisted marker (KTD-5) + an
  atlas-reader inclusion path so a subsequent `cluster build` can pick up evaluated
  failures. Test: after an evaluated `failure`, the atlas/eligibility reader
  reports the mechanism as newly eligible (R6). No auto-rerun.
- **U7. Fixtures + regressions + end-to-end (offline).** Reuse the M5.1 frontier
  fixtures. Cases: (a) deterministic tier decides `failure`, model tier would say
  `success` → stored `failure`/`deterministic-check`/`deterministic` (R3);
  (b) deterministic `violates` + counterexample-search finds none → `partial_success`;
  (c) counterexample-search finds a refuter → `failure`/`counterexample-search`;
  (d) no deterministic check applies → model tier → `single-model-judgment`;
  (e) nothing decides → `verification_blocked`; (f) `result` populated in tx (R5);
  (g) failure re-entry marker set (R6); (h) re-evaluate → new `evaluation_run`,
  prior untouched (R8). Full end-to-end:
  `seed → mine → challenge→surviving → frontier generate → evaluate → evaluation show`,
  asserting run `completed` and a strength-stamped verdict.
- **U8. Docs + quality gates.** `docs/evaluation.md` (verifier hierarchy, routing,
  strength recording, verdict vocabulary, atlas re-entry, holdout deferral, CLI,
  run lifecycle), EPIC.md **M5.2** status note, `docs/persistence.md` vNN update
  (shipped evaluation subset + `verifier_kind`/`verification_strength` addition +
  holdout deferral). Gates: `go build ./...`, `go vet ./...`, `go test ./...`
  green; `gofmt -l .` empty; a boundary test at every new seam
  (router/verifiers, provider, store, CLI).

---

## Verification Contract

- Build/vet/test green; `gofmt` clean; offline CI (no network/model).
- **R1 (strength structural):** an `evaluation` without a
  `verification_strength` is rejected by CHECK; two evaluations with the same
  `verdict` but different tiers carry different strengths (asserted).
- **R2/KTD-3 (cheapest-first, strongest-decisive):** the router unit test pins the
  order and asserts a cheaper decisive verdict is used and a
  `unknown`/`verification_blocked` tier is skipped; the stored strength never
  exceeds the deciding tier.
- **R3/KTD-4 (deterministic owns truth):** a fixture where the deterministic check
  says `failure` and the model tier "claims" `success` stores `failure` /
  `deterministic-check`. (Falsify: force routing to the model tier first and assert
  the test fails.)
- **R4 (verdict vocabulary):** a direct INSERT with a verdict outside the EPIC set
  is rejected by CHECK.
- **R5 (result populated in tx):** after `Evaluate`, the `frontier_proposal.result`
  equals the evaluation verdict; a forced failure between the two writes leaves
  neither (single transaction).
- **R6/KTD-5 (atlas re-entry):** an evaluated `failure`/`partial_failure` marks the
  mechanism eligible for the next clustering pass; the atlas reader reports it.
- **R7 (no fake precision):** confidence is stored as `confidence_ordinal`; a
  numeric metric uses `metric_scale='numeric'` and the exclusivity CHECK holds.
- **R8 (provenance/replay/lifecycle):** a model-tier evaluation resolves to a
  `provider_invocations` row (role `'evaluate'`) with retained payloads; a
  deterministic-tier evaluation records `tool_name/tool_version` and NO provider
  row; re-evaluation creates a new `evaluation_run`; raw UPDATE/DELETE aborts; the
  run shows `completed`/`failed` via `run show`.
- **R9 (holdout refused):** `Evaluate` with `mode='holdout'` returns a
  deferred-to-M7 error; the gate trigger independently blocks a raw holdout-mode
  insert missing its holdout FKs.
- **R10:** each command emits stable `--json` (contract test).

## Definition of Done

Given a problem's frontier proposals (from M5.1, `result` NULL), `newf evaluate`
runs under the real run lifecycle and routes each proposal cheap-first to the
strongest verifier that can decide it, writing a durable `Evaluation` that records
the verdict **and its verification strength**, populates the proposal's `result`
in the same transaction, and — for failures — makes the proposal's mechanism
eligible to re-enter the failure atlas. Deterministic verifiers (reusing the M4.2
predicate evaluator + counterexample search) own truth-sensitive verdicts; the
model tier is the last resort and is honestly stamped `single-model-judgment`;
nothing outside the EPIC verdict vocabulary can be stored; holdout mode is refused
pending M7. The evaluated set is durable, provenance-preserving, immutable, and
stable enough to become the input contract for **M6.1 (compress partial successes
into success invariants)** and the M7 holdout experiment's scored arm.

## Risks & Dependencies

- **Hard dependency on M5.1** (`frontier_proposal` with NULL `result`). U0 gates
  start; re-cement table/column names and the migration number against the landed
  frontier migration.
- **Strength laundering is the central risk.** The whole slice exists to prevent a
  model verdict from being stored as if deterministic. The `verifier_kind` +
  `verification_strength` CHECKs, the deterministic-first router, and the R3
  falsification test are the load-bearing guards; they must not be bypassable by a
  confident model tier.
- **Verifier expressiveness vs. v0 scope.** v0 ships `deterministic-check`
  (predicate violation), `counterexample-search`, and a `model-judgment` fixture.
  `reproducible-computation` / `independent-evidence` / `independent-critic` kinds
  are in the CHECK vocabulary (so strengths are stable) but their real adapters are
  deferred behind the same interface — the strength column is ready for them.
- **Holdout FK weight.** The holdout machinery is heavy and FK-dense; shipping only
  the nullable columns + gate triggers (not the tables) keeps this slice honest and
  M7-ready without a half-built leakage path.
- **Atlas re-entry without orchestration.** Re-entry is a persisted marker, not an
  auto-rerun; the operator/M6.2 triggers re-clustering. This matches AGENTS.md's
  "no orchestration before the typed operations exist."

## Sources & Research

- EPIC.md **M5.2 — Evaluation and verifier routing** (goal, verification hierarchy,
  expected outcomes, exit condition) and **Gate D** (search claims measurable).
- `AGENTS.md` — "Verification hierarchy" (`formal proof > … > single-model
  judgment`), "Deterministic or independently checkable tools own truth-sensitive
  operations", `ModelJudgment != Verification`, "Failure is a first-class
  artifact", "Do not fabricate fake precision", provider provenance/fixtures.
- `docs/toolbox-dsl.md` — `evaluate Proposal -> Outcome`, `evaluate frontier {
  proof_check symbolic_check computation counterexample_search }`, `evaluate
  --cheap-first`, and `FailureSpace(t+1) = FailureSpace(t) + newly_evaluated_failures`.
- `docs/persistence.md` (670–853) — authoritative `evaluation_run` / `evaluation`
  / `evaluation_metric` / `evaluation_run_metric` schema + holdout-gate triggers;
  (99) provider-role set incl. `'evaluate'`.
- Shipped upstream surfaces: `internal/invariant/{predicate,engine}.go`
  (`Evaluate` reused as the deterministic verifier), `internal/store/migrations.go`
  (current `v13`; guarded `provider_invocations.role` widening pattern),
  `internal/pipeline/app.go` (`finalizeRun`/`failRun`), `internal/domain/id.go`,
  and the M5.1 frontier surfaces once landed.

## Product Contract preservation

- No SQL/Cobra in `internal/domain`, `internal/invariant`, `internal/verify`, or
  `internal/provider`; SQLite behind `internal/store`; CLI wiring thin in
  `cmd/newf`.
- Provider coupling confined to `internal/provider`; the verifier model tier
  records its role `'evaluate'`; deterministic verifiers + the router are pure
  domain-adjacent code with no vendor concepts.
- **No silent epistemic promotion:** verification strength is stored explicitly;
  a model verdict can never be recorded as deterministic; evaluating a proposal
  does not change any invariant's state (M6 owns that).
- Additive migration with immutability triggers + one guarded, tested
  `provider_invocations.role` widening; historical rows never rewritten; the
  frontier proposal `result` is populated, not overwritten across re-evaluations
  (each re-evaluation is a new append-only `evaluation_run`).
