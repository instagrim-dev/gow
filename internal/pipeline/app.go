package pipeline

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/instagrim-dev/newf/internal/config"
	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/provider"
	"github.com/instagrim-dev/newf/internal/store"
)

type App struct {
	version             string
	now                 func() time.Time
	getwd               func() (string, error)
	stdin               io.Reader
	openStoreFn         func(context.Context, string) (string, problemStore, error)
	projectRevisionFn   func() string
	normalizers         map[string]provider.Normalizer
	invariantMinerFn    provider.InvariantMiner
	challengerFn        provider.Challenger
	successCompressorFn provider.SuccessCompressor
	generatorFn         provider.Generator
	modelVerifierFn     provider.ModelVerifier
	policyMutatorFn     provider.PolicyMutator
}

type problemStore interface {
	Close() error
	FindProblemBySlug(context.Context, string) (domain.Problem, bool, error)
	NextProblemSlug(context.Context, string) (string, error)
	CreateProblemWithRun(context.Context, domain.NewProblem, domain.NewRun) (domain.Problem, domain.Run, error)
	CreateRun(context.Context, domain.NewRun) (domain.Run, error)
	UpdateRunStatus(context.Context, string, domain.RunStatus, time.Time, *string) error
	GetProblem(context.Context, string) (domain.Problem, error)
	ListProblems(context.Context) ([]domain.Problem, error)
	GetRun(context.Context, string) (domain.Run, error)
	CreateSourceSnapshot(context.Context, store.SnapshotAdmission) (store.SnapshotAdmissionResult, error)
	ListSourcesByProblem(context.Context, string) ([]domain.Source, error)
	ListSourcesWithSnapshotStats(context.Context, string) ([]store.SourceSummary, error)
	ListSnapshotLineageForProblem(context.Context, string) ([]store.SnapshotLineageEntry, error)
	GetSource(context.Context, string) (domain.Source, error)
	GetSourceSnapshot(context.Context, string) (domain.SourceSnapshot, error)
	ListSourceSnapshots(context.Context, string) ([]domain.SourceSnapshot, error)
	FindEquivalentNormalization(context.Context, string, string, string) (store.ExistingNormalization, error)
	LatestNormalizationForSnapshot(context.Context, string) (domain.NormalizationRevision, bool, error)
	PersistNormalization(context.Context, store.NormalizationInput) (store.NormalizationWriteResult, error)
	ListApproaches(context.Context, string) ([]store.ApproachListItem, error)
	GetApproachDetail(context.Context, string) (store.ApproachDetail, error)
	GetMechanismDetail(context.Context, string) (store.ApproachDetail, error)
	AddInterpretationClaim(context.Context, store.InterpretationClaimRow, string) (store.InterpretationClaimRow, bool, error)
	ListInterpretationClaims(context.Context, string) ([]store.InterpretationClaimRow, error)
	ListInterpretationClaimsForProblem(context.Context, string) ([]store.ProblemInterpretationClaimRow, error)
	ListApproachRevisions(context.Context, string) (domain.Approach, []domain.ApproachRevision, error)
	SeedVocabulary(context.Context, store.VocabularySeedInput) error
	ListVocabularies(context.Context) ([]store.VocabularyRecord, error)
	ListTerms(context.Context, string, string) ([]store.TermRecord, error)
	ListRejected(context.Context, string) ([]string, error)
	GetTerm(context.Context, string, string) (store.TermRecord, error)
	PersistSignature(context.Context, store.SignatureRecord) (store.PersistSignatureResult, error)
	GetSignature(context.Context, string) (store.SignatureRecord, error)
	PersistComparison(context.Context, store.ComparisonRecord) error
	ListSignaturesForProblem(context.Context, string, string, string) ([]string, error)
	ListMechanismsForProblem(context.Context, string) ([]store.MechanismListItem, error)
	ListGenerationOccurrenceContents(context.Context, string) (map[string]store.OccurrenceContent, error)
	PersistClusterRun(context.Context, store.ClusterRunRecord) (store.PersistClusterRunResult, error)
	GetClusterRun(context.Context, string) (store.ClusterRunRecord, error)
	ListClusterRuns(context.Context, string) ([]store.ClusterRunRecord, error)
	LatestClusterRun(context.Context, string) (string, bool, error)
	// LatestClusterRunForVersions selects the challenge assessment population
	// (v34/S1): the newest cluster run whose signature schema and vocabulary
	// match the claim's discovery run, so widened evidence stays comparable.
	LatestClusterRunForVersions(context.Context, string, string, string) (string, bool, error)
	PersistFailureSpace(context.Context, store.FailureSpaceRecord) (store.PersistFailureSpaceResult, error)
	GetFailureSpace(context.Context, string) (store.FailureSpaceRecord, error)
	LatestFailureSpace(context.Context, string) (store.FailureSpaceRecord, bool, error)
	PersistInvariantRevision(context.Context, store.InvariantRevisionRecord) (store.PersistInvariantRevisionResult, error)
	LookupInvariantRevision(context.Context, store.InvariantReuseKey) (store.InvariantRevisionRecord, bool, error)
	GetInvariantRevision(context.Context, string) (store.InvariantRevisionRecord, error)
	ListInvariantRevisions(context.Context, string) ([]store.InvariantRevisionRecord, error)
	LatestInvariantRevision(context.Context, string) (string, bool, error)
	PersistChallengeCampaign(context.Context, store.ChallengeCampaignRecord) error
	GetInvariantState(context.Context, string) (store.InvariantStateRow, error)
	ListInvariantStates(context.Context, string, string) ([]store.InvariantStateRow, error)
	GetLatestCompatibleAuthority(context.Context, string, string) (store.CompatibleAuthorityRow, bool, error)
	GetWitnessAttemptBinding(context.Context, string) (store.WitnessAttemptBindingRow, bool, error)
	ListChallengesForInvariant(context.Context, string) ([]store.ChallengeRecord, error)
	FindInvariantRevisionForCandidate(context.Context, string) (string, error)
	PersistFrontierGeneration(context.Context, store.FrontierGenerationRecord) (store.PersistFrontierGenerationResult, error)
	GetFrontierGeneration(context.Context, string) (store.FrontierGenerationRecord, error)
	ListFrontierGenerations(context.Context, string) ([]store.FrontierGenerationRecord, error)
	LatestFrontierGeneration(context.Context, string) (string, bool, error)
	// FindProposalGeneration resolves the owning frontier_generation_run_id for a
	// (problem, proposal) — the generation that FIRST wrote the proposal, which
	// cross-run dedup preserves. Lets by-id evaluate reach a proposal even when
	// the LATEST generation deduped it and therefore owns zero proposal rows
	// (finding 2).
	FindProposalGeneration(context.Context, string, string) (string, bool, error)
	// LatestFrontierGenerationWithProposals resolves the most recent generation
	// that actually OWNS >=1 proposal row for the problem, so batch evaluate does
	// not go blind when a fully-deduped latest generation owns none (finding 2).
	LatestFrontierGenerationWithProposals(context.Context, string) (string, bool, error)
	// LatestProposalOccurrenceGeneration resolves the generation that most
	// recently EMITTED an interpretation of the proposal (occurrence binding) —
	// the default assessment context for by-id re-evaluation, so revised
	// content is reachable (87759d9 finding 2).
	LatestProposalOccurrenceGeneration(context.Context, string, string) (string, bool, error)
	// LatestFrontierGenerationWithOccurrences resolves the most recent
	// generation with >=1 occurrence binding — batch evaluation consumes
	// occurrence membership, not artifact ownership (87759d9 finding 2).
	LatestFrontierGenerationWithOccurrences(context.Context, string) (string, bool, error)
	// ListOccurrenceProposalRows returns the full proposal rows for a
	// generation's occurrence membership (what it emitted, owned or deduped).
	ListOccurrenceProposalRows(context.Context, string) ([]store.FrontierProposalRow, error)
	// RecordFailedProviderInvocation retains a REJECTED provider attempt's
	// payload envelope on the failed run (512bc54 f3).
	RecordFailedProviderInvocation(context.Context, store.FrontierProviderInvocation) error
	// ListProviderInvocationsForRun reads back a run's invocation envelopes,
	// including failed attempts.
	ListProviderInvocationsForRun(context.Context, string) ([]store.ProviderInvocationPayloadRow, error)
	RedundantAttackKeys(context.Context, string, int) ([]string, error)
	// Boundary deltas (v43, D5): the first persisted next-decision edge the
	// search-policy stage consumes from the challenge stage.
	ListBoundaryDeltasForProblem(context.Context, string) ([]store.ProblemBoundaryDeltaRow, error)
	ListBreakCohortRows(context.Context, string) ([]store.BreakCohortRow, error)
	// Normative review ledger (v44, G1 of the 2026-09-12 review-flow run).
	// These are NORMATIVE records — obligations, applicability, checks,
	// assessments — kept strictly apart from the scientific lifecycle above.
	// There is no coverage-status writer: coverage is generated from
	// LoadReviewCoverage.
	PersistReviewPolicy(context.Context, store.ReviewPolicyRecord) (store.ReviewPolicyRow, error)
	PersistReviewObligation(context.Context, store.ReviewObligationRow) (store.ReviewObligationRow, error)
	PersistReviewApplicabilityDecision(context.Context, store.ReviewApplicabilityDecisionRow) (store.ReviewApplicabilityDecisionRow, error)
	PersistReviewDependencyManifest(context.Context, store.ReviewDependencyManifestRow) (store.ReviewDependencyManifestRow, error)
	PersistReviewCheckAttempt(context.Context, store.ReviewCheckAttemptRow) (store.ReviewCheckAttemptRow, error)
	PersistReviewAssessment(context.Context, store.ReviewAssessmentRow) (store.ReviewAssessmentRow, error)
	LatestReviewPolicy(context.Context, string) (store.ReviewPolicyRow, bool, error)
	GetReviewPolicy(context.Context, string) (store.ReviewPolicyRow, error)
	LoadReviewCoverage(context.Context, string) (store.ReviewCoverage, error)
	PersistSuccessRevision(context.Context, store.SuccessRevisionRecord) (store.PersistSuccessRevisionResult, error)
	GetSuccessRevision(context.Context, string) (store.SuccessRevisionRecord, error)
	ListSuccessRevisions(context.Context, string) ([]store.SuccessRevisionRecord, error)
	LatestSuccessRevision(context.Context, string) (string, bool, error)
	// LatestSelectedSuccessRevision resolves the artifact chosen by the most
	// recent compression EXECUTION (v28) — current guidance follows the latest
	// selection, never MAX(revision), so a reused older (e.g. empty) artifact
	// displaces a higher-numbered stale one.
	LatestSelectedSuccessRevision(context.Context, string) (string, bool, error)
	// ListCompressionSelections returns the compression execution history
	// (v28, oldest first) — the audit trail behind current guidance.
	ListCompressionSelections(context.Context, string) ([]store.CompressionSelectionRow, error)
	PersistHoldoutSet(context.Context, store.HoldoutSetRecord) (store.HoldoutSetRecord, bool, error)
	GetHoldoutSet(context.Context, string) (store.HoldoutSetRecord, error)
	LatestHoldoutSet(context.Context, string) (string, bool, error)
	CountHoldoutSourceDating(context.Context, string) (int, int, error)
	// RecordHoldoutSourceDating / ListHoldoutSourceDating: the historical-mode
	// gate's write path — one externally auditable dated-evidence row per
	// withheld source, immutable, recorded via `experiment date-source`.
	RecordHoldoutSourceDating(context.Context, store.HoldoutSourceDatingRecord) (bool, error)
	ListHoldoutSourceDating(context.Context, string) ([]store.HoldoutSourceDatingRecord, error)
	RunLeakageCheck(context.Context, string, string, string, string) (store.LeakageCheckRecord, error)
	GetLeakageCheck(context.Context, string) (store.LeakageCheckRecord, error)
	ListTargetSignaturesForHoldout(context.Context, string) ([]store.TargetSignatureRow, error)
	// RecordExperimentExecutions / ListExperimentExecutions: v32 execution
	// attribution — which capture each execution assessed, even under
	// assessment-artifact reuse.
	RecordExperimentExecutions(context.Context, []store.ExperimentExecutionRow) error
	ListExperimentExecutions(context.Context, string) ([]store.ExperimentExecutionRow, error)
	PersistExperiment(context.Context, store.ExperimentRecord) (store.PersistExperimentResult, error)
	GetExperiment(context.Context, string) (store.ExperimentRecord, error)
	ListExperiments(context.Context, string) ([]store.ExperimentRecord, error)
	PersistPolicyRevision(context.Context, store.PolicyRevisionRecord) (store.PolicyRevisionRecord, bool, error)
	GetPolicyRevision(context.Context, string) (store.PolicyRevisionRecord, error)
	ListPolicyRevisions(context.Context, string) ([]store.PolicyRevisionRecord, error)
	LatestPolicyRevision(context.Context, string) (string, bool, error)
	// LatestSelectedPolicyRevision resolves the artifact chosen by the most
	// recent mutation EXECUTION (v30) — current policy follows the latest
	// selection, never MAX(revision) (P5 audit finding 1).
	LatestSelectedPolicyRevision(context.Context, string) (string, bool, error)
	PersistFrontierGenerationPolicy(context.Context, string, string, []store.FrontierGenerationPolicyRow) error
	GetFrontierGenerationPolicy(context.Context, string) (string, []store.FrontierGenerationPolicyRow, error)
	PersistEvaluationRun(context.Context, store.EvaluationRunRecord) (store.EvaluationRunRecord, error)
	GetEvaluationRun(context.Context, string) (store.EvaluationRunRecord, error)
	GetEvaluation(context.Context, string) (store.EvaluationRow, error)
	ListEvaluations(context.Context, string) ([]store.EvaluationRunRecord, error)
	ListEvaluatedFailures(context.Context, string) ([]store.EvaluatedFailureRow, error)
	// Evidence admission (v35/S2): the typed decision ledger bridging
	// evaluated failures into the atlas population, plus the content-addressed
	// read of the exact assessed proposal-signature bytes.
	PersistEvidenceAdmission(context.Context, store.EvidenceAdmissionRow) (store.EvidenceAdmissionRow, error)
	// PersistAdmittedFailure materializes one admitted evaluated failure
	// (snapshot + normalization + signature + admission decision) in a single
	// transaction — partial materializations must be impossible.
	PersistAdmittedFailure(context.Context, store.AdmittedFailureInput) (store.AdmittedFailureResult, error)
	ListEvidenceAdmissions(context.Context, string) ([]store.EvidenceAdmissionRow, error)
	// AdmittedSignatureKinds maps admitted signature ids to their ledger
	// observation kind so population views can label admitted members with
	// their epistemic provenance (2026-09-12 review F4).
	AdmittedSignatureKinds(context.Context, string) (map[string]string, error)
	GetProposalSignatureContentByHash(context.Context, string, string) (store.ProposalSignatureContentRow, bool, error)
	// Witness-backed evaluations (issue #23 slice 2, D2-C): a new assessment
	// pins the OCCURRENCE it is about (F2, 2026-09-12 review) — its generation,
	// that generation's population, and the exact content revision the
	// generation bound — via GetFrontierGeneration /
	// ListOccurrenceProposalRows / ListGenerationOccurrenceContents above.
	// Deliberately NOT a "latest revision" lookup: an assessment attached to
	// arbitrary newest bytes cannot be attributed to any occurrence.
	// Projection chain (v37/S5): concrete plan -> obligations -> decisions.
	GetProposalProblem(context.Context, string) (string, error)
	PersistProjection(context.Context, store.ProjectionRecord) (store.ProjectionRecord, error)
	PersistObligationDecision(context.Context, store.ProjectionObligationDecisionRow) error
	GetProjectionObligation(context.Context, string) (store.ProjectionObligationRow, store.ProjectionArtifactRow, error)
	ListProjectionsForProblem(context.Context, string) ([]store.ProjectionRecord, error)
	// Authored claim forms (v39/#21): quantifier + scope + role.
	PersistInvariantClaimForm(context.Context, store.InvariantClaimFormRow) error
	GetLatestInvariantClaimForm(context.Context, string) (store.InvariantClaimFormRow, bool, error)
	// Prospective two-observation episodes (v41/#23 consumer).
	PersistEpisode(context.Context, store.EpisodeRow, store.EpisodeCommitmentRow) error
	PersistEpisodeStepTwo(context.Context, store.EpisodeCommitmentRow) error
	PersistEpisodeObservation(context.Context, store.EpisodeObservationRow, string) error
	PersistEpisodeRevision(context.Context, store.EpisodeRevisionRow) error
	GetEpisode(context.Context, string) (store.EpisodeRecord, error)
	ListEpisodesForProblem(context.Context, string) ([]store.EpisodeRow, error)
}

