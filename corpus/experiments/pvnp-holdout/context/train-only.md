# pvnp-holdout permitted context: TRAIN ATLAS ONLY

Twelve recorded research programs on P vs NP and their outcomes. This is the complete atlas available to the proposer. No other material about this problem's research history is permitted context.

---

# Diagonalization and simulation program (pnp-01)

**Era:** 1960s–1975.

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
        "representations": [
          "Turing machine enumerations",
          "resource-bounded simulations"
        ],
        "assumptions": [
          "black-box simulation suffices for separation",
          "hierarchy-theorem engine transfers to P vs NP"
        ],
        "operators": [
          "diagonalization",
          "clocked simulation",
          "machine enumeration"
        ],
        "preserves": [
          "black-box relativizing simulation",
          "model-analysis-first direction"
        ],
        "breaks": [],
        "auxiliary_objects": [
          "universal machine",
          "enumeration of clocked machines"
        ],
        "locality": "global",
        "construction_mode": "constructive",
        "uncertainty_mode": "deterministic",
        "notes": "Never inspects internal structure of the simulated machine; input-output behavior only.",
        "completeness_scope": "declared_payload",
        "completeness_basis": "the embedded newf-normalize block is the complete declared payload for this corpus file: every representation/assumption/operator/preserves/breaks/auxiliary_object entry the file states is listed here and parsed by construction (deterministic embedded-payload parser). Author-attested 2026-09-12 as part of pvnp-holdout corpus preparation, before freeze and before any capture existed; declared_payload scope, NOT a mechanism-exhaustive extraction judgment.",
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
        "class": "failure",
        "boundary_statement": "every argument in this family relativizes, and contrary oracles exist for P vs NP",
        "boundary_conditions": [
          "argument unchanged under any shared oracle",
          "oracle A with P^A = NP^A and oracle B with P^B != NP^B both exist"
        ],
        "notes": "Blocked by the relativization wall (pnp-02), not by search effort."
      },
      "support": [
        {
          "field_path": "mechanism.operators",
          "support_kind": "explicit",
          "locator": "para:1"
        },
        {
          "field_path": "mechanism.preserves",
          "support_kind": "explicit",
          "locator": "para:2"
        },
        {
          "field_path": "outcome.class",
          "support_kind": "explicit",
          "locator": "para:3"
        },
        {
          "field_path": "outcome.boundary_statement",
          "support_kind": "explicit",
          "locator": "para:3"
        }
      ]
    }
  ]
}
newf-normalize -->

---

# Relativization barrier — contrary oracles (pnp-02)

**Era:** 1975. Anchor: Baker–Gill–Solovay, *Relativizations of the
P =? NP question*, SIAM J. Comput. 4(4), 1975.

This entry records the **wall itself** as a first-class failure structure.
Baker, Gill, and Solovay constructed an oracle `A` with `P^A = NP^A` and an
oracle `B` with `P^B ≠ NP^B`. Any proof technique that **relativizes** — that
remains valid when every machine in the argument is given the same oracle —
therefore cannot resolve P vs NP in either direction.

The barrier is a statement about a *conserved mechanism property*, not about
any single attempt: diagonalization, simulation, padding, and translation
arguments all treat computation as **black-box query behavior**, and black-box
query behavior is exactly what an oracle re-defines. The wall condemns the
property, and with it every method that preserves the property.

What the barrier does **not** say: it does not forbid separation proofs; it
forbids separation proofs that never open the box. Techniques that exploit the
internal, non-black-box structure of computation — completeness under local
reductions, the structure of specific circuit classes — are outside its scope.
That negative space is the barrier's own map of where a productive method must
live.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "pvnp/relativization-barrier",
      "label": "Relativization barrier (contrary oracles)",
      "description": "Contrary oracles show that no relativizing argument settles P vs NP; the black-box simulation property itself is condemned.",
      "mechanism": {
        "representations": [
          "oracle machines",
          "contrary oracle constructions"
        ],
        "assumptions": [
          "technique validity is preserved under a shared oracle"
        ],
        "operators": [
          "oracle construction",
          "stage-wise diagonalization against query behavior"
        ],
        "preserves": [
          "black-box relativizing simulation",
          "model-analysis-first direction"
        ],
        "breaks": [],
        "auxiliary_objects": [
          "oracle A with P^A = NP^A",
          "oracle B with P^B != NP^B"
        ],
        "locality": "global",
        "construction_mode": "constructive",
        "uncertainty_mode": "deterministic",
        "notes": "A meta-result: classifies which mechanism property makes an attempt hopeless.",
        "completeness_scope": "declared_payload",
        "completeness_basis": "the embedded newf-normalize block is the complete declared payload for this corpus file: every representation/assumption/operator/preserves/breaks/auxiliary_object entry the file states is listed here and parsed by construction (deterministic embedded-payload parser). Author-attested 2026-09-12 as part of pvnp-holdout corpus preparation, before freeze and before any capture existed; declared_payload scope, NOT a mechanism-exhaustive extraction judgment.",
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
        "class": "failure",
        "boundary_statement": "no relativizing argument can resolve P vs NP",
        "boundary_conditions": [
          "argument must survive substitution of any shared oracle",
          "both equality and separation oracles exist"
        ],
        "notes": "The wall condemns black-box query behavior; non-black-box structure is explicitly outside its scope."
      },
      "support": [
        {
          "field_path": "mechanism.auxiliary_objects",
          "support_kind": "explicit",
          "locator": "para:1"
        },
        {
          "field_path": "mechanism.preserves",
          "support_kind": "explicit",
          "locator": "para:2"
        },
        {
          "field_path": "outcome.boundary_statement",
          "support_kind": "explicit",
          "locator": "para:1"
        },
        {
          "field_path": "outcome.notes",
          "support_kind": "explicit",
          "locator": "para:3"
        }
      ]
    }
  ]
}
newf-normalize -->

