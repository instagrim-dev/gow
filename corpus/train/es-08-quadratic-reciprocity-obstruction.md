# Quadratic-reciprocity obstruction (es-08)

**Era:** Mordell (1960s), reframed by Elsholtz & Tao. Pre-cutoff.

This is not a construction but an **obstruction analysis** — the note that explains
*why* the identity/covering families (es-01, es-02) all stall on the same classes.
The congruence relations that carry polynomial identities are, in the relevant case,
governed by Jacobi symbols. An application of the quadratic reciprocity law shows
these relations can eliminate quadratic **non**-residues but never quadratic
**residues**: for the coprime data producing solutions, the relevant Jacobi symbols
are forced equal, so the congruence method is powerless exactly when `n` is a
quadratic residue (in particular whenever `n` is an odd perfect square, or lies in a
quadratic-residue class to small moduli).

The mechanism is global and existential in a meta sense: it reasons about the *space
of methods*, proving a negative structural fact rather than producing denominators.
It preserves — indeed pinpoints — the invariant that the whole identity/covering
program conserves: solvability-by-congruence is confined to the non-residue side.

Its "boundary" is that an obstruction is not a solution: it tells the search where
*not* to look and what any successful method must **break**, but it constructs
nothing. Its research value is high precisely as a candidate **failure invariant**:
every failed family here preserves the quadratic-residue wall.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "erdos-straus/quadratic-reciprocity-obstruction",
      "label": "Quadratic-reciprocity obstruction",
      "description": "Prove congruence/identity methods cannot eliminate quadratic-residue classes, via Jacobi-symbol reciprocity.",
      "mechanism": {
        "representations": ["Jacobi symbols", "residue/non-residue split"],
        "assumptions": ["identity methods reduce to congruence relations"],
        "operators": ["quadratic reciprocity", "obstruction argument"],
        "preserves": ["confinement of congruence methods to non-residues"],
        "breaks": [],
        "auxiliary_objects": ["quadratic reciprocity law"],
        "locality": "global",
        "construction_mode": "existential",
        "uncertainty_mode": "deterministic",
        "notes": "A negative structural result about the space of methods."
      },
      "outcome": {
        "class": "failure",
        "boundary_statement": "an obstruction locates the wall but constructs no solution",
        "boundary_conditions": ["quadratic-residue classes", "odd perfect squares"],
        "notes": "Candidate failure invariant: the QR wall is preserved by every congruence method."
      },
      "support": [
        {"field_path": "mechanism.operators", "support_kind": "explicit", "locator": "para:2"},
        {"field_path": "mechanism.preserves", "support_kind": "explicit", "locator": "para:2"},
        {"field_path": "outcome.boundary_statement", "support_kind": "inferred", "locator": "para:4"}
      ]
    }
  ]
}
newf-normalize -->
