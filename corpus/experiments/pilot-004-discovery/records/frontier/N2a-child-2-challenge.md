# N2a-child-2 challenge record — non-degenerate reciprocity carrier requirement

Candidate: **N2a-child-2** (proposed 2026-09-12 in the joint-crossing
cheapest-path execution record; entered `proposed` independently as a
derived hypothesis; requires its own challenge to advance).

Recorded: 2026-09-12.
Parent: derived from
[`joint-crossing-cheapest-path-execution.md`](./joint-crossing-cheapest-path-execution.md)
§Consequences → §New derived hypothesis. Not derived from N2a directly.
Sibling: N2a-child-1
([`../N2a-child-1-challenge.md`](../N2a-child-1-challenge.md)),
`challenged` state.
Canonical ID proposal: `joint_crossing_requires_nondegenerate_reciprocity_carrier`.

Predeclared decision rule (unchanged from
[`../N2a-child-1-challenge.md`](../N2a-child-1-challenge.md)):

- `surviving` iff ≥1 probe produces a *completed decisive negative* over a
  *nonempty* eligible population, and no probe forces `weaken`/`split`/`falsified`.
- `weakened` iff any probe produces a code-verifiable weakening.
- `split` iff C4 produces ≥2 admissible child predicates each grounding to
  nonempty pairwise-disjoint support.
- `falsified` iff C1 finds an in-corpus counterexample OR C2 produces a
  *completed* synthetic construction refuting the claim.
- `challenged` (resumable, not surviving) iff the campaign is entirely
  IIP/PE — no completed decisive negative earned.

Epistemic-tier legend (SGO / PE / CCS / IIP / CMA) inherited from parent
records.

---

## Canonical statement (from the falsification execution record)

> **N2a-child-2**: Any joint-crossing mechanism against N2a's C4
> boundary-delta requires a non-degenerate reciprocity carrier — i.e. a
> binary quadratic form of non-square fundamental discriminant with
> non-trivial genus theory, or an equivalent object.

Scope: mechanisms that jointly achieve polynomial main-term growth in the
averaged solution count *and* discriminatively cover the QR-survivor
congruence classes for the ES equation.

Predeclared falsification path: exhibit a joint-crossing mechanism whose
reciprocity-carrying apparatus has trivial genus theory, *or* whose
discrimination among QR-survivor classes proceeds through no
reciprocity-adjacent object at all.

---

## Formulation health check (before running probes)

Under AGENTS.md §Abstraction safety: an abstraction must state what it
preserves, admit falsification, and produce concrete predictions. This
claim has one identifiable sharpness problem *up front*:

> **The "or an equivalent object" clause is a movable goalpost.**

What counts as "equivalent to" a non-degenerate reciprocity carrier is
not fixed. Candidates that plausibly satisfy the clause:

- Classical binary quadratic forms of fundamental discriminant with
  `h(D) > 1` (the paradigm).
- Ideal class groups of imaginary quadratic fields `ℚ(√-n)` (essentially
  in bijection with the paradigm).
- Dirichlet L-functions `L(s, χ_D)` for the quadratic character `χ_D`
  (encoding the same reciprocity information analytically).
- Theta series and modular forms with quadratic-character nebentypus.
- Cusp forms of weight 2 with real quadratic Nebentypus (implicitly).

Candidates that would *not* satisfy the clause without stretching it:

- Additive-covering apparatuses (do not touch reciprocity).
- Ergodic/measure-theoretic methods (no direct reciprocity structure).
- Character-free additive combinatorics (Green–Tao–adjacent, though
  quadratic-character variants exist).

**Recorded up front:** the "or equivalent object" clause weakens the
claim's testability. This is registered as a **PE-tier ambiguity** that
C6 will fold into a weakening critique. The challenge proceeds against
the sharpest reading: reciprocity carriers include classical forms,
L-functions with quadratic characters, and their direct analogues; they
exclude additive-only apparatuses, ergodic methods, and character-free
additive combinatorics.

---

## Challenge execution

### C1 — Known corpus counterexample

**Test:** Is there a corpus mechanism that achieves joint crossing *and*
discriminates QR-survivor classes *without* a non-degenerate reciprocity
carrier?

**Corpus survey (SGO):**

- No atlas mechanism achieves joint crossing (this is the point of
  N2a's C4 boundary-delta — the two sub-obstructions are disjoint in
  the atlas). The eligible population is **empty**.

**C1: IIP.** No decisive negative earned; no counterexample possible.

---

### C2 — Synthetic counterexample

**Test:** Can we construct a joint-crossing mechanism whose apparatus
reaches QR-survivor classes without a non-degenerate reciprocity carrier?

**Attempted constructions (PE):**

