# Screen execution record: implementer-authored/v1 pack, runs 1–2

- **Pack:** `internal/sealedrun/pack_agent_sealed_v1.go` + `pack_controls_v1.go`, pre-committed at `8182e13` before first execution (post-run edits inadmissible; none occurred)
- **Evidence grade:** implementer-authored/v1 — the agent-sealed tier was not achieved (three custodian agent launches died on host infrastructure before authoring anything; transcripts retained). Grade meaning and mitigations in the pack's provenance header. **This is a screen-machinery and screen-protocol result, not a shaping-value verdict**
- **Arms (frozen before pack authoring):** H0 catalog-blind; H1 `h1-frequency/0` (ungated); HG `shape-selector/0` (relevance-gated)
- **Ceilings:** deterministic, zero provider spend, r=1 disclosed

## Run 1 — budget 500: ceiling effect, inconclusive (recorded, not discarded)

H0=H1=H1=21 identical; every completable episode completed under every ordering; scarcity never bound, so the only causal channel (rule order) could not matter. Roadmap disposition applied verbatim: *all arms hit the ceiling → screen inconclusive; repair the measurement, not the narrative.* Root cause: the pre-sealing sensitivity calibration step (handoff §7) was skipped by the implementer-custodian. Recorded as a protocol defect caught by the instrument.

## Calibration — arm-blind rule, stated before computation

`CalibrateH0MinBudgets`: per-episode minimum budget at which the **blind H0 arm** completes (binary search; completion monotone in budget; consults only catalog order — it cannot be steered by HG/H1 behavior it never computes, and run 1 observed zero inter-arm signal, so no directional knowledge existed to steer with). Result: 21 H0-completable episodes, minimums 1–72, **median 2** → run-2 budget = 2, applied as stated.

## Run 2 — calibrated budget 2: differentiated measurement

| Arm | Completions |
|---|---|
| H0 (blind) | **14** |
| H1 (ungated history) | 12 |
| HG (gated history) | 13 |

| Condition | Result | Reading |
|---|---|---|
| (a) no invalid certifications | pass | — |
| (b) HG ≥ H1 + 3 | **fail** (13 vs 15; margin +1) | the relevance gate beats ungated history, but by a third of the required margin |
| (c) control-strata loss ≤ 1 | pass (0) | HG never lost control episodes *relative to H1* |
| (d) ≥ 2 informative families | **fail** (1: fam-negsub) | the gate's win is concentrated, not diversified |
| (e) HG ≥ H0 | **fail** (13 vs 14) | **the H0 guard fired: history was net-negative overall** |

Episode-level signal (full traces in the committed test log): informative fam-negsub separated HG from both baselines (inf-01: H1+HG complete where H0 fails; inf-04: **only HG** completes); the blatant misleading episodes (mis-01/02) did exactly what the stratum is for — H0 completes, both history arms fail, because the lying history passes the v0 similarity gate and demotes the true reducers.

## Dispositions (roadmap §4 G4-lite / §8, applied verbatim)

1. **Spending rule: NOT satisfied.** No next shaping tranche is funded on this evidence. (At this evidence grade the rule could not have funded one anyway; the point is the instrument now produces real inputs.)
2. **H0-guard row applies:** *history handling is net-negative in this construction; repair history use before crediting or debiting shaping.* The concrete, localized defect: the v0 similarity gate admits structurally-lookalike lying histories (mis-01/02) at full weight. Candidate repairs for a **new controller version** (v0 stays frozen): failure-aware relevance weighting, or distrusting successes whose credited rules never reduce the current task. Any successor is a new version measured against a fresh pack.
3. **What survives:** HG > H1 (+1) is a weak, single-family signal that gating helps relative to ungated use of the same bytes — direction consistent with the development A/B, far below investment thresholds, at a grade that cannot fund anything.
4. **Protocol lesson promoted:** calibration is not optional. Any future pack runs calibration *before* sealing (handoff §7 already requires it; the runner now carries the arm-blind calibration function).

## Run 3 — v1 diagnostic on the same pack (adaptation reuse; development diagnosis, never confirmation)

Option (a) executed: `shape-selector/1` ("trust, then verify against the task") — a history claim earns preference only if the credited rule can strictly reduce the current task by one application (`rewrite.CanStrictlyReduce`); success claims that can't are demoted as distrusted; failure claims against demonstrably-reducing rules are distrusted and stay neutral. v0 remains frozen. v1's remaining lie surface is credit allocation among probe-passing rules — proven still-misleadable by construction in its tests, so measuring it stays meaningful.

