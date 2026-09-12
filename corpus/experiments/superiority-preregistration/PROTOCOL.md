# Superiority preregistration — resource-bounded verified-success advantage

**Status: `configured_not_executed`.** This is a *preregistration*, not a
result. No runs have been executed. It fixes the design so a later confirmatory
run cannot silently change endpoints, budgets, comparator, or analysis.

It is the empirical premise for the conditional theorem in the manuscript
(§9, *Conditional Superiority: When Shape-Guided Concentration Wins*). The
theorem states *when* a shape-guided method outperforms a comparator under a
fixed budget; this protocol specifies the experiment that would supply the
missing empirical fact. Executing it is future work; nothing here establishes
any advantage.

Machine-readable design: [`manifest.json`](manifest.json)
(`schema: newf-superiority-preregistration/v1`).

## 1. Claim under test

For a randomly selected task and starting history and a total resource budget
`B`, let `Y_pi(B) = 1` iff method `pi` achieves the verified objective within
`B`, else `0`. The estimand is

```
Delta(B) = E[Y_G(B) - Y_C(B)]
```

The **confirmatory claim** is `Delta(B) > delta` for a declared practical margin
`delta = 0.05` (five percentage points) over a **declared task population**,
with implementation versions, resource accounting, objective, budget `B`,
comparator, margin, and analysis all frozen before confirmatory testing.

Two boundaries (from §9.1):

- *"Conventional search" is a specified implementation, not a category.*
  No-free-lunch results preclude unrestricted superiority under all-objective
  averaging; they do not prevent advantages on structured populations. The
  question is which structure GoW exploits and whether the tested population
  contains it.
- *The advantage must be resource-bounded.* If GoW only reorganizes evidence
  the baseline already receives, an unbounded baseline could reproduce it. The
  claim lives in metareasoning: computation earns value through better external
  decisions per unit cost.

## 2. Arms

| Arm | Role | Description |
|---|---|---|
| `conventional_adaptive` | primary comparator | Same model, raw history, tools, checker, total budget. Ordinary generation, feedback, reflection, revision. **Not** prohibited from noticing interactions, drawing analogies, or reasoning about failures. |
| `gow_guided` | primary treatment | Representation selection, shape hypotheses, adversarial challenge, structural-intervention selection, projection, feedback-driven update. |
| `curation_only` | diagnostic ablation | Comparable organization/retrieval assistance **without** the shape-to-action policy. Isolates whether any gain is search control vs. convenient history. |

The primary result is `gow_guided` vs. `conventional_adaptive`. The ablation asks
whether the gain comes from **search control** rather than merely presenting the
history more conveniently. Do not cripple the comparator: a win whose rules
forbid the opponent from reasoning properly is not evidence.

## 3. Population and episodes

- **Domain:** a bounded domain with executable assessment — e.g. program-repair
  or small synthesis tasks with an exact checker over a declared finite
  specification. Erdős–Straus remains a motivating/exploratory case, but
  "resembles a research direction" is not interchangeable with checked task
  completion.
- **Episode:** one task plus a prior-work history with varied outcomes.
- **Holdout:** hold out task *structures* / generation regimes, not merely
  renamed identifiers. Include cases where the history contains transferable
  structure **and** cases where mapping should not help.
- **Independence unit:** the task/family is the higher-level unit. Stochastic
  reruns of one task are averaged or modeled *within* task and are **not**
  counted as independent episodes. 776 reruns of one task are not 776
  independent task families.

## 4. Budget accounting

Charge GoW's full cost: normalization, mapping, hypothesis generation,
adversarial challenge, projection, unsuccessful attempts, and verification.

Human intervention is either equalized across arms, prohibited during
evaluation, or explicitly counted and described — a human who supplies the
decisive abstraction is part of the method being evaluated. Report cold-start
and amortized costs separately under a declared deployment scenario; shared
preprocessing is not a free gift to GoW.

## 5. Outcomes

Primary `Y = 0` means **"no verified objective within budget"** — *not*
"mathematically false" or "impossible". Keep invalid output, incomplete
verification, timeout, unresolved evidence, and ordinary unsuccessful search as
separate secondary dispositions; never silently drop them from the denominator.
The endpoint measures verified achievement **under the declared evaluator**, not
an omniscient success rate.

