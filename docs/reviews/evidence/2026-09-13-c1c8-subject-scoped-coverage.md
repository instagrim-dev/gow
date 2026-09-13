# COVERAGE

Generated from the normative review ledger. Do not edit by hand: this file is
an export of records, not a status field. Change the records and regenerate.

## Decision

- **decision**: `ELIGIBLE_TO_ADVANCE`
- **policy**: `assessment-admission-decision@1` (whether the selected candidate may guide the next search action)
- **policy id**: `rpol_01M2AQT7G0KHR16YQ3QTCMR6AH`
- **owner**: repository-maintainer
- **authority**: docs/reviews/prompts/review-contract.md (four-record mapping) + docs/reviews/prompts/recipes/assessment-admission-decision.md@1
- **scope justification**: bounded to one obligation over the pinned checkout's assessment/admission/decision path; no claim about domain conjectures
- **evidence cutoff**: 2026-09-12T12:00:00Z

### Provenance

- **generator**: `coverage-generator/3`
- **generated at**: 2026-09-12T12:00:00Z
- **project revision**: `review-gate-fixture-revision-1`
- **contract**: `docs/reviews/prompts/review-contract.md`
- **recipe**: `docs/reviews/prompts/recipes/assessment-admission-decision.md@1`

Generation time is non-semantic metadata: repeated generation from identical
records changes that line and nothing else.

Eligibility is scoped permission to advance under this policy. It is not a
claim that any scientific hypothesis is true.

## Obligations

| obligation | mandatory | state | governing assessment | reasons |
| --- | --- | --- | --- | --- |
| `current-assessment-authority@1` | yes | `conforms` | `rasm_01M2AQT7G0KHR16YQ3VMCKB9J4` | — |

## Records

### `current-assessment-authority@1`

- **obligation id**: `robl_01M2AQT7G0KHR16YQ3QVTPHK53`
- **requirement**: A current decision uses assessments compatible with its declared evidence population and policy. Historical replay remains reproducible without restoring obsolete current authority.
- **owner**: repository-maintainer
- **state**: `conforms`

**Applicability**

- `rapp_01M2AQT7G0KHR16YQ3QYGGFNXX` subject `invariant:inv_01M2AQT7G0KHR16YQ3QQD6AAXX` decision `applies` by repository-maintainer: the subject is a persisted candidate whose lifecycle state is read to select the next frontier target, which is exactly the decision this policy names

**Assessments**

- `rasm_01M2AQT7G0KHR16YQ3RCXZX90V` outcome `conforms` subject `invariant:inv_01M2AQT7G0KHR16YQ3QQD6AAXX` context `cluster_run:clr_01M2AQT7G0KHR16YQ3Q6J7QBSC; generation:fgr_01M2AQT7G0KHR16YQ3R8PQ0P9A` by repository-gate:integration-test
  - argument: the campaign that produced the current state assessed cluster run clr_01M2AQT7G0KHR16YQ3Q6J7QBSC, which is the current compatible population; the generation therefore selected the candidate without a stale-authority exclusion
  - manifest: `rdep_01M2AQT7G0KHR16YQ3RCWGM5Y8`
  - project revision: `review-gate-fixture-revision-1`
  - contract: `docs/reviews/prompts/review-contract.md`
  - recipe: `docs/reviews/prompts/recipes/assessment-admission-decision.md@1`
  - evidence cutoff: `2026-09-12T12:00:00Z`
  - declared dependencies:
    - `assessment_population` = `clr_01M2AQT7G0KHR16YQ3Q6J7QBSC` — the obligation is about compatibility between the decision and its evidence population; a different current population changes what the decision may rely on
    - `policy_revision` = `assessment-admission-decision@1` — a semantically relevant policy revision changes the applicable acceptance criteria
    - `candidate_content` = `inv_01M2AQT7G0KHR16YQ3QQD6AAXX` — the assessment is about this exact candidate; different content is a different subject
    - `project_revision` = `review-gate-fixture-revision-1` — the obligation is about behavior of the authority-selection code; a different checkout can decide the next action differently
  - checks: `rchk_01M2AQT7G0KHR16YQ3R5ESPK5C`
  - stale: declared assessment_population dependency clr_01M2AQT7G0KHR16YQ3Q6J7QBSC is now clr_01M2AQT7G0KHR16YQ3TTEWF6DC: the obligation is about compatibility between the decision and its evidence population; a different current population changes what the decision may rely on (result retained as history; not usable for a current decision)
- `rasm_01M2AQT7G0KHR16YQ3VMCKB9J4` outcome `conforms` subject `invariant:inv_01M2AQT7G0KHR16YQ3QQD6AAXX` context `cluster_run:clr_01M2AQT7G0KHR16YQ3TTEWF6DC; generation:fgr_01M2AQT7G0KHR16YQ3VA6YQ1TV` by repository-gate:integration-test
  - argument: reassessment under the current population clr_01M2AQT7G0KHR16YQ3TTEWF6DC produced campaign run_01M2AQT7G0KHR16YQ3V51JG8FZ, so the state that governs the next decision was earned against the population that decision uses; the A0 assessment is retained unchanged as history
  - manifest: `rdep_01M2AQT7G0KHR16YQ3VK650F5G`
  - project revision: `review-gate-fixture-revision-1`
  - contract: `docs/reviews/prompts/review-contract.md`
  - recipe: `docs/reviews/prompts/recipes/assessment-admission-decision.md@1`
  - evidence cutoff: `2026-09-12T12:00:00Z`
  - declared dependencies:
    - `assessment_population` = `clr_01M2AQT7G0KHR16YQ3TTEWF6DC` — this assessment's authority is bounded to the population it examined
    - `policy_revision` = `assessment-admission-decision@1` — a semantically relevant policy revision changes the applicable acceptance criteria
    - `candidate_content` = `inv_01M2AQT7G0KHR16YQ3QQD6AAXX` — the assessment is about this exact candidate
    - `project_revision` = `review-gate-fixture-revision-1` — the obligation is about behavior of the authority-selection code; a different checkout can decide the next action differently
  - checks: `rchk_01M2AQT7G0KHR16YQ3VFM43QHC`

