---
id: assessment-admission-decision
revision: 1
kind: integration_review_recipe
status: executable_instructions_combined_test_not_implemented_here
covers: [assessment-identity, evidence-admission, decisions]
peers: [warrants, verification, governance, implementation]
contract: ../review-contract.md
---

# One obligation: assessment, admission, decision and derived coverage

Execute one integrated scenario and retain one evidence bundle. Do not build a
new assurance platform, run three domain sessions, or manufacture `COVERAGE.md`
first. This recipe specifies acceptance tests; its presence is not their execution.

## Copy-paste instruction

```text
Read AGENTS.md, docs/reviews/prompts/review-contract.md, TAXONOMY.md in that
folder, and this recipe. Pin the current checkout and decision policy. Trace
and exercise cases C1-C8 using the real migrated store and disposable fixtures.
Reuse existing component tests, but do not call them the combined scenario.
Record actual checks, blockers and findings in one evidence bundle. Generate
coverage from authoritative records only after that path exists. If the
required mapping or integrated test is absent, report the exact gap and the
smallest implementation slice; do not invent an assessment or parallel ledger.
Stay review-only unless implementation is separately authorized.
```

## The single obligation

`current-assessment-authority`, semantic revision 1:

> A current decision uses assessments compatible with its declared evidence
> population and policy. Historical replay remains reproducible without
> restoring obsolete current authority.

This is a norm, not a `CandidateInvariant`. Existing scientific claims can be
the subjects of the test; they must not be relabeled as review obligations.
Bind its definition, applicability, checks and assessment to a pinned decision
policy. The concrete named decision is whether the selected candidate may guide
the next search action under that policy, not whether a conjecture is true.

## Inputs, budget and source anchors

Require an authorized decision owner, evidence/admission/selection policy,
current project revision, exact fixtures and checker provenance, and the four
record mappings from the shared contract. Undefined mappings are findings.
Any required new normative type must be justified by this one missing behavior.

Budget: one scenario, cases C1-C8, two existing component-test invocations, one
run of repository gates, zero paid-provider calls, zero production DB/corpus
writes. At most one additional attempt per failed command for a diagnosed
harness/environment error; retain both attempts. Assertion failures are results,
not permission to retry. Resource overruns need a new authorization. Measure
human and compute costs separately. Missing execution access stops execution,
not source inspection. No independent model-review experiment is bundled here.

Authoring inspection anchor: `269de0d335804ceed955925df3ee5d3d02fd608b`.
Read current equivalents of:

- `internal/pipeline/challenge_population_integration_test.go`, especially
  `TestIntegrationChallengeAssessmentPopulation`;
- `internal/pipeline/admission_integration_test.go`, especially
  `TestIntegrationEvidenceAdmissionClosesReentry`;
- `internal/store/assessment_views.go`, the current population/admission writers
  and consumers, and their migrations;
- `docs/reviews/2026-09-12-structural-semantic-epistemic-review.md`,
  `docs/projection.md`, and relevant `internal/pipeline` projection tests.

At the anchor, the review record reports S1-S5 remediations, but records a
remaining limitation: discovery replay can re-earn survival against an old
population. Treat these as attributed source statements, not fresh execution.
The S1 test does not by itself assert the next current policy decision after
replay. The S2 positive path is operator-attested model judgment; it is not an
independently checked domain observation. Preserve these distinctions when
reusing the tests. Inspect current behavior rather than repeating stale findings.

## Procedure and acceptance cases

First freeze policy P1 and the obligation revision. Identify discovery population
D0, assessment A0, candidate content C, and the exact initial authority selection.
Use explicit store identities, not labels invented after the run.

