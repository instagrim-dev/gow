# Invariant challenge and falsification (M4.3)

`newf challenge` implements the toolbox `challenge` / `falsify` operators: it
**attacks** the `CandidateInvariant` records mined by M4.2 before they may
influence search. Every attack is a durable, typed challenge that produces
**evidence and an explicit state transition — never critic prose**. After this
slice, only challenged invariants can influence search policy: the M5.1
frontier reads exclusively `surviving` / `established` candidates.

## The lifecycle (state lives only in transitions)

```text
proposed -> challenged -> {surviving | weaken | falsified}
surviving -> {challenged | weaken | falsified | established}
weaken    -> {challenged | surviving | falsified}
falsified : terminal
```

There is **no state column**. The sole source of truth is the append-only
`invariant_state_transitions` ledger, guarded by a validating trigger
(sequence must match the atomically-allocated per-invariant counter; from_state
must match the current ledger head; only the transitions above are legal), read
through the `invariant_current_state` view. A cached column cannot drift from a
ledger that is the only writer surface.

## Challenge types and what confirms them

The challenger provider **proposes** attacks with a claimed verdict; pure
code-owned verifiers (`internal/invariant/challenge.go`) confirm or deny each
one against rehydrated signatures. `ModelJudgment != Verification`: an
unconfirmable claim is persisted for audit as `unconfirmed`, links **no**
evidence, and drives **no** transition (KTD-3) — a hypothesis is neither
strengthened nor weakened by a claim code cannot confirm.

| Type | Confirmed when (code-verified) | Lifecycle force |
|---|---|---|
| `known-counterexample` | a failure-side member in the atlas evaluates to `violates` (ambiguity is not a counterexample) | **falsified** |
| `synthetic-counterexample` | a provider-constructed approach (rehydrated with exhaustively-complete fields) evaluates to `violates`; persisted as a `synthetic_artifacts` row | **weaken** — constructibility shows the invariant is not conserved by necessity; only an observed, in-atlas counterexample falsifies the empirical regularity |
| `success-preserving` | a success/partial-success family's eligible members all satisfy the predicate (it does not discriminate outcome) | **weaken** |
| `bias-critique` | the deterministically **recomputed** distinct-family support falls below the mining threshold; the recount is the evidence (KTD-6) and is retained for audit even when unconfirmed | **weaken** |
| `split` | ≥2 valid child predicates, each distinct from the parent, grounding to **nonempty, pairwise-disjoint** supporting failure families (abstraction safety: ground or reject) | **weaken** + children persisted |
| `merge` | one child predicate covering the **union** of the parents' support whose contrast violations are **not lower than any parent's** (a merge that erases the outcome-separating axis is rejected) | **weaken** + child persisted |
| `independent-verification` | never provider-claimable; see the `established` gate below | surviving → **established** |

A campaign whose every attack failed confirmation leaves the invariant
**surviving** — the attacks were made and did not land. That is the only way to
earn `surviving`; it cannot be requested.

Confirmed split/merge children are persisted as **real candidate invariants**
in a new revision (miner version `challenge-split/v1` / `challenge-merge/v1`)
with support recomputed by the same engine mining uses, linked to the parent
through `invariant_lineage`. Children enter `proposed` and must survive their
own challenges — no inherited authority.

## The `established` gate (KTD-2)

The DB trigger permits `surviving → established` structurally; the pipeline
gates it **epistemically**. `newf invariant establish <id> --snapshot <snap>
--locator <loc>` refuses:

- any state other than `surviving`;
- any request without a persisted source snapshot + locator (independent,
  non-model evidence, recorded as `independent_source` challenge evidence).

No provider path reaches `established`: a challenger proposing
`independent-verification` is recorded inert. Model agreement — however
enthusiastic — tops out at `surviving`.

## Evidence links to real rows

The blueprint's sketched `evidence_record` table was never shipped, so
challenge evidence (`invariant_challenge_evidence`) references what actually
exists: `mechanism_clusters` (family evidence), `mechanism_signatures`
(counterexample members), `source_snapshots` (independent verification), plus a
structured detail (support recounts, grounding facts). `result_summary` is a
label over that evidence, never a substitute for it.

## Persistence (migration v13)

`invariant_challenges`, `invariant_transition_counters` (the **only** mutable
surface — the atomic sequence allocator), `invariant_state_transitions`
(+ validating trigger), `invariant_current_state` (view),
`invariant_challenge_evidence`, `synthetic_artifacts`,
`invariant_challenge_synthetic_artifacts`, `invariant_lineage` — all immutable
by `RAISE(ABORT)` trigger except the counter. v13 also widens
`provider_invocations.role` to admit `'challenge'` via the same guarded,
FK-safe in-place `writable_schema` CHECK edit v11 established (the table is a
live FK parent; it is never dropped). `validateSchemaTables` asserts the new
tables and that the role CHECK admits both `invariant` and `challenge`.

A campaign is persisted **atomically**: an illegal transition anywhere aborts
the whole campaign — no challenge rows, no partial evidence, no state change.

## Provenance, determinism, offline CI

Each campaign records one `provider_invocations` row (`role='challenge'`) with
raw request/response payloads for replay. Candidates are processed in id order
and challenge types in the provider's fixed order (KTD-4), so identical stores
produce identical transition sequences. The default challenger is
`DerivingFixtureChallenger` — deterministic rule-derived attacks
(known-counterexample, success-preserving, bias-critique, and split when the
predicate is boolean-composed). It deliberately proposes **no** synthetic
counterexample: a construction designed to violate always violates, so a
fixture-authored synthetic would weaken every candidate vacuously; meaningful
synthesis needs a real model. CI makes no network/model calls.

## CLI

```text
newf challenge <invariant-id>
newf challenge --problem <id> --all
newf invariant state <invariant-id>
newf invariant list --problem <id> --state surviving
newf invariant establish <invariant-id> --snapshot <snap-id> --locator <loc> [--note ...]
```

All support `--json`. `invariant list --state surviving|established` is the
frontier read surface for M5.1.

## Boundaries

- `internal/invariant/challenge.go` — pure verifiers (no SQL/Cobra/provider).
- `internal/provider/invariant_challenger.go` — `Challenger` interface +
  deriving fixture.
- `internal/store/challenge_store.go` + migration v13 — atomic campaign
  persistence, ledger, read surfaces.
- `internal/pipeline/challenge*.go` — campaign orchestration, verdict policy,
  the `established` gate, run lifecycle.
- `cmd/newf/challenge.go` — thin CLI wiring.
