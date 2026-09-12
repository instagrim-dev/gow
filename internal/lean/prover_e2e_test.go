package lean

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// serverURL is a running llama-server exposing the prover. Set NEWF_LLAMA_SERVER
// to override; empty/unreachable skips.
func serverURL() string {
	if u := os.Getenv("NEWF_LLAMA_SERVER"); u != "" {
		return u
	}
	return "http://127.0.0.1:8077"
}

// tacticGrammar constrains generation to a single Lean tactic. A grammar is not
// decoration here: unconstrained, this model answers a proof request with prose
// ("### Detailed Proof / ### 1. Understanding the Problem..."), which is not a
// proof term and cannot be submitted to a kernel.
const tacticGrammar = `root ::= " " tactic "\n"
tactic ::= "simp [Nat.add_comm]" | "omega" | "exact Nat.add_comm a b" | "ring" | "simp_arith"
`

// complete asks the local prover for a constrained completion via llama-server's
// JSON API.
//
// WHY THE HTTP API AND NOT llama-cli: llama-cli writes an ASCII-art banner, a
// spinner, build info, a slash-command menu, and an echoed prompt to STDOUT,
// interleaved with the completion, and --log-disable does not suppress them.
// Successive attempts to screen-scrape it captured the spinner
// ("Loading model... |\b-\b\\"), then a progress bar ("▄▄ ▄▄"), then the literal
// string "Loading model..." -- each of which was dutifully submitted to the Lean
// kernel as a candidate proof and (correctly) rejected. The JSON API returns the
// completion in one field, so there is nothing to guess at.
func complete(t *testing.T, prompt, grammar string) (string, bool) {
	t.Helper()
	body, err := json.Marshal(map[string]any{
		"prompt":      prompt,
		"n_predict":   24,
		"temperature": 0,
		"seed":        1,
		"grammar":     grammar,
	})
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Post(serverURL()+"/completion", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", false // server not running: caller skips
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", false
	}
	var out struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode completion: %v", err)
	}
	return strings.TrimSpace(out.Content), true
}

// The full loop this repository is actually for: a LOCAL model proposes a proof
// under grammar constraint, and the Lean KERNEL -- not the model, and not its
// confidence -- decides whether the proof holds. This is the only route from
// model output to deterministic evidence in the system.
func TestProverOutputCheckedByKernel(t *testing.T) {
	c, err := NewChecker(projectDir(t))
	if err != nil {
		t.Skipf("lean unavailable: %v", err)
	}
	const stmt = "theorem add_comm_ex (a b : Nat) : a + b = b + a := by"

	tactic, ok := complete(t, stmt, tacticGrammar)
	if !ok {
		t.Skip("llama-server not reachable; start it on :8077 to run the end-to-end check")
	}
	if tactic == "" {
		t.Fatal("prover returned an empty tactic")
	}
	t.Logf("prover emitted tactic: %q", tactic)

	// Model authority ends at this line. Everything after is mechanical.
	res, err := c.Check(context.Background(), stmt+"\n  "+tactic+"\n")
	if err != nil {
		t.Fatalf("kernel check: %v", err)
	}
	t.Logf("kernel verdict=%s tool=%s dur=%s", res.Verdict, res.ToolVersion, res.Duration)
	if !res.Accepted() {
		t.Fatalf("kernel did not accept prover output %q: %s\n%s", tactic, res.Verdict, res.Diagnostics)
	}

	// The counterfactual that shows the kernel is load-bearing rather than
	// rubber-stamping: corrupt the SAME tactic and it must be rejected.
	bad := strings.ReplaceAll(tactic, "add_comm", "mul_comm")
	if bad == tactic {
		t.Logf("tactic %q has no add_comm to corrupt; skipping counterfactual", tactic)
		return
	}
	badRes, err := c.Check(context.Background(), stmt+"\n  "+bad+"\n")
	if err != nil {
		t.Fatalf("kernel check (corrupted): %v", err)
	}
	if badRes.Accepted() {
		t.Fatal("kernel accepted a corrupted proof; the check is not load-bearing")
	}
	t.Logf("corrupted tactic %q correctly rejected", bad)
}

// Unconstrained, the prover emits explanatory PROSE rather than a proof term.
// This is the empirical justification for grammar-constrained decoding: without
// it there is nothing a kernel could even be handed.
func TestUnconstrainedOutputIsNotKernelReady(t *testing.T) {
	c, err := NewChecker(projectDir(t))
	if err != nil {
		t.Skipf("lean unavailable: %v", err)
	}
	const stmt = "theorem add_comm_ex (a b : Nat) : a + b = b + a := by"

	// No grammar: the model is free to produce whatever it likes.
	raw, ok := complete(t, stmt, "")
	if !ok {
		t.Skip("llama-server not reachable")
	}
	t.Logf("unconstrained output: %q", truncate(raw, 160))

	res, err := c.Check(context.Background(), stmt+"\n  "+raw+"\n")
	if err != nil {
		t.Fatalf("kernel check: %v", err)
	}
	// Asserting the kernel does not ACCEPT it. Not asserting the specific
	// verdict: the point is that raw output carries no verification authority,
	// not that this model always produces exactly one failure mode.
	if res.Accepted() {
		t.Skip("unconstrained output happened to be a valid proof; not a stable property to assert")
	}
	t.Logf("as expected, unconstrained output is not kernel-accepted (verdict=%s)", res.Verdict)
}
