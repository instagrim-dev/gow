package main

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/domain"
)

func TestCLIProblemLifecycle(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "workspace", "newf.db")

	initStdout := &bytes.Buffer{}
	if code := execute(context.Background(), []string{"--db", dbPath, "--json", "init", "Erdős-Straus conjecture"}, initStdout, &bytes.Buffer{}); code != 0 {
		t.Fatalf("execute(init) code = %d", code)
	}

	var initResponse struct {
		ProblemID string `json:"problem_id"`
		RunID     string `json:"run_id"`
		Created   bool   `json:"created"`
	}
	if err := json.Unmarshal(initStdout.Bytes(), &initResponse); err != nil {
		t.Fatalf("json.Unmarshal(init) error = %v", err)
	}
	if !initResponse.Created {
		t.Fatal("init created = false, want true")
	}

	listStdout := &bytes.Buffer{}
	if code := execute(context.Background(), []string{"--db", dbPath, "--json", "problem", "list"}, listStdout, &bytes.Buffer{}); code != 0 {
		t.Fatalf("execute(problem list) code = %d", code)
	}
	if !strings.Contains(listStdout.String(), initResponse.ProblemID) {
		t.Fatalf("problem list output does not contain problem ID %q", initResponse.ProblemID)
	}

	showStdout := &bytes.Buffer{}
	if code := execute(context.Background(), []string{"--db", dbPath, "--json", "problem", "show", initResponse.ProblemID}, showStdout, &bytes.Buffer{}); code != 0 {
		t.Fatalf("execute(problem show) code = %d", code)
	}
	if !strings.Contains(showStdout.String(), "Erdős-Straus conjecture") {
		t.Fatalf("problem show output = %q", showStdout.String())
	}

	runStdout := &bytes.Buffer{}
	if code := execute(context.Background(), []string{"--db", dbPath, "--json", "run", "show", initResponse.RunID}, runStdout, &bytes.Buffer{}); code != 0 {
		t.Fatalf("execute(run show) code = %d", code)
	}
	if !strings.Contains(runStdout.String(), `"operation": "init"`) {
		t.Fatalf("run show output = %q", runStdout.String())
	}
}

func TestCLIJSONError(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "workspace", "newf.db")
	stdout := &bytes.Buffer{}
	if code := execute(context.Background(), []string{"--db", dbPath, "--json", "problem", "show", "run_01K4Y8X6YJJ66Y5QY9G7DNE1H2"}, stdout, &bytes.Buffer{}); code == 0 {
		t.Fatal("execute(problem show) succeeded, want failure")
	}

	var response struct {
		OK    bool `json:"ok"`
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal(error) error = %v", err)
	}
	if response.OK {
		t.Fatal("error response ok = true, want false")
	}
	if response.Error.Code != "invalid_input" {
		t.Fatalf("error code = %q, want invalid_input", response.Error.Code)
	}
}

func TestCLIJSONNotFoundError(t *testing.T) {
	t.Parallel()

	dbPath := filepath.Join(t.TempDir(), "workspace", "newf.db")
	missingProblemID := domain.NewProblemID(time.Now().UTC())

	stdout := &bytes.Buffer{}
	if code := execute(context.Background(), []string{"--db", dbPath, "--json", "problem", "show", missingProblemID}, stdout, &bytes.Buffer{}); code == 0 {
		t.Fatal("execute(problem show) succeeded, want failure")
	}

	var response struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("json.Unmarshal(error) error = %v", err)
	}
	if response.Error.Code != "not_found" {
		t.Fatalf("error code = %q, want not_found", response.Error.Code)
	}
}
