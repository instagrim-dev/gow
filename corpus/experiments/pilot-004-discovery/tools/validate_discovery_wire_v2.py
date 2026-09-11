#!/usr/bin/env python3
"""discovery-wire/v1 validator, checker version 2 (note-scoped).

Usage: validate_discovery_wire_v2.py <capture.json> <bundle.md>

Supersedes validate_discovery_wire.py's checks for NEW validation runs while
the frozen v1 file and its recorded results stand untouched (the pilot-004
FREEZE pins v1's digest; this file is a separately identified checker).

v2 corrections over v1 (post-freeze review, 2026-09-11):
  1. Quotations are verified against the NAMED note's own section of the
     bundle, not the bundle as a whole — v1 accepted text from note A
     attributed to note B.
  2. `shared_by` requires >= 2 DISTINCT valid filenames — v1 counted list
     entries, so [A, A] passed the two-approaches condition.
All other checks are identical to v1.
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


def split_notes(bundle: str) -> dict:
    """Map note filename -> that note's own text (its bundle section)."""
    parts = re.split(r"<!-- train-source: (\S+) -->", bundle)
    # parts = [preamble, name1, text1, name2, text2, ...]
    sections = {}
    for i in range(1, len(parts) - 1, 2):
        sections[parts[i]] = parts[i + 1]
    return sections


def main() -> None:
    if len(sys.argv) != 3:
        fail("usage: validate_discovery_wire_v2.py <capture.json> <bundle.md>")
    raw = open(sys.argv[1], encoding="utf-8").read()
    bundle = open(sys.argv[2], encoding="utf-8").read()

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

    sections = split_notes(bundle)
    if len(sections) != 12:
        fail(f"bundle carries {len(sections)} notes, want 12")

    for i, p in enumerate(props):
        if not isinstance(p, dict) or set(p) != PROPERTY_FIELDS:
            fail(f"properties[{i}]: fields must be exactly {sorted(PROPERTY_FIELDS)}")
        for f in PROSE_FIELDS:
            if not isinstance(p[f], str) or not p[f].strip():
                fail(f"properties[{i}].{f} must be nonempty prose")
        shared = p["shared_by"]
        if not isinstance(shared, list):
            fail(f"properties[{i}].shared_by must be a list")
        distinct = set()
        for n in shared:
            if n not in sections:
                fail(f"properties[{i}].shared_by names unknown note {n!r}")
            distinct.add(n)
        if len(distinct) < 2:
            fail(f"properties[{i}].shared_by needs >=2 DISTINCT notes (got {sorted(distinct)})")
        passages = p["supporting_passages"]
        if not isinstance(passages, list) or not passages:
            fail(f"properties[{i}].supporting_passages must be a nonempty list")
        for j, sp in enumerate(passages):
            if not isinstance(sp, dict) or set(sp) != {"note", "quote"}:
                fail(f"properties[{i}].supporting_passages[{j}] must be exactly {{note, quote}}")
            if sp["note"] not in sections:
                fail(f"properties[{i}].supporting_passages[{j}] names unknown note {sp['note']!r}")
            if not isinstance(sp["quote"], str) or not sp["quote"].strip():
                fail(f"properties[{i}].supporting_passages[{j}].quote must be nonempty")
            if sp["quote"] not in sections[sp["note"]]:
                where = [name for name, text in sections.items() if sp["quote"] in text]
                fail(f"properties[{i}].supporting_passages[{j}]: quote is NOT in the named note {sp['note']!r}"
                     + (f" (it occurs in {where})" if where else " (it occurs nowhere in the bundle)"))

    print(f"valid: {len(props)} property proposal(s) [checker v2, note-scoped]")


if __name__ == "__main__":
    main()
