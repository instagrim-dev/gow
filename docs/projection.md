# The projection chain (`newf projection`, migration v37)

Structural review 2026-09-12, finding S5 ("high for the systems claim"): a
structural description plus directed-generation prose is not a concrete
mathematical construction or executable artifact, and a refinement described
in a transcript is not a typed object subsequent policy can consume. The
projection chain separates the four records the reviewer required and gives
each an owner:

| # | Record | Table | Owner |
|---|---|---|---|
| 1 | proposed structural change | `frontier_proposals` (existing) | frontier generation |
| 2 | concrete projection artifact | `projection_artifacts` | authored (operator or tool), never inferred from prose |
| 3 | verification obligation | `projection_obligations` + `projection_obligation_decisions` | code for `steps-compose`; external checker for `domain-realization` |
| 4 | domain observation | the `evaluations` row a discharge decision references | evaluation / S2 admission |

All rows are immutable; decisions are append-once (re-deciding is a loud
refusal); an artifact revision is append-only per proposal and identical
content is a refused duplicate.

## The artifact is a checkable plan, not prose

A `projection/v1` artifact declares `givens`, ordered `steps` (each with
`requires` / `provides` tokens and a statement), and `target` tokens:

```json
{
  "schema": "projection/v1",
  "givens": ["congruence-lattice"],
  "target": ["crossing-bound"],
  "steps": [
    {"name": "lift",  "statement": "lift to the covering classes",
     "requires": ["congruence-lattice"], "provides": ["cover"]},
    {"name": "bound", "statement": "bound crossings over the cover",
     "requires": ["cover"], "provides": ["crossing-bound"]}
  ]
}
```

`newf projection propose --proposal fpr_... --file plan.json` records the
artifact and decides composition **deterministically**: every step's
requirements must be met by givens or the provides of *strictly earlier*
steps (a later step cannot supply an earlier one), and every target token
must eventually be provided. Tokens are exact-match, artifact-scoped names
(v1); a canonical-vocabulary binding is future work and is not faked by
fuzzy matching.

## Composition failure refutes the abstract path before any domain work

A plan whose steps cannot compose is the reviewer's demanded failure case:
the `steps-compose` obligation is **failed by code** with the exact missing
tokens per step (typed `gaps`, not prose), and **no domain-realization
obligation is created** — there is no composed plan to realize. A fixed plan
is a new revision.

## What code cannot check is an open obligation, not an assumption

For a composing plan, a `domain-realization` obligation (checker `external`)
is created **open**. This is the honest, persisted form of "the domain
checker is missing": the default evaluator's abstention is a safeguard, and
the open obligation records exactly what remains unverified.

`newf projection discharge --obligation obl_... --status discharged|failed
--evaluation evl_... --note ...` records the operator's verdict backed by a
domain observation — an evaluation of the **same proposal** the artifact
projects (enforced). The decision stores the evaluation id as
`evidence_ref` and carries the observation's verification-strength label in
its basis; a model-judged observation stays labeled as such
(ModelJudgment != Verification).

## Inspection

`newf projection list --problem prb_...` (or `--json`) shows every artifact
revision with its obligations and decisions — the full chain from proposed
change to observation is auditable without reading any transcript.

## Regression tests

- `TestIntegrationProjectionChainFourRecords` — one path through all four
  records, ending in an honest failed domain observation; code-owned
  obligations refuse manual discharge; decisions are append-once.
- `TestIntegrationProjectionCompositionFailure` — the non-composing plan is
  refuted deterministically with exact gaps, leaves no realization
  obligation, refuses identical re-submission, and admits a fixed plan as a
  new revision.
- `internal/projection` unit tests pin the checker's ordering and gap
  semantics; `internal/store` tests pin revision allocation, append-once
  decisions, and SQL-level immutability.
