# Typed Toolbox DSL

This document defines a compact research algebra for `newf`: a problem-solving
interface expressed as composable, epistemically typed operations rather than a
sequence of free-form prompts.

The central idea is that the model should operate on the search process itself:
compress failure-space, infer boundaries, re-represent mechanisms, generate
counterfactuals, break candidate invariants, evaluate proposals, and mutate the
future search policy.

The DSL is intentionally small. It should be expressive enough to construct a
research workflow without becoming a taxonomy project that consumes the project
it was meant to help.

## Research algebra

The top-level flow is:

```text
Evidence
  -> Approaches
  -> FailureSpace
  -> CandidateFailureInvariants
  -> ChallengedFailureInvariants
  -> FrontierProposals
  -> EvaluatedResults
  -> SuccessSpace
  -> CandidateSuccessInvariants
  -> SearchPolicy'
```

For a community-resistant problem:

```text
failure-space
  -> failure invariant
  -> invariant break
  -> partial success
  -> success invariant
  -> generalized frontier
```

`solve` is one possible terminal state. It is not the primitive around which the
system is organized.

## Epistemic types

The type system exists primarily to prevent category errors.

```text
Problem
Source
Evidence
EvidenceSet
Approach
ApproachSet
Failure
FailureSpace
Mechanism
MechanismFamily
MechanismFamilies
Boundary
CandidateInvariant
SurvivingInvariant
EstablishedInvariant
Challenge
Counterfactual
Proposal
ProposalSet
Outcome
OutcomeSet
SuccessSpace
SearchPolicy
Run
```

The most important distinction is:

```text
Evidence != Hypothesis
CandidateInvariant != EstablishedInvariant
GeneratedInterpretation != SourceFact
```

A generated claim can become evidence only through an explicit verification or
observation operation that records provenance.

```text
Hypothesis --verify--> Evidence
CandidateInvariant --challenge*--> SurvivingInvariant
SurvivingInvariant --independent-proof--> EstablishedInvariant
```

No operator may silently promote epistemic status.

## Core operators

The initial operator vocabulary is deliberately constrained.

```text
ingest       Source[] -> EvidenceSet
normalize    EvidenceSet -> ApproachSet
cluster      ApproachSet -> MechanismFamilies

compress     Set -> CandidateInvariant[]
contrast     A B -> DifferenceStructure
boundary     SuccessSpace FailureSpace -> Boundary
abstract     X Level -> X'
rerepresent  X Representation -> X'

counterfact  Failure -> Counterfactual[]
transfer     Mechanism Domain -> Proposal[]
break        CandidateInvariant -> Proposal[]

challenge    CandidateInvariant -> Challenge
falsify      Claim -> Evidence
verify       Claim -> Evidence

frontier     Constraints -> ProposalSet
rank         ProposalSet Objective -> ProposalSet
evaluate     Proposal -> Outcome

expand       FailureSpace Axes -> FailureSpace
mutate       SearchPolicy Evidence -> SearchPolicy
```

Operators should be implemented as stable domain actions. Providers may perform
some of the inference behind them, but the provider is not the semantic owner of
the operation.

## Higher-order toolbox

The operators correspond to several useful higher-order LLM capabilities.

### Compress

Find recurring structure across many artifacts.

```text
compress failures
  seek conserved_structure
  across independent_clusters
```

This is not ordinary summarization. The output should identify candidate
properties that survive normalization across superficially different examples.

### Contrast

Find the smallest meaningful structural difference between populations.

```text
contrast failures partial_successes
  by mechanism
```

Contrast is useful when success and failure share most structure but differ in a
small enabling condition.

### Boundary

Infer the frontier separating working and non-working regimes.

```text
boundary successes failures
```

A boundary can be more informative than either population alone.

### Abstract

Move to a different level of explanation.

```text
abstract IF-03 higher
abstract IF-03 lower
```

This is useful when several candidate invariants may collapse into one broader
invariant, or one apparent invariant may split into unrelated mechanisms.

### Re-represent

Express the same research object in a new ontology while preserving provenance.

```text
rerepresent approach-17 as graph
rerepresent failure-space as constraint_system
```

A representation change is useful only if it exposes new structure. Cosmetic
rewriting should score as zero mechanistic novelty.

