# Pilot 005 — relational structure (PROTOCOL DRAFT — not frozen, not executed)

Status: **draft for operator review.** No database, no captures, no pins.
Design constraints inherited from
[`docs/theory/07-relational-structure.md`](../../../docs/theory/07-relational-structure.md)
(Proposition R, the correction, hypothesis H-R) and the shipped calibration
instrument `internal/relational`. The instrument control (XOR/main-effect
chain, deterministic checker) is **prerequisite calibration, not part of this
pilot's evidence**.

## Claim under test

> **H-R.** From a population of prior Work, the GoW loop can infer
> outcome-relevant relations among aspects of that Work and select next
> attempts that reach the objective at lower cost than (a) individual-feature
> analysis and (b) an ordinary interaction-aware method, given the same
> permitted evidence and comparable total budgets.

The interesting outcome is not "GoW noticed a relationship." It is: GoW
inferred a relation from prior Work, **committed to a prediction**, and
selected a next attempt whose **checked outcome supports that prediction** —
at a useful cost relative to the alternatives.

## Conditions

| Condition | Method | Purpose |
|---|---|---|
| **C1** | Individual-feature analysis: per-factor marginal success summaries only; next attempt chosen by best marginal cell | Tests what is lost by ignoring interactions |
| **C2** | Ordinary interaction-aware method: a standard model class that can represent interactions (e.g. decision tree / logistic regression with interaction terms / full factorial ANOVA-style selection), fit to the same observations | Tests whether established interaction modeling already suffices |
| **C3** | GoW relation inference and probe selection: session proposes executable `relational-claim/v1` documents from the observed Work, challenge pass, discriminating-probe selection | Tests the proposed contribution |

Constraints fixed by the theory doc (non-negotiable at freeze):

- All conditions receive the **same permitted evidence** (identical observed
  attempts, identical task metadata) and **comparable total resource
  budgets**. C3's mapping, interpretation, and challenge costs **count** —
  not just its final proposal. Budget unit and exchange rate (attempts vs.
  session tokens vs. wall time) to be fixed at freeze and recorded.
- C2 must NOT be forbidden to represent the answer. A baseline that cannot
  express the interaction would make the comparison vacuous (the correction
  in 07: "GoW beats a method forbidden to represent the answer" is a weak
  result).

## Task battery (frozen before any capture)

Executable finite tasks with hidden, known outcome rules, generated and
frozen as outcome **tables** (the `internal/relational.Task` discipline —
rule functions are consulted once and discarded). Three strata:

1. **Interaction-dependent** — marginals uninformative by construction
   (XOR-family over 2–4 factors, mixed arity; at least one 3-factor parity
   case where pairwise relations are also uninformative).
2. **Main-effect-sufficient** — individual effects fully determine outcome;
   the relational condition must NOT hallucinate interactions here (false-
   positive pressure).
3. **Insufficient evidence** — observation budgets too small to distinguish
   surviving explanations; the correct output in every condition is an
   explicit underdetermination report, not a confident relation
   (absence ≠ negation discipline).

Task rules and full tables stay **outside** every proposing session. Sessions
see only: factor names/domains and the observed attempts permitted by the
evidence schedule. Factor names are semantically bleached (`F1`, `F2`, …) so
no condition can pattern-match a famous rule from the name.

## Measures (two, kept separate)

1. **Held-out prediction quality** — accuracy of each condition's committed
   predictions on unobserved combinations, scored by the frozen table.
2. **Intervention efficiency** — number/cost of additional attempts needed
   to reach the objective (stated per task at freeze: e.g. "produce a
   relation/model that predicts the complete table" or "reach a success
   outcome"), including each condition's full budget spend.

Predicting outcomes and selecting useful interventions are not the same
achievement; a condition may win one and lose the other, and the report
records them separately.

## Controls

- **Permuted-label control for C3** — the same GoW sessions over
  observations whose outcomes are permuted (fixed seed recorded at freeze).
  Relation-discovery rates on permuted data bound the noise floor.
- **Calibration gate** — the `internal/relational` control chain must pass
  in CI at freeze time (it validates the instrument, contributes no
  evidence).

## Evaluator placement (anti-circularity)

```text
C3 proposes the relation and the probe.
The frozen deterministic checker (internal/relational) evaluates predictions
and probe outcomes against the frozen tables.
The report preserves disagreements and unresolved cases verbatim.
```

No GoW session interprets its own checked outcome into the score. Matches
the challenge contract: an unconfirmed model challenge must not silently
strengthen or weaken a hypothesis.

## Predeclared readings

- **C3 > C2 ≥ C1 on interaction stratum, parity elsewhere, within budget** —
  supports H-R as stated.
- **C2 ≈ C3** — locates the value in representation or workflow rather than
  a distinct search advantage. Informative; H-R not supported as a search
  claim; record as such (do not reframe post hoc).
- **C3 relation proposals ≈ permuted control** — relation discovery is
  noise; H-R falsified at this scale.
- **C3 confident relations on stratum 3 or hallucinated interactions on
  stratum 2** — epistemic-discipline failure, reported regardless of other
  results.
- **C3 wins on prediction but not efficiency (or vice versa)** — report both;
  no aggregation into a single scalar.

## Open items before freeze

- [ ] Budget unit and cross-condition exchange rate (operator decision).
- [ ] Concrete C2 implementation and its hyperparameter policy (must be
      fixed before any C3 session runs).
- [ ] Observation schedules per task (factorial coverage vs. sparse draws;
      the factorial-design references in `docs/research/summaries/` inform
      this).
- [ ] Wire format + strict preflight for C3 proposal capture through the
      real decode path (`relational.ParseRelation`) before sealing.
- [ ] Number of runs per condition and per stratum; seed registry.
- [ ] Freeze artifact with digests and attestations recorded before
      dispatch (inherits pilot-004's pre-capture freeze correction).
