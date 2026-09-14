package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"sort"

	"github.com/instagrim-dev/newf/internal/domain"
)

// EqualityGraphRow is the durable root of one bounded equality-saturation
// graph. Domain and start are retained as canonical JSON/text because graph
// semantics must be reconstructible without an in-memory engine.
type EqualityGraphRow struct {
	ID         string
	ProblemID  string
	DomainJSON string
	StartTerm  string
	EngineRef  string
	CreatedAt  string
}

// EqualityGraphRuleRow is one immutable, independently warranted dependency.
type EqualityGraphRuleRow struct {
	GraphID      string
	RuleIdentity string
	RuleName     string
	LeftTerm     string
	RightTerm    string
	WarrantJSON  string
	CreatedAt    string
}

// EqualityGraphEventRow records a graph state transition or a checked engine
// result. Dependencies are the exact active rule identities used by rebuild
// and reassessment events, not a mutable current-rule column.
type EqualityGraphEventRow struct {
	ID                      string
	GraphID                 string
	Ordinal                 int
	Kind                    string // union|withdraw|rebuild|reassess
	RuleIdentity            string
	Dependencies            []string
	PredictedExtractionCost int64
	MeasuredExecutionVisits int64
	MeasurementKnown        bool
	ResultJSON              string
	CreatedAt               string
}

// EqualityGraphRecord reads a graph with its complete dependency and event
// history. ActiveRuleIdentities is derived rather than stored.
type EqualityGraphRecord struct {
	Graph                EqualityGraphRow
	Rules                []EqualityGraphRuleRow
	Events               []EqualityGraphEventRow
	ActiveRuleIdentities []string
}

// PersistEqualityGraph writes a graph root and all admitted rule dependencies
// atomically. A partial dependency scope would make later withdrawal claims
// unauditable, so it is never committed.
func (s *Store) PersistEqualityGraph(ctx context.Context, graph EqualityGraphRow, rules []EqualityGraphRuleRow) error {
	if err := validateEqualityGraph(graph); err != nil {
		return err
	}
	if len(rules) == 0 {
		return fmt.Errorf("equality graph requires at least one admitted rule")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var present int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM problems WHERE id = ?`, graph.ProblemID).Scan(&present); err != nil {
		return err
	}
	if present == 0 {
		return fmt.Errorf("%w: problem %s", ErrNotFound, graph.ProblemID)
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO equality_graphs(id, problem_id, domain_json, start_term, engine_ref, created_at)
VALUES(?, ?, ?, ?, ?, ?)
`, graph.ID, graph.ProblemID, graph.DomainJSON, graph.StartTerm, graph.EngineRef, graph.CreatedAt); err != nil {
		return fmt.Errorf("persist equality graph: %w", err)
	}
	seen := make(map[string]bool, len(rules))
	for _, rule := range rules {
		rule.GraphID = graph.ID
		if err := validateEqualityGraphRule(rule); err != nil {
			return err
		}
		if seen[rule.RuleIdentity] {
			return fmt.Errorf("duplicate equality graph rule identity %q", rule.RuleIdentity)
		}
		seen[rule.RuleIdentity] = true
		if _, err := tx.ExecContext(ctx, `
INSERT INTO equality_graph_rules(graph_id, rule_identity, rule_name, left_term, right_term, warrant_json, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?)
`, rule.GraphID, rule.RuleIdentity, rule.RuleName, rule.LeftTerm, rule.RightTerm, rule.WarrantJSON, rule.CreatedAt); err != nil {
			return fmt.Errorf("persist equality graph rule: %w", err)
		}
	}
	return tx.Commit()
}

