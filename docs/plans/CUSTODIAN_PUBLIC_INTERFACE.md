# Custodian public interface

This document is the data-only interface paired with
`CUSTODIAN_INTERFACE_BRIEF.md`. It contains no protected case material and no
development examples. Use only an operator-supplied pinned `newf` executable.

## Inputs common to every check

Each command needs an operator-supplied SQLite path, policy ID, obligation ID,
case label, executor label, and explicit resource reservation. A saved check
does not create an assessment, admit a rewrite rule, verify custody, or grant
authority. Retain the JSON input bytes and the complete command response.

```text
newf --json review check-finite \
  --db <custodian-db> --policy <policy-id> --obligation <obligation-id> \
  --case <opaque-case-id> --executor <custodian-id> \
  --max-assignments <0..65536> --input <finite-claim.json>

newf --json review check-finite-instance \
  --db <custodian-db> --policy <policy-id> --obligation <obligation-id> \
  --case <opaque-case-id> --executor <custodian-id> \
  --max-instances <0..4096> --input <finite-instance-claim.json>

newf --json review check-observations \
  --db <custodian-db> --policy <policy-id> --obligation <obligation-id> \
  --case <opaque-case-id> --executor <custodian-id> \
  --max-submissions <0..16384> --input <observation-claim.json>
```

Each input is exactly one UTF-8 JSON object, at most 1 MiB. Unknown,
case-variant, and duplicate keys are refused. A reservation refusal assesses no
prefix of the input.

## Finite equivalence input: `finite-claim/1`

Required top-level fields are `schema`, `kind`, `source_ref`, `statement`,
`domain`, `left`, and `right`. `kind` is `finite_equivalence`. `domain` has
`width` (1 through 8) and a closed `variables` list. An expression is exactly
one of `var`, `const`, or `op` with `args`; supported operators are `not`,
`neg`, `shl1`, `shr1`, `and`, `or`, `xor`, `add`, `sub`, and `mul`.

The result is exhaustive only within the declared finite domain. `REFUTED`
includes a counterexample; missing or invalid premises and exhausted allowance
remain blocked records.

## Finite instance input: `finite-instance-claim/1`

Use the finite fields above, set `kind` to `finite_instance`, and add
`assignments`. It is an ordered array of objects with a `values` array. Every
value has `var` and `value`; every declared variable must be assigned exactly
once per assignment.

Agreement returns `INSTANCE_EVIDENCE_ONLY`; it is never a domain-wide
equivalence certificate. A supplied disagreement can return `REFUTED` for the
declared universal claim.

## Observation input: `observation-claim/1`

Required top-level fields are `schema`, `kind`, `source_ref`, `statement`,
`binding`, and `observations`. `binding` specifies `population`, `ordering`,
`stopping_rule`, `budget_min`, and `budget_max`. Each ordered observation has
matching `conditions` (including a strictly increasing `budget`) and `instances`.
Each instance has an opaque `id` and ordered `submissions`; every submission has
`move` and explicit Boolean `success`.

Supported kinds are `observed_rate_invariance`, `solved_monotonicity`, and
`probabilistic_property`. Observation populations, ordering, stopping rules,
budgets, instance IDs, and trace prefixes must match the claim's declared
comparison. Probabilistic-property routing records `NOT_ASSESSED`; it is not a
probability result.

## Check retrieval and explicit assessment

Retrieve a saved record without rerunning it:

```text
newf --json review check-show --db <custodian-db> <check-id>
```

An assessment is a separate operator-authorized action. It must use the exact
saved check and its scoped criterion. Never infer an assessment from a command
exit code or from a successful storage write.

## Content-free G1 metadata and seal

After task and answer manifests have been separately sealed in custodian storage,
create a `g1-pack/3` JSON object containing only: pack ID; distinct task and
answer manifest SHA-256, byte length, and locator; custody declarations; the
24/16/8 case counts; all five claim-kind counts totaling 48; current tool
contracts; progression criteria; and execution ceiling references. It must not
embed tasks, answers, raw outputs, or transcripts.

```text
newf --json g1 pack validate --input <g1-metadata.json>
newf --json g1 pack seal --input <g1-metadata.json> --out <new-seal.json>
newf --json g1 pack inspect <new-seal.json> --input <g1-metadata.json>
```

The validator binds current tool procedures, checker versions, and full registry
entries. A valid result is always `PREPARED_NOT_AUTHORIZED`: it neither verifies
custody nor authorizes protected execution.
