# (o-closed) Concrete-falsifiability priority — no additional ranking directive justified for this iteration

Recorded: 2026-09-12.
**2026-09-12 revision** — this record was rewritten to narrow the
warrant of the no-change decision. The first version overstated the
equivalence between declared evaluation cost and verifier readiness,
overstated the guarantee of the falsifiability floor, and
mislabeled a design disposition as a deterministic consequence.
Reviewer correction verbatim is preserved below in "Reviewer
correction 2026-09-12 (post-`12cbaa3`)".

## What (o) proposed to solve

The (g-scoped) investigation surfaced a design-space observation:
"the search policy could downweight proposals whose falsification
paths are non-concrete." User authorization on (g-scoped) explicitly
sanctioned bounded observations informing search priority.

(o) as first framed would have added a new policy directive — some
form of "penalize non-concrete falsification path." This loop's task
was to predeclare (o) rigorously before authorization.

## CMA on the existing policy engine (establishes the mechanism, not its adequacy)

`internal/policy/engine.go` and `internal/policy/apply.go` show that
the pipeline persists an ordinal ranking that consumes a candidate's
**declared** `EvaluationCost` and floor-protects one violator per
target.

**Directive vocabulary (verbatim from `engine.go:56-88`)**: every
policy directive targets a PERSISTED ROW by ID or predicate
fingerprint. Every target kind resolves to a real row (KTD-1). Free-
form falsification-path text is not a persisted-row target; a
directive keyed on it would fail the resolution contract.

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

The comparator is a base-objective ordering: mechanistic distance
desc, then info gain desc, then declared evaluation cost asc, then
hash. Within a violation tier and applied-bias band, the cheaper
declared cost sorts earlier.

**Falsifiability floor (`apply.go:160-190`)**: the floor selects, per
target invariant, the cheapest declared-cost violating candidate and
prevents its net applied-policy bias from becoming negative
(`ab.Net = 0` if it would otherwise be negative). What the floor
does NOT do:

- It does not establish that any verifier CAN decide the protected
  candidate's claim.
- It does not reserve execution budget for the protected candidate.
- It does not guarantee the protected candidate will be evaluated.
- It provides no witness attribution.

The floor is an anti-suppression protection against active-directive
policy penalties; it is not a "check exists" guarantee.

## Property distinction (correcting the earlier record)

**Declared evaluation cost and verifier readiness are different
properties.** The first record conflated them; this rewrite separates
them.

- A concretely-checkable claim can honestly declare `evaluation_cost:
  medium` or `high` when the concrete check is expensive.
- An abstract mechanism-family claim can declare `evaluation_cost:
  low` if the operator reads "cost" as reflecting the cost of
  observing an existing pipeline output rather than the cost of
  producing a domain outcome.

My H2-escape proposal (`fpr_01M2BCSS7EJ5WBYNST2J3F0ECY`) is the
counterexample from within this pilot's own records: it self-attests
`evaluation_cost: low` because its "falsification" is a pipeline-
verdict read, and it ranks at 0 in its violation tier. Yet (g-scoped)
recorded that this proposal has no candidate-specific witness
obligation. **Cost ranking is operating; verifier readiness is not
established.** The two properties do not coincide.

## SGO on M7 rank ordinals (consistency, not isolated cost demonstration)

The persisted `rank_ordinal` values in M7's `frontier_proposals`
table are consistent with `baseObjectiveLess` on the persisted
ordinal tuples. The table does not, on its own, isolate the cost
term — earlier ordering dimensions and applied-bias net contributions
are not held equal across the rank-0 rows. Consistency with the
comparator is what the SGO establishes; a controlled cost
discrimination would require holding mechanistic distance, info
gain, and applied bias equal. No new experiment is required; the
distinction matters only in the labeling.

## Reviewer correction 2026-09-12 (post-`12cbaa3`)

Verbatim reviewer commentary on the first version of this record:

