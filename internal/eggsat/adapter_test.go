package eggsat

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/finite"
	"github.com/instagrim-dev/newf/internal/rewrite"
)

func TestOptimizeReplaysAnExternalExtraction(t *testing.T) {
	t.Setenv("NEWF_EGGSAT_TEST_HELPER", "1")
	domain := finite.Domain{Width: 4, Vars: []string{"a"}}
	lhs := finite.Unary{Op: finite.OpNot, X: finite.Unary{Op: finite.OpNot, X: finite.Var{Name: "a"}}}
	rule := admitted(t, "double-not", lhs, finite.Var{Name: "a"}, domain)
	start := finite.Unary{Op: finite.OpNot, X: finite.Unary{Op: finite.OpNot, X: finite.Var{Name: "x"}}}

	client := Client{Binary: os.Args[0]}
	result, err := client.Optimize(context.Background(), start, finite.Domain{Width: 4, Vars: []string{"x"}}, []rewrite.Rule{rule}, Limits{Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if result.Best != "x" || result.BestCost != 1 {
		t.Fatalf("unexpected extracted expression: %+v", result)
	}
	if !result.EndpointVerified || result.Endpoint.Verdict != finite.VerdictHoldsOnDomain {
		t.Fatalf("endpoint was not independently verified: %+v", result.Endpoint)
	}
	if !strings.Contains(result.EngineExplanation, rule.Identity()) || len(result.Proof) != 1 || result.Proof[0].RuleIdentity != rule.Identity() {
		t.Fatalf("engine explanation was not replayed against the admitted rule: %+v", result)
	}
}

func TestOptimizeRejectsAForgedEngineProofBeforeEndpointReplay(t *testing.T) {
	t.Setenv("NEWF_EGGSAT_TEST_HELPER", "1")
	t.Setenv("NEWF_EGGSAT_TEST_REFUTE", "1")
	domain := finite.Domain{Width: 4, Vars: []string{"a"}}
	lhs := finite.Unary{Op: finite.OpNot, X: finite.Unary{Op: finite.OpNot, X: finite.Var{Name: "a"}}}
	rule := admitted(t, "double-not", lhs, finite.Var{Name: "a"}, domain)
	start := finite.Var{Name: "x"}
	client := Client{Binary: os.Args[0]}
	result, err := client.Optimize(context.Background(), start, finite.Domain{Width: 4, Vars: []string{"x"}}, []rewrite.Rule{rule}, Limits{Timeout: time.Second})
	if !errors.Is(err, ErrMalformedResponse) {
		t.Fatalf("expected forged proof rejection, got result=%+v err=%v", result, err)
	}
	if result.Endpoint.Verdict != "" {
		t.Fatalf("a forged proof must not reach endpoint replay: %+v", result.Endpoint)
	}
}

func TestDecodeTermRejectsAPatternOrOversizedTree(t *testing.T) {
	if _, err := decodeTerm("?x"); err == nil {
		t.Fatal("pattern variable must not be accepted as an engine result")
	}
	term := "x"
	for range finite.MaxExprDepth + 1 {
		term = "(not " + term + ")"
	}
	if _, err := decodeTerm(term); err == nil {
		t.Fatal("overdeep engine output must be rejected before finite evaluation")
	}
}

// TestExternalEngineE2E is an opt-in integration boundary. Unit tests use the
// helper subprocess above so Go CI needs no Rust toolchain; the maintained
// G2 adapter is exercised against the pinned binary when the caller supplies
// NEWF_EGGSAT_BINARY.
func TestExternalEngineE2E(t *testing.T) {
	binary := os.Getenv("NEWF_EGGSAT_BINARY")
	if binary == "" {
		t.Skip("NEWF_EGGSAT_BINARY is not set")
	}
	domain := finite.Domain{Width: 4, Vars: []string{"a"}}
	lhs := finite.Unary{Op: finite.OpNot, X: finite.Unary{Op: finite.OpNot, X: finite.Var{Name: "a"}}}
	rule := admitted(t, "double-not", lhs, finite.Var{Name: "a"}, domain)
	start := finite.Unary{Op: finite.OpNot, X: finite.Unary{Op: finite.OpNot, X: finite.Var{Name: "x"}}}

	result, err := (Client{Binary: binary}).Optimize(context.Background(), start, finite.Domain{Width: 4, Vars: []string{"x"}}, []rewrite.Rule{rule}, Limits{Iterations: 10, NodeLimit: 100, Timeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if result.Best != "x" || !result.EndpointVerified {
		t.Fatalf("external engine did not produce a verified reduction: %+v", result)
	}
}

func admitted(t *testing.T, name string, lhs, rhs finite.Expr, domain finite.Domain) rewrite.Rule {
	t.Helper()
	cert := finite.AssessEquivalence(finite.Binding{Sentence: name, Domain: domain}, lhs, rhs)
	rule, defects := rewrite.AdmitRule(name, cert, lhs, rhs, domain)
	if len(defects) > 0 {
		t.Fatalf("admission failed: %v", defects)
	}
	return rule
}

func TestMain(m *testing.M) {
	flag.Parse()
	if os.Getenv("NEWF_EGGSAT_TEST_HELPER") == "1" {
		runHelper()
		return
	}
	os.Exit(m.Run())
}

func runHelper() {
	var req request
	if err := json.NewDecoder(os.Stdin).Decode(&req); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(2)
	}
	if os.Getenv("NEWF_EGGSAT_TEST_REFUTE") == "1" {
		if len(req.Rules) != 1 {
			fmt.Fprint(os.Stderr, "expected one admitted rule")
			os.Exit(2)
		}
		_ = json.NewEncoder(os.Stdout).Encode(response{
			Schema:       ResponseSchema,
			Engine:       EngineVersion,
			Status:       "completed",
			StopReason:   "Saturated",
			Start:        req.Start,
			Best:         "0",
			OriginalCost: 1,
			BestCost:     1,
			Explanation:  "(Rewrite=> " + req.Rules[0].ID + " 0)",
			Proof: []wireProofStep{{
				Before:    req.Start,
				After:     "0",
				RuleID:    req.Rules[0].ID,
				Direction: "forward",
			}},
		})
		return
	}
	if len(req.Rules) != 1 {
		fmt.Fprint(os.Stderr, "expected one admitted rule")
		os.Exit(2)
	}
	_ = json.NewEncoder(os.Stdout).Encode(response{
		Schema:       ResponseSchema,
		Engine:       EngineVersion,
		Status:       "completed",
		StopReason:   "Saturated",
		Start:        req.Start,
		Best:         "x",
		OriginalCost: 3,
		BestCost:     1,
		Iterations:   1,
		EGraphNodes:  3,
		Explanation:  "(Rewrite=> " + req.Rules[0].ID + " x)",
		Proof: []wireProofStep{{
			Before:    req.Start,
			After:     "x",
			RuleID:    req.Rules[0].ID,
			Direction: "forward",
		}},
	})
}
