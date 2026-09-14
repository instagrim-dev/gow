---
artifact_kind: research-and-product-roadmap
status: proposed_not_authorized
revision: 0.3.0
supersedes_revision: 0.2.0
prepared_date_utc: "2026-09-13"
review_input_timestamp_utc: "2026-09-13T01:32:58Z"
repository_basis: cbc3685a0a2bb668f59fc7985680f4ccdb665957
scope: GoW shaping conception, early investment tests, evaluation custody, executable contracts, research validation, and commercialization
authorization: planning only; no spending, experiment dispatch, protocol freeze, or repository change
---

# GoW: from structural differences to a reusable search discipline

## Decision

Build toward a system that learns what successful Work must accomplish, selects or composes a realizable intervention, and checks it against the unchanged objective. Test the marginal value of history-conditioned shaping before committing to a broad implementation. Treat commercial usefulness and mathematical discovery as parallel outcomes, not substitutes for one another.

The next milestone is a small **residual obligation → applicable mathematical tool → executed check or intervention → attributed outcome** path, accompanied by a same-evidence comparison. Do not wait for the full shaping engine before asking whether the current shaping policy adds value.

**Three separate gates:** technical capability, research evidence, and permission to execute. Completing this roadmap document passes none of them. Proposed numbers below are explicit engineering/investment thresholds for a future authorized freeze, not measured results, sample-size guarantees, or scientific constants.

This revision retains the original nine-section organization, source distinctions, anti-goals, and responses to disconfirming evidence. Revision 0.3.0 folds in a second review's findings: it pre-commits the G4-lite run design and decision governance, adds an H0 guard to the spending rule, extends freeze discipline to comparator arms, fixes the custody-accounting boundary for the screen, operationalizes the custodian's interface boundary, and converts rule-withdrawal testing from prose into an owned exit condition. Repository observations [R] are tied to the stated commit; external methods [S] are their authors' contributions; design choices are proposals. Section 2 preserves the prior inspection's baseline rather than claiming the repository has not subsequently changed. [R4] was additionally read at the same pin for this revision.

**Provenance:** the preparation date is UTC; the review input is separately timestamped. A roadmap can be prepared after the commit it cites. Document creation, source revision, experiment registration, and execution are different events; this metadata does not claim a cryptographically attested creation instant or retroactive preregistration.

## 1. What the July 2026 publication actually supplies

The referenced article is *egg: Fast and Extensible Equality Saturation*, published online in Communications of the ACM on July 29, 2026. It explicitly identifies the original publication as POPL 2021; its arXiv preprint dates to April 2020. The useful foundation is therefore an existing body of techniques and implementations, not an engine first invented in July 2026. [S1, S2]

The fit is specific. E-graphs can compactly represent alternative expressions connected by equality rules; extraction selects a representation under a cost model. That supports GoW's encoding-removal and equivalent-construction search. It does not establish which abstractions a heterogeneous history of Work should use, discover every missing capability, prove a learned structural hypothesis, or guarantee that a preferred representation can be realized. [S1, S2]

Relevant neighboring components are already available:

| Foundation | Useful role in GoW | Remaining GoW obligation |
|---|---|---|
| egg / equality saturation | Represent and search equivalent expressions | Supply appropriate semantics, rules, side conditions, and extraction objective |
| egglog | Combine relational analysis with equality reasoning | Decide which relations have task relevance and which conclusions are warranted |
| Ruler | Infer candidate rewrite rules from a specified domain | Establish each rule with the appropriate validator and preserve its validity scope |
| Sketch-guided equality saturation | Guide search toward an intermediate structural pattern | Derive a useful sketch from Work history rather than merely hand-supplying it |
| DreamCoder | Learn reusable program abstractions alongside search | Demonstrate useful learning from GoW's particular histories and outcomes |
| EggMind, April 2026 preprint | Synthesize inspectable equality-saturation strategies using LLM guidance | Compare against this close neighbor instead of treating LLM + e-graph as sufficient novelty |

These are primary-source research connections, not claims that their evaluations validate GoW. [S2–S7]

**Research target:** the useful new contribution would be learning and exploiting task-relevant structure across heterogeneous Work, with attributable effects on subsequent decisions and checked outcomes. That target remains to be demonstrated.


## 2. Current maturity and owned workstreams

At the pinned baseline, durable evidence, exact witness checks, and attempt-to-output binding already provide an instrument foundation [R1]. Lean acceptance is explicitly separate from fidelity of the formal statement to the domain goal [R2]. The practitioner representation and action-selection prescriptions remain proposed [R3]. These observations are not a fresh build, execution audit, or inventory of other branches.

There is also an existing search-policy contract: immutable policy revisions, evidence-linked directives, and applied-bias logs [R4]. The omission in revision 0.1.0 was therefore not proof that policy persistence was absent from the repository. The missing roadmap obligation was to bind the **new shaping selector, rule selection, and representation choices** into an equally inspectable artifact. Reuse that substrate where it fits; do not create a parallel policy ledger by default.

| Axis | Pinned basis | Next demonstrable advance |
|---|---|---|
| Evidence custody | Durable records and attribution machinery | Exact subject/premise bindings, with measured custody cost |
| Domain checking | Witness checking and Lean adapter | Fixed-goal proof binding and independently replayable certificates |
| Structural interpretation | Proposed representations and hypotheses | Evidence-grounded residual obligation with an executable response |
| Search guidance | Scoped prior demonstrations | Incremental value over a capable method receiving the same history |
| Representation learning | Research agenda | Competing bases evaluated on future decisions |
| External task performance | Not established by these anchors | Protected, externally authored task families and independent reruns |
| Commercial repeatability | Separate and unestablished | Paid repeat use with observed delivery/support economics |

### Workstreams and responsibility

Names must be assigned in an implementation plan before the corresponding gate can be claimed complete. A role below is a requirement, not a claim that a collaborator has been recruited.

