# (l-coh reviewer-correction arc) — Consolidation Record: nine numbered corrections, current status, evidence tiers, and remaining obligations

**Status**: authored 2026-09-12 as a single canonical summary of the
pilot-004 (l-coh) reviewer-correction arc. Supersedes the scattered
CLOSURE-SCORECARD entries for narrative continuity; does not
replace them — the scorecard entries remain the authoritative
lifecycle-status source. This record answers "where does the arc
stand end-to-end and what remains?" in one place.

## Scope

The reviewer's message on 2026-09-12 (1:44 PM UTC-7) responded to
`l-coh-measurement-result.md` and delivered nine numbered
corrections. The arc that followed comprised seven discrete
executed loops (`l-coh-mechanism-test`, `l-population-validation`,
`source-lineage-check`, `m7-emission-nondivergence`,
`weaken-repair-policy`, `source-lineage-admission-gate`, and
pilot-005 `openitem-4-status`). This record summarizes:

1. What each of the nine corrections said.
2. What was accepted, discharged with evidence, or retained with
   narrowed warrant.
3. What evidence tier(s) support the current disposition.
4. What obligations remain open.

## Nine reviewer corrections and their current dispositions

### Correction 1 — Specify exactly what the negative COH-A result rejects

**Reviewer's point**: The negative COH-A result should not be described as
"the statistic is falsified" or "failure-population coherence is
falsified." What is rejected is narrower: **COH-A's numerical
conditions fail** (the stipulated 0.6 threshold is not met by the
observed 0.257 difference, ratio 2.5×). The statistic itself
produced a value; the rejected claim concerns its threshold or
explanatory role. Further, the operationalization substituted
outcome discrimination for within-population coherence — a
different quantity.

**Current disposition**: **Accepted.** `l-coh-measurement-result.md`
was revised to state the rejection precisely and to note the
operationalization substitution.

**Evidence tier**: text correction; no additional measurement.

**Remaining obligation**: none.

### Correction 2 — COH-B is not sufficiently defined to classify confidently; predeclared threshold-robustness contradicted

**Reviewer's point**: The predeclaration defined COH-B using "high,"
"low," and "comparable" without numerical definitions and overlaps
COH-C. Result cannot be confidently classified. Also: the
predeclaration claimed thresholds ≥0.4 preserve the qualitative
pattern; the reported tables contradict this (at 0.4, M7 has one
axis-value and pvnp has none; at 0.5, both have none).

**Current disposition**: **Accepted.** `l-coh-measurement-result.md`
was revised to state that COH-A was not obtained and that COH-B/C
adjudication is underdetermined by overlapping, incompletely
specified reading rules. The threshold-robustness claim was
withdrawn (the frozen 0.5 calculation is unaffected).

**Evidence tier**: text correction; no additional measurement.

**Remaining obligation**: none for this arc; a future coherence
measurement predeclaration should number its reading rules.

### Correction 3 — "n=2 truly independent" too strong; retain provenance evidence path

**Reviewer's point**: Discounting the four ES databases as four
independent replications is warranted; going all the way to "n=2
truly independent" between ES and P-vs-NP is not. Identical
aggregate distributions do not themselves prove identical
underlying observations. Family prevalence must actually count the
miner-eligible families, not every persisted signature.

**Current disposition**: **Accepted, with provenance validation
completed.** `l-coh-measurement-result.md` was revised to the
reviewer's defensible-replacement wording: "Five database
instances represent two substantive corpus lineages for this
comparison. The four ES instances do not provide four independent
research replications. Independence between the two corpus
constructions has not been established."

The **ES-side clause** was strengthened by
`source-lineage-check-result.md`: SGO-tier evidence that the four
ES databases share 12/13 `source_snapshots.sha256` values
byte-for-byte. See also
`source-lineage-admission-gate-result.md`: the finding is now
reachable via `newf source lineage-diff` (any operator can
reproduce with one command).

The **ES-vs-pvnp clause** remains open — source-byte disjointness
is confirmed (necessary condition for research independence), but
not sufficient.

**Evidence tier**: SGO for ES-side clause; ES-vs-pvnp clause
remains open.

**Remaining obligation**: ES-vs-pvnp research-methodological
independence beyond source bytes. Requires either (l-extension) or
deeper analysis.

### Correction 4 — Family prevalence must count miner-eligible families, not every persisted signature

**Reviewer's point**: Query-population validation is necessary
before the (l-coh) numerical tables are taken as authoritative.

