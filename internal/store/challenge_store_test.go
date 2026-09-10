package store

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/domain"
)

// persistSampleInvariant persists a mining revision and returns the store plus
// the candidate/run/problem ids the challenge layer needs.
func persistSampleInvariant(t *testing.T) (st *Store, problemID, runID, invariantID string) {
	t.Helper()
	st = openMigratedStore(t)
	rec := sampleRevision(t, st)
	res, err := st.PersistInvariantRevision(context.Background(), rec)
	if err != nil {
		t.Fatalf("persist revision: %v", err)
	}
	return st, rec.ProblemID, rec.RunID, res.Record.Candidates[0].ID
}

func sampleCampaign(problemID, runID, invariantID string, challenges ...ChallengeRecord) ChallengeCampaignRecord {
	now := formatTime(time.Date(2026, 9, 10, 13, 0, 0, 0, time.UTC))
	return ChallengeCampaignRecord{
		ProblemID:   problemID,
		RunID:       runID,
		InvariantID: invariantID,
		Invocation: InvariantProviderInvocation{
			ID: domain.NewProviderInvocationID(time.Now()), RunID: runID,
			ProviderName: "fixture", SchemaVersion: "invariant-predicate/v1",
			RequestHash: "rh", CreatedAt: now,
		},
		Challenges: challenges,
	}
}

func challengeRow(t *testing.T, typ, summary string, transitions ...string) ChallengeRecord {
	t.Helper()
	now := formatTime(time.Date(2026, 9, 10, 13, 0, 0, 0, time.UTC))
	return ChallengeRecord{
		ID:            domain.NewInvariantChallengeID(time.Now()),
		ChallengeType: typ,
		ResultSummary: summary,
		Detail:        "test",
		CreatedAt:     now,
		Transitions:   transitions,
	}
}

func TestChallengeCampaignPersistsEvidenceAndTransitions(t *testing.T) {
	st, problemID, runID, invID := persistSampleInvariant(t)
	ctx := context.Background()

	ch := challengeRow(t, "known-counterexample", "confirmed", "challenged", "falsified")
	ch.ClaimedVerdict = "violates"
	ch.Evidence = []ChallengeEvidenceRow{{Kind: "support_recount", Detail: "recount=0", Ordinal: 0}}
	if err := st.PersistChallengeCampaign(ctx, sampleCampaign(problemID, runID, invID, ch)); err != nil {
		t.Fatalf("persist campaign: %v", err)
	}

	state, err := st.GetInvariantState(ctx, invID)
	if err != nil {
		t.Fatalf("get state: %v", err)
	}
	if state.State != "falsified" {
		t.Fatalf("state = %q, want falsified", state.State)
	}

	history, err := st.ListChallengesForInvariant(ctx, invID)
	if err != nil {
		t.Fatalf("list challenges: %v", err)
	}
	if len(history) != 1 || history[0].ResultSummary != "confirmed" {
		t.Fatalf("history round-trip lost data: %+v", history)
	}
	if len(history[0].Evidence) != 1 || history[0].Evidence[0].Kind != "support_recount" {
		t.Fatalf("evidence round-trip lost data: %+v", history[0].Evidence)
	}
	if len(history[0].Transitions) != 2 || history[0].Transitions[1] != "falsified" {
		t.Fatalf("transitions round-trip lost data: %+v", history[0].Transitions)
	}
}

func TestUnconfirmedChallengeIsInert(t *testing.T) {
	st, problemID, runID, invID := persistSampleInvariant(t)
	ctx := context.Background()

	ch := challengeRow(t, "known-counterexample", "unconfirmed") // no transitions
	if err := st.PersistChallengeCampaign(ctx, sampleCampaign(problemID, runID, invID, ch)); err != nil {
		t.Fatalf("persist campaign: %v", err)
	}
	state, err := st.GetInvariantState(ctx, invID)
	if err != nil {
		t.Fatalf("get state: %v", err)
	}
	if state.State != "proposed" {
		t.Fatalf("unconfirmed challenge must not move state; got %q", state.State)
	}
}

