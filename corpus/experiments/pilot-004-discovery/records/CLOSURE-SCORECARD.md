# Pilot 004 — closure scorecard

Recorded: 2026-09-11
Based on: `records/RESULT.md` (this revision)

This scorecard groups proposal occurrences into distinct themes, separates
reference agreement from arm-evidence fidelity, and records the scope
objections that must be resolved before any candidate is admitted to the
pipeline. It does not re-adjudicate entries; it organises the existing record
for downstream use.

---

## 1. Frozen endpoint (predeclared success criteria)

| Criterion | Result |
|---|---|
| C1: D1 recovers ≥2 of {L1,L3,L4} in ≥2/3 runs, including L3 or L4 | ❌ 1/3 runs |
| C2: D1 over_merge + prohibited_promotion ≤ D0 | ❌ 2 vs 0 (conservative fallbacks) |
| C3: D0 novel rate < D1 novel rate | ✓ 36% vs 50% |

**Predeclared verdict: negative / inconclusive.**

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

## 5. Admission checklist for challenge step

For any candidate theme the operator selects for challenge:

| Gate | Requirement |
|---|---|
| Evidence bar | Confirmed by quorum `defensible_novel` (passed for N1–N5 entries, subject to scope notes) |
| Scope correction | E15: required; E03/N2b: es-07 inclusion must be justified; E10/E22: not applicable until conservative fallback resolved |
| Challenge: known counterexample | Attempt to find a known failed approach in the corpus that violates the property |
| Challenge: synthetic counterexample | Construct a synthetic failed approach that violates it |
| Challenge: success-preserving | Identify a success that still preserves the property |
| Admission | Operator attestation; new pinned vocabulary revision; reclustering + fresh database before support counted |

The negative pilot verdict does not gate the challenge step. Individual entry
validity is independent of the pilot's run-to-run consistency result.

---

## 6. Open questions (bounded, resolvable)

| ID | Question | Evidence needed |
|---|---|---|
| OQ1 | Is es-07's "finite verification" failure the same structural situation as es-02/es-10's proven-infinite survivor set? | Source passages from the frozen es-07 note |
| OQ2 | Does E08's locality grouping preserve or erase the L6 distinct-role finding? | Statement of what predicate the grouping asserts vs. what L6 preserves |
| OQ3 | Is E19's reference match L3, L4, or a conjunction? | Analyse whether the claim requires both class-union and identity-inheritance, or only one |
| OQ4 | What is the correct scope of E15's claim? | Distinguish annotation-pattern observation from mechanism-level conservation |

These are bounded research questions. They do not require a new harness phase.

---

## 7. Overall scorecard

| Dimension | Result |
|---|---|
| Candidate generation | Demonstrated — 23 occurrences captured, attributed, adjudicated |
| Predeclared success criterion | Not met (negative/inconclusive) |
| Reference rediscovery consistency | 1/3 D1 runs; L3 once (D1 only); L4 not recovered |
| Novel distinct themes | ~5 (N1–N5), from 10 occurrences |
| Entries requiring scope correction before admission | 2 (E15, E03/N2b es-07 inclusion) |
| Entries requiring conservative-fallback resolution | 2 (E10, E22) |
| Open bounded questions | 4 (OQ1–OQ4) |
| Same-model-family quorum | Confirmed — within-family consistency, not independent validation |
| Key seal gap | Confirmed — post-adjudication population; git history provides partial protection |
