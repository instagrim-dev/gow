# Protected G1 dispatch decision packet

**Status: proposed, not authorized.** This packet is the minimum operator
decision surface for a protected G1 dispatch. A blank, ambiguous, or unverified
field blocks dispatch. Completing this document does not create authority.

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
