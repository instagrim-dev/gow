package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
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

func TestCLISourceIngestLifecycle(t *testing.T) {
	t.Parallel()

	workspace := t.TempDir()
	dbPath := filepath.Join(workspace, ".newf", "newf.db")
	sourceFile := filepath.Join(workspace, "paper-a.md")
	if err := os.WriteFile(sourceFile, []byte("# paper-a\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(text) error = %v", err)
	}

	binaryFile := filepath.Join(workspace, "paper-b.pdf")
	if err := os.WriteFile(binaryFile, []byte("%PDF-1.4\nbinary\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(binary) error = %v", err)
	}

	initStdout := &bytes.Buffer{}
	if code := execute(context.Background(), []string{"--db", dbPath, "--json", "init", "Erdős-Straus conjecture"}, initStdout, &bytes.Buffer{}); code != 0 {
		t.Fatalf("execute(init) code = %d", code)
	}
	var initResponse struct {
		ProblemID string `json:"problem_id"`
	}
	decodeJSONBuffer(t, initStdout, &initResponse)

	ingestText := runCLIJSON(t, []string{"--db", dbPath, "--json", "ingest", sourceFile, "--problem", initResponse.ProblemID})
	first := ingestText["results"].([]any)[0].(map[string]any)
	sourceID := first["source_id"].(string)
	snapshotID := first["snapshot_id"].(string)
	if first["status"] != "created_snapshot" {
		t.Fatalf("first ingest status = %v", first["status"])
	}

	ingestTextAgain := runCLIJSON(t, []string{"--db", dbPath, "--json", "ingest", sourceFile, "--problem", initResponse.ProblemID})
	second := ingestTextAgain["results"].([]any)[0].(map[string]any)
	if second["status"] != "existing_snapshot" {
		t.Fatalf("second ingest status = %v", second["status"])
	}

	if err := os.WriteFile(sourceFile, []byte("# paper-a revised\n"), 0o644); err != nil {
		t.Fatalf("WriteFile(revision) error = %v", err)
	}
	ingestRevision := runCLIJSON(t, []string{"--db", dbPath, "--json", "ingest", sourceFile, "--problem", initResponse.ProblemID})
	third := ingestRevision["results"].([]any)[0].(map[string]any)
	if third["status"] != "new_revision" {
		t.Fatalf("revision ingest status = %v", third["status"])
	}

	ingestBinary := runCLIJSON(t, []string{"--db", dbPath, "--json", "ingest", binaryFile, "--problem", initResponse.ProblemID})
	fourth := ingestBinary["results"].([]any)[0].(map[string]any)
	if fourth["media_type"] != "application/pdf" {
		t.Fatalf("binary ingest media_type = %v, want application/pdf", fourth["media_type"])
	}

	sourceList := runCLIJSON(t, []string{"--db", dbPath, "--json", "source", "list", "--problem", initResponse.ProblemID})
	if len(sourceList["sources"].([]any)) != 2 {
		t.Fatalf("source list count = %d, want 2", len(sourceList["sources"].([]any)))
	}

	sourceShow := runCLIJSON(t, []string{"--db", dbPath, "--json", "source", "show", sourceID})
	if len(sourceShow["snapshots"].([]any)) != 2 {
		t.Fatalf("source show snapshots = %d, want 2", len(sourceShow["snapshots"].([]any)))
	}

	snapshotShow := runCLIJSON(t, []string{"--db", dbPath, "--json", "source", "snapshot", "show", snapshotID})
	objectAbs := snapshotShow["object_absolute_path"].(string)

	verify := runCLIJSON(t, []string{"--db", dbPath, "--json", "source", "snapshot", "verify", snapshotID})
	if verify["status"] != "verified" {
		t.Fatalf("verify status = %v, want verified", verify["status"])
	}

	if err := os.WriteFile(objectAbs, []byte("corrupt"), 0o644); err != nil {
		t.Fatalf("WriteFile(corrupt) error = %v", err)
	}
	verifyCorrupt := runCLIJSON(t, []string{"--db", dbPath, "--json", "source", "snapshot", "verify", snapshotID})
	if verifyCorrupt["status"] != "hash_mismatch" {
		t.Fatalf("verify corrupt status = %v, want hash_mismatch", verifyCorrupt["status"])
	}
}

func TestCLIIngestReportsPerInputFailures(t *testing.T) {
	t.Parallel()

	workspace := t.TempDir()
	dbPath := filepath.Join(workspace, ".newf", "newf.db")
	validFile := filepath.Join(workspace, "ok.md")
	if err := os.WriteFile(validFile, []byte("ok"), 0o644); err != nil {
		t.Fatalf("WriteFile(valid) error = %v", err)
	}

	initResponse := runCLIJSON(t, []string{"--db", dbPath, "--json", "init", "Problem"})
	problemID := initResponse["problem_id"].(string)

	result := runCLIJSON(t, []string{"--db", dbPath, "--json", "ingest", validFile, filepath.Join(workspace, "missing.md"), "--problem", problemID})
	var failed bool
	for _, item := range result["results"].([]any) {
		entry := item.(map[string]any)
		if entry["status"] == "failed" {
			failed = true
		}
	}
	if !failed {
		t.Fatalf("expected at least one failed ingest result: %#v", result["results"])
	}
}

func runCLIJSON(t *testing.T, args []string) map[string]any {
	t.Helper()
	stdout := &bytes.Buffer{}
	if code := execute(context.Background(), args, stdout, &bytes.Buffer{}); code != 0 {
		t.Fatalf("execute(%v) code = %d, stdout=%s", args, code, stdout.String())
	}
	var decoded map[string]any
	decodeJSONBuffer(t, stdout, &decoded)
	return decoded
}

func decodeJSONBuffer(t *testing.T, buffer *bytes.Buffer, target any) {
	t.Helper()
	if err := json.Unmarshal(buffer.Bytes(), target); err != nil {
		t.Fatalf("json.Unmarshal() error = %v raw=%s", err, buffer.String())
	}
}
