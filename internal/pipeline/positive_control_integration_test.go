package pipeline

import (
	"context"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/experiment"
	"github.com/instagrim-dev/newf/internal/provider"
)

// This file is the M7 POSITIVE CONTROL: a deliberately synthetic fixture with
// KNOWN GROUND TRUTH that drives the full research path to actual records, not
// just successful command exits:
//
//	scoped failure evidence  (3 failure/partial families sharing residue_locality)
//	-> comparable mechanisms (all axes resolved -> real clustering, not "cannot compare")
//	-> nonempty candidate set (miner derives preserves contains residue_locality, support>=2)
//	-> executed challenges   (a completed-negative campaign -> surviving)
//	-> eligible guided target (surviving invariant B3 may attack)
//	-> generated proposal    (a mechanism that VERIFIABLY breaks the invariant)
//	-> decisive recovery     (the proposal is mechanism-near the WITHHELD target)
//
// It is the complement to corpus/experiments/2026-09-10-esr-negative-control:
// that run proved the empty/unresolved-input handling paths (expected
// abstention); this proves the populated discovery path fires when usable
// evidence is actually supplied. The mutation battery below then flips each
// expected result for the corresponding reason, so a green result is evidence
// the path DISCRIMINATES, not merely that it runs.
//
// Two structural facts the persisted CLI/vocab path cannot express are handled
// here and documented in corpus/experiments/2026-09-10-positive-control:
//  1. The signature builder (internal/canon/signature.go) marks every set
//     field CompletenessUnobserved, so `contains`-absence evaluates to unknown,
//     never a verified violation. A verified break therefore requires a
//     signature whose preserves field is Complete — expressible only when the
//     proposed signature is authored directly (as a live generator would), not
//     rehydrated from a persisted record. The recovering generator below sets it.
//  2. The default deriving fixture generator's ONLY break preserves an
//     out-of-vocabulary id (core.property.global_coupling), so a vocab-normalized
//     target can never be mechanism-near it — which is exactly why the shipped
//     end-to-end test yields no_recovery. The positive control uses an in-vocab
//     complement (mean_growth_rate) shared by proposal and target so recovery is
//     reachable and decisive.

const (
	pcResidueLocality = "domain.number_theory.property.residue_locality"
	pcMeanGrowth      = "domain.number_theory.property.mean_growth_rate"
)

// recoveringGenerator is the positive-control stand-in for a live directed
// generator. Given >=1 surviving target (the B3 arm), it emits ONE proposal
// whose signature (a) preserves ONLY mean_growth_rate, exhaustively (so the
// engine can VERIFY it breaks preserves(residue_locality)), and (b) is
// mechanism-near the withheld target (which also preserves only mean_growth_rate)
// so recovery is decisive. Given NO targets (the B0 arm), it emits nothing —
// undirected search here honestly produces no proposal, so B0 vs B3 is a real
// guided-vs-undirected contrast, not two empty arms.
type recoveringGenerator struct{}

func (recoveringGenerator) Generate(_ context.Context, req provider.GenerationRequest) (provider.GenerationResponse, error) {
	var proposals []provider.FrontierProposal
	for _, tgt := range req.Targets {
		proposals = append(proposals, provider.FrontierProposal{
			ProposedSignature:         pcRecoverSignature(),
			TargetInvariantIDs:        []string{tgt.InvariantID},
			StructuralViolationClaim:  "abandons residue-local reasoning for a global mean-growth bound",
			NoveltyArgument:           "no known failure family preserves mean growth exhaustively while dropping residue locality",
			CheapestFalsificationPath: "check whether the mean-growth bound silently reintroduces a residue cover",
			ExpectedInformationGain:   domain.OrdinalMedium,
			EvaluationCost:            domain.OrdinalLow,
		})
	}
	return provider.GenerationResponse{
		Proposals:       proposals,
		Metadata:        provider.Metadata{ProviderName: "fixture", ProviderVersion: "test", ModelName: "positive-control-recovering", SchemaVersion: canon.SchemaMechanismV1},
		RequestPayload:  "req",
		ResponsePayload: "resp",
	}, nil
}

