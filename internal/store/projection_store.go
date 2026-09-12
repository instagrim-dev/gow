package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/instagrim-dev/newf/internal/domain"
)

// ProjectionArtifactRow is one persisted concrete projection plan revision
// (v37/S5): the authored, typed realization plan for a frontier proposal. The
// steps live in ContentJSON (projection.Artifact, schema-versioned) so the
// exact authored plan is replayable; ContentHash makes identical re-submission
// per proposal a refused duplicate rather than a silent new revision.
type ProjectionArtifactRow struct {
	ID            string
	ProblemID     string
	ProposalID    string
	RunID         string
	Revision      int
	AuthorKind    string // operator|tool
	SchemaVersion string
	ContentJSON   string
	ContentHash   string
	CreatedAt     string
}

// ProjectionObligationRow is one verification obligation derived from an
// artifact: what must be checked before the plan's claim can count. Kind
// 'steps-compose' is decided by code at projection time; 'domain-realization'
// is created only for composing plans and stays OPEN until an external
// observation decides it.
type ProjectionObligationRow struct {
	ID         string
	ArtifactID string
	Ordinal    int
	Kind       string // steps-compose|domain-realization
	Statement  string
	Checker    string // deterministic-check|external
	CreatedAt  string
	// Decision is the appended verdict, nil while the obligation is open.
	Decision *ProjectionObligationDecisionRow
}

// ProjectionObligationDecisionRow is the append-once verdict on one
// obligation. EvidenceKind/EvidenceRef identify the DOMAIN OBSERVATION backing
// an operator decision (an evaluation row id) or 'code-check' for the
// deterministic composition verdict.
type ProjectionObligationDecisionRow struct {
	ObligationID string
	RunID        string
	Status       string // discharged|failed
	DecidedBy    string // code|operator
	Basis        string
	EvidenceKind string // code-check|evaluation ('' = none recorded)
	EvidenceRef  string
	// EvaluationVerdict/EvaluationStrength are the typed verdict and
	// verification strength of the backing evaluation, copied verbatim at
	// decision time (v40, 2026-09-12 review F6). Empty for code decisions
	// (no backing evaluation) and pre-v40 history.
	EvaluationVerdict  string
	EvaluationStrength string
	CreatedAt          string
}

// ProjectionRecord bundles one artifact with its obligations for persistence
// and reads.
type ProjectionRecord struct {
	Artifact    ProjectionArtifactRow
	Obligations []ProjectionObligationRow
}

// GetProposalProblem resolves the owning problem of a frontier proposal via
// its generation run, reporting absence explicitly.
func (s *Store) GetProposalProblem(ctx context.Context, proposalID string) (string, error) {
	if err := domain.ValidateFrontierProposalID(proposalID); err != nil {
		return "", err
	}
	row := s.db.QueryRowContext(ctx, `
SELECT g.problem_id
FROM frontier_proposals p
JOIN frontier_generation_runs g ON g.id = p.frontier_generation_run_id
WHERE p.id = ?
`, proposalID)
	var problemID string
	if err := row.Scan(&problemID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("%w: frontier proposal %s", ErrNotFound, proposalID)
		}
		return "", err
	}
	return problemID, nil
}

