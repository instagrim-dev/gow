# 03 — Shape-Guided Search

*Operational theory. Layer 3 of the [theory series](./). The
[geometry](00-geometry-of-work.md) and [semantic model](01-semantic-model.md)
say what the objects are; the [epistemic model](02-epistemic-model.md) says what
we may believe about them. This document is the **algorithm**: how `newf` moves
through shape-space, and why the loop is recursive.*

The concrete stages that realize this loop are the EPIC milestones (M2–M6); this
document is their conceptual through-line.

---

## The recursive step

Shape-guided search is one step applied repeatedly:

```text
work
  -> geometry            reconstruct shape-space from the atlas of work
  -> boundary            locate a structural condition separating regimes
  -> structural delta    propose the change that crosses it
  -> classical projection compile the delta into a domain candidate
  -> verification        decide the candidate with the strongest verifier
  -> landing point       record verdict + verification strength
  -> geometry update     fold the landing point back into shape-space
```

Then repeat from the updated geometry. Each pass is a measurement of the terrain,
and the terrain is redrawn after every measurement. This is **recursive
cartography**: the output of search is not only a candidate but a better map.

## The step in detail

### work → geometry

Reconstruct shape-space from the observed atlas: normalize heterogeneous work
into outcome-blind `StructuralShape`s, canonicalize them to comparable
fingerprints, group them into mechanistic families, and partition those families
by outcome into **regimes**. Report coverage and gaps. The output is a stated
geometry: *N* distinct failure families, their boundaries, and where the sample
is thin.

### geometry → boundary

Mine the geometry for structure worth crossing. `ShapeClaim`s (candidate failure
invariants) are the most developed instrument, but they are **one instrument
among several** the geometry exposes:

- **clusters** — mechanistic families (regions).
- **voids** — unoccupied combinations of structure; under-sampled axes.
- **gradients** — directions along which outcome improves (contrast between a
  failure family and a nearer success family).
- **bridges** — structure shared between a failure region and a success region.
- **bottlenecks** — boundaries many independent failures pile up against.
- **separatrices** — the conserved properties that split the failure regime from
  the success regime; a surviving discriminative invariant is a separatrix.

A failure invariant is thus not the whole theory — it is the separatrix-shaped
reading of the geometry. Other readings (a void to expand into, a bottleneck to
route around) are equally legitimate boundaries to target.

#### Challenge is itself recursive cartography

Reconstructing a boundary is not a single measurement. Challenging a candidate
shape claim produces a **`boundary_delta`** — the minimal condition separating a
counterexample (or split axis) from the support population — and a `weaken` or
`split` with a *sharp* delta is a **resolution gain**, not a failure. The delta
opens a smaller loop nested inside this one:

```text
failure-space F0
  -> compress    -> candidate I0
  -> challenge   -> disposition ∈ {survive, weaken, split, falsify} + boundary_delta
       (if weaken/split)
  -> micro failure-space F1  = the boundary_delta's support population
  -> compress    -> candidate I1
  -> challenge   -> ...
```

So `geometry → boundary` is not one pass; it recursively sharpens the map's
resolution before the search ever commits to crossing. The same anti-fractal
discipline applies (a delta licenses at most one grounded child, entering
`proposed` with no inherited authority) — see the
[epistemic model](02-epistemic-model.md). A concrete instance is recorded in
[Finding 001](../findings/001-unencoded-shape-generation.md): the N2a candidate
survived challenge, and its most productive probe was not a refutation but a
`boundary_delta` that split one "structural ceiling" into two distinct
sub-mechanisms.

### boundary → structural delta

Choose a boundary and propose the `StructuralDelta` that crosses it: which
conserved property to break, and the argument that the break is **structural, not
cosmetic**. A proposal must be mechanistically distant from the known failure
families and state its nearest family and why it is still distinct. The
conceptual objective is:

```text
maximize(
    mechanistic_distance_from_known_failures
  + violation_of_surviving_failure_invariants
  + expected_information_gain
  - evaluation_cost
  - redundancy
)
```

This need not be a literal scalar. Ordinal, component-wise comparison is correct
for v0; a fabricated composite score would be false precision.

Only **surviving** (and operator-attested) invariants may be targeted — the
[epistemic model](02-epistemic-model.md) gates what the search is allowed to
believe about the boundary before spending effort crossing it.

### structural delta → classical projection

Compile the shape-space move into a candidate in the problem's classical
representation — the object a verifier can actually decide. See
[04-classical-projection.md](04-classical-projection.md). Crucially, the claimed
violation is **verified against the candidate's own shape**: a proposal that
*claims* to break `P` but whose signature still satisfies `P` is recorded
honestly as not-violated. Claiming a crossing is not making one.

### classical projection → verification → landing point

Route the projection through the strongest-decisive verifier and record the
verdict with its verification strength. The verdict vocabulary is
`failure | partial_failure | partial_success | success | unknown |
verification_blocked`. The landing point is the measurement.

### landing point → geometry update

This is what makes the loop recursive rather than a pipeline:

- A **failure / partial-failure** landing point re-enters the atlas as a new
  observed shape. The next `geometry` pass sees it, and the failure regime grows
  more mechanistically complete. A failed crossing has *measured* the boundary.
- A **partial-success / success** landing point feeds success compression: pair
  the broken `P` with the condition `C` under which breaking it progressed, and
  mutate the persisted, versioned **search policy** to prefer that structure,
  avoid surviving failure invariants, expand under-sampled voids, and penalize
  redundancy.

The updated geometry and the updated policy change *where the next step
searches*. Results change future search behavior as explicit persisted state —
never "the model will remember it from context."

## Why failed crossings are informative

Restating the [geometry](00-geometry-of-work.md) claim operationally:

- A candidate that **crossed** its boundary and then failed downstream
  establishes that the transition was traversable and exposes the *next*
  boundary. The separatrix it broke was real but not sufficient; there is more
  structure between here and success.
- A candidate that **failed to cross** the boundary it claimed shows the
  geometry was mis-read at that point — the reconstruction, not just the
  candidate, needs revision.

Both are gains. Neither is discarded. This is the operational meaning of
*failure is a first-class artifact*: the loop is designed so that a negative
result still advances the map.

## Failure invariants as one instrument

It is tempting to equate `newf` with "mine failure invariants." The operational
theory is broader: the failure invariant is the **separatrix instrument**. The
same geometry supports:

- expanding into a **void** (an under-sampled mechanism family) even when no
  invariant is surviving;
- routing around a **bottleneck** by targeting a boundary many failures share;
- following a **gradient** or **bridge** from a partial success back toward the
  failure regime to find the missing condition `C`.

A run can legitimately make progress by reading any of these, and the search
policy records which readings paid off so later runs weight them.

## Stopping conditions

A run may legitimately stop with:

```text
solved
invariant_established
frontier_exhausted
no_information_gain
budget_exhausted
insufficient_failure_diversity
verification_blocked
```

`no_information_gain` specifically forbids regenerating the same mechanism with
new adjectives; if the geometry stops changing, the honest move is to stop, not
to redecorate.

## Relationship to abstraction

Reconstructing geometry is itself an abstraction, and abstraction can silently
change the problem. Every geometry-level move that raises or lowers abstraction
must ground: candidate invariants are challenged **above and below** the level at
which they were inferred, and an abstraction that erases the outcome-separating
axis is rejected. See [abstraction safety](../abstraction-safety.md). Search that
climbs to a nicer vocabulary without a grounding round-trip is drift, not
progress.
