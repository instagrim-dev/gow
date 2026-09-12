# Review report — run-10-S4c-B

Reviewed revision: `5dd08618f10bde9ea0eca2aa8b2d39426898155c` (extracted tree; all paths below are relative to that tree root).

## 1. Decision, policy and scope

**Decision: READY for the inspected scope**, with two low-severity findings (one executed, one demonstrated static path) and named limitations. All three repository gates pass on the extracted tree: `go build ./...` (exit 0), `go test ./...` (exit 0, all packages ok), `gofmt -l .` (empty, exit 0).

Scope actually inspected — one causal path, end to end: assessment identity for frontier proposals, from context selection (`internal/pipeline/evaluation.go`), through transactional persistence (`internal/store/evaluation_store.go`), the occurrence-scoped read projection (`internal/store/assessment_views.go`), display/eligibility consumers (`internal/store/frontier_store.go`, `internal/pipeline/frontier.go`), and downstream success-cohort admission (`internal/store/success_store.go`, `internal/pipeline/successes.go`). Signature-population head selection (`signaturePopulationSQL`) and its two readers were inspected. Schema and immutability triggers were inspected in `internal/store/migrations.go` (v14–v27 sections).

Not judged: invariant challenge lifecycle, holdout/experiment machinery, provider adapters, policy mutation, cluster construction, and CLI ergonomics beyond the verdict-display path. A positive decision here is scoped to assessment bookkeeping only; it says nothing about search effectiveness or the truth of any research claim.

## 2. Responsibility and critical path

The critical path for "reassess a claim against new evidence, reproduce every earlier assessment, and show the right result for the requested context" is:

- **Context selection** — `internal/pipeline/evaluation.go:85-118`: explicit `GenerationID` pins historical replay; by-id defaults to `LatestProposalOccurrenceGeneration` (`internal/store/frontier_store.go:550-569`); batch defaults to `LatestFrontierGenerationWithOccurrences` (`frontier_store.go:576-593`). Both selectors order by `g.revision DESC, g.id DESC`, consistent with the cohort's `latest_occ` CTE (`success_store.go:91-100`), so "current view" agrees across evaluation and compression.
- **Content binding before verification** — `evaluation.go:130-136,214-226`: the assessed signature revision is loaded from the generation's occurrence binding (`frontier_generation_contents`, immutable, PK `(generation_run_id, proposal_id)`, `migrations.go:3098-3106`) *before* selection and verification; the persisted evaluation stores the supplied hash verbatim (`evaluation_store.go:79-84,160-164`), never a later "latest" lookup.
- **Atomic persistence** — `PersistEvaluationRun` (`evaluation_store.go:120-211`): one transaction writes run, evaluations, per-target verdicts (v26), metrics, the one-time `frontier_proposals.result` set (NULL→verdict only, trigger-enforced at `migrations.go:1985-2004`), and the `evaluated_failures` marker. Partial writes cannot become eligible evidence.
- **Occurrence-scoped reads** — `occurrenceResultSQL` (`internal/store/assessment_views.go:40-54`) pins verdict lookup to the requested generation, its cluster context, its occurrence membership, and the exact assessed content hash, with deterministic `created_at DESC, rowid DESC` tie-breaking. Both proposal readers (`frontier_store.go:307-311, 627-631`) overwrite the raw sticky column with this projection, so display and batch eligibility (`evaluation.go:149`) consume the ledger, not the legacy first-result field.
- **Cohort admission** — `ListBreakCohortRows` (`success_store.go:88-162`) selects one evaluation per proposal under selection-policy/v3 (content compatibility → decisiveness → strength → recency), joins the revision the selected evaluation *actually assessed*, admits breaks only from that evaluation's own recomputed target verdicts (`t.violated = 1 AND provenance <> 'unverified_legacy'`), and returns `LatestContentHash` so the caller (`internal/pipeline/successes.go:163-230`) marks stale-assessed members pending instead of splicing old outcomes onto new bytes.

## 3. Findings, limitations, protections, questions

### Findings

