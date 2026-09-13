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
