# WITHHELD TARGET — nonuniform ACC⁰ lower bounds via a supplied algorithm (pnp-target-02)

**Status: WITHHELD TARGET for the historical-holdout experiment. Do NOT
ingest into the train problem/workspace.** Historical claim intended:
dated_at 2011-06-08 (CCC 2011; journal version JACM 2014), strictly after the
2005-01-01 cutoff; evidence locator recorded via `experiment date-source` at
gate-lift time.

## The structural move, consummated

The companion entry (pnp-target-01) converts a mildly nontrivial circuit-SAT
algorithm into a lower bound. This entry **supplies the algorithm** for ACC⁰ —
the exact class where the train atlas's strongest program stalled for two
decades (pnp-06: no property found for composite moduli; pnp-09: the stall as
conserved structure).

The supplied ingredient: ACC⁰ circuits admit a known structural
transformation into depth-two circuits with a symmetric-function top gate
(quasipolynomial size), and dynamic programming plus fast rectangular matrix
multiplication evaluates such objects on all `2^n` inputs faster than
brute-force by the required superpolynomial savings. Feeding this concrete
satisfiability/evaluation algorithm through the conversion yields the
unconditional separation **NEXP ⊄ ACC⁰** — the first progress at that
frontier since pnp-06's era.

Why the walls do not bind, stated in the atlas's own vocabulary:

- **relativization (pnp-02)**: the algorithm exploits the anatomy of ACC⁰
  (the symmetric-top-gate normal form); the argument opens the box;
- **natural proofs (pnp-08)**: no large constructive property over truth
  tables appears anywhere — hardness emerges from a hierarchy contradiction
  about one function family, evading both largeness and constructivity;
- **the stall's conserved direction (pnp-09)**: reversed — the analysis
  object is an *algorithm*, and the circuit class's structure is consumed as
  the algorithm's fuel rather than as the property-bearing target.

The recovery question for the experiment: given only the train atlas, does
invariant-guided generation propose mechanisms whose signature is
**algorithm-supplying direction reversal against the stalled class** — not
another property search, not another restricted model, not another
conditional bridge.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "pvnp/acc0-lower-bound-via-supplied-algorithm",
      "label": "ACC0 lower bounds via a supplied evaluation algorithm",
      "description": "A concrete better-than-brute-force ACC0 satisfiability/evaluation algorithm (symmetric-top-gate normal form + fast matrix multiplication), fed through the algorithm-to-hardness conversion, yields NEXP not in ACC0 unconditionally.",
      "mechanism": {
        "representations": ["ACC0 circuits", "depth-two symmetric-top-gate normal form", "all-inputs evaluation tables"],
        "assumptions": ["structural normal form for ACC0 holds at quasipolynomial size"],
        "operators": ["circuit normal-form transformation", "dynamic-programming evaluation", "fast rectangular matrix multiplication", "algorithm-to-hardness conversion", "hierarchy contradiction"],
        "preserves": [],
        "breaks": ["model-analysis-first direction", "black-box relativizing simulation", "large constructive distinguishing property"],
        "auxiliary_objects": ["supplied ACC0 evaluation algorithm", "symmetric-function top gate"],
        "locality": "global",
        "construction_mode": "constructive",
        "uncertainty_mode": "deterministic",
        "notes": "Consummates the direction reversal at the exact class where the property-based program stalled; the class structure fuels the algorithm instead of bearing a property."
      },
      "outcome": {
        "class": "partial_success",
        "boundary_statement": "unconditional NEXP separation from ACC0, the first movement at the two-decade stall frontier",
        "boundary_conditions": ["separation reaches ACC0 only; TC0 and above require stronger supplied algorithms", "NEXP not yet lowered toward NP"],
        "notes": "Withheld historical target; structural move = supply an algorithm and reverse the analysis direction against the stalled class."
      },
      "support": [
        {"field_path": "mechanism.operators", "support_kind": "explicit", "locator": "para:2"},
        {"field_path": "mechanism.breaks", "support_kind": "explicit", "locator": "para:3"},
        {"field_path": "outcome.class", "support_kind": "explicit", "locator": "para:2"},
        {"field_path": "outcome.boundary_statement", "support_kind": "explicit", "locator": "para:2"}
      ]
    }
  ]
}
newf-normalize -->
