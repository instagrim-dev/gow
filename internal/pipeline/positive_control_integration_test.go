package pipeline

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/experiment"
	"github.com/instagrim-dev/newf/internal/invariant"
	"github.com/instagrim-dev/newf/internal/provider"
	"github.com/instagrim-dev/newf/internal/store"
)

// This file is the M7 POSITIVE CONTROL: a deliberately synthetic fixture with
// KNOWN GROUND TRUTH that drives the full research path to actual records, not
// just successful command exits:
//
//	scoped failure evidence  (the seeded corpus's failure-side families)
//	-> comparable mechanisms (all axes resolved -> real clustering, not "cannot compare")
//	-> nonempty candidate set (deterministic miner -> preserves contains mean_growth_rate)
//	-> executed challenges   (a completed-negative campaign -> surviving)
//	-> eligible guided target (surviving invariant B3 may attack)
//	-> generated proposal    (a mechanism that VERIFIABLY breaks the invariant,
//	   asserted by reading back the persisted per-target violation verdict)
//	-> decisive recovery     (the proposal is mechanism-near the WITHHELD target)
//
// It is the complement to corpus/experiments/2026-09-10-esr-negative-control:
// that run proved the empty/unresolved-input handling paths (expected
// abstention); this proves the populated discovery path fires when usable
// evidence is actually supplied.
//
// SCOPE (narrowed per review of 849a701): B0-vs-B3 here is a TARGET-CONDITIONED
// GENERATION POSITIVE CONTROL PAIRED WITH A NO-TARGET/NO-PROPOSAL CONTROL — the
// generator returns the authored signature per target and nothing without
// targets, so B0's emptiness and B3's output are encoded by the fixture. It
// shows the treatment path can be invoked and scored; it is NOT an
// effectiveness comparison against an undirected proposer.
//
// The mutation controls come in two tiers:
//   - component-level mutation controls (TestPositiveControlMutation*): direct
//     provider/evaluator calls pinning one boundary condition each;
//   - same-path integration mutations (TestIntegrationPositiveControlSamePath*,
//     TestIntegrationPositiveControlUnobservedCompleteness*): the SAME pipeline
//     configuration and persisted records, asserting the intermediate rows and
//     the downstream assessment across the storage/reporting boundary.
//
// Two structural facts the persisted CLI/vocab path cannot express are handled
// here and documented in corpus/experiments/2026-09-10-positive-control:
//  1. The signature builder (internal/canon/signature.go) marks every set
//     field CompletenessUnobserved, so the builder cannot establish
//     ABSENCE-BASED violations of positive `contains` predicates (absence stays
//     unknown unless the field was exhaustively extracted; enum mismatches and
//     negated predicates can still verify without proving set absence). A
//     verified absence-break therefore requires a signature whose preserves
//     field is Complete — expressible only when the proposed signature is
//     authored directly, not rehydrated from a persisted record. The generator
//     below sets it under this test's synthetic ground-truth convention; a
//     production fix must scope + justify completeness, never let a provider
//     assert it unqualified.
//  2. The default deriving fixture generator's ONLY break preserves an
//     out-of-vocabulary id (core.property.global_coupling), so a vocab-normalized
//     target can never be mechanism-near it — which is exactly why the shipped
//     end-to-end test yields no_recovery. The positive control uses an in-vocab
//     preserves id (residue_locality) shared by proposal and target so recovery
//     is reachable and decisive.

const (
	pcResidueLocality = "domain.number_theory.property.residue_locality"
	pcMeanGrowth      = "domain.number_theory.property.mean_growth_rate"
)

// pcGenerator is the positive-control stand-in for a live directed generator:
// TARGET-CONDITIONED by construction. Given >=1 surviving target (the B3 arm),
// it emits one proposal per authored signature, each targeting the first
// surviving invariant; given NO targets (the B0 arm) it emits nothing. B0's
// emptiness and B3's output are therefore ENCODED BY THE FIXTURE — this pairs a
// target-conditioned generation positive control with a no-target/no-proposal
// control; it is not an effectiveness comparison against an undirected proposer.
type pcGenerator struct {
	signatures []canon.MechanismSignature
}

// TrustedStructureAuthor: the positive-control generator authors SYNTHETIC
// GROUND TRUTH (incl. deliberate completeness), so it bypasses the untrusted
// proposal-admission boundary exactly like the in-repo deterministic fixtures.
func (pcGenerator) TrustedStructureAuthor() {}