**Check attempts**

- `rchk_01M2AQT7G0KHR16YQ3R5ESPK5C` case `C1` mode `executed` outcome `completed` procedure `app.ChallengeInvariants + app.GenerateFrontier` executor repository-gate:integration-test
  - procedure revision: `docs/reviews/prompts/recipes/assessment-admission-decision.md@1`
  - inputs: `invariant=inv_01M2AQT7G0KHR16YQ3QQD6AAXX; population_policy=latest`
  - environment: `go test ./internal/pipeline`
  - output: `challenge_run=run_01M2AQT7G0KHR16YQ3R03S7AA3; assessment_cluster_run=clr_01M2AQT7G0KHR16YQ3Q6J7QBSC`
  - resources: zero paid-provider calls; fixture providers only; disposable store
- `rchk_01M2AQT7G0KHR16YQ3SG3HG392` case `C2` mode `executed` outcome `completed` procedure `app.Evaluate + app.AdmitEvidence (rule pass)` executor repository-gate:integration-test
  - procedure revision: `docs/reviews/prompts/recipes/assessment-admission-decision.md@1`
  - inputs: `evaluation=evl_01M2AQT7G0KHR16YQ3RNSQA3PM`
  - environment: `go test ./internal/pipeline`
  - output: `withheld_basis=model-judged failure (single-model-judgment): a model verdict is a judgment, not a verified domain observation; admission requires operator attestation and stays labeled model-judged; population_unchanged=clr_01M2AQT7G0KHR16YQ3Q6J7QBSC`
  - resources: zero paid-provider calls; fixture providers only; disposable store
- `rchk_01M2AQT7G0KHR16YQ3V2CH51ZX` case `C3` mode `executed` outcome `completed` procedure `app.WitnessCheck(--procedure equal-denominator) + app.AdmitEvidence (rule) + app.BuildClustering` executor repository-gate:integration-test
  - procedure revision: `docs/reviews/prompts/recipes/assessment-admission-decision.md@1`
  - inputs: `proposal=fpr_01M2AQT7G0KHR16YQ3SQNRTXNJ; procedure=equal-denominator; params=n=7`
  - environment: `go test ./internal/pipeline`
  - output: `admitted_signature=msig_01M2AQT7G0KHR16YQ3TGWMT06B; content_hash=e8a13144e27b9db695a4409c9ab873598ca245f7867e241e0801eb749e42c467; observation_kind=domain-checked-failure; admitted_by=rule; attempt_binding_verified_by_recomputation=true; population_a1=clr_01M2AQT7G0KHR16YQ3TTEWF6DC`
  - resources: zero paid-provider calls; fixture providers only; disposable store
- `rchk_01M2AQT7G0KHR16YQ3VFM43QHC` case `C6` mode `executed` outcome `completed` procedure `app.ChallengeInvariants(population=latest) + app.GenerateFrontier` executor repository-gate:integration-test
  - procedure revision: `docs/reviews/prompts/recipes/assessment-admission-decision.md@1`
  - inputs: `invariant=inv_01M2AQT7G0KHR16YQ3QQD6AAXX; population=clr_01M2AQT7G0KHR16YQ3TTEWF6DC`
  - environment: `go test ./internal/pipeline`
  - output: `challenge_run=run_01M2AQT7G0KHR16YQ3V51JG8FZ; state_after=surviving; generation=fgr_01M2AQT7G0KHR16YQ3VA6YQ1TV`
  - resources: zero paid-provider calls; fixture providers only; disposable store
- `rchk_01M2AQT7G0KHR16YQ3W79EK2YD` case `C7` mode `executed` outcome `completed` procedure `app.ChallengeInvariants(population=discovery) + app.GenerateFrontier` executor repository-gate:integration-test
  - procedure revision: `docs/reviews/prompts/recipes/assessment-admission-decision.md@1`
  - inputs: `invariant=inv_01M2AQT7G0KHR16YQ3QQD6AAXX; population=clr_01M2AQT7G0KHR16YQ3Q6J7QBSC`
  - environment: `go test ./internal/pipeline`
  - output: `replay_run=run_01M2AQT7G0KHR16YQ3VNZR6NVJ; replay_state=surviving; compatible_current_authority_preserved=true; historical_row_durable=true`
  - resources: zero paid-provider calls; fixture providers only; disposable store

## Reading this document

- `unexamined` means no assessment exists. It is not a pass and not a failure.
- `inconclusive` means an assessment was made and reached no conclusion.
- `blocked` means a cited check could not execute. A blocker is retained, never
  converted into support for conformance.
- `not_applicable` requires an authorized applicability decision with a rationale.
  Missing implementation never lands here.
- A stale assessment keeps its historical outcome and loses current authority.
- An assessment whose declared current dependencies were not supplied to the
  generator is `compatibility unknown`: the historical outcome stands as history
  and does not grant a current decision. Supply the missing current values and
  regenerate to advance.
- The provenance block names the generator and the governing assessments' input
  revisions. A snapshot never expires as history, but reusing one for a CURRENT
  decision requires those inputs to still be compatible.
