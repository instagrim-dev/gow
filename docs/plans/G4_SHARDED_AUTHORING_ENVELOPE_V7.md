# G4 typed-history route-selection sharded authoring envelope v7

This public transport contract authorizes no model work by itself. A frozen
replacement authorization supplies the assignment, limits, endpoint, model,
materializer executable, and storage locations. The host sends one stateless
message for the assigned unit, supplies no tool schema, and never acts on a
model-emitted tool request.

The custodian returns one `g4-authoring-unit-intent/5` object. It selects
declared variables, a versioned recipe identifier, a compact endpoint term, a
fully typed nonempty history, and an answer justification. Each history
`start` is the recursive AST defined by the v7 typed-history specification,
not a compact expression string. The route schema is
`g4-custodian-route-selection/2`. The assigned id, stratum, and family are
frozen by the public plan and must be copied exactly. The answer endpoint is
the literal `HOLDS_ON_DECLARED_DOMAIN`.

The custodian does not emit an episode start term, rule catalog, target cost,
route steps, paths, bindings, premises, or intermediate states. The host
expands the selected recipe deterministically, reverse-constructs the episode
start, replays every step, and checks the endpoint on the declared finite
domain. The host also independently validates and deterministically renders
the history AST into the unchanged final history string without choosing or
repairing its content. Any omitted model field remains attributed to the host
rather than model capability.

The strict schema completely defines histories: every history item has a typed
expression-AST `start`, admitted `rules_applied`, nonnegative `final_cost` and
`target`, boolean `completed`, and nonempty `endpoint`. A rejected or
interrupted response remains capture only. A valid materialized unit is
protected staging only; it is not a pack, seal, evaluation, grade, spending
decision, or publication claim.
