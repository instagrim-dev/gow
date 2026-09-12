# Structural / Semantic / Epistemic Review — 2026-09-12

**Reviewed revision:** `3443a10` (CI green at that commit).
**Review method:** source-level review plus isolated SQLite checks. Explicitly
NOT an independent execution of the Go test suite and NOT a certification of
source mathematics.
**Reviewer:** the project's adversarial ChatGPT review process (per issue #16's
Purpose section — not independent outside-peer review).
**Received:** 2026-09-12, via operator paste into the working session.
**Remediation:** tracked in this document's Findings table; code remediation
landed in the commits referenced there.

## Remediation status

| Finding | Status | Where |
|---|---|---|
| S1 challenge assessment population | **Remediated** | migration v34 + `challenge_assessment_populations`; `newf challenge --population latest\|discovery` (default `latest`); out-of-scope counterexamples weaken rather than retroactively falsify; regression `TestIntegrationChallengeAssessmentPopulation` |
| S2 evidence admission / atlas materialization | **Remediated** | migration v35 + `evidence_admissions` ledger; `newf evidence admit` with typed observation kinds (structural-claim-failure never admissible; domain-checked rule-admitted; model-judged requires operator attestation); exact assessed content materialized as a revision of `frontier-proposal:<id>`; regression `TestIntegrationEvidenceAdmissionClosesReentry` |
| S3 first-verdict artifact result | **Remediated (PR #18, `5dd0861`)** | derived assessment views scope current results to generation/content/problem/cluster; `result` re-documented as immutable initial verdict |
| S4 historical interpretation revisions in default population | **Remediated (PR #18, `5dd0861`)** | `ListSignaturesForProblem` selects current approach heads; `ListHistoricalSignaturesForProblem` retains explicit all-history access |
| S5 typed boundary_delta / concrete projection | **Remediated** | v36: `challenge_boundary_deltas` — every confirmed challenge derives a typed `BoundaryDelta` in code (counterexample-separation, contrast-collapse, constructibility, support-recount, split-partition, merge-union) with the canonical separating condition, persisted and surfaced as `challenges[].boundary_delta`; v37: four-record projection chain (`frontier_proposals` → `projection_artifacts` → `projection_obligations`/decisions → referenced evaluation), `newf projection propose/discharge/list`, deterministic steps-compose checker refutes non-composing plans before domain work with typed gaps; regressions `TestIntegrationProjectionChainFourRecords`, `TestIntegrationProjectionCompositionFailure`; see `docs/projection.md` |
| Semantic: authored claim form (quantifier/scope) | **Remediated** | v39: `invariant_claim_forms` (append-only, immutable, latest governs) — `newf invariant claim --quantifier universal\|recurrent\|existential --role --scope --note`; `associationKindForCandidate` grants universal-counterexample treatment ONLY from an authored `universal` quantifier, never from measured coverage; regression: the S1 population test now demonstrates a lone counterexample is inconclusive until the operator authors the universal claim |
| Semantic: verification subject axis | **Remediated** | v38: `verification_subject ∈ annotation\|realization\|domain-goal` on evaluations and challenge evidence; `verify.Verifier.Subject()` is part of the registration and `Route` stamps it (deterministic tiers certify the ANNOTATION; the model tier judges the domain goal); NULL only for pre-v38 history — never backfilled retroactively |
| Epistemic: N2a supersession surfacing, prospective comparison | **Open** — operator workstream | current-claim views with supersession links; two-observation closed loop |

Known limitation recorded with S1: a `discovery`-policy replay of a weakened
invariant can re-earn `surviving` against the old population. The population
rows keep this auditable; ranking survival by population recency is future
state-machine work.

## Verdict (reviewer's words)

> GoW is a credible research workbench, but its core implementation does not
> yet close the learning loop described by the theory, and its experiments do
> not establish superior discovery performance. The strongest part is the
> distinction between proposed structure and warranted knowledge; the weakest
> part is carrying that distinction consistently through assessment,
> projection, and subsequent action.

- **Structural:** substantial implementation, with five consequential
  integration gaps.
- **Semantic:** coherent central model; several conceptual-to-code mappings
  promise more than the corresponding types enforce.
- **Epistemic:** meaningful safeguards and transparent corrections;
  effectiveness remains unestablished.

## Findings

### S1 (High) — Re-challenge cannot assess a newly selected evidence population

`challengeOne` resolves its evidence through
`candidate → original invariant revision → original failure space → original
cluster run → original challenge population`. Newly ingested and reclustered
evidence does not enter the known-counterexample check when an existing
invariant is re-challenged, while frontier generation uses the latest cluster
run. The map can advance while an invariant's search authority stays justified
against an older population.

**Required change:** distinguish the *discovery population* from the
*assessment population*. A challenge must identify both its claim scope and
the evidence manifest it ran against. Regression: derive a claim under
population A, add contradictory evidence B, assess under A+B; preserve the
historical statement about A; update generalization/search eligibility, not
history.

### S2 (High) — Failure re-entry is recorded but not consumed

`PersistEvaluationRun` writes `evaluated_failures` markers, but
`BuildClustering` obtains its population via `ListSignaturesForProblem`
(`mechanism_signatures → mechanisms → approach_revisions → approaches`),
which never reads the markers. A recorded failure is not yet a new observation
for subsequent clustering/mining.

**Required change:** explicit evidence admission and atlas materialization.
Do NOT indiscriminately admit every model-evaluated failure: a description
failing its own structural claim, an unrealizable projection, and a
domain-checked failed attempt are different observations with different
admission rules.

### S3 (High) — Artifact result retains the first verdict, not the latest

`UPDATE frontier_proposals SET result = ? WHERE id = ? AND result IS NULL` is
first-write-wins. Success-then-failure retains "success"; failure-then-success
retains "failure". The source comment says the field mirrors the latest
verdict; the migration permits only one-time assignment. Reviewer reproduced
both sequences in isolated SQLite.

**Required change:** derive current results from the evaluation ledger under
the relevant content and assessment context. An explicitly named
`initial_result` is defensible; an unqualified `result` that reads as current
is not. Globally-latest is also insufficient when content/targets/policy
changed.

### S4 (Medium) — Default atlas population includes historical interpretation revisions

`ListSignaturesForProblem` filters by problem/schema/vocabulary but not by
current approach revision; multiple interpretations of the same work item are
simultaneously eligible. Corrected interpretations can inflate apparent
support when the correction changes the mechanism family.

**Required change:** explicit population modes — current interpretations,
historical replay, or all-history — with source lineage part of population
selection.

### S5 (High for the systems claim C3) — Boundary refinement and projection are partly prose contracts

The glossary maps `boundary_delta` to `ChallengeResult.boundary_delta`; the Go
struct has `Confirmed`/`Outcome`/`Evidence`/`Detail` and no such field.
`Projection` maps to a candidate signature plus directed-generation prose. A
refinement described in a transcript is not a typed object later policy can
consume; a structural description plus prose is not a concrete realization.

**Required change:** separate records for (1) proposed structural change,
(2) concrete projection artifact, (3) verification obligation, (4) domain
observation. Demonstrate one path through all four, including a failing case
where the abstract path cannot compose concretely.

### What should remain intact (reviewer)

Exact occurrence-bound signature content; attribution of evaluations to the
bytes assessed; blocking rather than discarding stale targets;
coherence-guarded clustering; explicit inconclusive/inapplicable challenge
outcomes. "Substantive capabilities, not decorative provenance." Repairs must
strengthen these boundaries, not loosen them.

## Semantic findings

- **Claim form.** Sample recurrence, transformation invariance, and
  obstruction are distinct propositions. `associationKindForCandidate` treats
  measured full coverage as licensing universal-counterexample handling; an
  authored quantifier remains "future work". Required claim spec:
  `predicate + quantifier + scope + claim role + assessment context`.
  Truth, discrimination, and search eligibility must not collapse into one
  status.
- **Verification subject.** `deterministic` strength does not identify what
  was verified. Add a subject axis: annotation / realization / domain goal.
- **Geometry as operational relations.** Typed directed graph of admissible
  transformations with preconditions, preservation obligations, costs,
  realization evidence. Composability (A→B and B→C do not give A→C) must
  become a testable contract.
- **Complement geometry needs warranted exclusions.** Subtraction identifies
  the not-yet-excluded; it does not establish small/connected/feasible.
  Hard exclusions need scoped arguments; empirical disappointment should
  yield reversible preferences, not prohibitions.

## Epistemic findings

- Evidence supports a narrower claim than discovery effectiveness (Pilot-004
  `negative_inconclusive` stands; 10 `defensible_novel` occurrences ≈ 5
  themes; N2a admitted at model-judgment strength post-reduction; latest N2a
  child `challenged`, not surviving; XOR construction is a calibration chain).
- 6/12 vs 4/11 novel-occurrence counts must not displace the predeclared
  negative; occurrences/runs/themes/source-families are different units.
- The experimental contrast remains confounded (permuted-outcome control
  retained disclosing prose; same-family adjudication lanes).
- The historical-holdout retraction is important and correct; workflow
  withholding, chronology, and pretraining familiarity are separate
  contamination questions.
- N2a mathematical mistakes are corrected history, not new defects; the open
  issue is **propagation** — current summaries should surface reassessed
  strength directly with supersession links.
- The small formal results (XOR, 2ε bound, finite elimination) are correct,
  conditional, and identify obligations the system has not yet shown it meets.

## Positioning and recommended next step (reviewer)

Popper-style generate–test–constrain and MAP-Elites already occupy "failure is
data" and "search can use a map". The defensible opportunity: turn
heterogeneous histories of work into auditable, realizable decisions about
what to attempt next — and demonstrate when doing so earns its cost. Minimum
persuasive demonstration is a two-observation closed loop with an external
outcome check and full cost accounting.

> Pause theory expansion for one iteration. Fix assessment identity and
> derived views, implement typed projection and evidence admission, then run
> one prospective comparative episode.
