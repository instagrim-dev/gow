# Pilot 004 — discovery pilot result

Status: `negative_inconclusive`
Recorded: 2026-09-11
Commits: quorum adjudication `78e73d2`; result record corrections `0aac153`
Review: operator audit of `78e73d2` — no repository changes; corrections applied here.

---

## Predeclared verdict

The success criteria were fixed in `PROTOCOL-DRAFT.md` before any capture was
dispatched. All three criteria must be evaluated against the frozen run-level
endpoints, not a pooled tally.

### Criterion 1 — D1 run-level recovery of reference properties

> D1 produces ≥2 of {L1, L3, L4} as `matches_reference` in a majority of its
> runs (≥2 of 3), including at least one of {L3, L4}.

**FAILS.**

Reference-match distribution by run, after unblinding:

```
D1 run 1:  L1              (E17)
D1 run 2:  L1 + L3         (E01, E19)
D1 run 3:  L1              (E02)

D0 run 1:  L1              (E18)
D0 run 2:  L1              (E09)
D0 run 3:  L1              (E12)
```

One of three D1 runs passes the within-run threshold (d1-run2, which
produced both L1 and L3). The frozen criterion requires two of three.
**Not met.**

Numerical correction from the pooled table: the final aggregate contains
**six L1 assignments and one L3 assignment** (E01, E02, E09, E12, E17, E18
→ L1; E19 → L3). The minority rationale for E19 cited L4
(`identity_carried_solvability`) but was outvoted by L3. Crediting E19 with
both L3 and L4 would still leave the informative recovery concentrated in
one D1 run and would not satisfy the majority criterion. L4 was not
recovered in any run by majority.

Narrow favorable observation: L3 (`class_union_construction`) appeared once
in D1 run 2 and in no D0 run. That is the only unnamed-reference-property
hit distinguishing the two arms. It occurred in one run alongside one
disputed over-merge (E10) in the same run, and does not override the
majority-of-runs criterion. It is worth retaining as an observation.

### Criterion 2 — D1 error discipline

> D1's aggregate `over_merge` + `prohibited_promotion` count is ≤ D0's.

**FAILS.**

| Arm | `over_merge` | `prohibited_promotion` | Total |
|---|---|---|---|
| D1 | 2 (E10, E22 — both conservative fallbacks) | 0 | **2** |
| D0 | 0 | 0 | **0** |

Both D1 over-merge assignments are **disputed** (no majority; conservative
fallback applied per the predeclared rule). They should not be presented as
confirmed examples of over-merging; the conservative disposition is a
holding position, not a verdict that most reviewers found an error. The
correct description is: two D1 entries received a conservative fallback of
`over_merge` because no lane majority existed. D0 received no fallbacks.
The criterion still fails under this reading.

Both D1 disputed entries (E10, E22) targeted the same membership set
[es-01, es-03, es-07] and the same structural question as reference entry
L6: whether es-01's modular decomposition and es-03's local congruence
filter can be jointly labelled without erasing the distinct-role finding. One
lane voted `defensible_novel`, one `matches_reference (L6)`, one
`over_merge` for both entries — three irreconcilable readings of what "novel
vs. already known vs. category-violating" means for the same membership set.
That is a genuine scope ambiguity, not a resolved error.

### Criterion 3 — D0 novel rate below D1 novel rate

> D0 does not match the unnamed properties at D1's rate.

**PASSES.**

| Arm | `defensible_novel` occurrences | Total | Rate |
|---|---|---|---|
| D1 | 6 | 12 | 50% |
| D0 | 4 | 11 | 36% |

One of three criteria met.

### Overall verdict

**Verdict: negative / inconclusive.**

The two-criterion failure is decisive. Per the frozen protocol: "Anything
else is a negative or inconclusive result to be recorded as such."

---

## What "negative/inconclusive" names precisely

The verdict is specifically about **run-to-run consistency of rediscovery**.
It does not assess whether the individual entries are valid candidate
properties, whether the proposal operation has any value, or whether the
approach can succeed with design corrections.

