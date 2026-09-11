package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/instagrim-dev/newf/internal/domain"
)

// NormalizationInput is the complete, validated set of records for one
// normalization revision. It is written in a single transaction so a failure
// after provider success never leaves a half-created revision.
type NormalizationInput struct {
	Revision   domain.NormalizationRevision
	Invocation domain.ProviderInvocation
	Approaches []ApproachInput
}

// ApproachInput bundles one approach's logical identity with the derived
// interpretation records for this revision.
type ApproachInput struct {
	LogicalIdentity string
	Revision        domain.ApproachRevision
	Mechanism       domain.Mechanism
	Attributes      []domain.MechanismAttribute
	// FieldCompleteness carries the extractor's JUSTIFIED per-field
	// exhaustiveness declarations (v25). Optional; undeclared fields keep the
	// conservative unobserved default at signature build time.
	FieldCompleteness []domain.MechanismFieldCompleteness
	Outcome           domain.Outcome
	Boundaries        []domain.FailureBoundary
	Support           []domain.SourceSupport
}

// ApproachRef is the created identity trio for one approach in a revision.
type ApproachRef struct {
	ApproachID         string
	ApproachRevisionID string
	MechanismID        string
	CreatedApproach    bool
}

// NormalizationWriteResult reports the persisted IDs for a normalization write.
type NormalizationWriteResult struct {
	RevisionID string
	Approaches []ApproachRef
}

// ExistingNormalization is returned when an equivalent successful revision
// already exists for the same idempotency fingerprint.
type ExistingNormalization struct {
	Revision domain.NormalizationRevision
	Found    bool
}

// FindEquivalentNormalization looks up a prior successful (or skipped) revision
// for the same snapshot + schema version + config hash. This is the default
// idempotency key: identical source snapshot, normalization schema version, and
// provider/model/contract configuration.
func (s *Store) FindEquivalentNormalization(ctx context.Context, snapshotID, schemaVersion, configHash string) (ExistingNormalization, error) {
	if err := domain.ValidateSnapshotID(snapshotID); err != nil {
		return ExistingNormalization{}, err
	}
	row := s.db.QueryRowContext(ctx, `
SELECT id, problem_id, run_id, snapshot_id, provider_invocation_id, schema_version, config_hash, status, supersedes_revision_id, skip_reason, created_at
FROM normalization_revisions
WHERE snapshot_id = ? AND schema_version = ? AND config_hash = ? AND status IN ('succeeded', 'skipped')
ORDER BY created_at DESC, id DESC
LIMIT 1
`, snapshotID, schemaVersion, configHash)
	revision, err := scanNormalizationRevision(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ExistingNormalization{}, nil
		}
		return ExistingNormalization{}, err
	}
	return ExistingNormalization{Revision: revision, Found: true}, nil
}

// LatestNormalizationForSnapshot returns the most recent revision for a
// snapshot, used to set supersedes lineage on a forced re-normalization.
func (s *Store) LatestNormalizationForSnapshot(ctx context.Context, snapshotID string) (domain.NormalizationRevision, bool, error) {
	if err := domain.ValidateSnapshotID(snapshotID); err != nil {
		return domain.NormalizationRevision{}, false, err
	}
	row := s.db.QueryRowContext(ctx, `
SELECT id, problem_id, run_id, snapshot_id, provider_invocation_id, schema_version, config_hash, status, supersedes_revision_id, skip_reason, created_at
FROM normalization_revisions
WHERE snapshot_id = ?
ORDER BY created_at DESC, id DESC
LIMIT 1
`, snapshotID)
	revision, err := scanNormalizationRevision(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.NormalizationRevision{}, false, nil
		}
		return domain.NormalizationRevision{}, false, err
	}
	return revision, true, nil
}

