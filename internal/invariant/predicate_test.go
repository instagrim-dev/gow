package invariant

import (
	"errors"
	"strings"
	"testing"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/domain"
)

const (
	predIDResidue = "domain.number_theory.property.residue_locality"
	predIDSieve   = "core.operator.sieve"
	predIDBound   = "core.boundary.large_prime_case"
)

func predContains(field, id string) Node {
	return Node{Op: OpContains, Field: field, CanonicalID: id}
}

func predBaseSignature() canon.MechanismSignature {
	return canon.MechanismSignature{
		SchemaVersion:     canon.SchemaMechanismV1,
		VocabularyVersion: "mechanism/v1",
		Preserves: []canon.FieldClaim{{
			FieldKind:   domain.FieldPreserves,
			State:       domain.ResolutionResolved,
			CanonicalID: domain.CanonicalID(predIDResidue),
			Status:      domain.ClaimExplicit,
		}},
		Posture: canon.Posture{
			Locality:     domain.LocalityLocal,
			Construction: domain.ConstructionConstructive,
			Uncertainty:  domain.UncertaintyDeterministic,
		},
		OutcomeClass: domain.OutcomeFailure,
		Boundaries: []canon.Boundary{{
			State:       domain.ResolutionResolved,
			CanonicalID: domain.CanonicalID(predIDBound),
			Relation:    "stops_at",
			Status:      domain.ClaimExplicit,
		}},
		// Fully-specified synthetic signature: set fields are complete, so a
		// resolved-but-absent value is a verified negative (F3).
		SetFieldCompleteness: map[domain.FieldKind]domain.FieldCompleteness{
			domain.FieldRepresentation:  domain.CompletenessComplete,
			domain.FieldOperator:        domain.CompletenessComplete,
			domain.FieldAssumption:      domain.CompletenessComplete,
			domain.FieldPreserves:       domain.CompletenessComplete,
			domain.FieldBreaks:          domain.CompletenessComplete,
			domain.FieldAuxiliaryObject: domain.CompletenessComplete,
		},
	}
}

// KTD-5 normalization pass: single-child all/any flattens to the child and
// not(not(x)) folds to x, so degenerate nesting cannot mint a new identity.
func TestFingerprintNormalizesDegenerateNesting(t *testing.T) {
	plain := Predicate{Schema: PredicateSchemaV1, Root: predContains(FieldPreserves, predIDResidue)}
	wrapped := Predicate{Schema: PredicateSchemaV1, Root: Node{Op: OpAll, Children: []Node{predContains(FieldPreserves, predIDResidue)}}}
	doubleNeg := Predicate{Schema: PredicateSchemaV1, Root: Node{Op: OpNot, Children: []Node{{Op: OpNot, Children: []Node{predContains(FieldPreserves, predIDResidue)}}}}}
	if Fingerprint(plain) != Fingerprint(wrapped) {
		t.Fatal("single-child all should flatten to the child")
	}
	if Fingerprint(plain) != Fingerprint(doubleNeg) {
		t.Fatal("not(not(x)) should fold to x")
	}
}

// An enum axis recorded as its unknown sentinel is an epistemic gap: a
// predicate reading it yields unknown, never a coerced violates.
func TestEvaluateUnknownEnumAxisIsUnknown(t *testing.T) {
	sig := predBaseSignature()
	sig.Posture.Locality = domain.LocalityUnknown
	p := Predicate{Schema: PredicateSchemaV1, Root: Node{Op: OpIn, Field: FieldLocality, Values: []string{"local"}}}
	if got := Evaluate(p, sig); got != VerdictUnknown {
		t.Fatalf("Evaluate(unknown axis) = %s, want unknown", got)
	}
}

