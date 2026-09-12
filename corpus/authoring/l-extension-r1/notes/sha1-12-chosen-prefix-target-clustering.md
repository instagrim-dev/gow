# Chosen-prefix collision via target-difference birthday search (sha1-12)

**Era:** 2019. Source: `src-leurent-peyrin-2019`.

The 2017 collision meets the base objective but leaves a stronger capability
unreached: both messages must share a prefix. A **chosen-prefix** collision lets
each party pick its own arbitrary prefix first, which is what turns a collision
into a practical forgery against real protocols. Stevens' 2013 estimate put this
at 2^77.1 — far out of reach.

The mechanism reaches it by **restructuring the search into two phases**. Phase
one performs a birthday search that steers the two independently-chosen prefixes
toward a chaining-value difference lying in a *pre-defined target set*, rather
than toward any single value. Phase two eliminates the remaining difference with a
multi-block near-collision sequence, exploiting a **clustering effect**: many
distinct target differences are reachable by the same near-collision block work,
so the birthday phase only has to hit the set, not a point. Widening the target
from a point to a cluster is the whole idea.

The reported complexity is **between 2^66.9 and 2^69.4** — a range, not a point
estimate, "depending on assumptions about the cost of finding near-collision
blocks." Against the previous 2^77.1 that is a reduction of roughly 2^8 to 2^10.
The same technique also applies to MD5.

Scored against the objective this is a **partial success**. It concerns a stronger
objective than the base one and reaches it only in analysis: no chosen-prefix
collision was produced here. The authors themselves flag the near-collision-block
cost assumption as the source of the range, which is a candid statement that the
figure is assumption-dependent — notably the same class of dependency that sank
the withdrawn 2^52 claim in `sha1-07`, but here disclosed as a range rather than
presented as a point.

The practical outcome followed a year later. Leurent and Peyrin implemented the
attack at 2^63.4, producing the first actual chosen-prefix collision on two months
of 900 GTX 1060 GPUs, and demonstrated it against the PGP/GnuPG web of trust as
CVE-2019-14855, fixed in GnuPG 2.2.18. Their cost reporting is worth preserving
precisely: the paper estimates about $45,000, while the project page states the
attack "costed us about 75k USD" with $45,000 as a retrospective estimate at then-
current prices. Both figures are recorded and not reconciled. Even there, the
authors note that HMAC-SHA-1 "seems relatively safe" and that preimage resistance
"remains unbroken."

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "sha1/chosen-prefix-target-difference-clustering",
      "label": "Chosen-prefix collision via birthday search onto clustered target differences",
      "description": "Reach chosen-prefix collision by steering independently-chosen prefixes via birthday search toward a pre-defined set of target chaining-value differences, then cancelling the remainder with multi-block near-collisions that exploit a clustering effect.",
      "mechanism": {
        "representations": [
          "chosen-prefix collision with two independently selected prefixes",
          "pre-defined set of target chaining-value differences",
          "clustering effect over reachable target differences",
          "complexity as a range rather than a point estimate"
        ],
        "assumptions": [
          "near-collision block cost is bounded within the stated range",
          "many target differences are reachable by shared near-collision block work",
          "birthday search onto a set is materially cheaper than onto a point"
        ],
        "operators": [
          "two-phase attack decomposition",
          "birthday search onto a target difference set",
          "multi-block near-collision cancellation",
          "clustering effect exploitation"
        ],
        "preserves": [
          "joint local-collision analysis machinery",
          "multi-block near-collision structure",
          "differential path attack frame"
        ],
        "breaks": [
          "identical-prefix requirement",
          "single target difference requirement"
        ],
        "auxiliary_objects": [
          "target difference set",
          "near-collision block sequence",
          "birthday search structure"
        ],
        "locality": "mixed",
        "construction_mode": "constructive",
        "uncertainty_mode": "probabilistic",
        "notes": "Widening the birthday target from a point to a cluster is the core idea. The complexity is reported as an assumption-dependent range, disclosed rather than collapsed to a point.",
        "completeness_scope": "declared_payload",
        "completeness_basis": "Every representation/assumption/operator/preserves/breaks/auxiliary_object entry this note states is listed here and parsed by the deterministic embedded-payload parser. Authored 2026-09-12 by the agent identified in manifest.json authoring_provenance, from the read_scope recorded for src-leurent-peyrin-2019 (abstract and ePrint text), with the 2020 follow-on from src-leurent-peyrin-2020. declared_payload scope, NOT a mechanism-exhaustive extraction judgment.",
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
        "boundary_statement": "reduces chosen-prefix collision complexity from 2^77.1 to a range of 2^66.9 to 2^69.4 in analysis, but produces no chosen-prefix collision and the figure is explicitly assumption-dependent",
        "boundary_conditions": [
          "complexity is a range, dependent on assumptions about near-collision block cost",
          "no chosen-prefix collision artifact produced in this work",
          "concerns a stronger objective than the base identical-prefix collision",
          "the follow-on 2020 work implemented it at 2^63.4; its paper estimates about 45k USD while the project page reports about 75k USD actually spent"
        ],
        "notes": "Partial success relative to the declared objective. Preimage resistance and HMAC-SHA-1 remain unaffected per the authors of the follow-on work."
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
