# Pilot 001 — train-only mapping-and-abstraction review

Commissioned by `../records/STOP-readiness-2026-09-10.md` (revision 2).
This review exists because mining at min-support 2 produced zero candidates
and the recorded root cause is semantic, not mechanical: no surface label
recurs across two failure-side families, and only 2 of 85 decisive-axis
claims resolve under `mechanism/v1`.

## Scope and blinding

- **Inputs**: the twelve frozen train notes from pin `a868dec`
  (`corpus/train`, tree `da49ddcf…`) and the pre-capture snapshot digests in
  `../records/`. Nothing else.
- **The withheld target is out of scope.** No reviewer or proposing model
  may read `corpus/target/` or any artifact derived from it. A review
  session that has seen the target is disqualified from proposing or
  adjudicating entries.
- **Train-only includes contrast evidence.** Partial-success notes are in
  scope as contrast (e.g. `es-12`'s `congruence covering` under
  `partial_success`); entries must not be selected or rejected according to
  whether they help mining reach support.

## Roles

- **Proposer (model permitted)**: may propose relationships, extract
  supporting passages, identify counterexamples, and draft shared-property
  hypotheses. Proposals carry `status: proposed` and are never
  self-approved.
- **Adjudicator (operator)**: accepts, rejects, or defers each entry with a
  recorded reason. Only adjudicated entries have downstream effect.

## Relationship taxonomy (every entry declares exactly one kind)

| kind | claim | treatment on acceptance |
|---|---|---|
| `alias` | different labels mean the same operation in the relevant scope | approved alias in a **new pinned vocabulary revision** (never a silent mutation of `mechanism/v1`) |
| `component_of` / `specializes` | one operation is a constituent or specialization of another | identities preserved; relationship recorded; **no** support merge |
| `shared_property` | genuinely distinct operations share a restriction, assumption, or dependency | identities preserved; the property is recorded as a **source-grounded interpretation (hypothesis)**, encoded as its own canonical concept; its support must then be mined, evaluated and challenged like any candidate |
| `unresolved` | the evidence does not establish a relationship | left unresolved; no downstream effect |
| `distinct_on_tested_property` | the methods genuinely differ on the property being tested | distinction preserved as potential counterevidence; recorded so later candidates must face it |

"These methods share property P" is not the same claim as "these methods
are the same method." `alias` is the only kind that merges identities, and
it requires evidence of interchangeability in scope — co-occurrence on one
note is not sufficient (see `es-02`, whose `congruence covering` +
`class union` may be a method and one of its constituent operations).

## Ledger contract

Entries live in `adjudication-ledger.json`. Each entry records:

- `id`, `kind` (from the taxonomy), `labels` involved (with field kinds),
- `claim` — one sentence stating exactly what relationship is asserted,
- `scope` — where the claim is asserted to hold (and where it is not),
- `evidence` — source file + passage quotes (train notes only),
- `preserves_distinctions` — what the entry deliberately does NOT merge,
- `counterevidence` — known contrast occurrences, incl. partial-success uses,
- `status` — `proposed` | `accepted` | `rejected` | `deferred`,
- `adjudication` — operator reason, null until adjudicated.

## Freeze and downstream rules

1. **Decisions freeze before target recovery is revisited.** "Makes B3
   recover the target" must not become the criterion for accepting an
   entry; adjudication happens with the target unseen.
2. **Accepted aliases** publish as a new pinned vocabulary revision
   (`mechanism/v2` or scoped equivalent) with the ledger entry as
   provenance. `mechanism/v1` is immutable.
3. **Accepted shared properties** enter as hypotheses with
   `GeneratedInterpretation` provenance — they earn support only through
   the ordinary mine → challenge path, at the unchanged threshold.
4. **Reclustering is mandatory** after any accepted change, before support
   is counted again. The current seven failure-side groups are singleton
   isolates under a `degraded` comparison; they are not demonstrated
   mechanistic diversity and their count may not be carried forward.
5. Any accepted change ⇒ **new protocol revision** per RUNBOOK.md: new
   pinned hashes, fresh database, re-run readiness. Pilot 001's manifest
   stays null; a successor protocol inherits the design.

## What this review does and does not enable

Acceptance of entries here enables at most a **curated-feature pilot**:
reviewed mappings and challenged shared properties feed the deterministic
miner, and an external proposer tests whether those features improve
proposed directions. That is a legitimate smaller milestone. It does not
test the **discovery** claim — that a model can itself propose higher-order
shared properties from heterogeneous descriptions and have them survive
independent grounding and challenge. A discovery pilot needs its own
protocol; this review must not be retrofitted into evidence for it.