---

# AC⁰ lower bounds — parity via random restrictions (pnp-03)

**Era:** 1981–1986. Anchors: Ajtai 1983; Furst–Saxe–Sipser 1984;
Håstad 1986 (switching lemma).

The first escape from the relativization wall: **open the box**. Instead of
simulating machines, analyze the internal structure of a concrete
computational model — constant-depth, polynomial-size circuits (AC⁰) — and
prove that parity is not computable there.

The mechanism is the **random restriction**: fix a random subset of inputs to
random constants. Håstad's switching lemma shows a restricted small-width DNF
collapses to a small-depth decision tree with high probability; applying
restrictions level by level collapses a constant-depth circuit entirely, while
parity survives every restriction as parity of the live variables. The
contradiction gives exponential AC⁰ lower bounds.

This is a genuine partial success and its structure matters: the argument is
**non-relativizing** (it depends on circuit anatomy, not query behavior), it
proceeds by **model analysis first** — pick a weak circuit class, find a
distinguishing property, show a hard function lacks it — and its central
property (simplification under random restriction) holds for *most* functions
and is *efficiently certifiable*. The program's next decades try to push the
same posture up the circuit hierarchy; pnp-08 records where that posture
itself becomes the obstacle.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "pvnp/ac0-random-restrictions",
      "label": "AC0 lower bounds via random restrictions",
      "description": "Prove parity is outside AC0 by collapsing constant-depth circuits with random restrictions (switching lemma) while parity survives restriction.",
      "mechanism": {
        "representations": [
          "constant-depth circuits",
          "random restrictions",
          "decision trees"
        ],
        "assumptions": [
          "weak-class internal structure is analyzable",
          "hard function survives restriction"
        ],
        "operators": [
          "random restriction",
          "switching lemma",
          "depth reduction"
        ],
        "preserves": [
          "model-analysis-first direction",
          "large constructive distinguishing property"
        ],
        "breaks": [
          "black-box relativizing simulation"
        ],
        "auxiliary_objects": [
          "restriction distribution",
          "collapsed decision tree"
        ],
        "locality": "local",
        "construction_mode": "constructive",
        "uncertainty_mode": "probabilistic",
        "notes": "The distinguishing property (collapse under restriction) holds for most functions and is efficiently certifiable.",
        "completeness_scope": "declared_payload",
        "completeness_basis": "the embedded newf-normalize block is the complete declared payload for this corpus file: every representation/assumption/operator/preserves/breaks/auxiliary_object entry the file states is listed here and parsed by construction (deterministic embedded-payload parser). Author-attested 2026-09-12 as part of pvnp-holdout corpus preparation, before freeze and before any capture existed; declared_payload scope, NOT a mechanism-exhaustive extraction judgment.",
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
        "boundary_statement": "exponential lower bounds for AC0, but the method stalls below classes with parity gates or full depth",
        "boundary_conditions": [
          "restriction argument requires depth to be constant",
          "property is large and constructive (see pnp-08)"
        ],
        "notes": "First non-relativizing beachhead; posture inherited by the whole circuit program."
      },
      "support": [
        {
          "field_path": "mechanism.operators",
          "support_kind": "explicit",
          "locator": "para:2"
        },
        {
          "field_path": "mechanism.breaks",
          "support_kind": "explicit",
          "locator": "para:3"
        },
        {
          "field_path": "outcome.class",
          "support_kind": "explicit",
          "locator": "para:3"
        },
        {
          "field_path": "outcome.boundary_statement",
          "support_kind": "inferred",
          "locator": "para:3"
        }
      ]
    }
  ]
}
newf-normalize -->

---

# Monotone circuit lower bounds — the approximation method (pnp-04)

**Era:** 1985. Anchor: Razborov 1985 (superpolynomial monotone
lower bounds for clique).

Razborov's **approximation method** attacks a restricted but natural model:
monotone circuits (AND/OR only, no negation). Replace each gate of a purported
small monotone circuit for CLIQUE with an "approximating" gate drawn from a
structured lattice of simple set systems (sunflower-plucked positive
combinations). Each replacement introduces few errors on the chosen test
inputs; a small circuit therefore computes something close to a simple
function — but CLIQUE is provably far from every simple function on those test
distributions. The contradiction yields superpolynomial (later exponential)
monotone lower bounds for an NP-complete function.

