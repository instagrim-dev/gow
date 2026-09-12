# (l-coh) Coherence measurement result — COH-A numerical conditions not met; disposition NARROWED after reviewer correction

**Status**: executed 2026-09-12 at HEAD `00c3c58`; reviewer-corrected
same-day (record superseded by the "Corrected disposition" section
below; the earlier bottom-line section is retained for lineage).
Predeclaration: `records/l-coh-measurement-predeclaration.md`.

## Corrected disposition (authoritative)

> **COH-A's predeclared numerical conditions were not met.** ES had
> greater maximum posture-axis outcome discrimination than P-vs-NP
> in the reported tables, but this does not establish within-failure
> coherence, and the alternative reading rules do not uniquely
> classify the result. The four ES databases should not be counted
> as four independent research replications; source-level lineage
> and population-query validation remain necessary. The zero-cell
> observation suggests a conditional explanation through the
> existing success-preserving challenge rule, not an unconditional
> law of invariant survival. Its implementation implications can be
> tested directly; its usefulness for research requires separate
> prospective evidence.

Wording above adopted from the reviewer's suggested conclusion.

Concrete corrections applied:

1. **"Falsified" is the wrong description of the statistic.** The
   statistic produced a value. COH-A's threshold or explanatory role
   was rejected; the statistic itself was not falsified.
2. **Outcome discrimination is not within-population coherence.**
   The measure computes the largest between-outcome prevalence
   difference. Every failure and every success could be `local`
   (perfect within-failure homogeneity, zero discrimination). The
   operationalization substituted outcome discrimination for
   within-population coherence. The 2.5× ratio is arithmetically
   correct; it does not mean ES is "2.5× more coherent."
3. **COH-B is not sufficiently defined to classify the result.**
   The predeclaration uses "high", "low", "comparable" without
   numerical definitions and overlaps COH-C. My earlier
   justification (correct ordering + weaker magnitude) is not the
   stated decision rule. **COH-B/COH-C adjudication is
   underdetermined by the predeclaration.** Retrospectively
   repairing those rules while retaining the "predeclared
   classification" label is not permitted.
4. **The predeclaration's threshold-robustness assertion is
   falsified by the actual tables.** The claim that thresholds
   ≥ 0.4 preserve the qualitative pattern is wrong: at 0.4, M7 has
   one qualifying axis-value and pvnp has none; at 0.5, both have
   none. This does not invalidate the frozen 0.5 calculation. It
   invalidates the claimed threshold robustness.
5. **"n=2 truly independent observations" is too strong.** Identical
   aggregate distributions do not themselves prove identical
   underlying observations; different datasets can have the same
   histogram. Exactly tripled counts do not identify the cause
   (repeated source data, retained interpretation revisions, or
   join multiplicity can each produce that pattern). Independence
   between the two corpus lineages is not established. Defensible
   replacement:

   > Five database instances represent two substantive corpus
   > lineages for this comparison. The four ES instances do not
   > provide four independent research replications. Independence
   > between the two corpus constructions has not been established.

   **ES-side residual DISCHARGED 2026-09-12** by
   `(source-lineage-check)`. Positive source-byte evidence: the
   four ES databases share 12 of 13 `source_snapshots.sha256`
   values byte-for-byte, with identical `sources.logical_name`
   values on each shared source. The single non-shared row is
   M7's target file. See
   `records/source-lineage-check-result.md`. ES-vs-pvnp
   independence residual remains open (different source corpora,
   but "independence" in the stronger research sense requires
   more than a sha256 diff).

