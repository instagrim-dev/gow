# G4-lite protected-screen freeze contract

`newf g4 pack` prepares a **content-free** freeze receipt for the protected
G4-lite shaping screen. It binds the comparison design before protected episode
or answer material is authored or executed. The command never reads episodes,
histories, targets, answers, raw model output, traces, or observed results.

```bash
newf g4 pack validate --input g4-lite-metadata.json
newf g4 pack seal --input g4-lite-metadata.json --out g4-lite-pack.seal.json
newf g4 pack bind-execution --pre-execution-seal g4-lite-pack.seal.json --observed-metadata observed.json --out execution-binding.json
newf g4 pack inspect g4-lite-pack.seal.json --input g4-lite-metadata.json
```

The input schema is `g4-lite-pack/1`. It is one UTF-8 JSON object of at most
256 KiB. Unknown, duplicate, and case-variant fields are refused. A valid
manifest contains only identity references and declarations:

| Field | Required contract |
|---|---|
| `episode_manifest`, `answer_manifest` | Separate lower-case SHA-256 identities, positive byte lengths, and locators. No protected bytes appear in the record. |
| `calibration_manifest` | Identity of the separately retained open-case sensitivity calibration. Calibration is bound before sealing; it does not permit protected-case tuning. |
| `custody` | Declares `unexposed_to_implementation_cases`, `no_protected_content`, `separate_answer_manifest`, and a custody record. These remain declarations rather than independently verified facts. |
| `population` | Exactly 24 episodes: 12 `history_informative`, 6 `history_low_value`, and 6 `history_misleading`, with at least two construction families. |
| `arms` | Three distinct frozen controller snapshots: H0, H1, and HG. They must bind one model configuration digest, one tool-catalog digest, the current finite checker, and one per-arm resource-ceiling identity. H1 requires a recorded `non_implementer` review. Custody work is declared outside the per-arm ceiling and must later be metered. |
| `run_design` | Either the default three fixed seeds with a seed-manifest identity, or one disclosed budget-constrained run. Two-run or unsealed ad hoc designs are refused. |
| `endpoint` | The exact objective, including the declared target cost, under the same total resource cap. |
| `spending_rule` | The roadmap's fixed screen rule: no invalid certification, HG ≥ H1 + 3, HG loses ≤ 1 over low-value/misleading episodes, successful differences in at least two families, HG ≥ H0, and rounding against funding. |
| `execution` | The same resource-ceiling locator and nonnegative provider ceilings. `approval_ref` is an audit pointer only. |

The record is intentionally a **freeze**, not an execution or a result. Every
validation response and seal reports `PREPARED_NOT_AUTHORIZED`,
`protected_execution_authorized: false`, and `custody_verified: false`. An
`approval_ref`, a locally matching seal, or a declared provider budget cannot
change those values.

A later authorized custodian must retain the actual episode, answer,
calibration, arm-snapshot, resource, seed, and execution records at the
referenced identities. After execution, `bind-execution` writes a separate
`g4-lite-execution-binding/1` receipt that hashes the pre-execution seal and
content-free observed metadata; it establishes neither chronology nor custody
by itself. The custodian must then execute the complete 3 × 24 × runs grid; meter custody
separately; and submit the retained grid to the existing screen evaluator. A
screen outcome remains bounded to its sealed batch and cannot become a G4
confirmatory claim or funding authority through this command.
