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
//
// 2026-09-13 external review repairs (ordering policy UNCHANGED — the
// rule ordering emitted for any accepted input is byte-identical to the
// frozen v1; the repairs are to identity binding, claim wording, and
// work metering, and the snapshot/input hashes change to say so):
//
//   - finding 1: probe work is metered on the Decision so the runner can
//     charge it (the probes were previously uncharged pre-search work);
//   - finding 2: the demotion rationale states what the probe actually
//     checked (one-step, from the current start) instead of "can never
//     reduce this task";
//   - finding 3: the input hash binds the actual task expression and the
//     admitted rules' CONTENT, and the procedure refuses inputs whose
//     task disagrees with its declared rendering or whose rule names
//     collide with different content.
const ControllerVersionV1 = "shape-selector/1"

// probeDescription is the frozen probe's identity string, hashed into
// the snapshot. Its wording carries the probe's exact claim scope: a
// one-step check from the current start, not a reachability judgment.
const probeDescription = "single-application strict NodeCount reduction from the current task start (one-step probe; not a multi-step reachability claim)"

// ruleIdentity is the hashed identity view of everything that can change
// the v1 decision: the plain Input, the domain, the task expression's
// canonical rendering, the admitted rules' full content identities
// (rewrite.Rule.Identity: name, admitted domain, both sides), and the
// controller version. Names alone are NOT an identity — the 2026-09-13
// external review (finding 3) showed a name-only hash is blind to task
// or rule-content substitution that flips the probe.
type ruleIdentity struct {
	In      Input
	Domain  finite.Domain
	Task    string   // canonical rendering of the actual probed expression
	Rules   []string // rewrite.Rule.Identity() strings, sorted
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
//     (rewrite.ProbeStrictReduction); a credited rule with no one-step
//     reduction from the current start is demoted below neutral — a
//     declared heuristic, not a proof: the rule may still contribute
//     after a cost-neutral step (add-zero on add(0,x) via add-comm), and
//     the historical success may be genuine regardless of whether its
//     credited rule helps HERE;
//   - a rule seen only in relevant FAILURES is demoted only if it also
//     has no one-step reduction from the current start; a failure-credited
//     rule that demonstrably reduces has its failure claim distrusted and
//     stays neutral.
//
// The probe consults only the task and admitted rules — computable by
// any arm, no history content, no oracle access — so it grounds history
// claims without replacing the history channel: an uncredited reducing
// rule earns nothing and stays in catalog order.
//
// The probe is search work performed before the budgeted search: the
// Decision meters it (ProbeRuleApplications, ProbeCandidates) and the
// runner must charge it to the arm's task-directed ledger.
//
// Deliberately still misleadable: a lying history crediting a rule that
// DOES reduce this task (but wastefully) passes the probe and hurts.
// v1 narrows the lie surface; it does not pretend to close it.
//
// SelectV1 refuses inputs that violate its identity contract: a task
// expression that does not render to Input.TaskStart, an invalid task
// for the declared domain, or duplicate rule names carrying different
// rule content. A refusal is an error, never a silent decision.
func SelectV1(in InputV1) (Decision, error) {
	if defects := finite.ValidateExpr(in.Task, in.Domain); len(defects) > 0 {
		return Decision{}, fmt.Errorf("v1 task expression is not valid in the declared domain: %v", defects)
	}
	if got := finite.Render(in.Task); got != in.TaskStart {
		return Decision{}, fmt.Errorf("v1 identity contract violation: the task expression renders to %q but Input.TaskStart declares %q; the probe would ground history claims against a different task than the one hashed", got, in.TaskStart)
	}

	ruleByName := map[string]rewrite.Rule{}
	ruleIdentities := make([]string, 0, len(in.Rules))
	for _, r := range in.Rules {
		if prev, dup := ruleByName[r.Name()]; dup {
			if prev.Identity() != r.Identity() {
				return Decision{}, fmt.Errorf("v1 identity contract violation: two admitted rules share the name %q with different content (%s vs %s); a name is not a rule identity", r.Name(), prev.Identity(), r.Identity())
			}
			continue // exact duplicate: harmless, keep one
		}
		ruleByName[r.Name()] = r
		ruleIdentities = append(ruleIdentities, r.Identity())
	}
	sort.Strings(ruleIdentities)

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
		ControllerVersion: ControllerVersionV1,
		SnapshotHash: hashOf(struct {
			Version   string
			Threshold int
			Probe     string
		}{ControllerVersionV1, SimilarityThreshold, probeDescription}),
		InputHash: hashOf(ruleIdentity{In: in.Input, Domain: in.Domain, Task: finite.Render(in.Task), Rules: ruleIdentities, Version: ControllerVersionV1}),
	}

	// Task-grounding probe per catalog rule, metered: this is search
	// work (candidate rewrites materialized and costed) done before the
	// budgeted search, and it must appear on the decision so the runner
	// can charge it (2026-09-13 external review finding 1).
	reduces := map[string]bool{}
	for _, name := range catalog {
		r, ok := ruleByName[name]
		if !ok {
			continue // no admitted rule supplied: probe cannot pass, claim stays unverifiable
		}
		red, candidates := rewrite.ProbeStrictReduction(in.Task, r, in.Domain, rewrite.NodeCount)
		reduces[name] = red
		dec.ProbeRuleApplications++
		dec.ProbeCandidates += candidates
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

	// Rationale wording states exactly what the probe checked: "no
	// one-step strict NodeCount decrease from the current start" — never
	// "can never reduce this task", which exceeds the probe's claim
	// scope (finding 2), and never a verdict on whether the historical
	// attempt itself was truthful: current one-step usefulness and
	// historical truth are different claims.
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
					Rationale: fmt.Sprintf("credited by %d relevant completed attempt(s) but no one-step strict NodeCount decrease from the current start; preference withheld by declared heuristic (one-step probe — not a proof the rule can never contribute, nor that the history is false)", len(ev)),
				})
			}
			continue
		}
		if ev, ok := failureSupport[rule]; ok {
			if !reduces[rule] {
				demoted = append(demoted, Preference{
					Rule: rule, Direction: "avoid", Support: len(ev), Evidence: ev,
					Rationale: fmt.Sprintf("appears only in relevant failed attempt(s) (%d) and has no one-step strict NodeCount decrease from the current start", len(ev)),
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
	return dec, nil
}
