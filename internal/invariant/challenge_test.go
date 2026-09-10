package invariant

import (
	"strings"
	"testing"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/domain"
)

const (
	chIDResidue = "domain.number_theory.property.residue_locality"
	chIDSieve   = "core.operator.sieve"
	chIDGlobal  = "domain.number_theory.property.global_density"
	chIDBounds  = "core.assumption.effective_bounds"
)

func chPredicate(id string) Predicate {
	return Predicate{Schema: PredicateSchemaV1, Root: Node{Op: OpContains, Field: FieldPreserves, CanonicalID: id}}
}

// chSignature builds a resolved signature whose preserves set carries the given
// canonical ids, with the given outcome. The preserves field is marked
// exhaustively extracted so absence is a verified negative (violates), matching
// what the F3 completeness semantics require for counterexample confirmation.
func chSignature(sigID string, outcome domain.OutcomeClass, preserves ...string) canon.MechanismSignature {
	sig := canon.MechanismSignature{
		SchemaVersion:     canon.SchemaMechanismV1,
		VocabularyVersion: "mechanism/v1",
		SignatureID:       sigID,
		OutcomeClass:      outcome,
		Preserves:         []canon.FieldClaim{},
		SetFieldCompleteness: map[domain.FieldKind]domain.FieldCompleteness{
			domain.FieldPreserves: domain.CompletenessComplete,
		},
	}
	for _, id := range preserves {
		sig.Preserves = append(sig.Preserves, canon.FieldClaim{
			FieldKind:   domain.FieldPreserves,
			State:       domain.ResolutionResolved,
			CanonicalID: domain.CanonicalID(id),
			Status:      domain.ClaimExplicit,
		})
	}
	return sig
}

func chFamily(clusterID string, outcome domain.OutcomeClass, members ...canon.MechanismSignature) Family {
	fam := Family{ClusterID: clusterID, OutcomeClass: outcome}
	for _, m := range members {
		fam.Members = append(fam.Members, Member{SignatureID: m.SignatureID, Signature: m, OutcomeClass: m.OutcomeClass})
	}
	return fam
}

// Atlas: two failure families preserving residue locality, one failure family
// that does NOT (it preserves sieve + effective bounds), one partial-success
// family preserving residue + effective bounds.
func chAtlas() []Family {
	return []Family{
		chFamily("mcl_f1", domain.OutcomeFailure, chSignature("msig_f1", domain.OutcomeFailure, chIDResidue)),
		chFamily("mcl_f2", domain.OutcomeFailure, chSignature("msig_f2", domain.OutcomeFailure, chIDResidue, chIDGlobal)),
		chFamily("mcl_f3", domain.OutcomePartialFailure, chSignature("msig_f3", domain.OutcomePartialFailure, chIDSieve, chIDBounds)),
		chFamily("mcl_s1", domain.OutcomePartialSuccess, chSignature("msig_s1", domain.OutcomePartialSuccess, chIDResidue, chIDBounds)),
	}
}

func TestVerifyKnownCounterexample(t *testing.T) {
	families := chAtlas()
	// mcl_f3 violates preserves(residue) -> confirmed, evidence names it.
	res := VerifyKnownCounterexample(chPredicate(chIDResidue), families)
	if !res.Confirmed {
		t.Fatalf("expected confirmed, got %+v", res)
	}
	if len(res.Evidence) != 1 || res.Evidence[0].ClusterID != "mcl_f3" || res.Evidence[0].SignatureID != "msig_f3" {
		t.Fatalf("evidence must name the violating member: %+v", res.Evidence)
	}
	// A predicate every failure family satisfies has no known counterexample.
	all := Predicate{Schema: PredicateSchemaV1, Root: Node{Op: OpAny, Children: []Node{
		chPredicate(chIDResidue).Root, chPredicate(chIDSieve).Root,
	}}}
	if res := VerifyKnownCounterexample(all, families); res.Confirmed {
		t.Fatalf("expected unconfirmed, got %+v", res)
	}
}

func TestVerifyKnownCounterexampleUnknownIsNotACounterexample(t *testing.T) {
	sig := chSignature("msig_amb", domain.OutcomeFailure)
	sig.Preserves = append(sig.Preserves, canon.FieldClaim{
		FieldKind: domain.FieldPreserves, State: domain.ResolutionAmbiguous, Status: domain.ClaimAmbiguous,
	})
	families := []Family{chFamily("mcl_amb", domain.OutcomeFailure, sig)}
	if res := VerifyKnownCounterexample(chPredicate(chIDResidue), families); res.Confirmed {
		t.Fatalf("ambiguous member must not be a counterexample: %+v", res)
	}
}

func TestVerifySyntheticCounterexample(t *testing.T) {
	violating := chSignature("", domain.OutcomeFailure, chIDSieve)
	if res := VerifySyntheticCounterexample(chPredicate(chIDResidue), violating); !res.Confirmed {
		t.Fatalf("expected confirmed, got %+v", res)
	}
	satisfying := chSignature("", domain.OutcomeFailure, chIDResidue)
	if res := VerifySyntheticCounterexample(chPredicate(chIDResidue), satisfying); res.Confirmed {
		t.Fatalf("expected unconfirmed, got %+v", res)
	}
}

func TestVerifySuccessPreserving(t *testing.T) {
	families := chAtlas()
	// mcl_s1 (partial_success) preserves residue -> confirmed.
	res := VerifySuccessPreserving(chPredicate(chIDResidue), families)
	if !res.Confirmed || len(res.Evidence) != 1 || res.Evidence[0].ClusterID != "mcl_s1" {
		t.Fatalf("expected confirmed via mcl_s1, got %+v", res)
	}
	// No success family preserves sieve -> unconfirmed.
	if res := VerifySuccessPreserving(chPredicate(chIDSieve), families); res.Confirmed {
		t.Fatalf("expected unconfirmed, got %+v", res)
	}
}

