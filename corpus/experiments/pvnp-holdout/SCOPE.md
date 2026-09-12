# P vs NP historical-holdout protocol — SCOPE (pre-capture draft)

**Status: SCOPING DRAFT — not frozen, no corpus captured, no run defined.**
**Epistemic status of this document:** model-authored design artifact. Every
date and locator below is a *claim requiring operator audit* before freeze —
that audit is exactly what `experiment date-source --evidence` records. Nothing
here is admissible as dating evidence itself.

Issue: [#13](https://github.com/instagrim-dev/gow/issues/13). D7 disposition
(2026-09-12): retained, blocked-on-M7. The M7 code gap closed the same day
(`be09798`: dating write path + historical conclusion vocabulary), so the
blocker is now **this corpus** — dated, auditable holdout data.

## The experiment in one sentence

Given only mechanism-normalized P vs NP proof-strategy failures dated **before
2005-01-01**, does invariant-guided frontier generation recover the structural
move of the **algorithmic method** (Williams 2010–2011) better than undirected
baselines, under equal budgets, blinding enforced, and the conclusion earned
through the `historical` vocabulary gate?

## Why this cutoff and this target

**Cutoff: `2005-01-01T00:00:00Z`.** It cleanly separates the three classical
barriers and their era's failed programs (all pre-2005) from the two major
post-cutoff structural events (algebrization 2008, the algorithmic method
2010). It deliberately places Kabanets–Impagliazzo (STOC 2003) **inside**
train — see the honesty section below.

**Target: the algorithmic method** — Williams, *Improving exhaustive search
implies superpolynomial lower bounds* (STOC 2010) + *Non-uniform ACC circuit
lower bounds* (CCC 2011; JACM 2014). Chosen over algebrization because:

1. It is an **advance**, not another barrier: an unconditional lower bound
   (NEXP ⊄ ACC⁰) produced by a genuinely new mechanism.
2. Its structural move is crisply statable in mechanism vocabulary:
   *invert the algorithms→lower-bounds direction* — a nontrivial (better than
   brute-force) satisfiability algorithm for a circuit class, coupled to the
   nondeterministic time hierarchy, yields lower bounds against that class.
3. It **provably evades all three trained barriers** (non-relativizing,
   non-natural, non-algebrizing) — so recovering it requires violating the
   surviving failure structure, which is precisely the frontier-generation
   discipline the system claims to embody.

## Train inventory (12 sources, all pre-cutoff)

Dates below are publication-venue claims to be operator-audited; per-source
locators go in each file's provenance and, for the *target*, into
`holdout_source_dating` via `experiment date-source`.

| # | Source (mechanism family) | Claimed date | Anchor |
|---|---|---|---|
| 1 | Diagonalization / simulation program (pre-1975 posture) | 1960s–70s | Hartmanis–Stearns 1965; Cook 1971; Karp 1972 |
| 2 | **Relativization barrier** — contrary oracles | 1975 | Baker–Gill–Solovay, SIAM J. Comput. 4(4) |
| 3 | AC⁰ lower bounds — parity via random restrictions | 1981–1986 | Ajtai 1983; Furst–Saxe–Sipser 1984; Håstad 1986 |
| 4 | Monotone circuit lower bounds — approximation method | 1985 | Razborov 1985 (clique) |
| 5 | The monotone→general hope and its failure | 1985–1988 | Razborov 1985 (perfect matching, superpolynomial monotone); Tardos 1988 (exponential gap, Combinatorica) |
| 6 | AC⁰[p] — polynomial approximation over F_p | 1987 | Razborov 1987; Smolensky 1987 |
| 7 | Proof-complexity program — resolution lower bounds | 1985– | Haken 1985 (PHP); Beame–Pitassi survey 1998 |
| 8 | **Natural-proofs barrier** — largeness+constructivity vs PRFs | 1994 | Razborov–Rudich, STOC 1994 / JCSS 1997 |
| 9 | Circuit-lower-bound program stall post-natural-proofs | 1994–2004 | e.g. Razborov 1995 (independence); surveys |
| 10 | Geometric complexity theory — orbit closures, representation theory | 2001 | Mulmuley–Sohoni, SIAM J. Comput. 31(2) |
| 11 | Derandomization↔lower-bounds bridge | 2003 | **Kabanets–Impagliazzo, STOC 2003** (identity testing ⇒ arithmetic/Boolean lower bounds) |
| 12 | Time-space tradeoffs for SAT — indirect diagonalization | 1997–2004 | Fortnow 1997; Fortnow–van Melkebeek 2000 |

Each file follows the `corpus/train/` shape: prose + `newf-normalize` payload
with mechanism axes (representations, assumptions, operators, preserves,
breaks, locality, construction/uncertainty mode) and a failure/partial outcome
with boundary statements. The two barrier rows (2, 8) are the load-bearing
failure invariants; rows 11–12 are the partial-success structure nearest the
target.

## Target problem (quarantined; every source withheld + dated)

| Source | dated_at (claimed) | Evidence locator (to audit) |
|---|---|---|
| Williams — improving exhaustive search ⇒ lower bounds | 2010-06-05 (STOC 2010) | DOI 10.1145/1806689.1806723 |
| Williams — non-uniform ACC lower bounds | 2011-06-08 (CCC 2011) | DOI 10.1109/CCC.2011.36 (JACM 2014: 10.1145/2559903) |

Both dates are strictly after the cutoff, as `experiment date-source` enforces.
The target file's mechanism payload must state the structural move in the same
vocabulary the train files use, WITHOUT importing post-hoc labels the train
era lacked ("the algorithmic method" as a name postdates the move).

## Honesty constraints (freeze-blocking if violated)

1. **The KI03 precursor stays in train.** Excluding it would curate away the
   partial-success structure that makes the prediction question meaningful.
   Consequence, stated up front: recovery credit is graded — a proposal that
   merely re-lands KI03's identity-testing bridge is `mechanism-near` to
   *train*, not to the target; the target's distinct content is the coupling
   (circuit-analysis algorithm → nondeterministic time hierarchy → class
   lower bound). If `recovery-rule/v1` on `classify/v3` cannot separate
   those two, the experiment is **inconclusive by calibration**, and that
   result must be reported, not tuned away.
2. **Algebrization (2008) appears NOWHERE.** It is post-cutoff but is not the
   target; leaking it into train is a chronology violation, and adding it to
   the target set would dilute the recovery question. It is simply absent.
3. **No survey contamination.** Train prose must not cite or paraphrase
   post-2005 retrospectives (Fortnow 2009, Aaronson 2016). Each train file's
   provenance names only pre-cutoff primary anchors. This is the main
   authoring hazard since any modern author knows the ending; mitigation:
   every mechanism claim in a train file must be traceable to its pre-cutoff
   anchor, and the operator audit checks provenance before freeze.
4. **Blinding is code-audited anyway** (sha256 leakage check), but content
   identity cannot catch *paraphrase* leakage — barrier 3 is a human gate,
   recorded as an operator attestation in the freeze record.
5. **Model-judgment ceiling.** Unlike ES, no deterministic checker can score
   a proposal's *mathematical* merit here. What IS code-owned: the recovery
   classification (mechanism comparison under the pinned profile), budgets,
   blinding, and the conclusion vocabulary. The result claims recovery of a
   structural move, never correctness of a proof strategy.

## Execution shape (after freeze; per M7 + D6 discipline)

```text
init train + quarantined target (separate problems)
 -> ingest train (12 sources) / ingest target (withheld, 2 sources)
 -> experiment define --mode historical --cutoff 2005-01-01T00:00:00Z
 -> experiment date-source (x2, operator-audited locators)   # gate lift
 -> normalize -> signatures -> cluster -> failure-space -> mine -> challenge
 -> experiment run (b0_undirected + b3_invariant_guided, equal budgets)
 -> conclusion in {predicts_later_advance, fails_to_predict, inconclusive}
```

Preregistration (freeze) must fix, before any arm executes: budgets, arms,
comparison profile, recovery rule version, and the KI03-vs-target grading
note above. Freeze is a commit pinning this file plus the captured corpus,
mirroring `corpus/experiments/two-arm-comparison/` (protocol frozen at
`d75b4e0` before execution).

## What remains before freeze (operator work)

- [ ] Audit every claimed date/DOI above against primary records
- [x] Author the 12 train files + 2 target files (mechanism payloads) —
      drafted 2026-09-12 (`train/pnp-01..12`, `target/pnp-target-01..02`);
      smoke-validated: all 14 ingest and normalize cleanly, historical
      define + dating gate lifecycle verified on this corpus with
      PLACEHOLDER evidence locators (smoke DB destroyed; no experiment
      was run — execution stays behind freeze + audit)
- [x] Adversarial pass: search train prose for post-cutoff contamination —
      **done 2026-09-12, two-round blind review.** Round 1 (reviewer never
      shown target or SCOPE): verdict CONTAMINATED — apparatus vocabulary
      in-band ("cutoff"/"train atlas"/"Pre-cutoff" tags in train prose, one
      "as of 2005" slip), and pointer sentences in pnp-09/10/11/12 that
      editorialized toward the target (worst: a pnp-11 payload note that
      referenced "the withheld advance", and a recipe sentence naming
      "exhibiting an actual, unconditional algorithm and running such a
      bridge"). All remediated: apparatus vocabulary replaced with
      2004-honest dating, pointer/recipe sentences removed or reduced to
      period-recorded boundaries. Round 2 (fresh blind reviewer): verdict
      MINOR ISSUES — three residual pnp-12 items ("decades of refinement"
      anachronism, "hybrid ancestor" teleology, descendant-aware payload
      note), all fixed; the remaining top telegraph judged "fair in
      substance — inferable from honestly-recorded failure structure,
      which is the correct state for a valid holdout."
      **Accepted residual, with rationale:** pnp-11 is the only train
      family whose payload breaks `model-analysis-first direction`. This
      is machine-recoverable, and it stays: it is a true 2003 mechanism
      fact (the KI03 title itself states the direction), the mechanism
      axes are the experiment's intended discriminative substrate (as
      residue-locality is in the ES corpus), and SCOPE's honesty
      constraint #1 already prices it — a proposal that merely re-lands
      the KI03 bridge is mechanism-near to TRAIN, not the target, so the
      unique profile cannot be cashed in for recovery credit by itself.
- [ ] Fix budgets and arms; record the freeze commit
- [ ] Only then: `date-source` with audited locators, and run
