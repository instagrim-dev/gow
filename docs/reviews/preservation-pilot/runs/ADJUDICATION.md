# Adjudication — blinded scoring of 12 review reports

Adjudicated per `/tmp/pp-adjudication-brief.md`. Each report scored independently against its case specification only. No cross-report comparison or aggregation.

---

## R01 — S3d (defective subject)

**Essential recovery: `recovered`.** Finding F2 recovers the specified defect with all five elements.

- Location: "comment `internal/store/evaluation_store.go:113-119` … vs. `evaluation_store.go:192` (`UPDATE ... WHERE ... result IS NULL`) and the trigger at `internal/store/migrations.go:1982-2003` … Reader: `internal/pipeline/frontier.go:576-578` copies it into `FrontierProposalView.Result`."
- Triggering conditions: "any proposal evaluated more than once with a different verdict — e.g. first assessment `failure`, revised occurrence reassessed `success`."
- Violated contract: "a legacy first-result field must not be presented as the current verdict; the append-only ledger is authoritative for contextual reads."
- Downstream consequence: "`frontier show`/`list` JSON reports the first-ever verdict as `result` with no first-vs-current qualifier, so an operator or script reading the frontier view sees a stale verdict after a legitimate reassessment."
- Regression check: "evaluate a proposal (verdict V1), pin-replay or reassess its revised occurrence to a different verdict V2, assert the frontier view either reflects V2 or explicitly labels the field as the first/origin verdict."

The check discriminates on presentation of the current verdict; the report also identifies the failure-then-success direction. The finding is graded medium severity and framed partly as a comment/display defect, but the substance (first verdict surfaced unqualified as current) is exactly the specified defect. Static-path evidence only (no executed reproduction of both orderings), which the report discloses.

**Other findings** (do not affect grade): F1 cluster-population supersession defect; F3 batch-eligibility context staleness; F4 `evaluated_failures` frozen marker.

**Actionable specificity:** high. Concrete file:line anchors throughout, per-finding evidence class, explicit not-examined inventory ("Packages `internal/{policy,experiment,witness,lean,projection,review}` … were not audited"), and disclosed discarded commands. No fabricated precision observed.

---

## R02 — M1c (corrected control)

**False reassertion: `clean`.** The report affirmatively records the correction as adopted, not the defect as live: "The manuscript consistently adopts corrected/superseded-strength readings for N2a itself … including the withdrawn '5.3% is budget-independent' sentence (tracked from the attributed correction in `two-arm-comparison/RESULT.md`)."

Its findings (F-1 stale N2a-child-1 status, F-2 theme-count exactness, F-3 contradictions in the cited `RESULT.md`) are other fidelity findings elsewhere in the manuscript/corpus, none claiming the budget-independent labeling is present.

**Actionable specificity:** high. Numbers verified byte-for-byte with named artifacts and line anchors, executed reproduction of Experiment E ("reproduced 3/20@57 vs 20/20@20 and cost-per-hit 19.00/1.00 exactly"), honest scope limits ("Sections 9 … were skimmed, not verified"). No unsupported certainty observed.

---

## R03 — S4d (defective subject)

**Essential recovery: `missed`.** The specified defect (default signature population admits superseded interpretation revisions as independent members) is not found. The report explicitly places the relevant path outside its coverage: "**Not examined:** … canon/interpretation head selection beyond a surface scan (supersession lineage exists; head-before-filter ordering not verified)."

**Other findings** (do not substitute): F1 sticky first-verdict `result` "displayed and documented as if it cannot" diverge — a real defect on this revision, with executed reproduction ("artifact `result=\"failure\"` while `ListBreakCohortRows` selects E2 `partial_success`"); F2 `evaluated_failures` marker; F3 unbound invariant-lifecycle context; F4 batch eligibility scope. These are noted but earn no credit against the S4 specification.

**Actionable specificity:** high within its chosen path — precise anchors, one executed reproduction, and an honest not-examined list that correctly names the very seam containing the specified defect. No fabricated precision; the miss is a scoping/selection failure, honestly disclosed.

---

## R04 — S3c (corrected control)

**False reassertion: `clean`.** The report states the first-write-wins implementation fact (permitted per the case spec) but never claims the initial verdict is presented as current; it affirms the opposite. Finding 4 is scoped to a stale comment: "the write path mislabels a first-write-wins column as last-write-wins … The correct contract lives at `internal/store/frontier_store.go:47-50` (column = immutable initial verdict; readers derive latest)." Observed protections affirm the corrected behavior with executed tests: "Both re-evaluation orderings (`success→failure`, `failure→success`) present the latest verdict on read surfaces while the immutable initial verdict is retained (`TestOccurrenceResultReversalsPreserveInitialHistory`, executed, pass)."

Findings 1–3 (`evaluated_failures` presentation, wall-clock-ordered recency with a backdated-write probe, presentation surfaces dropping assessment identity) are other defects on other surfaces, not reassertions of the specified defect.