func TestStateMachineRejectsIllegalTransitions(t *testing.T) {
	st, problemID, runID, invID := persistSampleInvariant(t)
	ctx := context.Background()

	// proposed -> surviving skips 'challenged': the campaign aborts atomically.
	bad := challengeRow(t, "bias-critique", "confirmed", "surviving")
	err := st.PersistChallengeCampaign(ctx, sampleCampaign(problemID, runID, invID, bad))
	if err == nil || !strings.Contains(err.Error(), "invalid invariant state transition") {
		t.Fatalf("expected illegal-transition abort, got %v", err)
	}
	// The abort left no partial state: still proposed, no challenge rows.
	state, _ := st.GetInvariantState(ctx, invID)
	if state.State != "proposed" {
		t.Fatalf("aborted campaign must leave state proposed, got %q", state.State)
	}
	history, _ := st.ListChallengesForInvariant(ctx, invID)
	if len(history) != 0 {
		t.Fatalf("aborted campaign must persist nothing, got %d challenges", len(history))
	}

	// Reach falsified, then verify it is terminal.
	kill := challengeRow(t, "known-counterexample", "confirmed", "challenged", "falsified")
	if err := st.PersistChallengeCampaign(ctx, sampleCampaign(problemID, runID, invID, kill)); err != nil {
		t.Fatalf("persist kill campaign: %v", err)
	}
	after := challengeRow(t, "bias-critique", "confirmed", "challenged")
	err = st.PersistChallengeCampaign(ctx, sampleCampaign(problemID, runID, invID, after))
	if err == nil {
		t.Fatal("falsified must be terminal (no further transitions)")
	}
}

func TestTransitionRequiresCounterAllocation(t *testing.T) {
	st, _, runID, invID := persistSampleInvariant(t)
	ctx := context.Background()

	// A raw transition insert without allocating the counter must abort.
	now := formatTime(time.Date(2026, 9, 10, 13, 0, 0, 0, time.UTC))
	chID := domain.NewInvariantChallengeID(time.Now())
	// Seed a bare challenge row to satisfy the FK (via the campaign path with no
	// transitions, which is legal).
	pinv := domain.NewProviderInvocationID(time.Now())
	if _, err := st.db.ExecContext(ctx, `
INSERT INTO provider_invocations(id, run_id, role, provider_name, schema_version, request_hash, created_at)
VALUES(?, ?, 'challenge', 'fixture', 'sv', 'rh', ?)`, pinv, runID, now); err != nil {
		t.Fatalf("seed invocation: %v", err)
	}
	if _, err := st.db.ExecContext(ctx, `
INSERT INTO invariant_challenges(id, invariant_id, run_id, provider_invocation_id, challenge_type, result_summary, created_at)
VALUES(?, ?, ?, ?, 'bias-critique', 'confirmed', ?)`, chID, invID, runID, pinv, now); err != nil {
		t.Fatalf("seed challenge: %v", err)
	}
	_, err := st.db.ExecContext(ctx, `
INSERT INTO invariant_state_transitions(invariant_id, transition_seq, challenge_id, from_state, to_state, created_at)
VALUES(?, 1, ?, 'proposed', 'challenged', ?)`, invID, chID, now)
	if err == nil || !strings.Contains(err.Error(), "prior counter allocation") {
		t.Fatalf("expected counter-allocation abort, got %v", err)
	}
	// Wrong seq after a real allocation also aborts.
	if _, err := st.db.ExecContext(ctx, `INSERT INTO invariant_transition_counters(invariant_id, last_transition_seq) VALUES(?, 1)`, invID); err != nil {
		t.Fatalf("allocate: %v", err)
	}
	_, err = st.db.ExecContext(ctx, `
INSERT INTO invariant_state_transitions(invariant_id, transition_seq, challenge_id, from_state, to_state, created_at)
VALUES(?, 2, ?, 'proposed', 'challenged', ?)`, invID, chID, now)
	if err == nil || !strings.Contains(err.Error(), "atomically allocated") {
		t.Fatalf("expected seq-mismatch abort, got %v", err)
	}
	// from_state mismatch aborts too.
	_, err = st.db.ExecContext(ctx, `
INSERT INTO invariant_state_transitions(invariant_id, transition_seq, challenge_id, from_state, to_state, created_at)
VALUES(?, 1, ?, 'challenged', 'surviving', ?)`, invID, chID, now)
	if err == nil || !strings.Contains(err.Error(), "must match current invariant state") {
		t.Fatalf("expected from_state abort, got %v", err)
	}
}

