package pipeline

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/invariant"
	"github.com/instagrim-dev/newf/internal/provider"
	"github.com/instagrim-dev/newf/internal/store"
)

// biasOnlyChallenger proposes only a bias-critique, so a candidate whose
// support genuinely holds ends the campaign `surviving` deterministically.
type biasOnlyChallenger struct{}

func (biasOnlyChallenger) Challenge(_ context.Context, req provider.ChallengeRequest) (provider.ChallengeResponse, error) {
	return provider.ChallengeResponse{
		Proposals: []provider.ChallengeProposal{{
			Type:           invariant.ChallengeBiasCritique,
			Rationale:      "recount support under redundancy",
			ClaimedVerdict: "support_collapses",
		}},
		Metadata:        provider.Metadata{ProviderName: "fixture", ProviderVersion: "test", ModelName: "bias-only", SchemaVersion: invariant.PredicateSchemaV1},
		RequestPayload:  "req",
		ResponsePayload: "resp",
	}, nil
}

// minPreservesMiner proposes `preserves contains <id>` for the LEXICALLY
// SMALLEST preserves id across all families — insensitive to family order (and
// therefore to the randomness of freshly minted cluster ULIDs), so separate
// stores mine the same predicate.
type minPreservesMiner struct{}

func (minPreservesMiner) Identity() provider.MinerIdentity {
	return provider.MinerIdentity{
		ContractVersion: provider.InvariantMinerVersion,
		ProviderName:    "fixture", ProviderVersion: "v1", ModelName: "min-preserves",
	}
}

func (minPreservesMiner) Mine(_ context.Context, req provider.MiningRequest) (provider.MiningResponse, error) {
	var min string
	for _, f := range req.Families {
		for _, id := range f.Preserves {
			if min == "" || string(id) < min {
				min = string(id)
			}
		}
	}
	var proposals []provider.CandidateProposal
	if min != "" {
		proposals = append(proposals, provider.CandidateProposal{
			Predicate: invariant.Predicate{
				Schema: invariant.PredicateSchemaV1,
				Root:   invariant.Node{Op: invariant.OpContains, Field: invariant.FieldPreserves, CanonicalID: min},
			},
			Statement:        "failed methods preserve " + min,
			AbstractionLevel: "mechanism",
		})
	}
	return provider.MiningResponse{
		Proposals:       proposals,
		Metadata:        provider.Metadata{ProviderName: "fixture", ProviderVersion: "v1", ModelName: "min-preserves", SchemaVersion: invariant.PredicateSchemaV1},
		RequestPayload:  "req",
		ResponsePayload: "resp",
	}, nil
}

// mineOneCandidate seeds the corpus, clusters, builds the failure space, and
// mines one candidate, returning its id plus a snapshot usable as independent
// evidence.
func mineOneCandidate(t *testing.T, ctx context.Context, app *App, dbPath string) (problemID, invariantID, snapshotID string) {
	t.Helper()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	problemID, runID, snapshotID := seedProblemAndSnapshot(t, ctx, dbPath, now)
	seedClusterCorpus(t, ctx, app, dbPath, problemID, runID, snapshotID, canon.VocabularyMechanismV1)
	if _, err := app.BuildClustering(ctx, ClusterBuildInput{DBPath: dbPath, ProblemID: problemID}); err != nil {
		t.Fatalf("cluster build: %v", err)
	}
	if _, err := app.BuildFailureSpace(ctx, FailureSpaceBuildInput{DBPath: dbPath, ProblemID: problemID}); err != nil {
		t.Fatalf("failure-space build: %v", err)
	}
	mined, err := app.MineInvariants(ctx, InvariantMineInput{DBPath: dbPath, ProblemID: problemID, MinSupport: 1})
	if err != nil {
		t.Fatalf("mine: %v", err)
	}
	if mined.Revision.CandidateCount == 0 {
		t.Fatal("expected at least one mined candidate")
	}
	return problemID, mined.Revision.Candidates[0].ID, snapshotID
}

// inconclusiveOnlyChallenger proposes a single synthetic-counterexample with NO
// construction supplied. That attack is inadmissible: it evaluates no eligible
// population and reaches no definite negative. It exercises the F2 gate.
type inconclusiveOnlyChallenger struct{}

