---
artifact_contract: ce-unified-plan/v1
artifact_readiness: implementation-ready
execution: code
product_contract_source: ce-plan-bootstrap
origin: GitHub issue #20 (instagrim-dev/newf) — "M7 v0 experiment execution: arm runner, baselines, metrics, compare" (to be filed)
epic: EPIC.md milestone M7 — Historical holdout validation (v0 blinded benchmark; historical stays gated). Completes the EXECUTION half of plan 009.
depends_on: >-
  Plan 009 (blinded-benchmark v0) landed the DEFINITION half: migration v21
  experiment persistence (`holdout_sets`, `holdout_set_sources`,
  `holdout_source_dating`, `leakage_checks`, `experiment_runs`,
  `experiment_arms`, `experiment_metrics`, all immutable), the pure recovery
  detector (`internal/experiment/recovery.go`: `DetectRecovery`,
  `ComputeDiversity`, `MetricOrdinal`), the store surface (`RunLeakageCheck`,
  `PersistExperiment`, `GetExperiment`, `ListExperiments`,
  `ListProposalContents`, `CountHoldoutSourceDating`), the pipeline view/input
  types (`ExperimentRunInput/ShowInput/ListInput`, `ExperimentView`,
  `ExperimentArmView`, `ExperimentMetricView`, `LeakageCheckView`,
  `modeDisclaimer`), and `DefineExperiment`. This plan is ADDITIVE and adds NO
  migration (U0 re-verifies live `currentSchemaVersion`, observed = 21, and
  confirms no new tables are needed). It CONSUMES the M5.1 `GenerateFrontier`
  and M5.2 `Evaluate` services, the v17 `frontier_proposal_signatures` sidecar
  (via `ListProposalContents`), `canon.CompareWithProfile`, and plan 008's
  `--no-policy` baseline flag.
created: 2026-09-10
plan_type: feat
---

# feat: M7 v0 experiment execution — arm runner, baselines, metrics, compare

## Summary

Plan 009 defined the blinded-benchmark experiment artifact — persistence,
recovery detector, leakage-check store surface, view types, and
`DefineExperiment` — but the experiment cannot yet **run**. `RunExperiment`,
`ShowExperiment`, `ListExperiment`/`CompareExperiment`, the two baseline
provider roles (B1 summarize-next, B2 brainstorm), per-arm metric computation,
the `experiment` CLI command, and `docs/experiment.md` are all absent. This
slice ships that execution layer so the central thesis becomes a **measured**
claim rather than a demo:

> Given only the `train/` failure-space, does invariant-guided frontier
> generation (B3) recover the blinded `target/` family's structural move better
> than undirected/summary/brainstorm baselines (B0/B1/B2), under equal budgets,
> with blinding enforced and mode honestly stamped?

The decisive design moves are inherited, not reinvented:

- **One recovery rule for every arm** — `experiment.DetectRecovery` under the
  pinned profile. The baselines are never judged by a different rule (arm
  parity is a guard test), so a B3 win cannot come from a favorable metric.
