# Provenance Timeline — Geometry of Work

All timestamps are US Pacific Daylight Time (UTC−7), normalized from the git
commit record (`git log --date=format-local`). Where a milestone predates the
repository, the source is the author's report, labeled as such. Elapsed times
are measured from the author-reported inception. This file exists because the
paper's honesty about its own evidence depends on the reader knowing that
every artifact it cites was authored inside one compressed window — and on
not compressing the window further by retrodating the developed framework to
the origin point.

## Origin note (author-reported inception)

> The line of inquiry that developed into Geometry of Work began on
> September 10, 2026, at 03:49 PDT. Its initial focus was extracting shared
> structure from histories of unsuccessful problem-solving attempts. The
> broader formulation emerged through subsequent discussion, implementation,
> and experimental scrutiny.

The inception timestamp attaches to the **initial insight** (shared structure
across failed attempts), not to the present theory. The shape-space /
domain-space separation, the epistemic model, the challenge lifecycle, and
the boundary concept each carry their own later, commit-anchored timestamps
below. The progression is the record; the origin point alone is not.

## Stage timestamps

| Stage | Window (PDT) | Evidence class | Anchor |
|---|---|---|---|
| Inception — initial insight | 2026-09-10 03:49 | **Author-reported** (pre-repository) | origin note above |
| Formalization of the initial focus (failure-space invariant frontier, toolbox DSL, execution + abstraction-safety contracts) | 2026-09-10 04:52–05:43 | Commit-anchored | `dfb441c`…`3a8cd7a` |
| Implementation (v0 CLI slice through hardening v32) | 2026-09-10 05:51–22:26 | Commit-anchored | `260c74e`…`3607965` |
| Experiments (pilot-001 STOP through N5 vocabulary admission) | 2026-09-10 22:43 – 2026-09-11 11:22 | Commit-anchored | `a868dec`…`9af3376` |
| Broader GoW formulation (theory series, thesis documents, manuscript) | 2026-09-11 10:26–12:47 | Commit-anchored | `259c593`…`bde6d88` |
| External source-faithfulness review + reconciliation | 2026-09-11 ~12:46–14:29 | Review record + commits | `docs/reviews/2026-09-11-manuscript-source-faithfulness.md`; `23fcc64`, `39ba88d`, `1780933` |
| Software release (`newf` substrate only) | 2026-09-11 16:44 — tag `v1.0.0` | Commit-anchored | `a70c5f1`; scope split per `CHANGELOG.md` § Versioning |
| Public release (arXiv freeze) | **Pending** — blocked on external re-review | Gate | `paper/PUBLICATION-PLAN.md` stage 10 |

## Headline

| Fact | Value |
|---|---|
| Author-reported inception | **2026-09-10 03:49** |
| First commit (`dfb441c`) | 2026-09-10 04:52 (+1h 03m) |
| Manuscript at reviewed SHA (`4f3fa56`) | 2026-09-11 12:31 (+32h 42m) |
| External source-faithfulness review received | 2026-09-11 ~12:46 (+32h 57m) |
| R1–R5 reconciliation committed (`23fcc64`) | 2026-09-11 13:35 (+33h 46m) |
| Software tag `v1.0.0` (substrate only) | 2026-09-11 16:44 (+36h 55m) |
| Public release (arXiv manuscript) | pending |
| Total commits at reviewed SHA | 132 |

## Timeline

### Day 0 — 2026-09-10 (inception → working system → first pilot)