**Current disposition**: **Discharged.**
`l-population-validation-result.md` traced the miner's eligibility
rules (`fc ≥ 2` support gate and family-signature aggregation),
recomputed posture-axis prevalences across all five ES databases
plus pvnp, and verified the (l-coh) numerical tables were computed
for the correct population. All four ES databases produced
identical miner-eligible tables — consistent with the
byte-identical source lineage found in (source-lineage-check). A
narrative correction to pilot-003's counts was made in
`l-coh-measurement-result.md`.

**Evidence tier**: SGO (SQL verification against the actual
miner-eligible population).

**Remaining obligation**: none.

### Correction 5 — Zero-cell observation close to encoded rule; unconditional `iff` overgeneralizes

**Reviewer's point**: `VerifySuccessPreserving` confirms exactly
when at least one eligible success matches the predicate, but the
`iff` proposed was too strong. The correct refinement: "Among
emitted, support-valid candidates with comparable resolved
populations, absence of a preserving success prevents this
particular weakening route. Overall survival additionally depends
on the remaining challenge outcomes and claim semantics." Also:
the miner's emit-gate (≥ 2 distinct failure-side families, failure
prevalence > success prevalence) is an earlier selection gate that
must be considered — several posture-axis candidates in the (l-coh)
tables (`uncertainty=probabilistic` in both M7 and pvnp) are not
even emission-eligible.

**Current disposition**: **CMA-tier upgrade completed.**
`l-coh-mechanism-test-result.md` implemented a controlled
policy-mechanism test
(`internal/invariant/success_preserving_discrim_test.go`): two
fixtures with **identical discrimination** but different `sPrev`
values (Fixture A: 5/10 vs 0/10, discrim=+0.5 → confirmed-false,
completed-negative; Fixture B: 8/10 vs 3/10, discrim=+0.5 →
confirmed-true, contrast-collapse). Test passed, upgrading the
refinement to CMA-tier at the code-mechanism level.

Additional live-corpus cross-check:
`m7-emission-nondivergence-result.md` showed that M7's (a‴-B) run
persisted BOTH posture-axis candidates (`equals(locality, local)`
and `equals(construction, constructive)`), and their divergent
outcomes (`surviving` vs. `weakened`) provided a live real-data
instance of the correction 5 refinement in action.

Design-verification: `weaken-repair-policy-result.md` recorded
that the miner's directional emit-gate (`fPrev > sPrev`) and the
success-preserving probe's zero-cell binary threshold operate on
different scales by deliberate design — this asymmetry is
intentional, not a defect.

**Evidence tier**: CMA (controlled fixture test) + live corpus
cross-check + design-doctrine confirmation.

**Remaining obligation**: none for this arc.

### Correction 6 — Post-hoc discovery needs qualification, not epistemic prohibition

**Reviewer's point**: A post-hoc observation can motivate a valid
deduction about code (which the zero-cell mechanism does). A
preregistered prediction on a third corpus provides prospective
evidence but doesn't turn an empirical generalization into an
arithmetic or code-level deduction. The predeclaration's proposed
promotion of the coherence explanation to SGO upon obtaining COH-A
has the same problem — matching a threshold does not transform an
explanatory inference into a directly observed fact.

**Current disposition**: **Accepted.**
`l-coh-measurement-result.md` was revised to distinguish
discovery timing, evidence type, and strength of support. The
correction-7 (renumbered correction 5 in this consolidation) CMA
upgrade shows the reviewer-suggested "controlled policy-mechanism
test" was the right next test, and it was executed.

**Evidence tier**: text correction + subsequent CMA-tier test.

**Remaining obligation**: none.

### Correction 7 — The strongest new insight is about how the challenge policy converts corpus composition into lifecycle status

**Reviewer's overarching frame**: "The most important remaining
distinction is between discovering structure in research and
rediscovering the decision rule that labels that structure
`surviving`." The (l-coh) result is more informative about the
decision rule than about research coherence.

**Current disposition**: **Retained as the arc's central
insight.** All subsequent loops in this arc served this frame:
- `l-coh-mechanism-test` proved the decision rule discriminates
  by `sPrev` at fixed discrimination (CMA).
- `l-population-validation` proved the population was measured
  correctly (SGO).
- `source-lineage-check` proved the ES cross-database "consistency"
  reflected identical source bytes (SGO).
- `m7-emission-nondivergence` proved the decision rule's live
  application on M7's real (a‴-B) run matched the CMA prediction.
- `weaken-repair-policy` proved the terminal-`weaken` behavior on
  contrast-collapse is intentional design, not gap.
- `source-lineage-admission-gate` makes the source-byte diagnostic
  operator-reachable.

**Evidence tier**: four-layer evidence chain (CMA + SGO ×3 +
design-doctrine).

**Remaining obligation**: none for this arc; the frame will guide
future coherence-related work.

