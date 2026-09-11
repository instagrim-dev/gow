package pipeline

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/store"
)

// MechanismFixture is the offline, deterministic stand-in for #7's
// provider-driven normalize step. It carries the surface-labeled mechanism
// records a normalization revision would produce, so #9 has real mechanism rows
// to canonicalize and compare without any provider call.
//
// The shape mirrors the domain types (posture enums, attribute kinds, outcome
// class, boundaries, support) but uses only surface strings; identity and
// canonicalization are #9's job, not the fixture's.
type MechanismFixture struct {
	// Approaches is one-or-more normalized approaches derived from a single
	// (fixture) source snapshot, matching "one source -> multiple approaches".
	Approaches []MechanismFixtureApproach `json:"approaches"`
}

// MechanismFixtureApproach is a single normalized approach in a fixture.
type MechanismFixtureApproach struct {
	LogicalIdentity  string                    `json:"logical_identity"`
	Label            string                    `json:"label"`
	Description      string                    `json:"description,omitempty"`
	Locality         string                    `json:"locality"`
	ConstructionMode string                    `json:"construction_mode"`
	UncertaintyMode  string                    `json:"uncertainty_mode"`
	Notes            string                    `json:"notes,omitempty"`
	Representations  []string                  `json:"representations,omitempty"`
	Assumptions      []string                  `json:"assumptions,omitempty"`
	Operators        []string                  `json:"operators,omitempty"`
	Preserves        []string                  `json:"preserves,omitempty"`
	Breaks           []string                  `json:"breaks,omitempty"`
	AuxiliaryObjects []string                  `json:"auxiliary_objects,omitempty"`
	Outcome          MechanismFixtureOutcome   `json:"outcome"`
	Boundaries       []string                  `json:"boundaries,omitempty"`
	Support          []MechanismFixtureSupport `json:"support,omitempty"`
	// CompleteFields lists set-valued field kinds this authored fixture
	// declares exhaustively extracted (declared_payload scope). The fixture
	// is deterministic in-repo data — the same trust class as the fixture
	// normalizer's embedded-payload parser — so these declarations are
	// admitted `accepted` by code. Live provider paths never route here.
	CompleteFields []string `json:"complete_fields,omitempty"`
}

// MechanismFixtureOutcome is the normalized outcome for an approach.
type MechanismFixtureOutcome struct {
	Class             string `json:"class"`
	BoundaryStatement string `json:"boundary_statement,omitempty"`
	Notes             string `json:"notes,omitempty"`
}

// MechanismFixtureSupport records per-field provenance for a fixture approach.
type MechanismFixtureSupport struct {
	FieldPath   string `json:"field_path"`
	SupportKind string `json:"support_kind"`
	Locator     string `json:"locator,omitempty"`
	Confidence  string `json:"confidence,omitempty"`
}

// MechanismFixtureSeedInput seeds a mechanism fixture against an existing
// problem/run/snapshot in a workspace store.
type MechanismFixtureSeedInput struct {
	DBPath     string
	ProblemID  string
	RunID      string
	SnapshotID string
	// Exactly one of Path / Fixture must be provided.
	Path    string
	Fixture *MechanismFixture
}

// FixtureSeedResult reports the identities created by a fixture seed.
type FixtureSeedResult struct {
	RevisionID   string
	ApproachIDs  []string
	MechanismIDs []string
}

