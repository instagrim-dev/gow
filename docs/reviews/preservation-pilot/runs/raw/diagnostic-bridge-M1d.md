# Run report — manuscript claim fidelity

## 1. Decision, policy and scope

**Reviewed revision:** `00c3c5808b8df5a5257a34a9ffc695d406f68294`.

**Decision under review:** does every quantitative or epistemic claim in the results and conclusion-facing sections of `paper/geometry-of-work.tex` carry exactly the strength its cited evidence supports?

**Verdict: predominantly faithful, with three material fidelity defects and one appendix-level protocol misdescription.** Every headline number I could trace (Pilot-003 tables, Pilot-004 quorum counts and per-arm rates, negative-control zeros, Experiment E1/E2 quantities) reproduces exactly from the frozen artifacts or by re-execution of the in-tree deterministic runner. The defects are property *labels*, not values: one labeled invariance ("budget-independent") is refuted by direct execution over the unrestricted range; one adjective ("symmetric") overstates an asymmetric decisive negative; and the manuscript's evidence-integrity guarantee (hash-verifiable artifacts) fails at this revision for the exact artifact backing the Pilot-004 composite-endpoint ledger row.

**Policy applied:** value verification and label verification treated as separate obligations. For each quantity carrying a property label in the reviewed sections, I extracted the label as a proposition, stated its proof obligation, and ran the cheapest discriminating check executable inside the subject tree (metamorphic variation of the in-tree experiment runner, exact recomputation, artifact cross-reads). Verifying a number was never counted as discharging its label. The repository's own epistemic invariants (`AGENTS.md`: no silent promotion; `ModelJudgment != Verification`) were used as the governing contract for tier and strength language.

**Scope:** §7 Experimental Method, §8 Results (Experiments A–E), the Evidence Availability statement, and Appendices B (protocol details), C (complete tables), D (claim/evidence ledger), against `corpus/experiments/` and `corpus/evidence-supplement/` at the pinned revision. §9 (Conditional Superiority) self-declares "theoretical — no empirical claim" and was checked only for that declaration's consistency, which holds.

## 2. Responsibility and critical path

The paper's argument rests on a chain: (a) negative controls show the system abstains honestly; (b) Pilot-003/004 provide tier-4/5 observations, explicitly not established results; (c) Experiment E is the sole tier-2 (reproducible computation) evidence and carries the paper's only decisive quantitative comparison; (d) the claims ledger plus the hashed evidence supplement make every empirical sentence auditable.

The critical path for fidelity is therefore (c) and (d): E's numbers and labels must survive re-execution, and the audit guarantee must actually hold at the citing revision. Both were tested directly. E's *values* fully reproduce (runner re-executed: 3/20 @ 57 subs vs 20/20 @ 20 subs; E1 residue and witness recomputed exactly). But E carries the refuted "budget-independent" label (F1), and (d) fails the hash check for `CLOSURE-SCORECARD.md` and a manifest-listed missing file (F3). Responsibility for F1 is shared between the frozen artifact (`RESULT.md` line 44 makes the same unqualified claim) and the manuscript, which repeated it without the discriminating check; F2 and F4 are manuscript-authoring defects (the artifacts are correct and more precise than the prose); F3 is a release-engineering defect (post-manifest appends to a frozen-and-hashed record without amendment).

## 3. Findings, limitations, protections, questions

### Mandatory bridge-row table (labeled quantities in reviewed sections)

