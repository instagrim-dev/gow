# Pilot 001 — structural-recovery comparison (protocol scaffold)

Status: **template — not yet executed.** Every `[ ]` and `TBD` is an operator
decision or research judgment that must be filled before the run. The harness
half of readiness is computed by `newf experiment readiness --problem <id>`;
nothing in this file may claim readiness that command does not report, and
nothing that command reports substitutes for the attestations below.

## Claim under test (narrow, fixed before capture)

> Does explicit failure-derived structure (challenged, surviving invariants)
> improve the directions an external proposer proposes, measured as
> structural recovery of a withheld advance under `recovery-rule/v1` /
> `classify/v1`?

This is a structural-recovery pilot. It is NOT a novel-progress claim, NOT a
discovery-effectiveness result on its own, and a label-level match is never a
verified mathematical advance.

## 1. Corpus readiness (package 2 — research judgment)

- [ ] Selected failure cases each state: method M, under assumptions A and
      restrictions B, reaches documented boundary D — with a source locator.
      A paper stopping short of the conjecture is NOT a failure record.
- [ ] Outcome classes and scopes reviewed by: `TBD (name)` on `TBD (date)`.
- [ ] Vocabulary: the terms REQUIRED by this experiment's comparison axes
      resolve; all other terms deliberately left unresolved (claim narrowed
      accordingly). Reviewed mappings: `TBD`.
- [ ] Withheld target: `TBD (source)`, structural criterion: `TBD (what
      counts as recovering its move, stated before generation)`.
- [ ] Mechanical half: `newf experiment readiness` output attached below
      (`READY` required). Output: `TBD`.

## 2. Split freeze

- problem (train): `TBD (prb_...)`
- target problem: `TBD (prb_...)`
- holdout set: `TBD (hset_...)` — mode `blinded` (no chronological claim)
- frozen at commit: `TBD (sha)` — db snapshot hash: `TBD`

## 3. Capture protocol (the harness cannot establish these)

Same proposer, same configuration, both arms; predeclared budgets.

| field | B0 capture | B3 capture |
|---|---|---|
| model identity + version | TBD | TBD (must equal B0) |
| configuration (temp, seeds, etc.) | TBD | TBD (must equal B0) |
| prompt file (retained verbatim) | TBD | TBD |
| permitted context | source material only — **no inferred invariants, no target** | source material + surviving invariants — **no target** |
| capture file | `b0-captured.json` | `b3-captured.json` |
| capture sha256 | TBD | TBD |
| captured at | TBD | TBD |

- [ ] B0's context verifiably excluded the invariants (prompt retained).
- [ ] Neither prompt contains the withheld target's information.
- [ ] Wire files validate (`proposal-wire/v1`; explicit `target_invariant_ids`
      for B3 — an empty selection is a violation by design).

## 4. Execution

```bash
newf experiment readiness --problem <train>            # must be READY
newf experiment run --problem <train> \
  --b0-proposals-file b0-captured.json \
  --b3-proposals-file b3-captured.json \
  --count <N> --evaluation-budget <M>
newf experiment compare --problem <train>
```

Record from the run (cite EXECUTION rows, not "latest experiment"):

- execution run ids + per-arm generation ids + file sha256s
  (`experiment_executions`, v32): `TBD`
- experiment id + identity hash: `TBD`
- admission audit per arm (`admission_corrected/downgraded/stripped/rejected/overflow`): `TBD`
- budgets configured vs consumed: `TBD`
- raw result + comparison JSON committed beside this file.

## 5. Assessment (package 4 — research judgment)

- [ ] Independent reviewer (`TBD`) assesses, for every `recovered` and
      `decisive_no` classification, whether the comparator's verdict
      corresponds to the intended mechanism — using the pre-stated criterion
      from §1, blind to which arm produced the proposal where practicable.
- [ ] Disagreements recorded verbatim; the comparator's classification is
      never silently overridden — both are reported.
- [ ] Cost accounting: work spent deriving + challenging B3's invariants is
      reported alongside any effectiveness statement.

## 6. Interpretation discipline

- `unknown` / `inconclusive` outcomes are acceptable results, not failures of
  the pilot. An unintended all-target failure is a protocol defect.
- The gate is a meaningful, attributable comparison whose recorded result
  follows from the actual inputs — **not "B3 wins"**.
- Post-run findings are classified per the freeze-gate policy in
  `docs/experiment.md`: fix-before-interpreting / fix-alongside / defer.
- Known tracked limitation: interpret replayed experiments by explicit
  experiment id (see `docs/plans/2026-09-10-011-audit-current-state-readers.md`,
  finding 2).
