# 05 — Complement Geometry

*Theoretical extension. Layer 1 (semantic) of the [theory series](./),
companion to [00-geometry-of-work.md](00-geometry-of-work.md). Where ordinary
GoW infers shape from the structure of completed work, complement geometry
infers shape from the deformation of possibility-space that work induces. This
document develops the dual view and its operators.*

**Epistemic status:** hypothesis — theoretically motivated, formally stated,
not yet experimentally tested. The existing pilots (003, 004) do not address
complement geometry. This document records the idea at the precision required
to test it, not at the precision earned by evidence.

---

## The dual question

Ordinary GoW asks:

> What structure do these attempts share?

Complement geometry asks the dual:

> What structure is forced by everything these attempts make unavailable?

The two questions are not equivalent. The first reads the occupied regions of
shape-space; the second reads the space those regions exclude. Both are
legitimate geometric readings of the same population of Work.

## Ordinary GoW: shape from occupied space

The current formulation:

```text
W → S(W) → G(W)
```

Attempts are points or regions in shape-space. Their clustering, boundaries,
conserved properties, and outcome gradients reveal the forward geometry. We
infer shape from the things *present*.

## Complement geometry: shape from displaced possibility-space

Start with an ambient possibility-space Ω — the space of structurally
plausible work on a problem, before any evidence is accumulated.

Every meaningful attempt does not merely add a point to Ω. It **constrains**
it. A failed attempt may establish:

- not this mechanism
- not this scope
- not this combination of structure
- not this abstraction level
- not under these assumptions

Each piece of Work induces a constraint or exclusion:

```text
W_i  ⟹  C_i ⊆ Ω   (region excluded or constrained by W_i)
```

The surviving space after all accumulated work:

```text
Ω_W  =  Ω \ ⋃_i F_i
```

where F_i is the forbidden region established by the i-th piece of work.

Or, formulated as constraint intersection:

```text
Ω_W  =  ⋂_i C_i
```

where each C_i is the region compatible with the evidence from W_i.

**Layperson:** every serious attempt presses material out of the
possibility-space. Eventually the negative space itself takes on a shape.
That shape is not one of the attempts. It is what the attempts collectively
force to remain possible.

## The anti-vacuum

The core idea:

> Nothing occupies the residual region — but it is highly informative
> *because* everything around it exerts constraints on it.

The metaphor is a mold:

```text
██████████████████
██████      ██████
█████        █████
██████      ██████
██████████████████
```

The interesting object is the hole. The material around it determines its
geometry. So:

```text
Shape(solution) ≈ Shape(negative space induced by Work)
```

Not the solution itself — reality demands proof. But potentially the
**structural requirements** of a solution: the properties any successful
approach must satisfy given what failure has established.

## The dual thesis

The original governing principle:

> **Search in shape-space; verify in domain-space.**

The dual:

> **Let accumulated Work deform possibility-space until the surviving
> topology reveals what a successful move must look like.**

Or shorter:

> **Infer the missing shape from the space Work cannot occupy.**

This is strictly stronger than elimination. Ordinary elimination says:

```text
A failed. B failed. C failed. Try D.
```

The complement-geometric move says:

```text
A deforms the space this way.
B excludes another dimension.
C establishes this boundary.

Their combined deformation leaves a residual basin
with properties P, Q, R.

Generate something satisfying P, Q, R.
```

That is not brute-force subtraction. It is **shape emergence from constraint
accumulation**.

## Two complementary operators

GoW now has two readings of the same population:

### Forward geometry

```text
Work → Shape
```

"What's common to the things we've tried?"

Yields: clusters, invariants, regimes, trajectories, gradients.

### Complement geometry

```text
Work → Deformation(Ω) → Residual Shape
```

"What has accumulated Work forced the remaining possibility-space to look
like?"

Yields: forbidden regions, cavities, narrow corridors, bottlenecks, residual
degrees of freedom, **necessary structural features**.

The two can be compared:

```text
Observed Shape  ↔  Residual Shape
```

Agreement between the forward reading and the complement reading is a
convergence signal: the occupied structure and the excluded structure tell a
consistent story about where a solution must live.

Disagreement is also informative: the forward geometry may be seeing a
regularity that the complement geometry does not support as a constraint,
or vice versa.

## The collapse operator

The complement geometry's primitive operation:

```text
C(Ω | W) → R
```

where:

- Ω = plausible work-space before accumulated evidence
- W = existing work
- C = constraint-collapse operator
- R = residual structural region

Then:

```text
properties(R) = {p_1, p_2, ..., p_n}
```

becomes the generator input:

```text
collapse(workspace)
    ↓
residual region
    ↓
emergent properties
    ↓
classical projection
    ↓
candidate
```

The collapse is not probabilistic. It is constraint collapse: many degrees of
freedom are progressively removed until only a lower-dimensional structural
manifold remains.

## Failure as pressure

In this view, a failure does not have to tell you the answer. It only has to
impose a legitimate structural constraint:

```text
F_i  ⟹  ¬R_i
```

