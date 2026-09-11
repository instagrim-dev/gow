# N2a challenge record — intrinsic density/averaging ceiling

Candidate invariant: **N2a**
Recorded: 2026-09-11
Source occurrences: E13 (D0, d0-run3), E16 (D0, d0-run1), E20 (D1, d1-run3), E23 (D0, d0-run2)
All four unanimous (3-0) in quorum.

---

## Canonical statement (consolidated from four occurrences)

> **Density/averaging approaches produce a quantitative almost-all statement
> whose gap to "every n" is a structural feature of the method — not a
> shortfall of effort — because the method's rate of improvement is
> intrinsically incapable of driving the exceptional set to empty.**

Scope: **es-05** (averaging via Bombieri–Vinogradov) and **es-10** (Vaughan
congruence density). Both `partial_failure`.

Preserved distinctions:
- es-05 is probabilistic, global, bounds a *solution-count average* over a
  prime ensemble via sieve inequalities (Bombieri–Vinogradov,
  Brun–Titchmarsh, Turán–Kubilius).
- es-10 is deterministic, local-to-global, bounds the *density of unsolved n*
  by assembling identity-bearing congruence classes.
- They cap out for different structural reasons: es-05 because polylog average
  growth cannot force existence; es-10 because the QR-survivor classes are
  reciprocity-protected and no finite battery of congruence classes can touch
  them.

The shared property is the *shape of the boundary* (structural ceiling on
the rate of improvement), not the shared operator.

---

## Challenge execution

Per AGENTS.md §Candidate invariant discipline: seven challenge probes applied.

---

### C1 — Known failed approach that violates the property

**Test:** Is there a known failed approach in the corpus whose mechanism has
the density/averaging structure but whose boundary is *not* a structural
ceiling — i.e., the failure was due to effort or implementation, not method
shape?

**Result: no counterexample found.**

All other failure-side notes have different mechanism families:
- es-01, es-02, es-03 fail due to QR confinement — a different structural
  obstruction, not a rate ceiling.
- es-07 fails due to finite verification range — a generality failure, not a
  rate ceiling. The failure is structural (finite cannot certify universal)
  but the structure is different from a rate-limited convergence argument.
- es-08 is a meta-level obstruction, not a density approach.

No corpus note records a density/averaging approach that failed for a
non-structural reason while using the same rate-limited machinery.
**C1: property survives.**

---

### C2 — Synthetic failed approach that violates the property

**Test:** Can we construct a plausible density/averaging approach on the
Erdős–Straus problem that does NOT have a structural ceiling — i.e., one
whose average growth rate is sufficient to force universal existence?

**Result: the construction is exactly what the candidate identifies as the
missing move.**

The es-05 note states explicitly: "For some Diophantine problems a positive
average forces a solution to exist, but that argument needs *at least*
polynomial growth." If a hypothetical averaging argument achieved polynomial
(not polylog) growth of `Σ_{p≤N} f(p)`, that would cross the existence
threshold. Such an argument would be an averaging approach without the
structural ceiling.

However: no such argument exists in the corpus, and the note frames polynomial
growth as the required upgrade that the method cannot supply from within its
own machinery. A synthetic "averaging approach with polynomial average growth"
would need a fundamentally different sieve mechanism — one that the
failure-space shows has not been constructed.

This synthetic case is therefore the *target* the candidate invariant
identifies as beyond the ceiling — not a counterexample to it.
**C2: no falsifying synthetic construction exists; the synthetic target
reconfirms the property.**

---

### C3 — Success that preserves the property

**Test:** Is there a successful (or partial-success) approach in the corpus
that *also* has the density/averaging structure, shows the same rate-limited
almost-all ceiling, and yet crosses the survivor classes?

**Critical contrast case: es-06 (partial_success)**

es-06 reduces `4/n = 1/x+1/y+1/z` to a two-fraction representation question
and derives lower bounds `f(n) > 0` for almost all primes. Its boundary: "the
representation bounds [...] degrade exactly for primes that 'resemble a
perfect square' to small moduli — the quadratic-residue primes."

Two questions:
1. Does es-06 have the density/averaging structure?
2. Does es-06 have the *structural ceiling* the candidate identifies?