| Workstream | Accountable role | First deliverable | Staffing state in this roadmap |
|---|---|---|---|
| E0 — evaluation material and custody | Evaluation custodian | Source inventory, protected task packs, oracle, access ledger | Unassigned; must be distinct from the shaping implementer for external-evidence claims |
| T0 — transition and certificate contract | Engineering owner + semantics reviewer | A↔B↔C transition table implemented with mutation tests | Engineering delegation needed; reviewer unassigned |
| P0 — shaping-policy identity | Engineering owner | Immutable policy snapshot and decision/mutation trace | Reuse existing policy facilities after a fit check |
| X0 — early search-value screen | Evaluation custodian owns scoring; engineer owns runner | G4-lite same-evidence comparison | Requires E0, minimal T0/P0, and execution authorization |
| C0 — commercial discovery | Founder/commercial owner | Problem interviews and a scoped value hypothesis | Explicit time allocation; no assumed sales team |

**No independent evaluator available:** proceed with labeled development tests, not external validation. An isolated agent can reduce conversation leakage but shares model/training dependencies; a new seed or different document hash does not establish independent task authorship. Do not claim an unstaffed role is covered because its name appears in a table.

### E0: evaluation corpus is an explicit dependency

This revision supplies a collection plan, not a completed dataset. **Zero new sealed cases have been collected or independently audited by preparing this document.** The existing M1 budget-independence example is development material only.

Proposed acquisition channels are: provenance-linked examples and semantics from the cited EqSat work [S2–S5] for development; a custodian-authored generator against an implementation-independent task specification for sealed finite cases; and later contributor-supplied engineering histories for external transfer. Availability, license, suitability, and independence must be checked before admission. Public examples may be familiar to models and are not clean knowledge holdouts.

| Pack | Proposed size | Purpose | Who sees answers before testing |
|---|---:|---|---|
| Development | 12 open cases initially | Debug binding, certificates, and the runner | Implementer and custodian; never confirmatory evidence |
| G1 protected pack | 48 task cases: 24 applicable, 16 inapplicable, 8 underspecified | Check useful execution and both error directions | Custodian and outcome checker; implementer receives only interface/specification beforehand |
| G4-lite protected pack | 24 distinct task-and-history episodes | Early incremental-value screen | Custodian; separate from G1 and development items |
| G4 confirmatory pack | Not sized yet | Population-level effectiveness claim | Independent task custodian; size/design determined before execution from a separate design analysis |

A task is not a paraphrase count. Group by construction/template/source family; keep near-duplicates on one side of a split. Record exact objective, admissible semantics, source lineage, expected answer or oracle, permitted references, history construction, author/model provenance, exposure history, and family ID. Accept more than one valid tool or proof route; the oracle must judge the claim and scope, not demand a single reference name.

The custodian sees the supported interface but should not inspect the implementation's rule-specific examples, prompts, or failures before authoring the protected cases. **The supported interface means the registry's public statements: tool names, input types, premises, quantifier scope, and output kinds. It excludes worked examples, prompt/template text, failure logs, and development-case content.** Interface knowledge plus sensitivity calibration still couples task authoring to the registry; that residual registry-shaping exposure is not eliminated by author blinding and stays labeled on any resulting claim. Include tasks outside the initial registry's supported scope, conditions that invalidate an otherwise familiar rule, and alternative valid routes. An external domain review of protected-task assumptions is required before an external-evidence claim; when that reviewer is unstaffed, the pack's results are development evidence.

Before execution, seal source and answer manifests separately, test access controls, and scan adjudication inputs—including headings, paths, embedded metadata, and summaries—for arm identifiers. Retain raw outputs; any redacted view has a deterministic transformation record. Shared caches and prompts must not reveal another arm's outputs. An aggregate score leaked to the developer is still feedback; reuse after adaptation is development, unless a separately justified reusable-holdout procedure applies [S11].

**E0 exit:** all required cases exist; outcomes are independently checkable or explicitly adjudicated; family splits and authorization are fixed; access and leakage controls pass; no unresolved oracle disagreement is silently labeled truth; an external domain review of task assumptions is recorded, or its absence is recorded and the affected packs are labeled development evidence. Invalid cases discovered after execution are retained with an invalidity finding, not replaced until the result looks favorable.

## 3. Target architecture and transition contract

### A. Evidence and Work graph

Retain goals, contexts, attempted actions, outcomes, measurements, interpretations, proof/counterexample artifacts, source provenance, costs, and disagreements. Revisions change interpretations, not historical bytes. Hypotheses may motivate probes without becoming trusted equalities.

### B. Structural-intervention graph

Represent required capabilities, residual obligations, candidate transformations, applicability conditions, and alternative realizations. A repair edge is not necessarily an equality. A failed attempt and a successful repair must not be merged because they concern the same task.

### C. Scoped equivalence graphs

Represent equivalent expressions only under admitted rules and declared semantics. Guards, types, evaluation effects, and scope travel with those rules. Semantic resemblance proposes a correspondence; it never licenses an e-class union by itself.

### T0: transitions, failures, and dependency withdrawal

**Owner:** engineering owner; semantics reviewer approves the narrow language and trust assumptions. **Delivered during G1**, before any B→C rule admission; enforced again at G2. These are relation-specific contracts, not a universal confidence/promotion ladder.

| Transition | Required warrant | What the output establishes | Failure route |
|---|---|---|---|
| A → B: infer a residual or intervention | Identified evidence, exact objective, explicit interpretation | A proposed relation and next-action rationale | Missing evidence → unresolved residual; no invented mechanism defect |
| B → execution: select a tool/action | Applicable premises, compatible inputs, permission and remaining budget | Permission for this scoped attempt, not truth of the hypothesis | False premise → not applicable; unknown premise → applicability unresolved; authorization/budget failure → blocked |
| B → C: admit an equality rule | A specified guarded equivalence, semantics/quantifier scope, proof or exhaustive finite certificate with premise binding | Only the certified equality in that scope | Instance success or finite samples → remain B evidence; no generalized rewrite admission |
| C → B: extract a realization | Exact input, rule manifest, guards, extraction trace and predicted cost | Candidate realization plus a replayable explanation | Timeout → bounded best-found; unsupported derivation → rejected certificate |
| B/C → A: record execution/check | Actual output bytes, verifier identity, applicability, exact subject and costs | Whatever the check decided, at its actual scope | Invalid candidate, inconclusive result, timeout, and infrastructure failure remain distinct |
| A → policy revision | Explicit mutation trigger, eligible evidence, unchanged goal, bounded selection rule | A new policy proposal or an authorized policy-state update | Insufficient evidence → retain prior version or explicit unresolved update |
| New conflicting evidence → B/C dependents | Dependency graph identifies the affected warrant | Reassessment of exactly the dependent claims and decisions | Suspend affected eligibility; preserve historical checks; do not discard unrelated valid evidence |

