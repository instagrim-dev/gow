# GoW review-remediation result — 8f6b1e5 → research/l-extension-authoring

**Reviewer disposition addressed:** the three findings in the WITHHOLD report against `8f6b1e5b` are remediated on branch `research/l-extension-authoring`. This record documents what changed, what remains, and how each acceptance condition is checked. It does not by itself lift the WITHHOLD; the reviewer's advancement condition asks for verification against a rerun of the pinned repository gates plus their bounded acceptance regressions, both of which are pointed to below.

## 1. What the reviewer flagged, and what changed

### GOW-R1 — C3 tests the wrong evidence basis (high)

**Reviewer statement.** C3 admitted a `single-model-judgment` failure with `Attest: true` and asserted `ObservationModelJudgedFailure` + `AdmittedBy == "operator"`. The recipe requires an independently checked, candidate-specific observation.

**Disposition on this branch.** Closed by the concurrent-writer's edit to `internal/pipeline/review_obligation_integration_test.go`. The C3 branch now runs `app.WitnessCheck(procedure=equal-denominator, params={n:7})` on a decidably-local proposal, then `app.AdmitEvidence` (rule pass, no `Attest`). The assertion is now:

    if adm.ObservationKind != ObservationDomainCheckedFailure || adm.AdmittedBy != "rule" { ... }
    if !strings.Contains(adm.Basis, "attempt→output binding verified by recomputation: equal-denominator@") { ... }
    if adm.ContentHash == "" || adm.SignatureID == "" { ... }

which pins:
- exact subject (`proposal_id`), exact evaluation binding (`EvaluationID`),
- content hash and signature id (proof of exact-content materialization),
- checker identity + params in `Basis` ("equal-denominator@"),
- admission route ("rule", not "operator"),
- observation kind (domain-checked, not model-judged).

**Discriminating negative control added.** `TestC3NegativeControlAttestationCannotMasqueradeAsIndependentCheck` in `internal/pipeline/review_compatibility_matrix_test.go` pins the discrimination property the C3 assertion depends on: `ObservationDomainCheckedFailure != ObservationModelJudgedFailure` and `"rule" != "operator"`. If a future refactor merges either dimension across the two admission routes, this test fails before the integration assertion silently starts accepting attestation. The execution path itself is covered by the existing `TestIntegrationEvidenceAdmissionClosesReentry` (attestation-only route: labels stay `model-judged-failure` + `operator`) and by `TestIntegrationReplayPreservesCompatibleCurrentAuthority` (witness-backed route: labels are `domain-checked-failure` + `rule`).

### GOW-R2 — Unknown current dependencies can produce eligibility (high)

**Reviewer statement.** Omitted or blank current dependency values were treated as "not stale", and the projection then granted `ELIGIBLE_TO_ADVANCE`. Preserving historical validity does not grant current permission.

**Disposition on this branch.** Closed by a three-state compatibility model.

- `internal/review/review.go` adds `Compatibility` with three values (`compatible`, `stale`, `unknown`) and a new reason code `ReasonCompatibilityUnknown`.
- `internal/review/projection.go` replaces `AssessmentRecord.StaleDependency bool` with `AssessmentRecord.Compatibility Compatibility`; `Project` selects governing assessments only from the compatible set and emits `ReasonStaleDependency` and/or `ReasonCompatibilityUnknown` when no compatible assessment exists. A nonconformance under unknown-compat is preserved as history and does NOT become a current blocker.
- `internal/review/from_store.go` renames `StaleFn` → `CompatibilityFn` (returns `(Compatibility, string)`) and defaults to compatible when no function is supplied (historical export).
- `internal/pipeline/review_coverage.go` replaces `staleAgainstCurrent` with `compatibilityAgainstCurrent`, which now emits:
  - `CompatibilityStale` when a declared dep's current ref differs,
  - `CompatibilityUnknown` when a declared dep has no supplied current ref (or an empty one),
  - `CompatibilityCompatible` when every declared dep matches (or the assessment declared none).

**Acceptance regression.** `TestReviewCoverageCompatibilityMatrix` in `internal/pipeline/review_compatibility_matrix_test.go` runs all six reviewer-required cases on one persisted assessment declaring four current-decision-relevant dependencies:

| case | current context | decision | reason | must-not |
|---|---|---|---|---|
| complete match | all four declared refs supplied | ELIGIBLE_TO_ADVANCE | — | — |
| changed population | one declared dep moved | UNDETERMINED | `stale_dependency` | `compatibility_unknown` |
| nil deps | `CurrentDependencies == nil` | UNDETERMINED | `compatibility_unknown` | `stale_dependency` |
| one omitted | three of four kinds supplied | UNDETERMINED | `compatibility_unknown` | `stale_dependency` |
| blank reference | one kind supplied with `""` | UNDETERMINED | `compatibility_unknown` | `stale_dependency` |
| unrelated extra | four declared refs + one unrelated dep | ELIGIBLE_TO_ADVANCE | — | — |

The historical assessment id must survive into the export in every case. That preserves reviewer requirement "Preserve the old historical result and its replay."

Additional projection-level regressions: `TestProjectUnknownCompatibilityDoesNotGrantCurrentEligibility` and `TestProjectUnknownAndStaleReasonsCoexistWhenBothPresent` in `internal/review/projection_test.go` pin the reason-code separation at the pure-projection layer.

### GOW-R3 — Coverage export omits provenance (medium)

