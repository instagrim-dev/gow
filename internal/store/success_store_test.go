package store

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/domain"
)

func sampleSuccessRevision(t *testing.T, problemID, runID, cohortHash string) SuccessRevisionRecord {
	t.Helper()
	now := time.Now().UTC()
	return SuccessRevisionRecord{
		ID:                domain.NewSuccessRevisionID(now),
		ProblemID:         problemID,
		RunID:             runID,
		CompressorVersion: "success-compress/v1",
		PredicateSchema:   "invariant-predicate/v1",
		MinSupport:        1,
		CohortHash:        cohortHash,
		InvariantCount:    1,
		CreatedAt:         formatTime(now),
		Invocation: InvariantProviderInvocation{
			ID: domain.NewProviderInvocationID(now), RunID: runID,
			ProviderName: "fixture", SchemaVersion: "invariant-predicate/v1",
			RequestHash: "rh", CreatedAt: formatTime(now),
		},
		Invariants: []SuccessInvariantRow{{
			ID:                   domain.NewSuccessInvariantID(now),
			PredicateFingerprint: "fp-" + cohortHash,
			PredicateJSON:        `{"schema":"invariant-predicate/v1"}`,
			Statement:            "progress retains effective bounds",
			AbstractionLevel:     "mechanism",
			CoverageNum:          1, CoverageDen: 1,
			ExclusionNum: 1, ExclusionDen: 1,
			CoverageOrdinal: "high", ExclusionOrdinal: "high",
			DistinctSupport:       1,
			StrengthDeterministic: 1,
			Ordinal:               0,
		}},
	}
}

func TestPersistSuccessRevisionIdempotentAndRoundTrip(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()
	// Reuse the invariant-layer seeding for a real problem/run FK chain.
	problemID, runID, _, _, _ := seedInvariantPrereqs(t, st)

	rec := sampleSuccessRevision(t, problemID, runID, "cohort-a")
	res, err := st.PersistSuccessRevision(ctx, rec)
	if err != nil {
		t.Fatalf("persist: %v", err)
	}
	if !res.Created || res.Record.Revision != 1 {
		t.Fatalf("expected created revision 1, got created=%v rev=%d", res.Created, res.Record.Revision)
	}
	// Same identity tuple -> existing revision.
	again, err := st.PersistSuccessRevision(ctx, sampleSuccessRevision(t, problemID, runID, "cohort-a"))
	if err != nil {
		t.Fatalf("re-persist: %v", err)
	}
	if again.Created || again.Record.ID != res.Record.ID {
		t.Fatalf("re-persist must return the existing revision; created=%v", again.Created)
	}
	// A changed cohort hash (a new evaluation happened) -> next revision.
	next, err := st.PersistSuccessRevision(ctx, sampleSuccessRevision(t, problemID, runID, "cohort-b"))
	if err != nil {
		t.Fatalf("next persist: %v", err)
	}
	if !next.Created || next.Record.Revision != 2 {
		t.Fatalf("expected new revision 2, got created=%v rev=%d", next.Created, next.Record.Revision)
	}

	got, err := st.GetSuccessRevision(ctx, res.Record.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(got.Invariants) != 1 {
		t.Fatalf("round-trip lost invariants: %+v", got.Invariants)
	}
	si := got.Invariants[0]
	if si.CoverageNum != 1 || si.ExclusionNum != 1 || si.StrengthDeterministic != 1 || si.PredicateJSON == "" {
		t.Fatalf("round-trip lost counts/predicate: %+v", si)
	}

	latest, found, err := st.LatestSuccessRevision(ctx, problemID)
	if err != nil || !found || latest != next.Record.ID {
		t.Fatalf("latest = %q found=%v err=%v, want %q", latest, found, err, next.Record.ID)
	}
	list, err := st.ListSuccessRevisions(ctx, problemID)
	if err != nil || len(list) != 2 {
		t.Fatalf("list = %d revisions err=%v, want 2", len(list), err)
	}
}

func TestSuccessInvariantRowsImmutable(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()
	problemID, runID, _, _, _ := seedInvariantPrereqs(t, st)
	if _, err := st.PersistSuccessRevision(ctx, sampleSuccessRevision(t, problemID, runID, "cohort-imm")); err != nil {
		t.Fatalf("persist: %v", err)
	}
	if _, err := st.db.ExecContext(ctx, `UPDATE success_invariants SET statement='x'`); err == nil || !strings.Contains(err.Error(), "immutable") {
		t.Fatalf("expected immutability abort on UPDATE, got %v", err)
	}
	if _, err := st.db.ExecContext(ctx, `DELETE FROM success_invariant_revisions`); err == nil || !strings.Contains(err.Error(), "immutable") {
		t.Fatalf("expected immutability abort on DELETE, got %v", err)
	}
}

func TestSuccessInvariantRejectsNonProposedState(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()
	seedInvariantPrereqs(t, st)
	_, err := st.db.ExecContext(ctx, `
INSERT INTO success_invariants(id, success_revision_id, predicate_fingerprint, statement, abstraction_level, initial_state, progress_coverage_num, progress_coverage_den, nonprogressor_exclusion_num, nonprogressor_exclusion_den, coverage_ordinal, exclusion_ordinal, distinct_mechanism_support, ordinal)
VALUES('sinv_x','svr_x','fp','s','m','surviving',0,0,0,0,'low','low',0,0)`)
	if err == nil || !strings.Contains(err.Error(), "CHECK") {
		t.Fatalf("expected CHECK rejection of non-proposed state, got %v", err)
	}
}
