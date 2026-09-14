---
artifact_kind: prepared-future-custodian-public-interface
status: prepared_not_authorized
prepared_date_utc: '2026-09-13'
scope: future protected G1 evaluator only
supersedes: none; CUSTODIAN_PUBLIC_INTERFACE.md remains the frozen attempt-4 interface
---

# Custodian public interface v2

This is the data-only executable interface template paired with
`CUSTODIAN_INTERFACE_BRIEF_V2.md`. Before launch, the operator must replace the
policy-command placeholder with the frozen command text and pin the resulting
two-file instance and executable by SHA-256. This template contains no
protected case or answer material. It does not authorize a provider call,
protected dispatch, or an assessment.

## Command allowlist

After the authorization's one exact executable-SHA-256 check, the evaluator
may run only the following supplied executable commands. The frozen instance
of this interface provides the exact `review policy` command and fixes every
placeholder, including the storage root, policy definition, and resource
reservations.

```text
newf --db <custodian-db> --json review policy <exact operator-supplied flags>

newf --json review check-finite \
  --db <custodian-db> --policy <policy-id> --obligation <obligation-id> \
  --case <opaque-case-id> --executor <custodian-id> \
  --max-assignments <0..4096> --input <finite-claim.json>

newf --json review check-finite-instance \
  --db <custodian-db> --policy <policy-id> --obligation <obligation-id> \
  --case <opaque-case-id> --executor <custodian-id> \
  --max-instances <0..4096> --input <finite-instance-claim.json>

newf --json review check-observations \
  --db <custodian-db> --policy <policy-id> --obligation <obligation-id> \
  --case <opaque-case-id> --executor <custodian-id> \
  --max-submissions <0..16384> --input <observation-claim.json>

newf --json review check-show --db <custodian-db> --policy <policy-id> <check-id>

newf --json g1 pack validate --input <g1-metadata.json>
newf --json g1 pack seal --input <g1-metadata.json> --out <new-seal.json>
newf --json g1 pack inspect <new-seal.json> --input <g1-metadata.json>
```

The evaluator may create and hash JSON files only inside the approved storage
root. It may not call any other executable subcommand or inspect the supplied
binary by another tool. A requested action outside this list is a stop
condition, not an invitation to infer an alternative interface.

## Check input contract

Each check input is one UTF-8 JSON object, at most 1 MiB. Unknown,
case-variant, and duplicate keys are refused. A reservation refusal assesses
no prefix of the supplied input.

| Schema and `kind` | Required fields | Deterministic result boundary |
|---|---|---|
| `finite-claim/1`, `finite_equivalence` | `source_ref`, `statement`, `domain`, `left`, `right`; `domain.width` is 1–8 and `domain.variables` is closed | Exhaustive only over the declared finite domain. A counterexample is `REFUTED`; missing/invalid premises and exhausted reservations are blocked. |
| `finite-instance-claim/1`, `finite_instance` | finite fields plus ordered `assignments`; each assignment has exactly one value for every declared variable | Agreement is `INSTANCE_EVIDENCE_ONLY`, never a domain-equivalence certificate. A supplied disagreement may be `REFUTED`. |
| `observation-claim/1`, `observed_rate_invariance`, `solved_monotonicity`, or `probabilistic_property` | `source_ref`, `statement`, `binding`, `observations`; binding names population, ordering, stopping rule, `budget_min`, and `budget_max`; observations contain matching conditions, instances, and ordered submissions | Population, order, budgets, IDs, and solved prefixes must bind. A probability claim is `NOT_ASSESSED`, not a probability conclusion. |

Finite expressions are exactly `var`, `const`, or `op` with `args`; supported
operators are `not`, `neg`, `shl1`, `shr1`, `and`, `or`, `xor`, `add`, `sub`,
and `mul`. Observation submissions have `move` and an explicit Boolean
`success`.

## Content-free `g1-pack/3` metadata contract

The metadata is exactly one UTF-8 JSON object, at most 256 KiB. It has these
top-level fields and no others:

| Field | Exact required shape |
|---|---|
| `schema`, `pack_id` | `schema` is `g1-pack/3`; `pack_id` is nonempty and at most 256 bytes. |
| `task_manifest`, `answer_manifest` | Each has lower-case 64-character `sha256`, positive `byte_length`, and nonempty `locator`; identities must differ. Contents never appear in this metadata. |
| `custody` | `task_author_exposure` is `unexposed_to_implementation_cases`; `implementer_access` is `no_protected_content`; `answer_separation` is `separate_answer_manifest`; `record_ref` is nonempty. These are declarations, not verified facts. |
| `case_counts` | Exactly `applicable: 24`, `inapplicable: 16`, and `underspecified: 8`. |
| `claim_kind_counts` | Exactly five nonzero route fields—`finite_equivalence`, `finite_instance`, `observed_rate_invariance`, `solved_monotonicity`, and `probability_out_of_scope`—summing to 48. This mix is independent of the outcome strata. |
| `tool_contracts` | Exactly one entry per executed route, each with `kind`, `procedure`, `version`, and a lower-case 64-character `registry_sha256` matching the pinned executable's selected registry entry. |
| `progression` | `min_applicable_completed: 23`, `max_applicable_false_refusals: 1`, `max_inapplicable_false_certifications: 0`, `require_inapplicable_failed_condition: true`, `require_underspecified_missing_premise: true`, `max_underspecified_definite_conclusions: 0`, and `max_invalid_certified: 0`. |
| `execution` | Nonempty `resource_ceiling_ref`; nonnegative `provider_call_ceiling` and `provider_spend_cents`; `approval_ref` may be empty but never authorizes anything through metadata alone. |

`g1 pack validate`, `seal`, and `inspect` bind this content-free metadata and
its exact bytes. They never read task or answer contents. A `MATCH` inspection
or `PREPARED_NOT_AUTHORIZED` result proves neither custody nor dispatch
authority.

## Outcome recording

The evaluator records each case as `completed-valid`, `refused-applicable`,
`false-certification`, `blocked`, or `not-executed` under the separately
approved acceptance matrix. A blocked or not-executed case is not a refusal.
`matches_expected` is a custodian diagnostic field, not a case state and not a
substitute for certificate validity. No stratum arithmetic is reported until
the collection is complete; an invalid custody or interface breach takes
precedence over any partial outcome.
