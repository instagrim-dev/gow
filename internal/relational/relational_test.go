package relational

import (
	"errors"
	"testing"
)

// TestPropositionR_MarginalsUninformativeRelationComplete is Proposition R
// (docs/theory/07-relational-structure.md) checked numerically over the four
// equally likely combinations: each factor's marginal success rate is exactly
// 1/2 for every value, while the relation A ≠ B predicts every row.
func TestPropositionR_MarginalsUninformativeRelationComplete(t *testing.T) {
	task := XORTask()

	// Full-table observations: every combination once.
	var obs []Observation
	for _, a := range task.Assignments() {
		ok, err := task.Success(a)
		if err != nil {
			t.Fatal(err)
		}
		obs = append(obs, Observation{Assignment: a, Success: ok})
	}

	rep := Marginals(task, obs)
	for _, c := range rep.Cells {
		if c.Trials != 2 {
			t.Fatalf("cell %s=%s: want 2 trials over the full table, got %d", c.Factor, c.Value, c.Trials)
		}
		if got := c.Rate(); got != 0.5 {
			t.Fatalf("cell %s=%s: marginal success rate = %v, want exactly 0.5", c.Factor, c.Value, got)
		}
	}
	for _, f := range []FactorName{"A", "B"} {
		if rep.FactorInformative(f) {
			t.Fatalf("factor %s: marginal analysis should be uninformative on XOR", f)
		}
	}

	check, err := Check(task, SuccessRequiresNeq("A", "B"))
	if err != nil {
		t.Fatal(err)
	}
	if !check.Complete || check.Accuracy() != 1.0 {
		t.Fatalf("relation A≠B should predict the complete XOR table; got accuracy %v, counterexamples %v",
			check.Accuracy(), check.Counterexamples)
	}
}

// TestControl_InteractionPresenceChangesConclusion is the calibration
// requirement: the same proposed relation must be accepted on the task whose
// hidden rule is the interaction and rejected (at chance) on the task whose
// hidden rule is a main effect — and vice versa for the marginal explanation.
func TestControl_InteractionPresenceChangesConclusion(t *testing.T) {
	xor := XORTask()
	main := MainEffectTask()

	neq := SuccessRequiresNeq("A", "B")
	aIsOne := Relation{
		Schema: RelationSchemaV1,
		Root:   RelNode{Op: RelOpEquals, Factor: "A", Value: "1"},
	}

	cases := []struct {
		name       string
		task       *Task
		relation   Relation
		wantExact  bool
		wantChance bool
	}{
		{"xor accepts A≠B", xor, neq, true, false},
		{"xor rejects A=1 at chance", xor, aIsOne, false, true},
		{"main-effect accepts A=1", main, aIsOne, true, false},
		{"main-effect rejects A≠B at chance", main, neq, false, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rep, err := Check(tc.task, tc.relation)
			if err != nil {
				t.Fatal(err)
			}
			if tc.wantExact && !rep.Complete {
				t.Fatalf("want complete prediction, got accuracy %v with counterexamples %v", rep.Accuracy(), rep.Counterexamples)
			}
			if tc.wantChance && rep.Accuracy() != 0.5 {
				t.Fatalf("want chance accuracy 0.5, got %v", rep.Accuracy())
			}
		})
	}

	// The marginal picture inverts across the controls too: on the
	// main-effect task the A marginal IS informative.
	var mainObs []Observation
	for _, a := range main.Assignments() {
		ok, err := main.Success(a)
		if err != nil {
			t.Fatal(err)
		}
		mainObs = append(mainObs, Observation{Assignment: a, Success: ok})
	}
	if rep := Marginals(main, mainObs); !rep.FactorInformative("A") {
		t.Fatal("factor A should be marginally informative on the main-effect task")
	}
}

