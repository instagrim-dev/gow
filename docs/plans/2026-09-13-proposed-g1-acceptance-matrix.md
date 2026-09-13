---
artifact_kind: proposed-acceptance-matrix
status: proposed_awaiting_operator_approval
prepared_date_utc: '2026-09-13'
scope: protected G1 dispatch under the recorded decision-packet rows
authorizes: nothing; approval of this matrix fills one packet row only
---

# Proposed acceptance matrix: protected G1 dispatch

This drafts the acceptance rows the decision packet names as missing. It is
**proposed, not approved**. Operator approval of this document supplies the
packet's acceptance-criteria row; it does not supply dispatch authority or any
other open row. Thresholds below are the roadmap's designed-pack progression
rules, identical to the metadata-enforced `progression` gate in `g1-pack/3`;
they are a bounded engineering gate, not population error-rate estimates.

## Case-level outcomes

Every one of the 48 cases ends in exactly one recorded state:

| State | Meaning | Counts toward strata arithmetic? |
|---|---|---|
| `completed-valid` | The routed tool executed, its certificate binds the exact case input bytes, and the recorded verdict is consistent with the certificate | Yes |
| `refused-applicable` | An applicable case the system declined to check | Yes (as a false refusal) |
| `false-certification` | An inapplicable case certified, or any certificate inconsistent with its input | Yes (any occurrence fails the batch) |
| `blocked` | Resource stop, cancellation, or storage failure; receipt retained with reason and incurred costs | No — a blocked case is neither a completion nor a refusal |
| `not-executed` | Never dispatched within the wall clock | No |

Retry rule: a `blocked` case may be re-executed **at most once**, and only for
an infrastructure failure (storage, process interruption) — never to replace a
semantic result. Both receipts are retained; the second, if valid, supplies
the counted outcome. All checks are deterministic and offline; there is no
sampling variance to retry against.

## Stratum acceptance rows

| Stratum (fixed count) | Pass condition | Fail condition |
|---|---|---|
| Applicable (24) | ≥ 23 `completed-valid`; ≤ 1 `refused-applicable` | < 23 valid completions among executed cases |
| Inapplicable (16) | Zero certifications; every result records its specific inapplicability reason | Any `false-certification` |
| Underspecified (8) | Every result names the missing premise; no fabricated conclusion, no silently defaulted value | Any conclusion produced without its premise |
| All 48 | Zero invalid certifications anywhere | Any certificate that does not bind and match its exact input |

## Batch-level verdicts

The stratum arithmetic is evaluated **only on a complete collection**: all 48
cases in a counted state (`completed-valid`, `refused-applicable`, or
`false-certification`) after the retry rule. Exactly one verdict is recorded:

| Verdict | Condition | What it earns |
|---|---|---|
| `pass` | Complete collection; all four stratum rows pass | A costed result presented for the owner's next bounded investment decision. **It does not self-authorize spending, publication, or the next tranche.** |
| `evaluated-negative` | Complete collection; any stratum row fails | A completed evidence task at its recorded scope. Not license to tune the same exposed pack until it passes; a successor needs a new design revision. |
| `inconclusive-incomplete` | Any case `blocked`/`not-executed` after the retry rule, or the wall clock or a per-command reservation ended collection | No stratum arithmetic is computed or reported on the partial set. All partial receipts, blocked reasons, and incurred costs are retained. |
| `invalid` | Exposure or custody breach, manifest/identity mismatch, release-identity drift, ceiling breach, or an oracle found invalid | Recorded at its actual scope; earns nothing; the record is preserved, never rerun in place. |

`invalid` takes precedence over the other three; `inconclusive-incomplete`
takes precedence over `pass`/`evaluated-negative`.

## Validity preconditions (checked before arithmetic)

1. Executable and interface hashes match the packet's recorded release
   identity; the custodian ran only the pinned executable.
2. Task and answer manifests were sealed separately, after controller freeze,
   before execution; every case's input bytes match its manifest entry.
3. The custodian context received only the two-file packet plus the pinned
   executable; the observed isolation boundary was recorded at launch.
4. Provider calls and spend are zero; every reservation stayed within the
   recorded zero-spend ceiling.
5. No case matches previously exposed development content; renaming does not
   reset exposure.

A failed precondition makes the batch `invalid` regardless of case outcomes.

## Evidence grade

A `pass` or `evaluated-negative` from an actually recorded isolation boundary
is at best `agent-sealed/v1`: model-family-dependent, not independent human or
training-lineage evidence. If isolation fails at launch, authoring may continue
only as development-grade work and must be labeled as such; it is never
described as the completed agent-sealed pack.
