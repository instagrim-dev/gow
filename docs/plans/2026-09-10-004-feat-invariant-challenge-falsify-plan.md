---
artifact_contract: ce-unified-plan/v1
artifact_readiness: implementation-ready
execution: code
product_contract_source: ce-plan-bootstrap
origin: GitHub issue #13 (instagrim-dev/newf) — "Challenge and falsify candidate failure invariants" (to be filed)
epic: EPIC.md milestone M4.3 — Challenge and falsify candidate invariants
depends_on: >
  GitHub issue #12 (M4.2 candidate-invariant mining) — planned in
  docs/plans/2026-09-10-003-feat-candidate-invariant-mining-plan.md, lands the
  candidate_invariant / invariant_revision / invariant_support_* tables. NOTE:
  that plan reserved migration v10 for mining, but v10 was consumed by the #11
  cluster-identity hardening (input_set_hash + family coherence + mixed outcome).
  Mining therefore lands at v11 and THIS slice lands at v12. Also builds on #11
  (cluster/failure-space, v7/v8) and #9 (canonicalization/comparison, v4/v5/v6).
created: 2026-09-10
plan_type: feat
---

# feat: Challenge and falsify candidate failure invariants

## Summary

Implement the next epic slice (EPIC.md **M4.3 — Challenge and falsify candidate
invariants**): take the **`CandidateInvariant`** records produced by M4.2 (#12)
and **attack** each one before it may influence frontier allocation, recording
every attack as a durable, typed **`Challenge`** that produces **evidence and an
explicit state transition** — never critic prose.

This is the toolbox `challenge` / `falsify` operator (`docs/toolbox-dsl.md`):

```text
challenge  CandidateInvariant -> Challenge
falsify    Claim              -> Evidence

challenge IF-03 {
  falsify known_counterexample
  falsify synthetic_counterexample
  contrast successes
  split  lower_abstraction
  merge  higher_abstraction
  test   sampling_bias
}
```

The governing constraint is epistemic and is the whole point of the milestone:

> A challenge **attacks** an inferred invariant rather than rewarding agreement
> with it. Every challenge yields deterministic *evidence* plus a *state
> transition* on the invariant's append-only lifecycle. `established` is
> reachable **only** via evidence stronger than model consensus
> (`ModelJudgment != Verification`, `AGENTS.md`); this slice never lets a model
> self-certify an invariant into `established`.

The exit condition (EPIC.md M4.3): **only challenged invariants can influence
search policy.** After this slice, a `CandidateInvariant` carries a queryable
current state in `{proposed, challenged, surviving, weaken, falsified,
established}`, and M5.1 frontier generation may read only `surviving` (or, where
independent evidence exists, `established`) invariants.

This slice **consumes** the #12 substrate (`invariant_revision`,
`candidate_invariant`, `invariant_support_cluster`, `invariant_support_evidence`)
and the #11/#9 substrate it grounds back into (`mechanism_clusters`,
`cluster_members`, `mechanism_signatures`, `signature_outcomes`). It adds a
**`Challenger` provider role** (with a deterministic fixture), the revisioned
append-only **challenge + state-transition** persistence layer (migration
**v12**, matching the `docs/persistence.md` blueprint), pipeline services, and
CLI surface. It runs entirely offline in CI against project-authored fixtures.

---

## Problem Frame

After #12, `newf` can emit typed candidate invariants that name conserved failure
structure across independent families. But a candidate invariant is, by
`AGENTS.md`, **a hypothesis, not a law**. Left unchallenged it is exactly the
failure mode the whole project exists to avoid: sophisticated-sounding prose that
has never been attacked, silently steering search.

M4.3 turns each candidate into something that has *survived adversarial
pressure*, or been *weakened / split / merged / falsified*, with the reason
recorded as evidence. Concretely, the slice must answer, per candidate:

- Is there a **known** failed approach (already in the atlas) that violates the
  invariant? → `known-counterexample`.
- Can we **construct** a synthetic failed approach that violates it? →
  `synthetic-counterexample` (persisted as a `synthetic_artifact`).
- Does a **success / partial-success** family preserve it anyway (so it does not
  discriminate outcome)? → `success-preserving`.
- Does it **split** at a lower abstraction (two invariants masquerading as one)?
  → `split`.