Secondary efficiency measure: budget-capped time/cost to verified success
`E[min(T_pi, B)]` with `T_pi = infinity` on failure — so failures consume the
full horizon and the measure does not reward solving a few easy cases fast while
failing everywhere else.

## 6. Statistical analysis

Paired difference `D_i = Y_{G,i} - Y_{C,i} in {-1,0,1}`, `Delta_hat = mean(D_i)`.
Methods may adapt arbitrarily *within* an episode; the argument requires only
that episode **pairs** be independent under the frozen design.

- **Primary certificate (§9.3):** one-sided Hoeffding lower bound
  `L_alpha = Delta_hat - sqrt(2 log(1/alpha) / n)` at `alpha = 0.05`; success iff
  `L_alpha > delta`. This bounds the probability of *overstating* the advantage;
  it is not a posterior probability that GoW is better.
- **Secondary null test:** exact paired-discordance (McNemar) test on the
  discordant pairs against the equal-success null. The `p`-value addresses
  equality, not the five-point margin — report the effect estimate and bound
  beside it.

### Analysis discipline

- **Fixed sample:** sample size fixed before confirmatory runs. The fixed-`n`
  bound assumes `n` was **not** chosen by repeated inspection.
- **Sequential option:** if continuous monitoring is used, a prespecified
  confidence sequence / anytime-valid method replaces the fixed-`n` bound.
- **Selection vs. confirmation:** development runs for selection; **fresh**
  confirmatory episodes for the claim.
- **Multiple comparison:** comparing multiple baselines requires an explicit
  multiple-comparison plan or simultaneous bounds.

### Planning power (stipulated effects, not forecasts)

Exact one-sided paired-discordance power at `alpha = 0.05`, 80% target. Tests
superiority over zero, not the five-point margin.

| Pr(GoW-only) | Pr(baseline-only) | True advantage | Smallest `n` at 80% power |
|---:|---:|---:|---:|
| 30.0% | 10.0% | 20 pp | 67 |
| 20.0% | 10.0% | 10 pp | 199 |
| 17.5% | 12.5% | 5 pp | 776 |

## 7. Adaptive-loop (prospective) test

A batch result can show shape-conditioned proposals are better. The larger GoW
claim is that **map → probe → update improves subsequent search**. Each episode
preserves the sequence:

```
history
 -> chosen representation + shape hypothesis
 -> predicted consequence of a selected intervention
 -> actual domain candidate
 -> checked result
 -> updated next action
```

The update policy is frozen for the confirmatory study; it may revise
representations within an episode but cannot retroactively alter what an earlier
prediction meant. This separates **better starting guidance** (initial map
improves first proposals) from **better adaptation** (feedback yields better
subsequent attempts). Only the second substantiates the recursive-cartography
thesis.

## 8. Making the evidence specifically about GoW

A significant result establishes a difference between implementations, not the
*explanation*. The GoW account is supported to the extent these line up:

- **Action relevance** — changing/removing the inferred shape changes the next
  attempted work.
- **Prospective enrichment** — shape-guided proposals succeed more on outcomes
  not used to choose the shape.
- **Preservation** — the gain is not obtained by changing the goal or dropping
  hard cases.
- **Net efficiency** — the gain survives charging map + projection cost.
- **Adaptive value** — revised maps improve *later* decisions.

On a bounded domain, estimate `r`, `rho`, and the cost of using region `A`
directly and compare the observed advantage against the manuscript's
Proposition 9.2 prediction. This yields an interpretable result even when GoW
loses (e.g. "the map concentrated success, but projection overhead erased the
benefit"). A win that vanishes against `curation_only` locates the contribution
in evidence organization rather than the search policy. The provenance required
for attribution — provider/configuration metadata and raw request/response
records — is already produced by `newf`.

## 9. Explicitly not claimed

- Superiority over all search methods.
- Superiority on arbitrary future domains.
- That model-endorsement-rate differences equal a verified-search advantage.
- That the existing pilots (pilot-003, pilot-004) can be retrofitted into the
  paired tables above.
- Any result prior to execution of this design.
