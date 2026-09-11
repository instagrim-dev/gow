# Audit: "current state" readers (follow-on P5)

Scope: every non-test consumer of `frontier_proposals.result`,
`frontier_target_invariants`, and `MAX(revision)` / `ORDER BY revision DESC`
outside the queries fixed in the v26–v28 series. Method: classify each hit as
occurrence-aware / artifact-level / historical-by-design / defective. Audit
executed against `35db9f4`.

## Finding 1 — DEFECTIVE (fixed in this pass): policy revisions repeat the v28 pattern

`internal/store/policy_store.go`: `PersistPolicyRevision` dedup-reuses by
`(problem, evidence_cohort_hash, mutator_version, policy_schema)` — correct —
but `LatestPolicyRevision` selects `ORDER BY revision DESC`, and its consumers
are `internal/pipeline/policy.go:366` (`applySearchPolicy` — applied to every
frontier generation) and `:509` (policy show). Sequence: evidence E1 → P1;
evidence changes → P2; evidence reverts to E1's cohort hash → mutation reuses
P1, but generation keeps applying the higher-numbered stale P2. This is the
identical execution-vs-artifact confusion fixed for success compression in
v28. **Fix applied: v30 `policy_mutation_selections` + `LatestSelectedPolicyRevision`,
both consumers switched, guard-mutant regression added.**

## Finding 2 — flagged, not fixed: experiment "latest" under identity reversion

`internal/store/experiment_store.go:691` (`ORDER BY revision DESC`) serves
`experiment show/compare` "latest for the problem". `PersistExperiment`
dedup-reuses on identity; if the full assessment manifest reverts to an
earlier identity (reachable after A→B→A plus identical budgets/arms), the
reused older experiment is the current execution while show/compare read the
higher-numbered one. Lower severity: the consumers are reporting surfaces,
not a decision loop, and identity reversion requires the entire ordered
manifest to revert. Recommendation: apply the selection pattern if/when
experiments enter an automated loop; not fixed in this pass.

## Classified sound

- `experiment_store.go:571,606,659` (`ListProposalContents*` MAX-revision
  content joins): no non-test pipeline consumer exists today (interface-declared
  only); the M7 scoring path reads occurrence bindings instead. Hazard note:
  any future consumer inherits latest-revision semantics — prefer occurrence
  or selection-based readers.
- `experiment_store.go:750`: documented pre-v24 fallback, flagged
  `FallbackLatest` to the caller — honest by design.
- `frontier_store.go:452`: aggregates `frontier_target_invariants.invariant_id`
  only (attack-identity key for redundancy counting) — never reads `violated`
  currency.
- `frontier_store.go:350,373,418,549,574` and cluster-run latest readers:
  generation/cluster runs are execution records with monotonic revisions and
  no older-artifact reuse (cluster input sets are append-only, so an input-set
  hash cannot revert) — revision order IS execution order.
- `evaluation_store.go:386` (`evaluated_failures`): append-only event
  marker; the verdict is the writing evaluation's own assessment-context
  verdict — an event log, not a current-state view.
- MAX(revision) in `PersistSuccessRevision` / `PersistPolicyRevision` /
  `PersistExperiment` / `PersistFrontierGeneration`: revision-number
  allocators, not readers.
- `frontier_target_invariants` writes (`frontier_store.go:241`) and the
  per-proposal detail loader (`:310`): origin-time records retained by design;
  the only admission consumer is `evaluation_target_verdicts` (v26/v27).