**A checked intervention does not automatically become an equality.** For example, observing one successful realization of a repair establishes an instance-level result. A rewrite schema requires the different statement

`for every x in domain D, guard(x) implies eval(lhs(x)) = eval(rhs(x))`.

In plain language: the two expressions must have the same relevant behavior for every admitted input, not merely for the one successful attempt. A certificate exhaustive over a finite declared domain can support that domain, not a broader one. A proof of a guarded algebraic identity may support a larger domain if its premises and formalization match.

**C extraction has two checks.** Independent explanation replay establishes equality under the admitted rules; domain assessment establishes that the candidate meets the original objective. Equality alone does not establish improved performance, compliance with a non-equational requirement, or global optimality.

**Withdrawing a rule:** removing it from a rule list does not undo unions already made using it. Rebuild the affected C graph from valid seeds/rules, or use a verified dependency-aware repair mechanism. Re-evaluate downstream decisions whose warrants used that rule. An artifact with an independent surviving proof may remain usable. Preserve the old graph's provenance instead of relabeling its historical output.

```text
A: goal + Work history
  → B: candidate residual/intervention
  → applicability and authority check
      false premise       → not-applicable record in A → choose another tool or stop
      unresolved premise  → unresolved record in A → bounded evidence acquisition or stop
      no permission/budget→ blocked record in A → no unauthorized execution
      applicable          → direct realization, or C under admitted equalities
  → candidate + explanation
  → independent replay and domain check
      valid scoped result → A: scoped evidence + cost → explicit policy update
      counterexample      → A: refuted candidate/scoped claim → revise B
      unavailable/timeout → A: blocked/bounded result → fallback or stop
```

**T0 exit:** tests reject an instance-to-universal promotion, an unproved guard, incompatible semantics, a forged/reordered certificate, and stale-rule reuse; valid scoped transitions still succeed. Prevent refuse-everything conformance by exercising positive paths. T0 additionally records, for each admitted rule and dependent decision, the warrant dependencies that withdrawal will need. The full withdrawal test — admit a rule, form unions that depend on it, withdraw it, rebuild from valid seeds and rules, and verify that dependent decisions are re-evaluated while independently proven artifacts survive — is deferred to the G2 exit, where a real C graph exists; this sentence is the deferral decision, not an omission. This is one small executable contract, not a new graph platform.

### P0: shaping policy is a durable object

**Owner:** engineering owner. **First version required at G1; frozen per G4-lite run.** Use the existing policy revision/directive/provenance mechanism [R4] where compatible; record missing bindings rather than assume it already carries the fields below.

The conceptual snapshot identifies: parent/version; controller code or DSL; representation basis; tool/rule catalog; exact prompt/template and provider configuration when applicable; input evidence manifest; scoring/allocation/stopping rules; permitted mutation policy; and resource limits. Each decision records its snapshot, available alternatives, selected action, rationale, and observed costs. A mutation records triggering evidence, before/after identities, effective point, and authority.

These are fields to map to the existing implementation, not mandated new tables. No executable policy may exist solely as unretained conversational context. Reconstructing a stochastic run from recorded outputs is not the same as guaranteeing identical regenerated tokens; state the supported replay mode.

The controller's mutation rule is frozen across evaluation episodes. Permitted within-episode updates are part of the evaluated algorithm and must be recorded; changing the controller after viewing test results starts a new evaluation revision. Snapshot and freeze discipline applies to every evaluated arm, not only the treatment: a comparator whose scaffolding remains editable during or after sealing is not a capable baseline. EggMind's explicit EqSatL strategy artifact is a close precedent, not evidence that GoW has already implemented the corresponding binding [S7].

## 4. Maturity trajectory with an early investment checkpoint

Stages express capabilities and dependencies, not calendar commitments. E0, T0, and P0 accompany the first vertical slice. G4-lite runs before substantial G2/G3 expansion. A narrow early failure can stop investment in that implementation/domain without purporting to refute every possible GoW method.

### G0 — Stable research instrument

**Capability:** retain what was attempted, what was checked, what changed, and which evidence applies now.

**Work:** stabilize only the interfaces needed by the next slice. Preserve the store/provider/verification/review boundaries and outstanding integrity findings.

**Exit:** represent exact inputs, outputs, verification subject, decision policy, and costs without a new overarching state machine.

**Anti-goal:** another generic harness or review series with no pending decision it enables.

### G1 — Claim-to-tool execution, T0/P0, and a protected usefulness check

**Capability:** select an applicable mathematical operation and return a correctly scoped result. Start with observed-value invariance, monotonic accumulation under genuine trace extension, and finite equivalence. Other registry entries wait for an exercised need.

Each tool identifies its statement/implementation, input types, premises, supported quantifiers, outputs, and limits. Bind its invocation to the actual sentence or formal goal, measured object, transformation, and evidence. A theorem about two observed ratios is not automatically a theorem about an underlying probability.

**Self-contained example (development, not held-out evidence).** A deterministic local-search diagnostic on 20 fixed instances records 3 successes in 40 submissions at budget two, and the same 3 in 57 at budget three. The manuscript case called the latter rate budget-independent. Let `S` be old successes, `N>0` old submissions, `s` added successes, and `m>0` added submissions, with unchanged earlier observations. Then

