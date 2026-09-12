package provider

import (
	"testing"
)

// TestGenerationRequestFingerprintExpansions pins the request-boundary
// contract for policy expansions: an absent expansion set leaves the
// fingerprint IDENTICAL to the pre-expansion contract (nil and empty are the
// same request), and adding an expansion forks it.
func TestGenerationRequestFingerprintExpansions(t *testing.T) {
	base := GenerationRequest{ProblemID: "prb_x", Count: 3}
	if base.Fingerprint() != (GenerationRequest{ProblemID: "prb_x", Count: 3, Expansions: []GenerationExpansion{}}).Fingerprint() {
		t.Fatal("nil and empty expansions must fingerprint identically (omitempty back-compat)")
	}
	expanded := GenerationRequest{ProblemID: "prb_x", Count: 3, Expansions: []GenerationExpansion{
		{TargetKind: "refuted_boundary", TargetID: "fp-a", Weight: "medium", Source: "counterexample-separation"},
	}}
	if base.Fingerprint() == expanded.Fingerprint() {
		t.Fatal("an expansion must fork the request fingerprint")
	}
	// Order-independence: the fingerprint sorts expansions.
	a := GenerationRequest{ProblemID: "prb_x", Expansions: []GenerationExpansion{
		{TargetKind: "mechanism_family", TargetID: "cl_1", Weight: "medium"},
		{TargetKind: "refuted_boundary", TargetID: "fp-a", Weight: "medium"},
	}}
	b := GenerationRequest{ProblemID: "prb_x", Expansions: []GenerationExpansion{
		{TargetKind: "refuted_boundary", TargetID: "fp-a", Weight: "medium"},
		{TargetKind: "mechanism_family", TargetID: "cl_1", Weight: "medium"},
	}}
	if a.Fingerprint() != b.Fingerprint() {
		t.Fatal("expansion order must not fork the fingerprint")
	}
	// The serialized payload carries the expansions (this is what the
	// persisted invocation envelope shows an auditor).
	if expanded.Fingerprint() == "" {
		t.Fatal("fingerprint must be non-empty")
	}
}
