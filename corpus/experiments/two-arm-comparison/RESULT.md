# Two-arm comparison — RESULT

Recorded: 2026-09-12. Protocol frozen at `d75b4e0` **before execution**;
runner `main.go` (deterministic, exact-integer, zero model calls; wall-clock
0.11 s). Reproduce: `go run ./corpus/experiments/two-arm-comparison/`.

## Measured outcome (primary endpoint)

| Arm | Verified witnesses | Checker submissions | Cost per hit |
|---|---|---|---|
| U (undirected, cheap-local-first) | 3 / 20 | 57 | 19.00 |
| M (map-ordered: G1 first) | **20 / 20** | **20** | **1.00** |

**Frozen decision rule verdict: the map earns its cost on this class** —
Arm M ≥ hits (20 vs 3) with strictly fewer submissions (20 vs 57).

## Secondary observables

- **Local-move hit rate in the 1 (mod 24) class: 3/57 ≈ 5.3%** (L2 hit at
  p = 1000033, 1000273, 1001713; L1 and L3 hit nowhere). The map's implicit
  prediction — local/constructive moves rarely reach this class — is now a
  measured number, and the local family is demonstrably not a strawman: it
  does sometimes hit.
- **G1 hit 20/20, always at a submission cost of 1** (first admissible x;
  the x-range bound of 10⁴ was never stressed).
- **Arm U never reached G1**: with budget B = 3 and no abstentions, the
  undirected cheap-local-first ordering exhausts its budget inside the local
  family on 17/20 instances. Under this budget, the map's entire measured
  value is *knowing which family to skip*.

## Honest reading

1. **What is measured:** the ordering value of the map's failure geometry on
   this instance class at this budget — a real, preregistered, externally
   checked cost difference (57 vs 20 submissions for 3 vs 20 witnesses).
2. **What is not measured:** discovery. Both arms share a fixed 4-move pool;
   the map did not invent G1, it only promoted it. A map that must *find*
   a new mechanism family is a different (harder) experiment.
3. **Sensitivity, stated plainly:** the verdict is budget-dependent. At
   B = 4, Arm U would reach G1 after three local misses and likely hit,
   changing cost-per-hit to ~4 vs 1 rather than ∞-like. B = 3 was frozen in
   the protocol before execution, but readers should understand the
   contrast compresses as budget grows. The 5.3% local hit rate is the
   budget-independent number.
4. No claim about the Erdős–Straus conjecture follows; all 20 instances are
   small and each witness is independently checkable by one multiplication.

> **Correction (2026-09-12, attributed; original text above preserved):**
> point 3's closing sentence — "The 5.3% local hit rate is the
> budget-independent number" — is too strong and is withdrawn. The 3/57
> rate is conditioned on this run's frozen procedure: the L1→L2→L3
> submission ordering, the budget B = 3, and stop-on-hit truncation (a hit
> suppresses the instance's remaining local submissions — e.g., p = 1001713
> hit at L2, so L3 was never attempted there). The defensible statement:
> the local-family hit rate is *less budget-sensitive* than the
> cost-per-hit contrast, but it is a property of this ordering and stopping
> rule, not of the class alone. Raised by the 2026-09-12 external
> publication review (P1); the manuscript's Experiment E section was
> corrected the same day.

## Relation to the review

This closes the "prospective comparison" half of the 2026-09-12 review's
epistemic row at demonstration scale: episode-001 showed the closed loop;
this study shows a preregistered, matched-budget, externally checked cost
comparison where the map-guided arm wins under a frozen decision rule — and
records the conditions (budget, fixed pool, known class) under which that
verdict holds.
