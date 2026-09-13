# E0 custodian handoff: public interface for protected task authoring (revision 2)

- **Status:** prepared handoff; the custodian role's task-authoring half is unstaffed and nothing here authorizes execution
- **Revision 2 (2026-09-13):** remediates the D2 model review lane's findings (3 P1, 7 P2, 5 P3; recorded in [`015`](2026-09-12-015-custody-staffing-authorization-memo.md)). Review label: model-judgment defect-finder, same-model-family; not the external human review E0's exit requires
- **Roadmap:** [`v0.3.0`](../research/gow-shaping-maturity-roadmap-v0.3.0.md) §2 (E0), §4 (G1 gate, G4-lite) — normative for strata and arm definitions
- **Deliberate exclusions:** no worked examples, development-case content, known-failure identities, prompt text, or rule-specific test cases. The review lane confirmed this boundary holds

## 1. What the custodian owns

Authoring, sealing, and access control for protected packs; validation of histories and per-episode optimization targets; scoring custody for the G4-lite screen; **pre-sealing sensitivity calibration** (§7). The custodian sees this interface and the contracts below, not the implementation's rule-specific examples, prompts, failure logs, or development cases. An external domain review of task assumptions is required before any external-evidence claim.

## 2. Case format and claim-kind composition (P1-3, P2-7)

**Format:** protected cases are Go struct literals against the public APIs below, in a custodian-held package outside this repository until execution. No parser exists and none is implied.

**Claim-kind composition of the 48-case G1 pack:** the pack spans the registry, not only the finite kind — roadmap G1's capability includes observed-value invariance and monotonic accumulation. Proposed split, custodian-adjustable with the split recorded: 24 finite-kind (equivalence + instance), 16 measure-kind (rate invariance + monotonicity), 8 probabilistic/out-of-scope routing cases. A finite-only pack would leave two of the three seed tools untested and must be recorded as a deliberate narrowing if chosen.

**Measure-kind input interface** (previously omitted): a case supplies a `measure.Binding` (exact sentence, metric numerator/denominator, population, `BudgetRange` — range-scoped or unrestricted — ordering, stopping rule, claim kind) and `measure.Observation`s (declared `Conditions` + per-instance ordered `Traces` of submissions). Verdicts: `REFUTED`, `HOLDS_AT_COMPARED_POINTS`, `UNRESOLVED`, `INAPPLICABLE`, `NOT_ASSESSED`.

## 3. The finite language (unchanged; matches `internal/finite`)

Fixed-width words (1–8 bits); domain = width + closed variable set; total operators mod 2^width: `not, neg, shl1, shr1, and, or, xor, add, sub, mul`; constants reduced mod 2^width; **no division**. G4-lite domain: width 4, ≤3 variables (16³ = 4096 ≤ the cap).

## 4. Oracle outcomes, by entry point (P2-4/5/6, P3-11)

| Outcome | `AssessEquivalence` | `AssessInstances` |
|---|---|---|
| `HOLDS_ON_DECLARED_DOMAIN` | exhaustive agreement; supports exactly the declared domain | never (instance agreement is `INSTANCE_EVIDENCE_ONLY`) |
| `REFUTED` | first counterexample in **canonical order** (vars sorted, values ascending) | first disagreeing instance in **caller-supplied order** — and it DOES refute the universal domain claim |
| `UNRESOLVED` | domain **strictly greater than** 65,536 assignments (width-8 two-variable = exactly 65,536 still enumerates; oversized needs ≥3 variables at width ≥7) | **also**: empty instance set ("nothing was checked") |
| `INAPPLICABLE` | full refusal surface: undeclared variable, unknown operator, **nil/malformed nodes, empty variable names, depth > 64** | additionally: an instance missing a declared variable or exceeding the width |
| `INSTANCE_EVIDENCE_ONLY` | — | agreeing instances; can never be a domain verdict |

**Expressible refusal triggers for inapplicable cases** (struct literals can author these): unknown `Op` strings, nil children, empty variable names, undeclared variables, out-of-range widths, over-deep nesting, malformed instances. **Not externally authorable:** foreign node types (the `Expr` interface is closed to this package) — that trigger is internal-only and no case should target it.

Warrant consumption (`finite.VerifyRuleWarrant`) and the `toolreg` registry are as in revision 1. **Out-of-registry cases** (required stratum) expect one of two named outcomes: a `toolreg.Select` refusal error naming the kind, or `measure.VerdictNotAssessed` for probabilistic claims.