| Case | Operation | Required observable result |
| --- | --- | --- |
| C1: baseline | Assess C under P1/D0 and derive a decision | Exact assessment, check, manifest and policy IDs explain the selected action. No global artifact-result shortcut. |
| C2: withheld control | Record a generated/model-only failure without sufficient admission basis | It remains auditable but does not enter the policy's required observed population or change that decision's evidence basis. |
| C3: checked admission | Independently check a scope-matched candidate-specific observation, retain checker inputs/outputs, then run the actual admission path | Only the admitted exact content enters the next population A1; the observation retains its subject, strength, scope and lineage. |
| C4: relevant change | Make A1 or a semantically relevant policy revision current while C remains unchanged | The affected assessment becomes stale for this current request. Until reassessed, eligibility is not inherited from A0. |
| C5: unrelated control | Change an artifact outside the declared relevant dependency graph | A still-compatible assessment does not become stale merely because repository HEAD or a global document version changed. |
| C6: reassessment | Assess unchanged C against A1 under the pinned applicable policy | A distinct assessment drives the next decision; D0 and A0 remain reproducible. Newly out-of-scope evidence cannot rewrite a bounded historical claim. |
| C7: historical replay | Replay A0 after the newer assessment completes, including an equal-clock/order stress check | Historical result remains available; the next current decision still selects the compatible current context, not the last execution or obsolete survival. |
| C8: derived projection | Generate coverage twice from the same authoritative records; then regenerate after the relevant update | Substantive output is deterministic; only affected current assessments change. Unknown/unexamined and blocked/inconclusive remain distinguishable; no manual status patch. |

The independent check must decide the exact observation used in C3, not merely
return a stronger-looking verifier label. A fixture checker can exercise the
software contract but does not establish a real research claim. A valid witness
for one input is not a universal proof; an invalid submitted witness refutes that
witness, not the whole conjecture. Do not admit structural-description failure
as observed domain failure. If the available adapter cannot satisfy C3, record
that gap instead of substituting an operator-attested model judgment.

Prefer a temporary worktree and the real migrated store, not a mock that supplies
the desired outcomes. With implementation authorization, extend existing test
helpers for this single scenario, adding a narrow mapping only where necessary.
Do not introduce generalized orchestration or a new registry before this path
works. No implementation or persistent research-data edits are authorized by
this recipe alone.

## Executable baseline checks

These commands run component tests only. They do not implement C1-C8:

```sh
set -euo pipefail
for name in TestIntegrationChallengeAssessmentPopulation TestIntegrationEvidenceAdmissionClosesReentry; do
  go test ./internal/pipeline -list "^${name}$" | grep -Fx "$name" >/dev/null
  go test ./internal/pipeline -run "^${name}$" -count=1 -v
done
go build ./...
go test ./...
gofmt -l .
```

Inspect verbose output for skipped tests. The `-list` check prevents a renamed
or absent test from silently passing through an empty match. Formatting output
must be empty. Run the actual combined test separately if it exists and record
its exact name; never invent a test name and report an empty match as a pass.

## Evidence bundle and generated coverage

Retain one review-run bundle containing the policy/obligation revisions,
applicability decision, check attempts and exact logs, assessment references,
dependency manifest, subject-population identities, derived decisions and the
coverage-generation outputs. Keep machine-consumed state in the validated
storage mapping; exports are derived snapshots. Record unresolved mapping gaps
instead of building a hand-maintained substitute database.

The coverage export must include the contract's record classification and scope
basis. When historical reviews are later imported, include all relevant review
records and `docs/findings/`; retain each source revision and distinguish a
reported finding from a declared limitation or reconstructed scope. The D2
prose-retention restriction in Pilot-004's protocol is a declared limitation
with an interpretive rule, not new review yield. Keep E11's control/adjudication
question separate. Unknown historical denominators cannot become rates.

Passing all eight cases supports this one obligation's integrated behavior and
coverage derivability. Passing only the component tests does not. A blocker or
failure is a valid outcome, with a smallest coherent handoff. Only after this
slice succeeds should the [preservation pilot](preservation-pilot.md) proceed.
