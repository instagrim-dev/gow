package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/domain"
)

// Guard-mutation block, store layer (follow-on P1; newf-regress `mutants`
// contract). Each mutant runs the exact PRE-FIX query text the guard replaced
// against the same fixture, and requires that the NAMED semantic assertion
// fails against the mutant's well-formed result — while the production guard
// is asserted alongside, proving the regression detects the disabled guard.

// mutant use_origin_break_flag: disable revision-bound break admission by
// joining the origin-time frontier_target_invariants flags (the pre-v26
// admission source). must_fail = [NO_ORIGIN_FLAG_REUSE]: the origin violated
// flag admits a proposal whose assessed-revision break is unknown.
func TestGuardMutationUseOriginBreakFlag(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()

	rec := sampleFrontier(t, st) // origin target row: violates / violated=1
	res, err := st.PersistFrontierGeneration(ctx, rec)
	if err != nil {
		t.Fatalf("persist frontier: %v", err)
	}
	problemID := res.Record.ProblemID
	proposalID := res.Record.Proposals[0].ID
	targetInvID := res.Record.Proposals[0].Targets[0].InvariantID
	insertTestSignatureContent(t, st, proposalID, `{"schema_version":"mechanism/v1"}`, "cfp-gm1")

	now := time.Now().UTC()
	run := EvaluationRunRecord{
		ID: domain.NewEvaluationRunID(now), ProblemID: problemID, RunID: res.Record.RunID,
		Mode: "proposal", CreatedAt: formatTime(now),
		Evaluations: []EvaluationRow{{
			ID: domain.NewEvaluationID(now), ProposalID: proposalID, Verdict: "partial_success",
			VerifierKind: "model-judgment", VerificationStrength: "single-model-judgment",
			ConfidenceOrdinal: "medium", ToolName: "m", ToolVersion: "v1",
			// The assessed revision's break verdict is UNKNOWN.
			TargetVerdicts: []EvaluationTargetVerdictRow{{InvariantID: targetInvID, Verdict: "unknown", Violated: false}},
		}},
	}
	if _, err := st.PersistEvaluationRun(ctx, run); err != nil {
		t.Fatalf("persist evaluation: %v", err)
	}

	// GUARD ON: assessment-context admission excludes the proposal.
	rows, err := st.ListBreakCohortRows(ctx, problemID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(rows) != 0 {
		t.Fatalf("guard on: NO_ORIGIN_FLAG_REUSE must hold, got %+v", rows)
	}

	// MUTANT: the pre-v26 admission source. The named assertion fails: the
	// origin flag admits the proposal despite the unknown assessed break.
	var admitted int
	if err := st.db.QueryRowContext(ctx, `
SELECT COUNT(*)
FROM frontier_target_invariants t
JOIN frontier_proposals p ON p.id = t.proposal_id
JOIN evaluations e ON e.proposal_id = p.id
WHERE p.problem_id = ? AND t.violated = 1
`, problemID).Scan(&admitted); err != nil {
		t.Fatalf("mutant query: %v", err)
	}
	if admitted == 0 {
		t.Fatal("mutant did not disable the guard: NO_ORIGIN_FLAG_REUSE unexpectedly holds under the origin-flag join")
	}
}

// mutant use_max_revision_current_view: disable the latest-emitted-occurrence
// current view by computing "current" as MAX(revision) — the pre-v28 CTE. On
// an A -> B -> A replay the named assertion (assessed == current) fails: the
// mutant reports B as current even though the latest emitted occurrence is A.
func TestGuardMutationUseMaxRevisionCurrentView(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()

	contentA := `{"schema_version":"mechanism/v1","rev":"A"}`
	contentB := `{"schema_version":"mechanism/v1","rev":"B"}`
	sumA := sha256.Sum256([]byte(contentA))
	hashA := hex.EncodeToString(sumA[:])
	sumB := sha256.Sum256([]byte(contentB))
	hashB := hex.EncodeToString(sumB[:])

	gen1 := sampleFrontier(t, st)
	gen1.Proposals[0].SignatureJSON = contentA
	gen1.Proposals[0].CanonicalFingerprint = "cfp-gm2"
	res1, err := st.PersistFrontierGeneration(ctx, gen1)
	if err != nil {
		t.Fatalf("gen1: %v", err)
	}
	problemID := res1.Record.ProblemID
	proposalID := res1.Record.Proposals[0].ID
	targetInvID := res1.Record.Proposals[0].Targets[0].InvariantID

	gen2 := sampleFrontierReusingProblem(t, st, gen1)
	gen2.Proposals[0].SignatureJSON = contentB
	if _, err := st.PersistFrontierGeneration(ctx, gen2); err != nil {
		t.Fatalf("gen2: %v", err)
	}
	gen3 := sampleFrontierReusingProblem(t, st, gen2)
	gen3.Proposals[0].SignatureJSON = contentA // A re-emitted: A is CURRENT
	if _, err := st.PersistFrontierGeneration(ctx, gen3); err != nil {
		t.Fatalf("gen3: %v", err)
	}

	// A completed assessment of the re-emitted A occurrence.
	now := time.Now().UTC()
	run := EvaluationRunRecord{
		ID: domain.NewEvaluationRunID(now), ProblemID: problemID, RunID: res1.Record.RunID,
		Mode: "proposal", CreatedAt: formatTime(now),
		Evaluations: []EvaluationRow{{
			ID: domain.NewEvaluationID(now), ProposalID: proposalID, Verdict: "partial_success",
			VerifierKind: "model-judgment", VerificationStrength: "single-model-judgment",
			ConfidenceOrdinal: "medium", ToolName: "m", ToolVersion: "v1",
			SignatureContentHash: hashA,
			TargetVerdicts:       []EvaluationTargetVerdictRow{{InvariantID: targetInvID, Verdict: "violates", Violated: true}},
		}},
	}
	if _, err := st.PersistEvaluationRun(ctx, run); err != nil {
		t.Fatalf("persist evaluation: %v", err)
	}

	// GUARD ON: current view = latest emitted occurrence = A; not pending.
	rows, err := st.ListBreakCohortRows(ctx, problemID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(rows) != 1 || rows[0].ContentHash != hashA || rows[0].LatestContentHash != hashA {
		t.Fatalf("guard on: assessed==current must hold for the re-emitted A: %+v", rows)
	}

	// MUTANT: the pre-v28 current view (MAX revision) reports B as current.
	// The named assertion (assessed == current) fails for the semantic reason:
	// the assessed re-emitted occurrence would be wrongly pending.
	var mutantCurrent string
	if err := st.db.QueryRowContext(ctx, `
SELECT r.content_hash
FROM frontier_proposal_signature_revisions r
JOIN (SELECT proposal_id, MAX(revision) AS mr FROM frontier_proposal_signature_revisions GROUP BY proposal_id) lr
  ON lr.proposal_id = r.proposal_id AND lr.mr = r.revision
WHERE r.proposal_id = ?
`, proposalID).Scan(&mutantCurrent); err != nil {
		t.Fatalf("mutant query: %v", err)
	}
	if mutantCurrent != hashB {
		t.Fatalf("mutant must report B as current, got %s", mutantCurrent)
	}
	if mutantCurrent == hashA {
		t.Fatal("mutant did not disable the guard: assessed==current unexpectedly holds")
	}
}

// mutant use_latest_revision_not_selection: disable the compression selection
// log by reading MAX(revision) (the pre-v28 consumer, LatestSuccessRevision).
// After a retraction reuses the earlier empty artifact, the named assertion
// (current guidance == latest SELECTION) fails: the mutant returns the
// higher-numbered stale-supported revision.
func TestGuardMutationUseLatestRevisionNotSelection(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()
	problemID, runID, _, _, _ := seedInvariantPrereqs(t, st)

	// R1: empty cohort. R2: supported cohort.
	r1 := sampleSuccessRevision(t, problemID, runID, "cohort-empty")
	r1.InvariantCount = 0
	r1.Invariants = nil
	res1, err := st.PersistSuccessRevision(ctx, r1)
	if err != nil {
		t.Fatalf("persist R1: %v", err)
	}
	r2 := sampleSuccessRevision(t, problemID, runID, "cohort-supported")
	if _, err := st.PersistSuccessRevision(ctx, r2); err != nil {
		t.Fatalf("persist R2: %v", err)
	}

	// Retraction: a NEW execution reuses R1's identity (fresh run).
	now := time.Now().UTC()
	run3 := domain.NewRun{
		ID: domain.NewRunID(now), ProblemID: problemID, Operation: "successes compress",
		Status: domain.RunStatusInitialized, InputRef: "test", ToolName: "newf", ToolVersion: "test",
		StartedAt: now, CompletedAt: now,
	}
	if _, err := st.CreateRun(ctx, run3); err != nil {
		t.Fatalf("create run3: %v", err)
	}
	r3 := sampleSuccessRevision(t, problemID, run3.ID, "cohort-empty")
	r3.InvariantCount = 0
	r3.Invariants = nil
	res3, err := st.PersistSuccessRevision(ctx, r3)
	if err != nil {
		t.Fatalf("persist R3 (reuse): %v", err)
	}
	if res3.Created || res3.Record.ID != res1.Record.ID {
		t.Fatalf("retraction must reuse R1: %+v", res3.Record.ID)
	}

	// GUARD ON: the latest SELECTION is R1 (the reused empty artifact).
	selected, found, err := st.LatestSelectedSuccessRevision(ctx, problemID)
	if err != nil || !found {
		t.Fatalf("selected: %v %v", err, found)
	}
	if selected != res1.Record.ID {
		t.Fatalf("guard on: current guidance must be the reused R1, got %s", selected)
	}

	// MUTANT: the pre-v28 consumer. The named assertion fails: it returns the
	// higher-numbered stale-supported R2, resurrecting retracted support.
	mutant, found, err := st.LatestSuccessRevision(ctx, problemID)
	if err != nil || !found {
		t.Fatalf("mutant: %v %v", err, found)
	}
	if mutant == res1.Record.ID {
		t.Fatal("mutant did not disable the guard: selection unexpectedly holds under MAX(revision)")
	}
	if mutant != r2.ID {
		t.Fatalf("mutant must return the stale-supported R2, got %s", mutant)
	}
}
