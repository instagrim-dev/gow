package eggsat

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/finite"
	"github.com/instagrim-dev/newf/internal/store"
)

func TestExternalGraphWithdrawalRebuildsWithoutTheWithdrawnUnion(t *testing.T) {
	binary := os.Getenv("NEWF_EGGSAT_BINARY")
	if binary == "" {
		t.Skip("NEWF_EGGSAT_BINARY is not set")
	}
	searchDomain := finite.Domain{Width: 4, Vars: []string{"x"}}
	start := finite.Unary{Op: finite.OpNot, X: finite.Unary{Op: finite.OpNot, X: finite.Var{Name: "x"}}}
	graph, err := NewGraph(start, searchDomain)
	if err != nil {
		t.Fatal(err)
	}
	ruleDomain := finite.Domain{Width: 4, Vars: []string{"a"}}
	rule := admitted(t, "double-not", finite.Unary{Op: finite.OpNot, X: finite.Unary{Op: finite.OpNot, X: finite.Var{Name: "a"}}}, finite.Var{Name: "a"}, ruleDomain)
	if err := graph.Admit(rule); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 9, 13, 12, 0, 0, 0, time.UTC)
	ledger, err := store.Open(filepath.Join(t.TempDir(), "graph.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer ledger.Close()
	if err := ledger.Migrate(context.Background()); err != nil {
		t.Fatal(err)
	}
	problemID, runID := domain.NewProblemID(now), domain.NewRunID(now)
	if _, _, err := ledger.CreateProblemWithRun(context.Background(), domain.NewProblem{
		ID: problemID, Slug: "external-graph", Statement: "withdrawal rebuild test", Status: domain.ProblemStatusActive, CreatedAt: now, CreatedByRunID: runID,
	}, domain.NewRun{
		ID: runID, ProblemID: problemID, Operation: "test", Status: domain.RunStatusInitialized, InputRef: "test", ToolName: "test", ToolVersion: "test", StartedAt: now, CompletedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	graphID := domain.NewEqualityGraphID(now)
	if err := graph.Persist(context.Background(), ledger, graphID, problemID, now); err != nil {
		t.Fatal(err)
	}
	client := Client{Binary: binary}
	if err := graph.RecordUnion(context.Background(), ledger, graphID, domain.NewEqualityGraphEventID(now.Add(time.Minute)), rule.Identity(), now.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	withRule, err := graph.RebuildAndRecord(context.Background(), ledger, graphID, domain.NewEqualityGraphEventID(now.Add(2*time.Minute)), "rebuild", client, Limits{Timeout: time.Second}, now.Add(2*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if withRule.Best != "x" || !withRule.EndpointVerified || !withRule.MeasuredExecutionKnown || withRule.MeasuredExecution.NodeVisits != 16 {
		t.Fatalf("admitted graph result = %+v", withRule)
	}
	if err := graph.WithdrawAndRecord(context.Background(), ledger, graphID, domain.NewEqualityGraphEventID(now.Add(3*time.Minute)), rule.Identity(), now.Add(3*time.Minute)); err != nil {
		t.Fatal(err)
	}
	withoutRule, err := graph.RebuildAndRecord(context.Background(), ledger, graphID, domain.NewEqualityGraphEventID(now.Add(4*time.Minute)), "reassess", client, Limits{Timeout: time.Second}, now.Add(4*time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	if withoutRule.Best != "not(not(x))" || !withoutRule.EndpointVerified || len(withoutRule.Proof) != 0 || !withoutRule.MeasuredExecutionKnown || withoutRule.MeasuredExecution.NodeVisits != 48 {
		t.Fatalf("withdrawn graph result = %+v", withoutRule)
	}
	recorded, err := ledger.GetEqualityGraph(context.Background(), graphID)
	if err != nil {
		t.Fatal(err)
	}
	if len(recorded.ActiveRuleIdentities) != 0 || len(recorded.Events) != 4 || recorded.Events[3].MeasuredExecutionVisits != 48 {
		t.Fatalf("withdrawal ledger = %+v", recorded)
	}
}
