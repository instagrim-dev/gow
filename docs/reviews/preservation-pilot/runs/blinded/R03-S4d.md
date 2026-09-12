# Review report — run-06-S4d-B

## 1. Decision, policy and scope

**Reviewed revision:** `3443a1094407364be2e02264ffa1ee1bb0f3efde` (extracted tree; no VCS metadata consulted).

**Decision: NEEDS_CHANGES (documentation/display coherence), with the core assessment-identity machinery judged sound on the inspected paths.** All repository gates pass (`go build ./...`, `go test ./...`, `gofmt -l .` empty). The persistence layer implements a serious, layered assessment-identity model: append-only evaluation ledger, immutable content revisions, occurrence bindings, recomputed per-target verdicts with provenance, and a content-compatibility-first selection policy for cohort admission. One executed reproduction demonstrates a divergence between the documented contract for `frontier_proposals.result` and its actual sticky-first-verdict behavior, surfaced unqualified in the `frontier show` view. Remaining findings are static-path or inspected-only.

**Policy applied:** one selected causal path (frontier proposal → evaluation → persisted verdict → cohort admission → compression input) traced end-to-end; bounded adversarial cases against that path; repository gates executed; one focused reproduction test written and run in an isolated copy with per-test temp SQLite. No provider calls, no writes outside the extracted tree. A positive judgment is scoped strictly to inspected behavior; uninspected surfaces are named in section 4.

## 2. Responsibility and critical path

Concept-to-implementation map (as found):

- **Logical artifact:** `frontier_proposals` row, deduped on `UNIQUE(problem_id, proposal_hash)`; immutable by trigger (`internal/store/migrations.go`).
- **Exact interpretation:** `frontier_proposal_signature_revisions` — append-only `(proposal_id, sha256(signature_json))`, `PRIMARY KEY(proposal_id, revision)`, `UNIQUE(proposal_id, content_hash)` (`internal/store/migrations.go:3018–3027`); writer `insertSignatureRevision` (`internal/store/frontier_store.go:504–521`).
- **Occurrence:** `frontier_generation_contents` — immutable binding generation → (proposal, emitted content hash) (`internal/store/migrations.go:3098–3104`); writer `insertGenerationOccurrence` (`frontier_store.go:528–534`).
- **Assessment identity:** `evaluations` row carrying `signature_content_hash` (the exact bytes the verifier assessed, supplied by the pipeline, never a latest lookup — `internal/store/evaluation_store.go:79–84`), plus `evaluation_target_verdicts` (recomputed per-target verdicts with `provenance` column, v26/v27) and either deterministic tool identity or a `provider_invocations` row (role `evaluate`).
- **Execution identity:** `runs` row + `evaluation_runs` row; retries/re-evaluation are new append-only runs.
- **View identity:** explicit selectors — `GenerationID` pins historical/replay context; by-ID defaults to latest emitted occurrence (`LatestProposalOccurrenceGeneration`, `frontier_store.go:543–562`); batch defaults to latest generation with occurrence membership; cohort "current view" is the latest emitted occurrence, falling back to max revision only for pre-v24 history (`internal/store/success_store.go:90–112`).
- **Discovery identity:** origin generation + origin-time `frontier_target_invariants` flags are retained and explicitly *not* consumed for cohort admission (v26).
- **Evidence-manifest / assessment-policy identity: partially absent.** The evaluation run binds `frontier_generation_run_id` and `cluster_run_id`, but `invariant_revision_id`/`normalization_revision_id` columns are never populated by the writer (finding F3).

**Critical write/read trace:** `App.Evaluate` (`internal/pipeline/evaluation.go:62–308`) resolves the assessment context, enumerates occurrence membership, applies content-scoped eligibility (`HasEvaluationForContent`), recomputes target verdicts against the exact occurrence bytes, routes verifiers, and persists everything in one transaction via `PersistEvaluationRun` (`internal/store/evaluation_store.go:120–211`), which also does the one-time `result` populate (`WHERE result IS NULL`) and failure re-entry marker. Read side: `ListBreakCohortRows` (`success_store.go:85–178`) selects one evaluation per proposal under selection-policy/v3 (content compatibility → decisiveness → strength → recency), joins the *assessed* revision, flags binding-unknown legacy rows, and `buildBreakCohorts` (`internal/pipeline/successes.go:163–229`) excludes pending/ambiguous members with explicit counts.