`(S+s)/(N+m) − S/N = (N*s − S*m)/(N*(N+m))`.

The sign is negative at `S=3, N=40, s=0, m=17`. The ratio falls because added attempts all miss; no solved instance is lost. The counterexample refutes unrestricted observed-rate invariance, not a general success-probability claim. Exact sample binding and the trace-extension premise are part of the check. A restricted budget plateau is a different claim.

**Proposed G1 finite-pack gate (48 cases from E0):**

| Stratum | Requirement | What it prevents |
|---|---|---|
| 24 applicable cases | At least 23 correctly scoped completed checks; at most 1 false refusal; other non-completions also count against the 23 | Refuse-everything and blanket unknown policies |
| 16 inapplicable cases | Zero false acceptance/application certifications; all 16 identify an actual failed applicability condition | Forcing familiar mathematics onto the wrong subject |
| 8 underspecified cases | All 8 retain unresolved applicability and identify a missing premise; zero fabricated definite conclusions | Guessing a scope or premise to earn a pass |
| Across all 48 | Zero invalid results certified as valid; preserve every trace and outcome | Correct-looking output with a wrong binding or proof |

These are proposed **observed-case acceptance thresholds**, not bounds on population error probabilities. The allowances interact rather than compose: one false refusal plus any other non-completion already fails the 23-of-24 requirement; the bounds are not independent budgets. For intuition, even zero errors in 16 independent identically distributed Bernoulli trials has a one-sided 95% zero-event upper bound of `1 − 0.05^(1/16) ≈ 17.1%`; designed, related cases need not satisfy those assumptions. This pack is a bounded engineering gate, not evidence of a tiny real-world false-application rate. General error-rate claims require their own sampling design.

**Fixed-goal binding:** pin the allowed statement and environment; accept only evidence about that statement. Lean integration may be staged after an exact finite checker—the early comparison does not require a generalized proof-assistant interface first.

**Exit:** the finite-pack gate, T0's transition tests, and P0's decision replay pass. Existing M1 and published examples remain development material.

### G4-lite — early test of the shaping policy's marginal value

**Placement:** after the minimum G1/T0/P0 path, inside the first investment tranche; before full G2 integration and G3's synthesis build. Use a tiny bounded reference rewriter/enumerator if egg integration would delay the test. Charge that prototype effort; do not turn the early experiment into a second backend project.

**Question:** does an explicit history-conditioned shaping policy add enough value over a capable method with the **same history** to justify the next tranche?

A history-versus-no-history comparison alone confounds method with information. Proposed arms on E0's 24 protected episodes:

| Arm | Inputs and capability | Role |
|---|---|---|
| H0: no-history direct search | Current task, tools, checker, same task-directed resource ceiling; no prior-attempt history | Secondary diagnostic of the value of supplying history |
| H1: same-history direct search | Same task, raw history, tools/checker/model access and ceiling as HG; free to infer relationships | Primary capable comparator |
| HG: GoW shaping | Same raw evidence, explicit versioned shaping procedure, same tools/checker and ceiling | Proposed treatment |

Initial task domain: a pure finite expression language, for example 4-bit words with at most three variables and precisely specified operations. At most `16^3=4096` assignments permit a separate exact semantic oracle. Finite-domain claims remain finite. The objective combines semantic equivalence with a per-task optimization target selected by the custodian before arm execution. The target declares its cost model and tie rule in advance; a completion means meeting the exact objective, including the target threshold, under that declared model, and near-ties are resolved by the pre-declared rule rather than post-hoc measurement preference. Search effort and candidate execution cost are distinct measurements.

**Proposed population:** 12 history-informative episodes, 6 where history should add little, and 6 with irrelevant/misleading history. Distribute them over independently authored construction families. The custodian validates histories and targets without inspecting HG internals. These are 24 budgeted screen cases, not 24 independent domain replications; 3 arms give 72 primary episodes before any separately authorized reruns. Only the 12 history-informative episodes can discriminate shaping value; the screen's effective discriminating sample is 12, and the thresholds below are read against that fact.

**Arm parity.** All three arms are versioned and frozen under P0's snapshot discipline before sealed execution, with the same pinned model access, tool catalog, and checker versions. H1's scaffolding is authored or reviewed by someone other than the shaping implementer — the custodian or semantics reviewer — and that review is recorded, because a weakened comparator passes the margin threshold for free. A strengthened-after-sealing comparator equally voids the run.

**Custody accounting for the screen.** The per-arm resource ceiling covers task-directed search, inference, and checking, applied identically across arms. Mandatory custody writes — snapshots, decision traces, manifests — execute outside the ceiling but are metered and reported per arm, and the §5 full-cost figure accompanies the spending decision. Declare this boundary before execution; moving custody in or out of the ceiling after seeing results voids the comparison.

**Sensitivity calibration before sealing:** on open development cases, verify that the tool pool can express improvements and a reference procedure can reach them under the proposed budget. Avoid a floor/ceiling where no method can succeed or all succeed immediately. Preserve the calibration policy, not only its successful cases. Calibration is not permission to tune a protected task until HG wins.

**Default proposed primary endpoint:** number of protected tasks meeting the exact objective within the same task-directed resource cap. Custody is metered outside that cap and reported with the full-cost decision record; report total cost, quality, action trace, and per-stratum outcomes secondarily. Do not switch after seeing data to whichever metric favors HG.

**Run design and decision governance:** fix the number of runs per arm-episode cell before sealing — default three fixed seeds where providers are stochastic, one run only under a disclosed budget constraint. The evaluator compares exact run-summed integers: condition (b) requires an HG−H1 sum of at least `3 × runs`, and condition (c) permits a net H1−HG low-value/misleading sum of at most `1 × runs`; no separate per-arm rounding occurs. Condition (d) counts informative construction families with a positive run-summed HG advantage. The sealed batch's recorded outcome governs this tranche decision; separately budgeted repetitions inform the next evaluation revision and can neither rescue nor overturn the recorded decision.

