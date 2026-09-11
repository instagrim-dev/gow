# Theory series

This directory holds the **conceptual model** of `newf` — what the objects mean
and what we are allowed to believe about them — deliberately separated from
implementation docs, experiment records, and the delivery spine.

The separation is the point:

- **Theory** (this directory) — what the objects mean.
- **[Epistemics](02-epistemic-model.md)** (a layer within theory) — what we are
  allowed to believe about them.
- **[Findings](../findings/)** — what experiments have actually shown.
- **[Thesis](../thesis/why-solve-by-shape.md)** — why any sane person should care.

The theory is derived rigorously from what the implementation has taught us; the
thesis is the authored, project-level statement of the move itself; the
practitioner field guide is the reach-for-it-Monday-morning companion to both.

## Layers

| # | Document | Layer | Question it answers |
|---|---|---|---|
| 00 | [Geometry of Work](00-geometry-of-work.md) | Semantic (center) | What space does `newf` search, and what is its governing principle? |
| 01 | [Semantic Model](01-semantic-model.md) | Semantic | What are the descriptive objects and how do they relate? |
| 02 | [Epistemic Model](02-epistemic-model.md) | Epistemic | What may we believe, and how does belief strengthen? |
| 03 | [Shape-Guided Search](03-shape-guided-search.md) | Operational | What is the recursive algorithm over the geometry? |
| 04 | [Classical Projection](04-classical-projection.md) | Operational | How does a shape-space move get verified in domain-space? |
| 05 | [Complement Geometry](05-complement-geometry.md) | Semantic (dual) | What shape does accumulated Work force the remaining possibility-space to take? |
| 06 | [Situated in the Literature](06-situated-in-the-literature.md) | Research program | What established fields does GoW compose, what is the distinct lift, and what rigorous pieces are still missing? |
| — | [Paper Claims](00-paper-claims.md) | Claim registry | Which four claims may the manuscript make, and at what evidence standard? |
| — | [Glossary / ontology](glossary.md) | Anchor | What does each term mean, and its legacy implementation name? |

Alongside the theory series, the [thesis](../thesis/) directory holds the
authored position pieces:

| Document | Layer | Question it answers |
|---|---|---|
| [Why Map the Work?](../thesis/why-map-the-work.md) | Introduction | What does examining prior attempts teach that another attempt would not? |
| [Why Solve by Shape?](../thesis/why-solve-by-shape.md) | Thesis | Why is any of this worth building? |
| [When to Map the Work](../thesis/when-to-map-the-work.md) | Field guide | When should a practitioner reach for GoW, and how? |
| [How to Map the Work](../thesis/how-to-map-the-work.md) | Method | How to record attempts so the move is repeatable? |

## The governing principle

> **Search in shape-space; verify in domain-space.**

Everything else in the series elaborates, constrains, or operationalizes that
one sentence.

## Reading order

Read [00-geometry-of-work.md](00-geometry-of-work.md) first; it is the
conceptual center. Then the semantic and epistemic models (01, 02) for the
vocabulary and belief rules, then the two operational documents (03, 04) for the
algorithm. Keep the [glossary](glossary.md) open alongside — it maps every
conceptual term to the legacy implementation term so you never have to guess
whether `ShapeClaim` and `CandidateInvariant` are the same thing (they are).

## Relationship to implementation docs

Theory documents use **conceptual** terms. Implementation docs
(`docs/*.md` outside this directory), the [domain model](../domain-model.md),
code, and schema keep their **legacy** terms. The
[glossary](glossary.md) is the bridge, and its naming policy is explicit:
**document the cleaner model first; do not rename the database as a side effect.**
