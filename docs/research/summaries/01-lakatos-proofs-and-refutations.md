# Lakatos — *Proofs and Refutations* (1976)

| Field | Value |
|---|---|
| **Canonical citation** | Lakatos, I. (1976). *Proofs and Refutations: The Logic of Mathematical Discovery*. Cambridge University Press. |
| **Bib key** | `lakatos1976proofs` (present in `paper/references.bib`) |
| **Field** | Philosophy of mathematics / methodology of science |
| **GoW role** | Philosophical ancestor of the recursive challenge/refinement loop |
| **Epistemic status** | `GeneratedInterpretation` — written from model knowledge, **not** verified against the primary source. Verify every attribution before it enters a manuscript. |

## One-sentence core

Mathematical knowledge grows not by accumulating verified theorems but by a
dialectical cycle in which conjectures are refined through the analysis of the
counterexamples that refute them.

## The mechanism (typed loop)

```text
conjecture
 → proof (decomposition into lemmas)
 → counterexample (global: refutes the conjecture; local: refutes a lemma)
 → lemma-incorporation (locate the "guilty lemma"; make its hidden
   condition an explicit premise of a refined conjecture)
 → refined conjecture
 → repeat
```

Staged as a fictional classroom dialogue about the Euler characteristic
conjecture `V − E + F = 2` for polyhedra. Successive counterexamples (the
picture frame, the cylinder, twin tetrahedra, star polyhedra) each force a
refinement of what "polyhedron" means.

## Key load-bearing ideas

- **Lemma-incorporation** — the productive response to a counterexample is
  to find which implicit assumption it violates and promote that assumption
  into an explicit condition. The counterexample *localizes* the failure.
- **Monster-barring** (named as a pathology) — redefining terms ad hoc to
  exclude a counterexample without learning anything. The degenerate response.
- **Exception-barring** — restricting the conjecture's domain to dodge the
  counterexample; safer than monster-barring but less informative than
  lemma-incorporation.
- **Proof-generated concepts** — the concepts a field ends up with (e.g.
  "simply connected") are *produced by* the refutation process, not given in
  advance. The vocabulary itself moves.
- **Heuristic vs. deductivist style** — presenting mathematics as a finished
  deductive edifice hides the refutation history that made it intelligible.

## What it assumes is given in advance

- A single conjecture under refinement (not a population of attempts).
- Human mathematicians as the operators; no formal representation of the
  attempt history.
- Counterexamples arrive from unmodeled creative search.

## What it produces

A philosophy of recursive knowledge refinement: failure as the primary
engine of concept formation.

## Mapping to GoW vocabulary

| Lakatos | GoW / `newf` term |
|---|---|
| Conjecture | `ShapeClaim` (`CandidateInvariant`) |
| Counterexample | challenge counterexample; `Boundary` |
| Guilty lemma / hidden condition | `boundary_delta` |
| Lemma-incorporation | `weaken` / `split` disposition; `I0 → challenge → I1` |
| Monster-barring | the pathology the epistemic model's promotion gates exist to block |
| Proof-generated concepts | representation movement in `M(W)` |

## What GoW borrows

The entire challenge lifecycle (`02-epistemic-model.md`,
`03-shape-guided-search.md`) is Lakatosian: a refined invariant is expected to
expose a new micro-failure, and the dispositions `survive | weaken | split |
falsify` are lemma-incorporation made operational.

## Where GoW departs

Lakatos supplies the philosophy of refinement but **no Work geometry**: no
population of attempts as a space, no regimes, no distance/boundary structure,
no persisted provenance, no policy mutation. He describes one conjecture's
biography; GoW claims structure over the *corpus* of attempts.

## Reduction test (how this tradition attacks GoW)

> If GoW's outputs are fully explained by "run Lakatosian refinement with
> bookkeeping," the geometric vocabulary (regimes, voids, distances) is
> decoration and GoW reduces to Lakatos-plus-a-database.

Defense requires demonstrating that population-level structure (cluster
geometry, complement geometry) generates moves that single-claim refinement
would not.
