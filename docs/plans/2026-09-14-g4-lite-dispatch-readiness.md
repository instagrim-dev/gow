# G4-lite dispatch readiness

**Status: execution and preflight interfaces delivered; V1 is retained invalid.**
The public CLI has a bounded three-arm command for the disclosed deterministic
one-run design. A private V1 dispatch completed operationally but failed
frozen-controller identity conformance; see
[`2026-09-14-g4-lite-v1-invalid-disposition.md`](2026-09-14-g4-lite-v1-invalid-disposition.md).
No manifest or command establishes custody, approval, or a funding result.

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
`g4-resource-ceiling/1` artifact. It also requires the exact three
content-free runtime-identity artifacts that the final manifest freezes.
`g4 runtime-identity` exports the pinned executable's actual H0/H1/HG decision
identities; `g4 arm-preflight` must match those records before protected
authoring. Execution repeats the comparison before work, requires the
24-episode 12/6/6 population, supports only the disclosed
`single_run_budget_constrained` design, and executes H0/H1/HG only. It writes
a custodian-local `g4-lite-three-arm-execution/1` receipt with the explicit
label `protected-execution/custody-unverified`.

The command deliberately does not read answer material, compare actual arm
records to frozen snapshots, verify custody, accept an approval reference as
authority, or make a funding decision. Multi-seed execution remains a separate
extension. `shaping diagnose` remains the unrelated four-arm development
diagnostic.

Before a successor protected dispatch, run `g4 calibrate` on an open 24-case
pack with a new `--out` receipt path. Its receipt retains the H0 reference
calibration and any three-arm diagnostic independently, including partial
work and stop reasons. An H0 no-positive-sample outcome, an all-arm
completion ceiling, or a saturated H1 baseline is inconclusive for shaping
value; adjust only the open task-generation/resource procedure, record the
new calibration, and repeat the identity preflight before a new protected
design is frozen.

The retained [public CLI smoke](artifacts/2026-09-14-g4-calibration-cli/)
exercised this path on the published `a632086` executable: it retained 72 H0
reference probes and 72 diagnostic cells. Its deliberately simple open pack
ended 24/24/24 with `COMPLETION_CEILING`, so it verifies receipt retention and
status classification only. It is not the required non-ceiling calibration.

The retained [non-ceiling open calibration](artifacts/2026-09-14-g4-calibration-nonceiling/)
then exercised `b6f3e4a` with a synthetic, openly stored 12/6/6 development
pack. Its arm-blind H0 median was two expansions; the fixed diagnostic was
H0=24, H1=6, HG=24, with 78 retained reference probes and 72 diagnostic
cells. This satisfies the development calibration prerequisite for that exact
executable and resource vector. It remains construction-exposed calibration
evidence, never a protected-evaluation or shaping-value result.

## Required dispatch sequence

1. Pin the release; export and preflight the actual H0, H1, and HG runtime
   identities under the task-directed resource ceiling; then freeze those exact
   snapshot artifacts with the checker, tool, and model identities.
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

Pin the successor release and pass runtime-identity preflight under its chosen
resource vector. Retain a fresh open calibration if either controller or the
resource vector changes; otherwise the retained non-ceiling calibration above
is the development calibration record. A fresh custodian-held pack, real
arm/resource/seed artifacts, explicit execution authorization, separately
metered custody, and a grader remain required. The deterministic one-run
executor does not supply a multi-seed implementation or any external-evidence
authority.
