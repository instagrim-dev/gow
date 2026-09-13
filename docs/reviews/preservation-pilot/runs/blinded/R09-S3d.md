# Run 02 — assessment identity over evaluation results

Reviewed revision: `3443a1094407364be2e02264ffa1ee1bb0f3efde`

## 1. Decision, policy and scope

**Decision under review:** can a reader of the system's output surfaces distinguish an *initial* verdict from the *current* verdict for a frontier proposal, and does every surface that presents a proposal result say which one it presents?

**Decision: WITHHOLD.** At least one demonstrated blocking nonconformance applies: the primary proposal-result read surface (`frontier show`/`generate` JSON) presents the first-ever verdict under the unqualified key `result`, with no marker of initiality, no evaluation identity, and no assessed-content identity, while later evaluations with different verdicts exist in the same store (executed reproduction, F1 below). Two further surfaces (evaluation views, the evaluated-failures marker) also present outcomes without their assessment context (F2, F3).

Scope: the extracted source tree at the revision above, centered on the evaluation write path (`internal/store/evaluation_store.go`), the `frontier_proposals.result` column and its immutability trigger (`internal/store/migrations.go`), and every reader that presents a proposal outcome: `internal/pipeline/frontier.go`, `internal/pipeline/evaluation.go`, `internal/pipeline/output.go`, `internal/pipeline/successes.go`, `internal/store/success_store.go`, `internal/store/experiment_store.go`, and the CLI renderers in `cmd/newf/frontier.go` and `cmd/newf/evaluation.go`. Reproductions ran on isolated SQLite stores via `t.TempDir()`. Remaining uncertainties are listed in section 3 (open questions) and section 4 (not examined).

## 2. Responsibility and critical path

The causal path from a verdict to a reader:

1. `Evaluate` (`internal/pipeline/evaluation.go:61`) builds evaluation rows carrying the assessed `SignatureContentHash` and recomputed per-target verdicts.
2. `PersistEvaluationRun` (`internal/store/evaluation_store.go:120`) appends evaluation rows (append-only, re-evaluation is a new run) and, at line 192, sets `frontier_proposals.result` with `UPDATE … WHERE id = ? AND result IS NULL` — **first-write-wins**. The schema trigger (`internal/store/migrations.go:1985`) permits exactly one NULL→verdict set and aborts overwrites; the `WHERE result IS NULL` guard means a second evaluation's differing verdict is silently a 0-row update, not an error.
3. Readers then diverge:
   - `frontier show`/`generate` JSON presents `result` from that sticky column (`internal/pipeline/frontier.go:576-578` → `internal/pipeline/output.go:710`).
   - `evaluation show`/`list` presents per-evaluation verdicts (`internal/pipeline/evaluation.go:515-558` → `internal/pipeline/output.go:783-795`).
   - `evaluation failures` presents the `evaluated_failures` marker (`internal/store/evaluation_store.go:197-203`, `cmd/newf/evaluation.go:194`).
   - Success compression consumes evaluations via selection-policy/v3 (`internal/store/success_store.go:86-178`), *not* the sticky column.

Responsibility split: the store layer deliberately makes `result` a one-time initial verdict; the presentation layer is responsible for saying so and does not. The critical defect is at the presentation boundary, aggravated by contradictory in-code contracts about what `result` means (F4).

## 3. Findings, limitations, protections, questions

### F1 — sticky first verdict presented as the unqualified proposal `result`

- **Location:** `internal/pipeline/output.go:710` (field), `internal/pipeline/frontier.go:576-578` (population), write path `internal/store/evaluation_store.go:192`.
- **Triggering conditions:** any proposal evaluated two or more times where verdicts differ (explicit re-evaluation is a supported flow: `EvaluateInput.ProposalID`/`GenerationID`, `internal/pipeline/evaluation.go:20-34`).
- **Violated contract:** re-evaluation must append and must never be silently absorbed into an earlier verdict's presentation; a verdict without its assessment context is not a current result. The view exposes no evaluation id, assessed content hash, timestamp, or initial/current qualifier — only `"result": "success"`.
- **Downstream consequence:** in the success-then-failure ordering, `frontier show --json` reports `result=success` while the current (latest, equally strong) assessment is `failure`; any operator or tool consuming that JSON adopts a superseded verdict as the proposal's outcome. In failure-then-success the stale `failure` persists on the surface.
- **Discriminating regression check:** on one proposal persist success-then-failure (and the reverse); assert the proposal-outcome surface either (a) resolves the current verdict under an explicit selection policy, or (b) labels the sticky value as initial and carries its evaluation id. Present rendering fails both.
- **Evidence class:** executed reproduction (`TestReview_SuccessThenFailure_SurfacesDiverge`, `TestReview_FailureThenSuccess_StickyResultAndMarkerStale`, disposable file `internal/store/zz_review_identity_test.go` in the extracted tree; both PASS demonstrating the divergence).

