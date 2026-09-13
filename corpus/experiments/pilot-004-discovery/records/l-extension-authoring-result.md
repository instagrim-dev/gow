# (l-extension) — Authoring result: SHA-1 collision cryptanalysis candidate corpus

**Status**: authored 2026-09-12. Execution base `cbc3685a0a2bb668f59fc7985680f4ccdb665957`
(`origin/main` at start and unchanged at remote through this work).

Delivers the `(l-extension)` authoring path as a **candidate corpus**, and links
the companion `(es-pvnp-independence)` audit. Neither delivery is a replication,
a transfer experiment, or a lifecycle promotion.

Package: [`corpus/authoring/l-extension-r1/`](../../../authoring/l-extension-r1/README.md).
Audit: [`es-pvnp-independence-audit.md`](es-pvnp-independence-audit.md).

---

## 1. Four separate results, deliberately not merged

| Category | Status |
|---|---|
| **Authoring completion** | Complete: 12 Work records, 16-entry source register, dependency account |
| **Technical validation** | Passed: structural validation, two lineage comparisons, two controls, fixture normalization, 26 validator regressions, repository gates |
| **Research evidence** | **Not executed.** No mining, challenge, frontier generation, or evaluation |
| **Authorization** | **Unchanged.** `execution_authorization: null`; no gated path touched |

Success in the first two columns is not success in the last two.

## 2. Domain and selection

**Domain:** SHA-1 collision cryptanalysis. **Objective:** produce a collision for
full, unmodified SHA-1. **Span:** 1998–2020.

The selection rule was fixed **before** sources were chosen (seven
evidence-availability criteria, `SCOPE.md §1`), with expected mining results
explicitly excluded from the basis. Shortlist of three; Fermat's Last Theorem
rejected for lacking primary scoped accounts and for number-theory adjacency to
ES, Navier–Stokes rejected because an open problem yields an all-failure corpus
by domain structure rather than by curation.

Two properties decided it:

1. **Checkable outcomes.** A collision is a concrete object. The 2017 success was
   verified locally: both published PDFs hash to SHA-1
   `38762cf7f55934b34d179ae6a4c80cadccbb7f0a` with distinct SHA-256 digests. This
   is the first GoW corpus record whose outcome does not rest on annotator
   judgment.
2. **A real retraction.** ePrint 2009/259 was withdrawn by its own authors and
   its page still serves the retraction note. That is a failure record produced
   by the field's error-correction, not by a curator's decision.

## 3. Records and outcome spread

12 records, 13 distinct external work origins plus one artifact pair.

| Outcome class | Count | Records |
|---|---|---|
| `success` | 1 | `sha1-11` |
| `partial_success` | 6 | `sha1-03`, `-04`, `-05`, `-09`, `-10`, `-12` |
| `partial_failure` | 2 | `sha1-02`, `sha1-06` |
| `failure` | 3 | `sha1-01`, `sha1-07`, `sha1-08` |

**The distribution is natural, not balanced.** A fixed eligibility rule was
applied uniformly; no record was relabelled and no success was dropped to shape a
cell. Two failures are structurally informative rather than merely negative:

- **`sha1-07`** (withdrawn 2^52 path) fails on an **imported premise**, not on its
  own construction. The boomerang placement was never shown wrong; a borrowed cost
  evaluation from `sha1-06` was. This is the cleanest available instance of a
  complexity claim inheriting the fragility of an unverified input.
- **`sha1-08`** (SAT-encoded search) is a **generality control**. It discards the
  differential-path machinery every other record preserves, succeeds on MD4 and
  MD5, and cannot reach SHA-0 at an estimated 3 million CPU hours. It shows the
  preserved machinery is load-bearing rather than incidental.

## 4. Five factual corrections applied during source verification

Recorded because each would have produced a wrong record, and because the
correction rate is itself a fact about authoring from memory versus verification:

1. **Biham–Chen 2004** gives 65-round **SHA-0** collisions, *not* reduced-round
   SHA-1 collisions. Reduced SHA-1 belongs to a different paper, not in this corpus.
2. **The 2^63 announcement** is **Wang, Andrew Yao & Frances Yao** — not
   Wang–Yin–Yu. It was never published as a full paper, so under the eligibility
   rule it gets **no Work record**; it is retained as register context only.
3. **Nossum 2012** is a **preimage** master's thesis whose contribution is an
   encoding improvement. The citable SAT negative result is **Mironov–Zhang,
   SAT 2006**. Nossum is retained as an explicit register exclusion.
4. **SHAttered SHA-256 digests** are **locally computed**; no authoritative page
   publishes them.
5. **Chabaud–Joux's 2^61** is scoped by the abstract to the **compression
   function**, not the full hash. The paper also does not itself attribute SHA-1's
   resistance to the message-expansion rotation; that attribution is marked
   `inferred`.

