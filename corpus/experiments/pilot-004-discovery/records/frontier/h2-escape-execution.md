# (a‴-B) execution — posture-axis miner extension + H2 escape end-to-end

Recorded: 2026-09-12. Pipeline-space + code-space. User-authorized code
change: extend `DerivingFixtureInvariantMiner` to emit
`OpEquals(<posture>, value)` proposals (opt-in), then validate the H2
escape prediction from `records/enum-axis-invariant-survey.md`.

## Code change (summary)

- `internal/provider/invariant_fixture_derive.go`:
  - Added `emitPostureAxes bool` to `DerivingFixtureInvariantMiner`.
  - Added `NewDerivingFixtureInvariantMinerWithPostureAxes()` constructor.
  - `Identity().ModelName` differs between default and posture-axes
    variants — the reuse key is different, so enabling/disabling on the
    same problem/failure-space produces DIFFERENT revisions.
  - Added `derivePostureAxisProposals` helper implementing the
    classify()-mirror pre-filter (support ≥ 2 AND F prev > S prev).
- `internal/pipeline/invariant.go`: `InvariantMineInput.EmitPostureAxes`.
  Routes through to the WithPostureAxes constructor when the app has
  no injected miner (production path only; tests injecting a specific
  miner supply their own policy).
- `cmd/newf/invariant.go`: `--emit-posture-axes` boolean flag on
  `newf invariants mine`.
- `internal/provider/invariant_fixture_derive_posture_test.go`: 5
  unit tests covering default (no emission), enabled (emission of
  the recurring-worthy subset only), contrast-margin requirement,
  unknown-value skipping, and distinct reuse key.

## Backward compatibility

The default constructor `NewDerivingFixtureInvariantMiner()` behaves
IDENTICALLY to before this change. Every existing test that used the
default deriving miner continues to pass. The extension is opt-in via
constructor choice at the pipeline layer and via `--emit-posture-axes`
at the CLI layer. Full test suite green: `go test ./...`.

## M7 corpus validation

### Stage 1 — mining (SGO)

```
export NEWF_DB=.newf/m7/newf.db
./newf --json invariants mine \
    --problem prb_01M2B8ZMHRCJN4YVH4CP6SNXTD \
    --emit-posture-axes
```

Result: new invariant revision `ivr_01M2BCJZBSVP2XSCFNTBHWCJKH`,
distinct miner_version `invariant/v1+712e5351092a6ecd` from the
default miner's revision (`ivr_01M2B8ZNKFASFT9RARSB3R37B1`,
miner_version `invariant/v1+671cf5d70015c34a`). Distinct reuse keys →
distinct revisions, exactly as designed.

Candidates in the new revision (5):

| # | Predicate | Support | Association |
|---|---|---|---|
| 1 | `Contains(preserves, identity_carried_solvability)` | 3 | recurring |
| 2 | `Contains(preserves, confined_to_quadratic_nonresidues)` | 5 | recurring |
| 3 | **`Equals(locality, local)`** | **3** | **contrast_observed** |
| 4 | `Contains(preserves, class_union_construction)` | 2 | recurring |
| 5 | **`Equals(construction, constructive)`** | **3** | **contrast_observed** |

`contrast_observed` is a STRONGER association than `recurring`
(`engine.go:346-354`): support ≥ threshold AND at least one
success-side family violates the predicate. **All three
`Contains(preserves, X)` invariants sit at `recurring` (not
`contrast_observed`) because H2 makes them return `Unknown` on
contrast families — the completeness stripping specifically prevents
`contrast_observed` for that predicate shape.** Enum-axis predicates
escape this and reach the stronger association state naturally.

### Stage 2 — challenge (SGO)

```
./newf --json challenge inv_01M2BCJZBSVP2XSCFNTNTG8TDV
```

`Equals(locality, local)`: `state_before: proposed` →
`state_after: **surviving**`. Three probes:

- `known-counterexample: unconfirmed` — with 4 concrete failure-side
  violation members recorded (c2, c5, c7, c11 — the
  existential-global-either failures). The reason for `unconfirmed`
  was semantic, not structural: *"a contrast_observed claim is not
  refuted by isolated counterexamples (it asserts discrimination, not
  a universal regularity)"*. This is fundamentally different from the
  H2-corollary `unconfirmed` verdicts on `Contains` invariants, which
  return `unconfirmed` because completeness stripping makes them
  produce `Unknown` — not because the challenge correctly recognised
  the claim shape.
- `success-preserving: unconfirmed` — 0 success-preserving families
  (no success family satisfies `locality = local`).
- `bias-critique: unconfirmed` — support recount at 3, unchanged.

`Equals(construction, constructive)`: `state_before: proposed` →
`state_after: **weaken**`. `success-preserving: confirmed` — the
challenge found the c6 (constructive/global/deterministic partial
success) family, which satisfies the predicate but sits on the
success side. This weakens the invariant, exactly as the hand-eval
predicted (F prev = 3/7 = 0.43, S prev = 1/5 = 0.20 — the c6 outlier
is real).

### Stage 3 — frontier admission of a wire proposal (SGO)

Constructed a wire proposal declaring `posture.locality = "global"`
against `inv_01M2BCJZBSVP2XSCFNTNTG8TDV` (the surviving new
invariant):

```
./newf --json experiment validate-proposals \
    --problem prb_01M2B8ZMHRCJN4YVH4CP6SNXTD \
    --file .../records/frontier/h2-escape-test.wire.json
```

