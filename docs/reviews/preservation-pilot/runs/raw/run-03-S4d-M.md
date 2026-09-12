# Review report — population selection and interpretation revisions

Reviewed revision: `3443a1094407364be2e02264ffa1ee1bb0f3efde`

## 1. Decision, policy and scope

**Decision under review:** when one underlying work item (a logical approach) has multiple interpretation revisions, which revisions are eligible for the default signature population that feeds clustering, failure spaces, readiness gating, and invariant support — and whether that selection is explicit and typed.

**Decision: `WITHHOLD`.** At least one demonstrated, unresolved blocking nonconformance applies (finding F1, executed reproduction): superseded interpretation revisions of a single approach enter the default population as independent members, and with a changed mechanism family this moves invariant-support figures across the `recurring` threshold from one work item. Remaining uncertainties are listed in section 3.

**Policy provenance:** no versioned decision policy naming an accountable owner for this population-eligibility decision was found inside the subject tree. The obligations applied are: (a) a population is a conditioned selection, not "all rows matching a filter"; corrected interpretations of one approach must not count as independent support; (b) three access modes — current interpretation heads, pinned historical replay, all-history — must each be explicit and typed. These obligations are consistent with the tree's own doctrine (`AGENTS.md`: conditioned populations for claimed regularities; no silent epistemic promotion) and its documented data model (`docs/normalization.md:48-50`: re-normalizing attaches new revisions to the *same* logical approach rather than duplicating it).

**Scope:** SQLite store layer (`internal/store`), the population consumers (`internal/pipeline/cluster.go`, `internal/pipeline/readiness.go`, `internal/pipeline/invariant.go`, `internal/pipeline/mechanism.go`, `internal/pipeline/interpretation.go`), the pure invariant engine (`internal/invariant/engine.go`), and the governing docs (`docs/normalization.md`, `docs/mechanism-clustering.md`, `docs/invariant-mining.md`). Isolated disposable tests executed against a temp-dir SQLite store. Not in scope: experiment/holdout population, success compression, challenge campaign, CLI wiring depth (section 4).

## 2. Responsibility and critical path

The consequential causal path for this decision:

1. **Identity creation** — `internal/store/normalize.go:311-350` (`getOrCreateApproachTx`): approaches are keyed `(problem_id, logical_identity)`; re-normalization reuses the identity and derives `supersedes_revision_id` lineage in-transaction (`normalize.go:206-233`). Lineage is durable and correct.
2. **Revision multiplication** — a forced re-normalization (`internal/pipeline/normalize.go:36` `Force`, applied at `:212-230`) or a changed config hash (`internal/store/normalize.go:64-83`, `FindEquivalentNormalization` keys only on snapshot+schema+config) creates a second `approach_revision`, each with its **own mechanism row**.
3. **Signing** — `internal/pipeline/mechanism.go:70-135` builds a signature per *mechanism* under `(schema_version, vocabulary_version)`; `internal/store/canon_store.go:285-306` persists idempotently per `(mechanism_id, schema_version, vocabulary_version)`. Both revisions' mechanisms can be signed under one tuple.
4. **Population selection (critical seam)** — `internal/store/canon_store.go:488-514` `ListSignaturesForProblem` joins `mechanism_signatures → mechanisms → approach_revisions → approaches` filtered only by `problem_id` + version tuple. **No predicate distinguishes head from superseded revisions.**
5. **Consumers** — `internal/pipeline/cluster.go:90` (cluster build population), `internal/pipeline/readiness.go:126` (readiness counters); families flow into failure spaces (`cluster.go:377-469`) and invariant mining (`internal/pipeline/invariant.go:61-94` rehydrates persisted cluster members).
6. **Support accounting** — `internal/invariant/engine.go:283-308` caps support at one per distinct *family* (`ClusterID`), and `classify` (`engine.go:347-354`) mints `recurring`/`contrast_observed` at `minSupport` (default 2, `internal/pipeline/invariant.go:177`). Nothing anywhere dedupes by *approach*.

Responsibility: the store owns population semantics (step 4); the pipeline owns declaring which access mode a run used (steps 5-6). Both are silent today.

## 3. Findings, limitations, protections, questions

### F1 — Superseded interpretation revisions count as independent population members (blocking)

