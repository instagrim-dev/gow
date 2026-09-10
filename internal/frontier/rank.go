package frontier

import "sort"

// Rank orders a run's candidates by the ordinal objective (AGENTS.md /
// docs/toolbox-dsl.md):
//
//	maximize( mechanistic_distance
//	        + invariant_violation
//	        + expected_information_gain
//	        - evaluation_cost
//	        - redundancy )
//
// The comparison is LEXICOGRAPHIC over ordinal components — never a fabricated
// weighted float (AGENTS.md: "Do not fabricate fake precision"). The order is
// documented and deterministic:
//  1. code-confirmed violation of a target (a proposal that actually breaks a
//     surviving invariant outranks one that only claims to);
//  2. mechanistic distance from known failures (more distant = more novel);
//  3. expected information gain (provider ordinal);
//  4. lower evaluation cost;
//  5. down-rank near-duplicates: proposals sharing the same (nearest family set +
//     target set) are ordered after the first such proposal (redundancy penalty);
//  6. proposal hash, as a total-order tiebreak for determinism.
//
// Rank mutates and returns the slice for convenience.
func Rank(candidates []Candidate) []Candidate {
	// Compute a redundancy signature per candidate and a first-seen order so
	// later duplicates of the same directed attack sort lower.
	firstSeen := map[string]int{}
	redundancyRank := make([]int, len(candidates))
	for i := range candidates {
		key := redundancyKey(candidates[i])
		if _, ok := firstSeen[key]; !ok {
			firstSeen[key] = i
		}
		redundancyRank[i] = firstSeen[key]
	}

	idx := make([]int, len(candidates))
	for i := range idx {
		idx[i] = i
	}
	sort.SliceStable(idx, func(a, b int) bool {
		ca, cb := candidates[idx[a]], candidates[idx[b]]
		// 1. confirmed violation first.
		if ca.ViolatesAnyTarget != cb.ViolatesAnyTarget {
			return ca.ViolatesAnyTarget
		}
		// 2. mechanistic distance (desc).
		if ca.MechanisticDistance.Rank() != cb.MechanisticDistance.Rank() {
			return ca.MechanisticDistance.Rank() > cb.MechanisticDistance.Rank()
		}
		// 3. expected information gain (desc).
		if ca.ExpectedInformationGain.Rank() != cb.ExpectedInformationGain.Rank() {
			return ca.ExpectedInformationGain.Rank() > cb.ExpectedInformationGain.Rank()
		}
		// 4. evaluation cost (asc — cheaper to kill is better).
		if ca.EvaluationCost.Rank() != cb.EvaluationCost.Rank() {
			return ca.EvaluationCost.Rank() < cb.EvaluationCost.Rank()
		}
		// 5. redundancy: earlier first-seen of the same directed attack ranks higher.
		if redundancyRank[idx[a]] != redundancyRank[idx[b]] {
			return redundancyRank[idx[a]] < redundancyRank[idx[b]]
		}
		// 6. deterministic tiebreak.
		return ca.ProposalHash < cb.ProposalHash
	})

	ranked := make([]Candidate, len(candidates))
	for i, j := range idx {
		ranked[i] = candidates[j]
	}
	return ranked
}

// RedundancyKey exposes the directed-attack redundancy key for a candidate so
// other packages (e.g. search-policy application) can key penalize directives
// on the same notion of "same directed attack" that Rank uses internally.
func RedundancyKey(c Candidate) string {
	return redundancyKey(c)
}

// redundancyKey identifies proposals that are the "same directed attack": the
// same set of targeted invariants against the same nearest family set. Two such
// proposals differ only cosmetically for search purposes and the second is
// down-ranked.
func redundancyKey(c Candidate) string {
	key := "t:"
	for _, id := range c.TargetInvariantIDs { // already sorted+deduped
		key += id + ","
	}
	key += "|n:"
	ids := make([]string, 0, len(c.NearestClusters))
	for _, nc := range c.NearestClusters {
		ids = append(ids, nc.ClusterID)
	}
	sort.Strings(ids)
	for _, id := range ids {
		key += id + ","
	}
	return key
}
