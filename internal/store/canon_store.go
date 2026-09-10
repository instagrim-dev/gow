package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/instagrim-dev/newf/internal/domain"
)

// VocabularyRecord is the persisted header of a canonical vocabulary version.
type VocabularyRecord struct {
	Version   string
	CreatedAt string
	Notes     string
}

// TermRecord is one persisted canonical term with its aliases.
type TermRecord struct {
	VocabularyVersion string
	CanonicalID       string
	FieldKind         string
	Description       string
	ParentCanonicalID string
	Aliases           []string
}

// VocabularySeedInput is a full vocabulary definition to persist idempotently.
type VocabularySeedInput struct {
	Version   string
	Notes     string
	CreatedAt string
	Terms     []TermRecord
	// Rejected are normalized keys the vocabulary explicitly disallows. They are
	// persisted so rejection semantics survive a reload from SQLite (runtime
	// resolution rehydrates the vocabulary from the store, so an in-memory-only
	// rejected set would silently vanish after persistence).
	Rejected []string
}

// SeedVocabulary inserts a vocabulary version + terms + aliases if absent. It is
// idempotent: an already-present version is left untouched (rows are immutable).
func (s *Store) SeedVocabulary(ctx context.Context, input VocabularySeedInput) error {
	exists, err := s.vocabularyExists(ctx, input.Version)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
INSERT INTO canonical_vocabulary(version, created_at, notes) VALUES(?, ?, ?)
`, input.Version, input.CreatedAt, input.Notes); err != nil {
		return err
	}
	for _, term := range input.Terms {
		var parent any
		if term.ParentCanonicalID != "" {
			parent = term.ParentCanonicalID
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO canonical_terms(vocabulary_version, canonical_id, field_kind, description, parent_canonical_id)
VALUES(?, ?, ?, ?, ?)
`, input.Version, term.CanonicalID, term.FieldKind, term.Description, parent); err != nil {
			return err
		}
		for _, alias := range term.Aliases {
			// INSERT OR IGNORE is safe here because the alias key now includes
			// canonical_id: it ignores only exact-duplicate rows, and can never
			// collapse a second binding to a *different* canonical id (that row
			// has a different key). The alias namespace is scoped by field kind,
			// so the same phrase may bind to different canonical ids across
			// field kinds, and to multiple ids within one field kind (which the
			// resolver reports as ambiguous). No binding is silently dropped.
			if _, err := tx.ExecContext(ctx, `
INSERT OR IGNORE INTO canonical_term_aliases(vocabulary_version, canonical_id, field_kind, alias_normalized)
VALUES(?, ?, ?, ?)
`, input.Version, term.CanonicalID, term.FieldKind, alias); err != nil {
				return err
			}
		}
	}
	for _, rejected := range input.Rejected {
		if _, err := tx.ExecContext(ctx, `
INSERT OR IGNORE INTO canonical_rejected_terms(vocabulary_version, rejected_normalized)
VALUES(?, ?)
`, input.Version, rejected); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// ListRejected returns the persisted rejected keys for a vocabulary version.
func (s *Store) ListRejected(ctx context.Context, version string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT rejected_normalized FROM canonical_rejected_terms WHERE vocabulary_version = ? ORDER BY rejected_normalized`, version)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		out = append(out, key)
	}
	return out, rows.Err()
}

func (s *Store) vocabularyExists(ctx context.Context, version string) (bool, error) {
	row := s.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM canonical_vocabulary WHERE version = ?)`, version)
	var exists bool
	if err := row.Scan(&exists); err != nil {
		return false, err
	}
	return exists, nil
}

