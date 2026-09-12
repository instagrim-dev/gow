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
- `verification_subject` (v38, issue #21) — *what object the verdict is
  about*: `annotation` (the persisted normalized signature/claims),
  `realization` (the described mechanism is constructible as described), or
  `domain-goal` (the mechanism achieves the original goal). A deterministic
  check of a predicate over a signature establishes something about that
  signature — not that it faithfully describes a realizable mechanism, nor
  that the mechanism satisfies the domain goal; without this axis a truthful
  `deterministic` label can be read as certifying the wrong object. The
  subject is part of a verifier's *registration* (`Verifier.Subject()`), and
  `Route` stamps it from the deciding tier — a verifier cannot relabel its
  subject per verdict. The in-tree deterministic tiers certify the
  **annotation**; the model tier judges the **domain goal**. `NULL` only for
  pre-v38 history (never backfilled retroactively); new writes require it.

A `success` from a deterministic check and a `success` from a single model share
a verdict string but **can never share a strength**. Storing an evaluation
without a strength is rejected by CHECK, and without a subject by the writer.

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
2. **`counterexample-search`** (reproducible tier, **non-decisive by
   construction**) — a bounded, deterministic scan of the proposal's **recorded
   nearest** known failure families. The persisted verification context carries
   only per-target **predicate verdicts**, not a candidate-specific refutation
   witness, so this tier **cannot soundly decide a proposal failure** and never
   returns `failure` (H3). Two facts it must never conflate with refutation:
   the intended structural difference (the proposal *violates* a target that old
   failure families *satisfy*) is the signal frontier generation seeks, not a
   refutation; and a known failure family that *also* violates a broken target
   merely **shares a predicate bit** — a shared predicate verdict does not
   establish a shared mechanism or that the known failure transfers to the
   proposal (two different global constructions can both violate "local reasoning
   only"; one failing does not make the other fail). That signal is at most
   **"break previously observed"** (a novelty/sufficiency note), recorded but not
   decided. This tier confirms the break is real and otherwise returns a
   non-decisive verdict, deferring realizability to the model tier and never
   rewarding missing/`unknown` comparison evidence with `partial_success`. A
   decisive proposal failure awaits a future mechanism-level refuter that
   contradicts an **explicit claim about the proposal itself**.
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
targetable. If **any** targeted invariant has since become `weaken`/`falsified`,
the proposal is treated as **requiring reassessment**: it is recorded as
`verification_blocked` (deterministic tier) and is **not routed** at all (H5).
Erasing a stale target and then routing the reduced context would silently change
the question being evaluated and could let a decisive model adapter award an
unchanged proposal a *better* verdict merely because a hypothesis it targeted lost
credibility. Pinning the assessment to the exact targets — and refusing to answer
a different question in place of the original — is the guarantee: target
invalidation can never improve an unchanged proposal's evaluation. A fresh
challenge assessment may explicitly select newer evidence while retaining the
original discovery lineage.

### Verdict vocabulary

Exactly the EPIC.md M5.2 set, CHECK-enforced:
`failure | partial_failure | partial_success | success | unknown |
verification_blocked`.

### Failure re-enters the atlas — via explicit admission (v35/S2)

A `failure`/`partial_failure` evaluation writes an `evaluated_failures` marker so
the proposal's mechanism is eligible for the next `cluster build`
(`FailureSpace(t+1) ⊇ FailureSpace(t) + newly_evaluated_failures`). This is a
persisted, queryable flag surfaced by `newf evaluation failures` — not an
auto-rerun; re-clustering stays an explicit operator/policy decision.

The marker alone admits nothing. `ListSignaturesForProblem` — the population
`cluster build` and invariant mining consume — never reads it. The bridge is
`newf evidence admit` (structural review finding S2): a typed, immutable
decision per evaluated failure, recorded in the `evidence_admissions` ledger.
Observation kinds and their rules:

| Observation kind | What it is | Rule |
|---|---|---|
| `structural-claim-failure` | a `deterministic-check` failure: the description failed its own claimed break | never admissible (narrows the description space, not the observed-mechanism space) |
| `domain-checked-failure` | a `deterministic`/`reproducible` strength failure of an attempt | admitted by rule; `independent-evidence` strength requires operator attestation |
| `model-judged-failure` | a `model-judgment`/`independent-critic` strength failure | operator attestation only (`--evaluation --attest --note`), permanently labeled model-judged |

Admission materializes the **exact assessed signature content** (matched by the
evaluation's `signature_content_hash`; a verdict without persisted assessed
content is withheld — there is nothing admissible to materialize): a source
snapshot whose bytes are the assessed JSON, a new revision of the stable
approach `frontier-proposal:<id>`, and the signature persisted verbatim with
only the outcome set from the verdict (`inferred` provenance — tool-derived,
not source-explicit). Re-admitting a re-assessed proposal revises the same
logical approach, so the current-heads population (S4) keeps one current
interpretation instead of double-counting. Decisions are made once: an
evaluation is withheld at most once and admitted at most once; a
withheld-then-attested evaluation keeps both rows, so supersession is visible.

```text
newf evidence admit --problem <id>                # batch rule pass
newf evidence admit --problem <id> --evaluation <evl_> --attest --note "…"
newf evidence list --problem <id>                 # the admission ledger
```

### Persistence, provenance, lifecycle

- The proposal's `result` is populated in the **same transaction** as the
  evaluation (they can never diverge).
- Model-tier evaluations record a `provider_invocations` row (role `'evaluate'`)
  with retained request/response payloads; deterministic tiers record
  `tool_name`/`tool_version` and no provider row.
- A verifier tier that is **operationally unreachable** (wraps
  `verify.ErrVerifierUnavailable`) is recorded and skipped, not treated as an
  abstention or a run failure; if no reachable tier decides, the evaluation
  lands as `verification_blocked` with the unreachable tiers named in its
  notes. Any other verifier error still aborts the run.
- Runs use `running → completed/failed`; re-evaluation is a new append-only
  `evaluation_run`; all evaluation rows are immutable by trigger.
- `mode='holdout'` is refused at the service boundary AND by a gate trigger
  (deferred to M7); the nullable holdout columns are retained so M7 needs no
  schema retrofit.

### CLI

```text
newf evaluate <proposal-id> --problem <id>   # route one proposal
newf evaluate --problem <id>                 # every un-evaluated proposal in the latest generation
newf evaluate --problem <id> --generation <id>  # pin the assessment context to an explicit generation occurrence
newf evaluation list --problem <id>
newf evaluation show <evaluation-run-id>
newf evaluation failures --problem <id>      # proposals eligible for atlas re-entry
```

`--generation` pins the assessment context to a specific frontier generation's
occurrence membership (historical replay or a specific revised occurrence);
the default is the latest occurrence.

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
