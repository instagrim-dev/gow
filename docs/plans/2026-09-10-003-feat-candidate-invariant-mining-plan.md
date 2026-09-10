---
artifact_contract: ce-unified-plan/v1
artifact_readiness: implementation-ready
execution: code
product_contract_source: ce-plan-bootstrap
origin: GitHub issue #12 (instagrim-dev/newf) — "Mine candidate failure invariants from the failure-space" (to be filed)
epic: EPIC.md milestone M4.2 — Mine candidate failure invariants
depends_on: >-
  GitHub issue #11 (mechanism clustering + first FailureSpace artifact). Minimum
  substrate commit `151df63` ("fix(cluster): harden cluster-run identity, family
  coherence, and mixed-outcome semantics (#11)"), schema **v10**. That commit is
  what makes M4.2 viable: cluster identity now includes the signature population
  (`InputSetHash`), clustering enforces pairwise coherence rather than blind
  transitive closure, and mixed-outcome families are represented explicitly
  (`OutcomeClass` + `OutcomeMixed`, aggregate class `domain.OutcomeMixed`).
created: 2026-09-10
plan_type: feat
---

# feat: Mine candidate failure invariants from the FailureSpace

## Summary

Implement the next epic slice (EPIC.md **M4.2 — Mine candidate failure
invariants**): compress a problem's **FailureSpace** (the versioned, cluster-
partitioned artifact from #11) into explicit, typed **`CandidateInvariant`**
records that name conserved failure structure across the distinct mechanism
families in a problem — retaining support families, counterexamples, abstraction
level, epistemic composition, and an explicit `association_status`
(`recurring | discriminative | candidate_obstruction | unknown`).

This is the toolbox `compress` operator (`docs/toolbox-dsl.md`):

```text
compress  Set -> CandidateInvariant[]
compress failures
  seek conserved_structure
  across independent_clusters
```

It is the **first slice whose core operation is model-native and fuzzy**, not
deterministic. But "the model proposes, code computes support" is only *literally
true* if the invariant is a **machine-evaluable typed predicate**, not English
prose. Code cannot deterministically evaluate `"failed methods preserve
residue-class locality"`. The provider must therefore emit a typed predicate AST
over the same canonical fields #9 already fingerprints; the prose statement is a
**human rendering of the predicate, not its executable definition**:

```yaml
predicate:
  schema: invariant-predicate/v1
  all:
    - field: preserves
      op: contains
      canonical_id: domain.number_theory.property.residue_locality
    - field: locality           # posture axis
      op: in
      values: [local, mixed]
statement: "failed methods remain confined to residue-local reasoning"  # human render
```

The deterministic evaluator is then a real, code-owned function:

```text
P(signature) -> { satisfies | violates | unknown }
```

`unknown` when the predicate touches a field the signature left
ambiguous/`Incomparable` (ambiguity stays visible, never coerced). This is the
difference between building an invariant miner and building a very organized
fortune cookie.

### Epistemic chain (this slice does not claim upstream "truth")

Nothing before M4.2 is "truth"; it is deterministic derived structure *over a
model's interpretation*:

```text
SourceSnapshot                observed fact
    |
    v
NormalizationInterpretation   provider judgment (#7)
    |
    v
canonicalization / clustering deterministic derived structure (#9/#11)
    |
    v
CandidateInvariant            higher-order provider hypothesis (this slice)
```

Consequently the deterministic support an invariant earns **retains the
epistemic status of the underlying matched claims**: three families supported
entirely by `inferred` claims are *not* equivalent to three backed by `explicit`
source claims. Support aggregation carries those per-claim statuses forward.

The governing constraint is epistemic, not algorithmic:

> A mined invariant is a **hypothesis about conserved failure structure, not a
> discovered law** (`AGENTS.md`). It enters at `proposed`, carries full
> provenance and an explicit `association_status`
> (`recurring | discriminative | candidate_obstruction | unknown`), and is
> **never** promoted toward `established` by this slice. Frequency + outcome
> discrimination can establish a *recurring, discriminative association* — it
> cannot establish even a provisional obstruction merely because code counts N
> families. `candidate_obstruction` is a **model hypothesis flag**, not a
> deterministic promotion; M4.3 earns the right to strengthen it through
> challenge/falsification. The project exists because humans have spent centuries
> mistaking conserved correlations for reasons.

**Invariant identity is semantic, not textual.** A `CandidateInvariant`'s logical
identity is the **canonicalized predicate AST fingerprint**; the prose statement
lives on a revision and may re-phrase freely. Otherwise "preserves residue
locality" / "retains residue-class locality" / "remains confined to
residue-local reasoning" would become three invariants because a model
paraphrased itself.

The exit condition (EPIC.md M4.2): candidate invariants are **explicit typed
artifacts, not prose buried in a model response** — each an evaluable predicate,
keyed to the FailureSpace revision it was mined from, and linked to the distinct
mechanism families that support it and the counterexamples that bound it.

This slice **consumes** the #11 substrate (`failure_spaces`,
`failure_space_axes`, `cluster_runs` incl. `InputSetHash`, `mechanism_clusters`
incl. `OutcomeClass`/`OutcomeMixed`, `cluster_members`, `signature_outcomes`) and
#9's canonical signatures. It adds an **`Invariant`-miner provider role** (with a
deterministic fixture), a new revisioned persistence layer (migration **v11**),
pipeline services, and CLI surface. It runs entirely offline in CI.

---

## Problem Frame

After #11, `newf` can say *how many mechanistically distinct failure families a
problem has, which approaches are redundant, and which axes are under-sampled*,
and can emit a versioned `FailureSpace` revision. What it **cannot** yet do is
answer the question the failure-invariant loop actually turns on:

- **What structure is conserved across the *distinct mechanism families*?**
  i.e. which mechanistic properties (a preserved invariant, an operator family,
  a posture, a boundary relation) recur across families #11 marked
  mechanistically distinct / non-redundant, and therefore look like a real
  regularity rather than one repeatedly-published idea?

**Terminology discipline (do not overclaim independence).** #11 gives
**mechanistically distinct / non-redundant families under a particular
`ComparisonProfile`** — not statistical, historical, intellectual, or causal
independence. Two distinct clusters can still share a paper, author program, or
assumption lineage. This slice therefore counts a v0 quantity named
`distinct_mechanism_families` (a.k.a. `nonredundant_support_units`), **never**
"independent families". Source/program-independence is a *later* axis, not
claimed here.

