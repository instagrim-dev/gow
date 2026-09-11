# Finding 001 — LLMs Can Propose Unencoded Structural Regularities Over a Failure Corpus

**Status:** capability finding — provisional, narrowly scoped
**Source experiment:** [Pilot-004 (discovery pilot)](../../corpus/experiments/pilot-004-discovery/)
**Experiment verdict:** `negative_inconclusive` (predeclared aggregate endpoint not met)
**Recorded:** 2026-09-11

---

## Why this note exists separately from the experiment record

Pilot-004's own [`RESULT.md`](../../corpus/experiments/pilot-004-discovery/records/RESULT.md)
and [`CLOSURE-SCORECARD.md`](../../corpus/experiments/pilot-004-discovery/records/CLOSURE-SCORECARD.md)
are the **immutable experimental record**: the frozen protocol, the arms, the
quorum adjudication, the predeclared criteria, and the verdict. This is a
different kind of document. It records a **capability finding** — a claim about
what one constituent operation was observed to do — abstracted from the
experiment that happened to surface it.

The distinction matters because the experiment is *negative under its own frozen
effectiveness criterion* while simultaneously producing a *positive capability
observation* about one operation the system depends on. Those two facts must not
be allowed to overwrite each other. The finding is not "Pilot-004 succeeded" (it
did not); the finding is that a specific operation occurred at all.

See the [epistemic model](../theory/02-epistemic-model.md) for the belief rules
this note is disciplined by, and the
[thesis](../thesis/why-solve-by-shape.md) for why shape generation being possible
is the load-bearing question.

## The finding

> Given a corpus of failed and partial approaches to Erdős–Straus, an LLM
> proposed candidate shared structural properties (`ShapeClaim`s) that were
> **not present in the operator reference ledger**, and several of those
> candidates received provisionally defensible assessments under structured
> quorum criticism.

In the vocabulary of the [geometry of work](../theory/00-geometry-of-work.md):
**shape generation occurred.** The system exhibited, at least once, the ability
to read a failure corpus and emit a structural regularity that was not handed to
it. That was implicitly assumed to be one of the *hard* model-native operations;
this is the first direct observation that it happens.

## What the evidence actually was

From the frozen Pilot-004 record (23 candidate-property occurrences across six
sealed captures, quorum-adjudicated):

- **10 occurrences** received a `defensible_novel` disposition — i.e. candidate
  properties **absent from the reference ledger**. Collapsed by theme, these are
  approximately **5 distinct themes** (N1–N5), not 10 independent discoveries.
- The two cleanest, scope-objection-free themes were:
  - **N5** — *reorganisation without added existence* (es-09, es-11): unanimous
    (3-0), no scope objection.
  - **N2a** — *intrinsic density/averaging ceiling* (es-05, es-10): four
    independent formulations, each unanimous (3-0), appearing across both arms.
- **Recurrence across fresh runs** is itself informative: the proposer
  repeatedly converged on the same theme rather than generating arbitrary
  variation. That answers "does the model converge on something?" — a different
  question from "how many independent things did it find?"

## Scope — what this finding does *not* claim

This note is deliberately narrow. Each of the following is a **non-claim**, and
each corresponds to a real limitation in the frozen record:

1. **The frozen aggregate endpoint was not met.** Two of three predeclared
   criteria failed. D1 (guided) recovered an unnamed reference property (L3) in
   only **1 of 3 runs**; the majority-of-runs threshold required 2 of 3. L4 was
   not recovered in any run by majority. The experiment's verdict is
   `negative_inconclusive`, and that verdict stands.

2. **No candidate is an established invariant.** Every `defensible_novel` entry
   is a `proposed` `ShapeClaim` (`candidate ≠ established`). None has been
   grounded, challenged, or promoted. Eligibility for downstream use depends on
   each candidate's **own** challenge results, not on the pilot's outcome — a
   failed experiment does not invalidate its artifacts, and a successful one
   would not validate them.

3. **"Novel" means novel *relative to this ledger*.** It does **not** establish
   novelty in the mathematical literature, nor that the curated reference pass
   overlooked valid invariants.

4. **Quorum agreement is model judgment, not verification.** The three lanes were
   the same model family (claude / cursor-agent) under role-differentiated
   prompts. Agreement is *within-family consistency*, not independent-source
   validation. `ModelJudgment ≠ Verification` — quorum consensus tops out at
   model judgment under the [verification hierarchy](../theory/02-epistemic-model.md).

5. **The control has a known causal ambiguity.** The D0 control used
   outcome-permuted notes whose **prose still retained the original outcome
   information**. A D1 > D0 gap could reflect sensitivity to coherent-vs-corrupted
   descriptions rather than discovery of failure structure. The negative verdict
   moots this for now but the limitation is recorded.