## 3. Findings, limitations, protections, questions

### Findings

**F1 — Sticky first-verdict `result` diverges from the ledger and is displayed and documented as if it cannot.**
- **Location:** `internal/store/evaluation_store.go:192` (writer), `internal/store/evaluation_store.go:113–119` (contradicting doc comment), `internal/pipeline/frontier.go:576–577` + `internal/pipeline/output.go:710` (display), `docs/evaluation.md:112–113` (documented contract).
- **Triggering conditions:** any reassessment (explicit by-ID, or a revised occurrence evaluated in a later run) whose verdict differs from the first evaluation's verdict.
- **Violated contract:** `docs/evaluation.md` states the proposal's `result` is populated in the same transaction as the evaluation and "they can never diverge"; the `PersistEvaluationRun` comment says `result` "mirrors the latest verdict written here". The implementation is `UPDATE frontier_proposals SET result = ? WHERE id = ? AND result IS NULL` — a one-time set. After a divergent reassessment, `result` is the *first* verdict while the ledger's current-view evaluation says otherwise. `frontier show` emits this as a bare `result` field with no first-assessment qualification (the view doc says only "Result is empty until M5.2 evaluation populates it").
- **Downstream consequence:** an operator or tool reading `frontier list/show` sees a stale verdict presented as *the* result; decisions taken from that surface can contradict cohort admission (which correctly uses the ledger). This is exactly a legacy first-result field silently readable as current verdict.
- **Discriminating regression check:** persist E1=failure on revision R1, then E2=partial_success on revised occurrence R2; assert the display contract — either `result` is labeled/renamed first-assessment, or it is derived from the current-view selection. My disposable test (see §4) asserts the current divergent behavior: artifact `result="failure"` while `ListBreakCohortRows` selects E2 `partial_success`.
- **Evidence class:** **executed reproduction** (test run, passed, log captured).
- **Mitigating protection:** cohort admission, pending detection, and policy inputs never consume `result`; the blast radius is the human/JSON display surface and documentation.

**F2 — `evaluated_failures` cannot represent a second, distinct failure of the same proposal.**
- **Location:** `internal/store/migrations.go:2164–2170` (`proposal_id TEXT PRIMARY KEY`), `internal/store/evaluation_store.go:196–203` (`INSERT OR IGNORE`).
- **Triggering conditions:** a proposal fails under content revision R1; a revised interpretation R2 also fails in a later evaluation.
- **Violated contract:** occurrence/assessment identity on the atlas re-entry read path — the marker is artifact-keyed, so the second failure's `evaluation_id`/context is silently dropped (`INSERT OR IGNORE` on an existing PK), and `ListEvaluatedFailures` reports the first failure's evaluation as the re-entry context forever. The marker is also immutable by trigger, so it can never be updated.
- **Downstream consequence:** the next `cluster build` re-entry input attributes failure structure to stale evaluation identity; a revised-content failure is invisible as a distinct sample in `FailureSpace(t+1)`.
- **Discriminating regression check:** two failure runs on distinct content hashes for one proposal → assert either two markers (occurrence-keyed) or a documented single-marker semantics naming which evaluation the marker binds.
- **Evidence class:** demonstrated static path (schema + writer code; not executed).

