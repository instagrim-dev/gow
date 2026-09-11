# When to Map the Work

*A practitioner's field guide to Geometry of Work. The
[thesis](why-solve-by-shape.md) says why the approach is worth building; the
[theory series](../theory/) says what the objects mean; this document says when
a practitioner should reach for it and what to do next.*

---

## The instrument, not the solver

GoW is not a solver. It is an **instrument for a specific cognitive situation**:

> **"I have a lot of Work. Continued Work is not producing information."**

That is when to reach for it.

A calculator answers:

```
known operation + inputs → result
```

GoW is closer to a microscope, map, or diagnostic instrument:

```
history of Work → structure of search → next informative move
```

You don't reach for it because a problem is hard.
You reach for it because **the attempts themselves have become evidence**.

The analogous practitioner instincts:

```
System behaving strangely
    → reach for observability

Program behaving strangely
    → reach for debugger

Data behaving strangely
    → reach for statistics

Search behaving strangely
    → reach for Geometry of Work
```

---

## The trigger

A discipline needs a recognizable trigger before it needs terminology.

The GoW trigger is:

> **When another attempt is cheaper intellectually than understanding why the
> previous attempts cluster, stop generating and map the Work.**

Or, more operationally:

```
I have multiple serious attempts.
↓
They repeatedly stall, partially work, or disagree.
↓
I cannot explain the structural difference between those outcomes.
↓
Map the Work.
```

This is a recognizable practitioner situation. The prescription is specific:
**stop generating, and map**.

---

## The primitive is `map`, not `solve`

A practitioner should first think: **"Map the Work."**

The workflow then becomes cognitively small:

1. **Collect** — gather meaningful attempts with their outcomes.
2. **Shape** — describe each attempt structurally (representation, operators,
   assumptions, preserved properties), not just narratively.
3. **Locate** — assign each attempt to an outcome regime (failure,
   partial, success).
4. **Compare** — find what structurally differs between neighboring regimes.
5. **Find** — identify conserved structure within regimes and boundaries
   between them.
6. **Move** — generate Work that crosses or probes a promising boundary.
7. **Project** — translate that structural move back into the native
   domain representation.
8. **Observe** — record where the new Work lands, and return to step 1.

The loop is:

```
Map → Boundary → Move → Landing → Map'
```

GoW is not only analysis-before-action. **Action is how you probe the geometry.**
Each landing point updates the map.

---

## The fundamental artifact: `WorkMap`

GoW should return something a practitioner can ask for by name.

> "We're stuck. Build a WorkMap."

A WorkMap minimally contains:

```
Problem / objective

Work population
    the attempts and approaches on record

Shape dimensions
    what structurally varies across the Work

Outcome regimes
    failure / partial / success / unknown — where each attempt sits

Conserved shapes
    structure that persists within a regime (candidate regularity)

Boundaries
    structural conditions that separate regimes

Voids
    structurally plausible combinations not yet represented

Frontiers
    boundaries worth attempting to cross

Next probes
    Work that would maximally clarify the map
```

Someone doesn't need to understand `candidate_invariant`,
`ComparisonProfile`, or any migration number to ask for a WorkMap. That
reachability matters.

---

## Two modes: navigate and probe

Once you have a map, two fundamentally different questions become available.

### Navigate

> "We think we know where better outcomes are. How do we get there?"

```
current regime
→ boundary
→ structural delta
→ classical candidate
```

This is constructive. You have identified a structural change worth making and
you are generating Work that enacts it.

### Probe

> "We don't understand the geometry well enough yet. What Work would teach us
> the most?"

```
uncertain boundary
→ discriminating experiment
→ landing point
→ updated map
```

This is epistemic. You are not yet confident which map is true, and the best
next Work is not the Work most likely to solve the problem — it is the Work
most likely to **tell you which map is true**.

This second mode is experimental science hiding inside problem solving.
Distinguishing navigate from probe is one of GoW's most practical outputs.

---

## The mnemonic

```
Stuck → Map → Boundary → Move → Measure
```

**Stuck** — repeated Work is no longer informative.
**Map** — represent attempts by structural shape and outcome.
**Boundary** — find what separates regimes.
**Move** — generate Work that crosses or tests the boundary.
**Measure** — observe the landing point and update the map.

Each step is doable without software. Each step has an instrumented form in
`newf`. The mnemonic is the same at both levels.

---

## Three levels of practice

GoW does not require `newf` to be useful.

### Level 1 — GoW as a mental move (whiteboard)

You have six failed designs and two partial successes. Ask:

- What is structurally common among the failures?
- What changed in the partial successes?
- Which differences are incidental, and which separate outcome regimes?
- What is the smallest move across that boundary?

That is already GoW. No software. No ontology. No state machine.

### Level 2 — GoW as a method (semi-explicit)

Introduce explicit artifacts:

