---
artifact_contract: ce-unified-plan/v1
artifact_readiness: implementation-ready
execution: code
product_contract_source: ce-plan-bootstrap
origin: GitHub issue #19 (instagrim-dev/newf) — "Blinded-benchmark validation experiment (M7 v0)" (to be filed)
epic: EPIC.md milestone M7 — Historical holdout validation (v0 = blinded benchmark; historical mode stays gated)
depends_on: >-
  The full loop through M6.1 (commit `24cdc24`, schema v18 at drafting) and the
  in-flight M6.2 search-policy slice (plan 008, migration v19,
  `frontier_generation_policy`, `--no-policy` baseline). This plan is migration
  **v20** (U0 re-verifies `currentSchemaVersion` before numbering — the
  thrice-learned migration-drift lesson). It CONSUMES: the corpus
  train/target split (`corpus/README.md` — a SYNTHETIC BLINDED BENCHMARK,
  explicitly NOT a historical holdout), the M5.2 evaluation layer's reserved
  holdout hooks (`evaluation_runs.mode='holdout'` refused at the service
  boundary AND by a gate trigger; nullable `holdout_set_id` /
  `holdout_leakage_check_id` / `cutoff_time` / `baseline_type` /
  budget columns retained so M7 needs no schema retrofit), the v17
  `frontier_proposal_signatures` content sidecar (recovery detection compares
  proposal content, not hashes), and `canon.CompareWithProfile` (code-owned
  structural comparison).
created: 2026-09-10
plan_type: feat
---

# feat: Blinded-benchmark validation experiment (M7 v0)

## Summary

Implement the first scientifically meaningful end-to-end milestone — EPIC
**M7** — at the epistemic strength the substrate actually supports:

> Given only the `train/` failure-space, does invariant-guided frontier
> generation recover the **structural move** of a blinded `target/` family
> better than undirected baselines?

The governing constraint is the one the corpus reclassification already
cemented (`corpus/README.md`): the repository owns a **synthetic blinded
benchmark**, not a historical holdout. An earlier corpus version falsely
claimed the affine-lattice target was post-cutoff; that claim was removed and
must never re-enter through the experiment layer. This plan therefore ships
**two experiment modes with distinct, non-interchangeable epistemic claims**:

```text
mode=blinded     (SHIPS)  claim: structural-move recovery under enforced blinding.
                          No chronological claim is made or implied.
mode=historical  (GATED)  claim: prediction of a historically later advance.
                          Requires externally auditable, DATED source evidence
                          for every withheld artifact. The substrate has none,
                          so this mode remains refused end to end — exactly as
                          M5.2's holdout gate refuses today — with the refusal
                          reason naming what evidence would lift it.
```

A blinded-benchmark result can never be silently promoted into a historical
claim: mode is stamped on every experiment artifact, the two modes have
disjoint CHECK vocabularies for their conclusions, and rendering an experiment
omits no mode qualifier (`ModelJudgment != Verification` has a sibling here:
**BlindedRecovery != HistoricalPrediction**).

### What "recovery" means (code-owned)

The acceptance question turns on whether a generated proposal *recovers the
held-out structural move*. That judgment is deterministic, not model-voted:

```text
recovery(proposal, target_family) =
  canon.CompareWithProfile(proposal.ProposedSignature, target_representative)
    classification ∈ {identical, mechanism-near}      -> structural recovery
    surface-distinct+mechanism-near                    -> structural recovery
    mechanism-distinct / unknown                       -> no recovery
```

The proposal's canonical content comes from the v17
`frontier_proposal_signatures` sidecar; the target family's representative is
canonicalized from `target/` material ingested into a **quarantined** problem
workspace. Nothing about recovery is prose.

### Blinding is enforced, not promised

The experiment refuses to run (and refuses to *complete*) without a passing,
persisted **leakage check**: a code-computed audit that no train-side artifact
(source, snapshot, normalization, signature, cluster member, invariant support,
proposal provenance) derives from a withheld source. The check is a durable
row; the M5.2 evaluation gate for `mode='holdout'` lifts **only** in the
presence of a passing leakage check bound to the same experiment — the
condition the v15 gate comment reserved.

### Baselines and metrics

