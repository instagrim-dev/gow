package shape

import (
	"fmt"
	"sort"
)

// Select is the v0 shaping procedure. Deterministic: identical Input
// yields a byte-identical Decision. It emits the full enabled-rule
// ordering (preferred, then neutral in catalog order, then avoided) with
// per-preference evidence references. No rule is ever removed: demotion
// is ordering, not censorship, so the search stays complete under an
// unlimited budget and the selector's effect exists only under scarcity.
func Select(in Input) Decision {
	// Dedupe the catalog on entry, first occurrence wins: duplicate
	// names would double preferences and the enabled ordering
	// (adversarial review finding 4).
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
		ControllerVersion: ControllerVersion,
		SnapshotHash:      SnapshotHash(),
		InputHash:         hashOf(in),
	}

	taskFeatures := operatorMultiset(in.TaskStart)

	// Relevance gate (frozen threshold).
	relevant := make([]int, 0, len(in.History))
	for i, att := range in.History {
		if similarity(taskFeatures, operatorMultiset(att.Start)) >= SimilarityThreshold {
			relevant = append(relevant, i)
		}
	}
	dec.RelevantAttempts = relevant

	inCatalog := map[string]bool{}
	for _, r := range in.Catalog {
		inCatalog[r] = true
	}

	// Support counts over relevant attempts only. A rule outside the
	// admitted catalog earns nothing: preferences reorder admitted
	// rules, they never unlock.
	successSupport := map[string][]int{}
	failureSupport := map[string][]int{}
	for _, i := range relevant {
		att := in.History[i]
		seen := map[string]bool{}
		for _, rule := range att.RulesApplied {
			if !inCatalog[rule] || seen[rule] {
				continue
			}
			seen[rule] = true
			if att.Completed {
				successSupport[rule] = append(successSupport[rule], i)
			} else {
				failureSupport[rule] = append(failureSupport[rule], i)
			}
		}
	}

	var preferred, avoided []Preference
	for _, rule := range in.Catalog {
		if ev, ok := successSupport[rule]; ok {
			preferred = append(preferred, Preference{
				Rule: rule, Direction: "prefer", Support: len(ev), Evidence: ev,
				Rationale: fmt.Sprintf("appears in %d relevant completed attempt(s)", len(ev)),
			})
			continue
		}
		if ev, ok := failureSupport[rule]; ok {
			avoided = append(avoided, Preference{
				Rule: rule, Direction: "avoid", Support: len(ev), Evidence: ev,
				Rationale: fmt.Sprintf("appears only in relevant failed attempt(s) (%d), never in a relevant success", len(ev)),
			})
		}
	}
	// Preferred: by support desc, then catalog order (stable sort over
	// the catalog-ordered slice keeps ties deterministic).
	sort.SliceStable(preferred, func(i, j int) bool { return preferred[i].Support > preferred[j].Support })

	prefSet := map[string]bool{}
	for _, p := range preferred {
		prefSet[p.Rule] = true
	}
	avoidSet := map[string]bool{}
	for _, p := range avoided {
		avoidSet[p.Rule] = true
	}

	ordered := make([]string, 0, len(in.Catalog))
	for _, p := range preferred {
		ordered = append(ordered, p.Rule)
	}
	for _, rule := range in.Catalog { // neutral residue keeps catalog order
		if !prefSet[rule] && !avoidSet[rule] {
			ordered = append(ordered, rule)
		}
	}
	for _, rule := range in.Catalog { // avoided rules last, catalog order
		if avoidSet[rule] {
			ordered = append(ordered, rule)
		}
	}

	dec.EnabledRules = ordered
	dec.Preferences = append(preferred, avoided...)
	return dec
}
