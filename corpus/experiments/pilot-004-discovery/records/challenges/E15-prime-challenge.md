# E15' challenge record — annotation-pattern claim (Theme N4)

Candidate: **E15'** (proposed 2026-09-12 in
[`../E15-correction.md`](../E15-correction.md); revised, separately-
attributed correction of the original E15 entry).

Recorded: 2026-09-12.
Parent: original E15 ledger entry
[`../../adjudication-ledger.json`](../../adjudication-ledger.json) lines
546–583 (unanimous 3-0 defensible_novel, historical record retained).
Canonical ID proposal: `annotation_pattern_empty_breaks_es_01_03_10`.

Predeclared decision rule (fixed *before* running any probe, inherited
unchanged from
[`../frontier/N2a-child-2-challenge.md`](../frontier/N2a-child-2-challenge.md)):

- `surviving` iff ≥1 completed decisive negative over a nonempty eligible
  population, AND no probe forces `weakened`/`split`/`falsified`.
- `weakened` iff any probe produces a code-verifiable weakening.
- `falsified` iff C1 finds an in-corpus counterexample OR C2 produces a
  completed synthetic construction refuting the claim.
- `challenged` (resumable) iff every probe returns IIP/PE.

Two populations named up front (S1 discipline at
[`f0c0710`](https://github.com/instagrim-dev/gow/commit/f0c0710)):

- **Discovery population:** the Pilot-004 quorum result at `78e73d2` —
  where E15's original observation was mined.
- **Assessment population:** the current frozen atlas at
  `corpus/train/es-*.md`, HEAD `859cbbc` — where the seven probes
  actually run.

Epistemic-tier legend (unchanged): SGO / PE / CCS / IIP / CMA.

---

## Canonical statement (from E15-correction.md)

> **E15'**: The corpus annotations for **es-01** (Mordell polynomial
> identities), **es-03** (factorization scheme), and **es-10** (Vaughan
> congruence density) each record no broken properties — their frozen
> mechanism records carry `"breaks": []`. This is a syntactic pattern
> in the annotation records, not a mechanism-level conservation claim.

Scope: three specific atlas members' `breaks` field content in the
frozen records. Deliberately narrower than the original E15.

Predeclared falsification: an update to the frozen es-01, es-03, or es-10
records that supplies a non-empty `breaks` field.

---

## Challenge execution

### C1 — Known corpus counterexample

**Test:** Is there any atlas note among the shared_by set (es-01, es-03,
es-10) whose `breaks` field is *not* `[]` at the assessment population's
HEAD?

**Check (SGO, line-anchored, assessment population `859cbbc`):**

- `corpus/train/es-01-mordell-polynomial-identities.md` line 40:
  `"breaks": []` ✅
- `corpus/train/es-03-factorization-scheme.md` line 39:
  `"breaks": []` ✅
- `corpus/train/es-10-vaughan-congruence-density.md` line 36:
  `"breaks": []` ✅

All three annotations verified against the assessment population. No
counterexample.

**C1: no counterexample found (SGO). Property survives C1.**

---

### C2 — Synthetic counterexample

**Test:** Can a synthetic case be constructed that satisfies the
claim's antecedent (member of the shared_by set) but violates the
consequent (non-empty `breaks`)?

**Attempt:** By design the claim is *about* the frozen annotation
records themselves. A "synthetic case" that violates the annotation-
scope claim would have to *be* a modification of the frozen records —
which is exactly what the predeclared falsification path enumerates.
The synthetic-counterexample probe therefore collapses onto C1 for a
claim whose scope is bounded to specific atlas file contents at a
specific HEAD.

This is a **feature of the narrower revised scope**, not a defect: E15
was deliberately narrowed from mechanism-level to annotation-level in
[`../E15-correction.md`](../E15-correction.md) precisely to eliminate
the space of unmoored "synthetic mechanisms" that plagued the original.

**C2: not independently applicable (IIP by scope).** No decisive
negative earned, but no work is dropped — the check that C2 would
perform is subsumed by C1 for annotation-scope claims.

---

### C3 — Partial-success (or success) preserving the property

**Test:** Is there an atlas member with `outcome.class = partial_success`
that ALSO carries `breaks: []`? Such a case would preserve the property
across the outcome-class boundary, showing the annotation pattern is
not discriminating.

**Assessment population check (SGO, line-anchored):**

- `corpus/train/es-06-two-fraction-representation-bounds.md`:
  `outcome.class = partial_success` (line 43), `breaks: []` (line 35).
- `corpus/train/es-11-monks-velingker-structure.md`:
  `outcome.class = partial_success` (line 43), `breaks: []` (line 35).

**Two atlas members preserve the property across the partial_success
class boundary.** These are precisely the members that falsified E15's
*original* contrast_check overreach (defect 3 in the correction record).

C3 in the *original* E15's spirit would have expected the property to
NOT be preserved across the success boundary (that was the whole
substantive contrast). Under E15's *revised, annotation-scope* framing,
this is not a falsification — the revised claim explicitly does not
extend to a discrimination between outcome classes. But it IS a
**weakening**: the revised claim's inductive support relies on three
`partial_failure` observations, and the same field content occurs in
two `partial_success` records without changing outcome — the field is
weakly informative about outcome_class.

**C3: property preserved across outcome-class boundary. Under the
revised scope, this is a weakening landing (annotation pattern does
not discriminate between success and failure classes), not a
falsification. (SGO — weakens.)**

---

### C4 — Lower abstraction: does the claim split?

**Test:** Does the annotation-pattern claim decompose into sub-claims
with independent evidentiary bases?

**Decomposition (PE):**

- **(a)** es-01's `breaks: []` (Mordell polynomial identities).
- **(b)** es-03's `breaks: []` (factorization scheme).
- **(c)** es-10's `breaks: []` (Vaughan congruence density).

Each sub-claim is independently attested by its own annotation. The
three atlas members share no operator, no representation, and no
mathematical mechanism family in the frozen records — they are three
distinct approaches that happen to share the annotation-field-emptiness
property.

Do (a), (b), (c) each ground to nonempty pairwise-disjoint support?
Yes — each is one atlas note, three disjoint notes. But this
"decomposition" is trivial: the three-way conjunction that E15' claims
is literally the AND of three independent single-note claims. Under
`docs/invariant-challenge.md`'s split criterion (≥2 admissible child
predicates each grounding to nonempty pairwise-disjoint support), the
decomposition arithmetically satisfies the pattern.

