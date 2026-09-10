---
artifact_contract: ce-unified-plan/v1
artifact_readiness: implementation-ready
execution: code
product_contract_source: ce-plan-bootstrap
origin: GitHub issue #14 (instagrim-dev/newf) — "Generate frontier proposals against surviving invariants" (to be filed)
epic: EPIC.md milestone M5.1 — Generate frontier proposals against surviving invariants
depends_on: GitHub issue #13 (M4.3 — challenge/falsify candidate invariants) — provides the `surviving` invariant state this slice targets; #12 (M4.2 candidate invariants), #11 (clustering + FailureSpace), #9 (canonicalization/comparison) upstream
created: 2026-09-10
plan_type: feat
---

# feat: Generate frontier proposals against surviving invariants

## Summary

Implement the next epic slice (EPIC.md **M5.1 — Generate frontier proposals
against surviving invariants**): produce explicit, typed **`FrontierProposal`**
records that are *mechanistically distant* from known failure families and
*explicitly target* the structure a **surviving** failure invariant preserves —
each one carrying, per EPIC.md, its **target invariant(s)**, **claimed
structural violation**, **nearest known mechanism families**, **novelty
argument**, **cheapest falsification path**, **expected information gain**, and
**estimated evaluation cost**.

This is the toolbox `break` → `frontier`/`rank` bridge (`docs/toolbox-dsl.md`):

```text
frontier := generate against surviving(failure_invariants) {
  break invariant
  counterfact minimal_change
  rerepresent orthogonally
  transfer structural_analogy
  maximize information_gain
  minimize redundancy
}
```

> `break  CandidateInvariant -> Proposal[]` — *the principal bridge from failure
> compression to frontier generation*, generating mechanisms specifically
> designed **not** to preserve a candidate failure invariant
> (`docs/toolbox-dsl.md`).

The exit condition (EPIC.md M5.1): **frontier proposals are explainably unlike
known failures and are cheap to kill when wrong.** A proposal that cannot state
which surviving invariant it violates, how it is structurally distinct from the
nearest known family, or how it would be cheaply falsified, is not a frontier
proposal — it is undirected generation, which this slice must beat.

This slice **consumes** the #13 challenge substrate (only `state = 'surviving'`
invariants may be targeted), the #11 cluster/FailureSpace substrate (for the
nearest-family / mechanistic-distance computation over persisted signatures), and
#9 canonicalization. It adds a **`Generator` provider role** (`'generate'`, with
a deterministic fixture), the frontier persistence layer (a new additive
migration), a deterministic **ranking** service, pipeline services, and CLI
surface. It runs entirely offline in CI against project-authored fixtures.

---

## Problem Frame

After #13, `newf` has invariants that survived a recorded falsification battery.
The failure-invariant loop now turns on the symmetric move (`AGENTS.md`,
"Frontier generation discipline"):

> Do not ask for another arbitrary solution attempt. Generate proposals
> **against** surviving failure invariants.

The value of this slice is *not* generating ideas — a bare model does that. The
value is generating ideas that are **auditably directed**: each proposal names
the surviving structure it intends to break, is scored on **mechanistic distance
from the known failure families** (computed by code from persisted signatures,
not asserted), and records the **cheapest path to falsify it** so that a failed
proposal still yields information. This is exactly the objective AGENTS.md and
`docs/toolbox-dsl.md` state:

```text
maximize(
    mechanistic_distance_from_known_failures
  + violation_of_surviving_failure_invariants
  + expected_information_gain
  - evaluation_cost
  - redundancy
)
```

> The objective intentionally rewards proposals that are useful even when they
> fail. A low-probability proposal can be valuable if its failure cheaply
> eliminates a large region of conceptual search-space. (`docs/toolbox-dsl.md`)

Today the repository has the full upstream chain (sources → normalization →
signatures → clusters → FailureSpace → candidate invariants → challenge
lifecycle) but **no proposal artifact**: nothing can turn a surviving invariant
into a directed, scored, falsifiable proposal linked back to the structure it
attacks. Without it, M5.2 evaluation would have nothing typed to route to a
verifier, and the M7 holdout experiment would have no `B3` (invariant-guided
generation) arm to compare against its baselines.

### Governing constraints (from `AGENTS.md`, `docs/persistence.md`, `docs/toolbox-dsl.md`, `docs/abstraction-safety.md`, EPIC.md M5.1)

