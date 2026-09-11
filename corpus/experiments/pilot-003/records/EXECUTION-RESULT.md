# Pilot 003 — execution result (automatic classification only)

Status: `executed_awaiting_independent_review`. Experiment
`exp_01M27QA4137EYA7S1FPTF2Q0SM`, leakage audit passed, budgets 8/8 as
predeclared. v32 execution rows in `executions.json` bind arm → generation →
consumed file hash. Admission audits: B0 corrected=1, all other counters
zero for both arms.

Revision 2 (post-execution review): result classifications corrected. The
**primary B3 captured attempt is WIRE-INVALID** — the sealed `b3-raw.json`
fails the production importer (optional enums nested inside `mechanism`).
It is not a research failure and not `no_recovery`. What executed for B3 is
a **repaired-input assessment** of the recorded lossless derivative
`b3-consumed.json` (transform + digests in the capture record; repair
policy symmetric across arms — B0 required none and passed the importer as
captured). B3's transcript audit is reclassified
`protocol_deviation_pending_adjudication` (see capture record): the capture
rule was unconditional and no pre-capture exception existed; the
independent reviewer adjudicates admissibility.

## Automatic result under recovery-rule/v1 (classify/v1, mechanism/v3)

| Arm | Input class | Proposals | Recovered | Decisive non-recovery | Unknown |
|---|---|---:|---|---:|---:|
| B0 undirected | as captured | 7 | no | 2 | 5 |
| B3 invariant-guided | repaired-input derivative | 6 | no | 2 | 4 |

Recovery delta: **inconclusive** — "at least one arm was not decisively
assessed on this split; no recovery negative can be claimed."

Assessment-rule note: this pinned result was computed under `classify/v1`
(profile `f3334d08…`). The corrected production rule `classify/v2`
(completeness-aware absence: an unobserved empty decisive field is an
epistemic gap, never decisive negative evidence) now exists with a distinct
profile hash; a reassessment under it would be a **corrected assessment
with its own experiment identity**, reported separately — never substituted
for this result. Under v2 (whose
missing-data contract also guards mutual silence and partial-subset
mismatches), all four decisive_no counts above would degrade to unknown:
this frozen corpus declares no completeness, so no recorded-subset conflict
against it is decisive. The reviewer should weigh the decisive_no rows with
that caveat; decisive automatic negatives against this target require a
corpus revision with justified completeness (a research/protocol decision).

## Interpretation discipline (read before quoting any number)

1. **No arm recovered, and no recovery NEGATIVE can be claimed either.**
   The unknowns are proposals whose free-form surface labels do not resolve
   under the pinned `mechanism/v3`, making decisive fields incomparable.
   The classifier abstained; it did not judge those mechanisms distinct.
2. **The most important unknown is B3 rank 4** (also visible, weaker, in
   B0 rank 2's variety direction): its stated mechanism is fiberwise
   lattice-point counting with geometry-of-numbers estimation on the lifted
   variety — wording that a human reader may judge structurally close to
   the withheld target's move. Its labels (e.g. "fiberwise lattice-point
   counting", "geometry-of-numbers estimation") do not resolve under v3,
   so the automatic rule could not compare it. **Resolving those labels
   NOW — after seeing outcomes — would be outcome-selected vocabulary
   change and is prohibited.** Whether that proposal recovers the target's
   structural move is exactly the question the independent unlabeled
   review must answer; the automatic result neither supports nor refutes
   it.
3. **What the automatic result does establish**: the harness executed the
   full blinded path end-to-end on genuinely external captures — strict
   wire admission (including a real rejection, recorded verbatim, resolved
   by a machine-verified lossless re-nesting of two optional fields),
   equal budgets, target-side comparability (the reachability correction
   held: decisive non-recoveries ARE reachable now, where pilot-002 could
   only say unknown), and abstention where the vocabulary cannot compare.
4. **This is still the curated-feature pilot.** Nothing here tests whether
   a model can DISCOVER failure invariants; B3's treatment was the three
   adjudicated hypotheses with their declared challenge-coverage limits.

## What the independent reviewer receives

Unlabeled proposals (both arms shuffled, no arm labels, no automatic
classifications) and the predeclared rubric — not this file. Their
judgments belong in `independent-assessment.json`; review is not complete
until an actual reviewer supplies it. A post-review protocol revision may
canonicalize proposal wording for a REASSESSMENT arm, clearly labeled as
post-hoc and never replacing this pinned automatic result.
