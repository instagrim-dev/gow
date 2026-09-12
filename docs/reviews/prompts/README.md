# Research review workbench

**Decision:** retain an ownership map and a small, causal-path prompt set. The
one-domain/one-prompt series is withdrawn. This package records the agreed
closure; it does not claim an integrated test, coverage generator, or comparative
trial has been implemented or passed.

## Start with a decision and a path

| Need | Use |
| --- | --- |
| Validate assessment/admission/current-authority integration | [One-obligation recipe](recipes/assessment-admission-decision.md), first. |
| Inspect concrete identity/view behavior | [Existing identity prompt](assessment-identity-and-derived-views.md). |
| Inspect realization, verification and admission | [Existing projection/admission prompt](typed-projection-and-evidence-admission.md). |
| Compare modular and bespoke packaging | [Preservation protocol](recipes/preservation-pilot.md), only after the integrated slice succeeds. |
| Assign ownership and activate method/agentic obligations | [TAXONOMY.md](TAXONOMY.md); responsibility rows are not separate sessions. |
| Establish evidence, authority, state and reporting rules | [Shared contract](review-contract.md). |

```text
Read AGENTS.md and docs/reviews/prompts/review-contract.md.
Execute docs/reviews/prompts/recipes/assessment-admission-decision.md against
the selected current checkout and a pinned decision policy. Reuse the existing
component tests, trace the actual consumers and keep one evidence bundle.
Do not manufacture coverage, count a declared limitation as a new defect, or
claim that component tests prove the combined scenario. Report execution and
mapping blockers. Stay review-only unless implementation is separately authorized.
```

Legacy invocations of the two existing filenames remain valid. Their concrete
GoW examples are retained, with ownership metadata and agentic overlays. Their
legacy per-prompt verdicts do not aggregate into release eligibility.
Standalone distribution of a composed recipe must include its pinned contract
and required case bundle; an unavailable dependency is a blocker, not assumed text.

## What is delivered, and what is not

Four new Markdown documents: shared contract, taxonomy, integrated recipe and
preservation protocol. Three existing documents updated: this entry point and
the two compatibility prompts. No nineteen-file core, profile directory,
adapter schema, extra report template, runtime types, migrations or new corpus.
No automatic prompt-execution engine is implied by front matter.

The shared contract specifies authorized mandatory obligations, four distinct
record responsibilities, local semantic versioning, classified limitations and
findings, source-inference-report escalation, and reason-coded decision views.
The taxonomy preserves the completeness distinctions, reusable lenses, method
triggers and seam challenges. It leads with paths and gives agentic risks an
immediate home. Its nineteen provisional rows may merge while keeping owners.

`COVERAGE.md` is deliberately absent. Generate it from authoritative records
only after the single slice establishes that projection. Historical review
sources are not an examination denominator; reconstruct scope only with explicit
labels. No hand-authored ledger or status table should impersonate that export.

The preservation protocol is a qualitative twelve-run design with corrected
controls, separate setup/recurring/comparison costs and a break-even horizon.
No model trial was run. A reusable manuscript prompt and further generic
packaging are deferred; the manuscript escalation rule already lives in the
contract. Existing project reviews remain usable during this deferral.

## Acceptance and implementation order

1. Exercise the one obligation through the actual store, admission and decision
   consumers, including withheld, irrelevant-change and historical-replay controls.
2. Derive coverage from the same bundle. Scope the result to that obligation.
3. Run the preservation pilot only with frozen comparable inputs and explicit
   resource authorization. Consider transfer only after preservation passes.

Any missing normative storage mapping is a specific implementation slice, not
permission to turn a requirement into a `CandidateInvariant`. Preserve source
history and research epistemic types throughout. A fixture test is not a proof
of a scientific claim or a demonstration of better search performance.

## Authoring validation and limitations

Prepared on 2026-09-12 against `269de0d335804ceed955925df3ee5d3d02fd608b` through
the GitHub connector. Current review records and component-test sources were
inspected; historical anchors inside the compatibility prompts were not treated
as current findings. The final commit may have a later parent if concurrent
unrelated work advances main; recheck any cited behavior before execution.

Local GitHub checkout failed with `Could not resolve host: github.com`.
Consequently `go build ./...`, `go test ./...`, `gofmt -l .`, and the combined
scenario were not executed for this documentation change. These are explicit
unmet repository validation gates, not passing results. The package does not
alter runtime behavior. Markdown/metadata/link checks and remote file identity
verification are separate document checks, not substitutes for those gates.

Authoring scope cap: four new documents, three compatibility/entry-point edits,
zero production-state changes and zero paid-provider calls. Execution budgets
are declared per path, not multiplied by responsibility count.