func TestVerifyBiasCritique(t *testing.T) {
	families := chAtlas()
	// residue is supported by 2 distinct failure families; threshold 3 -> support
	// collapses below -> confirmed.
	res := VerifyBiasCritique(chPredicate(chIDResidue), families, 3)
	if !res.Confirmed {
		t.Fatalf("expected confirmed at threshold 3, got %+v", res)
	}
	if len(res.Evidence) != 1 || res.Evidence[0].Kind != EvidenceSupportRecount {
		t.Fatalf("recount must be the evidence: %+v", res.Evidence)
	}
	// Threshold 2 -> genuinely sufficient support -> unconfirmed (but the recount
	// is still recorded for audit).
	res = VerifyBiasCritique(chPredicate(chIDResidue), families, 2)
	if res.Confirmed {
		t.Fatalf("expected unconfirmed at threshold 2, got %+v", res)
	}
	if len(res.Evidence) != 1 {
		t.Fatalf("unconfirmed bias critique must still record the recount: %+v", res.Evidence)
	}
}

func TestVerifySplitGroundedDisjoint(t *testing.T) {
	// Parent: any(residue, sieve). Children residue / sieve ground to disjoint
	// failure families {f1,f2} / {f3}.
	parent := Predicate{Schema: PredicateSchemaV1, Root: Node{Op: OpAny, Children: []Node{
		chPredicate(chIDResidue).Root, chPredicate(chIDSieve).Root,
	}}}
	res := VerifySplit(parent, []Predicate{chPredicate(chIDResidue), chPredicate(chIDSieve)}, chAtlas())
	if !res.Confirmed {
		t.Fatalf("expected confirmed split, got %+v", res)
	}
	if len(res.Evidence) != 3 {
		t.Fatalf("expected 3 grounding rows (f1,f2 / f3), got %+v", res.Evidence)
	}
}

func TestVerifySplitRejectsOverlapAndUngrounded(t *testing.T) {
	// Overlapping children: residue {f1,f2} and global {f2} share supporter mcl_f2.
	overlapParent := Predicate{Schema: PredicateSchemaV1, Root: Node{Op: OpAny, Children: []Node{
		chPredicate(chIDResidue).Root, chPredicate(chIDGlobal).Root,
	}}}
	res := VerifySplit(overlapParent, []Predicate{chPredicate(chIDResidue), chPredicate(chIDGlobal)}, chAtlas())
	if res.Confirmed || !strings.Contains(res.Detail, "not disjoint") {
		t.Fatalf("expected overlap rejection, got %+v", res)
	}
	parent := chPredicate(chIDResidue)
	// Ungrounded child: nothing preserves an unused id.
	unused := chPredicate("core.operator.unused_everywhere")
	res = VerifySplit(parent, []Predicate{chPredicate(chIDSieve), unused}, chAtlas())
	if res.Confirmed || !strings.Contains(res.Detail, "grounds to no supporting") {
		t.Fatalf("expected ungrounded rejection, got %+v", res)
	}
	// A child identical to the parent is not a split.
	res = VerifySplit(parent, []Predicate{parent, chPredicate(chIDSieve)}, chAtlas())
	if res.Confirmed || !strings.Contains(res.Detail, "identical to the parent") {
		t.Fatalf("expected identity rejection, got %+v", res)
	}
}

func TestVerifyMergePreservesDiscrimination(t *testing.T) {
	families := chAtlas()
	// sieve discriminates (violated in s1); merging it into any(residue, sieve)
	// which s1 SATISFIES erases that signal -> rejected.
	parents := []Predicate{chPredicate(chIDResidue), chPredicate(chIDSieve)}
	child := Predicate{Schema: PredicateSchemaV1, Root: Node{Op: OpAny, Children: []Node{
		chPredicate(chIDResidue).Root, chPredicate(chIDSieve).Root,
	}}}
	res := VerifyMerge(parents, child, families)
	if res.Confirmed || !strings.Contains(res.Detail, "loses discrimination") {
		t.Fatalf("expected discrimination-loss rejection, got %+v", res)
	}
	// residue and effective-bounds both have zero contrast violations (s1
	// preserves both); their union covers {f1,f2,f3} and keeps contrast 0 -> a
	// merge that loses nothing -> confirmed.
	parents2 := []Predicate{chPredicate(chIDResidue), chPredicate(chIDBounds)}
	child2 := Predicate{Schema: PredicateSchemaV1, Root: Node{Op: OpAny, Children: []Node{
		chPredicate(chIDResidue).Root, chPredicate(chIDBounds).Root,
	}}}
	res = VerifyMerge(parents2, child2, families)
	if !res.Confirmed {
		t.Fatalf("expected confirmed merge, got %+v", res)
	}
	if len(res.Evidence) != 3 {
		t.Fatalf("expected grounding to f1,f2,f3, got %+v", res.Evidence)
	}
}

func TestVerifyMergeRejectsCoverageGap(t *testing.T) {
	families := chAtlas()
	// Parents residue {f1,f2} and bounds {f3} both have zero contrast, so the
	// discrimination gate passes; child = residue alone cannot cover bounds'
	// supporter mcl_f3 -> coverage rejection.
	res := VerifyMerge([]Predicate{chPredicate(chIDResidue), chPredicate(chIDBounds)}, chPredicate(chIDResidue), families)
	if res.Confirmed || !strings.Contains(res.Detail, "does not cover") {
		t.Fatalf("expected coverage rejection, got %+v", res)
	}
}
