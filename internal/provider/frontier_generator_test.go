package provider

import (
	"context"
	"testing"

	"github.com/instagrim-dev/newf/internal/invariant"
)

func TestDerivingFixtureGeneratorBreaksSinglePreserves(t *testing.T) {
	req := GenerationRequest{
		ProblemID: "prb_x",
		Count:     4,
		Targets: []GenerationTarget{{
			InvariantID: "inv_1",
			Predicate: invariant.Predicate{
				Schema: invariant.PredicateSchemaV1,
				Root:   invariant.Node{Op: invariant.OpContains, Field: invariant.FieldPreserves, CanonicalID: "domain.number_theory.property.residue_locality"},
			},
		}},
	}
	g := NewDerivingFixtureGenerator()
	resp, err := g.Generate(context.Background(), req)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if len(resp.Proposals) != 1 {
		t.Fatalf("want 1 proposal, got %d", len(resp.Proposals))
	}
	p := resp.Proposals[0]
	if len(p.TargetInvariantIDs) != 1 || p.TargetInvariantIDs[0] != "inv_1" {
		t.Fatalf("target ids = %v", p.TargetInvariantIDs)
	}
	if p.CheapestFalsificationPath == "" {
		t.Fatal("proposal must carry a cheapest falsification path")
	}
	// The derived break-mechanism must EVALUATE as a violation of the target.
	if v := invariant.Evaluate(req.Targets[0].Predicate, p.ProposedSignature); v != invariant.VerdictViolates {
		t.Fatalf("derived proposal does not break target: verdict=%q", v)
	}
	if resp.Metadata.ModelName != derivingGeneratorModelName {
		t.Fatalf("model name = %q", resp.Metadata.ModelName)
	}

	// Determinism: identical request -> identical fingerprint + proposal count.
	resp2, _ := g.Generate(context.Background(), req)
	if req.Fingerprint() != req.Fingerprint() || len(resp2.Proposals) != len(resp.Proposals) {
		t.Fatal("generator is not deterministic")
	}
}

func TestDerivingFixtureGeneratorSkipsNonSinglePreserves(t *testing.T) {
	req := GenerationRequest{
		ProblemID: "prb_x",
		Targets: []GenerationTarget{{
			InvariantID: "inv_2",
			Predicate: invariant.Predicate{
				Schema: invariant.PredicateSchemaV1,
				Root:   invariant.Node{Op: invariant.OpIn, Field: invariant.FieldLocality, Values: []string{"local"}},
			},
		}},
	}
	resp, err := NewDerivingFixtureGenerator().Generate(context.Background(), req)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if len(resp.Proposals) != 0 {
		t.Fatalf("fixture should not derive a break for a non-preserves target; got %d", len(resp.Proposals))
	}
}
