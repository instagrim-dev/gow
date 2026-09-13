# Run report — subject revision 5dd08618f10bde9ea0eca2aa8b2d39426898155c

## 1. Decision, policy and scope

**Decision: NEEDS_CHANGES (narrow), gate-level PASS.** Reviewed revision `5dd08618f10bde9ea0eca2aa8b2d39426898155c`, extracted as a clean tree (no VCS metadata consulted); Go 1.26.6 darwin/arm64; schema migrations run to version 33. All repository gates pass: `go build ./...` (exit 0), `go test ./...` (exit 0, all 17 packages ok), `gofmt -l .` empty (exit 0).

Scope: assessment identity and ledger-derived views — whether the system can reassess a claim against explicitly selected new evidence, reproduce every earlier assessment, and show the correct result for exactly the requested context. I traced the write path `Evaluate → PersistEvaluationRun` and the read paths `latestOccurrenceResult`, `ListOccurrenceProposalRows`, `ListBreakCohortRows`, `signaturePopulationSQL`/`ListHistoricalSignaturesForProblem`, and success-revision selection (`LatestSelectedSuccessRevision`). The core occurrence/content/verdict identity machinery is correct and well tested. The NEEDS_CHANGES verdict rests on one executed-reproduction finding: the persisted evaluation run does not bind the invariant-revision / normalization context it assessed under, leaving one component of the assessment tuple unreconstructible. Positive conclusions are scoped to inspected paths only; policy, cluster, interpretation, and experiment stores were not audited in depth (Section 4).

## 2. Responsibility and critical path

The critical assessment path is:

1. **Content identity** — `internal/store/frontier_store.go:511` `insertSignatureRevision`: append-only revisions keyed by sha256 of persisted bytes, `UNIQUE(proposal_id, content_hash)` + `PRIMARY KEY(proposal_id, revision)` (`internal/store/migrations.go:3018`), immutability triggers. Identical content is a revision no-op, so A→B→A re-emission reuses A's row and cannot create duplicate current-revision rows.
2. **Occurrence identity** — `frontier_generation_contents` (`migrations.go:3098`, immutable) binds (generation, proposal) → content hash; `insertGenerationOccurrence` (`frontier_store.go:535`). v24 migration backfills only the origin occurrence from revision 1 — an honest reconstruction, not invented history.
3. **Assessment write** — `internal/pipeline/evaluation.go:62` `Evaluate` resolves the assessment context (explicit `GenerationID` pin, by-id → latest occurrence generation, batch → latest generation with occurrence membership), recomputes per-target verdicts against the occurrence's exact bytes, and stores the **supplied** `signature_content_hash` verbatim (`internal/store/evaluation_store.go:160-163`) — never an independent "latest" lookup. `PersistEvaluationRun` (`evaluation_store.go:120`) writes run, evaluations, per-target verdicts, metrics, the one-time NULL→verdict proposal `result`, and failure markers in a single transaction.
4. **Contextual read** — `occurrenceResultSQL` (`internal/store/assessment_views.go:40`) scopes "current result" to the exact generation, cluster-run, and occurrence content hash, with `created_at DESC, rowid DESC` deterministic tie-break; both `loadFrontierGeneration` (`frontier_store.go:307`) and `ListOccurrenceProposalRows` (`frontier_store.go:627`) overwrite the raw sticky column with this projection.
5. **Cohort consumption** — `ListBreakCohortRows` (`internal/store/success_store.go:86`) selects one evaluation per proposal under selection-policy/v3 (content compatibility ≻ decisiveness ≻ strength ≻ recency), takes verdict/strength/evaluation-id from that single record, admits breaks only from the selected evaluation's own recomputed `evaluation_target_verdicts` (excluding `unverified_legacy` provenance), and flags hash-less evaluations on multi-revision proposals as `BindingUnknown` rather than binding them to current bytes.
6. **Current guidance** — `LatestSelectedSuccessRevision` (`success_store.go:477`) follows the latest compression *selection*, not `MAX(revision)`, so a deduped re-selection of an older artifact governs.

View semantics as implemented: *current* = latest assessment in the exact occurrence context; *historical* = generation-pinned replay, retained and reproducible; *all-history* = explicit flag in `signaturePopulationSQL` / `ListHistoricalSignaturesForProblem` (`canon_store.go:490`), with head selection (supersession before timestamps) preceding vocabulary filtering.

## 3. Findings, limitations, protections, questions

### Findings