type InitProblemInput struct {
	DBPath     string
	Statement  string
	Slug       string
	ForceNew   bool
	JSONOutput bool
}

type LookupInput struct {
	DBPath     string
	ID         string
	JSONOutput bool
}

type ListInput struct {
	DBPath     string
	JSONOutput bool
}

func New(version string) *App {
	app := &App{
		version: version,
		now: func() time.Time {
			return time.Now().UTC()
		},
		getwd: os.Getwd,
		stdin: os.Stdin,
		normalizers: map[string]provider.Normalizer{
			provider.FixtureProviderName: provider.NewFixtureNormalizer(),
		},
	}
	app.openStoreFn = app.defaultOpenStore
	app.projectRevisionFn = resolveProjectRevision
	return app
}

// projectRevisionUnknown is recorded when the checkout revision cannot be
// resolved. It is deliberately an explicit marker rather than an empty string:
// "we could not tell which code was assessed" is a real limitation that must
// survive into the export, not a field a reader can mistake for a clean answer.
const projectRevisionUnknown = "unknown"

// dirtyRevisionDigestLength is how much of the working-tree digest is kept. 48
// bits is far beyond what is needed to distinguish successive edits of one tree,
// and a full hash would bury the commit id it annotates.
const dirtyRevisionDigestLength = 12

