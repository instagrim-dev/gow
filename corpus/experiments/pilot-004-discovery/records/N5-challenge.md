# N5 challenge record — reorganisation without added existence

Candidate invariant: **N5**
Recorded: 2026-09-11
Source occurrences: E21 (D0, d0-run3)
Quorum vote: unanimous 3-0 (`defensible_novel`)
Scope: [es-09, es-11] — both `partial_success`

---

## Canonical statement

> **Reorganisation approaches (es-09, es-11) deliver structural improvements —
> unified representation, constraint geometry of solutions — but add no new
> existence at the QR survivor classes. Their partial_success status reflects
> progress in other dimensions (compression, constraint analysis, unification);
> the non-existence at QR primes is a separately conserved limitation.**

Preserved distinctions:
- es-09 is global, algebraic, existential: it lifts disparate parameterizations
  onto one higher-dimensional variety. Its operators are `algebraic lift`,
  `re-representation`, `unification`. It breaks surface distinctions between
  parameterizations. Its outcome is partial_success bounded by "adds no
  positivity for QR primes."
- es-11 is mixed in locality, existential: it studies algebraic relations and
  symmetries among solution denominators. Its operators are `structural
  constraint analysis`, `symmetry exploitation`. It preserves algebraic
  relations among denominators and explicitly records `breaks: []`. Its outcome
  is partial_success bounded by "structural constraints do not prove coverage of
  all primes."

The shared property is not that they are the same mechanism — they are
structurally distinct. The shared property is that both deliver **reorganisation
of existing knowledge without crossing the existence boundary at QR primes**.

---

## Challenge execution

Per AGENTS.md §Candidate invariant discipline: seven challenge probes.

---

### C1 — Known failed approach that violates the property

**Test:** Is there a known approach in the corpus with the reorganisation
character (structural improvement, unification, constraint analysis) that *does*
add new existence at QR primes?

**Result: no counterexample found.**

Systematically: every other corpus member either
(a) directly targets existence (and fails or partially succeeds for different
    reasons — es-01, es-02, es-03, es-05, es-06, es-07, es-08, es-10), or
(b) shares the reorganisation character and also fails to add QR existence.

The es-04 Type I/II classification (`partial_success`) classifies solution
shapes and partitions by divisibility but explicitly notes it "characterizes
solution shape but does not force existence for QR primes." es-04 is a
borderline case: it is closer to reorganisation/classification than to direct
existence-construction, and its boundary statement aligns with N5. It is a
potential **additional scope member**, not a counterexample.

es-12 (Type A/B congruence system, `partial_success`) is the most interesting
near-counterexample candidate. It uses congruence covering and coverage of
primes by Type A/B forms but records "coverage of all primes by Type A/B is
conjectural." The mechanism is downstream of es-11's structural analysis. Its
boundary statement is about conjectural coverage, not demonstrated existence.
es-12 does not violate N5: it remains in the partial_success regime without
demonstrated existence at QR primes.

**C1: property survives. No known corpus member with reorganisation character
demonstrates existence at QR primes.**

---

### C2 — Synthetic failed approach that violates the property

**Test:** Can we construct a plausible reorganisation approach that DOES add new
existence at QR primes?

**Analysis:**
The logical requirement for a counterexample is: a reorganisation/structural
approach (operators: lift, re-representation, unification, constraint analysis,
symmetry) that produces a new existence argument for at least one QR survivor
class — something the reorganisation structure itself generates, not inherited
from the methods being reorganised.

es-09's note is explicit: "the compression does not add positivity — it does
not force a solution to exist for every prime." If the variety lift generated
positivity arguments for QR primes from the unified structure (e.g., by showing
the intersection of rational slices at those primes is non-empty), that would
be a synthetic counterexample. But the note explicitly identifies this as what
the method CANNOT do: existence arguments at QR primes are exactly what the
unified variety structure fails to supply.

The distinction required for a counterexample is:

```
reorganisation that generates new existence
  vs.
reorganisation that preserves existing existence arguments
```

The second is what es-09 and es-11 do. The first is what N5 identifies as
absent. Constructing a synthetic approach of the first type would require adding
a positivity mechanism (e.g., lattice-point existence in the unified variety)
that goes beyond re-representation. That mechanism would no longer be purely
reorganisational — it would constitute a new mode of existence argument layered
on top of the reorganisation.

