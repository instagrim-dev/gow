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
