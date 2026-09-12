---
id: claim-form-obligations
revision: 1
kind: review_reference
status: proposed_templates_not_validated
origin: operator audit of preservation-pilot revision 1 (2026-09-12)
---

# Claim forms and their proof obligations

A reusable index for reviewers: when a manuscript or record attaches a
**property label** to a quantity, the label is a proposition with its own
proof obligation, decidable independently of whether the quantity's value
is correct. This index maps claim forms to obligations and cheap
discriminating checks. It exists because a real double miss showed that
reproducing a number is routinely mistaken for verifying the property
attributed to it.

These are **proposed reusable templates, not validated results**. Each use
must bind scope and preconditions to the actual subject. A knowledge base
should index claim forms and their obligations, not accumulate reminders
about individual past mistakes.

## The pattern

The established testing pattern is **metamorphic testing**: compare related
executions using a required relationship between their inputs and outputs.
The subject's claim supplies the required relationship; varying the named
condition supplies the transformation. The selection chain is:

```text
Established knowledge identifies a possible failure mechanism.
The claim determines the required property.
The subject's procedure determines which conditions can affect it.
An exact counterexample decides this particular claim.
```

The reference helps select the test. The counterexample establishes the
defect. For deterministic subjects, one admissible counterexample refutes
an invariance claim — no powered statistical experiment is required.

## The index

| Extracted claim form | Mathematical obligation | Cheap discriminating check |
|---|---|---|
| "X-independent" (budget-, order-, seed-…) | Invariance of the reported quantity over the specified (or, if unspecified, unrestricted) range of X | Hold everything else fixed; change X; compare |
| "Order-independent" | Invariance under the permitted permutations | Reorder inputs or operations |
| "Idempotent" | Repetition has the same relevant effect as one application | Apply twice; compare |
| "Monotonic" | A specified ordering is preserved | Search for one order-reversing pair |
| "Unconditional" / unqualified rate | The quantity does not depend on the procedure that selected its observations | Recompute under a changed selection procedure |

## A self-contained reference for rate claims

For a rate \(S/N\) (successes over submissions), adding a batch of \(s\)
successes among \(n>0\) further submissions changes the rate by

```text
(S+s)/(N+n) − S/N = (N·s − S·n) / (N·(N+n))
```

so the rate is unchanged **iff** `s/n = S/N`: added observations preserve a
rate only when they arrive at exactly the existing success proportion. Any
procedure change that admits observations at a different proportion changes
the rate — this is the whole content of many "X-independence" checks for
reported rates, and it is elementary algebra, not domain knowledge.

Two qualifications keep uses precise:

1. **Procedure-dependent selection does not automatically prove a rate
   changes** — numerator and denominator could change proportionally. The
   identity tells you what to compute; the computation decides.
2. **Restricted-range invariance is a different claim from unrestricted
   independence.** A plateau over part of the range does not rescue an
   unqualified label; conversely, refuting the unrestricted claim does not
   refute an explicitly range-scoped one.

## Worked example (the miss that motivated this index)

A manuscript labeled a local hit rate "the budget-independent
measurement." The required output was four lines:

> **Claim:** local hit rate is budget-independent.
> **Dependency:** budget changes which local submissions enter the
> denominator (stop-on-hit truncation).
> **Check:** recompute at budgets two and three, all else fixed.
> **Result:** 3/40 ≠ 3/57; unrestricted invariance is false.

Both reviewers of the defective revision reproduced 3/57 exactly and
affirmed the surrounding prose as adequately hedged. Reproducing the value
discharged nothing about the label. (Restricted qualification, recorded
for fairness: under the original ordering the local tally plateaus at
budgets ≥ 3; that supports a range-scoped claim only, which is not what
the defective sentence asserted.)

## Literature note (risk-flagging, not case-deciding)

Shin, Ramdas, Rinaldo, *On the Bias, Risk, and Consistency of Sample Means
in Multi-armed Bandits* ([arXiv:1902.00746](https://arxiv.org/abs/1902.00746))
distinguishes effects of adaptive sampling, stopping, choosing, and
rewinding on reported averages: properties of an average depend on how
observations enter it, not merely on whether the division was performed
correctly. Do **not** cite such results as direct proof that a particular
deterministic claim is false, and do not call a reported rate "biased"
without specifying the population quantity it is supposed to estimate. The
literature selects the failure mechanism to probe; the exact recomputation
decides the case.
