# 09 — The Falsification Review Primitive

*Operational layer of the [theory series](./). Where
[02](02-epistemic-model.md) states what we may believe and how belief
strengthens, this document names the single bounded operation that belief
must survive: the **falsification review**. The `challenge` operator on
candidate invariants ([02](02-epistemic-model.md),
[03](03-shape-guided-search.md)) and the standing falsification conditions
FC1–FC5 ([claim registry](00-paper-claims.md#falsification-conditions-for-gow-stage-8-material))
are both instances of this one primitive.*

**Epistemic status:** `method`. This document specifies a procedure a reader
could execute; it makes no empirical claim that the procedure has been run at
scale. Its constraints restate — in one operation-shaped contract — discipline
that already exists in [AGENTS.md](../../AGENTS.md) (challenge probes,
verification hierarchy, no silent promotion), the
[epistemic model](02-epistemic-model.md) (verdict lifecycle), and
[08](08-earning-operational-authority.md) (temporal anti-mythology rule).

---

## One-sentence definition

> **A falsification review is a bounded operation that takes a scoped claim
> with its epistemic status, derives the cheapest commitments the claim
> cannot survive being wrong about, executes the strongest available probe
> against one of them, and records the landing — where the only permitted
> outcomes strengthen the record, never the claim by default.**

The primitive's defining property is **asymmetry**: it can *damage* a claim on
its own authority, but it can never *promote* one. Survival is an event in the
record, not a credential.

## Signature

```text
falsify_review(claim, scope, budget) →
    verdict     ∈ {survives, weakened, split, falsified, undecidable}
  + commitment  (what the claim was held to, frozen before observation)
  + probe       (what was executed, at which verifier strength)
  + landing     (what was observed, with provenance)
  + delta       (how the claim's statement or scope changed, if it did)
  + policy_hint (how search policy should respond)
```

The five verdicts are the challenge lifecycle already in the
[epistemic model](02-epistemic-model.md) — which is the point: `challenge` on
a `CandidateInvariant` becomes *one instantiation* of the primitive, not a
special mechanism.

## The contract

### 1. Precondition — the claim must be reviewable

A claim enters review only with:

- a **scope**: what class of cases it quantifies over;
- an **epistemic status**: `hypothesis`, `GeneratedInterpretation`,
  `CandidateInvariant`, …;
- **provenance**: where the claim came from and what it rests on.

A claim that cannot state what would count against it is not rejected — it is
returned as `undecidable` with the *missing commitment* named. That return
value is itself informative: per
[08](08-earning-operational-authority.md), unfalsifiable-as-stated is the
failure mode that lends the rest of the theory hidden authority.

### 2. Commitment extraction — before any probe runs

The reviewer derives concrete, discriminating predictions from the claim and
**freezes them before observing outcomes**. This is
[08](08-earning-operational-authority.md)'s temporal anti-mythology rule
imported into review: a claim may use the last failure to learn, but
explaining it after the fact scores nothing.

For an invariant-shaped claim, the commitment menu is the existing seven-probe
discipline ([AGENTS.md](../../AGENTS.md)): known violator, synthetic violator,
preserving success, split under lower abstraction, merge under higher,
sampling bias, correlation-vs-obstruction. For other claim shapes the menu
differs, but the rule is constant: **no commitment, no review**.

### 3. Probe selection — cheapest discriminating, strongest verifier

Two orderings govern the choice:

- **Discrimination per cost.** Prefer the probe most likely to distinguish
  the claim from its nearest rival explanation — not the probe most likely to
  be passed. (This is the probe/navigate distinction of the
  [field guide](../thesis/when-to-map-the-work.md), applied to review.)
- **Verifier strength.** Execute at the strongest available rung —

  ```text
  formal proof / deterministic check
  > reproducible computation or experiment
  > independently sourced evidence
  > cross-model or independent critic agreement
  > single-model judgment
  ```

  — and **record the rung**, because a verdict is only as strong as its
  verifier. A survives-at-model-judgment is a different object from a
  survives-at-deterministic-check.

### 4. Postcondition — the record moves; status only moves down

- `falsified`, `weakened`, `split` — update the claim and its scope.
- `survives` — appends to the survival record and **does nothing else**.
  Promotion (e.g. to `EstablishedInvariant`) requires *independent* evidence
  through a separate operation, never accumulation of surviving reviews.
- `undecidable` — charges the budget and flags the claim. Repeated
  undecidability is a stopping-condition signal (`verification_blocked`), not
  a free pass.

## What makes it a *primitive*

**Composable and self-applicable.** The same operation runs on:

| Target | The review asks |
|---|---|
| Candidate invariant | Does a violator exist — known, synthetic, or across abstraction levels? |
| Abstraction | Does the grounding loop produce a concrete prediction that fails? ([abstraction safety](../abstraction-safety.md)) |
| Frontier proposal | Is the claimed structural violation real, and is the cheapest falsification path actually cheapest? |
| Representation in \(\mathcal M(W)\) | Do two reasonable bases produce incompatible consequential predictions? (FC1 in [08](08-earning-operational-authority.md)'s operational form) |
| GoW itself | FC1–FC5 are standing falsification reviews with the *project* as the claim; the [composite attack surface](../research/composite-attack-surface.md) A1–A10 is their queue. |

**Budgeted.** It consumes a declared budget and must terminate with a
verdict — including `undecidable` — rather than escalating into open-ended
investigation.

**Provenance-complete.** Everything needed to re-run or audit the review
(claim version, frozen commitments, probe, verifier rung, landing) persists;
a review that lives only in a transcript did not happen.

## A one-line worked instance

In the [teaching model](../thesis/why-map-the-work.md): claim = "completion
time is governed by gate scope"; frozen commitments = *10 ticks under
per-stream gate* **and** *no gain under single stream*; probe = deterministic
execution (top verifier rung, the
[checked implementation](../thesis/examples/why-map-the-work/)); landing =
both predictions hold; verdict = `survives` — and the claim's published form
is still the *narrower* one, because survival bounded the scope rather than
inflating it.

## Relationship to the rest of the series

- **[02-epistemic-model](02-epistemic-model.md)** supplies the verdict
  lifecycle and the belief rules the primitive enforces; this document is the
  operation-shaped restatement.
- **[03-shape-guided-search](03-shape-guided-search.md)**'s `challenge` step
  is the primitive instantiated on candidate invariants.
- **[08-earning-operational-authority](08-earning-operational-authority.md)**
  supplies the temporal rule (freeze commitments before observation) and the
  reason `undecidable` must be a first-class verdict.
- **[00-paper-claims](00-paper-claims.md)** FC1–FC5 are the primitive aimed
  at the project's own claims; the
  [composite attack surface](../research/composite-attack-surface.md) is the
  standing review queue.
- **[AGENTS.md](../../AGENTS.md)** contributes the seven challenge probes,
  the verification hierarchy, and the no-silent-promotion invariant the
  postcondition encodes.

