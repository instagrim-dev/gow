package pipeline

import (
	"context"
	"fmt"

	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/invariant"
	"github.com/instagrim-dev/newf/internal/store"
	"github.com/instagrim-dev/newf/internal/verify"
)

// --- inputs ---

// ChallengeInput attacks one candidate invariant (or, with All, every
// challengeable candidate for the problem). Population selects the evidence
// the campaign's searches run against (v34/S1):
//   - "latest" (default): the newest cluster run compatible with the claim's
//     schema/vocabulary — newly ingested, reclustered evidence enters the
//     known-counterexample and success-preserving checks;
//   - "discovery": historical replay against the population the claim was
//     mined over.
//
// Claim-scope attacks (bias-critique recount, split, merge, synthetic
// grounding) always run over the discovery population regardless of policy.
type ChallengeInput struct {
	DBPath      string
	InvariantID string
	ProblemID   string
	All         bool
	Population  string
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

// EstablishInput is the code-gated surviving->operator_attested transition (KTD-2):
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
	// VerificationSubject (v38/#21): what OBJECT this evidence is about —
	// annotation / realization / domain-goal. Empty only for pre-v38 history.
	VerificationSubject string `json:"verification_subject,omitempty"`
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
	// BoundaryDelta is the typed refinement a confirmed challenge derived
	// (v36/S5); absent for unconfirmed/inert challenges and pre-v36 history.
	BoundaryDelta *BoundaryDeltaView `json:"boundary_delta,omitempty"`
}

// BoundaryDeltaView is the machine-readable boundary refinement of one
// confirmed challenge: the minimal separating condition observed, in canonical
// (paraphrase-stable) predicate form.
type BoundaryDeltaView struct {
	Kind                 string   `json:"kind"`
	PredicateFingerprint string   `json:"predicate_fingerprint"`
	Condition            string   `json:"condition"`
	MeasuredSupport      int      `json:"measured_support,omitempty"`
	SupportThreshold     int      `json:"support_threshold,omitempty"`
	ChildFingerprints    []string `json:"child_fingerprints,omitempty"`
}

// InvariantChallengeReport is the campaign result for one candidate. The
// population fields (v34/S1) identify the claim's discovery population, the
// assessment population the campaign's evidence searches ran against, and the
// requested policy; they are empty only for pre-v34 history.
type InvariantChallengeReport struct {
	InvariantID            string          `json:"invariant_id"`
	StateBefore            string          `json:"state_before"`
	StateAfter             string          `json:"state_after"`
	Challenges             []ChallengeView `json:"challenges"`
	RunID                  string          `json:"run_id"`
	PopulationPolicy       string          `json:"population_policy,omitempty"`
	DiscoveryClusterRunID  string          `json:"discovery_cluster_run_id,omitempty"`
	AssessmentClusterRunID string          `json:"assessment_cluster_run_id,omitempty"`
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

// InvariantStateResponse is returned by `newf invariant state`. Claim is the
// governing authored claim form (v39/#21), nil when none has been authored —
// in which case a recurring candidate keeps association-level refutation
// semantics (no invented universality).
type InvariantStateResponse struct {
	OK         bool               `json:"ok"`
	Command    string             `json:"command"`
	Store      string             `json:"store"`
	Invariant  InvariantStateView `json:"invariant"`
	Claim      *ClaimFormView     `json:"claim,omitempty"`
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
			VerificationSubject: ev.VerificationSubject,
		})
	}
	if ch.Delta != nil {
		view.BoundaryDelta = &BoundaryDeltaView{
			Kind:                 ch.Delta.Kind,
			PredicateFingerprint: ch.Delta.PredicateFingerprint,
			Condition:            ch.Delta.Condition,
			MeasuredSupport:      ch.Delta.MeasuredSupport,
			SupportThreshold:     ch.Delta.SupportThreshold,
			ChildFingerprints:    ch.Delta.ChildFingerprints,
		}
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
	resp := InvariantStateResponse{OK: true, Command: "invariant state", Store: dbPath, Invariant: invariantStateView(state), Challenges: []ChallengeView{}}
	if form, found, err := repoStore.GetLatestInvariantClaimForm(ctx, input.InvariantID); err != nil {
		return InvariantStateResponse{}, err
	} else if found {
		v := claimFormView(form)
		resp.Claim = &v
	}
	for _, ch := range history {
		resp.Challenges = append(resp.Challenges, challengeView(ch))
	}
	return resp, nil
}

