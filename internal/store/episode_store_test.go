package store

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/domain"
)

// seedEpisodeProblem creates a migrated store with one problem.
func seedEpisodeProblem(t *testing.T) (*Store, string) {
	t.Helper()
	st := openMigratedStore(t)
	ctx := context.Background()
	problemID := "prb_01EPISODE000000000000000000"
	// One transaction: created_by_run_id and runs.problem_id are mutually
	// referential; the deferred FK resolves at commit.
	tx, err := st.db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin seed tx: %v", err)
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO problems(id, slug, statement, status, created_at, created_by_run_id)
VALUES(?, 'erdos-straus', '4/n as three unit fractions', 'open', '2026-09-12T17:00:00Z', 'run_01EPISODE00000000000000000R')
`, problemID); err != nil {
		t.Fatalf("seed problem: %v", err)
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO runs(id, problem_id, operation, status, input_ref, tool_name, tool_version, started_at, completed_at)
VALUES('run_01EPISODE00000000000000000R', ?, 'test', 'completed', '', 't', 'v', '2026-09-12T17:00:00Z', '2026-09-12T17:00:00Z')
`, problemID); err != nil {
		t.Fatalf("seed run: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit seed: %v", err)
	}
	return st, problemID
}

func sampleEpisode(problemID string, ts time.Time) (EpisodeRow, EpisodeCommitmentRow) {
	ep := EpisodeRow{
		ID: domain.NewEpisodeID(ts), ProblemID: problemID,
		Title:  "prospective two-observation loop",
		MapRef: "invr_MAP_A", ScoringRule: "hit iff observed witness verdict equals predicted verdict (code-derived)",
		CreatedAt: ts.Format("2006-01-02T15:04:05Z"),
	}
	c1 := EpisodeCommitmentRow{
		ID: domain.NewEpisodeCommitmentID(ts), EpisodeID: ep.ID, Step: 1,
		Action: "attempt witness construction for n=5 via identity family F1", Prediction: "F1 yields a valid witness for n=5",
		PredictedVerdict: "witness-valid", MapRef: "invr_MAP_A", Basis: "map A ranks F1 first",
		CreatedAt: ts.Format("2006-01-02T15:04:05Z"),
	}
	return ep, c1
}

