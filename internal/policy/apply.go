package policy

import (
	"sort"

	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/frontier"
	"github.com/instagrim-dev/newf/internal/invariant"
)

// AppliedBias is the per-candidate record of how a policy reordered it: the net
// ordinal nudge (positive = promoted, negative = suppressed) and which directive
// kinds fired. This is the queryable "why was this proposal favored or
// suppressed" surface (KTD-5); it is logged, never implicit.
type AppliedBias struct {
	ProposalHash    string
	Net             int // net signed nudge applied within the violation tier
	Preferred       bool
	Avoided         bool
	Penalized       bool
	FloorProtected  bool // suppression was floored to preserve the falsification surface
	FiredDirectives []Directive
}

// Result is the biased ordering plus the per-candidate applied-bias log.
type Result struct {
	Ranked []frontier.Candidate
	Bias   []AppliedBias
}

// Apply re-ranks already-Ranked candidates under a policy. It is a BOUNDED
// ordinal transform (KTD-2): the code-verified violation gate is inviolable — a
// non-violating proposal can never outrank a violating one no matter how
// strongly preferred — and every candidate is retained (suppression is a rank
// penalty, never a drop). For each targeted invariant the cheapest-falsification
// violating proposal is FLOOR-PROTECTED so policy can never render a run
// unfalsifiable (EPIC delivery principle 9).
//
// satisfiedPrefer maps a candidate's ProposalHash to the set of preferred
// success-invariant predicate fingerprints its proposed signature satisfies.
// The caller (pipeline) computes this by evaluating the success predicates
// against each candidate's signature in CODE (ModelJudgment != Verification);
// this pure package never trusts a provider claim of satisfaction. A nil map
// means no preference matches.
//
// Determinism: identical (policy, candidates, satisfiedPrefer) ⇒ identical
// output. An empty policy is an exact identity over ordering (KTD-6).
func Apply(pol SearchPolicy, candidates []frontier.Candidate, satisfiedPrefer map[string]map[string]bool) Result {
	res := Result{Ranked: append([]frontier.Candidate(nil), candidates...)}
	bias := make(map[string]*AppliedBias, len(candidates))
	for i := range res.Ranked {
		ab := &AppliedBias{ProposalHash: res.Ranked[i].ProposalHash}
		bias[res.Ranked[i].ProposalHash] = ab
		res.Bias = append(res.Bias, AppliedBias{}) // placeholder; filled after sort
	}

	if len(pol.Directives) == 0 {
		for i := range res.Ranked {
			res.Bias[i] = AppliedBias{ProposalHash: res.Ranked[i].ProposalHash}
		}
		return res
	}

	// Index directives by (kind, target) for O(1) lookup.
	preferFP := map[string]Directive{}
	avoidInv := map[string]Directive{}
	penalizeAttack := map[string]Directive{}
	for _, d := range pol.Directives {
		switch {
		case d.Kind == KindPrefer && d.TargetKind == TargetSuccessInvariant:
			preferFP[d.TargetID] = d
		case d.Kind == KindAvoid && d.TargetKind == TargetSurvivingInvariant:
			avoidInv[d.TargetID] = d
		case d.Kind == KindPenalize && d.TargetKind == TargetRedundantAttack:
			penalizeAttack[d.TargetID] = d
		}
		// KindExpand and KindPenalize/TargetRepeatedFailure bias GENERATION
		// (which families / mechanisms to draw from), not post-hoc ranking of an
		// already-generated set; they are carried in the persisted policy for the
		// generation-request path and do not fire here.
	}

	floor := computeFalsifiabilityFloor(candidates)

	for i := range res.Ranked {
		c := &res.Ranked[i]
		ab := bias[c.ProposalHash]

		// prefer: proposed signature satisfies a preferred success condition
		// (code-verified by the caller) ⇒ pull up.
		for fp := range satisfiedPrefer[c.ProposalHash] {
			if d, ok := preferFP[fp]; ok {
				ab.Net += weightMagnitude(d.Weight)
				ab.Preferred = true
				ab.FiredDirectives = append(ab.FiredDirectives, d)
			}
		}
		// avoid: proposal still SATISFIES (does not violate) an avoided surviving
		// invariant ⇒ re-enters known failure structure ⇒ suppress.
		for _, vc := range c.ViolationChecks {
			if d, ok := avoidInv[vc.InvariantID]; ok && vc.Verdict == invariant.VerdictSatisfies {
				ab.Net -= weightMagnitude(d.Weight)
				ab.Avoided = true
				ab.FiredDirectives = append(ab.FiredDirectives, d)
			}
		}
		// penalize: same directed attack as a redundant one ⇒ suppress.
		if d, ok := penalizeAttack[frontier.RedundancyKey(*c)]; ok {
			ab.Net -= weightMagnitude(d.Weight)
			ab.Penalized = true
			ab.FiredDirectives = append(ab.FiredDirectives, d)
		}

		// Falsifiability floor: a floor-protected candidate never ends net-negative.
		if floor[c.ProposalHash] && ab.Net < 0 {
			ab.Net = 0
			ab.FloorProtected = true
		}
	}

	sort.SliceStable(res.Ranked, func(a, b int) bool {
		ca, cb := res.Ranked[a], res.Ranked[b]
		if ca.ViolatesAnyTarget != cb.ViolatesAnyTarget {
			return ca.ViolatesAnyTarget // violation gate: inviolable
		}
		na, nb := bias[ca.ProposalHash].Net, bias[cb.ProposalHash].Net
		if na != nb {
			return na > nb
		}
		return baseObjectiveLess(ca, cb)
	})

	for i := range res.Ranked {
		res.Bias[i] = *bias[res.Ranked[i].ProposalHash]
	}
	return res
}

