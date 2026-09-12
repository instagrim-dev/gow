package domain

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

const (
	ProblemIDPrefix               = "prb_"
	RunIDPrefix                   = "run_"
	SourceIDPrefix                = "src_"
	SnapshotIDPrefix              = "snap_"
	NormalizationRevisionIDPrefix = "nrev_"
	ProviderInvocationIDPrefix    = "pinv_"
	ApproachIDPrefix              = "app_"
	ApproachRevisionIDPrefix      = "apr_"
	MechanismIDPrefix             = "mech_"
	OutcomeIDPrefix               = "out_"
	FailureBoundaryIDPrefix       = "fbnd_"
	MechanismSignatureIDPrefix    = "msig_"
	ComparisonRunIDPrefix         = "cmp_"
	ClusterRunIDPrefix            = "clr_"
	MechanismClusterIDPrefix      = "mcl_"
	FailureSpaceIDPrefix          = "fsp_"
	InvariantRevisionIDPrefix     = "ivr_"
	CandidateInvariantIDPrefix    = "inv_"
	InvariantChallengeIDPrefix    = "chl_"
	SyntheticArtifactIDPrefix     = "syn_"
	FrontierGenerationRunIDPrefix = "fgr_"
	FrontierProposalIDPrefix      = "fpr_"
	EvaluationRunIDPrefix         = "evr_"
	EvaluationIDPrefix            = "evl_"
	EvaluationMetricIDPrefix      = "evm_"
	EvidenceAdmissionIDPrefix     = "evad_"
	SuccessRevisionIDPrefix       = "svr_"
	SuccessInvariantIDPrefix      = "sinv_"
	HoldoutSetIDPrefix            = "hset_"
	LeakageCheckIDPrefix          = "lkc_"
	ExperimentIDPrefix            = "exp_"
	SearchPolicyRevisionIDPrefix  = "spr_"
	SearchPolicyDirectiveIDPrefix = "spd_"
	InterpretationClaimIDPrefix   = "icl_"
)

var (
	ErrInvalidProblemID               = errors.New("invalid problem id")
	ErrInvalidRunID                   = errors.New("invalid run id")
	ErrInvalidSourceID                = errors.New("invalid source id")
	ErrInvalidSnapshotID              = errors.New("invalid snapshot id")
	ErrInvalidNormalizationRevisionID = errors.New("invalid normalization revision id")
	ErrInvalidProviderInvocationID    = errors.New("invalid provider invocation id")
	ErrInvalidApproachID              = errors.New("invalid approach id")
	ErrInvalidApproachRevisionID      = errors.New("invalid approach revision id")
	ErrInvalidMechanismID             = errors.New("invalid mechanism id")
	ErrInvalidOutcomeID               = errors.New("invalid outcome id")
	ErrInvalidFailureBoundaryID       = errors.New("invalid failure boundary id")
	ErrInvalidMechanismSignatureID    = errors.New("invalid mechanism signature id")
	ErrInvalidComparisonRunID         = errors.New("invalid comparison run id")
	ErrInvalidClusterRunID            = errors.New("invalid cluster run id")
	ErrInvalidMechanismClusterID      = errors.New("invalid mechanism cluster id")
	ErrInvalidFailureSpaceID          = errors.New("invalid failure space id")
	ErrInvalidInvariantRevisionID     = errors.New("invalid invariant revision id")
	ErrInvalidCandidateInvariantID    = errors.New("invalid candidate invariant id")
	ErrInvalidInvariantChallengeID    = errors.New("invalid invariant challenge id")
	ErrInvalidSyntheticArtifactID     = errors.New("invalid synthetic artifact id")
	ErrInvalidFrontierGenerationRunID = errors.New("invalid frontier generation run id")
	ErrInvalidFrontierProposalID      = errors.New("invalid frontier proposal id")
	ErrInvalidEvaluationRunID         = errors.New("invalid evaluation run id")
	ErrInvalidEvaluationID            = errors.New("invalid evaluation id")
	ErrInvalidEvaluationMetricID      = errors.New("invalid evaluation metric id")
	ErrInvalidEvidenceAdmissionID     = errors.New("invalid evidence admission id")
	ErrInvalidSuccessRevisionID       = errors.New("invalid success revision id")
	ErrInvalidSuccessInvariantID      = errors.New("invalid success invariant id")
	ErrInvalidHoldoutSetID            = errors.New("invalid holdout set id")
	ErrInvalidLeakageCheckID          = errors.New("invalid leakage check id")
	ErrInvalidExperimentID            = errors.New("invalid experiment id")
	ErrInvalidSearchPolicyRevisionID  = errors.New("invalid search policy revision id")
	ErrInvalidSearchPolicyDirectiveID = errors.New("invalid search policy directive id")

	entropyMu sync.Mutex
	entropy   = ulid.Monotonic(defaultEntropy(), 0)
)

