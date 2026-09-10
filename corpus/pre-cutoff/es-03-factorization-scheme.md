# Factorization scheme (es-03)

**Era:** classical parameterization, refined through the modern literature. Pre-cutoff.

Starting from `4/n = 1/x + 1/y + 1/z` and clearing denominators, one seeks
solutions through an identity of the form

```text
(γA − c)(γB − c) = c²
```

which parameterizes candidate denominators nonlinearly in `n`. Solving requires
**enumerating divisors** of a quantity depending on `n` and applying local
congruence filters to keep the denominators positive integers.

This is a local, constructive mechanism: for a given `n` it produces concrete
`(x, y, z)` when a suitable divisor factorization exists, and the arithmetic is
deterministic. It is useful for exposing the local divisor structure of solutions
and connects directly to the Type I/Type II divisibility classification (es-04).

Its boundary is organizational rather than computational: the factorization view
gives no **global lattice organization** of the solution set. Whether a suitable
factorization exists for every `n` is exactly as hard as the conjecture; the
scheme does not force existence for the quadratic-residue primes.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "erdos-straus/factorization-scheme",
      "label": "Factorization scheme (γA−c)(γB−c)=c²",
      "description": "Nonlinear divisor-enumeration parameterization of denominators with local congruence filtering.",
      "mechanism": {
        "representations": ["divisor factorizations", "nonlinear parameterization"],
        "assumptions": ["suitable divisor exists", "positivity via local filters"],
        "operators": ["divisor enumeration", "local congruence filter", "factorization identity"],
        "preserves": ["local divisor structure"],
        "breaks": [],
        "auxiliary_objects": ["divisor lattice of an n-dependent quantity"],
        "locality": "local",
        "construction_mode": "constructive",
        "uncertainty_mode": "deterministic",
        "notes": "Per-n divisor search; no global organization of the solution set."
      },
      "outcome": {
        "class": "partial_failure",
        "boundary_statement": "no global lattice organization; existence unforced for QR primes",
        "boundary_conditions": ["quadratic-residue primes", "divisor search may fail to certify"],
        "notes": "Exposes local structure but does not globalize."
      },
      "support": [
        {"field_path": "mechanism.operators", "support_kind": "explicit", "locator": "para:2"},
        {"field_path": "mechanism.auxiliary_objects", "support_kind": "inferred", "locator": "para:2"},
        {"field_path": "outcome.boundary_statement", "support_kind": "explicit", "locator": "para:4"}
      ]
    }
  ]
}
newf-normalize -->
