package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/instagrim-dev/newf/internal/composition"
)

const compositionCLIAttempt = `{
  "schema":"composition-attempt/1",
  "task":{
    "id":"cli-double-normalize",
    "family":"finite-expression-normalization",
    "source_ref":"development/cli",
    "authoring_provenance":"CLI fixture; independent authorship not asserted",
    "domain":{"width":4,"variables":["x"]},
    "start":{"op":"not","args":[{"op":"not","args":[{"op":"add","args":[{"var":"x"},{"const":0}]}]}]},
    "objective":{"kind":"execution-cost-at-most","max_node_visits":16}
  },
  "residual":{"kind":"excessive-cost","source_ref":"history/one","detail":"prior attempt left redundant layers"},
  "capability_requirement":{"statement":"compose two warranted simplifications","delivers":["normalized"]},
  "initial_capabilities":["finite-word-semantics"],
  "action_menu":[{"op":"add","args":[{"var":"x"},{"const":0}]}],
  "intervention_schemas":[
    {"id":"double-not","statement":"double complement is identity","variables":["a"],"requires":["finite-word-semantics"],"provides":["outer-removed"],"left":{"op":"not","args":[{"op":"not","args":[{"var":"a"}]}]},"right":{"var":"a"}},
    {"id":"add-zero","statement":"adding zero is identity","variables":["a"],"requires":["outer-removed"],"provides":["normalized"],"left":{"op":"add","args":[{"var":"a"},{"const":0}]},"right":{"var":"a"}}
  ],
  "candidate":[
    {"schema_id":"double-not","direction":"forward","before":{"op":"not","args":[{"op":"not","args":[{"op":"add","args":[{"var":"x"},{"const":0}]}]}]},"after":{"op":"add","args":[{"var":"x"},{"const":0}]}},
    {"schema_id":"add-zero","direction":"forward","before":{"op":"add","args":[{"var":"x"},{"const":0}]},"after":{"var":"x"}}
  ]
}`

func TestCompositionCLICommitAndObserve(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "attempt.json")
	commitPath := filepath.Join(dir, "commitment.json")
	observationPath := filepath.Join(dir, "observation.json")
	if err := os.WriteFile(input, []byte(compositionCLIAttempt), 0600); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	if code := execute(context.Background(), []string{"--json", "composition", "commit", "--input", input, "--out", commitPath}, &stdout, &stderr); code != 0 {
		t.Fatalf("commit failed: code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	var commitResp compositionCommitResponse
	if err := json.Unmarshal(stdout.Bytes(), &commitResp); err != nil {
		t.Fatal(err)
	}
	if !commitResp.OK || commitResp.Fidelity.Status != "verified" || commitResp.Commitment != commitPath {
		t.Fatalf("commit response: %+v", commitResp)
	}
	rawCommit, err := os.ReadFile(commitPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := composition.DecodeCommitment(rawCommit); err != nil {
		t.Fatalf("saved commitment: %v", err)
	}

	stdout.Reset()
	if code := execute(context.Background(), []string{"--json", "composition", "observe", commitPath, "--out", observationPath}, &stdout, &stderr); code != 0 {
		t.Fatalf("observe failed: code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	var observeResp compositionObserveResponse
	if err := json.Unmarshal(stdout.Bytes(), &observeResp); err != nil {
		t.Fatal(err)
	}
	if !observeResp.OK || observeResp.Result.StructuralFidelity.Status != "verified" || observeResp.Result.OriginalObjective.Status != "met" {
		t.Fatalf("observe response: %+v", observeResp)
	}
	rawObservation, err := os.ReadFile(observationPath)
	if err != nil {
		t.Fatal(err)
	}
	var persisted composition.Observation
	if err := json.Unmarshal(rawObservation, &persisted); err != nil {
		t.Fatal(err)
	}
	if persisted.IndependentEndpoint != "HOLDS_ON_DECLARED_DOMAIN" || persisted.OriginalObjective.MeasuredNodeVisits != 16 {
		t.Fatalf("persisted observation: %+v", persisted)
	}

	stdout.Reset()
	if code := execute(context.Background(), []string{"--json", "composition", "commit", "--input", input, "--out", commitPath}, &stdout, &stderr); code == 0 || !strings.Contains(stdout.String(), "already exists") {
		t.Fatalf("existing commitment must not be overwritten: code=%d stdout=%s", code, stdout.String())
	}
}

func TestCompositionCLIRecordsStructuralFailure(t *testing.T) {
	dir := t.TempDir()
	input := filepath.Join(dir, "attempt.json")
	commitPath := filepath.Join(dir, "commitment.json")
	// The final candidate is deliberately inserted into the supplied menu.
	raw := strings.Replace(compositionCLIAttempt,
		`"action_menu":[{"op":"add","args":[{"var":"x"},{"const":0}]}]`,
		`"action_menu":[{"var":"x"}]`, 1)
	if err := os.WriteFile(input, []byte(raw), 0600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := execute(context.Background(), []string{"--json", "composition", "commit", "--input", input, "--out", commitPath}, &stdout, &stderr); code != 0 {
		t.Fatalf("refuted composition should be recorded, not dropped: code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	var resp compositionCommitResponse
	if err := json.Unmarshal(stdout.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Fidelity.Status != "refuted" || len(resp.Fidelity.Defects) == 0 {
		t.Fatalf("structural failure must remain inspectable: %+v", resp)
	}
}

func TestCompositionPublicationRaceRetainsCompletedPendingResult(t *testing.T) {
	path := filepath.Join(t.TempDir(), "commitment.json")
	_, pending, err := prepareCompositionOutput(path, "commitment")
	if err != nil {
		t.Fatal(err)
	}
	defer pending.Close()
	if err := os.WriteFile(path, []byte("other writer"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := publishCompositionJSON(pending, path, composition.Commitment{Schema: composition.CommitmentSchema}); err == nil || !strings.Contains(err.Error(), "complete result retained") {
		t.Fatalf("concurrent publication must retain a completed pending result: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "other writer" {
		t.Fatalf("concurrent output was overwritten: %q", got)
	}
	if _, err := os.Stat(pending.Name()); err != nil {
		t.Fatalf("pending result was discarded after publication conflict: %v", err)
	}
}
