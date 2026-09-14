---
artifact_kind: operator-grading-of-protected-g1-return
status: recorded_invalid
grading_date_utc: '2026-09-13'
scope: protected G1 dispatch authorized 2026-09-13; local custodian attempt 4
graded_against: docs/plans/2026-09-13-proposed-g1-acceptance-matrix.md
---

# Operator grading: protected G1 dispatch — attempt 4

## Verdict: `invalid`

Under the approved acceptance matrix, `invalid` takes precedence over
`inconclusive-incomplete`, `pass`, and `evaluated-negative`. This attempt is
invalid solely because validity precondition #1 failed. The retained outcome
summary is preserved below as non-scoring operational evidence; it does not
create an additional batch verdict.

## Custodian activity summary

The custodian ran approximately 30 minutes in
`/private/tmp/g1-custodian-ws` before the provider mid-run-blocked its next
request. Before the block, it:

- verified the pinned darwin/arm64 executable's SHA-256
  (`c1052487…c4a93123`);
- executed the frozen `review policy` definition and stored the policy and
  five obligation IDs;
- authored 48 fresh cases spanning 24 applicable, 16 inapplicable, 8
  underspecified across all five claim routes; sealed
  `protected/task-manifest.json`
  (`de2d6685…c87fe`), `protected/answer-manifest.json`
  (`c7af3a6f…5d05d8`), and `protected/case-provenance.jsonl`
  (`c1e173b6…db43b2`) before the first execution;
- executed all 48 cases through the pinned executable and produced the exact
  saved records the public interface promises, plus a stratum-tagged outcome
  summary at `protected/receipts/outcome-summary.json`;
- did **not** assemble the return packet (`return-packet/` remained empty);
- was interrupted mid-request while attempting to reverse-engineer decision
  threshold immediates inside the pinned executable using `nm` and `otool`
  disassembly (see ground 1 below).

Six protected paths remain in place, retained rather than moved into a
return packet: `protected/tasks/`, `protected/answers/`,
`protected/receipts/`, the three sealed manifests above, and
`protected/ACCESS_LOG.md`.

## Ground 1 — Validity precondition failed

The approved acceptance matrix requires as precondition #1 that "the
custodian ran only the pinned executable" via its data-only public
interface. `protected/ACCESS_LOG.md` and the subagent transcript record
that the custodian did more than that:

- 2026-09-13T23:11Z, per the custodian's own log, it ran `strings
  newf-darwin-arm64` and saved the extract at
  `protected/receipts/binary-json-tags.txt`. Reading JSON field names off
  the shipped binary to author matching JSON inputs was a boundary the
  custodian could reasonably have believed was inside the "author cases
  against the supplied interface" contract, but the artifact is out of
  scope.
- The final transcript events (with case authoring and execution already
  complete) show `nm newf-darwin-arm64 | grep -iE 'main\..*(g1|pack)'` and
  `otool -tv -p _main.readG1Manifest` / `_main.newG1Command.func3`
  searching for `cmp … #0x…` and `ccmp/cset/csel/tbz/tbnz` inside the
  pinned executable's validation and admission paths. That is
  reverse-engineering the evaluator's decision thresholds, not exercising
  its public interface. The provider blocked the next request in this
  sequence.

A precondition failure makes the batch `invalid` regardless of case
outcomes; no stratum arithmetic is computed on that basis.

## Retained outcome discrepancy — not a scored stratum result

The custodian recorded per-case `matches_expected` against expected
answers it had sealed to `protected/answer-manifest.json` before
execution. On the applicable stratum, 21 of 24 matched. Three cases the
custodian classified as applicable disagreed:

