package pipeline

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/instagrim-dev/newf/internal/config"
	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/provider"
	"github.com/instagrim-dev/newf/internal/store"
)

type App struct {
	version          string
	now              func() time.Time
	getwd            func() (string, error)
	stdin            io.Reader
	openStoreFn      func(context.Context, string) (string, problemStore, error)
	normalizers      map[string]provider.Normalizer
	invariantMinerFn provider.InvariantMiner
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
	GetSource(context.Context, string) (domain.Source, error)
	GetSourceSnapshot(context.Context, string) (domain.SourceSnapshot, error)
	ListSourceSnapshots(context.Context, string) ([]domain.SourceSnapshot, error)
	FindEquivalentNormalization(context.Context, string, string, string) (store.ExistingNormalization, error)
	LatestNormalizationForSnapshot(context.Context, string) (domain.NormalizationRevision, bool, error)
	PersistNormalization(context.Context, store.NormalizationInput) (store.NormalizationWriteResult, error)
	ListApproaches(context.Context, string) ([]store.ApproachListItem, error)
	GetApproachDetail(context.Context, string) (store.ApproachDetail, error)
	GetMechanismDetail(context.Context, string) (store.ApproachDetail, error)
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
	PersistClusterRun(context.Context, store.ClusterRunRecord) (store.PersistClusterRunResult, error)
	GetClusterRun(context.Context, string) (store.ClusterRunRecord, error)
	ListClusterRuns(context.Context, string) ([]store.ClusterRunRecord, error)
	LatestClusterRun(context.Context, string) (string, bool, error)
	PersistFailureSpace(context.Context, store.FailureSpaceRecord) (store.PersistFailureSpaceResult, error)
	GetFailureSpace(context.Context, string) (store.FailureSpaceRecord, error)
	LatestFailureSpace(context.Context, string) (store.FailureSpaceRecord, bool, error)
	PersistInvariantRevision(context.Context, store.InvariantRevisionRecord) (store.PersistInvariantRevisionResult, error)
	LookupInvariantRevision(context.Context, store.InvariantReuseKey) (store.InvariantRevisionRecord, bool, error)
	GetInvariantRevision(context.Context, string) (store.InvariantRevisionRecord, error)
	ListInvariantRevisions(context.Context, string) ([]store.InvariantRevisionRecord, error)
	LatestInvariantRevision(context.Context, string) (string, bool, error)
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
	return app
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