// SeedMechanismFixture builds a normalization revision from a fixture (file or
// in-memory) and persists it via the existing store.PersistNormalization. It is
// the offline stand-in for the provider normalize path: no provider concept
// leaks in and all domain validation runs unchanged.
func (a *App) SeedMechanismFixture(ctx context.Context, input MechanismFixtureSeedInput) (FixtureSeedResult, error) {
	fixture, err := resolveFixture(input)
	if err != nil {
		return FixtureSeedResult{}, err
	}
	if len(fixture.Approaches) == 0 {
		return FixtureSeedResult{}, fmt.Errorf("mechanism fixture has no approaches")
	}

	_, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return FixtureSeedResult{}, err
	}
	defer repoStore.Close()

	if _, err := repoStore.GetProblem(ctx, input.ProblemID); err != nil {
		return FixtureSeedResult{}, err
	}

	now := a.now()
	normInput, err := buildFixtureNormalizationInput(fixture, input, now)
	if err != nil {
		return FixtureSeedResult{}, err
	}

	result, err := repoStore.PersistNormalization(ctx, normInput)
	if err != nil {
		return FixtureSeedResult{}, err
	}

	seed := FixtureSeedResult{RevisionID: result.RevisionID}
	for _, ref := range result.Approaches {
		seed.ApproachIDs = append(seed.ApproachIDs, ref.ApproachID)
		seed.MechanismIDs = append(seed.MechanismIDs, ref.MechanismID)
	}
	return seed, nil
}

// SeedFixtureForProblemInput is the CLI-facing seed request: given only a
// problem, provision a run and a synthetic source snapshot from the fixture
// content, then seed the mechanism records. This is a manual-exploration
// convenience over SeedMechanismFixture; it stays fully offline and
// deterministic and reuses the existing store admission + normalization paths.
type SeedFixtureForProblemInput struct {
	DBPath     string
	ProblemID  string
	Path       string
	JSONOutput bool
}

// SeedMechanismFixtureForProblem provisions a run + source snapshot for a
// problem and seeds the given fixture in a single store session. The snapshot's
// bytes are the fixture file itself, so provenance points at real content.
func (a *App) SeedMechanismFixtureForProblem(ctx context.Context, input SeedFixtureForProblemInput) (SeedFixtureResponse, error) {
	if strings.TrimSpace(input.Path) == "" {
		return SeedFixtureResponse{}, fmt.Errorf("fixture path is required")
	}
	raw, err := os.ReadFile(input.Path)
	if err != nil {
		return SeedFixtureResponse{}, fmt.Errorf("read mechanism fixture: %w", err)
	}
	var fixture MechanismFixture
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&fixture); err != nil {
		return SeedFixtureResponse{}, fmt.Errorf("decode mechanism fixture %q: %w", input.Path, err)
	}
	if len(fixture.Approaches) == 0 {
		return SeedFixtureResponse{}, fmt.Errorf("mechanism fixture has no approaches")
	}

	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return SeedFixtureResponse{}, err
	}
	defer repoStore.Close()

	if _, err := repoStore.GetProblem(ctx, input.ProblemID); err != nil {
		return SeedFixtureResponse{}, err
	}

	now := a.now()

	// Provision an ingest run for the synthetic snapshot.
	ingestRun, err := repoStore.CreateRun(ctx, domain.NewRun{
		ID:          domain.NewRunID(now),
		ProblemID:   input.ProblemID,
		Operation:   "mechanism seed-fixture",
		Status:      domain.RunStatusRunning,
		InputRef:    "fixture:" + filepath.Base(input.Path),
		ToolName:    "newf",
		ToolVersion: a.version,
		StartedAt:   now,
		CompletedAt: now,
	})
	if err != nil {
		return SeedFixtureResponse{}, err
	}

	sum := sha256.Sum256(raw)
	digest := hex.EncodeToString(sum[:])
	logicalName := filepath.Base(input.Path)
	admission, err := repoStore.CreateSourceSnapshot(ctx, store.SnapshotAdmission{
		ProblemID:   input.ProblemID,
		Kind:        domain.SourceKindLocalPath,
		LogicalName: logicalName,
		Origin:      "fixture://" + logicalName,
		SHA256:      digest,
		ByteLength:  int64(len(raw)),
		MediaType:   "application/json",
		ObjectPath:  filepath.Join("sha256", digest[:2], digest),
		IngestRunID: ingestRun.ID,
		ObservedAt:  now,
	})
	if err != nil {
		a.failRun(ctx, repoStore, ingestRun.ID, err)
		return SeedFixtureResponse{}, err
	}

	normInput, err := buildFixtureNormalizationInput(fixture, MechanismFixtureSeedInput{
		ProblemID:  input.ProblemID,
		RunID:      ingestRun.ID,
		SnapshotID: admission.Snapshot.ID,
	}, now)
	if err != nil {
		a.failRun(ctx, repoStore, ingestRun.ID, err)
		return SeedFixtureResponse{}, err
	}

	result, err := repoStore.PersistNormalization(ctx, normInput)
	if err != nil {
		a.failRun(ctx, repoStore, ingestRun.ID, err)
		return SeedFixtureResponse{}, err
	}

	if err := a.finalizeRun(ctx, repoStore, ingestRun.ID, nil); err != nil {
		return SeedFixtureResponse{}, err
	}

	resp := SeedFixtureResponse{
		OK:         true,
		Command:    "mechanism seed-fixture",
		Store:      dbPath,
		ProblemID:  input.ProblemID,
		RunID:      ingestRun.ID,
		SnapshotID: admission.Snapshot.ID,
		RevisionID: result.RevisionID,
	}
	for _, ref := range result.Approaches {
		resp.ApproachIDs = append(resp.ApproachIDs, ref.ApproachID)
		resp.MechanismIDs = append(resp.MechanismIDs, ref.MechanismID)
	}
	return resp, nil
}

