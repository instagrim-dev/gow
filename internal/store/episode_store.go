package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/instagrim-dev/newf/internal/domain"
)

// Episode store: the prospective two-observation protocol (v41). The store
// enforces the ORDER of the loop transactionally — commitments before
// outcomes, revision before a second commitment, one observation per
// commitment forever — and the coherence of code-derived scores. Prose
// judgments never move a score: hit/miss is exactly (verdict ==
// predicted_verdict), checked again here at the persistence boundary.

// EpisodeRow is one preregistered episode.
type EpisodeRow struct {
	ID          string
	ProblemID   string
	Title       string
	MapRef      string
	ScoringRule string
	Status      string // open|completed|abandoned
	CreatedAt   string
	CompletedAt string
}

// EpisodeCommitmentRow is one frozen step commitment, made BEFORE its outcome.
type EpisodeCommitmentRow struct {
	ID               string
	EpisodeID        string
	Step             int
	Action           string
	Prediction       string
	PredictedVerdict string // witness-valid|witness-invalid
	MapRef           string
	Basis            string
	CreatedAt        string
}

// EpisodeObservationRow is the externally checked outcome for one commitment.
type EpisodeObservationRow struct {
	ID                   string
	CommitmentID         string
	Checker              string
	Payload              string
	Verdict              string // witness-valid|witness-invalid
	Score                string // hit|miss (code-derived)
	VerificationSubject  string // always domain-goal
	VerificationStrength string // always reproducible
	Detail               string
	CreatedAt            string
}

// EpisodeRevisionRow is the recorded map change between step 1 and step 2.
type EpisodeRevisionRow struct {
	ID           string
	EpisodeID    string
	AfterStep    int
	MapRefBefore string
	MapRefAfter  string
	WhatChanged  string
	Basis        string
	CreatedAt    string
}

// EpisodeRecord is one fully hydrated episode.
type EpisodeRecord struct {
	Episode      EpisodeRow
	Commitments  []EpisodeCommitmentRow
	Observations []EpisodeObservationRow
	Revisions    []EpisodeRevisionRow
}

// withEpisodeTx runs fn in one transaction, committing on nil error.
func (s *Store) withEpisodeTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

// PersistEpisode preregisters an episode WITH its step-1 commitment in one
// transaction: freezing map, action, prediction, and scoring rule is a single
// act — an episode without a committed first step is not preregistered.
func (s *Store) PersistEpisode(ctx context.Context, ep EpisodeRow, first EpisodeCommitmentRow) error {
	if err := domain.ValidateEpisodeID(ep.ID); err != nil {
		return err
	}
	if err := domain.ValidateEpisodeCommitmentID(first.ID); err != nil {
		return err
	}
	if first.Step != 1 {
		return fmt.Errorf("preregistration commits step 1; got step %d", first.Step)
	}
	return s.withEpisodeTx(ctx, func(tx *sql.Tx) error {
		var exists int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM problems WHERE id = ?`, ep.ProblemID).Scan(&exists); err != nil {
			return err
		}
		if exists == 0 {
			return fmt.Errorf("%w: problem %s", ErrNotFound, ep.ProblemID)
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO episodes(id, problem_id, title, map_ref, scoring_rule, status, created_at)
VALUES(?, ?, ?, ?, ?, 'open', ?)
`, ep.ID, ep.ProblemID, ep.Title, ep.MapRef, ep.ScoringRule, ep.CreatedAt); err != nil {
			return fmt.Errorf("persist episode: %w", err)
		}
		return insertEpisodeCommitment(ctx, tx, first)
	})
}

func insertEpisodeCommitment(ctx context.Context, tx *sql.Tx, c EpisodeCommitmentRow) error {
	if _, err := tx.ExecContext(ctx, `
INSERT INTO episode_commitments(id, episode_id, step, action, prediction, predicted_verdict, map_ref, basis, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)
`, c.ID, c.EpisodeID, c.Step, c.Action, c.Prediction, c.PredictedVerdict, c.MapRef, c.Basis, c.CreatedAt); err != nil {
		return fmt.Errorf("persist episode commitment: %w", err)
	}
	return nil
}

