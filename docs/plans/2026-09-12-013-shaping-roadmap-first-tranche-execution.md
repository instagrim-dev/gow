# Shaping-roadmap first tranche: execution record (T0/P0 fit-check, G1 seed tool 3, E0 inventory)

- **Status:** executed slice + recorded blockers; planning references remain proposals
- **Roadmap:** `gow_shaping_maturity_roadmap_v0.3.0.md` (external planning artifact, operator-held; revision 0.3.0)
- **Roadmap pin:** `cbc3685a0a2bb668f59fc7985680f4ccdb665957`
- **Execution baseline:** `e8c3da2` (HEAD at execution start; the pin's successor commits already delivered part of G1 — see fit-check)
- **Authorization:** operator instruction to begin execution (this session). Spending, provider dispatch, protocol freeze, and sealed-pack evaluation remain unauthorized; this record names those blockers instead of claiming them passed.

The roadmap's first commit sequence is: (1) map T0/P0 onto existing contracts with boundary tests; (2) build the small claim-tool path and independent reference evaluator; (3) complete E0 inventory and access controls under the separate custodian; (4) freeze the early comparison's manifests; (5) execute under ceilings; (6) record results and the spending decision. This tranche executes (1)–(3) to the extent a single operator-session honestly can, and records why (3)'s custodial half and (4)–(6) are blocked.

## Survey result: what the pin's successors already delivered

Between the roadmap pin and `e8c3da2`, work landed that satisfies part of G1 before this tranche began:

- **G1 seed tools 1–2 exist:** `internal/measure` (commit `a4ba2b0`) implements observed-value invariance and monotonic accumulation under genuine trace extension — exact `big.Int` cross-multiplication, the `N*s − S*m` sign rule, per-instance trace-prefix nesting verification, refusal of probabilistic claims with the reason recorded, and fixed-template certificates with scope guards.
- **The roadmap's development example is live:** the budget-independence (M1) case and its bridge-row diagnostic are covered by the preservation-pilot review at `e8c3da2`. Per the roadmap, it stays development material, never confirmatory evidence.
- **P0's substrate exists as documented:** the search-policy layer ([`docs/search-policy.md`](../search-policy.md), `internal/policy`, migration v19) provides immutable revisions, typed evidence-linked directives, per-directive provenance, and the applied-bias log.

Execution therefore started by mapping, not rebuilding.

## T0 fit-check: roadmap transitions → existing contracts

Relation-specific contracts, per the roadmap — not a universal promotion ladder. "Evidence" cites code/tests at `e8c3da2` plus this tranche's addition.

| Roadmap transition | Existing contract | Evidence | Gap |
|---|---|---|---|
| A → B: infer residual/intervention | Candidate invariants, frontier proposals, review findings, projection obligations are the B-objects; evidence rows are A | `internal/invariant`, `internal/frontier`, `internal/review`, `internal/projection` | No single first-class "residual obligation" type; nearest are review findings + projection obligations. Acceptable: the roadmap defers new types to exercised need |
| B → execution: select tool/action | `measure.Binding` premise checks (condition comparability, range membership, trace nesting); `witness.ExecuteAttempt` attempt-to-output binding | `internal/measure/certificate.go` applicability refusals; `internal/witness/attempt.go` (C3 remediation) | No unified tool registry object. Deliberate: three seed tools do not yet justify a registry abstraction (roadmap: "other registry entries wait for an exercised need") |
| B → C: admit equality rule | **New this tranche:** `internal/finite` — exhaustive finite-equivalence certificates are the first admission-grade warrant form | `internal/finite/certificate.go`; instance-to-universal rejection enforced as a verdict type (`INSTANCE_EVIDENCE_ONLY` can never be a domain verdict) and tested | No C graph exists yet (G2). Admission remains a review decision citing a certificate; no code path can admit a rule from instance evidence |
| C → B: extract realization | Not built | — | Correctly absent until G2 (equality-saturation adapter). Recorded as not-yet, not as gap-to-patch now |
| B/C → A: record execution/check | `review.CheckRecord` with executed-vs-blocked outcome separation; verifier tiers declare their subject | `internal/measure/review_adapter.go`, `internal/finite/review_adapter.go`, `internal/verify` (`Subject()` separates annotation from domain goal) | None material for this tranche |
| A → policy revision | Evidence-cohort-hashed immutable revisions; one admission gate (`policy.AdmitProposedDirective`); weight capped at evidence-derived medium | `internal/policy/admit.go`, `boundary_test.go`; `docs/search-policy.md` | The **shaping selector** the roadmap evaluates does not exist yet; P0's obligation is to bind it here when it does (below) |
| Conflicting evidence → dependents | Challenge boundary deltas + invariant lifecycle transitions carry typed refinements | `internal/policy` D5 consumption; challenge lifecycle | **No general warrant dependency graph.** Per roadmap 0.3.0 this is recorded T0 debt: dependency recording accompanies rule admission when a C graph exists; the withdrawal test (admit → union → withdraw → rebuild → verify) is a G2 exit condition |

## P0 fit-check: roadmap snapshot fields → search-policy substrate

| Roadmap P0 field | Existing binding | State |
|---|---|---|
| Parent/version | `search_policy_revisions.revision`, evidence-cohort hash | Bound |
| Input evidence manifest | `search_policy_provenance` per directive | Bound |
| Decision trace (snapshot, alternatives, selected, rationale) | Applied-bias log (`frontier_generation_policy`): per-proposal net bias + reason class, keyed to revision | Bound for the rerank lever; a future shaping selector must write the equivalent trace |
| Mutation record (trigger, before/after, authority) | Idempotent revision creation; inert-proposal ledger | Bound |
| Controller code/DSL identity | Not applicable yet — no shaping controller exists | **Open obligation**, recorded here rather than assumed satisfied |
| Prompt/template + provider configuration | `provider_invocations` rows carry role and payload | Bound at invocation level; per-snapshot aggregation open until a controller exists |
| Resource limits | Experiment arm budgets (`internal/store` experiment schema) | Bound at experiment level |

Conclusion: the substrate fits; **no parallel policy ledger is needed** (roadmap: reuse after fit check — this is that check). Two open bindings (controller identity, per-snapshot prompt aggregation) attach to the not-yet-built shaping selector, not to existing code.

## G1 seed registry state

| Seed tool | Package | State |
|---|---|---|
| Observed-value invariance | `internal/measure` | Shipped (`a4ba2b0`) |
| Monotonic accumulation under trace extension | `internal/measure` | Shipped (`a4ba2b0`) |
| Finite equivalence + exact exhaustive evaluator | `internal/finite` | **This tranche.** 1–8-bit words, closed variable set, total operators mod 2^width, division deliberately absent; canonical-order deterministic counterexamples; exhaustiveness cap with refusal (never sampling passed off as enumeration); `INSTANCE_EVIDENCE_ONLY` verdict type making instance-to-universal promotion unrepresentable |
| Idempotence, counterexample search (registry seeds 4–5) | — | Deferred to exercised need, per roadmap G1 |

`internal/finite` is also the exact semantic oracle for the G4-lite screen's 4-bit/3-variable task domain (`16^3 = 4096` assignments, well under the cap).

## E0 inventory and custody state

**Corpus inventory at `e8c3da2`:** `corpus/{authoring, evidence-supplement, experiments, research, target, train}` — all authored inside this development effort with model assistance, plus `fixtures/` and `testdata/`.

**Labeling decision:** every existing corpus item is **development material**. Zero sealed packs exist. Nothing in the current tree qualifies as a G1 protected pack, a G4-lite protected pack, or external evidence, because:

1. **The evaluation custodian is unstaffed.** The roadmap requires the custodian to be distinct from the shaping implementer for external-evidence claims. This session is one operator-directed agent; relabeling self-authored tests "independent" is expressly prohibited by roadmap §7.
2. **No sealed-pack authoring has been authorized.** Task-pack authoring under access controls is custodian work; doing it from this seat would burn the tasks (implementer exposure) before they were ever protected.

**Consequence (recorded, not lamented):** the G1 finite-pack gate (48 cases) and G4-lite (24 episodes) cannot be honestly run until a custodian exists. Development-labeled testing continues; external-evidence claims are withheld. This is exactly the roadmap's "no independent evaluator available" branch.

## Spending-gate position

| G4-lite prerequisite (roadmap §4) | State |
|---|---|
| Minimum G1 tool path | Met at development level (three seed tools, premise-checked, certificate-emitting) |
| T0 transition mapping + boundary tests | Met (this record + `internal/finite` tests; existing `internal/measure`, `internal/policy` boundary tests) |
| P0 policy identity | Substrate fit-checked; controller binding obligation attaches to the future selector |
| Exact oracle for the screen's task domain | Met (`internal/finite`) |
| E0 protected packs | **Blocked: custodian unstaffed; zero sealed cases** |
| Arm scaffolding (H0/H1/HG) frozen under parity discipline | Not built; H1 authorship/review requires a non-implementer |
| Execution authorization + resource ceilings | **Not granted; no provider dispatch performed in this tranche** |

**Spending decision status:** not reached — its inputs do not exist yet. The cheapest path to the screen is staffing the custodian role (or accepting a development-labeled, explicitly non-confirmatory dry run whose results cannot count toward the spending rule).

## Deliberately excluded from this tranche

Per roadmap first-30-days exclusions: no egg/Lean integration expansion, no second DSL, no scheduling synthesis, no new orchestration layer, no paid-delivery promise, no provider dispatch, no reinterpretation of previously gated pilots. Preservation-pilot, historical-rerun, and other authorization boundaries are unchanged.
