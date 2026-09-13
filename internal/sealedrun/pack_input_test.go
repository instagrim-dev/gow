package sealedrun

import (
	"strings"
	"testing"
)

func validPackJSON() string {
	return `{
		"schema": "shaping-pack/1",
		"label": "development/file-pack",
		"provenance": "implementer-authored decoder boundary fixture, 2026-09-13",
		"episodes": [{
			"id": "file-double-not",
			"stratum": "history_informative",
			"family": "double-not",
			"start": {"op": "not", "args": [{"op": "not", "args": [{"var": "x"}]}]},
			"variables": ["x"],
			"catalog": ["double-not"],
			"target_cost": 1,
			"history": [{
				"start": "(not (not x))",
				"rules_applied": ["double-not"],
				"final_cost": 1,
				"target": 1,
				"completed": true,
				"endpoint": "HOLDS_ON_DECLARED_DOMAIN"
			}]
		}]
	}`
}

func TestDecodeShapingPackRunsThroughTheRealDiagnostic(t *testing.T) {
	pack, err := DecodeShapingPack([]byte(validPackJSON()))
	if err != nil {
		t.Fatal(err)
	}
	if pack.Label != "development/file-pack" || len(pack.Episodes) != 1 {
		t.Fatalf("decoded pack lost identity: %+v", pack)
	}
	ep := pack.Episodes[0]
	if ep.Decl.ID != "file-double-not" || len(ep.History) != 1 || !ep.History[0].Completed || ep.History[0].FinalCost != 1 {
		t.Fatalf("decoded episode lost content: %+v", ep)
	}
	budget := ResourceBudget{Expansions: 2, RuleApplications: 128, Candidates: 256,
		HistoryBytes: 65536, CheckAssignments: 4096, MaxStates: 1024, MaxTermNodes: 1024}
	receipt, err := RunResourceDiagnostic(pack, budget, nil)
	if err != nil {
		t.Fatalf("decoded pack must execute through the unchanged runner: %v (%+v)", err, receipt)
	}
	if receipt.Assessment != "completed-development-diagnostic" || len(receipt.Cells) != 4 {
		t.Fatalf("expected a complete four-arm development diagnostic: %+v", receipt)
	}
}

// A pack file's label is a caller claim. Naming a stronger evidence tier in
// the file must not change the receipt's forced development evidence label.
func TestDecodeShapingPackLabelCannotUpgradeEvidence(t *testing.T) {
	raw := strings.Replace(validPackJSON(), `"development/file-pack"`, `"agent-sealed/v1"`, 1)
	pack, err := DecodeShapingPack([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	budget := ResourceBudget{Expansions: 2, RuleApplications: 128, Candidates: 256,
		HistoryBytes: 65536, CheckAssignments: 4096, MaxStates: 1024, MaxTermNodes: 1024}
	receipt, err := RunResourceDiagnostic(pack, budget, nil)
	if err != nil {
		t.Fatal(err)
	}
	if receipt.EvidenceLabel != ResourceEvidenceLabel {
		t.Fatalf("file label upgraded the evidence grade: %q", receipt.EvidenceLabel)
	}
	if receipt.SourcePackLabel != "agent-sealed/v1" {
		t.Fatalf("the unverified source claim must be preserved, not erased: %q", receipt.SourcePackLabel)
	}
}

func TestDecodeShapingPackRefusesAmbiguousAndMalformedInput(t *testing.T) {
	valid := validPackJSON()
	cases := []struct {
		name, raw, want string
	}{
		{"wrong schema", strings.Replace(valid, "shaping-pack/1", "shaping-pack/2", 1), "unsupported shaping pack schema"},
		{"duplicate key", strings.Replace(valid, `"label": "development/file-pack",`, `"label": "a", "label": "b",`, 1), "duplicate"},
		{"unknown key", strings.Replace(valid, `"family"`, `"surprise"`, 1), "unknown"},
		{"case-variant key", strings.Replace(valid, `"schema"`, `"Schema"`, 1), "unknown, case-variant"},
		{"missing label", strings.Replace(valid, `"label": "development/file-pack",`, `"label": " ",`, 1), "label and provenance are required"},
		{"missing target_cost", strings.Replace(valid, `"target_cost": 1,`, "", 1), "missing required fields: target_cost"},
		{"missing episode id", strings.Replace(valid, `"id": "file-double-not",`, `"id": "",`, 1), "id is required"},
		{"undeclared stratum", strings.Replace(valid, "history_informative", "definitely_fresh", 1), "stratum must be one of"},
		{"missing attempt verdict", strings.Replace(valid, `"completed": true,`, "", 1), "missing required fields: completed"},
		{"two expression forms", strings.Replace(valid, `{"var": "x"}`, `{"var": "x", "const": 0}`, 1), "exactly one of var, const, or op"},
		{"trailing object", valid + `{}`, "exactly one JSON object"},
		{"not utf-8", "\xff\xfe", "valid UTF-8"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := DecodeShapingPack([]byte(tc.raw)); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected refusal containing %q, got %v", tc.want, err)
			}
		})
	}
	if _, err := DecodeShapingPack(make([]byte, MaxShapingPackBytes+1)); err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("oversized input must be refused before parsing: %v", err)
	}
}

// Format admission and semantic validation have exactly one owner each: a
// well-formed pack naming a rule outside the admissible menu decodes, and the
// runner is the boundary that refuses it.
func TestDecodedPackSemanticDefectsAreRefusedByTheRunner(t *testing.T) {
	raw := strings.Replace(validPackJSON(), `"catalog": ["double-not"]`, `"catalog": ["invented-rule"]`, 1)
	pack, err := DecodeShapingPack([]byte(raw))
	if err != nil {
		t.Fatalf("menu membership is the runner's boundary, not the decoder's: %v", err)
	}
	budget := ResourceBudget{Expansions: 2, RuleApplications: 128, Candidates: 256,
		HistoryBytes: 65536, CheckAssignments: 4096, MaxStates: 1024, MaxTermNodes: 1024}
	receipt, err := RunResourceDiagnostic(pack, budget, nil)
	if err == nil || !strings.Contains(err.Error(), `unknown menu rule "invented-rule"`) {
		t.Fatalf("runner must refuse the unknown rule: %v (%+v)", err, receipt)
	}
}
