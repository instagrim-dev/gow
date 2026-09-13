# Automated differential characteristic search (sha1-04)

**Era:** 2006. Source: `src-decanniere-rechberger-2006`.

The 2005 attack's nonlinear early-round characteristic was built by hand. This
attempt removes the human from that step. The authors state the target plainly:
"the complex characteristics needed for the recent collision attacks on members
of the SHA family have been constructed manually by Wang et al. In this report,
we describe a method to search for them automatically."

The mechanism replaces manual construction with a **generalized-condition
representation plus guess-and-determine search**. Instead of committing to a
concrete bit pattern, each bit position carries a *set* of admissible
difference-and-value combinations; the search then repeatedly picks an
undetermined position, guesses a restriction, and propagates the consequences,
backtracking when a contradiction appears. Path construction becomes a
constraint-propagation problem over an under-determined object rather than a feat
of insight.

The concrete result is a two-block collision for **64-step SHA-1**, at an expected
work factor of about 2^35 compression function evaluations — against a prior
published best of 58 steps. This is a genuine, checkable artifact.

Relative to the objective this is a **partial success**, and its boundary is
unusually clear: automating characteristic search did **not** lower the complexity
of the attack on full SHA-1. The mechanism removed a labor bottleneck without
moving the cost frontier, which locates the real obstruction elsewhere — in the
probability of the later-round path, not in the difficulty of writing down the
early-round one. The method was later extended to 70 steps by the same authors
and to 72 and 73 steps by others, all short of 80.

There is a methodological point worth recording. This paper is also the
authoritative source for the fact that the widely-cited 2^63 improvement was
never published: it notes that for those "announced but ... unpublished
improvements ... no message modification costs are given, thus we lack
comparability here." An attempt that cannot be compared cannot be built on.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "sha1/automated-characteristic-search",
      "label": "Automated differential characteristic search over generalized conditions",
      "description": "Replace manual construction of the nonlinear early-round characteristic with automated guess-and-determine search over per-bit generalized conditions, propagating restrictions and backtracking on contradiction.",
      "mechanism": {
        "representations": [
          "generalized conditions: per-bit sets of admissible difference-value combinations",
          "differential characteristic as partially determined constraint object",
          "search tree over bit-position restrictions"
        ],
        "assumptions": [
          "characteristic construction is the binding constraint on attack reach",
          "local condition propagation approximates global path validity",
          "guess-and-determine search explores the useful characteristic space"
        ],
        "operators": [
          "generalized-condition encoding",
          "guess-and-determine search",
          "condition propagation",
          "backtracking on contradiction",
          "automated characteristic generation"
        ],
        "preserves": [
          "two-block collision structure",
          "local-collision substrate",
          "differential-path attack frame"
        ],
        "breaks": [
          "manual construction of the nonlinear characteristic",
          "insight-bound path discovery"
        ],
        "auxiliary_objects": [
          "generalized-condition table",
          "search tree",
          "64-step colliding message pair"
        ],
        "locality": "local",
        "construction_mode": "constructive",
        "uncertainty_mode": "deterministic",
        "notes": "Removed a labor bottleneck without moving the cost frontier for full SHA-1, which relocates the obstruction from characteristic construction to later-round path probability.",
        "completeness_scope": "declared_payload",
        "completeness_basis": "Every representation/assumption/operator/preserves/breaks/auxiliary_object entry this note states is listed here and parsed by the deterministic embedded-payload parser. Authored 2026-09-12 by the agent identified in manifest.json authoring_provenance, from the read_scope recorded for src-decanniere-rechberger-2006 (full text). declared_payload scope, NOT a mechanism-exhaustive extraction judgment.",
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
        "boundary_statement": "produces a real 64-step SHA-1 collision at about 2^35 evaluations but does not lower the complexity of the attack on full SHA-1",
        "boundary_conditions": [
          "automation addressed characteristic construction, not later-round path probability",
          "reach extended to 64 steps, later 70 by the same authors and 72-73 by others, all short of 80",
          "expected work factor about 2^35 compression function evaluations"
        ],
        "notes": "Partial success relative to the declared objective. The negative half of the result is the informative half: it shows the binding constraint was misidentified."
      },
      "support": [
        {
          "field_path": "mechanism.operators",
          "support_kind": "explicit",
          "locator": "para:3"
        },
        {
          "field_path": "mechanism.breaks",
          "support_kind": "explicit",
          "locator": "para:2"
        },
        {
          "field_path": "outcome.class",
          "support_kind": "explicit",
          "locator": "para:4"
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