Additionally, four unreconciled discrepancies are recorded rather than smoothed:
Manuel's ePrint and journal versions reach different conclusions; SHAttered
reports 100 GPU-years in the paper versus 110 in the vendor announcement;
Leurent–Peyrin estimate ~$45,000 against ~$75,000 actually spent; and
`shattered.io` is now a repurposed domain whose prose must be cited from the
Wayback capture.

## 5. Validation executed

All in **disposable databases** with explicit `--db` paths. No live `.newf/` store
read or written. Full record: `validation/RESULTS.md`.

| Comparison | Verdict | shared | left-only | right-only |
|---|---|---|---|---|
| new vs ES train | `disjoint` | 0 | 12 | 12 |
| new vs pvnp train | `disjoint` | 0 | 12 | 12 |
| ES vs pvnp train (**control**) | `disjoint` | 0 | — | — |
| new vs empty (**control**) | `empty` | — | — | — |

Populations are train inventories only on all three sides; targets excluded.

The two controls matter as much as the results. The ES-vs-pvnp control reproduces
the previously recorded verdict in a freshly built store, showing the harness is
not returning `disjoint` indiscriminately. The empty-population control returns
`empty` — inconclusive, never an approval.

`disjoint` supports **non-overlapping source-byte sets in the compared
inventories**. It is not research-methodological independence. No content was
reformatted or renamed to influence any verdict.

Fixture compatibility: 12/12 records normalized under `--provider fixture`, one
approach each. **Stopped there** — no mining or challenge, which would convert a
compatibility check into an unauthorized research run and would invite amending
source selection in response to lifecycle labels.

Repository gates: `go build ./...`, `go vet ./...`, `go test ./...` all pass;
`gofmt -l .` empty; `git diff --check` clean; 26 validator regressions pass.

## 6. Dependency account for the new corpus

Assessed by the **same seven dimensions** as the ES/pvnp audit, not a friendlier
standard. Full account: `DEPENDENCIES.md`.

**Two dimensions distinct (A raw-source identity, B work lineage); five shared by
construction (C selection, D annotation, E representation, F outcome assessment,
G execution).** Same ratio as the existing pair, for the same reason: the shared
dependencies live downstream of subject matter.

One genuine improvement, scoped narrowly: **dimension F is partly mitigated for
one record**. `sha1-11`'s `success` label is checkable by anyone with a hash
utility. The other eleven remain author-supplied readings. Even the checkable
fact is narrow — it confirms a collision exists, not the mechanism attribution or
cost figures around it.

**Dimension D's provenance gap is closed here only.** `manifest.json`
`authoring_provenance` records the authoring model and session — the gap the audit
found in ES and pvnp. Closing it for this corpus does nothing for them.

## 7. What this delivery does not establish

- **Not** independence between ES and pvnp. A third corpus is irrelevant to that
  question; the audit names the five dimensions that would have to change for the
  existing pair, and none is affected by adding records here.
- **Not** an independent replication. Dimensions C–G are shared with both
  predecessors, so agreement with either could reflect the shared instrument.
- **Not** prospective evidence. The authoring agent authored the dependency audit
  in the same session and knew the ES/pvnp mining outcomes before writing these
  records. **Not outcome-blind**, recorded in `SCOPE.md §6` and `manifest.json`.
- **Not** an increase to n=3 independent observations. It is three subject
  matters under one construction procedure.
- **Not** a PE→CMA/SGO upgrade for any coherence reading, and not cross-domain
  transfer of any invariant.

## 8. Outstanding, and the specific next check

**The semantic support review is the one substantive obligation.** Structural
validation cannot establish that a cited passage supports the claim attributed to
it; recording an unsupported claim as a linter success is precisely the failure
mode `sha1-07` documents. `validation/REVIEW.md` lists per-record obligations,
prioritizing the records whose load-bearing fields are `inferred` rather than
`explicit`, plus the seven source-level limitations.

**Next specific decision required of the operator:** whether to commission the
semantic support review of the 12 records — and if the corpus is ever to serve as
research evidence rather than candidate material, whether to commission
**re-annotation by a different, recorded annotator without exposure to the
ES/pvnp results**. That single change converts dimensions C, D, and F from
`shared_dependency_observed` to potentially distinct, and it is the only
available step that would make a third corpus an independent observation rather
than a third application of the same instrument.

Not requested and not authorized here: mining, challenge campaigns, frontier
generation, evaluation, pilot freeze, or any lifecycle promotion.

## 9. Gated paths inventoried, untouched

`(k-full)`, `(p)`, `(pilot-005-freeze)`, `(coherence-metric-pipeline)` — existing
blockers unchanged; no rerun, no policy change, no open decision chosen, no
threshold or metric added.

Note for the record: a **concurrent writer** committed preservation-pilot
execution and review-remediation work onto this same branch during this session
(`f95821f`, `d847fb0`, `b2da4b4`, `6327d57`, `7c51495`). That work is not part of
this delivery and was neither reviewed nor modified here. This delivery touched
only `corpus/authoring/l-extension-r1/`, the two audit files, `.gitignore`, and
this record.
