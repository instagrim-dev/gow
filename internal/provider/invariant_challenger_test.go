package provider

import (
	"context"
	"testing"

	"github.com/instagrim-dev/newf/internal/invariant"
)

func challengeRequest(t *testing.T, predJSON string) ChallengeRequest {
	t.Helper()
	p, err := invariant.ParsePredicate(predJSON)
	if err != nil {
		t.Fatalf("parse predicate: %v", err)
	}
	return ChallengeRequest{
		ProblemID:            "prb_x",
		InvariantID:          "inv_x",
		PredicateJSON:        predJSON,
		PredicateFingerprint: invariant.Fingerprint(p),
		MinSupport:           2,
		Families:             []MiningFamily{{ClusterID: "mcl_a"}},
	}
}

const leafPredJSON = `{"schema":"invariant-predicate/v1","root":{"op":"contains","field":"preserves","canonical_id":"domain.number_theory.property.residue_locality"}}`

const anyPredJSON = `{"schema":"invariant-predicate/v1","root":{"op":"any","children":[` +
	`{"op":"contains","field":"preserves","canonical_id":"domain.number_theory.property.residue_locality"},` +
	`{"op":"contains","field":"operators","canonical_id":"core.operator.sieve"}]}}`

func TestDerivingChallengerDeterministicAndFixedOrder(t *testing.T) {
	c := NewDerivingFixtureChallenger()
	req := challengeRequest(t, leafPredJSON)
	a, err := c.Challenge(context.Background(), req)
	if err != nil {
		t.Fatalf("challenge: %v", err)
	}
	b, _ := c.Challenge(context.Background(), req)
	if a.ResponsePayload != b.ResponsePayload {
		t.Fatal("challenger must be deterministic")
	}
	wantOrder := []invariant.ChallengeType{
		invariant.ChallengeKnownCounterexample,
		invariant.ChallengeSuccessPreserving,
		invariant.ChallengeBiasCritique,
	}
	if len(a.Proposals) != len(wantOrder) {
		t.Fatalf("expected %d proposals for a leaf predicate, got %d", len(wantOrder), len(a.Proposals))
	}
	for i, want := range wantOrder {
		if a.Proposals[i].Type != want {
			t.Fatalf("proposal %d type = %s, want %s", i, a.Proposals[i].Type, want)
		}
	}
	if a.Metadata.SchemaVersion != invariant.PredicateSchemaV1 || a.RequestPayload == "" {
		t.Fatal("metadata + request payload must be recorded for replay")
	}
}

func TestDerivingChallengerProposesSplitForBooleanPredicates(t *testing.T) {
	c := NewDerivingFixtureChallenger()
	resp, err := c.Challenge(context.Background(), challengeRequest(t, anyPredJSON))
	if err != nil {
		t.Fatalf("challenge: %v", err)
	}
	var split *ChallengeProposal
	for i := range resp.Proposals {
		if resp.Proposals[i].Type == invariant.ChallengeSplit {
			split = &resp.Proposals[i]
		}
	}
	if split == nil {
		t.Fatal("expected a split proposal for an any-composed predicate")
	}
	if len(split.Children) != 2 {
		t.Fatalf("expected 2 split children, got %d", len(split.Children))
	}
	for _, ch := range split.Children {
		if err := ch.Validate(); err != nil {
			t.Fatalf("split child must be valid: %v", err)
		}
	}
}

func TestDerivingChallengerRejectsUnparsablePredicate(t *testing.T) {
	c := NewDerivingFixtureChallenger()
	req := ChallengeRequest{PredicateJSON: `{"statement":"sounds plausible"}`}
	if _, err := c.Challenge(context.Background(), req); err == nil {
		t.Fatal("expected error for unparsable candidate predicate")
	}
}

func TestChallengeRequestFingerprintOrderIndependent(t *testing.T) {
	a := ChallengeRequest{InvariantID: "inv_1", Families: []MiningFamily{{ClusterID: "b"}, {ClusterID: "a"}}, SiblingFingerprints: []string{"z", "y"}}
	b := ChallengeRequest{InvariantID: "inv_1", Families: []MiningFamily{{ClusterID: "a"}, {ClusterID: "b"}}, SiblingFingerprints: []string{"y", "z"}}
	if a.Fingerprint() != b.Fingerprint() {
		t.Fatal("challenge request fingerprint must be order-independent")
	}
}
