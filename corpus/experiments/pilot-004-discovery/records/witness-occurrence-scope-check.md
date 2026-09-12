# Witness-occurrence attribution — fresh SGO scope-check on M7's proposal corpus

Predeclared as a **bounded scope check**, not proposal authoring. The
question: does any M7 occurrence carry, in its content, a legitimate
attribution obligation for the witness path?

## Contract being scoped against

`internal/witness/witness.go` implements an exact-integer domain-goal
check for Erdős–Straus with the following requirements:

- The proposal occurrence must be attributable via F2's pin-and-attach
  gate: a witness verdict must bind a specific `(frontier_generation_run_id, proposal_id, content_hash)`
  triple (`witness.go:127-128`, `witness.go:258`, `witness.go:285`).
- The claim being verified must be **decisive-strength** at
  `subject: domain-goal` — a concrete `(n, x, y, z)` integer 4-tuple
  such that `4·x·y·z = n·(y·z + x·z + x·y)`.
- Mandatory provenance note ("unnoted claims" are refused).
- On success, the verdict is stamped
  `verifier_kind=domain-witness` at
  `verification_strength=reproducible-computation`.

The witness path is the pipeline's only mechanism that can produce a
non-abstained `verdict != verification_blocked` at
`verification_strength = reproducible-computation` for M7's problem
class. Any proposal that lacks the concrete-tuple content described
above CANNOT invoke this path, regardless of the mechanism-level
sophistication of its claims.

## Method

Enumerated every row of `frontier_proposals` in `.newf/m7/newf.db` at
HEAD `aa3c4a3`, extracting `structural_violation_claim`,
`novelty_argument`, and `cheapest_falsification_path` for each. Scanned
the concatenated content for:

- Explicit `(n, x, y, z)` integer 4-tuples.
- Statements of `4/n = 1/x + 1/y + 1/z` or `4xyz = n(xy+yz+zx)` with
  numeric instantiations.
- Any reference to a specific `n` accompanied by a specific `(x, y, z)`.
- Any provenance annotation attributing an integer tuple to a source.

Also queried `evaluations` for `verifier_kind`, `verification_strength`,
and `verification_subject` values.

## Findings

### 1. Proposal-content survey (SGO, verbatim)

| Proposal | Content class | Concrete tuple? |
|---|---|---|
| `fpr_...E4T95` | B0/B1/B2 baseline restatement | No |
| `fpr_...PYDJ` | B0/B1/B2 baseline restatement | No |
| `fpr_...FAWA` | B0/B1/B2 baseline restatement | No |
| `fpr_...607H` | Generic re-representation (undirected) | No |
| `fpr_...FXYG` | B3 invariant-guided: "global coupling object" (abstract) | No |
| `fpr_...EQAQ` | B3 invariant-guided: "global coupling object" (abstract) | No |
| `fpr_...MVX3` | B3 invariant-guided: "global coupling object" (abstract) | No |
| `fpr_...D2YB1` (joint-crossing, N2a) | Mechanism-level: delta-method + composition-genus. Falsification step 4 mentions "for n ≤ 10^4 ... check whether the proposed apparatus produces representations." **Prospective**, no tuple in-proposal. | No |
| `fpr_...TFTE` (Brauer-Manin) | Mechanism-level: Br(S_n) + 2-descent. Falsification step 4 mentions "small-case verification against known representations for n in Mordell-covered classes." **Prospective**, no tuple in-proposal. | No |
| `fpr_...F0ECY` (H2-escape test) | Pipeline-mechanical (posture assertion) | No |

**Only occurrence of the word "witnesses"** in the entire M7 proposal
corpus appears in `fpr_...D2YB1`:

> "producing witnesses that simultaneously (a) achieve polynomial main-term growth and (b) cover the QR-survivor classes"

This is metaphorical usage denoting "instances-of-representation the
mechanism would produce," not a concrete `(n, x, y, z)` tuple asserted
with provenance. It does not satisfy the witness-check contract.

### 2. Evaluation-history survey (SGO, aggregate)

Query on `evaluations`:

```
n_evals: 3
verdict:              verification_blocked   (all 3)
verifier_kind:        model-judgment         (all 3)
verification_strength: single-model-judgment (all 3)
verification_subject: domain-goal            (all 3)
```

- Zero evaluations at `verifier_kind = domain-witness`.
- Zero evaluations at `verification_strength = reproducible-computation`.
- The F2 occurrence-attribution machinery has never been exercised on
  M7's corpus.

## Conclusion

**Scope check: negative.** M7's current proposal corpus contains no
in-scope obligation for the witness-occurrence attribution mechanism.
All ten proposals are either baseline restatements, generic
re-representations, mechanism-level structural claims (which stall at
`verification_blocked` under the deliberate non-decisive contracts of
`DeterministicCheck`, `CounterexampleSearch`, and the default model
tier), or pipeline-mechanical tests.

This generalizes and confirms the (g-scoped) finding by direct
enumeration of the full corpus (previously verified only for three
proposals).

## What discharge would require (unchanged)

Discharge of the witness-occurrence attribution obligation on M7
requires ONE of:

1. **Proposal-input work by an operator**: author a wire proposal
   whose content declares a specific `(n, x, y, z)` tuple as a
   decisive-strength domain-goal claim with attributed provenance.
