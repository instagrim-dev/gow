# (l-coh-mechanism-test) Result — Reading MECH-DISTINGUISH obtained; correction 7 upgrades to CMA-tier

**Status**: executed 2026-09-12. Predeclaration:
`records/l-coh-mechanism-test-predeclaration.md`. Test committed as
`internal/invariant/success_preserving_discrim_test.go`.

## Bottom line

**Predeclared Reading MECH-DISTINGUISH is obtained.** Both fixtures
satisfy all six frozen conditions:

- Fixture A (sPrev = 0, discrim = 0.5): `Outcome =
  completed_negative`, `Confirmed = false`, `Delta = nil`,
  `Evidence = []`.
- Fixture B (sPrev = 0.3, discrim = 0.5): `Outcome = confirmed`,
  `Confirmed = true`, `Delta.Kind = contrast-collapse`, `Evidence`
  has 3 entries, all `Kind = success_family`.

The `success-preserving` probe distinguishes the two fixtures
despite identical `fPrev - sPrev` discrimination. The distinguishing
signal is the presence of at least one eligible success family in
which every eligible member satisfies the predicate — exactly what
`VerifySuccessPreserving`'s implementation states.

## Epistemic upgrade

Correction 7 from the `(l-coh)` corrected disposition —

> Among emitted, support-valid candidates with comparable resolved
> populations, absence of a preserving success prevents this
> particular weakening route. Overall survival additionally depends
> on the remaining challenge outcomes and claim semantics.

— is now **CMA-tier at the probe verdict level**. The claim is a
code-derivable relationship, verified by:

1. Inspection of `internal/invariant/challenge.go:VerifySuccessPreserving`
   (already summarized in the reviewer's correction).
2. This test's deterministic pass on two controlled fixtures at
   HEAD `474813f`+1.

The narrower CMA scope is stated deliberately: it establishes
**how the probe verdict depends on the presence/absence of a
preserving success family** in fully-resolved singleton contrast
populations. It does NOT establish overall survival semantics
(explicit anti-scope from the predeclaration).

## Method

Test file `internal/invariant/success_preserving_discrim_test.go`
builds two `[]Family` slices with singleton members, each preserving
exactly one canonical id (`residue_locality` = X or `sieve` = Y).
Predicate: `contains(preserves, residue_locality)`. Every signature
has `SetFieldCompleteness[FieldPreserves] = complete`, so
`Evaluate` returns only decisive verdicts (never `Unknown`).

Populations:

| Fixture | fail-with-X | fail-with-Y | succ-with-X | succ-with-Y | fPrev | sPrev | discrim |
|---|---:|---:|---:|---:|---:|---:|---:|
| A | 5 | 5 | 0 | 10 | 0.500 | 0.000 | +0.500 |
| B | 8 | 2 | 3 | 7 | 0.800 | 0.300 | +0.500 |

Only `VerifySuccessPreserving` is exercised. The miner emission gate
and other probes are out of scope for this test (they are separately
tested in `challenge_test.go` and `engine_test.go`).

## What the test does NOT establish

Directly from the predeclaration's anti-scope:

- **Does not establish** that "no preserving success" ⟹ overall
  `surviving`. Survival requires the campaign to also avoid a
  confirmed known-counterexample and a support-collapse under
  bias-critique, and requires appropriate claim semantics.
- **Does not establish** that zero-cell selection identifies
  scientifically useful invariants. That is a research-usefulness
  claim requiring prospective corpus evidence — separate from
  this code-mechanism test.
- **Does not validate** the (l-coh) family-prevalence arithmetic.
  Population-query validation on the actual SQLite data remains a
  separate obligation (see (l-population-validation) in the
  scorecard).
- **Does not re-open** or narrow any of the other eight corrections
  in the (l-coh) corrected disposition. This test discharges only
  correction 7's code-level statement, at CMA-tier.

## Consequences for the (l-coh) record

Correction 7 in the authoritative disposition section is now
annotated **CMA-tier** rather than **PE-tier**. The rest of the
corrected disposition is unchanged. The (l) result's "narrowed
warrant" annotation is unchanged: the code-mechanism confirmation
does not restore the earlier research-coherence framing; it
confirms the *decision-rule* explanation the reviewer proposed.

## Consequences for search policy

None. Probe semantics were already implemented; no admission rules,
no verifier tiers, no policy directives, and no wire schema fields
change. `H3` gate preserved.

## Deliverables produced

- `internal/invariant/success_preserving_discrim_test.go` — the
  predeclared regression test. Passes deterministically.
- This result record.
- Scorecard entry to be updated to reflect the CMA-tier upgrade of
  correction 7.

## The reviewer's overarching frame, reinforced

> The most important remaining distinction is between discovering
> structure in research and rediscovering the decision rule that
> labels that structure `surviving`.

This test is squarely on the "decision rule" side. It formally
verifies part of the labeling mechanism. It contributes nothing to
the research-structure side, which remains an open agenda item that
(l-extension) — a third truly independent corpus in a non-ES,
non-P-vs-NP domain — would begin to address.
