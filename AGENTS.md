# AGENTS.md

This file defines how coding agents should work in `newf`.

`newf` is not a generic chatbot wrapper. It is a research-search system that treats failures as a first-class sample space, compresses them into candidate invariants, challenges those invariants, and generates frontier proposals that deliberately violate surviving failure structure.

The governing research loop is:

```text
failure-space
  -> failure invariant
  -> invariant break
  -> partial success
  -> success invariant
  -> generalized frontier
```

The repository owns the workflow. Models execute bounded inference operations inside it.

## Division of responsibility

### `newf` owns

- durable state;
- epistemic types;
- provenance and lineage;
- legal state transitions;
- search-policy state;
- execution budgets;
- stopping conditions;
- persistence and reproducibility;
- machine-readable contracts.

### Model providers own bounded fuzzy operations

Typical model-native operators include:

- `compress`
- `contrast`
- `abstract`
- `rerepresent`
- `counterfact`
- `transfer`
- `break`
- `challenge`
- candidate ranking where judgment is inherently fuzzy

A provider may implement several roles, but provider identity must not leak into domain semantics.

### Deterministic or independently checkable tools own truth-sensitive operations

Prefer external or deterministic checks for:

- arithmetic and symbolic computation;
- code execution;
- proof checking;
- theorem/proof assistant validation;
- database constraints;
- schema validation;
- source retrieval;
- reproducible experiments;
- counterexample search;
- final verification of claims when a stronger mechanism is available.

A model may propose a verification procedure. It should not silently certify its own output merely because it sounds convincing.

## Core rule: decompose work into native capabilities

Do not give a model an opaque instruction such as:

```text
solve the conjecture
```

Prefer explicit typed operations:

```text
ingest
-> normalize
-> cluster
-> compress
-> challenge
-> frontier
-> evaluate
-> boundary
-> compress successes
-> mutate policy
```

Each stage should consume persisted artifacts and produce persisted artifacts with provenance.

The model should not be required to simultaneously act as:

- researcher;
- workflow engine;
- scheduler;
- database;
- critic;
- judge;
- verifier;
- convergence controller.

That is application structure, not intelligence.

## Epistemic invariants

These distinctions are mandatory:

```text
Evidence != Hypothesis
SourceFact != GeneratedInterpretation
CandidateInvariant != EstablishedInvariant
ModelJudgment != Verification
Failure != UselessOutput
```

No operation may silently promote epistemic status.

Examples:

```text
Hypothesis --verify--> Evidence
CandidateInvariant --challenge*--> SurvivingInvariant
SurvivingInvariant --independent-proof--> EstablishedInvariant
```

If independent verification is unavailable, preserve the weaker type and record the limitation.

## Failure is a first-class artifact

A failed proposal is valuable when it narrows the search space, falsifies an invariant, identifies a boundary, or reveals redundancy.

Do not discard informative failures merely because the attempted result was negative.

Persist enough structure to answer:

- what mechanism was attempted;
- what assumptions it used;
- what properties it preserved;
- what candidate invariants it targeted;
- where it failed;
- how the failure was verified;
- what information was gained;
- how it should affect later search policy.

## Mechanistic novelty over surface novelty

Do not treat a new prompt, notation, analogy, or representation as a new research direction unless the underlying mechanism changes.

When comparing approaches, prefer dimensions such as:

- assumptions;
- operators;
- preserved properties;
- locality/globality;
- constructive/existential mode;
- deterministic/probabilistic mode;
- auxiliary-object use;
- search bounds;
- evaluation method.

A cosmetic rewrite should have near-zero frontier value.

## Candidate invariant discipline

An inferred invariant is a hypothesis about conserved failure structure, not a discovered law.

Every candidate invariant should be challenged by attempting to:

1. find a known failed approach that violates it;
2. construct a synthetic failed approach that violates it;
3. find a success that still preserves it;
4. lower the abstraction level and see whether it splits;
5. raise the abstraction level and see whether several invariants merge;
6. identify sampling or publication bias;
7. distinguish correlation from causal obstruction.

Only independently established evidence may promote an invariant to an established state.

## Frontier generation discipline

Do not ask for another arbitrary solution attempt.

Generate proposals against surviving failure invariants.

Every proposal should state:

- target invariant(s);
- claimed structural violation;
- nearest known mechanism family;
- why it is actually distinct;
- cheapest falsification path;
- expected information gain if it fails;
- estimated evaluation cost;
- provenance for all claims used to justify it.

Conceptually, prefer proposals that improve:

```text
mechanistic_distance
+ invariant_violation
+ expected_information_gain
- evaluation_cost
- redundancy
```

Do not fabricate fake precision. Component-wise or ordinal judgments are acceptable in v0.

## Search-policy mutation

Results should change future search behavior.

The search policy must be explicit, versioned, persisted, and inspectable.

Examples of policy changes:

- penalize mechanisms repeatedly shown redundant;
- avoid preserving a surviving failure invariant;
- prefer structures associated with partial success;
- expand under-sampled mechanism families;
- increase cheap-falsification pressure;
- lower or raise abstraction where candidate invariants are unstable.

Do not rely on "the model will remember this from context" as state management.

## Work execution

When implementing an issue:

