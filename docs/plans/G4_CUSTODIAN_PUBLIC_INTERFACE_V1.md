---
artifact_kind: prepared-future-g4-custodian-public-interface
status: prepared_not_authorized
prepared_date_utc: '2026-09-14'
scope: future protected G4-lite `/3` custodian only
paired_brief: G4_CUSTODIAN_INTERFACE_BRIEF_V1.md
---

# G4 custodian public interface v1

At dispatch, replace every angle-bracket value with an operator-supplied,
frozen value. This interface, its paired brief, and the supplied public
authoring specification are the custodian's complete workflow instructions.
They contain no protected episode or answer material. An omitted command,
field, or artifact is `interface_unrepresentable`.

## Supplied public packet

The operator supplies these exact files under `<dispatch-root>/packet/`:

- `newf` — the executable whose SHA-256 equals `<executable-sha256>`;
- `procedure.json` — `g4-lite-calibration-procedure/1` with the declared
  protected family prefix and primary resource vector;
- `resource-ceiling.json` — the primary `g4-resource-ceiling/1` vector;
- `h0-runtime-identity.json`, `h1-runtime-identity.json`, and
  `hg-runtime-identity.json` — executable-bound
  `g4-lite-arm-runtime-identity/2` records;
- `calibration-receipt.json`, `operator-bindings.json`, an exact copy of
  `g4-lite-pack-contract.md`, and `G4_PUBLIC_AUTHORING_SPEC_V1.md`; and
- this interface, the paired brief, and a frozen authorization produced from
  `G4_DISPATCH_AUTHORIZATION_TEMPLATE_V1.md`, naming their hashes,
  `<dispatch-id>`, `<custodian-id>`, `<output-root>`, the exact command below,
  resource/provider ceilings, stop conditions, cancellation rule, and grader.

The protected root is `<dispatch-root>/protected/`. The custodian may write
only there and to the approved content-free return path. The supplied authoring
specification is authoritative for the strict episode syntax, catalog, fixed
manifest values, and operator bindings; the custodian does not infer them from
the executable.

## Allowlisted commands

After the integrity hash check, the only executable commands are:

```text
<dispatch-root>/packet/newf --json g4 arm-preflight \
  --resource-ceiling <dispatch-root>/packet/resource-ceiling.json \
  --h0-snapshot <dispatch-root>/packet/h0-runtime-identity.json \
  --h1-snapshot <dispatch-root>/packet/h1-runtime-identity.json \
  --hg-snapshot <dispatch-root>/packet/hg-runtime-identity.json

<dispatch-root>/packet/newf --json g4 artifact-identity --input <dispatch-root>/protected/<bounded-artifact>.json

<dispatch-root>/packet/newf --json g4 pack validate --input <dispatch-root>/protected/g4-pack.json
<dispatch-root>/packet/newf --json g4 pack seal --input <dispatch-root>/protected/g4-pack.json --out <dispatch-root>/protected/g4-pack.seal.json
<dispatch-root>/packet/newf --json g4 pack inspect <dispatch-root>/protected/g4-pack.seal.json --input <dispatch-root>/protected/g4-pack.json

<dispatch-root>/packet/newf --json g4 execute \
  --manifest <dispatch-root>/protected/g4-pack.json \
  --episode-pack <dispatch-root>/protected/episodes.json \
  --resource-ceiling <dispatch-root>/packet/resource-ceiling.json \
  --h0-snapshot <dispatch-root>/packet/h0-runtime-identity.json \
  --h1-snapshot <dispatch-root>/packet/h1-runtime-identity.json \
  --hg-snapshot <dispatch-root>/packet/hg-runtime-identity.json \
  --generation-procedure <dispatch-root>/packet/procedure.json \
  --out <dispatch-root>/protected/execution-receipt.json

<dispatch-root>/packet/newf --json g4 pack bind-execution \
  --pre-execution-seal <dispatch-root>/protected/g4-pack.seal.json \
  --observed-metadata <dispatch-root>/protected/observed-metadata.json \
  --out <dispatch-root>/protected/execution-binding.json

<dispatch-root>/packet/newf --json g4 custodian-return validate \
  --input <dispatch-root>/protected/custodian-return.json
```

The authorization must separately name whether the `g4 execute` command is
permitted. Its absence means stop after sealing. A hash mismatch, resource
truncation, cancellation, unexpected artifact, exposure failure, or missing
metered cost is a stop; retain the affected receipt and do not classify it as a
screen result.

## Public schema requirements

The protected episode file is strict `shaping-pack/1`: `schema`, `label`,
`provenance`, and `episodes`. Each episode has `id`, `stratum`, `family`,
`start`, `variables`, `catalog`, and `target_cost`; included histories add
`start`, `rules_applied`, `final_cost`, `target`, `completed`, and `endpoint`.
The population is exactly 12 `history_informative`, 6 `history_low_value`, and
6 `history_misleading`, with at least two informative construction families.
The procedure's protected family prefix, minimum catalog-entry count, and
minimum expression-tree depth are mandatory; route construction evidence is
retained separately for the grader.

`g4-pack.json` is strict `g4-lite-pack/3` metadata. It references separately
sealed episode, answer, calibration, generation-procedure, arm-snapshot,
resource, and run-design artifacts by lower-case SHA-256, positive byte length,
and locator. It declares the fixed one-run design, exact spending arithmetic,
and custody limitations described by the frozen G4 contract. The executor
requires the procedure and all three `/2` arm identity bytes to match those
references before it opens a receipt.

`observed-metadata.json` is strict `g4-lite-observed-metadata/1`. It names the
pre-execution seal hash and bytes plus content references for the arm execution
manifest, resource ledger manifest, and result grid manifest. It never embeds
those private contents.

## Content-free return

Send only this content-free record to the named grader:

```json
{"schema":"g4-custodian-return/1","dispatch_id":"<dispatch-id>","release_revision":"<pinned-revision>","executable_sha256":"<pinned-sha256>","procedure":{"sha256":"<sha256>","byte_length":<positive>},"manifest":{"sha256":"<sha256>","byte_length":<positive>},"pre_execution_seal":{"sha256":"<sha256>","byte_length":<positive>},"execution_receipt":{"sha256":"<sha256>","byte_length":<positive>},"observed_metadata":{"sha256":"<sha256>","byte_length":<positive>},"execution_binding":{"sha256":"<sha256>","byte_length":<positive>},"completion_state":"<completed|execution_interrupted|resource_exhausted|verification_blocked|interface_unrepresentable>","custody_limitations":["<UPPERCASE_CONTENT_FREE_CODE>"],"blocked_actions":["<UPPERCASE_CONTENT_FREE_CODE>"]}
```

For a completed return, every listed artifact identity must be distinct. An
`execution_interrupted` return may omit the receipt only when interruption
prevents receipt publication; it retains the procedure, manifest, seal, and a
blocked-action code. For a pre-execution stop, retain the identities that
exist, omit unavailable artifact members, and use the matching completion
state. The custodian does not grade the run. The designated grader validates
its own `g4-lite-substantive-grade/2` return separately.