// ListVocabularies returns all persisted vocabulary versions, newest first.
func (s *Store) ListVocabularies(ctx context.Context) ([]VocabularyRecord, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT version, created_at, notes FROM canonical_vocabulary ORDER BY version`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []VocabularyRecord
	for rows.Next() {
		var v VocabularyRecord
		if err := rows.Scan(&v.Version, &v.CreatedAt, &v.Notes); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// ListTerms returns the terms of a vocabulary version, optionally filtered by
// field kind (pass "" for all), each with its aliases.
func (s *Store) ListTerms(ctx context.Context, version, fieldKind string) ([]TermRecord, error) {
	if exists, err := s.vocabularyExists(ctx, version); err != nil {
		return nil, err
	} else if !exists {
		return nil, fmt.Errorf("%w: vocabulary %s", ErrNotFound, version)
	}

	query := `SELECT canonical_id, field_kind, description, COALESCE(parent_canonical_id, '') FROM canonical_terms WHERE vocabulary_version = ?`
	args := []any{version}
	if fieldKind != "" {
		query += ` AND field_kind = ?`
		args = append(args, fieldKind)
	}
	query += ` ORDER BY canonical_id`

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []TermRecord
	for rows.Next() {
		t := TermRecord{VocabularyVersion: version}
		if err := rows.Scan(&t.CanonicalID, &t.FieldKind, &t.Description, &t.ParentCanonicalID); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range out {
		aliases, err := s.listAliases(ctx, version, out[i].CanonicalID)
		if err != nil {
			return nil, err
		}
		out[i].Aliases = aliases
	}
	return out, nil
}

// GetTerm returns a single term by (version, canonical_id).
func (s *Store) GetTerm(ctx context.Context, version, canonicalID string) (TermRecord, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT canonical_id, field_kind, description, COALESCE(parent_canonical_id, '')
FROM canonical_terms WHERE vocabulary_version = ? AND canonical_id = ?
`, version, canonicalID)
	t := TermRecord{VocabularyVersion: version}
	if err := row.Scan(&t.CanonicalID, &t.FieldKind, &t.Description, &t.ParentCanonicalID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return TermRecord{}, fmt.Errorf("%w: term %s in %s", ErrNotFound, canonicalID, version)
		}
		return TermRecord{}, err
	}
	aliases, err := s.listAliases(ctx, version, canonicalID)
	if err != nil {
		return TermRecord{}, err
	}
	t.Aliases = aliases
	return t, nil
}

func (s *Store) listAliases(ctx context.Context, version, canonicalID string) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `
SELECT alias_normalized FROM canonical_term_aliases
WHERE vocabulary_version = ? AND canonical_id = ? ORDER BY alias_normalized
`, version, canonicalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var a string
		if err := rows.Scan(&a); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// SignatureFieldClaimRow is one persisted field-claim row.
type SignatureFieldClaimRow struct {
	FieldKind          string
	SurfaceLabel       string
	ResolutionState    string
	CanonicalID        string
	ClaimStatus        string
	SupportSnapshotID  string
	SupportLocator     string
	Confidence         string
	ClassifierContract string
	Ordinal            int
}

// SignatureBoundaryRow is one persisted signature boundary.
type SignatureBoundaryRow struct {
	SurfaceLabel      string
	ResolutionState   string
	CanonicalID       string
	Relation          string
	ClaimStatus       string
	SupportSnapshotID string
	SupportLocator    string
	Ordinal           int
}

// SignatureRecord is the full persisted signature.
type SignatureRecord struct {
	ID                string
	MechanismID       string
	SchemaVersion     string
	VocabularyVersion string
	Fingerprint       string
	RunID             string
	CreatedAt         string
	OutcomeClass      string
	OutcomeStatus     string
	Posture           map[string]string
	PostureStatus     map[string]string
	FieldClaims       []SignatureFieldClaimRow
	Boundaries        []SignatureBoundaryRow
}

// PersistSignatureResult reports the persisted signature and whether it was new.
type PersistSignatureResult struct {
	Record  SignatureRecord
	Created bool
}

