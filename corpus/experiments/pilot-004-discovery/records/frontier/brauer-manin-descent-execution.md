# Brauer–Manin descent proposal — pipeline-space execution record

Recorded: 2026-09-12. Pipeline-space; consumes the predeclaration in
`brauer-manin-descent-proposal.md`.

## What was executed

**Recipe.** M7 fixture at HEAD `bee1fcc`, then:

```bash
# Predicate structure verification
./newf --json invariant show ivr_01M2B8ZNKFASFT9RARSB3R37B1
# -> confirms Contains(preserves, class_union_construction) and equivalents

# Preflight
./newf --json experiment validate-proposals \
    --problem prb_01M2B8ZMHRCJN4YVH4CP6SNXTD \
    --file corpus/experiments/pilot-004-discovery/records/frontier/brauer-manin-descent-proposal.wire.json
# -> ok:true, valid:true, proposals:1, permitted_targets: [3 inv_ ids]

# Pipeline execution
./newf --json frontier generate \
    --problem prb_01M2B8ZMHRCJN4YVH4CP6SNXTD \
    --proposals-file corpus/experiments/pilot-004-discovery/records/frontier/brauer-manin-descent-proposal.wire.json \
    --count 1
```

**Persisted state** (in `.newf/m7/newf.db`):

- Generation `fgr_01M2BABVW2SHP7R4E1YWNZ66CW` (revision 7 for this problem)
- Proposal `fpr_01M2BABVW2SHP7R4E1YZ7ETFTE`
- Proposal hash: `a0c4e06e047f84deaa4f4e82b74952f0eb0cff6a5b2b3448a4bc6fe4acb03c9a`

## Two hypotheses tested in one run

### H1 — Brauer–Manin admissibility (specific, per-proposal)

Prediction: the mechanism admits cleanly through
`provider.ParseWireProposals` → `canon.AdmitProposalSignature` → mining
persistence.

Observed:

| Metric | Value | Note |
|---|---|---|
| `admission_corrected` | 0 | No provider-supplied resolutions had to be corrected |
| `admission_downgraded` | 0 | No claims downgraded to unknown |
| `admission_stripped` | 0 | No claims stripped as inadmissible |
| `admission_rejected` | 0 | Proposal not rejected outright |
| Proposal persisted | Yes (`fpr_...`) | Full row in `frontier_proposals` |
| Mechanistic distance ordinal | `medium` | Not near any atlas family; not distant either |
| Nearest clusters | 12 (all `unknown` classification, `medium` proximity) | Reflects atlas-wide mechanistic novelty; expected for a genuinely new family |
| Rank | 0 | Only proposal in this run |

**Verdict: H1 confirmed.** The Brauer–Manin proposal is a well-formed
untrusted wire input under the current pipeline vocabulary. This is
SGO-tier at the code boundary.

Cavéat: `admission_corrected: 0` means no *resolution state change* was
recorded — but per `internal/canon/admission.go` this counts only when
the resolution result *differs from* the provider-supplied claim. An
untrusted wire's claims always come in with `State: ResolutionUnknown`
(see `internal/provider/untrusted_proposer.go` `wireSignature` line 271).
If code resolution also returns `ResolutionUnknown` (label absent from
vocabulary under this field kind), state didn't change and no correction
is counted. So `admission_corrected: 0` here is compatible with the
labels *staying* Unknown, not necessarily with the labels *resolving to
canonical IDs*. Which of the two happened is not visible in the summary
counters.

### H2 — untrusted wire proposals cannot verify absence on `Contains(preserves, X)` invariants (structural, load-bearing)

Prediction: all target verdicts return `unknown`, not `violated`. This is
a design property of the pipeline, not a proposal-specific outcome.

Observed:

```json
{
  "violates_any_target": false,
  "targets": [
    {
      "invariant_id": "inv_01M2B8ZNKFASFT9RARSDT5TY3R",
      "verdict": "unknown",
      "violated": false
    },
    {
      "invariant_id": "inv_01M2B8ZNKFASFT9RARSG8X978R",
      "verdict": "unknown",
      "violated": false
    }
  ]
}
```

