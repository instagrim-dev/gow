package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// OutcomeClass is the normalized epistemic result of an approach.
//
// It is interpretation, not verified truth: a normalized outcome records what
// the source material reports about how far an approach got, never an
// independent verification of correctness.
type OutcomeClass string

const (
	OutcomeUnknown        OutcomeClass = "unknown"
	OutcomeFailure        OutcomeClass = "failure"
	OutcomePartialFailure OutcomeClass = "partial_failure"
	OutcomePartialSuccess OutcomeClass = "partial_success"
	OutcomeSuccess        OutcomeClass = "success"
	// OutcomeMixed is an AGGREGATE-ONLY outcome class: it is never carried by a
	// single mechanism/signature (see Valid, which rejects it), but a mechanism
	// family or failure-space partition that spans more than one distinct member
	// outcome is reported as mixed rather than being compressed to the outcome of
	// a single representative. This preserves the truth that a family can contain
	// both failures and partial successes.
	OutcomeMixed OutcomeClass = "mixed"
)

func (c OutcomeClass) Valid() bool {
	switch c {
	case OutcomeUnknown, OutcomeFailure, OutcomePartialFailure, OutcomePartialSuccess, OutcomeSuccess:
		return true
	default:
		return false
	}
}

// ValidFamilyOutcome reports whether c is a valid AGGREGATE outcome for a
// mechanism family or failure-space partition: any signature-level class plus
// the aggregate-only "mixed".
func (c OutcomeClass) ValidFamilyOutcome() bool {
	return c.Valid() || c == OutcomeMixed
}

// Locality captures whether an approach reasons locally, globally, or a mix.
type Locality string

const (
	LocalityUnknown Locality = "unknown"
	LocalityLocal   Locality = "local"
	LocalityGlobal  Locality = "global"
	LocalityMixed   Locality = "mixed"
)

func (l Locality) Valid() bool {
	switch l {
	case LocalityUnknown, LocalityLocal, LocalityGlobal, LocalityMixed:
		return true
	default:
		return false
	}
}

// ConstructionMode captures constructive vs existential reasoning posture.
type ConstructionMode string

const (
	ConstructionUnknown      ConstructionMode = "unknown"
	ConstructionConstructive ConstructionMode = "constructive"
	ConstructionExistential  ConstructionMode = "existential"
	ConstructionMixed        ConstructionMode = "mixed"
)

func (m ConstructionMode) Valid() bool {
	switch m {
	case ConstructionUnknown, ConstructionConstructive, ConstructionExistential, ConstructionMixed:
		return true
	default:
		return false
	}
}

// UncertaintyMode captures deterministic vs probabilistic reasoning posture.
type UncertaintyMode string

const (
	UncertaintyUnknown       UncertaintyMode = "unknown"
	UncertaintyDeterministic UncertaintyMode = "deterministic"
	UncertaintyProbabilistic UncertaintyMode = "probabilistic"
	UncertaintyMixed         UncertaintyMode = "mixed"
)

func (m UncertaintyMode) Valid() bool {
	switch m {
	case UncertaintyUnknown, UncertaintyDeterministic, UncertaintyProbabilistic, UncertaintyMixed:
		return true
	default:
		return false
	}
}

// SupportKind is the provenance strength of a single normalized field.
//
// This is the epistemic boundary the normalization operator must never blur:
// a field that the model inferred is not the same as a field the source
// explicitly stated, and an unsupported generated claim is weaker still.
type SupportKind string

const (
	// SupportExplicit means the field is directly stated by source material,
	// ideally with a locator into the exact snapshot bytes.
	SupportExplicit SupportKind = "explicit"
	// SupportInferred means the model derived the field from source material
	// but the source did not state it verbatim.
	SupportInferred SupportKind = "inferred"
	// SupportUnsupported means the field is a generated claim with no source
	// grounding. It may still be recorded, but it must never be treated as
	// evidence.
	SupportUnsupported SupportKind = "unsupported"
)

func (k SupportKind) Valid() bool {
	switch k {
	case SupportExplicit, SupportInferred, SupportUnsupported:
		return true
	default:
		return false
	}
}

// NormalizationStatus is the lifecycle state of a normalization revision.
type NormalizationStatus string