// PersistSignature writes a mechanism signature transactionally. If a signature
// already exists for (mechanism_id, schema_version, vocabulary_version) it is
// returned unchanged with Created=false (idempotent), never rewritten.
func (s *Store) PersistSignature(ctx context.Context, record SignatureRecord) (PersistSignatureResult, error) {
	if err := domain.ValidateMechanismSignatureID(record.ID); err != nil {
		return PersistSignatureResult{}, err
	}
	if err := domain.ValidateMechanismID(record.MechanismID); err != nil {
		return PersistSignatureResult{}, err
	}

	existing, found, err := s.findSignature(ctx, record.MechanismID, record.SchemaVersion, record.VocabularyVersion)
	if err != nil {
		return PersistSignatureResult{}, err
	}
	if found {
		full, err := s.loadSignature(ctx, existing)
		if err != nil {
			return PersistSignatureResult{}, err
		}
		return PersistSignatureResult{Record: full, Created: false}, nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return PersistSignatureResult{}, err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
INSERT INTO mechanism_signatures(id, mechanism_id, schema_version, vocabulary_version, fingerprint, run_id, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?)
`, record.ID, record.MechanismID, record.SchemaVersion, record.VocabularyVersion, record.Fingerprint, record.RunID, record.CreatedAt); err != nil {
		return PersistSignatureResult{}, err
	}

	for _, c := range record.FieldClaims {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO signature_field_claims(signature_id, field_kind, surface_label, resolution_state, canonical_id, claim_status, support_snapshot_id, support_locator, confidence, classifier_contract, ordinal)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, record.ID, c.FieldKind, c.SurfaceLabel, c.ResolutionState, c.CanonicalID, c.ClaimStatus, c.SupportSnapshotID, c.SupportLocator, c.Confidence, c.ClassifierContract, c.Ordinal); err != nil {
			return PersistSignatureResult{}, err
		}
	}
	for axis, value := range record.Posture {
		status := record.PostureStatus[axis]
		if status == "" {
			status = "unknown"
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO signature_postures(signature_id, axis, value, claim_status) VALUES(?, ?, ?, ?)
`, record.ID, axis, value, status); err != nil {
			return PersistSignatureResult{}, err
		}
	}
	for _, b := range record.Boundaries {
		status := b.ClaimStatus
		if status == "" {
			status = "unknown"
		}
		if _, err := tx.ExecContext(ctx, `
INSERT INTO signature_boundaries(signature_id, surface_label, resolution_state, canonical_id, relation, claim_status, support_snapshot_id, support_locator, ordinal)
VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)
`, record.ID, b.SurfaceLabel, b.ResolutionState, b.CanonicalID, b.Relation, status, b.SupportSnapshotID, b.SupportLocator, b.Ordinal); err != nil {
			return PersistSignatureResult{}, err
		}
	}
	outcomeStatus := record.OutcomeStatus
	if outcomeStatus == "" {
		outcomeStatus = "unknown"
	}
	if _, err := tx.ExecContext(ctx, `
INSERT INTO signature_outcomes(signature_id, class, claim_status) VALUES(?, ?, ?)
`, record.ID, record.OutcomeClass, outcomeStatus); err != nil {
		return PersistSignatureResult{}, err
	}

	if err := tx.Commit(); err != nil {
		return PersistSignatureResult{}, err
	}
	return PersistSignatureResult{Record: record, Created: true}, nil
}