- **Ergodic/measure-theoretic approaches to ES.** No such joint-crossing
  ergodic mechanism is known in the literature or in this project's
  records. Ergodic methods have been applied to Diophantine problems
  (Bourgain, Green–Tao) but not to the ES joint-crossing question.
- **Circle method with pure additive characters.** The Hardy–Littlewood
  circle method uses additive characters `e(αn)`; achieving discrimination
  between QR-classes via *purely additive* character sums is not known
  to be possible. Character-sum arguments that reach QR-classes typically
  factor through the Pólya–Vinogradov inequality, which is a quadratic-
  character statement — i.e. a reciprocity carrier by another name.
- **Direct construction restricted to QR-survivor classes.** Choosing
  parameters `(x, y, z)` that lie in QR classes by direct construction
  requires *knowing* which `n` are QR-survivors, and that fact is defined
  by the Legendre symbol. Any systematic construction that discriminates
  by QR-ness *is* reciprocity work under a different label. This is a
  near-tautological observation, not a construction that falsifies the
  claim.

**Bounded structural argument (PE, not CMA):** the QR-survivor set is
*defined* by quadratic reciprocity at operative primes. Any apparatus
that systematically enumerates, samples, or covers the QR-survivor set
must, at some layer, evaluate Legendre symbols — either explicitly
(reciprocity carrier), or implicitly through an object whose theory is
reciprocity-adjacent (L-functions, class groups, quadratic-character
sums). This tension makes the claim *near-tautological* under its
sharpest reading, which is itself a challenge signal (see C7).

**C2: no completed construction (IIP + PE argument for near-tautology).**
No decisive negative; the *shape* of the claim starts to look definitional
under the sharpest reading.

---

### C3 — Partial-success (or success) preserving the property

**Test:** Does the corpus's partial success, es-06, preserve or violate
the claim?

**Analysis of es-06 (SGO):**

es-06 uses divisor-structure representation counting. Its lower bound
`f(n) > 0 for almost all primes` degrades **exactly at QR primes** —
this is recorded in
[`../N2a-challenge.md`](../N2a-challenge.md) lines 100–105.

es-06 uses no reciprocity carrier: its covering is by divisor structure,
which is additive/multiplicative in the divisor sense, not
reciprocity-carrying. And es-06 *fails* to cover QR-survivors — exactly
as N2a-child-2 predicts a non-carrier apparatus would.

**es-06 is a positive attestation of N2a-child-2:** a mechanism without
a non-degenerate reciprocity carrier failed to cover QR-survivors, which
is what N2a-child-2 says would happen.