Arms (EPIC): `B0` undirected generation (no invariant targets, `--no-policy`),
`B1` semantic-summary + propose-next-step (provider role, deriving fixture),
`B2` diverse brainstorming without failure invariants (provider role, deriving
fixture), `B3` invariant-guided `newf` (the shipped loop, optionally
policy-biased). Every arm runs under identical proposal/evaluation budgets
(persisted), and every arm's proposals flow through the SAME code-owned
recovery detector — the baselines are not judged by a different rule.

Metrics are exact counts + ordinal bands, never fabricated scalars, drawn from
the EPIC vocabulary: held-out family recovery (per arm: recovered yes/no +
nearest classification + rank of first recovering proposal),
structural-break recovery, invariant precision against counterexamples,
mechanistic diversity (distinct families among proposals), normalized
redundancy (dedup rate), information gain per evaluated proposal (ordinal),
and budget accounting.

The acceptance criterion is NOT "solve Erdős–Straus" and is NOT "the benchmark
proves historical prediction." It is: **the experiment artifact makes the
B3-vs-baselines comparison reproducible, leakage-audited, and honestly
labeled** — so that when a genuinely dated corpus exists, `mode=historical`
lifts with no schema or metric retrofit.

---

## Problem Frame

Every stage of the loop now ships (M1–M6.2 in flight), but nothing composes
them into an *experiment*: a versioned artifact that fixes the split, enforces
the blinding, runs the arms under equal budgets, applies one recovery rule,
and persists the comparison. Without it, "the loop works" is a demo claim, not
a measured one — precisely the "open-problem theater" AGENTS.md warns about.
And without the mode discipline, the first benchmark success would inevitably
be miscited as historical evidence — the exact failure the corpus README
documents happening once already.

### Governing constraints

- **Mode is first-class and non-launderable.** `blinded` and `historical`
  carry disjoint conclusion vocabularies; historical stays refused without
  dated-source evidence rows; every view/render stamps the mode.
- **Blinding by construction.** Withheld sources live in a quarantined problem
  (separate workspace/DB recommended, separate problem id required); the
  leakage check is code-computed over provenance joins and persisted; a failed
  or missing check blocks experiment completion and keeps the M5.2 holdout
  evaluation gate closed.
- **One recovery rule for all arms.** `canon.CompareWithProfile` under the
  pinned profile; classification thresholds are the experiment's provenance,
  not per-arm knobs.
- **Equal budgets, persisted.** `proposal_budget_count` /
  `evaluation_budget_count` (the v15 columns) bind every arm; exhausting a
  budget is a recorded stopping condition (`budget_exhausted`), not a silent
  truncation.
- **No promotion.** An experiment produces measurements; it does not change
  any invariant's lifecycle state, mutate policy, or write into the train
  atlas. Failed proposals still re-enter the atlas only through the existing
  `evaluated_failures` path.
- **Deterministic fixtures / offline CI**, additive migration **v20** with
  immutability triggers, run lifecycle `running → completed/failed`, and the
  guarded role-CHECK widening for the two new baseline provider roles
  (`'summarize-next'`, `'brainstorm'`) via the shipped in-place pattern.

---

## Scope

### In scope

1. **Experiment persistence (migration v20):** `holdout_sets` (mode
   `blinded|historical`, name, nullable `cutoff_time` — CHECK-required when
   historical), `holdout_set_sources` (withheld source links + the blueprint's
   problem-guard triggers), `holdout_source_dating` (historical mode only:
   per-source dated evidence — locator + date + provenance; its ABSENCE is
   what keeps historical refused), `leakage_checks` (code-computed audit
   result + per-category counts + pass/fail), `experiment_runs` (mode, split,
   budgets, profile, recovery-rule version, stopping condition),
   `experiment_arms` (B0–B3, generation/evaluation links, per-arm counts),
   `experiment_metrics` (metric name from the EPIC vocabulary + exact
   numerator/denominator + ordinal). All immutable.
