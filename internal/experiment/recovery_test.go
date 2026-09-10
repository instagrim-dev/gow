package experiment

import (
	"testing"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/domain"
)

func expSignature(preserves ...string) canon.MechanismSignature {
	sig := canon.MechanismSignature{
		SchemaVersion:     canon.SchemaMechanismV1,
		VocabularyVersion: "mechanism/v1",
		OutcomeClass:      domain.OutcomeUnknown,
		Preserves:         []canon.FieldClaim{},
		Operators:         []canon.FieldClaim{},
		Posture: canon.Posture{
			Locality:     domain.LocalityGlobal,
			Construction: domain.ConstructionConstructive,
			Uncertainty:  domain.UncertaintyDeterministic,
		},
		SetFieldCompleteness: map[domain.FieldKind]domain.FieldCompleteness{
			domain.FieldPreserves: domain.CompletenessComplete,
			domain.FieldOperator:  domain.CompletenessComplete,
		},
	}
	for _, id := range preserves {
		sig.Preserves = append(sig.Preserves, canon.FieldClaim{
			FieldKind: domain.FieldPreserves, State: domain.ResolutionResolved,
			CanonicalID: domain.CanonicalID(id), Status: domain.ClaimExplicit,
		})
	}
	return sig
}

const (
	expIDLattice = "core.auxiliary_object.affine_lattice"
	expIDResidue = "domain.number_theory.property.residue_locality"
)

func TestDetectRecoveryNearAndDistinct(t *testing.T) {
	target := expSignature(expIDLattice)
	profile := canon.ProfileMechanismV1()

	near := ProposalContent{ProposalID: "fpr_near", Rank: 1, Signature: expSignature(expIDLattice)}
	distinct := ProposalContent{ProposalID: "fpr_far", Rank: 0, Signature: expSignature(expIDResidue)}

	out := DetectRecovery([]ProposalContent{near, distinct}, target, profile)
	if !out.Recovered {
		t.Fatalf("expected recovery via the near proposal: %+v", out)
	}
	if out.FirstRecoveryRank != 1 {
		t.Fatalf("first recovery rank = %d, want 1 (the near proposal)", out.FirstRecoveryRank)
	}
	if len(out.Facts) != 2 {
		t.Fatalf("expected 2 facts, got %d", len(out.Facts))
	}
	// Rank order preserved: distinct proposal (rank 0) first, not recovered.
	if out.Facts[0].ProposalID != "fpr_far" || out.Facts[0].Recovered {
		t.Fatalf("rank-0 fact wrong: %+v", out.Facts[0])
	}
}

func TestDetectRecoveryNoRecovery(t *testing.T) {
	target := expSignature(expIDLattice)
	distinct := ProposalContent{ProposalID: "fpr_far", Rank: 0, Signature: expSignature(expIDResidue)}
	out := DetectRecovery([]ProposalContent{distinct}, target, canon.ProfileMechanismV1())
	if out.Recovered || out.FirstRecoveryRank != -1 {
		t.Fatalf("expected no recovery: %+v", out)
	}
}

// An ambiguous proposal read classifies unknown — an epistemic gap, never a
// coerced recovery.
func TestDetectRecoveryAmbiguousIsUnknown(t *testing.T) {
	target := expSignature(expIDLattice)
	amb := expSignature()
	amb.Preserves = append(amb.Preserves, canon.FieldClaim{
		FieldKind: domain.FieldPreserves, State: domain.ResolutionAmbiguous, Status: domain.ClaimAmbiguous,
	})
	out := DetectRecovery([]ProposalContent{{ProposalID: "fpr_amb", Rank: 0, Signature: amb}}, target, canon.ProfileMechanismV1())
	if out.Recovered {
		t.Fatalf("ambiguous content must not recover: %+v", out)
	}
	if out.Facts[0].Classification != canon.ClassUnknown {
		t.Fatalf("classification = %s, want unknown", out.Facts[0].Classification)
	}
}