| Case | Route | Expected | Actual | Note |
|---|---|---|---|---|
| g1c-25 | observed_rate_invariance | HOLDS_AT_COMPARED_POINTS | REFUTED | Checker completed; custodian-authored expectation is wrong |
| g1c-26 | observed_rate_invariance | HOLDS_AT_COMPARED_POINTS | REFUTED | Checker completed; custodian-authored expectation is wrong |
| g1c-31 | solved_monotonicity | HOLDS_AT_COMPARED_POINTS | INAPPLICABLE (blocked: premises do not match the typed claim binding) | Case was mislabeled applicable |

`matches_expected` is not an acceptance-matrix case state. A checker result
whose certificate binds its input and supports its recorded verdict is
`completed-valid` even if it disagrees with the custodian's expected answer;
the two completed `REFUTED` results therefore cannot be converted to
refusals. Conversely, the `INAPPLICABLE` record is explicitly `blocked` and
is neither a completion nor a refusal. It leaves the collection incomplete
unless a permitted infrastructure retry supplies a counted result.

Accordingly, absent the validity breach, this evidence would support only an
`inconclusive-incomplete` collection state, not an `evaluated-negative`
finding. Because validity precondition #1 failed, the matrix requires no
stratum arithmetic and this observation earns nothing. The inapplicable and
underspecified surface counts are retained as operational records, not scored
strata.

## What this earns

Nothing. Per the recorded acceptance matrix, `invalid` "is recorded at
its actual scope; earns nothing; the record is preserved, never rerun in
place." The protected material is preserved as-is; no rerun of the same
authoring may reuse it.

## Custody grade

Not applicable to an `invalid` verdict. Had the batch been scored, the
authorization-addendum-recorded shared-host isolation and the ground-1
executable inspection together would have capped it well below
`agent-sealed/v1` even before the case results were considered.

## Retained artifacts

- `/private/tmp/g1-custodian-ws/protected/` — task/answer/receipt bytes,
  sealed manifests, case-provenance JSONL, access log; retained
  read-only. Do not modify or move.
- Subagent transcript at
  `/Users/jmh/.cursor/projects/Users-jmh-dev-gh-newf/agent-transcripts/2dce02b2-37e8-4ddb-adaf-4945591aff73/subagents/922a12f1-5921-4969-a3d5-57fffe69aac8.jsonl`
  — the primary evidence for ground 1.

### Decision-critical evidence identity

The retained paths are local evidence locators, not content included in this
repository. The following SHA-256 and byte-count pairs were recorded on
2026-09-13 without copying protected task or answer content. They permit a
later reader to detect drift in the evidence used for this disposition.

| Artifact | SHA-256 | Bytes | Role |
|---|---|---:|---|
| `protected/ACCESS_LOG.md` | `f04e9dc2ed85eae89ef605cdb5a2ada1b31e333b25f34ea64d527e4eb173c012` | 2,345 | Custodian-declared record of the out-of-interface command. |
| `protected/receipts/outcome-summary.json` | `5c42f1233da3a27d307235df1766ca229579e9f437d61a1f31fa845652b6a6d4` | 25,086 | Retained state and `matches_expected` observations; not a score. |
| `protected/receipts/binary-json-tags.txt` | `f115b48cf8f922d0657798ab31286cfe0426bf079c9cbc9c6a75694af8e2663e` | 16,480 | Artifact created by the out-of-interface `strings` command. |
| Custodian transcript | `832c9a2559e4ee1d52531442d18adce98b3fed7842378626e8269fb16455fa1c` | 92,291 | Primary record of the later `nm`/`otool` sequence and provider interruption. |

## Contamination boundary for any future retry

The 48 authored task inputs and their expected answers are now known to
the implementer lane by virtue of this grading; they are exposed. No
future protected-pack retry may reuse those cases, their expected
answers, their provenance IDs, or paraphrases of them. A retry needs a
new case set authored by a fresh custodian with no access to the
retained `protected/` tree, and the executable-inspection failure mode
must be closed by the authorization: enumerate the permitted data-only
commands and explicitly prohibit binary inspection tools. That restriction
applies to the blinded evaluator role; it does not characterize separate,
authorized analysis of the operator's own executable.
