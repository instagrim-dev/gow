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

// ChallengeResult is the deterministic verdict on one proposed challenge.
type ChallengeResult struct {
	Confirmed bool
	Evidence  []ChallengeEvidence
	Detail    string
}

// unconfirmed builds the inert result (KTD-3): persisted for audit, linked to
// no evidence, driving no state transition.
func unconfirmed(reason string) ChallengeResult {
	return ChallengeResult{Confirmed: false, Detail: reason}
}

// AssociationKind is the coarse claim class a candidate asserts, mirroring the
// engine's code-visible AssociationStatus. Its refutation condition differs
// (F3): a `recurring` candidate claims a universal regularity over the failure
// cohort and is falsified by ONE known in-atlas counterexample; a
// `contrast_observed` candidate claims only that the predicate DISCRIMINATES
// outcomes (an association), which a single failure-side violation does not
// refute — it needs the discrimination itself to collapse. An `unknown`
// candidate makes no confirmed regularity claim, so a single counterexample is
// treated as inconclusive, not a falsification.
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
	for _, fam := range families {
		for _, m := range fam.Members {
			role, eligible := memberRole(fam, m)
			if !eligible || role != RoleSupport {
				continue
			}
			if Evaluate(pred, m.Signature) == VerdictViolates {
				ev = append(ev, ChallengeEvidence{
					Kind:        EvidenceCounterexampleMember,
					ClusterID:   fam.ClusterID,
					SignatureID: m.SignatureID,
					Detail:      "failure-side member violates the predicate",
				})
			}
		}
	}
	if len(ev) == 0 {
		return unconfirmed("no known failure-side member violates the predicate")
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
	return ChallengeResult{Confirmed: true, Evidence: ev,
		Detail: fmt.Sprintf("%d known counterexample member(s) refute the universal claim", len(ev))}
}

// VerifySyntheticCounterexample confirms a CONSTRUCTED failed approach violates
// the predicate. The synthetic signature is provider-authored and its claims are
// UNSUPPORTED; confirmation is code-owned and requires Evaluate to return
// `violates`. Crucially, a synthetic construction is a PROPOSAL, not a
// demonstrated construction: its set fields are `unobserved` (see
// syntheticSignature), so `contains(field, X)` on a description that merely
// omits X evaluates to `unknown`, not `violates` — omitting X from a
// description does not prove an admissible failed approach WITHOUT X exists
// (F3). A synthetic therefore weakens only when the queried structure is
// explicitly, exhaustively supplied AND the construction genuinely violates.
func VerifySyntheticCounterexample(pred Predicate, synthetic canon.MechanismSignature) ChallengeResult {
	if Evaluate(pred, synthetic) != VerdictViolates {
		return unconfirmed("synthetic construction does not violate the predicate (an omitted or unobserved field is a proposal, not a demonstrated counterexample)")
	}
	return ChallengeResult{Confirmed: true, Detail: "synthetic construction violates the predicate"}
}