// A predicate may not target an axis's unknown sentinel as a query value:
// "these failures all share locality=unknown" is a claim about missing data,
// not a conserved mechanistic property, so Validate rejects it rather than
// storing a predicate that Evaluate can never satisfy. Guards the F1 seam
// where enumFieldValues once admitted a value enumAxisValue always treats as
// known=false (validatable-but-permanently-unsatisfiable).
func TestValidateRejectsUnknownSentinelEnumQuery(t *testing.T) {
	cases := []struct {
		name  string
		field string
		value string
	}{
		{"locality", FieldLocality, string(domain.LocalityUnknown)},
		{"construction", FieldConstruction, string(domain.ConstructionUnknown)},
		{"uncertainty", FieldUncertainty, string(domain.UncertaintyUnknown)},
		{"outcome", FieldOutcome, string(domain.OutcomeUnknown)},
	}
	for _, tc := range cases {
		t.Run("equals_"+tc.name, func(t *testing.T) {
			p := Predicate{Schema: PredicateSchemaV1, Root: Node{Op: OpEquals, Field: tc.field, Values: []string{tc.value}}}
			if err := Validate(p); !errors.Is(err, ErrInvalidPredicate) {
				t.Fatalf("Validate(equals %s=%s) err = %v, want ErrInvalidPredicate", tc.field, tc.value, err)
			}
		})
		t.Run("in_"+tc.name, func(t *testing.T) {
			p := Predicate{Schema: PredicateSchemaV1, Root: Node{Op: OpIn, Field: tc.field, Values: []string{tc.value}}}
			if err := Validate(p); !errors.Is(err, ErrInvalidPredicate) {
				t.Fatalf("Validate(in %s=[%s]) err = %v, want ErrInvalidPredicate", tc.field, tc.value, err)
			}
		})
	}
}

// A concrete enum value query stays valid (the sentinel exclusion did not
// over-reject legitimate axis values).
func TestValidateAcceptsConcreteEnumQuery(t *testing.T) {
	p := Predicate{Schema: PredicateSchemaV1, Root: Node{Op: OpEquals, Field: FieldOutcome, Values: []string{string(domain.OutcomeFailure)}}}
	if err := Validate(p); err != nil {
		t.Fatalf("Validate(equals outcome=failure) = %v, want nil", err)
	}
}

// F3 essential regression: a missing extraction must not become negative
// evidence. When a set field was NOT exhaustively extracted, an absent value is
// an epistemic gap (unknown), and its negation stays unknown too. Only when the
// field is exhaustively extracted does an absent value become a verified
// negative (violates).
func TestEvaluateAbsentValueRespectsFieldCompleteness(t *testing.T) {
	// Field not extracted (unobserved): the queried id is absent.
	unobserved := canon.MechanismSignature{
		SchemaVersion: canon.SchemaMechanismV1, VocabularyVersion: "mechanism/v1",
		Preserves: []canon.FieldClaim{}, // empty, and completeness unset -> unobserved
	}
	contains := Predicate{Schema: PredicateSchemaV1, Root: predContains(FieldPreserves, predIDResidue)}
	notContains := Predicate{Schema: PredicateSchemaV1, Root: Node{Op: OpNot, Children: []Node{predContains(FieldPreserves, predIDResidue)}}}

	if got := Evaluate(contains, unobserved); got != VerdictUnknown {
		t.Fatalf("contains(X) on unobserved field = %s, want unknown", got)
	}
	if got := Evaluate(notContains, unobserved); got != VerdictUnknown {
		t.Fatalf("not(contains(X)) on unobserved field = %s, want unknown", got)
	}

	// Partial extraction: still not a verified negative.
	partial := unobserved
	partial.SetFieldCompleteness = map[domain.FieldKind]domain.FieldCompleteness{domain.FieldPreserves: domain.CompletenessPartial}
	if got := Evaluate(contains, partial); got != VerdictUnknown {
		t.Fatalf("contains(X) on partial field = %s, want unknown", got)
	}

	// Field exhaustively extracted; X absent: now a verified negative.
	complete := unobserved
	complete.SetFieldCompleteness = map[domain.FieldKind]domain.FieldCompleteness{domain.FieldPreserves: domain.CompletenessComplete}
	if got := Evaluate(contains, complete); got != VerdictViolates {
		t.Fatalf("contains(X) on complete field with X absent = %s, want violates", got)
	}
	if got := Evaluate(notContains, complete); got != VerdictSatisfies {
		t.Fatalf("not(contains(X)) on complete field with X absent = %s, want satisfies", got)
	}
}

