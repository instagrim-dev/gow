---
artifact_kind: nonprotected-provider-path-reproduction
status: incomplete
prepared_date_utc: '2026-09-13'
availability_preflight_utc: '2026-09-14T00:34:09Z'
scope: host-versus-provider interruption diagnosis
authorizes: no provider call, spend, protected-material access, or G1 dispatch
---

# Nonprotected provider-path reproduction

This capsule distinguishes a host-path interruption from a provider/model
interruption without reusing the invalid G1 pack. It is a diagnostic of one
specified interaction, not a test of the protected evaluator, research claim,
or custody boundary.

## Controlled task

Use one synthetic, nonprotected data-only task: author a small JSON input from
the public v2 schema, then state which allowlisted executable command would
consume it. The task must contain no supplied binary, source, protected input,
answer, transcript, API credential, production data, exploit content, or
instruction to inspect a binary. Keep the exact prompt and every attachment in
private local evidence; the committed record stores their SHA-256 values and
byte lengths only.

The first run tests this normal public-interface task. It does not replay the
attempt-4 conversation or establish why that later interaction was blocked.
Replaying a specific request needs a separate authorization, a determination
that the request contains no protected material, and a new private-evidence
record.

## Comparison contract

To test the host path, hold constant:

- exact prompt and attachments, identified by SHA-256 and byte length;
- selected model name and model version where the providers expose one;
- tool permissions, task budget, and any system/developer context supplied to
  the model; and
- one request at a time, with the same expected non-sensitive result.

Vary only the execution host and access route: one Cursor run and one direct,
approved provider route using that same model. If either route cannot offer the
same model, record `model_unavailable` and do not interpret the comparison as
an IDE-only test. Changing both host and model measures a combined product
stack, not Cursor alone.

## Availability preflight

At `2026-09-14T00:34:09Z`, a fresh local Cursor chat exposed only `Claude Opus
4.8 Bedrock (US) 300K` (`us.anthropic.claude-opus-4-8`). The recorded G1 Cursor
run used `Claude Opus 4.7 Bedrock (US) 300K`. The selected model therefore
cannot be held constant on the available Cursor route. No approved direct route
for that historical model was exposed to this task.

No synthetic prompt was transmitted, no provider request or spend occurred,
and no protected material was opened. This is an `incomplete` availability
preflight, not a Cursor control arm or a host-only comparison. A later run must
use a direct route and a Cursor route that both expose the same recorded model,
then create one new private evidence record for the paired interaction.

## Per-arm procedure

1. Create the synthetic task and compute the prompt and attachment identities.
   Record the selected model, host product/version, access route, timestamp,
   and configuration before sending the request.
2. Execute a single request. Record the full prompt, attachments, complete
   response, exact error, timestamp, and request ID in a private local
   evidence directory. Record only locators, hashes, and lengths in the
   committed report.
3. If a tool/action sequence begins, record the last completed action and the
   next requested action. Do not continue an interaction after it crosses the
   public task boundary.
4. Mark the arm `completed`, `provider_blocked`, `host_error`,
   `transport_error`, `model_unavailable`, or `not_run`. A provider block is
   evidence of an interruption, not an attribution to Cursor or misconduct.
5. Run the paired arm only with the identical interaction identity. Preserve
   both private-evidence directories and append the content-free report.

## Interpretation

| Paired result | Permitted conclusion |
|---|---|
| Same-model direct route completes and Cursor blocks | The interaction differs by host/integration/access path. The report does not identify which layer in that path caused it. |
| Both routes block | The task/model/account context remains a plausible shared cause; the result does not show that Cursor caused the interruption. |
| Both routes complete | The minimal task does not reproduce the interruption. It does not clear the prior interaction or authorize a protected retry. |
| One route lacks the model or approved access | No host-only comparison was run. Record the missing route rather than substituting another model. |

Use the report schema and synthetic example in
[`artifacts/2026-09-13-provider-path-reproduction/`](artifacts/2026-09-13-provider-path-reproduction/).
The report is an operational receipt: it does not alter the G1 grade, grant
provider access, or authorize a clean-room dispatch.
