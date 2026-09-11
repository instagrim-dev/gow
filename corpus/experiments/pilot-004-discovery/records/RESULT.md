# Pilot 004 — discovery pilot result

Status: `negative_inconclusive`
Recorded: 2026-09-11
Commits: quorum adjudication `78e73d2`; unblinding key `captures/adjudication-key.json`

---

## Predeclared verdict

The success criteria were fixed in `PROTOCOL-DRAFT.md` before any capture was
dispatched. All three criteria must be evaluated exactly as written; the result
cannot be adjusted post-hoc.

### Criterion 1 — Recovery of reference properties (D1 consistency)

> D1 produces ≥2 of {L1, L3, L4} as `matches_reference` in a majority of its
> runs (≥2 of 3), including at least one of {L3, L4}.

**FAILS.**

| Run | Entries | {L1,L3,L4} hits | Includes L3 or L4? | Passes? |
|---|---|---|---|---|
| d1-run1 | E05, E08, E15, E17 | E17 → L1 (1 hit) | No | ✗ |
| d1-run2 | E01, E07, E10, E19 | E01 → L1, E19 → L3 (2 hits) | Yes | ✓ |
| d1-run3 | E02, E03, E20, E22 | E02 → L1 (1 hit) | No | ✗ |

Result: 1 of 3 runs passes. Criterion requires ≥2. **Not met.**

L4 (`identity_carried_solvability`) was not recovered in any D1 run. L3
(`class_union_construction`) appeared in exactly one run (d1-run2, entry E19)
alongside two over-merges in the same run (E10 DISPUTED, E22 DISPUTED). The
one-run signal is preserved as an observation but does not satisfy the majority
test.

### Criterion 2 — Error discipline (D1 OM + PP ≤ D0 OM + PP)

> D1's aggregate `over_merge` + `prohibited_promotion` count is ≤ D0's.

**FAILS.**

| Arm | over_merge | prohibited_promotion | Total |
|---|---|---|---|
| D1 (12 entries) | 2 (E10, E22 — both DISPUTED, cemented conservative) | 0 | **2** |
| D0 (11 entries) | 0 | 0 | **0** |

D1 generated two over-merges; D0 generated none. **Not met.**

Both D1 over-merges targeted the same membership set [es-01, es-03, es-07]
and erased the distinction recorded in reference entry L6 (es-01's modular
decomposition versus es-03's local congruence filter have affirmatively
distinct roles). This is a recurring prompt-design failure: the D1 bundle
contains the training notes but provides no explicit instruction preventing
the L6 distinction from being collapsed. See §Learnings.

Note on disputed cements: the predeclared aggregation rule cements the most
conservative vote when no majority exists. If the operator independently
reverses E10 and E22 to `defensible_novel`, D1 OM drops to 0, matching D0 —
but this would require new evidence, not a recount. The conservative cements
are final under the current record.

### Criterion 3 — D0 novel rate below D1 novel rate

> D0 does not match the unnamed properties at D1's rate.

**PASSES.**

| Arm | defensible_novel | Total entries | Rate |
|---|---|---|---|
| D1 | 6 (E03, E05, E07, E08, E15, E20) | 12 | **50%** |
| D0 | 4 (E13, E16, E21, E23) | 11 | **36%** |

D0's novel rate (36%) is below D1's (50%). One criterion satisfied, two
not.

### Overall verdict

> Anything else is a negative or inconclusive result to be recorded as such.

**Verdict: negative / inconclusive.**

The two-criterion failure is decisive. The partial signal in d1-run2 is
preserved as an observation. No result stronger than "inconclusive" is
warranted.

---

## Unblinded arm breakdown

### D1 (guided — train-only bundle, L1/L3/L4 as surviving invariants supplied)

| Entry | Category | Ref match | Disputed | Run |
|---|---|---|---|---|
| E01 | `matches_reference` | L1 | — | d1-run2 |
| E02 | `matches_reference` | L1 | — | d1-run3 |
| E03 | `defensible_novel` | — | — | d1-run3 |
| E05 | `defensible_novel` | — | — | d1-run1 |
| E07 | `defensible_novel` | — | — | d1-run2 |
| E08 | `defensible_novel` | — | — | d1-run1 |
| E10 | `over_merge` | — | ⚠ conservative | d1-run2 |
| E15 | `defensible_novel` | — | — | d1-run1 |
| E17 | `matches_reference` | L1 | — | d1-run1 |
| E19 | `matches_reference` | L3 | — | d1-run2 |
| E20 | `defensible_novel` | — | — | d1-run3 |
| E22 | `over_merge` | — | ⚠ conservative | d1-run3 |