6. **Provenance caveat.** A key-seal gap existed (`key_sha256` recorded
   post-adjudication); git history and sealed capture digests provide partial
   tamper-evidence. Adjudication was blind by construction.

## Why the finding survives the negative verdict

The experiment measured **run-to-run consistency of rediscovery** — a claim about
the *method's reliability*. That is what failed. The capability finding is about
a *constituent operation*: did unencoded shape generation occur at all? It did.

These are separable because the [operational theory](../theory/03-shape-guided-search.md)
treats invariant/shape generation as **one instrument among several** in the
`geometry → boundary → structural delta` step. A pilot can fail to show that the
instrument *reliably beats a control* while still demonstrating that the
instrument *produces output of the claimed kind*. Finding 001 records only the
latter, at exactly the strength the evidence supports: **provisional, model-judged,
unpromoted.**

## Consequences for the pipeline

The finding changes what the next action is, not what is believed:

- The cleanest candidates (**N5**, **N2a**) are eligible to enter the challenge
  stage now — grounding, known/synthetic counterexample, success-preserving
  check, and abstraction split/merge — per the
  [epistemic model](../theory/02-epistemic-model.md). Survival there, not the
  pilot, determines admissibility.
- Two candidates need **scope correction before challenge** (E15's
  annotation-pattern-vs-mechanism-level claim and the es-07 inclusion in N2b);
  do not carry an unresolved scope question into challenge.
- Design learnings for the next pilot (run-level credit as a secondary metric,
  reference distinctions as explicit negative examples, a control without prose
  leakage, admission language that distinguishes "conflicts with corpus" from
  "correctly follows permuted input") are logged in the experiment record.

## Update — 2026-09-11: N2a challenged and survived

The primary next action above was executed. N2a (*intrinsic density/averaging
ceiling*, scope [es-05, es-10]) was run through the full seven-probe challenge
discipline
([`N2a-challenge.md`](../../corpus/experiments/pilot-004-discovery/records/N2a-challenge.md)).
Outcome, stated at exactly its earned strength:

- **Disposition: survives with one refinement.** No known-counterexample, no
  falsifying synthetic construction, and the strongest contrast case (es-06,
  `partial_success`) reaches its ceiling by a *different* mechanism (the QR wall),
  which preserves rather than refutes the distinction.
- **The productive probe was C4, and it produced a `boundary_delta`, not a
  refutation.** Lowering the abstraction level split the single "structural
  ceiling" into two distinct sub-mechanisms along an explicit axis —
  `rate_subthreshold` (es-05: polylog sieve-average growth is below the
  polynomial threshold to force existence) vs `qr_absolute_floor` (es-10: the
  reciprocity-protected QR-survivor classes cannot be emptied by any finite
  congruence battery). This is a **resolution gain**, exactly the
  `weaken`/`split`-with-sharp-`boundary_delta` case the
  [epistemic model](../theory/02-epistemic-model.md) and
  [shape-guided search](../theory/03-shape-guided-search.md) describe: the
  claim's scope did not shrink because it was wrong, but because the map gained
  resolution.
- **One derived child, `proposed` and unadmitted.** The C4 boundary licensed
  `N2a-child-1` (*obstruction independence*: crossing es-05's rate ceiling does
  not automatically cross es-10's QR floor), which enters `proposed` with **no
  inherited authority** and carries its own cheapest-falsification path —
  anti-fractal discipline holding exactly as specified.
- **Admitted as a `GeneratedInterpretation`, not verified evidence.** The refined
  property was recorded as vocabulary term
  `domain.number_theory.property.density_averaging_ceiling` in a **strict
  superset** revision (`mechanism/v4`, commit `bf331ae`), with the two
  sub-mechanisms recorded as scope notes. No corpus notes were altered.

This is the first end-to-end instance of the challenge → `boundary_delta` →
scoped refinement → attested-interpretation loop running on a candidate that the
system *generated itself*. It does **not** upgrade the finding's strength: the
property is now a challenge-survived, operator-attested interpretation claim, not
independently verified mathematics, and Pilot-004's aggregate verdict remains
`negative_inconclusive`. What it adds is evidence that the shape a model
proposed was *durable under adversarial criticism*, which is a strictly stronger
observation than "shape generation occurred at all."

## Provenance

- Immutable experiment record:
  [`corpus/experiments/pilot-004-discovery/`](../../corpus/experiments/pilot-004-discovery/)
  — `RESULT.md`, `CLOSURE-SCORECARD.md`, `INTERPRETATION-NOTES.md`, frozen
  protocol, quorum artifacts.
- Quorum adjudication commit `78e73d2`; result corrections `0aac153`.
- This finding cites that record and adds no new adjudication or mathematical
  verification.
