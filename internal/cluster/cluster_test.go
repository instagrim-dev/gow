package cluster

import (
	"testing"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/domain"
)

// resolved builds a resolved FieldClaim for a field kind + canonical id.
func resolved(kind domain.FieldKind, id string) canon.FieldClaim {
	return canon.FieldClaim{
		FieldKind:    kind,
		SurfaceLabel: id,
		State:        domain.ResolutionResolved,
		CanonicalID:  domain.CanonicalID(id),
		Status:       domain.ClaimExplicit,
	}
}

// unresolved builds an ambiguous (non-resolved) claim so a field is incomparable.
func unresolved(kind domain.FieldKind, label string) canon.FieldClaim {
	return canon.FieldClaim{
		FieldKind:    kind,
		SurfaceLabel: label,
		State:        domain.ResolutionAmbiguous,
		Status:       domain.ClaimInferred,
	}
}

// sig builds a minimal signature with the given mechanism id, decisive-field
// content, representation set, and outcome. Empty slices are fine.
func sig(mechID string, ops, preserves, reps []string, outcome domain.OutcomeClass) canon.MechanismSignature {
	s := canon.MechanismSignature{
		SchemaVersion:     canon.SchemaMechanismV1,
		VocabularyVersion: "mechanism/v1",
		MechanismID:       mechID,
		SignatureID:       "msig_" + mechID,
		OutcomeClass:      outcome,
		Posture: canon.Posture{
			Locality:     domain.LocalityLocal,
			Construction: domain.ConstructionConstructive,
			Uncertainty:  domain.UncertaintyDeterministic,
		},
		Representations:  []canon.FieldClaim{},
		Operators:        []canon.FieldClaim{},
		Assumptions:      []canon.FieldClaim{},
		Preserves:        []canon.FieldClaim{},
		Breaks:           []canon.FieldClaim{},
		AuxiliaryObjects: []canon.FieldClaim{},
		Boundaries:       []canon.Boundary{},
	}
	for _, o := range ops {
		s.Operators = append(s.Operators, resolved(domain.FieldOperator, o))
	}
	for _, p := range preserves {
		s.Preserves = append(s.Preserves, resolved(domain.FieldPreserves, p))
	}
	for _, r := range reps {
		s.Representations = append(s.Representations, resolved(domain.FieldRepresentation, r))
	}
	return s
}

func clusterOf(c Clustering, sigID string) int {
	for i, cl := range c.Clusters {
		for _, m := range cl.Members {
			if m.SignatureID == sigID {
				return i
			}
		}
	}
	return -1
}

func TestSameMechanismDifferentRepresentationClustersTogether(t *testing.T) {
	t.Parallel()
	// A and B share decisive fields (operators+preserves) but differ only in
	// representation -> surface-distinct+mechanism-near -> one family, redundant.
	a := sig("a", []string{"core.op.modular_decomposition"}, []string{"core.prop.residue_locality"}, []string{"core.rep.congruence_classes"}, domain.OutcomePartialFailure)
	b := sig("b", []string{"core.op.modular_decomposition"}, []string{"core.prop.residue_locality"}, []string{"core.rep.affine_lattice"}, domain.OutcomePartialFailure)

	got := BuildClustering([]canon.MechanismSignature{a, b}, Params{})
	if got.Coverage.DistinctFamilyCount != 1 {
		t.Fatalf("distinct families = %d, want 1", got.Coverage.DistinctFamilyCount)
	}
	if clusterOf(got, "msig_a") != clusterOf(got, "msig_b") {
		t.Fatal("a and b should be in the same family")
	}
	if got.Coverage.RedundantMemberCount == 0 {
		t.Fatal("surface-distinct+mechanism-near members should be flagged redundant")
	}
	if got.Status != StatusClean {
		t.Fatalf("status = %s, want clean", got.Status)
	}
}

func TestDistinctMechanismsStaySeparate(t *testing.T) {
	t.Parallel()
	// Different operators AND preserves -> mechanism-distinct -> two families.
	a := sig("a", []string{"core.op.modular_decomposition"}, []string{"core.prop.residue_locality"}, []string{"core.rep.congruence_classes"}, domain.OutcomePartialFailure)
	b := sig("b", []string{"core.op.energy_estimate"}, []string{"core.prop.energy_localization"}, []string{"core.rep.congruence_classes"}, domain.OutcomeFailure)

	got := BuildClustering([]canon.MechanismSignature{a, b}, Params{})
	if got.Coverage.DistinctFamilyCount != 2 {
		t.Fatalf("distinct families = %d, want 2", got.Coverage.DistinctFamilyCount)
	}
}

func TestTransitiveLinkageMergesFamily(t *testing.T) {
	t.Parallel()
	// a~b and b~c via shared decisive content -> one connected component.
	a := sig("a", []string{"op.x"}, []string{"p.1"}, []string{"r.1"}, domain.OutcomeFailure)
	b := sig("b", []string{"op.x"}, []string{"p.1"}, []string{"r.2"}, domain.OutcomeFailure)
	c := sig("c", []string{"op.x"}, []string{"p.1"}, []string{"r.3"}, domain.OutcomeFailure)
	got := BuildClustering([]canon.MechanismSignature{a, b, c}, Params{})
	if got.Coverage.DistinctFamilyCount != 1 {
		t.Fatalf("distinct families = %d, want 1 (transitive)", got.Coverage.DistinctFamilyCount)
	}
}

