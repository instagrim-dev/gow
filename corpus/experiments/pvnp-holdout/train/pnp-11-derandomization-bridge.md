# Derandomization ↔ lower bounds bridge (pnp-11)

**Era:** 2003. Pre-cutoff. Anchor: Kabanets–Impagliazzo, *Derandomizing
polynomial identity tests means proving circuit lower bounds*, STOC 2003.

The train atlas's nearest approach to a direction reversal. The
hardness-vs-randomness program (Nisan–Wigderson and successors) had long run
one way: lower bounds ⇒ pseudorandom generators ⇒ derandomization. Kabanets
and Impagliazzo proved a **converse**: derandomizing polynomial identity
testing (placing PIT in NSUBEXP) implies circuit lower bounds — either NEXP ⊄
P/poly or the permanent lacks polynomial-size arithmetic circuits.

The mechanism is an **algorithm-to-hardness implication**: assume a good
*algorithm* exists (deterministic PIT), combine it with completeness and
hierarchy machinery, and conclude that some explicit function is hard. The
distinguishing property here is not a truth-table property at all — the
argument is a conditional implication between an algorithmic event and a
lower-bound event, evading the natural-proofs frame (nothing large or
constructive is exhibited over random functions).

Its recorded boundary at the cutoff: the bridge is **conditional in the wrong
direction for separation** — it converts an unproven derandomization into an
unproven lower bound, trading one open problem for another. Nobody had
derived an *unconditional* lower bound by exhibiting an actual, unconditional
algorithm and running such a bridge on it. The atlas records the direction
reversal as demonstrated-in-principle, unconsummated-in-fact.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "pvnp/derandomization-lower-bound-bridge",
      "label": "Derandomization-to-lower-bounds bridge",
      "description": "Prove that derandomizing polynomial identity testing implies circuit lower bounds, inverting the classical hardness-to-randomness direction.",
      "mechanism": {
        "representations": ["identity-testing algorithms", "arithmetic circuits", "conditional implications"],
        "assumptions": ["deterministic subexponential PIT exists (hypothesis, not theorem)"],
        "operators": ["algorithm-to-hardness implication", "completeness leverage", "hierarchy-theorem coupling"],
        "preserves": [],
        "breaks": ["black-box relativizing simulation", "large constructive distinguishing property", "model-analysis-first direction"],
        "auxiliary_objects": ["polynomial identity tests", "permanent as hard candidate"],
        "locality": "global",
        "construction_mode": "existential",
        "uncertainty_mode": "deterministic",
        "notes": "First train family to reverse the analysis direction: from an algorithm's existence to a hardness conclusion. Conditional only."
      },
      "outcome": {
        "class": "partial_success",
        "boundary_statement": "the algorithm-to-hardness direction works in principle but only conditionally; no unconditional algorithm was supplied to consummate it",
        "boundary_conditions": ["hypothesis (derandomized PIT) itself open", "no unconditional lower bound extracted at cutoff"],
        "notes": "Direction reversal demonstrated-in-principle, unconsummated-in-fact; the atlas's nearest partial-success structure to the withheld advance."
      },
      "support": [
        {"field_path": "mechanism.operators", "support_kind": "explicit", "locator": "para:2"},
        {"field_path": "mechanism.breaks", "support_kind": "explicit", "locator": "para:2"},
        {"field_path": "outcome.class", "support_kind": "explicit", "locator": "para:3"},
        {"field_path": "outcome.boundary_statement", "support_kind": "explicit", "locator": "para:3"}
      ]
    }
  ]
}
newf-normalize -->
