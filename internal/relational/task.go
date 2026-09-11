// Package relational is the deterministic calibration instrument for the
// relational-structure control (docs/theory/07-relational-structure.md).
//
// It owns small, complete factorial tasks with known outcome rules and a tiny
// executable relation language over their factors. Its job is to make the
// chain
//
//	observed attempts
//	 → proposed relation (executable, e.g. "success requires A ≠ B")
//	 → predictions for unobserved combinations
//	 → a discriminating next attempt
//	 → checked outcome
//
// checkable by code end-to-end. A model may PROPOSE a relation; this package
// VERIFIES it against the complete finite task (ModelJudgment != Verification).
//
// Scope discipline: this is instrument calibration, not domain machinery. The
// tasks are synthetic and total (every factor combination has a recorded
// outcome), so evaluation here is deliberately two-valued — unlike
// invariant-predicate/v1, which must carry `unknown` because real corpus
// signatures are partial. Nothing here touches mechanism signatures, storage,
// or providers, and no orchestration layer is introduced.
package relational

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// FactorName identifies one controlled binary/enum choice in a task.
type FactorName string

// Value is one element of a factor's finite domain.
type Value string

// Factor is a named finite choice axis.
type Factor struct {
	Name   FactorName
	Domain []Value
}

// Assignment maps every factor of a task to one value — one attempt's shape.
type Assignment map[FactorName]Value

// ErrInvalidTask marks a malformed task definition.
var ErrInvalidTask = errors.New("invalid relational task")

// ErrIncompleteAssignment marks an assignment that does not cover the task's
// factors exactly.
var ErrIncompleteAssignment = errors.New("assignment does not match task factors")

// Task is a complete finite factorial task: every combination of factor
// values has a frozen success/failure outcome. The outcome table — not the
// rule function that generated it — is the truth the checker consults, so a
// frozen task cannot drift.
type Task struct {
	Name     string
	Factors  []Factor
	outcomes map[string]bool
	order    []Assignment
}

// NewTask enumerates the full factorial table for the given factors under
// rule and freezes it. The rule function is consulted exactly once per
// combination and then discarded.
func NewTask(name string, factors []Factor, rule func(Assignment) bool) (*Task, error) {
	if name == "" {
		return nil, fmt.Errorf("%w: empty name", ErrInvalidTask)
	}
	if len(factors) == 0 {
		return nil, fmt.Errorf("%w: no factors", ErrInvalidTask)
	}
	seen := map[FactorName]struct{}{}
	for _, f := range factors {
		if f.Name == "" {
			return nil, fmt.Errorf("%w: unnamed factor", ErrInvalidTask)
		}
		if _, dup := seen[f.Name]; dup {
			return nil, fmt.Errorf("%w: duplicate factor %q", ErrInvalidTask, f.Name)
		}
		seen[f.Name] = struct{}{}
		if len(f.Domain) < 2 {
			return nil, fmt.Errorf("%w: factor %q needs a domain of at least 2 values", ErrInvalidTask, f.Name)
		}
		vseen := map[Value]struct{}{}
		for _, v := range f.Domain {
			if v == "" {
				return nil, fmt.Errorf("%w: factor %q has an empty domain value", ErrInvalidTask, f.Name)
			}
			if _, dup := vseen[v]; dup {
				return nil, fmt.Errorf("%w: factor %q has duplicate value %q", ErrInvalidTask, f.Name, v)
			}
			vseen[v] = struct{}{}
		}
	}
	t := &Task{Name: name, Factors: factors, outcomes: map[string]bool{}}
	var build func(i int, acc Assignment)
	build = func(i int, acc Assignment) {
		if i == len(factors) {
			a := cloneAssignment(acc)
			t.order = append(t.order, a)
			t.outcomes[assignmentKey(a)] = rule(a)
			return
		}
		for _, v := range factors[i].Domain {
			acc[factors[i].Name] = v
			build(i+1, acc)
		}
		delete(acc, factors[i].Name)
	}
	build(0, Assignment{})
	return t, nil
}

// Assignments returns every combination in deterministic enumeration order.
func (t *Task) Assignments() []Assignment {
	out := make([]Assignment, len(t.order))
	for i, a := range t.order {
		out[i] = cloneAssignment(a)
	}
	return out
}

// Success returns the frozen outcome for one complete assignment.
func (t *Task) Success(a Assignment) (bool, error) {
	if err := t.validateAssignment(a); err != nil {
		return false, err
	}
	return t.outcomes[assignmentKey(a)], nil
}

func (t *Task) validateAssignment(a Assignment) error {
	if len(a) != len(t.Factors) {
		return fmt.Errorf("%w: got %d factors, task %q has %d", ErrIncompleteAssignment, len(a), t.Name, len(t.Factors))
	}
	for _, f := range t.Factors {
		v, ok := a[f.Name]
		if !ok {
			return fmt.Errorf("%w: missing factor %q", ErrIncompleteAssignment, f.Name)
		}
		if !valueInDomain(v, f.Domain) {
			return fmt.Errorf("%w: value %q not in domain of factor %q", ErrIncompleteAssignment, v, f.Name)
		}
	}
	return nil
}

func valueInDomain(v Value, domain []Value) bool {
	for _, d := range domain {
		if d == v {
			return true
		}
	}
	return false
}

func cloneAssignment(a Assignment) Assignment {
	out := make(Assignment, len(a))
	for k, v := range a {
		out[k] = v
	}
	return out
}

// assignmentKey is the canonical serialization used for table lookup and
// deterministic tie-breaking: sorted "factor=value" pairs.
func assignmentKey(a Assignment) string {
	parts := make([]string, 0, len(a))
	for k, v := range a {
		parts = append(parts, string(k)+"="+string(v))
	}
	sort.Strings(parts)
	return strings.Join(parts, ",")
}

// XORTask is the frozen interaction control: two binary factors A and B,
// success exactly when they differ. Each factor alone is uninformative
// (P(success | A=a) = P(success | B=b) = 1/2); the relation determines the
// outcome completely. This is Proposition R's four-row table, executable.
func XORTask() *Task {
	t, err := NewTask("control-xor",
		[]Factor{
			{Name: "A", Domain: []Value{"0", "1"}},
			{Name: "B", Domain: []Value{"0", "1"}},
		},
		func(a Assignment) bool { return a["A"] != a["B"] },
	)
	if err != nil {
		panic(err) // fixed literal definition; unreachable
	}
	return t
}

// MainEffectTask is the frozen no-interaction control over the same factors:
// success exactly when A = 1, regardless of B. The A marginal alone suffices;
// the XOR relation scores at chance here. A checker that reaches the same
// conclusion on both controls is not measuring the interaction.
func MainEffectTask() *Task {
	t, err := NewTask("control-main-effect",
		[]Factor{
			{Name: "A", Domain: []Value{"0", "1"}},
			{Name: "B", Domain: []Value{"0", "1"}},
		},
		func(a Assignment) bool { return a["A"] == "1" },
	)
	if err != nil {
		panic(err) // fixed literal definition; unreachable
	}
	return t
}
