# Quality-Diversity / MAP-Elites / Novelty Search

| Field | Value |
|---|---|
| **Canonical citation** | Mouret, J.-B. & Clune, J. (2015). "Illuminating Search Spaces by Mapping Elites." arXiv:1504.04909. Lehman, J. & Stanley, K. O. (2011). "Abandoning Objectives: Evolution Through the Search for Novelty Alone." *Evolutionary Computation* 19(2). Survey: Pugh, Soros & Stanley (2016), "Quality Diversity: A New Frontier for Evolutionary Computation," *Frontiers in Robotics and AI*. Learned descriptors: Cully (2019), AURORA. |
| **Bib keys** | `mouret2015illuminating`, `lehman2011abandoning` (present in `paper/references.bib`) |
| **Field** | Evolutionary computation |
| **GoW role** | License for preserving multiple regimes instead of scalarizing to one optimum |
| **Epistemic status** | `GeneratedInterpretation` — written from model knowledge, **not** verified against the primary source. Verify before citation. |

## One-sentence core

Instead of driving search toward one scalar optimum, *illuminate* a behavior
space: keep the best solution found in every behavioral niche, because
diversity of high performers is both the goal and the engine of progress.

## The mechanism (typed loop — MAP-Elites)

```text
behavior descriptor  b: solution → ℝ^d      (chosen in advance)
archive              grid over descriptor space, one elite per cell
loop:
    parent  ← random elite from archive
    child   ← mutate(parent)
    (b, f)  ← evaluate(child)               (descriptor + fitness)
    if cell(b) empty or f > fitness(cell(b)): archive[cell(b)] ← child
output: the full archive (a map of elites), not a single best
```

Novelty search is the purified precursor: select for *novelty of behavior
alone*, ignoring the objective — and it famously solves deceptive tasks that
objective-driven search fails.

## Key load-bearing ideas

- **Objectives can be deceptive** — the gradient of the objective may point
  away from the stepping stones that lead to it; rewarding novelty finds the
  stepping stones anyway (Lehman & Stanley's central result).
- **Illumination, not optimization** — the deliverable is the *map*: how good
  can you be at every point in behavior space? Peaks in different niches are
  different answers, all kept.
- **Elites in odd niches are stepping stones** — a mediocre-objective,
  novel-behavior solution is retained because its descendants may reach
  regions greedy search never touches.
- **Descriptor choice is the whole game** — the behavior characterization
  determines what diversity means; AURORA and successors *learn* descriptors
  (e.g. by autoencoder on trajectories) instead of hand-picking them.

## What it assumes is given in advance

- A behavior descriptor (or, in later work, a learning procedure for one).
- Cheap evaluation — archives are filled by millions of evaluations.
- A mutation operator over solutions.

## What it produces

An archive of diverse, locally-optimal solutions; a picture of the
performance potential of the entire behavior space.

## Mapping to GoW vocabulary

| QD | GoW / `newf` term |
|---|---|
| Behavior descriptor | `StructuralShape` (roughly — but see departure) |
| Archive cell / niche | `Regime` (roughly) |
| Elite | best-known Work exhibiting a shape |
| Illumination | the atlas: shape-space with outcome overlay |
| Descriptor learning (AURORA) | `M(W)` representation movement |

## What GoW borrows

The instinct that the space contains **multiple useful regimes, bridges, and
voids rather than one hill**, and that preserving structurally distinct
attempts — including failures — is more valuable than ranking everything on
one scale. "Mechanistic novelty over surface novelty" is a QD-flavored rule.

## Where GoW departs

QD describes *behavior* (what the solution does); GoW describes *mechanism*
(how the attempt was constructed: operators, assumptions, preserved
properties). QD needs millions of cheap evaluations; GoW works at n≈dozens
with expensive, graded verification. And QD has no epistemic lifecycle — an
archive entry is a measurement, never a claim under challenge.

## Reduction test (how this tradition attacks GoW)

> AURORA-style QD already learns its characterization from the data, so
> "GoW infers the descriptor too" is not the differentiator. If shape axes
> are just learned behavior descriptors and regimes are just niches, GoW is
> low-sample QD with an LLM mutation operator — and low-sample QD is known
> to be weak, because illumination needs density.

Defense: the differentiator must be mechanism-vs-behavior plus the epistemic
layer — show that a mechanism-described, challenge-disciplined map licenses
moves (e.g. complement-geometry probes, invariant-violating proposals) that a
behavior archive cannot represent, let alone justify.
