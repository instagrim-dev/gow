---
artifact_contract: ce-unified-plan/v1
artifact_readiness: implementation-ready
execution: code
product_contract_source: ce-plan-bootstrap
origin: GitHub issue #11 (instagrim-dev/newf) — "Implement mechanism clustering over canonical signatures"
epic: GitHub issue #10 (EPIC.md milestone M3 — Explicit failure-space)
depends_on: GitHub issue #9 (deterministic mechanism canonicalization and comparison) — landed on main (commit fefead0, schema v6)
created: 2026-09-10
plan_type: feat
---

# feat: Mechanism clustering over canonical signatures + first FailureSpace artifact

## Summary

Implement the next epic slice (EPIC.md **M3 — Explicit failure-space**, GitHub
issue **#11**): deterministic, inspectable, versioned **clustering** over the
canonical **mechanism signatures** produced by #9, plus the first explicit,
first-class **FailureSpace** artifact materialized from those clusters.

The governing constraint inherited from `AGENTS.md` and the #9 slice is *the
clustering layer must not hide behind embeddings*. Every grouping decision must
be reproducible from persisted, version-keyed inputs (signature fingerprints +
`canon.CompareWithProfile` field results under an explicit, persisted
**`ComparisonProfile`**) using an explicit, versioned, threshold-driven
agglomeration rule — never a learned metric, vector store, or model "are these
the same cluster?" call.

> **Alignment note (source state as of working tree after the #9 comparator
> correction).** The #9 comparator was refactored from a single hardcoded
> `decisiveFields` list into a caller-owned, versioned `ComparisonProfile`
> ("the comparator MEASURES every axis; the profile DECIDES which axes are
> decisive"). The refactor **explicitly names clustering (#11) as the consumer**
> that must be able to select the profile per experiment rather than inherit a
> hidden global definition of "mechanism". This plan therefore consumes
> `canon.CompareWithProfile(a, b, profile)` and `canon.ProfileMechanismV1()`
> (default: decisive set = `preserves, operator, assumption, breaks,
> auxiliary_object`; `representation` surface-only; `boundary`/`posture`/
> `outcome` measured but non-decisive by default), and **persists the profile
> version** on every cluster run. (At review time the comparator refactor is an
> in-progress uncommitted edit that does not yet compile — `Compare` return
> arity, a missing `classify(cmp, profile)`, and a dangling `decisiveFields`
> reference; this plan targets the refactor's *intended* profile contract, and
> U0 below records the dependency on that correction landing green.)

The exit condition (EPIC.md M3): given a problem's normalized approaches, `newf`
can state **how many mechanistically distinct families are represented, where
coverage is weak, and which failures are redundant** — and can emit a versioned
`FailureSpace` revision that later slices (#M4.1 failure-space enrichment, #M4.2
invariant mining) consume as their input contract.

This slice runs entirely offline against the existing project-authored
deterministic fixtures. It **consumes** the #9 substrate (`mechanism_signatures`,
`comparison_runs`, `internal/canon`) directly and adds only clustering +
failure-space tables (migrations **v7/v8**), a deterministic clustering engine in
a new pure package `internal/cluster`, pipeline services, and CLI surface.

---

## Problem Frame

After #9, `newf` can canonicalize a mechanism into a versioned signature with an
order-independent fingerprint and compare any two signatures component-wise. What
it **cannot** yet do is answer the population-level questions the failure-space
loop needs:

- How many *genuinely distinct* mechanism families are present among a problem's
  approaches (not "how many approaches")?
- Which approaches are *mechanistically redundant* (surface-distinct but
  mechanism-near) and therefore add no new failure information?
- Where is the mechanism space *under-sampled* (axes/families with thin or no
  coverage) — the signal that drives future frontier generation?

These are exactly the inputs the epic's next stages (candidate-invariant mining,
frontier generation) require, and they must be **trustworthy**: reproducible from
persisted inputs, provenance-preserving, and free of opaque similarity.

Today the repository has: problems/runs (v1), immutable sources/snapshots (v2),
the normalized-approach substrate (v3), and the canonicalization/comparison layer
(v4/v5/v6: `canonical_vocabulary`, `mechanism_signatures` + field/posture/
boundary/outcome projections, `comparison_runs`). This plan builds the clustering
+ failure-space layer **on top of** v6.

