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

// seedEvaluatedFailureForAdmission persists a proposal + failing evaluation and
// returns the ids an admission decision references.
func seedEvaluatedFailureForAdmission(t *testing.T, st *Store) (problemID, runID, proposalID, evaluationID string) {
	t.Helper()
	ctx := context.Background()
	run, pid := sampleEvaluationRun(t, st, "failure", "model-judgment", "single-model-judgment", true)
	if _, err := st.PersistEvaluationRun(ctx, run); err != nil {
		t.Fatalf("persist evaluation: %v", err)
	}
	return run.ProblemID, run.RunID, pid, run.Evaluations[0].ID
}

// A withheld decision round-trips with its refusing rule, and the same
// (evaluation, decision) pair is refused a second time: decisions are made
// once, never repeated or rewritten.
func TestEvidenceAdmissionWithheldRoundTripAndUniqueness(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()
	problemID, runID, proposalID, evaluationID := seedEvaluatedFailureForAdmission(t, st)
	now := time.Now().UTC()

	row := EvidenceAdmissionRow{
		ID:              domain.NewEvidenceAdmissionID(now),
		ProblemID:       problemID,
		RunID:           runID,
		ProposalID:      proposalID,
		EvaluationID:    evaluationID,
		Decision:        "withheld",
		ObservationKind: "model-judged-failure",
		AdmittedBy:      "rule",
		Basis:           "model-judged failure requires operator attestation",
		ContentHash:     "abc123",
		CreatedAt:       formatTime(now),
	}
	if _, err := st.PersistEvidenceAdmission(ctx, row); err != nil {
		t.Fatalf("persist withheld: %v", err)
	}

	got, err := st.ListEvidenceAdmissions(ctx, problemID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("want 1 admission, got %d", len(got))
	}
	r := got[0]
	if r.Decision != "withheld" || r.ObservationKind != "model-judged-failure" || r.AdmittedBy != "rule" ||
		r.Basis == "" || r.EvaluationID != evaluationID || r.SignatureID != "" {
		t.Fatalf("withheld row lost fields: %+v", r)
	}

	// Same decision twice is refused (UNIQUE(evaluation_id, decision)).
	dup := row
	dup.ID = domain.NewEvidenceAdmissionID(now.Add(time.Second))
	if _, err := st.PersistEvidenceAdmission(ctx, dup); err == nil {
		t.Fatal("duplicate withheld decision must be refused")
	}
}

// Decision/materialization coherence is refused before SQL: an admitted row
// without atlas ids and a withheld row with them are both invalid.
func TestEvidenceAdmissionDecisionCoherenceRefused(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()
	problemID, runID, proposalID, evaluationID := seedEvaluatedFailureForAdmission(t, st)
	now := time.Now().UTC()

	base := EvidenceAdmissionRow{
		ID:              domain.NewEvidenceAdmissionID(now),
		ProblemID:       problemID,
		RunID:           runID,
		ProposalID:      proposalID,
		EvaluationID:    evaluationID,
		ObservationKind: "domain-checked-failure",
		AdmittedBy:      "rule",
		Basis:           "test",
		CreatedAt:       formatTime(now),
	}

	admitted := base
	admitted.Decision = "admitted" // no materialization ids
	if _, err := st.PersistEvidenceAdmission(ctx, admitted); err == nil || !strings.Contains(err.Error(), "materialization ids") {
		t.Fatalf("admitted without atlas ids must be refused, got %v", err)
	}

	withheld := base
	withheld.Decision = "withheld"
	withheld.SignatureID = "msig_bogus"
	if _, err := st.PersistEvidenceAdmission(ctx, withheld); err == nil || !strings.Contains(err.Error(), "must not carry") {
		t.Fatalf("withheld with atlas ids must be refused, got %v", err)
	}
}

// Persisted admission decisions are immutable at the SQL layer.
func TestEvidenceAdmissionImmutable(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()
	problemID, runID, proposalID, evaluationID := seedEvaluatedFailureForAdmission(t, st)
	now := time.Now().UTC()

	row := EvidenceAdmissionRow{
		ID:              domain.NewEvidenceAdmissionID(now),
		ProblemID:       problemID,
		RunID:           runID,
		ProposalID:      proposalID,
		EvaluationID:    evaluationID,
		Decision:        "withheld",
		ObservationKind: "structural-claim-failure",
		AdmittedBy:      "rule",
		Basis:           "description failed its own claimed break",
		CreatedAt:       formatTime(now),
	}
	if _, err := st.PersistEvidenceAdmission(ctx, row); err != nil {
		t.Fatalf("persist: %v", err)
	}
	if _, err := st.db.ExecContext(ctx, `UPDATE evidence_admissions SET decision = 'admitted' WHERE id = ?`, row.ID); err == nil {
		t.Fatal("update must be refused by immutability trigger")
	}
	if _, err := st.db.ExecContext(ctx, `DELETE FROM evidence_admissions WHERE id = ?`, row.ID); err == nil {
		t.Fatal("delete must be refused by immutability trigger")
	}
}

