# (a‴) survey — enum-axis and Not-form invariants on the M7 corpus

Recorded: 2026-09-12. Design-space observation with pipeline-space
grounding. No code change proposed here without further authorization.

## Question

Path (a‴) in the CLOSURE-SCORECARD asked: mine an enum-axis or
Not-form invariant on the M7 failure population, because these
predicate shapes are the only ones on the current corpus that can
advance a wire proposal past the H2+H3 terminal state.

Before executing (a‴), two prior questions had to be answered:

1. Does the deriving miner *emit* enum-axis or Not-form predicates?
2. Would any such predicate reach `recurring` support on the M7
   failure population?

## Answer 1 — CLI-default miner cannot emit enum/Not shapes (CMA)

`internal/provider/invariant_fixture_derive.go` is the CLI default
(via `App.MineInvariants` → `provider.NewDerivingFixtureInvariantMiner`).
Its `Mine` method (lines 49-100) emits proposals of exactly one
shape:

```go
proposals = append(proposals, CandidateProposal{
    Predicate: invariant.Predicate{
        Schema: invariant.PredicateSchemaV1,
        Root:   invariant.Node{Op: invariant.OpContains, Field: occ.field, CanonicalID: string(id)},
    },
    ...
})
```

Field `occ.field` is always either `invariant.FieldPreserves` or
`invariant.FieldOperators`. **The miner cannot emit `OpEquals`,
`OpIn`, `OpNot`, `OpBoundary`, `OpAll`, or `OpAny`.** The predicate
schema (`internal/invariant/predicate.go` lines 42-53) admits all of
them, but the fixture miner has no template.

The alternate `FixtureInvariantMiner` (canned proposals keyed on the
request fingerprint) admits arbitrary predicate shapes, but cluster
IDs are freshly-minted ULIDs at CLI time so fingerprints cannot be
known ahead of time — the CLI cannot use this miner.

**Consequence:** the M7 corpus's three surviving invariants (L1
`Contains(preserves, density_ceiling)` support=5, L3
`Contains(preserves, sample_averaging)` support=2 at threshold, L4
`Contains(preserves, …)` support=3) are all `Contains(preserves, X)`
shape because that is the only shape the CLI-default miner can
produce.

## Answer 2 — Enum-axis wire proposals escape the H2 completeness stall (CMA)

The H2 mechanism (`internal/canon/admission.go` lines 108-124) strips
`SetFieldCompleteness` to `Unobserved` for untrusted wire signatures,
causing `Contains(preserves, X)` predicates to return `VerdictUnknown`
on absence rather than `VerdictViolates`.

**Enum-axis fields are NOT set fields.** `AdmitProposalSignature`
copies the entire signature via `out := sig` (line 48), then mutates
only:

- set-field claim slices (`Representations`, `Operators`,
  `Assumptions`, `Preserves`, `Breaks`, `AuxiliaryObjects`) via
  `admitClaims`;
- `Boundaries` (special handling);
- `SetFieldCompleteness` (stripped).

**`Posture` (locality/construction/uncertainty) is not touched.** It
flows through admission unchanged.

Enum-axis evaluation (`internal/invariant/predicate.go` lines
495-505):

```go
case OpIn, OpEquals:
    value, known := enumAxisValue(n.Field, sig)
    if !known {
        return VerdictUnknown
    }
    for _, v := range n.Values {
        if v == value {
            return VerdictSatisfies
        }
    }
    return VerdictViolates
```

The verdict is `VerdictSatisfies` or `VerdictViolates` — **decisive**
— whenever the posture axis value is `known` (non-empty, not
"unknown"). Wire proposals declare posture explicitly
(`internal/provider/untrusted_proposer.go` lines 39-41: `Locality`,
`ConstructionMode`, `UncertaintyMode` are required JSON fields),
so `enumAxisValue` returns `known=true` for admitted wire
signatures.

`OpNot` (lines 510-518) simply flips a decisive verdict, so
`Not(Equals(locality, X))` is also decisive when the underlying
enum axis is known.

