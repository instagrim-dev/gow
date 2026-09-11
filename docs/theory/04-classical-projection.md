# 04 — Classical Projection

*Layer 3 (operational) of the [theory series](./), companion to
[shape-guided search](03-shape-guided-search.md). This document governs the
single most delicate crossing in the system: taking a move made in shape-space
and compiling it back into the problem's classical representation so it can be
**verified in domain-space**. This is the concrete meaning of the second half of
the governing principle, "search in shape-space; verify in domain-space."*

---

## The crossing

Shape-guided search produces a `StructuralDelta`: a claim of the form "break
conserved property `P`; the break is structural because …; the nearest known
family is `F` but this is distinct because …". That object lives entirely in
shape-space. It cannot be verified there — shape-space has no notion of whether
the Erdős–Straus equation actually has a solution for a given `n`.

**Projection** is the compilation of the delta into a candidate expressed in the
problem's own classical representation: the object a deterministic check,
computation, proof obligation, or independent source can decide. The candidate's
observed result is the **LandingPoint**.

```text
StructuralDelta (shape-space)
   -> Projection (candidate in classical/domain representation)
   -> Verification (strongest decisive verifier)
   -> LandingPoint (verdict + verification strength)
   -> re-enters shape-space as observed work
```

## Why projection is separate from the delta

A move in shape-space and its domain-space image are different objects with
different truth conditions, and conflating them is the central failure mode the
[epistemic model](02-epistemic-model.md) guards against.

- The delta is a **hypothesis about structure**. It is model-authored and lives
  at model-judgment strength.
- The projection's landing point is a **measurement in the domain**. Its strength
  is whatever verifier decided it.

Keeping them separate is what stops a fluent structural argument from being
counted as a domain result. The delta says "this should cross"; only the landing
point says "it did / it didn't," and only at the strength of the mechanism that
checked.

## The internal-consistency check comes first

Before a projection is routed to a domain verifier, one thing is checkable
purely in shape-space: **does the candidate's own shape actually violate the
property it claims to violate?**

The candidate carries its own `StructuralShape`. Evaluate the target claim's
predicate against that shape:

- `violates` → the candidate really is a crossing of that boundary.
- `satisfies` → the candidate *claims* a break but does not embody one; record
  it honestly as not-violated. `ModelJudgment ≠ Verification`.
- `unknown` → ambiguous; keep it `unknown`, never coerce it to a violation.

This is the cheapest tier and it catches the most common defect: a proposal whose
prose asserts a structural move its own structure doesn't make. Only candidates
that pass internal consistency are worth spending a domain verifier on.

## The verifier hierarchy in projection

Domain-space verification routes **strongest-decisive-first**, not
cheapest-confident-first (though within a strength tier, cheaper is preferred):

```text
formal proof / deterministic check
  > reproducible computation or experiment
  > independently sourced evidence
  > cross-model / independent-critic agreement
  > single-model judgment
```

The ordering matters because it prevents strength laundering: a deterministic
check that returns `failure` can never be overridden by a later, more confident
model `success`. Every landing point is stamped with both the verifier used and
its verification strength; the verdict and its strength are one inseparable
record.

Concrete tiers in the current system: a deterministic-check tier that reuses the
predicate evaluator over the per-target violation verdicts; a
counterexample-search tier that scans the nearest known failure families for a
refuter; and a last-resort model-judgment tier. A model tier's confident
`success` is still the weakest evidence in the hierarchy.

## Projection as abstraction in reverse

Search abstracts *upward*: concrete work → shape-space. Projection grounds *back
downward*: a shape-space move → a concrete domain candidate. This is precisely
the `abstract` / `ground` pair required by
[abstraction safety](../abstraction-safety.md):

```text
abstract   Concrete work        -> shape-space geometry
ground     shape-space delta     -> concrete domain candidate -> test
```

An unprojectable delta is the shape-space analogue of an ungroundable
abstraction. If a structural move *cannot* be compiled into any candidate a
verifier can decide, it earns no search authority — however elegant it reads.
`cannot_be_operationalized` is a valid, recorded outcome; the response is not to
generate a more impressive noun.

## What a projection must carry

For provenance and replay, a projection records enough to answer later:

- which `StructuralDelta` (and which target claim) it compiles;
- the candidate's own `StructuralShape` and the internal-consistency verdict;
- which verifier tier decided it and at what strength;
- the resulting verdict;
- the re-entry marker if the landing point was a failure/partial-failure, so the
  mechanism rejoins the atlas on the next geometry pass.

Nothing in the chain should require "trust the transcript" as provenance.

## Failure at the projection boundary

Three distinct things can go wrong, and they are not the same:

1. **The delta doesn't embody its claimed break** (internal-consistency
   `satisfies`). The shape-space reasoning was inconsistent with its own
   candidate.
2. **The projection embodies the break but fails in the domain** (verifier
   returns failure/partial-failure). The crossing was real but insufficient —
   this is the informative "measurement of the geometry" that exposes the next
   boundary.
3. **The delta cannot be projected at all** (`verification_blocked` /
   `cannot_be_operationalized`). The move has no domain image to test.

Only (2) advances the map by crossing. (1) corrects the shape-space
reconstruction. (3) says the move never left shape-space. Recording which of the
three occurred is what makes the landing point a usable measurement rather than
an undifferentiated "it failed."

## One-line summary

> A structural move earns nothing until it is compiled into something the domain
> can decide — and then it earns exactly the strength of the mechanism that
> decided it.
