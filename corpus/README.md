# Erdős–Straus synthetic blinded benchmark corpus

This directory is a **synthetic blinded benchmark** for exercising the `newf`
invariant-guided search end to end. It is **not** a historical holdout, and it
is not a set of unit-test fixtures (the proof-of-shape fixtures live in
`../fixtures/`). It is project-authored mechanism material for the Erdős–Straus
conjecture, split into a **train** set and a **blinded target** set.

> **Why not a historical holdout (important):** an earlier version of this
> corpus labeled the affine-lattice / geometry-of-numbers target as
> *post-`2026-09-07`* and called the split a *historical holdout*. That claim was
> false and has been removed. The affine-lattice / linear-forms formulation of
> Erdős–Straus is **not** later than the pre-cutoff material: conditions linear
> in `n` describing the solution set as an affine class in `Z³`
> (`(4b−1)(4c−1)=4Pδ+1`) appear in the literature well before that date — the
> 2025 ED2 write-up (arXiv:2511.07465, §4.3 "Linear forms and lattices")
> explicitly cites *earlier* work for exactly this move. **Folder separation is
> not a historical chronology.** A genuine historical-holdout experiment (EPIC
> M7) requires an externally auditable cutoff and dated sources; this corpus
> does not provide that and must not be cited as historical evidence.

## What this benchmark IS and IS NOT

- **IS:** a blinded train/target split usable to check whether invariant-guided
  frontier generation, given only the `train/` families, recovers the
  *structural move* of a target family it was not shown. The blinding
  (target withheld from the working atlas) is real and enforced by the split.
- **IS NOT:** evidence that `newf` predicts a *historically later* advance. No
  chronological claim is made or implied. Any published-date ordering here is
  incidental and unaudited.

## Blinding discipline

- **Train set:** everything under `train/` is the working mechanism atlas the
  invariant miner runs against.
- **Blinded target:** `target/` stages one mechanism family that must be
  **withheld** from the train atlas during a run. The acceptance question is
  whether invariant-guided generation, given only `train/`, recovers the
  structural move in `target/`.
- Do **not** ingest `target/` into the same problem/workspace as `train/` when
  running the benchmark. Ingest it only to score recovery afterward.

## What is and is not claimed

These notes summarize real, published mechanism families and their known
structural boundaries (Mordell polynomial identities, the quadratic-residue
obstruction, Elsholtz–Tao Type I/II classification and averaging bounds,
factorization schemes, computational verification, covering-system attempts,
etc.). They are **project-authored summaries for normalization**, not original
mathematics and not verified proofs. Every normalized field records whether it
is `explicit` in the summarized source direction, `inferred`, or `unsupported`;
epistemic status never silently upgrades (see `../AGENTS.md`).

Each note's prose may name an approximate era for context, but that era is
**not audited** and carries no experimental weight. The only durable boundary is
the `train/` vs `target/` blinding split, not the prose and not any date.

## Layout

```text
corpus/
  README.md                 (this file)
  train/                    (working mechanism atlas; the miner sees these)
    es-01-mordell-polynomial-identities.md
    es-02-covering-system-attempt.md
    ...
  target/                   (blinded target family; withheld during a run)
    es-target-affine-lattice-linear-forms.md
```

## Running it

```bash
newf init "Erdős-Straus conjecture"                 # -> <prb>
newf ingest ./corpus/train --recursive --problem <prb>
newf normalize --problem <prb> --all
newf approach list --problem <prb>                  # the mechanism atlas
```

Scoring recovery (after invariant-guided generation) additionally ingests the
blinded target into a *separate* problem and compares mechanism signatures.
