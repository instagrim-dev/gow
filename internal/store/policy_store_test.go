package store

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/domain"
)

func samplePolicyRevision(t *testing.T, st *Store) PolicyRevisionRecord {
	t.Helper()
	problemID, runID, _, _, invariantID := seedFrontierPrereqs(t, st)
	now := time.Now().UTC()
	return PolicyRevisionRecord{
		ID:                 domain.NewSearchPolicyRevisionID(now),
		ProblemID:          problemID,
		RunID:              runID,
		MutatorVersion:     "policy-mutate/v1",
		PolicySchema:       "policy/v1",
		EvidenceCohortHash: "cohort-abc",
		DirectiveCount:     1,
		CreatedAt:          formatTime(now),
		Directives: []PolicyDirectiveRow{{
			ID:              domain.NewSearchPolicyDirectiveID(now),
			Kind:            "avoid",
			TargetKind:      "surviving_invariant",
			TargetID:        invariantID,
			Weight:          string(domain.OrdinalHigh),
			EpistemicSource: "operator_attested",
			Ordinal:         0,
			Provenance: []PolicyProvenanceRow{
				{EvidenceKind: "surviving_invariant", EvidenceRef: invariantID},
			},
		}},
	}
}

func TestPersistPolicyRevisionRoundTripAndIdempotent(t *testing.T) {
	t.Parallel()
	st := openTestStore(t)
	defer st.Close()
	ctx := context.Background()

	rec := samplePolicyRevision(t, st)
	got, created, err := st.PersistPolicyRevision(ctx, rec)
	if err != nil {
		t.Fatalf("PersistPolicyRevision() error = %v", err)
	}
	if !created {
		t.Fatal("first persist should be created=true")
	}
	if got.Revision != 1 {
		t.Fatalf("revision = %d, want 1", got.Revision)
	}
	if len(got.Directives) != 1 || got.Directives[0].TargetKind != "surviving_invariant" {
		t.Fatalf("directives not round-tripped: %+v", got.Directives)
	}
	if len(got.Directives[0].Provenance) != 1 {
		t.Fatalf("provenance not round-tripped: %+v", got.Directives[0].Provenance)
	}

	// Same identity tuple ⇒ idempotent, not a new revision.
	rec2 := rec
	rec2.ID = domain.NewSearchPolicyRevisionID(time.Now().UTC())
	again, created2, err := st.PersistPolicyRevision(ctx, rec2)
	if err != nil {
		t.Fatalf("second PersistPolicyRevision() error = %v", err)
	}
	if created2 {
		t.Fatal("re-persisting an unchanged cohort must be idempotent (created=false)")
	}
	if again.ID != got.ID {
		t.Fatalf("idempotent persist returned a different id: %s vs %s", again.ID, got.ID)
	}
}

func TestPersistPolicyRevisionNewCohortNextRevision(t *testing.T) {
	t.Parallel()
	st := openTestStore(t)
	defer st.Close()
	ctx := context.Background()

	rec := samplePolicyRevision(t, st)
	if _, _, err := st.PersistPolicyRevision(ctx, rec); err != nil {
		t.Fatalf("first persist: %v", err)
	}
	// New evidence ⇒ new cohort hash ⇒ next revision.
	rec2 := rec
	rec2.ID = domain.NewSearchPolicyRevisionID(time.Now().UTC().Add(time.Second))
	rec2.EvidenceCohortHash = "cohort-xyz"
	rec2.Directives[0].ID = domain.NewSearchPolicyDirectiveID(time.Now().UTC().Add(time.Second))
	got2, created2, err := st.PersistPolicyRevision(ctx, rec2)
	if err != nil {
		t.Fatalf("second persist: %v", err)
	}
	if !created2 || got2.Revision != 2 {
		t.Fatalf("new cohort must yield revision 2 (created), got created=%v rev=%d", created2, got2.Revision)
	}

	latest, found, err := st.LatestPolicyRevision(ctx, rec.ProblemID)
	if err != nil || !found {
		t.Fatalf("LatestPolicyRevision error=%v found=%v", err, found)
	}
	if latest != got2.ID {
		t.Fatalf("latest = %s, want %s", latest, got2.ID)
	}
}

func TestPolicyRevisionImmutable(t *testing.T) {
	t.Parallel()
	st := openTestStore(t)
	defer st.Close()
	ctx := context.Background()

	rec := samplePolicyRevision(t, st)
	got, _, err := st.PersistPolicyRevision(ctx, rec)
	if err != nil {
		t.Fatalf("persist: %v", err)
	}
	_, err = st.db.ExecContext(ctx, `UPDATE search_policy_revisions SET directive_count = 99 WHERE id = ?`, got.ID)
	if err == nil || !strings.Contains(err.Error(), "immutable") {
		t.Fatalf("expected immutability trigger to reject update, got %v", err)
	}
	_, err = st.db.ExecContext(ctx, `DELETE FROM search_policy_directives WHERE id = ?`, got.Directives[0].ID)
	if err == nil || !strings.Contains(err.Error(), "immutable") {
		t.Fatalf("expected immutability trigger to reject directive delete, got %v", err)
	}
}

func TestPolicyDirectiveKindCheckConstraint(t *testing.T) {
	t.Parallel()
	st := openTestStore(t)
	defer st.Close()
	ctx := context.Background()

	rec := samplePolicyRevision(t, st)
	rec.Directives[0].Kind = "bogus"
	if _, _, err := st.PersistPolicyRevision(ctx, rec); err == nil {
		t.Fatal("expected CHECK constraint to reject an invalid directive kind")
	}
}

func TestFrontierGenerationPolicyLogRoundTrip(t *testing.T) {
	t.Parallel()
	st := openTestStore(t)
	defer st.Close()
	ctx := context.Background()

	// Persist a generation (for the FK) and a policy revision, then the bias log.
	gen := sampleFrontier(t, st)
	genRes, err := st.PersistFrontierGeneration(ctx, gen)
	if err != nil {
		t.Fatalf("persist frontier generation: %v", err)
	}
	proposalID := genRes.Record.Proposals[0].ID

	rec := PolicyRevisionRecord{
		ID:                 domain.NewSearchPolicyRevisionID(time.Now().UTC()),
		ProblemID:          gen.ProblemID,
		RunID:              gen.RunID,
		MutatorVersion:     "policy-mutate/v1",
		PolicySchema:       "policy/v1",
		EvidenceCohortHash: "cohort-log",
		CreatedAt:          formatTime(time.Now().UTC()),
	}
	polRes, _, err := st.PersistPolicyRevision(ctx, rec)
	if err != nil {
		t.Fatalf("persist policy revision: %v", err)
	}

	rows := []FrontierGenerationPolicyRow{{ProposalID: proposalID, NetBias: -2, Avoided: true, FloorProtected: true}}
	if err := st.PersistFrontierGenerationPolicy(ctx, genRes.Record.ID, polRes.ID, rows); err != nil {
		t.Fatalf("persist generation policy: %v", err)
	}

	revID, got, err := st.GetFrontierGenerationPolicy(ctx, genRes.Record.ID)
	if err != nil {
		t.Fatalf("get generation policy: %v", err)
	}
	if revID != polRes.ID {
		t.Fatalf("policy revision id = %s, want %s", revID, polRes.ID)
	}
	if len(got) != 1 || got[0].NetBias != -2 || !got[0].Avoided || !got[0].FloorProtected {
		t.Fatalf("bias log not round-tripped: %+v", got)
	}
}
