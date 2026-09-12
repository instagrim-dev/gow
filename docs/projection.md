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
`evidence_ref`, and (v40) copies the backing evaluation's **verdict and
verification strength verbatim into typed columns**
(`evaluation_verdict`/`evaluation_strength`, surfaced in `--json`), so a
policy consumer can weigh a discharge without parsing the prose basis; a
model-judged observation stays labeled as such
(ModelJudgment != Verification).

Two coherence gates guard the decision (issue #23 closure):

- **Subject gate (v38):** the backing evaluation's `verification_subject`
  must be `domain-goal`. An annotation-subject certificate (a deterministic
  predicate check over the signature) is about the *description*, not the
  domain — accepting it here would be the wrong-object certification the
  subject axis exists to prevent. Pre-v38 evaluations with no recorded
  subject are refused for the same reason.
- **Verdict–status coherence:** `discharged` (realization holds) requires an
  unambiguous domain `success`; `failed` requires a failure-side verdict
  (`failure`/`partial_failure`). Partial, unknown, or blocked verdicts back
  neither — recording the stronger status over a weaker verdict is a silent
  epistemic promotion.

The strongest available backing is a **witness-backed evaluation**
(`newf witness check`, issue #23): an exact-integer domain check recorded at
`reproducible` strength, whose canonical claim the decision's basis and
typed columns carry.

> **Current limitation (recorded, 2026-09-12 review F10):** obligations and
> their decisions are persisted and inspectable, but no search-policy
> mutation consumes them yet. The typed columns exist so that consumption,
> when built, needs no prose parsing — the edge itself is not shipped.

## projection/v2: the semantic-preservation contract (issue #22, D3)

A projection is a re-representation, and `docs/abstraction-safety.md`
requires every non-trivial re-representation to state what it preserves,
what it loses, and how it grounds back. `projection/v2` makes that contract
part of the artifact (all six fields **required**; migration v42):

```json
{
  "schema": "projection/v2",
  "givens": ["..."], "steps": [ ... ], "target": ["..."],
  "source_domain": "unit-fraction identities over Z",
  "target_domain": "congruence covers of the moduli",
  "preserves": ["solution existence per residue class"],
  "loses": ["constructive witness values"],
  "correspondence": "one-way-implication",
  "grounding_plan": "each residue-class bound maps back to concrete n with a checkable witness obligation"
}
```

- `correspondence` ∈ {`equivalence`, `one-way-implication`, `analogy`,
  `unknown`} — closed vocabulary, **no default**: an unstated class is a
  parse error, not "unknown".
- `preserves` must be non-empty (an abstraction that cannot state what it
  preserves is defective); `loses` must be *present* — an explicit `[]` is
  the author's claim that nothing known is lost, a missing field is an
  unexamined loss surface and is refused.
- Strict parse both ways: a v2 artifact missing any field is refused, and a
  v1 artifact carrying v2 fields is refused (semantic claims may not ride a
  v1 declaration without the obligations they owe).
- **v1 artifacts keep parsing** and read as correspondence-unrecorded; there
  is no retroactive backfill.

A composing v2 artifact owes a **third obligation** of kind
`semantic-preservation` (checker `external`, ordinal 2), stating the claimed
contract verbatim. It stays open until discharged and passes exactly the
same gates as `domain-realization`: evaluation-backed only (manual discharge
without evidence is refused), subject `domain-goal`, verdict–status
coherence. Per the D2 seam (issue #23), a grounding-plan instrument is
naturally a witness-backed evaluation.

Composition checking is byte-for-byte v1-identical: v2 adds semantics, not a
new checker.

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
- `TestIntegrationWitnessBackedProjectionDischarge` — a domain-realization
  obligation discharged on an exact witness check, with both coherence gates
  proven: `discharged`-over-failure refused as an epistemic promotion, and an
  annotation-subject certificate refused as the wrong object.
- `internal/projection` unit tests pin the checker's ordering and gap
  semantics; `internal/store` tests pin revision allocation, append-once
  decisions, and SQL-level immutability.
