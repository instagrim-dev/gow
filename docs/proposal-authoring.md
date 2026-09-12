# Proposal authoring discipline

**Scope**: operator-facing conventions for authoring
`proposal-wire/v1` payloads that pass through
`newf frontier generate --proposals-file`. This document
concerns operator judgment; the code does not, and cannot,
enforce these conventions.

**Status**: living operator convention, revision 1
(2026-09-12). Extracted from
[`records/o-closed-lever-exists.md`](../corpus/experiments/pilot-004-discovery/records/o-closed-lever-exists.md)'s
follow-on observation and framed by the reviewer's correction
recorded in that same file. See
[`docs/frontier-generation.md`](frontier-generation.md) for
the wire schema itself.

## Why this document exists

Pilot-004 recorded a property distinction that the wire schema
cannot enforce: **declared `evaluation_cost` and actual verifier
readiness are different properties.** A cost attestation is a
self-report about the falsification path; a verifier readiness is
a claim about whether the pipeline (or an external check) can
decide the proposal's structural claim. The two can coincide, and
often should — but they do not automatically.

Pilot-004's own H2-escape proposal
(`fpr_01M2BCSS7EJ5WBYNST2J3F0ECY`) is the counterexample from
within the corpus: it declared `evaluation_cost: low`, and its
declared falsification path IS concretely checkable
(read the target verdict from `frontier_target_invariants`). That
attestation is defensible. But the (g-scoped) investigation
recorded that this proposal has no **candidate-specific witness
obligation** — no concrete `(n, x, y, z)` tuple bound to a
verifier contract. Cost-based prioritization operated; verifier
readiness for the underlying domain claim was not established.

The reviewer's frame:

> The authoring convention in (q) can improve this signal, but it
> remains an operator judgment requiring calibration.

This document is that convention. It does not turn declared cost
into verifier readiness. It aligns the operator's self-report with
the property the ranker will consume.

## What the ranker does with `evaluation_cost`

- Ordinal, `low` / `medium` / `high` (`domain.Ordinal`), never a
  fabricated scalar. Missing values are permitted; strict-decode
  rejects invalid values.
- Used as **the fourth ordering dimension** in
  `internal/policy/apply.go:baseObjectiveLess`, after violation
  tier, applied policy bias, mechanistic distance, and expected
  information gain. Within the same first three tiers, cheaper
  declared cost sorts earlier.
- The per-target falsifiability floor in `apply.go` selects the
  cheapest declared-cost **violating** candidate per target
  invariant and prevents its net applied-policy bias from
  becoming negative.

## What the ranker does NOT do with `evaluation_cost`

- Does **not** establish that any verifier can decide the
  protected candidate's claim.
- Does **not** reserve execution budget for the protected
  candidate.
- Does **not** guarantee the protected candidate will be
  evaluated.
- Does **not** attach a witness to the candidate.
- Does **not** discriminate concretely-checkable claims from
  claims whose falsification path exists only in prose.

The floor is an anti-suppression protection against
active-directive policy penalties; it is not a "check exists"
guarantee.

## Decision rule for self-attesting `evaluation_cost`

Read the value the author is about to write out loud, in the form
"the cheapest falsification path is `<cost>`-cost." Then decide
by the **falsification path**, not by the proposal's declared
information gain or novelty:

- **`low`**: The `cheapest_falsification_path` is a deterministic
  pipeline read against an existing verifier or an already-runnable
  external check. Concretely: the operator can name the exact
  command, the exact table/row the outcome lands in, and the exact
  decision rule (`v = violates` → falsified; `v = unknown` →
  not falsified). No domain reasoning is needed beyond executing
  the check. The H2-escape wire above is a legitimate `low`: one
  command, one table read, one decision rule.

