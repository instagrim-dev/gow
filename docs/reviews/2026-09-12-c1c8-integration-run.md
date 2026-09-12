# Review-run: one-obligation integration (C1–C8), 2026-09-12

> [!IMPORTANT]
> **Revision 2 (2026-09-12, after external critique of revision 1 at `00c3c58`).**
> Revision 1's closing decision `ELIGIBLE_TO_ADVANCE` is **retracted**: it
> conflated the coverage projection of a supplied review record with candidate
> eligibility, and rested on an unexamined C7 branch. Revision 1's text remains
> in git history (`1c9b412`); this revision supersedes its decision layer,
> retains its execution evidence, and adds a second, strengthened execution
> (run 2) that **demonstrated the previously open C7 defect (F-2)**.

Executed per `docs/reviews/prompts/README.md` →
`docs/reviews/prompts/recipes/assessment-admission-decision.md` under
`docs/reviews/prompts/review-contract.md`. This bundle records executed
checks; it is not whole-project completeness or scientific success.

## 1. Decisions (separated), policy, and scope

The critique's central correction is adopted: three different questions get
three different answers. None inherits another's verdict.

| Decision | Answer this bundle supports |
| --- | --- |
| May the exercised candidate guide the next search action in the final tested state? | **No.** The selector excludes it after the historical replay (run 2, C7). |
| Did the authority implementation preserve the tested safety boundaries? | **Partially.** Withholding, relevant-vs-unrelated discrimination, obsolete-restoration prevention, and membership-exact admission held. Preservation of valid current authority did **not** (F-2). |
| Has the one-obligation recipe been validated sufficiently to proceed to preservation testing? | **WITHHOLD** (revised from revision 1's eligibility and below the critique's floor of `UNDETERMINED`): run 2 examined the missing branch and demonstrated a blocking nonconformance of the obligation. Preservation testing remains gated. |

