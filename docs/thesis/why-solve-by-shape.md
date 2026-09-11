# Why Solve by Shape?

*This is a thesis, not a specification. The [theory series](../theory/) says
what the objects mean and what we are allowed to believe about them; the
[findings](../findings/) say what experiments have actually shown; the
[EPIC](../../EPIC.md) says what has been built. This document says why any of it
is worth building at all. For a reader-facing introduction that motivates the
move with a worked example before any vocabulary, see
[Why Map the Work?](why-map-the-work.md). For the practitioner's field guide —
when to reach for GoW and what to do — see
[When to Map the Work](when-to-map-the-work.md).*

---

## The move

The conventional representation of a hard problem contains both the object
being solved and the history of attempts to solve it. Most search operates
inside the former.

`newf` treats the latter as an object in its own right.

Attempts, failures, partial successes, and successes induce a structured
space of work. Conserved properties describe regions of that space.
Counterexamples reveal boundaries. Partial successes expose transitions.
Repeated failures identify bottlenecks. Unexplored combinations form voids.

The search problem can therefore be transformed:

```text
Do not search directly for the final solution.
Reconstruct the geometry of prior work.
Identify a structural transition toward a more successful regime.
Generate a candidate that crosses that boundary.
Compile that move back into the problem's classical representation.
Verify it there.
```

The system searches in shape-space and verifies in domain-space.

A candidate that fails after crossing one boundary is not merely a failed
candidate. It is a measurement of the geometry: it establishes that one
transition was traversable and exposes the next boundary.

Thus search becomes recursive cartography.

---

## Why this is not the usual thing

Most systems that "use an LLM on a hard problem" ask the model to grind harder
on one linear solution path. The model is asked to be, simultaneously,
researcher, workflow engine, scheduler, database, critic, judge, verifier, and
convergence controller. That is not intelligence; it is missing application
structure.

`newf` inverts the comparative advantage. The model's genuine strength is not
producing the final proof — deterministic and independently checkable tools are
better at truth-sensitive work. The model's strength is *reading structure*:
compressing many heterogeneous failed attempts into a candidate conserved
property, contrasting, re-representing, proposing a structural violation. So the
system uses the model as a **failure compressor and frontier generator**, and
keeps identity, provenance, epistemic state, verification, and search policy in
the repository where they can be made explicit, typed, replayable, and
falsifiable.

The distinction that makes the whole thing coherent is that shape-space and
domain-space are *different spaces with different truth conditions*. You are
allowed to reason fluently, associatively, and structurally in shape-space
because nothing there is being certified. The certification happens after the
move is compiled back down into the classical representation, where a
deterministic check, a computation, a proof obligation, or an independent source
decides the matter. Fluency generates leverage; grounding decides whether the
leverage still points at the original problem.

---

## Why a failed candidate is a measurement

The most important reframing is what happens on failure.

In direct search, a failed attempt is waste. You discard it and try again. This
is why the historical record of hard problems is biased toward what was worth
publishing: the failed programs, abandoned parameter choices, false lemmas, and
unproductive reformulations are under-recorded, precisely because a single
failure looked like nothing.

In shape-guided search, a failed candidate that crossed a boundary is a probe.
It tells you the transition was traversable and it exposes the next boundary. A
failed candidate that did *not* cross the boundary it claimed to cross tells you
your reconstruction of the geometry was wrong at that point. Either way the
failure narrows the space, and the narrowing is recorded as durable state rather
than lost to a chat transcript.

This is what it means to say **failure is a first-class sample space**, not a
regrettable byproduct.

---

## What we are *not* claiming

This thesis is deliberately narrow, and it is easy to inflate. It is worth
stating the non-claims plainly, because the discipline of the whole project
depends on them.

- We are **not** claiming that "geometry of work" is a metric space with a real
  distance function. It is, for now, a structured space with clusters, regions,
  boundaries, and ordinal comparisons. Fabricated precision is worse than
  honest coarseness.
- We are **not** claiming that an observed conserved property is an obstruction.
  *observed regularity ≠ obstruction.* A recurring shape is a hypothesis about
  the geometry, not a law of it.
- We are **not** claiming that a candidate produced by the system is a discovery.
  A model-proposed shape is model judgment; it becomes something stronger only
  after it survives challenge and, where a stronger mechanism exists,
  independent verification. *ModelJudgment ≠ Verification.*
- We are **not** claiming to have solved, or to be near solving, any open
  problem. The first serious benchmark is historical-holdout prediction, not
  open-problem theater.

The reason to write the thesis down anyway is that the *move* — treat the
history of attempts as a geometry, search in shape-space, verify in
domain-space, and let failed crossings measure the terrain — is the
project-level insight. It should have an authored statement before another
hundred implementation details bury the moment, and before implementation
terminology (which was chosen for SQLite migrations, not for concepts)
colonizes the idea.

---

## Why it might be false

A thesis worth stating is one that can fail. Concretely, "solve by shape" fails
if any of the following turns out to be true:

- **The geometry is an artifact of representation.** If the conserved properties
  the system finds are only shadows of how attempts were written up, then
  reconstructing the "shape of work" reconstructs nothing about the problem.
  (This is why abstraction must ground and invariants must be challenged above
  and below their level.)
- **Crossing a boundary carries no information.** If a candidate that violates a
  surviving failure invariant is no more likely to make progress than an
  undirected attempt, the transformation buys nothing. (This is exactly what the
  historical-holdout experiment is designed to test, against undirected and
  semantic-summary baselines.)
- **Shape generation is the easy part and grounding is intractable.** If moves
  are cheap to propose in shape-space but can never be compiled into a
  domain-space candidate that a real verifier can decide, then the system is an
  elegant prose generator.

The project's stopping conditions include `no_information_gain` and
`insufficient_failure_diversity` precisely because a null result is a legitimate
scientific outcome here. If failure history cannot be compressed into invariants
that predict productive search directions better than undirected generation,
the thesis needs revision — not more agent orchestration.

---

## Where this points

If searching in shape-space and verifying in domain-space beats the baselines
repeatedly on one problem, the next question is whether the effect survives a
domain jump — number theory to nonlinear PDE — with only adapters and verifiers
changing and the higher-order operators held constant. That is the litmus for
whether `newf` discovered a reusable research operator or merely became very
good at organizing one conjecture.

The wager is small to state and large if true:

> The history of failed work on a hard problem has usable shape, and the cheapest
> way to make progress is often to reconstruct that shape and cross one of its
> boundaries on purpose.
