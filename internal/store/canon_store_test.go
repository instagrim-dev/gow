package store

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/canon"
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

// TestRejectedTermSurvivesPersistenceRoundTrip guards the durable-rejected fix:
// a rejected key seeded into the store must still resolve to ResolutionRejected
// after the vocabulary is rebuilt from SQLite (runtime always reloads from the
// store, so an in-memory-only rejected set would silently vanish on reload).
func TestRejectedTermSurvivesPersistenceRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := openCanonStore(t, ctx)
	defer repo.Close()

	const version = "mechanism/v1"
	if err := repo.SeedVocabulary(ctx, VocabularySeedInput{
		Version:   version,
		Notes:     "test",
		CreatedAt: "2026-09-10T12:00:00Z",
		Terms: []TermRecord{{
			VocabularyVersion: version,
			CanonicalID:       "core.operator.modular_decomposition",
			FieldKind:         "operator",
			Description:       "test",
			Aliases:           []string{"modular decomposition"},
		}},
		Rejected: []string{canon.Normalize("handwaving")},
	}); err != nil {
		t.Fatalf("SeedVocabulary() error = %v", err)
	}

	// Rebuild the vocabulary purely from persisted rows, exactly as runtime does.
	terms, err := repo.ListTerms(ctx, version, "")
	if err != nil {
		t.Fatalf("ListTerms() error = %v", err)
	}
	defs := make([]canon.TermDef, 0, len(terms))
	for _, tr := range terms {
		defs = append(defs, canon.TermDef{
			CanonicalID: domain.CanonicalID(tr.CanonicalID),
			FieldKind:   domain.FieldKind(tr.FieldKind),
			Description: tr.Description,
			Parent:      domain.CanonicalID(tr.ParentCanonicalID),
			Aliases:     tr.Aliases,
		})
	}
	rejected, err := repo.ListRejected(ctx, version)
	if err != nil {
		t.Fatalf("ListRejected() error = %v", err)
	}
	if len(rejected) != 1 || rejected[0] != "handwaving" {
		t.Fatalf("ListRejected() = %v, want [handwaving]", rejected)
	}
	v, err := canon.BuildVocabularyWithRejected(version, defs, rejected)
	if err != nil {
		t.Fatalf("BuildVocabularyWithRejected() error = %v", err)
	}
	if got := v.Resolve(domain.FieldOperator, "Handwaving", true); got.State != domain.ResolutionRejected {
		t.Fatalf("reloaded Resolve(handwaving) = %q, want rejected (rejection did not survive persistence)", got.State)
	}
	// Sanity: a real term still resolves after the round trip.
	if got := v.Resolve(domain.FieldOperator, "modular decomposition", false); got.State != domain.ResolutionResolved {
		t.Fatalf("reloaded Resolve(modular decomposition) = %q, want resolved", got.State)
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

// TestAliasNamespaceIsPerFieldKind is the persistence guard for the #9 finding:
// the alias namespace is (vocabulary_version, field_kind, alias_normalized), so
// the same normalized phrase may legally bind to different canonical IDs in
// different field kinds. The prior PK (version, alias_normalized) + INSERT OR
// IGNORE silently dropped the second binding; this test fails if that returns.
func TestAliasNamespaceIsPerFieldKind(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := openCanonStore(t, ctx)
	defer repo.Close()

	// Same alias phrase "averaging" under two different field kinds.
	err := repo.SeedVocabulary(ctx, VocabularySeedInput{
		Version:   "mechanism/collide",
		Notes:     "cross-field-kind alias test",
		CreatedAt: "2026-09-10T12:00:00Z",
		Terms: []TermRecord{
			{
				VocabularyVersion: "mechanism/collide",
				CanonicalID:       "core.operator.density_averaging",
				FieldKind:         "operator",
				Aliases:           []string{"averaging"},
			},
			{
				VocabularyVersion: "mechanism/collide",
				CanonicalID:       "core.assumption.averaging_admissible",
				FieldKind:         "assumption",
				Aliases:           []string{"averaging"},
			},
		},
	})
	if err != nil {
		t.Fatalf("SeedVocabulary() with cross-field-kind alias error = %v", err)
	}

	// Both aliases must survive persistence — one per canonical id.
	opTerm, err := repo.GetTerm(ctx, "mechanism/collide", "core.operator.density_averaging")
	if err != nil {
		t.Fatalf("GetTerm(operator) error = %v", err)
	}
	asTerm, err := repo.GetTerm(ctx, "mechanism/collide", "core.assumption.averaging_admissible")
	if err != nil {
		t.Fatalf("GetTerm(assumption) error = %v", err)
	}
	if len(opTerm.Aliases) != 1 || opTerm.Aliases[0] != "averaging" {
		t.Fatalf("operator term aliases = %v, want [averaging] (alias silently dropped?)", opTerm.Aliases)
	}
	if len(asTerm.Aliases) != 1 || asTerm.Aliases[0] != "averaging" {
		t.Fatalf("assumption term aliases = %v, want [averaging] (alias silently dropped?)", asTerm.Aliases)
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

// TestSignatureCarriesPostureOutcomeBoundaryProvenance proves posture, outcome,
// and boundary provenance are persisted and reloaded rather than lost or
// silently defaulted to explicit. An unprovenanced posture axis / outcome must
// round-trip as its recorded status (here unknown), so downstream invariant
// mining reads provenance instead of assuming source backing.
func TestSignatureCarriesPostureOutcomeBoundaryProvenance(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := openCanonStore(t, ctx)
	defer repo.Close()
	seedTestVocabulary(t, ctx, repo)

	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	problemID, runID, snapshotID := seedSnapshotForNormalizeTests(t, ctx, repo)
	input := normalizationInput(t, problemID, runID, snapshotID, now, "approach-prov")
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
		Fingerprint:       "prov-fp",
		RunID:             runID,
		CreatedAt:         "2026-09-10T12:00:00Z",
		OutcomeClass:      "partial_success",
		OutcomeStatus:     "inferred",
		Posture:           map[string]string{"locality": "local", "construction": "constructive", "uncertainty": "deterministic"},
		PostureStatus:     map[string]string{"locality": "explicit", "construction": "unknown", "uncertainty": "unknown"},
		Boundaries: []SignatureBoundaryRow{
			{SurfaceLabel: "composite modulus", ResolutionState: "unknown", Relation: "stops_at", ClaimStatus: "unknown", Ordinal: 0},
		},
	}
	if _, err := repo.PersistSignature(ctx, record); err != nil {
		t.Fatalf("PersistSignature() error = %v", err)
	}

	got, err := repo.GetSignature(ctx, record.ID)
	if err != nil {
		t.Fatalf("GetSignature() error = %v", err)
	}
	if got.OutcomeStatus != "inferred" {
		t.Fatalf("outcome status = %q, want inferred (not silently promoted)", got.OutcomeStatus)
	}
	if got.PostureStatus["locality"] != "explicit" {
		t.Fatalf("locality provenance = %q, want explicit", got.PostureStatus["locality"])
	}
	if got.PostureStatus["construction"] != "unknown" {
		t.Fatalf("unprovenanced construction status = %q, want unknown (not explicit)", got.PostureStatus["construction"])
	}
	if len(got.Boundaries) != 1 || got.Boundaries[0].ClaimStatus != "unknown" {
		t.Fatalf("boundary provenance not round-tripped: %+v", got.Boundaries)
	}
}
