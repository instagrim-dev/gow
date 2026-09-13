# Shaping pack input contract (`shaping-pack/1`)

`newf shaping diagnose --pack-file <path>` executes an operator-authored pack
supplied as strict data instead of a built-in Go literal. This closes the
"packs are trusted Go code" authoring gap: a pack file is parsed, bounded, and
refused on ambiguity before any execution or receipt preparation, and it can
never raise its own evidence grade.

## Format

One JSON object, at most 1 MiB, valid UTF-8, no duplicate or case-variant
keys, no unknown fields, no trailing content.

```json
{
  "schema": "shaping-pack/1",
  "label": "development/example",
  "provenance": "who authored it, from what inputs, when",
  "episodes": [
    {
      "id": "unique-episode-id",
      "stratum": "history_informative",
      "family": "construction-family",
      "start": {"op": "not", "args": [{"op": "not", "args": [{"var": "x"}]}]},
      "variables": ["x"],
      "catalog": ["double-not"],
      "target_cost": 1,
      "history": [
        {
          "start": "(not (not x))",
          "rules_applied": ["double-not"],
          "final_cost": 1,
          "target": 1,
          "completed": true,
          "endpoint": "HOLDS_ON_DECLARED_DOMAIN"
        }
      ]
    }
  ]
}
```

- `label` and `provenance` are required. They are the caller's unverified
  claims and are retained verbatim in the receipt's `SourcePackLabel` and
  `SourceProvenance`.
- `stratum` must be one of `history_informative`, `history_low_value`,
  `history_misleading`.
- `start` uses the shared typed expression form from the finite claim
  contract: exactly one of `var`, `const`, or `op` plus `args`, under the same
  depth and node admission ceilings.
- `target_cost` and every history-attempt field are required explicitly. A
  missing field is a refusal, never a silently defaulted zero: a defaulted
  `completed: false` or `final_cost: 0` would shift selector preferences
  without an authored value.
- `family` and `history` are optional; an omitted history is an empty history.

## Validation ownership

The decoder owns format admission only. Domain validity, variable counts,
catalog-menu correspondence, duplicate episode IDs, node ceilings, and history
byte allowances are owned by the diagnostic runner — the same boundary that
validates the built-in packs. A well-formed file naming an unknown menu rule
decodes and is then refused by the runner, so there is exactly one semantic
validator.

## What a pack file cannot do

- **Upgrade evidence.** The runner forces
  `development/resource-diagnostic/freshness-unverified` on every receipt. A
  file labeled `agent-sealed/v1` keeps that string only as its recorded,
  unverified source claim.
- **Grant authority.** Executing a file pack is a development dispatch under
  the explicit resource vector. It creates no protected-batch, freshness, or
  spending status, and no completion arithmetic from it feeds any spending
  rule.
- **Reset exposure.** Renaming previously exposed task content does not make
  it fresh; freshness and custody are established by the separate sealed-run
  and dispatch-packet boundaries, not by pack authorship metadata.

`--pack` (built-in `smoke`/`development-v1`) and `--pack-file` are mutually
exclusive and one is required.
