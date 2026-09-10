---
artifact_contract: ce-unified-plan/v1
artifact_readiness: implementation-ready
execution: code
product_contract_source: ce-plan-bootstrap
origin: GitHub issue #16 (instagrim-dev/newf) — "Compress partial successes into success invariants" (to be filed)
epic: EPIC.md milestone M6.1 — Compress partial successes into success invariants
depends_on: >-
  M5.2 evaluation/verifier routing (commit `d03ea53`, schema v15) and the
  M4.3/M5.1 acceptance hardening (commit `a0591a4`, schema **v16** — lifecycle
  state `operator_attested`, earned survival, AdmitCandidate gate, in-tx derived
  children). This plan is migration **v17**. It consumes: `evaluations`
  (verdict + verifier_kind + verification_strength, CHECK-enforced),
  `frontier_proposals` (+ one-time `result`), `frontier_target_invariants`
  (code-verified per-target violation verdicts), `evaluated_failures`,
  `candidate_invariants`/`invariant_predicates` (typed predicates + lifecycle),
  and the #9/#11 canonical atlas. KNOWN SUBSTRATE GAP (verified against live
  main): `frontier.Proposal.ProposedSignature` exists in memory and feeds
  `proposal_hash`, but the proposed mechanism's canonical CONTENT is never
  persisted — v17 must add it (U1) before any success compression can evaluate
  a condition predicate against a successful proposal.
created: 2026-09-10
plan_type: feat
---

# feat: Compress partial successes into success invariants

## Summary

Implement EPIC **M6.1**: ask the symmetric question of the failure-invariant
loop —

> What common structure appears in proposals that cross a boundary the failure
> families could not cross?

and answer it with explicit, typed **`SuccessInvariant`** records:

```text
failure invariant: P remains preserved across failed families
success invariant: progress appears when P is broken under condition C
```

The decisive design move is that **`C` is a typed predicate and the cohort is
code-selected**. A success invariant is not "the model says breaking P helped."
It is:

- **P** — one or more *targeted* failure invariants that a proposal
  **code-verifiably violated** (`frontier_target_invariants.violated = 1`,
  computed by `invariant.Evaluate` at generation time, never claimed);
- **the break cohort** — every evaluated proposal that violated P, partitioned
  by its M5.2 verdict into **progressors** (`partial_success`/`success`) and
  **non-progressors** (`failure`/`partial_failure`);
- **C** — a typed `invariant-predicate/v1` condition, proposed by the provider
  (the `compress` operator over successes) but **evaluated by code** against
  every cohort member's persisted canonical signature: `C` must be satisfied by
  progressors and violated by (or absent from) non-progressors that broke the
  same P. Breaking P alone is not the finding — plenty of P-breakers fail;
  **`C` is what discriminates progress among P-breakers**, which is exactly why
  the EPIC says "`C` may be more useful than the raw symmetry break."

This reuses the entire M4.2 machinery (predicate contract, `Evaluate`,
epistemic composition, semantic fingerprint identity) over a new cohort — no
new predicate language, no new evaluator, no scalar scores. Everything
truth-sensitive is code-owned; the provider only authors candidate `C` ASTs and
prose renders (`ModelJudgment != Verification`).

**Epistemic discipline carried from M5.2:** a proposal's verdict has a
`verification_strength`. Success-invariant support therefore records its
**strength composition** (how many supporting verdicts were `deterministic`
vs `single-model-judgment`, etc.): three model-judged partial successes are
not the evidence three deterministic ones are, and the artifact says so
instead of laundering it.

