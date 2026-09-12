# Review report — population selection and interpretation revisions

Reviewed revision: `5dd08618f10bde9ea0eca2aa8b2d39426898155c`

## 1. Decision, policy and scope

**Decision under review:** when one underlying work item (approach) has multiple interpretation revisions, which revisions are eligible for the default signature population that feeds clustering, failure spaces, and invariant support — and whether that selection is explicit and typed across its three access modes (current heads, pinned historical replay, all-history).

**Decision projection: `WITHHOLD`.** One demonstrated, unresolved blocking nonconformance applies (Finding F1: identity drift across a superseded re-normalization double-counts one work item in the default population; executed reproduction). Remaining uncertainties are listed in section 3.

**Policy status:** no versioned decision policy naming an accountable owner, obligation IDs, or resource ceilings is present in the subject tree. Independently of F1, that gap would cap the projection at `UNDETERMINED` with reason `policy_missing_or_unauthorized`; the demonstrated blocker dominates. Review mode was read-only plus disposable tests on an isolated extracted tree with isolated SQLite stores (`t.TempDir()`); no providers, no repository modifications.

**Scope:** `internal/store` population selection and normalization lineage, their pipeline consumers (`cluster`, `readiness`, `invariant` inputs), schema constraints in `internal/store/migrations.go`, and identity handling in `internal/pipeline/normalize.go` / `internal/normalize/schema.go`. Everything else was out of scope (section 4).

## 2. Responsibility and critical path

The causal path from a source interpretation to an invariant-support figure is:

```text
snapshot -> normalization_revision (provider output; logical_identity per approach)
  -> approach (get-or-create by (problem_id, logical_identity))      internal/store/normalize.go:311
  -> approach_revision (auto-supersedes latest prior revision)       internal/store/normalize.go:212-228
  -> mechanism (UNIQUE per approach_revision)                        internal/store/migrations.go:166-174
  -> mechanism_signature (UNIQUE per (mechanism, schema, vocab))     internal/store/migrations.go:350-359
  -> default population = head revisions only                        internal/store/assessment_views.go:14-32
  -> cluster run (members pinned by signature id)                    internal/pipeline/cluster.go:90-143
  -> failure space -> invariant families -> DistinctFamilySupport,
     FailureCoverageNum/Den                                          internal/invariant/engine.go:118-126
```

The single load-bearing responsibility is `signaturePopulationSQL` (`internal/store/assessment_views.go:14`): for each approach it admits only the revision with no successor in the `approach_revisions.supersedes_revision_id` chain, choosing the head **before** applying the schema/vocabulary filter so a missing current signature returns nothing rather than resurrecting a superseded one. Everything downstream (cluster family counts, failure-space `distinct_family_count`, invariant `DistinctFamilySupport` and `FailureCoverage`) inherits its correctness. The critical weakness is that this supersession chain is derived **solely from logical-identity string equality** at persist time; the separately recorded normalization-revision supersession lineage (`normalization_revisions.supersedes_revision_id`, set by forced re-normalization at `internal/pipeline/normalize.go:211-217`) is never consumed by population selection.

## 3. Findings, limitations, protections, questions

### Finding F1 — identity drift across a superseded re-normalization double-counts one work item

- **Location:** `internal/store/assessment_views.go:14-32` (population selection consumes only approach-revision lineage); seam created at `internal/store/normalize.go:206-228` (lineage derived only from identity equality) and `internal/pipeline/normalize.go:196-263` (no reconciliation of provider-emitted identities against the revision being superseded).
- **Triggering conditions:** a snapshot is re-normalized (e.g. `normalize --force`, or a new provider/model/config hash) and the provider emits a different `logical_identity` string for the same underlying approach. Nothing constrains, cross-checks, or warns about identity drift; identities are free-form provider text (`internal/normalize/schema.go:86-88`, `docs/normalization.md:47-50`).
- **Violated contract:** a corrected interpretation of one approach must not count as independent support; the population layer's own contract, "A superseded interpretation is historical evidence, not another current observation" (`internal/store/canon_store.go:480-481`), is not honored when the supersession is recorded at the normalization-revision level rather than the approach-revision level.
- **Downstream consequence:** both the stale and corrected interpretations are unsuperseded heads, so the default population gains an extra member. With a changed mechanism family (the very case a correction exists for), the two signatures land in **different** cluster families, so the invariant engine's one-support-per-family cap does not mask it: `DistinctFamilySupport`, `FailureCoverageNum/Den` (`internal/invariant/engine.go:121-123`), cluster `family_count`, and failure-space `distinct_family_count` all inflate. A candidate invariant can cross `minSupport` on what is a single work item interpreted twice.
- **Discriminating regression check:** persist normalization N1 (identity `es/x`, signed), then N2 with `SupersedesRevisionID = N1` and identity `es/x-drifted` (signed, changed mechanism posture); assert `ListSignaturesForProblem` returns exactly 1 member. Today it returns 2.
- **Evidence class: executed reproduction.** A disposable test in `internal/store` driving only public APIs (`PersistNormalization` twice with the exact supersedes linkage the forced pipeline sets, `PersistSignature`, `ListSignaturesForProblem`) produced `2` default-population members for one snapshot's superseded interpretation chain (test failed as designed; suite otherwise green).
- **Severity / confidence:** high impact on invariant-support integrity; medium likelihood (depends on provider identity stability, which is unenforced and unmeasured). Confidence in the mechanism: high (executed).
- **Smallest coherent remedy:** when a normalization revision supersedes another for the same snapshot, exclude (or flag as conflicted) approach heads whose only support derives from the superseded normalization revision — or require an explicit identity mapping at re-normalization time.

