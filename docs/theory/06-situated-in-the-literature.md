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

**This is the canonical statement of the lift.** Every later formulation in
this document — and any novelty sentence in the manuscript — is a restatement
of this one and is labeled as such.

## The closest ancestors

Each unit below is a real, mature idea GoW overlaps with, in one fixed form:
the tradition's **pattern**, what it **gives** GoW, and what it **withholds**.
Every "withholds" line is a *claimed* difference, pending the reduction test
its linked attack states in full. The point of listing them is discipline: if
GoW reduces to any one of them, it is not new. Full treatments live in
[docs/research/summaries/](../research/summaries/); the steelmanned objections
in the [composite attack surface](../research/composite-attack-surface.md).

### Lakatos — *Proofs and Refutations*

**Pattern.**

```text
conjecture → proof decomposition → counterexample → locate the "guilty lemma"
           → incorporate the exposed condition into a refined conjecture
```

**Gives GoW.** The philosophy of recursive refinement. The pattern is almost
exactly the GoW/`newf` challenge loop ([02-epistemic-model.md](02-epistemic-model.md),
[03-shape-guided-search.md](03-shape-guided-search.md)) — `I0 → challenge →
boundary_delta → I1 → …` — including the observation that a refined invariant
tends to expose a new micro-failure.

**Withholds (claimed).** The *Work geometry* — the space, its regimes, and the
distance/boundary structure over a population of attempts. Lakatos describes
one conjecture's biography, not a corpus.

