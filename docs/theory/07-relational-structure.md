# 07 — Relational Structure

*Theory-series extension admitting relations among aspects of Work into the
conceptual model. It contains one small proved proposition (role: `theory`),
one executable instrument control (role: `implementation`, shipped), and one
clearly labeled hypothesis (role: `hypothesis`) that only the comparative
experiment can promote. Per the [claim registry](00-paper-claims.md), nothing
here is an `observation`.*

---

## Why this document exists

The prediction worth testing is:

> **GoW can identify outcome-relevant relationships among aspects of prior
> Work, and use those relationships to choose better next attempts than a
> specified alternative.**

That sentence contains an established mathematical possibility (relations can
carry information that no individual feature carries) and an untested claim
about this method (that GoW finds the right relations and turns them into
better Work). This document proves the first, calibrates the instrument for
it, and pre-commits the second to an experiment.

## Proposition R (role: `theory`)

> **Summaries of individual feature–outcome relationships can discard all of
> the predictive information available from their joint arrangement.**

**Setting.** Two independent, uniformly distributed binary choices \(A\) and
\(B\), with success \(Y = 1\) exactly when they differ.

**Claim.**

\[
P(Y=1 \mid A=a) \;=\; P(Y=1 \mid B=b) \;=\; \tfrac12
\quad\text{for every } a, b,
\]

yet

\[
P(Y=1 \mid A=a, B=b) =
\begin{cases}
1, & a \ne b,\\
0, & a = b.
\end{cases}
\]

**Proof.** The four combinations \((a,b) \in \{0,1\}^2\) are equally likely.
By the success rule, \(Y=1\) on \((0,1)\) and \((1,0)\), and \(Y=0\) on
\((0,0)\) and \((1,1)\). Fixing \(A=0\): the two equally likely rows are
\((0,0)\) with \(Y=0\) and \((0,1)\) with \(Y=1\), so
\(P(Y=1 \mid A=0)=\tfrac12\); the other three conditionals are identical by
symmetry. The joint conditional reads off the same four rows directly. ∎

**In ordinary language:** either choice alone is uninformative; their
relationship determines the outcome completely.

**What the proposition is and is not.** It is an illustration of established
informational *synergy* — the distinction between unique, redundant, and
jointly available information is part of the existing
information-decomposition literature (partial information decomposition;
Williams & Beer's framework is the usual anchor — *unverified pointer*, not a
checked citation). It is **not** a new GoW theorem, and nothing about it is
quantum. GoW claims no credit for the possibility, only an obligation to
handle it.

## The correction this document commits to

A feature vector containing both \(A\) and \(B\) **has not lost the
relationship** — a capable learner can infer \(Y = [A \ne B]\) from jointly
recorded features. Information is lost when:

1. only **marginal summaries** are retained (per-feature success rates), or
2. the **model class is restricted** to explanations that cannot represent
   the interaction.