**C3: property survives (SGO). Strong support (one atlas partial success
attesting the claim's prediction).**

---

### C4 — Lower abstraction: does the claim split?

**Test:** Does "non-degenerate reciprocity carrier" decompose into
independent sub-conditions with distinct predictive content?

**Decomposition (PE):**

The condition decomposes into:

- **(a)** The carrier discriminates congruence classes at some operative
  prime.
- **(b)** The discrimination is by reciprocity (Legendre symbols at
  ramified primes), not by additive residues.
- **(c)** The carrier has non-trivial internal structure (`h(D) > 1`, or
  the L-function analogue).

**Independence check (PE):**

- (a) without (b): additive residue covering (es-01, es-08, es-10). These
  discriminate but not by reciprocity — and they attest the QR-survivor
  floor rather than crossing it. Consistent with the claim.
- (b) without (c): trivial-class-group quadratic forms (e.g. the falsified
  Q in the parent record). These discriminate by reciprocity but have no
  internal structure to spread the discrimination across a class group.
  Also consistent with the claim.
- (a), (b), and (c) together: the classical `u² + nv²` paradigm and its
  analogues. This is the paradigm the claim points at.

The three sub-conditions each independently obtain in the corpus, but
they do **not** produce pairwise-disjoint corpus support: no atlas member
attests condition (b) alone or condition (c) alone as a joint-crossing
mechanism (because no atlas member achieves joint crossing). The
decomposition sharpens the claim but does not split it into two admissible
child predicates over disjoint support.

**C4: property survives at finer abstraction. Decomposition recorded, no
productive split into disjoint-support children (PE).**

---

### C5 — Higher abstraction: does the claim merge?

**Test:** Does N2a-child-2 merge with a broader invariant?

**Candidate merges:**

- **Broader principle: "obstructions of type X can only be overcome by
  machinery that natively speaks type X's language."** N2a-child-2 is
  the *reciprocity* instance of this broader principle. Merging would
  require attesting the principle at other types (locality, additivity,
  measure-theoretic) — which the corpus does not currently attest at
  scale. A premature merge would erase reciprocity-specific content.
- **With N2a itself:** N2a is a *ceiling* claim (the density-averaging
  approach cannot cross a growth threshold); N2a-child-2 is a
  *requirement* claim (any crossing requires a specific object). These
  are different logical shapes; merging would confuse them.
- **With L6 (distinct roles):** L6 is about role preservation across
  mechanisms; N2a-child-2 is about apparatus requirements for
  boundary-crossing. Different axes.

**C5: no productive merge available. Property is stable at its stated
abstraction level (SGO).**

---

### C6 — Sampling and publication bias

**Test:** Is the claim supported by the corpus at a strength appropriate
for its assertion?

**Support audit (SGO):**

Positive support:
- **es-06 (partial success):** exact fit for the claim's prediction —
  non-carrier apparatus, fails at QR-survivors.
- **es-01, es-08, es-10 (additive covering failures):** all consistent
  with the claim — non-carrier apparatuses that do not cross QR
  confinement.

Negative test:
- **The falsified frontier proposal (2026-09-12):** attempted a
  degenerate (discriminant-1) carrier and failed. This falsification
  is *one attempt* that failed at the reciprocity-carrier step — it
  attests the *necessity* of non-degeneracy, but does not test the
  *sufficiency* of non-degeneracy.

**Weakening landings identified:**

1. **The "or equivalent object" clause is a movable goalpost.** Under
   the sharpest reading (classical forms + L-functions with quadratic
   characters + direct analogues), the claim is testable. Under the
   loosest reading ("any object that discriminates QR-classes"), the
   claim becomes near-tautological because QR-classes are *defined* by
   reciprocity. This ambiguity is a genuine sharpness problem.
2. **The inductive base for necessity is n = 1 falsified attempt.** One
   falsified proposal is thin evidence for a universal requirement.
3. **The claim is at least partially definitional.** As noted in C2 and
   above: any apparatus that systematically discriminates QR-classes
   *must* evaluate Legendre symbols at some layer, either explicitly or
   through a reciprocity-adjacent object. The "requirement" may be
   built into the target set's definition rather than into the crossing
   mechanism.

**C6: real weakening landing (SGO on the count; PE on the definitional
reading).** This is a **weakening** signal on the general reading.

---

### C7 — Correlation vs causal obstruction

**Test:** Is the requirement of a non-degenerate reciprocity carrier
*causal* (reciprocity injection is structurally necessary to cross the
QR-survivor floor) or *correlational* (mechanisms that succeed happen
to use reciprocity carriers because that is the historically-developed
tool)?

**Causal-side argument (PE):**

The QR-survivor floor is *defined* by quadratic reciprocity: the survivors
are the classes for which `(4/p) = 1` at operative primes, or equivalent
Legendre-symbol conditions. Discriminating among these classes requires
information about `(⋅/p)`, and objects that carry this information
natively are the reciprocity carriers. This is a **near-tautological**
causal argument: the target set's *definition* forces the crossing
mechanism to interact with reciprocity.

**Correlational alternative (PE):**

The historical development of number theory has produced reciprocity
carriers as the *first* tool that discriminates QR-classes. It is
conceivable but not attested that a non-reciprocity-adjacent apparatus
could achieve the same discrimination through a mechanism not yet
categorised as reciprocity. (Ergodic-theoretic constructions on QR
classes, for example, are not attested but are not obviously impossible.)

**Assessment:**

The claim's causal core is **near-tautological**, which is itself a
challenge signal — a claim that cannot be falsified because its
target's definition forces the conclusion is not a *substantive*
invariant, it is a definitional restatement.

**C7: PE-strength causal argument, borderline tautological.** Not CMA.
Not a decisive negative, but a real challenge to the claim's *substantive
content* (as distinct from its truth).

---

## Challenge summary

| Probe | Verdict | Tier |
|---|---|---|
| C1 known counterexample | Eligible population empty | IIP |
| C2 synthetic counterexample | No completed construction; near-tautology observed | IIP + PE |
| C3 partial-success preservation | es-06 exactly attests the claim's prediction | **SGO — supports** |
| C4 lower abstraction | Three sub-conditions decompose; no disjoint-support split | PE |
| C5 higher abstraction | No productive merge | SGO |
| **C6 sampling bias** | **Weakening: n=1 falsified attempt; "equivalent object" is movable; partially definitional** | **SGO — weakens** |
| **C7 causal vs correlational** | **Near-tautological causal core; challenge to substantive content** | **PE — challenges** |

Completed decisive negatives: **0**.
Weakening landings: **2** (C6, C7).
Falsifications: **0**.
Splits: **0**.
Merges: **0**.
Positive support: **1 SGO landing** (C3, es-06 attestation).

---

## Disposition under the predeclared decision rule

`surviving` requires ≥1 completed decisive negative. That was not earned.
`falsified` requires a completed counterexample. Not earned.
`split` requires C4 to produce ≥2 admissible child predicates over
disjoint corpus support. Not earned.
`weakened` requires a code-verifiable weakening critique that shrinks
support or erases discrimination — **C6 and C7 both land here on the
general reading**.

**Disposition: `weakened`.** The claim is not surviving, not falsified.
Its formulation and its substantive content both take real hits:

- **Formulation:** the "or equivalent object" clause makes the claim
  slippery. Any operator acting on it should first pin the equivalence
  class (which is done informally here: classical forms + L-functions
  with quadratic characters + direct analogues; excluding additive-only,
  ergodic, and character-free additive combinatorics).
- **Substantive content:** the claim's causal core is near-tautological
  under its sharpest reading, because the QR-survivor set is *defined*
  by reciprocity. This is a challenge signal: a claim whose truth is
  built into its target set's definition contributes less to the search
  policy than a claim whose truth is a discovered structural fact.

C3 provides genuine positive support at SGO tier (one atlas partial
success attests the prediction), but this support is thin: es-06's
failure at QR primes is compatible with N2a-child-2 but is also directly
predicted by N2a itself. C3 is *consistent with* N2a-child-2, not
*discriminating for* it against N2a.

---

## Boundary-delta produced

The claim decomposes into a sharpness-refinement:

```text
N2a-child-2 (weakened, general reading):
  Any joint-crossing mechanism requires a non-degenerate reciprocity
  carrier or an equivalent object.
  (Movable-goalpost formulation; near-tautological under the sharpest
  reading; weak inductive base.)

N2a-child-2' (sharper, proposed):
  Any joint-crossing mechanism must, at some layer of its apparatus,
  evaluate quadratic-character information at the operative primes for
  the QR-survivor set. This layer may be:
    - a binary quadratic form of non-square fundamental discriminant
      (Gauss-composition machinery),
    - a Dirichlet L-function L(s, chi_D) or its analytic siblings,
    - an ideal class group of an imaginary quadratic field, or
    - a modular/theta object with quadratic-character nebentypus.
  Additive-only, character-free, or purely ergodic apparatuses are
  insufficient.
```

This sharper form (`N2a-child-2'`, proposed but not admitted) narrows
the claim into a form that is falsifiable by constructing a joint-
crossing mechanism whose reciprocity layer is *none* of the enumerated
types. Its own challenge is not run here.

---

## Search-policy implications (proposed, not applied)

Under the challenge protocol in
[`../../../../docs/invariant-challenge.md`](../../../../docs/invariant-challenge.md)
and AGENTS.md §Search-policy mutation:

1. **Do not admit N2a-child-2 to `surviving`.** Retain at `weakened`.
2. **Do not use N2a-child-2 to filter future frontier proposals**
   (weakened invariants do not influence policy per
   `internal/pipeline/frontier.go`'s targetable-states rule).
3. **Consider promoting `N2a-child-2'` to `proposed` for its own
   challenge.** This is a search-direction proposal, not a state
   change; the sharper form is falsifiable in a way the parent is not.
4. **Redundancy warning for future frontier proposals against N2a:**
   proposals that name a discriminant-1 form as their reciprocity
   carrier remain redundant with the joint-crossing execution record;
   proposals that name a *non-carrier* apparatus (additive-only,
   ergodic, character-free) should carry a separate novelty argument
   because they attack N2a-child-2' rather than the parent N2a.

Applying any of these is a separate typed policy-mutation operation.

---

## Provenance

- Parent (derived-from) record:
  [`joint-crossing-cheapest-path-execution.md`](./joint-crossing-cheapest-path-execution.md)
  §Consequences → §New derived hypothesis.
- Original frontier proposal:
  [`N2a-joint-crossing-proposal.md`](./N2a-joint-crossing-proposal.md).
- Sibling challenge record:
  [`../N2a-child-1-challenge.md`](../N2a-child-1-challenge.md).
- Atlas members referenced: es-01, es-06, es-08, es-10.
- Related epistemic discipline:
  [`../../../../docs/invariant-challenge.md`](../../../../docs/invariant-challenge.md)
  §The lifecycle and §Challenge types; AGENTS.md §Candidate invariant
  discipline; AGENTS.md §Abstraction safety (for the formulation health
  check).
- Executor caveat: this record is model-assisted. Its algebra (where
  relevant) is CMA; its literature identifications (what counts as a
  reciprocity carrier) are at literature level and should be re-checked
  by any operator acting on the search-policy implications. Same
  same-model-family caveat as the parent records.
