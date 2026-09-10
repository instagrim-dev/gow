---
artifact_contract: ce-unified-plan/v1
artifact_readiness: implementation-ready
execution: code
product_contract_source: ce-plan-bootstrap
origin: GitHub issue #18 (instagrim-dev/newf) — "Persist and mutate search policy" (to be filed)
epic: EPIC.md milestone M6.2 — Persist and mutate search policy
depends_on: >-
  M6.1 success compression (migration v17, `internal/success`, `success_store`,
  role `'success-compress'`) and the M5.1/M5.2 frontier + evaluation surfaces
  (schema v14/v15) plus the v16/v18 challenge-lifecycle hardening
  (`operator_attested`, challengeable attestation + resume). This plan is
  migration **v19** (live `currentSchemaVersion` verified = 18 at drafting;
  U0 re-verifies before numbering — the twice-learned migration-drift lesson).
  It CONSUMES: `success_invariants` (proposed, predicate-fingerprint identity,
  strength composition + coverage/exclusion counts), `candidate_invariants` /
  `invariant_current_state` (surviving/operator_attested targets), the
  `frontier.Rank` ordinal objective and `frontier.Family` set, `frontier_proposals`
  (redundancy history), and `evaluated_failures` (atlas re-entry). It PRODUCES a
  persisted, revisioned `SearchPolicy` that biases the NEXT `frontier generate`.
created: 2026-09-10
plan_type: feat
---

# feat: Persist and mutate search policy

## Summary

Implement EPIC **M6.2**: make future search behavior **explicit, versioned,
persisted, and inspectable**, so a run can reproduce *why* a frontier proposal
was favored or suppressed instead of relying on "the model remembered the last
failures from context" (AGENTS.md: *Do not rely on "the model will remember
this from context" as state management*).

A `SearchPolicy` is a typed, revisioned artifact — not prose, not a scalar —
built by **code** from the accumulated, code-owned evidence the loop already
persists, and consumed **deterministically** by the next `frontier generate`:

```text
prefer:  mechanism structure associated with code-verified partial success
avoid:   surviving failure invariants (their preserved structure)
expand:  under-sampled / uncovered mechanism families
penalize: mechanisms repeatedly shown redundant or repeatedly failing
```

The decisive design move — mirroring M4.2/M6.1 — is that **the policy is
derived from persisted evidence by code, and the model may only PROPOSE
mutations that code re-verifies before they take effect** (`ModelJudgment !=
Verification`). A directive that cannot be grounded in a persisted artifact
(a real surviving invariant id, a real success-invariant fingerprint, a real
under-covered family) is recorded inert and biases nothing.

The policy then re-enters the loop as a **deterministic, ordinal bias** over
the existing `frontier.Rank` objective — it never invents a fabricated scalar,
never overrides a code-verified violation check, and never suppresses a
proposal so completely that the run becomes unfalsifiable. Every applied bias
is logged against the generation run so the "why" is reconstructable.

This slice closes the recursive edge in the core loop:

```text
... -> SuccessInvariant (M6.1) -> SearchPolicyRevision (THIS slice)
    -> biased FrontierProposal (feeds back into M5.1 generation)
```

It does NOT run holdout scoring (M7), add cross-discipline adapters (M8), or
introduce a policy DSL parser (`docs/toolbox-dsl.md` is the semantic target,
not a language to implement now).

---

## Problem Frame

After M6.1 the system can name, per problem: the surviving failure invariants
(what conserved structure resists every failed family), and the candidate
success invariants (the condition `C` that discriminates progress among
P-breakers, with strength composition). After M5.1 it can generate proposals
ranked by a fixed ordinal objective. What it **cannot** do is let any of that
accumulated evidence change the *next* generation. Every `frontier generate`
today starts from the same fixed `frontier.Rank` ordering and the same
unbiased target/family set, regardless of what the last evaluation round
learned. The loop is not yet recursive in the one place the thesis needs:

```text
failure-space -> invariant -> break -> partial success -> success invariant
   -> SEARCH POLICY (this slice) -> biased frontier -> ...
```