func (inconclusiveOnlyChallenger) Challenge(_ context.Context, req provider.ChallengeRequest) (provider.ChallengeResponse, error) {
	return provider.ChallengeResponse{
		Proposals: []provider.ChallengeProposal{{
			Type:           invariant.ChallengeSyntheticCounterexample,
			Rationale:      "claims a construction but supplies none",
			ClaimedVerdict: "violates",
			// Synthetic intentionally nil: no construction to evaluate.
		}},
		Metadata:        provider.Metadata{ProviderName: "fixture", ProviderVersion: "test", ModelName: "inconclusive-only", SchemaVersion: invariant.PredicateSchemaV1},
		RequestPayload:  "req",
		ResponsePayload: "resp",
	}, nil
}

// TestIntegrationInconclusiveCampaignDoesNotEarnSurviving is the F2 regression:
// a campaign made up entirely of inadmissible/inconclusive attempts (here, a
// lone synthetic-counterexample with no construction) must NOT transition the
// invariant to `surviving`. Survival is earned by a completed applicable
// negative search, never granted by fallback. The invariant opens the campaign
// (-> challenged) and stops there.
func TestIntegrationInconclusiveCampaignDoesNotEarnSurviving(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = dataDrivenMiner{}
	app.challengerFn = inconclusiveOnlyChallenger{}

	_, invID, _ := mineOneCandidate(t, ctx, app, dbPath)

	resp, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID})
	if err != nil {
		t.Fatalf("challenge: %v", err)
	}
	if got := resp.Reports[0].StateAfter; got == "surviving" {
		t.Fatalf("an all-inconclusive campaign must NOT earn surviving (F2), got %q", got)
	}
	if got := resp.Reports[0].StateAfter; got != "challenged" {
		t.Fatalf("an all-inconclusive campaign should stop at challenged, got %q", got)
	}

	// G2 resumability: `challenged` must NOT be a dead end. A follow-up campaign
	// with a completed-negative attack (bias recount that holds) resumes the
	// undecided candidate and legitimately earns `surviving`.
	app.challengerFn = biasOnlyChallenger{}
	resume, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID})
	if err != nil {
		t.Fatalf("resume challenge on a challenged candidate must be allowed (G2), got error: %v", err)
	}
	if got := resume.Reports[0].StateAfter; got != "surviving" {
		t.Fatalf("a resumed campaign with a completed-negative search should reach surviving, got %q", got)
	}
}

// splitTooFewChildrenChallenger proposes a single split with ONE child. The
// split verifier rejects it as INADMISSIBLE (a split requires >=2 children); the
// pipeline must not count that rejected attack as a completed applicable check.
type splitTooFewChildrenChallenger struct{}

func (splitTooFewChildrenChallenger) Challenge(_ context.Context, req provider.ChallengeRequest) (provider.ChallengeResponse, error) {
	pred, _ := invariant.ParsePredicate(req.PredicateJSON)
	return provider.ChallengeResponse{
		Proposals: []provider.ChallengeProposal{{
			Type:           invariant.ChallengeSplit,
			Rationale:      "one child is not a split",
			ClaimedVerdict: "splits",
			Children:       []invariant.Predicate{pred}, // <2 children: inadmissible
		}},
		Metadata:        provider.Metadata{ProviderName: "fixture", ProviderVersion: "test", ModelName: "split-1", SchemaVersion: invariant.PredicateSchemaV1},
		RequestPayload:  "req",
		ResponsePayload: "resp",
	}, nil
}

// TestIntegrationInadmissibleSplitDoesNotEarnSurviving is the G2 regression for
// the specific leak the review flagged: an inadmissible split (fewer than two
// children) must not count as a completed applicable check. A campaign of only
// that attack opens (-> challenged) and stops there — never `surviving`.
func TestIntegrationInadmissibleSplitDoesNotEarnSurviving(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = dataDrivenMiner{}
	app.challengerFn = splitTooFewChildrenChallenger{}

	_, invID, _ := mineOneCandidate(t, ctx, app, dbPath)

	resp, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID})
	if err != nil {
		t.Fatalf("challenge: %v", err)
	}
	if got := resp.Reports[0].StateAfter; got != "challenged" {
		t.Fatalf("an inadmissible-split-only campaign must stop at challenged (G2), got %q", got)
	}
}

