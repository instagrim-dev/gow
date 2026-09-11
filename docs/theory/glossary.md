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

## Naming canon (roles, not synonyms)

The project's names are **assigned different jobs rather than treated as
interchangeable names for the whole theory**. Each name changes what is
promised to the reader; none changes the evidentiary status of the research
(promoting a name is a naming revision, not evidence). Geometry of Work
remains the umbrella.

| Role | Canonical name | Owning artifact(s) |
|---|---|---|
| Broad theory (umbrella) | **Geometry of Work** | [00-geometry-of-work](00-geometry-of-work.md); the theory series |
| Compositional and relational model | **The Grammar of Attempts** | [01-semantic-model](01-semantic-model.md), [07-relational-structure](07-relational-structure.md) |
| Method | **Consequential Cartography** | [03-shape-guided-search](03-shape-guided-search.md), [How to Map the Work](../thesis/how-to-map-the-work.md) |
| Operational cycle | **The Cartographer's Loop** | The C2 algorithm (`Map → Boundary → Move → Project → Measure`); reflexive form in [10-self-referential-cartography](10-self-referential-cartography.md) |
| Practitioner instruction | **Map the Work** | [Why](../thesis/why-map-the-work.md) / [When](../thesis/when-to-map-the-work.md) / [How](../thesis/how-to-map-the-work.md) to Map the Work |
| Authored thesis / book | **The Shape of Trying** | [docs/thesis/](../thesis/) as a collection |
| Complement extension | **Counterform** | [05-complement-geometry](05-complement-geometry.md) |
| Implementation | **`newf`** | this repository |

**Retired alternative:** *Cartographic Search* is not used as the method name.
It said how search is organized; **Consequential Cartography** says what makes
that organization worth using. Public-facing subtitle when clarity is needed:
*"A Method for Choosing Work through Maps of Prior Attempts."*

### The jobs, precisely

- **The Grammar of Attempts** — how Work is composed, constrained, and
  transformed. Emphasizes composition and transformation *within and between*
  attempts, where geometry emphasizes relationships *across a population*.
  The grammar describes the moves from which the geometry is constructed —
  a proposed connection, not a claim that the implementation already supplies
  a complete generative grammar. "Grammar" here is a **conceptual model**, not
  yet an executable formalism; "attempts" include successful constructions,
  partial results, experiments, and attempts to improve the map itself — not
  merely failed solution proposals.
- **Consequential Cartography** — why a map deserves to influence subsequent
  Work: **maps are evaluated through the predictions, decisions, and checked
  consequences they support — not merely their descriptive elegance.** A map
  can earn its cost by selecting a discriminating experiment, exposing an
  invalid assumption, or supporting a justified stopping decision; an
  increasingly elaborate retrospective explanation earns no automatic credit.
  This deliberately gives up counting *better organization alone* as evidence
  of a better search method (the anti-curation constraint; falsification
  conditions 3 and 5).
- **The Cartographer's Loop** — how mapping, action, observation, and revision
  operate together. Its promise is temporal: *the result of an attempt changes
  subsequent mapping or action.* A static atlas, one-off analysis, or batch of
  proposals may be stages in the loop; none is the feedback cycle itself. Two
  uses stay distinct: the **ordinary loop** (revise the map after observing
  new Work) and the **reflexive loop** (examine whether the map's own
  representation caused a bad prediction, then test an alternative —
  [10-self-referential-cartography](10-self-referential-cartography.md)). The
  name promises neither convergence, autonomous self-improvement, nor eventual
  success: a loop can terminate on budget or because its map is not helping.

### What the grammar name obligates (and where each obligation already lives)

| Obligation | Question | Existing artifact |
|---|---|---|
| Constituents | What are the meaningful parts of an attempt? | `StructuralShape` fields (operators, assumptions, representation, preserves, breaks, auxiliary objects) |
| Composition | How do their arrangements and dependencies matter? | signature axes + [07](07-relational-structure.md)'s relational claims (Proposition R: composition can carry what no part carries) |
| Transformations | Which changes produce a meaningfully different attempt? | `StructuralDelta`; mechanistic-novelty-over-surface-novelty rule |
| Conditions | Under what assumptions is a transformation applicable? | `assumptions` field; `Boundary`; challenge dispositions |
| Interpretation | What concrete Work would realize the description? | `Projection` (classical projection) |

A load-bearing distinction the grammar must keep explicit:

> **A well-formed description of an attempt is not necessarily a realizable
> attempt. A realizable attempt is not necessarily a successful one.**

(Description → realizability is decided at `Projection`; realizability →
success is decided at `LandingPoint`. The grammar proposes constructions;
checking decides them.)

### Usage rules

1. The one-paragraph coherent description:
   > **Geometry of Work studies the structure of problem-solving attempts.
   > The Grammar of Attempts describes their composition and possible
   > transformations. Consequential Cartography uses those descriptions to
   > choose subsequent Work through the Cartographer's Loop.**
2. **Public-facing identity uses only Geometry of Work and Consequential
   Cartography**; introduce the grammar and loop when explaining the method —
   otherwise the terminology becomes harder to learn than the idea.
3. These names create **no new software subsystems or research milestones**;
   they name different commitments within the same undertaking.
4. The database and code keep their legacy terms (naming policy above applies
   unchanged).

## Complement geometry terms (hypothesis-status)

*These terms are introduced in [05-complement-geometry.md](05-complement-geometry.md).
They are at hypothesis status — formally stated but not experimentally tested.
No legacy implementation terms exist yet.*

| Term | Meaning |
|---|---|
| `PossibilitySpace` (Ω) | The ambient space of structurally plausible work on a problem, before evidence is accumulated. Defining Ω rigorously is itself a research problem. |
| `Constraint` (C_i) | The region of Ω excluded or constrained by the evidence from a single piece of work W_i. |
| `ResidualRegion` (Ω_W) | The surviving possibility-space after accumulated work: Ω \ ⋃ F_i. Its shape is determined by the collective constraints of all prior work. |
| `AntiVacuum` | A residual region that is unoccupied by prior work but structurally determined by everything that surrounds it — a void whose geometry is informative. |
| `ConstraintCollapse` (C) | The operator C(Ω | W) → R that computes the residual structural region from accumulated work. Not probabilistic collapse — constraint collapse: progressive removal of degrees of freedom. |
| `ComplementGeometry` | The dual reading of a work population: inferring shape from the space work eliminates rather than from the space work occupies. |
| `ForwardGeometry` | The ordinary GoW reading: inferring shape from the structure of completed work (clusters, invariants, regimes). Named retroactively to distinguish it from its dual. |