**F3 — Evaluation run does not bind the invariant-lifecycle snapshot it consulted.**
- **Location:** `internal/pipeline/evaluation.go:210–219` (record construction omits `InvariantRevisionID`/`NormalizationRevisionID`), `internal/store/evaluation_store.go:141–145` (columns written as passed, i.e. NULL).
- **Triggering conditions:** auditing or replaying an evaluation after invariant lifecycle transitions (targetable → weaken/falsified) or a new mining revision.
- **Violated contract:** the assessment tuple's policy/target-set component. Target predicates are loaded from each candidate's *originating* revision (stable per candidate, `internal/pipeline/frontier.go:363–390`), and `evaluation_target_verdicts` enumerates which targets actually received verdicts — but *why* a target is absent (stale at assessment time) or why the verdict is `verification_blocked` is only reconstructible by correlating challenge-history timestamps, not by an explicit binding on the run row, despite the columns existing.
- **Downstream consequence:** an explicit reproduction of "which targetable set did this assessment see" degrades from a foreign-key read to a temporal inference; two assessments of identical bytes under different lifecycle states are distinguishable only indirectly.
- **Discriminating regression check:** evaluate; transition a target to falsified; re-evaluate by ID; assert the two evaluation runs are distinguishable by an explicit persisted context field rather than by timestamps.
- **Evidence class:** demonstrated static path.

**F4 — Batch eligibility is content-scoped across all contexts, not occurrence-scoped.**
- **Location:** `internal/store/evaluation_store.go:225–234` (`HasEvaluationForContent`: any evaluation, any generation, any verdict — including `unknown`/`verification_blocked`); consumed at `internal/pipeline/evaluation.go:150–160`.
- **Triggering conditions:** the same signature bytes are re-emitted in a later generation after the target set, cluster population, or predicates' lifecycle state changed; or the only prior assessment was non-decisive (`verification_blocked` from a stale target).
- **Violated contract:** the intentional repeated-batch no-op is declared for "an already-assessed identical occurrence"; this check instead treats *any* historical assessment of the bytes as covering the *new* occurrence's context. Batch never refreshes; only explicit by-ID or generation-pinned evaluation reaches it.
- **Downstream consequence:** the next decision can consume an old-context assessment as though the new occurrence were assessed; a permanently blocked verdict (H5) also permanently suppresses batch reassessment of those bytes.
- **Discriminating regression check:** same bytes emitted in generation G2 after a target-set change → assert batch either evaluates the new occurrence or the skip is documented as content-scoped by contract.
- **Evidence class:** inspected-only; moderate confidence this is a contract-wording gap rather than an implementation accident (comments assert intent for the identical-occurrence case only).

### Limitations (of the system, honestly recorded by it)

- Pre-v17 proposals without persisted content: attribution gap recorded (`SignatureContentHash` empty), counted `IneligibleUnpersisted`, never backfilled.
- Pre-v24 dedup occurrences historically unrecoverable; readers fall back to max revision and say so (`migrations.go:915–922`).
- Eligibility check and evaluation persist are separate transactions; concurrent batch evaluates could both assess the same occurrence (append-only ledger tolerates; no mixed-context write possible).

### Observed protections (verified by reading code and the passing test suite)

- Immutability triggers on evaluations, runs, target verdicts, revisions, occurrences, markers; v27 migration's trigger drop/recreate is a scoped one-time reclassification that only *weakens* claims (`recomputed` → `unverified_legacy`).
- Verifier-assessed bytes stored verbatim, never a latest lookup (`evaluation_store.go:79–84`; `TestIntegrationEvaluationStampsAssessedRevision`).
- Selection-policy/v3: current-content compatibility outranks stronger stale evidence; decisiveness gates before strength; deterministic tie-breaks (`success_store.go:113–136`); binding-unknown legacy rows flagged, never filled from current bytes.
- Cohort admission uses recomputed assessment-context verdicts, excludes `unverified_legacy`, and marks newer-unassessed interpretations pending (`successes.go:163–229`).
- A→B→A re-emission: current view is latest *emitted occurrence*, not max revision (`TestIntegrationReemittedOccurrenceRecompresses`).
- Stale targets produce `verification_blocked`, never a silently reduced question (H5, `evaluation.go:235–247`).
- Old-context replay cannot overwrite `result` (`WHERE result IS NULL`) or any immutable row.
- Guard-mutation tests assert that wrong-identity queries (origin flags, max-revision view, latest-not-selection) would fail tests (`internal/store/guard_mutation_test.go`, `internal/pipeline/guard_mutation_test.go`).
- Compression current guidance follows the execution/selection log, not `MAX(revision)` (v28, `success_store.go:470–492`).

### Open questions

