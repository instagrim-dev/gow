package invariant

import (
	"testing"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/domain"
)

const (
	idResidue = "domain.number_theory.property.residue_locality"
	idSieve   = "domain.number_theory.operator.sieve"
)

func pred(root Node) Predicate { return Predicate{Schema: PredicateSchemaV1, Root: root} }

func resolved(kind domain.FieldKind, id domain.CanonicalID, status domain.ClaimStatus) canon.FieldClaim {
	return canon.FieldClaim{FieldKind: kind, State: domain.ResolutionResolved, CanonicalID: id, Status: status}
}

func baseSig() canon.MechanismSignature {
	return canon.MechanismSignature{
		Representations: []canon.FieldClaim{}, Operators: []canon.FieldClaim{},
		Assumptions: []canon.FieldClaim{}, Preserves: []canon.FieldClaim{},
		Breaks: []canon.FieldClaim{}, AuxiliaryObjects: []canon.FieldClaim{},
		Boundaries: []canon.Boundary{},
	}
}

// --- predicate contract (U1) ---

func TestValidateRejectsProseOnlyAndMalformed(t *testing.T) {
	cases := map[string]Predicate{
		"wrong schema":  {Schema: "prose", Root: Node{Op: OpContains, Field: FieldPreserves, CanonicalID: idResidue}},
		"unknown op":    pred(Node{Op: "frobnicate"}),
		"bad canonical": pred(Node{Op: OpContains, Field: FieldPreserves, CanonicalID: "residue_locality"}),
		"not-set field": pred(Node{Op: OpContains, Field: FieldOutcome, CanonicalID: idResidue}),
		"in bad value":  pred(Node{Op: OpIn, Field: FieldLocality, Values: []string{"nowhere"}}),
		"not arity":     pred(Node{Op: OpNot, Children: []Node{}}),
	}
	for name, p := range cases {
		if err := Validate(p); err == nil {
			t.Errorf("%s: expected validation error", name)
		}
	}
}

func TestFingerprintOrderIndependentAndParaphraseFree(t *testing.T) {
	a := pred(Node{Op: OpAll, Children: []Node{
		{Op: OpContains, Field: FieldPreserves, CanonicalID: idResidue},
		{Op: OpIn, Field: FieldLocality, Values: []string{"mixed", "local"}},
	}})
	b := pred(Node{Op: OpAll, Children: []Node{
		{Op: OpIn, Field: FieldLocality, Values: []string{"local", "mixed"}},
		{Op: OpContains, Field: FieldPreserves, CanonicalID: idResidue},
	}})
	if Fingerprint(a) != Fingerprint(b) {
		t.Fatal("order-permuted equivalent predicates must share a fingerprint")
	}
	c := pred(Node{Op: OpContains, Field: FieldOperators, CanonicalID: idSieve})
	if Fingerprint(a) == Fingerprint(c) {
		t.Fatal("distinct predicates must not collide")
	}
}

func TestEvaluateResolvedAbsentIsViolatesAmbiguousIsUnknown(t *testing.T) {
	sig := baseSig()
	sig.Preserves = []canon.FieldClaim{resolved(domain.FieldPreserves, domain.CanonicalID(idResidue), domain.ClaimExplicit)}
	if got := Evaluate(pred(Node{Op: OpContains, Field: FieldPreserves, CanonicalID: idResidue}), sig); got != VerdictSatisfies {
		t.Fatalf("expected satisfies, got %s", got)
	}
	if got := Evaluate(pred(Node{Op: OpContains, Field: FieldPreserves, CanonicalID: idSieve}), sig); got != VerdictViolates {
		t.Fatalf("resolved-but-absent must be violates, got %s", got)
	}
	amb := baseSig()
	amb.Preserves = []canon.FieldClaim{{FieldKind: domain.FieldPreserves, State: domain.ResolutionAmbiguous, Status: domain.ClaimAmbiguous}}
	if got := Evaluate(pred(Node{Op: OpContains, Field: FieldPreserves, CanonicalID: idResidue}), amb); got != VerdictUnknown {
		t.Fatalf("ambiguous read must be unknown, got %s", got)
	}
}

