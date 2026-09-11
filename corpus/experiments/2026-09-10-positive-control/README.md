# M7 positive control — populated invariant discovery (known ground truth)

## Recorded conclusion

**Populated positive control: the discovery path fires and discriminates.** On a
deliberately synthetic fixture with known ground truth, the full research path
produced actual records — scoped failure evidence, a nonempty candidate set, a
completed-negative challenge campaign, a surviving invariant, a guided proposal
that *verifiably* breaks that invariant, and a **decisive structural recovery**
of the withheld target — while the undirected baseline honestly produced no
proposal. A four-part mutation battery then flips each expected result for the
corresponding reason, so a green result is evidence the path *discriminates*,
not merely that it runs.

This is the complement to `../2026-09-10-esr-negative-control/` (expected
abstention on unsuitable inputs). Together they are the negative-path control
and the populated positive control the review asked for — complementary, not
alternatives.

## Where it lives (auditable, not `/tmp`)

The control is executable and deterministic: it is a committed integration test,
`internal/pipeline/positive_control_integration_test.go`, run by
`go test ./internal/pipeline/ -run PositiveControl`. Ground truth is authored in
the fixture, so the assertions ARE the audit trail.

- commit: `760e54a661493dc28e1ea4686c76607b0d9213ef` (+ this change)
- provider: `fixture` (deterministic; no model training)
- recovery rule: `recovery-rule/v1`; profile `classify/v1` (`ProfileMechanismV1`)
- mode: `blinded`

## Required path — each link is an asserted record

```text
scoped failure evidence   3 families preserving residue_locality, all axes RESOLVED
  -> comparable mechanisms real clustering (distinct operators/representations), not "cannot compare"
  -> nonempty candidate    miner derives `preserves contains residue_locality` (support >= 2)
  -> executed challenges    a completed-negative campaign transitions the candidate to `surviving`
  -> eligible guided target the surviving invariant B3 may attack
  -> generated proposal     a mechanism whose preserves is COMPLETE and omits residue_locality
                            => `preserves(residue_locality)` evaluates to a VERIFIED violation
  -> decisive recovery      the proposal is mechanism-near the WITHHELD target (both preserve
                            mean_growth_rate, exhaustively) => experiment conclusion
                            `structural_recovery`, with unknown_count = unassessed_count = 0
```

`TestIntegrationPositiveControlDecisiveRecovery` asserts every link, plus the
scientific contrast: **B0 (undirected) is honest-empty; B3 (guided) recovers.**
B0 vs B3 is a real guided-vs-undirected comparison here, not two empty arms.

## Mutation battery — the result changes for the right reason

| Mutation | Expected change | Where |
|---|---|---|
| Remove shared support (1 family, support 1 < 2) | miner derives **NO** candidate → no failure-invariant task | `TestPositiveControlMutationRemoveEvidence` |
| Add a valid known counterexample to an overbroad (`recurring`) candidate | candidate is **falsified**; the same counterexample leaves a `contrast_observed` candidate surviving | `internal/invariant`: `TestVerifyKnownCounterexampleRefutationDependsOnClaimKind` (existing regression) |
| Make a decisive axis **unknown** (unresolved preserves on the target) | assessment is **`unknown`** — never coerced to recovery or decisive non-recovery | `TestPositiveControlMutationUnknownFieldIsInconclusive` |
| Exhaust the evaluation budget (2 targets, budget 1) | proposal is **`unassessed`**, consumed budget recorded exactly (1); unlimited budget → decisive_no | `TestPositiveControlMutationBudgetExhaustion` |

## Two structural findings this control surfaced

Building a control that reaches **verified violation → decisive recovery**
exposed two facts about the persisted CLI/vocab path — the same two blockers the
negative-path review predicted, now pinned mechanistically:

1. **The signature builder never marks a set field `Complete`.**
   `internal/canon/signature.go` sets every set field to
   `CompletenessUnobserved`. `invariant.Evaluate` returns `violates` for a
   `contains`-absence **only** when the field is `CompletenessComplete`;
   otherwise absence is `unknown` (an epistemic gap, correctly). So a mechanism
   rehydrated from a persisted record can never yield a *verified* break — only a
   directly-authored signature (as a live generator emits) can. The recovering
   generator here sets `preserves` complete on purpose. **Implication:** the
   corpus→signature path cannot today manufacture the "exhaustively extracted"
   completeness that verified negatives require; a real run needs a generator
   (or normalizer) that asserts field completeness with provenance.

2. **The default deriving fixture generator's only break is out-of-vocabulary.**
   `DerivingFixtureGenerator` proposes a mechanism preserving
   `core.property.global_coupling`, which is intentionally NOT in `mechanism/v1`
   (see `internal/pipeline/success_integration_test.go`, which asserts that
   condition is *inadmissible*). A vocab-normalized target can therefore never be
   mechanism-near it — which is exactly why the shipped end-to-end test yields
   `no_recovery`. The positive control uses an in-vocab complement
   (`mean_growth_rate`) shared by proposal and target so recovery is reachable
   and decisive. **Implication:** demonstrating recovery end-to-end requires a
   generator whose proposals live in the same resolved vocabulary as the target.

Neither is a bug in the harness; both are honest limits of the deterministic
fixtures. They are the concrete "corpus/generator interface" work the negative
run pointed at.

## What this validates — and what it does not

Validated: the mechanics of populated discovery — mining support thresholds,
challenge survival on a completed negative, verified structural violation,
decisive recovery classification, and the budget/unknown discipline — all
DISCRIMINATE under known ground truth.

**Not** validated: discovery *ability*. The proposal here is authored by a
fixture with the answer built in; it is not evidence that the system can find a
withheld structural move it was not handed. Testing the research thesis requires
independently assessed source cases and a proposal-producing system whose
outputs are not predetermined by the fixture. That is the next step after
mechanics.
