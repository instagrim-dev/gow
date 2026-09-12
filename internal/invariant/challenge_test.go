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

// chVocab builds the pinned test vocabulary the split/merge admissibility gate
// checks against. Every canonical id these challenge tests reference is
// registered as a preserves-field term so the tests exercise grounding /
// discrimination logic rather than failing the vocabulary gate. Grounding
// (whether any family PRESERVES an id) is orthogonal to admissibility (whether
// the id is IN the pinned vocabulary): `unused_everywhere` is admissible but
// ungrounded on purpose.
func chVocab(t *testing.T) *canon.Vocabulary {
	t.Helper()
	defs := []canon.TermDef{
		{CanonicalID: chIDResidue, FieldKind: domain.FieldPreserves},
		{CanonicalID: chIDSieve, FieldKind: domain.FieldPreserves},
		{CanonicalID: chIDGlobal, FieldKind: domain.FieldPreserves},
		{CanonicalID: chIDBounds, FieldKind: domain.FieldPreserves},
		{CanonicalID: "core.operator.unused_everywhere", FieldKind: domain.FieldPreserves},
	}
	vocab, err := canon.BuildVocabulary("mechanism/v1", defs)
	if err != nil {
		t.Fatalf("build test vocab: %v", err)
	}
	return vocab
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
	res := VerifyKnownCounterexample(chPredicate(chIDResidue), families, AssociationRecurring)
	if !res.Confirmed {
		t.Fatalf("expected confirmed, got %+v", res)
	}
	if len(res.Evidence) != 1 || res.Evidence[0].ClusterID != "mcl_f3" || res.Evidence[0].SignatureID != "msig_f3" {
		t.Fatalf("evidence must name the violating member: %+v", res.Evidence)
	}
	// v36/S5: a confirmed counterexample derives a typed boundary delta — the
	// claim predicate is the separating condition, in canonical form.
	if res.Delta == nil || res.Delta.Kind != DeltaCounterexampleSeparation ||
		res.Delta.PredicateFingerprint != Fingerprint(chPredicate(chIDResidue)) ||
		res.Delta.Condition != Condition(chPredicate(chIDResidue)) {
		t.Fatalf("confirmed counterexample must carry a typed delta: %+v", res.Delta)
	}
	// A predicate every failure family satisfies has no known counterexample.
	all := Predicate{Schema: PredicateSchemaV1, Root: Node{Op: OpAny, Children: []Node{
		chPredicate(chIDResidue).Root, chPredicate(chIDSieve).Root,
	}}}
	if res := VerifyKnownCounterexample(all, families, AssociationRecurring); res.Confirmed {
		t.Fatalf("expected unconfirmed, got %+v", res)
	}
}

func TestVerifyKnownCounterexampleUnknownIsNotACounterexample(t *testing.T) {
	sig := chSignature("msig_amb", domain.OutcomeFailure)
	sig.Preserves = append(sig.Preserves, canon.FieldClaim{
		FieldKind: domain.FieldPreserves, State: domain.ResolutionAmbiguous, Status: domain.ClaimAmbiguous,
	})
	families := []Family{chFamily("mcl_amb", domain.OutcomeFailure, sig)}
	if res := VerifyKnownCounterexample(chPredicate(chIDResidue), families, AssociationRecurring); res.Confirmed {
		t.Fatalf("ambiguous member must not be a counterexample: %+v", res)
	}
}

// TestVerifyKnownCounterexampleDispositionsSeparated is the H4 regression: an
// empty eligible population, an unknown-only population, and a fully-decided
// negative must NOT collapse to the same completed_negative outcome. Only the
// last is a completed negative that can earn survival.
func TestVerifyKnownCounterexampleDispositionsSeparated(t *testing.T) {
	pred := chPredicate(chIDResidue)

	// (a) No failure-side members at all -> inapplicable (empty population).
	empty := VerifyKnownCounterexample(pred, nil, AssociationRecurring)
	if empty.Confirmed || empty.Outcome != OutcomeInapplicable {
		t.Fatalf("empty population must be inapplicable, got outcome=%q confirmed=%v", empty.Outcome, empty.Confirmed)
	}

	// (b) Only an unknown-evaluating eligible member -> inconclusive.
	amb := chSignature("msig_amb", domain.OutcomeFailure)
	amb.Preserves = append(amb.Preserves, canon.FieldClaim{
		FieldKind: domain.FieldPreserves, State: domain.ResolutionAmbiguous, Status: domain.ClaimAmbiguous,
	})
	unknownOnly := VerifyKnownCounterexample(pred, []Family{chFamily("mcl_amb", domain.OutcomeFailure, amb)}, AssociationRecurring)
	if unknownOnly.Confirmed || unknownOnly.Outcome != OutcomeInconclusive {
		t.Fatalf("unknown-only population must be inconclusive, got outcome=%q confirmed=%v", unknownOnly.Outcome, unknownOnly.Confirmed)
	}

	// (c) A nonempty, fully-decided negative (every eligible member decisively
	// SATISFIES the predicate, none violates) -> completed_negative.
	sat := chSignature("msig_sat", domain.OutcomeFailure, chIDResidue) // satisfies preserves(residue)
	decided := VerifyKnownCounterexample(pred, []Family{chFamily("mcl_sat", domain.OutcomeFailure, sat)}, AssociationRecurring)
	if decided.Confirmed || decided.Outcome != OutcomeCompletedNegative {
		t.Fatalf("fully-decided negative must be completed_negative, got outcome=%q confirmed=%v", decided.Outcome, decided.Confirmed)
	}
}