// --- engine (U3) ---

func failFamily(id string, ids ...domain.CanonicalID) Family {
	sig := baseSig()
	for _, cid := range ids {
		sig.Preserves = append(sig.Preserves, resolved(domain.FieldPreserves, cid, domain.ClaimExplicit))
	}
	return Family{
		ClusterID: id, OutcomeClass: domain.OutcomeFailure,
		Members: []Member{{SignatureID: "msig_" + id, Signature: sig, OutcomeClass: domain.OutcomeFailure}},
	}
}

func containsResidue() Proposal {
	return Proposal{Predicate: pred(Node{Op: OpContains, Field: FieldPreserves, CanonicalID: idResidue}), Statement: "preserves residue locality"}
}

func TestEngineRecurringAcrossDistinctFamilies(t *testing.T) {
	fams := []Family{failFamily("a", idResidue), failFamily("b", idResidue), failFamily("c", idResidue)}
	got := EvaluateCandidates([]Proposal{containsResidue()}, fams, 2)
	if len(got) != 1 {
		t.Fatalf("expected 1 candidate, got %d", len(got))
	}
	c := got[0]
	if c.DistinctFamilySupport != 3 {
		t.Fatalf("expected 3 distinct-family support, got %d", c.DistinctFamilySupport)
	}
	if c.AssociationStatus != AssocRecurring {
		t.Fatalf("expected recurring, got %s", c.AssociationStatus)
	}
	if c.FailureCoverageNum != 3 || c.FailureCoverageDen != 3 {
		t.Fatalf("expected coverage 3/3, got %d/%d", c.FailureCoverageNum, c.FailureCoverageDen)
	}
}

func TestEnginePredicateMatchingNothingEarnsZeroSupport(t *testing.T) {
	fams := []Family{failFamily("a", idResidue), failFamily("b", idResidue)}
	// A proposal claiming a different id: code disproves it regardless of prose.
	got := EvaluateCandidates([]Proposal{{Predicate: pred(Node{Op: OpContains, Field: FieldPreserves, CanonicalID: idSieve}), Statement: "sieve everywhere"}}, fams, 2)
	if got[0].DistinctFamilySupport != 0 {
		t.Fatalf("model cannot self-certify: expected 0 support, got %d", got[0].DistinctFamilySupport)
	}
	if len(got[0].Counterexamples) != 2 {
		t.Fatalf("expected 2 counterexamples, got %d", len(got[0].Counterexamples))
	}
}

func TestEngineRedundantSupporterDoesNotInflate(t *testing.T) {
	a := failFamily("a", idResidue)
	b := failFamily("b", idResidue)
	b.Redundant = true
	got := EvaluateCandidates([]Proposal{containsResidue()}, []Family{a, b}, 2)
	if got[0].DistinctFamilySupport != 1 {
		t.Fatalf("redundant supporter must not inflate: expected 1, got %d", got[0].DistinctFamilySupport)
	}
	if got[0].AssociationStatus == AssocRecurring {
		t.Fatal("should not be recurring with support below threshold")
	}
}

func TestEngineDiscriminativeWhenAbsentInSuccess(t *testing.T) {
	fams := []Family{failFamily("a", idResidue), failFamily("b", idResidue)}
	// A success family that does NOT preserve residue (resolved, absent -> violates).
	succ := baseSig()
	fams = append(fams, Family{ClusterID: "s", OutcomeClass: domain.OutcomeSuccess,
		Members: []Member{{SignatureID: "msig_s", Signature: succ, OutcomeClass: domain.OutcomeSuccess}}})
	got := EvaluateCandidates([]Proposal{containsResidue()}, fams, 2)
	if got[0].AssociationStatus != AssocDiscriminative {
		t.Fatalf("expected discriminative, got %s", got[0].AssociationStatus)
	}
}