// PersistNormalization writes a full normalization revision transactionally.
// It validates every record before writing so invalid provenance is rejected
// by construction, and it reuses an existing Approach identity (never
// overwriting it) when one already exists for the problem.
func (s *Store) PersistNormalization(ctx context.Context, input NormalizationInput) (NormalizationWriteResult, error) {
	if err := input.Invocation.Validate(); err != nil {
		return NormalizationWriteResult{}, err
	}
	if err := input.Revision.Validate(); err != nil {
		return NormalizationWriteResult{}, err
	}
	if input.Revision.ProviderInvocationID != input.Invocation.ID {
		return NormalizationWriteResult{}, errors.New("revision provider_invocation_id must match invocation id")
	}
	for i := range input.Approaches {
		if err := validateApproachInput(input.Approaches[i], input.Revision.ID); err != nil {
			return NormalizationWriteResult{}, err
		}
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return NormalizationWriteResult{}, err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
INSERT INTO provider_invocations(id, run_id, role, provider_name, provider_version, model_name, schema_version, request_hash, request_payload, response_payload, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, input.Invocation.ID, input.Invocation.RunID, string(input.Invocation.Role), input.Invocation.ProviderName, input.Invocation.ProviderVersion, input.Invocation.ModelName, input.Invocation.SchemaVersion, input.Invocation.RequestHash, input.Invocation.RequestPayload, input.Invocation.ResponsePayload, formatTime(input.Invocation.CreatedAt)); err != nil {
		return NormalizationWriteResult{}, err
	}

	if _, err := tx.ExecContext(ctx, `
INSERT INTO normalization_revisions(id, problem_id, run_id, snapshot_id, provider_invocation_id, schema_version, config_hash, status, supersedes_revision_id, skip_reason, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, input.Revision.ID, input.Revision.ProblemID, input.Revision.RunID, input.Revision.SnapshotID, input.Revision.ProviderInvocationID, input.Revision.SchemaVersion, input.Revision.ConfigHash, string(input.Revision.Status), input.Revision.SupersedesRevisionID, input.Revision.SkipReason, formatTime(input.Revision.CreatedAt)); err != nil {
		return NormalizationWriteResult{}, err
	}

	refs := make([]ApproachRef, 0, len(input.Approaches))
	for i := range input.Approaches {
		ref, err := persistApproachTx(ctx, tx, input.Revision.ProblemID, input.Approaches[i])
		if err != nil {
			return NormalizationWriteResult{}, err
		}
		refs = append(refs, ref)
	}

	if err := tx.Commit(); err != nil {
		return NormalizationWriteResult{}, err
	}
	return NormalizationWriteResult{RevisionID: input.Revision.ID, Approaches: refs}, nil
}

func validateApproachInput(input ApproachInput, revisionID string) error {
	if input.LogicalIdentity == "" {
		return errors.New("approach logical_identity is required")
	}
	if err := input.Revision.Validate(); err != nil {
		return err
	}
	if input.Revision.NormalizationRevisionID != revisionID {
		return errors.New("approach revision normalization_revision_id mismatch")
	}
	if err := input.Mechanism.Validate(); err != nil {
		return err
	}
	if input.Mechanism.ApproachRevisionID != input.Revision.ID {
		return errors.New("mechanism approach_revision_id mismatch")
	}
	if err := input.Outcome.Validate(); err != nil {
		return err
	}
	if input.Outcome.ApproachRevisionID != input.Revision.ID {
		return errors.New("outcome approach_revision_id mismatch")
	}
	for _, attr := range input.Attributes {
		if !attr.Kind.Valid() {
			return fmt.Errorf("invalid mechanism attribute kind %q", attr.Kind)
		}
	}
	for _, boundary := range input.Boundaries {
		if err := boundary.Validate(); err != nil {
			return err
		}
	}
	for _, support := range input.Support {
		if err := support.Validate(); err != nil {
			return err
		}
		if support.ApproachRevisionID != input.Revision.ID {
			return errors.New("source support approach_revision_id mismatch")
		}
	}
	return nil
}

func persistApproachTx(ctx context.Context, tx *sql.Tx, problemID string, input ApproachInput) (ApproachRef, error) {
	approachID, createdApproach, err := getOrCreateApproachTx(ctx, tx, problemID, input.LogicalIdentity, input.Revision.CreatedAt)
	if err != nil {
		return ApproachRef{}, err
	}

	// Derive approach-revision lineage here, inside the transaction, because
	// the approach identity is only resolved above (get-or-create). The caller
	// never knows the approach id in advance, so it cannot supply the prior
	// revision link. An existing approach means this new revision supersedes
	// its immediately prior revision; a freshly created approach starts a new
	// chain. A caller-supplied SupersedesRevisionID (rare, e.g. explicit
	// backfill) is honored when set.
	supersedes := input.Revision.SupersedesRevisionID
	if supersedes == nil && !createdApproach {
		prior, found, priorErr := latestApproachRevisionIDTx(ctx, tx, approachID)
		if priorErr != nil {
			return ApproachRef{}, priorErr
		}
		if found {
			supersedes = &prior
		}
	}

	if _, err := tx.ExecContext(ctx, `
INSERT INTO approach_revisions(id, approach_id, normalization_revision_id, label, description, supersedes_revision_id, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?)
`, input.Revision.ID, approachID, input.Revision.NormalizationRevisionID, input.Revision.Label, input.Revision.Description, supersedes, formatTime(input.Revision.CreatedAt)); err != nil {
		return ApproachRef{}, err
	}

	if _, err := tx.ExecContext(ctx, `
INSERT INTO mechanisms(id, approach_revision_id, locality, construction_mode, uncertainty_mode, notes)
VALUES(?, ?, ?, ?, ?, ?)
`, input.Mechanism.ID, input.Revision.ID, string(input.Mechanism.Locality), string(input.Mechanism.ConstructionMode), string(input.Mechanism.UncertaintyMode), input.Mechanism.Notes); err != nil {
		return ApproachRef{}, err
	}

	for ordinal, attr := range input.Attributes {
		// Mechanism attributes are set-valued per (mechanism, kind): a repeated
		// value carries no additional information, so a duplicate is
		// intentionally ignored. This is distinct from source_supports below,
		// where a conflicting row is a hard error (see schema validation).
		if _, err := tx.ExecContext(ctx, `
INSERT OR IGNORE INTO mechanism_attributes(mechanism_id, kind, value, ordinal)
VALUES(?, ?, ?, ?)
`, input.Mechanism.ID, string(attr.Kind), attr.Value, ordinal); err != nil {
			return ApproachRef{}, err
		}
	}

	for _, fc := range input.FieldCompleteness {
		// Domain-validate before write: a declaration without a basis (or with
		// a vacuous 'unobserved' value) must never reach the table.
		declared := fc
		declared.MechanismID = input.Mechanism.ID
		if err := declared.Validate(); err != nil {
			return ApproachRef{}, fmt.Errorf("field completeness declaration: %w", err)
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO mechanism_field_completeness(mechanism_id, field_kind, completeness, basis)
VALUES(?, ?, ?, ?)
`, declared.MechanismID, string(declared.Kind), string(declared.Completeness), declared.Basis); err != nil {
			return ApproachRef{}, err
		}
	}

	if _, err := tx.ExecContext(ctx, `
INSERT INTO outcomes(id, approach_revision_id, class, boundary_statement, notes)
VALUES(?, ?, ?, ?, ?)
`, input.Outcome.ID, input.Revision.ID, string(input.Outcome.Class), input.Outcome.BoundaryStatement, input.Outcome.Notes); err != nil {
		return ApproachRef{}, err
	}

	for ordinal, boundary := range input.Boundaries {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO failure_boundaries(id, approach_revision_id, condition, ordinal)
VALUES(?, ?, ?, ?)
`, boundary.ID, input.Revision.ID, boundary.Condition, ordinal); err != nil {
			return ApproachRef{}, err
		}
	}

	for _, support := range input.Support {
		// Plain INSERT (not OR IGNORE): a second support row for the same
		// (approach_revision, snapshot, field_path) is an epistemic conflict,
		// not a benign duplicate. Schema validation rejects duplicate field
		// paths upstream; this is defense in depth so a differing support_kind
		// can never be silently dropped by the persistence layer.
		if _, err := tx.ExecContext(ctx, `
INSERT INTO source_supports(approach_revision_id, snapshot_id, field_path, support_kind, locator, confidence)
VALUES(?, ?, ?, ?, ?, ?)
`, input.Revision.ID, support.SnapshotID, support.FieldPath, string(support.SupportKind), support.Locator, support.Confidence); err != nil {
			return ApproachRef{}, err
		}
	}

	return ApproachRef{
		ApproachID:         approachID,
		ApproachRevisionID: input.Revision.ID,
		MechanismID:        input.Mechanism.ID,
		CreatedApproach:    createdApproach,
	}, nil
}

func getOrCreateApproachTx(ctx context.Context, tx *sql.Tx, problemID, logicalIdentity string, createdAt time.Time) (string, bool, error) {
	row := tx.QueryRowContext(ctx, `
SELECT id FROM approaches WHERE problem_id = ? AND logical_identity = ?
`, problemID, logicalIdentity)
	var existingID string
	err := row.Scan(&existingID)
	if err == nil {
		return existingID, false, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", false, err
	}

	approach := domain.Approach{
		ID:              domain.NewApproachID(createdAt),
		ProblemID:       problemID,
		LogicalIdentity: logicalIdentity,
		CreatedAt:       createdAt,
	}
	if err := approach.Validate(); err != nil {
		return "", false, err
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO approaches(id, problem_id, logical_identity, created_at)
VALUES(?, ?, ?, ?)
`, approach.ID, approach.ProblemID, approach.LogicalIdentity, formatTime(approach.CreatedAt)); err != nil {
		if isDuplicateApproachIdentityError(err) {
			row := tx.QueryRowContext(ctx, `
SELECT id FROM approaches WHERE problem_id = ? AND logical_identity = ?
`, problemID, logicalIdentity)
			var raceID string
			if scanErr := row.Scan(&raceID); scanErr != nil {
				return "", false, scanErr
			}
			return raceID, false, nil
		}
		return "", false, err
	}
	return approach.ID, true, nil
}

