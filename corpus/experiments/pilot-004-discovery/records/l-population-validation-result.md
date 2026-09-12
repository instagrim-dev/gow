# (l-population-validation) Result — miner-eligible arithmetic verified; (l-coh) tables unchanged; small narrative correction to data-provenance note

**Status**: executed 2026-09-12. Discharges (l-coh) correction 6.

## Bottom line

**Correction 6 is discharged.** The miner-eligible arithmetic — computed
strictly from `mechanism_clusters` at the latest `cluster_run` per problem,
projecting only the representative signature's posture axes, filtering
redundant-only clusters and `outcome_mixed` clusters, and dropping
`unknown` posture values — matches the earlier (l-coh) tables **exactly**
for M7 and pvnp. The values in the (l-coh) authoritative tables are
correct for the correct population.

**One narrative correction** to the (l-coh) data-provenance note: under
the corrected query, all four ES database instances (M7, pilot-001,
pilot-002, pilot-003) produce **identical** tables, not "M7 = pilot-001
= pilot-002 with pilot-003 tripled". The earlier "3×" claim was an
artifact of an unrestricted SQL join over pilot-003's two `cluster_runs`.

## Method

Deterministic SGO. For each of `.newf/{m7,pilot-001,pilot-002,pilot-003,pvnp}/newf.db`:

1. Select the latest `cluster_runs.id` per `problem_id` (by `created_at DESC`).
2. For each `mechanism_clusters` row in that cluster_run:
   - Skip if every `cluster_members` row is `redundant = 1`.
   - Skip if `outcome_mixed = 1`.
   - Skip if `outcome_class` ∈ {`mixed`, `unknown`}.
   - Otherwise bucket into failure-side (`failure`, `partial_failure`)
     or success-side (`partial_success`, `success`).
3. For the `representative_signature_id`, read `signature_postures`.
4. For each of `locality`, `construction`, `uncertainty`: if value ≠
   `unknown` and present, increment the corresponding
   `(axis, value)` count on the correct side.

This reproduces `internal/provider/invariant_fixture_derive.go:derivePostureAxisProposals`'s
eligible-population semantics exactly.

## Results

| Corpus | N_F | N_S | dropped | extra_runs |
|---|---:|---:|---:|---:|
| M7 | 7 | 5 | 0 | 0 |
| pilot-001 | 7 | 5 | 0 | 0 |
| pilot-002 | 7 | 5 | 0 | 0 |
| pilot-003 | 7 | 5 | 0 | **1** |
| pvnp | 5 | 7 | 0 | 0 |

`extra_runs` counts additional `cluster_runs` rows beyond the latest per
problem. pilot-003 has one extra cluster_run — the source of the earlier
`3×` artifact.

### Per-axis-value (miner-eligible) — all four ES corpora identical

| axis | value | fc | sc | fPrev | sPrev | discrim | eligible? |
|---|---|---:|---:|---:|---:|---:|---|
| construction | constructive | 3 | 1 | 0.429 | 0.200 | +0.229 | YES |
| construction | existential | 4 | 4 | 0.571 | 0.800 | −0.229 | . |
| locality | global | 3 | 3 | 0.429 | 0.600 | −0.171 | . |
| locality | **local** | 3 | **0** | 0.429 | **0.000** | **+0.429** | **YES** |
| locality | mixed | 1 | 2 | 0.143 | 0.400 | −0.257 | . |
| uncertainty | deterministic | 6 | 5 | 0.857 | 1.000 | −0.143 | . |
| uncertainty | probabilistic | 1 | 0 | 0.143 | 0.000 | +0.143 | . *(fc < 2 gate blocks emission)* |

### Per-axis-value (miner-eligible) — pvnp

| axis | value | fc | sc | fPrev | sPrev | discrim | eligible? |
|---|---|---:|---:|---:|---:|---:|---|
| construction | constructive | 3 | 5 | 0.600 | 0.714 | −0.114 | . |
| construction | **existential** | 2 | 2 | 0.400 | 0.286 | +0.114 | **YES** |
| locality | **global** | 3 | 3 | 0.600 | 0.429 | +0.171 | **YES** |
| locality | local | 2 | 4 | 0.400 | 0.571 | −0.171 | . |
| uncertainty | **deterministic** | 4 | 5 | 0.800 | 0.714 | +0.086 | **YES** |
| uncertainty | probabilistic | 1 | 2 | 0.200 | 0.286 | −0.086 | . |

Both tables **match** the (l-coh) authoritative sections exactly. The
persisted candidates from the actual (l) miner runs (M7 →
`equals(locality, local)` at `surviving`; pvnp → three `weaken` outcomes on
`equals(construction, existential)`, `equals(locality, global)`,
`equals(uncertainty, deterministic)`) are exactly the emission-eligible
set from these tables.

## What this validates and what it does NOT

**Validates (correction 6 discharged)**:
- The (l-coh) numerical prevalences count the correct population (the
  miner-eligible family set), not merely reachable signatures.
- Row-to-family mapping: 1-to-1 in these databases (all clusters are
  singletons with `member_count = 1`, `redundant = 0`).
- Normalization versions and cluster-algo versions are consistent within
  each latest run (single `cluster_run` per problem for M7/pilot-001/
  pilot-002/pvnp; for pilot-003 the latest is selected deterministically).

**Does NOT validate**:
- Source-level lineage between the four ES corpora. The reviewer's
  correction 5 explicitly noted that identical marginals do not prove
  identical underlying observations. This validation strengthens the
  "identical marginals" SGO but does not discharge correction 5's
  independence question. Different underlying data can have the same
  histogram.
- The **research** interpretation of the eligible-family table. The
  reviewer's overarching frame stands: this is about the arithmetic and
  its population, not about within-population coherence or
  cross-corpus independence.

## Correction to the (l-coh) data-provenance narrative

The earlier (l-coh) result record said:

> pilot-003's counts are exactly 3× M7's; same underlying distribution
> scaled by pipeline HEAD-time.

Under the corrected query, **pilot-003's counts equal M7's exactly**
(not 3×). The earlier "3×" observation was an artifact of joining every
signature × posture-axis without restricting to the latest cluster_run
per problem. pilot-003 has 2 `cluster_runs`; the unrestricted join
counted signatures from both, producing 3× the eligible signatures
(each cluster_run has 12 members). Under the miner-eligible query, only
the latest run participates.

This is a smaller correction than it sounds: the corrected finding is
even more consistent with the reviewer's independence caveat. All four
ES corpora produce **identical** miner-eligible tables. Whether the
underlying markdown/signature/cluster data really is copy-of-copy or
just happens to produce identical marginals remains open — the
reviewer's source/snapshot lineage validation task is untouched by
this recomputation.

## Consequences for the scorecard

- (l-coh) **correction 6 discharged**: the eligible-family arithmetic
  is verified. (l-coh) tables stand.
- (l-coh) **correction 5 unchanged**: independence between corpus
  lineages remains not established. This validation makes the
  "identical marginals" fact stronger, not weaker; it does not
  substitute for source-lineage validation.
- (l-coh) other corrections (1–4, 7–9) unchanged.
- (l-coh-mechanism-test) CMA upgrade of correction 7 unchanged.
- **H3 gate: preserved.** No admission rules changed. No verifier
  tiers. No policy directives. No wire schema fields.
- Pure SGO validation using existing SQLite schema and code semantics.

## Anti-hindsight note

The script used is a self-contained Python one-shot invoked in this
session's terminal history. The methodology exactly mirrors
`derivePostureAxisProposals`'s eligibility gates (see the file's
`for _, fam := range req.Families` loop). Anyone with the databases
can reproduce it in seconds.