### Counterfact

Find a minimal change under which a failure might cross into success.

```text
counterfact failure-21
  minimize intervention
```

Conceptually:

```text
Failure -> minimal Delta such that Failure + Delta may enter SuccessSpace
```

### Transfer

Move a structural mechanism between domains while preserving relationships, not
vocabulary.

```text
transfer mechanism-7 from topology to number_theory
```

The output must state the structural mapping and which relationships are claimed
to be preserved.

### Break

Generate mechanisms specifically designed not to preserve a candidate failure
invariant.

```text
break IF-03
  require structural_violation
```

This is the principal bridge from failure compression to frontier generation.

### Challenge / falsify

Attack an inferred invariant rather than reward agreement with it.

```text
challenge IF-03 {
  falsify known_counterexample
  falsify synthetic_counterexample
  contrast successes
  split lower_abstraction
  merge higher_abstraction
  test sampling_bias
}
```

A challenge produces evidence and a status transition. It does not simply return
critic prose.

### Mutate

Modify the future search policy based on accumulated evidence.

```text
policy := mutate policy using [IF-03, IS-02, outcome-41]
```

This is the key higher-order operation: results alter how future candidates are
generated rather than merely becoming more context in another chat turn.

## Pipeline syntax

The DSL should support composition using a pipe form:

```text
evidence
  |> normalize
  |> where outcome in [failure, partial_failure]
  |> cluster by mechanism
  |> compress seek conserved_structure
  |> challenge all
```

The pipe is dataflow, not shell text piping. Every stage has an input type and an
output type.

Invalid composition should fail before provider execution where possible.

```text
# invalid: EvidenceSet cannot be directly challenged

evidence |> challenge
```

## Block syntax

For named, repeatable research programs, use a declarative block:

```text
problem erdos_straus {
  claim:
    forall n >= 2:
      4/n = 1/x + 1/y + 1/z
      where x,y,z in positive_integers

  evidence := ingest(
    literature,
    known_constructions,
    computational_results,
    known_barriers
  )

  approaches := normalize evidence
    by [assumptions, operators, preserves, boundary, outcome]

  failures := approaches
    |> where outcome in [failure, partial_failure]

  families := cluster failures
    by mechanism
    require mechanistic_diversity

  failure_invariants := compress families
    seek conserved_structure
    across independent_clusters

  failure_invariants := challenge failure_invariants {
    falsify known_counterexample
    falsify synthetic_counterexample
    contrast successes
    split lower_abstraction
    merge higher_abstraction
    test sampling_bias
  }

  frontier := generate against surviving(failure_invariants) {
    break invariant
    counterfact minimal_change
    rerepresent orthogonally
    transfer structural_analogy
    maximize information_gain
    minimize redundancy
  }

  results := evaluate frontier {
    proof_check
    symbolic_check
    computation
    counterexample_search
  }

  success_space := results
    |> where outcome in [partial_success, success]

  success_invariants := compress success_space
    seek enabling_structure

  policy := mutate_search {
    avoid failure_invariants
    prefer success_invariants
    expand uncovered_mechanisms
  }

  repeat until [
    solved,
    invariant_established,
    frontier_exhausted,
    no_information_gain
  ]
}
```

The concrete parser may simplify this syntax in v0. The semantic model matters
more than preserving every piece of punctuation shown here.

## Erdős-Straus walkthrough

The conjecture is:

```text
forall n >= 2:
  4/n = 1/x + 1/y + 1/z
  where x,y,z are positive integers
```

A research session should resemble manipulation of a frontier, not a chat about
trying harder.

```text
> normalize corpus
18 approaches, 7 mechanism families

> compress failures --seek conserved_structure
IF-03: residue-local constructions dominate 5/7 failure families
IF-07: finite covering arguments leave structurally related survivor classes

> challenge IF-03 --all
IF-03 weaken
support: 4 independent families
counterexample: family-6
scope revised

> frontier IF-03
F-21 break locality via global auxiliary object
F-22 transfer probabilistic covering mechanism
F-23 counterfact minimal assumption weakening

> evaluate F-21 --cheap-first
partial_success:
  eliminates survivor family S4
  fails on S7

> boundary F-21
progress occurs when global coupling exists,
but only under divisor condition D

> compress successes
IS-02: global coupling + divisor condition D

> mutate-policy IF-03 IS-02

> frontier --next 20
```

