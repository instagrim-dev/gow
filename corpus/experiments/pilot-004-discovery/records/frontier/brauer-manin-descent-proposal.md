# Frontier proposal (pipeline-space) — Brauer–Manin / 2-descent construction on the ES surface

Recorded: 2026-09-12. Predeclared **before** wire authorship and pipeline run.

Context: first frontier proposal designed against the *actual* three surviving
invariants in the M7 live corpus, not against the paper-space N2a landscape.
Directly follows from the recommendation added in commit
[`66c96a5`](https://github.com/instagrim-dev/gow/commit/66c96a5) after the
pipeline-space vs paper-space distinction was named.

Epistemic-tier legend (unchanged from parent records):

```text
SGO  Source-grounded observation
PE   Proposed explanation
CCS  Completed counterexample search
IIP  Inapplicable or inconclusive probe
CMA  Checked mathematical argument
```

---

## Target invariants (from the live M7 corpus)

Three surviving invariants under `mechanism/v3`, all mined from the failure
population of `corpus/train/*.md` with reference-property enrichment from
`pilot-001-ledger`:

| Invariant ID | Statement | Reference | Enriched atlas members |
|---|---|---|---|
| `inv_01M2B8ZNKFASFT9RARSDT5TY3R` | `failed approaches share preserves(identity_carried_solvability)` | L4 | Mordell (es-01), Vaughan (es-10) |
| `inv_01M2B8ZNKFASFT9RARSDYY197J` | `failed approaches share preserves(confined_to_quadratic_nonresidues)` | L1 | Mordell, Covering (es-02), Factorization (es-03), Vaughan |
| `inv_01M2B8ZNKFASFT9RARSG8X978R` | `failed approaches share preserves(class_union_construction)` | L3 | Covering, Vaughan |

Corpus fixture: `.newf/m7/newf.db` (materialised by `run.sh` at HEAD `bee1fcc`).

**Explicit target selection:** L3 (`class_union_construction`) and L4
(`identity_carried_solvability`). Not L1 — attacking L1 requires either
using an existing atlas partial-success member's mechanism verbatim (which
fails the novelty gate) or making a claim about QR-class representation
construction that goes beyond what this proposal defensibly supports.

---

## Proposal statement

**A rational-point construction on the ES projective surface via
Brauer–Manin obstruction analysis and 2-descent.**

Let `S_n` be the projective closure of the affine surface
`4xyz = n(xy + yz + zx)` (the ES existence surface for `4/n = 1/x + 1/y + 1/z`).
`S_n` is a smooth projective surface over `Q` for `n` outside a finite set
of degenerate values; its geometric structure has been studied in
connection with exceptional-set bounds for ES (Elsholtz–Tao 2013).

**Mechanism:** for a target `n`, compute the Brauer group `Br(S_n)` and its
image in `∏_v Br(S_n × Q_v)` under the local invariant maps. When the
Brauer–Manin obstruction to the Hasse principle vanishes for `S_n`, apply
2-descent (in the classical sense, following Cassels–Guy and
Colliot-Thélène for cubic surfaces adapted to the specific structure of
`S_n`) to produce a rational point on `S_n(Q)`, yielding an explicit
`(x,y,z)` representation. Otherwise, the obstruction identifies the
specific arithmetic-cohomological reason `n` resists representation and
localises the failure to a computable invariant.

**Why this attacks L3 (`class_union_construction`):** the construction
proceeds by variety-theoretic descent on a single surface `S_n`, not by
covering `n` with congruence classes modulo small primes. The output is a
rational-point construction on one algebraic surface, not a union of
solution sets across arithmetic progressions.

**Why this attacks L4 (`identity_carried_solvability`):** no polynomial
identity is invoked. The construction proceeds by cohomological
computation (Brauer group of the surface, local invariant maps) and
descent-based effective search on a computable algebraic-geometric
object. Solvability is established, when it can be, by the vanishing of a
cohomological obstruction — not by writing down a solvability identity.

---

## Nearest known mechanism family (honest survey)