The quorum result provides considerably more diagnostic information than a
single pass/fail verdict. The remaining sections distinguish what the data
does and does not establish.

---

## Unblinded arm breakdown

### D1 (guided — train-only bundle, L1/L3/L4 supplied as surviving invariants)

| Entry | Category | Ref | Disputed | Run | Vote split |
|---|---|---|---|---|---|
| E01 | `matches_reference` | L1 | — | d1-run2 | 3-0 |
| E02 | `matches_reference` | L1 | — | d1-run3 | 3-0 |
| E03 | `defensible_novel` | — | — | d1-run3 | 2-1 |
| E05 | `defensible_novel` | — | — | d1-run1 | 3-0 |
| E07 | `defensible_novel` | — | — | d1-run2 | 3-0 |
| E08 | `defensible_novel` | — | — | d1-run1 | 2-1 |
| E10 | `over_merge` | — | ⚠ conservative | d1-run2 | 0 majority |
| E15 | `defensible_novel` | — | — | d1-run1 | 3-0 |
| E17 | `matches_reference` | L1 | — | d1-run1 | 3-0 |
| E19 | `matches_reference` | L3 | — | d1-run2 | 3-0 (L4 in 1 minority reason) |
| E20 | `defensible_novel` | — | — | d1-run3 | 3-0 |
| E22 | `over_merge` | — | ⚠ conservative | d1-run3 | 0 majority |

D1 summary: 4 `matches_reference`, 6 `defensible_novel`, 2 `over_merge` (conservative).

### D0 (control — outcome-permuted bundle; original prose retained)

| Entry | Category | Ref | Disputed | Run | Vote split |
|---|---|---|---|---|---|
| E04 | `unsupported` | — | ⚠ conservative | d0-run2 | 0 majority |
| E06 | `unsupported` | — | — | d0-run1 | 3-0 |
| E09 | `matches_reference` | L1 | — | d0-run2 | 2-1 |
| E11 | `unsupported` | — | — | d0-run1 | 2-1 |
| E12 | `matches_reference` | L1 | — | d0-run3 | 2-1 |
| E13 | `defensible_novel` | — | — | d0-run3 | 3-0 |
| E14 | `unsupported` | — | — | d0-run3 | 2-1 |
| E16 | `defensible_novel` | — | — | d0-run1 | 3-0 |
| E18 | `matches_reference` | L1 | — | d0-run1 | 2-1 |
| E21 | `defensible_novel` | — | — | d0-run2 | 3-0 |
| E23 | `defensible_novel` | — | — | d0-run2 | 3-0 |

D0 summary: 3 `matches_reference`, 4 `defensible_novel`, 4 `unsupported`.

---

## Vote-distribution summary

| Distribution | Count | Entries |
|---|---|---|
| Unanimous 3-0 | 13 | E01, E02, E05, E06, E07, E13, E15, E16, E17, E19, E20, E21, E23 |
| Majority 2-1 | 7 | E03, E08, E09, E11, E12, E14, E18 |
| No majority (conservative fallback) | 3 | E04, E10, E22 |

Seven entries carry a dissenting vote whose rationale names a substantive
objection — see §Testable disagreements.

---

## Novel occurrences — recurring themes, not ten distinct discoveries

Ten entries received `defensible_novel`. **These are proposal occurrences,
not ten distinct properties.** The lane rationales identify substantial
repetition:

- **E05 and E07** are two framings of the same [es-05, es-07, es-10]
  grouping ("intrinsically bounded evidence" and "quantitative measure
  cannot reach every-n"). Lane A flags this explicitly.
- **E13, E16, E20, E23** are four framings of the same [es-05, es-10]
  pairing at different abstraction levels (capped coverage profile,
  asymptotic thinning, too-weak quantitative statement, monotone-but-rate-
  limited). Lane A identifies these as duplicate novel claims.
- **E08, E10, E22** share the membership [es-01, es-03, es-07]; E08 was
  accepted as novel while E10 and E22 received conservative `over_merge`
  fallbacks for the same underlying locality-grouping question.

The defensible statement is:

> **Ten proposal occurrences received a `defensible_novel` disposition
> relative to the reference ledger. They contain recurring candidate themes
> whose distinct-property count has not been adjudicated.**

Approximate distinct themes in the novel set:

| Theme | Occurrences | Notes |
|---|---|---|
| Finite resource against demonstrated-infinite survivor | E03 (D1) | Contested by Lane B — see §Testable disagreements |
| Intrinsic quantitative ceiling / bounded evidence | E05, E07 (D1), E13, E16, E20, E23 (D0) | ~2 distinct properties across 6 occurrences |
| Local-only reasoning without global coupling | E08, E15 (D1) | E10, E22 disputed versions of same theme |
| Reorganisation without added existence | E21 (D0) | Distinct; es-09/es-11 as partial successes |

Recurrence across fresh runs is informative: it suggests the proposer
repeatedly finds a theme rather than generating arbitrary variation. This
answers a different question from distinctness — "does the model converge
on something?" not "how many independent things did it find?"

"Novel relative to this ledger" does not establish novelty in mathematical
literature or prove the curated pass overlooked valid invariants.

---

## Substantive concerns in specific entries

### E15 — annotation pattern versus mechanism-level claim

E15 was accepted unanimously. Its core claim:

> "Several stalled approaches break nothing about the underlying
> congruence/residue structure — their normalized mechanism records list no
> broken properties."

The supporting evidence is three quotations of `"breaks": []` from the
corpus records. That evidence directly supports:

> The corpus annotations record no broken properties for these mechanisms.

It does not directly support:

> The mechanisms break no relevant properties.

The second claim requires either that the annotations are exhaustive, or an
independent substantive argument. This is the same absence-versus-evidence
distinction corrected in the automatic comparator (classify/v2). A unanimous
quorum verdict should not reintroduce it through prose.

Additionally, E15's contrast check lists es-05 among partial successes that
"break something." The frozen corpus records es-05 as `partial_failure`, not
`partial_success`. This is a concrete status mismatch in a unanimously
accepted entry.

**Disposition:** Retain E15's quorum verdict as the historical record. Reopen
the substantive claim for two corrections before any pipeline admission: (1)
narrow the statement to the annotation pattern rather than the mechanism-level
conservation claim; (2) correct the es-05 contrast-check status. Any admitted
revision must be separately attributed.

### E10/E22 — conservative fallback ≠ confirmed error

As noted in §Criterion 2: the three lanes voted `defensible_novel`,
`matches_reference (L6)`, and `over_merge` respectively for both E10 and E22.
Two lanes found the claims supportable; they disagreed about whether novelty
or reference-match was the right category. One lane found an over-merge. The
conservative fallback correctly held the claim to a higher standard, but it
does not establish that most reviewers found a category error.

**Disposition:** Withhold from downstream guidance (conservative action). Do
not treat as confirmed examples of over-merging (unsupported conclusion).

### E19 — L3 match with an L4 claim in the minority rationale

E19 maps to L3 by 3-0 majority. Lane C's rationale argues the claim is
actually L4 (`identity_carried_solvability`) at L3's scope. The mapping
question — class-union construction, identity-carried solvability, or a
conjunction — is a bounded substantive question that the vote count alone
cannot resolve. E19's matched scope and its matched property meaning should be
addressed separately if it is advanced to challenge.

---

## D0 status objections as experimental artefacts

Several D0 entries were rejected partly because they cited outcome classes that
conflict with the original corpus. Some of these conflicts arise from the
experimental intervention, not from model error.

The D0 permutation record shows that d0-run1 (which produced E11) received
these outcome-class flips relative to the original:

| Note | Original class | D0 supplied class |
|---|---|---|
| es-09 | `partial_success` | `partial_failure` |
| es-11 | `partial_success` | `partial_failure` |
| es-03 | `partial_failure` | `partial_success` |
| es-07 | `partial_failure` | `partial_success` |

Lane B rejected E11 partly because "es-09 and es-11 are corpus-recorded as
partial_success" — but in the material d0-run1 actually received, they were
supplied as `partial_failure`. The rejection is grounded in the reference
corpus; the proposer operated on the permuted input.

A proposal can fail the reference-matching criterion while correctly using the
manipulated metadata it was given. **D0 status mismatches should not be
uniformly treated as examples of model hallucination or bad admission
behaviour.** The admission system should not encode "correctly following an
intentionally corrupted input is an error" without distinguishing which input
was authoritative.

This applies specifically to D0 entries. D1 entries received the unperturbed
training bundle; status mismatches in D1 entries are not explained by the
permutation.

---

## D0 reference-matching — prose-retention confirmation

D0 recovered L1 three times (E09, E12, E18) despite outcome-permuted
annotations. This is consistent with the pre-adjudication caution in
`records/INTERPRETATION-NOTES.md`:

> The model follows the original prose: outcome permutation has little effect.

The D0 bundle preserved the original training note text; only the outcome
metadata was permuted. The QR-confinement property is recoverable from prose
alone, independently of outcome labels. D0's three L1 recoveries are expected
under the prose-following explanation.

The same caution applies to D1's L1 recoveries, which is why the L3 recovery
in d1-run2 is the more informative signal: L3 requires reasoning across notes
(es-02 and es-10) rather than recovering a property named in a single note's
prose.

---

## Testable disagreements

The dissent identifies specific bounded questions for follow-up. These are
not noise.

**E03 (D1, 2-1: defensible_novel vs unsupported):** Lane B distinguishes a
finite computation's inability to certify a universal statement (es-07's
boundary: "finite verification is not a proof for all n") from a structural
obstruction leaving a proven-infinite survivor set (es-02, es-10). If these
are different structural situations, E03's shared_by set is not coherent. The
2-vote majority should answer this objection with source evidence, not rely on
the vote count.