The mechanistic posture is inherited from pnp-03 and sharpened: **model
analysis first** — fix the circuit class, build a *distinguishing property*
(approximability by the lattice of simple functions), show every small circuit
in the class has it and the hard function does not. The property is again
broadly shared (most monotone functions are hard this way) and certifiable by
combinatorial computation.

The result was the strongest concrete evidence yet that circuit analysis could
reach NP-completeness. Its limitation surfaced immediately — pnp-05 records
the failed hope of removing the monotonicity restriction.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "pvnp/monotone-approximation-method",
      "label": "Monotone approximation method",
      "description": "Superpolynomial monotone lower bounds for clique by gate-wise approximation within a lattice of simple set systems.",
      "mechanism": {
        "representations": [
          "monotone circuits",
          "lattice of approximator functions",
          "sunflower systems"
        ],
        "assumptions": [
          "gate-wise approximation errors stay controllable",
          "test distributions separate hard function from approximators"
        ],
        "operators": [
          "gate-by-gate approximation",
          "sunflower lemma",
          "error counting on test inputs"
        ],
        "preserves": [
          "model-analysis-first direction",
          "large constructive distinguishing property"
        ],
        "breaks": [
          "black-box relativizing simulation"
        ],
        "auxiliary_objects": [
          "approximator lattice",
          "positive/negative test distributions"
        ],
        "locality": "local",
        "construction_mode": "constructive",
        "uncertainty_mode": "deterministic",
        "notes": "Distinguishing property: approximability by simple functions; broadly shared and combinatorially certifiable.",
        "completeness_scope": "declared_payload",
        "completeness_basis": "the embedded newf-normalize block is the complete declared payload for this corpus file: every representation/assumption/operator/preserves/breaks/auxiliary_object entry the file states is listed here and parsed by construction (deterministic embedded-payload parser). Author-attested 2026-09-12 as part of pvnp-holdout corpus preparation, before freeze and before any capture existed; declared_payload scope, NOT a mechanism-exhaustive extraction judgment.",
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
        "boundary_statement": "superpolynomial monotone lower bounds for an NP-complete function; silent on general circuits",
        "boundary_conditions": [
          "negation-free model only",
          "approximator lattice exploits monotonicity essentially"
        ],
        "notes": "Strongest evidence to date that circuit analysis reaches NP-hardness; the monotone restriction is load-bearing (pnp-05)."
      },
      "support": [
        {
          "field_path": "mechanism.operators",
          "support_kind": "explicit",
          "locator": "para:1"
        },
        {
          "field_path": "mechanism.preserves",
          "support_kind": "explicit",
          "locator": "para:2"
        },
        {
          "field_path": "outcome.class",
          "support_kind": "explicit",
          "locator": "para:1"
        },
        {
          "field_path": "outcome.boundary_statement",
          "support_kind": "explicit",
          "locator": "para:3"
        }
      ]
    }
  ]
}
newf-normalize -->

---

# The monotone→general hope and its failure (pnp-05)

**Era:** 1985–1988. Anchors: Razborov 1985 (perfect matching has
superpolynomial monotone complexity); Tardos 1988 (exponential monotone/general
gap, Combinatorica).

After pnp-04, the natural program was **transfer**: if monotone lower bounds
can reach NP-complete functions, perhaps monotone complexity tracks general
complexity closely enough that monotone bounds for the right function separate
P from NP.

The hope failed on a concrete counterexample family. Razborov (1985) proved
superpolynomial monotone lower bounds for **perfect matching** — a problem in
P. Tardos (1988) sharpened the separation to exponential: a monotone function
computable by polynomial-size general circuits requires exponential-size
monotone circuits. Monotone complexity and general complexity are therefore
**exponentially divorced**, and no bound proven inside the negation-free model
transfers to the unrestricted question.