The exit condition (EPIC M6.1) is that **partial success changes the search
model rather than merely becoming another result row**: the persisted
`SuccessInvariant` set (state `proposed`, semantic predicate-fingerprint
identity, full provenance to evaluations/proposals/broken targets) is the
typed read surface **M6.2 search-policy mutation consumes** ("prefer structures
associated with partial success"). This slice produces that surface; it does
NOT mutate policy (M6.2), challenge success invariants (the M4.3 lifecycle can
be extended later), or run holdout scoring (M7).

---

## Problem Frame

After M5.2 the system can say, per frontier proposal: *what it targeted,
whether it really broke the target (code-verified), and how it fared under the
strongest available verifier — with the verdict's epistemic strength stamped on
it*. What it cannot do is the symmetric compression the research loop turns on:

```text
failure-space -> failure invariant -> invariant break -> partial success
   -> SUCCESS INVARIANT (this slice) -> generalized frontier / policy (M6.2+)
```

Today a `partial_success` evaluation is a result row. Nothing extracts the
conserved structure across progressors, nothing distinguishes "breaking P
helped" from "breaking P while preserving effective bounds helped," and M6.2
has no typed input to prefer.

### Governing constraints (AGENTS.md, docs/toolbox-dsl.md, EPIC M6.1)

- **Success invariant = machine-evaluable predicate.** `C` is an
  `invariant-predicate/v1` AST admitted by the same `AdmitCandidate` gate
  (mining grammar + pinned-vocabulary references + no outcome reads — a `C`
  that reads the outcome axis is tautological over an outcome-partitioned
  cohort and is rejected at validation).
- **Code selects the cohort; code computes discrimination.** The break cohort
  comes from persisted, code-verified rows (`frontier_target_invariants`,
  `evaluations`); the provider never nominates its own supporting evidence.
- **No silent epistemic promotion.** Success invariants enter and stay
  `proposed` in this slice; support carries the verification-strength
  composition of its underlying evaluations; a model-judged cohort is reported
  as such. No transition machinery ships here.
- **Symmetry with failure invariants, not a fork.** Same predicate schema, same
  fingerprint identity, same immutable revisioned persistence idiom, same run
  lifecycle — a later unified challenge lifecycle must not require a redesign.
- **Deterministic fixtures / offline CI.** The compressor role has a
  deterministic deriving fixture; no network/model in CI.
- **Additive persistence (v17)** with `RAISE(ABORT)` triggers; the one
  non-additive step is the established `provider_invocations.role`
  widening (adds `'success-compress'`) via the guarded in-place
  `writable_schema` CHECK edit (v11/v13/v14/v15 precedent).

---

## Scope

### In scope

1. **U1 — Persist proposed-mechanism canonical content** (the verified
   substrate gap): new `frontier_proposal_signatures` table
   (`proposal_id` PK → canonical signature JSON + fingerprint), written at
   generation time for new proposals AND insert-if-absent for deduped existing
   proposals (enrichment by INSERT, never UPDATE — proposal rows stay
   immutable). Proposals evaluated before v17 without persisted content are
   **ineligible for compression and reported as a named visibility gap**, never
   silently included or fabricated.
2. **`SuccessCompressor` provider role** (`internal/provider`, role
   `'success-compress'`): given a code-built `CompressionRequest` (per broken
   target P: the progressor/non-progressor cohort projections), propose
   candidate conditions `{predicate C, statement, abstraction_level}`.
   Deterministic deriving fixture for CLI/CI.
3. **Deterministic compression engine** (`internal/invariant` or sibling
   `internal/success`, pure): cohort partitioning, `Evaluate(C)` over every
   cohort signature, discrimination computation
   (`ProgressCoverage` = progressors satisfying C / eligible progressors;
   `NonProgressorExclusion` = non-progressors violating C / eligible
   non-progressors — both ordinal-banded, never a fabricated scalar),
   strength composition, distinct-mechanism support (dedup by canonical
   fingerprint so one mechanism re-proposed is one support unit).
4. **Persistence (migration v17):** `success_invariant_revisions`,
   `success_invariants` (semantic identity = predicate fingerprint;
   `initial_state CHECK ('proposed')`), `success_invariant_predicates`,
   `success_invariant_broken_targets` (the P links, FK to
   `candidate_invariants`), `success_invariant_cohort_evaluations` (per-member
   verdict + role progressor|non_progressor + verification strength), all
   immutable; plus the role-CHECK widening and `validateSchemaTables`
   extension (tables + role assertion).
5. **Pipeline + CLI + run lifecycle:** `newf successes compress --problem <id>
   [--min-support <n>]`, `newf success-invariant list|show`, all `--json`,
   runs `running → completed/failed`.
6. **Tests at every seam** (engine truth table incl. strength composition and
   pre-v17 ineligibility; store idempotency/immutability/migration-from-v16;
   pipeline end-to-end mine→challenge→generate→evaluate→compress; CLI
   contract) and **docs** (`docs/success-compression.md`, persistence v17,
   EPIC M6.1 status, README loop section).

### Deferred to Follow-Up Work

- **Challenge lifecycle for success invariants** — split to the M4.3-symmetric
  follow-up; sink: extend `invariant_challenges` FKs or a
  `success_invariant_challenges` sibling; trigger: first M6.2 policy consumer
  that needs `surviving` success invariants rather than `proposed` ones.
- **M6.2 search-policy mutation** — next epic slice; consumes this read
  surface.
- **Backfill of pre-v17 proposal signatures** — only via deterministic
  re-generation (dedup + enrichment makes this a natural `frontier generate`
  re-run); never by reconstructing content from a hash.

### Out of scope

- Holdout scoring (M7), cross-discipline litmus (M8), policy persistence
  (M6.2), any state beyond `proposed`, any scalar "success quality" score.

---

## Key Technical Decisions

- **KTD-1 — The break cohort is keyed on P, not on the whole run.** For each
  failure invariant P with ≥1 code-verified violating proposal, the cohort is
  every evaluated proposal with `frontier_target_invariants(P).violated = 1`
  AND a persisted signature AND a one-time `result`. Partition by result:
  progressors (`partial_success`,`success`) vs non-progressors
  (`failure`,`partial_failure`); `unknown`/`verification_blocked` members are
  **excluded from both sides and counted as a named ambiguity**, mirroring the
  M4.2 mixed-family discipline.
- **KTD-2 — C discriminates, or it is nothing.** A candidate condition earns
  support only from code-computed coverage AND exclusion: full coverage with
  zero exclusion means C is cohort-wide boilerplate, not the progress
  condition. Both quantities are ordinal bands over exact persisted counts
  (numerators/denominators stored; bands derived).
- **KTD-3 — Strength composition is first-class.** Each cohort evaluation and
  each success invariant aggregates `verification_strength` counts of the
  member verdicts (deterministic / reproducible / independent-evidence /
  independent-critic / single-model-judgment). No count is promoted; M6.2 can
  weight policy by strength, this slice only reports honestly.
- **KTD-4 — Same identity discipline as M4.2.** `success_invariants` dedup on
  `UNIQUE(revision, predicate_fingerprint)`; revisions idempotent on
  `(problem, evaluation cohort hash, compressor_version, predicate_schema,
  min_support)` where the cohort hash is a content hash over the sorted
  (proposal_id, result, violated-target-set) tuples — re-compressing an
  unchanged cohort returns the existing revision; a new evaluation changes the
  hash and yields the next revision.
- **KTD-5 — Enrichment by INSERT, never UPDATE.** `frontier_proposal_signatures`
  is a sibling row keyed by proposal id, so persisting content for an existing
  deduped proposal violates no immutability trigger; the signature JSON is the
  canonical serialization and must round-trip to the fingerprint embedded in
  `proposal_hash` (guard test).
- **KTD-6 — Role widening follows the shipped pattern.** v17 widens
  `provider_invocations.role` to include `'success-compress'` via
  `editTableCheckInPlace`, guarded and idempotent, with a from-pre-v17
  migration test (the v13 test is the template).

---

## Data flow

```text
evaluations (verdict + strength, M5.2)  +  frontier_target_invariants (violated, M5.1)
        |                                          |
        v                                          v
[ cohort builder, code ]  per broken failure invariant P:
    progressors / non-progressors / ambiguous  (persisted signatures only; pre-v17 gap reported)
        |
        v
[ SuccessCompressor provider ]  ->  candidate conditions {C predicate, statement}
        |
        v
[ AdmitCandidate gate ]  grammar + vocabulary refs + no outcome reads (tautology guard)
        |
        v
[ compression engine, pure code ]  Evaluate(C, every cohort signature)
    -> ProgressCoverage + NonProgressorExclusion (ordinal over exact counts)
    -> strength composition; distinct-mechanism support (fingerprint dedup)
        |
        v
SuccessInvariant(proposed)  --  P links + cohort evaluations + provenance   [immutable, v17]
        |
        v
[ M6.2 search-policy mutation ]  (out of scope here)
```

---

## Implementation Units

### U1. Proposed-signature persistence (substrate gap; `internal/store` + `internal/pipeline/frontier.go`)
Migration v17 part 1: `frontier_proposal_signatures(proposal_id PK REFERENCES
frontier_proposals(id), canonical_fingerprint TEXT NOT NULL, signature_json
TEXT NOT NULL, created_at)` + immutability triggers. `GenerateFrontier`
persists the canonical serialization for every proposal in the same
transaction as the proposal row, and inserts-if-absent for deduped existing
proposals (KTD-5). Store tests: round-trip; enrichment on dedup; immutability;
a rehydrated signature re-fingerprints to the stored fingerprint.

### U2. Cohort builder (`internal/pipeline` + store readers)
`ListBreakCohorts(problem)`: join `frontier_target_invariants(violated=1)` ×
`frontier_proposals(result NOT NULL)` × `frontier_proposal_signatures`,
grouped by target invariant id; rehydrate each member's
`canon.MechanismSignature` from `signature_json`. Members without persisted
signatures are returned in a named `IneligibleUnpersisted` count (pre-v17
gap); `unknown`/`verification_blocked` results in `Ambiguous`. Store test over
a seeded end-to-end fixture.

### U3. `SuccessCompressor` provider role (`internal/provider`)
`CompressionRequest` (per P: cohort projections — canonical content, results,
strengths; structured facts only) with an order-independent `Fingerprint()`;
`CompressionResponse` (`[]ConditionProposal{predicate, statement,
abstraction_level}` + metadata + payloads); `Identity()` per the M5.2 miner
precedent. `DerivingFixtureCompressor`: for every canonical id present in ALL
progressors' preserves/operators and ABSENT from ≥1 non-progressor, propose
`contains(field, id)` — transparent, deterministic, re-verified by code like
any model proposal. Tests: determinism, fixed order, invalid-predicate
rejection, fingerprint order-independence.

### U4. Compression engine (pure; no SQL/Cobra/provider)
`CompressCohort(proposals []ConditionProposal, cohort BreakCohort, minSupport
int) []SuccessCandidate`: per condition — AdmitCandidate gate, `Evaluate` over
every member, exact counts (coverage num/den, exclusion num/den), ordinal
bands, strength composition (KTD-3), distinct-mechanism support via canonical
fingerprint dedup, semantic fingerprint identity. Unit tests: a condition all
members satisfy earns coverage but zero exclusion → reported, never inflated
(KTD-2); ambiguous member lands in neither side; strength composition of a
model-judged cohort is `single-model-judgment`-dominant and never promoted;
outcome-reading C rejected by the gate; fingerprint stability under member
permutation.

### U5. Persistence (migration v17 part 2) + writer/reader
Tables per Scope §4 with immutability triggers; `PersistSuccessRevision`
(transactional, idempotent on KTD-4's tuple), `GetSuccessRevision`,
`ListSuccessInvariants(problem)`, `LatestSuccessRevision`. Widen the role
CHECK (`'success-compress'`, KTD-6) + extend `validateSchemaTables` (five new
tables + `frontier_proposal_signatures` + role assertion). Store tests:
idempotency; re-compress after a NEW evaluation → next revision; immutability;
`proposed`-only CHECK; from-pre-v17 role-migration test; fresh-migrate schema
validation.

### U6. Pipeline + CLI + run lifecycle
`newf successes compress --problem <id> [--min-support <n>]` (run
`running → completed/failed`; empty cohort is a legitimate empty revision,
never fabricated), `newf success-invariant list --problem <id>`,
`newf success-invariant show [revision-id] [--problem <id>]`; stable `--json`
views incl. broken-target links, coverage/exclusion counts + bands, strength
composition, and the named ineligibility/ambiguity counts. Register in
`cmd/newf/root.go`.

### U7. End-to-end + regression fixtures
Full offline chain: seed → signatures → cluster → failure-space → mine →
challenge (→ surviving) → frontier generate → evaluate → **compress** →
success-invariant show; assert (a) a code-verified P-breaker with
`partial_success` yields a success invariant linking exactly that P;
(b) a P-breaker that failed evaluation lands as non-progressor and C's
exclusion counts it; (c) re-compress unchanged cohort is idempotent;
(d) new evaluation → new revision; (e) run-status failure path; (f) pre-v17
unpersisted-signature proposals surface in the ineligibility count.

### U8. Docs + gates
`docs/success-compression.md` (the symmetric-compress contract, cohort
selection, C-discrimination discipline, strength composition, M6.2 handoff),
`docs/persistence.md` v17 section, EPIC M6.1 status note, README core-loop
update. Gates: `go build ./...`, `go vet ./...`, `go test ./...` green,
`gofmt -l .` empty; boundary test at every new seam.

---

## Verification Contract

- **Substrate gap closed:** a newly generated proposal has persisted canonical
  content; a deduped re-proposal is enriched by INSERT; the stored JSON
  re-fingerprints identically (falsify: corrupt the JSON, watch the guard
  fail). Pre-v17 proposals are counted ineligible, never silently dropped or
  fabricated.
- **Code owns the cohort:** a provider cannot add/remove cohort members —
  cohort tests build from persisted rows only; a fixture condition matching
  nothing earns zero support.
- **C discriminates:** coverage AND exclusion computed from exact persisted
  counts; the all-members-satisfy condition reports zero exclusion; ordinal
  bands never replace the stored numerators/denominators.
- **No promotion:** `initial_state != 'proposed'` rejected by CHECK; strength
  composition survives round-trip; a model-judged support never reports a
  stronger strength class.
- **Symmetry preserved:** success-invariant identity is the same semantic
  predicate fingerprint; paraphrased statements collapse; permuting cohort
  order changes nothing (KTD-4 hash is order-independent).
- **Migration:** from-v16 upgrade test (role CHECK gains `'success-compress'`;
  child FKs survive — `writable_schema` edit, no parent drop); idempotent on
  fresh v17; `validateSchemaTables` rejects a partially-upgraded store.
- **Lifecycle:** compress run shows `completed`/`failed` truthfully via
  `run show`; every artifact resolves back through evaluations → proposals →
  targeted invariants → signatures → snapshots.

## Definition of Done

Given a problem with evaluated frontier proposals, `newf successes compress`
produces immutable, revisioned `SuccessInvariant` records — each a validated,
machine-evaluable condition `C` linked to the code-verified broken failure
invariants P, with code-computed progress coverage AND non-progressor
exclusion over a code-selected cohort, verification-strength composition
retained, `proposed`-only, fully offline — a typed read surface M6.2 can
consume so that partial success changes the search model rather than remaining
a result row.

## Risks

- **Sparse cohorts.** Early corpora may yield one-member cohorts; min-support
  gating + honest empty revisions prevent fortune-cookie invariants. Mitigated,
  not hidden: counts are persisted.
- **Migration-number drift (recurring).** v15 and v16 were consumed while this
  plan was drafted; U0 of implementation re-verifies `currentSchemaVersion`
  against live main before numbering (the M4.2/M4.3 lesson, twice-learned).
- **Concurrent-writer churn.** Verifier/miner signatures have changed mid-slice
  before (AssociationKind, Identity()); re-cement U3/U4 against the live
  provider/invariant APIs at implementation start.
- **Tautological conditions.** Outcome-reading C is rejected at admission;
  cohort-wide boilerplate is exposed by zero exclusion (KTD-2).

## Sources

- `EPIC.md` M6.1 (goal, expected form, exit condition) + M6.2 (the consumer).
- `AGENTS.md` (success-invariant compression in the core loop; verification
  hierarchy; no promotion; deterministic fixtures).
- `docs/toolbox-dsl.md` (`compress` over successes; typed operators).
- `docs/evaluation.md` + migration v15 (verdict/strength contract),
  `docs/frontier-generation.md` + v14 (targets, violation verdicts,
  proposal_hash), `docs/invariant-mining.md` (predicate contract reused).
- Live surfaces verified at commit `a0591a4` (schema v16):
  `internal/store/migrations.go` (evaluations/proposals/targets DDL;
  `editTableCheckInPlace`), `internal/frontier/engine.go`
  (`ProposedSignature` in-memory only — the U1 gap),
  `internal/invariant/{predicate,engine,challenge}.go` (AdmitCandidate,
  Evaluate, epistemic composition), `internal/provider` (Identity() pattern).

## Product Contract

- CLI additions (stable `--json`): `successes compress`,
  `success-invariant list|show`.
- Read surface for M6.2: `success_invariants` filtered `proposed` (until the
  follow-up challenge lifecycle exists), with strength composition exposed so
  policy can weight evidence honestly.
- Epistemic guarantee: cohort selection, discrimination, and identity are
  code-owned; provider authors only candidate conditions; no state beyond
  `proposed`; verification strength is never laundered.

