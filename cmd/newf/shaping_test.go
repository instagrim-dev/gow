package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/instagrim-dev/newf/internal/sealedrun"
)

func shapingArgs(path string) []string {
	return []string{"--json", "shaping", "diagnose", "--pack", "smoke", "--out", path,
		"--expansions", "2", "--rule-applications", "128", "--candidates", "256", "--history-bytes", "65536",
		"--check-assignments", "4096", "--max-states", "1024", "--max-term-nodes", "1024"}
}

func TestShapingCLIExecutionReceiptAndInspection(t *testing.T) {
	path := filepath.Join(t.TempDir(), "receipt.json")
	var stdout, stderr bytes.Buffer
	if code := execute(context.Background(), shapingArgs(path), &stdout, &stderr); code != 0 {
		t.Fatalf("diagnose failed: code=%d stdout=%s stderr=%s", code, stdout.String(), stderr.String())
	}
	receipt, err := readShapingReceipt(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(receipt.Cells) != 4 || len(receipt.Completions) != 4 || receipt.BuildVersion != version || !strings.Contains(receipt.EvidenceLabel, "freshness-unverified") {
		t.Fatalf("incomplete or overstated receipt: %+v", receipt)
	}
	for _, c := range receipt.Cells {
		if !c.SearchStarted || c.Decision.ControllerVersion == "" || c.Decision.InputHash == "" || c.Search.Endpoint.Binding.Sentence == "" || c.Search.Endpoint.Left != c.Search.Original || c.Search.Endpoint.Right != c.Search.Best {
			t.Fatalf("real producer receipt is missing execution identity/checking: %+v", c)
		}
	}
	before, _ := os.ReadFile(path)
	stdout.Reset()
	if code := execute(context.Background(), []string{"--json", "shaping", "inspect", path}, &stdout, &stderr); code != 0 {
		t.Fatalf("inspect failed: code=%d %s %s", code, stdout.String(), stderr.String())
	}
	var response shapingResponse
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if !response.OK || !strings.Contains(response.Scope, "certificates not independently replayed") || response.Diagnostic.Assessment != "completed-development-diagnostic" {
		t.Fatalf("inspection must be bounded to reconstructed arithmetic: %+v", response)
	}
	after, _ := os.ReadFile(path)
	if !bytes.Equal(before, after) {
		t.Fatal("inspection mutated the append-only receipt")
	}
	stdout.Reset()
	if code := execute(context.Background(), shapingArgs(path), &stdout, &stderr); code == 0 || !strings.Contains(stdout.String(), "already exists") {
		t.Fatalf("a second run must never overwrite a receipt: code=%d %s", code, stdout.String())
	}
	after, _ = os.ReadFile(path)
	if !bytes.Equal(before, after) {
		t.Fatal("repeated output path changed the existing receipt")
	}
}

func TestShapingCLISetupFailureRemainsInspectable(t *testing.T) {
	path := filepath.Join(t.TempDir(), "setup-blocked.json")
	args := shapingArgs(path)
	for i := range args {
		if args[i] == "--check-assignments" {
			args[i+1] = "0"
		}
	}
	var out bytes.Buffer
	if code := execute(context.Background(), args, &out, &bytes.Buffer{}); code == 0 {
		t.Fatal("insufficient checking reservation must block execution")
	}
	receipt, err := readShapingReceipt(path)
	if err != nil || !strings.Contains(receipt.Error, "reserved assignments") {
		t.Fatalf("setup failure receipt must preserve the original reason: err=%v %+v", err, receipt)
	}
	out.Reset()
	if code := execute(context.Background(), []string{"--json", "shaping", "inspect", path}, &out, &bytes.Buffer{}); code != 0 {
		t.Fatalf("setup failure receipt must remain inspectable: %s", out.String())
	}
	var inspected shapingResponse
	if err := json.Unmarshal(out.Bytes(), &inspected); err != nil || inspected.RecordedError != receipt.Error || inspected.Diagnostic.Assessment != "blocked" {
		t.Fatalf("inspection must retain the setup failure as well as reassessment status: err=%v %s", err, out.String())
	}
}

func TestShapingCLICancellationPublishesReceiptBeforeFailure(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cancelled.json")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var out bytes.Buffer
	if code := execute(ctx, shapingArgs(path), &out, &bytes.Buffer{}); code == 0 || !strings.Contains(out.String(), "execution receipt saved") {
		t.Fatalf("cancelled execution must report its saved receipt before failure: code=%d %s", code, out.String())
	}
	receipt, err := readShapingReceipt(path)
	if err != nil {
		t.Fatalf("cancellation discarded the receipt: %v", err)
	}
	if receipt.Assessment != "blocked" || !strings.Contains(receipt.Error, "cancelled") || receipt.Completions != nil || receipt.BuildVersion != version {
		t.Fatalf("cancellation must retain its attributed refusal without completion totals: %+v", receipt)
	}
}

func TestShapingCLIRequiresExplicitResourceVector(t *testing.T) {
	path := filepath.Join(t.TempDir(), "unrequested.json")
	var out bytes.Buffer
	if code := execute(context.Background(), []string{"--json", "shaping", "diagnose", "--pack", "smoke", "--out", path}, &out, &bytes.Buffer{}); code == 0 || !strings.Contains(out.String(), "required flag") {
		t.Fatalf("missing resource allowances must refuse before dispatch: code=%d %s", code, out.String())
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("missing authorization vector should create no receipt: %v", err)
	}
}

func TestShapingCLIBlockedExecutionSavesReceiptBeforeError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "blocked.json")
	args := shapingArgs(path)
	for i := range args {
		if args[i] == "--candidates" {
			args[i+1] = "0"
		}
	}
	var out bytes.Buffer
	if code := execute(context.Background(), args, &out, &bytes.Buffer{}); code == 0 || !strings.Contains(out.String(), "execution receipt saved") {
		t.Fatalf("blocked execution must name its retained receipt: code=%d %s", code, out.String())
	}
	var errorEnvelope map[string]any
	if err := json.Unmarshal(out.Bytes(), &errorEnvelope); err != nil {
		t.Fatalf("blocked JSON output must be one complete error envelope: %v: %s", err, out.String())
	}
	receipt, err := readShapingReceipt(path)
	if err != nil {
		t.Fatal(err)
	}
	if receipt.Assessment != "blocked" || receipt.Error == "" || len(receipt.Cells) != 4 || receipt.Completions != nil {
		t.Fatalf("blocked receipt lost raw cells or manufactured totals: %+v", receipt)
	}
	out.Reset()
	if code := execute(context.Background(), []string{"--json", "shaping", "inspect", path}, &out, &bytes.Buffer{}); code != 0 {
		t.Fatalf("valid blocked receipt must remain inspectable: %s", out.String())
	}
	var inspected shapingResponse
	if err := json.Unmarshal(out.Bytes(), &inspected); err != nil || inspected.Diagnostic.Assessment != "blocked" || inspected.Diagnostic.Completions != nil {
		t.Fatalf("inspection promoted a blocked receipt: err=%v %s", err, out.String())
	}
}