The failure's structure is informative: the approximation method's power came
from an auxiliary restriction (no negations) that also *severed the bridge
back* to the real question. The conserved posture — prove hardness inside an
analyzable restricted model, then hope the restriction is inessential — is
recorded here as failing at the transfer step, with the restriction itself as
the boundary.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "pvnp/monotone-transfer-hope",
      "label": "Monotone-to-general transfer hope",
      "description": "Leverage monotone lower bounds into general-circuit separations by treating the negation-free restriction as inessential.",
      "mechanism": {
        "representations": [
          "monotone circuits",
          "general circuits",
          "slice/transfer arguments"
        ],
        "assumptions": [
          "monotone complexity approximates general complexity on monotone functions"
        ],
        "operators": [
          "restriction lifting",
          "complexity transfer"
        ],
        "preserves": [
          "model-analysis-first direction",
          "large constructive distinguishing property"
        ],
        "breaks": [],
        "auxiliary_objects": [
          "perfect matching function",
          "Tardos function"
        ],
        "locality": "local",
        "construction_mode": "constructive",
        "uncertainty_mode": "deterministic",
        "notes": "The transfer step, not the lower-bound step, is where the program dies.",
        "completeness_scope": "declared_payload",
        "completeness_basis": "the embedded newf-normalize block is the complete declared payload for this corpus file: every representation/assumption/operator/preserves/breaks/auxiliary_object entry the file states is listed here and parsed by construction (deterministic embedded-payload parser). Author-attested 2026-09-12 as part of pvnp-holdout corpus preparation, before freeze and before any capture existed; declared_payload scope, NOT a mechanism-exhaustive extraction judgment.",
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
        "class": "failure",
        "boundary_statement": "monotone and general complexity are exponentially separated, so negation-free bounds do not transfer",
        "boundary_conditions": [
          "matching in P has superpolynomial monotone complexity",
          "exponential monotone/general gap exists"
        ],
        "notes": "The auxiliary restriction that made analysis possible also severed the bridge back to P vs NP."
      },
      "support": [
        {
          "field_path": "mechanism.assumptions",
          "support_kind": "explicit",
          "locator": "para:1"
        },
        {
          "field_path": "outcome.class",
          "support_kind": "explicit",
          "locator": "para:2"
        },
        {
          "field_path": "outcome.boundary_statement",
          "support_kind": "explicit",
          "locator": "para:2"
        },
        {
          "field_path": "outcome.notes",
          "support_kind": "explicit",
          "locator": "para:3"
        }
      ]
    }
  ]
}
newf-normalize -->

---

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
        "representations": [
          "AC0[p] circuits",
          "low-degree polynomials over F_p"
        ],
        "assumptions": [
          "gate-wise polynomial approximation composes along constant depth",
          "hard function is far from all low-degree polynomials"
        ],
        "operators": [
          "polynomial approximation",
          "degree counting",
          "distance-from-polynomials argument"
        ],
        "preserves": [
          "model-analysis-first direction",
          "large constructive distinguishing property"
        ],
        "breaks": [
          "black-box relativizing simulation"
        ],
        "auxiliary_objects": [
          "approximating polynomial ensemble",
          "field F_p"
        ],
        "locality": "local",
        "construction_mode": "constructive",
        "uncertainty_mode": "probabilistic",
        "notes": "Distinguishing property: low-degree approximability; shared by most functions, certifiable by linear algebra.",
        "completeness_scope": "declared_payload",
        "completeness_basis": "the embedded newf-normalize block is the complete declared payload for this corpus file: every representation/assumption/operator/preserves/breaks/auxiliary_object entry the file states is listed here and parsed by construction (deterministic embedded-payload parser). Author-attested 2026-09-12 as part of pvnp-holdout corpus preparation, before freeze and before any capture existed; declared_payload scope, NOT a mechanism-exhaustive extraction judgment.",
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
        "boundary_statement": "exponential AC0[p] bounds for prime p; no analogous property found for composite moduli — ACC0 unseparated from NEXP as of 2004",
        "boundary_conditions": [
          "prime field structure essential to the approximation",
          "composite-modulus gates defeat the polynomial representation"
        ],
        "notes": "The stall at ACC0 is the sharpest open edge of the circuit program as of 2004."
      },
      "support": [
        {
          "field_path": "mechanism.operators",
          "support_kind": "explicit",
          "locator": "para:1"
        },
        {
          "field_path": "mechanism.preserves",
          "support_kind": "explicit",
          "locator": "para:2"
        },
        {
          "field_path": "outcome.class",
          "support_kind": "explicit",
          "locator": "para:1"
        },
        {
          "field_path": "outcome.boundary_statement",
          "support_kind": "explicit",
          "locator": "para:3"
        }
      ]
    }
  ]
}
newf-normalize -->

---

# Proof-complexity program — resolution lower bounds (pnp-07)

**Era:** 1985–2004. Anchors: Haken 1985 (exponential resolution
lower bounds for the pigeonhole principle); Beame–Pitassi 1998 (survey,
*Propositional proof complexity: past, present, future*).

A structurally different flank: instead of bounding the circuits that *decide*
NP problems, bound the proofs that *certify* their answers. Cook–Reckhow
reduce NP vs coNP to proof systems: if every propositional proof system has
hard tautologies, then NP ≠ coNP (and hence P ≠ NP). The program climbs the
proof-system hierarchy the way the circuit program climbs circuit classes.

Haken's 1985 exponential lower bound for **resolution** on the pigeonhole
principle was the breakthrough, via a bottleneck/width argument that is
mechanically kin to the circuit methods: identify a *distinguishing property*
of all short resolution proofs (narrowness / bottleneck counting) that the
target tautology's proofs cannot have. Later work extended to restricted
subsystems, but bounded-depth Frege with counting gates, and full Frege above
it, resisted — the strong systems stand unseparated as of 2004, in direct
parallel with the ACC⁰ stall of pnp-06.

