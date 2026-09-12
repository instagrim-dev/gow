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
| `known-counterexample` | a failure-side member in the atlas evaluates to `violates` (ambiguity is not a counterexample) **and** the candidate carries an **operator-authored** `universal` claim form (v39, issue #21); recurrence is a *frequency label*, not a quantifier, and — since v39 — measured full coverage of a finite sample no longer supplies an unstated universal domain. Absent an authored `universal` quantifier (`newf invariant claim`), a `recurring` candidate, a `contrast_observed` (association) claim, or an `unknown`-kind claim is **not** refuted by an isolated counterexample (G3): the violation is recorded, the claim is not refuted | **falsified** (authored-universal only) |
| `synthetic-counterexample` | a provider-constructed approach evaluates to `violates` **via admissibly supported structure**. The construction is a PROPOSAL: its set fields are `unobserved`, so an omitted/empty description reads as `unknown` (inert), and a violation resting on an **unsupported** present claim is likewise only a proposal — presence in a generated description is not stronger evidence of realizability than absence from it (G3) | **weaken** — constructibility (from supported structure) shows the invariant is not conserved by necessity; only an observed, in-atlas counterexample falsifies the empirical regularity |
| `success-preserving` | a success/partial-success family's eligible members all satisfy the predicate (it does not discriminate outcome) | **weaken** |
| `bias-critique` | the deterministically **recomputed** distinct-family support falls below the mining threshold; the recount is the evidence (KTD-6) and is retained for audit even when unconfirmed | **weaken** |
| `split` | ≥2 **admissible** child predicates (shared failure-mechanism + pinned-vocabulary gate), each distinct from the parent, each a **refinement** (grounding only to families the parent supports), grounding to **nonempty, pairwise-disjoint** supporting failure families (abstraction safety: ground or reject) | **weaken** + children persisted |
| `merge` | one **admissible** child predicate (same shared gate — this rejects an `outcome in [...]` child that would cover every failure family by definition) covering the **union** of the parents' support whose contrast violations are **not lower than any parent's** (a merge that erases the outcome-separating axis is rejected) | **weaken** + child persisted |
| `independent-verification` | never provider-claimable; see the `operator_attested` gate below | surviving → **operator_attested** |

Survival is **earned, not defaulted** (G2/H4). Each attack yields a typed
`CheckOutcome` — `inadmissible`, `inapplicable`, `inconclusive`,
`completed_negative`, or `confirmed` — instead of a default-true "applicable"
flag. A campaign transitions the invariant to **surviving** only when at least one
**`completed_negative`** attack ran a real determination over a **nonempty**
eligible population, **decisively resolved every applicable case**, and did not
land (a decisive negative). The three ways an attack can fail to confirm are held
distinct and none of them strengthen the hypothesis:

- **`inadmissible`** — the attack could not be evaluated at all (an `inadmissible`
  split with fewer than two children, a merge naming no partners, a synthetic with
  no construction, a provider over-claiming operator-only verification).
- **`inapplicable`** — the attack was well-formed but its eligible population was
  **empty** (zero known failure-side members to check; zero success-side contrast
  members). Zero eligible cases decide nothing.
- **`inconclusive`** — the attack ran but one or more eligible cases evaluated
  `unknown`, so the negative was never fully resolved (an unsupported synthetic
  proposal is also inconclusive).

An empty population and an unknown-only population are therefore **never** silently
counted as a completed negative. A campaign made up entirely of
inadmissible/inapplicable/inconclusive attempts opens the campaign
(`-> challenged`) and stops there — it does **not** earn `surviving`, and because
`challenged` is resumable a later campaign can still decide it. Survival cannot be
requested.

## Discovery vs assessment population (v34/S1)

A challenge campaign runs over **two explicitly identified populations**, fixing
the 2026-09-12 structural review's finding S1 (the map could advance while an
invariant's search authority stayed justified against an older population):

- **discovery population** — the cluster run the candidate's mining revision was
  derived over. This is the claim's *scope*: what the recorded statement is
  about.
- **assessment population** — the evidence the campaign's searches actually ran
  against. Under the default `--population latest` policy this is the newest
  cluster run with the **same schema and vocabulary versions** as the discovery
  run (predicate comparability is a precondition for widening evidence, never
  silently substituted); `--population discovery` replays the original
  population exactly.

Attacks route by what their claim is about:

| Attack | Population | Why |
|---|---|---|
| `known-counterexample`, `success-preserving` | **assessment** | evidence searches: "does *current* evidence contain a violator / a preserving success?" |
| `bias-critique`, `split`, `merge`, synthetic grounding | **discovery** | claim-scope operations: recounting support or partitioning the claim's own sample over a different population would conflate two claims |

A confirmed counterexample is additionally **scope-classified** by exact
signature membership in the discovery run:

- an **in-scope** violator (a member of the population the universal claim was
  made over) **falsifies** the historical claim itself;
- a violator found **only in the assessment expansion** does **not** rewrite
  history: the bounded statement about the discovery population stands, its
  *generalization* to current evidence is refuted, and the invariant transitions
  to **weaken** — dropping frontier eligibility without fabricating a
  retroactive falsification.

Membership is exact signature identity: a revised interpretation of an
in-discovery approach carries a new signature id and counts as new evidence,
because the claim was made over the signatures actually recorded.

Every population-assessed campaign persists its identity in
`challenge_assessment_populations` (`run_id`, `invariant_id`, both cluster-run
ids, the requested policy; immutable rows), so a reader can always tell which
evidence a campaign searched — it is never inferred from the mining revision.
Attestation campaigns (`invariant establish`) search no population and persist
no row.

**Recorded limitation:** survival re-earned by a later campaign is stamped with
that campaign's population; a `discovery` replay of a weakened invariant can
re-earn `surviving` against the old population. The transition ledger plus the
population rows keep this auditable, but the state summary alone does not rank
populations by recency — readers of `surviving` should check the latest
campaign's assessment population.

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
Migration v34 adds `challenge_assessment_populations` (immutable; one row per
population-assessed campaign) recording the discovery/assessment cluster-run
ids and the requested policy inside the same transaction.

Migration v36 (structural review S5, part A) adds `challenge_boundary_deltas`
(immutable; one row per CONFIRMED challenge, written in the same campaign
transaction): the **typed** boundary refinement the deterministic verifier
derived, so subsequent policy consumes a typed object instead of
reinterpreting a transcript. `kind` names what was observed —
`counterexample-separation` (the claim predicate is the separating condition;
the recorded counterexample members sit on its violating side),
`contrast-collapse` (the predicate fails to discriminate outcome),
`constructibility` (a supported synthetic violates it), `support-recount`
(measured support + threshold), `split-partition` / `merge-union` (ordered
derived-child predicate fingerprints) — and `condition` is the canonical,
paraphrase-stable serialization of the predicate. Disposition stays on the
challenge's persisted transitions; derived candidate ids stay on its
derived-children rows. Unconfirmed/inert challenges derive no delta. The
delta surfaces in `--json` as `challenges[].boundary_delta`.

> **Current limitation (recorded, 2026-09-12 review F10):** boundary deltas
> are persisted and surfaced but **no search-policy mutation consumes them
> yet** — "subsequent policy consumes a typed object" states the design
> intent, not shipped behavior. Until a policy revision reads
> `challenge_boundary_deltas`, the next-decision edge exists only as durable
> state.

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
newf challenge <invariant-id> [--population latest|discovery]
newf challenge --problem <id> --all [--population latest|discovery]
newf invariant state <invariant-id>
newf invariant list --problem <id> --state surviving
newf invariant establish <invariant-id> --snapshot <snap-id> --locator <loc> [--note ...]
```

All support `--json`. `invariant list --state surviving|operator_attested` is the
frontier read surface for M5.1 (both states are legal targets).

## `boundary_delta` — first-class challenge product

A challenge does not only pass or fail. The most valuable challenge outcome is
**narrowing the condition under which the parent invariant stops being true**.
That narrowed condition is a `boundary_delta`:

```text
InvariantBoundary: the minimal structural difference between the
    support population and the counterexample (or the split child)
    under which the parent claim ceases to hold.
```

This is distinct from both the counterexample itself and any derived child
hypothesis:

```text
counterexample
    ≠ boundary_delta (the counterexample is an instance; the boundary is the
                      minimal condition distinguishing it from the support)

boundary_delta
    ≠ child invariant (the delta is an observation about I; the child is a
                       new proposed claim that must be independently grounded)

repeated/grounded boundary_delta
    → candidate child hypothesis (enters proposed, not inherited state)
```

The reason for keeping these separate is anti-fractal discipline: a single
counterexample is not a license to instantiate ten new child hypotheses.
The sequence is:

```text
challenge(I)
→ boundary_delta: "what minimally distinguishes the counterexample from the
                   support population?"
→ derive child hypothesis if and only if the boundary is specific, grounded,
  and non-trivially distinct from the parent
→ child enters proposed with no inherited authority
→ challenge(child) independently
```

### `ChallengeResult` structure

A complete challenge result carries:

```text
ChallengeResult {
    invariant_id:         ID of the challenged invariant
    counterexample:       the specific failing instance (if found)
    boundary_delta:       the minimal condition separating the counterexample
                          from the support population
    disposition:          survive | weaken | split | falsify
    derived_candidate_ids: child candidates generated from the boundary
                           (empty unless disposition is split/weaken with a
                           specific boundary)
}
```

`disposition` maps to lifecycle transitions as before; `boundary_delta` is an
additional field that narrows what the transition *means*, not the transition
machinery itself.

### The recursive refinement loop

A challenge that produces a `boundary_delta` does not close the research loop —
it **opens a smaller one**:

```text
failure-space F0
    ↓ compress
candidate invariant I0
    ↓ challenge
    ├── survive:  I0 stands; boundary_delta = "no distinguishing condition found"
    ├── weaken:   I0 narrows; boundary_delta names the excluded condition;
    │             derived child H1 (proposed, no inherited status)
    ├── split:    I0 → {I0a, I0b}; boundary_delta names the bifurcation axis
    └── falsify:  I0 refuted; boundary_delta names the decisive counterstructure

    ↓ (if weaken or split)
micro-failure-space F1 (the boundary_delta's support population)
    ↓ compress
candidate invariant I1
    ↓ challenge
    ...
```

A `weaken` with a sharp `boundary_delta` is **not a disappointment** — it is the
mechanism by which the failure-space representation recursively gains resolution.
The challenge discipline was already designed for this: `split` and `weaken` were
always intended to produce a narrower surviving claim. `boundary_delta` makes the
narrowing artifact explicit.

### Resistance to hypothesis proliferation

The child hypothesis discipline prevents unbounded branching:

- A `boundary_delta` must be specific (names a structural condition, not "more
  research needed").
- A derived child must be non-trivially distinct from the parent (a paraphrase
  is not a child).
- Each child enters `proposed` and must survive its own challenge. No authority
  is inherited from the parent surviving a challenge.
- A boundary that requires more corpus data to test should be recorded as
  `challenged` (open campaign), not converted immediately into a new candidate.

## Boundaries

- `internal/invariant/challenge.go` — pure verifiers (no SQL/Cobra/provider).
- `internal/provider/invariant_challenger.go` — `Challenger` interface +
  deriving fixture.
- `internal/store/challenge_store.go` + migration v13 — atomic campaign
  persistence, ledger, read surfaces.
- `internal/pipeline/challenge*.go` — campaign orchestration, verdict policy,
  the `operator_attested` gate, run lifecycle.
- `cmd/newf/challenge.go` — thin CLI wiring.

## Authored claim forms (v39, issue #21)

Sample recurrence, transformation invariance, and obstruction are distinct
propositions. Before v39, `associationKindForCandidate` granted the universal
(falsifiable-by-one-counterexample) reading from *measured* full coverage over
the discovery population — a property of a finite recorded sample silently
fixing the refutation semantics of an unauthored proposition.

Since v39 the proposition shape is **authored**, never inferred:

```text
newf invariant claim --invariant inv_... \
  --quantifier universal|recurrent|existential \
  --role regularity|obstruction|enabling_condition|boundary_hypothesis \
  --scope "failure families of cluster run mcr_...; no claim beyond the corpus" \
  --note "basis for authoring"
```

- Forms live in `invariant_claim_forms` (append-only, immutable, latest form
  governs deterministically by `created_at, id`).
- The full spec is `predicate` (mined) + `quantifier` + `scope` + `claim_role`
  (authored) + assessment context (v34 discovery/assessment populations).
- **Universal treatment comes only from an authored `universal` quantifier.**
  Absent one, a `recurring` candidate keeps association semantics: a lone
  counterexample is recorded but inconclusive, and drives no transition.
- Authoring `universal` does not strengthen evidence — it makes the claim
  *more falsifiable* and records who fixed its shape and why. Truth (the
  recorded observation), discrimination (whether the property separates
  failures from successes — attacked by `success-preserving`), and search
  eligibility (lifecycle state) remain separate axes.
- `newf invariant state` shows the governing form, or states explicitly that
  none is authored.

Challenge evidence rows also carry `verification_subject` (v38): the
deterministic challenge verifiers run predicates over persisted normalized
signatures, so their evidence certifies the **annotation** — never
realizability or the domain goal. Operator-attested independent sources
(`newf invariant establish`) record `domain-goal` as their subject.

Regression: `TestIntegrationChallengeAssessmentPopulation` demonstrates the
gate end-to-end — the same in-scope violator is inconclusive before authoring
and weakens the invariant only after the operator authors the universal claim.
