# Natural-proofs barrier — largeness + constructivity vs PRFs (pnp-08)

**Era:** 1994. Pre-cutoff. Anchor: Razborov–Rudich, *Natural proofs*,
STOC 1994 / JCSS 55(1), 1997.

The second wall, and this one condemns the circuit program's own posture.
Razborov and Rudich observed that every known circuit lower bound (pnp-03,
pnp-04, pnp-06) proceeds through a **natural property**: a set of Boolean
functions that is (1) **constructive** — membership decidable in time
polynomial in the truth-table size — and (2) **large** — containing a
non-negligible fraction of all functions — and that contains the hard function
while excluding everything the small circuit class computes.

The barrier: if strong pseudorandom functions exist in the target class (as
widely believed for TC⁰ and above under standard cryptographic assumptions),
then **no natural property can separate that class from hard functions** — a
constructive, large property that rejects everything the class computes would
itself be a distinguisher breaking the PRF. The very features that made the
train methods checkable and generalizable (broadly shared, efficiently
certifiable distinguishing properties) are what the wall forbids.

The negative space is again a map: a successful argument against strong
classes must be **non-natural** — either non-constructive, or targeted at
rare/specific structure rather than a large property. As of the cutoff, no
known lower-bound technique for strong classes lived in that space; the
program did not know how to open the box without being natural about it.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "pvnp/natural-proofs-barrier",
      "label": "Natural-proofs barrier",
      "description": "Under standard cryptographic assumptions, no constructive large distinguishing property separates strong circuit classes from hard functions.",
      "mechanism": {
        "representations": ["natural properties over truth tables", "pseudorandom function families"],
        "assumptions": ["strong PRFs exist in the target class"],
        "operators": ["property-to-distinguisher reduction"],
        "preserves": ["model-analysis-first direction"],
        "breaks": [],
        "auxiliary_objects": ["pseudorandom function family", "truth-table property tester"],
        "locality": "global",
        "construction_mode": "existential",
        "uncertainty_mode": "probabilistic",
        "notes": "A meta-result condemning the conserved property of pnp-03/04/06: large constructive distinguishing properties."
      },
      "outcome": {
        "class": "failure",
        "boundary_statement": "no natural (constructive and large) property proves lower bounds against classes containing strong PRFs",
        "boundary_conditions": ["constructivity: property decidable in poly(truth-table) time", "largeness: property holds for a non-negligible fraction of functions", "cryptographic hardness assumption in the target class"],
        "notes": "Escape space is explicitly mapped: non-constructive or rare/specific properties; unoccupied at cutoff."
      },
      "support": [
        {"field_path": "mechanism.assumptions", "support_kind": "explicit", "locator": "para:2"},
        {"field_path": "outcome.boundary_statement", "support_kind": "explicit", "locator": "para:2"},
        {"field_path": "outcome.boundary_conditions", "support_kind": "explicit", "locator": "para:1"},
        {"field_path": "outcome.notes", "support_kind": "explicit", "locator": "para:3"}
      ]
    }
  ]
}
newf-normalize -->