Concretely, the following persisted evidence is currently inert with respect to
future search:

- **Surviving/operator_attested invariants** — the strongest known conserved
  failure structure. Policy should `avoid` re-proposing mechanisms that
  preserve it (a proposal that satisfies a surviving failure invariant is, by
  M4.2 semantics, mechanistically inside known failure structure).
- **Success invariants (M6.1)** — the discriminating condition `C`. Policy
  should `prefer` proposals whose signature satisfies a code-verified `C`,
  weighted honestly by that success invariant's verification-strength
  composition (three model-judged partial successes are not the evidence three
  deterministic ones are — KTD from M6.1 carries forward).
- **Coverage gaps** — the failure-space / cluster coverage report already knows
  which mechanism axes are under-sampled. Policy should `expand` there.
- **Redundancy + repeated failure** — `frontier_proposals` (same directed
  attack re-emitted) and `evaluated_failures` (mechanism re-entered the atlas)
  already record what keeps failing. Policy should `penalize` it.

### Governing constraints (AGENTS.md, docs/toolbox-dsl.md, EPIC M6.2)

- **Policy is persisted, versioned, inspectable data** — never model memory.
  A revision carries full provenance to the evidence rows that justify each
  directive.
- **Code derives; the model may only propose mutations code re-verifies.** A
  provider `mutate` proposal referencing a nonexistent/unresolved artifact is
  recorded inert (`ModelJudgment != Verification`). The default policy builder
  is a deterministic, offline derivation over persisted rows — CI needs no
  model.
- **Bias is ordinal and bounded, never a fabricated scalar.** Policy adjusts
  the existing lexicographic `frontier.Rank` objective via ordinal nudges and
  candidate-set membership; it MUST NOT override a code-verified violation
  check nor fully suppress the falsification surface (a run must remain
  falsifiable — EPIC delivery principle 9 / stopping-condition discipline).
- **No epistemic promotion.** Preferring structure associated with
  `partial_success` does not upgrade that partial success to success, nor a
  success invariant out of `proposed`.
- **Deterministic + reproducible.** Same problem state + same policy revision
  ⇒ same biased ranking; the applied bias is logged per generation run.
- **Additive persistence (v19)** with `RAISE(ABORT)` immutability triggers;
  the one guarded non-additive step is widening
  `provider_invocations.role` to admit `'policy-mutate'` via the established
  in-place `writable_schema` CHECK edit (v11/v13/v14/v15/v17 precedent).

---

## Scope

### In scope

1. **`internal/policy` (pure engine):** the `SearchPolicy` value type
   (prefer/avoid/expand/penalize directives, each a typed reference to a
   persisted artifact + an ordinal weight + a provenance handle), a
   deterministic `Derive(evidence) -> SearchPolicy` that builds directives from
   code-owned inputs, and an `Apply(policy, []frontier.Candidate) ->
   []frontier.Candidate` that re-ranks by a policy-biased objective. No SQL,
   Cobra, or provider concepts.
2. **`PolicyMutator` provider role** (`internal/provider`, role
   `'policy-mutate'`): given a code-built `PolicyEvidence` projection, PROPOSE
   candidate directives `{kind, target-reference, rationale}`. Each proposal is
   re-verified by code (the referenced artifact must resolve) before it can
   enter a revision. Deterministic deriving fixture for CLI/CI.
3. **Persistence (migration v19):** `search_policy_revisions` (per-problem,
   revisioned, run-lifecycled), `search_policy_directives` (kind +
   target-kind/target-id + ordinal weight + epistemic source), and
   `search_policy_provenance` (directive → justifying evidence rows:
   surviving-invariant id, success-invariant fingerprint, cluster/coverage-axis
   id, redundancy/failure counts), all immutable; the role-CHECK widening; and
   a `validateSchemaTables` extension (three tables + role assertion).
4. **Generation integration:** `frontier generate` optionally loads the latest
   policy revision for the problem and applies its bias in-code after
   `EvaluateProposals`/`Rank`, recording the `policy_revision_id` and a
   per-proposal applied-bias log on the generation run (so `frontier show`
   can answer "why this order"). A generation with no policy behaves exactly as
   today (backward compatible; the bias is opt-in and defaulted off unless a
   policy exists).
