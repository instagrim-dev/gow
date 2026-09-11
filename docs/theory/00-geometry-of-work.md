# 00 — Geometry of Work

*Semantic model. Layer 1 of the [theory series](./). This document defines the
conceptual center of `newf`: the space it searches, its objects, and the single
governing principle. It is intentionally about **meaning**, not schema — see the
[glossary](glossary.md) for the mapping to legacy implementation terms, and the
[domain model](../domain-model.md) for the concrete types.*

---

## The reframing

A hard problem, as conventionally represented, bundles two things:

1. the **object** being solved (the conjecture, the equation, the target); and
2. the **history of attempts** to solve it.

Direct search operates inside (1). `newf` takes (2) — the attempts, failures,
partial successes, and successes — and treats it as a structured object in its
own right: a **space of work**.

The claim is that this space has *geometry*. It has regions, boundaries, and
transitions, and those features are informative about the problem even before
the problem is solved.

## The objects of the geometry

The geometry is built from a small number of concepts. Each has a precise
meaning here and a legacy implementation term in the [glossary](glossary.md).

### Work-space (shape-space)

The space induced by all the work done on a problem. A point is a piece of work
described by its **shape**, not its outcome and not its surface prose.

`newf` searches *in this space*. This is what "search in shape-space" names.

### Shape

The **descriptive structure** of a piece of work, independent of whether it
succeeded: its representation, assumptions, operators, and preserved properties.

Two attempts with the same shape are the same point even if they are written in
different notation; two attempts with different shapes are different points even
if they used the same words. Shape is deliberately *outcome-blind* — attaching
the outcome is a separate act (see *Conditioning* in the
[epistemic model](02-epistemic-model.md)).

### Regime

A **region of shape-space associated with an outcome distribution**. The failure
regime is the region occupied by failed and partial-failed work; the success
regime is the region occupied by partial-success and success work. A region that
mixes both is `mixed` and must be split member-wise before it can support a
claim.

Regimes are how "some shapes fail and some shapes make progress" becomes a
statement about *where in the space* you are.

### Boundary

A **structural condition that separates regimes**, or that limits the scope of a
claim about a region. A counterexample is a boundary: it marks where a claimed
conserved property stops holding. A contrast between a failure family and a
success family is a boundary: it marks the structural difference that outcome
tracks.

### Trajectory

A path through shape-space traced by successive pieces of work. The history of a
problem is a set of trajectories; `newf` reconstructs their aggregate structure
rather than following any single one to its end.

### Structural delta

The **proposed change required to cross a boundary** — the specific structural
violation that would move work from the failure regime toward a more successful
regime. A frontier proposal is, conceptually, a structural delta plus the
argument that it is real rather than cosmetic.

### Projection

The **translation of a shape-space move into the problem's classical
representation**: a concrete candidate that a deterministic check, computation,
proof obligation, or independent source can actually decide. Search happens in
shape-space; a projection is how a shape-space move re-enters domain-space to be
tested.

## The governing principle

> **Search in shape-space; verify in domain-space.**

These are different spaces with different truth conditions, and keeping them
distinct is what makes the whole approach coherent:

- **Shape-space** is where the model's comparative advantage lives: compressing
  heterogeneous work into candidate conserved structure, contrasting regimes,
  proposing structural deltas. Reasoning here is allowed to be fluent and
  associative because nothing here is being *certified*.
- **Domain-space** is where truth lives: a projection is verified by the
  strongest available deterministic or independent mechanism, and its result is
  stamped with an explicit verification strength.

Fluency generates leverage in shape-space. Grounding in domain-space determines
whether that leverage still points at the original problem. Collapsing the two —
letting a confident shape-space argument stand in for a domain-space
verification — is the central error the [epistemic model](02-epistemic-model.md)
is built to prevent.

## Why geometry rather than a list

Raw attempt count is a bad proxy for knowledge. Ten syntactically different
attempts that preserve the same structure are one weak sample of the geometry,
not ten independent facts. Framing work as a *space with shape* forces the
useful questions:

- Which regions are densely occupied by failure? (bottlenecks)
- Which combinations of structure are unoccupied? (voids)
- What separates the failure regime from the nearest success? (boundaries)
- Which conserved property, if broken, would cross a boundary? (structural delta)

A flat list of approaches cannot ask these questions. A geometry can.

## What a failure *is* in this model

Because the objects are geometric, a failed candidate is not waste. A candidate
that crosses one boundary and then fails is a **measurement**: it establishes
that the transition was traversable and exposes the *next* boundary. A candidate
that fails to cross the boundary it claimed to cross is a measurement too: it
says the reconstruction of the geometry was wrong at that point.

This is why `newf` treats failure as a first-class sample space rather than a
regrettable byproduct — and why the operational loop
([03-shape-guided-search.md](03-shape-guided-search.md)) is recursive
cartography: every landing point updates the map.

## Relationship to the rest of the series

- The **[semantic model](01-semantic-model.md)** develops these objects into the
  full descriptive vocabulary and its relationships.
- The **[epistemic model](02-epistemic-model.md)** says what we are allowed to
  *believe* about points, regimes, and boundaries.
- **[Shape-guided search](03-shape-guided-search.md)** turns the geometry into an
  algorithm.
- **[Classical projection](04-classical-projection.md)** governs the shape → domain
  → verification crossing.
- The **[glossary](glossary.md)** maps every term to its legacy implementation
  name so the database is never renamed by accident.
