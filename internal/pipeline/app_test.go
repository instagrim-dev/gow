package pipeline

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/store"
)

func TestInitProblemFallsBackToExistingProblemOnDuplicateSlug(t *testing.T) {
	t.Parallel()

	firstNow := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	secondNow := firstNow.Add(time.Second)
	existingProblem := domain.Problem{
		ID:             domain.NewProblemID(firstNow),
		Slug:           "erdos-straus-conjecture",
		Statement:      "Erdos-Straus conjecture",
		Status:         domain.ProblemStatusActive,
		CreatedAt:      firstNow,
		CreatedByRunID: domain.NewRunID(firstNow),
	}
	fake := &fakeProblemStore{
		existingProblem: existingProblem,
	}

	nowCalls := 0
	app := &App{
		version: "dev",
		now: func() time.Time {
			nowCalls++
			if nowCalls == 1 {
				return firstNow
			}
			return secondNow
		},
		getwd: func() (string, error) {
			return "/workspace/repo", nil
		},
		openStoreFn: func(context.Context, string) (string, problemStore, error) {
			return "/workspace/repo/.newf/newf.db", fake, nil
		},
	}

	result, err := app.InitProblem(context.Background(), InitProblemInput{
		DBPath:    "/workspace/repo/.newf/newf.db",
		Statement: "Erdos-Straus conjecture",
	})
	if err != nil {
		t.Fatalf("InitProblem() error = %v", err)
	}

	if result.Created {
		t.Fatal("InitProblem() created = true, want false")
	}
	if result.ProblemID != existingProblem.ID {
		t.Fatalf("InitProblem() problem_id = %q, want %q", result.ProblemID, existingProblem.ID)
	}
	if fake.createdRun == nil {
		t.Fatal("fallback CreateRun() was not called")
	}
	if !fake.createdRun.StartedAt.Equal(secondNow) {
		t.Fatalf("fallback run StartedAt = %s, want %s", fake.createdRun.StartedAt, secondNow)
	}
}

func TestInitProblemForceNewUsesNextAvailableSlug(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	fake := &fakeProblemStore{
		nextSlug: "erdos-straus-conjecture-2",
	}

	app := &App{
		version: "dev",
		now: func() time.Time {
			return now
		},
		getwd: func() (string, error) {
			return "/workspace/repo", nil
		},
		openStoreFn: func(context.Context, string) (string, problemStore, error) {
			return "/workspace/repo/.newf/newf.db", fake, nil
		},
	}

	result, err := app.InitProblem(context.Background(), InitProblemInput{
		DBPath:    "/workspace/repo/.newf/newf.db",
		Statement: "Erdos-Straus conjecture",
		ForceNew:  true,
	})
	if err != nil {
		t.Fatalf("InitProblem() error = %v", err)
	}

	if !result.Created {
		t.Fatal("InitProblem() created = false, want true")
	}
	if fake.createdProblem == nil || fake.createdProblem.Slug != "erdos-straus-conjecture-2" {
		t.Fatalf("created problem slug = %#v, want erdos-straus-conjecture-2", fake.createdProblem)
	}
	if fake.createdCreateRun == nil || fake.createdCreateRun.InputRef != "problem_slug:erdos-straus-conjecture-2" {
		t.Fatalf("created run input_ref = %#v", fake.createdCreateRun)
	}
}

func TestInitProblemForceNewWithExplicitSlugUsesNextAvailableSlug(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	fake := &fakeProblemStore{
		nextSlug: "custom-slug-2",
	}

	app := &App{
		version: "dev",
		now: func() time.Time {
			return now
		},
		getwd: func() (string, error) {
			return "/workspace/repo", nil
		},
		openStoreFn: func(context.Context, string) (string, problemStore, error) {
			return "/workspace/repo/.newf/newf.db", fake, nil
		},
	}

	result, err := app.InitProblem(context.Background(), InitProblemInput{
		DBPath:    "/workspace/repo/.newf/newf.db",
		Statement: "Erdos-Straus conjecture",
		Slug:      "custom-slug",
		ForceNew:  true,
	})
	if err != nil {
		t.Fatalf("InitProblem() error = %v", err)
	}

	if !result.Created {
		t.Fatal("InitProblem() created = false, want true")
	}
	if fake.createdProblem == nil || fake.createdProblem.Slug != "custom-slug-2" {
		t.Fatalf("created problem slug = %#v, want custom-slug-2", fake.createdProblem)
	}
}

type fakeProblemStore struct {
	existingProblem         domain.Problem
	findCalls               int
	nextSlug                string
	returnDuplicateOnCreate bool
	createdProblem          *domain.NewProblem
	createdCreateRun        *domain.NewRun
	createdRun              *domain.NewRun
}

func (f *fakeProblemStore) Close() error { return nil }

func (f *fakeProblemStore) FindProblemBySlug(context.Context, string) (domain.Problem, bool, error) {
	f.findCalls++
	if f.findCalls == 1 {
		return domain.Problem{}, false, nil
	}
	return f.existingProblem, true, nil
}

func (f *fakeProblemStore) NextProblemSlug(context.Context, string) (string, error) {
	if f.nextSlug == "" {
		return "", errors.New("unexpected call")
	}
	return f.nextSlug, nil
}