**Label enforced in code:** `…+v1-adaptation-reuse-diagnostic` — v1 was designed after seeing this pack's failures; this run can diagnose the repair, never confirm value. Confirmation requires a pack authored after v1 froze.

| Arm | Run 2 (v0) | Run 3 (v1) |
|---|---|---|
| H0 | 14 | 14 |
| H1 | 12 | 12 |
| HG | 13 | **16** |

| Condition | Run 2 | Run 3 |
|---|---|---|
| (b) HG ≥ H1+3 | fail (+1) | **pass (+4)** |
| (c) control loss ≤ 1 | pass (0) | pass (**−3**: HG now *gains* on control strata — mis-01/02 recovered, exactly the targeted defect) |
| (d) ≥ 2 informative families | fail (1) | **fail (1)** — wins still concentrated in fam-negsub |
| (e) H0 guard | **fail** | **pass** (16 ≥ 14; history net-positive) |
| Rule | not satisfied | **not satisfied** ((d) alone) |

**Dispositions:** the targeted defect (lying histories passing the similarity gate) is repaired at diagnostic grade — the H0 guard cleared and the misleading stratum flipped from HG's worst to HG's best. The rule still fails on (d): family diversification. That is the honest open question for a *fresh* pack: whether v1's advantage generalizes beyond the neg-sub construction family or is an artifact of this pack's composition. No tranche is funded; the next admissible measurement is a post-v1-freeze pack (agent-sealed when infrastructure permits, or a second implementer pack under the same pre-commitment discipline, with the adaptation-reuse taint recorded).

## Upgrade path unchanged

Agent-sealed/v1 re-authoring by a clean-room custodian when agent infrastructure recovers; human tiers per the `018` packets. This pack then becomes development material.

## Attributed corrections — 2026-09-13 external review of `main` at `f7554cb` (additive; frozen text above retained verbatim)

The external review reproduced the completion table above with an independent Python reference implementation (exhaustive endpoint checks over the declared four-bit domains) and found the counts correct. It also found four defects in the record's *explanations and instrumentation*, corrected here without changing any frozen criterion, count, or disposition.

### 1. Run-2 condition (c): net zero was narrated as "never lost" — false

The run-2 table row reads "HG never lost control episodes *relative to H1*". The paired control outcomes were:

| Control episode | H1 | HG-v0 | Difference |
|---|---|---|---|
| `low-03` | Miss | Complete | HG gains one |
| `mis-06` | Complete | Miss | HG loses one |

Both arms completed **7/12 controls, but not the same seven**. Corrected reading: **control totals tied at seven; one paired gain offset one paired loss. Condition (c), defined on the net difference, passed.** The frozen net-loss criterion is unchanged; the "never lost" sentence above is retracted as an explanation. The evaluator now exports gross paired wins and losses beside the net figure on condition (c) so this conflation cannot recur silently (`internal/screen`).

`mis-06` also witnesses the attribution correction below: all of its histories pass the v0 similarity gate, and the H1/HG difference there comes from HG's failure-driven demotion of `mul-zero`, not from relevance filtering.

### 2. v0 H1/HG attribution: two mechanisms differ, not one

The frozen comparison changes both **relevance filtering** (HG gates history by task similarity; H1 uses all of it) and **failure handling** (HG demotes rules supported only by relevant failures; H1 ignores failed attempts entirely). An HG−H1 outcome difference therefore cannot be attributed to relevance filtering alone.

Three statements in the frozen text above are **retracted as explanations** on that basis. They are retracted, not rewritten — the frozen text stands verbatim, and an earlier version of this section wrongly said they "are corrected to 'the gated procedure'", asserting a substitution that was never applied to any of them (self-review of this note):

| Frozen site | Retracted phrase | Corrected reading |
|---|---|---|
| Run-2 conditions table, row (b) | "the relevance gate beats ungated history" | the *gated procedure* beats ungated history; which of its two mechanisms produced the margin is unresolved |
| Dispositions, item 3 | "a weak, single-family signal that **gating** helps" | a weak, single-family signal that the gated procedure helps |
| Run-3 prose | the same reading carried forward | as above |

`mis-06` is the concrete witness: all of its histories pass the similarity gate, so the H1/HG difference there comes from HG's failure-driven demotion of `mul-zero`, not from filtering any history out. The mechanism pair is recorded at `internal/shape/comparator.go`. Isolating either mechanism requires a comparator arm that differs in exactly one of them — future-pack design material, not a retrospective edit.

