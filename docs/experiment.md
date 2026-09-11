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
sidecar. Target representatives are the **frozen manifest**: exactly the
canonical signatures derived (via normalization provenance) from the holdout
set's REGISTERED withheld sources — the same population the leakage audit
inspects. Target material added to the quarantined problem after definition is
invisible to scoring, and the manifest (signature ids + content fingerprints)
is persisted per experiment (`experiment_targets`) and folded into the
experiment identity, so a changed target population is a **new** experiment,
never a silent recomputation. Ambiguous content classifies `unknown` — an
epistemic gap, never a coerced recovery. The rule version and the complete
comparison-profile hash are stamped on the experiment; changing either is a new
experiment identity.

## Arms and budgets

| arm | v0 status |
|---|---|
| `b0_undirected` | ships — no invariant targets, no policy. The offline deriving fixture honestly yields **zero** proposals (it derives from targets, and B0 has none); a live generator wired via `generatorFn` brainstorms freely under the same budget |
| `b1_semantic_summary` | ships — the summarize-next deriving fixture restates the dominant known failure family per family (role `'summarize-next'`); it targets no invariant and is expected to land mechanism-near the KNOWN failure, not the held-out advance |
| `b2_brainstorm` | ships — the undirected-brainstorm deriving fixture emits generic, canonically-redundant variations (role `'brainstorm'`); surface-only variation means one distinct mechanism and measurable redundancy |
| `b3_invariant_guided` | ships — the full loop (surviving targets + search policy), keyed on the dedup-stable set of proposals its generator produced this run so an unchanged atlas replays idempotently |

Every arm flows through **one** shared frontier core: distance, violation
verification, hashing, ranking, and persistence are identical across arms; only
target selection, policy, generator, and provenance role differ by arm. Each
arm's scored set is its EXPLICIT persisted membership
(`experiment_arm_proposals`: proposal id, rank, per-proposal assessment) — the
dedup-stable set of proposals *its* generator produced this run, resolved
through cross-run dedup to canonical persisted ids. Artifact deduplication and
experiment participation are different identities: an arm is never credited
with (or measured against) proposals another generation produced.

Both budgets are enforced and audited. The **proposal budget** caps the arm's
membership. The **evaluation budget** is spent in units of one proposal–target
comparison, in rank order, with actual consumption persisted per arm
(`evaluations_consumed`); a proposal the budget could not finish is recorded
`unassessed`. Hitting either budget records `budget_exhausted` (the AGENTS.md
stopping-condition vocabulary) — truncation is never silent. Per-proposal
assessments roll up as `recovered` / `decisive_no` / `unknown` / `unassessed`
counts, and the experiment conclusion respects them: `no_recovery` requires
EVERY membership proposal decisively assessed; unknown-only or budget-starved
populations conclude `inconclusive`, never a coerced negative. Metrics are
exact counts + derived ordinal bands. `held_out_family_recovery` encodes
OBSERVED recovery detections; when either compared arm is inconclusive
(unknown-only or budget-unassessed), `compare` reports the metric direction as
`incomparable` rather than treating 0 observed hits as a demonstrated negative (`held_out_family_recovery`,
`decisive_assessments`, `mechanistic_diversity`, `normalized_redundancy`;
verdict-dependent metrics land with M5.2 evaluation volume).

## External proposal arms (pilot)

`experiment run --b0-proposals-file <path> --b3-proposals-file <path>` routes
B0 and/or B3 through the untrusted proposer adapter instead of their
deterministic fixtures — the **minimum genuine comparison**: the same external
proposer answers B0 (its permitted context excludes the invariant targets) and
B3 (surviving invariants supplied), under the same predeclared budgets, and
both captured outputs enter through the production vocabulary-admission
boundary. B1/B2 stay scripted machinery controls and refuse a file.

The harness retains, per arm: the generation request (the permitted context),
the raw wire output verbatim, the admission audit
(`admission_corrected/downgraded/stripped/rejected/overflow`), and the
assessed membership. The **operator** retains the prompts that produced the
captured output — record them alongside the run in the experiment manifest.
Replaying with the same files is idempotent; a changed file is a new
assessment identity. For a cost-effectiveness claim, additionally account for
the work spent deriving and challenging B3's invariants — the harness does not
do that accounting for you.

