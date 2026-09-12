# Run 09 — M1c-B

Reviewed revision: `630e380b9584e378a1aa1360efe432d4651fa439`

## 1. Decision, policy and scope

**Decision: the manuscript's quantitative and epistemic claims are, with three specific exceptions below, stated at or below the strength of their cited frozen artifacts. No evidence-laundering or silent epistemic promotion was found in the results or conclusion sections. The exceptions are fidelity/staleness defects, not overreach of headline claims.**

Policy applied: claim-fidelity review of `paper/geometry-of-work.tex` against the frozen artifacts under `corpus/` at the pinned revision, governed by the repository's epistemic invariants in `AGENTS.md` (`ModelJudgment != Verification`; no silent promotion of epistemic status; claims carry recorded strength). Scope: every numeric result and epistemic-strength statement in the Results section (Experiments A–E), the abstract, the discussion/conclusion, and the appendix tables/ledger, checked against artifact bytes; deterministic claims re-executed where a runner exists in-tree. Out of scope: mathematical correctness of corpus content, the Section 9 theorem's proofs, related-work accuracy, and code architecture beyond what the claims depend on.

## 2. Responsibility and critical path

The manuscript's credibility rests on a chain: (a) frozen artifact bytes under `corpus/experiments/` are the sole warrant for every empirical sentence (asserted at `paper/geometry-of-work.tex:1117-1120` and in the Claim/Evidence Ledger, `paper/geometry-of-work.tex:2827+`); (b) where records were later corrected, the manuscript must track the *authoritative current claim* (`corpus/experiments/pilot-004-discovery/records/CURRENT-CLAIMS.md`), not the superseded verdict lines; (c) the single tier-2 claim (Experiment E) must be reproducible from the checked-in runner. Path (c) was verified by execution; path (a) was verified byte-for-byte for every number I checked; path (b) is where the defects live: the manuscript tracks corrections for N2a itself but not for its child hypothesis, and it inherits an approximate theme count as exact.

## 3. Findings, limitations, protections, questions

### Findings

**F-1. Stale lifecycle status for N2a-child-1 understates recorded adverse evidence.**
- Location: `paper/geometry-of-work.tex:1432-1434` ("A child hypothesis (N2a-child-1) was derived but not yet independently motivated by two corpus instances and therefore not yet admitted.")
- Triggering conditions: the sentence reflects the state at authoring of `corpus/experiments/pilot-004-discovery/records/N2a-challenge.md` (pre-challenge), but the same tree at this revision records the child as *challenged and not surviving*: `records/CURRENT-CLAIMS.md:31` ("`challenged`, not surviving... general reading... weakened at C6... one attempted-and-falsified construction... does not restore `surviving`") and `records/CLOSURE-SCORECARD.md` rows "N2a-child-1 challenged 2026-09-12" and "(partial rehabilitation)".
- Violated contract: the manuscript's own traceability commitment ("Every empirical sentence in this paper is intended to be traceable to a frozen artifact", `paper/geometry-of-work.tex:2829-2831`) combined with its adoption of `CURRENT-CLAIMS.md` as the authoritative supersession-linked view (ledger row at `paper/geometry-of-work.tex:2868`). Presenting a pre-challenge status where the authoritative record holds a post-challenge adverse status is a fidelity mismatch (understating negative evidence rather than overstating positive evidence, but still a mismatch).
- Downstream consequence: a reader concludes the child hypothesis is merely pending motivation/admission; the record shows it was challenged, its general reading weakened, and a rehabilitation construction falsified. The reader's model of the challenge lifecycle's output is materially wrong.
- Discriminating regression check: require that the manuscript's status words for `N2a-child-1` match the `CURRENT-CLAIMS.md` row (e.g., "subsequently challenged; general reading weakened; not admitted"); a grep for `N2a-child-1` in the `.tex` must co-occur with "challenged"/"weakened" or a `CURRENT-CLAIMS.md` cross-reference.
- Evidence class: demonstrated static path (both files inspected at the pinned revision).

