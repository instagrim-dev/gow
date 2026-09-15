---
artifact_kind: prepared-g4-sharded-authoring-protocol
status: prepared_not_authorized
scope: public delivery protocol for future protected G4 authoring
---

# G4 sharded authoring protocol v1

Protected authoring is delivered as 24 independently valid units, not as JSON
continuations. The frozen authorization assigns every unit ID, stratum,
required artifact types, order, per-response ceiling, total invocation count,
total token ceiling, and elapsed-time ceiling before any invocation.

Each stateless request receives only the public packet, its assigned unit, and
hash-pinned predecessor identities explicitly named by the authorization. A
response ending for length, malformed JSON, a missing unit, or a unit that
fails its schema is a failed unit. It is never spliced into a later response.

## Unit contract

Each returned unit is strict `g4-authoring-unit/1` with exactly `schema`,
`id`, `stratum`, `episode`, `answer`, and `route`. Its three artifact records
must bind to the same assigned ID. Host records—not model declarations—supply
custody observations, invocation usage, artifact hashes, and final manifest
references.

## Staging and publication

The host retains validated units in protected staging. Staging is not an
accepted pack. It assembles only after verifying all 24 assigned IDs, the
12/6/6 stratum population, unique IDs, exact answer and route coverage,
cross-document bindings, aggregate budgets, and frozen resource bindings.
The assembled package then undergoes the existing full host validation and
sealing path. Any missing or inconsistent unit blocks publication and sealing.

This protocol changes delivery mechanics only. It does not change population,
scoring, resource semantics, or execution authority.