// pcRecoverSignature is the mechanism the recovering generator proposes AND the
// signature the withheld target carries (constructed identically so the
// comparison is mechanism-near). preserves is COMPLETE and does NOT contain
// residue_locality, so evaluating preserves(residue_locality) against it is a
// VERIFIED violation, not an unknown.
func pcRecoverSignature() canon.MechanismSignature {
	return canon.MechanismSignature{
		SchemaVersion:     canon.SchemaMechanismV1,
		VocabularyVersion: canon.SchemaMechanismV1,
		Preserves: []canon.FieldClaim{{
			FieldKind: domain.FieldPreserves, State: domain.ResolutionResolved,
			CanonicalID: pcMeanGrowth, Status: domain.ClaimInferred,
		}},
		Representations:  []canon.FieldClaim{},
		Operators:        []canon.FieldClaim{},
		Assumptions:      []canon.FieldClaim{},
		Breaks:           []canon.FieldClaim{},
		AuxiliaryObjects: []canon.FieldClaim{},
		Posture:          canon.Posture{Locality: domain.LocalityGlobal, Construction: domain.ConstructionConstructive, Uncertainty: domain.UncertaintyDeterministic},
		OutcomeClass:     domain.OutcomeUnknown,
		SetFieldCompleteness: map[domain.FieldKind]domain.FieldCompleteness{
			domain.FieldPreserves:       domain.CompletenessComplete,
			domain.FieldOperator:        domain.CompletenessComplete,
			domain.FieldAssumption:      domain.CompletenessComplete,
			domain.FieldBreaks:          domain.CompletenessComplete,
			domain.FieldAuxiliaryObject: domain.CompletenessComplete,
		},
	}
}

// pcTargetFixture is the WITHHELD target's mechanism: a single approach that
// preserves mean growth (resolves to mean_growth_rate in mechanism/v1) with a
// global/constructive/deterministic posture — the mechanistic twin of the
// recovering proposal. It is the "advance that escaped the failure invariant":
// it does NOT preserve residue locality, so a proposal reproducing its move
// recovers it.
func pcTargetFixture() *MechanismFixture {
	return &MechanismFixture{
		Approaches: []MechanismFixtureApproach{{
			LogicalIdentity:  "withheld-mean-growth-advance",
			Label:            "Global mean-growth bound (withheld advance)",
			Locality:         "global",
			ConstructionMode: "constructive",
			UncertaintyMode:  "deterministic",
			Preserves:        []string{"mean growth rate"},
			Outcome:          MechanismFixtureOutcome{Class: "success"},
		}},
	}
}

// seedPositiveControl builds the full substrate: the TRAIN failure atlas mined
// + challenged to a surviving invariant, and a WITHHELD target problem whose
// canonical signature is the mean-growth advance. Returns the two problem ids
// and the surviving invariant id.
func seedPositiveControl(t *testing.T, ctx context.Context, app *App, dbPath string) (trainProblem, targetProblem, survivingInvID string) {
	t.Helper()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)

	trainProblem, invID, _ := mineOneCandidate(t, ctx, app, dbPath)
	resp, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID})
	if err != nil {
		t.Fatalf("challenge: %v", err)
	}
	if got := resp.Reports[0].StateAfter; got != "surviving" {
		t.Fatalf("candidate must survive a completed-negative campaign, got %q", got)
	}

	// Withheld target problem with the mean-growth advance signature.
	targetProblem, targetRun, targetSnap := seedOtherProblem(t, ctx, dbPath, now.Add(time.Hour))
	seed, err := app.SeedMechanismFixture(ctx, MechanismFixtureSeedInput{
		DBPath: dbPath, ProblemID: targetProblem, RunID: targetRun, SnapshotID: targetSnap,
		Fixture: pcTargetFixture(),
	})
	if err != nil {
		t.Fatalf("seed target fixture: %v", err)
	}
	if len(seed.MechanismIDs) != 1 {
		t.Fatalf("target fixture must produce exactly one mechanism, got %d", len(seed.MechanismIDs))
	}
	if _, err := app.SignatureMechanism(ctx, SignatureInput{DBPath: dbPath, MechanismID: seed.MechanismIDs[0], VocabVersion: canon.VocabularyMechanismV1}); err != nil {
		t.Fatalf("target signature: %v", err)
	}
	return trainProblem, targetProblem, invID
}

