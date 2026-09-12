---
id: research-review-ownership
revision: 1
kind: ownership_map
---

# Ownership map: paths first, responsibilities second

This is an anti-omission map, not nineteen prompts or an execution schedule.
Names are stable slugs; old numbers below are migration aliases only. Change the
partition when it improves accountability without leaving obligations ownerless.
The [contract](review-contract.md) governs applicability, evidence and decisions.

## Operator paths

| Path ID | Causal path and decision | Responsibility owners | Entry point |
| --- | --- | --- | --- |
| `assessment-admission-decision` | Evidence changes, reassessment, current authority and next action | `assessment-identity`, `evidence-admission`, `decisions` | [One-obligation integration recipe](recipes/assessment-admission-decision.md); existing [identity prompt](assessment-identity-and-derived-views.md) supplies concrete probes. |
| `realization-evidence` | Proposed change, concrete artifact, scoped check, admission | `transformation`, `verification`, `evidence-admission`; identity and decisions as peers | Existing [projection/admission prompt](typed-projection-and-evidence-admission.md). |
| `source-inference-report` | Source/result, warranted inference, outward claim | `warrants`, `inference`, `reporting`; comparative validation when utility is claimed | Historical manuscript review supplies cases; the contract's escalation rule applies. A reusable new prompt is deferred, not represented as implemented. |
| `review-preservation` | Equivalent information, defect/corrected controls, cost of packaging | `study-design`, `comparative-validation`, `decisions`, `reporting` | [Qualitative preservation protocol](recipes/preservation-pilot.md), gated on the integrated slice. |

These paths are initial entry points, not a complete list of research workflows.
Select additional obligations for the named decision without inventing one
session per owner. Composition is a recipe duty, not a twentieth domain verdict.

## Responsibility library

A finding has one primary owner; a recipe may cover several. Boundaries below
are assignment rules. Resolve collisions using independently remediable causes
and linked manifestations, not the file where an error was discovered.

| ID (old alias) | Owns | Boundary / assignment example | Peers and applicability trigger |
| --- | --- | --- | --- |
| `question` (01) | Research question, decision scope, success/failure conditions and exclusions | Defines what success would mean; does not claim it occurred | `warrants`, `comparative-validation`; every named research or release decision |
| `positioning` (02) | Prior-work search, strongest alternatives and contribution rationale | Search adequacy/novelty reasoning, not fidelity of a final sentence to its source | `sources`, `reporting`; novelty or contribution claim |
| `semantics` (03) | Constructs, units, meanings, quantifiers and equivalence relations | Unknown is not false; record identity implementation belongs to identity | `measurement`, `warrants`, `assessment-identity`; any typed or interpreted construct |
| `warrants` (04) | Claim specification, assumptions, dependencies and required justification | Specifies what would justify the claim; verification records what a particular check establishes | `verification`, `inference`, `evidence-admission`; any consequential claim |
| `sources` (05) | Acquisition, selection, authenticity, sampling and origin provenance | Selection/derivative-source contamination belongs here; long-term access belongs to stewardship | `positioning`, `measurement`, `stewardship`; evidence acquisition or reuse |
| `measurement` (06) | Instruments, operationalization, annotations and calibration | A model's labels are measurements/interpretations, not independent observations | `semantics`, `inference`; values or labels measure a construct |
| `assessment-identity` (07) | Artifact/revision/occurrence/context identity, history and current views | Attribution and context selection, not the admission criterion or downstream choice | `evidence-admission`, `decisions`; reassessment, revisions or cached results |
| `evidence-admission` (08) | Eligibility of scoped evidence, status changes, contradiction and withdrawal | Decides whether a checked result supports this claim/population; does not strengthen its checker | `verification`, `assessment-identity`, `decisions`; evidence can affect belief/action |
| `transformation` (09) | Reduction, abstraction, transfer, construction and grounded projection | Preservation and realization, not fluent descriptions or a changed problem | `semantics`, `verification`; source-to-target transformation |
| `study-design` (10) | Diagnostic contrast, controls, identification and advance commitments | A disclosed limitation is not automatically a finding; assess it against the actual claim | `measurement`, `inference`, `comparative-validation`; experiment or planned discrimination |
| `implementation` (11) | Protocol realization, runtime reliability, numerical and interface behavior | An operational timeout is not a domain counterexample | `verification`, `governance`, `stewardship`; apparatus, code or executed procedure |
| `verification` (12) | Exact check subject, applicability, decisiveness and adjudication | Records what was checked under which assumptions; a strong checker of the wrong subject is insufficient | `warrants`, `transformation`, `evidence-admission`; claimed check or adjudication |
| `inference` (13) | Result-to-conclusion reasoning, uncertainty and sensitivity | Invalid source argument stays an inference finding even when discovered through a manuscript | `warrants`, `study-design`, `reporting`; conclusions drawn from results |
| `comparative-validation` (14) | Claimed transfer, prospective performance and credible comparisons | Internal inference validity is separate; a correct procedure need not improve outcomes | `study-design`, `inference`, `decisions`; efficacy/generalization/utility claim |
| `reproduction` (15) | Independent reconstruction, repeatability, replication and tolerances | Exercises reconstruction; artifact availability alone belongs to stewardship | `implementation`, `verification`, `stewardship`; reconstructability or independent-validation claim |
| `decisions` (16) | Evidence-to-action rules, adaptation, total costs and stopping | Uses eligible current assessments; does not create evidence or silently change requirements | `assessment-identity`, `evidence-admission`, `comparative-validation`; allocation/release/next-action decision |
| `governance` (17) | Authority, permissions, harms, incentives, conflicts and exceptions | Technical access is not permission; generic review is not qualified approval | `implementation`, `stewardship`, `reporting`; universal screening, expanded for material risk |
| `stewardship` (18) | Custody, permitted access, metadata, reuse, preservation and maintenance | Sustains artifacts and custody records; acquisition selection belongs to sources | `sources`, `reproduction`, `governance`; shared or retained research artifacts |
| `reporting` (19) | Outward claim/source fidelity, completeness, labels and attribution | Faithful repetition of invalid inference escalates to inference with a linked reporting manifestation | `positioning`, `inference`, `governance`; manuscript, report, figure, documentation or summary |

