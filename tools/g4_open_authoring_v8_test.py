#!/usr/bin/env python3
"""Open end-to-end G4 authoring-to-materializer delivery v8."""
from __future__ import annotations

import json
import os
import sys
import tempfile
import time
from pathlib import Path

root = Path(__file__).parents[1]
sys.path.insert(0, str(root / "tools"))
from g4_open_authoring_v8 import Reject, build_request, invoke, run, validate_return  # noqa: E402


RECIPES = [
    "commute-add-eliminate",
    "commute-xor-eliminate",
    "commute-add-double-not-eliminate",
    "commute-add-neg-neg-eliminate",
    "commute-add-or-self-eliminate",
    "commute-add-and-self-eliminate",
    "commute-add-mul-one-eliminate",
]


def intent_for(assignment: dict[str, str], recipe: str = RECIPES[0]) -> dict:
    return {
        "schema": "g4-authoring-unit-intent/6",
        "id": assignment["id"],
        "stratum": assignment["stratum"],
        "episode": {
            "id": assignment["id"],
            "stratum": assignment["stratum"],
            "family": assignment["family"],
            "variables": ["x"],
            "history_intents": [{
                "recipe_id": "commute-add-eliminate",
                "endpoint": {"var": "x"},
            }],
        },
        "answer": {
            "schema": "g4-custodian-answer/1",
            "id": assignment["id"],
            "endpoint": "HOLDS_ON_DECLARED_DOMAIN",
            "justification": "public synthetic selection",
        },
        "route": {
            "schema": "g4-custodian-route-selection/3",
            "id": assignment["id"],
            "recipe_id": recipe,
            "endpoint": {"var": "x"},
        },
    }


class Response:
    status = 200

    def __init__(self, unit: dict):
        self.payload = json.dumps({
            "message": {"content": json.dumps(unit)},
            "prompt_eval_count": 1,
            "eval_count": 20,
        }).encode()

    def read(self) -> bytes:
        return self.payload


class Connection:
    def __init__(self, unit: dict):
        self.response, self.closed = Response(unit), False

    def request(self, *_args, **_kwargs):
        pass

    def getresponse(self) -> Response:
        return self.response

    def close(self):
        self.closed = True


def factory_for(unit: dict):
    return lambda *_args, **_kwargs: Connection(unit)


executable_value = os.environ.get("NEWF_G4_MATERIALIZER")
if not executable_value:
    raise SystemExit("NEWF_G4_MATERIALIZER must name the built public materializer executable")
executable = Path(executable_value).resolve()