// resolveProjectRevision reports the current checkout revision, content-addressing
// any uncommitted state.
//
// A hardcoded revision silently ages: an assessment recorded against a literal
// stays "current" across every later commit, and the stale literal is exactly
// the value a later reader would use to judge compatibility. Resolving at write
// time is what lets `project_revision` function as a real dependency.
//
// A bare `+dirty` marker is not enough, and that gap was a finding of its own
// (E3-2, docs/reviews/2026-09-12-e3-w2-migration-run.md). Two materially
// different trees at one commit stringify identically under a bare marker, so an
// assessment of pre-fix code and an assessment of post-fix code declare the same
// dependency value, neither can go stale relative to the other, and a
// demonstrated blocker can never be retired. In a repository whose norm is a
// dirty tree, that covers most changes. The digest closes it: the revision
// changes when the bytes change, with no commit in between.
//
// What the digest covers, exactly:
//
//   - tracked changes, staged and unstaged, via `git diff HEAD --binary`.
//     `--binary` is belt-and-braces rather than strictly required: a plain diff
//     renders a modified binary as "Binary files … differ", but its `index`
//     line still carries abbreviated before/after blob hashes, so content
//     changes do reach the digest either way. This was checked by mutation —
//     dropping `--binary` did not break the binary test — so the flag is kept
//     for the stronger reason that an abbreviated hash is a 7-hex-digit prefix
//     chosen for human display, while `--binary` embeds the actual content.
//   - untracked, non-ignored files, by path AND content.
//
// What it deliberately excludes: ignored files. In this repository
// `docs/reviews` is ignored, so writing a review does not change the revision —
// which is intended, since a review should not invalidate its own assessment.
func resolveProjectRevision() string {
	return resolveProjectRevisionIn("")
}