// TestIntegrationPositiveControlDecisiveRecovery is the populated positive
// control: the whole path produces actual records and the guided arm decisively
// RECOVERS the withheld structural move while the undirected baseline does not.
func TestIntegrationPositiveControlDecisiveRecovery(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = dataDrivenMiner{}
	app.challengerFn = biasOnlyChallenger{}
	app.generatorFn = recoveringGenerator{} // governs B0 (no targets) and B3 (targets)

	trainProblem, targetProblem, _ := seedPositiveControl(t, ctx, app, dbPath)

	if _, err := app.DefineExperiment(ctx, ExperimentDefineInput{DBPath: dbPath, ProblemID: trainProblem, TargetProblemID: targetProblem}); err != nil {
		t.Fatalf("define: %v", err)
	}

	res, err := app.RunExperiment(ctx, ExperimentRunInput{DBPath: dbPath, ProblemID: trainProblem})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	exp := res.Experiment

	// Substrate assertions: the discovery path produced real records, not exits.
	if !exp.LeakageCheck.Passed {
		t.Fatalf("clean split must pass the leakage audit: %+v", exp.LeakageCheck)
	}
	b0, b3 := armByName(t, exp, "b0_undirected"), armByName(t, exp, "b3_invariant_guided")

	// B0 (undirected): no targets -> the recovering generator emits nothing.
	if b0.ProposalCount != 0 || b0.Recovered {
		t.Fatalf("undirected B0 must be honest-empty here: %+v", b0)
	}
	// B3 (guided): generated against the surviving invariant AND decisively recovered.
	if b3.ProposalCount == 0 || b3.FrontierGenerationRun == "" {
		t.Fatalf("guided B3 must generate against the surviving invariant: %+v", b3)
	}
	if !b3.Recovered {
		t.Fatalf("guided B3 must decisively RECOVER the withheld target: %+v", b3)
	}
	if b3.UnknownCount != 0 || b3.UnassessedCount != 0 {
		t.Fatalf("recovery must be decisive (no unknown/unassessed): %+v", b3)
	}
	if exp.Conclusion != "structural_recovery" {
		t.Fatalf("conclusion must record a guided structural recovery, got %q", exp.Conclusion)
	}
}

func armByName(t *testing.T, exp ExperimentView, name string) ExperimentArmView {
	t.Helper()
	for i := range exp.Arms {
		if exp.Arms[i].Arm == name {
			return exp.Arms[i]
		}
	}
	t.Fatalf("arm %q missing from %+v", name, exp.Arms)
	return ExperimentArmView{}
}

// --- Mutation battery: the result must change for the corresponding reason ---

// TestPositiveControlMutationRemoveEvidence: with only ONE failure family
// preserving residue_locality (support 1 < 2), the CLI-default deriving miner
// derives NO candidate. Remove the shared support -> no failure-invariant task.
func TestPositiveControlMutationRemoveEvidence(t *testing.T) {
	req := miningReqSharing(pcResidueLocality, 1) // one family only
	miner := provider.NewDerivingFixtureInvariantMiner()
	resp, err := miner.Mine(context.Background(), req)
	if err != nil {
		t.Fatalf("mine: %v", err)
	}
	if len(resp.Proposals) != 0 {
		t.Fatalf("single-family support must yield NO candidate (support>=2 required), got %d", len(resp.Proposals))
	}

	// Restore: two families sharing the id -> exactly one candidate.
	resp2, err := miner.Mine(context.Background(), miningReqSharing(pcResidueLocality, 2))
	if err != nil {
		t.Fatalf("mine restored: %v", err)
	}
	if len(resp2.Proposals) != 1 {
		t.Fatalf("two families sharing the id must yield exactly one candidate, got %d", len(resp2.Proposals))
	}
}