**Verdict: H2 confirmed.** Both explicit L3 and L4 targets returned
`unknown` verdicts. Combined with the joint-crossing proposal
(`fpr_01M2B919QNAWFFM5A31KKD2YB1` at HEAD `66c96a5`) which also returned
`unknown` for all 3 targets, we have two independent data points on
`Contains(preserves, X)` predicates from wire-authored proposals with
distinct mechanism content, distinct labels, distinct
`admission_corrected` counts — and **identical `unknown` verdicts**.

## The mechanism behind H2 (CMA, traceable in the code)

Combine two facts from the code base:

**Fact 1** — `internal/canon/admission.go` lines 108–124:

```go
if res.CompletenessStripped || sig.SetFieldCompleteness != nil {
    out.SetFieldCompleteness = map[domain.FieldKind]domain.FieldCompleteness{
        domain.FieldRepresentation:  domain.CompletenessUnobserved,
        domain.FieldOperator:        domain.CompletenessUnobserved,
        domain.FieldAssumption:      domain.CompletenessUnobserved,
        domain.FieldPreserves:       domain.CompletenessUnobserved,
        domain.FieldBreaks:          domain.CompletenessUnobserved,
        domain.FieldAuxiliaryObject: domain.CompletenessUnobserved,
    }
}
```

For any untrusted proposal, `SetFieldCompleteness` is systematically
reset to `Unobserved` across all set-fields.

**Fact 2** — `internal/invariant/predicate.go` lines 448–456:

```go
// Value absent. This is a verified negative ONLY when the field was
// exhaustively extracted; otherwise the value could be absent merely
// because the field was never (fully) recorded, so absence is an
// epistemic gap, not evidence (F3). A missing completeness marker
// defaults to unobserved, so the safe (non-inflating) answer is unknown.
if sig.FieldCompleteness(setFieldKind(n.Field)) != domain.CompletenessComplete {
    return VerdictUnknown
}
return VerdictViolates
```

A `Contains(preserves, X)` predicate returns `VerdictViolates` **only**
when both:
- `X` is absent from the signature's resolved `preserves` claims, **AND**
- `FieldPreserves` is marked `CompletenessComplete`.

Fact 1 guarantees the second condition is false for every wire proposal.

**Composition.** Wire proposals systematically cannot return `VerdictViolates` against `Contains(preserves, X)` invariants. The best achievable is `VerdictSatisfies` (by explicitly claiming to preserve X) or `VerdictUnknown` (by any other means).

This is verified by two runs and is CMA-tier at the code boundary.

## Why the pipeline is designed this way

The design rationale is stated in the predicate.go comment quoted above
(F3 discipline: "absence is an epistemic gap, not evidence"). An
untrusted provider cannot make a claim about extraction completeness —
the provider has no basis to assert "my listed preserves are exhaustive
for this mechanism". Any such claim would be a `ModelJudgment` about
extraction fidelity masquerading as a `Verification` result — exactly
what AGENTS.md §Epistemic invariants prohibits.

The design is therefore not a bug. It is an operational instantiation of
`ModelJudgment != Verification`: to promote an untrusted proposal to a
`violated` verdict against a preserves-shape invariant would require the
proposal itself to attest that its declared preserves-list is exhaustive.
That attestation is not the provider's authority; it belongs to the
operator (via `newf evidence admit` under S2's typed observation-kind
rules) or to a trusted code-derived generator.

## What kind of frontier proposal CAN return `violated` on this corpus?

Predicate shapes that DO return `VerdictViolates` from a wire-authored
proposal (with no operator attestation), based on
`internal/invariant/predicate.go`:

| Predicate shape | Return `Violates` when… | Reachable from wire? |
|---|---|---|
| `Contains(preserves, X)` | X absent AND completeness = Complete | **No** — completeness stripped |
| `Contains(breaks, X)` | Same as above on breaks | **No** — same mechanism |
| `Boundary(id, relation)` | Boundary id present under different relation | Not exposed via wire schema |
| `Equals(locality, V)` | Wire declares different locality value | **Yes** — locality is declared verbatim on the wire |
| `In(construction_mode, [V1, V2])` | Wire declares value outside the set | **Yes** — same |
| `Not(Contains(preserves, X))` | X present in preserves (resolved) | **Yes** — declare X to satisfy inner Contains → outer Not violates |