**Proposed finite-screen spending rule:** continue to the next shaping tranche only if (a) no invalid result is certified, (b) HG's exact run-summed completions exceed H1 by at least `3 × runs`, (c) the net H1−HG run-summed loss across the 12 low-value/misleading-history tasks is at most `1 × runs`, (d) at least two informative construction families contain an episode with positive run-summed HG advantage, and (e) HG's exact run-summed completions are at least H0's overall. Complete the fixed batch under its authorized resource limits and retain all failures. These thresholds are deliberately visible investment judgments, not a power calculation or superiority certificate.

**Disposition:**

- Screen met: fund only the next bounded tranche; do not claim general superiority.
- H1 beats H0 but HG does not improve on H1: evidence may favor supplying history, not the GoW procedure. Prefer the simpler route pending contrary evidence.
- H0 beats both history arms: history handling is net-negative in this construction; repair history use before crediting or debiting shaping, and do not fund shaping on a margin over a history-confused H1.
- HG adds cost without the stipulated benefit on informative, measurable cases: stop or narrow this implementation/domain; a revised attempt needs a capped new plan and fresh evaluation.
- Leakage, invalid oracle, or floor/ceiling effects: the screen is inconclusive for shaping value and the investment gate is not earned. Repair the evaluation before building outward.
- With stochastic providers, follow the pre-committed run design and report seed/model sensitivity and uncertainty. Repetitions beyond the sealed design require a separate fixed budget and start a new evaluation revision — not retry-until-success, and not retry-until-failure.

This is a **spending gate, not a universal falsification theorem**. It does not test learned representation discovery, mechanism composition, or all future GoW implementations. Its purpose is to expose weak near-term economics before large engineering spend.

### G2 — Verified encoding removal and independent explanation replay

**Capability:** search equivalent implementations without weakening semantics, using one narrow language and a bounded engine.

Prefer a pinned egg adapter behind a small subprocess interface while Go owns orchestration/persistence [S1, S2]. Add egglog only for an exercised relational need [S3]. Do not port the engine to Go for stylistic consistency. Extend the early language only when its evaluator, guards, and equality contract remain explicit.

**Independent replay is an exit requirement, not polish.** Export the input/output terms, rule identities, substitutions, rewrite directions, guards, and dependency scope. A checker outside the search engine must replay the explanation against certified rules and bound premises. For the initial finite language, a separately implemented exhaustive evaluator can also validate endpoint equivalence. State shared trusted assumptions; separate processes alone do not ensure independent logic.

The engine's `Explanation::check_proof` checks explanations relative to supplied rules and documents limits on supported rule forms [S10]. It is useful, but it is not independent evidence that every supplied rule is semantically sound. The external checker must refuse unknown or uncertified steps. Mutate a rule, substitution, guard, source term, and final term: the replay must reject each relevant corruption.

**Exit:** a correctly checked cheaper realization; invalid-rewrite controls; independent explanation replay; graph/work/time bounds; separate predicted extraction cost and measured execution cost; and the T0-deferred withdrawal test — admit, union, withdraw, rebuild, verify — passing on this language's graphs. A budget stop is not saturation or global optimality.

**Anti-goal:** claiming this alone establishes learned geometry. A good equivalence optimizer can be a useful product even without incremental HG value.

### G3 — Residual-based shaping and composition

**Capability:** infer missing capabilities and construct a next attempt rather than only select a fixed menu item.

Use B's structured residuals—coverage gap, missing coupling, violated premise, excessive cost, unproved preservation, or unavailable evidence. Do not translate every epistemic gap into a physical or mathematical obstruction. Preserve interactions and viable alternative realizations.

Retrieve and compose intervention schemas, discharge preconditions, use C only for justified equalities, and record actual outcomes in A. Start with one bounded task family. A scheduling model with ordering/coordination constraints remains an option, not a second domain promised during the initial tranche.

**Exit:** before outcome observation, state a task-relevant capability requirement not handed to the proposer as the answer; construct a realization outside the supplied action menu; check structural fidelity and the original objective separately. Validate that successive transformations compose under compatible preconditions. Independently authored tasks and all failure routes are required.

**Anti-goal:** treating a plausible capability list, familiar theorem reproduction, or fixed-pool selection as new-mechanism discovery. Those can be useful under their own declared objectives.

### G4 — Confirmatory effectiveness evaluation

**Capability:** establish resource-bounded advantage for a specified population and implementation, not just pass a spending screen.

Build on G4-lite without reusing its exposed cases as confirmation. E0 must price/source a fresh task pack and appoint a custodian. Choose the task-family sampling design, practical margin, variance/dependence treatment, sample count and any sequential stopping rule before results. This roadmap cannot honestly supply a powered N before those choices and variance information exist.

Use a capable same-history baseline and a curation-only comparison. Where EqSat is central, use or compare a suitable unguided/sketch-guided strategy. EggMind remains a close strategy-synthesis comparator [S7]. A backend-by-guidance diagnostic can separate history-guidance value from engine value; do not require every ablation at once if a narrower experiment answers the pending claim.

Freeze the controller and catalog versions; permitted within-episode policy mutation is logged. Charge the full cost ledger in §5, **including custody, scope audits, reconciliation, maintenance, and independent replay**. Keep timeouts and verification failures in the primary denominator. Independent reproduction means a separate runner re-executes the pinned procedure, not a second reading of the same summaries.

**Exit:** uncertainty-qualified benefit on the declared population, assessed under the unchanged objective and total-cost endpoint, with independent rerun and leakage audit. A positive G4-lite screen is motivation, not part of this confirmatory sample.

### G5 — Representation learning and reusable capability libraries

**Capability:** learn useful Work descriptions and transformation libraries rather than merely choose a hand-authored vocabulary.

Compare bases through prospective decisions; record discarded distinctions and cases where they could matter. Learn candidate rules/fragments using appropriate methods, including Ruler or DreamCoder where suitable [S4, S6], but admit them into C only through T0. Every new rule's validity scope is checked separately from its usefulness.

