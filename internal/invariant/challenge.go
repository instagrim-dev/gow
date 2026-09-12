// Package invariant — deterministic challenge verifiers (M4.3, U3).
//
// A challenge ATTACKS a candidate invariant. The provider proposes challenges
// with a claimed verdict; these pure functions confirm or deny each claim
// against rehydrated signatures. A claim code cannot confirm is `unconfirmed`
// and inert: it neither strengthens nor weakens the hypothesis
// (ModelJudgment != Verification). No SQL, Cobra, or provider coupling.
package invariant

import (
	"fmt"
	"sort"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/domain"
)

// ChallengeType enumerates the attack families (docs/persistence.md blueprint;
// independent-verification is the code-gated path toward `established`).
type ChallengeType string

const (
	ChallengeKnownCounterexample     ChallengeType = "known-counterexample"
	ChallengeSyntheticCounterexample ChallengeType = "synthetic-counterexample"
	ChallengeSuccessPreserving       ChallengeType = "success-preserving"
	ChallengeSplit                   ChallengeType = "split"
	ChallengeMerge                   ChallengeType = "merge"
	ChallengeBiasCritique            ChallengeType = "bias-critique"
	ChallengeIndependentVerification ChallengeType = "independent-verification"
)

// Valid reports whether the challenge type is known.
func (t ChallengeType) Valid() bool {
	switch t {
	case ChallengeKnownCounterexample, ChallengeSyntheticCounterexample,
		ChallengeSuccessPreserving, ChallengeSplit, ChallengeMerge,
		ChallengeBiasCritique, ChallengeIndependentVerification:
		return true
	default:
		return false
	}
}

// ChallengeEvidence is one concrete evidence handle backing a confirmed
// challenge: a real cluster/signature, a recount, or a grounding fact.
// result_summary is a label over this evidence, never a substitute for it.
type ChallengeEvidence struct {
	Kind        string // counterexample_member | success_family | support_recount | grounding
	ClusterID   string
	SignatureID string
	Detail      string
}

const (
	EvidenceCounterexampleMember = "counterexample_member"
	EvidenceSuccessFamily        = "success_family"
	EvidenceSupportRecount       = "support_recount"
	EvidenceGrounding            = "grounding"
	EvidenceIndependentSource    = "independent_source"
)

// CheckOutcome is the explicit disposition of one attempted challenge (G2). It
// separates admissibility, decisiveness, and confirmation so the campaign
// verdict is derived from typed outcomes under a required-check policy — not
// from a default-true "applicable" flag. Only a `CompletedNegative` counts as a
// completed applicable negative search toward survival; an `Inadmissible` or
// `Inconclusive` attempt does neither strengthens nor weakens (KTD-3).
type CheckOutcome string

const (
	// OutcomeInadmissible: the attack could not be evaluated at all (validation
	// rejected it — e.g. a split with <2 children, a merge naming no partners, a
	// synthetic with no construction, a provider over-claiming operator-only
	// verification). It is not a search over any population.
	OutcomeInadmissible CheckOutcome = "inadmissible"
	// OutcomeInapplicable: the attack was well-formed but its eligible population
	// was EMPTY — there were zero applicable cases to check. Zero eligible cases
	// decide nothing (H4); it is neither a completed negative search nor a
	// confirmation, and never earns survival.
	OutcomeInapplicable CheckOutcome = "inapplicable"
	// OutcomeInconclusive: the attack ran but reached no definite result over its
	// eligible population (unresolved/unknown evidence, an unsupported synthetic
	// proposal). It neither confirms nor completes a negative search.
	OutcomeInconclusive CheckOutcome = "inconclusive"
	// OutcomeCompletedNegative: a decisive negative — the attack ran a real
	// determination over a NONEMPTY eligible population, every applicable case was
	// decisively resolved, and the claim did not land (e.g. every eligible failure
	// member was decisively checked and none violated; support holds under
	// recomputation). This is the only disposition that legitimately earns
	// `surviving`.
	OutcomeCompletedNegative CheckOutcome = "completed_negative"
	// OutcomeConfirmed: the attack landed (a confirmed counterexample / weakening).
	OutcomeConfirmed CheckOutcome = "confirmed"
)

// ChallengeResult is the deterministic verdict on one proposed challenge. Outcome
// is the typed disposition (G2); Confirmed is retained as the boolean shorthand
// for Outcome == OutcomeConfirmed. Delta is the typed boundary refinement a
// CONFIRMED challenge derives where code-derivable; nil otherwise.
type ChallengeResult struct {
	Confirmed bool
	Outcome   CheckOutcome
	Evidence  []ChallengeEvidence
	Detail    string
	Delta     *BoundaryDelta
}