// latestApproachRevisionIDTx returns the id of the most recent approach
// revision for an approach, scoped to the current transaction so a
// re-normalization in flight links to the revision it is about to supersede.
// Ordering matches listApproachRevisions (created_at DESC, id DESC).
func latestApproachRevisionIDTx(ctx context.Context, tx *sql.Tx, approachID string) (string, bool, error) {
	row := tx.QueryRowContext(ctx, `
SELECT id FROM approach_revisions
WHERE approach_id = ?
ORDER BY created_at DESC, id DESC
LIMIT 1
`, approachID)
	var id string
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}
		return "", false, err
	}
	return id, true, nil
}

func isDuplicateApproachIdentityError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed: approaches.problem_id, approaches.logical_identity")
}

func scanNormalizationRevision(row scanner) (domain.NormalizationRevision, error) {
	var revision domain.NormalizationRevision
	var status string
	var supersedes sql.NullString
	var skipReason sql.NullString
	var createdAt string
	if err := row.Scan(&revision.ID, &revision.ProblemID, &revision.RunID, &revision.SnapshotID, &revision.ProviderInvocationID, &revision.SchemaVersion, &revision.ConfigHash, &status, &supersedes, &skipReason, &createdAt); err != nil {
		return domain.NormalizationRevision{}, err
	}
	parsedCreatedAt, err := parseTime(createdAt)
	if err != nil {
		return domain.NormalizationRevision{}, fmt.Errorf("%w: %v", ErrCorruptStore, err)
	}
	revision.Status = domain.NormalizationStatus(status)
	revision.CreatedAt = parsedCreatedAt
	if supersedes.Valid {
		revision.SupersedesRevisionID = &supersedes.String
	}
	if skipReason.Valid {
		revision.SkipReason = &skipReason.String
	}
	return revision, nil
}

