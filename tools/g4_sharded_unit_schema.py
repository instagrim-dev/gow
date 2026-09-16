#!/usr/bin/env python3
"""Materialize and validate the public G4 sharded-authoring v2 response schema."""
from __future__ import annotations
import json
import re
from pathlib import Path

class ValidationError(ValueError): pass

_IDENTIFIER = re.compile(r'^[A-Za-z_][A-Za-z0-9_]*$')
_UNARY_OPERATORS = {'not', 'neg', 'shl1', 'shr1'}
_BINARY_OPERATORS = {'and', 'or', 'xor', 'add', 'sub', 'mul'}

def materialize(template: object, assignment: dict[str, str]) -> object:
    if set(assignment) != {'id', 'stratum'} or not all(isinstance(v, str) and v for v in assignment.values()):
        raise ValidationError('invalid_assignment')
    values = {'<assigned-id>': assignment['id'], '<assigned-stratum>': assignment['stratum']}
    def visit(value: object) -> object:
        if isinstance(value, str): return values.get(value, value)
        if isinstance(value, list): return [visit(x) for x in value]
        if isinstance(value, dict): return {k: visit(v) for k, v in value.items()}
        return value
    result = visit(template)
    if '<assigned-' in json.dumps(result, sort_keys=True): raise ValidationError('unresolved_assignment_placeholder')
    return result

def _check(value: object, schema: dict, path: str) -> None:
    kind = schema.get('type')
    if kind == 'object':
        if not isinstance(value, dict): raise ValidationError(path + ': expected object')
        for key in schema.get('required', []):
            if key not in value: raise ValidationError(path + ': missing ' + key)
        if schema.get('additionalProperties') is False:
            unknown = set(value) - set(schema.get('properties', {}))
            if unknown: raise ValidationError(path + ': unknown ' + sorted(unknown)[0])
        if len(value) < schema.get('minProperties', 0): raise ValidationError(path + ': empty object')
        for key, child in schema.get('properties', {}).items():
            if key in value: _check(value[key], child, path + '.' + key)
    elif kind == 'array':
        if not isinstance(value, list): raise ValidationError(path + ': expected array')
        if len(value) < schema.get('minItems', 0): raise ValidationError(path + ': too few items')
        if 'maxItems' in schema and len(value) > schema['maxItems']: raise ValidationError(path + ': too many items')
        for i, item in enumerate(value): _check(item, schema.get('items', {}), path + '[' + str(i) + ']')
    elif kind == 'string':
        if not isinstance(value, str): raise ValidationError(path + ': expected string')
        if len(value) < schema.get('minLength', 0): raise ValidationError(path + ': empty string')
    elif kind == 'integer':
        if not isinstance(value, int) or isinstance(value, bool): raise ValidationError(path + ': expected integer')
        if value < schema.get('minimum', -2**63): raise ValidationError(path + ': below minimum')
    if 'const' in schema and value != schema['const']: raise ValidationError(path + ': const mismatch')
    if 'enum' in schema and value not in schema['enum']: raise ValidationError(path + ': enum mismatch')

def _validate_expression(value: object, path: str, variables: set[str], depth: int = 0) -> int:
    if depth > 64:
        raise ValidationError(path + ': nesting exceeds 64')
    if not isinstance(value, dict):
        raise ValidationError(path + ': expected expression object')
    if set(value) == {'var'}:
        name = value['var']
        if not isinstance(name, str) or not _IDENTIFIER.fullmatch(name) or len(name.encode('utf-8')) > 128:
            raise ValidationError(path + ': invalid variable identifier')
        if name not in variables:
            raise ValidationError(path + ': variable not declared')
        return 1
    if set(value) == {'const'}:
        constant = value['const']
        if not isinstance(constant, int) or isinstance(constant, bool) or constant < 0:
            raise ValidationError(path + ': invalid constant')
        return 1
    if set(value) != {'op', 'args'}:
        raise ValidationError(path + ': invalid expression shape')
    operator, args = value['op'], value['args']
    if not isinstance(operator, str) or operator not in _UNARY_OPERATORS | _BINARY_OPERATORS:
        raise ValidationError(path + ': invalid expression operator')
    expected_arity = 1 if operator in _UNARY_OPERATORS else 2
    if not isinstance(args, list) or len(args) != expected_arity:
        raise ValidationError(path + ': invalid expression arity')
    nodes = 1
    for index, argument in enumerate(args):
        nodes += _validate_expression(argument, path + '.args[' + str(index) + ']', variables, depth + 1)
    if nodes > 4096:
        raise ValidationError(path + ': expression exceeds 4096 nodes')
    return nodes


def validate(unit: object, response_schema: dict) -> None:
    _check(unit, response_schema, '$')
    episode = unit['episode']
    variables = episode['variables']
    if len(set(variables)) != len(variables) or any(not _IDENTIFIER.fullmatch(name) or len(name.encode('utf-8')) > 128 for name in variables):
        raise ValidationError('$.episode.variables: invalid variable list')
    declared = set(variables)
    _validate_expression(episode['start'], '$.episode.start', declared)
    for index, step in enumerate(unit['route']['steps']):
        _validate_expression(step['before'], '$.route.steps[' + str(index) + '].before', declared)
        _validate_expression(step['after'], '$.route.steps[' + str(index) + '].after', declared)

def load_materialized(template_path: str, assignment: dict[str, str]) -> dict:
    return materialize(json.loads(Path(template_path).read_text()), assignment)
