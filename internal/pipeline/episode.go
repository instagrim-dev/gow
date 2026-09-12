// Package pipeline: the prospective two-observation episode protocol (v41;
// 2026-09-12 review, epistemic finding). The reviewer's minimum persuasive
// demonstration is a closed loop — freeze map, action, prediction, and
// scoring rule; obtain an externally checked outcome; revise the map; commit
// to a DIFFERENT next action; obtain and score the next outcome. The first
// miss remains a miss; the revision earns credit only on later evidence.
//
// The pipeline owns the loop's order and the code-derived scoring; the
// witness package owns the external outcome check; models own nothing here.
package pipeline

import (
	"context"
	"fmt"
	"strings"

	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/store"
	"github.com/instagrim-dev/newf/internal/witness"
)

// --- inputs ---

// PreregisterEpisodeInput drives `newf episode preregister`: the single act
// that freezes map, first action, prediction, and scoring rule.
type PreregisterEpisodeInput struct {
	DBPath           string
	ProblemID        string
	Title            string
	MapRef           string
	Action           string
	Prediction       string
	PredictedVerdict string // witness-valid|witness-invalid
	Note             string
	JSONOutput       bool
}

// ObserveEpisodeInput drives `newf episode observe`: run the external checker
// on the produced witness tuple and score the open commitment.
type ObserveEpisodeInput struct {
	DBPath    string
	EpisodeID string
	// WitnessTuple is "n,x,y,z" — the concrete outcome of the committed
	// action, checked exactly by the Erdős–Straus witness checker.
	WitnessTuple string
	Note         string
	JSONOutput   bool
}

// ReviseEpisodeInput drives `newf episode revise`.
type ReviseEpisodeInput struct {
	DBPath      string
	EpisodeID   string
	MapRefAfter string
	WhatChanged string
	Note        string
	JSONOutput  bool
}

// CommitEpisodeInput drives `newf episode commit` (step 2 only; step 1 is
// frozen at preregistration).
type CommitEpisodeInput struct {
	DBPath           string
	EpisodeID        string
	Action           string
	Prediction       string
	PredictedVerdict string
	Note             string
	JSONOutput       bool
}

// EpisodeShowInput drives `newf episode show` / `list`.
type EpisodeShowInput struct {
	DBPath     string
	EpisodeID  string
	ProblemID  string
	JSONOutput bool
}

// --- views ---

// EpisodeCommitmentView is one frozen step commitment.
type EpisodeCommitmentView struct {
	ID               string `json:"id"`
	Step             int    `json:"step"`
	Action           string `json:"action"`
	Prediction       string `json:"prediction"`
	PredictedVerdict string `json:"predicted_verdict"`
	MapRef           string `json:"map_ref"`
	Basis            string `json:"basis"`
	CreatedAt        string `json:"created_at"`
}

// EpisodeObservationView is one externally checked, code-scored outcome.
type EpisodeObservationView struct {
	ID                   string `json:"id"`
	CommitmentID         string `json:"commitment_id"`
	Checker              string `json:"checker"`
	Payload              string `json:"payload"`
	Verdict              string `json:"verdict"`
	Score                string `json:"score"`
	VerificationSubject  string `json:"verification_subject"`
	VerificationStrength string `json:"verification_strength"`
	Detail               string `json:"detail,omitempty"`
	CreatedAt            string `json:"created_at"`
}

// EpisodeRevisionView is the recorded map change between steps.
type EpisodeRevisionView struct {
	ID           string `json:"id"`
	AfterStep    int    `json:"after_step"`
	MapRefBefore string `json:"map_ref_before"`
	MapRefAfter  string `json:"map_ref_after"`
	WhatChanged  string `json:"what_changed"`
	Basis        string `json:"basis"`
	CreatedAt    string `json:"created_at"`
}

// EpisodeView is one full episode with its derived, code-owned summary.
type EpisodeView struct {
	ID           string                   `json:"id"`
	ProblemID    string                   `json:"problem_id"`
	Title        string                   `json:"title"`
	MapRef       string                   `json:"map_ref"`
	ScoringRule  string                   `json:"scoring_rule"`
	Status       string                   `json:"status"`
	CreatedAt    string                   `json:"created_at"`
	CompletedAt  string                   `json:"completed_at,omitempty"`
	Commitments  []EpisodeCommitmentView  `json:"commitments"`
	Observations []EpisodeObservationView `json:"observations"`
	Revisions    []EpisodeRevisionView    `json:"revisions,omitempty"`
	// StepScores maps step -> hit|miss for observed steps. Derived, never stored.
	StepScores map[int]string `json:"step_scores,omitempty"`
	// RevisionCredit is the reviewer's credit rule, code-derived: `earned`
	// iff step 2 was observed a hit AFTER a recorded revision; `not-earned`
	// iff step 2 was observed a miss; `unresolved` otherwise. The first miss
	// remains a miss regardless.
	RevisionCredit string `json:"revision_credit,omitempty"`
}

