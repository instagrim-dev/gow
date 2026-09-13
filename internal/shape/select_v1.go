package shape

import (
	"fmt"
	"sort"

	"github.com/instagrim-dev/newf/internal/finite"
	"github.com/instagrim-dev/newf/internal/rewrite"
)

// ControllerVersionV1 identifies the failure-aware v1 selector. v0
// (shape-selector/0) remains frozen and untouched; this is a successor
// version motivated by the 2026-09-13 screen run 2 (record 019): the
// H0 guard fired because v0's relevance gate admits structurally
// lookalike LYING histories at full weight.
const ControllerVersionV1 = "shape-selector/1"

// RuleSpecView is the hashed identity view of the admitted rules an
// InputV1 carries: names only, because rule semantics are pinned by the
// warrant-admitted menu, not by the selector.
type ruleIdentity struct {
	In      Input
	Domain  finite.Domain
	Rules   []string
	Version string
}

// InputV1 extends Input with what the task-grounding probe needs: the
// task as an expression (Input.TaskStart stays its canonical rendering;
// no parser exists and none is implied), the search domain, and the
// admitted rules matching the catalog names.
type InputV1 struct {
	Input
	Task   finite.Expr
	Domain finite.Domain
	Rules  []rewrite.Rule
}

// SelectV1 is the v1 procedure: v0's relevance gate plus task-grounded
// distrust of history claims. "Trust, then verify against the task":
//
//   - a rule credited by relevant SUCCESSES earns preference only if a
//     single application can strictly reduce the current task's cost
//     (rewrite.CanStrictlyReduce); a credited rule that can never reduce
//     this task is a DISTRUSTED claim and is demoted below neutral;
//   - a rule seen only in relevant FAILURES is avoided only if it also
//     cannot reduce this task; a failure-credited rule that demonstrably
//     reduces has its failure claim distrusted and stays neutral.
//
// The probe consults only the task and admitted rules — computable by
// any arm, no history content, no oracle access — so it grounds history
// claims without replacing the history channel: an uncredited reducing
// rule earns nothing and stays in catalog order.
//
// Deliberately still misleadable: a lying history crediting a rule that
// DOES reduce this task (but wastefully) passes the probe and hurts.
// v1 narrows the lie surface; it does not pretend to close it.
func SelectV1(in InputV1) Decision {
	seenName := map[string]bool{}
	catalog := make([]string, 0, len(in.Catalog))
	for _, r := range in.Catalog {
		if !seenName[r] {
			seenName[r] = true
			catalog = append(catalog, r)
		}
	}
	in.Catalog = catalog

	ruleByName := map[string]rewrite.Rule{}
	ruleNames := make([]string, 0, len(in.Rules))
	for _, r := range in.Rules {
		ruleByName[r.Name()] = r
		ruleNames = append(ruleNames, r.Name())
	}
	sort.Strings(ruleNames)

	dec := Decision{
		ControllerVersion: ControllerVersionV1,
		SnapshotHash: hashOf(struct {
			Version   string
			Threshold int
			Probe     string
		}{ControllerVersionV1, SimilarityThreshold, "single-application strict NodeCount reduction on the task start"}),
		InputHash: hashOf(ruleIdentity{In: in.Input, Domain: in.Domain, Rules: ruleNames, Version: ControllerVersionV1}),
	}

	// Task-grounding probe per catalog rule.
	reduces := map[string]bool{}
	for _, name := range catalog {
		r, ok := ruleByName[name]
		if !ok {
			continue // no admitted rule supplied: probe cannot pass, claim stays unverifiable
		}
		reduces[name] = rewrite.CanStrictlyReduce(in.Task, r, in.Domain, rewrite.NodeCount)
	}

	taskFeatures := operatorMultiset(in.TaskStart)
	relevant := make([]int, 0, len(in.History))
	for i, att := range in.History {
		if similarity(taskFeatures, operatorMultiset(att.Start)) >= SimilarityThreshold {
			relevant = append(relevant, i)
		}
	}
	dec.RelevantAttempts = relevant

	inCatalog := map[string]bool{}
	for _, r := range catalog {
		inCatalog[r] = true
	}
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

	var preferred, demoted []Preference
	for _, rule := range catalog {
		if ev, ok := successSupport[rule]; ok {
			if reduces[rule] {
				preferred = append(preferred, Preference{
					Rule: rule, Direction: "prefer", Support: len(ev), Evidence: ev,
					Rationale: fmt.Sprintf("credited by %d relevant completed attempt(s) AND a single application strictly reduces this task", len(ev)),
				})
			} else {
				demoted = append(demoted, Preference{
					Rule: rule, Direction: "avoid", Support: len(ev), Evidence: ev,
					Rationale: fmt.Sprintf("credited by %d relevant completed attempt(s) but can never reduce this task; the success claim is distrusted", len(ev)),
				})
			}
			continue
		}
		if ev, ok := failureSupport[rule]; ok {
			if !reduces[rule] {
				demoted = append(demoted, Preference{
					Rule: rule, Direction: "avoid", Support: len(ev), Evidence: ev,
					Rationale: fmt.Sprintf("appears only in relevant failed attempt(s) (%d) and cannot reduce this task", len(ev)),
				})
			}
			// else: failure claim distrusted (the rule demonstrably
			// reduces); stays neutral, no preference recorded.
		}
	}
	sort.SliceStable(preferred, func(i, j int) bool { return preferred[i].Support > preferred[j].Support })

	inPref := map[string]bool{}
	for _, p := range preferred {
		inPref[p.Rule] = true
	}
	inDem := map[string]bool{}
	for _, p := range demoted {
		inDem[p.Rule] = true
	}
	ordered := make([]string, 0, len(catalog))
	for _, p := range preferred {
		ordered = append(ordered, p.Rule)
	}
	for _, rule := range catalog {
		if !inPref[rule] && !inDem[rule] {
			ordered = append(ordered, rule)
		}
	}
	for _, rule := range catalog {
		if inDem[rule] {
			ordered = append(ordered, rule)
		}
	}
	dec.EnabledRules = ordered
	dec.Preferences = append(preferred, demoted...)
	return dec
}
