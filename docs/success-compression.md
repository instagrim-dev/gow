# Success compression (M6.1)

`newf successes compress` implements the symmetric question of the
failure-invariant loop:

> What common structure appears in proposals that cross a boundary the failure
> families could not cross?

```text
failure invariant: P remains preserved across failed families
success invariant: progress appears when P is broken under condition C
```

`C` may be more useful than the raw symmetry break — plenty of P-breakers still
fail. **`C` is what discriminates progress among P-breakers**, and everything
about it is code-owned except its authorship.

## The break cohort is code-selected (KTD-1)

For each failure invariant P with at least one **code-verified** violating
proposal (`frontier_target_invariants.violated = 1`, computed by
`invariant.Evaluate` at generation time — never claimed), the cohort is every
such proposal that has been evaluated (M5.2 one-time `result`), partitioned by
its verdict:

- **progressors** — `partial_success` / `success`;
- **non-progressors** — `failure` / `partial_failure`;
- `unknown` / `verification_blocked` members are excluded from both sides and
  **counted as ambiguous** (mirroring the M4.2 mixed-family discipline);
- proposals whose canonical content predates the v17 sidecar are **counted
  ineligible** (`ineligible_unpersisted`) — never silently included or
  fabricated; a deterministic `frontier generate` re-run enriches them.

The provider never nominates cohort members.

## Conditions are typed, admitted, and code-evaluated

The `SuccessCompressor` role (`'success-compress'`) receives compact structured
facts (per cohort: canonical content, results, strengths) and proposes
candidate conditions as `invariant-predicate/v1` ASTs — the same predicate
contract mining uses, through the same `AdmitCandidate` gate (grammar +
pinned-vocabulary references + **no outcome reads**: a `C` that reads the
outcome axis over an outcome-partitioned cohort is tautological). An
inadmissible condition is **skipped and counted** (`inadmissible_conditions`),
never silently dropped and never stored unevaluated.

Code then evaluates every admitted `C` against every cohort member's persisted
signature and computes:

```text
ProgressCoverage(C)        = progressors satisfying C / eligible progressors
NonProgressorExclusion(C)  = non-progressors violating C / eligible non-progressors
```

Both are stored as **exact counts**; ordinal bands (`low`/`medium`/`high`,
`unknown` when the denominator is empty) are derived from them, never in place
of them (KTD-2). Full coverage with zero exclusion is cohort-wide boilerplate
and is visibly reported as such, not inflated. `distinct_mechanism_support`
dedups supporting progressors by canonical mechanism fingerprint: one mechanism
re-proposed twice is one support unit.

## Verification strength is first-class (KTD-3)

Each supporting verdict carries its M5.2 `verification_strength`, and every
success invariant aggregates the composition
(deterministic / reproducible / independent-evidence / independent-critic /
single-model-judgment). Three model-judged partial successes are not the
evidence three deterministic ones are — the artifact says so; nothing is
promoted. M6.2 can weight policy by strength; this slice only reports honestly.

## Identity, persistence, provenance (migration v17)

- A success invariant's identity is the **semantic predicate fingerprint**
  (same canonicalization as M4.2); paraphrases collapse.
- The same condition confirmed against multiple broken targets is **one**
  invariant carrying every P link (`success_invariant_broken_targets`). The
  merge **never discards a cohort's evidence** (H2): the condition is evaluated
  against *every* contributing target's cohort, and the merged counts, support,
  and per-member verdicts are aggregated over the **union of all contributing
  cohorts' members, deduplicated by (role, proposal)**. A contradictory member
  present only in a later target's cohort is retained, and an overlapping member
  is counted once — never the first cohort's assessment attributed to both
  targets.
- Each cohort member's outcome carries the provenance of the **single evaluation**
  it came from (`evaluation_id`) — verdict, strength, and id are taken **together**
  from ONE evaluation record selected under `selection-policy/v1` — the STRONGEST
  verification class wins, ties break to the latest, so an accepted deterministic
  reassessment displaces an earlier model judgment or blocked result while a later
  model-judged "success" can never displace a deterministic "failure" (the
  earliest-result view stays in the append-only ledger for historical analysis),
  so a re-evaluation can never splice the first verdict onto a later evaluation's
  strength (H1).
- Revisions are immutable and idempotent on
  `(problem, cohort_hash, compressor_version, predicate_schema, min_support)`,
  where `cohort_hash` is a content hash over the sorted
  (target, proposal, **evaluation_id**, result, strength) tuples — a new
  evaluation changes the hash and yields the next revision, never a rewrite.
- Candidates enter and stay **`proposed`** (`CHECK`-enforced); the challenge
  lifecycle for success invariants is a follow-up slice.
- v17 also closes a substrate gap: `frontier_proposal_signatures` persists each
  proposal's canonical content (generation writes it; dedup **enriches by
  INSERT** — the immutable proposal row is never updated), guarded by a
  round-trip test that the stored JSON re-fingerprints identically. The one
  non-additive step widens `provider_invocations.role` to admit
  `'success-compress'` (the shipped in-place `writable_schema` pattern).
- An empty cohort or a fully-inadmissible proposal set yields a **legitimate
  empty revision** (run `completed`), never a fabricated invariant.

## CLI

```text
newf successes compress --problem <id> [--min-support <n>]
newf success-invariant list --problem <id>
newf success-invariant show [success-revision-id] [--problem <id>]
```

All support `--json`. Runs use the standard lifecycle
(`running → completed/failed`). Deterministic and offline: the default
`DerivingFixtureCompressor` proposes `contains(field, id)` for every canonical
id shared by all progressors and absent from ≥1 non-progressor — transparent,
and re-verified by code like any model proposal.

## M6.2 handoff

The `proposed` success-invariant set — with broken-target links, exact
coverage/exclusion counts, and strength composition — is the typed read surface
search-policy mutation consumes ("prefer structures associated with partial
success"). That is the M6.1 exit condition: partial success changes the search
model rather than remaining a result row.

## Boundaries

- `internal/success` — pure compression engine (no SQL/Cobra/provider).
- `internal/provider/success_compressor.go` — `SuccessCompressor` + fixture.
- `internal/store/success_store.go` + migration v17 — persistence + cohort rows.
- `internal/pipeline/successes.go` — cohort building, admission, run lifecycle.
- `cmd/newf/success.go` — thin CLI wiring.
