# Pilot 002 — pre-capture preparation record

Status: `prepared_not_captured`, **mechanically READY under the current
readiness implementation — NOT clearance for the primary recovery
comparison** (see the recovery-calibration limitation below, added after
post-preparation review). Preparation stops here by construction; the
capture protocol (inherited from `../../pilot-001/PROTOCOL.md`) takes over
only after the operator attestations are recorded AND the recovery
calibration below is corrected and pinned.

## Recovery-calibration limitation (post-preparation review, empirically confirmed)

The withheld target cannot receive a positive match under the pinned
recovery rule as prepared. `recovery-rule/v1` requires classification
`mechanism-near` (or its surface-distinct variant) under `classify/v1`;
`compareSetField` marks a field incomparable when EITHER side carries any
unresolved claim, and `mechanism-near` requires no incomparable decisive
field. The prepared target's decisive fields are: preserves 0/1 resolved,
operator 0/4, assumption 0/2, breaks 1/2, auxiliary_object 1/2 — every
decisive field carries at least one unresolved claim.

**Empirical probe (executed 2026-09-11 against a scratch copy of
`pre-capture.db`):** the target's exact self-comparison classifies
`unknown` with all six fields `incomparable`. No proposal can classify
better against the target than the target itself; recovery is unreachable
in this prepared state. The `withheld_target: ready` line is therefore a
weaker fact than it appears: the readiness check blocks only on ZERO
resolved decisive claims and does not test recovery reachability.

Bounded correction (reviewer-side, before any capture): calibrate the
evaluator against the actual frozen target (self-comparison positive-match
reachable; known-match recovered; known-different decisively
non-recovering; under-represented unknown), resolve the target's required
vocabulary mappings by ordinary canonicalization of what the target already
states (never favorable interpretation claims), and pin the resulting
revision. Unresolved decisive fields must not be ignored to obtain a green
check.

## Provenance

- Corpus pin: `a868dec` (train tree `da49ddcf…`, target blob `7ed727ab…`) —
  unchanged from pilot-001; the adjudication accepted no corpus change.
- Executable: substrate stages (init → ingest → normalize → interpretation
  → signature/v2 → cluster → failure-space → mine → challenge → define) ran
  at `df84bde` (mechanism/v2 + interpretation claims + contains-ordering
  fix). The readiness report was regenerated at `f7c7433`, which adds only
  the readiness `--vocab-version` population selector (no semantic change to
  any substrate stage; the earlier report at `df84bde` differed only by the
  false "no persisted signatures" blocker).
- Identifiers and DB snapshot digest: `runtime.json`.
- Interpretation claims: 8, per the adjudicated manifest in
  `../PROTOCOL-REVISION.md` (L1×4, L3×2, L4×2), each resolved under
  `mechanism/v2` with ledger provenance (`interpretations.json`).

## Outcome relative to pilot-001's STOP

| Check | pilot-001 | pilot-002 |
|---|---|---|
| failure_cohort | ready (7 families, degraded) | ready (7 families, **clean** clustering) |
| decisive_axis_resolution | ready-thin (2 resolved / 1 signature) | ready (12 resolved across **5** signatures) |
| surviving_invariants | **blocked** (0 candidates) | **ready (3 surviving)** |
| withheld_target | ready | ready |
| overall | NOT READY | **READY** |

The three survivors are exactly the adjudicated properties — support counted
honestly with interpretation-carried members as `inferred`
(explicit counts: L1=1 [es-08's own label], L3=0, L4=0):

- `confined_to_quadratic_nonresidues` — support 5 (L1)
- `identity_carried_solvability` — support 3 (L4)
- `class_union_construction` — support 2 (L3)

**Challenge-coverage qualification.** All three reached `surviving` through
the support-recount (bias-critique) branch. Their known-counterexample and
success-preserving checks report `unconfirmed`: several eligible
failure-side members evaluated `unknown` during counterexample checking
(L1: 2 of 7; L4: 4 of 7; L3: 5 of 7), and the success-contrast side remains
unresolved. The accurate framing is therefore: **three curated hypotheses
with sufficient recounted support; counterexample and success-contrast
assessment remain partly unresolved** — not three failure obstructions
demonstrated to withstand decisive adversarial testing. For the
curated-feature pilot these hypotheses are still the intervention, with
this unresolved challenge coverage declared as part of the treatment.

**Clustering qualification.** `clean` status means the comparison ran
without degradation, not comprehensive semantic resolution: the corpus
still carries 81 unresolved decisive claims, with resolved decisive content
in five of twelve signatures, and the run yields 12 singleton families. The
improvement to claim is that the pipeline now recognizes the adjudicated
properties and can count their support; the seven failure-side groups
remain tied to this vocabulary/profile with their remaining comparison
uncertainty.

## Epistemic framing (do not overclaim)

This readiness is the **curated-feature** milestone: operator-adjudicated
properties, encoded as GeneratedInterpretation hypotheses, mined and
challenged at the unchanged threshold. It demonstrates the substrate can
recognize usable evidence when usable evidence is supplied. It does NOT test
the discovery claim (a model proposing higher-order shared properties
itself) — that requires its own protocol. The completeness_admissions info
line (accepted=0) means absence-based verified violations remain
unreachable; assessments degrade to unknown honestly.

## Remaining before capture (operator)

1. **Recovery calibration (blocking, new):** the bounded correction above —
   evaluator calibration against the actual frozen target, target-side
   canonicalization in a pinned vocabulary revision, and a fresh protocol
   revision per the freeze rules. No captures may be produced before this
   lands; because none exist, the defect can be corrected without selecting
   changes according to B0/B3 outcomes.
2. The non-mechanical attestations from `newf experiment readiness` output
   and pilot-001's PROTOCOL: source/mapping review sign-off freeze, capture
   config recording, blinded prompt assembly (train-only context), and the
   independent assessment arrangement.

The DB snapshot (`pre-capture.db`) is retained locally with its digest in
`runtime.json`; the snapshot itself is not committed pending a
data-retention decision.