**E08/E10/E22 (D1, same membership [es-01, es-03, es-07]):** E08 was
accepted as novel (2-1: defensible_novel); E10 and E22 received conservative
over_merge fallbacks for proposals that are substantively similar. The question
is whether grouping es-01 and es-03 under a shared locality label preserves or
erases their L6-recorded distinct roles. Lane A accepted it in E08 but Lane C
rejected it in E10 and E22; Lane A also treated all three as equivalent.
Whether meaningful wording differences justify the asymmetric outcome is a
bounded question with a bounded answer.

**E19 (D1, 3-0 L3 with L4 in minority reasoning):** Is the claim
"solvability inherited from per-class identities" (L4) at the [es-02, es-10]
scope (L3), or is it a genuinely conjunctive property? This affects whether
L4 was marginally recovered in the pilot.

**E15 (D1, 3-0 defensible_novel but substantive scope concern):** Is the
correct claim about annotation patterns or about mechanism-level conservation?
The evidence base, stated as direct corpus inspection, only supports the
narrower claim. Resolving this requires no new evidence — only clarifying what
the passages actually establish.

---

## Quorum governance note

The three quorum lanes were dispatched to the same model family (claude /
cursor-agent) with role-differentiated prompts. Agreements across lanes
represent within-family consistency under different framings, not
independent-source validation. Different role labels do not on their own
establish three independent sources of expertise.

The quorum is a decision procedure that correctly implements the aggregation
rule and usefully organises judgment. It does not convert judgment into
verification. This is why the project's evidentiary standard requires
challenge to follow adjudication before a hypothesis is established —
quorum consensus is still model judgment under the verification hierarchy.

---

## Provenance limitations

