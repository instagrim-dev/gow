# (l-coh) Coherence measurement result — predeclared Reading COH-A falsified; sharper post-hoc signal recorded as hypothesis

**Status**: executed 2026-09-12 at HEAD `00c3c58`. Predeclaration:
`records/l-coh-measurement-predeclaration.md`.

## Bottom line

**Predeclared Reading COH-A is FALSIFIED.** The threshold of
`maxDiscrim ≥ 0.6` for ES corpora is not met (M7 = 0.429). The
predicate that DID reach `surviving` in the miner's actual run has a
maxDiscrim of 0.429, below the predeclared floor. The interpretive
"coherence" framing needs to be weakened.

**Predeclared Reading COH-B is OBTAINED.** ES maxDiscrim is ~2.5×
pvnp maxDiscrim (0.429 vs 0.171); the ordering matches the (l) PE
claim, but the magnitude is meaningfully smaller than "coherent
cluster vs heterogeneous grid" implied.

**Post-hoc signal (recorded as hypothesis, NOT established)**: the
sharper discriminator between ES and pvnp is `minSuccessPrev` over
failure-positive `(axis, value)` pairs — 0.000 in ES vs 0.286 in
pvnp. This is a **zero-cell** signal in the failure × success
posture crosstab, and it maps directly to the pipeline's
`success-preserving` challenge probe: the probe returns `confirmed`
(→ weaken) as soon as ANY success family preserves the predicate.

## Measurement (SGO)

Query: filtered per-train-problem posture distribution joined through
`mechanism_signatures → mechanisms → approach_revisions → approaches
→ problems`; success-side = `partial_success + success`; failure-side
= `failure + partial_failure`. Verbatim SQL and outputs in this
session's terminal record.

### Data-provenance note

- M7, pilot-001, pilot-002 have **identical** posture distributions
  under train-problem filtering. This is not evidence of independent
  replication — the corpora share source data and the same fixture
  normalizer.
- pilot-003's counts are **exactly 3× M7's**; same underlying
  distribution scaled by pipeline HEAD-time.
- Therefore the (i) headline "perfect replication across 4 corpora"
  is really *one* corpus observed four times at different HEADs.
  The **independent-observations count for the (l) hypothesis
  is 2**: ES-shape and P-vs-NP-shape.

### M7 (ES) — N_F=7 failure-side, N_S=5 success-side

| axis | value | fPrev | sPrev | discrim |
|---|---|---:|---:|---:|
| construction | constructive | 0.429 | 0.200 | +0.229 |
| construction | existential | 0.571 | 0.800 | −0.229 |
| locality | global | 0.429 | 0.600 | −0.171 |
| locality | **local** | 0.429 | **0.000** | **+0.429** |
| locality | mixed | 0.143 | 0.400 | −0.257 |
| uncertainty | deterministic | 0.857 | 1.000 | −0.143 |
| uncertainty | probabilistic | 0.143 | 0.000 | +0.143 |

- **maxDiscrim = +0.429** at `equals(locality, local)`
- **posDiscrimCount(≥0.5) = 0**
- **minSuccessPrev(post-hoc) = 0.000** at `equals(locality, local)`

### pvnp (P vs NP) — N_F=5, N_S=7

| axis | value | fPrev | sPrev | discrim |
|---|---|---:|---:|---:|
| construction | constructive | 0.600 | 0.714 | −0.114 |
| construction | existential | 0.400 | 0.286 | +0.114 |
| locality | global | 0.600 | 0.429 | +0.171 |
| locality | local | 0.400 | 0.571 | −0.171 |
| uncertainty | deterministic | 0.800 | 0.714 | +0.086 |
| uncertainty | probabilistic | 0.200 | 0.286 | −0.086 |

- **maxDiscrim = +0.171** at `equals(locality, global)`
- **posDiscrimCount(≥0.5) = 0**
- **minSuccessPrev(post-hoc) = 0.286** at `equals(uncertainty, probabilistic)`

## Predeclared readings — disposition

- **Reading COH-A (structural claim confirmed)**: **FALSIFIED**. My
  predeclared threshold required ES `maxDiscrim ≥ 0.6`; M7's actual
  maxDiscrim is 0.429. I explicitly predeclared that adjusting the
  threshold post-hoc does not count. The threshold falsifies.
- **Reading COH-B (partial support)**: **OBTAINED**. Ordering matches
  the (l) PE claim (ES 0.429 > pvnp 0.171, ~2.5×); magnitude weaker.
- **Reading COH-C (structural claim falsified)**: NOT obtained. The
  ordering direction is not reversed.