const (
	NormalizationStatusSucceeded NormalizationStatus = "succeeded"
	NormalizationStatusFailed    NormalizationStatus = "failed"
	NormalizationStatusSkipped   NormalizationStatus = "skipped"
)

func (s NormalizationStatus) Valid() bool {
	switch s {
	case NormalizationStatusSucceeded, NormalizationStatusFailed, NormalizationStatusSkipped:
		return true
	default:
		return false
	}
}

// ProviderRole records which model-native operator an invocation performed.
// Normalization must record role "normalize" even if the same model later
// performs clustering or invariant mining.
type ProviderRole string

const RoleNormalize ProviderRole = "normalize"

// NormalizationRevision is the durable, revisionable interpretation artifact.
//
// A revision is created (never overwritten) whenever provider/model, schema
// version, normalization contract, or source snapshot changes, or when a user
// forces re-normalization. Revisions link to the exact snapshot consumed and
// the provider invocation that produced them.
type NormalizationRevision struct {
	ID                   string
	ProblemID            string
	RunID                string
	SnapshotID           string
	ProviderInvocationID string
	SchemaVersion        string
	ConfigHash           string
	Status               NormalizationStatus
	SupersedesRevisionID *string
	SkipReason           *string
	CreatedAt            time.Time
}

func (r NormalizationRevision) Validate() error {
	if err := ValidateNormalizationRevisionID(r.ID); err != nil {
		return err
	}
	if err := ValidateProblemID(r.ProblemID); err != nil {
		return err
	}
	if err := ValidateRunID(r.RunID); err != nil {
		return err
	}
	if err := ValidateSnapshotID(r.SnapshotID); err != nil {
		return err
	}
	if err := ValidateProviderInvocationID(r.ProviderInvocationID); err != nil {
		return err
	}
	if strings.TrimSpace(r.SchemaVersion) == "" {
		return errors.New("normalization revision schema_version is required")
	}
	if strings.TrimSpace(r.ConfigHash) == "" {
		return errors.New("normalization revision config_hash is required")
	}
	if !r.Status.Valid() {
		return fmt.Errorf("invalid normalization status %q", r.Status)
	}
	if r.SupersedesRevisionID != nil {
		if err := ValidateNormalizationRevisionID(*r.SupersedesRevisionID); err != nil {
			return err
		}
	}
	if r.CreatedAt.IsZero() {
		return errors.New("normalization revision created_at is required")
	}
	return nil
}

// ProviderInvocation records replayable provenance for a single provider call.
// It must never store credentials or authorization headers.
type ProviderInvocation struct {
	ID              string
	RunID           string
	Role            ProviderRole
	ProviderName    string
	ProviderVersion string
	ModelName       string
	SchemaVersion   string
	RequestHash     string
	RequestPayload  string
	ResponsePayload string
	CreatedAt       time.Time
}

func (p ProviderInvocation) Validate() error {
	if err := ValidateProviderInvocationID(p.ID); err != nil {
		return err
	}
	if err := ValidateRunID(p.RunID); err != nil {
		return err
	}
	if p.Role != RoleNormalize {
		return fmt.Errorf("provider invocation role %q is not %q", p.Role, RoleNormalize)
	}
	if strings.TrimSpace(p.ProviderName) == "" {
		return errors.New("provider invocation provider_name is required")
	}
	if strings.TrimSpace(p.RequestHash) == "" {
		return errors.New("provider invocation request_hash is required")
	}
	if p.CreatedAt.IsZero() {
		return errors.New("provider invocation created_at is required")
	}
	return nil
}

// Approach is the stable logical identity of a normalized attempt. Multiple
// revisions of interpretation attach to one approach over time; a single
// source snapshot may yield multiple approaches.
type Approach struct {
	ID              string
	ProblemID       string
	LogicalIdentity string
	CreatedAt       time.Time
}

func (a Approach) Validate() error {
	if err := ValidateApproachID(a.ID); err != nil {
		return err
	}
	if err := ValidateProblemID(a.ProblemID); err != nil {
		return err
	}
	if strings.TrimSpace(a.LogicalIdentity) == "" {
		return errors.New("approach logical_identity is required")
	}
	if a.CreatedAt.IsZero() {
		return errors.New("approach created_at is required")
	}
	return nil
}

