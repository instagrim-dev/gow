# (source-lineage-check) Result — four ES databases share 12/13 source snapshots byte-for-byte; ES lineage is 1 corpus at 4 HEADs, not 4 independent replications

**Status**: executed 2026-09-12. Discharges the residual half of
`(l-coh)` correction 5 for the ES side; leaves the ES↔pvnp
independence question for a separate loop.

## Bottom line

**The four ES databases are demonstrably not four independent
research replications.** They share **12 of 13** `source_snapshots.sha256`
values byte-for-byte, with identical `sources.logical_name` values on
each of those 12 sources. The single non-shared row is M7's specific
target problem file (`es-target-affine-lattice-linear-forms.attested.md`).

The "identical miner-eligible marginals across 4 ES corpora"
observation from `(l-population-validation)` is now **explained at
CMA-tier**: the four DBs ingested the same physical training corpus.
Any downstream fingerprint differences (signatures, clusters,
input-set hashes) reflect pipeline-internal versioning across
different HEADs, not source-domain differences.

## Method

SGO. For each of `.newf/{m7, pilot-001, pilot-002, pilot-003}/newf.db`,
gather multiset values at five fingerprint layers:

1. `source_snapshots.sha256` — literal SHA-256 of ingested source content.
2. `sources.logical_name` — human-provided source names.
3. `mechanism_signatures.fingerprint` — canonical signature identity.
4. `mechanism_clusters.cluster_fingerprint` — cluster-level identity.
5. `cluster_runs.input_set_hash` — clustering input identity.

Also `problems.slug` for organizational-layer context.

Compare pairwise as set intersections. All values printed to the
session's terminal history for auditability.

## Results

### Layer 1 — source_snapshots.sha256 (13 rows each)

Pairwise intersection sizes:

|            | M7 | p-001 | p-002 | p-003 |
|---|---:|---:|---:|---:|
| M7         | 13 | 12    | 12    | 12    |
| pilot-001  | 12 | 13    | 13    | 13    |
| pilot-002  | 12 | 13    | 13    | 13    |
| pilot-003  | 12 | 13    | 13    | 13    |

- 12 SHA-256 values shared by all four DBs.
- 1 M7-unique SHA-256:
  `fd0dd43d7f5210041776cc2dd07ade16ea18e8e1f77b48148de88611b02619c4`
  corresponding to `es-target-affine-lattice-linear-forms.attested.md`
  (M7's specific target file).
- 0 uniques for pilot-001, pilot-002, pilot-003 (all their 13 sources
  are within the M7 sha256 union).

**Sample sha256 values match byte-for-byte across all four DBs**, e.g.
`09be207cf0eb3221c95b9e3bc3dd31204d3f5d2106ee583131d9d113c41dc46e`
is present in M7, pilot-001, pilot-002, pilot-003 alike.

### Layer 2 — sources.logical_name (13 rows each)

Same 12-shared + 1-M7-unique pattern. The 12 shared logical names are
present in all four DBs; M7 additionally carries `es-target-affine-lattice-linear-forms.attested.md`.

### Layer 3 — mechanism_signatures.fingerprint

Pairwise intersection sizes:

|            | M7 | p-001 | p-002 | p-003 |
|---|---:|---:|---:|---:|
| M7         | 11 | 0     | 0     | 11    |
| pilot-001  | 0  | 10    | 0     | 10    |
| pilot-002  | 0  | 0     | 11    | 0     |
| pilot-003  | 11 | 10    | 0     | 33    |

Interpretation: despite identical source bytes, signature fingerprints
differ across DBs. This is expected — the fingerprint hashes canonical
signature content plus schema/vocabulary versions and/or provider
invocation identity. pilot-003 was ingested at (or migrated through) a
HEAD compatible with M7's and pilot-001's fingerprints; pilot-002
sits at a HEAD where both differ.

Nothing here contradicts the source-layer finding; it constrains what
"identical" means at the persistence layer. The relevant identity for
the (l-coh) claim was posture-axis label values, not fingerprint bytes.

### Layer 4 — mechanism_clusters.cluster_fingerprint

Zero pairwise overlap across all pairs. Each DB has completely unique
cluster fingerprints. Expected: cluster fingerprints incorporate
run-scoped context.

### Layer 5 — cluster_runs.input_set_hash

Zero pairwise overlap. Each `cluster_run` has a distinct hash.

### Problem organization

- M7: `erdős-straus-{affine-lattice-target-m7-blinded, conjecture-m7-train}`
- pilot-001: `pilot-001-es-{target, train}`
- pilot-002: `pilot-002-es-{target, train}`
- pilot-003: `pilot-003-es-{target, train}`

Different naming conventions across pilots, but consistent
target/train sourcing.

## Why identical source bytes produce identical marginals but different fingerprints

- **Marginals** (`fPrev`, `sPrev` per posture axis-value): computed
  over aggregate label counts of the miner-eligible family
  population. If the normalizer's classification logic (locality/
  construction/uncertainty) is stable across HEADs on identical
  source content, the marginals are identical. Consistent with
  `(l-population-validation)`'s recomputed tables.
