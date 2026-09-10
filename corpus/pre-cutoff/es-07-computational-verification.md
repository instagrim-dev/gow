# Large-scale computational verification (es-07)

**Era:** ongoing; verified past `n = 10^17` and beyond by successive computations.
Pre-cutoff.

This approach abandons closed-form coverage and instead **checks the conjecture
directly** for all `n` up to a large bound by search. For each `n` (reduced to primes
via the multiplicative reduction), an algorithm enumerates candidate first
denominators `x` in the admissible range and tests whether `4/n − 1/x` splits into
two unit fractions. Efficient divisor-based search and sieving over residue classes
make the range enormous.

The mechanism is local and deterministic per `n`, and empirical in character: it
produces a certified solution for every `n` in the tested range and thereby rules out
small counterexamples. Recent computations even dispatch historically awkward
"Mordell-exceptional" quadratic-residue primes within the finite range.

Its boundary is definitional, not incidental: a finite verification, however large,
is **not a proof** for all `n`. It cannot certify the infinitely many quadratic-residue
primes beyond the bound. Its research value is as evidence and as a source of
concrete solved instances that constrain what any structural invariant must respect,
not as a closing argument.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "erdos-straus/computational-verification",
      "label": "Large-scale computational verification",
      "description": "Directly search and certify solutions for all n up to a large finite bound.",
      "mechanism": {
        "representations": ["explicit solved instances", "search over denominators"],
        "assumptions": ["finite range is checkable", "multiplicative reduction to primes"],
        "operators": ["bounded search", "sieving", "divisor test"],
        "preserves": ["certified per-instance solutions"],
        "breaks": ["claim of generality"],
        "auxiliary_objects": ["verification range bound", "solution certificates"],
        "locality": "local",
        "construction_mode": "constructive",
        "uncertainty_mode": "deterministic",
        "notes": "Empirical: certifies a finite range, including QR primes within it."
      },
      "outcome": {
        "class": "partial_failure",
        "boundary_statement": "finite verification is not a proof for all n",
        "boundary_conditions": ["infinitely many QR primes beyond the bound"],
        "notes": "Evidence and instance source, not closure."
      },
      "support": [
        {"field_path": "mechanism.operators", "support_kind": "explicit", "locator": "para:2"},
        {"field_path": "mechanism.breaks", "support_kind": "inferred", "locator": "para:4"},
        {"field_path": "outcome.boundary_statement", "support_kind": "explicit", "locator": "para:4"}
      ]
    }
  ]
}
newf-normalize -->