// unknown propagates through not: negating an unanswerable read does not make
// it answerable.
func TestEvaluateNotPreservesUnknown(t *testing.T) {
	sig := predBaseSignature()
	sig.Preserves = append(sig.Preserves, canon.FieldClaim{
		FieldKind: domain.FieldPreserves,
		State:     domain.ResolutionAmbiguous,
		Status:    domain.ClaimAmbiguous,
	})
	p := Predicate{Schema: PredicateSchemaV1, Root: Node{Op: OpNot, Children: []Node{predContains(FieldPreserves, predIDSieve)}}}
	if got := Evaluate(p, sig); got != VerdictUnknown {
		t.Fatalf("Evaluate(not unknown) = %s, want unknown", got)
	}
}

// Boundary predicates are relation-decisive: same canonical id under a
// different typed relation violates rather than satisfies.
func TestEvaluateBoundaryRelationDecisive(t *testing.T) {
	sig := predBaseSignature()
	match := Predicate{Schema: PredicateSchemaV1, Root: Node{Op: OpBoundary, CanonicalID: predIDBound, Relation: "stops_at"}}
	wrongRel := Predicate{Schema: PredicateSchemaV1, Root: Node{Op: OpBoundary, CanonicalID: predIDBound, Relation: "requires"}}
	if got := Evaluate(match, sig); got != VerdictSatisfies {
		t.Fatalf("Evaluate(matching boundary) = %s, want satisfies", got)
	}
	if got := Evaluate(wrongRel, sig); got != VerdictViolates {
		t.Fatalf("Evaluate(wrong relation) = %s, want violates", got)
	}
}

func TestMarshalCanonicalRoundTripPreservesFingerprint(t *testing.T) {
	p := Predicate{Schema: PredicateSchemaV1, Root: Node{Op: OpAll, Children: []Node{
		predContains(FieldPreserves, predIDResidue),
		{Op: OpIn, Field: FieldLocality, Values: []string{"mixed", "local"}},
	}}}
	raw, err := MarshalCanonical(p)
	if err != nil {
		t.Fatalf("MarshalCanonical() error: %v", err)
	}
	back, err := ParsePredicate(raw)
	if err != nil {
		t.Fatalf("ParsePredicate(canonical) error: %v", err)
	}
	if Fingerprint(back) != Fingerprint(p) {
		t.Fatal("canonical round-trip changed the fingerprint")
	}
	if !strings.Contains(raw, PredicateSchemaV1) {
		t.Fatal("canonical JSON must carry the schema version")
	}
}

func TestParsePredicateRejectsUnknownJSONFields(t *testing.T) {
	if _, err := ParsePredicate(`{"schema":"invariant-predicate/v1","root":{"op":"contains","field":"preserves","canonical_id":"core.operator.sieve"},"support_families":["mcl_x"]}`); !errors.Is(err, ErrInvalidPredicate) {
		t.Fatalf("ParsePredicate(extra fields) = %v, want ErrInvalidPredicate", err)
	}
}

func TestFieldsReadCollectsDistinctSortedFields(t *testing.T) {
	p := Predicate{Schema: PredicateSchemaV1, Root: Node{Op: OpAll, Children: []Node{
		predContains(FieldPreserves, predIDResidue),
		{Op: OpBoundary, CanonicalID: predIDBound},
		{Op: OpIn, Field: FieldLocality, Values: []string{"local"}},
		{Op: OpNot, Children: []Node{predContains(FieldPreserves, predIDSieve)}},
	}}}
	got := FieldsRead(p)
	want := []string{FieldBoundary, FieldLocality, FieldPreserves}
	if len(got) != len(want) {
		t.Fatalf("FieldsRead() = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("FieldsRead() = %v, want %v", got, want)
		}
	}
}

