---
artifact_contract: ce-unified-plan/v1
artifact_readiness: implementation-ready
execution: code
product_contract_source: ce-plan-bootstrap
origin: GitHub issue #9 (instagrim-dev/newf) — "Implement deterministic mechanism canonicalization and comparison"
created: 2026-09-10
plan_type: feat
---

# feat: Deterministic mechanism canonicalization and comparison

## Summary

Implement the deterministic comparability layer for `newf`: project normalized
approaches into versioned **canonical mechanism signatures**, resolve provider
surface labels to stable versioned canonical IDs through a persisted
**vocabulary**, produce **order-independent fingerprints**, and expose
**component-wise deterministic comparison** that distinguishes mechanistic from
surface diversity. The governing principle is *models discover candidate labels;
software owns identity* — the system must never persist a model's "are these the
same?" answer as truth.

This slice is the input contract for the next mechanism-clustering slice. It runs
entirely offline against project-authored deterministic fixtures.

**Settled scope decision (session-settled: user-directed): plan #9 as if #7
(approach normalization) is complete.** As of implementation time, **#7 has
already landed on this branch** (`currentSchemaVersion = 3`): the
`normalization_revisions` / `approaches` / `approach_revisions` / `mechanisms`
(inline `locality`/`construction_mode`/`uncertainty_mode` posture columns) /
`mechanism_attributes(kind,value,ordinal)` / `outcomes(class)` /
`failure_boundaries(condition,ordinal)` / `source_supports` substrate exists in
migration v3, with domain types in `internal/domain/normalize.go`, the store
writer/reader in `internal/store/normalize.go` (`PersistNormalization`,
`GetMechanismDetail`, `GetApproachDetail`, `ListApproaches`, …), and a registered
`newf mechanism show` command. **#9 therefore consumes this substrate directly
and does not recreate it.** The only stand-in #9 adds is a thin, offline
*fixture-seed loader* that constructs a `store.NormalizationInput` and calls the
existing `PersistNormalization` to populate mechanism rows without a provider
call (U1). New canonicalization tables are added as **migrations v4/v5/v6**.

---

## Problem Frame

`newf` cannot cluster or mine invariants over prose similarity. It needs a stable,
provenance-preserving, deterministic representation of *mechanism structure* so
that two approaches can be compared component-by-component (assumptions,
operators, preserved properties, representations, boundaries, posture) without a
model making the final identity decision.

Today the repository has problems/runs (migration v1), immutable
sources/snapshots (v2), and the **normalized-approach substrate (v3): approaches,
normalization revisions, approach revisions, mechanisms (with inline posture),
mechanism attributes, outcomes, failure boundaries, source supports** — the #7
substrate is *implemented*, not merely designed. This plan implements the
canonicalization/comparison/fingerprint layer **on top of** that existing
substrate, adding only a fixture-seed loader to populate it offline.

### Governing constraints (from `AGENTS.md` and `docs/abstraction-safety.md`)

- `newf` owns identity, canonicalization, fingerprints, comparison semantics.
- Canonical ID is the comparison identity; original wording is preserved as provenance.
- A canonical ID **does not upgrade** the truth status of the underlying claim
  (`explicit`/`inferred`/`ambiguous`/`unknown`/`unsupported` are preserved).
- Ambiguous/unknown mappings stay explicit; never silently coerce to the nearest term.
- Canonicalization is an abstraction boundary: every mapping records what wording
  was compressed, what canonical distinction was retained, what was lost, and what
  the classifier introduced. If two approaches become identical only because the
  vocabulary erased an outcome-predictive distinction, the normalization is defective
  (regression fixture required).
- Deterministic/independently-checkable operations (canonical sorting, set math,
  hashing, comparison) are owned by code, not the model.

---

## Scope Boundaries

### In scope

- Versioned mechanism-signature schema (`mechanism/v1`).
- Small stable semantic-spine vocabulary with namespaced canonical IDs
  (`core.*`, `domain.<field>.*`), versioned, with optional parent/child hierarchy.
- Deterministic canonicalization of already-classified semantic labels → canonical
  IDs, with explicit resolution states.
- Per-field claim support/provenance preservation.
- Deterministic, order-independent signature fingerprints.
- Component-wise deterministic comparison (set overlap/Jaccard, enum equality,
  hierarchy distance, boundary relation comparison) with explicit versioned weights.
- Mechanistic-vs-surface classification.
- CLI: `newf mechanism signature`, `newf mechanism compare`, `newf vocabulary list`,
  `newf vocabulary show`, `newf vocabulary resolve`, all with `--json`.
- Persistence of vocabulary terms/versions, rubrics, signatures, field provenance,
  resolution status, fingerprints, comparison-run metadata; historical signatures
  reproducible against the version that produced them.
- Offline deterministic fixture set (6 required cases) + CI.

### Deferred to Follow-Up Work

- Full #7 provider-driven `normalize` command and provider adapter layer (this plan
  uses a deterministic fixture loader as the mechanism-record source; see Assumptions).
- The bounded semantic-classification *provider call* (`locality/v1` rubric applied
  by a model). This plan persists rubric records and consumes pre-classified fixture
  labels; it does not call a provider to classify. Canonicalization operates on
  labels that are already classified.

### Out of scope (per issue, do not implement)

Clustering; invariant mining; embeddings/vector search; learned metric training;
autonomous ontology expansion; cross-domain transfer; proof verification; scalar
"research quality" scores.

---

## Assumptions

