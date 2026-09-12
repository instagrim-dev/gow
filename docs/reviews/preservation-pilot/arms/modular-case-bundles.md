---
id: preservation-pilot-modular-bundles
revision: 1
kind: frozen_comparison_arm_case_bundles
arm: modular
frozen: 2026-09-12
composes_with: ../../prompts/review-contract.md
---

# Modular arm — pinned case bundles (frozen)

Each bundle composes with the shared [review contract](../../prompts/review-contract.md)
and carries the **same substantive obligations, examples, symbols, and
evidence access** as the corresponding bespoke baseline prompt — packaging is
the only independent variable. A dispatched run receives: the contract, ONE
bundle below, checkout access at its pinned revision, and `AGENTS.md`. It is
denied `docs/reviews/`, post-revision commit messages, and remediation notes.
No bundle names a known finding or its location.

## Bundle E1 — assessment identity over evaluation results (S3 cases)

**Decision under review:** can a reader of the system's outputs distinguish
an *initial* verdict from the *current* verdict for a frontier proposal, and
does every surface that presents a result say which one it presents?

Obligations (same substance as the bespoke assessment-identity prompt):

- An assessment is identified by WHAT it assessed: content revision,
  generation occurrence, problem, and population context. A verdict without
  its context is not a current result.
- Re-evaluation appends; it must never be silently absorbed into an earlier
  verdict's presentation. Check write paths for first-write-wins or
  last-write-wins patterns (`INSERT OR IGNORE`, `UPDATE … WHERE … IS NULL`,
  unique keys) and ask of each: which read surfaces present this value, and
  do they qualify it?
- Symbols to inspect: `frontier_proposals.result`,
  `internal/store/evaluation_store.go` (persist path), any view/reader that
  reports a proposal outcome (CLI output structs included).
- Reproduce claims on an isolated SQLite store (`t.TempDir()`), both
  orderings: success-then-failure and failure-then-success.

Report per the contract's five sections; findings need actionable location,
triggering conditions, violated contract, downstream consequence, and a
discriminating regression check.

## Bundle E2 — population selection and interpretation revisions (S4 cases)

**Decision under review:** when the same underlying work item has multiple
interpretation revisions, which revisions are eligible for the default
population that supports invariants, and is that selection explicit?

Obligations (same substance as the bespoke assessment-identity prompt):

- A population is a conditioned selection, not "all rows that match a
  filter". Corrected interpretations of one approach must not count as
  independent support.
- Distinguish three access modes and check each is explicit and typed:
  current interpretation heads; pinned historical replay; all-history.
- Symbols to inspect: `ListSignaturesForProblem` and neighbors in
  `internal/store/canon_store.go`; normalization revisions and approach
  identity (`frontier-proposal:<id>` logical identity); cluster-run
  population inputs.
- Reproduce on an isolated store: sign one approach twice with a changed
  mechanism family; count default-population members; state which
  invariant-support figures would move.

## Bundle M — manuscript claim fidelity (M1 cases)

**Decision under review:** does every quantitative or epistemic claim in
`paper/geometry-of-work.tex`'s results and conclusion sections carry exactly
the strength its cited evidence supports?

Obligations (same substance as the frozen manuscript baseline):

- For each experiment section, list every numeric result WITH its producing
  conditions (design, budget, ordering, stopping rule, population); then
  check the prose for conditioned quantities presented as unconditional,
  demonstration-scale results presented as general, or verification tiers
  presented above their strength.
- Verify quoted numbers against the cited frozen artifacts under `corpus/`
  at the pinned revision; report mismatches, do not repair.
- `AGENTS.md` epistemic invariants govern; record the reviewed SHA; honest
  unresolved cases are reported separately from findings; generic warnings
  score nothing.