- **Blinding is the gate key** — an experiment cannot COMPLETE without a
  passing, persisted `leakage_checks` row bound to it; the M5.2
  `mode='holdout'` evaluation gate lifts only in its presence (the condition
  plan 009's v21 reserved). A missing/failed check keeps both closed.
- **Mode is non-launderable** — `blinded` and `historical` keep disjoint
  conclusion vocabularies (CHECK-enforced in v21); every rendered view carries
  `mode_disclaimer`; `historical` execution stays refused end-to-end without
  `holdout_source_dating` rows (`BlindedRecovery != HistoricalPrediction`).
- **Budgets are stopping conditions** — every arm runs under the same persisted
  `proposal_budget_count`/`evaluation_budget_count`; exhaustion records
  `budget_exhausted`, never a silent truncation.
- **No side effects** — an experiment measures; it never mutates invariant
  lifecycles, policy, or the train atlas (a digest guard proves it).

Fully offline and deterministic (baseline providers are transparent deriving
fixtures; CI makes no model calls). No migration — the v21 substrate is
sufficient; this is pure pipeline + provider + CLI + docs wiring.

---

## Problem Frame

The loop ships end to end (M1–M6.2) and plan 009 can DEFINE a holdout set, but
`newf experiment run` does not exist: there is no code that, given a holdout
set, (a) runs the leakage check and refuses on failure, (b) canonicalizes the
quarantined target family, (c) drives each requested arm's
generation→evaluation chain under equal budgets, (d) feeds every arm's proposals
through the single recovery detector, (e) computes the EPIC metric vocabulary as
exact counts + ordinals, and (f) persists an immutable, mode-stamped experiment
revision that `show`/`compare` read back. `DefineExperiment` produces the split;
nothing consumes it.

Concretely inert today:

- **The recovery detector** (`DetectRecovery`, `ComputeDiversity`) is pure and
  tested but has no caller — no arm feeds it proposal content.
- **The leakage checker** (`RunLeakageCheck`) exists but is not wired into
  experiment completion or the M5.2 holdout-gate lift.
- **`PersistExperiment`/`GetExperiment`/`ListExperiments`** and every
  `Experiment*View` type exist but are never written or rendered by a service.
- **Baselines B1/B2** have no provider role at all; B0 (the generator with no
  targets + `--no-policy`) and B3 (the shipped loop) have entrypoints but no
  arm harness that binds them to a budget and a holdout set.

### Governing constraints (EPIC M7, AGENTS.md, plan 009 KTDs)

- **Arm parity:** all arms share one recovery rule, one profile version, one
  budget pair; the rule/profile versions are the experiment's provenance, not
  per-arm knobs.
- **Blinding enforced, not promised:** completion requires `passed=1`; the
  holdout evaluation gate stays closed otherwise.
- **Mode honesty:** disjoint conclusion vocabularies; `historical` refused
  without dated evidence; disclaimer on every view.
- **No epistemic promotion / no side effects:** measurements only.
- **Deterministic + reproducible:** identical (split, budgets, rule/profile
  versions) ⇒ idempotent experiment revision; metrics re-derivable from rows.
- **Offline:** baseline providers are deriving fixtures; no network in CI.

---

## Scope

### In scope

1. **Baseline provider roles (`internal/provider`):** `SummarizeNextProposer`
   (B1) and `Brainstormer` (B2) interfaces + `DerivingFixture*`
   implementations, each with `Identity()`/payload retention per the
   M5.2/M6.1/M6.2 precedent, and a role widening for `'summarize-next'` /
   `'brainstorm'` in `provider_invocations.role`. **Substrate check at U0:** if
   plan 009's v21 already widened these roles, this adds NO migration; if not,
   it is a single guarded `editTableCheckInPlace` step (its own micro-migration
   at the next free integer). The fixtures are transparent rules over the same
   cohort facts B3 sees, so B1/B2 are honest weak baselines, not strawmen.
2. **Arm runner (`internal/pipeline`, `RunExperiment`):** given a holdout set,
   (a) start a run (`running`), (b) `RunLeakageCheck` and REFUSE completion on
   `passed=0`, (c) canonicalize the quarantined target representative, (d) for
   each requested arm build its generation+evaluation chain under the shared
   budgets, (e) collect proposal content via `ListProposalContents` and call
   `DetectRecovery`/`ComputeDiversity`, (f) compute `experiment_metrics`, (g)
   `PersistExperiment` transactionally and finalize the run. Idempotent on the
   plan-009 experiment identity tuple.
3. **Arm wiring (B0–B3):** B0 = `GenerateFrontier` with no invariant targets +
   `--no-policy`; B1 = `SummarizeNextProposer`; B2 = `Brainstormer`; B3 = the
   shipped generate→evaluate loop (policy-biased unless the run requests
   unbiased). Every arm's proposals are persisted through the existing frontier
   path so `ListProposalContents` and the recovery detector treat them
   identically.
4. **Metric computation (pure helper):** the EPIC vocabulary as exact
   numerator/denominator + `MetricOrdinal`: held-out family recovery (recovered
   + first-recovery rank + nearest classification), structural-break recovery,
   mechanistic diversity (distinct-family count), normalized redundancy (dedup
   rate), information gain per evaluated proposal (ordinal), budget accounting.
   Per-arm, per-metric rows.
5. **Show/List/Compare services + CLI:** `ShowExperiment`, `ListExperiments`,
   and a `CompareExperiment` read view (arms side-by-side on recovery + first
   rank + diversity). `newf experiment define|run|show|list|compare`, all
   `--json`, mode disclaimer on every render; register `newExperimentCommand`
   in `cmd/newf/root.go` (currently unregistered).
6. **Tests at every seam:** baseline provider determinism + payload retention;
   arm-parity (all arms share one recovery rule — falsify by giving B3 a
   different rule and watching parity fail); leakage gate (seeded leak blocks
   completion AND keeps the holdout eval gate closed; clean split passes);
   equal-budget enforcement + `budget_exhausted`; metric truth table; mode
   stamped in every view; historical refused without dating rows; idempotent
   re-run; **no-side-effects digest guard** over invariant/policy/atlas tables;
   end-to-end blinded experiment over corpus fixtures (B0 vs B3).
7. **Docs:** `docs/experiment.md` (mode discipline, leakage audit as gate key,
   the one recovery rule, arms + equal budgets, metric vocabulary, historical
   handoff), EPIC M7 status note ("v0 blinded benchmark delivered; historical
   gated"), `corpus/README.md` cross-link, README loop update.

### Deferred to Follow-Up Work

- **`mode=historical` execution** — needs an externally auditable dated corpus;
  the gate lift + `holdout_source_dating` schema already exist (plan 009), so
  the trigger is a corpus PR carrying dated evidence, not code.
- **Confidence calibration + synthetic-failure-usefulness metrics** — need real
  model-campaign volume fixtures cannot honestly simulate; additive
  `experiment_metrics` rows later.
- **B2 diversity beyond distinct-family counts**, **policy-conditioned
  generator prompting** (plan 008 deferral), **cross-discipline litmus (M8)**.

### Out of scope

- Any migration beyond an optional role-CHECK widen; any chronological claim
  from the shipped corpus; any experiment side effect on research state; scalar
  composite scores; solving Erdős–Straus.

---

## Key Technical Decisions

- **KTD-1 — The arm harness owns budgets and parity; arms are pluggable
  proposers.** `RunExperiment` fixes (recovery rule version, profile version,
  proposal budget, evaluation budget) ONCE and passes them to every arm. An arm
  is `{name, proposeFn}`; its proposals always flow through the existing
  frontier persistence + `ListProposalContents` + `DetectRecovery`. This is the
  structural guarantee of arm parity: the harness, not the arm, applies the
  rule.
- **KTD-2 — Completion is gated on a passing leakage check.** `RunExperiment`
  writes `leakage_checks` first; `passed=0` finalizes the run `failed` with a
  named reason and persists NO experiment conclusion. The M5.2 holdout-mode
  evaluation gate lifts only when a passing check bound to the holdout set
  exists — reusing plan 009's reserved condition, no schema change.
- **KTD-3 — B0 is the generator stripped, not a new engine.** B0 =
  `GenerateFrontier` with an empty surviving-target set + `--no-policy`; it
  exercises the identical ranking/evaluation path as B3 minus invariant
  guidance, so the B0-vs-B3 delta isolates exactly the thing under test.
- **KTD-4 — Baselines are transparent deriving fixtures.** B1/B2 emit
  deterministic proposals from the same cohort facts; their v0 value is the
  HARNESS (equal budgets, blinding, one rule), stated plainly in docs.
- **KTD-5 — Metrics are exact counts + ordinals, per arm.** No fabricated
  scalar; `experiment_metrics` rows carry numerator/denominator/ordinal.
  `compare` is a read view over rows; changing the metric set is additive.
- **KTD-6 — Idempotent + reproducible.** Re-running with the same (holdout set,
  budgets, rule/profile versions, arm set) returns the existing experiment
  revision; new inputs yield the next.
- **KTD-7 — No side effects, proven.** A digest over invariant-state, policy,
  and atlas tables is identical before/after a run; the experiment writes only
  experiment-namespaced rows (+ arm-owned generation/evaluation runs, which are
  normal loop artifacts, never promotions).

---

## Data flow

```text
DefineExperiment (plan 009) -> holdout_set (train problem, quarantined target, withheld sources)
        |
        v
[ RunExperiment: start run (running) ]
        |
        v
[ RunLeakageCheck ]  provenance joins: no train signature/cluster/invariant/proposal derives from a withheld source
        |  passed=0 -> finalize run FAILED, no conclusion; holdout eval gate stays closed  (KTD-2)
        |  passed=1
        v
[ canonicalize quarantined target representative ]  (canon, one profile version)
        |
        v
for each arm in {B0,B1,B2,B3} under SHARED (proposal_budget, evaluation_budget):   (KTD-1)
    propose -> persist via frontier path -> evaluate -> ListProposalContents
        |
        v
[ DetectRecovery + ComputeDiversity ]  ONE rule for every arm  ->  ArmRecovery + DiversityFacts
        |
        v
[ metric computation ]  exact num/den + MetricOrdinal, per arm            (KTD-5)
        |
        v
[ PersistExperiment (transactional, immutable, mode-stamped) ]  experiment_runs + arms + metrics + leakage link
        |
        v
ShowExperiment / CompareExperiment (read views; mode_disclaimer on every render)
```

---

## Implementation Units

### U0. Re-verify substrate before wiring
Confirm live `currentSchemaVersion` (observed = 21) and that
`experiment_runs`/`experiment_arms`/`experiment_metrics`/`leakage_checks` and
the disjoint conclusion-vocabulary CHECKs already exist (plan 009). Confirm
whether `provider_invocations.role` already admits `'summarize-next'`/
`'brainstorm'`; if not, this slice adds one guarded `editTableCheckInPlace`
micro-migration at the next free integer, else NO migration. Re-cement the
`GenerateFrontier`/`Evaluate` input signatures, `ListProposalContents` row
shape, `DetectRecovery`/`ComputeDiversity`/`MetricOrdinal` signatures, and
`PersistExperiment`'s `ExperimentRecord`/arm/metric row types against live
main (the repeatedly-learned concurrent-writer lesson).

### U1. Baseline provider roles (`internal/provider`)
`SummarizeNextProposer` (B1) and `Brainstormer` (B2) interfaces +
`DerivingFixture*` implementations that emit deterministic candidate signatures
from the same cohort facts B3 consumes; `Identity()` + payload retention +
order-independent fingerprint. Optional role widen (U0). Tests: determinism,
fixed order, payload retention, fingerprint order-independence.

### U2. Arm harness + `RunExperiment` (`internal/pipeline`)
Run lifecycle; leakage check + refusal (KTD-2); target canonicalization; the
`{name, proposeFn}` arm abstraction applying shared budgets (KTD-1); per-arm
generation→evaluation chains through the existing frontier path; recovery +
diversity via the pure detectors; `budget_exhausted` stopping condition;
`PersistExperiment` (idempotent, KTD-6). Tests: leakage-gate refusal + clean
pass; equal-budget enforcement + `budget_exhausted`; arm parity; idempotent
re-run; no-side-effects digest guard (KTD-7).

### U3. Arm wiring B0–B3
B0 = `GenerateFrontier` (no targets, `--no-policy`, KTD-3); B1/B2 = U1
providers; B3 = shipped generate→evaluate. All persist through the same
frontier path so the recovery detector treats every arm identically. Test:
B0-vs-B3 delta over corpus fixtures reflects invariant guidance alone.

### U4. Metric computation (pure helper)
EPIC vocabulary as exact num/den + `MetricOrdinal` per arm (KTD-5): family
recovery + first-recovery rank + nearest classification, structural-break
recovery, mechanistic diversity, normalized redundancy, information gain per
evaluated proposal, budget accounting. Truth-table tests.

### U5. Show / List / Compare + CLI
`ShowExperiment`, `ListExperiments`, `CompareExperiment` (side-by-side arm read
view). `newf experiment define|run|show|list|compare` with `--json`, mode
disclaimer on every render; register `newExperimentCommand` in
`cmd/newf/root.go`. Tests: mode stamped in every view; historical refused
without dating rows (schema-path only); `--json` contract stability.

### U6. End-to-end + regression fixtures
Full offline chain: train ingested + target quarantined (corpus split) →
`experiment define` → `experiment run --arms b0,b3` → leakage pass → recovery
computed by code → metrics persisted → `show`/`compare` reconstruct the
comparison from rows only; idempotent re-run returns the same revision; digest
guard proves no research-state mutation.

### U7. Docs + gates
`docs/experiment.md`, EPIC M7 status note, `corpus/README.md` cross-link,
README loop update. Gates: `go build ./...`, `go vet ./...`, `go test ./...`
green, `gofmt -l .` empty; boundary test at every new seam.

---

## Verification Contract

- **Arm parity:** every arm's proposals pass through one `DetectRecovery` under
  one profile/rule version; a test that hands B3 a different rule fails parity.
- **Blinding is the gate key:** a seeded leak (train normalization referencing
  a withheld target source) fails `RunLeakageCheck`, blocks experiment
  completion, and keeps the M5.2 `mode='holdout'` evaluation gate closed; a
  clean split passes and lifts it.
- **Equal budgets are stopping conditions:** all arms receive identical
  budgets; an arm hitting its budget records `budget_exhausted` and stops;
  unequal budgets refuse to run.
- **Mode honesty:** a blinded experiment cannot record a historical conclusion
  (CHECK probe from v21); `historical` refused end-to-end without
  `holdout_source_dating` rows; every view carries `mode_disclaimer`.
- **Reproducibility:** identical (split, budgets, rule/profile versions, arms)
  ⇒ idempotent experiment revision; metrics re-derivable from rows.
- **No side effects:** invariant-state, policy, and atlas table digests are
  identical before/after a run (KTD-7).

## Definition of Done

`newf experiment define/run/show/list/compare` executes a leakage-audited,
equal-budget, mode-stamped blinded-benchmark experiment over the corpus split,
running B0–B3 arms through one code-owned recovery rule, computing the EPIC
metric vocabulary as exact counts + ordinals, and persisting an immutable,
revisioned experiment artifact that `show`/`compare` reconstruct from rows
alone — with `historical` mode still structurally refused until dated evidence
exists, and a digest guard proving the experiment mutated no research state.
EPIC M7 v0 is delivered.

## Risks

- **Benchmark-to-history laundering** — the central risk (inherited from plan
  009); mitigated by v21 mode CHECKs, disjoint vocabularies, stamped views, and
  the corpus README cross-link. This slice adds the runtime paths, so its tests
  must re-assert the disclaimer on every new render.
- **Arm-parity erosion** — a future arm could quietly special-case recovery;
  the harness-owns-the-rule structure (KTD-1) + parity test are the guard.
- **Fixture-weak baselines** — B1/B2 are transparent rules, not models; the v0
  value is the HARNESS, stated in docs.
- **Migration/number drift (repeatedly learned)** — U0 re-verifies
  `currentSchemaVersion` and adds a role-widen micro-migration only if plan 009
  did not already cover it.
- **Concurrent-writer churn** — `GenerateFrontier`/`Evaluate`/store signatures
  shifted mid-slice before; re-cement U2/U3 against live APIs at start.

## Sources

- `EPIC.md` M7 (experiment shape, baselines B0–B3, metric vocabulary,
  acceptance criterion) + stopping conditions (`AGENTS.md`).
- `docs/plans/2026-09-10-009-feat-blinded-benchmark-validation-plan.md` — the
  definition half this plan completes (mode discipline, leakage gate, recovery
  rule, persistence v21).
- `corpus/README.md` — the blinded-benchmark reclassification and blinding
  discipline.
- Live surfaces at drafting (schema v21): `internal/pipeline/experiment.go`
  (`DefineExperiment` + all view/input types), `internal/experiment/recovery.go`
  (`DetectRecovery`/`ComputeDiversity`/`MetricOrdinal`),
  `internal/store/experiment_store.go` (`RunLeakageCheck`, `PersistExperiment`,
  `GetExperiment`, `ListExperiments`, `ListProposalContents`),
  `internal/pipeline/frontier.go` (`GenerateFrontier`, `--no-policy`),
  `internal/pipeline/evaluation.go` (`Evaluate`, holdout gate),
  `internal/store/migrations.go` (role CHECK precedent),
  `docs/search-policy.md` (`--no-policy` baseline).

## Product Contract

- CLI additions (stable `--json`): `experiment run|show|list|compare` (joining
  the already-shipped `experiment define`).
- Epistemic guarantee: blinded results are never presentable as historical;
  blinding is code-audited and gates completion + the holdout eval path;
  recovery is one deterministic rule across arms; budgets and stopping
  conditions are persisted; experiments have no side effects on the research
  state they measure.
