#!/usr/bin/env python3
"""Structural validator for the l-extension-r1 candidate authoring package.

This is an ARTIFACT CHECKER, not a gate engine. It has no --attest, --approve,
or --force-admit flag and cannot grant research authority. A pass means the
package is internally consistent and machine-readable. It does NOT mean any
cited passage supports the claim attributed to it: semantic support is a
human review obligation, recorded in validation/REVIEW.md.

Standard library only. Usage: python3 validate.py [package_dir]
Exit 0 = all checks pass; exit 1 = at least one defect.
"""

import hashlib
import json
import os
import re
import sys

SCHEMA_VERSION = "normalize/v1"
VALID_SUPPORT_KINDS = {"explicit", "inferred", "unsupported"}
VALID_LOCALITY = {"local", "global", "mixed", "unknown"}
VALID_CONSTRUCTION = {"constructive", "existential", "mixed", "unknown"}
VALID_UNCERTAINTY = {"deterministic", "probabilistic", "mixed", "unknown"}
VALID_OUTCOME_CLASS = {
    "success",
    "partial_success",
    "failure",
    "partial_failure",
    "unknown",
}
REQUIRED_MECHANISM_KEYS = {
    "representations",
    "assumptions",
    "operators",
    "preserves",
    "breaks",
    "auxiliary_objects",
    "locality",
    "construction_mode",
    "uncertainty_mode",
}
FIXTURE_START = "<!-- newf-normalize"
FIXTURE_END = "newf-normalize -->"
SHA256_RE = re.compile(r"^[0-9a-f]{64}$")


class Defects:
    """Accumulates defects. Nothing here can be waived by a flag."""

    def __init__(self):
        self.items = []

    def add(self, where, message):
        self.items.append((where, message))

    def __len__(self):
        return len(self.items)


def sha256_file(path):
    digest = hashlib.sha256()
    with open(path, "rb") as handle:
        for chunk in iter(lambda: handle.read(65536), b""):
            digest.update(chunk)
    return digest.hexdigest()


def extract_payload(text):
    """Return (payload_dict, error). Mirrors the fixture provider's delimiters."""
    start = text.find(FIXTURE_START)
    if start < 0:
        return None, "no newf-normalize block found"
    end = text.find(FIXTURE_END, start)
    if end < 0:
        return None, "newf-normalize block is not closed"
    raw = text[start + len(FIXTURE_START) : end]
    try:
        return json.loads(raw), None
    except json.JSONDecodeError as exc:
        return None, "invalid fixture JSON: %s" % exc


def safe_relpath(root, candidate):
    """Reject absolute paths and any path escaping the package root."""
    if os.path.isabs(candidate):
        return None
    resolved = os.path.realpath(os.path.join(root, candidate))
    root_real = os.path.realpath(root)
    if resolved != root_real and not resolved.startswith(root_real + os.sep):
        return None
    return resolved