func TestComputeDiversityDedupsByFingerprint(t *testing.T) {
	a := ProposalContent{ProposalID: "a", Signature: expSignature(expIDLattice)}
	b := ProposalContent{ProposalID: "b", Signature: expSignature(expIDLattice)} // same mechanism
	c := ProposalContent{ProposalID: "c", Signature: expSignature(expIDResidue)}
	d := ComputeDiversity([]ProposalContent{a, b, c})
	if d.DistinctMechanisms != 2 || d.RedundantProposals != 1 {
		t.Fatalf("diversity = %+v, want 2 distinct / 1 redundant", d)
	}
}

func TestMetricOrdinalBands(t *testing.T) {
	cases := []struct {
		num, den int
		want     domain.Ordinal
	}{
		{0, 0, domain.OrdinalUnknown},
		{3, 3, domain.OrdinalHigh},
		{2, 4, domain.OrdinalMedium},
		{1, 4, domain.OrdinalLow},
	}
	for _, c := range cases {
		if got := MetricOrdinal(c.num, c.den); got != c.want {
			t.Fatalf("MetricOrdinal(%d,%d) = %s, want %s", c.num, c.den, got, c.want)
		}
	}
}

// --- AssessProposals: enforced evaluation budget + honest assessment classes ---

// F3 regression: proposal_budget larger than evaluation_budget, with the only
// recovering proposal BEYOND the budget. The enforced budget must stop before
// evaluating it: no recovery may be reported, the unfinished proposal is
// unassessed, and consumption is auditable.
func TestAssessProposalsBudgetStopsBeforeRecovery(t *testing.T) {
	targets := []canon.MechanismSignature{expSignature(expIDLattice)}
	proposals := []ProposalContent{
		{ProposalID: "fpr_far", Rank: 0, Signature: expSignature(expIDResidue)},
		{ProposalID: "fpr_near", Rank: 1, Signature: expSignature(expIDLattice)}, // would recover
	}
	out := AssessProposals(proposals, targets, canon.ProfileMechanismV1(), 1)
	if out.RecoveredCount != 0 {
		t.Fatalf("budget=1 must stop before the recovering proposal: %+v", out)
	}
	if out.EvaluationsConsumed != 1 {
		t.Fatalf("consumed = %d, want 1", out.EvaluationsConsumed)
	}
	if out.UnassessedCount != 1 {
		t.Fatalf("unassessed = %d, want 1 (the never-compared proposal)", out.UnassessedCount)
	}
	if out.Proposals[1].Assessment != AssessmentUnassessed {
		t.Fatalf("rank-1 assessment = %s, want unassessed", out.Proposals[1].Assessment)
	}
}

// With sufficient budget the same population recovers, consumption reflects the
// early-exit comparison count, and the decisive non-recovery is counted.
func TestAssessProposalsSufficientBudgetRecovers(t *testing.T) {
	targets := []canon.MechanismSignature{expSignature(expIDLattice)}
	proposals := []ProposalContent{
		{ProposalID: "fpr_far", Rank: 0, Signature: expSignature(expIDResidue)},
		{ProposalID: "fpr_near", Rank: 1, Signature: expSignature(expIDLattice)},
	}
	out := AssessProposals(proposals, targets, canon.ProfileMechanismV1(), 0) // unlimited
	if out.RecoveredCount != 1 || out.FirstRecoveryRank != 1 {
		t.Fatalf("expected recovery at rank 1: %+v", out)
	}
	if out.DecisiveNoCount != 1 {
		t.Fatalf("decisive-no = %d, want 1", out.DecisiveNoCount)
	}
	if out.EvaluationsConsumed != 2 {
		t.Fatalf("consumed = %d, want 2", out.EvaluationsConsumed)
	}
}