Existing methods already model interactions; *"GoW beats a method forbidden
to represent the answer" would be a weak result.* Any comparative claim must
therefore include an ordinary interaction-aware baseline — this constraint is
pre-committed in the [experiment protocol](#the-comparative-experiment).

## What the current instrument can already express

No new primitive is required. Over finite axes, any relation is expressible
as a boolean expansion of literals — XOR over binary factors is exactly:

```text
any( all(A=0, B=1), all(A=1, B=0) )
```

The shipped `invariant-predicate/v1` grammar (`all`/`any`/`not` over
`equals`/`in` on enum axes) can represent this form today; what is *not*
established is that the mining/challenge chain ever proposes or selects such
a form. Representation capacity and discovery behavior are different claims —
conflating them is precisely the error the correction above forbids.

Per the recommendation this document implements: **no "entanglement"
primitive and no new orchestration layer.**

## The executable control (role: `implementation`, shipped)

`internal/relational` is the deterministic calibration instrument. It owns:

- two frozen complete factorial tasks: `control-xor` (success iff `A ≠ B`)
  and `control-main-effect` (success iff `A = 1`), as outcome *tables*, not
  rule functions;
- a tiny executable relation language (`relational-claim/v1`) in which a
  proposal must state something checkable — `success requires A ≠ B` as a
  validated JSON document, never "A and B are coupled";
- marginal analysis (`Marginals`, `FactorInformative`) so the information
  loss of individual-feature summaries is *demonstrated*, not asserted;
- the deterministic checker (`Check`) that evaluates a proposed relation
  against the complete finite task;
- the full required chain (`RunChain`):

```text
observed attempts
 → proposed relation
 → predictions for unobserved combinations
 → a discriminating next attempt        (deterministic max-disagreement probe)
 → checked outcome                      (frozen table decides, never the model)
```

The package tests are the control runs:

- **Proposition R numerically** — every marginal cell rate is exactly 0.5 on
  the full XOR table while `A ≠ B` predicts every row.
- **Interaction presence changes the conclusion** — `A ≠ B` is complete on
  `control-xor` and scores exactly chance (0.5) on `control-main-effect`;
  the marginal explanation `A = 1` inverts. A checker reaching the same
  conclusion on both controls would not be measuring the interaction.
- **The chain lands both ways** — on the XOR control the probe outcome
  supports the committed relational prediction and refutes the rival; on the
  main-effect control the *same chain refutes* the relational proposal and
  the marginal rival survives.
- **Degenerate honesty** — when no unobserved combination separates the
  candidates, or nothing is left to observe, the instrument reports that
  instead of inventing an informative probe.

**Scope statement.** This validates a **capability of the instrument**: the
chain is representable, executable, and deterministically checkable, and the
checker's verdict tracks the presence of the interaction. It does **not**
establish that GoW discovers useful relationships in difficult research.
Recognizing a familiar XOR table is a calibration, not the flagship
experiment.

## The hypothesis (role: `hypothesis`, untested)

> **H-R.** From a population of prior Work, the GoW loop (map → relation
> proposal → challenge → probe selection) can infer outcome-relevant
> relations and select next attempts that reach the objective at lower cost
> than (a) individual-feature analysis and (b) an ordinary interaction-aware
> method given the same evidence and budget.

Falsification paths:

- GoW's relation proposals fail the deterministic checker at rates
  indistinguishable from permuted-label proposals (relation discovery is
  noise);
- condition (b) matches GoW's held-out prediction and probe efficiency
  (value lives in established interaction modeling, and GoW's contribution
  reduces to representation/workflow — informative, but a different claim);
- GoW's costs (mapping, interpretation, challenge) exceed the attempts it
  saves (FC3: epistemic overhead without epistemic gain).

## The comparative experiment

Pre-registered as a draft protocol:
[`corpus/experiments/pilot-005-relational/PROTOCOL-DRAFT.md`](../../corpus/experiments/pilot-005-relational/PROTOCOL-DRAFT.md).
Design constraints fixed here so the protocol cannot drift:

| Condition | Purpose |
|---|---|
| C1 — individual-feature analysis | Tests what is lost by ignoring interactions |
| C2 — ordinary interaction-aware method | Tests whether established interaction modeling already suffices |
| C3 — GoW relation inference and probe selection | Tests the proposed contribution |

- All conditions receive the **same permitted evidence** and **comparable
  total resource budgets**; GoW's mapping, interpretation, and challenge
  costs count — not just its final proposal.
- Frozen executable tasks with hidden, known outcome rules: interaction-
  dependent tasks, tasks where individual effects suffice, and cases with
  insufficient evidence. Rules and evaluation cases stay outside proposing
  sessions.
- **Two measures, separately:** held-out prediction quality, and the
  number/cost of additional attempts needed to reach the objective.
  Predicting outcomes and selecting useful interventions are not the same
  achievement.

The interesting outcome is not "GoW noticed a relationship." It is:

> **GoW inferred a relationship from prior Work, committed to a prediction,
> and selected a next attempt whose checked outcome supports that
> prediction — at a useful cost relative to the alternatives.**

## Self-certification is excluded by construction

There is nothing circular about using GoW to organize its own experimental
history. Circularity would arise if GoW generated a relation, generated an
interpretation of the result, and treated agreement between those outputs as
validation. The evaluator stays outside that loop:

```text
GoW proposes the relation and probe.
A frozen domain checker evaluates the result.
The report preserves disagreements and unresolved cases.
```

This matches the existing challenge contract: an unconfirmed model challenge
must not silently strengthen or weaken a hypothesis
(`ModelJudgment ≠ Verification`; see the [epistemic model](02-epistemic-model.md)).

## Relationship to the rest of the series

- Proposition R sharpens why [01-semantic-model](01-semantic-model.md)'s
  shape axes cannot be consumed only marginally: `ShapeClaim` support
  computed feature-by-feature can be blind to exactly the structure that
  separates regimes.
- The discriminating probe is the [Probe reading](06-situated-in-the-literature.md)
  made deterministic in miniature (version-space/QBC disagreement; see the
  [active-learning summary](../research/summaries/05-active-learning.md)).
- The control discharges part of [attack A3's defense](../research/composite-attack-surface.md#a3)
  at calibration scale: the checker, not the proposing model, decides.
- H-R joins the labeled-hypothesis list in the
  [claim registry](00-paper-claims.md#what-goes-under-hypotheses--future-work).
