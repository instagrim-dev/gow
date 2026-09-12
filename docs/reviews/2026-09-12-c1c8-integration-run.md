# Review-run: one-obligation integration (C1–C8), 2026-09-12

Executed per `docs/reviews/prompts/README.md` →
`docs/reviews/prompts/recipes/assessment-admission-decision.md` under
`docs/reviews/prompts/review-contract.md`. This file is the single evidence
bundle for the run. It records executed checks; it is not a claim of
whole-project completeness or scientific success.

## 1. Decision, policy, and scope

- **Named decision:** may the selected candidate invariant guide the next
  search action under policy P1 (recipe wording, unchanged).
- **Policy P1 (revision 1):** owner `operator:jmh`; authority: the operator's
  explicit instruction of 2026-09-12 to run the prompt-based review workflow.
  One mandatory obligation: `current-assessment-authority`, semantic revision 1
  (recipe text, unchanged). Evidence cutoff: execution time (2026-09-12,
  frozen fixture clock `2026-09-12T13:00:00Z` inside the scenario store).
  Budgets: 8 cases, ≤2 attempts per failed command after diagnosis, **zero
  paid-provider calls, zero production DB/corpus writes** — both held.
- **Pinned checkout:** worktree at commit `7e24160` **plus substantial
  uncommitted concurrent changes** (a second writer was mid-flight adding the
  typed review-record/coverage store during this run). The C8 result is
  therefore pinned to worktree state, not to a commit — recorded as a
  limitation, not disguised.
- **Contract/recipe hashes (sha256):**
  - `review-contract.md` `dcc6f4e16e449cdaf52019f9e5108655fd5a00b21d56acfd12c9808e25fc888d`
  - `assessment-admission-decision.md` `acf5da5a8a78e89ea1851ba84e46f9f145ff0aa8edd84b33be0acf76ed066caf`
- **Mode:** review-only + disposable test, per contract. The combined scenario
  was implemented as a disposable Go test against the **real migrated store**
  (`store.Open` + `Migrate` on a `t.TempDir()` SQLite file), executed, retained
  as evidence, and **deleted from the source tree**. Procedure sha256
  `f0437229d794e2697443c0b97b382578cdf98a915d425f139e0c0f7410da6cce`; full
  source retained at `docs/reviews/evidence/2026-09-12-c1c8-procedure.go.txt`.

## 2. Responsibility and critical path

Path `assessment-admission-decision` (TAXONOMY): owners
`assessment-identity` (07), `evidence-admission` (08), `decisions` (16);
`verification` (12) exercised as peer through the exact-integer witness
checker. One causal path, one session — the taxonomy was not executed wholesale.

## 3. Cases executed and findings

### Baseline component checks (recipe's executable block)

Both named tests exist (`-list` guard passed) and passed verbatim:
`TestIntegrationChallengeAssessmentPopulation` (0.24s),
`TestIntegrationEvidenceAdmissionClosesReentry` (0.20s). Repository gates ran
once on the combined tree: `go build ./...` clean, `go test ./...` 20/20
packages ok, `gofmt -l .` empty.

**Source-anchor correction (supporting observation):** the recipe's authoring
anchor (`269de0d`) recorded a limitation that "the S1 test does not by itself
assert the next current policy decision after replay." That is no longer true
at the current checkout: the population test's section 5b asserts
`ExcludedStaleAuthority` **and** that the post-replay generation request
payload omits the obsolete-authority invariant. The stale finding was not
repeated; current behavior was executed instead.

### Combined scenario C1–C8 (all observables in `evidence/2026-09-12-c1c8-run-final.log`)

| Case | Result | Key identities |
| --- | --- | --- |
| C1 baseline | PASS | candidate `inv_…HNX4ED3J` assessed `surviving` over D0 `clr_…H5H71QKD` (campaign `run_…HQHWW4B2`); decision evidenced by the persisted generation request payload naming the invariant — no global artifact-result shortcut. |
| C2 withheld control | PASS | model-judged failure `evl_…JGSDME8V` recorded, rule pass **withheld** it; population unchanged (2 signatures). |
| C3 checked admission | PASS | deliberately invalid Erdős–Straus tuple `7,2,2,2` decided by the exact-integer checker (`witness-invalid`, reproducible strength); rule pass admitted it as `domain-checked-failure` with content hash; new population A1 `clr_…KWRBNA0M` = D0 + exactly the admitted signature. **No operator-attested model judgment was substituted.** |
| C4 relevant change | PASS | assessment became stale for current selection: excluded with reason `authority_assessed_obsolete_population`, naming both A0 and A1; generation request payload no longer targets the invariant. |
| C5 unrelated control | PASS | an episode preregistration (outside the declared dependency graph) did not stale the assessment; invariant still targeted. Executed **before** the relevant change. |
| C6 reassessment | PASS | distinct campaign `run_…MCJ1HRXW` assessed against A1 → `weaken`; the A0 campaign's population row remained durable and reproducible. |
| C7 historical replay | PASS (with a recorded vacuity — see OQ-1) | replay `run_…MX8R3QC4` assessed A0, re-earned `surviving` for its bounded claim; the **current** decision did not restore obsolete authority (stale exclusion; payload does not target). All three campaign population rows durable and distinct. |
| C8 derived projection | PASS (against in-flight surface — see L-1) | policy/applicability/check/assessment recorded through the four-record store; coverage generated twice from identical records → byte-identical document, decision `ELIGIBLE_TO_ADVANCE`; regenerated after a relevant dependency change → `UNDETERMINED [stale_dependency]`, distinguishable from unexamined. |

