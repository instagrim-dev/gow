package store

import (
	"context"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/domain"
)

func TestEqualityGraphLedgerDerivesWithdrawalScopeAndRetainsCostLedgers(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 13, 10, 0, 0, 0, time.UTC)
	st, problemID := seedEqualityGraphProblem(t, now)
	graph := EqualityGraphRow{
		ID: domain.NewEqualityGraphID(now), ProblemID: problemID,
		DomainJSON: `{"width":4,"variables":["x"]}`, StartTerm: "(not (not x))",
		EngineRef: "egg/0.11.0", CreatedAt: now.Format(time.RFC3339),
	}
	rule := EqualityGraphRuleRow{
		RuleIdentity: "double-not|w4|vars=[a]|not(not(a))=>a", RuleName: "double-not",
		LeftTerm: "not(not(a))", RightTerm: "a", WarrantJSON: `{"verdict":"HOLDS_ON_DECLARED_DOMAIN"}`,
		CreatedAt: now.Format(time.RFC3339),
	}
	if err := st.PersistEqualityGraph(ctx, graph, []EqualityGraphRuleRow{rule}); err != nil {
		t.Fatal(err)
	}
	union, err := st.AppendEqualityGraphEvent(ctx, EqualityGraphEventRow{
		ID: domain.NewEqualityGraphEventID(now.Add(time.Minute)), GraphID: graph.ID, Kind: "union", RuleIdentity: rule.RuleIdentity,
		CreatedAt: now.Add(time.Minute).Format(time.RFC3339),
	})
	if err != nil || union.Ordinal != 1 {
		t.Fatalf("union = %+v, %v", union, err)
	}
	rebuild, err := st.AppendEqualityGraphEvent(ctx, EqualityGraphEventRow{
		ID: domain.NewEqualityGraphEventID(now.Add(2 * time.Minute)), GraphID: graph.ID, Kind: "rebuild", Dependencies: []string{rule.RuleIdentity},
		PredictedExtractionCost: 1, MeasuredExecutionVisits: 16, MeasurementKnown: true, ResultJSON: `{"best":"x"}`,
		CreatedAt: now.Add(2 * time.Minute).Format(time.RFC3339),
	})
	if err != nil || rebuild.Ordinal != 2 {
		t.Fatalf("rebuild = %+v, %v", rebuild, err)
	}
	if _, err := st.AppendEqualityGraphEvent(ctx, EqualityGraphEventRow{
		ID: domain.NewEqualityGraphEventID(now.Add(3 * time.Minute)), GraphID: graph.ID, Kind: "withdraw", RuleIdentity: rule.RuleIdentity,
		CreatedAt: now.Add(3 * time.Minute).Format(time.RFC3339),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := st.AppendEqualityGraphEvent(ctx, EqualityGraphEventRow{
		ID: domain.NewEqualityGraphEventID(now.Add(4 * time.Minute)), GraphID: graph.ID, Kind: "withdraw", RuleIdentity: rule.RuleIdentity,
		CreatedAt: now.Add(4 * time.Minute).Format(time.RFC3339),
	}); err == nil {
		t.Fatal("a second withdrawal of the same rule must be refused")
	}
	reassess, err := st.AppendEqualityGraphEvent(ctx, EqualityGraphEventRow{
		ID: domain.NewEqualityGraphEventID(now.Add(5 * time.Minute)), GraphID: graph.ID, Kind: "reassess", Dependencies: []string{},
		PredictedExtractionCost: 3, MeasuredExecutionVisits: 48, MeasurementKnown: true, ResultJSON: `{"best":"(not (not x))"}`,
		CreatedAt: now.Add(5 * time.Minute).Format(time.RFC3339),
	})
	if err != nil || reassess.Ordinal != 4 {
		t.Fatalf("reassess = %+v, %v", reassess, err)
	}
	recorded, err := st.GetEqualityGraph(ctx, graph.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(recorded.Rules) != 1 || len(recorded.Events) != 4 || len(recorded.ActiveRuleIdentities) != 0 {
		t.Fatalf("graph ledger = %+v", recorded)
	}
	if !recorded.Events[1].MeasurementKnown || recorded.Events[1].PredictedExtractionCost != 1 || recorded.Events[1].MeasuredExecutionVisits != 16 {
		t.Fatalf("rebuild cost ledger was not retained: %+v", recorded.Events[1])
	}
	if _, err := st.db.ExecContext(ctx, `UPDATE equality_graph_events SET kind = 'union' WHERE id = ?`, reassess.ID); err == nil {
		t.Fatal("equality graph events must be immutable")
	}
}

func TestEqualityGraphRebuildRequiresExactActiveDependenciesAndMeasurement(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 13, 11, 0, 0, 0, time.UTC)
	st, problemID := seedEqualityGraphProblem(t, now)
	graph := EqualityGraphRow{ID: domain.NewEqualityGraphID(now), ProblemID: problemID, DomainJSON: `{}`, StartTerm: "x", EngineRef: "egg/0.11.0", CreatedAt: now.Format(time.RFC3339)}
	rule := EqualityGraphRuleRow{RuleIdentity: "r", RuleName: "r", LeftTerm: "x", RightTerm: "x", WarrantJSON: `{}`, CreatedAt: now.Format(time.RFC3339)}
	if err := st.PersistEqualityGraph(ctx, graph, []EqualityGraphRuleRow{rule}); err != nil {
		t.Fatal(err)
	}
	base := EqualityGraphEventRow{GraphID: graph.ID, Kind: "rebuild", Dependencies: []string{}, CreatedAt: now.Add(time.Minute).Format(time.RFC3339)}
	base.ID = domain.NewEqualityGraphEventID(now.Add(time.Minute))
	if _, err := st.AppendEqualityGraphEvent(ctx, base); err == nil {
		t.Fatal("rebuild without active dependency and measured cost must be refused")
	}
	base.ID = domain.NewEqualityGraphEventID(now.Add(2 * time.Minute))
	base.Dependencies = []string{rule.RuleIdentity}
	base.MeasurementKnown = true
	base.ResultJSON = `{}`
	if _, err := st.AppendEqualityGraphEvent(ctx, base); err != nil {
		t.Fatalf("exact active dependency scope with a known measurement must pass: %v", err)
	}
	base.ID = domain.NewEqualityGraphEventID(now.Add(3 * time.Minute))
	base.ResultJSON = "not-json"
	if _, err := st.AppendEqualityGraphEvent(ctx, base); err == nil {
		t.Fatal("a rebuild result must be valid JSON")
	}
}

func seedEqualityGraphProblem(t *testing.T, now time.Time) (*Store, string) {
	t.Helper()
	st := openMigratedStore(t)
	ctx := context.Background()
	problemID, runID := domain.NewProblemID(now), domain.NewRunID(now)
	if _, _, err := st.CreateProblemWithRun(ctx, domain.NewProblem{
		ID: problemID, Slug: "equality-graph", Statement: "bounded equality graph", Status: domain.ProblemStatusActive,
		CreatedAt: now, CreatedByRunID: runID,
	}, domain.NewRun{
		ID: runID, ProblemID: problemID, Operation: "test", Status: domain.RunStatusInitialized,
		InputRef: "test", ToolName: "test", ToolVersion: "test", StartedAt: now, CompletedAt: now,
	}); err != nil {
		t.Fatalf("seed equality graph problem: %v", err)
	}
	return st, problemID
}