// baseObjectiveLess is the tiebreak that preserves the pre-policy frontier
// objective within a (violation, net-bias) band: mechanistic distance desc,
// expected info gain desc, evaluation cost asc, then hash.
func baseObjectiveLess(ca, cb frontier.Candidate) bool {
	if ca.MechanisticDistance.Rank() != cb.MechanisticDistance.Rank() {
		return ca.MechanisticDistance.Rank() > cb.MechanisticDistance.Rank()
	}
	if ca.ExpectedInformationGain.Rank() != cb.ExpectedInformationGain.Rank() {
		return ca.ExpectedInformationGain.Rank() > cb.ExpectedInformationGain.Rank()
	}
	if ca.EvaluationCost.Rank() != cb.EvaluationCost.Rank() {
		return ca.EvaluationCost.Rank() < cb.EvaluationCost.Rank()
	}
	return ca.ProposalHash < cb.ProposalHash
}

// computeFalsifiabilityFloor selects, for each targeted invariant that has at
// least one confirmed-violating candidate, the cheapest-to-falsify violating
// candidate and marks it protected.
func computeFalsifiabilityFloor(candidates []frontier.Candidate) map[string]bool {
	type best struct {
		hash string
		cost int
	}
	byTarget := map[string]best{}
	for _, c := range candidates {
		if !c.ViolatesAnyTarget {
			continue
		}
		cost := c.EvaluationCost.Rank()
		if cost == 0 {
			cost = 99 // unknown cost sorts as most-expensive so a known-cheap wins
		}
		for _, vc := range c.ViolationChecks {
			if !vc.Violated {
				continue
			}
			cur, ok := byTarget[vc.InvariantID]
			if !ok || cost < cur.cost || (cost == cur.cost && c.ProposalHash < cur.hash) {
				byTarget[vc.InvariantID] = best{hash: c.ProposalHash, cost: cost}
			}
		}
	}
	protected := map[string]bool{}
	for _, b := range byTarget {
		protected[b.hash] = true
	}
	return protected
}

// weightMagnitude maps an ordinal weight to a small, bounded integer nudge so
// bias reorders WITHIN the violation tier without ever dominating the
// code-verified gate.
func weightMagnitude(o domain.Ordinal) int {
	switch o {
	case domain.OrdinalHigh:
		return 3
	case domain.OrdinalMedium:
		return 2
	case domain.OrdinalLow:
		return 1
	default:
		return 0
	}
}
