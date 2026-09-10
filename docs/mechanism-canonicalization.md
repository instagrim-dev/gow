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
| `rejected` | On the vocabulary's explicit rejected-terms list. |

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
`MechanismSignature` (from the #7 dotted-path support: `mechanism.locality`,
`mechanism.construction`, `mechanism.uncertainty`, `outcome.class`,
`outcome.boundary`) and are persisted alongside the signature, so invariant
mining reads their provenance rather than assuming source backing. Unprovenanced
posture/outcome round-trips as `unknown`
(`TestSignatureCarriesPostureOutcomeBoundaryProvenance`).

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
posture axes report enum equality; boundaries compare by canonical relation.
There is **no single scalar**.

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
