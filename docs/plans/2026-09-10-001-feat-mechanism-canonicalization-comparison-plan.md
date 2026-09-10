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

**Settled scope decision (session-settled: user-directed — chosen over
implementing #7 or blocking on #7): plan #9 as if #7 (approach normalization) is
already complete.** The plan assumes the `normalization_revision` /
`approach` / `mechanism` / `mechanism_representation|assumption|operator|preserves`
/ `mechanism_axis_value` / `outcome` / `failure_boundary` substrate designed in
`docs/persistence.md` already exists and is populated. Where that substrate is
needed to run and test #9 in isolation, the plan builds the minimal typed
mechanism-record tables and a fixture loader that stands in for #7's provider-driven
`normalize` — see U1 and the Assumptions section. This keeps #9 shippable and
fully offline while leaving #7's provider path to its own issue.

---

## Problem Frame

`newf` cannot cluster or mine invariants over prose similarity. It needs a stable,
provenance-preserving, deterministic representation of *mechanism structure* so
that two approaches can be compared component-by-component (assumptions,
operators, preserved properties, representations, boundaries, posture) without a
model making the final identity decision.

Today the repository has problems/runs (migration v1) and immutable
sources/snapshots (v2). The normalized-approach substrate is fully *designed* in
`docs/persistence.md` and `docs/domain-model.md` but not yet implemented. This
plan implements the canonicalization/comparison/fingerprint layer and the minimal
mechanism-record substrate it consumes.

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

- **A1 — #7 substrate treated as existing.** Per the settled scope decision, the
  normalized `approach` / `mechanism` / `mechanism_*` tables exist. Because they are
  not yet implemented in the repo, **U1 creates the minimal subset of those exact
  tables** (matching `docs/persistence.md` names/columns so #7 can extend, not
  rewrite) plus a fixture loader. If, at implementation time, #7 has already landed
  these tables, U1 reduces to "reuse existing tables + add the fixture loader path";
  do not duplicate or rename them.
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

- **KTD3 — Fingerprint = SHA-256 over a canonical JSON serialization of `{schema_version, vocabulary_version, sorted canonical-ID sets per field, posture enums, outcome.class, sorted boundary relations}` with provenance excluded.** Sets are sorted by canonical ID; posture is fixed-key ordered; JSON uses sorted keys and no insignificant whitespace. Rationale: reuses `crypto/sha256` already in the tree (`internal/pipeline/source.go`); order-independence and version-visibility fall out of the canonical serialization (satisfies R5). Provenance is excluded from the hashed body by default; a separate `provenance_fingerprint` may be added later if needed (not in v0).

- **KTD4 — Per-field comparison methods are explicit and typed, keyed by field kind.** Unordered canonical-ID sets (operators, assumptions, preserves, breaks, representations, auxiliary_objects) → set overlap + Jaccard (both reported). Posture (locality/construction/uncertainty) → enum equality. Boundaries → typed relation comparison over canonical IDs. Where the vocabulary declares parent/child, an explicit hierarchy distance is available. Rationale: matches the issue's "possible v0 choices" and the abstraction-safety requirement that behavior be inspectable, not buried in a distance function (R6, R13).

- **KTD5 — No single scalar in v0.** Comparison output is a per-field structure plus an ordinal similarity per field (`identical`/`high`/`low`/`none`/`incomparable`) and a categorical mechanistic-vs-surface classification. Surface similarity is computed only if fixtures/inputs carry surface text, and is tagged `auxiliary_only`. Rationale: issue explicitly forbids fake precision and forbids surface similarity defining identity (R6, R7, R12).

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

### Persistence shape (new tables, names aligned to `docs/persistence.md`)

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
    canonical.go                 # NEW — CanonicalID, ResolutionState, ClaimStatus, Posture enums + validation
    mechanism.go                 # NEW (A1) — minimal Approach/Mechanism/NormalizationRevision types
  store/
    migrations.go                # EXTEND — v4 (vocab/rubric), v5 (signatures), v6 (comparison); v3 mechanism substrate already exists
    canon_store.go               # NEW — vocab/signature/comparison repositories
    mechanism_store.go           # NEW (A1) — minimal mechanism-record repo + fixture loader
  pipeline/
    mechanism.go                 # NEW — SignatureMechanism / CompareMechanisms services
    vocabulary.go                # NEW — Vocabulary list/show/resolve services
    output.go                    # EXTEND — response view structs
cmd/newf/
    mechanism.go                 # NEW — `newf mechanism signature|compare`
    vocabulary.go                # NEW — `newf vocabulary list|show|resolve`
    root.go                      # EXTEND — register commands + error classes
testdata/
  fixtures/mechanism/            # NEW — 6 required deterministic fixtures
docs/
  mechanism-canonicalization.md  # NEW — contract/semantics doc
```

---

## Implementation Units

### U1. Minimal mechanism-record substrate + fixture loader (stands in for #7)

- **Goal:** Provide the assumed-existing normalized mechanism records #9 consumes, plus a deterministic offline loader that populates them from fixtures.
- **Requirements:** R1 (input side), R11, R15; enables A1/A2/A3.
- **Dependencies:** none.
- **Files:** `internal/domain/mechanism.go`, `internal/store/mechanism_store.go`, `internal/store/migrations.go` (migration v3), `internal/store/mechanism_store_test.go`, `testdata/fixtures/mechanism/` (initial fixtures used by loader tests).
- **Approach:**
  1. Add migration v3 creating the minimal subset of `docs/persistence.md` tables needed to hold a mechanism: `normalization_revision` (minimal columns), `approach`, `mechanism`, `mechanism_representation|assumption|operator|preserves`, `mechanism_axis_value`, `outcome`, `failure_boundary`, `outcome_boundary`. Use the exact names/columns from `docs/persistence.md` so #7 extends rather than rewrites. Include the immutable-source-style triggers only where the design marks these revisioned/append-only.
  2. Add domain types `NormalizationRevision`, `Approach`, `Mechanism` (with representation/assumptions/operators/preserves slices + axis values), `Outcome`, `FailureBoundary`, each with `Validate()`, following the `NewSource`/`Validate()` pattern. Include a `mch_` ID prefix (and any others needed) in `internal/domain/id.go`.
  3. Add a fixture loader on the store that reads a mechanism fixture (JSON) and inserts a normalization revision + approach(es) + mechanism(s) transactionally. This is the offline stand-in for #7's provider `normalize`.
- **Patterns to follow:** `internal/domain/source.go` (`New*`/`Validate`), `internal/store/store.go` (`CreateSourceSnapshot` transaction + scan helpers), `internal/store/migrations.go` (append a new numbered migration; bump `currentSchemaVersion`; add tables to `validateSchemaTables`).
- **Test scenarios:**
  - Migration v3 applies fresh and is idempotent; `SchemaVersion` reflects the bump.
  - Loading a single-mechanism fixture creates one revision, one approach, one mechanism with all field slices populated.
  - Loading a multi-approach fixture creates multiple approaches under one revision (issue: one source → multiple approaches).
  - Field-value rows preserve original surface labels (no canonicalization here).
  - Invalid fixture (missing required field) fails the transaction leaving no partial rows.
  - `Validate()` rejects bad IDs / empty required fields.
- **Verification:** A fixture file loads into queryable mechanism rows; `go test ./internal/store/...` passes offline.

### U2. Canonical domain types + versioned vocabulary with deterministic resolution

- **Goal:** Define the semantic-spine vocabulary (`mechanism/v1` terms), canonical-ID/resolution-state/claim-status types, and the pure deterministic `resolve()`.
- **Requirements:** R2, R4, R14, R15, KTD2, KTD6.
- **Dependencies:** U1 (for `field_kind` alignment only; can proceed in parallel on pure types).
- **Files:** `internal/domain/canonical.go`, `internal/canon/vocabulary.go`, `internal/canon/vocabulary_seed.go`, `internal/canon/vocabulary_test.go`.
- **Approach:**
  1. `internal/domain/canonical.go`: `CanonicalID string` with validation of the namespaced shape (`core.<kind>.<name>` / `domain.<field>.<kind>.<name>`); `ResolutionState` enum (`resolved|ambiguous|novel_candidate|unknown|rejected`) with an exhaustive-check helper; `ClaimStatus` enum (`explicit|inferred|ambiguous|unknown|unsupported`); `PostureAxis`/values enums (locality: local|global|mixed|unknown; construction: constructive|existential|mixed|unknown; uncertainty: deterministic|probabilistic|unknown); `FieldKind` enum for the spine (`representation|operator|assumption|preserves|breaks|auxiliary_object|outcome|boundary|posture`) — the set-field kinds match the existing `mechanism_attributes.kind` CHECK values exactly (`representation|assumption|operator|preserves|breaks|auxiliary_object`).
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
- **Files:** `internal/store/canon_store.go` (vocab portion), `internal/store/migrations.go` (migration v4, vocab + rubric tables), `internal/pipeline/vocabulary.go`, `internal/pipeline/output.go` (vocab views), `cmd/newf/vocabulary.go`, `cmd/newf/root.go` (register + error classes), `internal/store/canon_store_test.go`, `internal/pipeline/vocabulary_test.go`, `cmd/newf/root_test.go` (CLI JSON).
- **Approach:**
  1. Migration v4: `canonical_vocabulary`, `canonical_term`, `canonical_term_alias`, `classification_rubric` with immutability triggers; add to `validateSchemaTables`; bump `currentSchemaVersion`.
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
- **Files:** `internal/store/canon_store.go` (signature portion), `internal/store/migrations.go` (extend v4 or add v5 with signature tables), `internal/pipeline/mechanism.go` (SignatureMechanism service), `internal/pipeline/output.go` (signature views), `cmd/newf/mechanism.go`, `cmd/newf/root.go` (register), `internal/store/canon_store_test.go`, `internal/pipeline/mechanism_test.go`, `cmd/newf/root_test.go`.
- **Approach:**
  1. Migration: `mechanism_signature` (+ `UNIQUE(mechanism_id, schema_version, vocabulary_version)`), `signature_field_claim`, `signature_posture`, `signature_boundary`, `signature_outcome`, with append-only immutability triggers; register in `validateSchemaTables`; bump version.
  2. `SignatureMechanism` service: load mechanism (U1), resolve vocab version (default = latest seeded), `BuildSignature`, persist all rows transactionally under a new `run`, return the view. If a signature already exists for `(mechanism, schema, vocab)`, return it as `existing` (idempotent, matching the snapshot-admission `existing`/`created` status pattern) rather than recomputing/rewriting.
  3. CLI `mechanism signature <mechanism-id>` with `--vocab-version`, `--schema-version` (default `mechanism/v1`), `--json`. Human output shows fingerprint, per-field canonical IDs with resolution state + claim status, posture, outcome, boundaries. `--json` returns the full signature incl. provenance.
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
- **Files:** `internal/canon/compare.go`, `internal/canon/compare_test.go`, `internal/store/canon_store.go` (comparison_run persistence), `internal/store/migrations.go` (comparison tables), `internal/pipeline/mechanism.go` (CompareMechanisms service), `internal/pipeline/output.go` (comparison views), `cmd/newf/mechanism.go` (compare subcommand), `internal/pipeline/mechanism_test.go`, `cmd/newf/root_test.go`.
- **Approach:**
  1. `Compare(sigA, sigB, weights)`: for each set field compute intersection/union over **resolved** canonical IDs → `{overlap_count, jaccard}` and an ordinal (`identical|high|low|none`); elements with non-`resolved` state on either side are reported `incomparable` for that field rather than counted as agreement. Posture axes → enum equality. Boundaries → typed relation comparison. Weights come from an explicit, versioned `weights_version` table/struct (default `weights/v1`), configurable via flag; never collapse to one scalar.
  2. Classification: derive `mechanism-near`/`mechanism-distinct` from the aggregate of canonical-field ordinals (explicit, documented rule, e.g., near if preserves+operators+assumptions are high/identical and no material distinct field); compute optional surface similarity **only** from any surface text carried on inputs, tag it `auxiliary_only`, and combine into `surface-distinct+mechanism-near` / `surface-near+mechanism-distinct` / `mechanism-near` / `mechanism-distinct` / `unknown`. Surface similarity never feeds the mechanistic decision (R7).
  3. Persist `comparison_run` + `comparison_field_result` when not `--no-write`; return the view.
  4. CLI `mechanism compare <a> <b>` (`--weights-version`, `--no-write`, `--json`) producing the issue's example human table + full `--json`.
- **Patterns to follow:** `cmd/newf/source.go` two-arg command; `writeJSON`; deterministic table-driven tests.
- **Test scenarios:**
  - Identical signatures → every field `identical`, classification `mechanism-near`.
  - Same canonical IDs reached from different surface wording (the surface-distinct/mechanism-near fixture) → `preserves: identical`, classification `surface-distinct+mechanism-near`.
  - Textually similar but mechanistically different fixture → low canonical-field overlap, classification `surface-near+mechanism-distinct` (proves surface text does not drive identity, R7).
  - Jaccard math correct on known sets; enum posture equality correct.
  - Fields with ambiguous/unknown claims are reported `incomparable`, not silently equal.
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
  5. an **over-compression** case: under a deliberately lossy vocabulary version, two mechanisms with different outcomes canonicalize identically — the test asserts this collapse is **detectable** (e.g., signatures/fingerprints equal under lossy vocab but the mechanisms carry different `outcome.class`), demonstrating the defective-normalization failure mode;
  6. **ordering differences** that must yield the **same** fingerprint.
  Then an integration test that, in a temp workspace: migrates, loads fixture A and fixture B, builds both signatures, compares them, and asserts the expected classification per fixture pair.
- **Test scenarios / assertions:**
  - Case 1: fixtures A/B → equal fingerprints, `compare` → `surface-distinct+mechanism-near` (or `mechanism-near`).
  - Case 2: equal/near surface text → `mechanism-distinct` (`surface-near+mechanism-distinct`); asserts R7.
  - Case 3: the ambiguous field stays `ambiguous` through signature + comparison (`incomparable`), never guessed.
  - Case 4: the novel label persists as `novel_candidate`; it is not coerced to an existing canonical ID.
  - Case 5 (R9 regression): under lossy vocab the two different-outcome mechanisms collapse to one fingerprint; the test documents/asserts detection so the abstraction defect cannot pass silently. Under the correct `mechanism/v1` vocab they remain distinguishable.
  - Case 6: reordered field values → identical fingerprint.
  - Integration: signatures stable across runs; fingerprints deterministic; ambiguous fields stay ambiguous; no provenance lost end-to-end (support/status survive load→signature→json).
- **Execution note:** Author the abstraction-loss (case 5) regression test first and let it drive the shape of the lossy-vocab fixture; it is the highest-value guard in the slice per `AGENTS.md` abstraction-safety.
- **Verification:** `go test ./...` passes offline with no network/model access; the six fixtures are all exercised.

### U8. Documentation

- **Goal:** Document canonicalization semantics, the signature/fingerprint contract, resolution states, and the fixture provider, updating existing docs rather than duplicating contracts.
- **Requirements:** supports R1–R15 discoverability; repo `AGENTS.md` "update docs where contracts changed".
- **Dependencies:** U2–U7.
- **Files:** `docs/mechanism-canonicalization.md` (new), `README.md` (extend the executable-slice list with the new commands), light cross-links from `docs/domain-model.md`/`docs/persistence.md` only if a contract detail changed.
- **Approach:** Write the new doc: what canonicalization means and why it is not evidence promotion; the `mechanism/v1` signature shape; the five resolution states and the no-fuzzy-fallback rule; the fingerprint contract (what is/ isn't hashed); comparison methods + weights versioning + mechanistic-vs-surface classification; the abstraction-safety record answered by each mapping (compressed wording / retained distinction / lost / introduced); how to run the deterministic fixture path and the CLI with `--json`.
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

- **#7 substrate overlap (R-risk).** U1 creates a subset of tables that #7 will
  also want. Mitigation: use the exact `docs/persistence.md` names/columns so #7
  extends via later migrations; if #7 lands first, U1 degrades to "reuse + add
  fixture loader" (see A1). Reconcile forward, do not rename.
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
- `docs/persistence.md` — canonical table names/columns for mechanism substrate and revisioned/immutable patterns.
- `docs/domain-model.md` — `MechanismAxis`, versioned vocabulary tables, ownership boundaries.
- `docs/abstraction-safety.md` — abstraction record + predictive-discrimination criterion (drives U7 case 5).
- `docs/cli-design.md` — package boundaries and `--json` envelope conventions.
- Existing code: `internal/store/store.go`, `internal/store/migrations.go`, `internal/pipeline/source.go`, `cmd/newf/source.go`, `internal/domain/{id,source}.go` — established patterns to mirror.

## Product Contract preservation

No upstream `ce-brainstorm` Product Contract exists; origin is GitHub issue #9.
Requirements R1–R15 trace directly to the issue's acceptance criteria and
definition of done.