Transfer records the target-domain mapping, imported assumptions, losses, and new proof/check obligations. E0 supplies new structures and sources, not more phrasings of the old tasks.

**Exit:** a learned abstraction improves later work without the evaluator encoding that abstraction as the answer; a separately assessed domain demonstrates transfer or identifies its failure. No stage-wide calendar commitment.

### G6 — Flagship mathematical program

**Capability:** obtain independently inspectable constructions, lemmas, reductions, or counterexamples at their claimed scope, including a defensible novelty assessment where novelty is the objective.

Choose one program with its exact target, intermediate obligations, and appropriate checkers. A nearby restricted theorem or finite case can be valuable without being the original conjecture. Historical reconstruction calibrates tooling; model familiarity with the literature is not an unseen-task guarantee.

**External-scrutiny deliverable:** record a named consenting domain collaborator or external reviewer, expertise relevant to the target, non-participation/conflict disclosures, the exact packet reviewed, written objections, and their disposition. A workshop/journal submission and received substantive reports are another auditable route; submission alone is not endorsement. Friendly discussion or same-family model quorum cannot silently satisfy this gate.

**Milestone ladder:** reconstruct a known result; obtain a new checked intermediate result; secure documented external scrutiny; complete the dependency chain to the target; publish and seek community assessment. Work may begin in parallel under its own budget; no solution date is promised or scientific attempt prohibited because the question is open.

### G7 — Durable discipline and platform

**Capability:** other groups use, implement, extend, and challenge the method without the original author reconstructing it for them.

Publish stable contracts, decision-policy artifacts, rule/checker interfaces, worked failures, task provenance, and replication procedures. Keep selectors, representations, backends, and verifiers replaceable. Track research generality, operational adoption, and commercial retention separately.

**Exit:** independently maintained implementations or integrations reproduce scoped results and expose disagreements. A successor that removes an unnecessary layer is a reason to simplify the method, not automatically proof of every original claim.

## 5. Correctness, total cost, and the cost of the discipline

The objective remains: **minimize total effort to obtain and check an acceptable result, subject to unchanged semantics and acceptance requirements**. Soundness, completeness, optimality, and statistical confidence are separate claims. More efficient sufficient checking need not weaken correctness; a budgeted search may still sacrifice completeness.

Record costs by activity and by resource. An activity label prevents omitting custody; resource dimensions prevent pretending human time and GPU time are already comparable.

| Activity | Include |
|---|---|
| Task/history preparation | Goal formalization, source acquisition, normalization, task-specific mapping |
| Search and construction | Inference, semantic retrieval, graph building, exploration, realization, failed attempts |
| Verification | Premise binding, proof generation, independent explanation replay, domain checking, counterexample search |
| Custody and review discipline | Provenance writes/reads, hashes, manifests, version reconciliation, scope audits, review/adjudication and release checks |
| Operations and maintenance | Adapter/rule maintenance, migrations, monitoring, reproducibility repairs, support |
| Evaluation overhead | Protected-case authoring, oracle validation, blinding, scoring and independent replication |

For each event record human minutes, relevant CPU/GPU time, tokens/API charges when available, elapsed time, and storage/other material costs. Mark unavailable measures unknown. Assign each event once; do not count the same reviewer hour both as an activity charge and an additional generic human surcharge. Report developer/evaluator effort even when it is excluded from a per-task runtime estimand.

Distinguish cold-start task cost, reusable setup, recurring cost, maintenance over a specified horizon, and the cost of conducting the research comparison. Adoption arithmetic may exclude genuinely sunk costs, but must display them separately. Evaluate both methods' actual overhead symmetrically; shared infrastructure is not a license to make GoW-only scope auditing free.

Conversion rates and reuse horizon must be declared before outcome-based adoption decisions. If unavailable, report a resource vector and leave monetary superiority unresolved. No undocumented amortization. Report the benefit **before and after custody costs** as a diagnostic, while using the full-cost figure for the end-to-end claim.

Cache checked results only across contexts covered by their dependency identities. Bound custody work itself; the minimal sufficient record is preferable to repeated reviews that cannot change a pending decision. Optimize the discipline when it dominates, without weakening what a claimed certificate establishes.

The proof agenda covers exact arithmetic transitions, fixed-goal binding, guarded equalities, compatible composition, and replay soundness. Learned representation quality and comparative performance require their own evidence; none of these is a general convergence guarantee.

## 6. Commercial trajectory

Sell checked improvements and fewer wasted attempts, not prize expectations. A useful engineering tool may earn revenue before the broad research hypothesis is established; payment does not establish the hypothesis.

Start with a narrow computation or data-transformation component whose behavior and optimization objective can be checked. The deliverable is a reproducible change or bounded informative diagnosis, with assumptions, evidence, and total costs. Avoid beginning with claims of correctness for entire distributed systems or high-stakes clinical/legal/financial decisions.

| Commercial stage | Deliverable | Required evidence/owner |
|---|---|---|
| Customer discovery | 3–5 problem interviews; one written value/integration hypothesis | Founder/commercial owner; proposed target, not reported conversations |
| First paid engagement | Fixed scope, acceptance terms, deployment boundary, support allocation | Buyer agreement and a staffed delivery plan |
| Repeatability | Similar delivery for unrelated customers | Reuse measured after adapter/review/support costs |
| Product/platform | Private/local integration and maintained domain packages | Repeat use, renewals, sustainable support economics |

A paying design partner is no longer a default first-90-days deliverable. A funded engagement can be accepted, but its sales, integration, and support work must replace roadmap capacity or have additional ownership. Record this as a scope decision, not invisible founder overtime.

Customer value is checked savings plus explicitly agreed value of avoided work, minus inference, integration, verification, custody, maintenance, and support costs. Measured savings and hypothetical avoided costs remain separate. No price or market-demand forecast is validated here.

**Prize mechanics remain only here.** Clay's Millennium prizes are potential upside, not operating revenue. Its rules require a qualifying publication, at least two years after publication, and general community acceptance before consideration; direct solution submissions are not accepted [S8, S9]. Refresh applicable rules and target status before a prize-focused campaign. An arXiv upload or nearby checked statement is not a payment entitlement.

