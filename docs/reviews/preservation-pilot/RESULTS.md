# Preservation pilot — results (protocol revision 1, executed 2026-09-12)

> **Correction notice (2026-09-12, same day, post-publication audit):** §8
> below records attributed corrections to this report's blinding claim and
> to §4's causal explanation. The gate outcome (§1) is unchanged. Original
> text above §8 is preserved unedited.

Twelve runs executed under the frozen preparation package (`PREPARATION.md`,
authorization revision 1, frozen at `1bc14da`), one run per cell, no
retries, no aborts, all reports within the 2,500-word cap (1,702–2,058
words). One blinded adjudication pass (`runs/ADJUDICATION.md`), sealed-
mapping commitment `3af45374560dca1b7c43780773cd881e5fef2df8771680dfcd26b7a260dc961e`
verified byte-exact at unblinding (`runs/UNSEALED-MAPPING.txt`).

## 1. Decision under the frozen gate

> "A missing essential recovery or a false reassertion rejects the
> candidate packaging for this initial trial."

**The candidate (modular) packaging is REJECTED for this initial trial**:
its M1d run (R12) missed the specified manuscript defect. Per protocol, all
results are retained; any packaging revision starts a new protocol
revision, not a replacement successful run. **The transfer experiment
remains gated** (preservation did not pass).

Material context the gate does not capture: **the baseline arm missed the
same case** (R07), and additionally missed S4d (R03). The M1 miss is not
evidence of a modular-specific preservation failure; it is a double miss
(§4). No leaderboard claim is made in either direction — one observation
per cell.

## 2. Unblinded grades

| Case | Arm | Blind label (raw run) | Grade |
|---|---|---|---|
| S3d defective | Baseline | R01 (run-01) | recovered (static-path evidence, disclosed) |
| S3d defective | Modular | R09 (run-02) | recovered (executed, both orderings) |
| S4d defective | Baseline | R03 (run-06) | **missed** (path honestly listed as not-examined) |
| S4d defective | Modular | R08 (run-03) | recovered (executed reproduction, population=2) |
| M1d defective | Baseline | R07 (run-07) | **missed** (affirmed hedges present) |
| M1d defective | Modular | R12 (run-11) | **missed** (affirmed conditioning declared) |
| S3c control | Baseline | R11 (run-04) | clean |
| S3c control | Modular | R04 (run-12) | clean |
| S4c control | Baseline | R05 (run-10) | clean |
| S4c control | Modular | R06 (run-05) | clean |
| M1c control | Baseline | R02 (run-09) | clean |
| M1c control | Modular | R10 (run-08) | clean |

Arm totals — essential recoveries: baseline 1/3, modular 2/3. False
reassertions: 0/6 overall (0/3 per arm).

## 3. What the controls established

All six corrected controls were clean, including both S3c reports, which
each **stated the surviving first-write-wins `IS NULL` fact and explicitly
negated the presentation claim** — exactly the discrimination the S3c
control was designed to test (see `PREPARATION.md` §2). Neither arm falsely
reasserted a specified defect on its corrected control. On these six cases,
neither packaging fabricated defect persistence.

## 4. The double miss on M1d (retained finding about both packagings)

Both M1d runs re-executed the Experiment E runner, verified 3/57 ≈ 5.3%
numerically, and probed hedging language — and both then affirmed the
section as adequately conditioned without flagging the sentence that labels
that rate "the budget-independent measurement." The adjudicator graded both
blanket assurances as unsupported certainty relative to the checking
performed.

Case-selection insight retained for any future revision: the M1 defect is a
**mislabeled epistemic property of a correct number**, not a wrong number.
Both packagings biased their manuscript review toward numeric artifact
agreement, which this defect passes by construction. A future manuscript
case bundle or baseline would need an explicit obligation to check the
*label attached to* each verified quantity, not only the quantity — that is
a protocol-revision matter, not a permitted edit to this run.

Secondary observation (S4d): the baseline miss was a scoping/selection
failure — the report's own not-examined list names the seam containing the
specified defect. Honest disclosure, but no actionable-recovery credit.

## 5. Economics (reported separately from the quality gate)

Predeclared limits: no fixed effort/compute conversion rates, no token
metering in this harness (declared limitation, symmetric). Per the
protocol, the economic assessment is therefore **inconclusive**; the
qualitative gate above stands on its own.

Recorded proxies (symmetric accounting):

- Primary runs: 12 parallel dispatches launched ~15:45–15:47 PT, last
  report written 15:58 PT (~11–13 min wall each, overlapping). Report
  word counts — baseline mean ≈ 1,903 (1,715–2,009), modular mean ≈ 1,861
  (1,702–2,058). No arm-systematic recurring-cost difference observable at
  this resolution.
- Adjudication: one pass, ~16:01–16:08 PT.
- One-time setup recorded to arms: `S_M` includes authoring the three case
  bundles; `S_B` includes authoring the frozen manuscript baseline (did not
  previously exist). Case validation, freeze, dispatch tooling, blinding,
  and this report are costs of running the comparison itself — reported,
  and excludable as sunk from any later adoption arithmetic.
- Declared horizon: H ∈ {5, 10, 20}, H_decision = 10 — unused, because the
  quality gate rejected the candidate packaging before any break-even
  comparison applies.

