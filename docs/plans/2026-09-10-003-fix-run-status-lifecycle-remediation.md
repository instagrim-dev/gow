# Remediation: run-status lifecycle honesty

Date: 2026-09-10
Mode: holistic-review-remediator (delivery)

## Findings addressed

| # | Sev | Finding | Disposition |
|---|-----|---------|-------------|
| F1 | High | Runs report `completed`/`initialized` regardless of per-item outcome | **Fixed** |
| F2 | Medium | Run status write-once in storage (no update path) | **Fixed** (root enabler) |
| F3 | Low | `runs.status` schema has no `CHECK` constraint | **Deferred after attempt** (see below) |

## Root cause

Every pipeline command created its `runs` row with a terminal/initial status
(`completed`/`initialized`) at creation time and never transitioned it, because
the store exposed only `CreateRun` — no mutation path. Run-level telemetry
(`run show`) was therefore frozen before any work happened, so automation using
`run show` as source-of-truth saw false positives.

## Change

- `store.Store.UpdateRunStatus(ctx, id, status, completedAt, *errorSummary)` —
  the single, validating mutation path for run lifecycle. Rejects unknown run
  IDs (`ErrNotFound`) and out-of-domain statuses.
- `App.finalizeRun` / `App.failRun` helpers centralize the transition:
  - runs are created `running`;
  - `finalizeRun(nil failures)` → `completed`;
  - `finalizeRun(failures)` → `failed` with a deterministic, order-independent
    `error_summary` (`N item(s) failed: ...`, sorted);
  - `failRun` marks single-outcome commands `failed` on early error (best-effort;
    never masks the original error).
- Wired at all 7 caller sites: ingest, normalize, mechanism signature,
  mechanism compare (write path), cluster build, failure-space build,
  mechanism seed-fixture.

Command exit codes are intentionally unchanged: partial failure remains
informative-not-fatal per AGENTS.md ("failure is a first-class artifact"). The
run record — not the exit code — now carries degraded/failed status.

## Tests

- `internal/store`: `TestUpdateRunStatusTransitionsLifecycle`,
  `TestUpdateRunStatusRejectsUnknownRunAndStatus`.
- `cmd/newf`: `TestCLIIngestReportsPerInputFailures` extended to assert
  `run show` → `failed` + non-empty `error_summary`;
  `TestCLIIngestSuccessReportsCompletedRun` asserts a clean run → `completed`
  with no `error_summary`.

## F3 deferred-after-attempt rationale (reversibility cliff)

A DB-level `CHECK (status IN (...))` on `runs.status` requires a SQLite table
rebuild (no `ALTER ... ADD CHECK`). `runs(id)` is FK-referenced by ~7 tables
(`created_by_run_id`, `ingest_run_id`, and `run_id` in `provider_invocations`,
`normalization_revisions`, `mechanism_signatures`, `comparison_runs`,
`cluster_runs`, `failure_spaces`). A copy/drop/rename of `runs` under those
constraints is a high-risk, hard-to-reverse migration disproportionate to a Low
finding. The status invariant is already enforced in-code at two layers:
`NewRun.Validate` (creation) and `UpdateRunStatus` (transition). Defense-in-depth
via DB CHECK can be revisited if a rebuild of `runs` becomes necessary for
another reason.

## Unrelated failure reported (not repaired)

`internal/store` test `TestMigrateRejectsMissingSchemaTables` fails on the
current tree: it stamps `currentSchemaVersion` (9) as applied against an empty
DB and expects `ErrCorruptStore`, but `Migrate` returns nil. This is a
migration-versioning/test-contract mismatch introduced by the in-flight
clustering schema work (uncommitted churn in `internal/store/migrations.go`),
not by this remediation. Per AGENTS.md, reported rather than silently repaired.
The remediation slice (`cmd/newf`, `internal/pipeline`, and the new
`internal/store` run-status tests) builds and passes; `gofmt -l` is clean on all
touched files.
