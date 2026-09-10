# Mechanism clustering and the failure space

This document is the contract for `newf`'s deterministic mechanism-clustering
layer (issue #11). It builds directly on the canonical mechanism signatures and
component-wise comparator produced by the canonicalization layer (issue #9; see
[`mechanism-canonicalization.md`](mechanism-canonicalization.md)).

The governing principle is:

> Grouping is a reproducible function of signatures and an explicit profile —
> never an opaque embedding.

`newf` must never persist a model's "these belong together" answer as truth.
Clustering here is the connected-components of a linkage graph whose edges are
exactly `canon.CompareWithProfile` verdicts. Every family, representative, and
count is replayable from the persisted signatures plus the recorded profile and
algorithm version.

## What clustering is (and is not)

Clustering partitions the mechanism signatures of one problem, under one
`(schema_version, vocabulary_version)`, into **mechanism families**. Two
signatures share a family iff the comparator classifies them as `mechanism-near`
or `surface-distinct+mechanism-near` **and** no profile-decisive field was
incomparable. Grouping is transitive (connected components).

Clustering is **not**:

- an embedding, similarity threshold over vectors, or learned metric;
- evidence promotion — a family is a *candidate* grouping, not an established
  invariant;
- outcome inference — outcome classes are read from signatures, never inferred.

## The comparison profile decides; the comparator only measures

The comparator (`canon.CompareWithProfile`) measures every axis (all six set
fields, posture, outcome, and boundaries) regardless of which axes are decisive.
A caller-owned, versioned `canon.ComparisonProfile` names which axes are
**decisive** for identity. Clustering never inherits a hidden global definition
of "mechanism": the profile is an explicit input and is persisted on the run.

`ProfileMechanismV1` (the default) treats the operator / assumption / preserves /
breaks / auxiliary-object sets as decisive, uses representation as a *surface*
qualifier only, and treats posture and outcome as non-decisive. Swapping the
profile (e.g. making representation decisive) changes the family structure and
is recorded as a distinct run via `profile_version` + `thresholds_hash`.

## Determinism

`cluster.BuildClustering` is order-independent: signatures are sorted by
fingerprint before any pairwise work, so permuting the input yields
byte-identical clusters, representatives, and cluster fingerprints.

- **Representative**: the member with the lexicographically smallest signature
  fingerprint — a provenance-independent, stable choice.
- **Cluster fingerprint**: a `sha256` over the sorted member fingerprints plus
  the version tuple `(algo, profile, schema, vocabulary, thresholds)`.
- **Thresholds hash**: a `sha256` over the algorithm identity and the profile's
  decisive-axis selection, so a run under a different decisive set is a distinct,
  non-colliding run.

## Isolates and incomparability

A pair that is incomparable on any profile-decisive field (a non-resolved claim
on either side) never links — agreement cannot be asserted, so a link would be a
false merge. A singleton whose every candidate pair was incomparable is surfaced
as an **isolate**, never silently merged into a neighbor.

## Redundancy

Two members that are `surface-distinct+mechanism-near` add no new mechanism to a
family; the extra one is a redundant failure sample. Redundant members are
flagged (`redundant = 1`) and counted at the population level so the corpus's
true mechanism diversity is visible.

## Coverage and under-sampled axes

Each run reports, per axis, the number of distinct canonical values present
across families, and flags axes below the versioned `minAxisCoverage` threshold
as under-sampled. This is the signal that the failure corpus barely exercises an
axis, which biases any invariant mined from it.

## Discrimination-loss guard (population level)

Clustering reuses the system-owned `canon.AssertDiscriminationPreserved`
invariant across the whole population. If a lossy vocabulary erased every
non-outcome distinction between two mechanisms that have *different* outcome
classes, the run is marked `degraded` (still persisted, explicitly flagged with
the offending pairs) rather than presenting a clean family count over a
vocabulary that cannot tell successes from failures.

## The failure space

A `FailureSpace` is the first explicit, versioned failure-space artifact,
materialized from a cluster run. It:

- partitions families by the outcome class carried on their representative
  signature (never re-inferred);
- preserves the cluster's coverage / under-sampled report;
- carries `distinct_family_count` and `redundant_member_count`.

Failure spaces are immutable and revisioned per problem: a new cluster run yields
the next revision, never a rewrite. Materializing an already-materialized cluster
run returns the existing revision.

## Persistence

See [`persistence.md`](persistence.md) for the schema. All clustering and
failure-space tables are immutable by trigger. Runs are idempotent on their full
version tuple:

- `cluster_runs` is unique on
  `(problem, schema, vocabulary, profile, algo, thresholds)`;
- `failure_spaces` is unique on `(problem, cluster_run)` and `(problem, revision)`.

## CLI

```text
newf cluster build --problem <id> [--profile mechanism/v1] [--vocab-version ...] [--schema-version ...]
newf cluster show <cluster-run-id>
newf cluster list --problem <id>

newf failure-space build --problem <id> [--cluster-run <id>]
newf failure-space show   [--problem <id> | --id <failure-space-id>]
newf failure-space coverage [--problem <id> | --id <failure-space-id>]
```

`cluster build` requires that signatures already exist for the problem under the
chosen version tuple (run `newf mechanism signature` first). All commands support
`--json` for machine-readable output with stable contracts.