### 3. Run-3 cost accounting: the probes were uncharged, and the diagnostic does not isolate history value

v1 performs per-rule strict-reduction probes against the task **before** its budgeted search; H0/H1 run no such probes, and the runner charged only search expansions. The review's reference implementation counted **91 rule probes producing 188 candidate rewrites** across the 24 episodes, outside the reported cost. Additionally, a review-only, history-free counterfactual (same immediate-reduction probes, reducing rules first in catalog order, same budget-two search) **also completes 16 tasks and matches v1's completion status on all 24 episodes**. The frozen run-3 measurements stand as complete-policy comparisons; what is retracted is any reading of "13 → 16" as isolating *history-conditioned* shaping value. The repairs going forward: the selector meters its probe work on the `Decision` (`ProbeRuleApplications`, `ProbeCandidates`), the runner charges it into the arm's task ledger in one unit (candidate rewrites materialized, search + probes), expansion counts remain in the traces as the budget unit, unmeasured custody is recorded as unmeasured (never zero), and any future test of incremental history value must include a comparably capable task-only probe arm.

### 4. Corpus: `inf-06`'s one-node target is unattainable through its catalog

`inf-06` (start `add(mul(x, 0), mul(y, 1))`, target cost 1) is reachable only to `add(0, y)` under its catalog (`not-intro`, `mul-zero`, `mul-one`, right-zero `add(a,0)→a`): the catalog lacks left-zero elimination and `add-comm`, so the semantically valid one-node result `y` is structurally unreachable — an argument from the permitted rewrite structure, not from budget exhaustion. All arms fail it identically; recorded counts stand. Annotated at the spec site in `pack_agent_sealed_v1.go` (pack bytes retained). Related calibration repair: `CalibrateH0MinBudgets` no longer conflates "already completes with zero expansions" with "unreachable within cap" (unreachable episodes are now reported separately), and median derivation guards the empty case.

### 5. Run-3 claim scope: a one-step probe negative was narrated as permanent uselessness

The run-3 description above states that success claims whose credited rule "can't [strictly reduce the current task by one application]" are "demoted as distrusted". The emitted rationale went further, reading "can never reduce this task; the success claim is distrusted". Both exceed what the probe checked. The probe answers exactly one question: *can one application of this rule, at some position of the current start expression, strictly lower NodeCount?* A negative answer is not a reachability result, and it is not a verdict on the history. The external review's counterexample uses rules already on the menu:

```text
add(0, x) --add-comm--> add(x, 0) --add-zero--> x
```

`add-zero` cannot one-step reduce `add(0, x)`; it becomes the reducing step after a cost-neutral transformation. The diagnostic executes this path. Separately, a historical success can be genuine even when its credited rule does not help *this* task — current usefulness and historical truth are different claims.

Corrected reading, and the wording the selector now emits: **"no one-step strict NodeCount decrease from the current start."** Demotion remains a declared heuristic. It is not a proof of permanent uselessness and not a refutation of the history. A regression test fails the build on any rationale containing "can never reduce".

Three frozen sites carry the over-scoped wording and are **retracted as explanations** (the frozen text stands verbatim):

| Frozen site | Retracted phrase | Corrected reading |
|---|---|---|
| Dispositions, item 2 | "distrusting successes whose credited rules **never reduce the current task**" | distrusting successes whose credited rules show no one-step strict NodeCount decrease from the current start |
| Run-3 prose | "success claims that can't are demoted as **distrusted**" | preference is withheld by a declared heuristic; the claim is not adjudicated |
| Run-3 prose | "failure claims against demonstrably-reducing rules are **distrusted** and stay neutral" | the demotion is withheld because the rule demonstrably reduces; the historical failure is not called false |

"Distrusted" is the specific word to avoid: it reads as a verdict on whether the history is truthful, which the probe never tests. Current one-step usefulness and historical truth are different claims, and the shipped rationale says so explicitly.

**Symbol correction (same run-3 prose):** it cites `rewrite.CanStrictlyReduce` as the selector's probe. The selector calls **`rewrite.ProbeStrictReduction`** — the metered form. `CanStrictlyReduce` remains as the un-metered boolean convenience, and that un-metered form is precisely what finding 1 charged as uncharged work, so the citation names the one variant that would not have been chargeable.

