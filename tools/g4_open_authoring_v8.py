#!/usr/bin/env python3
"""Open-only end-to-end G4 authoring-to-materializer delivery v8.

This program is restricted to public synthetic inputs and authorizes no
invocation by itself. Its caller supplies the endpoint, model, frozen limits,
and open output locations. It treats every response as data: incomplete,
malformed, or schema-invalid responses are retained in capture and never reach
staging.
"""
from __future__ import annotations

import argparse
import contextlib
import http.client
import json
import os
import signal
import subprocess
import threading
import time
from collections import Counter
from pathlib import Path
from typing import Any, Callable

from g4_sharded_unit_schema import ValidationError
from g4_open_authoring_schema_v8 import load_materialized, validate_intent, validate_materialized


class Reject(ValueError):
    pass


class MaterializationReject(Reject):
    def __init__(self, classification: str):
        super().__init__(classification)
        self.classification = classification


class DeadlineExceeded(TimeoutError):
    pass


@contextlib.contextmanager
def _hard_deadline(seconds: float):
    """Interrupt one bounded invocation at an absolute wall-clock deadline."""
    if seconds <= 0:
        raise DeadlineExceeded("absolute_deadline_exceeded")
    if threading.current_thread() is not threading.main_thread() or not hasattr(signal, "setitimer"):
        raise Reject("absolute_deadline_unavailable")
    previous_handler = signal.getsignal(signal.SIGALRM)
    previous_delay, previous_interval = signal.getitimer(signal.ITIMER_REAL)
    started = time.monotonic()

    def expire(_signum, _frame):
        raise DeadlineExceeded("absolute_deadline_exceeded")

    signal.signal(signal.SIGALRM, expire)
    signal.setitimer(signal.ITIMER_REAL, min(seconds, previous_delay) if previous_delay > 0 else seconds)
    try:
        yield
    finally:
        signal.setitimer(signal.ITIMER_REAL, 0)
        signal.signal(signal.SIGALRM, previous_handler)
        if previous_delay > 0:
            remaining = max(0.000001, previous_delay - (time.monotonic() - started))
            signal.setitimer(signal.ITIMER_REAL, remaining, previous_interval)


def _read_json(path: Path) -> Any:
    return json.loads(path.read_text(encoding="utf-8"))


def _write_json(path: Path, value: Any, *, exclusive: bool = False) -> None:
    flags = "x" if exclusive else "w"
    with path.open(flags, encoding="utf-8") as handle:
        json.dump(value, handle, sort_keys=True, separators=(",", ":"))
        handle.write("\n")


def _atomic_write_json(path: Path, value: Any) -> None:
    if path.exists():
        raise Reject("staging_path_already_exists")
    temporary = path.with_name("." + path.name + ".tmp")
    if temporary.exists():
        raise Reject("staging_temporary_path_exists")
    _write_json(temporary, value, exclusive=True)
    os.replace(temporary, path)


def _append_json_line(path: Path, value: Any) -> None:
    with path.open("a", encoding="utf-8") as handle:
        json.dump(value, handle, sort_keys=True, separators=(",", ":"))
        handle.write("\n")


def build_request(packet: Path, assignment: dict[str, str], model: str, max_tokens: int) -> dict[str, Any]:
    if not isinstance(model, str) or not model:
        raise Reject("invalid_model")
    if not isinstance(max_tokens, int) or isinstance(max_tokens, bool) or max_tokens <= 0:
        raise Reject("invalid_max_tokens")
    template = load_materialized(str(packet / "G4_OPEN_AUTHORING_UNIT_TEMPLATE_V8.json"), assignment)
    instruction = template.get("instruction")
    if not isinstance(instruction, str) or not instruction:
        raise Reject("invalid_unit_template")
    route_spec = (packet / "G4_OPEN_AUTHORING_CONFORMANCE_V8.md").read_text(encoding="utf-8")
    history_spec = (packet / "G4_OPEN_AUTHORING_PROTOCOL_V8.md").read_text(encoding="utf-8")
    envelope = (packet / "G4_OPEN_AUTHORING_ENVELOPE_V8.md").read_text(encoding="utf-8")
    return {
        "model": model,
        "stream": False,
        "keep_alive": "0s",
        "format": template["response_format"],
        "options": {"num_predict": max_tokens},
        "messages": [
            {
                "role": "user",
                "content": "\n".join(
                    [
                        "You are a bounded custodian. No tools, files, commands, browser, or network access are supplied.",
                        envelope,
                        instruction,
                        "Assigned unit: " + json.dumps(assignment, sort_keys=True, separators=(",", ":")),
                        route_spec,
                        history_spec,
                    ]
                ),
            }
        ],
    }


