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

## [Unreleased]

### Added

- `internal/measure`: claim-aware measurement checker — a pure,
  deterministic verifier for property labels attached to recorded
  observations. Binds the exact claim (sentence, metric, population,
  budget range, ordering, stopping rule, observed-vs-probabilistic kind);
  validates applicability (per-instance trace-prefix extension, verdict
  retention, non-budget condition equality); computes exact transitions
  (big.Int cross-multiplication; weighted-average sign rule
  sign(N·s − S·m)); and emits an explanatory certificate that separates
  total achievement, cumulative yield, and marginal yield, with
  NOT-ASSESSED scope guards so a certificate answering one question cannot
  be consumed as an answer to another. Probabilistic/uncertainty claims
  are refused, not decided; equality at tested points never extends to
  untested budgets; zero denominators stay unresolved. Certificates adapt
  to `review.CheckRecord` evidence (`ToCheckRecord`; NOT_ASSESSED maps to
  `blocked`, never `completed`) — the checker grants itself no authority
  over policy. Regression suite covers the failure-only decrease,
  proportional-batch equality, ordering-change attribution refusal,
  mutated-history rejection, probabilistic refusal, zero-denominator
  handling, range scoping, monotonicity, and the motivating budget-sweep
  shape (0/20, 3/40, 3/57).

### Changed — 2026-09-13 external review remediations (`f7554cb` findings)

Five instrumentation/claim-scope corrections; all frozen counts, criteria,
and dispositions retained. Each is recorded with its own section in the
attributed correction note appended to
`docs/plans/2026-09-13-019-screen-execution-record.md` (§1–§4 for the
findings below and the corpus item, §5 for claim scope, §6 for bounded
work, §7 for the controller-identity consequence):

- **Probe accounting (finding 1):** `shape.SelectV1`'s pre-search
  strict-reduction probes are now metered on the `Decision`
  (`ProbeRuleApplications`, `ProbeCandidates`) and charged by the sealed
  runner into the arm's task ledger in one unit (candidate rewrites
  materialized: `rewrite.Result.Generated` + probe candidates); expansion
  counts stay in traces as the budget unit. `rewrite.ProbeStrictReduction`
  is the metered probe; `CanStrictlyReduce` remains as the boolean
  convenience. Unmeasured custody is now recorded as unmeasured
  (`screen.Execution.CustodyMeasured`) and reported UNKNOWN — never zero —
  with per-arm `CustodyKnown`/guarded `FullCost`.
- **Claim scope (finding 2):** the v1 demotion rationale now states what
  the probe checked — "no one-step strict NodeCount decrease from the
  current start" — never "can never reduce this task", and never a verdict
  on the history's truth. Ordering policy unchanged.
- **Input identity (finding 3):** `SelectV1` now returns an error and
  refuses a task expression disagreeing with its declared rendering or
  duplicate rule names with conflicting content; its input hash binds the
  actual task rendering and full rule content (`rewrite.Rule.Identity`),
  not names alone.
- **Net-vs-gross (finding 4):** screen condition (c) exports gross paired
  control wins/losses beside the net figure; run-2's "never lost"
  narration is retracted in the record (tied 7/12, one paired gain offset
  one paired loss; the frozen net criterion passed and is unchanged). The
  inaccurate "only the relevance gate" H1/HG attribution comment is
  corrected: relevance filtering AND failure-driven demotion both differ.
- **Bounded work (finding 5):** `finite` structural validation carries a
  traversal-work bound (shared-subexpression blowup refused as a resource
  refusal, not a semantic judgment); `rewrite.SearchBounded` adds
  generated-state and term-size ceilings plus cancellation, all recorded
  on the result (`StateBounded`, `TermSizeBounded`, `Cancelled`) with
  defaults far above every retained run. **Completed by the self-review
  entry below** — the first pass added the flags without a consumer, so
  resource truncation still reached the grid as a non-completion.