// AppendEqualityGraphEvent appends one transition after checking the current
// derived scope. A withdrawal cannot name a rule twice; a union cannot cite a
// withdrawn or unknown rule; rebuild/reassess must name exactly every active
// dependency and retain both cost ledgers.
func (s *Store) AppendEqualityGraphEvent(ctx context.Context, event EqualityGraphEventRow) (EqualityGraphEventRow, error) {
	if err := validateEqualityGraphEvent(event); err != nil {
		return EqualityGraphEventRow{}, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return EqualityGraphEventRow{}, err
	}
	defer tx.Rollback()
	active, err := activeEqualityRulesTx(ctx, tx, event.GraphID)
	if err != nil {
		return EqualityGraphEventRow{}, err
	}
	if err := validateEventScope(event, active); err != nil {
		return EqualityGraphEventRow{}, err
	}
	var ordinal int
	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(MAX(ordinal), 0) + 1 FROM equality_graph_events WHERE graph_id = ?`, event.GraphID).Scan(&ordinal); err != nil {
		return EqualityGraphEventRow{}, err
	}
	event.Ordinal = ordinal
	deps, err := json.Marshal(sortedUnique(event.Dependencies))
	if err != nil {
		return EqualityGraphEventRow{}, err
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO equality_graph_events(id, graph_id, ordinal, kind, rule_identity, dependency_json, predicted_extraction_cost, measured_execution_visits, measurement_known, result_json, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, event.ID, event.GraphID, event.Ordinal, event.Kind, event.RuleIdentity, string(deps), nullableCost(event.PredictedExtractionCost, event.MeasurementKnown), nullableCost(event.MeasuredExecutionVisits, event.MeasurementKnown), boolInt(event.MeasurementKnown), event.ResultJSON, event.CreatedAt); err != nil {
		return EqualityGraphEventRow{}, fmt.Errorf("append equality graph event: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return EqualityGraphEventRow{}, err
	}
	event.Dependencies = sortedUnique(event.Dependencies)
	return event, nil
}

// GetEqualityGraph returns the immutable graph ledger and its derived active
// rule scope after every recorded withdrawal.
func (s *Store) GetEqualityGraph(ctx context.Context, id string) (EqualityGraphRecord, error) {
	if err := domain.ValidateEqualityGraphID(id); err != nil {
		return EqualityGraphRecord{}, err
	}
	var record EqualityGraphRecord
	if err := s.db.QueryRowContext(ctx, `
SELECT id, problem_id, domain_json, start_term, engine_ref, created_at
FROM equality_graphs WHERE id = ?
`, id).Scan(&record.Graph.ID, &record.Graph.ProblemID, &record.Graph.DomainJSON, &record.Graph.StartTerm, &record.Graph.EngineRef, &record.Graph.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return EqualityGraphRecord{}, fmt.Errorf("%w: equality graph %s", ErrNotFound, id)
		}
		return EqualityGraphRecord{}, err
	}
	rules, err := s.db.QueryContext(ctx, `
SELECT graph_id, rule_identity, rule_name, left_term, right_term, warrant_json, created_at
FROM equality_graph_rules WHERE graph_id = ? ORDER BY rule_identity
`, id)
	if err != nil {
		return EqualityGraphRecord{}, err
	}
	defer rules.Close()
	for rules.Next() {
		var rule EqualityGraphRuleRow
		if err := rules.Scan(&rule.GraphID, &rule.RuleIdentity, &rule.RuleName, &rule.LeftTerm, &rule.RightTerm, &rule.WarrantJSON, &rule.CreatedAt); err != nil {
			return EqualityGraphRecord{}, err
		}
		record.Rules = append(record.Rules, rule)
	}
	if err := rules.Err(); err != nil {
		return EqualityGraphRecord{}, err
	}
	events, err := s.db.QueryContext(ctx, `
SELECT id, graph_id, ordinal, kind, rule_identity, dependency_json,
       COALESCE(predicted_extraction_cost, 0), COALESCE(measured_execution_visits, 0), measurement_known, result_json, created_at
FROM equality_graph_events WHERE graph_id = ? ORDER BY ordinal
`, id)
	if err != nil {
		return EqualityGraphRecord{}, err
	}
	defer events.Close()
	for events.Next() {
		var event EqualityGraphEventRow
		var raw string
		var known int
		if err := events.Scan(&event.ID, &event.GraphID, &event.Ordinal, &event.Kind, &event.RuleIdentity, &raw, &event.PredictedExtractionCost, &event.MeasuredExecutionVisits, &known, &event.ResultJSON, &event.CreatedAt); err != nil {
			return EqualityGraphRecord{}, err
		}
		if err := json.Unmarshal([]byte(raw), &event.Dependencies); err != nil {
			return EqualityGraphRecord{}, fmt.Errorf("equality graph event %s has invalid dependency JSON: %w", event.ID, err)
		}
		event.MeasurementKnown = known == 1
		record.Events = append(record.Events, event)
	}
	if err := events.Err(); err != nil {
		return EqualityGraphRecord{}, err
	}
	active, err := activeEqualityRules(ctx, s.db, id)
	if err != nil {
		return EqualityGraphRecord{}, err
	}
	record.ActiveRuleIdentities = active
	return record, nil
}

func validateEqualityGraph(graph EqualityGraphRow) error {
	if err := domain.ValidateEqualityGraphID(graph.ID); err != nil {
		return err
	}
	if err := domain.ValidateProblemID(graph.ProblemID); err != nil {
		return err
	}
	if graph.DomainJSON == "" || graph.StartTerm == "" || graph.EngineRef == "" || graph.CreatedAt == "" {
		return errors.New("equality graph requires domain, start term, engine reference, and creation time")
	}
	if !json.Valid([]byte(graph.DomainJSON)) {
		return errors.New("equality graph domain is not valid JSON")
	}
	return nil
}

func validateEqualityGraphRule(rule EqualityGraphRuleRow) error {
	if rule.RuleIdentity == "" || rule.RuleName == "" || rule.LeftTerm == "" || rule.RightTerm == "" || rule.WarrantJSON == "" || rule.CreatedAt == "" {
		return errors.New("equality graph rule requires identity, name, both terms, warrant, and creation time")
	}
	if !json.Valid([]byte(rule.WarrantJSON)) {
		return errors.New("equality graph rule warrant is not valid JSON")
	}
	return nil
}

func validateEqualityGraphEvent(event EqualityGraphEventRow) error {
	if err := domain.ValidateEqualityGraphEventID(event.ID); err != nil {
		return err
	}
	if err := domain.ValidateEqualityGraphID(event.GraphID); err != nil {
		return err
	}
	switch event.Kind {
	case "union", "withdraw", "rebuild", "reassess":
	default:
		return fmt.Errorf("unsupported equality graph event kind %q", event.Kind)
	}
	if event.CreatedAt == "" {
		return errors.New("equality graph event requires creation time")
	}
	if event.MeasurementKnown && (event.PredictedExtractionCost < 0 || event.MeasuredExecutionVisits < 0) {
		return errors.New("equality graph cost ledgers cannot be negative")
	}
	return nil
}

func validateEventScope(event EqualityGraphEventRow, active []string) error {
	canonicalDependencies := sortedUnique(event.Dependencies)
	if len(canonicalDependencies) != len(event.Dependencies) {
		return fmt.Errorf("%s dependency scope contains duplicate rule identities", event.Kind)
	}
	activeSet := make(map[string]bool, len(active))
	for _, id := range active {
		activeSet[id] = true
	}
	switch event.Kind {
	case "union":
		if event.RuleIdentity == "" || !activeSet[event.RuleIdentity] {
			return fmt.Errorf("union must cite one active admitted rule")
		}
	case "withdraw":
		if event.RuleIdentity == "" || !activeSet[event.RuleIdentity] {
			return fmt.Errorf("withdraw must cite one active admitted rule")
		}
	case "rebuild", "reassess":
		if event.RuleIdentity != "" {
			return fmt.Errorf("%s must not cite a single rule identity", event.Kind)
		}
		if !sameStrings(canonicalDependencies, active) {
			return fmt.Errorf("%s dependency scope does not equal the current active rule set", event.Kind)
		}
		if !event.MeasurementKnown {
			return fmt.Errorf("%s requires a known measured execution cost", event.Kind)
		}
		if !json.Valid([]byte(event.ResultJSON)) {
			return fmt.Errorf("%s requires a valid JSON result record", event.Kind)
		}
	}
	return nil
}

type equalityGraphQueryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}

func activeEqualityRules(ctx context.Context, db equalityGraphQueryer, graphID string) ([]string, error) {
	rows, err := db.QueryContext(ctx, `
SELECT r.rule_identity
FROM equality_graph_rules r
WHERE r.graph_id = ?
  AND NOT EXISTS (
    SELECT 1 FROM equality_graph_events e
    WHERE e.graph_id = r.graph_id AND e.kind = 'withdraw' AND e.rule_identity = r.rule_identity
  )
ORDER BY r.rule_identity
`, graphID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func activeEqualityRulesTx(ctx context.Context, tx *sql.Tx, graphID string) ([]string, error) {
	var exists int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM equality_graphs WHERE id = ?`, graphID).Scan(&exists); err != nil {
		return nil, err
	}
	if exists == 0 {
		return nil, fmt.Errorf("%w: equality graph %s", ErrNotFound, graphID)
	}
	return activeEqualityRules(ctx, tx, graphID)
}

func sortedUnique(in []string) []string {
	out := append([]string(nil), in...)
	sort.Strings(out)
	for i := 1; i < len(out); i++ {
		if out[i] == out[i-1] {
			return nil
		}
	}
	return out
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func nullableCost(value int64, known bool) any {
	if !known {
		return nil
	}
	return value
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
