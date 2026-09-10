package main

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/instagrim-dev/newf/internal/pipeline"
)

// E6 regression: cobra pre-RunE failures (wrong arg count, unknown flag) are
// USAGE errors attributed to the failing command — never `internal_error` on
// command "newf". Agents branch on this code; a misclassified usage error
// invites retrying an unretryable mistake.
func TestUsageErrorsAreClassifiedAndAttributed(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "newf.db")

	cases := []struct {
		name string
		args []string
	}{
		{"missing positional arg", []string{"--db", dbPath, "--json", "mechanism", "seed-fixture"}},
		{"unknown flag", []string{"--db", dbPath, "--json", "problem", "list", "--bogus"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stdout := &bytes.Buffer{}
			if code := execute(context.Background(), tc.args, stdout, &bytes.Buffer{}); code == 0 {
				t.Fatal("usage error must exit non-zero")
			}
			var resp pipeline.ErrorResponse
			if err := json.Unmarshal(stdout.Bytes(), &resp); err != nil {
				t.Fatalf("error output is not JSON: %v: %s", err, stdout.String())
			}
			if resp.Error.Code != "usage_error" {
				t.Fatalf("code = %q, want usage_error (message %q)", resp.Error.Code, resp.Error.Message)
			}
			if resp.Command == "newf" || resp.Command == "" {
				t.Fatalf("command = %q, want the failing subcommand path", resp.Command)
			}
		})
	}
}

// E1 regression at the CLI boundary: a flag-shaped --db value (empty shell
// var footgun) is rejected instead of silently creating a database named
// after the next flag.
func TestFlagShapedDBPathRejected(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	if code := execute(context.Background(), []string{"problem", "list", "--db", "--json"}, stdout, stderr); code == 0 {
		t.Fatal("flag-shaped --db value must fail")
	}
	combined := stdout.String() + stderr.String()
	if !bytes.Contains([]byte(combined), []byte("looks like a flag")) {
		t.Fatalf("expected the choke-point rejection message, got: %s", combined)
	}
}