## Correlation with miner productivity (SGO)

- M7 miner (a‴-B) run produced `equals(locality, local)` at
  `surviving`.
- pvnp miner (a‴-B) run produced 3 candidates, all reaching `weaken`.

Under the predeclared coherence measure, neither corpus should reach
`surviving` (both have `posDiscrimCount(≥0.5) = 0`). This is another
falsification signal for the predeclared measure: M7 DID produce a
`surviving` invariant despite `maxDiscrim = 0.429 < 0.5`.

## Post-hoc hypothesis (flagged, not established)

The `minSuccessPrev` observation is a **candidate refinement** of the
(l) PE claim. Under this hypothesis:

> **Enum-axis mining reaches `surviving` on `equals(axis, value)` iff
> the success-side prevalence of that `(axis, value)` is exactly zero.**

This maps directly to the code-owned `success-preserving` probe
(`internal/pipeline/invariant.go`, `internal/provider/invariant_challenger.go`):
the probe returns `confirmed → weaken` as soon as ANY success family
preserves the predicate. Zero success preservers = probe returns
`unconfirmed` = the candidate can survive.

Support for this hypothesis:
- M7: `equals(locality, local)` has sPrev = 0.000 → `surviving`.
- pvnp: minimum sPrev over failure-positive `(axis, value)` = 0.286
  → all candidates `weaken`.

**Why this is NOT established, only a hypothesis:**

1. **Post-hoc**: found by scanning the data after the predeclared
   measure failed. The predeclared measure has authority; the
   post-hoc measure does not.
2. **n=2 truly-independent-corpora**: two observations do not
   establish a pattern.
3. **Not a mechanism claim**: correctly explains the two observations
   but does not predict any third-corpus outcome yet.
4. The hypothesis would need to be **predeclared before the next
   corpus is executed** to earn any epistemic status.

## Update to the (l) result record

The (l) result's structural finding —

> *the enum-axis miner's recurring-invariant productivity depends on
> failure-population coherence in posture-axis space*

— should now be read as:

- **Correct in direction**: ES has stronger failure-side
  discrimination than pvnp on the enum axes.
- **Overstated in magnitude**: "coherent cluster vs heterogeneous
  grid" implies a large difference; measured difference is ~2.5× on
  maxDiscrim.
- **Wrong about the mechanism**: the actual driver of the
  ES-vs-pvnp `surviving` vs `weaken` distinction appears to be a
  **zero-cell effect** in the failure × success crosstab, not a
  discrimination magnitude. This is a hypothesis, not established.

The (l) record itself is not being rewritten (this record's job is
to record the correction). The scorecard entry will annotate the
narrowed warrant.

## What DOES this loop establish

- SGO measurement of failure/success posture distributions on 5
  persisted databases.
- Data-provenance finding: only 2 of the 5 are truly independent
  observations for cross-corpus claims about the miner. **This
  substantively weakens the (i) result too** — from "perfect
  replication across 4 corpora" to "perfect replication of ONE
  observation replicated across 4 pipeline HEADs plus 1 truly
  independent domain." The (i) claim is not falsified but its
  epistemic strength is lower than the earlier record suggested.
- Predeclared coherence measure is FALSIFIED at the threshold.
- Post-hoc `minSuccessPrev` signal is recorded as a candidate
  hypothesis for future predeclared replication.

## What this loop does NOT establish

- The post-hoc `minSuccessPrev` claim (see hypothesis flags above).
- Any claim about ES or P vs NP as domains beyond the SGO
  distributions.
- Any change in the (l) result record itself.
- Any discharge of C7/C8/witness-occurrence-attribution obligations.

## H3 gate status: PRESERVED

No admission rules changed. No new verifiers. No policy directives.
No wire schema fields. No pipeline capabilities added — this is
SGO measurement on existing SQLite state.

## Corrections propagated

- (l) result: warrant narrowed as described above.
- (i) result: data-provenance caveat added — the four ES corpora are
  one observation replicated across HEADs, not four independent
  observations. This is a substantive weakening.

## Verification tier

- Per-corpus posture distributions: **SGO** (verbatim sqlite3
  output on each persisted DB).
- Coherence measure per corpus: **CMA** (arithmetic on the SGO
  counts; independently reproducible from the tables).
- Predeclared threshold falsification: **CMA**.
- Data-provenance finding (identical distributions across M7/pilot-1/2/3):
  **SGO** (verbatim equality check).
- Post-hoc `minSuccessPrev` hypothesis: **PE** (recorded, not
  established); would need predeclared replication on a third
  independent corpus to earn CMA status.
