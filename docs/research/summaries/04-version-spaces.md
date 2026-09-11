# Version Spaces — Candidate Elimination

| Field | Value |
|---|---|
| **Canonical citation** | Mitchell, T. M. (1982). "Generalization as Search." *Artificial Intelligence* 18(2), 203–226. (Earlier: Mitchell (1977), "Version Spaces: A Candidate Elimination Approach to Rule Learning," IJCAI.) |
| **Bib key** | `TODO: add mitchell1982generalization` to `paper/references.bib` |
| **Field** | Machine learning (symbolic concept learning) |
| **GoW role** | Mature ancestor of "collapse the possibility-space via negative evidence" |
| **Epistemic status** | `GeneratedInterpretation` — written from model knowledge, **not** verified against the primary source. Verify before citation. |

## One-sentence core

Maintain the *entire set* of hypotheses consistent with the examples seen so
far, and let each positive or negative example monotonically shrink that set.

## The mechanism (typed loop)

```text
version space  VS ← all hypotheses in language H
for each example (x, label):
    VS ← { h ∈ VS : h(x) = label }
representation: VS is bounded by
    S — set of maximally specific consistent hypotheses
    G — set of maximally general consistent hypotheses
positive example → generalize S minimally; prune G
negative example → specialize G minimally; prune S
converged when S = G (singleton)
```

## Key load-bearing ideas

- **Learning as search through a hypothesis space** with a partial order
  (general-to-specific) — the framing that made "generalization as search"
  a slogan.
- **The boundary-set representation** — the surviving region is compactly
  represented by its frontier (`S`, `G`), not by enumeration. Structure of
  the surviving set is the learned object.
- **Negative evidence is first-class** — negative examples do real work
  (they specialize `G`); the method is symmetric in a way pure induction
  from positives is not.
- **Bias is necessary** — with an unrestricted hypothesis language the
  version space never converges; the *choice of H* carries all inductive
  power (Mitchell's "need for biases in learning generalizations").
- **Ambiguity is explicit** — while `S ≠ G`, the learner *knows* which
  queries the surviving hypotheses disagree on (the hook active learning
  later exploits).

## What it assumes is given in advance

- A fixed hypothesis language `H` with known semantics and a
  generality ordering.
- Noise-free labels (a single mislabeled example can collapse VS to ∅).

## What it produces

The surviving hypothesis region and its frontier — a fully explicit
representation of "everything still possible given the evidence."

## Mapping to GoW vocabulary

| Version spaces | GoW / `newf` term |
|---|---|
| Hypothesis space H | `PossibilitySpace` (Ω) |
| Consistency pruning | `Constraint` (C_i); `ConstraintCollapse` |
| Surviving VS | `ResidualRegion` (Ω_W) |
| S/G frontier | the *shape* of the residual region — the anti-vacuum's boundary |
| Disagreement region | where the next probe is maximally informative |

## What GoW borrows

The complement-geometry reading: evidence is primarily *eliminative*, and the
object worth representing is the surviving region and its frontier, not the
individual eliminations.

## Where GoW departs

Version spaces demand a fixed hypothesis language with known coordinates and
noise-free, binary-labeled evidence. GoW's Ω is over *structurally plausible
work*, its constraints are semantic and graded in evidential strength, and its
language is itself under revision. GoW also never gets consistency for free —
"consistent with a failure" is an interpretive judgment, not a boolean.

## Reduction test (how this tradition attacks GoW)

> Without a defined hypothesis language, "Ω shrinks" is not a mechanism but a
> metaphor. If GoW cannot state what Ω is over, membership, and what exactly a
> failure eliminates, the anti-vacuum is version-space vocabulary with the
> algebra removed.

Defense requires at minimum a typed predicate language for constraints
(`invariant-predicate/v1` is the seed) and an explicit account of Ω's scope —
already flagged in the glossary as "defining Ω rigorously is itself a
research problem."
