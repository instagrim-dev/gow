#!/usr/bin/env python3
"""discovery-wire/v1 validator — pilot-004's importer-path decode.

Usage: validate_discovery_wire.py <capture.json> <bundle.md>

Exit 0 iff the capture satisfies the full wire contract against the exact
bundle the proposer received. This is the decode path a capture must pass
BEFORE being declared sealable (pilot-003 correction: sealing checks weaker
than the decode contract are not validation).
"""
import json
import re
import sys

CAP = 5
PROSE_FIELDS = ("statement", "preserved_distinctions", "contrast_check", "falsification")
PROPERTY_FIELDS = set(PROSE_FIELDS) | {"shared_by", "supporting_passages"}


def fail(msg: str) -> None:
    print(f"INVALID: {msg}")
    sys.exit(1)


def main() -> None:
    if len(sys.argv) != 3:
        fail("usage: validate_discovery_wire.py <capture.json> <bundle.md>")
    raw = open(sys.argv[1], encoding="utf-8").read()
    bundle = open(sys.argv[2], encoding="utf-8").read()

    # Whole-payload strictness.
    try:
        decoder = json.JSONDecoder()
        doc, end = decoder.raw_decode(raw)
    except json.JSONDecodeError as e:
        fail(f"json: {e}")
    if raw[end:].strip():
        fail("trailing content after the wire document")

    if not isinstance(doc, dict) or set(doc) != {"schema_version", "properties"}:
        fail("top-level must be exactly {schema_version, properties}")
    if doc["schema_version"] != "discovery-wire/v1":
        fail(f"schema_version {doc['schema_version']!r}")
    props = doc["properties"]
    if not isinstance(props, list) or len(props) > CAP:
        fail(f"properties must be a list of 0..{CAP}")

    note_names = set(re.findall(r"<!-- train-source: (\S+) -->", bundle))
    if len(note_names) != 12:
        fail(f"bundle carries {len(note_names)} notes, want 12")

    for i, p in enumerate(props):
        if not isinstance(p, dict) or set(p) != PROPERTY_FIELDS:
            fail(f"properties[{i}]: fields must be exactly {sorted(PROPERTY_FIELDS)}")
        for f in PROSE_FIELDS:
            if not isinstance(p[f], str) or not p[f].strip():
                fail(f"properties[{i}].{f} must be nonempty prose")
        shared = p["shared_by"]
        if not isinstance(shared, list) or len(shared) < 2:
            fail(f"properties[{i}].shared_by needs >=2 notes (a property one approach has is not shared)")
        for n in shared:
            if n not in note_names:
                fail(f"properties[{i}].shared_by names unknown note {n!r}")
        passages = p["supporting_passages"]
        if not isinstance(passages, list) or not passages:
            fail(f"properties[{i}].supporting_passages must be a nonempty list")
        for j, sp in enumerate(passages):
            if not isinstance(sp, dict) or set(sp) != {"note", "quote"}:
                fail(f"properties[{i}].supporting_passages[{j}] must be exactly {{note, quote}}")
            if sp["note"] not in note_names:
                fail(f"properties[{i}].supporting_passages[{j}] names unknown note {sp['note']!r}")
            if not isinstance(sp["quote"], str) or not sp["quote"].strip():
                fail(f"properties[{i}].supporting_passages[{j}].quote must be nonempty")
            if sp["quote"] not in bundle:
                fail(f"properties[{i}].supporting_passages[{j}].quote is not a verbatim substring of the bundle")

    print(f"valid: {len(props)} property proposal(s)")


if __name__ == "__main__":
    main()
