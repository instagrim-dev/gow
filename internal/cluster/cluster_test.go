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

// Regression: two signatures with byte-identical resolved fingerprints where
// exactly one carries an ADDITIONAL ambiguous decisive claim. The ambiguous
// claim makes the pair incomparable-on-decisive, so they do NOT link -> two
// singleton clusters. Because the ambiguous claim is excluded from the
// signature fingerprint, both members share the same fingerprint; keying the
// cluster fingerprint on member fingerprints alone would collide and violate
// UNIQUE(cluster_run_id, cluster_fingerprint) at persist. Cluster fingerprints
// must stay distinct because signature IDs are folded into the identity.
func TestClusterFingerprintDistinctForAmbiguitySeparatedIsolates(t *testing.T) {
	t.Parallel()
	a := sig("a", []string{"op.x"}, []string{"p.1"}, nil, domain.OutcomeFailure)
	b := sig("b", []string{"op.x"}, []string{"p.1"}, nil, domain.OutcomeFailure)
	b.Operators = append(b.Operators, unresolved(domain.FieldOperator, "mystery op"))

	// Precondition: the two signatures share a resolved fingerprint, so the
	// collision is reachable and this test is exercising the real hazard.
	if canon.Fingerprint(a) != canon.Fingerprint(b) {
		t.Fatalf("precondition failed: fingerprints differ, collision not reachable")
	}

	got := BuildClustering([]canon.MechanismSignature{a, b}, Params{})
	if got.Coverage.DistinctFamilyCount != 2 {
		t.Fatalf("distinct families = %d, want 2", got.Coverage.DistinctFamilyCount)
	}
	if got.Clusters[0].Fingerprint == got.Clusters[1].Fingerprint {
		t.Fatalf("distinct clusters share cluster fingerprint %q (collision -> UNIQUE violation at persist)", got.Clusters[0].Fingerprint)
	}
}

// TestNonTransitiveChainStaysCoherent is the KTD-2 regression: the mechanism-near
// relation is NOT transitive. With A~B and B~C but A mechanism-distinct from C,
// naive single-linkage connected components would manufacture one family {A,B,C}
// containing a mechanism-distinct pair. The coherence guard must keep them
// coherent: no family may contain a mechanism-distinct pair.
//
// Operator Jaccard on a single decisive axis (the user's counterexample):
//
//	A = {op.a, op.b}        J(A,B) = 2/3 -> near
//	B = {op.a, op.b, op.c}  J(B,C) = 2/3 -> near
//	C = {op.b, op.c}        J(A,C) = 1/3 -> distinct
func TestNonTransitiveChainStaysCoherent(t *testing.T) {
	t.Parallel()
	a := sig("a", []string{"op.a", "op.b"}, nil, nil, domain.OutcomeFailure)
	b := sig("b", []string{"op.a", "op.b", "op.c"}, nil, nil, domain.OutcomeFailure)
	c := sig("c", []string{"op.b", "op.c"}, nil, nil, domain.OutcomeFailure)

	// Precondition: confirm the intended non-transitive relation actually holds
	// under the comparator, so the test exercises the real hazard.
	prof := canon.ProfileMechanismV1()
	nearAB := isLink(canon.CompareWithProfile(a, b, prof).Classification)
	nearBC := isLink(canon.CompareWithProfile(b, c, prof).Classification)
	distinctAC := isDistinct(canon.CompareWithProfile(a, c, prof).Classification)
	if !nearAB || !nearBC || !distinctAC {
		t.Fatalf("precondition failed: nearAB=%t nearBC=%t distinctAC=%t (need true,true,true)", nearAB, nearBC, distinctAC)
	}

	got := BuildClustering([]canon.MechanismSignature{a, b, c}, Params{})

	// Family coherence invariant: no family contains a mechanism-distinct pair.
	for i, cl := range got.Clusters {
		if cl.IntraVariation.MechanismDistinct != 0 {
			t.Fatalf("family %d contains %d mechanism-distinct pairs; coherence violated", i, cl.IntraVariation.MechanismDistinct)
		}
	}
	// A and C must NOT share a family (they are mechanism-distinct).
	if clusterOf(got, "msig_a") == clusterOf(got, "msig_c") {
		t.Fatal("mechanism-distinct A and C were merged into one family (single-linkage transitivity bug)")
	}
	// Determinism: reversed input yields the identical family partition.
	rev := BuildClustering([]canon.MechanismSignature{c, b, a}, Params{})
	if len(rev.Clusters) != len(got.Clusters) {
		t.Fatalf("coherent clustering not order-independent: %d vs %d families", len(rev.Clusters), len(got.Clusters))
	}
	for i := range got.Clusters {
		if got.Clusters[i].Fingerprint != rev.Clusters[i].Fingerprint {
			t.Fatalf("family %d fingerprint differs across input order", i)
		}
	}
}

