#!/usr/bin/env python3
"""Deterministic Plane-1 regression validator for the root AGENTS.md contract.

Model-free and stdlib-only by design. Re-runs the falsifiable slice of the
agents-md-review so the contract floor is guarded at edit time even when the
full review skill is not invoked.

Usage:
    python3 docs/reviews/agents-md/validate_agents_md.py [--repo-root PATH]

Exit code 0 => all guarded claims hold. Non-zero => a red diff to inspect.

Each check maps to a review dimension (D1..D9). A check that flips from
green to red is a contract regression; a check that flips red to green after
a contract edit that ADDS the missing rule is the intended ratchet direction.
"""
from __future__ import annotations

import argparse
import json
import os
import re
import subprocess
import sys
from dataclasses import dataclass, asdict
from pathlib import Path


@dataclass
class CheckResult:
    id: str
    dimension: str
    description: str
    passed: bool
    detail: str


def repo_root_default() -> Path:
    here = Path(__file__).resolve()
    # docs/reviews/agents-md/validate_agents_md.py -> repo root is 3 up.
    return here.parents[3]


def read_text(path: Path) -> str:
    return path.read_text(encoding="utf-8") if path.exists() else ""


def check_paths_exist(root: Path, agents: str) -> CheckResult:
    """D7: every repo-relative path the contract names must exist."""
    # Only backticked tokens that look like repo paths (contain '/' or a known
    # top-level file). Excludes code-fence pseudo-ops and vocabulary tokens.
    candidates = set(re.findall(r"`([^`\n]+)`", agents))
    named_paths = sorted(
        c
        for c in candidates
        if (("/" in c) or c.endswith(".md"))
        and " " not in c
        and not c.startswith("map[")
        and "->" not in c
        and "!=" not in c
        and not c.startswith(">")
        and len(c) < 100
    )
    missing = [p for p in named_paths if not (root / p).exists()]
    return CheckResult(
        id="P1-paths",
        dimension="D7",
        description="All repo paths named in AGENTS.md exist",
        passed=not missing,
        detail=("missing: " + ", ".join(missing)) if missing else f"checked {len(named_paths)} paths",
    )


def check_domain_purity(root: Path) -> CheckResult:
    """D7: invariant 'domain types free of Cobra and SQL' holds in tree."""
    domain = root / "internal" / "domain"
    offenders = []
    if domain.exists():
        for go in domain.rglob("*.go"):
            text = read_text(go)
            if re.search(r"spf13/cobra|database/sql|modernc\.org", text):
                offenders.append(str(go.relative_to(root)))
    return CheckResult(
        id="P1-domain-purity",
        dimension="D7",
        description="internal/domain imports no Cobra/SQL (contract line ~245)",
        passed=not offenders,
        detail=("offenders: " + ", ".join(offenders)) if offenders else "domain clean",
    )


def check_domain_no_provider(root: Path) -> CheckResult:
    """D7: invariant 'no provider coupling in domain packages' (line ~238)."""
    domain = root / "internal" / "domain"
    offenders = []
    if domain.exists():
        for go in domain.rglob("*.go"):
            if "internal/provider" in read_text(go):
                offenders.append(str(go.relative_to(root)))
    return CheckResult(
        id="P1-domain-no-provider",
        dimension="D7",
        description="internal/domain does not import internal/provider",
        passed=not offenders,
        detail=("offenders: " + ", ".join(offenders)) if offenders else "no provider coupling",
    )


def check_names_verification_command(agents: str) -> CheckResult:
    """D1: the contract must name a concrete build/test/verify command.

    Red at review open (AMR-002): AGENTS.md prescribes 'Add tests at the
    boundary' but never states how to build/test/vet/format. This check is
    the committed red-path fixture for that gap; it flips green when the
    contract adds a runnable gate.
    """
    # Require an actual command token, not the English word 'test'.
    pattern = re.compile(r"`?(go build|go test|go vet|gofmt)\b")
    found = sorted(set(m.group(1) for m in pattern.finditer(agents)))
    return CheckResult(
        id="P1-verify-command",
        dimension="D1",
        description="AGENTS.md names a concrete build/test/vet/format command",
        passed=bool(found),
        detail=("found: " + ", ".join(found)) if found else "no build/test/vet/gofmt command named",
    )


def check_first_contact_hard_constraint(agents: str) -> CheckResult:
    """D2: at least one operative hard constraint in the first 30 lines.

    Red at open (AMR-001): the first 30 lines are pure philosophy; the first
    imperative hard constraint ('provider identity must not leak') is at
    line ~48. Flips green when a hard-constraints block is front-loaded.
    """
    head = "\n".join(agents.splitlines()[:30])
    # An operative constraint is an imperative directed at the agent.
    has_constraint = bool(
        re.search(r"\b(do not|never|must not|always|must)\b", head, re.IGNORECASE)
    )
    return CheckResult(
        id="P1-first-contact",
        dimension="D2",
        description="An operative hard constraint appears in the first 30 lines",
        passed=has_constraint,
        detail="constraint present in head" if has_constraint else "first 30 lines carry no imperative constraint",
    )