The point of this interaction is not the literal example outputs. It is the
shape of the control loop:

```text
observe -> compress -> challenge -> perturb -> test -> learn -> change search
```

## Mechanistic distance

`newf` should distinguish surface diversity from mechanistic diversity.

A proposal that changes notation, prompt wording, or representation without
changing assumptions/operators/preserved structure is not a new frontier sample.

Mechanistic distance may consider components such as:

```text
assumption_distance
operator_distance
preserved_property_distance
representation_distance
locality_distance
construction_mode_distance
verification_mode_distance
```

These need not collapse to a single scalar in v0. Component-wise and ordinal
comparisons are preferable to fake precision.

## Frontier objective

A proposal-ranking objective may be expressed conceptually as:

```text
maximize(
    mechanistic_distance_from_known_failures
  + violation_of_surviving_failure_invariants
  + expected_information_gain
  - evaluation_cost
  - redundancy
)
```

The objective intentionally rewards proposals that are useful even when they
fail.

A low-probability proposal can be valuable if its failure cheaply eliminates a
large region of conceptual search-space.

## Failure as a first-class output

An evaluated proposal does not disappear when it fails.

```text
evaluate P-31
  -> Failure {
       mechanism: ...
       violated_invariants: ...
       preserved_invariants: ...
       boundary: ...
       evidence: ...
       information_gain: ...
     }
```

The failure returns to the atlas and can modify later compression.

```text
FailureSpace(t+1) = FailureSpace(t) + newly_evaluated_failures
```

This is how synthetic frontier generation expands the higher-order sample space.

## Invariant lifecycle

Candidate invariants should have explicit state.

```text
proposed
  -> challenged
  -> surviving
  -> weaken
  -> split
  -> merge
  -> falsified
  -> established
```

Not every transition is linear. `split` and `merge` produce lineage into new
candidate invariants.

`established` requires independent evidence appropriate to the domain. Model
agreement cannot establish an invariant.

## Search-policy state

The search policy should be inspectable data.

```yaml
policy:
  prefer:
    - global_coupling
    - auxiliary_object
  avoid:
    - residue_locality_only
  expand:
    - probabilistic_methods
    - geometric_methods
  penalties:
    redundancy: high
    cosmetic_rerepresentation: high
  budget:
    cheap_falsification_first: true
```

A model/provider may propose policy mutations, but the resulting policy must be
persisted as a revision with provenance.

## Terminal conditions

A workflow may stop for several reasons:

```text
solved
invariant_established
frontier_exhausted
no_information_gain
budget_exhausted
insufficient_failure_diversity
verification_blocked
```

`no_information_gain` and `insufficient_failure_diversity` are legitimate
research outcomes. They indicate that more frontier samples or new mechanism
families are needed rather than inviting an infinite loop of regenerated prose.

## CLI mapping

The initial CLI can expose the algebra without requiring users to author DSL
files immediately.

```text
newf ingest ...
newf normalize
newf cluster
newf invariants
newf challenge IF-03
newf frontier --against IF-03
newf evaluate P-21
newf boundary --success partial_success --failure failure
newf compress --successes
newf policy mutate
```

Later, a `newf apply <program.newf>` command can execute a declarative research
program once the operator semantics and persistence contracts have stabilized.

## Design constraints

1. Operators are typed domain actions, not aliases for prompts.
2. Provider calls return structured artifacts with provenance.
3. Generated interpretation never becomes source evidence implicitly.
4. Failure remains a durable output when it changes the known search-space.
5. Search policy is explicit, versioned, and inspectable.
6. Mechanistic novelty is preferred over surface novelty.
7. Cheap falsification is preferred before expensive elaboration.
8. Every higher-order operation must be independently evaluable.
9. The grammar should remain small enough to reason about compositionally.
10. Solving the original problem is an outcome of the system, not the only useful
    measure of progress.

## v0 grammar boundary

Do not implement a full parser first.

The first implementation should establish the typed domain operations and CLI
commands. A textual DSL should be added only after the operations have stable
input/output contracts.

Otherwise `newf` risks building a beautiful language for semantics it has not yet
managed to define, a surprisingly popular software-development genre.