// ApproachRevision is one interpretation of an approach produced by one
// normalization revision. It carries the human-facing label/description and
// links back to the mechanism, outcome, and source support for that reading.
type ApproachRevision struct {
	ID                      string
	ApproachID              string
	NormalizationRevisionID string
	Label                   string
	Description             string
	SupersedesRevisionID    *string
	CreatedAt               time.Time
}

func (r ApproachRevision) Validate() error {
	if err := ValidateApproachRevisionID(r.ID); err != nil {
		return err
	}
	if err := ValidateApproachID(r.ApproachID); err != nil {
		return err
	}
	if err := ValidateNormalizationRevisionID(r.NormalizationRevisionID); err != nil {
		return err
	}
	if strings.TrimSpace(r.Label) == "" {
		return errors.New("approach revision label is required")
	}
	if r.SupersedesRevisionID != nil {
		if err := ValidateApproachRevisionID(*r.SupersedesRevisionID); err != nil {
			return err
		}
	}
	if r.CreatedAt.IsZero() {
		return errors.New("approach revision created_at is required")
	}
	return nil
}

// Mechanism is the typed mechanistic representation of one approach revision.
// Extensible axes live in MechanismAxisValue rather than as fixed columns so
// the vocabulary can grow without rewriting prior records.
type Mechanism struct {
	ID                 string
	ApproachRevisionID string
	Locality           Locality
	ConstructionMode   ConstructionMode
	UncertaintyMode    UncertaintyMode
	Notes              string
}

func (m Mechanism) Validate() error {
	if err := ValidateMechanismID(m.ID); err != nil {
		return err
	}
	if err := ValidateApproachRevisionID(m.ApproachRevisionID); err != nil {
		return err
	}
	if !m.Locality.Valid() {
		return fmt.Errorf("invalid locality %q", m.Locality)
	}
	if !m.ConstructionMode.Valid() {
		return fmt.Errorf("invalid construction_mode %q", m.ConstructionMode)
	}
	if !m.UncertaintyMode.Valid() {
		return fmt.Errorf("invalid uncertainty_mode %q", m.UncertaintyMode)
	}
	return nil
}

// MechanismAttributeKind enumerates the list-valued mechanism dimensions.
type MechanismAttributeKind string

const (
	AttrRepresentation  MechanismAttributeKind = "representation"
	AttrAssumption      MechanismAttributeKind = "assumption"
	AttrOperator        MechanismAttributeKind = "operator"
	AttrPreserves       MechanismAttributeKind = "preserves"
	AttrBreaks          MechanismAttributeKind = "breaks"
	AttrAuxiliaryObject MechanismAttributeKind = "auxiliary_object"
)

func (k MechanismAttributeKind) Valid() bool {
	switch k {
	case AttrRepresentation, AttrAssumption, AttrOperator, AttrPreserves, AttrBreaks, AttrAuxiliaryObject:
		return true
	default:
		return false
	}
}

// Canonical SourceSupport.FieldPath values. FieldPath is free-form in the #7
// schema, but a stable canonical vocabulary is required so the #9 read side can
// join support rows to the fields they justify without silently defaulting an
// explicitly-supported field to unknown (which would promote/demote epistemic
// status). The canonical form matches the normalize JSON keys: plural list
// attributes, `_mode`-suffixed posture, and `boundary_statement`.
const (
	SupportPathRepresentations   = "mechanism.representations"
	SupportPathAssumptions       = "mechanism.assumptions"
	SupportPathOperators         = "mechanism.operators"
	SupportPathPreserves         = "mechanism.preserves"
	SupportPathBreaks            = "mechanism.breaks"
	SupportPathAuxiliaryObjects  = "mechanism.auxiliary_objects"
	SupportPathLocality          = "mechanism.locality"
	SupportPathConstructionMode  = "mechanism.construction_mode"
	SupportPathUncertaintyMode   = "mechanism.uncertainty_mode"
	SupportPathOutcomeClass      = "outcome.class"
	SupportPathBoundaryStatement = "outcome.boundary_statement"
)