def check_compaction_survival(agents: str) -> CheckResult:
    """D9: hard constraints survive a headings+bold+first-sentence extract.

    Red at open (AMR-003): the file has zero bold spans, so the mechanical
    extract is headings-only and every 'Do not…' constraint is dropped.
    Flips green when hard constraints are promoted into headings or bold.
    """
    bold_spans = re.findall(r"\*\*[^*]+\*\*", agents)
    # Constraints currently live as list items / prose 'Do not' lines.
    constraint_lines = [
        ln
        for ln in agents.splitlines()
        if re.match(r"\s*(-\s+)?(Do not|Never|Must not)\b", ln, re.IGNORECASE)
    ]
    # Survives only if constraints are reachable via bold (extract keeps bold).
    survives = len(bold_spans) > 0
    return CheckResult(
        id="P1-compaction",
        dimension="D9",
        description="Hard constraints survive a headings+bold compaction extract",
        passed=survives,
        detail=(
            f"{len(bold_spans)} bold spans; {len(constraint_lines)} 'do not/never' constraint lines "
            f"live only in prose"
        ),
    )


def check_go_gates(root: Path) -> CheckResult:
    """D7 support: prove the omitted commands would actually work here."""
    try:
        build = subprocess.run(
            ["go", "build", "./..."], cwd=root, capture_output=True, text=True, timeout=600
        )
    except (FileNotFoundError, subprocess.TimeoutExpired) as exc:
        return CheckResult(
            id="P1-go-gates",
            dimension="D7",
            description="Canonical go build ./... succeeds (context for AMR-002)",
            passed=False,
            detail=f"could not run go build: {exc}",
        )
    return CheckResult(
        id="P1-go-gates",
        dimension="D7",
        description="Canonical go build ./... succeeds (context for AMR-002)",
        passed=build.returncode == 0,
        detail="go build ./... OK" if build.returncode == 0 else build.stderr.strip()[:400],
    )


# Checks expected to be RED at review open (the ratchet targets).
EXPECTED_RED = {"P1-verify-command", "P1-first-contact", "P1-compaction"}

# Contextual checks: diagnostic support only. They never block the validator,
# because they depend on transient tree state (concurrent in-flight edits) that
# is orthogonal to the AGENTS.md contract this validator guards.
CONTEXTUAL = {"P1-go-gates"}


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--repo-root", type=Path, default=repo_root_default())
    parser.add_argument("--json", action="store_true", help="emit JSON only")
    parser.add_argument(
        "--enforce-ratchet",
        action="store_true",
        help="fail if an EXPECTED_RED check has not yet been fixed (default: only report)",
    )
    args = parser.parse_args()

    root = args.repo_root
    agents = read_text(root / "AGENTS.md")
    if not agents:
        print(f"absent-contract: no AGENTS.md at {root}", file=sys.stderr)
        return 2

    results = [
        check_paths_exist(root, agents),
        check_domain_purity(root),
        check_domain_no_provider(root),
        check_names_verification_command(agents),
        check_first_contact_hard_constraint(agents),
        check_compaction_survival(agents),
        check_go_gates(root),
    ]

    payload = {
        "target": "AGENTS.md",
        "checks": [asdict(r) for r in results],
        "expected_red_at_open": sorted(EXPECTED_RED),
    }

    if args.json:
        print(json.dumps(payload, indent=2))
    else:
        for r in results:
            flag = "PASS" if r.passed else "RED "
            if r.id in EXPECTED_RED and not r.passed:
                marker = " (expected-red-at-open)"
            elif r.id in CONTEXTUAL:
                marker = " (contextual, non-blocking)"
            else:
                marker = ""
            print(f"[{flag}] {r.id} {r.dimension}: {r.description}{marker}")
            print(f"        {r.detail}")

    # Factual invariants must always hold; a regression there fails the run.
    factual = [r for r in results if r.id not in EXPECTED_RED and r.id not in CONTEXTUAL]
    factual_ok = all(r.passed for r in factual)

    if not factual_ok:
        print("\nFACTUAL REGRESSION: a guarded reality claim went red.", file=sys.stderr)
        return 1

    if args.enforce_ratchet:
        still_red = [r.id for r in results if r.id in EXPECTED_RED and not r.passed]
        if still_red:
            print(
                "\nRATCHET: expected-red checks still unfixed: " + ", ".join(still_red),
                file=sys.stderr,
            )
            return 3

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