// EpisodeResponse is returned by the mutating episode commands.
type EpisodeResponse struct {
	OK      bool        `json:"ok"`
	Command string      `json:"command"`
	Store   string      `json:"store"`
	Episode EpisodeView `json:"episode"`
}

// EpisodesResponse is returned by `newf episode list`.
type EpisodesResponse struct {
	OK       bool          `json:"ok"`
	Command  string        `json:"command"`
	Store    string        `json:"store"`
	Episodes []EpisodeView `json:"episodes"`
}

func episodeView(rec store.EpisodeRecord) EpisodeView {
	v := EpisodeView{
		ID: rec.Episode.ID, ProblemID: rec.Episode.ProblemID, Title: rec.Episode.Title,
		MapRef: rec.Episode.MapRef, ScoringRule: rec.Episode.ScoringRule,
		Status: rec.Episode.Status, CreatedAt: rec.Episode.CreatedAt, CompletedAt: rec.Episode.CompletedAt,
		Commitments:  []EpisodeCommitmentView{},
		Observations: []EpisodeObservationView{},
	}
	stepByCommitment := map[string]int{}
	for _, c := range rec.Commitments {
		stepByCommitment[c.ID] = c.Step
		v.Commitments = append(v.Commitments, EpisodeCommitmentView{
			ID: c.ID, Step: c.Step, Action: c.Action, Prediction: c.Prediction,
			PredictedVerdict: c.PredictedVerdict, MapRef: c.MapRef, Basis: c.Basis, CreatedAt: c.CreatedAt,
		})
	}
	scores := map[int]string{}
	for _, o := range rec.Observations {
		v.Observations = append(v.Observations, EpisodeObservationView{
			ID: o.ID, CommitmentID: o.CommitmentID, Checker: o.Checker, Payload: o.Payload,
			Verdict: o.Verdict, Score: o.Score, VerificationSubject: o.VerificationSubject,
			VerificationStrength: o.VerificationStrength, Detail: o.Detail, CreatedAt: o.CreatedAt,
		})
		if step, ok := stepByCommitment[o.CommitmentID]; ok {
			scores[step] = o.Score
		}
	}
	for _, r := range rec.Revisions {
		v.Revisions = append(v.Revisions, EpisodeRevisionView{
			ID: r.ID, AfterStep: r.AfterStep, MapRefBefore: r.MapRefBefore, MapRefAfter: r.MapRefAfter,
			WhatChanged: r.WhatChanged, Basis: r.Basis, CreatedAt: r.CreatedAt,
		})
	}
	if len(scores) > 0 {
		v.StepScores = scores
	}
	// Credit rule: derived here, never stored (a derived view must not fossilize).
	switch {
	case len(rec.Revisions) == 0 || scores[2] == "":
		v.RevisionCredit = "unresolved"
	case scores[2] == "hit":
		v.RevisionCredit = "earned"
	default:
		v.RevisionCredit = "not-earned"
	}
	return v
}

func validWitnessVerdict(v string) bool { return v == "witness-valid" || v == "witness-invalid" }

// --- operations ---

