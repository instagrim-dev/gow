# (l) Cross-domain replication predeclaration — P vs NP holdout corpus

**Status**: PREDECLARED, not yet executed. Recorded to prevent post-hoc
reasoning.

## Question

Does the (a‴-B) enum-axis miner + (i) cross-corpus replication result
generalize beyond ES to a genuinely different mathematical domain?

## Target corpus

`corpus/experiments/pvnp-holdout/{train,target}` at HEAD `324ec5d`
(12 train files + 2 target files; adversarially-decontaminated per
`008b0c2`; each carries `newf-normalize` payload with
`locality`/`construction_mode`/`uncertainty_mode` axes).

## Method

1. Fresh isolated DB at `.newf/pvnp/newf.db`.
2. `newf ingest` train (12 files) + target (2 files, separate problem).
3. `newf normalize --provider fixture` (deterministic).
4. `newf mechanism signature`, `newf cluster build`,
   `newf failure-space build`.
5. `newf invariants mine --emit-posture-axes` (the (a‴-B) flag).
6. `newf challenge --all`.

## Predeclared readings

### Reading A — full replication (strongest)

The miner emits enum-axis candidate predicates on
`locality`/`construction`/`uncertainty` that reach `recurring` state
with support ≥ 2 and clean failure-side specificity, analogous to ES's
`equals(locality, local)` (support 3, 0/5 success prevalence).

Semantic-content sub-reading: at least one enum-axis predicate has an
interpretable failure-mode meaning in P vs NP (e.g. failure entries
share `locality=global` because black-box/oracle-style arguments
dominate the pre-2005 barrier programs).

**If reading A obtains**: the enum-axis miner is domain-general and
the (i) headline result is a mechanism, not an ES artifact.

### Reading B — partial replication

The miner emits enum-axis candidates, some reach `recurring`, but the
axis distribution or interpretability differs materially from ES
(e.g. `uncertainty` axis becomes discriminating in P vs NP where it
was flat in ES).

**If reading B obtains**: the miner generalizes as a mechanism, but
which specific axis carries the failure signature is corpus-specific.
This is still a positive replication finding.

### Reading C — non-replication

The miner emits few or no enum-axis candidates that reach `recurring`
on P vs NP. This can happen if (a) the P vs NP corpus has flat
enum-axis distributions, (b) the corpus size is too small (12 train)
to hit the default support threshold, or (c) enum-axis mining is an
ES-specific artifact.

**If reading C obtains**: reading matters — distinguish "corpus size
insufficient" from "the mechanism is ES-specific." The scorecard
would record this as a corpus-specific rather than domain-general
result.

### Reading D — degenerate

The pipeline fails at some stage before invariants can be mined
(ingest error, normalize error, empty cluster, empty failure-space).

**If reading D obtains**: not evidence about the miner; report the
stage of failure and stop.

## What DOES NOT count as evidence

- Recurring predicates from Contains-preserves shapes only (those
  can be produced without the (a‴-B) flag; they don't test the enum-axis
  emission).
- Predicates that only match by chance because of small corpus size —
  support ≥ 2 is the pipeline's own threshold, respect it, but flag any
  candidate reaching `recurring` at exactly support = 2 as marginal.
- Any post-hoc adjustment of predicate selection based on what "looks
  interpretable" — the challenge campaign's disposition is the authority.

## Verification tier

- SGO on pipeline output (verbatim `newf --json` output).
- CMA on miner behavior (already established in (a‴-B) and (i)).
- Semantic-content sub-reading is **PE** (proposed explanation) at
  best, not CMA — atlas-truthfulness for P vs NP failure structure is
  outside my epistemic reach as an agent, and the corpus itself is
  the operator's decontaminated construction.

## H3 gate status: PRESERVED

- No admission rules changed.
- No policy directives added.
- No verifier tier changed.
- No wire schema field added.

This is a pipeline execution on a fresh corpus, not an H3 escape
attempt or a proposal-authoring step.

## What this WILL and WILL NOT establish

WILL establish (if reading A or B obtains): the enum-axis mining
mechanism has cross-domain traction, extending the ES-only claim of (i).

WILL NOT establish:
- Any claim about P vs NP itself.
- Any discharge of C7/C8/witness-occurrence attribution obligations
  (these are review-integration obligations in the concurrent writer's
  framework, orthogonal to cross-domain miner replication).
- Any claim that the pilot-005-relational protocol can be frozen; that
  requires operator authorization on six open items.
