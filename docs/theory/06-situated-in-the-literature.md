# 06 — Situated in the Literature

*Research-program layer of the [theory series](./). Where 00–05 define Geometry
of Work (GoW) from the inside, this document places it from the outside: against
the mature neighboring fields it composes, the one lift that appears genuinely
distinct, and the rigorous pieces still missing. It is the honest "related work
+ open problems" companion to the [paper claim registry](00-paper-claims.md).*

**Epistemic status:** `interpretation` + `hypothesis`/`future work` per the
[claim registry](00-paper-claims.md). This document makes **no** novelty claim
of the form "nobody has thought of any of this." The opposite is true: pieces
are everywhere. The only defensible claim here is about the *level of
composition* and one *lift*. Named frameworks below are pointers to established
literatures, not verified citations — verify each against a primary source
before it enters a manuscript's related-work section.

---

## The one-sentence position

GoW is, roughly, **one meta-level above ordinary search**:

> Treat completed attempts themselves as the objects of a search space, infer
> semantic structure from their outcomes and relationships, then use that
> induced geometry to choose or construct the next attempt.

Most neighboring frameworks search a space of *candidates*, *hypotheses*, or
*instances*. GoW lifts landscape-based search into the space of the
problem-solving **Work** used to produce them. That lift is the thing to defend.

## The closest ancestors

Each row is a real, mature idea GoW overlaps with. The point of listing them is
discipline: if GoW reduces to any one of them, it is not new.

### Lakatos — *Proofs and Refutations* (recursive knowledge refinement)

The eerily close philosophical ancestor. Lakatos's pattern is:

```text
conjecture → proof decomposition → counterexample → locate the "guilty lemma"
           → incorporate the exposed condition into a refined conjecture
```

That is almost exactly the GoW/`newf` challenge loop
([02-epistemic-model.md](02-epistemic-model.md), [03-shape-guided-search.md](03-shape-guided-search.md)):

```text
I0 → challenge → boundary_delta → I1 → …
```

including the observation that a refined invariant tends to expose a new
micro-failure. Lakatos gives GoW a **philosophy of recursive refinement**. What
he does not give is the *Work geometry* — the space, its regimes, and the
distance/boundary structure over a population of attempts.

### CEGIS / CEGAR (counterexample-guided refinement)

From program synthesis and formal verification:

```text
candidate → checker → counterexample → constrained candidate space → new candidate
```

The load-bearing point is that a counterexample does not merely reject *one*
candidate — it excludes a whole *class* that would fail similarly. That is the
discrete, mature form of GoW's complement-geometry "anti-vacuum"
([05-complement-geometry.md](05-complement-geometry.md)):

```text
Ω_{t+1} = Ω_t \ F_t
```

CEGAR similarly begins with an abstraction and refines it via counterexamples
until the representation is adequate — a direct ancestor of GoW's
representation-refinement instinct. The difference: **CEGIS/CEGAR know the
specification and checker in advance.** GoW is harder/weirder because the
relevant *dimensions of the space* may themselves have to be inferred from
historical Work.

### Version spaces / active learning / Bayesian experimental design

Version-space learning maintains the hypothesis set consistent with positive and
negative examples; evidence shrinks it. Active learning focuses queries where
surviving hypotheses disagree (maximal information per observation). Bayesian
experimental design formalizes "probe where expected information gain is
highest." Together they are the mature ancestor of GoW's **Probe** reading:

```text
ambient possibility-space → evidence → surviving region → disagreement/boundary
                          → maximally informative next probe
next Work ≈ argmax_W  information gained about the geometry
```

The recurring difference: these frameworks search hypothesis/parameter spaces
with **known coordinates**. GoW wants to infer *the coordinates of useful
reasoning* from Work itself.

### Where "geometry" has precedent

- **Conceptual Spaces (Gärdenfors).** Concepts modeled geometrically over
  quality dimensions, with similarity and concept formation spatial. This gives
  philosophical legitimacy to GoW's core move: *semantic structure can carry
  useful geometry even when it is not ordinary Euclidean space* — precisely the
  disclaimer C1 already makes ("geometry" in the informal sense; see
  [00-paper-claims.md](00-paper-claims.md)).