// TestInputSetHashChangesWithPopulation is the KTD-1 regression: the cluster-run
// identity must depend on WHICH signatures were clustered, not just their count.
// Adding a signature (or swapping one for another of equal count) must change the
// input-set hash so re-clustering produces a new run instead of colliding with a
// stale run under the same version tuple.
func TestInputSetHashChangesWithPopulation(t *testing.T) {
	t.Parallel()
	a := sig("a", []string{"op.a"}, nil, nil, domain.OutcomeFailure)
	b := sig("b", []string{"op.b"}, nil, nil, domain.OutcomeFailure)
	c := sig("c", []string{"op.c"}, nil, nil, domain.OutcomeFailure)

	twelveA := BuildClustering([]canon.MechanismSignature{a, b}, Params{})
	plusC := BuildClustering([]canon.MechanismSignature{a, b, c}, Params{})
	if twelveA.InputSetHash == "" {
		t.Fatal("input set hash is empty for a non-empty population")
	}
	if twelveA.InputSetHash == plusC.InputSetHash {
		t.Fatal("adding a signature did not change the input-set hash (stale-run collision, KTD-1)")
	}

	// Same-count, different population must also differ.
	swap := BuildClustering([]canon.MechanismSignature{a, c}, Params{})
	if twelveA.InputSetHash == swap.InputSetHash {
		t.Fatal("different population of equal size shares an input-set hash (signature_count insufficient)")
	}

	// Order-independent for a fixed population.
	if BuildClustering([]canon.MechanismSignature{b, a}, Params{}).InputSetHash != twelveA.InputSetHash {
		t.Fatal("input-set hash is not order-independent")
	}
}

// TestFamilyOutcomeDistributionIsMixed is the KTD-9 regression: a family that
// spans more than one member outcome class must be reported as mixed, not
// compressed to the outcome of its representative signature.
func TestFamilyOutcomeDistributionIsMixed(t *testing.T) {
	t.Parallel()
	// a and b are the same mechanism (surface-distinct only) but carry DIFFERENT
	// outcome classes. Outcome is non-decisive for identity, so they form one
	// family; that family must be mixed.
	a := sig("a", []string{"op.x"}, []string{"p.1"}, []string{"r.1"}, domain.OutcomeFailure)
	b := sig("b", []string{"op.x"}, []string{"p.1"}, []string{"r.2"}, domain.OutcomePartialSuccess)

	got := BuildClustering([]canon.MechanismSignature{a, b}, Params{})
	if got.Coverage.DistinctFamilyCount != 1 {
		t.Fatalf("distinct families = %d, want 1 (same mechanism)", got.Coverage.DistinctFamilyCount)
	}
	fam := got.Clusters[0]
	if !fam.Mixed {
		t.Fatalf("family spanning failure+partial_success not marked mixed; outcomes=%v", fam.OutcomeClasses)
	}
	if fam.PrimaryOutcome() != domain.OutcomeMixed {
		t.Fatalf("PrimaryOutcome() = %q, want mixed", fam.PrimaryOutcome())
	}
	if len(fam.OutcomeClasses) != 2 {
		t.Fatalf("outcome distribution = %v, want 2 distinct classes", fam.OutcomeClasses)
	}
}

// TestSingleOutcomeFamilyIsNotMixed guards the boundary: a homogeneous family
// keeps its single outcome class and is not spuriously marked mixed.
func TestSingleOutcomeFamilyIsNotMixed(t *testing.T) {
	t.Parallel()
	a := sig("a", []string{"op.x"}, []string{"p.1"}, []string{"r.1"}, domain.OutcomeFailure)
	b := sig("b", []string{"op.x"}, []string{"p.1"}, []string{"r.2"}, domain.OutcomeFailure)

	got := BuildClustering([]canon.MechanismSignature{a, b}, Params{})
	fam := got.Clusters[0]
	if fam.Mixed {
		t.Fatal("homogeneous family incorrectly marked mixed")
	}
	if fam.PrimaryOutcome() != domain.OutcomeFailure {
		t.Fatalf("PrimaryOutcome() = %q, want failure", fam.PrimaryOutcome())
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
