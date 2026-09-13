# Review report — run-07-M1d-B

## 1. Decision, policy and scope

- **Reviewed revision:** `00c3c5808b8df5a5257a34a9ffc695d406f68294`, examined as an extracted archive tree with `docs/reviews/` removed.
- **Review subject:** the manuscript `paper/geometry-of-work.tex` at the pinned revision, assessed for **claim fidelity** — whether every quantitative or epistemic claim in the results and conclusion sections is stated with exactly the strength its cited evidence supports.
- **Policy applied:** `AGENTS.md` epistemic invariants govern (`ModelJudgment != Verification`; no silent promotion of epistemic status; `Evidence != Hypothesis`). Where the manuscript cites frozen artifacts under `corpus/`, quoted numbers were verified against artifact bytes at the pinned revision. Judgment-level review was distinguished from rule-based assessment. Mismatches are reported, not repaired.
- **Decision:** the manuscript's results and conclusion sections are, with two localized exceptions, faithful to the frozen artifacts — every headline number I checked (Pilot-003 arm counts and verdicts, Pilot-004 quorum counts and per-arm criteria, E1 episode figures, E2 comparison figures, appendix tables) reproduces exactly from the cited artifact bytes, and the paper's epistemic hedging consistently matches or understates artifact strength. The two exceptions are **stale status descriptions** of derived hypotheses (N2a-child-1 and Theme N4/E15) that describe an earlier state of the record than the one frozen at this revision. Both understate rather than overstate progress, but both disagree with the artifact that the paper's own claim ledger declares authoritative.

## 2. Responsibility and critical path

The manuscript's evidential load rests on five experiment subsections (§8.1–§8.5) plus three appendices (Complete Experiment Tables, Claim/Evidence Ledger, Falsification Conditions). The critical path for claim fidelity is:

1. **Abstract and Conclusion** (lines 223–261, 2371–2430) — the strongest-worded summaries; every number there must trace to §8 and then to `corpus/experiments/`.
2. **§8.3 Pilot-004 (Experiment C)** — the most quantitative section (23 entries, composite endpoint C1–C3, per-arm rates); source of the paper's only claimed discovery-adjacent observation. Sources: `quorum-result.json`, `adjudication-ledger.json`, `captures/adjudication-key.json`, `records/CLOSURE-SCORECARD.md`.
3. **§8.5 Experiment E** — the paper's only claimed tier-2 (reproducible computation) evidence; it must actually reproduce. Sources: `corpus/experiments/two-arm-comparison/` (runner + RESULT), `corpus/experiments/m7-blinded-run/records/episode-001-greedy-vs-global.md`.
4. **§8.4 Experiment D (N2a)** — the section where a superseded "survives challenge" verdict could most easily leak into prose; the governing artifact is `records/CURRENT-CLAIMS.md`, which declares itself authoritative over stale summaries.
5. **Appendix Claim/Evidence Ledger** — the paper's self-imposed contract that "every sentence agrees with the identified source."

I prioritized executing what could be executed (the E2 runner, the E1 arithmetic, per-arm recomputation of Pilot-004 statistics from the sealed key) and byte-checking the rest.

## 3. Findings, limitations, protections, questions

### Findings

**F1 — N2a-child-1 status is stale against the authoritative claim record.**
- **Location:** `paper/geometry-of-work.tex:1432–1434` ("A child hypothesis (N2a-child-1) was derived but not yet independently motivated by two corpus instances and therefore not yet admitted.")
- **Triggering conditions:** at the pinned revision, `corpus/experiments/pilot-004-discovery/records/N2a-child-1-challenge.md` (recorded 2026-09-12) exists and `records/CURRENT-CLAIMS.md` (the file the record set declares governs "when a summary elsewhere disagrees") records N2a-child-1 as **`challenged`, not surviving**, with a two-reading boundary split: the local reading (two distinct obstructions) survives at n=2 corpus instances; the general reading (no joint-crossing mechanism) was weakened at C6 and later gained one attempted-and-falsified construction. The manuscript sentence describes the child's pre-challenge state (as of `N2a-challenge.md` §Derived child hypothesis, 2026-09-11) and gives a support-based non-admission rationale that the newer record supersedes — in particular, "not yet independently motivated by two corpus instances" sits uneasily against the record's "local reading survives at n=2."
- **Violated contract:** the manuscript's own ledger standard (`geometry-of-work.tex:2823–2830`: every sentence must *agree with* the identified source at re-reading) and `CURRENT-CLAIMS.md`'s reading rule ("when a summary elsewhere disagrees with a row below, this file… governs").
- **Downstream consequence:** a reader concludes the child hypothesis is an unexamined derivation awaiting support, when the frozen record shows it was already challenged with a recorded split disposition; the reader also misses that one falsification attempt against its general reading was completed. The error direction is understatement, but it misrepresents which lifecycle stage the artifact reached.
- **Discriminating regression check:** for every hypothesis ID named in §8.4 (N2a, N2a-child-1, N2a-child-2, N5), diff the manuscript's status phrase against the corresponding `CURRENT-CLAIMS.md` row; the check passes only when each manuscript sentence names the disposition recorded there (`challenged, not surviving` for child-1).
- **Evidence class:** demonstrated static path (byte comparison of manuscript sentence vs. both records at the pinned revision).