// The content-addressed read returns exactly the revision whose content hash an
// evaluation assessed — and reports absence rather than substituting "latest".
func TestGetProposalSignatureContentByHash(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()
	_, _, proposalID, _ := seedEvaluatedFailureForAdmission(t, st)

	assessed := `{"schema_version":"mechanism/v1","note":"assessed"}`
	revised := `{"schema_version":"mechanism/v1","note":"revised"}`
	insertTestSignatureContent(t, st, proposalID, assessed, "cfp-assessed")
	insertTestSignatureContent(t, st, proposalID, revised, "cfp-revised")

	sum := sha256.Sum256([]byte(assessed))
	hash := hex.EncodeToString(sum[:])

	row, found, err := st.GetProposalSignatureContentByHash(ctx, proposalID, hash)
	if err != nil {
		t.Fatalf("get by hash: %v", err)
	}
	if !found || row.SignatureJSON != assessed || row.ContentHash != hash {
		t.Fatalf("want the ASSESSED revision, got found=%v %+v", found, row)
	}

	// Unknown hash: absent, not the latest revision.
	if _, found, err := st.GetProposalSignatureContentByHash(ctx, proposalID, "nope"); err != nil || found {
		t.Fatalf("unknown hash must report absence, got found=%v err=%v", found, err)
	}
	// Empty hash (evaluation without assessed content): absent by contract.
	if _, found, err := st.GetProposalSignatureContentByHash(ctx, proposalID, ""); err != nil || found {
		t.Fatalf("empty hash must report absence, got found=%v err=%v", found, err)
	}
}

// admittedFailureFixture builds one complete, valid materialization input for
// PersistAdmittedFailure. Every id is fresh; sha keys the content-addressed
// snapshot so two fixtures never dedup onto each other.
func admittedFailureFixture(t *testing.T, problemID, runID, proposalID, evaluationID, sha string, now time.Time) AdmittedFailureInput {
	t.Helper()
	invocationID := domain.NewProviderInvocationID(now)
	revisionID := domain.NewNormalizationRevisionID(now)
	approachRevisionID := domain.NewApproachRevisionID(now)
	return AdmittedFailureInput{
		Snapshot: SnapshotAdmission{
			ProblemID:   problemID,
			Kind:        domain.SourceKindLocalPath,
			LogicalName: "admitted-" + sha + ".json",
			Origin:      "evaluation:" + evaluationID,
			SHA256:      sha,
			ByteLength:  42,
			MediaType:   "application/json",
			ObjectPath:  "sha256/" + sha,
			IngestRunID: runID,
			ObservedAt:  now,
		},
		Normalization: NormalizationInput{
			Invocation: domain.ProviderInvocation{
				ID: invocationID, RunID: runID, Role: domain.RoleNormalize,
				ProviderName: "fixture", SchemaVersion: "normalize/v1",
				RequestHash: "rh-" + sha, CreatedAt: now,
			},
			Revision: domain.NormalizationRevision{
				ID: revisionID, ProblemID: problemID, RunID: runID,
				ProviderInvocationID: invocationID, SchemaVersion: "normalize/v1",
				ConfigHash: "ch", Status: domain.NormalizationStatusSucceeded, CreatedAt: now,
			},
			Approaches: []ApproachInput{{
				LogicalIdentity: "admitted-failure/" + sha,
				Revision: domain.ApproachRevision{
					ID: approachRevisionID, ApproachID: domain.NewApproachID(now),
					NormalizationRevisionID: revisionID, Label: "admitted " + sha, CreatedAt: now,
				},
				Mechanism: domain.Mechanism{
					ID: domain.NewMechanismID(now), ApproachRevisionID: approachRevisionID,
					Locality: domain.LocalityLocal, ConstructionMode: domain.ConstructionConstructive,
					UncertaintyMode: domain.UncertaintyDeterministic,
				},
				Outcome: domain.Outcome{
					ID: domain.NewOutcomeID(now), ApproachRevisionID: approachRevisionID,
					Class: domain.OutcomeFailure, BoundaryStatement: "assessed failure",
				},
			}},
		},
		Signature: SignatureRecord{
			ID: domain.NewMechanismSignatureID(now), SchemaVersion: "mechanism/v1",
			VocabularyVersion: "vocab/v1", Fingerprint: "fp-" + sha,
			RunID: runID, CreatedAt: formatTime(now), OutcomeClass: "failure",
		},
		Admission: EvidenceAdmissionRow{
			ID: domain.NewEvidenceAdmissionID(now), ProblemID: problemID, RunID: runID,
			ProposalID: proposalID, EvaluationID: evaluationID,
			Decision: "admitted", ObservationKind: "domain-checked-failure",
			AdmittedBy: "rule", Basis: "deterministic failure admissible",
			ContentHash: sha, CreatedAt: formatTime(now),
		},
	}
}

