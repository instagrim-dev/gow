---
name: newf-regress
description: >-
  Author temporal regression tests for newf's reassessment lifecycle
  (A -> B -> reassess B -> recompress) using the newf-regress/v1 operator
  vocabulary: specify what changes, what must remain identical, and which
  persisted assessment downstream compression must consume. Use when writing
  or reviewing tests for evidence revisions, generation occurrences,
  re-evaluation, success recompression, completeness admission, cohort
  admission, or when the user mentions temporal regressions, reassessment,
  occurrence selection, or coherent assessment tuples.
---

# newf-regress: temporal regression tests for the reassessment lifecycle

A test layer around newf's research operations — not another research workflow
engine. Every scenario makes selection explicit: no implicit `latest`, no
assumption that the generation that first owned a proposal is the generation
you want to evaluate.

## The contract being protected

> Current guidance follows the selected, admissible reassessment. Historical
> evidence remains unchanged. Neither is allowed to impersonate the other.

## The central distinction

```text
same mechanism
    ≠ same evidence revision
    ≠ same generation occurrence
    ≠ same assessment
```

Learning more about the same proposal must be possible without duplicating the
mechanism, overwriting history, or silently evaluating the old interpretation
again. The mechanism fingerprint deliberately excludes extraction completeness
and unresolved claims, so a revised interpretation dedups onto the same
proposal row while carrying evaluation-relevant changes.

## Implementation boundary (read first)

- **Thin Go integration-test adapter first**, with operators as typed actions
  in `internal/pipeline` integration tests. Add a text parser only afterward,
  if ever.
- Fixtures build prerequisites **through the ordinary services** (ingest,
  normalize, signature, cluster, failure-space, mine, challenge, generate,
  evaluate, compress) — never direct result-row insertion. Direct SQL is
  acceptable only for store-level probes that isolate one query's contract.
- A missing occurrence selector or authority check must surface as a test
  failure naming the missing capability (`capability_missing`), **not** a
  convenient fallback that makes the test green.
- Assertions read **independently persisted state** (store read-backs), never
  a provider's return value.

## Operator vocabulary → repository anchors

