# Publication Plan — Geometry of Work

**Target:** arXiv flagship paper introducing GoW as a discipline.
**Title:** Geometry of Work: Shape-Guided Search over Failed and Partial Solution Attempts
**Alternate subtitle:** Search in Shape-Space, Verify in Domain-Space
**Format:** LaTeX, single-column, ~15–20 main-text pages + appendices
**Categories:** cs.AI primary; cs.LG, math.HO cross-list candidates
**Author-reported inception:** 2026-09-10 03:49 PDT (UTC−7) — initial insight (shared structure from failed-attempt histories); the broader GoW formulation, implementation, experiments, and release carry separate stage timestamps (see `paper/HISTORY.md`)

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
| **6. Finish empirical evidence** | Experiment section + frozen artifacts | Every empirical sentence **agrees with** the identified frozen artifact (not merely "has a path to" one) | 🔄 **Reopened 2026-09-11** — source-faithfulness review (`docs/reviews/2026-09-11-manuscript-source-faithfulness.md`, finding 1) found the manuscript misstating pilot-003 arms/budget/classifier and pilot-004 inputs/treatment/endpoint; methods/results rebuilt from frozen protocols same day; gate closes after external re-review confirms sentence-level agreement |
| **7. One novel-shape loop closed** | One reference-absent hypothesis challenged end-to-end | At least one model-generated shape grounded and challenged, survive or fail; probe coverage and evidential character stated accurately | 🔄 **Reopened 2026-09-11** — N2a and N5 campaigns stand, but the review (finding 2) required attributed reassessments (now appended to both challenge records) qualifying C2/C7 (N2a) and C2/C3 (N5); manuscript Experiment D and vocabulary seed comments corrected; gate closes after re-review |
| **8. Limitations / falsifiability** | Explicit failure conditions | Paper states what evidence would make GoW uninteresting or wrong | ✅ **Done** — LaTeX §11 (7 limitations); Appendix E (F1–F5 falsification conditions with test procedures) |
| **9. Produce submission manuscript** | LaTeX + figures + bibliography | Clean compilation **and** structural consistency (no duplicate sections/labels; appendix tables agree with source records); artifact references; external review | 🔄 **Reopened 2026-09-11** — duplicate section declarations/labels removed and Appendix C lineage corrected per review finding 5; epistemic section (finding 3) and architecture statuses (finding 4) rewritten; external review and author-owned metadata still pending |
| **10. Freeze arXiv v1** | Tagged paper/repo release | Every empirical number traceable to an immutable artifact **and consistent with it**; findings 1–5 of the 2026-09-11 review closed by re-review | 🔄 **Blocked on re-review** — empirical consistency, epistemic consistency, and implementation traceability were reopened by the 2026-09-11 review; these are manuscript-consistency obligations, not author metadata; freeze only after an external pass confirms the reconciliation |

---

## Completion definition

The paper is ready to submit when ALL of these are true:

1. A skeptical reader can explain GoW without mentioning `newf`.
2. Every named construct has exactly one meaning (no "candidate invariant" / "failure invariant" semantic sludge in the paper prose).
3. The method is operational enough that another person could perform a small GoW analysis manually.
4. The software architecture clearly implements the theory rather than defining it, with each capability annotated implemented-and-exercised / performed-externally-with-artifacts / specified-extension.
5. Every empirical claim points to an immutable experiment artifact **and agrees with that artifact sentence-by-sentence** (a path alone is insufficient when the sentence contradicts its source).
6. At least one model-generated reference-absent shape has undergone an actual challenge lifecycle, with probe coverage and per-probe evidential character stated accurately. (**N2a, N5 campaigns recorded; reassessed 2026-09-11**)
7. Related work identifies the nearest intellectual ancestors and states a narrow novelty claim.
8. The paper contains explicit falsification conditions for the thesis.
9. A technically competent outsider reviews it without having participated in these conversations.
10. The LaTeX source, figures, bibliography, and references compile cleanly in an arXiv-compatible environment.

---

## What does NOT block arXiv v1

- A second discipline (strengthens generality but is not necessary if ES is explicitly an exploratory case study)
- A verified mathematical advance
- A live-provider architecture
- Fully automatic adjudication
- Statistical proof that GoW outperforms conventional search
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
