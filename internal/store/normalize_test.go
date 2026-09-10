package store

import (
	"context"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/domain"
)

func seedSnapshotForNormalizeTests(t *testing.T, ctx context.Context, repo *Store) (problemID, runID, snapshotID string) {
	t.Helper()
	problemID, ingestRunID := seedProblemForSourceTests(t, ctx, repo)
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	admission, err := repo.CreateSourceSnapshot(ctx, SnapshotAdmission{
		ProblemID:   problemID,
		Kind:        domain.SourceKindLocalPath,
		LogicalName: "approach.md",
		Origin:      "/tmp/approach.md",
		SHA256:      "deadbeef",
		ByteLength:  10,
		MediaType:   "text/markdown",
		ObjectPath:  "sha256/de/deadbeef",
		IngestRunID: ingestRunID,
		ObservedAt:  now,
	})
	if err != nil {
		t.Fatalf("CreateSourceSnapshot() error = %v", err)
	}
	return problemID, ingestRunID, admission.Snapshot.ID
}

func normalizationInput(t *testing.T, problemID, runID, snapshotID string, now time.Time, identities ...string) NormalizationInput {
	t.Helper()
	invocationID := domain.NewProviderInvocationID(now)
	revisionID := domain.NewNormalizationRevisionID(now)
	input := NormalizationInput{
		Invocation: domain.ProviderInvocation{
			ID:            invocationID,
			RunID:         runID,
			Role:          domain.RoleNormalize,
			ProviderName:  "fixture",
			SchemaVersion: "normalize/v1",
			RequestHash:   "reqhash",
			CreatedAt:     now,
		},
		Revision: domain.NormalizationRevision{
			ID:                   revisionID,
			ProblemID:            problemID,
			RunID:                runID,
			SnapshotID:           snapshotID,
			ProviderInvocationID: invocationID,
			SchemaVersion:        "normalize/v1",
			ConfigHash:           "confighash",
			Status:               domain.NormalizationStatusSucceeded,
			CreatedAt:            now,
		},
	}
	for _, identity := range identities {
		approachRevisionID := domain.NewApproachRevisionID(now)
		input.Approaches = append(input.Approaches, ApproachInput{
			LogicalIdentity: identity,
			Revision: domain.ApproachRevision{
				ID:                      approachRevisionID,
				ApproachID:              domain.NewApproachID(now),
				NormalizationRevisionID: revisionID,
				Label:                   "label:" + identity,
				CreatedAt:               now,
			},
			Mechanism: domain.Mechanism{
				ID:                 domain.NewMechanismID(now),
				ApproachRevisionID: approachRevisionID,
				Locality:           domain.LocalityLocal,
				ConstructionMode:   domain.ConstructionConstructive,
				UncertaintyMode:    domain.UncertaintyDeterministic,
			},
			Attributes: []domain.MechanismAttribute{
				{Kind: domain.AttrRepresentation, Value: "congruence classes"},
			},
			Outcome: domain.Outcome{
				ID:                 domain.NewOutcomeID(now),
				ApproachRevisionID: approachRevisionID,
				Class:              domain.OutcomePartialFailure,
				BoundaryStatement:  "uncovered residue families remain",
			},
			Boundaries: []domain.FailureBoundary{
				{ID: domain.NewFailureBoundaryID(now), ApproachRevisionID: approachRevisionID, Condition: "quadratic-residue survivors"},
			},
			Support: []domain.SourceSupport{
				{ApproachRevisionID: approachRevisionID, SnapshotID: snapshotID, FieldPath: "outcome.class", SupportKind: domain.SupportExplicit, Locator: "para:1"},
				{ApproachRevisionID: approachRevisionID, SnapshotID: snapshotID, FieldPath: "mechanism.preserves", SupportKind: domain.SupportInferred, Locator: "para:2"},
			},
		})
	}
	return input
}

func TestPersistNormalizationCreatesApproachAndRevision(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := openTestStore(t)
	defer repo.Close()

	problemID, runID, snapshotID := seedSnapshotForNormalizeTests(t, ctx, repo)
	now := time.Date(2026, 9, 10, 13, 0, 0, 0, time.UTC)
	result, err := repo.PersistNormalization(ctx, normalizationInput(t, problemID, runID, snapshotID, now, "erdos-straus/modular-residue-cover"))
	if err != nil {
		t.Fatalf("PersistNormalization() error = %v", err)
	}
	if len(result.Approaches) != 1 || !result.Approaches[0].CreatedApproach {
		t.Fatalf("expected one created approach, got %+v", result.Approaches)
	}

	detail, err := repo.GetApproachDetail(ctx, result.Approaches[0].ApproachID)
	if err != nil {
		t.Fatalf("GetApproachDetail() error = %v", err)
	}
	if detail.SnapshotID != snapshotID {
		t.Fatalf("detail snapshot = %q, want %q", detail.SnapshotID, snapshotID)
	}
	if detail.Invocation.ProviderName != "fixture" || detail.Invocation.Role != domain.RoleNormalize {
		t.Fatalf("provider metadata not persisted: %+v", detail.Invocation)
	}
}