**F1 — `evaluated_failures` marker is artifact-keyed first-write-wins; later distinct failure assessments cannot record their own marker.**
- Location: `internal/store/migrations.go:2164-2171` (PK `proposal_id`), `internal/store/evaluation_store.go:196-203` (`INSERT OR IGNORE`).
- Triggering conditions: a proposal fails once (e.g. `failure`), then a later assessment — possibly of a *revised* content revision in a new occurrence — fails differently (e.g. `partial_failure`).
- Violated contract: historical failure observations should remain auditable per assessment context. The marker table permanently attributes re-entry to the first failing `evaluation_id`/verdict; the second observation is invisible in this surface (only recoverable by scanning the evaluations ledger directly).
- Downstream consequence: currently contained — the only consumer is the CLI listing (`internal/pipeline/evaluation.go:480-501`, `cmd/newf/evaluation.go:133`), documented as "historical failure markers, not current verdicts". When a future cluster build consumes markers as re-entry input (the R6 intent stated at `migrations.go:2161-2163`), it will re-enter the *first* failure's context even when the informative failure was the revised one.
- Discriminating regression check: two failure verdicts for one proposal via `PersistEvaluationRun`; assert whether the marker surface can represent both. My probe (below) shows it cannot.
- Evidence class: **executed reproduction** — disposable test `TestReviewProbeEvaluatedFailureMarkerIsFirstWriteWins` (created in the extracted tree, `internal/store/zz_review_probe_test.go`): after `failure` then `partial_failure` on the same proposal, `ListEvaluatedFailures` returns one row with verdict `failure` and the first evaluation id, while the ledger retains 2 evaluations. Test passed (exit 0), confirming the behavior.

**F2 — `HasEvaluationForContent` is a dead production contract whose comment describes superseded eligibility semantics.**
- Location: `internal/store/evaluation_store.go:221-234` (implementation), `internal/pipeline/app.go:115-117` (interface declaration).
- Triggering conditions: any future caller adopting this method for batch eligibility.
- Violated contract: the method is content-scoped across *all* contexts (`WHERE proposal_id = ? AND COALESCE(signature_content_hash,'') = ?` — no generation, cluster, or occurrence predicate), while actual batch eligibility is occurrence-scoped (`evaluation.go:137-153` skips on the occurrence-projected `p.Result.Valid`). The interface comment still says "Batch evaluation eligibility is content-scoped", which no longer matches the shipped selection. A revived caller would let a content-only cache hit from an older generation suppress assessment under a new context — exactly the substitution `assessment_views.go:34-38` guards against.
- Downstream consequence: none today (no production caller; only the test fake at `internal/pipeline/app_test.go:465` implements it against the interface). Latent misuse hazard plus interface bloat.
- Discriminating regression check: compile-time — remove the method from the `problemStore` interface (or re-scope it to occurrence identity) and see what breaks; today, nothing does.
- Evidence class: **demonstrated static path** (exhaustive grep across the tree: definition, interface declaration, and one test fake are the only occurrences).

**F3 — Writer comment contradicts sticky-result behavior.**
- Location: `internal/store/evaluation_store.go:119` says "the proposal `result` mirrors the latest verdict written here", but the SQL at `evaluation_store.go:192` (`... WHERE id = ? AND result IS NULL`) and the trigger (`migrations.go:1982-2004`) make it a one-time *first* verdict. Peer comments (`frontier_store.go:47`, `success_store.go:14-16`) describe it correctly as sticky/first.
- Triggering conditions: a maintainer trusting the local comment when adding a reader of the raw column.
- Violated contract: documentation-behavior coherence; the first verdict may even be non-decisive (`verification_blocked` via the H5 path, `evaluation.go:226-237`), permanently occupying the one legal NULL→verdict transition.
- Downstream consequence: none at runtime today — both readers overwrite the column with the occurrence projection before exposure; verified by `TestOccurrenceResultReversalsPreserveInitialHistory` (`internal/store/assessment_views_test.go:100-133`).
- Discriminating regression check: existing test above already pins the actual (sticky) behavior; the fix is a comment correction.
- Evidence class: inspected-only (comment vs. SQL on the same screen).

### Limitations (honest gaps, not defects)

- Atlas re-entry is a marker only: failure markers do not materialize signatures or admit evidence; cluster builds do not consume them yet (`evaluation.go:476-479`). The admitted-population query for atlas re-entry therefore could not be traced — it does not exist.
- Pre-v17 evaluations carry empty `signature_content_hash`; they are explicit attribution gaps, counted (`IneligibleUnpersisted` / `BindingUnknown`, `successes.go:174-190`), never backfilled.
- Migrated-database behavior was exercised only through the repository's own migration tests, not against a real legacy corpus.
- `latestOccurrenceResult` is an N+1 per-proposal query in both list readers (`frontier_store.go:307,627`) — a scale concern, not a correctness one.

### Observed protections

