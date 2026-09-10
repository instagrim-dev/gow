package pipeline

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/instagrim-dev/newf/internal/config"
	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/store"
)

type App struct {
	version string
	now     func() time.Time
	getwd   func() (string, error)
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
	return &App{
		version: version,
		now: func() time.Time {
			return time.Now().UTC()
		},
		getwd: os.Getwd,
	}
}

func (a *App) InitProblem(ctx context.Context, input InitProblemInput) (InitResponse, error) {
	dbPath, repoStore, err := a.openStore(ctx, input.DBPath)
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
				Status:      domain.RunStatusSucceeded,
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
		if input.Slug != "" && nextSlug != slug {
			return InitResponse{}, fmt.Errorf("%w: slug %q already exists", domain.ErrInvalidSlug, slug)
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
		Status:      domain.RunStatusSucceeded,
		InputRef:    "problem_slug:" + slug,
		ToolName:    "newf",
		ToolVersion: a.version,
		StartedAt:   now,
		CompletedAt: now,
	}

	persistedProblem, persistedRun, err := repoStore.CreateProblemWithRun(ctx, problem, run)
	if err != nil {
		if !input.ForceNew && store.IsUniqueSlugError(err) {
			existing, found, findErr := repoStore.FindProblemBySlug(ctx, slug)
			if findErr != nil {
				return InitResponse{}, findErr
			}
			if found {
				fallbackRun, createErr := repoStore.CreateRun(ctx, domain.NewRun{
					ID:          domain.NewRunID(now),
					ProblemID:   existing.ID,
					Operation:   "init",
					Status:      domain.RunStatusSucceeded,
					InputRef:    "problem_slug:" + slug,
					ToolName:    "newf",
					ToolVersion: a.version,
					StartedAt:   now,
					CompletedAt: now,
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
	dbPath, repoStore, err := a.openStore(ctx, input.DBPath)
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
	dbPath, repoStore, err := a.openStore(ctx, input.DBPath)
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
	dbPath, repoStore, err := a.openStore(ctx, input.DBPath)
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

func (a *App) openStore(ctx context.Context, dbPath string) (string, *store.Store, error) {
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
