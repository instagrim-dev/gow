# pvnp-holdout RESULT — P vs NP historical holdout (issue #13)

Executed 2026-09-12. Freeze commit `cb62bf2`; dating-gate commit
`61bd1d3`; captures/run commit `7582a7d`. Store:
`records/pvnp.db` (schema and all pipeline rows), full JSON provenance
under `records/`.

## Primary endpoint (code-owned, mechanical)

Experiment `exp_01M2BW54D2VZX4AQJSXT03BMY4`, run
`run_01M2BW54D2V7X406MG6FJNQDVB`, **mode `historical`** through the
lifted dating gate (both withheld sources dated with audited
earliest-public evidence: 2010-05-04 and 2010-11-08, strictly
post-cutoff). Budgets 8/8. Leakage check `leakage-check/v1` passed:
0 snapshot / 0 normalization / 0 signature leaks.

| arm | proposals | recovered | nearest | decisive | unknown | unassessed |
|---|---|---|---|---|---|---|
| b0_undirected | 6 | false | mechanism-distinct | 1 | 3 | 2 |
| b3_invariant_guided | 5 | false | mechanism-distinct | 3 | 1 | 1 |

**Conclusion: `inconclusive`** — exactly the preregistered expectation
(SCOPE, "Predeclared consequences" §2): under `classify/v3`,
completeness admission strips at the untrusted boundary for externally
captured proposals, so the mechanical recovery rule cannot certify
recovery. This is the pinned contract, not a defect. The historical
vocabulary gate held: the run records the historical disclaimer and
never emitted a blinded-mode vocabulary word.

## Secondary endpoint (ModelJudgment, predeclared — never promoted)

Two independent blinded adjudicators judged the 11 proposals (shuffled
seed 20260912, arm-masked, `records/adjudication-packet.json`) against
the frozen target manifest under the predeclared criterion with the
KI03 distinctness guard (`records/ADJUDICATOR-INSTRUCTIONS.md`).

**Concordance 11/11** (`records/adjudication-concordance.json`). Both
adjudicators independently found:

- **P10 = B3 rank 0: `recovers`.** Supplies a modest-savings ACC0-SAT
  algorithm, converts through nondeterministic-hierarchy /
  indirect-diagonalization machinery with succinct completeness
  translations into an unconditional NEXP-not-in-ACC0-type bound,
  coupling stated as the load-bearing algorithm-first reversal; clears
  both horns of the distinctness guard (not the conditional PIT
  bridge, not the refuted-assumption skeleton).
- **P11 = B0 rank 0: `recovers`.** Same structural move, independently
  articulated by the undirected arm ("the hypothetical algorithm is to
  be supplied rather than refuted, with the refuted assumption instead
  being the circuit upper bound").
- All nine other proposals: `does_not_recover`, with concordant
  partial notes (P02 = supplied-algorithm posture without hierarchy
  machinery, TRAIN-near; P08 = hierarchy machinery without a supplied
  algorithm, TRAIN-near).
- Adjudicator A recorded one blemish in P10: "a nondeterministic
  ACC0-SAT algorithm" slightly misstates the target's setting; judged
  non-load-bearing.

## Comparative finding (the honest headline)

**Both arms' top-priority proposal recovered the withheld structural
move; the invariant-guided arm showed no differential over undirected
on the recovery question.** B0, given only the train sources, landed
the same algorithm-first ACC0 pipeline at rank 0 as B3 given the
weakened-candidate geometry. Under the adjudicated secondary endpoint
this experiment therefore does NOT evidence guidance value for the
zero-survivor weakened geometry; it evidences that the train atlas
itself (which deliberately contains KI03's conditional bridge, the
pnp-12 refutation skeleton, and the pnp-06 ACC0 stall edge, per
honesty constraint 1) makes the compositional move findable for this
proposer. Where the arms did differ: B3's recovering proposal named
the conserved postures it was departing (the weakened-candidate
vocabulary), and B3 produced fewer, more departure-concentrated
proposals (5 vs 6, with 3 decisive-no mechanical assessments vs 1).
No claim is made on n=1 captures per arm.

## Limitations (recorded, binding on any citation of this result)

1. **Model-judgment ceiling.** The recovery verdicts are ModelJudgment
   under the verification hierarchy — two-adjudicator agreement is
   still model judgment. Nothing here is independently verified
   mathematics; no proposal constructs the ACC0-SAT algorithm.
2. **Blinding imperfect.** Both adjudicators flagged inferable cues:
   B3-flavored vocabulary ("weakened candidates", "boundary deltas")
   and near-duplicate pairs (P10/P11, P03/P09) suggesting distinct
   conditions. Both certified judging on substance only; the
   limitation is reported, not waved off.
3. **Shared model family.** Proposers and adjudicators are instances
   of the same model family as the operator's session. Cross-instance
   agreement does not constitute independent-critic agreement in the
   strong sense.
4. **Training-data contamination is not excluded by the protocol.**
   The proposer model's weights postdate Williams 2010/2011. The
   historical gate certifies the *permitted context* excluded
   post-cutoff material; it cannot certify the model's parametric
   knowledge did. This is the structural weakness of every
   historical-holdout run with a modern model and is why the
   mechanical conclusion stays `inconclusive`. A blinded-mode run on a
   post-training-cutoff advance is the stronger successor experiment.
5. **n=1 capture per arm; single proposer configuration.**
6. **Corpus is project-authored study material**, adversarially
   audited for leakage/telegraphing (two rounds, plus a blind
   chronology audit), but not raw historical text.

## Verdict for issue #13

The M7 historical-holdout machinery is exercised end-to-end and
functioning: corpus -> vocabulary (mechanism/v6) -> mining ->
challenge -> zero-survivor geometry -> dating gate -> historical run ->
preregistered inconclusive + adjudicated secondary endpoint. The
system's honest-vocabulary discipline held at every stage. The
scientific finding on guidance value is null-at-n=1 with a
contamination caveat; the infrastructure finding is positive.