// TestVerifySuccessPreservingDispositionsSeparated is the H4 regression for the
// success-preserving verifier: empty contrast population is inapplicable, an
// unknown-evaluating contrast member leaves the negative inconclusive, and only
// a fully-decided non-preserving contrast population is a completed negative.
func TestVerifySuccessPreservingDispositionsSeparated(t *testing.T) {
	pred := chPredicate(chIDResidue)

	// (a) No contrast (success-side) members -> inapplicable.
	empty := VerifySuccessPreserving(pred, nil)
	if empty.Confirmed || empty.Outcome != OutcomeInapplicable {
		t.Fatalf("empty contrast population must be inapplicable, got outcome=%q confirmed=%v", empty.Outcome, empty.Confirmed)
	}

	// (b) A success-side member that evaluates unknown -> inconclusive.
	amb := chSignature("msig_amb_s", domain.OutcomePartialSuccess)
	amb.Preserves = append(amb.Preserves, canon.FieldClaim{
		FieldKind: domain.FieldPreserves, State: domain.ResolutionAmbiguous, Status: domain.ClaimAmbiguous,
	})
	unknownOnly := VerifySuccessPreserving(pred, []Family{chFamily("mcl_s_amb", domain.OutcomePartialSuccess, amb)})
	if unknownOnly.Confirmed || unknownOnly.Outcome != OutcomeInconclusive {
		t.Fatalf("unknown contrast member must be inconclusive, got outcome=%q confirmed=%v", unknownOnly.Outcome, unknownOnly.Confirmed)
	}

	// (c) A decisively non-preserving contrast family (violates the predicate) ->
	// completed_negative.
	nonPreserving := chSignature("msig_s_np", domain.OutcomePartialSuccess, chIDSieve) // does NOT preserve residue
	decided := VerifySuccessPreserving(pred, []Family{chFamily("mcl_s_np", domain.OutcomePartialSuccess, nonPreserving)})
	if decided.Confirmed || decided.Outcome != OutcomeCompletedNegative {
		t.Fatalf("decisively non-preserving contrast must be completed_negative, got outcome=%q confirmed=%v", decided.Outcome, decided.Confirmed)
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

// TestVerifySyntheticEmptyDescriptionIsProposalNotCounterexample is the F3
// regression: an EMPTY (unobserved) synthetic description must NOT confirm a
// contains-based counterexample. Marking the field unobserved (as the real
// pipeline syntheticSignature now does) means absence evaluates to `unknown`,
// so `contains(preserves, residue)` on an empty synthetic is inert — proving
// only that a DESCRIPTION can omit residue, not that an admissible failed
// approach without it exists.
func TestVerifySyntheticEmptyDescriptionIsProposalNotCounterexample(t *testing.T) {
	empty := canon.MechanismSignature{
		SchemaVersion:     canon.SchemaMechanismV1,
		VocabularyVersion: "mechanism/v1",
		OutcomeClass:      domain.OutcomeFailure,
		Preserves:         []canon.FieldClaim{},
		// unobserved: absence is an epistemic gap, not a verified negative (F3).
		SetFieldCompleteness: map[domain.FieldKind]domain.FieldCompleteness{
			domain.FieldPreserves: domain.CompletenessUnobserved,
		},
	}
	if res := VerifySyntheticCounterexample(chPredicate(chIDResidue), empty); res.Confirmed {
		t.Fatalf("empty/unobserved synthetic must not confirm a counterexample (F3): %+v", res)
	}
}

// TestVerifyKnownCounterexampleRefutationDependsOnClaimKind is the F3 regression
// for claim-kind-scoped refutation: the SAME failure-side violation falsifies a
// `recurring` (universal-regularity) candidate but leaves a `contrast_observed`
// (association/discrimination) candidate unconfirmed — a lone counterexample
// does not refute an association claim.
func TestVerifyKnownCounterexampleRefutationDependsOnClaimKind(t *testing.T) {
	families := chAtlas() // mcl_f3 violates preserves(residue)
	if res := VerifyKnownCounterexample(chPredicate(chIDResidue), families, AssociationRecurring); !res.Confirmed {
		t.Fatalf("recurring claim must be falsified by a known counterexample: %+v", res)
	}
	res := VerifyKnownCounterexample(chPredicate(chIDResidue), families, AssociationContrastObserved)
	if res.Confirmed {
		t.Fatalf("contrast_observed claim must NOT be falsified by an isolated counterexample (F3): %+v", res)
	}
	if len(res.Evidence) == 0 {
		t.Fatalf("the violation should still be recorded as audit evidence: %+v", res)
	}
}

// TestVerifyMergeRejectsOutcomeReadingChild is the F4 regression: a merged child
// that reads the outcome axis (e.g. outcome in [failure, partial_failure]) earns
// coverage/contrast by definition and identifies no mechanism. The shared
// AdmitCandidate gate inside VerifyMerge must reject it, closing the target leak
// the mining fix excludes.
func TestVerifyMergeRejectsOutcomeReadingChild(t *testing.T) {
	families := chAtlas()
	parents := []Predicate{chPredicate(chIDResidue), chPredicate(chIDBounds)}
	outcomeChild := Predicate{Schema: PredicateSchemaV1, Root: Node{
		Op: OpIn, Field: FieldOutcome, Values: []string{string(domain.OutcomeFailure), string(domain.OutcomePartialFailure)},
	}}
	res := VerifyMerge(parents, outcomeChild, families, chVocab(t))
	if res.Confirmed || !strings.Contains(res.Detail, "not admissible") {
		t.Fatalf("outcome-reading merge child must be rejected as inadmissible (F4): %+v", res)
	}
}

// TestVerifySplitRejectsChildOutsideParentSupport is the F4/refinement
// regression: a split child that grounds to a failure family the PARENT does
// not support is a new invariant, not a refinement of the parent, and must be
// rejected even when the children are pairwise-disjoint and distinct.
func TestVerifySplitRejectsChildOutsideParentSupport(t *testing.T) {
	// Parent = any(residue, global), supported by {f1, f2}. Children residue
	// {f1,f2} and bounds {f3} are disjoint and both distinct from the parent, but
	// bounds grounds to f3, a failure family the parent does NOT support.
	parent := Predicate{Schema: PredicateSchemaV1, Root: Node{Op: OpAny, Children: []Node{
		chPredicate(chIDResidue).Root, chPredicate(chIDGlobal).Root,
	}}}
	res := VerifySplit(parent, []Predicate{chPredicate(chIDResidue), chPredicate(chIDBounds)}, chAtlas(), chVocab(t))
	if res.Confirmed || !strings.Contains(res.Detail, "not a refinement") {
		t.Fatalf("split child outside parent support must be rejected as non-refinement (F4): %+v", res)
	}
}

// TestVerifySyntheticUnsupportedPresenceIsProposalNotCounterexample is the G3
// regression: a provider can assert an UNSUPPORTED present property that violates
// a negated predicate (here `not contains(preserves, X)` with X present but
// unsupported). Presence in a generated description is not stronger evidence of
// realizability than absence from it, so the synthetic is a PROPOSAL — it must
// not confirm a weakening without admissibly supported structure.
func TestVerifySyntheticUnsupportedPresenceIsProposalNotCounterexample(t *testing.T) {
	negated := Predicate{Schema: PredicateSchemaV1, Root: Node{
		Op:       OpNot,
		Children: []Node{{Op: OpContains, Field: FieldPreserves, CanonicalID: chIDSieve}},
	}}
	// X (sieve) is present but only as an UNSUPPORTED model assertion.
	unsupported := canon.MechanismSignature{
		SchemaVersion:     canon.SchemaMechanismV1,
		VocabularyVersion: "mechanism/v1",
		OutcomeClass:      domain.OutcomeFailure,
		Preserves: []canon.FieldClaim{{
			FieldKind:   domain.FieldPreserves,
			State:       domain.ResolutionResolved,
			CanonicalID: domain.CanonicalID(chIDSieve),
			Status:      domain.ClaimUnsupported,
		}},
		SetFieldCompleteness: map[domain.FieldKind]domain.FieldCompleteness{
			domain.FieldPreserves: domain.CompletenessComplete,
		},
	}
	// Sanity: the construction really does violate the negated predicate.
	if Evaluate(negated, unsupported) != VerdictViolates {
		t.Fatalf("test setup: unsupported-present synthetic should violate the negated predicate")
	}
	res := VerifySyntheticCounterexample(negated, unsupported)
	if res.Confirmed {
		t.Fatalf("an unsupported present property must not confirm a synthetic counterexample (G3): %+v", res)
	}
	if res.Outcome != OutcomeInconclusive {
		t.Fatalf("an unsupported-present synthetic should be inconclusive (a proposal), got %q", res.Outcome)
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
	res := VerifySplit(parent, []Predicate{chPredicate(chIDResidue), chPredicate(chIDSieve)}, chAtlas(), chVocab(t))
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
	res := VerifySplit(overlapParent, []Predicate{chPredicate(chIDResidue), chPredicate(chIDGlobal)}, chAtlas(), chVocab(t))
	if res.Confirmed || !strings.Contains(res.Detail, "not disjoint") {
		t.Fatalf("expected overlap rejection, got %+v", res)
	}
	// Ungrounded child: nothing preserves an unused id. The parent must be the
	// composite any(residue, sieve) so the grounded sibling (sieve -> f3) stays
	// INSIDE the parent's support and the ungrounded branch is what fires.
	compositeParent := Predicate{Schema: PredicateSchemaV1, Root: Node{Op: OpAny, Children: []Node{
		chPredicate(chIDResidue).Root, chPredicate(chIDSieve).Root,
	}}}
	unused := chPredicate("core.operator.unused_everywhere")
	res = VerifySplit(compositeParent, []Predicate{chPredicate(chIDSieve), unused}, chAtlas(), chVocab(t))
	if res.Confirmed || !strings.Contains(res.Detail, "grounds to no supporting") {
		t.Fatalf("expected ungrounded rejection, got %+v", res)
	}
	// A child identical to the parent is not a split.
	parent := chPredicate(chIDResidue)
	res = VerifySplit(parent, []Predicate{parent, chPredicate(chIDSieve)}, chAtlas(), chVocab(t))
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
	res := VerifyMerge(parents, child, families, chVocab(t))
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
	res = VerifyMerge(parents2, child2, families, chVocab(t))
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
	res := VerifyMerge([]Predicate{chPredicate(chIDResidue), chPredicate(chIDBounds)}, chPredicate(chIDResidue), families, chVocab(t))
	if res.Confirmed || !strings.Contains(res.Detail, "does not cover") {
		t.Fatalf("expected coverage rejection, got %+v", res)
	}
}

// TestBoundaryDeltasTyped pins v36/S5: every confirmed challenge derives a
// typed BoundaryDelta with the fields its kind requires, and unconfirmed
// results carry none. The delta is code-derived — subsequent policy consumes a
// typed object, never a transcript.
func TestBoundaryDeltasTyped(t *testing.T) {
	families := chAtlas()

	// support-recount: measured support + threshold are recorded.
	recount := VerifyBiasCritique(chPredicate(chIDResidue), families, 5)
	if !recount.Confirmed || recount.Delta == nil || recount.Delta.Kind != DeltaSupportRecount ||
		recount.Delta.MeasuredSupport != 2 || recount.Delta.SupportThreshold != 5 {
		t.Fatalf("recount delta must carry measured/threshold: %+v", recount.Delta)
	}
	// A negative recount (support holds) derives no delta.
	if hold := VerifyBiasCritique(chPredicate(chIDResidue), families, 1); hold.Confirmed || hold.Delta != nil {
		t.Fatalf("a completed-negative recount must carry no delta: %+v", hold)
	}

	// contrast-collapse: the success family preserving the predicate.
	preserve := VerifySuccessPreserving(chPredicate(chIDResidue), families)
	if !preserve.Confirmed || preserve.Delta == nil || preserve.Delta.Kind != DeltaContrastCollapse ||
		preserve.Delta.Condition != Condition(chPredicate(chIDResidue)) {
		t.Fatalf("success-preserving delta must carry the collapsed condition: %+v", preserve.Delta)
	}

	// split-partition: child fingerprints in proposal order.
	parent := Predicate{Schema: PredicateSchemaV1, Root: Node{Op: OpAny, Children: []Node{
		chPredicate(chIDResidue).Root, chPredicate(chIDSieve).Root,
	}}}
	children := []Predicate{chPredicate(chIDResidue), chPredicate(chIDSieve)}
	split := VerifySplit(parent, children, families, chVocab(t))
	if !split.Confirmed || split.Delta == nil || split.Delta.Kind != DeltaSplitPartition ||
		len(split.Delta.ChildFingerprints) != 2 ||
		split.Delta.ChildFingerprints[0] != Fingerprint(children[0]) ||
		split.Delta.ChildFingerprints[1] != Fingerprint(children[1]) {
		t.Fatalf("split delta must carry ordered child fingerprints: %+v", split.Delta)
	}
	if split.Delta.PredicateFingerprint != Fingerprint(parent) {
		t.Fatalf("split delta must be measured against the parent claim: %+v", split.Delta)
	}
}
