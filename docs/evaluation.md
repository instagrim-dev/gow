# Evaluation design (`newf` v0)

## Implemented: verifier routing (#15, M5.2, `mode='proposal'`)

M5.2 ships the toolbox `evaluate` operator for the `proposal` mode: it takes the
frontier proposals M5.1 produced (whose `result` is NULL) and routes each to the
**strongest verifier that can actually decide it**, recording a durable
`Evaluation` whose **verification strength is explicit**. The M7 holdout
experiment described further below is deferred; this section documents what
ships now.

### The one load-bearing idea

An outcome is worth exactly as much as the mechanism that produced it, and that
strength is **stored, not implied** (`AGENTS.md` verification hierarchy;
`ModelJudgment != Verification`). Every `evaluation` carries two columns beyond
the verdict:

- `verifier_kind` — *what ran*: `deterministic-check`, `counterexample-search`,
  `reproducible-computation`, `independent-evidence`, `independent-critic`, or
  `model-judgment`.
- `verification_strength` — *its position in the hierarchy*: `deterministic` >
  `reproducible` > `independent-evidence` > `independent-critic` >
  `single-model-judgment`.

A `success` from a deterministic check and a `success` from a single model share
a verdict string but **can never share a strength**. Storing an evaluation
without a strength is rejected by CHECK.

### The verifier hierarchy (`internal/verify`, pure)

Three tiers ship in v0 (the middle three `verifier_kind` values are in the CHECK
vocabulary so strengths are stable, but their real adapters are deferred behind
the same interface):

1. **`deterministic-check`** (strongest, cheapest) — reuses the M4.2 predicate
   evaluator over the per-target violation verdicts M5.1 already computed and
   persisted. It decides only the code-certain negative: if the proposed
   mechanism still *satisfies* a targeted invariant, the claimed break did not
   happen → deterministic `failure`. A confirmed break is left non-decisive here
   (necessary but not sufficient for success), deferring the positive verdict to
   the next tier.
2. **`counterexample-search`** (reproducible) — a bounded, deterministic scan of
   the proposal's **recorded nearest** known failure families for a genuine
   *refuter*. A refuter must contradict a claim about the **proposed mechanism
   itself**, not merely differ from it: a known failure family that makes the
   **same** break (also *violates* a target the proposal broke) yet still failed
   is a refuter → `failure` / `counterexample-search`. The intended structural
   difference — the proposal *violates* a target that old failure families
   *satisfy* — is exactly the signal frontier generation seeks, **never** a
   refutation. A bounded search that finds no refuter is a **negative search
   result**, not progress: it returns a non-decisive verdict (deferring
   realizability to the model tier) rather than rewarding missing or `unknown`
   comparison evidence with `partial_success` (G1).
3. **`model-judgment`** (weakest, last resort) — a provider `Verifier` (a
   deterministic `FixtureVerifier` in CI, role `'evaluate'`). Consulted only when
   no stronger tier decides, and always stamped `single-model-judgment`.

### Cheap-first, strongest-decisive routing

`Route` orders verifiers by **hierarchy strength before cost**, then runs them
until one returns a *decisive* verdict (not `unknown`/`verification_blocked`).
Ordering by strength first is the anti-laundering guard: a confident model
`success` can never preempt a deterministic check that is also able to decide,
even if the model tier declares a cheaper cost. The deciding verdict's strength
is **clamped to the verifier's registered tier** (`StrengthForKind`): a verifier
may under-report its strength but can never launder a stronger one than its
registration — a model-kind adapter returning a valid `deterministic` strength is
still recorded as `single-model-judgment` (G5). If nothing decides, the stored
verdict is `verification_blocked` stamped with the weakest tier tried — an honest
"we could not verify", never a guess.

### Pinned evaluation context

A proposal is evaluated against the **currently targetable** invariants. A cached
per-target verdict from generation time is included only when its target is still
targetable; if a targeted invariant has since become `weaken`/`falsified` it is
dropped from **both** the deterministic input and the comparison population
together (G4). An unchanged proposal therefore cannot gain a *better* evaluation
merely because a hypothesis it targeted became less credible, and its comparison
evidence cannot silently disappear while its claimed break persists. If every
claimed target is stale the context is empty and routes to a non-decisive result,
never a free `partial_success`.

### Verdict vocabulary

Exactly the EPIC.md M5.2 set, CHECK-enforced:
`failure | partial_failure | partial_success | success | unknown |
verification_blocked`.

### Failure re-enters the atlas

A `failure`/`partial_failure` evaluation writes an `evaluated_failures` marker so
the proposal's mechanism is eligible for the next `cluster build`
(`FailureSpace(t+1) ⊇ FailureSpace(t) + newly_evaluated_failures`). This is a
persisted, queryable flag surfaced by `newf evaluation failures` — not an
auto-rerun; re-clustering stays an explicit operator/policy decision.

