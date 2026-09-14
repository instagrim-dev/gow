---
artifact_kind: prepared-future-custodian-public-interface
status: prepared_not_authorized
prepared_date_utc: '2026-09-14'
scope: future protected G1 evaluator only
supersedes: none; CUSTODIAN_PUBLIC_INTERFACE.md remains the frozen attempt-4 interface
---

# Custodian public interface v2

This is a preparation template. At dispatch, the operator must replace every
angle-bracket token with a frozen value, include the paired brief and pinned
executable, record their SHA-256 identities, and supply an explicit authorization.
The resulting instance is the only interface a custodian receives. This template
contains no protected case or answer material and does not itself authorize a
dispatch, spending, or assessment.


This file is complete for the stated data-only custody role. It defines the
only executable commands, input schemas, sealed private records, and
content-free metadata shape. A field or command absent from this document is
`interface_unrepresentable`; stop rather than inspect the executable or invent
it.

Only the commands below may operate on the supplied executable after the exact
integrity command in the paired brief. The storage root is `<dispatch-root>/protected`; the
database is `<dispatch-root>/protected/custodian.db`; executor is `<custodian-id>`.

## First command: frozen policy definition

Run this command verbatim before authoring a case. Retain its JSON output and
record the returned policy and obligation IDs. The five `--obligation` IDs map
to their respective `key` values below.

```text
<dispatch-root>/packet/newf --db <dispatch-root>/protected/custodian.db --json review policy \
  --key g1-v2-protected-usefulness --revision 1 \
  --decision-name "protected G1 v2 usefulness progression" \
  --owner operator --authority "<operator-approval-text>" \
  --scope "48 fresh custodian-authored cases across five typed claim routes on the pinned executable" \
  --case-budget 48 --attempt-budget 2 --provider-call-budget 0 \
  --obligation 'key=finite-equivalence-route;revision=1;requirement=execute finite_equivalence claims exactly as supplied;acceptance=certificate binds exact input bytes and verdict is consistent;applicability=cases declaring kind finite_equivalence;owner=operator;mandatory=true' \
  --obligation 'key=finite-instance-route;revision=1;requirement=execute finite_instance claims exactly as supplied;acceptance=instance evidence or refutation binds exact input bytes;applicability=cases declaring kind finite_instance;owner=operator;mandatory=true' \
  --obligation 'key=observed-rate-route;revision=1;requirement=execute observed_rate_invariance claims exactly as supplied;acceptance=result binds exact input bytes and declared comparison;applicability=cases declaring kind observed_rate_invariance;owner=operator;mandatory=true' \
  --obligation 'key=solved-monotonicity-route;revision=1;requirement=execute solved_monotonicity claims exactly as supplied;acceptance=result binds exact input bytes and verified trace prefixes;applicability=cases declaring kind solved_monotonicity;owner=operator;mandatory=true' \
  --obligation 'key=probabilistic-routing-route;revision=1;requirement=route probabilistic_property claims to explicit non-assessment;acceptance=NOT_ASSESSED recorded without an invented probability;applicability=cases declaring kind probabilistic_property;owner=operator;mandatory=true'
```

## Allowed check commands

For each authored input, substitute only the retained policy ID, its matching
obligation ID, and a unique opaque case ID. The input must remain under
`<dispatch-root>/protected/tasks/`.

```text
<dispatch-root>/packet/newf --db <dispatch-root>/protected/custodian.db --json review check-finite --policy <policy-id> --obligation <finite-equivalence-route-id> --case <opaque-case-id> --executor <custodian-id> --max-assignments <1..4096> --input <finite-claim.json>
<dispatch-root>/packet/newf --db <dispatch-root>/protected/custodian.db --json review check-finite-instance --policy <policy-id> --obligation <finite-instance-route-id> --case <opaque-case-id> --executor <custodian-id> --max-instances <1..4096> --input <finite-instance-claim.json>
<dispatch-root>/packet/newf --db <dispatch-root>/protected/custodian.db --json review check-observations --policy <policy-id> --obligation <observed-rate-route-id|solved-monotonicity-route-id|probabilistic-routing-route-id> --case <opaque-case-id> --executor <custodian-id> --max-submissions <1..16384> --input <observation-claim.json>
<dispatch-root>/packet/newf --db <dispatch-root>/protected/custodian.db --json review check-show --policy <policy-id> <check-id>
```