2. **Existing proposal + operator-supplied witness**: an operator
   attaches an `(n, x, y, z)` tuple to an existing proposal and
   supplies provenance. The pipeline's F2 gate would refuse
   attachment if the tuple is not attributable to the specific
   generation-occurrence membership of that proposal.

Both are **operator-input operations**. Neither is agent-tractable
without external authorization, and neither is a code-change concern.

## Verification tier

- Enumeration of proposal count and per-proposal content: **SGO**
  (verbatim `sqlite3` output on `.newf/m7/newf.db`).
- Absence of witness-strength evaluations: **SGO** (aggregate query on
  `evaluations`).
- Contract requirements: **CMA** (verbatim `witness.go` reading).
- "The witness path has never been exercised on M7": **CMA + SGO**
  (contract + zero rows at required verifier_kind).
- Discharge of the witness-occurrence attribution obligation: **not
  claimed**. Not agent-discharge­able; recorded as a bounded scope
  check confirming the barrier's structure.

## H3 gate status: PRESERVED (unchanged)

No admission rules changed, no new evaluations authored, no proposal
inputs constructed. This is a scope check on the existing corpus, not
an escape attempt.

## Boundary note

A **bounded "no matching witness obligation found"** observation may
inform search-priority discussions (e.g., the (p) generator-bias
follow-on referenced in `records/o-closed-lever-exists.md` gains
concreteness here: the current generator produces zero witness-carrying
proposals out of ten, so any move toward concrete-witness generation
would be a strictly new capability rather than a bias adjustment to an
existing signal). This observation is non-decisive about the domain
goal and does not alter any admission rule.

---

## Clarification appended 2026-09-12 after concurrent-writer discovery

At the time of authoring this scope check, I could not locate the
reviewer's "C7" and "C8" labels against any documented obligation set
(the paper-space challenge protocol has only C1–C7 as probe kinds).
The scope check was authored assuming "witness-occurrence attribution"
referred to F2 gate exercise on M7 proposal content specifically.

A concurrent writer's uncommitted-at-time-of-scoping additions to the
tree now surface a **different, code-owned C1–C8 framework**:

- `internal/pipeline/review.go` (new)
- `internal/pipeline/review_coverage.go` (new)
- `internal/pipeline/review_c1c8_disposable_test.go` (disposable test
  documenting cases C1–C8 for obligation `current-assessment-authority`
  under policy `P1`, per recipe
  `docs/reviews/prompts/recipes/assessment-admission-decision.md`).

Under that framework:

- **C1** = baseline assessment under P1/D0.
- **C2** = withheld control — model-only failure stays out of the
  population.
- **C3** = independently checked observation enters through real
  admission (witness path integration at the assessment-population
  boundary).
- **C4** = relevant change — eligibility not inherited from A0.
- **C5** = unrelated write must not stale.
- **C6** = reassessment against A1 under the pinned policy.
- **C7** = **historical replay must not displace compatible current
  context** (NOT the paper-space "correlation vs causal" probe).
- **C8** = **derived coverage projection** — previously a recorded
  gap; the concurrent writer's `review_coverage.go` appears to close
  it in code.

Under this reading, the reviewer's "C7 obligation, witness-occurrence
attribution obligation, C8 obligation" refers to review-INTEGRATION
obligations at the pipeline↔review boundary, NOT to paper-space
challenge probes and NOT to M7 proposal witness-content provisioning.

## What this scope check DOES establish (unchanged, still valid)

The SGO on M7's proposal corpus stands: 0/10 proposals carry an
in-content `(n, x, y, z)` claim, 0/3 evaluations use the witness path.
This is factual observation about the corpus, independent of what the
reviewer's terminology maps to.

## What this scope check DOES NOT establish (corrected)

- **Discharge of the C7 obligation** (review-recipe sense): the
  scope check does not address historical-replay-versus-current-authority
  staleness gates. The concurrent writer's C7 case-under-test
  in `review_c1c8_disposable_test.go` is where that obligation is
  being exercised.
- **Discharge of the C8 obligation** (review-recipe sense): the
  scope check does not address the derived-coverage-projection gap.
  The concurrent writer's `review_coverage.go` and the `GenerateReviewCoverage`
  App method appear to be filling that gap in code.
- **Discharge of the witness-occurrence attribution obligation**
  (review-recipe C3 integration sense): the scope check confirms
  the M7 corpus has no in-content witness claims, which is orthogonal
  to whether the F2 attribution machinery integrates correctly with
  the assessment-population and coverage projection when it IS
  invoked. Case C3 in the concurrent writer's disposable test
  exercises that integration directly.

## Correct scope of this record's contribution

- **Bounded, valid SGO** on M7's proposal-content witness-obligation
  distribution.
- **Corroborates** the design-space observation about the
  generator↔verifier mismatch surfaced in (g-scoped) and (o-closed).
- **Does not discharge** any of the three obligations the reviewer
  identified; those are review-INTEGRATION obligations that a
  concurrent writer's in-flight code changes appear to be addressing.
- **Does not conflict with** the concurrent writer's work; the two
  investigations are targeting adjacent but distinct questions.