1. Read `README.md` and relevant material in `docs/` before changing architecture.
2. Read the full issue and identify the smallest coherent vertical slice.
3. Preserve existing names and contracts unless the issue explicitly requires a design change.
4. Prefer one working path end-to-end over broad scaffolding with no usable behavior.
5. Add tests at the boundary where behavior is introduced.
6. Keep machine-readable output stable once exposed.
7. Persist long-running or research-significant intermediate state instead of leaving it only in terminal output.
8. Do not add provider coupling to domain packages.
9. Do not add orchestration abstractions before the underlying typed operations exist.
10. Update docs only where behavior or contracts changed.

## Go implementation bias

Unless an accepted design says otherwise:

- keep domain types free of Cobra and SQL concerns;
- keep CLI wiring thin;
- put SQLite behind storage/repository boundaries;
- use explicit migrations;
- prefer transactions for multi-record provenance writes;
- make invalid provenance transitions difficult by construction;
- prefer typed structs over `map[string]any` for core domain records;
- use JSON only where the shape is genuinely open-ended;
- keep interfaces small and consumer-owned;
- avoid one-package-per-noun architecture;
- add abstractions after repeated pressure, not preemptively.

## Provider implementation bias

Provider adapters should:

- expose role-specific structured operations;
- return structured outputs where practical;
- record model/provider/version/configuration metadata;
- preserve prompt/input/output provenance needed for replay or audit;
- distinguish provider failure from semantic evaluation failure;
- avoid embedding vendor-specific concepts into domain records;
- support deterministic fixtures for tests.

The same underlying model can satisfy multiple roles, but each invocation must record the role it performed.

## Verification hierarchy

Prefer the strongest available verifier:

```text
formal proof / deterministic check
> reproducible computation or experiment
> independently sourced evidence
> cross-model or independent critic agreement
> single-model judgment
```

Higher layers do not automatically invalidate lower ones, but the stored evidence strength must remain explicit.

Consensus among models is still model judgment.

## Evaluation before open-problem theater

The primary early benchmark is historical holdout prediction.

Given only pre-cutoff failures and partial results, test whether the system can:

- infer a useful failure invariant;
- survive adversarial challenge;
- generate proposals that recover or approximate the structural move of a held-out later advance;
- beat undirected solution generation and ordinary semantic summarization.

Do not claim success because generated research prose looks sophisticated.

## Stopping conditions

A run may legitimately stop with:

```text
solved
invariant_established
frontier_exhausted
no_information_gain
budget_exhausted
insufficient_failure_diversity
verification_blocked
```

`no_information_gain` is not permission to regenerate the same mechanism with different adjectives.

## Current repository direction

Start with the CLI and durable research substrate before building a textual DSL parser or autonomous orchestration layer.

The intended order is roughly:

```text
problem/workspace persistence
-> source/evidence ingestion
-> approach normalization
-> mechanism clustering
-> invariant mining
-> invariant challenge
-> frontier generation
-> evaluation
-> success compression
-> search-policy mutation
-> historical holdout experiment
```

The typed toolbox described in `docs/toolbox-dsl.md` is the semantic target. CLI commands should expose those semantics incrementally.

## Abstraction safety

Abstraction and re-representation are high-leverage model-native operations, but they must not silently change the problem being solved.

The governing rule is:

> An abstraction earns search authority only when it improves compression or transfer while preserving predictive discrimination over the source problem.

Every non-trivial abstraction must record:

- the source artifact(s);
- the mapping into the new representation or explanatory level;
- properties claimed to be preserved;
- information known to be lost;
- assumptions or structure introduced by the abstraction;
- whether claimed correspondence is equivalence, one-way implication, analogy, or unknown;
- a grounding plan back into concrete source-domain cases;
- verification status and counterexamples.

Do not allow `abstract` to be a one-way escape into nicer prose. Pair it with `ground`:

```text
abstract   Concrete -> Abstraction
ground     Abstraction -> ConcretePrediction[]
```

The required validation loop is:

```text
source cases
-> abstract
-> derive prediction
-> ground
-> test against original or held-out cases
-> retain | weaken | split | falsify
```

Reject or weaken abstractions that:

- cannot state what they preserve;
- erase distinctions that separate known successes from failures;
- import unrecorded assumptions from another ontology or discipline;
- cannot produce concrete predictions;
- fail round-trip grounding;
- reduce predictive discrimination;
- remain too vague to operationalize.

Compression alone is not success. If two mechanism families become indistinguishable only because the abstraction erased the variable that determines outcome, the abstraction is defective.

Cross-domain `transfer` must likewise operate on structural relations rather than labels, record imported assumptions, and ground the transferred relation into target-domain predictions before it can affect search policy.

Candidate invariants should be challenged above and below the abstraction level where they were inferred. An invariant that survives controlled abstraction changes is stronger than one that exists only in one vocabulary.

See `docs/abstraction-safety.md` for the full contract.

## Definition of good agent work

A good contribution makes the research state more explicit, typed, testable, reproducible, and falsifiable.

A bad contribution hides uncertainty behind prose, adds orchestration without observable state, lets generated claims masquerade as evidence, or creates ten abstractions where one transaction and a struct would have sufficed.