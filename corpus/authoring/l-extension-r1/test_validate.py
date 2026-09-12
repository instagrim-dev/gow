#!/usr/bin/env python3
"""Regression tests for validate.py.

Covers the boundary cases the implementation plan requires (§7 C4), including
the ones that must NEVER pass: stale manifest hashes, missing evidence
locators, escaping paths, and any research-authority escape hatch.

Run: python3 -m unittest discover -s corpus/authoring/l-extension-r1 -p 'test_*.py'
"""

import ast
import hashlib
import json
import os
import shutil
import tempfile
import unittest

import validate

MINIMAL_PAYLOAD = {
    "schema_version": "normalize/v1",
    "approaches": [
        {
            "logical_identity": "demo/one",
            "label": "Demo",
            "description": "Demo approach.",
            "mechanism": {
                "representations": ["r"],
                "assumptions": ["a"],
                "operators": ["o"],
                "preserves": ["p"],
                "breaks": [],
                "auxiliary_objects": [],
                "locality": "local",
                "construction_mode": "constructive",
                "uncertainty_mode": "deterministic",
                "notes": "n",
            },
            "outcome": {
                "class": "failure",
                "boundary_statement": "bounded",
                "boundary_conditions": [],
                "notes": "n",
            },
            "support": [
                {"field_path": "outcome.class", "support_kind": "explicit", "locator": "para:1"}
            ],
        }
    ],
}


def note_text(payload, body="Paragraph one.\n\nParagraph two.\n"):
    return "# Demo (demo-01)\n\n%s\n%s\n%s\n%s\n" % (
        body,
        validate.FIXTURE_START,
        json.dumps(payload, indent=2),
        validate.FIXTURE_END,
    )


class PackageFixture:
    """Builds a throwaway minimal package on disk."""

    def __init__(self, root):
        self.root = root
        os.makedirs(os.path.join(root, "notes"))
        for name in ("SCOPE.md", "README.md", "DEPENDENCIES.md"):
            self.write(name, "# %s\n" % name)
        self.write_note("demo-01.md", MINIMAL_PAYLOAD)
        self.write_json(
            "sources.json",
            {
                "sources": [
                    {
                        "id": "src-demo",
                        "locator": "https://example.org/demo",
                        "verification_status": "verified",
                        "original_source_bytes_retained": False,
                    }
                ]
            },
        )
        self.rewrite_manifest()

    def path(self, rel):
        return os.path.join(self.root, rel)

    def write(self, rel, text):
        with open(self.path(rel), "w", encoding="utf-8") as handle:
            handle.write(text)

    def write_json(self, rel, obj):
        self.write(rel, json.dumps(obj, indent=2) + "\n")

    def write_note(self, name, payload, body="Paragraph one.\n\nParagraph two.\n"):
        self.write(os.path.join("notes", name), note_text(payload, body))

    def load(self, rel):
        with open(self.path(rel), "r", encoding="utf-8") as handle:
            return json.load(handle)

    def rewrite_manifest(self):
        notes = sorted(os.listdir(self.path("notes")))
        files = []
        for rel in ["SCOPE.md", "README.md", "DEPENDENCIES.md", "sources.json"] + [
            os.path.join("notes", n) for n in notes
        ]:
            files.append({"path": rel, "sha256": validate.sha256_file(self.path(rel))})
        self.write_json(
            "manifest.json",
            {
                "artifact_status": "candidate_authoring_package",
                "record_count": len(notes),
                "files": files,
                "execution_authorization": None,
            },
        )


