---
artifact_kind: prepared-future-g4-dispatch-authorization-template
status: prepared_not_authorized
prepared_date_utc: '2026-09-14'
scope: future protected G4-lite `/3` dispatch only
---

# G4 dispatch authorization template v1

This template has no authority while any placeholder remains. It does not
create a release pin, custody finding, execution permission, grading result,
spending approval, or publication permission. Before a dispatch, the operator
must fill every value from observed records, freeze this completed document,
and record its SHA-256 beside the custodian packet. Do not write that digest
inside this document. A missing, ambiguous, or mismatched field is a stop.

## Dispatch identity

| Field | Required filled value |
|---|---|
| Dispatch ID | `<dispatch-id>` |
| Effective operator reference and timestamp | `<operator-reference-and-rfc3339-timestamp>` |
| Named custodian and fresh-context identifier | `<custodian-id-and-fresh-context-id>` |
| Named substantive grader and permitted recipient | `<grader-id-and-recipient>` |
| Cancellation authority and signal | `<cancellation-authority-and-signal>` |

## Frozen public packet

| Artifact | Required filled identity |
|---|---|
| Release | `<full-commit-sha>`, build command, executable SHA-256, platform, and Go version |
| G4 contract | SHA-256 and exact supplied copy of `docs/g4-lite-pack-contract.md` |
| Public authoring specification | SHA-256 and exact supplied copy of `G4_PUBLIC_AUTHORING_SPEC_V1.md` |
| Open calibration receipt | SHA-256, byte length, and supplied path |
| Operator bindings | SHA-256, byte length, and supplied path for `operator-bindings.json` |
| Generation procedure | SHA-256, byte length, and supplied path |
| Resource ceiling | SHA-256, byte length, and supplied path |
| H0/H1/HG runtime identities | SHA-256 and byte length of each `/2` artifact, plus passed `g4 arm-preflight` output identity |
| Custodian brief | SHA-256 of the frozen `G4_CUSTODIAN_INTERFACE_BRIEF_V1.md` instance |
| Public interface | SHA-256 of the frozen `G4_CUSTODIAN_PUBLIC_INTERFACE_V1.md` instance |
| This authorization | Its SHA-256 is recorded beside the frozen document after every row is filled and before protected authoring begins; it is not a field inside this document |

The executable hash must equal each supplied arm identity's
`executable_sha256`. The resource ceiling and arm snapshots must be exactly the
preflighted bytes. A later rebuild, changed file, or changed runtime identity
requires a new authorization and preflight.

## Custody and storage boundary

- Public packet root: `<dispatch-root>/packet/`
- Empty custodian-controlled protected root: `<dispatch-root>/protected/`
- Content-free return path: `<approved-return-path>`
- Custody record and access-log path: `<custody-record-and-access-log-path>`
- Access controls and implementer exclusion: `<observed-controls-and-limitations>`
- Provider/network/package/browser permissions: `none`

The custodian receives only the frozen public packet, pinned executable, and
empty protected root. It must not access repository history, source, tests,
development packs, previous protected packs, answers, transcripts, or binary
inspection tools. Any boundary failure is retained as a stop, never turned
into a screen result.

## Permitted work and limits

The allowlisted commands are exactly those in the frozen public interface.
Set execution permission explicitly:

```text
execute_permitted: <true|false>
```

If `execute_permitted` is `false` or absent, the custodian may preflight,
author, validate, and seal but must stop before `g4 execute`. If it is `true`,
the authorization additionally names these fixed limits:

| Limit | Required filled value |
|---|---|
| Task-directed resource ceiling | `<exact-resource-ceiling-identity>` |
| Provider call ceiling and spend ceiling | `0 calls; 0 cents` |
| Separately metered custody ceiling | `<custody-metering-method-and-limit>` |
| Wall-clock ceiling | `<duration>` |
| Allowed output paths | `<execution-receipt-observed-metadata-binding-and-return-paths>` |
| Stop conditions | `<hash-mismatch-resource-truncation-cancellation-exposure-failure-missing-cost-and-other-listed-stops>` |

The authorization grants no permission to inspect binary contents, disclose
protected material, modify frozen public artifacts, classify a stop as a
screen outcome, decide spending, or publish a claim.

## Grader protected-evidence grant and completion handoff

Before execution, the operator creates a read-only protected-artifact mapping
for the named substantive grader and records its access-control mechanism and
locator here: `<grader-read-only-grant-and-locator>`. It must include the final
manifest, episode and answer artifacts, calibration and generation procedure,
arm snapshots, resource and run-design records, construction-route evidence,
custody/access records, execution receipt or interruption record, observed
metadata, execution binding, and resource ledger/result grid manifests. The
custodian cannot alter that mapping after the final manifest seal.

The grader's input is this authorized protected evidence. Its outgoing result
is only a content-free `g4-lite-substantive-grade/2` judgment. The custodian
also validates and sends its content-free `g4-custodian-return/1` record. No
return record establishes custody or grants a funding decision.
