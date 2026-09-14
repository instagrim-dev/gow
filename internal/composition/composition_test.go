package composition

import (
	"strings"
	"testing"
)

const validAttempt = `{
  "schema": "composition-attempt/1",
  "task": {
    "id": "task-double-normalize",
    "family": "finite-expression-normalization",
    "source_ref": "custodian-development-fixture/1",
    "authoring_provenance": "development fixture; independent authorship not asserted",
    "domain": {"width": 4, "variables": ["x"]},
    "start": {"op": "not", "args": [{"op": "not", "args": [{"op": "add", "args": [{"var": "x"}, {"const": 0}]}]}]},
    "objective": {"kind": "execution-cost-at-most", "max_node_visits": 16}
  },
  "residual": {
    "kind": "excessive-cost",
    "source_ref": "attempt-history/step-1",
    "detail": "the prior realization retained two redundant layers under the declared evaluator"
  },
  "capability_requirement": {
    "statement": "construct a semantics-preserving realization that discharges both redundant layers",
    "delivers": ["normalized"]
  },
  "initial_capabilities": ["finite-word-semantics"],
  "action_menu": [
    {"op": "add", "args": [{"var": "x"}, {"const": 0}]},
    {"op": "not", "args": [{"op": "not", "args": [{"var": "x"}]}]}
  ],
  "intervention_schemas": [
    {
      "id": "double-not",
      "statement": "double complement is identity on declared finite words",
      "variables": ["a"],
      "requires": ["finite-word-semantics"],
      "provides": ["outer-layer-removed"],
      "left": {"op": "not", "args": [{"op": "not", "args": [{"var": "a"}]}]},
      "right": {"var": "a"}
    },
    {
      "id": "add-zero",
      "statement": "adding zero is identity on declared finite words",
      "variables": ["a"],
      "requires": ["outer-layer-removed"],
      "provides": ["normalized"],
      "left": {"op": "add", "args": [{"var": "a"}, {"const": 0}]},
      "right": {"var": "a"}
    }
  ],
  "candidate": [
    {
      "schema_id": "double-not",
      "direction": "forward",
      "before": {"op": "not", "args": [{"op": "not", "args": [{"op": "add", "args": [{"var": "x"}, {"const": 0}]}]}]},
      "after": {"op": "add", "args": [{"var": "x"}, {"const": 0}]}
    },
    {
      "schema_id": "add-zero",
      "direction": "forward",
      "before": {"op": "add", "args": [{"var": "x"}, {"const": 0}]},
      "after": {"var": "x"}
    }
  ]
}`

func TestCommitThenObserveSeparatesFidelityFromObjective(t *testing.T) {
	commitment, err := Commit([]byte(validAttempt))
	if err != nil {
		t.Fatalf("commit: %v", err)
	}
	if commitment.StructuralFidelity.Status != "verified" || commitment.StructuralFidelity.Final != "x" {
		t.Fatalf("pre-observation commitment must replay a composed realization: %+v", commitment.StructuralFidelity)
	}
	if strings.Contains(string(commitment.Attempt), "measured_node_visits") {
		t.Fatal("commitment must retain task input, not an objective observation")
	}

	observation, err := Observe(commitment)
	if err != nil {
		t.Fatalf("observe: %v", err)
	}
	if observation.StructuralFidelity.Status != "verified" || observation.IndependentEndpoint != "HOLDS_ON_DECLARED_DOMAIN" {
		t.Fatalf("observation must independently replay semantic fidelity: %+v", observation)
	}
	if observation.OriginalObjective.Status != "met" || observation.OriginalObjective.MeasuredNodeVisits != 16 || observation.OriginalObjective.Assignments != 16 {
		t.Fatalf("objective must be separately measured over the declared domain: %+v", observation.OriginalObjective)
	}
}