D1 summary: 4 `matches_reference` (all L1 or L3; L4 absent), 6 `defensible_novel`,
2 `over_merge`.

### D0 (control — outcome-permuted bundle; original prose retained)

| Entry | Category | Ref match | Disputed | Run |
|---|---|---|---|---|
| E04 | `unsupported` | — | ⚠ conservative | d0-run2 |
| E06 | `unsupported` | — | — | d0-run1 |
| E09 | `matches_reference` | L1 | — | d0-run2 |
| E11 | `unsupported` | — | — | d0-run1 |
| E12 | `matches_reference` | L1 | — | d0-run3 |
| E13 | `defensible_novel` | — | — | d0-run3 |
| E14 | `unsupported` | — | — | d0-run3 |
| E16 | `defensible_novel` | — | — | d0-run1 |
| E18 | `matches_reference` | L1 | — | d0-run1 |
| E21 | `defensible_novel` | — | — | d0-run2 |
| E23 | `defensible_novel` | — | — | d0-run2 |

D0 summary: 3 `matches_reference` (all L1), 4 `defensible_novel`, 4 `unsupported`.

---

## Observations (not result claims)

### D0 reference-matching explained by prose retention

D0 recovered L1 (QR-confinement) three times despite outcome-permuted
annotations. This is consistent with the pre-adjudication caution in
`records/INTERPRETATION-NOTES.md`:

> The model follows the original prose: outcome permutation has little effect.

The D0 bundle preserves the original training note text; the permuted field is
the outcome metadata annotation. A model reading the prose can reconstruct
the QR-confinement property without access to a guided bundle. Therefore D0's
three L1 recoveries are expected under the "prose-following" explanation and do
not constitute evidence of failure-structure discovery in the D0 arm.

This observation also limits what D1's L1 recoveries show: the two-explanation
problem (failure-structure discovery vs. prose-following) applies to both arms.
D1's additional L3 recovery in run2 is the only case where the prose-following
explanation is weaker, since L3 (`class_union_construction`) is a
relationship between es-02 and es-10 that requires reasoning across notes
rather than copying a single note's language.

### d1-run2 as the strongest single-run signal

d1-run2 is the only run that passes Criterion 1 individually. It produced:
- E01 → L1 (QR confinement)
- E19 → L3 (class union construction) ← includes L3 ✓
- E07 → defensible_novel (intrinsic quantitative ceiling)
- E10 → over_merge DISPUTED (local-coupling erasure of L6 distinction)

The L3 recovery is the highest-quality signal in the pilot: it requires
recognising that es-02 (covering) and es-10 (density assembly) share a
union-construction property despite different objectives. It does not appear
in any other run. One run is weak evidence; it motivates a follow-on but
does not discharge the majority criterion.

### Ten novel candidates — ready for challenge

The quorum judged 10 entries `defensible_novel`. A `defensible_novel` verdict
means the passages support the property, distinctions are preserved, and the
contrast check is honest. This is a proposed candidate invariant — the model
has done the compression step the research loop expects.

The negative pilot verdict is about run-to-run *rediscovery reliability*, not
about the validity of individual entries. The admission gate for a candidate
invariant is: (1) surviving the evidence bar, and (2) surviving challenge.
Condition 1 is met by the quorum verdict. Condition 2 has not yet been tested.

One asymmetry between arms: the 4 D0 novel entries have the prose-retention
ambiguity (the model may be following original outcome language rather than
discovering structure). The 6 D1 novel entries are less ambiguous in origin.
This does not invalidate D0 entries — it qualifies what their source establishes.

**D1 novel (6):**

| Entry | Property (summary) | Scope |
|---|---|---|
| E03 | Finite resource against demonstrated-infinite survivor set | es-02, es-07, es-10 |
| E05 | Intrinsically bounded evidence type cannot upgrade to universal existence | es-05, es-07, es-10 |
| E07 | Quantitative progress measure cannot reach every-n | es-05, es-07, es-10 (same as E05, tighter framing) |
| E08 | Local-only reasoning, no global coupling | es-01, es-03, es-07 |
| E15 | breaks=[] for all three (conserves structure) | es-01, es-03, es-10 |
| E20 | Global quantitative statement too weak to force per-prime existence | es-05, es-10 |

**D0 novel (4):**

| Entry | Property (summary) | Scope |
|---|---|---|
| E13 | Intrinsically capped almost-all coverage profile | es-05, es-10 |
| E16 | Asymptotic thinning incapable of reaching empty exceptional set | es-05, es-10 |
| E21 | Reorganisation without new existence (es-09, es-11 partial successes) | es-09, es-11 |
| E23 | Monotonically improving but rate-limited, infinite uncovered residue | es-05, es-10 |

