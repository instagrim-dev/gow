package pipeline

import (
	"context"
	"strings"
	"testing"
	"time"
)

// The reviewer's minimum persuasive demonstration (2026-09-12 review,
// epistemic finding), end-to-end at the App boundary: freeze map, action,
// prediction, and scoring rule; obtain an externally checked outcome; revise
// the map; commit to a DIFFERENT next action; obtain and score the next
// outcome. The first miss remains a miss; the revision earns credit only on
// later evidence — and that credit is code-derived from exact-integer witness
// checks, never authored.
func TestIntegrationEpisodeTwoObservationLoop(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2026, 9, 12, 18, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	problemID, _, _ := seedProblemAndSnapshot(t, ctx, dbPath, now)

	// 1. Preregister: map, first action, prediction, scoring rule — one act.
	pre, err := app.PreregisterEpisode(ctx, PreregisterEpisodeInput{
		DBPath: dbPath, ProblemID: problemID,
		Title:  "prospective loop: witness construction for n=5",
		MapRef: "invr_MAP_A",
		Action: "construct witness for n=5 via family F1 (greedy z)", Prediction: "F1 yields a valid witness for n=5",
		PredictedVerdict: "witness-valid", Note: "map A ranks F1 first",
	})
	if err != nil {
		t.Fatalf("preregister: %v", err)
	}
	ep := pre.Episode
	if ep.Status != "open" || len(ep.Commitments) != 1 || ep.Commitments[0].Step != 1 {
		t.Fatalf("preregistration must freeze exactly the step-1 commitment: %+v", ep)
	}
	if ep.RevisionCredit != "unresolved" {
		t.Fatalf("credit before any outcome must be unresolved, got %q", ep.RevisionCredit)
	}

	// 2. A revision before evidence is refused: revision responds to outcomes.
	if _, err := app.ReviseEpisode(ctx, ReviseEpisodeInput{
		DBPath: dbPath, EpisodeID: ep.ID, MapRefAfter: "invr_MAP_B", WhatChanged: "premature", Note: "premature",
	}); err == nil {
		t.Fatal("revision before the step-1 observation must be refused")
	}

	// 3. Observe step 1: the action produced the near-miss tuple (2,4,21) —
	// the exact checker refutes it. Predicted valid, observed invalid: MISS.
	obs1, err := app.ObserveEpisode(ctx, ObserveEpisodeInput{
		DBPath: dbPath, EpisodeID: ep.ID, WitnessTuple: "5,2,4,21", Note: "F1 output",
	})
	if err != nil {
		t.Fatalf("observe step 1: %v", err)
	}
	if got := obs1.Episode.StepScores[1]; got != "miss" {
		t.Fatalf("step-1 score = %q, want miss (the first miss remains a miss)", got)
	}
	o1 := obs1.Episode.Observations[0]
	if o1.Verdict != "witness-invalid" || o1.VerificationSubject != "domain-goal" || o1.VerificationStrength != "reproducible" {
		t.Fatalf("step-1 outcome must be an externally checked domain observation: %+v", o1)
	}
	if !strings.Contains(o1.Detail, "identity fails") {
		t.Fatalf("the refutation must name the exact failure: %q", o1.Detail)
	}

	// 4. Step 2 before a recorded revision is refused: no re-rolls.
	if _, err := app.CommitEpisodeStepTwo(ctx, CommitEpisodeInput{
		DBPath: dbPath, EpisodeID: ep.ID, Action: "construct witness for n=5 via family F2",
		Prediction: "F2 yields a valid witness", PredictedVerdict: "witness-valid", Note: "retry",
	}); err == nil || !strings.Contains(err.Error(), "no recorded map revision") {
		t.Fatalf("step 2 without a revision must be refused: %v", err)
	}

	// 5. Revise the map in response to the miss.
	if _, err := app.ReviseEpisode(ctx, ReviseEpisodeInput{
		DBPath: dbPath, EpisodeID: ep.ID, MapRefAfter: "invr_MAP_B",
		WhatChanged: "penalize family F1 for n≡1 (mod 4); promote F2", Note: "step-1 refutation",
	}); err != nil {
		t.Fatalf("revise: %v", err)
	}

	// 6. An identical action is refused; a different action commits.
	if _, err := app.CommitEpisodeStepTwo(ctx, CommitEpisodeInput{
		DBPath: dbPath, EpisodeID: ep.ID, Action: "construct witness for n=5 via family F1 (greedy z)",
		Prediction: "same again", PredictedVerdict: "witness-valid", Note: "same",
	}); err == nil || !strings.Contains(err.Error(), "DIFFERENT next action") {
		t.Fatalf("identical step-2 action must be refused: %v", err)
	}
	c2resp, err := app.CommitEpisodeStepTwo(ctx, CommitEpisodeInput{
		DBPath: dbPath, EpisodeID: ep.ID, Action: "construct witness for n=5 via family F2 (fixed x=2, y=4)",
		Prediction: "F2 yields a valid witness for n=5", PredictedVerdict: "witness-valid", Note: "map B ranks F2 first",
	})
	if err != nil {
		t.Fatalf("commit step 2: %v", err)
	}
	if got := c2resp.Episode.Commitments[1].MapRef; got != "invr_MAP_B" {
		t.Fatalf("step 2 must be committed under the REVISED map, got %q", got)
	}

	// 7. Observe step 2: the revised action produced the valid witness
	// 4/5 = 1/2 + 1/4 + 1/20. HIT; the episode completes; credit is earned —
	// on later evidence only.
	obs2, err := app.ObserveEpisode(ctx, ObserveEpisodeInput{
		DBPath: dbPath, EpisodeID: ep.ID, WitnessTuple: "5,2,4,20", Note: "F2 output",
	})
	if err != nil {
		t.Fatalf("observe step 2: %v", err)
	}
	final := obs2.Episode
	if final.Status != "completed" {
		t.Fatalf("episode must complete on the second observation, got %q", final.Status)
	}
	if final.StepScores[1] != "miss" || final.StepScores[2] != "hit" {
		t.Fatalf("scores = %+v; the first miss must remain a miss", final.StepScores)
	}
	if final.RevisionCredit != "earned" {
		t.Fatalf("revision credit = %q, want earned (step-2 hit after a recorded revision)", final.RevisionCredit)
	}

	// 8. A completed episode accepts no further observations.
	if _, err := app.ObserveEpisode(ctx, ObserveEpisodeInput{
		DBPath: dbPath, EpisodeID: ep.ID, WitnessTuple: "5,2,4,20",
	}); err == nil {
		t.Fatal("a completed episode must refuse further observations")
	}
}

