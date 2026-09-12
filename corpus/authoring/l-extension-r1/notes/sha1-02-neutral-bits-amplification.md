# Neutral bits and near-collisions on SHA-0 (sha1-02)

**Era:** 2004. Source: `src-biham-chen-2004`.

Where the 1998 attack was purely analytic, this attempt adds a **search-control**
mechanism. The key observation is that a conforming message pair has many
**neutral bits**: bit positions that can be flipped without disturbing the
differential path for a substantial stretch of rounds. The paper states that "the
messages have many neutral bits, some of which do not affect the differences for
about 15--20 rounds." Flipping them generates many candidate pairs from one
expensively-found conforming pair, so the cost of satisfying the early rounds is
amortized rather than repaid on every trial.

The mechanism is a **hybrid**: it keeps the differential path from the analytic
family and bolts on a cheap amplifier over the message space. That amplifier is
generic — it does not care how the path was found — which is why it survived into
every later attack in this domain even after the paths themselves were replaced.

The results are real but scoped to SHA-0: two near-collisions of the **full SHA-0
compression function**, with up to 142 of 160 output bits equal, and full
collisions of **65-round reduced SHA-0**, against a prior best of 35 rounds. The
paper also reports that 82-round SHA-0 is weaker than the standard 80-round
SHA-0 — strength is not monotone in round count, which undercuts any assumption
that reduced-round results extrapolate smoothly.

Relative to the objective this is a **partial failure**: nothing here collides
SHA-1, and even on SHA-0 the full-function result is a *near*-collision, not a
collision. Two boundaries are worth separating. A near-collision leaves residual
output difference and must be composed with something else to become a
collision. And a reduced-round collision is not evidence about the full function,
especially given this paper's own non-monotonicity finding.

One correction is recorded deliberately, because the mistake is easy and
consequential: this paper's reduced-round collisions are **SHA-0**, not SHA-1.
Reduced-round SHA-1 collisions belong to Biham, Chen, Joux, Carribault, Lemuet
and Jalby at EUROCRYPT 2005, a different publication that this corpus does not
include.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "sha1/neutral-bits-amplification",
      "label": "Neutral-bit amplification over a fixed differential path",
      "description": "Amplify one expensively-found conforming message pair into many candidates by flipping bit positions that leave the differential path undisturbed for 15-20 rounds, amortizing early-round conformance cost.",
      "mechanism": {
        "representations": [
          "differential path with per-round conformance conditions",
          "neutral-bit set over the message block",
          "near-collision as bounded residual output difference"
        ],
        "assumptions": [
          "some message bits leave early-round path conformance invariant",
          "amortizing early-round cost dominates total search cost",
          "reduced-round behavior informs full-round behavior"
        ],
        "operators": [
          "neutral-bit identification",
          "neutral-bit flipping",
          "differential path conformance testing",
          "round reduction"
        ],
        "preserves": [
          "analytic differential path from the linearized family",
          "single-block message structure"
        ],
        "breaks": [
          "one-trial-per-path-search cost model"
        ],
        "auxiliary_objects": [
          "conforming seed message pair",
          "neutral-bit index set"
        ],
        "locality": "local",
        "construction_mode": "constructive",
        "uncertainty_mode": "probabilistic",
        "notes": "The amplifier is generic with respect to how the path was obtained, which is why it outlived the paths it was introduced with. The paper's 82-vs-80-round finding contradicts its own third assumption.",
        "completeness_scope": "declared_payload",
        "completeness_basis": "Every representation/assumption/operator/preserves/breaks/auxiliary_object entry this note states is listed here and parsed by the deterministic embedded-payload parser. Authored 2026-09-12 by the agent identified in manifest.json authoring_provenance, from the read_scope recorded for src-biham-chen-2004 (abstract). declared_payload scope, NOT a mechanism-exhaustive extraction judgment.",
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
        "boundary_statement": "achieves near-collisions of the full SHA-0 compression function and collisions of 65-round reduced SHA-0, but no collision of full SHA-0 and nothing on SHA-1",
        "boundary_conditions": [
          "near-collision leaves residual output difference requiring further composition",
          "reduced-round results do not extrapolate: the paper reports 82-round SHA-0 weaker than 80-round SHA-0",
          "target is SHA-0, not SHA-1"
        ],
        "notes": "Partial failure relative to the declared objective. The neutral-bits technique itself is the durable contribution and transferred to later SHA-1 attacks."
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
          "field_path": "outcome.class",
          "support_kind": "explicit",
          "locator": "para:4"
        },
        {
          "field_path": "outcome.boundary_statement",
          "support_kind": "explicit",
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
