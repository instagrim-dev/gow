#!/usr/bin/env python3
"""Validate the open G4 host-derived-state v8 contract."""
from __future__ import annotations

import json
import re
from pathlib import Path
from typing import Any

from g4_sharded_unit_schema import ValidationError, _validate_expression


_IDENTIFIER = re.compile(r"^[A-Za-z_][A-Za-z0-9_]*$")


_UNARY_OPERATORS = {"not", "neg", "shl1", "shr1"}
_BINARY_OPERATORS = {"and", "or", "xor", "add", "sub", "mul"}
_MAX_UINT64 = (1 << 64) - 1
_MAX_DEPTH = 64
_MAX_NODES = 4096


def _resolve_ref(schema: dict[str, Any], root: dict[str, Any]) -> dict[str, Any]:
    reference = schema.get("$ref")
    if reference is None:
        return schema
    if reference != "#/$defs/expression":
        raise ValidationError("unsupported schema reference")
    target = root.get("$defs", {}).get("expression")
    if not isinstance(target, dict):
        raise ValidationError("missing expression schema definition")
    return target


def _check(value: object, schema: dict[str, Any], path: str, root: dict[str, Any] | None = None) -> None:
    if root is None:
        root = schema
    schema = _resolve_ref(schema, root)
    if "oneOf" in schema:
        matches = 0
        for alternative in schema["oneOf"]:
            try:
                _check(value, alternative, path, root)
                matches += 1
            except ValidationError:
                pass
        if matches != 1:
            raise ValidationError(path + ": expected exactly one expression form")
        return
    kind = schema.get("type")
    if kind == "object":
        if not isinstance(value, dict):
            raise ValidationError(path + ": expected object")
        for key in schema.get("required", []):
            if key not in value:
                raise ValidationError(path + ": missing " + key)
        properties = schema.get("properties", {})
        additional = schema.get("additionalProperties", True)
        for key, child in value.items():
            if key in properties:
                _check(child, properties[key], path + "." + key, root)
            elif additional is False:
                raise ValidationError(path + ": unknown " + key)
            elif isinstance(additional, dict):
                _check(child, additional, path + "." + key, root)
        if len(value) < schema.get("minProperties", 0):
            raise ValidationError(path + ": too few properties")
    elif kind == "array":
        if not isinstance(value, list):
            raise ValidationError(path + ": expected array")
        if len(value) < schema.get("minItems", 0):
            raise ValidationError(path + ": too few items")
        if "maxItems" in schema and len(value) > schema["maxItems"]:
            raise ValidationError(path + ": too many items")
        for index, item in enumerate(value):
            _check(item, schema.get("items", {}), path + "[" + str(index) + "]", root)
    elif kind == "string":
        if not isinstance(value, str):
            raise ValidationError(path + ": expected string")
        if len(value) < schema.get("minLength", 0):
            raise ValidationError(path + ": empty string")
        if len(value.encode("utf-8")) > schema.get("maxLength", 1 << 60):
            raise ValidationError(path + ": string too long")
    elif kind == "integer":
        if not isinstance(value, int) or isinstance(value, bool):
            raise ValidationError(path + ": expected integer")
        if value < schema.get("minimum", -(1 << 63)) or value > schema.get("maximum", 1 << 63):
            raise ValidationError(path + ": integer outside range")
    elif kind == "boolean":
        if not isinstance(value, bool):
            raise ValidationError(path + ": expected boolean")
    if "const" in schema and value != schema["const"]:
        raise ValidationError(path + ": const mismatch")
    if "enum" in schema and value not in schema["enum"]:
        raise ValidationError(path + ": enum mismatch")


def validate_intent(unit: object, response_schema: dict[str, Any]) -> None:
    _check(unit, response_schema, "$")
    assert isinstance(unit, dict)
    variables = unit["episode"]["variables"]
    if len(set(variables)) != len(variables) or any(not _IDENTIFIER.fullmatch(name) for name in variables):
        raise ValidationError("$.episode.variables: invalid variable list")
    declared = set(variables)
    for index, item in enumerate(unit["episode"]["history_intents"]):
        _validate_history_expression(item["endpoint"], "$.episode.history_intents[" + str(index) + "].endpoint", declared)
    _validate_history_expression(unit["route"]["endpoint"], "$.route.endpoint", declared)