def extract_unit(raw: bytes) -> tuple[dict[str, Any], dict[str, int]]:
    try:
        envelope = json.loads(raw.decode("utf-8"))
        content = envelope["message"]["content"]
        if not isinstance(content, str):
            raise TypeError("content is not a string")
        unit = json.loads(content)
    except (UnicodeDecodeError, json.JSONDecodeError, KeyError, TypeError) as error:
        raise Reject("incomplete_or_malformed_provider_response") from error
    if not isinstance(unit, dict):
        raise Reject("unit_is_not_an_object")
    usage = {
        name: envelope[name]
        for name in ("prompt_eval_count", "eval_count")
        if isinstance(envelope.get(name), int) and not isinstance(envelope[name], bool) and envelope[name] >= 0
    }
    return unit, usage


def materialize_with_executable(unit: dict[str, Any], capture: Path, executable: Path, timeout_seconds: float) -> dict[str, Any]:
    if not executable.is_file():
        raise Reject("materializer_executable_missing")
    if timeout_seconds <= 0:
        raise Reject("materializer_process_timeout")
    intent_path = capture / "unit-intent.json"
    _write_json(intent_path, unit, exclusive=True)
    try:
        result = subprocess.run(
            [str(executable), "--json", "g4", "materialize-authoring-unit", "--input", str(intent_path)],
            stdin=subprocess.DEVNULL,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            check=False,
            timeout=min(30, max(0.001, timeout_seconds - min(0.01, timeout_seconds / 2))),
        )
    except subprocess.TimeoutExpired as error:
        raise Reject("materializer_process_timeout") from error
    if result.stderr:
        (capture / "materializer-stderr.txt").write_bytes(result.stderr)
    if result.returncode != 0:
        raise Reject("materializer_process_failed")
    try:
        response = json.loads(result.stdout)
    except (UnicodeDecodeError, json.JSONDecodeError) as error:
        raise Reject("materializer_response_malformed") from error
    _write_json(capture / "materializer-response.json", response, exclusive=True)
    if not isinstance(response, dict) or response.get("ok") is not True:
        failure = response.get("failure") if isinstance(response, dict) else None
        classification = failure.get("class") if isinstance(failure, dict) else "MATERIALIZATION_REJECTED"
        raise MaterializationReject(str(classification))
    materialized = response.get("unit")
    if not isinstance(materialized, dict):
        raise Reject("materializer_unit_missing")
    return materialized


def artifact_identity(executable: Path, path: Path, timeout_seconds: float = 30) -> dict[str, Any]:
    if not executable.is_file() or not path.is_file():
        raise Reject("identity_input_missing")
    try:
        result = subprocess.run(
            [str(executable), "--json", "g4", "artifact-identity", "--input", str(path)],
            stdin=subprocess.DEVNULL,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            check=False,
            timeout=timeout_seconds,
        )
    except subprocess.TimeoutExpired as error:
        raise Reject("identity_process_timeout") from error
    if result.returncode != 0:
        raise Reject("identity_process_failed")
    try:
        identity = json.loads(result.stdout)
    except (UnicodeDecodeError, json.JSONDecodeError) as error:
        raise Reject("identity_response_malformed") from error
    if (
        not isinstance(identity, dict)
        or set(identity) != {"schema", "sha256", "byte_length"}
        or identity.get("schema") != "g4-artifact-identity/1"
        or not isinstance(identity.get("sha256"), str)
        or len(identity["sha256"]) != 64
        or not isinstance(identity.get("byte_length"), int)
        or isinstance(identity["byte_length"], bool)
        or identity["byte_length"] <= 0
    ):
        raise Reject("identity_response_invalid")
    return identity


