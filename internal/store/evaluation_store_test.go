package store

import (
	"context"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/domain"
)

// persistOneProposal persists a frontier generation and returns a real proposal
// id (with NULL result) plus its problem/run ids, for evaluation tests.
func persistOneProposal(t *testing.T, st *Store) (problemID, runID, proposalID string) {
	t.Helper()
	ctx := context.Background()
	rec := sampleFrontier(t, st)
	res, err := st.PersistFrontierGeneration(ctx, rec)
	if err != nil {
		t.Fatalf("persist frontier: %v", err)
	}
	return res.Record.ProblemID, res.Record.RunID, res.Record.Proposals[0].ID
}

func sampleEvaluationRun(t *testing.T, st *Store, verdict, kind, strength string, withProvider bool) (EvaluationRunRecord, string) {
	t.Helper()
	problemID, runID, proposalID := persistOneProposal(t, st)
	now := time.Now().UTC()
	e := EvaluationRow{
		ID:                   domain.NewEvaluationID(now),
		ProposalID:           proposalID,
		Verdict:              verdict,
		VerifierKind:         kind,
		VerificationStrength: strength,
		VerificationSubject:  subjectForTestKind(kind),
		ConfidenceOrdinal:    "medium",
	}
	if withProvider {
		e.Invocation = &EvaluationProviderInvocation{
			ID: domain.NewProviderInvocationID(now), RunID: runID,
			ProviderName: "fixture", ProviderVersion: "v1", ModelName: "deterministic-fixture",
			SchemaVersion: "evaluate/v1", RequestHash: "rh", RequestPayload: "req", ResponsePayload: "resp",
			CreatedAt: formatTime(now),
		}
	} else {
		e.ToolName = "newf-deterministic-check"
		e.ToolVersion = "evaluate/v1"
	}
	return EvaluationRunRecord{
		ID:                      domain.NewEvaluationRunID(now),
		ProblemID:               problemID,
		RunID:                   runID,
		FrontierGenerationRunID: "",
		Mode:                    "proposal",
		CreatedAt:               formatTime(now),
		Evaluations:             []EvaluationRow{e},
	}, proposalID
}

func TestPersistEvaluationRunPopulatesResultInTx(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()
	run, proposalID := sampleEvaluationRun(t, st, "partial_success", "deterministic-check", "deterministic", false)

	if _, err := st.PersistEvaluationRun(ctx, run); err != nil {
		t.Fatalf("persist evaluation: %v", err)
	}
	// R5: the proposal result mirrors the verdict, set in the same tx.
	var result string
	if err := st.db.QueryRowContext(ctx, `SELECT COALESCE(result,'') FROM frontier_proposals WHERE id = ?`, proposalID).Scan(&result); err != nil {
		t.Fatalf("read result: %v", err)
	}
	if result != "partial_success" {
		t.Fatalf("proposal result = %q, want partial_success", result)
	}
	// Round-trip: verdict + kind + strength survive.
	got, err := st.GetEvaluationRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(got.Evaluations) != 1 {
		t.Fatalf("want 1 evaluation, got %d", len(got.Evaluations))
	}
	e := got.Evaluations[0]
	if e.Verdict != "partial_success" || e.VerifierKind != "deterministic-check" || e.VerificationStrength != "deterministic" {
		t.Fatalf("round-trip lost strength: %+v", e)
	}
	if e.ToolName == "" || e.ProviderInvocationID != "" {
		t.Fatalf("deterministic tier must record tool identity and NO provider invocation: %+v", e)
	}
}

func TestPersistEvaluationRunModelTierRecordsProvider(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()
	run, _ := sampleEvaluationRun(t, st, "success", "model-judgment", "single-model-judgment", true)
	if _, err := st.PersistEvaluationRun(ctx, run); err != nil {
		t.Fatalf("persist: %v", err)
	}
	got, _ := st.GetEvaluationRun(ctx, run.ID)
	if got.Evaluations[0].ProviderInvocationID == "" {
		t.Fatal("model tier must resolve to a provider invocation (role evaluate)")
	}
	// The invocation exists with role 'evaluate'.
	var role string
	if err := st.db.QueryRowContext(ctx, `SELECT role FROM provider_invocations WHERE id = ?`, got.Evaluations[0].ProviderInvocationID).Scan(&role); err != nil {
		t.Fatalf("read invocation: %v", err)
	}
	if role != "evaluate" {
		t.Fatalf("role = %q, want evaluate", role)
	}
}