5. **Pipeline + CLI + run lifecycle:** `newf policy mutate --problem <id>`
   (derive + optionally fold in verified provider proposals; run
   `running → completed/failed`), `newf policy list|show [--json]`. Register in
   `cmd/newf/root.go`.
6. **Tests at every seam** (engine derivation truth table + bias-application
   determinism + falsifiability-floor guard; store idempotency / immutability /
   migrate-from-v18; provider inert-proposal rejection; end-to-end
   mine→challenge→generate→evaluate→compress→**mutate**→generate-again showing
   the ranking changed for a grounded reason) and **docs**
   (`docs/search-policy.md`, persistence v19, EPIC M6.2 status, README loop
   section + `toolbox-dsl` cross-link).

### Deferred to Follow-Up Work

- **Policy-conditioned PROMPTING of the generator** (biasing what the provider
  proposes, not just how code ranks it) — trigger: evidence that post-hoc
  re-ranking under-samples preferred regions; sink: extend `GenerationRequest`
  with a policy projection. This slice biases ranking + candidate-set
  membership only, which is the falsifiable, code-owned core.
- **Challenge lifecycle for policy revisions** — policies stay at their derived
  state; no promotion machinery here.
- **Multi-objective / learned weights** — ordinal nudges only; no scalar
  scoring or tuning loop.

### Out of scope

- Holdout scoring (M7), cross-discipline litmus (M8), a DSL parser, any
  fabricated numeric policy score, any override of a code-verified violation
  check.

---

## Key Technical Decisions

- **KTD-1 — Directives are typed references, never free strings.** Each
  directive names a `target_kind` (`surviving_invariant` | `success_invariant`
  | `mechanism_family` | `redundant_attack` | `repeated_failure`) and a
  `target_id` that MUST resolve to a persisted row at derivation time. An
  unresolved reference is dropped from the active policy and recorded in
  provenance as inert. This is the structural expression of "code derives; the
  model only proposes."
- **KTD-2 — Bias is a bounded ordinal transform of the EXISTING objective.**
  `Apply` inserts policy adjustments as ADDITIONAL, lower-priority sort keys
  beneath the code-verified `ViolatesAnyTarget` gate — it can reorder within a
  violation tier and adjust the `MechanisticDistance`/redundancy nudges, but it
  can NEVER float a non-violating proposal above a violating one, nor drop a
  proposal from the run entirely (suppression is a rank penalty, floored so the
  cheapest-falsification proposal for each surviving target always survives —
  the falsifiability floor).
- **KTD-3 — Strength-weighted preference (carries M6.1 KTD-3).** A `prefer`
  directive sourced from a success invariant inherits that invariant's
  strength composition; a model-judgment-dominant `C` produces a WEAKER nudge
  than a deterministic one. The weight is an ordinal band derived from the
  composition, never a scalar, and never promoted.
- **KTD-4 — Identity + idempotency like M4.2/M6.1.** `search_policy_revisions`
  dedup on `(problem, evidence_cohort_hash, mutator_version, policy_schema)`
  where the cohort hash is a content hash over the sorted justifying evidence
  (surviving-invariant ids + state, success-invariant fingerprints + strength,
  coverage-axis ids, redundancy/failure counts). Re-deriving over unchanged
  evidence returns the existing revision; new evidence yields the next.
- **KTD-5 — Application is logged, not implicit.** When `frontier generate`
  applies a policy, it persists `policy_revision_id` on the generation run and,
  per proposal, the net ordinal bias applied (and which directives fired). This
  is the "reproduce why a proposal was favored or suppressed" exit condition —
  made a queryable artifact, not a log line.
- **KTD-6 — Backward compatible + opt-in.** No policy ⇒ identical behavior to
  today. Policy application is deterministic given (problem state, revision);
  it is defaulted on only when a revision exists, and `--no-policy` forces the
  unbiased path for A/B and holdout-baseline use (M7 B0/B3 contrast).
