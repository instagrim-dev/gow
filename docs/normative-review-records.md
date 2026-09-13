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

There is no coverage-status column and no `review set-status` command. The
readers are `Store.LoadReviewCoverage` for a whole-policy historical export and
`Store.LoadReviewCoverageForSubject` for one exact subject; the only writer of a
document is `review.RenderCoverage`. To change what coverage says you must
record the applicability decision, check attempt, or assessment that justifies
it.

`RenderCoverage` emits a generated-at line as non-semantic metadata. Repeated
generation from identical records preserves every substantive byte, so a
regenerated file differs meaningfully only when the records or the declared
current dependency values differ.

A current decision about a subject must use the subject-scoped path
(`newf review coverage --subject <exact-ref>`). That projection includes only
applicability decisions and assessments for the named subject, so a legitimate
`does_not_apply` decision for B cannot make A unresolved, and B cannot inherit
A's assessment merely because both share a policy and obligation. Omitting
`--subject` is a historical whole-policy export, not a per-subject decision.

An assessment's cited applicability decision must match the assessment's policy,
obligation and subject. The store checks this before insert and the schema
trigger enforces it for direct SQL writes.

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
  --subject <exact-ref> --current assessment_population=clr_... --out COVERAGE.md
```

## Execute a typed finite claim

`review check-finite` takes an operator-supplied formal target, selects the
registered `finite_equivalence` checker, executes it and writes its actual
certificate to the existing check ledger. This entry point does not infer the
claim kind or translate source prose. `source_ref` records the operator's source
reference; correspondence between that source and the typed target remains an
explicit, separately assessed premise.

Input is one strict UTF-8 JSON object, at most 1 MiB:

```json
{
  "schema": "finite-claim/1",
  "kind": "finite_equivalence",
  "source_ref": "development:four-bit-add-zero",
  "statement": "x + 0 equals x over the declared four-bit domain",
  "domain": {"width": 4, "variables": ["x"]},
  "left": {"op": "add", "args": [{"var": "x"}, {"const": 0}]},
  "right": {"var": "x"}
}
```

An expression has exactly one form: `var`, unsigned `const`, or `op` with
`args`. Unary operators are `not`, `neg`, `shl1`, `shr1`; binary operators are
`and`, `or`, `xor`, `add`, `sub`, `mul`. Arithmetic uses words modulo `2^width`.
Width is 1–8, identifiers are plain names of at most 128 bytes, and each
expression is bounded to 4,096 nodes and depth 64 (root depth zero). Variables
must be declared without duplicates. An explicit empty variable array supports
constant expressions. Unknown, duplicate and case-variant JSON keys are refused.
There is no code evaluation, division, sampled fallback or provider call.

Define a policy and an obligation with acceptance criteria for the intended
finite question using `review policy`. Supply their returned IDs:

```sh
newf --db /tmp/finite-review.db --json review check-finite \
  --policy rpol_... --obligation robl_... --case add-zero \
  --executor local-operator --input /tmp/claim.json --max-assignments 16
newf --db /tmp/finite-review.db --json review check-show rchk_... \
  --policy rpol_...
