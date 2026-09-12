package policy

import (
	"testing"

	"github.com/instagrim-dev/newf/internal/domain"
)

// TestDeriveRefutedBoundaries pins the D5 edge (v43): separation-class
// boundary deltas earn one expand directive per predicate fingerprint;
// bookkeeping-class and structural deltas earn nothing here.
func TestDeriveRefutedBoundaries(t *testing.T) {
	pol := Derive(Evidence{RefutedBoundaries: []RefutedBoundaryEvidence{
		{PredicateFingerprint: "fp-a", DeltaKind: "counterexample-separation", ChallengeID: "icl_1"},
		{PredicateFingerprint: "fp-a", DeltaKind: "constructibility", ChallengeID: "icl_2"}, // dedupes onto fp-a
		{PredicateFingerprint: "fp-b", DeltaKind: "constructibility", ChallengeID: "icl_3"},
		{PredicateFingerprint: "fp-c", DeltaKind: "contrast-collapse", ChallengeID: "icl_4"},   // epistemic defect: no directive
		{PredicateFingerprint: "fp-d", DeltaKind: "support-recount", ChallengeID: "icl_5"},     // epistemic defect: no directive
		{PredicateFingerprint: "fp-e", DeltaKind: "split-partition", ChallengeID: "icl_6"},     // children own their lifecycle
		{PredicateFingerprint: "", DeltaKind: "counterexample-separation", ChallengeID: "icl"}, // no identity: skipped
	}})
	if len(pol.Directives) != 2 {
		t.Fatalf("want 2 expand directives (fp-a deduped, fp-b), got %+v", pol.Directives)
	}
	for _, d := range pol.Directives {
		if d.Kind != KindExpand || d.TargetKind != TargetRefutedBoundary || d.Weight != domain.OrdinalMedium {
			t.Fatalf("boundary directive shape wrong: %+v", d)
		}
	}
	if pol.Directives[0].TargetID != "fp-a" || pol.Directives[0].Source != "counterexample-separation" {
		t.Fatalf("fp-a directive must keep the FIRST separation delta's kind as source: %+v", pol.Directives[0])
	}
	if pol.Directives[1].TargetID != "fp-b" || pol.Directives[1].Source != "constructibility" {
		t.Fatalf("fp-b directive wrong: %+v", pol.Directives[1])
	}
}

// TestAdmitProposedRefutedBoundary pins provider parity: an expand proposal at
// a separation-backed fingerprint admits (capped medium); anything else is
// inert — a bookkeeping delta is not evidence FOR expansion, and no other kind
// may target a refuted boundary.
func TestAdmitProposedRefutedBoundary(t *testing.T) {
	ev := Evidence{RefutedBoundaries: []RefutedBoundaryEvidence{
		{PredicateFingerprint: "fp-a", DeltaKind: "counterexample-separation", ChallengeID: "icl_1"},
		{PredicateFingerprint: "fp-c", DeltaKind: "contrast-collapse", ChallengeID: "icl_4"},
	}}

	d, ok := AdmitProposedDirective(KindExpand, TargetRefutedBoundary, "fp-a", ev)
	if !ok || d.Weight != domain.OrdinalMedium || d.Source != "provider:counterexample-separation" {
		t.Fatalf("separation-backed expand must admit at medium: %+v ok=%v", d, ok)
	}
	if _, ok := AdmitProposedDirective(KindExpand, TargetRefutedBoundary, "fp-c", ev); ok {
		t.Fatal("a bookkeeping-class delta must not admit an expand proposal")
	}
	if _, ok := AdmitProposedDirective(KindExpand, TargetRefutedBoundary, "fp-x", ev); ok {
		t.Fatal("an unresolvable fingerprint must be inert")
	}
	if _, ok := AdmitProposedDirective(KindAvoid, TargetRefutedBoundary, "fp-a", ev); ok {
		t.Fatal("only expand may target a refuted boundary")
	}
}
