# Mechanism canonicalization and comparison

This document is the contract for `newf`'s deterministic mechanism-identity
layer (issue #9). It builds on the normalized `Approach` / `Mechanism` /
`Outcome` / `FailureBoundary` substrate produced by `newf normalize` (see
[`normalization.md`](normalization.md)).

The governing principle is:

> Models discover candidate labels; software owns identity.

`newf` must never persist a model's "are these the same?" answer as truth. The
fuzzy, truth-sensitive step (semantic classification of a surface phrase into a
field label) belongs to normalization / a later slice. This layer owns only the
**deterministic** operations: resolving already-classified labels to canonical
IDs, building versioned signatures, hashing fingerprints, and comparing
signatures component-wise.

## What canonicalization is (and is not)

Canonicalization maps a classified surface label (e.g. `"works
residue-by-residue"`) to a stable, versioned **canonical ID** (e.g.
`domain.number_theory.property.residue_locality`). The canonical ID is the
comparison identity; the original wording is preserved as provenance.

Canonicalization is **not** evidence promotion. A canonical ID never upgrades
the epistemic status of the underlying claim: an `inferred` claim stays
`inferred` after it resolves. Ambiguous/unknown/novel labels stay explicit and
are never coerced to the nearest term.

## Canonical IDs

Canonical IDs are namespaced and lower_snake:

```text
core.<kind>.<name>                    e.g. core.operator.modular_decomposition
domain.<field>.<kind>.<name>          e.g. domain.number_theory.property.residue_locality
```

## The vocabulary

A **vocabulary** is an immutable, versioned semantic spine (`mechanism/v1`). Each
term has a canonical ID, a field kind, optional aliases, and an optional parent
(for hierarchy distance). The vocabulary is seeded from an in-repo definition on
first store use and is immutable per version at the database layer (triggers).

Revisions never mutate an existing version: `mechanism/v2` is a strict
superset of `mechanism/v1` adding the three shared-property concepts accepted
by the pilot-001 mapping review (adjudication ledger L1/L3/L4). Those terms
are GeneratedInterpretation hypotheses — the only source labels they alias
are ones that directly NAME the property (es-08's obstruction label for L1,
es-02's preserves label for L4); no alias merges two distinct operations, and
es-12's form-indexed variant deliberately stays unresolved as a preserved
distinction. Property claims reach signatures through `interpretation_claims`
(migration `v33`, `docs/persistence.md`), entering as `inferred` claims with
`interpretation:` provenance — never by editing sources.

`mechanism/v3` (recovery-calibration correction) is a strict superset of v2
adding terms that canonicalize what the withheld benchmark target's payload
explicitly states, so its decisive fields become comparable and
`recovery-rule/v1` can reach a positive match. These are ordinary
canonicalizations, not interpretations; the breaks-kind QR-confinement term
is a field-kind-scoped sibling of the v2 preserves-kind concept (following
the v1 `residue_locality` / `residue_class_locality` precedent), and the
documented alias additions to existing terms exist only in v3
(`mechanismV3ExtraAliases`). A superset regression pins v2-in-v3.

The set-valued field kinds (`representation`, `assumption`, `operator`,
`preserves`, `breaks`, `auxiliary_object`) match the existing
`mechanism_attributes.kind` values one-to-one, so canonicalization maps directly
onto attribute rows. Posture (`locality`, `construction`, `uncertainty`) is
**not** re-canonicalized: it reuses the existing validated enum columns on the
`mechanisms` row, so there is no parallel posture ontology.

### Resolution states

Resolution is a deterministic lookup against a vocabulary version, with **no
fuzzy/embedding fallback**. The five fixed states are:

| State | Meaning |
|---|---|
| `resolved` | Matched exactly one canonical ID or alias. |
| `ambiguous` | Matched more than one candidate (all recorded, none chosen). |
| `novel_candidate` | No match, but the caller flagged the label as novel. |
| `unknown` | No match, not flagged novel. |
| `rejected` | On the vocabulary's explicit rejected-terms list (persisted; see below). |

Only `resolved` claims contribute a canonical ID to the fingerprint body. The
other states are persisted with their surface label and state, and are surfaced
by comparison as `incomparable` — never silently dropped or coerced.

### Provenance is preserved, never promoted

Every field claim carries a preserved `claim_status` (`explicit` / `inferred` /
`unsupported` / `unknown`). Canonicalization never upgrades it: a claim whose
label resolves to a canonical ID keeps its original status. Crucially, **absence
of a `source_support` row does not mean explicit** — #7 does not require support
for every populated field, so a missing row maps to `unknown`, not `explicit`.
Defaulting to explicit would silently promote an unprovenanced value into a
source-backed claim, violating the epistemic-status hard constraint; this is
guarded by the provenance assertions in the mechanism integration test.

This holds for the non-vocabulary fields too: posture axes, outcome class, and
boundaries each carry their own preserved claim status through
`MechanismSignature`. The support field paths are the verbatim #7 normalize JSON
keys: list attributes are plural (`mechanism.operators`,
`mechanism.representations`, `mechanism.assumptions`, `mechanism.preserves`,
`mechanism.breaks`, `mechanism.auxiliary_objects`), posture is
`mechanism.locality` / `mechanism.construction_mode` / `mechanism.uncertainty_mode`,
outcome is `outcome.class`, and the boundary is `outcome.boundary_statement`.
Because `source_support.field_path` is free-form in #7, the #9 reader resolves
each field against a canonical-first list of accepted paths
(`domain.AttributeSupportPaths` / `PostureSupportPaths` / `OutcomeSupportPaths`),
tolerating earlier singular aliases (e.g. `mechanism.operator`) so authored
provenance is never demoted to `unknown` by spelling drift. The
`outcome.boundary_statement` support attaches only to the statement-derived
boundary; enumerated `boundary_conditions` have no per-condition support path and
carry `unknown` rather than inheriting the statement's provenance (no broadcast).
Unprovenanced posture/outcome round-trips as `unknown`
(`TestSignatureCarriesPostureOutcomeBoundaryProvenance`); the end-to-end join is
guarded by `TestIntegrationCanonicalSupportPathsJoin`.

### Alias namespace (per field kind)

Aliases are namespaced by field kind: the resolver keys the alias index on
`(field_kind, normalized_alias)`, and persistence keys `canonical_term_aliases`
on `(vocabulary_version, field_kind, alias_normalized, canonical_id)`. Two
consequences follow, both intentional and now consistent between the in-memory
resolver and SQLite:

- The **same phrase may map to different canonical IDs in different field
  kinds** (e.g. an "averaging" operator vs. an "averaging" assumption). The
  earlier alias key omitted `field_kind`, so a phrase reused across field kinds
  was silently dropped at seed time by `INSERT OR IGNORE`; the corrected key
  represents both bindings. Guarded by `TestAliasNamespaceIsPerFieldKind`.
- Within a single field kind a phrase **may** still bind to more than one
  canonical ID; the resolver reports this as the explicit `ambiguous` state
  (candidates recorded, none chosen). Including `canonical_id` in the primary
  key lets the store represent that legitimately while still forbidding
  exact-duplicate rows, so `INSERT OR IGNORE` can never collapse two *different*
  bindings.

```bash
newf vocabulary list
newf vocabulary list --version mechanism/v1 --field operator
newf vocabulary show core.operator.modular_decomposition
newf vocabulary resolve "works residue-by-residue" --field preserves
newf vocabulary resolve "a brand new move" --field operator --novel
```

## Signatures and fingerprints

`newf mechanism signature <mechanism-id>` builds a versioned
`mechanism/v1` **signature**: every spine field is an always-present key
(possibly empty) holding canonicalized field claims, plus posture, outcome
class, and boundaries. Support/status/classifier-contract provenance is copied
through unchanged.

To get mechanism IDs offline without a provider, `newf mechanism seed-fixture
<path> --problem <id>` provisions a run + synthetic source snapshot from a
project-authored fixture and seeds the mechanism records via the existing
normalization write path.

The **fingerprint** is `SHA-256` over a canonical JSON body:

```text
{schema_version, vocabulary_version,
 sorted resolved canonical-ID set per field,
 posture enums, outcome.class,
 sorted resolved boundary relations}
```

Sets are sorted and JSON keys are fixed, so the fingerprint is
**order-independent** and **version-visible** (changing schema or vocabulary
version changes the fingerprint). Provenance (support, confidence, surface
wording, non-`resolved` claims) is **excluded**, so changing only provenance
never changes the fingerprint.

### Fingerprint equality is resolved-identity equality only

Two signatures that share resolved sets / posture / outcome / boundaries but
differ in `ambiguous`/`unknown`/`novel_candidate` claims will hash **equal**.
This is a deliberate, documented boundary — it is *not* a claim of full
mechanistic sameness. Before any identity conclusion, `compare` marks a field
`incomparable` if either side has a non-`resolved` claim, so a consumer cannot
over-trust fingerprint equality.

Fingerprint identity is also **conditional on a fixed classifier contract**:
because which labels reach `resolved` depends on the upstream classified label,
signatures produced under different classifier contracts are not
identity-comparable.

Signatures are immutable and version-keyed by `(mechanism_id, schema_version,
vocabulary_version)`. Re-canonicalizing under a newer vocabulary creates a *new*
row; the old one stays reproducible. Re-running under the same versions is
idempotent (`existing`).

## Comparison

`newf mechanism compare <a> <b>` is **component-wise** and deterministic. Per set
field it reports overlap count, union count, Jaccard, and an ordinal
(`identical`/`high`/`low`/`none`/`incomparable`) over resolved canonical IDs;
posture axes report enum equality; boundaries compare over the same
`canonicalID|relation` composite the fingerprint uses, so the **typed relation
is decisive**: `stops_at(X)` and `requires(X)` never compare identical even
though they share a canonical ID. There is **no single scalar**.

Weights are explicit and versioned (`weights/v1`); an unknown weights version is
an error rather than a silent default swap.

### Comparison profiles: the comparator measures, the caller decides

`Compare` (and `CompareWithProfile`) **measure every axis** — all six set fields,
posture, outcome, and boundaries — regardless of which axes are treated as
decisive. A `ComparisonProfile` (versioned) then decides *which* measured axes
determine the mechanistic verdict for a given experiment or domain. There is no
universal hardcoded decisive-field list: mechanistic distance is not universal,
so clustering (#11) and invariant mining choose the profile rather than inherit
a hidden global definition of "mechanism". A non-decisive axis is still fully
measured and reported; it is simply not consulted for that profile's verdict, so
no measurement is ever hidden by axis selection. This is guarded by
`TestComparisonProfileIsCallerControlled`, which shows the same signatures
classifying differently under two profiles while their component measurements
stay identical.

### Default profile (`classify/v1` / `ProfileMechanismV1`)

The default profile's **decisive** fields are `preserves`, `operators`,
`assumptions`, `breaks`, and `auxiliary_objects` — the conserved/violated
properties, structural moves, and auxiliary constructions that predict outcome.
`representation` is the surface qualifier; `boundaries` and `posture` are
measured but non-decisive **by default** (a caller may promote posture/outcome
or add boundaries as decisive via a different profile).

### Production assessment profile (`classify/v3` / `ProfileMechanismV3`)

`classify/v3` (v2's absence/subset guards plus stable positive evidence:
nonempty-set agreement is decisive only when both sides justify
completeness) is the production **assessment** rule used by `experiment run`
and `experiment readiness`: the same decisive axis selection as v1 with
completeness-aware absence enabled — under its missing-data contract (see
[`experiment.md`](experiment.md)), an empty decisive field without a `complete`
justification is an epistemic gap (`incomparable`), never decisive negative
evidence. The flag participates in the profile hash, so `classify/v2` has a
distinct hash and a v2 verdict can never masquerade as (or be silently
substituted for) a pinned v1 verdict. Clustering (#11) and frontier ranking
deliberately remain on `classify/v1` pending a separate research decision:
switching them would change persisted mechanism-family identity.

### Profile hash: version strings are not enough for #11

A profile carries a human-authored `Version` (e.g. `classify/v1`), but a version
string alone is only reproducible if nobody edits its meaning later.
`ComparisonProfile.Hash()` therefore returns a content hash over the profile's
decisive axis selection (sorted decisive set fields, surface field, and the
posture/outcome decisiveness flags) — deliberately **excluding** the mutable
`Version` string. Each comparison verdict carries this `profile_hash`, and
clustering (#11) persists it alongside a cluster run so a later edit to what
`classify/v1` means yields a *different* hash, making the drift tamper-evident
instead of silent. `TestComparisonProfileHashStableAndContentSensitive` guards
that the hash is order-independent over decisive fields yet changes when any
decisive axis is added or a decisiveness flag flips.

### Rejected terms are durable

The vocabulary's explicit rejected-terms list is **persisted** (table
`canonical_rejected_terms`, keyed by `(vocabulary_version, rejected_normalized)`,
immutable per version). Runtime resolution always rehydrates the vocabulary from
SQLite, so an in-memory-only rejected set would silently vanish after
persistence and a rejected label would resolve as `unknown`/`novel_candidate`
instead of `rejected`. Seeding threads the rejected keys through
`VocabularySeedInput.Rejected`, and reload rebuilds them via
`BuildVocabularyWithRejected`; `TestRejectedTermSurvivesPersistenceRoundTrip`
asserts a rejected label still resolves to `rejected` after a full
seed → persist → reload cycle.

Rationale: a different *representation* of the same structural move is a surface
change. But a **break** (an invariant the approach deliberately violates) and an
**auxiliary object** (a lattice, a convex body, an averaging kernel newly
introduced) are themselves mechanism changes, not rewordings — the project
thesis is that a representation change or auxiliary object *can be the mechanism
break itself*. The Erdős–Straus → affine-lattice family is exactly such a case:
it shares operators/assumptions with congruence-local families but is productive
precisely because it **breaks** residue-class locality and introduces an affine
lattice. Treating `breaks`/`auxiliary_objects` as non-decisive would let the
comparator rate that genuine mechanism break as `mechanism-near`, erasing the
variable that determines outcome — the abstraction-safety failure this system
exists to prevent. This contract is guarded by
`TestBreakIsDecisive_ESAffineLattice`.

| Verdict | Condition |
|---|---|
| `mechanism-near` | Every decisive field is `identical`/`high`; none `low`/`none`/`incomparable`. |
| `mechanism-distinct` | Any decisive field is `low`/`none` (so differing only in what an approach *breaks* or in its auxiliary construction is enough to be distinct). |
| `unknown` | Otherwise (e.g. all decisive fields incomparable). |

The verdict is qualified by the surface axis into
`surface-distinct+mechanism-near` / `surface-near+mechanism-distinct` when the
representation field disagrees/agrees. Surface text similarity is auxiliary-only
diagnostic data and **never** drives the mechanistic decision.

```bash
newf mechanism compare <mechanism-a> <mechanism-b> --json
newf mechanism compare <mechanism-a> <mechanism-b> --no-write   # do not persist
```

## Abstraction safety

Canonicalization is an abstraction boundary. Per [`abstraction-safety.md`](abstraction-safety.md),
an abstraction must not erase a distinction that separates known successes from
failures. The system-owned invariant `AssertDiscriminationPreserved` enforces
this: for every pair of signatures whose `outcome.class` differs, it recomputes
an **outcome-excluded** fingerprint; if those are equal, the vocabulary has
collapsed every non-outcome distinction between two mechanisms known to differ in
outcome — a discrimination-loss defect. This is a code-owned check, not a
hand-verified equality, so a lossy vocabulary introduced later is still caught.

## Determinism and offline operation

The whole path runs offline against project-authored deterministic fixtures
(`testdata/fixtures/mechanism/`). No network or model access is used by any code
path exercised by tests or the CLI commands above.

> Note on evaluation: the six fixtures test contract-consistency (resolution,
> fingerprint, comparison, abstraction-loss detection) — **not** discrimination
> power on unseen mechanisms. The historical-holdout evaluation that `AGENTS.md`
> names as the primary benchmark is a later slice.
