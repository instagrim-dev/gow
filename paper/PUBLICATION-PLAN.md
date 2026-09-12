# Publication Plan — Geometry of Work

**Target:** arXiv flagship paper introducing GoW as a discipline.
**Title:** Geometry of Work: Shape-Guided Search over Failed and Partial Solution Attempts
**Alternate subtitle:** Search in Shape-Space, Verify in Domain-Space
**Format:** LaTeX, single-column, ~15–20 main-text pages + appendices
**Categories:** cs.AI primary; cs.LG, math.HO cross-list candidates
**Author-reported inception:** 2026-09-10 03:49 PDT (UTC−7) — initial insight (shared structure from failed-attempt histories); the broader GoW formulation, implementation, experiments, and release carry separate stage timestamps (see `paper/HISTORY.md`)

**Publication review gate:** [#16 — adversarial manuscript review](https://github.com/instagrim-dev/gow/issues/16).
**Review scope (updated 2026-09-12 UTC):** At the author's direction, the project's adversarial ChatGPT review process owns the commit-specific PASS/HOLD decision; the human author remains the publication authority. This replaces the earlier uninvolved-outsider requirement for this publication gate, not by claiming that requirement was met, but by explicitly changing the review arrangement. It is not independent human/outside-peer review, mathematical verification, or independent replication. Independent review remains useful but is not represented as completed. Gate #16 remains **HOLD** until its checks, including evidence availability and exact submission-package validation, are satisfied and a PASS is recorded for a specific revision.

**Core thesis:**

> When repeated Work stops producing information, treat the Work itself as data.

$$
W \xrightarrow{S} \mathcal{G}(W) \xrightarrow{B} \Delta S \xrightarrow{\Pi} W' \xrightarrow{O} \mathcal{G}'
$$

---

## The four claims (from `docs/theory/00-paper-claims.md`)

| ID | Role | Statement |
|---|---|---|
| C1 | Theory | A history of problem-solving work can be represented as a structured search space whose geometry is informative about how search is failing or progressing |
| C2 | Method | Conserved shapes, regimes, boundaries, and structural deltas provide a disciplined way to choose informative next work |
| C3 | Systems | `newf` operationalizes GoW with typed provenance, epistemic state tracking, adversarial challenge, and domain projection |
| C4 | Empirical (narrow) | In project-authored Erdős–Straus corpora, LLMs proposed reference-absent structural regularities; shape-guided context produced proposals judged closer to a withheld move; experiments do **not** establish mathematical validity, general effectiveness, or autonomous theorem solving |

**Everything else:** explicitly under hypotheses / future work.

---

## 10-stage completion plan

| Stage | Deliverable | Gate | Status |
|---|---|---|---|
| **0. Freeze claims** | `docs/theory/00-paper-claims.md` | Every major sentence classified | ✅ **Done** — 4 claims frozen; sentence-level classification key written; falsification conditions stated |
| **1. Semantic foundation** | GoW ontology + glossary | Shape, Work, outcome, regime, boundary, delta, projection, landing point defined without `newf` terminology | ✅ **Done** — `docs/theory/00-04` + glossary; LaTeX §3 has full definitions for regime, boundary, boundary delta, structural delta, projection, landing point |
| **2. Epistemic foundation** | Claims/evidence model | Conditioning, claim role, epistemic status separated; no invariant laundering | ✅ **Done** — three-axis model in domain doc; LaTeX §4 has full prose: four separations, verification hierarchy, epistemic state transitions, adversarial challenge, anti-fractal discipline |
| **3. Formalize the method** | Algorithm box | Reader can reconstruct Map→Boundary→Move→Measure from definitions | ✅ **Done** — Algorithm 1 + full step-by-step prose walkthrough + recursive-refinement discussion; Figure 2 (shape-space / domain-space loop) |
| **4. Position against prior work** | Related-work section | Matrix complete; narrow novelty claim stated | ✅ **Done** — 12 traditions compared paragraph-by-paragraph, narrow novelty claim stated as conjunction of features no single tradition fills |
| **5. Reduce `newf` to paper-relevant** | System architecture section | Only machinery necessary to instantiate GoW | ✅ **Done** — LaTeX §6 has 9-layer architecture + three epistemic guarantees + Figure 4 (architecture diagram) |
| **6. Finish empirical evidence** | Experiment section + frozen artifacts | Every empirical sentence **agrees with** the identified frozen artifact (not merely "has a path to" one) | ✅ **Closed (internal remediation) 2026-09-11** — source-faithfulness review (`docs/reviews/2026-09-11-manuscript-source-faithfulness.md`, finding 1) found the manuscript misstating pilot-003 arms/budget/classifier and pilot-004 inputs/treatment/endpoint; methods/results rebuilt from frozen protocols (`23fcc64`); a follow-up QC pass (`1780933`) additionally corrected the pilot-004 C3 endpoint to report the frozen reference-recovery measurement rather than the defensible-novel rate. Sentence-level agreement re-verified against frozen records same day. **Adversarial re-review under #16 remains open but is tracked separately (see row 10) and does not gate this internal-remediation row or the `v1.0.0` software tag.** |
| **7. One novel-shape loop closed** | One reference-absent hypothesis challenged end-to-end | At least one model-generated shape grounded and challenged, survive or fail; probe coverage and evidential character stated accurately | ✅ **Closed (internal remediation) 2026-09-11** — N2a and N5 campaigns stand; the review (finding 2) required attributed reassessments (appended to both challenge records) qualifying C2/C7 (N2a) and C2/C3 (N5); manuscript Experiment D and vocabulary seed comments corrected (`23fcc64`). The QC pass (`1780933`) further narrowed claim strength in the detailed challenge-outcomes paragraph. A second adversarial review pass (corrections in `635e6d8`) found the narrowed detail had not propagated everywhere: the progression figure, its caption, the N2a Disposition paragraph, the Discussion section, and the Appendix claims-ledger row still said "survived + refinement" / "completed lifecycle" — all five now consistently read "seven-part challenge recorded; refinement proposed; coverage incomplete." **Adversarial re-review under #16 tracked separately; not a gate on this internal-remediation row.** |
| **8. Limitations / falsifiability** | Explicit failure conditions | Paper states what evidence would make GoW uninteresting or wrong | ✅ **Done** — LaTeX §11 (7 limitations); Appendix E (F1–F5 falsification conditions with test procedures) |
| **9. Produce submission manuscript** | LaTeX + figures + bibliography | Clean compilation **and** structural consistency (no duplicate sections/labels; appendix tables agree with source records); artifact references; review gate #16 | ✅ **Closed (internal remediation) 2026-09-11** — duplicate section declarations/labels removed and Appendix C lineage corrected per review finding 5; epistemic section (finding 3) and architecture statuses (finding 4) rewritten (`23fcc64`); AI-use disclosure corrected to match the frozen record (blinded same-family session, tier-5) plus two missing citations added (`39ba88d`); disclosure further distinguishes AI-assisted drafting from human approval/freezing, and an Evidence Availability section was added (`1780933`). A second adversarial review pass (corrections in `635e6d8`) found the disclosure still overstated Pilot-004's contract as governed by `recovery-rule/v1` (Pilot-004 actually ran under a separate discovery-wire contract and predeclared quorum categories) and softened the provider-identity framing to align with the Limitations section's own same-family caveat; also corrected an ATP/CEGAR related-work overclaim ("proof failures in ATP are dead ends" — falsified by TrialMaster, added to references.bib) and the accompanying CEGAR mischaracterization. A same-day residue sweep (`314550f`) additionally reconciled the Acknowledgments section, which still said provider identities are "not load-bearing for any claim" after the Disclosure section had already been corrected to say the opposite for external-validity claims, and removed 4 unused duplicate bibliography entries. Author-owned metadata (name, ORCID) is present. CI (`paper.yml`) now fails the build on invoked draft macros, placeholder markers, or a rendered "??" cross-reference, so this row's structural-consistency gate is continuously enforced, not just checked once. **Adversarial review gate #16 and distribution-license/endorsement confirmation remain open, tracked in row 10 — not a `v1.0.0` gate.** |
| **10. Freeze arXiv v1** | Tagged paper/repo release | Every empirical number traceable to an immutable artifact **and consistent with it**; a commit-specific PASS recorded under review gate #16; exact source package validated | 🔄 **Bibliographic integrity and source-package criteria satisfied at `c2a1b1e`; PASS must now be recorded for a post-`c2a1b1e` revision.** A full 36/36-entry bibliographic audit against primary sources (DBLP, publisher pages, arXiv, ACL Anthology) landed in `c2a1b1e`: 28 entries verified clean, 8 corrected — including one apparently nonexistent citation (`palermo2004dse`, replaced with the authors' real 2005 J. Embedded Computing paper), a misspelled author surname + wrong year (`anderson2005root` → Andersen/2006), a wrong ACL page range (`an2024trialmaster` → 776–790), and a misspelled author (`lightman2023process` → Yura Burda). The exact source package was validated by clean-room compile: `geometry-of-work.tex` + `references.bib` alone produce the PDF and `.bbl` with zero bibtex warnings and zero undefined citations; arXiv's `xelatex` processor (available since Nov 2025, TeX Live 2025) matches the tectonic/XeTeX toolchain — submission should select `xelatex` and include the generated `.bbl`. Gate #16 finding 4 (evidence deposit) resolved by public repository. The repository was flipped to public visibility at `github.com/instagrim-dev/gow` under the Apache License 2.0 (`LICENSE` file added at this commit); a full-history `gitleaks` scan across 168 commits reported no leaks. The evidence supplement (`corpus/evidence-supplement/manifest.json`) now manifests **71 files** after two recorded amendments on 2026-09-12: (1) rehash of three pilot-004 records that received additive current-claim banners plus addition of `CURRENT-CLAIMS.md` (the authoritative supersession view); (2) addition of the five Experiment E artifacts when the manuscript began citing them. **The manuscript itself changed after the last review pass:** `57c3247`/`a1491b2` added Experiment E (`sec:results-episode` — preregistered closed-loop episode + matched-budget comparison, the paper's only tier-2 rows) and propagated it through the abstract, §9, Conclusion, Future Work, claims ledger, and tier annotations. Any #16 PASS recorded for an earlier commit does not cover these additions. This is a repository-based deposit; a canonical-archive DOI (Zenodo/OSF) is not obtained but is compatible with a later strengthening step. The three-lane independent-quorum review at `a98c93a` remains the recorded substitute for the uninvolved-human-reviewer criterion (still explicitly model-judgment, not peer review). The human author retains publication authority and controls the actual arXiv upload, license selection (`CC BY 4.0` recommended for the paper), and endorsement. **The `newf` software `v1.0.0` tag is unchanged and remains independent of this row.** |

---

## Completion definition

The paper is ready to submit when ALL of these are true:

1. A skeptical reader can explain GoW without mentioning `newf`.
2. Every named construct has exactly one meaning (no "candidate invariant" / "failure invariant" semantic sludge in the paper prose).
3. The method is operational enough that another person could perform a small GoW analysis manually.
4. The software architecture clearly implements the theory rather than defining it, with each capability annotated implemented-and-exercised / performed-externally-with-artifacts / specified-extension.
5. Every empirical claim points to an immutable experiment artifact **and agrees with that artifact sentence-by-sentence** (a path alone is insufficient when the sentence contradicts its source).
6. At least one model-generated reference-absent shape has undergone an actual challenge lifecycle, with probe coverage and per-probe evidential character stated accurately. (**N2a, N5 campaigns recorded; reassessed 2026-09-11; authoritative current-claim/supersession views added 2026-09-12** — `records/CURRENT-CLAIMS.md`)
7. Related work identifies the nearest intellectual ancestors and states a narrow novelty claim.
8. The paper contains explicit falsification conditions for the thesis.
9. The project's adversarial publication review records a PASS in #16 for the specific manuscript revision; its scope and participation in the project are disclosed, without representing it as independent human/outside-peer review.
10. The LaTeX source, figures, bibliography, and references compile cleanly in an arXiv-compatible environment.

---

## What does NOT block arXiv v1

- A second discipline (strengthens generality but is not necessary if ES is explicitly an exploratory case study)
- A verified mathematical advance
- A live-provider architecture
- Fully automatic adjudication
- Statistical proof that GoW outperforms conventional search — §9 ("Conditional
  Superiority", added `2bd5336`) is a conditional theorem plus a preregistered
  test design (`corpus/experiments/superiority-preregistration/`), not a
  result; the empirical premise is recorded as an open hypothesis (`H-SUP` in
  `docs/theory/00-paper-claims.md`), not established, and is not required for
  this row. Experiment E2 (2026-09-12, `corpus/experiments/two-arm-comparison/`)
  is a demonstration-scale data point of the right shape — frozen protocol,
  matched budget, exact checking — but explicitly not the specified experiment
  and asserts no statistical claim
- A complete universal ontology of Work

---

## Figures (mandatory)

| # | Description | Status |
|---|---|---|
| 1 | Geometry of Work: work items in failure/partial/success regimes; conserved shapes; boundary-crossing delta | ✅ `fig:gow-schematic` (TikZ, inline) |
| 2 | Search in shape-space, verify in domain-space: full pipeline loop | ✅ `fig:loop` (TikZ, inline) |
| 3 | Epistemic decomposition: shape × conditioning × claim role × epistemic state | ✅ `fig:epistemic-decomp` (TikZ, inline) |
| 4 | `newf` architecture: the 9-layer GoW loop as implemented | ✅ `fig:newf-arch` (TikZ, inline) |
| 5 | Pilot progression: negative controls → curated-feature → discovery → novel-shape challenge | ✅ `fig:progression` (TikZ, inline) |

---

## Near-term writing priorities (what is writable now)

These sections are fully writable from existing artifacts. No new experiments needed.

| Priority | Section | Source material |
|---|---|---|
| 1 | §3 Semantic Model (complete prose) | `docs/theory/00-01`, `glossary.md` |
| 2 | §4 Epistemic Model (complete prose) | `docs/domain-model.md` three-axis model; `docs/invariant-challenge.md` |
| 3 | §5 Method prose + recursive-refinement discussion | `docs/theory/03-shape-guided-search.md`, `N2a-challenge.md` boundary_delta |
| 4 | §10 Related Work (full text from matrix) | `docs/research/related-work-matrix.md`, `docs/theory/06-situated-in-the-literature.md` |
| 5 | §8A Negative controls | M7 run + pilot-001/002 records |
| 6 | §8B Curated-feature pilot | Pilot-003 frozen record |
| 7 | §8C Shape discovery | Pilot-004 frozen result + quorum |
| 8 | §8D Novel-hypothesis challenge | N2a-challenge.md (full lifecycle) |
| 9 | §6 System | EPIC.md, `docs/` implementation docs |
| 10 | §11 Limitations, §13 Conclusion | `docs/theory/00-paper-claims.md` §Falsification |

---

## File map

| File | Role |
|---|---|
| `paper/geometry-of-work.tex` | Main manuscript (LaTeX) |
| `paper/references.bib` | Bibliography |
| `paper/PUBLICATION-PLAN.md` | This file |
| `paper/HISTORY.md` | Provenance timeline (inception → milestones, commit-anchored) |
| `docs/theory/00-paper-claims.md` | Claim registry (stage 0 authority) |
| `docs/research/related-work-matrix.md` | Related-work comparison matrix |
| `docs/theory/06-situated-in-the-literature.md` | Literature positioning (draft §10 source) |
| `corpus/experiments/pilot-003/` | Frozen Pilot-003 artifacts |
| `corpus/experiments/pilot-004-discovery/` | Frozen Pilot-004 artifacts |
| `corpus/experiments/m7-blinded-run/` | M7 blinded run (negative controls + no_recovery) |
| `docs/findings/001-unencoded-shape-generation.md` | Finding 001 (C4 evidence) |
| `corpus/experiments/pilot-004-discovery/records/N2a-challenge.md` | N2a challenge lifecycle (stage 7 complete) |
| `corpus/experiments/pilot-004-discovery/records/N5-challenge.md` | N5 challenge lifecycle (stage 7 additional) |
| `corpus/experiments/pilot-004-discovery/records/CURRENT-CLAIMS.md` | Authoritative current claims + supersession links (governs over original verdict lines) |
| `corpus/experiments/m7-blinded-run/records/episode-001-greedy-vs-global.md` | Experiment E1: closed-loop episode (tier-2, `revision_credit: earned`) |
| `corpus/experiments/two-arm-comparison/` | Experiment E2: matched-budget comparison (protocol frozen pre-execution; deterministic runner) |
| `corpus/evidence-supplement/` | Hashed evidence selection (71 files) + reading guide; amendments recorded in manifest |