## 7. Capacity-based first tranche and replanning points

The original per-stage dates are removed. The first tranche has a scope cap and decision checkpoints; calendar labels below are review windows after authorization, not promised completion dates.

**Proposed capacity for the initial investment decision:** 15–20 focused engineering person-days, 5–8 evaluation-custodian person-days, 2–3 semantics-review person-days, and up to 2 founder days for customer discovery. These are estimates to approve or revise before execution, not logged costs. Custodian/reviewer effort is additional capacity, not silently included in one engineer's week. If the engineer covers a second role, subtract that effort from implementation time and disclose lost independence. At most one major implementation slice is active. The first checkpoint's scope is expected to press against this estimate; if it does, the honest output is the named readiness/capacity blocker, not silent scope stretch or undisclosed role doubling.

| Checkpoint window | Maximum planned scope | Decision output | Excluded unless separately staffed |
|---|---|---|---|
| First roughly 30 calendar days | Minimum G1 + E0/T0/P0; tiny rewriter if needed; G4-lite once its inputs and ceilings are ready | Early spending decision, or an exact readiness/capacity blocker | Full egg/Lean integration, second DSL, scheduling synthesis, paid delivery promise |
| Remainder of the first roughly 90 days | Only if the screen earns investment: finish one G2 backend with independent replay, or prioritize the cheapest remaining binding/evaluation defect | Reproducible bounded capability; plan and price the next fresh test | Mandatory G3, powered G4, broad corpus expansion, simultaneous commercial deployment |
| Later tranche | G3 composition and G4 confirmation after their prerequisites | New authorization/scope decision based on evidence | Automatic advancement because time elapsed |

An outstanding E0 staffing or dataset dependency cannot be repaired by labeling self-authored tests independent. Ship a development result and hold the external-evidence claim. Missing input/tool/model budgets block dispatch, not the ability to prepare code or a proposed protocol. Existing preservation-pilot, historical-rerun, and other authorization boundaries remain unchanged.

**First commit sequence, when implementation is delegated:** (1) map T0/P0 onto existing contracts and add positive/negative boundary tests; (2) build the small claim-tool path and independent reference evaluator; (3) complete E0 inventory and access controls under the separate custodian; (4) freeze the early comparison's policy, task and cost manifests; (5) execute only under explicit ceilings; (6) record all results and the spending decision before expanding G2/G3. No general framework is a prerequisite for these commits.

## 8. Evidence that changes the roadmap

| Observation | Required response |
|---|---|
| Claim binding repeatedly selects the wrong subject despite correct arithmetic | Narrow to structured/approved statements; repair binding before adding tools |
| Inapplicable cases are refused but applicable cases are also mostly refused | Fail G1 usefulness gate; no conformance claim from safe inactivity |
| Same-history direct search matches/exceeds shaping in a sensitive G4-lite screen | Pause this shaping investment; simplify, narrow, or redesign under a capped new plan |
| H1 beats no-history H0 while HG adds nothing over H1 | Attribute benefit to history access, not GoW-specific processing |
| H0 outperforms both history arms in the screen | Treat history handling as net-negative in this construction; repair history use and re-screen before funding shaping |
| All arms hit the floor/ceiling or the oracle is invalid | Classify the screen as inconclusive; repair measurement, not the narrative |
| Held-out success relies on leakage or registry-shaped tasks | Withdraw the unsupported generalization, mark affected cases development/contaminated, move to genuinely separate task authoring and a new sealed pack |
| EqSat helps but the extra GoW layer does not | Ship the useful backend; retain the broader method claim as unestablished |
| A successful intervention is admitted as an unproved general equality | Suspend affected rule/graph authority; rebuild dependent equivalence state; inspect derived claims |
| Certificate replay fails despite engine acceptance | Reject the certificate at that scope; retain the candidate for independent investigation |
| An arithmetic/semantic rule is valid but custody costs erase its search benefit | Reduce redundant custody or find justified reuse; do not hide the cost |
| Learned bases fit history but fail on prospective interventions | Change the selection criterion; do not relabel past misses as prediction success |
| A proposed boundary excludes a known valid solution | Reopen that boundary and dependent decisions; do not blame projection automatically |
| Customers require bespoke work with little reuse | Remain services-led or narrow scope; no unsupported platform economics |
| Other groups cannot reproduce the method without its author | Improve the operative contract and teaching before claiming a discipline |

An investment pause is not a universal theorem about all GoW variants. Conversely, an invalid early test does not earn permission to expand unboundedly. Separate the failure of an instrument, a measurement, a scoped hypothesis, and an adoption decision.

## 9. Completion rule and north-star demonstration

> Given several attempts at a fixed task, GoW identifies a task-relevant distinction not explicitly supplied as the answer, uses it to retrieve or compose an applicable mathematical operation, produces a concrete next attempt, and checks that attempt against the original objective. A versioned policy and checked transition chain attribute what the representation contributed and record total effort, including evidence custody.

The first tranche need not achieve every part of this north star. It must expose whether the simplest operative policy has enough marginal value to justify building toward it.

Do not make all maturity levels prerequisites for revenue or publication. Do not count a paid engagement as scientific generality, a formally valid output as evidence of search advantage, a task-specific failure as universal refutation, or a recorded map as a productive search policy.

**Revision 0.3.0 is complete as a planning artifact when:** both review rounds' dispositions are explicit; E0/T0/P0 have named responsibilities and exit conditions; both G1 error directions are specified; G4-lite's spending rule, run design, arm parity, custody-accounting boundary, and H0 guard are fixed before sealing; the withdrawal test has an explicit owner and exit; costs and capacity include the discipline; and unanswered staffing/budget questions stay visible. It does not freeze those proposed thresholds or authorize execution.

## References and inspected repository anchors


### Repository anchors (source inspection, not fresh execution)

