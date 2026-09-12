# (l-extension-r1) — SHA-1 collision cryptanalysis candidate corpus

**Artifact status:** candidate authoring package, structurally validated.
**Not** a frozen experiment, **not** an executed study, **not** an independent
replication.

Twelve Work records covering attempts to collide the SHA-1 hash function,
1998–2020, authored from verified primary locators.

## What this is

| | |
|---|---|
| Domain | SHA-1 collision cryptanalysis |
| Objective | Produce a collision for the full, unmodified SHA-1 |
| Work records | 12 |
| Source register | 16 entries (14 usable, 1 excluded, 1 register-context only) |
| Distinct external work origins | 13 publications plus 1 artifact pair |
| Outcome spread | 1 success, 6 partial success, 2 partial failure, 3 failure |
| Schema | `normalize/v1`, embedded `newf-normalize` payloads |
| `execution_authorization` | `null` |

## Why this domain

Selected on evidence availability under the seven criteria fixed in
[`SCOPE.md`](SCOPE.md) **before** sources were chosen. Two properties made it
win over the other shortlisted candidates:

- **Checkable outcomes.** A collision is a concrete object. The 2017 success is
  verifiable by anyone who computes a digest — the strongest outcome evidence
  available in any GoW corpus to date.
- **A real retraction.** `sha1-07` is a withdrawn paper whose ePrint page serves
  its own retraction note. That is a failure record produced by the field's
  error-correction, not by a curator's judgment.

## Reading order

The records form a mechanism arc; reading in order shows what each attempt
preserved and broke.

| Record | Attempt | Outcome |
|---|---|---|
| `sha1-01` | Linearized local-collision codewords | failure |
| `sha1-02` | Neutral-bit amplification | partial failure |
| `sha1-03` | Multi-block message modification | partial success |
| `sha1-04` | Automated characteristic search | partial success |
| `sha1-05` | Boomerang auxiliary differentials | partial success |
| `sha1-06` | Disturbance vector classification | partial failure |
| `sha1-07` | Composed 2^52 path — **withdrawn** | failure |
| `sha1-08` | SAT-encoded collision search | failure |
| `sha1-09` | Joint local-collision analysis | partial success |
| `sha1-10` | Freestart collision, full 80 steps | partial success |
| `sha1-11` | **First full SHA-1 collision** | success |
| `sha1-12` | Chosen-prefix via target clustering | partial success |

## Read this before using the records

Three limitations, stated up front rather than buried:

1. **Not outcome-blind.** The authoring agent knew the ES/pvnp mining outcomes
   before writing these records. See [`SCOPE.md`](SCOPE.md) §6.
2. **Not independent of the existing corpora.** Five of seven dependency
   dimensions are shared with ES and pvnp by construction. See
   [`DEPENDENCIES.md`](DEPENDENCIES.md).
3. **Cannot establish ES/pvnp independence.** A third corpus is irrelevant to
   whether the existing two were independently constructed.

## Files

| Path | Contents |
|---|---|
| `SCOPE.md` | Objective, selection rule, shortlist and rejections, exposure record |
| `sources.json` | 16-entry source register with verification status per entry |
| `notes/*.md` | 12 Work records with embedded `newf-normalize` payloads |
| `DEPENDENCIES.md` | Seven-dimension dependency account vs ES and pvnp |
| `manifest.json` | Tracked contents, full SHA-256 per file, authoring provenance |
| `validate.py` | Structural validator (stdlib only, no authority flags) |
| `test_validate.py` | Regressions for the validator's boundary cases |
| `validation/` | Retained commands, outputs, and the review obligation |

## Verification

```bash
python3 corpus/authoring/l-extension-r1/validate.py
python3 -m unittest discover -s corpus/authoring/l-extension-r1 -p 'test_*.py'
```

A validator pass means the package is internally consistent and
machine-readable. It does **not** establish that any cited passage supports the
claim attributed to it; that is a human review obligation recorded in
`validation/REVIEW.md`.

## Permitted next action

Structural inspection and review. **Not authorized:** mining, challenge
campaigns, frontier generation, evaluation, freeze, or lifecycle promotion. Only
the operator can change `execution_authorization`.
