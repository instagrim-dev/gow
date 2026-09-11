# Pilot 002 — pre-capture preparation record

Status: `prepared_not_captured`, **mechanically READY**. Preparation stops
here by construction; the capture protocol (inherited from
`../../pilot-001/PROTOCOL.md`) takes over only after the operator
attestations are recorded.

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

The non-mechanical attestations from `newf experiment readiness` output and
pilot-001's PROTOCOL: source/mapping review sign-off freeze, capture config
recording, blinded prompt assembly (train-only context), and the independent
assessment arrangement. The DB snapshot (`pre-capture.db`) is retained
locally with its digest in `runtime.json`; the snapshot itself is not
committed pending a data-retention decision.
