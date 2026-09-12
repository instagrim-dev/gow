# E15 correction record — scope narrowing, es-05 classification fix, and contrast overreach

**Original entry:** E15 (D1, unanimous 3-0 defensible_novel) —
[`adjudication-ledger.json`](../../adjudication-ledger.json) at lines 546–583.

**Original claim (verbatim from ledger entry_id E15):**

> "Several stalled approaches break nothing about the underlying
> congruence/residue structure — their normalized mechanism records list no
> broken properties — consistent with the obstruction note's claim that
> every failed family preserves the quadratic-residue wall."

**Disposition instruction (from
[`RESULT.md`](../RESULT.md) §E15, lines 241–245):**

> "Retain E15's quorum verdict as the historical record. Reopen the
> substantive claim for two corrections before any pipeline admission: (1)
> narrow the statement to the annotation pattern rather than the mechanism-
> level conservation claim; (2) correct the es-05 contrast-check status. Any
> admitted revision must be separately attributed."

Recorded: 2026-09-12. This record supplies both required corrections and
documents a **third defect** found by CMA-verifying E15's contrast_check
against the frozen atlas. The three defects are separated so an admission
gate can evaluate each independently. Nothing here promotes E15's quorum
verdict beyond its `defensible_novel` status; the verdict stays as
historical record per the disposition instruction.

Epistemic-tier legend (inherited from parent records):

```text
SGO  Source-grounded observation (verifiable against the frozen corpus)
PE   Proposed explanation (plausible argument; not independently verified)
CMA  Checked mathematical argument (steps individually verified)
```

---

## Corpus verification (SGO — all citations against the frozen atlas)

The frozen atlas notes were checked directly:

| Note | `outcome.class` | `breaks` |
|---|---|---|
| `corpus/train/es-01-mordell-polynomial-identities.md` | `partial_failure` (line 48) | `[]` (line 40) |
| `corpus/train/es-02-covering-system-attempt.md` | `failure` (line 46) | `["class-local isolation"]` (line 38) |
| `corpus/train/es-03-factorization-scheme.md` | `partial_failure` (line 47) | `[]` (line 39) |
| `corpus/train/es-04-type-i-ii-classification.md` | `partial_success` (line 44) | `["classification by producing congruence"]` (line 36) |
| `corpus/train/es-05-averaging-bombieri-vinogradov.md` | **`partial_failure`** (line 44) | `["per-n constructivity"]` (line 36) |
| `corpus/train/es-06-two-fraction-representation-bounds.md` | **`partial_success`** (line 43) | **`[]`** (line 35) |
| `corpus/train/es-09-higher-dimensional-variety-lift.md` | `partial_success` (line 45) | `["surface distinctions between parameterizations"]` (line 37) |
| `corpus/train/es-10-vaughan-congruence-density.md` | `partial_failure` (line 44) | `[]` (line 36) |
| `corpus/train/es-11-monks-velingker-structure.md` | **`partial_success`** (line 43) | **`[]`** (line 35) |
| `corpus/train/es-12-type-a-b-congruence-system.md` | `partial_success` (line 46) | `["classification by producing congruence"]` (line 38) |

All rows verified against the atlas commits at
[`corpus/train/`](../../../../corpus/train/) as of this record. Bold cells
are the ones the defects below turn on.

---

## Three defects in the original E15 entry

### Defect 1 — Scope overreach in the main claim (OQ4-flagged)

