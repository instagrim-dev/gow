# 01 — Semantic Model

*Layer 1 (semantic) of the [theory series](./). Where
[00-geometry-of-work.md](00-geometry-of-work.md) introduces the space and its
governing principle, this document develops the full **descriptive** vocabulary:
the objects, their fields, and their relationships — what the objects **mean**,
independent of what we are allowed to believe about them (that is the
[epistemic model](02-epistemic-model.md)) and independent of how they are stored
(that is the [domain model](../domain-model.md)).*

Every conceptual term below maps to a legacy implementation term in the
[glossary](glossary.md).

---

## The descriptive/epistemic split

The single most important thing the semantic model does is **separate
description from belief**. A `StructuralShape` describes a piece of work; it does
not assert that the work is good, that its structure is conserved, or that the
structure obstructs progress. Those are further, stronger claims layered on top,
and each gets its own object so it can carry its own evidence and its own
epistemic state.

Reading order of increasing commitment:

```text
StructuralShape          "this work has this structure"           (description)
  + Conditioning         "observed over this outcome population"   (context)
  = ShapeClaim           "this structure is conserved across it"   (hypothesis)
  + explanatory step     "this structure prevents progress"        (ObstructionHypothesis)
```

Nothing in the semantic layer permits jumping to a later line. That discipline
belongs to the [epistemic model](02-epistemic-model.md); the semantic model just
makes sure each line is a *different object*.

## StructuralShape

The descriptive structure of a piece of work, outcome-blind.

Conceptual fields (the legacy `MechanismSignature` / `Mechanism` realize these):

- **representation** — the objects the work reasons over.
- **assumptions** — what the work takes as given.
- **operators** — the moves the work performs.
- **preserves** — the properties the work keeps invariant.
- **axes** — versioned mechanistic-diversity coordinates (local/global,
  constructive/existential, deterministic/probabilistic, …).

Two rules make shape a usable coordinate rather than prose:

1. **Models discover candidate labels; software owns identity.** A provider may
   propose the surface labels for these fields, but they are deterministically
   resolved against a versioned vocabulary to stable canonical IDs, and the
   whole shape gets an order-independent fingerprint. An unmatched label stays
   `unknown` / `novel_candidate` / `ambiguous`; it is never coerced to the
   nearest term.
2. **Shape is compared component-wise.** Two shapes are compared per field, never
   by an opaque scalar, yielding a mechanistic-vs-surface classification.

Surface novelty (new wording) is not shape novelty (new structure). This is the
foundation of *mechanistic novelty over surface novelty*.

## Regime

A region of shape-space with an associated outcome distribution. Regimes are
built by grouping shapes into mechanistic families (deterministic,
profile-driven, inspectable — not embeddings) and then partitioning by outcome.

- A family all of whose members failed is in the **failure regime**.
- A family all of whose members progressed is in the **success regime**.
- A family spanning both outcomes is **`mixed`** and must be split member-wise
  before any claim rests on it. Compressing a mixed family to its
  representative's outcome is a discrimination-loss error and is guarded against.

The point of regimes is that they let the system state not "we have 42 failures"
but "we have 7 mechanistically distinct failure families, with these gaps and
these boundaries."

## ShapeClaim

A claim that some `StructuralShape` (or a boolean composition of shape
conditions) is **conserved across a population**. Its durable identity is a
machine-evaluable typed predicate; the prose statement is a human render of the
predicate, not its definition.

A `ShapeClaim` is inseparable from its **Conditioning** — the population it was
observed over:

- **failure coverage** — the fraction of the failure/partial-failure families
  whose members satisfy the predicate.
- **contrast** — the behavior of the predicate over the partial-success/success
  families.

The same predicate observed over different populations is a different claim.
"Confined to residue-local reasoning across the failed families" and "…across
all families" are not the same statement, and the second is not the first with
more confidence. Support is a **mechanistic-non-redundancy** count of distinct
families — never a claim of statistical or historical independence.

A `ShapeClaim` carries a descriptive `association_status`:

- **recurring** — code-assigned from measured failure coverage.
- **discriminative** — code-assigned from measured contrast (it separates
  outcomes).
- **candidate_obstruction** — see below; recorded only as a flagged model
  hypothesis.
- **unknown** — on ambiguity.