func TestChallengeRowsImmutable(t *testing.T) {
	st, problemID, runID, invID := persistSampleInvariant(t)
	ctx := context.Background()
	ch := challengeRow(t, "known-counterexample", "confirmed", "challenged")
	if err := st.PersistChallengeCampaign(ctx, sampleCampaign(problemID, runID, invID, ch)); err != nil {
		t.Fatalf("persist campaign: %v", err)
	}
	if _, err := st.db.ExecContext(ctx, `UPDATE invariant_challenges SET result_summary='unconfirmed'`); err == nil {
		t.Fatal("expected immutability abort on challenge UPDATE")
	}
	if _, err := st.db.ExecContext(ctx, `DELETE FROM invariant_state_transitions`); err == nil {
		t.Fatal("expected immutability abort on transition DELETE")
	}
}

func TestListInvariantStatesFiltersByState(t *testing.T) {
	st, problemID, runID, invID := persistSampleInvariant(t)
	ctx := context.Background()
	ch := challengeRow(t, "bias-critique", "unconfirmed", "challenged", "surviving")
	ch.ResultSummary = "unconfirmed"
	if err := st.PersistChallengeCampaign(ctx, sampleCampaign(problemID, runID, invID, ch)); err != nil {
		t.Fatalf("persist campaign: %v", err)
	}
	surviving, err := st.ListInvariantStates(ctx, problemID, "surviving")
	if err != nil {
		t.Fatalf("list surviving: %v", err)
	}
	if len(surviving) != 1 || surviving[0].InvariantID != invID {
		t.Fatalf("expected the surviving candidate, got %+v", surviving)
	}
	falsified, err := st.ListInvariantStates(ctx, problemID, "falsified")
	if err != nil {
		t.Fatalf("list falsified: %v", err)
	}
	if len(falsified) != 0 {
		t.Fatalf("expected no falsified candidates, got %+v", falsified)
	}
}

// sampleDerivedChildren builds the derived split-children revision a confirmed
// split challenge carries, keyed to the parent's substrate. Fresh ids each call
// (mimicking the pipeline rebuilding the record on every campaign); the store's
// identity-tuple lookup must reuse the existing revision on replay.
func sampleDerivedChildren(rec InvariantRevisionRecord) *DerivedChildren {
	now := time.Now().UTC()
	return &DerivedChildren{
		Relation: "split",
		Revision: InvariantRevisionRecord{
			ID:              domain.NewInvariantRevisionID(now),
			ProblemID:       rec.ProblemID,
			FailureSpaceID:  rec.FailureSpaceID,
			ClusterRunID:    rec.ClusterRunID,
			RunID:           rec.RunID,
			MinerVersion:    "challenge-split/v1",
			PredicateSchema: "invariant-predicate/v1",
			MinSupport:      rec.MinSupport,
			CandidateCount:  1,
			CreatedAt:       formatTime(now),
			Invocation: InvariantProviderInvocation{
				ID: domain.NewProviderInvocationID(now), RunID: rec.RunID,
				ProviderName: "fixture", SchemaVersion: "invariant-predicate/v1",
				RequestHash: "derived-rh", CreatedAt: formatTime(now),
			},
			Candidates: []CandidateInvariantRow{{
				ID:                   domain.NewCandidateInvariantID(now),
				PredicateFingerprint: "child-fp-1",
				PredicateJSON:        `{"schema":"invariant-predicate/v1"}`,
				Statement:            "split child 0",
				AbstractionLevel:     "mechanism",
				AssociationStatus:    "unknown",
				Ordinal:              0,
			}},
		},
	}
}