**But the substantive question is: does the split make the composite
claim more informative than its parts?** Three isolated annotation-
empty observations, each supported by exactly one atlas note, do not
compose into a *pattern* claim in the sense E15' intends. E15' is
essentially the observation "three atlas records happen to share an
annotation state," which is a corpus-fact but not a mechanism-level
finding.

**C4: three-way split arithmetically available, but the split reveals
the claim is a conjunction of atomic observations rather than a
composite property. (PE — challenges without falsifying.)**

---

### C5 — Higher abstraction: does the claim merge?

**Test:** Does E15' merge with a broader annotation-pattern
observation?

**Candidate merges:**

- **With a corpus-hygiene observation:** "The frozen atlas's `breaks`
  field is annotated inconsistently — five records carry `[]`
  regardless of outcome_class." This is a genuinely broader claim;
  the merge would absorb E15' into a **corpus-hygiene finding**
  rather than an atlas invariant.

  Under AGENTS.md §Abstraction safety: this merge would erase
  distinctions between the three shared_by members and other atlas
  members with `[]` — it would say something *about the atlas's
  annotation practice*, not about es-01/es-03/es-10 specifically. That
  reframing is more accurate to what the evidence supports.

- **With N4 theme itself:** E15' is the sole occupant of Theme N4;
  there is no other candidate to merge with within N4.

