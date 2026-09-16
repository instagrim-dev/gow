---
artifact_kind: prepared-g4-route-selection-specification
status: prepared_not_authorized
scope: public model-output contract for sharded authoring v7
---

# G4 versioned route-selection specification v7

The final `shaping-pack/1` episode format is unchanged. The intermediate
response is `g4-authoring-unit-intent/5`, and its route is
`g4-custodian-route-selection/2`. The model selects exactly one pinned recipe
and a compact endpoint term. The route contains exactly `schema`, `id`,
`recipe_id`, and `endpoint_term`.

The allowed recipe identifiers remain:

- `commute-add-eliminate`
- `commute-xor-eliminate`
- `commute-add-double-not-eliminate`
- `commute-add-neg-neg-eliminate`
- `commute-add-or-self-eliminate`
- `commute-add-and-self-eliminate`
- `commute-add-mul-one-eliminate`

The host supplies exact rules, order, paths, complete bindings, premise
references, catalog, target cost, episode start, and intermediate route states.
It does not search for a replacement recipe or endpoint. An unknown recipe is
rejected before materialization. Every materialized route step records
`construction_origin: host_expanded_from_model_recipe_selection`.

Version 7's separate typed-history specification changes only
`episode.history[].start`; it does not transfer route selection to the host.
Successful materialization establishes replayability under the pinned rewrite
engine. It does not establish history value, live-provider reliability, or
permission to execute a protected evaluation.