def validate_return(record: Any, *, plan_path: Path, assembled_path: Path, executable: Path) -> None:
    if not isinstance(record, dict) or set(record) != {"schema", "status", "unit_count", "plan", "assembled"}:
        raise Reject("return_shape_invalid")
    if record.get("schema") != "g4-open-authoring-return/1" or record.get("status") != "completed":
        raise Reject("return_status_invalid")
    assembled = _read_json(assembled_path)
    if record.get("unit_count") != assembled.get("unit_count"):
        raise Reject("return_unit_count_mismatch")
    if record.get("plan") != artifact_identity(executable, plan_path) or record.get("assembled") != artifact_identity(executable, assembled_path):
        raise Reject("return_identity_mismatch")


def invoke(
    *,
    packet: Path,
    plan_path: Path,
    assignment: dict[str, str],
    capture: Path,
    staging: Path,
    endpoint: str,
    model: str,
    max_tokens: int,
    timeout_seconds: int,
    materializer_executable: Path,
    connection_factory: Callable[..., Any] = http.client.HTTPConnection,
    materializer: Callable[[dict[str, Any], Path, Path, float], dict[str, Any]] = materialize_with_executable,
    wall_clock_seconds: float | None = None,
) -> int:
    if capture.exists():
        raise Reject("capture_path_already_exists")
    if not staging.is_dir():
        raise Reject("staging_directory_missing")
    if not isinstance(timeout_seconds, int) or isinstance(timeout_seconds, bool) or timeout_seconds <= 0:
        raise Reject("invalid_timeout")
    plan = _read_json(plan_path)
    assignments = _validate_plan(plan)
    if assignment not in assignments:
        raise Reject("assignment_not_in_frozen_plan")
    per_response = plan["per_response"]
    if max_tokens != per_response["max_tokens"]:
        raise Reject("max_tokens_does_not_match_frozen_plan")
    if timeout_seconds != per_response["max_elapsed_seconds"]:
        raise Reject("timeout_does_not_match_frozen_plan")
    effective_seconds = float(timeout_seconds if wall_clock_seconds is None else wall_clock_seconds)
    if effective_seconds <= 0 or effective_seconds > timeout_seconds:
        raise Reject("invalid_wall_clock_deadline")
    deadline_at = time.monotonic() + effective_seconds
    request = build_request(packet, assignment, model, max_tokens)
    response_schema = request["format"]
    capture.mkdir(mode=0o700, parents=True)
    _write_json(capture / "request.json", request, exclusive=True)
    connection = None
    try:
        with _hard_deadline(deadline_at - time.monotonic()):
            host, port_text = endpoint.rsplit(":", 1)
            port = int(port_text)
            connection = connection_factory(host, port, timeout=max(0.000001, deadline_at - time.monotonic()))
            connection.request(
                "POST",
                "/api/chat",
                body=json.dumps(request, separators=(",", ":")).encode("utf-8"),
                headers={"Content-Type": "application/json"},
            )
            response = connection.getresponse()
            raw = response.read()
    except (OSError, ValueError, http.client.HTTPException, Reject) as error:
        _write_json(
            capture / "receipt.json",
            {"schema": "g4-open-authoring-unit-invocation/8", "status": "transport_failed", "tool_schema_supplied": False, "host_tool_dispatch": "absent", "reason": type(error).__name__},
            exclusive=True,
        )
        return 2
    finally:
        if connection is not None:
            connection.close()
    (capture / "response.json").write_bytes(raw)
    receipt: dict[str, Any] = {
        "schema": "g4-open-authoring-unit-invocation/8",
        "status": "completed" if response.status == 200 else "http_failed",
        "http_status": response.status,
        "tool_schema_supplied": False,
        "host_tool_dispatch": "absent",
    }
    if response.status != 200:
        _write_json(capture / "receipt.json", receipt, exclusive=True)
        return 2
    try:
        with _hard_deadline(deadline_at - time.monotonic()):
            unit, usage = extract_unit(raw)
            validate_intent(unit, response_schema)
            materialized = materializer(unit, capture, materializer_executable, max(0.000001, deadline_at - time.monotonic()))
            validate_materialized(materialized, assignment)
    except (Reject, ValidationError, DeadlineExceeded) as error:
        receipt["status"] = "rejected_before_staging"
        receipt["reason"] = str(error).split(":", 1)[0]
        if isinstance(error, MaterializationReject):
            receipt["failure_class"] = error.classification
        _write_json(capture / "receipt.json", receipt, exclusive=True)
        return 2
    receipt["status"] = "validated_staged"
    receipt["usage"] = usage
    _atomic_write_json(staging / (assignment["id"] + ".json"), materialized)
    _write_json(capture / "receipt.json", receipt, exclusive=True)
    return 0


