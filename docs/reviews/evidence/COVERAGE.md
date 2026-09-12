# COVERAGE

Generated from the normative review ledger. Do not edit by hand: this file is
an export of records, not a status field. Change the records and regenerate.

## Decision

- **decision**: `ELIGIBLE_TO_ADVANCE`
- **policy**: `assessment-admission-decision@1` (whether the selected candidate may guide the next search action)
- **policy id**: `rpol_01M2AQT7G0HTRWQP87T2X4HNYD`
- **owner**: repository-maintainer
- **authority**: docs/reviews/prompts/review-contract.md (four-record mapping) + docs/reviews/prompts/recipes/assessment-admission-decision.md@1
- **scope justification**: bounded to one obligation over the pinned checkout's assessment/admission/decision path; no claim about domain conjectures
- **evidence cutoff**: 2026-09-12T12:00:00Z

Eligibility is scoped permission to advance under this policy. It is not a
claim that any scientific hypothesis is true.

## Obligations

| obligation | mandatory | state | governing assessment | reasons |
| --- | --- | --- | --- | --- |
| `current-assessment-authority@1` | yes | `conforms` | `rasm_01M2AQT7G0HTRWQP87XM3B7EFJ` | — |

## Records

### `current-assessment-authority@1`

- **obligation id**: `robl_01M2AQT7G0HTRWQP87T5YT4Q7E`
- **requirement**: A current decision uses assessments compatible with its declared evidence population and policy. Historical replay remains reproducible without restoring obsolete current authority.
- **owner**: repository-maintainer
- **state**: `conforms`

**Applicability**

- `rapp_01M2AQT7G0HTRWQP87T6H115A2` subject `invariant:inv_01M2AQT7G0HTRWQP87SYYE8PQ3` decision `applies` by repository-maintainer: the subject is a persisted candidate whose lifecycle state is read to select the next frontier target, which is exactly the decision this policy names

**Assessments**

- `rasm_01M2AQT7G0HTRWQP87TVZK2PB2` outcome `conforms` subject `invariant:inv_01M2AQT7G0HTRWQP87SYYE8PQ3` context `cluster_run:clr_01M2AQT7G0HTRWQP87SHM6RNYE; generation:fgr_01M2AQT7G0HTRWQP87TP5VXACZ` by repository-gate:integration-test
  - argument: the campaign that produced the current state assessed cluster run clr_01M2AQT7G0HTRWQP87SHM6RNYE, which is the current compatible population; the generation therefore selected the candidate without a stale-authority exclusion
  - manifest: `rdep_01M2AQT7G0HTRWQP87TTGHHMEA`
  - checks: `rchk_01M2AQT7G0HTRWQP87TG3FRXGF`
  - stale: declared assessment_population dependency clr_01M2AQT7G0HTRWQP87SHM6RNYE is now clr_01M2AQT7G0HTRWQP87WKHQKFKD: the obligation is about compatibility between the decision and its evidence population; a different current population changes what the decision may rely on (result retained; not usable for a current decision)
- `rasm_01M2AQT7G0HTRWQP87XM3B7EFJ` outcome `conforms` subject `invariant:inv_01M2AQT7G0HTRWQP87SYYE8PQ3` context `cluster_run:clr_01M2AQT7G0HTRWQP87WKHQKFKD; generation:fgr_01M2AQT7G0HTRWQP87X6ZD3CJX` by repository-gate:integration-test
  - argument: reassessment under the current population clr_01M2AQT7G0HTRWQP87WKHQKFKD produced campaign run_01M2AQT7G0HTRWQP87X10TPCZ9, so the state that governs the next decision was earned against the population that decision uses; the A0 assessment is retained unchanged as history
  - manifest: `rdep_01M2AQT7G0HTRWQP87XHSM2JCR`
  - checks: `rchk_01M2AQT7G0HTRWQP87XEEF7MSD`

**Check attempts**

- `rchk_01M2AQT7G0HTRWQP87TG3FRXGF` case `C1` mode `executed` outcome `completed` procedure `app.ChallengeInvariants + app.GenerateFrontier` executor repository-gate:integration-test
  - output: `challenge_run=run_01M2AQT7G0HTRWQP87TAAYDB5X; assessment_cluster_run=clr_01M2AQT7G0HTRWQP87SHM6RNYE`
  - resources: zero paid-provider calls; fixture providers only; disposable store
- `rchk_01M2AQT7G0HTRWQP87VXBX8ARX` case `C2` mode `executed` outcome `completed` procedure `app.Evaluate + app.AdmitEvidence (rule pass)` executor repository-gate:integration-test
  - output: `withheld_basis=model-judged failure (single-model-judgment): a model verdict is a judgment, not a verified domain observation; admission requires operator attestation and stays labeled model-judged; population_unchanged=clr_01M2AQT7G0HTRWQP87SHM6RNYE`
  - resources: zero paid-provider calls; fixture providers only; disposable store
- `rchk_01M2AQT7G0HTRWQP87WYH5C8PX` case `C3` mode `executed` outcome `completed` procedure `app.AdmitEvidence (operator attestation) + app.BuildClustering` executor repository-gate:integration-test
  - output: `admitted_signature=msig_01M2AQT7G0HTRWQP87W834YNQN; content_hash=cd5f13371651fa86e3edc340cc0c9f4db24dacfb590470d0f8857f51aa22692f; population_a1=clr_01M2AQT7G0HTRWQP87WKHQKFKD`
  - resources: zero paid-provider calls; fixture providers only; disposable store
- `rchk_01M2AQT7G0HTRWQP87XEEF7MSD` case `C6` mode `executed` outcome `completed` procedure `app.ChallengeInvariants(population=latest) + app.GenerateFrontier` executor repository-gate:integration-test
  - output: `challenge_run=run_01M2AQT7G0HTRWQP87X10TPCZ9; state_after=surviving; generation=fgr_01M2AQT7G0HTRWQP87X6ZD3CJX`
  - resources: zero paid-provider calls; fixture providers only; disposable store
- `rchk_01M2AQT7G0HTRWQP87YAFZ0PGS` case `C7` mode `executed` outcome `completed` procedure `app.ChallengeInvariants(population=discovery) + app.GenerateFrontier` executor repository-gate:integration-test
  - output: `replay_run=run_01M2AQT7G0HTRWQP87XQAYEH3H; replay_state=surviving; compatible_current_authority_preserved=true; historical_row_durable=true`
  - resources: zero paid-provider calls; fixture providers only; disposable store

## Reading this document

- `unexamined` means no assessment exists. It is not a pass and not a failure.
- `inconclusive` means an assessment was made and reached no conclusion.
- `blocked` means a cited check could not execute. A blocker is retained, never
  converted into support for conformance.
- `not_applicable` requires an authorized applicability decision with a rationale.
  Missing implementation never lands here.
- A stale assessment keeps its historical outcome and loses current authority.
