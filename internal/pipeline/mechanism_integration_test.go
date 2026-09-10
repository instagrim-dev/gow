package pipeline

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/store"
)

func loadSignatureForTest(ctx context.Context, t *testing.T, dbPath, id string) (store.SignatureRecord, error) {
	t.Helper()
	repo := openTestStore(t, ctx, dbPath)
	defer repo.Close()
	return repo.GetSignature(ctx, id)
}

func fixturePath(name string) string {
	return filepath.Join("..", "..", "testdata", "fixtures", "mechanism", name)
}

// seedAndSignature loads a single-approach fixture and returns the signature
// response for its one mechanism.
func seedAndSignature(t *testing.T, ctx context.Context, app *App, dbPath, problemID, runID, snapshotID, fixture, vocabVersion string) SignatureResponse {
	t.Helper()
	seed, err := app.SeedMechanismFixture(ctx, MechanismFixtureSeedInput{
		DBPath:     dbPath,
		ProblemID:  problemID,
		RunID:      runID,
		SnapshotID: snapshotID,
		Path:       fixturePath(fixture),
	})
	if err != nil {
		t.Fatalf("SeedMechanismFixture(%s) error = %v", fixture, err)
	}
	if len(seed.MechanismIDs) != 1 {
		t.Fatalf("fixture %s produced %d mechanisms, want 1", fixture, len(seed.MechanismIDs))
	}
	sig, err := app.SignatureMechanism(ctx, SignatureInput{
		DBPath:       dbPath,
		MechanismID:  seed.MechanismIDs[0],
		VocabVersion: vocabVersion,
	})
	if err != nil {
		t.Fatalf("SignatureMechanism(%s) error = %v", fixture, err)
	}
	return sig
}

func compareFixtures(t *testing.T, ctx context.Context, app *App, dbPath, problemID, runID, snapshotID, fa, fb string) CompareResponse {
	t.Helper()
	seedA, err := app.SeedMechanismFixture(ctx, MechanismFixtureSeedInput{DBPath: dbPath, ProblemID: problemID, RunID: runID, SnapshotID: snapshotID, Path: fixturePath(fa)})
	if err != nil {
		t.Fatalf("seed %s: %v", fa, err)
	}
	seedB, err := app.SeedMechanismFixture(ctx, MechanismFixtureSeedInput{DBPath: dbPath, ProblemID: problemID, RunID: runID, SnapshotID: snapshotID, Path: fixturePath(fb)})
	if err != nil {
		t.Fatalf("seed %s: %v", fb, err)
	}
	cmp, err := app.CompareMechanisms(ctx, CompareInput{
		DBPath:       dbPath,
		MechanismAID: seedA.MechanismIDs[0],
		MechanismBID: seedB.MechanismIDs[0],
	})
	if err != nil {
		t.Fatalf("CompareMechanisms(%s,%s): %v", fa, fb, err)
	}
	return cmp
}

func TestIntegrationCase1SameCanonical(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	problemID, runID, snapshotID := seedProblemAndSnapshot(t, ctx, dbPath, now)

	cmp := compareFixtures(t, ctx, app, dbPath, problemID, runID, snapshotID, "case1_same_canonical_a.json", "case1_same_canonical_b.json")
	if cmp.FingerprintA != cmp.FingerprintB {
		t.Fatalf("case1 fingerprints differ: %s vs %s", cmp.FingerprintA, cmp.FingerprintB)
	}
	switch cmp.Comparison.Classification {
	case string(canon.ClassMechanismNear), string(canon.ClassSurfaceDistinctMechNear):
	default:
		t.Fatalf("case1 classification = %q, want a mechanism-near variant", cmp.Comparison.Classification)
	}
}