- **Corpus:** `inf-06`'s one-node target annotated structurally
  unreachable through its catalog (pack bytes retained);
  `sealedrun.CalibrateH0MinBudgets` separates "unreachable within cap"
  from "completes with zero expansions" and callers guard the empty
  median case.

### Changed — 2026-09-13 self-review of the remediation above

An adversarial pass over the remediation commit found six defects in it.
All are corrected here; every frozen count still reproduces (run 1
21/21/21, run 2 14/12/13, run 3 14/12/16, calibrated budget 2), and
correction sections §5–§7 were added to the execution record:

- **A resource stop was still a semantic non-completion.** The bound flags
  landed with no consumer: a search truncated by the state or term-size
  ceiling reached `screen.Evaluate` as `Completed: false`, indistinguishable
  from a searched-and-failed miss. `sealedrun` now aborts a run whose
  search was resource-truncated (`searchBlockedReason`), and
  `screen.Execution.MeasurementBlocked` / `BlockedReason` make a blocked
  cell refuse the batch instead of being scored. `BudgetExhausted` is
  excluded by design: the budget is the declared measurement parameter.
- **`shape-selector/1` gained a second procedure under one identifier.**
  The repairs changed a frozen probe parameter (snapshot hash moved) and
  made the selector refuse inputs it previously decided, which
  `internal/shape/shape.go`'s "any change is a new version" rule forbids.
  Bumped to **`shape-selector/2`** (`SelectV2`, `InputV2`,
  `RunDiagnosticV2`, `RunConfirmatoryV2`); `/1` stays frozen and owns
  record 019's run-3 figures. The confirmatory freeze anchor no longer
  hard-codes a commit hash that silently pointed at the superseded
  procedure; outcome labels now carry the controller version.
- **The term-size ceiling was charged after the work it prevented.**
  Successors were rendered and then measured, so refusal cost what it
  refused (a measured audit: 66s for one expansion, 16,384 oversize terms
  rendered and discarded). Size is now measured on the tree before
  rendering, and the bound moved to **admission**: `finite.MaxExprNodes`
  bounds expression tree size, so every consumer of a validated
  expression — search keys, identity hashes, oracle replay — inherits a
  bounded rendering. `rewrite`'s ceiling is defined as that constant.
- **A non-matching probe was free.** The runner charged
  `ProbeCandidates` only, so a rule probed against a task it never matches
  performed a full positional traversal at zero cost. `probeWork` now
  charges applications and candidates.
- **A vacuous identity subtest.** One subtest changed the task *and* its
  rendering, then asserted the hash moved — which passed on the unfixed
  code, since `Input` already carried `TaskStart`. Replaced with the
  invariant that actually holds (every accepted input satisfies
  `Render(Task) == TaskStart`, so the hashed `Input` pins the probed
  expression), and the redundant hashed `Task` field was dropped.
- **The correction note claimed coverage it did not have.** It carried
  four sections for six corrections; findings 2 and 5 were absent while
  the changelog asserted five. Sections §5 (claim scope), §6 (bounded
  work) and §7 (controller identity) added.

### Added — Lean 4 kernel as a deterministic verifier tier

`internal/lean` (commit `5cba8b7`): an asymmetric deterministic verifier —
kernel-checked acceptance is `deterministic` strength; rejection or
unavailability degrades, never blocks, the evaluation route.

### Added — local-model fingerprint split

`internal/provider/localfp` (commit `8f6b1e5`): semantic fingerprints
(what was asked) separated from execution fingerprints (how the local
runtime was configured), so replay identity survives runtime upgrades.

### Added — mechanism/v6 vocabulary + first executed historical holdout

- **`mechanism/v6`** (commit `aaca121`): verbatim per-label canonicalization
  of the pvnp-holdout corpus; strict superset of v5 with regression tests
  pinning the no-merging discipline (KI03 implication vs target conversion
  stay distinct IDs).