## 5. Underspecified stratum: operational definition (P1-1)

"Underspecified" means **a missing premise the case never declares**, and the expected outcome must *identify the missing premise* — it is NOT the enumeration cap (an oversized domain is fully specified; its `UNRESOLVED` names no missing premise and does not satisfy this stratum). Authorable forms, per kind:

- **measure kinds:** an invariance claim with no declared budget range where range membership decides applicability; observations whose nesting is undeclared/unverifiable (→ `INAPPLICABLE`/`UNRESOLVED` with the premise named in the reason); a claim sentence whose metric or population binding is absent.
- **finite kinds:** an instance case whose assignments omit a declared variable (the reason names the unassigned variable); a claim sentence that does not fix the domain the expressions quantify over, forcing the binding to carry the omission.

Acceptance for the 8 underspecified cases: the recorded reason must name the missing premise; a cap-refusal or a failed-premise refusal does not count.

## 6. G4-lite optimization targets: checkable interface (P1-2)

The admissible cost vocabulary is the runner's: `rewrite.CostModel`, default `rewrite.NodeCount` (one unit per node); per-operator weighted variants are admissible if declared as an exact integer weight table. Ties break deterministically by canonical rendering (the runner's rule). A target is `(cost model, integer threshold)`; **completion** = `BestCost ≤ threshold` **and** an oracle-verified endpoint (`EndpointVerified`) — an unverified candidate completes nothing. The evaluator is `rewrite.Search` + `screen.Evaluate`; no separate cost checker is outstanding.

## 7. Scoring custody: operative arithmetic and calibration (P2-9, P2-10)

- **Operative rule:** `internal/screen`'s exact-integer sum comparison over a complete arm×episode×run grid (condition (b): `HG_sum ≥ H1_sum + 3·r`) — this supersedes the roadmap's rounded-average phrasing, per the operator's recorded correction. Runs-per-cell (`r`) is fixed before sealing (default 3 with stochastic providers; 1 only for deterministic procedures, disclosed).
- **Sensitivity calibration (pre-sealing, custodian-owned, development cases only):** verify the rule pool can express improvements and a reference procedure (the bounded rewriter) reaches them under the proposed budget; avoid floors/ceilings where no arm or every arm completes. Preserve the calibration policy, not only its successful cases. Calibration is the residual registry-coupling channel — it stays labeled on the pack, and is never permission to tune a protected task until an arm wins.

## 8. Pack shapes and per-case record (P2-8, P3-12/13)

G1 pack: 48 = 24 applicable / 16 inapplicable / 8 underspecified. G4-lite pack: 24 episodes = 12 informative / 6 low-value / 6 misleading (strata defined in roadmap §4, normative). **Every episode of every stratum carries a construction-family ID** — `screen.Evaluate` refuses empty families. Condition (d)'s family-diversity guard counts **informative episodes only** (control-stratum differences cannot satisfy it; superseded revision-2's earlier all-strata description after the tranche adversarial review, finding 2). ≥2 families among informative episodes or (d) is unsatisfiable.

Each case records (full roadmap E0 list): exact objective; **admissible semantics**; source lineage; expected answer or oracle outcome — **plus, for instance-level cases, the assignment set**; **permitted references**; history construction; **author/model provenance**; **exposure history**; family ID.

## 9. Access boundary

| Party | May see before sealing | Must not see |
|---|---|---|
| Custodian | this document; public package docs of `internal/finite`, `internal/toolreg`, `internal/measure`, **`internal/screen`**, `internal/rewrite` (cost models); roadmap §4 stratum/arm definitions | implementation test files, development cases, prompts, failure logs, review records naming specific identities |
| Implementer | pack sizes and strata (public in the roadmap) | protected case content, answer manifests, aggregate scores during adaptation |

Seal source and answer manifests separately; test access controls; scan adjudication inputs for arm identifiers; retain raw outputs. Score leakage to the implementer is feedback; reuse after adaptation is development.

## 10. Outstanding (ownership, not engineering)

1. Human task-authoring custodian (D1: operator holds mechanics; unexposed human before sealed claims).
2. External domain review before external-evidence claims (D2: model lane executed — this revision is its remediation; human reviewer still required).
3. D3b execution authorization; arm parity review of H1 (custodian-held, per D4).
