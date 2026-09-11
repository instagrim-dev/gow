package main

import "testing"

// findConfig returns the named configuration from the chapter's set.
func findConfig(t *testing.T, name string) Config {
	t.Helper()
	for _, c := range Configurations() {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("no configuration named %q", name)
	return Config{}
}

// runChecked runs a configuration and fails the test if correctness — all
// updates applied, per-stream order preserved — is violated. Tick counts are
// meaningless if the run cheated.
func runChecked(t *testing.T, c Config) Result {
	t.Helper()
	r := Run(c)
	if err := VerifyCorrectness(r); err != nil {
		t.Fatalf("%s: correctness violated: %v", c.Name, err)
	}
	return r
}

// Test 1 — Attempt 1: raising workers from one to four leaves the global
// gate in place, and completion stays at 40 ticks.
func TestAttempt1FourWorkersGlobalGate(t *testing.T) {
	r := runChecked(t, findConfig(t, "Attempt 1 (4 workers)"))
	if r.Ticks != 40 {
		t.Errorf("ticks = %d, want 40", r.Ticks)
	}
}

// Test 2 — Attempt 2: doubling workers to eight changes nothing while the
// gate remains global.
func TestAttempt2EightWorkersGlobalGate(t *testing.T) {
	r := runChecked(t, findConfig(t, "Attempt 2 (8 workers)"))
	if r.Ticks != 40 {
		t.Errorf("ticks = %d, want 40", r.Ticks)
	}
}

// Test 3 — Attempt 3: round-robin scheduling spreads preparation across all
// four streams, but the global gate still admits one commit per tick.
func TestAttempt3RoundRobinGlobalGate(t *testing.T) {
	r := runChecked(t, findConfig(t, "Attempt 3 (round-robin)"))
	if r.Ticks != 40 {
		t.Errorf("ticks = %d, want 40", r.Ticks)
	}
}

// Test 4 — Intervention: identical to Attempt 3 except gate scope, and
// completion drops to 10 ticks. The single-variable discipline is asserted,
// not narrated: the two configs must differ in Gate and in nothing else.
func TestInterventionPerStreamGate(t *testing.T) {
	attempt3 := findConfig(t, "Attempt 3 (round-robin)")
	intervention := findConfig(t, "Intervention (per-stream gate)")

	if intervention.Workers != attempt3.Workers ||
		intervention.Scheduling != attempt3.Scheduling ||
		intervention.Streams != attempt3.Streams ||
		intervention.PerStream != attempt3.PerStream {
		t.Fatalf("intervention must differ from attempt 3 in gate scope only:\n  attempt 3:    %+v\n  intervention: %+v",
			attempt3, intervention)
	}
	if intervention.Gate == attempt3.Gate {
		t.Fatal("intervention does not change the gate scope")
	}

	r := runChecked(t, intervention)
	if r.Ticks != 10 {
		t.Errorf("ticks = %d, want 10", r.Ticks)
	}
}

// Test 5 — Control: with the same intervention but a single 40-update
// stream, the per-stream gate degenerates into a global gate and the gain
// vanishes.
func TestControlSingleStreamRemovesGain(t *testing.T) {
	r := runChecked(t, findConfig(t, "Control (single stream)"))
	if r.Ticks != 40 {
		t.Errorf("ticks = %d, want 40 (the intervention's gain must vanish)", r.Ticks)
	}
}

// Test 6 — The claim itself: ticks = largest number of updates sharing one
// gate scope, for every chapter configuration and for a sweep of worker
// counts at or above the stream count. Correctness holds throughout, and
// runs are deterministic.
func TestGateScopeClaimAndCorrectness(t *testing.T) {
	for _, c := range Configurations() {
		r := runChecked(t, c)
		if want := PredictTicks(c); r.Ticks != want {
			t.Errorf("%s: ticks = %d, prediction = %d", c.Name, r.Ticks, want)
		}

		// Determinism: a second run must land identically.
		r2 := Run(c)
		if r2.Ticks != r.Ticks || len(r2.Applied) != len(r.Applied) {
			t.Errorf("%s: nondeterministic run (ticks %d vs %d)", c.Name, r.Ticks, r2.Ticks)
			continue
		}
		for i := range r.Applied {
			if r.Applied[i] != r2.Applied[i] {
				t.Errorf("%s: nondeterministic apply order at index %d", c.Name, i)
				break
			}
		}
	}

	// Worker count at or above the stream count does not affect the outcome,
	// under either gate scope.
	for _, gate := range []GateScope{GateGlobal, GatePerStream} {
		for workers := 4; workers <= 12; workers++ {
			c := Config{
				Name:       "sweep",
				Workers:    workers,
				Scheduling: SchedRoundRobin,
				Gate:       gate,
				Streams:    4,
				PerStream:  10,
			}
			r := runChecked(t, c)
			if want := PredictTicks(c); r.Ticks != want {
				t.Errorf("gate=%s workers=%d: ticks = %d, prediction = %d",
					gate, workers, r.Ticks, want)
			}
		}
	}
}
