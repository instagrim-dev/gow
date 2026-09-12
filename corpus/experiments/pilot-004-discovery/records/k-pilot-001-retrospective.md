# (k-restricted) Retrospective challenge on pilot-001's new invariants

Recorded: 2026-09-12. Pipeline-space (SGO on pilot-001).

## Bounded scope

This is a **retrospective research artifact**, not a re-execution of
pilot-001. No B0/B3 captures are being regenerated; no
`manifest.json` values are being modified; the historical
`STOP-readiness-2026-09-10.md` record stands as the pilot's declared
outcome. The only new durable state in `.newf/pilot-001/newf.db` is
additional invariant revisions and challenge assessments, which are
labelled and timestamped 2026-09-12 and are additive rather than
supersessive.

The purpose of this record is to answer: **given the (a‴-B)
extension developed independently on pilot-004 and validated on M7,
what would pilot-001's substrate look like at the invariant-mining
and challenge stages today?** The answer is a research finding, not
a claim about what pilot-001's B0/B3 captures would have produced.

## Historical STOP reason (pilot-001)

`corpus/experiments/pilot-001/records/readiness.json`, verbatim:

```
"check": "surviving_invariants",
"status": "blocked",
"detail": "no surviving candidate invariant; mine and run the
          challenge campaign first — B3 has no eligible guided target"
```

`corpus/experiments/pilot-001/records/mining.json`, verbatim:

```
"miner_version": "invariant/v1+671cf5d70015c34a",
"candidate_count": 0,
"candidates": null
```

Pilot-001 was stopped because the pre-extension miner produced zero
candidates, which made the surviving_invariants readiness gate
unsatisfiable and left B3 with no guided-context material.

## Retrospective mining + challenge with (a‴-B)

Extended miner (`invariant/v1+712e5351092a6ecd`) on the SAME failure
space + cluster run:

| candidate | assoc | support | challenge disposition |
|---|---|---|---|
| equals(locality, local) | contrast_observed | 3 | **surviving** |
| equals(construction, constructive) | contrast_observed | 3 | weaken (success-preserving confirmed) |

Verified via `sqlite3 .newf/pilot-001/newf.db`:

```
inv_01M2BDMGB64VR9RVFM2WMP67PD  equals(locality, local)         surviving
inv_01M2BDMGB64VR9RVFM2YT46HPV  equals(construction, constructive) weaken
```

**One surviving invariant is now present on pilot-001** for the
first time in its history.

The `equals(construction, constructive)` weakening replays the M7
result exactly: success-preserving probe confirms a partial_success
family that preserves the predicate. The atlas-truthfulness of this
weakening was confirmed independently on M7 (`records/m7-atlas-truth-check.md`).

## Retrospective readiness (informational only, not a claim of readiness)

Re-running `experiment readiness` against pilot-001 today:

```
ready: False
  failure_cohort              ready
  decisive_axis_resolution    ready
  completeness_admissions     info
  surviving_invariants        ready      surviving=1
  withheld_target             ready
  recovery_reachability       blocked    1 of 1 target(s) cannot receive
                                          a positive match under
                                          recovery-rule/v1
  recovery_criterion          info       recovery-rule/v1 under classify/v3
```

Two distinct observations:

1. **The specific gate that stopped pilot-001 historically
   (`surviving_invariants: blocked → no surviving candidate
   invariant`) is now `ready: surviving=1`.** The (a‴-B) extension
   resolves the exact failure that produced pilot-001's STOP.

2. **A different gate now blocks readiness**
   (`recovery_reachability: blocked`), which was not in pilot-001's
   original readiness manifest at all. This is a schema-evolution
   effect — new readiness checks have been introduced since
   pilot-001's preparation date (the classify profile is now v3,
   not v1). Any legitimate re-execution of pilot-001 would need to
   freshly certify against the current schema, which is a separate
   question outside the scope of this record.

## Cross-pilot comparison

For completeness, pilots 2 and 3 had `surviving_invariants: ready`
in their original readiness (3 recurring Contains-preserves
invariants each). The (a‴-B) extension adds 2 stronger
(`contrast_observed`) invariants to their revisions but does not
change their historical readiness disposition on the
`surviving_invariants` gate. Their downstream substrate would have
had richer material to challenge and target, but this is a smaller
delta than pilot-001's zero→one surviving change.

**Pilot-001 is the unique case among persisted ES pilots where the
(a‴-B) extension resolves the specific historical STOP reason.**
This is not a claim that (a‴-B) would have produced any particular
B0/B3 outcome — the pilot's B3 arm would still need to receive the
new surviving invariant as guided context and the recovery
classification would still need to be independently assessed.

## Consequence

This closes the retrospective portion of path (k) at pilot-001. The
retrospective finding is:

- The (a‴-B) extension retroactively resolves pilot-001's historical
  STOP gate.
- Pilot-001's B0/B3 execution remains a distinct research question
  requiring a fresh readiness certification under the current schema
  and separate authorization from the operator, because that would
  produce new capture artifacts that are epistemically distinct from
  the historical `prepared_not_captured` record.

## Verification tier

- Historical STOP + mining zero-count: **SGO** (verbatim reads of
  pilot-001 record files).
- New invariant candidates + challenge dispositions: **SGO** (direct
  sqlite reads + deterministic challenge command output).
- Retrospective readiness: **SGO** (direct CLI output).
- The claim that (a‴-B) "would have" changed pilot-001's outcome
  had it been available: **not claimed here**. Only the
  substrate-level claim (the gate that stopped it is now
  satisfiable) is asserted.
