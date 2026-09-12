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

**[Superseded 2026-09-12: both starting points were executed — challenge
records exist and were subsequently reassessed. Current strength for both
lives in `records/CURRENT-CLAIMS.md`, which governs over the pre-challenge
characterizations below.]**

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
| Code-owned challenge campaign 2026-09-12 (all 3 M7 invariants) | Pipeline-space. `newf challenge --problem prb_...NXTD --all` at HEAD `269de0d`. Persisted 6 new challenge campaign rows across 3 invariants × 3 probes (`known-counterexample`, `success-preserving`, `bias-critique`). **All 3 invariants remain `surviving`** (bias-critique's `surviving` transition overrides KCE's intermediate `challenged`). All 9 probes return `unconfirmed` — H2 corollary confirmed: completeness-stripping affects code-owned challenges too, not just wire proposals. **Support strength (pipeline-visible):** L1=5 (margin +3), L4=3 (margin +1), L3=2 (**at threshold**). L3 is one cluster-removal away from mining failure. **Protocol delta finding:** the code-owned 3-probe protocol is a proper subset of the paper-space 7-probe protocol (C1-C7); paper-space C2/C5/C6/C7 are absent from code but produced substantive weakenings in prior loops (C5 → E15' hygiene merge; C6/C7 → N2a-child-2 movable-goalpost + near-tautology weakenings). Paper-space challenge runs on invariants that ALSO have pipeline-space state produce *strictly more* information than the pipeline's own campaign — not a translation. Record: `records/challenges/m7-code-owned-challenge-campaign.md`. |
| Evidence-admission attempt 2026-09-12 — H3 finding | Pipeline-space. Evaluated both wire proposals via `newf evaluate`: **both returned `verdict: verification_blocked`** with `verifier_kind: model-judgment`, `verification_strength: single-model-judgment`, "no verifier returned a decisive verdict". Evaluations `evl_01M2BB0BT7VJYK17CZWM2HJPA1` (joint-crossing) and `evl_01M2BB33P6EAJS6K8Z7D3Y0AGH` (Brauer-Manin) persisted. **Admission refused for both** at the "not a recorded evaluated failure" barrier — with AND without `--attest --note` (attestation controls admissibility *within* `evaluated_failures`; it does not add rows to that table). CMA trace: `internal/store/evaluation_store.go` lines 214-220 (R6: only `failure/partial_failure` verdicts enter `evaluated_failures`) composed with `internal/pipeline/admission.go` line 234 (selection refuses non-evaluated-failures before attestation runs). **H3 finding:** wire-authored proposals whose evaluation ends at `verification_blocked` are terminally-stuck at admitted-and-ranked pipeline state; no code-owned pathway advances them further. My prior "predicted verdict: structural-claim-failure = never admissible" was wrong at two levels — classification requires deterministic-check verifier kind (which didn't fire), AND `verification_blocked` doesn't reach classification anyway. Both proposals now confirmed at terminal pipeline state; paper-space CMA falsification remains real research content but is not code-consumable. Record: `records/frontier/evidence-admission-attempt.md`. |
| Enum-axis invariant survey 2026-09-12 — (a‴) design-space | Paper-space design analysis with pipeline-space grounding (SGO + CMA). Before executing (a‴), asked two prior questions: (1) can the miner emit enum/Not shapes? (2) would any reach `recurring` on M7? **Answer 1 (CMA):** the CLI-default `DerivingFixtureInvariantMiner` (`invariant_fixture_derive.go` lines 49-100) is hardcoded to emit only `OpContains(preserves\|operators)`. Cannot emit `OpEquals`/`OpIn`/`OpNot`/`OpBoundary`/`OpAll`/`OpAny`. M7's L1/L3/L4 are all `Contains(preserves, X)` because that is the only shape the CLI miner can produce. **Answer 2 (SGO + CMA):** hand-evaluated candidate predicates against the 7 failure-side + 5 success-side families (verbatim sqlite3 output). **C1 `Equals(locality, local)` reaches `recurring` at support=3 with 100% failure-side specificity (0/5 success-side prevalence).** C4 `Equals(construction, constructive)` support=3, 1/5 success prevalence. C8 `Not(Equals(locality, global))` support=4, 2/5 success prevalence. **Structural insight (CMA):** enum-axis predicates escape H2 because `AdmitProposalSignature` does not touch `Posture`; enum-axis evaluation (predicate.go:495-505) returns decisive `Satisfies`/`Violates` when the axis is known. Wire proposals declare posture explicitly (`untrusted_proposer.go:39-41,389`) so the enum axis is always known. **This is the escape hatch from H2+H3 for wire-authored proposals.** Path (a‴) forks into two code-change routes: (a‴-A) add `--proposals-file` to `newf invariants mine`, or (a‴-B) extend the deriving miner to emit enum-axis proposals. Not executed here; recorded as durable prediction and code-change roadmap. Record: `records/enum-axis-invariant-survey.md`. |
| (a‴-B) executed 2026-09-12 — H2 escape end-to-end | Code-space + pipeline-space (user-authorized). Extended `DerivingFixtureInvariantMiner` with opt-in `EmitPostureAxes` mode (distinct `ModelName` → distinct reuse key; backward-compatible default preserves all 200+ existing tests). Added `--emit-posture-axes` CLI flag. **Mining stage:** re-mining M7 produced rev 2 (`ivr_01M2BCJZBSVP2XSCFNTBHWCJKH`) with 5 candidates including two NEW shapes: `Equals(locality, local)` [support=3, association=`contrast_observed` — **stronger than the existing 3 Contains invariants' `recurring`**] and `Equals(construction, constructive)` [support=3, `contrast_observed`]. **Challenge stage:** `Equals(locality, local)` → **`surviving`** with 4 concrete failure-side counterexample members recorded; `Equals(construction, constructive)` → `weaken` because the challenge correctly identified c6 (constructive/global/deterministic partial_success) as a success-preserving family — exactly the hand-eval prediction. **Frontier stage (HEADLINE):** wire proposal declaring `posture.locality=global` against `inv_01M2BCJZBSVP2XSCFNTNTG8TDV` returned **`verdict: violates`** — **the first decisive violation verdict from any wire-authored proposal in M7 history.** All prior wire proposals against `Contains(preserves, X)` invariants returned `verdict: unknown`. **H2 ESCAPED end-to-end.** **Evaluation stage:** verdict `verification_blocked` — H3 remains closed and is now demonstrably ORTHOGONAL to H2 (both H2-stalled and H2-escaped proposals reach the same H3 barrier). Records: `records/frontier/h2-escape-execution.md` and `records/frontier/h2-escape-test.wire.json`. Persisted proposal: `fpr_01M2BCSS7EJ5WBYNST2J3F0ECY`. |
| (h) atlas-truth check 2026-09-12 — weakening is atlas-correct + axis co-variation surfaced | SGO + CMA. Verified c6's `posture.construction=constructive` against `corpus/train/es-12-type-a-b-congruence-system.md`. Verbatim: *"each form comes with an explicit polynomial, and by construction none of the admitted congruences is a quadratic residue"*; *"constructive-per-form"*. es-12 IS legitimately constructive → the `weaken` disposition on `equals(construction, constructive)` is **atlas-correct**; no atlas remediation needed. **Surface finding from the survey**: on M7, `locality=local` and `construction=constructive` are **perfectly co-varying on the failure side** — the three failure-side constructive families (c0=es-03, c3=es-07, c4=es-01) are EXACTLY the three failure-side local families. `All(loc=local, con=constructive)` reduces to `equals(locality, local)` on this corpus. **Convergence with paper-space:** the three failure-side `local/constructive` families are all local-modulus explicit-construction attempts (Mordell polynomial identities, factorization schemes, computational verification). This is essentially the N2a chain and density-and-averaging finding, re-derived through support/contrast counting on posture axes rather than through CMA on covering-density arithmetic. The mining engine, once authorised to emit enum-axis proposals, reproduces the paper-space discovery via an independent route. Record: `records/m7-atlas-truth-check.md`. |
| (i) mining ↔ CMA convergence 2026-09-12 — pipeline blind-spot closed on 4 persisted ES corpora | SGO on 4 corpora. Ran (a‴-B)-extended miner (`--emit-posture-axes`) against M7, pilot-001, pilot-002, pilot-003 train corpora. **All four produce IDENTICAL enum-axis emissions:** `equals(locality, local)` (support=3, contrast_violating=5/5, `contrast_observed`) + `equals(construction, constructive)` (support=3, contrast_violating=4/5, `contrast_observed`). Same predicate fingerprints; same support/contrast counts; same association status. **HEADLINE finding on pilot-001**: original mining record (`corpus/experiments/pilot-001/records/mining.json`) shows `"candidate_count": 0`, `"candidates": null` — the pre-extension miner told operators there were ZERO invariants discoverable. The extended miner today finds TWO `contrast_observed` invariants (the strongest association state) on the same corpus. **Not a corpus-idiosyncratic finding** — the delta is identical (+2 contrast_observed invariants) across all four corpora. This is a **systematic corpus-independent blind spot in the pre-extension miner**, now closed. Pilots 2 and 3 executed their subsequent challenge and frontier stages against 3 `recurring` invariants without knowing 2 stronger `contrast_observed` invariants existed in the same data. Consequence: the extended miner is a viable retroactive-enrichment step for historical pilots. Convergence hypothesis (paper-space CMA vs pipeline-space mining) **confirmed on the ES domain**. Cross-domain generalization remains open (pilot-005-relational not materialized). Record: `records/i-mining-convergence.md`. |
| (k-restricted) 2026-09-12 — pilot-001's specific STOP gate is now satisfiable | SGO on pilot-001. Bounded retrospective (not a re-execution). Challenged the two new invariants created by (i)'s mining pass on pilot-001. **Result**: `equals(locality, local)` → **surviving** (4 failure-side violations recorded, 5 success-side non-preservers); `equals(construction, constructive)` → `weaken` (success-preserving confirmed, same as M7, atlas-truthful per (h)). **Pilot-001 now has ONE surviving invariant for the first time in its history.** Re-running `experiment readiness` today: the `surviving_invariants` gate — the SPECIFIC gate that stopped pilot-001 historically (`no surviving candidate invariant; B3 has no eligible guided target`) — is now `ready: surviving=1`. A DIFFERENT gate (`recovery_reachability: blocked`) now blocks readiness, but that gate did not exist in pilot-001's original readiness manifest (schema evolution; classify/v1 → classify/v3). **This is the unique case among persisted ES pilots where (a‴-B) resolves the specific historical STOP.** Pilots 2 and 3 both had `surviving_invariants: ready` originally, so their historical dispositions are unchanged by the extension (though their substrates are enriched). Only the substrate-level claim is asserted; no claim is made about what pilot-001's B0/B3 captures would have produced had the extension been available. Record: `records/k-pilot-001-retrospective.md`. |
| (m) 2026-09-12 — recovery_reachability is legitimate hardening, not a bug; all historical pilots regressed | CMA + SGO on 4 corpora. Traced `readiness.go:216-236`: the check does a target self-comparison under `ProfileMechanismV3`. Verbatim from the code comment: *"no proposal can classify better against a target than the target itself, so a non-recovering self-comparison means recovery is unreachable for ANY proposal."* This is a mechanical necessity, not a schema-migration bug. **Cross-corpus diagnosis:** M7 (freshly normalized) target has resolved=11, unresolved=0 → **ready**. Pilots 1, 2, 3 (older schema) targets have resolved=2/2/13, **unresolved=9/9/9** → **blocked** on all three. **All three historical ES pilots' targets fail the current bar.** classify/v3's completeness-aware absence semantics correctly refuse targets with unresolved decisive-field claims. **Definitive regression evidence:** pilot-003's original `readiness.json` recorded `recovery_reachability: ready`; today under classify/v3 the SAME target blocks — proving the check existed then and produced a passing verdict under the older profile. **Design-space observation** (recorded, not proposed): the pipeline uses the CURRENT classification profile for readiness on any corpus, creating a "corpus-versioning ratchet" where historical pilots fall out of admissibility as classify semantics tighten. A **pinned-profile readiness** mode (check against pilot's historical profile, recorded on manifest) would separate "would proceed today under today's rules" from "would proceed today under its historical rules". **Consequence for (k-full)**: full re-execution of any historical pilot requires re-normalization of its target(s) with full decisive-field resolution — an atlas-quality task with 9 unresolved claims to canonicalize per historical target, plus operator authorization. Neither prerequisite is a code change. Record: `records/m-recovery-reachability.md`. |
| (g-scoped) 2026-09-12 — H3 escape investigation: witness path exists, no in-scope obligation matches; H3 gate preserved | CMA + SGO. **Pre-work rescoping**: my predeclared (g-A) was inconsistent with the pipeline's documented design. Verbatim from `internal/verify/deterministic.go`: *"A confirmed violation is deliberately left NON-decisive (unknown) here; 'the invariant is broken' is necessary but not sufficient for success"*; *"It never rewards missing/unknown comparison evidence with partial_success"*; and `evaluation.go:320-322`: the default model tier abstains *"so a bare deployment never launders judgment."* The H3 stall on M7 is not a design gap — it's this design behaving correctly. **Witness path (CMA)**: `internal/witness/witness.go` implements an exact-integer domain-goal check for ES: `4·x·y·z == n·(y·z + x·z + x·y)` over big.Int, with mandatory provenance note, stamped as reproducible-computation strength at subject domain-goal. Refuses malformed tuples and unnoted claims. **SGO on production usage**: zero witness-check evaluations across M7, pilot-001, pilot-002, pilot-003. The pipeline's designed H3 escape has never been exercised on real corpus data. **SGO on candidate obligations**: all persisted target-violating proposals in M7 make structural mechanism-family claims (introduce a global coupling object; degeneracy analysis on survivor classes) with no concrete `(n,x,y,z)` tuple attributable to the proposal's mechanism. Attaching an unrelated tuple would be misattribution — explicitly forbidden by user authorization. **Remaining verification obligations**: for M7 target-violating proposals, person-days of domain mathematics to construct the claimed global coupling object explicitly before any witness tuple becomes derivable. Not a code change; not attestation. Real domain research work. **Design-space observation** (recorded, not proposed): generator↔verifier mismatch — strongest deterministic verifier expects concrete claims, generator produces abstract mechanism-family claims. Search policy could downweight non-concrete falsification paths as a "bounded observation" per authorization ("may inform search priority while remaining non-decisive about the domain goal"). **H3 gate status: preserved.** No admission rules changed, no verifier rewards absence with a positive verdict, blocked evaluation on the H2-escape proposal remains blocked. Record: `records/g-scoped-h3-investigation.md`. |
| (o-closed) 2026-09-12 — no additional ranking directive justified for this iteration (record narrowed post-`12cbaa3` per reviewer) | CMA + SGO, warrant narrowed. **Reviewer correction** on the first version of this record: closing (o) as "no additional ranking directive justified" is defensible; closing it as "concrete falsifiability is already established by the ranker" is not. The record was rewritten to reflect this narrower disposition. **What the CMA actually establishes**: (a) every existing directive TargetKind resolves to a persisted-row identity (KTD-1); a free-form falsification-path text directive would fail resolution. (b) `baseObjectiveLess` orders `MechanisticDistance desc → ExpectedInformationGain desc → EvaluationCost asc → hash`; declared cost is a tiebreaker. (c) `computeFalsifiabilityFloor` selects the cheapest declared-cost violator per target and prevents its net policy bias from becoming negative — it does NOT establish that a check exists, reserve execution budget, or guarantee evaluation. **Property distinction (corrected)**: declared evaluation cost and verifier readiness are different properties. The H2-escape proposal (`fpr_01M2BCSS7EJ5WBYNST2J3F0ECY`) is the counterexample within this pilot's own records — it self-attests `evaluation_cost: low` and ranks first, yet (g-scoped) established it has no candidate-specific witness obligation. **What the M7 rank table shows (corrected)**: consistency with the comparator on persisted ordinal tuples, NOT an isolated demonstration of the cost term (multiple rank-0 rows; earlier ordering dimensions and applied bias not held equal). **Closure wording (reviewer-supplied, adopted verbatim)**: *"Path (o) closed without implementation. Existing ordinal evaluation-cost ranking and per-target anti-suppression protection make an additional cost-priority directive unjustified for this iteration. These mechanisms consume declared cost; they do not establish concrete verifier readiness or witness attribution. The next research input must supply those properties explicitly. H3 remains unchanged, and review-integration obligations retain their separate status."* **Separately open (not discharged by this loop)**: C7, witness-occurrence attribution, C8. Pausing exploratory code-change work is not evidence of end-to-end conformance on those obligations. **H3 gate status: preserved (unchanged).** Record: `records/o-closed-lever-exists.md` (rewritten). |
| Entries requiring conservative-fallback resolution | 2 (E10, E22) |
| Open bounded questions | 4 (OQ1–OQ4); **OQ1/OQ3/OQ4 resolved 2026-09-11; OQ2 deferred** |
| N2a admitted | Yes — `density_averaging_ceiling` (mechanism/v4); C4 boundary_delta recorded. **[Paper-space only; no live SQLite persistence — see `records/pipeline-space-vs-paper-space.md` (2026-09-12).]** **[Current claim (2026-09-12): reassessed strength governs — challenge *recorded*, refinement *proposed*, coverage *incomplete*; see `records/CURRENT-CLAIMS.md`.]** |
| N5 admitted | Yes — `reorganisation_without_qr_existence` (mechanism/v5); C4 boundary_delta recorded. **[Paper-space only; no live SQLite persistence.]** **[Current claim (2026-09-12): six-of-seven probe coverage, C3 inapplicable; see `records/CURRENT-CLAIMS.md`.]** |
| N2a-child-1 challenged 2026-09-12 | State: `challenged` (not surviving). 0 completed decisive negatives; 1 weakening landing at C6 (n=1 corpus mechanism per obstruction; joint-crossing untested). Boundary-delta produced: two-reading split (local reading survives at n=2; general reading weakened by C6). Record: `records/N2a-child-1-challenge.md`. **[Paper-space; no code-owned challenge campaign row.]** |
| Frontier proposal 2026-09-12 (against N2a's C4 boundary-delta) | Recorded: **joint-crossing proposal** — δ-method counting identity composed with composition-genus covering, claiming to break both `rate_subthreshold` (es-05) and `qr_confinement` (es-10) in the same mechanism. State: **`falsified_by_cheapest_path`** at step 2 (2026-09-12): the named form `Q(u,v) = uv + λn(u+v)` has discriminant 1, one class, zero reciprocity-indexed generic characters — Component B's covering apparatus is reciprocity-trivial as stated. CMA-tier algebra. Wire payload preflighted; step 1/3/4 became n/a-given-step-2. Records: `records/frontier/N2a-joint-crossing-proposal.md`, `records/frontier/joint-crossing-cheapest-path-execution.md`. **Pipeline-space update 2026-09-12**: same wire also run end-to-end through `newf frontier generate --proposals-file` against the live M7 corpus at HEAD `d4596c3`, producing persisted proposal `fpr_01M2B919QNAWFFM5A31KKD2YB1` (generation `fgr_...84FMC`, revision 6). Code-owned violation verdict: **`violates_any_target: false` with all targets returning `verdict: unknown`** — the pipeline's violation check could not confirm the claimed structural violation against the M7 corpus's L1/L3/L4-shaped surviving invariants (N2a is not persisted in the M7 corpus). CMA falsification and pipeline `unknown` verdict are separate observations answering different questions. |
| N2a-child-2 proposed 2026-09-12 → **challenged 2026-09-12** | State: **`weakened`** (not surviving, not falsified). Two weakening landings: C6 (n=1 falsified attempt; "or equivalent object" is a movable goalpost) and C7 (causal core is near-tautological — QR-survivors are *defined* by reciprocity, so any covering apparatus must evaluate Legendre-symbol information by definition). One positive support at SGO tier: C3 (es-06 partial success attests the claim's prediction, but non-discriminatingly against N2a). Boundary-delta: sharper form `N2a-child-2'` proposed (enumerates the reciprocity-layer types: forms, L-functions, class groups, modular/theta objects), not admitted. Record: `records/frontier/N2a-child-2-challenge.md`. **[Paper-space; no code-owned challenge campaign row.]** |
| N2a-child-1 general reading (partial rehabilitation) | The C6 weakening on "no joint-crossing mechanism can exist" gains one attempted-and-falsified construction. Net movement: "no attempted construction" → "one attempt falsified at reciprocity-carrier step." Weak evidence for the general reading; does not restore it to `surviving`. |
| Cluster revision under v5 | `clr_01M28W6AF7DBAX5363TVXSN3NR` (12 singletons, clean) |
| **Recommended next action** | **Revised 2026-09-12 (thirteenth revision) after (o-closed) record narrowed per reviewer:** Paths (a″), (c), (a‴-B), (h), (i), (k-restricted), (m), (g-scoped), (o) closed for this iteration. **The existing ordinal evaluation-cost ranking and per-target anti-suppression protection consume declared cost; they do NOT establish concrete verifier readiness or witness attribution. Any next research input intended to close that gap must supply candidate-specific claim form, attributed provenance, and verifier execution explicitly — none of which is a ranking concern.** Remaining priorities are research-execution, corpus-materialization, or design-space observations awaiting downstream authority: **(l)** Cross-domain replication — materialize pilot-005-relational or another non-ES pilot corpus. **(k-full)** Historical pilot re-execution — multi-loop, atlas-quality re-normalization + operator authorization. **(p)** Frontier generator bias toward concrete-witness-producing mechanism families (design-space; needs authorization). **(q)** Authoring-discipline documentation — operator-judgment signal requiring calibration, not a code-owned discriminator. **(n)** Pinned-profile readiness (design-space, unchanged). **(e)** C6 design-space (unchanged). **(b)** Materialise a Pilot-004-specific SQLite fixture. **(j)** Axis co-variation as corpus signal (design-space, unchanged). **(a‴-A)** Lower priority. **Separately open — NOT discharged by this loop**: C7 obligation, witness-occurrence attribution obligation, C8 obligation. Pausing exploratory code-change work is not evidence of end-to-end conformance on those obligations. **The corpus of open code-change paths for this pilot iteration is small; the corpus of open review-integration obligations is not.** OQ2, N5-child-1, N2b es-07, atlas `breaks` field inconsistency remain as atlas-quality tasks. |

The negative pilot result limits the claim about the *method's reliability*.
It does not invalidate the candidate invariants. A challenged and surviving
candidate is admissible regardless of whether the pilot that produced it met
its aggregate endpoint.