### Findings (classified per contract)

- **F-1 — demonstrated static path, reviewer-discovered, primary owner
  `evidence-admission` (08).** `evaluated_failures` is keyed per proposal with
  `INSERT OR IGNORE` (`internal/store/evaluation_store.go`, R6 marker write).
  A proposal whose first failure was model-judged (withheld) keeps that first
  marker forever: a **later witness-checked (reproducible-strength) failure of
  the same proposal never reaches `AdmitEvidence`** — the rule pass sees only
  the old evaluation and skips it as already decided. Stronger evidence is
  silently shadowed by a weaker earlier marker. Demonstrated during scenario
  construction (retained: `evidence/2026-09-12-c1c8-run-f3-shadowing.log`,
  where the witness evaluation is absent from the admission response); the
  scenario then used distinct proposals for C2/C3. Expected: a
  stronger-verdict evaluation should become visible to admission (marker per
  evaluation, or upgrade-on-stronger-verdict). Severity: moderate (blocks
  epistemic upgrade of real observations); confidence: high (reproduced).
  Discriminating regression check: witness-check a proposal that already has a
  withheld model-judged failure; assert the rule pass admits the
  domain-checked failure.
- **OQ-1 — open question (untested hypothesis), owner `decisions` (16) with
  `assessment-identity` (07).** C7's positive half — "the next current
  decision still selects the **compatible current** context, not the last
  execution" — was satisfied only vacuously here, because C6 ended `weaken`
  (non-targetable), leaving nothing compatible to select. The current-state
  authority join selects the campaign behind the **latest state transition**
  (`invariantStateAuthorityJoin`), and every campaign writes transitions; so a
  replay executed **after** a compatible current `surviving` assessment would
  take latest-transition authority with an obsolete population and could
  displace a live current authority into stale exclusion. Not asserted a
  defect — not demonstrated. Discriminating check: variant scenario where C6
  ends `surviving` over A1, then replay A0, then assert the invariant is
  still targeted.
- **L-1 — declared limitation.** C8 was exercised against the concurrent
  writer's **uncommitted** four-record/coverage implementation
  (`internal/store/review_store.go`, `review_coverage_store.go`,
  `internal/pipeline/review.go`, `review_coverage.go`, `internal/review/`).
  The behavior observed is exactly the contract's projection (three-valued
  decision, six reason codes, deterministic export, `stale_dependency`
  distinguishable), but the result is pinned to worktree state and must be
  re-confirmed once that slice lands at a commit.
- **P-1 — protections observed (supporting).** The withheld control (C2), the
  narrow staleness gate (C4/C5 discriminate relevant from unrelated change),
  durable per-campaign population identity (C6/C7), and the payload-level
  decision boundary (targeting asserted on the persisted provider request, not
  on a response label).

## 4. Checks, coverage, and resources

- **Executed:** 2 component tests; 1 full gates run (`go build ./...`,
  `go test ./...` 20/20 ok, `gofmt -l .` empty); 6 executions of the combined
  scenario (4 failed runs retained: 2 concurrent-writer build breaks, 1
  fixture-generator gap, 1 F-1 manifestation, 1 obligation-key error; final
  run green). Zero paid-provider calls; zero production DB/corpus writes; all
  scenario writes on a `t.TempDir()` store.
- **Blockers encountered:** transient test-build break from the concurrent
  writer's in-flight `problemStore.GetReviewPolicy` interface change; resolved
  forward (their own fake stub landed; my temporary duplicate was removed).
  Both attempts retained per budget.
- **Coverage statement (scoped):** this run supports **review completion and
  conformance for the one obligation only**. It is not taxonomic coverage, not
  whole-project completeness, and not scientific success. The C8 coverage
  export proves that one obligation's projection derivability.
- **Decision projection (this run's named decision):** `ELIGIBLE_TO_ADVANCE`
  under P1 as pinned — every mandatory obligation (one) has current
  conformance support from the executed scenario, with F-1 recorded against a
  peer obligation, OQ-1 as an open question on an undemonstrated branch, and
  L-1 as a revision-pinning limitation. None is an unresolved demonstrated
  blocking nonconformance for this decision.

## 5. Remediation handoffs (≤3)

1. **F-1 (`evidence-admission`):** make stronger-verdict evaluations of an
   already-marked proposal visible to admission (marker per evaluation or
   verdict upgrade). Invalidation: any change to `evaluated_failures` keying.
   Acceptance: the discriminating regression check in F-1 passes.
2. **OQ-1 (`decisions`):** add the surviving-C6-then-replay variant as a
   committed integration case asserting the compatible current authority is
   selected over the last-executed obsolete campaign. Acceptance: the variant
   passes, or the demonstrated failure becomes a new finding with this bundle
   as provenance.
3. **L-1 (owner of the in-flight review slice):** land and pin the
   four-record/coverage implementation, then re-run C8 (this bundle's
   procedure, `evidence/2026-09-12-c1c8-procedure.go.txt`) against the pinned
   commit. Acceptance: identical projection behavior at a named revision.

Only after this slice's handoffs settle should the preservation pilot
(`recipes/preservation-pilot.md`) proceed, per the recipe's gating.
