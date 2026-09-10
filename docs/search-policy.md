# Search policy (M6.2)

`newf policy mutate` makes future search behavior **explicit, versioned,
persisted, and inspectable** — the last stage of the governing loop before the
historical-holdout experiment:

```text
success invariant / surviving failure invariant / coverage gap / repeated failure
  -> code-owned policy evidence
  -> typed directives (prefer | avoid | expand | penalize)
  -> bounded ordinal bias over the next frontier generation
  -> reproducible applied-bias log
```

Results must change future search behavior, and the change must be inspectable
rather than living "in the model's context". A policy revision is a durable,
immutable artifact; a biased generation records exactly why each proposal was
favored or suppressed.

## Code owns policy identity (KTD-1)

`newf` builds the policy evidence from persisted rows only:

- **successes** — the latest success-compression revision's invariants (M6.1),
  each carrying its verification-strength composition;
- **surviving** — `surviving` / `operator_attested` candidate invariants (the
  conserved failure structure to steer away from);
- **uncovered families** — under-sampled coverage axes from the latest cluster
  run;
- **repeated failures** — re-entered evaluated failures (`failure` /
  `partial_failure`) from the M5.2 failure atlas.

The provider (`role='policy-mutate'`) MAY propose directives, but every proposal
is **re-verified against the resolvable evidence set** before it counts: a
directive whose target does not resolve to a persisted row is recorded
`inert_proposals` and biases nothing. `ModelJudgment != Verification`. The
code-owned derivation is the source of truth; a provider fold-in can only add
verified directives, never invent targets. `--no-provider` runs the pure code
derivation.

## Typed directives with strength-weighted preference

Directives are typed over typed targets:

| kind | target kind | meaning |
|---|---|---|
| `prefer` | `success_invariant` | favor proposals whose signature satisfies a supported success condition |
| `avoid` | `surviving_invariant` | steer away from re-preserving conserved failure structure |
| `expand` | `mechanism_family` | sample under-covered mechanism families |
| `penalize` | `redundant_attack` / `repeated_failure` | dampen mechanisms shown redundant or repeatedly failing |

Each directive carries an ordinal `weight`, an `epistemic_source`, and
provenance back to the justifying evidence rows. Preference weight for a success
invariant is derived from the **strongest verification class present in its
support composition** — never promoted beyond what the counts show.

## Bounded ordinal bias with an inviolable violation gate (KTD-2)

`policy.Apply` re-ranks the already-`Rank`-ordered candidates. It is a
**bounded ordinal transform**, not an override:

- the **code-verified violation gate is inviolable** — a non-violating proposal
  can never outrank a violating one, no matter how strongly preferred;
- **no proposal is ever dropped** — suppression is a rank penalty, not a
  removal;
- for each targeted invariant the **cheapest-falsification** violating proposal
  is **floor-protected**, so policy can never render a run unfalsifiable (the
  falsifiability floor);
- preference matches are computed in **code** by evaluating each preferred
  success invariant's stored predicate against every candidate's proposed
  signature — the policy names an invariant; code decides which proposals
  satisfy it.

## Idempotent, revisioned persistence (migration `v19`)

A mutation pass is idempotent on the order-independent **evidence-cohort hash**:
re-mutating over unchanged evidence returns the existing revision; new evidence
yields the next `revision`, never a rewrite. The layer is additive and
immutable:

- `search_policy_revisions` — one pass (`evidence_cohort_hash`, `revision`,
  `directive_count`, `inert_proposals`, optional `policy-mutate` invocation);
- `search_policy_directives` — the typed bias (`kind` / `target_kind` CHECKed);
- `search_policy_provenance` — per-directive justifying evidence references;
- `frontier_generation_policy` — the applied-bias log.

The one non-additive step widens `provider_invocations.role` to admit
`'policy-mutate'` via the established guarded in-place CHECK edit. See
[`persistence.md`](persistence.md).

## The applied-bias log makes "why" reproducible

When a `frontier generate` run is biased by a policy, `newf` records — per
newly-persisted proposal — the net ordinal bias and whether it was `preferred`,
`avoided`, `penalized`, or `floor_protected`, keyed to the policy revision that
produced it. A generation with **no policy writes nothing**: unbiased ==
absence. `frontier generate --no-policy` is a first-class unbiased baseline for
the M7 directed-vs-undirected contrast.

This satisfies the M6.2 exit condition: *a new run can reproduce why a
particular frontier proposal was favored or suppressed.*

## CLI

```bash
# derive a policy from accumulated evidence (idempotent per unchanged cohort)
newf policy mutate --problem <id>
newf policy mutate --problem <id> --no-provider   # pure code derivation

# inspect
newf policy list --problem <id>
newf policy show --problem <id>                   # latest
newf policy show <policy-revision-id>

# apply (default) or run the unbiased baseline
newf frontier generate --problem <id>
newf frontier generate --problem <id> --no-policy
```

All commands are deterministic and offline, and emit a stable `--json`
envelope.