```

All execution flags are required. `--max-assignments` is a reservation from
0 through 65,536, not permission to sample. If the full declared domain exceeds
it, no assignments are checked. The retained `UNRESOLVED` receipt identifies a
resource refusal; it does not label a fully specified domain as a missing
premise. Missing domain, width, variable array or either top-level expression
also produces `UNRESOLVED`, with the absent fields identified. Invalid supplied
premises produce `INAPPLICABLE`. Those verdicts become blocked check records;
equality and a concrete counterexample become completed records with their
distinct verdicts retained.

Successful storage returns exit zero and `persisted: true`, including when the
check's outcome is blocked. Operators must inspect `check.Outcome` and
`receipt.certificate.Verdict`; process success reports receipt delivery.
Interrupt/SIGTERM cancellation is cooperative and retains an unresolved check
under a separate five-second storage timeout. If storage fails after execution,
the command exits nonzero and returns the unpersisted receipt in `result` in
the JSON error envelope (or as JSON on stdout in text mode). Preserve that output
for recovery. An abrupt kill or power loss during execution can lose the current
attempt; this command does not implement incremental crash recovery.

Each receipt retains exact input bytes in `input_json`, their SHA-256, source
reference, selected tool and procedure revision, environment, certificate,
checked assignments and elapsed execution time. CPU and custody costs remain
unmeasured. `check-show` reads the stored row, including the receipt in
`OutputRef`; it does not rerun the checker or authenticate the record's author.
The original JSON can be submitted again for a new check attempt.

The returned `subject_ref` is `finite-claim:sha256:<exact-input-hash>`. Use it in
`review applicability` and `review assess`, cite the returned check ID and
declare `--depends 'candidate_content=<subject_ref>=exact input defines the target'`.
Supply `--subject <subject_ref>` and
`--current candidate_content=<subject_ref>` when generating current coverage.
Changing the exact input invalidates current authority under that dependency.
A direct `finite-claim:` assessment cannot cite this command's receipt for
different input bytes. Broader subjects still require an explicit relevance
argument and dependency manifest.

No assessment is automatic: the assessor compares the certificate with the
obligation's criteria. A counterexample can support `nonconforms` for an
equivalence requirement, while a blocked check cannot support `conforms`.
`ELIGIBLE_TO_ADVANCE` remains scoped permission under the recorded policy; this
path neither admits rewrite rules nor establishes a scientific invariant.
Other registered claim kinds have no execution surface in this command.

`TestCLIFiniteCheckRealLedgerAndScopedAssessment` exercises positive,
counterexample, invalid-premise, missing-premise and resource-refusal cases
through the real CLI, SQLite ledger and coverage projection. Separate tests
cover cancellation and a failed storage insert. These are engineering checks,
not the protected G1 usefulness evaluation.

## Execute typed finite instance evidence

`review check-finite-instance` evaluates the expressions at supplied points
only. It uses a separate `finite-instance-claim/1` schema and returns
`INSTANCE_EVIDENCE_ONLY` when every supplied point agrees. That verdict is not
an exhaustive finite-domain equivalence certificate, cannot warrant a rewrite
rule, and says nothing about unlisted assignments. A single admissible supplied
counterexample is different: it returns `REFUTED`, because one counterexample
does refute the declared universal equality over its finite domain.

```json
{
  "schema": "finite-instance-claim/1",
  "kind": "finite_instance",
  "source_ref": "development:two-four-bit-points",
  "statement": "x + 0 agrees with x at the supplied points",
  "domain": {"width": 4, "variables": ["x"]},
  "left": {"op": "add", "args": [{"var": "x"}, {"const": 0}]},
  "right": {"var": "x"},
  "assignments": [
    {"values": [{"var": "x", "value": 0}]},
    {"values": [{"var": "x", "value": 7}]}
  ]
}
```

Variable bindings are arrays, not JSON maps. Each assignment must bind every
declared variable exactly once; undeclared variables, duplicate names and
duplicate whole assignments are `INAPPLICABLE`. `value` is an unsigned integer,
so zero is explicit and never a missing value. A null or absent `assignments`
field is `UNRESOLVED`; an explicit empty array is also `UNRESOLVED`, because no
point was checked. Out-of-domain values are rejected by the finite checker.

```sh
newf --db /tmp/instance-review.db --json review check-finite-instance \
  --policy rpol_... --obligation robl_... --case two-points \
  --executor local-operator --input /tmp/points.json --max-instances 2
newf --db /tmp/instance-review.db --json review check-show rchk_... \
  --policy rpol_...