// TestChain_XOR runs the full required chain on the interaction control:
// observed attempts → proposed relation → predictions for unobserved
// combinations → a discriminating next attempt → checked outcome.
func TestChain_XOR(t *testing.T) {
	task := XORTask()

	// Two observed failures whose marginals are all zero — individually
	// uninformative, jointly suggestive: both lie on the diagonal A=B.
	obs := []Observation{
		{Assignment: Assignment{"A": "0", "B": "0"}, Success: false},
		{Assignment: Assignment{"A": "1", "B": "1"}, Success: false},
	}
	rep := Marginals(task, obs)
	for _, f := range []FactorName{"A", "B"} {
		if rep.FactorInformative(f) {
			t.Fatalf("factor %s should be marginally uninformative on the observed diagonal", f)
		}
	}

	// The proposed relation and its rival disagree on every unobserved row.
	proposed := SuccessRequiresNeq("A", "B") // success iff A ≠ B
	rival := Relation{                       // success iff A = B (fits the two failures? no — it predicts success on them; a live rival nonetheless)
		Schema: RelationSchemaV1,
		Root:   RelNode{Op: RelOpFactorEq, Left: "A", Right: "B"},
	}

	rec, err := RunChain(task, proposed, []Relation{rival}, obs)
	if err != nil {
		t.Fatal(err)
	}

	// Commitments: exactly the two unobserved off-diagonal rows, both
	// predicted successes.
	if len(rec.Predictions) != 2 {
		t.Fatalf("want 2 predictions for unobserved combinations, got %d", len(rec.Predictions))
	}
	for _, p := range rec.Predictions {
		if p.Assignment["A"] == p.Assignment["B"] {
			t.Fatalf("prediction targets an observed diagonal row: %v", p.Assignment)
		}
		if !p.Success {
			t.Fatalf("A≠B must predict success on %v", p.Assignment)
		}
	}

	// The probe is deterministic (tie broken by canonical key) and separates
	// the candidates.
	if rec.Probe["A"] != "0" || rec.Probe["B"] != "1" {
		t.Fatalf("want deterministic probe A=0,B=1, got %v", rec.Probe)
	}
	if !rec.ProbePredicted {
		t.Fatal("proposed relation must commit to success on the probe")
	}
	if !rec.ProbeOutcome {
		t.Fatal("frozen XOR table must return success on A=0,B=1")
	}
	if !rec.ProbeSupports {
		t.Fatal("checked outcome should support the committed prediction")
	}
	if !rec.FullTableReport.Complete {
		t.Fatal("proposed relation should predict the complete table")
	}

	// The rival is refuted by the same checked outcome.
	if rival.Holds(rec.Probe) {
		t.Fatal("rival A=B must predict failure on the probe")
	}
}

// TestChain_MainEffect_RefutesInteractionClaim: on the no-interaction control
// the same chain must land against the relational proposal — the checked
// probe outcome contradicts A≠B's committed prediction.
func TestChain_MainEffect_RefutesInteractionClaim(t *testing.T) {
	task := MainEffectTask()
	obs := []Observation{
		{Assignment: Assignment{"A": "0", "B": "0"}, Success: false},
		{Assignment: Assignment{"A": "1", "B": "1"}, Success: true},
	}
	proposed := SuccessRequiresNeq("A", "B")
	rival := Relation{
		Schema: RelationSchemaV1,
		Root:   RelNode{Op: RelOpEquals, Factor: "A", Value: "1"},
	}
	rec, err := RunChain(task, proposed, []Relation{rival}, obs)
	if err != nil {
		t.Fatal(err)
	}
	// Probe A=0,B=1: proposed predicts success, table says failure.
	if rec.Probe["A"] != "0" || rec.Probe["B"] != "1" {
		t.Fatalf("want deterministic probe A=0,B=1, got %v", rec.Probe)
	}
	if rec.ProbeSupports {
		t.Fatal("checked outcome must refute A≠B on the main-effect task")
	}
	if rec.FullTableReport.Complete {
		t.Fatal("A≠B must not predict the complete main-effect table")
	}
	// The rival explanation survives the same probe.
	rivalRep, err := Check(task, rival)
	if err != nil {
		t.Fatal(err)
	}
	if !rivalRep.Complete {
		t.Fatal("A=1 should predict the complete main-effect table")
	}
}

// TestParseRelation_ExecutableDocument: the exact wire form the control
// demands — "success requires A ≠ B" as a validated document, with prose-only
// or ill-typed proposals rejected before scoring.
func TestParseRelation_ExecutableDocument(t *testing.T) {
	task := XORTask()

	r, err := ParseRelation(`{"schema":"relational-claim/v1","root":{"op":"factor_neq","left":"A","right":"B"}}`, task)
	if err != nil {
		t.Fatal(err)
	}
	rep, err := Check(task, r)
	if err != nil {
		t.Fatal(err)
	}
	if !rep.Complete {
		t.Fatal("parsed A≠B document should predict the complete XOR table")
	}

	rejects := []struct {
		name string
		raw  string
	}{
		{"wrong schema", `{"schema":"relational-claim/v2","root":{"op":"factor_neq","left":"A","right":"B"}}`},
		{"unknown factor", `{"schema":"relational-claim/v1","root":{"op":"factor_neq","left":"A","right":"C"}}`},
		{"self relation", `{"schema":"relational-claim/v1","root":{"op":"factor_neq","left":"A","right":"A"}}`},
		{"value outside domain", `{"schema":"relational-claim/v1","root":{"op":"equals","factor":"A","value":"2"}}`},
		{"extraneous fields", `{"schema":"relational-claim/v1","root":{"op":"factor_neq","left":"A","right":"B","value":"1"}}`},
		{"unknown op", `{"schema":"relational-claim/v1","root":{"op":"coupled","left":"A","right":"B"}}`},
		{"unknown json field", `{"schema":"relational-claim/v1","root":{"op":"factor_neq","left":"A","right":"B"},"prose":"A and B are coupled"}`},
		{"empty boolean", `{"schema":"relational-claim/v1","root":{"op":"all"}}`},
	}
	for _, tc := range rejects {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := ParseRelation(tc.raw, task); !errors.Is(err, ErrInvalidRelation) {
				t.Fatalf("want ErrInvalidRelation, got %v", err)
			}
		})
	}
}