// ApproachListItem summarizes one approach and its latest revision for listing.
type ApproachListItem struct {
	Approach       domain.Approach
	LatestRevision domain.ApproachRevision
	OutcomeClass   string
	SnapshotID     string
	RevisionCount  int
}

// ListApproaches returns approaches for a problem with their latest revision.
func (s *Store) ListApproaches(ctx context.Context, problemID string) ([]ApproachListItem, error) {
	if err := domain.ValidateProblemID(problemID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT id, problem_id, logical_identity, created_at
FROM approaches
WHERE problem_id = ?
ORDER BY created_at ASC, id ASC
`, problemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var approaches []domain.Approach
	for rows.Next() {
		var approach domain.Approach
		var createdAt string
		if err := rows.Scan(&approach.ID, &approach.ProblemID, &approach.LogicalIdentity, &createdAt); err != nil {
			return nil, err
		}
		parsed, parseErr := parseTime(createdAt)
		if parseErr != nil {
			return nil, fmt.Errorf("%w: %v", ErrCorruptStore, parseErr)
		}
		approach.CreatedAt = parsed
		approaches = append(approaches, approach)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	items := make([]ApproachListItem, 0, len(approaches))
	for _, approach := range approaches {
		revisions, err := s.listApproachRevisions(ctx, approach.ID)
		if err != nil {
			return nil, err
		}
		item := ApproachListItem{Approach: approach, RevisionCount: len(revisions)}
		if len(revisions) > 0 {
			latest := revisions[0]
			item.LatestRevision = latest
			outcome, _, outErr := s.getOutcome(ctx, latest.ID)
			if outErr != nil {
				return nil, outErr
			}
			item.OutcomeClass = string(outcome.Class)
			snapshotID, snapErr := s.snapshotForNormalizationRevision(ctx, latest.NormalizationRevisionID)
			if snapErr != nil {
				return nil, snapErr
			}
			item.SnapshotID = snapshotID
		}
		items = append(items, item)
	}
	return items, nil
}

// ApproachDetail is the full read model for one approach revision.
type ApproachDetail struct {
	Approach      domain.Approach
	Revision      domain.ApproachRevision
	Normalization domain.NormalizationRevision
	Invocation    domain.ProviderInvocation
	SnapshotID    string
	Mechanism     domain.Mechanism
	Attributes    []domain.MechanismAttribute
	// FieldCompleteness is the mechanism's justified per-field exhaustiveness
	// declarations (v25); empty when the extractor declared none.
	FieldCompleteness []domain.MechanismFieldCompleteness
	Outcome           domain.Outcome
	Boundaries        []domain.FailureBoundary
	Support           []domain.SourceSupport
	RevisionCount     int
}

// GetApproachDetail loads the latest revision detail for an approach.
func (s *Store) GetApproachDetail(ctx context.Context, approachID string) (ApproachDetail, error) {
	if err := domain.ValidateApproachID(approachID); err != nil {
		return ApproachDetail{}, err
	}
	approach, err := s.getApproach(ctx, approachID)
	if err != nil {
		return ApproachDetail{}, err
	}
	revisions, err := s.listApproachRevisions(ctx, approachID)
	if err != nil {
		return ApproachDetail{}, err
	}
	if len(revisions) == 0 {
		return ApproachDetail{}, fmt.Errorf("%w: approach %s has no revisions", ErrNotFound, approachID)
	}
	return s.approachDetailForRevision(ctx, approach, revisions[0], len(revisions))
}

// MechanismListItem is one mechanism summary row for problem-level discovery
// (E2: the provenance chain must be walkable without raw SQL).
type MechanismListItem struct {
	MechanismID        string
	ApproachID         string
	ApproachRevisionID string
	Label              string
	LogicalIdentity    string
	OutcomeClass       string
	SignatureCount     int
}

// ListMechanismsForProblem returns every persisted mechanism for a problem
// (via approach-revision provenance) with its owning approach identity and
// how many canonical signatures have been computed for it.
func (s *Store) ListMechanismsForProblem(ctx context.Context, problemID string) ([]MechanismListItem, error) {
	if err := domain.ValidateProblemID(problemID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT m.id, a.id, ar.id, ar.label, a.logical_identity, COALESCE(o.class, 'unknown'),
       (SELECT COUNT(*) FROM mechanism_signatures ms WHERE ms.mechanism_id = m.id)
FROM mechanisms m
JOIN approach_revisions ar ON ar.id = m.approach_revision_id
JOIN approaches a ON a.id = ar.approach_id
LEFT JOIN outcomes o ON o.approach_revision_id = ar.id
WHERE a.problem_id = ?
ORDER BY m.id
`, problemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []MechanismListItem
	for rows.Next() {
		var item MechanismListItem
		if err := rows.Scan(&item.MechanismID, &item.ApproachID, &item.ApproachRevisionID, &item.Label, &item.LogicalIdentity, &item.OutcomeClass, &item.SignatureCount); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// GetMechanismDetail loads the approach detail owning a mechanism.
func (s *Store) GetMechanismDetail(ctx context.Context, mechanismID string) (ApproachDetail, error) {
	if err := domain.ValidateMechanismID(mechanismID); err != nil {
		return ApproachDetail{}, err
	}
	row := s.db.QueryRowContext(ctx, `
SELECT ar.id
FROM mechanisms m
JOIN approach_revisions ar ON ar.id = m.approach_revision_id
WHERE m.id = ?
`, mechanismID)
	var approachRevisionID string
	if err := row.Scan(&approachRevisionID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ApproachDetail{}, fmt.Errorf("%w: mechanism %s", ErrNotFound, mechanismID)
		}
		return ApproachDetail{}, err
	}
	revision, err := s.getApproachRevision(ctx, approachRevisionID)
	if err != nil {
		return ApproachDetail{}, err
	}
	approach, err := s.getApproach(ctx, revision.ApproachID)
	if err != nil {
		return ApproachDetail{}, err
	}
	revisions, err := s.listApproachRevisions(ctx, approach.ID)
	if err != nil {
		return ApproachDetail{}, err
	}
	return s.approachDetailForRevision(ctx, approach, revision, len(revisions))
}

// ListApproachRevisions returns all revisions for an approach, newest first.
func (s *Store) ListApproachRevisions(ctx context.Context, approachID string) (domain.Approach, []domain.ApproachRevision, error) {
	if err := domain.ValidateApproachID(approachID); err != nil {
		return domain.Approach{}, nil, err
	}
	approach, err := s.getApproach(ctx, approachID)
	if err != nil {
		return domain.Approach{}, nil, err
	}
	revisions, err := s.listApproachRevisions(ctx, approachID)
	if err != nil {
		return domain.Approach{}, nil, err
	}
	return approach, revisions, nil
}

func (s *Store) approachDetailForRevision(ctx context.Context, approach domain.Approach, revision domain.ApproachRevision, revisionCount int) (ApproachDetail, error) {
	normalization, err := s.getNormalizationRevision(ctx, revision.NormalizationRevisionID)
	if err != nil {
		return ApproachDetail{}, err
	}
	invocation, err := s.getProviderInvocation(ctx, normalization.ProviderInvocationID)
	if err != nil {
		return ApproachDetail{}, err
	}
	mechanism, attributes, err := s.getMechanism(ctx, revision.ID)
	if err != nil {
		return ApproachDetail{}, err
	}
	outcome, _, err := s.getOutcome(ctx, revision.ID)
	if err != nil {
		return ApproachDetail{}, err
	}
	boundaries, err := s.listFailureBoundaries(ctx, revision.ID)
	if err != nil {
		return ApproachDetail{}, err
	}
	support, err := s.listSourceSupports(ctx, revision.ID)
	if err != nil {
		return ApproachDetail{}, err
	}
	completeness, err := s.listFieldCompleteness(ctx, mechanism.ID)
	if err != nil {
		return ApproachDetail{}, err
	}
	return ApproachDetail{
		Approach:          approach,
		Revision:          revision,
		Normalization:     normalization,
		Invocation:        invocation,
		SnapshotID:        normalization.SnapshotID,
		Mechanism:         mechanism,
		Attributes:        attributes,
		FieldCompleteness: completeness,
		Outcome:           outcome,
		Boundaries:        boundaries,
		Support:           support,
		RevisionCount:     revisionCount,
	}, nil
}

// listFieldCompleteness loads the justified per-field exhaustiveness
// declarations for one mechanism (v25). Deterministic order by field kind.
func (s *Store) listFieldCompleteness(ctx context.Context, mechanismID string) ([]domain.MechanismFieldCompleteness, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT field_kind, completeness, basis
FROM mechanism_field_completeness WHERE mechanism_id = ? ORDER BY field_kind
`, mechanismID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.MechanismFieldCompleteness
	for rows.Next() {
		var kind, completeness, basis string
		if err := rows.Scan(&kind, &completeness, &basis); err != nil {
			return nil, err
		}
		out = append(out, domain.MechanismFieldCompleteness{
			MechanismID:  mechanismID,
			Kind:         domain.MechanismAttributeKind(kind),
			Completeness: domain.FieldCompleteness(completeness),
			Basis:        basis,
		})
	}
	return out, rows.Err()
}

func (s *Store) getApproach(ctx context.Context, approachID string) (domain.Approach, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, problem_id, logical_identity, created_at
FROM approaches WHERE id = ?
`, approachID)
	var approach domain.Approach
	var createdAt string
	if err := row.Scan(&approach.ID, &approach.ProblemID, &approach.LogicalIdentity, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Approach{}, fmt.Errorf("%w: approach %s", ErrNotFound, approachID)
		}
		return domain.Approach{}, err
	}
	parsed, err := parseTime(createdAt)
	if err != nil {
		return domain.Approach{}, fmt.Errorf("%w: %v", ErrCorruptStore, err)
	}
	approach.CreatedAt = parsed
	return approach, nil
}