**Consequence:** enum-axis or Not-of-enum invariants would receive
decisive verdicts from wire-authored frontier proposals, unlike
`Contains(preserves, X)` invariants which return `Unknown`. This is
the escape hatch from H2 and, downstream, the escape hatch from H3.

## Answer 3 — Hand-evaluation of candidate enum/Not predicates on M7 families (SGO + CMA)

### Source-grounded observation

M7 corpus cluster reps and posture axes (verbatim from
`sqlite3 -header .newf/m7/newf.db`):

```
c_ord|outcome_class    |axis        |value        |claim_status
0    |partial_failure  |locality    |local        |unknown
0    |partial_failure  |construction|constructive |unknown
0    |partial_failure  |uncertainty |deterministic|unknown
1    |partial_success  |locality    |global       |unknown
1    |partial_success  |construction|existential  |unknown
1    |partial_success  |uncertainty |deterministic|unknown
2    |partial_failure  |locality    |global       |unknown
2    |partial_failure  |construction|existential  |unknown
2    |partial_failure  |uncertainty |probabilistic|explicit
3    |partial_failure  |locality    |local        |unknown
3    |partial_failure  |construction|constructive |unknown
3    |partial_failure  |uncertainty |deterministic|unknown
4    |partial_failure  |locality    |local        |unknown
4    |partial_failure  |construction|constructive |unknown
4    |partial_failure  |uncertainty |deterministic|unknown
5    |failure          |locality    |global       |unknown
5    |failure          |construction|existential  |unknown
5    |failure          |uncertainty |deterministic|unknown
6    |partial_success  |locality    |global       |unknown
6    |partial_success  |construction|constructive |unknown
6    |partial_success  |uncertainty |deterministic|unknown
7    |failure          |locality    |global       |unknown
7    |failure          |construction|existential  |unknown
7    |failure          |uncertainty |deterministic|unknown
8    |partial_success  |locality    |global       |unknown
8    |partial_success  |construction|existential  |unknown
8    |partial_success  |uncertainty |deterministic|unknown
9    |partial_success  |locality    |mixed        |unknown
9    |partial_success  |construction|existential  |unknown
9    |partial_success  |uncertainty |deterministic|unknown
10   |partial_success  |locality    |mixed        |unknown
10   |partial_success  |construction|existential  |unknown
10   |partial_success  |uncertainty |deterministic|unknown
11   |partial_failure  |locality    |mixed        |unknown
11   |partial_failure  |construction|existential  |unknown
11   |partial_failure  |uncertainty |deterministic|unknown
```

All 12 clusters are isolates (member_count=1, isolate=1), so each
cluster = one family. Failure-side (`partial_failure` +
`failure`) = 7 families: {c0, c2, c3, c4, c5, c7, c11}. Success-side
(`partial_success`) = 5 families: {c1, c6, c8, c9, c10}.

### Checked mathematical argument (per-candidate)

Support = distinct failure-side families where the predicate returns
`VerdictSatisfies`. Contrast prevalence: same denominator on success
side. `MinSupport` default (per `internal/pipeline/invariant.go` line
177) = 2. Association reaches `recurring` when
`support >= minSupport` AND failure-prevalence > success-prevalence
(engine.go `classify`).

