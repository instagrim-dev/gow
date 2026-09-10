# Abstraction Safety

`newf` treats abstraction as a first-class research operator because it is one of the strongest ways an LLM can expose hidden structure across heterogeneous approaches. It is also one of the easiest ways to produce elegant nonsense.

The central failure mode is **abstraction drift**: a representation or explanatory-level change removes distinctions that matter to the original claim, yet downstream reasoning proceeds as if nothing was lost.

## Core rule

An abstraction is useful only when it improves compression or transfer **without destroying predictive discrimination over the original problem**.

Conceptually:

```text
X --abstract(phi)--> X'
```

For a property or prediction that matters to the research claim, `newf` should be able to state whether:

```text
P(X) <-> P'(phi(X))
```

actually holds, only holds in one direction, holds under explicit conditions, or is unknown.

Abstraction must never silently upgrade one of those weaker relationships into equivalence.

## Required abstraction record

Every non-trivial abstraction should emit a structured artifact containing at least:

```yaml
abstraction:
  source_id: ...
  target_representation: ...
  mapping: ...
  preserves:
    - ...
  loses:
    - ...
  introduces:
    - ...
  assumptions:
    - ...
  claimed_equivalences:
    - ...
  one_way_implications:
    - ...
  unresolved_correspondence:
    - ...
  grounding_plan: ...
  verification_status: untested
```

`preserves`, `loses`, and `introduces` are mandatory concepts even if one list is empty. Unknown information should be recorded as unknown rather than omitted.

## Main failure modes

### False equivalence

Two mechanisms can look identical only because the abstraction erased the distinction that determines success or failure.

Compression is not evidence of equivalence.

### Invariant destruction

An abstraction can erase the very invariant `newf` is trying to discover. If algebraic, geometric, and computational failures are all reduced to something as broad as `constraint reduction`, the representation may compress beautifully while discriminating nothing useful.

### Ontology leakage

Cross-domain transfer can import hidden assumptions from the source domain. Calling an optimization transition a `phase transition`, for example, does not grant it the mathematical properties of a physical phase transition.

Mappings must identify which relations are transferred and which are merely analogous.

### Unverifiable altitude

Higher-level claims become easier to phrase and harder to falsify. Terms such as `local`, `global`, `structural`, `coupled`, or `emergent` are not useful invariants until operationalized against concrete cases.

### Semantic hallucination

LLMs can generate coherent abstractions more easily than they can establish that those abstractions preserve the truth conditions of the source problem. Fluency is not grounding.

## Bidirectional abstraction

`abstract` should be paired with `ground`.

```text
abstract   Concrete -> Abstraction
ground     Abstraction -> ConcretePrediction[]
```

The minimum research loop is:

```text
source cases
  -> abstract
  -> derive prediction / classification
  -> ground into concrete cases
  -> test against source domain
  -> retain | weaken | split | falsify abstraction
```

Conceptually:

```text
X -> X' -> X_hat
```

and the system should retain evidence about the discrepancy between `X` and the grounded reconstruction `X_hat`.

The exact distance function is domain-specific. The important point is that the round trip produces testable consequences.

## Predictive discrimination criterion

An abstraction earns search authority only if it preserves or improves the ability to distinguish outcomes that matter.

For failure-space work, useful tests include:

- can the abstraction still distinguish known failures from known partial successes?
- can it predict held-out failure boundaries?
- does it merge mechanism families that later require different search policies?
- do frontier proposals generated from the abstraction retain the claimed structural violation when grounded?
- does a cross-domain transfer survive concrete counterexamples in the target domain?

A simpler description is:

```text
better compression + preserved prediction = useful abstraction
better compression + worse prediction     = abstraction drift
```

## Abstraction lineage

Abstractions are derived artifacts and require provenance.

A higher-level abstraction must link to:

- source artifacts;
- operator/provider invocation;
- abstraction level or target representation;
- explicit preservation/loss claims;
- grounding attempts;
- counterexamples;
- revisions, splits, merges, or falsification.

Do not overwrite an abstraction when grounding reveals a defect. Preserve the original and create a weakened or revised descendant.

## Interaction with invariants

Candidate failure invariants inferred at one abstraction level should be challenged above and below that level.

```text
I(level=n)
  -> abstract higher
  -> ground lower
  -> compare support / counterexamples
```

This prevents two common errors:

1. a low-level invariant that is only an artifact of representation;
2. a high-level invariant so broad that it no longer predicts the success/failure boundary.

An invariant that survives controlled abstraction changes is stronger evidence than one that exists only in one vocabulary.

## Interaction with transfer

`transfer` across disciplines must include an abstraction map and a grounding test in the target domain.

```text
source mechanism
  -> abstract structural relation
  -> transfer relation
  -> ground target-domain prediction
  -> test
```

Do not transfer labels when the useful object is a relation.

The transfer artifact should distinguish:

- structural isomorphism or claimed correspondence;
- analogy only;
- imported assumptions;
- target-domain counterexamples;
- verification strength.

## Suggested DSL shape

```text
A := abstract failure-space higher {
  preserve [outcome_boundary, operator_relations]
  report_loss
  report_introduced_assumptions
}

P := ground A into original_domain {
  predict outcomes
  reconstruct representative cases
}

test P against held_out_cases

accept A only_if predictive_discrimination >= baseline
```

The concrete parser does not need to support this syntax in v0. These are semantic requirements for the eventual typed operators.

## Stopping / rejection conditions

Reject or weaken an abstraction when:

```text
cannot_state_preservation
cannot_generate_grounded_prediction
merges_known_distinct_outcomes
requires_unrecorded_assumptions
fails_round_trip
reduces_predictive_discrimination
cannot_be_operationalized
```

`cannot_operationalize` is a valid outcome. The response is not to generate a more impressive noun.

## Design consequence

The higher-order toolbox should therefore contain both:

```text
abstract  X Level -> Abstraction
ground    Abstraction Domain -> ConcretePrediction[]
```

and treat the pair as one epistemic control surface.

`abstract` generates leverage.

`ground` determines whether that leverage still points at the original problem.