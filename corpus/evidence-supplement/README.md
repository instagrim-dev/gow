# Evidence Supplement — Reading Guide

**Status: selected and hashed, not yet publicly deposited.** This directory
freezes *which* frozen records back the empirical claims in
`paper/geometry-of-work.tex`, and their exact byte content (`manifest.json`),
so that a later public deposit (Zenodo / OSF / a read-only mirror repository)
has a fixed, auditable selection to publish rather than an ad hoc export.

This is not a copy of the files — it is a manifest (paths + SHA-256 hashes)
pointing at the existing locations inside `corpus/experiments/`. Publishing
this supplement means depositing the *listed files*, not this repository.

**Amendment 2026-09-12** (recorded in `manifest.json` → `amendments`): three
pilot-004 records were rehashed after receiving additive, attributed
current-claim banners/annotations (supersession views, `faff2a6`), and
`records/CURRENT-CLAIMS.md` was added to the selection — depositing without
it would publish the superseded N2a/N5 claim strength that the 2026-09-12
external review objected to. No historical text was edited; the amendments
follow the redaction/frozen-artifact discipline below.

## What this is not

- **Not the whole repository.** Only the three frozen experiment corpora the
  manuscript's Results section and Appendix claims ledger
  (`\label{app:claims}` in `paper/geometry-of-work.tex`) depend on are in
  scope: `m7-blinded-run`, `pilot-003`, `pilot-004-discovery`.
- **Not a security-reviewed release.** A first-pass grep sweep for common
  secret patterns (`api_key`, `Bearer `, `sk-ant-`, `sk-proj-`, etc.) and email
  addresses found nothing across the selection, but this is not an exhaustive
  security review. See `manifest.json`'s `open_blockers`.
- **Not yet public.** No deposit host has been chosen. That is an author
  decision (hosting cost, DOI-minting, licensing terms) this pass does not
  make — see `manifest.json`'s `open_blockers`.

## How to verify the selection is unmodified

```bash
cd <repo root>
python3 - <<'PY'
import hashlib, json
m = json.load(open("corpus/evidence-supplement/manifest.json"))
bad = []
for e in m["files"]:
    h = hashlib.sha256(open(e["path"], "rb").read()).hexdigest()
    if h != e["sha256"]:
        bad.append(e["path"])
print("OK" if not bad else f"MISMATCH: {bad}")
PY
```

## Reading guide — claims ledger to source file

Each row of the manuscript's Appendix claims ledger
(`paper/geometry-of-work.tex`, `\label{app:claims}`) is answered by one or more
files in this selection:

| Claims-ledger row (abbreviated) | Answering file(s) |
|---|---|
| System refuses to manufacture failure support when none exists | `corpus/experiments/2026-09-10-esr-negative-control/` *(not included in this supplement — outside the three named corpora; see Explicit exclusions)* |
| M7 Stage 1: pristine target yields `inconclusive` for all arms | `m7-blinded-run/RESULT.md`, `m7-blinded-run/records/readiness.json` |
| M7 Stage 2: operator-attested target yields `no_recovery` | `m7-blinded-run/records/compare-b0-b3.json`, `m7-blinded-run/target-attested/es-target-affine-lattice-linear-forms.attested.md` |
| Blinding audit clean (0 leakage) | `m7-blinded-run/RESULT.md`, `m7-blinded-run/run.sh` (reproduction script) |
| B0: 7 proposals, 0 full recovery, 1 rank-2 partial | `pilot-003/records/INDEPENDENT-REVIEW-RESULT.md` |
| B3: 6 proposals, 1 full recovery (rank 4), 1 rank-5 partial | `pilot-003/records/INDEPENDENT-REVIEW-RESULT.md` |
| B3 rank 4 = fiberwise lattice-point / geometry-of-numbers proposal | `pilot-003/records/INDEPENDENT-REVIEW-RESULT.md`, `pilot-003/records/CLOSURE.md` |
| Automatic assessment: no recovery under `classify/v1`; `classify/v2` reassessment post-hoc | `pilot-003/records/EXECUTION-RESULT.md`, `pilot-003/records/reassessment-v2/` |
| Pilot-003 protocol / arms / budgets | `pilot-003/PROTOCOL-REVISION.md` |
| Pilot-003 external-review packet + reviewer instructions | `pilot-003/review-derivative/external-review-packet.json`, `pilot-003/review-derivative/REVIEWER-INSTRUCTIONS.md`, `pilot-003/review-derivative/derivation-record.json` |
| Pilot-004: 23 proposal occurrences, three-lane quorum | `pilot-004-discovery/quorum-result.json`, `pilot-004-discovery/quorum-lane-{A,B,C}-votes.json` |
| 7 matches_reference / 10 defensible_novel / 4 unsupported / 2 over_merge | `pilot-004-discovery/quorum-result.json`, `pilot-004-discovery/adjudication-ledger.json` |
| Predeclared composite endpoint not met | `pilot-004-discovery/records/CLOSURE-SCORECARD.md` |
| Pilot-004 design / arms / freeze attestation / contract | `pilot-004-discovery/PROTOCOL-DRAFT.md`, `pilot-004-discovery/FREEZE.md`, `pilot-004-discovery/WIRE.md` |
| N2a (density-averaging ceiling) proposed as reference-absent hypothesis | `pilot-004-discovery/quorum-result.json` (entries E13, E16, E20, E23) |
| N2a seven-part challenge: refinement proposed, coverage incomplete | `pilot-004-discovery/records/N2a-challenge.md` (current-claim banner atop; superseded verdict line flagged in place) |
| N5 seven-probe campaign | `pilot-004-discovery/records/N5-challenge.md` (current-claim banner atop) |
| Authoritative current claim + supersession links (N2a, N5, children, E15) | `pilot-004-discovery/records/CURRENT-CLAIMS.md` — **read this first**; it governs over original verdict lines per the 2026-09-12 review's propagation finding |
| Interpretation notes / vocabulary admission context | `pilot-004-discovery/records/INTERPRETATION-NOTES.md` |
| Reference-absent shape generation observed | `docs/findings/001-unencoded-shape-generation.md` *(not included — outside the three named corpora)* |

Rows marked "not included" point at manuscript-cited artifacts outside the
three frozen experiment corpora this supplement selects; they are excluded per
`manifest.json`'s stated `selection_scope`, not omitted by oversight.

## Verification-tier reminder

Per `paper/geometry-of-work.tex` §Verification-tier annotations: everything in
this supplement is tier-4 (same-family multi-lane agreement) or tier-5
(single-model judgment) evidence. Nothing here is tier-1 (formal proof) or
tier-2 (reproducible deterministic computation). Reading these records proves
what was recorded and when — it does not upgrade any claim's verification
tier.

Separately from this selection: the repository now contains tier-2
(reproducible deterministic computation) artifacts — the closed-loop episode
(`corpus/experiments/m7-blinded-run/records/episode-001-greedy-vs-global.md`)
and the two-arm comparison (`corpus/experiments/two-arm-comparison/`), both
adjudicated by the exact-integer witness checker. They are **not in this
supplement** because the manuscript's claims ledger does not yet cite them;
if a manuscript revision adds them, they must be added to `manifest.json` by
a recorded amendment, not silently.

## Redaction discipline (if a redaction pass is later requested)

Any redacted copy must use a distinguishable filename or suffix (e.g.
`-public.md`) and must never edit an original frozen file in place. This
mirrors the repository's general frozen-artifact discipline: corrections are
additive and attributed, not silent rewrites of the historical record.
