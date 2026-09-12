# Blinded structural-recovery adjudication — instructions (pvnp-holdout)

You are asked for a bounded assessment of a research CRITERION — not to
resolve P vs NP and not to certify any proposal's mathematics.

Input (below): 11 research-direction proposals for the P vs NP problem,
in shuffled order with masked identifiers, plus the frozen target
manifest and the predeclared criterion. You are deliberately NOT given:
which experimental condition produced any proposal, any prior judgments
or tallies, or any automatic classification. Do not seek them.

## Frozen target manifest

The withheld advance this experiment tests recovery of is the
**algorithm-to-lower-bound conversion** (the structural move of
Williams 2010/2011):

1. SUPPLY a concrete, constructed satisfiability algorithm for a
   restricted circuit class (ACC0) that beats exhaustive search by a
   modest (superpolynomial-savings) margin — the algorithm is actually
   delivered, not hypothesized;
2. CONVERT the algorithm's existence, through nondeterministic
   time-hierarchy machinery and succinct completeness translations,
   into an UNCONDITIONAL circuit lower bound (NEXP not in ACC0);
3. The coupling is load-bearing: faster exhaustive search for the class
   IS the hardness evidence; analysis direction runs algorithm->model,
   not model->function.

## Predeclared criterion (grading, honesty constraint 1)

For EACH proposal, record passage-level reasons (quote the passages) for:

1. Does it propose SUPPLYING (constructing) a nontrivial satisfiability
   or exhaustive-search-beating algorithm for a restricted circuit
   class, rather than consuming such an algorithm as an open hypothesis
   or refuting an assumed one?
2. Does it describe CONVERTING that algorithm's existence into an
   unconditional lower bound via nondeterministic time-hierarchy /
   indirect-diagonalization machinery?
3. Does it state the COUPLING as load-bearing (the algorithm is the
   hardness evidence), i.e. the analysis direction is algorithm-first?
4. Distinctness guard: a proposal that re-lands the TRAIN atlas's
   conditional identity-testing bridge (consuming an open
   derandomization hypothesis, "if PIT derandomizes then lower bounds")
   is mechanism-near TRAIN, NOT the target. Recovery credit requires
   the supplied-algorithm -> hierarchy-machinery -> unconditional-bound
   content. Likewise a pipeline that REFUTES its assumed algorithm
   (contradiction skeleton) is TRAIN-near, not the target.

Then a verdict per proposal, exactly one of: `recovers` /
`does_not_recover` / `cannot_assess`; partial annotations are permitted
as secondary notes. Base every verdict on what the proposal ACTUALLY
DESCRIBES — do not supply missing details on a proposal's behalf, and
record any place where a generous completion would be needed.

Procedural-cue treatment (required): if you suspect a proposal reveals
its experimental condition, note the cue and continue judging on
substance; report the limitation rather than certify blinding.

Deliverable: one JSON document:
{"schema":"pvnp-adjudication/v1","verdicts":[{"proposal_id":"P01",
"criterion_1":true|false,"criterion_2":true|false,"criterion_3":true|false,
"train_near_flag":true|false,"verdict":"recovers|does_not_recover|cannot_assess",
"quoted_evidence":["..."],"notes":"..."}],"procedural_cue_notes":"...",
"summary":"..."}
Your verdicts will be recorded verbatim as ModelJudgment — they are
never promoted to the mechanical conclusion vocabulary.