| # | Candidate predicate | Failure-side matches | Support | F prev | S prev | F > S | Reaches `recurring`? |
|---|---|---|---|---|---|---|---|
| C1 | `Equals(locality, local)` | c0, c3, c4 | **3** | 3/7 = 0.43 | **0/5 = 0.00** | ✓ | **YES — 100% specificity** |
| C2 | `Equals(locality, global)` | c2, c5, c7 | 3 | 3/7 = 0.43 | 3/5 = 0.60 | ✗ | No (F < S) |
| C3 | `Equals(locality, mixed)` | c11 | 1 | — | — | — | No (support < 2) |
| C4 | `Equals(construction, constructive)` | c0, c3, c4 | **3** | 3/7 = 0.43 | 1/5 = 0.20 | ✓ | **YES** |
| C5 | `Equals(construction, existential)` | c2, c5, c7, c11 | 4 | 4/7 = 0.57 | 4/5 = 0.80 | ✗ | No (F < S) |
| C6 | `Equals(uncertainty, deterministic)` | c0, c3, c4, c5, c7, c11 | 6 | 6/7 = 0.86 | 5/5 = 1.00 | ✗ | No (F < S) |
| C7 | `Equals(uncertainty, probabilistic)` | c2 | 1 | — | — | — | No (support < 2) |
| C8 | `Not(Equals(locality, global))` | c0, c3, c4, c11 | **4** | 4/7 = 0.57 | 2/5 = 0.40 | ✓ | **YES** |
| C9 | `Not(Equals(construction, existential))` | c0, c3, c4 | 3 | 3/7 = 0.43 | 1/5 = 0.20 | ✓ | YES (dual of C4) |
| C10 | `Not(Equals(uncertainty, deterministic))` | c2 | 1 | — | — | — | No (support < 2) |
| C11 | `All(Equals(loc, local), Equals(con, constructive), Equals(unc, deterministic))` | c0, c3, c4 | 3 | 3/7 = 0.43 | **0/5 = 0.00** | ✓ | YES (100% specificity; redundant to C1) |

**Three genuinely-distinct enum-axis or Not-form shapes reach
`recurring` on M7:** C1, C4, C8.

**C1 is the strongest candidate**: support=3, success-side prevalence
= 0.00 (perfect discrimination), and the smallest single-axis
predicate. C4 is a good complement (also perfect discrimination
across the constructive axis, one success-side match). C8 has the
largest support but weaker discrimination.

## Meaning (proposed explanation)