2. **Leakage checker** (`internal/store` + pure summary): provenance joins
   proving no train-side signature/cluster/invariant/proposal derives from a
   withheld source; persisted as `leakage_checks`; wired into the M5.2
   holdout-gate lift (the gate trigger's reserved condition).
3. **Recovery detector** (`internal/experiment`, pure): quarantined target
   family canonicalization → `CompareWithProfile` against every arm proposal's
   v17 sidecar content → per-proposal classification + per-arm recovery facts.
4. **Baseline provider roles** (`internal/provider`): `SummarizeNextProposer`
   (B1) and `Brainstormer` (B2) interfaces + deriving fixtures; B0 = the
   existing generator invoked with no invariant targets and `--no-policy`;
   B3 = the shipped loop.
5. **Pipeline + CLI:** `newf experiment define --problem <id> --target-problem
   <id> [--mode blinded]`, `newf experiment run <experiment-id> [--arms b0,b3]
   [--proposal-budget n] [--evaluation-budget n]`, `newf experiment show`,
   `newf experiment compare`; all `--json`; run lifecycle throughout.
6. **Tests:** leakage check catches a seeded leak (and passes clean);
   historical mode refused without dating rows and lifts with them (schema
   path only — no historical data ships); recovery detector truth table over
   fixture signatures; equal-budget enforcement + `budget_exhausted`; arm
   determinism; mode stamped in every view; end-to-end blinded experiment over
   the corpus fixtures with B0 vs B3.
7. **Docs:** `docs/experiment.md` (mode discipline, leakage audit, recovery
   rule, baselines, metrics), persistence v20 section, EPIC M7 status note
   (explicitly "v0 = blinded benchmark; historical gated"), corpus README
   cross-link, README loop update.

### Deferred to Follow-Up Work

- **`mode=historical` execution** — split until an externally auditable dated
  corpus exists; sink: `holdout_source_dating` rows + the gate lift already
  shipped here; trigger: a corpus PR carrying dated, source-locatable
  evidence per withheld artifact.
- **B2 diversity scoring beyond distinct-family counts**; **confidence
  calibration** and **synthetic-failure usefulness** metrics (EPIC lists
  them; they need M4.3 synthetic-artifact volume that fixtures cannot honestly
  simulate yet) — sink: `experiment_metrics` rows (additive), trigger: first
  real-model campaign.
- **Cross-discipline litmus (M8).**

### Out of scope

- Any historical/chronological claim from the shipped corpus; any change to
  invariant lifecycles, policy, or the train atlas from within an experiment;
  scalar composite scores; solving Erdős–Straus.

---

## Key Technical Decisions

- **KTD-1 — Two problems, one experiment.** The blinded target is ingested
  into a separate, quarantined problem (`--target-problem`); the experiment
  row binds (train problem, target problem, holdout set). Cross-problem
  provenance joins are exactly what the leakage checker audits; the
  quarantine makes accidental co-ingestion structurally visible instead of
  merely procedurally forbidden.
- **KTD-2 — The leakage check is the gate key.** `leakage_checks(id, holdout
  set, per-category leak counts, passed)` is written only by the code auditor.
  Experiment completion requires `passed=1`; the M5.2 holdout-mode gate
  condition becomes "a passing leakage check bound to this evaluation's
  holdout set exists" — the reserved lift, no schema retrofit.
- **KTD-3 — Recovery is a comparison classification, not a threshold knob.**
  `recovery-rule/v1` = pinned profile + the classification set
  {identical, mechanism-near, surface-distinct+mechanism-near} counting as
  recovery. The rule version is stamped on the experiment; changing the rule
  is a new experiment, never a re-interpretation.
- **KTD-4 — Arms are rows, not modes of one run.** Each arm gets its own
  generation/evaluation chain under the same budgets, persisted with links;
  `experiment_metrics` rows are per-arm and per-metric with exact
  numerator/denominator. Comparison is a read view over rows.
- **KTD-5 — Disjoint conclusion vocabularies.**
  `blinded  -> conclusion IN ('structural_recovery','no_recovery','inconclusive')`
  `historical -> conclusion IN ('predicts_later_advance','fails_to_predict','inconclusive')`
  enforced by CHECK against the mode column; a blinded experiment structurally
  cannot record a historical conclusion.
- **KTD-6 — Budgets are stopping conditions.** Arms record
  `stopping_condition IN ('completed','budget_exhausted','no_information_gain',
  'verification_blocked')` from the AGENTS.md vocabulary; truncation is never
  silent.

---

## Implementation Units

- **U0.** Re-verify substrate: `currentSchemaVersion` (v19 expected after the
  in-flight M6.2 lands; this slice is v20 — renumber if drifted AGAIN), the
  v15 holdout gate trigger text, the v17 sidecar surface, plan-008's
  `--no-policy` baseline flag.
- **U1.** Migration v20 (tables above + immutability + problem-guard triggers
  + role widen for `'summarize-next'`/`'brainstorm'`) + `validateSchemaTables`
  extension (tables + role assertions).
- **U2.** Leakage checker (store joins + pure summary + persistence) and the
  holdout-gate lift keyed on a passing check. Tests: seeded leak caught per
  category; clean split passes; gate stays closed on fail/missing.
- **U3.** `internal/experiment` recovery detector (pure) + truth-table tests
  over fixture signatures (near/distinct/unknown, sidecar round-trip).
- **U4.** Baseline provider roles + deriving fixtures (B1 summarize-next, B2
  brainstorm) with determinism + payload-retention tests; B0 wiring through
  the existing generator with no targets + no policy.
- **U5.** Pipeline services (`DefineExperiment`, `RunExperiment`,
  `ShowExperiment`, `CompareExperiment`) with run lifecycle, equal-budget
  enforcement, per-arm chains, metric computation. CLI + `--json` contracts.
- **U6.** End-to-end blinded experiment over corpus fixtures (train ingested,
  target quarantined): leakage pass, B0 + B3 arms, recovery facts computed by
  code, mode stamped everywhere, idempotent re-run returns the existing
  experiment revision.
- **U7.** Docs + EPIC status + gates (`go build/vet/test`, `gofmt -l` empty).

---

## Verification Contract

- **Mode honesty:** a blinded experiment cannot record a historical
  conclusion (CHECK probe); `mode=historical` refused end-to-end without
  `holdout_source_dating` rows and lifts with them (schema-path test only);
  every emitted view carries the mode.
- **Blinding:** a deliberately leaked source (target source referenced by a
  train-side normalization) fails the leakage check with the category named;
  experiment completion and the holdout evaluation gate both stay closed.
- **Recovery is code-owned:** fixture proposals structurally near the target
  classify as recovery, distinct ones do not, ambiguous content is
  `inconclusive` — the same detector for every arm (falsify: give B3 a
  different rule and watch the arm-parity test fail).
- **Budgets:** an arm hitting its budget records `budget_exhausted` and stops;
  arms receive identical budgets (probe: unequal budgets refuse to run).
- **Reproducibility:** identical split + budgets + rule version ⇒ idempotent
  experiment revision; metrics are exact counts + ordinals, re-derivable from
  persisted rows.
- **No side effects:** after an experiment, invariant states, policies, and
  the train atlas are byte-identical (probe by digest over the affected
  tables).

## Definition of Done

`newf experiment define/run/show/compare` executes a leakage-audited, equal-
budget, mode-stamped blinded-benchmark experiment over the corpus split,
computing held-out structural-move recovery with one code-owned rule across
B0/B1/B2/B3 arms, persisted as an immutable, revisioned artifact — and the
`historical` mode remains structurally refused until dated evidence exists,
so the first genuinely historical run needs data, not code.

## Risks

- **Benchmark-to-history laundering** — the central risk; mitigated by mode
  CHECKs, disjoint vocabularies, stamped views, and the corpus README
  cross-link. A reviewer should be unable to quote any artifact as historical
  evidence without visible fraud.
- **Migration drift (fourth occurrence likely)** — M6.2 is in flight at v19;
  U0 renumbers.
- **Fixture-weak baselines** — B1/B2 fixtures are transparent rules, not
  models; the comparison's v0 value is the HARNESS (budgets, blinding,
  recovery rule), stated as such in docs.
- **Concurrent-writer churn** — policy slice touches `frontier.Rank` and
  pipeline surfaces; re-cement U4/U5 against live APIs at implementation
  start.

## Sources

- `EPIC.md` M7 (experiment shape, baselines B0–B3, metric vocabulary,
  acceptance criterion) + stopping conditions (AGENTS.md).
- `corpus/README.md` — the blinded-benchmark reclassification and blinding
  discipline this plan elevates into enforced code.
- `docs/persistence.md` holdout blueprint (`holdout_set`, source links +
  problem guards, family labels) and the v15 section's holdout-REFUSAL gate.
- Live surfaces at `24cdc24` (+ plan 008 in flight): `internal/store/
  migrations.go` (gate trigger, role CHECK), `internal/pipeline/evaluation.go`
  (mode refusal), `frontier_proposal_signatures` (v17 sidecar),
  `canon.CompareWithProfile`, `docs/search-policy.md` (`--no-policy`).

## Product Contract

- CLI additions (stable `--json`): `experiment define|run|show|compare`.
- Epistemic guarantee: blinded results are never presentable as historical;
  blinding is code-audited; recovery is one deterministic rule across arms;
  budgets and stopping conditions are persisted; experiments have no side
  effects on the research state they measure.