| # | Extracted claim → | Mathematical obligation → | Cheap discriminating check → | Result |
|---|---|---|---|---|
| B1 | Local-move hit rate 3/57 ≈ 5.3% is "budget-independent" (tex:1495; artifact RESULT.md:43–44) | Rate invariant over budget B, unrestricted (label unqualified) | Rerun runner copy with B varied, all else fixed | **EXECUTED — REFUTED.** B=2 → 3/40 = 7.5%; B=3 → 3/57; B=4 → 3/57. Invariance holds only on the restricted range B≥3 (all three local moves admitted); rate changed because added observations arrive at a different proportion |
| B2 | Experiment E is "tier-2 (reproducible computation)" (tex:1458–1459) | Every reported E quantity re-derivable by deterministic computation from the tree | Re-execute runner; recompute E1 arithmetic | **EXECUTED — PASS for E2 and E1 arithmetic** (all values matched); E1's 97 s wall-clock, 5 CLI invocations, and `revision_credit: earned` derive from a live DB not in the tree — inspected via typed export only |
| B3 | Checker is "exact-integer... no floats, no tolerance" (tex:1461–1465) | Verdicts decided by exact identity `4xyz = n(yz+xz+xy)` over big ints | Recompute greedy tuple and witness exactly; run package tests | **EXECUTED — PASS.** Greedy sides differ by exactly 1 at ~8.7×10³⁸ scale; witness identity exact; `go test ./internal/witness` ok |
| B4 | Greedy miss margin "1 part in 10³⁸" (tex:1477–1478) | Relative miss of frozen step-1 tuple ≈ 10⁻³⁸ | Exact `Fraction` recomputation of 4/p − Σ1/xᵢ | **EXECUTED — PASS.** Residue = 1/2.17×10⁴⁴ ⇒ relative miss 10⁻³⁸·⁹; matches artifact byte-for-byte |
| B5 | p = 1000033 = "smallest prime ≥ 10⁶ ≡ 1 (mod 24)" (tex:1472–1474) | No smaller qualifying prime exists | Enumerate 10⁶..1000033 | **EXECUTED — PASS.** Only other candidate 1000009 = 293×3413 (composite) |
| B6 | M7 Stage 2 produces a "symmetric, decisive negative" (tex:1207) | Both arms decisively assessed with the same negative verdict | Read `compare-b0-b3.json` arm records | **EXECUTED READ — REFUTED.** B0: proposal_count 0, decisive_count 0 (unassessed); B3: 3 proposals, decisive, recovered=false. The negative is one-armed; symmetry held only for Stage-1 abstention |
| B7 | B3 rank-4 recovery is "tier-5 (single-model judgment), not tier-1–2" (tex:1265–1267) | Evidence chain contains no verifier stronger than one model family | Inspect review + reassessment records | **INSPECTED — PASS.** Artifact matches; paper never promotes it; same-family reassessment correctly kept at tier-5 |
| B8 | Defensible-novel rates 36% (D0) vs 50% (D1), scoped as novelty-classification, not the C3 endpoint (tex:1336–1339) | Rates recompute from cemented categories under the arm assignment; conditioning stated | Join `quorum-result.json` with `captures/adjudication-key.json` | **EXECUTED — PASS.** 4/11 = 36.4%, 6/12 = 50.0%; C1 (1-of-3 D1 runs), C2 (2 vs 0), C3 per-reference facts (L1: 3 D1 + 3 D0; L3: D1 only; L4: none) all recompute exactly |
| B9 | Two-arm runner "deterministic... zero model calls" (tex:1457–1458; RESULT.md) | Repeated runs identical; no provider path | Run twice; inspect source | **EXECUTED — PASS.** Identical outputs; `ProbablyPrime(64)` is exact below 2⁶⁴ so instance selection is effectively deterministic; no provider imports |
| B10 | Quorum aggregation: "2-of-3 majority determines cemented_category; non-unanimous outcomes recorded as disputed" (tex:2699–2700) | Every cemented category the output of the stated rule | Recompute lane agreement for all 23 entries | **EXECUTED — MISDESCRIBED.** 3 of 23 entries (E04, E10, E22) had *no* majority; artifact rule cements the most conservative vote and flags disputed. Five 2-of-3 non-unanimous entries are *not* disputed, contradicting both the appendix sentence and the Table-C caption's own definition of *Disp.* (tex:2738) |

### Findings

