# The circuit program's stall after natural proofs (pnp-09)

**Era:** 1994–2004. Pre-cutoff. Anchors: Razborov 1995 (*Unprovability of
lower bounds…*, bounded-arithmetic independence); contemporaneous surveys of
the post-natural-proofs decade.

This entry records the decade of **conserved failure** after pnp-08, because
the absence of movement is itself structure. Between 1994 and the cutoff, no
lower bound was proven against any circuit class at or above ACC⁰ for any
function in NEXP, despite the escape space being explicitly mapped
(non-constructive or rare properties).

Attempts in the period conserved the old posture in new vocabulary:
derandomization-flavored properties, extensions of the polynomial method to
composite moduli, and formal-independence results (Razborov 1995: relevant
lower-bound statements are unprovable in certain bounded-arithmetic systems
whose reasoning power mirrors the natural-proofs frame). Each either
re-encountered the largeness/constructivity trap or retreated to restricted
models where the trap does not bind.

The candidate invariant this decade sharpens: **the direction of analysis was
always model→function** — fix the circuit class, seek a property over truth
tables that all small circuits in it respect. Both walls (pnp-02, pnp-08) are
statements about *that direction*. Nothing in the train atlas reverses the
direction — nothing derives circuit structure from the existence of an
*algorithm*. The reversal is precisely what the atlas never tried.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "pvnp/post-natural-proofs-stall",
      "label": "Post-natural-proofs stall (1994-2004)",
      "description": "A decade of attempts against ACC0 and above that either re-encounter the natural-proofs trap or retreat to restricted models; no NEXP lower bound at cutoff.",
      "mechanism": {
        "representations": ["truth-table properties", "restricted circuit models", "bounded-arithmetic systems"],
        "assumptions": ["a property-based route exists that evades largeness or constructivity"],
        "operators": ["property search", "polynomial-method extension", "independence proof"],
        "preserves": ["model-analysis-first direction", "large constructive distinguishing property"],
        "breaks": [],
        "auxiliary_objects": ["bounded-arithmetic fragments"],
        "locality": "local",
        "construction_mode": "existential",
        "uncertainty_mode": "deterministic",
        "notes": "Absence of movement recorded as structure: the model-to-function direction of analysis is conserved across every attempt."
      },
      "outcome": {
        "class": "failure",
        "boundary_statement": "no lower bound against ACC0 or above for any NEXP function in the decade after the natural-proofs barrier",
        "boundary_conditions": ["attempts conserve the model-to-function analysis direction", "mapped escape space (non-constructive or rare properties) unoccupied"],
        "notes": "The atlas-wide conserved property is the analysis direction itself; no train family derives structure from an algorithm's existence."
      },
      "support": [
        {"field_path": "mechanism.preserves", "support_kind": "explicit", "locator": "para:3"},
        {"field_path": "outcome.class", "support_kind": "explicit", "locator": "para:1"},
        {"field_path": "outcome.boundary_statement", "support_kind": "explicit", "locator": "para:1"},
        {"field_path": "outcome.notes", "support_kind": "explicit", "locator": "para:3"}
      ]
    }
  ]
}
newf-normalize -->