Enough constraints produce a progressive compression:

```text
Ω  →[F_1]→  Ω_1  →[F_2]→  Ω_2  →[F_3]→  ···  →[F_n]→  Ω_n
```

with:

```text
dim(Ω_n)  ≪  dim(Ω)
```

Crucially, GoW lets you compress **semantically**, not by enumerating
candidates. A failure might remove:

> all approaches preserving class locality

rather than:

> candidate numbers 1 through 4,000.

The difference between semantic exclusion and enumeration is the difference
between a constraint that removes a structural dimension and one that removes
a finite number of points.

## Partial success as directional force

Failures tell you where not to live.

Partial successes tell you which direction the surviving region slopes.

- **Failure** supplies exclusion and boundaries.
- **Partial success** supplies gradient — the direction of outcome improvement.
- **Success** supplies attractor evidence.

The work-space begins to look like a **field**:

```text
F_Work(S) = pressure toward or away from structural state S
```

where accumulated evidence induces directional force:

```text
              success attractor
                    ◎
                 ↗  ↑  ↖
              ↗     |    ↖

 failure   ◉←── residual ──→◉ failure
 basin          cavity         basin

              ↘     |    ↙
                 ↘  ↓  ↙
              partial-success
                 gradient
```

Not physics. But a useful computational metaphor: the residual region is not
just defined by what is excluded but *oriented* by what partially succeeded.

## The void as structured absence

The most interesting case is a region with no existing Work in it.

```text
Failure regime A
Failure regime B
Partial-success regime C

There is a structurally coherent region R between B and C.
No historical approach occupies R.
```

If the surrounding boundaries all imply compatible constraints, the void may
have a surprisingly precise shape — more precise than any individual attempt
could establish by itself.

Then:

> The next candidate is generated not by extrapolating an existing method but
> by **materializing the missing region**.

This changes the prompt to the model from:

> "Invent a new idea."

to:

> "Instantiate an object that occupies this currently empty but structurally
> constrained region."

A much better-conditioned problem.

## The complementarity principle

> **Complementarity Principle of Work Geometry.**
> Work informs search in two ways: through the structures instantiated by
> completed attempts (*forward geometry*) and through the regions of
> possibility-space those attempts exclude or constrain (*complement
> geometry*). Sufficiently structured negative evidence can induce a residual
> geometry whose conserved properties specify candidate directions not
> instantiated by prior Work.

Shorter:

> **The shape of what remains can be inferred from the shape of everything
> that failed to occupy it.**

## Why community-resistant problems are attractive

A heavily worked problem with thousands of failed approaches may look
hopeless. Under the complement view:

```text
many diverse failures  ≠  no information
```

Potentially:

```text
many diverse, well-characterized failures  ⟹  highly constrained residual geometry
```

Community resistance can be an **asset**:

- A young problem has a giant unexplored Ω.
- A century-old resistant problem may have Ω_n with dim(Ω_n) ≪ dim(Ω).

The question mark is still unsolved. But the cavity around it may be
increasingly well defined.

```text
█████████████████████
████████    █████████
██████        ███████
██████   ?    ███████
██████        ███████
████████    █████████
█████████████████████
```

That is the anti-vacuum: a void that is not empty but structurally determined
by its surroundings.

## The three modes of failed Work

```text
Conventional search:   failed Work = consumed effort
Geometry of Work:      failed Work = material     (shape from occupied space)
Complement Geometry:   failed Work = pressure     (shape from excluded space)
```

The first discards. The second reads. The third inverts.

GoW as a complete discipline has two modes:

```text
infer shape from Work
    +
infer shape from the space Work eliminates
```

The second may ultimately be the more radical idea.

## What this document does not claim

- That complement geometry is immediately operationalizable in `newf`. The
  current system has voids as a concept
  ([03-shape-guided-search.md](03-shape-guided-search.md)) but does not
  implement the constraint-collapse operator or residual-region inference.
- That "possibility-space" Ω is well-defined without further work. Defining
  Ω rigorously is itself a research problem.
- That the anti-vacuum produces solutions. It produces **structural
  requirements** for candidates — the projection into domain-space and
  classical verification remain mandatory.
- That any of this is tested. This is a theoretical extension at hypothesis
  status.

## Relationship to the rest of the series

- The **[Geometry of Work](00-geometry-of-work.md)** is the forward theory;
  this document is the dual.
- The **[semantic model](01-semantic-model.md)** provides the shape vocabulary;
  complement geometry adds constraint, exclusion, and residual-region concepts.
- The **[epistemic model](02-epistemic-model.md)** applies unchanged: a
  complement-inferred property is a `ShapeClaim` at `proposed` status until
  challenged. Model-authored residual-region properties are model judgment,
  not verification.
- **[Shape-guided search](03-shape-guided-search.md)** already includes voids
  as one instrument the geometry exposes. Complement geometry develops the
  void concept into a full dual operator.
- **[Classical projection](04-classical-projection.md)** remains the mandatory
  crossing: a residual-region property is not a domain-space result until
  projected and verified.