The distinction between "5 approaches all do X" and "5 *distinct families* all
do X" is exactly what #11's redundancy/family structure feeds: the former can be
publication bias; the latter is the first weak signal of conserved failure
structure. This slice turns that structure into typed, evaluable, challengeable
artifacts that the M4.3 challenge lifecycle and M5.1 frontier generator consume.

**Failure support vs. success contrast are separate axes** (a failure invariant
that "spans ≥2 outcome classes" is conceptually muddy). Define instead:

```text
FailureCoverage(I)  = (failure families satisfying I) / (eligible failure families)
Contrast(I)         = how differently I behaves in partial-success/success families
```

- **Failure / partial-failure** families provide **support**.
- **Partial-success / success** families provide **contrast or counterevidence**.
- **Mixed** families (which #11 now preserves via `OutcomeMixed`) are evaluated
  **member-wise**, or marked **ambiguous** — never shoved onto either side.

Today the repository has: problems/runs (v1), immutable sources/snapshots (v2),
normalized approaches (v3), canonicalization/comparison (v4/v5/v6), and the
clustering + first FailureSpace layer with the #11 hardening consolidated through
schema **v10** (cluster-run identity incl. `InputSetHash`, family coherence,
mixed-outcome persistence). This plan builds the candidate-invariant layer **on
top of v10** as migration **v11**.

### Governing constraints (from `AGENTS.md`, `docs/abstraction-safety.md`, EPIC.md M4.2, `docs/toolbox-dsl.md`)

- **Invariant = machine-evaluable predicate.** The durable invariant identity is
  a typed predicate AST (`invariant-predicate/v1`) over canonical fields; prose
  is a human rendering. `P(signature) -> {satisfies|violates|unknown}` is a pure,
  code-owned function. A model that emits only prose is rejected by the schema.
- **No silent epistemic promotion.** A mined invariant is created `proposed` and
  stays `proposed` in this slice; **no** transition machinery ships here (that is
  M4.3). The DB `CHECK` enforces `proposed`.
- **Upstream is interpretation, not truth.** Support retains the epistemic status
  (`explicit`/`inferred`/`ambiguous`/`unknown`/`unsupported`) of the underlying
  matched claims; aggregation reports the support's epistemic composition and
  never launders `inferred` support into `explicit`.
- **Association, not causation.** Every candidate carries
  `association_status ∈ {recurring, discriminative, candidate_obstruction,
  unknown}`. Code may deterministically assign `recurring`/`discriminative` from
  measured coverage/contrast; `candidate_obstruction` is a **model hypothesis
  flag** the code records but does not itself certify; `established` is
  unreachable here.
- **Distinct families, not "independent".** Support counts
  `distinct_mechanism_families` under the pinned `ComparisonProfile`, discounting
  #11-redundant members. No independence claim beyond mechanistic non-redundancy.
- **Failure support vs. success contrast are separate.** `FailureCoverage` over
  failure/partial-failure families; `Contrast` over partial-success/success;
  mixed families evaluated member-wise or ambiguous.
- **Provenance + replay + run lifecycle.** Each mining run is created `running`,
  then `finalizeRun`→`completed` or `failRun`→`failed` (the current run
  lifecycle); the miner invocation + raw payloads are persisted for replay.
- **Abstraction safety.** A candidate states an explicit `abstraction_level` and
  is groundable back to concrete member signatures via stored support links
  (round-trip *prediction* is M4.3; the links are persisted here).
- **Deterministic fixtures / offline CI.** The default miner is a deterministic
  fixture; no network/model call in CI.
- **Additive, immutable persistence.** New tables in migration **v11** with
  `RAISE(ABORT, ...)` triggers; revisioned append-only; re-mining a FailureSpace
  makes a new revision, never a rewrite. No edits to v1–v10 **except** the
  `provider_invocations.role` constraint (see the migration note below), which
  currently permits only `role = 'normalize'`.

---

## Scope Boundaries

### In scope

- A **typed predicate contract** `invariant-predicate/v1` (`internal/invariant`,
  pure): an AST over canonical fields (set fields `preserves`/`operators`/…,
  posture axes, boundary relation, outcome) with a small op set
  (`contains`/`in`/`equals`/`all`/`any`/`not`), a **canonical serialization** and
  **predicate fingerprint** (semantic identity, KTD-5/change 7), and a validator
  (rejects unknown fields/ops/canonical IDs).
- A **deterministic evaluator** `Evaluate(predicate, signature) ->
  {satisfies|violates|unknown}` (pure) — `unknown` when the predicate reads a
  field the signature left ambiguous/`Incomparable`. This is the code-owned core
  that makes "code computes support" literally true.
- An **`Invariant` miner provider role** (`internal/provider`) with a
  deterministic fixture (`FixtureInvariantMiner`) that, given a compact
  `MiningRequest` (per family: resolved canonical content + outcome + redundancy
  flag), proposes candidates as **`{predicate, statement, abstraction_level}`** —
  the predicate is the executable definition; the statement is a human render.
- **Deterministic support + contrast computation** (`internal/invariant`, pure):
  evaluate each candidate's predicate against every eligible signature; aggregate
  **member-wise → family** (`supports|violates|unknown`, mixed families
  member-wise); compute `FailureCoverage` (over failure/partial-failure families)
  and `Contrast` (over partial-success/success families) as **separate** axes;
  carry the **epistemic composition** of matched claims (counts of
  explicit/inferred/…); count `distinct_mechanism_families` discounting
  #11-redundant members.
- **Association status (code-assigned where deterministic):** `recurring` /
  `discriminative` derived from measured coverage + contrast; `candidate_obstruction`
  recorded **only** as a model-proposed flag (never code-certified); `unknown`
  when a decisive field is ambiguous across families. Never `established`.
- **Revisioned persistence (migration v11):** immutable, version-keyed
  `invariant_revisions`, `candidate_invariants` (semantic identity =
  `predicate_fingerprint`; `initial_state = 'proposed'` by `CHECK`),
  `invariant_predicates` (the stored AST), `invariant_family_evaluations`
  (per-family verdict + epistemic composition), `invariant_counterexamples`, plus
  the **`provider_invocations.role` generalization** (see migration note) and the
  `provider_invocations` link. All with immutability triggers.
- **Run lifecycle integration:** `invariants mine` creates its run `running`,
  then `finalizeRun`→`completed` / `failRun`→`failed`; `run show` reflects the
  real outcome.
- **CLI (all `--json`):** `newf invariants mine --failure-space <id>`,
  `newf invariant list --problem <id>`, `newf invariant show <id>`.
- **Deterministic fixture tests (offline):** predicate validation +
  fingerprint order-independence; evaluator truth table incl. `unknown`; support
  computed from data (not trusted from the model); distinct-family counting
  honors #11 redundancy; failure-coverage vs. success-contrast separation;
  mixed-family member-wise handling; epistemic-composition carry-through;
  `association_status` assignment; `proposed`-only state; provenance + run-status
  round-trip; re-mine → new revision (never rewrite); CLI contract.
- **Docs:** `docs/invariant-mining.md`, README section, EPIC.md **M4.2** status
  note, and `docs/persistence.md` v11 update.

### Deferred to Follow-Up Work

- **Invariant challenge/falsification (EPIC.md M4.3, next slice):** all state
  transitions beyond `proposed` (`challenged`/`surviving`/`weakened`/`split`/
  `merged`/`falsified`/`established`), the `invariant_challenge` /
  `invariant_state_transition` tables + the seq-counter trigger from the
  `docs/persistence.md` blueprint, and synthetic-counterexample construction.
  This slice ships **only** the `proposed` artifact + its deterministic support.
- **Abstraction split/merge** of candidates (M4.3 `abstract`/`ground` loop).
- **Real (non-fixture) provider adapter.** The interface + fixture ship here; a
  vendor adapter is a separate, provider-only change.

### Out of scope (do not implement)

- Frontier generation (M5.1), evaluation/verifier routing (M5.2), success
  compression (M6), search-policy mutation, historical-holdout scoring.
- Any promotion toward `established`, or any state other than `proposed`.
- Embeddings / vector search / learned similarity in the support path (support
  is computed from #9 canonical signatures deterministically).
- A scalar "invariant quality" score.

---

## Assumptions

1. **#11 hardening is landed** (schema **v10**, minimum commit `151df63`).
   `internal/store` exposes `GetFailureSpace`, `ListClusterRuns`/`GetClusterRun`
   (cluster records carry `InputSetHash`, and each cluster carries `OutcomeClass`
   + `OutcomeMixed`), cluster members, and per-signature projections;
   `internal/canon` exposes `MechanismSignature` + `Evaluate`-able resolved
   fields. `domain.OutcomeMixed` is an **aggregate-only** class (never on a
   single signature) with `Cluster.PrimaryOutcome()`; `OutcomeClass.ValidFamilyOutcome()`
   admits it.
2. **The FailureSpace is the mining input**, pinned to one `(problem,
   cluster_run)` and thus one `(schema_version, vocabulary_version,
   profile_version, input_set_hash)` tuple. Candidates are keyed to the
   FailureSpace revision they were mined from.
3. **Distinct-family counting = #11 non-redundant family structure** under the
   pinned `ComparisonProfile`; redundant members do not inflate
   `distinct_mechanism_families`. This is *not* an independence claim (change 4).
4. **New ID classes** follow the ULID+prefix convention in
   `internal/domain/id.go`: add `InvariantRevisionIDPrefix = "ivr_"`,
   `CandidateInvariantIDPrefix = "inv_"` with `New*`/`Validate*` and `id_test.go`
   round-trip + cross-class rejection tests.
5. **Provider provenance + the role constraint.** The miner invocation is
   recorded in `provider_invocations`, whose `role` column **currently has
   `CHECK (role IN ('normalize'))`** (`migrations.go:119`). Migration v11 must
   therefore **generalize that constraint** (see KTD-8 / U5 for the two options).
6. **Run lifecycle** is the current create-`running` → `finalizeRun`/`failRun`
   contract (`app.go:354/370`); `run show` reflects real outcome. `mine` uses it.
7. **Offline CI:** the default miner is `FixtureInvariantMiner`; no network/model.

If any surface differs at implementation time (renamed store/canon API, changed
run lifecycle, or a since-relaxed role constraint), re-cement against the live
API before U4/U5; do not fabricate a shape.

---

## Requirements

- **R1. Typed, evaluable invariants.** Each `CandidateInvariant` is a validated
  `invariant-predicate/v1` AST over canonical fields with a semantic identity
  (`predicate_fingerprint`); the prose statement is a human render, not the
  definition. `Evaluate(predicate, signature) -> {satisfies|violates|unknown}` is
  pure and total.
- **R2. Code computes support; the model cannot self-certify it.** Support and
  contrast are computed by evaluating predicates against persisted signatures;
  the model's proposed predicate is executed, not trusted. A model whose
  predicate matches nothing simply earns zero support.
- **R3. Distinct-family support (not "independence").** Support counts
  `distinct_mechanism_families` discounting #11-redundant members; the term
  "independent" is not used for this v0 quantity.
- **R4. Failure coverage vs. success contrast are separate axes.**
  `FailureCoverage` over failure/partial-failure families; `Contrast` over
  partial-success/success families; mixed families member-wise or ambiguous.
- **R5. Association status, code-assigned where deterministic.**
  `recurring`/`discriminative` from measured coverage/contrast;
  `candidate_obstruction` only as a model-proposed flag (never code-certified);
  `unknown` on ambiguity. Nothing reaches `established`; state stays `proposed`.
- **R6. Epistemic status retained.** Aggregated support carries the
  explicit/inferred/… composition of the matched claims; `inferred`-only support
  is reported as such and never laundered into `explicit`.
- **R7. Provenance, replay, run lifecycle.** Each run records
  provider/model/config + raw payloads, links every candidate to its
  family evaluations + the FailureSpace/cluster-run it was mined from, and uses
  the create-`running` → `finalizeRun`/`failRun` lifecycle.
- **R8. Reproducibility / immutability.** Invariant revisions are immutable and
  revisioned per problem; re-mining creates a new revision, never a rewrite;
  identical fixture + versions ⇒ identical `predicate_fingerprint`s.
- **R9. CLI + `--json`; offline determinism.** `mine`/`list`/`show` emit stable
  human + JSON; all tests use fixtures (no network/model).

Requirements (R1–R9) and implementation units (U1–U8) are separate axes.

---

## Key Technical Decisions

- **KTD-1: Mining input is a compact, deterministic FailureSpace projection.**
  Code builds a `MiningRequest` from persisted data: per distinct family, the
  representative signature's resolved canonical content (per field kind), posture
  enums, boundary relations, outcome class + `OutcomeMixed`, and the #11
  redundancy flag. The provider sees structured facts, never free text. Note the
  provider request uses the **representative** per family (compact), but the
  deterministic support/contrast computation (KTD-3) loads **every non-redundant
  member** signature so member-wise mixed-family splitting and per-member
  epistemic composition are computed from real per-member data, not the
  representative alone.

- **KTD-2: The invariant IS a typed predicate; the model authors the AST, code
  evaluates it.** `invariant-predicate/v1` is an AST over canonical fields:
  set fields via `contains`/`any`/`all` over canonical IDs, posture axes via
  `in`/`equals` over enum values, boundary via relation match, plus boolean
  `all`/`any`/`not`. `Evaluate(pred, sig) -> {satisfies|violates|unknown}` is
  pure; `unknown` iff the predicate reads a field the signature left
  ambiguous/`Incomparable`. The evaluator distinguishes **field-unresolved** from
  **value-absent**: it uses `hasUnresolved` (`signature.go:246`) as the
  `unknown`-trigger — a field with any non-resolved claim → `unknown`; a field
  that is **present and resolved but does not contain the queried ID** →
  `violates` (not `unknown`). Empty-but-resolved set fields therefore evaluate a
  `contains`/`any` as `violates`, matching `compareSetField`'s treatment of empty
  resolved sets (`compare.go:268-284`). The model's role is confined to
  *proposing* the AST + prose; **support is `Evaluate` over persisted
  signatures**, so "code computes support" is literally true.
  `ModelJudgment != Verification`.

- **KTD-3: Support/contrast aggregate member-wise → family, retaining epistemic
  status.** `Evaluate` runs **per member signature**, not per cluster. For each
  candidate: evaluate against every eligible member signature, roll up to families
  (`supports` if the family's non-redundant members satisfy; `unknown`
  propagates). **Per-member outcome is not on `ClusterMemberRow`** — that row
  carries only `SignatureID`/`MechanismID`/`Redundant`/`Ordinal`
  (`cluster_store.go:13-18`). The only per-member outcome available is
  `signature_outcomes.class`, reached by loading each member's signature
  (`GetSignature`, `canon_store.go:454`). A cluster's `OutcomeMixed` flag
  (`cluster_store.go:29`) is therefore only the **trigger** to split that family's
  members by their own `signature_outcomes.class` into failure-support vs.
  success-contrast; non-mixed families use the cluster `OutcomeClass`. Compute
  `FailureCoverage` over failure/partial-failure families and `Contrast` over
  partial-success/success families **separately** (change 6). Each family
  evaluation stores the **epistemic composition** (counts of explicit/inferred/…
  among the matched claims), so `inferred`-only support is visible (change 3).

- **KTD-4: Distinct-family counting, not independence (change 4).**
  `distinct_mechanism_families(I)` = families whose non-redundant members satisfy
  `I`, with #11-redundant duplication discounted. Named
  `distinct_mechanism_families` / `nonredundant_support_units`; never
  "independent". Source/program independence is a deferred axis.

- **KTD-5: Semantic identity = canonical predicate fingerprint (change 7).**
  A candidate's logical identity is `sha256` over the **canonicalized predicate
  AST** — *not* the prose. Prose statements live on the revision and may re-phrase
  freely. Two paraphrases of the same predicate collapse to one invariant. To be
  deterministic and paraphrase-proof the canonicalizer applies a fixed
  **normalization pass** before hashing: (1) canonical IDs are already exact
  lower_snake at this layer (`domain/canonical.go`), so no aliasing; (2) leaf
  operand value lists (`in`/`any`/`all`-over-IDs) are sorted by a total order;
  (3) boolean `all`/`any` child lists are sorted by each child's own fingerprint
  (recursively) so nesting order is irrelevant; (4) single-child `all`/`any` are
  flattened to the child, and `not(not(x))` folds to `x`; (5) op serialization is
  fixed. This guarantees two structurally-equal ASTs hash identically and two
  genuinely different predicates do not collide. (If v1 chooses to *reject* rather
  than *normalize* degenerate nesting/double-negation, say so explicitly in U1 so
  the implementer does not guess — but the sort-and-flatten pass is the default.)

- **KTD-6: Association status, code-assigned where deterministic (change 5).**
  `recurring` when `distinct_mechanism_families ≥ min_support`; `discriminative`
  when it is `recurring` **and** shows a **deterministic outcome `Contrast`** —
  defined as: the predicate is `satisfies` across its supporting failure families
  yet `violates` (not `unknown`) in **≥1 eligible partial-success/success
  family** (so "behaves differently in success" is a reproducible count, not a
  vibe); `candidate_obstruction` **only** if the model
  proposed that hypothesis flag — recorded as a *model claim*, never code-
  certified; `unknown` when a read field is ambiguous across families. No
  `established`. Mathematics has no "causal status"; the term is `association_status`.

- **KTD-7: `proposed`-only state, enforced in the schema.**
  `candidate_invariants.initial_state TEXT NOT NULL DEFAULT 'proposed' CHECK
  (initial_state = 'proposed')`; the M4.3 transition machinery is absent.

- **KTD-8: Migration v11 is additive + one guarded constraint change.** New
  tables with `RAISE(ABORT, ...)` triggers. Because `provider_invocations.role`
  is `CHECK (role IN ('normalize'))` (`migrations.go:119`), reusing that table for
  the miner requires **generalizing the role constraint**. Two options; the plan
  takes **(a)**:
  - **(a) Generalize `provider_invocations.role`** via a v11 introspective table
    rebuild. `provider_invocations` is a **parent** table — `normalization_revisions.provider_invocation_id REFERENCES provider_invocations(id)`
    (`migrations.go:143`) — and the whole `Migrate` runs in **one transaction**
    with `PRAGMA foreign_keys = ON` (`store.go:56/69`); SQLite **ignores**
    `PRAGMA foreign_keys` toggles mid-transaction, so the rebuild cannot disable
    FK enforcement to "cheat". The rebuild must therefore follow the **exact,
    already-shipped and passing v10 template** `rebuildClusterRunsWithInputSetHash`
    (`migrations.go:754`), which rebuilds a parent table (`cluster_runs`, itself
    referenced by `mechanism_clusters`/`failure_spaces`/etc.) safely:
    `DROP TRIGGER` (immutability update/delete) → `CREATE TABLE
    provider_invocations__v11 (... CHECK (role IN ('normalize','invariant')) ...)`
    → `INSERT INTO …__v11 SELECT … FROM provider_invocations` → `DROP TABLE
    provider_invocations` → `ALTER TABLE provider_invocations__v11 RENAME TO
    provider_invocations` → recreate indexes + immutability triggers. This relies
    on modern `ALTER TABLE … RENAME`'s **child-FK propagation** (default
    `legacy_alter_table = OFF`): the `REFERENCES provider_invocations(id)` text in
    `normalization_revisions` is rewritten to the renamed table, and the old
    parent is dropped only after rows are copied. The whole step is **guarded**
    (run only if `invariant` is not already permitted — detect by attempting the
    generalized shape's presence, or by a `PRAGMA`/probe insert-rollback) so it is
    **idempotent and a no-op on a fresh DB** whose v11 already permits `invariant`.
    A store test MUST insert a `normalization_revision` (with its
    `provider_invocation_id` FK) under the `normalize`-only shape **before** v11
    and assert that FK still **resolves after** v11 — proving the child FK
    survived the parent rebuild, not merely that `invariant` is now allowed.
  - **(b)** Introduce a separate `inference_invocations` table (generalized
    role) and leave `provider_invocations` for normalize. Rejected for v0: it
    forks the provenance story; (a) keeps one auditable invocation table.
  Either way, migration v11 is **not** purely "add new tables" — the role
  constraint must be handled explicitly, FK-safely, and with a from-pre-v11 test.

---

## High-Level Technical Design

### Data flow (the revised core)

```text
FailureSpace
    |
    v
store.GetFailureSpace + GetClusterRun + members -> per family: all non-redundant member canon.MechanismSignature   [readers; U4]
    (per-member outcome via signature_outcomes.class; per-axis claim status via the U4 epistemic reader)
    |
    v
model proposes (provider.InvariantMiner, fixture in CI):                                  [U2]
    typed predicate (invariant-predicate/v1) + prose interpretation
    |
    v
deterministic evaluator: Evaluate(predicate, every eligible signature)                    [pure; U1, KTD-2]
    -> satisfies | violates | unknown
    |
    v
family aggregation (member-wise; mixed families member-wise):                             [pure; U3, KTD-3]
    supports | violates | unknown  + underlying epistemic strengths (explicit/inferred/…)
    |
    v
failure coverage (failure/partial-failure) + success contrast (partial-success/success)   [pure; U3, KTD-3/6]
    -> association_status (recurring | discriminative | candidate_obstruction* | unknown)
       (*model-proposed flag only)
    |
    v
CandidateInvariant(proposed)  — semantic identity = predicate_fingerprint                 [immutable; U5, KTD-5/7]
    |
    v
[M4.3 challenge]  (out of scope here)
```

The run wrapping this is created `running`, then `finalizeRun`→`completed` or
`failRun`→`failed`.

### Persistence shape (new tables — additive on v10, + one guarded constraint change)

Migration **v11** (candidate invariants):

```sql
-- One mining pass over a FailureSpace revision. Immutable; revisioned per problem.
CREATE TABLE invariant_revisions (
  id TEXT PRIMARY KEY,                            -- ivr_<ulid>
  problem_id TEXT NOT NULL REFERENCES problems(id),
  failure_space_id TEXT NOT NULL REFERENCES failure_spaces(id),
  cluster_run_id TEXT NOT NULL REFERENCES cluster_runs(id),
  run_id TEXT NOT NULL REFERENCES runs(id),
  provider_invocation_id TEXT NOT NULL REFERENCES provider_invocations(id),
  miner_version TEXT NOT NULL,                    -- 'invariant/v1'
  predicate_schema TEXT NOT NULL,                 -- 'invariant-predicate/v1'
  min_support INTEGER NOT NULL,                   -- distinct-family threshold (provenance)
  revision INTEGER NOT NULL,                      -- monotonic per problem
  candidate_count INTEGER NOT NULL,
  created_at TEXT NOT NULL,
  UNIQUE(problem_id, failure_space_id, miner_version, predicate_schema, min_support),
  UNIQUE(problem_id, revision)
);

CREATE TABLE candidate_invariants (
  id TEXT PRIMARY KEY,                            -- inv_<ulid>
  invariant_revision_id TEXT NOT NULL REFERENCES invariant_revisions(id),
  predicate_fingerprint TEXT NOT NULL,            -- KTD-5, SEMANTIC identity
  statement TEXT NOT NULL,                        -- human render of the predicate
  abstraction_level TEXT NOT NULL,
  initial_state TEXT NOT NULL DEFAULT 'proposed'
                 CHECK (initial_state = 'proposed'),               -- KTD-7
  association_status TEXT NOT NULL
                 CHECK (association_status IN ('recurring','discriminative',
                                               'candidate_obstruction','unknown')),  -- KTD-6
  obstruction_is_model_hypothesis INTEGER NOT NULL DEFAULT 0,      -- 1 iff model-proposed
  distinct_family_support INTEGER NOT NULL,       -- CODE-computed (KTD-4)
  failure_coverage_num INTEGER NOT NULL,          -- FailureCoverage numerator
  failure_coverage_den INTEGER NOT NULL,          -- eligible failure families
  support_explicit_count INTEGER NOT NULL,        -- epistemic composition (change 3)
  support_inferred_count INTEGER NOT NULL,
  support_other_count INTEGER NOT NULL,           -- ambiguous/unknown/unsupported
  confidence_ordinal TEXT NOT NULL DEFAULT '',
  ordinal INTEGER NOT NULL,
  UNIQUE(invariant_revision_id, predicate_fingerprint)
);

-- The stored predicate AST (canonical serialization; source of the fingerprint).
CREATE TABLE invariant_predicates (
  invariant_id TEXT PRIMARY KEY REFERENCES candidate_invariants(id),
  predicate_json TEXT NOT NULL                    -- canonicalized invariant-predicate/v1
);

-- Per-family evaluation verdict + epistemic composition (grounding + contrast).
CREATE TABLE invariant_family_evaluations (
  invariant_id TEXT NOT NULL REFERENCES candidate_invariants(id),
  cluster_id TEXT NOT NULL REFERENCES mechanism_clusters(id),
  outcome_class TEXT NOT NULL,                    -- incl. 'mixed'
  role TEXT NOT NULL CHECK (role IN ('support','contrast')),   -- change 6 split
  verdict TEXT NOT NULL CHECK (verdict IN ('satisfies','violates','unknown','member_mixed')),
  explicit_count INTEGER NOT NULL,                -- matched-claim epistemic composition
  inferred_count INTEGER NOT NULL,
  other_count INTEGER NOT NULL,
  PRIMARY KEY(invariant_id, cluster_id)
);

-- Known counterexamples: eligible failure families that VIOLATE the predicate.
CREATE TABLE invariant_counterexamples (
  invariant_id TEXT NOT NULL REFERENCES candidate_invariants(id),
  cluster_id TEXT NOT NULL REFERENCES mechanism_clusters(id),
  reason TEXT NOT NULL,                           -- which predicate clause it violates
  PRIMARY KEY(invariant_id, cluster_id)
);
-- + immutability triggers (UPDATE/DELETE RAISE(ABORT,...)) for each new table.

-- Guarded constraint change (KTD-8a): generalize provider_invocations.role from
-- CHECK (role IN ('normalize')) to CHECK (role IN ('normalize','invariant')) via
-- an introspective table rebuild, run ONLY if 'invariant' is not already allowed
-- (idempotent; no-op on a fresh DB whose v11 already permits it). This is a
-- FK-safe rebuild of a PARENT table: normalization_revisions.provider_invocation_id
-- REFERENCES provider_invocations(id) (migrations.go:143). See KTD-8a for the exact
-- mechanism and the child-FK-survival requirement.
```

> **Blueprint reconciliation.** `docs/persistence.md` sketches a forward-looking
> `candidate_invariant` schema with a prose `statement` as the artifact and the
> full challenge lifecycle. This slice **replaces the prose-as-definition** with
> a typed predicate (statement demoted to a human render), ships the
> **`proposed`-only** subset + code-computed evaluations, and **intentionally does
> not create** the lifecycle tables — `invariant_challenge`,
> `invariant_state_transition`, the `invariant_transition_counter` seq trigger,
> the `invariant_current_state` view, and `invariant_lineage` — all deferred to
> **M4.3**. A reviewer can confirm no lifecycle machinery leaked into v11 by
> checking those five names are absent from the migration. Names are pluralized to
> match the shipped convention (`mechanism_clusters`, `signature_outcomes`).

### Output structure

New pipeline views (`internal/pipeline/output.go`), all `--json`-stable:

- `InvariantRevisionView` — id, problem, failure_space_id, cluster_run_id,
  miner_version, predicate_schema, min_support, revision, candidate_count,
  created_at.
- `CandidateInvariantView` — id, predicate_fingerprint, predicate (AST),
  statement, abstraction_level, association_status,
  obstruction_is_model_hypothesis, distinct_family_support, failure_coverage
  (num/den), support epistemic composition (explicit/inferred/other), family
  evaluations (cluster id, outcome, role support|contrast, verdict, epistemic
  counts), counterexamples (cluster id + reason).

Human output mirrors the toolbox example style, e.g.
`IF-01: preserves(residue_locality) ∧ locality∈{local,mixed} — recurring across 4 distinct families; discriminative (absent in 2/3 partial-success); support 3 explicit / 1 inferred`.

---

## Implementation Units

### U1. Predicate contract + evaluator (`internal/invariant`, pure — no SQL/Cobra/provider)
Define `invariant-predicate/v1`: the AST types (leaf field predicates over set
fields / posture axes / boundary relation / outcome with ops
`contains`/`any`/`all`/`in`/`equals`/`not`, plus boolean composition), a
**validator** (unknown field/op/canonical-ID → error), a **canonical
serialization** and `PredicateFingerprint()` (sorted operands, normalized IDs;
KTD-5), and `Evaluate(pred, canon.MechanismSignature) ->
{satisfies|violates|unknown}` (pure; `unknown` on ambiguous/`Incomparable`
reads). Unit tests: validation rejects malformed ASTs; fingerprint is
order-independent and paraphrase-free (two structurally-equal ASTs → same hash;
a genuine op/ID change → different hash); evaluator truth table incl. `unknown`.

### U2. Miner provider role + deterministic fixture (`internal/provider`)
`provider.InvariantMiner` interface (`Mine(ctx, MiningRequest) (MiningResponse,
error)`); `MiningResponse` carries `[]CandidateProposal{predicate, statement,
abstraction_level, association_hypothesis?}` + `Metadata` + raw payloads (mirrors
`NormalizeResponse`). `FixtureInvariantMiner` keyed on the `MiningRequest`
fingerprint → canned proposals for the project fixtures. Provider tests:
deterministic output; role recorded; a fixture proposal whose predicate is
invalid is surfaced as an error, not silently stored. **No domain/SQL coupling.**

### U3. Support/contrast engine (`internal/invariant`, pure)
`BuildMiningRequest(families) MiningRequest` (KTD-1) and
`EvaluateCandidates(proposals, families) []Candidate` computing, per candidate:
per-member `Evaluate` rolled up to family verdicts (mixed families split
member-wise by each member's own `signature_outcomes.class`, since per-member
outcome is not on `ClusterMemberRow` — KTD-3), `FailureCoverage`
(failure/partial-failure) and `Contrast` (partial-success/success) as separate
axes, `distinct_family_support` discounting #11 redundancy (KTD-4), the
**epistemic composition** of matched claims for the axis each predicate reads
(KTD-3, via the U4 epistemic reader), counterexamples, and
`association_status` (KTD-6; `candidate_obstruction` only when the proposal
carried that hypothesis flag, recorded as model-hypothesis). The engine is pure
over already-loaded family structs (member signatures + per-member outcome +
per-axis claim status); the N-member `GetSignature` loads happen in the U4 reader,
not here. Unit tests: a
proposal whose predicate matches nothing → zero support; redundant supporters do
not inflate `distinct_family_support`; failure-support vs. success-contrast
separation; mixed family split member-wise (members land on different sides by
their own outcome, never the aggregate); `inferred`-only support reported as such
(not laundered); ambiguous decisive read → `unknown`; a model over-claiming
`candidate_obstruction` is stored flagged as hypothesis, never as a code verdict.

### U4. FailureSpace → families reader (`internal/store` / pipeline)
Given a `failure_space_id`, return the distinct families with each member
signature rehydrated to `canon.MechanismSignature` **plus each member's per-claim
epistemic status across all read axes**. **Do not reuse `signatureFromRecord`
(`internal/pipeline/mechanism.go:390`) for this.** That helper is lossy for
epistemic purposes: it copies set-field `Status` (`:416`) but sets **no** `Status`
on boundaries (`:437-443`) and never carries posture/outcome provenance — even
though `store.SignatureRecord` persists it all (`OutcomeStatus` `canon_store.go:267`,
`PostureStatus map[string]string` `:269`, boundary `ClaimStatus` `:251`, all
populated by `GetSignature` at `:429/445/455`). Reusing it would make every
`boundary`/`posture`/`outcome` predicate's epistemic composition structurally
empty (silently `other`/`unknown`), an **unearned** status the epistemic-retention
requirement (R6) forbids. U4 therefore adds a reader that maps
`SignatureRecord.ClaimStatus` (set fields), boundary `ClaimStatus`, `PostureStatus`
per axis, and `OutcomeStatus` into whatever structure U3 aggregates, so the
`explicit`/`inferred`/… of a matched claim is available **for the specific axis a
predicate reads**. Reuse #11 cluster-member + #9 signature readers for the
signature payload itself (add a thin `ListFamiliesForFailureSpace` if none
exists), but the epistemic-status mapping is new. Store tests over seeded #11
fixtures, including one asserting that an `inferred`-only **boundary** (and an
`inferred`-only **outcome**) surfaces as `inferred`, not `other`.

### U5. Invariant persistence (migration v11) + writer/reader
Migration v11: the new tables + immutability triggers **and** the guarded
`provider_invocations.role` generalization (KTD-8a: FK-safe introspective rebuild
following the shipped `rebuildClusterRunsWithInputSetHash` template
(`migrations.go:754`) to `CHECK (role IN ('normalize','invariant'))`, run only if
not already permitted).
**`validateSchemaTables` extension is mandatory, not optional.** That function
(`store.go:912`) runs inside `Migrate`'s single transaction on **every** store
open and returns `ErrCorruptStore` on any missing table (`:938`) or missing
newest-migration column (`:947`); forgetting it breaks every `Open`, including all
pre-existing tests. v11 MUST therefore:
- append these five to the table-existence list: `invariant_revisions`,
  `candidate_invariants`, `invariant_predicates`, `invariant_family_evaluations`,
  `invariant_counterexamples` (keep the existing v10 tables);
- **keep** the v10 column checks (`cluster_runs.input_set_hash`,
  `mechanism_clusters.outcome_class`/`outcome_mixed`);
- prove the role CHECK was generalized. A column-existence probe cannot see a
  `CHECK`, so guard it with a **probe insert** of a synthetic `invariant`-role row
  in a savepoint that is rolled back (or an equivalent assertion), so a DB that
  silently retained the `normalize`-only CHECK is reported corrupt rather than
  accepted.

`store.PersistInvariantRevision(...)`
(transactional, idempotent on `UNIQUE(problem, failure_space, miner_version,
predicate_schema, min_support)` — re-run returns the existing revision) and
`GetInvariantRevision`/`ListCandidateInvariants(problem)`/`GetCandidateInvariant(id)`.
Store tests: idempotency; immutability (raw UPDATE/DELETE aborts); `proposed`
`CHECK` rejects other initial states; `association_status` `CHECK`; provenance +
family-evaluation round-trip; revision monotonicity; a **fresh-migrate schema
validation** for the new tables; and a **pre-v11 `provider_invocations` role
migration test** that (i) constructs the `normalize`-only `CHECK` shape,
(ii) inserts a `provider_invocation` + a `normalization_revision` whose
`provider_invocation_id` FK points at it, (iii) runs v11, then asserts both that
an `invariant`-role insert now succeeds **and** that the pre-existing
`normalization_revision`'s FK still resolves (the child FK survived the parent
rebuild); idempotent on a fresh DB.

### U6. Pipeline services + CLI + run lifecycle
`newf invariants mine --failure-space <id> [--min-support <n>] [--force]`,
`newf invariant list --problem <id>`, `newf invariant show <id>`, all `--json`.
`mine` **creates its run `running`**, wires reader → `FixtureInvariantMiner` →
`EvaluateCandidates` → persist (recording the provider invocation), then
`finalizeRun`→`completed` on success or `failRun`→`failed` on error (so
`run show` reflects the real outcome). Register in `cmd/newf/root.go`; CLI
integration tests (mine → list → show; `--force`/`--min-support` change → new
revision else existing; run status transitions asserted via `run show`;
`--json` contract in `output_test.go`).

### U7. Deterministic fixtures + regressions + end-to-end
Reuse #9/#11 fixtures; add: (a) a predicate satisfied across ≥`min_support`
distinct families → `recurring`; (b) same predicate but absent in
partial-success families → also `discriminative`; (c) two supporters #11-redundant
→ `distinct_family_support` drops below threshold → not `recurring`;
(d) `inferred`-only support → reported epistemic composition, not `explicit`;
(e) a mixed-outcome family → members split member-wise by their own
`signature_outcomes.class` (a failing member counts as support, a
partial-success member as contrast), never forced onto a side as the aggregate;
(f) an ambiguous decisive read → `unknown`; (g) a model proposing
`candidate_obstruction` → stored as model-hypothesis flag, `association_status`
reflects only what code measured; (h) re-mine same FailureSpace → identical
`predicate_fingerprint`s (new revision, prior untouched). Full offline
end-to-end: `seed → signatures → cluster build → failure-space build →
invariants mine → invariant show`, asserting run status `completed`.

### U8. Docs + quality gates
`docs/invariant-mining.md` (predicate contract + evaluator, `MiningRequest`,
member-wise support/contrast, epistemic-composition carry-through, distinct-family
vs. independence, `association_status`, `proposed`-only, CLI, run lifecycle),
README section, EPIC.md **M4.2** status note, `docs/persistence.md` v11 update,
and a forward note in `docs/mechanism-clustering.md`. Gates: `go build ./...`,
`go vet ./...`, `go test ./...` green; `gofmt -l .` empty; a boundary test at
every new seam (predicate/evaluator, engine, provider, store, CLI).

---

## Verification Contract

- Build/vet/test green; `gofmt` clean; offline CI (no network/model).
- **R1 (typed + evaluable):** the evaluator truth table (satisfies/violates/
  `unknown` on ambiguous read) is asserted; a prose-only "predicate" is rejected
  by the validator. `PredicateFingerprint` is order-independent and identical for
  two paraphrased prose statements sharing one AST, distinct for a real op/ID
  change.
- **R2/KTD-2 (support is code-owned):** a fixture proposal whose predicate
  matches no signature earns zero support (the model cannot assert coverage the
  `Evaluate` disproves). (Falsify: temporarily trust a model-supplied support set
  and assert the test fails.)
- **R3/KTD-4 (distinct families, not independence):** a fixture with two
  #11-redundant supporters asserts `distinct_family_support` counts the family
  once; flipping the redundant flag changes the count deterministically. The term
  "independent" appears nowhere in the emitted views.
- **R4 (coverage vs. contrast):** `FailureCoverage` is computed only over
  failure/partial-failure families and `Contrast` only over
  partial-success/success; a mixed family is split **member-wise by each member's
  own `signature_outcomes.class`** (its aggregate `OutcomeMixed` flag only
  triggers the split) and never silently assigned to a side.
- **R5/KTD-6/7 (association, no promotion):** `recurring`/`discriminative` are
  code-assigned from measured coverage/contrast; a model-proposed
  `candidate_obstruction` is stored with `obstruction_is_model_hypothesis = 1`
  and is never produced by code alone; a direct `INSERT` with `initial_state !=
  'proposed'` or an unknown `association_status` is rejected by `CHECK`; no path
  yields `established`.
- **R6 (epistemic retention):** an `inferred`-only-supported candidate reports
  `support_inferred_count > 0, support_explicit_count = 0`; support is never
  laundered to `explicit`.
- **R7 (provenance/replay/lifecycle):** every candidate resolves back through
  family evaluations → signatures → mechanisms → snapshots and to the
  `provider_invocations` row (whose `role` now legally = `invariant`) with
  retained payloads; the mining run shows `completed` on success and `failed`
  (with summary) on a forced error, via `run show`.
- **R8 (immutability/reproducibility):** re-mining is idempotent on the version
  tuple; a differing `min-support`/miner version creates a new revision; raw
  `UPDATE`/`DELETE` on any v11 table aborts (direct SQL probe); a **pre-v11
  `provider_invocations` role migration test** starts from the `normalize`-only
  `CHECK`, inserts a `normalization_revision` whose `provider_invocation_id` FK
  points at a seeded invocation, runs v11, and asserts both that `invariant` is
  now permitted **and** that the pre-existing child FK still resolves (the parent
  rebuild did not orphan it); idempotent on fresh; `validateSchemaTables`
  rejects a DB left on the `normalize`-only CHECK; permuting family input order
  yields identical `predicate_fingerprint`s.
- **R9:** each command emits stable `--json` (contract test).

## Definition of Done

Given a problem's materialized FailureSpace, `newf invariants mine` runs under
the real run lifecycle (`running` → `completed`/`failed`) and produces explicit,
typed `CandidateInvariant` records — each a **validated, machine-evaluable
predicate** (prose is a human render), `proposed`, with **code-computed** family
evaluations (`Evaluate` over persisted signatures), a `FailureCoverage`/`Contrast`
split, `distinct_mechanism_families` support that discounts #11 redundancy, the
retained epistemic composition of matched claims, counterexamples, and an
`association_status` the model cannot inflate — persisted as an immutable,
provenance-preserving, revisioned artifact, entirely offline. The set is stable
enough to become the input contract for **M4.3 (challenge and falsify)**.

## Risks & Dependencies

- **First fuzzy operator, contained by the predicate contract.** The model only
  *authors the AST + prose*; code *evaluates* it. If a real adapter later emits
  weaker structure, the validator rejects it; the domain + evaluator are
  unaffected. This is the change that makes "code computes support" literally
  true rather than aspirational.
- **Predicate expressiveness vs. `invariant-predicate/v1` scope.** v1 covers set
  membership over canonical IDs, posture/outcome enums, and boundary relations —
  enough for conserved-structure invariants over #9 signatures. Richer predicates
  (counts, cross-field relations) are a future `invariant-predicate/v2` behind
  the same fingerprint discipline; v1 deliberately ships small.
- **Distinct ≠ independent.** `distinct_mechanism_families` is mechanistic
  non-redundancy under one `ComparisonProfile`, **not** statistical/historical/
  intellectual/causal independence (shared paper, author program, or assumption
  lineage are not excluded). Source/program independence is a deferred axis;
  nothing in this slice claims it.
- **Over-merge inherited from `cluster/v1`.** #11 now enforces pairwise coherence
  (not blind transitive closure) and preserves mixed outcomes, which mitigates
  this; residual over-merge is surfaced via counterexamples and is M4.3's to
  stress. Documented, not hidden.
- **`provider_invocations.role` constraint.** v11 must generalize the
  `normalize`-only `CHECK`; the guarded introspective rebuild (KTD-8a) is
  required and tested from the pre-v11 shape. This is the one non-additive step.
- **Depends on #11 hardening** (schema v10, commit `151df63`). Re-cement U3/U5
  against the live `internal/canon`/`internal/store` API if a surface was
  renamed.

## Sources & Research

- EPIC.md **M4.2 — Mine candidate failure invariants** (goal, CLI, exit
  condition) and **Gate B/Gate C** prerequisites (satisfied by #9/#11).
- `docs/toolbox-dsl.md` (`compress: Set -> CandidateInvariant[]`,
  "seek conserved_structure across independent_clusters", explicit candidate
  state).
- `AGENTS.md` (candidate-invariant discipline, `ModelJudgment != Verification`,
  no epistemic promotion, provider provenance/fixtures).
- `docs/abstraction-safety.md` (grounding + preserved-property recording).
- `docs/persistence.md` `candidate_invariant*` blueprint (predicate replaces
  prose-as-definition; `proposed`-only subset shipped here).
- Live surfaces (verified against commit `151df63`, schema v10):
  `internal/store/{failure_space_store,cluster_store,canon_store,migrations}.go`
  (incl. `provider_invocations` role `CHECK`), `internal/cluster/cluster.go`
  (`InputSetHash`, `PrimaryOutcome`, `OutcomeMixed`), `internal/canon/{signature,
  compare,fingerprint}.go`, `internal/provider/{provider,fixture}.go`,
  `internal/pipeline/app.go` (`finalizeRun`/`failRun`), `internal/domain/id.go`,
  `internal/domain/normalize.go` (`OutcomeMixed` aggregate-only semantics).

## Product Contract preservation

- No SQL/Cobra in `internal/domain`, `internal/invariant`, or
  `internal/provider`; SQLite behind `internal/store`; CLI wiring thin in
  `cmd/newf`.
- Provider coupling confined to `internal/provider`; the miner invocation records
  its role; the predicate AST + evaluator are pure domain-adjacent code with no
  vendor concepts.
- **No epistemic-status promotion:** candidates are `proposed` only; support,
  contrast, epistemic composition, and code-assignable `association_status` are
  computed; `candidate_obstruction` is a recorded model hypothesis; `established`
  is unreachable in this slice.
- Additive migration v11 with immutability triggers + one guarded, tested
  `provider_invocations.role` generalization; historical revisions never
  rewritten.