// resolveProjectRevisionIn resolves the revision of the checkout containing dir.
// An empty dir means the current working directory. The parameter exists so the
// resolver is testable against a fixture repository without a process-wide
// chdir.
func resolveProjectRevisionIn(dir string) string {
	head, err := gitOutput(dir, "rev-parse", "HEAD")
	if err != nil {
		return projectRevisionUnknown
	}
	revision := strings.TrimSpace(string(head))
	if revision == "" {
		return projectRevisionUnknown
	}

	status, err := gitOutput(dir, "status", "--porcelain")
	if err != nil {
		// The commit is known but cleanliness is not. Saying so beats implying
		// the tree was clean.
		return revision + "+dirty:" + projectRevisionUnknown
	}
	if len(strings.TrimSpace(string(status))) == 0 {
		return revision
	}

	digest, ok := workingTreeDigest(dir)
	if !ok {
		return revision + "+dirty:" + projectRevisionUnknown
	}
	return revision + "+dirty:" + digest
}

// workingTreeDigest hashes the uncommitted content of the checkout containing
// dir. ok=false means the digest could not be computed, which the caller must
// report rather than paper over.
func workingTreeDigest(dir string) (string, bool) {
	sum := sha256.New()

	trackedDiff, err := gitOutput(dir, "diff", "HEAD", "--binary")
	if err != nil {
		return "", false
	}
	sum.Write(trackedDiff)

	// `ls-files --others --exclude-standard -z` lists untracked, non-ignored
	// files one per NUL, unquoted. It is used in preference to parsing porcelain
	// status because that output collapses untracked directories and quotes
	// unusual paths, both of which would make the digest depend on formatting
	// rather than on content.
	untracked, err := gitOutput(dir, "ls-files", "--others", "--exclude-standard", "-z")
	if err != nil {
		return "", false
	}
	paths := make([]string, 0, 8)
	for _, p := range strings.Split(string(untracked), "\x00") {
		if p != "" {
			paths = append(paths, p)
		}
	}
	// Sorted so the digest depends on the set of files, not on git's ordering.
	sort.Strings(paths)

	root := dir
	if root == "" {
		root = "."
	}
	for _, p := range paths {
		sum.Write([]byte("\x00untracked:" + p + "\x00"))
		content, readErr := os.ReadFile(filepath.Join(root, p))
		if readErr != nil {
			// A file that vanished mid-walk, or is unreadable, makes the digest
			// incomplete. Refuse rather than emit a value that looks precise.
			return "", false
		}
		sum.Write(content)
	}

	return hex.EncodeToString(sum.Sum(nil))[:dirtyRevisionDigestLength], true
}

