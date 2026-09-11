# M7 positive control — populated invariant discovery (known ground truth)

## Recorded conclusion

**Populated integration-control package: the discovery path fires and
discriminates on this fixture.** On a deliberately synthetic fixture with known
ground truth, the full research path produced actual records — scoped failure
evidence, a nonempty candidate set, a completed-negative challenge campaign, a
surviving invariant, a generated proposal whose **persisted violation verdict is
read back as `violates`**, and a **decisive structural recovery** of the
withheld target. Mutation controls at two tiers flip each expected result for
the corresponding reason.

Complement to `../2026-09-10-esr-negative-control/` (expected abstention on
unsuitable inputs). Together: negative-path control + populated positive
control.

## Where it lives (auditable, not `/tmp`)

Executable and deterministic: `internal/pipeline/positive_control_integration_test.go`,
run by `go test ./internal/pipeline/ -run PositiveControl`. Ground truth is
authored in the fixture, so the assertions are the audit trail.

- provider: `fixture` (deterministic; no model training)
- recovery rule: `recovery-rule/v1`; profile `classify/v1`
- mode: `blinded`
- miner: `minPreservesMiner` (deterministic mined predicate); challenger:
  `biasOnlyChallenger` — a **support recount**, NOT a known-counterexample
  attack. The counterexample refutation contract is covered separately (below).

## Required path — each link is an asserted persisted record

```text
scoped failure evidence   corpus failure-side families (resolved axes)
-> comparable mechanisms  real clustering, not "cannot compare"
-> nonempty candidate     mined predicate `preserves contains mean_growth_rate`
-> executed challenge     completed-negative campaign -> `surviving`
-> eligible guided target the surviving invariant B3 may attack
-> generated proposal     preserves ONLY residue_locality, preserves field Complete
-> verified break         READ BACK: persisted per-target verdict == violates
-> decisive recovery      proposal mechanism-near withheld target -> structural_recovery
```

The **read-back chain** ties: expected predicate fingerprint → surviving
invariant ID → proposal target ID → exact persisted signature revision
(occurrence content hash + JSON) → violation verdict `violates` → recovered
experiment member. This distinguishes *recovering a representation* from
*establishing a decidable predicate violation* — `AssessProposals` compares
proposal↔target and never consumes the violation verdict, so `structural_recovery`
alone would not establish the intervening verified-break step.

## Scope: what B0-vs-B3 is here

**A target-conditioned generation positive control, paired with a
no-target/no-proposal control.** The generator returns the authored matching
signature per surviving target and nothing when there are no targets; B0's
emptiness and B3's output are encoded by the fixture. It shows the treatment
path can be invoked and scored. It does **not** show an undirected proposer
attempted and failed, nor that the inferred invariant supplied information the
generator needed. The effectiveness comparison is reserved for the later
experiment with a non-predetermined proposer.

## Mutation controls — two tiers

Same-path integration mutations (same pipeline configuration, fresh workspaces,
persisted records asserted across the storage/reporting boundary):

| Mutation | Expected change | Test |
|---|---|---|
| Support threshold: 2 vs 1 failure families sharing the id, CLI-default deriving miner, explicit MinSupport 2, via `MineInvariants` | candidate with persisted support ≥ 2 vs **zero** candidates | `SamePathMatrix/support_threshold` |
| Unknown target axis: withheld target's preserves label does not resolve (real admission path) | persisted arm `unknown_count=1`, no recovery, conclusion `inconclusive` (via `ShowExperiment`) | `SamePathMatrix/unknown_target_axis` |
| Budget exhaustion: 2 distinct proposals, evaluation budget 1, full experiment run | persisted `decisive=1 unassessed=1 consumed=1`, stopping `budget_exhausted`, conclusion `inconclusive` — never coerced to `no_recovery` | `SamePathMatrix/budget_exhaustion` |
| Completeness flip: same resolved features, preserves completeness Complete → Unobserved | persisted violation verdict degrades to **`unknown`** while recovery stays decisive | `UnobservedCompletenessBreaksVerification` |

Component-level mutation controls (direct provider/evaluator calls pinning one
boundary each): `TestPositiveControlMutation{RemoveEvidence,UnknownFieldIsInconclusive,BudgetExhaustion}`.
The known-counterexample refutation contract (one in-atlas counterexample
falsifies a `recurring` claim; a `contrast_observed` claim survives it) is the
existing regression `TestVerifyKnownCounterexampleRefutationDependsOnClaimKind`
in `internal/invariant` — component coverage, distinct from this control's
bias-recount campaign.

Note: these are **input mutations**. The separate verification-contract
requirement — disabling a guard and confirming the test catches the defect — is
not replaced by them and remains future work.

## Two structural findings (narrowed)

1. **The current builder cannot establish absence-based violations of positive
   `contains` predicates when it leaves the relevant set field unobserved.**
   `internal/canon/signature.go` marks every set field
   `CompletenessUnobserved`; `invariant.Evaluate` returns `violates` for
   absence only under `CompletenessComplete`. (Enum mismatches and negated
   predicates can still verify without proving set absence.) The fixture
   authors the flag under its synthetic ground-truth convention. **The
   production fix must not be "let the model set Complete"**: completeness
   needs a defined scope and justification — "all entries in this declared
   field were parsed" is enforceable; "all properties preserved by this method
   were identified" is a much stronger claim, and they cannot share an
   unqualified authority level.

2. **The default deriving fixture generator's only break preserves an
   out-of-vocabulary id** (`core.property.global_coupling`, intentionally
   inadmissible — see `success_integration_test.go`), so a vocab-normalized
   target can never be mechanism-near it; that is why the shipped end-to-end
   test yields `no_recovery`. This control uses an in-vocab preserves id
   (`residue_locality`) shared by proposal and target. The **target** side
   reaches its representation through the real vocabulary-admission path; the
   **proposal** side still directly authors resolved claims. The next control
   should demonstrate a proposal reaching the same comparable representation
   through the intended admission path, without trusting a provider-supplied
   `resolved` status or silently inventing an alias.

## What this validates — and what it does not

Validated: the mechanics of populated discovery — mining support thresholds,
challenge survival on a completed negative, persisted verified structural
violation, decisive recovery classification, and the budget/unknown discipline —
all discriminate under known ground truth, with intermediate records asserted.

**Not** validated: discovery ability. The proposal is authored by a fixture with
the answer built in. Testing the research thesis requires independently assessed
source cases and a proposal-producing system whose outputs are not predetermined
by the fixture.

**Next milestone (per review): a persisted-input positive control** — the same
synthetic scenario entering through ordinary ingestion and normalization,
carrying justified completeness and vocabulary mappings through storage, then
challenged, generated against, and assessed, with exact records asserted at each
boundary. Deterministic proposals suffice; a live model is not needed to prove
the interface works.

## Validation scope

Gates (`go build`, `go vet`, `go test ./...`, `gofmt -l`) are author-reported
from the local run; the reviewing party did not independently execute them.