// PersistEpisodeStepTwo appends the step-2 commitment. The transaction
// refuses unless: the episode is open, step 1 exists AND is observed, a
// revision is recorded, and the new action differs from step 1's action. The
// order of the loop is a persistence invariant, not a convention.
func (s *Store) PersistEpisodeStepTwo(ctx context.Context, c EpisodeCommitmentRow) error {
	if err := domain.ValidateEpisodeCommitmentID(c.ID); err != nil {
		return err
	}
	if c.Step != 2 {
		return fmt.Errorf("PersistEpisodeStepTwo persists step 2; got step %d", c.Step)
	}
	return s.withEpisodeTx(ctx, func(tx *sql.Tx) error {
		var status string
		if err := tx.QueryRowContext(ctx, `SELECT status FROM episodes WHERE id = ?`, c.EpisodeID).Scan(&status); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("%w: episode %s", ErrNotFound, c.EpisodeID)
			}
			return err
		}
		if status != "open" {
			return fmt.Errorf("episode %s is %s; step 2 requires an open episode", c.EpisodeID, status)
		}
		var step1ID, step1Action string
		err := tx.QueryRowContext(ctx, `SELECT id, action FROM episode_commitments WHERE episode_id = ? AND step = 1`, c.EpisodeID).Scan(&step1ID, &step1Action)
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("episode %s has no step-1 commitment", c.EpisodeID)
		}
		if err != nil {
			return err
		}
		var observed int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM episode_observations WHERE commitment_id = ?`, step1ID).Scan(&observed); err != nil {
			return err
		}
		if observed == 0 {
			return fmt.Errorf("step 1 of episode %s is not yet observed: the first outcome must be obtained before the next action is committed", c.EpisodeID)
		}
		var revisions int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM episode_revisions WHERE episode_id = ?`, c.EpisodeID).Scan(&revisions); err != nil {
			return err
		}
		if revisions == 0 {
			return fmt.Errorf("episode %s has no recorded map revision: step 2 must follow an explicit revision, not a re-roll", c.EpisodeID)
		}
		if c.Action == step1Action {
			return fmt.Errorf("step-2 action is identical to step 1 (%q): the revision must produce a DIFFERENT next action", step1Action)
		}
		return insertEpisodeCommitment(ctx, tx, c)
	})
}

// PersistEpisodeObservation records the externally checked outcome for one
// commitment, append-once. The score must equal the code-derived comparison
// of verdict against the commitment's predicted verdict — re-checked here so
// a caller cannot persist a flattering score. When the observed commitment is
// step 2, the episode completes in the same transaction.
func (s *Store) PersistEpisodeObservation(ctx context.Context, o EpisodeObservationRow, completedAt string) error {
	if err := domain.ValidateEpisodeObservationID(o.ID); err != nil {
		return err
	}
	return s.withEpisodeTx(ctx, func(tx *sql.Tx) error {
		var episodeID, predicted string
		var step int
		err := tx.QueryRowContext(ctx, `SELECT episode_id, step, predicted_verdict FROM episode_commitments WHERE id = ?`, o.CommitmentID).Scan(&episodeID, &step, &predicted)
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: commitment %s", ErrNotFound, o.CommitmentID)
		}
		if err != nil {
			return err
		}
		var status string
		if err := tx.QueryRowContext(ctx, `SELECT status FROM episodes WHERE id = ?`, episodeID).Scan(&status); err != nil {
			return err
		}
		if status != "open" {
			return fmt.Errorf("episode %s is %s; observations require an open episode", episodeID, status)
		}
		wantScore := "miss"
		if o.Verdict == predicted {
			wantScore = "hit"
		}
		if o.Score != wantScore {
			return fmt.Errorf("score %q contradicts code-derived comparison (verdict %q vs predicted %q => %q); scores are never authored", o.Score, o.Verdict, predicted, wantScore)
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO episode_observations(id, commitment_id, checker, payload, verdict, score, verification_subject, verification_strength, detail, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, o.ID, o.CommitmentID, o.Checker, o.Payload, o.Verdict, o.Score, o.VerificationSubject, o.VerificationStrength, o.Detail, o.CreatedAt); err != nil {
			return fmt.Errorf("persist episode observation: %w", err)
		}
		if step == 2 {
			if _, err := tx.ExecContext(ctx, `UPDATE episodes SET status = 'completed', completed_at = ? WHERE id = ?`, completedAt, episodeID); err != nil {
				return fmt.Errorf("complete episode: %w", err)
			}
		}
		return nil
	})
}

// PersistEpisodeRevision records the map change after step 1, append-once.
// It refuses when step 1 is unobserved (a revision must respond to evidence)
// and when the map reference did not change.
func (s *Store) PersistEpisodeRevision(ctx context.Context, r EpisodeRevisionRow) error {
	if err := domain.ValidateEpisodeRevisionID(r.ID); err != nil {
		return err
	}
	return s.withEpisodeTx(ctx, func(tx *sql.Tx) error {
		var status string
		if err := tx.QueryRowContext(ctx, `SELECT status FROM episodes WHERE id = ?`, r.EpisodeID).Scan(&status); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("%w: episode %s", ErrNotFound, r.EpisodeID)
			}
			return err
		}
		if status != "open" {
			return fmt.Errorf("episode %s is %s; revisions require an open episode", r.EpisodeID, status)
		}
		var step1ID string
		err := tx.QueryRowContext(ctx, `SELECT id FROM episode_commitments WHERE episode_id = ? AND step = 1`, r.EpisodeID).Scan(&step1ID)
		if err != nil {
			return err
		}
		var observed int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM episode_observations WHERE commitment_id = ?`, step1ID).Scan(&observed); err != nil {
			return err
		}
		if observed == 0 {
			return fmt.Errorf("step 1 of episode %s is not yet observed: a revision responds to evidence, it does not precede it", r.EpisodeID)
		}
		if r.MapRefBefore == r.MapRefAfter {
			return fmt.Errorf("map_ref did not change (%q): a revision that changes nothing is not a revision", r.MapRefBefore)
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO episode_revisions(id, episode_id, after_step, map_ref_before, map_ref_after, what_changed, basis, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?)
`, r.ID, r.EpisodeID, r.AfterStep, r.MapRefBefore, r.MapRefAfter, r.WhatChanged, r.Basis, r.CreatedAt); err != nil {
			return fmt.Errorf("persist episode revision: %w", err)
		}
		return nil
	})
}

