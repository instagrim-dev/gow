# Constraint Propagation / Feasible-Region Contraction

| Field | Value |
|---|---|
| **Canonical citation** | Mackworth, A. K. (1977). "Consistency in Networks of Relations." *Artificial Intelligence* 8(1), 99–118 (AC-3). Precursor: Waltz, D. (1975), line-labeling constraint filtering. Textbook: Dechter, R. (2003). *Constraint Processing*. Morgan Kaufmann. |
| **Bib key** | `TODO: add mackworth1977consistency`, `dechter2003constraint` to `paper/references.bib` |
| **Field** | Constraint satisfaction / AI |
| **GoW role** | The nearest dual to "collapse the space around it" — the anti-vacuum as semantic constraint propagation |
| **Epistemic status** | `GeneratedInterpretation` — written from model knowledge, **not** verified against the primary source. Verify before citation. |

## One-sentence core

Before (and during) search, mechanically shrink each variable's domain by
removing values that cannot participate in any solution consistent with the
constraints — and let each removal trigger further removals until a fixpoint.

## The mechanism (typed operator)

```text
CSP = (variables X, domains D, constraints C)
feasible region:  Ω' = { x ∈ Ω : C_1(x) ∧ C_2(x) ∧ … ∧ C_n(x) }
propagation (arc consistency, AC-3):
    repeat until fixpoint:
        for constraint c over (x_i, x_j):
            remove from D_i every value with no support in D_j
outcomes: domain wipe-out (provably infeasible) | fixpoint (smaller Ω') | singleton (solved)
```

Propagation interleaves with search: each tentative assignment triggers
further contraction; failure (wipe-out) triggers backtracking with the
learned infeasibility.

## Key load-bearing ideas

- **Deduction before search** — every value removed by propagation is a
  region of the search space that never has to be visited; inference and
  search trade off.
- **Local consistency ≠ global solvability** — arc consistency prunes
  soundly but incompletely; a fixpoint with nonempty domains does not imply a
  solution exists. Stronger consistencies (path, k-consistency) prune more
  at higher cost.
- **Propagation cascades** — one removal enables others; the fixpoint is a
  global consequence of local rules.
- **The residual domain is the state of knowledge** — what search knows is
  exactly the current contracted domains; the shape of what *remains* directs
  where to branch next (e.g. fail-first heuristics branch on the smallest
  domain).

## What it assumes is given in advance

- Explicit variables with enumerable (or interval) domains.
- Constraints with checkable, native semantics (`x > 4`, `alldifferent`).
- Soundness: a removed value is *provably* in no solution.

## What it produces

A contracted feasible region, infeasibility proofs (wipe-outs), and branching
guidance from the shape of the residual domains.

## Mapping to GoW vocabulary

| Constraint propagation | GoW / `newf` term |
|---|---|
| Ω and domains | `PossibilitySpace` (Ω) |
| Constraint C_i | `Constraint` — but semantic: "must break class locality" |
| Fixpoint region Ω' | `ResidualRegion` (Ω_W) |
| Domain wipe-out | a falsified direction; provable dead region |
| Fail-first branching on residual shape | probing where the anti-vacuum is most determined |

## What GoW borrows

The dual reading itself: the anti-vacuum is, stated non-cinematically,
**semantic constraint propagation over a learned representation of Work** —
accumulated evidence removes degrees of freedom, and the *shape of what
remains* is an object worth reasoning about.

## Where GoW departs

CSP constraints are native, sound, and checkable; every removal is a theorem.
GoW's constraints are interpretive structural claims with graded evidential
strength — a "removal" is a hypothesis that can itself be challenged, weakened,
or falsified. GoW therefore needs the epistemic lifecycle that CSP gets free
from soundness.

## Reduction test (how this tradition attacks GoW)

> Propagation's power is soundness: removed means *provably* removed. If GoW
> "removes" regions on the strength of model judgment, the residual region is
> not a state of knowledge but a state of opinion — and cascading opinions
> compound error rather than deduction. Unsound propagation is just bias
> with extra steps.

Defense: never let a semantic constraint contract Ω at full strength without
independent verification; carry verification strength on every constraint and
let the residual region be explicitly stratified by it (strongly excluded /
weakly excluded / merely unsampled). Absence ≠ negation is already the
glossary's rule; the stratification makes it operational.