6. **Population-query validation is a prerequisite the record did
   not discharge.** Family prevalence must count the eligible
   families used by the miner and challenger, not every persisted
   signature. The documented query joins through signatures and
   approach revisions; it does not show how the measured rows were
   restricted to the actual selected cluster population or
   deduplicated into families. Arithmetic can be correct for the
   wrong population. Retained SQL, source/snapshot lineage,
   selected cluster-run IDs, normalization versions, and
   row-to-family mapping remain necessary before making the
   provenance conclusion definitive.

   **DISCHARGED 2026-09-12** by
   `(l-population-validation)`. A deterministic recomputation
   using the exact eligibility semantics of
   `derivePostureAxisProposals` (latest `cluster_run` per problem,
   `mechanism_clusters` representative signatures, `outcome_mixed`/
   `unknown` outcomes excluded, redundant-only clusters excluded,
   `unknown` posture values excluded) reproduces the numerical
   tables in this record exactly for M7 and pvnp. All four ES
   corpora produce identical miner-eligible tables under this
   query. The tables' values are correct for the correct
   population. See
   `records/l-population-validation-result.md`.
7. **The zero-cell "iff" overgeneralizes.** The code-derivable
   relationship (from `VerifySuccessPreserving` inspection) is:
   the probe confirms exactly when at least one eligible success
   matches the predicate, in the fully-resolved singleton case.
   **"No preserving success" is not equivalent to "reaches
   surviving."** The campaign can still falsify or weaken through
   another probe; survival additionally requires at least one
   completed negative and no overriding confirmed falsification.
   Also, the posture-axis miner's emission gate requires ≥ 2
   distinct failure-side families and `fPrev > sPrev`, so ineligible
   axis-values do not participate. The useful refinement is:

   > Among emitted, support-valid candidates with comparable
   > resolved populations, absence of a preserving success
   > prevents this particular weakening route. Overall survival
   > additionally depends on the remaining challenge outcomes and
   > claim semantics.

   This explains an implementation mechanism (code-derivable). It
   does not show that zero-cell selection identifies
   scientifically useful invariants; nor does zero matches among
   five observed successes establish zero prevalence beyond the
   sample.

   **UPGRADED 2026-09-12 to CMA-tier** by
   `(l-coh-mechanism-test)`. A controlled fixture test at
   `internal/invariant/success_preserving_discrim_test.go` shows
   `VerifySuccessPreserving` returns `OutcomeCompletedNegative` on
   a fully-resolved singleton fixture with sPrev = 0 and
   `OutcomeConfirmed` on a fixture with sPrev = 0.3, holding
   `fPrev - sPrev = 0.5` fixed. Both predeclared conditions plus
   Delta/Evidence structure hold. See
   `records/l-coh-mechanism-test-result.md`.
8. **Attribution correction on the numerical minimum.** The
   earlier record attributed pvnp's `minSuccessPrev = 0.286` to
   `equals(uncertainty, probabilistic)`. That candidate fails
   BOTH miner emission gates (fc = 1 < 2; and fPrev = 0.200 <
   sPrev = 0.286, so the discrim > 0 gate also fails). The
   correct emission-eligible attribution is
   `equals(construction, existential)` at the same numerical
   value 2/7 = 0.286. The pvnp candidates actually persisted by
   the (a‴-B) run — `equals(construction, existential)`,
   `equals(locality, global)`, `equals(uncertainty, deterministic)` —
   are exactly the emission-eligible set. M7's minimum-sPrev
   attribution to `equals(locality, local)` at 0.000 was already
   correct (that candidate IS emission-eligible).
9. **Post-hoc discovery needs qualification, not an epistemic
   prohibition.** The earlier record said the post-hoc hypothesis
   must be predeclared on a third corpus to earn any epistemic
   status. That conflates discovery timing, evidence type, and
   strength of support. A post-hoc observation can motivate a
   valid deduction about code. A preregistered prediction can
   still use the wrong population. A third corpus provides
   prospective evidence about research usefulness; it does not
   turn an empirical generalization into an arithmetic or
   code-level deduction. The proposed promotion of the coherence
   explanation to SGO upon obtaining COH-A had the same
   category error: matching a threshold does not transform an
   explanatory inference into a directly observed fact. Keep
   counts, calculations, code implications, and research
   interpretation separately attributed.

## Emission-eligible recomputation (SGO)

Verified 2026-09-12 by `(l-population-validation)`; population is the
miner's actual eligible-family set (see
`records/l-population-validation-result.md`).

