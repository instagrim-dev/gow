# Withdrawn 2^52 differential path (sha1-07)

**Era:** 2009. Source: `src-mcdonald-hawkes-pieprzyk-2009` (withdrawn).

This attempt **composes two other attempts' results** and claims the best
complexity yet: "a new differential path which can be used in a collision attack
with complexity of O(2^52). This is currently the lowest complexity attack on
SHA-1." The mechanism is composition rather than new analysis. The authors' own
decomposition, quoted in the withdrawal note, is explicit: the Round 1 nonlinear
path uses five auxiliary differentials in a boomerang attack — the technique from
`sha1-05` — while the Rounds 2–4 linear path is taken with a cost evaluation of
2^57 from Manuel (`sha1-06`). The overall attack then requires 2^52 conforming
message pairs for Round 1, achieved by message modification.

The paper was **withdrawn by its own authors**, and the withdrawal note states
the reason: "The cost evaluation for point 2 above was reported by Manuel in
November 2008. We have recently discovered this evaluation to be incorrect. This
implies the complexity given in this paper is also incorrect. We have decided to
withdraw the current paper."

Scored against the objective this is a **failure**, and it is the most
methodologically informative record in the corpus for a specific reason: the
failure is not in the attempt's own construction. The boomerang placement and the
message-modification scheme were not shown wrong. What failed was an **imported
premise** — a borrowed cost figure the authors did not independently verify. The
composition was sound; one input was not.

The boundary this establishes is about **dependency rather than technique**: a
complexity claim assembled from another paper's unverified cost evaluation
inherits that evaluation's fragility completely. RFC 6194 independently notes
that Manuel's 2^51 claim "is absent from the published conference paper," so the
weakness was visible in the source's own publication history.

This record exists because the field corrected itself. The ePrint page still
serves the paper and its own retraction, which is why this corpus cites the
withdrawal directly rather than paraphrasing a secondary account of it. One
scope caution: the claim that this result was announced at the Eurocrypt 2009
rump session could not be verified against any retrievable programme, so this
record asserts the withdrawal, not an announcement venue.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "sha1/composed-boomerang-linear-path-withdrawn",
      "label": "Composed boomerang and linear-path differential attack (withdrawn)",
      "description": "Compose a boomerang-based nonlinear Round 1 path with a borrowed linear path cost evaluation for Rounds 2-4 to claim a 2^52 collision attack; withdrawn by its authors when the borrowed cost evaluation proved incorrect.",
      "mechanism": {
        "representations": [
          "differential path partitioned into Round 1 nonlinear and Rounds 2-4 linear segments",
          "borrowed per-segment cost evaluation as a composable quantity",
          "conforming message pair count as the attack cost unit"
        ],
        "assumptions": [
          "the borrowed Rounds 2-4 cost evaluation of 2^57 is correct",
          "segment costs compose to a valid overall complexity",
          "five auxiliary differentials transfer to this path unchanged"
        ],
        "operators": [
          "result composition across published attempts",
          "boomerang auxiliary differential placement",
          "message modification",
          "complexity derivation from borrowed cost evaluation"
        ],
        "preserves": [
          "differential path attack frame",
          "boomerang auxiliary-differential technique",
          "message modification for early-round conformance"
        ],
        "breaks": [],
        "auxiliary_objects": [
          "five auxiliary differentials",
          "Manuel's Rounds 2-4 linear differential path",
          "borrowed 2^57 cost evaluation"
        ],
        "locality": "mixed",
        "construction_mode": "constructive",
        "uncertainty_mode": "probabilistic",
        "notes": "The attempt's own construction was never shown incorrect. What failed was an imported premise the authors did not independently verify.",
        "completeness_scope": "declared_payload",
        "completeness_basis": "Every representation/assumption/operator/preserves/breaks/auxiliary_object entry this note states is listed here and parsed by the deterministic embedded-payload parser. Authored 2026-09-12 by the agent identified in manifest.json authoring_provenance, from the read_scope recorded for src-mcdonald-hawkes-pieprzyk-2009 (ePrint page and withdrawal note). declared_payload scope, NOT a mechanism-exhaustive extraction judgment.",
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
        "boundary_statement": "withdrawn by its own authors: the borrowed 2^57 cost evaluation for the Rounds 2-4 linear path was discovered to be incorrect, invalidating the derived 2^52 complexity",
        "boundary_conditions": [
          "failure located in an imported premise, not in the attempt's own construction",
          "a complexity claim assembled from another paper's unverified cost evaluation inherits that evaluation's fragility",
          "the source's own publication history already showed the weakness: RFC 6194 notes the related 2^51 claim is absent from the published conference paper",
          "no collision or reduced-round artifact was produced"
        ],
        "notes": "Failure relative to the declared objective, produced by the field's own error correction rather than by a curator's judgment. The rump-session announcement venue is unverified and is not asserted."
      },
      "support": [
        {
          "field_path": "mechanism.assumptions",
          "support_kind": "explicit",
          "locator": "para:2"
        },
        {
          "field_path": "mechanism.operators",
          "support_kind": "explicit",
          "locator": "para:2"
        },
        {
          "field_path": "mechanism.auxiliary_objects",
          "support_kind": "explicit",
          "locator": "para:2"
        },
        {
          "field_path": "outcome.class",
          "support_kind": "explicit",
          "locator": "para:3"
        },
        {
          "field_path": "outcome.boundary_statement",
          "support_kind": "explicit",
          "locator": "para:3"
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
