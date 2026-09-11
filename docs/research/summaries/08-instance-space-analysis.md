# Instance Space Analysis (ISA)

| Field | Value |
|---|---|
| **Canonical citation** | Smith-Miles, K. et al. — e.g. Smith-Miles, K. & Lopes, L. (2012). "Measuring instance difficulty for combinatorial optimization problems." *Computers & Operations Research*; Smith-Miles, K. & Muñoz, M. A. (2023). "Instance Space Analysis for Algorithm Testing: Methodology and Software Tools." *ACM Computing Surveys*. Platform: MATILDA (University of Melbourne). |
| **Bib key** | `TODO: add smithmiles2023isa` to `paper/references.bib` |
| **Field** | Operations research / algorithm selection & testing |
| **GoW role** | Structurally the nearest relative to geometry-of-performance; the "lift" is defined against it |
| **Epistemic status** | `GeneratedInterpretation` — written from model knowledge, **not** verified against the primary source. Verify before citation. |

## One-sentence core

Map a benchmark's problem *instances* into a low-dimensional feature space,
overlay each algorithm's performance, and read off where algorithms succeed,
where they fail, where the benchmark has holes, and which features predict
difficulty.

## The mechanism (typed pipeline)

```text
instances I
 → feature extraction        f: I → ℝ^k     (difficulty-correlated features)
 → feature selection          (subset that best separates performance)
 → projection to 2D           (e.g. PILOT: performance-aware linear projection)
 → performance overlay        per-algorithm good/bad regions ("footprints")
 → footprint analysis         area, purity, density of each algorithm's region
 → gap detection              unsampled/underrepresented regions of the space
 → instance generation        synthesize new instances to fill gaps; iterate
```

Rooted in Rice's (1976) algorithm selection framework: features → space →
per-algorithm performance mapping.

## Key load-bearing ideas

- **Performance is regional, not global** — "algorithm A beats B" is almost
  always false as stated and true only over a region; the footprint is the
  honest unit of comparison.
- **Benchmarks have geometry and bias** — standard test suites cluster in
  small pockets of the instance space; conclusions drawn from them silently
  inherit that sampling bias.
- **The map drives generation** — when the space has holes, ISA *generates
  new instances* targeted at the holes. The loop map → gap → generate →
  re-map is sequential design over a feature space.
- **Features are selected for outcome-discrimination** — a feature earns its
  axis by separating good from bad performance, not by being descriptive.

## What it assumes is given in advance

- A candidate feature library for instances (human-engineered).
- Algorithms whose performance on an instance is cheaply measurable.
- Instances as the unit: each point is a *problem*, not an *attempt*.

## What it produces

A 2D atlas of the problem class: per-algorithm footprints, difficulty
gradients, benchmark-bias diagnosis, and targeted new instances.

## Mapping to GoW vocabulary

| ISA | GoW / `newf` term |
|---|---|
| Instance features | `StructuralShape` axes |
| Instance space | shape-space (the atlas) |
| Performance footprint | `Regime` |
| Footprint boundary | `Boundary` between regimes |
| Gap in sampled space | `AntiVacuum` / void |
| Targeted instance generation | `Projection` of a probe |

## What GoW borrows

Nearly the whole geometric reading: outcome-conditioned regions over a
feature-described population, boundary and hole analysis, and the discipline
of generating new points where the map is uninformative.

## Where GoW departs

**What counts as a point.** ISA maps *problem instances*; classical fitness
landscapes map *candidate solutions*; GoW maps *attempts* — proofs, methods,
experiments, arguments — described by mechanism structure. Consequences:

- ISA's features are engineered up front from a mature library; GoW's shape
  axes must be inferred from the Work itself (`M(W)` may move).
- ISA's outcome measure is cheap and repeatable (run the solver); GoW's
  landing points are expensive, graded in verification strength, and rare.
- ISA's population is large and homogeneous; GoW's is small and
  heterogeneous.

## Reduction test (how this tradition attacks GoW)

> Call each attempt an "instance," call its mechanism fields "features," and
> GoW *is* ISA with n≈20 points, unvalidated features, and no projection
> method — i.e., ISA done badly. The change of point-type must be shown to
> change the method, not just the noun.

Defense: identify operations GoW performs that ISA cannot express — the
epistemic lifecycle over claims about the space, challenge-driven axis
splitting, and complement-geometry reasoning over *excluded* work — and show
at least one produces a move ISA's pipeline would not.
