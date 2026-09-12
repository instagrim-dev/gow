# Preservation pilot — revision 2 proposal (preparation only, not authorized)

Status: **proposal**. Revision 1 executed and rejected the candidate
packaging under the frozen gate (`RESULTS.md`). Per protocol, revising the
packaging starts a new protocol revision with fresh authorization; this
document is that revision's preparation seed. **No runs are authorized**:
dispatch requires a new owner ceiling declaration (revision-1 ceilings are
spent).

## What revision 2 changes, and why

> **Corrected 2026-09-12 (external audit):** the original framing below —
> that the packagings lacked a label-checking obligation — was wrong. Both
> frozen packagings already required listing each result's producing
> conditions (budget, ordering, stopping rule) and checking the prose
> against them (`arms/manuscript-baseline.md:31`,
> `arms/modular-case-bundles.md:76`). The revision-1 miss was a **failure
> to enforce an existing obligation, cause unresolved**
> (`RESULTS.md` §8.3). Revision 2 therefore proposes an **enforcement
> mechanism**, not a new requirement — and it is a proposal to test, not a
> demonstrated cure.

The design change is motivated by the retained double miss on M1d
(`RESULTS.md` §4 as corrected by §8.3): both packagings verified every
number against its artifact yet neither discharged the already-present
conditioning-check obligation.

**Change (applied to BOTH arms, symmetrically):** make the existing
obligation operationally hard to skip — for every verified quantity the
report must emit a visible row:

> **quantity → producing conditions → claimed property → discriminating
> check**

where "claimed property" is the label the prose attaches to the quantity
(unconditional vs conditioned, budget-independent vs budget-sensitive,
general vs demonstration-scale, tier-N vs tier-M) and the discriminating
check states what would falsify that label. A missing row for a quantity
that appears in the manuscript is itself a reportable omission.

Revision 2 must also fix a **procedural defect**: mechanical blinding must
strip or rewrite report title lines and verify the packet with a
first-line census before adjudication (revision 1 leaked arm identity in
6 of 12 titles; `RESULTS.md` §8.1).

No other packaging text changes. Changing only the manuscript obligation's
enforcement in both arms preserves the packaging contrast as the sole
independent variable and does not reward either arm for revision-1
performance.

## Case discipline for revision 2

- Revision-1 cases are **burned for discovery purposes**: all six were
  exposed to this session's lineage and their answers now exist in
  committed reports. They may serve as smoke-test controls only if both
  arms' runs are executed by agents with no access to `docs/reviews/`
  (the deletion control already enforces this), but a fresh
  manuscript-fidelity defect pair with independently verified
  defective/corrected revisions is strongly preferred for the M-case.
- Candidate M-case correction (audit, 2026-09-12): the M-A finding was
  **found by all four manuscript runs in revision 1** — it is
  answer-exposed, not fresh. Once remediated it yields a defective/
  corrected revision pair usable only as a **known positive control**,
  not as unseen fault-family evidence and not as proof that the missed
  budget-independence check has been repaired. A genuinely fresh M-case
  must come from a defect no revision-1 report surfaced.
- S3/S4 engineering cases may be reused unchanged (the packaging change
  does not touch them) or replaced by pairs derived from any remediated
  triage lead (e.g., E-D once fixed).

## What must be redone before dispatch

1. Owner declares fresh ceilings (same simplicity discipline as §7 of
   `PREPARATION.md`).
2. Freeze revised arm files with hashes; re-validate every case pair per
   exact revision; new seeded dispatch order with a NEW seed string
   (`preservation-pilot-r2-<date>`).
3. Blinding, adjudication brief (including any new case's subtlety notes),
   and the reject-on-false-reassertion gate carry over unchanged.

Nothing in this proposal was tuned against a frozen transfer test; the
transfer experiment remains gated on a future preservation pass.