// gitOutput runs one git command in dir, returning its stdout.
func gitOutput(dir string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	return cmd.Output()
}

// projectRevision returns the resolver's answer, tolerating a nil hook so a
// zero-value App in a test cannot panic here.
func (a *App) projectRevision() string {
	if a.projectRevisionFn == nil {
		return projectRevisionUnknown
	}
	return a.projectRevisionFn()
}

func (a *App) InitProblem(ctx context.Context, input InitProblemInput) (InitResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return InitResponse{}, err
	}
	defer repoStore.Close()

	slug, err := domain.CanonicalSlug(input.Slug, input.Statement)
	if err != nil {
		return InitResponse{}, err
	}

	if !input.ForceNew {
		if existing, found, findErr := repoStore.FindProblemBySlug(ctx, slug); findErr != nil {
			return InitResponse{}, findErr
		} else if found {
			runTime := a.now()
			run := domain.NewRun{
				ID:          domain.NewRunID(runTime),
				ProblemID:   existing.ID,
				Operation:   "init",
				Status:      domain.RunStatusInitialized,
				InputRef:    "problem_slug:" + slug,
				ToolName:    "newf",
				ToolVersion: a.version,
				StartedAt:   runTime,
				CompletedAt: runTime,
			}
			persistedRun, createErr := repoStore.CreateRun(ctx, run)
			if createErr != nil {
				return InitResponse{}, createErr
			}
			return InitResponse{
				OK:        true,
				Command:   "init",
				ProblemID: existing.ID,
				RunID:     persistedRun.ID,
				Store:     dbPath,
				Created:   false,
				Problem:   existing.Statement,
			}, nil
		}
	}

	if input.ForceNew {
		nextSlug, slugErr := repoStore.NextProblemSlug(ctx, slug)
		if slugErr != nil {
			return InitResponse{}, slugErr
		}
		slug = nextSlug
	}

	now := a.now()
	runID := domain.NewRunID(now)
	problem := domain.NewProblem{
		ID:             domain.NewProblemID(now),
		Slug:           slug,
		Statement:      input.Statement,
		Status:         domain.ProblemStatusActive,
		CreatedAt:      now,
		CreatedByRunID: runID,
	}
	run := domain.NewRun{
		ID:          runID,
		ProblemID:   problem.ID,
		Operation:   "init",
		Status:      domain.RunStatusInitialized,
		InputRef:    "problem_slug:" + slug,
		ToolName:    "newf",
		ToolVersion: a.version,
		StartedAt:   now,
		CompletedAt: now,
	}

	persistedProblem, persistedRun, err := repoStore.CreateProblemWithRun(ctx, problem, run)
	if err != nil {
		if !input.ForceNew && errors.Is(err, store.ErrDuplicateSlug) {
			existing, found, findErr := repoStore.FindProblemBySlug(ctx, slug)
			if findErr != nil {
				return InitResponse{}, findErr
			}
			if found {
				fallbackNow := a.now()
				fallbackRun, createErr := repoStore.CreateRun(ctx, domain.NewRun{
					ID:          domain.NewRunID(fallbackNow),
					ProblemID:   existing.ID,
					Operation:   "init",
					Status:      domain.RunStatusInitialized,
					InputRef:    "problem_slug:" + slug,
					ToolName:    "newf",
					ToolVersion: a.version,
					StartedAt:   fallbackNow,
					CompletedAt: fallbackNow,
				})
				if createErr != nil {
					return InitResponse{}, createErr
				}
				return InitResponse{
					OK:        true,
					Command:   "init",
					ProblemID: existing.ID,
					RunID:     fallbackRun.ID,
					Store:     dbPath,
					Created:   false,
					Problem:   existing.Statement,
				}, nil
			}
		}
		return InitResponse{}, err
	}

	return InitResponse{
		OK:        true,
		Command:   "init",
		ProblemID: persistedProblem.ID,
		RunID:     persistedRun.ID,
		Store:     dbPath,
		Created:   true,
		Problem:   persistedProblem.Statement,
	}, nil
}

