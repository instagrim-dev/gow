package pipeline

import (
	"context"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/invariant"
	"github.com/instagrim-dev/newf/internal/provider"
)

// Guard-mutation block (follow-on P1; newf-regress `mutants` contract): each
// test DISABLES one guard in isolation — expressed as the exact pre-fix
// behavior the guard replaced — and requires that the NAMED semantic assertion
// fails against the mutant's result for the semantic reason. A compiler error,
// crash, or unrelated failure does not count as catching the bug: every mutant
// here produces a well-formed result that is WRONG in the named way, and the
// production guard is asserted alongside so the pair proves detection.

// mutant use_origin_occurrence: disable explicit occurrence selection by
// resolving the proposal's OWNING generation (the pre-v26 by-id selector)
// instead of its latest occurrence. must_fail = [B_OCCURRENCE, B_BYTES].
func TestGuardMutationUseOriginOccurrence(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = minPreservesMiner{}
	app.challengerFn = biasOnlyChallenger{}
	app.generatorFn = pcGenerator{signatures: []canon.MechanismSignature{pcSignature(pcResidueLocality, "", true)}}

	problemID, invID, _ := mineOneCandidate(t, ctx, app, dbPath)
	if resp, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID}); err != nil || resp.Reports[0].StateAfter != "surviving" {
		t.Fatalf("challenge: %v / %+v", err, resp.Reports)
	}
	gen1, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("generate A: %v", err)
	}
	proposalID := gen1.Generation.Proposals[0].ID
	app.generatorFn = pcGenerator{signatures: []canon.MechanismSignature{pcSignature(pcResidueLocality, "", false)}}
	gen2, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("generate B: %v", err)
	}

	repo := openTestStore(t, ctx, dbPath)
	defer repo.Close()
	occ1, _ := repo.ListGenerationOccurrenceContents(ctx, gen1.Generation.ID)
	occ2, _ := repo.ListGenerationOccurrenceContents(ctx, gen2.Generation.ID)
	hashA, hashB := occ1[proposalID].ContentHash, occ2[proposalID].ContentHash

	// GUARD ON: default by-id selection reaches B (B_OCCURRENCE, B_BYTES hold).
	prod, err := app.Evaluate(ctx, EvaluateInput{DBPath: dbPath, ProblemID: problemID, ProposalID: proposalID})
	if err != nil {
		t.Fatalf("production evaluate: %v", err)
	}
	prodRun, _ := repo.GetEvaluationRun(ctx, prod.Run.ID)
	if prodRun.Evaluations[0].SignatureContentHash != hashB {
		t.Fatalf("guard on: B_BYTES must hold, got %s", prodRun.Evaluations[0].SignatureContentHash)
	}

	// MUTANT: the disabled guard resolved the OWNER generation. Express it
	// exactly: the pre-fix selector still exists as FindProposalGeneration.
	ownerGen, found, err := repo.FindProposalGeneration(ctx, problemID, proposalID)
	if err != nil || !found {
		t.Fatalf("owner selector: %v %v", err, found)
	}
	mut, err := app.Evaluate(ctx, EvaluateInput{DBPath: dbPath, ProblemID: problemID, ProposalID: proposalID, GenerationID: ownerGen})
	if err != nil {
		t.Fatalf("mutant evaluate: %v", err)
	}
	mutRun, _ := repo.GetEvaluationRun(ctx, mut.Run.ID)

	// must_fail B_OCCURRENCE: the mutant's assessment context is NOT B's
	// occurrence generation.
	if mut.Run.FrontierGenerationRunID == gen2.Generation.ID {
		t.Fatal("mutant did not disable the guard: B_OCCURRENCE unexpectedly holds")
	}
	// must_fail B_BYTES: the mutant assessed A's bytes, not B's — the named
	// assertion fails for the semantic reason (stale occurrence selected).
	if got := mutRun.Evaluations[0].SignatureContentHash; got != hashA || got == hashB {
		t.Fatalf("mutant must reproduce the pre-fix defect (assess A): got %s (A=%s B=%s)", got, hashA, hashB)
	}
}

