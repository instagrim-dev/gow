package frontier

import (
	"testing"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/invariant"
)

const (
	residueLocality = "domain.number_theory.property.residue_locality"
	globalCoupling  = "core.property.global_coupling"
)

// sig builds a minimal resolved signature preserving the given canonical ids
// with the given posture locality.
func sig(locality domain.Locality, preserves ...string) canon.MechanismSignature {
	claims := make([]canon.FieldClaim, 0, len(preserves))
	for _, id := range preserves {
		claims = append(claims, canon.FieldClaim{
			FieldKind:    domain.FieldPreserves,
			SurfaceLabel: id,
			State:        domain.ResolutionResolved,
			CanonicalID:  domain.CanonicalID(id),
			Status:       domain.ClaimExplicit,
		})
	}
	return canon.MechanismSignature{
		SchemaVersion:     canon.SchemaMechanismV1,
		VocabularyVersion: canon.SchemaMechanismV1,
		Preserves:         claims,
		Representations:   []canon.FieldClaim{},
		Operators:         []canon.FieldClaim{},
		Assumptions:       []canon.FieldClaim{},
		Breaks:            []canon.FieldClaim{},
		AuxiliaryObjects:  []canon.FieldClaim{},
		Posture:           canon.Posture{Locality: locality, Construction: domain.ConstructionConstructive, Uncertainty: domain.UncertaintyDeterministic},
		OutcomeClass:      domain.OutcomePartialFailure,
		// preserves is exhaustively extracted in fixtures, so an absent id is a
		// verified negative (F3) rather than an epistemic gap.
		SetFieldCompleteness: map[domain.FieldKind]domain.FieldCompleteness{
			domain.FieldPreserves: domain.CompletenessComplete,
		},
	}
}

func preservesPredicate(id string) invariant.Predicate {
	return invariant.Predicate{
		Schema: invariant.PredicateSchemaV1,
		Root:   invariant.Node{Op: invariant.OpContains, Field: invariant.FieldPreserves, CanonicalID: id},
	}
}

func TestEvaluateProposalsViolationVerified(t *testing.T) {
	// Surviving invariant: failed methods preserve residue locality.
	target := SurvivingInvariant{
		InvariantID:          "inv_1",
		PredicateFingerprint: "fp",
		Predicate:            preservesPredicate(residueLocality),
	}
	targets := map[string]SurvivingInvariant{"inv_1": target}

	// Family that preserves residue locality (the known failure structure).
	families := []Family{{ClusterID: "mcl_a", OutcomeClass: domain.OutcomePartialFailure, Representative: sig(domain.LocalityLocal, residueLocality)}}

	// A proposal that BREAKS residue locality: its signature preserves global
	// coupling instead (resolved, so the preserves field is NOT ambiguous).
	breaking := Proposal{
		ProposedSignature:         sig(domain.LocalityGlobal, globalCoupling),
		TargetInvariantIDs:        []string{"inv_1"},
		StructuralViolationClaim:  "replaces residue-local reasoning with a global coupling object",
		CheapestFalsificationPath: "check whether the global object collapses to a residue cover",
		NoveltyArgument:           "introduces a global auxiliary object absent from every failure family",
		ExpectedInformationGain:   domain.OrdinalHigh,
		EvaluationCost:            domain.OrdinalLow,
	}

	got := EvaluateProposals([]Proposal{breaking}, families, targets, canon.ProfileMechanismV1())
	if len(got) != 1 {
		t.Fatalf("want 1 candidate, got %d", len(got))
	}
	c := got[0]
	if !c.ViolatesAnyTarget {
		t.Fatalf("expected the proposal to violate the target invariant; checks=%+v", c.ViolationChecks)
	}
	if c.ViolationChecks[0].Verdict != invariant.VerdictViolates {
		t.Fatalf("verdict = %q, want violates", c.ViolationChecks[0].Verdict)
	}
	// Global-coupling proposal is mechanistically distant from the residue-local family.
	if c.MechanisticDistance != domain.OrdinalHigh {
		t.Fatalf("mechanistic distance = %q, want high", c.MechanisticDistance)
	}
}

func TestEvaluateProposalsClaimRefutedWhenPreserved(t *testing.T) {
	// A proposal that CLAIMS to break residue locality but whose signature STILL
	// preserves it: code refutes the claim (satisfies => not violated).
	target := SurvivingInvariant{InvariantID: "inv_1", Predicate: preservesPredicate(residueLocality)}
	targets := map[string]SurvivingInvariant{"inv_1": target}
	families := []Family{{ClusterID: "mcl_a", Representative: sig(domain.LocalityLocal, residueLocality)}}

	confidentButNear := Proposal{
		ProposedSignature:         sig(domain.LocalityLocal, residueLocality),
		TargetInvariantIDs:        []string{"inv_1"},
		StructuralViolationClaim:  "totally different, trust me",
		CheapestFalsificationPath: "n/a",
	}
	got := EvaluateProposals([]Proposal{confidentButNear}, families, targets, canon.ProfileMechanismV1())
	if got[0].ViolatesAnyTarget {
		t.Fatal("code should refute a confident claim when the signature still preserves the invariant")
	}
	if got[0].ViolationChecks[0].Verdict != invariant.VerdictSatisfies {
		t.Fatalf("verdict = %q, want satisfies", got[0].ViolationChecks[0].Verdict)
	}
	// Near an existing family => low mechanistic distance, not trusted as novel.
	if got[0].MechanisticDistance != domain.OrdinalLow {
		t.Fatalf("distance = %q, want low (proposal resembles a known family)", got[0].MechanisticDistance)
	}
}

