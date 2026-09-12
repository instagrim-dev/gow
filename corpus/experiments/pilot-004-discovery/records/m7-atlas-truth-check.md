# (h) atlas-truth check on the `equals(construction, constructive)` weakening

Recorded: 2026-09-12. Pipeline-space + atlas-quality check (SGO).

## Question

The (a‴-B) run weakened `equals(construction, constructive)` because the
challenge's success-preserving probe found c6 as a success-side family
satisfying the predicate. Was c6's `posture.construction=constructive`
atlas-truthful, or was the weakening premature?

## Result: atlas-truthful (SGO)

**c6 = `erdos-straus/type-a-b-congruence-system` (es-12).**

Verbatim from `corpus/train/es-12-type-a-b-congruence-system.md`:

- Commentary (para 3): *"each form comes with an explicit polynomial,
  and by construction none of the admitted congruences is a quadratic
  residue"*; *"constructive-per-form"*.
- newf-normalize block: `"construction_mode": "constructive"`.

es-12 IS legitimately constructive: each Type A/B solution form is
associated with an explicit polynomial that produces the ES witness
for that form. The mining engine's classification is atlas-truthful,
and the challenge's weakening on `equals(construction, constructive)`
is **atlas-correct**. Path (h) closes without an atlas remediation.

## Deeper observation surfaced by the survey (CMA)

Full atlas × posture cross-table on M7 (SGO):

| c | locality | construction | atlas | outcome |
|---|---|---|---|---|
| 0 | local | constructive | es-03 factorization-scheme | partial_failure |
| 1 | global | existential | es-09 higher-dimensional-variety-lift | partial_success |
| 2 | global | existential | es-04 averaging-bv-count | partial_failure |
| 3 | local | constructive | es-07 computational-verification | partial_failure |
| 4 | local | constructive | es-01 mordell-polynomial-identities | partial_failure |
| 5 | global | existential | es-08 quadratic-reciprocity-obstruction | failure |
| 6 | global | **constructive** | **es-12 type-a-b-congruence-system** | **partial_success** |
| 7 | global | existential | es-02 covering-system-attempt | failure |
| 8 | global | existential | es-06 two-fraction-representation-bounds | partial_success |
| 9 | mixed | existential | es-10 monks-velingker-structure | partial_success |
| 10 | mixed | existential | es-11 type-i-ii-classification | partial_success |
| 11 | mixed | existential | es-05 vaughan-congruence-density | partial_failure |

**Observation A — axis co-variation on the failure side.** On M7, the
three failure-side `construction=constructive` families (c0, c3, c4)
are EXACTLY the three failure-side `locality=local` families. On this
corpus the axes are perfectly co-varying on the failure side. This means:

```
All(equals(locality, local), equals(construction, constructive))
  reduces to
equals(locality, local)
```

on M7. The joint predicate CANNOT beat single-axis `locality=local`
on this data — it produces the same 3-family support AND additionally
picks up the c6 (partial_success) counterexample, weakening the joint
below where the single-axis predicate stands.

This is a compact concrete argument that **conjunctive posture
predicates, if the miner were extended to emit them (a hypothetical
(a‴-B') extension), would not necessarily strengthen the invariant
set on this corpus.** The value of conjunctive predicates depends on
axis independence — a corpus property not present here.

**Observation B — the three failure-side `local/constructive` entries
are a semantically-coherent class.** All three (es-01, es-03, es-07)
are **local-modulus explicit-construction attempts**:

- es-01 (Mordell polynomial identities): "For many residue classes of
  n, a fixed polynomial identity in n writes..." — per-residue-class
  polynomial identity.
- es-03 (factorization scheme): "Starting from 4/n = 1/x + 1/y + 1/z
  and clearing denominators, one seeks solutions through an identity
  of the form..." — explicit factorization form.
- es-07 (computational verification): "checks the conjecture...
  n = 10^17..." — case-by-case explicit witness.

The success side has ZERO local mechanisms. This is essentially the
paper-space N2a chain and density-and-averaging finding, re-expressed
on the enum axis via support/contrast counting rather than via the
CMA covering-density arithmetic. The mining engine, extended with
posture-axis emission, reproduces the paper-space discovery through
a completely different route.

## Interpretation

The `equals(locality, local)` surviving invariant compresses a
well-known number-theoretic pattern: **local-modulus explicit-
construction attempts fail on ES**. Success requires either global
scope (with existential construction, per es-06/es-09) or mixed scope
(with existential construction, per es-10/es-11) or the specific
constructive-global exception (es-12, which is self-declared
conjectural).

The atlas has been correctly labelled. The mining engine, once
authorised to emit enum-axis proposals, discovers the pattern
already present in the paper-space records. This is a genuine
convergence: the same substantive finding arrived at through two
independent routes (CMA on covering-density arithmetic in the paper
space; support/contrast counting on posture axes in the pipeline
space).

## Follow-on observations (design-space, not proposed as changes here)

1. **Axis co-variation on discovery is a corpus signal.** If two
   posture axes are perfectly co-varying on the failure side, the
   engine could report this as a "structural redundancy" hint — the
   two axes are not measuring independent things on THIS problem.
   Not a change proposal; recorded for downstream authority.

2. **The mining engine's convergence with paper-space CMA
   discoveries is testable in principle.** For every paper-space
   invariant recorded across the pilots, one could ask: does the
   enum-axis-extended miner produce a support/contrast-matching
   invariant when it runs on the same corpus? If yes, the miner is
   a viable independent verification route for paper-space findings.
   If no, either the miner is missing some predicate shape or the
   paper-space finding is not corpus-visible. Not a change proposal
   for this loop; a research-methodology observation.

## Verification tier

- c6 atlas identity + posture: **SGO** (verbatim sqlite output + atlas
  .md content).
- Weakening is atlas-correct: **CMA** (direct consequence of the SGO
  above — c6 is partial_success AND construction=constructive, exactly
  what the success-preserving probe checks).
- Axis co-variation observation: **CMA** (deterministic counting on the
  posture cross-table).
- Semantic-class observation: **CMA** with **SGO** references (verbatim
  commentary from each atlas .md).
- Mining↔CMA convergence claim: **PE** (a plausible research-methodology
  hypothesis; would require running the extended miner on other
  pilots' corpora to confirm).

## Consequences for the closure scorecard (proposed)

- Path (h) closes; no atlas remediation needed.
- The observation about axis co-variation is recorded as a
  design-space observation for downstream authority (does not propose
  a code change here).
- The observation about mining↔CMA convergence is recorded as a
  potential future research-methodology check (does not propose
  execution here).
