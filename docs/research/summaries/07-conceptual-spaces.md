# Conceptual Spaces — Gärdenfors

| Field | Value |
|---|---|
| **Canonical citation** | Gärdenfors, P. (2000). *Conceptual Spaces: The Geometry of Thought*. MIT Press. Follow-up: Gärdenfors (2014), *The Geometry of Meaning*. |
| **Bib key** | `TODO: add gardenfors2000conceptual` to `paper/references.bib` |
| **Field** | Cognitive science / philosophy of mind |
| **GoW role** | Philosophical license for "semantic structure can carry useful geometry" |
| **Epistemic status** | `GeneratedInterpretation` — written from model knowledge, **not** verified against the primary source. Verify before citation. |

## One-sentence core

Concepts are best modeled neither symbolically nor connectionistically but
*geometrically*: as regions in a space whose axes are quality dimensions, with
similarity as distance and natural concepts as convex regions.

## The mechanism (typed structure)

```text
quality dimensions  (hue, brightness, pitch, weight, temperature, …)
 → domains          (integral dimension groups, e.g. color = hue×sat×brightness)
 → conceptual space (a product of domains with a similarity metric)
 → properties       (regions within one domain)
 → concepts         (regions spanning domains, correlated)
 → prototypes       (central points; categories as Voronoi cells around them)
```

## Key load-bearing ideas

- **A third level of representation** — between symbolic logic (too rigid for
  similarity and learning) and subsymbolic vectors (too opaque for meaning),
  geometry gives similarity, vagueness, and concept formation for free.
- **Naturalness = convexity** — the criterion that a "natural" property is a
  *convex* region: if two instances have the property, everything between
  them does too. This is a substantive, falsifiable constraint on what counts
  as a well-formed concept, not decoration.
- **Betweenness and distance are the primitives** — the geometry is defined
  by which judgments (`b is between a and c`; `a is more similar to b than to
  c`) the space must respect.
- **Prototype effects fall out** — Voronoi tessellation around prototypes
  explains graded membership and category boundaries.
- **Dimensions can be learned** — quality dimensions may be innate, culturally
  acquired, or extracted from similarity data (e.g. by multidimensional
  scaling); the space is not necessarily given.

## What it assumes is given in advance

A stock of quality dimensions per domain (even if learned offline) and a
similarity structure. It is a *representational theory*, not a search
algorithm: it says what concepts are, not what to do next.

## What it produces

Philosophical and formal legitimacy for non-Euclidean, non-physical semantic
geometry — plus concrete design criteria (convexity, betweenness,
prototype structure) that any claimed semantic geometry can be tested against.

## Mapping to GoW vocabulary

| Conceptual spaces | GoW / `newf` term |
|---|---|
| Quality dimensions | shape axes (`representation`, `assumptions`, `operators`, `preserves`) |
| Domain | an axis group of `StructuralShape` |
| Convex region | `Regime` (aspirationally) |
| Prototype | a cluster's canonical signature (`MechanismSignature`) |
| Betweenness | what `CompareWithProfile`'s ordinal comparison gestures at |

## What GoW borrows

The core license behind claim C1: "geometry" in the informal sense — a
structured space with regions, boundaries, and trajectories — can be
meaningful over semantic dimensions without being a metric space.

## Where GoW departs

Gärdenfors models *concepts held by an agent*; GoW models *completed attempts
at a problem*, conditioned on outcome, with an epistemic lifecycle over claims
about the space. Conceptual spaces have no notion of regimes-by-outcome,
challenge, or search policy.

## Reduction test (how this tradition attacks GoW)

> Gärdenfors pays for the word "geometry" with axioms: betweenness,
> convexity, a similarity metric. GoW currently pays with component-wise
> ordinal comparison. If GoW can state no analog of the convexity criterion —
> no property its regimes must have to count as well-formed regions — then
> "geometry of work" is metaphor debt, and Conceptual Spaces is the creditor.

Defense: adopt a naturalness criterion for regimes (e.g., a regime claim must
be closed under some stated family of shape interpolations, or be split), and
test whether observed regimes satisfy it.