// TestIntegrationChallengeCampaign runs the deriving challenger against a mined
// candidate and asserts: the run completes, every executed challenge is
// persisted with its verified result, the first challenge opens the campaign
// (-> challenged), and the final state is one the trigger-guarded ledger
// reached (never proposed, never established via a provider).
func TestIntegrationChallengeCampaign(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = dataDrivenMiner{}

	problemID, invID, _ := mineOneCandidate(t, ctx, app, dbPath)

	resp, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID})
	if err != nil {
		t.Fatalf("challenge: %v", err)
	}
	if len(resp.Reports) != 1 {
		t.Fatalf("expected 1 report, got %d", len(resp.Reports))
	}
	report := resp.Reports[0]
	if report.StateBefore != "proposed" {
		t.Fatalf("state before = %q, want proposed", report.StateBefore)
	}
	if report.StateAfter == "proposed" || report.StateAfter == "operator_attested" {
		t.Fatalf("state after = %q: a campaign must move past proposed and can never reach operator_attested", report.StateAfter)
	}
	if len(report.Challenges) < 3 {
		t.Fatalf("expected >=3 executed challenges, got %d", len(report.Challenges))
	}
	if len(report.Challenges[0].Transitions) == 0 || report.Challenges[0].Transitions[0] != "challenged" {
		t.Fatalf("first challenge must open the campaign, got %+v", report.Challenges[0].Transitions)
	}
	// Run lifecycle reflects the real outcome.
	run, err := app.ShowRun(ctx, LookupInput{DBPath: dbPath, ID: report.RunID})
	if err != nil {
		t.Fatalf("run show: %v", err)
	}
	if run.Run.Status != "completed" {
		t.Fatalf("challenge run status = %q, want completed", run.Run.Status)
	}
	// State is queryable and history round-trips.
	state, err := app.ShowInvariantState(ctx, InvariantStateInput{DBPath: dbPath, InvariantID: invID})
	if err != nil {
		t.Fatalf("invariant state: %v", err)
	}
	if state.Invariant.State != report.StateAfter {
		t.Fatalf("state view %q != report %q", state.Invariant.State, report.StateAfter)
	}
	if len(state.Challenges) != len(report.Challenges) {
		t.Fatalf("history %d != campaign %d", len(state.Challenges), len(report.Challenges))
	}
	// A falsified candidate is terminal: a second campaign is refused.
	if report.StateAfter == "falsified" {
		if _, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID}); err == nil {
			t.Fatal("falsified must be terminal")
		}
	}
	_ = problemID
}

// TestIntegrationEstablishedIsCodeGated proves KTD-2 end to end: a candidate
// that SURVIVES its campaign cannot be established by model judgment (no
// snapshot -> refused), only by operator-supplied independent evidence; and a
// non-surviving candidate is refused regardless.
func TestIntegrationEstablishedIsCodeGated(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = dataDrivenMiner{}
	app.challengerFn = biasOnlyChallenger{}

	problemID, invID, snapshotID := mineOneCandidate(t, ctx, app, dbPath)

	// A proposed (unchallenged) candidate cannot be established at all.
	if _, err := app.EstablishInvariant(ctx, EstablishInput{DBPath: dbPath, InvariantID: invID, SnapshotID: "snap_x", Locator: "l"}); err == nil {
		t.Fatal("only a surviving invariant may be established")
	}

	// bias-only campaign whose support holds -> surviving.
	resp, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID})
	if err != nil {
		t.Fatalf("challenge: %v", err)
	}
	if resp.Reports[0].StateAfter != "surviving" {
		t.Fatalf("state after bias-only campaign = %q, want surviving", resp.Reports[0].StateAfter)
	}

	// Surviving + NO independent evidence -> refused (the gate).
	_, err = app.EstablishInvariant(ctx, EstablishInput{DBPath: dbPath, InvariantID: invID})
	if err == nil || !strings.Contains(err.Error(), "independent evidence") {
		t.Fatalf("expected independent-evidence refusal, got %v", err)
	}

	// Evidence from ANOTHER problem is refused: `operator_attested` must not be
	// mintable from any bytes anywhere in the store.
	otherSnap := seedOtherProblemSnapshot(t, ctx, dbPath, now.Add(time.Hour))
	_, err = app.EstablishInvariant(ctx, EstablishInput{DBPath: dbPath, InvariantID: invID, SnapshotID: otherSnap, Locator: "sec. 1"})
	if err == nil || !strings.Contains(err.Error(), "belongs to problem") {
		t.Fatalf("expected cross-problem evidence refusal, got %v", err)
	}

	// Surviving + a real persisted snapshot -> operator_attested.
	est, err := app.EstablishInvariant(ctx, EstablishInput{
		DBPath: dbPath, InvariantID: invID,
		SnapshotID: snapshotID, Locator: "sec. 3, theorem 2", Note: "independent proof",
	})
	if err != nil {
		t.Fatalf("establish: %v", err)
	}
	if est.Invariant.State != "operator_attested" {
		t.Fatalf("state = %q, want operator_attested", est.Invariant.State)
	}

	// The read surface exposes it under --state operator_attested.
	list, err := app.ListInvariantStates(ctx, InvariantStatesInput{DBPath: dbPath, ProblemID: problemID, State: "operator_attested"})
	if err != nil {
		t.Fatalf("list established: %v", err)
	}
	if len(list.Invariants) != 1 || list.Invariants[0].InvariantID != invID {
		t.Fatalf("expected the established candidate on the read surface, got %+v", list.Invariants)
	}

	// D1 (frontier read surface): an operator_attested invariant remains a legal
	// frontier target — attestation strengthens the invariant's evidence and must
	// never REMOVE it from search-policy influence.
	gen, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("frontier generate: %v", err)
	}
	targeted := false
	for _, p := range gen.Generation.Proposals {
		for _, tgt := range p.Targets {
			if tgt.InvariantID == invID {
				targeted = true
			}
		}
	}
	if !targeted {
		t.Fatalf("operator_attested invariant %s must remain a frontier target; proposals: %+v", invID, gen.Generation.Proposals)
	}
}