**F1 — Unqualified "budget-independent" label on the 3/57 rate is false over the unrestricted range.**
Location: `paper/geometry-of-work.tex:1495`; root artifact `corpus/experiments/two-arm-comparison/RESULT.md:43–44`. Triggering condition: any reader varies the budget below 3 (the label invites exactly this use — it is offered as *the* transportable number). Violated contract: an invariance label is a proposition over its stated range; unqualified means unrestricted (also the manuscript's own conditioned-vs-unconditional discipline, tex §7). Downstream consequence: the "budget-independent measurement" is quoted as the class-level local hit rate; anyone re-deriving cost-per-hit at another budget from it inherits a wrong denominator model (at B=2 the true rate is 7.5%). Discriminating regression check: run the runner with `const budget = 2`; assert the printed local rate equals the labeled one — currently fails. Evidence class: **executed reproduction** (budget-2 and budget-4 variants run; outputs above). Note the rate is additionally sample-scoped (first 20 instances), which "in this class" soft-pedals.

**F2 — "Symmetric, decisive negative" overstates M7 Stage 2.**
Location: `paper/geometry-of-work.tex:1207`; contradicting artifact `corpus/experiments/m7-blinded-run/records/compare-b0-b3.json` (baseline arm: 0 proposals, 0 decisive) and `RESULT.md` ("B3 alone is decisively no_recovery"). Triggering condition: reading the sentence as parallel to Stage 1's genuinely symmetric abstention, which the sentence's own structure invites. Violated contract: verification-tier fidelity — presenting a one-armed decisive outcome as two-armed strengthens the negative control beyond its evidence; the immediately preceding sentence (B0-vs-B3 inconclusive because B0 emits zero proposals) contradicts the adjective in-paragraph. Downstream consequence: the "first decisive negative result in the corpus" claim reads as broader than it is. Regression check: assert no prose adjective claims multi-arm symmetry where `compare-b0-b3.json` shows `unassessed_count`/zero-proposal arms. Evidence class: **demonstrated static path** (artifact read; no execution needed).

**F3 — The Evidence Availability integrity guarantee fails at this revision.**
Location: `paper/geometry-of-work.tex:2552` ("Content integrity can be verified against the recorded hashes") vs `corpus/evidence-supplement/manifest.json`. Executed sweep of all 71 manifest rows: 69 ok, **1 hash mismatch** (`corpus/experiments/pilot-004-discovery/records/CLOSURE-SCORECARD.md` — the cited source for the composite-endpoint ledger row, tex:2850), **1 missing file** (`corpus/experiments/pilot-003/records/pre-capture.db`, absent from the tree though the availability text promises every cited record "in its exact byte-form"). Triggering condition: any reader performs the verification the paper invites. Cause visible in-tree: the scorecard accumulated post-manifest appendix rows (entries pinned to later HEADs) after its recorded 2026-09-12 rehash, without a new amendment. Violated contract: the manifest's own freeze rule ("this manifest freezes WHICH files and their byte content") and the paper's auditability claim. Downstream consequence: the audit chain for Pilot-004's headline negative verdict is broken precisely where it matters; a skeptical reader cannot distinguish annotation drift from result tampering. Regression check: CI step hashing every manifest row at the paper-citing revision; currently fails on two rows. Evidence class: **executed reproduction**.

**F4 — Appendix B misdescribes the quorum aggregation rule that decided C2.**
Location: `paper/geometry-of-work.tex:2699–2700` and Table-C caption tex:2736–2739, vs `quorum-result.json` `aggregation_rule`. The artifact rule has a no-majority → most-conservative-vote fallback; it decided 3 of 23 entries, including **both** over-merge entries (E10, E22: votes defensible_novel/matches_reference/over_merge) whose counts fail criterion C2. The appendix says majority determines category and non-unanimous ⇒ disputed; in fact only no-majority entries are disputed and five 2-of-3 entries (E03, E08, E09, E12, E18) have objecting lanes yet Disp.=no. The main text (tex:1324–1325) honestly says "both conservative fallbacks", so this is an appendix-internal inconsistency, but C2's failure is decided by the fallback, so the rule's correct statement is load-bearing. Regression check: recompute the appendix Agree/Disp. columns from the artifact and diff against the stated definitions. Evidence class: **executed reproduction** (all 23 agreements recomputed).

### Limitations and honest unresolved cases (not findings)

- E1's wall-clock (97 s), CLI-invocation count, and code-derived `revision_credit` rest on a live SQLite store not checked in; the typed JSON export and the arithmetic were verified, the persistence-trigger claims were not re-executable here.
- Freeze-ordering claims ("protocol frozen at d75b4e0 before execution"; manifest `basis_commit`) are commit-history assertions not decidable from the subject tree alone; I treated them as recorded, not verified.
- N2a/N5 challenge content is model-judgment by declaration; the paper's per-probe honesty (C2/C7 not discharged) matches the ledger, but I did not re-read the full challenge records line-by-line.
- `m7-blinded-run/run.sh` (claimed "fully offline and deterministic") was not executed within the time budget.

### Protections observed (working as designed)

Conditioning discipline in §8 is generally strong: the C3 endpoint-definition note is propagated correctly from the scorecard into prose and ledger; the 36%/50% figures are explicitly demoted to descriptive; tier language is consistently self-limiting; the negative-control zeros (0 invariants, 0 proposals, six content axes empty, outcome/posture axes non-empty) all match `failure-space.json` and the compare records; `go build ./...` passes on the whole tree.

### Questions for the authors

1. Is "budget-independent" intended as "invariant for all budgets admitting the full local family (B≥3)"? If so, say that; if not, what range is claimed?
2. Should `pre-capture.db` be in the public tree, or should the manifest and availability text record its exclusion explicitly?

## 4. Checks, coverage and resources

**Commands executed (all inside the extracted subject tree; exit status in brackets):**

- `git archive <SHA> | tar -x` setup extraction [0]
- `go run ./corpus/experiments/two-arm-comparison/` — twice, identical output [0]
- budget-metamorphic variants: `sed 's/budget = 3/budget = {2,4}/'` into scratch copies + `go run` [0]
- `python3` primality/enumeration check for p=1000033 [0]
- `python3` exact-Fraction recomputation of E1 greedy residue, relative margin, and witness identity [0]
- `python3` joins/recomputations over `quorum-result.json`, `quorum-adjudication-packet.json`, `captures/adjudication-key.json` (category counts, per-arm rates, per-run C1 facts, lane-agreement for all 23 entries) [0]
- `python3` reads of `compare-b0-b3.json` (M7 and negative-control) and `failure-space.json` [0]
- `python3` SHA-256 sweep of all 71 `corpus/evidence-supplement/manifest.json` rows [0; 2 integrity failures found]
- `go build ./...` [0]; `go test ./internal/witness/` [0]
- assorted `grep`/`sed`/`ls`/`wc` over the paper and artifacts [0]

**Files consulted:** `paper/geometry-of-work.tex` (§§7–9, Results, Evidence Availability, Appendices B–D); `corpus/experiments/two-arm-comparison/{PROTOCOL.md,RESULT.md,main.go}`; `corpus/experiments/m7-blinded-run/{RESULT.md,records/episode-001-greedy-vs-global.md,records/compare-b0-b3.json}`; `corpus/experiments/pilot-003/records/{EXECUTION-RESULT.md,INDEPENDENT-REVIEW-RESULT.md}`; `corpus/experiments/pilot-004-discovery/{quorum-result.json,quorum-adjudication-packet.json,captures/adjudication-key.json,records/CLOSURE-SCORECARD.md}`; `corpus/experiments/2026-09-10-esr-negative-control/{README.md,compare-b0-b3.json,failure-space.json}`; `corpus/evidence-supplement/manifest.json`; `AGENTS.md`.

**Not examined:** pilot-001/002/005, pvnp-holdout, superiority-preregistration directories; pilot-003 raw captures/prompts and the wire-invalidity/repair records; N2a/N5 challenge records and `CURRENT-CLAIMS.md` in full; `m7-blinded-run/run.sh` execution; §9's proofs; `internal/` beyond the witness package; `docs/`; `gofmt`/full `go test`.

**Wall-clock estimate:** ≈ 23 minutes.

## 5. Remediation handoffs

1. **Re-scope the "budget-independent" label** in `paper/geometry-of-work.tex:1495` and `corpus/experiments/two-arm-comparison/RESULT.md:43–44` (e.g. "invariant for every budget that admits all three local moves, B≥3" or publish the per-(instance,move) hit matrix, which *is* budget-free). Acceptance: a budget-2 run of the runner no longer contradicts any printed label; the label states its range explicitly.
2. **Correct M7 Stage-2 symmetry wording** at `paper/geometry-of-work.tex:1207` and align the Appendix-B aggregation-rule sentence and Table-C caption (tex:2699–2700, 2736–2739) with the artifact rule (conservative no-majority fallback; disputed ⇔ no majority). Acceptance: prose asserts a decisive negative for B3 only, with B0 unassessed; recomputing Agree/Disp. from `quorum-result.json` under the stated definitions reproduces the printed table.
3. **Restore evidence-supplement integrity** — amend `manifest.json` with a rehash (and attributed reason) for `CLOSURE-SCORECARD.md`, and either check in `pilot-003/records/pre-capture.db` or record its exclusion in both manifest and Evidence Availability text. Acceptance: the full-manifest SHA-256 sweep reports 0 mismatches and 0 missing files at the revision the paper cites, and a repeatable check guards it.