**Discrepancy:** The claim as stated is *mechanism-level* — it asserts
that the shared_by mechanisms "break nothing about the underlying
congruence/residue structure." The supporting evidence is three
`"breaks": []` annotations. The evidence supports only the *annotation-
level* observation ("the frozen records list no broken properties");
the *mechanism-level* claim ("the mechanisms break no relevant property
in fact") requires either exhaustive annotation coverage or an
independent substantive argument, neither of which is supplied.

**Source:** [`RESULT.md`](../RESULT.md) §E15, lines 222–234. This is the
same absence-versus-evidence distinction that classify/v2 was hardened
against.

**Correction required:** narrow the claim to the annotation pattern.

### Defect 2 — es-05 classification error in contrast_check (OQ4-flagged)

**Original contrast_check quote (from ledger):**

> "The partial successes each record a non-empty break — es-04 and es-12
> break 'classification by producing congruence', **es-05** breaks
> per-n constructivity, es-09 breaks surface distinctions between
> parameterizations — so breaking some inherited structural commitment
> co-occurs with partial success in this corpus, [...]"

**Discrepancy:** es-05 is annotated `partial_failure`, not `partial_success`
(atlas line 44). Its `breaks` field content is quoted correctly, but its
outcome_class categorisation is wrong.

**Source:** [`RESULT.md`](../RESULT.md) §E15, lines 236–239.

**Correction required:** remove es-05 from the "partial successes" list
in the contrast_check.

### Defect 3 — Contrast_check overreach (newly found via CMA verification)

**Discrepancy:** the contrast_check asserts "the partial successes
**each** record a non-empty break." Verified against the frozen atlas,
this is false: **es-06** (partial_success, atlas line 43) and **es-11**
(partial_success, atlas line 43) both record `breaks: []`. The
substantive shape of the contrast — that non-empty break co-occurs with
partial success — is not attested by the atlas.

**Not flagged by OQ4.** OQ4's resolution called out only the es-05
categorisation error; the deeper problem that the underlying correlation
fails within the corpus was not noted. This record surfaces it now.

**Impact:** correcting only defects 1 and 2 would leave the substantive
contrast intact, and that substantive contrast is itself non-corpus-
consistent. Correcting all three is necessary before any admission of a
revised E15.

**Correction required:** replace the "each records a non-empty break"
quantifier with an accurate summary of the actual partial_success break
patterns (mixed: three with non-empty breaks; two with `breaks: []`).

---

## Revised, admissible E15 (separately attributed)

Statement (revised):

> **E15' (proposed 2026-09-12; scope: annotation-pattern observation):**
>
> The corpus annotations for **es-01** (Mordell polynomial identities),
> **es-03** (factorization scheme), and **es-10** (Vaughan congruence
> density) each record no broken properties — their frozen mechanism
> records carry `"breaks": []`. This is a syntactic pattern in the
> annotation records, not a mechanism-level conservation claim: no
> independent argument is offered that these mechanisms break no
> relevant property *in fact*.

Contrast (revised — CMA-verified against the atlas):

> The frozen atlas does not exhibit a "partial_success ⇔ non-empty
> break" correlation. Of the six atlas members annotated
> `partial_success`:
>
> - **es-04, es-09, es-12** record non-empty breaks
>   (`"classification by producing congruence"`; `"surface distinctions
>   between parameterizations"`; `"classification by producing
>   congruence"` respectively);
> - **es-06, es-11** record `"breaks": []`.
>
> es-05 is `partial_failure` (not `partial_success`) and records
> `"per-n constructivity"` as its break. The annotation-pattern claim
> above is therefore an observation about three specific
> `partial_failure` records with empty breaks; it is neither
> a mechanism-level conservation claim, nor a claim that discriminates
> partial_success from partial_failure via the breaks field.

Falsification (revised):

> An admissible falsification is: an update to the frozen es-01, es-03,
> or es-10 records that supplies a non-empty `breaks` field. The
> revised claim does not extend to mechanism-level behaviour, so a
> mechanism-level counterexample (a mechanism from the shared_by set
> that "in fact" breaks something) is *not* a falsification of E15' as
> revised — it may inform a future, stronger claim, but E15' explicitly
> stays at annotation scope.

Attribution:

- Original quorum verdict on E15: unanimous 3-0 defensible_novel,
  historical record retained per RESULT.md §E15.
- **E15' revised claim:** separately attributed to this correction
  record (2026-09-12). Not admitted; enters `proposed`. Any admission
  requires operator attestation per AGENTS.md §Candidate invariant
  discipline.

---

## Admission gate for E15' (proposed decision rule)

Under the pipeline's admission discipline (
[`docs/invariant-challenge.md`](../../../../docs/invariant-challenge.md)
§Admission; AGENTS.md §Candidate invariant discipline), E15' admission
requires:

1. **Scope resolution:** the three OQ4-flagged corrections plus the
   defect-3 correction must all be applied to the revised claim
   (**satisfied here**).
2. **Challenge campaign:** the seven probes (C1–C7) must be run against
   E15' before admission. Not run here.
3. **Operator attestation:** operator review of the corrections against
   the frozen atlas. Not performed here.
4. **Vocabulary revision:** if admitted, a new pinned vocabulary
   revision + reclustering under it, before support counts.

Step 1 is complete. Steps 2–4 remain outstanding. This record moves E15
from "correction outstanding" to "corrected, ready for challenge."

---

## Notes on the N4 theme

E15 is the sole entry in Theme N4 (structural `breaks=[]` pattern) per
[`CLOSURE-SCORECARD.md`](../CLOSURE-SCORECARD.md) §3 → N4. With E15'
now the admissible-shape statement, Theme N4's admission gate advances
from "correction outstanding" to "ready for the seven-probe challenge."

The N4 theme's coherence is somewhat weakened by defect 3: the frozen
atlas's `breaks` field does not correlate with outcome_class. This is
a **structural observation about the atlas's annotation completeness**,
not about the mechanisms themselves. If the challenge campaign for E15'
runs, C6 (sampling bias) should incorporate this observation directly:
the annotation-pattern claim's inductive base (three `breaks: []`
records among `partial_failure` outcomes) is bounded, and the field's
overall discrimination is weak (breaks-null occurs in both outcome
classes).

---

## Search-policy implications (proposed, not applied here)

1. **Do not use E15's original mechanism-level claim as guidance.**
   Redundant with the falsified defects 1 and 3.
2. **Do consider E15' for the challenge queue.** Corrected and
   attributed; its scope is honest.
3. **Redundancy warning for future annotation-pattern claims:** any
   future proposal that draws mechanism-level conclusions from atlas
   annotation-field patterns should be gated against E15's original
   defect-1 shape. The atlas's annotations are not exhaustive; absence
   of a field entry is not evidence of the mechanism's actual
   behaviour.
4. **Recommendation for atlas hygiene** (out-of-scope for this record):
   an atlas revision that either explicitly marks `breaks` fields as
   exhaustive/non-exhaustive per record, or removes empty `breaks`
   entries from partial_success rows to avoid future
   annotation-pattern confusion, would sharpen the corpus. Not
   performed here.

---

## Provenance

- Original entry: `adjudication-ledger.json` E15 (lines 546–583).
- Disposition instruction: `records/RESULT.md` §E15 (lines 212–245).
- OQ4 resolution: `records/CLOSURE-SCORECARD.md` §6 OQ4 (resolved
  2026-09-11).
- Frozen atlas notes: `corpus/train/es-{01,02,03,04,05,06,09,10,11,12}-*.md`
  (line-anchored citations in the verification table above).
- Executor: model-assisted, single-operator recorded. The line-number
  citations against the frozen atlas are CMA (an independent reader can
  reproduce every check with `rg`); the identification of defect 3 as
  a discovered problem (not just an applied OQ4 correction) is SGO
  against the atlas. Same same-model-family caveat as the parent
  records — this is not independent human verification.
