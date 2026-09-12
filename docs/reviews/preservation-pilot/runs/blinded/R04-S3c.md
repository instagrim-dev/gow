# Review run S3c-M — assessment identity over evaluation results

Reviewed revision: `5dd08618f10bde9ea0eca2aa8b2d39426898155c`

## 1. Decision, policy and scope

**Decision under review:** can a reader of this system's outputs distinguish an *initial* verdict from the *current* verdict for a frontier proposal, and does every surface that presents a proposal outcome say which one it presents?

**Policy basis:** the review obligations supplied for this run: (a) an assessment is identified by what it assessed (content revision, generation occurrence, problem, population context) — a verdict without its context is not a current result; (b) re-evaluation appends and must never be silently absorbed into an earlier verdict's presentation; write paths with first-write-wins / last-write-wins patterns must be traced to every read surface that presents the value.

**Scope:** the extracted source tree at the revision above, only. Mode: read-only inspection plus disposable tests on isolated SQLite (`t.TempDir()`). No providers, no network, no repository writes other than this report.

**Decision projection:** `WITHHOLD` for the claim "every result-presenting surface correctly qualifies which verdict it presents." Two demonstrated nonconformances apply (findings 1 and 3 below). The core write/read substrate itself (append-only evaluations, ledger-derived occurrence result) is sound and well-tested; the gaps are at specific presentation surfaces and at one ordering assumption. Remaining uncertainties are listed in section 3.

## 2. Responsibility and critical path

The consequential causal path for a proposal verdict:

1. **Write:** `Store.PersistEvaluationRun` (`internal/store/evaluation_store.go:120`) writes append-only `evaluations` rows carrying `verdict`, `verifier_kind`, `verification_strength`, and the exact `signature_content_hash` the verifier assessed, plus per-target verdicts (`evaluation_target_verdicts`, v26). In the same transaction it (a) sets `frontier_proposals.result` once, `NULL → verdict` only (`evaluation_store.go:192`, enforced by trigger `frontier_proposals_immutable_update`, `internal/store/migrations.go:1985–2005`), and (b) writes an `evaluated_failures` marker via `INSERT OR IGNORE` keyed on `proposal_id` alone (`evaluation_store.go:196–203`; PK at `migrations.go:2165`; immutability triggers at `migrations.go:2220–2228`).
2. **Derive:** `latestOccurrenceResult` (`internal/store/assessment_views.go:56`, SQL at `:40–54`) projects the *current* verdict per (generation occurrence, exact content hash, cluster context), ordered by `e.created_at DESC, e.rowid DESC`.
3. **Read:** `GetFrontierGeneration` and `ListOccurrenceProposalRows` (`internal/store/frontier_store.go:307,627`) overwrite the loaded `result` column with the derived occurrence result before returning; the immutable column value is documented as "the initial verdict for compatibility" (`frontier_store.go:47–50`).
4. **Present:** `frontier show/list` JSON (`FrontierProposalView.Result`, `internal/pipeline/output.go:710`), `evaluation show/list` (`EvaluationView`, `output.go:783–795`; human table `cmd/newf/evaluation.go:163–169`), and `evaluation failures` (`cmd/newf/evaluation.go:121–149,186–198`; pipeline `internal/pipeline/evaluation.go:476–501`).
5. **Consume:** batch evaluation eligibility skips proposals whose *occurrence-scoped* result is already set (`internal/pipeline/evaluation.go:141–153`); re-evaluation context resolution is explicit and occurrence-aware (`evaluation.go:76–108`).

Responsibility split: the store layer owns identity and immutability (strong); the presentation layer owns qualification of which verdict a surface shows (the weak link).

## 3. Findings, limitations, protections, questions

### Finding 1 — `evaluated_failures` presents a superseded verdict as current re-entry eligibility