Every claim input is one UTF-8 JSON object no larger than 1 MiB. Unknown,
case-variant, and duplicate field names are rejected at every nesting level.
No command assesses a prefix after a resource reservation refusal.

For a successfully persisted check response, the opaque identity required by
`check-show` is the exact JSON field `check.ID`. Preserve the complete response
in protected storage, record `check.ID`, `check.Outcome`, `check.Blocker`, and
`receipt.certificate.Verdict` when present, then use that exact `check.ID` in
`check-show`. `persisted: true` confirms storage only; it does not make an
assessment or a certificate verdict. A command error or a response without a
persisted check has no `check.ID`; retain it as an interruption or blocked
receipt and do not fabricate an ID.

### Schema 1: finite equivalence

A `finite-claim/1` root has exactly: `schema`, `kind`, `source_ref`,
`statement`, `domain`, `left`, and `right`.

- `schema` is exactly `finite-claim/1`; `kind` is exactly
  `finite_equivalence`.
- `source_ref` and `statement` are nonempty text.
- `domain` has exactly `width` (integer 1 through 8) and `variables`
  (a nonempty array of distinct plain-text identifiers, each at most 128 bytes).
- `left` and `right` are expression objects. Each has exactly one form:
  `{"var":"name"}`, `{"const":nonnegative-integer}`, or
  `{"op":"operator","args":[expression]}` /
  `{"op":"operator","args":[expression,expression]}`.
  A `var` must name a declared variable. `args` is allowed only with `op`.
  The unary operators are `not`, `neg`, `shl1`, `shr1`; binary
  operators are `and`, `or`, `xor`, `add`, `sub`, `mul`.
  Maximum expression depth is 64 and maximum nodes per side is 4096.

```json
{"schema":"finite-claim/1","kind":"finite_equivalence","source_ref":"custodian:fresh","statement":"x plus zero equals x over this declared four-bit domain","domain":{"width":4,"variables":["x"]},"left":{"op":"add","args":[{"var":"x"},{"const":0}]},"right":{"var":"x"}}
```

A fully bound exhaustive agreement is a finite-domain certificate. A found
counterexample is `REFUTED`; an absent or invalid premise is blocked.

### Schema 2: finite instances

A `finite-instance-claim/1` root has the exact Schema 1 fields plus
`assignments`. It must use `schema: "finite-instance-claim/1"` and
`kind: "finite_instance"`. `assignments` is an array of 1 through 4096
objects. Each assignment has exactly `values`, an array containing exactly
one `{"var":"declared-name","value":nonnegative-integer}` for every declared
variable: no extra, missing, or duplicate variables. Assignments must be
distinct. Values must fit the declared width.

```json
{"schema":"finite-instance-claim/1","kind":"finite_instance","source_ref":"custodian:fresh","statement":"x plus zero agrees at two declared points","domain":{"width":4,"variables":["x"]},"left":{"op":"add","args":[{"var":"x"},{"const":0}]},"right":{"var":"x"},"assignments":[{"values":[{"var":"x","value":0}]},{"values":[{"var":"x","value":7}]}]}
```

Agreement is `INSTANCE_EVIDENCE_ONLY`, never an exhaustive-equivalence
certificate. A disagreement is `REFUTED`.

### Schema 3: observations

An `observation-claim/1` root has exactly `schema`, `kind`, `source_ref`,
`statement`, `binding`, and `observations`. It uses a `kind` of
`observed_rate_invariance`, `solved_monotonicity`, or
`probabilistic_property`.

- `source_ref` and `statement` are nonempty text and each at most 4096 bytes.
- `binding` has exactly `population`, `ordering`, `stopping_rule`,
  `budget_min`, and `budget_max`. The labels are nonempty and at most 256
  bytes; budgets are integers, nonnegative, and ordered.
- `observations` has 2 through 64 records. Each has exactly `conditions`
  and `instances`. `conditions` has exactly the matching three labels and
  one nonnegative `budget`; labels must equal the binding and budgets must
  strictly increase in input order and remain in the binding range.
- Each observation has at least one `instances` record. Instance `id` is
  nonempty, at most 128 bytes, unique within that observation, and the complete
  ID set must be identical across observations. Each instance has exactly
  `id` and `submissions`; submissions have exactly nonempty `move` (at
  most 1024 bytes) and explicit Boolean `success`. Across input there may be
  at most 4096 instance records and 16384 submissions.
