# Disturbance vector classification (sha1-06)

**Era:** 2008–2011. Source: `src-manuel-2008`.

Previous attempts each selected a disturbance vector and built an attack on it.
This attempt **steps up an abstraction level** and asks what the space of
disturbance vectors looks like. The result is a classification: "all published
disturbance vectors can be classified into two types of vectors, type-I and
type-II." The equivalence is structural — vectors related by cyclic shift of
32-bit words yield the same extended expanded message — so the apparently
distinct vectors used across a decade of attacks collapse into two families.
Type-I contains Wang et al.'s vector; type-II contains Jutla and Patthak's
Codeword2.

This is genuine compression of the search space, and it is the kind of move that
should be valuable: it replaces "pick a vector and try" with "enumerate the
families and evaluate each." The mechanism is *systematic generation plus cost
evaluation* rather than differential-path construction.

Its outcome is where this record earns its place, because the same work exists in
two versions with **different conclusions**, and the difference is not cosmetic.
The ePrint version claimed a theoretical attack complexity of 2^51 hash function
calls, with a disturbance vector carrying a cost evaluation of 2^57. The
published journal version drops the 2^51 claim entirely and concludes instead
that the most efficient disturbance vector is Codeword2 — adding that "the common
assumption of local collision independence is **flawed**." RFC 6194 records the
divergence: the 2^51 claim "is absent from the published conference paper."

Scored against the objective, this is a **partial failure**. The classification
survives and is useful; the cost evaluation that made it look like a frontier
advance did not. And the retracted number did damage before it was corrected:
the 2^57 cost evaluation from this work is precisely what a subsequent paper
built its 2^52 claim on, and that paper was withdrawn when this evaluation was
found incorrect (`sha1-07`).

The corpus records both versions and does **not** reconcile them. The finding
about flawed local-collision independence is arguably the more durable
contribution, and it is an *assumption-level* correction: it says the whole
family had been composing local collisions as though they were independent when
they are not.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "sha1/disturbance-vector-classification",
      "label": "Systematic classification and generation of disturbance vectors",
      "description": "Move up an abstraction level from selecting a disturbance vector to classifying the space of them, showing all published vectors fall into two shift-equivalent families, and evaluating each family's cost.",
      "mechanism": {
        "representations": [
          "disturbance vector space partitioned into type-I and type-II families",
          "shift equivalence over 32-bit words as the classifying relation",
          "cost evaluation per disturbance vector"
        ],
        "assumptions": [
          "published vectors exhaust the useful vector space",
          "cost evaluation per vector predicts attack complexity",
          "local collisions compose independently (later identified by this same work as flawed)"
        ],
        "operators": [
          "equivalence classification",
          "systematic vector generation",
          "cost evaluation",
          "abstraction level raising"
        ],
        "preserves": [
          "linearized local-collision substrate",
          "disturbance-vector-driven attack frame"
        ],
        "breaks": [
          "per-attack ad hoc vector selection",
          "local collision independence assumption"
        ],
        "auxiliary_objects": [
          "type-I and type-II family representatives",
          "Codeword2 disturbance vector",
          "per-vector cost table"
        ],
        "locality": "global",
        "construction_mode": "constructive",
        "uncertainty_mode": "deterministic",
        "notes": "The classification survived; the cost evaluation did not. The two published versions reach different conclusions and are recorded without reconciliation.",
        "completeness_scope": "declared_payload",
        "completeness_basis": "Every representation/assumption/operator/preserves/breaks/auxiliary_object entry this note states is listed here and parsed by the deterministic embedded-payload parser. Authored 2026-09-12 by the agent identified in manifest.json authoring_provenance, from the read_scope recorded for src-manuel-2008 (ePrint abstract, INRIA version, journal abstract). declared_payload scope, NOT a mechanism-exhaustive extraction judgment.",
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
        "class": "partial_failure",
        "boundary_statement": "the two-family classification stands, but the ePrint version's 2^51 complexity claim and its 2^57 per-vector cost evaluation did not survive: the claim is absent from the published version and the cost evaluation was later found incorrect",
        "boundary_conditions": [
          "ePrint 2008/469 and the Designs Codes and Cryptography version reach different conclusions and are not interchangeable",
          "the incorrect 2^57 cost evaluation propagated into a subsequent withdrawn paper",
          "RFC 6194 records that the 2^51 claim is absent from the published conference paper"
        ],
        "notes": "Partial failure relative to the declared objective. The durable contribution may be the assumption-level correction that local collision independence is flawed, which indicts how the whole family had been composing local collisions."
      },
      "support": [
        {
          "field_path": "mechanism.representations",
          "support_kind": "explicit",
          "locator": "para:2"
        },
        {
          "field_path": "mechanism.operators",
          "support_kind": "inferred",
          "locator": "para:3"
        },
        {
          "field_path": "mechanism.breaks",
          "support_kind": "explicit",
          "locator": "para:4"
        },
        {
          "field_path": "outcome.class",
          "support_kind": "inferred",
          "locator": "para:5"
        },
        {
          "field_path": "outcome.boundary_statement",
          "support_kind": "explicit",
          "locator": "para:4"
        },
        {
          "field_path": "outcome.boundary_conditions",
          "support_kind": "explicit",
          "locator": "para:4"
        }
      ]
    }
  ]
}
newf-normalize -->
