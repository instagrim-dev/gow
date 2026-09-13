# Joint local-collision analysis (sha1-09)

**Era:** 2013. Source: `src-stevens-2013`.

Every prior attempt in this corpus composed local collisions while treating them
as approximately independent, patching the resulting error with **heuristic
corrections**. Manuel had already identified that assumption as flawed
(`sha1-06`). This attempt removes it properly. Joint local-collision analysis
computes, for a given set of *dependent* local collisions, "the theoretical
maximum success probability ... as well as the smallest set of message conditions
that attains this probability."

The mechanism change is a shift from **estimate to optimum**. Prior work asked
"what does this path cost, approximately?" and answered with a product of
per-step probabilities plus a fudge factor. JLCA asks "what is the best
achievable probability over this whole dependent set, and what minimal condition
set achieves it?" — turning cost analysis into an exact optimization over the
joint structure rather than an approximation over parts. The dependency that
earlier work treated as noise becomes the object of analysis.

The results: an **implemented, open-source** near-collision attack at complexity
equivalent to 2^57.5 SHA-1 compressions, plus attack estimates of approximately
2^61 for identical-prefix collision and 2^77.1 for chosen-prefix collision. The
2^77.1 figure is the first chosen-prefix collision attack on SHA-1 — a
strictly stronger capability than the base objective, and the baseline that later
work beat.

Relative to the objective this is a **partial success**. The near-collision
attack is implemented and open, which is a real artifact; the collision figures
are estimates. No SHA-1 collision was produced.

The boundary worth recording is the split between what was implemented and what
was estimated. Only the near-collision attack was run. This distinction is
exactly what the two records that follow test: `sha1-10` executes a freestart
collision and finds the theoretical estimate optimistic, and `sha1-11` executes
the real thing at 2^63.1 against a 2^61 estimate. An estimate in this domain is a
lower bound on reality, and the gap is measurable.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "sha1/joint-local-collision-analysis",
      "label": "Optimal joint local-collision analysis",
      "description": "Replace heuristic dependency corrections with exact optimization over sets of dependent local collisions, computing the maximum achievable success probability and the smallest message-condition set attaining it.",
      "mechanism": {
        "representations": [
          "set of dependent local collisions as a joint probabilistic object",
          "message conditions as a minimizable set",
          "attack cost as an optimum rather than an estimate"
        ],
        "assumptions": [
          "local collision dependencies are tractable jointly",
          "the minimal condition set attaining maximum probability is computable",
          "compression-equivalent complexity units are comparable across attacks"
        ],
        "operators": [
          "joint probability optimization",
          "dependency-exact cost analysis",
          "message condition set minimization",
          "near-collision attack implementation"
        ],
        "preserves": [
          "local collision substrate",
          "differential path attack frame",
          "multi-block collision structure"
        ],
        "breaks": [
          "local collision independence assumption",
          "heuristic dependency correction",
          "per-part probability approximation"
        ],
        "auxiliary_objects": [
          "joint local-collision probability model",
          "open-source near-collision attack implementation"
        ],
        "locality": "mixed",
        "construction_mode": "constructive",
        "uncertainty_mode": "probabilistic",
        "notes": "Turns the dependency that earlier work treated as noise into the object of analysis, moving cost from approximation over parts to optimization over the joint structure.",
        "completeness_scope": "declared_payload",
        "completeness_basis": "Every representation/assumption/operator/preserves/breaks/auxiliary_object entry this note states is listed here and parsed by the deterministic embedded-payload parser. Authored 2026-09-12 by the agent identified in manifest.json authoring_provenance, from the read_scope recorded for src-stevens-2013 (abstract and IACR-hosted PDF). declared_payload scope, NOT a mechanism-exhaustive extraction judgment.",
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
        "boundary_statement": "implements an open-source near-collision attack at 2^57.5 compression-equivalents and estimates identical-prefix collision at about 2^61 and chosen-prefix at about 2^77.1, but produces no SHA-1 collision",
        "boundary_conditions": [
          "only the near-collision attack was implemented; the collision figures are estimates",
          "2^77.1 is the first chosen-prefix collision attack on SHA-1, a stronger capability than the base objective",
          "later executed work shows estimates in this domain run optimistic relative to realized cost"
        ],
        "notes": "Partial success relative to the declared objective. The implemented-versus-estimated split is directly tested by the two records that follow."
      },
      "support": [
        {
          "field_path": "mechanism.breaks",
          "support_kind": "explicit",
          "locator": "para:2"
        },
        {
          "field_path": "mechanism.operators",
          "support_kind": "explicit",
          "locator": "para:3"
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
          "locator": "para:6"
        }
      ]
    }
  ]
}
newf-normalize -->
