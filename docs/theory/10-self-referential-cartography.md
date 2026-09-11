# 09 — Self-Referential Programmatic Cartography

*Theory-series extension stating what it means, and does not mean, for a GoW
implementation to map its own problem-solving activity — including its own
representational choices — and use that map to guide subsequent Work. Its
epistemic status is `interpretation` + `hypothesis`/`future work` per the
[claim registry](00-paper-claims.md): it defines a capability and a validation
loop, and pre-commits the substantive claim to evidence. Nothing here is an
`observation`. The literature pointers (metareasoning; Russell & Wefald) are
**unverified pointers**, not checked citations — verify against primary
sources before any manuscript use.*

---

## Why this document exists

The [three-space decomposition](06-situated-in-the-literature.md#three-nested-spaces)
(`D ⇆ W ⇆ M(W)`) already names the layer where GoW selects among competing
representations of Work. This document answers a narrower question about that
layer:

> **Can a GoW implementation treat its own mapping and search decisions as
> Work — recording them, criticizing them, revising the representation that
> produced them, and letting the revision have an attributable, testable
> consequence for what it attempts next?**

The short answer, as an architectural capability, is yes, and *self-referential
programmatic cartography* is an accurate name for it — **provided it means the
system can map its own problem-solving activity and use that map to guide
subsequent Work.** It must not imply that describing itself establishes the
correctness of its descriptions. The boundary is **self-reference, not
self-certification**.

## The phrase, unpacked

> **A GoW implementation can enable self-referential programmatic cartography:
> constructing and revising explicit maps of problem-solving Work — including
> its own representational choices, search decisions, and outcomes — and using
> those maps to select subsequent Work.**

Each word carries a specific obligation, and each has an explicit disclaimer:

- **Self-referential** — the mapping process itself becomes eligible Work.
  Choosing a representation, grouping attempts, proposing a boundary, and
  selecting a probe can all be examined, not just the domain candidates they
  produce. *It does not mean the system's account of itself is privileged or
  self-validating.*
- **Programmatic** — these operations have explicit inputs, outputs,
  provenance, and executable relationships. *It does not require every
  interpretation to be deterministic or every decision to exclude human
  judgment.*
- **Cartography** — the output represents relationships, constraints, possible
  transformations, and uncertainty. *It need not be a literal spatial
  visualization or a metric geometry* — consistent with the operational
  meanings assigned to "distance," "boundary," and "trajectory" in
  [08](08-earning-operational-authority.md#give-each-geometric-term-an-operational-meaning).
- **Search-connected** — the map influences what the system attempts next.
  *Without that feedback it is self-documentation or analytical curation, not
  an adaptive search method* — the same distinction
  [FC5](00-paper-claims.md#falsification-conditions-for-gow-stage-8-material)
  draws between a curation framework and a search method.

## Two capabilities, not one

There are two distinct self-referential capabilities, and only the second is
the substantive extension.

**Mapping its own attempts** — "Which strategies did I try, under which
conditions, and what happened?" This is Work-space (`W`) bookkeeping about the
domain candidates. GoW already does a form of this.

**Mapping its own representational choices** — "Which distinctions did my
previous map erase? Which predictions failed because of that? Would another
representation support better decisions?" This is the substantive extension
into `M(W)`: treating **competing representations of Work as objects of inquiry
themselves**. That is exactly the conceptual role already assigned to the
[Theory-of-Work space](06-situated-in-the-literature.md#three-nested-spaces),
and the selection problem over that space is the open contract stated in
[08 §1](08-earning-operational-authority.md#what-should-select-among-bases-in-mathcal-mw).

The first is a ledger. The second is model selection over `M(W)` with the
selection *history itself* admitted as Work.

## The qualifying loop

A qualifying implementation supports a loop in which a representation change has
an attributable consequence, and subsequent evidence tests that consequence:

```text
Record its mapping and search decisions.
    ↓
Identify a limitation in the current representation.
    ↓
Propose a revised representation and a testable prediction.
    ↓
Use it to choose new domain Work.
    ↓
Check the outcome against the unchanged objective.
    ↓
Retain, revise, or reject the representation.
```

> **Merely feeding its previous output back into another prompt would not
> demonstrate this loop.** The representation change needs an attributable
> consequence for action, and subsequent evidence must test that consequence.

This is the same discipline [08 §4](08-earning-operational-authority.md#4-convergence-prevent-retrospective-repair-from-masquerading-as-learning)
requires of convergence — freeze the current representation, selected action,
predicted consequences, and interpretation rule before observing the result;
score that commitment; then allow revision. Self-reference does not relax that
requirement; it is the case where the object being revised is the mapper's own
representation.

## The boundary is self-reference, not self-certification

This capability is closely related to **metareasoning**: evaluating
computational choices by their consequences for external actions. The
metareasoning tradition (Russell & Wefald — *unverified pointer*) grounds the
utility of a computation in its ability to affect those external actions. The
cartographic formulation here is a **particular way of representing and
controlling that process** — not a claim to have invented self-directed
reasoning, and not a claim stronger than the metareasoning standard it borrows.

The safeguards are the same ones the series already imposes, applied to the
self-referential case:

- **Retain the previous predictions.** A revised map does not get to overwrite
  what its predecessor forecast.
- **Keep the objective fixed during the comparison.** The unchanged objective
  is what the outcome is checked against (cf. the fixed objective in the
  [08 convergence contract](08-earning-operational-authority.md#the-anti-mythology-rule-is-temporal)).
- **Do not let a revised map retroactively turn its predecessor's mistake into
  a successful prediction.** This is the temporal anti-mythology rule of
  [08 §4](08-earning-operational-authority.md#the-anti-mythology-rule-is-temporal),
  and it is precisely what forbids self-certification.

The existing [challenge contract](02-epistemic-model.md) supplies the
complementary guard: an unconfirmed model claim about a representation must not
silently strengthen or weaken a hypothesis's status
(`ModelJudgment ≠ Verification`). Self-reference gives the mapper a new class
of things to make claims *about*; it does not grant those claims a new
evidential standard.

## The hypothesis (role: `hypothesis`, untested)

> **H-SRC.** A GoW implementation can (a) record its own mapping and search
> decisions as Work, (b) detect a decision-relevant limitation in its current
> representation, (c) commit to a revised representation and a discriminating
> prediction *before* seeing the outcome, and (d) select subsequent domain Work
> whose checked outcome — against an unchanged objective — supports that
> prediction at a useful cost relative to not revising the representation.

Falsification paths (each inherits an existing falsification condition):

- **Self-documentation only.** The recorded self-map never changes what is
  attempted next; the loop is missing its search-connected edge
  ([FC5](00-paper-claims.md#falsification-conditions-for-gow-stage-8-material):
  curation, not search).
- **Retrospective repair.** Revisions improve the *account* of past failures
  without improving prospective decisions; frozen pre-revision scores do not
  get better ([FC3](00-paper-claims.md#falsification-conditions-for-gow-stage-8-material)
  and the [08 §4](08-earning-operational-authority.md#4-convergence-prevent-retrospective-repair-from-masquerading-as-learning)
  anti-mythology rule).
- **Prompt echo.** The "representation change" reduces to re-feeding prior
  output into another prompt with no attributable, separable consequence for
  the selected action.
- **Overhead without gain.** The extra self-mapping, adjudication, and
  explanation cost exceeds the Work it saves
  ([FC3](00-paper-claims.md#falsification-conditions-for-gow-stage-8-material)).

## What this document is and is not

**It is:** a definition of a capability, a name that survives scrutiny, and a
validation loop that connects a representation change to a testable
consequence. Its strongest defensible meaning is:

> **the cartographer's own Work becomes part of the territory it can study —
> while the usefulness of each revised map still has to be earned through
> subsequent Work.**

**It is not:** a claim that GoW has demonstrated this loop, that describing
itself certifies its descriptions, or that self-reference constitutes a new
verification mechanism. Recording and criticizing one's own representational
choices is a capability; establishing that the revised representation was
*better* is a separate achievement that only the prospective loop above can
earn.

## Relationship to the rest of the series

- The `M(W)` layer this document makes self-referential is defined in
  [06 — three nested spaces](06-situated-in-the-literature.md#three-nested-spaces);
  H-SRC admits the *selection history over `M(W)`* as Work.
- The representation-selection contract H-SRC leans on is
  [08 §1](08-earning-operational-authority.md#1-representation-learning-select-for-decisions-not-retrospective-discrimination-alone),
  and its anti-self-certification safeguard is the temporal anti-mythology rule
  of [08 §4](08-earning-operational-authority.md#the-anti-mythology-rule-is-temporal).
- The self-reference/self-certification boundary is enforced by the same
  `ModelJudgment ≠ Verification` guard as the
  [epistemic model](02-epistemic-model.md) and the challenge discipline used in
  [07](07-relational-structure.md#self-certification-is-excluded-by-construction).
- H-SRC joins the labeled-hypothesis list in the
  [claim registry](00-paper-claims.md#what-goes-under-hypotheses--future-work).