### Persistence, provenance, lifecycle

- The proposal's `result` is populated in the **same transaction** as the
  evaluation (they can never diverge).
- Model-tier evaluations record a `provider_invocations` row (role `'evaluate'`)
  with retained request/response payloads; deterministic tiers record
  `tool_name`/`tool_version` and no provider row.
- Runs use `running → completed/failed`; re-evaluation is a new append-only
  `evaluation_run`; all evaluation rows are immutable by trigger.
- `mode='holdout'` is refused at the service boundary AND by a gate trigger
  (deferred to M7); the nullable holdout columns are retained so M7 needs no
  schema retrofit.

### CLI

```text
newf evaluate <proposal-id> --problem <id>   # route one proposal
newf evaluate --problem <id>                 # every un-evaluated proposal in the latest generation
newf evaluation list --problem <id>
newf evaluation show <evaluation-run-id>
newf evaluation failures --problem <id>      # proposals eligible for atlas re-entry
```

All commands emit stable `--json`. Human output leads with verdict AND strength,
e.g. `partial_success  [counterexample-search / reproducible]`, so a reader never
sees an outcome without its epistemic strength.

### Deliberately out of scope here

Compressing partial successes into success invariants (M6.1), mutating search
policy (M6.2), the holdout experiment and baseline arms (M7), and real external
verifier adapters. Evaluating a proposal does **not** change any invariant's
state.

---

## Primary benchmark: historical holdout prediction

v0 evaluates whether failure-space compression predicts productive frontier directions better than baselines, not whether it “solves” open problems.

## First-class workflow

1. Select a problem with known historical partial advances.
2. Define cutoff timestamp/version and hold out later mechanism family.
3. Ingest only pre-cutoff sources into atlas.
4. Normalize and cluster pre-cutoff approaches.
5. Mine candidate failure invariants and challenge them.
6. Generate frontier proposals against surviving invariants.
7. Evaluate whether proposals recover held-out mechanism family or its structural break.

Each stage is persisted in its own run/revision artifacts. `normalization_revision`
stores the training snapshot/filter that was used, and `evaluation_run` links
the evaluation stage to the selected upstream revision IDs plus the typed budget
and judge configuration used to compare baselines.

## Holdout dataset contracts

- `holdout_set` persisted artifact (referenced by `--holdout-set-id`) with:
  - problem ID
  - cutoff
  - held-out source IDs
  - held-out mechanism family label(s)
- `normalization_revision` persists:
  - the source filter / excluded holdout set used to build the training split
  - a manifest hash for the selected training evidence
  - per-evidence snapshot rows so leakage checks can compare exact contents, not just timestamps
- No held-out source/evidence may appear in normalization/clustering revisions used for generation.
- `generate` and holdout-mode `evaluate` require a passed `holdout_leakage_check` for the selected `holdout_set` + `normalization_revision` pair before proposals are scored.

## Metrics

- **Held-out mechanism-family recovery**: whether top-k proposals map to held-out family or equivalent structural break.
- **Invariant precision vs known counterexamples**: fraction of invariants surviving known valid counterexamples.
- **Mechanistic diversity**: spread across mechanism-axis values among generated proposals.
- **Normalized redundancy**: duplicate/surface-variant rate after normalization.
- **Information gain per evaluated proposal**: ordinal gain from falsification/partial-success outcomes.
- **Confidence calibration**: agreement between confidence bins and realized outcomes.
- **Synthetic-failure usefulness**: rate synthetic failures weaken/falsify invariants or improve future proposal quality.

## Baselines

Required baseline families:

1. **Undirected generation baseline**  
   Prompt model to “find a solution direction” without failure invariants.
2. **Semantic summarization baseline**  
   Summarize known approaches and generate next ideas from summary, without mechanistic invariant lifecycle.

Compare identical budget envelopes:

- same provider class where possible,
- same proposal count,
- same evaluation budget,
- same judge/evaluator process.

## Evaluation outputs

Machine (`--json`) includes:

- evaluation run metadata (cutoff, holdout IDs, baseline type, budget envelope, judge provider/config),
- proposal-level verdicts and confidence,
- typed holdout-match rows (`source_recovery` targets a held-out source ID; `family_recovery` and `structural_break` target a held-out family label, each with a match verdict),
- per-evaluation metric rows plus run-level aggregate metric rows, each with scales, comparators, and ordinal scale/version metadata where applicable,
- leakage/precondition status.

Human output includes:

- concise pass/fail against hypothesis,
- top recovered holdout links (or misses),
- which invariants were most predictive vs misleading,
- recommended next experiment.

## Falsifiable hypothesis

`newf` should beat undirected generation and semantic-summary baselines on held-out mechanism-family recovery and information gain per evaluated proposal under matched evaluation cost.