Note: E13, E16, E20, E23 are four framings of the same [es-05, es-10]
pairing at different abstraction levels. E05 and E07 are two framings of
the same [es-05, es-07, es-10] grouping. Collapsing these would reduce 10
novel candidates to approximately 4–5 distinct properties. The next step
for these entries is the challenge step (can an adversarial case falsify
them?), not re-adjudication. No admission decision has been made; that is
operator work.

---

## Provenance limitations

### Key seal gap

`adjudication-ledger.json` has `key_sealed: null` and `key_sha256: null`.
The adjudication key (`captures/adjudication-key.json`) was produced by the
executing agent and committed before the quorum was dispatched. It was not
formally hashed into the ledger before adjudication began. The adjudication
itself was blind by construction — the quorum agents operated from
`quorum-lane-brief.json` which contains no arm assignments — but the
tamper-evidence seal between the key file and the ledger was not completed.

**Limitation:** a post-hoc key substitution could not be detected from the
ledger record alone. The git commit history and the sealed capture digests
provide partial protection, but the formal seal is absent. Future protocols
should include the key SHA256 in FREEZE.md before any capture is dispatched.

### Same-model-family limitation

All three quorum lanes were dispatched to the same model family (claude,
cursor-agent). Lane diversity was role-based (research / engineering /
generalist), not model-diverse. Agreements across lanes therefore represent
within-family consistency rather than independent-source agreement. The
aggregation rule treats majority as cement; this caveat applies to all
cemented results.

### D0 control interpretation

Per `records/INTERPRETATION-NOTES.md`, the D0 design is "unperturbed notes
versus outcome-permuted notes whose prose still retains the original outcome
information." A D1 > D0 score — had it been achieved — would have been
consistent with both failure-structure discovery and sensitivity to
coherent-versus-corrupted descriptions. The negative verdict makes this
moot for now but records the limitation for future protocol design.

---

## Learnings for future protocols

**L1 — Run-level credit:** One of three D1 runs passed Criterion 1. A
future protocol could record run-level pass/fail as a secondary metric
without changing the majority criterion. This preserves the majority bar
while making the partial signal visible.

**L2 — L6 erasure failure mode:** Both D1 over-merges targeted [es-01,
es-03, es-07] and erased the L6 distinction. The D1 prompt did not
explicitly instruct the model to preserve distinctions recorded in the
reference. Future prompts should include the reference distinctions (not
only the reference properties) as negative examples: "Do not group es-01
and es-03 under a shared label without preserving their
decomposition-vs-filtering distinction."

**L3 — L4 absence:** The `identity_carried_solvability` property (L4) was
not recovered in any D1 run. The three D1 runs collectively covered all
twelve training notes through their proposals; the property exists in the
training evidence. Its absence may reflect the prompt's framing, the
property's subtlety (it appears as operator/preserved/assumption across
three notes rather than as a single named pattern), or run-count limits.
Three runs per arm is insufficient to characterise absence.

**L4 — Novel candidates as scoping input:** The 10 novel entries, if
subsequently challenged and admitted, would expand the property space
available to guide future experiments. The four [es-05, es-10] framings
suggest a density/averaging ceiling property may be admissible under a
tighter scope; this would be L7 or similar if admitted.

---

## Next steps (operator-owned)

The following are decisions, not tasks. None may be executed by the agent
without explicit operator instruction.

1. **Admission and challenge (D-C1):** Review the arm-labeled novel candidates
   above and decide which to advance to the challenge step. Challenge means:
   attempt to find a known failed approach that violates the property, a
   synthetic failed approach that violates it, and a success that preserves it
   (per AGENTS.md §Candidate invariant discipline). Properties that survive
   challenge can be admitted as interpretation claims in a new vocabulary
   revision. A positive challenge outcome is the correct gate — not the pilot
   overall verdict.

2. **Ledger update (D-B1):** After this result record is committed, the
   executing agent may update `adjudication-ledger.json` to replace `SEALED`
   capture references with actual arm/run assignments from the key, and
   populate `key_sha256` from the file digest.

3. **Follow-on protocol:** If a follow-on discovery pilot is planned,
   incorporate L1–L4 above before freezing the next design.

---

*Review scope: quorum votes in `quorum-result.json`, adjudication key in
`captures/adjudication-key.json`, predeclared criteria in `PROTOCOL-DRAFT.md`,
and pre-adjudication interpretation notes in `records/INTERPRETATION-NOTES.md`.
Verdict is mechanical from the predeclared criteria; no new judgment was
applied to reach `negative_inconclusive`.*