// Boundary-delta kinds: which structural refinement a confirmed challenge
// observed. Each kind fixes which BoundaryDelta fields are meaningful.
const (
	// DeltaCounterexampleSeparation: the claim predicate itself is the
	// separating condition; the recorded counterexample members sit on its
	// violating side (known-counterexample confirmed).
	DeltaCounterexampleSeparation = "counterexample-separation"
	// DeltaContrastCollapse: the predicate fails to separate outcomes — one or
	// more success-side families preserve it (success-preserving confirmed).
	DeltaContrastCollapse = "contrast-collapse"
	// DeltaConstructibility: a supported synthetic construction violates the
	// predicate, so the regularity is not conserved by necessity.
	DeltaConstructibility = "constructibility"
	// DeltaSupportRecount: the deterministic recount moved the claim's measured
	// support below its admission threshold (bias-critique confirmed).
	DeltaSupportRecount = "support-recount"
	// DeltaSplitPartition: the claim is >=2 invariants masquerading as one; the
	// children partition its support (split confirmed).
	DeltaSplitPartition = "split-partition"
	// DeltaMergeUnion: several claims merge into one child covering the union
	// of their support without discrimination loss (merge confirmed).
	DeltaMergeUnion = "merge-union"
)

// BoundaryDelta is the typed structural refinement a confirmed challenge
// derives: the minimal condition separating the observed refuter/refinement
// from the claim's support population (glossary `boundary_delta`). It is
// code-derived at confirmation time — never provider prose — so subsequent
// policy can consume it without reinterpreting a transcript. The disposition
// lives on the challenge's persisted transitions and any derived candidate ids
// on its derived-children rows; the delta carries only the boundary itself.
type BoundaryDelta struct {
	// Kind is one of the Delta* constants.
	Kind string
	// PredicateFingerprint is the semantic identity of the claim predicate the
	// delta is measured against.
	PredicateFingerprint string
	// Condition is the canonical serialization of the separating condition (the
	// canonicalized predicate AST — deterministic, paraphrase-stable).
	Condition string
	// MeasuredSupport / SupportThreshold are meaningful only for
	// support-recount deltas (both zero otherwise is impossible: a recount
	// confirms only when measured < threshold and thresholds are >= 1).
	MeasuredSupport  int
	SupportThreshold int
	// ChildFingerprints are the derived child predicates' fingerprints for
	// split-partition (>=2) and merge-union (exactly 1) deltas, in proposal
	// order.
	ChildFingerprints []string
}

// Condition returns the canonical, paraphrase-stable serialization of the
// predicate — the same rendering Fingerprint hashes. It is the typed textual
// form a BoundaryDelta records as its separating condition.
func Condition(p Predicate) string {
	return serializeNode(Canonicalize(p).Root)
}

// unconfirmed builds an INCONCLUSIVE inert result (KTD-3): persisted for audit,
// linked to no evidence, driving no transition. Callers that mean a decisive
// negative search use completedNegative; callers rejecting an inadmissible
// attack use inadmissible.
func unconfirmed(reason string) ChallengeResult {
	return ChallengeResult{Confirmed: false, Outcome: OutcomeInconclusive, Detail: reason}
}

// completedNegative builds a decisive-negative result: the attack ran over an
// eligible population and the claim did not land. It carries any audit evidence
// (e.g. a support recount) and counts toward survival.
func completedNegative(reason string, ev ...ChallengeEvidence) ChallengeResult {
	return ChallengeResult{Confirmed: false, Outcome: OutcomeCompletedNegative, Detail: reason, Evidence: ev}
}

// inadmissible builds a rejected-attack result: the attack could not be
// evaluated over any population (validation failure). It never counts toward
// survival and drives no transition.
func inadmissible(reason string) ChallengeResult {
	return ChallengeResult{Confirmed: false, Outcome: OutcomeInadmissible, Detail: reason}
}

// inapplicable builds an empty-population result: the attack was well-formed but
// no eligible cases existed, so it decides nothing (H4). It never counts toward
// survival.
func inapplicable(reason string) ChallengeResult {
	return ChallengeResult{Confirmed: false, Outcome: OutcomeInapplicable, Detail: reason}
}