def check_note(path, text, seen_ids, defects):
    """Validate one Work record's prose + embedded fixture payload."""
    where = os.path.basename(path)
    payload, err = extract_payload(text)
    if err:
        defects.add(where, err)
        return
    if payload.get("schema_version") != SCHEMA_VERSION:
        defects.add(
            where,
            "schema_version must be %r, got %r"
            % (SCHEMA_VERSION, payload.get("schema_version")),
        )
    approaches = payload.get("approaches")
    if not isinstance(approaches, list) or not approaches:
        defects.add(where, "payload has no approaches")
        return

    # Paragraph count bounds the para:N locators the payload may cite.
    body = text[: text.find(FIXTURE_START)]
    paragraph_count = len([p for p in body.split("\n\n") if p.strip()])

    for index, approach in enumerate(approaches):
        tag = "%s approach[%d]" % (where, index)
        ident = approach.get("logical_identity")
        if not ident:
            defects.add(tag, "missing logical_identity")
        elif ident in seen_ids:
            defects.add(tag, "duplicate logical_identity %r (also in %s)" % (ident, seen_ids[ident]))
        else:
            seen_ids[ident] = where
        for field in ("label", "description"):
            if not approach.get(field):
                defects.add(tag, "missing %s" % field)

        mechanism = approach.get("mechanism")
        if not isinstance(mechanism, dict):
            defects.add(tag, "missing mechanism")
            continue
        missing = REQUIRED_MECHANISM_KEYS - set(mechanism)
        if missing:
            defects.add(tag, "mechanism missing keys: %s" % ", ".join(sorted(missing)))
        for key, allowed in (
            ("locality", VALID_LOCALITY),
            ("construction_mode", VALID_CONSTRUCTION),
            ("uncertainty_mode", VALID_UNCERTAINTY),
        ):
            value = mechanism.get(key)
            if value is not None and value not in allowed:
                defects.add(tag, "invalid %s %r" % (key, value))
        for key in ("representations", "assumptions", "operators", "preserves", "breaks", "auxiliary_objects"):
            value = mechanism.get(key)
            if key in mechanism and not isinstance(value, list):
                defects.add(tag, "mechanism.%s must be a list" % key)

        outcome = approach.get("outcome")
        if not isinstance(outcome, dict):
            defects.add(tag, "missing outcome")
        else:
            cls = outcome.get("class")
            if cls not in VALID_OUTCOME_CLASS:
                defects.add(tag, "invalid outcome.class %r" % cls)
            if not outcome.get("boundary_statement"):
                defects.add(tag, "missing outcome.boundary_statement")

        support = approach.get("support")
        if not isinstance(support, list) or not support:
            defects.add(tag, "missing support rows")
            continue
        supported_paths = set()
        for j, row in enumerate(support):
            rtag = "%s support[%d]" % (tag, j)
            kind = row.get("support_kind")
            if kind not in VALID_SUPPORT_KINDS:
                defects.add(rtag, "invalid support_kind %r" % kind)
            field_path = row.get("field_path")
            if not field_path:
                defects.add(rtag, "missing field_path")
            else:
                supported_paths.add(field_path)
            locator = row.get("locator") or ""
            if not locator:
                defects.add(rtag, "missing locator")
            elif locator.startswith("para:"):
                try:
                    num = int(locator.split(":", 1)[1])
                except ValueError:
                    defects.add(rtag, "malformed para locator %r" % locator)
                else:
                    if num < 1 or num > paragraph_count:
                        defects.add(
                            rtag,
                            "locator %r points outside the note's %d paragraphs"
                            % (locator, paragraph_count),
                        )
        # outcome.class is the load-bearing annotation; it must be supported.
        if "outcome.class" not in supported_paths:
            defects.add(tag, "outcome.class has no support row")


def check_manifest(root, manifest, defects):
    """Verify manifest hashes against actual file bytes.

    The manifest cannot hash its own final bytes; its digest is reported
    separately by the caller rather than stored inside itself.
    """
    if manifest.get("independent") is not None:
        defects.add("manifest.json", "manifest must not emit an `independent` field")
    if "execution_authorization" not in manifest:
        defects.add("manifest.json", "missing execution_authorization key")
    elif manifest["execution_authorization"] is not None:
        defects.add(
            "manifest.json",
            "execution_authorization must be null; only the operator may set it",
        )
    files = manifest.get("files")
    if not isinstance(files, list) or not files:
        defects.add("manifest.json", "missing files list")
        return
    for entry in files:
        rel = entry.get("path")
        if not rel:
            defects.add("manifest.json", "file entry missing path")
            continue
        resolved = safe_relpath(root, rel)
        if resolved is None:
            defects.add("manifest.json", "path %r is absolute or escapes the package" % rel)
            continue
        if not os.path.isfile(resolved):
            defects.add("manifest.json", "listed file is missing: %s" % rel)
            continue
        declared = entry.get("sha256")
        if not declared or not SHA256_RE.match(str(declared)):
            defects.add("manifest.json", "%s: sha256 must be a full 64-hex digest" % rel)
            continue
        actual = sha256_file(resolved)
        if actual != declared:
            defects.add(
                "manifest.json",
                "%s: hash mismatch (manifest %s..., actual %s...)"
                % (rel, declared[:12], actual[:12]),
            )


