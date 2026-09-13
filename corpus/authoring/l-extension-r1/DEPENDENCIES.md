# DEPENDENCIES — how this corpus relates to ES and P-vs-NP

Companion to [`es-pvnp-independence-audit.md`](../../experiments/pilot-004-discovery/records/es-pvnp-independence-audit.md),
using that audit's seven dimensions so this corpus is assessed by the same
standard rather than a friendlier one.

**Headline:** source origins are expected disjoint from both existing lineages.
**Everything downstream of the sources is shared with them by construction.**
This corpus does not add an independent observation to the ES/pvnp pair.

---

## 1. Dimension-by-dimension

| # | Dimension | Disposition vs ES and pvnp | Basis |
|---|---|---|---|
| A | Raw-source identity | `distinct_origins_supported` | IACR/USENIX cryptanalysis publications; no source shared with either lineage. Verified by the byte comparison in `validation/`. |
| B | Underlying work lineage | `distinct_origins_supported` | No cited author appears in ES or pvnp. Stevens, Wang, Peyrin, Leurent, Joux, Manuel, Biham, Chabaud, Mironov, Zhang are absent from both. |
| C | Corpus selection | `shared_dependency_observed` | Selection rule authored by the same agent that authored the notes. No independent party screened sources. |
| D | Annotation construction | `shared_dependency_observed` | Same `normalize/v1` template, same field structure, same `completeness_scope`/`field_completeness` convention as pvnp. |
| E | Representation | `shared_dependency_observed` | Same three posture axes, same code-validated enums from `internal/domain/normalize.go`. No new vocabulary. |
| F | Outcome assessment | `shared_dependency_observed` (partly mitigated) | Outcome labels are author-supplied, as in ES and pvnp. **Mitigation:** for `sha1-11` the outcome is independently checkable — see §3. |
| G | Execution | `shared_dependency_observed` | Same binary, same `--provider fixture` normalizer parsing the author's own embedded JSON. |

**Two dimensions distinct, five shared.** The same ratio the audit found for the
ES/pvnp pair, for the same reason: the shared dependencies live downstream of
subject matter.

## 2. What this corpus cannot be used to claim

1. **That ES and pvnp are independent of each other.** A third corpus is
   irrelevant to that question. The audit names the five dimensions that would
   have to change for the existing pair, and none of them is affected by adding
   records here.
2. **That this corpus is an independent replication.** Dimensions C–G are shared
   with both predecessors. Agreement between this corpus and either of them could
   reflect the shared instrument — the same confound the audit identified.
3. **That a mining result here is prospective evidence.** The authoring agent
   knew the ES/pvnp mining outcomes before writing these records
   (`SCOPE.md §6.2`). Not outcome-blind. A held-out or historical-prediction
   reading requires a separately approved design with a different annotator.
4. **That n has increased to 3 independent observations.** It has increased to
   three *subject matters* under one construction procedure.

## 3. One genuine improvement, scoped narrowly

**Dimension F is partly mitigated, for one record only.** In ES and pvnp every
outcome label rests on the author's reading of the literature. Here, `sha1-11`'s
`outcome.class: success` is checkable independently of any annotation: the two
published PDFs both hash to SHA-1 `38762cf7f55934b34d179ae6a4c80cadccbb7f0a`
with distinct SHA-256 digests, verified by local computation and reproducible by
anyone with a hash utility.

That is a **verifier-grade outcome for one record**, not for the corpus. The
other eleven records' outcome classes remain author-supplied readings of primary
accounts, exactly as in the existing corpora. Note also that even the checkable
fact is narrow: it confirms *a collision exists*, not the mechanism attribution,
the boundary conditions, or the cost figures around it.

**Dimension D's provenance gap is closed for this corpus.** The audit found that
neither ES nor pvnp records its authoring model, provider, or session.
`manifest.json` `authoring_provenance` records this corpus's. That closes the gap
here and does nothing for them.

## 4. Adjacency worth stating

SHA-1 cryptanalysis and P-vs-NP both sit under theoretical computer science —
closer disciplinary neighbours than ES. No cited author, publication, or example
is shared, and the byte comparison in `validation/` is the evidence. But the
adjacency is recorded rather than suppressed, because an audit should be able to
find this stated by the author rather than discover it.

Concretely: `sha1-08` concerns SAT solving and `pnp-11` concerns derandomization
and polynomial identity testing. Both touch computational complexity. They cite
disjoint literatures and were written from disjoint sources, but a reviewer
asking "are these really unrelated fields?" deserves the honest answer that they
are adjacent fields with disjoint sources, not distant ones.

## 5. Where the shared dependencies would have to be broken

Inherited from the audit's table, unchanged, because they apply identically here:

- **C:** selection criteria authored by someone who did not author the notes.
- **D:** a different annotator, with identity recorded.
- **E:** a representation not derived from `internal/normalize` + `internal/canon`.
- **F:** outcome labels from a party without reference exposure — or, better,
  extended use of checkable artifacts as in §3.
- **G:** a second implementation cross-checking normalization.

The cheapest real improvement available in this domain is **F via §3**: SHA-1
cryptanalysis supplies checkable outcome objects, so a future design could
require artifact verification rather than annotator judgment wherever the sources
provide one. That is a property of the domain, not of this authoring pass.
