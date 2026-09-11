package pipeline

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/normalize"
	"github.com/instagrim-dev/newf/internal/provider"
	"github.com/instagrim-dev/newf/internal/store"
)

var (
	// ErrUnknownProvider marks a request for an unregistered normalization provider.
	ErrUnknownProvider = errors.New("unknown provider")
	// ErrNoEligibleSnapshots marks a normalize request with nothing to process.
	ErrNoEligibleSnapshots = errors.New("no eligible snapshots to normalize")
)

// NormalizeInput drives the normalize operator.
type NormalizeInput struct {
	DBPath        string
	ProblemID     string
	SourceOrSnap  string // source-id or snapshot-id; empty when All is set
	All           bool
	Provider      string
	Model         string
	SchemaVersion string
	Force         bool
	JSONOutput    bool
}

// ApproachListInput lists approaches for a problem.
type ApproachListInput struct {
	DBPath     string
	ProblemID  string
	JSONOutput bool
}

const defaultNormalizeProvider = provider.FixtureProviderName

// Normalize converts eligible immutable source snapshots into typed approach /
// mechanism revisions. It preserves the epistemic boundary between source
// material and generated interpretation: every persisted field records whether
// it is explicitly source-backed, model-inferred, or unsupported.
func (a *App) Normalize(ctx context.Context, input NormalizeInput) (NormalizeResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return NormalizeResponse{}, err
	}
	defer repoStore.Close()

	if _, err := repoStore.GetProblem(ctx, input.ProblemID); err != nil {
		return NormalizeResponse{}, err
	}

	providerName := input.Provider
	if providerName == "" {
		providerName = defaultNormalizeProvider
	}
	normalizer, ok := a.normalizers[providerName]
	if !ok {
		return NormalizeResponse{}, fmt.Errorf("%w: %q", ErrUnknownProvider, providerName)
	}

	schemaVersion := input.SchemaVersion
	if schemaVersion == "" {
		schemaVersion = normalize.SchemaVersion
	}

	snapshots, err := a.resolveNormalizeSnapshots(ctx, repoStore, input)
	if err != nil {
		return NormalizeResponse{}, err
	}

	now := a.now()
	run, err := repoStore.CreateRun(ctx, domain.NewRun{
		ID:          domain.NewRunID(now),
		ProblemID:   input.ProblemID,
		Operation:   "normalize",
		Status:      domain.RunStatusRunning,
		InputRef:    normalizeInputRef(input),
		ToolName:    "newf",
		ToolVersion: a.version,
		StartedAt:   now,
		CompletedAt: now,
	})
	if err != nil {
		return NormalizeResponse{}, err
	}

	objectsRoot := filepath.Join(filepath.Dir(dbPath), "objects")
	configHash := normalizeConfigHash(schemaVersion, providerName, input.Model)

	response := NormalizeResponse{
		OK:        true,
		Command:   "normalize",
		Store:     dbPath,
		ProblemID: input.ProblemID,
		RunID:     run.ID,
		Provider:  providerName,
		Schema:    schemaVersion,
	}

	for _, snapshot := range snapshots {
		result := a.normalizeOneSnapshot(ctx, repoStore, normalizer, snapshot, objectsRoot, run.ID, input, schemaVersion, configHash)
		response.Results = append(response.Results, result)
	}

	sort.Slice(response.Results, func(i, j int) bool {
		return response.Results[i].SnapshotID < response.Results[j].SnapshotID
	})

	var failures []string
	for _, result := range response.Results {
		if result.Status == "failed" {
			failures = append(failures, fmt.Sprintf("%s: %s", result.SnapshotID, result.Message))
		}
	}
	if err := a.finalizeRun(ctx, repoStore, run.ID, failures); err != nil {
		return NormalizeResponse{}, err
	}

	return response, nil
}