// AssociationKind is the refutation-governing claim class the CALLER derives for
// a candidate. It is deliberately distinct from the engine's measured
// association_status: recurrence is a frequency label, not a logical quantifier
// (G3). The pipeline passes `recurring` here ONLY when the corpus actually
// exhibits universality over the eligible failure families (full failure
// coverage); a recurrence that is not universal is passed as
// `contrast_observed`. Refutation conditions differ: a `recurring` (verified
// universal) claim is falsified by ONE known in-atlas counterexample; a
// `contrast_observed` (association) claim is not refuted by an isolated
// counterexample; an `unknown` claim makes no confirmed regularity assertion, so
// a lone counterexample is inconclusive.
type AssociationKind string

const (
	AssociationRecurring        AssociationKind = "recurring"
	AssociationContrastObserved AssociationKind = "contrast_observed"
	AssociationUnknown          AssociationKind = "unknown"
)

// VerifyKnownCounterexample confirms the claim that a KNOWN failed approach in
// the atlas violates the predicate: some failure-side member evaluates to
// `violates` (not `unknown` — ambiguity is not a counterexample). The refutation
// consequence depends on the candidate's association kind (F3): a single known
// violation FALSIFIES a `recurring` (universal-regularity) claim, but only a
// `contrast_observed` (association/discrimination) claim survives such a
// violation — for it, isolated failure-side violations are recorded as evidence
// but do not by themselves refute the association. An `unknown`-kind candidate
// makes no confirmed regularity claim, so a lone counterexample is inconclusive.
func VerifyKnownCounterexample(pred Predicate, families []Family, kind AssociationKind) ChallengeResult {
	var ev []ChallengeEvidence
	eligible, unknown := 0, 0
	for _, fam := range families {
		for _, m := range fam.Members {
			role, ok := memberRole(fam, m)
			if !ok || role != RoleSupport {
				continue
			}
			eligible++
			switch Evaluate(pred, m.Signature) {
			case VerdictViolates:
				ev = append(ev, ChallengeEvidence{
					Kind:        EvidenceCounterexampleMember,
					ClusterID:   fam.ClusterID,
					SignatureID: m.SignatureID,
					Detail:      "failure-side member violates the predicate",
				})
			case VerdictUnknown:
				unknown++
			}
		}
	}
	if len(ev) == 0 {
		// No violation found. Distinguish the three collapsed dispositions (H4):
		// an empty eligible population decides nothing; unresolved members are
		// inconclusive; only a fully-decided nonempty population is a completed
		// negative that may count toward survival.
		switch {
		case eligible == 0:
			return inapplicable("no eligible failure-side member to check for a counterexample (empty population)")
		case unknown > 0:
			return unconfirmed(fmt.Sprintf("%d of %d eligible failure-side member(s) evaluated unknown; no decisive counterexample and not a completed negative search", unknown, eligible))
		default:
			return completedNegative(fmt.Sprintf("all %d eligible failure-side member(s) decisively checked; none violates the predicate", eligible))
		}
	}
	sort.Slice(ev, func(i, j int) bool {
		if ev[i].ClusterID != ev[j].ClusterID {
			return ev[i].ClusterID < ev[j].ClusterID
		}
		return ev[i].SignatureID < ev[j].SignatureID
	})
	// A known counterexample falsifies ONLY a universal-regularity (recurring)
	// claim. A contrast/association claim is not refuted by isolated failure-side
	// violations (its claim is about discrimination, not universality); the
	// violation is retained as audit evidence but the result is inconclusive.
	if kind != AssociationRecurring {
		r := unconfirmed(fmt.Sprintf("%d failure-side violation(s) recorded, but a %q claim is not refuted by isolated counterexamples (it asserts discrimination, not a universal regularity)", len(ev), kind))
		r.Evidence = ev
		return r
	}
	return ChallengeResult{Confirmed: true, Outcome: OutcomeConfirmed, Evidence: ev,
		Detail: fmt.Sprintf("%d known counterexample member(s) refute the universal claim", len(ev)),
		Delta: &BoundaryDelta{
			Kind:                 DeltaCounterexampleSeparation,
			PredicateFingerprint: Fingerprint(pred),
			Condition:            Condition(pred),
		}}
}

