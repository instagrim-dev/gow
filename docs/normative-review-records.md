# Normative review records

This document describes the **normative review ledger** (`internal/review`,
`internal/store/review_store.go`, migration `v44`) and how coverage is generated
from it.

It exists because of a specific failure: a review run could reach cases C1–C8 of
`docs/reviews/prompts/recipes/assessment-admission-decision.md`, find no place to
put a policy, an applicability decision, a check attempt or an assessment, and
therefore could only have produced a hand-written `COVERAGE.md`. A
hand-maintained coverage table is indistinguishable from a derived one once it is
committed, so the gate stayed `UNDETERMINED` until the mapping existed.

## Normative is not scientific

These records are deliberately separate from the scientific lifecycle:

```text
obligation        REQUIRES a property
CandidateInvariant CLAIMS a regularity over a conditioned population
```

There is no promotion path between them. A mined predicate can be the *subject*
of an obligation's assessment; it can never be relabeled as one. The tables have
their own id kinds (`rpol_`, `robl_`, `rapp_`, `rdep_`, `rchk_`, `rasm_`) and
share no rows with invariants, challenges or evidence admissions.

## The four record responsibilities

| Responsibility | Table | What it pins |
| --- | --- | --- |
| Decision policy | `review_policies` | the authority under which a decision may be reached at all, its named decision, scope justification and budgets |
| Obligation | `review_obligations` | one individually versioned requirement plus its acceptance criteria and applicability rule |
| Applicability | `review_applicability_decisions` | whether an obligation applies to an exact subject, with rationale and authorizer |
| Assessment | `review_assessments` | one outcome for one obligation revision against an exact subject and context, with its argument, assessor, applicability decision, dependency manifest and check references |

Two supporting records carry the weight that makes the above checkable:

- `review_check_attempts` — what was actually run or read, by whom, where, with
  which procedure revision, and what came of it.
- `review_dependency_manifests` (+ `review_manifest_dependencies`) — what the
  assessment's meaning rests on, and **why each dependency is relevant**.

## Coverage is generated, never stored

There is no coverage-status column and no `review set-status` command. The only
reader is `Store.LoadReviewCoverage`, and the only writer of a document is
`review.RenderCoverage`. To change what coverage says you must record the
applicability decision, check attempt, or assessment that justifies it.

`RenderCoverage` emits no generation timestamp. Repeated generation from
identical records is byte-identical, so a regenerated file differs only when the
records or the declared current dependency values differ.

### Derived decisions

```text
WITHHOLD             a demonstrated, unresolved blocking nonconformance applies
UNDETERMINED         mandatory applicability, authority or evidence is unresolved
ELIGIBLE_TO_ADVANCE  scoped permission to advance under this policy
```

`ELIGIBLE_TO_ADVANCE` is permission, not a claim that a hypothesis is true.

Precedence is fixed: a demonstrated blocker outranks every uncertainty, but the
uncertainties are still listed. A blocker never hides an unexamined obligation.

### States that must not collapse

| State | Meaning |
| --- | --- |
| `unexamined` | no assessment exists — not a pass, not a failure |
| `inconclusive` | an assessment was made and reached no conclusion |
| `blocked` | a cited check could not execute |
| `not_applicable` | an authorized decision with a rationale says it does not apply |
| `conforms` | assessed as conforming, with completed **executed** check support |
| `nonconforms` | a demonstrated blocker when the obligation is mandatory |
| `applicability_unresolved` | no applicability decision, or conflicting ones |

Specific refusals encoded in the projection and the schema:

- an **inspected** procedure cannot be recorded as `completed` — inspection is
  not execution, so source reading cannot certify an unexecuted check;
- a **blocked** attempt cannot support a `conforms` assessment (refused by both
  the pipeline validator and a schema trigger);
- **absence** of an assessment reports `unexamined`; no empty assessment is
  created to claim coverage;
- **conflicting** applicability decisions stay unresolved rather than being
  settled by recency;
- a **demonstrated nonconformance** is not erased by a later favorable
  assessment; the disagreement is reported as a contradiction;
- a **vacuous** policy (no mandatory obligation) and an **unauthorized** policy
  (missing owner, authority source, or scope justification) cannot grant
  eligibility.

## Staleness is relevance, not difference

An assessment goes stale only when a dependency **kind it declared**, with a
stated reason, now has a different ref. This is the C4/C5 boundary:

- **C4** — the declared assessment population moves: the assessment is stale for
  a current request, and eligibility is not inherited. Reassessment is the only
  path back, so the gate is not a dead end.
- **C5** — an artifact outside the manifest changes: nothing goes stale. A
  moving repository `HEAD` or a bumped global document version is not evidence
  that a bounded assessment stopped applying.

A dependency kind the caller does not supply is **unknown**, not changed:
incomplete inputs must not invalidate valid assessments. A stale assessment keeps
its historical outcome and loses only current authority.

## Surfaces

```sh
newf review policy --key assessment-admission-decision --revision 1 \
  --decision-name <decision> --owner <who> --authority <source> --scope <why> \
  --obligation 'key=...;revision=1;requirement=...;acceptance=...;applicability=...;owner=...;mandatory=true'

newf review assess --policy rpol_... --obligation robl_... \
  --subject <exact-ref> --context <ref> --outcome conforms --argument <why> \
  --assessor <who> --applicability rapp_... --check rchk_... \
  --depends 'project_revision=<sha>=<why relevant>'

newf review applicability --policy rpol_... --obligation robl_... \
  --subject <exact-ref> --decision applies --rationale <why> --authorizer <who>

newf review check --policy rpol_... --obligation robl_... --case C1 \
  --procedure <what> --procedure-revision <rev> --executor <who> \
  --environment <where> --mode executed --outcome completed --output <ref>

newf review coverage --policy-key assessment-admission-decision \
  --current assessment_population=clr_... --out COVERAGE.md
```

## The integrated gate

`TestIntegrationCurrentAssessmentAuthorityObligation`
(`internal/pipeline/review_obligation_integration_test.go`) instantiates policy
P1 and the single obligation `current-assessment-authority@1` through the real
migrated store and exercises C1–C8 end to end: baseline, withheld model-only
failure, checked admission, relevant change, unrelated control, reassessment,
historical replay, and deterministic projection — including a control policy
proving `unexamined` and execution-blocked stay distinguishable.

Generate the evidence-bundle export from that run with:

```sh
NEWF_REVIEW_COVERAGE_OUT=docs/reviews/evidence/COVERAGE.md \
  go test ./internal/pipeline -run '^TestIntegrationCurrentAssessmentAuthorityObligation$' -count=1
```
