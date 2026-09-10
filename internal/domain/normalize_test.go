package domain

import (
	"testing"
	"time"
)

func nowForTest() time.Time {
	return time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
}

func TestOutcomeClassValidity(t *testing.T) {
	t.Parallel()
	for _, class := range []OutcomeClass{OutcomeUnknown, OutcomeFailure, OutcomePartialFailure, OutcomePartialSuccess, OutcomeSuccess} {
		if !class.Valid() {
			t.Fatalf("OutcomeClass %q should be valid", class)
		}
	}
	if OutcomeClass("bogus").Valid() {
		t.Fatal("bogus outcome class should be invalid")
	}
}

func TestSupportKindValidity(t *testing.T) {
	t.Parallel()
	for _, kind := range []SupportKind{SupportExplicit, SupportInferred, SupportUnsupported} {
		if !kind.Valid() {
			t.Fatalf("SupportKind %q should be valid", kind)
		}
	}
	if SupportKind("guessed").Valid() {
		t.Fatal("guessed support kind should be invalid")
	}
}

func TestProviderInvocationRejectsNonNormalizeRole(t *testing.T) {
	t.Parallel()
	invocation := ProviderInvocation{
		ID:           NewProviderInvocationID(nowForTest()),
		RunID:        NewRunID(nowForTest()),
		Role:         ProviderRole("cluster"),
		ProviderName: "fixture",
		RequestHash:  "hash",
		CreatedAt:    nowForTest(),
	}
	if err := invocation.Validate(); err == nil {
		t.Fatal("expected non-normalize role to be rejected")
	}
}

func TestNormalizationRevisionRequiresSchemaVersion(t *testing.T) {
	t.Parallel()
	revision := NormalizationRevision{
		ID:                   NewNormalizationRevisionID(nowForTest()),
		ProblemID:            NewProblemID(nowForTest()),
		RunID:                NewRunID(nowForTest()),
		SnapshotID:           NewSnapshotID(nowForTest()),
		ProviderInvocationID: NewProviderInvocationID(nowForTest()),
		ConfigHash:           "hash",
		Status:               NormalizationStatusSucceeded,
		CreatedAt:            nowForTest(),
	}
	if err := revision.Validate(); err == nil {
		t.Fatal("expected missing schema_version to be rejected")
	}
	revision.SchemaVersion = "normalize/v1"
	if err := revision.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestSourceSupportValidation(t *testing.T) {
	t.Parallel()
	support := SourceSupport{
		ApproachRevisionID: NewApproachRevisionID(nowForTest()),
		SnapshotID:         NewSnapshotID(nowForTest()),
		FieldPath:          "mechanism.locality",
		SupportKind:        SupportInferred,
	}
	if err := support.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	support.SupportKind = "guessed"
	if err := support.Validate(); err == nil {
		t.Fatal("expected invalid support kind to be rejected")
	}
}