**F1 — Evaluation runs never record their invariant/normalization context (Medium).**
- **Location:** `internal/pipeline/evaluation.go:201-210` (record construction); columns exist and are honored at `internal/store/evaluation_store.go:141-144`.
- **Triggering conditions:** every `newf evaluate` invocation. Target predicates are resolved from *live* challenge state at call time (`survivingInvariants`, `internal/pipeline/frontier.go:363`, states `surviving`/`operator_attested`), but the record sets neither `InvariantRevisionID` nor `NormalizationRevisionID`.
- **Violated contract:** the assessment identity tuple must bind "which target claim revisions and policy context were used this time." The schema provides `evaluation_runs.invariant_revision_id` / `normalization_revision_id` for exactly this; the writer leaves them NULL unconditionally.
- **Downstream consequence:** after challenge states move (e.g. a target later weakens), an auditor cannot reconstruct from the run row which invariant-state snapshot defined the target set at assessment time. The persisted `evaluation_target_verdicts` pin which targets *were* assessed, but "target absent because stale at assessment time (H5 block)" is indistinguishable from "never targeted" without cross-correlating challenge timestamps. Replay of the assessment context is therefore incomplete, though no wrong verdict is displayed.
- **Discriminating regression check:** after a full mine→challenge→generate→evaluate pass, assert `GetEvaluationRun(runID).InvariantRevisionID != ""`; then change a challenge state, re-evaluate, and assert the two runs record different invariant contexts.
- **Evidence class: executed reproduction.** A disposable pipeline test (fixture providers, `t.TempDir()` SQLite) ran the full flow and read back `InvariantRevisionID=""`, `NormalizationRevisionID=""` while `ClusterRunID` was correctly bound. Probe file removed afterward.

**F2 — CLI/JSON evaluation views omit the assessed content revision and per-target verdicts (Low–Medium).**
- **Location:** `internal/pipeline/output.go:783-795` (`EvaluationView`); mapping at `internal/pipeline/evaluation.go:504-547` drops `SignatureContentHash` and `TargetVerdicts`.
- **Triggering conditions:** any `newf evaluate` / `evaluation show` / `evaluation list` on a proposal with multiple content revisions.
- **Violated contract:** "every displayed or consumed verdict retains … scope and relevant provenance." The ledger stores the assessed hash and recomputed target verdicts, but the machine-readable surface does not expose them.
- **Downstream consequence:** an operator comparing the revised-occurrence assessment with a pinned historical replay sees two JSON records that differ only in verdict/notes; which bytes each assessed is invisible without raw SQL. View identity is preserved in storage but not at the exposed boundary.
- **Discriminating regression check:** in the existing `TestIntegrationReassessRevisedOccurrence` setup, assert the `evaluation show` JSON for runs A and B carries distinct `signature_content_hash` values.
- **Evidence class: demonstrated static path** (struct definition plus mapping code; storage side confirmed executed via the suite).

**F3 — `HasEvaluationForContent` is dead surface with a stale eligibility claim (Low).**
- **Location:** `internal/pipeline/app.go:115-117` (interface), `internal/store/evaluation_store.go:221-234` (implementation + comment).
- **Triggering conditions:** none at runtime — no production caller exists in `internal/pipeline` or `cmd/newf` (grep confirmed; only fakes in tests implement it).
- **Violated contract:** the comment states batch eligibility is content-scoped via this method; actual batch eligibility is occurrence-result-scoped (`evaluation.go:142-153` skipping `p.Result.Valid`, which derives from `latestOccurrenceResult`). The actual mechanism is *stronger* (context-scoped), so behavior is correct, but the interface advertises a check nothing performs.
- **Downstream consequence:** a future maintainer may call it believing it encodes the eligibility policy; it answers "any evaluation of these bytes, in any context," which would reintroduce the cross-generation cache-hit bug the comments elsewhere warn against.
- **Discriminating regression check:** remove the method (or wire it where content-scoped short-circuiting is genuinely wanted) and confirm `go build ./... && go test ./...` stay green.
- **Evidence class: inspected-only** (exhaustive symbol search).

**F4 — `PersistEvaluationRun` doc claims `result` "mirrors the latest verdict"; it is first-verdict sticky (Doc-only, Low).**
- **Location:** `internal/store/evaluation_store.go:113-119` (comment) vs `:192` (`UPDATE … WHERE id = ? AND result IS NULL`).
- **Triggering conditions:** any re-evaluation of a proposal.
- **Violated contract:** internal documentation accuracy. Behavior is safe — all shipped readers overwrite the column with `latestOccurrenceResult` — but the comment invites a raw read of `frontier_proposals.result` as "latest."
- **Downstream consequence:** misleads future direct-store consumers; `BreakCohortRow`'s doc (`success_store.go:14-15`) correctly calls it the sticky first result, so the two comments contradict.
- **Discriminating regression check:** existing `TestOccurrenceResultReversalsPreserveInitialHistory` already pins behavior; fix is a comment correction.
- **Evidence class: inspected-only.**

### Limitations

- I did not audit policy directive derivation, cluster membership reads, interpretation-store head selection internals, or experiment/holdout stores beyond confirming `signaturePopulationSQL` semantics and the policy-selection guard tests exist and pass.
- Concurrency behavior (`-race`, concurrent revision insertion, injected transaction failure) was not exercised beyond the suite's own coverage; `PersistEvaluationRun`'s single-transaction shape was verified statically plus via existing tests.
- Fresh-database migration was exercised implicitly by every integration test; a migrated-from-legacy database was not constructed independently — legacy-path conclusions rely on `TestLegacyResultNeedsMatchingAssessmentContext` and the v27 backfill code inspection.
- A shared temp log path I initially used appeared to contain unrelated content; I discarded it and re-ran all gates with clean, uniquely named logs. Reported exit codes come from the clean re-run only.