// ValidateForMining prohibits any read of the outcome axis in a mined
// failure-mechanism predicate, recursively: `outcome in [...]` (and any nesting
// of it) would earn support/contrast by definition, identifying no mechanism.
// The general Validate still accepts these — the restriction is mining-scope (F2).
func TestValidateForMiningProhibitsOutcomeReads(t *testing.T) {
	cases := map[string]Predicate{
		"top-level in": {Schema: PredicateSchemaV1, Root: Node{Op: OpIn, Field: FieldOutcome, Values: []string{"failure", "partial_failure"}}},
		"equals":       {Schema: PredicateSchemaV1, Root: Node{Op: OpEquals, Field: FieldOutcome, Values: []string{"failure"}}},
		"nested under all": {Schema: PredicateSchemaV1, Root: Node{Op: OpAll, Children: []Node{
			predContains(FieldPreserves, predIDResidue),
			{Op: OpIn, Field: FieldOutcome, Values: []string{"failure"}},
		}}},
		"nested under not": {Schema: PredicateSchemaV1, Root: Node{Op: OpNot, Children: []Node{
			{Op: OpEquals, Field: FieldOutcome, Values: []string{"success"}},
		}}},
	}
	for name, p := range cases {
		// The general grammar accepts the outcome read...
		if err := Validate(p); err != nil {
			t.Fatalf("%s: general Validate should accept outcome reads, got %v", name, err)
		}
		// ...but the mining contract rejects it.
		if err := ValidateForMining(p); !errors.Is(err, ErrInvalidPredicate) {
			t.Fatalf("%s: ValidateForMining must reject outcome read, got %v", name, err)
		}
	}
}

// ValidateForMining still accepts a mechanism predicate that reads no outcome.
func TestValidateForMiningAcceptsMechanismPredicate(t *testing.T) {
	p := Predicate{Schema: PredicateSchemaV1, Root: Node{Op: OpAll, Children: []Node{
		predContains(FieldPreserves, predIDResidue),
		{Op: OpIn, Field: FieldLocality, Values: []string{"local", "mixed"}},
	}}}
	if err := ValidateForMining(p); err != nil {
		t.Fatalf("ValidateForMining(mechanism predicate) = %v, want nil", err)
	}
}

// ValidateReferences rejects a syntactically-valid canonical_id that is not in
// the pinned vocabulary, and one whose field kind does not match the queried
// set field, while accepting a genuine term for its own field (F3).
func TestValidateReferencesAgainstVocabulary(t *testing.T) {
	vocab := canon.MechanismV1()
	// Pick a real preserves term and a real operator term from the seed.
	preservesTerms := vocab.Terms(domain.FieldPreserves)
	operatorTerms := vocab.Terms(domain.FieldOperator)
	if len(preservesTerms) == 0 || len(operatorTerms) == 0 {
		t.Skip("seed vocabulary lacks preserves/operator terms")
	}
	realPreserve := string(preservesTerms[0].CanonicalID)
	realOperator := string(operatorTerms[0].CanonicalID)

	// A real preserves term for the preserves field: accepted.
	ok := Predicate{Schema: PredicateSchemaV1, Root: predContains(FieldPreserves, realPreserve)}
	if err := ValidateReferences(ok, vocab); err != nil {
		t.Fatalf("ValidateReferences(real preserves term) = %v, want nil", err)
	}

	// A nonexistent (but syntactically valid) term: rejected.
	ghost := Predicate{Schema: PredicateSchemaV1, Root: predContains(FieldPreserves, "domain.number_theory.property.nonexistent_term")}
	if err := ValidateReferences(ghost, vocab); !errors.Is(err, ErrInvalidPredicate) {
		t.Fatalf("ValidateReferences(nonexistent term) = %v, want ErrInvalidPredicate", err)
	}

	// A real operator term used under contains(preserves): field-kind mismatch, rejected.
	mismatch := Predicate{Schema: PredicateSchemaV1, Root: predContains(FieldPreserves, realOperator)}
	if err := ValidateReferences(mismatch, vocab); !errors.Is(err, ErrInvalidPredicate) {
		t.Fatalf("ValidateReferences(field-kind mismatch) = %v, want ErrInvalidPredicate", err)
	}
}