func (a *App) normalizeOneSnapshot(
	ctx context.Context,
	repoStore problemStore,
	normalizer provider.Normalizer,
	snapshot domain.SourceSnapshot,
	objectsRoot string,
	runID string,
	input NormalizeInput,
	schemaVersion string,
	configHash string,
) NormalizeResult {
	result := NormalizeResult{SnapshotID: snapshot.ID}

	if !input.Force {
		existing, findErr := repoStore.FindEquivalentNormalization(ctx, snapshot.ID, schemaVersion, configHash)
		if findErr != nil {
			result.Status = "failed"
			result.Reason = classifyNormalizeError(findErr)
			result.Message = findErr.Error()
			return result
		}
		if existing.Found {
			result.Status = "duplicate_existing"
			result.RevisionID = existing.Revision.ID
			if existing.Revision.SkipReason != nil {
				result.Reason = *existing.Revision.SkipReason
			}
			return result
		}
	}

	content, readErr := readSnapshotContent(objectsRoot, snapshot)
	if readErr != nil {
		result.Status = "failed"
		result.Reason = "source_snapshot_unreadable"
		result.Message = readErr.Error()
		return result
	}

	req := normalize.Request{
		ProblemID:     input.ProblemID,
		SnapshotID:    snapshot.ID,
		MediaType:     snapshot.MediaType,
		Content:       content,
		SchemaVersion: schemaVersion,
	}
	response, provErr := normalizer.Normalize(ctx, req)
	if provErr != nil {
		result.Status = "failed"
		result.Reason = classifyNormalizeError(provErr)
		result.Message = provErr.Error()
		return result
	}

	providerResult := response.Result
	if validationErr := providerResult.Validate(); validationErr != nil {
		result.Status = "failed"
		result.Reason = "schema_validation_failed"
		result.Message = validationErr.Error()
		return result
	}

	now := a.now()
	invocation := domain.ProviderInvocation{
		ID:              domain.NewProviderInvocationID(now),
		RunID:           runID,
		Role:            domain.RoleNormalize,
		ProviderName:    response.Metadata.ProviderName,
		ProviderVersion: response.Metadata.ProviderVersion,
		ModelName:       response.Metadata.ModelName,
		SchemaVersion:   schemaVersion,
		RequestHash:     hashString(response.RequestPayload),
		RequestPayload:  response.RequestPayload,
		ResponsePayload: response.ResponsePayload,
		CreatedAt:       now,
	}

	var supersedes *string
	if input.Force {
		if latest, found, latestErr := repoStore.LatestNormalizationForSnapshot(ctx, snapshot.ID); latestErr == nil && found {
			superseded := latest.ID
			supersedes = &superseded
		}
	}

	if providerResult.Skipped {
		skipReason := string(providerResult.SkipReason)
		revision := domain.NormalizationRevision{
			ID:                   domain.NewNormalizationRevisionID(now),
			ProblemID:            input.ProblemID,
			RunID:                runID,
			SnapshotID:           snapshot.ID,
			ProviderInvocationID: invocation.ID,
			SchemaVersion:        schemaVersion,
			ConfigHash:           configHash,
			Status:               domain.NormalizationStatusSkipped,
			SupersedesRevisionID: supersedes,
			SkipReason:           &skipReason,
			CreatedAt:            now,
		}
		writeResult, writeErr := repoStore.PersistNormalization(ctx, store.NormalizationInput{
			Revision:   revision,
			Invocation: invocation,
		})
		if writeErr != nil {
			result.Status = "failed"
			result.Reason = "persistence_failed"
			result.Message = writeErr.Error()
			return result
		}
		result.Status = "skipped"
		result.Reason = skipReason
		result.RevisionID = writeResult.RevisionID
		result.Warnings = providerResult.Warnings
		return result
	}

	revision := domain.NormalizationRevision{
		ID:                   domain.NewNormalizationRevisionID(now),
		ProblemID:            input.ProblemID,
		RunID:                runID,
		SnapshotID:           snapshot.ID,
		ProviderInvocationID: invocation.ID,
		SchemaVersion:        schemaVersion,
		ConfigHash:           configHash,
		Status:               domain.NormalizationStatusSucceeded,
		SupersedesRevisionID: supersedes,
		CreatedAt:            now,
	}

	approaches, buildErr := buildApproachInputs(revision, snapshot.ID, providerResult, now)
	if buildErr != nil {
		result.Status = "failed"
		result.Reason = "schema_validation_failed"
		result.Message = buildErr.Error()
		return result
	}

	writeResult, writeErr := repoStore.PersistNormalization(ctx, store.NormalizationInput{
		Revision:   revision,
		Invocation: invocation,
		Approaches: approaches,
	})
	if writeErr != nil {
		result.Status = "failed"
		result.Reason = "persistence_failed"
		result.Message = writeErr.Error()
		return result
	}

	result.Status = "created"
	result.RevisionID = writeResult.RevisionID
	result.Warnings = providerResult.Warnings
	for _, ref := range writeResult.Approaches {
		result.Approaches = append(result.Approaches, NormalizeApproachResult{
			ApproachID:      ref.ApproachID,
			RevisionID:      ref.ApproachRevisionID,
			MechanismID:     ref.MechanismID,
			CreatedApproach: ref.CreatedApproach,
		})
	}
	return result
}

