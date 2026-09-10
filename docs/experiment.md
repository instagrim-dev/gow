# Blinded-benchmark validation experiment (M7 v0)

`newf experiment` composes the whole loop into a measured experiment:

> Given only the `train/` failure-space, does invariant-guided frontier
> generation recover the **structural move** of a blinded `target/` family
> better than undirected baselines?

## Two modes, disjoint epistemic claims

**BlindedRecovery != HistoricalPrediction.** The shipped corpus is a
**synthetic blinded benchmark** (`corpus/README.md`), not a historical holdout
— an earlier corpus version falsely claimed otherwise and was corrected. The
experiment layer makes that discipline structural:

| mode | claim | conclusion vocabulary (CHECK-enforced) | executable? |
|---|---|---|---|
| `blinded` | structural-move recovery under enforced blinding; **no chronological claim** | `structural_recovery` / `no_recovery` / `inconclusive` | yes |
| `historical` | prediction of a historically later advance | `predicts_later_advance` / `fails_to_predict` / `inconclusive` | **refused** until EVERY withheld source carries dated evidence (`holdout_source_dating`) |

Every view carries a `mode_disclaimer`; a blinded experiment structurally
cannot record a historical conclusion. Lifting `historical` needs **data**
(auditable dated sources), not code.

## Blinding is code-audited (KTD-2)

`experiment define --problem <train> --target-problem <target>` binds a TRAIN
problem to a QUARANTINED target problem (a separate problem id, enforced by
CHECK) whose every source is withheld. `experiment run` first executes the
**leakage check**: a content-identity audit (sha256) proving no train-problem
snapshot, normalization revision, or mechanism signature derives from withheld
bytes. The persisted `leakage_checks` row is the completion key — a failed
audit fails the run and persists **no** experiment; it is also the condition
that lifts the M5.2 `mode='holdout'` evaluation gate (the v15 trigger's
reserved lift, replaced in v21).

## One recovery rule for every arm (KTD-3)

`recovery-rule/v1`: a proposal recovers the target family iff
`canon.CompareWithProfile(proposal_content, target_representative)` classifies
as `mechanism-near` (incl. `surface-distinct+mechanism-near`) under the pinned
profile. Proposal content comes from the v17 `frontier_proposal_signatures`
sidecar; target signatures are rehydrated with full provenance from the
quarantined problem. Ambiguous content classifies `unknown` — an epistemic gap,
never a coerced recovery. The rule version is stamped on the experiment;
changing the rule is a new experiment identity.

## Arms and budgets

| arm | v0 status |
|---|---|
| `b0_undirected` | ships — no invariant targets, no policy. The offline fixture honestly yields **zero** proposals (a fixture cannot brainstorm); a live model generates freely under the same budget |
| `b1_semantic_summary`, `b2_brainstorm` | schema-supported, **execution-refused** until their baseline provider roles gain deriving fixtures (v21 already admits roles `'summarize-next'`/`'brainstorm'`) |
| `b3_invariant_guided` | ships — the full loop, keyed on the problem's persisted proposal set so an unchanged atlas replays idempotently |

All arms share one persisted proposal/evaluation budget; hitting it records
`budget_exhausted` (the AGENTS.md stopping-condition vocabulary) — truncation
is never silent. Metrics are exact counts + derived ordinal bands
(`held_out_family_recovery`, `mechanistic_diversity`, `normalized_redundancy`;
verdict-dependent metrics land with M5.2 evaluation volume).

## Persistence (migration v21) and side-effect freedom

`holdout_sets` (+ withheld-source links with problem-guard triggers,
`holdout_source_dating`), `leakage_checks`, `experiment_runs` (idempotent on an
identity hash over split + rule + profile + budgets + arms + the persisted
proposal set), `experiment_arms`, `experiment_metrics` — all immutable.
Experiments **measure** the research state and never mutate it: no invariant
lifecycle change, no policy write, no train-atlas write.

## CLI

```text
newf experiment define --problem <train> --target-problem <target> [--mode blinded|historical] [--cutoff <t>] [--name <n>]
newf experiment run [--problem <train> | --holdout-set <id>] [--arms b0_undirected,b3_invariant_guided] [--proposal-budget n]
newf experiment show [experiment-id] [--problem <id>]
newf experiment list --problem <id>
```

All support `--json`; runs use `running → completed/failed`. Fully offline.

## The v0 value is the harness

Offline fixtures make B0 trivially empty and B1/B2 unavailable — the v0 claim
is NOT "B3 beats baselines." It is that the **harness** exists: leakage-audited
blinding, equal budgets, one code-owned recovery rule, mode-honest immutable
artifacts. The first live-model campaign and the first dated corpus drop into
it without schema or metric retrofit.