func (s *Store) listApproachRevisions(ctx context.Context, approachID string) ([]domain.ApproachRevision, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, approach_id, normalization_revision_id, label, description, supersedes_revision_id, created_at
FROM approach_revisions
WHERE approach_id = ?
ORDER BY created_at DESC, id DESC
`, approachID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var revisions []domain.ApproachRevision
	for rows.Next() {
		revision, scanErr := scanApproachRevision(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		revisions = append(revisions, revision)
	}
	return revisions, rows.Err()
}

func (s *Store) getApproachRevision(ctx context.Context, id string) (domain.ApproachRevision, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, approach_id, normalization_revision_id, label, description, supersedes_revision_id, created_at
FROM approach_revisions WHERE id = ?
`, id)
	revision, err := scanApproachRevision(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ApproachRevision{}, fmt.Errorf("%w: approach revision %s", ErrNotFound, id)
		}
		return domain.ApproachRevision{}, err
	}
	return revision, nil
}

func scanApproachRevision(row scanner) (domain.ApproachRevision, error) {
	var revision domain.ApproachRevision
	var supersedes sql.NullString
	var createdAt string
	if err := row.Scan(&revision.ID, &revision.ApproachID, &revision.NormalizationRevisionID, &revision.Label, &revision.Description, &supersedes, &createdAt); err != nil {
		return domain.ApproachRevision{}, err
	}
	parsed, err := parseTime(createdAt)
	if err != nil {
		return domain.ApproachRevision{}, fmt.Errorf("%w: %v", ErrCorruptStore, err)
	}
	revision.CreatedAt = parsed
	if supersedes.Valid {
		revision.SupersedesRevisionID = &supersedes.String
	}
	return revision, nil
}