func buildApproachInputs(revision domain.NormalizationRevision, snapshotID string, result normalize.Result, now time.Time) ([]store.ApproachInput, error) {
	inputs := make([]store.ApproachInput, 0, len(result.Approaches))
	for _, approach := range result.Approaches {
		approachRevisionID := domain.NewApproachRevisionID(now)
		mechanismID := domain.NewMechanismID(now)
		outcomeID := domain.NewOutcomeID(now)

		mechanism := domain.Mechanism{
			ID:                 mechanismID,
			ApproachRevisionID: approachRevisionID,
			Locality:           approach.Mechanism.Locality,
			ConstructionMode:   approach.Mechanism.ConstructionMode,
			UncertaintyMode:    approach.Mechanism.UncertaintyMode,
			Notes:              approach.Mechanism.Notes,
		}

		attributes := buildAttributes(mechanismID, approach.Mechanism)

		// Justified completeness declarations (v25): schema validation already
		// enforced valid keys/values + a non-empty basis; sort for determinism.
		var completeness []domain.MechanismFieldCompleteness
		if len(approach.Mechanism.FieldCompleteness) > 0 {
			kinds := make([]string, 0, len(approach.Mechanism.FieldCompleteness))
			for kind := range approach.Mechanism.FieldCompleteness {
				kinds = append(kinds, kind)
			}
			sort.Strings(kinds)
			for _, kind := range kinds {
				completeness = append(completeness, domain.MechanismFieldCompleteness{
					MechanismID:  mechanismID,
					Kind:         domain.MechanismAttributeKind(kind),
					Completeness: domain.FieldCompleteness(approach.Mechanism.FieldCompleteness[kind]),
					Basis:        approach.Mechanism.CompletenessBasis,
				})
			}
		}

		boundaries := make([]domain.FailureBoundary, 0, len(approach.Outcome.BoundaryConditions))
		for _, condition := range approach.Outcome.BoundaryConditions {
			boundaries = append(boundaries, domain.FailureBoundary{
				ID:                 domain.NewFailureBoundaryID(now),
				ApproachRevisionID: approachRevisionID,
				Condition:          condition,
			})
		}

		support := make([]domain.SourceSupport, 0, len(approach.Support))
		for _, field := range approach.Support {
			support = append(support, domain.SourceSupport{
				ApproachRevisionID: approachRevisionID,
				SnapshotID:         snapshotID,
				FieldPath:          field.FieldPath,
				SupportKind:        field.SupportKind,
				Locator:            field.Locator,
				Confidence:         field.Confidence,
			})
		}

		inputs = append(inputs, store.ApproachInput{
			LogicalIdentity: approach.LogicalIdentity,
			Revision: domain.ApproachRevision{
				ID:                      approachRevisionID,
				ApproachID:              domain.NewApproachID(now), // placeholder; store resolves real identity
				NormalizationRevisionID: revision.ID,
				Label:                   approach.Label,
				Description:             approach.Description,
				CreatedAt:               now,
			},
			Mechanism:         mechanism,
			Attributes:        attributes,
			FieldCompleteness: completeness,
			Outcome: domain.Outcome{
				ID:                 outcomeID,
				ApproachRevisionID: approachRevisionID,
				Class:              approach.Outcome.Class,
				BoundaryStatement:  approach.Outcome.BoundaryStatement,
				Notes:              approach.Outcome.Notes,
			},
			Boundaries: boundaries,
			Support:    support,
		})
	}
	return inputs, nil
}