class ValidatorTest(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.mkdtemp()
        self.pkg = PackageFixture(self.tmp)

    def tearDown(self):
        shutil.rmtree(self.tmp, ignore_errors=True)

    def run_validator(self):
        return validate.main(["validate.py", self.tmp])

    def assertFails(self):
        self.assertEqual(1, self.run_validator())

    # --- baseline -------------------------------------------------------

    def test_minimal_valid_package_passes(self):
        self.assertEqual(0, self.run_validator())

    def test_no_research_authority_escape_hatch(self):
        """No authority flag may exist in code.

        Scanned over non-docstring code strings only: the module docstring names
        these flags in order to disclaim them, so a raw substring search over the
        whole file would match its own disclaimer.
        """
        path = os.path.join(
            os.path.dirname(os.path.abspath(validate.__file__)), "validate.py"
        )
        with open(path, encoding="utf-8") as handle:
            tree = ast.parse(handle.read())

        # Collect docstring nodes by identity so cleaned-text mismatches cannot
        # cause a docstring to be scanned as if it were code.
        docstring_nodes = set()
        for node in ast.walk(tree):
            if isinstance(node, (ast.Module, ast.FunctionDef, ast.ClassDef)):
                body = getattr(node, "body", None)
                if (
                    body
                    and isinstance(body[0], ast.Expr)
                    and isinstance(body[0].value, ast.Constant)
                    and isinstance(body[0].value.value, str)
                ):
                    docstring_nodes.add(id(body[0].value))

        code_strings = [
            node.value
            for node in ast.walk(tree)
            if isinstance(node, ast.Constant)
            and isinstance(node.value, str)
            and id(node) not in docstring_nodes
        ]
        for escape in ("--attest", "--approve", "--force-admit", "--admit", "--promote"):
            for value in code_strings:
                self.assertNotIn(
                    escape,
                    value,
                    "validator must not expose %r (found in %r)" % (escape, value),
                )

    def test_validator_defines_no_flag_parser(self):
        """No argument parser means no room for an authority flag to appear."""
        path = os.path.join(
            os.path.dirname(os.path.abspath(validate.__file__)), "validate.py"
        )
        with open(path, encoding="utf-8") as handle:
            tree = ast.parse(handle.read())
        imported = set()
        for node in ast.walk(tree):
            if isinstance(node, ast.Import):
                imported.update(alias.name for alias in node.names)
            elif isinstance(node, ast.ImportFrom) and node.module:
                imported.add(node.module)
        self.assertNotIn("argparse", imported)
        self.assertNotIn("optparse", imported)

    def test_validator_takes_no_authority_flags(self):
        """The only accepted argument is a package directory."""
        self.assertEqual(0, validate.main(["validate.py", self.tmp]))
        with self.assertRaises(SystemExit) as raised:
            validate.main(["validate.py", self.tmp, "--attest"])
        self.assertNotEqual(0, raised.exception.code)

    # --- mandated boundary cases (plan §7 C4) ---------------------------

    def test_changed_source_bytes_under_stale_manifest_fails(self):
        """Integrity failure: content edited without a manifest revision."""
        with open(self.pkg.path("SCOPE.md"), "a", encoding="utf-8") as handle:
            handle.write("an edit the manifest does not know about\n")
        self.assertFails()

    def test_truncated_display_prefix_hash_rejected(self):
        """The manifest records full SHA-256, not display prefixes."""
        manifest = self.pkg.load("manifest.json")
        manifest["files"][0]["sha256"] = manifest["files"][0]["sha256"][:12]
        self.pkg.write_json("manifest.json", manifest)
        self.assertFails()

    def test_missing_listed_file_fails(self):
        os.remove(self.pkg.path("notes/demo-01.md"))
        self.assertFails()

    def test_escaping_path_rejected(self):
        manifest = self.pkg.load("manifest.json")
        manifest["files"].append({"path": "../../etc/passwd", "sha256": "0" * 64})
        self.pkg.write_json("manifest.json", manifest)
        self.assertFails()

    def test_absolute_path_rejected(self):
        manifest = self.pkg.load("manifest.json")
        manifest["files"].append({"path": "/etc/passwd", "sha256": "0" * 64})
        self.pkg.write_json("manifest.json", manifest)
        self.assertFails()

    def test_duplicate_logical_id_fails(self):
        self.pkg.write_note("demo-02.md", MINIMAL_PAYLOAD)
        self.pkg.rewrite_manifest()
        self.assertFails()

    def test_missing_evidence_locator_fails(self):
        """No fabricated support: a support row without a locator is a defect."""
        payload = json.loads(json.dumps(MINIMAL_PAYLOAD))
        del payload["approaches"][0]["support"][0]["locator"]
        self.pkg.write_note("demo-01.md", payload)
        self.pkg.rewrite_manifest()
        self.assertFails()

    def test_unsupported_outcome_class_fails(self):
        """outcome.class is load-bearing; it may not go unsupported."""
        payload = json.loads(json.dumps(MINIMAL_PAYLOAD))
        payload["approaches"][0]["support"] = [
            {"field_path": "mechanism.operators", "support_kind": "explicit", "locator": "para:1"}
        ]
        self.pkg.write_note("demo-01.md", payload)
        self.pkg.rewrite_manifest()
        self.assertFails()

    def test_locator_beyond_paragraph_count_fails(self):
        """A locator must point at text that exists."""
        payload = json.loads(json.dumps(MINIMAL_PAYLOAD))
        payload["approaches"][0]["support"][0]["locator"] = "para:99"
        self.pkg.write_note("demo-01.md", payload)
        self.pkg.rewrite_manifest()
        self.assertFails()

    def test_invalid_fixture_json_fails(self):
        self.pkg.write(
            "notes/demo-01.md",
            "# Demo\n\nBody.\n\n%s\n{not json,}\n%s\n"
            % (validate.FIXTURE_START, validate.FIXTURE_END),
        )
        self.pkg.rewrite_manifest()
        self.assertFails()

    def test_unclosed_fixture_block_fails(self):
        self.pkg.write("notes/demo-01.md", "# Demo\n\nBody.\n\n%s\n{}\n" % validate.FIXTURE_START)
        self.pkg.rewrite_manifest()
        self.assertFails()

    def test_invalid_support_kind_fails(self):
        payload = json.loads(json.dumps(MINIMAL_PAYLOAD))
        payload["approaches"][0]["support"][0]["support_kind"] = "guessed"
        self.pkg.write_note("demo-01.md", payload)
        self.pkg.rewrite_manifest()
        self.assertFails()

    def test_invalid_posture_enum_fails(self):
        payload = json.loads(json.dumps(MINIMAL_PAYLOAD))
        payload["approaches"][0]["mechanism"]["locality"] = "somewhat-local"
        self.pkg.write_note("demo-01.md", payload)
        self.pkg.rewrite_manifest()
        self.assertFails()

    def test_unresolved_vocabulary_is_reported_not_aliased(self):
        """An unrecognized enum value is a defect, never silently mapped."""
        payload = json.loads(json.dumps(MINIMAL_PAYLOAD))
        payload["approaches"][0]["mechanism"]["construction_mode"] = "semi-constructive"
        self.pkg.write_note("demo-01.md", payload)
        self.pkg.rewrite_manifest()
        self.assertFails()

    def test_wrong_schema_version_fails(self):
        payload = json.loads(json.dumps(MINIMAL_PAYLOAD))
        payload["schema_version"] = "normalize/v2"
        self.pkg.write_note("demo-01.md", payload)
        self.pkg.rewrite_manifest()
        self.assertFails()

    def test_cited_source_absent_from_register_fails(self):
        """A note may not cite a source id the register does not resolve."""
        self.pkg.write_note(
            "demo-01.md",
            MINIMAL_PAYLOAD,
            body="Paragraph one cites src-not-registered.\n\nParagraph two.\n",
        )
        self.pkg.rewrite_manifest()
        self.assertFails()

    def test_source_without_verification_status_fails(self):
        sources = self.pkg.load("sources.json")
        del sources["sources"][0]["verification_status"]
        self.pkg.write_json("sources.json", sources)
        self.pkg.rewrite_manifest()
        self.assertFails()

    def test_source_must_state_whether_bytes_retained(self):
        sources = self.pkg.load("sources.json")
        del sources["sources"][0]["original_source_bytes_retained"]
        self.pkg.write_json("sources.json", sources)
        self.pkg.rewrite_manifest()
        self.assertFails()

    def test_record_count_mismatch_fails(self):
        manifest = self.pkg.load("manifest.json")
        manifest["record_count"] = 11
        self.pkg.write_json("manifest.json", manifest)
        self.assertFails()

    # --- authority boundaries -------------------------------------------

    def test_independent_true_rejected(self):
        """The package may never assert an independence verdict."""
        manifest = self.pkg.load("manifest.json")
        manifest["independent"] = True
        self.pkg.write_json("manifest.json", manifest)
        self.assertFails()

    def test_self_granted_execution_authorization_rejected(self):
        manifest = self.pkg.load("manifest.json")
        manifest["execution_authorization"] = "authorized"
        self.pkg.write_json("manifest.json", manifest)
        self.assertFails()

    def test_missing_execution_authorization_key_fails(self):
        manifest = self.pkg.load("manifest.json")
        del manifest["execution_authorization"]
        self.pkg.write_json("manifest.json", manifest)
        self.assertFails()


if __name__ == "__main__":
    unittest.main()
