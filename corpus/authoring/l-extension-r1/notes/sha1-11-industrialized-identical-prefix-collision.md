# First collision for full SHA-1 (sha1-11)

**Era:** 2017. Source: `src-shattered-2017`, artifact `src-shattered-artifacts`.

This attempt achieves the declared objective. It is the "first known instance of a
collision for full SHA-1," realized as two PDF files with different
arbitrarily-chosen visual contents that share a SHA-1 digest.

The mechanism adds no fundamentally new cryptanalytic idea. It **industrializes**
the existing one: the joint local-collision analysis of `sha1-09` supplies the
path, the GPU infrastructure proven by the freestart attack in `sha1-10` supplies
the compute, and the remaining work is engineering a distributed computation
across data centers. Reported cost: 2^63.1 compression function calls,
approximately 6,500 CPU years and 100 GPU years. The mechanism change relative to
its predecessors is one of *scale and orchestration*, and that is precisely what
had been missing.

The outcome is **success**, and this corpus's objective is met here. Two
boundaries must stay attached to it.

First, it is an **identical-prefix** collision, in the authors' words: "This is an
identical-prefix collision attack, where a given prefix P is extended with two
distinct near-collision block pairs such that they collide for any suffix S." The
attacker does not control two *different* chosen prefixes. That stronger
capability is a separate objective, reached later (`sha1-12`). Preimage
resistance is untouched.

Second, the realized cost **exceeded the estimate**. The 2^61 figure from
`sha1-09` became 2^63.1 in execution, and the authors attribute the gap to GPU
efficiency loss and "the inefficiency we encountered in actually launching a large
scale computation distributed over several data centers." This is the corpus's
cleanest evidence that theoretical estimates in this domain are optimistic lower
bounds.

Two provenance facts are recorded because they matter to anyone re-checking this.
The paper states **100 GPU years** while Google's announcement states **110**;
both are recorded and not reconciled. And `shattered.io` no longer serves the
original project page — the live domain hosts unrelated commercial content — so
project *claims* must be cited from the Wayback capture. The two PDFs at
`shattered.io/static/` are still genuine: both were downloaded and both hash to
SHA-1 `38762cf7f55934b34d179ae6a4c80cadccbb7f0a` with distinct SHA-256 digests,
differing at byte 193. That local verification, not the citation, is the strongest
evidence in this corpus — the outcome is checkable by anyone with a hash utility.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "sha1/industrialized-identical-prefix-collision",
      "label": "Industrialized identical-prefix collision at data-center scale",
      "description": "Achieve a real collision for full SHA-1 by combining the existing joint local-collision path with proven GPU infrastructure and engineering a distributed computation across data centers, rather than by new cryptanalysis.",
      "mechanism": {
        "representations": [
          "identical-prefix collision: shared prefix extended by two distinct near-collision block pairs",
          "colliding PDF pair as the delivered artifact",
          "attack cost as realized distributed computation"
        ],
        "assumptions": [
          "existing differential path analysis is sufficient without new cryptanalytic ideas",
          "distributed orchestration overhead is manageable at this scale",
          "a chosen file format can absorb the colliding blocks while differing visually"
        ],
        "operators": [
          "distributed computation orchestration",
          "GPU-parallel near-collision block search",
          "near-collision block pair chaining",
          "file-format engineering for visual difference",
          "executed collision production"
        ],
        "preserves": [
          "joint local-collision analysis machinery",
          "GPU attack infrastructure from the freestart attack",
          "multi-block identical-prefix structure",
          "differential path attack frame"
        ],
        "breaks": [
          "absence of a realized full-SHA-1 collision",
          "single-facility compute scale"
        ],
        "auxiliary_objects": [
          "two colliding PDF files",
          "near-collision block pairs",
          "multi-data-center GPU and CPU infrastructure"
        ],
        "locality": "mixed",
        "construction_mode": "constructive",
        "uncertainty_mode": "probabilistic",
        "notes": "The mechanism change relative to its predecessors is scale and orchestration rather than new cryptanalysis, and that is exactly what had been missing.",
        "completeness_scope": "declared_payload",
        "completeness_basis": "Every representation/assumption/operator/preserves/breaks/auxiliary_object entry this note states is listed here and parsed by the deterministic embedded-payload parser. Authored 2026-09-12 by the agent identified in manifest.json authoring_provenance, from the read_scope recorded for src-shattered-2017 (full text) and local hash verification of src-shattered-artifacts. declared_payload scope, NOT a mechanism-exhaustive extraction judgment.",
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
        "class": "success",
        "boundary_statement": "achieves the objective with a real identical-prefix collision for full SHA-1 at 2^63.1 compression calls, but does not provide chosen-prefix capability and does not affect preimage resistance",
        "boundary_conditions": [
          "identical-prefix only: the attacker does not control two different chosen prefixes",
          "realized cost 2^63.1 exceeded the 2^61 estimate, attributed by the authors to GPU efficiency loss and distributed-launch inefficiency",
          "preimage resistance is unaffected",
          "paper states 100 GPU years while the vendor announcement states 110; both recorded, not reconciled"
        ],
        "notes": "Success relative to the declared objective, independently checkable: both published PDFs were downloaded and hash to SHA-1 38762cf7f55934b34d179ae6a4c80cadccbb7f0a with distinct SHA-256 digests."
      },
      "support": [
        {
          "field_path": "mechanism.operators",
          "support_kind": "inferred",
          "locator": "para:3"
        },
        {
          "field_path": "mechanism.preserves",
          "support_kind": "explicit",
          "locator": "para:3"
        },
        {
          "field_path": "outcome.class",
          "support_kind": "explicit",
          "locator": "para:2"
        },
        {
          "field_path": "outcome.boundary_statement",
          "support_kind": "explicit",
          "locator": "para:5"
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