func (a *App) ShowProblem(ctx context.Context, input LookupInput) (ProblemShowResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return ProblemShowResponse{}, err
	}
	defer repoStore.Close()

	problem, err := repoStore.GetProblem(ctx, input.ID)
	if err != nil {
		return ProblemShowResponse{}, err
	}

	return ProblemShowResponse{
		OK:      true,
		Command: "problem show",
		Store:   dbPath,
		Problem: problemView(problem),
	}, nil
}

func (a *App) ListProblems(ctx context.Context, input ListInput) (ProblemListResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return ProblemListResponse{}, err
	}
	defer repoStore.Close()

	problems, err := repoStore.ListProblems(ctx)
	if err != nil {
		return ProblemListResponse{}, err
	}

	views := make([]ProblemView, 0, len(problems))
	for _, problem := range problems {
		views = append(views, problemView(problem))
	}

	return ProblemListResponse{
		OK:       true,
		Command:  "problem list",
		Store:    dbPath,
		Problems: views,
	}, nil
}

func (a *App) ShowRun(ctx context.Context, input LookupInput) (RunShowResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return RunShowResponse{}, err
	}
	defer repoStore.Close()

	run, err := repoStore.GetRun(ctx, input.ID)
	if err != nil {
		return RunShowResponse{}, err
	}

	return RunShowResponse{
		OK:      true,
		Command: "run show",
		Store:   dbPath,
		Run:     runView(run),
	}, nil
}

