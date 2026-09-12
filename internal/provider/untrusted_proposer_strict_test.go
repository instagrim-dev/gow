package provider

import (
	"errors"
	"strings"
	"testing"
)

// F2/F3 regressions: encoding/json is last-wins on duplicate keys and
// case-insensitive on struct tags, so DisallowUnknownFields alone does not
// close the wire vocabulary. The token-level strict pre-pass must reject
// duplicates and case-variant/unknown keys byte-exactly, at every nesting
// level, while leaving valid payloads untouched.
func TestParseWireProposalsStrictKeyDiscipline(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		raw  string
	}{
		{
			name: "duplicate top-level schema_version",
			raw: strings.Replace(validWire,
				`"schema_version": "proposal-wire/v1",`,
				`"schema_version": "proposal-wire/v0", "schema_version": "proposal-wire/v1",`, 1),
		},
		{
			name: "duplicate structural_violation_claim in a proposal",
			raw: strings.Replace(validWire,
				`"structural_violation_claim": "drops the shared property",`,
				`"structural_violation_claim": "first claim", "structural_violation_claim": "drops the shared property",`, 1),
		},
		{
			name: "duplicate locality inside the mechanism object",
			raw: strings.Replace(validWire,
				`"locality": "global",`,
				`"locality": "local", "locality": "global",`, 1),
		},
		{
			name: "case-variant top-level SCHEMA_VERSION",
			raw:  strings.Replace(validWire, `"schema_version"`, `"SCHEMA_VERSION"`, 1),
		},
		{
			name: "case-variant nested Mechanism",
			raw:  strings.Replace(validWire, `"mechanism"`, `"Mechanism"`, 1),
		},
		{
			name: "case-variant nested Preserves",
			raw:  strings.Replace(validWire, `"preserves"`, `"Preserves"`, 1),
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if tc.raw == validWire {
				t.Fatal("test bug: strings.Replace changed nothing")
			}
			_, err := ParseWireProposals(tc.raw, []string{"inv_x"})
			if err == nil {
				t.Fatal("payload must be rejected")
			}
			if !errors.Is(err, ErrProposalWireViolation) {
				t.Fatalf("want ErrProposalWireViolation, got %v", err)
			}
		})
	}
}

// The pre-pass must not change what a VALID payload means: the minimal
// fixture that passed before the strict key pass still parses.
func TestParseWireProposalsStrictPassAcceptsValidPayload(t *testing.T) {
	t.Parallel()
	proposals, err := ParseWireProposals(validWire, []string{"inv_x"})
	if err != nil {
		t.Fatalf("valid payload must still parse: %v", err)
	}
	if len(proposals) != 1 {
		t.Fatalf("want 1 proposal, got %d", len(proposals))
	}
	if got := proposals[0].TargetInvariantIDs; len(got) != 1 || got[0] != "inv_x" {
		t.Fatalf("break-all default must survive the pre-pass: %+v", got)
	}
}

// A large-but-under-limit valid payload is unaffected: the pre-pass adds key
// discipline, not a new size gate (MaxProposalResponseBytes already exists
// upstream of the decode and stays the only size limit).
func TestParseWireProposalsStrictPassUnaffectedBySize(t *testing.T) {
	t.Parallel()
	bulk := strings.Repeat("a very long novelty argument ", 4096) // ~120KB, well under the 4MB cap
	raw := strings.Replace(validWire, `"novelty_argument": "new",`,
		`"novelty_argument": "`+bulk+`",`, 1)
	if len(raw) > MaxProposalResponseBytes {
		t.Fatal("test bug: payload accidentally exceeds the wire size limit")
	}
	if _, err := ParseWireProposals(raw, []string{"inv_x"}); err != nil {
		t.Fatalf("large valid payload must still parse: %v", err)
	}
}

// Malformed JSON must never become acceptable because of the pre-pass: the
// pass defers syntax judgment to the main decode, which still rejects.
func TestParseWireProposalsStrictPassKeepsMalformedRejected(t *testing.T) {
	t.Parallel()
	for name, raw := range map[string]string{
		"truncated":     `{"schema_version": "proposal-wire/v1", "proposals": [`,
		"bare garbage":  `not-json-at-all`,
		"unclosed key":  `{"schema_version`,
		"number as key": `{42: "x"}`,
	} {
		if _, err := ParseWireProposals(raw, nil); !errors.Is(err, ErrProposalWireViolation) {
			t.Fatalf("%s: malformed JSON must stay a wire violation, got %v", name, err)
		}
	}
}
