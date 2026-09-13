# Review report — run-01-S3d-B

Reviewed revision: `3443a1094407364be2e02264ffa1ee1bb0f3efde` (extracted tree; all paths below are relative to it).

## 1. Decision, policy and scope

**Decision: NEEDS_CHANGES**, scoped to the inspected assessment-identity and derived-view paths. Gate-level results are clean: `go build ./...` passes, `go test ./...` passes across all packages, and `gofmt -l .` is empty. The evaluation write path, success-cohort read path, and their schema enforcement are unusually disciplined (transactional writes, immutability triggers, supplied-hash attribution, content-compatibility-ranked selection). The NEEDS_CHANGES verdict rests on one demonstrated population-selection defect in the clustering read path (Finding 1) plus a writer-comment/behavior contradiction on the sticky proposal `result` (Finding 2). The core question — can the system reassess a claim against explicitly selected new evidence, reproduce earlier assessments, and show the correct result for the requested context — is answered **yes for the evaluation/compression pipeline** and **no for the clustering population**, where supersession of interpretations is not honored.

Scope: I inspected `internal/store` (evaluation, success, frontier, canon, normalize, migrations), `internal/pipeline` (evaluation, successes, cluster, frontier views, reassessment/guard-mutation tests), `internal/verify`, and `internal/domain/id.go`. I did not audit `internal/{policy,experiment,witness,lean,projection,review}`, challenge/interpretation stores, CLI wiring beyond spot checks, or migrated legacy databases. No clean bill of health is claimed for those.

## 2. Responsibility and critical path

The critical write/read path for one assessment:

1. **Selection** — `internal/pipeline/evaluation.go:87-165`: context resolution (explicit `GenerationID` pin, by-id → latest occurrence generation, batch → latest generation with occurrence membership), then content-scoped eligibility via `HasEvaluationForContent`.
2. **Computation** — `verificationContextForProposal` (`evaluation.go:415-444`) recomputes per-target verdicts against the occurrence-bound signature bytes; stale (weaken/falsified) targets block routing entirely (`evaluation.go:235-247`) rather than silently shrinking the question.
3. **Persistence** — `internal/store/evaluation_store.go:120-211` (`PersistEvaluationRun`): one transaction covering the run row, per-evaluation rows carrying the **supplied** `signature_content_hash` (never a latest-revision lookup), per-target verdicts (`evaluation_target_verdicts`), metrics, provider invocation, the one-time `frontier_proposals.result` set, and the `evaluated_failures` marker. Rollback on any error; immutability triggers (`migrations.go:2176-2233`) prevent post-hoc rewriting.
4. **Loading/display** — `loadEvaluationRun` rehydrates verdict + verifier kind + strength + target verdicts; `evaluationRunView` preserves them.
5. **Downstream consumption** — `internal/store/success_store.go:85-178` (`ListBreakCohortRows`): current view = latest emitted occurrence (`latest_occ` CTE), evaluation selection ranks content compatibility → decisiveness → verification strength → recency, `BindingUnknown` legacy rows and stale-content members become *pending*, never support (`internal/pipeline/successes.go:163-232`). `success_compression_selections` (v28) makes current guidance follow the latest execution, not `MAX(revision)`.

View semantics as implemented: **current** = latest emitted occurrence + content-compatible evaluation; **historical** = pinned `GenerationID` replay, append-only rows retained; **all-history** = `ListEvaluations` / `ListCompressionSelections` audit ordering. The clustering path (`internal/pipeline/cluster.go:90-105`) sits outside this discipline — see Finding 1.

## 3. Findings, limitations, protections, questions

### Findings