def _validate_plan(plan: Any) -> list[dict[str, str]]:
    if not isinstance(plan, dict) or plan.get("schema") not in {"g4-open-authoring-plan/8", "g4-open-authoring-live-plan/8"} or plan.get("unit_schema") != "g4-authoring-unit-intent/6" or plan.get("materialized_unit_schema") != "g4-authoring-unit-materialized/1":
        raise Reject("invalid_plan_schema")
    units = plan.get("units")
    expected_count = 24 if plan["schema"] == "g4-open-authoring-plan/8" else 3
    if not isinstance(units, list) or plan.get("unit_count") != len(units) or len(units) != expected_count:
        raise Reject("invalid_plan_population")
    expected = Counter({"history_informative": 12, "history_low_value": 6, "history_misleading": 6}) if expected_count == 24 else Counter({"history_informative": 1, "history_low_value": 1, "history_misleading": 1})
    if any(
        not isinstance(unit, dict)
        or set(unit) != {"id", "stratum", "family"}
        or not isinstance(unit["id"], str)
        or not unit["id"]
        or not isinstance(unit["stratum"], str)
        or not isinstance(unit["family"], str)
        or not unit["family"]
        for unit in units
    ):
        raise Reject("invalid_plan_assignment")
    if len({unit["id"] for unit in units}) != len(units) or Counter(unit["stratum"] for unit in units) != expected:
        raise Reject("invalid_plan_assignments")
    expected_families = (["open-inf-a"] * 6 + ["open-inf-b"] * 6 + ["open-low"] * 6 + ["open-misleading"] * 6) if expected_count == 24 else ["open-live-inf", "open-live-low", "open-live-misleading"]
    if [unit["family"] for unit in units] != expected_families:
        raise Reject("invalid_plan_family_allocation")
    per, aggregate = plan.get("per_response"), plan.get("aggregate")
    if not isinstance(per, dict) or not isinstance(aggregate, dict):
        raise Reject("invalid_plan_limits")
    required = ("max_tokens", "max_elapsed_seconds", "retries")
    if set(per) != set(required) or any(not isinstance(per[key], int) or isinstance(per[key], bool) for key in required):
        raise Reject("invalid_per_response_limits")
    if per["max_tokens"] <= 0 or per["max_elapsed_seconds"] <= 0 or per["retries"] != 0:
        raise Reject("unsupported_per_response_limits")
    required = ("max_invocations", "max_tokens", "max_elapsed_seconds")
    if set(aggregate) != set(required) or any(not isinstance(aggregate[key], int) or isinstance(aggregate[key], bool) for key in required):
        raise Reject("invalid_aggregate_limits")
    if aggregate["max_invocations"] < len(units) or aggregate["max_tokens"] < len(units) * per["max_tokens"] or aggregate["max_elapsed_seconds"] < len(units) * per["max_elapsed_seconds"]:
        raise Reject("aggregate_budget_exhausted")
    return units


