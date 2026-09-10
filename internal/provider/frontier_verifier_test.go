package provider

import (
	"context"
	"testing"

	"github.com/instagrim-dev/newf/internal/verify"
)

func TestFixtureModelVerifierStampsSingleModelJudgment(t *testing.T) {
	m := NewFixtureModelVerifier(verify.VerdictSuccess, "medium")
	d, err := m.Verify(context.Background(), verify.VerificationContext{ProposalID: "fpr_x", ClaimedViolation: "breaks X"})
	if err != nil {
		t.Fatal(err)
	}
	if d.Kind != verify.KindModelJudgment || d.Strength != verify.StrengthSingleModelJudgment {
		t.Fatalf("model tier must stamp model-judgment/single-model-judgment; got %q/%q", d.Kind, d.Strength)
	}
	if d.Verdict != verify.VerdictSuccess || d.ConfidenceOrdinal != "medium" {
		t.Fatalf("got %q/%q", d.Verdict, d.ConfidenceOrdinal)
	}
	req, resp := m.LastPayloads()
	if req == "" || resp == "" {
		t.Fatal("model tier must retain request/response payloads for provenance")
	}
	if m.Metadata().SchemaVersion != VerifierVersion {
		t.Fatalf("metadata schema = %q", m.Metadata().SchemaVersion)
	}
}

func TestFixtureModelVerifierDefaultsToUnknown(t *testing.T) {
	m := NewFixtureModelVerifier("", "")
	d, _ := m.Verify(context.Background(), verify.VerificationContext{})
	if d.Verdict != verify.VerdictUnknown {
		t.Fatalf("a bare fixture must not manufacture a decisive verdict; got %q", d.Verdict)
	}
}