### F2 — evaluation read surface drops the assessed-content identity the store retains

- **Location:** `internal/pipeline/output.go:783-795` (`EvaluationView` has no `signature_content_hash` and no target-verdict fields); `internal/pipeline/evaluation.go:527-539` (`evaluationRunView` copies neither `e.SignatureContentHash` nor `e.TargetVerdicts`); human renderer `cmd/newf/evaluation.go:163-168` likewise.
- **Triggering conditions:** any proposal with multiple signature revisions (re-emission/dedup flow, v24 occurrence bindings). Two evaluations of *different* content revisions render identically except verdict.
- **Violated contract:** an assessment is identified by what it assessed — content revision, occurrence, context. The store persists exactly this (`evaluations.signature_content_hash`, `evaluation_target_verdicts`, written at `internal/store/evaluation_store.go:160-176`) and its own comments insist the hash is stored verbatim so the result is never stamped with bytes the verifier never saw; the presentation layer then discards it.
- **Downstream consequence:** a reader of `evaluation show`/`list` cannot tell whether a verdict assessed the proposal's current interpretation or a stale one, so "current verdict" is not reconstructible from the exposed surfaces at all — only by raw SQL.
- **Discriminating regression check:** evaluate revision A, re-emit content B, re-evaluate; assert the two evaluations' rendered views differ in an assessed-content field. Currently they differ only in id/verdict.
- **Evidence class:** demonstrated static path (struct and mapping inspected; persistence of the dropped fields confirmed by existing passing store tests).

### F3 — `evaluated_failures` marker frozen at first failing write, presented unqualified

- **Location:** `internal/store/evaluation_store.go:197-203` (`INSERT OR IGNORE`), schema `internal/store/migrations.go:2164-2170` (`proposal_id TEXT PRIMARY KEY`), surface `cmd/newf/evaluation.go:194` and `ListEvaluatedFailures` (`internal/store/evaluation_store.go:381`).
- **Triggering conditions:** (a) a later failing evaluation with a different verdict (`failure` vs `partial_failure`) or newer evaluation id — the marker keeps the first `evaluation_id`/`verdict` forever; (b) success-then-failure — the sticky `result` says `success` while a failure marker is inserted, so two surfaces disagree about the same proposal with neither disclosing which evaluation it reflects; (c) failure-then-success — the marker legitimately remains as history but is rendered without any "superseded by" qualification.
- **Violated contract:** each surface presenting a result must say which assessment it presents; first-write-wins absorption into an earlier record's presentation.
- **Downstream consequence:** a later `cluster build` consulting the re-entry list (its stated purpose, R6/KTD-5) ingests a verdict/evaluation binding that may not be the proposal's current failure assessment; operators see contradictory `frontier show` and `evaluation failures` outputs.
- **Discriminating regression check:** persist failure then partial_failure; assert the failures surface reflects (or explicitly qualifies) the newer assessment. Currently frozen at the first.
- **Evidence class:** executed reproduction (`TestReview_FailureThenPartialFailure_MarkerFrozen`, `TestReview_SuccessThenFailure_SurfacesDiverge`; PASS).

### F4 — contradictory in-code contracts about what `result` means