func TestShapingReceiptStrictInputBoundary(t *testing.T) {
	for _, tc := range []struct{ name, data, want string }{
		{"unknown field", `{"Version":"` + sealedrun.ResourceDesignVersion + `","InjectedAuthority":true}`, "unknown field"},
		{"trailing object", `{"Version":"` + sealedrun.ResourceDesignVersion + `"} {}`, "exactly one"},
		{"null", `null`, "unsupported receipt version"},
		{"oversize", strings.Repeat(" ", maxShapingReceiptBytes+1), "inspection limit"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "input.json")
			if err := os.WriteFile(path, []byte(tc.data), 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := readShapingReceipt(path); err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("wanted %q refusal, got %v", tc.want, err)
			}
		})
	}
}

func TestShapingReceiptConcurrentPublicationDoesNotOverwrite(t *testing.T) {
	path := filepath.Join(t.TempDir(), "receipt.json")
	_, pending, err := prepareShapingReceipt(path)
	if err != nil {
		t.Fatal(err)
	}
	defer pending.Close()
	const other = "another writer's receipt"
	if err := os.WriteFile(path, []byte(other), 0600); err != nil {
		t.Fatal(err)
	}
	if err := publishShapingReceipt(pending, path, sealedrun.ResourceReceipt{Version: sealedrun.ResourceDesignVersion}); err == nil || !strings.Contains(err.Error(), "complete receipt retained") {
		t.Fatalf("concurrent output must refuse publication and retain pending receipt: %v", err)
	}
	data, _ := os.ReadFile(path)
	if string(data) != other {
		t.Fatal("publication replaced concurrent writer's receipt")
	}
	if _, err := readShapingReceipt(pending.Name()); err != nil {
		t.Fatalf("complete pending receipt was not retained: %v", err)
	}
}