func (s *Store) findSignature(ctx context.Context, mechanismID, schemaVersion, vocabVersion string) (string, bool, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id FROM mechanism_signatures
WHERE mechanism_id = ? AND schema_version = ? AND vocabulary_version = ?
`, mechanismID, schemaVersion, vocabVersion)
	var id string
	if err := row.Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}
		return "", false, err
	}
	return id, true, nil
}

// GetSignature loads a full signature by ID.
func (s *Store) GetSignature(ctx context.Context, id string) (SignatureRecord, error) {
	if err := domain.ValidateMechanismSignatureID(id); err != nil {
		return SignatureRecord{}, err
	}
	return s.loadSignature(ctx, id)
}

func (s *Store) loadSignature(ctx context.Context, id string) (SignatureRecord, error) {
	row := s.db.QueryRowContext(ctx, `
SELECT id, mechanism_id, schema_version, vocabulary_version, fingerprint, run_id, created_at
FROM mechanism_signatures WHERE id = ?
`, id)
	rec := SignatureRecord{Posture: map[string]string{}, PostureStatus: map[string]string{}}
	if err := row.Scan(&rec.ID, &rec.MechanismID, &rec.SchemaVersion, &rec.VocabularyVersion, &rec.Fingerprint, &rec.RunID, &rec.CreatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SignatureRecord{}, fmt.Errorf("%w: signature %s", ErrNotFound, id)
		}
		return SignatureRecord{}, err
	}

	claimRows, err := s.db.QueryContext(ctx, `
SELECT field_kind, surface_label, resolution_state, canonical_id, claim_status, support_snapshot_id, support_locator, confidence, classifier_contract, ordinal
FROM signature_field_claims WHERE signature_id = ? ORDER BY field_kind, ordinal
`, id)
	if err != nil {
		return SignatureRecord{}, err
	}
	defer claimRows.Close()
	for claimRows.Next() {
		var c SignatureFieldClaimRow
		if err := claimRows.Scan(&c.FieldKind, &c.SurfaceLabel, &c.ResolutionState, &c.CanonicalID, &c.ClaimStatus, &c.SupportSnapshotID, &c.SupportLocator, &c.Confidence, &c.ClassifierContract, &c.Ordinal); err != nil {
			return SignatureRecord{}, err
		}
		rec.FieldClaims = append(rec.FieldClaims, c)
	}
	if err := claimRows.Err(); err != nil {
		return SignatureRecord{}, err
	}

	postureRows, err := s.db.QueryContext(ctx, `SELECT axis, value, claim_status FROM signature_postures WHERE signature_id = ?`, id)
	if err != nil {
		return SignatureRecord{}, err
	}
	defer postureRows.Close()
	for postureRows.Next() {
		var axis, value, status string
		if err := postureRows.Scan(&axis, &value, &status); err != nil {
			return SignatureRecord{}, err
		}
		rec.Posture[axis] = value
		rec.PostureStatus[axis] = status
	}
	if err := postureRows.Err(); err != nil {
		return SignatureRecord{}, err
	}

	boundaryRows, err := s.db.QueryContext(ctx, `
SELECT surface_label, resolution_state, canonical_id, relation, claim_status, support_snapshot_id, support_locator, ordinal
FROM signature_boundaries WHERE signature_id = ? ORDER BY ordinal
`, id)
	if err != nil {
		return SignatureRecord{}, err
	}
	defer boundaryRows.Close()
	for boundaryRows.Next() {
		var b SignatureBoundaryRow
		if err := boundaryRows.Scan(&b.SurfaceLabel, &b.ResolutionState, &b.CanonicalID, &b.Relation, &b.ClaimStatus, &b.SupportSnapshotID, &b.SupportLocator, &b.Ordinal); err != nil {
			return SignatureRecord{}, err
		}
		rec.Boundaries = append(rec.Boundaries, b)
	}
	if err := boundaryRows.Err(); err != nil {
		return SignatureRecord{}, err
	}

	outcomeRow := s.db.QueryRowContext(ctx, `SELECT class, claim_status FROM signature_outcomes WHERE signature_id = ?`, id)
	if err := outcomeRow.Scan(&rec.OutcomeClass, &rec.OutcomeStatus); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return SignatureRecord{}, err
	}
	return rec, nil
}

// ListSignaturesForProblem returns the persisted signature ids for a problem
// under one (schema_version, vocabulary_version), newest-mechanism first is not
// meaningful here so results are ordered by signature id for determinism. It
// joins signatures back to their owning problem through
// mechanism -> approach_revision -> approach.
//
// Clustering keys on a single version tuple (KTD-1), so callers pass the exact
// schema + vocabulary version; a signature under a different version is a
// different clustering run and is excluded here.
func (s *Store) ListSignaturesForProblem(ctx context.Context, problemID, schemaVersion, vocabVersion string) ([]string, error) {
	if err := domain.ValidateProblemID(problemID); err != nil {
		return nil, err
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT sig.id
FROM mechanism_signatures sig
JOIN mechanisms m ON m.id = sig.mechanism_id
JOIN approach_revisions ar ON ar.id = m.approach_revision_id
JOIN approaches a ON a.id = ar.approach_id
WHERE a.problem_id = ? AND sig.schema_version = ? AND sig.vocabulary_version = ?
ORDER BY sig.id
`, problemID, schemaVersion, vocabVersion)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

// ComparisonFieldResultRow is one persisted per-field comparison result.
type ComparisonFieldResultRow struct {
	FieldKind    string
	OverlapCount int
	UnionCount   int
	Jaccard      float64
	Ordinal      string
	Incomparable bool
}

// ComparisonRecord is the full persisted comparison run.
type ComparisonRecord struct {
	ID              string
	SignatureAID    string
	SignatureBID    string
	WeightsVersion  string
	ClassifyVersion string
	Classification  string
	RunID           string
	CreatedAt       string
	Fields          []ComparisonFieldResultRow
}

// PersistComparison writes a comparison run + field results transactionally.
func (s *Store) PersistComparison(ctx context.Context, record ComparisonRecord) error {
	if err := domain.ValidateComparisonRunID(record.ID); err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
INSERT INTO comparison_runs(id, signature_a_id, signature_b_id, weights_version, classify_version, classification, run_id, created_at)
VALUES(?, ?, ?, ?, ?, ?, ?, ?)
`, record.ID, record.SignatureAID, record.SignatureBID, record.WeightsVersion, record.ClassifyVersion, record.Classification, record.RunID, record.CreatedAt); err != nil {
		return err
	}
	for _, f := range record.Fields {
		if _, err := tx.ExecContext(ctx, `
INSERT INTO comparison_field_results(comparison_run_id, field_kind, overlap_count, union_count, jaccard, ordinal, incomparable)
VALUES(?, ?, ?, ?, ?, ?, ?)
`, record.ID, f.FieldKind, f.OverlapCount, f.UnionCount, f.Jaccard, f.Ordinal, boolToInt(f.Incomparable)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