func buildAttributes(mechanismID string, mechanism normalize.Mechanism) []domain.MechanismAttribute {
	var attributes []domain.MechanismAttribute
	appendKind := func(kind domain.MechanismAttributeKind, values []string) {
		for _, value := range values {
			attributes = append(attributes, domain.MechanismAttribute{
				MechanismID: mechanismID,
				Kind:        kind,
				Value:       value,
			})
		}
	}
	appendKind(domain.AttrRepresentation, mechanism.Representations)
	appendKind(domain.AttrAssumption, mechanism.Assumptions)
	appendKind(domain.AttrOperator, mechanism.Operators)
	appendKind(domain.AttrPreserves, mechanism.Preserves)
	appendKind(domain.AttrBreaks, mechanism.Breaks)
	appendKind(domain.AttrAuxiliaryObject, mechanism.AuxiliaryObjects)
	return attributes
}

func (a *App) resolveNormalizeSnapshots(ctx context.Context, repoStore problemStore, input NormalizeInput) ([]domain.SourceSnapshot, error) {
	if input.All {
		sources, err := repoStore.ListSourcesByProblem(ctx, input.ProblemID)
		if err != nil {
			return nil, err
		}
		var snapshots []domain.SourceSnapshot
		seen := map[string]struct{}{}
		for _, source := range sources {
			sourceSnapshots, snapErr := repoStore.ListSourceSnapshots(ctx, source.ID)
			if snapErr != nil {
				return nil, snapErr
			}
			// Latest snapshot per source is eligible; ListSourceSnapshots is
			// newest-first.
			if len(sourceSnapshots) > 0 {
				latest := sourceSnapshots[0]
				if _, dup := seen[latest.ID]; !dup {
					seen[latest.ID] = struct{}{}
					snapshots = append(snapshots, latest)
				}
			}
		}
		if len(snapshots) == 0 {
			return nil, ErrNoEligibleSnapshots
		}
		return snapshots, nil
	}

	ref := input.SourceOrSnap
	if ref == "" {
		return nil, errors.New("--source <source-id|snapshot-id> is required unless --all is set")
	}

	if snapErr := domain.ValidateSnapshotID(ref); snapErr == nil {
		snapshot, err := repoStore.GetSourceSnapshot(ctx, ref)
		if err != nil {
			return nil, err
		}
		if err := a.assertSnapshotBelongsToProblem(ctx, repoStore, snapshot, input.ProblemID); err != nil {
			return nil, err
		}
		return []domain.SourceSnapshot{snapshot}, nil
	}

	if srcErr := domain.ValidateSourceID(ref); srcErr == nil {
		source, err := repoStore.GetSource(ctx, ref)
		if err != nil {
			return nil, err
		}
		if source.ProblemID != input.ProblemID {
			return nil, fmt.Errorf("%w: source %s does not belong to problem %s", store.ErrNotFound, ref, input.ProblemID)
		}
		snapshots, err := repoStore.ListSourceSnapshots(ctx, source.ID)
		if err != nil {
			return nil, err
		}
		if len(snapshots) == 0 {
			return nil, ErrNoEligibleSnapshots
		}
		return []domain.SourceSnapshot{snapshots[0]}, nil
	}

	return nil, fmt.Errorf("%w: %q is not a source or snapshot id", domain.ErrInvalidSnapshotID, ref)
}

func (a *App) assertSnapshotBelongsToProblem(ctx context.Context, repoStore problemStore, snapshot domain.SourceSnapshot, problemID string) error {
	source, err := repoStore.GetSource(ctx, snapshot.SourceID)
	if err != nil {
		return err
	}
	if source.ProblemID != problemID {
		return fmt.Errorf("%w: snapshot %s does not belong to problem %s", store.ErrNotFound, snapshot.ID, problemID)
	}
	return nil
}

