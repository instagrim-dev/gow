# Averaging / Bombieri–Vinogradov solution count (es-05)

**Era:** Elsholtz & Tao, *Counting the number of solutions to the Erdős–Straus
equation on unit fractions* (2013). Pre-cutoff.

Let `f(n)` count the solutions of `4/n = 1/x + 1/y + 1/z`. This approach studies the
**average** of `f(p)` over primes `p ≤ N` using analytic number theory: the
Bombieri–Vinogradov theorem, the Brun–Titchmarsh inequality, the Erdős divisor
bound, and the Turán–Kubilius inequality. The result bounds the average number of
solutions polylogarithmically in `N`.

The mechanism is global and probabilistic in posture: it reasons about the
distribution of solution counts across a large ensemble of primes rather than about
any single `n`. It couples many moduli through sieve estimates and averages.

Its boundary is intrinsic to the method: the average grows only polylogarithmically.
For some Diophantine problems a positive average forces a solution to exist, but that
argument needs at least polynomial growth. The slow polylog growth here means the
averaging bound **cannot** be upgraded to "every prime has a solution" — a typical
prime has few solutions, and rare primes could in principle have none. The method
gives strong evidence but no existence proof for the quadratic-residue primes.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "erdos-straus/averaging-bv-count",
      "label": "Averaging solution count via Bombieri–Vinogradov",
      "description": "Bound the average number of solutions over primes ≤ N using sieve and analytic inequalities.",
      "mechanism": {
        "representations": ["solution-count function f(n)", "average over primes"],
        "assumptions": ["ensemble averaging is informative", "sieve estimates apply"],
        "operators": ["averaging", "sieve estimation", "divisor bound"],
        "preserves": ["mean solution density"],
        "breaks": ["per-n constructivity"],
        "auxiliary_objects": ["Bombieri–Vinogradov theorem", "Brun–Titchmarsh inequality", "Turán–Kubilius inequality"],
        "locality": "global",
        "construction_mode": "existential",
        "uncertainty_mode": "probabilistic",
        "notes": "Distributional statement about many primes at once."
      },
      "outcome": {
        "class": "partial_failure",
        "boundary_statement": "polylog average growth is too slow to force existence",
        "boundary_conditions": ["rare primes could have zero solutions", "quadratic-residue primes"],
        "notes": "Strong evidence, not an existence proof."
      },
      "support": [
        {"field_path": "mechanism.auxiliary_objects", "support_kind": "explicit", "locator": "para:2"},
        {"field_path": "mechanism.uncertainty_mode", "support_kind": "explicit", "locator": "para:3"},
        {"field_path": "outcome.boundary_statement", "support_kind": "explicit", "locator": "para:4"}
      ]
    }
  ]
}
newf-normalize -->