// PreregisterEpisode freezes the episode: map reference, scoring rule, and
// the step-1 commitment (action + falsifiable prediction) in one transaction.
// The scoring rule is fixed by the protocol and recorded verbatim so a later
// reader need not trust this code's history: hit iff the checker verdict
// equals the predicted verdict.
func (a *App) PreregisterEpisode(ctx context.Context, input PreregisterEpisodeInput) (EpisodeResponse, error) {
	for _, req := range []struct{ name, val string }{
		{"--title", input.Title}, {"--map", input.MapRef}, {"--action", input.Action},
		{"--prediction", input.Prediction}, {"--note", input.Note},
	} {
		if strings.TrimSpace(req.val) == "" {
			return EpisodeResponse{}, fmt.Errorf("%s is required: preregistration freezes every element of the loop", req.name)
		}
	}
	if !validWitnessVerdict(input.PredictedVerdict) {
		return EpisodeResponse{}, fmt.Errorf("--predict must be witness-valid or witness-invalid; got %q", input.PredictedVerdict)
	}
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return EpisodeResponse{}, err
	}
	defer repoStore.Close()

	now := a.now()
	ep := store.EpisodeRow{
		ID: domain.NewEpisodeID(now), ProblemID: input.ProblemID,
		Title: strings.TrimSpace(input.Title), MapRef: strings.TrimSpace(input.MapRef),
		ScoringRule: "hit iff " + witness.CheckerName + "/" + witness.CheckerVersion + " verdict equals the commitment's predicted verdict; scores are code-derived, never authored",
		CreatedAt:   now.Format(timeLayout),
	}
	c1 := store.EpisodeCommitmentRow{
		ID: domain.NewEpisodeCommitmentID(now), EpisodeID: ep.ID, Step: 1,
		Action: strings.TrimSpace(input.Action), Prediction: strings.TrimSpace(input.Prediction),
		PredictedVerdict: input.PredictedVerdict, MapRef: ep.MapRef,
		Basis: strings.TrimSpace(input.Note), CreatedAt: now.Format(timeLayout),
	}
	if err := repoStore.PersistEpisode(ctx, ep, c1); err != nil {
		return EpisodeResponse{}, err
	}
	rec, err := repoStore.GetEpisode(ctx, ep.ID)
	if err != nil {
		return EpisodeResponse{}, err
	}
	return EpisodeResponse{OK: true, Command: "episode preregister", Store: dbPath, Episode: episodeView(rec)}, nil
}

// ObserveEpisode runs the external checker against the produced witness tuple
// for the episode's open commitment and persists the code-scored outcome. The
// checker's verdict is the outcome; the operator supplies only the tuple the
// attempted action actually produced.
func (a *App) ObserveEpisode(ctx context.Context, input ObserveEpisodeInput) (EpisodeResponse, error) {
	claim, err := witness.ParseErdosStraus(input.WitnessTuple)
	if err != nil {
		return EpisodeResponse{}, err
	}
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return EpisodeResponse{}, err
	}
	defer repoStore.Close()

	rec, err := repoStore.GetEpisode(ctx, input.EpisodeID)
	if err != nil {
		return EpisodeResponse{}, err
	}
	observed := map[string]bool{}
	for _, o := range rec.Observations {
		observed[o.CommitmentID] = true
	}
	var open *store.EpisodeCommitmentRow
	for i := range rec.Commitments {
		if !observed[rec.Commitments[i].ID] {
			open = &rec.Commitments[i]
			break
		}
	}
	if open == nil {
		return EpisodeResponse{}, fmt.Errorf("episode %s has no open commitment to observe", input.EpisodeID)
	}

	verdict := "witness-valid"
	detail := "identity holds exactly over big integers"
	if cerr := claim.Check(); cerr != nil {
		verdict = "witness-invalid"
		detail = cerr.Error()
	}
	if strings.TrimSpace(input.Note) != "" {
		detail = detail + "; operator note: " + strings.TrimSpace(input.Note)
	}
	score := "miss"
	if verdict == open.PredictedVerdict {
		score = "hit"
	}
	now := a.now()
	o := store.EpisodeObservationRow{
		ID: domain.NewEpisodeObservationID(now), CommitmentID: open.ID,
		Checker: witness.CheckerName + "/" + witness.CheckerVersion, Payload: claim.Canonical(),
		Verdict: verdict, Score: score,
		VerificationSubject: "domain-goal", VerificationStrength: "reproducible",
		Detail: detail, CreatedAt: now.Format(timeLayout),
	}
	if err := repoStore.PersistEpisodeObservation(ctx, o, now.Format(timeLayout)); err != nil {
		return EpisodeResponse{}, err
	}
	rec, err = repoStore.GetEpisode(ctx, input.EpisodeID)
	if err != nil {
		return EpisodeResponse{}, err
	}
	return EpisodeResponse{OK: true, Command: "episode observe", Store: dbPath, Episode: episodeView(rec)}, nil
}