// TestIntegrationAttestationIsNotVerification is the F1 regression: an operator
// attestation must NOT masquerade as machine-confirmed evidence. A valid,
// same-problem snapshot whose bytes are UNRELATED to the invariant, paired with
// an INVENTED locator, still transitions only to `operator_attested` — never to
// a status named `established` (which no longer exists) — and the recorded
// challenge is labeled an operator attestation, not a verification of the claim.
// The system does not inspect snapshot content or validate the locator; the
// honest status name is the whole point.
func TestIntegrationAttestationIsNotVerification(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = dataDrivenMiner{}
	app.challengerFn = biasOnlyChallenger{}

	_, invID, snapshotID := mineOneCandidate(t, ctx, app, dbPath)
	if _, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID}); err != nil {
		t.Fatalf("challenge: %v", err)
	}

	// Unrelated text + an invented locator: the strongest honest status is
	// operator_attested, explicitly NOT a machine-verified `established`.
	est, err := app.EstablishInvariant(ctx, EstablishInput{
		DBPath: dbPath, InvariantID: invID,
		SnapshotID: snapshotID, Locator: "appendix Z, line 999 (invented)",
		Note: "operator attests without machine verification",
	})
	if err != nil {
		t.Fatalf("attest: %v", err)
	}
	if est.Invariant.State == "established" {
		t.Fatal("F1: no status may be named `established` (attestation is not verification)")
	}
	if est.Invariant.State != "operator_attested" {
		t.Fatalf("state = %q, want operator_attested", est.Invariant.State)
	}

	// G4: attestation must NOT make the hypothesis immune to further challenge.
	// A campaign against an operator_attested invariant is accepted (not rejected
	// as unchallengeable) and drives a legitimate transition.
	postAttest, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID})
	if err != nil {
		t.Fatalf("operator_attested must remain challengeable (G4), got error: %v", err)
	}
	switch postAttest.Reports[0].StateAfter {
	case "surviving", "challenged", "weaken", "falsified":
		// any legitimate post-challenge state is fine; the point is it was attackable
	default:
		t.Fatalf("challenge of operator_attested produced unexpected state %q", postAttest.Reports[0].StateAfter)
	}
}