**C2: the synthetic construction needed to violate N5 requires a positivity
mechanism that is categorically distinct from reorganisation. N5 identifies
exactly that gap. No vacuous synthetic violation is possible; the required
construction is the research target.**

---

### C3 — Success that preserves the property

**Test:** Is there a success-side approach that *also* reorganises without adding
QR existence — i.e., where the conserved property discriminates correctly?

**Assessment:**
The corpus contains no `success` outcomes; all outcomes are
failure/partial_failure/partial_success. This means C3's standard form
(success that preserves or violates the property) cannot be fully executed on
this corpus.

Within partial_success members:
- es-04, es-06, es-09, es-11, es-12 are all partial_success.
- All of them share the property of not demonstrating existence at QR primes.

The property is therefore **not discriminating** within the partial_success
regime: every partial_success member satisfies it. This is by design — N5 is
a property of partial_success approaches, not a discriminator within that
regime.

What N5 claims to discriminate is the **partial_success vs. success**
boundary: no approach has crossed the QR existence boundary; all partial
successes share that feature. This claim cannot be tested for discrimination
until a `success` outcome appears in the corpus.

**C3 result: property conserved across all partial_success members (consistent
with N5). Cannot test discrimination against success because no success
outcomes are in the corpus. Claim is non-discriminating within partial_success
— by construction, not by weakness.**

---

### C4 — Lower the abstraction level: does the property split?

**Test:** At a more concrete level, are es-09 and es-11 sharing the same
feature or do the "reorganisation without existence" claim split?

**es-09 at concrete level:**
The variety lift reorganises at the level of the *parameterisation family*: it
unifies es-01/es-02/es-03/es-04-style approaches into rational slices of one
variety. The non-existence is non-existence *by the unified structure itself* —
the variety knows where the QR wall is, but knowing it geometrically is not
crossing it.

**es-11 at concrete level:**
The structural constraint analysis reorganises at the level of *individual
solutions*: it maps out the constraint geometry of what any solution must look
like and how solutions for related `n` connect. The non-existence is
non-existence *by the constraint structure* — the constraints narrow the search
space but do not fill it.

**Split analysis:**
The two are reorganising at **different levels of the problem hierarchy**:
- es-09: parameterisation-level unification (global, across method families)
- es-11: solution-level constraint analysis (mixed, within one problem instance)

However, both land at the same boundary for the same reason: the reorganisation
exposes and clarifies the structure but does not supply the positivity argument
that would force a solution to exist for QR primes. The non-existence is a
consequence of what these operations *are* (re-representation and constraint
analysis are not existence proofs), not of their level.

**boundary_delta from C4:**
```
axis:       level of reorganisation
es-09:      parameterisation-family unification (global, inter-method)
es-11:      solution-constraint geometry (mixed, intra-problem)
difference: distinct structural levels; shared ceiling (no positivity generated
            at either level from reorganisation alone)
```

The property survives as stated. The C4 boundary identifies that reorganisation
at multiple structural levels *independently* fails to generate QR existence —
which actually strengthens the claim's generality.

**C4: property survives. Boundary_delta records distinct structural levels;
the failure to add existence is independently verified at both.**

---

### C5 — Raise the abstraction level: does N5 merge with other invariants?

**Test:** At a higher level, does N5 merge with L1 (QR confinement) or N2a
(density/averaging ceiling)?

**Against L1 (confined_to_quadratic_nonresidues):**
L1 applies to methods whose *reach is confined by congruence structure* —
they cannot touch QR classes. N5 applies to methods that *reorganise without
adding existence* — they are not confined in reach; they simply do not add
positivity. These are different:
- L1's members (es-01, es-02, es-03) fail because their mechanism is
  congruence-local and cannot escape QR residue classes.
- N5's members (es-09, es-11) succeed at reorganising/unifying — they are
  not confined; they have a global view — but reorganisation is not existence.

A higher abstraction "neither crosses the QR wall" covers both but erases the
mechanism distinction. L1's obstruction is at the *reach level*; N5's
limitation is at the *generative level*. Different.