// VerifySyntheticCounterexample evaluates a provider-authored CONSTRUCTED
// approach against the predicate. A synthetic construction is a PROPOSAL, never
// a demonstrated construction, and confirmation is code-owned (G3):
//
//   - Its set fields are `unobserved` (see syntheticSignature), so
//     `contains(field, X)` on a description that merely omits X evaluates to
//     `unknown`, not `violates` — omitting X does not prove an admissible failed
//     approach WITHOUT X exists.
//   - Its present claims are `unsupported` model assertions. Presence in a
//     generated description is not stronger evidence of realizability than
//     absence from it, so a violation that rests on unsupported present claims
//     is still only a proposal: it is recorded inert and does NOT weaken the
//     invariant. A synthetic can confirm a weakening only when the violation is
//     driven by admissibly SUPPORTED structure (a value whose claim status is
//     explicit/inferred, not unsupported) — which the current fixture path never
//     supplies, so today no synthetic weakens by fiat.
//
// This keeps description-omission and unsupported-description-presence as
// proposals, not confirmed grounds for weakening.
func VerifySyntheticCounterexample(pred Predicate, synthetic canon.MechanismSignature) ChallengeResult {
	if Evaluate(pred, synthetic) != VerdictViolates {
		return unconfirmed("synthetic construction does not violate the predicate (an omitted or unobserved field is a proposal, not a demonstrated counterexample)")
	}
	if !hasSupportedClaim(synthetic) {
		return unconfirmed("synthetic construction violates only via unsupported model-asserted structure; a proposal, not a demonstrated admissible failed approach (G3)")
	}
	return ChallengeResult{Confirmed: true, Outcome: OutcomeConfirmed, Detail: "synthetic construction violates the predicate via supported structure",
		Delta: &BoundaryDelta{
			Kind:                 DeltaConstructibility,
			PredicateFingerprint: Fingerprint(pred),
			Condition:            Condition(pred),
		}}
}

// hasSupportedClaim reports whether any resolved set-field claim in the
// signature carries an admissible support status (explicit or inferred), i.e.
// evidence beyond a bare unsupported model assertion. Used to gate synthetic
// confirmation (G3): a construction whose violating structure is entirely
// unsupported is a proposal, not a demonstrated counterexample.
func hasSupportedClaim(sig canon.MechanismSignature) bool {
	sets := [][]canon.FieldClaim{
		sig.Representations, sig.Operators, sig.Assumptions,
		sig.Preserves, sig.Breaks, sig.AuxiliaryObjects,
	}
	for _, claims := range sets {
		for _, c := range claims {
			if c.State != domain.ResolutionResolved {
				continue
			}
			if c.Status == domain.ClaimExplicit || c.Status == domain.ClaimInferred {
				return true
			}
		}
	}
	return false
}

// VerifySuccessPreserving confirms that a success/partial-success family
// PRESERVES the invariant anyway — evidence the predicate does not discriminate
// outcome. Confirmed when every eligible contrast-side member of some family
// satisfies (unknown members disqualify that family: ambiguity is not
// preservation).
func VerifySuccessPreserving(pred Predicate, families []Family) ChallengeResult {
	var ev []ChallengeEvidence
	totalEligible, sawUnknown := 0, false
	for _, fam := range families {
		eligible, allSatisfy := 0, true
		for _, m := range fam.Members {
			role, ok := memberRole(fam, m)
			if !ok || role != RoleContrast {
				continue
			}
			eligible++
			switch Evaluate(pred, m.Signature) {
			case VerdictSatisfies:
				// preserves
			case VerdictUnknown:
				allSatisfy = false
				sawUnknown = true
			default:
				allSatisfy = false
			}
		}
		totalEligible += eligible
		if eligible > 0 && allSatisfy {
			ev = append(ev, ChallengeEvidence{
				Kind:      EvidenceSuccessFamily,
				ClusterID: fam.ClusterID,
				Detail:    "success/partial-success family preserves the predicate",
			})
		}
	}
	if len(ev) == 0 {
		// No preserving family found. Distinguish the collapsed dispositions (H4):
		// zero eligible contrast members is inapplicable; unresolved members leave
		// the negative undecided (inconclusive); only a fully-decided nonempty
		// contrast population is a completed negative.
		switch {
		case totalEligible == 0:
			return inapplicable("no eligible success-side member to check for preservation (empty contrast population)")
		case sawUnknown:
			return unconfirmed("success-side preservation undecided: one or more eligible contrast members evaluated unknown; not a completed negative search")
		default:
			return completedNegative(fmt.Sprintf("all %d eligible success-side member(s) decisively checked; none preserves the predicate", totalEligible))
		}
	}
	sort.Slice(ev, func(i, j int) bool { return ev[i].ClusterID < ev[j].ClusterID })
	return ChallengeResult{Confirmed: true, Outcome: OutcomeConfirmed, Evidence: ev,
		Detail: fmt.Sprintf("%d success-preserving family(ies)", len(ev)),
		Delta: &BoundaryDelta{
			Kind:                 DeltaContrastCollapse,
			PredicateFingerprint: Fingerprint(pred),
			Condition:            Condition(pred),
		}}
}

