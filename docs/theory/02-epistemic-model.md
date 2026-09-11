# 02 — Epistemic Model

*Claims, Evidence, and Promotion. Layer 2 of the [theory series](./). The
[semantic model](01-semantic-model.md) says what the objects **mean**. This
document says what we are **allowed to believe** about them, and by what
transitions belief may strengthen. It is the discipline that keeps the geometry
honest.*

Legacy implementation terms are in the [glossary](glossary.md); the concrete
lifecycle machinery is documented in
[invariant-challenge.md](../invariant-challenge.md).

---

## Why this layer exists

Searching in shape-space is fluent and productive precisely because it is not
being certified as it happens. That same fluency is dangerous: a coherent
structural argument *feels* like knowledge. The epistemic model exists to make
sure it is never *counted* as knowledge until something other than fluency has
confirmed it.

The whole model reduces to keeping four pairs of things separate, and defining
the only legal ways to move between them.

## A claim is three orthogonal axes, not one ladder

Before the separations, fix the shape of a claim. A `ShapeClaim`
(`CandidateInvariant`) is **not** a point on a single confidence ladder. It is a
claim along **three independent axes** that must never be collapsed into one
enum:

```text
Invariant = shape × observed_regime × epistemic_status
```

- **shape** — the conserved structure being asserted (see the
  [semantic model](01-semantic-model.md)). It does not change when confidence
  changes or when scope narrows.
- **observed_regime** — the population the shape was observed over:
  `failure_conditioned`, `success_conditioned`, `mixed`, or (future)
  `frontier_conditioned`. This is the *Conditioning* axis.
- **epistemic_status** — how much pressure the claim has survived: `proposed`,
  `challenged`, `surviving`, `weakened`, `falsified`, `operator_attested`.

The axes are independent, and two consequences follow immediately:

- **"Candidate" is a status word (`proposed`), not a kind of invariant.** A
  failure-conditioned shape at `surviving` is the *same shape claim* as at
  `proposed`, with more evidence — not a stronger species of thing.
- **A success-conditioned shape is not a "positive invariant" on a different
  ladder.** It is the same type of record with a different `observed_regime`.

The four separations below are the discipline that keeps these three axes from
silently bleeding into one another.

## The four separations

### 1. Shape ≠ outcome-conditioning

