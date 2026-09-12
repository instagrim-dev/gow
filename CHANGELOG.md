# Changelog

All notable changes to `newf` are documented here. Format loosely follows
[Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## Versioning

This repo has two independently versioned surfaces, and `v1.0.0` versions
**only the first**:

- **The `newf` software/substrate** (CLI surface, domain types, storage
  schema, docs) — tagged `v1.0.0` here. A breaking change to a shipped CLI
  flag, JSON output shape, or storage schema migration warrants a major bump;
  additive commands/fields warrant a minor bump; docs/internal-only changes
  warrant a patch bump.
- **The `geometry-of-work` arXiv manuscript** (`paper/`) — versioned
  separately as "arXiv v1", "v2", etc., per `paper/PUBLICATION-PLAN.md` and
  `paper/HISTORY.md`. Freezing arXiv v1 is an external, human-reviewed event
  and is **not** gated on, or gating, this repo's software tags.

**Out of scope for `v1.0.0`:** M8 (cross-discipline portability litmus,
`EPIC.md` §M8) has no code or CLI surface yet and is explicitly not required
for v0/`v1.0.0` per `EPIC.md`'s own exit criteria; resolving the Erdős–Straus
conjecture is likewise explicitly not a `v1.0.0` requirement (`AGENTS.md`,
`EPIC.md`) — see `docs/plans/2026-09-11-012-plan-v1.0.0-gap-analysis.md`.

## [1.0.0] — 2026-09-11

First tagged release of the `newf` research substrate: a provenance-heavy
local CLI that implements the governing loop (failure-space → invariant →
challenge → frontier → evaluation → success compression → policy mutation)
against an Erdős–Straus conjecture corpus used as the running case study.

### Added — durable substrate and CLI surface (M0–M6.2)

- `newf init` / `ingest` / `problem` / `source` / `run` — provenance-first
  problem persistence and immutable, content-addressed source ingestion
  (`260c74e`, `770ce39`).
- `newf normalize` / `approach` — approach normalization into typed
  mechanism records with per-field explicit/inferred support tracking
  (`55dc654`, `ff6f870`).
- `newf mechanism signature` / `compare` / `seed-fixture` — deterministic
  mechanism canonicalization and profile-based comparison (`fefead0`,
  `1d5741f`, `42353c0`).
- `newf cluster` / `failure-space` — mechanism clustering and the
  `FailureSpace` artifact, with a discrimination-loss alarm on abstraction
  that erases outcome-determining distinctions (`d56cd00`, `151df63`).
- `newf invariants mine` / `invariant challenge`/`show`/`establish` —
  candidate failure-invariant mining and adversarial challenge/falsification,
  with the tier-1/independent-evidence `establish` path deliberately
  code-gated (`c048a69`, `008d3a8`).
- `newf frontier generate`/`list`/`show` — frontier proposal generation
  against surviving invariants (`008d3a8`, `a0591a4`).
- `newf evaluate` / `evaluation list`/`show`/`failures` — verifier routing
  with recorded verification strength, keeping `ModelJudgment != Verification`
  explicit in the schema (`d03ea53`, `40960ed`).
- `newf successes compress` / `success-invariant list`/`show` — symmetric
  partial-success compression into success invariants (`24cdc24`).
- `newf policy mutate`/`list`/`show` — explicit, versioned search-policy
  mutation that biases subsequent `frontier generate` runs (`4bfe614`,
  `d9fec05`).
- `newf mechanism/v2`–`v4`, `classify/v1`–`v3`, `recovery-rule/v1`,
  `proposal-wire/v1` — the vocabulary/classification/recovery/untrusted-import
  contracts hardened across the pilot campaigns (multiple commits,
  2026-09-10–11).

### Added — M7 blinded-benchmark harness (partial scope, as designed)

- `newf experiment define`/`run`/`readiness`/`show`/`validate-proposals`/
  `list`/`compare` — the BLINDED-BENCHMARK holdout harness: four arms
  (`b0_undirected`, `b1`–`b3`) executed offline through one shared frontier
  core, with equal persisted budgets and a code-owned recovery rule
  (`ec2ce51`, `bc10b35`, `9d9f528`).
- `mode=historical` execution is intentionally refused at the service
  boundary pending an operator-supplied dated-evidence artifact
  (`holdout_source_dating`) — a data gap, not a code gap (`EPIC.md` §M7).
  This is unchanged by `v1.0.0` and is not a release blocker.
- M7 executed end-to-end from `corpus/` in an isolated DB via a reproducible
  offline runbook (`corpus/experiments/m7-blinded-run/`), 2026-09-11.

### Added — Erdős–Straus corpus and pilot campaigns

- Pre-cutoff Erdős–Straus atlas and source-linked research collection
  (`8551dbe`, `1f711d0`).
- Pilot-001 (mapping/abstraction review, STOP disposition), Pilot-002
  (successor-protocol readiness), Pilot-003 (curated-feature blinded
  comparison, B0/B3 arms, `classify/v1`, independent review), and Pilot-004
  (discovery protocol, three-lane quorum adjudication over 23 proposal
  occurrences — 7 matches-reference, 10 defensible-novel, 2 over-merge,
  4 unsupported, 3 disputed) — frozen under `corpus/experiments/`.
- N2a and N5 novel-shape challenge lifecycles (`density_averaging_ceiling`,
  `reorganisation_without_qr_existence`) run against Pilot-004 discovery
  output, each with attributed reassessments narrowing claim strength after
  external review (`be9d3df`, `9af3376`, and the 2026-09-11 source-fidelity
  remediation below).

### Added — flagship arXiv manuscript (`paper/`)

- `paper/geometry-of-work.tex` scaffolded, all ten `PUBLICATION-PLAN.md`
  stages progressed, five figures (TikZ), full bibliography, and author
  metadata (`fe5f412` → `4f3fa56`).
- 2026-09-11 source-faithfulness review
  (`docs/reviews/2026-09-11-manuscript-source-faithfulness.md`) identified
  five submission blockers (methods/results misstating frozen pilot-003/004
  records, unsupported challenge claims, implemented-vs-proposed conflation,
  document-consistency defects, missing AI-use disclosure); all five
  remediated same day (`23fcc64`), with a follow-on residue sweep
  correcting the AI-disclosure content itself plus two missing citations
  (`39ba88d`), and a further QC correction pass narrowing endpoint reporting,
  claim strength, and literature contrasts, and adding CI enforcement
  against invoked draft macros/placeholders/rendered `??` refs (`1780933`).
- **External re-review of the manuscript is tracked separately in
  `paper/PUBLICATION-PLAN.md` (stage 10) and is explicitly decoupled from
  this software tag** — see
  `docs/plans/2026-09-11-012-plan-v1.0.0-gap-analysis.md`.

### Added — CI and release engineering

- `.github/workflows/ci.yml` — `gofmt`, `go vet`, `go build`, `go test`, and
  M7 reproducibility gates (`87fbf97`).
- `.github/workflows/paper.yml` — LaTeX compile gate, now also failing on
  invoked draft macros, placeholder markers, and rendered `??`
  cross-references (`87fbf97`, `1780933`).
- `newf version` / `newf --version` — build-time version reporting via
  `-ldflags "-X main.version=..."`, recorded as `tool_version` on every
  persisted provenance `Run` (this release).

### Deferred (recorded, not silently dropped)

- Dedicated `.github/workflows/release.yml` release-automation workflow and a
  `scripts/release-readiness.sh` preflight script — the manual checklist in
  `docs/plans/2026-09-11-012-plan-v1.0.0-gap-analysis.md` §3 is the gate for
  this tag; automating it is deferred to the next tag under normal
  "add abstractions after repeated pressure" discipline (`AGENTS.md`).
- A live-provider (real HTTP) integration test for `internal/provider`
  against a local mock server — CI remains deterministic-fixture-only;
  pilot campaigns captured real LLM calls out-of-band via the untrusted
  `proposal-wire/v1` boundary instead of the in-process provider path.
  Deferred pending an operator decision on whether this needs its own
  build-tag-gated CI job (open question OQ-3 in the gap-analysis plan).
- A `docs/experiment.md` addendum spelling out the exact dated-evidence bar
  that would lift the `mode=historical` gate — deferred; no dated corpus
  exists yet to motivate writing the bar precisely.

[1.0.0]: https://github.com/instagrim-dev/gow/releases/tag/v1.0.0
