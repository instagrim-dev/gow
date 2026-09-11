# Pilot 004 — interpretation notes (recorded before adjudication)

## Control framing (post-freeze review, 2026-09-11)

The D0 control must be reported precisely as:

> **Unperturbed notes versus outcome-permuted notes whose prose still
> retains the original outcome information.**

The protocol's "conservative control" phrasing asserted more than is
established. Two effects are consistent with the frozen design:

```text
The model follows the original prose:
    outcome permutation has little effect.

The model responds to contradictions between prose and metadata:
    output changes because the input is internally conflicting.
```

Therefore a future D1 > D0 score may reflect sensitivity to coherent versus
corrupted descriptions, not solely successful discovery of failure
structure. This is a competing explanation recorded IN ADVANCE of
adjudication, not an observed result. The frozen design is retained
unchanged; a positive score does not automatically discharge the protocol's
broader causal claim, and the eventual result record must carry this
limitation alongside any comparison.

## Validator lineage

The frozen checker (`tools/validate_discovery_wire.py`, digest pinned in
FREEZE.md) accepted quotations found anywhere in the bundle and counted
`shared_by` entries rather than distinct notes. Checker v2
(`tools/validate_discovery_wire_v2.py`) scopes quotations to the named
note's own section and requires distinct filenames.
`captures/supplementary-validation.json` records that all six sealed
captures, bytes unchanged (digests re-verified), PASS the stricter checker
— the v1 holes were real but unexploited by the actual captures. The frozen
validator's historical results stand; a quotation matching its note still
does not establish that it supports the proposed property — that remains
the operator's adjudication.
