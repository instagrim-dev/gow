# HELD-OUT TARGET — Affine-lattice / geometry-of-numbers linear forms

**Status: HELD OUT for the historical-holdout experiment. Do NOT ingest into the
pre-cutoff problem/workspace.** Staged as the later-productive mechanism the
invariant-guided search must try to recover from pre-cutoff failure structure alone.

**Era:** modern constructive program (post-cutoff relative to `2026-09-07` for the
purposes of this experiment).

## The structural move

Every pre-cutoff family that stalls (es-01, es-02, es-08, es-10, es-12) preserves the
same invariant: it reasons through **residue-class locality** and congruence
identities, and is therefore confined by the quadratic-reciprocity wall to the
non-residue side. The candidate failure invariant across the atlas is:

> Methods built from residue-class congruence identities cannot cross the
> quadratic-residue boundary.

The held-out family **deliberately breaks that invariant.** Instead of covering `n`
by residue classes, it recasts the problem as **conditions linear in `n`** and reads
the solution set as an **affine class in `Z³`** (an affine lattice / affine
sublattice). A representative linear form is

```text
(4b − 1)(4c − 1) = 4 P δ + 1
```

which is linear in `P` and describes admissible `(b, c, δ)` as lattice points in an
affine subspace. Existence then becomes a **geometry-of-numbers** question — does the
affine class contain a lattice point in the positive cone? — attacked with lattice
enumeration, convex-body / Minkowski-style arguments, and proofs of convergence,
rather than with per-class polynomial identities.

The mechanistic distinctions from the pre-cutoff atlas:

- locality: **global-geometric**, not residue-local;
- representation: **affine lattice in `Z³`**, not congruence classes;
- operators: **geometry of numbers / lattice enumeration**, not modular identity;
- it *breaks* residue-locality and the confinement-to-non-residues invariant instead
  of preserving it;
- it introduces an auxiliary object — the affine lattice and its convex body — that
  no pre-cutoff family used.

This is the "recover the structural move" target: the point is not the exact
identity, but that the productive direction **abandons residue-class locality for an
affine-lattice / geometry-of-numbers representation**, which is exactly the invariant
the failure families all conserved.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "erdos-straus/affine-lattice-linear-forms",
      "label": "Affine-lattice / geometry-of-numbers linear forms",
      "description": "Recast solvability as linear conditions in n describing an affine class in Z^3, then attack existence with geometry of numbers.",
      "mechanism": {
        "representations": ["affine lattice in Z^3", "linear forms in n", "convex body / positive cone"],
        "assumptions": ["solution set is an affine class", "lattice-point existence is decidable via geometry of numbers"],
        "operators": ["lattice enumeration", "geometry of numbers", "convergence proof", "linearization in n"],
        "preserves": ["positivity of denominators"],
        "breaks": ["residue-class locality", "confinement to quadratic non-residues"],
        "auxiliary_objects": ["affine sublattice", "Minkowski-style convex body"],
        "locality": "global",
        "construction_mode": "constructive",
        "uncertainty_mode": "deterministic",
        "notes": "Abandons congruence-class covering for an affine-lattice representation; targets the QR wall directly."
      },
      "outcome": {
        "class": "partial_success",
        "boundary_statement": "lattice organization reaches the residue-protected classes the congruence program could not",
        "boundary_conditions": ["global convergence of the lattice enumeration must be established"],
        "notes": "Held-out productive direction; breaks the shared failure invariant."
      },
      "support": [
        {"field_path": "mechanism.representations", "support_kind": "explicit", "locator": "structural move"},
        {"field_path": "mechanism.breaks", "support_kind": "explicit", "locator": "structural move"},
        {"field_path": "mechanism.operators", "support_kind": "explicit", "locator": "structural move"},
        {"field_path": "outcome.boundary_statement", "support_kind": "inferred", "locator": "structural move"}
      ]
    }
  ]
}
newf-normalize -->