- **Location:** `internal/store/canon_store.go:488-514` (query, no head/superseded filter); consumed at `internal/pipeline/cluster.go:90` and `internal/pipeline/readiness.go:126`.
- **Triggering conditions:** the same snapshot/problem is re-normalized for the same `logical_identity` (via `normalize --force`, `internal/pipeline/normalize.go:212-230`, or any config-hash change defeating the idempotency key at `internal/store/normalize.go:64-83`), and both revisions' mechanisms are signed under the same `(schema, vocabulary)` tuple.
- **Violated contract:** the population supporting invariants must be a conditioned selection; a corrected interpretation of one approach is the *same* work item (`docs/normalization.md:48-50`) and must not contribute independent support. `docs/mechanism-clustering.md:22-24` defines the clustered population as "the mechanism signatures of one problem, under one `(schema_version, vocabulary_version)`" with no revision conditioning — the doc contract itself omits the condition.
- **Downstream consequence:** when the correction changes the mechanism family (exactly the epistemically significant case), the two signatures land in different families: family counts in the failure space inflate; readiness `failure_cohort` (`readiness.go:90-115`) counts both toward the mining threshold; `DistinctFamilySupport` (`engine.go:283-308`) counts one approach twice; with `defaultMinSupport = 2` (`internal/pipeline/invariant.go:177`) **a single approach interpreted twice can alone mint `recurring` (or `contrast_observed`) association status** — a claimed cross-approach regularity from one conditioned sample. `FailureCoverageNum/Den`, `ContrastEligibleDen`, and coverage axes move likewise.
- **Discriminating regression check:** isolated store; persist one approach (`logical_identity` fixed) through two normalization revisions with changed mechanism posture; sign both mechanisms under one tuple; assert the default population has 1 member (head mode) or that the selection carries an explicit typed access mode. My disposable test does exactly this and fails today with population = 2.
- **Evidence class: executed reproduction.** Test `TestDefaultPopulationCountsSupersededInterpretationRevisions` (disposable, in the extracted tree only) — output: `default population members for one approach with two interpretation revisions: 2`.

### F2 — Post-signing interpretation claims are silently excluded from the population (separate root cause)

- **Location:** `internal/store/canon_store.go:296-306` (`PersistSignature` returns the existing record unchanged for a known `(mechanism, schema, vocab)`); `internal/pipeline/mechanism.go:96-112` (adjudicated interpretation claims merge only at signature *build* time); `internal/pipeline/interpretation.go:78-141` (`AddInterpretation` performs no staleness marking or warning against existing signatures).
- **Triggering conditions:** an operator adjudicates an interpretation claim (`interpretation add`) for a mechanism after `mechanism signature` has already run for that mechanism under the tuple.
- **Violated contract:** the "current interpretation heads" access mode must be explicit; here the *current* adjudicated interpretation state and the *persisted* signature silently diverge — a correction becomes a no-op for every downstream population without any recorded marker.
- **Downstream consequence:** clustering, readiness counters, and invariant support keep evaluating pre-correction claims indefinitely; the epistemic state of the population is neither head-current nor labeled historical.
- **Discriminating regression check:** sign a mechanism; add an interpretation claim; re-run signature build; assert the response either creates a successor signature identity, marks the prior one stale for current-decision use, or emits an explicit staleness warning — today it returns `status: existing` with the claim absent.
- **Evidence class: demonstrated static path** (both code paths read end-to-end; not executed as one flow).

F1 and F2 are independently remediable: F1's remedy is an explicit revision-conditioned population selector; F2's is staleness signaling between interpretation adjudication and persisted signatures. They share the manifestation "population is not the current-heads view."

### Access-mode inventory (obligation (b))

- **Current interpretation heads:** absent for populations. Head semantics exist only for display reads (`internal/store/normalize.go:410-467` `ListApproaches` / `GetApproachDetail` use latest revision).
- **Pinned historical replay:** present and sound downstream — cluster runs persist exact members and an `InputSetHash` over the exact clustered population (`internal/pipeline/cluster.go:206-270`; `docs/mechanism-clustering.md:66-70`), and invariant mining rehydrates from the *persisted* cluster rows, not a fresh query (`internal/pipeline/invariant.go:61-94`).
- **All-history:** the implicit, undeclared default of `ListSignaturesForProblem`. No caller or record states this choice.

### Limitations

- The full `cluster build → failure-space build → invariant mine` pipeline was not executed over the reproduction store; the support-figure movement in F1 is traced statically from the executed 2-member population through the engine's family-cap code.
- Two early code searches executed against an unintended directory outside the subject tree because of a lost shell variable; their outputs were discarded unread beyond noticing the error, and both searches were re-run inside the subject tree before use.
- The experiment/holdout target population (`ListTargetSignaturesForHoldout`, `internal/pipeline/readiness.go:170+`) was not examined for the same conditioning issue.
- Test exit codes piped through `tail` masked shell-level exit status in two invocations; failure/success was read from the test output itself.

### Observed protections

