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
