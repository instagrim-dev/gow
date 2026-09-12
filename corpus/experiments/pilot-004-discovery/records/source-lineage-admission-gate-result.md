# (source-lineage-admission-gate) Result — `newf source lineage-diff` diagnostic implemented; admission rules preserved; reproduces the (source-lineage-check) evidence via a single CLI invocation

**Status**: implemented and merged into the local tree 2026-09-12.
No admission rules changed. No source-admission behavior changed.
A new read-only diagnostic surface added.

## Bottom line

The `(source-lineage-check)` finding — that four ES databases share
12/13 source SHA-256 values byte-for-byte and are therefore not four
independent research replications — is now reachable via a single
CLI invocation that any operator can run before sealing a fresh
corpus. The gate is **descriptive**, not admission-blocking: it
reports overlap and emits a categorical verdict; the operator
decides what to do with the finding.

## Command

```
newf [--json] --db <left-db> source lineage-diff \
  --problem <left-problem-id> \
  --against-problem <right-problem-id> \
  [--against-db <right-db>] \
  [--shared-sample <int>]
```

Both sides are problem-scoped by construction: operators must name
a specific problem on each side. The diff cannot silently compare a
specific corpus against an unrelated aggregate.

## Verdict labels

Six possible verdicts, exhaustive over `(leftCount, rightCount,
sharedCount)`:

- `empty`: at least one side has no snapshots.
- `disjoint`: both sides non-empty; zero shared SHA-256 values.
- `identical`: both sides non-empty; SHA-256 sets are equal.
- `subset_left`: left is a non-empty proper subset of right.
- `subset_right`: right is a non-empty proper subset of left.
- `partial_overlap`: some but not all shared; neither side a subset.

None of these gates admission or writes any store row. The verdict
is a compact summary of the read-only overlap; operators consult it
before sealing a fresh corpus.

## Live cross-check against the (source-lineage-check) finding

Three live invocations against the existing pilot databases
reproduce the source-lineage-check evidence:

**M7 train (12 sources) vs pilot-001 train (12 sources), cross-DB:**

```
Source lineage diff (verdict: identical)
  Left:  problem=prb_01M2B8ZMHRCJN4YVH4CP6SNXTD  store=.newf/m7/newf.db     snapshots=12
  Right: problem=prb_01M27GTZHJN2V42W73DFVBY2SK  store=.newf/pilot-001/newf.db  snapshots=12
  shared:     12
  left-only:  0
  right-only: 0
```

Reproduces the (source-lineage-check) result: pilot-001 is
byte-identical to M7 at the source layer.

**M7 train vs pvnp train, cross-DB:**

```
Source lineage diff (verdict: disjoint)
  Left:  problem=prb_01M2B8ZMHRCJN4YVH4CP6SNXTD  store=.newf/m7/newf.db    snapshots=12
  Right: problem=prb_01M2BKVV5A7HD9HSAQKG7XEDST  store=.newf/pvnp/newf.db  snapshots=12
  shared:     0
  left-only:  12
  right-only: 12
```

Reproduces the (source-lineage-check) result: ES and P-vs-NP source
corpora are byte-disjoint.

**M7 train vs M7 target within-DB:**

```
Source lineage diff (verdict: disjoint)
  Left:  problem=prb_01M2B8ZMHRCJN4YVH4CP6SNXTD  snapshots=12
  Right: problem=prb_01M2B8ZMS5VYXY44JJWXR9MYJV  snapshots=1
  shared:     0
```

M7's 13 sources split cleanly into 12 train + 1 target (the
attested affine-lattice file), consistent with the M7 corpus
construction.

## Implementation surface

New code paths added (all read-only, none touching admission):

- `internal/store/store.go`:
  - `SnapshotLineageEntry` (SHA-256 + logical-name projection).
  - `ListSnapshotLineageForProblem(ctx, problemID) ([]entry, error)`.
- `internal/pipeline/output.go`:
  - `SourceLineageDiffSide`, `SourceLineageOverlapEntry`,
    `SourceLineageDiffResponse`, `SourceLineageDiffInput`.