func resolveFixture(input MechanismFixtureSeedInput) (MechanismFixture, error) {
	if input.Fixture != nil {
		return *input.Fixture, nil
	}
	if strings.TrimSpace(input.Path) == "" {
		return MechanismFixture{}, fmt.Errorf("mechanism fixture: one of Path or Fixture is required")
	}
	raw, err := os.ReadFile(input.Path)
	if err != nil {
		return MechanismFixture{}, fmt.Errorf("read mechanism fixture: %w", err)
	}
	var fixture MechanismFixture
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&fixture); err != nil {
		return MechanismFixture{}, fmt.Errorf("decode mechanism fixture %q: %w", input.Path, err)
	}
	return fixture, nil
}

func buildFixtureNormalizationInput(fixture MechanismFixture, input MechanismFixtureSeedInput, now time.Time) (store.NormalizationInput, error) {
	invocationID := domain.NewProviderInvocationID(now)
	revisionID := domain.NewNormalizationRevisionID(now)

	normInput := store.NormalizationInput{
		Invocation: domain.ProviderInvocation{
			ID:            invocationID,
			RunID:         input.RunID,
			Role:          domain.RoleNormalize,
			ProviderName:  "fixture",
			SchemaVersion: "normalize/v1",
			RequestHash:   "fixture",
			CreatedAt:     now,
		},
		Revision: domain.NormalizationRevision{
			ID:                   revisionID,
			ProblemID:            input.ProblemID,
			RunID:                input.RunID,
			SnapshotID:           input.SnapshotID,
			ProviderInvocationID: invocationID,
			SchemaVersion:        "normalize/v1",
			ConfigHash:           "fixture",
			Status:               domain.NormalizationStatusSucceeded,
			CreatedAt:            now,
		},
	}

	for i := range fixture.Approaches {
		approachInput, err := fixtureApproachToInput(fixture.Approaches[i], revisionID, input.SnapshotID, now)
		if err != nil {
			return store.NormalizationInput{}, fmt.Errorf("approach %d: %w", i, err)
		}
		normInput.Approaches = append(normInput.Approaches, approachInput)
	}
	return normInput, nil
}

