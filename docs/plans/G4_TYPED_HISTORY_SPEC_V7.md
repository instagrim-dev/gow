---
artifact_kind: prepared-g4-typed-history-specification
status: prepared_not_authorized
scope: public model-output contract for sharded authoring v7
---

# G4 typed-history specification v7

Version 7 changes exactly one model-authored value from v6: every
`episode.history[].start` is an expression AST instead of a compact expression
string. The response schema is `g4-authoring-unit-intent/5`. The route is
`g4-custodian-route-selection/2`; its `recipe_id` and compact `endpoint_term`
contract is otherwise unchanged.

A history start has exactly one recursive form:

```text
{"var":"<declared-identifier>"}
{"const":<nonnegative-uint64>}
{"op":"<allowed-operator>","args":[<expression>]}
{"op":"<allowed-operator>","args":[<expression>,<expression>]}
```

The unary operators are `not`, `neg`, `shl1`, and `shr1`. The binary operators
are `and`, `or`, `xor`, `add`, `sub`, and `mul`. A host validator requires the
exact object shape, checks operator arity, requires every variable to appear in
the episode's declared variable list, bounds identifiers to 128 UTF-8 bytes,
accepts constants only in the unsigned 64-bit range, and enforces the existing
64-level and 4,096-node expression ceilings for the fixed four-bit domain.

After validation, the host renders the AST deterministically into the unchanged
compact history string consumed by the final episode contract. Rendering is
structural: it preserves operator and argument order and has no recovery,
guessing, rewrite, or fallback path. Distinct admitted ASTs therefore retain
distinct compact histories unless their exact tree, labels, and leaves are
identical.

All other history fields remain model-authored: `rules_applied`, `final_cost`,
`target`, `completed`, and `endpoint`. The host does not derive or repair those
claims. Successful AST validation and rendering establish representation
validity only; they do not establish that a historical route occurred or was
valuable.

The materialized unit records the boundary
`model_authored_typed_history_start_host_validated_and_canonicalized_model_selected_versioned_recipe_and_endpoint_host_expanded_route_only`.
Each route step continues to record
`host_expanded_from_model_recipe_selection`. This credits the model for the
typed history tree, recipe, endpoint, variables, remaining history claims, and
answer justification; it credits the host for history validation and
serialization and for the mechanical route expansion.