### Governing constraints (from `AGENTS.md`, `docs/abstraction-safety.md`, EPIC.md M3)

- **No embeddings / no learned metric / no model identity call.** Grouping is a
  deterministic function of persisted signature fingerprints and
  `canon.CompareWithProfile` field results under an explicit, persisted
  `ComparisonProfile`. (`AGENTS.md`: deterministic/independently-checkable
  operations are owned by code, not the model.)
- **Inspectable + versioned.** Any heuristic used to group mechanisms is explicit,
  named, and version-stamped (`cluster/v1`); a cluster run records the exact
  algorithm version, the **`ComparisonProfile.Version`** (which now subsumes the
  former `weights_version`/`classify_version` pair — both derive from the
  profile), vocabulary version(s), and thresholds so the grouping is replayable.
- **No epistemic-status promotion.** A cluster is a `CandidateMechanismFamily`
  (a grouping hypothesis), never an `EstablishedInvariant`. Membership carries the
  underlying signatures' provenance; clustering does not upgrade any claim's truth
  status (`explicit`/`inferred`/`ambiguous`/`unknown`/`unsupported` preserved).
- **Abstraction safety.** Clustering is an abstraction boundary: a cluster must
  record what it merged, on what basis, and what distinction it *erased*. If two
  approaches with **different outcome classes** are merged into one family purely
  because the grouping erased an outcome-predictive distinction, that is a
  discrimination-loss defect (regression fixture required — extends #9's
  `AssertDiscriminationPreserved`).
- **Incomparable is not "far".** Signatures with non-resolved (ambiguous/unknown)
  decisive fields cannot be asserted to agree; per #9's `FieldResult.Incomparable`,
  they must remain explicitly *unclustered* (or singletons) rather than being
  forced into a family by default.
- **Historical signatures stay reproducible.** Clustering keys on
  `(schema_version, vocabulary_version, profile_version, cluster_algo_version,
  thresholds_hash)`; a run under new versions creates a new cluster run and a new
  FailureSpace revision, never rewriting prior ones.

---

## Scope Boundaries

### In scope

- A pure, deterministic clustering engine (`internal/cluster`, no SQL/Cobra/
  provider) that consumes `canon.MechanismSignature` + pairwise `canon.Comparison`
  and produces stable, reproducible `MechanismCluster` assignments under an
  explicit, versioned threshold rule (`cluster/v1`).
- Cluster outputs: membership, a deterministically-chosen **representative
  signature** per cluster, **intra-cluster variation**, **inter-cluster distance**,
  and a **coverage/diversity report** (distinct-family count, redundancy set,
  under-sampled axes).
- The first explicit, first-class **FailureSpace** artifact: a versioned revision
  materialized from a cluster run, distinguishing outcome classes already present
  in signatures (`success`/`partial_failure`/`failure`/… from `signature_outcomes`)
  and preserving cluster coverage + provenance.
- Persistence (migrations **v7/v8**): immutable, version-keyed `cluster_runs`,
  `mechanism_clusters`, `cluster_members`, `failure_spaces`, and
  `failure_space_axes` (coverage), all with immutability triggers matching the
  repo pattern.
- CLI surface (all `--json`): `newf cluster build`, `newf cluster show`,
  `newf cluster list`, `newf failure-space build`, `newf failure-space show`,
  `newf failure-space coverage`.
- Deterministic fixture-driven tests (offline, no network/model): reproducibility,
  order-independence, version isolation, redundancy detection, discrimination-loss
  regression, and end-to-end integration over the existing mechanism fixtures.
- Docs: `docs/mechanism-clustering.md` + README section + EPIC.md M3 status note.

### Deferred to Follow-Up Work

- **FailureSpace enrichment (EPIC.md M4.1):** synthetic-failure injection,
  richer under-sampled-axis analytics beyond first-pass coverage, `failure-space`
  coverage deltas across revisions. This slice ships the *first* FailureSpace
  artifact; deeper analytics are the next issue.
- **Alternative clustering algorithms** (e.g. connected-components vs
  threshold-agglomerative variants) beyond the single versioned `cluster/v1` rule.
  The engine is built to admit a second `cluster_algo_version` without rewrite,
  but only one algorithm ships here.