func NewProblemID(now time.Time) string {
	return newID(ProblemIDPrefix, now)
}

func NewRunID(now time.Time) string {
	return newID(RunIDPrefix, now)
}

func ValidateProblemID(id string) error {
	return validateID(id, ProblemIDPrefix, ErrInvalidProblemID)
}

func ValidateRunID(id string) error {
	return validateID(id, RunIDPrefix, ErrInvalidRunID)
}

func NewSourceID(now time.Time) string {
	return newID(SourceIDPrefix, now)
}

func NewSnapshotID(now time.Time) string {
	return newID(SnapshotIDPrefix, now)
}

func ValidateSourceID(id string) error {
	return validateID(id, SourceIDPrefix, ErrInvalidSourceID)
}

func ValidateSnapshotID(id string) error {
	return validateID(id, SnapshotIDPrefix, ErrInvalidSnapshotID)
}

func NewNormalizationRevisionID(now time.Time) string {
	return newID(NormalizationRevisionIDPrefix, now)
}

func NewProviderInvocationID(now time.Time) string {
	return newID(ProviderInvocationIDPrefix, now)
}

func NewApproachID(now time.Time) string {
	return newID(ApproachIDPrefix, now)
}

func NewApproachRevisionID(now time.Time) string {
	return newID(ApproachRevisionIDPrefix, now)
}

func NewMechanismID(now time.Time) string {
	return newID(MechanismIDPrefix, now)
}

func NewOutcomeID(now time.Time) string {
	return newID(OutcomeIDPrefix, now)
}

func NewFailureBoundaryID(now time.Time) string {
	return newID(FailureBoundaryIDPrefix, now)
}

func NewMechanismSignatureID(now time.Time) string {
	return newID(MechanismSignatureIDPrefix, now)
}

func NewComparisonRunID(now time.Time) string {
	return newID(ComparisonRunIDPrefix, now)
}

func ValidateMechanismSignatureID(id string) error {
	return validateID(id, MechanismSignatureIDPrefix, ErrInvalidMechanismSignatureID)
}

func ValidateComparisonRunID(id string) error {
	return validateID(id, ComparisonRunIDPrefix, ErrInvalidComparisonRunID)
}

func NewClusterRunID(now time.Time) string {
	return newID(ClusterRunIDPrefix, now)
}

func NewMechanismClusterID(now time.Time) string {
	return newID(MechanismClusterIDPrefix, now)
}

func NewFailureSpaceID(now time.Time) string {
	return newID(FailureSpaceIDPrefix, now)
}

func ValidateClusterRunID(id string) error {
	return validateID(id, ClusterRunIDPrefix, ErrInvalidClusterRunID)
}

func ValidateMechanismClusterID(id string) error {
	return validateID(id, MechanismClusterIDPrefix, ErrInvalidMechanismClusterID)
}

func ValidateFailureSpaceID(id string) error {
	return validateID(id, FailureSpaceIDPrefix, ErrInvalidFailureSpaceID)
}

func NewInvariantRevisionID(now time.Time) string {
	return newID(InvariantRevisionIDPrefix, now)
}

func NewCandidateInvariantID(now time.Time) string {
	return newID(CandidateInvariantIDPrefix, now)
}

func NewInterpretationClaimID(now time.Time) string {
	return newID(InterpretationClaimIDPrefix, now)
}

func ValidateInvariantRevisionID(id string) error {
	return validateID(id, InvariantRevisionIDPrefix, ErrInvalidInvariantRevisionID)
}

func ValidateCandidateInvariantID(id string) error {
	return validateID(id, CandidateInvariantIDPrefix, ErrInvalidCandidateInvariantID)
}

func NewInvariantChallengeID(now time.Time) string {
	return newID(InvariantChallengeIDPrefix, now)
}

func NewSyntheticArtifactID(now time.Time) string {
	return newID(SyntheticArtifactIDPrefix, now)
}

func ValidateInvariantChallengeID(id string) error {
	return validateID(id, InvariantChallengeIDPrefix, ErrInvalidInvariantChallengeID)
}

func ValidateSyntheticArtifactID(id string) error {
	return validateID(id, SyntheticArtifactIDPrefix, ErrInvalidSyntheticArtifactID)
}

func NewFrontierGenerationRunID(now time.Time) string {
	return newID(FrontierGenerationRunIDPrefix, now)
}

func NewFrontierProposalID(now time.Time) string {
	return newID(FrontierProposalIDPrefix, now)
}

func ValidateFrontierGenerationRunID(id string) error {
	return validateID(id, FrontierGenerationRunIDPrefix, ErrInvalidFrontierGenerationRunID)
}

func ValidateFrontierProposalID(id string) error {
	return validateID(id, FrontierProposalIDPrefix, ErrInvalidFrontierProposalID)
}