// VerifySuccessPreserving confirms that a success/partial-success family
// PRESERVES the invariant anyway — evidence the predicate does not discriminate
// outcome. Confirmed when every eligible contrast-side member of some family
// satisfies (unknown members disqualify that family: ambiguity is not
// preservation).
func VerifySuccessPreserving(pred Predicate, families []Family) ChallengeResult {
	var ev []ChallengeEvidence
	for _, fam := range families {
		eligible, allSatisfy := 0, true
		for _, m := range fam.Members {
			role, ok := memberRole(fam, m)
			if !ok || role != RoleContrast {
				continue
			}
			eligible++
			if Evaluate(pred, m.Signature) != VerdictSatisfies {
				allSatisfy = false
			}
		}
		if eligible > 0 && allSatisfy {
			ev = append(ev, ChallengeEvidence{
				Kind:      EvidenceSuccessFamily,
				ClusterID: fam.ClusterID,
				Detail:    "success/partial-success family preserves the predicate",
			})
		}
	}
	if len(ev) == 0 {
		return unconfirmed("no success-side family preserves the predicate")
	}
	sort.Slice(ev, func(i, j int) bool { return ev[i].ClusterID < ev[j].ClusterID })
	return ChallengeResult{Confirmed: true, Evidence: ev,
		Detail: fmt.Sprintf("%d success-preserving family(ies)", len(ev))}
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
		r := unconfirmed("support holds under recomputation: " + detail)
		r.Evidence = ev // the recount is still recorded for audit
		return r
	}
	return ChallengeResult{Confirmed: true, Evidence: ev, Detail: detail}
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
		return unconfirmed("a split requires >=2 child predicates")
	}
	parentFP := Fingerprint(parent)
	parentSupport := supportingFamilies(parent, families)
	seen := map[string]struct{}{}
	supports := make([]map[string]struct{}, 0, len(children))
	var ev []ChallengeEvidence
	for i, child := range children {
		if err := AdmitCandidate(child, vocab); err != nil {
			return unconfirmed(fmt.Sprintf("child %d not admissible: %v", i, err))
		}
		fp := Fingerprint(child)
		if fp == parentFP {
			return unconfirmed(fmt.Sprintf("child %d is semantically identical to the parent", i))
		}
		if _, dup := seen[fp]; dup {
			return unconfirmed(fmt.Sprintf("child %d duplicates another child", i))
		}
		seen[fp] = struct{}{}
		sup := supportingFamilies(child, families)
		if len(sup) == 0 {
			return unconfirmed(fmt.Sprintf("child %d grounds to no supporting failure family", i))
		}
		// Refinement: a split partitions the PARENT's support, so every family a
		// child grounds to must be a family the parent also supports. A child
		// grounding to a family outside the parent's support is a new invariant,
		// not a refinement of this one (docs/abstraction-safety.md).
		for fam := range sup {
			if _, ok := parentSupport[fam]; !ok {
				return unconfirmed(fmt.Sprintf("child %d grounds to family %s outside the parent's support (not a refinement)", i, fam))
			}
		}
		supports = append(supports, sup)
	}
	for i := range supports {
		for j := i + 1; j < len(supports); j++ {
			for fam := range supports[i] {
				if _, overlap := supports[j][fam]; overlap {
					return unconfirmed(fmt.Sprintf("children %d and %d share supporting family %s (not disjoint)", i, j, fam))
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
	return ChallengeResult{Confirmed: true, Evidence: ev,
		Detail: fmt.Sprintf("split grounds %d children to disjoint support", len(children))}
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
		return unconfirmed("a merge requires >=2 parent predicates")
	}
	if err := AdmitCandidate(child, vocab); err != nil {
		return unconfirmed(fmt.Sprintf("child not admissible: %v", err))
	}
	childSupport := supportingFamilies(child, families)
	childContrast := contrastViolations(child, families)
	union := map[string]struct{}{}
	for i, p := range parents {
		if err := AdmitCandidate(p, vocab); err != nil {
			return unconfirmed(fmt.Sprintf("parent %d not admissible: %v", i, err))
		}
		for fam := range supportingFamilies(p, families) {
			union[fam] = struct{}{}
		}
		if pc := contrastViolations(p, families); childContrast < pc {
			return unconfirmed(fmt.Sprintf("merge loses discrimination: child contrast %d < parent %d contrast %d", childContrast, i, pc))
		}
	}
	for fam := range union {
		if _, ok := childSupport[fam]; !ok {
			return unconfirmed(fmt.Sprintf("child support does not cover parent-supported family %s", fam))
		}
	}
	var ev []ChallengeEvidence
	for _, fam := range sortedKeys(childSupport) {
		ev = append(ev, ChallengeEvidence{Kind: EvidenceGrounding, ClusterID: fam,
			Detail: "merged child grounds to family " + fam})
	}
	return ChallengeResult{Confirmed: true, Evidence: ev,
		Detail: fmt.Sprintf("merge covers %d united support families without discrimination loss", len(union))}
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
