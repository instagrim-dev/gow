package projection

import (
	"testing"
)

func artifact(givens []string, target []string, steps ...Step) Artifact {
	return Artifact{Schema: SchemaProjectionV1, Givens: givens, Steps: steps, Target: target}
}

func TestCheckCompositionComposes(t *testing.T) {
	a := artifact(
		[]string{"congruence-lattice"},
		[]string{"crossing-bound"},
		Step{Name: "lift", Statement: "lift to the covering system", Requires: []string{"congruence-lattice"}, Provides: []string{"cover"}},
		Step{Name: "bound", Statement: "bound crossings over the cover", Requires: []string{"cover"}, Provides: []string{"crossing-bound"}},
	)
	res := CheckComposition(a)
	if !res.OK || len(res.Gaps) != 0 {
		t.Fatalf("expected composition, got %+v", res)
	}
}

// The reviewer's demanded failure mode: a proposed abstract path whose
// concrete steps cannot compose is refuted deterministically, with the exact
// missing tokens named — before any domain work.
func TestCheckCompositionNamesExactGaps(t *testing.T) {
	a := artifact(
		nil,
		[]string{"final-bound"},
		Step{Name: "lift", Statement: "lift", Provides: []string{"cover"}},
		// requires a token NOTHING provides, plus one only a LATER step provides
		// (order matters: later steps cannot supply earlier ones).
		Step{Name: "glue", Statement: "glue", Requires: []string{"cover", "uniformity", "descent-datum"}},
		Step{Name: "descend", Statement: "descend", Provides: []string{"descent-datum"}},
	)
	res := CheckComposition(a)
	if res.OK || len(res.Gaps) != 2 {
		t.Fatalf("expected 2 gaps (glue + target), got %+v", res)
	}
	g := res.Gaps[0]
	if g.StepName != "glue" || g.StepIndex != 1 || len(g.Missing) != 2 ||
		g.Missing[0] != "descent-datum" || g.Missing[1] != "uniformity" {
		t.Fatalf("gap must name the failing step and its sorted missing tokens: %+v", g)
	}
	if res.Gaps[1].StepName != "target" || res.Gaps[1].StepIndex != -1 || res.Gaps[1].Missing[0] != "final-bound" {
		t.Fatalf("unmet target must be a gap: %+v", res.Gaps[1])
	}
}

func TestParseRejectsProseOnlyAndDuplicates(t *testing.T) {
	if _, err := Parse([]byte(`{"schema":"projection/v1","steps":[]}`)); err == nil {
		t.Fatal("a plan with no steps must be rejected as prose")
	}
	if _, err := Parse([]byte(`{"schema":"projection/v1","steps":[
		{"name":"a","statement":"s"},{"name":"a","statement":"s2"}]}`)); err == nil {
		t.Fatal("duplicate step names must be rejected")
	}
	if _, err := Parse([]byte(`{"schema":"projection/v2","steps":[{"name":"a","statement":"s"}]}`)); err == nil {
		t.Fatal("unknown schema must be rejected")
	}
	if _, err := Parse([]byte(`{"schema":"projection/v1","surprise":1,"steps":[{"name":"a","statement":"s"}]}`)); err == nil {
		t.Fatal("unknown fields must be rejected")
	}
}
