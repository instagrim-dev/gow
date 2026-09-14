# G4-lite protected-screen final-pack contract

`newf g4 pack` validates and seals the **final, content-free** G4-lite pack
manifest. It is used only after a custodian has authored and separately sealed
the protected episode and answer manifests. The command never reads episode
contents, histories, targets, answers, raw model output, traces, or results.

```bash
newf g4 pack validate --input g4-lite-metadata.json
newf g4 pack seal --input g4-lite-metadata.json --out g4-lite-pack.seal.json
newf g4 pack bind-execution --pre-execution-seal g4-lite-pack.seal.json --observed-metadata observed.json --out execution-binding.json
newf g4 pack inspect g4-lite-pack.seal.json --input g4-lite-metadata.json
newf g4 runtime-identity --resource-ceiling resource.json --arm H0
newf g4 calibrate --episode-pack open-episodes.json --resource-ceiling resource.json
newf g4 arm-preflight --resource-ceiling resource.json --h0-snapshot h0.json --h1-snapshot h1.json --hg-snapshot hg.json
newf g4 execute --manifest g4-lite-metadata.json --episode-pack episodes.json \
  --resource-ceiling resource.json --h0-snapshot h0.json --h1-snapshot h1.json \
  --hg-snapshot hg.json --out protected/execution-receipt.json
```

The final input schema is `g4-lite-pack/2`. It is one UTF-8 JSON object of at
most 256 KiB. Unknown, duplicate, and case-variant fields are refused. The
manifest contains only identity references and declarations:

| Field | Required contract |
|---|---|
| `episode_manifest`, `answer_manifest` | Separate lower-case SHA-256 identities, positive byte lengths, and locators. No protected bytes appear in the record. |
| `calibration_manifest` | Identity of the separately retained open-case sensitivity calibration. It documents pre-seal calibration and does not permit protected-case tuning. |
| `custody` | Declares `unexposed_to_implementation_cases`, `no_protected_content`, `separate_answer_manifest`, and a custody record. These remain declarations rather than independently verified facts. |
| `population` | Exactly 24 episodes: 12 `history_informative`, 6 `history_low_value`, and 6 `history_misleading`, with at least two construction families. |
| `arms` | Three distinct frozen controller snapshots: H0, H1, and HG. They bind one model configuration digest, one tool-catalog digest, the current finite checker, and one per-arm task-directed resource-ceiling identity. H1 requires a recorded `non_implementer` review. Custody work is outside that ceiling and must later be metered. |
| `run_design` | Either the default three fixed seeds with a seed-manifest identity, or one disclosed budget-constrained run. Two-run or ad hoc designs are refused. |
| `endpoint` | The exact objective, including the declared target cost, under the same **task-directed** resource cap. Separately metered custody is excluded from this endpoint and retained for the full-cost decision record. |
| `spending_rule` | Exact run-summed arithmetic: HG must exceed H1 by at least `3 × runs`; H1 minus HG on low-value/misleading cases is a net run-summed loss of at most `1 × runs`; qualifying family advantages are positive run-summed HG advantages on informative episodes only. The rule also requires no invalid certification and HG at least H0. |
| `execution` | The same task-directed resource-ceiling locator and nonnegative provider ceilings. `approval_ref` is an audit pointer only. |

## Deterministic one-run executor

Before a custodian authors protected cases, use `g4 runtime-identity` once for
each arm under the proposed resource ceiling. It emits a content-free
`g4-lite-arm-runtime-identity/1` record with the compiled controller ID and
the decision-snapshot hash it will actually use. The field
`decision_snapshot_sha256` is the compiled controller's frozen-parameter
identity, encoded as `go-json-sha256/1`; it is deliberately distinct from the
SHA-256 of the JSON artifact that records it. The latter is what the manifest's
per-arm `snapshot` reference binds.

Create one such artifact for H0, H1, and HG, then run `g4 arm-preflight` with
the same resource ceiling. It compares each declared runtime identity against
the pinned executable before protected authoring begins. A mismatch is a stop,
not a reason to alter a protected pack. The final manifest must reference those
exact three artifact bytes. `g4 execute` repeats both checks before it opens a
receipt, so a changed controller label, decision snapshot, or snapshot artifact
cannot consume a cell.

`g4 execute` is the public data-only executor for a final manifest with
`single_run_budget_constrained`. Before doing work it verifies the exact bytes
of the supplied episode, resource, and runtime-identity artifacts against the
identities in the final manifest. The episode artifact is `shaping-pack/1` and
must contain the fixed 24-episode 12/6/6 population. The resource artifact is
one strict JSON object:

```json
{"schema":"g4-resource-ceiling/1","expansions":1,"rule_applications":1,"candidates":1,"history_bytes":1,"check_assignments":1,"max_states":1,"max_term_nodes":1}
```

All numeric values are explicit nonnegative ceilings except `max_states` and
`max_term_nodes`, which must be positive. The command executes only H0, H1,
and HG, writes a new custodian-local receipt, and labels it
`protected-execution/custody-unverified`. It never reads answers, establishes
custody, validates an authorization reference, or scores/funds the batch.

## Open sensitivity calibration

`g4 calibrate` accepts an **open** 24-episode 12/6/6 `shaping-pack/1` and a
candidate resource vector. It first finds each H0 minimum expansion allowance
without evaluating H1 or HG, then uses the median positive H0 minimum as a
fixed proposed expansion allowance for one disclosed three-arm diagnostic. Its
JSON output binds the open input and resource bytes, records all H0 minima and
unreachable episodes, reports completions for each arm, and never grants
protected-dispatch readiness.

`H0_COMPLETION_CEILING` means every H0 completion minimum was zero. A
`COMPLETION_CEILING` means H1 or HG completes every open episode at the
proposed allowance. Either status is an inconclusive sensitivity diagnostic:
it provides no shaping-value conclusion and must be resolved with the open
task-generation/resource procedure before a successor protected design is
frozen. Record both successful and failed calibrations; do not tune a fresh
protected pack from completion counts.

## Required sequence

1. Freeze the resource specification at a pinned release. Export H0, H1, and
   HG runtime identities from that executable, record the three content-free
   runtime-identity artifacts, and make `g4 arm-preflight` pass. This is a
   design-stage reference; it does not include future episode or answer hashes.
2. The custodian authors the fresh protected episodes and separately seals the
   episode, answer, calibration, arm-snapshot, resource, and seed artifacts.
3. The custodian builds and seals the final `g4-lite-pack/2` manifest with
   those real identities. It must not invent future hashes or mutate a prior
   final-pack seal.
4. Under a separate execution authorization, the custodian executes the full
   3 × 24 × runs grid and meters custody separately.
5. The custodian records `g4-lite-observed-metadata/1` and runs
   `bind-execution`. The observed record names the exact pre-execution seal
   plus content-free arm-execution, resource-ledger, and result-grid manifest
   identities. The command validates both inputs and their seal binding.
6. A grader separately compares actual arm/resource records with the frozen
   manifest and evaluates the grid. `bind-execution` does not perform that
   comparison, establish chronology, verify custody, or grant authority.

Every final-manifest validation response and seal reports
`PREPARED_NOT_AUTHORIZED`, `protected_execution_authorized: false`, and
`custody_verified: false`. An `approval_ref`, locally matching seal, or
provider budget cannot change those values. A screen outcome remains bounded
to its sealed batch and cannot become a G4 confirmatory claim or funding
authority through this command.