**Key seal gap:** `adjudication-ledger.json` had `key_sealed: null` and
`key_sha256: null` when adjudication began. The key was produced and committed
before quorum dispatch but its SHA256 was not formally recorded in the ledger.
The adjudication itself was blind by construction — quorum agents operated from
`quorum-lane-brief.json` which contains no arm assignments — but tamper-evidence
between key file and ledger was absent. Git commit history and sealed capture
digests provide partial protection. Future protocols should include the key
SHA256 in FREEZE.md before any capture is dispatched.

**Same-model-family quorum:** See §Quorum governance note.

**D0 causal ambiguity:** Per `records/INTERPRETATION-NOTES.md`, a D1 > D0
score would have been consistent with both failure-structure discovery and
sensitivity to coherent-vs-corrupted descriptions. The negative verdict makes
the causal question moot for now but records the limitation for future design.

---

## Learnings for future protocols

**L1 — Run-level credit:** One of three D1 runs passed Criterion 1. A future
protocol could record run-level pass/fail as a secondary metric without
changing the majority criterion.

**L2 — L6 erasure failure mode:** Both D1 over-merge fallbacks grouped es-01
and es-03 under a shared label despite L6's distinct-role finding. Future D1
prompts should include the reference distinctions as explicit negative
examples.

**L3 — L4 absence:** `identity_carried_solvability` was not recovered in any
D1 run by majority. Its cross-note character (operator in es-01, preserved
property in es-02, assumption in es-10) may make it harder to compress into a
single-session proposal. Three runs per arm is insufficient to characterise
absence.

**L4 — Annotation-pattern versus mechanism-level claims:** The E15 finding
shows that proposals derived from structured fields (empty lists, metadata
counts) need to state their evidential scope precisely. A future admission
checklist should distinguish "corpus annotation records X" from "mechanism
exhibits X."

**L5 — D0 permutation disambiguation:** Rejection rationales for D0 entries
should distinguish "entry conflicts with original corpus" from "entry
correctly follows supplied permuted input." The current rejection language
conflates them.

---

## Summary

Pilot-004 produced 23 candidate-property occurrences across six captures.
Quorum adjudication assigned seven reference matches and ten proposal
occurrences with a `defensible_novel` disposition, with recurring themes and
unresolved scope questions across the ten. D1 recovered an unnamed reference
property (L3) in one of three runs, below the predeclared majority-of-runs
requirement; the overall success criterion was not met. The result demonstrates
candidate-generation capability and yields useful hypotheses and diagnostic
disagreements. It does not establish ten distinct discoveries, reliable
discrimination of failure structure from the control, or verification by
reviewer consensus.

**Operational milestone reached:** non-predetermined discovery proposals were
captured, attributed, and subjected to structured criticism.

**Predeclared discovery-success milestone not reached.**

---

## Next steps (operator-owned)

1. **Closure scorecard:** Freeze the endpoint result; group repeated
   hypotheses without deleting occurrences; distinguish reference agreement from
   fidelity to each arm's supplied evidence; resolve specific scope objections
   before promoting candidates into downstream guidance. See
   `records/CLOSURE-SCORECARD.md`.

2. **Challenge step for novel candidates:** For each candidate theme selected
   by the operator, apply the predeclared challenge discipline (known
   counterexample, synthetic counterexample, success-preserving check). A
   surviving challenged property may be admitted as an interpretation claim.
   The negative pilot verdict does not gate this step; individual entry validity
   is independent of pilot consistency.

3. **Scope corrections before any admission:** E15 requires narrowing before
   pipeline admission (annotation pattern vs. mechanism-level claim; es-05
   contrast status). E10/E22 require substantive resolution before any
   downstream guidance use (do not rely on conservative fallback category).

4. **Follow-on protocol revisions:** Incorporate learnings L1–L5 before
   designing the next pilot.

---

*Operator review scope: committed quorum result at `78e73d2`, adjudication key,
frozen protocol, and selected supporting records. Arm/run counts derived
mechanically from the key; no new blinded adjudication or mathematical
verification. Corrections applied to this file at `0aac153` and this revision.*
