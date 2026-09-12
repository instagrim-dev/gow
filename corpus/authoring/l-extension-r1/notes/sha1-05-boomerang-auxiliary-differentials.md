# Boomerang auxiliary differentials (sha1-05)

**Era:** 2007. Source: `src-joux-peyrin-2007`.

This attempt **transfers a block-cipher technique** into hash cryptanalysis. The
boomerang attack was developed for ciphers; the observation here is that its
structure can be repurposed as a stronger version of neutral bits. The authors
state the move directly: "we show that another tool of block cipher analysis, the
boomerang attack, can also be used in this context. In particular, we show that
using this boomerang attack as a neutral bits tool, it becomes possible to lower
the complexity of the attacks on SHA-1."

Mechanistically, the improvement is about **how far conformance can be carried
for free**. Plain neutral bits and message modification maintain path conformance
to roughly step 25. By placing **five auxiliary differentials** — small
quartet-structured perturbations borrowed from the boomerang frame — the authors
maintain conformance to **step 28 or beyond**. Each extra step of free
conformance removes a probabilistic factor from the search, and the reported
improvement is a **relative factor of 32**.

This is a mechanism-transfer record rather than a frontier-moving one, and its
classification depends on that distinction. It is a **partial success**: the
transfer works, the speedup is real, and the technique persisted into the attacks
that eventually succeeded. But no collision was produced, no reduced-round
artifact is claimed here, and the objective is untouched.

Two honesty constraints are recorded because they bound what this record may be
used for. First, the verified figure is a **relative factor of 32**, not an
absolute complexity: no absolute figure was confirmable, so none is encoded. Any
absolute number attached to this work elsewhere should be treated as unverified.
Second, the technical claims here were read from the ECRYPT Hash Workshop 2007
version on the author's site, not the CRYPTO 2007 proceedings version; the
bibliography is confirmed by DOI but the text is a sibling version. That is a
provenance limitation, recorded rather than smoothed over.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "sha1/boomerang-auxiliary-differentials",
      "label": "Boomerang-derived auxiliary differentials as a neutral-bits amplifier",
      "description": "Transfer the boomerang attack from block-cipher cryptanalysis into SHA-1 collision search, placing auxiliary differentials that extend free path conformance several steps beyond what neutral bits or message modification achieve.",
      "mechanism": {
        "representations": [
          "auxiliary differential as quartet-structured perturbation",
          "boomerang quartet adapted to a compression-function setting",
          "path conformance depth measured in steps"
        ],
        "assumptions": [
          "block-cipher boomerang structure transfers to hash compression functions",
          "auxiliary differentials can be placed without invalidating the main path",
          "extending free conformance depth reduces total search cost proportionally"
        ],
        "operators": [
          "cross-domain technique transfer",
          "auxiliary differential placement",
          "boomerang quartet construction",
          "conformance-depth extension"
        ],
        "preserves": [
          "differential path attack frame",
          "neutral-bit amplification role",
          "multi-block collision structure"
        ],
        "breaks": [
          "step-25 free-conformance ceiling"
        ],
        "auxiliary_objects": [
          "five auxiliary differentials",
          "boomerang quartet"
        ],
        "locality": "local",
        "construction_mode": "constructive",
        "uncertainty_mode": "probabilistic",
        "notes": "A mechanism-transfer record: the contribution is importing a technique from another subfield, not moving the complexity frontier by new analysis of SHA-1 itself.",
        "completeness_scope": "declared_payload",
        "completeness_basis": "Every representation/assumption/operator/preserves/breaks/auxiliary_object entry this note states is listed here and parsed by the deterministic embedded-payload parser. Authored 2026-09-12 by the agent identified in manifest.json authoring_provenance, from the read_scope recorded for src-joux-peyrin-2007 (ECRYPT workshop version; CRYPTO proceedings version not retrieved). declared_payload scope, NOT a mechanism-exhaustive extraction judgment.",
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
        "boundary_statement": "extends free path conformance from about step 25 to step 28 or beyond for a relative improvement factor of 32, but produces no collision and no absolute complexity figure is verified",
        "boundary_conditions": [
          "improvement is a relative factor of 32; no absolute complexity figure is verified for this work",
          "technical claims read from the ECRYPT Hash Workshop 2007 version, not the CRYPTO 2007 proceedings version",
          "no collision artifact produced at any round count"
        ],
        "notes": "Partial success relative to the declared objective. The technique persisted into the attacks that eventually succeeded, which is its durable value."
      },
      "support": [
        {
          "field_path": "mechanism.operators",
          "support_kind": "explicit",
          "locator": "para:2"
        },
        {
          "field_path": "mechanism.auxiliary_objects",
          "support_kind": "explicit",
          "locator": "para:3"
        },
        {
          "field_path": "mechanism.breaks",
          "support_kind": "explicit",
          "locator": "para:3"
        },
        {
          "field_path": "outcome.class",
          "support_kind": "inferred",
          "locator": "para:4"
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
