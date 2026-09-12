# (l) Cross-domain replication result — P vs NP holdout corpus

**Status**: executed 2026-09-12 at HEAD `324ec5d`. Predeclaration:
`records/l-pvnp-replication-predeclaration.md`.

## Result: **Reading B (partial replication)**

The (a‴-B) enum-axis miner emitted candidate predicates on the P vs NP
corpus with the same emission rule as ES; challenge disposition
diverged — where ES's `equals(locality, local)` reached `surviving`
with 100% failure-side specificity, all three P vs NP enum-axis
candidates reached `weaken` with success-side prevalences comparable
to or exceeding failure-side coverage.

**The mining mechanism generalizes; the specific axis-signature
discriminating power is corpus-specific.**

## Method (verbatim replay)

DB: `.newf/pvnp/newf.db` at HEAD `324ec5d`.

```text
newf init "P vs NP historical holdout (train)"        → prb_...EDST
newf init "P vs NP algorithmic method (target)"       → prb_...RJD2
newf ingest ./corpus/experiments/pvnp-holdout/train   → 12 sources
newf ingest ./corpus/experiments/pvnp-holdout/target  → 2  sources
newf normalize --provider fixture                     → 12 approaches on train
newf mechanism signature --vocab-version mechanism/v3 → 12 signatures
newf cluster build                                    → 12 singletons
newf failure-space build                              → fsp_...AQJN
                                                        distinct_family_count=12
                                                        redundant_member_count=0
                                                        outcome: 5 failure + 7 partial_success
newf invariants mine --emit-posture-axes              → 3 candidates persisted
newf challenge --all                                  → all 3 → weaken
```

## Persisted candidates (SGO, verbatim)

| ID | Predicate | Support | Failure cov | Success prev | State | Association |
|---|---|---|---|---|---|---|
| `inv_...F6M7FKR` | `equals(construction, existential)` | 2 | 2/5 (40%) | 5/7 (71%) — miner-time; challenge shows 2/7 preserve → weaken | weaken | contrast_observed |
| `inv_...GEPTA22` | `equals(uncertainty, deterministic)` | 4 | 4/5 (80%) | 5/7 preserve → weaken | weaken | contrast_observed |
| `inv_...HG59EMN` | `equals(locality, global)` | 3 | 3/5 (60%) | 3/7 preserve → weaken | weaken | contrast_observed |

Challenge campaign detail per candidate:

- **known-counterexample**: `unconfirmed` (a `contrast_observed` claim
  is not refuted by isolated counterexamples).
- **success-preserving**: `confirmed` with `contrast-collapse` delta
  on each — this is what drove the `weaken` transition.
- **bias-critique**: `unconfirmed` (support holds under recomputation).

## Comparison to (i) — the cross-corpus ES replication

| Corpus | `equals(locality, ?)` result | `equals(construction, ?)` result | Headline |
|---|---|---|---|
| M7 (ES) | `equals(locality, local)` support=3, 0/5 success → **surviving** | `equals(construction, constructive)` support=3, 1/5 success → **weaken** (atlas-truthful per (h)) | (a‴-B) enum-axis escape verified |
| pilot-001 (ES) | `equals(locality, local)` reproduced → surviving (per (k-restricted)) | reproduced → weaken | (i) headline: blind spot closed |
| pilot-002 (ES) | reproduced → surviving | reproduced → weaken | (i) headline |
| pilot-003 (ES) | reproduced → surviving | reproduced → weaken | (i) headline |
| **pvnp** (P vs NP) | `equals(locality, global)` support=3, 3/7 success → **weaken** | `equals(construction, existential)` support=2, 2/7 success → **weaken** | **(l): mechanism replicates, discrimination corpus-specific** |

Key structural difference: ES's failed approaches predominantly attack
the problem from a `local, constructive` posture (per-residue-class
methods, explicit identities). The failure-side forms a coherent
posture cluster with a nearly-empty success-side prevalence for the
same axis-values, driving the `surviving` disposition. P vs NP's
failed approaches span the full posture grid — relativization is
`global, existential`; monotone circuit bounds are `local, constructive`;
AC⁰[p] is `local, constructive`; GCT is `global, constructive-ish`;
etc. No single posture-axis value is a coherent failure-signature at
the pre-cutoff barrier programs.

## Predeclared readings — disposition

- **Reading A** (full replication with surviving state): NOT obtained.
  No pvnp enum-axis candidate reached `surviving` after challenge.