- Do several candidates **merge** at a higher abstraction? → `merge`.
- Is its support an artifact of **sampling / publication bias** (support that
  collapses once #11 redundancy is applied)? → `bias-critique`.

Each challenge type maps to the blueprint's
`challenge_type IN ('known-counterexample','synthetic-counterexample',
'success-preserving','split','merge','bias-critique')` and drives a transition on
the state machine already specified by `invariant_state_transition`'s trigger.

### Governing constraints (from `AGENTS.md`, `docs/abstraction-safety.md`, EPIC.md M4.3, `docs/toolbox-dsl.md`, `docs/persistence.md`)

- **No silent epistemic promotion; `established` is code-gated.** The provider
  may propose a verdict, but the transition to `established` is permitted by the
  pipeline **only** when an independent, non-model evidence record supports it
  (`invariant_support_evidence.relation = 'supports'` sourced from a deterministic
  or independently-sourced `evidence_record`, per the verification hierarchy).
  Model agreement alone can reach at most `surviving`.
- **A challenge produces evidence, not prose.** Every `invariant_challenge` row
  is backed by a concrete artifact: a known counterexample points at real
  `cluster_members` / `evidence_record`s; a synthetic counterexample persists a
  `synthetic_artifact`; a success-preserving challenge points at a
  success/partial-success family; a bias-critique records the recomputed
  independent-support count. `result_summary` is a label over that evidence,
  never a substitute for it.
- **Deterministic support is recomputed by code.** Counterexample and
  bias-critique challenges **recompute** the invariant's independent support from
  persisted signatures (reusing #11 redundancy structure and #12's independence
  threshold). If the model claims a counterexample that the deterministic check
  does not confirm, the challenge is recorded as **unconfirmed** and yields **no**
  weakening/falsifying transition. `ModelJudgment != Verification`.
- **Append-only, monotonic state machine.** State lives only in
  `invariant_state_transition`, guarded by the blueprint trigger:
  `proposed→challenged`; `challenged→{surviving,weaken,falsified}`;
  `surviving→{challenged,weaken,falsified,established}`;
  `weaken→{challenged,surviving,falsified}`. `transition_seq` is allocated
  atomically via `invariant_transition_counter`. No table is ever rewritten;
  `invariant_current_state` is the read view.
- **Abstraction safety with round-trip grounding.** `split`/`merge` challenges
  create `invariant_lineage` rows (`split|merge|weaken`) and require the child
  invariant(s) to be **groundable** back to concrete member signatures
  (`docs/abstraction-safety.md`): a split that cannot ground each child into a
  distinct concrete case is rejected. Merge must preserve predictive
  discrimination; a merge that erases the axis separating outcomes is rejected.
- **Independence over count.** Bias-critique applies #11 redundancy: support that
  is really one repeatedly-published idea collapses to a single independent
  family and can drop a candidate below #12's `candidate_obstruction` threshold,
  transitioning it toward `weaken`.
- **Provenance + replay.** Each challenge run records provider/model/config
  metadata and raw request/response payloads (reusing the #7
  `provider_invocations` discipline) so a challenge campaign is auditable and
  replayable.
- **Deterministic fixtures / offline CI.** The default `Challenger` provider in
  tests is a deterministic fixture; no network/model call in CI.
- **Additive, immutable persistence.** New tables in migration **v12** with the
  repo's `RAISE(ABORT, ...)` immutability triggers (challenges and transitions
  are immutable rows; the *sequence* of transitions is the mutation surface). No
  edits to v1–v11.

---

## Scope

### In scope

1. **`Challenger` provider role** (`internal/provider`): given a candidate
   invariant + its persisted support and the problem's atlas, propose challenges
   of each type with a natural-language rationale and a *claimed* verdict.
   Deterministic fixture-backed for tests; records role + model/provider/config.
