---
artifact_kind: case-provenance-format
status: operator_approved_2026-09-13
prepared_date_utc: '2026-09-13'
scope: custodian-authored cases for the protected G1 pack
authorizes: nothing beyond the case-provenance row; dispatch authority remains separate
---

# Case provenance and exposure-record format

Approved as drafted by explicit operator selection on 2026-09-13.

One JSON-lines file per pack, `case-provenance.jsonl`, held in custodian
storage beside the task manifest and hashed into it. One line per case, 48
lines total, written at authoring time before any execution. Fields:

| Field | Contract |
|---|---|
| `case_id` | Unique within the pack; matches the task-manifest entry |
| `input_sha256`, `input_bytes` | Exact task-input identity; must match the manifest |
| `answer_sha256` | Identity of the sealed expected answer; content stays in the answer manifest |
| `stratum` | `applicable`, `inapplicable`, or `underspecified` (the 24/16/8 rows) |
| `claim_route` | One of the five registered tool routes |
| `author` | Custodian runtime identity: agent kind, model family, configuration reference |
| `authored_at_utc` | Creation instant; must postdate the recorded controller freeze and predate first execution |
| `method` | `authored-fresh-in-custodian-context`; any other value blocks the case |
| `exposure` | `none` or an explicit list of prior contexts. Any nonempty value blocks the case from the protected pack; renaming or trivially mutating exposed content does not make it `none` |
| `derived_from` | Empty, or references that make the case development-grade |

Rules:

1. The file is append-only during authoring and sealed (SHA-256) with the task
   manifest before execution; a post-execution edit invalidates the batch.
2. Declarations are custodian claims, not verified facts; they are graded at
   most `agent-sealed/v1` together with the recorded isolation boundary.
3. A case whose provenance line is missing, duplicated, or inconsistent with
   the manifests is `invalid` under the approved acceptance matrix.
4. The implementer lane never reads this file's case contents; only its hash
   and per-stratum counts travel in the return packet.
