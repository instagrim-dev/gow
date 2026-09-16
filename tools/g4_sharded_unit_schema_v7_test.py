#!/usr/bin/env python3
"""Public structural regressions for the G4 typed-history v7 schema."""
from __future__ import annotations

import copy
import json
import sys
from pathlib import Path

root = Path(__file__).parents[1]
sys.path.insert(0, str(root / "tools"))
from g4_sharded_unit_schema import ValidationError  # noqa: E402
from g4_sharded_unit_schema_v7 import (  # noqa: E402
    canonicalize_history_start,
    load_materialized,
    validate_intent,
)


def intent_for(assignment: dict[str, str], start: object | None = None) -> dict:
    if start is None:
        start = {"op": "add", "args": [{"const": 0}, {"var": "x"}]}
    return {
        "schema": "g4-authoring-unit-intent/5",
        "id": assignment["id"],
        "stratum": assignment["stratum"],
        "episode": {
            "id": assignment["id"],
            "stratum": assignment["stratum"],
            "family": assignment["family"],
            "variables": ["x"],
            "history": [{
                "start": start,
                "rules_applied": ["add-comm", "add-zero"],
                "final_cost": 1,
                "target": 1,
                "completed": True,
                "endpoint": "HOLDS_ON_DECLARED_DOMAIN",
            }],
        },
        "answer": {
            "schema": "g4-custodian-answer/1",
            "id": assignment["id"],
            "endpoint": "HOLDS_ON_DECLARED_DOMAIN",
            "justification": "public synthetic typed history",
        },
        "route": {
            "schema": "g4-custodian-route-selection/2",
            "id": assignment["id"],
            "recipe_id": "commute-add-eliminate",
            "endpoint_term": "x",
        },
    }


def schema_for(assignment: dict[str, str]) -> dict:
    return load_materialized(
        str(root / "docs" / "plans" / "G4_SHARDED_UNIT_REQUEST_TEMPLATE_V7.json"),
        assignment,
    )["response_format"]


assignments = [
    {"id": "unit-00", "stratum": "history_informative", "family": "protected-inf-a"},
    {"id": "unit-12", "stratum": "history_low_value", "family": "protected-low"},
    {"id": "unit-18", "stratum": "history_misleading", "family": "protected-misleading"},
]
for assignment in assignments:
    validate_intent(intent_for(assignment), schema_for(assignment))

assignment = assignments[0]
template = schema_for(assignment)
unit = intent_for(assignment)


def rejected(label: str, candidate: dict) -> None:
    try:
        validate_intent(candidate, template)
        raise AssertionError(label + " accepted")
    except ValidationError:
        pass


cases: list[tuple[str, callable]] = [
    ("raw-string", lambda c: c["episode"]["history"][0].__setitem__("start", "add(0, x)")),
    ("unknown-field", lambda c: c["episode"]["history"][0]["start"].__setitem__("left", {"var": "x"})),
    ("unknown-operator", lambda c: c["episode"]["history"][0].__setitem__("start", {"op": "divide", "args": [{"var": "x"}, {"const": 1}]})),
    ("wrong-unary-arity", lambda c: c["episode"]["history"][0].__setitem__("start", {"op": "not", "args": [{"var": "x"}, {"var": "x"}]})),
    ("wrong-binary-arity", lambda c: c["episode"]["history"][0].__setitem__("start", {"op": "add", "args": [{"var": "x"}]})),
    ("undeclared-variable", lambda c: c["episode"]["history"][0].__setitem__("start", {"var": "y"})),
    ("negative-constant", lambda c: c["episode"]["history"][0].__setitem__("start", {"const": -1})),
    ("boolean-constant", lambda c: c["episode"]["history"][0].__setitem__("start", {"const": True})),
    ("uint64-overflow", lambda c: c["episode"]["history"][0].__setitem__("start", {"const": 1 << 64})),
]
for label, mutate in cases:
    candidate = copy.deepcopy(unit)
    mutate(candidate)
    rejected(label, candidate)

deep: object = {"var": "x"}
for _ in range(66):
    deep = {"op": "not", "args": [deep]}
candidate = copy.deepcopy(unit)
candidate["episode"]["history"][0]["start"] = deep
rejected("depth-overflow", candidate)

large: object = {"var": "x"}
for _ in range(12):
    large = {"op": "add", "args": [copy.deepcopy(large), copy.deepcopy(large)]}
candidate = copy.deepcopy(unit)
candidate["episode"]["history"][0]["start"] = large
rejected("node-overflow", candidate)

for owner, field, value in (
    ("episode", "start_term", "add(0, x)"),
    ("episode", "catalog", ["add-comm", "add-zero"]),
    ("episode", "target_cost", 1),
    ("route", "steps", []),
    ("route", "rule", "add-zero"),
):
    candidate = copy.deepcopy(unit)
    candidate[owner][field] = value
    rejected("removed-" + field, candidate)

left = {"op": "add", "args": [{"const": 0}, {"var": "x"}]}
right = {"op": "add", "args": [{"var": "x"}, {"const": 0}]}
left_rendered = canonicalize_history_start(left, {"x"})
right_rendered = canonicalize_history_start(right, {"x"})
assert left_rendered == "add(0, x)"
assert right_rendered == "add(x, 0)"
assert left_rendered != right_rendered

print(json.dumps({
    "schema": "g4-typed-history-schema-regression/1",
    "cases": [
        "three_strata_identical_validation_path",
        "raw_string_rejected",
        "unknown_fields_and_operators_rejected",
        "wrong_arity_rejected",
        "undeclared_variables_rejected",
        "invalid_constants_rejected",
        "depth_and_node_overflow_rejected",
        "distinct_asts_have_distinct_canonical_histories_without_fallback",
        "removed_model_fields_rejected",
    ],
}))
