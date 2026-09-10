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
		// Synthetic test signatures are fully specified by their author, so their
		// set fields are `complete`: an absent value is a verified negative. F3
		// tests that need an unrecorded field override this explicitly.
		SetFieldCompleteness: completeAllSetFields(),
	}
}

// completeAllSetFields marks every set-valued field complete, the default for
// fully-specified synthetic test signatures.
func completeAllSetFields() map[domain.FieldKind]domain.FieldCompleteness {
	return map[domain.FieldKind]domain.FieldCompleteness{
		domain.FieldRepresentation:  domain.CompletenessComplete,
		domain.FieldOperator:        domain.CompletenessComplete,
		domain.FieldAssumption:      domain.CompletenessComplete,
		domain.FieldPreserves:       domain.CompletenessComplete,
		domain.FieldBreaks:          domain.CompletenessComplete,
		domain.FieldAuxiliaryObject: domain.CompletenessComplete,
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

// A redundant paraphrase WITHIN a family does not add a support unit (the
// distinct-family cap holds) and, critically, does not remove the family's
// existing support unit or hide a counterexample. This is the F1 contract:
// deduplicate the count, not the evidence.
func TestEngineRedundantMemberDoesNotAddOrRemoveSupport(t *testing.T) {
	two := []Family{failFamily("a", idResidue), failFamily("b", idResidue)}
	base := EvaluateCandidates([]Proposal{containsResidue()}, two, 2)
	if base[0].DistinctFamilySupport != 2 {
		t.Fatalf("two distinct families must give support 2, got %d", base[0].DistinctFamilySupport)
	}

	// Add a paraphrase (a second member reading the same id) inside family "a".
	// Support must stay 2: the paraphrase neither inflates (cap at one per
	// family) nor erases family "a"'s existing support unit.
	paraphrased := []Family{failFamily("a", idResidue), failFamily("b", idResidue)}
	dupSig := baseSig()
	dupSig.Preserves = append(dupSig.Preserves, resolved(domain.FieldPreserves, domain.CanonicalID(idResidue), domain.ClaimExplicit))
	paraphrased[0].Members = append(paraphrased[0].Members, Member{
		SignatureID: "msig_a_paraphrase", Signature: dupSig, OutcomeClass: domain.OutcomeFailure,
	})
	paraphrased[0].Redundant = true // the family-level hint is set; it must not gate support
	got := EvaluateCandidates([]Proposal{containsResidue()}, paraphrased, 2)
	if got[0].DistinctFamilySupport != 2 {
		t.Fatalf("adding a paraphrase must preserve support 2 (not add, not remove), got %d", got[0].DistinctFamilySupport)
	}
	if got[0].AssociationStatus != AssocRecurring {
		t.Fatalf("support 2 >= minSupport 2 must be recurring, got %s", got[0].AssociationStatus)
	}
	if len(got[0].Counterexamples) != 0 {
		t.Fatalf("a supporting paraphrase must not introduce a counterexample, got %d", len(got[0].Counterexamples))
	}
}

// A redundant member that VIOLATES the predicate must still surface as a
// counterexample: a redundant family flag cannot suppress evaluation of its
// members (the pre-F1 representative-only fallback could have hidden this).
func TestEngineRedundantFamilyStillYieldsCounterexample(t *testing.T) {
	fams := []Family{failFamily("a", idResidue), failFamily("b", idResidue)}
	// Family "b" is flagged redundant but its member does NOT preserve residue
	// (it reads sieve instead): resolved-absent -> violates -> counterexample.
	bSig := baseSig()
	bSig.Preserves = []canon.FieldClaim{resolved(domain.FieldPreserves, domain.CanonicalID(idSieve), domain.ClaimExplicit)}
	fams[1].Members = []Member{{SignatureID: "msig_b", Signature: bSig, OutcomeClass: domain.OutcomeFailure}}
	fams[1].Redundant = true
	got := EvaluateCandidates([]Proposal{containsResidue()}, fams, 2)
	if got[0].DistinctFamilySupport != 1 {
		t.Fatalf("only family a supports, got %d", got[0].DistinctFamilySupport)
	}
	if len(got[0].Counterexamples) != 1 || got[0].Counterexamples[0].ClusterID != "b" {
		t.Fatalf("redundant family b must still surface as a counterexample, got %+v", got[0].Counterexamples)
	}
}

func TestEngineDiscriminativeWhenAbsentInSuccess(t *testing.T) {
	fams := []Family{failFamily("a", idResidue), failFamily("b", idResidue)}
	// A success family that does NOT preserve residue (resolved, absent -> violates).
	succ := baseSig()
	fams = append(fams, Family{ClusterID: "s", OutcomeClass: domain.OutcomeSuccess,
		Members: []Member{{SignatureID: "msig_s", Signature: succ, OutcomeClass: domain.OutcomeSuccess}}})
	got := EvaluateCandidates([]Proposal{containsResidue()}, fams, 2)
	if got[0].AssociationStatus != AssocContrastObserved {
		t.Fatalf("expected contrast_observed, got %s", got[0].AssociationStatus)
	}
	// Complete contrast counts are recorded (F2): one eligible success family,
	// one violating (residue absent where the failures preserve it).
	if got[0].ContrastEligibleDen != 1 || got[0].ContrastViolatingNum != 1 {
		t.Fatalf("expected contrast 1/1, got %d/%d", got[0].ContrastViolatingNum, got[0].ContrastEligibleDen)
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