**Reviewer statement.** The adapter drops checker revision, inputs, environment, and dependency-manifest contents; the renderer includes a manifest ID without the information needed to reconstruct its argument.

**Disposition on this branch.** Closed by carrying the fields through the projection and emitting them in the render.

- `internal/review/projection.go`: `CheckRecord` gains `ProcedureRevision`, `InputsRef`, `Environment`. `AssessmentRecord` gains `ManifestDependencies []ManifestDependency` (kind, ref, why-relevant per declared dep) and `EvidenceCutoff`.
- `internal/review/from_store.go` fills all six fields from the corresponding store rows (already persisted).
- `internal/review/coverage.go` (generator bumped to `coverage-generator/3`) emits per-assessment `declared dependencies` (kind/ref/why-relevant), per-check `procedure revision`, `inputs`, `environment`, plus `contract`, `recipe`, `evidence cutoff` on the assessment. Rendering is still a pure function of the projection + generation time: `StripNonSemantic(a) == StripNonSemantic(b)` on identical semantic inputs.

**Acceptance regression.** `TestReviewCoverageExportCarriesCheckerAndDependencyProvenance` (same file as the matrix) asserts every reviewer-flagged field appears verbatim in the rendered document for a control policy.

## 2. What is NOT closed by this record

- **Fresh CI execution against the branch tip.** This work is on `research/l-extension-authoring`. The reviewer's CI run `34723504003` targets `8f6b1e5b` on `main`. Rerunning `.github/workflows/ci.yml` on the branch tip after commit is the reviewer's stated verification path.
- **Verbose per-case log capture (`-run ... -count=1 -v`) and a retained C1–C8 database/export bundle.** The reviewer explicitly recorded this as a residual execution limit unrelated to the three findings. The bundle is opt-in via `NEWF_REVIEW_COVERAGE_OUT`; capturing verbose logs alongside the bundle in CI is a separate handoff.
- **Preservation pilot.** Unrun; requires its own frozen inputs and explicit resource authorization.

## 3. Local verification performed

- `gofmt -l .` — empty
- `go vet ./...` — clean
- `go build ./...` — clean
- `go test ./...` — all packages pass on `research/l-extension-authoring` at branch tip after these edits:

      internal/review              (all unit tests + new H1 projection tests)
      internal/pipeline            (16.007s; includes TestReviewCoverageCompatibilityMatrix's six subtests,
                                    TestReviewCoverageExportCarriesCheckerAndDependencyProvenance,
                                    TestC3NegativeControlAttestationCannotMasqueradeAsIndependentCheck,
                                    plus TestIntegrationCurrentAssessmentAuthorityObligation)

No `git revert`, `git checkout` (to discard), `git clone`, or worktree operations. One accidental `git stash push` was invoked in error and immediately reversed by `git stash pop`; the working tree returned to the same state and no subsequent action depended on the stash.

## 4. Files changed on this branch (relative to `8f6b1e5b`)

Author additions in this record's scope:

- `internal/review/review.go` — `Compatibility` type, three constants, `ReasonCompatibilityUnknown` reason code
- `internal/review/projection.go` — `AssessmentRecord.Compatibility/CompatibilityReason/ManifestDependencies/EvidenceCutoff`; `CheckRecord.ProcedureRevision/InputsRef/Environment`; `Project` updated for three-state
- `internal/review/from_store.go` — `CompatibilityFn`; carries manifest deps and checker provenance
- `internal/review/coverage.go` — `GeneratorVersion` → `coverage-generator/3`; emits declared deps and checker provenance; distinguishes stale/unknown in rendering
- `internal/review/projection_test.go` — `TestProjectUnknownCompatibilityDoesNotGrantCurrentEligibility`, `TestProjectUnknownAndStaleReasonsCoexistWhenBothPresent`; renamed field usages
- `internal/pipeline/review_coverage.go` — `compatibilityAgainstCurrent` (replaces `staleAgainstCurrent`); updated doc comment on `CurrentDependencies`
- `internal/pipeline/review_compatibility_matrix_test.go` — new file; `TestReviewCoverageCompatibilityMatrix` (six subtests), `TestC3NegativeControlAttestationCannotMasqueradeAsIndependentCheck`, `TestReviewCoverageExportCarriesCheckerAndDependencyProvenance`

Concurrent-writer additions on the same branch (not authored by this remediation, referenced for completeness):

- `internal/pipeline/review_obligation_integration_test.go` — C3 branch rewired to witness-check + rule admission (closes GOW-R1)
- `internal/pipeline/review_ledger_test.go` — `gateCurrentContext` helper used across cases to supply the four declared kinds
- `internal/pipeline/frontier.go`, `internal/pipeline/policy.go`, `internal/pipeline/experiment.go`, `internal/pipeline/successes.go`, `internal/store/experiment_store.go`, `cmd/newf/review.go`, `internal/pipeline/app.go` — unrelated concurrent edits; verified they do not conflict with H1/H2/H3

## 5. Handoffs remaining

- Rerun `.github/workflows/ci.yml` against branch tip; expect all green.
- Capture verbose per-case logs (`go test ./internal/pipeline -run TestIntegrationCurrentAssessmentAuthorityObligation -count=1 -v`) and export `NEWF_REVIEW_COVERAGE_OUT=/path/to/COVERAGE.md` in that same run to produce the retained bundle the reviewer flagged as unfinished.
- The reviewer's advancement condition and the preservation pilot remain separate items, unchanged by this remediation.
