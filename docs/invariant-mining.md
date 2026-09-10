# Invariant mining (M4.2)

`newf invariants mine` implements the toolbox `compress` operator
(`compress: Set -> CandidateInvariant[]`): it compresses a materialized
[failure space](mechanism-clustering.md) into explicit, typed
**`CandidateInvariant`** records that name structure conserved across the
problem's distinct mechanism families.

This is the first slice whose core operation is **model-native and fuzzy**. The
containment contract is the typed predicate: the model *authors* an executable
predicate AST; **code evaluates it** against persisted signatures. "The model
proposes, code computes support" is literally true, not aspirational
(`ModelJudgment != Verification`).

## The predicate contract (`invariant-predicate/v1`)

An invariant's durable, executable definition is a validated AST over the same
canonical fields #9 fingerprints (`internal/invariant`, pure — no SQL, Cobra,
or provider coupling). The prose `statement` is a human render only.

Leaf operators:

| Op | Fields | Meaning |
|---|---|---|
| `contains` | `representations` `operators` `assumptions` `preserves` `breaks` `auxiliary_objects` | the field's resolved canonical IDs include `canonical_id` |
| `in` / `equals` | `locality` `construction` `uncertainty` `outcome` | the enum axis is one of `values` / equals the single value |
| `boundary` | — | some boundary matches `canonical_id`, and `relation` when given (the typed relation is decisive: `stops_at(X)` never matches `requires(X)`) |

Boolean composition: `all`, `any` (≥1 child), `not` (exactly 1 child).

The validator rejects unknown ops/fields, malformed canonical IDs, invalid enum
values, extraneous node fields, unknown JSON keys, and prose-only proposals —
before anything is stored.

### Evaluation: `Evaluate(predicate, signature) → satisfies | violates | unknown`

The evaluator is pure and total, and it distinguishes **field-unresolved** from
**value-absent**:

- a read field containing any non-resolved claim (ambiguous / novel / unknown /
  rejected) → `unknown` — ambiguity stays visible, never coerced;
- a **resolved** field that simply lacks the queried ID → `violates` (an empty
  resolved set violates a `contains`, matching how comparison treats empty
  resolved sets);
- an enum axis recorded as its `unknown` sentinel → `unknown`;
- `unknown` propagates through boolean composition (`not(unknown) = unknown`;
  `all`/`any` short-circuit only on a decisive child).

### Semantic identity: the predicate fingerprint

A candidate's logical identity is `sha256` over the **canonicalized** AST plus
schema version — never the prose, so a model paraphrasing itself cannot mint
three invariants from one predicate. The canonicalization pass (KTD-5):

1. leaf value lists sorted;
2. boolean child lists sorted by each child's own canonical serialization
   (nesting order is irrelevant);
3. single-child `all`/`any` flattened to the child;
4. `not(not(x))` folded to `x`;
5. fixed op serialization.

The canonical JSON is persisted (`invariant_predicates.predicate_json`) as the
exact bytes the identity derives from.

## Mining flow

```text
FailureSpace (pinned to one cluster run / input_set_hash)
    -> families reader: every NON-redundant member signature rehydrated
       WITH epistemic provenance; per-member outcome from signature_outcomes.class
    -> MiningRequest: compact per-family projection (structured facts, no free text)
    -> miner proposes: typed predicate + prose render (+ optional obstruction hypothesis)
    -> validator rejects malformed predicates (run fails, nothing stored)
    -> Evaluate(predicate, every eligible member signature)     [code-owned]
    -> member-wise -> family rollup; FailureCoverage vs Contrast [code-owned]
    -> CandidateInvariant(proposed), immutable revision           [schema v11]
```

The wrapping run is created `running`, then `finalizeRun` → `completed` or
`failRun` → `failed`, so `run show` reflects the real outcome.

### The epistemic reader

The families reader deliberately does **not** reuse the comparison rehydrator
(`signatureFromRecord`), which drops boundary claim status and posture/outcome
provenance. `signatureFromRecordWithProvenance` maps the persisted
`claim_status` for set-field claims, boundaries, each posture axis, and the
outcome, so the epistemic strength of a matched claim is available **for the
specific axis a predicate reads**. A missing persisted status is preserved as
`unknown`, never promoted to `explicit`.

### Support vs. contrast (separate axes)

"Spans ≥2 outcome classes" is conceptually muddy for a *failure* invariant, so
the two axes are kept separate:

```text
FailureCoverage(I) = failure families satisfying I / eligible failure families
Contrast(I)        = I violates (not unknown) in ≥1 partial-success/success family
```

