# G1 protected-pack metadata contract

`newf g1 pack` prepares a **content-free** metadata record for the protected
G1 usefulness pack described in the roadmap and public custodian handoff. It
does not receive protected cases, expected answers, raw outputs, or agent
transcripts. Task and answer material remain separate custodian-held manifests;
the local record contains only their independent SHA-256 identities, lengths,
and locators.

This is preparation work. Validation and sealing do not execute a check, verify
the custodian's access boundary, or authorize a protected dispatch. Every valid
result reports `PREPARED_NOT_AUTHORIZED`, with both
`protected_execution_authorized` and `custody_verified` false. An
`approval_ref` is retained as an auditable pointer only; the command cannot
establish that the referenced approval is genuine, current, or sufficient.

## Input schema: `g1-pack/3`

The input is one UTF-8 JSON object, limited to 256 KiB. Unknown,
case-variant, and duplicate keys are refused. The fixed fields are:

| Field | Required contract |
|---|---|
| `task_manifest`, `answer_manifest` | Distinct SHA-256 identity, positive byte length, and custodian locator; no embedded material |
| `custody` | Declares `unexposed_to_implementation_cases`, `no_protected_content`, `separate_answer_manifest`, and a custody-record reference. These are retained declarations, not verified facts. |
| `case_counts` | The acceptance strata: exactly 24 applicable, 16 inapplicable, and 8 underspecified cases. |
| `claim_kind_counts` | Capability coverage, recorded separately from the outcome strata: all five routes must be exercised and total 48 cases. The custodian may choose the mix; it must not be inferred from the acceptance strata. |
| `tool_contracts` | Exactly one entry for finite equivalence, finite instance, observed-rate invariance, solved monotonicity, and probabilistic-property routing. Each must name the currently shipped procedure, checker version, and SHA-256 of the complete selected registry entry. Free-text aliases, stale versions, and changed premises/quantifiers/outputs/limits are refused. |
| `progression` | The roadmap's observed-case gate: 23 of 24 applicable checks completed, at most one false refusal, zero inapplicable false certifications, every underspecified result names a missing premise, and zero invalid certifications. These are a bounded engineering gate, not a population error-rate estimate. |
| `execution` | An explicit resource-ceiling reference plus nonnegative provider ceilings. `approval_ref` may be empty; a nonempty reference still does not authorize execution here. |

An example uses synthetic manifest identities only:

```bash
newf --json g1 pack validate \
  --input docs/plans/artifacts/2026-09-13-g1-pack-contract/example-pack.json
```

## Immutable seal receipt

```bash
newf --json g1 pack seal \
  --input docs/plans/artifacts/2026-09-13-g1-pack-contract/example-pack.json \
  --out /secure-custodian-location/g1-pack.seal.json
```

Sealing hashes the exact metadata bytes and writes a `g1-pack-seal/1` receipt
through an atomic hard-link publication. An existing output is refused. The
receipt preserves the validation result and its scope restriction; it is not a
seal of the referenced task or answer material, an access-control test, a
provider invocation, or a protected evaluation result.

Inspect a retained receipt without rerunning or dispatching anything:

```bash
newf --json g1 pack inspect /secure-custodian-location/g1-pack.seal.json \
  --input docs/plans/artifacts/2026-09-13-g1-pack-contract/example-pack.json
```

Inspection strictly parses the receipt and reports `MATCH`, `MISMATCH`, or
`NOT_CHECKED` for its exact metadata-byte binding. `MATCH` proves only that the
receipt names those bytes and pack ID. It does not promote the custody,
authorization, or content assertions.

Before any protected G1 work, the operator must separately establish a concrete
authorization, actual custody boundary, and execution ceilings. Those acts are
outside this command and remain visible rather than being inferred from a JSON
field or a local receipt.
