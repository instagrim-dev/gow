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
