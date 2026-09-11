# CEGAR — Counterexample-Guided Abstraction Refinement

| Field | Value |
|---|---|
| **Canonical citation** | Clarke, E., Grumberg, O., Jha, S., Lu, Y., Veith, H. (2003). "Counterexample-Guided Abstraction Refinement for Symbolic Model Checking." *Journal of the ACM* 50(5), 752–794. (Conference version: CAV 2000.) |
| **Bib key** | `clarke2003cegar` (present in `paper/references.bib`) |
| **Field** | Formal verification / model checking |
| **GoW role** | Direct ancestor of the representation-refinement instinct; closest structural analog per the related-work matrix |
| **Epistemic status** | `GeneratedInterpretation` — written from model knowledge, **not** verified against the primary source. Verify before citation. |

## One-sentence core

Verify a large system by checking a small abstraction of it, and when the
abstraction produces a spurious counterexample, use that counterexample to
refine the abstraction until it is adequate for the property being checked.

## The mechanism (typed loop)

```text
abstraction  Â ← abstract(system M, property φ)
result       ← model-check(Â, φ)
if result = holds                → φ holds on M (abstraction is conservative)
if result = counterexample ĉ:
    if ĉ concretizable on M      → real bug; done
    else (spurious)              → refine Â to eliminate ĉ; loop
```

Typically the abstraction is predicate abstraction: states are grouped by
which of a finite predicate set they satisfy; refinement adds predicates that
distinguish the states the spurious trace conflated.

## Key load-bearing ideas

- **Spurious counterexamples are information about the representation, not
  the system.** A failed check can indict the *abstraction* rather than the
  artifact — the loop repairs the map, not the territory.
- **Adequate abstraction, not maximal detail** — refine only enough to decide
  the property at hand; the representation is property-relative.
- **Conservative abstraction** — the abstraction over-approximates behavior,
  so "holds on Â" soundly implies "holds on M"; the asymmetry between
  positive and negative verdicts is engineered, not accidental.
- **Refinement is localized** — the spurious trace tells you *where* the
  abstraction lost a distinction that matters.

## What it assumes is given in advance

- The concrete system `M` (fully formal, executable/checkable).
- The property `φ` (formal specification).
- A sound abstraction and concretization framework (e.g., predicate
  abstraction with a decision procedure).

## What it produces

Either a proof that `φ` holds, a real counterexample, or (as a byproduct) an
abstraction just fine-grained enough to decide `φ` — an earned representation.

## Mapping to GoW vocabulary

| CEGAR | GoW / `newf` term |
|---|---|
| Abstraction Â | a candidate representation in `M(W)` |
| Spurious counterexample | evidence that the current shape basis conflates outcome-relevant distinctions |
| Predicate addition (refinement) | `split` disposition; new shape axis |
| Adequacy for φ | abstraction-safety's predictive-discrimination test |

## What GoW borrows

The principle that **the representation is itself refinable and the refinement
is driven by observed failures of the representation** — the seed of `M(W)`
model selection and of `docs/abstraction-safety.md`'s round-trip grounding.

## Where GoW departs

CEGAR refines the abstraction of *one* formal system against *one* formal
property, with a sound concretization check to classify counterexamples as
real vs. spurious. GoW refines a representation of a **heterogeneous
population of attempts** with no formal ground truth and no sound spuriousness
test — which is precisely why its epistemic model must be carried explicitly
rather than derived from soundness theorems.

## Reduction test (how this tradition attacks GoW)

> CEGAR earns each refinement with a soundness argument. If GoW refines its
> shape basis whenever the geometry "fails to explain" an outcome, without a
> spuriousness test, refinement is unfalsifiable curve-fitting — CEGAR's loop
> with the load-bearing part deleted.

Defense requires a GoW analog of the spuriousness check: a criterion that
distinguishes "the representation was wrong" from "the claim was wrong"
before permitting a basis change.
