# Invariant challenge and falsification (M4.3)

`newf challenge` implements the toolbox `challenge` / `falsify` operators: it
**attacks** the `CandidateInvariant` records mined by M4.2 before they may
influence search. Every attack is a durable, typed challenge that produces
**evidence and an explicit state transition — never critic prose**. After this
slice, only challenged invariants can influence search policy: the M5.1
frontier reads `surviving` and `operator_attested` candidates — the
challenged, unfalsified structure (see `frontier-generation.md`).

## The lifecycle (state lives only in transitions)

```text
proposed          -> challenged
challenged        -> {challenged | surviving | weaken | falsified}
surviving         -> {challenged | weaken | falsified | operator_attested}
weaken            -> {challenged | surviving | falsified}
operator_attested -> {challenged | weaken | falsified}
falsified : terminal
```

`challenged` is **resumable**, not a dead end (G2): a campaign that reaches no
decisive outcome parks the invariant at `challenged`, and a later campaign
(`challenged -> challenged` opens it again) can attack it with a stronger
challenger or a new failure population. `operator_attested` is **challengeable**
(G4): operator attestation records a human assertion, not machine verification,
so it never makes a hypothesis immune to further challenge.

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
| `known-counterexample` | a failure-side member in the atlas evaluates to `violates` (ambiguity is not a counterexample) **and** the candidate is a verified-universal claim; recurrence is a *frequency label*, not a quantifier, so `recurring` grants the universal (falsifiable-by-one-counterexample) reading **only** when the corpus is actually universal over the eligible failure population (full failure coverage). A `recurring` candidate that is not fully covering, a `contrast_observed` (association) claim, or an `unknown`-kind claim is **not** refuted by an isolated counterexample (G3) | **falsified** (verified-universal only) |
| `synthetic-counterexample` | a provider-constructed approach evaluates to `violates` **via admissibly supported structure**. The construction is a PROPOSAL: its set fields are `unobserved`, so an omitted/empty description reads as `unknown` (inert), and a violation resting on an **unsupported** present claim is likewise only a proposal — presence in a generated description is not stronger evidence of realizability than absence from it (G3) | **weaken** — constructibility (from supported structure) shows the invariant is not conserved by necessity; only an observed, in-atlas counterexample falsifies the empirical regularity |
| `success-preserving` | a success/partial-success family's eligible members all satisfy the predicate (it does not discriminate outcome) | **weaken** |
| `bias-critique` | the deterministically **recomputed** distinct-family support falls below the mining threshold; the recount is the evidence (KTD-6) and is retained for audit even when unconfirmed | **weaken** |
| `split` | ≥2 **admissible** child predicates (shared failure-mechanism + pinned-vocabulary gate), each distinct from the parent, each a **refinement** (grounding only to families the parent supports), grounding to **nonempty, pairwise-disjoint** supporting failure families (abstraction safety: ground or reject) | **weaken** + children persisted |
| `merge` | one **admissible** child predicate (same shared gate — this rejects an `outcome in [...]` child that would cover every failure family by definition) covering the **union** of the parents' support whose contrast violations are **not lower than any parent's** (a merge that erases the outcome-separating axis is rejected) | **weaken** + child persisted |
| `independent-verification` | never provider-claimable; see the `operator_attested` gate below | surviving → **operator_attested** |

Survival is **earned, not defaulted** (G2). Each attack now yields a typed
`CheckOutcome` — `inadmissible`, `inconclusive`, `completed_negative`, or
`confirmed` — instead of a default-true "applicable" flag. A campaign transitions
the invariant to **surviving** only when at least one **`completed_negative`**
attack ran a real determination over an eligible population and did not land (a
decisive negative). Validation rejections (an `inadmissible` split with fewer
than two children, a merge naming no partners, a synthetic with no construction,
a provider over-claiming operator-only verification) and undecided attacks
(`inconclusive`) count toward neither survival nor refutation. A campaign made up
entirely of inadmissible/inconclusive attempts opens the campaign
(`-> challenged`) and stops there — it does **not** earn `surviving`, and because
`challenged` is resumable a later campaign can still decide it. Survival cannot be
requested.

Confirmed split/merge children are persisted as **real candidate invariants**
in a new revision whose derivation identity (the `miner_version` reuse key)
**folds the relation, the parent invariant set, and the canonical child
predicate fingerprints** (F5), so two different parents (or child sets) splitting
under the same failure space + threshold never collide on the reuse key. Support
is recomputed by the same engine mining uses; children are linked to the parent
through `invariant_lineage`, are written in the **same campaign transaction**
(no orphaned children on a later failure), enter `proposed`, and must survive
their own challenges — no inherited authority.

## The `operator_attested` gate (KTD-2)

The DB trigger permits `surviving → operator_attested` structurally; the pipeline
gates it **epistemically**. `newf invariant establish <id> --snapshot <snap>
--locator <loc>` refuses:

- any state other than `surviving`;
- any request without a persisted source snapshot + locator (independent,
  non-model evidence, recorded as `independent_source` challenge evidence);
- any snapshot that does not belong to the invariant's problem (evidence is
  problem-scoped; ingest the source under this problem first).

The state records a provenance-bearing OPERATOR ATTESTATION, not machine
confirmation — hence `operator_attested`, never "established".

No provider path reaches `operator_attested`: a challenger proposing
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

All support `--json`. `invariant list --state surviving|operator_attested` is the
frontier read surface for M5.1 (both states are legal targets).

## Boundaries

- `internal/invariant/challenge.go` — pure verifiers (no SQL/Cobra/provider).
- `internal/provider/invariant_challenger.go` — `Challenger` interface +
  deriving fixture.
- `internal/store/challenge_store.go` + migration v13 — atomic campaign
  persistence, ledger, read surfaces.
- `internal/pipeline/challenge*.go` — campaign orchestration, verdict policy,
  the `operator_attested` gate, run lifecycle.
- `cmd/newf/challenge.go` — thin CLI wiring.
