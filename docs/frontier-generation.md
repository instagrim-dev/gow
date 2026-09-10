# Frontier generation (M5.1)

Frontier generation takes a problem's **surviving** candidate invariants — the
conserved failure structure that has withstood challenge — and generates typed
**break-proposals**: candidate mechanisms that deliberately violate that
structure, are mechanistically distant from every known failure family, and are
cheap to kill when wrong.

This is the `break` / frontier operator in the research loop:

```text
failure-space -> failure invariant -> invariant break -> partial success -> ...
```

The repository owns everything truth-sensitive. A model authors a candidate
mechanism plus prose; **code decides** the nearest family, the mechanistic
distance, whether the claimed violation actually holds, and the ranking.

## What may be targeted

Only invariants whose **current lifecycle state is `surviving`** are legal
targets. `proposed` (unchallenged), `weaken`, `falsified`, and `established`
invariants are excluded. Surviving state is read from the
`invariant_current_state` view (M4.3); there is no state column to trust — state
lives only in the transition ledger.

If a problem has no surviving invariant, generation is a legitimate empty
outcome (the run still completes and persists an empty, provenance-bearing
generation), never a fabricated proposal.

## Division of responsibility

| Concern | Owner |
|---|---|
| candidate mechanism signature + directed-generation prose | **provider** (Generator role, `generate`) |
| nearest failure family + `mechanistic_distance` ordinal | **code** (`internal/frontier`, via `canon.CompareWithProfile`) |
| structural-violation verification | **code** (`invariant.Evaluate(predicate, proposed_signature)`) |
| proposal dedup identity (`proposal_hash`) | **code** |
| ranking by the ordinal objective | **code** |
| durable state, revisioning, provenance | **`newf`** (`internal/store`, migration `v14`) |

`ModelJudgment != Verification`: a proposal that *claims* a break but whose
signature still `satisfies` the invariant is recorded honestly as **not
violated**. The F3 completeness discipline keeps an ambiguous read `unknown`
(never a coerced `violates`).

## Each proposal records

Per the epic (`EPIC.md` M5.1):

```text
target invariant(s)          -- frontier_target_invariants (+ code-verified verdict)
claimed structural violation -- structural_violation_claim
nearest known families       -- frontier_nearest_clusters (+ code classification)
novelty argument             -- novelty_argument
cheapest falsification path  -- cheapest_falsification_path
expected information gain     -- expected_information_gain_ordinal (provider ordinal)
evaluation cost              -- evaluation_cost_ordinal (provider ordinal)
```

## Mechanistic distance

For each proposal, code compares the proposed signature to every cluster
representative under the pinned `canon.ProfileMechanismV1()` and takes the
**closest** classification. The overall `mechanistic_distance_ordinal` is the
inverse of the closest proximity:

- `mechanism-near` (incl. `surface-distinct+mechanism-near`) → **low** distance;
- `mechanism-distinct` / `surface-near+mechanism-distinct` → **high** distance;
- `unknown` → **medium**.

An empty family set yields **high** distance (nothing to be near). Ordinals are
`low`/`medium`/`high` bands (`domain.Ordinal`) — never a fabricated scalar.

## Violation verification

For each targeted surviving invariant, code evaluates the invariant's
`invariant-predicate/v1` predicate against the proposed signature:

- `violates` → the proposal **breaks** the invariant (a confirmed structural
  violation, `violated = true`);
- `satisfies` → the proposal **still preserves** the invariant (claim refuted);
- `unknown` → the proposed signature is ambiguous on the read axis (an epistemic
  gap, not a break).

A proposal violating ≥1 target on a code-decidable axis is flagged
`violates_any_target`.

## Ranking objective

Proposals are ranked lexicographically over ordinal components (no fabricated
weighted scalar), realizing the epic objective:

```text
maximize( mechanistic_distance + invariant_violation + expected_information_gain
        - evaluation_cost - redundancy )
```

Order: (1) code-confirmed violation first; (2) mechanistic distance; (3)
expected information gain; (4) lower evaluation cost; (5) redundancy penalty
(proposals sharing the same target set + nearest family set sort after the first
such proposal); (6) `proposal_hash` as a deterministic tiebreak.

## Deduplication

Each proposal's stable identity is a `proposal_hash` over its targets (sorted,
deduped), the proposed mechanism's canonical fingerprint, and its
trimmed/lowercased violation claim. Cosmetic-only variants collapse to one, and
`frontier_proposals` enforces `UNIQUE(problem_id, proposal_hash)` so a re-derived
identical directed attack never double-writes across runs.

## Persistence (schema `v14`)

Migration `v14` adds `frontier_generation_runs`, `frontier_proposals`,
`frontier_target_invariants`, and `frontier_nearest_clusters` (all immutable by
trigger; `frontier_proposals` is append-only except a **one-time `result` set**
for M5.2). It widens `provider_invocations.role` in place to admit `'generate'`
(FK-safe `writable_schema` edit, guarded and idempotent — the same pattern v11
used for `invariant` and v13 for `challenge`). Each generation gets a per-problem
monotonic `revision`. See [`persistence.md`](persistence.md).

Every proposal's `result` column is deliberately left **NULL**; populating it
(routing proposals to the strongest available verifier, recording verification
strength, and re-entering failed proposals into the failure atlas) is **M5.2**.

## Run lifecycle

`frontier generate` runs under the standard run lifecycle: a `frontier generate`
run is created `running`, the generator is invoked, code scores + verifies +
ranks, the generation is persisted, and the run is finalized `completed` (or
`failed` with the cause on any error). Deterministic and offline via the default
`DerivingFixtureGenerator`.

## CLI

```bash
# Generate proposals against the problem's surviving invariants.
newf frontier generate --problem <problem-id> [--count 8] [--json]

# Inspect.
newf frontier list --problem <problem-id> [--json]
newf frontier show [<generation-id>] [--problem <problem-id>] [--json]
```

The default generator (`DerivingFixtureGenerator`) is deterministic: for a
surviving invariant of the form `preserves contains X`, it proposes a mechanism
that introduces a global-coupling object (and reasons globally) instead of
preserving `X` — a mechanism designed *not* to preserve the invariant. Code
still re-scores and violation-verifies everything it proposes.