**F1 — Cluster population ignores approach supersession (severity: high for invariant-mining semantics).**
- Location: `internal/store/canon_store.go:488-501` (`ListSignaturesForProblem` SQL); consumers `internal/pipeline/cluster.go:90`, `internal/pipeline/readiness.go:126`.
- Triggering conditions: a source approach is re-normalized. `internal/pipeline/normalize.go:211-231` creates a new `approach_revisions` row superseding the prior one (lineage confirmed by `internal/store/normalize_test.go:172-202`); each revision owns its own mechanism. If both revisions' mechanisms are signed under the same `(schema_version, vocabulary_version)`, the selection query — which joins `mechanism_signatures → mechanisms → approach_revisions → approaches` with **no head/supersession predicate** — returns both signatures.
- Violated contract: current-population selection must honor explicit supersession before version filtering; one logical approach contributes one current interpretation. Two corollary violations: (a) a superseded interpretation remains a current failure-space member indefinitely; (b) if the *head* revision is unsigned, the old revision's signature silently stands in for the approach (an unsigned head resurrects an old signature).
- Downstream consequence: `cluster build` clusters both revisions; `familiesForFailureSpace` (`internal/pipeline/invariant.go:61-90`) then feeds inflated member/support counts into invariant mining, so a single re-normalized approach can double-count toward candidate-invariant support, and frontier nearest-family comparisons inherit the stale interpretation.
- Discriminating regression check: normalize one snapshot, sign, re-normalize the same approach with changed resolved content, sign the new mechanism, run `cluster build`; assert the persisted cluster-run population contains exactly one signature per approach (the head's), and that an unsigned head yields an explicit gap rather than the superseded signature.
- Evidence class: demonstrated static path (SQL inspected; supersession lineage proven by existing executed tests; no head predicate exists anywhere on this path).

**F2 — `frontier_proposals.result` writer comment contradicts sticky-first behavior; views surface it unqualified (severity: medium).**
- Location: comment `internal/store/evaluation_store.go:113-119` ("the proposal `result` mirrors the latest verdict written here") vs. `evaluation_store.go:192` (`UPDATE ... WHERE ... result IS NULL`) and the trigger at `internal/store/migrations.go:1982-2003` (one-time NULL→verdict; overwrite aborts). Reader: `internal/pipeline/frontier.go:576-578` copies it into `FrontierProposalView.Result`.
- Triggering conditions: any proposal evaluated more than once with a different verdict — e.g. first assessment `failure`, revised occurrence reassessed `success`.
- Violated contract: a legacy first-result field must not be presented as the current verdict; the append-only ledger is authoritative for contextual reads.
- Downstream consequence: `frontier show`/`list` JSON reports the first-ever verdict as `result` with no first-vs-current qualifier, so an operator or script reading the frontier view sees a stale verdict after a legitimate reassessment. The high-value consumers are protected (break-cohort selection ignores `result` and uses selection-policy/v3; batch eligibility uses it only as the documented pre-v17 fallback at `evaluation.go:158`), so this is a display/API-semantics defect plus a false contract comment, not corrupted state.
- Discriminating regression check: evaluate a proposal (verdict V1), pin-replay or reassess its revised occurrence to a different verdict V2, assert the frontier view either reflects V2 or explicitly labels the field as the first/origin verdict.
- Evidence class: demonstrated static path (SQL, trigger, and view mapping all inspected; behavior forced by the trigger).

**F3 — Batch eligibility key omits assessment context beyond content bytes (severity: low; partially intentional).**
- Location: `internal/store/evaluation_store.go:221-234` (`HasEvaluationForContent`: keyed on `(proposal_id, signature_content_hash)` across **all** generations); consumer `internal/pipeline/evaluation.go:150-158`.
- Triggering conditions: unchanged content re-emitted in a later generation (A→B→A re-emission, exercised for compression by `TestIntegrationReemittedOccurrenceRecompresses`) after the comparison population, cluster run, or target set has changed.
- Violated contract: an assessment of old context must not be treated as an assessment of the new context merely because bytes match. The repeated-batch no-op for an *identical* occurrence is intentional and preserved; the gap is that the eligibility predicate cannot distinguish "same occurrence re-run" from "same bytes, new generation/population".
- Downstream consequence: batch `evaluate` on generation G3 skips the re-emitted proposal; the next decision then consumes the G1-context evaluation as current support (it ranks content-compatible in `success_store.go:113-136`) even though its nearest-family comparison evidence came from G1's cluster run. Mitigation: explicit by-id or `--generation` reassessment reaches the new context, and comments declare re-evaluation explicit — so this is context *staleness*, not silent rewriting.
- Discriminating regression check: G1 emit+assess content A; admit new failure families / new cluster run; G3 re-emit A; assert batch evaluate either reassesses or reports the skipped proposal's assessment context (generation/cluster-run id) so the staleness is visible.
- Evidence class: demonstrated static path.

**F4 — `evaluated_failures` marker frozen at first failure (severity: low).**
- Location: `internal/store/evaluation_store.go:196-203` (`INSERT OR IGNORE`), schema `migrations.go:2164-2171` (`proposal_id` PRIMARY KEY), immutability triggers `migrations.go:2220-2233`.
- Triggering conditions: a proposal fails, is later reassessed (revised occurrence) and fails again with a different verdict or evaluation.
- Violated contract: marker claims to record "newly-evaluated failure" re-entry, but only the first failure's `evaluation_id`/`verdict` can ever exist; a later distinct failure event is unrecordable by construction.
- Downstream consequence: minor — the only consumer is the CLI listing (`cmd/newf/evaluation.go:133`) feeding explicit re-entry decisions; re-entry presence is still correct, but its provenance can point at a superseded assessment.
- Discriminating regression check: two failure evaluations for one proposal with different verdicts; assert the marker either records both events or documents first-event-only semantics at the read site.
- Evidence class: demonstrated static path.

### Limitations

- No new executable reproductions were written; all findings are static traces over code whose surrounding behavior is covered by the executed suite.
- Fresh-schema behavior only; migrated legacy databases (pre-v17/v24/v26/v27 data) were not constructed.
- `-race` and concurrency injection were not run.
- Two early inspection commands executed in the wrong working directory due to shell-state loss; their output was discarded and every cited fact was re-derived from the subject tree.
- Packages `internal/{policy,experiment,witness,lean,projection,review}`, challenge/interpretation/failure-space stores, and CLI ergonomics were not audited.

### Observed protections

- Single-transaction evaluation persistence with rollback; partial writes cannot become eligible evidence (`evaluation_store.go:135-211`).
- Assessed-bytes attribution is supplied by the verifier path and stored verbatim — no independent "latest" lookup can stamp a computation with bytes it never saw (`evaluation_store.go:79-84`, `evaluation.go:188-190`).
- Immutability triggers on evaluations, runs, metrics, markers, approach revisions, signatures (`migrations.go`, throughout).
- Selection-policy/v3: content compatibility outranks decisiveness outranks strength; blockers are never support; `BindingUnknown` legacy rows and newer-unassessed-content members become pending, counted, never fabricated (`success_store.go:39-136`, `successes.go:163-232`).
- Stale-target H5 blocking: a weakened/falsified target forces explicit reassessment instead of silently shrinking the question (`evaluation.go:235-247`).
- Deterministic tier decides only the code-certain negative; empty target sets yield `unknown`, never fabricated coverage (`internal/verify/deterministic.go:32-49, 86-102`).
- v28 compression selections make current guidance follow the latest execution, not `MAX(revision)`, with executed guard-mutation tests (`internal/store/guard_mutation_test.go`).
- Monotonic ULID minting under a mutex plus `created_at DESC, id DESC` tie-breaks give deterministic selection under equal timestamps (`internal/domain/id.go:85-86,349-354`).
- Reassessment lifecycle covered by executed integration tests: revised-occurrence reassessment, re-emission recompression, batch reaching revised content, pending-interval-not-support (`internal/pipeline/reassessment_integration_test.go`).

### Open questions

- Is the dual-signature population of F1 reachable through the standard CLI workflow (does the signing command only ever target the latest mechanism?), or does it require signing an older mechanism id explicitly? Either way the store boundary accepts it.
- What is the intended external semantics of the frontier view's `result` field — origin verdict or current verdict?
- Should batch-eligibility answers carry the assessed generation/cluster-run id so context staleness is operator-visible?

## 4. Checks, coverage and resources

Commands executed (subject tree unless noted; all exit 0 unless stated):

- `git archive 3443a10... | tar -x -C $TMP && rm -rf $TMP/docs/reviews` — exit 0 (setup, run against the source repository once, before inspection began).
- `go build ./...` — exit 0.
- `go test ./...` — exit 0 (`TEST_EXIT=0`; all packages ok, `internal/pipeline` 9.98s, `internal/store` 5.57s).
- `gofmt -l .` — exit 0, empty output.
- ~8 `grep`/`sed` inspection commands over store/pipeline/domain/verify sources — exit 0.
- Two early commands ran with a wrong working directory (shell state loss); output discarded, facts re-derived (noted under Limitations).

Files consulted (partial reads noted): `internal/store/evaluation_store.go`, `internal/store/success_store.go`, `internal/store/frontier_store.go` (rows/contract sections), `internal/store/canon_store.go` (signature listing), `internal/store/migrations.go` (evaluation/frontier/marker schema + triggers), `internal/store/normalize.go` (grep), `internal/store/normalize_test.go` (grep), `internal/store/guard_mutation_test.go` (test names), `internal/pipeline/evaluation.go`, `internal/pipeline/successes.go` (cohort building), `internal/pipeline/cluster.go` (population build), `internal/pipeline/frontier.go` (view mapping), `internal/pipeline/canon_views.go`, `internal/pipeline/invariant.go` (grep), `internal/pipeline/output.go` (grep), `internal/pipeline/reassessment_integration_test.go` (three test bodies), `internal/pipeline/guard_mutation_test.go` (test names), `internal/verify/deterministic.go`, `internal/domain/id.go` (id minting).

NOT examined: `cmd/newf/*` beyond greps; `internal/{policy,experiment,witness,lean,projection,review,canon,cluster,frontier,invariant,success}` algorithm internals; challenge/interpretation/failure-space/experiment stores; migration behavior on populated legacy databases; `corpus/`, `docs/`, `paper/`, `fixtures/`, `testdata/`.

Wall-clock estimate: ~25 minutes.

## 5. Remediation handoffs

1. **Head-select the clustering population (F1).** Change `ListSignaturesForProblem` (or add a head-scoped variant used by `cluster build` and `readiness`) to select, per approach, the signature of its latest non-superseded revision, applying supersession before `(schema, vocabulary)` filtering; an unsigned head must surface as an explicit gap, not fall back to the superseded signature. Acceptance: new store/pipeline test in which one approach carries two signed revisions — cluster population contains exactly the head's signature; a second test with an unsigned head asserts the approach is reported missing rather than represented by the old signature; existing cluster/invariant integration tests still pass; migration/replay note documenting that historical cluster runs are immutable artifacts and are not retroactively re-membered.
2. **Reconcile the `result` contract (F2, folds in F4).** Either (a) fix the comment at `evaluation_store.go:113-119` to state first-verdict-once semantics and rename/annotate the view field (`origin_result` or a doc-tagged qualifier in `FrontierProposalView`), or (b) implement genuine mirroring with an explicit trigger change. Acceptance: a regression test evaluating one proposal to two different verdicts asserts the displayed field matches the declared contract; `evaluated_failures` first-event-only semantics documented at the read site or the marker keyed by `(proposal_id, evaluation_id)`.
3. **Make skipped-eligibility context visible (F3).** When batch `evaluate` skips a proposal as already-assessed, include the assessing evaluation's generation/cluster-run binding in the response (or a `stale_context` flag when it differs from the generation under evaluation), preserving the intentional identical-occurrence no-op. Acceptance: integration test with A→B→A re-emission across changed cluster runs asserts the skip is reported with the original assessment context; no automatic re-evaluation is introduced.
