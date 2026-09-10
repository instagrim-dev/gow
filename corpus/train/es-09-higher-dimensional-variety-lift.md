# Higher-dimensional variety lift (es-09)

**Era:** Elsholtz & Tao (2013). Pre-cutoff.

Many of the classical parameterizations (Mordell identities, factorization schemes,
Type I/II forms) look unrelated on the surface. This approach **lifts** the
two-dimensional Erdős–Straus surface into a higher-dimensional algebraic variety on
which those parameterizations appear as different rational slices of a single object.
Unifying them this way lets one transport counting and existence arguments uniformly
instead of re-deriving each family.

The mechanism is global, algebraic, and existential: it re-represents the problem at
a higher abstraction level to expose shared structure. It preserves the solution sets
of the individual parameterizations while abstracting away their surface differences.

Its boundary is the abstraction-safety limit the project cares about: the lift
compresses many methods into one, but the compression does **not** add positivity —
it does not force a solution to exist for every prime, and in particular does not
cross the quadratic-residue wall. The unified variety is a better *organization* of
known mechanisms, not a new source of existence. It is a clean test case for whether
a re-representation "earns search authority" by improving transfer without losing
predictive discrimination.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "erdos-straus/higher-dimensional-variety-lift",
      "label": "Higher-dimensional variety lift",
      "description": "Lift known parameterizations onto one higher-dimensional variety to unify counting and existence arguments.",
      "mechanism": {
        "representations": ["higher-dimensional variety", "rational slices"],
        "assumptions": ["parameterizations are slices of one object", "lift preserves solution sets"],
        "operators": ["algebraic lift", "re-representation", "unification"],
        "preserves": ["solution sets of unified parameterizations"],
        "breaks": ["surface distinctions between parameterizations"],
        "auxiliary_objects": ["unifying variety"],
        "locality": "global",
        "construction_mode": "existential",
        "uncertainty_mode": "deterministic",
        "notes": "Abstraction/transfer move; compression without added positivity."
      },
      "outcome": {
        "class": "partial_success",
        "boundary_statement": "unifies methods but adds no positivity for QR primes",
        "boundary_conditions": ["quadratic-residue primes", "existence not gained by lifting"],
        "notes": "Better organization, same wall."
      },
      "support": [
        {"field_path": "mechanism.operators", "support_kind": "explicit", "locator": "para:2"},
        {"field_path": "mechanism.breaks", "support_kind": "inferred", "locator": "para:2"},
        {"field_path": "outcome.boundary_statement", "support_kind": "explicit", "locator": "para:4"}
      ]
    }
  ]
}
newf-normalize -->