| Atlas member | Relationship |
|---|---|
| es-09 (higher-dimensional variety lift, partial_success) | **Closest.** Lifts ES to higher-dimensional varieties for parameterization unification. Different operator (unification, not obstruction analysis) and different goal (surface distinctions between parameterizations, not rational-point construction). |
| es-06 (two-fraction representation bounds, partial_success) | Elsholtz–Tao's exceptional-set bound uses surface-theoretic input. Different goal (upper-bound exceptional set, not construct representations) and different technique (character-sum + geometric averaging, not Brauer group + descent). |
| es-11 (Monks–Velingker structural analysis, partial_success) | Structural analysis of the solution set as an algebraic object. Different operator (constraint analysis of an existing solution set, not construction of new points via cohomological obstruction). |

**Not in the atlas:** any use of the Brauer group of `S_n`, any 2-descent
computation on `S_n`, any Brauer–Manin obstruction analysis for ES
representation existence. Confirmed by the atlas survey run at HEAD
`bee1fcc` (see `/tmp/atlas_survey.py`).

**External literature anchor:**

- Colliot-Thélène, Sansuc, and Swinnerton-Dyer's work on Brauer–Manin
  for cubic surfaces provides the general framework. `S_n` is degree-3 in
  `(x,y,z)` and the machinery applies with modifications.
- Cassels–Guy's 2-descent methods for cubic surfaces adapt directly to
  `S_n` in principle.
- Elsholtz–Tao 2013 uses `S_n` for exceptional-set bounds, so the surface
  is known and its geometric structure is documented.

**Genuine mechanistic distinctness, not surface novelty:**

The proposal introduces two operators (`brauer_group_computation`,
`variety_descent`) and one auxiliary object (`projective_es_surface_S_n`)
that no atlas member and no near-neighbour uses for representation
construction. This is mechanistic content in AGENTS.md's sense
(§Mechanistic novelty over surface novelty).

---

## Cheapest falsification path

Four steps ordered by cost. Each step's negative would falsify or weaken
the proposal at a distinct level of specificity.

**Step 1 — Brauer group of `S_n` is trivial (or trivially non-trivial) for
QR-class `n`.** Literature/computation check, hours. If `Br(S_n)/Br(Q) = 0`
for all `n` (i.e. the obstruction vanishes trivially), then the proposal's
"identify the specific arithmetic-cohomological reason" claim collapses
to "there is no obstruction to identify" — the mechanism has no
discriminating content. If instead `Br(S_n)/Br(Q)` is large and
uncomputable, the operator `brauer_group_computation` is not effective as
stated and the proposal weakens to a non-constructive existence claim.

