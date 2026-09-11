# Related Work Matrix

*Structured comparison of GoW against its intellectual neighbors. Each row
is a research tradition; columns are the features that collectively
distinguish GoW. The paper's novelty claim will be stated as the conjunction
of columns no single row fills.*

---

## Column definitions

| Column | Question |
|---|---|
| **representation_of_attempts** | Does the work explicitly represent prior attempts as structured objects? |
| **uses_failures** | Does the work use failure information? |
| **uses_successes** | Does the work use success information? |
| **outcome_conditioned_regimes** | Does the work model regions of the attempt space conditioned on outcome? |
| **infers_conserved_structure** | Does the work infer properties conserved across a population of attempts? |
| **boundary_concept** | Does the work have a concept of structural boundary between outcome regimes? |
| **generates_next_work** | Does the work use the analysis to generate the next attempt? |
| **projects_to_domain** | Is the generated work verified in the original domain (not just in the meta-space)? |
| **updates_from_landing** | Does the result of verification feed back into the analysis? |
| **epistemic_provenance** | Does the work track epistemic status and provenance of claims? |

---

## Matrix

| Tradition | repr | fail | succ | regimes | conserved | boundary | gen | proj | update | epist | Key difference from GoW |
|---|---|---|---|---|---|---|---|---|---|---|---|
| **Metaheuristics / search-space methods** (SA, GA, etc.) | solution encodings | implicit (fitness) | yes | no | no (fitness landscape) | no explicit | yes | yes (solutions are domain objects) | yes (fitness) | no | Operates on solution space, not attempt-history space; no structural description of *why* attempts fail |
| **Novelty search / quality-diversity** (Lehman & Stanley; MAP-Elites) | behavior descriptors | yes (archive) | yes | behavior niches | yes (niche boundaries) | implicit (niche edges) | yes | yes | yes (archive update) | no | Searches *behavior* space, not *structural-description* space; diversity is behavioral, not mechanistic; no epistemic lifecycle |
| **Case-based reasoning / analogical reasoning** | case libraries | yes (failed cases retrievable) | yes | no explicit regimes | similarity, not conserved structure | no | yes (adapted solutions) | partial (case retention) | no | Retrieves *similar* cases, not *structurally characterized populations*; no regime/boundary/delta vocabulary |
| **Automated scientific discovery** (Bacon, Eureka, AI Scientist) | data/observations | partial | partial | no | yes (laws/patterns) | no | yes (hypotheses) | yes (experiment loop) | partial | Discovers patterns in *data*, not in *prior work*; no failure-population geometry; discovery is the goal, not search guidance |
| **Automated theorem proving** (Lean, Isabelle, CoqHammer) | proof states | yes (failed tactics) | yes (successful proofs) | no | no (tactic libraries) | no | yes (tactic selection) | yes (learning from proofs) | formal (proof objects) | Searches *proof space*, not *attempt-history space*; failures are dead ends, not structured measurements |
| **Conjecture generation** (Graffiti, TxGraffiti, AI conjecturing) | graph/object properties | partial | partial | no | yes (conjectured invariants) | no | yes (conjectures) | partial | no | Generates conjectures from *object properties*, not from *failure populations*; no outcome-conditioned regimes |
| **Program synthesis / CEGAR** | program space / abstraction–refinement | yes (counterexamples) | yes (correct programs) | yes (abstraction partitions) | yes (loop invariants) | yes (counterexample-guided refinement) | yes | yes (concrete programs) | yes (refinement loop) | partial (abstraction history) | **Closest neighbor.** Refines an abstraction using counterexamples. Difference: CEGAR refines *one program's correctness proof*, not a *population of heterogeneous attempts*; the "geometry" is over a single synthesis target, not a corpus of prior work |
| **Meta-learning / learning-to-search** (MAML, learning curve extrapolation) | task representations | yes (poor-performing configs) | yes (good configs) | implicit (task similarity) | yes (transferable knowledge) | implicit | yes (adapted strategies) | yes | no | Learns *across tasks*, not *across attempts on one task*; the meta-space is task-indexed, not attempt-indexed |
| **Failure analysis / root-cause analysis** | failure reports | yes (primary data) | partial (baseline) | partial (failure modes) | yes (failure patterns) | partial (failure mode boundaries) | no (diagnosis, not generation) | no | no | Analyzes failures but does not *generate next work*; diagnostic, not generative; no shape-space search |
| **Design-space exploration** (DSE, Pareto optimization) | design parameters | yes (infeasible designs) | yes (Pareto-optimal) | yes (feasibility regions) | partial (constraints) | yes (feasibility boundaries) | yes | yes (concrete designs) | yes | no | Operates on *parameter space* with known dimensions, not on *structurally described attempt history*; the space is given, not inferred |
| **LLM self-reflection / debate / verifier** (Reflexion, debate, Process RM) | generation traces | yes (self-critique) | yes | no | no | no | yes (revised generation) | partial (may re-verify) | no | Reflects on *its own single generation*, not on a *population of structurally described prior work*; no persistent geometry across attempts |

