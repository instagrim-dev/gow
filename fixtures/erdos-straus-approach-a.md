# Modular residue-cover construction (Approach A)

This project-authored note sketches a **modular residue-cover** attack on the
Erdős–Straus conjecture, that for every integer `n >= 2` the equation
`4/n = 1/x + 1/y + 1/z` admits a solution in positive integers.

The construction fixes a small modulus and partitions candidate `n` into
residue classes. For each admissible class we exhibit an explicit polynomial
identity in `n` that produces a unit-fraction decomposition. Because the
identities are keyed to a fixed modulus, the argument is strictly *local*: it
reasons one residue class at a time and never couples classes together.

The approach is fully constructive and deterministic: given `n` in a covered
class, the identity writes down `x`, `y`, `z` directly. It preserves
residue-locality and fixed-modulus structure throughout.

It is a partial result. A stubborn family of residue classes, notably several
quadratic-residue survivor classes, is left uncovered by any fixed modulus of
the sizes tried here. Progress stops at exactly those uncovered residue
families.

This fixture exists only to demonstrate the normalization shape; it does not
claim new mathematics.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "erdos-straus/modular-residue-cover",
      "label": "modular residue-cover construction",
      "description": "Fix a small modulus, partition n by residue class, and give explicit polynomial identities per admissible class.",
      "mechanism": {
        "representations": ["congruence classes", "polynomial identities"],
        "assumptions": ["finite residue cover", "fixed modulus"],
        "operators": ["modular decomposition", "polynomial identity"],
        "preserves": ["residue locality", "fixed-modulus structure"],
        "breaks": [],
        "auxiliary_objects": [],
        "locality": "local",
        "construction_mode": "constructive",
        "uncertainty_mode": "deterministic",
        "notes": "Per-class explicit decomposition; classes never coupled."
      },
      "outcome": {
        "class": "partial_failure",
        "boundary_statement": "uncovered residue families remain",
        "boundary_conditions": ["quadratic-residue survivor classes"],
        "notes": "No fixed modulus of the tried sizes covers all classes."
      },
      "support": [
        {"field_path": "mechanism.representations", "support_kind": "explicit", "locator": "para:2"},
        {"field_path": "mechanism.preserves", "support_kind": "inferred", "locator": "para:3"},
        {"field_path": "outcome.class", "support_kind": "explicit", "locator": "para:4"},
        {"field_path": "outcome.boundary_statement", "support_kind": "explicit", "locator": "para:4"}
      ]
    }
  ]
}
newf-normalize -->