- **Signature fingerprints**: hash canonical signature content
  including schema/vocabulary version metadata, which changes
  across HEADs. So identical source content at different HEADs
  produces different fingerprints — even though the classification
  labels used for marginals are stable.
- **Cluster/input_set fingerprints**: incorporate run-scoped
  identifiers, so they never overlap across independent
  ingestion runs.

## Consequences for (l-coh) correction 5

The reviewer's correction 5 said:

> "n=2 truly independent" is too strong. Identical aggregate
> distributions do not themselves prove identical underlying
> observations; different datasets can have the same histogram.
> Independence between the two corpus lineages (ES-shape and
> P-vs-NP-shape) has not been established either.
>
> Defensible replacement:
>
> Five database instances represent two substantive corpus
> lineages for this comparison. The four ES instances do not
> provide four independent research replications. Independence
> between the two corpus constructions has not been established.

**Half of this residual is now definitively closed:**

- On the ES side, **positive evidence** now exists that the four ES
  instances are not four independent replications. Their source
  content is 12/13 SHA-256-identical.
- On the ES-vs-pvnp side, the independence question remains open:
  they clearly use different source corpora (M7's affine-lattice /
  divisor-sum ES material vs pvnp's barrier-program corpus), but
  "independence" in a stronger sense (mutually informative,
  non-overlapping mechanism spaces, etc.) requires more work than
  a sha256 diff.

The reviewer's defensible-replacement wording remains **exactly the
right claim to keep**. This validation strengthens the "four ES
instances" clause with source-byte evidence and leaves the "two
corpus constructions" clause open as before.

## What this does NOT establish

- Does NOT establish that ES and pvnp are statistically independent
  research replications. Different source domains do not
  automatically imply independence in the research sense.
- Does NOT change the (l-coh) numerical tables (unaltered).
- Does NOT change (l-coh) correction 7's CMA-tier upgrade
  (unaltered — that's a code-mechanism claim, unrelated to source
  lineage).
- Does NOT change H3 admission rules, verifier tiers, policy
  directives, or wire schema fields.
- Does NOT establish whether the 4-way source sharing was
  intentional (deliberate corpus re-use across pilots) or
  incidental (each pilot happened to seed from the same
  attested-file directory). The record captures the SGO fact;
  intent is separate.

## Consequences for the scorecard

- (l-coh) correction 5 ES-side residual: **discharged** with
  source-byte evidence.
- (l-coh) correction 5 ES-vs-pvnp-independence residual:
  **unchanged** — remains open.
- All other (l-coh) corrections unchanged.
- (l-coh) numerical tables unchanged.
- (l-coh-mechanism-test) CMA upgrade unchanged.
- (l-population-validation) arithmetic-verification unchanged.
- **H3 gate: preserved.**
- Pure SGO validation via SQL over existing persisted schema.

## Practical downstream implication

For any future (l-extension) work — authoring a third truly
independent research corpus — the source-lineage layer now has a
concrete definition to satisfy: the new corpus's sha256 set must
be disjoint from both the ES sha256 set (which the four current
ES DBs share) and the pvnp sha256 set. This is a cheap admission
check and worth automating in the pipeline as a corpus-onboarding
gate.

## Reviewer's overarching frame, once more

> The most important remaining distinction is between discovering
> structure in research and rediscovering the decision rule that
> labels that structure `surviving`.

This validation is entirely on the "corpus provenance" side, not
either side of that distinction. It disambiguates whether identical
marginals reflect identical source data (yes, 12/13 sha256 shared)
or a coincidental histogram convergence (no).
