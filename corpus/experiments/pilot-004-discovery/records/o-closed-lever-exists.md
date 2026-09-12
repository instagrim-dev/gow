# (o-closed) Concrete-falsifiability priority — lever already exists

Recorded: 2026-09-12. Pipeline-space (CMA + SGO).

## What (o) proposed to solve

The (g-scoped) investigation surfaced a design-space observation:
"the search policy could downweight proposals whose falsification
paths are non-concrete." User authorization on (g-scoped) explicitly
sanctioned bounded observations informing search priority.

(o) as first framed would have added a new policy directive — some
form of "penalize non-concrete falsification path." The task of this
loop was to predeclare (o) rigorously before authorization.

## CMA on the existing policy engine

`internal/policy/engine.go` and `internal/policy/apply.go` reveal
the pipeline already carries the concrete-falsifiability lever, and
carries it at exactly the right position.

**Directive vocabulary (verbatim from `engine.go:56-88`)**: every
policy directive targets a PERSISTED ROW by ID or predicate
fingerprint — success invariants, surviving invariants, mechanism
families, redundant attacks, mechanism fingerprints, refuted-
boundary predicates. Every target kind resolves to a real row (KTD-1).
Free-form falsification-path text is not a persisted-row target;
a directive keyed on it would fail the resolution contract.

**Ranking function (`apply.go:142-156`)**:

```go
func baseObjectiveLess(ca, cb frontier.Candidate) bool {
    if ca.MechanisticDistance.Rank() != cb.MechanisticDistance.Rank() {
        return ca.MechanisticDistance.Rank() > cb.MechanisticDistance.Rank()
    }
    if ca.ExpectedInformationGain.Rank() != cb.ExpectedInformationGain.Rank() {
        return ca.ExpectedInformationGain.Rank() > cb.ExpectedInformationGain.Rank()
    }
    if ca.EvaluationCost.Rank() != cb.EvaluationCost.Rank() {
        return ca.EvaluationCost.Rank() < cb.EvaluationCost.Rank()
    }
    return ca.ProposalHash < cb.ProposalHash
}
```

The ordering is: mechanistic distance (novelty) > info gain (learning
value) > **evaluation cost (concrete-falsifiability, cheaper wins)**
> hash. Concrete-falsifiability is a first-class tiebreaker.

**Falsifiability floor (`apply.go:160-190`)**: separately, the pipeline
protects the cheapest-to-falsify violating candidate per target
invariant. Verbatim from `apply.go:34-36`: *"the code-verified violation
gate is inviolable — a non-violating proposal can never outrank a
violating one no matter how strongly preferred — and every candidate
is retained."* The falsification surface is code-guaranteed.

## SGO on M7 to verify the lever is consumed

Ranked M7 proposals by `rank_ordinal`:

| id | eval_cost | info_gain | mech_dist | violates | rank |
|---|---|---|---|---|---|
| fpr_...STE4T95 | low | low | medium | 0 | 0 |
| fpr_...RS607H | low | low | medium | 0 | 0 |
| fpr_...DF6FXYG | low | medium | medium | **1** | 0 |
| fpr_...KD2YB1 | medium | high | medium | 0 | 0 |
| fpr_...ETFTE  | medium | high | medium | 0 | 0 |
| **fpr_...J3F0ECY** (my H2-escape) | **low** | high | medium | **1** | **0** |
| fpr_...XRDPYDJ | low | low | medium | 0 | 1 |
| fpr_...F2DEQAQ | low | medium | medium | 1 | 1 |
| fpr_...MDFAWA | low | low | medium | 0 | 2 |
| fpr_...HR7MVX3 | low | medium | medium | 1 | 2 |

The concrete-eval-cost lever is being consumed. My H2-escape
proposal, self-attested `evaluation_cost: low`, sits at rank 0 in
its violation tier — exactly what the ranker's design predicts for
a `low/high/medium/violates=1` tuple. No policy change is needed
to give concrete-falsifiable proposals priority; they already have
it.

## Verdict on (o)

**No code change required.** The lever the user's authorization
sanctioned already exists at the correct position in the ordering,
with a documented boundary (never dominates the violation gate,
never breaks falsifiability). Adding a new directive kind for
"non-concrete falsification path" would either:

1. Duplicate the existing `EvaluationCost` signal — proposals with
   concrete falsification paths ALREADY self-attest lower
   evaluation_cost; the ranker already prefers them. A duplicative
   directive would be redundant.
2. Rank on falsification-text patterns — model-judgment over prose,
   masquerading as code-owned discrimination. KTD-3 forbids this.
3. Rank on a new declared wire field — wire schema evolution, and
   the new field would encode the same information `evaluation_cost`
   already does.

## Where the actual research signal lives

The observation from (g-scoped) was: *"none of the persisted target-
violating proposals has a candidate-specific witness obligation."*
Reading that observation carefully:

- The **ranker's** concrete-falsifiability lever is a signal about
  proposal ORDERING within a corpus.
- The **corpus-level** observation from (g-scoped) is a signal about
  the WHOLE POPULATION of proposals: none is concrete enough to enter
  the witness path.

These are different objects. The ranker cannot fix a corpus that has
no concrete proposals — it can only choose the cheapest-to-evaluate
proposal from among those available. The corpus-level signal is about
what the GENERATOR (or the human wire authors) produce, not about how
the ranker orders them.

## Consequence for authoring discipline (recorded, not proposed as change)

Wire authors have a legitimate handle: `evaluation_cost` self-
attestation. An honestly-authored concrete proposal declares
`evaluation_cost: low` when a witness tuple or exact check is
derivable from its mechanism; a mechanism-family claim requiring
domain math declares `medium/high`. The ranker will then favor the
concrete when both exist. This is the pipeline's designed
integration point for the observation from (g-scoped): the operator's
authoring judgment IS the code-owned discriminator, expressed via
the self-attested ordinal.

## Consequence for the frontier generator (recorded, not proposed as change)

The stronger, longer-horizon research move is not a ranker change
but a GENERATOR change: bias the frontier generator toward
mechanisms that CAN produce concrete witnesses. For ES specifically,
the atlas contains mechanism families (es-01 Mordell polynomial
identities, es-03 factorization scheme) that produce witness tuples
for specific residue classes. A generator that seeded new proposals
from those families would produce concrete-witness-path candidates
naturally. Recorded for downstream authority; not proposed as change.

## H3 gate status: PRESERVED (unchanged)

- No admission rules changed.
- No new verifier tier added.
- No policy directive kind added.
- No wire schema field added.
- The ranker's evaluation_cost tiebreaker continues to operate as
  the pipeline authors designed it.

## Verification tier

- Code trace of policy engine directive vocabulary: **CMA** (verbatim
  reading of `engine.go` and `apply.go`).
- Existing evaluation_cost lever consumed by the ranker: **SGO**
  (verbatim sqlite output showing consistent rank_ordinal assignment
  matching `baseObjectiveLess` on the ordinal tuple).
- Verdict "no code change required": **CMA** (deterministic consequence
  of the two observations above — the lever exists, is consumed, and
  is at the right position; a new directive would duplicate or launder).
- Corpus-level observation about the current proposal population:
  **SGO** (a summary of (g-scoped)'s findings, unchanged here).
- Authoring-discipline observation: **PE** (a plausible research-
  methodology hypothesis; recorded, not proposed as change).
- Generator-bias observation: **PE** (a longer-horizon design-space
  observation; recorded, not proposed as change).