## 6. Interpretation limits (predeclared)

- One observation per cell: the seeded dispatch order is **reproducibly
  shuffled, not statistically randomized** — it controls obvious
  sequencing effects only, and averages nothing away.
- Both arms ran on the same session model: shared model provenance is
  shared provenance, not independent scientific evidence.
- A pass would not have established generality, an error-rate advantage, or
  cost noninferiority; a fortiori, this rejection establishes no general
  inferiority of modular packaging — on the engineering cases the modular
  arm preserved both recoveries while the baseline preserved one, at n=1
  per cell.
- Two raw modular reports leaked the phrase "the bundle obligation"
  (packaging self-reference) despite the envelope rule; blinded copies
  received one mechanical symmetric redaction, recorded in the mapping.
  Raw files are verbatim.
- The run-7 dispatch had one transport-level re-issue before any agent
  existed (`runs/DISPATCH.md`); the 12-primary-dispatch ceiling was
  respected.

## 7. Disposition

- Preservation: **not passed** (frozen gate; missing essential recoveries).
- False-fabrication risk on these six cases: not observed in either arm.
- Transfer experiment: remains gated.
- Next legitimate step: a **new protocol revision** may revise both arms'
  manuscript-case obligations (label-checking obligation, §4) and re-run
  under a fresh authorization. Nothing in this run's outputs may tune a
  supposedly frozen transfer test.

## 8. Attributed corrections (2026-09-12, external audit)

An independent post-publication audit (operator-supplied, exact-integer
diagnostic included) established the following. Each correction is
additive; the original sections above are preserved as written.

### 8.1 Arm-blinding was NOT achieved — procedural deviation

The "blinded" packet leaks identity in report title lines, which the
mechanical redaction never inspected. Full census (first lines, verified
2026-09-12):

```text
ARM-IDENTIFYING (6):  R01 "run-01-S3d-B"   R02 "Run 09 — M1c-B"
                      R03 "run-06-S4d-B"   R04 "…S3c-M…"
                      R05 "run-10-S4c-B"   R07 "run-07-M1d-B"
RUN-NUMBER ONLY (1):  R09 "Run 02 — …"
CLEAN (5):            R06, R08, R10, R11, R12
```

The auditor's three confirmed leaks were an undercount; the true count is
six arm-identifying titles plus one run-number title. Consequences,
recorded exactly:

- **Successful arm-blinding is not established.** The sealed-mapping
  commitment verifies content consistency between sealed and unsealed
  bytes; it does not establish concealment, and the packet contents
  demonstrably did not conceal.
- **This does not demonstrate the adjudicator used the leaked tokens**,
  and it does not reverse the double miss, which is directly inspectable
  in the raw reports (§8.2) and was independently confirmed by the
  auditor without relying on the adjudication.
- The false-reassertion endpoint (0/6) and recovery grades are quote-
  anchored in `ADJUDICATION.md` and re-checkable against the raw reports
  without trusting blinding; the *blinded* character of the adjudication
  is what cannot be claimed.
- **The original packet is preserved unmodified** (`runs/blinded/`); this
  deviation is recorded here rather than repaired, per the no-silent-
  replacement rule. Any future revision must strip title lines as part of
  the mechanical blinding step and verify with a first-line census.

### 8.2 The double miss is independently confirmed; the defect is real

The auditor re-derived the local-move generation and exact-integer
checking from the pinned runner and varied one condition at a time
(deterministic audit calculations, not pilot runs): budget 1 → 0/20,
budget 2 → 3/40 (7.5%), budget 3 → 3/57 (5.263%), L2-first ordering →
3/54, no early stopping → 3/60; budget 4 plateaus under the original
ordering. Budgets one and two falsify the unrestricted independence
claim: **the number is correct; the property attributed to it is not.**
Both raw M1d reports (run-07 baseline, run-11 modular) identify the
correct subject revision, include Experiment E in scope, and do not
discharge the obligation. The miss classification stands without the
adjudicator.

### 8.3 §4's causal explanation is RETRACTED as stated

§4 attributed the double miss to a missing label-checking obligation.
That is contradicted by the frozen packagings themselves: the baseline's
rule 2 (`arms/manuscript-baseline.md:31`) and bundle M
(`arms/modular-case-bundles.md:76`) both already require listing each
result's producing conditions — explicitly including budget, ordering,
and stopping rule — and then checking the prose against them. The correct
statement is:

> **Both frozen prompts already required this check. The miss was a
> failure to enforce an existing obligation; the cause of non-enforcement
> is unresolved.**

The "numeric-agreement bias" hypothesis in §4 likewise remains
unestablished; the modular report's own `Disp.`-caption finding (a label
defect found while affirming the values correct) is a counterexample to
any strong form of it. §4's retained case-selection insight survives only
in weakened form: the M1 defect species (mislabeled property of a correct
number) is empirically hard for both packagings under these conditions,
for reasons not yet determined.

### 8.4 Endpoint characterization tightened

The control endpoint is zero specified false reassertions across **six
corrected-control reports (three per arm)** — not twelve opportunities.
Economics remains inconclusive by rule; similar report lengths do not
establish comparable internal inference costs, so §5's "no observable
recurring-cost difference at this resolution" should be read as a
statement about the proxy's resolution, not about costs.
