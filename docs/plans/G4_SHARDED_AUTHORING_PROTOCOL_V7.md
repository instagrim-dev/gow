---
artifact_kind: prepared-g4-typed-history-authoring-protocol
status: prepared_not_authorized
scope: public delivery protocol for a future replacement G4 authoring attempt
supersedes: G4_SHARDED_AUTHORING_PROTOCOL_V6.md for a newly authorized packet only
---

# G4 typed-history route-selection sharded authoring protocol v7

Version 7 retains the frozen 24-unit `24/2048/45/49152/1080` zero-retry
schedule and the v6 family allocation: units 00–05 use `protected-inf-a`,
06–11 use `protected-inf-b`, 12–17 use `protected-low`, and 18–23 use
`protected-misleading`.

Each invocation has one absolute 45-second wall-clock deadline covering the
request, complete response read, schema validation, materialization, and final
unit validation. The launcher caps that deadline by the remaining aggregate
1,080-second allowance. Deadline expiry closes the connection or terminates
the materializer, records the attempted invocation, stages nothing for that
unit, and stops the zero-retry run.

Each response is a complete `g4-authoring-unit-intent/5` object whose route is
`g4-custodian-route-selection/2`. The model selects one of the seven pinned
versioned recipes, an endpoint term, variables, a fully typed nonempty history,
and an answer justification. Every history `start` is a recursive expression
AST; raw compact strings are rejected. The schema binds the assigned id,
stratum, and family. It rejects model-authored episode start, catalog, target,
route steps, or rule metadata as unknown fields.

The transport loads only the v7 envelope, request template, route-selection
specification, and typed-history specification into the model request. The
forced response schema is the provider boundary. The host independently
validates exact AST shape, operators, arity, declared variables, constants,
depth, and nodes before invoking the pinned Go
`g4 materialize-authoring-unit` command. The materializer expands the selected
recipe to exact rules, paths, bindings, premises, catalog, target cost, episode
start, and states; canonicalizes the model-authored history AST to the
unchanged compact final history string; then replays the route and independently
checks the endpoint. There is no history fallback or inferred replacement.

Only a successfully materialized unit reaches protected staging. Assembly
requires exactly the frozen 24 assignments, revalidates every materialized
unit, and writes staging only. A missing, rejected, duplicated, or interrupted
unit prevents assembly. Existing full-pack semantic, endpoint, resource, and
execution checks remain authoritative.

The model remains responsible for the history AST, the remaining history
claims, recipe, endpoint, variables, and justification. The host owns AST
validation and serialization plus deterministic route expansion. Neither
responsibility may be attributed to the other. Packet preparation and public
regressions do not authorize protected authoring or evaluation.