| Corpus | Candidate (emission-eligible) | sPrev | Persisted at (l)? |
|---|---|---:|---|
| M7 | `equals(construction, constructive)` | 0.200 | (not present in the final `candidate_invariants` table; see M7 clarification below) |
| M7 | `equals(locality, local)` | **0.000** | **YES, → `surviving`** |
| pvnp | `equals(construction, existential)` | **0.286** | YES, → `weaken` |
| pvnp | `equals(locality, global)` | 0.429 | YES, → `weaken` |
| pvnp | `equals(uncertainty, deterministic)` | 0.714 | YES, → `weaken` |

**M7 clarification**: `equals(construction, constructive)` is
emission-eligible (fc = 3, fPrev > sPrev), but the (a‴-B) M7 run
persisted only `equals(locality, local)` at `surviving`. This
divergence between "would-be-eligible" and "actually persisted"
belongs in the campaign-run investigation (why only one of two
eligible candidates survived to persistence at that HEAD), not in
(l-coh)'s scope. The eligibility semantics themselves are verified.

## Reviewer's message preserved verbatim

> **Accept the threshold rejection and the correction to the
> replication claim. Do not yet accept "COH-B obtained," "two
> truly independent observations," or the unconditional zero-cell
> `iff` as established conclusions.** The strongest new insight
> is probably about **how the challenge policy converts corpus
> composition into lifecycle status**, not about research
> coherence.

Full reviewer text is preserved in this session's chat log; the
corrections above adopt each of its numbered concerns.

## Concrete next test (reviewer-suggested, not yet executed)

Construct two resolved singleton-family fixtures with the same
discrimination difference:

| Fixture | Failure prevalence | Success prevalence | Discrimination |
|---|---:|---:|---:|
| A | 5/10 | 0/10 | 0.5 |
| B | 8/10 | 3/10 | 0.5 |

Keep emission/support gates satisfied and the claim and challenge
configuration fixed. The `success-preserving` probe should
distinguish these cases despite identical discrimination. That
tests whether the proposed **code-mechanism** explanation captures
the pipeline's behavior. Testing whether that behavior predicts
useful research outcomes is a separate experiment.

This is a code-behavior test, not a corpus replication. Its result
would upgrade the correction-7 refinement from PE (interpretive)
to CMA (code-level deduction), without requiring a third domain
corpus.

## The distinction that matters most

> **The most important remaining distinction is between
> discovering structure in research and rediscovering the decision
> rule that labels that structure `surviving`.**

The earlier record's post-hoc `minSuccessPrev` observation is
probably an instance of the latter, not the former. That framing is
adopted here as the record's operating summary.

## Earlier bottom-line (retained for lineage; SUPERSEDED by the corrected disposition above)

**⚠ The three claims below were written before the reviewer
correction. They are retained for provenance ONLY. Do not cite them
as this record's conclusions; use the "Corrected disposition"
section at the top instead.**

**~~Predeclared Reading COH-A is FALSIFIED.~~** *[Corrected: the
statistic produced a value; COH-A's numerical conditions were not
met, but "falsified" mis-describes the statistic itself. See
correction 1.]* The threshold of `maxDiscrim ≥ 0.6` for ES corpora
is not met (M7 = 0.429). The predicate that DID reach `surviving`
in the miner's actual run has a maxDiscrim of 0.429, below the
predeclared floor. *[The "coherence framing needs to be weakened"
conclusion is retained but sharpened: the statistic measured
outcome discrimination, not within-population coherence — see
correction 2.]*

**~~Predeclared Reading COH-B is OBTAINED.~~** *[Corrected:
withdrawn. COH-B is not sufficiently defined by the predeclaration
to classify this result. See correction 3. Adjudication between
COH-B and COH-C is underdetermined. The 2.5× ratio observation is
retained as arithmetic but does not warrant either label.]*

