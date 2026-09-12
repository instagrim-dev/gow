# (g-scoped) H3 escape investigation — witness path exists, no in-scope obligation matches

Recorded: 2026-09-12. Pipeline-space (SGO + code trace).

## User authorization scope (verbatim)

> Preserve the H3 abstention gate. Do not make CounterexampleSearch
> return success or partial_success from absence of a matching failure
> family, and do not admit verification_blocked as an observed domain
> outcome.
>
> Close only the proposed novelty-to-success shortcut. Identify one
> bounded, candidate-specific verification obligation and the cheapest
> check capable of deciding it. Inspect the existing witness-check
> path before proposing new infrastructure. Use it only where the
> exact claim, candidate content, witness provenance, and scope match;
> do not attach an unrelated valid or invalid tuple merely to obtain
> an admissible verdict.
>
> Preserve the blocked evaluation. Any subsequently obtained result
> must be a separate, attributed assessment. An operator may authorize
> investigation or supply evidence, but attestation alone must not
> manufacture a domain result. A bounded "no matching break found"
> observation may inform search priority while remaining non-decisive
> about the domain goal.
>
> Report either the scoped checked result or the remaining
> verification obligation and its cost. Do not change admission rules
> merely to make the loop advance.

## Pre-work rescoping

The predeclared (g-A) — "add a deterministic-check verifier that
consumes target-violation as evidence for partial_success" — is
inconsistent with the pipeline's documented design boundary.
Verbatim from `internal/verify/deterministic.go`:

- lines 19-22: "A confirmed violation is deliberately left NON-
  decisive (unknown) here: 'the invariant is broken' is necessary
  but not sufficient for success, so the positive verdict is deferred
  to counterexample search."
- `CounterexampleSearch` comment: "It never rewards missing/unknown
  comparison evidence with partial_success."
- `internal/pipeline/evaluation.go:320-322`: the default model tier
  is `FixtureModelVerifier(VerdictUnknown, "low")` because "a
  conservative default model tier: it abstains (unknown) rather than
  manufacturing a verdict, so a bare deployment never launders
  judgment."

The H3 stall on M7 is exactly this design behaving correctly: no
verifier can decide, so the pipeline honestly records
`verification_blocked`. (g-A) as sketched would violate a documented
epistemic-safety boundary.

## The witness path (existing designed H3 escape)

`internal/witness/witness.go`, `cmd/newf/witness.go`,
`internal/pipeline/witness.go`, and
`internal/pipeline/witness_integration_test.go` implement an
exact-integer domain-goal verifier for Erdős–Straus:

- **Contract**: `4·x·y·z == n·(y·z + x·z + x·y)` over exact big
  integers, for a positive-integer tuple `(n, x, y, z)`.
- **Provenance**: an operator-supplied `--note` is REQUIRED and
  refused if absent. Verbatim from the integration test:
  `"produced by the attempted mechanism's step 3 (test fixture)"`.
- **Stamped verdict on witness-valid**: `success` at subject
  domain-goal, kind reproducible-computation, strength reproducible.
- **Stamped verdict on witness-invalid**: `failure` at same subject
  and strength; canonical claim recorded on the evaluation.
- **Malformed tuple**: refused as input error before any write; no
  domain verdict.
- **Missing note**: refused with the string "provenance" in the
  error.

This is the pipeline's own answer to "how does a deterministic H3
escape work" — and it exists today, works, and has an integration
test.

## SGO: production usage of the witness path

Queried all four persisted corpora for evaluations attributed to the
witness checker (`tool_name = 'erdos-straus-witness'` or
`verifier_kind = 'reproducible-computation'`):

```
=== m7 ===
(zero rows)
=== pilot-001 ===
(zero rows)
=== pilot-002 ===
(zero rows)
=== pilot-003 ===
(zero rows)
```

**The witness path has never been exercised on real corpus data.**
Only the integration test uses it.

## SGO: candidate-specific witness obligations in persisted proposals

Surveyed frontier proposals across all four corpora looking for a
claim whose exact form matches the witness path's contract.

### M7 target-violating proposals (`violates_any_target=1`)

Five proposals qualify. Representative example
(`fpr_01M2B8ZNPJ88YP59KPEDF6FXYG`):

- **Structural violation claim**: "introduces a global coupling
  object so the mechanism no longer preserves
  `domain.number_theory.property.identity_carried_solvability`"
- **Cheapest falsification path**: "check whether the global object
  degenerates back into a finite residue cover on the survivor
  classes"

**This is a structural degeneracy check, not a witness-tuple check.**
The proposal names a mechanism FAMILY (introduce a global coupling
object) but does not supply a concrete constructed instance. There
is no `(n, x, y, z)` tuple to submit to the witness checker with
legitimate attribution.

The same pattern holds across all five M7 target-violating proposals
and the M7 non-target-violating proposals — every falsification path
is a structural claim or a signature-vs-family-representative
comparison, none is a witness-tuple identity check.

### H2-escape test proposal (`fpr_01M2BCSS7EJ5WBYNST2J3F0ECY`)

- **Claim**: about pipeline H2 escape, not about a domain goal.
- **Falsification**: "Run `newf frontier generate --proposals-file`
  with this wire, then read the proposal's target verdict from
  `frontier_target_invariants`."
