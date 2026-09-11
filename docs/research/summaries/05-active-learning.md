# Active Learning

| Field | Value |
|---|---|
| **Canonical citation** | Settles, B. (2009). *Active Learning Literature Survey*. CS Technical Report 1648, University of Wisconsin–Madison. Foundational: Angluin (1988) "Queries and Concept Learning"; Seung, Opper, Sompolinsky (1992) "Query by Committee"; Lewis & Gale (1994) "A Sequential Algorithm for Training Text Classifiers." |
| **Bib key** | `TODO: add settles2009active` to `paper/references.bib` |
| **Field** | Machine learning |
| **GoW role** | Mature ancestor of the **Probe** reading — choose the next Work for information, not for success |
| **Epistemic status** | `GeneratedInterpretation` — written from model knowledge, **not** verified against the primary source. Verify before citation. |

## One-sentence core

A learner that chooses which examples to label next can reach the same
accuracy with far fewer labels than one fed random examples, by querying where
its current knowledge is weakest.

## The mechanism (typed loop)

```text
model      ← train(labeled pool L)
query x*   ← argmax_x  utility(x | model, unlabeled pool U)
label y*   ← oracle(x*)
L ← L ∪ {(x*, y*)}; loop until budget or convergence
```

Principal utility families:

- **Uncertainty sampling** — query where the model is least confident
  (least-confident / margin / entropy).
- **Query-by-committee (QBC)** — train a committee of consistent hypotheses;
  query where they *disagree most*. This is the version-space connection:
  disagreement marks the surviving region's interior boundary.
- **Expected model change / expected error reduction** — query what would
  most alter or improve the model.
- **Density weighting** — discount informative-but-unrepresentative outliers.

## Key load-bearing ideas

- **Labels are expensive; choose them adversarially against your own
  ignorance.** The query is an experiment on the hypothesis space.
- **Disagreement localizes information** — where surviving hypotheses agree,
  observation is redundant; where they disagree, one observation splits
  the space (QBC can achieve exponential label-complexity savings in
  favorable settings).
- **Scenarios** — membership query synthesis (the learner *constructs* the
  query), stream-based selection, and pool-based sampling. Query synthesis
  is the closest to GoW: the probe is generated, not selected.
- **Pathologies** — synthesized queries can be uninterpretable to the oracle;
  sampling bias in the labeled set can mislead later training.

## What it assumes is given in advance

- A fixed model/hypothesis class and feature representation.
- A reliable oracle whose per-query cost is roughly uniform.
- A utility function computable *before* the label is known.

## What it produces

A sequence of queries approximately maximizing information about the target
concept per unit labeling cost.

## Mapping to GoW vocabulary

| Active learning | GoW / `newf` term |
|---|---|
| Query x* | next `Work` (Probe mode) |
| Oracle label | `LandingPoint` (`Evaluation`) |
| Committee disagreement region | boundary/void where the geometry is least determined |
| Query synthesis | `Projection` of an informative structural delta |
| Label budget | evaluation cost in the frontier score |

## What GoW borrows

`next Work ≈ argmax_W information gained about the geometry` — the frontier
discipline's `expected_information_gain − evaluation_cost` term is
active-learning utility, ordinal rather than probabilistic.

## Where GoW departs

Active learning queries points in a **known feature space** against a fixed
hypothesis class, with a cheap, reliable oracle. GoW's probes are whole pieces
of Work in a space whose coordinates are inferred; its oracle is a costly,
tiered verification hierarchy; and each probe can change the representation in
which utility was computed — a moving-target problem active learning does not
face.

## Reduction test (how this tradition attacks GoW)

> If the shape vocabulary is frozen, "probe the boundary" is pool-based QBC
> with an LLM committee — established method, weaker oracle. And if utility
> cannot be computed even ordinally before evaluation, "maximally informative
> next Work" is a slogan, not a selection rule.

Defense requires showing the selection rule is *executable* (predeclared,
comparable, auditable) and that it beats undirected generation on information
per evaluation — falsification condition 2 territory.