### Out of scope (per issue #11 and EPIC.md, do not implement)

- Candidate-invariant mining (#M4.2) and invariant challenge (#M4.3).
- Embeddings / vector search / learned metric training / opaque similarity.
- Frontier generation, evaluation, search-policy mutation, historical-holdout
  scoring, cross-discipline portability.
- Any model call in the clustering path (clustering is fully deterministic code).
- Scalar "research quality" scores.

---

## Assumptions

1. **#9 is landed and stable on `main`** (commit `fefead0`/`1d5741f`,
   `currentSchemaVersion = 6`). `internal/canon` exports `MechanismSignature`,
   `BuildSignature`, `Fingerprint`, and — **after the comparator correction** —
   `CompareWithProfile(a, b, ComparisonProfile) Comparison`,
   `ProfileMechanismV1() ComparisonProfile`, plus the retained backward-compat
   `Compare(a, b, weightsVersion) (Comparison, error)` wrapper. `Comparison` now
   also carries a measured, non-decisive-by-default `Boundary FieldResult`. The
   default profile's decisive set is `{preserves, operator, assumption, breaks,
   auxiliary_object}`; `representation` is the surface qualifier. `internal/store`
   exposes `PersistSignature`/`GetSignature` and the
   `mechanism_signatures`/`comparison_runs` tables. **The alias correction**
   namespaces `canonical_term_aliases` by `field_kind` and permits multiple
   canonical ids per `(version, field_kind, alias_normalized)`, surfaced as an
   explicit `ambiguous` resolution rather than a silently-dropped binding — this
   feeds the `Incomparable`/ambiguous path clustering relies on (KTD-4).
   **Dependency:** the comparator refactor is currently an uncommitted,
   non-compiling in-progress edit; U0 gates this slice on it landing green.
2. **Signatures are the clustering input, not raw mechanisms.** A cluster run
   operates over the set of `mechanism_signatures` for a problem under one pinned
   `(schema_version, vocabulary_version)`; the plan adds a store reader to list
   signatures by problem if one is not already present.
3. **Determinism inputs are total.** Given the same signature set and the same
   pinned versions + thresholds, `cluster/v1` yields byte-identical membership,
   representatives, and fingerprints regardless of input ordering.
4. **Outcome class is already on the signature** (`signature_outcomes.class`), so
   the FailureSpace artifact partitions by existing enum values without inventing
   new outcome semantics.
5. **New ID classes** follow the existing ULID+prefix convention in
   `internal/domain/id.go` (add `ClusterRunIDPrefix`, `MechanismClusterIDPrefix`,
   `FailureSpaceIDPrefix` with matching `New*`/`Validate*` and `id_test.go`
   round-trip + cross-class rejection tests).
6. **Offline CI.** All tests use project-authored fixtures; no network/model.

If assumption (1) is ever false (surfaces renamed), re-cement against the live
`internal/canon` API before U2; do not fabricate an API shape.

---

## Requirements

- **R1. Deterministic, versioned clustering.** Group mechanism signatures for a
  problem into `MechanismCluster`s using an explicit `cluster/v1` rule over
  `canon.CompareWithProfile` results under a persisted `ComparisonProfile`;
  identical inputs + profile + versions + thresholds ⇒ identical output. No
  embeddings, no model call.
- **R2. Representative + variation + distance.** Each cluster exposes a
  deterministically-selected representative signature, an intra-cluster variation
  measure (component-wise), and inter-cluster distance is reported pairwise
  between representatives.
- **R3. Coverage/diversity report.** Report distinct-family count, the redundancy
  set (surface-distinct + mechanism-near members that add no new mechanism), and
  under-sampled axes (canonical field kinds / posture axes thinly represented).
- **R4. First-class FailureSpace artifact.** Materialize a versioned FailureSpace
  revision from a cluster run, partitioned by outcome class, preserving cluster
  coverage and full provenance back to signatures/mechanisms/snapshots.
- **R5. Provenance + epistemic preservation.** Cluster membership and FailureSpace
  entries retain the underlying signatures' provenance and claim statuses; nothing
  is promoted. A cluster is a candidate family, not an established invariant.
- **R6. Abstraction-loss guard.** Detect and refuse-to-silently-merge the case
  where clustering would merge different outcome classes only because a distinction
  was erased; surface it as an explicit discrimination-loss finding (regression
  fixture).
- **R7. Reproducibility across versions.** Cluster runs and FailureSpace revisions
  are immutable and version-keyed; re-running under new versions creates new rows,
  never rewrites history.
- **R8. CLI + `--json`.** All new commands expose stable human + JSON output.
- **R9. Offline determinism in CI.** End-to-end fixture tests prove stable cluster
  ids/fingerprints, order-independence, version isolation, and redundancy/
  discrimination-loss behavior with no network/model access.

Requirements (R1–R9) and implementation units (U1–U8) are separate axes.

---

## Key Technical Decisions

- **KTD-1: Clustering input is the persisted signature, keyed by version tuple.**
  A cluster run selects all `mechanism_signatures` for a problem sharing one
  `(schema_version, vocabulary_version)` and clusters those. Signatures under a
  different vocabulary/schema version are a *different* run. Rationale: keeps the
  grouping reproducible and prevents cross-version contamination (R1, R7).

- **KTD-2: `cluster/v1` = deterministic threshold-agglomerative over
  `canon.CompareWithProfile`, not a distance-matrix black box.** Two signatures
  are *same-family-linkable* iff `canon.CompareWithProfile(a,b,profile).Classification ∈
  {mechanism-near, surface-distinct+mechanism-near}` under the **persisted
  `ComparisonProfile`** (default `ProfileMechanismV1()`, version `classify/v1`;
  a caller may select a domain-specific profile, recorded on the run). Families
  are the connected components of the linkage graph over resolved-comparable
  pairs. Any pair that is `Incomparable` on a **decisive** field of the active
  profile does **not** create a link (KTD-4). Rationale: connected-components
  over an explicit, already-versioned classification is fully inspectable and
  reuses #9's profile-driven decisive-set rule verbatim — no new similarity
  math, no threshold guesswork beyond the classification #9 already owns (R1).
  The comparator's design note names clustering as the reason profiles are
  caller-owned: this slice **chooses and persists** the profile rather than
  inheriting a hidden global "mechanism" definition. A `min_link` threshold
  param exists and defaults to the classification rule; it is recorded in
  `thresholds_hash` so a future `cluster/v2` (or an alternate profile) can
  tighten linkage without rewriting `v1` runs.

- **KTD-3: Deterministic representative + deterministic cluster id.** The
  representative is the member with the lexicographically-smallest `fingerprint`
  (total order, provenance-excluded, already order-independent per #9). The
  cluster's own identity fingerprint is `sha256` over the sorted set of member
  fingerprints + the version tuple, so cluster identity is order-independent and
  reproducible (mirrors #9's `Fingerprint` discipline) (R1, R2, R7).

- **KTD-4: Incomparable stays unclustered.** A signature that is `Incomparable`
  with every other signature on a decisive field forms a **singleton** family
  flagged `incomparable_isolate`, never silently merged. This makes ambiguity
  visible in coverage rather than hidden by optimistic grouping (R5, R6).

- **KTD-5: Redundancy = intra-family `surface-distinct+mechanism-near` pairs.**
  The redundancy set is computed from stored per-field comparison results, not
  recomputed heuristically, so "which failures are redundant" is auditable back to
  a `comparison_runs` row (R3).

- **KTD-6: Under-sampled axes from canonical coverage counts.** For each canonical
  field kind and posture axis, count distinct canonical IDs / enum values present
  across families; axes below a versioned `min_axis_coverage` are reported as
  under-sampled. Deterministic and inspectable (R3).

- **KTD-7: FailureSpace partitions by existing `signature_outcomes.class`.** No new
  outcome ontology; the FailureSpace groups clusters by the outcome enum already on
  signatures and preserves member provenance (R4, R5).

- **KTD-8: Discrimination-loss guard reuses #9's invariant, lifted to families.**
  Extend the abstraction-safety check: if a single family contains members with
  materially different outcome classes whose *only* distinguishing decisive-field
  content was erased by the active (possibly lossy) vocabulary, emit a
  `DiscriminationLoss` finding and mark the cluster run `degraded` (still
  persisted, explicitly flagged) rather than presenting a clean family count (R6).

- **KTD-9: Everything immutable + additive migrations.** New tables in v7/v8 with
  the repo's `RAISE(ABORT, ...)` immutability triggers; no edits to v1–v6.

---

## High-Level Technical Design

### Data flow

```text
problem_id
  -> store.ListSignaturesForProblem(problem, schema_version, vocab_version)   [reader; U3]
       -> []canon.MechanismSignature (rehydrated from mechanism_signatures + projections)
  -> canon/cluster.BuildClustering(signatures, params{profile=canon.ProfileMechanismV1(), algo=cluster/v1, thresholds})  [U2]
       -> pairwise canon.CompareWithProfile(a,b,profile) (resolved-comparable only) -> linkage graph
       -> connected components -> []MechanismCluster{members, representative, intra_variation}
       -> inter-cluster distances (representative-vs-representative)
       -> coverage/diversity report (distinct families, redundancy set, under-sampled axes)
       -> AssertFamilyDiscriminationPreserved(...) -> []DiscriminationLoss   [U2, KTD-8]
  -> store.PersistClusterRun(run, clusters, members, coverage)   [immutable, version-keyed; U4]
  -> pipeline: newf cluster build|show|list (--json)             [U5]