| Time | Elapsed | Milestone | Anchor |
|---|---|---|---|
| 03:49 | 0h | **Author-reported inception** — initial insight: extracting shared structure from histories of unsuccessful attempts (pre-repository) | origin note |
| 04:52 | +1.0h | Repository init: failure-space invariant frontier defined | `dfb441c` |
| 05:11–05:43 | +1.4h | Typed toolbox DSL, agent execution contract, abstraction-safety contract | `1bd226d`, `7183a71`, `3a8cd7a` |
| 05:51 | +2.0h | First code slice: provenance-first v0 CLI, SQLite-backed problem/run records | `260c74e` |
| 06:07–06:50 | +2.3h | Immutable source ingestion; approach normalization; mechanism records | `770ce39`, `55dc654` |
| 07:00 | +3.2h | Erdős–Straus historical-holdout corpus (pre-cutoff atlas + quarantined target) | `8551dbe` |
| 07:10–08:25 | +3.4h | Deterministic canonicalization/comparison; clustering; failure-space artifact | `fefead0`, `d56cd00` |
| 10:07–11:25 | +6.3h | Invariant mining (M4.2); challenge/falsify (M4.3); frontier generation (M5.1) | `c048a69`, `008d3a8` |
| 11:59–12:44 | +8.2h | Evaluation + verifier routing (M5.2); acceptance-gate hardening | `d03ea53` |
| 12:55–13:32 | +9.1h | Success compression (M6.1); search-policy mutation (M6.2) | `24cdc24`, `4bfe614` |
| 14:01 | +10.2h | Blinded-benchmark validation experiment (M7 v0) | `ec2ce51` |
| 14:34–22:26 | +10.8h | Hardening passes v23–v32: signature-revision binding, completeness admission, occurrence-based reassessment, untrusted proposer wire, external arms | `760e54a` … `3607965` |
| 22:43–23:03 | +18.9h | **Pilot-001**: protocol + readiness — honest STOP (no surviving invariants) | `a868dec`, `b0ad3ee` |
| 23:18–23:54 | +19.5h | Pilot-001 revised: mapping review, L1–L6 ledger adjudicated; mechanism/v2 + interpretation claims | `70e8f51`, `df84bde` |

### Day 1 — 2026-09-11 (pilots 002–004 → challenges → manuscript → review)

| Time | Elapsed | Milestone | Anchor |
|---|---|---|---|
| 00:01 | +20.2h | **Pilot-002**: successor protocol mechanically READY (3 surviving invariants) | `c457109` |
| 00:24–00:55 | +20.6h | **Pilot-003**: prepared under mechanism/v3; B0/B3 captured from isolated sessions; blinded comparison executed — no recovery, delta inconclusive | `6cd1c0c`, `f2dff65`, `dfe892e` |
| 01:59 | +22.2h | Pilot-003 blinded independent review (model judgment): B3 rank 4 judged to recover the withheld move | `c85f40c` |
| 02:41 | +22.9h | Pilot-003 CLOSE: exploratory milestone met, recovery provisional; external-review packet prepared | `ec8d3e5` |
| 02:50–03:06 | +23.0h | Pilot-003 post-hoc classify/v2 reassessment; **Pilot-004** protocol drafted, decisions cemented, bundles/wire/prompts built | `583bfc4`, `aa0b95f` |
| 05:06 | +25.3h | Pilot-004 FREEZE: operator attestation; capture opens | `7e58c2c` |
| 05:22–06:24 | +25.6h | Six captures sealed (3×D1, 3×D0 incl. one recorded host-fault blocker + resolution); 23-entry arm-blind adjudication ledger | `5674d28`, `dc7b678` |
| 07:47–09:14 | +28.0h | Three-lane quorum adjudication: 7 reference / 10 novel / 4 unsupported / 2 over-merge / 3 disputed | `3f2d987`, `78e73d2` |
| 09:27–09:39 | +29.6h | Pilot-004 result records (predeclared verdict: negative/inconclusive); **N2a challenge record** | `8307b22`, `be9d3df` |
| 09:47 | +30.0h | mechanism/v4 admitted (`density_averaging_ceiling`) | `bf331ae` |
| 10:26–11:29 | +30.6h | Theory series, practitioner field guide, M7 blinded-run record, claim registry update | `144a3e8`, `7280550` |
| 11:01 | +31.2h | **Manuscript scaffold**: claims frozen, complement geometry | `fe5f412` |
| 11:22 | +31.6h | N5 challenge + mechanism/v5 admitted (`reorganisation_without_qr_existence`) | `9af3376` |
| 11:48–12:07 | +32.0h | Manuscript sections filled; appendices; five TikZ figures; compile verified | `80bcecf`, `9aaade9` |
| 12:15–12:31 | +32.4h | Author metadata + ORCID → **reviewed SHA `4f3fa56`** | `e55e7e7`, `4f3fa56` |
| ~12:46 | +32.9h | **External source-faithfulness review received** (5 findings; 7/7 spot-checkable claims locally confirmed) | `docs/reviews/2026-09-11-manuscript-source-faithfulness.md` |
| 12:46–12:47 | +33.0h | Theory series 07–10, thesis worked example, pilot-005 comparative protocol committed | `bde6d88`, `71196af` |
| ~13:00–13:20 | +33.5h | **R1–R5 reconciliation authored**: methods/results rebuilt from frozen records; N2a/N5 attributed reassessments; epistemic + architecture sections corrected; assembly deduplicated; publication gates reopened pending external re-review | committed at `23fcc64` (below) |
| 13:07 | +33.3h | Authoritative GoW naming canon adopted | `321c356` |
| 13:30 | +33.7h | Manuscript redesigned as a monograph | `c0777fc` |
| 13:35 | +33.8h | **R1–R5 reconciliation committed** (source-faithfulness blockers 1–5) | `23fcc64` |
| 13:36 | +33.8h | CI wired: repo gates + M7 reproducibility + paper compile with duplicate-label guard | `87fbf97` |
| 13:41–14:29 | +33.9h | Source-fidelity residue sweep; v1.0.0 gap analysis; QC correction pass (endpoint reporting, claim strength) | `39ba88d`, `4e38ecb`, `1780933` |
| 16:44 | +36.9h | **Software tag `v1.0.0`** — versions the `newf` substrate only; the arXiv manuscript is versioned separately and remains gated | `a70c5f1` |
| 17:19–17:20 | +37.5h | CHANGELOG corrections (quorum count 23; canonical repo URL after rename to `instagrim-dev/gow`) | `7ec89dc`, `9895dc0` |