func TestIncomparableDecisiveFieldStaysIsolate(t *testing.T) {
	t.Parallel()
	// b has an ambiguous (unresolved) operator -> incomparable on a decisive
	// field vs everyone -> isolate singleton, never merged (KTD-4).
	a := sig("a", []string{"op.x"}, []string{"p.1"}, nil, domain.OutcomeFailure)
	b := sig("b", nil, []string{"p.1"}, nil, domain.OutcomeFailure)
	b.Operators = append(b.Operators, unresolved(domain.FieldOperator, "mystery op"))

	got := BuildClustering([]canon.MechanismSignature{a, b}, Params{})
	if got.Coverage.DistinctFamilyCount != 2 {
		t.Fatalf("distinct families = %d, want 2", got.Coverage.DistinctFamilyCount)
	}
	bIdx := clusterOf(got, "msig_b")
	if bIdx < 0 || !got.Clusters[bIdx].Isolate {
		t.Fatal("b should be an incomparable isolate singleton")
	}
}

func TestOrderIndependence(t *testing.T) {
	t.Parallel()
	a := sig("a", []string{"op.x"}, []string{"p.1"}, []string{"r.1"}, domain.OutcomeFailure)
	b := sig("b", []string{"op.x"}, []string{"p.1"}, []string{"r.2"}, domain.OutcomeFailure)
	c := sig("c", []string{"op.y"}, []string{"p.2"}, []string{"r.3"}, domain.OutcomeSuccess)

	forward := BuildClustering([]canon.MechanismSignature{a, b, c}, Params{})
	reverse := BuildClustering([]canon.MechanismSignature{c, b, a}, Params{})

	if len(forward.Clusters) != len(reverse.Clusters) {
		t.Fatalf("cluster count differs: %d vs %d", len(forward.Clusters), len(reverse.Clusters))
	}
	for i := range forward.Clusters {
		if forward.Clusters[i].Fingerprint != reverse.Clusters[i].Fingerprint {
			t.Fatalf("cluster %d fingerprint differs across input order: %s vs %s", i, forward.Clusters[i].Fingerprint, reverse.Clusters[i].Fingerprint)
		}
		if forward.Clusters[i].RepresentativeSignatureID != reverse.Clusters[i].RepresentativeSignatureID {
			t.Fatalf("cluster %d representative differs across input order", i)
		}
	}
	if forward.ThresholdsHash != reverse.ThresholdsHash {
		t.Fatal("thresholds hash must be order-independent")
	}
}

func TestUnderSampledAxes(t *testing.T) {
	t.Parallel()
	// One signature -> every set axis has <=1 distinct value -> under-sampled.
	a := sig("a", []string{"op.x"}, []string{"p.1"}, []string{"r.1"}, domain.OutcomeFailure)
	got := BuildClustering([]canon.MechanismSignature{a}, Params{})
	foundOps := false
	for _, ax := range got.Coverage.Axes {
		if ax.Axis == "operator" {
			foundOps = true
			if !ax.UnderSampled {
				t.Fatalf("operator axis with 1 distinct value should be under-sampled")
			}
		}
	}
	if !foundOps {
		t.Fatal("operator axis missing from coverage report")
	}
}

func TestDiscriminationLossDegradesRun(t *testing.T) {
	t.Parallel()
	// Two signatures identical in all resolved non-outcome structure but with
	// DIFFERENT outcome classes: the vocabulary erased every distinction that
	// separates a success from a failure -> discrimination loss -> degraded.
	a := sig("a", []string{"op.x"}, []string{"p.1"}, []string{"r.1"}, domain.OutcomeFailure)
	b := sig("b", []string{"op.x"}, []string{"p.1"}, []string{"r.1"}, domain.OutcomeSuccess)
	got := BuildClustering([]canon.MechanismSignature{a, b}, Params{})
	if got.Status != StatusDegraded {
		t.Fatalf("status = %s, want degraded", got.Status)
	}
	if len(got.DiscriminationLoss) == 0 {
		t.Fatal("expected a discrimination-loss finding")
	}
}

func TestProfileSwapChangesFamilies(t *testing.T) {
	t.Parallel()
	// a and b differ ONLY in representation. Under the default profile
	// (representation surface-only) they are mechanism-near -> one family.
	// Under a profile that makes representation decisive, they become distinct.
	a := sig("a", []string{"op.x"}, []string{"p.1"}, []string{"r.1"}, domain.OutcomeFailure)
	b := sig("b", []string{"op.x"}, []string{"p.1"}, []string{"r.2"}, domain.OutcomeFailure)

	def := BuildClustering([]canon.MechanismSignature{a, b}, Params{})
	if def.Coverage.DistinctFamilyCount != 1 {
		t.Fatalf("default profile families = %d, want 1", def.Coverage.DistinctFamilyCount)
	}

	repDecisive := canon.ComparisonProfile{
		Version:           "classify/rep-decisive-test",
		DecisiveSetFields: []domain.FieldKind{domain.FieldRepresentation, domain.FieldOperator, domain.FieldPreserves},
		SurfaceField:      "",
	}
	swapped := BuildClustering([]canon.MechanismSignature{a, b}, Params{Profile: repDecisive})
	if swapped.Coverage.DistinctFamilyCount != 2 {
		t.Fatalf("rep-decisive profile families = %d, want 2", swapped.Coverage.DistinctFamilyCount)
	}
	if swapped.ProfileVersion == def.ProfileVersion {
		t.Fatal("profile version should differ and be recorded")
	}
	if swapped.ThresholdsHash == def.ThresholdsHash {
		t.Fatal("thresholds hash must change with the decisive set")
	}
}