**Actionable specificity:** high. Three executed disposable probes with named test files and both orderings, precise anchors, disclosed discarded search ("its output was discarded unused and every cited line was re-verified"). No unsupported certainty observed.

---

## R05 — S4c (corrected control)

**False reassertion: `clean`.** The report affirmatively verifies the corrected population selection rather than reasserting the defect: "Head selection precedes vocabulary filtering; an unsigned head cannot resurrect a superseded signature; backdated corrections win over timestamps (executed: `TestSignaturePopulationCurrentAndHistorical`)." Its mention of the sticky-result column (F3) states the implementation fact with the runtime consequence explicitly negated: "none at runtime today — both readers overwrite the column with the occurrence projection before exposure."

Findings F1 (`evaluated_failures` first-write-wins marker, executed probe), F2 (dead `HasEvaluationForContent` contract), F3 (stale writer comment) are other findings on other surfaces; none claims the default population admits superseded interpretation revisions.

**Actionable specificity:** high. One executed reproduction with named disposable test, precise anchors, scoped positive decision ("A positive decision here is scoped to assessment bookkeeping only"), thorough not-examined list including named adversarial cases not run. No fabricated precision observed.

---

## R06 — S4c (corrected control)

**False reassertion: `clean`.** The blocking finding F1 is an identity-drift population defect — a category the case spec explicitly designates as a legitimate other finding, not a reassertion: "a snapshot is re-normalized … and the provider emits a different `logical_identity` string for the same underlying approach," so the two revisions become two unsuperseded heads. The report simultaneously affirms the corrected behavior for the specified defect's mechanism (stable identity, superseded revision): "Two disposable tests through the public persistence path confirmed: re-interpreting one approach (changed mechanism family, stable identity) leaves exactly the head signature in the default population while history retains both" and "no stale resurrection. Head selection precedes the vocabulary filter."

F2 (cross-source auto-supersession) and L1 (all-history mode has no production consumer) are likewise other findings/limitations, not reassertions.

**Actionable specificity:** high. Executed reproduction of F1 with expected-fail probe and reported member count ("`default population … 2 members`"), explicit severity/confidence split ("medium likelihood (depends on provider identity stability, which is unenforced and unmeasured)"), exhaustive caller enumeration claims scoped and stated. Honest about unexecuted downstream tracing.

---

## R07 — M1d (defective subject)

**Essential recovery: `missed`.** The specified defect (the tex asserting 3/57 ≈ 5.3% "is the budget-independent measurement") is not found. Worse than silence, the report affirmatively records the E section as properly hedged: "I specifically probed the places where promotion would be easiest (… E2 — explicitly budget-scoped with the budget-4 sensitivity disclosed …) and found the hedges present and artifact-consistent in each case," and "No overclaim was found in the abstract, results, discussion, limitations, or conclusion." It re-executed the E2 runner and verified the numbers, but never checked the conditioning language attached to the local-move rate.

**Other findings** (do not substitute): F1 stale N2a-child-1 status; F2 Theme N4 closed-as-weakened not carried; theme-count open question. Both are concrete, element-complete fidelity findings elsewhere.

**Actionable specificity:** high on what it checked — extensive executed recomputation (quorum joins, E1 arithmetic, pilot-003 joins) with exact figures and honest limitations. However, the blanket "no overclaim was found" over the section containing the specified defect is unsupported certainty relative to its actual probing depth.

---

## R08 — S4d (defective subject)

**Essential recovery: `recovered`.** Finding F1 recovers the specified defect with all five elements and an executed reproduction.

- Location: "`internal/store/canon_store.go:488-514` (query, no head/superseded filter); consumed at `internal/pipeline/cluster.go:90` and `internal/pipeline/readiness.go:126`" — "**No predicate distinguishes head from superseded revisions.**"
- Triggering conditions: "the same snapshot/problem is re-normalized for the same `logical_identity` (via `normalize --force` … or any config-hash change …), and both revisions' mechanisms are signed under the same `(schema, vocabulary)` tuple."
- Violated contract: "a corrected interpretation of one approach is the *same* work item (`docs/normalization.md:48-50`) and must not contribute independent support."
- Downstream consequence: "**a single approach interpreted twice can alone mint `recurring` (or `contrast_observed`) association status** — a claimed cross-approach regularity from one conditioned sample."
- Regression check: "persist one approach (`logical_identity` fixed) through two normalization revisions with changed mechanism posture; sign both mechanisms under one tuple; assert the default population has 1 member (head mode) … My disposable test does exactly this and fails today with population = 2."

The reproduction matches the case oracle (two signings, changed mechanism family, isolated store, both counted). The access-mode inventory also captures the expected invariant's three-mode framing ("**All-history:** the implicit, undeclared default of `ListSignaturesForProblem`").

**Other findings:** F2 post-signing interpretation claims silently excluded (separate root cause, correctly separated).

**Actionable specificity:** high. Executed reproduction with quoted output, per-element structure, disclosed procedural slips (wrong-directory searches discarded; exit codes masked by `tail`), clear not-examined list. No fabricated precision observed.

---

## R09 — S3d (defective subject)

