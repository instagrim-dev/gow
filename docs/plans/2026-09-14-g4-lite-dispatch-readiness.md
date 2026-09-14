# G4-lite dispatch readiness

**Status: execution and preflight interfaces delivered; V1 is invalid and V2S
is inconclusive.** The public CLI has a bounded three-arm command for the
disclosed deterministic one-run design. A private V1 dispatch completed
operationally but failed frozen-controller identity conformance; see
[`2026-09-14-g4-lite-v1-invalid-disposition.md`](2026-09-14-g4-lite-v1-invalid-disposition.md).
The fresh V2S run passed its content-free integrity grade but saturated all
three arms; see [`2026-09-14-g4-v2s-inconclusive-disposition.md`](2026-09-14-g4-v2s-inconclusive-disposition.md).
No manifest or command establishes custody, approval, or a funding result.

## Delivered preparation

`g4-lite-pack/3` validates and seals the final content-free manifest after the
custodian has separately produced real episode, answer, calibration, arm,
resource, and seed identities. `g4 pack bind-execution` validates the final
pack seal and a bounded `g4-lite-observed-metadata/1` record, then binds their
exact bytes. See [`docs/g4-lite-pack-contract.md`](../g4-lite-pack-contract.md).

This does not authorize execution, validate custody, compare actual arm or
resource records against their frozen identities, score the grid, or decide
funding.

## Delivered execution seam

`newf g4 execute` accepts a final `g4-lite-pack/3` manifest, the exact
separately held `shaping-pack/1` episode artifact, and a bounded
`g4-resource-ceiling/1` artifact. It also requires the exact three
content-free `/2` runtime-identity artifacts that the final manifest freezes.
For a new `/3` manifest, it also requires the exact frozen generation-procedure
artifact and checks its protected family namespace, minimum catalog-entry
count, minimum expression-tree depth, and exact primary resource vector against
the supplied episode pack. Historical `/2`
manifests remain readable as completed evidence but cannot carry this binding.
`g4 runtime-identity` exports the running executable's SHA-256 together with
its actual H0/H1/HG decision identities; `g4 arm-preflight` must match those
records before protected
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

Before a successor protected dispatch, freeze a
`g4-lite-calibration-procedure/1` and run `g4 calibrate-procedure` on its
open 24-case pack. It fixes the generation controls, open/protected family and
exposure separation, and every resource vector before the command executes.
Its receipt retains each H0 reference calibration and three-arm diagnostic,
including partial work and stop reasons, and reports whether the one-run
margin and two-informative-family conditions are even attainable on the open
design. An H0 no-positive-sample outcome, an all-arm completion ceiling, a
saturated H1 baseline, or failed open headroom is inconclusive for shaping
value; revise only a new open procedure and repeat identity preflight before
a new protected design is frozen.

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

The [runtime-identity preflight](artifacts/2026-09-14-g4-runtime-preflight/)
then exported H0/H1/HG from `af9d2cf` under that resource vector and passed
`g4 arm-preflight`. Each exported decision-snapshot identity matches the
corresponding identity retained by the non-ceiling calibration. This completes
the public calibration/preflight preparation for the unchanged controller and
resource vector; it remains non-authorizing.

V2S nevertheless saturated its fresh 12/6/6 protected grid at H0=24, H1=24,
HG=24. The independent content-free grade found no artifact or identity
invalidity, but classified the result `INCONCLUSIVE_INCOMPLETE` because the
one-run ceiling produced no discriminating evaluation outcome. The retained
open calibration did not carry enough task-generation/resource pressure into
that protected population. It is preserved; it is not rerun.

The retained [frozen multi-choice open procedure](artifacts/2026-09-14-g4-calibration-procedure/)
records expression-tree depth, catalog-entry count, a declared enabling-step
count, target/history declarations, family, exposure, and no-post-target-
adjustment controls. It does not establish rewrite distance, live route
branching, enabling necessity, or target construction; those need custodian
evidence and substantive-grader assessment. Its prior `/2` result used the
derived H0 median for the high-vector diagnostic and remains a historical
preparation artifact only. The successor [exact-vector calibration](artifacts/2026-09-14-g4-calibration-procedure-exact/)
retains each declared diagnostic at its own resource vector. Its open results leave a maximum possible HG-over-H1
advantage of 18 and two informative families with H1 headroom, so the stated
one-run margin and family thresholds are attainable in this open construction.
This is stronger preparation, not protected evidence or a dispatch decision.

The [current executable-bound `/3` preflight](artifacts/2026-09-14-g4-runtime-preflight-v5-executable-bound/)
was built from `ef58ccd` under that procedure's primary resource vector. It
exports `/2` arm identities that bind H0/H1/HG controller snapshots to
executable SHA-256 `4f05f1186c1e49a148698c9e02b3c1f969ab97a4b3588e86edb27e133aa80766`
and passes `g4 arm-preflight`. The prior `/1` identity preflights remain
historical evidence only; `/3` execution now requires executable-bound `/2`
records. This completes release and controller-identity preparation for a
fresh `/3` custodian pack; it remains non-authorizing.

## Required dispatch sequence

1. Freeze and retain a fresh procedure before protected authoring, then pin
   the release and export/preflight the actual H0, H1, and HG runtime
   identities under the task-directed resource ceiling; then freeze those exact
   snapshot artifacts with the checker, tool, and model identities.
2. A fresh custodian context authors 24 protected episodes and separate answer,
   calibration, arm-snapshot, resource, and seed artifacts. The custodian
   records actual isolation, access controls, exposure history, and H1 review.
3. The custodian creates and seals the final `g4-lite-pack/3` metadata using
   those real identities.
4. The operator supplies an explicit authorization naming the executable hash,
   permitted command, resource and provider ceilings, storage location,
   recipient, run design, and stop conditions.
5. The custodian executes the complete grid, retains all cells and cost
   ledgers, creates observed metadata, and binds it to the earlier seal.
6. A substantive grader separately checks artifact identity/parity, protected
   answers and result quality, resource compliance, and spending arithmetic;
   it returns only a content-free `g4-lite-substantive-grade/2` judgment.
   Completed pass/negative returns need all assessments; early invalid or
   incomplete returns retain available identities, assessment state, and a
   stopping reason. `g4 grade validate`
   verifies the return contract without reading protected evidence.

## Remaining dispatch inputs

The generation/resource procedure, open multi-choice calibration, and current
runtime preflight are delivered. The remaining external gate is a fresh
custodian-held pack authored under that frozen procedure, followed by explicit
execution authorization, separately metered custody, and the substantive
grader. Repeat runtime preflight if the executable or resource vector changes.
The deterministic one-run executor does not supply a multi-seed implementation
or any external-evidence authority.
