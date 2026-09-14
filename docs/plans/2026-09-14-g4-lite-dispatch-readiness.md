# G4-lite dispatch readiness

**Status: execution interface delivered; dispatch remains unauthorized.** The public CLI now has a bounded three-arm command for the disclosed deterministic one-run design, but no manifest or command establishes custody, approval, or a funding result.

## Delivered preparation

`g4-lite-pack/2` validates and seals the final content-free manifest after the
custodian has separately produced real episode, answer, calibration, arm,
resource, and seed identities. `g4 pack bind-execution` validates the final
pack seal and a bounded `g4-lite-observed-metadata/1` record, then binds their
exact bytes. See [`docs/g4-lite-pack-contract.md`](../g4-lite-pack-contract.md).

This does not authorize execution, validate custody, compare actual arm or
resource records against their frozen identities, score the grid, or decide
funding.

## Delivered execution seam

`newf g4 execute` accepts a final `g4-lite-pack/2` manifest, the exact
separately held `shaping-pack/1` episode artifact, and a bounded
`g4-resource-ceiling/1` artifact. It verifies both content identities before
work, requires the 24-episode 12/6/6 population, supports only the disclosed
`single_run_budget_constrained` design, and executes H0/H1/HG only. It writes
a custodian-local `g4-lite-three-arm-execution/1` receipt with the explicit
label `protected-execution/custody-unverified`.

The command deliberately does not read answer material, compare actual arm
records to frozen snapshots, verify custody, accept an approval reference as
authority, or make a funding decision. Multi-seed execution remains a separate
extension. `shaping diagnose` remains the unrelated four-arm development
diagnostic.

## Required dispatch sequence

1. Pin the release that exposes `g4 execute` and freeze the H0, H1, HG, checker, tool, model, and task-directed resource
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

## Remaining dispatch inputs

A fresh custodian-held pack, real arm/resource/seed artifacts, explicit
execution authorization, separately metered custody, and a grader remain
required. The deterministic one-run executor does not supply a multi-seed
implementation or any external-evidence authority.
