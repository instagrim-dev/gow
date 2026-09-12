package provider

import (
	"context"
	"testing"

	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/invariant"
)

// TestDerivingMinerDefaultDoesNotEmitPostureAxes documents the backward-
// compatibility contract of NewDerivingFixtureInvariantMiner: even when the
// families ARE eligible for posture-axis emission, the DEFAULT constructor
// does not emit them.
func TestDerivingMinerDefaultDoesNotEmitPostureAxes(t *testing.T) {
	// Two failure families sharing (locality=local, construction=constructive,
	// uncertainty=deterministic), plus one contrast family. Under the
	// posture-axes constructor these would emit `Equals(locality, local)` and
	// `Equals(construction, constructive)` (contrast=0). The default must not.
	req := postureAxisRequest()
	m := NewDerivingFixtureInvariantMiner()
	resp, err := m.Mine(context.Background(), req)
	if err != nil {
		t.Fatalf("mine: %v", err)
	}
	for _, p := range resp.Proposals {
		if p.Predicate.Root.Op == invariant.OpEquals {
			t.Fatalf("default deriving miner must not emit OpEquals proposals, got %+v", p.Predicate.Root)
		}
	}
	if m.Identity().ModelName != derivingMinerModelName {
		t.Fatalf("default miner ModelName = %q, want %q", m.Identity().ModelName, derivingMinerModelName)
	}
}

// TestDerivingMinerPostureAxesEmitsRecurringOnly verifies the (a‴-B) pre-filter:
// only (axis, value) pairs with failure-side count >= 2 AND failure prevalence
// strictly exceeding success prevalence are emitted. Ordering is deterministic
// (axis, then value alphabetical).
func TestDerivingMinerPostureAxesEmitsRecurringOnly(t *testing.T) {
	req := postureAxisRequest()
	m := NewDerivingFixtureInvariantMinerWithPostureAxes()
	resp, err := m.Mine(context.Background(), req)
	if err != nil {
		t.Fatalf("mine: %v", err)
	}

	// Failure-side (2): both (local, constructive, deterministic).
	// Success-side (1): (global, constructive, deterministic).
	// Predictions per the classify-mirror pre-filter:
	//   locality=local: F=2/2=1.0, S=0/1=0.0 -> emit
	//   construction=constructive: F=2/2=1.0, S=1/1=1.0 -> DO NOT emit (F not > S)
	//   uncertainty=deterministic: F=2/2=1.0, S=1/1=1.0 -> DO NOT emit (F not > S)
	//   locality=global: F=0/2=0.0 (support<2) -> DO NOT emit
	got := postureEmissions(resp.Proposals)
	want := []axisValue{{axis: invariant.FieldLocality, value: string(domain.LocalityLocal)}}
	if !axisValueSlicesEqual(got, want) {
		t.Fatalf("emitted posture-axis proposals mismatch:\n  got  %+v\n  want %+v\n  all: %+v", got, want, proposalRoots(resp.Proposals))
	}

	// ModelName must differ from the default so the reuse key differs.
	if m.Identity().ModelName == derivingMinerModelName {
		t.Fatalf("posture-axes miner ModelName must differ from default; both were %q", derivingMinerModelName)
	}
	if m.Identity().ModelName != derivingMinerModelNamePostureAxes {
		t.Fatalf("posture-axes miner ModelName = %q, want %q", m.Identity().ModelName, derivingMinerModelNamePostureAxes)
	}
}

// TestDerivingMinerPostureAxesRequiresContrastMargin checks that a posture axis
// value dominant on the failure side but ALSO dominant on the success side is
// NOT emitted. This is the anti-inflation guard: without it, a corpus in
// which one axis value happens to be modal on both sides would clutter the
// candidate set with non-recurring proposals.
func TestDerivingMinerPostureAxesRequiresContrastMargin(t *testing.T) {
	// Three failure families and three success families, all with
	// uncertainty=deterministic. Nothing on the uncertainty axis
	// discriminates: F=3/3=1.0, S=3/3=1.0 -> not emitted.
	// Locality: 3 failure=local, 3 success=global. F prev locality=local = 1.0,
	// S prev = 0.0 -> emit.
	fam := func(cid string, out domain.OutcomeClass, loc domain.Locality) MiningFamily {
		return MiningFamily{
			ClusterID: cid, OutcomeClass: out,
			Locality:     loc,
			Construction: domain.ConstructionConstructive,
			Uncertainty:  domain.UncertaintyDeterministic,
		}
	}
	req := MiningRequest{Families: []MiningFamily{
		fam("f1", domain.OutcomeFailure, domain.LocalityLocal),
		fam("f2", domain.OutcomeFailure, domain.LocalityLocal),
		fam("f3", domain.OutcomeFailure, domain.LocalityLocal),
		fam("s1", domain.OutcomePartialSuccess, domain.LocalityGlobal),
		fam("s2", domain.OutcomePartialSuccess, domain.LocalityGlobal),
		fam("s3", domain.OutcomePartialSuccess, domain.LocalityGlobal),
	}}
	m := NewDerivingFixtureInvariantMinerWithPostureAxes()
	resp, err := m.Mine(context.Background(), req)
	if err != nil {
		t.Fatalf("mine: %v", err)
	}
	got := postureEmissions(resp.Proposals)
	want := []axisValue{{axis: invariant.FieldLocality, value: string(domain.LocalityLocal)}}
	if !axisValueSlicesEqual(got, want) {
		t.Fatalf("posture-axis emissions mismatch:\n  got  %+v\n  want %+v", got, want)
	}
}

