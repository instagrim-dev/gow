#!/usr/bin/env python3
"""Offline integrity/annotation lint; not a theorem checker or Go-schema validator."""
from __future__ import annotations

import hashlib
import json
from pathlib import Path
import re
import sys
from urllib.parse import urlparse

BLOCK = re.compile(r"<!-- newf-normalize\s*(.*?)\s*newf-normalize -->", re.S)


def validate(root: Path) -> tuple[int, int]:
    manifest = json.loads((root / "sources.json").read_text(encoding="utf-8"))
    if manifest.get("schema") != "newf-research-source-register/v1":
        raise ValueError("unsupported source-register schema")
    for key in ("benchmark_training_eligible", "historical_holdout_eligible",
                "original_source_hashes_available"):
        if manifest.get(key) is not False:
            raise ValueError(f"{key} must remain false")
    entries = manifest["entries"]
    if manifest["entry_count"] != len(entries):
        raise ValueError("entry count mismatch")
    ids, paths, annotated = set(), set(), 0
    for entry in entries:
        eid, relative = entry["id"], Path(entry["path"])
        if eid in ids or relative.as_posix() in paths:
            raise ValueError(f"duplicate id or path: {eid}")
        ids.add(eid)
        paths.add(relative.as_posix())
        if relative.is_absolute() or ".." in relative.parts:
            raise ValueError(f"unsafe path: {relative}")
        path = (root / relative).resolve()
        if root.resolve() not in path.parents:
            raise ValueError(f"path escapes collection: {relative}")
        bucket = entry["collection"]
        if bucket not in {"notes", "exploratory", "quarantine"} or relative.parts[0] != bucket:
            raise ValueError(f"invalid bucket: {eid}")
        raw = path.read_bytes()
        if hashlib.sha256(raw).hexdigest() != entry["note_sha256"]:
            raise ValueError(f"note hash mismatch: {eid}")
        text = raw.decode("utf-8")
        if not text.startswith(f"# {eid} ") or not entry.get("read_scope"):
            raise ValueError(f"missing identity or read scope: {eid}")
        if entry["independent_verification"] != "not_performed" or entry["original_source_bytes_archived"] is not False:
            raise ValueError(f"unsupported verification/archive claim: {eid}")
        for url in entry["primary_urls"]:
            parsed = urlparse(url)
            if parsed.scheme != "https" or not parsed.netloc or url not in text:
                raise ValueError(f"invalid or missing source URL: {eid}")
        blocks = BLOCK.findall(text)
        if bucket == "quarantine":
            if blocks or entry["annotation_status"] != "quarantined_unverified_claim":
                raise ValueError(f"quarantine cannot contain normalization: {eid}")
            continue
        if len(blocks) != 1 or entry["annotation_status"] != "curator_inferred":
            raise ValueError(f"expected one curator annotation: {eid}")
        doc = json.loads(blocks[0])
        if doc["schema_version"] != "normalize/v1" or len(doc["approaches"]) != 1:
            raise ValueError(f"invalid normalization envelope: {eid}")
        approach = doc["approaches"][0]
        if approach["logical_identity"] != f"erdos-straus/research/{eid}":
            raise ValueError(f"wrong logical identity: {eid}")
        mechanism, outcome = approach["mechanism"], approach["outcome"]
        for key, allowed in {
            "locality": {"local", "global", "mixed", "unknown"},
            "construction_mode": {"constructive", "existential", "mixed", "unknown"},
            "uncertainty_mode": {"deterministic", "probabilistic", "mixed", "unknown"},
        }.items():
            if mechanism[key] not in allowed:
                raise ValueError(f"invalid {key}: {eid}")
        if outcome["class"] not in {"partial_success", "unknown"} or outcome["class"] != entry["outcome_interpretation"]:
            raise ValueError(f"unsupported outcome: {eid}")
        if bucket == "exploratory" and outcome["class"] != "unknown":
            raise ValueError(f"exploratory outcome must remain unknown: {eid}")
        support = approach["support"]
        support_paths = [s["field_path"] for s in support]
        if len(support_paths) != len(set(support_paths)):
            raise ValueError(f"duplicate field support: {eid}")
        expected = {f"mechanism.{k}" for k, v in mechanism.items() if v and k != "notes"}
        expected |= {"outcome.class", "outcome.boundary_statement", "outcome.boundary_conditions"}
        if set(support_paths) != expected:
            raise ValueError(f"missing/unexpected field provenance: {eid}")
        for item in support:
            if item["support_kind"] != "inferred":
                raise ValueError(f"curation promoted to source evidence: {eid}")
            if not item["locator"].startswith("heading:") or "## " + item["locator"][8:] not in text:
                raise ValueError(f"invalid locator: {eid}")
        annotated += 1
    actual = {p.relative_to(root).as_posix() for bucket in ("notes", "exploratory", "quarantine")
              for p in (root / bucket).glob("*.md")}
    if actual != paths:
        raise ValueError("registry does not cover exactly the source cards")
    return len(entries), annotated


if __name__ == "__main__":
    try:
        root = Path(sys.argv[1]) if len(sys.argv) > 1 else Path(__file__).resolve().parent
        count, annotated = validate(root)
        print(f"PASS: {count} source cards; {annotated} inferred annotations; integrity and admission lint only")
    except (OSError, ValueError, KeyError, TypeError, IndexError) as exc:
        print(f"FAIL: {exc}", file=sys.stderr)
        raise SystemExit(1)
