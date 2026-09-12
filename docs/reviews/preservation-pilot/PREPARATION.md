# Preservation pilot — frozen preparation package (revision 1)

**Status: CEILINGS DECLARED — TWELVE RUNS AUTHORIZED (authorization
revision 1, §7).** Per
[`recipes/preservation-pilot.md`](../prompts/recipes/preservation-pilot.md),
preparation was frozen first with §7 unfilled (zero dispatches); the owner
then declared ceilings on 2026-09-12. Exactly the twelve runs of §4 plus one
blinded adjudication pass are authorized — no retries, no sanity runs, no
case or prompt refinement after outputs are observed.

**Prerequisite confirmed.** The integrated obligation slice has a retained
passing assessment: run 3 of the C1–C8 scenario
([report](../2026-09-12-c1c8-integration-run.md) §0, revision 3, commit
`0c4a75a`) recorded a derived `conforms` and projected
`ELIGIBLE_TO_ADVANCE` (scoped) on the remediated tree base `9b901f7`.

**Frozen inputs (sha256, first 16 hex).**

| Input | Hash prefix |
| --- | --- |
| Protocol `recipes/preservation-pilot.md` | `5dc450c15aa1c9ce` (post status-header update in the same commit; body unchanged) |
| Shared contract `review-contract.md` | `dcc6f4e16e449cda` |
| Baseline engineering prompt `assessment-identity-and-derived-views.md` (rev 2) | `314d10c7fa2d622e` |
| Baseline engineering prompt `typed-projection-and-evidence-admission.md` (rev 2) | `7a7f7621da7fa72a` |
| Authored manuscript baseline `arms/manuscript-baseline.md` | `4009a54a471c4a65` |
| Modular case bundles `arms/modular-case-bundles.md` | `f38b32b02d3d1089` |

## 1. Cases (three defective + three matched corrected controls)

Historical findings are **case-selection leads, not current truth**. Each
defective/corrected pair below was validated against its exact revision on
2026-09-12 (validation commands in §2); no common pre-fix commit was guessed.

| Case | Defect | Defective rev | Corrected rev | Expected invariant | Reproduction oracle |
| --- | --- | --- | --- | --- | --- |
| **S3d / S3c** | First-verdict/current-view confusion: `frontier_proposals.result` is first-write-wins (`… WHERE result IS NULL`) while presented as current | `3443a10` | `5dd0861` | A reader of "current result" must receive the verdict computed against the current generation/content/problem/cluster context, never the immutable initial verdict presented unqualified | Success-then-failure and failure-then-success sequences on an isolated SQLite store: defective rev surfaces the FIRST verdict as current; corrected rev derives the current result from the evaluation ledger (`internal/store/assessment_views.go`, absent at `3443a10`) and re-documents `result` as the immutable initial verdict |
| **S4d / S4c** | Historical interpretation revisions inside the default atlas population: `ListSignaturesForProblem` admits every revision of the same approach simultaneously | `3443a10` | `5dd0861` | Population selection must distinguish current interpretation heads from historical replay and all-history access; corrected interpretations must not inflate apparent support | Sign the same approach twice with a changed mechanism family on an isolated store: defective rev counts both members in the default population; corrected rev selects current approach heads (`ListSignaturesForProblem`) and gates all-history behind `ListHistoricalSignaturesForProblem` |
| **M1d / M1c** | Manuscript-fidelity overclaim: the Experiment E local hit rate `3/57 ≈ 5.3%` asserted as "the budget-independent measurement" when it is conditioned on the frozen move ordering, budget B=3, and stop-on-hit truncation | `00c3c58` | `630e380` | A reported statistic must carry its conditioning; a budget/ordering/truncation-dependent rate must not be labeled budget-independent | `paper/geometry-of-work.tex` line ~1495: defective rev contains "is the budget-independent measurement"; corrected rev states the conditioning ("conditioned on the frozen move ordering, the budget, and stop-on-hit truncation — … not budget-independent") |

Why each correction resolves its defect: S3/S4 — commit `5dd0861` adds
derived, context-scoped current-result views and current-heads population
selection while retaining explicit historical access (the defect was
unqualified currency, not the existence of history). M1 — commit `630e380`
withdraws the unconditional label and states the exact conditioning in place
(pre-freeze manuscript correction; the frozen `RESULT.md` was corrected by
additive attributed note, preserved).

**Case-source discipline.** The original structural review
(`docs/reviews/2026-09-12-structural-semantic-epistemic-review.md`) is a case
source, **not** a prompt and **not** an input to any arm. The actual
manuscript-review instruction was operator-supplied in-session and is not
retained; per protocol, an explicit manuscript baseline is **authored and
frozen** at [`arms/manuscript-baseline.md`](arms/manuscript-baseline.md).

