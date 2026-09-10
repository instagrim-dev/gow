package provider

import (
	"context"
	"testing"
)

func TestDerivingFixturePolicyMutatorEmitsCodeDerivableDirectives(t *testing.T) {
	t.Parallel()
	ev := PolicyEvidence{
		ProblemID: "prb_x",
		Successes: []PolicySuccessFact{
			{PredicateFingerprint: "fp1", Strength: "deterministic", DistinctSupport: 2},
			{PredicateFingerprint: "fp2", Strength: "deterministic", DistinctSupport: 0}, // dropped: no support
		},
		Surviving:         []PolicySurvivingFact{{InvariantID: "inv_a", Attested: true}},
		UncoveredFamilies: []string{"mcl_1"},
		RedundantAttacks:  []string{"k1"},
		RepeatedFailures:  []string{"fpmech"},
	}
	resp, err := NewDerivingFixturePolicyMutator().Mutate(context.Background(), ev)
	if err != nil {
		t.Fatalf("Mutate() error = %v", err)
	}
	// prefer(fp1) + avoid(inv_a) + expand(mcl_1) + penalize(k1) + penalize(fpmech) = 5.
	if len(resp.Proposals) != 5 {
		t.Fatalf("expected 5 proposals (unsupported success dropped), got %d: %+v", len(resp.Proposals), resp.Proposals)
	}
	if resp.Metadata.ProviderVersion != PolicyMutatorVersion {
		t.Fatalf("provider version = %q, want %q", resp.Metadata.ProviderVersion, PolicyMutatorVersion)
	}
	if resp.RequestPayload == "" || resp.ResponsePayload == "" {
		t.Fatal("payloads must be retained for provenance")
	}
}

func TestDerivingFixturePolicyMutatorIsDeterministic(t *testing.T) {
	t.Parallel()
	ev := PolicyEvidence{Surviving: []PolicySurvivingFact{{InvariantID: "inv_a"}, {InvariantID: "inv_b"}}}
	a, _ := NewDerivingFixturePolicyMutator().Mutate(context.Background(), ev)
	b, _ := NewDerivingFixturePolicyMutator().Mutate(context.Background(), ev)
	if len(a.Proposals) != len(b.Proposals) {
		t.Fatal("mutator not deterministic")
	}
	for i := range a.Proposals {
		if a.Proposals[i] != b.Proposals[i] {
			t.Fatalf("mutator not deterministic at %d", i)
		}
	}
}

func TestPolicyEvidenceFingerprintOrderIndependent(t *testing.T) {
	t.Parallel()
	a := PolicyEvidence{
		ProblemID: "prb_x",
		Successes: []PolicySuccessFact{{PredicateFingerprint: "z"}, {PredicateFingerprint: "a"}},
		Surviving: []PolicySurvivingFact{{InvariantID: "inv_z"}, {InvariantID: "inv_a"}},
	}
	b := PolicyEvidence{
		ProblemID: "prb_x",
		Successes: []PolicySuccessFact{{PredicateFingerprint: "a"}, {PredicateFingerprint: "z"}},
		Surviving: []PolicySurvivingFact{{InvariantID: "inv_a"}, {InvariantID: "inv_z"}},
	}
	if a.Fingerprint() != b.Fingerprint() {
		t.Fatal("evidence fingerprint must be order-independent")
	}
}