cluster_run_id
  -> store.GetClusterRun(...) + signature_outcomes.class per member
  -> pipeline.BuildFailureSpace -> failure_spaces + failure_space_axes  [immutable revision; U6]
  -> newf failure-space build|show|coverage (--json)            [U6]
```

### Persistence shape (new tables — additive, on top of the existing v6 substrate)

Migration **v7** (clustering):

```sql
-- A single deterministic clustering pass over one version tuple of signatures.
CREATE TABLE cluster_runs (
  id TEXT PRIMARY KEY,                         -- clr_<ulid>
  problem_id TEXT NOT NULL REFERENCES problems(id),
  run_id TEXT NOT NULL REFERENCES runs(id),    -- provenance/exec record
  schema_version TEXT NOT NULL,                -- signature schema pinned
  vocabulary_version TEXT NOT NULL REFERENCES canonical_vocabulary(version),
  profile_version TEXT NOT NULL,               -- canon.ComparisonProfile.Version
                                               -- (subsumes former weights/classify pair)
  cluster_algo_version TEXT NOT NULL,          -- 'cluster/v1'
  thresholds_hash TEXT NOT NULL,               -- sha256 of canonical params json
                                               -- (includes the profile's decisive-set)
  signature_count INTEGER NOT NULL,
  family_count INTEGER NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('clean','degraded')),  -- KTD-8
  created_at TEXT NOT NULL,
  UNIQUE(problem_id, schema_version, vocabulary_version,
         profile_version, cluster_algo_version, thresholds_hash)
);