- **Policy provenance.** Run 1's policy was persisted *after* C1–C7 with an
  authored `conforms` — a predeclaration failure, acknowledged. Run 2 persists
  policy **P1r2** and the applicability decision **before any observation row**
  on a strictly monotonic store clock, and its decision name states explicitly
  that it governs the review record, *not* candidate targetability. The
  assessment outcome in run 2 is **derived from observed case results at
  runtime** (it recorded `nonconforms` because C7's positive branch failed),
  not authored in advance.
- **Pinned tree (run 2).** Base commit `7c54dcc` **plus** the retained
  complete manifest: tracked patch
  `evidence/2026-09-12-worktree-7c54dcc-tracked.patch`, untracked source
  archive `…-untracked.tgz`, and its sha256 manifest `…-untracked.sha256`.
  This identifies the bytes actually built. Run 1 (base `7e24160`) remains
  identified only by base + disclosure — its results beyond what run 2
  reconfirmed carry that limitation.
- **Procedure hashes (sha256).** Run 1
  `f0437229d794e2697443c0b97b382578cdf98a915d425f139e0c0f7410da6cce`; run 2
  `7c37b47857c887f3afb37332ad0a1e0ae41c316e08b2d7857cbe0171629cec82`
  (`evidence/2026-09-12-c1c8-v2-procedure.go.txt`). Contract/recipe hashes as
  in revision 1. Mode both runs: review-only + disposable test on the real
  migrated store (`t.TempDir()` SQLite), zero provider calls, zero production
  writes; disposable sources deleted after retention.

## 2. Responsibility and critical path

Unchanged: path `assessment-admission-decision`; owners `assessment-identity`
(07), `evidence-admission` (08), `decisions` (16); `verification` (12) as peer.

## 3. Findings and case results

### Run 2 case results (`evidence/2026-09-12-c1c8-v2-run-final.log`)

| Case | Result | Strengthening applied vs run 1 |
| --- | --- | --- |
| C1 | PASS | Targets parsed from the typed `GenerationRequest.targets` JSON field, not substring search. |
| C2 | PASS | Population compared as a **membership set**, not a count. |
| C3 | PASS (narrowed claim, see below) | Tuple `7,5,5,5` **derived** from a declared bounded procedure (equal-denominator probe `x=y=z=round(3n/4)`, n=7) whose output *is* the checked tuple; A1 membership asserted **set-equal** to D0 ∪ {admitted signature}. |
| C4 | PASS | Unchanged semantics; typed-target assertion. |
| C5 | PASS | Unchanged semantics; typed-target assertion. |
| C6 | PASS | Engineered to end `surviving` over A1 (locality-local admitted signature) and the positive next decision **asserted**: candidate targeted again. |
| C7 negative | PASS | Replay reproduces its bounded A0 result; population rows durable and distinct. |
| C7 positive | **FAIL → F-2** | Precondition enforced (`t.Fatal` if C6 ≠ surviving), branch executed, displacement demonstrated. |
| C8 | PASS | Policy predeclared; assessment outcome **derived** (`nonconforms`); projection tracked it (`WITHHOLD`); determinism byte-checked; staleness produced by a **real pipeline change** (second admission → A2, dependency re-derived from live store) → `UNDETERMINED [stale_dependency]`. Coverage documents retained (`…-v2-coverage-1.md`, `…-v2-coverage-3-stale.md`). |

### F-2 — demonstrated defect (new in run 2), primary owner `decisions` (16) with `assessment-identity` (07)

A historical replay **displaces valid current authority**. Sequence: C6 earns
`surviving` over the *current* population A1 (campaign `run_…WVKVBC5Q`);
candidate is targeted. A discovery-population replay then re-earns its bounded
`surviving` over A0; because current-state authority follows the **latest
state transition's campaign** (`invariantStateAuthorityJoin`), the next
generation excludes the candidate as `authority_assessed_obsolete_population`
— naming A0 as assessed — even though a compatible current surviving
assessment exists one transition earlier. The failure direction is
conservative (loss of availability, not unsound targeting), but it violates
the recipe's C7 requirement that the current decision "select the compatible
current context, not the last execution." If deliberate, it needs an explicit
contract decision; nothing in the code comments claims it is. Severity:
moderate; confidence: high (reproduced, full identities in the run-2 log).
Smallest coherent remedy: authority selection should prefer the latest
*population-compatible* campaign rather than the latest campaign
unconditionally. Discriminating regression check: exactly the run-2 C7
positive branch.

**Contradicted artifact (verified 2026-09-12 14:06):** the concurrently
generated export `docs/reviews/evidence/COVERAGE.md` (untracked at
verification time; policy `rpol_01M2AQT7G0HXWKFJ6Q4VTNBJKR`, generated 13:14,
**before** F-2 was demonstrated) still projects `ELIGIBLE_TO_ADVANCE`. Its own
records carry the F-2 signature: check C6 records `state_after=surviving` over
the current population A1, and check C7 then records
`current_selection_excluded=true` — i.e., the compatible current authority was
displaced, yet the governing assessment remains `conforms`. Its generator
(`TestIntegrationCurrentAssessmentAuthorityObligation`) asserts the exclusion
as C7's *pass* condition, so regeneration reproduces the eligibility verdict;
the export is not hand-editable by design. That gate encodes the F-2 behavior
as conformance: any `ELIGIBLE_TO_ADVANCE` derived from it is superseded by
this bundle's demonstrated nonconformance until remediation handoff 1 lands
(fixing the gate or recording the explicit contract decision).

**Resolution (2026-09-12 14:2x, handoff 1 landed):** current-authority
selection is now population-compatibility-aware
(`store.GetLatestCompatibleAuthority` consulted at the frontier decision
boundary): a replay's latest transition no longer displaces a campaign that
assessed the current compatible population and earned the same targetable
state; a compatible campaign that earned a *different* state still grants
nothing, so the 5b negative control (weaken) still excludes. The committed
regression `TestIntegrationReplayPreservesCompatibleCurrentAuthority` is
exactly the run-2 C7 positive branch and passes; the obligation gate's C7 was
corrected to assert preservation (`compatible_current_authority_preserved`)
instead of encoding the displacement as its pass condition, and
`COVERAGE.md` was regenerated from the corrected gate. F-2 is closed for the
exercised path; its `ELIGIBLE_TO_ADVANCE` verdict is no longer contradicted
on this point.

### F-1 — reclassified per critique: **reproduced admission omission, supported by a static root-cause trace**

Retained from revision 1 (runtime evidence:
`evidence/2026-09-12-c1c8-run-f3-shadowing.log`; static trace: per-proposal
`INSERT OR IGNORE` in the R6 marker write). Repair principle revised per the
critique: the acceptance requirement is **per-evaluation visibility under
exact context** — every applicable evaluation remains independently visible to
admission even when the proposal already carries an earlier marker or
decision — *not* strength-ranked replacement. Companion constraint: finer
marker granularity must not let repeated checks of the same underlying
observation inflate the atlas. The run-1 same-proposal failure **remains part
of the result**; the distinct-proposal layout in both final harnesses is a
diagnostic restriction, is hereby made explicit, and narrows what C3 admission
evidence covers (distinct-proposal path only).

**Resolution (2026-09-12 14:3x, handoff 2 landed):** the `evaluated_failures`
re-entry marker is now keyed **per evaluation** (migration v45 rebuilds the
per-proposal-PK table in place, preserving rows and immutability triggers;
fresh stores carry the new shape from the baseline DDL). Admission already
keyed its decision ledger per evaluation, so the later witness-checked failure
of an already-decided proposal now simply appears and is rule-admitted under
its own context; the earlier withheld model judgment is **preserved, not
replaced** — visibility, not strength ranking. The no-inflation companion holds
structurally: materialization uses the stable logical identity
`frontier-proposal:<id>`, so a second admitted evaluation of the same proposal
produces a new revision of the same approach and the current-heads population
does not grow a second member. Committed regressions:
`TestIntegrationLaterStrongerEvaluationReachesAdmission` (the exact run-1
shadowing case: withheld model judgment → witness-invalid → rule-admitted
exactly once, rerun skips both, population +1 total across two admitted
evaluations) and `TestMigrateV45RebuildsPerProposalMarkerTable` (upgrade path
on a real pre-v45 store). The distinct-proposal restriction on C3's admission
evidence is now unnecessary for *visibility*; mechanism-to-output attribution
(handoff 3) remains open.

### C3 — narrowed claim (adopted verbatim in substance)

What both runs establish: exact arithmetic checking, evaluation persistence,
rule admission, and membership-exact population growth for a supplied invalid
tuple. Run 2 adds procedure↔output agreement: the tuple is computed by the
declared bounded probe, and its failure is recorded as refuting that bounded
attempt only. What neither run establishes: that a proposal's *executed
mechanism* produced the tuple — the proposal-to-attempt binding is still a
declared fixture relationship, now stated rather than implicit. Full
mechanism-to-output attribution remains future work at the adapter boundary.

### Protections observed (retained)

Withholding control; relevant-vs-unrelated change discrimination; obsolete-
replay restoration prevention; durable per-campaign population identity;
typed-payload decision boundary (run 2); derived-not-authored assessment
outcome with `WITHHOLD` projection (run 2).

## 4. Checks, attempts, and resources

**Attempt ledger (corrected: revision 1 said "four failed runs" while
enumerating five; five failures + one success = six executions is right for
run 1).**

| # | Procedure rev | Outcome | Classification |
| --- | --- | --- | --- |
| 1.1–1.2 | disposable-1 | build failure | environment (concurrent `problemStore.GetReviewPolicy` interface change; second attempt authorized as diagnosed harness error) |
| 1.3 | disposable-1 | assertion failure | harness gap (deriving fixture generator emits no proposals for this corpus) — logs `…-c1c8-run1-attempt3.log` |
| 1.4 | disposable-1a | assertion failure | **completed negative result — F-1 manifestation** (`…-run-f3-shadowing.log`); the distinct-proposal variant afterward is diagnostic follow-on, not a retry erasing this result |
| 1.5 | disposable-1b | runtime error | harness error (obligation map key `key@revision`) |
| 1.6 | disposable-1b | PASS | run-1 final (`…-c1c8-run-final.log`) |
| 2.1 | disposable-2 | PASS (harness) with **C7-positive FAIL recorded as F-2** | run-2 final (`…-c1c8-v2-run-final.log`) |

Attempt logs 1.1–1.5 now published beside the finals. Gates: run once on the
run-1 tree (20/20 ok); on the run-2 tree, `go vet` + the scenario package ran
clean and the disposable source was removed before any commit. Zero paid
calls; zero production writes; both runs on disposable stores. The store
clocks are synthetic (run 2: strictly monotonic from a fixed base) — they
establish **in-store ordering** (predeclaration before observations), and are
not claims about wall-clock time; wall-clock provenance is this bundle's git
history.

**Coverage statement (scoped):** run 2 supports review completion for the one
obligation with a **demonstrated nonconformance**; the C8 projection exercised
`ELIGIBLE_TO_ADVANCE`-vs-`WITHHOLD` tracking of a derived outcome plus
`stale_dependency`; the remaining reason codes were not exercised and no such
claim is made.

## 5. Remediation handoffs (≤3)

1. **F-2 (`decisions`/`assessment-identity`): LANDED 2026-09-12.**
   Current-authority selection is population-compatibility-aware; acceptance
   met by the committed regression
   `TestIntegrationReplayPreservesCompatibleCurrentAuthority` (run-2 C7
   positive branch) with the 5b negative control still passing.
2. **F-1 (`evidence-admission`): LANDED 2026-09-12.** Per-evaluation admission
   visibility under exact context (migration v45: per-evaluation marker key),
   with the no-inflation companion constraint holding structurally via
   current-heads revisioning. Acceptance met by the committed regression
   `TestIntegrationLaterStrongerEvaluationReachesAdmission`: the same-proposal
   witness-after-model-judged case is rule-admissible exactly once.
3. **Attribution slice (`verification`/`transformation`):** an adapter-level
   binding from a proposal's executed bounded attempt to its emitted tuple, so
   C3 can assert mechanism-to-output attribution instead of declaring it.
   Acceptance: witness admission carries a checkable attempt→output link.

> The final disposable harness passed on a now-reconstructable snapshot (base
> `7c54dcc` + retained patch and untracked archive). It provides positive
> evidence for the exercised admission and authority-exclusion paths; an
> earlier variant reproduced evaluation shadowing in admission (F-1); and the
> re-executed positive current-authority-preservation branch **failed,
> demonstrating F-2**. C3 demonstrates supplied-witness checking and
> membership-exact admission without mechanism-to-output attribution. C8
> demonstrates deterministic projection of derived review records, including
> `WITHHOLD` on demonstrated nonconformance — not candidate eligibility. The
> full one-obligation recipe is **not validated**; preservation testing
> remains gated.