### Pilot freeze gate

Harness development stops for the pilot when this chain demonstrably holds:

```text
one pinned, eligible split
-> non-predetermined baseline and guided proposals
-> bounded, attributable admission and assessment
-> exact retained inputs, outputs, and uncertainty
-> reproducible report
-> evidence revision can withdraw or restore downstream guidance
```

The gate is not "B3 wins" — it is "the comparison is meaningful and the
recorded result follows from the actual inputs." Findings after the freeze are
classified by consequence: fix-before-interpreting (claim scope, target
leakage, evidence misattribution, lost revisions, unjustified verdicts),
fix-alongside (audit ergonomics, reporting, resource accounting), defer
(generalized machinery whose absence does not change this experiment's
conclusion).

## Persistence (migrations v21 + v22) and side-effect freedom

`holdout_sets` (+ withheld-source links with problem-guard triggers,
`holdout_source_dating`), `leakage_checks`, `experiment_runs` (idempotent on an
identity hash over split + rule + profile version **and hash** + budgets +
arms + the frozen target-manifest fingerprints + the per-arm membership set),
`experiment_arms` (with assessment counts and consumed evaluations),
`experiment_arm_proposals` (explicit membership), `experiment_targets` (frozen
manifest), `experiment_metrics` — all immutable.
Experiments **measure** the research state and never mutate it: no invariant
lifecycle change, no policy write, no train-atlas write.

## CLI

```text
newf experiment define --problem <train> --target-problem <target> [--mode blinded|historical] [--cutoff <t>] [--name <n>]
newf experiment run [--problem <train> | --holdout-set <id>] [--arms b0_undirected,b1_semantic_summary,b2_brainstorm,b3_invariant_guided] [--proposal-budget n] [--evaluation-budget n]
newf experiment show [experiment-id] [--problem <id>]
newf experiment list --problem <id>
newf experiment compare [experiment-id] [--problem <id>] [--baseline <arm>] [--treatment <arm>]
```

All support `--json`; runs use `running → completed/failed`. Fully offline.

`compare` reports a **within-experiment**, apples-to-apples arm delta (default
`b0_undirected` vs `b3_invariant_guided`): both arms share the same holdout
split, recovery rule, comparison profile, and shared budget. It reports exact
counts, an ordinal direction (by exact ratio when comparable, else band), and
the recovery delta only — a single deterministic split supports **no**
statistical-significance claim, and the interpretation string never implies one.
Each arm carries an explicit recovery **status** — `recovered`, `no_recovery`,
or `inconclusive` — computed with the same F5 gate the run-level conclusion
uses: a non-recovery is a decisive negative only when EVERY membership proposal
was decisively assessed. An arm with unknown/unassessed proposals (or an empty
own generation) is `inconclusive` and is **never** coerced into the negative
side of the delta (`neither`/`baseline-only` require a decisive `no_recovery`
on the arm being called out; otherwise the delta reads `*-inconclusive` /
`inconclusive`).

Two arms that derive the same mechanism legitimately share a deduped
`proposal_id` (frontier proposals dedup on `(problem_id, proposal_hash)`). Arm
isolation therefore does not rely on `proposal_id` uniqueness: each arm computes
its OWN arm-local `member_rank` and assessment, and experiment identity +
membership key on `(arm, proposal_hash)` — the arm's own derivation — so a
replay whose proposals all dedup stays idempotent and no arm reuses another
arm's rank.

## The v0 value is the harness

All four arms now execute offline against deterministic fixtures: B0 is
honestly empty, B1/B2 produce baseline-shaped proposals (summary restatement /
generic redundancy), and B3 runs the directed loop. The v0 claim is still NOT
"B3 beats baselines" — the shipped corpus is a single synthetic split with
fixture generators, so no arm outcome is a scientific result. The claim is that
the **harness** exists: leakage-audited blinding, equal budgets, one code-owned
recovery rule across every arm, a non-inflating compare, and mode-honest
immutable artifacts. The first live-model campaign and the first dated corpus
drop into it without schema or metric retrofit.