// ReviseEpisode records the map change after the step-1 outcome. The store
// refuses a revision before evidence or one whose map reference is unchanged.
func (a *App) ReviseEpisode(ctx context.Context, input ReviseEpisodeInput) (EpisodeResponse, error) {
	if strings.TrimSpace(input.MapRefAfter) == "" || strings.TrimSpace(input.WhatChanged) == "" || strings.TrimSpace(input.Note) == "" {
		return EpisodeResponse{}, fmt.Errorf("--map, --changed, and --note are required: a revision records what changed and why")
	}
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return EpisodeResponse{}, err
	}
	defer repoStore.Close()

	rec, err := repoStore.GetEpisode(ctx, input.EpisodeID)
	if err != nil {
		return EpisodeResponse{}, err
	}
	now := a.now()
	rev := store.EpisodeRevisionRow{
		ID: domain.NewEpisodeRevisionID(now), EpisodeID: input.EpisodeID, AfterStep: 1,
		MapRefBefore: rec.Episode.MapRef, MapRefAfter: strings.TrimSpace(input.MapRefAfter),
		WhatChanged: strings.TrimSpace(input.WhatChanged), Basis: strings.TrimSpace(input.Note),
		CreatedAt: now.Format(timeLayout),
	}
	if err := repoStore.PersistEpisodeRevision(ctx, rev); err != nil {
		return EpisodeResponse{}, err
	}
	rec, err = repoStore.GetEpisode(ctx, input.EpisodeID)
	if err != nil {
		return EpisodeResponse{}, err
	}
	return EpisodeResponse{OK: true, Command: "episode revise", Store: dbPath, Episode: episodeView(rec)}, nil
}

// CommitEpisodeStepTwo commits the revised next action. The store refuses it
// before the step-1 observation, before a recorded revision, or when the
// action is identical to step 1.
func (a *App) CommitEpisodeStepTwo(ctx context.Context, input CommitEpisodeInput) (EpisodeResponse, error) {
	if strings.TrimSpace(input.Action) == "" || strings.TrimSpace(input.Prediction) == "" || strings.TrimSpace(input.Note) == "" {
		return EpisodeResponse{}, fmt.Errorf("--action, --prediction, and --note are required: step 2 is a full commitment, not a retry")
	}
	if !validWitnessVerdict(input.PredictedVerdict) {
		return EpisodeResponse{}, fmt.Errorf("--predict must be witness-valid or witness-invalid; got %q", input.PredictedVerdict)
	}
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return EpisodeResponse{}, err
	}
	defer repoStore.Close()

	rec, err := repoStore.GetEpisode(ctx, input.EpisodeID)
	if err != nil {
		return EpisodeResponse{}, err
	}
	mapRef := rec.Episode.MapRef
	if n := len(rec.Revisions); n > 0 {
		mapRef = rec.Revisions[n-1].MapRefAfter
	}
	now := a.now()
	c2 := store.EpisodeCommitmentRow{
		ID: domain.NewEpisodeCommitmentID(now), EpisodeID: input.EpisodeID, Step: 2,
		Action: strings.TrimSpace(input.Action), Prediction: strings.TrimSpace(input.Prediction),
		PredictedVerdict: input.PredictedVerdict, MapRef: mapRef,
		Basis: strings.TrimSpace(input.Note), CreatedAt: now.Format(timeLayout),
	}
	if err := repoStore.PersistEpisodeStepTwo(ctx, c2); err != nil {
		return EpisodeResponse{}, err
	}
	rec, err = repoStore.GetEpisode(ctx, input.EpisodeID)
	if err != nil {
		return EpisodeResponse{}, err
	}
	return EpisodeResponse{OK: true, Command: "episode commit", Store: dbPath, Episode: episodeView(rec)}, nil
}

// ShowEpisode returns one full episode.
func (a *App) ShowEpisode(ctx context.Context, input EpisodeShowInput) (EpisodeResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return EpisodeResponse{}, err
	}
	defer repoStore.Close()
	rec, err := repoStore.GetEpisode(ctx, input.EpisodeID)
	if err != nil {
		return EpisodeResponse{}, err
	}
	return EpisodeResponse{OK: true, Command: "episode show", Store: dbPath, Episode: episodeView(rec)}, nil
}

// ListEpisodes returns all episodes for a problem.
func (a *App) ListEpisodes(ctx context.Context, input EpisodeShowInput) (EpisodesResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return EpisodesResponse{}, err
	}
	defer repoStore.Close()
	rows, err := repoStore.ListEpisodesForProblem(ctx, input.ProblemID)
	if err != nil {
		return EpisodesResponse{}, err
	}
	resp := EpisodesResponse{OK: true, Command: "episode list", Store: dbPath, Episodes: []EpisodeView{}}
	for _, row := range rows {
		rec, err := repoStore.GetEpisode(ctx, row.ID)
		if err != nil {
			return EpisodesResponse{}, err
		}
		resp.Episodes = append(resp.Episodes, episodeView(rec))
	}
	return resp, nil
}
