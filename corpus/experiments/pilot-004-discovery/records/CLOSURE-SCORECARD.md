# Pilot 004 — closure scorecard

Recorded: 2026-09-11
Based on: `records/RESULT.md` (this revision)

This scorecard groups proposal occurrences into distinct themes, separates
reference agreement from arm-evidence fidelity, and records the scope
objections that must be resolved before any candidate is admitted to the
pipeline. It does not re-adjudicate entries; it organises the existing record
for downstream use.

---

## 0. Category note — paper-space vs pipeline-space

Every row in this scorecard is one of two structural categories. The
distinction was made explicit at commit
[`66c96a5`](https://github.com/instagrim-dev/gow/commit/66c96a5) after
running the joint-crossing proposal wire through `newf frontier generate`
end-to-end and observing that its target invariants returned
`verdict: unknown` — because those invariants are not persisted in the live
corpus. Full record: `records/pipeline-space-vs-paper-space.md`.

- **Paper-space** — markdown/CMA-tier research reasoning. Durable via git
  history and line-anchored corpus citations. Not consumed by any code-owned
  typed operation. All admission/challenge/frontier-generation rows in
  §§1–7 authored before 2026-09-12 are paper-space by default.
- **Pipeline-space** — code-owned persisted state in a live SQLite corpus
  (invariants, cluster runs, frontier generations, proposals, evaluations,
  admissions). Consumed by the pipeline's typed operators. Currently
  materialised only by `corpus/experiments/m7-blinded-run/run.sh` into
  `.newf/m7/newf.db` (not checked in); Pilot-004-specific pipeline-space
  state does not yet exist.

**Reading rule.** Rows carrying an explicit `[Paper-space only; no live
SQLite persistence]` or `[Paper-space; no code-owned ... row]` annotation
are paper-space (the annotations were retrofitted in `66c96a5`). Rows
carrying persisted IDs like `fpr_...`, `fgr_...`, `inv_...`, or
`clr_...` are pipeline-space. Unannotated rows without persisted IDs are
paper-space by default.

**Both categories are legitimate.** The distinction is honesty, not a
value ranking. Paper-space records ratchet the research chain forward
under AGENTS.md's discipline (typed epistemic status, CMA-verifiable
citations, durable git provenance); pipeline-space records are what
future automated operators consume. Confusion between them is what the
naming prevents.

**Class-move opportunities** (the classes this distinction now names):

| From | To | Move |
|---|---|---|
| Paper-space challenge campaign (e.g. N2a-child-1) | Pipeline-space challenge campaign | Run `newf challenge` against a persisted surviving invariant with a code-owned campaign row |
| Paper-space admission (e.g. N2a `surviving` at mechanism/v4) | Pipeline-space admission | Materialise a corpus at the target vocab and persist the predicate |
| Paper-space frontier proposal | Pipeline-space frontier proposal | Serialise to `proposal-wire/v1` and run through `newf frontier generate --proposals-file` |
| Paper-space evidence claim | Pipeline-space admission | Run `newf evidence admit` under S2's typed-observation-kind rules |

---

## 1. Frozen endpoint (predeclared success criteria)

