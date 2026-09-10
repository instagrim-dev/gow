# Type I / Type II solution classification (es-04)

**Era:** Mordell's two cases; sharpened by Elsholtz & Tao (2011–2013). Pre-cutoff.

Rather than classify by the congruence that produces a solution, this approach
classifies by the **divisibility structure** of the solution itself. For prime `n`,
Elsholtz and Tao show every solution is one of two types:

- **Type I:** exactly one of `x, y, z` is divisible by `n`;
- **Type II:** exactly two of them are divisible by `n`.

(For composite `n`, most solutions on average are of neither type, but for primes
these exhaust the possibilities.) The classification is transversal: it cuts across
the residue-class families and lets one count solutions of each type and relate
them to divisor sums.

The mechanism is mixed in locality (it reasons about global solution structure via
local divisibility) and existential in mode: it characterizes what solutions *must*
look like without constructing them. Its boundary is that a structural
classification of solutions does not by itself force a solution to *exist* for the
hard quadratic-residue primes — it reorganizes the problem rather than closing it.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "erdos-straus/type-i-ii-classification",
      "label": "Type I / Type II solution classification",
      "description": "Classify solutions by how many denominators n divides; for prime n only two types occur.",
      "mechanism": {
        "representations": ["solution divisibility types", "divisor sums"],
        "assumptions": ["n prime restricts to two types"],
        "operators": ["structural classification", "divisibility partition"],
        "preserves": ["solution divisibility invariants"],
        "breaks": ["classification by producing congruence"],
        "auxiliary_objects": ["Type I and Type II solution counts"],
        "locality": "mixed",
        "construction_mode": "existential",
        "uncertainty_mode": "deterministic",
        "notes": "Transversal classification; characterizes rather than constructs."
      },
      "outcome": {
        "class": "partial_success",
        "boundary_statement": "characterizes solution shape but does not force existence for QR primes",
        "boundary_conditions": ["quadratic-residue primes", "existence not implied by classification"],
        "notes": "Reorganizes the problem; foundation for counting and averaging."
      },
      "support": [
        {"field_path": "mechanism.operators", "support_kind": "explicit", "locator": "para:2"},
        {"field_path": "mechanism.representations", "support_kind": "explicit", "locator": "para:2"},
        {"field_path": "outcome.class", "support_kind": "inferred", "locator": "para:4"},
        {"field_path": "outcome.boundary_statement", "support_kind": "explicit", "locator": "para:4"}
      ]
    }
  ]
}
newf-normalize -->
