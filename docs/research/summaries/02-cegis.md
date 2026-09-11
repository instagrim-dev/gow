# CEGIS — Counterexample-Guided Inductive Synthesis

| Field | Value |
|---|---|
| **Canonical citation** | Solar-Lezama, A. (2008). *Program Synthesis by Sketching*. PhD thesis, UC Berkeley. Earlier: Solar-Lezama, A., Tancau, L., Bodík, R., Seshia, S., Saraswat, V. (2006). "Combinatorial Sketching for Finite Programs." ASPLOS. |
| **Bib key** | `solar2008program` (present in `paper/references.bib`) |
| **Field** | Program synthesis / formal methods |
| **GoW role** | Strongest ancestor of the anti-vacuum / constraint-collapse formulation |
| **Epistemic status** | `GeneratedInterpretation` — written from model knowledge, **not** verified against the primary source. Verify before citation. |

## One-sentence core

Synthesize a program by alternating a *synthesizer* that proposes a candidate
consistent with all counterexamples seen so far and a *verifier* that either
certifies the candidate or returns a new counterexample.

## The mechanism (typed loop)

```text
candidate  c ← synthesize(consistent with E)      E = counterexample set
verdict    ← verify(c, spec)
if verdict = correct        → done
if verdict = counterexample e → E ← E ∪ {e}; loop
```

In sketching, the search space is a *partial program* (a sketch with holes);
the synthesizer solves for hole values, typically via SAT/SMT.

## Key load-bearing ideas

- **A counterexample kills a class, not a candidate.** One input `e` on which
  `c` misbehaves excludes *every* candidate that misbehaves on `e`. Each
  iteration removes a (possibly large) region of candidate space:
  `Ω_{t+1} = Ω_t \ F_t`.
- **Inductive generalization from finitely many examples** — the synthesizer
  only ever sees a finite `E`, yet must produce a candidate correct on all
  inputs; the verifier closes the gap.
- **Separation of proposer and checker** — the synthesizer may be heuristic;
  soundness lives entirely in the verifier.
- **The sketch bounds the space** — tractability comes from a human-supplied
  candidate representation with known coordinates (the holes).

## What it assumes is given in advance

- A **specification** (functional correctness criterion).
- A **complete, sound verifier**.
- A **candidate representation** (the sketch / DSL / grammar) with known,
  finite or effectively enumerable coordinates.

## What it produces

A concrete artifact certified correct against the specification, plus (as a
byproduct) the counterexample set that carved the space.

## Mapping to GoW vocabulary

| CEGIS | GoW / `newf` term |
|---|---|
| Candidate space Ω | `PossibilitySpace` (Ω) |
| Counterexample-excluded class F_t | `Constraint` (C_i) |
| Surviving candidate space | `ResidualRegion` (Ω_W) |
| Verifier verdict | `LandingPoint` (`Evaluation`) |
| Synthesize-from-E | `Projection` with constraints |

## What GoW borrows

The exact form of the anti-vacuum (`05-complement-geometry.md`): failure is
information about the *remaining* space, and the informative unit is the
excluded class, not the rejected instance.

## Where GoW departs

CEGIS knows its specification, verifier, and candidate coordinates **in
advance**. GoW's harder problem is that the relevant *dimensions of the
space* must be inferred from historical Work; its constraints are semantic
("must break class locality"), not native predicates over a fixed encoding;
and its "verifier" is a tiered strength hierarchy, not a sound oracle.

## Reduction test (how this tradition attacks GoW)

> If GoW's shape vocabulary can be fixed up front as a finite feature grammar,
> then GoW is CEGIS over that grammar with a weak verifier — a known method
> with an unsound checker, which is strictly worse than CEGIS, not novel.

Defense requires showing the representation genuinely moves (`M(W)` model
selection) in response to Work, and that this movement earns predictive power.