| Criterion | Result |
|---|---|
| C1: D1 recovers ≥2 of {L1,L3,L4} in ≥2/3 runs, including L3 or L4 | ❌ 1/3 runs |
| C2: D1 over_merge + prohibited_promotion ≤ D0 | ❌ 2 vs 0 (conservative fallbacks) |
| C3 (frozen protocol): "D0 does not match the unnamed properties at D1's rate" | reference-recovery comparison — did **not** favor D0 (L3 recovered in D1 only; L4 not recovered; D0's L1 matches partially explained by prose retention) |

> **Endpoint-definition note (2026-09-11).** The frozen protocol defines C3 in
> terms of *reference-property* recovery ("the unnamed properties"). An earlier
> row here paraphrased C3 as "D0 novel rate < D1 novel rate (36% vs 50%)" — that
> is a *defensible-novel classification* rate, a **different measurement** from
> reference-property recovery. The frozen protocol controls: C3 is a
> reference-recovery comparison. On that measurement D0 did not exceed D1. The
> 36%/50% novel rates are retained below only as a separately labeled
> descriptive observation, not as the C3 endpoint.

**Predeclared verdict: negative / inconclusive.**

Descriptive-only (not the C3 endpoint): defensible-novel classification rate
D0 36% vs D1 50%.

---

## 2. Reference-property occurrences

Seven entries matched a reference property. Grouped by reference ID:

| Reference | Entries | Arm | Unanimous? |
|---|---|---|---|
| L1 — confined_to_quadratic_nonresidues | E01, E02, E17 (D1), E09, E12, E18 (D0) | mixed | yes for E01, E02, E17; 2-1 for E09, E12, E18 |
| L3 — class_union_construction | E19 (D1) | D1 | yes (L4 in 1 minority reason) |
| L4 — identity_carried_solvability | — | — | not recovered by majority |

L1 was reproduced six times across both arms. L3 appeared once, in D1 only.
L4 was not recovered. The L1 recoveries in D0 are partially explained by
prose retention (the original outcome language was present in D0's prose
despite permuted metadata). L3's appearance in D1 only is the pilot's
narrowest favorable signal.

---

## 3. Novel occurrences — grouped into distinct themes

Ten entries received `defensible_novel`. Collapsed into approximate distinct
themes:

### Theme N1 — Finite resource against infinite survivor set

| Entry | Arm | Run | Vote | Status |
|---|---|---|---|---|
| E03 | D1 | d1-run3 | 2-1 (DN, US, DN) | **Contested — open scope question** |

Claim: the mechanism spends a finite resource (finite covering family, finite
verification range, finite class battery) against a survivor set demonstrated
to be infinite. Shared by es-02, es-07, es-10.

Lane B contention: es-07's boundary is "finite verification cannot certify a
universal statement" — a generality failure, not a structural obstruction
against a proven-infinite set. These may not be the same structural situation.

**Before admitting:** answer whether es-07's survivor set is demonstrated
infinite in the same sense as es-02 and es-10, or whether the three entries
share only the nominal "finite vs. all" shape.

---

### Theme N2 — Intrinsic quantitative ceiling

Six occurrences of the same underlying [es-05, es-10] property (4 entries)
and an extension to [es-05, es-07, es-10] (2 entries):

| Entry | Arm | Scope | Vote | Framing |
|---|---|---|---|---|
| E05 | D1 | es-05, es-07, es-10 | 3-0 | "intrinsically bounded evidence cannot upgrade to universal existence" |
| E07 | D1 | es-05, es-07, es-10 | 3-0 | "quantitative progress measure cannot reach every-n" |
| E13 | D0 | es-05, es-10 | 3-0 | "intrinsically capped almost-all coverage profile" |
| E16 | D0 | es-05, es-10 | 3-0 | "asymptotic thinning incapable of driving exceptional set to empty" |
| E20 | D1 | es-05, es-10 | 3-0 | "global quantitative statement intrinsically too weak to force per-prime existence" |
| E23 | D0 | es-05, es-10 | 3-0 | "monotonically improving but rate-limited; infinite uncovered residue" |

Approximate distinct properties:

- **N2a**: [es-05, es-10] — intrinsic ceiling on density/averaging approaches; the
  gap to "all n" is structural, not a shortfall of effort. (E13, E16, E20, E23
  are four framings of this.)
- **N2b**: [es-05, es-07, es-10] — intrinsically bounded evidence type that cannot be
  upgraded to universal existence, extending N2a to include computational
  verification. (E05 and E07 are two framings of this.)

E05/E07 appeared in D1; E13/E16/E23 appeared in D0. E20 appeared in D1.
The property appears in both arms, which is informative about reproducibility
but not arm-discriminating.

N2b (es-07 inclusion) is subject to the same scope question as N1: whether
es-07's boundary (generality failure) is the same structural situation as the
density/averaging intrinsic ceiling.

---

### Theme N3 — Local-only reasoning without global coupling

Three occurrences, two with a parallel contested family:

| Entry | Arm | Vote | Note |
|---|---|---|---|
| E08 | D1 | 2-1 (DN, DN, OM) | accepted; Lane C dissent on L6 distinction |
| E10 | D1 | 0-majority (DN, MR, OM) | conservative `over_merge` — open |
| E22 | D1 | 0-majority (DN, MR, OM) | conservative `over_merge` — open |

E08 was accepted as novel. E10 and E22 were not (conservative fallback).
All three share membership [es-01, es-03, es-07] and make similar locality
claims. The asymmetry may reflect wording differences; it may also reflect
unstable category boundaries at the novelty/reference-match/over-merge
junction. See §Testable disagreement in RESULT.md.

E15 is related: it makes a locality-adjacent claim (breaks=[] for the same
es-01/es-03/es-10 subset) but is in a different theme.

---

### Theme N4 — Structural breaks=[] pattern

| Entry | Arm | Vote | Note |
|---|---|---|---|
| E15 | D1 | 3-0 (DN, DN, DN) | **scope correction required before admission** |

Unanimous. Claim: es-01, es-03, es-10 list no broken properties in their
corpus annotations.

Scope correction required: the evidence (`"breaks": []`) directly supports
"annotations record no broken properties"; it does not directly support
"mechanisms break nothing." See RESULT.md §E15. Additionally, E15's contrast
check lists es-05 as a partial success with a non-empty break; the frozen
corpus records es-05 as `partial_failure`. The contrast statement needs
correction.

**Before admitting:** narrow the claim to the annotation-pattern observation;
correct the es-05 contrast-check status. Revised claim must be separately
attributed.

---

### Theme N5 — Reorganisation without added existence

| Entry | Arm | Vote | Note |
|---|---|---|---|
| E21 | D0 | 3-0 (DN, DN, DN) | scope: es-09, es-11 (both partial_success) |

Unanimous. Claim: es-09 and es-11 deliver reorganisation without new existence
for the survivor classes. Their partial_success status does not contradict the
claim — their successes lie elsewhere (compression/constraint analysis); the
non-existence at QR primes is a separately recorded property.

This is the cleanest novel entry: high vote confidence, no scope objection,
D0 origin (weaker attribution but the property itself is uncontested).

---

## 4. Unsupported entries — key lessons

Four entries received `unsupported` (E04, E06, E11, E14).

**E11 (D0, 2-1)** — Lane B rejected it partly on corpus-status grounds. The
D0 permutation supplied es-09 and es-11 as `partial_failure` and es-03/es-07
as `partial_success`. E11's apparent status errors are partly artefacts of the
experimental intervention, not purely model errors. See RESULT.md
§D0 status objections.

**E06, E14 (D0, 3-0 and 2-1)** — Included partial_success notes in "stalled
approaches" framing. These are genuine misreadings of the unperturbed corpus;
the D0 permutation did not alter es-09, es-11, or es-12 for d0-run1/d0-run3
in ways that would explain these misclassifications.

**E04 (D0, conservative fallback: PP, US, DN)** — Three incompatible readings:
prohibited_promotion, unsupported, defensible_novel. The claim (aggregate/
characterising mode) is acknowledged by all lanes as corpus-wide; the dispute
is whether that breadth makes it prohibited. Conservative fallback: unsupported.

---

## 5. Challenge checklist — what determines individual candidate status

The pilot's aggregate verdict limits claims about *method reliability*. It
does not gate individual candidate validity. A candidate from a negative
pilot can be grounded, challenged, and survive; a candidate from a positive
pilot can be falsified. What determines downstream use is the candidate's
own challenge results, not whether the originating experiment succeeded.

For each candidate theme the operator selects:

| Step | Requirement | Notes |
|---|---|---|
| Scope resolution | Confirm shared_by set is coherent; settle any open scope question (OQ1–OQ4) | Do this *before* challenge, not during |
| Challenge: known counterexample | Find a known failed approach in the corpus that violates the property | Per AGENTS.md §Candidate invariant discipline |
| Challenge: synthetic counterexample | Construct a synthetic failed approach that violates it | — |
| Challenge: success-preserving | Find a success that still preserves the property | — |
| Challenge: abstraction level | Lower and raise abstraction; check if candidate splits or merges | — |
| Challenge: sampling/redundancy | Distinguish correlation from causal obstruction | — |
| Admission | Operator attestation; new pinned vocabulary revision; reclustering + fresh database before support counted | Only after surviving challenge |

**Recommended starting points** (cleanest candidates, no open scope questions):

1. **N5 — Reorganisation without added existence** (E21, D0, 3-0): es-09 and
   es-11 deliver reorganisation at higher structural levels without adding
   existence at the QR survivor classes. Unanimous, no scope objection, and
   the contrast (es-06 adds representation lower bounds, so reorganisation
   alone is not the success path) is clearly stated.

2. **N2a — Intrinsic density/averaging ceiling** (E13, E16, E20, E23, mixed
   arms, 3-0 each): es-05 and es-10 share a structural ceiling — the
   almost-all gap cannot be closed within the method's own machinery. Four
   independent formulations, all unanimous, appearing in both arms. The
   recurrence strengthens the case that the property is real and reproducible.

---

## 6. Open questions (bounded, resolvable)

| ID | Question | Evidence needed | **Resolution** |
|---|---|---|---|
| OQ1 | Is es-07's "finite verification" failure the same structural situation as es-02/es-10's proven-infinite survivor set? | Source passages from the frozen es-07 note | **Resolved (2026-09-11): NO.** es-07's boundary is *epistemological*: a finite computation cannot certify a universal claim regardless of its size. es-02/es-10's boundary is *mathematical*: the QR-survivor set is proven non-empty and infinite by Dirichlet + quadratic reciprocity. These are structurally distinct. N2b (es-07 inclusion) and N1 (es-02/es-10 grouping with es-07) should exclude es-07 or explicitly split the es-07 membership into a separate epistemological-limitation claim. |
| OQ2 | Does E08's locality grouping preserve or erase the L6 distinct-role finding? | Statement of what predicate the grouping asserts vs. what L6 preserves | **Deferred.** Requires reading the L6 predicate and what role distinction it preserves. |
| OQ3 | Is E19's reference match L3, L4, or a conjunction? | Analyse whether the claim requires both class-union and identity-inheritance, or only one | **Resolved (2026-09-11):** E19 describes class-union construction as its primary operator and references identity-carried solvability as an upstream dependency. The quorum correctly assigned `matches_reference`; the match is to L3 (class_union_construction) as the asserted property; L4 (identity_carried_solvability) is a stated prerequisite, not the claim itself. The reference match is L3, not a conjunction, and the L4 reference does not require a separate match. |
| OQ4 | What is the correct scope of E15's claim? | Distinguish annotation-pattern observation from mechanism-level conservation | **Resolved (2026-09-11):** E15's evidential basis is `breaks: []` annotations in three corpus notes. This supports a narrower claim: *"the annotation records no broken properties for these mechanisms."* It does not support a mechanism-level conservation claim ("the mechanisms break no relevant properties"), which would require evidence that the annotations are exhaustive. The correct scope is annotation-pattern observation. E15 also contains a contrast error: it lists es-05 among partial successes, but es-05 is annotated `partial_failure`. These two corrections (narrower claim + contrast fix) are required before E15 can be admitted; no challenge should proceed until they are applied. |

These are bounded research questions. They do not require a new harness phase.

---

## 7. Overall scorecard

| Dimension | Result |
|---|---|
| Candidate generation | Demonstrated — 23 occurrences captured, attributed, adjudicated |
| Predeclared success criterion | Not met (negative/inconclusive) |
| Reference rediscovery consistency | 1/3 D1 runs; L3 once (D1 only); L4 not recovered |
| Novel distinct themes | ~5 (N1–N5), from 10 occurrences |
| N4 admission gate | **Closed 2026-09-12:** E15 correction record applied at commit [`859cbbc`](https://github.com/instagrim-dev/gow/commit/859cbbc). Revised claim **E15'** entered `proposed`, then challenged via seven probes at HEAD `859cbbc`. **Disposition: `weakened`** — 2 weakening landings (C3: es-06 and es-11 preserve the property across outcome_class boundary; C6: support set equals evidence base + field non-discriminating + category tension), 1 merge signal (C5: productive merge into a corpus-hygiene observation), 2 category challenges (C4: three-way arithmetic split reveals atomicity; C7: descriptive-correlational-by-construction claim is not an invariant candidate in AGENTS.md's sense). Boundary-delta: two derived observations, neither admitted — `E15''` (corpus-hygiene: the atlas's breaks field is not consistently discriminating; 5 records carry `breaks: []` across outcome_class), and `E15'''` (atomic three-record listing at HEAD). N4 theme closes as `weakened` — this is not a failure of the pilot; it is the correct epistemic verdict on an annotation-scope observation. Record: `records/challenges/E15-prime-challenge.md`. **[Paper-space; no code-owned challenge campaign row.]** |
| Pipeline-space vs paper-space distinction | **Recorded 2026-09-12** at HEAD `d4596c3`. The M7 recipe (`corpus/experiments/m7-blinded-run/run.sh`) was executed end-to-end against current head, producing a live SQLite fixture at `.newf/m7/newf.db` with 3 surviving invariants (L1/L3/L4-shaped), 12 clusters, and full experiment run. Joint-crossing proposal wire also run through `newf frontier generate --proposals-file` against this live corpus: **persisted proposal `fpr_01M2B919QNAWFFM5A31KKD2YB1`** with `violates_any_target: false, verdict: unknown` (N2a not persisted in M7 corpus; violation cannot be code-verified). Reveals: prior reciprocity-axis chain has been operating in *paper-space* (markdown records), not *pipeline-space* (code-owned persisted state). Both are legitimate; the confusion was implicit and is now explicit. Record: `records/pipeline-space-vs-paper-space.md`. |
| Frontier proposal 2026-09-12 — Brauer–Manin descent (against L3 + L4 in live M7 corpus) | Pipeline-space. Predeclared in `records/frontier/brauer-manin-descent-proposal.md`; wire authored, preflighted (`experiment validate-proposals` ok:true, valid:true, permitted_targets=3); run through `newf frontier generate --proposals-file` at HEAD `bee1fcc`. **Persisted proposal `fpr_01M2BABVW2SHP7R4E1YZ7ETFTE`** (generation `fgr_...66CW`, revision 7), hash `a0c4e06e047f84deaa4f4e82b74952f0eb0cff6a5b2b3448a4bc6fe4acb03c9a`. Mechanism: Brauer group computation + Brauer-Manin obstruction analysis + 2-descent on the projective ES surface `S_n: 4xyz = n(xy+yz+zx)`; external anchors Colliot-Thélène/Cassels-Guy/Elsholtz-Tao. Explicit targets: L3 (`class_union_construction`) + L4 (`identity_carried_solvability`); admission_corrected=0, admission_downgraded=0, admission_stripped=0 (labels resolved cleanly on first pass). **Code-owned verdict:** `violates_any_target: false`, both target verdicts `unknown`. Mechanistic distance: `medium`; nearest clusters: 12 with `unknown` classification. Record: `records/frontier/brauer-manin-descent-execution.md`. |
| Pipeline structural finding 2026-09-12 (H2, CMA-tier) | Two independent wire-authored proposals (`fpr_01M2B919...D2YB1` and `fpr_01M2BABV...TFTE`) with distinct content, distinct labels, distinct `admission_corrected` counts (1 vs 0), both returned `verdict: unknown` on every `Contains(preserves, X)` target. Traced to `internal/canon/admission.go` lines 108-124 (SetFieldCompleteness reset to Unobserved for every untrusted proposal) composed with `internal/invariant/predicate.go` lines 448-456 (Contains returns `VerdictViolates` only when completeness = Complete). Composition: **untrusted wire proposals systematically cannot return `VerdictViolates` on `Contains(preserves, X)` invariants** — a design property implementing AGENTS.md's `ModelJudgment != Verification` at the code boundary. Wire-attackable predicate shapes on the current corpus: none (all 3 surviving invariants are Contains-preserves). Enum-axis (Locality/ConstructionMode/UncertaintyMode) and Not-form predicates would be wire-attackable, but the M7 mining did not produce any. Full trace: `records/frontier/brauer-manin-descent-execution.md` §"The mechanism behind H2". |
| Entries requiring conservative-fallback resolution | 2 (E10, E22) |
| Open bounded questions | 4 (OQ1–OQ4); **OQ1/OQ3/OQ4 resolved 2026-09-11; OQ2 deferred** |
| N2a admitted | Yes — `density_averaging_ceiling` (mechanism/v4); C4 boundary_delta recorded. **[Paper-space only; no live SQLite persistence — see `records/pipeline-space-vs-paper-space.md` (2026-09-12).]** |
| N5 admitted | Yes — `reorganisation_without_qr_existence` (mechanism/v5); C4 boundary_delta recorded. **[Paper-space only; no live SQLite persistence.]** |
| N2a-child-1 challenged 2026-09-12 | State: `challenged` (not surviving). 0 completed decisive negatives; 1 weakening landing at C6 (n=1 corpus mechanism per obstruction; joint-crossing untested). Boundary-delta produced: two-reading split (local reading survives at n=2; general reading weakened by C6). Record: `records/N2a-child-1-challenge.md`. **[Paper-space; no code-owned challenge campaign row.]** |
| Frontier proposal 2026-09-12 (against N2a's C4 boundary-delta) | Recorded: **joint-crossing proposal** — δ-method counting identity composed with composition-genus covering, claiming to break both `rate_subthreshold` (es-05) and `qr_confinement` (es-10) in the same mechanism. State: **`falsified_by_cheapest_path`** at step 2 (2026-09-12): the named form `Q(u,v) = uv + λn(u+v)` has discriminant 1, one class, zero reciprocity-indexed generic characters — Component B's covering apparatus is reciprocity-trivial as stated. CMA-tier algebra. Wire payload preflighted; step 1/3/4 became n/a-given-step-2. Records: `records/frontier/N2a-joint-crossing-proposal.md`, `records/frontier/joint-crossing-cheapest-path-execution.md`. **Pipeline-space update 2026-09-12**: same wire also run end-to-end through `newf frontier generate --proposals-file` against the live M7 corpus at HEAD `d4596c3`, producing persisted proposal `fpr_01M2B919QNAWFFM5A31KKD2YB1` (generation `fgr_...84FMC`, revision 6). Code-owned violation verdict: **`violates_any_target: false` with all targets returning `verdict: unknown`** — the pipeline's violation check could not confirm the claimed structural violation against the M7 corpus's L1/L3/L4-shaped surviving invariants (N2a is not persisted in the M7 corpus). CMA falsification and pipeline `unknown` verdict are separate observations answering different questions. |
| N2a-child-2 proposed 2026-09-12 → **challenged 2026-09-12** | State: **`weakened`** (not surviving, not falsified). Two weakening landings: C6 (n=1 falsified attempt; "or equivalent object" is a movable goalpost) and C7 (causal core is near-tautological — QR-survivors are *defined* by reciprocity, so any covering apparatus must evaluate Legendre-symbol information by definition). One positive support at SGO tier: C3 (es-06 partial success attests the claim's prediction, but non-discriminatingly against N2a). Boundary-delta: sharper form `N2a-child-2'` proposed (enumerates the reciprocity-layer types: forms, L-functions, class groups, modular/theta objects), not admitted. Record: `records/frontier/N2a-child-2-challenge.md`. **[Paper-space; no code-owned challenge campaign row.]** |
| N2a-child-1 general reading (partial rehabilitation) | The C6 weakening on "no joint-crossing mechanism can exist" gains one attempted-and-falsified construction. Net movement: "no attempted construction" → "one attempt falsified at reciprocity-carrier step." Weak evidence for the general reading; does not restore it to `surviving`. |
| Cluster revision under v5 | `clr_01M28W6AF7DBAX5363TVXSN3NR` (12 singletons, clean) |
| **Recommended next action** | **Revised 2026-09-12 (second revision) after pipeline structural finding H2:** path (a) as originally written targeted an unreachable outcome — untrusted wire proposals cannot return `verdict: violated` on `Contains(preserves, X)` invariants under the current pipeline design (traced in `records/frontier/brauer-manin-descent-execution.md`). Four paths in revised priority order: **(a′)** Mine an enum-axis or Not-form invariant on the current M7 failure population (e.g. `equals(locality, local)` for the Mordell family or `not(contains(preserves, class_local_isolation))` for es-02). Enum-axis and Not-form predicates ARE wire-attackable; a survived invariant of such shape becomes the first target on this corpus that a wire proposal can drive to `violated`. This is the highest-leverage path *for the pipeline-space discipline*. **(a″)** Run `newf challenge` directly against the M7 surviving invariants (code-owned campaign path, distinct from wire-authored frontier proposals). Produces code-owned challenge campaign rows against the three L-shape invariants; not gated by the completeness-stripping issue that blocks wire violation verdicts. **(c)** Run `newf evidence admit` against the persisted joint-crossing proposal `fpr_01M2B919...D2YB1` under S2's typed-observation-kind rules. Predicted verdict: `structural-claim-failure` — **never admissible** by rule. Testable end-to-end now. **(b)** Materialise a Pilot-004-specific SQLite fixture with N2a persisted at mechanism/v4 — distinct research campaign, not a recipe run. Separately: OQ2 remains deferred; N5-child-1 remains blocked on predicate formalization; N2b es-07 inclusion correction still outstanding. Corpus-hygiene follow-on from the E15' challenge: atlas's `breaks` field inconsistency remains as an atlas-quality task. |

The negative pilot result limits the claim about the *method's reliability*.
It does not invalidate the candidate invariants. A challenged and surviving
candidate is admissible regardless of whether the pilot that produced it met
its aggregate endpoint.
