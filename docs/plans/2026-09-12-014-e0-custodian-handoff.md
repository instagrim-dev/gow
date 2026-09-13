# E0 custodian handoff: public interface for protected task authoring

- **Status:** prepared handoff; the custodian role is unstaffed and nothing here authorizes execution
- **Roadmap:** `gow_shaping_maturity_roadmap_v0.3.0.md` §2 (E0), §4 (G1 gate, G4-lite)
- **Audience:** the future evaluation custodian — a person or process distinct from the shaping implementer
- **Deliberate exclusions:** this document contains the supported public interface only. It contains no worked examples, no development-case content, no known-failure identities, no prompt or template text, and no rule-specific test cases — supplying those would burn the custodian's independence before it existed. The implementer must not see the protected cases this document enables before sealing.

## 1. What the custodian owns

Authoring, sealing, and access control for the protected task packs; validation of histories and per-task optimization targets; scoring custody for the G4-lite screen. The custodian sees this interface and the oracle contract below. The custodian does not inspect the implementation's rule-specific examples, prompts, failure logs, or development cases before authoring (roadmap E0). An external domain review of task assumptions is required before any external-evidence claim; absent that review, packs are development evidence.

## 2. Public task interface: the finite expression language

A task's semantic substrate is the seed language of `internal/finite` (public contract; the package documentation is the normative statement):

- **Values:** fixed-width machine words, width 1–8 bits. The G4-lite screen's proposed domain is width 4 with at most three variables (≤ 16³ = 4096 assignments).
- **Domain declaration:** a width plus a closed set of named variables. Every claim quantifies over exactly this declaration.
- **Operators (all total, all modulo 2^width):** unary `not`, `neg`, `shl1`, `shr1`; binary `and`, `or`, `xor`, `add`, `sub`, `mul`. Division does not exist in the language.
- **Constants:** literals reduced modulo 2^width.
- **Out of scope by construction:** floating point, unbounded integers, shift-by-variable, division, undefined behavior of any kind.

A protected case supplies: an exact claim sentence; a domain declaration; expressions in the language above; the expected oracle outcome (or the assignment set for instance-level cases); source lineage and construction-family ID; and, for G4-lite episodes, the task history and a pre-declared optimization target with its cost model and tie rule.

## 3. Oracle interface

The oracle is `finite.AssessEquivalence` (exhaustive) and `finite.AssessInstances` (instance-level), procedure revision `finite-equivalence-checker/1`. Its complete verdict vocabulary:

| Verdict | Meaning | Custodian-relevant property |
|---|---|---|
| `HOLDS_ON_DECLARED_DOMAIN` | every assignment enumerated and agreeing | supports exactly the declared domain; scope guards attached |
| `REFUTED` | exact counterexample recorded | deterministic: first counterexample in canonical order (variables sorted, values ascending) |
| `INSTANCE_EVIDENCE_ONLY` | supplied assignments agree | can never be a domain verdict; a case testing instance-vs-domain confusion should expect this |
| `INAPPLICABLE` | a premise failed (undeclared variable, unknown operator, invalid width) | the failed premise is named; useful for the inapplicable stratum |
| `UNRESOLVED` | domain exceeds the exhaustiveness cap (65,536 assignments) | refusal, not sampling; useful for underspecified/oversized strata |

Warrant consumption (`finite.VerifyRuleWarrant`) additionally rejects certificates presented for a different rule, width, domain, or with inconsistent coverage, and replays the enumeration. Custodian cases may target this boundary.

The claim-selection layer is `internal/toolreg`: five registered claim kinds (`observed_rate_invariance`, `solved_monotonicity`, `probabilistic_property`, `finite_equivalence`, `finite_instance`), each with declared premises, quantifier scope, outputs, and limits. Unregistered kinds are refused by name. Cases outside the registry's supported scope are a required stratum (roadmap E0).

## 4. Case-count proposal (roadmap revision 0.3.0, unchanged)

| Pack | Size | Composition |
|---|---|---|
| G1 protected pack | 48 | 24 applicable, 16 inapplicable, 8 underspecified; strata thresholds per roadmap §4 G1 |
| G4-lite protected pack | 24 episodes | 12 history-informative, 6 low-value, 6 misleading; ≥2 construction families among informative episodes (the spending rule's condition (d) needs family diversity to be satisfiable) |

Group by construction/template/source family; near-duplicates stay on one side of a split. Accept more than one valid tool or proof route; the oracle judges the claim and scope, not a reference name. Include: tasks outside the registry's supported scope; conditions invalidating an otherwise-familiar rule; alternative valid routes.

## 5. Access boundary

| Party | May see before sealing | Must not see before sealing |
|---|---|---|
| Custodian | this document; the public package documentation of `internal/finite`, `internal/toolreg`, `internal/measure`; the oracle verdict vocabulary | implementation test files, development cases, prompts, failure logs, review records naming specific identities |
| Implementer | the pack sizes and strata (already public in the roadmap) | any protected case content, answer manifests, or aggregate scores during adaptation |

Seal source and answer manifests separately; test access controls before execution; scan adjudication inputs for arm identifiers; retain raw outputs. An aggregate score leaked to the implementer is feedback; reuse after adaptation is development (roadmap E0, [S11]).

## 6. What remains outstanding (ownership decisions, not engineering)

1. **Staffing the custodian** — required before any pack is authored; self-authoring by the implementer cannot be relabeled independent.
2. **External domain review** of protected-task assumptions — required before external-evidence claims.
3. **Execution authorization and resource ceilings** — required before any G4-lite arm runs; the screen's decision arithmetic (`internal/screen`) is built and synthetically tested, and refuses gate eligibility for anything not labeled custodian-sealed.
4. **Arm scaffolding under parity discipline** — H1's scaffolding needs non-implementer authorship or review (roadmap 0.3.0 arm parity).

An implementer-authored dry run can expose broken plumbing; it cannot substitute for protected evaluation or count toward the spending rule.
