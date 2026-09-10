# Averaged covering-system heuristic (Approach B)

This project-authored note records a very differently worded attempt on the
Erdős–Straus conjecture. Rather than writing per-class identities, it studies
the *density* of representable `n` using an averaged covering-system argument.

The idea is global: instead of treating each residue class in isolation, it
couples many moduli at once and estimates how often a random-looking `n` fails
to be caught by the combined covering system. The reasoning is existential and
probabilistic — it argues that solutions *exist* for almost all `n` without
constructing `x`, `y`, `z` explicitly, and it introduces an auxiliary sieve
weight as a bookkeeping object.

The result is a strong asymptotic statement: the set of exceptional `n` has
density zero. But it is only a partial result for the full conjecture, because
a sparse exceptional set survives the averaging and the method says nothing
constructive about any individual exceptional `n`. Progress stops at that
sparse surviving exceptional set.

Although the prose looks nothing like the modular-identity note, both attempts
ultimately fail on structurally related survivor classes — which is exactly the
comparison the normalized representation should expose.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "erdos-straus/averaged-covering-density",
      "label": "averaged covering-system density argument",
      "description": "Couple many moduli and bound the density of n not caught by the combined covering system.",
      "mechanism": {
        "representations": ["covering systems", "density estimates"],
        "assumptions": ["asymptotic regime", "combined moduli coupling"],
        "operators": ["averaging", "sieve weighting"],
        "preserves": ["residue-structure survivors"],
        "breaks": ["residue locality"],
        "auxiliary_objects": ["auxiliary sieve weight"],
        "locality": "global",
        "construction_mode": "existential",
        "uncertainty_mode": "probabilistic",
        "notes": "Couples moduli; argues existence for almost all n."
      },
      "outcome": {
        "class": "partial_success",
        "boundary_statement": "sparse exceptional set survives averaging",
        "boundary_conditions": ["density-zero exceptional survivors"],
        "notes": "No constructive statement about individual exceptional n."
      },
      "support": [
        {"field_path": "mechanism.locality", "support_kind": "explicit", "locator": "para:2"},
        {"field_path": "mechanism.breaks", "support_kind": "inferred", "locator": "para:2"},
        {"field_path": "outcome.class", "support_kind": "inferred", "locator": "para:3"},
        {"field_path": "outcome.boundary_statement", "support_kind": "explicit", "locator": "para:3"}
      ]
    }
  ]
}
newf-normalize -->