- **Only surviving invariants may be targeted.** A proposal's
  `frontier_target_invariant` links must reference invariants whose *current*
  state (from #13's `invariant_current_state` view) is `surviving`. Targeting a
  `proposed`/`challenged`/`weaken`/`falsified` invariant is rejected at the
  service boundary and by a code check against the view — this is the mechanism
  that makes "only challenged invariants influence search policy" (Gate C) real
  at the generation stage.
- **No fake precision.** Per AGENTS.md, the scored components are **ordinal**, not
  invented floats: `expected_information_gain_ordinal`, `evaluation_cost_ordinal`,
  and the ranking are component-wise/ordinal (matching the
  `frontier_proposal` schema in `docs/persistence.md`, which stores ordinals). Do
  not collapse the objective to a single fabricated scalar.
- **Mechanistic distance is computed, not claimed.** The provider proposes a
  candidate mechanism + a natural-language novelty argument; **code computes** the
  nearest known cluster(s) and the mechanistic-distance ordinal via
  `internal/canon` comparison against persisted cluster representatives, and
  records them in `frontier_nearest_cluster`. A provider claim of novelty that
  code finds to be near an existing family is recorded as low-distance, not
  trusted. `ModelJudgment != Verification`; *mechanistic* novelty over surface
  novelty (AGENTS.md).
- **Structural violation must be concrete.** `structural_violation_claim` names
  which preserved property of the target invariant the proposal breaks; where the
  property is expressible over canonical signature axes, code checks that the
  proposed mechanism's signature actually differs on that axis (a
  `break`-verification analogous to M4.3's counterexample check).
- **Every proposal carries a cheapest falsification path.** Non-empty
  `cheapest_falsification_path` is required (NOT NULL in schema); a proposal with
  no stated way to be killed is rejected — this is what makes failure informative
  and feeds the atlas (`Failure != UselessOutput`).
- **Redundancy is penalized, not silently emitted.** Two proposals with the same
  `proposal_hash` collapse (UNIQUE(problem_id, proposal_hash)); near-duplicate
  proposals (same nearest family + same target + same violated axis) are down-
  ranked by the deterministic ranker, matching the policy bias against
  `cosmetic_rerepresentation`.
- **Holdout leakage discipline is wired but optional here.** The schema links a
  `frontier_generation_run` to an optional `holdout_leakage_check`. This slice
  persists the nullable link and, when a holdout cutoff is active, refuses to
  target/justify a proposal using post-cutoff material — the full check is M7, but
  the hook and the "no post-cutoff leakage into justification" guard are honored
  now so M7 does not have to retrofit generation.
- **Provenance + replay.** Each generation run records provider/model/config +
  raw request/response payloads (`provider_invocations`, role `'generate'`), and
  is append-only: re-running produces a new run + new proposals (dedup by hash),
  never a rewrite.
- **Additive, immutable persistence.** New tables in a migration **after M4.3's
  challenge tables**, with the repo's `RAISE(ABORT, …)` immutability triggers. No
  edits to prior migrations.
- **Deterministic fixtures / offline CI.** The default `Generator` provider in
  tests is a deterministic fixture; no network call in CI.

---

## Scope Boundaries

### In scope

- `Generator` provider role (`'generate'`) interface + deterministic fixture.
- Frontier persistence: `frontier_generation_run`, `frontier_proposal`,
  `frontier_target_invariant`, `frontier_nearest_cluster` (transcribed from
  `docs/persistence.md` 534–569, adapted to live table names), plus the nullable
  `holdout_leakage_check` link column.
- Deterministic **mechanistic-distance / nearest-family** computation over
  persisted cluster representatives (`internal/canon`), and a deterministic
  **ranker** implementing the ordinal objective (component-wise, no fake scalar).
- A `break`-verification check: where the violated property is a canonical axis,
  confirm the proposed mechanism's signature differs on it.
- Pipeline services: `GenerateFrontier` (against one/all surviving invariants for
  a problem), plus `proposal list/show` readers.
- CLI surface (see below).
- Fixtures + regressions + one offline end-to-end integration test.
- Docs: a new `docs/frontier-generation.md` + EPIC.md M5.1 status update.

### Deferred to follow-up work

- Evaluation / verifier routing and recording proposal `result` (M5.2 —
  `frontier_proposal.result` stays NULL here).
- Full holdout leakage enforcement + `holdout_set`/`holdout_leakage_check`
  population (M7); this slice only persists the nullable link + the no-post-cutoff
  guard.
- Learned/scalar scoring or auto-tuned weights (only ordinal, component-wise now).

### Out of scope (per issue and EPIC.md, do not implement)

- `break` producing *executed* mechanisms or any solving attempt.
- Search-policy mutation (M6.2) — proposals do not yet feed policy.
- Embedding/vector novelty — distance is deterministic over canonical signatures.
- Editing M4.2/M4.3 logic or their providers.

---

## Requirements

1. `GenerateFrontier` targets only `surviving` invariants (checked against
   `invariant_current_state`); targeting any other state is rejected.
2. Each proposal persists all EPIC.md-required fields: target invariant(s),
   `structural_violation_claim`, nearest known families (`frontier_nearest_cluster`
   with proximity ordinal), `novelty_argument`, `cheapest_falsification_path`
   (required, non-empty), `expected_information_gain_ordinal`,
   `evaluation_cost_ordinal`.
3. Nearest-family + mechanistic-distance ordinal are computed by code from
   persisted signatures, not taken from the provider.
4. Proposals dedup by `proposal_hash` (UNIQUE per problem); the ranker down-ranks
   near-duplicates and orders by the ordinal objective.
5. A generation run records provider/model/config + raw payload provenance and is
   append-only; role `'generate'` accepted by `provider_invocations`.
6. When a holdout cutoff is active, a proposal justified using post-cutoff
   material is refused.
7. CLI can generate, list, and show proposals with full provenance, human +
   `--json`.
8. The whole slice runs offline in CI against deterministic fixtures.

---

## Key Technical Decisions

- **KTD-1 — Transcribe the frontier schema from `docs/persistence.md`.** The
  tables, ordinal fields, and `UNIQUE(problem_id, proposal_hash)` dedup are
  already the contract; the migration copies them, fixing FK names to the shipped
  `mechanism_clusters` / `candidate_invariant` / cluster-run tables and adding the
  nullable `holdout_leakage_check_id` link now (populated in M7).
- **KTD-2 — Ordinal, component-wise scoring only.** The ranker sorts proposals by
  a documented lexicographic/ordinal comparison over
  (violation strength, mechanistic distance, expected information gain, −cost,
  −redundancy). No weighted float. This is deliberate per AGENTS.md ("Do not
  fabricate fake precision") and keeps the objective auditable.
- **KTD-3 — Code owns distance + violation; provider owns candidate + argument.**
  Mirrors the M4.3 separation: the `Generator` proposes a mechanism spec + prose;
  `internal/canon` computes nearest cluster + whether the claimed violated axis
  actually differs. A confident-but-near proposal is recorded honestly as
  low-distance and down-ranked.
- **KTD-4 — `cheapest_falsification_path` is mandatory.** Enforced NOT NULL +
  non-empty at the service boundary. A proposal with no kill path is rejected,
  because its failure could not add information (the whole point of the objective).
- **KTD-5 — Fixture-first provider** enables a fully offline end-to-end test:
  seed → mine (fixture) → challenge (fixture, reach `surviving`) → generate
  (fixture) → assert scored, directed, deduped proposals linked to the surviving
  invariant and its nearest family.

---

## High-Level Technical Design

### Data flow

```text
invariant_current_state (state='surviving', from #13)
  -> GenerateFrontier(problem, [invariant_ids], count)
       -> Generator provider proposes { mechanism_spec, structural_violation_claim,
                                        novelty_argument, cheapest_falsification_path,
                                        eig_ordinal, cost_ordinal }
       -> code: nearest cluster(s) + mechanistic_distance ordinal via canon
       -> code: verify claimed axis violation against target invariant support
       -> write frontier_generation_run, frontier_proposal (dedup by hash),
          frontier_target_invariant, frontier_nearest_cluster
       -> deterministic ranker orders the run's proposals (ordinal objective)
  -> proposal list/show reflect scored, directed proposals
```

### Persistence shape (new tables — additive, on top of M4.3)

Transcribed from `docs/persistence.md` 534–569 (adapted names):
`frontier_generation_run`, `frontier_proposal`, `frontier_target_invariant`,
`frontier_nearest_cluster`, with immutability triggers; nullable
`holdout_leakage_check_id` retained for M7.

## Output Structure

```text
newf frontier generate --problem <id> [--against <invariant-id>]... [--count <n>]
newf frontier list --problem <id> [--target <invariant-id>]
newf frontier show <proposal-id>
```

Human output leads with the ranked proposals (violation → nearest family →
kill-path), each tagged with its target surviving invariant; `--json` emits full
records + provenance + the computed distance/violation verdicts.

## Implementation Units

- **U0. Dependency gate:** M4.3 (challenge lifecycle + `invariant_current_state`)
  landed green; record its migration version so this slice's migration is `+1`,
  and confirm the `provider_invocations.role` CHECK includes `'generate'` (widen
  if not).
- **U1.** ID classes + domain scaffolding (`FrontierGenerationRunID`,
  `ProposalID`, typed ordinal enums, proposal-hash canonicalization).
- **U2.** Migration (frontier tables + triggers), transcribed from persistence
  blueprint; store writer/reader with hash dedup in a transaction.
- **U3.** Pure deterministic nearest-family / mechanistic-distance + violation
  verifier over canonical signatures (no SQL/provider), and the ordinal ranker.
- **U4.** `Generator` provider role interface + deterministic fixture.
- **U5.** `GenerateFrontier` pipeline service (surviving-only guard, code-owned
  scoring) + `proposal list/show` readers.
- **U6.** CLI wiring (thin).
- **U7.** Fixtures + regressions (surviving-only enforcement; distance computed
  not trusted; missing kill-path rejected; hash dedup; ordinal ranking order) +
  one offline end-to-end test.
- **U8.** Docs: `docs/frontier-generation.md` + EPIC.md M5.1 status.

## Verification Contract

- `go build ./...`, `go test ./...`, `gofmt -l .` empty.
- Storage tests: proposals immutable; hash dedup enforced; target links reject
  non-surviving invariants.
- Pipeline tests: a provider proposal whose novelty is code-refuted is stored as
  low mechanistic distance and down-ranked (KTD-3); a proposal missing a
  falsification path is rejected (KTD-4); ranker order matches the documented
  ordinal objective (KTD-2).
- End-to-end (offline): fixtures → mine → challenge→surviving → generate → ranked,
  directed, deduped proposals linked to the surviving invariant + nearest family.

## Definition of Done

A fresh clone can, offline, take a surviving invariant and produce a ranked set of
directed frontier proposals — each stating the invariant it violates, its nearest
known family (computed, not claimed), a mandatory cheapest falsification path, and
ordinal information-gain/cost — such that M5.2 can route them to verifiers and M7
can use them as the `B3` invariant-guided arm. No proposal targets a
non-surviving invariant; no proposal lacks a kill path; scoring is ordinal, not a
fabricated scalar.

## Risks & Dependencies

- **Hard dependency on M4.3** (`surviving` state + current-state view). U0 gates
  start.
- **Distance/violation fidelity:** the code-owned distance + break-verification
  (U3) is the load-bearing guard that keeps this from degenerating into
  undirected generation; it must not be bypassable by a confident provider.
- **Ordinal-vs-scalar temptation:** resist collapsing the objective to a float;
  U7 pins the ordinal comparison.
- **Holdout hook:** persist the nullable link + no-post-cutoff guard now to avoid
  an M7 retrofit, but do not implement the full leakage check here.

## Sources & Research

- `AGENTS.md` — "Frontier generation discipline", "Mechanistic novelty over
  surface novelty", "Do not fabricate fake precision", verification hierarchy.
- `docs/persistence.md` (lines 534–569, and 99 for the provider-role set incl.
  `'generate'`) — authoritative frontier proposal schema.
- `docs/toolbox-dsl.md` — `break`/`frontier`/`rank` operators + the maximize
  objective.
- EPIC.md M5.1 + Gate C→D transition; end-to-end CLI shape; artifact chain.

## Product Contract preservation

Domain stays free of SQL/Cobra; SQLite behind `internal/store`; provider behind
`internal/provider` with recorded role/model/config; no proposal may target a
non-surviving invariant; scoring is ordinal/component-wise; machine-readable
output additive and stable; `frontier_proposal.result` left for M5.2.