**C5: productive merge available with a broader corpus-hygiene
observation.** This is a real merge signal — E15' is more honestly
framed as a corpus-annotation-practice observation than as an atlas
invariant. (SGO — challenges via merge availability.)

---

### C6 — Sampling and publication bias

**Test:** Is E15''s inductive base sufficient for the claim it makes?

**Support audit (SGO):**

Positive support:
- es-01, es-03, es-10 all carry `breaks: []` at HEAD — three records.

Anti-discrimination evidence (from C3):
- es-06 and es-11 also carry `breaks: []` with different outcome_class.
  The field state is non-discriminating.

Redundancy check:
- The three shared_by members are the *three* atlas notes whose
  annotations were used to derive the claim. The claim's support set
  IS its evidence base. This is the definition of a redundant inference:
  E15' asserts a pattern that its evidence exhaustively is.

**Weakening landings:**

1. **The claim's support set is exactly its evidence base.** E15'
   observes that three specific records have a specific annotation
   state. That is not a generalisable pattern; it is a listing of
   three records.
2. **The annotation field is non-discriminating across outcome_class.**
   Combined with C3's finding, E15' cannot function as a
   discriminator, only as an enumeration.
3. **The claim's admission would put a corpus-fact enumeration into
   the invariant table.** Whether that is a category error depends on
   the invariant table's admission discipline; it is at minimum a
   category tension.

**C6: real weakening landing (SGO on 1–2; PE on 3). Substantial
weakening.**

---

### C7 — Correlation vs causal obstruction

**Test:** Does E15' assert a *causal* claim (the mechanisms
structurally break nothing), a *correlational* claim (the annotations
happen to be empty), or a *definitional* claim (E15' is literally the
statement that these three annotations are empty)?

Under the revised scope in [`../E15-correction.md`](../E15-correction.md):
E15' is deliberately narrowed to the **annotation-level**
observation — it explicitly disclaims mechanism-level content. So the
claim is **descriptive-correlational-at-most** by construction.

This is not a defect of E15' as revised — it is exactly what the
correction record intended. But it *is* a challenge signal:

**A descriptive-correlational claim is not a candidate invariant in
the sense AGENTS.md §Candidate invariant discipline defines the
term.** Invariants are hypotheses about conserved failure structure;
E15' is a hypothesis about *what the annotation records say*, which is
a claim about the corpus's file contents, not about its mathematical
structure.

**C7: E15' as revised is descriptive-correlational, not causal. This
is a definitional feature of the revision, and it challenges the claim's
appropriateness as an invariant candidate. (PE — challenges the
category, not the truth.)**

---

## Challenge summary

| Probe | Verdict | Tier |
|---|---|---|
| C1 known counterexample | none in atlas (all three annotations verified) | SGO |
| C2 synthetic counterexample | scope-collapses onto C1 | IIP |
| C3 partial-success preservation | **es-06 and es-11 preserve the property across outcome_class boundary — weakens** | **SGO — weakens** |
| C4 lower abstraction | three-way arithmetic split available, but reveals the claim is a conjunction of atomic observations | PE — challenges |
| C5 higher abstraction | **productive merge into a corpus-hygiene observation available** | **SGO — challenges via merge** |
| C6 sampling bias | **weakening: support set equals evidence base; non-discriminating field; category tension** | **SGO — weakens** |
| C7 causal vs correlational | **descriptive-correlational by construction; challenges the invariant-candidate category** | PE — challenges |

Completed decisive negatives: **0** (C1 supports, none of the negatives
achieved by decisive attack).
Weakening landings: **2** (C3, C6).
Merge signals: **1** (C5).
Category challenges: **2** (C4, C7).
Falsifications: **0**.
Splits: **0** admissible (C4's arithmetic split reveals atomicity, not
productive splitting).

---

## Disposition under the predeclared decision rule

