# Two partial approaches to Erdős–Straus in one note

This project-authored survey note describes two materially different partial
approaches to the Erdős–Straus conjecture in a single source. It exists to
prove that normalization can emit multiple distinct approaches from one source
snapshot rather than forcing a one-source-one-approach mapping.

## Section 1: local modular identities

The first approach fixes a small modulus and produces explicit polynomial
identities per admissible residue class. It is local, constructive, and
deterministic, and it stalls on stubborn quadratic-residue survivor classes.

## Section 2: global averaged covering

The second approach couples many moduli and estimates the density of `n` not
caught by the combined covering system. It is global, existential, and
probabilistic, and it stalls on a sparse density-zero exceptional set.

The two sections share almost no surface vocabulary, yet both stop on
structurally related survivor classes.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "erdos-straus/modular-residue-cover",
      "label": "modular residue-cover construction",
      "description": "Fix a small modulus and give explicit polynomial identities per admissible residue class.",
      "mechanism": {
        "representations": ["congruence classes", "polynomial identities"],
        "assumptions": ["finite residue cover", "fixed modulus"],
        "operators": ["modular decomposition", "polynomial identity"],
        "preserves": ["residue locality", "fixed-modulus structure"],
        "locality": "local",
        "construction_mode": "constructive",
        "uncertainty_mode": "deterministic"
      },
      "outcome": {
        "class": "partial_failure",
        "boundary_statement": "uncovered residue families remain",
        "boundary_conditions": ["quadratic-residue survivor classes"]
      },
      "support": [
        {"field_path": "mechanism.representations", "support_kind": "explicit", "locator": "section:1"},
        {"field_path": "outcome.class", "support_kind": "explicit", "locator": "section:1"}
      ]
    },
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
        "uncertainty_mode": "probabilistic"
      },
      "outcome": {
        "class": "partial_success",
        "boundary_statement": "sparse exceptional set survives averaging",
        "boundary_conditions": ["density-zero exceptional survivors"]
      },
      "support": [
        {"field_path": "mechanism.locality", "support_kind": "explicit", "locator": "section:2"},
        {"field_path": "outcome.boundary_statement", "support_kind": "explicit", "locator": "section:2"}
      ]
    }
  ]
}
newf-normalize -->
