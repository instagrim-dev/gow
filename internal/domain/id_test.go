package domain

import (
	"testing"
	"time"
)

func TestGeneratedIDsValidate(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	problemID := NewProblemID(now)
	runID := NewRunID(now)
	sourceID := NewSourceID(now)
	snapshotID := NewSnapshotID(now)

	if err := ValidateProblemID(problemID); err != nil {
		t.Fatalf("ValidateProblemID() error = %v", err)
	}
	if err := ValidateRunID(runID); err != nil {
		t.Fatalf("ValidateRunID() error = %v", err)
	}
	if err := ValidateSourceID(sourceID); err != nil {
		t.Fatalf("ValidateSourceID() error = %v", err)
	}
	if err := ValidateSnapshotID(snapshotID); err != nil {
		t.Fatalf("ValidateSnapshotID() error = %v", err)
	}
}

func TestValidateRejectsCrossClassIDs(t *testing.T) {
	t.Parallel()

	runID := NewRunID(time.Now().UTC())
	if err := ValidateProblemID(runID); err == nil {
		t.Fatal("ValidateProblemID() succeeded for run ID")
	}
}