- **KTD-7 — Role widening follows the shipped pattern.** v19 widens
  `provider_invocations.role` to include `'policy-mutate'` via
  `editTableCheckInPlace`, guarded + idempotent, with a from-v18 migration
  test (the v13/v15/v17 tests are the template).

---

## Data flow

```text
surviving/operator_attested invariants (M4.3)   success_invariants (M6.1, proposed + strength)
        |                                                 |
        v                                                 v
coverage/redundancy/failure evidence (cluster, frontier_proposals, evaluated_failures)
        |
        v
[ code evidence builder ]  ->  PolicyEvidence (typed, resolved references only)
        |
        v
[ PolicyMutator provider ]  ->  candidate directives {kind, target-ref, rationale}
        |
        v
[ code re-verification ]  every target-ref must resolve to a persisted row (else inert)
        |
        v
[ policy.Derive, pure ]  prefer/avoid/expand/penalize + ordinal weights + provenance
        |
        v
SearchPolicyRevision (immutable, revisioned, v19)  --  directives + provenance
        |
        v
[ frontier generate ]  policy.Apply over Rank output (bounded ordinal bias; violation gate untouched)
        |
        v
biased proposals + logged applied-bias (policy_revision_id on the run)   [reproducible "why"]
```

---

## Implementation Units

### U0. Re-verify substrate before numbering
Confirm live `currentSchemaVersion` (expected 18) and the exact signatures of
`ListSuccessRevisions`/`SuccessInvariantRow` (strength composition +
coverage/exclusion), `frontier.Rank`/`frontier.Candidate`, the coverage report
reader, and `editTableCheckInPlace`. Number this migration **v19** (or the
next free integer if more landed). Re-cement U2/U3 against the live
provider/invariant APIs (the recurring concurrent-writer lesson).

### U1. Domain ids + policy value type (`internal/domain`, `internal/policy`)
Add `SearchPolicyRevisionIDPrefix`/`SearchPolicyDirectiveIDPrefix`
(+ `New*`/`Validate*` + errors + id tests). Pure `internal/policy`:
`Directive{Kind, TargetKind, TargetID, Weight domain.Ordinal, Source}`,
`SearchPolicy{Revision, Directives}`, and the `Kind`/`TargetKind` enums with
`Valid()`. No imports of SQL/Cobra/provider.

### U2. Evidence builder + `policy.Derive` (pure engine)
`Derive(evidence PolicyEvidence) SearchPolicy`: map surviving invariants →
`avoid` (weight from evidence strength: operator_attested > surviving), success
invariants → `prefer` (weight = strength-composition band, KTD-3), uncovered
axes → `expand`, redundant/repeatedly-failing mechanisms → `penalize`. Fully
deterministic, order-stable, drops unresolved references. Unit tests: each
directive kind from a seeded evidence fixture; strength-weighted preference
ordering; unresolved reference dropped; empty evidence ⇒ empty (legitimate)
policy.

### U3. `policy.Apply` bias + falsifiability floor (pure engine)
`Apply(policy SearchPolicy, candidates []frontier.Candidate) []frontier.Candidate`:
re-rank beneath the `ViolatesAnyTarget` gate using policy nudges; NEVER
reorders across the violation tier; floors suppression so each surviving
target keeps its cheapest-falsification proposal. Returns the biased order plus
a per-candidate `AppliedBias` log. Unit tests: violation gate is inviolable
(a preferred non-violator cannot outrank a violator, KTD-2); suppression floor
preserves falsification surface; identical input+policy ⇒ identical output
(determinism); no-policy ⇒ identity.

### U4. `PolicyMutator` provider role (`internal/provider`)
`PolicyEvidence` projection (typed, resolved references + strength facts) with
order-independent `Fingerprint()`; `PolicyMutationResponse`
(`[]DirectiveProposal{kind, target_kind, target_id, rationale}` + metadata +
payloads); `Identity()` per the M5.2/M6.1 precedent.
`DerivingFixturePolicyMutator`: re-emit exactly the code-derivable directives
(transparent, deterministic), so the fixture path equals the pure derivation
and CI needs no model. Tests: determinism, fixed order, inert rejection of a
proposal whose `target_id` does not resolve, fingerprint order-independence.