func TestIntegrationCase2SurfaceNearMechanismDistinct(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	problemID, runID, snapshotID := seedProblemAndSnapshot(t, ctx, dbPath, now)

	cmp := compareFixtures(t, ctx, app, dbPath, problemID, runID, snapshotID, "case2_surface_near_a.json", "case2_mechanism_distinct_b.json")
	switch cmp.Comparison.Classification {
	case string(canon.ClassMechanismDistinct), string(canon.ClassSurfaceNearMechDistinct):
	default:
		t.Fatalf("case2 classification = %q, want a mechanism-distinct variant", cmp.Comparison.Classification)
	}
	if cmp.FingerprintA == cmp.FingerprintB {
		t.Fatal("case2 mechanistically-distinct fixtures produced equal fingerprints")
	}
}

func TestIntegrationCase4NovelCandidatePreserved(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	problemID, runID, snapshotID := seedProblemAndSnapshot(t, ctx, dbPath, now)

	sig := seedAndSignature(t, ctx, app, dbPath, problemID, runID, snapshotID, "case4_novel_candidate.json", "")
	// The novel/unknown operator is retained with a non-resolved state and no
	// canonical ID (never coerced).
	var found bool
	for _, c := range sig.Signature.FieldClaims {
		if c.FieldKind == "operator" {
			found = true
			if c.CanonicalID != "" {
				t.Fatalf("novel operator coerced to canonical id %q", c.CanonicalID)
			}
			if c.ResolutionState == "resolved" {
				t.Fatalf("novel operator marked resolved")
			}
		}
	}
	if !found {
		t.Fatal("operator claim not present in signature")
	}
}

func TestIntegrationCase6OrderingSameFingerprint(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	problemID, runID, snapshotID := seedProblemAndSnapshot(t, ctx, dbPath, now)

	cmp := compareFixtures(t, ctx, app, dbPath, problemID, runID, snapshotID, "case6_ordering_a.json", "case6_ordering_b.json")
	if cmp.FingerprintA != cmp.FingerprintB {
		t.Fatalf("case6 reordered fixtures produced different fingerprints: %s vs %s", cmp.FingerprintA, cmp.FingerprintB)
	}
}

