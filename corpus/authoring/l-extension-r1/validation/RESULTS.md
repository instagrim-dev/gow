# Validation record — (l-extension-r1)

Executed 2026-09-12 at repository head `cbc3685a0a2bb668f59fc7985680f4ccdb665957`.
All commands ran against **disposable databases** under `/tmp/lext-validation/`
with explicit `--db` paths. No live `.newf/` store was read or written.

## 1. What was checked, and what each check does not establish

| Check | Result | Establishes | Does NOT establish |
|---|---|---|---|
| Structural validation | PASS | Package is internally consistent and machine-readable | That any cited passage supports its claim |
| Source-byte lineage vs ES | `disjoint` | Non-overlapping source-byte sets in the compared inventories | Research-methodological independence |
| Source-byte lineage vs pvnp | `disjoint` | Same | Same |
| Fixture normalization | 12/12 | Technical schema compatibility | Any scientific or research validity |
| Empty-population control | `empty` | The diagnostic reports inconclusive, not approval | — |
| Validator regressions | 26 pass | The checker rejects the mandated defect classes | — |

## 2. Commands and results

### Structural validation

```bash
python3 corpus/authoring/l-extension-r1/validate.py
```

```text
work records:   12
logical ids:    12 unique
cited sources:  15
manifest sha256: 6b097f85b8d511b1ea516589967f9396bafa9a7aafe6e1eb1a4f0681f037bfa1
PASS — structural checks only.
```

Exit 0. The manifest digest is reported separately because a manifest cannot
contain the hash of its own final bytes.

### Disposable store construction

```bash
newf --json --db /tmp/lext-validation/new.db  init "sha1-collision-l-extension-r1"
newf --json --db /tmp/lext-validation/es.db   init "es-source-only-comparison"
newf --json --db /tmp/lext-validation/pvnp.db init "pvnp-source-only-comparison"
```

Every `init`-returned problem ID was captured and reused in subsequent commands;
no persisted ID from an existing database was assumed to exist here.

| Store | Problem ID | Ingested | Population |
|---|---|---|---|
| new | `prb_01M2BZDZS0X3QZFKJYAKP72JT0` | 12 | `corpus/authoring/l-extension-r1/notes` |
| ES | `prb_01M2BZE0073P99TTPKEYNGXT00` | 12 | `corpus/train` (**train only**) |
| pvnp | `prb_01M2BZE0C80442HQPCSV6DD00X` | 12 | `corpus/experiments/pvnp-holdout/train` (**train only**) |

Populations are **train inventories only**. Targets are excluded from all three
sides so the comparison covers like with like.

### Source-only lineage comparisons

```bash
newf --json --db "$NEW_DB" source lineage-diff \
  --problem "$NEW_P" --against-db "$ES_DB" --against-problem "$ES_P"

newf --json --db "$NEW_DB" source lineage-diff \
  --problem "$NEW_P" --against-db "$PVNP_DB" --against-problem "$PVNP_P"
```

| Comparison | Verdict | shared | left-only | right-only |
|---|---|---|---|---|
| new vs ES train | `disjoint` | 0 | 12 | 12 |
| new vs pvnp train | `disjoint` | 0 | 12 | 12 |
| ES vs pvnp train (control) | `disjoint` | 0 | — | — |

Retained: `lineage-new-vs-es.json`, `lineage-new-vs-pvnp.json`,
`lineage-es-vs-pvnp.json`.

The third row is a **control**: it reproduces the already-recorded ES/pvnp
verdict from `source-lineage-admission-gate-result.md` in a freshly built store,
confirming the harness behaves as previously reported rather than producing
`disjoint` indiscriminately.

**No overlap required explanation.** No content was reformatted, renamed, or
paraphrased to influence any of these verdicts.

### Empty-population negative control

```bash
newf --json --db "$NEW_DB" source lineage-diff \
  --problem "$NEW_P" --against-db "$EMPTY_DB" --against-problem "$EMPTY_P"
```

Verdict: **`empty`**. Retained: `lineage-new-vs-empty.json`. Confirms that a
missing comparison population yields an inconclusive verdict, never an
independence approval. `empty` is not a pass.

### Fixture compatibility smoke test

```bash
newf --json --db "$NEW_DB" normalize --problem "$NEW_P" --all --provider fixture
```

`ok: true`; 12 of 12 records normalized, one approach each, all `created`.
Summary retained in `normalize-summary.json` with scratch-store IDs omitted as
disposable-database noise.

**Stopped here deliberately.** No mining, clustering, failure-space build,
challenge, or frontier generation was run. Doing so would convert a compatibility
check into an unauthorized research run, and would also invite amending source
selection in response to lifecycle labels — which the selection rule forbids.

### Validator regressions

```bash
python3 -m unittest discover -s corpus/authoring/l-extension-r1 -p 'test_*.py'
```

`Ran 26 tests ... OK`. Coverage includes: stale manifest under changed bytes,
truncated display-prefix hashes, missing and escaping/absolute paths, duplicate
logical IDs, missing evidence locators, unsupported `outcome.class`, locators
pointing past the note's paragraphs, invalid fixture JSON, unclosed payload
blocks, invalid support kinds, invalid posture enums, unresolved vocabulary
(reported, never aliased), wrong schema version, cited source IDs absent from the
register, sources missing verification status or byte-retention statement,
record-count mismatch, self-granted `execution_authorization`, and `independent:
true`.

## 3. What was not verified by machine

**The semantic support review is outstanding.** Content hashes and schema checks
cannot establish that a cited passage supports the claim attributed to it. See
[`REVIEW.md`](REVIEW.md) for the specific per-record obligations, including the
seven source-level limitations carried in `manifest.json` `unresolved_items`.

## 4. Reproduction

The scratch databases were disposable and are not retained; `.newf/` is
gitignored and was untouched. To reproduce, rerun the commands above against
fresh `--db` paths. The corpus files and the ES/pvnp train inventories are
tracked, so the inputs are reconstructable from a clone even though the stores
are not.