- Occurrence view cannot cross generations or content revisions; equal-clock ties broken by `rowid` (executed: `TestOccurrenceResultDoesNotCrossGenerationOrContent`, `TestLegacyResultNeedsMatchingAssessmentContext`).
- Head selection precedes vocabulary filtering; an unsigned head cannot resurrect a superseded signature; backdated corrections win over timestamps (executed: `TestSignaturePopulationCurrentAndHistorical`).
- Guard-mutation tests actively mutate the cohort SQL toward known-wrong shapes (origin break flags, max-revision current view, latest-revision join, selection-vs-latest) and assert the wrongness is detectable (`internal/store/guard_mutation_test.go:23-297`).
- Re-emitted identical content in a new generation still requires its own occurrence assessment (executed: `TestIntegrationReemittedContentRequiresOccurrenceAssessment`).
- Stale targets yield `verification_blocked` rather than a silently narrowed question (`evaluation.go:226-237`); blockers are selected into cohorts only when nothing decisive exists and are counted ambiguous, never support (`success_store.go:63-70`, `successes.go:216-222`).
- Immutability triggers cover the evaluation ledger, occurrence bindings, and signature revisions; `PRAGMA foreign_keys = ON` at open and `foreign_key_check` before migration commit (`internal/store/store.go:56,132-140`).

### Open questions

- Can a `frontier_generation_contents` row exist whose `content_hash` matches no retained revision row? If so, `current_rev` is empty and a hash-less legacy evaluation ranks tier-0 in `selected_eval` (`success_store.go:118-121`); the outer projection still empties its content (counted ineligible), so no support is fabricated, but the selection ordering would be surprising. Writer-side coupling was not traced to closure.
- Is the permanent batch-skip after a non-decisive occurrence verdict (`evaluation.go:149`: any `Result.Valid`, including `verification_blocked`) the intended operator workflow, given only explicit by-id re-evaluation escapes it?

## 4. Checks, coverage and resources

Commands executed (all inside the extracted tree unless noted):

- `git -C <repo> archive 5dd08618… | tar -x -C $TMP && rm -rf $TMP/docs/reviews` — exit 0 (setup).
- `go build ./...` — exit 0.
- `go test ./...` — exit 0 (all 17 packages ok, including `internal/store` 6.3s, `internal/pipeline` 10.5s).
- `gofmt -l .` — empty output, exit 0.
- `go test ./internal/store/ -run TestReviewProbeEvaluatedFailureMarkerIsFirstWriteWins -v` — exit 0, PASS (my disposable probe, F1).
- Assorted `grep`/`sed` inspections — exit 0.

Files consulted: `AGENTS.md` (via injected rules), `internal/store/assessment_views.go`, `assessment_views_test.go` (partial), `evaluation_store.go` (full), `frontier_store.go` (result/selector regions), `success_store.go` (lines 1–230, plus grep of remainder), `migrations.go` (v14–v27 comments; frontier/evaluation/revision/occurrence DDL regions), `store.go` (pragma region), `guard_mutation_test.go` (test names), `internal/pipeline/evaluation.go` (lines 1–300, 470–501), `pipeline/successes.go` (lines 140–260), `pipeline/app.go` (grep), `pipeline/frontier.go` (grep), `cmd/newf/evaluation.go` (grep), `internal/pipeline/assessment_views_integration_test.go` (test name).

NOT examined: `internal/{invariant,cluster,policy,experiment,relational,verify,canon,normalize,provider,config,domain}` internals beyond grep hits; challenge lifecycle and bounded-claim scope semantics; holdout/leakage machinery; concurrency under `-race`; migrated legacy databases; CLI output stability; `docs/`, `corpus/`, `paper/`, `fixtures/`. Required adversarial cases not covered by this pass: concurrent revision insertion / injected transaction failure; cross-problem same-local-ID contamination beyond the `p.problem_id = g.problem_id` guard inspection; empty-evidence-population compression behavior (inspected counters only); the full discovery→A0→A1 seam as an executed end-to-end (covered piecewise by existing tests, not by one seam test).

Wall-clock estimate: ~22 minutes.

## 5. Remediation handoffs

1. **Re-key or version the failure re-entry marker** (F1). Either widen `evaluated_failures` to PK `(proposal_id, evaluation_id)` (append-only history of re-entries) or add an explicit "marker is origin-failure-only" contract at the future cluster-consumption seam. Acceptance: a second, later failure assessment of a revised occurrence is representable and listable; existing single-failure behavior unchanged; migration preserves existing marker rows verbatim; my F1 probe scenario yields two auditable observations.
2. **Retire or re-scope `HasEvaluationForContent`** (F2). Delete the interface method and store implementation, or rename/rescope it to occurrence identity and correct the `app.go:115-117` comment to match the shipped occurrence-scoped eligibility. Acceptance: `go build ./...` clean with no production caller regression; no interface method whose documented semantics contradict `evaluation.go:137-153`.
3. **Correct the sticky-result comment** (F3) at `evaluation_store.go:113-119` to state the first-verdict-once semantics (matching `frontier_store.go:47`), and note that the first verdict may be non-decisive. Acceptance: comment matches the `WHERE result IS NULL` SQL and the trigger contract; no behavior change; existing reversal-history test still passes.