func TestEvaluationFailureReEntersAtlas(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()
	run, proposalID := sampleEvaluationRun(t, st, "failure", "deterministic-check", "deterministic", false)
	if _, err := st.PersistEvaluationRun(ctx, run); err != nil {
		t.Fatalf("persist: %v", err)
	}
	// R6: the failed proposal is now a queryable re-entered failure.
	failures, err := st.ListEvaluatedFailures(ctx, run.ProblemID)
	if err != nil {
		t.Fatalf("list evaluated failures: %v", err)
	}
	if len(failures) != 1 || failures[0].ProposalID != proposalID || failures[0].Verdict != "failure" {
		t.Fatalf("expected the failed proposal to re-enter the atlas, got %+v", failures)
	}
}

func TestEvaluationVerdictAndStrengthChecks(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()
	problemID, runID, _ := persistOneProposal(t, st)
	now := formatTime(time.Now().UTC())
	// A run row to hang evaluations off.
	evrID := domain.NewEvaluationRunID(time.Now().UTC())
	if _, err := st.db.ExecContext(ctx, `INSERT INTO evaluation_runs(id, problem_id, run_id, mode, routing_policy, evaluation_count, created_at) VALUES(?,?,?,'proposal','cheap-first',0,?)`, evrID, problemID, runID, now); err != nil {
		t.Fatalf("seed run: %v", err)
	}
	// R4: a verdict outside the EPIC set is rejected by CHECK.
	if _, err := st.db.ExecContext(ctx, `INSERT INTO evaluations(id, evaluation_run_id, verdict, verifier_kind, verification_strength, created_at) VALUES(?,?,?,?,?,?)`,
		domain.NewEvaluationID(time.Now().UTC()), evrID, "totally_solved", "deterministic-check", "deterministic", now); err == nil {
		t.Fatal("a verdict outside the EPIC vocabulary must be rejected")
	}
	// R1: an evaluation without a strength is rejected (NOT NULL).
	if _, err := st.db.ExecContext(ctx, `INSERT INTO evaluations(id, evaluation_run_id, verdict, verifier_kind, created_at) VALUES(?,?,?,?,?)`,
		domain.NewEvaluationID(time.Now().UTC()), evrID, "success", "model-judgment", now); err == nil {
		t.Fatal("an evaluation without a verification_strength must be rejected")
	}
}

func TestEvaluationHoldoutModeRefused(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()
	problemID, runID, _ := persistOneProposal(t, st)
	now := formatTime(time.Now().UTC())
	// R9: a raw holdout-mode insert is blocked by the gate trigger.
	if _, err := st.db.ExecContext(ctx, `INSERT INTO evaluation_runs(id, problem_id, run_id, mode, routing_policy, evaluation_count, created_at) VALUES(?,?,?,'holdout','cheap-first',0,?)`,
		domain.NewEvaluationRunID(time.Now().UTC()), problemID, runID, now); err == nil {
		t.Fatal("holdout mode must be refused by the gate trigger (deferred to M7)")
	}
}

func TestEvaluationRunImmutable(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()
	run, _ := sampleEvaluationRun(t, st, "success", "model-judgment", "single-model-judgment", true)
	if _, err := st.PersistEvaluationRun(ctx, run); err != nil {
		t.Fatalf("persist: %v", err)
	}
	if _, err := st.db.ExecContext(ctx, `UPDATE evaluations SET verdict = 'failure' WHERE evaluation_run_id = ?`, run.ID); err == nil {
		t.Fatal("evaluations must be immutable")
	}
	if _, err := st.db.ExecContext(ctx, `UPDATE evaluation_runs SET evaluation_count = 99 WHERE id = ?`, run.ID); err == nil {
		t.Fatal("evaluation runs must be immutable")
	}
}

func TestV15ProviderRoleAllowsEvaluate(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()
	tx, err := st.db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("tx: %v", err)
	}
	defer tx.Rollback()
	allows, err := providerRoleAllows(ctx, tx, "evaluate")
	if err != nil {
		t.Fatalf("providerRoleAllows: %v", err)
	}
	if !allows {
		t.Fatal("v15 must widen provider_invocations.role to permit 'evaluate'")
	}
}

// subjectForTestKind supplies the v38 verification subject the production
// verifiers of each kind actually declare (deterministic tiers certify
// annotations; the model tier judges the domain goal).
func subjectForTestKind(kind string) string {
	switch kind {
	case "deterministic-check", "counterexample-search", "reproducible-computation":
		return "annotation"
	default:
		return "domain-goal"
	}
}
