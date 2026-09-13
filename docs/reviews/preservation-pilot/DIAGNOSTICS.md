# Mechanism diagnostics (2026-09-12) — outside the preservation protocol

These are labeled diagnostics, not preservation-pilot runs: revision-1
ceilings are spent, revision 2 is unauthorized, and nothing here counts as
pilot evidence. Purpose: check whether the corrected record's claims and
the proposed revision-2 enforcement mechanism hold up before any fresh
authorization bets on them.

## D1 — executed verification of the audit's budget sweep (RESULTS §8.2)

Method: extracted the pinned runner at `00c3c58`
(`corpus/experiments/two-arm-comparison/main.go` + `internal/witness` +
module files) via `git archive` into a tmpdir; swept `const budget` by
one-token substitution; executed each variant (`go run .`).

```text
budget 1: local submissions 20, hits 0    (0%)
budget 2: local submissions 40, hits 3    (7.5%)
budget 3: local submissions 57, hits 3    (5.263%)
budget 4: local submissions 57, hits 3    (plateau, original ordering)
```

Result: the external audit's exact-integer diagnostic is **confirmed by
executing the pinned runner itself** — same code path as the manuscript's
number, one condition varied. Budgets one and two falsify the unrestricted
budget-independence label; the ≥3 plateau supports only a range-scoped
claim the defective sentence did not make. RESULTS §8.2 is now grounded in
an executed reproduction in this repository's lineage, not only the
auditor's independent translation.

## D2 — does the claim-form bridge cause the catch? (single dispatch)

Question: revision 2 proposes the bridge row (extracted claim →
mathematical obligation → discriminating check → result) as an enforcement
mechanism. Both revision-1 packagings contained the conditioning obligation
and both reviewers failed to enforce it; a mechanism that merely restates
the obligation would fail the same way. This diagnostic tests the row
format on the burned M1d case.

Design and contamination controls:

- One fresh-context agent, defective tree at `00c3c58` via `git archive`
  with `docs/reviews/` deleted (committed answers unreachable).
- Packaging: revision-1 bundle M (byte-identical) + the bridge obligation
  + the claim-form reference **with the worked example removed and the
  motivating-incident sentence genericized** (the frozen reference names
  the M1 answer; handing it over unscrubbed would make the diagnostic
  circular). Scrub verified: zero occurrences of 3/57, 3/40, "budget-
  independent measurement", "Experiment E", "stop-on-hit" in the handed
  copy. The claim FORM "X-independent" remains — that is the mechanism
  under test, not the answer.
- Honest residual contamination: the case is answer-exposed in this
  session's lineage and in the public repo history; the agent cannot reach
  either, but this remains a known-positive-control diagnostic, not
  discovery evidence and not a revision-2 result.

Outcome: recorded below when the dispatch returns.

### D2 outcome (2026-09-12, ~16:45 PT)

**The bridge caused both selection and execution of the discriminating
check — the burned M1 defect was caught.** The run's mandatory bridge-row
table contains, as row B1: extracted claim (the 3/57 "budget-independent"
label at `tex:1495` and its root artifact `RESULT.md:43–44`) → obligation
(rate invariant over budget, unrestricted since the label is unqualified) →
check (rerun a runner copy with the budget varied, all else fixed) →
result: **EXECUTED — REFUTED** (B=2 → 3/40 = 7.5%; B=3 → 3/57; B=4 →
3/57), with the restricted-range qualification (invariance holds only for
B ≥ 3) and the accounting explanation in ratio-identity terms ("rate
changed because added observations arrive at a different proportion").

Beyond the target, the same mechanism produced: F1's root-cause split
(the frozen artifact makes the same unqualified claim; the manuscript
repeated it without the discriminating check); a second label defect
("symmetric" overstating an asymmetric decisive negative); a real
evidence-integrity finding (the `CLOSURE-SCORECARD.md` hash-chain break at
`00c3c58` — independently discovered here, and known to have been fixed
later at `630e380`, which corroborates the diagnostic's accuracy); and a
remediation proposal better than re-labeling (publish the per-
(instance,move) hit matrix, which is genuinely budget-free).

Honest limits, unchanged from the design: n = 1 on a known positive
control; the claim form "X-independent" was present in the scrubbed
reference (that is the mechanism, but a form-level hint nonetheless); the
revision-1 comparison baseline had different agents on a different day.
This diagnostic demonstrates the bridge row CAN convert the prose
obligation into an executed refutation on this case; it does not measure
how often it does, which is what an authorized revision 2 would test.

Report retained at `runs/raw/diagnostic-bridge-M1d.md` (2,224 words,
within cap).