### U5. Persistence (migration v19) + writer/reader
`search_policy_revisions` (problem, revision, mutator_version, policy_schema,
evidence_cohort_hash UNIQUE per KTD-4, run lifecycle), `search_policy_directives`
(revision FK, kind, target_kind, target_id, weight, epistemic_source),
`search_policy_provenance` (directive FK → justifying evidence row refs), all
with immutability triggers. `PersistPolicyRevision` (transactional, idempotent
on KTD-4 tuple), `GetPolicyRevision`, `ListPolicyRevisions`,
`LatestPolicyRevision`. Widen role CHECK (`'policy-mutate'`, KTD-7) + extend
`validateSchemaTables`. Store tests: idempotency; re-derive after new evidence
⇒ next revision; immutability; from-v18 role-migration test; fresh-migrate
schema validation.

### U6. Generation integration + applied-bias log
`GenerateFrontier` loads `LatestPolicyRevision(problem)` (unless `--no-policy`),
maps it to `policy.SearchPolicy`, and calls `policy.Apply` on the ranked
candidates before persistence; records `policy_revision_id` on the generation
run and the per-proposal applied-bias in a new
`frontier_proposal_policy_bias` child (or a column on the proposal row —
decide at U0 against the immutable-row constraint; a sibling child is the
safe default). `frontier show` surfaces it. Backward-compatible: no revision ⇒
unchanged output (KTD-6).

### U7. Pipeline + CLI + run lifecycle
`newf policy mutate --problem <id> [--no-provider]` (run
`running → completed/failed`; empty evidence ⇒ legitimate empty revision,
never fabricated), `newf policy list --problem <id>`,
`newf policy show [revision-id] [--problem <id>]`, all `--json` with directive
kinds, target references, ordinal weights, epistemic source, and provenance
handles. Register in `cmd/newf/root.go`.

### U8. End-to-end + regression fixtures
Full offline chain: seed → cluster → failure-space → mine → challenge
(→ surviving) → frontier generate → evaluate → compress → **policy mutate** →
**frontier generate again**; assert (a) a `prefer` from a code-verified success
invariant reorders a preferred proposal UP within its violation tier;
(b) an `avoid` from a surviving invariant penalizes a matching proposal without
dropping the falsification floor; (c) the violation gate is never crossed;
(d) `policy_revision_id` + applied-bias are logged on the second generation and
reconstruct the "why"; (e) re-mutate over unchanged evidence is idempotent;
(f) new evidence ⇒ new revision; (g) `--no-policy` reproduces the unbiased
order (M7 baseline contrast).

### U9. Docs + gates
`docs/search-policy.md` (the derive/verify/apply contract, ordinal bias,
falsifiability floor, strength weighting, M7 baseline handoff),
`docs/persistence.md` v19 section, EPIC M6.2 status note, README core-loop +
`toolbox-dsl` cross-link. Gates: `go build ./...`, `go vet ./...`,
`go test ./...` green, `gofmt -l .` empty; boundary test at every new seam.

---

## Verification Contract

- **Policy is derived from persisted evidence, not model memory:** every active
  directive resolves to a real row; a fixture proposal with an unresolved
  target is recorded inert and biases nothing (falsify: point a directive at a
  deleted id, watch it drop).
- **Bias respects the code-verified gate (KTD-2):** a preferred non-violating
  proposal can never outrank a violating one; property test over permuted
  candidate sets.
- **Falsifiability floor holds:** for every surviving target, the
  cheapest-falsification proposal survives suppression; a run never becomes
  unfalsifiable through policy.
- **Strength is never laundered (KTD-3):** a `prefer` from a
  single-model-judgment-dominant success invariant produces a strictly weaker
  ordinal nudge than one from a deterministic-dominant invariant; round-trips
  through persistence.
- **Reproducible "why" (KTD-5):** the second generation persists
  `policy_revision_id` and per-proposal applied bias; `frontier show`
  reconstructs the favored/suppressed rationale from stored rows only.
