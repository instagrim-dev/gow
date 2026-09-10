# Erdős–Straus historical-holdout corpus

This directory is the **research corpus** for the first `newf` historical-holdout
experiment (EPIC milestone M7). It is not a set of unit-test fixtures — the
proof-of-shape fixtures live in `../fixtures/`. This is the actual pre-cutoff
failure/partial-success material the invariant-guided search runs against.

Each `*.md` note describes one **mechanistically distinct** approach to the
Erdős–Straus conjecture (that for every integer `n ≥ 2` the equation
`4/n = 1/x + 1/y + 1/z` has a solution in positive integers), and carries an
embedded `newf-normalize` block (`normalize/v1`) so the whole tree flows
`ingest → normalize → typed mechanism atlas` deterministically via the offline
fixture provider.

## Holdout discipline

The experiment tests whether failure history compressed into invariants predicts
a *later* productive research direction better than undirected generation. To
keep that falsifiable:

- **Cutoff:** `2026-09-07`. Everything under `pre-cutoff/` is material whose
  mechanism was known and published on or before the cutoff.
- **Held-out target:** `holdout/` stages one later-productive mechanism family
  that must be **hidden** from the pre-cutoff corpus during the experiment. The
  acceptance question is whether invariant-guided frontier generation, given
  only `pre-cutoff/`, recovers the *structural move* in `holdout/`.
- Do **not** ingest `holdout/` into the same problem/workspace as `pre-cutoff/`
  when running the holdout. Ingest it only to score recovery afterward.

## What is and is not claimed

These notes summarize real, published mechanism families and their known
structural boundaries (Mordell polynomial identities, the quadratic-residue
obstruction, Elsholtz–Tao Type I/II classification and averaging bounds,
factorization schemes, computational verification, covering-system attempts,
etc.). They are **project-authored summaries for normalization**, not original
mathematics and not verified proofs. Every normalized field records whether it
is `explicit` in the summarized source direction, `inferred`, or `unsupported`;
epistemic status never silently upgrades (see `../AGENTS.md`).

Each note's prose names an approximate era so the cutoff is auditable, but the
durable holdout boundary is the `pre-cutoff/` vs `holdout/` split, not the prose.

## Layout

```text
corpus/
  README.md                 (this file)
  pre-cutoff/               (≤ 2026-09-07 mechanism families; the working atlas)
    es-01-mordell-polynomial-identities.md
    es-02-covering-system-attempt.md
    ...
  holdout/                  (later productive family; hidden during the experiment)
    es-holdout-affine-lattice-linear-forms.md
```

## Running it

```bash
newf init "Erdős-Straus conjecture"                 # -> <prb>
newf ingest ./corpus/pre-cutoff --recursive --problem <prb>
newf normalize --problem <prb> --all
newf approach list --problem <prb>                  # the mechanism atlas
```

Scoring recovery (after invariant-guided generation) additionally ingests the
holdout into a *separate* problem and compares mechanism signatures.
