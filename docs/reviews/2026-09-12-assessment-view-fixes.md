# Review fixes: current populations and assessment views

Base: `3443a1094407364be2e02264ffa1ee1bb0f3efde`.
Scope: deterministic read-path defects from the structural, semantic, and
epistemic review. This is a partial implementation of that review, not closure
of the research program or every integration gap.

## Implemented

New clustering reads the current interpretation head of each logical approach,
not every historical revision. Explicit supersession takes precedence over
observation timestamps. When several unsuperseded heads exist, created_at/id
selects one deterministically, consistent with the existing head-selection
ordering. Competing interpretations do not count as independent samples.
The head is selected before the schema/vocabulary filter: an unnormalized or
unsigned current interpretation cannot resurrect a superseded signature.
`ListHistoricalSignaturesForProblem` retains explicit all-history access;
existing cluster runs retain their pinned membership and replay behavior.

`frontier_proposals.result` remains an immutable initial-verdict compatibility
field. `GetFrontierGeneration` and `ListOccurrenceProposalRows` now derive their
returned `Result` from the evaluation ledger for the selected generation,
problem, cluster run, and exact occurrence content. A later success can supersede
a failure, and a later failure can supersede a success, without rewriting either.
Equal-clock evaluations use ledger insertion order. These read projections report
recorded assessments; they do not certify the underlying mathematical claim.
Unattributed or incompatible evaluations do not supply a current result.

Batch `evaluate` uses that occurrence-scoped assessment rather than a global
content-only cache hit. Re-emitting identical bytes under a new generation is
eligible for assessment; repeating an already-assessed occurrence stays a no-op.
Explicit by-ID evaluation remains available for reassessment.

Historical experiment records, outcome labels, vocabulary judgments, immutable
triggers, and the initial-verdict write behavior are unchanged. Existing raw SQL
consumers of `frontier_proposals.result` must treat it as the initial result, not
the current result. The old writer comment describing it as the latest verdict
is superseded by this contract and the read-path comments.

## Regression coverage

Store tests cover current versus historical membership, backdated supersession,
missing current signatures, problem/vocabulary isolation, both verdict reversal
orders, generation and content isolation, legacy empty-content history, and
unchanged initial-result storage. Evaluation read tests use the fully migrated
production schema. A pipeline integration test checks that identical re-emitted
content is assessed in its new generation and that a repeated batch is a no-op.

## Still open

1. Challenge assessment population: add an explicit assessment manifest distinct
   from the discovery population, with claim scope and fresh search eligibility.
   An expanded sample must not retroactively falsify a bounded historical claim.
2. Evidence admission and atlas materialization: `evaluated_failures` is a
   historical marker, not an admitted domain observation. Do not merge
   model judgments or failed description checks into observed mathematical work.
3. Typed boundary deltas and concrete projections: connect a concrete artifact,
   claim-specific verification obligation, checker coverage, and domain outcome.
   A signature plus prose is not a checked realization.
4. Persistent assessment identity: version target-state and verification-policy
   changes explicitly. Exact generation/content attribution is necessary but
   does not itself resolve every possible change to an assessment context.
5. Current claim summaries and prospective research validation: surface the N2a
   reassessment consistently and run the prospective comparison without changing
   historical endpoints or inflating the negative pilot into effectiveness.

These require separate tested vertical slices. None is claimed fixed merely by
adding a type, a comment, or another successful fixture run.
