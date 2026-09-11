# Pilot 002 — protocol revision (successor to pilot-001)

Status: preparation. This revision exists because pilot-001 stopped at
readiness (`../pilot-001/records/STOP-readiness-2026-09-10.md`, revision 2)
and its mapping-and-abstraction review was adjudicated
(`../pilot-001/review/adjudication-ledger.json`, L1–L6 accepted;
Decision Handoff Packet validated). Per the freeze rules in
`../pilot-001/review/MAPPING-REVIEW.md`, an accepted vocabulary change
requires a new protocol revision: new pins, fresh database, reclustering,
readiness re-run.

## What changed relative to pilot-001 (and what did not)

Changed:

1. **Vocabulary: `mechanism/v2`** — a strict superset of the immutable
   `mechanism/v1`, adding three GeneratedInterpretation property concepts
   (ledger L1/L3/L4). No alias merges two operations; the only aliased
   source labels are ones that directly name a property (es-08 → L1,
   es-02 → L4). es-12's form-indexed variant deliberately does not resolve.
2. **Interpretation claims (migration v33)** — the accepted shared
   properties are attached to the mechanisms in each entry's adjudicated
   scope via `newf interpretation add`, with the ledger entry as required
   provenance. They enter signatures as `inferred` claims with
   `interpretation:` locators; frozen sources and per-field support rows are
   untouched.
3. **Contains-ordering fix** — verified presence via a resolved claim now
   beats unresolved co-claims (absence semantics unchanged). Recorded here
   because it changes evaluation decisiveness relative to the pilot-001
   executable.

Unchanged (deliberately):

- **Corpus pin**: the ledger accepted no corpus change, so the dataset is
  the same frozen train/target trees from `a868dec`
  (train tree `da49ddcf…`, target blob `7ed727ab…`).
- **Miner and threshold**: `DerivingFixtureInvariantMiner`, min-support 2.
  A readiness failure remains a stop, not a tuning instruction.
- **Design values**: arms, budgets (8/8), blinded mode, capture protocol,
  and interpretation discipline inherit from `../pilot-001/PROTOCOL.md`.
  This is still at most a **curated-feature pilot**: reviewed properties
  feed the deterministic miner; it does not test the discovery claim.

## Interpretation application manifest (operator-adjudicated scopes)

Applied to the TRAIN problem only, before signing. `field=preserves` for all.

| Ledger entry | Property label | Mechanisms (by logical identity) |
|---|---|---|
| L1 | `confined to quadratic nonresidues` | mordell-polynomial-identities (es-01), covering-system-attempt (es-02), factorization-scheme (es-03), vaughan-congruence-density (es-10) |
| L3 | `class union construction` | covering-system-attempt (es-02), vaughan-congruence-density (es-10) |
| L4 | `identity carried solvability` | mordell-polynomial-identities (es-01), vaughan-congruence-density (es-10) |

Not applied anywhere: es-08 (its own preserves label names L1 and resolves by
alias), es-02 for L4 (its own label resolves by alias), and every
partial-success note (contrast is measured, not manufactured). L2/L5/L6 are
non-merging relationship records; they add no claims.

The withheld target is signed under `mechanism/v2` with NO interpretation
claims: the target's recovery must come from the proposals, not from
operator annotation of the target.

## Stop conditions

Identical to pilot-001: a failed readiness check stops preparation; the
named blockers are preserved and any dataset/vocabulary change requires a
further protocol revision. Reclustering results are recorded verbatim —
degraded clustering remains degraded in the record even if counts pass.