A `StructuralShape` is outcome-blind. Attaching a population ("observed across
the failed families") is a separate, explicit act. Therefore:

- A shape claim's meaning **includes** its conditioning. The predicate plus
  "over the failure regime" is a different belief from the same predicate "over
  all regimes."
- **`candidate ≠ failure-conditioned`.** A frontier candidate is a *proposal
  about shape-space*, not a member of the observed failure population. It is
  never allowed to be counted as evidence for a failure invariant. A generated
  attempt is a `SyntheticArtifact`; it cannot be promoted into the observed
  atlas without independent source-backed evidence.

### 2. Observed regularity ≠ obstruction

That a property recurs across the failure regime is a **descriptive** fact.
That the property *prevents* progress is an **explanatory** claim. The system
records the interpretive standing of a shape claim on a separate axis,
`claim_role`, precisely so recurrence cannot masquerade as obstruction:

```text
regularity          — shape conserved across the sampled population; no causal
                      reading claimed. CODE-ASSIGNED ONLY.
obstruction         — regularity + discrimination against successes + challenge
                      survival; the shape may be causally blocking progress.
enabling_condition  — regularity conserved across successes; the shape may be
                      necessary for progress.
boundary_hypothesis — the minimal structural difference between a failure regime
                      and a success regime; the Δ(F, S) below.
```

- Code assigns **only** `regularity` automatically. Every other role requires
  explicit evidence and operator/model attestation. `ModelJudgment ≠ Verification`.
- In the implementation this is why `candidate_obstruction` is never derived
  from measured quantities: obstruction is an interpretation a model may *flag*,
  carried as a hypothesis, and it earns no authority from the fact that the
  underlying shape recurs.

Correlation between a shape and the failure outcome is not causal obstruction.
Distinguishing them requires ruling out sampling and publication bias, which is
one of the challenge operations below.

#### Δ(F, S): the regime boundary hypothesis

The most actionable derived object is not a failure-shape or a success-shape on
its own, but the **structural difference between them**:

```text
Δ(failure_conditioned_shape F, success_conditioned_shape S)
    = "what changed between mechanisms that stay trapped and mechanisms that
       make progress?"
```

This is the `boundary_hypothesis` role, and it is the epistemic name for what
[shape-guided search](03-shape-guided-search.md) calls a **StructuralDelta** and
what a separatrix reading of the geometry points at. After independent challenge
and survival, `Δ(F, S)` becomes a candidate obstruction/enabling-condition pair
— which may be the actual structural content of a "big move" on the problem. It
is `proposed` like anything else and inherits no authority.

### 3. Model judgment ≠ verification

A provider *proposes*; code or an independent mechanism *confirms*.

- When a challenger claims an attack lands, pure code-owned verifiers re-run the
  determination against persisted shapes. A claim that code **cannot** confirm
  is recorded inert (`unconfirmed`): it links no evidence and drives no state
  change. A hypothesis is neither strengthened nor weakened by a claim code
  cannot check.
- A verdict is always stamped with a **verification strength**, drawn from the
  hierarchy below. A verdict without its strength is meaningless; the two are
  stored as one inseparable fact.
- Consensus among models — however enthusiastic, however many lanes — is still
  model judgment. Role-differentiated prompts to the same model family are
  *within-family consistency*, not independent sources.

### 4. Absence ≠ negation

An unobserved or empty field reads as `unknown` (inert), never as `false`.

- A synthetic construction's fields are `unobserved`: an omitted description is
  `unknown`, not a violation. Presence in a generated description is not stronger
  evidence of realizability than absence from it.
- An attack whose eligible population is empty is `inapplicable` (it decides
  nothing), not a confirmed negative.
- An attack that ran but left an eligible case `unknown` is `inconclusive`, not
  a completed negative.

This is the same absence-versus-evidence distinction that separates "the corpus
annotations record no broken properties" from "the mechanism breaks nothing" —
a distinction a single unanimous quorum verdict is not allowed to erase (see
[Finding 001](../findings/001-unencoded-shape-generation.md) §E15).

## The verification hierarchy

Belief strength is not a scalar the model sets; it is the class of mechanism
that confirmed the claim. Prefer the strongest available:

```text
formal proof / deterministic check
  > reproducible computation or experiment
  > independently sourced evidence
  > cross-model / independent-critic agreement
  > single-model judgment
```

Higher layers do not erase lower ones — a stored verdict keeps the exact
strength that produced it — but a stronger layer, when available, is preferred,
and a deterministic result can never be overridden by a confident model
"success." Routing is strongest-decisive-first, not cheapest-confident-first.

## Epistemic states and legal promotion

A shape claim (`CandidateInvariant`) occupies exactly one `EpistemicState`, and
state lives **only** in an append-only, trigger-guarded transition ledger — there
is no mutable state column that could drift.

```text
proposed          -> challenged
challenged        -> {challenged | surviving | weaken | falsified}
surviving         -> {challenged | weaken | falsified | operator_attested}
weaken            -> {challenged | surviving | falsified}
operator_attested -> {challenged | weaken | falsified}
falsified : terminal
```

Reading of each state:

- **proposed** — a hypothesis with computed support. Believed only as a
  candidate; influences nothing downstream.
- **challenged** — under attack, and **resumable**: a campaign that reaches no
  decisive outcome parks here, and a later, stronger campaign can reopen it. Not
  a dead end.
- **surviving** — attacked and not refuted. This is *earned, not defaulted*: a
  campaign reaches `surviving` only when at least one attack ran a real
  determination over a **nonempty** eligible population, decisively resolved
  every applicable case, and did not land. A campaign made entirely of
  inadmissible/inapplicable/inconclusive attempts opens the campaign and stops —
  survival cannot be requested. Only surviving (and operator-attested) claims may
  influence search.
- **weaken** — a non-fatal attack landed (synthetic constructibility,
  success-preservation, support collapse under redundancy, a grounded
  split/merge). The claim is not conserved by necessity. A `weaken` that carries
  a **sharp `boundary_delta`** (see below) is a *resolution gain*, not a
  disappointment: the claim's scope narrows to a more accurate region.
- **falsified** — terminal. A confirmed known counterexample refutes it, but
  **only** when the claim is a verified-universal claim over a fully-covered
  failure population. Recurrence is a *frequency label*, not a quantifier: a
  `recurring` claim that is not fully covering is not refuted by an isolated
  counterexample.
- **operator_attested** — the strongest reachable state, and still
  **challengeable**. It records that an operator attached independent,
  problem-scoped, source-backed evidence (a snapshot + locator). It is an
  **attestation of that evidence, not a machine verification of the claim
  against its predicate**, which is exactly why it is named `operator_attested`
  and never "established." No provider path reaches it; model agreement tops out
  at `surviving`.

Split and merge are **lineage relations**, not states: split produces ≥2 child
claims, merge produces one; children enter `proposed` and must survive their own
challenges — no inherited authority.

## The challenge operations (what may move belief)

Every candidate must be attacked before it controls search allocation. The
challenger proposes; code confirms. The operations:

1. **known-counterexample** — a real failure-side member violates the predicate.
   *(falsifies, verified-universal only)*
2. **synthetic-counterexample** — a constructed approach violates it via
   *admissibly supported* structure. *(weaken — constructibility ≠ observed
   refutation)*
3. **success-preserving** — a success family's members all satisfy it, so it
   does not discriminate outcome. *(weaken)*
4. **abstraction split** — the claim splits into ≥2 grounded child refinements at
   a lower level. *(weaken + children)*
5. **abstraction merge** — several claims collapse into one at a higher level
   *without* erasing the outcome-separating axis. *(weaken + child)*
6. **bias-critique** — the recomputed distinct-family support falls below
   threshold; the recount is the evidence. *(weaken)*
7. **independent-verification** — never provider-claimable; the operator-attested
   gate.

Each attack yields a typed outcome — `inadmissible`, `inapplicable`,
`inconclusive`, `completed_negative`, or `confirmed` — never a default-true
"applicable" flag. The three ways to *fail to confirm* are held distinct and
none of them strengthen the hypothesis.

## The boundary_delta: a challenge produces a measurement, not just a verdict

A challenge does not only move the epistemic status; it produces a
**`boundary_delta`** — the *minimal structural condition that separates the
counterexample (or the split axis) from the support population*. This is the
epistemic realization of a **Boundary**, and it is kept strictly distinct from
its neighbors:

```text
counterexample   ≠ boundary_delta   (an instance vs. the condition that isolates it)
boundary_delta   ≠ child claim      (an observation about I vs. a new proposed claim)
```

A complete challenge result therefore carries `{invariant_id, counterexample,
boundary_delta, disposition ∈ survive|weaken|split|falsify, derived_candidate_ids}`.
The `boundary_delta` narrows what a transition *means*; it is not new transition
machinery.

**Anti-fractal discipline.** A single counterexample is not a license to spawn
ten child hypotheses. A `boundary_delta` may license *at most* a derived child
only when it is specific (names a structural condition, not "more research
needed"), grounded to concrete support, and non-trivially distinct from the
parent. The child enters `proposed` with **no inherited authority** and must
survive its own challenge. A boundary that needs more corpus data to test is
recorded as `challenged` (open campaign), not converted into a candidate.

This is what makes challenge **recursive**: a `weaken`/`split` with a sharp
`boundary_delta` opens a smaller failure-space (the delta's support population),
which compresses to a narrower candidate, which is challenged in turn. See the
recursive refinement loop in [shape-guided search](03-shape-guided-search.md).

## Challenge and refinement semantics

- **Challenge is adversarial by construction.** Survival is the *absence of a
  landed attack after a real attempt*, not the *presence of agreement*. This is
  why an empty or unknown-only campaign cannot grant survival.
- **Refinement preserves discrimination.** A split may ground its children only
  to families the parent supported; a merge may not lower the contrast that
  separated outcomes. An abstraction that compresses beautifully but merges known
  distinct outcomes is rejected — compression is not success (see
  [abstraction safety](../abstraction-safety.md)).
- **Attestation is not immunity.** `operator_attested` remains challengeable, so
  new evidence can still reopen it.

## What "no invariant" means

A null result is a legitimate outcome, not a failure of the machine. No surviving
invariant may mean the sample is too small, the failures are not mechanistically
diverse, several unrelated mechanisms coexist, the relevant invariant lives at a
different abstraction level, or there is no compact invariant. The stopping
conditions `no_information_gain` and `insufficient_failure_diversity` name these
honestly. `no_information_gain` is **not** permission to regenerate the same
mechanism with different adjectives.

## One-line summary

> Describe freely, believe reluctantly, promote only by confirmation, and keep
> the strength of every belief attached to the belief.