The parallel is the informative structure: two independent programs, both
model-analysis-first, both powered by large certifiable distinguishing
properties, both stalled at the level where the analyzed object becomes
expressive enough to *count*.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "pvnp/proof-complexity-resolution",
      "label": "Proof-complexity program (resolution lower bounds)",
      "description": "Attack NP vs coNP by proving superpolynomial proof-size lower bounds system-by-system, starting from resolution on the pigeonhole principle.",
      "mechanism": {
        "representations": [
          "propositional proof systems",
          "resolution refutations",
          "tautology families"
        ],
        "assumptions": [
          "per-system lower bounds accumulate toward all-systems hardness",
          "width/bottleneck properties certify proof size"
        ],
        "operators": [
          "bottleneck counting",
          "width lower bounds",
          "restriction on proofs"
        ],
        "preserves": [
          "model-analysis-first direction",
          "large constructive distinguishing property"
        ],
        "breaks": [
          "black-box relativizing simulation"
        ],
        "auxiliary_objects": [
          "pigeonhole tautologies",
          "proof-width measure"
        ],
        "locality": "local",
        "construction_mode": "constructive",
        "uncertainty_mode": "deterministic",
        "notes": "Mechanically kin to the circuit program: distinguishing property over all short proofs in a fixed system.",
        "completeness_scope": "declared_payload",
        "completeness_basis": "the embedded newf-normalize block is the complete declared payload for this corpus file: every representation/assumption/operator/preserves/breaks/auxiliary_object entry the file states is listed here and parsed by construction (deterministic embedded-payload parser). Author-attested 2026-09-12 as part of pvnp-holdout corpus preparation, before freeze and before any capture existed; declared_payload scope, NOT a mechanism-exhaustive extraction judgment.",
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
        "boundary_statement": "exponential bounds for resolution and weak subsystems; Frege and stronger systems unseparated as of 2004",
        "boundary_conditions": [
          "method stalls where the proof system can count",
          "per-system progress does not compose into all-systems claims"
        ],
        "notes": "Stall level parallels the circuit program's ACC0 edge: expressiveness sufficient for counting defeats the known properties."
      },
      "support": [
        {
          "field_path": "mechanism.operators",
          "support_kind": "explicit",
          "locator": "para:2"
        },
        {
          "field_path": "mechanism.preserves",
          "support_kind": "inferred",
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
          "locator": "para:2"
        }
      ]
    }
  ]
}
newf-normalize -->

---

# Natural-proofs barrier — largeness + constructivity vs PRFs (pnp-08)

**Era:** 1994. Anchor: Razborov–Rudich, *Natural proofs*,
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
known lower-bound methods checkable and generalizable (broadly shared, efficiently
certifiable distinguishing properties) are what the wall forbids.

The negative space is again a map: a successful argument against strong
classes must be **non-natural** — either non-constructive, or targeted at
rare/specific structure rather than a large property. As of 2004, no
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
        "representations": [
          "natural properties over truth tables",
          "pseudorandom function families"
        ],
        "assumptions": [
          "strong PRFs exist in the target class"
        ],
        "operators": [
          "property-to-distinguisher reduction"
        ],
        "preserves": [
          "model-analysis-first direction"
        ],
        "breaks": [],
        "auxiliary_objects": [
          "pseudorandom function family",
          "truth-table property tester"
        ],
        "locality": "global",
        "construction_mode": "existential",
        "uncertainty_mode": "probabilistic",
        "notes": "A meta-result condemning the conserved property of pnp-03/04/06: large constructive distinguishing properties.",
        "completeness_scope": "declared_payload",
        "completeness_basis": "the embedded newf-normalize block is the complete declared payload for this corpus file: every representation/assumption/operator/preserves/breaks/auxiliary_object entry the file states is listed here and parsed by construction (deterministic embedded-payload parser). Author-attested 2026-09-12 as part of pvnp-holdout corpus preparation, before freeze and before any capture existed; declared_payload scope, NOT a mechanism-exhaustive extraction judgment.",
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
        "class": "failure",
        "boundary_statement": "no natural (constructive and large) property proves lower bounds against classes containing strong PRFs",
        "boundary_conditions": [
          "constructivity: property decidable in poly(truth-table) time",
          "largeness: property holds for a non-negligible fraction of functions",
          "cryptographic hardness assumption in the target class"
        ],
        "notes": "Escape space is explicitly mapped: non-constructive or rare/specific properties; unoccupied as of 2004."
      },
      "support": [
        {
          "field_path": "mechanism.assumptions",
          "support_kind": "explicit",
          "locator": "para:2"
        },
        {
          "field_path": "outcome.boundary_statement",
          "support_kind": "explicit",
          "locator": "para:2"
        },
        {
          "field_path": "outcome.boundary_conditions",
          "support_kind": "explicit",
          "locator": "para:1"
        },
        {
          "field_path": "outcome.notes",
          "support_kind": "explicit",
          "locator": "para:3"
        }
      ]
    }
  ]
}
newf-normalize -->

---

# The circuit program's stall after natural proofs (pnp-09)

**Era:** 1994–2004. Anchors: Razborov 1995 (*Unprovability of
lower bounds…*, bounded-arithmetic independence); contemporaneous surveys of
the post-natural-proofs decade.

