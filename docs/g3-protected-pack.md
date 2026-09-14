# G3 protected-pack preparation

`newf g3 pack` prepares a strictly content-free receipt for a future G3
evaluation. It binds separate task and answer manifest identities and declared
coverage; it never reads task inputs, sealed answers, candidate submissions,
commitments, observations, or receipts.

```bash
newf g3 pack validate --input metadata.json
newf g3 pack seal --input metadata.json --out g3-pack-seal.json
newf g3 pack inspect g3-pack-seal.json --input metadata.json
```

The metadata schema is `g3-pack/1`. It requires two distinct manifest
references (`task_manifest` and `answer_manifest`), a custody declaration, a
current composition/finite-checker contract, resource ceilings, and one
declared case for each route:

- faithful realization with objective met;
- faithful realization with objective missed;
- refuted action-menu selection;
- refuted missing precondition; and
- refuted unjustified equality.

Its exact public JSON shape is:

```json
{
  "schema": "g3-pack/1",
  "pack_id": "opaque-pack-id",
  "task_manifest": {"sha256": "...", "byte_length": 1, "locator": "protected/tasks/MANIFEST.json"},
  "answer_manifest": {"sha256": "...", "byte_length": 1, "locator": "protected/answers/MANIFEST.json"},
  "custody": {"task_author_exposure": "unexposed_to_implementation_cases", "implementer_access": "no_protected_content", "answer_separation": "separate_answer_manifest", "record_ref": "protected/custody.json"},
  "coverage": {"total": 5, "faithful_objective_met": 1, "faithful_objective_miss": 1, "menu_selection": 1, "missing_precondition": 1, "unjustified_equality": 1},
  "tool_contract": {"task_schema": "composition-task/1", "candidate_schema": "composition-candidate/1", "commitment_schema": "composition-commitment/1", "observation_schema": "composition-observation/1", "checker_version": "finite-equivalence-checker/1"},
  "execution": {"resource_ceiling_ref": "operator reference", "provider_call_ceiling": 0, "provider_spend_cents": 0, "approval_ref": ""}
}
```

All manifest digests are lower-case SHA-256 strings. The coverage counts must
be derived from saved observations; do not alter them to make this record
valid. A missing required route leaves the pack unsealed and records a failed
evaluation rather than a pass.

Those are declared coverage counts, not outcomes established by metadata
validation. The custodian must author task files as `composition-task/1`, seal
answers separately before a candidate is evaluated, and give the proposer only
the permitted task-facing material. A candidate submission must use
`composition-candidate/1` and bind the exact compact task SHA-256.

Every validation response and seal reports `PREPARED_NOT_AUTHORIZED`.
`g3 pack` cannot verify author independence or access isolation, authorize a
protected run, or upgrade a declared custody record into evidence. A valid
sealed pack becomes evidence only after an externally authorized custodian
procedure executes and grades the held material under its declared boundary.