// VerifyBiasCritique recomputes the invariant's distinct-family support from
// persisted signatures (the deterministic counterpart of the model's
// sampling/publication-bias claim) and confirms iff the recomputed support
// falls below minSupport. The recount is the evidence (KTD-6).
func VerifyBiasCritique(pred Predicate, families []Family, minSupport int) ChallengeResult {
	support := supportingFamilies(pred, families)
	detail := fmt.Sprintf("recomputed distinct-family support = %d (threshold %d)", len(support), minSupport)
	ev := []ChallengeEvidence{{Kind: EvidenceSupportRecount, Detail: detail}}
	if len(support) >= minSupport {
		r := completedNegative("support holds under recomputation: "+detail, ev...)
		return r
	}
	return ChallengeResult{Confirmed: true, Outcome: OutcomeConfirmed, Evidence: ev, Detail: detail,
		Delta: &BoundaryDelta{
			Kind:                 DeltaSupportRecount,
			PredicateFingerprint: Fingerprint(pred),
			Condition:            Condition(pred),
			MeasuredSupport:      len(support),
			SupportThreshold:     minSupport,
		}}
}

// VerifySplit confirms that the parent predicate is two-or-more invariants
// masquerading as one: every proposed child must be ADMISSIBLE (the shared
// failure-mechanism + pinned-vocabulary gate, F4), semantically distinct from
// the parent, an actual REFINEMENT of the parent (every family a child grounds
// to must also be a parent-supporting family — disjoint support alone does not
// prove the children partition the PARENT), and grounded in a NONEMPTY,
// PAIRWISE-DISJOINT set of supporting failure families
// (docs/abstraction-safety.md: a split that cannot ground each child into
// distinct concrete cases is rejected).
func VerifySplit(parent Predicate, children []Predicate, families []Family, vocab *canon.Vocabulary) ChallengeResult {
	if len(children) < 2 {
		return inadmissible("a split requires >=2 child predicates")
	}
	parentFP := Fingerprint(parent)
	parentSupport := supportingFamilies(parent, families)
	seen := map[string]struct{}{}
	supports := make([]map[string]struct{}, 0, len(children))
	var ev []ChallengeEvidence
	for i, child := range children {
		if err := AdmitCandidate(child, vocab); err != nil {
			return inadmissible(fmt.Sprintf("child %d not admissible: %v", i, err))
		}
		fp := Fingerprint(child)
		if fp == parentFP {
			return inadmissible(fmt.Sprintf("child %d is semantically identical to the parent", i))
		}
		if _, dup := seen[fp]; dup {
			return inadmissible(fmt.Sprintf("child %d duplicates another child", i))
		}
		seen[fp] = struct{}{}
		sup := supportingFamilies(child, families)
		if len(sup) == 0 {
			return inadmissible(fmt.Sprintf("child %d grounds to no supporting failure family", i))
		}
		// Refinement: a split partitions the PARENT's support, so every family a
		// child grounds to must be a family the parent also supports. A child
		// grounding to a family outside the parent's support is a new invariant,
		// not a refinement of this one (docs/abstraction-safety.md).
		for fam := range sup {
			if _, ok := parentSupport[fam]; !ok {
				return inadmissible(fmt.Sprintf("child %d grounds to family %s outside the parent's support (not a refinement)", i, fam))
			}
		}
		supports = append(supports, sup)
	}
	for i := range supports {
		for j := i + 1; j < len(supports); j++ {
			for fam := range supports[i] {
				if _, overlap := supports[j][fam]; overlap {
					return inadmissible(fmt.Sprintf("children %d and %d share supporting family %s (not disjoint)", i, j, fam))
				}
			}
		}
	}
	for i, sup := range supports {
		for _, fam := range sortedKeys(sup) {
			ev = append(ev, ChallengeEvidence{
				Kind:      EvidenceGrounding,
				ClusterID: fam,
				Detail:    fmt.Sprintf("child %d grounds to family %s", i, fam),
			})
		}
	}
	childFPs := make([]string, 0, len(children))
	for _, child := range children {
		childFPs = append(childFPs, Fingerprint(child))
	}
	return ChallengeResult{Confirmed: true, Outcome: OutcomeConfirmed, Evidence: ev,
		Detail: fmt.Sprintf("split grounds %d children to disjoint support", len(children)),
		Delta: &BoundaryDelta{
			Kind:                 DeltaSplitPartition,
			PredicateFingerprint: parentFP,
			Condition:            Condition(parent),
			ChildFingerprints:    childFPs,
		}}
}