- **Location:** `internal/store/evaluation_store.go:196–203` (write), `internal/store/migrations.go:2164–2170` (PK `proposal_id`), `migrations.go:2220–2228` (immutable), `cmd/newf/evaluation.go:123–127` (help copy: "made their mechanism eligible for inclusion in the next `cluster build`"), `cmd/newf/evaluation.go:188` (empty state: "no evaluated failures eligible for atlas re-entry"), `cmd/newf/evaluation.go:192–196` (table renders `VERDICT` unqualified).
- **Triggering conditions:** any re-evaluation of a proposal after its first failing evaluation. Two demonstrated shapes: (a) `failure → partial_failure`: the second marker insert is silently dropped (`INSERT OR IGNORE` on an existing PK), so the surface shows the first verdict and first evaluation ID; (b) `failure → success`: the marker survives and the proposal is still listed as an evaluated failure "eligible for atlas re-entry" while every occurrence read surface reports `success`.
- **Violated contract:** re-evaluation must never be silently absorbed into an earlier verdict's presentation; a surface presenting a result must say which one it presents. The pipeline's own comment concedes these are "historical failure markers, not current verdicts" (`internal/pipeline/evaluation.go:476–479`), but the user-facing copy and column naming assert current eligibility.
- **Downstream consequence:** an operator (or policy step) deciding what to feed the next `cluster build` reads a stale severity (`failure` vs `partial_failure` matters for failure-space structure) or re-enters a mechanism whose current assessment is `success`, contaminating the failure atlas input.
- **Discriminating regression check:** on an isolated store, persist two evaluation runs for one proposal/content (both orderings `failure→partial_failure` and `failure→success`); assert the failures surface either reflects the latest verdict or explicitly labels the marker as historical *and* the fixture's current verdict is reachable from the same output. (Executed as disposable probes 1–2; both divergences reproduced.)
- **Evidence class:** executed reproduction (store layer; CLI copy inspected).

### Finding 2 — "Latest" occurrence verdict is wall-clock-ordered, not ledger-ordered; a backdated re-evaluation is hidden

- **Location:** `internal/store/assessment_views.go:53` (`ORDER BY e.created_at DESC, e.rowid DESC`), contradicting the same file's stated contract "Latest means the last recorded assessment in THIS occurrence's context" (`assessment_views.go:34`). `created_at` is caller-supplied verbatim (`internal/store/evaluation_store.go:161–163`; pipeline stamps `a.now()` at `internal/pipeline/evaluation.go:209`) with no monotonicity validation at the write boundary.
- **Triggering conditions:** a re-evaluation run persisted with a `created_at` earlier than a prior run's — clock regression, cross-host skew, or programmatic backfill through the exported `PersistEvaluationRun`. The `rowid` tiebreak applies only on exact string equality.
- **Violated contract:** the newest recorded assessment is not presented as current; the earlier verdict silently retains the "current" presentation — a timestamp-mediated last-write-loses.
- **Downstream consequence:** every occurrence read surface (frontier views) and batch eligibility (`p.Result.Valid` skip, `internal/pipeline/evaluation.go:149–151`) act on the wrong current verdict; a corrected assessment can be permanently invisible without any error.
- **Discriminating regression check:** persist evaluation A (verdict `failure`, t=12:00) then evaluation B (verdict `success`, t=11:00, recorded later); assert the occurrence result presents the last-*recorded* verdict, or that the write of a non-monotonic `created_at` is rejected. (Executed as disposable probe 3; hidden-verdict behavior reproduced.)
- **Evidence class:** executed reproduction (constructed timestamps; no in-the-wild skew observed).

### Finding 3 — Presentation surfaces drop the assessment identity the store already persists

- **Location:** `internal/pipeline/output.go:783–795` (`EvaluationView` has no `signature_content_hash` field and no target verdicts), `internal/pipeline/evaluation.go:516–528` (mapping drops `SignatureContentHash` and `TargetVerdicts`), `cmd/newf/evaluation.go:163–169` (human table likewise); `internal/pipeline/output.go:710` (`FrontierProposalView.Result` is a bare string with no evaluation ID, verifier kind, strength, or "derived latest for this occurrence" qualification); stale comment `output.go:696–698` ("Result is empty until M5.2 evaluation populates it" — it is now a ledger-derived, re-evaluation-sensitive projection).
- **Triggering conditions:** a proposal re-emitted with revised content (cross-generation dedup + new occurrence binding) evaluated under both revisions: `evaluation show` renders two evaluations whose views are indistinguishable in content context; `frontier show --json` renders a `result` whose producing evaluation cannot be identified from the output. The human `frontier show` table shows no result column at all, so the initial/current distinction is unreachable there.
- **Violated contract:** an assessment is identified by *what* it assessed; a verdict without its context is not a current result. The store persists exactly the needed identity (`evaluations.signature_content_hash`, `evaluation_target_verdicts`), so this is presentation-layer loss, one root cause across two manifestations.
- **Downstream consequence:** audit or replay from machine-readable CLI output cannot bind a verdict to the interpretation it assessed; a reader can misattribute an old-revision verdict to the current interpretation — precisely the confusion the v26/F1 store work was built to prevent.
- **Discriminating regression check:** fixture with one proposal, two signature revisions, one evaluation each; assert `evaluation show --json` output distinguishes the two rows by assessed content revision, and `frontier show --json` names the evaluation (or content hash) behind `result`.
- **Evidence class:** demonstrated static path (structs and mapping inspected; underlying store fields confirmed populated by executed tests).