- `solved_monotonicity` uses exactly two observations and needs an actual
  extension of every solved trace. `probabilistic_property` is an unconditional
  `NOT_ASSESSED` route: it does not inspect a sample or produce a probability
  conclusion. Do not use that route for an underspecified case, because absent
  `binding` or `observations` does not change the refusal's `NOT_ASSESSED`
  verdict. An underspecified case must instead use a nonprobabilistic
  observation claim with a genuinely absent `binding` or `observations`, which
  records `UNRESOLVED` and names the missing premise.

```json
{"schema":"observation-claim/1","kind":"observed_rate_invariance","source_ref":"custodian:fresh","statement":"The observed success rate is unchanged at the two declared budgets","binding":{"population":"instance a","ordering":"recorded order","stopping_rule":"declared budget","budget_min":1,"budget_max":2},"observations":[{"conditions":{"population":"instance a","ordering":"recorded order","stopping_rule":"declared budget","budget":1},"instances":[{"id":"a","submissions":[{"move":"m1","success":true}]}]},{"conditions":{"population":"instance a","ordering":"recorded order","stopping_rule":"declared budget","budget":2},"instances":[{"id":"a","submissions":[{"move":"m1","success":true},{"move":"m2","success":true}]}]}]}
```

## Frozen case allocation and case-state mapping

Use this exact cross-stratum allocation; it leaves no discretionary route or
stratum pairing:

| Stratum | Route and count | Required terminal result |
|---|---|---|
| Applicable (24) | 12 `finite_equivalence`, 4 `finite_instance`, 8 `probabilistic_property` | The finite routes use fully bound inputs. The probability routes expect `NOT_ASSESSED`; this is a correct terminal nonassessment, not a refusal to run the public route. |
| Inapplicable (16) | 8 `finite_instance`, 8 `solved_monotonicity` | Supply a typed premise defect so the retained verdict is `INAPPLICABLE` and names the defect. It must never certify the claim. |
| Underspecified (8) | 8 `observed_rate_invariance` | Omit one real required premise (`binding` or `observations`) so the retained verdict is `UNRESOLVED` and names that exact absence. |

This allocation totals the frozen 12/12/8/8/8 route composition and 24/16/8
strata. Do not repurpose a reservation refusal, malformed JSON rejected before
persistence, cancellation, or storage failure as an inapplicable or
underspecified case.

The G1 case-state and the review ledger outcome have separate meanings. Assign
`completed-valid` whenever the routed public command persists a receipt whose
exact input binding and terminal verdict match the sealed expected answer,
including `NOT_ASSESSED` for the probability route, `INAPPLICABLE` for an
inapplicable input, and `UNRESOLVED` that names the sealed missing premise. The
CLI records these three semantic non-certifications as `check.Outcome: "blocked"`;
that ledger word does **not** make the G1 case-state `blocked`. Reserve the G1
case-state `blocked` only for a resource stop, cancellation, process
interruption, or storage failure. Assign `false-certification` only if an
inapplicable case is certified or the receipt contradicts its exact input;
assign `refused-applicable` only if an otherwise applicable public route is
actually declined. For a matching run under this allocation, all 48 case states
are `completed-valid`.

## Private records sealed before checks

Create these directories: `tasks/`, `answers/`, and `receipts/` beneath
the storage root. Before the first check, create the following separate JSON
records. They are protected material and never copied to the return root.

**`tasks/MANIFEST.json`** is one object with exactly
`{"schema":"g1-v2-task-manifest/1","entries":[...] }`. It has exactly 48
entries, each with exactly: `case_id`, `input_path`, `input_sha256`,
`input_bytes`, `stratum`, `claim_route`, `author`, `authored_at_utc`,
`method`, `exposure`, `derived_from`.

- `case_id` is unique and opaque; `input_path` is a relative `tasks/` path;
  the SHA-256 is lower-case 64 hex characters; `input_bytes` is positive.
- `stratum` is one of `applicable`, `inapplicable`, `underspecified`;
  counts are exactly 24/16/8. `claim_route` is one of `finite_equivalence`,
  `finite_instance`, `observed_rate_invariance`, `solved_monotonicity`,
  `probabilistic_property`; route counts are exactly 12/12/8/8/8.
- `author` is exactly `<custodian-id>`; `method` is exactly
  `authored-fresh-in-custodian-context`; `exposure` is exactly `none`; and
  `derived_from` is exactly the empty string.

**`answers/MANIFEST.json`** is one object with exactly
`{"schema":"g1-v2-answer-manifest/1","entries":[...] }`. It has the same
48 unique case IDs, each with exactly `case_id`, `answer_path`,
`answer_sha256`, `answer_bytes`, `sealed_before_execution`. The path is a
relative `answers/` path, digest/byte fields follow the task rules, and
`sealed_before_execution` is true.