func readSnapshotContent(objectsRoot string, snapshot domain.SourceSnapshot) (string, error) {
	absPath := filepath.Join(objectsRoot, filepath.FromSlash(snapshot.ObjectPath))
	raw, err := os.ReadFile(absPath)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func normalizeConfigHash(schemaVersion, providerName, model string) string {
	return hashString(fmt.Sprintf("schema=%s|provider=%s|model=%s", schemaVersion, providerName, model))
}

func hashString(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func normalizeInputRef(input NormalizeInput) string {
	if input.All {
		return "normalize:all:" + input.ProblemID
	}
	return "normalize:" + input.SourceOrSnap
}

func classifyNormalizeError(err error) string {
	switch {
	case errors.Is(err, normalize.ErrSchemaViolation):
		return "schema_validation_failed"
	case errors.Is(err, provider.ErrProviderTransport):
		return "provider_unavailable"
	default:
		return "internal_error"
	}
}

// ListApproaches returns the normalized approaches for a problem.
func (a *App) ListApproaches(ctx context.Context, input ApproachListInput) (ApproachListResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return ApproachListResponse{}, err
	}
	defer repoStore.Close()

	if _, err := repoStore.GetProblem(ctx, input.ProblemID); err != nil {
		return ApproachListResponse{}, err
	}

	items, err := repoStore.ListApproaches(ctx, input.ProblemID)
	if err != nil {
		return ApproachListResponse{}, err
	}

	views := make([]ApproachListView, 0, len(items))
	for _, item := range items {
		views = append(views, ApproachListView{
			ID:              item.Approach.ID,
			ProblemID:       item.Approach.ProblemID,
			LogicalIdentity: item.Approach.LogicalIdentity,
			Label:           item.LatestRevision.Label,
			OutcomeClass:    item.OutcomeClass,
			SnapshotID:      item.SnapshotID,
			RevisionCount:   item.RevisionCount,
			LatestRevision:  item.LatestRevision.ID,
		})
	}

	return ApproachListResponse{
		OK:         true,
		Command:    "approach list",
		Store:      dbPath,
		ProblemID:  input.ProblemID,
		Approaches: views,
	}, nil
}

// ShowApproach returns the latest revision detail for one approach.
func (a *App) ShowApproach(ctx context.Context, input LookupInput) (ApproachShowResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return ApproachShowResponse{}, err
	}
	defer repoStore.Close()

	detail, err := repoStore.GetApproachDetail(ctx, input.ID)
	if err != nil {
		return ApproachShowResponse{}, err
	}

	return ApproachShowResponse{
		OK:                    true,
		Command:               "approach show",
		Store:                 dbPath,
		ApproachID:            detail.Approach.ID,
		ProblemID:             detail.Approach.ProblemID,
		LogicalIdentity:       detail.Approach.LogicalIdentity,
		Revision:              approachRevisionView(detail.Revision),
		NormalizationRevision: detail.Normalization.ID,
		SnapshotID:            detail.SnapshotID,
		RunID:                 detail.Normalization.RunID,
		Provider:              providerInvocationView(detail.Invocation),
		Mechanism:             mechanismView(detail.Mechanism, detail.Attributes),
		Outcome:               outcomeView(detail.Outcome, detail.Boundaries),
		Support:               fieldSupportViews(detail.Support),
		RevisionCount:         detail.RevisionCount,
	}, nil
}

// ShowApproachRevisions returns all revisions for one approach.
func (a *App) ShowApproachRevisions(ctx context.Context, input LookupInput) (ApproachRevisionsResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return ApproachRevisionsResponse{}, err
	}
	defer repoStore.Close()

	approach, revisions, err := repoStore.ListApproachRevisions(ctx, input.ID)
	if err != nil {
		return ApproachRevisionsResponse{}, err
	}

	views := make([]ApproachRevisionView, 0, len(revisions))
	for _, revision := range revisions {
		views = append(views, approachRevisionView(revision))
	}

	return ApproachRevisionsResponse{
		OK:         true,
		Command:    "approach revisions",
		Store:      dbPath,
		ApproachID: approach.ID,
		ProblemID:  approach.ProblemID,
		Revisions:  views,
	}, nil
}

