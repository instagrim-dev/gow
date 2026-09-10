# v0 implementation plan (vertical slices)

Each slice must ship a usable CLI behavior, persisted artifacts, and tests.

## Slice 1 — initialize problem + SQLite store

- Commands: `newf init`
- Deliver:
  - migration runner
  - `problem`, `run`, `run_event` tables
  - stable ID generation
- Tests:
  - DB bootstrap/migrations
  - init idempotency with duplicate slug
  - `--json` output contract

## Slice 2 — ingest source metadata + immutable evidence

- Commands: `newf ingest`
- Deliver:
  - `source`, `evidence_record` immutable writes
  - dedupe by canonical ref + content hash
  - explicit ingest failure artifacts
- Tests:
  - duplicate ingest behavior
  - immutability guardrails (no update path)
  - malformed source handling

## Slice 3 — normalize approaches into typed mechanisms

- Commands: `newf normalize`
- Deliver:
  - normalization revision model
  - `approach`, `mechanism`, `outcome`, `failure_boundary`
  - mechanism-axis vocabulary versioning
- Tests:
  - schema validation of normalized outputs
  - re-normalization creates new revision, preserves old
  - provenance links to evidence and provider calls

## Slice 4 — cluster and inspect mechanism families

- Commands: `newf cluster`
- Deliver:
  - cluster revision tables + memberships
  - mechanistic diversity summaries
- Tests:
  - deterministic clustering fixture
  - cluster membership persistence
  - revision lineage for reclustering

## Slice 5 — infer candidate invariants

- Commands: `newf invariants`
- Deliver:
  - candidate invariant creation (`proposed`)
  - support-cluster links
- Tests:
  - minimum support-cluster enforcement
  - invariant JSON/human output formats

## Slice 6 — challenge and falsify invariants

- Commands: `newf challenge <invariant-id>`
- Deliver:
  - challenge record types
  - lifecycle transitions: surviving/weakened/split/merged/falsified/established
  - lineage edges for split/merge
- Tests:
  - valid/invalid transitions
  - `established` requires independent evidence class
  - challenge provenance retention

## Slice 7 — generate frontier proposals

- Commands: `newf generate --against <id> --count <n>`
- Deliver:
  - proposal schema with novelty + falsification path + info gain + cost ordinals
  - invariant targeting + nearest cluster linkage
- Tests:
  - proposals must target surviving invariants only
  - duplicate structural proposal handling

## Slice 8 — evaluate and record outcomes

- Commands: `newf evaluate`
- Deliver:
  - proposal and holdout evaluation modes
  - metric persistence + baseline comparison
- Tests:
  - metric computation fixtures
  - baseline parity checks (same budget envelope)
  - confidence calibration bucket output

## Slice 9 — compress success structure + end-to-end holdout

- Commands: `newf compress`
- Deliver:
  - success invariant revisions
  - link success invariants to boundary crossings and prior failure invariants
  - one scripted historical holdout run in repo fixtures
- Tests:
  - success invariant derivation from partial-success set
  - end-to-end historical holdout regression fixture

## Cross-slice quality gates

- Every command supports stable `--json`.
- IDs emitted are consumable by later commands.
- No generated interpretation is written as source evidence.
- Long-running command status inspectable through persisted run events.
- Failures remain retained as useful artifacts.
