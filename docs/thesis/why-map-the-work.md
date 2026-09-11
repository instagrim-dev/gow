# Why Map the Work?

*A reader-facing introduction to Geometry of Work. This chapter comes before
the vocabulary: it exists so a reader can experience the reasoning move once,
concretely, before any term of art is introduced. The
[thesis](why-solve-by-shape.md) argues why the approach is worth building; the
[field guide](when-to-map-the-work.md) says when a practitioner should reach
for it; [How to Map the Work](how-to-map-the-work.md) gives the reusable
recording method. This chapter only asks one question.*

> **Scope note.** The worked example below is a synthetic teaching model — a
> deterministic tick-based simulation, small enough to read in one sitting. It
> is **not** a hardware benchmark, **not** a new concurrency discovery, and
> **not** evidence that Geometry of Work outperforms conventional analysis. A
> competent engineer would find the same bottleneck by ordinary reasoning. Its
> only job is to let you experience the reasoning move before learning the
> terminology.

---

## The question

> **When does examining prior attempts teach us something that simply making
> another attempt would not?**

The default answer in most engineering practice is "rarely." A failed attempt
is discarded and the next attempt begins. Sometimes that is correct: when
attempts are cheap and independent, generating another one is the fastest way
to learn.

But attempts are not always independent. Sometimes several serious attempts
fail *the same way*, and the sameness is itself information — information that
no single additional attempt will surface, because each new attempt is built
from the same unexamined assumption that shaped the previous ones.

This chapter walks through one small, fully checkable case where three attempts
carry more information together than any of them carries alone.

---

## The setup

A system must apply **forty updates** spread across **four independent
streams** — ten updates per stream. Updates within a stream must apply in
order; updates in different streams have no ordering relationship at all.

Time is measured in **ticks** of a deterministic model. Each tick, workers may
prepare updates, and a *commit gate* admits prepared updates into the applied
state. The gate exists to protect ordering. As originally built, the gate is
**global**: it admits at most one commit per tick, across the whole system.

The team wants the forty updates applied in fewer ticks. Three attempts follow.

### Attempt 1 — add workers

One worker is obviously too few. The team raises the worker count from one to
four.

**Result: 40 ticks.** Exactly as before.

### Attempt 2 — add more workers

Perhaps four was not enough. The team raises the count to eight, so workers
outnumber streams two to one.

**Result: 40 ticks.**

### Attempt 3 — smarter scheduling

Perhaps the workers are colliding — all draining one stream while others sit
idle. The team replaces naive stream assignment with round-robin scheduling
that spreads workers evenly across streams every tick.

**Result: 40 ticks.**

Three attempts, three mechanisms, one number. At this point the ordinary
response is a fourth attempt: more workers still, work stealing, a priority
queue. Each of those is another sample from the same family.

The alternative is to stop generating and look at the attempts as a set.

---

## The unchanged restriction

Lay the three attempts side by side and describe each one *structurally* — not
"what we tried" but "what the attempt varied and what it held fixed":

| | Workers | Scheduling | Commit gate scope | Ticks |
|---|---|---|---|---|
| Attempt 1 | 1 → 4 | naive | **global** | 40 |
| Attempt 2 | 4 → 8 | naive | **global** | 40 |
| Attempt 3 | 8 | naive → round-robin | **global** | 40 |

Two columns varied. One column never did. Every attempt, whatever else it
changed, retained the global commit gate — and every attempt produced exactly
40 ticks, which happens to equal the total number of updates.

That coincidence is the first genuine piece of information the attempts
produce *as a set*: the completion time tracks the update count, not the
worker count and not the schedule. No single attempt shows this. The
comparison does.

Note what has and has not been established. We have observed a **commonality**
across three attempts. We have not yet shown it is an **explanation**. The
gate might be the bottleneck, or the sameness might be an artifact of how few
configurations were tried.

---

## A testable explanation

Turn the commonality into a claim that sticks its neck out:

> **Claim.** Completion time is governed by the commit gate's scope, not by
> worker count or scheduling. A gate admits one commit per tick per scope. So
> the model predicts:
>
> *ticks = the largest number of updates sharing a single gate scope.*

This claim makes two predictions that differ from "add more parallelism":

1. **Intervention.** With a global gate, all 40 updates share one scope, so
   40 ticks — regardless of workers or schedule. If the gate's scope is
   narrowed to *per stream* (one commit per stream per tick), the largest
   per-scope load is 10 updates, so completion should drop to **10 ticks** —
   again regardless of worker count above four.

2. **Control.** If the same forty updates are placed in a *single* stream,
   the per-stream gate degenerates back into a global gate: largest per-scope
   load is 40, so **40 ticks**, and the intervention's gain should vanish
   entirely.

The second prediction is what makes the explanation honest. An explanation
that only predicts success where you hope to find it is a slogan. This one
names the condition under which its own intervention must fail to help.

---

## A different intervention, and checked results

The intervention changes exactly one thing the three attempts never touched:
the gate's **scope**, from global to per-stream. Ordering within each stream is
still enforced — that is the invariant the gate exists to protect — but streams
no longer wait on each other's commits.

