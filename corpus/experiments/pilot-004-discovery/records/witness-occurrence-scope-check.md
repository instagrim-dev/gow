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