// PersistProjection writes one artifact, its obligations, and any synchronous
// code decisions in a single transaction (multi-record provenance). The
// artifact's revision is allocated here (max+1 per proposal); duplicate
// content per proposal is refused loudly.
func (s *Store) PersistProjection(ctx context.Context, rec ProjectionRecord) (ProjectionRecord, error) {
	if err := domain.ValidateProjectionArtifactID(rec.Artifact.ID); err != nil {
		return ProjectionRecord{}, err
	}
	for i := range rec.Obligations {
		if err := domain.ValidateProjectionObligationID(rec.Obligations[i].ID); err != nil {
			return ProjectionRecord{}, err
		}
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ProjectionRecord{}, err
	}
	defer func() { _ = tx.Rollback() }()

	var dup int
	if err := tx.QueryRowContext(ctx, `
SELECT COUNT(*) FROM projection_artifacts WHERE proposal_id = ? AND content_hash = ?
`, rec.Artifact.ProposalID, rec.Artifact.ContentHash).Scan(&dup); err != nil {
		return ProjectionRecord{}, err
	}
	if dup > 0 {
		return ProjectionRecord{}, fmt.Errorf("an identical projection artifact already exists for proposal %s (content hash %s); author a revised plan instead of re-submitting", rec.Artifact.ProposalID, rec.Artifact.ContentHash)
	}
	if err := tx.QueryRowContext(ctx, `
SELECT COALESCE(MAX(revision), 0) + 1 FROM projection_artifacts WHERE proposal_id = ?
`, rec.Artifact.ProposalID).Scan(&rec.Artifact.Revision); err != nil {
		return ProjectionRecord{}, err
	}

	if _, err := tx.ExecContext(ctx, `
INSERT INTO projection_artifacts(id, problem_id, proposal_id, run_id, revision, author_kind, schema_version, content_json, content_hash, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, rec.Artifact.ID, rec.Artifact.ProblemID, rec.Artifact.ProposalID, rec.Artifact.RunID, rec.Artifact.Revision,
		rec.Artifact.AuthorKind, rec.Artifact.SchemaVersion, rec.Artifact.ContentJSON, rec.Artifact.ContentHash, rec.Artifact.CreatedAt); err != nil {
		return ProjectionRecord{}, fmt.Errorf("persist projection artifact: %w", err)
	}
	for i := range rec.Obligations {
		ob := &rec.Obligations[i]
		ob.ArtifactID = rec.Artifact.ID
		if _, err := tx.ExecContext(ctx, `
INSERT INTO projection_obligations(id, artifact_id, ordinal, kind, statement, checker_kind, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?)
`, ob.ID, ob.ArtifactID, ob.Ordinal, ob.Kind, ob.Statement, ob.Checker, ob.CreatedAt); err != nil {
			return ProjectionRecord{}, fmt.Errorf("persist projection obligation: %w", err)
		}
		if ob.Decision != nil {
			d := ob.Decision
			d.ObligationID = ob.ID
			if _, err := tx.ExecContext(ctx, `
INSERT INTO projection_obligation_decisions(obligation_id, run_id, status, decided_by, basis, evidence_kind, evidence_ref, evaluation_verdict, evaluation_strength, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, d.ObligationID, d.RunID, d.Status, d.DecidedBy, d.Basis, nullable(d.EvidenceKind), nullable(d.EvidenceRef), nullable(d.EvaluationVerdict), nullable(d.EvaluationStrength), d.CreatedAt); err != nil {
				return ProjectionRecord{}, fmt.Errorf("persist projection decision: %w", err)
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return ProjectionRecord{}, err
	}
	return rec, nil
}

// PersistObligationDecision appends the verdict for one OPEN obligation. The
// PRIMARY KEY makes re-deciding a decided obligation a loud refusal.
func (s *Store) PersistObligationDecision(ctx context.Context, d ProjectionObligationDecisionRow) error {
	if err := domain.ValidateProjectionObligationID(d.ObligationID); err != nil {
		return err
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO projection_obligation_decisions(obligation_id, run_id, status, decided_by, basis, evidence_kind, evidence_ref, evaluation_verdict, evaluation_strength, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, d.ObligationID, d.RunID, d.Status, d.DecidedBy, d.Basis, nullable(d.EvidenceKind), nullable(d.EvidenceRef), nullable(d.EvaluationVerdict), nullable(d.EvaluationStrength), d.CreatedAt)
	if err != nil {
		return fmt.Errorf("persist obligation decision: %w", err)
	}
	return nil
}

// GetProjectionObligation loads one obligation with its artifact linkage and
// any decision.
func (s *Store) GetProjectionObligation(ctx context.Context, obligationID string) (ProjectionObligationRow, ProjectionArtifactRow, error) {
	if err := domain.ValidateProjectionObligationID(obligationID); err != nil {
		return ProjectionObligationRow{}, ProjectionArtifactRow{}, err
	}
	row := s.db.QueryRowContext(ctx, `
SELECT o.id, o.artifact_id, o.ordinal, o.kind, o.statement, o.checker_kind, o.created_at,
       a.id, a.problem_id, a.proposal_id, a.run_id, a.revision, a.author_kind, a.schema_version, a.content_json, a.content_hash, a.created_at
FROM projection_obligations o
JOIN projection_artifacts a ON a.id = o.artifact_id
WHERE o.id = ?
`, obligationID)
	var ob ProjectionObligationRow
	var art ProjectionArtifactRow
	if err := row.Scan(&ob.ID, &ob.ArtifactID, &ob.Ordinal, &ob.Kind, &ob.Statement, &ob.Checker, &ob.CreatedAt,
		&art.ID, &art.ProblemID, &art.ProposalID, &art.RunID, &art.Revision, &art.AuthorKind, &art.SchemaVersion, &art.ContentJSON, &art.ContentHash, &art.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ProjectionObligationRow{}, ProjectionArtifactRow{}, fmt.Errorf("%w: projection obligation %s", ErrNotFound, obligationID)
		}
		return ProjectionObligationRow{}, ProjectionArtifactRow{}, err
	}
	if err := s.loadObligationDecision(ctx, &ob); err != nil {
		return ProjectionObligationRow{}, ProjectionArtifactRow{}, err
	}
	return ob, art, nil
}

// ListProjectionsForProblem returns every artifact of a problem (revision
// order per proposal, then creation order) with obligations and decisions.
func (s *Store) ListProjectionsForProblem(ctx context.Context, problemID string) ([]ProjectionRecord, error) {
	if err := domain.ValidateProblemID(problemID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT id, problem_id, proposal_id, run_id, revision, author_kind, schema_version, content_json, content_hash, created_at
FROM projection_artifacts WHERE problem_id = ?
ORDER BY proposal_id, revision
`, problemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ProjectionRecord
	for rows.Next() {
		var a ProjectionArtifactRow
		if err := rows.Scan(&a.ID, &a.ProblemID, &a.ProposalID, &a.RunID, &a.Revision, &a.AuthorKind, &a.SchemaVersion, &a.ContentJSON, &a.ContentHash, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, ProjectionRecord{Artifact: a})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range out {
		obs, err := s.listObligationsForArtifact(ctx, out[i].Artifact.ID)
		if err != nil {
			return nil, err
		}
		out[i].Obligations = obs
	}
	return out, nil
}

func (s *Store) listObligationsForArtifact(ctx context.Context, artifactID string) ([]ProjectionObligationRow, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, artifact_id, ordinal, kind, statement, checker_kind, created_at
FROM projection_obligations WHERE artifact_id = ? ORDER BY ordinal
`, artifactID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ProjectionObligationRow
	for rows.Next() {
		var ob ProjectionObligationRow
		if err := rows.Scan(&ob.ID, &ob.ArtifactID, &ob.Ordinal, &ob.Kind, &ob.Statement, &ob.Checker, &ob.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, ob)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range out {
		if err := s.loadObligationDecision(ctx, &out[i]); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (s *Store) loadObligationDecision(ctx context.Context, ob *ProjectionObligationRow) error {
	row := s.db.QueryRowContext(ctx, `
SELECT obligation_id, run_id, status, decided_by, basis, COALESCE(evidence_kind,''), COALESCE(evidence_ref,''), COALESCE(evaluation_verdict,''), COALESCE(evaluation_strength,''), created_at
FROM projection_obligation_decisions WHERE obligation_id = ?
`, ob.ID)
	var d ProjectionObligationDecisionRow
	err := row.Scan(&d.ObligationID, &d.RunID, &d.Status, &d.DecidedBy, &d.Basis, &d.EvidenceKind, &d.EvidenceRef, &d.EvaluationVerdict, &d.EvaluationStrength, &d.CreatedAt)
	switch {
	case err == nil:
		ob.Decision = &d
	case errors.Is(err, sql.ErrNoRows):
		// open obligation
	default:
		return err
	}
	return nil
}
