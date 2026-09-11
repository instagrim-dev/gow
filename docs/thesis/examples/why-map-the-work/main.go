// Command why-map-the-work is a standalone, deterministic teaching model for
// the chapter docs/thesis/why-map-the-work.md.
//
// It is a synthetic tick-based simulation — not a hardware benchmark and not
// a concurrency discovery. Forty updates are spread across independent
// streams; updates within a stream must apply in order. Each tick, workers
// prepare stream heads and a commit gate admits prepared updates. The only
// structural variable that governs completion time is the gate's scope
// (global vs per-stream).
//
// Run it to print the results table used in the chapter:
//
//	go run ./docs/thesis/examples/why-map-the-work
//
// The table's numbers are checked by the tests in this package.
package main

import (
	"fmt"
	"os"
)

// GateScope determines how many commits the gate admits per tick.
type GateScope int

const (
	// GateGlobal admits at most one commit per tick across the whole system.
	GateGlobal GateScope = iota
	// GatePerStream admits at most one commit per stream per tick.
	GatePerStream
)

func (g GateScope) String() string {
	switch g {
	case GateGlobal:
		return "global"
	case GatePerStream:
		return "per-stream"
	default:
		panic(fmt.Sprintf("unknown GateScope %d", int(g)))
	}
}

// Scheduling determines which stream each worker targets on a tick.
type Scheduling int

const (
	// SchedNaive sends every worker to the lowest-index stream with pending
	// work. Redundant preparation is wasted effort.
	SchedNaive Scheduling = iota
	// SchedRoundRobin spreads workers across streams: worker i targets
	// stream (i mod S), falling forward to the next nonempty stream.
	SchedRoundRobin
)

func (s Scheduling) String() string {
	switch s {
	case SchedNaive:
		return "naive"
	case SchedRoundRobin:
		return "round-robin"
	default:
		panic(fmt.Sprintf("unknown Scheduling %d", int(s)))
	}
}

// Update is one unit of work: the seq-th update of stream Stream.
type Update struct {
	Stream int
	Seq    int
}

// Config is one fully specified attempt at the problem.
type Config struct {
	Name       string
	Workers    int
	Scheduling Scheduling
	Gate       GateScope
	Streams    int // number of independent streams
	PerStream  int // updates per stream
	Predicted  int // ticks predicted by the gate-scope claim
}

// Result is the measured outcome of running a Config, with enough recorded
// state to verify correctness independently of the tick count.
type Result struct {
	Config  Config
	Ticks   int
	Applied []Update // global apply order, one entry per committed update
}

// TotalUpdates is Streams × PerStream.
func (c Config) TotalUpdates() int { return c.Streams * c.PerStream }

// PredictTicks states the gate-scope claim as a computation: completion time
// equals the largest number of updates sharing a single gate scope.
func PredictTicks(c Config) int {
	switch c.Gate {
	case GateGlobal:
		return c.TotalUpdates()
	case GatePerStream:
		return c.PerStream // largest per-scope load; streams are even here
	default:
		panic(fmt.Sprintf("unknown GateScope %d", int(c.Gate)))
	}
}

// Run executes the deterministic tick model until all updates are applied.
func Run(c Config) Result {
	if c.Workers < 1 || c.Streams < 1 || c.PerStream < 1 {
		panic("Run: workers, streams, and per-stream update count must be >= 1")
	}

	// next[s] is the sequence number of the next unapplied update in stream s.
	next := make([]int, c.Streams)
	applied := make([]Update, 0, c.TotalUpdates())
	remaining := c.TotalUpdates()

	ticks := 0
	for remaining > 0 {
		ticks++

		// Phase 1: workers prepare stream heads. Only the head of a stream
		// can be prepared, because in-stream order must be preserved.
		prepared := make([]bool, c.Streams)
		for w := 0; w < c.Workers; w++ {
			s := pickStream(c, w, next)
			if s >= 0 {
				prepared[s] = true // duplicate preparation is wasted, not extra
			}
		}

		// Phase 2: the commit gate admits prepared updates.
		admitted := 0
		for s := 0; s < c.Streams; s++ {
			if !prepared[s] {
				continue
			}
			if c.Gate == GateGlobal && admitted >= 1 {
				continue // global scope: one commit per tick, system-wide
			}
			applied = append(applied, Update{Stream: s, Seq: next[s]})
			next[s]++
			remaining--
			admitted++
		}
	}

	return Result{Config: c, Ticks: ticks, Applied: applied}
}