// ShowMechanism returns the mechanism detail for one mechanism id.
func (a *App) ShowMechanism(ctx context.Context, input LookupInput) (MechanismShowResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return MechanismShowResponse{}, err
	}
	defer repoStore.Close()

	detail, err := repoStore.GetMechanismDetail(ctx, input.ID)
	if err != nil {
		return MechanismShowResponse{}, err
	}

	return MechanismShowResponse{
		OK:          true,
		Command:     "mechanism show",
		Store:       dbPath,
		MechanismID: detail.Mechanism.ID,
		ApproachID:  detail.Approach.ID,
		RevisionID:  detail.Revision.ID,
		Mechanism:   mechanismView(detail.Mechanism, detail.Attributes),
		Outcome:     outcomeView(detail.Outcome, detail.Boundaries),
	}, nil
}

func approachRevisionView(revision domain.ApproachRevision) ApproachRevisionView {
	return ApproachRevisionView{
		ID:                      revision.ID,
		ApproachID:              revision.ApproachID,
		NormalizationRevisionID: revision.NormalizationRevisionID,
		Label:                   revision.Label,
		Description:             revision.Description,
		SupersedesRevisionID:    revision.SupersedesRevisionID,
		CreatedAt:               revision.CreatedAt.Format(timeLayout),
	}
}

func providerInvocationView(invocation domain.ProviderInvocation) ProviderInvocationView {
	return ProviderInvocationView{
		ID:              invocation.ID,
		Role:            string(invocation.Role),
		ProviderName:    invocation.ProviderName,
		ProviderVersion: invocation.ProviderVersion,
		ModelName:       invocation.ModelName,
		SchemaVersion:   invocation.SchemaVersion,
		RequestHash:     invocation.RequestHash,
	}
}

func mechanismView(mechanism domain.Mechanism, attributes []domain.MechanismAttribute) MechanismView {
	view := MechanismView{
		ID:               mechanism.ID,
		ApproachRevision: mechanism.ApproachRevisionID,
		Locality:         string(mechanism.Locality),
		ConstructionMode: string(mechanism.ConstructionMode),
		UncertaintyMode:  string(mechanism.UncertaintyMode),
		Notes:            mechanism.Notes,
	}
	for _, attr := range attributes {
		switch attr.Kind {
		case domain.AttrRepresentation:
			view.Representations = append(view.Representations, attr.Value)
		case domain.AttrAssumption:
			view.Assumptions = append(view.Assumptions, attr.Value)
		case domain.AttrOperator:
			view.Operators = append(view.Operators, attr.Value)
		case domain.AttrPreserves:
			view.Preserves = append(view.Preserves, attr.Value)
		case domain.AttrBreaks:
			view.Breaks = append(view.Breaks, attr.Value)
		case domain.AttrAuxiliaryObject:
			view.AuxiliaryObjects = append(view.AuxiliaryObjects, attr.Value)
		}
	}
	return view
}

func outcomeView(outcome domain.Outcome, boundaries []domain.FailureBoundary) OutcomeView {
	view := OutcomeView{
		Class:             string(outcome.Class),
		BoundaryStatement: outcome.BoundaryStatement,
		Notes:             outcome.Notes,
	}
	for _, boundary := range boundaries {
		view.BoundaryConditions = append(view.BoundaryConditions, boundary.Condition)
	}
	return view
}

func fieldSupportViews(supports []domain.SourceSupport) []FieldSupportView {
	views := make([]FieldSupportView, 0, len(supports))
	for _, support := range supports {
		views = append(views, FieldSupportView{
			FieldPath:   support.FieldPath,
			SupportKind: string(support.SupportKind),
			Locator:     support.Locator,
			Confidence:  support.Confidence,
		})
	}
	return views
}