func (s *Store) getNormalizationRevision(ctx context.Context, id string) (domain.NormalizationRevision, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, problem_id, run_id, snapshot_id, provider_invocation_id, schema_version, config_hash, status, supersedes_revision_id, skip_reason, created_at
FROM normalization_revisions WHERE id = ?
`, id)
	revision, err := scanNormalizationRevision(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.NormalizationRevision{}, fmt.Errorf("%w: normalization revision %s", ErrNotFound, id)
		}
		return domain.NormalizationRevision{}, err
	}
	return revision, nil
}

func (s *Store) snapshotForNormalizationRevision(ctx context.Context, id string) (string, error) {
	revision, err := s.getNormalizationRevision(ctx, id)
	if err != nil {
		return "", err
	}
	return revision.SnapshotID, nil
}

func (s *Store) getProviderInvocation(ctx context.Context, id string) (domain.ProviderInvocation, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, run_id, role, provider_name, provider_version, model_name, schema_version, request_hash, request_payload, response_payload, created_at
FROM provider_invocations WHERE id = ?
`, id)
	var invocation domain.ProviderInvocation
	var role string
	var createdAt string
	if err := row.Scan(&invocation.ID, &invocation.RunID, &role, &invocation.ProviderName, &invocation.ProviderVersion, &invocation.ModelName, &invocation.SchemaVersion, &invocation.RequestHash, &invocation.RequestPayload, &invocation.ResponsePayload, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ProviderInvocation{}, fmt.Errorf("%w: provider invocation %s", ErrNotFound, id)
		}
		return domain.ProviderInvocation{}, err
	}
	parsed, err := parseTime(createdAt)
	if err != nil {
		return domain.ProviderInvocation{}, fmt.Errorf("%w: %v", ErrCorruptStore, err)
	}
	invocation.Role = domain.ProviderRole(role)
	invocation.CreatedAt = parsed
	return invocation, nil
}