def _validate_history_expression(value: object, path: str, variables: set[str], depth: int = 0) -> int:
    if depth > _MAX_DEPTH:
        raise ValidationError(path + ": expression exceeds maximum depth 64")
    if not isinstance(value, dict):
        raise ValidationError(path + ": expected expression object")
    if set(value) == {"var"}:
        name = value["var"]
        if not isinstance(name, str) or not _IDENTIFIER.fullmatch(name) or len(name.encode("utf-8")) > 128:
            raise ValidationError(path + ": invalid variable identifier")
        if name not in variables:
            raise ValidationError(path + ": variable not declared")
        return 1
    if set(value) == {"const"}:
        constant = value["const"]
        if not isinstance(constant, int) or isinstance(constant, bool) or not 0 <= constant <= _MAX_UINT64:
            raise ValidationError(path + ": invalid constant")
        return 1
    if set(value) != {"op", "args"}:
        raise ValidationError(path + ": invalid expression shape")
    operator, args = value["op"], value["args"]
    if not isinstance(operator, str) or operator not in _UNARY_OPERATORS | _BINARY_OPERATORS:
        raise ValidationError(path + ": invalid expression operator")
    expected = 1 if operator in _UNARY_OPERATORS else 2
    if not isinstance(args, list) or len(args) != expected:
        raise ValidationError(path + ": invalid expression arity")
    nodes = 1
    for index, argument in enumerate(args):
        nodes += _validate_history_expression(argument, path + ".args[" + str(index) + "]", variables, depth + 1)
        if nodes > _MAX_NODES:
            raise ValidationError(path + ": expression exceeds 4096 nodes")
    return nodes


def canonicalize_history_start(value: object, variables: set[str]) -> str:
    """Validate and render one AST without rewriting, guessing, or fallback."""
    _validate_history_expression(value, "$.episode.history.start", variables)
    assert isinstance(value, dict)
    if "var" in value:
        return str(value["var"])
    if "const" in value:
        return str(value["const"])
    return str(value["op"]) + "(" + ", ".join(canonicalize_history_start(argument, variables) for argument in value["args"]) + ")"