CREATE TABLE mechanism_clusters (
  id TEXT PRIMARY KEY,                          -- mcl_<ulid>
  cluster_run_id TEXT NOT NULL REFERENCES cluster_runs(id),
  cluster_fingerprint TEXT NOT NULL,            -- KTD-3, order-independent
  representative_signature_id TEXT NOT NULL REFERENCES mechanism_signatures(id),
  member_count INTEGER NOT NULL,
  intra_variation TEXT NOT NULL,                -- ordinal/component summary (json)
  isolate INTEGER NOT NULL DEFAULT 0,           -- 1 = incomparable_isolate (KTD-4)
  ordinal INTEGER NOT NULL,                     -- stable display order
  UNIQUE(cluster_run_id, cluster_fingerprint)
);

CREATE TABLE cluster_members (
  cluster_id TEXT NOT NULL REFERENCES mechanism_clusters(id),
  signature_id TEXT NOT NULL REFERENCES mechanism_signatures(id),
  mechanism_id TEXT NOT NULL REFERENCES mechanisms(id),
  redundant INTEGER NOT NULL DEFAULT 0,         -- KTD-5, surface-distinct+mech-near
  PRIMARY KEY(cluster_id, signature_id)
);

CREATE TABLE cluster_distances (
  cluster_run_id TEXT NOT NULL REFERENCES cluster_runs(id),
  cluster_a_id TEXT NOT NULL REFERENCES mechanism_clusters(id),
  cluster_b_id TEXT NOT NULL REFERENCES mechanism_clusters(id),
  classification TEXT NOT NULL,                 -- rep-vs-rep canon.Classification
  ordinal_summary TEXT NOT NULL,                -- per-field ordinal (json)
  PRIMARY KEY(cluster_run_id, cluster_a_id, cluster_b_id)
);
-- + immutability triggers (UPDATE/DELETE RAISE(ABORT,...)) for each table.
```

Migration **v8** (failure-space):

```sql
CREATE TABLE failure_spaces (
  id TEXT PRIMARY KEY,                           -- fsp_<ulid>
  problem_id TEXT NOT NULL REFERENCES problems(id),
  cluster_run_id TEXT NOT NULL REFERENCES cluster_runs(id),
  run_id TEXT NOT NULL REFERENCES runs(id),
  revision INTEGER NOT NULL,                     -- monotonic per problem
  distinct_family_count INTEGER NOT NULL,
  redundant_member_count INTEGER NOT NULL,
  created_at TEXT NOT NULL,
  UNIQUE(problem_id, cluster_run_id)
);

