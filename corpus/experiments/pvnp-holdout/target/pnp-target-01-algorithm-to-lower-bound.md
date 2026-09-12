# WITHHELD TARGET — faster exhaustive search implies lower bounds (pnp-target-01)

**Status: WITHHELD TARGET for the historical-holdout experiment. Do NOT
ingest into the train problem/workspace.** Historical claim intended:
dated_at 2010-06-05 (STOC 2010), strictly after the 2005-01-01 cutoff;
evidence locator recorded via `experiment date-source` at gate-lift time.

## The structural move

Every train family analyzes the **model side**: fix a computational class,
find a property its members respect, show a hard function lacks it (pnp-03,
04, 06, 07, 10) — or simulate members as black boxes (pnp-01, 12). Both walls
are statements about those postures (pnp-02, pnp-08). The one train family
that gestures at reversal (pnp-11) is conditional: it consumes a *hypothetical*
algorithm.

The target move **consummates the reversal unconditionally in principle**: it
proves that a *slightly-better-than-brute-force* satisfiability algorithm for
a circuit class `C` — savings as small as superpolynomial over `2^n` — implies
`NEXP ⊄ C`. The engine couples the assumed algorithm to the
**nondeterministic time hierarchy** through succinct/compressed completeness:
if circuit-SAT for `C` is even mildly easy and NEXP ⊆ C, one can guess-and-check
succinct witnesses fast enough to violate the hierarchy. The contradiction
converts an algorithmic *upper* bound into a circuit *lower* bound.

Mechanistic distinctions from the train atlas:

- **direction**: algorithm→hardness, run on an algorithm to be *supplied*,
  not hypothesized-and-refuted (contrast pnp-11, pnp-12);
- **no distinguishing property**: nothing large or constructive over truth
  tables is exhibited — the natural-proofs frame (pnp-08) does not bind;
- **non-black-box**: the required SAT algorithm must exploit the anatomy of
  class `C`, so the argument does not relativize (pnp-02 does not bind);
- **auxiliary object**: the mild-savings circuit-analysis algorithm itself —
  an object no train family used as a lower-bound ingredient.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "pvnp/algorithm-to-lower-bound-conversion",
      "label": "Faster exhaustive search implies circuit lower bounds",
      "description": "A mildly-better-than-brute-force circuit-SAT algorithm for class C, coupled to the nondeterministic time hierarchy via succinct completeness, yields NEXP not in C.",
      "mechanism": {
        "representations": ["circuit-SAT instances", "succinct witness encodings", "nondeterministic time hierarchy"],
        "assumptions": ["a supplied SAT algorithm with superpolynomial savings over brute force", "succinct completeness for NEXP verification"],
        "operators": ["algorithm-to-hardness conversion", "succinct witness compression", "hierarchy contradiction"],
        "preserves": [],
        "breaks": ["model-analysis-first direction", "black-box relativizing simulation", "large constructive distinguishing property"],
        "auxiliary_objects": ["mild-savings circuit-analysis algorithm", "succinct SAT instances"],
        "locality": "global",
        "construction_mode": "constructive",
        "uncertainty_mode": "deterministic",
        "notes": "Inverts the conserved analysis direction: the lower bound is manufactured FROM an algorithm, with no truth-table property exhibited."
      },
      "outcome": {
        "class": "partial_success",
        "boundary_statement": "reduces circuit lower bounds to designing mildly nontrivial satisfiability algorithms for the target class",
        "boundary_conditions": ["a concrete algorithm for the class must still be supplied", "savings threshold must beat brute force superpolynomially"],
        "notes": "Withheld historical target; the companion entry supplies the algorithm for ACC0 and consummates the conversion."
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