> Closing (o) as "no additional ranking directive justified" is
> defensible. Closing it as "concrete falsifiability is already
> established by the ranker" is not. I checked `12cbaa3`; the
> implementation supports the narrower conclusion, while the new
> record overstates the equivalence between evaluation cost and
> checkability.
>
> - Cost and concreteness are different properties. A concrete check
>   can be expensive; an abstract proposal can declare
>   `evaluation_cost: low`. Your own record supplies the
>   counterexample: the H2-escape proposal declares low cost and
>   ranks first, while the investigation reports that the relevant
>   population lacks candidate-specific witness obligations. That
>   demonstrates cost-based prioritization, not verifier readiness.
>   The authoring convention in (q) can improve this signal, but it
>   remains an operator judgment requiring calibration.
> - The ranking protection is narrower than "the falsification
>   surface is code-guaranteed." With active directives, Apply
>   orders by violation tier, then net policy bias, then
>   baseObjectiveLess. The floor selects the cheapest declared-cost
>   violator per target and prevents its net bias becoming negative.
>   It does not establish that a check exists, reserve execution
>   budget, or guarantee that the protected candidate will be
>   evaluated.
> - The recorded rank table supports consistency, not an isolated
>   demonstration of the cost term. It omits generation/occurrence
>   identifiers and applied policy bias, and contains several
>   rank-zero entries. To demonstrate cost discrimination
>   specifically, compare candidates within the same ranking context
>   with earlier ordering dimensions held equal.
> - "No code change justified" is a design disposition, not a
>   deterministic consequence. Code inspection establishes the
>   existing mechanism. Whether that mechanism adequately represents
>   the desired preference depends on the validity of its inputs
>   and the decision being supported. The record's classification of
>   the no-change verdict as a deterministic consequence should
>   therefore be narrowed.
> - Campaign closure and review closure must stay separate. This
>   commit changes only the scorecard and the new research record.
>   It does not discharge the previously identified C7, witness-
>   occurrence attribution, or C8 obligations.
>
> Keep the no-change decision. Narrow its warrant.

## Closure wording (reviewer-supplied, adopted)

> **Path (o) closed without implementation.** Existing ordinal
> evaluation-cost ranking and per-target anti-suppression protection
> make an additional cost-priority directive unjustified for this
> iteration. These mechanisms consume declared cost; they do not
> establish concrete verifier readiness or witness attribution. The
> next research input must supply those properties explicitly. H3
> remains unchanged, and review-integration obligations retain their
> separate status.

## Where the signal that (o) targeted actually lives

The observation from (g-scoped) — "none of the persisted target-
violating proposals has a candidate-specific witness obligation" — is
a corpus-level signal about the proposal population, not a lack in
the ranker. A ranker cannot manufacture verifier readiness from
declared cost alone. Concrete verifier readiness requires each of:

1. a candidate-specific claim whose form matches an existing verifier
   contract (e.g. a specific `(n, x, y, z)` tuple for the witness
   path);
2. attributed provenance connecting the claim to the proposal's
   mechanism;
3. execution against the verifier — code-persisted, code-attributed.

None of these is a ranking concern. The next research input to close
this gap must supply properties (1)–(3) explicitly on a proposal, not
change the ranker.

## Follow-on observations (recorded, not proposed as change)

- **(q) Authoring discipline** — operators authoring wire proposals
  should self-attest `evaluation_cost: low` only when the
  falsification path IS concretely checkable, and use `medium/high`
  when domain math is required. This is an operator-judgment signal
  that requires calibration; it is not a code-owned discriminator.
- **(p) Generator bias** — a longer-horizon research direction is to
  bias the frontier generator toward mechanisms that produce concrete
  witnesses (e.g., seeded from es-01 Mordell polynomial identities,
  es-03 factorization scheme). This would supply property (1) above
  by construction. Recorded for downstream authority; not proposed
  here.

## Separately open (not discharged by this loop)

- **C7 obligation** (as previously identified by review).
- **Witness-occurrence attribution obligation.**
- **C8 obligation.**

Pausing pilot-004's exploratory code-change work is not evidence of
end-to-end conformance on these obligations. They retain their
separate status.

## H3 gate status: PRESERVED (unchanged)

- No admission rules changed.
- No new verifier tier added.
- No policy directive kind added.
- No wire schema field added.
- The ranker's evaluation_cost tiebreaker continues to operate as
  the pipeline authors designed it.

## Verification tier (corrected)

- Existence of the policy directive vocabulary and the ranking
  comparator: **CMA** (verbatim reading of `engine.go`, `apply.go`).
- Existence of the anti-suppression floor and its exact effect: **CMA**
  (verbatim reading of `apply.go:160-190`).
- M7 rank ordinals consistent with the comparator: **SGO on
  consistency**, not an isolated demonstration of the cost term.
- Property distinction between declared cost and verifier readiness:
  **CMA + SGO counterexample** (the H2-escape proposal declares
  `low` and ranks first, yet has no witness obligation).
- "No code change justified for this iteration": **design
  disposition**, not a deterministic consequence. Downgraded from
  the earlier record's CMA label.
- Adequacy of the existing signal for future decisions: **not
  established** by this loop.
- Discharge of C7, witness-occurrence attribution, C8: **not claimed
  and not established** by this loop.
