# (i) Mining ↔ CMA convergence — enum-axis miner replicates on all persisted ES pilots

Recorded: 2026-09-12. Pipeline-space (SGO on 4 corpora).

## Question

Does the (a‴-B)-extended enum-axis miner reproduce useful invariants
on pilot corpora that predate the extension, or was the M7 finding
corpus-specific?

## Method (SGO)

Ran the extended miner (`--emit-posture-axes`) against each persisted
ES train corpus (m7, pilot-001, pilot-002, pilot-003), all using the
default `DerivingFixtureInvariantMinerWithPostureAxes` (miner_version
`invariant/v1+712e5351092a6ecd`). Compared to each corpus's recorded
original mining pass (miner_version `invariant/v1+671cf5d70015c34a`,
which lacked posture-axis emission).

## Result: perfect replication across all four persisted ES corpora

| Corpus | Original mining | Extended re-mining | Delta |
|---|---|---|---|
| M7 | 3 candidates (all `recurring` Contains) | 5 candidates | +2 `contrast_observed` |
| pilot-001 train | **0 candidates** | 2 candidates | **+2 `contrast_observed`** |
| pilot-002 train | 3 candidates (all `recurring` Contains) | 5 candidates | +2 `contrast_observed` |
| pilot-003 train | 3 candidates (all `recurring` Contains) | 5 candidates | +2 `contrast_observed` |

The two new invariants are IDENTICAL across all four corpora, with
identical support and contrast counts:

- `equals(locality, local)`: support=3, contrast_violating=5/5, `contrast_observed`
- `equals(construction, constructive)`: support=3, contrast_violating=4/5, `contrast_observed`

## The pilot-001 finding is the strongest

Pilot-001's original mining record (`corpus/experiments/pilot-001/records/mining.json`)
verbatim:

```
"miner_version": "invariant/v1+671cf5d70015c34a",
"revision": 1,
"candidate_count": 0,
"candidates": null
```

The pre-extension miner told pilot-001's operators there was nothing
to challenge, nothing to frontier against — **zero invariants** in a
corpus that today reveals two `contrast_observed` invariants (the
strongest association state available). This is not a marginal
extension; it is a **systematic blind spot closed**.

The pilot-001 record is the strongest evidence because the
before/after delta is total: from "no invariants discoverable" to
"two contrast-observed invariants across three failure-side families
each." Pipeline downstream stages — challenge, frontier generation,
evaluation — could not have engaged pilot-001 in any research-
meaningful way given zero mining output; today they can.

## Pilots 2 and 3 exhibit the same structural gap

Pilots 2 and 3 both recorded 3 `recurring` Contains-preserves
invariants and were driven through subsequent challenge and frontier
stages against those recurring targets. The re-mining today shows
that on each of those corpora, TWO ADDITIONAL invariants with
STRONGER association status (`contrast_observed` vs `recurring`)
were present in the data all along. Both pilots' subsequent research
activity was blind to those invariants.

## Verification tier and reproducibility

- **SGO tier.** All four corpus queries are direct sqlite reads +
  deterministic mining passes. The extended miner is offline and
  deterministic; running the same command against the same DB
  produces the same mining revision (subject to ULID timestamps for
  the revision ID). The candidate predicates, support values, and
  contrast counts are reproducible.
- Reproduction: from repo root,
  ```
  for p in pilot-001 pilot-002 pilot-003; do
    prb=$(sqlite3 .newf/$p/newf.db "SELECT id FROM problems WHERE slug LIKE '%train';")
    ./bin/newf --db .newf/$p/newf.db invariants mine \
      --problem "$prb" --emit-posture-axes
  done
  ```
- Note: each run creates a new invariant revision in the target DB.
  This does not destroy or supersede historical revisions; downstream
  interpretation heads continue to point at whatever revision the
  original pilot cited unless a subsequent challenge or frontier
  command elects the new revision.

## Interpretation

Two coupled findings:

1. **The pre-extension `DerivingFixtureInvariantMiner` had a
   systematic corpus-independent blind spot** on enum-axis posture
   structure. The blind spot was not corpus-idiosyncratic — it
   presented identically on M7, pilot-001, pilot-002, and pilot-003.
   The (a‴-B) extension closes this blind spot deterministically.

2. **The enum-axis pattern is corpus-independent within the ES
   domain.** All four independently-classified atlases produce the
   same two enum-axis invariants with the same support/contrast
   counts. This is what one would expect if the posture axes are
   picking up a genuine structural property of the ES failure space
   rather than a classification artifact of a particular atlas.

## Consequence for the research substrate

The extended miner is now a viable "discovery routing" step for
paper-space research. When a paper-space CMA argument uncovers a
structural pattern, the extended miner can be pointed at the same
corpus to check whether the pattern manifests as a support/contrast-
observable invariant in the pipeline state. When the pipeline is the
only writer (as in pilot-001, with zero prior invariants), the
extended miner can surface starting-point invariants that the paper-
space researcher would otherwise have to hypothesize from scratch.

The convergence hypothesis (paper-space CMA vs pipeline-space mining)
is **confirmed on the ES domain**. The generalization to other
domains remains open — pilot-005-relational, if materialized, would
be the natural next replication target.

## Notes on the cross-pilot replication

- The extended miner produces **only 2 candidates** on pilot-001
  because pilot-001's corpus lacks the Contains-preserves annotations
  that produce the 3 additional `recurring` candidates on pilots 2, 3,
  and M7. This is a **corpus-annotation** difference, not a mining
  logic difference. The enum-axis path is more robust to annotation
  gaps because posture fields (locality, construction, uncertainty)
  are enum-valued and mandatory in the atlas normalize schema.

- Pilots 2, 3, and M7 all reproduce the exact same 3 recurring
  Contains-preserves invariants because they share substantially
  the same ES atlas material. The stability of these three across
  corpora is itself a signal that they represent genuine ES failure-
  space structure and not a mining artifact.

- The predicate_fingerprint values for `equals(locality, local)` and
  `equals(construction, constructive)` are identical across all four
  corpora (SGO: verified by cross-corpus sqlite queries), which is
  expected — the fingerprint hashes only the predicate, not the
  corpus.

## Consequences for the closure scorecard

- Path (i) closes: convergence hypothesis confirmed on all persisted
  ES corpora. Recorded as SGO.
- Follow-on observation: the mining engine can retroactively enrich
  historical pilots that were closed with insufficient invariant
  yield. Not proposed as an execution here.
- Path (g) — escape H3 — remains the highest-value next research
  target (unchanged priority).
