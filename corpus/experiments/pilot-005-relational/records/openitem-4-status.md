# Pilot-005 open item #4 — status inspection: wire format and strict preflight substantively present in code; only an operator-facing spec doc remains for the checkbox

**Status**: inspection loop, 2026-09-12. Does not freeze pilot-005;
does not add new code, tests, or documentation. Records the
substantive completion state so the operator's freeze checklist can
proceed with an accurate picture.

## Item under review

Pilot-005 `PROTOCOL-DRAFT.md`, open items before freeze, item #4:

> Wire format + strict preflight for C3 proposal capture through the
> real decode path (`relational.ParseRelation`) before sealing.

## Current substantive state

**Wire format**: `relational-claim/v1`, defined at
`internal/relational/relation.go`:

- `RelationSchemaV1 = "relational-claim/v1"`.
- `Relation` struct: `{schema, root}`.
- `RelNode` struct with a tiny AST: `factor_eq`, `factor_neq`,
  `equals`, `all`, `any`, `not`.
- `FactorName`, `Value`, `Assignment`, `Factor` types live in
  `internal/relational/task.go`.

**Strict preflight**: `ParseRelation(raw, t *Task) (Relation, error)`
in `internal/relational/relation.go`:

1. `json.Decoder.DisallowUnknownFields()` — rejects any unknown JSON
   key at any nesting depth.
2. Schema-version check: rejects anything but exactly
   `relational-claim/v1`.
3. Per-op field discipline:
   - `factor_eq` / `factor_neq`: require `left`, `right`; forbid
     `factor`, `value`, `children`; reject self-relation
     (`left == right`).
   - `equals`: require `factor`, `value`; forbid `left`, `right`,
     `children`; require value ∈ factor domain.
   - `all` / `any`: require ≥1 child; forbid the atomic fields;
     recurse.
   - `not`: require exactly 1 child; forbid atomic fields; recurse.
4. Task-vocabulary check: every factor name must resolve to a
   task-declared `Factor`.
5. Factor-domain check: `equals`'s value must lie in the target
   factor's `Domain`.
6. Unknown-op rejection.

**Test coverage** (`internal/relational/relational_test.go:TestParseRelation_ExecutableDocument`):

- Accepts the canonical `success requires A ≠ B` document and
  round-trips it through `Check` to verify the parsed AST is
  executable.
- Rejects, each with `errors.Is(err, ErrInvalidRelation)`:
  wrong schema, unknown factor, self-relation, value outside
  domain, extraneous AST-node field, unknown op, unknown JSON
  field, empty boolean composition.

## What "before sealing" would still need

Reading item #4 literally, the wire-format grammar and its decoder
are code-complete and code-tested. What remains — and what the
"before sealing" clause seems to require — is a **brief
operator-facing spec document** that presents the grammar without
requiring the reader to consult Go source. A candidate structure
(not authored here, requiring operator authorization to freeze the
draft):

1. Purpose: what `relational-claim/v1` is and is not for.
2. JSON shape: top-level `schema` + `root`; the AST node types
   with their permitted fields.
3. Validation rules restated as prose (mirroring
   `validateRelNode`).
4. Example round-trip: an XOR relation JSON document and its
   parsed executable form.
5. Rejection examples: at least the eight test cases from
   `TestParseRelation_ExecutableDocument` with the exact
   `ErrInvalidRelation` reasons.
6. Non-goals: it is not a schema for corpus signatures; not a
   substitute for `invariant-predicate/v1`; not for representing
   epistemic uncertainty (relational tasks are total — see the
   `Holds` comment at line 150 of `relation.go`).

The theory doc (`docs/theory/07-relational-structure.md:110`) already
names the grammar in passing but does not present its shape as an
operator-consumable spec.

## Recommended disposition

**Item #4 is ~90% complete at code level.** The strict preflight
already runs on every `ParseRelation` call and has been exercised
against the exact spectrum of rejection cases the theory doc
mentions. Adding wider adversarial coverage (fuzzing? property
tests? extra-large payload limits?) is orthogonal design work,
not blocked by anything visible in the current draft.

**The remaining checkbox gap is documentation, not code.** Whether
this is enough to close item #4 at freeze depends on the operator's
sealing policy — is a `docs/relational-claim-v1.md` a required
artifact, or does the code-level spec + test coverage satisfy the
"before sealing" clause?

## What this record does NOT do

- Does not freeze pilot-005.
- Does not author the operator-facing spec document unauthorized.
- Does not add tests, fuzzing, or property coverage to
  `internal/relational`.
- Does not decide any of items #1, #2, #3, #5, #6 in the pilot-005
  open list — those are budget/hyperparameter/schedule/seed/artifact
  decisions requiring operator judgment.
- Does not preempt pilot-005's freeze workflow. This is an
  inspection outcome, offered to help the operator's freeze
  checklist proceed with an accurate picture of item #4's state.

## Consequences

- Pilot-005 PROTOCOL-DRAFT open items #1, #2, #3, #5, #6 remain
  strictly operator-authorized.
- Pilot-005 open item #4 can be re-scoped to either "author the
  operator-facing spec doc" (a small documentation task) or "close
  as substantively met, defer spec doc to post-freeze" (an
  operator policy call).
- H3 gate: preserved.
- No code, tests, or documentation added.
- No changes to pilot-004 records or scorecard beyond this
  inspection record's existence.

## Recommendation to the operator

At the next authorized session, decide between:

- **Option A**: author `docs/relational-claim-v1.md` (est. one page)
  and check item #4 as complete.
- **Option B**: check item #4 as substantively complete (code +
  tests + theory-doc reference), and note the spec doc as a
  low-priority post-freeze task.

Either closure allows pilot-005 to proceed to the remaining
five-item block, all of which still require operator authorization.
