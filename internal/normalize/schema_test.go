package normalize

import (
	"errors"
	"testing"

	"github.com/instagrim-dev/newf/internal/domain"
)

func validApproach() Approach {
	return Approach{
		LogicalIdentity: "erdos-straus/modular-residue-cover",
		Label:           "modular residue-cover construction",
		Mechanism: Mechanism{
			Locality:         domain.LocalityLocal,
			ConstructionMode: domain.ConstructionConstructive,
			UncertaintyMode:  domain.UncertaintyDeterministic,
		},
		Outcome: Outcome{Class: domain.OutcomePartialFailure},
		Support: []FieldSupport{
			{FieldPath: "outcome.class", SupportKind: domain.SupportExplicit},
		},
	}
}

func TestResultValidateAcceptsWellFormed(t *testing.T) {
	t.Parallel()
	result := Result{SchemaVersion: SchemaVersion, Approaches: []Approach{validApproach()}}
	if err := result.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestResultValidateRejectsMissingSchemaVersion(t *testing.T) {
	t.Parallel()
	result := Result{Approaches: []Approach{validApproach()}}
	if err := result.Validate(); !errors.Is(err, ErrSchemaViolation) {
		t.Fatalf("Validate() error = %v, want schema violation", err)
	}
}

func TestResultValidateRejectsEmptyNonSkipped(t *testing.T) {
	t.Parallel()
	result := Result{SchemaVersion: SchemaVersion}
	if err := result.Validate(); !errors.Is(err, ErrSchemaViolation) {
		t.Fatalf("Validate() error = %v, want schema violation", err)
	}
}

func TestResultValidateRejectsInvalidEnum(t *testing.T) {
	t.Parallel()
	approach := validApproach()
	approach.Mechanism.Locality = "sideways"
	result := Result{SchemaVersion: SchemaVersion, Approaches: []Approach{approach}}
	if err := result.Validate(); !errors.Is(err, ErrSchemaViolation) {
		t.Fatalf("Validate() error = %v, want schema violation", err)
	}
}

func TestResultValidateRejectsInvalidSupportKind(t *testing.T) {
	t.Parallel()
	approach := validApproach()
	approach.Support = []FieldSupport{{FieldPath: "mechanism.locality", SupportKind: "guessed"}}
	result := Result{SchemaVersion: SchemaVersion, Approaches: []Approach{approach}}
	if err := result.Validate(); !errors.Is(err, ErrSchemaViolation) {
		t.Fatalf("Validate() error = %v, want schema violation", err)
	}
}

func TestResultValidateRejectsDuplicateIdentity(t *testing.T) {
	t.Parallel()
	result := Result{SchemaVersion: SchemaVersion, Approaches: []Approach{validApproach(), validApproach()}}
	if err := result.Validate(); !errors.Is(err, ErrSchemaViolation) {
		t.Fatalf("Validate() error = %v, want schema violation", err)
	}
}

func TestResultValidateRejectsDuplicateFieldPath(t *testing.T) {
	t.Parallel()
	approach := validApproach()
	// A field carries exactly one provenance strength; two support entries for
	// the same field_path (here explicit vs unsupported) would blur the
	// epistemic boundary and must be rejected at the contract boundary.
	approach.Support = []FieldSupport{
		{FieldPath: "outcome.class", SupportKind: domain.SupportExplicit},
		{FieldPath: "outcome.class", SupportKind: domain.SupportUnsupported},
	}
	result := Result{SchemaVersion: SchemaVersion, Approaches: []Approach{approach}}
	if err := result.Validate(); !errors.Is(err, ErrSchemaViolation) {
		t.Fatalf("Validate() error = %v, want schema violation for duplicate field_path", err)
	}
}

func TestResultValidateSkippedRequiresReason(t *testing.T) {
	t.Parallel()
	result := Result{SchemaVersion: SchemaVersion, Skipped: true}
	if err := result.Validate(); !errors.Is(err, ErrSchemaViolation) {
		t.Fatalf("Validate() error = %v, want schema violation for missing skip_reason", err)
	}
}

func TestResultValidateSkippedRejectsApproaches(t *testing.T) {
	t.Parallel()
	result := Result{
		SchemaVersion: SchemaVersion,
		Skipped:       true,
		SkipReason:    SkipUnsupportedContentRepresentation,
		Approaches:    []Approach{validApproach()},
	}
	if err := result.Validate(); !errors.Is(err, ErrSchemaViolation) {
		t.Fatalf("Validate() error = %v, want schema violation for approaches on skip", err)
	}
}

func TestResultValidateSkippedAcceptsTypedReason(t *testing.T) {
	t.Parallel()
	result := Result{SchemaVersion: SchemaVersion, Skipped: true, SkipReason: SkipTextExtractionRequired}
	if err := result.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

// Field-completeness declarations are strong claims: admissible only with a
// typed scope, a non-empty basis, valid set-field keys, and non-vacuous values
// (v25/v26). Validation admits the CLAIM; acceptance is decided by code.
func TestResultValidateFieldCompletenessRequiresBasisAndScope(t *testing.T) {
	t.Parallel()
	approach := validApproach()
	approach.Mechanism.FieldCompleteness = map[string]string{"preserves": "complete"}
	approach.Mechanism.CompletenessScope = "declared_payload"
	result := Result{SchemaVersion: SchemaVersion, Approaches: []Approach{approach}}
	if err := result.Validate(); !errors.Is(err, ErrSchemaViolation) {
		t.Fatalf("a basis-free completeness declaration must be rejected, got %v", err)
	}

	approach.Mechanism.CompletenessBasis = "all entries of the declared payload's preserves list were parsed"
	approach.Mechanism.CompletenessScope = ""
	result = Result{SchemaVersion: SchemaVersion, Approaches: []Approach{approach}}
	if err := result.Validate(); !errors.Is(err, ErrSchemaViolation) {
		t.Fatalf("a scope-free completeness declaration must be rejected, got %v", err)
	}

	approach.Mechanism.CompletenessScope = "declared_payload"
	result = Result{SchemaVersion: SchemaVersion, Approaches: []Approach{approach}}
	if err := result.Validate(); err != nil {
		t.Fatalf("a justified, scoped declaration must validate, got %v", err)
	}
}

func TestResultValidateFieldCompletenessRejectsBadKeyAndVacuousValue(t *testing.T) {
	t.Parallel()
	approach := validApproach()
	approach.Mechanism.CompletenessBasis = "basis"
	approach.Mechanism.CompletenessScope = "declared_payload"
	approach.Mechanism.FieldCompleteness = map[string]string{"outcome": "complete"} // not a set field
	if err := (Result{SchemaVersion: SchemaVersion, Approaches: []Approach{approach}}).Validate(); !errors.Is(err, ErrSchemaViolation) {
		t.Fatalf("a non-set-field key must be rejected, got %v", err)
	}

	approach.Mechanism.FieldCompleteness = map[string]string{"preserves": "unobserved"} // vacuous
	if err := (Result{SchemaVersion: SchemaVersion, Approaches: []Approach{approach}}).Validate(); !errors.Is(err, ErrSchemaViolation) {
		t.Fatalf("a vacuous 'unobserved' declaration must be rejected, got %v", err)
	}
}
