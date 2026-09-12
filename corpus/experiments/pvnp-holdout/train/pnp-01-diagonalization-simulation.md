# Diagonalization and simulation program (pnp-01)

**Era:** 1960s–1975. Pre-cutoff.

The founding program attacks separation directly with the tools that built
complexity theory: **diagonalization** and **simulation**. The time hierarchy
theorem (Hartmanis–Stearns 1965) separates deterministic time classes by
constructing a machine that simulates and contradicts every machine in the
smaller class. After Cook (1971) and Karp (1972) located NP-completeness, the
natural hope was that the same engine — enumerate machines, simulate, diagonalize
against — would separate P from NP.

The mechanism treats machines as **black boxes to be simulated**: the
diagonalizing construction never inspects *how* a machine computes, only its
input–output behavior under a resource bound. This is the program's strength
(it is robust and fully rigorous) and, as pnp-02 records, exactly its wall.

The attempt fails structurally, not for lack of ingenuity: every argument in
this family **relativizes** — it goes through unchanged if all machines are
given the same oracle. Baker–Gill–Solovay (1975) constructed oracles making the
P-vs-NP question resolve both ways, so no relativizing argument can settle it.
The program's conserved property — black-box simulation — is precisely what the
wall condemns.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "pvnp/diagonalization-simulation",
      "label": "Diagonalization / simulation program",
      "description": "Separate P from NP by enumerating machines and diagonalizing against resource-bounded simulations, as in the time hierarchy theorem.",
      "mechanism": {
        "representations": ["Turing machine enumerations", "resource-bounded simulations"],
        "assumptions": ["black-box simulation suffices for separation", "hierarchy-theorem engine transfers to P vs NP"],
        "operators": ["diagonalization", "clocked simulation", "machine enumeration"],
        "preserves": ["black-box relativizing simulation", "model-analysis-first direction"],
        "breaks": [],
        "auxiliary_objects": ["universal machine", "enumeration of clocked machines"],
        "locality": "global",
        "construction_mode": "constructive",
        "uncertainty_mode": "deterministic",
        "notes": "Never inspects internal structure of the simulated machine; input-output behavior only."
      },
      "outcome": {
        "class": "failure",
        "boundary_statement": "every argument in this family relativizes, and contrary oracles exist for P vs NP",
        "boundary_conditions": ["argument unchanged under any shared oracle", "oracle A with P^A = NP^A and oracle B with P^B != NP^B both exist"],
        "notes": "Blocked by the relativization wall (pnp-02), not by search effort."
      },
      "support": [
        {"field_path": "mechanism.operators", "support_kind": "explicit", "locator": "para:1"},
        {"field_path": "mechanism.preserves", "support_kind": "explicit", "locator": "para:2"},
        {"field_path": "outcome.class", "support_kind": "explicit", "locator": "para:3"},
        {"field_path": "outcome.boundary_statement", "support_kind": "explicit", "locator": "para:3"}
      ]
    }
  ]
}
newf-normalize -->