The runnable model in
[`examples/why-map-the-work/`](examples/why-map-the-work/) implements all five
configurations and checks them:

| Configuration | Gate scope | Streams | Predicted | Measured | Correct? |
|---|---|---|---|---|---|
| Attempt 1 (4 workers) | global | 4 × 10 | 40 | 40 | yes |
| Attempt 2 (8 workers) | global | 4 × 10 | 40 | 40 | yes |
| Attempt 3 (round-robin) | global | 4 × 10 | 40 | 40 | yes |
| Intervention (per-stream gate) | per-stream | 4 × 10 | 10 | **10** | yes |
| Control (single stream) | per-stream | 1 × 40 | 40 | 40 | yes |

"Correct" here means checked, not assumed: every configuration applies all
forty updates, and every stream's applied order equals its input order. The
speedup is not purchased by weakening the guarantee the gate was protecting.

The control row matters as much as the intervention row. The per-stream gate
buys nothing when there is one stream — exactly as the claim predicted. The
gain is a property of the *relationship* between gate scope and stream
structure, not a property of the new gate in isolation.

---

## The narrower claim

What is actually established at the end, stated with the boundaries visible:

> In this model, completion time equals the largest number of updates sharing
> a commit-gate scope. Worker count (at or above the stream count) and
> scheduling policy do not affect it. Narrowing gate scope helps exactly in
> proportion to how evenly the update load divides across scopes, and not at
> all when it doesn't divide.

Notice how much smaller that is than "per-stream gates are faster." The
narrower claim says when the intervention helps, by how much, and when it
doesn't — and each part of it was checked, including the part where it fails.

---

## The reasoning move, named after the fact

Here is what happened, as a sequence:

```text
three attempts
→ one unchanged restriction        (found by comparing, not by attempting)
→ a testable explanation           (predicts intervention AND control)
→ a different intervention         (varies the thing no attempt varied)
→ checked results                  (correctness verified, not assumed)
→ a narrower claim                 (states its own boundary)
```

The pivotal step is the second one. The unchanged restriction — the global
gate — was invisible to each attempt individually because every attempt
inherited it as background. It became visible only when the attempts were laid
side by side and described structurally. That is the answer to the opening
question:

> Examining prior attempts teaches something a new attempt would not **when
> the attempts share an unexamined restriction, because a new attempt drawn
> from the same family will inherit the restriction rather than reveal it.**

Everything else in this book is elaboration of that move: how to record
attempts so the comparison is possible ([How to Map the
Work](how-to-map-the-work.md)), when the move is worth making at all ([When to
Map the Work](when-to-map-the-work.md)), why we believe it scales past toy
models ([Why Solve by Shape?](why-solve-by-shape.md)), and what the
instrumented version looks like when the corpus of attempts is too large to
lay on a table ([the theory series](../theory/)).

---

## Objections a skeptical reader should raise

**"Any decent engineer would have found the gate."** Yes — in this model,
deliberately. The example is sized so the move is visible, not so the move is
necessary. The claim being illustrated is not "this analysis is hard" but
"this analysis has a shape, and the shape is the same when the problem is too
large to eyeball." Whether the move outperforms conventional analysis at scale
is an open experimental question, not something this chapter asserts.

**"Three attempts is a tiny sample. The 'unchanged restriction' could be
coincidence."** Correct, and that is why the middle step exists. The
commonality alone proved nothing; it was promoted only after it generated a
prediction that discriminated between explanations — including a control
configuration designed to remove the gain. Commonality is a hypothesis about
the attempts. It earns nothing until it survives a test it could have failed.

**"You designed the model so the gate was the bottleneck. Of course the story
works."** Also correct — this is the teaching-model disclaimer again, and it
is why the results table is *checked by tests rather than narrated*. The model
is honest about what it is: a controlled environment in which the reasoning
move can be run end-to-end and every arrow in the sequence verified. It is
evidence that the move is coherent and mechanizable, not evidence that it
wins.

**"Isn't this just root-cause analysis / Amdahl's law / theory of
constraints?"** The ingredients are old, and no originality is claimed for
them. What the rest of the book develops is the discipline around the move:
recording attempts structurally *before* you are stuck, distinguishing
observed commonality from established obstruction, requiring controls before
promotion, and keeping the epistemic status of every claim explicit when the
analysis is performed by a machine rather than an engineer at a whiteboard.

---

## Completion notes (for the author — remove before publication)

- The voice above is deliberately neutral. Passages that should carry your
  authorial voice: the opening two paragraphs of "The question," and the
  paragraph after the named reasoning move.
- The objections section is written in the order a skeptical technical reader
  is likely to raise them; reorder if your intended audience differs.
- The results table is generated by `go run ./docs/thesis/examples/why-map-the-work`
  and checked by its tests. If the model changes, regenerate the table rather
  than editing it by hand.
- Terminology is intentionally absent: "map," "shape," "invariant,"
  "boundary," and "frontier" first appear for the reader in the field guide.
  Keep it that way — this chapter's contract is *experience before
  vocabulary*.
