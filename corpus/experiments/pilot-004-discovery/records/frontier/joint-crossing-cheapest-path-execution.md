# Cheapest-path execution — joint-crossing proposal falsified at step 2

Executes the predeclared cheapest-falsification-path for the joint-crossing
frontier proposal recorded at commit
[`6ce385c`](https://github.com/instagrim-dev/gow/commit/6ce385c). Under the
predeclared decision rule (fixed *before* executing any step), the verdict is
**`falsified_by_cheapest_path`**.

Recorded: 2026-09-12. Executor: model-assisted with the checked algebra below.
The algebra is **CMA-tier** (Checked Mathematical Argument): the identities
below are elementary and directly verifiable, not model judgment.

Epistemic-tier legend (unchanged from the parent records):

```text
SGO  Source-grounded observation
PE   Proposed explanation
CCS  Completed counterexample search
IIP  Inapplicable or inconclusive probe
CMA  Checked mathematical argument
```

---

## What was executed

Predeclared cheapest-path from
[`N2a-joint-crossing-proposal.md`](./N2a-joint-crossing-proposal.md)
§Cheapest falsification path:

| Step | Test | Predeclared cost |
|---|---|---|
| 1 | Bombieri–Vinogradov admissibility of the operative modulus range | hours |
| 2 | **Composition-genus non-emptiness for `Q(u,v) = uv + λn(u+v)` over QR-survivors for small operative primes** | hours |
| 3 | Joint-indexing precondition (character-sum uniformity within the genus) | days |
| 4 | Small-case verification for `n ≤ 10^4` in QR-survivor classes | days |

Step 2 was executed first because it is the more decisive of the two "hours"
steps: Component B's entire covering apparatus depends on it, and its failure
mode is strong (the apparatus doesn't apply at all) whereas step 1's failure
mode is quantitative (the modulus range is too narrow).

---

## Step 2: computation

### Object under test

`Q(u, v) = uv + λn(u + v)` — the binary quadratic form the proposal named as
its reciprocity-carrying covering apparatus for Component B.

### Homogeneous part

As a form `au² + buv + cv²` the homogeneous part of Q has
`(a, b, c) = (0, 1, 0)`. Its discriminant is:

```
D = b² − 4ac = 1² − 4·0·0 = 1                    (CMA)
```

### Reduction to normal form (linear substitution)

Substitute `u = X − λn`, `v = Y − λn`:

```
Q(u, v) = uv + λn(u + v)
        = (X − λn)(Y − λn) + λn(X − λn) + λn(Y − λn)
        = XY − λnX − λnY + λ²n²
                + λnX − λ²n²
                + λnY − λ²n²
        = XY − λ²n²                                (CMA)
```

So `Q(u, v) = m` is equivalent to `XY = m + λ²n²` where `X = u + λn` and
`Y = v + λn`. Q is (up to an integer translation) the split hyperbolic form
`XY`.

### Class-group / genus-theory implications of `D = 1`

Standard results in the classical theory of binary quadratic forms (Gauss;
see e.g. Cox, *Primes of the form x² + ny²*, chs. 2–3):

1. `D = 1` is a **perfect-square discriminant**. A discriminant is
   *fundamental* only when it is not a non-trivial square (with additional
   4-divisibility conditions). `D = 1` is degenerate.

2. `h(1) = 1` (one class of forms of discriminant 1, the split form `XY`).

3. **`g(1) = 1` — one genus.** The genus contains the unique class.

4. The **generic characters** that discriminate genera — the reciprocity-
   indexed characters `(·/p)` for primes `p | D` — are indexed by the set of
   primes dividing `D`. With `D = 1`, this set is empty. **There are zero
   reciprocity characters attached to Q's genus.**

### Representability by Q

`XY = m + λ²n²` has integer solutions `(X, Y) ∈ ℤ²` for every integer `m`:
take `X = 1`, `Y = m + λ²n²`. Q represents every integer without any
congruence or reciprocity restriction.

### Verdict on Component B (proposal quote)

The frontier proposal
([`N2a-joint-crossing-proposal.md`](./N2a-joint-crossing-proposal.md),
§The proposal → Component B) stated:

> The composition genus of `Q(u, v) = uv + λn(u + v)` indexes classes by
> reciprocity symbols and covers the QR-survivor congruence classes natively.

Both claims fail against the object as literally named:

- **"indexes classes by reciprocity symbols":** false — Q has one class and
  zero reciprocity-indexed generic characters (discriminant 1).
- **"covers the QR-survivor congruence classes natively":** vacuously true in
  the wrong direction — Q represents *every* integer, so it does not
  *selectively cover* QR-survivors. It covers everything indiscriminately,
  which is not the covering apparatus Component B needs. A covering
  apparatus must *discriminate* — that is what makes it inject reciprocity
  into the counting.

**Component B does not do what the proposal claimed. This is a negative at
cheapest-path step 2.**

---

## Step 1: n/a-given-step-2

Component A's Bombieri–Vinogradov admissibility check was scoped to establish
uniformity *over Component B's composition genus*. Component B has no
non-trivial composition genus in the reciprocity-carrying sense — the object
it was supposed to average over does not exist as described. Step 1 becomes
vacuous *for this specific proposal*.

If a rescued proposal supplied a genuinely non-degenerate reciprocity-carrying
form (see §Rescue attempts below), step 1 would have to be re-scoped against
that form's actual genus structure.

---

## Steps 3–4: n/a-given-step-2

Steps 3 (joint-indexing precondition) and 4 (small-case verification) were
predicated on Component B's apparatus existing as stated. They do not have
an object to run against.

---

## Verdict under the predeclared decision rule

Predeclared rule from
[`N2a-joint-crossing-proposal.md`](./N2a-joint-crossing-proposal.md)
§Predeclared decision rule:

> `falsified_by_cheapest_path`: the cheapest falsification path is executed
> and returns a negative — either the polynomial main term does not exist
> under the proposed apparatus, or the QR-covering step does not extend to
> the QR-survivor classes, or the two steps cannot be composed without one
> of them destroying the other's precondition.

Step 2 returned a strong negative on the middle clause: **the QR-covering
step does not extend to the QR-survivor classes because it does not
discriminate any classes at all.** The proposal is
**`falsified_by_cheapest_path`**.

Frontier proposal state:

```text
admitted                  ✅ (wire-admissible; commit 6ce385c preflight passed)
violates_target           N/A (no live corpus; not evaluated by internal/frontier)
falsified_by_cheapest_path ✅ THIS VERDICT
```

---

## Rescue attempts — what would a fixed proposal look like?

Honest scope: the falsification lands on the **specific `Q` named in the
proposal**, not on the *class* of possible joint-crossing mechanisms. A
rescued proposal must supply a different reciprocity-carrying object.

### Candidate rescue: `Q'(u, v) = u² + nv²`

Discriminant `−4n`. For `n` squarefree and appropriate residue mod 4, the
form class group of `Q'` has non-trivial genus theory:

- Number of generic characters: `ω(n) + ε` (where `ε ∈ {0, 1, 2}` depends
  on `n mod 4`).
- Number of genera: `2^{ω(n)+ε−1}`.
- Genera are indexed by tuples of Legendre symbols `(m/p)` for `p | n` —
  genuinely reciprocity-carrying.

`Q'(u, v) = u² + nv²` **would** be a covering apparatus that indexes
classes by reciprocity symbols. This is the machinery classical composition-
genus theory attaches to.

**But this is a different proposal.** Specifically:

- The ES equation `4/n = 1/x + 1/y + 1/z` does not naturally express `x`,
  `y`, or `z` as values of `u² + nv²`. After clearing denominators, the ES
  equation is `4xyz = n(xy + xz + yz)` — a Diophantine surface, not a value
  of a binary quadratic form in the natural variables.
- Even if some substitution `x = f(u, v)` could be forced, the *joint-
  indexing move* — choosing the δ-method modulus parameter within `Q'`'s
  genus over the operative small primes — becomes structurally different
  because `Q'`'s genus is indexed differently.

A rescued proposal built around `Q'` would need to:

1. Establish a substitution that plausibly connects `Q'`'s representation
   to the ES linear form. (Not obvious; not known to exist.)
2. Re-do the δ-method / Bombieri–Vinogradov admissibility check against
   `Q'`'s actual modulus range (which involves the discriminant `−4n`, not
   an arbitrary parameter).
3. Verify the joint-indexing precondition against `Q'`'s specific genus
   structure.

Under AGENTS.md §Mechanistic novelty over surface novelty, that is a new
proposal, not a repair of this one. **The falsification of the original
proposal stands.**

### What the rescue attempt *demonstrates*

The rescue exercise sharpens the residual research question:

> Any joint-crossing mechanism composing an averaged counting identity with
> a reciprocity-carrying covering must supply a **non-degenerate**
> reciprocity carrier — a binary quadratic form of non-square fundamental
> discriminant, or an equivalent object with non-trivial genus theory —
> *and* an argument that this object plausibly meshes with the ES linear
> form.

The falsified proposal supplied neither: its `Q` was degenerate, and its
joint-indexing move was scoped to a genus that does not exist.

---

## Consequences

### For the frontier proposal

- **State:** `falsified_by_cheapest_path` (persisted). The proposal remains
  in the durable record; its wire payload, preflight log, and this
  execution record together form a complete failed-attempt artifact per
  AGENTS.md §Failure is a first-class artifact.
- **What was learned:** the specific compositional move
  "polynomial-rate δ-method identity × composition-genus covering of
  `uv + λn(u+v)`" cannot work because the covering component is
  reciprocity-trivial.
- **Redundancy warning:** any future frontier proposal that names a
  discriminant-1 binary quadratic form as its reciprocity carrier is
  redundant with this record and should be rejected at the
  novelty-argument step.

### For N2a-child-1 (the parent challenge's derived hypothesis)

Under the search-policy implications predeclared at
[`N2a-joint-crossing-proposal.md`](./N2a-joint-crossing-proposal.md)
§Search-policy implications:

> If the proposal reaches `falsified_by_cheapest_path` at step 1 or 2,
> **promote N2a-child-1's general reading** by one weakening-tier — the
> "no joint-crossing mechanism can exist" claim gains a falsified attempt,
> weakening the C6 landing.

**Applied here:** N2a-child-1's C6 weakening on the general reading
("no single mechanism can jointly cross both obstructions") is now
*partially rehabilitated*. The rehabilitation is bounded:

- **What the falsification does establish (CMA):** the *specific*
  discriminant-1 attempt does not achieve joint crossing.
- **What the falsification does not establish:** that *no* joint-crossing
  mechanism exists. One attempt has been falsified; the rescue analysis
  above shows a non-degenerate reciprocity carrier is a *necessary*
  condition but is not itself a joint-crossing mechanism.
- **Net movement:** N2a-child-1's general reading moves from "no
  attempted construction; C6 weakens" to "one attempted construction,
  falsified at its reciprocity-carrier step." This is *weak* evidence for
  the general reading, of a form C6 explicitly asked for.

### New derived hypothesis (proposed, entering `proposed` independently)

```text
N2a-child-2 (proposed 2026-09-12):
  Any joint-crossing mechanism against N2a's C4 boundary-delta requires a
  non-degenerate reciprocity carrier — i.e. a binary quadratic form of
  non-square fundamental discriminant with non-trivial genus theory, or
  an equivalent object.

  Cheapest falsification: exhibit a joint-crossing mechanism whose
  reciprocity-carrying apparatus has trivial genus theory. (This
  falsification is itself constructive: it would falsify by producing a
  mechanism the current record cannot construct.)
```

This is `proposed`, not `surviving`. Its own challenge is not run here.

---

## Provenance

- Parent proposal record:
  [`N2a-joint-crossing-proposal.md`](./N2a-joint-crossing-proposal.md).
- Wire payload:
  [`joint-crossing-proposal.wire.json`](./joint-crossing-proposal.wire.json).
- Preflight log: [`preflight.log`](./preflight.log).
- Predeclared cheapest-path steps: parent record §Cheapest falsification
  path, four ordered steps.
- Predeclared decision rule: parent record §Predeclared decision rule,
  fixed *before* any step execution.
- Predeclared search-policy implications: parent record §Search-policy
  implications, three conditional moves.
- Classical genus-theory reference (external, not verified here — cited at
  literature level): Cox, *Primes of the form x² + ny²*, chs. 2–3.

Executor caveat: this record is model-assisted. The algebra
(discriminant computation, linear substitution, representability check) is
elementary and directly verifiable — an independent reader can reproduce
every line of it from the definitions above in a few minutes. That places
this record at CMA-tier for the computational core. The identifications
of *what classical genus theory says about D=1* are at literature level
and should be re-checked by any operator who wants independent
confirmation before acting on the search-policy implications.