**F-2. "Five distinct themes" states an unadjudicated, approximate count as exact, and conflicts with the manuscript's own appendix.**
- Location: `paper/geometry-of-work.tex:245` (abstract: "collapsing into five distinct themes"), `:1316` (results table: "10 (collapsed to five distinct themes)"), `:1341` ("five recurring themes were identified").
- Triggering conditions: the primary record says the ten defensible-novel occurrences "contain recurring candidate themes whose distinct-property count has not been adjudicated" (`corpus/experiments/pilot-004-discovery/records/RESULT.md`, §Novel occurrences) and its theme table marks one row "~2 distinct properties across 6 occurrences"; the closure scorecard says "Novel distinct themes | **~5** (N1–N5)" (`records/CLOSURE-SCORECARD.md:302`). The manuscript's own appendix then lists **six** theme labels (N1, N2a, N2b, N3, N4, N5) as the collapse (`paper/geometry-of-work.tex:2790-2806`).
- Violated contract: claims must be stated with the strength the cited evidence supports; the record's count is explicitly approximate and unadjudicated, and the body's "five" is inconsistent with the appendix's six-label enumeration (resolvable only by silently merging N2a/N2b into one theme, which the body does at `:1341-1343` but the appendix does not).
- Downstream consequence: a reader cites "five distinct discoveries-adjacent themes" as an adjudicated result; the record explicitly warns the distinct-property count was never adjudicated ("The result does not establish ten distinct discoveries", RESULT.md §Summary).
- Discriminating regression check: abstract and results should read "approximately five recurring themes (distinct-property count not adjudicated)" or the appendix should present the same five-way grouping as the body; check = the count named in abstract/results equals the count of theme labels enumerated in the appendix, and carries the record's hedge.
- Evidence class: demonstrated static path.

**F-3. The cited Pilot-004 result record contradicts the manuscript on two points the manuscript silently corrects.**
- Location: manuscript `paper/geometry-of-work.tex:1322-1334` (C3 grading) and `:1289-1292` (D1 blinding), against the cited artifact `corpus/experiments/pilot-004-discovery/records/RESULT.md` (cited as a frozen artifact at `paper/geometry-of-work.tex:2713`).
- Triggering conditions: (i) `RESULT.md` §Criterion 3 is headed "D0 novel rate below D1 novel rate" and graded "**PASSES**" ("One of three criteria met"), whereas the manuscript grades C3 "as frozen" as at-most-weakly-consistent, following the scorecard's endpoint-definition note (`records/CLOSURE-SCORECARD.md:68-75`), which is the better reading of the frozen protocol ("the unnamed properties" = L3/L4 per `PROTOCOL-DRAFT.md:63-64`). (ii) `RESULT.md:118` labels arm D1 "guided — train-only bundle, **L1/L3/L4 supplied as surviving invariants**", contradicting the frozen protocol ("train notes themselves are the entire permitted context", `PROTOCOL-DRAFT.md:58-59`) and the manuscript's blinding claim; I verified the D1 prompt bytes (`prompts/d1-prompt.txt`) contain no reference-ledger identifiers or supplied invariants, so the manuscript's description is correct and the artifact header line is wrong.
- Violated contract: the manuscript cites `RESULT.md` as warrant while departing from two of its statements without flagging the conflict; a verifying reader following the citation finds "C3 PASSES" and a blinding-violating arm description, and cannot tell whether the paper or the record is authoritative without independently discovering the scorecard note and the prompt bytes.
- Downstream consequence: traceability failure in exactly the direction the paper claims to prevent — the paper's numbers survive checking, but its cited record actively misleads on two points. The blinding header line, if trusted, would invalidate C1's meaning (recovering supplied invariants is not discovery).
- Discriminating regression check: (i) manuscript footnote or ledger row noting that `RESULT.md`'s C3 grading is superseded by the scorecard's endpoint-definition note; (ii) a correction line in `RESULT.md` (attributed, per repo convention) fixing the D1 arm header. Check = following each manuscript citation reproduces the manuscript's statement without consulting uncited files.
- Evidence class: demonstrated static path (all four artifacts byte-inspected; prompt grep executed).

