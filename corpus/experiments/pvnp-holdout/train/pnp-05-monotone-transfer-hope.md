# The monotone→general hope and its failure (pnp-05)

**Era:** 1985–1988. Pre-cutoff. Anchors: Razborov 1985 (perfect matching has
superpolynomial monotone complexity); Tardos 1988 (exponential monotone/general
gap, Combinatorica).

After pnp-04, the natural program was **transfer**: if monotone lower bounds
can reach NP-complete functions, perhaps monotone complexity tracks general
complexity closely enough that monotone bounds for the right function separate
P from NP.

The hope failed on a concrete counterexample family. Razborov (1985) proved
superpolynomial monotone lower bounds for **perfect matching** — a problem in
P. Tardos (1988) sharpened the separation to exponential: a monotone function
computable by polynomial-size general circuits requires exponential-size
monotone circuits. Monotone complexity and general complexity are therefore
**exponentially divorced**, and no bound proven inside the negation-free model
transfers to the unrestricted question.

The failure's structure is informative: the approximation method's power came
from an auxiliary restriction (no negations) that also *severed the bridge
back* to the real question. The conserved posture — prove hardness inside an
analyzable restricted model, then hope the restriction is inessential — is
recorded here as failing at the transfer step, with the restriction itself as
the boundary.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "pvnp/monotone-transfer-hope",
      "label": "Monotone-to-general transfer hope",
      "description": "Leverage monotone lower bounds into general-circuit separations by treating the negation-free restriction as inessential.",
      "mechanism": {
        "representations": ["monotone circuits", "general circuits", "slice/transfer arguments"],
        "assumptions": ["monotone complexity approximates general complexity on monotone functions"],
        "operators": ["restriction lifting", "complexity transfer"],
        "preserves": ["model-analysis-first direction", "large constructive distinguishing property"],
        "breaks": [],
        "auxiliary_objects": ["perfect matching function", "Tardos function"],
        "locality": "local",
        "construction_mode": "constructive",
        "uncertainty_mode": "deterministic",
        "notes": "The transfer step, not the lower-bound step, is where the program dies."
      },
      "outcome": {
        "class": "failure",
        "boundary_statement": "monotone and general complexity are exponentially separated, so negation-free bounds do not transfer",
        "boundary_conditions": ["matching in P has superpolynomial monotone complexity", "exponential monotone/general gap exists"],
        "notes": "The auxiliary restriction that made analysis possible also severed the bridge back to P vs NP."
      },
      "support": [
        {"field_path": "mechanism.assumptions", "support_kind": "explicit", "locator": "para:1"},
        {"field_path": "outcome.class", "support_kind": "explicit", "locator": "para:2"},
        {"field_path": "outcome.boundary_statement", "support_kind": "explicit", "locator": "para:2"},
        {"field_path": "outcome.notes", "support_kind": "explicit", "locator": "para:3"}
      ]
    }
  ]
}
newf-normalize -->