func (g pcGenerator) Generate(_ context.Context, req provider.GenerationRequest) (provider.GenerationResponse, error) {
	var proposals []provider.FrontierProposal
	if len(req.Targets) > 0 {
		for i, sig := range g.signatures {
			proposals = append(proposals, provider.FrontierProposal{
				ProposedSignature:         sig,
				TargetInvariantIDs:        []string{req.Targets[0].InvariantID},
				StructuralViolationClaim:  "drops the shared mean-growth property for residue-local structure (variant " + string(rune('a'+i)) + ")",
				NoveltyArgument:           "no known failure family carries this exact resolved structure",
				CheapestFalsificationPath: "check the residue-local structure degenerates into the known families",
				ExpectedInformationGain:   domain.OrdinalMedium,
				EvaluationCost:            domain.OrdinalLow,
			})
		}
	}
	return provider.GenerationResponse{
		Proposals:       proposals,
		Metadata:        provider.Metadata{ProviderName: "fixture", ProviderVersion: "test", ModelName: "positive-control-generator", SchemaVersion: canon.SchemaMechanismV1},
		RequestPayload:  "req",
		ResponsePayload: "resp",
	}, nil
}

// pcSignature authors a directly-constructed canonical signature: preserves
// exactly one resolved id, optionally one resolved operator, with the preserves
// field marked Complete (exhaustively extracted) unless completePreserves is
// false. Under the mined invariant `preserves contains mean_growth_rate`, a
// signature preserving ONLY residue_locality with Complete preserves is a
// VERIFIED violation (absence is decidable); with Unobserved preserves the
// same structure evaluates unknown — that distinction is exactly what the
// completeness-flip mutation pins.
func pcSignature(preservesID, operatorID string, completePreserves bool) canon.MechanismSignature {
	sig := canon.MechanismSignature{
		SchemaVersion:     canon.SchemaMechanismV1,
		VocabularyVersion: canon.SchemaMechanismV1,
		Preserves: []canon.FieldClaim{{
			FieldKind: domain.FieldPreserves, State: domain.ResolutionResolved,
			CanonicalID: domain.CanonicalID(preservesID), Status: domain.ClaimInferred,
		}},
		Representations:  []canon.FieldClaim{},
		Operators:        []canon.FieldClaim{},
		Assumptions:      []canon.FieldClaim{},
		Breaks:           []canon.FieldClaim{},
		AuxiliaryObjects: []canon.FieldClaim{},
		Posture:          canon.Posture{Locality: domain.LocalityGlobal, Construction: domain.ConstructionConstructive, Uncertainty: domain.UncertaintyDeterministic},
		OutcomeClass:     domain.OutcomeUnknown,
	}
	if operatorID != "" {
		sig.Operators = append(sig.Operators, canon.FieldClaim{
			FieldKind: domain.FieldOperator, State: domain.ResolutionResolved,
			CanonicalID: domain.CanonicalID(operatorID), Status: domain.ClaimInferred,
		})
	}
	if completePreserves {
		sig.SetFieldCompleteness = map[domain.FieldKind]domain.FieldCompleteness{
			domain.FieldPreserves:       domain.CompletenessComplete,
			domain.FieldOperator:        domain.CompletenessComplete,
			domain.FieldAssumption:      domain.CompletenessComplete,
			domain.FieldBreaks:          domain.CompletenessComplete,
			domain.FieldAuxiliaryObject: domain.CompletenessComplete,
		}
	}
	return sig
}

// pcExpectedPredicate is the KNOWN ground-truth predicate the deterministic
// miner must produce on this corpus: minPreservesMiner picks the lexically
// smallest preserves id across all families, which for the four-families corpus
// is mean_growth_rate. The read-back chain asserts the persisted candidate
// carries exactly this predicate's fingerprint.
func pcExpectedPredicate() invariant.Predicate {
	return invariant.Predicate{
		Schema: invariant.PredicateSchemaV1,
		Root:   invariant.Node{Op: invariant.OpContains, Field: invariant.FieldPreserves, CanonicalID: pcMeanGrowth},
	}
}

