# Multi-block message modification on full SHA-1 (sha1-03)

**Era:** 2005. Source: `src-wang-yin-yu-2005`.

This is the attempt that first brought full SHA-1 under the birthday bound, and
it does so by **abandoning the single-block frame**. Two structural moves
combine. First, the differential path is split across **two message blocks**: the
first block is allowed to end in a near-collision, and the second cancels the
residual difference. This relaxes the requirement that any one block achieve
everything, which is what made the earlier single-block searches infeasible.
Second, the early rounds are not searched but **solved** — hand-crafted
*message modification* directly forces conformance for the first rounds, so
probabilistic search only pays for the later rounds.

Crucially, the nonlinear portion of the path is **built by hand**. The linearized
codeword machinery supplies the later rounds; a human supplies the first-round
nonlinear characteristic. This division is the attempt's defining feature and its
limiting one.

The reported result is a collision attack on full 80-step SHA-1 with complexity
"less than 2^69 hash operations ... the first attack on the full 80-step SHA-1
with complexity less than the 2^80 theoretical bound."

No collision of full SHA-1 was produced. What was produced is a **real collision
of 58-step SHA-1**, found with fewer than 2^33 hash operations. The paper also
gives 75-round at 2^78 and 70-round at 2^68.

Scored against the objective, this is a **partial success**: it establishes that
full SHA-1 is theoretically breakable and demonstrates the mechanism on a
reduction, but the object the objective asks for does not exist at this point.
The gap between 2^69 as an estimate and a produced collision is a real boundary,
not a formality — it took twelve more years and two further mechanism changes to
close.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "sha1/multiblock-message-modification",
      "label": "Multi-block differential path with hand-crafted message modification",
      "description": "Split the differential path across two message blocks so the first ends in a near-collision and the second cancels the residual, and force early-round conformance directly by hand-crafted message modification rather than by search.",
      "mechanism": {
        "representations": [
          "two-block differential path",
          "near-collision as intermediate state rather than failure",
          "disturbance-vector-derived linear path for later rounds",
          "hand-built nonlinear characteristic for early rounds"
        ],
        "assumptions": [
          "residual difference from block one can be cancelled by block two",
          "early-round conditions can be satisfied deterministically by message modification",
          "hand construction of the nonlinear path is feasible at this scale"
        ],
        "operators": [
          "multi-block path decomposition",
          "message modification",
          "near-collision chaining",
          "manual nonlinear characteristic construction",
          "round reduction for demonstration"
        ],
        "preserves": [
          "linearized-codeword analysis of the later rounds",
          "local-collision substrate",
          "human-directed path construction"
        ],
        "breaks": [
          "single-block collision requirement",
          "search-only conformance for early rounds"
        ],
        "auxiliary_objects": [
          "disturbance vector",
          "message modification condition set",
          "intermediate near-collision state"
        ],
        "locality": "mixed",
        "construction_mode": "constructive",
        "uncertainty_mode": "probabilistic",
        "notes": "The manual construction of the nonlinear path is the attempt's defining and limiting feature: it is what makes the attack work and what caps its reach, since a human cannot search the characteristic space.",
        "completeness_scope": "declared_payload",
        "completeness_basis": "Every representation/assumption/operator/preserves/breaks/auxiliary_object entry this note states is listed here and parsed by the deterministic embedded-payload parser. Authored 2026-09-12 by the agent identified in manifest.json authoring_provenance, from the read_scope recorded for src-wang-yin-yu-2005 (full text). declared_payload scope, NOT a mechanism-exhaustive extraction judgment.",
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
        "class": "partial_success",
        "boundary_statement": "establishes a sub-birthday complexity estimate of 2^69 for full SHA-1 but produces no full-SHA-1 collision; the only produced collision is for 58-step SHA-1",
        "boundary_conditions": [
          "2^69 is an estimate, not an executed computation",
          "produced collision is 58-step, at under 2^33 hash operations",
          "nonlinear early-round characteristic is hand-built and does not scale by search"
        ],
        "notes": "Partial success relative to the declared objective. Closing the estimate-to-artifact gap required twelve further years and two further mechanism changes."
      },
      "support": [
        {
          "field_path": "mechanism.operators",
          "support_kind": "explicit",
          "locator": "para:2"
        },
        {
          "field_path": "mechanism.breaks",
          "support_kind": "inferred",
          "locator": "para:2"
        },
        {
          "field_path": "mechanism.notes",
          "support_kind": "explicit",
          "locator": "para:3"
        },
        {
          "field_path": "outcome.class",
          "support_kind": "inferred",
          "locator": "para:6"
        },
        {
          "field_path": "outcome.boundary_statement",
          "support_kind": "explicit",
          "locator": "para:5"
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