**F2 — Theme N4 (E15) described as pending admission; the artifact closed it as `weakened`.**
- **Location:** `paper/geometry-of-work.tex:2795–2796` (appendix theme list: "N4 structural `breaks=[]` annotation-pattern observation (E15; scope correction required before admission)").
- **Triggering conditions:** `corpus/experiments/pilot-004-discovery/records/CLOSURE-SCORECARD.md` (N4 admission gate row) records the gate **closed 2026-09-12**: the corrected claim E15′ entered `proposed`, was challenged via seven probes, and landed **`weakened`** (two weakening landings, one merge signal, two category challenges); derived observations E15″/E15‴ were not admitted. `CURRENT-CLAIMS.md` (E15 row) confirms: quorum verdict retained as historical record only; E15′ challenged and `weakened` — not admitted as surviving.
- **Violated contract:** same ledger standard as F1; also the no-silent-status rule read in reverse — the manuscript keeps a candidate alive ("required before admission" implies admission is still the expected next step) that the frozen record has already adjudicated.
- **Downstream consequence:** a reader tallying the five themes treats N4 as a live pending candidate rather than a closed, weakened one, slightly inflating the surviving-novelty inventory implied by the appendix.
- **Discriminating regression check:** parse the appendix theme bullets; for each theme with a challenge or closure record under `records/` or `records/challenges/`, require the bullet to carry the recorded terminal disposition (`weakened`/closed for N4). Passes only when the N4 bullet names the closure.
- **Evidence class:** demonstrated static path.

No overclaim was found in the abstract, results, discussion, limitations, or conclusion: I specifically probed the places where promotion would be easiest (B3 rank-4 "recovery" — consistently labeled tier-5 model judgment; N2a "survival" — consistently downgraded to "recorded challenge, coverage incomplete," matching the attributed reassessment; E2 — explicitly budget-scoped with the budget-4 sensitivity disclosed; §9 — explicitly marked "theoretical, no empirical claim") and found the hedges present and artifact-consistent in each case.

### Limitations (of this review)

- The LaTeX was not compiled; line-anchored citations refer to source lines, and no rendered-output check (e.g., broken refs, figure/caption drift) was performed.
- The mathematical content of §9 (conditional enrichment theorem, Hoeffding certificate) and §Formal Definitions was not verified; I checked only that the section claims no empirical result.
- The M7 runbook (`corpus/experiments/m7-blinded-run/run.sh`) was not executed; M7 Stage 1/2 and negative-control claims were checked against `RESULT.md`/`README.md`/JSON bytes only.
- `references.bib`, Related Work, and the pilot-001/002/005 and pvnp-holdout artifact sets were not examined (the results/conclusion sections do not draw quantitative claims from them).
- The blinded-review over-read reassessment (§8.2 caveat 5) and the operator audit of treatment-revealing language were accepted from the manuscript's description; the underlying `review-derivative/` and `SECOND-ADJUDICATION-RESULT.md` bytes were not read line-by-line.

### Observed protections

- **E2 reproduces exactly** (executed): `go run ./corpus/experiments/two-arm-comparison/` yields 3/20 hits at 57 submissions (U) vs 20/20 at 20 (M), decision `true` — identical to `RESULT.md` and to every manuscript occurrence (abstract, §8.5, ledger).
- **E1 arithmetic re-derived independently** (executed): recomputing the greedy 3-term truncation at p=1000033 gives an identity miss of exactly 1 on ~3×10³⁸-magnitude sides ("1 part in 10³⁸" is fair), and the recorded witness (250009, 83339083442, 718489947739347664906) satisfies 4xyz = n(yz+xz+xy) exactly. The 97 s wall-clock and `revision_credit: earned` match the episode record.
- **Pilot-004 statistics recompute from primary artifacts** (executed): joining `quorum-result.json` with the sealed `adjudication-key.json` reproduces every per-arm figure in §8.3 — category totals 7/10/4/2 of 23; D1 novel rate 50% (6/12) vs D0 36% (4/11); L1 matched 3×D1 + 3×D0; L3 once, D1 only; L4 never; C1 satisfied in exactly 1 of 3 D1 runs; C2 over-merges 2 (both D1) vs 0. The appendix's 23-row table, including the Agree and Disputed columns, matches vote-level recomputation row for row.
- **Pilot-003 figures match key + judgments** (executed): 7 B0 / 6 B3 proposals; the only `recovers` verdict is P10 = B3 rank 4; partials P07 = B0 rank 2 and P04 = B3 rank 5; automatic counts (B0: 2 decisive_no/5 unknown; B3: 2/4) match `recovery-facts.json`. The wire-invalid B3 primary capture and repaired-derivative assessment are disclosed in the manuscript exactly as `EXECUTION-RESULT.md` records them.
- **Supersession discipline is real and propagated:** `CURRENT-CLAIMS.md` exists, is dated, and the manuscript's N2a and N5 prose (§8.4, ledger, limitations §12) matches the *downgraded* strengths (C2→IIP+PE, C7→SGO+PE; N5 six-of-seven with C3 inapplicable), including the corrected exceptional-set gloss.
- **Tree is healthy:** `go build ./...`, `go test ./...` (all packages ok), and `gofmt -l .` (empty) all pass.