// attributeSupportPaths maps each list attribute kind to its canonical support
// field path plus historically-authored aliases that must still join. The
// alias set exists because two conventions appear in real data: the plural
// canonical form (corpus) and an earlier singular form (older fixtures). A
// support row under any listed alias resolves to the same field so no authored
// provenance is dropped by spelling drift.
var attributeSupportPaths = map[MechanismAttributeKind][]string{
	AttrRepresentation:  {SupportPathRepresentations, "mechanism.representation"},
	AttrAssumption:      {SupportPathAssumptions, "mechanism.assumption"},
	AttrOperator:        {SupportPathOperators, "mechanism.operator"},
	AttrPreserves:       {SupportPathPreserves},
	AttrBreaks:          {SupportPathBreaks},
	AttrAuxiliaryObject: {SupportPathAuxiliaryObjects, "mechanism.auxiliary_object"},
}

// postureSupportPaths maps each posture axis to its canonical support path plus
// accepted aliases (the `_mode` suffix is canonical; the bare form is tolerated).
var postureSupportPaths = map[string][]string{
	"locality":     {SupportPathLocality},
	"construction": {SupportPathConstructionMode, "mechanism.construction"},
	"uncertainty":  {SupportPathUncertaintyMode, "mechanism.uncertainty"},
}

// outcomeSupportPaths lists accepted paths for the outcome class.
var outcomeSupportPaths = []string{SupportPathOutcomeClass}

// AttributeSupportPaths returns the accepted support field paths for a list
// attribute kind (canonical first). Empty when the kind is unknown.
func AttributeSupportPaths(kind MechanismAttributeKind) []string {
	return attributeSupportPaths[kind]
}

// PostureSupportPaths returns the accepted support field paths for a posture
// axis ("locality" / "construction" / "uncertainty"), canonical first.
func PostureSupportPaths(axis string) []string {
	return postureSupportPaths[axis]
}

// OutcomeSupportPaths returns the accepted support field paths for the outcome
// class, canonical first.
func OutcomeSupportPaths() []string {
	return outcomeSupportPaths
}

// MechanismAttribute is a single list-valued mechanism entry. Namespaced kind
// + value keeps the vocabulary extensible without schema churn.
type MechanismAttribute struct {
	MechanismID string
	Kind        MechanismAttributeKind
	Value       string
}

// Outcome is the normalized result class plus its boundary statement.
type Outcome struct {
	ID                 string
	ApproachRevisionID string
	Class              OutcomeClass
	BoundaryStatement  string
	Notes              string
}

func (o Outcome) Validate() error {
	if err := ValidateOutcomeID(o.ID); err != nil {
		return err
	}
	if err := ValidateApproachRevisionID(o.ApproachRevisionID); err != nil {
		return err
	}
	if !o.Class.Valid() {
		return fmt.Errorf("invalid outcome class %q", o.Class)
	}
	return nil
}

// FailureBoundary is a normalized boundary condition where an approach stops.
type FailureBoundary struct {
	ID                 string
	ApproachRevisionID string
	Condition          string
}

func (b FailureBoundary) Validate() error {
	if err := ValidateFailureBoundaryID(b.ID); err != nil {
		return err
	}
	if err := ValidateApproachRevisionID(b.ApproachRevisionID); err != nil {
		return err
	}
	if strings.TrimSpace(b.Condition) == "" {
		return errors.New("failure boundary condition is required")
	}
	return nil
}

// SourceSupport records the provenance strength of one normalized field.
//
// FieldPath identifies which normalized field it justifies (e.g.
// "mechanism.locality", "outcome.boundary"). Locator points into the source
// snapshot bytes where available. Confidence is an optional ordinal hint.
type SourceSupport struct {
	ApproachRevisionID string
	SnapshotID         string
	FieldPath          string
	SupportKind        SupportKind
	Locator            string
	Confidence         string
}

func (s SourceSupport) Validate() error {
	if err := ValidateApproachRevisionID(s.ApproachRevisionID); err != nil {
		return err
	}
	if err := ValidateSnapshotID(s.SnapshotID); err != nil {
		return err
	}
	if strings.TrimSpace(s.FieldPath) == "" {
		return errors.New("source support field_path is required")
	}
	if !s.SupportKind.Valid() {
		return fmt.Errorf("invalid support kind %q", s.SupportKind)
	}
	return nil
}