def validate_materialized(unit: object, assignment: dict[str, str]) -> None:
    if not isinstance(unit, dict) or unit.get("schema") != "g4-authoring-unit-materialized/1":
        raise ValidationError("$: invalid materialized schema")
    if unit.get("endpoint_verification") != "HOLDS_ON_DECLARED_DOMAIN":
        raise ValidationError("$.endpoint_verification: independent endpoint did not hold")
    if unit.get("construction_boundary") != "model_selected_typed_endpoints_and_versioned_recipes_host_derived_history_and_route_state_independent_endpoint_checks":
        raise ValidationError("$.construction_boundary: invalid recipe-selection attribution")
    if unit.get("id") != assignment["id"] or unit.get("stratum") != assignment["stratum"]:
        raise ValidationError("$: materialized assignment mismatch")
    episode, answer, route = unit.get("episode"), unit.get("answer"), unit.get("route")
    if not all(isinstance(value, dict) for value in (episode, answer, route)):
        raise ValidationError("$: missing materialized subdocument")
    if episode.get("id") != assignment["id"] or episode.get("stratum") != assignment["stratum"] or answer.get("id") != assignment["id"] or route.get("id") != assignment["id"]:
        raise ValidationError("$: nested materialized assignment mismatch")
    if episode.get("family") != assignment["family"]:
        raise ValidationError("$.episode.family: materialized assignment mismatch")
    if answer.get("schema") != "g4-custodian-answer/1" or answer.get("endpoint") != "HOLDS_ON_DECLARED_DOMAIN" or not isinstance(answer.get("justification"), str) or not answer["justification"]:
        raise ValidationError("$.answer: invalid")
    variables = episode.get("variables")
    if not isinstance(variables, list) or not variables or len(variables) > 3 or len(set(variables)) != len(variables) or not all(isinstance(name, str) and _IDENTIFIER.fullmatch(name) for name in variables):
        raise ValidationError("$.episode.variables: invalid")
    declared = set(variables)
    previous = episode.get("start")
    start_nodes = _validate_expression(previous, "$.episode.start", declared)
    catalog = episode.get("catalog")
    if not isinstance(catalog, list) or not catalog or len(set(catalog)) != len(catalog) or not all(isinstance(rule, str) and rule for rule in catalog):
        raise ValidationError("$.episode.catalog: invalid")
    history = episode.get("history")
    history_keys = {"start", "rules_applied", "final_cost", "target", "completed", "endpoint"}
    if not isinstance(history, list) or not history:
        raise ValidationError("$.episode.history: invalid")
    for index, item in enumerate(history):
        path = "$.episode.history[" + str(index) + "]"
        if not isinstance(item, dict) or set(item) != history_keys:
            raise ValidationError(path + ": invalid shape")
        if not isinstance(item["start"], str) or not item["start"] or not isinstance(item["endpoint"], str) or not item["endpoint"]:
            raise ValidationError(path + ": invalid string")
        if not isinstance(item["rules_applied"], list) or not all(isinstance(rule, str) and rule for rule in item["rules_applied"]):
            raise ValidationError(path + ".rules_applied: invalid")
        if any(not isinstance(item[name], int) or isinstance(item[name], bool) or item[name] < 0 for name in ("final_cost", "target")):
            raise ValidationError(path + ": invalid cost")
        if not isinstance(item["completed"], bool):
            raise ValidationError(path + ".completed: invalid")
    steps = route.get("steps")
    if route.get("schema") != "g4-custodian-construction-route/2" or not isinstance(steps, list) or len(steps) < 2:
        raise ValidationError("$.route: invalid materialized selection")
    strict_reduction = False
    cost_neutral = False
    endpoint_nodes = start_nodes
    for index, step in enumerate(steps):
        path = "$.route.steps[" + str(index) + "]"
        if not isinstance(step, dict) or step.get("direction") != "forward" or step.get("construction_origin") != "host_derived_from_model_typed_endpoint_and_recipe":
            raise ValidationError(path + ": invalid host-expanded selection step")
        materialized_path = step.get("path")
        if materialized_path is None:
            materialized_path = []
        if not isinstance(materialized_path, list) or not all(isinstance(part, int) and not isinstance(part, bool) and part in (0, 1) for part in materialized_path):
            raise ValidationError(path + ".path: invalid")
        before_nodes = _validate_expression(step.get("before"), path + ".before", declared)
        after_nodes = _validate_expression(step.get("after"), path + ".after", declared)
        if step.get("before") != previous:
            raise ValidationError(path + ": route continuity mismatch")
        if after_nodes > before_nodes:
            raise ValidationError(path + ": recipe step grows")
        strict_reduction = strict_reduction or after_nodes < before_nodes
        cost_neutral = cost_neutral or after_nodes == before_nodes
        if step.get("rule") == "not-intro" or step.get("rule") not in catalog:
            raise ValidationError(path + ".rule: invalid materialized recipe rule")
        if step.get("premise_refs") != ["g4-menu-rule:" + str(step.get("rule"))]:
            raise ValidationError(path + ".premise_refs: invalid")
        previous = step.get("after")
        endpoint_nodes = after_nodes
    if not strict_reduction or not cost_neutral:
        raise ValidationError("$.route: selection must expand to neutral and reducing steps")
    target_cost = episode.get("target_cost")
    if not isinstance(target_cost, int) or isinstance(target_cost, bool) or not endpoint_nodes <= target_cost < start_nodes:
        raise ValidationError("$.episode.target_cost: outside endpoint/start node bounds")


def load_materialized(template_path: str, assignment: dict[str, str]) -> dict[str, Any]:
    if set(assignment) != {"id", "stratum", "family"} or not all(isinstance(value, str) and value for value in assignment.values()):
        raise ValidationError("invalid_assignment")
    replacements = {
        "<assigned-id>": assignment["id"],
        "<assigned-stratum>": assignment["stratum"],
        "<assigned-family>": assignment["family"],
    }

    def visit(value: object) -> object:
        if isinstance(value, str):
            return replacements.get(value, value)
        if isinstance(value, list):
            return [visit(item) for item in value]
        if isinstance(value, dict):
            return {key: visit(item) for key, item in value.items()}
        return value

    result = visit(json.loads(Path(template_path).read_text(encoding="utf-8")))
    if not isinstance(result, dict):
        raise ValidationError("invalid template")
    if "<assigned-" in json.dumps(result, sort_keys=True):
        raise ValidationError("unresolved_assignment_placeholder")
    return result