**Against N2a (density_averaging_ceiling):**
N2a's members fail because their improvement rate hits an intrinsic ceiling
(polylog too slow; QR floor). N5's members partially succeed via
reorganisation. The failure modes are orthogonal: N2a is about rate of
convergence toward existence; N5 is about the absence of a positivity
mechanism in an otherwise successful reorganisation operation.

**C5 result: N5 is stable. Does not merge with L1 (different mechanism of
non-achievement) or N2a (different type of limitation).**

---

### C6 — Sampling and publication bias

**Test:** Is the corpus biased toward approaches with this reorganisation-without-existence
structure?

**Assessment:**
- es-09 (Elsholtz–Tao 2013) is a published result. The "adds no positivity"
  observation is made explicitly in the published note, not imputed.
- es-11 (Monks–Velingker) is separately published. Its boundary statement
  ("structural constraints do not prove coverage of all primes") is explicit
  in the source.
- Both are naturally included because the corpus was built to cover known
  pre-cutoff approaches; reorganisation approaches appear because they are a
  real strategy, not because they were selected to support this property.

**C6 result: no meaningful sampling bias. Both members are independently
published; the boundary observations are explicitly sourced.**

---

### C7 — Correlation versus causal obstruction

**Test:** Is N5 observational (these approaches happen to not add QR existence)
or is the absence of existence-addition *structural in the mechanism type*?

**es-09 causal argument:**
The variety lift's operators are `algebraic lift`, `re-representation`,
`unification`. None of these operators produce a positivity argument from
scratch — they reorganise what is already there. The note is explicit: "the
compression does not add positivity." The causal connection is: re-representation
preserves existing solution sets but does not generate new ones; therefore it
cannot force existence where no solution currently exists. This is a structural
property of the operator class, not an accident of this particular instance.

**es-11 causal argument:**
Constraint analysis and symmetry exploitation characterise the *shape* of
solutions. They can rule out certain forms and link classes, but ruling out
bad forms or linking classes is a different operation from guaranteeing that at
least one good form exists. The note: "structural constraints do not prove
coverage of all primes." The causal connection is: constraint geometry is a
*necessary* but not sufficient tool for existence; constraint satisfaction
does not entail non-emptiness of the feasible set.

**C7 result: both non-existence results are causally explained by the operator
type. Reorganisation/constraint operators do not generate positivity; therefore
they cannot supply existence arguments at QR primes.**

---

## Challenge summary

| Probe | Result |
|---|---|
| C1: known failed approach violating property | No counterexample — property survives |
| C2: synthetic failed approach violating property | Required construction needs a positivity mechanism categorically distinct from reorganisation; this names the research gap, not a violation |
| C3: success preserving property | No success outcomes in corpus; all partial_success members satisfy property — consistent, not discriminating |
| C4: lower abstraction — splits? | **Boundary_delta produced:** es-09 = parameterisation-level unification; es-11 = solution-constraint geometry. Different structural levels; both independently fail to generate QR existence. **Strengthens** generality. |
| C5: higher abstraction — merges? | Stable — does not merge with L1 (reach mechanism differs) or N2a (limitation type differs) |
| C6: sampling bias | None — two independent publications, explicit boundary statements |
| C7: correlation vs causal | Both causally explained by operator type (reorganisation/constraint analysis cannot generate positivity) |

**Verdict: property survives challenge. No refinement required beyond the
C3 caveat.**

---

## C3 caveat — discrimination is unverifiable on this corpus

The standard success-contrast test cannot run because no `success` outcomes
are in the corpus. N5 claims to name a property of partial_success approaches
*relative to the unfound success regime*. That is the entire point of the
claim — no approach has crossed this boundary yet.

The claim is therefore:
- **Supported** by the two partial_success members and causally grounded (C7)
- **Non-discriminating within partial_success** (by construction, not weakness)
- **Untestable against success** on this corpus (corpus is bounded at partial_success)

This caveat must be recorded. It is not grounds to reject the claim; it is an
honest statement of what the corpus can and cannot establish. The claim becomes
testable when a success outcome appears — at which point, if that success was
generated by an operator that adds positivity on top of reorganisation, N5's
Δ(F,S) prediction is immediately testable.

---

## C4 boundary_delta and derived child hypothesis

The C4 boundary identifies that reorganisation-without-existence holds
**independently** at two distinct structural levels. This licenses one child
hypothesis:

> **N5-child-1 (proposed): Reorganisation at any structural level of the
> Erdős–Straus problem (parameterisation-family, solution-constraint, or
> other) is insufficient alone to generate QR existence. Existence at QR
> primes requires an operator that is categorically distinct from
> reorganisation/re-representation — specifically, one that generates
> positivity rather than merely rearranging known structure.**

Canonical ID proposal: `reorganisation_existence_gap`

This child sharpens N5 from an observation about two specific approaches into
a general claim about the operator class. It is the `Δ(F,S)` direction:

```
failure-conditioned shape:
    reorganisation operators do not generate QR existence

success-conditioned shape (hypothetical):
    some as-yet-unknown existence-generating operator does

Δ(N5, success-shape) hypothesis:
    the enabling condition is adding a positivity/existence operator
    that reorganisation approaches lack
```

This child is **not yet admitted**. It requires: operator review, grounding to
specific corpus passages (which are available — both notes state the boundary
explicitly), and a formal challenge run. The cheapest falsification path: find
or construct an approach whose primary operator is reorganisation/unification
and which also demonstrates existence for at least one QR survivor class.

---

## Admission recommendation

**Recommended for admission** as an interpretation claim under the canonical
statement as written. No refinement of the statement is required (the C4
sub-mechanisms strengthen rather than split the claim).

Admission requires:
- Operator attestation
- New vocabulary revision (mechanism/v5 or addition to v4) adding
  `reorganisation_without_qr_existence`
- Interpretation claims for es-09 and es-11
- No changes to original corpus notes
- Reclustering + fresh database before support is counted

Suggested canonical ID: `domain.number_theory.property.reorganisation_without_qr_existence`
Provenance: `GeneratedInterpretation`, source occurrence E21 from
pilot-004-discovery quorum, challenge record this file.

**C3 caveat to include in the vocabulary entry description**: this property is
non-discriminating within the partial_success regime by construction; discrimination
against success outcomes is untestable until a success appears in the corpus.

---

## Attributed reassessment (2026-09-11, source-faithfulness review)

*Appended per the external review recorded at
`docs/reviews/2026-09-11-manuscript-source-faithfulness.md` (finding 2). The
original record above is retained unchanged. Classification legend: SGO
(source-grounded observation), PE (proposed explanation), CCS (completed
counterexample search), IIP (inapplicable or inconclusive probe), CMA
(checked mathematical argument).*

**R-1 (C3): untestable is not passed.** C3 correctly records that success
discrimination *cannot be tested* on this corpus (no `success` outcomes
exist). Any downstream summary that counts C3 among probes the property
"survived" converts an **IIP** into a passed test. The accurate statement:
six probes produced results consistent with N5; C3 was structurally
inapplicable and remains an open obligation that activates when a success
outcome enters the corpus.

**R-2 (C2): hypothetical reasoning is not a completed search.** The C2
section argues that a violating construction "would no longer be purely
reorganisational" — a claim about the *definition* of the property's scope,
reached by reasoning over note quotations. No synthetic approach was
constructed or checked. C2's status is **PE** (a useful characterization of
the research target), not **CCS**. Its conclusion that "no vacuous synthetic
violation is possible" is definitional scoping, not a searched-and-exhausted
result.

**R-3: the operator-type implication is not established.** Downstream
comments state that absence of added existence *follows from* the
reorganisation operator type. The record supports: the two in-scope corpus
members (es-09, es-11) each *describe themselves* as reorganising without
adding positivity (**SGO**), and a proposed account of why that might be
inherent to the operator type (**PE**). A general causal impossibility —
*no* reorganisation-type operator can add existence — is not established by
two source descriptions and is not claimed at **CMA** strength anywhere in
this record.

**Net effect.** The verdict "property survives challenge, no refinement
required beyond the C3 caveat" stands *at model-judgment strength for this
corpus*, with the caveat promoted from a footnote to a coverage limit: probe
coverage is six-of-seven, C3 inapplicable by corpus construction. The
vocabulary entry (`reorganisation_without_qr_existence`, mechanism/v5)
remains admitted as a `GeneratedInterpretation`; its seed comment must not
claim unqualified seven-probe survival (corrected in
`internal/canon/vocabulary_seed.go`, same date).
