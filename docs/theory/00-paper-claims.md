# 00 — Paper Claims

*Claim registry for the Geometry of Work manuscript. Every major sentence
in the paper must trace to exactly one entry here, classified by role.
This document is the pre-writing constraint: write it before prose, and
do not let the manuscript make a claim this registry does not contain.*

---

## The four claims

### C1 — Conceptual

> A history of problem-solving work can itself be represented as a structured
> search space whose geometry contains useful information about how search is
> failing or progressing.

**Role:** theory

**What it requires of the paper:**
A self-contained definition of Work, StructuralShape, Regime, Boundary,
and the geometry they induce — intelligible without reading `newf` source code
or knowing the Erdős–Straus conjecture. A reader who understands C1 can explain
GoW to a colleague without mentioning software.

**What it does not claim:**
That the geometry is unique, optimal, or complete. That it always contains
useful information. That "geometry" here means a metric space — the term is
used in the informal mathematical sense of "structured space with regions,
boundaries, and trajectories."

---

### C2 — Methodological

> Conserved shapes, outcome regimes, boundaries, and structural deltas provide
> a disciplined way to choose informative next work rather than continuing
> linear search.

**Role:** theory + method

**What it requires of the paper:**
The `Map → Boundary → Move → Project → Measure` algorithm, stated precisely
enough that a reader could execute a small GoW analysis manually with pen and
paper (or with any tool, not necessarily `newf`). The epistemic model that
prevents shape-space fluency from masquerading as domain-space evidence.

**What it does not claim:**
That the method is provably better than alternatives. That it converges.
That the epistemic model is complete. That a single pass of the algorithm
is sufficient — the claim is that the *loop* is informative, not that any
single step is decisive.

---

### C3 — Systems

> `newf` operationalizes this methodology using typed, provenance-carrying
> representations, epistemic state tracking, adversarial challenge, structural
> comparison, and projection into candidate work.

**Role:** implementation

**What it requires of the paper:**
A system architecture section showing how `newf` realizes the GoW loop:
work ingestion → structural normalization → shape-space construction →
shape proposal → challenge lifecycle → boundary/frontier generation →
domain projection → assessment → map update. The technical contribution
framed as: how do you build software that lets fuzzy model-native operations
coexist with deterministic evidence and immutable provenance?

**What it does not claim:**
That `newf` is the only or best possible implementation. That its specific
design choices (SQLite, Go, CLI-first) are load-bearing for the theory.
That the system is production-ready or feature-complete.

---

### C4 — Empirical (deliberately narrow)

> In project-authored Erdős–Straus study corpora, LLMs can propose structural
> regularities not explicitly supplied as reference properties, and
> shape-guided context can produce domain proposals judged closer to a withheld
> structural move; the experiments do **not** establish mathematical validity,
> general discovery effectiveness, or autonomous theorem solving.

**Role:** observation

**What it requires of the paper:**
- Negative controls demonstrating epistemic behavior (the system refuses to
  manufacture structure when preconditions are absent).
- Pilot-003 results: curated-feature guided search. Automatic assessment
  inconclusive; separate model review found B3's proposal matched the intended
  structural move; same-family caveat; criterion recovery not mathematical
  validity.
- Pilot-004 results: shape discovery. 23 proposal occurrences; reference
  matches; provisionally defensible novel occurrences; predeclared aggregate
  criterion not met; capability finding that unencoded shape generation occurred.
- At least one novel-shape challenge lifecycle (N2a or another candidate)
  run end-to-end, whether the hypothesis survives, weakens, splits, or is
  falsified.
- Every number traceable to an immutable frozen experiment artifact.

**What it does not claim:**
Mathematical validity of any generated structure. That GoW outperforms
conventional search statistically — the manuscript's §9 gives only a
*conditional* theorem and the (unrun) preregistered test for such a claim; no
empirical superiority is asserted. That the results generalize to other
domains. That LLM-generated shapes are "discoveries" in the mathematical
sense. That quorum agreement constitutes independent verification.

---

## Sentence-level classification key

Every substantive sentence in the manuscript should carry one of these roles
(implicitly or explicitly traceable to this document):

| Role | Meaning | Evidence standard |
|---|---|---|
| **theory** | Definitional or axiomatic content of GoW | Internal consistency; no empirical claim |
| **method** | Algorithmic or procedural claim about how GoW operates | Operational: a reader could execute it |
| **implementation** | Claim about `newf` specifically | Traceable to source code or design doc |
| **observation** | Empirical claim about what happened in experiments | Traceable to frozen artifact with provenance |
| **interpretation** | Authored reading of observations | Explicitly flagged as interpretation; alternatives noted |
| **hypothesis** | Claim about what GoW might do, untested | Explicitly flagged; falsification path stated |
| **speculation** | Broader implication or future direction | Explicitly in future-work or discussion section |

---

## What goes under hypotheses / future work

The following are interesting and should appear in the paper, but as
explicitly labeled hypotheses or future directions — not as claims:

- **Complement geometry / anti-vacuum.** Work informs search not only through
  the structures attempts instantiate (forward geometry) but through the
  regions of possibility-space they exclude (complement geometry). Sufficiently
  structured negative evidence may induce a residual geometry whose conserved
  properties specify candidate directions not instantiated by prior Work.
  The constraint-collapse operator C(Ω | W) → R, the void-as-structured-
  absence concept, and the field interpretation (failure as pressure, partial
  success as gradient, success as attractor) are formally stated in
  `docs/theory/05-complement-geometry.md` but untested. This is a paper-
  worthy corollary, not a paper-worthy claim, until experimental evidence
  exists. See below for the complementarity principle and its placement.
