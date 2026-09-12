# (l-coh-mechanism-test) Predeclaration — controlled policy-mechanism test for the `success-preserving` probe

**Purpose**: Test whether the reviewer's correction-7 refinement of the
zero-cell hypothesis correctly captures the pipeline's code behavior.
If confirmed, correction 7 upgrades from **PE (interpretive)** to
**CMA (code-derivable)** without requiring a third research corpus.

**Predeclared prior to writing any code, executed at HEAD** (to be
recorded at execution time).

## The claim being tested

From `(l-coh)` corrected disposition, correction 7:

> Among emitted, support-valid candidates with comparable resolved
> populations, absence of a preserving success prevents this
> particular weakening route. Overall survival additionally depends
> on the remaining challenge outcomes and claim semantics.

Restated as a probe-level prediction:

> `VerifySuccessPreserving(pred, families)` returns different verdicts
> for two fixtures that share every other resolved-population
> property including `fPrev - sPrev` discrimination, provided the
> only difference is whether **any eligible success family satisfies
> `pred`**.

## Why this test is bounded

- Runs entirely in-process against `internal/invariant.VerifySuccessPreserving`.
- Uses the existing `chVocab`, `chPredicate`, `chSignature`, `chFamily`
  test helpers from `internal/invariant/challenge_test.go`.
- No SQLite, no CLI, no provider, no fixtures beyond the assembled
  `[]Family` slice.
- Two `t.Run` subtests, no state carried between them.

## Fixture protocol (frozen before implementation)

Vocabulary: reuse `chVocab(t)` from `challenge_test.go`.

Predicate under test: `contains(preserves, domain.number_theory.property.residue_locality)`
(constructed via `chPredicate(chIDResidue)`).

Every member is a **fully resolved singleton family**:
`SetFieldCompleteness[FieldPreserves] = CompletenessComplete`,
`preserves` claims marked `Resolved`/`Explicit`. This makes every
`Evaluate` call return `VerdictSatisfies` or a decisive negative
(never `VerdictUnknown`), matching the reviewer's "fully resolved
singleton success families" scope.

### Fixture A — sPrev(X) = 0, discrim = 0.5

- **10 failure-side singleton families**, each with one member:
  - 5 members preserve `chIDResidue` (X) → satisfy predicate.
  - 5 members preserve `chIDSieve` (Y), not X → violate predicate.
- **10 success-side (partial_success) singleton families**, each with
  one member:
  - **0 members preserve X.** All 10 preserve Y only → all violate
    predicate.
- fPrev(X) = 5/10 = 0.500; sPrev(X) = 0/10 = 0.000; discrim = **+0.500**.

### Fixture B — sPrev(X) = 0.3, discrim = 0.5

- **10 failure-side singleton families**:
  - 8 preserve X → satisfy predicate.
  - 2 preserve Y only → violate predicate.
- **10 success-side singleton families**:
  - **3 preserve X** → satisfy predicate.
  - 7 preserve Y only → violate predicate.
- fPrev(X) = 8/10 = 0.800; sPrev(X) = 3/10 = 0.300; discrim = **+0.500**.

Both fixtures satisfy the miner emission gates (fc ≥ 2 for X on
failure side, fPrev > sPrev). Both share the same numerical
discrimination. The only structural difference is `sPrev(X)`.

## Predeclared readings

Let `resA = VerifySuccessPreserving(pred, familiesA)` and
`resB = VerifySuccessPreserving(pred, familiesB)`.

### Reading MECH-DISTINGUISH (correction 7 holds)

- `resA.Outcome == OutcomeCompletedNegative`
- `resA.Confirmed == false`
- `resB.Outcome == OutcomeConfirmed`
- `resB.Confirmed == true`
- `resB.Evidence` names ≥ 3 preserving-family cluster IDs
- `resB.Delta.Kind == DeltaContrastCollapse`

If all six hold, correction 7 is **code-verified**. It upgrades from
PE to **CMA on the probe verdict claim**. The probe's binary
threshold on `∃ preserving-success` is confirmed to be the
code-level mechanism that distinguishes these fixtures despite
identical discrimination.

### Reading MECH-COLLAPSE (correction 7 does NOT hold)

- Both fixtures return the same outcome (either both
  `OutcomeCompletedNegative` or both `OutcomeConfirmed`).

If this happens, correction 7's proposed code-level relationship is
falsified. The probe would then be operating on something other
than the presence/absence of a preserving success. The (l-coh)
Correction section would need further revision.

### Reading MECH-PARTIAL (partial code-level match)

- One of the six MECH-DISTINGUISH conditions holds numerically but
  another doesn't (e.g., correct outcome directions but no `Delta`,
  or `Evidence` misses expected clusters).

Any deviation from the six-condition ground is treated as PARTIAL
and prevents the CMA promotion.

## What this test does NOT establish

Explicitly out of scope:

- **Does not establish** that "no preserving success" ⟹ overall
  `surviving`. Overall survival depends on all challenge outcomes
  (known-counterexample, bias-critique, split when applicable) and
  on claim-status semantics.
- **Does not establish** that zero-cell selection identifies
  scientifically useful invariants. That is a research-usefulness
  claim requiring prospective corpus evidence.
- **Does not establish** that the emission gate correctly
  reflects the (l-coh) tables — that is a separate
  `(l-population-validation)` task on the actual SQLite data.
- **Does not establish** the coherence-in-research claim from the
  original (l) framing. The reviewer's overarching point stands:
  this is about the decision rule labeling structure `surviving`,
  not about research coherence.

## Epistemic status after execution

- **On MECH-DISTINGUISH obtained**: correction 7 statement earns
  **CMA-tier** support (deterministic code-derivable relationship
  from `VerifySuccessPreserving` semantics under fully-resolved
  singleton contrast populations). The record is annotated
  accordingly. No policy change follows; probe semantics were
  already implemented.
- **On MECH-COLLAPSE**: correction 7 statement is falsified at
  code level. The (l-coh) record's authoritative disposition
  section requires further correction.
- **On MECH-PARTIAL**: correction 7 statement remains PE-tier;
  narrower CMA claim (whatever the code actually does) is recorded
  as a distinct observation.

## Anti-hindsight rules

- Fixtures assembled and predictions written **before** looking at
  test output.
- No adjustments to fixture population sizes after seeing verdicts.
- If MECH-COLLAPSE occurs, the collapse must be recorded literally
  — no reinterpretation of "the fixture was mis-constructed"
  without documenting the specific construction defect.
- Test committed as a permanent regression artifact regardless of
  outcome direction.

## Deliverables

1. New file: `internal/invariant/success_preserving_discrim_test.go`
   with two subtests (`FixtureA_ZeroPreservingSuccess`,
   `FixtureB_ThreePreservingSuccesses`).
2. Post-execution result record:
   `records/l-coh-mechanism-test-result.md`.
3. Scorecard entry updated with CMA/PE tier per outcome.