func (a *App) defaultOpenStore(ctx context.Context, dbPath string) (string, problemStore, error) {
	cwd, err := a.getwd()
	if err != nil {
		return "", nil, err
	}

	resolvedPath, err := config.ResolveDBPath(cwd, dbPath, nil)
	if err != nil {
		return "", nil, err
	}

	repoStore, err := store.Open(resolvedPath)
	if err != nil {
		return "", nil, err
	}

	if err := repoStore.Migrate(ctx); err != nil {
		repoStore.Close()
		return "", nil, err
	}

	if err := a.seedVocabularies(ctx, repoStore); err != nil {
		repoStore.Close()
		return "", nil, err
	}

	return resolvedPath, repoStore, nil
}

func problemView(problem domain.Problem) ProblemView {
	return ProblemView{
		ID:             problem.ID,
		Slug:           problem.Slug,
		Statement:      problem.Statement,
		Status:         string(problem.Status),
		CreatedAt:      problem.CreatedAt.Format(time.RFC3339Nano),
		CreatedByRunID: problem.CreatedByRunID,
	}
}

func runView(run domain.Run) RunView {
	return RunView{
		ID:           run.ID,
		ProblemID:    run.ProblemID,
		ParentRunID:  run.ParentRunID,
		Operation:    run.Operation,
		Status:       string(run.Status),
		InputRef:     run.InputRef,
		ToolName:     run.ToolName,
		ToolVersion:  run.ToolVersion,
		StartedAt:    run.StartedAt.Format(time.RFC3339Nano),
		CompletedAt:  run.CompletedAt.Format(time.RFC3339Nano),
		ErrorSummary: run.ErrorSummary,
	}
}

