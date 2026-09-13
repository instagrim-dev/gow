# Protected G1 dispatch decision packet

**Status: decision-complete; dispatch authorized 2026-09-13; execution in
flight.** This packet is the minimum operator decision surface for a
protected G1 dispatch. A blank, ambiguous, or unverified field blocks
dispatch. Completing this document does not create authority beyond the
recorded operator decisions below; the observed-isolation row is graded from
the custodian's return packet.

## Fixed engineering contract

The dispatched executable must expose the two-file custodian packet:

- [`CUSTODIAN_INTERFACE_BRIEF.md`](CUSTODIAN_INTERFACE_BRIEF.md)
- [`CUSTODIAN_PUBLIC_INTERFACE.md`](CUSTODIAN_PUBLIC_INTERFACE.md)

The G1 metadata contract is `g1-pack/3`: 48 cases, exactly 24 applicable, 16
inapplicable, and 8 underspecified; all five claim routes; separate task and
answer manifests; and content-free metadata sealing. A valid metadata seal is
always `PREPARED_NOT_AUTHORIZED`.

## Operator decisions required before dispatch

| Required decision | Required recorded value | Blocks when absent |
|---|---|---|
| Release identity | Immutable committed revision, executable SHA-256, executable build command, and public-interface SHA-256 | A dirty working tree or an unpinned executable cannot be dispatched. |
| Frozen execution contract | Controller version and snapshot; tool-registry identity; catalog; budget vector; policy ID; obligation IDs; assessment criteria | The custodian must not choose or mutate these values. |
| Custodian isolation | Actual runtime, fresh-context identifier, access controls, custodian storage locator, and permitted recipients | A role label or new conversation alone is insufficient. |
| Content boundary | Protected task storage, separate answer storage, access-log location, and explicit implementer denial of access | Task/answer contents cannot enter the implementation lane. |
| Case plan | Pack ID; the fixed 24/16/8 outcome strata; route coverage mix totaling 48; source/provenance and exposure-record format | No exposed or self-authored development task may enter this pack. |
| Resource ceiling | Per-command assignment/instance/submission reservations; storage limits; wall-clock limit; provider-call and spend ceilings; cancellation handling | Existing development ceilings do not authorize this new vector. |
| Dispatch authority | Operator approval reference, effective date, permitted commands, permitted outputs, and named outcome recipient | An `approval_ref` in metadata does not establish this decision. |
| Acceptance criteria | The applicable row(s) from the missing `ACCEPTANCE_MATRIX.md`, including treatment of blocked, invalid, and incomplete cases | The observed 23-of-24 progression criterion is not a substitute for an approval. |

## Recorded operator decisions (2026-09-13, operator session)

Recorded from an explicit operator selection in the working session on
2026-09-13. Scope of the approval: **record these rows**; it did not grant
dispatch authority.

### Release identity — RECORDED

| Field | Recorded value |
|---|---|
| Committed revision | `0cd7284c6b418a3d3c17f9387e894492a5a2f690` (published `main`, clean tree, CI passed) |
| Build command | `CGO_ENABLED=0 go build -trimpath -o newf ./cmd/newf` with Go 1.26.6 (`go 1.25.0` module floor) |
| Executable SHA-256 | `5d052272922483cd03bcbb1186d05e59b5f2da9aa11d9fbc1927c701e9d72690` (byte-identical across two independent builds) |
| Public interface SHA-256 | `e3a2ada6582d047b064b6dbd76f1c769bb9403a2b82478cdb976b1df920e2c59` (`CUSTODIAN_PUBLIC_INTERFACE.md`) |
| Interface brief SHA-256 | `dc8fb24e1abf18039f4122d3264bbd8f44462efe632ae29ac77213c231425c4e` (`CUSTODIAN_INTERFACE_BRIEF.md`) |

### Resource ceiling — RECORDED

Zero-spend development-class vector, chosen explicitly as a new vector (no
carry-over from D3a-class ceilings):

- offline execution; **zero provider calls and zero provider spend**;
- at most **4096 reserved assignments per check command** (the fixed
  four-bit, 1–3-variable domain bound);
- at most **48 cases** (the fixed `g1-pack/3` composition);
- **15-minute wall clock** for the authoring/checking batch;
- cooperative cancellation required; an interrupted check is retained as
  blocked, never as false.