func TestObjectiveMissRemainsSeparateFromStructuralFidelity(t *testing.T) {
	raw := strings.Replace(validAttempt,
		`"max_node_visits": 16`,
		`"max_node_visits": 15`, 1)
	commitment, err := Commit([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	if commitment.StructuralFidelity.Status != "verified" {
		t.Fatalf("cost threshold must not alter structural fidelity: %+v", commitment.StructuralFidelity)
	}
	observation, err := Observe(commitment)
	if err != nil {
		t.Fatal(err)
	}
	if observation.StructuralFidelity.Status != "verified" || observation.OriginalObjective.Status != "not-met" ||
		observation.OriginalObjective.MeasuredNodeVisits != 16 {
		t.Fatalf("a faithful realization that misses cost must remain a distinct outcome: %+v", observation)
	}
}

func TestCommitRefutesMenuSelectionAndObservationPreservesNonEvaluation(t *testing.T) {
	raw := strings.Replace(validAttempt,
		`{"op": "add", "args": [{"var": "x"}, {"const": 0}]},
    {"op": "not", "args": [{"op": "not", "args": [{"var": "x"}]}]}`,
		`{"var": "x"}`, 1)
	commitment, err := Commit([]byte(raw))
	if err != nil {
		t.Fatalf("commit refuted candidate: %v", err)
	}
	if commitment.StructuralFidelity.Status != "refuted" || !strings.Contains(strings.Join(commitment.StructuralFidelity.Defects, " | "), "action menu") {
		t.Fatalf("final menu selection must be a recorded refutation: %+v", commitment.StructuralFidelity)
	}
	observation, err := Observe(commitment)
	if err != nil {
		t.Fatalf("observe refuted candidate: %v", err)
	}
	if observation.OriginalObjective.Status != "not-evaluated" || observation.IndependentEndpoint != "not-evaluated" {
		t.Fatalf("a fidelity defect must not be mislabeled as an objective miss: %+v", observation)
	}
}

func TestCommitRefutesUndischargedPreconditionAndUnjustifiedEquality(t *testing.T) {
	t.Run("undischarged precondition", func(t *testing.T) {
		raw := strings.Replace(validAttempt,
			`"initial_capabilities": ["finite-word-semantics"]`,
			`"initial_capabilities": ["other-capability"]`, 1)
		commitment, err := Commit([]byte(raw))
		if err != nil {
			t.Fatalf("commit: %v", err)
		}
		if commitment.StructuralFidelity.Status != "refuted" || !strings.Contains(strings.Join(commitment.StructuralFidelity.Defects, " | "), "undischarged preconditions") {
			t.Fatalf("missing precondition must be recorded: %+v", commitment.StructuralFidelity)
		}
	})
	t.Run("unjustified equality", func(t *testing.T) {
		raw := strings.Replace(validAttempt,
			`"right": {"var": "a"}`,
			`"right": {"const": 0}`, 1)
		commitment, err := Commit([]byte(raw))
		if err != nil {
			t.Fatalf("commit: %v", err)
		}
		if commitment.StructuralFidelity.Status != "refuted" || !strings.Contains(strings.Join(commitment.StructuralFidelity.Defects, " | "), "unjustified equality") {
			t.Fatalf("false equality must never enter the composed trace: %+v", commitment.StructuralFidelity)
		}
	})
}

func TestStrictInputAndCommitmentBinding(t *testing.T) {
	if _, err := DecodeAttempt([]byte(strings.Replace(validAttempt, `"schema": "composition-attempt/1"`, `"schema": "composition-attempt/1", "unexpected": true`, 1))); err == nil {
		t.Fatal("unknown attempt field must be refused")
	}
	commitment, err := Commit([]byte(validAttempt))
	if err != nil {
		t.Fatal(err)
	}
	commitment.Attempt = []byte(strings.Replace(string(commitment.Attempt), "task-double-normalize", "task-tampered-normalize", 1))
	if _, err := Observe(commitment); err == nil || !strings.Contains(err.Error(), "digest") {
		t.Fatalf("tampered retained input must be refused before observation: %v", err)
	}
}