- The falsification already completed; no further domain-goal
  verification obligation.

### Pilot-003 witness-adjacent proposal (`fpr_01M27QA41G7B5WN1354R9EP9TH`)

Uses "witness" in the sense of "small nonresidue witness modulo n",
a different mathematical object from an ES tuple witness. Its
falsification path is "attempt to write any candidate witness-
indexed formula and test it on the first few hundred survivor-class
primes" — a computational task that COULD produce (n, x, y, z)
tuples if the candidate formula were written, but the formula does
not yet exist.

## Answer to the user's question

**No scoped checked result is available on the current corpus.**
Every persisted frontier proposal makes a mechanism-family claim
whose verification obligation is either (a) structural degeneracy
analysis on a not-yet-constructed object, (b) construction plus
testing of a candidate formula that does not yet exist, or (c) an
already-completed pipeline-mechanism claim rather than a domain
claim. The witness path's contract does not match any of these.

**Attaching a witness tuple to any of these proposals would be
misattribution.** For example, a tuple produced by an es-01 (Mordell
polynomial identity) mechanism would be a legitimate witness of ES
for a specific n, but attaching it to an M7 proposal that claims to
introduce a global coupling object would falsely represent the
proposal's mechanism as having produced the tuple. This is exactly
the pattern the user's authorization explicitly forbids: "do not
attach an unrelated valid or invalid tuple merely to obtain an
admissible verdict."

## Remaining verification obligations and their costs

For the interesting M7 target-violating proposals to become witness-
checkable, someone must:

1. **Construct the claimed global coupling object explicitly.**
   Cost: person-days of domain mathematics per proposal (algebraic-
   geometric or analytic construction). Output is a mathematical
   artifact, not a witness tuple.
2. **If the construction produces witnesses at all**, derive from
   the construction one or more concrete `(n, x, y, z)` tuples
   attributable to the mechanism. Cost: additional computation on
   the constructed object.
3. **Only then** does the witness path apply, and it applies as an
   ordinary reproducible-computation verdict — success or failure
   depending on whether the produced tuple satisfies the identity.

For pilot-003's witness-parametrized proposal (SGO from its own
declared falsification path): "whiteboard-plus-script cost" —
bounded, but requires actual domain work.

**None of these obligations is a code change.** All are research
work on the domain problem. This is not a pipeline defect; it is a
reflection of the underlying research difficulty. The pipeline's
verifier is set up to consume concrete claims; the frontier generator
produces mostly abstract mechanism-family claims because that is
what novel research in ES currently looks like.

## Design-space observation (recorded, not proposed as change)

**Generator↔verifier mismatch.** The strongest deterministic
verifier (witness check) is set up for CONCRETE claims: a specific
tuple for a specific n. The frontier generator produces ABSTRACT
claims: mechanism-family declarations targeted at invariants. The
mismatch is not a defect; it is a diagnostic. The mismatch reveals
what the pipeline's actual scientific frontier is: **the cheapest
research work that turns an abstract mechanism-family claim into a
concrete witness claim.**

For ES specifically: the atlas already contains mechanisms that
produce concrete witnesses (es-01, es-03, es-07, es-12). A frontier
proposal that CONSTRUCTS a new mechanism producing witnesses for
previously-unsolved n directly enters the witness path with a
straightforward reproducible-computation verdict. Constructing such
a mechanism is precisely the ES problem the field has been working
on for decades. The generator↔verifier gap is a reflection of the
underlying difficulty, not a research-infrastructure defect.

Observation for search policy (per authorization: "bounded 'no
matching break found' observation may inform search priority while
remaining non-decisive about the domain goal"): the search policy
could downweight proposals whose falsification paths are non-
concrete (structural-only, no candidate tuple derivable), and
upweight proposals whose falsification paths are concrete
(candidate formula + prime range + tuple attribution). This would
route generator effort toward proposals that can be H3-decided,
without any admission-rule change. Recorded for downstream authority;
not proposed as change here.

## H3 gate status: preserved

- No admission rules changed.
- No verifier rewards absence with a positive verdict.
- The blocked evaluation on the H2-escape test proposal
  (`fpr_01M2BCSS7EJ5WBYNST2J3F0ECY`) remains blocked and is not
  attested.
- No new domain-goal verdict manufactured.
- The M7 target-violating proposals remain honestly non-decidable
  by the current verifier chain.

## Verification tier

- Code trace of verifier design boundaries: **CMA** (verbatim
  reading of `deterministic.go`, `verify.go`,
  `frontier_verifier.go`, `evaluation.go`).
- Witness path contract: **CMA** (verbatim reading of
  `witness.go`, `witness_integration_test.go`, `cmd/newf/witness.go`).
- Zero production usage of witness path across four corpora:
  **SGO** (direct sqlite reads).
- No candidate-specific obligation matches the witness contract:
  **SGO** (verbatim structural_violation_claim and
  cheapest_falsification_path from persisted `frontier_proposals`).
- Remaining verification obligation costs: **PE** (informed by the
  proposals' own falsification-path declarations; not independently
  estimated).
- Generator↔verifier mismatch as a diagnostic: **PE** (a plausible
  design-space observation; recorded for downstream authority).
