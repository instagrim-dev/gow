---
artifact_kind: method-contract
status: proposed
revision: 0.1.0
opened: 2026-09-12
basis: remote main fdf7f5e (2026-09-12 14:20 PDT)
authority: explicitly heuristic decision policy adopted under owner direction
scope: what a practitioner is instructed to do, and the ceiling on how strongly
  the book may instruct it
---

# Method contract — Map the Work, v0.1.0

*The compact contract the [book lane](README.md) writes against. It fixes the
intended practitioner, the mapping trigger, the representation-selection
procedure, the action-selection rule, the evidence requirements, and the
stopping conditions — so that prescriptive chapters cannot silently settle
those choices in prose.*

**Status: `proposed`, not adopted.** No chapter may cite this contract as
settled doctrine until the [adoption gate](#8-adoption-and-change-policy) is
recorded. Nothing here is a validated result.

---

## 0. What this contract is, and is not

It is a **versioned, mostly heuristic decision policy** for a practitioner with
a stalled body of work. Its purpose is usability and accountability: a reader
should be able to execute it, see where it is guessing, and notice when it
fails them.

It is **not**:

- evidence that following it improves outcomes;
- a claim that its defaults are optimal, or even good, in a given setting;
- a uniquely correct ontology of work;
- a substitute for the domain's own verification.

> Permission to advance under a policy is not evidence that a hypothesis is
> true. A rule can be enforced and still be unwise.
> (See [normative review records](../normative-review-records.md) — the same
> separation the implementation enforces between `ELIGIBLE_TO_ADVANCE` and a
> scientific claim.)

### Authority tiers

Every instruction below carries exactly one tier. The book's prose may never
be more certain than the tier of the instruction it teaches.

| Tier | Meaning | What prose may say |
|---|---|---|
| `derived` | Follows from a definition or an epistemic rule already adopted in the [theory series](../theory/) — mostly refusals (do not promote status, do not assume the check). | "This is required by the discipline; here is why the alternative is incoherent." |
| `heuristic` | A declared default with no evidence of optimality. | "This is our default and its cost; here is when to override it." |
| `conditional-evidence` | Supported by a specific frozen artifact, **only** under that artifact's declared conditions. | "In *this* setting, *this* was observed." Never "in general." |

Every instruction also states **preconditions** (when it applies),
**discretion** (what the practitioner must choose), and **failure mode** (how
following it goes wrong).

---

## 1. Intended practitioner and task class

**Intended practitioner:** someone with several serious attempts on one
objective, who can describe an attempt in writing and can obtain an outcome
verdict they do not personally control.

**Preconditions for the whole contract (`derived`, from what the method
consumes):**

| # | Precondition | Why it is load-bearing |
|---|---|---|
| P1 | **Several comparable attempts** on one stated objective (three is the practical floor for comparison; two supports contrast only). | Comparison is the only input the method has. |
| P2 | **An outcome check the practitioner does not adjudicate by opinion** — a test, a measurement, an exact checker, a proof, a third party. | Without it, every map is self-certifying. |
| P3 | **An attempt costs materially more than describing an attempt.** | If trying is cheaper than writing it down, try. |
| P4 | **It is possible to record a prediction before seeing the outcome.** | Predeclaration is the method's only defence against retrospective repair. |

**Out of scope (declared, not hedged):** single-attempt problems; tasks whose
outcome cannot be checked independently of the practitioner's judgment; tasks
where attempts are near-free; objectives that change faster than the record
accumulates (yesterday's attempts describe a different problem); and settings
where the record is someone else's curated selection of attempts and the
selection rule is unknown.

**Failure mode of ignoring scope:** the method still *runs* — it produces a
map, a boundary, and a confident next move — because nothing in it detects that
its inputs were unsuitable. A tidy map over an unsuitable population is the
most expensive output this discipline can produce.

---

## 2. The mapping trigger and the allocation rule

*Answers: what makes another unit of mapping or investigation worth more than
another unit of direct work?*

### M1 — Decision-relevance trigger

**Tier:** `heuristic`

> **Map or probe when a specific named uncertainty, if resolved, would change
> which action you take next. Otherwise act on the best current option.**

Operational test — write the uncertainty as a question with at least two
candidate answers, and name the next action you would take under each:

```text
uncertainty   "is the ceiling caused by the shared gate, or by capacity?"
answer A      gate      → next action: narrow the gate's scope
answer B      capacity  → next action: add capacity
verdict       two answers, two different actions → mapping is owed
```

```text
uncertainty   "which of these three failures was most embarrassing?"
answer A/B/C  → next action: unchanged in all three
verdict       no action depends on the answer → mapping is not owed; act
```

- **Preconditions:** P1–P4. At least one admissible action exists.
- **Discretion:** the granularity of "action" (a practitioner who calls
  everything "keep working" will never owe a map; one who splits actions into
  a hundred variants will never stop mapping). State the granularity.
- **Failure mode:** indefinite analysis presented as progress. The trigger is
  satisfiable by *any* named uncertainty, so it must be paired with M2.

### M2 — Declared spending limit

**Tier:** `heuristic`

> **Before mapping, declare the budget (time, tokens, attempts, money) and what
> you will do when it is exhausted.** Exhaustion is not permission to extend;
> it is an instruction to act on the best current option and record the map as
> unfinished.

- **Discretion:** the budget's size, and its unit. Use the unit that actually
  binds you.
- **Failure mode:** budget renegotiated mid-run in the direction of more
  mapping, every time. Record extensions as revisions, with the reason.

### M3 — Four-way allocation

**Tier:** `heuristic`

| Situation | Choose | What you produce |
|---|---|---|
| No decision-relevant uncertainty; an admissible action exists | **Act** | An attempt, recorded per §3 |
| Decision-relevant uncertainty; existing attempts are not yet comparable | **Map** | A comparison over work already done — no new attempts |
| Map yields two or more explanations that recommend different actions | **Probe** | Work designed per §4 to discriminate them |
| A stopping condition in §6 holds | **Stop** | A record of why, and what may now be claimed |

"Map" spends description effort on work already paid for; "probe" spends new
attempt effort to buy a distinction. They have different prices and should not
be chosen by mood.

- **Failure mode:** treating **probe** as a softer word for **act**. A probe
  that would teach nothing if it failed is an attempt, and should be budgeted
  as one.

---

## 3. Representation-selection procedure

*Answers: how does a practitioner choose a defensible representation?*

**The standard (`derived`, from
[08 — Earning Operational Authority](../theory/08-earning-operational-authority.md)):**

> A description of work earns the right to direct work when it preserves the
> distinctions needed to predict the consequences of the **actions available to
> you** — not merely enough to separate the outcome labels of attempts already
> finished.

Two failed attempts that both "missed the deadline" may need opposite repairs.
A description that groups them is accurate about outcomes and useless for
choosing.

**Minimum closure is a procedure, not a correct ontology.** This contract does
not claim there is one right map. It claims a reader can follow R1–R6 and end
up with a defensible one, or with a recorded disagreement between two.

### R1 — Declare the unit of attempt

**Tier:** `heuristic`

> **An attempt is the smallest unit of work that received its own outcome
> check.** Declare the unit before describing anything, and keep it fixed
> across the record.

- A change to *what was submitted* is a **new attempt**.
- A change to *the record of an attempt* is an **amendment**: append it, mark
  it, keep the original text. Corrections never overwrite.
- **Discretion:** where to cut a long session into units. Cut at checks.
- **Failure mode:** unit drift — early "attempts" are whole weeks and later
  ones are single edits, so the comparison silently compares different things.

### R2 — Retain context that could change an action

**Tier:** `heuristic`

> **Retain any condition that could change the outcome verdict or the set of
> admissible actions:** objective, budget, workload, resources, tooling and
> its version, checker and its version, who ran it.

Test for inclusion: *if this field had been different, might I choose a
different next action?* If yes, it belongs in the description — including the
things that felt too obvious to write down. Those are exactly the fields every
attempt inherits unexamined.

- **Failure mode:** "held fixed" recorded as blank, which reads later as
  "nothing was held fixed."

### R3 — No outcome leakage

**Tier:** `derived`

> **A descriptor that was not available before the attempt's outcome is
> inadmissible.** Neither is one that encodes the outcome (author's confidence
> after the fact, "the approach that was always doomed").

- **Failure mode:** a map with beautiful separation and no predictive use,
  because its coordinates are the labels in disguise.

### R4 — Construct at least two candidate descriptions

**Tier:** `heuristic`

> **Whenever the map will direct work, build a second plausible description**
> at a different granularity or in a different vocabulary — not a strawman.

- **Discretion:** how different the alternative must be. Prefer one that a
  competent colleague would actually have written.
- **Failure mode:** single-map lock-in, where the first vocabulary chosen
  determines every later boundary and nobody notices it could have been
  otherwise.

### R5 — Find where the candidates diverge

**Tier:** `heuristic`

> **State the case where the two descriptions recommend different next
> actions.** If they never diverge, keep the cheaper one — the choice was
> immaterial. If they diverge, that divergence is your probe (§4).

- **Failure mode:** comparing maps on elegance. Maps are compared on the
  actions they recommend and the outcomes those actions get.

### R6 — Compare prospectively; never re-describe backwards

**Tier:** `derived`

> **After the outcome, record which description's recommendation the evidence
> favored.** Do not re-describe past attempts so the surviving map looks
> inevitable.

This is the temporal anti-mythology rule
([08 §4](../theory/08-earning-operational-authority.md)): reuse of the archive
is not free evidence, and a record edited after the outcome cannot support the
prediction it now appears to have made.

- **Failure mode:** convergence theater — a map that improves only because its
  history keeps being rewritten.

**What remains open:** R1–R6 are a procedure for heterogeneous work histories,
not a demonstrated one. No experiment in this repository shows that two
practitioners following R1–R6 on the same uncurated history produce maps that
recommend compatible actions. That is the
[usability trial](../../corpus/experiments/contract-usability-trial/PROTOCOL.md),
and it is unexecuted.

---

## 4. Action selection: designing the next unit of work

*Replaces the earlier categorical probe prescription (vary one factor, predict
a number, maximize learning), which was stronger than the theory supports.*

### A1 — The design rule

**Tier:** `heuristic`

> **Choose a design that distinguishes the relevant explanations; predeclare a
> checkable prediction; preserve the original objective and its correctness
> conditions; and justify the design's cost against the decision it informs.**

Admissible designs include, and are not limited to:

| Design | Suits |
|---|---|
| Single-factor control | One suspected factor, cheap runs, interactions implausible or already bounded |
| Factorial / fractional factorial | Two or more factors that may interact |
| Ablation | Removing a component to test whether it carries the effect |
| Formal counterexample search | Claims stated as universal properties |
| Categorical or set-membership prediction | Outcomes that are not numeric |
| Natural comparison inside the existing record | When the discriminating pair was already run and nobody compared them |

**One-factor-at-a-time is one admissible design, not the definition of a
probe.** Varying a single factor cannot detect interactions among factors; the
NIST/SEMATECH *e-Handbook of Statistical Methods* makes this limitation of
one-factor-at-a-time procedures explicit in its design-of-experiments material
([§5.1–5.3](https://www.itl.nist.gov/div898/handbook/pri/pri.htm)). A discipline
whose whole interest is relational structure must not mandate a design that can
hide it.

- **Discretion:** the design; how many explanations count as "relevant".
- **Failure mode:** a design that discriminates nothing because both
  explanations predict the same result, dressed up as rigor by its control.

### A2 — Predeclare a prediction that can fail

**Tier:** `derived`

> **Record the prediction before the run, in a form whose failure you would
> recognize.** Numeric where the outcome is numeric ("10 ticks", not "faster");
> categorical, ordinal, set-membership, or existence-of-counterexample where it
> is not.

A prediction stated after the result is narration. A prediction no outcome
could contradict is decoration.

- **Failure mode:** predictions so wide they always hold; or a numeric form
  demanded where the domain has no numbers, which pushes practitioners into
  inventing a metric to satisfy the ritual.

### A3 — Preserve the objective and re-check correctness

**Tier:** `derived`

> **Every run re-verifies the original correctness conditions.** An outcome
> that matches the prediction by breaking the invariant the old constraint was
> protecting is a regression with good marketing, not a success.

- **Failure mode:** the objective quietly relaxing across the run series until
  the "improvement" is measured against a different problem.

### A4 — Justify the cost

**Tier:** `heuristic`

> **State what the design costs and which decision its result changes.** If no
> decision changes, the design is entertainment; if the cost exceeds the value
> of the decision, act instead.

- **Failure mode:** cost accounted only in the cheap unit (checker
  submissions) while the expensive unit (analysis, construction, human
  attention) goes unmeasured. See §5 E5.

### A5 — When two maps recommend different actions

**Tier:** `heuristic`

> **Prefer the affordable design that discriminates them.** If none is
> affordable: act on the recommendation whose failure is cheaper and more
> informative, and record the divergence as unresolved — not as settled by the
> action you happened to take.

- **Failure mode:** the map that produced the chosen action being retroactively
  credited with a distinction it never earned.

---

## 5. Evidence requirements

### E1 — Record what checked it, at the strength it has

**Tier:** `derived`

Record the check and its tier — formal proof or deterministic check >
reproducible computation or experiment > independently sourced evidence >
independent critic agreement > single judgment. Consensus among judges is still
judgment.

### E2 — Keep the conclusions separate

**Tier:** `derived`

A successful check licenses a narrower conclusion than it feels like it does.
The separations the book must keep — checked result, versus the additional
conclusion that still needs its own support — are enumerated in
[warrant boundaries](warrant-boundaries.md). These are **different
propositions, not successive confidence levels**.

### E3 — Attribution is a separate warrant

**Tier:** `derived`

> **Checking an artifact does not establish that the proposed mechanism
> produced it.** Attributing an output to a mechanism requires a binding from
> the executed attempt to the output.

This is open in the implementation too: the C1–C8 integration review records
that supplied-witness checking and membership-exact admission are demonstrated
**without** mechanism-to-output attribution, and lists the attribution slice as
future work
([review](../reviews/2026-09-12-c1c8-integration-run.md)).

### E4 — No silent promotion; misses stay in the record

**Tier:** `derived`

Hypothesis does not become evidence, and a surviving explanation does not
become an established one, without an independent mechanism that could have
refused. Failed predictions remain in the record with their original wording.

### E5 — Total cost, not the convenient cost

**Tier:** `derived`

A claim about efficiency accounts for map construction, analysis, internal
computation before a candidate is submitted, interpretation, and human effort —
with any reuse or amortization stated. Checker submissions are one component,
not the total.

### E6 — Exclusion needs a warrant covering the region

**Tier:** `derived`

> **A failed attempt supports a preference against similar attempts. Excluding
> a region of possibility requires a warrant that covers that region** — a
> proof, an exhaustive check over a declared finite space, or a demonstrated
> property of the region itself.

This is why Counterform (complement geometry) remains a labeled hypothesis in
the [claim registry](../theory/00-paper-claims.md) and may not appear in the
book as a capability.

---

## 6. Stopping conditions

A mapping or probing episode stops with exactly one of these, recorded:

| Condition | Holds when | What may be claimed | Repository vocabulary |
|---|---|---|---|
| `decision_resolved` | The uncertainty that triggered mapping is resolved enough that one action is chosen | The action and its basis — not that the map is correct | `solved` / `invariant_established` (only with an independent mechanism) |
| `no_decision_relevant_uncertainty` | Every remaining question leaves the next action unchanged | Nothing about the map's truth | `no_information_gain` |
| `budget_exhausted` | The M2 limit is reached | The map as unfinished, with its open questions | `budget_exhausted` |
| `no_discriminating_design` | Candidate explanations exist but no affordable design separates them | The divergence, explicitly unresolved | `frontier_exhausted` |
| `check_unavailable` | The outcome cannot be checked at any admissible strength | Only what weaker evidence supports, marked as such | `verification_blocked` |
| `population_inadequate` | The attempts on record are too few, too similar, or selected by an unknown rule | Nothing about the problem; something about the record | `insufficient_failure_diversity` |
| `objective_changed` | The objective moved; the record describes a different problem | Historical outcomes only; current authority is lost | (stale assessment, see [normative records](../normative-review-records.md)) |

`no_information_gain` is not permission to regenerate the same mechanism with
different adjectives.

- **Failure mode:** stopping without recording which condition applied, which
  makes a budget exhaustion indistinguishable later from a resolved question.

---

## 7. What this contract does not decide

Held open deliberately, and named in the book as open:

1. **How much mapping is right** in a given setting. M1/M2 give a usable rule,
   not an allocation optimum.
2. **Whether the procedure converges** on a stable representation.
3. **Whether a map-directed practitioner outperforms a competent undirected
   one** at equal total cost. One frozen comparison shows ordering value over a
   fixed move pool under a declared budget
   ([two-arm comparison](../../corpus/experiments/two-arm-comparison/RESULT.md));
   an adaptive baseline, a curation-only ablation, held-out structures, and a
   statistical design remain unrun
   ([preregistration](../../corpus/experiments/superiority-preregistration/PROTOCOL.md)).
4. **Whether any of this transfers** beyond the domains exercised here.

---

## 8. Adoption and change policy

**Adoption gate.** This revision becomes `adopted` when all of the following
are recorded:

| # | Gate | State |
|---|---|---|
| A1 | The contract is self-contained: a reader can execute §2–§6 without the theory series or `newf` | 🔄 to be judged by an outside reader |
| A2 | Every instruction carries tier, preconditions, discretion, failure mode | ✅ this revision |
| A3 | The [usability trial](../../corpus/experiments/contract-usability-trial/PROTOCOL.md) is executed — two independent participants, a predeclared battery containing negative controls, findings recorded whatever they say | 🔄 `draft_not_frozen` |
| A4 | Owner decision recorded here | 🔄 open |

A3 is a **usability and decision-behavior test, not a validation threshold**. A
trial in which participants follow the instructions does not make the
contract's recommendations correct; a trial in which they cannot follow them, or
in which two participants following them recommend incompatible actions, is
direct evidence that the instructions are defective or underspecified.

**Change policy.** Revisions are versioned here with the reason. An instruction
may be weakened, split, or removed at any time; **strengthening** one — moving
`heuristic` toward `derived`, or narrowing permitted discretion — requires
either a derivation from an already-adopted rule or a frozen artifact, cited
inline. Prose in the book may not exceed the tier of the instruction it
teaches; where they conflict, this contract wins and the prose is a defect.

---

## Provenance

- Basis: remote `main` at `fdf7f5e` (2026-09-12 14:20 PDT), plus the assessment
  recorded in the [book lane charter](README.md#origin).
- Standard for representation selection:
  [`docs/theory/08-earning-operational-authority.md`](../theory/08-earning-operational-authority.md) §1, §4.
- Epistemic rules: [`docs/theory/02-epistemic-model.md`](../theory/02-epistemic-model.md),
  [`docs/abstraction-safety.md`](../abstraction-safety.md).
- Practice sources being replaced or narrowed:
  [`docs/thesis/how-to-map-the-work.md`](../thesis/how-to-map-the-work.md) §3,
  [`docs/thesis/when-to-map-the-work.md`](../thesis/when-to-map-the-work.md) trigger.
- Evidence limits: [`docs/theory/00-paper-claims.md`](../theory/00-paper-claims.md),
  [`docs/reviews/2026-09-12-c1c8-integration-run.md`](../reviews/2026-09-12-c1c8-integration-run.md).
- Design-of-experiments caveat: NIST/SEMATECH *e-Handbook of Statistical
  Methods*, Ch. 5 (Process Improvement),
  <https://www.itl.nist.gov/div898/handbook/pri/pri.htm>.