### Limitations (of this review)

- Sections 9 (conditional-superiority theorem), 10.5 (complement geometry), Related Work, and Appendix A formal definitions were skimmed, not verified; the theorem carries an explicit "no empirical claim" banner, lowering fidelity risk.
- Pilot-001/002/005, `pvnp-holdout`, and `superiority-preregistration` artifacts were not opened except where the manuscript's checked claims cited them.
- The full `go test ./...` suite was not run (only `internal/witness`); build succeeded, so untested packages pose no risk to the manuscript-fidelity verdict.
- Bibliography (`references.bib`) accuracy was not checked.

### Observed protections

- Experiment E is genuinely reproducible: the checked-in deterministic runner reproduced 3/20@57 vs 20/20@20 and cost-per-hit 19.00/1.00 exactly; the E1 witness identity is exact under arbitrary-precision arithmetic; the "1 part in 10^38" greedy-miss margin recomputes to 10^38.9; the instance rule (smallest prime ≥10^6, ≡1 mod 24) independently yields 1000033.
- The manuscript consistently adopts corrected/superseded-strength readings for N2a itself (abstract, `:1423-1434`, ledger `:2868`), including the withdrawn "5.3% is budget-independent" sentence (tracked from the attributed correction in `two-arm-comparison/RESULT.md`) and the corrected exceptional-set gloss (R-3).
- Pilot-003 and Pilot-004 numbers verified byte-for-byte: proposal counts (7/6, budget 8), automatic assessments (2 decisive_no/5 unknown; 2/4), review table (1 full recovery B3 rank 4; partials rank 2/rank 5), quorum tallies (7/10/4/2 of 23; six L1 + one L3; run-level distribution; 50% vs 36%), negative-control zeros, M7 Stage 1/2 dispositions, wire-invalid B3 capture handling, identifier-masking (not full concealment) caveat.
- The motivating example is explicitly labeled a stylized composite with an itemized list of which elements are inventions (`:446-461`) — an unusually strong protection against narrative laundering.
- Tier assignments are conservative throughout: the blinded reviewer is reported as tier-5, same-family reassessment explicitly "not independent verification", quorum consensus explicitly "still model judgment".

### Open questions

