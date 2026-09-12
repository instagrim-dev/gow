---
id: preservation-pilot-manuscript-baseline
revision: 1
kind: frozen_comparison_arm_prompt
arm: baseline
case_family: manuscript-fidelity
frozen: 2026-09-12
---

# Manuscript fidelity review (baseline arm, frozen)

The actual manuscript-review instruction that historically surfaced the M1
finding was operator-supplied in-session and is not retained; per the
preservation protocol this baseline is **authored and frozen before testing**.
It is a bespoke, project-specific prompt. It contains no reference to any
specific known finding.

## Instruction

Review the manuscript `paper/geometry-of-work.tex` at the pinned revision you
have been given, for **claim fidelity**: whether every quantitative or
epistemic claim in the results and conclusion sections is stated with exactly
the strength its cited evidence supports.

Operating rules:

1. Read `AGENTS.md` first; its epistemic invariants govern
   (`ModelJudgment != Verification`; no silent promotion of epistemic
   status). Record the commit SHA you reviewed.
2. For each experiment section, list every numeric result and the conditions
   under which it was produced (design, budget, ordering, stopping rule,
   population). Then check the surrounding prose: does any sentence present
   a conditioned quantity as unconditional, a demonstration-scale result as
   general, or a weaker verification tier as a stronger one?
3. For each candidate finding, report: the actionable location (file, line,
   quoted sentence); the triggering conditions (what makes the claim
   overreach); the violated contract (which stated rule or cited artifact it
   contradicts); the downstream consequence (what a reader would wrongly
   conclude); and a discriminating check (what edit or comparison would
   verify the fix).
4. Where the manuscript cites frozen artifacts under `corpus/`, verify the
   quoted numbers against the artifact bytes at the pinned revision. Report
   mismatches exactly; do not repair them.
5. Report honest unresolved cases separately from findings. Do not pad with
   generic warnings; an observation without an actionable location scores
   nothing and wastes budget.
6. You may not consult `docs/reviews/`, commit messages after your pinned
   revision, or any remediation notes. Your evidence is the manuscript, the
   cited artifacts, and the code at the pinned revision.
