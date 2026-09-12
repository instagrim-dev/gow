---
id: review-preservation
revision: 1
kind: qualitative_comparison_protocol
status: ceilings_declared_twelve_runs_authorized
preparation: ../../preservation-pilot/PREPARATION.md
covers: [study-design, comparative-validation, decisions, reporting]
contract: ../review-contract.md
prerequisite: assessment-admission-decision
---

# Preservation before generalization

This is a qualitative falsification pilot for prompt packaging, not a powered
noninferiority study, novel-defect discovery result, or permission to expand the
prompt library. Dispatch only after the integrated obligation slice succeeds
and a named owner freezes the decision policy and resource authorization.

## Operator instruction

```text
Read the shared review contract and this protocol. Confirm a retained passing
assessment for the integrated obligation slice. Prepare the six case bundles,
equivalent-information arms and rubric below. Freeze them, access permissions,
model/tool settings, budgets and reuse horizon before any dispatch. Do not run
paid providers without explicit authorization. Keep runs isolated; adjudicate
blindly and retain every output. Report qualitative preservation, control
errors and economics separately. Do not call this a discovery experiment.
```

## Cases, arms and freeze

Use S3 (first-verdict/current-view confusion), S4 (historical revisions in the
current population), and one manuscript-fidelity defect whose defective and
corrected revisions have been independently verified. Pair each with its matched
corrected control. Pin exact relevant file sets, revisions, expected invariant,
reproduction oracle, and why the correction resolves that defect. Do not label
an entire corrected repository defect-free or count unrelated real findings as
false positives. Historical findings are case-selection leads, not current truth.

The original structural and manuscript reviews under `docs/reviews/` are case
sources, not prompts. Recover the actual manuscript-review instruction when
available, or author and freeze an explicit manuscript baseline before testing.
Neither existing engineering prompt is a manuscript comparator. Validate S3/S4
pre-fix and corrected behavior; do not guess a common pre-fix commit for all cases.
Exclude remedial commit messages, answer-containing reports and inconsistent
source access unless both arms explicitly receive them as substantive input.

| Arm | Information and procedure |
| --- | --- |
| Baseline | Bespoke, project-specific prompt with the relevant substantive cases. |
| Modular | Shared contract plus pinned project-specific case bundle containing the same substantive obligations, examples, symbols and evidence access. |

The independent variable is packaging and retrieval/composition overhead, not
extra hints or weaker evidence. Keep the same model/version, tools, settings,
access and token-accounting policy. Charge modular retrieval and contract tokens
to that arm. Freeze reviewer/judge instructions and handle identification of arms
so the adjudicator does not know which packaging produced each report.

## Run and resource budget

The initial pilot is six isolated cases (three defective, three corrected) by
two arms, one run per cell: **twelve runs**. Do not place paired defective and
corrected cases in one context. Randomize dispatch order, retain aborted runs,
and use fresh contexts without prior outputs or the decision discussion. Shared
model provenance remains shared provenance, not independent scientific evidence.

No repetitions or paid calls are authorized by this document. Before dispatch,
the owner must declare total model-call/token/currency ceilings, per-run limits,
human adjudication budget, environment controls and the realistic future reuse
horizon H. Missing ceilings permit preparation only. All preparation and run
costs must be recorded; use no retry-until-success policy.

Three repetitions per cell would require **thirty-six runs**, under a separately
authorized revision. That count alone does not establish statistical power.
The twelve-run pilot estimates neither a low false-assertion rate nor a moderate
cost-regression bound. Do not attach a numeric noninferiority claim to it.

## Rubric and gate

A recovered defect must name the actionable location, triggering conditions,
violated contract, downstream consequence and a discriminating regression check.
A generic warning receives no actionable-recovery credit. Equivalent substantive
information makes recovery a smoke test of preservation, not discovery evidence.

The discriminating observations are:

- whether an arm falsely reasserts the specified defect on its corrected control;
- whether it preserves actionable specificity and honest unresolved cases;
- whether retrieval, execution and reconciliation cost justify modular packaging.

A missing essential recovery or a false reassertion rejects the candidate
packaging for this initial trial. Retain all results; revising the packaging
starts a new protocol revision, not a replacement successful run. When neither
failure occurs, record feasibility on these six cases and uncertainty. A pass
does not establish generality, an error-rate advantage or cost noninferiority.
No new domain-prompt series follows automatically.

## Costs and break-even decision

Report human effort and compute separately. Convert them into one cost unit only
using rates fixed before results. Separate incremental one-time adoption setup,
recurring per-review cost (including amortizable maintenance assumptions), and
the one-time cost of running this comparison. Do not hide adapter authoring in
runtime or mix it into recurring cost. Record sunk baseline investment separately;
compare future avoidable costs for the adoption decision.

For modular M and baseline B, under explicit constant recurring-cost assumptions:

```text
C_M(H) = S_M + H * r_M
C_B(H) = S_B + H * r_B
```

S is future one-time setup cost, r is recurring cost per comparable review, and
H is the number of future reviews. In words: setup plus the accumulated review
cost. When extra modular setup is positive and recurring savings are positive:

```text
H_star = ceil((S_M - S_B) / (r_B - r_M))
```

This is the first nonnegative break-even horizon under those assumptions. With
positive extra setup and no recurring savings, there is no cost-based break-even.
Use the direct cost inequality for other sign combinations. Report a sensitivity
range for uncertain effort, maintenance and task mix; pilot estimates are not
precise forecasts. Quality or safety improvements may separately justify cost,
but require an explicit decision rationale, not a claim of savings.

The decision compares H_star with the owner's predeclared plausible H. An unknown
horizon or unsupported cost model leaves the economic assessment inconclusive.
Retain the qualitative quality gate separately. The cost of the comparison is
reported even if excluded as a sunk cost from subsequent adoption arithmetic.

## Later transfer experiment

Only after preservation passes, freeze prompts before exposing genuinely held-out
fault variants and clean controls. That separately authorized experiment tests
transfer beyond answer-bearing examples. Do not reuse preservation cases as
unseen discovery evidence or use this pilot's output to tune a supposedly frozen
transfer test. Report both shared dependencies and contamination limits.
