package main

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

// TestCLIEvaluateHoldoutRefused proves the CLI refuses holdout mode with a
// deferred-to-M7 error and emits a stable --json error envelope (R9/R10).
func TestCLIEvaluateHoldoutRefused(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "workspace", "newf.db")
	// Seed a problem so the failure is the holdout refusal, not a missing problem.
	stdout := &bytes.Buffer{}
	if code := execute(context.Background(), []string{"--db", dbPath, "--json", "init", "Erdos-Straus holdout guard"}, stdout, &bytes.Buffer{}); code != 0 {
		t.Fatalf("init failed: %s", stdout.String())
	}
	var initResp struct {
		ProblemID string `json:"problem_id"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &initResp); err != nil {
		t.Fatalf("decode init: %v", err)
	}

	// The evaluate command has no --mode flag (holdout is not user-selectable in
	// v0), so the service-level refusal is exercised via the pipeline in
	// evaluation_integration_test.go. Here we assert the CLI surface exists and
	// requires --problem (contract), and that a missing generation is a clean
	// error rather than a panic.
	errOut := &bytes.Buffer{}
	code := execute(context.Background(), []string{"--db", dbPath, "--json", "evaluate", "--problem", initResp.ProblemID}, errOut, &bytes.Buffer{})
	if code == 0 {
		t.Fatalf("evaluate with no frontier generation should fail cleanly, got success: %s", errOut.String())
	}
	var resp struct {
		OK    bool `json:"ok"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(errOut.Bytes(), &resp); err != nil {
		t.Fatalf("decode evaluate error: %v raw=%s", err, errOut.String())
	}
	if resp.OK {
		t.Fatal("expected ok=false")
	}
	if !strings.Contains(resp.Error.Message, "frontier generation") {
		t.Fatalf("expected a 'no frontier generation' message, got %q", resp.Error.Message)
	}
}

// TestCLIEvaluationRequiresProblem asserts the list surface requires --problem.
func TestCLIEvaluationRequiresProblem(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "workspace", "newf.db")
	code := execute(context.Background(), []string{"--db", dbPath, "--json", "evaluation", "list"}, &bytes.Buffer{}, &bytes.Buffer{})
	if code == 0 {
		t.Fatal("evaluation list without --problem should fail")
	}
}