- **failure / partial_failure** families provide *support*;
- **partial_success / success** families provide *contrast or counterevidence*;
- **mixed** families (aggregate `OutcomeMixed`, preserved by #11) are split
  **member-wise by each member's own `signature_outcomes.class`** — the
  aggregate flag only triggers the split; members are never forced onto a side;
- `unknown`-outcome families are eligible for neither axis.

Eligible failure families that **violate** the predicate are persisted as
counterexamples.

### Distinct families, not "independence"

`distinct_family_support` counts families whose non-redundant members satisfy
the predicate, discounting #11-redundant members. This is **mechanistic
non-redundancy under one `ComparisonProfile`** — not statistical, historical,
intellectual, or causal independence (two distinct clusters can still share a
paper or author program). The word "independent" appears nowhere in the emitted
views. Source/program independence is a deferred axis.

### Epistemic composition is retained

Each family evaluation and each candidate aggregates the claim statuses of
exactly the fields its predicate read: `explicit` / `inferred` / `other`
(ambiguous, unknown, unsupported) counts. Three families supported entirely by
`inferred` claims are *not* equivalent to three backed by explicit source
claims, and the counts make that visible; `inferred` support is never laundered
into `explicit` (R6).

### `association_status`

Code assigns only what it measured (mathematics has no "causal status"):

| Status | Assigned by | Rule |
|---|---|---|
| `recurring` | code | `distinct_family_support ≥ min_support` |
| `discriminative` | code | recurring **and** violates in ≥1 eligible partial-success/success family |
| `candidate_obstruction` | — | **never code-assigned**; a model-proposed hypothesis recorded via `obstruction_is_model_hypothesis = 1`. M4.3 earns the right to strengthen it through challenge/falsification |
| `unknown` | code | threshold not met, or decisive reads ambiguous |

Candidates enter and stay `proposed` in this slice — the DB `CHECK` enforces
it, and no transition machinery exists here (that is M4.3). `established` is
unreachable.

## Persistence (migration v11)

Immutable, per-problem revisioned tables (all with `RAISE(ABORT)` triggers):
`invariant_revisions` (idempotent on
`(problem, failure_space, miner_version, predicate_schema, min_support)`),
`candidate_invariants` (identity = `predicate_fingerprint`;
`initial_state CHECK ('proposed')`), `invariant_predicates` (canonical AST
JSON), `invariant_family_evaluations` (per-family verdict + role + epistemic
counts), `invariant_counterexamples`. Re-mining an unchanged failure space
returns the existing revision; a changed `min-support`/miner version creates
the next revision, never a rewrite.

The mining invocation is recorded in the shared `provider_invocations` table
with `role='invariant'`. That required the **one non-additive step** in v11:
the pre-v11 `CHECK (role IN ('normalize'))` is generalized in place via a
guarded `PRAGMA writable_schema` DDL-text edit (widening one CHECK enum) —
chosen over a drop/rename parent rebuild because `normalization_revisions`
holds live FKs into the table and `Migrate` runs in a single `foreign_keys=ON`
transaction. The edit never drops the parent, so child FKs are preserved by
construction; it is guarded (no-op on fresh v11 databases), followed by an
`integrity_check`, and covered by a from-pre-v11 test asserting both that
`invariant` is permitted and that a pre-existing `normalization_revision` FK
still resolves. See [persistence.md](persistence.md).

`validateSchemaTables` asserts the five new tables **and** that the
`provider_invocations.role` CHECK admits `invariant` (a column probe cannot see
a CHECK, so it reads the table DDL directly), so a database that silently
retained the normalize-only CHECK is reported corrupt.

## CLI

```text
newf invariants mine --problem <id> [--failure-space <id>] [--min-support <n>]
newf invariant list --problem <id>
newf invariant show [invariant-revision-id] [--problem <id>]
```

All support `--json` with stable views. `mine` defaults to the latest failure
space and `min-support` 2. Human output leads with state, association,
support, coverage, and the epistemic split (`e2/i1/o0`).

## Determinism and offline CI

The default miner is a deterministic fixture. `FixtureInvariantMiner` replays
canned proposals keyed on the order-independent `MiningRequest` fingerprint
(tests); `DerivingFixtureInvariantMiner` (CLI default) derives proposals by a
fixed transparent rule — propose `contains(field, id)` for every canonical ID
shared by ≥2 distinct failure-side families. Derived or canned, every proposal
is validated and re-evaluated by the engine exactly like a model's: it earns
only the support `Evaluate` proves. A real provider adapter replaces the
fixture behind the same `InvariantMiner` interface without touching the domain
or evaluator. CI makes no network or model calls.

## Boundaries

- `internal/invariant` — predicate contract + evaluator + support/contrast
  engine; pure (no SQL/Cobra/provider).
- `internal/provider` — `InvariantMiner` interface + fixtures; no domain/SQL
  coupling.
- `internal/store` — migration v11 + revisioned persistence.
- `internal/pipeline` — epistemic reader, mining service, run lifecycle.
- `cmd/newf` — thin CLI wiring.

Deferred to M4.3: all state transitions beyond `proposed`
(`challenged`/`surviving`/`weakened`/`split`/`merged`/`falsified`/
`established`), the `invariant_challenge`/`invariant_state_transition`
lifecycle tables, synthetic-counterexample construction, and abstraction
split/merge.