## What this timeline is evidence for

1. **Same-window authorship.** The corpus, the system that analyzes it, the
   experiments, and the manuscript were authored by the same operator-plus-model
   process inside ~33 hours. This is exactly the condition falsification
   condition F5 (transfer) exists to test against, and the paper's Limitations
   section says so explicitly. Nothing in this timeline supports a claim of
   longitudinal validation, independent replication, or cross-domain transfer.

2. **The epistemic machinery ran in real time, not retrospectively.**
   Predeclared endpoints preceded captures (`7e58c2c` FREEZE before any
   pilot-004 capture); the pilot-001 STOP, the pilot-003 wire-invalid capture,
   the pilot-004 host-fault blocker, and the negative pilot-004 verdict were
   all recorded as they happened rather than repaired afterward. The one place
   the narrative drifted from the frozen record — the manuscript itself — was
   caught by external review within fifteen minutes of the reviewed SHA and is
   reconciled as *correction, not confirmation* (per
   `docs/theory/09-falsification-review.md`).

3. **Compression is a cost claim, not a quality claim.** A 33-hour idea-to-
   manuscript arc demonstrates throughput of the operator-plus-model process.
   It does not strengthen any scientific claim in the paper, and the paper
   must not cite it as if it did.

4. **The framework is not retrodated to the inception timestamp.** What began
   at 03:49 was the initial insight; the theory series, semantic model,
   epistemic model, and manuscript acquired their timestamps when they were
   committed, more than a day later. Claiming the present formulation existed
   at inception would be a provenance promotion of exactly the kind the
   epistemic model prohibits for every other artifact in this repository.

## Correction discipline

This file is append-only in spirit: correct errors by adding dated notes, not
by rewriting rows. Timestamps derive from the commit record and can be
re-verified with:

```bash
git log --reverse --format='%h %ad %s' --date=format-local:'%Y-%m-%d %H:%M'
```

### Correction note — 2026-09-11 (evening)

A same-day adversarial review found this file stale within hours of creation:
status rows still said "working tree" after the reconciliation was committed
(`23fcc64`), the headline said "Public release: pending" without
distinguishing the software surface after tag `v1.0.0` was pushed, and the
timeline ended at ~13:20 while eight committed events followed. Status-bearing
cells were updated to current fact, the Day-1 table was extended through
17:20, and the stage table gained the software-release row with its scope
split (`CHANGELOG.md` § Versioning: `v1.0.0` versions the substrate; the
arXiv manuscript is gated separately). The lesson is recorded rather than
erased: a provenance document is itself a claim about the repository and
drifts like any other claim unless re-verified against the commit record.
