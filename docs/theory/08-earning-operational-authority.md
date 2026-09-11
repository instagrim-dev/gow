# 08 — Earning Operational Authority

*Research-program layer of the [theory series](./), sharpening the
[four missing rigorous pieces](06-situated-in-the-literature.md#four-missing-rigorous-pieces-the-research-program)
named in [06](06-situated-in-the-literature.md). Where 06 places GoW against
its neighbors and names the gaps, this document states what closing each gap
would actually require — as contracts, not aspirations.*

**Epistemic status:** `interpretation` + `hypothesis`/`future work` per the
[claim registry](00-paper-claims.md). This is a **proposed formal direction
for GoW, not a claim that these guarantees are already implemented or
demonstrated.** The two mathematical results included (a decision-loss bound
and a candidate-elimination bound) are elementary conditional arguments,
labeled as such — their value is in identifying what would need to be
preserved or assumed, not in their novelty. Named literatures (approximate
state abstraction, abstract interpretation, causal abstraction, causal
representation learning, prequential evaluation, adaptive data analysis) are
pointers, not verified citations — verify against primary sources before any
manuscript use.

---

## The one question behind the four gaps

The four gaps are real, but they are not four independent missing chapters.
They meet at one question:

> **Can GoW learn a representation of prior Work that preserves the
> distinctions needed to choose — and successfully realize — the next
> action?**

A representation can explain the archive beautifully and still recommend
useless interventions. A plausible intervention can be impossible to
construct. A successful construction can be credited to the wrong feature. An
adaptive system can explain every disappointment retrospectively without
becoming better at choosing.

That is the chain the theory needs to make accountable.

The [claim registry](00-paper-claims.md) correctly leaves convergence,
comparative effectiveness, and complement geometry unestablished. The
[abstraction-safety contract](../abstraction-safety.md) already asks whether
an abstraction merges cases requiring different search policies and whether
generated moves survive grounding. **Those are the strongest starting points —
not the visual resemblance to a geometric space.**

The four gaps, sharpened:

| Existing heading ([06](06-situated-in-the-literature.md#four-missing-rigorous-pieces-the-research-program)) | More precise missing contract |
|---|---|
| Representation learning | **Which abstractions preserve useful decisions, at an acceptable cost?** |
| Geometry | **Which relationships and transformations are meaningful and realizable?** |
| Causality | **When does changing the represented structure change the result, and what does that establish?** |
| Convergence | **How does the system earn credit from future evidence rather than repair its account of the past?** |

---

## 1. Representation learning: select for decisions, not retrospective discrimination alone

### The deeper gap

"Preserves predictive discrimination" ([abstraction safety](../abstraction-safety.md))
is necessary for many uses, but underspecified: **prediction of what?**

A representation might perfectly separate historically successful and
unsuccessful Work using features that are unavailable before execution,
reflect author skill, or encode the outcome itself. It could also correctly
classify two approaches as failures while erasing the distinction that
determines how to repair them.

For example, two implementations both miss a deadline. One is limited by a
shared gate; the other by insufficient processing capacity. A representation
grouping them as "insufficient throughput" describes their outcomes
accurately. It is inadequate for choosing between narrowing the gate and
adding capacity.

The stronger criterion is:

> **A useful Work-shape basis preserves the distinctions needed to predict the
> consequences of available next actions — not merely the labels attached to
> completed attempts.**

This has a substantive mathematical ancestor. Approximate state-abstraction
research (the MDP-abstraction line of Li–Walsh–Littman and successors)
evaluates abstraction by the quality of behavior it preserves, including
bounds on lost optimality. State-action abstraction also studies which
abstract actions can preserve useful policies. GoW can borrow that standard
without assuming historical research attempts already satisfy the assumptions
of a Markov decision process.

### A concrete selection target

Let:

\[
\phi_m(w,c)=z
\]

where \(w\) is a Work item, \(c\) its relevant context, and \(m\) a candidate
representation in \(\mathcal M(W)\)
(the [Theory-of-Work space](06-situated-in-the-literature.md)). The output
\(z\) is its shape description.

**In ordinary language:** the same implementation can have a different
significance under a different workload, objective, or resource budget. Those
conditions must not disappear when it becomes a point on the map.

Now let \(Q(w,c,a)\) denote the expected value of performing a particular
admissible action \(a\), under a fixed objective and budget. An abstract model
predicts:

\[
\widehat Q_m(\phi_m(w,c),a).
\]

The representation is useful when those predictions support good action
selection.

Here is a modest conditional result we can establish immediately. Suppose, for
a fixed set of feasible actions,

\[
\left|Q(w,c,a)-\widehat Q_m(\phi_m(w,c),a)\right|\leq\varepsilon
\]

for every available action. Selecting the action with the highest predicted
value then loses at most \(2\varepsilon\) relative to the best action.

The argument is simply:

\[
Q(a^*)\leq \widehat Q(a^*)+\varepsilon
\leq \widehat Q(\hat a)+\varepsilon
\leq Q(\hat a)+2\varepsilon.
\]

**Translation:** prediction error can overrate the chosen action and
underrate the best action; together those errors cost at most twice the error
bound.

This is an elementary decision bound, **not a novel GoW theorem**. Its value
is that it identifies what a representation would need to preserve. The
difficult research problem is obtaining an adequate bound — or convincing
prospective evidence of useful action selection — from sparse, heterogeneous
Work histories.

### What should select among bases in \(\mathcal M(W)\)?

The proposed hierarchy:

**First, admissibility.** A basis must retain the problem specification,
provenance, relevant context, and uncertainty. It cannot win by encoding
outcomes into supposedly pre-outcome features or by quietly changing the task.

**Second, prospective decision performance.** Compare what policies using the
competing bases actually select and achieve on unseen episodes, under the same
total resource budget.

**Third, economy.** Among comparably useful bases, prefer the one requiring
less extraction, adjudication, computation, and explanation.

Compression becomes a cost advantage or tie-breaker — not proof of usefulness.

The comparison also needs an ordinary interaction-aware baseline, not merely a
deliberately weak feature summary — the same discipline
[pilot-005](../../corpus/experiments/pilot-005-relational/PROTOCOL-DRAFT.md)
already imposes on H-R. Otherwise the
experiment establishes that a richer representation beats an impoverished one,
rather than establishing value for GoW's particular procedure.

### What this changes about falsification condition 1

**Different geometries are not automatically evidence that geometry is
vacuous** ([FC1](00-paper-claims.md#falsification-conditions-for-gow-stage-8-material)).

Two representations can use different coordinates while making equivalent
useful predictions and selecting equivalent actions. Conversely, two diagrams
can look alike while prescribing different interventions.

The operational form of condition 1 should be:

> Reasonable changes in representation repeatedly produce incompatible,
> consequential predictions or actions, and those differences cannot be
> resolved by prospective evidence.

That is a sharper failure condition than requiring one uniquely correct map.
It also avoids treating the inferred geometry as an intrinsic object that must
exist independently of the task.

**Minimum closure for this gap:** a declared representation-selection rule, a
decision-relevant comparison task, and a prospective test showing whether a
basis earns its additional cost.

---

## 2. Geometry: define permitted transformations before demanding a metric

### "Not yet a metric" is not necessarily a deficiency

GoW does not need Euclidean distance, a smooth manifold, or a single
similarity score to become rigorous. It needs to specify **which relationships
justify which operations**.

The proposed starting object is a **typed, directed graph of admissible
transformations**, not a point cloud.

Its nodes represent Work or scoped classes of Work. Its edges represent
changes such as replacing an operator, weakening an assumption, adding an
auxiliary construction, changing a coordination structure, or refining a
representation. Each edge carries its preconditions, preservation obligations,
evidence status, and cost information.

This connects naturally to abstract interpretation (Cousot–Cousot): reasoning
in an abstract domain is useful when its relationship to concrete computations
is specified and sound, even when the abstraction is deliberately incomplete.
That literature offers more than a geometric metaphor; it supplies a model for
making the abstraction–concrete relationship explicit.

### Give each geometric term an operational meaning

| Term | Proposed formal meaning |
|---|---|
| **Distance** | Cost or an ordering over admissible transformation paths — not necessarily semantic similarity |
| **Boundary** | A transition or cut separating regions under an explicitly stated outcome criterion |
| **Void** | A described region with no observed member; feasibility remains a separate question |
| **Trajectory** | A sequence of actual or proposed transformations, retaining the concrete artifacts and observations |
| **Invariant** | A property preserved under a specified class of transformations and conditions |
| **Regime** | A region associated with a scoped outcome pattern; initially descriptive, not necessarily causal |

Several consequences follow.

**Distance can be asymmetric.** Adding a restrictive assumption may be easy;
removing dependence on it may be difficult. A symmetric similarity measure
does not describe that asymmetry.

**No known path is not a proof of impossibility.** In an incompletely explored
graph, failure to find a route establishes a knowledge gap, not infinite
distance.

**A void need not be a missing invention.** It can be an unrecorded region, an
infeasible combination, or a distortion introduced by the representation.

**A boundary is not automatically minimal.** One successful intervention
establishes a witnessed transition. Establishing the smallest necessary change
requires additional comparisons or an argument.

These distinctions materially qualify the "fabric collapsing onto the
solution" prose of [05-complement-geometry](05-complement-geometry.md). That
metaphor was ahead of the formal justification.

### The less obvious problem: abstract paths may not compose

Suppose the map contains:

```text
Region A → Region B → Region C
```

The first edge may be witnessed by a concrete construction that enters one
part of B. The second edge may be available only from a different,
incompatible part of B.

Both edges can be individually legitimate while the apparent A→C route is
unrealizable.

This is a serious candidate for a central GoW obligation:

> **A proposed trajectory through shape-space must admit a compatible sequence
> of domain-level realizations.**

Checking each structural change independently is insufficient if their
preconditions do not compose.

This is where the registry's
**[condition 4](00-paper-claims.md#falsification-conditions-for-gow-stage-8-material)
— domain projection as the bottleneck — belongs**. It is not an unrelated
fifth concern. It is the connection between geometry and action. A map can be
descriptively accurate while its apparent shortcuts cannot be constructed.

### Separate observed regularity from transformation invariance

A sampled population can share a property without establishing that any
operation preserves it.

For a stronger invariant claim, specify something like:

\[
p(T(w))=p(w)
\]

for the allowed transformations \(T\), within a stated scope.

**Translation:** name what is allowed to change and what is claimed to remain
unchanged.

The earlier population-level shape claims remain useful. They simply do not
acquire this stronger preservation meaning merely because we call them
invariants.

**Minimum closure for this gap:** a transformation vocabulary, explicit edge
preconditions and preservation obligations, and at least one demonstrated
abstract path whose concrete realization is checked. A universal metric can
wait.

---

## 3. Causality: distinguish usefulness, intervention effects, and impossibility

### There are three different targets here

| Claim | What would support it |
|---|---|
| "This shape predicts outcomes." | Successful prediction on relevant unseen cases |
| "Changing this feature changes outcomes." | A well-defined intervention with an appropriate comparison or justified causal identification |
| "Any method retaining this restriction cannot succeed." | A scoped impossibility argument, exhaustive result in a finite setting, or another guarantee supporting that quantifier |

Those are not successive confidence levels on one scale. They are different
propositions.

A predictive representation can help search without establishing causes. A
demonstrated causal improvement does not prove that the modified feature is
necessary for every successful method.

**Therefore,
[condition 2](00-paper-claims.md#falsification-conditions-for-gow-stage-8-material)
— shape-space search being dominated — is a test of utility, not by itself a
test of causal validity.** GoW could improve search using correlations, or
identify a real causal restriction while spending too much to make the detour
worthwhile.

### "Change the shape" is not yet a well-defined intervention

Consider:

> Add global coupling.

That could mean shared state, message passing, a global constraint solver, a
proof lemma linking cases, or a completely different construction. Those
implementations may have different consequences.

The causal question becomes meaningful only after specifying what changes in
domain-space and what remains fixed.

A useful consistency target is:

\[
\phi_m(T_a(w))\approx \overline T_a(\phi_m(w)).
\]

**Translation:** executing the concrete intervention and then describing its
shape should agree with what the abstract transformation predicted. For
stochastic systems, the agreement concerns outcome distributions, not
necessarily identical individual realizations.

This is close to existing work on causal abstraction: models at different
levels should agree about the effects of appropriately related interventions.
Rubenstein and colleagues formalize such consistency between structural
equation models, emphasizing the importance of specifying the interventions
themselves.

It also narrows the novelty claim. Learning useful high-level causal variables
is already an explicit research problem in causal representation learning
(the Schölkopf-line literature). GoW's research opportunity is not simply "the
coordinates can be learned"; it is whether this can be done productively over
heterogeneous histories of problem-solving Work.

### GoW needs two grounding routes, not one universal causal checker

**For empirical systems**, the route is controlled intervention, matched
comparisons, or another defensible identification strategy. The claim remains
scoped to the tested conditions and assumptions.

**For mathematical or formally specified systems**, an obstruction can be
established deductively.

The [teaching example](../thesis/why-map-the-work.md) illustrates the second
route. Under its specified gate semantics, forty updates needing one admission
each and a gate admitting at most one per tick imply at least forty ticks.
That establishes a restriction on **every implementation retaining that gate
under those assumptions**, not merely a pattern among three runs. The
per-stream implementation is then separately
[checked against the unchanged task](../thesis/examples/why-map-the-work/).

The architecture should therefore connect to domain-appropriate evidence
procedures, rather than expect `challenge` to manufacture causal authority
from agreement.

### This determines when complement geometry may exclude regions

The strong exclusion claim is:

\[
\forall w\in\mathcal C,\quad p(w)\Rightarrow \neg G(w),
\]

where \(\mathcal C\) is a defined method class and \(G(w)\) means satisfying
the original goal.

**Translation:** within this class, retaining property \(p\) prevents success.

A single failed attempt usually does not establish that statement. It may
refute a prediction about that attempt, weaken a hypothesis, or motivate
another probe.

Consequently, GoW should distinguish **hard exclusion justified by an
argument** from **a provisional preference against a region**. Otherwise the
"negative-space collapse" of [05](05-complement-geometry.md) simply conceals
an aggressive generalization policy.

Even legitimate accumulated exclusions need not produce a small, connected, or
realizable residual region. They can leave disconnected possibilities — or
expose inconsistent assumptions. Those outcomes are information too.

**Minimum closure for this gap:** one explicit intervention contract and one
scoped case showing either a checked intervention effect or a demonstrated
obstruction. General causal discovery over all Work is a later ambition.

---

## 4. Convergence: prevent retrospective repair from masquerading as learning

### First decide what is supposed to converge

At least three possibilities exist:

**The representation stabilizes.** Successive maps stop changing
substantially.

**Predictions improve.** The maps make better forecasts on evidence not used
to construct them.

**Search improves.** The resulting policy reaches a fixed objective with
better outcomes or lower total cost.

These can diverge. A representation can stabilize around a false explanation.
A growing representation can keep improving predictions. Predictions can
improve without helping the choices that matter.

The proposal here makes **prospective decision performance** the principal
target, with prediction quality and representation complexity as supporting
diagnostics.

### The anti-mythology rule is temporal

> **A revision may use the last failure to learn, but cannot count explaining
> that failure as a successful prediction.**

For each episode, freeze the current representation, selected action,
predicted consequences, and interpretation rule before observing the result.
Score that commitment. Then allow revision.

The revised model must earn validation on subsequent evidence; its parent's
misses remain in the record.

This follows the logic of prequential evaluation (Dawid): assess sequential
forecasts against the observations that subsequently arrive, rather than
judging a final model only by how well it fits its construction data.

A history ledger alone is insufficient. The ledger must support a score that
cannot be improved merely by rewriting the representation after seeing the
answer.

### Reusing the archive is not free evidence

Suppose the system keeps proposing abstractions until one fits the same twelve
notes. Even a nominally held-out subset can become part of the training
process if its results repeatedly inform revisions.

Adaptive-data-analysis research (the Dwork et al. reusable-holdout line)
specifically addresses this problem: repeated adaptation to evaluation
feedback can overfit the holdout itself. The need for protected evaluation
applies to GoW's representation-selection loop as much as to its final
proposals.

For GoW, that suggests evaluating on fresh episodes or protected source
families, not only new phrasings of the same mechanisms. Competing
representations should also face shared probes or comparable independent
episodes. A model should not appear more accurate merely because its own
policy selects easier cases.

### A limited convergence statement is available now

Consider a deliberately bounded setting:

- There is a **fixed finite collection** of candidate models of Work.
- One candidate correctly predicts all observations in the scope.
- A trusted evaluator returns exact outcomes.
- Whenever candidate models disagree, an admissible distinguishing probe can
  be selected and executed.
- Inconsistent candidates are removed.

Then each distinguishing observation removes at least one incorrect candidate
while retaining the correct one. With \(K\) candidates, at most \(K-1\) such
eliminations are possible before a unique candidate remains — provided the
candidates are distinguishable by the available probes. Otherwise the result
is an indistinguishable class of models.

That is an elementary conditional argument (version-space elimination in
GoW clothing — see the [version-spaces summary](../research/summaries/04-version-spaces.md)),
**not a guarantee for current GoW**.

Its limits expose the real research work:

**The representation class is not fixed.** GoW can invent new dimensions and
hypotheses.

**The evaluator is not always decisive.** An unknown result may eliminate
nothing.

**A distinguishing probe may be expensive or unrealizable.** The elimination
bound says nothing about the cost of finding it.

**The true account may not be in the candidate set.** Eliminating disagreement
does not establish truth under misspecification.

Most importantly, identifying the best model of a bounded task is not the same
as solving the task.

### What can be guaranteed before general convergence?

A useful first target is **disciplined non-convergence**: the system can fail
to improve, but cannot represent that failure as accumulated confirmation.

That requires persistent prospective scores, explicit revision costs, retained
alternatives, and a stop-or-act decision when further modeling is not
justified. Uncertainty may increase after a good experiment; monotonic
confidence is not a sound progress requirement.

The success measure should therefore include the cost and consequences of the
decisions the map produces — not the number of refinements, accepted claims,
or newly named regions.

This aligns with
[conditions 3 and 5](00-paper-claims.md#falsification-conditions-for-gow-stage-8-material),
but keeps them distinct. Condition 3 concerns whether epistemic machinery
earns its overhead. Condition 5 asks whether the procedure works beyond
histories prepared by its own practitioner. Neither follows merely from an
internally consistent sequence of refinements.

**Minimum closure for this gap:** a prospective adaptive episode with a fixed
external objective, retained prediction errors, bounded total cost, and a
comparator. Broader convergence claims need stronger assumptions and further
results.

---

## What the theory actually needs next

The four gaps can be joined by one modest organizing requirement:

> **A GoW representation earns operational authority when it supports grounded
> transformations and improves prospective decisions under a declared
> objective, evidence regime, and resource budget.**

That yields a coherent research program rather than four unbounded fields to
solve.

| Gap | First defensible theoretical contribution | Necessary empirical counterpart |
|---|---|---|
| Representation | A decision-preservation criterion and conditional loss bound | Competing bases selecting next Work on unseen episodes |
| Geometry | A typed transformation graph with explicit realization obligations | A proposed path successfully grounded, or a spurious path detected |
| Causality | Separate contracts for prediction, intervention, and obstruction | A scoped intervention or checked impossibility argument |
| Convergence | A bounded elimination result plus prospective revision discipline | An adaptive loop that improves — or visibly fails to improve — against a comparator |

**Universal solutions to all four are not prerequisites for publishing. What
is required is stopping their unanswered parts from supplying hidden authority
to the rest of the theory.**

In particular:

A useful map need not be unique.
A rigorous geometry need not be metric.
A useful search policy need not establish causality for every feature.
A legitimate adaptive method need not guarantee that it solves every problem.

But it must make commitments that its own representational flexibility cannot
erase.

**The deepest missing piece is therefore not "more geometry." It is a contract
connecting representation changes to realizable actions and future evidence.
Once that contract is explicit, GoW becomes a testable search methodology
rather than an indefinitely extensible account of why past Work failed.**

## Relationship to the rest of the series

- Sharpens the [four missing rigorous pieces](06-situated-in-the-literature.md#four-missing-rigorous-pieces-the-research-program)
  of [06](06-situated-in-the-literature.md) into per-gap contracts with
  minimum-closure conditions.
- Reinterprets falsification conditions
  [FC1, FC2, FC4](00-paper-claims.md#falsification-conditions-for-gow-stage-8-material)
  operationally (FC1 as unresolvable consequential disagreement; FC2 as
  utility not causality; FC4 as the geometry–action connection, via path
  composition).
- Qualifies the field prose of [05-complement-geometry](05-complement-geometry.md):
  exclusion requires an argument with a quantifier, not accumulated
  disappointment.
- Extends the [abstraction-safety](../abstraction-safety.md)
  predictive-discrimination test to a decision-preservation criterion over
  \(\mathcal M(W)\).
- The deductive-obstruction route is exhibited end-to-end by the
  [teaching model](../thesis/why-map-the-work.md) and its
  [checked implementation](../thesis/examples/why-map-the-work/).
- The convergence discipline is the theoretical home of the anti-mythology
  concern pressed by [attack A2](../research/composite-attack-surface.md#a2).

