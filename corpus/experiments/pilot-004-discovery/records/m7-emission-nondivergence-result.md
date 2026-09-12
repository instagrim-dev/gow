# (m7-emission-divergence) Result — NON-ISSUE; both posture-axis candidates were emitted, and their divergent outcomes are a live real-data instance of correction 7's CMA statement

**Status**: executed 2026-09-12. Retracts the "(m7-emission-divergence)"
task from the scorecard's next-action list as a **non-divergence**.
Confirms correction 7 at the empirical (real-data) tier.

## Bottom line

There is **no divergence** to investigate. M7's `(a‴-B)` miner run
(revision 2) emitted **both** posture-axis candidates flagged by
`(l-population-validation)`:

- `equals(locality, local)` — ordinal 2, → **surviving**
- `equals(construction, constructive)` — ordinal 4, → **weaken**

The earlier framing "M7 persisted only `equals(locality, local)` at
`surviving`, when `equals(construction, constructive)` was also
emission-eligible" was **not** an emission problem. It was an
artifact of a state-filtered query: I looked only at rows with
`state='surviving'` and missed the co-persisted `weaken` row.

The divergence in **final state** between the two candidates is
exactly the pattern that correction 7's CMA statement predicts,
now observed at real-data tier on M7's train corpus.

## Method

SGO query against `.newf/m7/newf.db`:

- `invariant_revisions`: two mining runs on the same M7 train
  problem (`prb_01M2B8ZMHRCJN4YVH4CP6SNXTD`) at the same cluster
  run. Revision 1 (17:00 UTC, `min_support=2`, `candidate_count=3`)
  is the base miner; revision 2 (18:03 UTC, `candidate_count=5`)
  is the `(a‴-B)` posture-extension miner.
- `candidate_invariants` × `invariant_predicates` × `invariant_current_state`
  full join, no state filter.
- `invariant_challenges` × `invariant_state_transitions` per
  candidate for the probe-level detail.

## Live evidence

### Revision 2 (`(a‴-B)`) emitted 5 candidates

| Ord. | Predicate | assoc_status | contrast_viol / elig | Final state |
|---:|---|---|---:|---|
| 0 | `contains(preserves, identity_carried_solvability)` | recurring | 0/5 | proposed *(re-emit of Rev-1)* |
| 1 | `contains(preserves, confined_to_quadratic_nonresidues)` | recurring | 0/5 | proposed *(re-emit of Rev-1)* |
| 2 | **`equals(locality, local)`** | contrast_observed | **5/5** | **surviving** |
| 3 | `contains(preserves, class_union_construction)` | recurring | 0/5 | proposed *(re-emit of Rev-1)* |
| 4 | **`equals(construction, constructive)`** | contrast_observed | **4/5** | **weaken** |

Ordinals 0, 1, 3 duplicate Revision 1's fingerprints, so they inherit
their earlier-revision `surviving` state via the fingerprint-canonical
identity and are not re-challenged in Revision 2 (their
`invariant_current_state` shows `proposed` because Rev-2's rows are
distinct persistence identifiers but represent the same predicate).
The two **new** predicates ordinals 2 and 4 are the posture-axis
extension's contribution; both go through challenge in Revision 2.

### Probe-level challenge detail

Both candidates receive **the same three code-owned probes**
(`known-counterexample`, `success-preserving`, `bias-critique`).
Two of the three probes return identically:

| Probe | ord. 2 (locality=local) | ord. 4 (construction=constructive) |
|---|---|---|
| known-counterexample | unconfirmed (contrast_observed doesn't refute this way) | unconfirmed (same reason) |
| bias-critique | unconfirmed (recomputed support = 3 ≥ 2) | unconfirmed (recomputed support = 3 ≥ 2) |
| **success-preserving** | **unconfirmed** — "all 5 eligible success-side member(s) decisively checked; none preserves the predicate" | **confirmed** — "1 success-preserving family(ies)" |

The single differentiating signal is the `success-preserving`
probe's outcome — `completedNegative` on the sPrev=0 candidate,
`Confirmed` on the sPrev=1/5 = 0.2 candidate. This is the exact
mechanism the `(l-coh-mechanism-test)` fixture test operationalized
synthetically at HEAD `20058a4`.

### `contrast_violating_num` / `contrast_eligible_den` reconciliation

- Ordinal 2, `equals(locality, local)`: `contrast_violating_num = 5`,
  `contrast_eligible_den = 5`. All 5 successes violate → sc = 0 →
  sPrev = 0. Matches `(l-population-validation)`'s table.
- Ordinal 4, `equals(construction, constructive)`: `contrast_violating_num = 4`,
  `contrast_eligible_den = 5`. 4 successes violate, **1 success
  preserves** → sc = 1 → sPrev = 0.2. Matches
  `(l-population-validation)`'s table.

The persistence-layer arithmetic (`candidate_invariants` columns), the
challenge-layer arithmetic (`invariant_challenges.result_summary`
detail), and the SGO SQL recomputation all agree.

## Consequences

**Correction 7 tier upgrade** (from `(l-coh-mechanism-test)`):
already at CMA-tier. This loop **does not further upgrade** the tier,
because CMA is the highest tier for a code-derivable relationship
verifiable by inspection + deterministic test. What this loop adds is
a **third orthogonal source of evidence**:

1. Code inspection of `VerifySuccessPreserving` (reviewer's message,
   2026-09-12).
2. Synthetic fixture test on hand-built `[]Family` slices with
   identical discrim and different sPrev — `(l-coh-mechanism-test)`.
3. **Live real-data instance**: M7's (a‴-B) run naturally produced
   the two-fixture pattern on real ES corpus data. `sPrev` is the
   sole differentiator between `surviving` and `weaken` on two
   posture-axis candidates emitted in the same campaign.

Nothing else about (l-coh) or its corrections needs to change:

- (l-coh) tables: unchanged.
- (l-coh) correction 5: ES-side residual remains discharged by
  `(source-lineage-check)`; ES-vs-pvnp residual remains open.
- (l-coh) correction 6: remains discharged by `(l-population-validation)`.
- (l-coh) correction 7: remains CMA-tier from `(l-coh-mechanism-test)`;
  now with a live-data instance for cross-check.
- All other corrections unchanged.
- H3 gate: preserved.

## What this does NOT establish

- **Does not** re-open any admissibility rule, verifier tier, policy
  directive, or wire schema field. Pure SGO inspection of persisted
  campaign state.
- **Does not** establish that `equals(construction, constructive)`
  is a scientifically wrong invariant — it was `weakened` (not
  `falsified`), meaning the pipeline flagged it for revision, not
  refutation. Whether a weakened candidate should be repaired,
  narrowed, or discarded is a separate question outside this
  investigation.
- **Does not** establish that the M7 (a‴-B) run's outcomes on other
  candidates were correct — only that these two posture-axis
  candidates traversed the pipeline as designed.
- **Does not** address the ES-vs-pvnp independence question.

## Retraction

The `(m7-emission-divergence)` task entered the next-action list in
the nineteenth-revision scorecard on the assumption that only
`equals(locality, local)` was persisted. That assumption was wrong.
The task is now retired as a non-issue. The scorecard's twentieth
revision will retire it and add this loop's finding as a live
cross-check of correction 7.
