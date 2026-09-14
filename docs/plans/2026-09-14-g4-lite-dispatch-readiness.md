# G4-lite dispatch readiness

**Status: preparation only. Do not dispatch a protected G4-lite batch from
this document.** The final-pack contract is implemented, but the public CLI
has no three-arm protected-screen execution command.

## Delivered preparation

`g4-lite-pack/2` validates and seals the final content-free manifest after the
custodian has separately produced real episode, answer, calibration, arm,
resource, and seed identities. `g4 pack bind-execution` validates the final
pack seal and a bounded `g4-lite-observed-metadata/1` record, then binds their
exact bytes. See [`docs/g4-lite-pack-contract.md`](../g4-lite-pack-contract.md).

This does not authorize execution, validate custody, compare actual arm or
resource records against their frozen identities, score the grid, or decide
funding.

## Blocking execution seam

The shipped public commands are intentionally insufficient for a protected
G4-lite run:

- `newf g4 pack {validate,seal,bind-execution,inspect}` handles content-free
  metadata only.
- `newf shaping diagnose --pack-file` accepts episode data, but runs the
  four-arm development diagnostic, forces a development evidence label, and
  does not bind a `g4-lite-pack/2` manifest.

The internal runner and screen arithmetic do not substitute for a public,
data-only custodian interface. Running them through tests, source edits, or an
ad hoc Go program would violate the public-interface boundary of a protected
batch.

## Required sequence after the executor exists

1. Pin the release that exposes a three-arm G4-lite execution command and
   freeze the H0, H1, HG, checker, tool, model, and task-directed resource
   identities.
2. A fresh custodian context authors 24 protected episodes and separate answer,
   calibration, arm-snapshot, resource, and seed artifacts. The custodian
   records actual isolation, access controls, exposure history, and H1 review.
3. The custodian creates and seals the final `g4-lite-pack/2` metadata using
   those real identities.
4. The operator supplies an explicit authorization naming the executable hash,
   permitted command, resource and provider ceilings, storage location,
   recipient, run design, and stop conditions.
5. The custodian executes the complete grid, retains all cells and cost
   ledgers, creates observed metadata, and binds it to the earlier seal.
6. A grader separately checks artifact identity/parity and evaluates the fixed
   screen rule. It must retain invalid, incomplete, and blocked outcomes.

## Next engineering slice

Provide a data-only `g4` execution command that accepts only a sealed final
manifest, the separately held episode and resource artifacts, and explicit
bounded output paths. It must verify supplied artifact identities before work,
execute exactly H0/H1/HG under the frozen run design, retain task-directed and
custody ledgers separately, reject resource-truncated cells rather than scoring
them as misses, and emit a content-free observed-metadata record for later
binding. That command must still report unverified custody and cannot create
execution authority from an `approval_ref`.