func miningReqSharing(id string, nFamilies int) provider.MiningRequest {
	req := provider.MiningRequest{}
	for i := 0; i < nFamilies; i++ {
		req.Families = append(req.Families, provider.MiningFamily{
			ClusterID:    "clus_" + string(rune('a'+i)),
			OutcomeClass: domain.OutcomeFailure,
			Preserves:    []domain.CanonicalID{domain.CanonicalID(id)},
		})
	}
	return req
}

// TestPositiveControlMutationUnknownFieldIsInconclusive: make the target's
// decisive axis UNKNOWN (an unresolved preserves claim). The comparison is then
// incomparable on preserves -> the proposal is assessed `unknown`, never coerced
// into recovery OR decisive non-recovery. Uncertainty survives.
func TestPositiveControlMutationUnknownFieldIsInconclusive(t *testing.T) {
	profile := canon.ProfileMechanismV1()
	proposal := experiment.ProposalContent{ProposalID: "fpr_p", Rank: 0, Signature: pcRecoverSignature()}

	// Target with an UNRESOLVED preserves claim -> incomparable on the decisive axis.
	unknownTarget := pcRecoverSignature()
	unknownTarget.Preserves = []canon.FieldClaim{{
		FieldKind: domain.FieldPreserves, State: domain.ResolutionUnknown, SurfaceLabel: "some unmapped property",
	}}

	out := experiment.AssessProposals([]experiment.ProposalContent{proposal}, []canon.MechanismSignature{unknownTarget}, profile, 0)
	if out.RecoveredCount != 0 {
		t.Fatalf("an unresolved decisive axis must NOT be coerced into recovery: %+v", out)
	}
	if out.DecisiveNoCount != 0 {
		t.Fatalf("an unresolved decisive axis must NOT be coerced into a decisive negative: %+v", out)
	}
	if out.UnknownCount != 1 {
		t.Fatalf("an unresolved decisive axis must be assessed unknown: %+v", out)
	}
}

// TestPositiveControlMutationBudgetExhaustion: with two targets and a budget of
// exactly one comparison, the (non-recovering-first) proposal cannot finish its
// comparisons and is UNASSESSED — budget is a stopping condition with recorded
// consumption, never a silent inclusion or a coerced negative.
func TestPositiveControlMutationBudgetExhaustion(t *testing.T) {
	profile := canon.ProfileMechanismV1()
	// A proposal that is DISTINCT from both targets so no early recovery shortcut
	// ends the scan before the budget bites.
	distinct := pcRecoverSignature()
	distinct.Preserves = []canon.FieldClaim{{
		FieldKind: domain.FieldPreserves, State: domain.ResolutionResolved,
		CanonicalID: pcResidueLocality, Status: domain.ClaimInferred,
	}}
	proposal := experiment.ProposalContent{ProposalID: "fpr_p", Rank: 0, Signature: distinct}

	targetA := pcRecoverSignature()
	targetB := pcRecoverSignature()
	targets := []canon.MechanismSignature{targetA, targetB}

	// Budget 1: only the first of two comparisons runs -> proposal UNASSESSED.
	starved := experiment.AssessProposals([]experiment.ProposalContent{proposal}, targets, profile, 1)
	if starved.UnassessedCount != 1 {
		t.Fatalf("a budget that cannot finish a proposal's comparisons must leave it unassessed: %+v", starved)
	}
	if starved.EvaluationsConsumed != 1 {
		t.Fatalf("consumed budget must be recorded exactly (1), got %d", starved.EvaluationsConsumed)
	}
	if starved.DecisiveNoCount != 0 {
		t.Fatalf("budget exhaustion must not manufacture a decisive negative: %+v", starved)
	}

	// Unlimited budget: both comparisons run -> decisive_no (distinct from both).
	full := experiment.AssessProposals([]experiment.ProposalContent{proposal}, targets, profile, 0)
	if full.DecisiveNoCount != 1 {
		t.Fatalf("with budget to finish, a distinct proposal is a decisive negative: %+v", full)
	}
}