func fixtureApproachToInput(fx MechanismFixtureApproach, revisionID, snapshotID string, now time.Time) (store.ApproachInput, error) {
	if strings.TrimSpace(fx.LogicalIdentity) == "" {
		return store.ApproachInput{}, fmt.Errorf("logical_identity is required")
	}
	approachRevisionID := domain.NewApproachRevisionID(now)
	mechanismID := domain.NewMechanismID(now)
	outcomeID := domain.NewOutcomeID(now)

	mechanism := domain.Mechanism{
		ID:                 mechanismID,
		ApproachRevisionID: approachRevisionID,
		Locality:           domain.Locality(fx.Locality),
		ConstructionMode:   domain.ConstructionMode(fx.ConstructionMode),
		UncertaintyMode:    domain.UncertaintyMode(fx.UncertaintyMode),
		Notes:              fx.Notes,
	}

	attributes := make([]domain.MechanismAttribute, 0)
	appendAttrs := func(kind domain.MechanismAttributeKind, values []string) {
		for _, v := range values {
			attributes = append(attributes, domain.MechanismAttribute{
				MechanismID: mechanismID,
				Kind:        kind,
				Value:       v,
			})
		}
	}
	appendAttrs(domain.AttrRepresentation, fx.Representations)
	appendAttrs(domain.AttrAssumption, fx.Assumptions)
	appendAttrs(domain.AttrOperator, fx.Operators)
	appendAttrs(domain.AttrPreserves, fx.Preserves)
	appendAttrs(domain.AttrBreaks, fx.Breaks)
	appendAttrs(domain.AttrAuxiliaryObject, fx.AuxiliaryObjects)

	boundaries := make([]domain.FailureBoundary, 0, len(fx.Boundaries))
	for _, condition := range fx.Boundaries {
		boundaries = append(boundaries, domain.FailureBoundary{
			ID:                 domain.NewFailureBoundaryID(now),
			ApproachRevisionID: approachRevisionID,
			Condition:          condition,
		})
	}

	support := make([]domain.SourceSupport, 0, len(fx.Support))
	for _, s := range fx.Support {
		support = append(support, domain.SourceSupport{
			ApproachRevisionID: approachRevisionID,
			SnapshotID:         snapshotID,
			FieldPath:          s.FieldPath,
			SupportKind:        domain.SupportKind(s.SupportKind),
			Locator:            s.Locator,
			Confidence:         s.Confidence,
		})
	}

	completeness := make([]domain.MechanismFieldCompleteness, 0, len(fx.CompleteFields))
	for _, kindStr := range fx.CompleteFields {
		kind := domain.MechanismAttributeKind(kindStr)
		if !kind.Valid() {
			return store.ApproachInput{}, fmt.Errorf("complete_fields: invalid field kind %q", kindStr)
		}
		completeness = append(completeness, domain.MechanismFieldCompleteness{
			MechanismID:    mechanismID,
			Kind:           kind,
			Completeness:   domain.CompletenessComplete,
			Scope:          domain.ScopeDeclaredPayload,
			Basis:          "authored fixture payload lists this field exhaustively",
			Admission:      domain.CompletenessAccepted,
			AdmissionBasis: "deterministic in-repo fixture payload (declared_payload scope)",
		})
	}

	return store.ApproachInput{
		LogicalIdentity: fx.LogicalIdentity,
		Revision: domain.ApproachRevision{
			ID:                      approachRevisionID,
			ApproachID:              domain.NewApproachID(now),
			NormalizationRevisionID: revisionID,
			Label:                   fallback(fx.Label, fx.LogicalIdentity),
			Description:             fx.Description,
			CreatedAt:               now,
		},
		Mechanism:  mechanism,
		Attributes: attributes,
		Outcome: domain.Outcome{
			ID:                 outcomeID,
			ApproachRevisionID: approachRevisionID,
			Class:              domain.OutcomeClass(fx.Outcome.Class),
			BoundaryStatement:  fx.Outcome.BoundaryStatement,
			Notes:              fx.Outcome.Notes,
		},
		Boundaries:        boundaries,
		Support:           support,
		FieldCompleteness: completeness,
	}, nil
}

func fallback(primary, secondary string) string {
	if strings.TrimSpace(primary) != "" {
		return primary
	}
	return secondary
}