### Custodian runtime — RECORDED

A **fresh cloud agent on its own VM and branch**, receiving only the two-file
packet (`CUSTODIAN_INTERFACE_BRIEF.md`, `CUSTODIAN_PUBLIC_INTERFACE.md`) plus
the pinned executable identified above. The observed boundary (runtime
identifier, access controls, storage locators, access log) must still be
recorded **at launch time**; this row names the chosen mechanism, not yet an
observed isolation fact.

### Acceptance criteria — RECORDED

The [G1 acceptance matrix](2026-09-13-proposed-g1-acceptance-matrix.md) was
approved as drafted by explicit operator selection on 2026-09-13: case states,
stratum rows identical to the `g1-pack/3` progression gate, four batch
verdicts with `invalid` > `inconclusive-incomplete` > `pass`/
`evaluated-negative` precedence, a single infrastructure-only retry rule,
validity preconditions, and the `agent-sealed/v1` grade ceiling.

### Case-provenance format — RECORDED

The [case provenance and exposure-record format](2026-09-13-proposed-g1-case-provenance-format.md)
was approved as drafted by explicit operator selection on 2026-09-13:
append-only JSONL sealed with the task manifest before execution, explicit
per-case exposure declarations with `none` required for protected cases, and
manifest-consistency rules feeding the approved acceptance matrix.

### Dispatch authority — RECORDED

Granted by the explicit operator instruction **"dispatch"** on 2026-09-13 in
the working session, following the recorded rows above. Effective
immediately. Permitted commands: exactly the data-only check, retrieval, and
`g1 pack` commands in `CUSTODIAN_PUBLIC_INTERFACE.md`, under the recorded
zero-spend ceiling. Permitted outputs: the return packet defined in the
dispatch authorization. Named outcome recipient: the operator, via the
committed return packet on the custodian branch. The full authorization text
is retained at
[`2026-09-13-g1-dispatch-authorization.md`](2026-09-13-g1-dispatch-authorization.md)
and rides with the custodian packet as `DISPATCH_AUTHORIZATION.md`.

### Frozen execution contract — RECORDED

- Tool registry, procedures, and checker versions: as shipped at revision
  `0cd7284` (the pinned executable is the frozen tool interface).
- Dispatch executable (linux/amd64 for the cloud custodian VM): SHA-256
  `006f466a1dac2e7370a26be8c25b75cb7ff85094cc6cca3159de1bab47dc37d8`, built
  from an exact export of revision `0cd7284` with
  `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -buildvcs=false ./cmd/newf`,
  Go 1.26.6, byte-identical across two independent builds. The earlier
  recorded darwin/arm64 hash `5d052272…9d72690` remains the local-platform
  identity; it embeds the VCS stamp of `0cd7284` and reproduces only at that
  revision.
- Policy and obligations: the exact `review policy` definition frozen in the
  authorization (`g1-protected-usefulness@1`, five mandatory route
  obligations). IDs are storage handles generated in the custodian database
  and are recorded in the return packet before case authoring.
- Budget vector: the recorded zero-spend ceiling.

### Content boundary — RECORDED (procedural, with declared limitations)

Protected material lives only on the custodian's working branch
(`protected/tasks/`, `protected/answers/`, `protected/receipts/`,
`protected/case-provenance.jsonl`), with an append-only self-declared
`protected/ACCESS_LOG.md`. Declared limitations, recorded rather than hidden:
git provides no enforced read barrier — implementer-lane reads of the pushed
branch are technically possible and are not mechanically logged; the access
log and exposure record are custodian declarations. The clean-room base
branch `g1-custodian-cleanroom` (`742bb689…`) is an orphan commit containing
only the two-file packet, the dispatch authorization, and the pinned
executable; it carries no repository history. The custodian VM can, in
principle, fetch other refs of this repository; the authorization forbids it
and requires an explicit declaration either way. Under the approved
acceptance matrix these limitations cap the grade at `agent-sealed/v1`
(model-family-dependent) at best.

### Still open — recordable only at completion

| Row | Status |
|---|---|
| Custodian isolation (observed) | The custodian records the actually observed runtime boundary in its custody/exposure record at launch; graded when the return packet is read. |

### Launch record

