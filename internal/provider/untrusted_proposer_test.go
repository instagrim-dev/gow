package provider

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/instagrim-dev/newf/internal/domain"
)

type stringTransport struct{ raw string }

func (t stringTransport) Fetch(context.Context, GenerationRequest) (string, error) {
	return t.raw, nil
}

type failingTransport struct{}

func (failingTransport) Fetch(context.Context, GenerationRequest) (string, error) {
	return "", ErrProviderTransport
}

const validWire = `{
  "schema_version": "proposal-wire/v1",
  "proposals": [{
    "mechanism": {
      "preserves": ["residue locality"],
      "operators": ["modular decomposition"],
      "locality": "global",
      "construction_mode": "constructive",
      "uncertainty_mode": "deterministic"
    },
    "structural_violation_claim": "drops the shared property",
    "novelty_argument": "new",
    "cheapest_falsification_path": "check",
    "expected_information_gain": "medium",
    "evaluation_cost": "low"
  }]
}`

func TestUntrustedProposerParsesLabelOnlyClaims(t *testing.T) {
	t.Parallel()
	p := NewUntrustedProposer(stringTransport{raw: validWire}, Metadata{ProviderName: "test"})
	resp, err := p.Generate(context.Background(), GenerationRequest{
		Targets: []GenerationTarget{{InvariantID: "inv_x"}},
	})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(resp.Proposals) != 1 {
		t.Fatalf("want 1 proposal, got %d", len(resp.Proposals))
	}
	prop := resp.Proposals[0]
	if len(prop.TargetInvariantIDs) != 1 || prop.TargetInvariantIDs[0] != "inv_x" {
		t.Fatalf("proposal must target the supplied surviving set: %+v", prop.TargetInvariantIDs)
	}
	sig := prop.ProposedSignature
	if len(sig.Preserves) != 1 || sig.Preserves[0].SurfaceLabel != "residue locality" {
		t.Fatalf("label must be carried: %+v", sig.Preserves)
	}
	// The wire cannot pre-assert authority: claims arrive UNRESOLVED with no
	// canonical id, and no completeness map exists.
	if sig.Preserves[0].State != domain.ResolutionUnknown || sig.Preserves[0].CanonicalID != "" {
		t.Fatalf("wire claims must be unresolved: %+v", sig.Preserves[0])
	}
	if sig.SetFieldCompleteness != nil {
		t.Fatalf("the wire must not express completeness: %+v", sig.SetFieldCompleteness)
	}
	// Payload retention: raw response persisted verbatim for audit.
	if resp.ResponsePayload != validWire || resp.RequestPayload == "" {
		t.Fatal("request/response payloads must be retained")
	}
}

func TestUntrustedProposerRejectsSmuggledAuthority(t *testing.T) {
	t.Parallel()
	// canonical_id / field_completeness are not part of the wire schema;
	// strict decoding rejects them VISIBLY instead of silently dropping.
	smuggled := strings.Replace(validWire, `"preserves": ["residue locality"],`,
		`"preserves": ["residue locality"], "field_completeness": {"preserves": "complete"},`, 1)
	p := NewUntrustedProposer(stringTransport{raw: smuggled}, Metadata{})
	if _, err := p.Generate(context.Background(), GenerationRequest{}); !errors.Is(err, ErrProposalWireViolation) {
		t.Fatalf("smuggled authority fields must be a wire violation, got %v", err)
	}
}

func TestUntrustedProposerRejectsBadEnumAndVersion(t *testing.T) {
	t.Parallel()
	badEnum := strings.Replace(validWire, `"locality": "global"`, `"locality": "everywhere"`, 1)
	p := NewUntrustedProposer(stringTransport{raw: badEnum}, Metadata{})
	if _, err := p.Generate(context.Background(), GenerationRequest{}); !errors.Is(err, ErrProposalWireViolation) {
		t.Fatalf("invalid posture enum must be a wire violation, got %v", err)
	}
	badVersion := strings.Replace(validWire, "proposal-wire/v1", "proposal-wire/v99", 1)
	p = NewUntrustedProposer(stringTransport{raw: badVersion}, Metadata{})
	if _, err := p.Generate(context.Background(), GenerationRequest{}); !errors.Is(err, ErrProposalWireViolation) {
		t.Fatalf("wrong schema version must be a wire violation, got %v", err)
	}
}

func TestUntrustedProposerDistinguishesTransportFailure(t *testing.T) {
	t.Parallel()
	p := NewUntrustedProposer(failingTransport{}, Metadata{})
	_, err := p.Generate(context.Background(), GenerationRequest{})
	if !errors.Is(err, ErrProviderTransport) || errors.Is(err, ErrProposalWireViolation) {
		t.Fatalf("transport failure must stay distinct from a schema violation: %v", err)
	}
}

// The adapter must never be a trusted structure author: its output exists to
// exercise the admission boundary.
func TestUntrustedProposerIsNotTrusted(t *testing.T) {
	t.Parallel()
	var g Generator = NewUntrustedProposer(stringTransport{}, Metadata{})
	if _, trusted := g.(TrustedStructureAuthor); trusted {
		t.Fatal("UntrustedProposer must not implement TrustedStructureAuthor")
	}
}
