# Dispatch log — preservation pilot, authorization revision 1

Written BEFORE any run output exists. Dispatched 2026-09-12, ~15:45 PT, by
the operator's session agent under the ceilings of `../PREPARATION.md` §7
(frozen at `1bc14da`).

## Delivery mechanism

Each run is one isolated background agent (fresh context, session model,
identical toolset). Packaging is NOT paraphrased into prompts: byte-exact
copies of the frozen inputs were staged at `/tmp/pp-arms/` and hash-verified
against the frozen package before dispatch; each run is instructed to read
only its own staged packaging file(s):

```text
314d10c7fa2d622e  baseline-engineering.md   (= frozen assessment-identity rev 2)
dcc6f4e16e449cda  contract.md               (= frozen review-contract.md)
4009a54a471c4a65  manuscript-baseline.md    (= frozen authored baseline)
cd79cc559e8dd0b4  bundle-e1.md              (byte-slice of frozen bundles file)
7092f8f7acfeea65  bundle-e2.md              (byte-slice of frozen bundles file)
423a0e8262b46467  bundle-m.md               (byte-slice of frozen bundles file)
```

Subject delivery: `git archive <REV> | tar -x -C $(mktemp -d)` followed by
`rm -rf <tmp>/docs/reviews` — no `.git`, hence no commit messages; the
denial list is enforced by deletion, symmetrically, not only by instruction.

## Launch order (frozen, §4)

| # | Cell | Revision | Packaging |
|---|------|----------|-----------|
| 1 | S3d-B | `3443a1094407364be2e02264ffa1ee1bb0f3efde` | baseline-engineering |
| 2 | S3d-M | `3443a1094407364be2e02264ffa1ee1bb0f3efde` | contract + bundle-e1 |
| 3 | S4d-M | `3443a1094407364be2e02264ffa1ee1bb0f3efde` | contract + bundle-e2 |
| 4 | S3c-B | `5dd08618f10bde9ea0eca2aa8b2d39426898155c` | baseline-engineering |
| 5 | S4c-M | `5dd08618f10bde9ea0eca2aa8b2d39426898155c` | contract + bundle-e2 |
| 6 | S4d-B | `3443a1094407364be2e02264ffa1ee1bb0f3efde` | baseline-engineering |
| 7 | M1d-B | `00c3c5808b8df5a5257a34a9ffc695d406f68294` | manuscript-baseline |
| 8 | M1c-M | `630e380b9584e378a1aa1360efe432d4651fa439` | contract + bundle-m |
| 9 | M1c-B | `630e380b9584e378a1aa1360efe432d4651fa439` | manuscript-baseline |
| 10 | S4c-B | `5dd08618f10bde9ea0eca2aa8b2d39426898155c` | baseline-engineering |
| 11 | M1d-M | `00c3c5808b8df5a5257a34a9ffc695d406f68294` | contract + bundle-m |
| 12 | S3c-M | `5dd08618f10bde9ea0eca2aa8b2d39426898155c` | contract + bundle-e1 |

Runs are launched as parallel background agents in this order. With one
observation per cell, the seeded order controls obvious sequencing effects
only; launch order equals the frozen order, but wall-clock completion order
is not controlled — recorded here as an environment note, symmetric across
arms.

Raw outputs land at `raw/run-NN-<cell>.md`. Aborted or timed-out runs are
retained as-is and never redispatched. Blinding happens after collection:
opaque labels `R01…R12` with a sealed mapping; the adjudicator (a fresh
isolated agent) receives only blinded reports, case specifications, the S3c
subtlety note, and the rubric.

## Transport note (recorded before any run output existed)

Run 7 (M1d-B): the first dispatch attempt failed with a malformed tool-call
error in the dispatch harness BEFORE any agent was created — zero model
exposure, zero work performed. The dispatch was re-issued once, identical
prompt bytes. This is transport repair, not a run retry; the 12-primary-
dispatch ceiling is intact (12 agents created, one each).

Dispatched agent IDs, in launch order:
22ad3aef (S3d-B), 9b76ba42 (S3d-M), c4748fa4 (S4d-M), 96625acf (S3c-B),
ed2f16c7 (S4c-M), 74566c8a (S4d-B), 62c764cc (M1d-B), 5d59c6b2 (M1c-M),
5cbd6327 (M1c-B), b1bebc1c (S4c-B), 78e22cc9 (M1d-M), 86088d9d (S3c-M).
