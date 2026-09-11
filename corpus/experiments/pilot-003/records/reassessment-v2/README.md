# Pilot 003 — post-hoc classify/v2 reassessment (accounting only)

Explicitly **post hoc**: executed after the captures, the automatic v1
result, the model review, and closure. It is a separate accounting exercise
with its own experiment identity (`exp_01M27XY3DJ3032CVN6MYPANZ3F`, profile
`classify/v2`; the pinned primary remains
`exp_01M27QA4137EYA7S1FPTF2Q0SM` under `classify/v1`). It replaces nothing.

## Result

Same sealed inputs (b0-raw, b3-consumed), same budgets 8/8:

| Arm | Proposals | Recovered | Decisive | Unknown |
|---|---:|---|---:|---:|
| B0 undirected | 7 | no | 0 | 7 |
| B3 invariant-guided | 6 | no | 0 | 6 |

Conclusion: `inconclusive`. All four v1 `decisive_no` rows degraded to
`unknown`, confirming the EXECUTION-RESULT caveat: this frozen corpus
declares no completeness, so under the corrected missing-data contract no
recorded-subset conflict against it is decisive — the v1 decisive negatives
were exactly the recorded-description differences the contract now refuses
to promote.

## What this does and does not say

- It says the corrected rule behaves as documented on real captured data,
  and that the automatic layer's honest ceiling on this corpus (no
  completeness, out-of-vocabulary wording) is abstention in both
  directions.
- It does NOT resolve whether P10's prose satisfies the intended mechanism
  criterion (that is the external human review's question), does NOT
  change the closed run's conclusion, and any vocabulary change that would
  make P10 comparable remains prohibited as outcome-selected.
