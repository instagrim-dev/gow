package pipeline

import (
	"context"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/canon"
)

func TestIntegrationReemittedContentRequiresOccurrenceAssessment(t *testing.T) {
	ctx := context.Background()
	app, dbPath := newRealStoreApp(t, time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC))
	app.invariantMinerFn = minPreservesMiner{}
	app.challengerFn = biasOnlyChallenger{}
	app.generatorFn = pcGenerator{signatures: []canon.MechanismSignature{pcSignature(pcResidueLocality, "", true)}}
	problemID, invID, _ := mineOneCandidate(t, ctx, app, dbPath)
	if _, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID}); err != nil {
		t.Fatal(err)
	}
	first, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatal(err)
	}
	if len(first.Generation.Proposals) != 1 {
		t.Fatalf("want one proposal: %+v", first.Generation)
	}
	if result, err := app.Evaluate(ctx, EvaluateInput{DBPath: dbPath, ProblemID: problemID}); err != nil || result.Run.EvaluationCount != 1 {
		t.Fatalf("first assessment: %+v / %v", result, err)
	}
	second, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatal(err)
	}
	if len(second.Generation.Proposals) != 0 {
		t.Fatal("expected artifact deduplication")
	}
	result, err := app.Evaluate(ctx, EvaluateInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil || result.Run.EvaluationCount != 1 || result.Run.FrontierGenerationRunID != second.Generation.ID {
		t.Fatalf("content-only cache suppressed new-context assessment: %+v / %v", result, err)
	}
	// Repeating the same occurrence is still idempotent in batch mode.
	result, err = app.Evaluate(ctx, EvaluateInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil || result.Run.EvaluationCount != 0 {
		t.Fatalf("same occurrence was evaluated twice: %+v / %v", result, err)
	}
}