This entry records the decade of **conserved failure** after pnp-08, because
the absence of movement is itself structure. Between 1994 and 2004, no
lower bound was proven against any circuit class at or above ACC⁰ for any
function in NEXP, despite the escape space being explicitly mapped
(non-constructive or rare properties).

Attempts in the period conserved the old posture in new vocabulary:
derandomization-flavored properties, extensions of the polynomial method to
composite moduli, and formal-independence results (Razborov 1995: relevant
lower-bound statements are unprovable in certain bounded-arithmetic systems
whose reasoning power mirrors the natural-proofs frame). Each either
re-encountered the largeness/constructivity trap or retreated to restricted
models where the trap does not bind.

The candidate invariant this decade sharpens: **the direction of analysis was
always model→function** — fix the circuit class, seek a property over truth
tables that all small circuits in it respect. Both walls (pnp-02, pnp-08) are
statements about *that direction*: relativization condemns black-box
simulation of the model, natural proofs condemns large certifiable properties
of the model. The decade's attempts vary the property and vary the model, but
the analyzed object is the model in every case.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "pvnp/post-natural-proofs-stall",
      "label": "Post-natural-proofs stall (1994-2004)",
      "description": "A decade of attempts against ACC0 and above that either re-encounter the natural-proofs trap or retreat to restricted models; no NEXP lower bound as of 2004.",
      "mechanism": {
        "representations": [
          "truth-table properties",
          "restricted circuit models",
          "bounded-arithmetic systems"
        ],
        "assumptions": [
          "a property-based route exists that evades largeness or constructivity"
        ],
        "operators": [
          "property search",
          "polynomial-method extension",
          "independence proof"
        ],
        "preserves": [
          "model-analysis-first direction",
          "large constructive distinguishing property"
        ],
        "breaks": [],
        "auxiliary_objects": [
          "bounded-arithmetic fragments"
        ],
        "locality": "local",
        "construction_mode": "existential",
        "uncertainty_mode": "deterministic",
        "notes": "Absence of movement recorded as structure: the model-to-function direction of analysis is conserved across every attempt of the decade.",
        "completeness_scope": "declared_payload",
        "completeness_basis": "the embedded newf-normalize block is the complete declared payload for this corpus file: every representation/assumption/operator/preserves/breaks/auxiliary_object entry the file states is listed here and parsed by construction (deterministic embedded-payload parser). Author-attested 2026-09-12 as part of pvnp-holdout corpus preparation, before freeze and before any capture existed; declared_payload scope, NOT a mechanism-exhaustive extraction judgment.",
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
        "class": "failure",
        "boundary_statement": "no lower bound against ACC0 or above for any NEXP function in the decade after the natural-proofs barrier",
        "boundary_conditions": [
          "attempts conserve the model-to-function analysis direction",
          "mapped escape space (non-constructive or rare properties) unoccupied"
        ],
        "notes": "Attempts vary the property and the model; the analyzed object is the model in every case."
      },
      "support": [
        {
          "field_path": "mechanism.preserves",
          "support_kind": "explicit",
          "locator": "para:3"
        },
        {
          "field_path": "outcome.class",
          "support_kind": "explicit",
          "locator": "para:1"
        },
        {
          "field_path": "outcome.boundary_statement",
          "support_kind": "explicit",
          "locator": "para:1"
        },
        {
          "field_path": "outcome.notes",
          "support_kind": "explicit",
          "locator": "para:3"
        }
      ]
    }
  ]
}
newf-normalize -->

---

# Geometric complexity theory — orbit closures and representation theory (pnp-10)

**Era:** 2001. Anchor: Mulmuley–Sohoni, *Geometric complexity
theory I*, SIAM J. Comput. 31(2), 2001.

GCT is the field's most radical re-representation: recast permanent vs
determinant (an algebraic proxy for NP vs P) as a question about **orbit
closures** under the general linear group action, and seek **representation-
theoretic obstructions** — irreducible modules whose multiplicities
distinguish the two orbit closures.

Mechanistically, GCT is a deliberate response to pnp-08: the sought
obstructions are *rare, highly structured* objects (specific irreducible
representations), not large properties over random truth tables — the program
is explicitly designed to be non-natural. It is also non-relativizing, since
it works with the algebraic anatomy of specific polynomials.

