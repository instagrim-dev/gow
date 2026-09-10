package provider

import (
	"context"
	"testing"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/domain"
)

func familiesFixture() []GenerationFamily {
	return []GenerationFamily{
		{ClusterID: "clus_a", OutcomeClass: domain.OutcomeUnknown, Preserves: []domain.CanonicalID{"core.property.residue_locality"}, Locality: domain.LocalityLocal},
		{ClusterID: "clus_b", OutcomeClass: domain.OutcomeUnknown, Preserves: []domain.CanonicalID{"core.property.finite_cover"}, Locality: domain.LocalityLocal},
	}
}

func TestSummarizeNextProposerDerivesPerFamilyRestatement(t *testing.T) {
	req := GenerationRequest{ProblemID: "prob_x", Count: 4, Families: familiesFixture()}
	resp, err := NewSummarizeNextProposer().Generate(context.Background(), req)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(resp.Proposals) != 2 {
		t.Fatalf("want one proposal per family (2), got %d", len(resp.Proposals))
	}
	// B1 must claim no structural violation and target no invariant: it is a
	// summary continuation, not a directed break.
	for _, p := range resp.Proposals {
		if len(p.TargetInvariantIDs) != 0 {
			t.Fatalf("summarize-next must target no invariants, got %v", p.TargetInvariantIDs)
		}
	}
	// Provenance must be distinct from the directed generator.
	if resp.Metadata.ModelName != summarizeNextModelName {
		t.Fatalf("model name = %q, want %q", resp.Metadata.ModelName, summarizeNextModelName)
	}
	if resp.Metadata.ProviderVersion != baselineProviderVersion {
		t.Fatalf("provider version = %q, want %q", resp.Metadata.ProviderVersion, baselineProviderVersion)
	}
	// The first proposal preserves the family's own preserved property (a
	// non-break) — verifies it will land mechanism-near the KNOWN family.
	if got := resp.Proposals[0].ProposedSignature.Preserves[0].CanonicalID; got != "core.property.residue_locality" {
		t.Fatalf("restatement preserves %q, want the family's own property", got)
	}
}

func TestSummarizeNextProposerEmptyFamiliesIsEmpty(t *testing.T) {
	resp, err := NewSummarizeNextProposer().Generate(context.Background(), GenerationRequest{ProblemID: "prob_x", Count: 3})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(resp.Proposals) != 0 {
		t.Fatalf("no families must yield no proposals, got %d", len(resp.Proposals))
	}
}

func TestBrainstormerEmitsCountGenericRedundantProposals(t *testing.T) {
	req := GenerationRequest{ProblemID: "prob_x", Count: 3, Families: familiesFixture()}
	resp, err := NewBrainstormer().Generate(context.Background(), req)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(resp.Proposals) != 3 {
		t.Fatalf("want Count (3) proposals, got %d", len(resp.Proposals))
	}
	if resp.Metadata.ModelName != brainstormModelName {
		t.Fatalf("model name = %q, want %q", resp.Metadata.ModelName, brainstormModelName)
	}
	// Brainstorm ignores families/invariants: no proposal targets anything.
	for _, p := range resp.Proposals {
		if len(p.TargetInvariantIDs) != 0 {
			t.Fatalf("brainstorm must target no invariants, got %v", p.TargetInvariantIDs)
		}
	}
	// All brainstorm proposals share ONE canonical fingerprint (surface label
	// varies, mechanism does not) — the redundancy the diversity metric measures.
	first := canon.Fingerprint(resp.Proposals[0].ProposedSignature)
	for i, p := range resp.Proposals[1:] {
		if fp := canon.Fingerprint(p.ProposedSignature); fp != first {
			t.Fatalf("brainstorm proposal %d fingerprint %q != %q; surface-only variation must not change canonical content", i+1, fp, first)
		}
	}
}

func TestBrainstormerZeroCountEmitsOne(t *testing.T) {
	resp, err := NewBrainstormer().Generate(context.Background(), GenerationRequest{ProblemID: "prob_x"})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(resp.Proposals) != 1 {
		t.Fatalf("zero count must emit one proposal, got %d", len(resp.Proposals))
	}
}
