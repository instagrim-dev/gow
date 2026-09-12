# Monotone circuit lower bounds — the approximation method (pnp-04)

**Era:** 1985. Anchor: Razborov 1985 (superpolynomial monotone
lower bounds for clique).

Razborov's **approximation method** attacks a restricted but natural model:
monotone circuits (AND/OR only, no negation). Replace each gate of a purported
small monotone circuit for CLIQUE with an "approximating" gate drawn from a
structured lattice of simple set systems (sunflower-plucked positive
combinations). Each replacement introduces few errors on the chosen test
inputs; a small circuit therefore computes something close to a simple
function — but CLIQUE is provably far from every simple function on those test
distributions. The contradiction yields superpolynomial (later exponential)
monotone lower bounds for an NP-complete function.

The mechanistic posture is inherited from pnp-03 and sharpened: **model
analysis first** — fix the circuit class, build a *distinguishing property*
(approximability by the lattice of simple functions), show every small circuit
in the class has it and the hard function does not. The property is again
broadly shared (most monotone functions are hard this way) and certifiable by
combinatorial computation.

The result was the strongest concrete evidence yet that circuit analysis could
reach NP-completeness. Its limitation surfaced immediately — pnp-05 records
the failed hope of removing the monotonicity restriction.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "pvnp/monotone-approximation-method",
      "label": "Monotone approximation method",
      "description": "Superpolynomial monotone lower bounds for clique by gate-wise approximation within a lattice of simple set systems.",
      "mechanism": {
        "representations": ["monotone circuits", "lattice of approximator functions", "sunflower systems"],
        "assumptions": ["gate-wise approximation errors stay controllable", "test distributions separate hard function from approximators"],
        "operators": ["gate-by-gate approximation", "sunflower lemma", "error counting on test inputs"],
        "preserves": ["model-analysis-first direction", "large constructive distinguishing property"],
        "breaks": ["black-box relativizing simulation"],
        "auxiliary_objects": ["approximator lattice", "positive/negative test distributions"],
        "locality": "local",
        "construction_mode": "constructive",
        "uncertainty_mode": "deterministic",
        "notes": "Distinguishing property: approximability by simple functions; broadly shared and combinatorially certifiable."
      },
      "outcome": {
        "class": "partial_success",
        "boundary_statement": "superpolynomial monotone lower bounds for an NP-complete function; silent on general circuits",
        "boundary_conditions": ["negation-free model only", "approximator lattice exploits monotonicity essentially"],
        "notes": "Strongest evidence to date that circuit analysis reaches NP-hardness; the monotone restriction is load-bearing (pnp-05)."
      },
      "support": [
        {"field_path": "mechanism.operators", "support_kind": "explicit", "locator": "para:1"},
        {"field_path": "mechanism.preserves", "support_kind": "explicit", "locator": "para:2"},
        {"field_path": "outcome.class", "support_kind": "explicit", "locator": "para:1"},
        {"field_path": "outcome.boundary_statement", "support_kind": "explicit", "locator": "para:3"}
      ]
    }
  ]
}
newf-normalize -->