C1 in natural language: "failed approaches to ES have `locality =
local`; no successful approach does". This aligns with the
paper-space N2a chain (local density-and-averaging ceilings) but
lifts the invariant from the SET AXIS (`preserves` contains
`density_ceiling`) to the ENUM AXIS (`locality = local`). It is a
higher-level structural claim about the *shape* of the failing
mechanism family, not the *content* of what they preserve.

C4 in natural language: "failed approaches to ES have `construction
= constructive`; only one successful approach does (c6 =
constructive/global/deterministic)." This is weaker than C1 because
of the c6 counterexample. The specificity drop (0.00 → 0.20) suggests
constructive-and-local is doing the work, not constructive alone.

C11 (conjunctive) formalises "constructive-and-local-and-deterministic
failures" — the classic explicit-construction / small-modulus / bounded
class of ES attempts. Redundant to C1 on this corpus because c0, c3, c4
share all three values.

## Why this matters for the pipeline (design-space claim)

If the deriving miner were extended to emit
`OpEquals(FieldPosture, value)` predicates for every posture axis
value where the failure-side count ≥ `MinSupport` (or a stricter
"failure-side and NOT success-side" filter), the M7 corpus would gain
at least three new invariants — the first M7 invariants for which:

1. **Wire proposals produce decisive violation verdicts** (H2
   escaped — enum-axis evaluation does not depend on
   `SetFieldCompleteness`).
2. **Decisive violation verdicts can produce `failure`/`partial_failure`
   at evaluation**, entering `evaluated_failures` (H3 escaped, subject
   to whether the verifier hierarchy actually reaches a decisive
   verdict when the violation is decisive).
3. **Admission is reachable** by rule (for `deterministic`/
   `reproducible` strength failures) or by operator attestation (for
   `model-judged-failure`).

The structural asymmetry is:

> `Contains(preserves, X)` invariants stall at H2 because
> wire signatures cannot claim
> `SetFieldCompleteness[FieldPreserves] = Complete` — admission
> strips that claim.
>
> `Equals(<posture>, X)` invariants do not stall at H2 because
> posture axes are single-valued, wire signatures declare them
> explicitly, and admission does not touch them.

## Verification tier of each claim in this record

- Miner-shape restriction: **CMA** — traced through
  `invariant_fixture_derive.go` at HEAD `fd4c3ef`.
- Admission does-not-touch-posture: **CMA** — traced through
  `internal/canon/admission.go`.
- Enum-axis evaluation is decisive: **CMA** — traced through
  `internal/invariant/predicate.go`.
- Family posture-axis values on M7: **SGO** — verbatim sqlite3
  output.
- Support and contrast counts: **CMA** — deterministic counting.
- "Reaches `recurring` on M7": **CMA** — direct application of
  `classify()` from engine.go.
- "Wire proposals produce decisive verdicts on these": **PE** —
  predicted by the code trace; testable end-to-end by supplying a
  proposal and observing evaluation output.
- "Admission is reachable": **PE** — direct consequence of the
  above IF the verifier hierarchy produces a decisive verdict.
  The current fixture verifier's behaviour under decisive-violation
  input has not been observed; it might still return
  `verification_blocked` for reasons unrelated to H2/H3.

## What this loop did NOT do

- Did **not** modify `DerivingFixtureInvariantMiner` to emit
  enum-axis proposals. That is a code change (~30 LOC + tests +
  reuse-key rev bump because it changes miner_version). Not
  authorized here.
- Did **not** add a `--proposals-file` path to
  `newf invariants mine` (branch A of (a‴)). Also a code change.
- Did **not** actually invoke the pipeline with an enum-axis
  proposal to confirm the H2/H3-escape prediction. That requires
  either of the code changes above OR a Go test that constructs
  the proposal in-process (feasible but out of loop scope).

## Consequences for the closure scorecard (proposed)

1. **Path (a‴) refined.** The mining-based fork ("run the miner
   again") is impossible on the current codebase: the CLI-default
   miner has no enum/Not templates and no CLI supplies canned
   proposals. Two code-change forks remain:
   - **(a‴-A)** Add `--proposals-file` to
     `newf invariants mine`. Small; mirrors the existing
     `frontier generate --proposals-file` shape.
   - **(a‴-B)** Extend `DerivingFixtureInvariantMiner.Mine` to
     emit `OpEquals(<posture>, value)` where failure-side count ≥
     `MinSupport` and failure prevalence > success prevalence.
     Small; the classification rule is exactly the one already
     in engine.go.
2. **The prediction is durable and falsifiable.** If (a‴-A) or
   (a‴-B) is executed, C1 `Equals(locality, local)` should mine
   at support=3 on the M7 corpus with 100% failure-side
   specificity, and a wire proposal declaring `posture.locality =
   "global"` against it should return decisive
   `VerdictViolates` at frontier ranking, which will let us test
   whether H3 (`verification_blocked` at evaluation) actually
   holds or breaks on decisive-violation input.
3. **Records the structural insight**: the H2 stall is
   set-field-completeness-scoped, not predicate-scope-scoped.
   Any predicate shape that does not depend on
   `SetFieldCompleteness` (enum axes, boundaries with resolved
   labels, `Contains(operators, X)` on `FieldOperator` completeness)
   escapes it.

## Notes and honest self-critique

- The support/contrast counts assume `member_count=1` and
  `isolate=1` for every M7 cluster, verified in the SQL survey.
  If new members were later added to any cluster (e.g. via a
  cluster rebuild), the counts might change; the mining engine
  caps distinct-family support at one per cluster.
- The 0/5 success-side prevalence for C1 is a fragile boundary. A
  single success-side cluster with `locality = local` would
  weaken C1's specificity. This is exactly the kind of
  fragility the paper-space challenge protocol's C6
  (support-set variability) probe would flag; recorded in the
  scorecard's design-space observation.
- "PE" tier used for the pipeline-end-to-end predictions is
  honest: the code trace is CMA, but the FULL behaviour
  (miner → engine → admission → frontier → evaluate) has not
  been observed for enum-axis predicates and could exhibit
  interactions this analysis missed.