// TestDerivingMinerPostureAxesIgnoresUnknownValues verifies that an
// axis value of "" or "unknown" on a failure family is not counted (the
// enum-axis validator would reject an equals("unknown") predicate anyway).
func TestDerivingMinerPostureAxesIgnoresUnknownValues(t *testing.T) {
	// Two failure families: one with locality=local, one with locality=unknown.
	// Under the pre-filter, locality=local has count=1 (below min support=2),
	// so no locality-axis proposal is emitted.
	req := MiningRequest{Families: []MiningFamily{
		{ClusterID: "f1", OutcomeClass: domain.OutcomeFailure, Locality: domain.LocalityLocal},
		{ClusterID: "f2", OutcomeClass: domain.OutcomeFailure, Locality: domain.LocalityUnknown},
	}}
	m := NewDerivingFixtureInvariantMinerWithPostureAxes()
	resp, err := m.Mine(context.Background(), req)
	if err != nil {
		t.Fatalf("mine: %v", err)
	}
	if got := postureEmissions(resp.Proposals); len(got) != 0 {
		t.Fatalf("expected no posture-axis proposals when only one family has a known value; got %+v", got)
	}
}

// TestDerivingMinerPostureAxesReuseKeyDiffers verifies that enabling
// posture-axis emission changes the miner Identity's Version() output, so the
// invariant-revision reuse key differs from a non-posture mining pass.
func TestDerivingMinerPostureAxesReuseKeyDiffers(t *testing.T) {
	a := NewDerivingFixtureInvariantMiner().Identity().Version()
	b := NewDerivingFixtureInvariantMinerWithPostureAxes().Identity().Version()
	if a == b {
		t.Fatalf("posture-axes miner must produce a distinct Identity.Version() from default; both were %q", a)
	}
}

// --- helpers ---

// postureAxisRequest returns a canonical 2-failure + 1-success request used
// by the tests above: failure families share (local, constructive, det) and
// the contrast family is (global, constructive, det).
func postureAxisRequest() MiningRequest {
	fam := func(cid string, out domain.OutcomeClass, loc domain.Locality) MiningFamily {
		return MiningFamily{
			ClusterID: cid, OutcomeClass: out,
			Locality:     loc,
			Construction: domain.ConstructionConstructive,
			Uncertainty:  domain.UncertaintyDeterministic,
			Preserves:    []domain.CanonicalID{"domain.number_theory.property.residue_locality"},
		}
	}
	return MiningRequest{Families: []MiningFamily{
		fam("f1", domain.OutcomeFailure, domain.LocalityLocal),
		fam("f2", domain.OutcomeFailure, domain.LocalityLocal),
		fam("s1", domain.OutcomePartialSuccess, domain.LocalityGlobal),
	}}
}

type axisValue struct {
	axis  string
	value string
}

func postureEmissions(proposals []CandidateProposal) []axisValue {
	out := make([]axisValue, 0)
	for _, p := range proposals {
		if p.Predicate.Root.Op != invariant.OpEquals {
			continue
		}
		if len(p.Predicate.Root.Values) != 1 {
			continue
		}
		out = append(out, axisValue{axis: p.Predicate.Root.Field, value: p.Predicate.Root.Values[0]})
	}
	return out
}

func axisValueSlicesEqual(a, b []axisValue) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func proposalRoots(proposals []CandidateProposal) []invariant.Node {
	out := make([]invariant.Node, 0, len(proposals))
	for _, p := range proposals {
		out = append(out, p.Predicate.Root)
	}
	return out
}