- OQ-A: Is `RESULT.md:118`'s "L1/L3/L4 supplied as surviving invariants" a copy-paste error from Pilot-003's B3 description? The prompt bytes say yes, but the record itself should say so.
- OQ-B: The abstract says defensible-novel occurrences were "not present in the reference ledger"; classification was by the same-family quorum, and RESULT.md warns novelty is "relative to this ledger" only — the abstract's parenthetical preserves this, but "defensible-novel" as a headline term may still read stronger than "candidate themes, distinctness unadjudicated".
- OQ-C: Does any consumer of the appendix's six-label theme list (N1–N5 with N2a/N2b split) depend on the body's five-count? (Determines which side F-2's fix should change.)

## 4. Checks, coverage and resources

Commands executed (all inside the extracted subject tree; exit status in brackets):

- `git archive 630e380b... | tar -x` into temp dir; removed `docs/reviews` [0]
- `grep` section map of `paper/geometry-of-work.tex` [0]
- `python3` tally of `pilot-004-discovery/quorum-result.json` — 23 entries; 7/10/4/2; L1×6, L3×1 [0]
- `python3` witness/identity check: exact identity true; greedy residual; smallest qualifying prime = 1000033 [0]
- `python3` greedy-3 truncation: sides differ by 1 at ~8.7×10^38 (≈1 part in 10^38.9) [0]
- `go build ./...` [0]
- `go run ./corpus/experiments/two-arm-comparison/` — reproduced U 3/20@57 (19.00), M 20/20@20 (1.00), verdict true [0]
- `go test ./internal/witness/` — ok [0]
- `gofmt -l .` — empty [0]
- `python3`/`grep` inspections of `recovery-facts.json`, `experiment.json` (budget 8) [0]
- assorted `grep`/`sed` reads of records listed below [0]

Files consulted (relative to subject tree): `paper/geometry-of-work.tex` (abstract, §§1–2, 3–4 excerpts, 7–8 fully, 9 banner, 10–13 excerpts, appendices B–E excerpts); `AGENTS.md` (epistemic invariants); `corpus/experiments/pilot-004-discovery/`: `quorum-result.json`, `PROTOCOL-DRAFT.md`, `records/RESULT.md` (fully), `records/CLOSURE-SCORECARD.md` (targeted), `records/CURRENT-CLAIMS.md` (targeted), `records/N2a-challenge.md` (header, verdict, reassessment), `prompts/d1-prompt.txt` (grep), `bundles/` (listing); `corpus/experiments/pilot-003/`: `records/recovery-facts.json`, `records/INDEPENDENT-REVIEW-RESULT.md` (targeted), `records/EXECUTION-RESULT.md` (targeted), `records/experiment.json`, `PROTOCOL-REVISION.md` (grep); `corpus/experiments/two-arm-comparison/`: `RESULT.md` (fully), `main.go` (executed); `corpus/experiments/m7-blinded-run/`: `RESULT.md` (targeted), `records/episode-001-greedy-vs-global.md` (fully); `corpus/experiments/2026-09-10-esr-negative-control/README.md` (targeted).

Not examined: `internal/` source beyond `internal/witness` tests; `cmd/`; `docs/`; `fixtures/`, `testdata/`; pilots 001/002/005; `pvnp-holdout`; `superiority-preregistration`; `corpus/train|target|research|evidence-supplement` note bytes; `references.bib`; Section 9 proofs; `paper/HISTORY.md`, `paper/PUBLICATION-PLAN.md`.

Wall-clock estimate: ~22 minutes.

## 5. Remediation handoffs

**H-1 (F-1): Update the N2a-child-1 sentence to the authoritative status.** Replace `paper/geometry-of-work.tex:1432-1434` with the `CURRENT-CLAIMS.md` disposition (challenged; general reading weakened at C6; one falsified rehabilitation attempt; not admitted) and cross-reference the record. Acceptance: manuscript wording matches the `CURRENT-CLAIMS.md:31` row's operative status terms; a reader following the citation finds no stronger or weaker status than stated.

**H-2 (F-2): Reconcile the theme count and carry the record's hedge.** Either state "approximately five recurring themes (distinct-property count not adjudicated; see `records/RESULT.md`)" in abstract and §8.3, or regroup the appendix list into the same five-way collapse the body uses (N2a/N2b as one theme, matching `CLOSURE-SCORECARD.md` Theme N2). Acceptance: the count in abstract/results equals the count of enumerated labels in the appendix, and the unadjudicated-count hedge appears at first use.

**H-3 (F-3): Flag the two superseded statements in the cited Pilot-004 result record.** Add an attributed correction line to `records/RESULT.md` (per the repo's existing correction convention, cf. `two-arm-comparison/RESULT.md`) covering (i) §Criterion 3's "PASSES" grading (superseded by the scorecard endpoint-definition note) and (ii) the D1 arm header's "L1/L3/L4 supplied as surviving invariants" (contradicted by `PROTOCOL-DRAFT.md` and `prompts/d1-prompt.txt`); optionally add one manuscript footnote at `:1322` noting the supersession. Acceptance: following every manuscript citation for Pilot-004 reproduces the manuscript's statements without consulting uncited files; the blinding description is consistent across protocol, record, and manuscript.