## Reused lenses

Structural: objects, dependencies and boundaries. Semantic: retained meaning.
Epistemic: warranted conclusion and uncertainty. Operational: executable and
recoverable procedure within budget. Responsibility: permissions, risks and
accountability. Screen all five; expand only applicable questions. They are
analytical facets, not mutually independent dimensions or five separate reports.

## Method triggers, without a profile directory

| Trigger | Additional obligations and required local cases |
| --- | --- |
| Formal/mathematical claims | Quantifiers, assumptions, proof composition, exact witnesses and trusted-checker limits; no statistical test in place of proof. |
| Computational/simulation work | Numerical error, modeled-domain validity, environment/dependency versions and nondeterminism. |
| Experimental empirical work | Allocation, control fidelity, blinding where meaningful, predeclared endpoints and protocol deviations. |
| Observational/causal work | Confounding, selection, identification assumptions, estimand and transport limits. |
| Qualitative/historical/interpretive work | Source context, coding trace, rival readings and reflexivity; do not demand literal repetition of an event. |
| Evidence synthesis | Search/screening coverage, duplicate studies, quality, heterogeneity and publication bias; activate appropriate reporting guidance. |
| Agentic/adaptive work | The mandatory-risk overlays below apply for GoW; they are not optional low-priority modifiers. |
| Sensitive people/data/organisms/infrastructure | Qualified oversight, permission, containment, harm, retention and release restrictions. |

### Interim home for agentic obligations

| Risk | Primary owner / peers | Concrete case and execution home |
| --- | --- | --- |
| Model as instrument | `measurement` / `verification`, `inference` | Change annotator/prompt/model while retaining the artifact; do not transfer calibration or call repeat agreement independent. Both existing prompts' agentic overlays apply. |
| Generated versus observed inputs | `evidence-admission` / `sources`, `verification` | Withhold a model-only failure; retain scope when a genuinely checked observation enters. Integrated recipe cases C2-C3 and projection prompt. |
| Memory/context and policy drift | `assessment-identity` / `decisions` | Unchanged content, changed evidence; replay must not restore obsolete authority. Integrated recipe C4-C7 and identity prompt. |
| Holdout reuse or post-outcome adaptation | `study-design` / `comparative-validation`, `decisions` | A used holdout becomes development evidence; freeze fresh cases before transfer claims. Preservation protocol and both prompts' overlays. |
| Tool authority and prompt injection | `governance` / `implementation`, `evidence-admission` | A source or provider payload requests forged status, a write, or unapproved tool use; it has no authority to authorize either. Contract restrictions and projection prompt overlay. |

A YAML adapter cannot replace coupled examples such as supersession-before-filtering
or a concrete exact-integer witness. Keep those cases in project-specific prompts
and fixtures until their preservation is tested.

## Mandatory seam challenges, selected for scope

| Change or challenge | What must remain distinct | Responsibility peers |
| --- | --- | --- |
| Definition or measurement rubric changes | Original result versus newly interpreted result | `semantics`, `measurement`, `assessment-identity`, `inference`, `reporting` |
| Source withdrawn/excluded | Preserved history versus current support | `sources`, `evidence-admission`, `assessment-identity`, `decisions`, `reporting` |
| Same candidate; new evidence/policy | Artifact equality versus assessment equivalence | `assessment-identity`, `evidence-admission`, `verification`, `decisions` |
| Description passes without realization | Well-formed annotation versus domain result | `transformation`, `implementation`, `verification`, `evidence-admission` |
| Analysis changes after outcomes | Exploration versus previously committed confirmation | `study-design`, `inference`, `comparative-validation`, `reporting` |
| Mistaken procedure reruns successfully | Repeat execution versus valid inference | `implementation`, `verification`, `inference`, `reproduction` |
| Accessible but unauthorized artifact | Technical access versus permission | `governance`, `stewardship`, `implementation` |
| Revised policy chooses the next action | Evidence available then versus future information | `assessment-identity`, `evidence-admission`, `comparative-validation`, `decisions` |

This table is a test plan, not results or a coverage denominator. One recipe
traces the selected seam; peers do not rerun it as separate domain sessions.

## Maintenance and failure conditions

After ten completed, applicable review opportunities with no distinct use of a
row, review it for consolidation. Ten is an initial maintenance interval, not a
validated law. Merge/retire only if obligations retain an explicit owner or no
longer apply to supported scope. Keep old ID aliases. Never delete rare safety
or release obligations merely because no defect was found.

The package fails its purpose when ownership remains ambiguous, a consequential
handoff has no test, statuses become independently edited duplicates, known
controls are falsely reported defective, or modular packaging loses actionable
specificity without an acceptable quality/cost justification. The integrated
slice tests derivability; the preservation trial tests packaging. Neither
establishes universal taxonomy completeness or scientific efficacy.
