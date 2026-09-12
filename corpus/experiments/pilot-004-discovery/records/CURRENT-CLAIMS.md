# Current claims — authoritative view with supersession links

Recorded: 2026-09-12
Provenance: remediates the propagation finding in
[`docs/reviews/2026-09-12-structural-semantic-epistemic-review.md`](../../../../docs/reviews/2026-09-12-structural-semantic-epistemic-review.md)
(§Epistemic findings: "current summaries should surface reassessed strength
directly with supersession links").

## Reading rule

This file is the **entry point** for the current evidential status of every
Pilot-004 claim whose original record was later reassessed, corrected, or
superseded. Historical records are retained unchanged (per the
source-faithfulness review's own instruction); each carries a banner pointing
here. **When a summary elsewhere disagrees with a row below, this file and
the linked reassessment govern.** All statuses below are paper-space
(markdown/CMA-tier) unless a persisted ID is shown; see
[`pipeline-space-vs-paper-space.md`](pipeline-space-vs-paper-space.md).

Epistemic-tier legend (from the source-faithfulness review):
SGO source-grounded observation · PE proposed explanation · CCS completed
counterexample search · IIP inapplicable/inconclusive probe · CMA checked
mathematical argument.

## Claim table

| Claim | Current status (authoritative) | Superseded / reassessed by | Original (historical) record |
|---|---|---|---|
| **N2a** — intrinsic density/averaging ceiling (es-05, es-10) | Admitted as `density_averaging_ceiling` (mechanism/v4) at **model-judgment strength for this corpus**. Seven-part challenge **recorded**; **refinement proposed** (rate-ceiling vs reach-ceiling split); **coverage incomplete** — C2 reclassified IIP+PE (no completed synthetic search; "reconfirms" withdrawn), C7 reclassified SGO+PE (causal insufficiency not established at CMA). Not "survived challenge" in the unqualified sense the original verdict line states. Three mathematical glosses corrected (average↛pointwise; exceptional-set bound grows, its *density* decays). Paper-space only; no live SQLite persistence. | [`N2a-challenge.md` §Attributed reassessment](N2a-challenge.md#attributed-reassessment-2026-09-11-source-faithfulness-review) (R-1…R-4); vocabulary seed comment corrected same date (`internal/canon/vocabulary_seed.go`) | [`N2a-challenge.md`](N2a-challenge.md) (verdict line "survives challenge with one refinement required" is the superseded strength) |
| **N5** — reorganisation without added existence (es-09, es-11) | Admitted as `reorganisation_without_qr_existence` (mechanism/v5) at **model-judgment strength for this corpus**. Probe coverage **six-of-seven**: C3 (success-preserving) is structurally inapplicable on a corpus with no `success` outcomes — an IIP, not a passed probe; it re-activates as an open obligation when a success enters the corpus. C2 reclassified PE (definitional scoping, not a completed search). The operator-type causal implication ("reorganisation *cannot* add existence") is **not established**; only the two in-scope self-descriptions (SGO) plus a PE. Paper-space only. | [`N5-challenge.md` §Attributed reassessment](N5-challenge.md#attributed-reassessment-2026-09-11-source-faithfulness-review) (R-1…R-3) | [`N5-challenge.md`](N5-challenge.md) |
| **N2a-child-1** — obstruction independence of the two ceilings | **`challenged`, not surviving.** Two-reading split (boundary delta): the *local* reading (the two corpus ceilings are distinct obstructions) survives at n=2; the *general* reading (no joint-crossing mechanism can exist) was weakened at C6 (n=1 corpus mechanism per obstruction; joint-crossing untested at authoring time). Partial rehabilitation 2026-09-12: the general reading gains exactly one attempted-and-**falsified** construction (joint-crossing, below) — moving it from "untested" to "one attempt falsified," which does not restore `surviving`. | [`N2a-child-1-challenge.md` §Disposition](N2a-child-1-challenge.md); rehabilitation note in [`CLOSURE-SCORECARD.md`](CLOSURE-SCORECARD.md) | [`N2a-child-1-challenge.md`](N2a-child-1-challenge.md) |
| **N2a-child-2** — non-degenerate reciprocity-carrier requirement | **`weakened`, not surviving, not falsified.** Two weakening landings: C6 ("or equivalent object" is a movable goalpost; n=1 falsified attempt) and C7 (near-tautology: QR-survivors are *defined* by reciprocity). One non-discriminating SGO support (C3/es-06). Sharper enumerated form **N2a-child-2′** proposed, **not admitted**. | [`frontier/N2a-child-2-challenge.md`](frontier/N2a-child-2-challenge.md) | same record (proposal §) |
| **Joint-crossing proposal** — δ-method × composition-genus covering, against N2a's C4 boundary delta | **`falsified_by_cheapest_path`** at step 2 (CMA-tier algebra): the named form `Q(u,v) = uv + λn(u+v)` has discriminant 1 → one class → zero reciprocity-indexed generic characters; Component B is reciprocity-trivial as stated. Separately, the same wire run through the live M7 corpus persisted `fpr_01M2B919QNAWFFM5A31KKD2YB1` with code-owned verdict `unknown` — a different observation answering a different question (N2a is not persisted in M7). | [`frontier/joint-crossing-cheapest-path-execution.md`](frontier/joint-crossing-cheapest-path-execution.md) | [`frontier/N2a-joint-crossing-proposal.md`](frontier/N2a-joint-crossing-proposal.md) |
| **E15** — structural `breaks=[]` conservation claim (Theme N4) | Quorum verdict (`defensible_novel`, 3-0) retained **as historical record only**. Substantive claim reopened and corrected: narrowed to an annotation-pattern claim (**E15′**), es-05 contrast status fixed, and a third defect recorded (atlas `breaks` field does not discriminate outcome class). E15′ was then challenged and landed **`weakened`** — not admitted as surviving. | [`E15-correction.md`](E15-correction.md) (three defects) → [`challenges/E15-prime-challenge.md` §Disposition](challenges/E15-prime-challenge.md) | Ledger entry E15 ([`adjudication-ledger.json`](../adjudication-ledger.json) lines 546–583) |

## What supersession means here

- **Nothing is withdrawn silently.** Every original record is intact,
  including its original (now-superseded) verdict lines; the reassessments
  are attributed and dated appendices within the same files.
- **Strength moves down, provenance stays.** The reassessments reclassify
  evidential character (e.g. CCS→IIP, CMA→PE); they do not delete candidates.
  Both N2a and N5 remain admitted `GeneratedInterpretation` vocabulary
  entries; what is superseded is the advertised strength of their survival.
- **Downstream consumers.** The manuscript already reads the corrected
  strength ("refinement proposed with coverage incomplete, not survival" —
  propagated across five locations at `635e6d8`); the vocabulary seed
  comments were corrected 2026-09-11. Any future summary should cite this
  file rather than the original verdict lines.

## Maintenance rule

When a new reassessment, correction, or supersession is appended to any
Pilot-004 claim record: (1) add or update its row here the same day;
(2) add a banner at the top of the affected record pointing here;
(3) never edit the historical text being superseded.