## 2. Per-revision validation (executed during preparation)

```text
S3d  git show 3443a10:internal/store/evaluation_store.go | grep 'result IS NULL'   → present (first-write-wins)
S3c  git ls-tree 3443a10 internal/store/ | grep -c assessment_views                → 0 (no derived views)
     git ls-tree 5dd0861 internal/store/ | grep assessment_views                   → assessment_views.go + test present
     (the IS NULL write REMAINS at 5dd0861 by design: re-documented immutable INITIAL verdict; the correction is the derived current view)
S4d  git show 3443a10:internal/store/canon_store.go → ListSignaturesForProblem without current-heads selection
S4c  git show 5dd0861:internal/store/canon_store.go → current-heads ListSignaturesForProblem + explicit ListHistoricalSignaturesForProblem
M1d  git show 00c3c58:paper/geometry-of-work.tex | grep 'budget-independent'       → "is the budget-independent measurement"
M1c  git show 630e380:paper/geometry-of-work.tex | grep 'budget-independent'      → "… but not budget-independent"
```

S3's corrected control has a subtlety adjudicators must know: the
first-write-wins UPDATE is **still present** at `5dd0861`. Reasserting "the
UPDATE is first-write-wins" against S3c is TRUE but is **not** the specified
defect; the specified defect is the *unqualified presentation of the initial
verdict as current*, which `5dd0861` resolves. A false-reassertion judgment
must be made against the specified defect, not the surviving (now-documented)
write pattern.

## 3. Arms

| Arm | Information and procedure |
| --- | --- |
| **Baseline (B)** | Bespoke project-specific prompt. Engineering cases (S3/S4): `assessment-identity-and-derived-views.md` rev 2 verbatim. Manuscript case (M1): the authored frozen `arms/manuscript-baseline.md`. |
| **Modular (M)** | `review-contract.md` (shared) + the pinned per-case bundle from `arms/modular-case-bundles.md`, carrying the same substantive obligations, examples, symbols, and evidence access as the baseline. Retrieval and contract tokens are charged to this arm. |

The independent variable is **packaging and retrieval/composition overhead
only**. Both arms receive: checkout access at the case's pinned revision,
`AGENTS.md`, and the case's pinned relevant file set. Both arms are
**denied**: remedial commit messages (any commit after the case's pinned
revision), all of `docs/reviews/` (answer-containing reports), the evidence
supplement manifest annotations that postdate the pinned revision, and this
preparation package. Same model/version, tools, settings, token-accounting
policy for both arms (recorded at dispatch; §7).

## 4. Run matrix and frozen dispatch order

Twelve isolated runs (6 cases × 2 arms, one run per cell). Paired defective
and corrected cases never share a context; every run gets a fresh context
with no prior outputs and no decision discussion.

Dispatch order was fixed by seeded shuffle
(`sha256("preservation-pilot-r1-2026-09-12")` =
`9f8065e81c817f2b839b12c25bd26a39697e8154281d531fd173896eec0db506`, Python
`random.Random(seed).shuffle`):

```text
1:S3d-B  2:S3d-M  3:S4d-M  4:S3c-B  5:S4c-M  6:S4d-B
7:M1d-B  8:M1c-M  9:M1c-B  10:S4c-B  11:M1d-M  12:S3c-M
```

Aborted runs are retained and reported; no retry-until-success.

## 5. Blinding and adjudication

Each run's report is stripped of arm-identifying headers and assigned an
opaque label `R01…R12` by a person/process other than the adjudicator; the
mapping is sealed until scoring is complete. The adjudicator receives: the
case's specified defect, expected invariant, reproduction oracle, the §2
S3-subtlety note, and the rubric below — never the arm identity or token
accounting (costs are reported separately after unblinding).

## 6. Rubric and gate (from the protocol, applied per run)

Actionable-recovery credit requires ALL of: actionable location; triggering
conditions; violated contract; downstream consequence; a discriminating
regression check. Generic warnings score zero. The discriminating
observations, judged per arm across its six runs:

1. false reassertion of the specified defect on a corrected control —
   **any occurrence rejects the arm's packaging for this trial**;
2. a missing essential recovery on a defective case — **rejects likewise**;
3. preservation of actionable specificity and honest unresolved cases;
4. retrieval/execution/reconciliation cost (economics reported separately,
   §8; never mixed into the quality gate).

A pass records feasibility on these six cases and uncertainty. It does not
establish generality, an error-rate advantage, or cost noninferiority. No new
domain-prompt series follows automatically.

## 7. Resource ceilings — DECLARED (authorization revision 1, 2026-09-12)