// finalizeRun transitions a just-executed run to its terminal lifecycle state so
// `run show` reflects the real command outcome instead of the status chosen
// before work happened. An empty failures slice yields `completed`; otherwise
// the run is marked `failed` with a deterministic, order-independent
// error_summary aggregating the failing items. The failure detail lives at the
// per-item result level; the run-level summary is a stable count + joined
// reasons so automation treating `run show` as source-of-truth is not misled.
func (a *App) finalizeRun(ctx context.Context, repoStore problemStore, runID string, failures []string) error {
	if len(failures) == 0 {
		return repoStore.UpdateRunStatus(ctx, runID, domain.RunStatusCompleted, a.now(), nil)
	}
	sorted := make([]string, len(failures))
	copy(sorted, failures)
	sort.Strings(sorted)
	summary := fmt.Sprintf("%d item(s) failed: %s", len(sorted), strings.Join(sorted, "; "))
	return repoStore.UpdateRunStatus(ctx, runID, domain.RunStatusFailed, a.now(), &summary)
}

// failRun marks a single-outcome run as failed with the error as its summary. It
// is best-effort on an already-failing path: the caller returns the original
// error regardless, so a telemetry-write failure here must not mask the real
// cause. Used by commands that either fully succeed or return early with an
// error, where a per-item failures slice does not apply.
func (a *App) failRun(ctx context.Context, repoStore problemStore, runID string, cause error) {
	summary := cause.Error()
	_ = repoStore.UpdateRunStatus(ctx, runID, domain.RunStatusFailed, a.now(), &summary)
}