### Open questions

- **Theme count presentation:** the abstract and §8.3 say ten defensible-novel occurrences collapse "into five distinct themes" (N2a/N2b counted as one N2, matching the scorecard's "~5 (N1–N5)"), but the appendix itemizes six bullets (N1, N2a, N2b, N3, N4, N5). A reader counting bullets gets six. Which presentation is intended as canonical?
- **HISTORY.md scope:** the +38.1h freeze-time figure is present at `paper/HISTORY.md:105`; whether later remediation entries were appended after that row (as the limitations paragraph implies) was not determinable from the tree alone.

## 4. Checks, coverage and resources

Commands executed (all inside the extracted subject tree; exit status in parentheses):

- `git archive` + extract + remove `docs/reviews` (0) — setup.
- `ls` / `wc -l` on tree, `paper/`, `corpus/experiments/` (0).
- Grep of manuscript section map (1 on shell-escaped variant, then 0 via search tool).
- Python: count/categorize `quorum-result.json` entries (0); join with `adjudication-key.json` for per-arm rates, reference matches, C1/C2 recomputation (0); recompute appendix Agree/Disputed columns from lane votes (0).
- Python: pilot-003 `review-key.json` × `independent-assessment.json` join (0); `recovery-facts.json` per-arm counts (0).
- Python: independent recomputation of E1 greedy expansion, residual, identity gap, and exact witness check (0).
- `go run ./corpus/experiments/two-arm-comparison/` (0) — full E2 reproduction.
- `go build ./...` (0), `go test ./...` (0, all packages ok), `gofmt -l .` (0, empty).
- Targeted greps of `RESULT.md` (m7, two-arm, episode record), `EXECUTION-RESULT.md`, `CLOSURE-SCORECARD.md`, negative-control README/JSONs, `vocabulary_seed.go`, `paper/HISTORY.md` (0; one grep returned 1 on a no-match probe of an absent RESULT.md path before `ls` fallback).

Files consulted (relative to subject tree): `paper/geometry-of-work.tex` (sections 1, 7–14, appendices; ~1,300 lines read), `paper/HISTORY.md` (spot), `AGENTS.md`, `corpus/experiments/pilot-004-discovery/{quorum-result.json, adjudication-ledger.json, decision-handoff-packet.json, captures/adjudication-key.json, records/{CLOSURE-SCORECARD.md, CURRENT-CLAIMS.md, N2a-challenge.md (child section), N2a-child-1-challenge.md}}`, `corpus/experiments/pilot-003/{review-packet.json, records/{review-key.json, independent-assessment.json, recovery-facts.json, holdout.json, EXECUTION-RESULT.md}}`, `corpus/experiments/two-arm-comparison/{PROTOCOL.md, RESULT.md, main.go (via go run)}`, `corpus/experiments/m7-blinded-run/{RESULT.md, records/episode-001-greedy-vs-global.md}`, `corpus/experiments/2026-09-10-esr-negative-control/{README.md, compare-b0-b3.json, compare-b2-b3.json}`, `internal/canon/vocabulary_seed.go` (spot), directory listings of `corpus/experiments/*`.

Not examined: LaTeX compilation and rendered output; §9 and Appendix A mathematics; `references.bib`; Related Work sources; pilot-001, pilot-002, pilot-005, pvnp-holdout, superiority-preregistration contents; `m7-blinded-run/run.sh` execution; pilot-003 `captures/` prompt/raw bytes and `review-derivative/`; pilot-004 `bundles/` and `prompts/` bytes beyond entry counts; `docs/` prose; `internal/` beyond build/test and one seed file.

Wall-clock estimate: ~20 minutes.

## 5. Remediation handoffs

1. **Update the N2a-child-1 sentence (§8.4).** Replace the "derived but not yet… admitted" clause with the disposition recorded in `records/CURRENT-CLAIMS.md` (challenged, not surviving; local reading survives at n=2; general reading weakened with one falsified crossing attempt). **Acceptance:** the manuscript sentence names the `challenged` disposition and cites `N2a-child-1-challenge.md`; the F1 regression check (per-hypothesis status diff against `CURRENT-CLAIMS.md`) passes for all four hypothesis IDs.
2. **Close Theme N4 in the appendix.** Amend the N4 bullet to state that the corrected claim E15′ was challenged and landed `weakened` and the admission gate is closed, citing `records/challenges/E15-prime-challenge.md`. **Acceptance:** the appendix bullet carries the terminal disposition; the F2 regression check (theme bullets vs. closure records) passes for all themes with challenge records.
3. **Reconcile the theme count presentation.** Either merge the appendix's N2a/N2b bullets under a single N2 heading with two framings, or change "five distinct themes" to a formulation that survives bullet-counting. **Acceptance:** the abstract, §8.3, and the appendix present the same theme cardinality, and the scorecard's "~5 (N1–N5)" remains the cited source.