On (1): es-06 uses "reduction to two fractions" and "representation counting"
— it counts representations across ranges and derives distributional lower
bounds. This is *close* to the density/averaging family but not identical:
es-06 reasons about divisor structure of `4/n − 1/x` per residue class
rather than averaging a solution-count function via sieve. It has existential,
global character but its machinery differs.

On (2): es-06's boundary does degrade at QR primes, but the note does NOT
attribute this to an intrinsic rate limit of its method. The boundary is the
QR wall (same as L1's scope), not a "rate of improvement too slow to converge."
es-06 does not carry the internal ceiling claim: it does not say the gap to
"all n" is because its bound grows too slowly, but rather that divisor
structure fails at QR primes.

**Assessment:** es-06 is a genuine contrast but does not falsify N2a. es-06
reaches almost-all by a different mechanism and its ceiling is the QR wall,
not a rate-limiting structural ceiling internal to an averaging/density
argument. The candidate's distinction between "rate-limited convergence
argument" and "QR-wall bounded method" is real and preserved.
**C3: property survives; contrast correctly distinguishes boundary type.**

---

### C4 — Lower the abstraction level: does the property split?

**Test:** At a more concrete level, are es-05 and es-10 failing for the same
reason or for different reasons that happen to share an abstract label?

**es-05 at concrete level:**
The polylog growth barrier is specific: `Σ f(p) ~ C (log N)^k` for some k,
and the threshold for forcing existence is `Σ f(p) ≫ N^ε` for some ε > 0. The
gap is a specific asymptotic statement about growth rates of averages, and it
is intrinsic to the sieve machinery (the Bombieri–Vinogradov theorem gives
smooth error bounds, not polynomial main terms for solution counts).

**es-10 at concrete level:**
The congruence density bound gives `|{n ≤ N : n unsolved}| ≤ N · exp(−c
(log N)^{2/3})`. This decays to zero as N → ∞ but never reaches zero for any
finite N, because the QR-protected classes always survive. The "intrinsic
ceiling" here is the QR wall: the exceptional set cannot empty because
reciprocity prevents the congruence battery from covering QR classes.

**Split analysis:**
At concrete level, the two ceilings are *mechanistically different*:
- es-05's ceiling: rate of the averaging bound is too slow (polylog < polynomial
  threshold for existence forcing).
- es-10's ceiling: the exceptional set cannot reach zero because it is
  *structurally protected by QR reciprocity*, not because the density
  converges too slowly toward zero.

This reveals a **partial split at the concrete level**: the "intrinsic
structural ceiling" abstracts over two distinct mechanisms. The abstract
property survives as a description of *both*, but the concrete failure
structure is different:
- es-05 fails because *the method's quantity converges too slowly to cross
  the existence threshold*.
- es-10 fails because *the method's quantity cannot reach the target value
  (zero) regardless of how many steps are taken, due to an external obstruction*.

The candidate is tighter and more accurate when stated as:
> **Each approach produces a monotone-improving almost-all statement that
> cannot reach universality: es-05 because the growth rate is subthreshold;
> es-10 because the QR wall provides an absolute floor.**

**C4 result: the property survives but splits at one level. Recommend
recording the two mechanisms distinctly rather than collapsing them.**

---

### C5 — Raise the abstraction level: do several invariants merge?

**Test:** At a higher level, does N2a merge with L1 (QR confinement) or
with the N3 locality candidate?

**Against L1 (confined_to_quadratic_nonresidues):**
L1 is about mechanism confinement — the congruence-carried reach cannot touch
QR classes. N2a is about measurement ceiling — the rate of improvement cannot
cross a threshold. These are distinct: es-05 is NOT in L1's scope (its
boundary is the growth rate, not QR confinement), and es-10's N2a aspect
(density never reaches zero) overlaps with L1 but the *reason* is different
from the reason L1 applies to es-01/es-02/es-03/es-08. At higher abstraction,
one could say both are "approaches that cannot cover the QR survivors," but
that loses the mechanistic distinction (confinement of reach vs. rate of
convergence). The merge is defective — it erases predictive information.

**Against N3 (local-only reasoning):**
N2a's members are global (es-05 probabilistic-global, es-10 local-to-global).
L3 and the locality family are local/per-class. No merger.

