# Time-space tradeoffs for SAT — indirect diagonalization (pnp-12)

**Era:** 1997–2004. Pre-cutoff. Anchors: Fortnow 1997 (nondeterministic
polynomial time vs small-space); Fortnow–van Melkebeek 2000 (time-space
tradeoffs for satisfiability).

The period's only *unconditional* lower bounds touching SAT itself:
satisfiability cannot be solved simultaneously in time `n^c` and space `n^{o(1)}`
for small constants `c` (e.g. `c < φ` and successive improvements). The method
is **indirect diagonalization**: assume SAT is easy in both time and space,
use completeness to translate the assumption into unlikely inclusions between
nondeterministic and alternating time classes, speed up alternations under
the space bound, and contradict a hierarchy theorem.

Mechanistically this family is a hybrid ancestor: it runs an
**assumption-to-contradiction pipeline through hierarchy theorems** — assume
an algorithm exists, derive class collapses, contradict a hierarchy — which is
the same skeleton the pnp-11 bridge uses, and it is unconditional. But the
assumed algorithm is refuted rather than supplied: the argument consumes a
*hypothetical* algorithm to power a contradiction, and its strength is capped
by how much work the hierarchy theorems can do, yielding bounds that are
model-specific (simultaneous time-space) and quantitatively weak (small
polynomial exponents).

Boundary at cutoff: decades of refinement moved the exponent slightly; the
method showed no path to super-polynomial bounds or to separating P from NP.
Its structural contribution to the atlas is the working, unconditional
skeleton: algorithmic assumption → completeness translation → hierarchy
contradiction.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "pvnp/sat-time-space-tradeoffs",
      "label": "SAT time-space tradeoffs (indirect diagonalization)",
      "description": "Unconditional lower bounds against simultaneous time-space algorithms for SAT via assumption-to-contradiction pipelines through hierarchy theorems.",
      "mechanism": {
        "representations": ["alternating time classes", "simultaneous time-space machines", "completeness translations"],
        "assumptions": ["hypothetical fast low-space SAT algorithm (for contradiction)"],
        "operators": ["indirect diagonalization", "alternation speedup", "hierarchy contradiction", "completeness translation"],
        "preserves": ["black-box relativizing simulation"],
        "breaks": ["large constructive distinguishing property"],
        "auxiliary_objects": ["alternation hierarchy", "clocked simulations"],
        "locality": "global",
        "construction_mode": "constructive",
        "uncertainty_mode": "deterministic",
        "notes": "Unconditional skeleton: assumed algorithm -> class collapse -> hierarchy contradiction. Assumed algorithm is refuted, never supplied."
      },
      "outcome": {
        "class": "partial_success",
        "boundary_statement": "unconditional but model-specific and quantitatively weak; small-exponent time-space bounds only, no route to superpolynomial separation",
        "boundary_conditions": ["strength capped by hierarchy-theorem leverage", "bounds tied to the simultaneous time-space model"],
        "notes": "Contributes the pipeline skeleton the pnp-11 bridge shares; largely relativizing, hence also wall-bounded."
      },
      "support": [
        {"field_path": "mechanism.operators", "support_kind": "explicit", "locator": "para:1"},
        {"field_path": "mechanism.notes", "support_kind": "explicit", "locator": "para:2"},
        {"field_path": "outcome.class", "support_kind": "explicit", "locator": "para:1"},
        {"field_path": "outcome.boundary_statement", "support_kind": "explicit", "locator": "para:3"}
      ]
    }
  ]
}
newf-normalize -->