with tempfile.TemporaryDirectory() as temporary:
    work = Path(temporary)
    packet = work / "packet"
    packet.mkdir()
    for name in (
        "G4_OPEN_AUTHORING_UNIT_TEMPLATE_V8.json",
        "G4_OPEN_AUTHORING_ENVELOPE_V8.md",
        "G4_OPEN_AUTHORING_CONFORMANCE_V8.md",
        "G4_OPEN_AUTHORING_PROTOCOL_V8.md",
    ):
        source = root / "docs" / "plans" / name
        (packet / name).write_bytes(source.read_bytes())
    plan_path = root / "docs" / "plans" / "G4_OPEN_AUTHORING_PLAN_V8.json"
    plan = json.loads(plan_path.read_text())
    assignment = plan["units"][0]

    request = build_request(packet, assignment, "synthetic", 2048)
    request_text = request["messages"][0]["content"]
    assert "The host derives all starting expressions" in request_text
    assert "G4_ROUTE" not in request_text
    assert "route-recipe authoring supplement v5" not in request_text
    assert request["format"]["properties"]["episode"]["properties"]["family"]["const"] == assignment["family"]
    assert set(path.name for path in packet.iterdir()) == {
        "G4_OPEN_AUTHORING_UNIT_TEMPLATE_V8.json",
        "G4_OPEN_AUTHORING_ENVELOPE_V8.md",
        "G4_OPEN_AUTHORING_CONFORMANCE_V8.md",
        "G4_OPEN_AUTHORING_PROTOCOL_V8.md",
    }

    def invoke_case(label: str, unit: dict) -> tuple[int, dict, Path]:
        staging = work / (label + "-staging")
        staging.mkdir()
        capture = work / (label + "-capture")
        code = invoke(
            packet=packet,
            plan_path=plan_path,
            assignment=assignment,
            capture=capture,
            staging=staging,
            endpoint="127.0.0.1:1",
            model="synthetic",
            max_tokens=2048,
            timeout_seconds=45,
            materializer_executable=executable,
            connection_factory=factory_for(unit),
        )
        return code, json.loads((capture / "receipt.json").read_text()), staging

    for recipe in RECIPES:
        code, receipt, staging = invoke_case(recipe, intent_for(assignment, recipe))
        assert code == 0 and receipt["status"] == "validated_staged", (recipe, code, receipt)
        materialized = json.loads(next(staging.iterdir()).read_text())
        assert materialized["construction_boundary"] == "model_selected_typed_endpoints_and_versioned_recipes_host_derived_history_and_route_state_independent_endpoint_checks"
        assert materialized["episode"]["history"][0]["start"] == "add(0, x)"
        assert materialized["episode"]["history"][0]["rules_applied"] == ["add-comm", "add-zero"]
        assert len(materialized["route"]["steps"]) >= 2
        assert all(step["construction_origin"] == "host_derived_from_model_typed_endpoint_and_recipe" for step in materialized["route"]["steps"])

    invalid = intent_for(assignment)
    invalid["route"]["recipe_id"] = "invented-recipe"
    code, receipt, staging = invoke_case("unknown-recipe", invalid)
    assert code == 2 and receipt["status"] == "rejected_before_staging" and not list(staging.iterdir())

    malformed = intent_for(assignment)
    malformed["route"]["endpoint"] = {"op": "add", "args": [{"var": "x"}]}
    code, receipt, staging = invoke_case("malformed-route-endpoint", malformed)
    assert code == 2 and receipt["status"] == "rejected_before_staging" and not list(staging.iterdir())

    undeclared = intent_for(assignment)
    undeclared["episode"]["history_intents"][0]["endpoint"] = {"var": "y"}
    code, receipt, staging = invoke_case("undeclared-history-variable", undeclared)
    assert code == 2 and receipt["status"] == "rejected_before_staging" and not list(staging.iterdir())

    class MalformedResponse(Response):
        def __init__(self):
            self.payload = b'{"message":{"content":"{truncated"}}'

    malformed_capture_2 = work / "malformed-provider-capture-2"
    malformed_staging_2 = work / "malformed-provider-staging-2"
    malformed_staging_2.mkdir()

    class MalformedConnection(Connection):
        def __init__(self):
            self.response, self.closed = MalformedResponse(), False

    code = invoke(
        packet=packet,
        plan_path=plan_path,
        assignment=assignment,
        capture=malformed_capture_2,
        staging=malformed_staging_2,
        endpoint="127.0.0.1:1",
        model="synthetic",
        max_tokens=2048,
        timeout_seconds=45,
        materializer_executable=executable,
        connection_factory=lambda *_args, **_kwargs: MalformedConnection(),
    )
    malformed_receipt = json.loads((malformed_capture_2 / "receipt.json").read_text())
    assert code == 2 and malformed_receipt["status"] == "rejected_before_staging" and not list(malformed_staging_2.iterdir())

    class SlowResponse(Response):
        def read(self) -> bytes:
            time.sleep(2)
            return self.payload

    class SlowConnection(Connection):
        def __init__(self, unit: dict):
            self.response, self.closed = SlowResponse(unit), False

    slow_stream_staging = work / "slow-stream-staging"
    slow_stream_staging.mkdir()
    slow_stream_capture = work / "slow-stream-capture"
    slow_stream_started = time.monotonic()
    code = invoke(
        packet=packet,
        plan_path=plan_path,
        assignment=assignment,
        capture=slow_stream_capture,
        staging=slow_stream_staging,
        endpoint="127.0.0.1:1",
        model="synthetic",
        max_tokens=2048,
        timeout_seconds=45,
        wall_clock_seconds=0.05,
        materializer_executable=executable,
        connection_factory=lambda *_args, **_kwargs: SlowConnection(intent_for(assignment)),
    )
    slow_stream_receipt = json.loads((slow_stream_capture / "receipt.json").read_text())
    assert code == 2 and slow_stream_receipt["status"] == "transport_failed"
    assert slow_stream_receipt["reason"] == "DeadlineExceeded" and not list(slow_stream_staging.iterdir())
    assert time.monotonic() - slow_stream_started < 1

    sleeper = work / "slow-materializer.py"
    sleeper.write_text("#!/usr/bin/env python3\nimport time\ntime.sleep(2)\n")
    sleeper.chmod(0o700)
    slow_materializer_staging = work / "slow-materializer-staging"
    slow_materializer_staging.mkdir()
    slow_materializer_capture = work / "slow-materializer-capture"
    slow_materializer_started = time.monotonic()
    code = invoke(
        packet=packet,
        plan_path=plan_path,
        assignment=assignment,
        capture=slow_materializer_capture,
        staging=slow_materializer_staging,
        endpoint="127.0.0.1:1",
        model="synthetic",
        max_tokens=2048,
        timeout_seconds=45,
        wall_clock_seconds=0.2,
        materializer_executable=sleeper,
        connection_factory=factory_for(intent_for(assignment)),
    )
    slow_materializer_receipt = json.loads((slow_materializer_capture / "receipt.json").read_text())
    assert code == 2 and slow_materializer_receipt["status"] == "rejected_before_staging"
    assert slow_materializer_receipt["reason"] == "materializer_process_timeout" and not list(slow_materializer_staging.iterdir())
    assert time.monotonic() - slow_materializer_started < 1

    def full_invoker(*, assignment, capture, staging, **kwargs):
        recipe = RECIPES[int(assignment["id"].split("-")[1]) % len(RECIPES)]
        return invoke(
            assignment=assignment,
            capture=capture,
            staging=staging,
            connection_factory=factory_for(intent_for(assignment, recipe)),
            **kwargs,
        )

    attempt = work / "positive-attempt"
    progress, stop = work / "positive-progress.jsonl", work / "positive-stop.json"
    code = run(
        packet=packet,
        plan_path=plan_path,
        attempt=attempt,
        progress_path=progress,
        stop_path=stop,
        endpoint="127.0.0.1:1",
        model="synthetic",
        max_tokens=2048,
        timeout_seconds=45,
        materializer_executable=executable,
        invoker=full_invoker,
    )
    assert code == 0
    terminal = json.loads(stop.read_text())
    assert terminal["status"] == "open_delivery_completed" and terminal["attempted_invocations"] == 24 and terminal["validated_staged_units"] == 24
    assembled = json.loads((attempt / "assembled-staging.json").read_text())
    assert assembled["schema"] == "g4-open-authoring-assembled-staging/8" and assembled["unit_count"] == 24
    assert {unit["episode"]["family"] for unit in assembled["units"]} == {
        "open-inf-a", "open-inf-b", "open-low", "open-misleading",
    }
    returned = json.loads((attempt / "return.json").read_text())
    validate_return(returned, plan_path=plan_path, assembled_path=attempt / "assembled-staging.json", executable=executable)
    tampered = dict(returned)
    tampered["unit_count"] = 23
    try:
        validate_return(tampered, plan_path=plan_path, assembled_path=attempt / "assembled-staging.json", executable=executable)
        raise AssertionError("tampered return accepted")
    except Reject:
        pass

    interrupted_calls = 0

    def interrupted_invoker(*, assignment, capture, staging, **kwargs):
        global interrupted_calls
        interrupted_calls += 1
        if interrupted_calls == 6:
            capture.mkdir()
            (capture / "receipt.json").write_text(json.dumps({
                "schema": "g4-open-authoring-unit-invocation/8",
                "status": "transport_failed",
                "reason": "TimeoutError",
            }))
            return 2
        return invoke(
            assignment=assignment,
            capture=capture,
            staging=staging,
            connection_factory=factory_for(intent_for(assignment)),
            **kwargs,
        )

    interrupted_attempt = work / "interrupted-attempt"
    interrupted_progress, interrupted_stop = work / "interrupted-progress.jsonl", work / "interrupted-stop.json"
    code = run(
        packet=packet,
        plan_path=plan_path,
        attempt=interrupted_attempt,
        progress_path=interrupted_progress,
        stop_path=interrupted_stop,
        endpoint="127.0.0.1:1",
        model="synthetic",
        max_tokens=2048,
        timeout_seconds=45,
        materializer_executable=executable,
        invoker=interrupted_invoker,
    )
    assert code == 2 and not (interrupted_attempt / "assembled-staging.json").exists()
    assert len(list((interrupted_attempt / "staging").glob("*.json"))) == 5

print(json.dumps({
    "schema": "g4-open-authoring-delivery-regression/1",
    "schedule": {
        "invocations": 24,
        "per_tokens": 2048,
        "per_seconds": 45,
        "aggregate_tokens": 49152,
        "aggregate_seconds": 1080,
    },
    "recipes_materialized": RECIPES,
    "positive": "complete_24_unit_assembly_identity_and_return",
    "interruption": "five_staged_no_assembly",
    "deadlines": "slow_stream_and_slow_materializer_terminated_without_staging",
    "request_inputs": "v8_open_envelope_template_and_host_derived_state_contract_only",
}))