- **First `mode=historical` execution** (commits `cb62bf2`→`e3caac2`,
  `corpus/experiments/pvnp-holdout/`): preregistered freeze, blind
  chronology audit, `experiment date-source` for both withheld sources,
  blinded external B0/B3 captures, clean leakage audit. Mechanical
  conclusion `inconclusive` exactly as preregistered; secondary
  ModelJudgment endpoint: 11/11 adjudicator concordance, both arms' rank-0
  proposals recover the withheld move (null-at-n=1 on guidance value;
  binding limitations in `RESULT.md`). Discharges the 1.0.0 "Deferred"
  note on the historical dating gate: `docs/experiment.md` now documents
  the exact dated-evidence bar and a dated corpus exists.

### Fixed — provenance-bearing paths fail loudly

Commit `f95821f` (rode along with the preservation-pilot lane): four
swallowed errors (frontier signature marshal, experiment capture-file
hash, policy evidence reads, success-compression vocabulary resolution)
now fail their stage instead of silently degrading persisted provenance;
`RecordExperimentExecutions` and `RunLeakageCheck` now run in
transactions.

### Added — checkable attempt→output binding (v46, review attribution slice)

Remediates the C3 attribution finding of the 2026-09-12 C1–C8 review run: a
supplied tuple proves only that the tuple fails the identity, not that the
proposal's executed bounded attempt produced it.

- **`newf witness check --procedure <name> --param k=v`** — executes a
  registered deterministic bounded-attempt procedure (`equal-denominator`,
  `greedy`; versioned executor in `internal/witness`) whose output IS the
  checked tuple. Abstention is an input-level refusal (nothing persisted);
  `--procedure` and `--tuple` are mutually exclusive.
- **Migration `v46`** — `witness_attempt_bindings`: one immutable binding
  (procedure, executor version, canonical params, canonical tuple) per
  evaluation, written in the same transaction as the evaluation.
- **Admission recheck** — rule admission recomputes the recorded procedure
  over the recorded params and names the verified binding in the basis; a
  binding that fails or refuses recomputation degrades rule admission to
  withholding (operator attestation is the recorded escape). The
  supplied-tuple path keeps its explicitly weaker provenance, recorded in the
  evaluation notes. Scope: the binding attributes the tuple to an identified
  executed bounded attempt; attempt→proposal mechanism fidelity remains a
  separate, recorded judgment.
- **Regressions** — `TestIntegrationWitnessAttemptBinding` plus executor unit
  tests (determinism, canonical-params round trip, recheck mismatch, version
  refusal, abstention).

### Added — book lane and method contract (docs only)

Opens the authored-book lane (*The Shape of Trying*) **without** freezing GoW as
a generally validated prescriptive discipline, per the 2026-09-12 assessment of
`fdf7f5e`. No code, CLI, schema, or claim status changes.

- **`docs/book/method-contract.md`** (v0.1.0, `proposed`) — one compact contract
  fixing the intended practitioner and task class, the mapping trigger and
  spending limit, the representation-selection procedure, the action-selection
  rule, the evidence requirements, and seven stopping conditions. Every
  instruction carries an authority tier (`derived` / `heuristic` /
  `conditional-evidence`) plus its preconditions, discretion, and failure mode.
  Strengthening an instruction requires a derivation or a frozen artifact;
  weakening one requires nothing.
- **`docs/book/warrant-boundaries.md`** — the checked-result / additional-
  conclusion ledger (witness checked ≠ mechanism attributed; success ≠
  explanation; one improvement ≠ class exclusion; `ELIGIBLE_TO_ADVANCE` ≠ true),
  three laundering routes, Counterform's constrained treatment, and sentence
  tagging for chapter drafts.
- **`docs/book/README.md`, `docs/book/OUTLINE.md`** — charter with three
  authority tiers and gates B0–B5 (B5, freezing prescriptive chapters, is
  blocked on the usability trial; B3, drafting the core practice, is not), and a
  20-chapter outline where each chapter carries its tier and an explicit
  may-not-claim list.