2. **Deterministic challenge verification** (`internal/invariant` or extend the
   #12 package): for each proposed challenge, code confirms/denies it against
   persisted signatures (counterexample membership, success-preserving family
   existence, recomputed independent support for bias-critique, groundability for
   split/merge).
3. **Persistence (migration v12)**: `invariant_challenge`,
   `invariant_transition_counter`, `invariant_state_transition` (+ validating
   trigger + `invariant_current_state` view), `invariant_challenge_source_evidence`,
   `synthetic_artifact`, `invariant_challenge_synthetic_artifact`,
   `invariant_lineage` — exactly the `docs/persistence.md` blueprint, all
   immutable by trigger.
4. **Pipeline services**: `ChallengeInvariant` (single), `ChallengeAll`
   (campaign over a problem's `proposed` candidates), plus read services
   `ShowInvariantState` / `ListInvariantsByState`.
5. **CLI**: `newf challenge <invariant-id>`, `newf challenge --problem <id> --all`,
   `newf invariant state <invariant-id>`, `newf invariant list --problem <id>
   [--state surviving]`. All `--json` with stable contracts.
6. **Docs**: new `docs/invariant-challenge.md`; extend `docs/persistence.md`
   "Implemented …" section to mark the challenge tables shipped; EPIC.md M4.3
   status.
7. **Tests**: provider fixture; deterministic-verification unit tests; state-
   machine trigger tests (every legal + representative illegal transition);
   end-to-end challenge campaign integration test; `established` code-gate test.

### Out of scope (explicit)

- M5.1 frontier generation (only the `surviving`/`established` read surface it
  will consume is delivered).
- Search-policy mutation (M6.2).
- Any *new* candidate mining (M4.2 / #12 owns creation).
- Network model calls; live provider adapters beyond the fixture + interface.

---

## Key Technical Decisions

- **KTD-1 — State lives only in transitions.** No `current_state` column on
  `candidate_invariant`; the sole source of truth is `invariant_state_transition`
  read through `invariant_current_state`. Prevents drift between a cached column
  and the ledger.
- **KTD-2 — `established` requires non-model evidence.** The pipeline refuses a
  `surviving→established` transition unless a supporting `evidence_record` exists
  whose source ranks above single-model judgment in the verification hierarchy.
  Encoded as a pipeline precondition **and** documented; the DB trigger allows the
  transition structurally, the pipeline gates it epistemically.
- **KTD-3 — Unconfirmed challenges are inert.** A proposed challenge whose
  deterministic check fails is persisted (for audit) with `result_summary`
  = `unconfirmed`, links to no evidence, and drives **no** state transition. The
  hypothesis is neither strengthened nor weakened by a model claim code can't
  confirm.
- **KTD-4 — Campaign determinism.** `ChallengeAll` processes candidates in a
  deterministic order (by invariant id) and, within a candidate, challenge types
  in a fixed order; the fixture provider is pure, so a campaign is replayable and
  the resulting transition sequence is stable.
- **KTD-5 — Split/merge grounded or rejected.** A `split` must produce ≥2 child
  invariants each groundable to a disjoint concrete support subset; a `merge`
  must produce one child whose support is the union AND whose discrimination is
  not worse than the parents'. Failing grounding → challenge recorded
  `unconfirmed`, no lineage row, no transition.
- **KTD-6 — Bias-critique reuses #11 redundancy.** Independent support is
  recomputed with redundant members collapsed; the recomputed count is the
  evidence. This is the deterministic counterpart to the model's bias claim.

---

## Data Flow

```text
candidate_invariant (proposed, from #12)
  + invariant_support_cluster / _evidence
  + atlas: mechanism_clusters, cluster_members, mechanism_signatures, signature_outcomes
        |
        v
[ Challenger provider ]  -> proposed challenges {type, rationale, claimed_verdict}
        |
        v
[ deterministic verification (code) ]
   known-counterexample     -> confirm member violates invariant
   synthetic-counterexample -> persist synthetic_artifact; confirm it violates
   success-preserving       -> confirm a success family preserves it
   split / merge            -> confirm groundability + discrimination; build lineage
   bias-critique            -> recompute independent support (collapse redundancy)
        |
        v
[ persist ]  invariant_challenge (+ source_evidence / synthetic links / lineage)
        |     invariant_transition_counter (atomic seq)  ->  invariant_state_transition
        v
invariant_current_state (view)  ->  CLI / M5.1 read surface (surviving | established only)
```

---

## Implementation Units

- **U0. Dependency gate.** #12 (mining, v11) is landed and green. Reconcile the
  reserved-migration-number drift: mining = v11, this slice = v12. Confirm
  `candidate_invariant.initial_state` CHECK is `proposed` and support tables
  exist. (If #12 is not yet merged, this slice is blocked; record the blocker.)
- **U1. Migration v12.** Add the blueprint challenge/transition/synthetic/lineage
  tables + the `invariant_state_transition_validate_insert` trigger +
  `invariant_current_state` view + immutability triggers on the row tables.
  Extend `validateSchemaTables` with the new tables (and, per the v10 precedent,
  assert a representative column). Bump `currentSchemaVersion = 12`.
- **U2. `Challenger` provider role.** Interface + deterministic fixture provider
  in `internal/provider`; role recorded per invocation; raw payload provenance.
- **U3. Deterministic verifiers.** Pure functions (no SQL) over rehydrated
  signatures/support for each challenge type, returning `confirmed | unconfirmed`
  + the concrete evidence handle. Unit-tested in isolation.
- **U4. Store layer.** `PersistChallenge` (transactional: challenge row + its
  evidence/synthetic/lineage links + the atomic counter allocation + the
  transition row), `GetInvariantState`, `ListInvariantsByState`. Reuses
  `invariant_current_state`.
- **U5. Pipeline services.** `ChallengeInvariant`, `ChallengeAll`,
  `ShowInvariantState`, `ListInvariants`. Enforces KTD-2 (`established` gate) and
  KTD-3 (unconfirmed inert). Threads output views.
- **U6. CLI.** `newf challenge …`, `newf invariant state|list …`, `--json`
  contracts + human tables.
- **U7. Docs + EPIC.** `docs/invariant-challenge.md`; persistence + EPIC updates.

---

## Verification Contract

- `go build ./...`, `go vet ./...`, `go test ./...` green; `gofmt -l .` empty.
- **State-machine trigger tests**: assert every legal transition succeeds and a
  representative set of illegal ones (`proposed→surviving`,
  `falsified→anything`, wrong `transition_seq`, `from_state` mismatch) abort.
- **Verifier unit tests**: each challenge type confirmed on a fixture that truly
  exhibits it and `unconfirmed` on one that does not.
- **`established` gate test**: `surviving→established` refused with only model
  judgment present; permitted once an independent supporting `evidence_record`
  exists (proven a true gate by removing the evidence and observing refusal).
- **Bias-critique test**: a candidate whose support collapses under redundancy
  transitions toward `weaken`; one with genuinely independent support does not.
- **Split grounding test**: a split whose children cannot ground to disjoint
  support is recorded `unconfirmed` with no lineage/transition.
- **Campaign integration test**: seed fixture atlas → mine (fixture) → challenge
  --all → assert deterministic final states + replay idempotence.
- **Regressions proven true** by neutralizing each guard (KTD-2/KTD-3/KTD-5) and
  observing the targeted test fail.

## Definition of Done

- A `CandidateInvariant` has a queryable current state; only `surviving` /
  `established` are exposed to the (future) frontier read surface.
- No challenge exists without backing evidence; no `established` without
  independent evidence; no state except through an append-only, trigger-guarded
  transition.
- Docs + EPIC reflect shipped behavior; migration is introspective/idempotent and
  the integrity check detects an incomplete v12 upgrade.

## Risks

- **R1 — #12 not merged.** Blocks U0. Mitigation: gate explicitly; do not
  duplicate mining tables.
- **R2 — Migration-number drift** (v10 taken by clustering). Mitigation: this
  plan fixes mining=v11, challenge=v12; U0 reconciles the #12 plan's stale "v10".
- **R3 — Over-trusting model verdicts.** Mitigation: KTD-3 makes unconfirmed
  challenges inert; deterministic recomputation is the evidence.
- **R4 — `established` leakage.** Mitigation: KTD-2 pipeline gate + explicit test.

## Sources

- `EPIC.md` M4.3 (challenge families, lifecycle, exit condition).
- `docs/persistence.md` (candidate_invariant / invariant_challenge /
  invariant_state_transition trigger / invariant_current_state / synthetic_artifact
  / invariant_lineage — the authoritative schema this slice implements).
- `docs/toolbox-dsl.md` (`challenge`/`falsify` operators; "a challenge produces
  evidence and a status transition, not critic prose").
- `docs/abstraction-safety.md` (split/merge grounding + discrimination preservation).
- `AGENTS.md` (candidate-invariant discipline; verification hierarchy;
  `ModelJudgment != Verification`; no silent promotion).
- `docs/plans/2026-09-10-003-feat-candidate-invariant-mining-plan.md` (#12 substrate
  this consumes).

## Product Contract

- **CLI additions** (stable `--json`): `challenge`, `invariant state`,
  `invariant list --state`.
- **Read surface for M5.1**: `invariant_current_state` filtered to
  `surviving`/`established`.
- **Epistemic guarantee**: every state change is evidence-backed and append-only;
  `established` is unreachable by model consensus alone.