CREATE TABLE failure_space_axes (
  failure_space_id TEXT NOT NULL REFERENCES failure_spaces(id),
  axis_kind TEXT NOT NULL,                       -- field kind or posture axis
  distinct_value_count INTEGER NOT NULL,
  under_sampled INTEGER NOT NULL DEFAULT 0,      -- KTD-6
  PRIMARY KEY(failure_space_id, axis_kind)
);
-- + immutability triggers for both tables.
```

### Cluster linkage rule (deterministic, `cluster/v1`)

```text
for each unordered pair (a,b) of signatures under the pinned version tuple:
    cmp = canon.CompareWithProfile(a, b, profile)   -- profile owns the decisive set
    if any profile-decisive-field FieldResult.Incomparable:  no link      (KTD-4)
    elif cmp.Classification in {mechanism-near,
                                surface-distinct+mechanism-near}:  link(a,b)
    else:                                             no link
families = connected_components(nodes=signatures, edges=links)
singletons with zero comparable neighbors -> isolate=1 (incomparable_isolate)
representative(family) = argmin_fingerprint(member)                (KTD-3)
redundant(member) = exists intra-family pair classed surface-distinct+mechanism-near (KTD-5)
```

> Default profile decisive set (`ProfileMechanismV1`): `preserves, operator,
> assumption, breaks, auxiliary_object`. `representation` is the surface
> qualifier; `boundary` (now measured on `Comparison.Boundary`), `posture`, and
> `outcome` are measured but non-decisive unless a domain profile opts in. The
> exact decisive set is folded into `thresholds_hash` and pinned by
> `profile_version` so the grouping is reproducible.

---

## Output Structure

New pipeline views (in `internal/pipeline/output.go`), all `--json`-stable:

- `ClusterRunView` — id, problem, version tuple, `thresholds_hash`, family_count,
  status (`clean`/`degraded`), created_at.
- `MechanismClusterView` — id, cluster_fingerprint, representative signature id +
  fingerprint, member_count, intra_variation, isolate flag, members
  (signature id, mechanism id, redundant flag).
- `ClusterCoverageView` — distinct family count, redundancy set (member ids),
  under-sampled axes (axis kind + distinct value count).
- `ClusterDistanceView` — per representative-pair classification + ordinal summary.
- `FailureSpaceView` — id, revision, cluster_run_id, distinct_family_count,
  redundant_member_count, per-outcome-class family breakdown.
- `FailureSpaceCoverageView` — axes with under-sampled flags + distinct counts.

Human output mirrors #9's inspectable style (e.g. a coverage summary:
`7 distinct mechanism families; 3 redundant approaches; under-sampled: breaks, auxiliary_object`).

---

## Implementation Units

### U0. Dependency gate: the #9 comparator correction lands green
This slice consumes `canon.CompareWithProfile` / `canon.ProfileMechanismV1` and
the `Comparison.Boundary` field. At plan time these exist only as an
**uncommitted, non-compiling** in-progress edit (`Compare` return arity, missing
`classify(cmp, profile)`, dangling `decisiveFields`). U0 is satisfied when that
correction compiles and `go build ./... && go test ./internal/canon/...` is green
on the branch this slice builds from. If the profile API shape changes before it
lands, re-cement U2/KTD-2 against the final API. **No new code in U0** — it is a
precondition check, not a rewrite of #9's in-flight work.

### U1. New ID classes + domain scaffolding
Add `ClusterRunIDPrefix = "clr_"`, `MechanismClusterIDPrefix = "mcl_"`,
`FailureSpaceIDPrefix = "fsp_"` to `internal/domain/id.go` with `New*`/`Validate*`
and matching `id_test.go` round-trip + cross-class rejection tests. Add any small
pure domain enums the clustering result needs (e.g. `ClusterStatus`
`clean|degraded`, isolate flag) with `Valid()` methods. **No SQL/Cobra here.**

### U2. Pure clustering engine `internal/cluster` (no SQL/Cobra/provider)
`BuildClustering(signatures []canon.MechanismSignature, params ClusterParams)
(Clustering, []canon.DiscriminationLoss, error)` where
`ClusterParams{Profile canon.ComparisonProfile; AlgoVersion string; …thresholds}`
carries the caller-owned profile. Internally uses
`canon.CompareWithProfile(a, b, params.Profile)` and treats a pair as linkable
only when no **profile-decisive** field is `Incomparable` (KTD-2/KTD-4).
Implements KTD-3..KTD-6 and the lifted discrimination-loss guard (KTD-8). Pure
functions over `canon` types; deterministic, order-independent (sort inputs by
fingerprint internally). Unit tests: linkage rule under `ProfileMechanismV1`,
connected-components, representative selection, cluster fingerprint
order-independence, redundancy detection, isolate handling, under-sampled-axis
counting, discrimination-loss on the lossy vocabulary, **and a profile-swap test
proving an alternate decisive set (e.g. making `representation` decisive) changes
family assignments deterministically and is recorded via `profile_version`.**

### U3. Signature-set reader
Add `store.ListSignaturesForProblem(ctx, problemID, schemaVersion, vocabVersion)
([]SignatureRecord, error)` (or reuse an existing reader if present) plus a
rehydrator to `canon.MechanismSignature`. Store test over seeded fixtures.

### U4. Cluster persistence (migration v7) + writer/reader
Migration v7 tables + immutability triggers. `store.PersistClusterRun(...)`
(transactional, idempotent on the `UNIQUE` version tuple — re-run returns the
existing run) and `store.GetClusterRun(...)`/`ListClusterRuns(...)`. Store tests:
idempotency, immutability (raw UPDATE/DELETE aborts), full round-trip.

### U5. Cluster pipeline services + CLI
`pipeline` service wiring: `newf cluster build --problem <id> [--vocabulary
<version>] [--schema-version <v>] [--profile <profile-version>] [--force]`,
`newf cluster show <cluster-run-id>`, `newf cluster list --problem <id>`, all
`--json`. `--profile` selects the persisted `ComparisonProfile` (default
`classify/v1` = `ProfileMechanismV1`); the resolved profile version is recorded
on the run. Register in `cmd/newf/root.go`; CLI integration tests (build → show →
list; `--force` creates a new run only when a version/profile/threshold differs,
else returns existing).

### U6. FailureSpace artifact (migration v8) + services + CLI
Migration v8 tables + triggers. `pipeline.BuildFailureSpace(clusterRunID)`
partitioning by `signature_outcomes.class`, computing coverage axes (KTD-6/7).
CLI: `newf failure-space build --problem <id>` (builds from latest/`--cluster-run`),
`newf failure-space show <id>`, `newf failure-space coverage <id>`, all `--json`.
Store + CLI integration tests including revision monotonicity and immutability.

### U7. Deterministic fixtures + regressions + end-to-end integration
Reuse the six #9 mechanism fixtures; add fixture coverage for: (a) two
surface-distinct + mechanism-near approaches → same family + one flagged
redundant; (b) surface-near + mechanism-distinct → distinct families; (c) an
ambiguous/unknown decisive field → `incomparable_isolate` singleton; (d) lossy
vocabulary merges differing outcome classes → `degraded` run + `DiscriminationLoss`
finding (regression); (e) input-ordering permutation → identical cluster ids/
fingerprints/failure-space. Full offline end-to-end test:
`seed → signature → cluster build → failure-space build → coverage`.

### U8. Documentation
`docs/mechanism-clustering.md` (algorithm `cluster/v1`, determinism contract,
persistence, CLI, epistemic/abstraction constraints), README section, and an
EPIC.md **M3** status note. Update `docs/persistence.md` with v7/v8.

---

## Verification Contract

- `go build ./...`, `go vet ./...`, `go test ./...` green; `gofmt -l .` empty.
- **R1/R9 (determinism):** a test permutes signature input order and asserts
  byte-identical cluster ids, cluster fingerprints, family_count, and resulting
  FailureSpace.
- **R7 (version isolation + immutability):** re-running `cluster build` with the
  same version tuple returns the existing run (idempotent); a differing
  vocabulary/threshold creates a new run; raw `UPDATE`/`DELETE` on any v7/v8 table
  is aborted by trigger (direct SQL probe, mirroring the #6/#9 immutability tests).
- **R3 (coverage):** fixtures assert distinct-family count, redundancy set, and
  under-sampled axes match expected values.
- **R6/KTD-8 (discrimination loss):** the lossy-vocabulary fixture yields a
  `degraded` cluster run + explicit `DiscriminationLoss` finding; the non-lossy
  vocabulary over the same inputs is `clean`.
- **R4/R5 (failure-space provenance):** every FailureSpace member resolves back to
  its signature → mechanism → snapshot; no claim status is upgraded.
- **R8:** every new command emits stable `--json` (contract test in
  `output_test.go`).
- CI runs entirely offline (no network/model).

## Definition of Done

Given a problem's normalized approaches, `newf` can deterministically group their
mechanism signatures into inspectable, versioned families; state how many
mechanistically distinct families exist, which approaches are redundant, and which
axes are under-sampled; and emit an immutable, provenance-preserving FailureSpace
revision — all offline, with no embeddings and no model making the grouping
decision. The FailureSpace artifact is stable enough to become the input contract
for the next slice (M4.1 failure-space enrichment / M4.2 invariant mining).

## Risks & Dependencies

- **Depends on #9** (`internal/canon` + `mechanism_signatures`/`comparison_runs`,
  schema v6). If the canon API differs from Assumptions §1 at implementation time,
  re-cement U2 against the live API before writing the engine.
- **Connected-components can over-merge via transitivity** (a~b, b~c ⇒ a,b,c one
  family even if a≁c). This is an accepted, *documented* property of `cluster/v1`
  and is exactly why the discrimination-loss guard (KTD-8) and redundancy reporting
  exist; a future `cluster/v2` may add linkage tightening via the already-present
  `thresholds_hash` without rewriting v1 runs.
- **Isolates are a feature, not a bug:** ambiguous signatures must stay visible as
  singletons; tests lock this so future "tidy up singletons" changes cannot silently
  absorb them.

## Sources & Research

- GitHub issue #11 (this slice) and #10 / `EPIC.md` **M3 — Explicit failure-space**.
- GitHub issue #9 + `docs/mechanism-canonicalization.md` (comparison/signature/
  fingerprint contract this slice consumes).
- `AGENTS.md` (no-embedding, epistemic-status, provenance constraints),
  `docs/abstraction-safety.md` (abstraction-loss / discrimination-preservation),
  `docs/toolbox-dsl.md` (operator semantics), `docs/persistence.md`.
- Live tree surfaces: `internal/canon/{signature,fingerprint,compare}.go`,
  `internal/store/{canon_store,migrations}.go`, `internal/domain/id.go`.

## Product Contract preservation

- No SQL/Cobra in `internal/domain` or `internal/cluster`; SQLite behind
  `internal/store`; CLI wiring thin in `cmd/newf`.
- No provider coupling anywhere in this slice (clustering is deterministic).
- No epistemic-status promotion: clusters are `CandidateMechanismFamily`
  groupings; the FailureSpace preserves member claim statuses and provenance.
- Additive migrations v7/v8 with immutability triggers; historical runs/revisions
  never rewritten.