def check_sources(root, sources, notes_cited, defects):
    """Verify the source register and that every cited source id resolves."""
    entries = sources.get("sources")
    if not isinstance(entries, list) or not entries:
        defects.add("sources.json", "missing sources list")
        return set()
    ids = set()
    for entry in entries:
        sid = entry.get("id")
        if not sid:
            defects.add("sources.json", "source entry missing id")
            continue
        if sid in ids:
            defects.add("sources.json", "duplicate source id %r" % sid)
        ids.add(sid)
        if not entry.get("locator"):
            defects.add("sources.json", "%s: missing locator" % sid)
        if entry.get("verification_status") is None:
            defects.add("sources.json", "%s: missing verification_status" % sid)
        if entry.get("original_source_bytes_retained") is None:
            defects.add("sources.json", "%s: must state original_source_bytes_retained" % sid)
        digest = entry.get("artifact_sha256")
        if digest is not None and not SHA256_RE.match(str(digest)):
            defects.add("sources.json", "%s: artifact_sha256 must be a full digest" % sid)
    unknown = notes_cited - ids
    for sid in sorted(unknown):
        defects.add("notes", "cited source id %r is not in sources.json" % sid)
    return ids


SOURCE_ID_RE = re.compile(r"\bsrc-[a-z0-9][a-z0-9-]*\b")


def main(argv):
    args = argv[1:]
    # Deliberately no flags. This is an artifact checker; it has no authority to
    # grant, so it accepts nothing that could look like granting any.
    if len(args) > 1 or (args and args[0].startswith("-")):
        sys.stderr.write(
            "usage: validate.py [package_dir]\n"
            "This validator accepts no flags. It performs structural checks only\n"
            "and cannot attest, approve, or admit anything.\n"
        )
        raise SystemExit(2)
    root = os.path.abspath(args[0] if args else os.path.dirname(__file__) or ".")
    defects = Defects()

    for required in ("SCOPE.md", "README.md", "DEPENDENCIES.md", "manifest.json", "sources.json"):
        if not os.path.isfile(os.path.join(root, required)):
            defects.add(required, "required package file is missing")

    notes_dir = os.path.join(root, "notes")
    note_paths = []
    if not os.path.isdir(notes_dir):
        defects.add("notes/", "notes directory is missing")
    else:
        note_paths = sorted(
            os.path.join(notes_dir, name)
            for name in os.listdir(notes_dir)
            if name.endswith(".md")
        )

    seen_ids = {}
    notes_cited = set()
    for path in note_paths:
        with open(path, "r", encoding="utf-8") as handle:
            text = handle.read()
        check_note(path, text, seen_ids, defects)
        notes_cited.update(SOURCE_ID_RE.findall(text))

    sources_path = os.path.join(root, "sources.json")
    if os.path.isfile(sources_path):
        try:
            with open(sources_path, "r", encoding="utf-8") as handle:
                sources = json.load(handle)
        except json.JSONDecodeError as exc:
            defects.add("sources.json", "invalid JSON: %s" % exc)
        else:
            check_sources(root, sources, notes_cited, defects)

    manifest_path = os.path.join(root, "manifest.json")
    manifest_digest = None
    if os.path.isfile(manifest_path):
        try:
            with open(manifest_path, "r", encoding="utf-8") as handle:
                manifest = json.load(handle)
        except json.JSONDecodeError as exc:
            defects.add("manifest.json", "invalid JSON: %s" % exc)
        else:
            check_manifest(root, manifest, defects)
            manifest_digest = sha256_file(manifest_path)
            declared = manifest.get("record_count")
            if declared is not None and declared != len(note_paths):
                defects.add(
                    "manifest.json",
                    "record_count %r does not match %d note files"
                    % (declared, len(note_paths)),
                )

    print("package:        %s" % root)
    print("work records:   %d" % len(note_paths))
    print("logical ids:    %d unique" % len(seen_ids))
    print("cited sources:  %d" % len(notes_cited))
    if manifest_digest:
        # Reported separately: a manifest cannot contain the hash of its own bytes.
        print("manifest sha256: %s" % manifest_digest)

    if len(defects):
        print("\nDEFECTS (%d):" % len(defects))
        for where, message in defects.items:
            print("  %-52s %s" % (where, message))
        print("\nFAIL")
        return 1

    print("\nPASS — structural checks only.")
    print("This does NOT establish that any cited passage supports the claim")
    print("attributed to it, and is not a research validation or attestation.")
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv))