// GetEpisode hydrates one full episode record.
func (s *Store) GetEpisode(ctx context.Context, id string) (EpisodeRecord, error) {
	if err := domain.ValidateEpisodeID(id); err != nil {
		return EpisodeRecord{}, err
	}
	var rec EpisodeRecord
	row := s.db.QueryRowContext(ctx, `
SELECT id, problem_id, title, map_ref, scoring_rule, status, created_at, COALESCE(completed_at,'')
FROM episodes WHERE id = ?
`, id)
	e := &rec.Episode
	if err := row.Scan(&e.ID, &e.ProblemID, &e.Title, &e.MapRef, &e.ScoringRule, &e.Status, &e.CreatedAt, &e.CompletedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return EpisodeRecord{}, fmt.Errorf("%w: episode %s", ErrNotFound, id)
		}
		return EpisodeRecord{}, err
	}
	crows, err := s.db.QueryContext(ctx, `
SELECT id, episode_id, step, action, prediction, predicted_verdict, map_ref, basis, created_at
FROM episode_commitments WHERE episode_id = ? ORDER BY step
`, id)
	if err != nil {
		return EpisodeRecord{}, err
	}
	defer crows.Close()
	for crows.Next() {
		var c EpisodeCommitmentRow
		if err := crows.Scan(&c.ID, &c.EpisodeID, &c.Step, &c.Action, &c.Prediction, &c.PredictedVerdict, &c.MapRef, &c.Basis, &c.CreatedAt); err != nil {
			return EpisodeRecord{}, err
		}
		rec.Commitments = append(rec.Commitments, c)
	}
	if err := crows.Err(); err != nil {
		return EpisodeRecord{}, err
	}
	orows, err := s.db.QueryContext(ctx, `
SELECT o.id, o.commitment_id, o.checker, o.payload, o.verdict, o.score, o.verification_subject, o.verification_strength, o.detail, o.created_at
FROM episode_observations o JOIN episode_commitments c ON c.id = o.commitment_id
WHERE c.episode_id = ? ORDER BY c.step
`, id)
	if err != nil {
		return EpisodeRecord{}, err
	}
	defer orows.Close()
	for orows.Next() {
		var o EpisodeObservationRow
		if err := orows.Scan(&o.ID, &o.CommitmentID, &o.Checker, &o.Payload, &o.Verdict, &o.Score, &o.VerificationSubject, &o.VerificationStrength, &o.Detail, &o.CreatedAt); err != nil {
			return EpisodeRecord{}, err
		}
		rec.Observations = append(rec.Observations, o)
	}
	if err := orows.Err(); err != nil {
		return EpisodeRecord{}, err
	}
	rrows, err := s.db.QueryContext(ctx, `
SELECT id, episode_id, after_step, map_ref_before, map_ref_after, what_changed, basis, created_at
FROM episode_revisions WHERE episode_id = ? ORDER BY after_step
`, id)
	if err != nil {
		return EpisodeRecord{}, err
	}
	defer rrows.Close()
	for rrows.Next() {
		var r EpisodeRevisionRow
		if err := rrows.Scan(&r.ID, &r.EpisodeID, &r.AfterStep, &r.MapRefBefore, &r.MapRefAfter, &r.WhatChanged, &r.Basis, &r.CreatedAt); err != nil {
			return EpisodeRecord{}, err
		}
		rec.Revisions = append(rec.Revisions, r)
	}
	return rec, rrows.Err()
}

// ListEpisodesForProblem returns episode headers for a problem, oldest first.
func (s *Store) ListEpisodesForProblem(ctx context.Context, problemID string) ([]EpisodeRow, error) {
	if err := domain.ValidateProblemID(problemID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT id, problem_id, title, map_ref, scoring_rule, status, created_at, COALESCE(completed_at,'')
FROM episodes WHERE problem_id = ? ORDER BY created_at, id
`, problemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []EpisodeRow
	for rows.Next() {
		var e EpisodeRow
		if err := rows.Scan(&e.ID, &e.ProblemID, &e.Title, &e.MapRef, &e.ScoringRule, &e.Status, &e.CreatedAt, &e.CompletedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}
