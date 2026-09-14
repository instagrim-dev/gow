# Nonprotected G1 observation mutation check

This development-level control runs four fresh `observation-claim/1` inputs through a fresh local review ledger at `9b896e7`. It is not a protected G1 v2 rerun, score, or full roadmap closure.

The [rate control](summary.json) changes only the high-budget second-submission `success` value. Its completed, assessor-invoked result changes from `HOLDS_AT_COMPARED_POINTS` to `REFUTED`, so the observed-rate path is not accepting the baseline independently of supplied data.

The solved-set control changes only the high-budget shared-prefix `success` value. The baseline has a verified extension and holds. The mutant makes that trace non-nested; the assessor runs and returns `INAPPLICABLE`, preserving the premise guard instead of treating an invalid comparison as a monotonicity refutation. A genuine nested extension cannot lose an already solved prefix instance, so this is the discriminating negative control for the solved path.

No provider calls, protected data, mutation of shipped code, or protected-case allocation were used.
