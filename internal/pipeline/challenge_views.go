package pipeline

import (
	"context"
	"fmt"

	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/invariant"
	"github.com/instagrim-dev/newf/internal/store"
)

// --- inputs ---

// ChallengeInput attacks one candidate invariant (or, with All, every
// challengeable candidate for the problem).
type ChallengeInput struct {
	DBPath      string
	InvariantID string
	ProblemID   string
	All         bool
	JSONOutput  bool
}

// InvariantStateInput reads one candidate's lifecycle state.
type InvariantStateInput struct {
	DBPath      string
	InvariantID string
	JSONOutput  bool
}

// InvariantStatesInput lists candidates with their current state.
type InvariantStatesInput struct {
	DBPath     string
	ProblemID  string
	State      string
	JSONOutput bool
}

// EstablishInput is the code-gated surviving->established transition (KTD-2):
// it requires INDEPENDENT, non-model evidence — a persisted source snapshot +
// locator — and is never produced by a provider.
type EstablishInput struct {
	DBPath      string
	InvariantID string
	SnapshotID  string
	Locator     string
	Note        string
	JSONOutput  bool
}

// --- views ---

// ChallengeEvidenceView is one evidence handle behind a challenge.
type ChallengeEvidenceView struct {
	Kind        string `json:"kind"`
	ClusterID   string `json:"cluster_id,omitempty"`
	SignatureID string `json:"signature_id,omitempty"`
	SnapshotID  string `json:"snapshot_id,omitempty"`
	Detail      string `json:"detail,omitempty"`
}

// ChallengeView is one persisted attack + its verified outcome.
type ChallengeView struct {
	ID             string                  `json:"id"`
	ChallengeType  string                  `json:"challenge_type"`
	ClaimedVerdict string                  `json:"claimed_verdict,omitempty"`
	ResultSummary  string                  `json:"result_summary"`
	Detail         string                  `json:"detail,omitempty"`
	Evidence       []ChallengeEvidenceView `json:"evidence,omitempty"`
	Transitions    []string                `json:"transitions,omitempty"`
}

// InvariantChallengeReport is the campaign result for one candidate.
type InvariantChallengeReport struct {
	InvariantID string          `json:"invariant_id"`
	StateBefore string          `json:"state_before"`
	StateAfter  string          `json:"state_after"`
	Challenges  []ChallengeView `json:"challenges"`
	RunID       string          `json:"run_id"`
}

// ChallengeCommandResponse is returned by `newf challenge`.
type ChallengeCommandResponse struct {
	OK      bool                       `json:"ok"`
	Command string                     `json:"command"`
	Store   string                     `json:"store"`
	Reports []InvariantChallengeReport `json:"reports"`
}

// InvariantStateView is one candidate + its current lifecycle state.
type InvariantStateView struct {
	InvariantID          string `json:"invariant_id"`
	State                string `json:"state"`
	AsOf                 string `json:"as_of"`
	Statement            string `json:"statement"`
	PredicateFingerprint string `json:"predicate_fingerprint"`
	InvariantRevisionID  string `json:"invariant_revision_id"`
	AssociationStatus    string `json:"association_status"`
}

// InvariantStateResponse is returned by `newf invariant state`.
type InvariantStateResponse struct {
	OK         bool               `json:"ok"`
	Command    string             `json:"command"`
	Store      string             `json:"store"`
	Invariant  InvariantStateView `json:"invariant"`
	Challenges []ChallengeView    `json:"challenges"`
}

// InvariantStatesResponse is returned by `newf invariant list --state`.
type InvariantStatesResponse struct {
	OK         bool                 `json:"ok"`
	Command    string               `json:"command"`
	Store      string               `json:"store"`
	State      string               `json:"state,omitempty"`
	Invariants []InvariantStateView `json:"invariants"`
}

// EstablishResponse is returned by `newf invariant establish`.
type EstablishResponse struct {
	OK        bool               `json:"ok"`
	Command   string             `json:"command"`
	Store     string             `json:"store"`
	Invariant InvariantStateView `json:"invariant"`
}

func invariantStateView(r store.InvariantStateRow) InvariantStateView {
	return InvariantStateView{
		InvariantID:          r.InvariantID,
		State:                r.State,
		AsOf:                 r.AsOf,
		Statement:            r.Statement,
		PredicateFingerprint: r.PredicateFingerprint,
		InvariantRevisionID:  r.InvariantRevisionID,
		AssociationStatus:    r.AssociationStatus,
	}
}

func challengeView(ch store.ChallengeRecord) ChallengeView {
	view := ChallengeView{
		ID:             ch.ID,
		ChallengeType:  ch.ChallengeType,
		ClaimedVerdict: ch.ClaimedVerdict,
		ResultSummary:  ch.ResultSummary,
		Detail:         ch.Detail,
		Transitions:    ch.Transitions,
	}
	for _, ev := range ch.Evidence {
		view.Evidence = append(view.Evidence, ChallengeEvidenceView{
			Kind: ev.Kind, ClusterID: ev.ClusterID, SignatureID: ev.SignatureID,
			SnapshotID: ev.SnapshotID, Detail: ev.Detail,
		})
	}
	return view
}