- **Distinct-family support cap** (`internal/invariant/engine.go:283-290`): intra-family paraphrases — including a superseded revision whose signature stays mechanism-near its successor — cannot inflate support. The defect surfaces only when the correction crosses a family boundary.
- **Pinned replay integrity:** `InputSetHash` + persisted member rows make historical cluster runs replayable and immune to later population drift.
- **Stable approach identity with durable lineage:** `(problem_id, logical_identity)` uniqueness, in-transaction supersedes derivation, and a race-safe duplicate-identity recovery (`internal/store/normalize.go:311-350`), covered by `internal/store/normalize_test.go:152-204`.
- **Audit surface exists:** `ListMechanismsForProblem` exposes `LogicalIdentity` + per-mechanism `SignatureCount` (`internal/store/normalize.go:508-549`), and `ListInterpretationClaimsForProblem` joins claims to logical identity (`internal/store/interpretation_store.go:112-142`), so the double-membership is *discoverable* by hand even though no population path uses it.
- **Immutability discipline:** signatures, revisions, and cluster runs are insert-only with `RAISE(ABORT)` triggers (`internal/store/migrations.go:405-412`); nothing here rewrites history.

### Open questions

- Is all-history ever the intended mining default (treating each interpretation as an independent *reading* rather than an independent *approach*)? If so it must be declared, and support figures per-approach-deduped anyway.
- Should a superseded revision's signature be excluded from head-mode populations, or retained with a `superseded` marker so both projections stay derivable?
- How does population conditioning interact with vocabulary succession (readiness parameterizes `vocabVersion` at `readiness.go:118-126`) — do version tuples partition revisions cleanly in real corpora, or can one approach's revisions span tuples and evade even per-tuple analysis?

## 4. Checks, coverage and resources

**Commands executed (inside the extracted subject tree unless noted):**

| Command | Result |
| --- | --- |
| `git archive <SHA> \| tar -x` into temp dir; remove excluded subtree | exit 0 |
| `ls` / `rg` inventory and symbol searches (~8 invocations) | exit 0 (one `rg` exit 1 = no matches; one failed `cd` exit 1, retried) |
| `go build ./...` | exit 0 |
| `go test ./internal/store/ -run TestDefaultPopulationCountsSupersededInterpretationRevisions -v` | test FAIL — intended reproduction: population = 2 for one approach (first attempt FK-failed until vocabulary seeded; then reproduced) |
| `go vet ./internal/store/` | exit 0 |
| `gofmt -l internal/store/` | clean |
| `go test ./internal/store/` (full package) | 1 failure — only the reproduction test; baseline green |
| `go test ./internal/invariant/ ./internal/cluster/` | ok / ok |

**Files consulted (relative to subject tree):** `internal/store/canon_store.go` (full), `internal/store/normalize.go` (full), `internal/store/interpretation_store.go` (full), `internal/store/normalize_test.go` (full), `internal/store/canon_store_test.go` (partial), `internal/store/migrations.go` (targeted sections), `internal/pipeline/cluster.go` (full), `internal/pipeline/invariant.go` (lines 1-190 + targeted greps), `internal/pipeline/readiness.go` (lines 90-180), `internal/pipeline/mechanism.go` (lines 1-135), `internal/pipeline/interpretation.go` (lines 78-145), `internal/invariant/engine.go` (full), `docs/normalization.md`, `docs/mechanism-clustering.md`, `docs/invariant-mining.md` (targeted sections), `AGENTS.md`, plus one disposable test file authored in `internal/store/`.

**Not examined:** `internal/experiment`, `internal/success`, `internal/verify`, `internal/policy`, `internal/frontier` (beyond one grep), `internal/relational`, `internal/config`, `cmd/newf` command wiring, challenge/evaluation stores, `corpus/`, `paper/`, `fixtures/`, remaining docs.

**Wall-clock estimate:** ~24 minutes (setup 3, exploration 10, reproduction 5, report 6).

## 5. Remediation handoffs

1. **Explicit access mode on the population selector.** Extend `ListSignaturesForProblem` (or add a sibling) with a typed mode: `current_heads` (only signatures whose mechanism belongs to each approach's latest / non-superseded revision) vs `all_history`; default cluster build and readiness to `current_heads`. **Acceptance:** the F1 regression check passes (population = 1 in head mode for one approach with two revisions); `all_history` remains available and must be requested by name; no existing green test regresses.
2. **Record the population basis in cluster-run provenance.** Persist the access mode and revision-selection basis on `cluster_runs` (feeding `InputSetHash` inputs) so failure spaces and invariant revisions inherit an explicit conditioning statement. **Acceptance:** a new cluster run's record names its access mode; two runs over the same corpus differing only in mode produce distinct input-set hashes; historical runs remain readable unchanged.
3. **Staleness signaling between interpretation adjudication and signatures (F2).** When `interpretation add` targets a mechanism that already has a signature under any tuple, either mark that signature stale-for-current-decisions or return an explicit warning naming the affected signature ids; `mechanism signature` on a stale pair must not silently answer `existing`. **Acceptance:** the F2 regression check passes at the `AddInterpretation` boundary; no signature row is mutated (immutability preserved); readiness output distinguishes stale-backed counts or excludes them with a stated reason.