- **`medium`**: The falsification path requires **bounded** domain
  math or small new code, on the order of hours of operator time
  or a single controlled calculation. Concretely: the operator can
  name the exact quantities to compute (e.g., "verify the form
  `Q(u,v) = uv + λn(u+v)` has discriminant 1 and one class"), and
  the check terminates on a specific arithmetic outcome. The
  N2a joint-crossing proposal's step-2 discriminant check
  is a legitimate `medium`.

- **`high`**: The falsification path requires **unbounded** domain
  reasoning, empirical search over a large parameter space, novel
  verifier implementation, or expert domain judgment. Concretely:
  the operator cannot describe the check as a bounded procedure
  terminating on a specific outcome without committing to
  substantial new work.

If the operator cannot state the falsification path in one of
these three registers, the proposal is not ready to submit. The
missing property is a candidate-specific falsification path, not
a value for `evaluation_cost`.

## Anti-pattern: declaring `low` from optimism

Do not attest `low` because:

- the proposal *feels* cheap to evaluate ("obvious counterexample
  exists somewhere");
- the model output declared `low` without concrete grounding;
- the falsification path is described in prose but no specific
  witness contract exists;
- the operator prefers the proposal to rank ahead of others in
  the same target;
- adjacent proposals declared `low` and this one should be
  "consistent."

Any of these is grounds for `medium` at minimum. The reviewer's
observation from (o-closed) — "an abstract proposal can declare
`evaluation_cost: low`" — is the failure mode this rule targets.

## Anti-pattern: declaring `high` from indecision

Do not attest `high` merely because:

- the operator is uncertain how expensive the check will be;
- the proposal's target invariant is broadly stated;
- the check has multiple valid variants and the operator has not
  chosen one.

`high` is a substantive claim that the falsification path IS
unbounded or novel-verifier-required. Indecision is a signal that
the falsification path itself is not yet concrete. Address the
falsification path first, then attest the cost.

## Composition with `cheapest_falsification_path`

The wire schema already requires
`cheapest_falsification_path` as a prose field. The two fields
should be **consistent with each other by construction**:

- If the prose describes a one-command pipeline read →
  `evaluation_cost: low`.
- If the prose describes a specific bounded calculation →
  `evaluation_cost: medium`.
- If the prose describes an open-ended domain investigation →
  `evaluation_cost: high`.

A `low` cost with a prose path that opens "we conjecture that..."
is an authoring error the pipeline cannot catch. The reviewer
signal that would surface this is *not* the ranker but the
follow-up review protocol on the proposal's outcome.

## Calibration signal

There is no automated calibration procedure. Calibration happens
by review:

- When a `low`-declared proposal fails to reach a verdict, the
  disposition may be a witness gap, not a candidate defect.
  Record the gap in the proposal's follow-on record.
- When a `medium`-declared proposal turns out to require
  substantially more work than the operator predicted, revise the
  authoring convention entry (this file) with a concrete example.
- When a `high`-declared proposal turns out to have a cheap
  falsification path the author missed, record the discovered
  path in the proposal's follow-on record; the wire attestation
  is immutable but the corpus record is not.

The signal from the ranker is honest exactly to the extent that
declared costs are honest.

## Relation to `expected_information_gain`

`expected_information_gain` and `evaluation_cost` are the two
optional ordinal fields on `proposal-wire/v1`. They compose in the
ranking objective but express different properties:

- `expected_information_gain` — how informative a *positive*
  outcome would be if the check succeeds.
- `evaluation_cost` — how expensive the check is to *run*.

A `high` × `high` proposal (informative but expensive) can be
legitimate; the ranker will place it behind cheaper high-info
peers within the same target tier. A `low` × `low` proposal is
the ideal, provided both attestations are honest.

## What this document does NOT do

- Does NOT change the wire schema. `evaluation_cost` remains
  optional and ordinal; strict decoding still rejects invalid
  values without consulting this document.
- Does NOT change the ranker. `baseObjectiveLess` and the
  falsifiability floor in `internal/policy/apply.go` continue to
  operate as designed.
- Does NOT introduce any new verifier tier, admission rule, or
  policy directive.
- Does NOT promote declared cost into a verifier-readiness claim.
- Does NOT close the C7, C8, or witness-occurrence attribution
  obligations recorded elsewhere in pilot-004. Those retain their
  separate status.

## Reference material

- [`docs/frontier-generation.md`](frontier-generation.md) — wire
  schema and ranker mechanism.
- [`records/o-closed-lever-exists.md`](../corpus/experiments/pilot-004-discovery/records/o-closed-lever-exists.md) —
  the (o-closed) property-distinction analysis and the reviewer
  correction that framed this document.
- [`records/l-coh-arc-consolidation.md`](../corpus/experiments/pilot-004-discovery/records/l-coh-arc-consolidation.md) —
  the reviewer-correction arc summary that motivated the frame.
- `internal/policy/apply.go` — the ranker implementation this
  document references.
- `internal/provider/untrusted_proposer.go` — the strict-decode
  path that admits `proposal-wire/v1` values.
