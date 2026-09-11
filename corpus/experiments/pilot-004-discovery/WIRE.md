# Pilot 004 — discovery wire (`discovery-wire/v1`)

The capture output contract for both arms. The committed validator
(`tools/validate_discovery_wire.py`) IS this pilot's importer-path decode:
per the pilot-003 correction, a capture is declared sealable only after the
validator passes on the exact raw bytes — sealing checks weaker than the
decode contract are not validation.

One JSON document, no Markdown fence:

```json
{
  "schema_version": "discovery-wire/v1",
  "properties": [
    {
      "statement": "one-sentence property the named approaches share",
      "shared_by": ["es-01-....md", "es-02-....md"],
      "supporting_passages": [
        {"note": "es-01-....md", "quote": "verbatim passage"}
      ],
      "preserved_distinctions": "what this property deliberately does NOT merge",
      "contrast_check": "result of checking the property against the partial-success notes",
      "falsification": "what observation would refute the property"
    }
  ]
}
```

Rules (enforced by the validator):

- exactly the top-level fields shown; unknown fields anywhere are violations;
- 0–5 properties (cap 5, decision D1). An empty list is permitted and
  honest: a proposer that finds nothing shared must not invent;
- every `shared_by` and `supporting_passages[].note` entry must be one of
  the 12 bundle filenames; `shared_by` needs ≥2 entries (a property one
  approach has is not shared);
- every `supporting_passages[].quote` must be a verbatim substring of the
  named note's bundle text (checked against the arm's own bundle);
- all prose fields non-empty; whole-payload strictness (no trailing
  content).
