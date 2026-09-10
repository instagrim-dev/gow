package provider

import (
	"context"
	"testing"

	"github.com/instagrim-dev/newf/internal/invariant"
)

func validProposal() CandidateProposal {
	return CandidateProposal{
		Predicate: invariant.Predicate{
			Schema: invariant.PredicateSchemaV1,
			Root:   invariant.Node{Op: invariant.OpContains, Field: invariant.FieldPreserves, CanonicalID: "domain.number_theory.property.residue_locality"},
		},
		Statement:        "failed methods preserve residue locality",
		AbstractionLevel: "mechanism",
	}
}

func TestFixtureMinerDeterministicAndRecordsRole(t *testing.T) {
	req := MiningRequest{ProblemID: "prb_x", FailureSpaceID: "fsp_x", MinSupport: 2, Families: []MiningFamily{{ClusterID: "mcl_a"}}}
	m := NewFixtureInvariantMiner(map[string][]CandidateProposal{req.Fingerprint(): {validProposal()}})
	a, err := m.Mine(context.Background(), req)
	if err != nil {
		t.Fatalf("mine: %v", err)
	}
	b, _ := m.Mine(context.Background(), req)
	if a.ResponsePayload != b.ResponsePayload {
		t.Fatal("fixture miner must be deterministic")
	}
	if len(a.Proposals) != 1 {
		t.Fatalf("expected 1 proposal, got %d", len(a.Proposals))
	}
	if a.Metadata.SchemaVersion != invariant.PredicateSchemaV1 {
		t.Fatalf("metadata must record predicate schema, got %q", a.Metadata.SchemaVersion)
	}
}

func TestFixtureMinerRejectsInvalidPredicate(t *testing.T) {
	req := MiningRequest{ProblemID: "prb_x", FailureSpaceID: "fsp_x"}
	bad := CandidateProposal{Predicate: invariant.Predicate{Schema: invariant.PredicateSchemaV1, Root: invariant.Node{Op: "frobnicate"}}, Statement: "x"}
	m := NewFixtureInvariantMiner(map[string][]CandidateProposal{req.Fingerprint(): {bad}})
	if _, err := m.Mine(context.Background(), req); err == nil {
		t.Fatal("expected error for invalid predicate, got nil")
	}
}

func TestFixtureMinerUnregisteredIsEmpty(t *testing.T) {
	m := NewFixtureInvariantMiner(nil)
	resp, err := m.Mine(context.Background(), MiningRequest{ProblemID: "prb_y"})
	if err != nil {
		t.Fatalf("unregistered request must not error: %v", err)
	}
	if len(resp.Proposals) != 0 {
		t.Fatalf("expected empty proposals, got %d", len(resp.Proposals))
	}
}

func TestMiningRequestFingerprintOrderIndependent(t *testing.T) {
	a := MiningRequest{ProblemID: "p", FailureSpaceID: "f", Families: []MiningFamily{{ClusterID: "b"}, {ClusterID: "a"}}}
	b := MiningRequest{ProblemID: "p", FailureSpaceID: "f", Families: []MiningFamily{{ClusterID: "a"}, {ClusterID: "b"}}}
	if a.Fingerprint() != b.Fingerprint() {
		t.Fatal("request fingerprint must be family-order independent")
	}
}