func NewEvaluationRunID(now time.Time) string {
	return newID(EvaluationRunIDPrefix, now)
}

func NewEvaluationID(now time.Time) string {
	return newID(EvaluationIDPrefix, now)
}

func NewEvaluationMetricID(now time.Time) string {
	return newID(EvaluationMetricIDPrefix, now)
}

func ValidateEvaluationRunID(id string) error {
	return validateID(id, EvaluationRunIDPrefix, ErrInvalidEvaluationRunID)
}

func ValidateEvaluationID(id string) error {
	return validateID(id, EvaluationIDPrefix, ErrInvalidEvaluationID)
}

func ValidateEvaluationMetricID(id string) error {
	return validateID(id, EvaluationMetricIDPrefix, ErrInvalidEvaluationMetricID)
}

func NewEvidenceAdmissionID(now time.Time) string {
	return newID(EvidenceAdmissionIDPrefix, now)
}

func ValidateEvidenceAdmissionID(id string) error {
	return validateID(id, EvidenceAdmissionIDPrefix, ErrInvalidEvidenceAdmissionID)
}

func NewSuccessRevisionID(now time.Time) string {
	return newID(SuccessRevisionIDPrefix, now)
}

func NewSuccessInvariantID(now time.Time) string {
	return newID(SuccessInvariantIDPrefix, now)
}

func ValidateSuccessRevisionID(id string) error {
	return validateID(id, SuccessRevisionIDPrefix, ErrInvalidSuccessRevisionID)
}

func ValidateSuccessInvariantID(id string) error {
	return validateID(id, SuccessInvariantIDPrefix, ErrInvalidSuccessInvariantID)
}

func NewHoldoutSetID(now time.Time) string {
	return newID(HoldoutSetIDPrefix, now)
}

func NewLeakageCheckID(now time.Time) string {
	return newID(LeakageCheckIDPrefix, now)
}

func NewExperimentID(now time.Time) string {
	return newID(ExperimentIDPrefix, now)
}

func ValidateHoldoutSetID(id string) error {
	return validateID(id, HoldoutSetIDPrefix, ErrInvalidHoldoutSetID)
}

func ValidateLeakageCheckID(id string) error {
	return validateID(id, LeakageCheckIDPrefix, ErrInvalidLeakageCheckID)
}

func ValidateExperimentID(id string) error {
	return validateID(id, ExperimentIDPrefix, ErrInvalidExperimentID)
}

func NewSearchPolicyRevisionID(now time.Time) string {
	return newID(SearchPolicyRevisionIDPrefix, now)
}

func NewSearchPolicyDirectiveID(now time.Time) string {
	return newID(SearchPolicyDirectiveIDPrefix, now)
}

func ValidateSearchPolicyRevisionID(id string) error {
	return validateID(id, SearchPolicyRevisionIDPrefix, ErrInvalidSearchPolicyRevisionID)
}

func ValidateSearchPolicyDirectiveID(id string) error {
	return validateID(id, SearchPolicyDirectiveIDPrefix, ErrInvalidSearchPolicyDirectiveID)
}

func ValidateNormalizationRevisionID(id string) error {
	return validateID(id, NormalizationRevisionIDPrefix, ErrInvalidNormalizationRevisionID)
}

func ValidateProviderInvocationID(id string) error {
	return validateID(id, ProviderInvocationIDPrefix, ErrInvalidProviderInvocationID)
}

func ValidateApproachID(id string) error {
	return validateID(id, ApproachIDPrefix, ErrInvalidApproachID)
}

func ValidateApproachRevisionID(id string) error {
	return validateID(id, ApproachRevisionIDPrefix, ErrInvalidApproachRevisionID)
}

func ValidateMechanismID(id string) error {
	return validateID(id, MechanismIDPrefix, ErrInvalidMechanismID)
}

func ValidateOutcomeID(id string) error {
	return validateID(id, OutcomeIDPrefix, ErrInvalidOutcomeID)
}

func ValidateFailureBoundaryID(id string) error {
	return validateID(id, FailureBoundaryIDPrefix, ErrInvalidFailureBoundaryID)
}

func newID(prefix string, now time.Time) string {
	entropyMu.Lock()
	defer entropyMu.Unlock()

	return prefix + ulid.MustNew(ulid.Timestamp(now.UTC()), entropy).String()
}

func validateID(id, prefix string, sentinel error) error {
	if !strings.HasPrefix(id, prefix) {
		return fmt.Errorf("%w: expected prefix %q", sentinel, prefix)
	}

	raw := strings.TrimPrefix(id, prefix)
	if _, err := ulid.ParseStrict(raw); err != nil {
		return fmt.Errorf("%w: %s", sentinel, err)
	}

	return nil
}

func defaultEntropy() io.Reader {
	return rand.Reader
}
