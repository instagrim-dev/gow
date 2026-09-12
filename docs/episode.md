# Prospective two-observation episodes (v41)

The 2026-09-12 review's epistemic conclusion: GoW does not need another
explanation of why a map could help — it needs a checked case showing whether
the map's next decision was worth its cost. The minimum persuasive
demonstration is a closed loop:

```text
freeze map, action, prediction, and scoring rule
→ obtain an externally checked outcome
→ revise the map
→ commit to a DIFFERENT next action
→ obtain and score the next outcome
```

The first miss remains a miss. The revision earns credit only on later
evidence. This document describes the typed substrate that makes such an
episode auditable; running one against a live problem is an operator act on
top of it.

## Records (migration v41)

| Table | Record | Mutability |
|---|---|---|
| `episodes` | frozen preregistration: problem, title, `map_ref`, `scoring_rule`, status | preregistration fields frozen by trigger; only `status`/`completed_at` may change |
| `episode_commitments` | one per step (1, 2): action, falsifiable prediction, predicted verdict, map ref, basis | immutable, `UNIQUE(episode_id, step)` |
| `episode_observations` | the externally checked outcome: checker identity, canonical payload, verdict, code-derived score | immutable, append-once per commitment |
| `episode_revisions` | the map change between steps: before/after refs, what changed, basis | immutable, one per episode |

Order is a persistence invariant, enforced transactionally:

- no revision before the step-1 observation (a revision responds to evidence);
- no step-2 commitment before the step-1 observation AND a recorded revision;
- the step-2 action must differ from step 1 (a re-roll is not a revised decision);
- a revision whose map reference did not change is refused;
- one observation per commitment, forever; the second observation completes
  the episode in the same transaction;
- a completed episode accepts nothing further.

## Scoring is code-owned

`hit` iff the checker verdict equals the commitment's `predicted_verdict` —
derived by the pipeline and re-derived at the persistence boundary, so a
caller cannot store a flattering score. The scoring rule is recorded verbatim
on the episode at preregistration.

The derived `revision_credit` in the episode view (`earned` / `not-earned` /
`unresolved`) is computed on read, never stored: `earned` iff step 2 was
observed a hit after a recorded revision. A second miss completes the episode
honestly as `not-earned` — a legitimate stopping state.

## The external outcome mechanism: `internal/witness`

The first decisive domain checker (issue #23): single-instance Erdős–Straus
witness claims, decided exactly over `math/big`:

```text
(x, y, z) witnesses 4/n = 1/x + 1/y + 1/z  ⇔  4·x·y·z == n·(y·z + x·z + x·y)
```

No floats, no tolerance, no judgment. Observations record the v38 axes
honestly: `verification_subject = domain-goal` (the verdict is about the
domain object, not an annotation) and `verification_strength = reproducible`
(an in-repo exact computation, not a formal proof of any surrounding claim).
Rejections name the exact identity failure with both sides shown.

## CLI

```text
newf episode preregister --problem prb_... --title ... --map invr_... \
  --action "..." --prediction "..." --predict witness-valid --note "..."
newf episode observe --episode epi_... --witness "5,2,4,20" [--note ...]
newf episode revise  --episode epi_... --map invr_... --changed "..." --note "..."
newf episode commit  --episode epi_... --action "..." --prediction "..." \
  --predict witness-valid --note "..."
newf episode show    --episode epi_... | --problem prb_...
```

## Regression tests

- `internal/witness/witness_test.go` — valid witnesses, the issue-#23
  near-miss `(2,4,21)`, non-positive/missing values, exactness at 10^30 scale.
- `internal/store/episode_store_test.go` — loop-order refusals, authored-score
  refusal, append-once, completion, frozen preregistration, SQL immutability.
- `internal/pipeline/episode_integration_test.go` — the full loop at the App
  boundary (miss → revise → different action → hit → credit `earned`) and the
  honest negative (second miss → credit `not-earned`).

## What this does not claim

The substrate makes a prospective episode auditable; it does not run one. A
persuasive comparative episode still requires: a live map whose revision is
produced by the pipeline's own machinery (not authored ad hoc), an action
selection actually driven by that map, and cost accounting for curation,
normalization, mapping, challenge, generation, and verification — not merely
final proposal cost. Those remain operator workstream on top of this record
structure.