```
Work item    Shape    Outcome    Regime
Shape claim    Boundary    Structural delta    Projection    Landing point
```

Maintain these manually or semi-automatically. This is where GoW becomes
comparable to:

- root-cause analysis,
- threat modeling,
- experimental design,
- systems modeling,
- design-space exploration.

A repeatable discipline with reviewable outputs.

### Level 3 — GoW as an instrumented system (`newf`)

The machine handles:

```
large Work corpus
→ normalization
→ shape extraction
→ redundancy collapse
→ candidate regularities
→ challenge
→ boundary inference
→ frontier generation
→ replay
```

`newf` is **an implementation of the discipline**, not the discipline itself.
That distinction prevents conceptual grief. The discipline can be taught,
applied, and evaluated at Level 1 and Level 2 before Level 3 exists.

---

## The core question, reduced

If Geometry of Work has a one-sentence summary analogous to "calculus studies
change," it is:

> **Study the relationship between the shape of Work and the outcomes of Work.**

Formally:

```
G : S(W) → O
```

where `W` is Work, `S(W)` is its structural shape, and `O` is the observed
outcome.

The practical objective is:

```
Find ΔS such that O(S + ΔS) > O(S)
```

Find the smallest meaningful structural change that moves Work into a better
outcome regime. The rest of the theory — clustering, shape claims, challenge,
frontier generation, projection — elaborates how to do that reliably.

---

## The operation vocabulary

The "buttons" of GoW:

```
map(work)             → WorkMap

compress(region)      → candidate conserved shapes within that regime

contrast(A, B)        → structural differences between regimes A and B

boundary(F, S)        → the structural condition separating failure regime F
                         from success regime S

void(work_map)        → structurally plausible combinations not in the Work

challenge(shape)      → probe whether the shape survives attempted falsification

probe(boundary)       → generate Work that tests a specific boundary

navigate(current,     → generate Work that moves from current to target
         target)        across identified boundaries

project(delta,        → translate a structural delta into the problem's
        domain)         classical representation

update(landing)       → incorporate a landing point into the map
```

Many of these operations are epistemic and abductive rather than deterministic.
Their output is therefore **claims + provenance + uncertainty** rather than
values. The existing challenge discipline already embodies that: candidate
structure remains hypothesis until attacked, and challenge can survive, weaken,
split, or falsify it.

---

## What Pilot-004 showed

Pilot-004 is the first experimental evidence that **shape generation** — the
`compress` operation, extracting candidate conserved shapes from a failure
corpus — may be model-native enough to automate. Four formulations of the same
structural regularity (N2a: density/averaging ceiling) emerged independently
from multiple blind runs and survived quorum criticism. That candidate was
then run through a recorded seven-probe challenge campaign and survived at
model-judgment strength with one required refinement (a rate-ceiling vs
reach-ceiling boundary delta), and was admitted as a
`GeneratedInterpretation` claim — with two probes' evidential character
qualified in the record's attributed reassessment (no completed
counterexample search at C2; causal probe C7 remains a proposed
explanation).

That matters for the field guide because it means Level 3 is no longer
hypothetical for at least one operation. **Shape proposal occurred.** The
question it opens is not "can the model do this" but "does doing this
recursively refine the failure-space representation in a way that generates
better frontiers" — which is the next experiment.

---

## The recommended first teaching artifact

The first artifact for someone learning GoW should be:

1. This field guide: the trigger, the mnemonic, the WorkMap definition, and the
   two modes.

2. A worked case: the **same problem** approached first conventionally (generate
   more attempts, iterate on the same approach) and then geometrically (map the
   attempts already made, identify the boundary, generate the smallest
   discriminating Work). Show the structural difference in what the two
   approaches see and what each can ask.

The worked case should not start with a solved problem. It should start with a
case where conventional attempts have stalled — because that is the exact
situation the trigger identifies.

Neither the formal theory nor `newf` should be the first artifact. Both require
vocabulary that presupposes the motive. The motive is: **you are stuck, and the
existing Work contains more information than you have extracted from it.**

---

## Relationship to the rest of the series

- **[Why Map the Work?](why-map-the-work.md)** — the reader-facing
  introduction: experience the reasoning move once before any terminology.
- **[How to Map the Work](how-to-map-the-work.md)** — the reusable recording
  method: attempts, commonality vs explanation, discriminating probes, honest
  updates.
- **[Why Solve by Shape?](why-solve-by-shape.md)** — the thesis: why this
  is worth building at all.
- **[00 — Geometry of Work](../theory/00-geometry-of-work.md)** — the
  conceptual model: the objects and the governing principle.
- **[Findings 001](../findings/001-unencoded-shape-generation.md)** — the
  first experimental evidence that shape generation is model-native.
- **[Glossary](../theory/glossary.md)** — maps every term here to its legacy
  implementation name in the code.
