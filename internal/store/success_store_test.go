package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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

// insertEvalTargetVerdict writes the assessment-context break verdict (v26)
// that cohort admission consumes — used where tests insert evaluations
// directly instead of going through the pipeline.
func insertEvalTargetVerdict(t *testing.T, st *Store, evalID, invID, verdict string, violated bool) {
	t.Helper()
	if _, err := st.db.ExecContext(context.Background(), `
INSERT INTO evaluation_target_verdicts(evaluation_id, invariant_id, verdict, violated) VALUES(?, ?, ?, ?)`,
		evalID, invID, verdict, boolToInt(violated)); err != nil {
		t.Fatalf("insert eval target verdict: %v", err)
	}
}

// H1 regression: ListBreakCohortRows must return a COHERENT (verdict, strength,
// evaluation_id) triple from ONE evaluation. A re-evaluation must not let the
// cohort query splice the first (sticky) verdict onto a later evaluation's
// strength. We persist E1 (partial_success / single-model-judgment) which sets
// the sticky result, then insert a later E2 (failure / deterministic). The
// cohort row must reflect E1 entirely — never partial_success + deterministic.
func TestListBreakCohortRowsSelectsOneEvaluationRecord(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()

	rec := sampleFrontier(t, st)
	res, err := st.PersistFrontierGeneration(ctx, rec)
	if err != nil {
		t.Fatalf("persist frontier: %v", err)
	}
	problemID := res.Record.ProblemID
	proposalID := res.Record.Proposals[0].ID
	targetInvID := res.Record.Proposals[0].Targets[0].InvariantID

	// Persist canonical content so the proposal is cohort-eligible.
	insertTestSignatureContent(t, st, proposalID, `{"schema_version":"mechanism/v1"}`, "cfp-1")

	// E1: the earliest evaluation. partial_success / single-model-judgment. This
	// sets the sticky frontier_proposals.result via PersistEvaluationRun.
	now := time.Now().UTC()
	e1ID := domain.NewEvaluationID(now)
	run := EvaluationRunRecord{
		ID:        domain.NewEvaluationRunID(now),
		ProblemID: problemID,
		RunID:     res.Record.RunID,
		Mode:      "proposal",
		CreatedAt: formatTime(now.Add(-time.Hour)), // earlier
		Evaluations: []EvaluationRow{{ID: e1ID, ProposalID: proposalID, Verdict: "partial_success", VerifierKind: "model-judgment", VerificationStrength: "single-model-judgment", ConfidenceOrdinal: "medium", ToolName: "m", ToolVersion: "v1",
			TargetVerdicts: []EvaluationTargetVerdictRow{{InvariantID: targetInvID, Verdict: "violates", Violated: true}}}},
	}
	if _, err := st.PersistEvaluationRun(ctx, run); err != nil {
		t.Fatalf("persist E1: %v", err)
	}

	// E2: a LATER re-evaluation. failure / deterministic. Inserted directly with a
	// strictly later created_at into the same run (a real re-evaluation would be a
	// new run; the direct insert keeps the fixture small).
	e2ID := domain.NewEvaluationID(now.Add(time.Second))
	if _, err := st.db.ExecContext(ctx, `
INSERT INTO evaluations(id, evaluation_run_id, proposal_id, verdict, verifier_kind, verification_strength, confidence_ordinal, tool_name, tool_version, created_at)
VALUES(?, ?, ?, 'failure', 'deterministic-check', 'deterministic', 'high', 'd', 'v1', ?)`,
		e2ID, run.ID, proposalID, formatTime(now)); err != nil {
		t.Fatalf("insert E2: %v", err)
	}
	insertEvalTargetVerdict(t, st, e2ID, targetInvID, "violates", true)

	rows, err := st.ListBreakCohortRows(ctx, problemID)
	if err != nil {
		t.Fatalf("list cohort rows: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("want 1 cohort row, got %d", len(rows))
	}
	r := rows[0]
	// selection-policy/v1: the DETERMINISTIC re-evaluation (E2) displaces the
	// earlier model judgment — a refuted success cannot remain active support.
	// The triple must still be coherent from ONE record (H1): NEVER the first
	// verdict spliced onto a later evaluation's strength.
	if r.Result != "failure" || r.Strength != "deterministic" || r.EvaluationID != e2ID {
		t.Fatalf("cohort row must be the coherent SELECTED triple; got result=%q strength=%q eval=%q (want failure/deterministic/%s)", r.Result, r.Strength, r.EvaluationID, e2ID)
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

// F2 probe #1: verification_blocked followed by an equal-strength successful
// reassessment. selection-policy/v1 ties break to the LATEST, so the accepted
// reassessment is selected — a blocked first attempt is not pinned forever.
func TestListBreakCohortRowsLatestWinsAtEqualStrength(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()

	rec := sampleFrontier(t, st)
	res, err := st.PersistFrontierGeneration(ctx, rec)
	if err != nil {
		t.Fatalf("persist frontier: %v", err)
	}
	problemID := res.Record.ProblemID
	proposalID := res.Record.Proposals[0].ID
	targetInvID := res.Record.Proposals[0].Targets[0].InvariantID
	insertTestSignatureContent(t, st, proposalID, `{"schema_version":"mechanism/v1"}`, "cfp-1")

	now := time.Now().UTC()
	blockedID := domain.NewEvaluationID(now)
	run := EvaluationRunRecord{
		ID:        domain.NewEvaluationRunID(now),
		ProblemID: problemID,
		RunID:     res.Record.RunID,
		Mode:      "proposal",
		CreatedAt: formatTime(now.Add(-time.Hour)),
		Evaluations: []EvaluationRow{{ID: blockedID, ProposalID: proposalID, Verdict: "verification_blocked", VerifierKind: "model-judgment", VerificationStrength: "single-model-judgment", ConfidenceOrdinal: "low", ToolName: "m", ToolVersion: "v1",
			TargetVerdicts: []EvaluationTargetVerdictRow{{InvariantID: targetInvID, Verdict: "violates", Violated: true}}}},
	}
	if _, err := st.PersistEvaluationRun(ctx, run); err != nil {
		t.Fatalf("persist blocked: %v", err)
	}
	successID := domain.NewEvaluationID(now.Add(time.Second))
	if _, err := st.db.ExecContext(ctx, `
INSERT INTO evaluations(id, evaluation_run_id, proposal_id, verdict, verifier_kind, verification_strength, confidence_ordinal, tool_name, tool_version, created_at)
VALUES(?, ?, ?, 'partial_success', 'model-judgment', 'single-model-judgment', 'medium', 'm', 'v1', ?)`,
		successID, run.ID, proposalID, formatTime(now)); err != nil {
		t.Fatalf("insert reassessment: %v", err)
	}
	insertEvalTargetVerdict(t, st, successID, targetInvID, "violates", true)

	rows, err := st.ListBreakCohortRows(ctx, problemID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("want 1 row, got %d", len(rows))
	}
	if rows[0].EvaluationID != successID || rows[0].Result != "partial_success" {
		t.Fatalf("latest equal-strength reassessment must be selected; got %+v", rows[0])
	}
}

// Round-2 F2 regression: an unassessed NEWER revision must not inherit the old
// evaluation's outcome. The cohort row binds the evaluation to the revision it
// ACTUALLY assessed (R1) and exposes the latest hash (R2) so the caller can
// mark the member pending instead of splicing.
func TestListBreakCohortRowsBindsAssessedRevision(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()

	rec := sampleFrontier(t, st)
	res, err := st.PersistFrontierGeneration(ctx, rec)
	if err != nil {
		t.Fatalf("persist frontier: %v", err)
	}
	problemID := res.Record.ProblemID
	proposalID := res.Record.Proposals[0].ID
	targetInvID := res.Record.Proposals[0].Targets[0].InvariantID

	r1JSON := `{"schema_version":"mechanism/v1","completeness":"complete"}`
	insertTestSignatureContent(t, st, proposalID, r1JSON, "cfp-r1")
	sumR1 := sha256.Sum256([]byte(r1JSON))
	r1Hash := hex.EncodeToString(sumR1[:])

	// E1 assessed R1 (recorded hash).
	now := time.Now().UTC()
	e1ID := domain.NewEvaluationID(now)
	run := EvaluationRunRecord{
		ID:        domain.NewEvaluationRunID(now),
		ProblemID: problemID,
		RunID:     res.Record.RunID,
		Mode:      "proposal",
		CreatedAt: formatTime(now),
		Evaluations: []EvaluationRow{{
			ID: e1ID, ProposalID: proposalID, Verdict: "partial_success",
			VerifierKind: "model-judgment", VerificationStrength: "single-model-judgment",
			ConfidenceOrdinal: "medium", ToolName: "m", ToolVersion: "v1",
			SignatureContentHash: r1Hash,
			TargetVerdicts:       []EvaluationTargetVerdictRow{{InvariantID: targetInvID, Verdict: "violates", Violated: true}},
		}},
	}
	if _, err := st.PersistEvaluationRun(ctx, run); err != nil {
		t.Fatalf("persist E1: %v", err)
	}

	// R2: a newer, UNASSESSED interpretation.
	r2JSON := `{"schema_version":"mechanism/v1","completeness":"unobserved"}`
	insertTestSignatureContent(t, st, proposalID, r2JSON, "cfp-r2")
	sumR2 := sha256.Sum256([]byte(r2JSON))
	r2Hash := hex.EncodeToString(sumR2[:])

	rows, err := st.ListBreakCohortRows(ctx, problemID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("want 1 row, got %d", len(rows))
	}
	r := rows[0]
	if r.ContentHash != r1Hash {
		t.Fatalf("row must bind the ASSESSED revision R1; got %s", r.ContentHash)
	}
	if r.LatestContentHash != r2Hash {
		t.Fatalf("latest hash must expose R2 for pending detection; got %s", r.LatestContentHash)
	}
	if r.SignatureJSON != r1JSON {
		t.Fatalf("content must be the assessed bytes (R1), got %s", r.SignatureJSON)
	}
}

// Round-2 F4 regression: a non-decisive blocker must not outrank a later
// completed reassessment. Decisiveness gates eligibility BEFORE the strength
// hierarchy; the control (decisive deterministic failure vs later weaker model
// success) is covered by TestListBreakCohortRowsSelectsOneEvaluationRecord.
func TestListBreakCohortRowsDecisiveBeatsBlocked(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()

	rec := sampleFrontier(t, st)
	res, err := st.PersistFrontierGeneration(ctx, rec)
	if err != nil {
		t.Fatalf("persist frontier: %v", err)
	}
	problemID := res.Record.ProblemID
	proposalID := res.Record.Proposals[0].ID
	targetInvID := res.Record.Proposals[0].Targets[0].InvariantID
	insertTestSignatureContent(t, st, proposalID, `{"schema_version":"mechanism/v1"}`, "cfp-blocked")

	// Earlier: verification_blocked at DETERMINISTIC strength (the pipeline
	// emits exactly this for stale targets).
	now := time.Now().UTC()
	blockedID := domain.NewEvaluationID(now)
	run := EvaluationRunRecord{
		ID:        domain.NewEvaluationRunID(now),
		ProblemID: problemID,
		RunID:     res.Record.RunID,
		Mode:      "proposal",
		CreatedAt: formatTime(now.Add(-time.Hour)),
		Evaluations: []EvaluationRow{{
			ID: blockedID, ProposalID: proposalID, Verdict: "verification_blocked",
			VerifierKind: "deterministic-check", VerificationStrength: "deterministic",
			ConfidenceOrdinal: "high", ToolName: "d", ToolVersion: "v1",
			TargetVerdicts: []EvaluationTargetVerdictRow{{InvariantID: targetInvID, Verdict: "violates", Violated: true}},
		}},
	}
	if _, err := st.PersistEvaluationRun(ctx, run); err != nil {
		t.Fatalf("persist blocked: %v", err)
	}
	// Later: a COMPLETED reassessment at weaker strength.
	successID := domain.NewEvaluationID(now.Add(time.Second))
	if _, err := st.db.ExecContext(ctx, `
INSERT INTO evaluations(id, evaluation_run_id, proposal_id, verdict, verifier_kind, verification_strength, confidence_ordinal, tool_name, tool_version, created_at)
VALUES(?, ?, ?, 'partial_success', 'model-judgment', 'single-model-judgment', 'medium', 'm', 'v1', ?)`,
		successID, run.ID, proposalID, formatTime(now)); err != nil {
		t.Fatalf("insert reassessment: %v", err)
	}
	insertEvalTargetVerdict(t, st, successID, targetInvID, "violates", true)

	rows, err := st.ListBreakCohortRows(ctx, problemID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("want 1 row, got %d", len(rows))
	}
	if rows[0].EvaluationID != successID || rows[0].Result != "partial_success" {
		t.Fatalf("a completed reassessment must displace a non-decisive blocker; got %+v", rows[0])
	}
}

// v26 finding-3 regression, direction 1 (violates -> unknown): the ORIGIN
// frontier_target_invariants row says violated, but the selected evaluation's
// own recomputed break verdict for its assessed revision is unknown. The
// cohort must EXCLUDE the proposal — an origin-time break flag must not ride
// revised content.
func TestListBreakCohortRowsExcludesWhenAssessedBreakDegraded(t *testing.T) {
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
	insertTestSignatureContent(t, st, proposalID, `{"schema_version":"mechanism/v1"}`, "cfp-degraded")

	now := time.Now().UTC()
	run := EvaluationRunRecord{
		ID:        domain.NewEvaluationRunID(now),
		ProblemID: problemID,
		RunID:     res.Record.RunID,
		Mode:      "proposal",
		CreatedAt: formatTime(now),
		Evaluations: []EvaluationRow{{
			ID: domain.NewEvaluationID(now), ProposalID: proposalID, Verdict: "partial_success",
			VerifierKind: "model-judgment", VerificationStrength: "single-model-judgment",
			ConfidenceOrdinal: "medium", ToolName: "m", ToolVersion: "v1",
			TargetVerdicts: []EvaluationTargetVerdictRow{{InvariantID: targetInvID, Verdict: "unknown", Violated: false}},
		}},
	}
	if _, err := st.PersistEvaluationRun(ctx, run); err != nil {
		t.Fatalf("persist evaluation: %v", err)
	}

	rows, err := st.ListBreakCohortRows(ctx, problemID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(rows) != 0 {
		t.Fatalf("a degraded (unknown) assessed break must be EXCLUDED despite the origin violated flag; got %+v", rows)
	}
}

// v26 finding-3 regression, direction 2 (unknown -> violates): the ORIGIN row
// says unknown/not-violated, but the selected evaluation VERIFIED the break
// for its assessed revision. The cohort must ADMIT the proposal on the
// assessment context, not exclude it on the stale origin flag.
func TestListBreakCohortRowsAdmitsWhenAssessedBreakVerified(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()

	rec := sampleFrontier(t, st)
	targetInvID := rec.Proposals[0].Targets[0].InvariantID
	rec.Proposals[0].Targets[0] = FrontierTargetRow{InvariantID: targetInvID, Verdict: "unknown", Violated: false}
	rec.Proposals[0].ViolatesAnyTarget = false
	res, err := st.PersistFrontierGeneration(ctx, rec)
	if err != nil {
		t.Fatalf("persist frontier: %v", err)
	}
	problemID := res.Record.ProblemID
	proposalID := res.Record.Proposals[0].ID
	insertTestSignatureContent(t, st, proposalID, `{"schema_version":"mechanism/v1"}`, "cfp-verified")

	now := time.Now().UTC()
	evalID := domain.NewEvaluationID(now)
	run := EvaluationRunRecord{
		ID:        domain.NewEvaluationRunID(now),
		ProblemID: problemID,
		RunID:     res.Record.RunID,
		Mode:      "proposal",
		CreatedAt: formatTime(now),
		Evaluations: []EvaluationRow{{
			ID: evalID, ProposalID: proposalID, Verdict: "partial_success",
			VerifierKind: "model-judgment", VerificationStrength: "single-model-judgment",
			ConfidenceOrdinal: "medium", ToolName: "m", ToolVersion: "v1",
			TargetVerdicts: []EvaluationTargetVerdictRow{{InvariantID: targetInvID, Verdict: "violates", Violated: true}},
		}},
	}
	if _, err := st.PersistEvaluationRun(ctx, run); err != nil {
		t.Fatalf("persist evaluation: %v", err)
	}

	rows, err := st.ListBreakCohortRows(ctx, problemID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(rows) != 1 || rows[0].EvaluationID != evalID || rows[0].TargetInvariantID != targetInvID {
		t.Fatalf("a verified assessed break must be ADMITTED despite the stale origin flag; got %+v", rows)
	}
}

// selection-policy/v3 regression: content compatibility outranks the strength
// hierarchy. A DETERMINISTIC, decisive evaluation of STALE content must not
// shadow a weaker (model-judgment) but CURRENT-revision reassessment — the
// current one is selected, its bytes are the cohort content, and the member is
// not pending (assessed == latest).
func TestListBreakCohortRowsCurrentContentOutranksStrongerStale(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()

	rec := sampleFrontier(t, st)
	res, err := st.PersistFrontierGeneration(ctx, rec)
	if err != nil {
		t.Fatalf("persist frontier: %v", err)
	}
	problemID := res.Record.ProblemID
	proposalID := res.Record.Proposals[0].ID
	targetInvID := res.Record.Proposals[0].Targets[0].InvariantID

	// Revision A (stale after B lands) and revision B (current).
	aJSON := `{"schema_version":"mechanism/v1","rev":"A"}`
	insertTestSignatureContent(t, st, proposalID, aJSON, "cfp-a")
	sumA := sha256.Sum256([]byte(aJSON))
	hashA := hex.EncodeToString(sumA[:])
	bJSON := `{"schema_version":"mechanism/v1","rev":"B"}`
	insertTestSignatureContent(t, st, proposalID, bJSON, "cfp-b")
	sumB := sha256.Sum256([]byte(bJSON))
	hashB := hex.EncodeToString(sumB[:])

	now := time.Now().UTC()
	staleID := domain.NewEvaluationID(now)
	currentID := domain.NewEvaluationID(now.Add(time.Second))
	run := EvaluationRunRecord{
		ID:        domain.NewEvaluationRunID(now),
		ProblemID: problemID,
		RunID:     res.Record.RunID,
		Mode:      "proposal",
		CreatedAt: formatTime(now),
		Evaluations: []EvaluationRow{
			{
				// STALE content, decisive, STRONGEST tier, even LATER timestamp
				// below via direct insert would not help it — compatibility is
				// ranked first.
				ID: staleID, ProposalID: proposalID, Verdict: "failure",
				VerifierKind: "deterministic-check", VerificationStrength: "deterministic",
				ConfidenceOrdinal: "high", ToolName: "d", ToolVersion: "v1",
				SignatureContentHash: hashA,
				TargetVerdicts:       []EvaluationTargetVerdictRow{{InvariantID: targetInvID, Verdict: "violates", Violated: true}},
			},
			{
				// CURRENT content, decisive, weaker (model) tier.
				ID: currentID, ProposalID: proposalID, Verdict: "partial_success",
				VerifierKind: "model-judgment", VerificationStrength: "single-model-judgment",
				ConfidenceOrdinal: "medium", ToolName: "m", ToolVersion: "v1",
				SignatureContentHash: hashB,
				TargetVerdicts:       []EvaluationTargetVerdictRow{{InvariantID: targetInvID, Verdict: "violates", Violated: true}},
			},
		},
	}
	if _, err := st.PersistEvaluationRun(ctx, run); err != nil {
		t.Fatalf("persist evaluations: %v", err)
	}

	rows, err := st.ListBreakCohortRows(ctx, problemID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("want 1 row, got %d", len(rows))
	}
	r := rows[0]
	if r.EvaluationID != currentID || r.Result != "partial_success" {
		t.Fatalf("the CURRENT-revision reassessment must be selected over stronger stale evidence; got %+v", r)
	}
	if r.ContentHash != hashB || r.SignatureJSON != bJSON {
		t.Fatalf("cohort content must be the current assessed bytes (B); got hash=%s", r.ContentHash)
	}
	if r.ContentHash != r.LatestContentHash {
		t.Fatalf("current-compatible selection must not read as pending: %s vs %s", r.ContentHash, r.LatestContentHash)
	}

	// Control: with ONLY the stale evaluation available, it is still selected
	// (strongest among what exists) and the hash pair exposes pending-ness —
	// stale evidence is reachable as history, never silently current.
	st2 := openMigratedStore(t)
	rec2 := sampleFrontier(t, st2)
	res2, err := st2.PersistFrontierGeneration(ctx, rec2)
	if err != nil {
		t.Fatalf("persist frontier 2: %v", err)
	}
	p2 := res2.Record.Proposals[0].ID
	inv2 := res2.Record.Proposals[0].Targets[0].InvariantID
	insertTestSignatureContent(t, st2, p2, aJSON, "cfp-a2")
	insertTestSignatureContent(t, st2, p2, bJSON, "cfp-b2")
	staleOnly := domain.NewEvaluationID(now)
	run2 := EvaluationRunRecord{
		ID:        domain.NewEvaluationRunID(now),
		ProblemID: res2.Record.ProblemID,
		RunID:     res2.Record.RunID,
		Mode:      "proposal",
		CreatedAt: formatTime(now),
		Evaluations: []EvaluationRow{{
			ID: staleOnly, ProposalID: p2, Verdict: "failure",
			VerifierKind: "deterministic-check", VerificationStrength: "deterministic",
			ConfidenceOrdinal: "high", ToolName: "d", ToolVersion: "v1",
			SignatureContentHash: hashA,
			TargetVerdicts:       []EvaluationTargetVerdictRow{{InvariantID: inv2, Verdict: "violates", Violated: true}},
		}},
	}
	if _, err := st2.PersistEvaluationRun(ctx, run2); err != nil {
		t.Fatalf("persist stale-only: %v", err)
	}
	rows2, err := st2.ListBreakCohortRows(ctx, res2.Record.ProblemID)
	if err != nil {
		t.Fatalf("list 2: %v", err)
	}
	if len(rows2) != 1 || rows2[0].EvaluationID != staleOnly {
		t.Fatalf("with no current-compatible assessment the stale one is still selected: %+v", rows2)
	}
	if rows2[0].ContentHash != hashA || rows2[0].LatestContentHash != hashB {
		t.Fatalf("stale-only selection must expose the hash mismatch for pending detection: %+v", rows2[0])
	}
}