// mutant trust_nonempty_basis: disable the completeness authority check by
// overlaying EVERY persisted declaration regardless of admission — the
// pre-v26 mapping. must_fail = [DECLARATION_NOT_AUTHORITY]: absence on the
// declared field becomes a verified violation from an unaccepted provider
// claim, exactly the laundering the guard prevents.
func TestGuardMutationTrustNonemptyBasis(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	payload := controlTrainResult(true)
	for i := range payload.Approaches {
		payload.Approaches[i].Mechanism.CompletenessBasis = "The model believes extraction is exhaustive."
	}
	app.normalizers = map[string]provider.Normalizer{
		provider.FixtureProviderName: provider.NewFixtureNormalizer(),
		"model-sim":                  cannedNormalizer{result: payload},
	}
	doc := writeControlDoc(t, t.TempDir(), "mutant-basis.md", payload)
	init, err := app.InitProblem(ctx, InitProblemInput{DBPath: dbPath, Statement: "guard mutant (basis)", Slug: "mutant-basis"})
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	if _, err := app.IngestSources(ctx, IngestInput{DBPath: dbPath, ProblemID: init.ProblemID, Paths: []string{doc}}); err != nil {
		t.Fatalf("ingest: %v", err)
	}
	if _, err := app.Normalize(ctx, NormalizeInput{DBPath: dbPath, ProblemID: init.ProblemID, All: true, Provider: "model-sim"}); err != nil {
		t.Fatalf("normalize: %v", err)
	}

	repo := openTestStore(t, ctx, dbPath)
	defer repo.Close()
	approaches, err := repo.ListApproaches(ctx, init.ProblemID)
	if err != nil || len(approaches) == 0 {
		t.Fatalf("approaches: %v %d", err, len(approaches))
	}
	// Pick the residue-preserving failure approach: its payload declares
	// preserves complete with a confident but unaccepted basis.
	for _, item := range approaches {
		d, err := repo.GetApproachDetail(ctx, item.Approach.ID)
		if err != nil {
			t.Fatalf("detail: %v", err)
		}
		if len(d.FieldCompleteness) == 0 {
			continue
		}
		pred := invariant.Predicate{
			Schema: invariant.PredicateSchemaV1,
			Root:   invariant.Node{Op: invariant.OpContains, Field: invariant.FieldPreserves, CanonicalID: pcMeanGrowth},
		}

		// GUARD ON: the production mapping filters to ACCEPTED declarations;
		// absence stays unknown (DECLARATION_NOT_AUTHORITY holds).
		vocab := canon.MechanismV1()
		prodSig := canon.BuildSignature(mechanismInputFromDetail(d), vocab)
		prodVerdict := invariant.Evaluate(pred, prodSig)

		// MUTANT: the pre-v26 overlay consumed every declaration.
		in := mechanismInputFromDetail(d)
		in.DeclaredCompleteness = map[domain.FieldKind]domain.FieldCompleteness{}
		for _, fc := range d.FieldCompleteness {
			kind, kerr := domain.AttributeFieldKind(fc.Kind)
			if kerr != nil {
				continue
			}
			in.DeclaredCompleteness[kind] = fc.Completeness // admission ignored
		}
		mutSig := canon.BuildSignature(in, vocab)
		mutVerdict := invariant.Evaluate(pred, mutSig)

		// Only the mean-growth-absent approaches discriminate.
		if prodVerdict == invariant.VerdictSatisfies {
			continue
		}
		if prodVerdict != invariant.VerdictUnknown {
			t.Fatalf("guard on: absence must stay unknown, got %q", prodVerdict)
		}
		// must_fail DECLARATION_NOT_AUTHORITY: under the mutant the same
		// absence becomes a verified violation — from provider prose alone.
		if mutVerdict != invariant.VerdictViolates {
			t.Fatalf("mutant must reproduce the laundering (violates), got %q", mutVerdict)
		}
		return
	}
	t.Fatal("no discriminating approach found in the fixture")
}
