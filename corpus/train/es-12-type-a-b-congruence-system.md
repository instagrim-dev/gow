# Type A / Type B congruence system (es-12)

**Era:** *A Complete Congruence System for the Erdős–Straus Conjecture*
(arXiv:2404.01508, April 2024). Pre-cutoff.

Building on Elsholtz–Tao and Monks–Velingker, this approach classifies solutions
**by their form** rather than by the congruence that produces them, defining two new
solution forms — **Type A** and **Type B** — and associating a congruence and a
general polynomial to each. It then proposes a system of congruences that (it
conjectures) always yields a Type A or Type B solution and, again conjecturally,
covers all prime numbers.

The mechanism is global and constructive-per-form: each form comes with an explicit
polynomial, and by construction none of the admitted congruences is a quadratic
residue (respecting Mordell's obstruction, es-08). It couples the transversal
solution-form view with the covering-system goal (es-02), trying to reach coverage
without hitting the reciprocity wall by choosing forms cleverly.

Its boundary is honest and self-declared: it is **conjectural** that every prime has
a Type A or Type B solution. Either a counterexample prime exists (lacking both
forms), or the proposal is an alternative formulation "as difficult or more difficult
to prove than the original". It advances the organization of the problem but does not
close it.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "erdos-straus/type-a-b-congruence-system",
      "label": "Type A / Type B congruence system",
      "description": "Classify solutions by form (Type A/B), attach a polynomial and congruence to each, and propose a covering system avoiding quadratic residues.",
      "mechanism": {
        "representations": ["Type A/B solution forms", "congruence system", "general polynomials"],
        "assumptions": ["forms cover all primes", "admitted congruences avoid quadratic residues"],
        "operators": ["form classification", "polynomial association", "congruence covering"],
        "preserves": ["Mordell non-residue restriction", "polynomial-identity solvability per form"],
        "breaks": ["classification by producing congruence"],
        "auxiliary_objects": ["Type A congruence+polynomial", "Type B congruence+polynomial"],
        "locality": "global",
        "construction_mode": "constructive",
        "uncertainty_mode": "deterministic",
        "notes": "Couples transversal form view with the covering goal."
      },
      "outcome": {
        "class": "partial_success",
        "boundary_statement": "coverage of all primes by Type A/B is conjectural",
        "boundary_conditions": ["a prime lacking both Type A and Type B would be a counterexample", "may be as hard as the original"],
        "notes": "Self-declared conjectural completeness."
      },
      "support": [
        {"field_path": "mechanism.representations", "support_kind": "explicit", "locator": "para:2"},
        {"field_path": "mechanism.preserves", "support_kind": "explicit", "locator": "para:3"},
        {"field_path": "outcome.boundary_statement", "support_kind": "explicit", "locator": "para:4"},
        {"field_path": "outcome.class", "support_kind": "inferred", "locator": "para:4"}
      ]
    }
  ]
}
newf-normalize -->