- **2026-09-13, attempt 1 (cloud, session-default model):** infrastructure
  failure before any custodian context was created — the cloud environment
  rejected the model ("The AI model you selected isn't available", cloud
  subagent `bc-74952bfe-d1ba-4283-a1e4-e47baba19e76`, server request
  `78935238-d60b-452d-b8f7-18a3e7688a5c`). No protected material was created
  or exposed; the clean-room branch was untouched.
- **2026-09-13, attempt 2 (cloud, `composer-2.5-fast`):** same infrastructure
  failure (cloud subagent `bc-9ad0d862-e868-4bb3-b5cc-f12a2545cf70`, server
  request `aa7d071c-fe98-46c0-95fe-5847c8e345de`). No custodian context was
  created.

- **2026-09-13, attempt 3 (local fresh-context subagent):** operator
  explicitly substituted the mechanism after the cloud failures. The
  custodian runs in a dedicated scratch workspace containing only the
  two-file packet, the frozen authorization, a local-substitution addendum,
  and a pinned darwin/arm64 executable (SHA-256
  `c105248752eb2e5e708ebe9c1a48f30d389a822036452e61f4ef41c0c4a93123`, built
  from the same exact export of revision `0cd7284` with
  `CGO_ENABLED=0 go build -trimpath -buildvcs=false ./cmd/newf`,
  byte-identical across two builds). The addendum records the isolation
  downgrade plainly: the custodian shares the implementer host; the
  workspace boundary is procedural, not enforced; the achievable grade is at
  most `agent-sealed/v1` with a declared shared-host limitation, else
  development grade. Protected storage moves from a git branch to the
  workspace directory; all other frozen values are unchanged. The unused
  clean-room branch `g1-custodian-cleanroom` is retained for a future cloud
  retry.

- **2026-09-13, attempt 3 (local fresh-context subagent, launched):**
  infrastructure failure at the first model request — the provider blocked
  the request under its usage policy before the custodian produced a single
  action (subagent `f69c37e5-9ca2-4d58-b0fb-45831b92c3c8`). The workspace
  was untouched afterward (the five staged files only; no protected
  material was created) and the transcript contains only the dispatched
  prompt. Assessed cause: the prompt's confidentiality-heavy phrasing
  pattern-matched a provider policy filter; the underlying task (authoring
  held-out JSON test cases for a local offline checker) is benign.
- **2026-09-13, attempt 4 (local fresh-context subagent, replanned
  prompt):** same mechanism and identical frozen documents; the dispatch
  prompt was rewritten in plain language describing the held-out test-set
  task directly, with the same workspace-only boundary and summary
  restrictions. No contract value changed.

Per the operative handoff, a failed isolated-agent launch is preserved as an
infrastructure failure; it does not make implementer-authored cases
independent, and development-grade authoring must not be described as the
completed agent-sealed task. The recorded cloud-agent custodian mechanism
cannot currently be realized; substituting a different mechanism (for
example, a local subagent sharing the implementer host, with strictly weaker
observable isolation) is a new operator decision, not a repair this lane may
make silently.

## Required pre-dispatch artifacts

The operator retains or verifies these before authoring cases:

1. A clean committed release and reproducible executable hash.
2. The completed decision values above, including actual—not asserted—isolation
   and access controls.
3. A fresh custodian context that receives only the two-file packet and the
   pinned executable.
4. Separate empty protected-storage locations for tasks, answers, and execution
   receipts, with an access log.
5. A manifest format for exact task/answer bytes, SHA-256, byte lengths,
   provenance, exposure history, and creation times.

## Permitted dispatch sequence after approval

1. Freeze and record the approved release and execution contract.
2. Start the isolated custodian context and record the observed boundary.
3. Author and separately seal task and answer manifests in custodian storage.
4. Run only the approved data-only check commands with the approved
   reservations; retain every result, including refusals and interruptions.
5. Create, validate, and immutably seal the content-free G1 metadata manifest.
6. Send only the approved return packet to the designated recipient. Do not
   reveal protected contents or interim aggregate outcomes to the implementer.

## Explicit non-decisions

This packet does not approve provider use, spending, publication, outreach,
G4/G2 work, an external-evidence claim, or a transition to a higher epistemic
state. `agent-sealed/v1` may be recorded only from actual isolation and
post-freeze evidence; it remains model-family-dependent.