**Enumeration correction (same run-3 prose):** it opens "Option (a) executed". No lettered enumeration of these repairs exists — dispositions item 2 lists two *unlabeled* candidate repairs, and `2026-09-12-017`'s only lettered options are forks F1–F3 (inputs / output levers / policy-identity home), none of which is this repair. Worse than a dangling label: read in the order item 2 lists them, "(a)" points at **failure-aware relevance weighting**, while the shipped selector implements the **second** alternative — distrusting success claims whose credited rule shows no one-step reduction. The correct reading is "the second candidate repair from dispositions item 2 executed"; relevance weighting was never implemented.

### 5a. What the probe meter shows once the work is charged

The per-arm cost ledger is now logged beside the completion counts (task cost = candidate rewrites materialized, split into search-generated and selector-probe components). At the calibrated budget, run 3 reports:

| Arm | Charged task cost | search-generated | selector probe |
|---|---:|---:|---:|
| H0 | 325 | 325 | 0 |
| H1 | 322 | 322 | 0 |
| HG (`shape-selector/2`) | **566** | 287 | **279** |

The probe component (279 = 91 rule applications + 188 candidate rewrites) **exceeds** the search work it saves: HG expands the fewest candidates of any arm (287 vs 325) and is nonetheless the **most expensive** arm once its pre-search probes are charged — a 74% cost premium over H1 for its four extra completions. Under the previous accounting HG appeared cheapest. This is a measured consequence of finding 1, not a new experiment: the figures come from the same retained run whose completion counts are unchanged. It sharpens the external review's point that "13 → 16" is not a clean capability gain, and it is the number a spending decision should see beside the margin.

### 5b. Input identity: the decision hash omitted inputs that change the decision

The selector's input hash covered the embedded textual `Input`, domain, rule **names**, and version. It did not bind rule bodies, and the procedure never checked that the actual probed expression agreed with its declared rendering. Two source-derived counterexamples from the external review:

| Change | Old hash payload | Decision change |
|---|---|---|
| Hold `TaskStart` fixed; change the actual task from `not(not(x))` to `x` | unchanged | double-not goes from reducing to non-reducing |
| Replace a rule with a different valid rule bearing the same name | unchanged | probe result and preference change |

The retained run is **not** corrupted: the runner constructs consistent inputs from a fixed menu, so neither substitution occurred. This was an exposed API/provenance defect — the identity contract was not enforced at the boundary.

Both are now closed by **refusal** rather than by hashing more: the selector errors on a task expression that does not render to `Input.TaskStart`, and on duplicate rule names carrying different content (exact duplicates are tolerated). Rule content enters the hash through `rewrite.Rule.Identity()` (name, admitted domain, both rendered sides). The task needs no separate hashed copy — every accepted input satisfies `Render(Task) == TaskStart`, so the hashed `Input` already pins the probed expression. A self-review caught the first attempt hashing a redundant `Task` field whose value could not diverge on any accepted input, and the test asserting it moved would have passed without the fix; both are gone.

### 6. Bounded work, and what the first repair left open

The external review showed that a maximum depth of 64 bounds nothing in practice: a depth-30 expression whose `Binary` children share one subexpression value costs 2^31−1 logical node visits while staying far under the depth limit. Search had a related boundary — one counted expansion applies all rules at all positions and materializes every successor, with `not-intro` admitting growth and no ceiling on generated state or term size.

A self-review of the first repair found it **incomplete in two ways**, both now closed:

- The repair bounded the *validator's own* traversal at 2^20 visits while still admitting expressions whose canonical rendering was megabytes. Every downstream consumer — search visited-set keys, selector identity hashes, oracle replay — then paid that size repeatedly. The bound is now on **tree size at admission** (`finite.MaxExprNodes`), so a validated expression is bounded for every consumer, and `rewrite`'s successor ceiling is defined as that same constant rather than an independent number that could drift from it.
- The search's size ceiling **rendered each successor and then measured the string**, paying the cost it existed to refuse; a measured audit search spent 66 seconds on one expansion, rendering 16,384 oversize terms and retaining none. Successors are now measured on the tree and dropped at construction.

The load-bearing repair is at the consumer: **a resource stop is not a non-completion.** The first repair added `StateBounded`, `TermSizeBounded`, and `Cancelled` to the search result, and *no consumer read them* — so a truncated search reached the screen grid as `Completed: false`, indistinguishable from a searched-and-failed miss. That is precisely the semantic refutation the roadmap forbids. The runner now aborts a run whose search was resource-truncated, and `screen.Evaluate` refuses a batch containing a cell marked `MeasurementBlocked` rather than scoring it. `BudgetExhausted` is deliberately excluded from that set: the budget is the declared measurement parameter, and stopping on it is the measurement working as designed. No retained run is affected — every pack expression is tens of nodes — but the recorded figures now rest on a measurement that cannot silently become a truncation.

