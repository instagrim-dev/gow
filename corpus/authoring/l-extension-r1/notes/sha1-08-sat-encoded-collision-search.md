# SAT-solver collision search (sha1-08)

**Era:** 2006. Source: `src-mironov-zhang-2006`.

This attempt **discards the differential-path frame entirely**. Instead of
constructing a characteristic and reasoning about its probability, it encodes the
collision condition as a Boolean satisfiability instance — the compression
function as a circuit, the requirement that two distinct inputs produce equal
output as a constraint — and hands the whole thing to a general-purpose SAT
solver. The mechanism substitutes a **generic automated reasoner** for
domain-specific analysis. No disturbance vector, no local collisions, no
hand-built path.

On weaker functions in the same family this works outright. The authors report a
collision for MD4 in under 10 minutes and one for MD5 in approximately 100 hours,
using off-the-shelf solvers.

It does not scale. For SHA-0 the authors report: "we estimate that generating a
full collision using SatELiteGTI would require approximately **3 million CPU
hours** ... A successful SAT-solver-aided attack on SHA-0 is still a theoretical
possibility." SHA-1 was left as an explicit open problem — not attempted, not
estimated.

Relative to the objective this is a **failure**, and its value is as a
**negative control on generality**. The attempt shows that the domain's progress
was not merely a matter of applying enough automated search. The
differential-path machinery every other record in this corpus preserves is not
incidental scaffolding that a strong enough solver could bypass; it is what makes
the problem tractable at all. A method that solves MD4 in minutes cannot reach
SHA-0 in three million CPU hours, and the gap is structural rather than a matter
of degree.

The boundary is worth stating precisely: SAT-based collision search never reached
**full SHA-0**, and never attempted **full SHA-1**. Absence of an attempt is not
evidence about difficulty, so this record makes no claim about what SAT solvers
could achieve on SHA-1 — only what was tried and what was reported.

A correction is recorded here deliberately. A frequently-cited candidate for
"SAT-based SHA-1 attack" is Nossum's 2012 master's thesis, but that work is a
**preimage** study whose main contribution is an encoding improvement for 32-bit
modular addition; it contains no collision claim. Attributing a failed SAT
collision attempt to it would be wrong. The register retains it as an explicit
exclusion (`src-nossum-2012-excluded`).

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "sha1/sat-encoded-collision-search",
      "label": "SAT-encoded collision search with a general-purpose solver",
      "description": "Encode the collision condition for the compression function directly as a Boolean satisfiability instance and delegate the search to an off-the-shelf SAT solver, using no differential path or disturbance vector.",
      "mechanism": {
        "representations": [
          "compression function as a Boolean circuit",
          "collision condition as a CNF satisfiability instance",
          "message pair as a satisfying assignment"
        ],
        "assumptions": [
          "generic automated search can substitute for domain-specific differential analysis",
          "solver performance on weaker family members predicts performance on stronger ones",
          "the collision condition admits a tractable CNF encoding"
        ],
        "operators": [
          "CNF encoding of the compression function",
          "general-purpose SAT solving",
          "clause learning and propagation",
          "solver-driven search"
        ],
        "preserves": [
          "the collision condition itself as the only stated constraint"
        ],
        "breaks": [
          "differential path construction",
          "disturbance vector selection",
          "local collision substrate",
          "domain-specific message modification"
        ],
        "auxiliary_objects": [
          "CNF instance",
          "off-the-shelf SAT solver"
        ],
        "locality": "global",
        "construction_mode": "existential",
        "uncertainty_mode": "deterministic",
        "notes": "A negative control on generality: it discards everything the successful family preserves, succeeds on MD4 and MD5, and fails to reach SHA-0 by a structural margin rather than a matter of degree.",
        "completeness_scope": "declared_payload",
        "completeness_basis": "Every representation/assumption/operator/preserves/breaks/auxiliary_object entry this note states is listed here and parsed by the deterministic embedded-payload parser. Authored 2026-09-12 by the agent identified in manifest.json authoring_provenance, from the read_scope recorded for src-mironov-zhang-2006 (full text). declared_payload scope, NOT a mechanism-exhaustive extraction judgment.",
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
        "boundary_statement": "succeeds on MD4 and MD5 but does not scale to SHA-0, estimated at approximately 3 million CPU hours, and full SHA-1 was left as an explicit open problem rather than attempted",
        "boundary_conditions": [
          "MD4 collision in under 10 minutes; MD5 collision in approximately 100 hours",
          "full SHA-0 collision estimated at approximately 3 million CPU hours and never produced",
          "full SHA-1 never attempted, so no evidence about SHA-1 difficulty is claimed here"
        ],
        "notes": "Failure relative to the declared objective. Informative as a generality control: it shows the differential-path machinery is load-bearing rather than incidental."
      },
      "support": [
        {
          "field_path": "mechanism.operators",
          "support_kind": "inferred",
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
          "locator": "para:6"
        }
      ]
    }
  ]
}
newf-normalize -->