---

## The plausible novelty claim

No single row above fills all columns. The nearest is **CEGAR/program synthesis**,
which has counterexample-guided refinement, abstraction partitions, and a
feedback loop — but operates on a single synthesis target rather than a
heterogeneous population of prior work, and does not track epistemic provenance
of structural claims.

The narrow, defensible novelty claim:

> **GoW treats a heterogeneous historical population of completed Work as a
> first-class geometric object, separates structural shape from outcome and
> epistemic status, and searches by inferring and traversing regime boundaries
> before projecting a move back into the domain.**

The *conjunction* is the novelty, not any single ingredient. Specifically:

1. The **object of analysis** is a population of prior work (not a single program,
   not raw data, not a behavior descriptor).
2. The **representation** is structural shape independent of outcome (not fitness,
   not behavior, not similarity).
3. The **search operates on inferred geometry** over that population (regimes,
   boundaries, conserved structure) rather than on the domain directly.
4. **Verification returns to the domain** — shape-space reasoning generates
   candidates; domain-space verification decides them.
5. **Epistemic status is tracked** — claims about the geometry carry provenance,
   conditioning, and explicit promotion rules.
6. **A dual complement reading is available** (hypothesis-status) — inferring
   residual structural requirements from the possibility-space that accumulated
   work excludes, not only from the structure work instantiates. No tradition
   above formalizes this dual as a first-class operator.

---

## Research tasks (papers to read and cite)

### High priority (must appear in related-work section)

- [ ] Lehman & Stanley (2011) — Abandoning objectives: novelty search
- [ ] Mouret & Clune (2015) — MAP-Elites / quality-diversity
- [ ] Langley et al. (1987) — BACON and scientific discovery
- [ ] Lu et al. (2024) — The AI Scientist
- [ ] Clarke et al. (2003) — Counterexample-Guided Abstraction Refinement (CEGAR)
- [ ] Solar-Lezama (2008) — Program synthesis (SKETCH)
- [ ] Fajtlowicz (1988) — Graffiti conjecturer
- [ ] Kolodner (1993) — Case-based reasoning
- [ ] Gentner (1983) — Structure-mapping theory (analogical reasoning)
- [ ] Shinn et al. (2023) — Reflexion (LLM self-reflection)
- [ ] Finn et al. (2017) — MAML (meta-learning)
- [ ] Romera-Paredes et al. (2024) — FunSearch (LLM + evolutionary search for math)
- [ ] Trinh et al. (2024) — AlphaGeometry
- [ ] de Moura & Bjørner (2008) — Z3 / SMT solving
- [ ] Bansal et al. (2019) — HOList (learning to prove theorems)

### Medium priority (should appear if space permits)

- [ ] Derner & Batistič (2023) — LLM-based theorem proving survey
- [ ] Li et al. (2024) — LeanDojo
- [ ] Polu & Sutskever (2020) — Generative language modeling for theorem proving
- [ ] Wagstaff et al. (2023) — Design-space exploration survey
- [ ] Falkner et al. (2018) — BOHB (Bayesian optimization + bandit)
- [ ] Chipman (1993) / Ortiz et al. (2023) — Failure analysis surveys
- [ ] Du et al. (2023) — LLM debate
- [ ] Irving et al. (2018) — AI safety via debate
- [ ] Cobbe et al. (2021) — Verifiers for math (process reward models)
- [ ] Lightman et al. (2023) — Process reward models
- [ ] Norvig (1992) — Paradigms of AI Programming (search-space methods)
- [ ] Mitchell (1993) — Analogy-Making as Perception (Copycat)

### Lower priority (mention if relevant)

- [ ] Lenat (1983) — AM / Eureka
- [ ] Colton (2002) — HR (automated theory formation)
- [ ] Lakatos (1976) — Proofs and Refutations (philosophical ancestor)
- [ ] Polya (1945) — How to Solve It
- [ ] Wang et al. (2024) — MathChat / ChatDev for math
- [ ] Jiang et al. (2022) — Draft, Sketch, and Prove (autoformalization)
- [ ] Pei et al. (2023) — Can LLMs generate novel research ideas?

---

## Notes

- The matrix will be refined as papers are read. Entries marked with
  qualifiers ("partial", "implicit") need source-backed verification.
- The novelty claim should be tested by asking: "Could someone do exactly
  this with tool X?" If yes, that tool fills the row and the claim needs
  narrowing.
- CEGAR is the most important comparison to get right because it is the
  closest structural analog. The paper should dedicate a paragraph to the
  precise distinction.
