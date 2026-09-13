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
(auditable dated sources), not new machinery: the write path is
`experiment date-source`, and the epistemic burden of producing real dated
evidence stays with the operator. The gate was first lifted and a historical
run executed on 2026-09-12 for the P-vs-NP corpus
(`corpus/experiments/pvnp-holdout/` — see its `RESULT.md`).

## The historical dating gate (`experiment date-source`)

A historical run is refused until **every** withheld source carries one
dated-evidence row in `holdout_source_dating`:

```text
newf experiment date-source --holdout-set <id> --source <id> \
  --dated-at <RFC3339> --evidence <URL/DOI/archive-ref> [--provenance <note>]
```

The contract, enforced at record time:

- **Historical sets only.** A blinded set makes no chronological claim, so
  there is nothing to date; the verb refuses.
- **Membership.** The source must be a withheld member of the holdout set.
- **Chronology.** `dated_at` must parse as RFC3339 and be **strictly after**
  the set's cutoff — a withheld source dated at or before the cutoff cannot
  be a *historically later* advance, and admitting it would let a run
  "predict" something that predates its own training boundary.
- **Auditability.** An evidence locator is required; a dating claim without
  auditable evidence is just an assertion.
- **Immutability.** Rows are immutable (trigger + verb): an identical
  re-record is an idempotent no-op, a contradicting attestation is refused —
  a changed dating claim needs a new holdout set, never a silent update.

Once the gate opens, `experiment run` executes the same leakage-audited,
equal-budget loop, and the conclusion is recorded through the **historical**
vocabulary (`predicts_later_advance` / `fails_to_predict` / `inconclusive`) —
the mapping from the one code-owned recovery outcome forks on mode
(`experimentConclusion`), and the schema CHECK aborts any cross-vocabulary
write. What the gate does **not** do: it does not verify that the dating
evidence is true. `evidence_locator` makes the claim auditable; auditing it
is operator work, recorded as provenance, not code-certified truth.

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
A supplied file must name a selected external arm (B0/B3): a file for an
unselected or scripted arm is rejected in preflight, before any run record
exists — an explicitly supplied input either participates or is rejected,
never silently ignored.

**Execution identity vs assessment identity (v32).** Experiment identity keys
on the ASSESSED structure (ordered proposal identities/content, targets,
budgets, arms) — so a changed capture whose assessed structure is unchanged
(e.g. only prose that `ProposalHash` excludes) correctly REUSES the experiment
artifact. It is still a different execution of a different capture: every
execution appends `experiment_executions` rows recording, per arm, its own
run id, the experiment it selected (created or reused), the generation it
actually produced (whose invocation retains the new raw payload), and the
sha256 of the consumed proposals file. Pilot manifests should cite those
execution rows — run id, generation id, file hash — not "latest experiment".
For a cost-effectiveness claim, additionally account for the work spent
deriving and challenging B3's invariants — the harness does not do that
accounting for you.

The harness cannot establish from the files alone that both captures came
from the same model/configuration, that B0's context genuinely lacked the
invariants when the output was produced, or that the proposer was unaware of
the withheld target. Those are PILOT PROTOCOL obligations: retain actual
model identity, configuration, prompts, permitted context, and capture hashes
alongside the run.

`newf experiment readiness --problem <id>` computes the MECHANICAL half of
the readiness decision, read-only: eligible failure cohort (vs the mining
support threshold), decisive-axis resolution on the train population (under
`--vocab-version`, default `mechanism/v1`), completeness admissions
(informational — zero accepted means absence-based verification degrades to
unknown, honestly), surviving invariants, an assessable withheld target, and
**recovery reachability**: a target with any unresolved claim on a decisive
field cannot receive a positive match from ANY proposal under
`recovery-rule/v1` (its own self-comparison classifies unknown), so the
check blocks until the target's stated labels are canonicalized in a pinned
vocabulary revision — never by ignoring unresolved fields, and never by
adding interpretation claims to the target. The evaluator-side calibration
against the actual frozen target (self-match reachable / known-match
recovered / known-different decisively non-recovering / under-represented
unknown) is pinned in `internal/pipeline/recovery_calibration_integration_test.go`.
The non-mechanical half is emitted as operator attestations, never assumed.
The pilot template lives at `corpus/experiments/pilot-001/PROTOCOL.md`.

`newf experiment validate-proposals --file <capture> [--problem <id>]`
preflights a captured proposals file through the EXACT importer decode path
(`provider.ParseWireProposals` — the same implementation `experiment run`
consumes), so "validated" can never mean anything weaker than what the
importer enforces. With `--problem`, the problem's surviving invariants are
the permitted targets (B3 semantics); without it no targets are permitted
(B0 semantics). Read-only. A capture must pass this preflight before being
declared importable; sealing checks alone (JSON parse + schema_version) are
not wire validation — the pilot-003 B3 capture passed sealing and failed
the importer on field nesting.

Assessment runs under the corrected `classify/v3` profile, whose
missing-data contract is: **recorded-set similarity is always observable,
but a decisive judgment requires enough information that unrecorded members
cannot overturn it — in BOTH directions.** Concretely — one side empty without a `complete`
justification is an epistemic gap (never decisive negative evidence); two
empty unjustified sides are an epistemic gap (mutual silence is never
POSITIVE evidence either — an unobserved pair must not read as recovered);
recorded AGREEMENT and recorded DISAGREEMENT are BOTH decisive only when
both sides are justified complete: a partial side's unrecorded members
could contain exactly the missing elements (overturning a mismatch) or
diverge entirely (overturning an apparent match — {X} vs {X} recorded can
complete to similarity 1/5, below the near threshold). "Both descriptions
affirm X" is supported; "their sets are sufficiently similar" is not, so
recorded overlap survives only as a diagnostic on withdrawn axes.
Justified-complete empty fields participate decisively in both directions.
Practical consequence: against a corpus with no completeness declarations
the automatic layer abstains in BOTH directions, and untrusted proposals —
whose completeness admission strips — can earn neither automatic verdict;
their recovery question is judgment-level until a justified-completeness
pathway exists. Each contract participates in the profile hash
(`classify/v1` → v2 → v3 are pairwise distinct), so pinned results keep
their identity; a reassessment under a corrected rule is a corrected
assessment with its own experiment identity, never a silent substitution.

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
newf experiment date-source --holdout-set <id> --source <id> --dated-at <t> --evidence <locator> [--provenance <p>]
newf experiment run [--problem <train> | --holdout-set <id>] [--arms b0_undirected,b1_semantic_summary,b2_brainstorm,b3_invariant_guided] [--proposal-budget n] [--evaluation-budget n] [--b0-proposals-file <path>] [--b3-proposals-file <path>]
newf experiment readiness --problem <id> [--holdout-set <id>] [--min-support n] [--vocab-version <v>]
newf experiment validate-proposals --file <capture> [--problem <id>]
newf experiment show [experiment-id] [--problem <id>]
newf experiment list --problem <id>
newf experiment compare [experiment-id] [--problem <id>] [--baseline <arm>] [--treatment <arm>]
newf evaluate --problem <id> [--generation <id>]    # pin the assessment to an explicit generation occurrence
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
