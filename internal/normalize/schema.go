// Package normalize defines the provider-independent normalization contract:
// the request handed to a Normalizer and the structured result it must return.
//
// This package owns schema validation for normalized output. It depends only
// on the domain layer, never on any provider SDK, so the contract stays
// portable across provider adapters and deterministic fixtures.
package normalize

import (
	"errors"
	"fmt"
	"strings"

	"github.com/instagrim-dev/newf/internal/domain"
)

// SchemaVersion is the current normalization contract version. It is persisted
// on every revision and provider invocation so re-normalization under a new
// contract produces a new revision rather than silently overwriting history.
const SchemaVersion = "normalize/v1"

// Request is the material and contract handed to a Normalizer. It contains only
// what this operator needs: the source snapshot identity, its decoded text, and
// the schema version to target. Providers must not need any other domain state.
type Request struct {
	ProblemID     string
	SnapshotID    string
	MediaType     string
	Content       string
	SchemaVersion string
}

// FieldSupport is the per-field provenance the provider must attach so the
// pipeline can persist explicit-vs-inferred distinctions. FieldPath uses dotted
// paths like "mechanism.locality" or "outcome.boundary".
type FieldSupport struct {
	FieldPath   string             `json:"field_path"`
	SupportKind domain.SupportKind `json:"support_kind"`
	Locator     string             `json:"locator,omitempty"`
	Confidence  string             `json:"confidence,omitempty"`
}

// Mechanism is the typed mechanistic reading extracted for one approach.
type Mechanism struct {
	Representations  []string                `json:"representations,omitempty"`
	Assumptions      []string                `json:"assumptions,omitempty"`
	Operators        []string                `json:"operators,omitempty"`
	Preserves        []string                `json:"preserves,omitempty"`
	Breaks           []string                `json:"breaks,omitempty"`
	AuxiliaryObjects []string                `json:"auxiliary_objects,omitempty"`
	Locality         domain.Locality         `json:"locality"`
	ConstructionMode domain.ConstructionMode `json:"construction_mode"`
	UncertaintyMode  domain.UncertaintyMode  `json:"uncertainty_mode"`
	Notes            string                  `json:"notes,omitempty"`
}

// Outcome is the normalized result and boundary for one approach.
type Outcome struct {
	Class              domain.OutcomeClass `json:"class"`
	BoundaryStatement  string              `json:"boundary_statement,omitempty"`
	BoundaryConditions []string            `json:"boundary_conditions,omitempty"`
	Notes              string              `json:"notes,omitempty"`
}

// Approach is one materially distinct approach extracted from the snapshot.
type Approach struct {
	// LogicalIdentity stabilizes approach identity across re-normalizations of
	// the same snapshot so revisions attach to the same logical Approach.
	LogicalIdentity string         `json:"logical_identity"`
	Label           string         `json:"label"`
	Description     string         `json:"description,omitempty"`
	Mechanism       Mechanism      `json:"mechanism"`
	Outcome         Outcome        `json:"outcome"`
	Support         []FieldSupport `json:"support,omitempty"`
}

// Result is the structured provider output. A provider that cannot process the
// snapshot must set Skipped with a typed SkipReason rather than guessing.
type Result struct {
	SchemaVersion string     `json:"schema_version"`
	Approaches    []Approach `json:"approaches"`
	Skipped       bool       `json:"skipped,omitempty"`
	SkipReason    SkipReason `json:"skip_reason,omitempty"`
	Warnings      []string   `json:"warnings,omitempty"`
}

// SkipReason is a typed reason a snapshot could not be normalized. These are
// distinct from provider transport failures.
type SkipReason string

const (
	SkipUnsupportedContentRepresentation SkipReason = "unsupported_content_representation"
	SkipTextExtractionRequired           SkipReason = "text_extraction_required"
	SkipNoApproachesFound                SkipReason = "no_approaches_found"
)

func (r SkipReason) Valid() bool {
	switch r {
	case SkipUnsupportedContentRepresentation, SkipTextExtractionRequired, SkipNoApproachesFound:
		return true
	default:
		return false
	}
}

// ErrSchemaViolation marks structurally invalid provider output. It is a
// semantic normalization failure, not a provider transport failure.
var ErrSchemaViolation = errors.New("normalization schema violation")

func schemaErr(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrSchemaViolation, fmt.Sprintf(format, args...))
}

// Validate enforces the structural contract on provider output. It does not
// attempt to verify the correctness of the interpretation, only that the shape
// and enum values are well-formed and internally consistent.
func (r Result) Validate() error {
	if strings.TrimSpace(r.SchemaVersion) == "" {
		return schemaErr("schema_version is required")
	}
	if r.Skipped {
		if !r.SkipReason.Valid() {
			return schemaErr("invalid skip_reason %q", r.SkipReason)
		}
		if len(r.Approaches) != 0 {
			return schemaErr("skipped result must not contain approaches")
		}
		return nil
	}
	if r.SkipReason != "" {
		return schemaErr("skip_reason set on non-skipped result")
	}
	if len(r.Approaches) == 0 {
		return schemaErr("non-skipped result must contain at least one approach")
	}
	seenIdentity := map[string]struct{}{}
	for i, approach := range r.Approaches {
		if err := approach.validate(i); err != nil {
			return err
		}
		if _, dup := seenIdentity[approach.LogicalIdentity]; dup {
			return schemaErr("duplicate approach logical_identity %q", approach.LogicalIdentity)
		}
		seenIdentity[approach.LogicalIdentity] = struct{}{}
	}
	return nil
}

func (a Approach) validate(index int) error {
	if strings.TrimSpace(a.LogicalIdentity) == "" {
		return schemaErr("approach[%d] logical_identity is required", index)
	}
	if strings.TrimSpace(a.Label) == "" {
		return schemaErr("approach[%d] label is required", index)
	}
	if !a.Mechanism.Locality.Valid() {
		return schemaErr("approach[%d] invalid locality %q", index, a.Mechanism.Locality)
	}
	if !a.Mechanism.ConstructionMode.Valid() {
		return schemaErr("approach[%d] invalid construction_mode %q", index, a.Mechanism.ConstructionMode)
	}
	if !a.Mechanism.UncertaintyMode.Valid() {
		return schemaErr("approach[%d] invalid uncertainty_mode %q", index, a.Mechanism.UncertaintyMode)
	}
	if !a.Outcome.Class.Valid() {
		return schemaErr("approach[%d] invalid outcome class %q", index, a.Outcome.Class)
	}
	for j, support := range a.Support {
		if strings.TrimSpace(support.FieldPath) == "" {
			return schemaErr("approach[%d].support[%d] field_path is required", index, j)
		}
		if !support.SupportKind.Valid() {
			return schemaErr("approach[%d].support[%d] invalid support_kind %q", index, j, support.SupportKind)
		}
	}
	// A normalized field has exactly one provenance strength (explicit /
	// inferred / unsupported). Two support entries for the same field_path
	// would otherwise silently collapse at persistence and could blur the
	// epistemic boundary the operator must never blur, so reject the conflict
	// at the contract boundary rather than dropping one row.
	seenFieldPath := make(map[string]struct{}, len(a.Support))
	for j, support := range a.Support {
		key := strings.TrimSpace(support.FieldPath)
		if _, dup := seenFieldPath[key]; dup {
			return schemaErr("approach[%d].support[%d] duplicate field_path %q", index, j, key)
		}
		seenFieldPath[key] = struct{}{}
	}
	return nil
}
