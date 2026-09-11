# Fitness Landscapes

| Field | Value |
|---|---|
| **Canonical citation** | Wright, S. (1932). "The roles of mutation, inbreeding, crossbreeding and selection in evolution." *Proceedings of the Sixth International Congress of Genetics*, 356–366. Modern treatments: Kauffman, S. (1993). *The Origins of Order* (NK landscapes); Stadler, P. F. (2002). "Fitness landscapes." |
| **Bib key** | `TODO: add wright1932roles` to `paper/references.bib` |
| **Field** | Evolutionary biology / optimization theory |
| **GoW role** | The classical baseline GoW claims to generalize: `candidate → fitness` becomes `shape of attempt → outcome` |
| **Epistemic status** | `GeneratedInterpretation` — written from model knowledge, **not** verified against the primary source. Verify before citation. |

## One-sentence core

Search and evolution can be pictured as movement on a surface where position
is a candidate (genotype, solution, configuration), height is fitness, and
the surface's topography — peaks, valleys, ruggedness, neutrality — governs
what search dynamics can achieve.

## The mechanism (typed structure)

```text
landscape = (X, N, f)
  X — configuration space (genotypes / solutions)
  N — neighborhood or move operator (what counts as "adjacent")
  f — fitness function  f: X → ℝ
dynamics: populations climb gradients, drift across neutral networks,
          get trapped on local optima, cross valleys rarely
```

## Key load-bearing ideas

- **Topography predicts search behavior** — ruggedness (local-optima
  density), epistasis (interaction between components; tunable in Kauffman's
  NK model), neutrality, and deceptiveness determine whether hill-climbing,
  recombination, or drift can succeed.
- **The landscape is operator-relative** — change the move operator and the
  same `f` yields a different topography; "distance" is a property of the
  search process, not the problem alone.
- **Local optima and valley-crossing** — the canonical explanation for why
  greedy improvement stalls and why escaping requires structured
  perturbation.
- **A metaphor with a formal core and known abuses** — the picture is
  seductive; the literature itself warns that low-dimensional intuitions
  (2D surfaces) mislead badly in high-dimensional combinatorial spaces.

## What it assumes is given in advance

- The configuration space and its encoding.
- The neighborhood/move operator.
- A scalar (or at least comparable) fitness for every point, cheap to query.

## What it produces

A predictive vocabulary connecting problem structure to search-algorithm
performance — the base layer under metaheuristics, ISA, and QD.

## Mapping to GoW vocabulary

| Fitness landscapes | GoW / `newf` term |
|---|---|
| Configuration x | a candidate in domain-space (not yet Work) |
| Fitness f(x) | outcome — but GoW refuses the scalar; outcomes are typed regimes |
| Neighborhood N | shape adjacency (implicitly: small structural deltas) |
| Local optimum | a regime that absorbs continued linear search |
| Valley crossing | crossing a `Boundary` via a `StructuralDelta` |

## What GoW borrows

The founding move — take a population of evaluated points and treat their
arrangement as terrain that should steer search — and the warning that comes
with it: topography claims are operator-relative and easy to over-read.

## Where GoW departs

Three refusals: (1) points are *attempts*, not candidates; (2) outcome is a
typed regime with verification strength, not scalar fitness; (3) neither the
encoding nor the neighborhood is given — both are inferred and revisable.
Each refusal buys expressiveness and costs rigor.

## Reduction test (how this tradition attacks GoW)

> A landscape needs (X, N, f) to be defined. GoW has no explicit N (what is
> adjacent to a proof attempt?) and a non-scalar f. If GoW borrows landscape
> vocabulary — peaks, voids, trajectories — without any (X, N, f) analog, it
> inherits the metaphor's seductiveness and none of its predictive content:
> the known abuse case of this literature.

Defense: state the GoW analog explicitly — X = canonical signatures, N =
declared structural-delta operators, f = outcome regime + verification
strength — and confine topographic language to what that triple supports.
