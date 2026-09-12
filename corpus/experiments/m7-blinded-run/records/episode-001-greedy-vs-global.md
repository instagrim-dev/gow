# Episode 001 — greedy-3 local truncation vs map-guided global search

Recorded: 2026-09-12
Episode: `epi_01M2BED4J0GN24NV0P1GJTZRXQ` (v41 protocol, `completed`,
`revision_credit: earned` — code-derived)
Store: `.newf/m7/newf.db` (live M7 corpus; not checked in — full typed export
at [`episode-001-greedy-vs-global.json`](episode-001-greedy-vs-global.json))
Provenance: the "minimum persuasive demonstration" requested by
`docs/reviews/2026-09-12-structural-semantic-epistemic-review.md`
(§Positioning): freeze map/action/prediction/scoring → externally checked
outcome → revise map → commit a DIFFERENT action → score the next outcome.

## What this episode is and is not

**Is:** the first operator-run closed loop through the v41 episode substrate
against a live map, with both outcomes decided by the exact-integer
Erdős–Straus checker (`erdos-straus-witness/v1`, subject `domain-goal`,
strength `reproducible`), predictions frozen before computation, and the
revision's credit derived by code from the second observation.

**Is not:** evidence about the Erdős–Straus conjecture (both actions are
elementary, known techniques; p=1000033 is not a mathematically interesting
instance), and not yet the *comparative* demonstration against an unmapped
baseline arm — this is one arm, map-guided, showing the loop closes and the
accounting works. Claim strength: the protocol demonstrably orders, checks,
and scores a real decision sequence; nothing here promotes the map's
invariants beyond their existing `surviving` (model-judgment) state.

## Preregistration honesty

The instance was fixed by a deterministic rule authored **before** any
computation: *p = smallest prime ≥ 10⁶ with p ≡ 1 (mod 24)* (the QR-hard
residue class from the corpus's failure geometry). The operator had not
computed p, the greedy expansion, or any feasibility check when the step-1
commitment was persisted; the rule leaves no room for instance steering.
Both step-2 search parameters (x-range bound, acceptance rule) were likewise
committed before execution. The v41 triggers enforce what prose cannot:
preregistration fields are frozen, observations append once, the step-2
commitment was refused until a revision existed, and scores are re-derived
at the persistence boundary.

## The loop

| Step | Committed action | Predicted | Checked outcome | Score |
|---|---|---|---|---|
| 1 | A1 — Fibonacci–Sylvester greedy expansion of 4/p truncated at 3 unit fractions: the *local, constructive, identity-carried* posture that all four surviving invariants of frozen map `ivr_01M2BCJZBSVP2XSCFNTBHWCJKH` flag as failure-side | `witness-valid` | `witness-invalid` — residue 1/217073791956662567555644065945239226028624401; checker shows both sides: 4xyz = …988 vs p(yz+xz+xy) = …987 | **miss** |
| revision | `epr_01M2BEF1Y1YQQV1Q4QC6VT34EG` — map ref moved from the frozen revision to revision+`epo_01M2BEE0DR8R8MXAXQN703TE2W`: the locality boundary is now *instance-checked* inside the episode, not just corpus-statistical; next action must violate `locality=local` | — | — | — |
| 2 | A2 — global divisor-structure search: for x from ⌈p/4⌉, reduce (4x−p)/px to a′/b′, find divisor d of b′² with d ≡ −b′ (mod a′), take y=(b′+d)/a′, z=(b′+b′²/d)/a′ | `witness-valid` | `witness-valid` — (1000033, 250009, 83339083442, 718489947739347664906), identity exact | **hit** |

**Code-derived verdict: `revision_credit: earned`** — step 2 was observed a
hit after a recorded revision whose committed action differed structurally
from step 1.

## The instructive contrast

Both actions started at the **same locality point** x = 250009. Greedy's
local rounding chose y = 83339083433 and missed the identity by 1 part in
10³⁸ — a miss no floating-point check would catch, and exactly why the
checker is exact-integer. The global move consulted the divisor lattice of
b′² (found d = 29 satisfying the congruence), chose y nine larger, and is
exact. The difference between miss and hit was not effort or range — it was
*which structure the action consulted*, which is the map's `locality=local`
axis made concrete at a single instance. This mirrors, at toy scale, the
paper-space N2a geometry (local moves stall; the boundary is structural),
without claiming to add evidence for N2a itself
(see `../../pilot-004-discovery/records/CURRENT-CLAIMS.md`).

## Full cost accounting

| Item | Cost |
|---|---|
| Wall-clock, preregistration → completion | **97 s** (preregistered 18:34:48Z → final observation 18:36:25Z, per persisted timestamps) |
| Compute | 2 python3 runs, < 1 s CPU total (sympy `isprime`/`factorint`, exact `Fraction` arithmetic) |
| Model/provider calls | **0** — zero provider spend; every verdict is code-owned |
| CLI invocations | 5 (`preregister`, `observe`, `revise`, `commit`, `observe`) + 1 `show` export |
| Records persisted | 1 episode, 2 commitments, 2 observations, 1 revision (all immutable) |
| Human/operator time | ~10 min authoring commitments and notes |

The map's *marginal* cost inside this episode was one refused posture (don't
re-roll a local identity) — it cost nothing and redirected step 2 to the
structure that worked. Whether that redirect beats an unmapped baseline arm
at matched cost is the **next** experiment (comparative, two-arm), not a
claim this episode makes.

## Reproduction

```bash
export NEWF_DB=.newf/m7/newf.db   # materialise via corpus/experiments/m7-blinded-run/run.sh
./bin/newf episode show --episode epi_01M2BED4J0GN24NV0P1GJTZRXQ
# or, from the checked-in export:
python3 -m json.tool corpus/experiments/m7-blinded-run/records/episode-001-greedy-vs-global.json
```

The witness identity itself is independently checkable anywhere:
4·250009·83339083442·718489947739347664906 =
1000033·(83339083442·718489947739347664906 + 250009·718489947739347664906 +
250009·83339083442).
