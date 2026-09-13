# D3a development dry run: execution record

- **Authorization:** operator approved D3a as proposed (2026-09-12; ceilings in [`2026-09-12-015`](2026-09-12-015-custody-staffing-authorization-memo.md))
- **Execution form:** `internal/dryrun` — a committed deterministic test, so the run re-executes in CI on every push rather than living in one terminal session
- **Evidence label:** `development-dry-run`. This outcome cannot count toward the spending rule and the code enforces that (`GateEligible` is constitutionally false; the label is echoed, not trusted)

## What ran

The full loop, end to end, through the real components with no shortcuts:

1. **B→C admission:** three development rules (`double-not`, `add-zero`, `xor-self-zero`) admitted through `rewrite.AdmitRule`, i.e. through `finite.VerifyRuleWarrant`'s structural checks and independent replay — the dry run does not bypass the warrant boundary.
2. **Episodes:** the 24-episode roadmap population shape (12 informative across two families, 6 low-value, 6 misleading), self-authored, deterministic, with completable and uncompletable targets mixed (18 completed, 6 not — variety asserted, degenerate floors/ceilings refused).
3. **Arms:** H0, H1, HG all ran the **identical** deterministic bounded-rewriter procedure. No shaping selector exists (decision D6), and the dry run refuses to fake one; "history" inputs are deliberately absent.
4. **Checking:** every episode's endpoint independently replayed by the exhaustive oracle; a completion required both the cost target and the verified endpoint.
5. **Decision arithmetic:** the complete 24×3×1 grid evaluated by `screen.Evaluate` — population conformance, all five conditions, both cost ledgers.

## Resources against ceilings

| Ceiling (authorized) | Used |
|---|---|
| 10,000 expansions per cell | 500 budget; no episode exhausted it |
| 24 development episodes | 24 |
| 15 minutes wall clock | < 1 second |
| Zero provider spend | zero (no provider exists in the loop) |
| Runs per cell | 1, disclosed: the procedure is deterministic, run variance is structurally zero |

## Outcome (the honest expected one)

| Condition | Result |
|---|---|
| (a) no invalid certifications | pass |
| (b) HG margin ≥ 3·r over H1 | **fail at margin exactly zero** — identical arms |
| (c) control-strata loss ≤ r | pass (zero loss) |
| (d) ≥ 2 differing families | **fail** (no differing episodes exist) |
| (e) HG ≥ H0 | pass (tie) |
| ArithmeticSatisfied / RuleSatisfied / GateEligible | false / false / false |

The test asserts this shape: a dry run that satisfied the spending rule would be a dry-run bug, because nothing that could produce a margin exists yet.

## Addendum: shaped dry run (D6 plan steps 3–4)

After the D6 v0 selector shipped (`internal/shape`, misleadability proven 8-vs-19 expansions both directions), the dry run gained a second, three-arm form (`TestShapedDevelopmentDryRun`): H0 = catalog order, history withheld; H1 = implementer-authored capable comparator (global success-frequency ordering, ungated — sealed runs require non-implementer H1 review per D4); HG = `shape.Select`. Per-episode histories under a 12-expansion budget: informative episodes carry relevant successes plus irrelevant distractor noise (the ungated H1 is misled; HG's relevance gate filters it), misleading episodes carry lying relevant history (both history arms hurt), low-value episodes carry none (all arms equal).

**THE MARGIN IS DESIGNED, NOT DISCOVERED** — the implementer constructed episodes so the arms differ, to validate margin machinery, P0 attribution (every HG decision checked for controller version, snapshot hash, input hash, full ordering), and the screen's positive arithmetic path (rule satisfied on a conforming population). The outcome remains gate-ineligible with the rule satisfied — the exact boundary the screen must hold. No shaping-value claim in any direction; that requires custodian-authored episodes (D3b chain).

P0 status change: the fit-check's two open controller bindings ([`013`](2026-09-12-013-shaping-roadmap-first-tranche-execution.md)) are closed by `internal/shape`.

## What this run established, and only this

- The plumbing is sound: admission → search → oracle replay → grid → decision arithmetic compose without manual glue, on a conforming population, with real nonzero task-cost metering.
- Custody cost was **not** metered (recorded as 0 with a note); the sealed screen must meter it per §5 of the roadmap.
- Nothing about shaping value was measured, in either direction. The next object that could change condition (b) is the shaping selector (decision D6), and after it, the custodian-sealed screen (D3b).