**~~Post-hoc signal (recorded as hypothesis, NOT established)~~**:
*[Corrected: the framing "requires predeclared replication on a
third corpus to earn epistemic status" conflated discovery timing,
evidence type, and strength of support. See correction 9. The
correct treatment is: absence of a preserving success is a
code-derivable necessary condition for the success-preserving
weakening route (via `VerifySuccessPreserving` inspection), not an
`iff` for `surviving`. See correction 7. The attribution of pvnp's
0.286 minimum has been corrected from `uncertainty=probabilistic`
(ineligible) to `construction=existential` (eligible, same
numerical value). See correction 8.]*

## Measurement (SGO)

Query: filtered per-train-problem posture distribution joined through
`mechanism_signatures → mechanisms → approach_revisions → approaches
→ problems`; success-side = `partial_success + success`; failure-side
= `failure + partial_failure`. Verbatim SQL and outputs in this
session's terminal record.

### Data-provenance note

- M7, pilot-001, pilot-002, **and pilot-003** have **identical**
  posture distributions under the miner-eligible query
  (`internal/provider/invariant_fixture_derive.go:derivePostureAxisProposals`
  semantics: latest `cluster_run` per problem, `mechanism_clusters`
  representative signatures only, `outcome_mixed`/`unknown` excluded,
  `redundant`-only clusters excluded, `unknown` postures excluded).
  Under this query pilot-003 = M7, not 3× M7. The earlier "3×"
  observation in this record was an artifact of joining every
  signature × posture-axis without restricting to the latest
  cluster_run; pilot-003 has two `cluster_runs`, the others one.
  See `records/l-population-validation-result.md`.
- **What the SGO establishes**: the four ES database instances
  produce identical miner-eligible marginals under a query that
  faithfully implements the miner's eligibility semantics. This
  strengthens the "consistent-with-shared-lineage" observation.
  They should not be counted as four independent research
  replications under this measurement.
- **What the SGO does NOT establish**: that only two truly
  independent observations exist. Different underlying data can
  produce the same histogram. Independence between the two corpus
  lineages (ES-shape and P-vs-NP-shape) has not been established
  either. Defensible replacement (unchanged from earlier):

  > Five database instances represent two substantive corpus
  > lineages for this comparison. The four ES instances do not
  > provide four independent research replications. Independence
  > between the two corpus constructions has not been
  > established.