def _validate_run_plan(plan: Any) -> list[dict[str, str]]:
    assignments = _validate_plan(plan)
    per, aggregate = plan["per_response"], plan["aggregate"]
    if plan["schema"] == "g4-open-authoring-plan/8":
        expected_assignments = ["unit-" + str(index).zfill(2) for index in range(24)]
        expected_order = "unit-00-through-unit-23"
        expected_per = {"max_tokens": 2048, "max_elapsed_seconds": 45, "retries": 0}
        expected_aggregate = {"max_invocations": 24, "max_tokens": 49152, "max_elapsed_seconds": 1080}
    else:
        expected_assignments = ["live-open-00", "live-open-01", "live-open-02"]
        expected_order = "live-open-00-through-live-open-02"
        expected_per = {"max_tokens": 1536, "max_elapsed_seconds": 90, "retries": 0}
        expected_aggregate = {"max_invocations": 3, "max_tokens": 4608, "max_elapsed_seconds": 270}
    if plan.get("generation_order") != expected_order or [assignment["id"] for assignment in assignments] != expected_assignments:
        raise Reject("generation_order_does_not_match_frozen_schedule")
    if per != expected_per or aggregate != expected_aggregate:
        raise Reject("aggregate_limits_must_equal_frozen_schedule")
    return assignments