- **Location:** `internal/store/evaluation_store.go:119` claims "the proposal `result` mirrors the latest verdict written here"; the code at line 192 implements first-write-wins; `internal/store/success_store.go:15` calls the same column "the sticky frontier_proposals.result" set by "the earliest evaluation" — which itself misstates the selection policy implemented directly below it (selection-policy/v3 ranks content compatibility, then decisiveness, then strength, ties to the *latest*, `internal/store/success_store.go:113-133`).
- **Triggering conditions:** any maintainer or reviewer deciding which surface presents which verdict from the documented contracts.
- **Violated contract:** the write path's documented behavior must match its implemented behavior; here the same column is documented as both latest-wins and first-wins in the same package.
- **Downstream consequence:** future readers of the persist path will reasonably build "current verdict" consumers on top of `result` believing it mirrors the latest evaluation.
- **Discriminating regression check:** none executable; the smallest remedy is correcting `evaluation_store.go:119` and the `BreakCohortRow` header to state first-write/initial semantics and the actual selection ordering.
- **Evidence class:** inspected-only (documentation vs executed behavior from F1's reproduction).

### Limitations (project-declared)

- Pre-v17 proposals have no persisted signature content; batch eligibility falls back to the artifact-level sticky result (`internal/pipeline/evaluation.go:158-159`) — declared as an attribution gap, recorded, not faked.
- Hash-less legacy evaluations on multi-revision proposals are declared `BindingUnknown` and excluded from support (`internal/store/success_store.go:30-37`).
- Holdout evaluation mode is explicitly refused pending M7 (`internal/pipeline/evaluation.go:62-64`).

### Observed protections

- Immutability trigger permits exactly one NULL→verdict set and aborts overwrites/deletes (`internal/store/migrations.go:1982-2010`; covered by `internal/store/frontier_store_test.go:176-188`).
- Evaluations are append-only; re-evaluation is a new run with full provenance (verifier kind, strength, provider invocation).
- Success-cohort admission ignores the sticky column: selection-policy/v3 selects one coherent evaluation, prefers current-content assessments, returns `ContentHash` + `LatestContentHash` + `BindingUnknown` so callers detect pending reassessment (`internal/store/success_store.go:39-162`); H1 splice protection has an executed test (`internal/store/success_store_test.go:130-192`).
- Batch evaluation eligibility is content-scoped via `HasEvaluationForContent` (`internal/store/evaluation_store.go:225`), so an earlier artifact-level result cannot hide a never-assessed occurrence.
- `EvaluationView` always pairs verdict with verifier kind and strength (`internal/pipeline/output.go:780-789`).
- v28 compression selections make current guidance follow the latest execution rather than `MAX(revision)` (`internal/store/success_store.go:470-492`).

### Open questions

- Is `result` *intended* as a permanent "initial verdict" field? The trigger and success-store comments suggest yes; the persist-path comment says the opposite (F4). If yes, the JSON key and its doc comment (`output.go:697-698`) still assert no initiality qualification.
- The human `frontier show` renderer prints no outcome column at all (`cmd/newf/frontier.go:126-148`) — deliberate omission or gap? It cannot mislead, but it also cannot inform.
- Should `evaluation failures` reflect verdict *upgrades* between the two failure classes, or is first-marker-wins intended atlas semantics? No comment addresses the differing-verdict case.

## 4. Checks, coverage and resources

Commands executed (in the extracted tree; exit statuses as reported):

| Command | Exit |
| --- | --- |
| `git archive <SHA> \| tar -x -C $TMP` + `rm -rf $TMP/docs/reviews` (setup) | 0 |
| `go build ./...` | 0 |
| `go test ./internal/store -run 'TestReview_' -v` (3 disposable probes, all PASS, demonstrating F1/F3 in both orderings) | 0 |
| `go test ./...` (full suite, all packages ok) | 0 |
| assorted `grep`/`sed` inspections of the files below | 0 |

Files consulted (relative to subject tree): `internal/store/evaluation_store.go`, `internal/store/evaluation_store_test.go`, `internal/store/frontier_store.go` (result-relevant sections), `internal/store/frontier_store_test.go` (trigger coverage), `internal/store/success_store.go`, `internal/store/success_store_test.go` (H1/splice sections), `internal/store/experiment_store.go` (proposal-content queries; confirmed they read content, not `result`), `internal/store/migrations.go` (frontier_proposals trigger, `evaluated_failures` schema, v24 backfill), `internal/pipeline/evaluation.go`, `internal/pipeline/frontier.go` (view mapping), `internal/pipeline/output.go` (view structs), `internal/pipeline/successes.go` (cohort consumption, skimmed), `cmd/newf/frontier.go`, `cmd/newf/evaluation.go`. One disposable test file was created: `internal/store/zz_review_identity_test.go` (in the temporary tree only).

Not examined: `internal/pipeline/experiment.go` and holdout scoring paths beyond confirming their store queries do not read `result`; policy/challenge/invariant stores; `corpus/`, `paper/`, `docs/` prose beyond grep for result-semantics claims; concurrency behavior of the write path; CLI ergonomics tests. No blockers: build and tests ran normally offline.

Wall-clock estimate: ~20 minutes.

## 5. Remediation handoffs

1. **Qualify or re-derive the proposal-outcome surface.** Either rename/augment `FrontierProposalView.Result` to expose `initial_result` plus the setting evaluation id and assessed content hash, or compute a `current_result` via an explicit selection policy (reusing selection-policy/v3 semantics). **Acceptance:** the F1 regression (both orderings) shows a surface that distinguishes initial from current, and no JSON key presents an unqualified superseded verdict; existing suite stays green.
2. **Expose assessment identity on evaluation views.** Add `signature_content_hash` (and ideally the per-target verdicts) to `EvaluationView` and both renderers. **Acceptance:** the F2 regression — two evaluations of different revisions render distinguishably; empty hash renders as an explicit attribution gap, not omitted silently.
3. **Reconcile the `result` and marker contracts.** Correct `evaluation_store.go:119` and the `BreakCohortRow` header to state first-write/initial semantics and the real selection ordering; decide and document `evaluated_failures` differing-verdict semantics (freeze-with-qualifier or update), and test the failure-then-partial_failure case. **Acceptance:** no in-tree comment claims `result` mirrors the latest verdict; the F3 frozen-marker case is either documented intended behavior surfaced with a qualifier or updates to the newer assessment, with a test pinning the choice.