// VerifyMerge confirms that several parent predicates merge into one child at
// a higher abstraction WITHOUT losing predictive discrimination: the child must
// be ADMISSIBLE (the shared failure-mechanism + pinned-vocabulary gate — this
// is what rejects a merged child of `outcome in [failure, partial_failure]`
// that would otherwise cover every failure family and violate every success
// case by definition, F4), its supporting-family set must cover the union of
// the parents', and its contrast violations must be at least each parent's (a
// merge that erases the axis separating outcomes is rejected).
func VerifyMerge(parents []Predicate, child Predicate, families []Family, vocab *canon.Vocabulary) ChallengeResult {
	if len(parents) < 2 {
		return inadmissible("a merge requires >=2 parent predicates")
	}
	if err := AdmitCandidate(child, vocab); err != nil {
		return inadmissible(fmt.Sprintf("child not admissible: %v", err))
	}
	childSupport := supportingFamilies(child, families)
	childContrast := contrastViolations(child, families)
	union := map[string]struct{}{}
	for i, p := range parents {
		if err := AdmitCandidate(p, vocab); err != nil {
			return inadmissible(fmt.Sprintf("parent %d not admissible: %v", i, err))
		}
		for fam := range supportingFamilies(p, families) {
			union[fam] = struct{}{}
		}
		if pc := contrastViolations(p, families); childContrast < pc {
			return inadmissible(fmt.Sprintf("merge loses discrimination: child contrast %d < parent %d contrast %d", childContrast, i, pc))
		}
	}
	for fam := range union {
		if _, ok := childSupport[fam]; !ok {
			return inadmissible(fmt.Sprintf("child support does not cover parent-supported family %s", fam))
		}
	}
	var ev []ChallengeEvidence
	for _, fam := range sortedKeys(childSupport) {
		ev = append(ev, ChallengeEvidence{Kind: EvidenceGrounding, ClusterID: fam,
			Detail: "merged child grounds to family " + fam})
	}
	return ChallengeResult{Confirmed: true, Outcome: OutcomeConfirmed, Evidence: ev,
		Detail: fmt.Sprintf("merge covers %d united support families without discrimination loss", len(union)),
		Delta: &BoundaryDelta{
			// The delta is measured against the ACTING parent (parents[0] by the
			// pipeline's contract); the merge partners' identities travel on the
			// derived-children lineage rows.
			Kind:                 DeltaMergeUnion,
			PredicateFingerprint: Fingerprint(parents[0]),
			Condition:            Condition(parents[0]),
			ChildFingerprints:    []string{Fingerprint(child)},
		}}
}

// supportingFamilies returns the DISTINCT failure-side families whose eligible
// members all satisfy the predicate (the same member-wise rollup the mining
// engine applies; redundant members were excluded upstream by the reader).
func supportingFamilies(pred Predicate, families []Family) map[string]struct{} {
	out := map[string]struct{}{}
	for _, fam := range families {
		eligible, allSatisfy := 0, true
		for _, m := range fam.Members {
			role, ok := memberRole(fam, m)
			if !ok || role != RoleSupport {
				continue
			}
			eligible++
			if Evaluate(pred, m.Signature) != VerdictSatisfies {
				allSatisfy = false
			}
		}
		if eligible > 0 && allSatisfy {
			out[fam.ClusterID] = struct{}{}
		}
	}
	return out
}

// contrastViolations counts contrast-side families in which the predicate is
// violated by at least one eligible member (the discrimination signal).
func contrastViolations(pred Predicate, families []Family) int {
	n := 0
	for _, fam := range families {
		violated := false
		for _, m := range fam.Members {
			role, ok := memberRole(fam, m)
			if !ok || role != RoleContrast {
				continue
			}
			if Evaluate(pred, m.Signature) == VerdictViolates {
				violated = true
			}
		}
		if violated {
			n++
		}
	}
	return n
}

func sortedKeys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