- **`corpus/experiments/contract-usability-trial/`** (`draft_not_frozen`) —
  PROTOCOL plus `manifest.json` (`schema: newf-usability-trial/v1`). Tests
  whether the contract helps practitioners make **appropriate decisions,
  including the decision not to map**; a completed WorkMap is not the endpoint.
  Three-case battery with preregistered adverse observations: one suitable
  decision-relevant case, one inadequate-population control, one
  decision-irrelevant-uncertainty control. Two independent participants on all
  three cases (six participant-case observations, **not** six replications),
  with commitments sealed for all three cases **before** the instrument bundle is
  released, compared on recommended actions rather than map appearance. No global
  pass condition; the predeclared finding of primary interest is two participants
  following the instructions and recommending incompatible actions, which locates
  an underspecified prescription. Case selection is "predeclared rule without
  outcome-based substitution" — the word *uncurated* is withdrawn, since
  inclusion criteria are themselves selection. Two freezes: **F1 fixes the
  selection policy** for cases *and* for the instrument's warrant extracts
  (source-ledger revision, domain-to-row inclusion criteria, permitted editing,
  treatment of qualifications / conflicting warrants / missing support); **F2
  fixes what the policy produced** (packets, evaluator key, participant-visible
  extract text, digests). Methodological warrants belong to the bundle; objective-
  relevant domain facts belong to the case-information boundary and must be
  available before and after exposure, or the treatment is renamed *contract plus
  additional domain guidance*. A supplement may not complete N1's missing
  evidence. Missing or noncompliant extracts **block exposure**. Consent and
  out-of-repository handling of raw records are launch requirements. Conclusions
  are limited to observed usability and decision behavior — not superiority, not
  general reproducibility, not causal effectiveness.

### Changed — narrowed practitioner prescriptions

- **`docs/thesis/how-to-map-the-work.md`** §3 rewritten. The earlier text made
  one-factor-at-a-time variation, a numeric prediction, and "maximize what it
  teaches" the *defining* properties of a probe. One-factor-at-a-time cannot
  detect interactions (NIST/SEMATECH e-Handbook, Ch. 5), which is a poor mandate
  for a method whose central interest is relational structure. Replaced with a
  design rule (discriminate the relevant explanations; predeclare a checkable
  prediction; preserve the objective and its correctness conditions; justify the
  cost) plus an admissible-design catalog. §2 now also requires a second
  candidate description and its divergence case; §4 adds the
  no-retrospective-re-description rule; a new closing section names the three
  things the page deliberately does not settle.
- **`docs/thesis/when-to-map-the-work.md`** gains an allocation rule: map or
  probe when a named uncertainty would change the next action, under a declared
  spending limit with a stated stopping condition; otherwise act. Marked a
  declared default, not a result. "Stop generating and map" is now explicitly a
  bounded spend, and "next probes" is work that would change a decision rather
  than work that maximally clarifies the map.
- **`docs/theory/00-paper-claims.md`** adds `H-MC`: the contract's `heuristic`
  instructions are not claims of the registry, and the registry's scope over the
  book lane is stated (same claims, one extra constraint — prescriptions carry
  tiers).

### Fixed — per-evaluation admission visibility (v45, review finding F-1)

Remediates F-1 of the 2026-09-12 C1–C8 review run (reproduced admission
omission): the `evaluated_failures` re-entry marker was keyed per **proposal**
with `INSERT OR IGNORE`, so a proposal whose first failure was model-judged
(withheld) permanently shadowed a later witness-checked failure of the same
proposal from admission.

- **Migration `v45`** — rebuilds `evaluated_failures` with a per-**evaluation**
  primary key (rows and immutability triggers preserved; introspective and
  idempotent; fresh stores get the new shape from the baseline DDL). Every
  applicable evaluation now remains independently visible to admission under
  its exact context; earlier decisions are preserved, not replaced — visibility
  is not strength-ranked replacement.
