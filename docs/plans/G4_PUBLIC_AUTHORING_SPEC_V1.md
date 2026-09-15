---
artifact_kind: prepared-future-g4-public-authoring-spec
status: prepared_not_authorized
prepared_date_utc: '2026-09-14'
scope: public inputs for a future protected G4-lite `/3` custodian
---

# G4 public authoring specification v1

This is the complete public syntax and fixed-contract supplement for the G4
custodian. It contains no protected episode, answer, route, calibration-case,
or execution-result material. The operator supplies it with the paired
custodian brief, public interface, G4 contract, frozen procedure, calibration
receipt, resource ceiling, arm identities, and `operator-bindings.json`.
Missing or changed public bytes are `interface_unrepresentable`.

## Protected episode syntax

The episode file is one UTF-8 `shaping-pack/1` JSON object of at most 1 MiB.
Its exact top-level keys are `schema`, `label`, `provenance`, and `episodes`.
Every episode requires `id`, `stratum`, `start`, `variables`, `catalog`, and
`target_cost`; `family` and `history` are optional. The only strata are
`history_informative`, `history_low_value`, and `history_misleading`. The final
population is exactly 12, 6, and 6 respectively, with at least two informative
families.

Each expression is exactly one of:

```text
{"var":"<plain-identifier>"}
{"const":<nonnegative-integer>}
{"op":"<operator>","args":[<one-or-two-expressions>]}
```

Allowed unary operators are `not`, `neg`, `shl1`, and `shr1`. Allowed binary
operators are `and`, `or`, `xor`, `add`, `sub`, and `mul`. Variables are plain
nonempty identifiers (letters, digits, underscore; not beginning with a
digit), appear in the episode's 1–3 variable list, and are evaluated as
4-bit words. Expressions admit at most 64 nesting levels, 4,096 tree nodes,
and 128 bytes per identifier. Every history item supplies `start`,
`rules_applied`, `final_cost`, `target`, `completed`, and `endpoint`.

The complete admissible catalog is:

```text
double-not  neg-neg       add-zero      sub-zero      xor-zero
xor-self-zero  or-self    and-self      mul-one       mul-zero
not-intro   add-comm      xor-comm
```

No other catalogue name is valid. Catalog rules are independently admitted by
the finite checker at execution; inclusion does not establish that an authored
episode has a desired route or result.

## Operator bindings

The operator supplies one content-free, schema-shaped
`operator-bindings.json` before protected authoring. Its values are copied
into the final manifest; the custodian must not choose or alter them. This is
an operator packet contract: `g4 pack validate` independently validates the
corresponding manifest values, while dispatch integrity binds this file's
exact bytes.

```json
{"schema":"g4-public-operator-bindings/1","model_config_sha256":"<64-lowercase-hex>","tool_catalog_sha256":"<64-lowercase-hex>","checker_version":"finite-equivalence-checker/1","h1_review_ref":"<content-free-reference>","h1_reviewer_role":"non_implementer"}
```

The supplied H0/H1/HG runtime identities provide the three controller IDs and
exact snapshot artifact identities. The supplied resource ceiling provides the
single primary resource vector. The operator-retained open calibration receipt
is represented to the custodian only by the content-free
`calibration-manifest-reference.json`; its identity is the
`calibration_manifest` reference. The full receipt is not a custodian input.
The supplied generation procedure is the
`generation_procedure_manifest` reference.

## Final manifest skeleton

The operator supplies this skeleton with all public references and bindings
filled from the frozen packet. The custodian fills only its own pack ID,
protected-artifact identities and locators, custody-record reference, and the
single disclosed budget-constraint reference. It does not invent or tune any
fixed value.

```json
{"schema":"g4-lite-pack/3","pack_id":"<custodian-pack-id>","episode_manifest":{"sha256":"<custodian>","byte_length":<positive>,"locator":"<custodian>"},"answer_manifest":{"sha256":"<custodian>","byte_length":<positive>,"locator":"<custodian>"},"calibration_manifest":{"sha256":"<supplied-calibration>","byte_length":<supplied>,"locator":"<supplied>"},"generation_procedure_manifest":{"sha256":"<supplied-procedure>","byte_length":<supplied>,"locator":"<supplied>"},"custody":{"episode_author_exposure":"unexposed_to_implementation_cases","implementer_access":"no_protected_content","answer_separation":"separate_answer_manifest","record_ref":"<custodian>"},"population":{"total":24,"history_informative":12,"history_low_value":6,"history_misleading":6,"min_families":2},"arms":{"h0":{"controller_id":"<supplied-H0>","snapshot":{"sha256":"<supplied>","byte_length":<supplied>,"locator":"<supplied>"}},"h1":{"controller_id":"<supplied-H1>","snapshot":{"sha256":"<supplied>","byte_length":<supplied>,"locator":"<supplied>"}},"hg":{"controller_id":"<supplied-HG>","snapshot":{"sha256":"<supplied>","byte_length":<supplied>,"locator":"<supplied>"}},"model_config_sha256":"<operator-binding>","tool_catalog_sha256":"<operator-binding>","checker_version":"finite-equivalence-checker/1","resource_ceiling":{"sha256":"<supplied>","byte_length":<supplied>,"locator":"<supplied>"},"custody_outside_ceiling":true,"h1_review_ref":"<operator-binding>","h1_reviewer_role":"non_implementer"},"run_design":{"runs_per_cell":1,"seed_policy":"single_run_budget_constrained","budget_constraint_ref":"<custodian>"},"endpoint":{"kind":"exact_objective_within_same_task_directed_resource_cap/1","includes_target_cost":true,"same_task_directed_resource_ceiling":true},"spending_rule":{"max_invalid_certified":0,"min_hg_over_h1":3,"max_hg_loss_low_and_misleading":1,"min_difference_families":2,"require_hg_at_least_h0":true,"decision_arithmetic":"run_summed_exact/1","control_loss_arithmetic":"net_control_stratum_run_summed/1","family_advantage_arithmetic":"informative_positive_run_summed/1","task_directed_resources_only":true},"execution":{"resource_ceiling_ref":"<same-supplied-resource-locator>","provider_call_ceiling":0,"provider_spend_cents":0,"approval_ref":"<operator-authorization-reference>"}}
```

Replace every placeholder with valid JSON before invoking `g4 pack validate`.
The skeleton is a public construction aid, not an authorization or a grading
claim.