```

All flags are required. `--max-instances` reserves 0–4,096 supplied assignment
records. A reservation shortfall rejects the entire input without checking a
prefix. The 1 MiB strict-JSON limit, duplicate/case-variant key rejection,
expression limits, word widths and closed declared variable set match the finite
equivalence path. This command observes cancellation before and after the
bounded synchronous checker, then uses a separate five-second storage timeout
to retain a cooperative cancellation receipt. CPU and custody costs are
unmeasured; provider calls are zero.

The receipt's subject is `finite-instance-claim:sha256:<exact-input-hash>`.
Use that exact value for applicability, assessment and a
`candidate_content` dependency. A direct typed assessment cannot cite a check
for different bytes. Pass the same value to `review coverage --subject` when
deriving current coverage. `INSTANCE_EVIDENCE_ONLY` may support an explicitly scoped
obligation about these exact points, but it must not be described as evidence of
domain equality. Assessment, policy mutation and rule admission remain separate
operator actions. `finite.VerifyRuleWarrant` rejects this verdict by design.

`UNRESOLVED` and `INAPPLICABLE` records are blocked. Stored instance evidence
and counterexamples are completed checks; exit success reports receipt storage,
not a policy conclusion. If the database write fails after execution, the command
returns a nonzero JSON error carrying the unpersisted result under `result`.
`TestCLIFiniteInstanceCheckLedgerAndScopedAssessment` covers agreement,
counterexample, missing/invalid input, resource refusal, stale dependencies and
the domain-warrant rejection. These are engineering checks, not protected G1
usefulness evidence.

## Execute a typed observation claim

`review check-observations` routes `observed_rate_invariance` and
`solved_monotonicity` to the existing deterministic measurement tools. The
registered `probabilistic_property` route produces an explicit `NOT_ASSESSED`
refusal. No route interprets source prose, authenticates supplied executions or
independently verifies their success labels. Conclusions concern the submitted
records under the declared conditions.

The `observation-claim/1` schema accepts ordered traces, never supplied aggregate
counts. This example has a falling recorded rate, from 1/1 to 1/2, while retaining
the same solved instance:

```json
{
  "schema": "observation-claim/1",
  "kind": "observed_rate_invariance",
  "source_ref": "development:retained-prefix",
  "statement": "The observed success-per-submission rate is unchanged at budgets 1 and 2",
  "binding": {
    "population": "instance a", "ordering": "recorded order",
    "stopping_rule": "declared budget", "budget_min": 1, "budget_max": 2
  },
  "observations": [
    {
      "conditions": {"population": "instance a", "ordering": "recorded order", "stopping_rule": "declared budget", "budget": 1},
      "instances": [{"id": "a", "submissions": [{"move": "m1", "success": true}]}]
    },
    {
      "conditions": {"population": "instance a", "ordering": "recorded order", "stopping_rule": "declared budget", "budget": 2},
      "instances": [{"id": "a", "submissions": [{"move": "m1", "success": true}, {"move": "m2", "success": false}]}]
    }
  ]
}
```

The metric is fixed: **recorded successful submissions / recorded submissions**.
Counts and solved-instance totals are computed from traces. A false success
flag must be explicit; an absent or null flag is a missing premise, not a failed
submission. An explicit empty submissions array has zero counts; an undefined
rate never becomes zero or equality. Each observation needs at least one
instance, with unique nonempty IDs and explicit nonempty move labels.

All non-budget conditions must match the binding, not merely each other.
Instance ID sets must also match. Budgets must be nonnegative, strictly
increasing in input order and inside the explicit inclusive bound range.
Observed-rate comparison accepts 2–64 observations. Solved-set retention
accepts exactly two and requires unchanged per-instance trace prefixes, including
move labels and verdicts. A non-nested execution is `INAPPLICABLE`; it is not a
refutation of the conditional retention claim. Rate comparison can still compare
non-nested executions over the same instance set, but supplies no incremental
decomposition for them.

Input is one strict UTF-8 JSON object under 1 MiB. Unknown, duplicate and
case-variant keys are refused, including nested submission keys. Across the
whole input, execution is limited to 64 observations, 4,096 instance records
and 16,384 submission records. Every appearance counts, including repeated
prefixes. Source references and statements have a 4,096-byte limit each;
condition labels 256 bytes; instance IDs 128 bytes; move labels 1,024 bytes.

```sh
newf --db /tmp/observation-review.db --json review check-observations \
  --policy rpol_... --obligation robl_... --case retained-prefix \
  --executor local-operator --input /tmp/observations.json --max-submissions 3
newf --db /tmp/observation-review.db --json review check-show rchk_... \
  --policy rpol_...
```

All flags are required. `--max-submissions` reserves 0–16,384 input records;
it is not a count of CPU instructions or a changed interpretation of each
observation's declared budget. Exceeding a limit refuses the entire comparison
without sampling. The receipt records input submission count, the reservation,
whether the selected assessor was invoked, and elapsed binding/assessment time.
CPU and custody costs remain unmeasured; provider calls are zero. The probability
refusal needs only schema, kind, source reference and statement; any supplied
observations still pass through input resource admission.

`observation-claim-check/1` preserves the legacy `measure-checker/1` arithmetic
and records `UNRESOLVED`, `INAPPLICABLE` and `NOT_ASSESSED` as **blocked** checks.
The older measurement adapter and historical rows retain their existing
semantics. A completed result is either `REFUTED` or
`HOLDS_AT_COMPARED_POINTS`; neither asserts a property of untested budgets or
underlying probabilities. For the example above, rate invariance is refuted,
while a separately bound solved-set claim holds for the verified pair.

Persistence, cold inspection and explicit assessment follow the finite path.
Use the returned `observation-claim:sha256:<exact-input-hash>` subject reference
for applicability, assessment and the `candidate_content` dependency. Direct
typed-claim assessments reject receipts for different input bytes. Saved
refusals return exit zero with `persisted: true`; a failed insert returns the
unpersisted result with a nonzero exit. Pass the same value to
`review coverage --subject` for current coverage. Inspect the certificate and
check outcome.

Cancellation is checked before and after the bounded synchronous assessor,
not within every trace traversal. A canceled attempt retains an unresolved
receipt using a separate five-second storage timeout. An abrupt kill or power
loss can lose an in-flight attempt. No automatic assessment, source attestation,
policy mutation or scientific promotion occurs.

`TestCLIObservationCheckScopedLedger` verifies the actual CLI-to-SQLite-to-review
path, including exact rates, retained solved instances, binding mismatches,
missing premises, changed populations, non-nesting and refused conformity.
Cancellation and failed-insert tests verify the retained-result paths. These
engineering cases do not complete the protected G1 usefulness gate.

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
