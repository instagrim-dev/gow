# ESR negative-path control — fixture integration smoke test (expected abstention)

## Recorded conclusion

**Fixture integration smoke test: expected abstention.** The pipeline completed
using research-only curator annotations. No eligible failure-side support was
available, invariant-guided generation was not exercised, and both generated
baseline proposals were structurally unresolved. Experiment and comparison
outputs remained inconclusive without claiming recovery or decisive
non-recovery. **This run supports the tested empty/unresolved-input handling
paths; it does not validate populated invariant discovery, challenge coverage,
budget exhaustion, or comparative research effectiveness.**

This is the **negative-path control**. The complementary positive control (a
synthetic fixture with known ground truth that actually exercises the discovery
path) lives at `../2026-09-10-esr-positive-control/`.

## What this run does and does not support

| Reported observation | Supports | Does NOT establish |
|---|---|---|
| No eligible failures → 0 mined candidates | fixture path did not manufacture failure support | correct mining on a populated failure-space |
| 0 candidates → 0 B3 proposals | B3 did not invent targets when none existed | correct challenge / survival / guided generation |
| Two proposals compare as `unknown` | unresolved inputs not coerced to decisive match/negative | correct recovery detection when comparisons are possible |
| `inconclusive` / `incomparable` | uncertainty survived the reported path | correct interpretation across populated / budget-exhausted cases |
| Unannotated quarantine note rejected | fixture normalization precondition enforced (format/admission) | a substantive verification gate against an unproved math claim with a well-formed payload |

A guard cannot be validated by never reaching it. With `candidate_count: 0`,
this run does not exercise the challenge "unconfirmed attacks leave the
hypothesis unchanged" protection, and with two total proposals and no attempted
overrun it does not demonstrate budget enforcement.

## Two independent blockers (not isolated by this run)

1. **No eligible failures** → no supported failure-invariant task. The
   failure-space materialized as 7 `partial_success` + 4 `unknown`, **zero
   `failure`** families.
2. **No interpretable mechanism features** → no grounded structural comparison.
   All content axes (assumption, operator, preserves, auxiliary_object, breaks,
   representation) resolved to **0 canonical IDs**.

Because both co-occur, this run cannot attribute the empty candidate set to
either alone. It also cannot substantiate that the reported families are
*mechanistically distinct*; with unresolved content axes, clustering separation
may reflect inability to compare rather than demonstrated difference.

## Governance

The collection is `benchmark_training_eligible: false` and
`historical_holdout_eligible: false` (see `corpus/research/.../sources.json`).
Blinded mode addresses the absence of a chronological claim; it does **not**
grant benchmark eligibility. **The collection remained benchmark-ineligible. We
used it for a research-only execution test and do not interpret the output as an
eligible benchmark result.**

## Run manifest

- commit: `760e54a661493dc28e1ea4686c76607b0d9213ef`
- provider: `fixture` (deterministic; no model training occurred)
- mode: `blinded` (recovery-rule/v1, profile classify/v1)
- train problem: `prb_01M26SR297SH09DMYK82HBNE9G` — 11 curator notes **loaded
  into the train-side workspace** (not "trained")
- withheld target problem: `prb_01M26SR29JJGSP2TKM3KMBZMF8` — `esr-11`
  (Ghermoul, polynomial families) **withheld** for assessment (not "quarantined")
- holdout set: `hset_01M26SSYHF7CNJ62TBRE70K59P`
- experiment: `exp_01M26ST4EV1MPRAC4GCNE0XN4P`, run `run_01M26ST4EVTZ0XK037DNJ62SFC`
- budgets: proposal **configured 8**; evaluation **configured 40**
- conclusion: `inconclusive`

### Per-arm (execution vs research status)

| arm | execution_status | research_status | proposals | recovered | decisive | unknown | evals_consumed |
|---|---|---|---|---|---|---|---|
| b0_undirected | completed | skipped_no_eligible_invariants (empty) | 0 | false | 0 | 0 | 0 |
| b1_semantic_summary | completed | executed_baseline (unresolved) | 1 | false | 0 | 1 | 1 |
| b2_brainstorm | completed | executed_baseline (unresolved) | 1 | false | 0 | 1 | 1 |
| b3_invariant_guided | completed | skipped_no_eligible_invariants | 0 | false | 0 | 0 | 0 |

`execution_status` / `research_status` are **proposed reporting fields** (not
current schema) to keep `completed` from reading as "the intervention was
delivered." B3 was correctly *inhibited*, not *exercised*: no eligible failure
cohort, no candidate invariants, no surviving invariants, no guided proposals.
Comparing B0 vs B3 tests reporting for two empty arms; B2 vs B3 tests an
unresolved fixture proposal against an unavailable treatment.

## Split manifest

Train (loaded into workspace): esr-01, esr-02, esr-03, esr-04, esr-05, esr-06,
esr-07, esr-08, esr-09, esr-10, esr-13.
Withheld target: esr-11.
Excluded (no normalization payload, by corpus design): esr-12 (quarantine).

## Artifacts

- `experiment.json` — full experiment view
- `compare-b0-b3.json`, `compare-b2-b3.json` — comparison outputs
- `failure-space.json` — materialized failure-space (0 failures; unresolved axes)

## Reproduce

```bash
export NEWF_DB=/tmp/newf-esr/newf.db
C=corpus/research/erdos-straus-2026-09-10
newf init "…train…" --slug esr-train
newf init "…target…" --slug esr-target --new-problem
# train = notes/* + exploratory/esr-09,esr-10 ; target = exploratory/esr-11
newf ingest <train sources> --recursive --problem $TRAIN
newf ingest $C/exploratory/esr-11-*.md --problem $TARGET
newf normalize --all --problem $TRAIN ; newf normalize --all --problem $TARGET
newf mechanism signature --problem $TRAIN ; newf mechanism signature --problem $TARGET
newf cluster build --problem $TRAIN
newf failure-space build --problem $TRAIN
newf invariants mine --problem $TRAIN          # candidate_count: 0 (no failures)
newf challenge --all --problem $TRAIN           # nothing to challenge
newf experiment define --problem $TRAIN --target-problem $TARGET --name esr-holdout
newf experiment run --problem $TRAIN --arms b0_undirected,b1_semantic_summary,b2_brainstorm,b3_invariant_guided --proposal-budget 8 --evaluation-budget 40
newf experiment compare --problem $TRAIN
```
