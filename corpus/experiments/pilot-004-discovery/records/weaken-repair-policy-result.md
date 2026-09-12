# (weaken-repair-policy) Result — no gap: bookkeeping-class weaken is terminal by explicit design; `contrast-collapse` deliberately earns no directive

**Status**: executed 2026-09-12. Design-inspection loop.
Closes `(weaken-repair-policy)` as **no code change warranted**;
records a design observation about the probe-vs-miner threshold
asymmetry.

## Bottom line

The pipeline's designed response to a `weaken` verdict driven by a
`contrast-collapse` boundary delta is **exactly what M7 exhibits**:

1. Persist the typed `contrast-collapse` boundary delta.
2. Do NOT emit any domain-search directive.
3. Do NOT construct a narrower child invariant automatically.
4. Do NOT insert any `invariant_lineage` row.
5. Leave the candidate at `weaken` state, permanently unless a fresh
   mining pass or a model-authored challenger proposes a new
   candidate.

This is not a gap. It is the search-policy doctrine at
`docs/search-policy.md` and `docs/invariant-challenge.md` executing
as designed.

## Evidence chain

### Doctrine (`docs/invariant-challenge.md` v43 / decision D5):

> Boundary deltas are now consumed by search-policy derivation. Every
> confirmed **separation-class** delta (`counterexample-separation`,
> `constructibility`) earns one `expand`/`refuted_boundary`
> directive at its predicate fingerprint, with provenance naming
> the confirmed challenge. **Bookkeeping-class deltas
> (`contrast-collapse`, `support-recount`) record epistemic defects
> of the claim, not domain structure, and earn no directive**;
> split/merge deltas flow through the child invariants' own
> lifecycle.

### Code (`internal/policy/engine.go`)

The delta-class distinction is realized as `ExpansionBearingDelta`:

```
func ExpansionBearingDelta(kind string) bool {
    switch kind {
    case "counterexample-separation", "constructibility":
        return true
    }
    return false
}
```

`contrast-collapse` and `support-recount` return false, so no
`RefutedBoundaryEvidence` is ever produced for them; the
downstream policy engine emits no `expand` directive.

### Regression asserting the doctrine (`internal/policy/boundary_test.go`)

Line 53: `t.Fatal("a bookkeeping-class delta must not admit an expand proposal")`.
This is a permanent guard against silently promoting bookkeeping
deltas.

### Live SGO (M7)

The persisted state of `equals(construction, constructive)`
(`inv_01M2BCJZBSVP2XSCFNTRCT3D40`) matches the design exactly:

- `challenge_boundary_deltas` row: `kind = "contrast-collapse"`,
  `predicate_fingerprint = b53548d1...`, `condition =
  "equals(construction,constructive)"`, `child_fingerprints = null`.
- `invariant_lineage`: zero rows involving this invariant as parent.
- `invariant_current_state`: `weaken` at 2026-09-12T18:05:17.
- No subsequent challenge, no derived child, no state transition
  beyond `challenged → weaken`.

pvnp exhibits the same pattern for all three of its weakened
posture-axis candidates: three `contrast-collapse` deltas
persisted, zero lineage rows.

## Why the design chooses this

Restated from the doctrine:

- A `contrast-collapse` finding says "the predicate does not
  separate failure-side from success-side outcomes: at least one
  success family satisfies it." That is a statement about the
  **claim's fitness** as a discriminator, not about the domain's
  mechanism space.
- A `counterexample-separation` finding says "we found a domain
  failure member that violates the predicate, and its structural
  distance from the support families is $δ$." That is a statement
  about the **domain**.
- Search policy exists to bias domain search. Bookkeeping deltas
  don't inform domain search because they don't describe domain
  structure; they describe how well the analyst's claim fits it.
- Refinement (weakening → narrower child) requires knowing *how* to
  narrow. Code-owned refinement can only apply generic splits
  (via `VerifySplit` on `OpAny`/`OpAll` roots); it cannot invent a
  domain-specific narrowing predicate. For `OpEquals(<axis>, <val>)`
  candidates, no code-owned split rule applies (no boolean children
  to partition). Any narrowing must come from a model-authored
  challenger or a fresh mining pass. This is deliberate: the
  code refuses to invent domain structure.

## Design observation (recorded, not a defect)

The miner's emission gate (`internal/provider/invariant_fixture_derive.go:derivePostureAxisProposals`)
and the `success-preserving` probe (`internal/invariant/challenge.go:VerifySuccessPreserving`)
apply thresholds on different scales:

| Stage | Threshold | Effect |
|---|---|---|
| Miner emit | `fPrev > sPrev` AND `fc ≥ 2` | Directional discrimination + support |
| success-preserving | `∃ eligible success family every member of which satisfies pred` | Binary zero-cell on success side |

The miner admits directional predicates with `sPrev > 0`; the probe
weakens any predicate with `sPrev > 0` at singleton contrast
populations. This asymmetry is exactly what correction 7's CMA
statement captures. It is not a bug — it is a legitimate design
choice: the miner emits candidates in a broader class; the
challenger enforces a stricter class for `surviving`.

For pilot-004's research questions, this asymmetry matters:

- Every `contrast_observed` `equals(axis, value)` candidate the
  miner emits AND which has at least one preserving success
  reaches `weaken` and stays there.
- A domain with any success-side heterogeneity across posture
  axes will not produce `surviving` enum-axis invariants unless
  a truly zero-cell axis-value exists.
- This is the reviewer's overarching point restated at the code
  level: the pipeline's decision rule for `surviving` is a
  zero-cell rule, not a directional-discrimination rule.

## What this does NOT establish

- **Does not** recommend implementing an automatic split for
  `contrast-collapse` weakened candidates. The doctrine explicitly
  routes such refinements through model-authored challengers.
  A code-owned automatic split for `OpEquals` predicates would
  require inventing a splitting axis, which the design refuses.
- **Does not** propose changing `ExpansionBearingDelta` to admit
  bookkeeping deltas. The doctrine explicitly excludes them and
  the regression test enforces the exclusion.
- **Does not** propose loosening `VerifySuccessPreserving`'s
  zero-cell threshold. Correction 7's CMA statement (and its
  live-data instance on M7 (a‴-B)) confirms the current behavior
  as designed.
- **Does not** re-open any of the (l-coh) corrections. This is a
  design-inspection loop, not a research correction.

## Consequences for the scorecard

- `(weaken-repair-policy)` closed as **design-verified, no gap**.
- H3 gate: preserved.
- No code changes, no admission rules, no verifier tiers, no
  policy directives, no wire schema fields.
- Adds one design observation to the record: the miner-emission
  threshold and the success-preserving probe's threshold apply on
  different scales by explicit design (correction 7's asymmetry).

## The reviewer's overarching frame, restated at code level

> The most important remaining distinction is between discovering
> structure in research and rediscovering the decision rule that
> labels that structure `surviving`.

This loop confirms that the "decision rule" side is the pipeline's
zero-cell rule enforced by `VerifySuccessPreserving`, and that this
rule is deliberately terminal on code-only campaigns (no automatic
domain-refinement). Any research-usefulness claim about `surviving`
invariants inherits this labeling geometry; the pipeline does not
try to soften the geometry retrospectively when it produces
sparse `surviving` results.