func TestEvaluateProposalsUnknownWhenAmbiguous(t *testing.T) {
	target := SurvivingInvariant{InvariantID: "inv_1", Predicate: preservesPredicate(residueLocality)}
	targets := map[string]SurvivingInvariant{"inv_1": target}
	// Proposed signature leaves preserves unresolved => the read axis is ambiguous.
	ambiguous := canon.MechanismSignature{
		SchemaVersion: canon.SchemaMechanismV1, VocabularyVersion: canon.SchemaMechanismV1,
		Preserves:       []canon.FieldClaim{{FieldKind: domain.FieldPreserves, SurfaceLabel: "vague", State: domain.ResolutionUnknown, Status: domain.ClaimInferred}},
		Representations: []canon.FieldClaim{}, Operators: []canon.FieldClaim{}, Assumptions: []canon.FieldClaim{},
		Breaks: []canon.FieldClaim{}, AuxiliaryObjects: []canon.FieldClaim{},
		Posture:      canon.Posture{Locality: domain.LocalityLocal, Construction: domain.ConstructionConstructive, Uncertainty: domain.UncertaintyDeterministic},
		OutcomeClass: domain.OutcomePartialFailure,
	}
	p := Proposal{ProposedSignature: ambiguous, TargetInvariantIDs: []string{"inv_1"}, StructuralViolationClaim: "x", CheapestFalsificationPath: "y"}
	got := EvaluateProposals([]Proposal{p}, nil, targets, canon.ProfileMechanismV1())
	if got[0].ViolatesAnyTarget {
		t.Fatal("ambiguous read must not count as a confirmed violation")
	}
	if got[0].ViolationChecks[0].Verdict != invariant.VerdictUnknown {
		t.Fatalf("verdict = %q, want unknown", got[0].ViolationChecks[0].Verdict)
	}
}

func TestProposalHashDedupAndStability(t *testing.T) {
	p := Proposal{
		ProposedSignature:        sig(domain.LocalityGlobal, globalCoupling),
		TargetInvariantIDs:       []string{"inv_1"},
		StructuralViolationClaim: "Breaks Residue Locality",
	}
	// Cosmetic-only variant: reordered/duplicated targets, whitespace+case in claim.
	pDup := Proposal{
		ProposedSignature:        sig(domain.LocalityGlobal, globalCoupling),
		TargetInvariantIDs:       []string{"inv_1", "inv_1"},
		StructuralViolationClaim: "  breaks residue locality  ",
	}
	if ProposalHash(p) != ProposalHash(pDup) {
		t.Fatal("cosmetic-only variants must share a proposal hash")
	}
	got := EvaluateProposals([]Proposal{p, pDup}, nil, map[string]SurvivingInvariant{}, canon.ProfileMechanismV1())
	if len(got) != 1 {
		t.Fatalf("duplicate proposals must collapse; got %d", len(got))
	}
}

func TestRankObjectiveOrder(t *testing.T) {
	// c1: confirmed violation, distant, high EIG, low cost -> should rank first.
	// c2: no violation, distant.
	// c3: no violation, near (low distance).
	c1 := Candidate{ProposalHash: "a", ViolatesAnyTarget: true, MechanisticDistance: domain.OrdinalHigh, ExpectedInformationGain: domain.OrdinalHigh, EvaluationCost: domain.OrdinalLow}
	c2 := Candidate{ProposalHash: "b", ViolatesAnyTarget: false, MechanisticDistance: domain.OrdinalHigh, ExpectedInformationGain: domain.OrdinalMedium, EvaluationCost: domain.OrdinalMedium}
	c3 := Candidate{ProposalHash: "c", ViolatesAnyTarget: false, MechanisticDistance: domain.OrdinalLow, ExpectedInformationGain: domain.OrdinalHigh, EvaluationCost: domain.OrdinalLow}
	ranked := Rank([]Candidate{c3, c2, c1})
	if ranked[0].ProposalHash != "a" || ranked[1].ProposalHash != "b" || ranked[2].ProposalHash != "c" {
		t.Fatalf("rank order = %s,%s,%s; want a,b,c", ranked[0].ProposalHash, ranked[1].ProposalHash, ranked[2].ProposalHash)
	}
}

func TestRankDownRanksRedundant(t *testing.T) {
	// Two proposals with identical target+nearest-family signature: the second is
	// down-ranked despite identical component ordinals.
	base := Candidate{ViolatesAnyTarget: true, MechanisticDistance: domain.OrdinalHigh, ExpectedInformationGain: domain.OrdinalHigh, EvaluationCost: domain.OrdinalLow, TargetInvariantIDs: []string{"inv_1"}, NearestClusters: []NearestCluster{{ClusterID: "mcl_a"}}}
	first := base
	first.ProposalHash = "aaa"
	dupAttack := base
	dupAttack.ProposalHash = "bbb"
	ranked := Rank([]Candidate{dupAttack, first})
	if ranked[0].ProposalHash != "aaa" {
		t.Fatalf("first-seen directed attack should rank first; got %s", ranked[0].ProposalHash)
	}
}