def run(
    *,
    packet: Path,
    plan_path: Path,
    attempt: Path,
    progress_path: Path,
    stop_path: Path,
    endpoint: str,
    model: str,
    max_tokens: int,
    timeout_seconds: int,
    materializer_executable: Path,
    invoker: Callable[..., int] = invoke,
    assembler: Callable[..., dict[str, Any]] = None,
) -> int:
    if assembler is None:
        assembler = assemble
    plan = _read_json(plan_path)
    assignments = _validate_run_plan(plan)
    if max_tokens != plan["per_response"]["max_tokens"] or timeout_seconds != plan["per_response"]["max_elapsed_seconds"]:
        raise Reject("run_limits_do_not_match_frozen_plan")
    if attempt.exists() or progress_path.exists() or stop_path.exists():
        raise Reject("run_path_already_exists")
    if not progress_path.parent.is_dir() or not stop_path.parent.is_dir():
        raise Reject("run_log_parent_missing")
    attempt.mkdir(mode=0o700, parents=False)
    captures, staging, assignments_root = attempt / "captures", attempt / "staging", attempt / "assignments"
    for directory in (captures, staging, assignments_root):
        directory.mkdir(mode=0o700)
    aggregate_limit = plan["aggregate"]["max_elapsed_seconds"]
    started = time.monotonic()
    attempted = staged = 0

    def terminal(status: str, action: str | None, reason: str, *, assembled: bool = False) -> int:
        elapsed = round(time.monotonic() - started, 3)
        record: dict[str, Any] = {
            "schema": "g4-sharded-authoring-run/1",
            "status": status,
            "reason": reason,
            "attempted_invocations": attempted,
            "validated_staged_units": staged,
            "aggregate_elapsed_seconds": elapsed,
            "aggregate_elapsed_ceiling_seconds": aggregate_limit,
            "assembled_staging_created": assembled,
            "open_response_contents_inspected": False,
        }
        if action is not None:
            record["blocked_action"] = action
        _write_json(stop_path, record, exclusive=True)
        return 0 if status == "open_delivery_completed" else 2

    for assignment in assignments:
        elapsed_before = time.monotonic() - started
        if elapsed_before + timeout_seconds > aggregate_limit:
            _append_json_line(progress_path, {"schema": "g4-sharded-authoring-progress/1", "unit_id": assignment["id"], "status": "not_started", "reason": "aggregate_budget_insufficient_for_full_authorized_invocation", "attempted_invocations": attempted, "validated_staged_units": staged, "aggregate_elapsed_seconds": round(elapsed_before, 3), "open_response_contents_inspected": False})
            return terminal("verification_blocked", "BUDGET_EXHAUSTED", "aggregate_budget_insufficient_for_next_full_invocation")
        assignment_path = assignments_root / (assignment["id"] + ".json")
        _write_json(assignment_path, assignment, exclusive=True)
        capture = captures / assignment["id"]
        attempted += 1
        invocation_started = time.monotonic()
        remaining_aggregate = aggregate_limit - (invocation_started - started)
        code = invoker(packet=packet, plan_path=plan_path, assignment=assignment, capture=capture, staging=staging, endpoint=endpoint, model=model, max_tokens=max_tokens, timeout_seconds=timeout_seconds, materializer_executable=materializer_executable, wall_clock_seconds=min(float(timeout_seconds), remaining_aggregate))
        receipt_path = capture / "receipt.json"
        receipt = _read_json(receipt_path) if receipt_path.is_file() else {"status": "receipt_missing"}
        entry = {"schema": "g4-sharded-authoring-progress/1", "unit_id": assignment["id"], "status": receipt.get("status", "receipt_missing"), "reason": receipt.get("reason"), "usage": receipt.get("usage", {}), "attempted_invocations": attempted, "validated_staged_units": staged, "invocation_elapsed_seconds": round(time.monotonic() - invocation_started, 3), "aggregate_elapsed_seconds": round(time.monotonic() - started, 3), "open_response_contents_inspected": False}
        if code != 0 or receipt.get("status") != "validated_staged":
            _append_json_line(progress_path, entry)
            action = "TRANSPORT_FAILURE" if receipt.get("status") == "transport_failed" else "AUTHORING_OUTPUT_FAILURE"
            return terminal("verification_blocked", action, str(receipt.get("status", "receipt_missing")))
        staged += 1
        entry["validated_staged_units"] = staged
        _append_json_line(progress_path, entry)
    try:
        result = assembler(plan_path=plan_path, template_path=packet / "G4_OPEN_AUTHORING_UNIT_TEMPLATE_V8.json", staging=staging)
        assembled_path = attempt / "assembled-staging.json"
        _atomic_write_json(assembled_path, result)
    except (Reject, ValidationError, OSError, json.JSONDecodeError):
        return terminal("verification_blocked", "ASSEMBLY_VALIDATION_FAILURE", "assembly_rejected")
    _append_json_line(progress_path, {"schema": "g4-sharded-authoring-progress/1", "unit_id": None, "status": "assembled_staging", "attempted_invocations": attempted, "validated_staged_units": staged, "aggregate_elapsed_seconds": round(time.monotonic() - started, 3), "open_response_contents_inspected": False})
    try:
        returned = {
            "schema": "g4-open-authoring-return/1",
            "status": "completed",
            "unit_count": len(assignments),
            "plan": artifact_identity(materializer_executable, plan_path),
            "assembled": artifact_identity(materializer_executable, assembled_path),
        }
        return_path = attempt / "return.json"
        _atomic_write_json(return_path, returned)
        validate_return(returned, plan_path=plan_path, assembled_path=assembled_path, executable=materializer_executable)
    except (Reject, OSError, json.JSONDecodeError):
        return terminal("verification_blocked", "RETURN_VALIDATION_FAILURE", "return_rejected", assembled=True)
    _append_json_line(progress_path, {"schema": "g4-sharded-authoring-progress/1", "unit_id": None, "status": "return_validated", "attempted_invocations": attempted, "validated_staged_units": staged, "aggregate_elapsed_seconds": round(time.monotonic() - started, 3), "open_response_contents_inspected": False})
    return terminal("open_delivery_completed", None, "complete_population_assembled_identified_and_return_validated", assembled=True)


