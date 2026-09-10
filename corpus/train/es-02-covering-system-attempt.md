# Covering-system attempt (es-02)

**Era:** classical to modern; the hope predates and motivates Mordell's obstruction.
Pre-cutoff.

The idea couples the per-class polynomial identities (es-01) into a **covering
system of congruences**: choose finitely many moduli and residue classes so that
every integer `n` falls into at least one class that admits an identity. If such a
finite covering existed, it would resolve the conjecture globally by reducing it to
a finite check.

Unlike a single identity, this approach reasons globally — it seeks a finite family
of congruences whose union is all of `Z`. It is existential in posture: it asks
whether *some* covering exists, not constructing one class at a time for its own
sake.

The attempt fails for a structural reason, not for lack of effort: the congruence
relations that carry polynomial identities can only eliminate quadratic
*non*-residues. By quadratic reciprocity, no finite covering built from these
identities can catch the quadratic-residue classes (see es-08). The survivor set
is therefore non-empty for every candidate covering, and it contains infinitely
many primes (Dirichlet). The covering hope stalls on the same quadratic-residue
wall.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "erdos-straus/covering-system-attempt",
      "label": "Covering-system of congruences",
      "description": "Couple per-class identities into a finite covering system meant to catch every n.",
      "mechanism": {
        "representations": ["covering systems", "congruence classes"],
        "assumptions": ["finite covering suffices", "identities compose across classes"],
        "operators": ["congruence covering", "class union"],
        "preserves": ["polynomial-identity solvability per class"],
        "breaks": ["class-local isolation"],
        "auxiliary_objects": ["system of moduli"],
        "locality": "global",
        "construction_mode": "existential",
        "uncertainty_mode": "deterministic",
        "notes": "Seeks a finite union of solvable classes covering Z."
      },
      "outcome": {
        "class": "failure",
        "boundary_statement": "no finite covering eliminates quadratic-residue classes",
        "boundary_conditions": ["quadratic-residue survivor classes", "infinitely many survivor primes"],
        "notes": "Blocked by the quadratic-reciprocity obstruction, not by search effort."
      },
      "support": [
        {"field_path": "mechanism.operators", "support_kind": "explicit", "locator": "para:2"},
        {"field_path": "mechanism.breaks", "support_kind": "inferred", "locator": "para:2"},
        {"field_path": "outcome.class", "support_kind": "explicit", "locator": "para:4"},
        {"field_path": "outcome.boundary_statement", "support_kind": "explicit", "locator": "para:4"}
      ]
    }
  ]
}
newf-normalize -->
