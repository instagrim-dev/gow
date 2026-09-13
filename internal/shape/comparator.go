package shape

import (
	"fmt"
	"sort"
)

// ComparatorVersion identifies the frozen H1 comparator procedure; bump on
// any change. It is frozen BEFORE any sealed pack exists, same as the HG
// controller, so pack authorship cannot be conditioned on either arm.
const ComparatorVersion = "h1-frequency/0"

// SelectUngatedFrequency is the H1 arm: a capable deterministic use of
// the same History bytes with NO relevance gate — rules ordered by global
// success frequency (rules seen in any completed attempt, however
// dissimilar), ties and zero-support rules in catalog order. Its
// deliberate weakness relative to HG is exactly the treatment under
// test: HG's only additional privilege is the relevance gate.
//
// Parity discipline: same Input, same Decision shape, same catalog
// dedupe, same identity fields (its own version and snapshot), so the
// two arms are auditable side by side.
func SelectUngatedFrequency(in Input) Decision {
	seenName := map[string]bool{}
	catalog := make([]string, 0, len(in.Catalog))
	for _, r := range in.Catalog {
		if !seenName[r] {
			seenName[r] = true
			catalog = append(catalog, r)
		}
	}
	in.Catalog = catalog

	dec := Decision{
		ControllerVersion: ComparatorVersion,
		SnapshotHash:      hashOf(struct{ Version string }{ComparatorVersion}),
		InputHash:         hashOf(in),
	}

	inCatalog := map[string]bool{}
	for _, r := range catalog {
		inCatalog[r] = true
	}
	support := map[string][]int{}
	for i, att := range in.History {
		if !att.Completed {
			continue
		}
		seen := map[string]bool{}
		for _, rule := range att.RulesApplied {
			if !inCatalog[rule] || seen[rule] {
				continue
			}
			seen[rule] = true
			support[rule] = append(support[rule], i)
		}
	}

	ordered := append([]string(nil), catalog...)
	sort.SliceStable(ordered, func(i, j int) bool { return len(support[ordered[i]]) > len(support[ordered[j]]) })
	dec.EnabledRules = ordered
	for _, rule := range ordered {
		if ev, ok := support[rule]; ok {
			dec.Preferences = append(dec.Preferences, Preference{
				Rule: rule, Direction: "prefer", Support: len(ev), Evidence: ev,
				Rationale: fmt.Sprintf("appears in %d completed attempt(s), relevance not assessed (ungated comparator)", len(ev)),
			})
		}
	}
	return dec
}
