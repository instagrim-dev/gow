# Differential collisions on SHA-0 (sha1-01)

**Era:** 1998. Source: `src-chabaud-joux-1998`.

The founding attack on the SHA family treats the compression function as a
**linear code over local collisions**. A local collision is a small
perturbation injected into one step and cancelled over the following few steps;
the message expansion is linearized, disturbance patterns are treated as
codewords, and a collision becomes a search for a low-weight codeword whose
corrections can all be satisfied simultaneously. The mechanism is *analytic*:
the attacker reasons about the function's algebraic structure rather than
searching message space.

Against SHA-0 this succeeds as theory. The paper reports a theoretical attack on
the SHA-0 **compression function** at complexity 2^61, well under the 2^80
birthday bound. No collision was produced; the result is a complexity estimate.

Against SHA-1 the same mechanism fails, and the authors say so: "In the case of
SHA-1, this method is unable to find collisions faster than the birthday
paradox." The single added rotation in SHA-1's message expansion destroys the
low-weight structure the method depends on — perturbations propagate instead of
staying confined, so the codewords the search needs do not exist at usable
weight. Relative to the objective of colliding full SHA-1, this attempt
**fails**, and it fails informatively: it locates the exact feature that
resists.

A scope caution belongs in the record. The 2^61 figure is stated by the abstract
for the compression function. And while the SHA-0/SHA-1 delta *is* a rotation in
the message expansion, this paper's abstract attributes SHA-1's resistance only
to "the transition to version 1" — the mechanism-level attribution is
established by later work, not by this one, and is recorded here as inferred.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "sha1/linear-code-local-collisions",
      "label": "Differential collisions via linearized local-collision codewords",
      "description": "Attack the SHA compression function by linearizing the message expansion, treating disturbance patterns as codewords of a linear code, and searching for a low-weight codeword whose local-collision corrections are simultaneously satisfiable.",
      "mechanism": {
        "representations": [
          "linearized message expansion over GF(2)",
          "disturbance vector as codeword of a linear code",
          "local collision as cancelled six-step perturbation"
        ],
        "assumptions": [
          "the nonlinear step functions can be approximated linearly for path-search purposes",
          "low codeword weight implies low collision search cost",
          "local collisions compose independently"
        ],
        "operators": [
          "linearization of the step function",
          "low-weight codeword search",
          "local-collision construction",
          "differential path assembly"
        ],
        "preserves": [
          "algebraic analysis of the compression function",
          "single-block differential structure"
        ],
        "breaks": [],
        "auxiliary_objects": [
          "linear code of expanded message differences",
          "local-collision template"
        ],
        "locality": "local",
        "construction_mode": "constructive",
        "uncertainty_mode": "deterministic",
        "notes": "The mechanism is analytic rather than search-driven: it reasons about algebraic structure. Its dependence on low-weight codewords is exactly what SHA-1's added expansion rotation removes.",
        "completeness_scope": "declared_payload",
        "completeness_basis": "Every representation/assumption/operator/preserves/breaks/auxiliary_object entry this note states is listed here and parsed by the deterministic embedded-payload parser. Authored 2026-09-12 by the agent identified in manifest.json authoring_provenance, from the read_scope recorded for src-chabaud-joux-1998 (abstract only). declared_payload scope, NOT a mechanism-exhaustive extraction judgment.",
        "field_completeness": {
          "representation": "complete",
          "assumption": "complete",
          "operator": "complete",
          "preserves": "complete",
          "breaks": "complete",
          "auxiliary_object": "complete"
        }
      },
      "outcome": {
        "class": "failure",
        "boundary_statement": "the method cannot beat the birthday bound on SHA-1, because the added rotation in SHA-1's message expansion destroys the low-weight codeword structure it requires",
        "boundary_conditions": [
          "applies to SHA-0's unrotated message expansion",
          "SHA-1's single-bit expansion rotation is present",
          "2^61 figure is scoped to the compression function, not the full hash"
        ],
        "notes": "Scored against the declared objective (collide full SHA-1) this is a failure. On SHA-0 it is a theoretical success at 2^61 with no produced collision. The failure is informative: it identifies the resisting feature."
      },
      "support": [
        {
          "field_path": "mechanism.operators",
          "support_kind": "explicit",
          "locator": "para:2"
        },
        {
          "field_path": "mechanism.representations",
          "support_kind": "inferred",
          "locator": "para:2"
        },
        {
          "field_path": "outcome.class",
          "support_kind": "explicit",
          "locator": "para:4"
        },
        {
          "field_path": "outcome.boundary_statement",
          "support_kind": "inferred",
          "locator": "para:4"
        },
        {
          "field_path": "outcome.boundary_conditions",
          "support_kind": "explicit",
          "locator": "para:5"
        }
      ]
    }
  ]
}
newf-normalize -->
