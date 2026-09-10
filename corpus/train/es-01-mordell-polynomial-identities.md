# Mordell polynomial identities (es-01)

**Era:** classical (Mordell, *Diophantine Equations*, 1969); systematized through
the 20th century. Pre-cutoff.

For many residue classes of `n`, a fixed **polynomial identity** in `n` writes
down `4/n = 1/x + 1/y + 1/z` directly. The canonical example is `n ≡ 2 (mod 3)`:

```text
4/n = 1/n + 1/((n+1)/3) + 1/(n(n+1)/3)
```

Mordell listed identities covering `n ≡ 2 (mod 3)`, `3 (mod 4)`, `2 or 3 (mod 5)`,
`3, 5, 6 (mod 7)`, and `5 (mod 8)`. Each identity is local to a residue class and
deterministic: given `n` in a covered class, the denominators are computed by the
identity with no search.

Because these identities are keyed to fixed moduli, they reason one class at a
time and never couple classes. They resolve every class **except** those where
`n` behaves like a perfect square to small moduli. The residues left uncovered by
all such identities reduce to `n ≡ 1 (mod 24)` for the small-modulus family, and
more fundamentally to the quadratic-residue classes (see es-08). Progress stops
exactly at those survivor classes.

This note summarizes a known mechanism family; it is not new mathematics.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "erdos-straus/mordell-polynomial-identities",
      "label": "Mordell polynomial identities",
      "description": "Fixed polynomial identities in n give explicit unit-fraction decompositions per residue class.",
      "mechanism": {
        "representations": ["residue classes", "polynomial identities"],
        "assumptions": ["fixed small modulus", "class-local reasoning"],
        "operators": ["polynomial identity", "modular decomposition"],
        "preserves": ["residue locality", "fixed-modulus structure"],
        "breaks": [],
        "auxiliary_objects": [],
        "locality": "local",
        "construction_mode": "constructive",
        "uncertainty_mode": "deterministic",
        "notes": "Per-class explicit denominators; classes never coupled."
      },
      "outcome": {
        "class": "partial_failure",
        "boundary_statement": "quadratic-residue survivor classes remain uncovered",
        "boundary_conditions": ["n a quadratic residue to small moduli", "n = 1 mod 24 survivors"],
        "notes": "No fixed identity family covers the quadratic-residue classes."
      },
      "support": [
        {"field_path": "mechanism.operators", "support_kind": "explicit", "locator": "para:2"},
        {"field_path": "mechanism.preserves", "support_kind": "inferred", "locator": "para:4"},
        {"field_path": "outcome.class", "support_kind": "explicit", "locator": "para:4"},
        {"field_path": "outcome.boundary_statement", "support_kind": "explicit", "locator": "para:4"}
      ]
    }
  ]
}
newf-normalize -->
