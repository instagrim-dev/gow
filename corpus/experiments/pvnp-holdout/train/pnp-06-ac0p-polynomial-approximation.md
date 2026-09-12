# AC⁰[p] — polynomial approximation over finite fields (pnp-06)

**Era:** 1987. Anchors: Razborov 1987 (majority vs parity gates);
Smolensky 1987 (algebraic method over F_p).

The circuit program's second beachhead extends pnp-03 past parity gates. For
constant-depth circuits with AND/OR/MOD-p gates (AC⁰[p], p prime), the
**algebraic approximation method** represents each gate by a low-degree
polynomial over F_p that agrees with it on most inputs; composing along
constant depth, every small circuit is approximated by a low-degree
polynomial. Smolensky showed MOD-q (q a different prime) is far from every
low-degree polynomial over F_p, giving exponential AC⁰[p] lower bounds.

The mechanistic signature is again the inherited posture: model analysis
first, with a **distinguishing property** — approximability by low-degree
F_p-polynomials — that every small circuit in the class has and the hard
function lacks. The property is once more broadly shared (a random function is
far from all low-degree polynomials) and certifiable by linear algebra.

The wall stands exactly one gate type away: for **composite** moduli (AC⁰[6]),
and a fortiori for TC⁰ and NC¹, no analogous property was found. Two decades
of effort produced no lower bound against ACC⁰ (constant depth, AND/OR/MOD-m
for composite m) for any function in NEXP — the sharpest statement of the
circuit program's stall as of 2004.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "pvnp/ac0p-polynomial-approximation",
      "label": "AC0[p] algebraic approximation",
      "description": "Exponential lower bounds for constant-depth circuits with MOD-p gates by approximating gates with low-degree polynomials over F_p.",
      "mechanism": {
        "representations": ["AC0[p] circuits", "low-degree polynomials over F_p"],
        "assumptions": ["gate-wise polynomial approximation composes along constant depth", "hard function is far from all low-degree polynomials"],
        "operators": ["polynomial approximation", "degree counting", "distance-from-polynomials argument"],
        "preserves": ["model-analysis-first direction", "large constructive distinguishing property"],
        "breaks": ["black-box relativizing simulation"],
        "auxiliary_objects": ["approximating polynomial ensemble", "field F_p"],
        "locality": "local",
        "construction_mode": "constructive",
        "uncertainty_mode": "probabilistic",
        "notes": "Distinguishing property: low-degree approximability; shared by most functions, certifiable by linear algebra."
      },
      "outcome": {
        "class": "partial_success",
        "boundary_statement": "exponential AC0[p] bounds for prime p; no analogous property found for composite moduli — ACC0 unseparated from NEXP as of 2004",
        "boundary_conditions": ["prime field structure essential to the approximation", "composite-modulus gates defeat the polynomial representation"],
        "notes": "The stall at ACC0 is the sharpest open edge of the circuit program as of 2004."
      },
      "support": [
        {"field_path": "mechanism.operators", "support_kind": "explicit", "locator": "para:1"},
        {"field_path": "mechanism.preserves", "support_kind": "explicit", "locator": "para:2"},
        {"field_path": "outcome.class", "support_kind": "explicit", "locator": "para:1"},
        {"field_path": "outcome.boundary_statement", "support_kind": "explicit", "locator": "para:3"}
      ]
    }
  ]
}
newf-normalize -->