func (s *Store) getMechanism(ctx context.Context, approachRevisionID string) (domain.Mechanism, []domain.MechanismAttribute, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, approach_revision_id, locality, construction_mode, uncertainty_mode, notes
FROM mechanisms WHERE approach_revision_id = ?
`, approachRevisionID)
	var mechanism domain.Mechanism
	var locality, construction, uncertainty string
	if err := row.Scan(&mechanism.ID, &mechanism.ApproachRevisionID, &locality, &construction, &uncertainty, &mechanism.Notes); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Mechanism{}, nil, fmt.Errorf("%w: mechanism for revision %s", ErrNotFound, approachRevisionID)
		}
		return domain.Mechanism{}, nil, err
	}
	mechanism.Locality = domain.Locality(locality)
	mechanism.ConstructionMode = domain.ConstructionMode(construction)
	mechanism.UncertaintyMode = domain.UncertaintyMode(uncertainty)

	rows, err := s.db.QueryContext(ctx, `
SELECT kind, value
FROM mechanism_attributes
WHERE mechanism_id = ?
ORDER BY kind ASC, ordinal ASC
`, mechanism.ID)
	if err != nil {
		return domain.Mechanism{}, nil, err
	}
	defer rows.Close()
	var attributes []domain.MechanismAttribute
	for rows.Next() {
		var kind, value string
		if err := rows.Scan(&kind, &value); err != nil {
			return domain.Mechanism{}, nil, err
		}
		attributes = append(attributes, domain.MechanismAttribute{
			MechanismID: mechanism.ID,
			Kind:        domain.MechanismAttributeKind(kind),
			Value:       value,
		})
	}
	return mechanism, attributes, rows.Err()
}

func (s *Store) getOutcome(ctx context.Context, approachRevisionID string) (domain.Outcome, bool, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, approach_revision_id, class, boundary_statement, notes
FROM outcomes WHERE approach_revision_id = ?
`, approachRevisionID)
	var outcome domain.Outcome
	var class string
	if err := row.Scan(&outcome.ID, &outcome.ApproachRevisionID, &class, &outcome.BoundaryStatement, &outcome.Notes); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Outcome{}, false, nil
		}
		return domain.Outcome{}, false, err
	}
	outcome.Class = domain.OutcomeClass(class)
	return outcome, true, nil
}