// A revision that does not pay off is recorded as exactly that: the second
// miss completes the episode with revision_credit = not-earned. Honest
// negatives are a legitimate stopping state, not a failure of the protocol.
func TestIntegrationEpisodeRevisionNotEarned(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2026, 9, 12, 19, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	problemID, _, _ := seedProblemAndSnapshot(t, ctx, dbPath, now)

	pre, err := app.PreregisterEpisode(ctx, PreregisterEpisodeInput{
		DBPath: dbPath, ProblemID: problemID, Title: "loop where the revision fails",
		MapRef: "invr_MAP_A", Action: "family F1 for n=7", Prediction: "valid witness",
		PredictedVerdict: "witness-valid", Note: "map A",
	})
	if err != nil {
		t.Fatalf("preregister: %v", err)
	}
	epID := pre.Episode.ID
	if _, err := app.ObserveEpisode(ctx, ObserveEpisodeInput{DBPath: dbPath, EpisodeID: epID, WitnessTuple: "7,2,7,14"}); err != nil {
		t.Fatalf("observe step 1: %v", err) // 1/2+1/7+1/14 = 5/7 != 4/7: miss
	}
	if _, err := app.ReviseEpisode(ctx, ReviseEpisodeInput{
		DBPath: dbPath, EpisodeID: epID, MapRefAfter: "invr_MAP_B", WhatChanged: "swap family", Note: "miss",
	}); err != nil {
		t.Fatalf("revise: %v", err)
	}
	if _, err := app.CommitEpisodeStepTwo(ctx, CommitEpisodeInput{
		DBPath: dbPath, EpisodeID: epID, Action: "family F3 for n=7", Prediction: "valid witness",
		PredictedVerdict: "witness-valid", Note: "map B",
	}); err != nil {
		t.Fatalf("commit: %v", err)
	}
	// The revised action ALSO produced an invalid tuple: second miss.
	obs, err := app.ObserveEpisode(ctx, ObserveEpisodeInput{DBPath: dbPath, EpisodeID: epID, WitnessTuple: "7,3,4,5"})
	if err != nil {
		t.Fatalf("observe step 2: %v", err)
	}
	if obs.Episode.Status != "completed" || obs.Episode.RevisionCredit != "not-earned" {
		t.Fatalf("a failed revision must complete honestly with credit not-earned: status=%q credit=%q",
			obs.Episode.Status, obs.Episode.RevisionCredit)
	}
}
