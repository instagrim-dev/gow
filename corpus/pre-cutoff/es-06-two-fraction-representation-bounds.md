# Two-unit-fraction representation bounds (es-06)

**Era:** Browning & Elsholtz (2010–2011); used inside the Elsholtz–Tao counting work.
Pre-cutoff.

This approach reduces `4/n = 1/x + 1/y + 1/z` to the question of representing a
rational as a sum of **two** unit fractions once the first denominator is fixed: for
a chosen `x`, one asks whether `4/n − 1/x = 1/y + 1/z` is solvable. Bounds on the
number of representations of a rational `a/b` as `1/y + 1/z` (a divisor-counting
question about `b` and `a`) then control how many full solutions exist.

The mechanism is global and existential: it counts representations across ranges of
`n` and uses those counts to derive lower bounds `f(n) > 0` for many `n`. It couples
the three-denominator problem to a cleaner two-denominator divisor problem.

Its boundary: the representation bounds give lower bounds that hold for almost all
primes and for infinitely many `n`, but they degrade exactly for primes that "resemble
a perfect square" to small moduli — the quadratic-residue primes — where the divisor
structure of `4/n − 1/x` does not guarantee a two-fraction representation. The method
covers typical primes, not the survivor class.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "erdos-straus/two-fraction-representation-bounds",
      "label": "Two-unit-fraction representation bounds",
      "description": "Fix one denominator, reduce to representing a rational as two unit fractions, and bound representation counts.",
      "mechanism": {
        "representations": ["two-unit-fraction representations", "divisor counts"],
        "assumptions": ["fixing one denominator is without loss", "divisor counts control existence"],
        "operators": ["reduction to two fractions", "representation counting"],
        "preserves": ["existence via representation lower bounds"],
        "breaks": [],
        "auxiliary_objects": ["representation-count estimates for a/b = 1/y + 1/z"],
        "locality": "global",
        "construction_mode": "existential",
        "uncertainty_mode": "deterministic",
        "notes": "Lower bounds for almost all primes; couples 3-term to 2-term problem."
      },
      "outcome": {
        "class": "partial_success",
        "boundary_statement": "covers almost all primes but degrades for perfect-square-like primes",
        "boundary_conditions": ["quadratic-residue primes", "no guaranteed 2-fraction representation there"],
        "notes": "Almost-all, not all."
      },
      "support": [
        {"field_path": "mechanism.operators", "support_kind": "explicit", "locator": "para:2"},
        {"field_path": "mechanism.preserves", "support_kind": "inferred", "locator": "para:3"},
        {"field_path": "outcome.boundary_statement", "support_kind": "explicit", "locator": "para:4"}
      ]
    }
  ]
}
newf-normalize -->