**Step 2 — Brauer–Manin obstruction on `S_n` for small QR-class `n`.**
Explicit computation of the local invariant maps at small primes for
`n ∈ {1, 25, 49, 121, ...}` (small QR values that resist Mordell). Days.
If the obstruction is trivial for all these `n`, the surface has plenty
of local-global points — but Elsholtz–Tao already show ES has solutions
for these `n` (they're just not covered by Mordell), so no new content.
If the obstruction is non-trivial, the proposal's "Brauer–Manin picks out
the arithmetic reason for confinement" claim is directly tested.

**Step 3 — 2-descent effectiveness on `S_n`.** For a specific target `n`
where Brauer–Manin does NOT obstruct, attempt an explicit 2-descent
(compute `Sel_2(S_n)`, produce a rational point via descent). Weeks. If
the descent fails to produce a point despite no Brauer obstruction, the
proposal weakens — either the surface admits an obstruction outside the
Brauer group, or the descent apparatus is not effective for `S_n`. Either
way the mechanism does not deliver as stated.

**Step 4 — verify against ES's known partial-success regime.** For `n`
where Mordell already produces representations, check whether the
proposed apparatus recovers them (or an alternative representation on
`S_n`). If not, the proposal weakens to "cannot even recover the easy
cases via descent."

**Cost hierarchy:** step 1 is a literature check + explicit `Br(S_n)`
computation for one small `n`. Step 2 is `SageMath`-level algebra on
several small `n`. Step 3 is a research-level computation (weeks). Step
4 is small-case verification (days).

A negative at step 1 falsifies the mechanism-effectiveness claim. Steps
2–4 progressively weaken; only step 3's positive result would earn a
full evaluation.

---

## Expected information gain

**High** — the proposal targets 2 of 3 surviving invariants (L3, L4) with
a mechanism family absent from both the atlas and the atlas's
near-neighbours. Whether the proposal succeeds or fails, the outcome
narrows the search space:

- If step 1 falsifies: `brauer_group_computation` is non-effective on ES's
  specific surface — a mechanistic-distance data point that eliminates a
  family of proposals.
- If step 2 obstructs: identifies a specific arithmetic-cohomological
  reason for QR-confinement, potentially strengthening L1's causal core.
- If step 3 succeeds: a new construction mechanism for ES representations
  that does not preserve L3 or L4 — genuine progress on the frontier.
- If step 4 fails: the mechanism is not competitive with Mordell on the
  easy cases, so its value is limited to the hard cases (QR classes) where
  it hasn't been tested.

## Evaluation cost

**Medium** — steps 1–2 are hours to days; steps 3–4 are weeks. A full
evaluation is research-level effort. Preflight and pipeline admission are
seconds.

---

## Provenance

- **Atlas source:** `corpus/train/es-01, es-02, es-03, es-06, es-09,
  es-10, es-11` at HEAD `bee1fcc`. Survey via `/tmp/atlas_survey.py`.
- **Reference-property enrichment source:** M7 recipe
  `corpus/experiments/m7-blinded-run/run.sh` lines 45–65 (the 8 L-claim
  interpretation lines applied during pipeline build).
- **Live invariant targets:** `.newf/m7/newf.db` at HEAD `bee1fcc`, three
  surviving invariants under `mechanism/v3` (IDs listed in the target
  table above).
- **External literature (paper-space citations, not verified via the
  pipeline):** Colliot-Thélène, Sansuc, Swinnerton-Dyer (Brauer–Manin
  framework); Cassels-Guy (2-descent on cubic surfaces);
  Elsholtz-Tao 2013 (surface `S_n` for exceptional-set bounds).
- **Category (per §0 of CLOSURE-SCORECARD):** this predeclaration is
  paper-space; the wire authorship + pipeline run that follows is
  pipeline-space.

---

## Verification tier

- Atlas survey outcomes: SGO (verbatim from newf-normalize JSON blocks in
  the atlas files).
- Target invariant IDs and predicates: SGO (from `newf --json invariant
  list --state surviving` at run time in the prior loop).
- Mechanism-family absence claim ("no atlas member uses Brauer group +
  2-descent for representation construction"): SGO on the atlas survey +
  PE on the taxonomy interpretation. An independent operator could
  disagree about whether es-09's "algebraic lift" is close enough to
  Brauer-Manin descent to count as a near-neighbour rather than a
  distinct family; my judgment is that lifting-for-unification and
  Brauer-Manin-for-existence are different operators.
- External literature citations: paper-space, not verified through the
  pipeline. The `bree_group_computation` operator's effectiveness on
  `S_n` specifically is PE — I have not computed `Br(S_n)` for any `n`.
- The proposal is a **frontier direction candidate**, not a solution.
  Its admission to the pipeline (below) tests whether the code-owned
  admission boundary can recognise mechanistic novelty on this axis; it
  does not test whether the direction actually works.

Same same-model-family caveat as parent records — not independent human
verification.

---

## Next step

Author the `proposal-wire/v1` JSON payload, preflight via
`newf experiment validate-proposals`, and run through
`newf frontier generate --proposals-file` against
`prb_01M2B8ZMHRCJN4YVH4CP6SNXTD` in `.newf/m7/newf.db`.

Read the code-owned violation verdict (`violates_any_target`,
per-target `verdict`, `admission_corrected` count) as evidence. Expected
outcomes and their interpretations are stated in
`records/pipeline-space-vs-paper-space.md` (§What the M7 fixture *does*
enable).