- **Relational structure (H-R).** GoW can infer outcome-relevant *relations*
  among aspects of prior Work (not just per-feature regularities) and use
  them to select next attempts at lower cost than individual-feature
  analysis and an ordinary interaction-aware baseline. The mathematical
  possibility is established (Proposition R, proved in
  `docs/theory/07-relational-structure.md`: marginal summaries can discard
  all jointly available predictive information); the instrument capability
  is calibrated by the executable XOR/main-effect control in
  `internal/relational`; the *method* claim awaits the pre-registered
  comparative experiment (`corpus/experiments/pilot-005-relational/`).
  Until then the extension is admissible in the conceptual model only as a
  labeled hypothesis.
- **Self-referential programmatic cartography (H-SRC).** A GoW implementation
  can treat its own mapping and search decisions as Work — including its
  representational choices — detect a decision-relevant limitation in its
  current representation, commit to a revised representation and a
  discriminating prediction *before* seeing the outcome, and select next
  domain Work whose checked outcome (against an unchanged objective) supports
  that prediction at a useful cost relative to not revising. The capability is
  architecturally admissible (it is the `M(W)` selection history admitted as
  Work) and its boundary is stated in
  `docs/theory/10-self-referential-cartography.md`: **self-reference, not
  self-certification** — describing itself never establishes the correctness
  of its descriptions. The safeguards are the temporal anti-mythology rule and
  representation-selection contract of
  [08](08-earning-operational-authority.md) plus the challenge discipline
  (`ModelJudgment ≠ Verification`). Until the prospective loop is run against a
  comparator, the extension is admissible only as a labeled hypothesis.
- **Conditional resource-bounded superiority (H-SUP).** For a declared task
  population with executable verification, a fixed GoW implementation achieves a
  higher probability of a verified result within the same total budget than a
  fixed conventional-search comparator. The **conditional theorem is
  established** (theory): manuscript §9 proves the enrichment/overhead criteria
  — `p_G/p = ρ/r` (Prop. 9.1), the exact and approximate budgeted-superiority
  criteria (Prop. 9.2), and the filter-model cost condition (Cor. 9.1) — and the
  finite-sample certificate (Hoeffding lower bound, Prop. 9.3; McNemar exact
  test). These are `theory` claims under stated assumptions and make **no**
  empirical claim. The **empirical premise is a labeled hypothesis**: that a
  fixed GoW implementation attains sufficient enrichment net of overhead on a
  real population. It is *not* established here and is *not* claimed. The
  confirmatory design is preregistered at
  `corpus/experiments/superiority-preregistration/`
  (`status: configured_not_executed`). All numeric examples in §9 are
  explicitly hypothetical calculations, not results. This maps onto
  falsification condition 2 ("shape-space search is dominated").
- GoW applies to domains beyond number theory.
- The failure-invariant instrument is the most productive reading of the geometry.
- Shape-guided search converges faster than undirected search.
- The recursive challenge refinement produces diminishing-radius coverage.
- GoW can operate without LLMs (manually, or with other tools).
- The epistemic model prevents all forms of evidence laundering.
- Micro-failures (within-step geometry) are as productive as macro-failures.
- Human practitioners already do implicit GoW; the framework makes it explicit.
- The search-policy mutation mechanism produces measurable improvement.

Each of these could become a claim in a future paper with the right evidence.
None is supported by the current experimental record.

## Situating GoW against neighboring fields

The manuscript's related-work section is drafted in
[`06-situated-in-the-literature.md`](06-situated-in-the-literature.md). It is
`interpretation` + `hypothesis`, never `observation`: it makes **no** "nobody
thought of this" claim. It maps each GoW operation to its closest established
relative (Lakatos, CEGIS/CEGAR, version spaces / active learning / Bayesian
experimental design, Conceptual Spaces, Instance Space Analysis,
Quality-Diversity, constraint propagation, inverse design), names the one lift
that appears distinct (search over the population of *Work*, with the
representation itself allowed to move), and states the `D ⇆ W ⇆ M(W)` three-space
decomposition. Its four "missing rigorous pieces" — representation learning,
geometry, causality, convergence — are the research program, and each maps onto
a falsification condition below. Any novelty sentence in the manuscript must
survive comparison against Lakatos and Instance Space Analysis first.

---

## Falsification conditions for GoW (Stage 8 material)

The paper must state what future evidence would make GoW uninteresting or wrong:

1. **Geometry is vacuous.** If the structural descriptions assigned to work
   are so model-dependent that different reasonable description procedures
   produce incommensurable geometries, then the "geometry" is a projection
   artifact rather than a property of the work.

2. **Shape-space search is dominated.** If undirected generation (without
   shape-space reasoning) consistently produces equal or better proposals
   measured by structural proximity to withheld advances, then the
   shape-space detour adds cost without information.

3. **Epistemic overhead without epistemic gain.** If the challenge lifecycle
   and provenance machinery produce no measurably different outcomes from
   unstructured "try again with the failures in context," then the
   epistemics are ceremony rather than discipline.

4. **Domain projection is the bottleneck.** If model-generated structural
   deltas cannot be reliably compiled into domain-verifiable candidates
   (i.e., projection failure is the norm rather than the exception), then
   shape-space reasoning is systematically disconnected from domain-space
   truth.

5. **The method does not transfer.** If GoW only produces interesting
   observations on problems whose failure history was curated by the GoW
   practitioner, then it is a curation framework rather than a search
   method.

---

## Provenance

- Theory docs: `docs/theory/00–04`, `glossary.md`
- Pilot-003 frozen record: `corpus/experiments/pilot-003/`
- Pilot-004 frozen record: `corpus/experiments/pilot-004-discovery/`
- Finding 001: `docs/findings/001-unencoded-shape-generation.md`
- N2a challenge: `corpus/experiments/pilot-004-discovery/records/N2a-challenge.md`