- **Family prevalence arithmetic is now verified** to count the
  eligible families the miner and challenger actually operated on.
  See (l-population-validation) result record. Correction 6 from
  the reviewer's list is discharged; the values in the tables
  below are correct for the correct population.

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
- **minSuccessPrev(post-hoc)** = 0.286 at `equals(construction, existential)` *(emission-eligible; the earlier record's attribution to `equals(uncertainty, probabilistic)` was an ineligible-candidate error — see correction 8. Numerical value 2/7 is coincidentally the same.)*

## Predeclared readings — disposition (SUPERSEDED; see Corrected disposition at top of record)

- **Reading COH-A (structural claim confirmed)**: numerical
  conditions not met (M7 maxDiscrim = 0.429 < 0.6 threshold).
  Earlier framing "FALSIFIED" is retained here for lineage but is
  mis-descriptive per correction 1.
- **Reading COH-B (partial support)**: **WITHDRAWN.** COH-B is not
  sufficiently defined by the predeclaration to classify this
  result; earlier "OBTAINED" classification is retracted per
  correction 3.
- **Reading COH-C (structural claim falsified)**: adjudication vs
  COH-B is underdetermined by the predeclaration.
- **Threshold-robustness assertion** ("≥ 0.4 preserves the
  qualitative pattern"): falsified by the tables per correction 4.

## Correlation with miner productivity (SGO)

- M7 miner (a‴-B) run produced `equals(locality, local)` at
  `surviving`.
- pvnp miner (a‴-B) run produced 3 candidates, all reaching `weaken`.

Under the predeclared coherence measure, neither corpus should reach
`surviving` (both have `posDiscrimCount(≥0.5) = 0`). This is another
falsification signal for the predeclared measure: M7 DID produce a
`surviving` invariant despite `maxDiscrim = 0.429 < 0.5`.

## Post-hoc hypothesis (SUPERSEDED by correction 7; see top-of-record)

**⚠ The unconditional `iff` below is retracted. The reviewer's
correction gives the narrower, code-derivable statement adopted by
this record. The section is retained for lineage.**

The `minSuccessPrev` observation was originally recorded as a
candidate refinement of the (l) PE claim. Under the retracted
hypothesis:

> ~~**Enum-axis mining reaches `surviving` on `equals(axis, value)` iff
> the success-side prevalence of that `(axis, value)` is exactly zero.**~~

The refined, code-derivable statement (per correction 7):

> Among emitted, support-valid candidates with comparable resolved
> populations, absence of a preserving success prevents this
> particular weakening route. Overall survival additionally depends
> on the remaining challenge outcomes and claim semantics.

This maps directly to the code-owned `success-preserving` probe
(`internal/pipeline/invariant.go`, `internal/provider/invariant_challenger.go`):
`VerifySuccessPreserving` confirms exactly when at least one
eligible success family satisfies the predicate (in the
fully-resolved singleton case). "No preserving success" is
NECESSARY but not sufficient for `surviving`; the campaign can
still falsify or weaken through another probe.

Support (SGO on 2 corpus lineages, NOT independent replications):
- M7: `equals(locality, local)` has sPrev = 0.000 → `surviving`.
- pvnp: minimum sPrev over emission-eligible candidates = 0.286 at
  `equals(construction, existential)` → all candidates `weaken`.

**Why this is a code-behavior claim rather than a research claim:**

The narrower statement above is derivable from `VerifySuccessPreserving`'s
implementation via CMA inspection. It does not require a third
corpus. The **research** question — whether zero-cell selection
identifies scientifically useful invariants — is separate and
requires prospective evidence.

The earlier record's post-hoc epistemic-prohibition framing ("must
be predeclared on a third corpus") was itself confused (per
correction 9): it conflated discovery timing, evidence type, and
strength of support. A code-derivable relationship is code-derivable
regardless of when it was noticed; a prospective replication tests
usefulness, not the deductive link.

## Update to the (l) result record

The (l) result's structural finding —

> *the enum-axis miner's recurring-invariant productivity depends on
> failure-population coherence in posture-axis space*

— should now be read as:

- **Direction consistent with earlier interpretation**: ES had
  stronger between-outcome discrimination than pvnp on the enum
  axes (0.429 vs 0.171 on maxDiscrim).
- **Does not establish within-failure coherence**: the statistic
  measures between-outcome discrimination, not within-population
  coherence. Every failure and every success could share a value
  (perfect homogeneity, zero discrimination). See correction 2.
- **Explanation via challenge policy, not via "coherence"**: the
  actual driver of the ES-vs-pvnp `surviving` vs `weaken`
  distinction is the `success-preserving` probe's binary threshold,
  not a magnitude of discrimination. That is a code-level
  observation about how the pipeline converts corpus composition
  into lifecycle status — the reviewer's phrasing.

The (l) record itself is not being rewritten (this record's job is
to record the correction). The scorecard entry annotates the
narrowed warrant.

## What DOES this loop establish

- SGO measurement of failure/success posture distributions on 5
  persisted databases, under a documented query.
- Consistent-with-shared-lineage SGO signal: identical distributions
  across M7/pilot-001/pilot-002 under the query; pilot-003 tripled.
  This weakens (i) from "perfect replication across 4 corpora" to
  "consistent with one observation replicated across 4 HEADs plus
  1 different domain," pending source/snapshot lineage validation.
- **What we can say without over-claiming**: the four ES database
  instances should not be counted as four independent research
  replications under this measurement. Independence between the
  two corpus constructions is not established either.
- COH-A's predeclared numerical conditions were not met.
- Code-derivable relationship: absence of a preserving success is a
  necessary condition for the success-preserving weakening route;
  survival additionally depends on other campaign outcomes.
- Attribution correction on the numerical minimum (correction 8).

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