func (s *Store) listFailureBoundaries(ctx context.Context, approachRevisionID string) ([]domain.FailureBoundary, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT id, approach_revision_id, condition
FROM failure_boundaries
WHERE approach_revision_id = ?
ORDER BY ordinal ASC, id ASC
`, approachRevisionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var boundaries []domain.FailureBoundary
	for rows.Next() {
		var boundary domain.FailureBoundary
		if err := rows.Scan(&boundary.ID, &boundary.ApproachRevisionID, &boundary.Condition); err != nil {
			return nil, err
		}
		boundaries = append(boundaries, boundary)
	}
	return boundaries, rows.Err()
}

func (s *Store) listSourceSupports(ctx context.Context, approachRevisionID string) ([]domain.SourceSupport, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT approach_revision_id, snapshot_id, field_path, support_kind, locator, confidence
FROM source_supports
WHERE approach_revision_id = ?
ORDER BY field_path ASC
`, approachRevisionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var supports []domain.SourceSupport
	for rows.Next() {
		var support domain.SourceSupport
		var kind string
		if err := rows.Scan(&support.ApproachRevisionID, &support.SnapshotID, &support.FieldPath, &kind, &support.Locator, &support.Confidence); err != nil {
			return nil, err
		}
		support.SupportKind = domain.SupportKind(kind)
		supports = append(supports, support)
	}
	return supports, rows.Err()
}

// GetProviderInvocation exposes one recorded provider invocation for audit
// (retained request/response payloads + request hash).
func (s *Store) GetProviderInvocation(ctx context.Context, id string) (domain.ProviderInvocation, error) {
	return s.getProviderInvocation(ctx, id)
}
