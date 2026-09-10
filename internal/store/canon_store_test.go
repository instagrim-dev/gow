package store

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/domain"
)

func openCanonStore(t *testing.T, ctx context.Context) *Store {
	t.Helper()
	repo, err := Open(filepath.Join(t.TempDir(), "newf.db"))
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	if err := repo.Migrate(ctx); err != nil {
		t.Fatalf("Migrate() error = %v", err)
	}
	return repo
}

func seedTestVocabulary(t *testing.T, ctx context.Context, repo *Store) {
	t.Helper()
	err := repo.SeedVocabulary(ctx, VocabularySeedInput{
		Version:   "mechanism/v1",
		Notes:     "test",
		CreatedAt: "2026-09-10T12:00:00Z",
		Terms: []TermRecord{
			{
				VocabularyVersion: "mechanism/v1",
				CanonicalID:       "core.operator.modular_decomposition",
				FieldKind:         "operator",
				Description:       "test",
				Aliases:           []string{"modular decomposition"},
			},
		},
	})
	if err != nil {
		t.Fatalf("SeedVocabulary() error = %v", err)
	}
}

func TestSeedVocabularyIdempotentAndImmutable(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := openCanonStore(t, ctx)
	defer repo.Close()

	seedTestVocabulary(t, ctx, repo)
	seedTestVocabulary(t, ctx, repo) // second call must be a no-op, not an error

	vocabs, err := repo.ListVocabularies(ctx)
	if err != nil {
		t.Fatalf("ListVocabularies() error = %v", err)
	}
	if len(vocabs) != 1 {
		t.Fatalf("expected 1 vocabulary after double seed, got %d", len(vocabs))
	}

	// Immutability trigger: UPDATE must abort.
	_, err = repo.db.ExecContext(ctx, `UPDATE canonical_terms SET description = 'x' WHERE canonical_id = 'core.operator.modular_decomposition'`)
	if err == nil {
		t.Fatal("UPDATE on canonical_terms succeeded, want immutability abort")
	}
}

func TestListTermsUnknownVocabularyNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := openCanonStore(t, ctx)
	defer repo.Close()

	if _, err := repo.ListTerms(ctx, "mechanism/does-not-exist", ""); err == nil {
		t.Fatal("ListTerms on missing vocabulary succeeded, want not_found")
	}
}

func TestPersistSignatureIdempotentAndImmutable(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := openCanonStore(t, ctx)
	defer repo.Close()
	seedTestVocabulary(t, ctx, repo)

	// A signature needs a real mechanism + run (FKs). Seed via normalization.
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	problemID, runID, snapshotID := seedSnapshotForNormalizeTests(t, ctx, repo)
	input := normalizationInput(t, problemID, runID, snapshotID, now, "approach-x")
	writeResult, err := repo.PersistNormalization(ctx, input)
	if err != nil {
		t.Fatalf("PersistNormalization() error = %v", err)
	}
	mechanismID := writeResult.Approaches[0].MechanismID

	record := SignatureRecord{
		ID:                domain.NewMechanismSignatureID(now),
		MechanismID:       mechanismID,
		SchemaVersion:     "mechanism/v1",
		VocabularyVersion: "mechanism/v1",
		Fingerprint:       "abc123",
		RunID:             runID,
		CreatedAt:         "2026-09-10T12:00:00Z",
		OutcomeClass:      "partial_success",
		Posture:           map[string]string{"locality": "local", "construction": "constructive", "uncertainty": "deterministic"},
		FieldClaims: []SignatureFieldClaimRow{
			{FieldKind: "operator", SurfaceLabel: "modular decomposition", ResolutionState: "resolved", CanonicalID: "core.operator.modular_decomposition", ClaimStatus: "explicit", Ordinal: 0},
		},
	}

	first, err := repo.PersistSignature(ctx, record)
	if err != nil {
		t.Fatalf("PersistSignature() error = %v", err)
	}
	if !first.Created {
		t.Fatal("first PersistSignature Created = false, want true")
	}

	// Idempotent: same (mechanism, schema, vocab) returns existing.
	dup := record
	dup.ID = domain.NewMechanismSignatureID(now.Add(time.Second))
	dup.Fingerprint = "different-but-ignored"
	second, err := repo.PersistSignature(ctx, dup)
	if err != nil {
		t.Fatalf("second PersistSignature() error = %v", err)
	}
	if second.Created {
		t.Fatal("second PersistSignature Created = true, want existing")
	}
	if second.Record.ID != first.Record.ID {
		t.Fatalf("existing signature returned different id: %s vs %s", second.Record.ID, first.Record.ID)
	}
	if second.Record.Fingerprint != "abc123" {
		t.Fatalf("existing signature fingerprint = %q, want original abc123", second.Record.Fingerprint)
	}

	// Immutability trigger.
	if _, err := repo.db.ExecContext(ctx, `UPDATE mechanism_signatures SET fingerprint = 'z' WHERE id = ?`, first.Record.ID); err == nil {
		t.Fatal("UPDATE on mechanism_signatures succeeded, want immutability abort")
	}
}