- **Instance Space Analysis (ISA).** Maps problem *instances* by features,
  overlays algorithm performance, and identifies success/failure regions, holes,
  and difficulty-associated features — even iterating by generating data where
  the map has gaps. Structurally the nearest relative to GoW's
  geometry-of-performance. The lift is *what counts as a point* (below).
- **Fitness landscapes.** Map candidate solutions/configurations to a scalar
  fitness. The classical baseline GoW generalizes.

### Quality-Diversity (QD / MAP-Elites)

Illuminates a behavioral feature space rather than driving everything to one
scalar optimum; different niches hold different high performers. Compatible with
GoW's instinct that there may be **multiple regimes, bridges, voids, and
boundaries** rather than one hill. But QD generally *starts* with a behavior
characterization; GoW's harder problem is that the characterization may itself
have to emerge from comparing Work.

### Constraint propagation / feasible-region contraction

The nearest dual to "collapse the space around it." Begin with `Ω`; every
observation imposes a constraint and the feasible region shrinks:

```text
Ω' = { x ∈ Ω : C_1(x) ∧ C_2(x) ∧ … ∧ C_n(x) }
```

GoW's twist: the constraints are not native problem constraints (`x > 4`) but
**semantic structural constraints** ("must break class locality", "cannot rely
solely on asymptotic coverage"). So the anti-vacuum is, less cinematically but
more usefully, **semantic constraint propagation over a learned representation
of Work**.

### Inverse design

Materials science distinguishes forward discovery from **inverse design**:
specify desired functionality, then search for structures realizing it. GoW's
[classical projection](04-classical-projection.md) is an inverse-design step:

```text
desired structural delta (cross boundary B, break P, retain Q,R)
   → inverse projection → domain-native candidate
```

"Derive desired properties in shape-space, then solve the inverse problem of
constructing an object that has them."

## The synthesis table

| GoW operation | Closest established relative |
|---|---|
| Treat semantic structure geometrically | Conceptual Spaces |
| Map features against outcomes | Instance Space Analysis / fitness landscapes |
| Preserve diverse regimes instead of scalarizing | Quality-Diversity / MAP-Elites |
| Learn from counterexamples | Lakatos / CEGIS / CEGAR |
| Collapse possibilities via negative evidence | Version spaces / constraint propagation |
| Choose maximally informative next Work | Active learning / Bayesian experimental design |
| Construct something satisfying an inferred target shape | Inverse design |
| Iterate from an observed landing point | Sequential experimental design |

Assembled, GoW looks far less like an isolated invention and far more like a
**missing synthesis sitting between several mature fields**.

## What appears genuinely distinct

Most neighbors assume, in advance, one or more of: the feature space, the
hypothesis language, the specification, the fitness function, the search
neighborhood, the candidate representation, or the target property.

GoW's more ambitious loop lets the **representation itself move**:

```text
Work
 → infer representation
 → infer geometry
 → infer boundary
 → derive target structural move
 → construct domain candidate
 → observe outcome
 → revise representation
```

It is not merely searching a landscape. It is partially **learning what
landscape would make the history of Work intelligible**, acting inside that
landscape, and testing whether the representation predicted something useful.
That is closer to **scientific theory formation** than to standard optimization.

## Three nested spaces

The most useful decomposition of what GoW is actually doing:

```text
D  — Domain space        ordinary objects: proofs, programs, molecules,
                          architectures, legal arguments, experiments
W  — Work space          attempts over those objects, described by shape:
                          operators, assumptions, representation, preserved
                          and broken properties, auxiliary objects.
                          This is where `newf` currently lives.
M(W) — Theory-of-Work     competing representations of W: candidate geometries.
       space              "maybe class-locality is the axis" / "challenge shows
                          that's too broad — split it" / "this basis predicts
                          outcomes better."
```

The forward geometry ([00](00-geometry-of-work.md)) and its dual
([05](05-complement-geometry.md)) both live in `W`. The genuinely spicy part is
`M(W)`: GoW is doing **model selection over candidate geometries**. The full
loop is a three-way traffic:

```text
D ⇆ W ⇆ M(W)
```

Actions discovered in `M(W)` compile back through `W` into `D` via
[classical projection](04-classical-projection.md). The current
[three-axis invariant model](../domain-model.md) (`shape × observed_regime ×
epistemic_status`) is an early, concrete piece of `M(W)`: a candidate
representation that challenge can `split`/`weaken`, i.e. model selection in
miniature.

## Disciplinary description

Deliberately not "a new branch of mathematics or AI." The defensible
descriptions today:

> **Geometry of Work is a proposed meta-search methodology combining semantic
> representation learning, outcome-conditioned landscape analysis,
> counterexample-guided refinement, and inverse design over populations of
> prior problem-solving attempts.**

or, more provocatively:

> **GoW lifts landscape-based search from the space of candidate solutions into
> the space of problem-solving Work itself.**

```text
classical landscape:   candidate        → fitness
GoW:                   shape of attempt → outcome
eventually:            history of shapes/outcomes → geometry → next attempt
```

## Four missing rigorous pieces (the research program)

If GoW is the larger pattern, the theory currently lacks four things. These are
`hypothesis`/`future work` per the [claim registry](00-paper-claims.md), and
each aligns with an existing GoW falsification condition.

1. **Representation learning.** What makes one Work-shape basis better than
   another? Needs a criterion beyond "compresses nicely" — the
   [abstraction-safety](../abstraction-safety.md) predictive-discrimination
   test is the seed, but the *selection* rule over bases in `M(W)` is unstated.
   *(Ties to falsification condition 1: geometry-is-vacuous.)*

2. **Geometry.** What do "distance," "boundary," "void," and "trajectory"
   formally mean when the dimensions are symbolic/semantic and partly learned?
   `newf` uses component-wise ordinal comparison (`CompareWithProfile`), which
   is a deliberate refusal of false precision, not yet a metric.

3. **Causality.** How to distinguish a merely outcome-correlated shape from a
   genuine obstruction/enabler? The [epistemic model](02-epistemic-model.md)
   already forbids the leap (`observed regularity ≠ obstruction`; `claim_role`
   gates it), but forbidding the error is not the same as having a method that
   *establishes* causal role. *(Ties to condition 2: shape-space dominated.)*

4. **Convergence.** Under what conditions does repeated `map → probe → update`
   actually improve the geometry rather than produce an increasingly elaborate
   mythology? This is the one that keeps everyone employed, and it is the
   sharpest risk: a representation that can always be refined to "explain" the
   last failure is unfalsifiable. *(Ties to conditions 3 and 5.)*

These are the *right* holes — the kind that appear when something is becoming a
research program rather than a clever workflow.

## The read

GoW does not appear to be one thing someone else already named. It appears to
be an independent convergence on a **composition of several deep patterns**,
with one potentially novel lift:

> from reasoning about candidates / hypotheses / instances,
> to reasoning geometrically about **the accumulated population of Work** used
> to create them.

The comparisons to study hardest before any novelty claim are **Lakatos**
(conceptually closest to the recursive micro-failure/refinement cycle) and
**Instance Space Analysis** (structurally closest to geometry-of-performance),
with **CEGIS / version spaces** the strongest ancestors of the anti-vacuum
collapse. If GoW survives serious comparison against those without reducing to
one of them, the interesting thesis is not "we invented failure analysis." It is:

> **We unified counterexample refinement, semantic geometry, outcome
> landscapes, and inverse construction at the level of Work itself.**

## Relationship to the rest of the series

- The lift is over [00-geometry-of-work.md](00-geometry-of-work.md)'s objects;
  the dual "collapse" is [05-complement-geometry.md](05-complement-geometry.md).
- The `M(W)` model-selection layer is the formal home of what the
  [epistemic model](02-epistemic-model.md) enforces (challenge can `split`/
  `weaken` a representation) and what [abstraction safety](../abstraction-safety.md)
  disciplines (a new basis must preserve predictive discrimination).
- Every claim here is bounded by the [paper claim registry](00-paper-claims.md);
  nothing in this document is an `observation` — the empirical record is
  [Finding 001](../findings/001-unencoded-shape-generation.md) and the pilots.
