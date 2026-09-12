# N2a-child-1 challenge record — obstruction independence of the density/averaging ceilings

Candidate: **N2a-child-1** (proposed 2026-09-11 in `N2a-challenge.md` §Derived
child hypothesis; entered `proposed` independently of N2a; requires its own
challenge to advance).

Recorded: 2026-09-12.
Parent: N2a (`density_averaging_ceiling`, admitted 2026-09-11 to
`mechanism/v4`; C4 boundary_delta recorded, coverage incomplete at C2/C7
per attributed reassessment).
Canonical ID proposal: `density_averaging_ceiling_obstruction_independence`.

Epistemic-tier legend used inline (same as N2a/N5 attributed reassessments):

```text
SGO  Source-grounded observation (verifiable against the frozen corpus)
PE   Proposed explanation (plausible argument; not independently verified)
CCS  Completed counterexample search (nonempty search actually performed)
IIP  Inapplicable or inconclusive probe (no completed decisive attack)
CMA  Checked mathematical argument (steps individually verified)
```

---

## Canonical statement (from N2a-challenge.md, restated here)

> **N2a-child-1**: The rate-limited ceiling (es-05's mechanism) and the
> reciprocity-floor ceiling (es-10's mechanism) are distinct obstructions.
> A proposal that crosses one does not automatically cross the other. In
> particular, achieving polynomial average growth (crossing es-05's
> ceiling) within a congruence-class-structured framework would still
> face the QR-survivor floor (es-10's mechanism) unless it also breaks
> QR confinement.

Scope: es-05 (Bombieri–Vinogradov style averaging) and es-10 (Vaughan
congruence density) — the same two mechanisms N2a covers. N2a-child-1
does not extend the population; it *decouples* the sub-obstructions
attributed to those two mechanisms.

Predeclared falsification path (from N2a-challenge.md line 315–319):
find (or construct) a single mechanism that simultaneously (a) achieves
polynomial main-term growth in the averaged solution count and
(b) covers, or otherwise bypasses, the QR-survivor congruence classes.

Predeclared decision rule: `surviving` requires at least one completed
decisive negative over a nonempty eligible population;
IIP/PE/inapplicable probes do not earn survival even in aggregate.

---

## Challenge execution

Seven probes applied per AGENTS.md §Candidate invariant discipline and the
protocol in `docs/invariant-challenge.md`.

---

### C1 — Known corpus counterexample to the decoupling

**Test:** Is there a corpus failure whose mechanism *already* achieves
both crossings (polynomial average growth AND QR coverage), showing that
the two obstructions co-occur or that one crossing automatically
delivers the other?

**Corpus survey (SGO):**

- es-01, es-02, es-03, es-08: QR-confinement failures. None uses an
  averaging apparatus that would engage the rate obstruction; they fail
  on QR classes, i.e. exhibit exactly one of the two obstructions.
- es-05: rate-limited averaging. Its failure trace is the polylog-vs-
  polynomial growth gap; the note does not discuss QR-class coverage as
  its failure locus — QR would only matter *after* rate crossing.
- es-07: finite-verification failure (epistemological, per OQ1's
  resolution). Neither of N2a-child-1's obstructions is the operative
  one.
- es-10: QR-density failure. Its exceptional-set floor is the QR
  survivors; growth rate is not identified as its ceiling.

No corpus member exhibits both obstructions simultaneously, and no
corpus member exhibits one obstruction *because it crossed the other*.
The atlas therefore does not contain a counterexample to the
decoupling.

**C1: no counterexample (SGO). Property survives C1.**

Bias note carried forward to C6: this outcome is unremarkable because
n=1 corpus mechanism represents each obstruction and there is no
mechanism in the population that has been *pushed against both barriers
in a single attempt*.

---

### C2 — Synthetic counterexample: construct a joint-crossing mechanism

**Test:** Can we construct — with admissibly supported structure — a
plausible mechanism that jointly achieves polynomial average growth AND
covers the QR-survivor classes, thereby falsifying the decoupling?

**Attempted construction shape (PE):**

A joint-crossing mechanism would have to compose:

1. A **counting apparatus** whose main term for solution counts of
   `4/n = 1/x + 1/y + 1/z` (or an equivalent restated identity) scales
   as `N^ε` (some ε > 0) rather than `(log N)^k`. Candidates: a
   circle-method / Fourier-analytic identity with a nontrivial main
   term; a delta-method decomposition; a moment identity for a divisor
   sum that admits polynomial-rate estimates.

2. A **QR-covering apparatus** that either (i) generates witnesses in
   the QR-survivor classes directly (via a construction of Mordell type
   restricted to those classes), or (ii) imports reciprocity into a
   covering argument so that the congruence battery is not restricted
   to QR-touchable classes.

**Assessment (SGO on frozen record; PE on construction):**

No such joint construction is exhibited in the atlas, in the frozen
Pilot-003/004 proposals, or in the N2a/N5 challenge records. A search
of the corpus for "polynomial-rate main term" mechanisms returns
nothing; a search for "QR class direct construction on n ≡ QR"
similarly returns nothing.

Producing a completed synthetic construction is not accomplished in
this record. The two candidate apparatuses above are **PE**: each is a
plausible shape a construction could take, but each also imports its
own independent research problem (there is no known polynomial-rate
solution-count identity for the Erdős–Straus family; there is no known
reciprocity-carried covering argument for its QR-survivor classes).

That the joint move requires *two* apparently-independent research
advances is itself an argument *for* the decoupling — but this argument
runs at PE strength, not CMA strength: it observes only that the shape
of a joint crossing decomposes cleanly into two currently-unsolved
sub-problems, not that the two sub-problems are *provably*
independent.

**C2: no completed synthetic construction (IIP).** The absence of a
construction does not confirm N2a-child-1 (per the R-1 discipline in
N2a-challenge.md's attributed reassessment). A **PE** describing the
shape of a joint-crossing mechanism and its two independent
sub-obstacles is retained.

---

### C3 — Partial-success (or success) that preserves the decoupling

**Test:** Does the corpus's one partial-success member, es-06,
demonstrate one obstruction *without* the other — providing evidence
that the two occur separably rather than jointly?

**Analysis (SGO on es-06 record):**

es-06 uses divisor-structure representation counting (reduction to
two-fraction question with lower bounds `f(n) > 0` for almost all
primes). Its stated failure boundary is the QR wall: "the
representation bounds [...] degrade exactly for primes that 'resemble
a perfect square' to small moduli — the quadratic-residue primes"
(quoted in N2a-challenge.md line 100–105).

Regarding N2a-child-1's two obstructions:

- **QR-survivor floor:** es-06 exhibits this obstruction (its lower
  bound degrades at QR primes).
- **Rate-subthreshold:** es-06 is *outside the scope* of N2a-child-1's
  rate-subthreshold claim, because es-06 does not attempt existence-
  forcing via averaging in the sieve/density sense. It does not face
  the rate ceiling because it does not engage the averaging apparatus
  at all.

es-06 therefore shows *one* obstruction (QR wall) active in the
absence of the *other* (rate-subthreshold, which is not engaged rather
than crossed). This is consistent with the decoupling — the
obstructions do not automatically co-occur — but the inference is
weaker than "es-06 crossed one obstruction, still faced the other,"
because the un-faced obstruction was un-engaged rather than un-crossed.

**C3: consistent with decoupling; not a strong discrimination (SGO,
inapplicable-in-the-strict-sense).** No decisive negative earned here.

---

### C4 — Lower abstraction: does the decoupling split further?

**Test:** Do the two sub-obstructions themselves decompose, and if so
do the decompositions preserve or erase the decoupling?

**Decomposition (PE):**

*Rate-subthreshold obstruction (es-05):*
- (a) Sieve error bounds (Bombieri–Vinogradov) are smooth: they preserve
  average behaviour on residue classes but do not admit polynomial main
  terms.
- (b) The main term for the ES solution-count average scales as a
  product of divisor density × admissibility density, yielding
  `(log N)^k` under standard sieve heuristics.

*QR-floor obstruction (es-10):*
- (c) Congruence covering by any finite battery of moduli cannot inject
  quadratic-reciprocity structure; QR classes remain outside the reach
  of any such cover.
- (d) The QR-survivor density is bounded below by a positive constant
  of `∏(1 − 1/(2p))`-type shape, giving a strictly positive floor for
  the exceptional set's density.

**Cross-checking for shared machinery:** components (a)/(b) sit in the
analytic-sieve / distribution-of-solution-counts machinery; components
(c)/(d) sit in the algebraic-reciprocity / congruence-cover machinery.
The four components partition cleanly into two disjoint mathematical
families under the standard categorisation of number-theoretic
obstructions.

C4 therefore does not produce two admissible child predicates in the
sense of the challenge protocol (`split` requires
pairwise-disjoint support among corpus members; here the decomposition
is *within* each of es-05 and es-10, not across new corpus support).

**C4: property survives at finer abstraction; no productive split into
two child predicates with disjoint corpus support (PE).**

---

### C5 — Higher abstraction: does the decoupling merge with anything?

**Test:** At a higher abstraction level, does N2a-child-1 merge with an
existing invariant or another N-family theme?

**Candidate merges:**

- **With L6 (distinct-role finding):** L6 is a role-preservation
  finding across corpus mechanisms, not a claim about *obstruction
  independence*. A merge would erase the concrete "these two
  mathematical obstructions arise from disjoint machineries" content
  and replace it with generic role-distinctness.
- **With N3 (locality-without-global-coupling):** N3 concerns local vs
  global reasoning modes. N2a-child-1 concerns obstruction independence
  within the density/averaging family. Different axes.
- **With a hypothetical meta-invariant "distinct obstructions require
  distinct machinery":** such a meta-invariant is not attested in the
  corpus and would be an abstraction jump that AGENTS.md §Abstraction
  safety would reject as erasing predictive content until grounded.

**C5: no productive merge available. Property is stable at its stated
abstraction level (SGO).**

---

### C6 — Sampling and publication bias

**Test:** Is the decoupling supported by the corpus at a strength
appropriate for its assertion, or is the evidence base too sparse for
the generalisation to hold?

**Support audit (SGO):**

The decoupling's inductive base is:
- n = 1 corpus mechanism for the rate-subthreshold obstruction (es-05).
- n = 1 corpus mechanism for the QR-floor obstruction (es-10).
- n = 0 corpus mechanisms attempting the joint crossing.

The assertion "the two obstructions are distinct" is trivially true
at n = 2 (two different mechanisms have two different-looking
obstructions). The substantive assertion — that no *joint*-crossing
mechanism exists or can exist — is not tested by the corpus: it is a
claim about a population (joint-crossing attempts) with zero observed
members.

Under the bias-critique verdict rule in `docs/invariant-challenge.md`
(a recomputed support below the mining threshold weakens the
hypothesis), N2a-child-1 sits at the boundary of admissible support.
The parent N2a is supported by four occurrences (E13/E16/E20/E23,
unanimous). N2a-child-1 draws on the same two atlas members and adds
a decoupling claim whose support is *the pair being observed to differ
in obstruction*, not multiple independent observations of the
decoupling.

**C6: real bias-critique landing. The decoupling holds observationally
over n = 2 but is under-tested as a general claim; the "no joint
crossing exists" reading is not supported by the corpus and rests on
absence of an unattempted move. (SGO on the count; PE on the general
reading.)** This is a **weakening** signal under the protocol.

---

### C7 — Correlation versus causal obstruction

**Test:** Is the decoupling causally grounded (the two obstructions
arise from structurally disjoint mathematical constraints and cannot
be jointly overcome by a single mechanism as a matter of structural
necessity) or merely correlational (the two corpus mechanisms happen
to fail for two different-looking reasons)?

**Causal-side argument (PE):**

The rate obstruction reflects an analytic fact about smooth
error-term behaviour in sieve inequalities: Bombieri–Vinogradov, the
Brun–Titchmarsh inequality, and the Turán–Kubilius framework all
produce average bounds without polynomial main terms *within their
own machinery* — the polynomial main term would require a
non-sieve-based counting identity.

The QR-floor obstruction reflects an algebraic fact about
reciprocity: a finite congruence battery of moduli cannot induce
quadratic-reciprocity-carried covering of the QR-survivor classes
because reciprocity is a global multiplicative property not
transportable through additive covering.

These are answers to different mathematical questions
(*How fast does the averaged count grow?* vs *Which n are forever
outside any finite congruence cover?*). Jointly overcoming them
requires composing two apparatuses drawn from different
mathematical families — a compositional independence that is
*structurally suggestive*.

**Limitations of the causal argument (honest labelling):**

The argument above is **PE**, not **CMA**. What it establishes:
- The two obstructions currently arise from operationally disjoint
  machineries in the corpus.
- A composition would require importing structure from both
  machineries into a single mechanism.

What it does *not* establish:
- That no single mechanism can, in principle, simultaneously exhibit
  polynomial main-term growth and inject reciprocity into a covering
  argument. Absence of a known such mechanism is not proof of
  impossibility.
- That the two mathematical questions are provably independent (a
  cross-machinery identity linking sieve main terms to
  reciprocity-covering density has not been ruled out and has not
  been sought).

**C7: PE-strength causal reasoning. Not CMA. No verified proof that the
decoupling is structurally necessary rather than observed.**

---

## Challenge summary

| Probe | Verdict | Tier |
|---|---|---|
| C1 known counterexample | No counterexample in atlas | SGO |
| C2 synthetic counterexample | No completed construction | IIP (+ PE) |
| C3 partial-success preservation | Consistent, not discriminating | SGO |
| C4 lower abstraction | Survives; no productive split | PE |
| C5 higher abstraction | No productive merge | SGO |
| C6 sampling bias | Weakening: n = 1 per obstruction; joint-crossing untested | **SGO — weakens** |
| C7 causal vs correlational | PE-level causal argument, not verified | PE |

Completed decisive negatives: **0** (C2 incomplete; no probe cleanly
resolved every applicable case over a nonempty eligible population in
the decisive-negative sense required by the protocol).

Weakening landings: **1** (C6).

Falsifications: **0**.

Splits: **0** (C4 produced a decomposition but not two admissible
child predicates over disjoint corpus support).

Merges: **0**.

---

## Disposition under the predeclared decision rule

Recall the predeclared rule (top of this record): `surviving` requires
at least one *completed decisive negative* over a *nonempty* eligible
population. That was not earned in this campaign — C2 is IIP, and every
other non-weakening probe returned either SGO-consistent or PE-level
verdicts without executing a decisive-negative attack in the sense of
the protocol.

Additionally, **C6 produces a weakening** on the general reading of the
decoupling (the "no joint-crossing mechanism can exist" reading) even
while the local reading (the two corpus members do exhibit different
obstructions) survives as observation.

**Disposition:** the campaign closes at state **`challenged`** (per
the lifecycle in `docs/invariant-challenge.md` §The lifecycle: an
"all-inconclusive campaign does not" earn surviving; the state remains
resumable). N2a-child-1 is **not admitted** as a surviving candidate
at this time. It is not falsified.

**Boundary-delta produced:** the decoupling admits a two-reading
split that this campaign did not fold into a single predicate:

```text
N2a-child-1 (local reading, admissible under this record):
  The two mechanisms es-05 and es-10 exhibit different sub-obstructions.
  (Trivially SGO at n = 2.)

N2a-child-1 (general reading, weakened by C6):
  No single mechanism can jointly cross both obstructions.
  (Under-tested at the corpus level; C7 causal argument is PE, not CMA.)
```

The two readings are not interchangeable, and the campaign's C6
weakening applies only to the general reading.

---

## Coverage note (following the N2a/N5 discipline)

**Probes not fully discharged, honest count:**

- **C2** (synthetic-counterexample construction) is IIP. No completed
  construction was produced; the PE describing the shape of a joint
  crossing does not substitute for the construction under the protocol.
- **C7** (causal argument) is PE, not CMA. Independent mathematical
  verification of the two obstructions' structural independence is not
  achieved.

Two of seven probes are therefore not fully discharged, mirroring the
C2/C7 pattern in the parent N2a challenge (per its attributed
reassessment, 2026-09-11). This is a **recorded challenge with
incomplete coverage plus one weakening landing**, not a survival across
all seven probes.

---

## Search-policy implications (proposed, not yet applied)

If the policy layer chooses to act on this record, three moves are
suggested by what the campaign found:

1. **Do not promote N2a-child-1 to `surviving`.** Retain it at
   `challenged`, resumable.
2. **Prefer frontier proposals that attempt the joint crossing.** The
   most information-dense next research move against N2a and
   N2a-child-1 is exactly the mechanism C2 could not construct: a
   proposal that exhibits polynomial main-term growth *and* covers or
   avoids QR-survivor classes. Under the frontier-eligibility rule
   (`internal/pipeline/frontier.go`), only `surviving` or
   `operator_attested` invariants influence proposal generation — this
   preference must be recorded on N2a (which is `surviving` with a
   recorded C4 boundary-delta), not on N2a-child-1, until the latter's
   state changes.
3. **Do not treat the C6 weakening as evidence that the decoupling is
   false.** Weakening reduces the epistemic authority of the general
   reading; it does not falsify the local reading.

Applying (2) — i.e. actually generating a frontier proposal against the
decoupling — is a separate, subsequent typed operation (`newf frontier
generate`) whose eligibility set is limited by the code-owned
`internal/pipeline/frontier.go` gate. This record proposes the search
direction; it does not perform the proposal step.

---

## Provenance

- Parent record: `corpus/experiments/pilot-004-discovery/records/N2a-challenge.md`
  (in particular §Derived child hypothesis, lines 297–323, where
  N2a-child-1 was first stated with its cheapest-falsification path).
- Corpus atlas members referenced: es-01, es-02, es-03, es-05, es-06,
  es-07, es-08, es-10 (as summarised in this project's Erdős–Straus
  atlas; each has a frozen note under `corpus/`).
- Related epistemic discipline: `docs/invariant-challenge.md` §The
  lifecycle and §Challenge types; AGENTS.md §Candidate invariant
  discipline.
- Sibling records with the same protocol: `N2a-challenge.md`,
  `N5-challenge.md` (both with attributed reassessments recording
  IIP/PE downgrades; this record applies those labels inline from the
  start rather than appending them post-hoc).
- Author process: model-assisted, single-operator recorded. Same
  same-model-family caveat as the parent challenges (tier-4/5, per the
  manuscript's Limitations §12 and Experiment D limitation added
  2026-09-11 at commit `db4cb6e`). This is not independent human
  verification.