### Correction 8 — Predeclaration's proposed SGO promotion on COH-A match is not defensible

**Reviewer's point**: Matching a threshold does not transform an
explanatory inference into a directly observed fact. Keep counts,
calculations, code implications, and research interpretation
separately attributed.

**Current disposition**: **Accepted.** No promotion of any
coherence explanation to SGO was performed on the basis of the
frozen calculation. The CMA-tier upgrade in correction 5 was
supported by a controlled test, not by threshold matching.

**Evidence tier**: text discipline.

**Remaining obligation**: none.

### Correction 9 — Retain source-lineage validation, cluster-run IDs, and row-to-family mapping before making provenance conclusions definitive

**Reviewer's point**: Before making the "identical ES source
data" conclusion definitive, retain the SQL, source/snapshot
lineage, selected cluster-run IDs, normalization versions, and
row-to-family mapping.

**Current disposition**: **Discharged.**
`source-lineage-check-result.md` recorded SGO evidence across five
fingerprint layers (source_snapshots.sha256, sources.logical_name,
mechanism_signatures.fingerprint, mechanism_clusters.cluster_fingerprint,
cluster_runs.input_set_hash). All exact values persisted to
terminal history at time of check. `l-population-validation-result.md`
recorded the row-to-family mapping via the miner's eligibility
rules.

**Evidence tier**: SGO.

**Remaining obligation**: none for this arc.

## Evidence-tier summary

| Correction | Disposition | Best evidence tier |
|---|---|---|
| 1 (COH-A rejection scope) | Accepted (text) | Text correction |
| 2 (COH-B underspecification) | Accepted (text) | Text correction |
| 3 (independence claim) | Accepted; ES-side discharged, ES-vs-pvnp open | SGO for ES-side clause |
| 4 (population correctness) | Discharged | SGO |
| 5 (zero-cell `iff` refinement) | CMA-tier | CMA (controlled test) + SGO + design-doctrine |
| 6 (post-hoc qualification) | Accepted (text) | Text correction |
| 7 (overarching frame) | Retained as insight | Four-layer evidence chain |
| 8 (SGO promotion discipline) | Accepted (text) | Text discipline |
| 9 (provenance retention) | Discharged | SGO |

## Remaining obligations

- **ES-vs-pvnp research-methodological independence** (correction
  3, ES-vs-pvnp clause): open. Source-byte disjointness confirmed
  (necessary), but not sufficient. Requires (l-extension) or
  deeper analysis.
- **(l-extension) authoring**: a truly third-independent research
  corpus in a non-ES, non-P-vs-NP domain, sha256-disjoint from
  both known lineages. The `newf source lineage-diff` diagnostic
  now supports this as a freeze-checklist item.

## What the arc did NOT do

- Did NOT change any admission rule, verifier tier, or policy
  directive.
- Did NOT modify H3 admission-gate behavior.
- Did NOT establish research-methodological independence between
  ES and pvnp.
- Did NOT promote any coherence explanation to SGO status.
- Did NOT rerun challenge campaigns or modify existing (l-coh)
  numerical tables (unchanged).

## What the arc DID do

- Corrected the (l-coh) result narrative on all nine reviewer
  points.
- Upgraded the zero-cell refinement to CMA-tier via controlled
  fixture test.
- Discharged the ES-side independence residual with SGO-tier
  source-byte evidence.
- Validated the (l-coh) tables' population arithmetic.
- Documented the miner emit-gate ↔ success-preserving probe
  asymmetry as intentional design.
- Made the source-lineage diagnostic operator-reachable via
  `newf source lineage-diff`.

## Consequences for the scorecard

Every entry in the "closed" list of the CLOSURE-SCORECARD's
recommended-next-action cell (twenty-third revision) traces to a
correction disposition above. The scorecard remains the
lifecycle-status source of truth; this record is the arc-level
summary the operator can consult once instead of reconstructing it
from scattered per-loop records.

## Consequences for future work

The reviewer's overarching frame (correction 7) sets the target for
subsequent coherence-related work: **distinguish structure in
research from the decision rule that labels structure `surviving`.**
Two future paths align with this frame:

1. **(l-extension)** ⭐: a third truly-independent corpus would
   probe whether the decision rule's `surviving` labels track a
   research property that generalizes across substantive domains,
   or whether they primarily reflect the rule's implementation.

2. **(coherence-metric-pipeline)**: less compelling given the CMA
   evidence for correction 5. Would only earn its keep if a future
   observation calls for the metric's discipline as a first-class
   pipeline capability rather than as an ad-hoc measurement.

Any other coherence hypothesis proposed in future work must be
disciplined against this distinction from the outset.