// F5 regression: a fully-compared population whose comparisons are all unknown
// is an UNKNOWN assessment, never decisive non-recovery.
func TestAssessProposalsUnknownOnlyIsNotDecisive(t *testing.T) {
	targets := []canon.MechanismSignature{expSignature(expIDLattice)}
	amb := expSignature()
	amb.Preserves = append(amb.Preserves, canon.FieldClaim{
		FieldKind: domain.FieldPreserves, State: domain.ResolutionAmbiguous, Status: domain.ClaimAmbiguous,
	})
	out := AssessProposals([]ProposalContent{{ProposalID: "fpr_amb", Rank: 0, Signature: amb}}, targets, canon.ProfileMechanismV1(), 0)
	if out.UnknownCount != 1 || out.DecisiveNoCount != 0 {
		t.Fatalf("unknown-only must stay unknown: %+v", out)
	}
	if out.Proposals[0].Assessment != AssessmentUnknown {
		t.Fatalf("assessment = %s, want unknown", out.Proposals[0].Assessment)
	}
}

// F1 regression: when two proposals carry the SAME rank (reachable when a
// pipeline arm mixes newly-written and cross-generation-deduped proposals whose
// per-generation rank_ordinal collides), the assessment order — and therefore
// FirstRecoveryRank, membership order, and budget consumption — must still be
// deterministic. The (rank, proposalID) total order makes it so: with a budget
// that admits exactly one comparison, the lexicographically-first proposal id
// is always the one evaluated, regardless of input slice order.
func TestAssessProposalsCollidingRanksAreDeterministic(t *testing.T) {
	targets := []canon.MechanismSignature{expSignature(expIDLattice)}
	// Both rank 0. "fpr_a" (recovers) sorts before "fpr_b" (distinct) by id.
	recovers := ProposalContent{ProposalID: "fpr_a", Rank: 0, Signature: expSignature(expIDLattice)}
	distinct := ProposalContent{ProposalID: "fpr_b", Rank: 0, Signature: expSignature(expIDResidue)}

	for _, order := range [][]ProposalContent{{recovers, distinct}, {distinct, recovers}} {
		out := AssessProposals(order, targets, canon.ProfileMechanismV1(), 1) // budget: one comparison
		if len(out.Proposals) == 0 || out.Proposals[0].ProposalID != "fpr_a" {
			t.Fatalf("tiebreak must evaluate fpr_a first regardless of input order: %+v", out.Proposals)
		}
		if out.RecoveredCount != 1 || out.FirstRecoveryRank != 0 {
			t.Fatalf("fpr_a recovers within the single-comparison budget: %+v", out)
		}
		// The second proposal never got its comparison → unassessed, not decisive.
		if out.UnassessedCount != 1 || out.DecisiveNoCount != 0 {
			t.Fatalf("budget-starved second proposal must be unassessed: %+v", out)
		}
	}
}

// F1 sibling regression: DetectRecovery shares the same total-order tiebreak, so
// colliding ranks yield a stable fact order independent of input order.
func TestDetectRecoveryCollidingRanksAreDeterministic(t *testing.T) {
	target := expSignature(expIDLattice)
	a := ProposalContent{ProposalID: "fpr_a", Rank: 0, Signature: expSignature(expIDResidue)}
	b := ProposalContent{ProposalID: "fpr_b", Rank: 0, Signature: expSignature(expIDLattice)}
	first := DetectRecovery([]ProposalContent{a, b}, target, canon.ProfileMechanismV1())
	second := DetectRecovery([]ProposalContent{b, a}, target, canon.ProfileMechanismV1())
	if len(first.Facts) != 2 || first.Facts[0].ProposalID != "fpr_a" || first.Facts[1].ProposalID != "fpr_b" {
		t.Fatalf("fact order must be id-stable under equal ranks: %+v", first.Facts)
	}
	if second.Facts[0].ProposalID != first.Facts[0].ProposalID || second.Facts[1].ProposalID != first.Facts[1].ProposalID {
		t.Fatalf("fact order must not depend on input order: %+v vs %+v", first.Facts, second.Facts)
	}
}
