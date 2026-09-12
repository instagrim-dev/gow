# (es-pvnp-independence) — Corpus dependence audit: seven dimensions, attributed dispositions

**Status**: authored 2026-09-12. Execution pin: `cbc3685a0a2bb668f59fc7985680f4ccdb665957`
(`origin/main` at audit time, verified unchanged during the audit).

This record discharges the **ES-vs-pvnp clause of correction 3**, left open by
[`l-coh-arc-consolidation.md`](l-coh-arc-consolidation.md) ("Source-byte
disjointness confirmed (necessary), but not sufficient") and listed as
`deferred` in [`CLOSURE-SCORECARD.md`](CLOSURE-SCORECARD.md) `(es-pvnp-independence)`.

It is an **audit of construction dependencies**, not an independence verdict
manufactured to close a row. The supported finding is **dependence**, and that
is a completed result, not a failure to reach one.

Machine-readable companion: [`es-pvnp-independence-audit.json`](es-pvnp-independence-audit.json).

---

## 1. The governing conclusion

> **The ES and P-vs-NP corpora draw on disjoint primary subject-matter sources
> — zero shared citations, zero shared cited authors, and code-verified
> SHA-256 disjointness across their retained source snapshots. They share the
> corpus author, the selection authority, the annotation template, the posture
> vocabulary, the outcome-label source, the decision rule, and the entire
> execution path. Independent corpus construction is therefore NOT established;
> what is established is independent *subject matter* under common
> construction.**

Two corpora can describe unrelated mathematics and still be one research
instrument applied twice. That is what the evidence supports here.

### Why "different hashes" was never sufficient

The prior evidence — `disjoint` at the source-byte layer — is real and is
retained. But the quantity it measures sits **upstream of every dependency that
matters for a replication claim**. The decisive structural fact:

> `internal/provider/fixture.go` derives each corpus's normalized mechanism axes
> by **parsing the corpus author's own embedded `newf-normalize` JSON**
> (`fixtureBlockStart = "<!-- newf-normalize"`). For both corpora, the
> representation the miner, challenger, and recovery rule operate on was
> hand-authored by the same author, in the same template, under the same
> vocabulary.

Source-byte disjointness does not touch that. Two disjoint source sets,
hand-annotated by one author into one schema, yield one annotation lineage.

---

## 2. Units being compared (A1)

Distinguishing corpora from database copies, per correction 3's warning against
counting four ES databases as four replications:

| Unit | ES | P-vs-NP |
|---|---|---|
| Substantive corpus lineage | 1 | 1 |
| Train notes | 12 (`corpus/train/es-01..12`) | 12 (`corpus/experiments/pvnp-holdout/train/pnp-01..12`) |
| Target notes | 1 (`corpus/target/es-target-affine-lattice-linear-forms.md`) | 2 (`pvnp-holdout/target/pnp-target-01,-02`) |
| Database instances | 4 (`.newf/{m7,pilot-001,pilot-002,pilot-003}`) | 1 (`.newf/pvnp`) |
| Independent research replications | **1** | **1** |

The four ES databases share 12/13 `source_snapshots.sha256` byte-for-byte
([`source-lineage-check-result.md`](source-lineage-check-result.md)) — one corpus
at four HEADs. **n = 2 corpus lineages, not 5.**

Neither corpus retains primary source bytes. Both are project-authored
summaries (`corpus/README.md`: "project-authored summaries for normalization,
not original mathematics and not verified proofs"; `pvnp-holdout/RESULT.md`:
"not raw historical text"). So "raw-source identity" is verified against
**stated citation anchors**, one authored layer removed from the literature.

---

## 3. Dependency matrix

Per-dimension dispositions are **audit-file labels only** — not domain enums,
not lifecycle states, and not collapsible into a numeric independence score.

| # | Dimension | Disposition | Basis |
|---|---|---|---|
| A | Raw-source identity | `distinct_origins_supported` (qualified) | 0 shared citations; but neither corpus retains source bytes, and ES has only 2 machine-resolvable identifiers across 13 files |
| B | Underlying work lineage | `distinct_origins_supported` | 0 shared authors across 49 surnames; 0 cross-citation; `disjoint` SHA-256 verdict |
| C | Corpus selection | `shared_dependency_observed` | Same sole git author; pvnp created 2 days after ES; **ES has no recorded selection criteria at all** |
| D | Annotation construction | `shared_dependency_observed` | Identical `normalize/v1` field structure (pvnp additive-only); near-identical prompts; `pvnp SCOPE.md:66` states pvnp follows `corpus/train/`'s shape |
| E | Representation / vocabulary | `shared_dependency_observed` | Same posture enums, same `mechanism/v*` seed file, same comparison profile sha `34b4bf21…` |
| F | Outcome / assessment | `shared_dependency_observed` | `recovery-rule/v1`, `leakage-check/v1`, same adjudication rubric; outcome labels author-supplied on both sides |
| G | Execution | `shared_dependency_observed` | One binary, one fixture normalizer parsing author-written JSON, one verb chain |

**Two dimensions support distinct origins. Five record shared dependency.** The
two that support distinctness are the two furthest from the decision rule.

### A — Raw-source identity

ES cites: Mordell (1969), Elsholtz & Tao (2011–2013), Browning & Elsholtz
(2010–2011), Vaughan (1970), Monks & Velingker, arXiv:2404.01508.
P-vs-NP cites: Baker–Gill–Solovay (1975), Razborov–Rudich (1994/1997),
Mulmuley–Sohoni (2001), Kabanets–Impagliazzo (2003), Håstad, Smolensky, and
others. **Intersection: empty.**

Qualifications that keep this from being strong evidence:

- ES has **2** machine-resolvable identifiers (arXiv IDs) across 13 files; the
  rest is prose in an `**Era:**` line that `corpus/README.md` explicitly
  declares "not audited" and carrying "no experimental weight."
- P-vs-NP corpus files contain **zero** arXiv/DOI/URL tokens; its DOIs live in
  `SCOPE.md` and two `date-source-*.json` records.
- Neither corpus records **which passages were actually read** during authoring.
  The sibling ES literature collection
  (`corpus/research/erdos-straus-2026-09-10/sources.json`) does record
  `read_scope` per source — and marks every entry
  `original_source_bytes_archived: false`, `annotation_status: curator_inferred`,
  `independent_verification: not_performed`. **No equivalent register exists for
  either benchmark corpus.**

### B — Underlying work lineage

Strongest independence evidence in the repository. Disjoint authorship,
no cross-citation in either direction (`rg 'Erd|erdos|residue'` over
`pvnp-holdout/train/*.md` → zero hits), and the code-verified verdict:

```text
Source lineage diff (verdict: disjoint)
  Left:  problem=prb_01M2B8ZMHRCJN4YVH4CP6SNXTD  store=.newf/m7/newf.db    snapshots=12
  Right: problem=prb_01M2BKVV5A7HD9HSAQKG7XEDST  store=.newf/pvnp/newf.db  snapshots=12
  shared:     0
```

Verified against *stated* anchors, not against what was consulted.

### C — Corpus selection

| Path | Creating commit | Date | Author |
|---|---|---|---|
| `corpus/README.md` | `8551dbe` | 2026-09-10 07:00 −0700 | `jmh <5620391+instagrim-dev@…>` |
| `corpus/train/` | `42353c0` (rename of `pre-cutoff/`) | 2026-09-10 07:37 −0700 | same |
| `pvnp-holdout/SCOPE.md` | `7c7d289` | 2026-09-12 12:28 −0700 | same |
| `pvnp-holdout/train/` | `4d80e41` | 2026-09-12 12:37 −0700 | same |

One git identity created both, two days apart, ES first.

The asymmetry is itself a finding. pvnp has an explicit preregistered selection
specification with numbered honesty constraints — the KI03 precursor deliberately
retained, algebrization deliberately absent, no post-2005 survey contamination —
plus a two-round blind contamination review. **ES has no inclusion/exclusion
criteria document.** Its nearest rationale is the origin commit body listing
twelve "mechanistically distinct approach families." No second party reviewed
either selection.

### D — Annotation construction

Extracted payload key sets are identical, pvnp additive-only:

| Level | ES (13 files) | P-vs-NP (14 files) |
|---|---|---|
| top | `approaches`, `schema_version` | same |
| approach | `description`, `label`, `logical_identity`, `mechanism`, `outcome`, `support` | same |
| mechanism | `assumptions`, `auxiliary_objects`, `breaks`, `construction_mode`, `locality`, `notes`, `operators`, `preserves`, `representations`, `uncertainty_mode` | same 10 **+** `completeness_basis`, `completeness_scope`, `field_completeness` |
| outcome | `boundary_conditions`, `boundary_statement`, `class`, `notes` | same |

All 27 files declare `"schema_version": "normalize/v1"`. The dependency is
stated in the repository itself:

> "Each file follows the `corpus/train/` shape: prose + `newf-normalize` payload
> with mechanism axes" — `pvnp-holdout/SCOPE.md:66`

Prompts are near-identical across corpora (`pilot-001/prompts/b0.md` vs
`pvnp-holdout/captures/b0-prompt.md`): same headings, same wire-schema
paragraphs, same constraint list, with the domain sentence swapped.

**Authoring agent identity: NOT RECORDED for either corpus.** ES attributes only
"project-authored"; pvnp only "Author-attested 2026-09-12", repeated verbatim in
all 14 files' `completeness_basis`. The repo *does* record this for downstream
capture steps (`pilot-003/captures/config.json` names proposer and
`model_identity`), which makes the absence at corpus-authoring time a gap in
provenance rather than a convention. **"Not recorded" is the audit finding.**

### E — Representation / vocabulary

Exact shared identifiers: `normalize/v1` (`internal/normalize/schema.go:20`);
the three posture axes with code-validated enums
(`internal/domain/normalize.go:336-342`); the `mechanism/v*` vocabulary family
seeded from one file (`internal/canon/vocabulary_seed.go`); and comparison
profile `classify/v3` with profile-sha
`34b4bf219dcacebd7c58b1917723d525f19491918e705357c5c6fd63c6ec91b2` appearing in
readiness records for **both** corpora.

Revision numbers differ (ES `mechanism/v3`, pvnp `mechanism/v6`); pvnp's SCOPE
pins v6 as a declared **superset** of the same lineage. The vocabulary family and
code path do not differ.

### F — Outcome / assessment

Outcome labels are author-supplied inside each note's own payload
(`outcome.class` with a self-declared `support` array), on both sides. No
independent labeller is recorded for either corpus.

Shared rule identifiers: `recovery-rule/v1`
(`internal/experiment/recovery.go:23`), `leakage-check/v1`
(`internal/store/experiment_store.go:245`), and an adjudication rubric that is
the same template with the domain criterion swapped — same 4-question
passage-level structure, same three-verdict enum, same mandatory
procedural-cue clause.

pvnp records the consequence explicitly; ES does not:

> "**Shared model family.** Proposers and adjudicators are instances of the same
> model family as the operator's session. Cross-instance agreement does not
> constitute independent-critic agreement in the strong sense."
> — `pvnp-holdout/RESULT.md:91-94`

### G — Execution

Both corpora: `--provider fixture` normalization (one implementation), the same
verb chain (`init → ingest → normalize → mechanism signature → cluster build →
failure-space build → invariants mine → challenge`), the same miner eligibility
code path (`derivePostureAxisProposals`), the same `mechanism/*` seed, the same
`classify/v3` profile, the same `recovery-rule/v1`.

Asymmetry: ES has a committed reproducible runner
(`m7-blinded-run/run.sh`); pvnp has none — its execution is reconstructable only
from JSON records plus the verb list in `l-pvnp-replication-result.md`. And the
five `.newf/*.db` stores backing the lineage evidence are **gitignored**, so
those SGO claims are not reproducible from a fresh clone.

---

## 4. What the shared dependencies constrain

Claims that the five shared dimensions **block**, until an independent
construction exists:

1. **"ES and pvnp are two independent research replications."** Not supported.
   They are two subject matters under one construction procedure.
2. **"Agreement between the corpora corroborates a domain-general property."**
   Constrained: shared annotator, template, vocabulary, and decision rule mean
   agreement can reflect the instrument rather than the domains. This is the
   direct application of correction 7's frame — *structure in research* versus
   *the decision rule that labels structure `surviving`*.
3. **"Divergence between the corpora reflects a domain difference."** Also
   constrained, in the opposite direction: divergence could reflect the
   vocabulary revision difference (`mechanism/v3` vs `/v6`) or the annotation
   choices of one authoring pass, not the mathematics.
4. **"The (l-pvnp-replication) partial-replication reading generalizes."** Its
   own record already caps this at PE tier and states "one non-ES corpus is one
   observation, not a distribution." This audit adds the reason that observation
   is not independent: same author, template, vocabulary, rule, and runner.

Claims **unaffected** by this audit:

- The source-byte `disjoint` verdict itself (SGO, retained).
- Every deterministic code-level result in the (l-coh) arc — correction 5's CMA
  fixture test concerns code behavior, not corpus provenance.
- pvnp's chronology discipline and leakage audit, which are internal to that
  corpus's construction.

---

## 5. What would establish independent construction

Named so the gap is actionable rather than rhetorical:

| Dimension | What would move it to `distinct_origins_supported` |
|---|---|
| C selection | Inclusion/exclusion criteria authored by a party who did not author the notes, recorded before source screening |
| D annotation | A different annotator (human or a recorded, differently-identified model session), with the authoring identity actually recorded |
| E representation | A schema/vocabulary not derived from `internal/normalize` + `internal/canon`, or the same corpus independently re-annotated under a second vocabulary |
| F assessment | Outcome labels supplied by someone other than the note author, ideally by a checker or a party without reference exposure |
| G execution | A second implementation cross-checking normalization and mining |

A third corpus does **not** discharge any of these for the existing pair. It
broadens the evidence base and carries its own dependency account. **Adding a
third corpus cannot retroactively make ES and pvnp independent of each other.**

---

## 6. Evidence tier and limits

**Tier: SGO** (source-grounded observation) for every disposition — each rests
on repository bytes, git history, or code identifiers cited above, not on model
judgment.

What this audit did **not** do:

- Did not rerun any historical pilot, mine, challenge, or evaluation.
- Did not modify frozen records, thresholds, policies, or vocabulary.
- Did not change any claim's tier in `CURRENT-CLAIMS.md` or the manuscript.
- Did not establish that the corpora are *dependent* in a way that invalidates
  their internal results — construction dependence constrains cross-corpus
  replication claims, not each corpus's own deterministic findings.
- Did not determine whether the shared-annotator dependency was intentional.
  It is recorded as observed, with the provenance gap named.

Unresolved after this audit:

1. Authoring-agent identity for both corpora (model/provider/session).
2. Which source passages were actually consulted during authoring.
3. Whether the pvnp author had ES payloads in context while writing — `SCOPE.md:66`
   makes it likely; nothing records it as fact.
4. Reproducibility of the lineage evidence from a fresh clone (`.newf/` is
   gitignored).

---

## 7. Provenance

- Dimension evidence: repository read at `cbc3685`, git history via
  `git log --diff-filter=A`, and code identifiers quoted inline.
- Prior records this audit builds on and does not replace:
  [`source-lineage-check-result.md`](source-lineage-check-result.md),
  [`source-lineage-admission-gate-result.md`](source-lineage-admission-gate-result.md),
  [`l-population-validation-result.md`](l-population-validation-result.md),
  [`l-pvnp-replication-result.md`](l-pvnp-replication-result.md),
  [`l-pvnp-replication-predeclaration.md`](l-pvnp-replication-predeclaration.md),
  [`l-coh-arc-consolidation.md`](l-coh-arc-consolidation.md).
- Historical prose in those records is preserved verbatim; where this audit
  reaches a stronger or narrower statement, it says so here rather than editing
  the source record.
