package store

import (
	"context"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/domain"
)

func seedProjectionArtifact(t *testing.T, st *Store, problemID, runID, proposalID, contentHash string, withOpenObligation bool) ProjectionRecord {
	t.Helper()
	now := time.Date(2026, 9, 12, 14, 0, 0, 0, time.UTC)
	created := formatTime(now)
	rec := ProjectionRecord{
		Artifact: ProjectionArtifactRow{
			ID:            domain.NewProjectionArtifactID(now),
			ProblemID:     problemID,
			ProposalID:    proposalID,
			RunID:         runID,
			AuthorKind:    "operator",
			SchemaVersion: "projection/v1",
			ContentJSON:   `{"schema":"projection/v1","steps":[{"name":"a","statement":"s"}]}`,
			ContentHash:   contentHash,
			CreatedAt:     created,
		},
		Obligations: []ProjectionObligationRow{{
			ID: domain.NewProjectionObligationID(now), Ordinal: 0, Kind: "steps-compose",
			Statement: "steps compose", Checker: "deterministic-check", CreatedAt: created,
			Decision: &ProjectionObligationDecisionRow{
				RunID: runID, Status: "discharged", DecidedBy: "code",
				Basis: "all steps compose", EvidenceKind: "code-check", EvidenceRef: contentHash,
				CreatedAt: created,
			},
		}},
	}
	if withOpenObligation {
		rec.Obligations = append(rec.Obligations, ProjectionObligationRow{
			ID: domain.NewProjectionObligationID(now.Add(time.Second)), Ordinal: 1,
			Kind: "domain-realization", Statement: "plan is realizable in the domain",
			Checker: "external", CreatedAt: created,
		})
	}
	persisted, err := st.PersistProjection(context.Background(), rec)
	if err != nil {
		t.Fatalf("persist projection: %v", err)
	}
	return persisted
}

// The projection chain round-trips: artifact revision allocation is
// append-only per proposal, obligations keep their synchronous code decision,
// open obligations read back with no decision, and duplicate content per
// proposal is refused.
func TestProjectionRoundTripRevisionsAndDuplicateContent(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()
	problemID, runID, proposalID, _ := seedEvaluatedFailureForAdmission(t, st)

	first := seedProjectionArtifact(t, st, problemID, runID, proposalID, "hash-1", true)
	if first.Artifact.Revision != 1 {
		t.Fatalf("first artifact revision = %d, want 1", first.Artifact.Revision)
	}
	second := seedProjectionArtifact(t, st, problemID, runID, proposalID, "hash-2", false)
	if second.Artifact.Revision != 2 {
		t.Fatalf("second artifact revision = %d, want 2", second.Artifact.Revision)
	}

	recs, err := st.ListProjectionsForProblem(ctx, problemID)
	if err != nil {
		t.Fatalf("list projections: %v", err)
	}
	if len(recs) != 2 {
		t.Fatalf("want 2 artifacts, got %d", len(recs))
	}
	if len(recs[0].Obligations) != 2 {
		t.Fatalf("first artifact must carry 2 obligations, got %d", len(recs[0].Obligations))
	}
	compose := recs[0].Obligations[0]
	if compose.Kind != "steps-compose" || compose.Decision == nil ||
		compose.Decision.Status != "discharged" || compose.Decision.DecidedBy != "code" ||
		compose.Decision.EvidenceKind != "code-check" || compose.Decision.EvidenceRef != "hash-1" {
		t.Fatalf("code decision round-trip lost data: %+v", compose.Decision)
	}
	realize := recs[0].Obligations[1]
	if realize.Kind != "domain-realization" || realize.Checker != "external" || realize.Decision != nil {
		t.Fatalf("open obligation must read back with no decision: %+v", realize)
	}

	// Identical content per proposal is a refused duplicate, not a new revision.
	dup := first
	dup.Artifact.ID = domain.NewProjectionArtifactID(time.Now())
	dup.Obligations = nil
	if _, err := st.PersistProjection(ctx, dup); err == nil {
		t.Fatal("duplicate artifact content per proposal must be refused")
	}
}

// Obligation decisions are append-once and the chain rows are immutable at
// the SQL layer.
func TestProjectionDecisionAppendOnceAndImmutability(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()
	problemID, runID, proposalID, _ := seedEvaluatedFailureForAdmission(t, st)
	rec := seedProjectionArtifact(t, st, problemID, runID, proposalID, "hash-1", true)
	open := rec.Obligations[1]

	decision := ProjectionObligationDecisionRow{
		ObligationID: open.ID, RunID: runID, Status: "failed", DecidedBy: "operator",
		Basis: "external run refuted step 2", EvidenceKind: "evaluation", EvidenceRef: "evl_x",
		CreatedAt: formatTime(time.Date(2026, 9, 12, 15, 0, 0, 0, time.UTC)),
	}
	if err := st.PersistObligationDecision(ctx, decision); err != nil {
		t.Fatalf("persist decision: %v", err)
	}
	if err := st.PersistObligationDecision(ctx, decision); err == nil {
		t.Fatal("re-deciding a decided obligation must be refused (append-once)")
	}

	ob, art, err := st.GetProjectionObligation(ctx, open.ID)
	if err != nil {
		t.Fatalf("get obligation: %v", err)
	}
	if art.ProposalID != proposalID || ob.Decision == nil || ob.Decision.Status != "failed" ||
		ob.Decision.EvidenceRef != "evl_x" {
		t.Fatalf("obligation/decision round-trip lost data: ob=%+v art=%+v", ob, art)
	}

	if _, err := st.db.ExecContext(ctx, `UPDATE projection_artifacts SET content_json = '{}' WHERE id = ?`, rec.Artifact.ID); err == nil {
		t.Fatal("projection artifacts must be immutable")
	}
	if _, err := st.db.ExecContext(ctx, `UPDATE projection_obligation_decisions SET status = 'discharged' WHERE obligation_id = ?`, open.ID); err == nil {
		t.Fatal("obligation decisions must be immutable")
	}
	if _, err := st.db.ExecContext(ctx, `DELETE FROM projection_obligations WHERE id = ?`, open.ID); err == nil {
		t.Fatal("projection obligations must be immutable")
	}
}