// The loop's order is a persistence invariant: no step 2 before step 1 is
// observed and a revision exists; no revision before evidence; no second
// observation of a commitment; scores are code-derived, never authored.
func TestEpisodeLoopOrderEnforced(t *testing.T) {
	st, problemID := seedEpisodeProblem(t)
	ctx := context.Background()
	ts := time.Date(2026, 9, 12, 17, 30, 0, 0, time.UTC)
	ep, c1 := sampleEpisode(problemID, ts)
	if err := st.PersistEpisode(ctx, ep, c1); err != nil {
		t.Fatalf("preregister: %v", err)
	}

	c2 := EpisodeCommitmentRow{
		ID: domain.NewEpisodeCommitmentID(ts.Add(2 * time.Hour)), EpisodeID: ep.ID, Step: 2,
		Action: "attempt witness construction for n=5 via identity family F2", Prediction: "F2 yields a valid witness",
		PredictedVerdict: "witness-valid", MapRef: "invr_MAP_B", Basis: "revised map ranks F2 first",
		CreatedAt: ts.Add(2 * time.Hour).Format("2006-01-02T15:04:05Z"),
	}
	if err := st.PersistEpisodeStepTwo(ctx, c2); err == nil || !strings.Contains(err.Error(), "not yet observed") {
		t.Fatalf("step 2 before step-1 observation must be refused: %v", err)
	}
	rev := EpisodeRevisionRow{
		ID: domain.NewEpisodeRevisionID(ts.Add(time.Hour)), EpisodeID: ep.ID, AfterStep: 1,
		MapRefBefore: "invr_MAP_A", MapRefAfter: "invr_MAP_B",
		WhatChanged: "penalize family F1 after miss", Basis: "step-1 outcome",
		CreatedAt: ts.Add(time.Hour).Format("2006-01-02T15:04:05Z"),
	}
	if err := st.PersistEpisodeRevision(ctx, rev); err == nil || !strings.Contains(err.Error(), "not yet observed") {
		t.Fatalf("revision before evidence must be refused: %v", err)
	}

	// Observe step 1: a MISS (predicted valid, observed invalid). The store
	// re-derives the score; a flattering 'hit' is refused.
	o1 := EpisodeObservationRow{
		ID: domain.NewEpisodeObservationID(ts.Add(30 * time.Minute)), CommitmentID: c1.ID,
		Checker: "erdos-straus-witness/v1", Payload: `{"n":"5","x":"2","y":"4","z":"21"}`,
		Verdict: "witness-invalid", Score: "hit",
		VerificationSubject: "domain-goal", VerificationStrength: "reproducible",
		Detail: "identity fails", CreatedAt: ts.Add(30 * time.Minute).Format("2006-01-02T15:04:05Z"),
	}
	if err := st.PersistEpisodeObservation(ctx, o1, ""); err == nil || !strings.Contains(err.Error(), "never authored") {
		t.Fatalf("authored score contradicting the code derivation must be refused: %v", err)
	}
	o1.Score = "miss"
	if err := st.PersistEpisodeObservation(ctx, o1, ""); err != nil {
		t.Fatalf("observe step 1: %v", err)
	}
	dup := o1
	dup.ID = domain.NewEpisodeObservationID(ts.Add(31 * time.Minute))
	if err := st.PersistEpisodeObservation(ctx, dup, ""); err == nil {
		t.Fatal("a commitment must be observed at most once")
	}

	// Step 2 still refused: no revision yet.
	if err := st.PersistEpisodeStepTwo(ctx, c2); err == nil || !strings.Contains(err.Error(), "no recorded map revision") {
		t.Fatalf("step 2 without a revision must be refused: %v", err)
	}
	badRev := rev
	badRev.MapRefAfter = badRev.MapRefBefore
	if err := st.PersistEpisodeRevision(ctx, badRev); err == nil || !strings.Contains(err.Error(), "did not change") {
		t.Fatalf("no-op revision must be refused: %v", err)
	}
	if err := st.PersistEpisodeRevision(ctx, rev); err != nil {
		t.Fatalf("revision after evidence: %v", err)
	}

	// Identical action re-roll refused; different action admitted.
	sameAction := c2
	sameAction.Action = c1.Action
	if err := st.PersistEpisodeStepTwo(ctx, sameAction); err == nil || !strings.Contains(err.Error(), "DIFFERENT next action") {
		t.Fatalf("identical step-2 action must be refused: %v", err)
	}
	if err := st.PersistEpisodeStepTwo(ctx, c2); err != nil {
		t.Fatalf("step 2 commit: %v", err)
	}

	// Observe step 2: a HIT completes the episode in the same transaction.
	o2 := EpisodeObservationRow{
		ID: domain.NewEpisodeObservationID(ts.Add(3 * time.Hour)), CommitmentID: c2.ID,
		Checker: "erdos-straus-witness/v1", Payload: `{"n":"5","x":"2","y":"4","z":"20"}`,
		Verdict: "witness-valid", Score: "hit",
		VerificationSubject: "domain-goal", VerificationStrength: "reproducible",
		Detail: "identity holds", CreatedAt: ts.Add(3 * time.Hour).Format("2006-01-02T15:04:05Z"),
	}
	completedAt := ts.Add(3 * time.Hour).Format("2006-01-02T15:04:05Z")
	if err := st.PersistEpisodeObservation(ctx, o2, completedAt); err != nil {
		t.Fatalf("observe step 2: %v", err)
	}

	rec, err := st.GetEpisode(ctx, ep.ID)
	if err != nil {
		t.Fatalf("get episode: %v", err)
	}
	if rec.Episode.Status != "completed" || rec.Episode.CompletedAt != completedAt {
		t.Fatalf("episode must complete on step-2 observation: %+v", rec.Episode)
	}
	if len(rec.Commitments) != 2 || len(rec.Observations) != 2 || len(rec.Revisions) != 1 {
		t.Fatalf("full record: %d commitments, %d observations, %d revisions", len(rec.Commitments), len(rec.Observations), len(rec.Revisions))
	}
	if rec.Observations[0].Score != "miss" || rec.Observations[1].Score != "hit" {
		t.Fatalf("scores must round-trip: %+v", rec.Observations)
	}

	// A completed episode accepts nothing further.
	late := EpisodeRevisionRow{
		ID: domain.NewEpisodeRevisionID(ts.Add(4 * time.Hour)), EpisodeID: ep.ID, AfterStep: 1,
		MapRefBefore: "invr_MAP_B", MapRefAfter: "invr_MAP_C",
		WhatChanged: "post-hoc", Basis: "post-hoc", CreatedAt: completedAt,
	}
	if err := st.PersistEpisodeRevision(ctx, late); err == nil {
		t.Fatal("a completed episode must refuse further revisions")
	}
}

// Preregistration fields are frozen and all episode records are immutable at
// the SQL layer.
func TestEpisodeImmutability(t *testing.T) {
	st, problemID := seedEpisodeProblem(t)
	ctx := context.Background()
	ep, c1 := sampleEpisode(problemID, time.Date(2026, 9, 12, 17, 0, 0, 0, time.UTC))
	if err := st.PersistEpisode(ctx, ep, c1); err != nil {
		t.Fatalf("preregister: %v", err)
	}
	if _, err := st.db.ExecContext(ctx, `UPDATE episodes SET scoring_rule = 'looser' WHERE id = ?`, ep.ID); err == nil {
		t.Fatal("scoring rule must be frozen at preregistration")
	}
	if _, err := st.db.ExecContext(ctx, `UPDATE episodes SET map_ref = 'invr_OTHER' WHERE id = ?`, ep.ID); err == nil {
		t.Fatal("frozen map reference must be immutable")
	}
	if _, err := st.db.ExecContext(ctx, `UPDATE episode_commitments SET predicted_verdict = 'witness-invalid' WHERE id = ?`, c1.ID); err == nil {
		t.Fatal("a commitment's prediction must be immutable (no post-hoc prediction edits)")
	}
	if _, err := st.db.ExecContext(ctx, `DELETE FROM episode_commitments WHERE id = ?`, c1.ID); err == nil {
		t.Fatal("commitments must not be deletable")
	}
}
