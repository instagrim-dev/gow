# Pilot 001 — preparation STOP (readiness not met)

Status: `prepared_not_captured`, **NOT READY**. Per RUNBOOK.md, a failed
readiness check is a stop, not an instruction to weaken the threshold, alter
the vocabulary, or manufacture completeness. No captures were produced; no
prompts were assembled; the manifest's null values remain null.

Revision 2 (2026-09-10, post-review): scoped the miner's conclusion,
qualified the degraded seven-family count, and replaced the two-way
disposition with the commissioned mapping-and-abstraction review
(`../review/MAPPING-REVIEW.md`). Revision 1 is preserved at `b0ad3ee`.

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
| failure_cohort | ready (provisional) | failure-side families = 7 (failure:2, partial_failure:5); partial_success:5 — see qualification below |
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
   So resolving the existing labels 1:1 would STILL mine zero candidates.

**Scope of the miner's conclusion.** `DerivingFixtureInvariantMiner`
proposes only a single `contains` predicate for a canonical id appearing in
the operators/preserves projection of ≥2 failure-side families; it is a
deterministic fixture, not fuzzy compression. Zero candidates therefore
means: *this fixture found no sufficiently repeated encoded operator or
preserved property.* It does **not** mean no shared failure structure
exists, nor that synonym mappings are the only way to expose it. Different
mechanisms can share a failure-relevant property without being aliases of
one another, and this miner cannot discover that relationship unless
someone has already encoded the shared property.

**Qualification of the seven-family count.** The stored clustering result
is `degraded`: twelve signatures became twelve singleton isolates
(`clusters.json` → `cluster_run.status = "degraded"`, `family_count = 12`,
all `member_count = 1`). The seven failure-side groups satisfy the numeric
gate, but separation caused by unresolved comparisons is not demonstrated
mechanistic diversity. Record as: **seven failure-side singleton groups
under a degraded comparison, pending semantic mapping review.** Any
approved mapping changes must be followed by reclustering before support is
counted again — it would be invalid to improve feature overlap while
continuing to claim the original seven support units.

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

The next bounded task is a **train-only mapping-and-abstraction review**
(commissioned in `../review/MAPPING-REVIEW.md`), not immediate corpus
replacement and not synonym consolidation as the only productive outcome.
The review must admit several relationship types:

| Relationship found in the source descriptions | Treatment |
|---|---|
| Different labels mean the same operation in the relevant scope | Approved aliases to one canonical concept |
| One operation is a component or specialization of another | Preserve identities; record the relationship |
| Different operations share a restriction, assumption, or dependency | Preserve identities; propose a shared property |
| The evidence does not establish a relationship | Leave unresolved |
| The methods genuinely differ on the property being tested | Preserve that distinction as potential counterevidence |

"These methods share property P" is not the same claim as "these methods
are the same method." Concretely: `es-02` lists both `congruence covering`
and `class union` on one note — evidence of a method and possibly one of
its constituent operations, not evidence the two labels are
interchangeable. And `es-12` uses the exact label `congruence covering`
under a `partial_success` outcome — it cannot help reach the failure-side
support threshold, but it is mandatory contrast evidence for any future
candidate involving covering and must not disappear from the review.

Review outputs route by kind: genuine synonyms become approved aliases in a
**new pinned vocabulary revision** (not a silent mutation of
`mechanism/v1`); shared properties of genuinely distinct methods are
recorded as **source-grounded interpretations/hypotheses** whose support
must then be mined, evaluated and challenged — never disguised as aliases.
A model may propose mappings, extract passages, identify counterexamples,
and draft shared-property hypotheses; it must not approve its own
interpretation as authoritative. Keep the review train-only and freeze its
decisions before revisiting target recovery: "makes B3 recover the target"
must not become the criterion for accepting a mapping. Either outcome
requires a **new protocol revision** (RUNBOOK.md): changed dataset or
vocabulary ⇒ new pinned hashes, fresh database, reclustering, re-run
readiness.

**Interpretation limit.** This STOP exposes missing semantic preparation —
the protocol selected this corpus before that readiness had been
established. It is not a negative result about the research thesis. Two
distinct future experiments follow from here: a **curated-feature pilot**
(reviewed mappings and challenged shared properties feed the deterministic
miner; an external proposer tests whether those features improve proposed
directions) and a **discovery pilot** (a model also proposes higher-order
shared properties from heterogeneous descriptions, and those
interpretations undergo independent grounding and challenge). The first is
a legitimate smaller milestone; it does not test the full claim that a
model can discover the failure invariant itself, and the current fixture
cannot stand in for that capability merely because its support calculation
is correct.

Nothing in this stop is a defect in the harness path: ingest, normalization,
signatures (12/12 + 1), clustering (degraded, recorded as such),
failure-space, mining, challenge (vacuous — zero candidates), split
definition, and the readiness report all executed and are recorded in this
directory.
