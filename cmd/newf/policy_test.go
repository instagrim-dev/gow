package main

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

// TestCLIPolicyMutateRequiresProblem proves the `policy mutate` surface exists
// and requires --problem (contract), emitting a stable --json error envelope.
func TestCLIPolicyMutateRequiresProblem(t *testing.T) {
	t.Parallel()
	dbPath := filepath.Join(t.TempDir(), "workspace", "newf.db")
	errOut := &bytes.Buffer{}
	code := execute(context.Background(), []string{"--db", dbPath, "--json", "policy", "mutate"}, errOut, &bytes.Buffer{})
	if code == 0 {
		t.Fatalf("policy mutate without --problem should fail, got success: %s", errOut.String())
	}
	var resp struct {
		OK    bool `json:"ok"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(errOut.Bytes(), &resp); err != nil {
		t.Fatalf("decode error envelope: %v (%s)", err, errOut.String())
	}
	if resp.OK || !strings.Contains(resp.Error.Message, "problem") {
		t.Fatalf("expected a --problem-required error envelope, got %+v", resp)
	}
}

// TestCLIPolicyMutateEmptyProblem proves mutate on a seeded-but-empty problem
// produces a legitimate empty policy revision (no fabricated directives) and a
// clean --json envelope.
func TestCLIPolicyMutateEmptyProblem(t *testing.T) {
	t.Parallel()
	dbPath := filepath.Join(t.TempDir(), "workspace", "newf.db")
	stdout := &bytes.Buffer{}
	if code := execute(context.Background(), []string{"--db", dbPath, "--json", "init", "Erdos-Straus policy smoke"}, stdout, &bytes.Buffer{}); code != 0 {
		t.Fatalf("init failed: %s", stdout.String())
	}
	var initResp struct {
		ProblemID string `json:"problem_id"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &initResp); err != nil {
		t.Fatalf("decode init: %v", err)
	}

	out := &bytes.Buffer{}
	code := execute(context.Background(), []string{"--db", dbPath, "--json", "policy", "mutate", "--problem", initResp.ProblemID}, out, &bytes.Buffer{})
	if code != 0 {
		t.Fatalf("policy mutate on empty problem should succeed, got failure: %s", out.String())
	}
	var resp struct {
		OK       bool `json:"ok"`
		Revision struct {
			DirectiveCount int `json:"directive_count"`
		} `json:"revision"`
	}
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		t.Fatalf("decode mutate: %v (%s)", err, out.String())
	}
	if !resp.OK {
		t.Fatalf("expected ok envelope, got %s", out.String())
	}
	if resp.Revision.DirectiveCount != 0 {
		t.Fatalf("empty problem must yield 0 directives, got %d", resp.Revision.DirectiveCount)
	}
}