- [R1] [GoW changelog at cbc3685](https://github.com/instagrim-dev/gow/blob/cbc3685a0a2bb668f59fc7985680f4ccdb665957/CHANGELOG.md): v44–v46 review, admission, and attempt-output binding descriptions.
- [R2] [Lean verifier adapter at cbc3685](https://github.com/instagrim-dev/gow/blob/cbc3685a0a2bb668f59fc7985680f4ccdb665957/internal/lean/verifier.go): accepted formal statement versus unverified domain-goal fidelity.
- [R3] [Method contract v0.1.0 at cbc3685](https://github.com/instagrim-dev/gow/blob/cbc3685a0a2bb668f59fc7985680f4ccdb665957/docs/book/method-contract.md): proposed practitioner scope, representation selection, allocation, and authority tiers.

### Primary research and institutional sources

- [S1] Willsey et al., [*egg: Fast and Extensible Equality Saturation*](https://doi.org/10.1145/3815481), Communications of the ACM 69(8), 105–113; online July 29, 2026. The article identifies its original POPL 2021 publication.
- [S2] Willsey et al., [original egg paper](https://arxiv.org/abs/2004.03082), arXiv first submitted April 7, 2020; POPL 2021; DOI 10.1145/3434304.
- [S3] Zhang et al., [*Better Together: Unifying Datalog and Equality Saturation*](https://arxiv.org/abs/2304.04332), PLDI 2023.
- [S4] Nandi et al., [*Rewrite Rule Inference Using Equality Saturation*](https://arxiv.org/abs/2108.10436), 2021; DOI 10.1145/3485496.
- [S5] Koehler, Trinder, and Steuwer, [*Sketch-Guided Equality Saturation*](https://arxiv.org/abs/2111.13040), arXiv first submitted 2021.
- [S6] Ellis et al., [*DreamCoder: Growing generalizable, interpretable knowledge with wake-sleep Bayesian program learning*](https://arxiv.org/abs/2006.08381), arXiv first submitted 2020.
- [S7] Yin et al., [*LLM-Guided Strategy Synthesis for Scalable Equality Saturation*](https://arxiv.org/abs/2604.17364), April 19, 2026 preprint (EggMind). Reported capabilities are the authors' findings, not an independent replication in this review.
- [S8] Clay Mathematics Institute, [Millennium Prize Problems](https://www.claymath.org/millennium-problems/).
- [S9] Clay Mathematics Institute, [Rules for the Millennium Prize Problems](https://www.claymath.org/millennium-problems/rules/).


### Additions and source-check scope for revision 0.2.0

- [R4] [Search-policy contract at cbc3685](https://github.com/instagrim-dev/gow/blob/cbc3685a0a2bb668f59fc7985680f4ccdb665957/docs/search-policy.md): immutable revisions, directives, provenance and applied-bias logs. Read for this revision; a documented current substrate, not proof that every new shaping field is implemented.
- [S10] [egg 0.11.0 Explanation API](https://docs.rs/egg/0.11.0/egg/struct.Explanation.html): explanation serialization and `check_proof` relative to supplied rules, including its documented supported-rule limitation. This motivates, but does not implement, the independent replay requirement.
- [S11] Dwork et al., [Generalization in Adaptive Data Analysis and Holdout Reuse](https://arxiv.org/abs/1506.02629), 2015. Supports the warning that adaptive feedback can overfit a reused holdout; it does not validate the proposed finite-pack thresholds.

The original [R1–R3] baseline and [S1–S9] bibliography are retained. The reviewer supplied fresh spot-checks of [R1–R3], [S1] and [S7]. In preparing this revision, [R4], [S2], [S7], [S9], [S10], and [S11] were read; direct retrieval of [S1] was blocked by HTTP 403, so no new independent verification of its July date is claimed here. No source supplies the proposed task counts, staffing allocations, or spending thresholds: those are explicitly authored design choices.

### Changes in revision 0.3.0

Revision 0.3.0 responds to a second review of revision 0.2.0. That review independently confirmed that [R4] and [S10] resolve at their stated pins and that the revision's arithmetic identities check out (`(S+s)/(N+m) − S/N = (N·s − S·m)/(N(N+m))`; `1 − 0.05^(1/16) ≈ 17.1%`; `16^3 = 4096`; pack and arm counts). No new sources were added; every change is a design pre-commitment:

1. **G4-lite run design and governance** (finding: the spending rule fired deterministically on an acknowledged-noisy single run): pre-committed runs per cell, threshold evaluation on per-run averages rounded against funding, and sealed-batch decision authority that later reruns cannot rescue or overturn.
2. **H0 guard** (finding: HG could pass by out-margining a history-confused H1 while losing to the no-history arm): spending-rule condition (e), a new disposition branch, and a new §8 trigger row.
3. **Arm parity** (finding: only HG received freeze discipline; H1's strength was authorship-dependent): P0 snapshot/freeze discipline for all arms, non-implementer authorship or review of H1's scaffolding, pinned model/tool/checker versions.
4. **Custody-accounting boundary** (finding: whether custody counted inside the per-arm ceiling could flip the screen): custody metered outside the ceiling, reported per arm, boundary declared before execution.
5. **Custodian boundary** (finding: "supported interface" was asserted, not defined, and calibration re-couples authoring to the registry): interface contents defined, residual registry-shaping exposure labeled, external domain review made an E0 exit condition with a stated fallback.
6. **Withdrawal test** (finding: rule withdrawal was specified in prose but never tested): warrant-dependency recording added to T0's exit, and the admit→union→withdraw→rebuild→verify test added to G2's exit as an explicit deferral decision.
7. **Smaller items**: G1 stratum interaction stated; G4-lite per-task target requires a pre-declared cost model and tie rule; §7 names the expected capacity pressure and its honest output; the screen's effective discriminating sample (12 informative episodes) is stated next to the population design.

The 0.2.0 thresholds themselves (48/24 case counts, the ±3 margin) are retained unchanged: the second review challenged their decision integrity, not their magnitudes, and magnitude changes without new evidence would be tuning, not correction.