### Observed protections

- Immutability triggers on signature revisions, occurrence bindings, evaluations, and target verdicts; `UNIQUE(proposal_id, content_hash)` prevents current-view join multiplication.
- Cross-problem guards: `Evaluate` rejects a generation from another problem (`evaluation.go:113-115`); `occurrenceResultSQL` re-checks `p.problem_id = g.problem_id` and cluster-run equality in SQL.
- Deterministic tie-breaks everywhere ordering matters (`created_at DESC, rowid/id DESC`), tested at `internal/store/assessment_views_test.go:100,135,171`.
- Selection-policy/v3 ranks content compatibility above strength, so a stronger stale evaluation cannot override a completed reassessment (executed: `TestIntegrationReassessRevisedOccurrence`).
- `BindingUnknown` handling refuses to bind hash-less legacy evaluations to current bytes on multi-revision proposals; `unverified_legacy` target-verdict provenance never admits cohort support.
- Mutation-style guard tests (`internal/store/guard_mutation_test.go`, `internal/pipeline/guard_mutation_test.go`) actively assert that the *wrong* selectors (origin break flags, MAX(revision), latest-revision-not-selection) would fail — unusually strong regression discipline.
- H5 stale-target handling records `verification_blocked` instead of silently shrinking the question; conservative default model verifier abstains rather than manufacturing verdicts.

### Open questions

- Is run-level invariant-context binding (F1) intentionally deferred (per-target verdict rows may have been judged sufficient), or an oversight? The unused schema columns suggest the latter.
- `evaluation_runs.normalization_revision_id` — what writer was ever intended to populate it for proposal-mode runs?
- Should `evaluated_failures` markers carry the assessed content hash? Currently they are keyed `(proposal_id, evaluation_id)` with verdict only; downstream atlas re-entry semantics were out of my inspected scope.

## 4. Checks, coverage and resources

Commands executed (inside the extracted subject tree; all exit statuses as shown):

- `go build ./...` — exit 0
- `gofmt -l .` — empty output, exit 0
- `go test ./...` — exit 0 (17/17 packages ok; pipeline 9.5s, store 5.5s)
- `go test ./internal/pipeline -run TestReviewProbeEvaluationRunAttribution -v` — probe test **failed as predicted** (both attribution fields empty), demonstrating F1; probe file then deleted and suite-green state restored
- ripgrep symbol searches for `HasEvaluationForContent`, `latestOccurrenceResult`, `InvariantRevisionID`, `survivingInvariants`, raw `result` readers, schema versions

Files consulted (relative to subject tree): `AGENTS.md`; `internal/store/assessment_views.go`, `assessment_views_test.go`, `evaluation_store.go`, `frontier_store.go` (lines 200–712), `success_store.go`, `canon_store.go` (population query), `migrations.go` (v23/v24/v26/v27 sections, version list), `guard_mutation_test.go`; `internal/pipeline/evaluation.go`, `frontier.go` (340–420), `reassessment_integration_test.go`, `assessment_views_integration_test.go` (inventory), `guard_mutation_test.go` (inventory), `app.go` (interface), `output.go` (views); `internal/domain/id.go` (ULID minting — collision-safe under mutex).

Not examined: `internal/store/{policy_store,cluster_store,failure_space_store,interpretation_store,experiment_store,challenge_store,invariant_store,normalize}.go` internals beyond targeted greps; `cmd/newf` command wiring beyond the evaluation surface; `docs/` prose; corpus/paper/fixtures; concurrency under `-race`; migrated-legacy database construction; holdout/experiment identity.

Wall-clock estimate: ~22 minutes.

## 5. Remediation handoffs

1. **Bind assessment context on evaluation runs (F1).** In `Evaluate`, resolve and set `InvariantRevisionID` (the revision(s) backing the targetable states used for predicates — if multiple, introduce a run→invariant-state snapshot join table rather than faking a single ID) and the current `NormalizationRevisionID`. *Acceptance:* new integration test asserts non-empty context on a persisted run; a challenge-state change between two evaluations yields distinguishable recorded contexts; existing suite stays green; NULLs on pre-existing rows remain as honest attribution gaps (no backfill from current state).
2. **Expose assessed revision and target verdicts in evaluation views (F2).** Add `signature_content_hash` and `target_verdicts` (additive, `omitempty`) to `EvaluationView` and populate in `evaluationRunView`. *Acceptance:* `evaluation show` JSON for the revised-occurrence vs pinned-replay runs in the existing reassessment scenario carries the two distinct hashes; no existing field changes shape.
3. **Retire or wire `HasEvaluationForContent`; fix the sticky-result comment (F3+F4).** Either delete the interface method and store implementation or give it a real, correctly documented caller; rewrite the `PersistEvaluationRun` comment to say the proposal `result` is the immutable *initial* verdict and contextual reads go through `latestOccurrenceResult`. *Acceptance:* build/test/gofmt gates green; no remaining comment claims batch eligibility is enforced by this method.