- **Reading B** (partial replication — mechanism generalizes, axis
  distribution differs): **OBTAINED**. Three candidates emitted with
  the same emission criterion, all weakening under the same
  success-preserving probe, driven by cross-domain differences in
  failure-structure coherence.
- **Reading C** (non-replication — miner emitted nothing): NOT
  obtained.
- **Reading D** (pipeline failure): NOT obtained.

The `--json invariants mine` output initially showed 0 candidates,
which briefly looked like Reading C. That was a misreading: the
`--json` output enumerates fresh candidates from a run, not persisted
rows; the persisted `candidate_invariants` table contains all 3.

## Semantic-content interpretation (PE-tier, not CMA)

The enum-axis miner is measuring **failure-structure coherence along
posture axes**. In ES, that coherence is high (the classical barriers
to ES-style representations concentrate on local-modulus-explicit
attacks). In P vs NP, that coherence is low at the pre-2005 window
because each barrier program (relativization, natural proofs, GCT
stall, etc.) posts a structurally different mechanism family and is
NOT a homogeneous "attempted approach" the way ES failure records are.

This tells us something the ES corpus alone could not: **the
enum-axis miner's power depends on the failure population being a
coherent class in posture-axis space.** When failure-space
heterogeneity is high (many distinct mechanism families, each with
its own posture profile), the miner still emits candidates but
challenge correctly weakens them because success-side preservation
signals cross-cut posture-axis membership.

This is a **domain-general property of the mechanism**, not a
domain-specific outcome. It clarifies (i)'s interpretation: the
recurring enum-axis invariants in ES corpora are informative about
ES's failure-structure coherence, not about a universal
failure-invariant-mining law.

## Verification tier

- Persistence of 3 candidates: **SGO** (verbatim `sqlite3 candidate_invariants`).
- Weaken transitions with `contrast-collapse` delta: **SGO**
  (verbatim `newf challenge` output).
- Cross-corpus comparison with (i) result: **SGO** on tabulated states.
- Structural interpretation ("ES has coherent failure posture; P vs NP
  does not"): **PE** (proposed explanation). Atlas-truthfulness for
  P vs NP is outside my epistemic reach; the operator-attested pvnp
  corpus is the authority.
- Domain-general property claim: **PE**, one-corpus replication.
  Full CMA would require additional non-ES, non-P-vs-NP corpus.

## What this DOES establish

- The (a‴-B) enum-axis miner mechanism operates correctly on a
  non-ES corpus (P vs NP). No implementation bug in cross-domain
  application.
- The (i) headline result (perfect cross-corpus replication of
  `equals(locality, local)` at `surviving` across ES corpora) is a
  **corpus-property finding**, not a universal invariant-mining
  result. It reflects ES's failure-structure homogeneity along
  posture axes.
- **Reading B outcome refines the (i) claim**: enum-axis mining is
  a domain-general mechanism whose recurring-invariant productivity
  depends on failure-population coherence in posture-axis space.

## What this does NOT establish

- Any claim about P vs NP itself.
- Any claim that pvnp's negative discrimination result predicts a
  positive or negative on other non-ES corpora — one non-ES corpus
  is one observation, not a distribution.
- Discharge of any review-integration obligation (C7, C8,
  witness-occurrence attribution) — those are orthogonal.
- Freeze-readiness of pilot-005-relational — that pilot requires
  operator authorization on six open items and is not exercised by
  this run.

## H3 gate status: PRESERVED

- No admission rules changed.
- No new policy directives.
- No verifier tier changes.
- No wire schema field additions.

This is a corpus-execution replication test, not an H3 escape attempt.

## Follow-on questions (recorded, not proposed)

- **Third-corpus check (l-extension)**: if a non-ES, non-P-vs-NP
  corpus were materialized (e.g. one of the existing draft protocols
  or the pilot-005-relational once frozen), would it produce more
  Reading-B outcomes, or would it reproduce Reading A? Without a
  distribution over corpora, "corpus-property finding" is a
  one-observation claim.
- **Coherence-metric derivation**: could the pipeline itself compute
  a failure-structure-coherence score that predicts enum-axis
  miner productivity, so a corpus's amenability to enum-axis
  invariants is machine-observable before challenge? This would be
  a new pipeline capability, not a bug fix. Design-space only;
  needs authorization.
- **Adversarial-population contamination check**: the pvnp corpus
  was decontaminated per `008b0c2`. If contamination were
  reintroduced (post-cutoff barrier programs leaking into
  pre-cutoff train), would the failure-population become more
  coherent (spurious Reading A) or less coherent (stronger Reading
  B)? Not investigated in this loop.