### 7. Controller identity: the corrections made a new version, and it is labeled as one

The repairs in §3, §5, §5b, and §6 leave the **rule ordering unchanged** for every accepted input — run 3 still reports 14/12/16 — but they do not leave the *procedure* unchanged: it now refuses inputs its predecessor decided, and its frozen probe-description parameter changed, moving its snapshot hash (`dc88c9fc…` → `5f0dc8e2…`). `internal/shape/shape.go` states the governing rule: "any change is a new version." Keeping the `shape-selector/1` label over changed frozen parameters would have put two procedures behind one identifier, and disclosure in a comment does not discharge that rule.

The controller is therefore **`shape-selector/2`**. Consequences recorded here:

- Record 019's retained run-3 figures **belong to `shape-selector/1`**, which stays frozen and recoverable from git at `f7554cb`. The re-run under `/2` reproduces them, and its outcome label names `/2` so the two measurements cannot be conflated.
- The confirmatory runner's freeze anchor was a hard-coded commit hash naming `/1`'s freeze. Once `/1`'s parameters changed, that hash silently anchored to a superseded procedure, and a pack authored between the two commits would have been admitted as confirmatory evidence for a controller frozen *after* it. The anchor is now expressed as the version identity plus the emitted label — a hash in a comment cannot stay correct across a version bump; a version string can.
- The bump makes one more frozen sentence stale, **retracted as an explanation** (frozen text stands verbatim): the run-3 line "**Label enforced in code:** `…+v1-adaptation-reuse-diagnostic`" describes a label no code path emits. The label emitted today is `implementer-authored/v1+shape-selector/2-adaptation-reuse-diagnostic` — the `implementer-authored/v1` prefix is the pack's evidence tier, unrelated to the controller version, and the controller component now names `/2`. Adaptation-reuse enforcement is unchanged and asserted in the runner's test.
- **The rule that failed here now has a mechanical guard.** `shape.go`'s "any change is a new version" was doctrine with nothing checking it, which is why a frozen parameter moved under a fixed label in the first place. `shape.frozenSnapshotV2` pins the frozen-parameter hash (`da3e7c650d0d…`), and a regression test fails if a parameter moves — the only correct repair being a version bump and a pin update in the same change. Editing the pin alone to restore green reproduces the original defect, and the test says so in its failure message.

### 8. One frozen figure class DID change: per-arm task cost

Every completion count, condition verdict, and disposition above is unchanged and reproduces. One recorded figure class does move, and saying otherwise would be the same accounting dishonesty this note exists to correct: **per-arm `TaskCost`**. It previously counted search expansions only; it now counts candidate rewrites materialized plus charged selector probe work (§3, §5a). Any comparison against a task-cost figure recorded before this change will disagree, by construction — that was the point of finding 1. Completion counts, which every disposition rests on, are untouched.

Two accounting limits worth stating rather than discovering later:

- The charged unit **sums two measures** — rules probed (`ProbeRuleApplications`) and candidates materialized (`ProbeCandidates`). A probe's positional traversal is charged a flat 1 rather than a cost proportional to task size, so the figure is a lower bound on probe work, not a proportional model. It is nonzero, which is what finding 1 required; it is not calibrated.
- `Generated` counts a nested oversize drop as **one** dropped successor even when the drop suppresses a whole successor family. Immaterial to every retained run (`dropped == 0` throughout, since no pack expression approaches the admission ceiling), but the figure would understate work on a truncated run — and a truncated run is now refused outright, so it cannot reach a recorded result.

*Scope and attribution. Sections §1–§5b answer the 2026-09-13 external review, one section per finding plus the corpus item: §1 net-vs-gross (finding 4), §2 attribution and §3 cost accounting (finding 1), §4 corpus, §5 claim scope (finding 2), §5b input identity (finding 3), §6 bounded work (finding 5). Sections §5a, §6–§8 are attributed to two passes over that remediation: a self-review, and an independent validator that re-derived the frozen counts and mutation-tested each repair.*

*What §2, §5, and §7 retract are **explanations**: the frozen text stands verbatim and the retraction tables name each site rather than asserting a substitution that was never applied — an earlier version of §2 made exactly that false claim and is itself corrected here. Every completion count, condition verdict, and disposition above is unchanged and reproduces. **One recorded figure class did change** — per-arm task cost, by construction of finding 1 — and §8 states it rather than leaving it inside a blanket "nothing changed" claim. The code repairs land across four commits, all referenced from `CHANGELOG.md`.*