func (f *fakeProblemStore) CreateProblemWithRun(_ context.Context, problem domain.NewProblem, run domain.NewRun) (domain.Problem, domain.Run, error) {
	if f.returnDuplicateOnCreate || f.existingProblem.ID != "" {
		return domain.Problem{}, domain.Run{}, store.ErrDuplicateSlug
	}
	f.createdProblem = &problem
	f.createdCreateRun = &run
	return domain.Problem{
			ID:             problem.ID,
			Slug:           problem.Slug,
			Statement:      problem.Statement,
			Status:         problem.Status,
			CreatedAt:      problem.CreatedAt,
			CreatedByRunID: problem.CreatedByRunID,
		}, domain.Run{
			ID:          run.ID,
			ProblemID:   run.ProblemID,
			Operation:   run.Operation,
			Status:      run.Status,
			InputRef:    run.InputRef,
			ToolName:    run.ToolName,
			ToolVersion: run.ToolVersion,
			StartedAt:   run.StartedAt,
			CompletedAt: run.CompletedAt,
		}, nil
}

func (f *fakeProblemStore) CreateRun(_ context.Context, run domain.NewRun) (domain.Run, error) {
	f.createdRun = &run
	return domain.Run{
		ID:          run.ID,
		ProblemID:   run.ProblemID,
		Operation:   run.Operation,
		Status:      run.Status,
		InputRef:    run.InputRef,
		ToolName:    run.ToolName,
		ToolVersion: run.ToolVersion,
		StartedAt:   run.StartedAt,
		CompletedAt: run.CompletedAt,
	}, nil
}

func (f *fakeProblemStore) GetProblem(context.Context, string) (domain.Problem, error) {
	return domain.Problem{}, errors.New("unexpected call")
}

func (f *fakeProblemStore) ListProblems(context.Context) ([]domain.Problem, error) {
	return nil, errors.New("unexpected call")
}

func (f *fakeProblemStore) GetRun(context.Context, string) (domain.Run, error) {
	return domain.Run{}, errors.New("unexpected call")
}

func (f *fakeProblemStore) CreateSourceSnapshot(context.Context, store.SnapshotAdmission) (store.SnapshotAdmissionResult, error) {
	return store.SnapshotAdmissionResult{}, errors.New("unexpected call")
}

func (f *fakeProblemStore) ListSourcesByProblem(context.Context, string) ([]domain.Source, error) {
	return nil, errors.New("unexpected call")
}

func (f *fakeProblemStore) ListSourcesWithSnapshotStats(context.Context, string) ([]store.SourceSummary, error) {
	return nil, errors.New("unexpected call")
}

func (f *fakeProblemStore) GetSource(context.Context, string) (domain.Source, error) {
	return domain.Source{}, errors.New("unexpected call")
}

func (f *fakeProblemStore) GetSourceSnapshot(context.Context, string) (domain.SourceSnapshot, error) {
	return domain.SourceSnapshot{}, errors.New("unexpected call")
}

func (f *fakeProblemStore) ListSourceSnapshots(context.Context, string) ([]domain.SourceSnapshot, error) {
	return nil, errors.New("unexpected call")
}

func (f *fakeProblemStore) FindEquivalentNormalization(context.Context, string, string, string) (store.ExistingNormalization, error) {
	return store.ExistingNormalization{}, errors.New("unexpected call")
}

func (f *fakeProblemStore) LatestNormalizationForSnapshot(context.Context, string) (domain.NormalizationRevision, bool, error) {
	return domain.NormalizationRevision{}, false, errors.New("unexpected call")
}

func (f *fakeProblemStore) PersistNormalization(context.Context, store.NormalizationInput) (store.NormalizationWriteResult, error) {
	return store.NormalizationWriteResult{}, errors.New("unexpected call")
}

func (f *fakeProblemStore) ListApproaches(context.Context, string) ([]store.ApproachListItem, error) {
	return nil, errors.New("unexpected call")
}

func (f *fakeProblemStore) GetApproachDetail(context.Context, string) (store.ApproachDetail, error) {
	return store.ApproachDetail{}, errors.New("unexpected call")
}

func (f *fakeProblemStore) GetMechanismDetail(context.Context, string) (store.ApproachDetail, error) {
	return store.ApproachDetail{}, errors.New("unexpected call")
}

func (f *fakeProblemStore) ListApproachRevisions(context.Context, string) (domain.Approach, []domain.ApproachRevision, error) {
	return domain.Approach{}, nil, errors.New("unexpected call")
}

func (f *fakeProblemStore) SeedVocabulary(context.Context, store.VocabularySeedInput) error {
	return nil
}

func (f *fakeProblemStore) ListVocabularies(context.Context) ([]store.VocabularyRecord, error) {
	return nil, errors.New("unexpected call")
}

func (f *fakeProblemStore) ListTerms(context.Context, string, string) ([]store.TermRecord, error) {
	return nil, errors.New("unexpected call")
}

func (f *fakeProblemStore) GetTerm(context.Context, string, string) (store.TermRecord, error) {
	return store.TermRecord{}, errors.New("unexpected call")
}

func (f *fakeProblemStore) PersistSignature(context.Context, store.SignatureRecord) (store.PersistSignatureResult, error) {
	return store.PersistSignatureResult{}, errors.New("unexpected call")
}

func (f *fakeProblemStore) GetSignature(context.Context, string) (store.SignatureRecord, error) {
	return store.SignatureRecord{}, errors.New("unexpected call")
}

func (f *fakeProblemStore) PersistComparison(context.Context, store.ComparisonRecord) error {
	return errors.New("unexpected call")
}
