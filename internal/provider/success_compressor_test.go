package provider

import (
	"context"
	"testing"

	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/invariant"
)

func compressorRequest() CompressionRequest {
	return CompressionRequest{
		ProblemID:  "prb_x",
		MinSupport: 1,
		Cohorts: []BreakCohortFacts{{
			TargetInvariantID: "inv_P",
			Progressors: []CohortMemberFacts{
				{ProposalID: "fpr_1", Result: domain.OutcomePartialSuccess, Strength: "deterministic",
					Preserves: []domain.CanonicalID{"core.assumption.effective_bounds", "domain.number_theory.property.global_density"}},
				{ProposalID: "fpr_2", Result: domain.OutcomeSuccess, Strength: "deterministic",
					Preserves: []domain.CanonicalID{"core.assumption.effective_bounds"}},
			},
			NonProgressors: []CohortMemberFacts{
				{ProposalID: "fpr_3", Result: domain.OutcomeFailure, Strength: "deterministic",
					Preserves: []domain.CanonicalID{"domain.number_theory.property.global_density"}},
			},
		}},
	}
}

// The deriving rule: bounds is shared by ALL progressors and absent from the
// non-progressor -> proposed. global is not shared by all progressors -> not.
func TestDerivingCompressorDerivesDiscriminatingCondition(t *testing.T) {
	c := NewDerivingFixtureCompressor()
	resp, err := c.Compress(context.Background(), compressorRequest())
	if err != nil {
		t.Fatalf("compress: %v", err)
	}
	if len(resp.Proposals) != 1 {
		t.Fatalf("expected exactly 1 derived condition, got %d: %+v", len(resp.Proposals), resp.Proposals)
	}
	p := resp.Proposals[0]
	if p.TargetInvariantID != "inv_P" || p.Predicate.Root.CanonicalID != "core.assumption.effective_bounds" {
		t.Fatalf("unexpected derived condition: %+v", p)
	}
	if err := p.Predicate.Validate(); err != nil {
		t.Fatalf("derived condition must be valid: %v", err)
	}
	if resp.Metadata.SchemaVersion != invariant.PredicateSchemaV1 || resp.RequestPayload == "" {
		t.Fatal("metadata + payloads must be recorded for replay")
	}
}

func TestDerivingCompressorDeterministic(t *testing.T) {
	c := NewDerivingFixtureCompressor()
	a, _ := c.Compress(context.Background(), compressorRequest())
	b, _ := c.Compress(context.Background(), compressorRequest())
	if a.ResponsePayload != b.ResponsePayload {
		t.Fatal("compressor must be deterministic")
	}
}

// With no non-progressors, every all-progressor-shared id is proposed.
func TestDerivingCompressorNoContrastProposesShared(t *testing.T) {
	req := compressorRequest()
	req.Cohorts[0].NonProgressors = nil
	c := NewDerivingFixtureCompressor()
	resp, err := c.Compress(context.Background(), req)
	if err != nil {
		t.Fatalf("compress: %v", err)
	}
	if len(resp.Proposals) != 1 || resp.Proposals[0].Predicate.Root.CanonicalID != "core.assumption.effective_bounds" {
		t.Fatalf("expected shared-preserves condition, got %+v", resp.Proposals)
	}
}

func TestCompressionRequestFingerprintOrderIndependent(t *testing.T) {
	a := compressorRequest()
	b := compressorRequest()
	b.Cohorts[0].Progressors = []CohortMemberFacts{b.Cohorts[0].Progressors[1], b.Cohorts[0].Progressors[0]}
	if a.Fingerprint() != b.Fingerprint() {
		t.Fatal("request fingerprint must be member-order independent")
	}
}
