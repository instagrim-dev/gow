# Monks–Velingker structural solution analysis (es-11)

**Era:** Monks & Velingker and related structural work; used as input to later
congruence-system proposals. Pre-cutoff.

This approach studies the **algebraic structure of the solution set** directly:
relations among the denominators, symmetries of solutions, and how solutions of one
`n` induce solutions of related `n`. It seeks structural constraints that any
solution triple must satisfy and uses them to organize the search and to connect
distinct residue classes through shared solution forms.

The mechanism is mixed in locality and existential in mode: it characterizes the
constraint geometry of solutions rather than constructing them for a fixed `n`. It
preserves the algebraic relations among denominators and is a natural upstream input
to the Type A/B congruence-system program (es-12).

Its boundary is that structural relations, however tight, remain **conjectural about
coverage**: they narrow what solutions look like and link classes, but they do not
prove that every prime admits at least one solution. The hardest quadratic-residue
primes still require an existence argument the structure alone does not supply.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "erdos-straus/monks-velingker-structure",
      "label": "Structural solution-set analysis",
      "description": "Analyze algebraic relations and symmetries among solution denominators to constrain and link classes.",
      "mechanism": {
        "representations": ["solution-set relations", "denominator symmetries"],
        "assumptions": ["structural constraints link residue classes"],
        "operators": ["structural constraint analysis", "symmetry exploitation"],
        "preserves": ["algebraic relations among denominators"],
        "breaks": [],
        "auxiliary_objects": ["induced-solution maps between related n"],
        "locality": "mixed",
        "construction_mode": "existential",
        "uncertainty_mode": "deterministic",
        "notes": "Constraint geometry of solutions; upstream of Type A/B systems."
      },
      "outcome": {
        "class": "partial_success",
        "boundary_statement": "structural constraints do not prove coverage of all primes",
        "boundary_conditions": ["quadratic-residue primes", "coverage remains conjectural"],
        "notes": "Links classes; existence still owed."
      },
      "support": [
        {"field_path": "mechanism.operators", "support_kind": "explicit", "locator": "para:2"},
        {"field_path": "mechanism.preserves", "support_kind": "inferred", "locator": "para:2"},
        {"field_path": "outcome.boundary_statement", "support_kind": "explicit", "locator": "para:4"}
      ]
    }
  ]
}
newf-normalize -->
