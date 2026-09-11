# Review — Manuscript source-faithfulness at `4f3fa56`

**Kind:** external full-implementation review (publication / evidentiary readiness)
**Target:** remote `main` @ `4f3fa56` (24 commits past prior review baseline `3ac3e6e`)
**Received:** 2026-09-11 (~12:46 PDT; 15 minutes after the reviewed SHA was
committed — see the dated timeline at `../../paper/HISTORY.md`)
**Reviewer scope (as stated):** manuscript body and appendices, challenge
records, selected frozen experiment artifacts, vocabulary changes, adjacent
runtime contracts. Reviewer did not compile LaTeX, run the Go suite, inspect
local transcripts, or audit the bibliography. No repository changes were made
by the reviewer.
**Local verification:** see [appendix](#local-verification-2026-09-11) —
every spot-checkable claim tested locally was **confirmed**.
**Disposition:** manuscript reconciliation required before freeze; findings
tracked in the [remediation checklist](#remediation-checklist).

---

## Verbatim review

> **Remote `main` is `4f3fa56`, 24 commits beyond the last full implementation
> review at `3ac3e6e`. There is now a substantial manuscript, theory series,
> and expanded experiment record—but I would not freeze the paper for
> submission yet.** The main problem is no longer missing prose: **the
> manuscript sometimes changes what the experiments did, overstates the
> challenge evidence, and describes proposed semantics as implemented
> capabilities.**

### 1. High — The paper misstates the experimental designs and frozen results

**Locations:** `paper/geometry-of-work.tex`, Experimental Method, Results B–D,
and Appendix B.

| Subject | Manuscript says | Retained record says |
|---|---|---|
| Pilot-003 arms | B0, B1, B3 | The recorded experiment contains **B0 and B3** |
| Pilot-003 budget | Six proposals each; elsewhere "6–7 proposal slots" | **Budget eight per arm**; actual outputs seven and six |
| Pilot-003 classifier | `classify/v3` | Original experiment is pinned to **`classify/v1`** |
| Pilot-004 inputs | Context drawn from signatures and vocabulary | **Train-note bundles**; vocabulary, interpretation claims, and reference ledger excluded |
| Pilot-004 treatment | "Unmentored" versus "mentored … with challenge scaffolding" | Same instructions; **unperturbed outcomes versus permuted outcome objects** |
| Pilot-004 failed endpoint | Aggregate novel rate below a threshold | Required **reference-property recovery in a majority of runs**, error-count comparison, and unnamed-property comparison |

There is also an earlier instance: the motivating example invents four
mechanistic families, an averaging ceiling shared by every stalled approach,
and a particular boundary-crossing construction, then says the example is
drawn from the pilots. Those details are not supported by the cited experiment
narrative. **Either identify it as an explicitly fictional illustration or
reconstruct it from actual records.**

**Required correction.** Rebuild the experiment descriptions from the frozen
protocols and outputs, not from summaries of summaries. Preserve separate
fields for **configured budget, submitted count, assessed count, rule
version, original capture, repaired derivative, and later reassessment**. The
correct Pilot-004 endpoint remains useful even though it was not met.
Replacing it with a novel-rate threshold creates a different experiment
retrospectively. **A claim/evidence table containing an artifact path is not
sufficient when the sentence contradicts that artifact.**

### 2. High — The novel-hypothesis challenge records overstate what was established

**Locations:** `records/N2a-challenge.md`, `records/N5-challenge.md`,
manuscript Experiment D, and the v4/v5 vocabulary comments.

**N2a contains unsupported mathematical implications.** The challenge record
treats polynomial growth of a solution-count average as sufficient to cross an
existence threshold, and says that failing to construct a falsifying example
"reconfirms" the property. **Average growth alone cannot establish positivity
at every input.** Counterexample to the general implication:

\[
f(n)=
\begin{cases}
2n,&n\text{ even},\\
0,&n\text{ odd}.
\end{cases}
\]

For even \(N\), the average over \(1,\ldots,N\) is \(N/2+1\): polynomial
growth, with half the inputs still zero. This does **not** refute a particular
Erdős–Straus theorem; it shows the record needs the exact additional
assumptions connecting its averaged quantity to pointwise existence.

The record also writes an exceptional-set bound of the form
\(N\exp[-c(\log N)^{2/3}]\) and says it decays to zero. For fixed \(c>0\),
**the displayed expression grows without bound**; its *ratio to* \(N\) tends
to zero. Vanishing density, an empty exceptional set, and a positive density
floor must not be interchanged.

**The manuscript does not faithfully summarize even that record.** Experiment
D changes N2a's scope from **es-05/es-10** to "all stalled … mechanisms,"
identifies es-01/es-02 as partial successes, substitutes es-01 for the
record's es-06 contrast, and reports a higher-level merge where the challenge
record says the proposed merge is defective.

**N5's seven-probe summary also needs qualification.** N5's C3 explicitly
says success discrimination **cannot be tested because the corpus has no
success outcomes**; its synthetic-counterexample discussion reasons about a
hypothetical construction. Nevertheless the vocabulary comments state the
property survived all seven probes and that absence of existence follows from
the operator type.

**Required correction.** Retain the original challenge notes, but add an
attributed reassessment distinguishing:

```text
Source-grounded observation
Proposed explanation
Completed counterexample search
Inapplicable or inconclusive probe
Checked mathematical argument
```

"No counterexample constructed" does not reconfirm a hypothesis. An untestable
contrast does not become a passed discrimination test. Two source descriptions
do not, by themselves, establish a general causal impossibility. **The
candidates remain worth investigating. What is not yet supported is the
advertised strength of their survival and causal grounding.**

### 3. High — The epistemic section reintroduces the promotion the project prohibits

**Location:** `paper/geometry-of-work.tex`, Epistemic Model.

The manuscript says verification requires a stronger mechanism, including "at
minimum an independent model family that did not generate the original claim,"
and elsewhere describes domain-space verification as "inherently
authoritative." Both are too strong: an independent model can corroborate
while remaining model judgment; a domain-space check can be incomplete,
assumption-dependent, or aimed at the wrong claim. The governing distinction
is `ModelJudgment != Verification`.

Related lifecycle discrepancy: the manuscript summarizes `surviving` as "all
probes failed to falsify," while the documented implementation permits
survival after at least one completed-negative attack, with other attacks
potentially inapplicable or inconclusive. Materially different coverage
claims.

**Required correction.** Describe evidence by **what was checked, against
which claim, under which assumptions, and with what unresolved coverage**.
Provider independence is a useful attribute, not a substitute for that
account.

### 4. High (systems claim) — Architecture blends implemented behavior with proposed extensions

**Location:** manuscript System section and Figure 4.

- **Frontier eligibility:** paper says surviving *or weakened* claims feed
  generation. Production `targetableStates` contains only `surviving` and
  `operator_attested`; `weaken` is explicitly excluded.
- **Boundary persistence:** paper says boundary deltas are persisted as
  first-class records; the inspected challenge store has no dedicated
  `boundary_delta` field — the conceptual `ChallengeResult` is documentation,
  not an implemented typed persistence contract.
- **Challenge execution:** paper describes every candidate receiving a
  seven-probe lifecycle; the default challenger proposes a narrower attack
  set and deliberately does not synthesize counterexamples. The N2a/N5
  Markdown analyses are not evidence that the CLI campaign executed those
  seven checks.

**Required correction.** Annotate the architecture with three explicit
statuses: **implemented and exercised** / **performed externally or manually,
with retained artifacts** / **specified extension, not yet demonstrated**.
Keep the larger GoW loop in the theory section; show the implemented subset
separately. The review materials do not yet demonstrate the stronger
self-referential cartography loop (comparing representational choices,
prospectively selecting a revised basis, testing improvement); that remains a
research direction.

### 5. Medium — The completion gate checks presence, not consistency

**Location:** `paper/PUBLICATION-PLAN.md` and manuscript assembly.

The tracker marks empirical evidence complete and submission
"ready-modulo-user" — premature given the mismatches above. Assembly defects:
duplicate section declarations and labels for **Motivating Example, System,
Experimental Method, Limitations, Future Work, Conclusion**, plus the opening
subsection. Appendix C misassigns novel-theme lineage: it assigns E13/E20 to
N1 and E15 to N5, while the closure scorecard assigns **E03 to N1,
E13/E16/E20/E23 to N2a, E15 to the annotation-pattern theme, and E21 to N5**.

**Fix the assembly and replace "every claim has a path" with "every claim
agrees with the identified source."** The tracker should reopen empirical
consistency, epistemic consistency, and implementation traceability.

### Overall disposition (reviewer)

The repository has made a real transition: theory, teaching material,
experiment closure, and manuscript content now exist as separate artifacts.
The paper's retained limitations (synthetic corpus, same-family adjudication,
failed aggregate endpoint, no mathematical verification, no convergence
guarantee) should stay. **The next step is manuscript reconciliation, not
another experiment or infrastructure layer.** Priorities:

1. Rebuild the methods/results sections directly from frozen protocols and outputs.
2. Reassess N2a/N5's challenge conclusions; preserve unsupported steps as unresolved.
3. Align the epistemic and systems sections with actual evidence and implementation.
4. Run a document-consistency pass before external technical review.

**Bottom line: the writing has advanced faster than the evidence
reconciliation. The research program remains worth developing, but the current
paper is not merely awaiting acknowledgments and a tag. It needs a
source-faithfulness pass before its polished narrative can be trusted.**

---

## Local verification (2026-09-11)

Spot-checks executed against the local checkout at `4f3fa56` (local HEAD ==
reviewed SHA) immediately after receiving the review. Every checkable claim
tested was **confirmed**:

| Review claim | Local check | Result |
|---|---|---|
| Pilot-003 recorded arms are B0 and B3 | `corpus/experiments/pilot-003/captures/capture-record.json` — `arms` contains `b0_undirected`, `b3_invariant_guided` only | **Confirmed** |
| Actual outputs seven and six | same record: `proposals: 7` (B0), `proposals: 6` (B3) | **Confirmed** |
| Budget eight per arm | pilot-003 records contain `budget_count: 8` (×3) | **Confirmed** |
| Experiment pinned to `classify/v1` | `grep -r 'classify/v'` over pilot-003: `classify/v1` ×10, `classify/v2` ×6, **no `classify/v3`** | **Confirmed** |
| Manuscript says `classify/v3`, B1 arm, "6–7 proposal slots" | `paper/geometry-of-work.tex` lines 855, 881, 926, 946, 1646, 1653, 1655, 1698 | **Confirmed** |
| `targetableStates` excludes weakened | `internal/pipeline/frontier.go:32` — `[]string{"surviving", "operator_attested"}` | **Confirmed** |
| Duplicate section declarations and labels | duplicated `\label{}`: `sec:conclusion`, `sec:example`, `sec:future`, `sec:limitations`, `sec:method-exp`, `sec:system` — exactly the six sections named | **Confirmed** |

Not yet locally verified (require careful reading, not greps): the Pilot-004
input/treatment/endpoint rows, the N2a/N5 mathematical analysis, the Experiment
D scope changes, and the Appendix C lineage table. The pattern of seven-for-seven
confirmations on the checkable subset warrants treating the remainder as
credible pending item-by-item reconciliation.

## Remediation checklist

Tracked here so the review does not live only in a transcript. Order follows
the reviewer's priorities.

- [x] **R1 — Methods/results rebuild.** Re-derive Experimental Method,
  Results B–D, and Appendix B from the frozen protocols and capture records;
  carry separate fields for configured budget / submitted / assessed / rule
  version / original capture / repaired derivative / reassessment. Restore
  Pilot-004's actual inputs, treatment, and (failed) endpoint. Resolve the
  motivating example: label fictional or reconstruct from records.
- [x] **R2 — N2a/N5 reassessment.** Append attributed reassessments to
  `N2a-challenge.md` / `N5-challenge.md` using the five-way distinction
  (source-grounded observation / proposed explanation / completed
  counterexample search / inapplicable-or-inconclusive probe / checked
  mathematical argument). Correct the average-growth→pointwise-existence gap
  and the \(N\exp[-c(\log N)^{2/3}]\) decay misstatement. Downgrade the v4/v5
  vocabulary comments' seven-probe survival claims. Fix Experiment D's
  scope/contrast/merge misreporting.
- [x] **R3 — Epistemic section.** Remove "independent model family" as a
  verification floor and "inherently authoritative" for domain checks; state
  the checked/against-what/under-which-assumptions/with-what-coverage form.
  Align the `surviving` definition with the implemented coverage semantics.
- [x] **R4 — Architecture statuses.** Annotate the System section and Figure
  4 with implemented-and-exercised / performed-externally-with-artifacts /
  specified-extension. Correct frontier eligibility (`surviving`,
  `operator_attested` only), boundary-delta persistence, and challenge
  execution descriptions.
- [x] **R5 — Assembly and gate.** Deduplicate the six section
  declarations/labels; fix Appendix C lineage per the closure scorecard
  (E03→N1; E13/E16/E20/E23→N2a; E15→annotation-pattern; E21→N5). Reopen
  empirical consistency, epistemic consistency, and implementation
  traceability in `PUBLICATION-PLAN.md`; the gate criterion becomes "every
  claim agrees with the identified source."

## Relationship to repository discipline

This review is a
[falsification review](../theory/09-falsification-review.md) against the
manuscript-as-claim: verdict **weakened** (narrative damaged; underlying
program intact), verifier rung *independently sourced evidence* (external
reviewer against frozen artifacts, spot-confirmed locally at deterministic
rung for seven claims). Per the primitive's postcondition, the tracker's
"ready-modulo-user" status cannot survive this record; per
[08](../theory/08-earning-operational-authority.md), repairing the
manuscript's account *after* this observation earns no predictive credit —
the reconciliation must be scored as correction, not confirmation.

