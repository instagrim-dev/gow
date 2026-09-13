# Protected G1 dispatch decision packet

**Status: partially recorded, not dispatch-authorized.** This packet is the
minimum operator decision surface for a protected G1 dispatch. A blank,
ambiguous, or unverified field blocks dispatch. Completing this document does
not create authority. The recorded-decision section below captures actual
operator selections; the open rows are the remaining blockers.

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

### Still open — blocks dispatch

| Row | Why it remains open |
|---|---|
| Dispatch authority | The recorded approval scope was "record rows", explicitly not dispatch. Needs approval reference, effective date, permitted commands/outputs, and named outcome recipient. |
| Content boundary | Protected task/answer storage locations and access-log location do not exist yet; they are created at launch and recorded then. |
| Custodian isolation (observed) | Recorded mechanism above; the actual observed boundary is recordable only when the custodian context starts. |

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