// TestReChallengeAfterSplitWeakenIsNotBricked regresses the duplicate-lineage
// abort: a second campaign on a split-weakened invariant re-confirms the same
// split (fresh record ids, same identity tuple). The store must reuse the
// existing derived revision and ignore the duplicate lineage rows instead of
// aborting the whole campaign forever.
func TestReChallengeAfterSplitWeakenIsNotBricked(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()
	rec := sampleRevision(t, st)
	res, err := st.PersistInvariantRevision(ctx, rec)
	if err != nil {
		t.Fatalf("persist revision: %v", err)
	}
	invID := res.Record.Candidates[0].ID

	first := challengeRow(t, "split", "confirmed", "challenged", "weaken")
	first.Derived = sampleDerivedChildren(rec)
	if err := st.PersistChallengeCampaign(ctx, sampleCampaign(rec.ProblemID, rec.RunID, invID, first)); err != nil {
		t.Fatalf("first campaign: %v", err)
	}

	// Re-challenge: weaken -> challenged -> weaken, same split re-confirmed.
	second := challengeRow(t, "split", "confirmed", "challenged", "weaken")
	second.Derived = sampleDerivedChildren(rec) // fresh ids, same identity tuple
	if err := st.PersistChallengeCampaign(ctx, sampleCampaign(rec.ProblemID, rec.RunID, invID, second)); err != nil {
		t.Fatalf("re-challenge must not abort on duplicate lineage: %v", err)
	}

	// Exactly one derived revision and one lineage row survive both campaigns.
	var revCount, lineageCount int
	if err := st.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM invariant_revisions WHERE miner_version = 'challenge-split/v1'`).Scan(&revCount); err != nil {
		t.Fatalf("count revisions: %v", err)
	}
	if revCount != 1 {
		t.Fatalf("derived revision must be reused on replay, got %d", revCount)
	}
	if err := st.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM invariant_lineage WHERE parent_invariant_id = ?`, invID).Scan(&lineageCount); err != nil {
		t.Fatalf("count lineage: %v", err)
	}
	if lineageCount != 1 {
		t.Fatalf("lineage must not duplicate on replay, got %d", lineageCount)
	}
	state, err := st.GetInvariantState(ctx, invID)
	if err != nil {
		t.Fatalf("get state: %v", err)
	}
	if state.State != "weaken" {
		t.Fatalf("state = %q, want weaken", state.State)
	}
}

// TestAbortedCampaignPersistsNoDerivedChildren regresses the atomicity escape:
// derived children ride INSIDE the campaign transaction, so a campaign that
// aborts (illegal transition) must leave no derived revision, no candidates,
// and no lineage — not orphaned provenance-dangling rows.
func TestAbortedCampaignPersistsNoDerivedChildren(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()
	rec := sampleRevision(t, st)
	res, err := st.PersistInvariantRevision(ctx, rec)
	if err != nil {
		t.Fatalf("persist revision: %v", err)
	}
	invID := res.Record.Candidates[0].ID

	bad := challengeRow(t, "split", "confirmed", "surviving") // illegal from proposed
	bad.Derived = sampleDerivedChildren(rec)
	if err := st.PersistChallengeCampaign(ctx, sampleCampaign(rec.ProblemID, rec.RunID, invID, bad)); err == nil {
		t.Fatal("expected illegal-transition abort")
	}

	var revCount, lineageCount int
	if err := st.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM invariant_revisions WHERE miner_version = 'challenge-split/v1'`).Scan(&revCount); err != nil {
		t.Fatalf("count revisions: %v", err)
	}
	if revCount != 0 {
		t.Fatalf("aborted campaign must persist no derived revision, got %d", revCount)
	}
	if err := st.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM invariant_lineage WHERE parent_invariant_id = ?`, invID).Scan(&lineageCount); err != nil {
		t.Fatalf("count lineage: %v", err)
	}
	if lineageCount != 0 {
		t.Fatalf("aborted campaign must persist no lineage, got %d", lineageCount)
	}
}