### Finding 4 — Write-path documentation contradicts the implemented first-write-wins column

- **Location:** `internal/store/evaluation_store.go:117–119` ("the proposal `result` mirrors the latest verdict written here") vs. `evaluation_store.go:192` (`UPDATE … WHERE … result IS NULL`) and trigger `migrations.go:1982–2005` (any overwrite of a non-null result aborts). The correct contract lives at `internal/store/frontier_store.go:47–50` (column = immutable initial verdict; readers derive latest).
- **Triggering conditions:** any future maintainer reading only the write path before building a new read surface on the raw column.
- **Violated contract:** the write path mislabels a first-write-wins column as last-write-wins — the exact ambiguity this review's obligations target.
- **Downstream consequence:** a new reader of `frontier_proposals.result` built on the comment's promise would present the initial verdict as current (the two agree only until the first re-evaluation).
- **Discriminating regression check:** none executable for prose; the behavioral half is already pinned by `TestOccurrenceResultReversalsPreserveInitialHistory` (`internal/store/assessment_views_test.go:100–133`). Remedy is a comment correction.
- **Evidence class:** inspected-only (the underlying column behavior itself was executed and confirmed initial-only).

### Limitations

- Store-layer probes only; the CLI binary was not driven end-to-end (surfaces assessed from renderers and view structs).
- Success-compression cohort admission and experiment surfaces were spot-checked (`internal/store/success_store.go:24,171`; `internal/store/experiment_store.go` reads content, not `result`), not fully reviewed; `success_store_test.go:192` suggests the cohort selects a coherent evaluation triple, untested here.
- Legacy pre-v17/pre-v24 fallbacks (empty content hash, owned-row fallback `internal/pipeline/evaluation.go:125–127`) assessed only via existing tests.
- Early in the session one recursive text search accidentally executed against a different checkout; its output was discarded unused and every cited line was re-verified against the subject tree by absolute path.

### Observed protections

- The occurrence-result projection scopes by generation, exact content hash, cluster context, and mode, and cannot cross generations or content revisions (`assessment_views.go:40–54`; `TestOccurrenceResultDoesNotCrossGenerationOrContent`, executed, pass).
- Both re-evaluation orderings (`success→failure`, `failure→success`) present the latest verdict on read surfaces while the immutable initial verdict is retained (`TestOccurrenceResultReversalsPreserveInitialHistory`, executed, pass).
- Comprehensive immutability triggers: evaluations, runs, metrics, target verdicts, occurrence bindings, and one-time `NULL→verdict` on proposals (`migrations.go:1982–2230, 3105–3117`).
- Evaluation stores the *supplied* assessed content hash verbatim, never a latest-revision lookup (`evaluation_store.go:79–84`); batch eligibility is content- and occurrence-scoped (`evaluation.go:129–153`); equal-clock ties resolve by ledger insertion order and are tested (`assessment_views_test.go:74–75`).
- The human evaluation table deliberately never shows a verdict without its verification strength (`cmd/newf/evaluation.go:163–165`).

### Open questions

