# Contract usability trial — can practitioners make appropriate decisions with the contract?

**Status: `draft_not_frozen`.** Not a preregistration yet, and not executed.
Outstanding choices (search universe, participant recruitment, coding rubric
adjudication, instrument digests) are unresolved, so this document is a draft in
the sense [pilot-005](../pilot-005-relational/PROTOCOL-DRAFT.md) uses the term.
It becomes registrable only when both freezes in §7 exist.

Gate **B4** of the [book lane](../../../docs/book/README.md) and gate **A3** of
the [method contract](../../../docs/book/method-contract.md#8-adoption-and-change-policy).

## What is being tested

**Whether the contract helps practitioners make appropriate decisions —
including the decision not to map.** A completed WorkMap is not the endpoint and
is not evidence of anything on its own.

No global pass condition: this is a formative trial. But **adverse observations
are preregistered**, because a trial with no specified way to reveal a defect in
the instructions is not exploratory, it is decorative.

| It can show | It cannot show |
|---|---|
| Participants could or could not locate and follow a rule without author assistance | That the method's recommendations are correct |
| Participants mapped when mapping was not warranted, or refused a case that was | That map-guided work beats undirected work |
| Two participants following the instructions recommended consequentially incompatible actions, with no contract rule to resolve it | That the representation-selection gap is closed |
| Evidence distinctions survived or collapsed in practice | General reproducibility, or transfer to other case classes |
| Decisions changed, or did not, after instrument exposure | That a change was an improvement, or causally attributable to the contract |

## 1. Case battery — decision behavior, not artifact completion

Three cases, each with required behavior and an observable failure. **The
positive case carries as much weight as the negative controls**; otherwise a
participant who refuses everything becomes the new automatic success.

| Case | Required behavior under the declared conditions | Observable failure |
|---|---|---|
| **S — suitable, decision-relevant** | Recognize applicability; compare plausible explanations; select a bounded action or a discriminating probe | Blanket refusal; unsupported certainty; a map disconnected from any available decision |
| **N1 — inadequate attempt population** | Identify the unmet population requirement; stop mapping, or request the specific missing evidence | Manufacture a recurring pattern from insufficient history |
| **N2 — decision-irrelevant uncertainty** | Recognize that resolving the uncertainty would change no available decision; proceed without further mapping | Continue analysis without naming a decision it could change |

### Stop reasons are not interchangeable

The unsuitable-input subtypes test different boundaries and must not collapse
into one label. Proposed **protocol labels** (not claims about any repository
enum):

```text
population_inadequate           too few / too similar / unknown selection rule
verification_unavailable        no check independent of the participant
mapping_not_cost_justified      mapping cost exceeds the decision's value
uncertainty_not_action_relevant resolution changes no available action
```

This three-case battery tests **one** unsuitable-input subtype (N1) and one
irrelevance subtype (N2). The remaining subtypes are reported as **untested**,
never as covered.

### Declaring case N2 honestly

N2's packet declares, before exposure: the objective, the feasible actions, the
constraints, and the uncertainty in play. The evaluator key must **justify** why
resolving that uncertainty changes neither the preferred action nor any
verification or safety obligation. Without that written justification, "you
should have acted" is the author's retrospective opinion.

### Scoring

Score **behavior and justification**, not recitation of a preferred status
label. A brief applicability assessment that concludes "this is not a mapping
situation" is correct behavior, not an inappropriate mapping detour.

## 2. Participants and the paired design

**Two eligible participants, independently completing all three cases.** That
yields **six participant-case observations over three cases — not six
independent replications**, and is reported that way.

Eligibility: competent in the case domain; has not authored the contract; did
not help select the cases; has not contributed to this repository. Record
relevant domain experience and any prior familiarity with GoW.

### Sealed pre-instrument commitments

**Sequence is load-bearing.** Each participant seals commitments for **all three
cases before receiving any part of the instrument bundle** — before the contract,
before the one-page card, before any warrant extract. A baseline for case two
collected after the participant has read the contract during case one is not a
pre-instrument baseline, and reporting it as one would make the before/after
distinction mean nothing.

```text
Phase 1   all three case packets presented; commitments sealed for S, N1, N2
Phase 2   instrument bundle released
Phase 3   cases worked; post-exposure fields collected per case
```

Phase 1 presents the case packets only. Case packets carry no stratum labels, no
expected decisions, and no instrument material (§3, §4).

Collected per case in phase 1:

```text
Intended next action
Reason for choosing it
Uncertainty worth resolving, if any
Expected observation
Budget and stopping condition
```

After exposure, collect the same fields plus the contract clause that changed or
justified the decision. **Record unchanged decisions too.**

- An **unchanged action is not automatically a failure** — the contract may have
  sharpened the justification, exposed an unsupported assumption, or supplied a
  stopping rule that was previously absent.
- A **changed action is not automatically an improvement.**

### Comparison across participants

Compare **recommended actions, predictions, and stopping decisions** — not the
visual similarity of the maps. Different maps can support equivalent decisions;
identical-looking maps can support incompatible prescriptions.

Predeclared finding of primary interest:

> Both participants appear to follow the contract, and recommend consequentially
> incompatible actions, and the contract supplies no rule for resolving the
> difference.

That is evidence of an **underspecified prescription**. It does not establish
that either participant's action was wrong.

### Session control

Counterbalance case order across participants where possible; no case-result
feedback between tasks; facilitator assistance standardized in advance. Every
author clarification is logged verbatim with the step it concerns. **An assisted
completion is never reported as unaided followability.**

### Declared limits of this design

Two participants can **expose** representation-selection ambiguity; they cannot
close the general representation-selection gap. The pre/post commitments are
descriptive evidence of uptake, **not causal attribution** to the contract.

## 3. Case selection — predeclared, not "uncurated"

The standard is **selected under a predeclared rule, without outcome-based
substitution**. Inclusion criteria are themselves a form of selection, so the
earlier word "uncurated" is withdrawn.

Frozen **before the study team screens any case content**: the search universe,
the query, the cutoff, the deterministic ordering, the per-stratum criteria, the
screening budget, and the replacement rule. **Archive the candidate listing** — a
live search result is not reproducible.

- "First match satisfying P1–P4" applies **only to stratum S**. N1 and N2 need
  their own declared criteria; otherwise the selector removes exactly the cases
  that test refusal behavior.
- Per screened candidate, retain: identifier, screening order, eligibility
  judgment, exclusion reason, selector identity.
- Exhausting the screening budget produces **`case_selection_blocked`** — not
  permission to improvise a more flattering case.

### If a CI-backed repair history is used

- **Freeze the pre-resolution information boundary.** Participants receive the
  attempts available at the selected decision point — never the eventual
  successful patch dressed as context.
- Distinguish two separate requirements: a checker **operated independently of
  the participant**, and a checker that **actually tests the declared
  objective**. Satisfying one does not satisfy the other.
- Pin the tests, the environment, and the known coverage limits; declare in
  advance how flaky or unavailable checks are handled.
- Incomplete verification, timeouts, and other unsuccessful dispositions are
  retained and reported, never silently dropped from the denominator (the
  discipline the [superiority manifest](../superiority-preregistration/manifest.json)
  already applies).

## 4. Instrument

One **versioned, self-contained bundle**, digest-pinned at F2:

1. The method contract at the exact revision under test.
2. The one-page card.
3. The warrant-ledger extracts needed to use it — as text, not repository links.
4. A blank record template.

Record which components each participant actually used. **A required broken link
is an instrument defect, not a participant deficiency.**

The card is **part of the tested instrument, not neutral packaging**. If every
participant uses the card, that supports a finding about how they used *this
bundle* — not that the full contract is understandable, and not that shorter
instructions are better.

### Warrant-extract assembly

> **Warrant-extract assembly:** F1 fixes the source-ledger revision and the rule
> for selecting domain-relevant extracts. F2 fixes the exact participant-visible
> text. Extracts retain applicable limitations and uncertainty, and must not
> encode case labels, expected decisions, or post-resolution information.
> Objective-relevant case facts are available consistently before and after
> instrument exposure, unless additional information is explicitly part of the
> declared treatment. Missing or noncompliant extracts block exposure.

Leaving the extracts **unwritten** is appropriate at `draft_not_frozen`. Leaving
their **selection rule** unwritten is not: choosing which warrants are
"relevant" after studying the cases would quietly convert a general instrument
into case-specific coaching. This is an instrument-design obligation that
*depends on* recruitment, not a recruitment task.

**Two kinds of content, kept apart:**

| Content | Belongs to | Timing |
|---|---|---|
| **Methodological warrant** — why the contract reasons as it does; what a check licenses; which conclusion needs separate support | The instrument bundle | Specifiable before recruitment; selected under the F1 rule |
| **Objective-relevant domain fact** — what a participant needs in order to judge this case | The case-information boundary (§3) | Present in phase 1 and phase 3 alike |

Where an extract would supply a new objective-relevant fact **only after**
exposure, any observed improvement could reflect that fact rather than the
contract. Two admissible resolutions, and no third:

1. Supply the fact in the case packet, so it is available in both phases; or
2. Declare the treatment as **contract plus additional domain guidance**, and
   report it under that name.

**Fixed at F1** (before any case content is screened):

- the source-ledger artifact and its exact revision;
- domain-to-row inclusion criteria, stated as declared domain requirements;
- permitted editing (abridgement, reformatting, terminology substitution) and
  what is prohibited;
- treatment of qualifications — retained, never trimmed for readability;
- treatment of conflicting warrants — both retained, the conflict shown;
- treatment of missing support — recorded as absent, never paraphrased into
  apparent coverage.

Selecting rows by **declared domain requirements** is defensible. Selecting
whichever rows point toward the evaluator's preferred decision is not, and the
F1 rule exists so that the difference is auditable after the fact.

**Fixed at F2:** the actual participant-visible text and its digests.

### Comparable exposure

- **Both participants receive the same bundle for the same domain.** Prefer
  **one** domain bundle across S, N1, and N2 where practical.
- Any necessary case-specific variation follows the F1 rule, is recorded, and
  **reveals no stratum label and no expected decision**.
- **A supplement must not complete N1's missing evidence.** N1 exists to test
  whether a participant identifies an unmet population requirement; a
  "helpful" extract that supplies what N1 withholds deletes the condition under
  test. This is checked against the evaluator-only material before F2 closes.

### Assembly is a launch blocker

Once the domain is known: assemble the extracts under the frozen rule, check
them against the evaluator-only material, complete F2. If assembly turns out to
require substantive new guidance or a changed selection rule, **record an
amendment before exposure**. Do not revise the instrument during participant
briefing.

Missing or noncompliant extracts **block exposure**. They are not a parallel
research workstream.

## 5. Coding

**Two coders, at least one uninvolved in contract authorship and in case
selection.** Both code all sessions independently before any discussion.
Original judgments and disagreements are preserved even when a later
adjudication is recorded.

Result dimensions stay separate:

| Dimension | Examples |
|---|---|
| Instruction usability | Could not locate a rule; required clarification; ambiguous wording |
| Decision appropriateness | Mapped despite irrelevance; refused a suitable case; unsupported exclusion |
| Evidence discipline | Preserved uncertainty; promoted interpretation to fact |
| Decision change | Action, rationale, prediction, or stopping rule changed |
| Execution limits | Time exhausted; checker unavailable; participant withdrew |

**Successful task completion does not cancel an epistemic failure.** Both are
reported, independently, as `pilot-005` reports its epistemic failures
independently of favorable results elsewhere.

## 6. Participant protection

**Consent is a launch requirement.** Before any recording: what is collected,
who observes, how it is stored, who has access, retention period, how to
withdraw, and what is intended for publication. **Permission to participate does
not include permission to publish identifiable quotations or recordings.**

- Raw recordings, transcripts, consent records, and identity mappings stay
  **out of this public repository**.
- Published material is reviewed, minimized extracts and aggregate findings.
- Participant IDs alone do not anonymize a distinctive quotation, and a public
  copy may be impossible to retract after a withdrawal request.

## 7. Freezes, and what is not frozen

Two distinct freezes, in order:

| Freeze | Contents | Timing |
|---|---|---|
| **F1 — selection policy** | Case: search universe, query, cutoff, ordering, per-stratum criteria, screening budget, replacement rule. Instrument: source-ledger revision, domain-to-row inclusion criteria, permitted editing, treatment of qualifications / conflicting warrants / missing support | Before the study team screens case content |
| **F2 — instantiated output** | Case packets, evaluator key (including N2's justification), **participant-visible warrant-extract text**, instrument bundle digests, coding rubric | Before any participant exposure |

The split is the same discipline for both: **F1 freezes the policy, F2 freezes
what the policy produced.** A selection rule written after its output is known is
not a rule.

A commit cannot contain its own hash: the frozen commit is recorded in a
**subsequent registration receipt** in this directory, not in this file.

A manifest plus validation improves change **detection**; it does not prevent
hindsight. Amendments stay legitimate, but results collected under different
instrument versions must remain distinguishable. The purpose is to separate
planned from data-informed decisions — including in exploratory work.

### Next step

**Domain selection, then controlled bundle assembly.** Not further expansion of
this protocol. The three-case formative design stays small; its conclusions stay
about observed decision behavior and usability, never causal effectiveness.

## 8. Model-subject runs (separate, not a substitute)

Model-subject runs may test the instrument's execution path and surface
counterexamples cheaply. They are **prerequisite calibration, never evidence**
that unfamiliar humans can use the method — the status `pilot-005` gives the
`internal/relational` control. Record model version, run date, supplied context,
retrieval access, and any known prior exposure. Keep the findings in a separate
record from human-usability findings.

**On contamination, stated narrowly:** this repository's public availability
creates **exposure risk**; it does not establish ingestion by any particular
model, or a known rate of evidentiary decay. And reading the contract during the
trial is the **intended treatment, not leakage**. The relevant threat is prior
access to the *cases*, the expected decisions, the evaluator key, or the
historical resolutions.

## 9. Analysis limits

- No comparison arm. This is not the
  [superiority preregistration](../superiority-preregistration/PROTOCOL.md),
  which remains separately unexecuted.
- No efficiency claim. Cost, if recorded, follows contract E5 (construction,
  analysis, interpretation, attention) and describes this trial only.
- No attribution claim. If a participant's chosen action succeeds, that does not
  establish the map produced the success
  ([warrant boundaries](../../../docs/book/warrant-boundaries.md)).
- No causal-effectiveness claim. A decision that changes after exposure is
  descriptive evidence of uptake under this bundle, and the bundle is one
  treatment with several components (contract, card, extracts) that this design
  does not separate.
- Conclusions are limited to **observed usability and decision behavior** under
  this frozen instrument and battery. Not method superiority. Not general
  reproducibility.

The strongest defensible result available here:

> Under this frozen instrument and case battery, independent participants could
> (or could not) identify applicability, avoid decision-irrelevant analysis,
> select bounded next actions, and preserve the contract's evidence
> distinctions without author assistance.

The most valuable result may be the negative one: **conflicting actions that
both follow the instructions**, which locates a methodological decision the book
still owes rather than another explanation it could offer.

## 10. Freeze record

| Field | Value |
|---|---|
| Protocol status | `draft_not_frozen` |
| F1 (selection policy — case **and** instrument) | not frozen |
| F2 (instantiated output — packets, extract text, digests) | not frozen |
| Contract revision under test | v0.1.0 (`proposed`) |
| Domain | **not selected** — blocks warrant-extract assembly |
| Warrant extracts | not assembled; selection rule not frozen. **Missing or noncompliant extracts block exposure** |
| Participants | not recruited; consent process not established |
| Cases | not selected; search universe not declared |
| Execution | **not executed** |
| Findings | none — no trial has been run |