- Is `frontier show`'s `result` intended as first-assessment-only? No document says so; `docs/evaluation.md` asserts the opposite.
- Should batch eligibility be occurrence-scoped (proposal, generation, content) rather than (proposal, content)?
- Cross-problem contamination: `HasEvaluationForContent` and cohort CTEs key on globally-unique proposal IDs; I did not adversarially test same-content proposals under two problems.

## 4. Checks, coverage and resources

**Commands executed (all inside the extracted tree; exit statuses):**
- `go build ./...` — exit 0.
- `gofmt -l .` — exit 0, empty output.
- `go test ./...` — exit 0 (all 17 packages ok; `internal/pipeline` 9.5s, `internal/store` 5.5s).
- `go test ./internal/store/ -run TestReviewReproStickyResultDivergesFromCurrentEvaluation -v` — exit 0; PASS; logged `artifact result="failure"; selected eval=… verdict="partial_success"` (disposable test file `internal/store/zz_review_repro_test.go` created in the temp copy only).

**Files consulted (relative to subject tree):** `AGENTS.md`; `internal/store/evaluation_store.go`, `frontier_store.go`, `success_store.go`, `migrations.go` (targeted segments), `evaluation_store_test.go`, `success_store_test.go` (fixtures + case inventory), `frontier_store_test.go` (helper), `guard_mutation_test.go` (test names); `internal/pipeline/evaluation.go`, `frontier.go` (segments), `successes.go` (cohort building), `invariant.go` (segments), `challenge_views.go` (segments), `output.go` (view struct), `reassessment_integration_test.go` / `evaluation_integration_test.go` / `guard_mutation_test.go` (test inventories); `docs/evaluation.md` (targeted), `docs/success-compression.md` (grep); `internal/store/normalize.go`, `canon_store.go`, `interpretation_store.go` (quick supersession/head scan only).

**Not examined:** experiment arms (`internal/store/experiment_store.go`, `internal/pipeline/experiment*.go`) including `experiment_arm_proposals.signature_content_hash` consumers; policy read paths in depth (`policy_store.go`, `policy.go` beyond the guard test names); challenge state machine internals; cluster/failure-space stores; canon/interpretation head selection beyond a surface scan (supersession lineage exists; head-before-filter ordering not verified); CLI command wiring; holdout refusal gates beyond test names; migrated-from-legacy databases (fresh-DB tests only); `-race` runs; the required cross-problem contamination and injected-transaction-failure adversarial cases (rollback-on-error is structural via `defer tx.Rollback()`, inspected only).

**Wall-clock estimate:** ~24 minutes.

## 5. Remediation handoffs

**H1 — Reconcile `frontier_proposals.result` semantics across writer comment, docs, and view.** Either (a) declare it first-assessment-only: fix `docs/evaluation.md:112` ("can never diverge") and the `PersistEvaluationRun` comment ("mirrors the latest verdict"), and rename or annotate the `result` JSON field (e.g. `first_result`), or (b) derive the displayed verdict from the current-view ledger selection. **Acceptance:** a regression test persisting E1=failure then E2(revised occurrence)=partial_success asserts the displayed field matches the declared semantics; docs and code comments agree; no consumer change to cohort admission.

**H2 — Cement the batch-eligibility scope contract.** Decide whether `HasEvaluationForContent` should be occurrence-scoped (add generation to the key) or whether cross-generation identical-bytes skip is the declared no-op semantics; document the decision where the repeated-batch no-op is described. **Acceptance:** a test covering "same bytes, new generation, changed target set" pins the chosen behavior; `verification_blocked`-only prior assessments have an explicit documented refresh path.

**H3 — Bind assessment-time policy context and fix failure-marker identity.** Populate `evaluation_runs.invariant_revision_id` (or an explicit as-of/target-set snapshot field) in `App.Evaluate`; re-key or supplement `evaluated_failures` so a second failure under a distinct content revision is representable (occurrence-keyed marker or per-content marker), with a migration that leaves legacy markers as explicit first-failure records. **Acceptance:** two-failure scenario yields distinguishable, auditable markers (or documented single-marker semantics); an evaluation run's consulted target-set context is readable by foreign key, not timestamp inference; migration tested on a populated legacy database.