func TestEngineInferredOnlySupportNotLaundered(t *testing.T) {
	sig := baseSig()
	sig.Preserves = []canon.FieldClaim{resolved(domain.FieldPreserves, domain.CanonicalID(idResidue), domain.ClaimInferred)}
	fam := Family{ClusterID: "a", OutcomeClass: domain.OutcomeFailure,
		Members: []Member{{SignatureID: "msig_a", Signature: sig, OutcomeClass: domain.OutcomeFailure}}}
	got := EvaluateCandidates([]Proposal{containsResidue()}, []Family{fam}, 1)
	if got[0].SupportEpistemic.Explicit != 0 || got[0].SupportEpistemic.Inferred != 1 {
		t.Fatalf("inferred-only support must report inferred=1 explicit=0, got %+v", got[0].SupportEpistemic)
	}
}

func TestEngineMixedFamilySplitsMemberWise(t *testing.T) {
	failSig := baseSig()
	failSig.Preserves = []canon.FieldClaim{resolved(domain.FieldPreserves, domain.CanonicalID(idResidue), domain.ClaimExplicit)}
	// A partial-success member that also preserves residue (so it would satisfy):
	psSig := baseSig()
	psSig.Preserves = []canon.FieldClaim{resolved(domain.FieldPreserves, domain.CanonicalID(idResidue), domain.ClaimExplicit)}
	fam := Family{
		ClusterID: "m", OutcomeMixed: true, OutcomeClass: domain.OutcomeMixed,
		Members: []Member{
			{SignatureID: "msig_f", Signature: failSig, OutcomeClass: domain.OutcomeFailure},
			{SignatureID: "msig_p", Signature: psSig, OutcomeClass: domain.OutcomePartialSuccess},
		},
	}
	got := EvaluateCandidates([]Proposal{containsResidue()}, []Family{fam}, 1)
	c := got[0]
	// The mixed family must appear on BOTH support and contrast sides, labeled mixed.
	var sawSupport, sawContrast bool
	for _, fe := range c.FamilyEvaluations {
		if fe.OutcomeClass != "mixed" {
			t.Fatalf("mixed family evaluation must be labeled mixed, got %q", fe.OutcomeClass)
		}
		if fe.Role == RoleSupport {
			sawSupport = true
		}
		if fe.Role == RoleContrast {
			sawContrast = true
		}
	}
	if !sawSupport || !sawContrast {
		t.Fatalf("mixed family must split member-wise onto both sides (support=%v contrast=%v)", sawSupport, sawContrast)
	}
}

func TestEngineAmbiguousDecisiveReadIsUnknownNotSupport(t *testing.T) {
	amb := baseSig()
	amb.Preserves = []canon.FieldClaim{{FieldKind: domain.FieldPreserves, State: domain.ResolutionAmbiguous, Status: domain.ClaimAmbiguous}}
	fam := Family{ClusterID: "a", OutcomeClass: domain.OutcomeFailure,
		Members: []Member{{SignatureID: "msig_a", Signature: amb, OutcomeClass: domain.OutcomeFailure}}}
	got := EvaluateCandidates([]Proposal{containsResidue()}, []Family{fam}, 1)
	if got[0].DistinctFamilySupport != 0 {
		t.Fatalf("ambiguous read must not count as support, got %d", got[0].DistinctFamilySupport)
	}
	if got[0].AssociationStatus != AssocUnknown {
		t.Fatalf("expected unknown, got %s", got[0].AssociationStatus)
	}
}

func TestEngineObstructionIsModelHypothesisNotCodeVerdict(t *testing.T) {
	fams := []Family{failFamily("a", idResidue), failFamily("b", idResidue)}
	prop := containsResidue()
	prop.ObstructionHypothesis = true
	got := EvaluateCandidates([]Proposal{prop}, fams, 2)
	if !got[0].ObstructionIsHypothesis {
		t.Fatal("obstruction hypothesis flag must be recorded")
	}
	if got[0].AssociationStatus == AssocCandidateObstruction {
		t.Fatal("code must never assign candidate_obstruction as association_status")
	}
	if got[0].AssociationStatus != AssocRecurring {
		t.Fatalf("association_status must reflect measurement (recurring), got %s", got[0].AssociationStatus)
	}
}