**C5 result: N2a does not merge with existing reference properties or other
novel candidates at higher abstraction. The property is stable.**

---

### C6 — Sampling and publication bias

**Test:** Does the corpus over-represent approaches with this ceiling because
the corpus was curated to contain well-characterised failed partial approaches?

**Assessment:**
The corpus was curated to contain documented pre-cutoff approaches to the
Erdős–Straus conjecture. Density/averaging approaches appear in the literature
precisely because they are a natural analytic-number-theory strategy. The
property is attested by two independent research groups (es-05: Elsholtz–Tao
2013; es-10: Vaughan 1970 and successors). The ceiling observation in both
notes is a mathematical statement about the method's asymptotic behaviour, not
a subjective annotation.

**C6 result: no meaningful sampling bias identified. Both members are
independently published; the ceiling observation is internally derived.**

---

### C7 — Correlation versus causal obstruction

**Test:** Is N2a observational (these approaches happen to have an almost-all
ceiling) or causal (the ceiling is structurally necessitated by the method's
machinery)?

**es-05:** The causal argument is explicit in the note: the growth rate is
polylogarithmic by the sieve machinery, and the threshold for existence-forcing
is polynomial. The gap is not accidental; it follows from the specific tools
used (Bombieri–Vinogradov gives smooth error bounds, not existence-forcing main
terms). The ceiling is mechanistically explained.

**es-10:** The causal argument is also explicit: the exceptional set cannot
reach zero because the QR-survivor classes are reciprocity-protected. Vaughan's
bound improves the density of the exceptional set but the lower bound on its
size is set by the QR obstruction. The ceiling follows from an external
structural constraint, not from convergence speed.

**C7 result: both ceilings are causally explained within their notes, not
merely observed. The candidate is mechanistically grounded for both members,
though the mechanisms differ (C4).**

---

## Challenge summary

| Probe | Result |
|---|---|
| C1: known failed approach violating property | No counterexample — property survives |
| C2: synthetic failed approach violating property | Synthetic case is the target the property names — reconfirms |
| C3: success preserving the property | es-06 contrast is genuine; es-06's ceiling is QR-wall, not rate-limited — property survives with distinction intact |
| C4: lower abstraction — splits? | **Partial split:** es-05 ceiling = rate subthreshold; es-10 ceiling = QR absolute floor. Abstract property survives; concrete mechanisms differ |
| C5: higher abstraction — merges? | Does not merge with L1 or N3 — stable |
| C6: sampling bias | None identified — two independent publications |
| C7: correlation vs causal | Both ceilings causally explained within notes |

**Verdict: property survives challenge with one refinement required.**

---

## Refinement required before admission

The abstract statement "intrinsic structural ceiling" is accurate but
conflates two mechanistically different ceilings. The admitted property should
record the distinction:

> **N2a (refined): Density/averaging approaches produce a monotone-improving
> almost-all statement whose gap to "every n" cannot be closed within the
> method's own machinery: es-05 because the sieve-average growth rate is
> polylogarithmic and therefore subthreshold for existence-forcing; es-10
> because the QR-survivor classes provide an absolute floor that the
> congruence battery cannot remove (the QR obstruction, independently named
> by L1).**

This preserved-distinctions note makes the property more informative than the
generic "structural ceiling" label: a future proposal that achieves polynomial
average growth (crossing es-05's barrier) would not be falsified by N2a; a
proposal that crosses the QR wall (crossing es-10's barrier, outside L1's
scope) would falsify both N2a and L1.

The scope remains [es-05, es-10]. Separate admission records for the two
sub-mechanisms are not required, but the refinement should be reflected in the
interpretation claim if admitted.

---

## Admission recommendation

**Recommended for admission** as an interpretation claim under the refined
statement above. Suggested canonical ID: `density_averaging_ceiling`.

Admission requires:
- Operator attestation of the refined statement
- New pinned vocabulary revision (mechanism/v4 or similar)
- No changes to original corpus notes
- Reclustering + fresh database before support is counted
- The two sub-mechanisms (rate-limited vs QR-floor) recorded as scope notes,
  not as separate claims

Provenance: `GeneratedInterpretation`, source occurrences E13/E16/E20/E23 from
pilot-004-discovery quorum, challenge record this file.
