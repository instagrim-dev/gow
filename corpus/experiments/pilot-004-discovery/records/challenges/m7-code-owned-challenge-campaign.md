# Code-owned challenge campaign on M7 surviving invariants — execution and protocol delta

Recorded: 2026-09-12. Pipeline-space execution + comparison to paper-space
protocol.

Context: after the [Brauer-Manin execution
record](brauer-manin-descent-execution.md) surfaced structural finding H2
(wire proposals cannot verify absence on `Contains(preserves, X)`
invariants), the natural next code-owned pathway was `newf challenge` —
distinct from the wire-authored frontier path, not gated by
completeness-stripping in the same way. This record executes that path
and compares its output to the paper-space 7-probe protocol used in
prior loops (N2a-child-1, N2a-child-2, E15').

## What was executed

Two invocations against the live M7 corpus at HEAD `269de0d`:

```bash
export NEWF_DB=.newf/m7/newf.db

# Single-invariant campaign
./newf --json challenge inv_01M2B8ZNKFASFT9RARSDT5TY3R

# Full-problem campaign
./newf --json challenge --problem prb_01M2B8ZMHRCJN4YVH4CP6SNXTD --all
```

## Results — all 3 surviving invariants

The pipeline's code-owned challenge protocol runs a **3-probe campaign**
per invariant: `known-counterexample`, `success-preserving`,
`bias-critique`. All three invariants ended in state `surviving`.

### Full-probe outcome matrix

| Invariant | Ref | State before → after | KCE unknowns | Support | Threshold margin |
|---|---|---|---|---|---|
| `inv_...T5TY3R` (`identity_carried_solvability`) | L4 | `surviving` → `surviving` | 4/7 | 3 | +1 |
| `inv_...DYY197J` (`confined_to_quadratic_nonresidues`) | L1 | `surviving` → `surviving` | 2/7 | 5 | +3 |
| `inv_...G8X978R` (`class_union_construction`) | L3 | `surviving` → `surviving` | 5/7 | 2 | **0 (at threshold)** |

- **KCE unknowns** = number of failure-side members whose signature
  returned `unknown` under `known-counterexample` (out of 7 eligible).
  High KCE unknowns are a direct consequence of H2 — the pipeline
  cannot verify absence on the failure-side signatures either.
- **Support** = distinct-family support at the mining threshold of 2.
  L3 is exactly at threshold; L4 has margin +1; L1 has margin +3.
- **Contrast violations** = 0/5 for every invariant: no partial-success
  cluster's signature violated any of the three. Same H2 root cause.

Every probe returned `result_summary: unconfirmed` for the same reason:
the pipeline's completeness-stripping design (see
`brauer-manin-descent-execution.md`) makes absence-tests return
`unknown`, and `unconfirmed` is the correct verdict when a negative
search has not been completed.

### Persisted state

Six new challenge campaign rows in `.newf/m7/newf.db` after my re-run
(three from the single-invariant call, three from the `--all` call).
The M7 recipe's initial `challenge --all` at build time already
persisted the first set; these are re-runs producing additional
persisted campaigns.

Example row from L4's re-run:

```
chl_01M2BANTW1W0P8YXVSHYDGEE7K   known-counterexample   unconfirmed   [challenged]
chl_01M2BANTW1W0P8YXVSJ05CX4DD   success-preserving     unconfirmed   []
chl_01M2BANTW1W0P8YXVSJ1AYZP7W   bias-critique          unconfirmed   [surviving]
```

The transitions vector on each row records the state each probe would
transition the invariant to *in isolation*; the final `state_after`
resolves them jointly, with `bias-critique`'s `surviving` transition
overriding `known-counterexample`'s intermediate `challenged` when
support holds.

## Substantive finding — protocol delta

The code-owned protocol is a **proper subset of the paper-space 7-probe
protocol** used in prior loops. Mapping:

| Paper-space probe (my labelling) | Code-owned analog | In pipeline? |
|---|---|---|
| **C1** — family coverage / support recount | `bias-critique` | ✅ |
| C2 — mechanism-use scan (does any atlas member use this exact operator?) | — | ❌ paper-space only |
| **C3** — success-side preservation (is this property discriminating?) | `success-preserving` | ✅ |
| **C4** — known counterexample search | `known-counterexample` | ✅ |
| C5 — merge signal (would this claim merge with a nearby observation?) | — | ❌ paper-space only |
| C6 — support-set variability (would remove-one leave the claim standing?) | — | ❌ paper-space only |
| C7 — causal-core sanity (is the claim near-tautological?) | — | ❌ paper-space only |

**Four paper-space probes are not in the code-owned protocol.** Two of
them produced substantive weakenings in prior challenge runs:

