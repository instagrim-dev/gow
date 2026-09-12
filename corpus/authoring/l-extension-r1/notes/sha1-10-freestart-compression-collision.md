# Freestart collision on the full compression function (sha1-10)

**Era:** 2015–2016. Source: `src-stevens-karpman-peyrin-2016`.

Every prior attempt aimed at the objective as stated and fell short of it. This
attempt **weakens the objective deliberately** and hits the weakened version
exactly. A freestart collision drops the requirement that the computation begin
at SHA-1's standard initial value: the attacker may choose the chaining value.
What remains is a collision of the full internal compression function — "an
explicit freestart colliding pair for SHA-1, i.e. a collision for its internal
compression function ... reaching **all 80 out of 80 steps**."

The mechanism is *relaxation plus execution*. Freeing the chaining value removes
the hardest conformance constraints, which converts an infeasible search into a
feasible one, and the attack is then actually **run on hardware**: "Only 10 days
of computation on a 64-GPU cluster were necessary ... for a cost of approximately
2^57.5 calls to the compression function." The move from analysis to executed
computation is itself the change — the GPU implementation is part of the
mechanism, not an incidental detail.

The authors are explicit about the limit, and this record follows their wording
rather than overstating it: "Freestart collisions do not directly imply a
collision for the full hash function. However, this work is an important
milestone towards an actual SHA-1 collision."

Scored against the declared objective this is a **partial success**: full 80
steps, real artifact, executed cost — but on a modified starting condition, so it
is not a SHA-1 collision. The boundary is precisely the chaining value.

What makes this record structurally interesting is that the relaxation was
**instrumented**. Running the weakened attack produced a measured cost for real
GPU execution, which calibrated the estimates feeding the unrelaxed attack. The
attempt is a probe as much as a result: it converts an untested theoretical
estimate into a measured engineering quantity, and the full collision followed
about sixteen months later using the same infrastructure.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "sha1/freestart-compression-collision",
      "label": "Freestart collision by relaxing the initial chaining value",
      "description": "Deliberately weaken the objective by letting the attacker choose the chaining value, making the search feasible, and execute the attack on GPU hardware to produce a real collision of the full 80-step compression function.",
      "mechanism": {
        "representations": [
          "compression function with attacker-chosen chaining value",
          "freestart collision as an achievable relaxation of the target",
          "attack cost as measured GPU execution rather than estimate"
        ],
        "assumptions": [
          "relaxing the initial value preserves the attack's structural relevance to the unrelaxed target",
          "measured freestart cost calibrates estimates for the unrelaxed attack",
          "GPU implementation efficiency is representative"
        ],
        "operators": [
          "objective relaxation",
          "chaining value freeing",
          "GPU-parallel collision search",
          "executed computation",
          "cost measurement"
        ],
        "preserves": [
          "joint local-collision analysis machinery",
          "full 80-step coverage",
          "differential path attack frame"
        ],
        "breaks": [
          "fixed standard initial value requirement",
          "analysis-only evaluation of attack cost"
        ],
        "auxiliary_objects": [
          "64-GPU cluster",
          "explicit freestart colliding pair",
          "GPU attack implementation"
        ],
        "locality": "mixed",
        "construction_mode": "constructive",
        "uncertainty_mode": "probabilistic",
        "notes": "Functions as a probe as much as a result: it converts an untested theoretical estimate into a measured engineering quantity on real hardware.",
        "completeness_scope": "declared_payload",
        "completeness_basis": "Every representation/assumption/operator/preserves/breaks/auxiliary_object entry this note states is listed here and parsed by the deterministic embedded-payload parser. Authored 2026-09-12 by the agent identified in manifest.json authoring_provenance, from the read_scope recorded for src-stevens-karpman-peyrin-2016 (abstract and IACR-hosted PDF). declared_payload scope, NOT a mechanism-exhaustive extraction judgment.",
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
        "boundary_statement": "produces a real colliding pair for the full 80-step compression function in 10 days on a 64-GPU cluster at about 2^57.5 compression calls, but with an attacker-chosen chaining value, so it is not a collision of SHA-1",
        "boundary_conditions": [
          "attacker chooses the chaining value rather than using SHA-1's standard initial value",
          "authors state freestart collisions do not directly imply a collision for the full hash function",
          "full 80 of 80 steps are covered"
        ],
        "notes": "Partial success relative to the declared objective, in the authors' own framing as a milestone. The full collision followed about sixteen months later on the same infrastructure."
      },
      "support": [
        {
          "field_path": "mechanism.operators",
          "support_kind": "explicit",
          "locator": "para:3"
        },
        {
          "field_path": "mechanism.breaks",
          "support_kind": "inferred",
          "locator": "para:2"
        },
        {
          "field_path": "outcome.class",
          "support_kind": "inferred",
          "locator": "para:5"
        },
        {
          "field_path": "outcome.boundary_statement",
          "support_kind": "explicit",
          "locator": "para:3"
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