- **No-inflation companion** — unchanged by design: admission materializes
  under the stable logical identity `frontier-proposal:<id>`, so a second
  admitted evaluation of the same proposal revises the same approach and the
  current-heads population does not grow a second member.
- **Regressions** — `TestIntegrationLaterStrongerEvaluationReachesAdmission`
  (the exact shadowing case, exactly-once admission, no-inflation) and
  `TestMigrateV45RebuildsPerProposalMarkerTable` (upgrade path on a real
  pre-v45 store).

### Added — normative review records (v44)

Remediates G1 of the 2026-09-12 review-flow run, whose disposition was
`UNDETERMINED` because the review contract's four record responsibilities had no
storage mapping: coverage could only have been hand-written.

- **`internal/review`** — pure decision projection and deterministic
  `COVERAGE.md` generator. Derives `WITHHOLD` / `UNDETERMINED` /
  `ELIGIBLE_TO_ADVANCE` from records alone, with reason codes. Emits no
  generation timestamp, so identical records render byte-identical output.
- **Migration `v44`** — `review_policies`, `review_obligations`,
  `review_policy_obligations`, `review_applicability_decisions`,
  `review_dependency_manifests` (+ dependencies), `review_check_attempts`,
  `review_assessments` (+ check references), with immutability triggers and a
  trigger refusing to link a blocked check attempt to a `conforms` assessment.
  These are normative records; they share nothing with the scientific lifecycle
  and have no promotion path into it.
- **`newf review applicability|check|coverage`** — record applicability
  decisions and check attempts, and generate coverage. There is deliberately no
  status writer: coverage is derived on every read.
- **`TestIntegrationCurrentAssessmentAuthorityObligation`** — cases C1–C8 of
  `docs/reviews/prompts/recipes/assessment-admission-decision.md` for the single
  obligation `current-assessment-authority@1`, through the real migrated store,
  plus a control policy proving `unexamined` and execution-`blocked` remain
  distinguishable from each other and from a pass.
- **`docs/normative-review-records.md`** — the mapping, the refusals it encodes,
  and the C4/C5 staleness boundary (relevance, not "HEAD moved").

Encoded refusals: an inspected procedure cannot be recorded as `completed`; a
blocked attempt cannot support conformance; absence of an assessment reports
`unexamined` rather than a pass; conflicting applicability stays unresolved
instead of being settled by recency; a demonstrated nonconformance is not erased
by a later favorable assessment; and vacuous or unauthorized policies cannot
grant eligibility.

### Changed — build requirements

- **Minimum Go toolchain is standardized at 1.25.** The floor was already
  declared by the `go` directive in `go.mod`; it is now stated for operators in
  `README.md`, made a hard constraint in `AGENTS.md`, and enforced by
  `toolchain_test.go` so it cannot regress or be shadowed by a competing pin.
- `toolchain_test.go` — repository-scoped gate asserting that the `go` (and any
  `toolchain`) directive meets the 1.25 floor, that every `actions/setup-go`
  step derives its version from `go.mod` rather than hardcoding one, and that
  any hosted-container pin in `.codex/environments/environment.toml` does not
  undercut the floor (that file is gitignored and per-operator, so the last
  check skips in CI). Motivated by a hosted runner defaulting to Go 1.23.2,
  which cannot build this module and previously failed with no pointer to the
  cause. Each assertion was mutation-checked: lowering the directive, adding a
  below-floor `toolchain` line, hardcoding `go-version:`, dropping the version
  source from one of two `setup-go` steps, redirecting `go-version-file` away
  from `go.mod`, and lowering the container pin all turn the gate red.
- `.gitignore` — ignore `/bin/`, the conventional `go build -o bin/` output
  directory (a binary built for one runner's `GOARCH` is not portable to
  another), and `.codex/`, autogenerated per-operator agent-host config that is
  not a repository contract.

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