| Operator | Responsibility | Where it lives in newf |
|---|---|---|
| `seed` | Build fixture prerequisites through ordinary services | `mineOneCandidate`, `seedPositiveControl`, `ingestNormalizeSign`, `newRealStoreApp` (`internal/pipeline/*_integration_test.go`) |
| `pin` | Freeze target, population, vocabulary, policy context | fixed `now`, `minPreservesMiner` (deterministic mined predicate), a pinned vocabulary version (`mechanism/v1`; `mechanism/v2`/`v3` exist for interpretation-claim and target-canonicalization scenarios — pin ONE per scenario), `biasOnlyChallenger` |
| `declare` / `admit` | Separate an asserted completeness claim from accepted, scoped authority | payload `field_completeness` + `completeness_scope` + `completeness_basis`; code-decided `admission` in `buildApproachInputs` (v26). Only `declared_payload` scope via the deterministic embedded-payload parser is `accepted` |
| `revise` / `emit` | Persist another interpretation without overwriting the earlier one | swap `app.generatorFn` (`pcGenerator` + `pcSignature`) between `GenerateFrontier` calls; same fingerprint ⇒ dedup + new `frontier_proposal_signature_revisions` row + `frontier_generation_contents` binding |
| `reassess` | Evaluate the explicitly selected occurrence; persist its exact binding | `Evaluate(EvaluateInput{ProposalID, GenerationID})`; by-id default = latest occurrence; `GenerationID` pins historical replay |
| `recompress` | Rebuild guidance from the declared current view and compatible assessments | `CompressSuccesses`; cohort via `ListBreakCohortRows` (joins the SELECTED evaluation's `evaluation_target_verdicts`, never origin-time `frontier_target_invariants` flags) |
| `read` / `expect` | Assert independently read persisted state | `openTestStore` + `GetEvaluationRun` (incl. `TargetVerdicts`), `ListGenerationOccurrenceContents`, `GetFrontierGeneration`, `ListBreakCohortRows` |
| `snapshot` / `replay` | Verify historical immutability and exact-context reproducibility | re-read pre-mutation rows post-mutation; `Evaluate` with pinned `GenerationID`; immutability triggers reject UPDATE/DELETE |

`recompress` is a test-facing name for success compression. It does not
introduce a second compression algorithm.

## Authoring workflow

1. **Name the scenario** after the transition under test
   (`unknown_to_break`, `break_to_unknown`, `a_b_a_replay`,
   `untrusted_completeness_cannot_self_certify`, `idempotent_recompress`).
2. **Seed + pin**: fixed `now`, deterministic miner/challenger, authored
   generator signatures. The candidate invariant, its predicate fingerprint,
   and the expected verdicts are **known ground truth**, written down in the
   test before execution.
3. **Emit A, reassess A, recompress** — assert the baseline expectations,
   then `snapshot` the persisted rows (ids + content hashes + verdicts).
4. **Admit or reject the completeness claim** explicitly. Nonempty provider
   prose is not authority: assert the persisted `admission`
   (`accepted` / `declared_only`) and its recorded basis.
5. **Revise → emit B**: assert `SAME_MECHANISM` (fingerprint),
   `SAME_PROPOSAL` (id), `NEW_CONTENT` (content hash differs).
6. **Reassess B by explicit occurrence**: assert the evaluation names B's
   generation, B's content hash, and B's recomputed break verdict.
7. **Recompress with both assessments available**: assert the coherent tuple
   (see below), the support counts, and `HISTORY_RETAINED` (the snapshot
   re-reads unchanged).
8. **Cover the pending interval**: current = B, only A's assessment exists ⇒
   `pending_reassessment`, progressor_count 0. "We have some assessment for
   this proposal" must never read as "we assessed this interpretation."
9. **Add the guard mutations** (below) and the audit record.

## The crucial assertion: B_COHERENT_TUPLE

The proposal occurrence, signature revision, target predicate, break verdict,
and outcome assessment must all come from the **same explicitly bound
assessment**:

```text
proposal occurrence
+ signature content hash        (evaluations.signature_content_hash)
+ target predicate/context
+ predicate verdict             (evaluation_target_verdicts)
+ outcome evaluation            (one selected record; selection-policy/v3)
```

A success outcome cannot supply a missing invariant break. A historical break
cannot authorize a new interpretation.

## Scenario matrix — both directions, always

Only testing "new evidence improves the result" leaves half the contract
untested.

| Scenario | Expected behavior |
|---|---|
| Unobserved → accepted complete | Break changes `unknown` → `violates`; the reassessed occurrence becomes eligible support |
| Accepted complete → unobserved | Break returns to `unknown`; current support disappears; historical support remains queryable |
| Provider asserts completeness without valid authority | Declaration retained (`declared_only`), acceptance rejected, absence stays unknown |
| A → B → A | Re-emitted A is assessed as A despite B's retained revision; pinned replay is exact |
| Identical recompression repeated | Same manifest reuses its revision; no duplicated support |
| Pending interval (current=B, evidence=EA only) | `pending_reassessment`, progressor_count 0 |

## Guard mutations — prove the tests detect the bugs

For each guard, disable it in isolation and require that **at least one named
semantic assertion fails**. A compiler error, crash, missing adapter, or
unrelated failure does **not** count as catching the bug.

```text
mutants {
  use_origin_occurrence   { disable = explicit_occurrence_selection;    must_fail = [B_OCCURRENCE, B_BYTES]; }
  use_origin_break_flag   { disable = revision_bound_break_admission;   must_fail = [NO_ORIGIN_FLAG_REUSE]; }
  trust_nonempty_basis    { disable = completeness_authority_check;     must_fail = [DECLARATION_NOT_AUTHORITY]; }
}
```

In Go, implement a mutant as a scoped behavioral substitution (e.g. resolving
the owner generation instead of the occurrence selector, or querying
`frontier_target_invariants` instead of `evaluation_target_verdicts`) and
assert the named expectation fails for the semantic reason. Input mutations
(changing fixtures) are useful but do not replace guard mutations.

## Audit and replay record

Each run records: the exact manifests, selected references (generation ids,
content hashes, evaluation ids), expected/actual assertion values, and
provider/verifier payload references. Prefer asserting these into the test
failure messages so a red run is its own audit trail.

## Scope honesty

The synthetic fixture's completeness authority covers **its declared list**,
not every property a real method might preserve. Its outcome verifier checks a
small synthetic task independently of the invariant predicate. State this in
the test file header; these controls validate mechanics, not discovery
ability.

## Existing anchors to extend (do not duplicate)

- `internal/pipeline/reassessment_integration_test.go` — the full lifecycle:
  A → B reassess-B (`ReassessRevisedOccurrence`), A → B → A recompression
  (`ReemittedOccurrenceRecompresses`), retraction → selection → policy
  (`CompressionSelectionGovernsPolicy`), and the pending interval
  (`PendingIntervalIsNotSupport`, incl. `HISTORY_RETAINED` snapshots).
- `internal/pipeline/guard_mutation_test.go` +
  `internal/store/guard_mutation_test.go` — the guard-mutation block:
  `use_origin_occurrence`, `trust_nonempty_basis`, `use_origin_break_flag`,
  `use_max_revision_current_view`, `use_latest_revision_not_selection`,
  `use_latest_policy_revision_not_selection` — each pairs the production
  guard with the exact pre-fix behavior and requires the named semantic
  assertion to fail under the mutant. `ListCompressionSelectionsExecutionOrder`
  (store) pins the selection-log ordering the last two mutants depend on.
- `internal/pipeline/persisted_input_control_integration_test.go` —
  ordinary-path seeding + completeness admission regressions.
- `internal/pipeline/frontier_admission_integration_test.go` — the untrusted
  proposal-admission boundary (`canon.AdmitProposalSignature`), external
  proposal arms + preflight, execution attribution under artifact reuse, and
  the experiment-readiness checks (incl. vocab parameterization and the
  real-self-comparison recovery-reachability regression).
- `internal/store/success_store_test.go` — cohort admission probes: both
  verdict directions, legacy binding-unknown handling, current-content
  vs stale strength, v27 reclassification.
- `internal/pipeline/success_integration_test.go` — recompress idempotence at
  BOTH levels: revision-ID reuse and, on a populated revision, per-invariant
  support counts unchanged with a new selection row appended
  (`RecompressSameManifestNoDuplicateSupport`).
- `internal/pipeline/positive_control_integration_test.go` — the populated
  positive control (decisive recovery, completeness-flip, input-mutation
  battery, same-path matrix incl. budget exhaustion).
- `internal/pipeline/interpretation_integration_test.go` — interpretation
  claims (v33) enter as `inferred`, unblock mining, require provenance, and
  change nothing when absent.
- `internal/pipeline/recovery_calibration_integration_test.go` — evaluator
  calibration against the frozen benchmark target (v2 defect pin, v3
  self-match, variant recovery, omitted/unresolved unknowns, v1 legacy pins).
- `internal/pipeline/proposals_validate_test.go` — the capture preflight uses
  the importer decode path and agrees with generation on attested targets.
- `internal/pipeline/experiment_compare_test.go` — epistemic non-promotion of
  inconclusive assessments in arm comparison.
- `internal/canon/vocabulary_v2_test.go` / `vocabulary_v3_test.go` — vocabulary
  revision superset + resolution pins; `TestCompletenessAwareAbsence`
  (`internal/canon/compare_test.go`) — the classify/v2 missing-data contract.
- `internal/provider/untrusted_proposer_test.go` — proposal-wire/v1 strictness
  (smuggled authority, whole-payload decoding, target attribution).
- Remaining gap: no text parser/runner for the DSL exists (the typed Go
  adapter is the implementation). The former idempotent-recompress gap is
  closed by `RecompressSameManifestNoDuplicateSupport`.

## Full syntax reference

For the complete proposed `newf-regress/v1` case syntax (the
`unknown_to_break` scenario, operator semantics, and mutant grammar), see
[reference.md](reference.md).
