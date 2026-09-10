# Vaughan-style congruence verification (es-10)

**Era:** Vaughan (1970) and successors. Pre-cutoff.

This approach estimates the number of `n ≤ N` for which the conjecture is *not yet*
settled by known congruence identities, and drives that count down. Vaughan showed
the number of exceptional `n` up to `N` is small (bounded by an expression decaying
like `N · exp(−c (log N)^{2/3})`), by assembling a large battery of congruence
classes that admit identities and bounding the leftover.

The mechanism is local-to-global: it works class by class (local identities) but its
payoff is a **global density** statement about how thin the unsolved set is. It is
deterministic and existential in flavor — it counts the survivors rather than
constructing solutions for them.

Its boundary is the familiar one from a different angle: the exceptional set is thin
but **non-empty and infinite**, and it is exactly the quadratic-residue survivor
family (the `n ≡ 1 (mod 24)` style classes and their refinements). Making the density
smaller never reaches zero, because the surviving classes are protected by the
reciprocity obstruction (es-08). It sharpens "almost all `n`" without closing "all
`n`".

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "erdos-straus/vaughan-congruence-density",
      "label": "Vaughan-style congruence density bound",
      "description": "Bound the density of unsolved n by assembling many identity-bearing congruence classes.",
      "mechanism": {
        "representations": ["congruence classes", "density of exceptional set"],
        "assumptions": ["many classes admit identities", "leftover is boundable"],
        "operators": ["class assembly", "density estimation"],
        "preserves": ["solvability on the assembled classes"],
        "breaks": [],
        "auxiliary_objects": ["exceptional-set density bound"],
        "locality": "mixed",
        "construction_mode": "existential",
        "uncertainty_mode": "deterministic",
        "notes": "Local identities aggregated into a global thinness statement."
      },
      "outcome": {
        "class": "partial_failure",
        "boundary_statement": "exceptional set is thin but infinite and protected by the QR wall",
        "boundary_conditions": ["quadratic-residue survivor classes", "density never reaches zero"],
        "notes": "Sharpens almost-all; does not reach all."
      },
      "support": [
        {"field_path": "mechanism.operators", "support_kind": "explicit", "locator": "para:2"},
        {"field_path": "mechanism.representations", "support_kind": "explicit", "locator": "para:2"},
        {"field_path": "outcome.boundary_statement", "support_kind": "explicit", "locator": "para:4"}
      ]
    }
  ]
}
newf-normalize -->