Result: `valid: true`, permitted_targets now includes
`inv_01M2BCJZBSVP2XSCFNTNTG8TDV` alongside L1/L3/L4. The new
enum-axis invariant is targetable by wire proposals — the first
`Equals(<axis>, value)` targetable invariant in M7 history.

### Stage 4 — frontier ranking with decisive verdict (SGO — HEADLINE)

```
./newf --json frontier generate --problem ... --proposals-file ...
```

Result:

```
generation_run_id: run_01M2BCSS7EJ5WBYNST2C1ZCT1V
proposals count: 1
  proposal_id: fpr_01M2BCSS7EJ5WBYNST2J3F0ECY
  violates_any_target: True
  targets:
    inv: inv_01M2BCJZBSVP2XSCFNTNTG8TDV
    verdict: violates
    violated: True
```

**H2 has been escaped end-to-end.** The wire proposal produced a
decisive `verdict: violates` — the FIRST time any wire-authored
proposal in the M7 corpus has produced a violation verdict other than
`unknown`. Both previously-persisted wire proposals against
`Contains(preserves, X)` invariants (joint-crossing `fpr_...D2YB1`
and Brauer-Manin `fpr_...TFTE`) returned `unknown` on all targets.
The enum-axis invariant escapes that stall exactly as the
`records/enum-axis-invariant-survey.md` CMA prediction stated.

### Stage 5 — evaluation still blocks at H3 (SGO)

```
./newf --json evaluate fpr_01M2BCSS7EJ5WBYNST2J3F0ECY \
    --problem prb_01M2B8ZMHRCJN4YVH4CP6SNXTD
```

Result:

```
verdict: verification_blocked
verifier_kind: model-judgment
verification_strength: single-model-judgment
notes: no verifier returned a decisive verdict
```

**H3 is still closed. This is expected and important.** H2 and H3 are
independent gates:

- **H2** governs whether a wire proposal can produce a decisive
  violation verdict against its target invariant, and is a property
  of `invariant.Evaluate` on the wire signature (per predicate.go).
- **H3** governs whether the pipeline's verifier hierarchy can decide
  whether the proposal AS A WHOLE is a `failure` / `partial_failure`
  / `success` outcome, and is a property of the verifier chain
  (deterministic-check → counterexample-search → model-judgment).

Escaping H2 does not automatically escape H3. A proposal can decisively
violate its target invariant AND still be an inconclusive-outcome
research contribution. The H3 gate is orthogonal.

The evaluation is now persisted as
`evl_01M2BCT64PPDDB9NS6JJF3V64G`. The proposal shares the same
terminal-state pattern as the two prior wire proposals (H1 + H3 =
blocked at evaluation), but with the crucial difference that the H2
gate is now demonstrably passable.

## Consequences for the pipeline (design-space observations)

1. **Enum-axis invariants are code-verifiable through wire proposals;
   set-field-completeness-dependent invariants are not.** This is a
   substantive design property, not a bug. The predicate shape
   choice determines what class of proposals can attack it.
2. **`contrast_observed` reveals a mining-stage prevalence dimension
   that `Contains(preserves, X)` invariants cannot access on this
   corpus.** Any invariant mined at `contrast_observed` has
   *stronger* code-measured evidence than one at `recurring`, and
   should probably be preferred in downstream ranking.
3. **The challenge protocol correctly recognises association-type
   semantics.** `contrast_observed` claims are correctly identified
   as discrimination claims that cannot be refuted by isolated
   counterexamples. This is not the H2-corollary
   "everything-returns-unconfirmed" pattern; it is a substantively
   different reason for `unconfirmed` that depends on the claim
   shape.
4. **The evaluator's `verification_blocked` outcome now applies
   uniformly to all three persisted wire proposals across two
   different H2 disposition classes** (stalled at H2 for
   joint-crossing/Brauer-Manin; escaped H2 for h2-escape-test).
   This tells us the H3 stall is NOT a consequence of H2 —
   it's a property of the verifier chain's ability to reach a
   decisive proposal-outcome verdict, which is independent.

## What was NOT done

- Did not modify default miner behaviour (backward compatibility
  preserved).
- Did not challenge `equals(construction, constructive)` further —
  its `weaken` disposition is well-explained by the c6 counterexample
  and there is no active hypothesis about how to strengthen it.
- Did not attempt to escape H3 — that is a separate, distinct research
  route (see the closure scorecard's design-space observation about
  the three-gate ratchet).
- Did not attempt to admit the newly-persisted wire proposal
  `fpr_01M2BCSS...` via `newf evidence admit`. Since its evaluation
  ends at `verification_blocked` (same as the two prior proposals),
  the H3 finding predicts admission will refuse at the same barrier
  ("not a recorded evaluated failure"). Recording this prediction
  without spending a loop to re-observe it.

## Verification tier

- Code change is CMA (traced through the git diff at HEAD).
- Unit tests are SGO (pass under `go test`).
- M7 corpus outcomes at each stage are SGO — verbatim `newf --json`
  output.
- H2 escape claim is directly demonstrated: `verdict: violates`
  is decisive by definition of the `Verdict` enum in `predicate.go`.
- H3 orthogonality claim is CMA (code trace) + SGO (both H2-stall
  and H2-escape proposals produced `verification_blocked` at
  evaluation).
- Design-space observations about association strength and search
  policy implications are PE — proposals for downstream authority to
  consider.
