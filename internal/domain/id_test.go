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

	normalizationRevisionID := NewNormalizationRevisionID(now)
	providerInvocationID := NewProviderInvocationID(now)
	approachID := NewApproachID(now)
	approachRevisionID := NewApproachRevisionID(now)
	mechanismID := NewMechanismID(now)
	outcomeID := NewOutcomeID(now)
	failureBoundaryID := NewFailureBoundaryID(now)

	if err := ValidateNormalizationRevisionID(normalizationRevisionID); err != nil {
		t.Fatalf("ValidateNormalizationRevisionID() error = %v", err)
	}
	if err := ValidateProviderInvocationID(providerInvocationID); err != nil {
		t.Fatalf("ValidateProviderInvocationID() error = %v", err)
	}
	if err := ValidateApproachID(approachID); err != nil {
		t.Fatalf("ValidateApproachID() error = %v", err)
	}
	if err := ValidateApproachRevisionID(approachRevisionID); err != nil {
		t.Fatalf("ValidateApproachRevisionID() error = %v", err)
	}
	if err := ValidateMechanismID(mechanismID); err != nil {
		t.Fatalf("ValidateMechanismID() error = %v", err)
	}
	if err := ValidateOutcomeID(outcomeID); err != nil {
		t.Fatalf("ValidateOutcomeID() error = %v", err)
	}
	if err := ValidateFailureBoundaryID(failureBoundaryID); err != nil {
		t.Fatalf("ValidateFailureBoundaryID() error = %v", err)
	}

	clusterRunID := NewClusterRunID(now)
	mechanismClusterID := NewMechanismClusterID(now)
	failureSpaceID := NewFailureSpaceID(now)

	if err := ValidateClusterRunID(clusterRunID); err != nil {
		t.Fatalf("ValidateClusterRunID() error = %v", err)
	}
	if err := ValidateMechanismClusterID(mechanismClusterID); err != nil {
		t.Fatalf("ValidateMechanismClusterID() error = %v", err)
	}
	if err := ValidateFailureSpaceID(failureSpaceID); err != nil {
		t.Fatalf("ValidateFailureSpaceID() error = %v", err)
	}

	invariantRevisionID := NewInvariantRevisionID(now)
	candidateInvariantID := NewCandidateInvariantID(now)

	if err := ValidateInvariantRevisionID(invariantRevisionID); err != nil {
		t.Fatalf("ValidateInvariantRevisionID() error = %v", err)
	}
	if err := ValidateCandidateInvariantID(candidateInvariantID); err != nil {
		t.Fatalf("ValidateCandidateInvariantID() error = %v", err)
	}

	frontierGenerationRunID := NewFrontierGenerationRunID(now)
	frontierProposalID := NewFrontierProposalID(now)

	if err := ValidateFrontierGenerationRunID(frontierGenerationRunID); err != nil {
		t.Fatalf("ValidateFrontierGenerationRunID() error = %v", err)
	}
	if err := ValidateFrontierProposalID(frontierProposalID); err != nil {
		t.Fatalf("ValidateFrontierProposalID() error = %v", err)
	}
}

func TestValidateRejectsCrossClassIDs(t *testing.T) {
	t.Parallel()

	runID := NewRunID(time.Now().UTC())
	if err := ValidateProblemID(runID); err == nil {
		t.Fatal("ValidateProblemID() succeeded for run ID")
	}

	approachID := NewApproachID(time.Now().UTC())
	if err := ValidateApproachRevisionID(approachID); err == nil {
		t.Fatal("ValidateApproachRevisionID() succeeded for approach ID")
	}
	if err := ValidateMechanismID(approachID); err == nil {
		t.Fatal("ValidateMechanismID() succeeded for approach ID")
	}

	snapshotID := NewSnapshotID(time.Now().UTC())
	if err := ValidateNormalizationRevisionID(snapshotID); err == nil {
		t.Fatal("ValidateNormalizationRevisionID() succeeded for snapshot ID")
	}

	clusterRunID := NewClusterRunID(time.Now().UTC())
	if err := ValidateMechanismClusterID(clusterRunID); err == nil {
		t.Fatal("ValidateMechanismClusterID() succeeded for cluster run ID")
	}
	if err := ValidateFailureSpaceID(clusterRunID); err == nil {
		t.Fatal("ValidateFailureSpaceID() succeeded for cluster run ID")
	}

	invariantRevisionID := NewInvariantRevisionID(time.Now().UTC())
	if err := ValidateCandidateInvariantID(invariantRevisionID); err == nil {
		t.Fatal("ValidateCandidateInvariantID() succeeded for invariant revision ID")
	}
	if err := ValidateInvariantRevisionID(clusterRunID); err == nil {
		t.Fatal("ValidateInvariantRevisionID() succeeded for cluster run ID")
	}

	frontierGenerationRunID := NewFrontierGenerationRunID(time.Now().UTC())
	if err := ValidateFrontierProposalID(frontierGenerationRunID); err == nil {
		t.Fatal("ValidateFrontierProposalID() succeeded for frontier generation run ID")
	}
	if err := ValidateFrontierGenerationRunID(clusterRunID); err == nil {
		t.Fatal("ValidateFrontierGenerationRunID() succeeded for cluster run ID")
	}
}