// ShowInvariantState reads one candidate's current state + challenge history.
func (a *App) ShowInvariantState(ctx context.Context, input InvariantStateInput) (InvariantStateResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return InvariantStateResponse{}, err
	}
	defer repoStore.Close()

	state, err := repoStore.GetInvariantState(ctx, input.InvariantID)
	if err != nil {
		return InvariantStateResponse{}, err
	}
	history, err := repoStore.ListChallengesForInvariant(ctx, input.InvariantID)
	if err != nil {
		return InvariantStateResponse{}, err
	}
	resp := InvariantStateResponse{OK: true, Command: "invariant state", Store: dbPath, Invariant: invariantStateView(state)}
	for _, ch := range history {
		resp.Challenges = append(resp.Challenges, challengeView(ch))
	}
	return resp, nil
}

// ListInvariantStates lists candidates + current states, optionally filtered.
// Filtered to surviving/established this is the M5.1 frontier read surface.
func (a *App) ListInvariantStates(ctx context.Context, input InvariantStatesInput) (InvariantStatesResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return InvariantStatesResponse{}, err
	}
	defer repoStore.Close()

	rows, err := repoStore.ListInvariantStates(ctx, input.ProblemID, input.State)
	if err != nil {
		return InvariantStatesResponse{}, err
	}
	resp := InvariantStatesResponse{OK: true, Command: "invariant list", Store: dbPath, State: input.State}
	for _, r := range rows {
		resp.Invariants = append(resp.Invariants, invariantStateView(r))
	}
	return resp, nil
}

// EstablishInvariant performs the code-gated surviving->established transition
// (KTD-2). The gate is epistemic, not structural: the DB trigger permits the
// transition, but this service refuses it without independent, non-model
// evidence — a persisted source snapshot (+ locator) recorded as
// independent_source evidence on an independent-verification challenge. No
// provider path can reach this; model agreement tops out at `surviving`.
func (a *App) EstablishInvariant(ctx context.Context, input EstablishInput) (EstablishResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return EstablishResponse{}, err
	}
	defer repoStore.Close()

	state, err := repoStore.GetInvariantState(ctx, input.InvariantID)
	if err != nil {
		return EstablishResponse{}, err
	}
	if state.State != "surviving" {
		return EstablishResponse{}, fmt.Errorf("invariant %s is %q; only a surviving invariant can be established", input.InvariantID, state.State)
	}
	if input.SnapshotID == "" || input.Locator == "" {
		return EstablishResponse{}, fmt.Errorf("establishing requires independent evidence: --snapshot and --locator are mandatory (model judgment alone reaches at most surviving)")
	}
	snapshot, err := repoStore.GetSourceSnapshot(ctx, input.SnapshotID)
	if err != nil {
		return EstablishResponse{}, fmt.Errorf("independent evidence snapshot: %w", err)
	}

	revID, err := repoStore.FindInvariantRevisionForCandidate(ctx, input.InvariantID)
	if err != nil {
		return EstablishResponse{}, err
	}
	revision, err := repoStore.GetInvariantRevision(ctx, revID)
	if err != nil {
		return EstablishResponse{}, err
	}

	now := a.now()
	run, err := repoStore.CreateRun(ctx, domain.NewRun{
		ID:          domain.NewRunID(now),
		ProblemID:   revision.ProblemID,
		Operation:   "invariant establish",
		Status:      domain.RunStatusRunning,
		InputRef:    "invariant:" + input.InvariantID,
		ToolName:    "newf",
		ToolVersion: a.version,
		StartedAt:   now,
		CompletedAt: now,
	})
	if err != nil {
		return EstablishResponse{}, err
	}

	detail := "operator-supplied independent verification"
	if input.Note != "" {
		detail = input.Note
	}
	campaign := store.ChallengeCampaignRecord{
		ProblemID:   revision.ProblemID,
		RunID:       run.ID,
		InvariantID: input.InvariantID,
		Invocation: store.InvariantProviderInvocation{
			ID:            domain.NewProviderInvocationID(now),
			RunID:         run.ID,
			ProviderName:  "operator",
			ModelName:     "independent-verification",
			SchemaVersion: invariant.PredicateSchemaV1,
			RequestHash:   "independent:" + snapshot.ID,
			CreatedAt:     now.Format(timeLayout),
		},
		Challenges: []store.ChallengeRecord{{
			ID:             domain.NewInvariantChallengeID(now),
			InvariantID:    input.InvariantID,
			ChallengeType:  string(invariant.ChallengeIndependentVerification),
			ClaimedVerdict: "supports",
			ResultSummary:  "confirmed",
			Detail:         detail,
			CreatedAt:      now.Format(timeLayout),
			Evidence: []store.ChallengeEvidenceRow{{
				Kind:       invariant.EvidenceIndependentSource,
				SnapshotID: snapshot.ID,
				Detail:     input.Locator,
				Ordinal:    0,
			}},
			Transitions: []string{"established"},
		}},
	}
	if err := repoStore.PersistChallengeCampaign(ctx, campaign); err != nil {
		a.failRun(ctx, repoStore, run.ID, err)
		return EstablishResponse{}, err
	}
	if err := a.finalizeRun(ctx, repoStore, run.ID, nil); err != nil {
		return EstablishResponse{}, err
	}

	after, err := repoStore.GetInvariantState(ctx, input.InvariantID)
	if err != nil {
		return EstablishResponse{}, err
	}
	return EstablishResponse{OK: true, Command: "invariant establish", Store: dbPath, Invariant: invariantStateView(after)}, nil
}
