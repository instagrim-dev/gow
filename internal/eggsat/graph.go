package eggsat

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/instagrim-dev/newf/internal/finite"
	"github.com/instagrim-dev/newf/internal/rewrite"
	"github.com/instagrim-dev/newf/internal/store"
)

// Graph is the Go-owned logical dependency graph for one saturation request.
// egg state never survives a run: Rebuild starts a fresh subprocess from this
// active-rule set, which makes a withdrawal observable rather than an attempt
// to undo an opaque in-process union.
type Graph struct {
	start  finite.Expr
	domain finite.Domain
	rules  map[string]rewrite.Rule
}

// NewGraph validates the semantic root before any rule can be admitted.
func NewGraph(start finite.Expr, d finite.Domain) (*Graph, error) {
	if defects := finite.ValidateExpr(start, d); len(defects) > 0 {
		return nil, fmt.Errorf("invalid equality graph start: %v", defects)
	}
	return &Graph{start: start, domain: finite.Domain{Width: d.Width, Vars: append([]string(nil), d.Vars...)}, rules: make(map[string]rewrite.Rule)}, nil
}

// Admit adds one already-warranted rule to the graph. A changed rule cannot
// reuse an identity, and duplicate admission is refused instead of ignored.
func (g *Graph) Admit(rule rewrite.Rule) error {
	if g == nil {
		return fmt.Errorf("nil equality graph")
	}
	if err := rule.ValidateForDomain(g.domain); err != nil {
		return err
	}
	id := rule.Identity()
	if _, exists := g.rules[id]; exists {
		return fmt.Errorf("rule %q is already active in this graph", id)
	}
	g.rules[id] = rule
	return nil
}

// Withdraw removes exactly one admitted dependency. It never mutates a rule
// definition; Rebuild must be called to obtain a candidate under the new
// active set.
func (g *Graph) Withdraw(identity string) error {
	if g == nil {
		return fmt.Errorf("nil equality graph")
	}
	if _, exists := g.rules[identity]; !exists {
		return fmt.Errorf("rule %q is not active in this graph", identity)
	}
	delete(g.rules, identity)
	return nil
}