func TestIntegrationSignatureIdempotentAndDeterministic(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	problemID, runID, snapshotID := seedProblemAndSnapshot(t, ctx, dbPath, now)

	seed, err := app.SeedMechanismFixture(ctx, MechanismFixtureSeedInput{
		DBPath: dbPath, ProblemID: problemID, RunID: runID, SnapshotID: snapshotID,
		Path: fixturePath("single_modular_descent.json"),
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	first, err := app.SignatureMechanism(ctx, SignatureInput{DBPath: dbPath, MechanismID: seed.MechanismIDs[0]})
	if err != nil {
		t.Fatalf("first signature: %v", err)
	}
	if first.Status != "created" {
		t.Fatalf("first signature status = %q, want created", first.Status)
	}
	second, err := app.SignatureMechanism(ctx, SignatureInput{DBPath: dbPath, MechanismID: seed.MechanismIDs[0]})
	if err != nil {
		t.Fatalf("second signature: %v", err)
	}
	if second.Status != "existing" {
		t.Fatalf("second signature status = %q, want existing", second.Status)
	}
	if first.Signature.Fingerprint != second.Signature.Fingerprint {
		t.Fatal("idempotent signature produced different fingerprints")
	}
	if first.Signature.ID != second.Signature.ID {
		t.Fatal("idempotent signature produced different IDs (rewrote)")
	}

	// Provenance survived: the explicit operator support is preserved.
	var sawExplicit bool
	for _, c := range first.Signature.FieldClaims {
		if c.FieldKind == "operator" && c.ClaimStatus == "explicit" {
			sawExplicit = true
		}
	}
	if !sawExplicit {
		t.Fatal("explicit claim status not preserved through signature")
	}
}

func TestIntegrationVersionIsolation(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	problemID, runID, snapshotID := seedProblemAndSnapshot(t, ctx, dbPath, now)

	seed, err := app.SeedMechanismFixture(ctx, MechanismFixtureSeedInput{
		DBPath: dbPath, ProblemID: problemID, RunID: runID, SnapshotID: snapshotID,
		Path: fixturePath("case5_lossy_a.json"),
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	v1sig, err := app.SignatureMechanism(ctx, SignatureInput{DBPath: dbPath, MechanismID: seed.MechanismIDs[0], VocabVersion: canon.VocabularyMechanismV1})
	if err != nil {
		t.Fatalf("v1 signature: %v", err)
	}
	v0sig, err := app.SignatureMechanism(ctx, SignatureInput{DBPath: dbPath, MechanismID: seed.MechanismIDs[0], VocabVersion: canon.VocabularyMechanismV0Lossy})
	if err != nil {
		t.Fatalf("v0 signature: %v", err)
	}
	// Distinct rows, and the v1 signature is unchanged by computing the v0 one.
	if v1sig.Signature.ID == v0sig.Signature.ID {
		t.Fatal("two vocab versions shared a signature row")
	}
	reRead, err := app.SignatureMechanism(ctx, SignatureInput{DBPath: dbPath, MechanismID: seed.MechanismIDs[0], VocabVersion: canon.VocabularyMechanismV1})
	if err != nil {
		t.Fatalf("re-read v1: %v", err)
	}
	if reRead.Signature.Fingerprint != v1sig.Signature.Fingerprint {
		t.Fatal("v1 fingerprint changed after computing v0 signature")
	}
}

// TestIntegrationCase5AbstractionLoss is the R9 regression: the system-owned
// AssertDiscriminationPreserved invariant fires under the lossy vocab (two
// different-outcome mechanisms collapse) and does not fire under mechanism/v1.
func TestIntegrationCase5AbstractionLoss(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	problemID, runID, snapshotID := seedProblemAndSnapshot(t, ctx, dbPath, now)

	seedA, err := app.SeedMechanismFixture(ctx, MechanismFixtureSeedInput{DBPath: dbPath, ProblemID: problemID, RunID: runID, SnapshotID: snapshotID, Path: fixturePath("case5_lossy_a.json")})
	if err != nil {
		t.Fatalf("seed a: %v", err)
	}
	seedB, err := app.SeedMechanismFixture(ctx, MechanismFixtureSeedInput{DBPath: dbPath, ProblemID: problemID, RunID: runID, SnapshotID: snapshotID, Path: fixturePath("case5_lossy_b.json")})
	if err != nil {
		t.Fatalf("seed b: %v", err)
	}

	build := func(mechID, vocab string) canon.MechanismSignature {
		sig, err := app.SignatureMechanism(ctx, SignatureInput{DBPath: dbPath, MechanismID: mechID, VocabVersion: vocab})
		if err != nil {
			t.Fatalf("signature %s @ %s: %v", mechID, vocab, err)
		}
		rec, err := loadSignatureForTest(ctx, t, dbPath, sig.Signature.ID)
		if err != nil {
			t.Fatalf("load signature: %v", err)
		}
		return signatureFromRecord(rec)
	}

	v1 := []canon.MechanismSignature{
		build(seedA.MechanismIDs[0], canon.VocabularyMechanismV1),
		build(seedB.MechanismIDs[0], canon.VocabularyMechanismV1),
	}
	if losses := canon.AssertDiscriminationPreserved(v1); len(losses) != 0 {
		t.Fatalf("v1 reported discrimination loss: %+v", losses)
	}

	v0 := []canon.MechanismSignature{
		build(seedA.MechanismIDs[0], canon.VocabularyMechanismV0Lossy),
		build(seedB.MechanismIDs[0], canon.VocabularyMechanismV0Lossy),
	}
	losses := canon.AssertDiscriminationPreserved(v0)
	if len(losses) != 1 {
		t.Fatalf("lossy vocab reported %d losses, want 1: %+v", len(losses), losses)
	}
}