// TestDiscriminatingProbe_Degenerate: when nothing separates the candidates,
// or nothing is left to observe, the instrument says so instead of inventing
// an informative probe.
func TestDiscriminatingProbe_Degenerate(t *testing.T) {
	task := XORTask()

	// Identical candidates: no combination separates them.
	same := SuccessRequiresNeq("A", "B")
	if _, err := DiscriminatingProbe(task, []Relation{same, same}, nil); err == nil {
		t.Fatal("want error when no combination separates the candidates")
	}

	// Fully observed task: nothing left to probe.
	var obs []Observation
	for _, a := range task.Assignments() {
		ok, err := task.Success(a)
		if err != nil {
			t.Fatal(err)
		}
		obs = append(obs, Observation{Assignment: a, Success: ok})
	}
	eq := Relation{Schema: RelationSchemaV1, Root: RelNode{Op: RelOpFactorEq, Left: "A", Right: "B"}}
	if _, err := DiscriminatingProbe(task, []Relation{same, eq}, obs); err == nil {
		t.Fatal("want error when every combination is already observed")
	}

	// Fewer than two candidates is not a probe-selection problem.
	if _, err := DiscriminatingProbe(task, []Relation{same}, nil); err == nil {
		t.Fatal("want error for a single candidate")
	}
}

// TestRelationExpressibility_BooleanExpansion: the theory doc claims a finite
// relation is expressible WITHOUT a new primitive as a boolean expansion over
// literals (the invariant-predicate/v1 bridge). Verify the expansion of XOR
// is exactly equivalent to factor_neq on the complete table.
func TestRelationExpressibility_BooleanExpansion(t *testing.T) {
	task := XORTask()
	expansion := Relation{
		Schema: RelationSchemaV1,
		Root: RelNode{Op: RelOpAny, Children: []RelNode{
			{Op: RelOpAll, Children: []RelNode{
				{Op: RelOpEquals, Factor: "A", Value: "0"},
				{Op: RelOpEquals, Factor: "B", Value: "1"},
			}},
			{Op: RelOpAll, Children: []RelNode{
				{Op: RelOpEquals, Factor: "A", Value: "1"},
				{Op: RelOpEquals, Factor: "B", Value: "0"},
			}},
		}},
	}
	neq := SuccessRequiresNeq("A", "B")
	for _, a := range task.Assignments() {
		if expansion.Holds(a) != neq.Holds(a) {
			t.Fatalf("boolean expansion diverges from factor_neq on %v", a)
		}
	}
	rep, err := Check(task, expansion)
	if err != nil {
		t.Fatal(err)
	}
	if !rep.Complete {
		t.Fatal("boolean expansion of XOR should predict the complete table")
	}
}

// TestNewTask_Validation guards the task-definition boundary.
func TestNewTask_Validation(t *testing.T) {
	bad := []struct {
		name    string
		factors []Factor
	}{
		{"no factors", nil},
		{"duplicate factor", []Factor{{Name: "A", Domain: []Value{"0", "1"}}, {Name: "A", Domain: []Value{"0", "1"}}}},
		{"singleton domain", []Factor{{Name: "A", Domain: []Value{"0"}}}},
		{"duplicate value", []Factor{{Name: "A", Domain: []Value{"0", "0"}}}},
		{"empty value", []Factor{{Name: "A", Domain: []Value{"", "1"}}}},
	}
	for _, tc := range bad {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := NewTask("t", tc.factors, func(Assignment) bool { return false }); !errors.Is(err, ErrInvalidTask) {
				t.Fatalf("want ErrInvalidTask, got %v", err)
			}
		})
	}

	task := XORTask()
	if _, err := task.Success(Assignment{"A": "0"}); !errors.Is(err, ErrIncompleteAssignment) {
		t.Fatalf("want ErrIncompleteAssignment for partial assignment, got %v", err)
	}
	if _, err := task.Success(Assignment{"A": "0", "B": "2"}); !errors.Is(err, ErrIncompleteAssignment) {
		t.Fatalf("want ErrIncompleteAssignment for out-of-domain value, got %v", err)
	}
}