- **C5 (merge signal)** produced the "corpus-hygiene observation" merge
  disposition for `E15'` at commit
  [`d4596c3`](https://github.com/instagrim-dev/gow/commit/d4596c3) —
  the invariant candidate should have been split up-front into
  E15-hygiene (an atlas-quality claim) and E15-mechanism (an actual
  invariant candidate); C5 caught that.
- **C6 (support-set variability)** produced the "movable goalpost"
  weakening for `N2a-child-2` at commit
  [`029f20e`](https://github.com/instagrim-dev/gow/commit/029f20e) —
  the claim's support set relied on "or equivalent object" language
  that let contested cases join or leave the support; C6 caught that.
- **C7 (causal-core sanity)** produced the "near-tautological"
  weakening for `N2a-child-2` at the same commit — the invariant's
  causal core was near-definitional; C7 caught that.

**C2 (mechanism-use scan)** was used in `N2a-child-1` at commit
[`3443a10`](https://github.com/instagrim-dev/gow/commit/3443a10) for
atlas-coverage attribution but did not by itself cause a weakening.

**Pipeline-visible sensitivity at L3.** The code-owned protocol's
bias-critique correctly reports L3's distinct-family support at 2 =
threshold. But it doesn't emit a "fragility" signal — the invariant
appears to survive equally well as L1 (support 5) or L4 (support 3),
even though L3 is one cluster removal away from mining failure. My
paper-space C6 would flag this as "support-set variability = high
sensitivity"; the code-owned bias-critique returns the same
`unconfirmed` label for L1 (huge margin), L4 (small margin), and L3
(zero margin). This is real diagnostic information the code-owned
protocol drops.

## Not a bug — a specification gap

The paper-space additions (C2/C5/C6/C7) are legitimate research
extensions that produced value in 4 of 4 prior challenge runs (E15',
N2a-child-1, N2a-child-2 got weakenings from them). But they are *not
wrong* to be absent from the code-owned protocol. The code-owned
protocol implements a specific, versioned discipline; adding a probe
requires a spec update, a canonical predicate implementation, and
persistence-schema support. This record is *not* proposing that C2/C5/C6/C7
be added; it is naming the delta so the paper-space campaigns' outputs
can be understood as **pipeline-external observations that the code
does not currently reproduce**.

Under the pipeline-space / paper-space distinction (see
CLOSURE-SCORECARD §0), a paper-space challenge run is:

- ✅ a genuine research artefact under AGENTS.md's discipline
- ✅ durable via git history and CMA-verifiable citations
- ❌ **not** a code-owned campaign row and not consumed by any typed
  pipeline operator
- ⚠️ contains signals (C2/C5/C6/C7) that even a follow-on pipeline-space
  campaign would not surface

The last point is new — it means paper-space challenge runs on
invariants that ALSO have pipeline-space state produce *strictly more*
information than the pipeline's own campaign. The paper-space run is
not merely a "translation" of the pipeline campaign into markdown.

## H2 corollary — code-owned campaigns share the completeness limitation

Every code-owned probe on every invariant returned `unconfirmed`. The
common reason across all 9 probe runs (3 invariants × 3 probes) is:
one or more eligible members (failure-side, or contrast-side, or the
signature under bias-critique) evaluated `unknown` under the invariant
predicate. This is the exact same completeness-stripping mechanism that
H2 identifies for wire proposals: `internal/canon/admission.go` resets
completeness for untrusted signatures, `internal/invariant/predicate.go`
returns `unknown` on absence when completeness ≠ `Complete`.

So H2 is not specific to the wire-authored frontier path. It affects
the code-owned challenge path too. Any signature that lacks a
`CompletenessComplete` marker (which is every atlas member and every
wire proposal under current design) will contribute to `unknown`
verdicts in absence-based checks.

**The only way to get decisive absence-based verdicts under the
current design is for signatures to carry `CompletenessComplete`
markers.** That is a code path change (the normalization layer would
need to author such markers based on operator attestation or specific
extraction confidence), not a challenge-protocol change.

## Downstream consequences (proposed, not applied here)

1. **Do not run additional paper-space challenges on the M7 invariants
   expecting parity with the code-owned campaign.** The two protocols
   produce different, non-overlapping information; running both is
   legitimate (paper-space adds C2/C5/C6/C7), but the pipeline-space
   record won't be updated by a paper-space run — even under the §0
   Class-move table's "Paper-space challenge campaign → Pipeline-space
   challenge campaign" move.
2. **L3 is the invariant most vulnerable to challenge.** Its support
   sits at threshold. A future atlas addition that fails to preserve
   L3 (or fails to preserve L1/L4 while contributing to L3's support
   population) could drop L3 below threshold on the next mining run.
   The code-owned bias-critique doesn't emit this signal but the raw
   `distinct_family_support: 2` value is code-owned data.
3. **CompletenessComplete markers would radically change what the
   pipeline can verify.** Every `unconfirmed` in this campaign, and
   every `unknown` verdict in the two wire-frontier runs, would flip
   to a decisive verdict. This is a code-path decision (via
   normalization or operator attestation), not a research decision.
   Whether the pipeline SHOULD accept CompletenessComplete assertions
   is a separate question — currently the design deliberately doesn't,
   for exactly the reasons AGENTS.md §Epistemic invariants states.
4. **Adding C6 (support-set variability) to the code-owned protocol
   would be low-cost.** The predicate `min(distinct_family_support -
   threshold) / support` is derivable from data already persisted. The
   protocol change would add a fourth probe with a computable verdict.
   Adding C7 (causal-core sanity) is much harder — near-tautology
   detection requires semantic reasoning about the predicate's
   definition, not just counting. Recording as a design-space
   observation, not a change proposal.

## Verification tier

- Challenge campaign results (per-probe verdicts, support values,
  transitions): **SGO** — verbatim from `newf --json challenge` output.
- Mining support metrics (`distinct_family_support`,
  `failure_coverage`, `contrast_violating`): **SGO** — from
  `newf --json invariant show`.
- Protocol-delta table (C1-C7 → code-owned mapping): **PE** — the C1-C7
  labels are my paper-space taxonomy; the mapping to code-owned probe
  names is model-assisted. An independent operator could disagree
  about whether my C1 = `bias-critique` (they both do support
  recomputation but weight it differently) or whether C3 =
  `success-preserving` (semantic near-equivalent, not identical).
- H2 corollary: **CMA** — same code trace as
  `brauer-manin-descent-execution.md` §H2, applied to challenge
  probe outcomes instead of frontier violation outcomes. The mechanism
  is identical.
- Downstream consequence claims: **PE** — model-assisted routing.
- Same same-model-family caveat as parent records — not independent
  human verification.