// ListInvariantStates lists candidates + current states, optionally filtered.
// Filtered to surviving/operator_attested this is the M5.1 frontier read
// surface (see docs/frontier-generation.md).
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
	resp := InvariantStatesResponse{OK: true, Command: "invariant list", Store: dbPath, State: input.State, Invariants: []InvariantStateView{}}
	for _, r := range rows {
		resp.Invariants = append(resp.Invariants, invariantStateView(r))
	}
	return resp, nil
}

// EstablishInvariant performs the code-gated surviving->operator_attested
// transition (KTD-2). The gate is epistemic, not structural: the DB trigger
// permits the transition, but this service refuses it without independent,
// non-model evidence — a persisted, same-problem source snapshot (+ locator)
// recorded as independent_source evidence on an independent-verification
// challenge. No provider path can reach this; model agreement tops out at
// `surviving`.
//
// It records an OPERATOR ATTESTATION, not a machine verification (F1): the
// service does not inspect the snapshot content, validate the locator, prove
// independence, or check a proof against the predicate. The resulting status is
// therefore `operator_attested` — useful provenance that must NOT masquerade as
// machine-confirmed evidence. A claim-specific verification contract (a
// deterministic/formal check that actually validates the claim) is future work;
// until it exists, this is the strongest status the system honestly offers, and
// it is explicitly an attestation.
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
		return EstablishResponse{}, fmt.Errorf("invariant %s is %q; only a surviving invariant can be attested", input.InvariantID, state.State)
	}
	if input.SnapshotID == "" || input.Locator == "" {
		return EstablishResponse{}, fmt.Errorf("establishing requires independent evidence: --snapshot and --locator are mandatory (model judgment alone reaches at most surviving)")
	}

	revID, err := repoStore.FindInvariantRevisionForCandidate(ctx, input.InvariantID)
	if err != nil {
		return EstablishResponse{}, err
	}
	revision, err := repoStore.GetInvariantRevision(ctx, revID)
	if err != nil {
		return EstablishResponse{}, err
	}

	snapshot, err := repoStore.GetSourceSnapshot(ctx, input.SnapshotID)
	if err != nil {
		return EstablishResponse{}, fmt.Errorf("independent evidence snapshot: %w", err)
	}
	// Evidence must belong to the invariant's problem: `operator_attested` is the
	// strongest epistemic status in the system, and "any bytes anywhere in the
	// store" is not a gate. Cross-problem evidence requires ingesting the source
	// under this problem first, which records the provenance intentionally.
	source, err := repoStore.GetSource(ctx, snapshot.SourceID)
	if err != nil {
		return EstablishResponse{}, fmt.Errorf("independent evidence source: %w", err)
	}
	if source.ProblemID != revision.ProblemID {
		return EstablishResponse{}, fmt.Errorf("independent evidence snapshot %s belongs to problem %s, not %s; ingest the source under this problem to record its provenance", snapshot.ID, source.ProblemID, revision.ProblemID)
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

	detail := "operator-attested independent evidence (not machine-verified)"
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
			ModelName:     "operator-attestation",
			SchemaVersion: invariant.PredicateSchemaV1,
			RequestHash:   "independent:" + snapshot.ID,
			CreatedAt:     now.Format(timeLayout),
		},
		Challenges: []store.ChallengeRecord{{
			ID:            domain.NewInvariantChallengeID(now),
			InvariantID:   input.InvariantID,
			ChallengeType: string(invariant.ChallengeIndependentVerification),
			// The operator ATTESTS support; the system records the attestation and
			// its evidence handle. It does NOT itself verify the claim against the
			// predicate, so the resulting lifecycle state is `operator_attested`, not
			// a machine-confirmed `established` (F1: ModelJudgment/OperatorJudgment
			// != Verification). `confirmed` here means "the attestation was recorded",
			// scoped by the operator_attested target below.
			ClaimedVerdict: "operator-attests-support",
			ResultSummary:  "confirmed",
			Detail:         detail,
			CreatedAt:      now.Format(timeLayout),
			Evidence: []store.ChallengeEvidenceRow{{
				Kind:       invariant.EvidenceIndependentSource,
				SnapshotID: snapshot.ID,
				Detail:     input.Locator,
				Ordinal:    0,
				// v38 (#21): an independent source is consulted about the DOMAIN
				// claim the invariant makes — not about our annotation formatting.
				// The subject axis keeps that explicit; the strength stays
				// operator-attested (this records an attestation, not a
				// machine verification).
				VerificationSubject: string(verify.SubjectDomainGoal),
			}},
			Transitions: []string{"operator_attested"},
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