- **A1 — #7 substrate exists and is consumed directly.** The
  `approaches` / `normalization_revisions` / `approach_revisions` / `mechanisms` /
  `mechanism_attributes` / `outcomes` / `failure_boundaries` / `source_supports`
  tables (migration v3) and their domain types (`internal/domain/normalize.go`:
  `Approach`, `ApproachRevision`, `Mechanism`, `MechanismAttribute`, `Outcome`,
  `FailureBoundary`, `SourceSupport`, `NormalizationRevision`) already exist. **U1
  does not create tables or domain types.** U1 adds only a fixture-seed loader that
  builds a `store.NormalizationInput` and calls the existing
  `store.PersistNormalization` to populate mechanism rows offline (the deterministic
  stand-in for #7's provider `normalize`). #9 reads mechanisms via the existing
  `store.GetMechanismDetail` / `GetApproachDetail`.
- **A2 — Classification is upstream.** Mechanism field values arrive already
  semantically classified (the provider/rubric step is #7's / a later slice's job).
  #9 canonicalizes classified labels deterministically. The `classifier_contract`
  version travels with each claim as provenance.
- **A3 — Offline determinism.** No network/model access in any code path exercised
  by tests or the required CLI commands. Vocabulary is seeded from an in-repo
  versioned definition.

---

## Requirements

Traceability back to issue #9 acceptance criteria (AC) and definition of done (DoD).

- **R1** — Normalized approaches project into versioned canonical mechanism signatures (`mechanism/v1`). (AC1)
- **R2** — Canonical IDs are stable, versioned, and distinct from source wording. (AC2)
- **R3** — Every canonicalized field retains explicit epistemic/source provenance (`status`, `support{snapshot_id, locator}`, `confidence`, `classifier_contract`). A canonical ID never upgrades claim status. (AC3)
- **R4** — Ambiguous/unknown/novel/rejected mappings remain explicit; no silent coercion to nearest term. (AC4)
- **R5** — Signatures have deterministic, order-independent fingerprints; provenance metadata does not change semantic identity; schema/vocab version changes are visible in the fingerprint. (AC5)
- **R6** — Comparison is component-wise and deterministic (per field: operators, assumptions, preserves, breaks, representations, boundaries, posture, auxiliary_objects). No forced single scalar. (AC6)
- **R7** — Surface text similarity is retained only as auxiliary diagnostic data; it never defines mechanistic identity. (AC7)
- **R8** — Vocabulary evolution does not rewrite historical signatures; a signature is reproducible against the exact vocab/schema version that produced it. (AC8)
- **R9** — Abstraction-loss regression is covered by a fixture/test (over-compression that erases a success/failure distinction is detectable). (AC9)
- **R10** — CLI and `--json` expose signatures and comparison results for: `mechanism signature`, `mechanism compare`, `vocabulary list|show|resolve`. (AC10)
- **R11** — CI runs entirely offline with deterministic fixtures. (AC11)
- **R12** — Comparator emits the mechanistic-vs-surface classification (`surface-distinct+mechanism-near`, `surface-near+mechanism-distinct`, `mechanism-near`, `mechanism-distinct`, `unknown`). (DoD)
- **R13** — Comparison weights, if present, are explicit, versioned, and configurable; no behavior hidden in opaque distance. (issue "Comparison algorithms")
- **R14** — Vocabulary resolution states are the fixed set: `resolved`, `ambiguous`, `novel_candidate`, `unknown`, `rejected`. (issue "Canonicalization semantics")
- **R15** — Domain packages carry no SQL/Cobra/vendor coupling (repo `AGENTS.md` Go bias).

---

## Key Technical Decisions

- **KTD1 — Signature is a derived, revisioned, immutable record keyed by `(mechanism_id, schema_version, vocabulary_version)`.** A signature is computed from a mechanism revision under a specific schema+vocab version and never rewritten. Re-canonicalizing under a newer vocabulary creates a *new* signature row; the old one stays reproducible (satisfies R8). Rationale: mirrors the existing immutable-derived-record pattern (`source_snapshots`, and the `*_revision` design in `docs/persistence.md`), and makes "vocabulary evolution does not rewrite history" true by construction rather than by discipline.

- **KTD2 — Canonicalization consumes already-classified labels; it does not classify.** The input to canonicalization is a `(field, classified_label, support, status, classifier_contract)` tuple. Resolution is a deterministic lookup/alias-match against the vocabulary version, producing one of the five resolution states. Rationale: keeps the truth-sensitive fuzzy step (semantic classification) out of #9 per `AGENTS.md`; #9 owns only the deterministic identity resolution. Alias matching is exact-normalized-string + explicit alias table only — **no fuzzy/embedding match** (chosen over nearest-term matching: nearest-term silently coerces and violates R4).

- **KTD3 — Fingerprint = SHA-256 over a canonical JSON serialization of `{schema_version, vocabulary_version, sorted canonical-ID sets per field, posture enums, outcome.class, sorted boundary relations}` with provenance excluded.** Sets are sorted by canonical ID; posture is fixed-key ordered; JSON uses sorted keys and no insignificant whitespace. Rationale: reuses `crypto/sha256` already in the tree (`internal/pipeline/source.go`); order-independence and version-visibility fall out of the canonical serialization (satisfies R5). Provenance is excluded from the hashed body by default; a separate `provenance_fingerprint` may be added later if needed (not in v0). **Fingerprint equality asserts *resolved-identity equality only* (adversarial P1):** the body is built from `resolved` claims plus posture/outcome/boundaries, so two signatures sharing those but differing in `ambiguous`/`novel_candidate`/`unknown` claims hash equal. This is a deliberate, documented boundary — it is *not* a claim of full mechanistic sameness. To prevent a downstream consumer over-trusting equality, **comparison (U6) must surface any non-`resolved` claim on either side as `incomparable` before any identity conclusion**, and U7 adds a fixture asserting this (equal-resolved / differing-unresolved pair is flagged `incomparable`, not merged). **Identity is also conditional on a fixed `classifier_contract`:** because which claims reach `resolved` depends on the upstream classified label, fingerprints produced under different `classifier_contract` versions are not identity-comparable; the plan states this explicitly and comparison records the contract so cross-contract comparison is not silently treated as sameness.

- **KTD4 — Per-field comparison methods are explicit and typed, keyed by field kind.** Unordered canonical-ID sets (operators, assumptions, preserves, breaks, representations, auxiliary_objects) → set overlap + Jaccard (both reported). Posture (locality/construction/uncertainty) → enum equality. Boundaries → typed relation comparison over canonical IDs. Where the vocabulary declares parent/child, an explicit hierarchy distance is available. Rationale: matches the issue's "possible v0 choices" and the abstraction-safety requirement that behavior be inspectable, not buried in a distance function (R6, R13).

- **KTD5 — No single scalar in v0.** Comparison output is a per-field structure plus an ordinal similarity per field (`identical`/`high`/`low`/`none`/`incomparable`) and a categorical mechanistic-vs-surface classification. Surface similarity is computed only if fixtures/inputs carry surface text, and is tagged `auxiliary_only`. Rationale: issue explicitly forbids fake precision and forbids surface similarity defining identity (R6, R7, R12). **The mechanistic-vs-surface predicate is explicit, versioned (`classify/v1`), and documented**, not a loose heuristic: `mechanism-near` iff every *decisive* field (`preserves`, `operators`, `assumptions`) is `identical`-or-`high` **and** no decisive field is `low`/`none`/`incomparable`; `mechanism-distinct` iff any decisive field is `low`/`none`; otherwise (decisive fields all `incomparable`, or mixed with non-decisive disagreement only) `unknown`. `representations`/`boundaries`/`posture` are non-decisive in v1 and affect only the surface axis. The choice of decisive fields is an explicit invariant claim recorded in the doc (U8) with rationale, and U7 adds a fixture where a decisive-vs-non-decisive disagreement flips the classification, so the weighting is falsifiable rather than assumed.

- **KTD6 — Vocabulary is seeded from an in-repo versioned Go definition, persisted on migrate/first-use, and immutable per version.** New terms/versions append; existing `(version, canonical_id)` rows are immutable (triggers, mirroring the source-immutability pattern). Rationale: deterministic, offline, and makes "historical signatures reproducible" enforceable (R8, R11).

- **KTD7 — Package layout follows `docs/cli-design.md`.** New domain types in `internal/domain/`; new tables/repos in `internal/store/`; a new `internal/canon/` package owns pure canonicalization/fingerprint/comparison logic (no SQL, no Cobra); orchestration in `internal/pipeline/`; Cobra wiring in `cmd/newf/`. Rationale: preserves existing boundaries and the repo's "domain types free of Cobra/SQL" rule (R15).

- **KTD8 — Signature schema is `mechanism/v1` with fields: `representations, operators, assumptions, preserves, breaks, auxiliary_objects` (canonical-ID sets), `posture{locality, construction, uncertainty}` (enums), `outcome{class}`, `boundaries[]` (canonical ID + relation).** Matches the issue's example shape. `breaks` and `auxiliary_objects` may be empty but are always present keys. Rationale: stable key presence is required for deterministic fingerprints and comparisons across sparse signatures.

---

## High-Level Technical Design

### Data flow

```text
(existing #7 substrate, assumed)
  mechanism revision
    -> classified field labels + per-claim support/status/classifier_contract
        |
        v
  [internal/canon] resolve(label, field, vocab_version)
    -> canonical_id + resolution_state (resolved|ambiguous|novel_candidate|unknown|rejected)
    -> claim{canonical_id, status, support, confidence, classifier_contract}   (provenance preserved)
        |
        v
  [internal/canon] buildSignature(schema=mechanism/v1, vocab_version)
    -> MechanismSignature{fields..., posture, outcome, boundaries}
        |
        +--> canonicalSerialize -> sha256 -> fingerprint         (order-independent, version-visible)
        +--> persist signature + field-claim provenance rows      (immutable, revisioned)
        |
        v
  [internal/canon] compare(sigA, sigB)
    -> per-field {overlap, jaccard, ordinal} + posture enum eq + boundary relation cmp
    -> mechanistic-vs-surface classification (+ optional auxiliary surface diagnostic)
```

### Persistence shape (new tables — additive, on top of the existing v3 substrate)

**Vocabulary reconciliation (scope-guardian).** The existing v3 substrate stores
posture as inline enum columns on `mechanisms` (`locality`, `construction_mode`,
`uncertainty_mode`) — there is no separate `mechanism_axis_*` table in the live
schema, so #9's posture canonicalization reads those columns directly (no new
axis vocabulary needed for posture; posture "canonical IDs" are the existing
validated enum values). The new `canonical_term` vocabulary is therefore scoped
**only to the set fields** carried in `mechanism_attributes` (`representation`,
`assumption`, `operator`, `preserves`, `breaks`, `auxiliary_object`) plus
`boundary`. `canonical_term.field_kind` reuses those exact `mechanism_attributes.kind`
values so canonicalization maps 1:1 onto attribute rows and does not introduce a
parallel posture ontology.

```text
canonical_vocabulary(version PK, created_at, notes)
canonical_term(vocabulary_version, canonical_id, field_kind, description,
               parent_canonical_id NULL, PRIMARY KEY(vocabulary_version, canonical_id))
canonical_term_alias(vocabulary_version, canonical_id, alias_normalized,
               PRIMARY KEY(vocabulary_version, alias_normalized))
classification_rubric(contract_version PK, field_kind, definition_json, created_at)

mechanism_signature(id PK, mechanism_id, schema_version, vocabulary_version,
               fingerprint, created_at, run_id,
               UNIQUE(mechanism_id, schema_version, vocabulary_version))
signature_field_claim(signature_id, field_kind, canonical_id, resolution_state,
               claim_status, support_snapshot_id NULL, support_locator NULL,
               confidence NULL, classifier_contract NULL, surface_label,
               PRIMARY KEY(signature_id, field_kind, canonical_id, surface_label))
signature_posture(signature_id, axis, value, PRIMARY KEY(signature_id, axis))
signature_boundary(signature_id, canonical_id, relation, PRIMARY KEY(signature_id, canonical_id, relation))
signature_outcome(signature_id PK, class)

comparison_run(id PK, signature_a_id, signature_b_id, weights_version,
               classification, created_at, run_id)
comparison_field_result(comparison_run_id, field_kind, overlap_json,
               jaccard REAL NULL, ordinal, PRIMARY KEY(comparison_run_id, field_kind))
```

Immutability triggers on `canonical_vocabulary`, `canonical_term`,
`canonical_term_alias`, `mechanism_signature`, and its child rows (append-only,
matching the existing source/snapshot trigger pattern).

### Resolution state machine (deterministic, no fuzzy fallback)

```text
label --normalize--> key
  key exact-matches a canonical_id in vocab_version      -> resolved
  key exact-matches an alias in vocab_version            -> resolved (to aliased canonical_id)
  key matches >1 canonical_id/alias                      -> ambiguous   (record all candidates)
  key matches 0, but caller flagged it novel             -> novel_candidate
  key matches 0, not flagged novel                       -> unknown
  key present on an explicit rejected-terms list         -> rejected
```

Only `resolved` claims contribute a canonical ID to the fingerprint body.
`ambiguous`/`novel_candidate`/`unknown`/`rejected` are persisted as field claims
with their state and surface label, and are surfaced in comparison as
`incomparable` for that element rather than silently dropped or coerced.

---

## Output Structure

```text
internal/
  canon/                         # NEW — pure logic, no SQL/Cobra
    vocabulary.go                # versioned vocab definition + seed + resolve()
    vocabulary_seed.go           # mechanism/v1 canonical terms + aliases (in-repo data)
    signature.go                 # MechanismSignature type + buildSignature()
    fingerprint.go               # canonical serialization + sha256
    compare.go                   # per-field comparison + classification
    *_test.go
  domain/
    canonical.go                 # NEW — CanonicalID, ResolutionState, ClaimStatus, FieldKind enums + validation
                                 # (Approach/Mechanism/etc. already exist in internal/domain/normalize.go — reused, not recreated)
  store/
    migrations.go                # EXTEND — v4 (vocab/rubric), v5 (signatures), v6 (comparison); v3 mechanism substrate already exists
    canon_store.go               # NEW — vocab/signature/comparison repositories
    fixture_seed.go              # NEW (A1) — offline loader: build store.NormalizationInput + call existing PersistNormalization
  pipeline/
    mechanism.go                 # NEW — SignatureMechanism / CompareMechanisms services
    vocabulary.go                # NEW — Vocabulary list/show/resolve services
    output.go                    # EXTEND — response view structs
cmd/newf/
    mechanism.go                 # EXTEND newMechanismCommand (in approach.go today) — add `signature` + `compare` subcommands to existing `newf mechanism show`
    vocabulary.go                # NEW — `newf vocabulary list|show|resolve`
    root.go                      # EXTEND — register vocabulary command + error classes
testdata/
  fixtures/mechanism/            # NEW — 6 required deterministic fixtures
docs/
  mechanism-canonicalization.md  # NEW — contract/semantics doc
```

---

## Implementation Units

### U1. Offline fixture-seed loader over the existing #7 substrate

- **Goal:** Populate the *existing* v3 mechanism substrate offline (no provider call) so #9 has mechanism records to canonicalize and compare.
- **Requirements:** R1 (input side), R11, R15; enables A1/A2/A3.
- **Dependencies:** none (consumes existing v3 tables/types/store).
- **Files:** `internal/store/fixture_seed.go`, `internal/store/fixture_seed_test.go`, `testdata/fixtures/mechanism/` (initial fixtures used by loader tests). **No new migration, no new domain types.**
- **Approach:**
  1. Add a fixture-seed loader (`LoadMechanismFixture` on `*store.Store`, or a small `internal/pipeline` helper) that reads a mechanism fixture JSON and builds a `store.NormalizationInput` — `NormalizationRevision` + `ProviderInvocation` (role `normalize`, marked fixture-sourced) + `[]ApproachInput{LogicalIdentity, Revision, Mechanism, Attributes, Outcome, Boundaries, Support}` — then calls the **existing** `store.PersistNormalization`. This is the deterministic offline stand-in for #7's provider `normalize`.
  2. The fixture JSON schema mirrors the existing domain types (`domain.Mechanism` posture enums, `domain.MechanismAttribute{Kind,Value,Ordinal}`, `domain.Outcome{Class}`, `domain.FailureBoundary{Condition,Ordinal}`, `domain.SourceSupport{...}`). Loader relies on the existing `Validate()` methods; it does not re-validate independently.
  3. Reuse existing IDs (`domain.NewMechanismID`, `NewApproachID`, `NewNormalizationRevisionID`, etc.) — all already defined in `internal/domain/id.go`. **Do not add ID prefixes.**
- **Patterns to follow:** `internal/store/normalize.go` (`PersistNormalization`, `NormalizationInput`/`ApproachInput` shapes, `persistApproachTx`); existing fixture patterns under `testdata/`.
- **Test scenarios:**
  - Loading a single-mechanism fixture creates one normalization revision, one approach + revision, one mechanism with all attribute kinds populated (queryable via `GetMechanismDetail`).
  - Loading a multi-approach fixture creates multiple approaches under one revision (one source → multiple approaches).
  - Attribute rows preserve original surface labels/values (no canonicalization here).
  - Re-loading the same logical identity reuses the approach (via `getOrCreateApproachTx`) rather than duplicating.
  - Invalid fixture (failing a `domain.*.Validate()`) aborts the transaction leaving no partial rows.
- **Verification:** A fixture file loads into rows readable by the existing `GetMechanismDetail`/`GetApproachDetail`; `go test ./internal/store/...` passes offline.

### U2. Canonical domain types + versioned vocabulary with deterministic resolution

- **Goal:** Define the semantic-spine vocabulary (`mechanism/v1` terms), canonical-ID/resolution-state/claim-status types, and the pure deterministic `resolve()`.
- **Requirements:** R2, R4, R14, R15, KTD2, KTD6.
- **Dependencies:** U1 (for `field_kind` alignment only; can proceed in parallel on pure types).
- **Files:** `internal/domain/canonical.go`, `internal/canon/vocabulary.go`, `internal/canon/vocabulary_seed.go`, `internal/canon/vocabulary_test.go`.
- **Approach:**
  1. `internal/domain/canonical.go`: `CanonicalID string` with validation of the namespaced shape (`core.<kind>.<name>` / `domain.<field>.<kind>.<name>`); `ResolutionState` enum (`resolved|ambiguous|novel_candidate|unknown|rejected`) with an exhaustive-check helper; `ClaimStatus` enum (`explicit|inferred|ambiguous|unknown|unsupported`); `FieldKind` enum for the spine (`representation|operator|assumption|preserves|breaks|auxiliary_object|outcome|boundary|posture`) — the set-field kinds match the existing `mechanism_attributes.kind` CHECK values exactly (`representation|assumption|operator|preserves|breaks|auxiliary_object`). **Posture enums are NOT redefined here** — reuse the existing `domain.Locality`/`domain.ConstructionMode`/`domain.UncertaintyMode` from `internal/domain/normalize.go` (already validated: local|global|mixed|unknown, constructive|existential|mixed|unknown, deterministic|probabilistic|unknown), and reuse `domain.OutcomeClass` for outcome.
  2. `vocabulary_seed.go`: the in-repo `mechanism/v1` vocabulary — a small set of `core.*` terms plus a couple of `domain.number_theory.*` terms (e.g. `core.operator.modular_decomposition`, `core.representation.congruence_classes`, `domain.number_theory.property.residue_locality`), each with normalized aliases (e.g. `"works residue-by-residue"`, `"local congruence argument"` → `domain.number_theory.property.residue_locality`), optional parent links, and a `rejected` list.
  3. `vocabulary.go`: `Vocabulary` value object built from the seed; `Normalize(label)` (lowercase, collapse whitespace/punctuation deterministically); `Resolve(field, label, novelFlag) -> Resolution{state, canonicalID, candidates}` implementing the KTD2/state-machine rules with **no fuzzy fallback**.
- **Patterns to follow:** exhaustive switch with `never`-style default per workspace TS rule analog in Go (return error on unknown enum); table-driven tests like existing `_test.go` files.
- **Test scenarios:**
  - Canonical ID validation accepts well-formed namespaced IDs, rejects malformed.
  - Exact canonical-ID match → `resolved`.
  - Alias match (all seeded aliases) → `resolved` to the correct canonical ID.
  - Two differently-worded aliases for the same property resolve to the same canonical ID (supports the "surface-distinct/mechanism-near" fixture).
  - Label matching multiple terms/aliases → `ambiguous` with all candidates recorded.
  - Unmatched label, not novel-flagged → `unknown`; novel-flagged → `novel_candidate`.
  - Rejected-list label → `rejected`.
  - `Normalize` is idempotent and case/whitespace-insensitive in a deterministic way.
  - Resolution never coerces to nearest term (explicitly assert an unknown does not map to a similar canonical ID).
- **Verification:** `go test ./internal/canon/...` passes; resolution is a pure function of `(vocab_version, field, label, novelFlag)`.

### U3. Vocabulary persistence + `newf vocabulary list|show|resolve`

- **Goal:** Persist vocabulary versions/terms/aliases/rubrics immutably; expose read + resolve CLI with `--json`.
- **Requirements:** R4, R8 (version isolation), R10, R14, KTD6.
- **Dependencies:** U2.
- **Files:** `internal/store/canon_store.go` (vocab portion), `internal/store/migrations.go` (**migration v4**, vocab + rubric tables), `internal/pipeline/vocabulary.go`, `internal/pipeline/output.go` (vocab views), `cmd/newf/vocabulary.go`, `cmd/newf/root.go` (register + error classes), `internal/store/canon_store_test.go`, `internal/pipeline/vocabulary_test.go`, `cmd/newf/root_test.go` (CLI JSON).
- **Approach:**
  1. Migration **v4** (`currentSchemaVersion` 3 → 4): `canonical_vocabulary`, `canonical_term`, `canonical_term_alias`, `classification_rubric` with immutability triggers; add to `validateSchemaTables`; bump `currentSchemaVersion`.
  2. On migrate/first-use, seed the `mechanism/v1` vocabulary + `locality/v1` rubric record from U2's in-repo definition (idempotent insert; existing rows immutable).
  3. Repository methods: `ListVocabularies`, `GetVocabulary(version)`, `ListTerms(version, fieldKind?)`, `GetTerm(version, canonicalID)`.
  4. Pipeline services + `--json` response structs following the `SourceListResponse`/`writeJSON` pattern; `resolve` service calls U2's pure `Resolve` against a persisted version and returns the resolution + candidates + the (unchanged) claim status semantics.
  5. CLI: `vocabulary list` (versions/terms), `vocabulary show <canonical-id>` (term + aliases + parent + which versions contain it), `vocabulary resolve <candidate-label>` (`--field`, `--vocab-version`, `--novel`), all honoring global `--json`.
- **Patterns to follow:** `cmd/newf/source.go` command wiring + `writeJSON`; `internal/pipeline/source.go` service + view mapping; `internal/store/store.go` transaction + immutability triggers.
- **Test scenarios:**
  - Seeded vocabulary present after migrate; re-migrate does not duplicate or mutate (`existing`/idempotent).
  - Attempted UPDATE/DELETE on a term/alias row aborts (immutability trigger).
  - `resolve` returns each of the five states for representative labels; `--json` shape is stable and includes `state`, `canonical_id`, `candidates`.
  - `resolve` of an ambiguous label lists all candidates and does not pick one.
  - `vocabulary show` of a term reports its aliases and parent (hierarchy).
  - Unknown canonical ID → `not_found` error class in `--json`.
  - Two vocabulary versions coexist; `list`/`show` scope to the requested version.
- **Verification:** All three vocabulary commands run offline in a temp workspace with stable `--json`; immutability enforced by DB.

### U4. Signature builder + deterministic order-independent fingerprint

- **Goal:** Project a mechanism into a `mechanism/v1` `MechanismSignature` and compute its fingerprint.
- **Requirements:** R1, R3, R5, R8, KTD1, KTD3, KTD8.
- **Dependencies:** U2 (resolve), U1 (mechanism input).
- **Files:** `internal/canon/signature.go`, `internal/canon/fingerprint.go`, `internal/canon/signature_test.go`, `internal/canon/fingerprint_test.go`.
- **Approach:**
  1. `MechanismSignature` type: `SchemaVersion`, `VocabularyVersion`, field sets of `Claim{CanonicalID, ResolutionState, Status, Support, Confidence, ClassifierContract, SurfaceLabel}` for each spine field (always-present keys, may be empty), `Posture{Locality, Construction, Uncertainty}`, `Outcome{Class}`, `Boundaries[]{CanonicalID, Relation}`.
  2. `BuildSignature(mechanism, vocab, schemaVersion)`: for each field value, call `Resolve`; preserve support/status/classifier_contract from the input claim unchanged (assert a canonical ID never alters `Status`).
  3. `Fingerprint(sig)`: build the canonical body = `{schema_version, vocabulary_version, per-field sorted []canonical_id of only resolved claims, posture (fixed key order), outcome.class, boundaries sorted by (canonical_id, relation)}`; serialize with sorted keys + compact form; `sha256.Sum256`; hex-encode. Provenance (support/confidence/surface labels/resolution states other than resolved) is excluded from the body.
- **Patterns to follow:** `crypto/sha256`/`encoding/hex` usage in `internal/pipeline/source.go`.
- **Test scenarios:**
  - Two mechanisms whose fields differ only in ordering produce identical signatures and identical fingerprints (R5 order-independence).
  - Two differently-worded mechanisms that canonicalize to the same IDs produce equal fingerprints.
  - Changing `vocabulary_version` or `schema_version` changes the fingerprint even with identical fields (version visibility).
  - Changing only provenance (support locator, confidence, surface wording) does **not** change the fingerprint.
  - A field with an `ambiguous`/`unknown` claim is excluded from the hashed body but retained on the signature struct.
  - Canonical ID does not upgrade claim status: an `inferred` claim stays `inferred` on the built signature.
  - Empty `breaks`/`auxiliary_objects` still yield a stable fingerprint (present keys).
- **Verification:** `Fingerprint` is a deterministic pure function; golden-fingerprint test is stable across runs and Go map iteration order.

### U5. Signature persistence + `newf mechanism signature`

- **Goal:** Persist immutable, version-keyed signatures with field-level provenance; expose `mechanism signature <mechanism-id>` with `--json`.
- **Requirements:** R1, R3, R5, R8, R10, R15, KTD1, KTD7.
- **Dependencies:** U3 (vocab persistence), U4 (builder/fingerprint), U1 (mechanism rows).
- **Files:** `internal/store/canon_store.go` (signature portion), `internal/store/migrations.go` (**migration v5**, signature tables), `internal/pipeline/mechanism.go` (SignatureMechanism service), `internal/pipeline/output.go` (signature views), `cmd/newf/mechanism.go`, `cmd/newf/root.go` (register), `internal/store/canon_store_test.go`, `internal/pipeline/mechanism_test.go`, `cmd/newf/root_test.go`.
- **Approach:**
  1. Migration **v5** (4 → 5): `mechanism_signature` (+ `UNIQUE(mechanism_id, schema_version, vocabulary_version)`), `signature_field_claim`, `signature_posture`, `signature_boundary`, `signature_outcome`, with append-only immutability triggers; register in `validateSchemaTables`; bump version.
  2. `SignatureMechanism` service: load mechanism (U1), resolve vocab version (default = latest seeded), `BuildSignature`, persist all rows transactionally under a new `run`, return the view. If a signature already exists for `(mechanism, schema, vocab)`, return it as `existing` (idempotent, matching the snapshot-admission `existing`/`created` status pattern) rather than recomputing/rewriting.
  3. CLI: add a `signature <mechanism-id>` subcommand to the **existing** `newMechanismCommand` (currently in `cmd/newf/approach.go`, alongside `show`), with `--vocab-version`, `--schema-version` (default `mechanism/v1`), `--json`. Human output shows fingerprint, per-field canonical IDs with resolution state + claim status, posture, outcome, boundaries. `--json` returns the full signature incl. provenance.
- **Patterns to follow:** `CreateSourceSnapshot` (transaction + `existing`/`created` status), `cmd/newf/source.go` wiring, `writeJSON`.
- **Test scenarios:**
  - First `signature` call creates the row + child rows and returns `created`; second identical call returns `existing` with the same fingerprint and does not write a duplicate.
  - Signature under a different `--vocab-version` creates a distinct row; the original is untouched (R8 reproducibility).
  - Attempted UPDATE/DELETE of a persisted signature row aborts.
  - `--json` output includes fingerprint, every field claim with `resolution_state`, `claim_status`, `support`, and is stable/composable (IDs usable by `compare`).
  - Ambiguous/unknown field claims persist with their state and surface label (R4 preserved through storage).
  - Failed transaction (e.g., unknown mechanism ID) leaves no partial signature.
- **Verification:** `newf mechanism signature <id> --json` in a temp workspace returns a stable signature; re-run is idempotent; DB enforces immutability.

### U6. Component-wise comparison + mechanistic-vs-surface classification + `newf mechanism compare`

- **Goal:** Compare two signatures per-field deterministically, classify mechanistic-vs-surface, optionally persist the comparison run; expose `mechanism compare <a> <b>` with `--json`.
- **Requirements:** R6, R7, R12, R13, KTD4, KTD5.
- **Dependencies:** U4 (types), U5 (persisted signatures to compare).
- **Files:** `internal/canon/compare.go`, `internal/canon/compare_test.go`, `internal/store/canon_store.go` (comparison_run persistence), `internal/store/migrations.go` (**migration v6**, comparison tables), `internal/pipeline/mechanism.go` (CompareMechanisms service), `internal/pipeline/output.go` (comparison views), `cmd/newf/mechanism.go` (add `compare` subcommand to the existing `newMechanismCommand`), `internal/pipeline/mechanism_test.go`, `cmd/newf/root_test.go`.
- **Approach:**
  1. `Compare(sigA, sigB, weights)`: for each set field compute intersection/union over **resolved** canonical IDs → `{overlap_count, jaccard}` and an ordinal (`identical|high|low|none`); elements with non-`resolved` state on either side are reported `incomparable` for that field rather than counted as agreement. Posture axes → enum equality. Boundaries → typed relation comparison. Weights come from an explicit, versioned `weights_version` table/struct (default `weights/v1`), configurable via flag; never collapse to one scalar.
  2. Classification per the versioned `classify/v1` predicate (KTD5): derive `mechanism-near`/`mechanism-distinct`/`unknown` from the decisive fields (`preserves`, `operators`, `assumptions`) only; `representations`/`boundaries`/`posture` are non-decisive and feed only the surface axis. Compute optional surface similarity **only** from any surface text carried on inputs, tag it `auxiliary_only`, and combine into `surface-distinct+mechanism-near` / `surface-near+mechanism-distinct` / `mechanism-near` / `mechanism-distinct` / `unknown`. Surface similarity never feeds the mechanistic decision (R7). Any non-`resolved` claim on either side makes its field `incomparable` and cannot count toward `mechanism-near`.
  3. Persist `comparison_run` + `comparison_field_result` when not `--no-write`; return the view.
  4. CLI `mechanism compare <a> <b>` (`--weights-version`, `--no-write`, `--json`) producing the issue's example human table + full `--json`.
- **Patterns to follow:** `cmd/newf/source.go` two-arg command; `writeJSON`; deterministic table-driven tests.
- **Test scenarios:**
  - Identical signatures → every field `identical`, classification `mechanism-near`.
  - Same canonical IDs reached from different surface wording (the surface-distinct/mechanism-near fixture) → `preserves: identical`, classification `surface-distinct+mechanism-near`.
  - Textually similar but mechanistically different fixture → low canonical-field overlap, classification `surface-near+mechanism-distinct` (proves surface text does not drive identity, R7).
  - Jaccard math correct on known sets; enum posture equality correct.
  - Fields with ambiguous/unknown claims are reported `incomparable`, not silently equal.
  - **Resolved-identity boundary (adversarial P1):** two signatures with identical `resolved` sets/posture/outcome/boundaries but differing `ambiguous`/`unknown` claims — equal fingerprints, yet `compare` flags the differing field `incomparable` and does **not** emit `mechanism-near` on the strength of the equal fingerprint alone.
  - **Decisive-vs-non-decisive flip:** a pair agreeing on `representations` but disagreeing on `preserves` classifies `mechanism-distinct`; the reverse (agree on decisive, differ only on `representations`) classifies `mechanism-near` (or `surface-distinct+mechanism-near`) — proving the `classify/v1` weighting.
  - Weights are read from the versioned config and reflected in output; changing `--weights-version` is visible and does not hide behavior.
  - `--no-write` performs comparison without persisting; default persists a `comparison_run` retrievable by ID.
  - No single scalar is emitted (assert output shape is per-field + categorical).
- **Verification:** `newf mechanism compare A B --json` returns per-field results + classification deterministically; `--no-write` respected; comparison reproducible.

### U7. Deterministic fixture set + abstraction-loss regression + end-to-end integration test

- **Goal:** Author the six required project-authored fixtures and the offline end-to-end integration test that exercises load → signature → signature → compare.
- **Requirements:** R7, R9, R11, and integration verification from the issue.
- **Dependencies:** U1, U5, U6.
- **Files:** `testdata/fixtures/mechanism/*.json` (the six cases), `internal/canon/abstraction_loss_test.go`, `cmd/newf/mechanism_integration_test.go` (or `internal/pipeline/mechanism_integration_test.go`), possibly a small second `mechanism/v0-lossy` vocabulary in `vocabulary_seed.go` for the over-compression case.
- **Approach:** Author fixtures covering exactly the issue's six cases:
  1. two differently-worded approaches that canonicalize to the **same** signature;
  2. two textually similar approaches that differ **mechanistically**;
  3. an **ambiguous** classification that must remain unresolved;
  4. a **novel candidate** term absent from the current vocabulary;
  5. an **over-compression** case: under a deliberately lossy vocabulary version, two mechanisms that differ in `outcome.class` have their *outcome-predictive operators* merged to one canonical ID. Detection is a **system-owned invariant, not a hand-checked equality**: implement `AssertDiscriminationPreserved(mechanisms, vocab)` in `internal/canon` that, for every pair whose `outcome.class` differs, recomputes an **outcome-excluded** canonical body (schema+vocab+resolved field sets+posture+boundaries, *without* `outcome.class`) and flags a discrimination-loss defect if those bodies are equal — i.e. the vocabulary erased every non-outcome distinction between two mechanisms known to differ in outcome. The test asserts this invariant *fires* under the lossy vocab and *does not fire* under `mechanism/v1`. This keeps `outcome.class` in the real identity fingerprint (KTD3/KTD8) while still catching a vocabulary that collapses the fields that *should* have predicted the outcome difference;
  6. **ordering differences** that must yield the **same** fingerprint.
  Then an integration test that, in a temp workspace: migrates, loads fixture A and fixture B, builds both signatures, compares them, and asserts the expected classification per fixture pair.
- **Test scenarios / assertions:**
  - Case 1: fixtures A/B → equal fingerprints, `compare` → `surface-distinct+mechanism-near` (or `mechanism-near`).
  - Case 2: equal/near surface text → `mechanism-distinct` (`surface-near+mechanism-distinct`); asserts R7.
  - Case 3: the ambiguous field stays `ambiguous` through signature + comparison (`incomparable`), never guessed.
  - Case 4: the novel label persists as `novel_candidate`; it is not coerced to an existing canonical ID.
  - Case 5 (R9 regression): `AssertDiscriminationPreserved` **fires** under the lossy vocab (outcome-excluded bodies of the two different-outcome mechanisms are equal) and **does not fire** under `mechanism/v1`. The defect cannot pass silently because the invariant is code-owned, not author-checked.
  - Case 6: reordered field values → identical fingerprint.
  - Integration: signatures stable across runs; fingerprints deterministic; ambiguous fields stay ambiguous; no provenance lost end-to-end (support/status survive load→signature→json).
- **Execution note:** Author the abstraction-loss (case 5) regression test first and let it drive the shape of the lossy-vocab fixture; it is the highest-value guard in the slice per `AGENTS.md` abstraction-safety.
- **Verification:** `go test ./...` passes offline with no network/model access; the six fixtures are all exercised.

### U8. Documentation

- **Goal:** Document canonicalization semantics, the signature/fingerprint contract, resolution states, and the fixture provider, updating existing docs rather than duplicating contracts.
- **Requirements:** supports R1–R15 discoverability; repo `AGENTS.md` "update docs where contracts changed".
- **Dependencies:** U2–U7.
- **Files:** `docs/mechanism-canonicalization.md` (new), `README.md` (extend the executable-slice list with the new commands), light cross-links from `docs/domain-model.md`/`docs/persistence.md` only if a contract detail changed.
- **Approach:** Write the new doc: what canonicalization means and why it is not evidence promotion; the `mechanism/v1` signature shape; the five resolution states and the no-fuzzy-fallback rule; the fingerprint contract (what is/isn't hashed) including that **fingerprint equality is resolved-identity equality only** and is **conditional on a fixed `classifier_contract`**; comparison methods + weights versioning + the **`classify/v1` decisive-field rule (why `preserves`/`operators`/`assumptions` are decisive and posture/representations/boundaries are not)**; the abstraction-safety record answered by each mapping (compressed wording / retained distinction / lost / introduced) and the `AssertDiscriminationPreserved` invariant; an explicit **"fixtures test contract-consistency, not discrimination power"** caveat naming historical-holdout as the deferred primary benchmark; how to run the deterministic fixture path and the CLI with `--json`.
- **Test scenarios:** `Test expectation: none — documentation only.`
- **Verification:** New commands appear in `README.md`; the contract doc matches implemented behavior; no duplicated contract text.

---

## Verification Contract

- `go build ./cmd/newf` succeeds.
- `go test ./...` passes fully offline (no network, no model/provider calls) — satisfies R11.
- Fresh temp workspace end-to-end: `init` → load mechanism fixtures → `mechanism signature A` → `mechanism signature B` → `mechanism compare A B`, each with and without `--json`, producing stable IDs and fingerprints.
- Determinism: repeated `signature`/`compare` on the same inputs yield byte-identical `--json` (modulo run/timestamps), identical fingerprints.
- Immutability: UPDATE/DELETE on vocabulary and signature rows aborts at the DB layer.
- Version isolation: computing a signature under a second vocabulary version leaves the first signature and its fingerprint unchanged.
- Provenance: `claim_status` and `support` survive load → signature → `--json`; a canonical ID never changes a claim's status.
- Abstraction-loss regression (U7 case 5) is present and fails loudly if over-compression silently erases the outcome distinction.

## Definition of Done

Given two normalized research approaches (supplied via the deterministic fixture
substrate), `newf` deterministically produces canonical, provenance-preserving
`mechanism/v1` signatures with order-independent fingerprints, and explains how
the two are similar or different across assumptions, operators, preserved
properties, representations, boundaries, and posture — component-wise, without a
model making the identity decision. Ambiguous/unknown/novel mappings remain
explicit, surface similarity never defines mechanistic identity, vocabulary
evolution does not rewrite historical signatures, and the whole path runs offline
in CI. The resulting signatures are stable enough to become the input contract
for the next mechanism-clustering slice.

---

## Risks & Dependencies

- **#7 substrate is live, not designed (resolved).** #7 landed on this branch
  (migration v3, `internal/domain/normalize.go`, `internal/store/normalize.go`,
  `newf mechanism show`). #9 consumes it directly; U1 is only a fixture-seed loader
  over `PersistNormalization`, and new tables are v4/v5/v6. No table/type/command
  is recreated. If v3 is later renumbered by a rebase, only the migration constants
  shift — the additive v4/v5/v6 design is unaffected.
- **Fixtures are self-authored (closed-loop caveat).** The six fixtures and the
  tiny `mechanism/v1` vocabulary are co-designed, so they test *internal
  consistency and the resolution/fingerprint/compare contracts*, **not**
  discrimination power on unseen mechanisms. This is acceptable for the v0 slice
  but is explicitly **not** the historical-holdout evaluation `AGENTS.md` names as
  the primary benchmark; that evaluation is deferred to a later slice. Documented in
  U8 so no reader mistakes fixture pass for validated discrimination.
- **Fingerprint stability across Go versions.** Mitigation: canonical JSON with
  sorted keys + explicit field ordering + golden test (KTD3); never rely on
  struct/map iteration order.
- **Vocabulary design bikeshed.** Mitigation: keep `mechanism/v1` deliberately
  tiny (a handful of `core.*` + 1–2 `domain.number_theory.*` terms) — just enough
  to exercise all six fixtures; expansion is explicitly out of scope.
- **Over-reach into #7's classification.** Mitigation: A2 fixes the boundary —
  #9 canonicalizes already-classified labels and does not call a provider.

## Sources & Research

- Issue #9 (instagrim-dev/newf) — full requirements, six-fixture set, acceptance criteria.
- `AGENTS.md` — division of responsibility, epistemic invariants, abstraction safety, Go bias.
- `docs/persistence.md` — *design* reference for revisioned/immutable patterns; note the **live v3 schema is authoritative** where it diverges (posture is inline columns on `mechanisms`, not a `mechanism_axis_*` table; boundaries are `failure_boundaries(condition,ordinal)`).
- `docs/domain-model.md` — `MechanismAxis`, versioned vocabulary tables, ownership boundaries.
- `docs/abstraction-safety.md` — abstraction record + predictive-discrimination criterion (drives U7 case 5).
- `docs/cli-design.md` — package boundaries and `--json` envelope conventions.
- Existing code: `internal/domain/normalize.go` (Approach/Mechanism/Outcome/… types + `Validate()`), `internal/store/normalize.go` (`PersistNormalization`, `NormalizationInput`, `GetMechanismDetail`), `internal/store/store.go`, `internal/store/migrations.go` (v3 substrate + migration pattern), `internal/pipeline/source.go`, `cmd/newf/{source,approach}.go` (`newMechanismCommand`), `internal/domain/{id,source}.go` — established patterns to mirror and the substrate #9 consumes.

## Product Contract preservation

No upstream `ce-brainstorm` Product Contract exists; origin is GitHub issue #9.
Requirements R1–R15 trace directly to the issue's acceptance criteria and
definition of done.