// pcTargetFixture is the WITHHELD target's mechanism: a single approach whose
// preserves label resolves through the REAL vocabulary-admission path. With
// "residue locality" it resolves to residue_locality — the mechanistic twin of
// the recovering proposal (the "advance that escaped the failure invariant": it
// does NOT preserve mean_growth_rate). With an unmapped label the claim stays
// UNRESOLVED, which the unknown-axis mutation uses.
func pcTargetFixture(preservesLabel string) *MechanismFixture {
	return &MechanismFixture{
		Approaches: []MechanismFixtureApproach{{
			LogicalIdentity:  "withheld-advance",
			Label:            "Withheld structural advance",
			Locality:         "global",
			ConstructionMode: "constructive",
			UncertaintyMode:  "deterministic",
			Preserves:        []string{preservesLabel},
			Outcome:          MechanismFixtureOutcome{Class: "success"},
		}},
	}
}

// seedPositiveControl builds the full substrate: the TRAIN failure atlas mined
// + challenged to a surviving invariant (deterministically preserves contains
// mean_growth_rate via minPreservesMiner), and a WITHHELD target problem whose
// canonical signature resolves through the vocabulary-admission path. Returns
// the two problem ids and the surviving invariant id.
func seedPositiveControl(t *testing.T, ctx context.Context, app *App, dbPath, targetPreservesLabel string) (trainProblem, targetProblem, survivingInvID string) {
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

	// Withheld target problem: label -> canonical id through the REAL admission path.
	targetProblem, targetRun, targetSnap := seedOtherProblem(t, ctx, dbPath, now.Add(time.Hour))
	seed, err := app.SeedMechanismFixture(ctx, MechanismFixtureSeedInput{
		DBPath: dbPath, ProblemID: targetProblem, RunID: targetRun, SnapshotID: targetSnap,
		Fixture: pcTargetFixture(targetPreservesLabel),
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
// control: the whole path produces actual records and the target-conditioned
// arm decisively RECOVERS the withheld structural move while the no-target arm
// stays empty by construction. The read-back chain ties expected predicate
// fingerprint -> surviving invariant id -> proposal target id -> persisted
// signature revision -> violation verdict == violates -> recovered member.
func TestIntegrationPositiveControlDecisiveRecovery(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = minPreservesMiner{} // deterministic mined predicate (mean_growth_rate)
	app.challengerFn = biasOnlyChallenger{}    // a support recount, NOT a counterexample attack
	// Governs B0 (no targets -> nothing) and B3 (targets -> the authored break).
	app.generatorFn = pcGenerator{signatures: []canon.MechanismSignature{pcSignature(pcResidueLocality, "", true)}}

	trainProblem, targetProblem, invID := seedPositiveControl(t, ctx, app, dbPath, "residue locality")

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

	// B0 (no-target control): the target-conditioned generator emits nothing.
	if b0.ProposalCount != 0 || b0.Recovered {
		t.Fatalf("no-target B0 must be empty by construction: %+v", b0)
	}
	// B3 (target-conditioned): generated against the surviving invariant AND
	// decisively recovered.
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

	// --- Read-back chain (review of 849a701, finding 1): the recovery verdict
	// alone does not establish the intervening VERIFIED-BREAK step, because
	// AssessProposals compares proposal<->target and never consumes the
	// invariant-violation verdict. Assert the production path computed AND
	// RETAINED that result, tied together by ids, not by construction.
	repo := openTestStore(t, ctx, dbPath)
	defer repo.Close()

	// (1) expected predicate fingerprint -> surviving invariant id.
	assertCandidateFingerprint(t, ctx, repo, trainProblem, invID, pcExpectedPredicate().Fingerprint())

	// (2) recovered member -> proposal id.
	recoveredID := ""
	for _, m := range b3.Members {
		if m.Assessment == "recovered" {
			recoveredID = m.ProposalID
		}
	}
	if recoveredID == "" {
		t.Fatalf("no recovered membership row persisted: %+v", b3.Members)
	}

	// (3) proposal target id + persisted violation verdict + signature revision.
	gen, err := repo.GetFrontierGeneration(ctx, b3.FrontierGenerationRun)
	if err != nil {
		t.Fatalf("read back B3 generation: %v", err)
	}
	var prop *store.FrontierProposalRow
	for i := range gen.Proposals {
		if gen.Proposals[i].ID == recoveredID {
			prop = &gen.Proposals[i]
		}
	}
	if prop == nil {
		t.Fatalf("recovered proposal %s not owned by B3 generation %s", recoveredID, gen.ID)
	}
	if len(prop.Targets) != 1 || prop.Targets[0].InvariantID != invID {
		t.Fatalf("recovered proposal must target exactly the surviving invariant %s, got %+v", invID, prop.Targets)
	}
	if prop.Targets[0].Verdict != "violates" || !prop.Targets[0].Violated || !prop.ViolatesAnyTarget {
		t.Fatalf("persisted violation verdict must be a VERIFIED break: %+v", prop.Targets[0])
	}

	// (4) the exact persisted signature revision the arm assessed.
	occ, err := repo.ListGenerationOccurrenceContents(ctx, gen.ID)
	if err != nil {
		t.Fatalf("read back occurrence contents: %v", err)
	}
	oc, ok := occ[recoveredID]
	if !ok || oc.ContentHash == "" || oc.SignatureJSON == "" {
		t.Fatalf("recovered proposal must have a persisted signature revision: %+v", oc)
	}
	if !strings.Contains(oc.SignatureJSON, pcResidueLocality) {
		t.Fatalf("persisted signature revision must carry the authored preserves id, got %s", oc.SignatureJSON)
	}
}

// assertCandidateFingerprint reads back the persisted candidate invariant and
// asserts it carries exactly the expected predicate fingerprint — the head of
// the verified-break chain.
func assertCandidateFingerprint(t *testing.T, ctx context.Context, repo *store.Store, problemID, invID, wantFP string) {
	t.Helper()
	revs, err := repo.ListInvariantRevisions(ctx, problemID)
	if err != nil {
		t.Fatalf("list invariant revisions: %v", err)
	}
	for _, header := range revs {
		rev, err := repo.GetInvariantRevision(ctx, header.ID)
		if err != nil {
			t.Fatalf("get invariant revision %s: %v", header.ID, err)
		}
		for _, c := range rev.Candidates {
			if c.ID == invID {
				if c.PredicateFingerprint != wantFP {
					t.Fatalf("surviving invariant %s predicate fingerprint = %q, want %q", invID, c.PredicateFingerprint, wantFP)
				}
				return
			}
		}
	}
	t.Fatalf("surviving invariant %s not found in persisted revisions", invID)
}

// TestIntegrationPositiveControlUnobservedCompletenessBreaksVerification is the
// targeted completeness-flip mutation (review of 849a701, finding 1): the SAME
// resolved features, but preserves completeness drops from Complete to
// Unobserved. The invariant check must become UNKNOWN (absence is no longer
// decidable) while structural similarity to the withheld target is unchanged —
// recovery stays decisive. This proves the control distinguishes RECOVERING A
// REPRESENTATION from ESTABLISHING A DECIDABLE PREDICATE VIOLATION.
func TestIntegrationPositiveControlUnobservedCompletenessBreaksVerification(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = minPreservesMiner{}
	app.challengerFn = biasOnlyChallenger{}
	// Identical resolved structure; ONLY completeness flips to unobserved.
	app.generatorFn = pcGenerator{signatures: []canon.MechanismSignature{pcSignature(pcResidueLocality, "", false)}}

	trainProblem, targetProblem, invID := seedPositiveControl(t, ctx, app, dbPath, "residue locality")
	if _, err := app.DefineExperiment(ctx, ExperimentDefineInput{DBPath: dbPath, ProblemID: trainProblem, TargetProblemID: targetProblem}); err != nil {
		t.Fatalf("define: %v", err)
	}
	res, err := app.RunExperiment(ctx, ExperimentRunInput{DBPath: dbPath, ProblemID: trainProblem})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	b3 := armByName(t, res.Experiment, "b3_invariant_guided")
	if b3.ProposalCount == 0 {
		t.Fatalf("B3 must still generate: %+v", b3)
	}
	// Recovery is UNCHANGED: comparison reads resolved ids, not completeness.
	if !b3.Recovered || b3.UnknownCount != 0 {
		t.Fatalf("recovery must remain decisive under unobserved completeness: %+v", b3)
	}

	// But the persisted violation verdict must degrade to UNKNOWN: absence of
	// mean_growth_rate is no longer decidable without exhaustive extraction.
	repo := openTestStore(t, ctx, dbPath)
	defer repo.Close()
	gen, err := repo.GetFrontierGeneration(ctx, b3.FrontierGenerationRun)
	if err != nil {
		t.Fatalf("read back generation: %v", err)
	}
	if len(gen.Proposals) == 0 {
		t.Fatal("generation owns no proposals")
	}
	prop := gen.Proposals[0]
	if len(prop.Targets) != 1 || prop.Targets[0].InvariantID != invID {
		t.Fatalf("proposal must target the surviving invariant %s: %+v", invID, prop.Targets)
	}
	if prop.Targets[0].Verdict != "unknown" || prop.Targets[0].Violated || prop.ViolatesAnyTarget {
		t.Fatalf("unobserved completeness must persist an UNKNOWN violation verdict, got %+v", prop.Targets[0])
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

// TestPositiveControlMutationRemoveEvidence is a COMPONENT-LEVEL mutation
// control (direct provider call; the same-path integration variant is
// TestIntegrationPositiveControlSamePathMatrix/support_threshold): with only
// ONE failure family preserving residue_locality, the CLI-default deriving
// miner derives NO candidate. Remove the shared support -> no
// failure-invariant task.
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

// TestPositiveControlMutationUnknownFieldIsInconclusive is a COMPONENT-LEVEL
// mutation control (direct evaluator call; the same-path integration variant is
// TestIntegrationPositiveControlSamePathMatrix/unknown_target_axis): make the
// target's decisive axis UNKNOWN (an unresolved preserves claim). The
// comparison is then incomparable on preserves -> the proposal is assessed
// `unknown`, never coerced into recovery OR decisive non-recovery.
func TestPositiveControlMutationUnknownFieldIsInconclusive(t *testing.T) {
	profile := canon.ProfileMechanismV1()
	proposal := experiment.ProposalContent{ProposalID: "fpr_p", Rank: 0, Signature: pcSignature(pcResidueLocality, "", true)}

	// Target with an UNRESOLVED preserves claim -> incomparable on the decisive axis.
	unknownTarget := pcSignature(pcResidueLocality, "", true)
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

// TestPositiveControlMutationBudgetExhaustion is a COMPONENT-LEVEL mutation
// control (direct evaluator call; the same-path integration variant is
// TestIntegrationPositiveControlSamePathMatrix/budget_exhaustion): with two
// targets and a budget of exactly one comparison, the proposal cannot finish
// its comparisons and is UNASSESSED — budget is a stopping condition with
// recorded consumption, never a silent inclusion or a coerced negative.
func TestPositiveControlMutationBudgetExhaustion(t *testing.T) {
	profile := canon.ProfileMechanismV1()
	// A proposal that is DISTINCT from both targets so no early recovery shortcut
	// ends the scan before the budget bites.
	proposal := experiment.ProposalContent{ProposalID: "fpr_p", Rank: 0, Signature: pcSignature(pcMeanGrowth, "", true)}

	targetA := pcSignature(pcResidueLocality, "", true)
	targetB := pcSignature(pcResidueLocality, "", true)
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

// pcAtlasFixture authors a failure atlas with nShared failure approaches all
// preserving "residue locality" (each with a DISTINCT operator so they cluster
// as separate mechanism families) plus one partial_success contrast that also
// preserves it. Every label resolves in mechanism/v1, so the corpus enters
// through the REAL vocabulary-admission path.
func pcAtlasFixture(nShared int) *MechanismFixture {
	operators := []string{"modular decomposition", "density averaging"}
	fx := &MechanismFixture{}
	for i := 0; i < nShared; i++ {
		fx.Approaches = append(fx.Approaches, MechanismFixtureApproach{
			LogicalIdentity:  "failed-residue-" + string(rune('a'+i)),
			Label:            "Failed residue-local method " + string(rune('A'+i)),
			Locality:         "local",
			ConstructionMode: "constructive",
			UncertaintyMode:  "deterministic",
			Operators:        []string{operators[i%len(operators)]},
			Preserves:        []string{"residue locality"},
			Outcome:          MechanismFixtureOutcome{Class: "failure"},
		})
	}
	fx.Approaches = append(fx.Approaches, MechanismFixtureApproach{
		LogicalIdentity:  "contrast-partial",
		Label:            "Partial success preserving residue locality",
		Locality:         "local",
		ConstructionMode: "constructive",
		UncertaintyMode:  "deterministic",
		Assumptions:      []string{"residue independence"},
		Preserves:        []string{"residue locality"},
		Outcome:          MechanismFixtureOutcome{Class: "partial_success"},
	})
	return fx
}

// seedAtlasAndMine seeds pcAtlasFixture(nShared) into a fresh problem and runs
// signature -> cluster -> failure-space -> MineInvariants with the CLI-DEFAULT
// deriving miner and an EXPLICIT MinSupport, returning the mined revision view.
func seedAtlasAndMine(t *testing.T, ctx context.Context, app *App, dbPath string, nShared, minSupport int) InvariantMineResponse {
	t.Helper()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	problemID, runID, snapshotID := seedProblemAndSnapshot(t, ctx, dbPath, now)
	seed, err := app.SeedMechanismFixture(ctx, MechanismFixtureSeedInput{
		DBPath: dbPath, ProblemID: problemID, RunID: runID, SnapshotID: snapshotID,
		Fixture: pcAtlasFixture(nShared),
	})
	if err != nil {
		t.Fatalf("seed atlas: %v", err)
	}
	for _, mechID := range seed.MechanismIDs {
		if _, err := app.SignatureMechanism(ctx, SignatureInput{DBPath: dbPath, MechanismID: mechID, VocabVersion: canon.VocabularyMechanismV1}); err != nil {
			t.Fatalf("signature %s: %v", mechID, err)
		}
	}
	if _, err := app.BuildClustering(ctx, ClusterBuildInput{DBPath: dbPath, ProblemID: problemID}); err != nil {
		t.Fatalf("cluster build: %v", err)
	}
	if _, err := app.BuildFailureSpace(ctx, FailureSpaceBuildInput{DBPath: dbPath, ProblemID: problemID}); err != nil {
		t.Fatalf("failure-space build: %v", err)
	}
	mined, err := app.MineInvariants(ctx, InvariantMineInput{DBPath: dbPath, ProblemID: problemID, MinSupport: minSupport})
	if err != nil {
		t.Fatalf("mine: %v", err)
	}
	return mined
}

// TestIntegrationPositiveControlSamePathMatrix is the same-path integration
// mutation matrix (review of 849a701, finding 2): each row uses the SAME
// pipeline configuration in a fresh workspace, mutates one input, and asserts
// both the intermediate persisted records and the downstream result across the
// storage/reporting boundary — not a direct provider/evaluator call.
func TestIntegrationPositiveControlSamePathMatrix(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)

	// Row 1 — support threshold, through MineInvariants with the CLI-DEFAULT
	// deriving miner and an EXPLICIT MinSupport of 2. Two failure families
	// sharing residue_locality mine exactly that candidate with support 2;
	// removing one family (same path, same config) yields ZERO candidates.
	t.Run("support_threshold", func(t *testing.T) {
		wantFP := invariant.Predicate{
			Schema: invariant.PredicateSchemaV1,
			Root:   invariant.Node{Op: invariant.OpContains, Field: invariant.FieldPreserves, CanonicalID: pcResidueLocality},
		}.Fingerprint()

		app, dbPath := newRealStoreApp(t, now) // default (deriving) miner: invariantMinerFn nil
		populated := seedAtlasAndMine(t, ctx, app, dbPath, 2, 2)
		if populated.Revision.CandidateCount == 0 {
			t.Fatal("two failure families sharing residue_locality must mine a candidate")
		}
		found := false
		for _, c := range populated.Revision.Candidates {
			if c.PredicateFingerprint == wantFP {
				found = true
				if c.DistinctFamilySupport < 2 {
					t.Fatalf("persisted support = %d, want >= 2", c.DistinctFamilySupport)
				}
			}
		}
		if !found {
			t.Fatalf("mined candidates missing the shared-preserves predicate: %+v", populated.Revision.Candidates)
		}

		appStarved, dbStarved := newRealStoreApp(t, now)
		starved := seedAtlasAndMine(t, ctx, appStarved, dbStarved, 1, 2)
		if starved.Revision.CandidateCount != 0 {
			t.Fatalf("one failure family must mine ZERO candidates through the same path, got %d", starved.Revision.CandidateCount)
		}
	})

	// Row 2 — unknown target axis, through the FULL experiment path: the
	// withheld target's preserves label does not resolve, so the persisted arm
	// carries unknown_count=1, no recovery, and the experiment conclusion stays
	// inconclusive — read back via ShowExperiment (the reporting boundary).
	t.Run("unknown_target_axis", func(t *testing.T) {
		app, dbPath := newRealStoreApp(t, now)
		app.invariantMinerFn = minPreservesMiner{}
		app.challengerFn = biasOnlyChallenger{}
		app.generatorFn = pcGenerator{signatures: []canon.MechanismSignature{pcSignature(pcResidueLocality, "", true)}}

		trainProblem, targetProblem, _ := seedPositiveControl(t, ctx, app, dbPath, "an entirely unmapped surface property")
		if _, err := app.DefineExperiment(ctx, ExperimentDefineInput{DBPath: dbPath, ProblemID: trainProblem, TargetProblemID: targetProblem}); err != nil {
			t.Fatalf("define: %v", err)
		}
		if _, err := app.RunExperiment(ctx, ExperimentRunInput{DBPath: dbPath, ProblemID: trainProblem}); err != nil {
			t.Fatalf("run: %v", err)
		}
		shown, err := app.ShowExperiment(ctx, ExperimentShowInput{DBPath: dbPath, ProblemID: trainProblem})
		if err != nil {
			t.Fatalf("show: %v", err)
		}
		b3 := armByName(t, shown.Experiment, "b3_invariant_guided")
		if b3.ProposalCount != 1 || b3.Recovered {
			t.Fatalf("unresolved target axis must not recover: %+v", b3)
		}
		if b3.UnknownCount != 1 || b3.DecisiveCount != 0 {
			t.Fatalf("persisted arm must carry the unknown, not a coerced negative: %+v", b3)
		}
		if shown.Experiment.Conclusion != "inconclusive" {
			t.Fatalf("conclusion must stay inconclusive, got %q", shown.Experiment.Conclusion)
		}
	})

	// Row 3 — budget exhaustion, through the FULL experiment path: two distinct
	// non-recovering proposals under an evaluation budget of 1. The first is
	// decisively assessed; the second is UNASSESSED with the consumption
	// persisted; the stopping condition records budget_exhausted; and the
	// conclusion is inconclusive, never a coerced no_recovery.
	t.Run("budget_exhaustion", func(t *testing.T) {
		app, dbPath := newRealStoreApp(t, now)
		app.invariantMinerFn = minPreservesMiner{}
		app.challengerFn = biasOnlyChallenger{}
		// Two DISTINCT proposals (different resolved operators), both mechanism-
		// distinct from the withheld target (which has no operators).
		app.generatorFn = pcGenerator{signatures: []canon.MechanismSignature{
			pcSignature(pcResidueLocality, "core.operator.modular_decomposition", true),
			pcSignature(pcResidueLocality, "core.operator.density_averaging", true),
		}}

		trainProblem, targetProblem, _ := seedPositiveControl(t, ctx, app, dbPath, "residue locality")
		if _, err := app.DefineExperiment(ctx, ExperimentDefineInput{DBPath: dbPath, ProblemID: trainProblem, TargetProblemID: targetProblem}); err != nil {
			t.Fatalf("define: %v", err)
		}
		if _, err := app.RunExperiment(ctx, ExperimentRunInput{DBPath: dbPath, ProblemID: trainProblem, EvaluationBudget: 1}); err != nil {
			t.Fatalf("run: %v", err)
		}
		shown, err := app.ShowExperiment(ctx, ExperimentShowInput{DBPath: dbPath, ProblemID: trainProblem})
		if err != nil {
			t.Fatalf("show: %v", err)
		}
		b3 := armByName(t, shown.Experiment, "b3_invariant_guided")
		if b3.ProposalCount != 2 || b3.Recovered {
			t.Fatalf("expected two non-recovering proposals: %+v", b3)
		}
		if b3.DecisiveCount != 1 || b3.UnassessedCount != 1 || b3.EvaluationsConsumed != 1 {
			t.Fatalf("budget of 1 must persist decisive=1 unassessed=1 consumed=1: %+v", b3)
		}
		if b3.StoppingCondition != "budget_exhausted" {
			t.Fatalf("stopping condition must record budget_exhausted, got %q", b3.StoppingCondition)
		}
		if shown.Experiment.Conclusion != "inconclusive" {
			t.Fatalf("an unassessed proposal must keep the conclusion inconclusive, got %q", shown.Experiment.Conclusion)
		}
	})
}
