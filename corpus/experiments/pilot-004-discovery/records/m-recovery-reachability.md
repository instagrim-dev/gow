# (m) Investigation of `recovery_reachability` schema-evolution blocker

Recorded: 2026-09-12. Pipeline-space (SGO on 4 corpora + code trace).

## Question

The (k-restricted) retrospective on pilot-001 surfaced a new
readiness blocker: `recovery_reachability: blocked` under classify/v3.
Is this a schema-migration artifact (bug), a legitimate hardening of
the classification contract (feature), or something else?

## Code trace (CMA)

`internal/pipeline/readiness.go:216-236` — the check does a
**self-comparison** of every holdout-target signature:

```
profile := canon.ProfileMechanismV3()
...
sig := signatureFromRecordWithProvenance(rec)
cls := canon.CompareWithProfile(sig, sig, profile).Classification
if experiment.IsRecovering(cls) {
    reachableTargets++
}
```

The mechanical claim, verbatim from the code's own comment
(lines 202-215):

> **no proposal can classify better against a target than the target
> itself, so a non-recovering self-comparison means recovery is
> unreachable for ANY proposal.**

> the prior approximation missed EMPTY decisive fields without a
> `complete` justification, which classify/v2's missing-data
> contract also makes incomparable.

> The correction path is reviewer-side canonicalization of stated
> labels and/or justified completeness — never ignoring the gap,
> and never interpretation claims on the target.

This is not a bug. The check is a mechanical necessity: if a target
signature cannot self-classify as "recovering" under the current
classification profile, no proposal signature can outperform it under
that classifier, so recovery under `recovery-rule/v1` is
categorically unreachable. The check was strengthened when
classify/v2 (and now classify/v3) tightened the missing-data
contract.

## Cross-corpus diagnosis (SGO)

Ran `experiment readiness` on the train problem of each persisted
corpus:

| Corpus | withheld_target claims | recovery_reachability | classify profile |
|---|---|---|---|
| M7 | resolved=11, unresolved=**0** | **ready** (all 1 targets pass) | classify/v3 |
| pilot-001 | resolved=2, unresolved=**9** | blocked (1/1 target) | classify/v3 |
| pilot-002 | resolved=2, unresolved=**9** | blocked (1/1 target) | classify/v3 |
| pilot-003 | resolved=13, unresolved=**9** | blocked (2/2 targets) | classify/v3 |

**The determining variable is unresolved-claims count.** M7's target
has ZERO unresolved decisive-field claims; the three historical
pilots' targets each have 9. Under classify/v3's missing-data
contract, an unresolved decisive-field claim makes the field
"incomparable" against itself, which drops the overall classification
out of the recovery set.

Pilot-003 has a specific SGO anchor: its `records/readiness.json`
originally recorded `recovery_reachability: ready` — proving the
check existed then but produced a passing verdict under the older
profile. Today under classify/v3, pilot-003 blocks. This is a
clean regression from classify/vN to classify/v3 semantics, not an
absence-of-check that was later added.

## Verdict on (m)

**`recovery_reachability` is a legitimate hardening of the
classification contract, not a bug.** classify/v3's completeness-
aware absence semantics correctly refuse to admit targets whose
decisive fields are unresolved without a `complete` justification.
The check enforces exactly what the code comment describes: a
mechanical impossibility (target can't self-recover ⇒ nothing can
recover it).

**All three historical ES pilots' targets fail the current bar.**
This is a corpus-versioning consequence, not a defect:

- Pilots 1-3 were normalized under an earlier schema that admitted
  unresolved-claim targets.
- Their targets are frozen artifacts of that era.
- classify/v3 correctly refuses to score against them without
  re-normalization.

**M7's target passes** because it was freshly normalized under the
current schema with full decisive-field resolution (11 resolved,
0 unresolved).

## Design-space observation (recorded, not a change proposal)

The pipeline currently checks target readiness against the **current**
classification profile. This is correct for today's experiments but
creates a **corpus-versioning ratchet**: as classification semantics
tighten, historical pilots' targets fall out of admissibility.

Two framings, both legitimate:

1. **Epistemic-safety framing:** the ratchet is desirable. classify/v1
   was admitting targets that are not self-consistent under stricter
   semantics; retiring them prevents false-positive matches.
2. **Research-reproducibility framing:** the ratchet is partially
   costly. Historical pilots' STOP dispositions are locked in the
   past, but retrospective research on those corpora (e.g., testing
   whether a new capability like (a‴-B) would have changed outcomes)
   requires either target re-normalization or a mode that pins the
   readiness check to the pilot's historical profile.

A speculative resolution would be **pinned-profile readiness** — an
opt-in mode where readiness is checked against the profile in effect
at pilot preparation time (recorded on the manifest). Not proposed as
a change here; recorded as a design-space observation for downstream
authority.

## Consequence for path (k-full)

Full re-execution of any historical ES pilot under classify/v3
requires re-normalization of its target(s) with full decisive-field
resolution. This is an **atlas-quality task** (canonicalize each
unresolved field claim against the vocabulary, or admit completeness
with a documented basis), not a schema-migration or code-change task.
The correction path is exactly what the readiness check's error
message says: "canonicalize the target's stated labels and/or justify
field completeness in a pinned revision (do not ignore the gap, do
not add interpretation claims to the target)."

The corollary is that path (k-full) has two prerequisites:

1. Operator authorization to modify historical pilot state (either
   in-place target re-normalization or a fresh derived-pilot corpus).
2. Atlas-quality work to canonicalize each of the 9 unresolved
   decisive-field claims in each historical target (some of which
   may honestly have no canonical mapping and require a
   completeness-with-basis admission instead).

Neither prerequisite is a code change. Both are research-execution
tasks requiring operator authorization.

## Verification tier

- Code trace of the check: **CMA** (verbatim reading of
  `readiness.go:216-236` and its comment).
- Cross-corpus SGO on unresolved-claim counts and check outcomes:
  **SGO** (direct CLI + sqlite reads).
- Pilot-003 historical `recovery_reachability: ready` → today
  `blocked`: **SGO** (verbatim from `corpus/experiments/pilot-003/records/readiness.json`
  compared to current CLI output).
- Verdict "not a bug, legitimate hardening": **PE** based on the
  above SGO and CMA.
- Design-space observation about pinned-profile readiness: **PE**
  (recorded, not authorised).