- `internal/pipeline/source.go`:
  - `SourceLineageDiff(ctx, input) (response, error)` — orchestrator,
    handles same-DB and cross-DB comparisons.
  - `buildLineageDiffResponse` — pure set-arithmetic + sample
    construction, factored for testability.
  - `lineageVerdict(leftCount, rightCount, sharedCount) string`.
- `internal/pipeline/app.go`:
  - `ListSnapshotLineageForProblem` added to `problemStore`
    interface.
- `internal/pipeline/app_test.go`:
  - Fake-store stub for the new method.
- `cmd/newf/source.go`:
  - `newf source lineage-diff` subcommand with human and JSON
    output modes.

Tests added:

- `internal/store/store_test.go:TestListSnapshotLineageForProblem`
  — SQL-level projection: three cases including same-bytes-
  different-name discipline (an operator relabeling a source
  keeps both names surfaced in the overlap sample).
- `internal/pipeline/source_lineage_diff_test.go`:
  - `TestSourceLineageDiff_VerdictMatrix` — 8 sub-cases covering
    every verdict label with predeclared inputs and outputs.
  - `TestSourceLineageDiff_SampleLimit` — verifies
    `shared_sample` honors the sample limit AND preserves left/
    right logical names verbatim.
  - `TestLineageVerdict_Enumeration` — pins the state-space
    enumeration for the verdict function.

Verification: `gofmt -l .` empty, `go vet ./...` clean, `go build
./...` clean, `go test ./...` all-green.

## Admission-rules preservation (explicit)

- **No changes** to `CreateSourceSnapshot`, `SnapshotAdmission`,
  `validateSnapshotAdmission`, `createSourceSnapshotTx`,
  `getOrCreateSourceTx`, or any other admission path.
- **No new blocking** at ingest time. The diagnostic is entirely
  post-hoc: an operator runs it against already-ingested data.
- **No admission-side reads** consult lineage overlap.
- **No policy directives** emitted from the diagnostic.
- **No new epistemic promotion**. The diagnostic reports a
  descriptive verdict, not an evidential claim about research
  independence.

## Consequences for (l-coh) correction 5

The reviewer's defensible-replacement wording — "The four ES
instances do not provide four independent research replications.
Independence between the two corpus constructions has not been
established." — remains the correct claim to keep. The diagnostic
now lets any operator verify the ES-side clause with a single
command against any two ES-lineage databases.

The ES-vs-pvnp independence residual is unchanged: the diagnostic
confirms source-byte disjointness (necessary for research
independence) but does not establish research-methodological
independence (a stronger claim requiring separate evidence).

## Consequences for (l-extension)

If and when an operator authors a third truly-independent research
corpus, the freeze checklist can now include a single-command
lineage check:

```
newf --db <new-corpus.db> source lineage-diff \
  --problem <new-corpus-train-problem> \
  --against-db <es-corpus.db> \
  --against-problem <es-corpus-train-problem>
```

Expected verdict: `disjoint`. Any other verdict (`identical`,
`subset_*`, `partial_overlap`) is a source-lineage red flag the
operator must resolve before treating the new corpus as
independent.

## What this does NOT establish

- Does NOT establish research-methodological independence between
  disjoint corpora. `disjoint` at the source-byte layer is a
  necessary condition, not a sufficient one.
- Does NOT change any admission rule, verifier tier, or policy
  directive.
- Does NOT modify H3 admission-gate behavior.
- Does NOT persist a lineage record to the store. The diagnostic
  is stateless.
- Does NOT rank or score corpora against each other. The verdict
  is descriptive.
- Does NOT enable cross-cluster or cross-signature lineage
  inference. It operates strictly at the SHA-256 layer for source
  bytes.

## Consequences for the scorecard

- `(source-lineage-admission-gate)`: **closed**, implemented as
  `newf source lineage-diff` diagnostic. All tests green.
- `(source-lineage-check)`: unchanged — the SGO evidence remains
  the epistemic warrant; this loop makes it operator-reachable.
- `(l-coh)` correction 5 ES-side clause: unchanged. Now
  verifiable by any operator in one command.
- `(l-coh)` correction 5 ES-vs-pvnp clause: unchanged. Still open
  as a broader independence question.
- **H3 gate: preserved.**
- **Admission rules: preserved.**