The M7 corpus's three surviving invariants are all `Contains(preserves,
X)`, so none of them are wire-attackable to `Violates`. Enum-axis
invariants (Locality, ConstructionMode, UncertaintyMode) or `Not`-form
invariants could be attacked, but the mining pipeline at
`mechanism/v3` did not produce any.

## Downstream consequences (proposed, not applied here)

1. **Rename the Recommended-Next-Action path (a).** The stated aim
   ("engages the pipeline's admission+ranking+persistence path, and
   lands on invariants that are in the live corpus") is achieved. But
   the stated hoped-for outcome ("produces code-verifiable violation
   verdicts (not `unknown`)") is unreachable via wire-authored
   proposals against `Contains(preserves, X)` invariants. Either:
   - The next iteration mines invariants of a shape that IS
     wire-attackable (enum-axis or Not-form), or
   - The next iteration uses a trusted code-derived generator, or
   - The next iteration accepts that the wire path produces
     `admission-ranked-persisted` proposals with `unknown` verdicts as
     the code-owned artefact, and operator attestation via
     `newf evidence admit` is where truth-tier promotion happens.
2. **The joint-crossing proposal at `fpr_01M2B919...D2YB1` and this
   Brauer-Manin proposal at `fpr_01M2BABV...TFTE` are structurally
   equivalent in verdict.** Both are ranked, persisted, admitted
   frontier proposals with `unknown` verdicts against `Contains(preserves,
   X)` invariants. Their differences are:
   - Content (joint-crossing attacks reciprocity/rate; Brauer-Manin
     attacks class-union/identity)
   - Admission_corrected (1 vs 0)
   - Mechanistic distance ordinal (both `medium`, but nearest-cluster
     patterns differ)
3. **The Brauer-Manin proposal itself is still a legitimate frontier
   direction.** The pipeline can't code-verify its violation claims, but
   the mathematical content (Colliot-Thélène / Cassels-Guy / Elsholtz-Tao
   anchor) is defensible paper-space content. Falsification via steps
   1–4 of the predeclared cheapest-path is a research-level task
   (weeks); its persistence in the pipeline records the direction as
   ranked and admitted.

## Verification tier

- H1 execution outcomes (admission counts, persisted IDs, hashes,
  ordinals, verdicts): **SGO** — verbatim from `newf --json` CLI output.
- H2 predicted-and-observed outcome: **SGO** for both runs' verdicts.
- H2's design-rationale explanation (why the pipeline strips
  completeness): **CMA** — traced through `admission.go` and
  `predicate.go` at HEAD `bee1fcc`; comments in the source explicitly
  cite the F3 discipline.
- Downstream consequences: **PE** — model-assisted routing decisions
  that an independent operator could disagree with.
- Same same-model-family caveat as parent records — not independent
  human verification.

## Recommended pivot

Return to the CLOSURE-SCORECARD's revised recommended-next-action list.
Path (a) as written targets an unreachable outcome; suggested rewrite
below reflects H2's finding.

- **(a′)** — Author a new frontier proposal targeting an enum-axis
  invariant, IF the M7 corpus contains one. (It does not, per the
  survey above; this reduces to "mine differently first.")
- **(a″)** — Or: run `newf challenge` against the M7 surviving
  invariants (code-owned campaign path, distinct from wire-authored
  frontier proposals).
- **(c)** — Run `newf evidence admit` against the persisted joint-crossing
  proposal `fpr_01M2B919...D2YB1` under S2's typed-observation-kind rules.
  S2's predicted verdict remains "never admissible" for the
  structural-claim-failure class. Still testable end-to-end now.
- **(b)** — Materialise a Pilot-004-specific SQLite fixture. Same as
  before; separate campaign.

New addition to the ledger from this loop:

- **(d)** — Mine an enum-axis invariant on the current M7 failure
  population, e.g. `equals(locality, local)` for the class-local Mordell
  family. If such an invariant survives challenge, it becomes the first
  wire-attackable target on this corpus.