→ *Deeper:* [summary 01](../research/summaries/01-lakatos-proofs-and-refutations.md)
· *Reduction test:* [A2](../research/composite-attack-surface.md#a2)

### CEGIS / CEGAR — counterexample-guided refinement

**Pattern.**

```text
candidate → checker → counterexample → constrained candidate space → new candidate
```

**Gives GoW.** The load-bearing point that a counterexample does not merely
reject *one* candidate — it excludes a whole *class* that would fail similarly:
the discrete, mature form of the complement-geometry "anti-vacuum"
([05-complement-geometry.md](05-complement-geometry.md)), `Ω_{t+1} = Ω_t \ F_t`.
CEGAR adds the representation-refinement instinct: refine the abstraction via
counterexamples until it is adequate.

**Withholds (claimed).** CEGIS/CEGAR know the specification and checker in
advance, and CEGAR earns each refinement with a spuriousness proof. GoW is
harder/weirder because the relevant *dimensions of the space* may themselves
have to be inferred from historical Work — with no soundness theorem to lean on.

→ *Deeper:* [summary 02](../research/summaries/02-cegis.md),
[summary 03](../research/summaries/03-cegar.md)
· *Reduction tests:* [A2](../research/composite-attack-surface.md#a2),
[A8](../research/composite-attack-surface.md#a8)

### Version spaces — candidate elimination

**Pattern.** Maintain the entire hypothesis set consistent with positive and
negative examples; each observation monotonically shrinks it, and the
surviving region is represented by its frontier.

**Gives GoW.** The eliminative reading of evidence behind the anti-vacuum:
the object worth representing is the surviving region and its boundary, not
the individual eliminations.

**Withholds (claimed).** A fixed hypothesis language with known semantics and
noise-free, boolean-labeled evidence. GoW's constraints are semantic, graded
in strength, and stated in a language that is itself under revision.

→ *Deeper:* [summary 04](../research/summaries/04-version-spaces.md)
· *Reduction test:* [A4](../research/composite-attack-surface.md#a4)

### Active learning / Bayesian experimental design

**Pattern.**

```text
ambient possibility-space → evidence → surviving region → disagreement/boundary
                          → maximally informative next probe
next Work ≈ argmax_W  information gained about the geometry
```

**Gives GoW.** The mature ancestor of the **Probe** reading: query where
surviving hypotheses disagree; choose experiments by expected information
gain; iterate from the observed landing point (sequential design).

**Withholds (claimed).** These frameworks search hypothesis/parameter spaces
with **known coordinates**, a fixed model class, and a likelihood. GoW wants
to infer *the coordinates of useful reasoning* from Work itself, with no
forward model from shape to outcome.

→ *Deeper:* [summary 05](../research/summaries/05-active-learning.md),
[summary 06](../research/summaries/06-bayesian-experimental-design.md)
· *Reduction test:* [A3](../research/composite-attack-surface.md#a3)

### Conceptual Spaces — Gärdenfors

**Pattern.** Concepts modeled geometrically over quality dimensions, with
similarity as distance and natural concepts as convex regions.

**Gives GoW.** Philosophical legitimacy for the core move: *semantic structure
can carry useful geometry even when it is not ordinary Euclidean space* —
precisely the disclaimer C1 already makes ("geometry" in the informal sense;
see [00-paper-claims.md](00-paper-claims.md)).

**Withholds (claimed).** Gärdenfors pays for the word "geometry" with axioms
(betweenness, convexity, a metric); GoW currently pays with component-wise
ordinal comparison and owes an analog of the naturalness criterion.

→ *Deeper:* [summary 07](../research/summaries/07-conceptual-spaces.md)
· *Reduction test:* [A6](../research/composite-attack-surface.md#a6)

### Instance Space Analysis

**Pattern.** Map problem *instances* by features, overlay algorithm
performance, and identify success/failure regions, holes, and
difficulty-associated features — iterating by generating data where the map
has gaps.

**Gives GoW.** Nearly the whole geometric reading: outcome-conditioned
regions, boundary and hole analysis, and the discipline of generating new
points where the map is uninformative. Structurally the nearest relative, and
the comparison to study hardest before any novelty claim.

**Withholds (claimed).** The lift is *what counts as a point*: ISA maps
instances with engineered features over large homogeneous populations; GoW
maps attempts described by inferred mechanism structure over small
heterogeneous ones.

→ *Deeper:* [summary 08](../research/summaries/08-instance-space-analysis.md)
· *Reduction tests:* [A1](../research/composite-attack-surface.md#a1),
[A7](../research/composite-attack-surface.md#a7)

### Fitness landscapes

**Pattern.** Map candidate solutions/configurations to a scalar fitness; the
surface's topography (peaks, valleys, ruggedness) governs what search can do.

**Gives GoW.** The classical baseline it generalizes — and the standing
warning that topographic language is operator-relative and easy to over-read.

**Withholds (claimed).** A defined (X, N, f) triple. GoW must state its analog
explicitly or confine its geometric vocabulary.

→ *Deeper:* [summary 09](../research/summaries/09-fitness-landscapes.md)
· *Reduction tests:* [A6](../research/composite-attack-surface.md#a6),
[A7](../research/composite-attack-surface.md#a7)

### Quality-Diversity — MAP-Elites

**Pattern.** Illuminate a behavioral feature space rather than driving
everything to one scalar optimum; different niches hold different high
performers.

**Gives GoW.** The instinct that there may be **multiple regimes, bridges,
voids, and boundaries** rather than one hill, and that structurally distinct
attempts — including failures — are worth preserving over ranking.

**Withholds (claimed).** QD describes *behavior*; GoW describes *mechanism*.
And while later QD (AURORA-style) learns its characterization too, it carries
no epistemic lifecycle: an archive entry is a measurement, never a claim under
challenge.

→ *Deeper:* [summary 10](../research/summaries/10-quality-diversity.md)
· *Reduction tests:* [A9](../research/composite-attack-surface.md#a9),
[A7](../research/composite-attack-surface.md#a7)

### Constraint propagation — feasible-region contraction

**Pattern.**

```text
Ω' = { x ∈ Ω : C_1(x) ∧ C_2(x) ∧ … ∧ C_n(x) }
```

**Gives GoW.** The nearest dual to "collapse the space around it": the
anti-vacuum is, less cinematically but more usefully, **semantic constraint
propagation over a learned representation of Work** — constraints such as
"must break class locality" rather than `x > 4`.

**Withholds (claimed).** Soundness. In CSP a removed value is *provably* in no
solution; GoW's removals are interpretive claims with graded strength, which
is why the epistemic lifecycle must be carried explicitly.

→ *Deeper:* [summary 11](../research/summaries/11-constraint-propagation.md)
· *Reduction test:* [A4](../research/composite-attack-surface.md#a4)

### Inverse design

**Pattern.**

```text
desired structural delta (cross boundary B, break P, retain Q,R)
   → inverse projection → domain-native candidate
```

**Gives GoW.** The shape of [classical projection](04-classical-projection.md):
derive desired properties in shape-space, then solve the inverse problem of
constructing an object that has them.

**Withholds (claimed).** The enabling asset — a cheap, trusted forward model
that scores candidates before the expensive step. GoW has no shape→outcome
evaluator; its affordable proxy is structural self-consistency checking, and
projection failure (falsification condition 4) is the standing risk.

→ *Deeper:* [summary 12](../research/summaries/12-inverse-design.md)
· *Reduction test:* [A8](../research/composite-attack-surface.md#a8)

## The synthesis table (canonical)

*This is the single canonical copy; other documents link here rather than
duplicating it.*

| GoW operation | Closest established relative | Summary | Reduction test |
|---|---|---|---|
| Treat semantic structure geometrically | Conceptual Spaces | [07](../research/summaries/07-conceptual-spaces.md) | [A6](../research/composite-attack-surface.md#a6) |
| Map features against outcomes | Instance Space Analysis / fitness landscapes | [08](../research/summaries/08-instance-space-analysis.md), [09](../research/summaries/09-fitness-landscapes.md) | [A1](../research/composite-attack-surface.md#a1), [A7](../research/composite-attack-surface.md#a7) |
| Preserve diverse regimes instead of scalarizing | Quality-Diversity / MAP-Elites | [10](../research/summaries/10-quality-diversity.md) | [A9](../research/composite-attack-surface.md#a9) |
| Learn from counterexamples | Lakatos / CEGIS / CEGAR | [01](../research/summaries/01-lakatos-proofs-and-refutations.md), [02](../research/summaries/02-cegis.md), [03](../research/summaries/03-cegar.md) | [A2](../research/composite-attack-surface.md#a2) |
| Collapse possibilities via negative evidence | Version spaces / constraint propagation | [04](../research/summaries/04-version-spaces.md), [11](../research/summaries/11-constraint-propagation.md) | [A4](../research/composite-attack-surface.md#a4) |
| Choose maximally informative next Work | Active learning / Bayesian experimental design | [05](../research/summaries/05-active-learning.md), [06](../research/summaries/06-bayesian-experimental-design.md) | [A3](../research/composite-attack-surface.md#a3) |
| Construct something satisfying an inferred target shape | Inverse design | [12](../research/summaries/12-inverse-design.md) | [A8](../research/composite-attack-surface.md#a8) |
| Iterate from an observed landing point | Sequential experimental design | [06](../research/summaries/06-bayesian-experimental-design.md) | [A3](../research/composite-attack-surface.md#a3) |

Assembled, GoW looks far less like an isolated invention and far more like a
**missing synthesis sitting between several mature fields**.

## What appears genuinely distinct

*This section is the operational restatement of the
[canonical lift statement](#the-one-sentence-position).*

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

Stated through the [naming canon](glossary.md#naming-canon-roles-not-synonyms)
(roles, not synonyms — a naming assignment, not an evidentiary change):

> **Geometry of Work studies the structure of problem-solving attempts. The
> Grammar of Attempts describes their composition and possible
> transformations. Consequential Cartography uses those descriptions to choose
> subsequent Work through the Cartographer's Loop.**

## Four missing rigorous pieces (the research program)

If GoW is the larger pattern, the theory currently lacks four things. These are
`hypothesis`/`future work` per the [claim registry](00-paper-claims.md), and
each aligns with an existing GoW falsification condition.
[08-earning-operational-authority](08-earning-operational-authority.md)
sharpens each into a precise missing contract with a minimum-closure
condition; the headings below are the coarse form.

1. **Representation learning.** What makes one Work-shape basis better than
   another? Needs a criterion beyond "compresses nicely" — the
   [abstraction-safety](../abstraction-safety.md) predictive-discrimination
   test is the seed, but the *selection* rule over bases in `M(W)` is unstated.
   *(Ties to falsification condition 1: geometry-is-vacuous; pressed hardest by
   [attack A4](../research/composite-attack-surface.md#a4).)*

2. **Geometry.** What do "distance," "boundary," "void," and "trajectory"
   formally mean when the dimensions are symbolic/semantic and partly learned?
   `newf` uses component-wise ordinal comparison (`CompareWithProfile`), which
   is a deliberate refusal of false precision, not yet a metric.
   [07-relational-structure](07-relational-structure.md) sharpens the stakes:
   marginal reads of shape axes can be blind to outcome-determining relations
   (Proposition R), so whatever geometry is adopted must be able to carry
   relational structure. *(Pressed hardest by
   [attack A6](../research/composite-attack-surface.md#a6).)*

3. **Causality.** How to distinguish a merely outcome-correlated shape from a
   genuine obstruction/enabler? The [epistemic model](02-epistemic-model.md)
   already forbids the leap (`observed regularity ≠ obstruction`; `claim_role`
   gates it), but forbidding the error is not the same as having a method that
   *establishes* causal role. *(Ties to condition 2: shape-space dominated;
   pressed hardest by [attack A5](../research/composite-attack-surface.md#a5).)*

4. **Convergence.** Under what conditions does repeated `map → probe → update`
   actually improve the geometry rather than produce an increasingly elaborate
   mythology? This is the one that keeps everyone employed, and it is the
   sharpest risk: a representation that can always be refined to "explain" the
   last failure is unfalsifiable. *(Ties to conditions 3 and 5; pressed hardest
   by [attack A2](../research/composite-attack-surface.md#a2).)*

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

That sentence is the manuscript-facing form of the
[canonical lift statement](#the-one-sentence-position), not a new claim.

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
- Per-tradition summaries of the ancestors named above live in
  [docs/research/summaries/](../research/summaries/); the strongest objections
  across the composite are consolidated in the
  [composite attack surface](../research/composite-attack-surface.md).
- [07-relational-structure](07-relational-structure.md) admits relational
  structure into the conceptual model (Proposition R proved; instrument
  calibrated by `internal/relational`); whether GoW exploits it is
  pre-registered as hypothesis H-R in
  [pilot-005](../../corpus/experiments/pilot-005-relational/PROTOCOL-DRAFT.md).
- [08-earning-operational-authority](08-earning-operational-authority.md)
  takes the four pieces above and states, per gap, the contract whose absence
  currently lends the theory unearned authority — with the minimum closure
  each requires.