// seedOtherProblemSnapshot creates a SECOND problem with its own snapshot, for
// asserting that establish-evidence is problem-scoped.
func seedOtherProblemSnapshot(t *testing.T, ctx context.Context, dbPath string, now time.Time) string {
	t.Helper()
	repo := openTestStore(t, ctx, dbPath)
	defer repo.Close()

	runID := domain.NewRunID(now)
	problemID := domain.NewProblemID(now)
	problem, run, err := repo.CreateProblemWithRun(ctx, domain.NewProblem{
		ID: problemID, Slug: "other-problem", Statement: "Other problem",
		Status: domain.ProblemStatusActive, CreatedAt: now, CreatedByRunID: runID,
	}, domain.NewRun{
		ID: runID, ProblemID: problemID, Operation: "init", Status: domain.RunStatusInitialized,
		InputRef: "problem_slug:other-problem", ToolName: "newf", ToolVersion: "test",
		StartedAt: now, CompletedAt: now,
	})
	if err != nil {
		t.Fatalf("seed other problem: %v", err)
	}
	admission, err := repo.CreateSourceSnapshot(ctx, store.SnapshotAdmission{
		ProblemID: problem.ID, Kind: domain.SourceKindLocalPath,
		LogicalName: "other.md", Origin: "/tmp/other.md", SHA256: "cafebabe",
		ByteLength: 8, MediaType: "text/markdown", ObjectPath: "sha256/ca/cafebabe",
		IngestRunID: run.ID, ObservedAt: now,
	})
	if err != nil {
		t.Fatalf("seed other snapshot: %v", err)
	}
	return admission.Snapshot.ID
}

// TestIntegrationChallengeAllIsDeterministic runs --all twice on identical
// stores and asserts identical final states in identical candidate order
// (KTD-4: campaign determinism / replayability).
func TestIntegrationChallengeAllIsDeterministic(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)

	runCampaign := func() []string {
		app, dbPath := newRealStoreApp(t, now)
		app.invariantMinerFn = minPreservesMiner{}
		problemID, _, _ := mineOneCandidate(t, ctx, app, dbPath)
		resp, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, ProblemID: problemID, All: true})
		if err != nil {
			t.Fatalf("challenge --all: %v", err)
		}
		var out []string
		for _, r := range resp.Reports {
			out = append(out, r.StateAfter)
			for _, ch := range r.Challenges {
				out = append(out, ch.ChallengeType+"="+ch.ResultSummary)
			}
		}
		return out
	}

	a, b := runCampaign(), runCampaign()
	if len(a) != len(b) {
		t.Fatalf("campaign shapes differ: %v vs %v", a, b)
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("campaign not deterministic at %d: %q vs %q", i, a[i], b[i])
		}
	}
}

// TestDerivationMinerVersionIdentity is the F5 regression for derivation
// identity: the derived-revision reuse key is (problem, failure_space,
// miner_version, predicate_schema, min_support), so miner_version MUST fold the
// parent set and canonical child fingerprints — otherwise two DIFFERENT parents
// (or child sets) splitting under the same failure space + threshold collide on
// the key and the second silently reuses the first's children. This proves the
// version is (a) stable for identical inputs, and DISTINCT for (b) a different
// parent set and (c) a different child set.
func TestDerivationMinerVersionIdentity(t *testing.T) {
	child := func(id string) invariant.Predicate {
		return invariant.Predicate{Schema: invariant.PredicateSchemaV1, Root: invariant.Node{Op: invariant.OpContains, Field: invariant.FieldPreserves, CanonicalID: id}}
	}
	base := derivationMinerVersion(invariant.ChallengeSplit,
		[]string{"fp_parentA"},
		[]invariant.Predicate{child("core.operator.a"), child("core.operator.b")})

	// (a) identical inputs (order-insensitive) -> identical identity.
	same := derivationMinerVersion(invariant.ChallengeSplit,
		[]string{"fp_parentA"},
		[]invariant.Predicate{child("core.operator.b"), child("core.operator.a")})
	if base != same {
		t.Fatalf("identical derivation inputs must share identity: %q vs %q", base, same)
	}

	// (b) different parent set -> different identity (the collision the review names).
	otherParent := derivationMinerVersion(invariant.ChallengeSplit,
		[]string{"fp_parentB"},
		[]invariant.Predicate{child("core.operator.a"), child("core.operator.b")})
	if base == otherParent {
		t.Fatalf("distinct parents must not collide on derivation identity (F5): %q", base)
	}

	// (c) different child set -> different identity.
	otherChildren := derivationMinerVersion(invariant.ChallengeSplit,
		[]string{"fp_parentA"},
		[]invariant.Predicate{child("core.operator.a"), child("core.operator.c")})
	if base == otherChildren {
		t.Fatalf("distinct child sets must not collide on derivation identity (F5): %q", base)
	}

	// (d) relation participates in identity too.
	asMerge := derivationMinerVersion(invariant.ChallengeMerge,
		[]string{"fp_parentA"},
		[]invariant.Predicate{child("core.operator.a"), child("core.operator.b")})
	if base == asMerge {
		t.Fatalf("relation must participate in derivation identity: %q", base)
	}
}
