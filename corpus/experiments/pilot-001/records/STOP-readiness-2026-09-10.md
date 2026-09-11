# Pilot 001 — preparation STOP (readiness not met)

Status: `prepared_not_captured`, **NOT READY**. Per RUNBOOK.md, a failed
readiness check is a stop, not an instruction to weaken the threshold, alter
the vocabulary, or manufacture completeness. No captures were produced; no
prompts were assembled; the manifest's null values remain null.

## Execution facts

- prepared at: 2026-09-10 23:00 PT (2026-09-11T06:00Z), executable commit
  `5eea271` (see `executable-commit.txt`), source pin `a868dec` verified
  (train tree `da49ddcf…`, target blob `7ed727ab…`).
- gates: `go build/vet/test` green, `gofmt` clean, isolated binary built.
- identifiers (also in `runtime.json`):
  train `prb_01M27GTZHJN2V42W73DFVBY2SK`,
  target `prb_01M27GTZHW9NDHDGQ2RKZEQRY0`,
  holdout `hset_01M27GTZRCS648Z6DVZZRAWK74` (blinded).
- pre-capture DB snapshot: `pre-capture.db` (sha256 in `runtime.json`;
  retained locally, deliberately not committed — digest is the durable record
  pending an explicit data-retention decision).

## Readiness result (`readiness.json`)

| check | status | fact |
|---|---|---|
| failure_cohort | ready | failure-side families = 7 (failure:2, partial_failure:5); partial_success:5 |
| decisive_axis_resolution | ready (thin) | 12 signatures; resolved decisive claims = 2, unresolved = 83; only 1 signature has any resolved decisive content |
| completeness_admissions | info | accepted = 0 → absence-based verified violations unreachable (assessments degrade to unknown, honestly) |
| **surviving_invariants** | **blocked** | mining at min-support 2 produced **0 candidates**; B3 has no eligible guided target |
| withheld_target | ready | 1 signed target; resolved = 2, unresolved = 9 |
| recovery_criterion | info | recovery-rule/v1 under classify/v1, profile `f3334d08…` |

## Root cause (mechanical, from the snapshot — read-only queries)

Mining found no candidate because no canonical id recurs across ≥2
failure-side families. Two independent contributors:

1. **Vocabulary resolution is near-empty on the train corpus.** Of 85
   decisive-axis claims, 2 resolve (`modular decomposition`,
   `residue locality`, both on one note); `mechanism/v1` was seeded for the
   fixture corpora, not these notes.
2. **More fundamentally: no surface label — resolved or not — is shared by
   two failure-side families.** Every operator/preserves label occurs in
   exactly one family (`congruence covering` ×2 occurrences but one family).
   So resolving the existing labels 1:1 would STILL mine zero candidates:
   support requires the mapping review to adjudicate which distinct surface
   phrasings denote the same mechanism-level term (e.g. whether any of
   `congruence covering` / `class union` / `class assembly` /
   `local congruence filter` are one canonical operator — a research
   judgment, not a harness decision).

The full label frequency table is reproducible from `pre-capture.db`:

```sql
SELECT c.field_kind, c.surface_label, COUNT(DISTINCT cm.cluster_id) ff
FROM signature_field_claims c
JOIN cluster_members cm ON cm.signature_id = c.signature_id
JOIN signature_outcomes so ON so.signature_id = c.signature_id
WHERE so.class IN ('failure','partial_failure')
  AND c.field_kind IN ('preserves','operator')
GROUP BY 1,2 ORDER BY ff DESC;
```

## Disposition

This is exactly the Package-2 work the protocol's §1 checkboxes reserve for
review: reviewed vocabulary mappings (alias adjudication for the terms THIS
experiment's comparison axes require — not the full Erdős–Straus ontology)
and, if the mapping review concludes the notes genuinely describe disjoint
mechanisms, a corpus revision with justified scopes. Either outcome requires a
**new protocol revision** (RUNBOOK.md): changed dataset or vocabulary ⇒ new
pinned hashes, fresh database, re-run readiness.

Nothing in this stop is a defect in the harness path: ingest, normalization,
signatures (12/12 + 1), clustering, failure-space, mining, challenge (vacuous
— zero candidates), split definition, and the readiness report all executed
and are recorded in this directory.
