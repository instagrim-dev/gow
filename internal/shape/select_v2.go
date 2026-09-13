package shape

import (
	"fmt"
	"sort"

	"github.com/instagrim-dev/newf/internal/finite"
	"github.com/instagrim-dev/newf/internal/rewrite"
)

// ControllerVersionV2 identifies the failure-aware selector that grounds
// history claims against the task. Its predecessor lineage:
//
//   - shape-selector/0 (v0, frozen): relevance gate only. The
//     2026-09-13 screen run 2 (record 019) fired the H0 guard because
//     that gate admits structurally lookalike LYING histories at full
//     weight.
//   - shape-selector/1 (frozen at f7554cb, recoverable from git): added
//     the task-grounding probe. It produced record 019's run-3 numbers
//     (H0 14 / H1 12 / HG 16) and is the identity those numbers belong
//     to.
//   - shape-selector/2 (this procedure): the 2026-09-13 external review
//     repairs. The RULE ORDERING emitted for any accepted input is
//     unchanged from /1 — run 3 reproduces 16 — but the procedure is not
//     the same procedure: it refuses inputs /1 decided, and its frozen
//     probe parameter changed, so its snapshot hash moved. Keeping the
//     /1 label over changed frozen parameters would have put two
//     procedures behind one identifier, which shape.go's versioning rule
//     ("any change is a new version") forbids. The bump is that rule
//     being obeyed, not a behavioral claim.
//
// The /2 repairs are:
//
//   - finding 1: probe work is metered on the Decision so the runner can
//     charge it (the probes were previously uncharged pre-search work);
//   - finding 2: the demotion rationale states what the probe actually
//     checked (one-step, from the current start) instead of "can never
//     reduce this task";
//   - finding 3: the input hash binds the admitted rules' CONTENT, and
//     the procedure refuses inputs whose task disagrees with its declared
//     rendering or whose rule names collide with different content.
const ControllerVersionV2 = "shape-selector/2"

// probeDescription is the frozen probe's identity string, hashed into
// the snapshot. Its wording carries the probe's exact claim scope: a
// one-step check from the current start, not a reachability judgment.
const probeDescription = "single-application strict NodeCount reduction from the current task start (one-step probe; not a multi-step reachability claim)"

// ruleIdentity is the hashed identity view of everything that can change
// the decision: the plain Input (whose TaskStart is the task's canonical
// rendering — SelectV2 refuses any input where the actual expression
// disagrees, so Input pins the task), the domain, the admitted rules'
// full content identities (rewrite.Rule.Identity: name, admitted domain,
// both sides), and the controller version.
//
// Rule NAMES alone are not an identity: the 2026-09-13 external review
// (finding 3) showed a name-only hash is blind to rule-content
// substitution that flips the probe. The task is bound by the refusal,
// not by a second hashed copy of the same rendering — the self-review of
// that repair found a redundant Task field whose value could never
// diverge from In.TaskStart on any accepted input, so it proved nothing
// and is not carried here.
type ruleIdentity struct {
	In      Input
	Domain  finite.Domain
	Rules   []string // rewrite.Rule.Identity() strings, sorted
	Version string
}

// InputV2 extends Input with what the task-grounding probe needs: the
// task as an expression (Input.TaskStart stays its canonical rendering;
// no parser exists and none is implied), the search domain, and the
// admitted rules matching the catalog names.
type InputV2 struct {
	Input
	Task   finite.Expr
	Domain finite.Domain
	Rules  []rewrite.Rule
}

// SelectV2 is v0's relevance gate plus task-grounded distrust of history
// claims. "Trust, then verify against the task":
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
// SelectV2 refuses inputs that violate its identity contract: a task
// expression that does not render to Input.TaskStart, an invalid task
// for the declared domain, or duplicate rule names carrying different
// rule content. A refusal is an error, never a silent decision.
func SelectV2(in InputV2) (Decision, error) {
	if defects := finite.ValidateExpr(in.Task, in.Domain); len(defects) > 0 {
		return Decision{}, fmt.Errorf("selector task expression is not valid in the declared domain: %v", defects)
	}
	if got := finite.Render(in.Task); got != in.TaskStart {
		return Decision{}, fmt.Errorf("selector identity contract violation: the task expression renders to %q but Input.TaskStart declares %q; the probe would ground history claims against a different task than the one hashed", got, in.TaskStart)
	}

	ruleByName := map[string]rewrite.Rule{}
	ruleIdentities := make([]string, 0, len(in.Rules))
	for _, r := range in.Rules {
		if prev, dup := ruleByName[r.Name()]; dup {
			if prev.Identity() != r.Identity() {
				return Decision{}, fmt.Errorf("selector identity contract violation: two admitted rules share the name %q with different content (%s vs %s); a name is not a rule identity", r.Name(), prev.Identity(), r.Identity())
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
		ControllerVersion: ControllerVersionV2,
		SnapshotHash: hashOf(struct {
			Version   string
			Threshold int
			Probe     string
		}{ControllerVersionV2, SimilarityThreshold, probeDescription}),
		InputHash: hashOf(ruleIdentity{In: in.Input, Domain: in.Domain, Rules: ruleIdentities, Version: ControllerVersionV2}),
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
