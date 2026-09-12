package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/domain"
)

// This tests the actual population reader against its minimal relational
// boundary. Evaluation tests below use the full migrated production schema.
func TestSignaturePopulationCurrentAndHistorical(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	db.SetMaxOpenConns(1)
	st := &Store{db: db}
	ctx := context.Background()
	problemID := domain.NewProblemID(time.Now().UTC())
	for _, ddl := range []string{
		`CREATE TABLE approaches (id TEXT PRIMARY KEY, problem_id TEXT)`,
		`CREATE TABLE approach_revisions (id TEXT PRIMARY KEY, approach_id TEXT, supersedes_revision_id TEXT, created_at TEXT)`,
		`CREATE TABLE mechanisms (id TEXT PRIMARY KEY, approach_revision_id TEXT)`,
		`CREATE TABLE mechanism_signatures (id TEXT PRIMARY KEY, mechanism_id TEXT, schema_version TEXT, vocabulary_version TEXT)`,
	} {
		if _, err := db.Exec(ddl); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(`INSERT INTO approaches VALUES ('a', ?), ('other', 'other-problem')`, problemID); err != nil {
		t.Fatal(err)
	}
	for _, stmt := range []string{
		`INSERT INTO approach_revisions VALUES ('old', 'a', NULL, '2026-09-12'), ('new', 'a', 'old', '2026-09-11'), ('other', 'other', NULL, '2026-09-13')`,
		`INSERT INTO mechanisms VALUES ('mo', 'old'), ('mn', 'new'), ('mx', 'other')`,
		`INSERT INTO mechanism_signatures VALUES ('sig-old', 'mo', 's1', 'v1'), ('sig-new', 'mn', 's1', 'v1'), ('sig-other', 'mx', 's1', 'v1'), ('sig-v2', 'mn', 's1', 'v2')`,
	} {
		if _, err := db.Exec(stmt); err != nil {
			t.Fatal(err)
		}
	}
	current, err := st.ListSignaturesForProblem(ctx, problemID, "s1", "v1")
	if err != nil || !reflect.DeepEqual(current, []string{"sig-new"}) {
		t.Fatalf("current population must follow supersession even when backdated: %v / %v", current, err)
	}
	history, err := st.ListHistoricalSignaturesForProblem(ctx, problemID, "s1", "v1")
	if err != nil || !reflect.DeepEqual(history, []string{"sig-new", "sig-old"}) {
		t.Fatalf("history must retain both interpretations: %v / %v", history, err)
	}
	// The latest interpretation has no signature in v1. Filtering by vocabulary
	// must not quietly bring its superseded predecessor back into the atlas.
	if _, err := db.Exec(`INSERT INTO approach_revisions VALUES ('pending', 'a', 'new', '2026-09-10')`); err != nil {
		t.Fatal(err)
	}
	current, err = st.ListSignaturesForProblem(ctx, problemID, "s1", "v1")
	if err != nil || len(current) != 0 {
		t.Fatalf("missing current signature resurrected old evidence: %v / %v", current, err)
	}
	history, err = st.ListHistoricalSignaturesForProblem(ctx, problemID, "s1", "v1")
	if err != nil || len(history) != 2 {
		t.Fatalf("a correction must not erase history: %v / %v", history, err)
	}
}

func recordAssessmentViewTest(t *testing.T, st *Store, gen FrontierGenerationRecord, proposalID, hash, verdict string) {
	t.Helper()
	// Equal clocks exercise insertion-order tie-breaking, not lexical/random IDs.
	now := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	_, err := st.PersistEvaluationRun(context.Background(), EvaluationRunRecord{
		ID:                      domain.NewEvaluationRunID(now),
		ProblemID:               gen.ProblemID,
		RunID:                   gen.RunID,
		FrontierGenerationRunID: gen.ID,
		ClusterRunID:            gen.ClusterRunID,
		Mode:                    "proposal",
		CreatedAt:               formatTime(now),
		Evaluations: []EvaluationRow{{
			ID:                   domain.NewEvaluationID(now),
			ProposalID:           proposalID,
			SignatureContentHash: hash,
			Verdict:              verdict,
			VerifierKind:         "deterministic-check",
			VerificationStrength: "deterministic",
			ToolName:             "assessment-read-test",
			ToolVersion:          "v1",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestOccurrenceResultReversalsPreserveInitialHistory(t *testing.T) {
	for _, verdicts := range [][2]string{{"success", "failure"}, {"failure", "success"}} {
		t.Run(verdicts[0]+"-to-"+verdicts[1], func(t *testing.T) {
			st := openMigratedStore(t)
			ctx := context.Background()
			gen := sampleFrontier(t, st)
			gen.Proposals[0].SignatureJSON = `{"schema_version":"mechanism/v1"}`
			gen.Proposals[0].CanonicalFingerprint = "assessment-fingerprint"
			if _, err := st.PersistFrontierGeneration(ctx, gen); err != nil {
				t.Fatal(err)
			}
			pid := gen.Proposals[0].ID
			hash := fmt.Sprintf("%x", sha256.Sum256([]byte(gen.Proposals[0].SignatureJSON)))
			for _, verdict := range verdicts {
				recordAssessmentViewTest(t, st, gen, pid, hash, verdict)
			}
			got, err := st.GetFrontierGeneration(ctx, gen.ID)
			if err != nil || len(got.Proposals) != 1 || got.Proposals[0].Result.String != verdicts[1] {
				t.Fatalf("owned read must expose latest exact assessment: %+v / %v", got, err)
			}
			rows, err := st.ListOccurrenceProposalRows(ctx, gen.ID)
			if err != nil || len(rows) != 1 || rows[0].Result.String != verdicts[1] {
				t.Fatalf("occurrence read must agree: %+v / %v", rows, err)
			}
			var initial string
			if err := st.db.QueryRow(`SELECT result FROM frontier_proposals WHERE id = ?`, pid).Scan(&initial); err != nil {
				t.Fatal(err)
			}
			if initial != verdicts[0] {
				t.Fatalf("historical initial verdict was rewritten: %q", initial)
			}
		})
	}
}

func TestOccurrenceResultDoesNotCrossGenerationOrContent(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()
	first := sampleFrontier(t, st)
	first.Proposals[0].SignatureJSON = `{"schema_version":"mechanism/v1"}`
	first.Proposals[0].CanonicalFingerprint = "assessment-fingerprint"
	if _, err := st.PersistFrontierGeneration(ctx, first); err != nil {
		t.Fatal(err)
	}
	pid := first.Proposals[0].ID
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(first.Proposals[0].SignatureJSON)))
	recordAssessmentViewTest(t, st, first, pid, hash, "success")
	second := sampleFrontierReusingProblem(t, st, first)
	if _, err := st.PersistFrontierGeneration(ctx, second); err != nil {
		t.Fatal(err)
	}
	rows, err := st.ListOccurrenceProposalRows(ctx, second.ID)
	if err != nil || len(rows) != 1 || rows[0].Result.Valid {
		t.Fatalf("identical content in a new generation is not already assessed: %+v / %v", rows, err)
	}
	recordAssessmentViewTest(t, st, second, pid, "wrong-content-hash", "success")
	rows, err = st.ListOccurrenceProposalRows(ctx, second.ID)
	if err != nil || len(rows) != 1 || rows[0].Result.Valid {
		t.Fatalf("wrong-content assessment must not match: %+v / %v", rows, err)
	}
	recordAssessmentViewTest(t, st, second, pid, hash, "failure")
	rows, err = st.ListOccurrenceProposalRows(ctx, second.ID)
	if err != nil || len(rows) != 1 || rows[0].Result.String != "failure" {
		t.Fatalf("second generation must use its own assessment: %+v / %v", rows, err)
	}
	original, err := st.GetFrontierGeneration(ctx, first.ID)
	if err != nil || len(original.Proposals) != 1 || original.Proposals[0].Result.String != "success" {
		t.Fatalf("later context contaminated historical generation: %+v / %v", original, err)
	}
}

func TestLegacyResultNeedsMatchingAssessmentContext(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()
	gen := sampleFrontier(t, st) // no persisted signature content or binding
	if _, err := st.PersistFrontierGeneration(ctx, gen); err != nil {
		t.Fatal(err)
	}
	pid := gen.Proposals[0].ID
	recordAssessmentViewTest(t, st, gen, pid, "", "failure")
	recordAssessmentViewTest(t, st, gen, pid, "", "success")
	got, err := st.GetFrontierGeneration(ctx, gen.ID)
	if err != nil || len(got.Proposals) != 1 || got.Proposals[0].Result.String != "success" {
		t.Fatalf("legacy generation must still derive its scoped latest result: %+v / %v", got, err)
	}
}