// pickStream returns the stream worker w targets this tick, or -1 if no
// stream has pending work.
func pickStream(c Config, w int, next []int) int {
	start := 0
	if c.Scheduling == SchedRoundRobin {
		start = w % c.Streams
	}
	for i := 0; i < c.Streams; i++ {
		s := (start + i) % c.Streams
		if next[s] < c.PerStream {
			return s
		}
	}
	return -1
}

// VerifyCorrectness checks the two conditions every configuration must
// preserve: all updates applied exactly once, and each stream's applied
// order equal to its input order. It returns an error describing the first
// violation, or nil.
func VerifyCorrectness(r Result) error {
	if got, want := len(r.Applied), r.Config.TotalUpdates(); got != want {
		return fmt.Errorf("applied %d updates, want %d", got, want)
	}
	seen := make([]int, r.Config.Streams) // next expected seq per stream
	for i, u := range r.Applied {
		if u.Stream < 0 || u.Stream >= r.Config.Streams {
			return fmt.Errorf("applied[%d]: stream %d out of range", i, u.Stream)
		}
		if u.Seq != seen[u.Stream] {
			return fmt.Errorf("applied[%d]: stream %d got seq %d, want %d (order violated)",
				i, u.Stream, u.Seq, seen[u.Stream])
		}
		seen[u.Stream]++
	}
	for s, n := range seen {
		if n != r.Config.PerStream {
			return fmt.Errorf("stream %d: applied %d updates, want %d", s, n, r.Config.PerStream)
		}
	}
	return nil
}

// Configurations returns the five configurations discussed in the chapter,
// with predictions filled in from the gate-scope claim.
func Configurations() []Config {
	base := Config{Streams: 4, PerStream: 10}
	configs := []Config{
		{Name: "Attempt 1 (4 workers)", Workers: 4, Scheduling: SchedNaive, Gate: GateGlobal,
			Streams: base.Streams, PerStream: base.PerStream},
		{Name: "Attempt 2 (8 workers)", Workers: 8, Scheduling: SchedNaive, Gate: GateGlobal,
			Streams: base.Streams, PerStream: base.PerStream},
		{Name: "Attempt 3 (round-robin)", Workers: 8, Scheduling: SchedRoundRobin, Gate: GateGlobal,
			Streams: base.Streams, PerStream: base.PerStream},
		// The intervention differs from Attempt 3 in gate scope only.
		{Name: "Intervention (per-stream gate)", Workers: 8, Scheduling: SchedRoundRobin, Gate: GatePerStream,
			Streams: base.Streams, PerStream: base.PerStream},
		// The control keeps the intervention but removes the divisible load.
		{Name: "Control (single stream)", Workers: 8, Scheduling: SchedRoundRobin, Gate: GatePerStream,
			Streams: 1, PerStream: 40},
	}
	for i := range configs {
		configs[i].Predicted = PredictTicks(configs[i])
	}
	return configs
}

func main() {
	fmt.Println("Why Map the Work? — teaching model (synthetic; not a benchmark)")
	fmt.Println()
	fmt.Printf("%-32s %-8s %-12s %-11s %-8s %-9s %-9s %s\n",
		"Configuration", "Workers", "Scheduling", "Gate", "Streams", "Predicted", "Measured", "Correct?")
	failed := false
	for _, c := range Configurations() {
		r := Run(c)
		correct := "yes"
		if err := VerifyCorrectness(r); err != nil {
			correct = "NO: " + err.Error()
			failed = true
		}
		fmt.Printf("%-32s %-8d %-12s %-11s %d x %-4d %-9d %-9d %s\n",
			c.Name, c.Workers, c.Scheduling, c.Gate, c.Streams, c.PerStream,
			c.Predicted, r.Ticks, correct)
	}
	fmt.Println()
	fmt.Println("Claim: ticks = largest number of updates sharing one gate scope.")
	if failed {
		os.Exit(1)
	}
}
