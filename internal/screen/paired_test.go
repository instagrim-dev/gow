package screen

import (
	"reflect"
	"testing"
)

func TestPairedReversalsSurviveEpisodeAggregateTie(t *testing.T) {
	d := Design{Episodes: []Episode{{ID: "control", Family: "one", Stratum: StratumMisleading}}, RunsPerCell: 3, EvidenceLabel: "development"}
	var execs []Execution
	for _, arm := range allArms {
		for run := 1; run <= 3; run++ {
			execs = append(execs, Execution{Arm: arm, EpisodeID: "control", Run: run,
				Completed: (arm == ArmH1 && run == 1) || (arm == ArmHG && run == 2)})
		}
	}
	out, err := Evaluate(d, execs)
	if err != nil {
		t.Fatal(err)
	}
	want := ControlDiagnostics{PairedExecutionWins: 1, PairedExecutionLosses: 1}
	if out.Controls != want {
		t.Fatalf("paired reversal lost under aggregate tie: got %+v, want %+v", out.Controls, want)
	}
	c := condition(t, out, "c")
	if c.Left != out.Controls.PairedExecutionLosses-out.Controls.PairedExecutionWins || !c.Satisfied {
		t.Fatalf("paired counts must reconcile with unchanged net criterion: %+v", c)
	}
	for i, j := 0, len(execs)-1; i < j; i, j = i+1, j-1 {
		execs[i], execs[j] = execs[j], execs[i]
	}
	reversed, err := Evaluate(d, execs)
	if err != nil || !reflect.DeepEqual(out, reversed) {
		t.Fatalf("pairing must use episode and repetition, not slice order: %v", err)
	}
}

func TestSingleRunDiagnosticUnitsCoincide(t *testing.T) {
	d := Design{Episodes: []Episode{{ID: "control", Family: "one", Stratum: StratumLowValue}}, RunsPerCell: 1}
	var execs []Execution
	for _, arm := range allArms {
		execs = append(execs, Execution{Arm: arm, EpisodeID: "control", Run: 1, Completed: arm == ArmHG})
	}
	out, err := Evaluate(d, execs)
	if err != nil {
		t.Fatal(err)
	}
	if out.Controls != (ControlDiagnostics{EpisodeAggregateWins: 1, PairedExecutionWins: 1}) {
		t.Fatalf("r=1 must preserve the old episode counts: %+v", out.Controls)
	}
}
