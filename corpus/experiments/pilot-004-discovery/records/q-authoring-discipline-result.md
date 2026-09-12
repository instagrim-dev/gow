# (q) Result — proposal authoring discipline documented at `docs/proposal-authoring.md`; wire schema and ranker code unchanged

**Status**: closed 2026-09-12. Documentation-only. No code
changes. No admission rules changed. No verifier tiers changed.
No policy directives changed. H3 gate preserved.

## What (q) proposed to solve

The (o-closed) reviewer correction observed that declared
`evaluation_cost` and actual verifier readiness are different
properties. A cost attestation is a self-report about the
falsification path; verifier readiness is a claim about whether
the pipeline (or an external check) can decide the proposal's
structural claim. The two do not automatically coincide, and the
wire schema alone cannot align them.

The reviewer's frame:

> The authoring convention in (q) can improve this signal, but it
> remains an operator judgment requiring calibration.

(q) closes this gap at the documentation layer — not at the
schema, verifier, or ranker layer — by making the operator
convention explicit.

## What (q) delivered

**New document**: [`docs/proposal-authoring.md`](../../../docs/proposal-authoring.md).

Contents:

1. **Property distinction**: declared `evaluation_cost` is
   self-reported cost; verifier readiness is a separate property.
   Reproduces the (o-closed) counterexample
   (`fpr_01M2BCSS7EJ5WBYNST2J3F0ECY`, a legitimate `low`
   attestation without a candidate-specific witness obligation).
2. **What the ranker DOES with `evaluation_cost`**: fourth
   ordering dimension in `baseObjectiveLess`; per-target
   falsifiability floor protects one violator per target.
3. **What the ranker DOES NOT do with `evaluation_cost`**: no
   verifier existence guarantee; no budget reservation; no
   evaluation guarantee; no witness attachment; no concreteness
   discrimination.
4. **Decision rule** for self-attesting the ordinal:
   - `low` = deterministic pipeline read or already-runnable
     external check.
   - `medium` = bounded domain math or small new code.
   - `high` = unbounded domain reasoning, empirical parameter
     search, novel verifier implementation, or expert judgment.
5. **Two anti-patterns**:
   - Declaring `low` from optimism (the failure mode
     (o-closed) identified).
   - Declaring `high` from indecision (declared cost is a claim
     about the falsification path, not about the operator's
     confidence).
6. **Composition with `cheapest_falsification_path`**: the
   ordinal must be consistent with the prose by construction.
7. **Calibration signal**: no automated procedure; calibration
   happens by review, feeding back into the authoring convention.
8. **Relation to `expected_information_gain`**: two orthogonal
   optional ordinals expressing different properties.
9. **Explicit non-scope**: does not change the wire schema, the
   ranker, admission, verifier tiers, or policy directives; does
   not close C7/C8/witness-occurrence attribution obligations.
10. **Reference material** cross-linking `frontier-generation.md`,
    `o-closed-lever-exists.md`, `l-coh-arc-consolidation.md`,
    `internal/policy/apply.go`, and
    `internal/provider/untrusted_proposer.go`.

**Cross-link added**:
[`docs/frontier-generation.md`](../../../docs/frontier-generation.md)'s
wire-semantics section now points to `docs/proposal-authoring.md`
for the calibration convention.

## Explicit preservation

- No code changes.
- No schema changes (`proposal-wire/v1` unchanged).
- No `internal/provider/untrusted_proposer.go` changes.
- No `internal/policy/apply.go` changes.
- No `baseObjectiveLess` changes.
- No falsifiability-floor changes.
- No admission-rule changes.
- No verifier-tier changes.
- H3 gate preserved.

## What (q) does NOT do

- Does NOT enforce the authoring convention. The code cannot
  detect a `low` attestation from optimism versus one grounded in
  a concrete falsification path; only review can.
- Does NOT promote declared cost into verifier readiness.
- Does NOT propose a new admission gate at ingest time.
- Does NOT close the C7, C8, or witness-occurrence attribution
  obligations recorded elsewhere in pilot-004.
- Does NOT calibrate against a specific historical corpus (the
  operator-judgment signal is calibrated per-review, not
  ahead-of-time).
- Does NOT replace any existing documentation. The wire schema
  reference stays in `frontier-generation.md`; the authoring
  convention lives in a separate operator-facing document.

## Consequences for the scorecard

- `(q)` closed as documentation-complete. Authoring convention
  documented at `docs/proposal-authoring.md`, cross-linked from
  `docs/frontier-generation.md`.
- `(o)` disposition unchanged. The (o-closed) narrower warrant
  remains the operative statement; (q) does not alter its scope.

## Consequences for future work

Future proposal capture prompts, capture reviews, and pilot
freeze checklists can now cite `docs/proposal-authoring.md` as
the operator convention. A future observation that a `low`-
declared proposal failed to reach a verdict due to a witness
gap should be recorded in the proposal's follow-on record with
a reference to this convention, so calibration accumulates in
the corpus rather than in operator memory.