- Is `evaluated_failures` intended as a one-shot historical event log (then its `verdict` column and "eligible for re-entry" copy overstate) or a current eligibility queue (then it must reflect re-evaluation)? The schema comment (`migrations.go:2161–2163`) and pipeline comment pull in opposite directions from the CLI copy.
- Should evaluation recency trust caller-supplied clocks at all, versus ledger insertion order or an explicit supersession link (the sibling signature projection already prefers explicit supersession over timestamps, `assessment_views.go:9–13`)?
- Should `FrontierProposalRow`/`View` carry result provenance (evaluation ID, verifier strength) so the frontier surface meets the same "no outcome without strength" bar as the evaluation surface?

## 4. Checks, coverage and resources

Commands executed (exit status):

- `git archive <SHA> | tar -x` into temp dir; removed excluded subtree — 0
- `go build ./...` — 0 (`BUILD_OK`)
- `go test ./internal/store -run 'TestReview|TestOccurrenceResult|TestLegacyResult|TestPersistEvaluationRun|TestEvaluationFailureReEntersAtlas' -v` — 0; existing identity tests pass; 3 disposable probes each reproduced the targeted behavior (`CONFIRMED` log lines)
- `go test ./...` — 0 (all 17 packages pass, including the disposable probes)
- `gofmt -l .` — 0, empty output
- assorted `grep`/`sed`/`wc` inspection commands — 0 (one recursive grep invalidated and discarded per Limitations)

Disposable artifact created inside the extracted tree only: `internal/store/review_disposable_test.go` (3 probes: marker-keeps-first-verdict, marker-survives-success, backdated-re-evaluation-hidden).

Files consulted (relative to subject tree): `internal/store/evaluation_store.go`, `internal/store/assessment_views.go`, `internal/store/assessment_views_test.go`, `internal/store/frontier_store.go`, `internal/store/frontier_store_test.go` (helpers), `internal/store/migrations.go` (proposal/evaluation/occurrence DDL and trigger sections), `internal/store/evaluation_store_test.go` (test inventory), `internal/store/success_store.go` (partial), `internal/store/success_store_test.go` (one assertion), `internal/store/experiment_store.go` (proposal-content queries), `internal/pipeline/evaluation.go`, `internal/pipeline/frontier.go` (view mapping), `internal/pipeline/output.go` (view structs), `internal/pipeline/app.go` (interface line), `cmd/newf/evaluation.go`, `cmd/newf/frontier.go`; directory listings of `cmd/newf`, `internal/store`, `internal/pipeline`.

Not examined: other CLI commands (ingest/normalize/cluster/invariant/challenge/policy/success/experiment surfaces beyond result reads), `internal/experiment`, provider adapters, `internal/canon`, holdout paths, `docs/`, `corpus/`, `paper/`, migration history outside the cited sections, and any CLI end-to-end invocation.

Wall-clock estimate: ~20 minutes.

## 5. Remediation handoffs

1. **Reconcile the `evaluated_failures` surface with its historical semantics.** Either reword the command help, empty-state, and view to name the marker as "first failing evaluation (historical)" and surface the current occurrence verdict alongside, or make the marker append-only per `(proposal_id, evaluation_id)`. *Acceptance:* under both probe scenarios (`failure→partial_failure`, `failure→success`) the failures output names the marker's historical status and the current verdict is obtainable from the same surface; regression tests encode both orderings. *Invalidated if* the marker is redefined as a pure event log with presentation renamed accordingly.
2. **Carry assessment identity through the presentation layer.** Add `signature_content_hash` (and target verdicts) to `EvaluationView`; attach result provenance (evaluation ID and/or assessed content hash, verification strength) to `FrontierProposalView.Result`; correct the stale comments at `internal/store/evaluation_store.go:117–119` and `internal/pipeline/output.go:696–698`. *Acceptance:* a two-revision fixture renders distinguishable evaluation rows in JSON; `frontier show --json` identifies which evaluation produced `result`; write-path comment matches the `WHERE result IS NULL` behavior.
3. **Make occurrence recency ledger-ordered or clock-validated.** Order `occurrenceResultSQL` by insertion order (rowid / run sequence) with `created_at` as display metadata, or reject non-monotonic `created_at` per (proposal, content, generation) at `PersistEvaluationRun`. *Acceptance:* probe 3 (later-recorded, earlier-stamped assessment) presents the last-recorded verdict or fails the write with an explicit error; all existing reversal and equal-clock tie tests still pass.