// ActiveRuleIdentities reports the exact dependency scope in stable order.
func (g *Graph) ActiveRuleIdentities() []string {
	if g == nil {
		return nil
	}
	ids := make([]string, 0, len(g.rules))
	for id := range g.rules {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// Persist records the graph root and its complete admitted dependency scope.
// Call it only after admission: later lifecycle changes are append-only events
// through RecordUnion, WithdrawAndRecord, and RebuildAndRecord.
func (g *Graph) Persist(ctx context.Context, ledger *store.Store, graphID, problemID string, createdAt time.Time) error {
	if g == nil {
		return fmt.Errorf("nil equality graph")
	}
	if ledger == nil {
		return fmt.Errorf("nil equality graph ledger")
	}
	domainJSON, err := json.Marshal(g.domain)
	if err != nil {
		return fmt.Errorf("encode equality graph domain: %w", err)
	}
	rules := make([]store.EqualityGraphRuleRow, 0, len(g.rules))
	for _, id := range g.ActiveRuleIdentities() {
		exported := g.rules[id].Export()
		warrant, err := json.Marshal(exported.Warrant)
		if err != nil {
			return fmt.Errorf("encode rule warrant %q: %w", id, err)
		}
		rules = append(rules, store.EqualityGraphRuleRow{
			RuleIdentity: exported.Identity, RuleName: exported.Name,
			LeftTerm: finite.Render(exported.Left), RightTerm: finite.Render(exported.Right),
			WarrantJSON: string(warrant), CreatedAt: createdAt.UTC().Format(time.RFC3339Nano),
		})
	}
	return ledger.PersistEqualityGraph(ctx, store.EqualityGraphRow{
		ID: graphID, ProblemID: problemID, DomainJSON: string(domainJSON),
		StartTerm: finite.Render(g.start), EngineRef: EngineVersion,
		CreatedAt: createdAt.UTC().Format(time.RFC3339Nano),
	}, rules)
}

// RecordUnion appends the admitted dependency used by an engine union.
func (g *Graph) RecordUnion(ctx context.Context, ledger *store.Store, graphID, eventID, ruleIdentity string, at time.Time) error {
	if g == nil || ledger == nil {
		return fmt.Errorf("equality graph and ledger are required")
	}
	if _, active := g.rules[ruleIdentity]; !active {
		return fmt.Errorf("rule %q is not active in this graph", ruleIdentity)
	}
	_, err := ledger.AppendEqualityGraphEvent(ctx, store.EqualityGraphEventRow{
		ID: eventID, GraphID: graphID, Kind: "union", RuleIdentity: ruleIdentity,
		CreatedAt: at.UTC().Format(time.RFC3339Nano),
	})
	return err
}

// WithdrawAndRecord appends the durable withdrawal before changing the
// in-memory active set, so a failed persistence operation cannot make a rule
// disappear only from the process that happened to attempt it.
func (g *Graph) WithdrawAndRecord(ctx context.Context, ledger *store.Store, graphID, eventID, ruleIdentity string, at time.Time) error {
	if g == nil || ledger == nil {
		return fmt.Errorf("equality graph and ledger are required")
	}
	if _, active := g.rules[ruleIdentity]; !active {
		return fmt.Errorf("rule %q is not active in this graph", ruleIdentity)
	}
	if _, err := ledger.AppendEqualityGraphEvent(ctx, store.EqualityGraphEventRow{
		ID: eventID, GraphID: graphID, Kind: "withdraw", RuleIdentity: ruleIdentity,
		CreatedAt: at.UTC().Format(time.RFC3339Nano),
	}); err != nil {
		return err
	}
	return g.Withdraw(ruleIdentity)
}

// Rebuild runs the external engine from the current active rules, then makes
// the independent proof and endpoint checks through Client. An unmeasured
// finite execution cost is retained as unknown; it cannot be presented as a
// zero-cost realization.
func (g *Graph) Rebuild(ctx context.Context, client Client, limits Limits) (Result, error) {
	if g == nil {
		return Result{}, fmt.Errorf("nil equality graph")
	}
	ids := g.ActiveRuleIdentities()
	rules := make([]rewrite.Rule, 0, len(ids))
	for _, id := range ids {
		rules = append(rules, g.rules[id])
	}
	return client.Optimize(ctx, g.start, g.domain, rules, limits)
}

// RebuildAndRecord runs a fresh engine process from the active scope and
// appends a rebuild/reassessment event carrying the verified result and both
// cost ledgers. An unverified endpoint or unknown execution measurement cannot
// become a durable G2 result.
func (g *Graph) RebuildAndRecord(ctx context.Context, ledger *store.Store, graphID, eventID, kind string, client Client, limits Limits, at time.Time) (Result, error) {
	result, err := g.Rebuild(ctx, client, limits)
	if err != nil {
		return result, err
	}
	if ledger == nil {
		return result, fmt.Errorf("nil equality graph ledger")
	}
	if !result.EndpointVerified {
		return result, fmt.Errorf("refusing to persist %s: endpoint is not independently verified", kind)
	}
	if !result.MeasuredExecutionKnown {
		return result, fmt.Errorf("refusing to persist %s: measured execution cost is unknown", kind)
	}
	payload, err := json.Marshal(result)
	if err != nil {
		return result, fmt.Errorf("encode equality graph %s result: %w", kind, err)
	}
	_, err = ledger.AppendEqualityGraphEvent(ctx, store.EqualityGraphEventRow{
		ID: eventID, GraphID: graphID, Kind: kind, Dependencies: g.ActiveRuleIdentities(),
		PredictedExtractionCost: result.BestCost, MeasuredExecutionVisits: result.MeasuredExecution.NodeVisits,
		MeasurementKnown: true, ResultJSON: string(payload), CreatedAt: at.UTC().Format(time.RFC3339Nano),
	})
	return result, err
}
