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
newf g4 calibrate --episode-pack open-episodes.json --resource-ceiling resource.json --out calibration-receipt.json
newf g4 calibrate-procedure --episode-pack open-episodes.json --procedure procedure.json --out calibration-receipt.json
newf g4 arm-preflight --resource-ceiling resource.json --h0-snapshot h0.json --h1-snapshot h1.json --hg-snapshot hg.json
newf g4 execute --manifest g4-lite-metadata.json --episode-pack episodes.json \
  --resource-ceiling resource.json --h0-snapshot h0.json --h1-snapshot h1.json \
  --hg-snapshot hg.json --generation-procedure procedure.json --out protected/execution-receipt.json
```

The final input schema is `g4-lite-pack/3`. It is one UTF-8 JSON object of at
most 256 KiB. Unknown, duplicate, and case-variant fields are refused. The
manifest contains only identity references and declarations:

| Field | Required contract |
|---|---|
| `episode_manifest`, `answer_manifest` | Separate lower-case SHA-256 identities, positive byte lengths, and locators. No protected bytes appear in the record. |
| `calibration_manifest` | Identity of the separately retained open-case sensitivity calibration. It documents pre-seal calibration and does not permit protected-case tuning. |
| `generation_procedure_manifest` | Identity of the frozen `g4-lite-calibration-procedure/1` used for open calibration and protected authoring. The executor receives those bytes, verifies this identity, and checks the protected pack's family prefix, catalog branching, and rewrite depth against it. |
| `custody` | Declares `unexposed_to_implementation_cases`, `no_protected_content`, `separate_answer_manifest`, and a custody record. These remain declarations rather than independently verified facts. |
| `population` | Exactly 24 episodes: 12 `history_informative`, 6 `history_low_value`, and 6 `history_misleading`, with at least two construction families. |
| `arms` | Three distinct frozen controller snapshots: H0, H1, and HG. They bind one model configuration digest, one tool-catalog digest, the current finite checker, and one per-arm task-directed resource-ceiling identity. H1 requires a recorded `non_implementer` review. Custody work is outside that ceiling and must later be metered. |
| `run_design` | Either the default three fixed seeds with a seed-manifest identity, or one disclosed budget-constrained run. Two-run or ad hoc designs are refused. |
| `endpoint` | The exact objective, including the declared target cost, under the same **task-directed** resource cap. Separately metered custody is excluded from this endpoint and retained for the full-cost decision record. |
| `spending_rule` | Exact run-summed arithmetic: HG must exceed H1 by at least `3 × runs`; H1 minus HG on low-value/misleading cases is a net run-summed loss of at most `1 × runs`; qualifying family advantages are positive run-summed HG advantages on informative episodes only. The rule also requires no invalid certification and HG at least H0. |
| `execution` | The same task-directed resource-ceiling locator and nonnegative provider ceilings. `approval_ref` is an audit pointer only. |

### Exact manifest field names

The final manifest decoder rejects unknown keys. In particular, the custody
declarations are **values** under these four keys; the declaration words are
not JSON field names:

```json
"custody": {
  "episode_author_exposure": "unexposed_to_implementation_cases",
  "implementer_access": "no_protected_content",
  "answer_separation": "separate_answer_manifest",
  "record_ref": "protected/custody.json"
}
```

Every content reference uses `sha256`, `byte_length`, and `locator`. For the
disclosed deterministic one-run design, `run_design` must use
`runs_per_cell: 1`, `seed_policy: "single_run_budget_constrained"`, one
nonempty `budget_constraint_ref`, and **omit** `seed_manifest`. The
`execution.resource_ceiling_ref` must equal `arms.resource_ceiling.locator`.
The required `population` keys are `total`, `history_informative`,
`history_low_value`, `history_misleading`, and `min_families` (with values
24, 12, 6, 6, and at least 2, respectively).
The strict open episode input is `shaping-pack/1`: its top-level keys are
`schema`, `label`, `provenance`, and `episodes`; every episode requires `id`,
`stratum`, `family`, `start`, `variables`, `catalog`, and `target_cost`.
Included history records require `start`, `rules_applied`, `final_cost`,
`target`, `completed`, and `endpoint`; `history[].start` is the canonical
rendered expression **string**, whereas the episode `start` is a finite
expression JSON object. `history[].endpoint` is likewise an oracle-verdict
string (for example, `HOLDS_ON_DECLARED_DOMAIN`), not a Boolean.

Observed metadata uses `schema: "g4-lite-observed-metadata/1"`, the exact
pre-execution seal's `pre_execution_seal_sha256` and
`pre_execution_seal_bytes`, and three content references named
`arm_execution_manifest`, `resource_ledger_manifest`, and
`result_grid_manifest`. Each of those three values uses the same `sha256`,
`byte_length`, `locator` shape. `g4 pack bind-execution` validates and binds
those byte identities; it does not establish that their private contents
conform to the frozen design.

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
of the supplied episode, resource, runtime-identity, and (for `/3`) frozen
generation-procedure artifacts against the identities in the final manifest.
It uses the procedure only to enforce its protected family namespace, catalog
branching, and rewrite-depth controls; it does not inspect answers or verify
custody. The episode artifact is `shaping-pack/1` and
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

`g4 calibrate` accepts an **open** 24-episode 12/6/6 `shaping-pack/1`, a
candidate resource vector, and a new `--out` path. It first finds each H0
minimum expansion allowance without evaluating H1 or HG, under the declared
state, term, rule-application, candidate, and cancellation limits. Every
independent binary-search probe gets a fresh rule/candidate allowance. The
median positive H0 minimum becomes the fixed proposed expansion allowance for
one disclosed three-arm diagnostic.

The output path receives an atomic
`g4-lite-sensitivity-calibration-receipt/1` record. It retains every H0
reference probe with its bounded search result, alongside minimums,
unreachable episodes, proposed allowance, and any reference-phase stop. If the
diagnostic starts, it is retained as a distinct nested resource receipt,
including cells, decisions, work records, and any stop reason. A
blocked phase returns a command error after its receipt is saved; it cannot
support a favorable sensitivity conclusion (`CALIBRATION_BLOCKED` or
`DIAGNOSTIC_BLOCKED`).

No positive H0 minimum exists when every start already meets its target
(`H0_COMPLETION_CEILING`), every task is unreachable inside the expansion cap
(`H0_UNREACHABLE_WITHIN_CAP`), or the open pack mixes those two conditions
(`H0_NO_POSITIVE_CALIBRATION_SAMPLE`). Those outcomes do not run a diagnostic
because no median positive allowance exists. `COMPLETION_CEILING` requires
**all three** arms to complete every open episode. The stated open sensitivity
criterion is remaining comparator headroom: a positive H0 calibration sample
and at least one H1 noncompletion (`COMPARATOR_HEADROOM_REMAINS`). HG may
complete all 24 open episodes while H1 remains below 24; that still has useful
comparator headroom. `SENSITIVITY_CRITERION_UNMET` means the positive H0
sample exists but H1 is saturated without an all-arm ceiling. Every status is
diagnostic only and never grants protected-dispatch readiness. Record both
successful and failed calibrations; do not tune a fresh protected pack from
completion counts. Whenever its three-arm diagnostic completes, the legacy
receipt also records the open margin and informative-family headroom fields;
the status label itself remains the older comparator-only classification.

For a prospective protected cycle, use the versioned
`g4-lite-calibration-procedure/1` input with `g4 calibrate-procedure`. It
freezes the exact public pack used for calibration, at least two concrete
resource vectors, one predeclared primary vector, and the authoring controls:
minimum rewrite depth, branching alternatives, cost-neutral enabling steps,
independently verified targets, and an independently specified history method.
It also declares non-overlapping open/protected family prefixes and the
required exposure boundary. The procedure must declare that it is frozen
before protected authoring and that protected targets will not be adjusted
after targeted performance. The command refuses an open pack whose identity,
family namespace, catalog alternatives, or expression depth violate that
procedure.

Its `/2` receipt retains every declared resource choice, including a blocked
attempt, plus two open-only feasibility diagnostics for each completed
three-arm run: the maximum possible HG-over-H1 advantage (`24 - H1`), which
must be at least three for the one-run margin to be attainable, and the number
of informative families containing an H1 noncompletion, which must be at
least two for the family condition to be attainable. These diagnostics do not
require H0 failure or HG-over-H0 success and do not add rules to an already
sealed or executed protected batch. They are a design check for a fresh
procedure, not a spending, grading, custody, or dispatch decision.

## Required sequence

1. Freeze the generation/resource procedure before protected authoring. Run it
   against an implementation-exposed, family-separated open pack and retain
   every resource response. This fixes the protected construction controls and
   does not disclose a future protected pack.
2. Freeze the resource specification at a pinned release. Export H0, H1, and
   HG runtime identities from that executable, record the three content-free
   runtime-identity artifacts, and make `g4 arm-preflight` pass. This is a
   design-stage reference; it does not include future episode or answer hashes.
3. The custodian authors the fresh protected episodes and separately seals the
   episode, answer, calibration, arm-snapshot, resource, and seed artifacts.
4. The custodian builds and seals the final `g4-lite-pack/3` manifest with
   those real identities. It must not invent future hashes or mutate a prior
   final-pack seal.
5. Under a separate execution authorization, the custodian executes the full
   3 × 24 × runs grid and meters custody separately.
6. The custodian records `g4-lite-observed-metadata/1` and runs
   `bind-execution`. The observed record names the exact pre-execution seal
   plus content-free arm-execution, resource-ledger, and result-grid manifest
   identities. The command validates both inputs and their seal binding.
7. A substantive grader separately compares actual arm/resource records with
   the frozen manifest, assesses the protected answers and result quality,
   verifies resource compliance and spending arithmetic, and returns only a
   content-free judgment. `bind-execution` does not perform that comparison,
   establish chronology, verify custody, or grant authority.

Every final-manifest validation response and seal reports
`PREPARED_NOT_AUTHORIZED`, `protected_execution_authorized: false`, and
`custody_verified: false`. An `approval_ref`, locally matching seal, or
provider budget cannot change those values. A screen outcome remains bounded
to its sealed batch and cannot become a G4 confirmatory claim or funding
authority through this command.
