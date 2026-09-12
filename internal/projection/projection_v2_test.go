package projection

import (
	"strings"
	"testing"
)

// v2Base is a complete, valid projection/v2 artifact (issue #22, D3): the v1
// composition shape plus the full semantic-preservation contract.
const v2Base = `{
  "schema": "projection/v2",
  "givens": ["congruence-lattice"],
  "target": ["crossing-bound"],
  "steps": [
    {"name": "lift", "statement": "lift to the covering system",
     "requires": ["congruence-lattice"], "provides": ["cover"]},
    {"name": "bound", "statement": "bound crossings over the cover",
     "requires": ["cover"], "provides": ["crossing-bound"]}
  ],
  "source_domain": "unit-fraction identities over Z",
  "target_domain": "congruence covers of the moduli",
  "preserves": ["solution existence per residue class"],
  "loses": ["constructive witness values"],
  "correspondence": "one-way-implication",
  "grounding_plan": "each residue-class cover bound maps back to a family of concrete n with a checkable witness obligation"
}`

// TestParseV2RoundTrip: acceptance clause 1 — a complete v2 fixture parses,
// carries every semantic field typed, and composition checking is unchanged.
func TestParseV2RoundTrip(t *testing.T) {
	a, err := Parse([]byte(v2Base))
	if err != nil {
		t.Fatalf("parse v2: %v", err)
	}
	if a.Schema != SchemaProjectionV2 {
		t.Fatalf("schema: %q", a.Schema)
	}
	if a.SourceDomain == "" || a.TargetDomain == "" || len(a.Preserves) != 1 ||
		len(a.Loses) != 1 || a.Correspondence != CorrespondenceOneWayImplication || a.GroundingPlan == "" {
		t.Fatalf("v2 semantic fields must round-trip typed: %+v", a)
	}
	if res := CheckComposition(a); !res.OK {
		t.Fatalf("composition must be v1-identical: %+v", res)
	}
}

// TestParseV2ExplicitEmptyLosesAllowed: an explicit empty loses list is the
// author's claim that nothing known is lost — present, therefore valid.
func TestParseV2ExplicitEmptyLosesAllowed(t *testing.T) {
	raw := strings.Replace(v2Base, `"loses": ["constructive witness values"]`, `"loses": []`, 1)
	if _, err := Parse([]byte(raw)); err != nil {
		t.Fatalf("explicit empty loses must parse: %v", err)
	}
}

// TestParseV2MissingFieldRefused: acceptance clause 2 — a v2 artifact missing
// ANY semantic field is refused, field by field.
func TestParseV2MissingFieldRefused(t *testing.T) {
	cases := []struct {
		name   string
		remove string
		want   string
	}{
		{"source_domain", `"source_domain": "unit-fraction identities over Z",`, "source_domain is required"},
		{"target_domain", `"target_domain": "congruence covers of the moduli",`, "target_domain is required"},
		{"preserves", `"preserves": ["solution existence per residue class"],`, "preserves is required"},
		{"loses", `"loses": ["constructive witness values"],`, "loses is required"},
		{"correspondence", `"correspondence": "one-way-implication",`, "correspondence must be one of"},
		{"grounding_plan", `,
  "grounding_plan": "each residue-class cover bound maps back to a family of concrete n with a checkable witness obligation"`, "grounding_plan is required"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw := strings.Replace(v2Base, tc.remove, "", 1)
			if raw == v2Base {
				t.Fatalf("fixture edit did not apply for %s", tc.name)
			}
			_, err := Parse([]byte(raw))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("missing %s must be refused with %q, got %v", tc.name, tc.want, err)
			}
		})
	}
}

// TestParseV2EmptyPreservesRefused: an abstraction that cannot state what it
// preserves is defective (docs/abstraction-safety.md).
func TestParseV2EmptyPreservesRefused(t *testing.T) {
	raw := strings.Replace(v2Base, `"preserves": ["solution existence per residue class"]`, `"preserves": []`, 1)
	if _, err := Parse([]byte(raw)); err == nil || !strings.Contains(err.Error(), "preserves is required and non-empty") {
		t.Fatalf("empty preserves must be refused, got %v", err)
	}
}

// TestParseV2InvalidCorrespondenceRefused: the class vocabulary is closed and
// has no default.
func TestParseV2InvalidCorrespondenceRefused(t *testing.T) {
	raw := strings.Replace(v2Base, `"correspondence": "one-way-implication"`, `"correspondence": "strong-analogy"`, 1)
	if _, err := Parse([]byte(raw)); err == nil || !strings.Contains(err.Error(), "correspondence must be one of") {
		t.Fatalf("out-of-vocabulary correspondence must be refused, got %v", err)
	}
}

// TestParseV1BackCompat: acceptance clause 3 — a v1 artifact still parses and
// carries no semantic fields (correspondence unrecorded, never defaulted).
func TestParseV1BackCompat(t *testing.T) {
	raw := `{
  "schema": "projection/v1",
  "steps": [{"name": "s", "statement": "do the thing", "provides": ["done"]}],
  "target": ["done"]
}`
	a, err := Parse([]byte(raw))
	if err != nil {
		t.Fatalf("v1 must keep parsing: %v", err)
	}
	if a.Correspondence != "" || a.SourceDomain != "" || a.Preserves != nil {
		t.Fatalf("v1 must carry no semantic fields: %+v", a)
	}
}

// TestParseV1WithV2FieldsRefused: strict parse — semantic claims may not ride
// a v1 schema declaration without the v2 obligations they owe.
func TestParseV1WithV2FieldsRefused(t *testing.T) {
	raw := `{
  "schema": "projection/v1",
  "steps": [{"name": "s", "statement": "do the thing", "provides": ["done"]}],
  "target": ["done"],
  "correspondence": "equivalence"
}`
	if _, err := Parse([]byte(raw)); err == nil || !strings.Contains(err.Error(), "carries projection/v2 semantic fields") {
		t.Fatalf("v2 fields on v1 must be refused, got %v", err)
	}
}