- **Identity/idempotency (KTD-4):** re-mutate over an unchanged evidence cohort
  returns the same revision; a new evaluation/compression changes the hash and
  yields the next; cohort hash is order-independent.
- **Migration:** from-v18 upgrade test (role CHECK gains `'policy-mutate'`;
  child FKs survive the `writable_schema` edit); idempotent on fresh v19;
  `validateSchemaTables` rejects a partially-upgraded store.
- **Backward compatibility (KTD-6):** no policy revision ⇒ byte-identical
  generation output vs. pre-v19; `--no-policy` forces the unbiased path.

## Definition of Done

Given a problem with success invariants and surviving failure invariants,
`newf policy mutate` produces an immutable, revisioned `SearchPolicy` — typed
prefer/avoid/expand/penalize directives, each a resolved reference to a
persisted artifact with an ordinal weight and full provenance — and the next
`newf frontier generate` applies that policy as a bounded, deterministic
ordinal bias over the existing objective (never crossing the code-verified
violation gate, never breaching the falsifiability floor), logging the applied
bias so a reader can reproduce why any proposal was favored or suppressed.
Partial success and surviving failure structure now change the search model
rather than remaining result rows — fully offline, no epistemic promotion.

## Risks

- **Bias that hides the falsification surface.** A strong `avoid`/`penalize`
  could suppress the very proposals that would falsify a shaky invariant.
  Mitigated by the KTD-2 floor (cheapest-falsification proposal per surviving
  target is never suppressed) and a guard test.
- **Over-fitting to sparse evidence.** Early corpora yield thin cohorts; a
  single partial success should not dominate policy. Mitigated by ordinal
  (not scalar) weights and strength weighting; counts persisted, not hidden.
- **Migration-number drift (recurring, twice-learned).** v15–v18 were consumed
  across recent slices; U0 re-verifies `currentSchemaVersion` against live main
  before numbering.
- **Concurrent-writer churn.** Provider `Identity()`, verifier, and frontier
  signatures have shifted mid-slice before; re-cement U2/U3/U4/U6 against the
  live APIs at implementation start.
- **Immutable-row vs. applied-bias log.** Proposal rows are immutable; the
  applied-bias must be a sibling child written in the same generation
  transaction, not an UPDATE (decided at U0/U6).

## Sources

- `EPIC.md` M6.2 (goal, policy shape, exit condition) + M6.1 (the producer of
  the consumed success-invariant surface) + delivery principle 8 (search policy
  is persisted state) and 9 (falsifiability before theater).
- `AGENTS.md` (search-policy mutation section: explicit, versioned, persisted,
  inspectable; no "model remembers from context"; ModelJudgment != Verification;
  deterministic fixtures).
- `docs/toolbox-dsl.md` (`mutate SearchPolicy Evidence -> SearchPolicy`;
  search-policy state YAML; prefer/avoid/expand/penalize; ordinal over scalar).
- Live surfaces verified at drafting (schema v18):
  `internal/success/engine.go` + `internal/store/success_store.go`
  (success-invariant + strength surface), `internal/frontier/{engine,rank}.go`
  (the ordinal objective this slice biases), `internal/pipeline/frontier.go`
  (target/family selection + `Rank` call site — the integration point),
  `internal/store/migrations.go` (`editTableCheckInPlace`, role-CHECK precedent),
  `internal/provider` (`Identity()` pattern).

## Product Contract

- CLI additions (stable `--json`): `policy mutate`, `policy list|show`; new
  `--no-policy` flag on `frontier generate` for unbiased/baseline runs.
- Read surface for M7: a persisted, revisioned `SearchPolicy` per problem, and
  a per-generation applied-bias log — the B3 (invariant-guided) arm of the
  holdout baselines, contrastable against `--no-policy` (B0-adjacent).
- Epistemic guarantee: policy is code-derived from persisted evidence and
  provider proposals are code-re-verified; bias is bounded and ordinal; the
  code-verified violation gate and the falsifiability floor are inviolable; no
  state promotion; verification strength is never laundered into preference.