> **Note on axes.** A `ShapeClaim` is *not* a point on a single ladder. It is a
> claim along three orthogonal axes — `shape × observed_regime ×
> epistemic_status` — with `claim_role` (regularity / obstruction /
> enabling_condition / boundary_hypothesis) recorded separately again. The
> [epistemic model](02-epistemic-model.md) develops these; here it is enough
> that *conditioning is a distinct axis from the shape itself*. "Confined to
> residue-local reasoning" is one shape; observing it `failure_conditioned`
> versus `success_conditioned` are different claims about the same shape.

## ObstructionHypothesis

The stronger, **explanatory** claim that a conserved shape *prevents* progress —
that it is not merely present in the failure regime but is why that regime
fails.

This is a genuinely different object from a `ShapeClaim`, and the semantic model
refuses to let a recurrence masquerade as it. `observed regularity ≠
obstruction`. In the implementation this is exactly why `candidate_obstruction`
is never code-assigned from measured quantities: obstruction is an interpretation
a model may *flag*, carried as a hypothesis, and it earns no additional authority
from the fact that the underlying shape recurs.

## Boundary

A structural condition that separates regimes or limits a claim.

Three concrete kinds appear across the system:

- a **`FailureBoundary`**: the normalized obstacle a failed attempt hit.
- a **counterexample**: a member (real or constructed) whose shape *violates* a
  `ShapeClaim`, marking where the claim stops holding.
- a **contrast boundary**: the structural difference between a failure family
  and a success family that outcome tracks.

Boundaries are the features a search wants to cross.

## StructuralDelta and Projection

- A **StructuralDelta** is the proposed structural change that would cross a
  boundary: the violation a frontier proposal claims, stated as *which* conserved
  property it breaks and *why the break is structural rather than cosmetic*. When
  the boundary being crossed is the difference between the failure regime and a
  success regime, the delta is exactly the regime boundary hypothesis
  **`Δ(F, S)`** — the `claim_role = boundary_hypothesis` object of the
  [epistemic model](02-epistemic-model.md).
- A **Projection** is the delta compiled into the problem's classical
  representation — a concrete candidate a verifier can decide. See
  [04-classical-projection.md](04-classical-projection.md).

The delta lives in shape-space; the projection is its image in domain-space.
A `boundary_delta` produced by a *challenge* is the same kind of object observed
in reverse: the minimal condition separating a counterexample from the support
population. Both are **Boundary** measurements — one proposed to be crossed, one
observed while defending a claim.

## Success condition `C`

Progress is rarely "just break `P`." The valuable object is usually the
*condition* under which breaking a failure invariant `P` yields progress:

```text
failure invariant:  P is preserved across the failed families
success invariant:  progress appears when P is broken under condition C
```

`C` is a `ShapeClaim`-shaped object conditioned on the **break cohort**
(work that crossed the boundary), separating progressors from non-progressors.
Merely breaking `P` may be necessary but not sufficient; `C` is what makes it
sufficient, and it is often more useful than the raw symmetry break.

## LandingPoint

The observed result of executing/verifying a projection: a verdict together with
the **verifier used** and the **verification strength** of that verifier. A
landing point is not just "it worked / it didn't"; it is a stamped measurement
whose epistemic weight is inseparable from its verdict.

A failure/partial-failure landing point re-enters the atlas as a new observed
shape, expanding the geometry. This closes the loop from search back to
description.

## Relationship map (conceptual)

```text
StructuralShape --grouped/partitioned--> Regime
StructuralShape --conserved across?----> ShapeClaim --conditioned by--> Conditioning
ShapeClaim -----explanatory upgrade?---> ObstructionHypothesis (flagged, not derived)
ShapeClaim -----limited/separated by---> Boundary
Boundary -------crossed by-------------> StructuralDelta --compiled to--> Projection
Projection -----verified at------------> LandingPoint --re-enters--> StructuralShape (atlas)
break cohort ---compressed to----------> success condition C (paired with broken P)
```

## What the semantic model refuses to do

- It refuses to store a shape and its outcome as one inseparable thing.
  Conditioning is always an explicit, separable act.
- It refuses to treat surface restatement as new structure.
- It refuses to let compression stand as evidence of equivalence — two families
  that merge under an abstraction that erased the outcome-separating axis are not
  thereby the same (see [abstraction safety](../abstraction-safety.md)).
- It refuses to promote recurrence to obstruction.

Each refusal is a semantic boundary that the [epistemic model](02-epistemic-model.md)
then enforces as belief discipline.