func countRows(t *testing.T, st *Store, table string) int {
	t.Helper()
	var n int
	if err := st.db.QueryRow(`SELECT COUNT(*) FROM ` + table).Scan(&n); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return n
}

// ensureTestVocabulary registers the canonical vocabulary version the fixture
// signature declares (mechanism_signatures.vocabulary_version FK).
func ensureTestVocabulary(t *testing.T, st *Store) {
	t.Helper()
	if _, err := st.db.Exec(`INSERT OR IGNORE INTO canonical_vocabulary(version, notes, created_at) VALUES('vocab/v1','',?)`, formatTime(time.Now().UTC())); err != nil {
		t.Fatalf("seed vocabulary: %v", err)
	}
}

// One admitted failure materializes in a SINGLE transaction: the committed
// state carries the snapshot, normalization revision, signature, and the
// admission row pointing at all four resolved materialization ids.
func TestPersistAdmittedFailureCommitsAllRows(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()
	problemID, runID, proposalID, evaluationID := seedEvaluatedFailureForAdmission(t, st)
	ensureTestVocabulary(t, st)
	now := time.Now().UTC()

	res, err := st.PersistAdmittedFailure(ctx, admittedFailureFixture(t, problemID, runID, proposalID, evaluationID, "aaaa1111", now))
	if err != nil {
		t.Fatalf("persist admitted failure: %v", err)
	}
	if res.SnapshotID == "" || res.ApproachID == "" || res.ApproachRevisionID == "" || res.MechanismID == "" || res.SignatureID == "" {
		t.Fatalf("materialization ids must all be resolved: %+v", res)
	}

	got, err := st.ListEvidenceAdmissions(ctx, problemID)
	if err != nil || len(got) != 1 {
		t.Fatalf("want 1 admission, got %d (%v)", len(got), err)
	}
	r := got[0]
	if r.Decision != "admitted" || r.ApproachID != res.ApproachID || r.ApproachRevisionID != res.ApproachRevisionID ||
		r.MechanismID != res.MechanismID || r.SignatureID != res.SignatureID {
		t.Fatalf("admission row must carry the SAME resolved ids as the write: %+v vs %+v", r, res)
	}
	var sigCount int
	if err := st.db.QueryRow(`SELECT COUNT(*) FROM mechanism_signatures WHERE id = ?`, res.SignatureID).Scan(&sigCount); err != nil || sigCount != 1 {
		t.Fatalf("signature must be committed: count=%d err=%v", sigCount, err)
	}
}

// C1 atomicity: a failure at the FINAL insert (the admission decision row)
// must roll back every atlas-population row written earlier in the sequence.
// Population rows without the admission decision that justifies them would be
// unauditable atlas contamination.
func TestPersistAdmittedFailureRollsBackAtomically(t *testing.T) {
	st := openMigratedStore(t)
	ctx := context.Background()
	problemID, runID, proposalID, evaluationID := seedEvaluatedFailureForAdmission(t, st)
	ensureTestVocabulary(t, st)
	now := time.Now().UTC()

	// First admission succeeds and occupies UNIQUE(evaluation_id, 'admitted').
	if _, err := st.PersistAdmittedFailure(ctx, admittedFailureFixture(t, problemID, runID, proposalID, evaluationID, "bbbb2222", now)); err != nil {
		t.Fatalf("first admission: %v", err)
	}

	tables := []string{"source_snapshots", "normalization_revisions", "approaches", "approach_revisions", "mechanisms", "mechanism_signatures", "evidence_admissions"}
	before := make(map[string]int, len(tables))
	for _, tb := range tables {
		before[tb] = countRows(t, st, tb)
	}

	// Second admission for the SAME evaluation: snapshot, normalization, and
	// signature inserts all succeed, then the admission insert violates
	// UNIQUE(evaluation_id, decision). The whole transaction must vanish.
	_, err := st.PersistAdmittedFailure(ctx, admittedFailureFixture(t, problemID, runID, proposalID, evaluationID, "cccc3333", now.Add(time.Second)))
	if err == nil {
		t.Fatal("duplicate admitted decision must fail")
	}
	for _, tb := range tables {
		if after := countRows(t, st, tb); after != before[tb] {
			t.Errorf("%s: mid-sequence failure leaked rows: before=%d after=%d", tb, before[tb], after)
		}
	}
}
