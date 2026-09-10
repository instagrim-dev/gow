# Erdős–Straus research corpus — 2026-09-10

Thirteen source-linked, project-authored research cards for the equation
`4/n = 1/x + 1/y + 1/z`, with positive integer denominators and `n >= 2`.
This is a literature expansion, **not an expansion of the existing blinded
benchmark** and not evidence that the full conjecture has been resolved.

## Admission boundary

Use a **new research problem and database**. Do not merge this collection into
`corpus/train/`, ingest it with `corpus/target/`, or reuse an old experiment's
blinding claim after exposing its participants to this literature. This batch
may reveal mechanisms related to the existing target. A new experiment needs
an explicitly reviewed split and exposure/provenance record; different file
hashes alone cannot rule out semantic leakage.

The collection contains **9 research notes**, **3 optional exploratory notes**,
and **1 quarantined solution claim**. These are curation buckets, not a quality
ranking or a claim that the first bucket is independently verified. Some papers
refine mechanisms already represented in the synthetic corpus. Thirteen papers
are not thirteen independent failure families.

## Source map

| ID | Source / selected date | Contribution to inspect | Important limit |
|---|---|---|---|
| [01](notes/esr-01-bright-loughran-brauer-manin.md) | Bright–Loughran, 2020 | Arithmetic geometry; existence versus strong approximation | A local compatibility set is not a positive integral witness |
| [02](notes/esr-02-elsholtz-tao-counting.md) | Elsholtz–Tao, v6, 2015 | Type I/II counts and polynomial congruence classification | An average is not every prime; pins a corrected version |
| [03](notes/esr-03-elsholtz-planitzer-enumeration.md) | Elsholtz–Planitzer, 2018 | Relative-GCD patterns and enumeration complexity | Upper bounds do not establish nonemptiness |
| [04](notes/esr-04-salez-modular-sieve.md) | Salez, 2014 | Modular equations, sieve, finite verification | Reported bound is finite; code not rerun |
| [05](notes/esr-05-mihnea-bogdan-verification.md) | Mihnea–Bogdan, 2025 | Parallel verification extending the Salez program | Same program lineage; not independent methodological support |
| [06](notes/esr-06-bradford-ionascu-euclidean-rings.md) | Bradford–Ionascu, v2, 2014 | Transfer to norm-Euclidean rings | Enlarging the ring changes the original problem |
| [07](notes/esr-07-bradford-pythagorean-bezout.md) | Bradford, 2021 | Pythagorean/Bézout auxiliary-object transformation | Necessary relations do not guarantee reconstruction |
| [08](notes/esr-08-bradford-divisor-congruences.md) | Bradford, 2024 | Smallest-denominator/divisor search | Correspondence does not prove universal pair existence |
| [09](exploratory/esr-09-bello-benito-fernandez-divisor.md) | Bello-Hernández–Benito–Fernández, 2026 | Divisor parametrization and bounded search | Prime/composite and bounded/unbounded domains differ |
| [10](exploratory/esr-10-mballa-perfect-square-reduction.md) | Mballa, v2, 2026 | Perfect-square/discriminant reformulation | Required square existence remains conjectural |
| [11](exploratory/esr-11-ghermoul-polynomial-families.md) | Ghermoul, 2025 | Polynomial families with conditional coverage | Family generation is not surjectivity |
| [12](quarantine/esr-12-bradford-solution-claim.md) | Bradford, 2026 | Full-solution claim retained for verification triage | No correctness verdict; excluded from normalization |
| [13](notes/esr-13-chamberland-prime-representation.md) | Chamberland, 2026 | Published prime Type II representation criterion | Theorem 1 and coverage Conjecture 2 are different claims |

Each linked card contains its primary-source URLs, selected version, inspected
sections, source-attributed account, curator interpretation, overlap notes,
and a proposed grounding task. The tasks are **not reported as executed**.
The source date is the selected arXiv version date or the stated publication
date, not a verified discovery date. Dates do not make this a historical holdout.

## Normalization and evidence strength

The 12 non-quarantined cards include `newf-normalize` contract blocks compatible
in shape with `docs/normalization.md`. Every populated annotated mechanism or
outcome field is tagged **inferred**. These blocks are curator interpretations
of the linked text, not verbatim source claims, author-approved metadata,
independent evidence, or proof certificates. The fixture provider imports the
blocks; it does not read the linked papers or discover their mechanisms.

`partial_success` in these blocks means a **source-reported, scoped result**
such as a finite verification, bound, or correspondence. It does not assert a
proof of Erdős–Straus. `unknown` must not become a failed approach, and a paper
that does not settle the conjecture must not be labeled a failed proof.
For outcome-sensitive experiments, independently review the claim scope and
verification strength before admitting these annotations as support/contrast.

Empty lists mean unrecorded structure, not verified absence. No field is marked
exhaustively complete. New vocabulary may remain unresolved under the current
ontology; do not force aliases just to improve clustering coverage. Shared
papers, authors, or method lineages must be retained when counting support.

## Use in a separate workspace

From the repository root, with a built `newf` binary:

```bash
./newf --db .newf/es-literature.db init "Erdos-Straus literature research"
# Substitute the returned problem ID below.
./newf --db .newf/es-literature.db ingest \
  ./corpus/research/erdos-straus-2026-09-10/notes --recursive --problem <problem-id>
./newf --db .newf/es-literature.db normalize --problem <problem-id> --all --provider fixture
```

Optional exploratory admission is a separate decision: ingest `exploratory/`
only after reviewing its three cards. Do not recursively ingest the collection
root: its README, registry and quarantine are not mechanism-training material.
No benchmark configuration, existing train/target file, or vocabulary is changed
by this batch.

## Integrity, licensing, and limits

`sources.json` is a bibliographic/curation register, not the application's
`SourceSnapshot` table. `note_sha256` hashes **our note bytes**, not the original
paper. Primary PDFs/HTML/source archives are not vendored; no original-source
content hash or download provenance is invented. Follow the cited versioned
links and archive source bytes through a separate authorized ingestion step.
Published papers retain their own rights and licenses; these cards paraphrase
short, selected findings rather than reproduce the papers.

```bash
python3 corpus/research/erdos-straus-2026-09-10/validate.py
```

The validator checks registry paths, note hashes, admission buckets, embedded
JSON and conservative annotation tags. It does not verify mathematics, fetch
sources, run the original computations, or substitute for the Go schema/tests.
This source-only contribution was locally linted; the repository Go build/test
suite and CLI ingestion were not executed in the curation environment.
