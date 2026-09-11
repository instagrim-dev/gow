package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/instagrim-dev/newf/internal/domain"
)

// nullIfEmpty returns a NULL sql value for empty strings, else the string. Used
// so optional FK/text columns store NULL rather than "".
func nullIfEmpty(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

// now returns the current UTC time for minting fallback ids.
func now() time.Time { return time.Now().UTC() }

// EvaluationProviderInvocation is the model-tier evaluation invocation recorded
// in the shared provider_invocations table with role='evaluate' (permitted
// after the v15 role generalization). It is written ONLY for model-judgment
// evaluations; deterministic tiers record tool identity on the evaluation row
// instead and carry no provider invocation.
type EvaluationProviderInvocation struct {
	ID              string
	RunID           string
	ProviderName    string
	ProviderVersion string
	ModelName       string
	SchemaVersion   string
	RequestHash     string
	RequestPayload  string
	ResponsePayload string
	CreatedAt       string
}

// EvaluationMetricRow is one typed metric attached to an evaluation. Exactly one
// of the value columns is set per metric_scale (enforced by the table CHECK).
type EvaluationMetricRow struct {
	ID                  string
	MetricName          string
	MetricScale         string // 'numeric'|'ordinal'|'categorical'
	NumericValue        sql.NullFloat64
	OrdinalValue        string
	OrdinalScaleKey     string
	OrdinalScaleVersion string
	CategoricalValue    string
	Comparator          string
}

// EvaluationTargetVerdictRow is one per-target break verdict an evaluation
// computed against its assessed signature content revision (v26).
type EvaluationTargetVerdictRow struct {
	InvariantID string
	Verdict     string // 'satisfies'|'violates'|'unknown'
	Violated    bool
}

// EvaluationRow is one persisted evaluation: the verdict PLUS its verifier kind
// and verification strength (KTD-1/R1), optional confidence ordinal, and either
// deterministic tool identity or a model-tier provider invocation.
type EvaluationRow struct {
	ID                   string
	ProposalID           string
	Verdict              string
	VerifierKind         string
	VerificationStrength string
	ConfidenceOrdinal    string
	ToolName             string
	ToolVersion          string
	ProviderInvocationID string
	Notes                string
	// SignatureContentHash is the EXACT signature content revision the verifier
	// assessed, supplied by the evaluation stage (round-2 F1). The writer stores
	// this value verbatim — never an independent "latest revision" lookup, which
	// could stamp the result with bytes the verifier never saw. Empty for
	// pre-v17 proposals without persisted content (attribution gap, recorded).
	SignatureContentHash string
	// TargetVerdicts are the per-target break verdicts this evaluation ACTUALLY
	// computed against the assessed content revision (v26). Success-cohort
	// admission consumes THESE, never the origin-time frontier_target_invariants
	// flags, so a revised interpretation cannot ride an origin break verdict
	// (or be excluded by one).
	TargetVerdicts []EvaluationTargetVerdictRow
	// Invocation, when set (model tier), is written to provider_invocations with
	// role='evaluate' and its ID linked from the evaluation row.
	Invocation *EvaluationProviderInvocation
	Metrics    []EvaluationMetricRow
}

// EvaluationRunRecord is the full persisted evaluation pass.
type EvaluationRunRecord struct {
	ID                      string
	ProblemID               string
	RunID                   string
	FrontierGenerationRunID string
	InvariantRevisionID     string
	ClusterRunID            string
	NormalizationRevisionID string
	Mode                    string // must be 'proposal' in M5.2
	RoutingPolicy           string
	EvaluationCount         int
	CreatedAt               string
	Evaluations             []EvaluationRow
}

// PersistEvaluationRun writes an evaluation pass transactionally. For each
// evaluation it (a) optionally writes the model-tier provider invocation
// (role='evaluate'), (b) writes the evaluation row with verdict + kind +
// strength, (c) writes its metrics, (d) **populates the proposal's `result` in
// the SAME transaction** (R5), and (e) writes an `evaluated_failures` marker for
// failure/partial_failure verdicts (R6). Re-evaluation is a NEW run (append-only,
// R8); the proposal `result` mirrors the latest verdict written here.
func (s *Store) PersistEvaluationRun(ctx context.Context, record EvaluationRunRecord) (EvaluationRunRecord, error) {
	if err := domain.ValidateEvaluationRunID(record.ID); err != nil {
		return EvaluationRunRecord{}, err
	}
	if err := domain.ValidateProblemID(record.ProblemID); err != nil {
		return EvaluationRunRecord{}, err
	}
	if record.Mode == "" {
		record.Mode = "proposal"
	}
	if record.RoutingPolicy == "" {
		record.RoutingPolicy = "cheap-first"
	}
	record.EvaluationCount = len(record.Evaluations)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return EvaluationRunRecord{}, err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
INSERT INTO evaluation_runs(id, problem_id, run_id, frontier_generation_run_id, invariant_revision_id, cluster_run_id, normalization_revision_id, holdout_set_id, holdout_leakage_check_id, mode, routing_policy, evaluation_count, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?, NULL, NULL, ?, ?, ?, ?)
`, record.ID, record.ProblemID, record.RunID, nullIfEmpty(record.FrontierGenerationRunID), nullIfEmpty(record.InvariantRevisionID), nullIfEmpty(record.ClusterRunID), nullIfEmpty(record.NormalizationRevisionID), record.Mode, record.RoutingPolicy, record.EvaluationCount, record.CreatedAt); err != nil {
		return EvaluationRunRecord{}, err
	}

	for _, e := range record.Evaluations {
		providerInvocationID := nullIfEmpty(e.ProviderInvocationID)
		if e.Invocation != nil {
			inv := e.Invocation
			if _, err := tx.ExecContext(ctx, `
INSERT INTO provider_invocations(id, run_id, role, provider_name, provider_version, model_name, schema_version, request_hash, request_payload, response_payload, created_at)
VALUES(?, ?, 'evaluate', ?, ?, ?, ?, ?, ?, ?, ?)
`, inv.ID, inv.RunID, inv.ProviderName, inv.ProviderVersion, inv.ModelName, inv.SchemaVersion, inv.RequestHash, inv.RequestPayload, inv.ResponsePayload, inv.CreatedAt); err != nil {
				return EvaluationRunRecord{}, err
			}
			providerInvocationID = sql.NullString{String: inv.ID, Valid: true}
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO evaluations(id, evaluation_run_id, proposal_id, verdict, verifier_kind, verification_strength, confidence_ordinal, tool_name, tool_version, provider_invocation_id, notes, created_at, signature_content_hash)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, e.ID, record.ID, nullIfEmpty(e.ProposalID), e.Verdict, e.VerifierKind, e.VerificationStrength, nullIfEmpty(e.ConfidenceOrdinal), nullIfEmpty(e.ToolName), nullIfEmpty(e.ToolVersion), providerInvocationID, nullIfEmpty(e.Notes), record.CreatedAt, e.SignatureContentHash); err != nil {
			return EvaluationRunRecord{}, err
		}
		// v26: the per-target break verdicts computed against the assessed
		// content revision — the assessment-context tuple cohort admission
		// consumes.
		for _, tv := range e.TargetVerdicts {
			if _, err := tx.ExecContext(ctx, `
INSERT INTO evaluation_target_verdicts(evaluation_id, invariant_id, verdict, violated)
VALUES(?, ?, ?, ?)
`, e.ID, tv.InvariantID, tv.Verdict, boolToInt(tv.Violated)); err != nil {
				return EvaluationRunRecord{}, err
			}
		}
		for _, m := range e.Metrics {
			metricID := m.ID
			if metricID == "" {
				metricID = domain.NewEvaluationMetricID(now())
			}
			if _, err := tx.ExecContext(ctx, `
INSERT INTO evaluation_metrics(id, evaluation_id, metric_name, metric_scale, numeric_value, ordinal_value, ordinal_scale_key, ordinal_scale_version, categorical_value, comparator, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, metricID, e.ID, m.MetricName, m.MetricScale, m.NumericValue, nullIfEmpty(m.OrdinalValue), nullIfEmpty(m.OrdinalScaleKey), nullIfEmpty(m.OrdinalScaleVersion), nullIfEmpty(m.CategoricalValue), nullIfEmpty(m.Comparator), record.CreatedAt); err != nil {
				return EvaluationRunRecord{}, err
			}
		}
		// R5: populate the proposal result in the SAME transaction. The frontier
		// proposal immutability trigger permits a one-time NULL -> verdict set.
		if e.ProposalID != "" {
			if _, err := tx.ExecContext(ctx, `UPDATE frontier_proposals SET result = ? WHERE id = ? AND result IS NULL`, e.Verdict, e.ProposalID); err != nil {
				return EvaluationRunRecord{}, err
			}
			// R6: failure/partial_failure re-enters the atlas as a queryable marker.
			if e.Verdict == "failure" || e.Verdict == "partial_failure" {
				if _, err := tx.ExecContext(ctx, `
INSERT OR IGNORE INTO evaluated_failures(proposal_id, evaluation_id, problem_id, verdict, created_at)
VALUES(?, ?, ?, ?, ?)
`, e.ProposalID, e.ID, record.ProblemID, e.Verdict, record.CreatedAt); err != nil {
					return EvaluationRunRecord{}, err
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return EvaluationRunRecord{}, err
	}
	return record, nil
}

// GetEvaluationRun loads a full evaluation pass by id.
func (s *Store) GetEvaluationRun(ctx context.Context, id string) (EvaluationRunRecord, error) {
	if err := domain.ValidateEvaluationRunID(id); err != nil {
		return EvaluationRunRecord{}, err
	}
	return s.loadEvaluationRun(ctx, id)
}

func (s *Store) loadEvaluationRun(ctx context.Context, id string) (EvaluationRunRecord, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, problem_id, run_id, COALESCE(frontier_generation_run_id,''), COALESCE(invariant_revision_id,''), COALESCE(cluster_run_id,''), COALESCE(normalization_revision_id,''), mode, routing_policy, evaluation_count, created_at
FROM evaluation_runs WHERE id = ?
`, id)
	var rec EvaluationRunRecord
	if err := row.Scan(&rec.ID, &rec.ProblemID, &rec.RunID, &rec.FrontierGenerationRunID, &rec.InvariantRevisionID, &rec.ClusterRunID, &rec.NormalizationRevisionID, &rec.Mode, &rec.RoutingPolicy, &rec.EvaluationCount, &rec.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return EvaluationRunRecord{}, fmt.Errorf("%w: evaluation run %s", ErrNotFound, id)
		}
		return EvaluationRunRecord{}, err
	}
	evalRows, err := s.db.QueryContext(ctx, `
SELECT id, COALESCE(proposal_id,''), verdict, verifier_kind, verification_strength, COALESCE(confidence_ordinal,''), COALESCE(tool_name,''), COALESCE(tool_version,''), COALESCE(provider_invocation_id,''), COALESCE(notes,''), COALESCE(signature_content_hash,'')
FROM evaluations WHERE evaluation_run_id = ? ORDER BY id
`, id)
	if err != nil {
		return EvaluationRunRecord{}, err
	}
	defer evalRows.Close()
	for evalRows.Next() {
		var e EvaluationRow
		if err := evalRows.Scan(&e.ID, &e.ProposalID, &e.Verdict, &e.VerifierKind, &e.VerificationStrength, &e.ConfidenceOrdinal, &e.ToolName, &e.ToolVersion, &e.ProviderInvocationID, &e.Notes, &e.SignatureContentHash); err != nil {
			return EvaluationRunRecord{}, err
		}
		rec.Evaluations = append(rec.Evaluations, e)
	}
	if err := evalRows.Err(); err != nil {
		return EvaluationRunRecord{}, err
	}
	for i := range rec.Evaluations {
		if err := s.loadEvaluationMetrics(ctx, &rec.Evaluations[i]); err != nil {
			return EvaluationRunRecord{}, err
		}
		if err := s.loadEvaluationTargetVerdicts(ctx, &rec.Evaluations[i]); err != nil {
			return EvaluationRunRecord{}, err
		}
	}
	return rec, nil
}

// loadEvaluationTargetVerdicts rehydrates the assessment-context break
// verdicts (v26) so audit and regression paths can read exactly what the
// evaluation computed.
func (s *Store) loadEvaluationTargetVerdicts(ctx context.Context, e *EvaluationRow) error {
	rows, err := s.db.QueryContext(ctx, `
SELECT invariant_id, verdict, violated
FROM evaluation_target_verdicts WHERE evaluation_id = ? ORDER BY invariant_id
`, e.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var tv EvaluationTargetVerdictRow
		var violated int
		if err := rows.Scan(&tv.InvariantID, &tv.Verdict, &violated); err != nil {
			return err
		}
		tv.Violated = violated != 0
		e.TargetVerdicts = append(e.TargetVerdicts, tv)
	}
	return rows.Err()
}

func (s *Store) loadEvaluationMetrics(ctx context.Context, e *EvaluationRow) error {
	rows, err := s.db.QueryContext(ctx, `
SELECT metric_name, metric_scale, numeric_value, COALESCE(ordinal_value,''), COALESCE(ordinal_scale_key,''), COALESCE(ordinal_scale_version,''), COALESCE(categorical_value,''), COALESCE(comparator,'')
FROM evaluation_metrics WHERE evaluation_id = ? ORDER BY metric_name
`, e.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var m EvaluationMetricRow
		if err := rows.Scan(&m.MetricName, &m.MetricScale, &m.NumericValue, &m.OrdinalValue, &m.OrdinalScaleKey, &m.OrdinalScaleVersion, &m.CategoricalValue, &m.Comparator); err != nil {
			return err
		}
		e.Metrics = append(e.Metrics, m)
	}
	return rows.Err()
}

// ListEvaluations returns the evaluations for a problem, newest run first, as
// flat rows joined with their run id.
func (s *Store) ListEvaluations(ctx context.Context, problemID string) ([]EvaluationRunRecord, error) {
	if err := domain.ValidateProblemID(problemID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT id FROM evaluation_runs WHERE problem_id = ? ORDER BY created_at DESC, id DESC
`, problemID)
	if err != nil {
		return nil, err
	}
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		ids = append(ids, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	out := make([]EvaluationRunRecord, 0, len(ids))
	for _, id := range ids {
		rec, err := s.loadEvaluationRun(ctx, id)
		if err != nil {
			return nil, err
		}
		out = append(out, rec)
	}
	return out, nil
}

// GetEvaluation loads a single evaluation row (with metrics) by id.
func (s *Store) GetEvaluation(ctx context.Context, id string) (EvaluationRow, error) {
	if err := domain.ValidateEvaluationID(id); err != nil {
		return EvaluationRow{}, err
	}
	row := s.db.QueryRowContext(ctx, `
SELECT id, COALESCE(proposal_id,''), verdict, verifier_kind, verification_strength, COALESCE(confidence_ordinal,''), COALESCE(tool_name,''), COALESCE(tool_version,''), COALESCE(provider_invocation_id,''), COALESCE(notes,''), COALESCE(signature_content_hash,'')
FROM evaluations WHERE id = ?
`, id)
	var e EvaluationRow
	if err := row.Scan(&e.ID, &e.ProposalID, &e.Verdict, &e.VerifierKind, &e.VerificationStrength, &e.ConfidenceOrdinal, &e.ToolName, &e.ToolVersion, &e.ProviderInvocationID, &e.Notes, &e.SignatureContentHash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return EvaluationRow{}, fmt.Errorf("%w: evaluation %s", ErrNotFound, id)
		}
		return EvaluationRow{}, err
	}
	if err := s.loadEvaluationMetrics(ctx, &e); err != nil {
		return EvaluationRow{}, err
	}
	return e, nil
}

// ListEvaluatedFailures returns the proposals a prior evaluation marked as
// re-entered failures for a problem (R6): the input a subsequent cluster build
// may include as newly-evaluated failure structure.
func (s *Store) ListEvaluatedFailures(ctx context.Context, problemID string) ([]EvaluatedFailureRow, error) {
	if err := domain.ValidateProblemID(problemID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT proposal_id, evaluation_id, verdict, created_at FROM evaluated_failures WHERE problem_id = ? ORDER BY created_at, proposal_id
`, problemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EvaluatedFailureRow
	for rows.Next() {
		var r EvaluatedFailureRow
		if err := rows.Scan(&r.ProposalID, &r.EvaluationID, &r.Verdict, &r.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// EvaluatedFailureRow is one re-entered failure marker.
type EvaluatedFailureRow struct {
	ProposalID   string
	EvaluationID string
	Verdict      string
	CreatedAt    string
}