Owner declaration received 2026-09-12 (operator, in-session). Ceilings are
deliberately simple so they cannot become a second treatment variable.

```text
total model-call ceiling:   13 agent dispatches: exactly 12 primary review runs
                            + exactly 1 blinded adjudication pass. No retries,
                            no continuation/resume calls, no sanity runs.
total token ceiling:        enforced structurally, identically across arms:
                            one dispatch per run, one report file per run,
                            report capped at 2,500 words. Fine-grained token
                            metering is not exposed by the execution harness;
                            declared limitation. Proxy accounting retained:
                            wall time + report word count per run.
total currency ceiling:     zero external paid API calls. All 13 dispatches
                            execute inside the operator's existing session;
                            the hard dispatch count above IS the currency
                            ceiling (derived, not an independent soft limit).
per-run limits:             exactly 1 isolated agent dispatch; fixed max
                            report output (2,500 words); 30-minute wall
                            timeout; an incomplete run is retained as aborted
                            and never redispatched. The run is agentic (an
                            internal tool loop); internal loop calls are not
                            separately meterable — declared limitation,
                            identical across arms.
human adjudication budget:  one blinded adjudication pass over all 12
                            reports by a FRESH isolated agent that receives
                            only blinded reports + case specs + rubric.
                            Predeclared tie/ambiguity procedure: (a) a
                            false-reassertion judgment requires an
                            affirmative assertion of the SPECIFIED defect on
                            a corrected control, not a mere mention or
                            hedged question; (b) an essential-recovery
                            judgment that cannot be made decisively is
                            recorded "ambiguous" and does NOT credit the arm
                            (severity favors rejection of unclear recovery,
                            never rejection-by-unclear-reassertion).
environment controls:       identical across arms: same session model for
                            every dispatch ("inherit"); same toolset; single
                            self-contained prompt with the frozen packaging
                            content pasted VERBATIM; temperature/seed not
                            exposed by the harness (declared limitation,
                            symmetric). Case checkout rule: first action is
                            `git archive <REV> | tar -x -C <fresh tmpdir>`;
                            the extracted tree (no .git, hence no commit
                            messages) is the entire review subject; nothing
                            under the live working tree may be read
                            afterwards. Denial list (docs/reviews/, post-
                            revision commits, remediation notes, this
                            package) OVERRIDES any instruction embedded in a
                            packaging (the bespoke engineering prompt
                            internally cites an answer-containing fix report;
                            that citation is inert under this control).
                            Uniform report envelope for blinding: both arms
                            emit the same five section headers, must not
                            name/quote their packaging, and must list every
                            file consulted and every command executed.
reuse horizon H:            H ∈ {5, 10, 20} comparable future reviews
                            (sensitivity range), H_decision = 10 — a
                            planning estimate chosen independently of pilot
                            results, before any run.
```

**Dispatch is now authorized for exactly the twelve runs of §4, in the
frozen order, plus one adjudication pass — nothing else.** Case refinement,
prompt edits after observing outputs, and repetitions remain unauthorized;
any packaging revision starts a new protocol revision.

**Reporting interpretation note (predeclared):** the seeded dispatch order
is reproducibly shuffled, not statistically randomized in any averaging
sense — with one observation per cell it controls obvious sequencing effects
only. Results reporting must carry this interpretation.

## 8. Cost model (skeleton, rates to be fixed BEFORE results)

```text
C_M(H) = S_M + H·r_M        C_B(H) = S_B + H·r_B
H* = ceil((S_M − S_B)/(r_B − r_M))   when S_M−S_B > 0 and r_B−r_M > 0
```

- `S_M` (future one-time modular setup): case-bundle authoring is COUNTED
  here (it is adapter authoring, not runtime); the bundle in this package is
  part of `S_M`, recorded as preparation effort.
- `S_B`: manuscript-baseline authoring is counted to B's setup (it did not
  previously exist as a frozen artifact).
- `r_M`, `r_B`: recurring per-review cost including amortizable maintenance
  assumptions — estimated from the pilot's token/effort accounting.
- Sunk investment in the existing bespoke prompts is recorded separately and
  excluded from the adoption comparison (future avoidable costs only).
- The cost of running this comparison is reported even if excluded as sunk.
- Conversion of human effort and compute into one unit requires rates fixed
  before results; none are fixed yet. An unknown H or unsupported cost model
  leaves the economic assessment **inconclusive** by rule.

## 9. What follows a pass (not authorized here)

Only after preservation passes may a separately authorized transfer
experiment freeze prompts and expose genuinely held-out fault variants and
clean controls. Preservation cases must not be reused as unseen discovery
evidence; this pilot's output must not tune the transfer test.
