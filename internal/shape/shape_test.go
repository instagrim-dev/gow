package shape

import (
	"reflect"
	"testing"
)

func catalog() []string { return []string{"double-not", "add-zero", "xor-self-zero"} }

// A history-free input yields the catalog order unchanged: no evidence,
// no reordering, deterministic hashes present.
func TestNoHistoryYieldsCatalogOrder(t *testing.T) {
	dec := Select(Input{TaskStart: "not(not(x))", Target: 1, Catalog: catalog()})
	if !reflect.DeepEqual(dec.EnabledRules, catalog()) {
		t.Fatalf("no evidence must mean no reordering: %v", dec.EnabledRules)
	}
	if len(dec.Preferences) != 0 || len(dec.RelevantAttempts) != 0 {
		t.Fatalf("no history can justify preferences: %+v", dec)
	}
	if dec.SnapshotHash == "" || dec.InputHash == "" || dec.ControllerVersion != ControllerVersion {
		t.Fatalf("identity fields missing: %+v", dec)
	}
}

// Identical inputs yield byte-identical decisions.
func TestSelectIsDeterministic(t *testing.T) {
	in := Input{
		TaskStart: "xor(not(not(x)), not(not(x)))",
		Target:    1,
		Catalog:   catalog(),
		History: History{
			{Start: "xor(not(not(y)), not(not(y)))", RulesApplied: []string{"xor-self-zero"}, FinalCost: 1, Target: 1, Completed: true, Endpoint: "HOLDS_ON_DECLARED_DOMAIN"},
		},
	}
	a, b := Select(in), Select(in)
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("nondeterministic decision:\n%+v\n%+v", a, b)
	}
}

// Relevant completed attempts promote their rules; rules seen only in
// relevant failures are demoted to the end; every preference carries its
// evidence indices.
func TestPreferAndAvoidWithEvidence(t *testing.T) {
	in := Input{
		TaskStart: "xor(not(not(x)), not(not(x)))",
		Target:    1,
		Catalog:   catalog(),
		History: History{
			{Start: "xor(not(not(y)), not(not(y)))", RulesApplied: []string{"xor-self-zero"}, Completed: true},
			{Start: "xor(not(not(y)), not(not(y)))", RulesApplied: []string{"add-zero"}, Completed: false},
		},
	}
	dec := Select(in)
	if dec.EnabledRules[0] != "xor-self-zero" {
		t.Fatalf("supported success rule must lead: %v", dec.EnabledRules)
	}
	if dec.EnabledRules[len(dec.EnabledRules)-1] != "add-zero" {
		t.Fatalf("failure-only rule must trail: %v", dec.EnabledRules)
	}
	for _, p := range dec.Preferences {
		if len(p.Evidence) == 0 || p.Rationale == "" {
			t.Fatalf("preference without evidence or rationale: %+v", p)
		}
	}
}

// Irrelevant history (below the similarity threshold) biases nothing:
// the gate is real, not decorative.
func TestIrrelevantHistoryBiasesNothing(t *testing.T) {
	in := Input{
		TaskStart: "xor(not(not(x)), not(not(x)))", // xor+not heavy
		Target:    1,
		Catalog:   catalog(),
		History: History{
			{Start: "mul(add(a, b), sub(a, b))", RulesApplied: []string{"add-zero"}, Completed: true},
		},
	}
	dec := Select(in)
	if len(dec.RelevantAttempts) != 0 || len(dec.Preferences) != 0 {
		t.Fatalf("dissimilar history must be gated out: %+v", dec)
	}
	if !reflect.DeepEqual(dec.EnabledRules, catalog()) {
		t.Fatalf("ordering must be unchanged: %v", dec.EnabledRules)
	}
}

// A rule outside the admitted catalog earns nothing: preferences reorder,
// they never unlock.
func TestHistoryCannotUnlockUnadmittedRules(t *testing.T) {
	in := Input{
		TaskStart: "not(not(x))",
		Target:    1,
		Catalog:   []string{"double-not"},
		History: History{
			{Start: "not(not(y))", RulesApplied: []string{"forbidden-rule", "double-not"}, Completed: true},
		},
	}
	dec := Select(in)
	for _, r := range dec.EnabledRules {
		if r == "forbidden-rule" {
			t.Fatal("an unadmitted rule appeared in the enabled ordering")
		}
	}
	if len(dec.EnabledRules) != 1 {
		t.Fatalf("only the admitted catalog may be ordered: %v", dec.EnabledRules)
	}
}

// Demotion is ordering, not censorship: every catalog rule appears
// exactly once in the enabled ordering, whatever the history says.
func TestNoRuleIsEverRemoved(t *testing.T) {
	in := Input{
		TaskStart: "xor(not(not(x)), not(not(x)))",
		Target:    1,
		Catalog:   catalog(),
		History: History{
			{Start: "xor(not(not(y)), not(not(y)))", RulesApplied: []string{"add-zero", "double-not", "xor-self-zero"}, Completed: false},
		},
	}
	dec := Select(in)
	seen := map[string]int{}
	for _, r := range dec.EnabledRules {
		seen[r]++
	}
	for _, r := range catalog() {
		if seen[r] != 1 {
			t.Fatalf("rule %q appears %d times; demotion must not censor", r, seen[r])
		}
	}
}