Each referenced answer is one object with exactly `schema`, `case_id`,
`expected_certificate_verdict`, `expected_ledger_outcome`,
`expected_check_outcome`, `rationale`, `authored_at_utc`, `author`,
`sealed_before_execution`. It uses
`schema: "g1-v2-expected-answer/1"`, the matching case ID,
`author: "<custodian-id>"`, and `sealed_before_execution: true`.
`expected_certificate_verdict` is the exact checker value expected from the
protected response: finite equivalence uses one of `HOLDS_ON_DECLARED_DOMAIN`,
`REFUTED`, `UNRESOLVED`, or `INAPPLICABLE`; finite instance uses one of
`INSTANCE_EVIDENCE_ONLY`, `REFUTED`, `UNRESOLVED`, or `INAPPLICABLE`; observation
uses one of `HOLDS_AT_COMPARED_POINTS`, `REFUTED`, `UNRESOLVED`, `INAPPLICABLE`,
or `NOT_ASSESSED`; use `NONE` only when the expected response has no certificate.
`expected_ledger_outcome` is exactly `completed` or `blocked` and records the
expected CLI ledger word separately. `expected_check_outcome` is the
custodian's intended G1 case classification—one of `completed-valid`,
`refused-applicable`, `false-certification`, `blocked`, or `not-executed`.
For a persisted CLI response, compare the expected certificate with
`receipt.certificate.Verdict` and compare the expected ledger word with
`check.Outcome`; then apply the frozen case-state mapping above. `rationale` is
nonempty and explains the declared expectation using only the case's typed data.

**`case-provenance.jsonl`** has exactly 48 UTF-8 JSON lines, one per case,
each with exactly: `schema`, `case_id`, `input_sha256`, `input_bytes`,
`answer_sha256`, `answer_bytes`, `stratum`, `claim_route`, `author`,
`authored_at_utc`, `method`, `exposure`, `derived_from`. `schema` is
`g1-v2-case-provenance/1`; values must agree with both manifests.

After creating those records and before any check, write
`PRE_EXECUTION_SEAL.json` with exactly `schema`, `task_manifest_sha256`,
`task_manifest_bytes`, `answer_manifest_sha256`, `answer_manifest_bytes`,
`case_provenance_sha256`, `case_provenance_bytes`, `sealed_at_utc`, and
`first_check_not_started`. Use `schema: "g1-v2-pre-execution-seal/1"` and
`first_check_not_started: true`. Its three identities must refer to the exact
bytes already written. Hash all three records and the seal; enter their
identities in `ACCESS_LOG.md` before running a check.

## Content-free metadata commands

After both manifests are sealed, create `g1-metadata.json` under the storage
root and run only:

```text
<dispatch-root>/packet/newf --json g1 pack validate --input <g1-metadata.json>
<dispatch-root>/packet/newf --json g1 pack seal --input <g1-metadata.json> --out <dispatch-root>/protected/g1-pack.seal.json
<dispatch-root>/packet/newf --json g1 pack inspect <dispatch-root>/protected/g1-pack.seal.json --input <g1-metadata.json>
```

The metadata root has exactly these fields: `schema`, `pack_id`,
`task_manifest`, `answer_manifest`, `custody`, `case_counts`,
`claim_kind_counts`, `tool_contracts`, `progression`, and `execution`.
Use this complete shape, replacing only the manifest identities with actual
sealed values:

```json
{"schema":"g1-pack/3","pack_id":"<dispatch-id>","task_manifest":{"sha256":"<actual lower-case SHA-256>","byte_length":<actual positive bytes>,"locator":"custodian://g1-v2/task-manifest"},"answer_manifest":{"sha256":"<actual lower-case SHA-256>","byte_length":<actual positive bytes>,"locator":"custodian://g1-v2/answer-manifest"},"custody":{"task_author_exposure":"unexposed_to_implementation_cases","implementer_access":"no_protected_content","answer_separation":"separate_answer_manifest","record_ref":"custodian://g1-v2/access-log"},"case_counts":{"applicable":24,"inapplicable":16,"underspecified":8},"claim_kind_counts":{"finite_equivalence":12,"finite_instance":12,"observed_rate_invariance":8,"solved_monotonicity":8,"probability_out_of_scope":8},"tool_contracts":[{"kind":"finite_equivalence","procedure":"finite-claim-check/1","version":"finite-equivalence-checker/1","registry_sha256":"95458ca6a49e875342bd7e00bb9da26fcb558df439accf94b3055bffa2313375"},{"kind":"finite_instance","procedure":"finite-instance-claim-check/1","version":"finite-equivalence-checker/1","registry_sha256":"fde8ffc552a3b54c2a4e60192b63d6d7a0ffe0e229c6a29f5bbcb0b1e215f30e"},{"kind":"observed_rate_invariance","procedure":"observation-claim-check/1","version":"measure-checker/1","registry_sha256":"e6c686e5ba1e311b283d164e1b15b1f01aa93a4e94e0ec96e3fb8cec5bfb2ca5"},{"kind":"solved_monotonicity","procedure":"observation-claim-check/1","version":"measure-checker/1","registry_sha256":"94c5b431e32ca5cf04761f8e3a0e963fd7c0e4f9c00dd04433c9b139571e7377"},{"kind":"probabilistic_property","procedure":"observation-claim-check/1","version":"measure-checker/1","registry_sha256":"3b3945ca6a85347ac1ae72e81032a6f33a5ee5e608c31d35576e3394d02c60a5"}],"progression":{"min_applicable_completed":23,"max_applicable_false_refusals":1,"max_inapplicable_false_certifications":0,"require_inapplicable_failed_condition":true,"require_underspecified_missing_premise":true,"max_underspecified_definite_conclusions":0,"max_invalid_certified":0},"execution":{"resource_ceiling_ref":"<dispatch-id>:offline-zero-spend","provider_call_ceiling":0,"provider_spend_cents":0,"approval_ref":"<operator-approval-ref>"}}
```

Metadata is content-free: do not place task, answer, raw output, expected
answer, or case result content in it. A MATCH seal binds only exact metadata
bytes; it does not prove custody or independently authorize execution.

## Content-free return packet

Write `G1_V2_RETURN_PACKET.json` in the return root using this content-free
shape. Replace angle-bracket values only; do not add fields.

```json
{"schema":"g1-v2-return-packet/1","dispatch_id":"<dispatch-id>","release_revision":"<pinned-revision>","executable_sha256":"<pinned-sha256>","policy_id":"<policy-id>","obligation_ids":{"finite-equivalence-route@1":"<id>","finite-instance-route@1":"<id>","observed-rate-route@1":"<id>","solved-monotonicity-route@1":"<id>","probabilistic-routing-route@1":"<id>"},"manifest_identities":{"task_manifest":{"sha256":"<lower-case sha256>","byte_length":<positive>},"answer_manifest":{"sha256":"<lower-case sha256>","byte_length":<positive>},"case_provenance":{"sha256":"<lower-case sha256>","byte_length":<positive>},"pre_execution_seal":{"sha256":"<lower-case sha256>","byte_length":<positive>}},"metadata_identity":{"sha256":"<lower-case sha256>","byte_length":<positive>},"seal_identity":{"sha256":"<lower-case sha256>","byte_length":<positive>},"custody_limitations":["<content-free limitation>"],"blocked_actions":["<content-free action or empty list>"],"stratum_outcome_counts":{"applicable":{"completed-valid":<nonnegative>,"refused-applicable":<nonnegative>,"false-certification":<nonnegative>,"blocked":<nonnegative>,"not-executed":<nonnegative>},"inapplicable":{"completed-valid":<nonnegative>,"refused-applicable":<nonnegative>,"false-certification":<nonnegative>,"blocked":<nonnegative>,"not-executed":<nonnegative>},"underspecified":{"completed-valid":<nonnegative>,"refused-applicable":<nonnegative>,"false-certification":<nonnegative>,"blocked":<nonnegative>,"not-executed":<nonnegative>}},"elapsed_seconds":<nonnegative>,"completion_state":"<completed|interface_unrepresentable|execution_interrupted|resource_exhausted|verification_blocked>"}
```

Each stratum object totals its frozen population (24, 16, or 8). If metadata or
a seal was never created because the interface was terminally unrepresentable,
use a zero-byte identity with an all-zero SHA-256 and include the reason in
`blocked_actions`; never replace the identity object with a raw explanation.
No raw, case-specific, expected-answer, result, or score content is allowed.
Write `G1_V2_RETURN_PACKET.sha256` containing the SHA-256 of its exact JSON
bytes. The final response reports only the completion state, both return paths,
and that digest.