### Finding F2 — cross-source auto-supersession conflates correction with independent attestation

- **Location:** `internal/store/normalize.go:212-228` (`persistApproachTx`): any new revision for an existing approach automatically supersedes the latest prior revision, regardless of whether it re-interprets the same snapshot or comes from a completely different source snapshot.
- **Triggering conditions:** two different snapshots (independent sources) yield the same `logical_identity` for a problem. The second write silently demotes the first source's interpretation to historical.
- **Violated contract:** supersession is the record meaning "corrected interpretation of the same observation"; an independent attestation from another source is not a correction. Recording it as supersession destroys the distinction the population layer relies on (and that F1's remedy would rely on even more).
- **Downstream consequence:** the population count stays at 1 (no inflation — consistent with one-member-per-approach), but the head's mechanism content, outcome class, and epistemic claims are last-writer-wins by ingestion order. A lower-quality later source silently replaces a better interpretation as the approach's current evidence; no conflict or attestation record exists to detect it.
- **Discriminating regression check:** ingest two snapshots; normalize both to identity `shared` with differing mechanism postures; assert either (a) both attestations remain queryable as current with an explicit recorded head choice, or (b) a typed conflict record exists. Today neither exists; `approach_revisions.supersedes_revision_id` links across snapshots unconditionally.
- **Evidence class: demonstrated static path** (code path is unconditional; not executed end-to-end). Related to F1 by a shared underlying cause (identity string is the sole lineage carrier) but independently remediable, so recorded separately.
- **Severity / confidence:** medium; the design comment at `internal/store/normalize.go:212-218` declares the behavior deliberately for re-normalization, but is silent on the multi-source case, so this may be a deliberate boundary — see open question Q1.

### Limitations (declared or structural, not defects)

- **L1 — all-history access mode has no production consumer.** `ListHistoricalSignaturesForProblem` (`internal/store/canon_store.go:488-492`) is explicit and typed at the store layer but is referenced only by tests; no pipeline method or CLI flag reaches it (`internal/pipeline/app.go:68` exposes only the default reader; `cmd/newf/cluster.go:55-58` has no history flag). The audit population is currently unreachable by an operator.
- **L2 — non-forced re-normalization records no normalization-revision lineage.** A new config hash creates a new normalization revision with `supersedes = nil` (`internal/pipeline/normalize.go:211-217` sets it only under `Force`), so even the unconsumed lineage of F1 is absent in that path.
- **L3 — cluster runs do not record an access-mode label.** The population basis is always "current heads at build time", implicit rather than stated; replay integrity is nevertheless preserved because member signature IDs are pinned per run.

### Observed protections

- **P1 — head-only default population is real and executed-verified.** Two disposable tests through the public persistence path confirmed: re-interpreting one approach (changed mechanism family, stable identity) leaves exactly the head signature in the default population while history retains both; two genuinely distinct approaches contribute two members (no over-collapse).
- **P2 — no stale resurrection.** Head selection precedes the vocabulary filter; the existing test `internal/store/assessment_views_test.go:49-69` covers both the backdated-correction and missing-current-signature cases.
- **P3 — structural duplicate guards.** `UNIQUE(approach_revision_id)` on `mechanisms`, `UNIQUE(mechanism_id, schema_version, vocabulary_version)` on `mechanism_signatures`, and `UNIQUE(approach_id, normalization_revision_id)` on `approach_revisions` make per-tuple duplicate members impossible by construction; immutability triggers prevent rewriting revisions or signatures (`internal/store/migrations.go:242-251, 405-412`).
- **P4 — duplicate identities within one provider result are rejected** at schema validation (`internal/normalize/schema.go:160-163`), before persistence.
- **P5 — invariant support caps one per family** (`internal/invariant/engine.go:5-11`), which masks same-family paraphrase inflation (though not F1's cross-family case).
- **P6 — pinned replay is typed:** cluster runs persist exact member signature IDs and an `input_set_hash`, and failure spaces/invariants consume cluster runs by ID (`internal/pipeline/invariant.go:208`).

### Open questions

- **Q1:** is cross-snapshot auto-supersession (F2) intended semantics for "two sources attest the same approach"? If yes, the intended meaning of `supersedes_revision_id` should be documented as "replaces as current" rather than "corrects", and F1's remedy must not rely on it meaning correction.
- **Q2:** what is the observed rate of provider `logical_identity` drift across re-normalizations? No measurement exists; F1's practical severity depends on it.
- **Q3:** when multiple unsuperseded heads exist, the deterministic `created_at DESC, id DESC` pick (`internal/store/assessment_views.go:11-13, 29`) selects one competing interpretation silently. Should competing heads be surfaced as a degraded/conflicted population status instead?

## 4. Checks, coverage and resources

Commands executed (all inside the extracted tree; exit status in brackets):

1. `git archive 5dd08618… | tar -x` into a fresh temp dir; `docs/reviews` removed — setup [0].
2. `go build ./...` [0].
3. `go test ./internal/store/ -run 'TestRepro' -v` — 2 disposable reproduction tests (head-only population; distinct approaches stay separate): both PASS [0].
4. `go test ./internal/store/ -run 'TestProbeIdentityDrift' -v` — disposable probe: FAIL with `default population … 2 members` — the F1 executed reproduction [1, expected].
5. Disposable test files removed; pristine-tree gates: `go build ./...` [0], `go test ./...` — all 17 packages `ok` [0], `gofmt -l .` — empty [0].

Files consulted (relative to subject tree): `internal/store/canon_store.go`, `internal/store/assessment_views.go`, `internal/store/assessment_views_test.go`, `internal/store/normalize.go`, `internal/store/normalize_test.go`, `internal/store/migrations.go` (schema/trigger sections), `internal/store/interpretation_store.go` (partial), `internal/pipeline/cluster.go`, `internal/pipeline/normalize.go`, `internal/pipeline/readiness.go` (population call site), `internal/pipeline/mechanism.go` (entry points), `internal/pipeline/app.go` (store interface), `internal/pipeline/fixture.go` (partial), `internal/pipeline/invariant.go` (cluster-run pinning), `internal/invariant/engine.go` (support model), `internal/normalize/schema.go` (identity validation), `cmd/newf/cluster.go` and `cmd/newf/normalize.go` (flag surfaces), `docs/normalization.md` (identity contract), plus grep-level contact with fixtures under `fixtures/` and `testdata/fixtures/mechanism/`.

Not examined: challenge, frontier generation/admission, evaluation, success-compression, experiment, and policy subsystems beyond their population touchpoints; provider adapters; the clustering engine internals (`internal/cluster`); comparison profiles (`internal/canon`); the occurrence-result projection in the second half of `assessment_views.go`; `paper/`, `corpus/`, and remaining `docs/`. No claims are made about those areas. `ListSignaturesForProblem` callers were enumerated exhaustively; `ListHistoricalSignaturesForProblem` non-test callers were enumerated exhaustively (zero).

Resources: single pass, one reviewer session, local `go` toolchain only, isolated SQLite per test. Wall-clock estimate: ~22 minutes.

## 5. Remediation handoffs

1. **Close the identity-drift hole in default-population eligibility (F1).** Either consume normalization-revision supersession lineage in population selection (exclude/flag approach heads supported only by a superseded normalization revision of the same snapshot), or require an explicit old-identity→new-identity mapping when a re-normalization supersedes a prior revision. **Acceptance:** the F1 regression check (section 3) yields a 1-member default population or a typed, operator-visible conflict; existing tests in `internal/store/assessment_views_test.go` and the full suite stay green; invalidated if a new lineage rule resurrects superseded signatures under a vocabulary where the head is unsigned.
2. **Separate correction from independent attestation in approach-revision lineage (F2).** Auto-supersede only when the new revision re-interprets the same snapshot (or is explicitly marked a correction); otherwise retain both attestations with a recorded head choice or a typed conflict. **Acceptance:** the F2 regression check (section 3) passes; population count for one logical identity remains 1; no existing supersession test regresses.
3. **Make the three access modes operator-complete and self-describing (L1/L3).** Expose the all-history audit population through the pipeline/CLI, and record the population basis (access mode + selection rule version) on cluster runs. **Acceptance:** an operator can produce the audit population for a problem without SQL; new cluster runs carry an explicit basis field; historical runs remain readable unchanged.