def assemble(*, plan_path: Path, template_path: Path, staging: Path) -> dict[str, Any]:
    plan = _read_json(plan_path)
    assignments = _validate_plan(plan)
    if not staging.is_dir():
        raise Reject("staging_directory_missing")
    files = sorted(staging.iterdir())
    expected_names = {assignment["id"] + ".json" for assignment in assignments}
    if {path.name for path in files} != expected_names:
        raise Reject("missing_or_unexpected_units")
    units: list[dict[str, Any]] = []
    for assignment in assignments:
        unit = _read_json(staging / (assignment["id"] + ".json"))
        try:
            validate_materialized(unit, assignment)
        except ValidationError as error:
            raise Reject("staged_unit_invalid") from error
        units.append(unit)
    return {
        "schema": "g4-open-authoring-assembled-staging/8",
        "unit_schema": "g4-authoring-unit-materialized/1",
        "unit_count": len(units),
        "publication": "eligible_for_full_host_validation_only",
        "units": units,
    }
def main() -> int:
    parser = argparse.ArgumentParser()
    commands = parser.add_subparsers(dest="command", required=True)
    invoke_parser = commands.add_parser("invoke")
    invoke_parser.add_argument("--packet", required=True, type=Path)
    invoke_parser.add_argument("--plan", required=True, type=Path)
    invoke_parser.add_argument("--assignment", required=True, type=Path)
    invoke_parser.add_argument("--capture", required=True, type=Path)
    invoke_parser.add_argument("--staging", required=True, type=Path)
    invoke_parser.add_argument("--endpoint", required=True)
    invoke_parser.add_argument("--model", required=True)
    invoke_parser.add_argument("--max-tokens", required=True, type=int)
    invoke_parser.add_argument("--timeout-seconds", required=True, type=int)
    invoke_parser.add_argument("--materializer-executable", required=True, type=Path)
    run_parser = commands.add_parser("run")
    run_parser.add_argument("--packet", required=True, type=Path)
    run_parser.add_argument("--plan", required=True, type=Path)
    run_parser.add_argument("--attempt", required=True, type=Path)
    run_parser.add_argument("--progress", required=True, type=Path)
    run_parser.add_argument("--stop", required=True, type=Path)
    run_parser.add_argument("--endpoint", required=True)
    run_parser.add_argument("--model", required=True)
    run_parser.add_argument("--max-tokens", required=True, type=int)
    run_parser.add_argument("--timeout-seconds", required=True, type=int)
    run_parser.add_argument("--materializer-executable", required=True, type=Path)
    assemble_parser = commands.add_parser("assemble")
    assemble_parser.add_argument("--plan", required=True, type=Path)
    assemble_parser.add_argument("--template", required=True, type=Path)
    assemble_parser.add_argument("--staging", required=True, type=Path)
    assemble_parser.add_argument("--out", required=True, type=Path)
    args = parser.parse_args()
    try:
        if args.command == "invoke":
            assignment = _read_json(args.assignment)
            if not isinstance(assignment, dict) or set(assignment) != {"id", "stratum", "family"}:
                raise Reject("invalid_assignment")
            return invoke(packet=args.packet, plan_path=args.plan, assignment=assignment, capture=args.capture, staging=args.staging, endpoint=args.endpoint, model=args.model, max_tokens=args.max_tokens, timeout_seconds=args.timeout_seconds, materializer_executable=args.materializer_executable)
        if args.command == "run":
            return run(packet=args.packet, plan_path=args.plan, attempt=args.attempt, progress_path=args.progress, stop_path=args.stop, endpoint=args.endpoint, model=args.model, max_tokens=args.max_tokens, timeout_seconds=args.timeout_seconds, materializer_executable=args.materializer_executable)
        if args.out.exists():
            raise Reject("assembly_output_already_exists")
        result = assemble(plan_path=args.plan, template_path=args.template, staging=args.staging)
        _atomic_write_json(args.out, result)
        return 0
    except (Reject, ValidationError, OSError, json.JSONDecodeError) as error:
        parser.error(str(error))
    return 2


if __name__ == "__main__":
    raise SystemExit(main())
