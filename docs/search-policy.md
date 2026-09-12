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
passes through the ONE engine-owned admission gate
(`policy.AdmitProposedDirective`) — the same evidence requirements `Derive`
applies to its own output. A resolvable reference is **not** sufficient
evidence for the requested action: the (kind, target-kind) pairing must be one
`Derive` itself emits, the target must resolve to a persisted evidence row, the
target-specific support gates hold (a success invariant with zero distinct
support earns no preference, from a provider exactly as from code), and the
admitted weight is the evidence-derived weight **capped at medium** — a
provider can confirm evidence-backed bias, never amplify it. Anything failing a
gate is recorded `inert_proposals` and biases nothing.
`ModelJudgment != Verification`. The code-owned derivation is the source of
truth; a provider fold-in can only add gate-passing directives, never invent
targets or bypass support requirements. `--no-provider` runs the pure code
derivation.

## Typed directives with strength-weighted preference

Directives are typed over typed targets:

| kind | target kind | meaning |
|---|---|---|
| `prefer` | `success_invariant` | favor proposals whose signature satisfies a supported success condition |
| `avoid` | `surviving_invariant` | steer away from re-preserving conserved failure structure |
| `expand` | `mechanism_family` | sample under-covered mechanism families *(derived + persisted; generation-path application deferred — see below)* |
| `expand` | `refuted_boundary` | expand where a confirmed challenge PROVED believed failure structure violable (v43, D5) *(same generation-path deferral as `mechanism_family`)* |
| `penalize` | `redundant_attack` | dampen directed attacks seen on ≥2 distinct proposals (down-rank in the applied rerank) |
| `penalize` | `repeated_failure` | dampen repeatedly-failing mechanisms *(generation-path lever; deferred — see below)* |

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

## Applied vs. deferred levers

The applied levers — those that change the *current* rerank — are `prefer`
(success invariant), `avoid` (surviving invariant), and `penalize`
(redundant_attack). These fire in `policy.Apply` over the ranked candidate set,
and their per-proposal effect is recorded in the applied-bias log.

Two lever kinds are **derived and persisted for provenance/inspection but not
yet applied**, because they are *generation-path* levers (they change which
families/mechanisms are drawn *before* ranking) and the generation-request path
does not yet consume the persisted policy:

- `expand` (`mechanism_family`) — would broaden sampling into under-covered
  families;
- `expand` (`refuted_boundary`) — would direct generation toward predicates a
  confirmed challenge proved violable;
- `penalize` (`repeated_failure`) — would dampen repeatedly-failing mechanisms.

They are carried in the revision so an operator can see the accumulated
intent, and `policy.Apply` intentionally does not fire them (a comment in
`internal/policy/apply.go` marks the deferral). Wiring the generation request to
consume policy is deferred to a follow-up so this slice stays a bounded,
testable rerank rather than a change to generation semantics. Until then,
`expand`/`repeated_failure` directives do not alter search behavior, and this is
stated rather than implied.

## Next-decision edges: which persisted feedback feeds `Derive` (decision D5, v43)

Three durable next-decision edges exist in the store. The 2026-09-12 decision
pass (D5) ordered their consumption; the first is now active:

1. **`challenge_boundary_deltas` — consumed (v43).** A confirmed challenge's
   typed boundary refinement. Only **separation-class** deltas
   (`counterexample-separation`, `constructibility`) earn a directive: they
   record a domain artifact that actually violated believed failure structure,
   which is precisely the frontier doctrine's cheapest expansion direction.
   Bookkeeping-class deltas (`contrast-collapse`, `support-recount`) expose
   epistemic defects of the *claim*, not domain structure; split/merge deltas
   flow through the child invariants' own lifecycle. Identity is the predicate
   fingerprint (repeated refutations dedupe); provenance rows name each
   justifying confirmed challenge; weight is `medium` (a demonstrated
   violation earns a direction, not a mandate). The doctrine lives in one
   place: `policy.ExpansionBearingDelta`.
2. **Projection obligation decisions — not consumed (rejected for first
   place).** Discharge decisions are proposal-scoped: keying directives on
   them would emit unbounded single-use rows — the same recorded objection
   that defers `repeated_failure` derivation. Trigger to revisit: a
   mechanism-level key for obligation outcomes.
3. **Episode outcomes (v41) — not consumed (rejected for first place).** The
   two-observation protocol is new; deriving policy from n≈1 episodes would
   bias search on anecdote, and the episode lane's own consumption protocol
   (revision credit) is still forming. Trigger to revisit: enough completed
   episodes for hit/miss to be population evidence.

Provider proposals may cite refuted boundaries (`expand`/`refuted_boundary` by
fingerprint); the admission gate re-verifies the delta class exactly as
`Derive` does, and caps the weight at `medium`.

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