`surviving` requires ≥1 completed decisive negative. Not earned.
`falsified` requires a counterexample or completed synthetic
refutation. Not earned.
`split` requires ≥2 admissible child predicates with disjoint support.
C4 shows arithmetic decomposability but reveals the claim is an
enumeration, not a productive split.
`weakened` requires a code-verifiable weakening critique. **Two
independent weakening landings (C3, C6), plus a merge signal (C5) and
two category challenges (C4, C7).**

**Disposition: `weakened`.** E15' does not admit as a surviving
invariant.

Under AGENTS.md §Candidate invariant discipline, the strongest reading
of this challenge is that **E15' is category-appropriately a corpus-
hygiene observation, not an atlas invariant**. The revision correctly
removed E15's mechanism-level overreach; the challenge now reveals
that after that removal, the residual claim is not the kind of thing
the invariant machinery is for.

---

## Boundary-delta produced

The challenge extracts two related but distinct observations, of which
only one is category-appropriate for the invariant table:

```text
E15''  (proposed, corpus-hygiene observation — NOT invariant scope):
  The frozen atlas's `breaks` field is not consistently discriminating
  between outcome_class values. Five records (es-01, es-03, es-06,
  es-10, es-11) carry `breaks: []`; the field's discrimination between
  partial_success and partial_failure is weak. This is a corpus-
  annotation-practice observation.

E15''' (proposed, atomic listing — NOT productive invariant):
  At HEAD 859cbbc, the corpus/train/es-01, es-03, and es-10 files each
  carry `breaks: []` in their frozen mechanism records. This is a
  three-record enumeration.
```

Neither of these is admitted to the invariant table. The corpus-hygiene
observation may be worth acting on as an atlas-quality task (out of
scope for this record); the atomic listing is durable by virtue of
git history and requires no further pipeline action.

---

## Search-policy implications (proposed, not applied)

1. **Do not admit E15' as a surviving invariant.** Retain at
   `weakened`. Under `internal/pipeline/frontier.go`'s targetable-states
   rule, weakened invariants do not influence frontier generation.
2. **Do not promote E15 or E15' to a frontier target.** They are
   corpus-fact observations, not conserved-failure-structure
   hypotheses.
3. **Consider a separate atlas-hygiene task:** either mark the atlas's
   `breaks` fields as exhaustive/non-exhaustive per record, or
   normalise the field so its content correlates with outcome_class in
   a documented way. Out of scope for this record but recorded as a
   corpus-quality follow-on.
4. **N4 theme's admission gate closes as `weakened`.** N4 was E15's
   sole occupant; the theme itself does not admit a surviving
   candidate. This is not a failure of the pilot — it is the correct
   epistemic verdict on an annotation-scope observation.
5. **Redundancy warning for future candidate-invariant proposals:**
   any proposal whose evidence base is `field-value patterns in the
   frozen annotation records` (rather than mechanism-level structural
   observations) should be gated against E15's original defect-1 shape.

Applying any of these is a separate typed policy-mutation operation.

---

## Provenance

- Parent (revised claim) record:
  [`../E15-correction.md`](../E15-correction.md).
- Original ledger entry: `adjudication-ledger.json` E15 (lines
  546–583).
- Disposition instruction: `records/RESULT.md` §E15 (lines 212–245).
- Discovery population: Pilot-004 quorum result at commit `78e73d2`.
- Assessment population: current atlas at commit `859cbbc`
  (`corpus/train/es-*.md`).
- S1 (discovery vs assessment population fix): commit
  [`f0c0710`](https://github.com/instagrim-dev/gow/commit/f0c0710) —
  the discipline this challenge follows.
- Related epistemic discipline: `docs/invariant-challenge.md` §The
  lifecycle and §Challenge types; AGENTS.md §Candidate invariant
  discipline, §Abstraction safety.
- Executor: model-assisted, single-operator recorded. Line-anchored
  citations are CMA (independently reproducible with `rg` on the
  frozen atlas); category-tension observations (C4, C7) are PE.
  Same same-model-family caveat as parent records — not independent
  human verification.