func TestPersistNormalizationMultipleApproachesFromOneSnapshot(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := openTestStore(t)
	defer repo.Close()

	problemID, runID, snapshotID := seedSnapshotForNormalizeTests(t, ctx, repo)
	now := time.Date(2026, 9, 10, 13, 0, 0, 0, time.UTC)
	result, err := repo.PersistNormalization(ctx, normalizationInput(t, problemID, runID, snapshotID, now,
		"erdos-straus/modular-residue-cover", "erdos-straus/averaged-covering-density"))
	if err != nil {
		t.Fatalf("PersistNormalization() error = %v", err)
	}
	if len(result.Approaches) != 2 {
		t.Fatalf("approaches = %d, want 2", len(result.Approaches))
	}

	items, err := repo.ListApproaches(ctx, problemID)
	if err != nil {
		t.Fatalf("ListApproaches() error = %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("ListApproaches len = %d, want 2", len(items))
	}
}

func TestReNormalizationCreatesRevisionNotOverwrite(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := openTestStore(t)
	defer repo.Close()

	problemID, runID, snapshotID := seedSnapshotForNormalizeTests(t, ctx, repo)
	first := normalizationInput(t, problemID, runID, snapshotID, time.Date(2026, 9, 10, 13, 0, 0, 0, time.UTC), "erdos-straus/modular-residue-cover")
	firstResult, err := repo.PersistNormalization(ctx, first)
	if err != nil {
		t.Fatalf("PersistNormalization(first) error = %v", err)
	}

	second := normalizationInput(t, problemID, runID, snapshotID, time.Date(2026, 9, 10, 14, 0, 0, 0, time.UTC), "erdos-straus/modular-residue-cover")
	second.Revision.SupersedesRevisionID = &first.Revision.ID
	secondResult, err := repo.PersistNormalization(ctx, second)
	if err != nil {
		t.Fatalf("PersistNormalization(second) error = %v", err)
	}
	if firstResult.Approaches[0].ApproachID != secondResult.Approaches[0].ApproachID {
		t.Fatal("re-normalization must reuse the same logical approach identity")
	}
	if secondResult.Approaches[0].CreatedApproach {
		t.Fatal("second normalization should not create a new approach")
	}

	approach, revisions, err := repo.ListApproachRevisions(ctx, firstResult.Approaches[0].ApproachID)
	if err != nil {
		t.Fatalf("ListApproachRevisions() error = %v", err)
	}
	if approach.ID != firstResult.Approaches[0].ApproachID {
		t.Fatalf("unexpected approach %q", approach.ID)
	}
	if len(revisions) != 2 {
		t.Fatalf("historical revisions = %d, want 2 (history preserved)", len(revisions))
	}

	// The approach-revision lineage must be derived at persistence time, not
	// injected by the caller: the newest revision (index 0, newest-first) must
	// supersede the immediately prior approach revision, and the original must
	// begin a fresh chain. This asserts the production derivation path, not a
	// hand-set fixture field.
	newest, oldest := revisions[0], revisions[1]
	if newest.SupersedesRevisionID == nil {
		t.Fatal("newest approach revision must record supersedes lineage")
	}
	if *newest.SupersedesRevisionID != oldest.ID {
		t.Fatalf("approach revision lineage = %q, want prior revision %q", *newest.SupersedesRevisionID, oldest.ID)
	}
	if oldest.SupersedesRevisionID != nil {
		t.Fatalf("first approach revision must not supersede anything, got %q", *oldest.SupersedesRevisionID)
	}
}

func TestPersistNormalizationDistinguishesSupportKinds(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := openTestStore(t)
	defer repo.Close()

	problemID, runID, snapshotID := seedSnapshotForNormalizeTests(t, ctx, repo)
	now := time.Date(2026, 9, 10, 13, 0, 0, 0, time.UTC)
	result, err := repo.PersistNormalization(ctx, normalizationInput(t, problemID, runID, snapshotID, now, "erdos-straus/modular-residue-cover"))
	if err != nil {
		t.Fatalf("PersistNormalization() error = %v", err)
	}

	detail, err := repo.GetApproachDetail(ctx, result.Approaches[0].ApproachID)
	if err != nil {
		t.Fatalf("GetApproachDetail() error = %v", err)
	}
	var explicit, inferred int
	for _, support := range detail.Support {
		switch support.SupportKind {
		case domain.SupportExplicit:
			explicit++
		case domain.SupportInferred:
			inferred++
		}
		if support.SnapshotID != snapshotID {
			t.Fatalf("support links snapshot %q, want %q", support.SnapshotID, snapshotID)
		}
	}
	if explicit == 0 || inferred == 0 {
		t.Fatalf("expected both explicit and inferred support; explicit=%d inferred=%d", explicit, inferred)
	}
}

func TestPersistNormalizationRollsBackOnInvalidApproach(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := openTestStore(t)
	defer repo.Close()

	problemID, runID, snapshotID := seedSnapshotForNormalizeTests(t, ctx, repo)
	now := time.Date(2026, 9, 10, 13, 0, 0, 0, time.UTC)
	input := normalizationInput(t, problemID, runID, snapshotID, now, "erdos-straus/modular-residue-cover")
	// Corrupt the outcome class to force a validation failure after the
	// revision/invocation would otherwise be written.
	input.Approaches[0].Outcome.Class = "bogus"

	if _, err := repo.PersistNormalization(ctx, input); err == nil {
		t.Fatal("PersistNormalization() succeeded, want validation failure")
	}

	items, err := repo.ListApproaches(ctx, problemID)
	if err != nil {
		t.Fatalf("ListApproaches() error = %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected no approaches after rollback, got %d", len(items))
	}
	if _, err := repo.getNormalizationRevision(ctx, input.Revision.ID); err == nil {
		t.Fatal("normalization revision persisted despite rollback")
	}
}

func TestPersistNormalizationRejectsConflictingSupportKinds(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := openTestStore(t)
	defer repo.Close()

	problemID, runID, snapshotID := seedSnapshotForNormalizeTests(t, ctx, repo)
	now := time.Date(2026, 9, 10, 13, 0, 0, 0, time.UTC)
	input := normalizationInput(t, problemID, runID, snapshotID, now, "erdos-straus/modular-residue-cover")
	// Two support rows for the same field_path with differing epistemic
	// strength must never silently collapse: the persistence layer rejects the
	// conflict (defense in depth behind schema validation).
	revID := input.Approaches[0].Revision.ID
	input.Approaches[0].Support = []domain.SourceSupport{
		{ApproachRevisionID: revID, SnapshotID: snapshotID, FieldPath: "outcome.class", SupportKind: domain.SupportExplicit},
		{ApproachRevisionID: revID, SnapshotID: snapshotID, FieldPath: "outcome.class", SupportKind: domain.SupportUnsupported},
	}

	if _, err := repo.PersistNormalization(ctx, input); err == nil {
		t.Fatal("PersistNormalization() accepted conflicting support kinds, want error")
	}

	items, err := repo.ListApproaches(ctx, problemID)
	if err != nil {
		t.Fatalf("ListApproaches() error = %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected rollback to leave no approaches, got %d", len(items))
	}
}

func TestFindEquivalentNormalizationIdempotency(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := openTestStore(t)
	defer repo.Close()

	problemID, runID, snapshotID := seedSnapshotForNormalizeTests(t, ctx, repo)
	now := time.Date(2026, 9, 10, 13, 0, 0, 0, time.UTC)
	input := normalizationInput(t, problemID, runID, snapshotID, now, "erdos-straus/modular-residue-cover")
	if _, err := repo.PersistNormalization(ctx, input); err != nil {
		t.Fatalf("PersistNormalization() error = %v", err)
	}

	existing, err := repo.FindEquivalentNormalization(ctx, snapshotID, "normalize/v1", "confighash")
	if err != nil {
		t.Fatalf("FindEquivalentNormalization() error = %v", err)
	}
	if !existing.Found || existing.Revision.ID != input.Revision.ID {
		t.Fatalf("expected equivalent revision %q, got %+v", input.Revision.ID, existing)
	}

	missing, err := repo.FindEquivalentNormalization(ctx, snapshotID, "normalize/v2", "confighash")
	if err != nil {
		t.Fatalf("FindEquivalentNormalization(other schema) error = %v", err)
	}
	if missing.Found {
		t.Fatal("different schema version must not be treated as equivalent")
	}
}

func TestGetMechanismDetailResolvesApproach(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	repo := openTestStore(t)
	defer repo.Close()

	problemID, runID, snapshotID := seedSnapshotForNormalizeTests(t, ctx, repo)
	now := time.Date(2026, 9, 10, 13, 0, 0, 0, time.UTC)
	result, err := repo.PersistNormalization(ctx, normalizationInput(t, problemID, runID, snapshotID, now, "erdos-straus/modular-residue-cover"))
	if err != nil {
		t.Fatalf("PersistNormalization() error = %v", err)
	}
	detail, err := repo.GetMechanismDetail(ctx, result.Approaches[0].MechanismID)
	if err != nil {
		t.Fatalf("GetMechanismDetail() error = %v", err)
	}
	if detail.Approach.ID != result.Approaches[0].ApproachID {
		t.Fatalf("mechanism resolved to approach %q, want %q", detail.Approach.ID, result.Approaches[0].ApproachID)
	}
	if len(detail.Attributes) == 0 {
		t.Fatal("expected mechanism attributes")
	}
}