Its recorded status as of 2004 is a **program, not a result**: no
unconditional lower bound for any explicit function had been derived from it,
and the representation-theoretic multiplicity problems it reduces to
(Kronecker coefficients, plethysm) were — and were understood to be — deep
open problems themselves. The structural lesson: escaping
both walls is *possible in posture* (rare, structured, non-black-box), but the
only occupant of that space traded the walls for open problems at
least as hard.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "pvnp/gct-orbit-closures",
      "label": "Geometric complexity theory (orbit-closure obstructions)",
      "description": "Recast permanent vs determinant as orbit-closure separation and seek rare representation-theoretic obstructions, deliberately evading naturalness.",
      "mechanism": {
        "representations": [
          "orbit closures under GL-action",
          "irreducible representations",
          "multiplicity obstructions"
        ],
        "assumptions": [
          "obstruction multiplicities are computable or boundable",
          "algebraic proxy transfers to Boolean separation"
        ],
        "operators": [
          "symmetry classification",
          "multiplicity comparison",
          "algebraic-geometric degeneration"
        ],
        "preserves": [
          "model-analysis-first direction"
        ],
        "breaks": [
          "black-box relativizing simulation",
          "large constructive distinguishing property"
        ],
        "auxiliary_objects": [
          "permanent and determinant orbit closures",
          "Kronecker/plethysm coefficients"
        ],
        "locality": "global",
        "construction_mode": "existential",
        "uncertainty_mode": "deterministic",
        "notes": "Explicitly engineered to be non-natural: obstructions are rare and structured, not large properties.",
        "completeness_scope": "declared_payload",
        "completeness_basis": "the embedded newf-normalize block is the complete declared payload for this corpus file: every representation/assumption/operator/preserves/breaks/auxiliary_object entry the file states is listed here and parsed by construction (deterministic embedded-payload parser). Author-attested 2026-09-12 as part of pvnp-holdout corpus preparation, before freeze and before any capture existed; declared_payload scope, NOT a mechanism-exhaustive extraction judgment.",
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
        "boundary_statement": "a posture that evades both walls exists, but yields no unconditional bound and reduces to open representation-theoretic problems",
        "boundary_conditions": [
          "Kronecker/plethysm positivity problems open",
          "algebraic-to-Boolean transfer unestablished"
        ],
        "notes": "Occupies the mapped escape space in posture only; no unconditional bound extracted from it."
      },
      "support": [
        {
          "field_path": "mechanism.breaks",
          "support_kind": "explicit",
          "locator": "para:2"
        },
        {
          "field_path": "mechanism.representations",
          "support_kind": "explicit",
          "locator": "para:1"
        },
        {
          "field_path": "outcome.class",
          "support_kind": "explicit",
          "locator": "para:3"
        },
        {
          "field_path": "outcome.boundary_statement",
          "support_kind": "explicit",
          "locator": "para:3"
        }
      ]
    }
  ]
}
newf-normalize -->

---

# Derandomization ↔ lower bounds bridge (pnp-11)

**Era:** 2003. Anchor: Kabanets–Impagliazzo, *Derandomizing
polynomial identity tests means proving circuit lower bounds*, STOC 2003.

A converse to the hardness-vs-randomness program. That program
(Nisan–Wigderson and successors) had long run one way: lower bounds ⇒
pseudorandom generators ⇒ derandomization. Kabanets
and Impagliazzo proved a **converse**: derandomizing polynomial identity
testing (placing PIT in NSUBEXP) implies circuit lower bounds — either NEXP ⊄
P/poly or the permanent lacks polynomial-size arithmetic circuits.

The mechanism is an **algorithm-to-hardness implication**: assume a good
*algorithm* exists (deterministic PIT), combine it with completeness and
hierarchy machinery, and conclude that some explicit function is hard. The
distinguishing property here is not a truth-table property at all — the
argument is a conditional implication between an algorithmic event and a
lower-bound event, evading the natural-proofs frame (nothing large or
constructive is exhibited over random functions).

Its recorded boundary as of 2004: the bridge is **conditional in the wrong
direction for separation** — it converts an unproven derandomization into an
unproven lower bound, trading one open problem for another. The hypothesis
it consumes — a deterministic subexponential identity test — was itself an
open algorithmic problem, so no unconditional consequence had been extracted.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "pvnp/derandomization-lower-bound-bridge",
      "label": "Derandomization-to-lower-bounds bridge",
      "description": "Prove that derandomizing polynomial identity testing implies circuit lower bounds, inverting the classical hardness-to-randomness direction.",
      "mechanism": {
        "representations": [
          "identity-testing algorithms",
          "arithmetic circuits",
          "conditional implications"
        ],
        "assumptions": [
          "deterministic subexponential PIT exists (hypothesis, not theorem)"
        ],
        "operators": [
          "algorithm-to-hardness implication",
          "completeness leverage",
          "hierarchy-theorem coupling"
        ],
        "preserves": [],
        "breaks": [
          "black-box relativizing simulation",
          "large constructive distinguishing property",
          "model-analysis-first direction"
        ],
        "auxiliary_objects": [
          "polynomial identity tests",
          "permanent as hard candidate"
        ],
        "locality": "global",
        "construction_mode": "existential",
        "uncertainty_mode": "deterministic",
        "notes": "Reverses the classical analysis direction: from an algorithm's existence to a hardness conclusion. Conditional only.",
        "completeness_scope": "declared_payload",
        "completeness_basis": "the embedded newf-normalize block is the complete declared payload for this corpus file: every representation/assumption/operator/preserves/breaks/auxiliary_object entry the file states is listed here and parsed by construction (deterministic embedded-payload parser). Author-attested 2026-09-12 as part of pvnp-holdout corpus preparation, before freeze and before any capture existed; declared_payload scope, NOT a mechanism-exhaustive extraction judgment.",
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
        "boundary_statement": "the algorithm-to-hardness implication holds only conditionally; the hypothesis it consumes is itself open",
        "boundary_conditions": [
          "hypothesis (derandomized PIT) itself open",
          "no unconditional lower bound extracted as of 2004"
        ],
        "notes": "Converts one open problem into another; the implication's hypothesis is an algorithmic event no one can yet supply."
      },
      "support": [
        {
          "field_path": "mechanism.operators",
          "support_kind": "explicit",
          "locator": "para:2"
        },
        {
          "field_path": "mechanism.breaks",
          "support_kind": "explicit",
          "locator": "para:2"
        },
        {
          "field_path": "outcome.class",
          "support_kind": "explicit",
          "locator": "para:3"
        },
        {
          "field_path": "outcome.boundary_statement",
          "support_kind": "explicit",
          "locator": "para:3"
        }
      ]
    }
  ]
}
newf-normalize -->

