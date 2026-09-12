# Preservation pilot — revision 2 proposal (preparation only, not authorized)

Status: **proposal**. Revision 1 executed and rejected the candidate
packaging under the frozen gate (`RESULTS.md`). Per protocol, revising the
packaging starts a new protocol revision with fresh authorization; this
document is that revision's preparation seed. **No runs are authorized**:
dispatch requires a new owner ceiling declaration (revision-1 ceilings are
spent).

## What revision 2 changes, and why

The single design change is motivated by the retained double miss on M1d
(`RESULTS.md` §4): both packagings verified every number against its
artifact and then affirmed the section as properly hedged, because both
bias manuscript review toward **numeric agreement** — which a mislabeled
epistemic property of a correct number passes by construction.

**Change (applied to BOTH arms, symmetrically):** add one obligation to the
manuscript case packaging —

> For every verified quantity, separately check the *label attached to it*:
> every classification of a statistic (unconditional vs conditioned,
> budget-independent vs budget-sensitive, general vs demonstration-scale,
> tier-N vs tier-M) is itself a claim requiring verification against the
> producing conditions. Verifying the number does not verify its label.
> Report quantity-label pairs, not quantities.

No other packaging text changes. Changing only the manuscript obligation in
both arms preserves the packaging contrast as the sole independent variable
and does not reward either arm for revision-1 performance.

## Case discipline for revision 2

- Revision-1 cases are **burned for discovery purposes**: all six were
  exposed to this session's lineage and their answers now exist in
  committed reports. They may serve as smoke-test controls only if both
  arms' runs are executed by agents with no access to `docs/reviews/`
  (the deletion control already enforces this), but a fresh
  manuscript-fidelity defect pair with independently verified
  defective/corrected revisions is strongly preferred for the M-case.
- Candidate fresh M-case source: the M-A finding (stale N2a-child-1
  characterization, `FINDINGS-TRIAGE.md`) once remediated will yield a new
  independently verified defective/corrected revision pair — with the
  attractive property that it is, again, a status/label defect rather than
  a numeric defect.
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