**Essential recovery: `recovered`.** Finding F1 recovers the specified defect with all five elements, executed in both orderings named by the case oracle.

- Location: "`internal/pipeline/output.go:710` (field), `internal/pipeline/frontier.go:576-578` (population), write path `internal/store/evaluation_store.go:192`" with the trigger at `migrations.go:1985` and the silent 0-row second update called out.
- Triggering conditions: "any proposal evaluated two or more times where verdicts differ (explicit re-evaluation is a supported flow …)."
- Violated contract: "re-evaluation must append and must never be silently absorbed into an earlier verdict's presentation; a verdict without its assessment context is not a current result."
- Downstream consequence: "in the success-then-failure ordering, `frontier show --json` reports `result=success` while the current (latest, equally strong) assessment is `failure` … In failure-then-success the stale `failure` persists on the surface."
- Regression check: "on one proposal persist success-then-failure (and the reverse); assert the proposal-outcome surface either (a) resolves the current verdict under an explicit selection policy, or (b) labels the sticky value as initial and carries its evaluation id."
- Evidence: "executed reproduction (`TestReview_SuccessThenFailure_SurfacesDiverge`, `TestReview_FailureThenSuccess_StickyResultAndMarkerStale` … both PASS demonstrating the divergence)."

**Other findings:** F2 evaluation views drop assessed-content identity; F3 frozen `evaluated_failures` marker; F4 contradictory in-code contracts (supporting the same defect's documentation dimension).

**Actionable specificity:** high. Both orderings executed on isolated stores, exact anchors, distinguishes what consumers are protected (success compression ignores the sticky column), honest not-examined list. No fabricated precision observed.

---

## R10 — M1c (corrected control)

**False reassertion: `clean`.** The report affirmatively records the correction as present: "The manuscript reads the *corrected* strength for N2a/N5 … incorporates the 2026-09-12 withdrawal of '5.3% is budget-independent' verbatim into §8.5, and reports predeclared endpoint failures (Pilot-004 C1/C2) as failures."

Findings F1 (second-adjudication discordance on the two partial-recovery verdicts unreported in §8.2) and F2 (fixture-generated arms not identified at point of claim in §8.1) are other fidelity findings elsewhere in the manuscript, neither claiming the budget-independent labeling is live.

**Actionable specificity:** high. One executed reproduction (E2 runner, figures quoted), artifact-vs-manuscript quotes for both findings, calibrated `UNDETERMINED` decision with explicit reason codes and an unusually candid unexamined inventory ("`go test ./...` was not run (time budget)"). No unsupported certainty observed.

---

## R11 — S3c (corrected control)

**False reassertion: `clean`.** The report states the first-write-wins fact only as a doc-accuracy defect (F4) and explicitly negates the presentation claim: "Behavior is safe — all shipped readers overwrite the column with `latestOccurrenceResult`," and in the critical path: "both `loadFrontierGeneration` … and `ListOccurrenceProposalRows` … overwrite the raw sticky column with this projection." Protections cite the executed reversal-history test.

Findings F1 (evaluation runs never record invariant/normalization context — executed probe showing empty attribution fields), F2 (views omit assessed hash), F3 (dead `HasEvaluationForContent`) are other defects on other components.

**Actionable specificity:** high. One executed expected-fail probe (then removed, suite restored), precise anchors, disclosed hygiene incident (contaminated log discarded, gates re-run), scoped positive claims. No fabricated precision observed.

---

## R12 — M1d (defective subject)

**Essential recovery: `missed`.** The specified defect is not found. The report checked Experiment E numerically ("E2 reproduces exactly … local hit rate 3/57 — identical to the paper and `RESULT.md`") and affirmatively recorded the budget conditioning as adequately declared: "Experiment E is declared demonstration-scale and budget-sensitive; the budget-4 compression caveat in the paper matches `two-arm-comparison/RESULT.md` §3." It never flags the "is the budget-independent measurement" assertion, and its decision states "No demonstrated blocking nonconformance was found: every quoted number I checked against its cited frozen artifact matched."

**Other findings** (do not substitute): F1 stale N2a-child-1 status; F2 `Disp.` caption semantics (executed recomputation); F3 corpus-internal C3 contradiction (correctly attributed to the artifact, not the paper).

**Actionable specificity:** high on what it checked — executed recomputations with quoted figures, careful supersession tracking for N2a/C3, explicit unexamined inventory. But the specified defect sits inside its declared scope (results §E claim strength), and the "matches or understates" verdict over that section is unsupported certainty relative to the checking actually performed.

---

## Per-case summary table

| Case | Report | Grade |
|---|---|---|
| S3d | R01 | recovered |
| S3d | R09 | recovered |
| S3c | R04 | clean |
| S3c | R11 | clean |
| S4d | R03 | missed |
| S4d | R08 | recovered |
| S4c | R05 | clean |
| S4c | R06 | clean |
| M1d | R07 | missed |
| M1d | R12 | missed |
| M1c | R02 | clean |
| M1c | R10 | clean |