---

# Time-space tradeoffs for SAT — indirect diagonalization (pnp-12)

**Era:** 1997–2004. Anchors: Fortnow 1997 (nondeterministic
polynomial time vs small-space); Fortnow–van Melkebeek 2000 (time-space
tradeoffs for satisfiability).

The period's only *unconditional* lower bounds touching SAT itself:
satisfiability cannot be solved simultaneously in time `n^c` and space `n^{o(1)}`
for small constants `c` (e.g. `c < φ` and successive improvements). The method
is **indirect diagonalization**: assume SAT is easy in both time and space,
use completeness to translate the assumption into unlikely inclusions between
nondeterministic and alternating time classes, speed up alternations under
the space bound, and contradict a hierarchy theorem.

Mechanistically this family hybridizes the two flanks: it runs an
**assumption-to-contradiction pipeline through hierarchy theorems** — assume
an algorithm exists, derive class collapses, contradict a hierarchy — and it
is unconditional. But the
assumed algorithm is refuted rather than supplied: the argument consumes a
*hypothetical* algorithm to power a contradiction, and its strength is capped
by how much work the hierarchy theorems can do, yielding bounds that are
model-specific (simultaneous time-space) and quantitatively weak (small
polynomial exponents).

Boundary as of 2004: successive refinements moved the exponent slightly; the
method showed no path to super-polynomial bounds or to separating P from NP.
The bounds remain model-specific and quantitatively weak, capped by how
much leverage the hierarchy theorems provide.

<!-- newf-normalize
{
  "schema_version": "normalize/v1",
  "approaches": [
    {
      "logical_identity": "pvnp/sat-time-space-tradeoffs",
      "label": "SAT time-space tradeoffs (indirect diagonalization)",
      "description": "Unconditional lower bounds against simultaneous time-space algorithms for SAT via assumption-to-contradiction pipelines through hierarchy theorems.",
      "mechanism": {
        "representations": [
          "alternating time classes",
          "simultaneous time-space machines",
          "completeness translations"
        ],
        "assumptions": [
          "hypothetical fast low-space SAT algorithm (for contradiction)"
        ],
        "operators": [
          "indirect diagonalization",
          "alternation speedup",
          "hierarchy contradiction",
          "completeness translation"
        ],
        "preserves": [
          "black-box relativizing simulation"
        ],
        "breaks": [
          "large constructive distinguishing property"
        ],
        "auxiliary_objects": [
          "alternation hierarchy",
          "clocked simulations"
        ],
        "locality": "global",
        "construction_mode": "constructive",
        "uncertainty_mode": "deterministic",
        "notes": "Unconditional skeleton: assumed algorithm -> class collapse -> hierarchy contradiction. Assumed algorithm is refuted, never supplied.",
        "completeness_scope": "declared_payload",
        "completeness_basis": "the embedded newf-normalize block is the complete declared payload for this corpus file: every representation/assumption/operator/preserves/breaks/auxiliary_object entry the file states is listed here and parsed by construction (deterministic embedded-payload parser). Author-attested 2026-09-12 as part of pvnp-holdout corpus preparation, before freeze and before any capture existed; declared_payload scope, NOT a mechanism-exhaustive extraction judgment.",
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
        "boundary_statement": "unconditional but model-specific and quantitatively weak; small-exponent time-space bounds only, no route to superpolynomial separation",
        "boundary_conditions": [
          "strength capped by hierarchy-theorem leverage",
          "bounds tied to the simultaneous time-space model"
        ],
        "notes": "Largely relativizing, hence also wall-bounded; strength capped by hierarchy-theorem leverage."
      },
      "support": [
        {
          "field_path": "mechanism.operators",
          "support_kind": "explicit",
          "locator": "para:1"
        },
        {
          "field_path": "mechanism.notes",
          "support_kind": "explicit",
          "locator": "para:2"
        },
        {
          "field_path": "outcome.class",
          "support_kind": "explicit",
          "locator": "para:1"
        },
        {
          "field_path": "outcome.boundary_statement",
          "support_kind": "explicit",
          "locator": "para:3"
        }
      ]
    }
  ]
}
newf-normalize -->
