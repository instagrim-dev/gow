# Pilot 003 — pre-capture preparation record

Status: `prepared_not_captured`, **mechanically READY including recovery
reachability**. Preparation stops here by construction; the capture protocol
(inherited from `../../pilot-001/PROTOCOL.md`) takes over after the operator
attestations are recorded.

## Provenance

- Corpus pin: `a868dec` (unchanged; hashes verified).
- Executable: all stages ran at the commit in `executable-commit.txt`
  (single-stage provenance this time — the executable already contained
  mechanism/v3, the `recovery_reachability` check, and the calibration
  harness when preparation started).
- Vocabulary: `mechanism/v3`. Interpretation claims: 8, identical scopes to
  pilot-002 (`interpretations.json`, all resolved).
- Identifiers and DB snapshot digest: `runtime.json`.

## The correction this revision carries

Pilot-002's defect — target self-comparison `unknown`, recovery unreachable
— is corrected by target-side canonicalization only. Live probe, recorded in
`target-self-comparison.json`:

```text
pilot-002 (v2): classification=unknown, all six fields incomparable
pilot-003 (v3): classification=mechanism-near, incomparable fields: none
```

## Replication check (stop condition of PROTOCOL-REVISION.md)

The train substrate replicated pilot-002 exactly, as the v3 superset design
requires (no train label resolves differently under v3):

- clustering `clean`, 12 families;
- same three candidates with identical support/epistemics:
  `confined_to_quadratic_nonresidues` 5 (explicit=1), 
  `identity_carried_solvability` 3 (explicit=0),
  `class_union_construction` 2 (explicit=0);
- all three `surviving` via the same recount branch — the pilot-002
  challenge-coverage qualification carries over verbatim: **three curated
  hypotheses with sufficient recounted support; known-counterexample and
  success-preserving checks remain partly unresolved** and are part of the
  declared treatment.

## Readiness

All gating checks `ready`, including the new `recovery_reachability`
(all targets' decisive fields fully resolved; a positive match is
reachable). Evaluator calibration is pinned in
`internal/pipeline/recovery_calibration_integration_test.go` (v2-defect pin,
v3 self-match, variant-worded recovery, decisive non-recovery, unresolved
unknown).

Epistemic framing unchanged from pilot-002: curated-feature milestone, not a
discovery result; completeness admissions remain zero (absence-based
verified violations unreachable; assessments degrade to unknown honestly).

## Remaining before capture (operator)

The non-mechanical attestations: source/mapping review sign-off freeze,
capture config recording, blinded prompt assembly (train-only context;
survivor IDs/statements/ASTs to B3 only), and the independent assessment
arrangement. The DB snapshot (`pre-capture.db`) is retained locally with its
digest in `runtime.json`, not committed pending a data-retention decision.
