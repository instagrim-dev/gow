# Glossary / ontology (`docs/theory`)

This is the semantic anchor for the theory series. It fixes the meaning of the
conceptual vocabulary *before* implementation terminology overloads it, and it
maps each concept onto the **legacy implementation term** that already exists in
the code and database.

> **Do not rename the database.** The code has accumulated working semantics
> around `candidate_invariant`, `frontier_proposal`, `evaluation`, and friends.
> A mass terminology refactor now would celebrate conceptual progress by
> breaking SQLite migrations. The rule is: **document the cleaner conceptual
> model first; map legacy implementation terms onto it; rename later, if ever,
> as a deliberate migration.** The conceptual term and the implementation term
> are allowed to differ, and this table is the bridge.

## Core terms

| Term | Meaning | Legacy implementation term(s) |
|---|---|---|
| `StructuralShape` | Descriptive structure of work, independent of outcome. The typed content that says *what a piece of work is*, not whether it succeeded. | `MechanismSignature` (canonical, fingerprinted); `Mechanism` fields (`representation`/`assumptions`/`operators`/`preserves`/axes) |
| `Regime` | Region of shape-space associated with an outcome distribution. | `MechanismCluster` partitioned by outcome inside a `FailureSpace` (`failure` / `partial_failure` / `partial_success` / `success` / `mixed`) |
| `ShapeClaim` | Claim that some structure is conserved across a population. Descriptive, not explanatory. | `CandidateInvariant` (its `invariant-predicate/v1` predicate) with `association_status ∈ {recurring, discriminative, unknown}` |
| `Conditioning` | The population a shape was observed over. A distinct axis (`observed_regime`) from the shape itself — the same shape observed over failures vs. successes is a different claim. | `observed_regime ∈ {failure_conditioned, success_conditioned, mixed, frontier_conditioned}`; realized by `FailureCoverage` vs `Contrast` axes; mixed families split member-wise |
| `ObstructionHypothesis` | Stronger *explanatory* claim that a shape prevents progress. Not the same as observing the shape. | `claim_role = obstruction` (evidence/operator-gated); `association_status = candidate_obstruction` — recorded only as a **flagged model hypothesis**, never code-assigned |
| `Boundary` | Structural condition separating regimes, or limiting the scope of a shape claim. | `FailureBoundary`; the counterexample that refutes a universal `ShapeClaim`; a `Contrast` violation that separates outcomes; a challenge's `boundary_delta` |
| `StructuralDelta` | The proposed change required to cross a boundary — the structural violation a proposal claims. When it separates a failure regime from a success regime it is the regime boundary hypothesis `Δ(F, S)`. | `FrontierProposal.StructuralClaim` (verified against the proposal's signature); `claim_role = boundary_hypothesis` |
| `Projection` | Translation of a shape-space move into a classical / domain candidate that can be verified. | The proposal's candidate `MechanismSignature` plus its directed-generation prose; the compile-down step ahead of `evaluate` |
| `LandingPoint` | The observed result of executing/verifying that candidate. | `Evaluation` (`verdict` + `verifier_kind` + `verification_strength`); `evaluated_failures` re-entry marker |
| `EpistemicState` | Where a claim sits on the epistemic-status axis (one of three orthogonal axes). | `InvariantState`: `proposed` / `challenged` / `surviving` / `weaken(ed)` / `falsified` / `operator_attested` |

## Secondary terms used across the series

| Term | Meaning | Legacy implementation term(s) |
|---|---|---|
| Shape-space | The space induced by attempts/failures/successes, in which `newf` searches. | The atlas of canonical signatures + clusters + failure-space |
| Domain-space | The problem's classical representation, in which `newf` verifies. | The verifier tiers in `internal/verify`; source-backed `EvidenceRecord`s |
| Success condition `C` | The structure that must *begin* to hold for progress, paired with a broken failure invariant `P`. | `SuccessInvariant` (its predicate `C`, linked to broken `CandidateInvariant` `P`) |
| Search policy | Persisted, versioned bias over where to search next. | `SearchPolicyRevision` + directives (`prefer`/`avoid`/`expand`/`penalize`) |
| Geometry update | The change to shape-space caused by a landing point. | New `cluster build` after `evaluated_failures` re-entry; next `policy mutate`; next invariant revision |
| `observed_regime` | The conditioning axis of a claim. | `failure_conditioned` / `success_conditioned` / `mixed` / `frontier_conditioned` |
| `claim_role` | The interpretive standing of a shape claim, recorded separately from shape, regime, and status. | `regularity` (code-assigned only) / `obstruction` / `enabling_condition` / `boundary_hypothesis` |
| `boundary_delta` | The minimal structural condition separating a counterexample (or split axis) from a claim's support population — a `Boundary` observed while challenging. | `ChallengeResult.boundary_delta` (+ `disposition ∈ survive\|weaken\|split\|falsify`, `derived_candidate_ids`) |
| `Δ(F, S)` | The regime boundary hypothesis: the structural difference between a failure-conditioned shape and a success-conditioned shape. The most actionable derived object. | `claim_role = boundary_hypothesis`; the target a `StructuralDelta` aims to cross |

## Terms that are deliberately *not* synonyms

These distinctions are load-bearing; collapsing any of them is the failure mode
the [epistemic model](02-epistemic-model.md) exists to prevent.

- `ShapeClaim` (a structure recurs) **≠** `ObstructionHypothesis` (a structure
  prevents progress). *observed regularity ≠ obstruction.*
- `Conditioning: failure` (observed on failures) **≠** `candidate ≠
  failure-conditioned` — a candidate proposal is not part of the observed
  failure population; it is a hypothesis about shape-space.
- `ModelJudgment` (a provider proposed it) **≠** `Verification` (code or an
  independent mechanism confirmed it).
- Absence (`unobserved` / empty set → `unknown`) **≠** negation (a confirmed
  violation). An omitted field is inert, not false.
- `Projection` (shape → candidate) **≠** `LandingPoint` (candidate → observed
  result). Proposing a crossing is not the same as landing across.
- `shape` **≠** `observed_regime` **≠** `epistemic_status` **≠** `claim_role`.
  These are four independent facts about a claim; a `CandidateInvariant` is their
  product, never a single position on one ladder. "Candidate" is an
  epistemic_status (`proposed`), not a kind of invariant.

## Naming policy

1. Theory documents use the **conceptual** terms above.
2. Implementation docs (`docs/*.md` outside `theory/`), code, and schema keep
   their **legacy** terms.
3. When a theory document must refer to something concrete, it names the
   conceptual term and cites the legacy term in parentheses on first use.
4. Any future rename is a scheduled migration with its own plan — not a
   side effect of a docs change.
