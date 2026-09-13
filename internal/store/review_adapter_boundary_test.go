package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/finite"
	"github.com/instagrim-dev/newf/internal/measure"
	"github.com/instagrim-dev/newf/internal/review"
	"github.com/instagrim-dev/newf/internal/store"
)

// openMigrated mirrors the in-package helper; this test lives in the
// external test package because internal/review imports internal/store,
// so the in-package helper would create an import cycle.
func openMigrated(t *testing.T) *store.Store {
	t.Helper()
	st, err := store.Open(t.TempDir() + "/boundary.sqlite")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := st.Migrate(context.Background()); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	t.Cleanup(func() { st.Close() })
	return st
}

// The persistence half of the adapter-to-consumer contract: check records
// produced by the measure and finite certificate adapters must survive the
// REAL persistence boundary — the schema's mode/outcome CHECK constraints
// and PersistReviewCheckAttempt's blocked-requires-blocker gate. The
// review-projection half lives in internal/toolreg.
func TestCertificateAdapterRecordsPersistAtTheReviewBoundary(t *testing.T) {
	st := openMigrated(t)
	ctx := context.Background()
	now := time.Now().UTC()
	created := now.Format(time.RFC3339)

	// Seed the FK targets: one obligation pinned into one policy.
	obl := store.ReviewObligationRow{
		ID: domain.NewReviewObligationID(now), ObligationKey: "adapter-boundary",
		SemanticRevision: 1, Requirement: "certificates must persist",
		AcceptanceCriteria: "row accepted", ApplicabilityRule: "always",
		PrimaryOwner: "test", CreatedAt: created,
	}
	if _, err := st.PersistReviewObligation(ctx, obl); err != nil {
		t.Fatalf("seed obligation: %v", err)
	}
	pol := store.ReviewPolicyRecord{
		Policy: store.ReviewPolicyRow{
			ID: domain.NewReviewPolicyID(now), PolicyKey: "adapter-boundary",
			Revision: 1, DecisionName: "persist", Owner: "test",
			AuthoritySource: "test", ScopeJustification: "adapter boundary test",
			CreatedAt: created,
		},
		Obligations: []store.ReviewPolicyObligationRow{{ObligationID: obl.ID, Mandatory: true}},
	}
	if _, err := st.PersistReviewPolicy(ctx, pol); err != nil {
		t.Fatalf("seed policy: %v", err)
	}

	toRow := func(rec review.CheckRecord) store.ReviewCheckAttemptRow {
		return store.ReviewCheckAttemptRow{
			ID:                domain.NewReviewCheckAttemptID(now),
			ObligationID:      obl.ID,
			PolicyID:          pol.Policy.ID,
			CaseLabel:         rec.CaseLabel,
			ProcedureRef:      rec.ProcedureRef,
			ProcedureRevision: rec.ProcedureRevision,
			InputsRef:         rec.InputsRef,
			Executor:          rec.Executor,
			Environment:       rec.Environment,
			Mode:              rec.Mode,
			Outcome:           rec.Outcome,
			OutputRef:         rec.OutputRef,
			Blocker:           rec.Blocker,
			StartedAt:         rec.StartedAt,
			EndedAt:           rec.EndedAt,
			CreatedAt:         created,
		}
	}

	// Completed path: a finite equivalence decision persists.
	holds := finite.AssessEquivalence(
		finite.Binding{Sentence: "xor commutes", Domain: finite.Domain{Width: 4, Vars: []string{"x", "y"}}},
		finite.Binary{Op: finite.OpXor, X: finite.Var{Name: "x"}, Y: finite.Var{Name: "y"}},
		finite.Binary{Op: finite.OpXor, X: finite.Var{Name: "y"}, Y: finite.Var{Name: "x"}},
	)
	completedRec, err := holds.ToCheckRecord("unused", "C1", "boundary test", "", "go test", now, now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.PersistReviewCheckAttempt(ctx, toRow(completedRec)); err != nil {
		t.Fatalf("a completed adapter record must persist (schema mode/outcome CHECK): %v", err)
	}

	// Blocked path, finite: an oversized-domain refusal persists WITH its
	// blocker (the gate rejects blocked rows with an empty blocker).
	over := finite.AssessEquivalence(
		finite.Binding{Sentence: "oversized", Domain: finite.Domain{Width: 8, Vars: []string{"a", "b", "c"}}},
		finite.Var{Name: "a"}, finite.Var{Name: "a"},
	)
	blockedRec, err := over.ToCheckRecord("unused", "C2", "boundary test", "", "go test", now, now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.PersistReviewCheckAttempt(ctx, toRow(blockedRec)); err != nil {
		t.Fatalf("a blocked adapter record must persist with its blocker: %v", err)
	}

	// Blocked path, measure: the probabilistic refusal persists too.
	prob := measure.AssessProbabilistic(measure.Binding{Sentence: "p constant", Kind: measure.ClaimProbabilisticProperty})
	probRec, err := prob.ToCheckRecord("unused", "C3", "boundary test", "", "go test", now, now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.PersistReviewCheckAttempt(ctx, toRow(probRec)); err != nil {
		t.Fatalf("measure's blocked record must persist with its blocker: %v", err)
	}

	// The gate itself still refuses a blocker-less blocked row: the
	// adapters satisfy the contract, they do not weaken it.
	bare := toRow(blockedRec)
	bare.ID = domain.NewReviewCheckAttemptID(now.Add(time.Second))
	bare.Blocker = ""
	if _, err := st.PersistReviewCheckAttempt(ctx, bare); err == nil {
		t.Fatal("a blocked row without a blocker must still be refused")
	}
}
