package pipeline

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/invariant"
	"github.com/instagrim-dev/newf/internal/provider"
	"github.com/instagrim-dev/newf/internal/store"
)

// untrustedGenerator simulates a LIVE MODEL proposer: it does NOT implement
// provider.TrustedStructureAuthor, so the production admission boundary must
// re-resolve its claims and strip its self-granted completeness.
type untrustedGenerator struct {
	proposals []provider.FrontierProposal
}

func (g untrustedGenerator) Generate(_ context.Context, req provider.GenerationRequest) (provider.GenerationResponse, error) {
	out := make([]provider.FrontierProposal, 0, len(g.proposals))
	if len(req.Targets) > 0 {
		for _, p := range g.proposals {
			p.TargetInvariantIDs = []string{req.Targets[0].InvariantID}
			out = append(out, p)
		}
	}
	return provider.GenerationResponse{
		Proposals:       out,
		Metadata:        provider.Metadata{ProviderName: "model-sim", ProviderVersion: "v1", ModelName: "untrusted-proposer", SchemaVersion: canon.SchemaMechanismV1},
		RequestPayload:  "req",
		ResponsePayload: "resp",
	}, nil
}

// TestIntegrationUntrustedProposalAdmission is the 040b8c9 finding-2
// regression at the PRODUCTION boundary: an untrusted proposer supplies (a) a
// claim whose surface label maps to canonical A while it asserts canonical B
// with state=resolved plus a self-granted `complete` flag, and (b) a
// version-incompatible proposal. The pipeline must code-resolve the label
// (never consume B), strip the completeness (so no absence-based verified
// break arises from provider assertion), and reject the incompatible
// proposal — with the raw provider response still auditable in the persisted
// invocation payload.
func TestIntegrationUntrustedProposalAdmission(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = minPreservesMiner{} // surviving invariant: preserves contains mean_growth_rate
	app.challengerFn = biasOnlyChallenger{}

	// (a) Inconsistent claim + self-granted completeness. If the provider's
	// resolution were consumed, the signature would PRESERVE mean_growth
	// (satisfies -> decisive refutation); if its completeness were accepted,
	// absence would be a verified violation. Neither may happen.
	inconsistent := canon.MechanismSignature{
		SchemaVersion:     canon.SchemaMechanismV1,
		VocabularyVersion: canon.VocabularyMechanismV1,
		Preserves: []canon.FieldClaim{{
			FieldKind: domain.FieldPreserves, SurfaceLabel: "residue locality",
			State: domain.ResolutionResolved, CanonicalID: pcMeanGrowth, Status: domain.ClaimInferred,
		}},
		Posture: canon.Posture{Locality: domain.LocalityGlobal, Construction: domain.ConstructionConstructive, Uncertainty: domain.UncertaintyDeterministic},
		SetFieldCompleteness: map[domain.FieldKind]domain.FieldCompleteness{
			domain.FieldPreserves: domain.CompletenessComplete,
		},
	}
	// (b) Version-incompatible proposal: must be rejected outright.
	incompatible := inconsistent
	incompatible.VocabularyVersion = "mechanism/v99"

	app.generatorFn = untrustedGenerator{proposals: []provider.FrontierProposal{
		{ProposedSignature: inconsistent, StructuralViolationClaim: "claims to break the invariant"},
		{ProposedSignature: incompatible, StructuralViolationClaim: "wrong vocabulary version"},
	}}

	problemID, invID, _ := mineOneCandidate(t, ctx, app, dbPath)
	if resp, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID}); err != nil || resp.Reports[0].StateAfter != "surviving" {
		t.Fatalf("challenge: %v / %+v", err, resp.Reports)
	}
	gen, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(gen.Generation.Proposals) != 1 {
		t.Fatalf("the version-incompatible proposal must be rejected: got %d proposals", len(gen.Generation.Proposals))
	}
	prop := gen.Generation.Proposals[0]

	// The persisted violation verdict must be UNKNOWN: the label code-resolves
	// to residue_locality (not the provider's mean_growth, which would read
	// SATISFIES) and the stripped completeness makes absence undecidable (a
	// verified VIOLATES from a self-granted flag is exactly the laundering the
	// gate exists to stop).
	if len(prop.Targets) != 1 || prop.Targets[0].InvariantID != invID {
		t.Fatalf("proposal must target the surviving invariant: %+v", prop.Targets)
	}
	if prop.Targets[0].Verdict != "unknown" || prop.Targets[0].Violated {
		t.Fatalf("an untrusted proposal must not earn a decisive break verdict from its own assertions: %+v", prop.Targets[0])
	}

	// The persisted admitted signature carries the CODE resolution + the
	// admission contract stamp, never the provider's canonical id.
	repo := openTestStore(t, ctx, dbPath)
	defer repo.Close()
	occ, err := repo.ListGenerationOccurrenceContents(ctx, gen.Generation.ID)
	if err != nil {
		t.Fatalf("occurrences: %v", err)
	}
	sigJSON := occ[prop.ID].SignatureJSON
	if !strings.Contains(sigJSON, pcResidueLocality) || strings.Contains(sigJSON, pcMeanGrowth) {
		t.Fatalf("admitted signature must carry the code resolution, never the provider's id: %s", sigJSON)
	}
	if !strings.Contains(sigJSON, canon.ProposalAdmissionContract) {
		t.Fatalf("admitted claims must be stamped with the admission contract: %s", sigJSON)
	}
	if strings.Contains(sigJSON, `"complete"`) {
		t.Fatalf("provider-declared completeness must be stripped from the admitted signature: %s", sigJSON)
	}

	// v29: the admission audit is PERSISTED on the generation — one corrected
	// claim, one stripped completeness, one rejected proposal — so an operator
	// can see what admission changed without diffing raw payloads.
	genRec, err := repo.GetFrontierGeneration(ctx, gen.Generation.ID)
	if err != nil {
		t.Fatalf("read back generation record: %v", err)
	}
	if genRec.AdmissionCorrected < 1 || genRec.AdmissionStripped != 1 || genRec.AdmissionRejected != 1 {
		t.Fatalf("admission audit must persist (corrected>=1 stripped=1 rejected=1): %+v",
			[]int{genRec.AdmissionCorrected, genRec.AdmissionDowngraded, genRec.AdmissionStripped, genRec.AdmissionRejected})
	}
	if gen.Generation.AdmissionCorrected != genRec.AdmissionCorrected || gen.Generation.AdmissionRejected != genRec.AdmissionRejected {
		t.Fatalf("the response view must expose the persisted audit: %+v vs %+v", gen.Generation, genRec)
	}
}

// TestIntegrationProposalsFileEntersThroughAdmission is the P7 end-to-end:
// captured model output (proposal-wire/v1, label-only) fed via
// `frontier generate --proposals-file` flows through the UNTRUSTED adapter and
// the production admission boundary — resolvable labels are code-resolved
// under the pinned vocabulary, unmapped labels stay unresolved, every claim
// carries the admission stamp, the audit is persisted, and no verified break
// arises without accepted completeness.
func TestIntegrationProposalsFileEntersThroughAdmission(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = minPreservesMiner{} // surviving invariant: preserves contains mean_growth_rate
	app.challengerFn = biasOnlyChallenger{}

	problemID, invID, _ := mineOneCandidate(t, ctx, app, dbPath)
	if resp, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID}); err != nil || resp.Reports[0].StateAfter != "surviving" {
		t.Fatalf("challenge: %v / %+v", err, resp.Reports)
	}

	wire := `{
  "schema_version": "proposal-wire/v1",
  "proposals": [{
    "mechanism": {
      "preserves": ["residue locality", "an entirely unmapped surface property"],
      "locality": "global",
      "construction_mode": "constructive",
      "uncertainty_mode": "deterministic"
    },
    "structural_violation_claim": "abandons the shared mean-growth property",
    "novelty_argument": "differs from every known family",
    "cheapest_falsification_path": "check degeneration",
    "expected_information_gain": "medium",
    "evaluation_cost": "low"
  }]
}`
	path := filepath.Join(t.TempDir(), "captured-model-output.json")
	if err := os.WriteFile(path, []byte(wire), 0o644); err != nil {
		t.Fatalf("write wire file: %v", err)
	}

	gen, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID, ProposalsFile: path})
	if err != nil {
		t.Fatalf("generate from proposals file: %v", err)
	}
	if len(gen.Generation.Proposals) != 1 {
		t.Fatalf("want 1 admitted proposal, got %d", len(gen.Generation.Proposals))
	}
	prop := gen.Generation.Proposals[0]

	// No verified break from label-only input: completeness is inexpressible
	// on the wire and never granted by admission.
	if len(prop.Targets) != 1 || prop.Targets[0].InvariantID != invID ||
		prop.Targets[0].Verdict != "unknown" || prop.ViolatesAnyTarget {
		t.Fatalf("wire input must not earn a verified break: %+v", prop.Targets)
	}

	repo := openTestStore(t, ctx, dbPath)
	defer repo.Close()
	occ, err := repo.ListGenerationOccurrenceContents(ctx, gen.Generation.ID)
	if err != nil {
		t.Fatalf("occurrences: %v", err)
	}
	var persisted canon.MechanismSignature
	if err := json.Unmarshal([]byte(occ[prop.ID].SignatureJSON), &persisted); err != nil {
		t.Fatalf("unmarshal persisted signature: %v", err)
	}
	if len(persisted.Preserves) != 2 {
		t.Fatalf("want 2 preserves claims, got %+v", persisted.Preserves)
	}
	byLabel := map[string]canon.FieldClaim{}
	for _, c := range persisted.Preserves {
		byLabel[c.SurfaceLabel] = c
		if c.ClassifierContract != canon.ProposalAdmissionContract {
			t.Fatalf("every admitted claim must carry the admission stamp: %+v", c)
		}
	}
	if got := byLabel["residue locality"]; got.State != domain.ResolutionResolved || got.CanonicalID != pcResidueLocality {
		t.Fatalf("resolvable label must be CODE-resolved under the pinned vocabulary: %+v", got)
	}
	if got := byLabel["an entirely unmapped surface property"]; got.State == domain.ResolutionResolved || got.CanonicalID != "" {
		t.Fatalf("an unmapped label must stay unresolved (no invented alias): %+v", got)
	}

	// The audit is persisted (label-only unresolved claims re-resolve, so both
	// claims count as corrections) and the raw wire payload is retained.
	genRec, err := repo.GetFrontierGeneration(ctx, gen.Generation.ID)
	if err != nil {
		t.Fatalf("generation record: %v", err)
	}
	if genRec.AdmissionCorrected < 1 || genRec.AdmissionRejected != 0 {
		t.Fatalf("admission audit must record the corrections: %+v",
			[]int{genRec.AdmissionCorrected, genRec.AdmissionDowngraded, genRec.AdmissionStripped, genRec.AdmissionRejected})
	}
	// Raw wire payload retention is proven at the adapter layer
	// (TestUntrustedProposerParsesLabelOnlyClaims); the generation read-back
	// does not rehydrate invocation payloads.
}

// seedTwoSurvivors builds a corpus whose failure families share BOTH a
// preserves id (residue_locality) and an operator id (modular_decomposition),
// so the CLI-default miner derives TWO candidates; both survive a bias-only
// campaign. Returns (problemID, preservesInvariantID, operatorInvariantID).
func seedTwoSurvivors(t *testing.T, ctx context.Context, app *App, dbPath string) (string, string, string) {
	t.Helper()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	problemID, runID, snapshotID := seedProblemAndSnapshot(t, ctx, dbPath, now)
	fixture := &MechanismFixture{Approaches: []MechanismFixtureApproach{
		{
			LogicalIdentity: "fail-shared-a", Label: "Failed shared A",
			Locality: "local", ConstructionMode: "constructive", UncertaintyMode: "deterministic",
			Operators: []string{"modular decomposition"}, Preserves: []string{"residue locality"},
			Outcome: MechanismFixtureOutcome{Class: "failure"},
		},
		{
			LogicalIdentity: "fail-shared-b", Label: "Failed shared B",
			Locality: "local", ConstructionMode: "constructive", UncertaintyMode: "deterministic",
			Operators: []string{"modular decomposition"}, Preserves: []string{"residue locality"},
			Assumptions: []string{"residue independence"}, // keeps the families distinct
			Outcome:     MechanismFixtureOutcome{Class: "failure"},
		},
		{
			LogicalIdentity: "contrast", Label: "Partial success contrast",
			Locality: "global", ConstructionMode: "constructive", UncertaintyMode: "deterministic",
			Preserves: []string{"mean growth rate"},
			Outcome:   MechanismFixtureOutcome{Class: "partial_success"},
		},
	}}
	seed, err := app.SeedMechanismFixture(ctx, MechanismFixtureSeedInput{DBPath: dbPath, ProblemID: problemID, RunID: runID, SnapshotID: snapshotID, Fixture: fixture})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	for _, mechID := range seed.MechanismIDs {
		if _, err := app.SignatureMechanism(ctx, SignatureInput{DBPath: dbPath, MechanismID: mechID, VocabVersion: canon.VocabularyMechanismV1}); err != nil {
			t.Fatalf("signature: %v", err)
		}
	}
	if _, err := app.BuildClustering(ctx, ClusterBuildInput{DBPath: dbPath, ProblemID: problemID}); err != nil {
		t.Fatalf("cluster: %v", err)
	}
	if _, err := app.BuildFailureSpace(ctx, FailureSpaceBuildInput{DBPath: dbPath, ProblemID: problemID}); err != nil {
		t.Fatalf("failure-space: %v", err)
	}
	mined, err := app.MineInvariants(ctx, InvariantMineInput{DBPath: dbPath, ProblemID: problemID, MinSupport: 2})
	if err != nil {
		t.Fatalf("mine: %v", err)
	}
	if mined.Revision.CandidateCount != 2 {
		t.Fatalf("want 2 candidates (shared preserves + shared operator), got %d", mined.Revision.CandidateCount)
	}
	var preservesInv, operatorInv string
	wantPreserves := invariant.Predicate{Schema: invariant.PredicateSchemaV1,
		Root: invariant.Node{Op: invariant.OpContains, Field: invariant.FieldPreserves, CanonicalID: pcResidueLocality}}.Fingerprint()
	for _, c := range mined.Revision.Candidates {
		if c.PredicateFingerprint == wantPreserves {
			preservesInv = c.ID
		} else {
			operatorInv = c.ID
		}
		if resp, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: c.ID}); err != nil || resp.Reports[0].StateAfter != "surviving" {
			t.Fatalf("challenge %s: %v / %+v", c.ID, err, resp.Reports)
		}
	}
	if preservesInv == "" || operatorInv == "" {
		t.Fatalf("could not identify both survivors: %+v", mined.Revision.Candidates)
	}
	return problemID, preservesInv, operatorInv
}

// TestIntegrationExplicitTargetingFixesTheClaim is the 512bc54 finding-1
// regression: with TWO survivors, an explicitly single-targeted wire proposal
// whose mechanism SATISFIES the untargeted operator invariant must NOT be
// refuted by that untargeted invariant — while the same wire under the
// documented break-all default IS decisively refuted.
func TestIntegrationExplicitTargetingFixesTheClaim(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.challengerFn = biasOnlyChallenger{}

	problemID, preservesInv, operatorInv := seedTwoSurvivors(t, ctx, app, dbPath)

	wireFor := func(targetLine string) string {
		return `{
  "schema_version": "proposal-wire/v1",
  "proposals": [{` + targetLine + `
    "mechanism": {
      "preserves": ["mean growth rate"],
      "operators": ["modular decomposition"],
      "locality": "global",
      "construction_mode": "constructive",
      "uncertainty_mode": "deterministic"
    },
    "structural_violation_claim": "escapes residue locality",
    "novelty_argument": "keeps the operator, changes the preserved property",
    "cheapest_falsification_path": "check the preserved set",
    "expected_information_gain": "medium",
    "evaluation_cost": "low"
  }]
}`
	}

	// Explicitly targeted: preserving the UNTARGETED operator invariant is not
	// a refutation; the intended target stays unknown (no accepted
	// completeness), so evaluation is non-decisive — never `failure`.
	path := filepath.Join(t.TempDir(), "targeted.json")
	if err := os.WriteFile(path, []byte(wireFor(`
    "target_invariant_ids": ["`+preservesInv+`"],`)), 0o644); err != nil {
		t.Fatalf("write wire: %v", err)
	}
	gen, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID, ProposalsFile: path})
	if err != nil {
		t.Fatalf("generate targeted: %v", err)
	}
	prop := gen.Generation.Proposals[0]
	if len(prop.Targets) != 1 || prop.Targets[0].InvariantID != preservesInv {
		t.Fatalf("the persisted claim must be the EXPLICIT subset: %+v", prop.Targets)
	}
	res, err := app.Evaluate(ctx, EvaluateInput{DBPath: dbPath, ProblemID: problemID, ProposalID: prop.ID})
	if err != nil {
		t.Fatalf("evaluate targeted: %v", err)
	}
	if v := res.Run.Evaluations[0].Verdict; v == "failure" {
		t.Fatalf("an untargeted invariant must not refute an explicitly targeted proposal (got %q)", v)
	}

	// Control — break-all default: the same mechanism SATISFIES the operator
	// invariant it now implicitly claims to break -> deterministic failure.
	pathAll := filepath.Join(t.TempDir(), "break-all.json")
	if err := os.WriteFile(pathAll, []byte(wireFor("")), 0o644); err != nil {
		t.Fatalf("write wire: %v", err)
	}
	genAll, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID, ProposalsFile: pathAll})
	if err != nil {
		t.Fatalf("generate break-all: %v", err)
	}
	propAll := genAll.Generation.Proposals[0]
	if len(propAll.Targets) != 2 {
		t.Fatalf("break-all default must claim both survivors: %+v", propAll.Targets)
	}
	resAll, err := app.Evaluate(ctx, EvaluateInput{DBPath: dbPath, ProblemID: problemID, ProposalID: propAll.ID})
	if err != nil {
		t.Fatalf("evaluate break-all: %v", err)
	}
	if v := resAll.Run.Evaluations[0].Verdict; v != "failure" {
		t.Fatalf("break-all semantics stay strict: preserving %s must refute (got %q)", operatorInv, v)
	}
}

// TestIntegrationCountBoundsExternalProposals is the 512bc54 finding-2
// regression: the requested count bounds admission/scoring on the external
// path — never left to provider cooperation. Excess proposals are
// deterministically truncated in wire order and the overflow is persisted.
func TestIntegrationCountBoundsExternalProposals(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = minPreservesMiner{}
	app.challengerFn = biasOnlyChallenger{}

	problemID, invID, _ := mineOneCandidate(t, ctx, app, dbPath)
	if resp, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID}); err != nil || resp.Reports[0].StateAfter != "surviving" {
		t.Fatalf("challenge: %v / %+v", err, resp.Reports)
	}

	wire := `{
  "schema_version": "proposal-wire/v1",
  "proposals": [
    {"mechanism": {"preserves": ["residue locality"], "locality": "global", "construction_mode": "constructive", "uncertainty_mode": "deterministic"},
     "structural_violation_claim": "first", "novelty_argument": "n1", "cheapest_falsification_path": "c1"},
    {"mechanism": {"preserves": ["mean growth rate"], "locality": "global", "construction_mode": "constructive", "uncertainty_mode": "deterministic"},
     "structural_violation_claim": "second", "novelty_argument": "n2", "cheapest_falsification_path": "c2"}
  ]
}`
	path := filepath.Join(t.TempDir(), "two-proposals.json")
	if err := os.WriteFile(path, []byte(wire), 0o644); err != nil {
		t.Fatalf("write wire: %v", err)
	}
	gen, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID, ProposalsFile: path, Count: 1})
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	if len(gen.Generation.Proposals) != 1 {
		t.Fatalf("count=1 must bound the scored membership, got %d proposals", len(gen.Generation.Proposals))
	}
	if gen.Generation.Proposals[0].StructuralViolationClaim != "first" {
		t.Fatalf("truncation must be deterministic in wire order: %+v", gen.Generation.Proposals[0].StructuralViolationClaim)
	}
	repo := openTestStore(t, ctx, dbPath)
	defer repo.Close()
	rec, err := repo.GetFrontierGeneration(ctx, gen.Generation.ID)
	if err != nil {
		t.Fatalf("record: %v", err)
	}
	if rec.AdmissionOverflow != 1 || rec.RequestedCount != 1 {
		t.Fatalf("overflow disposition must be persisted (overflow=1, requested=1): overflow=%d requested=%d", rec.AdmissionOverflow, rec.RequestedCount)
	}
	// 622fb6e finding 1 (ordering): the cap bounds ADMISSION work, not only
	// scoring. Each submitted proposal carries one unresolved label whose
	// admission is one correction — a persisted count of 1 (not 2) proves
	// vocabulary resolution ran only on the capped subset. Checking the final
	// proposal count alone cannot catch the ordering defect.
	if rec.AdmissionCorrected != 1 {
		t.Fatalf("admission must run on the CAPPED subset only (corrected=1, not 2): got %d", rec.AdmissionCorrected)
	}
}

// TestIntegrationRejectedWirePayloadIsRetained is the 512bc54 finding-3
// regression: a rejected wire payload (forbidden authority fields) fails the
// run AND survives as a durable invocation envelope on that failed run — the
// exact attempted bytes are readable back from the audit trail.
func TestIntegrationRejectedWirePayloadIsRetained(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = minPreservesMiner{}
	app.challengerFn = biasOnlyChallenger{}

	problemID, invID, _ := mineOneCandidate(t, ctx, app, dbPath)
	if resp, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID}); err != nil || resp.Reports[0].StateAfter != "surviving" {
		t.Fatalf("challenge: %v / %+v", err, resp.Reports)
	}

	smuggled := `{
  "schema_version": "proposal-wire/v1",
  "proposals": [{
    "mechanism": {"preserves": ["residue locality"], "field_completeness": {"preserves": "complete"}, "locality": "global", "construction_mode": "constructive", "uncertainty_mode": "deterministic"},
    "structural_violation_claim": "x", "novelty_argument": "y", "cheapest_falsification_path": "z"
  }]
}`
	path := filepath.Join(t.TempDir(), "smuggled.json")
	if err := os.WriteFile(path, []byte(smuggled), 0o644); err != nil {
		t.Fatalf("write wire: %v", err)
	}
	if _, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID, ProposalsFile: path}); err == nil {
		t.Fatal("smuggled authority fields must reject the generation")
	}

	repo := openTestStore(t, ctx, dbPath)
	defer repo.Close()
	run, found, err := repo.LatestRunForOperation(ctx, problemID, "frontier generate")
	if err != nil || !found {
		t.Fatalf("failed run must exist: %v %v", err, found)
	}
	if run.Status != domain.RunStatusFailed {
		t.Fatalf("run status = %q, want failed", run.Status)
	}
	invs, err := repo.ListProviderInvocationsForRun(ctx, run.ID)
	if err != nil || len(invs) != 1 {
		t.Fatalf("the rejected attempt must leave ONE invocation envelope: %v %d", err, len(invs))
	}
	if invs[0].ResponsePayload != smuggled {
		t.Fatalf("the EXACT attempted payload must be readable back, got %q", invs[0].ResponsePayload)
	}
	if invs[0].RequestPayload == "" || invs[0].ProviderName != "external-file" {
		t.Fatalf("the envelope must carry request + provider identity: %+v", invs[0])
	}
}

// auditFailingStore delegates everything except failed-invocation retention,
// simulating a storage failure during rejection handling.
type auditFailingStore struct {
	problemStore
}

func (s auditFailingStore) RecordFailedProviderInvocation(context.Context, store.FrontierProviderInvocation) error {
	return errors.New("simulated audit storage failure")
}

// TestIntegrationAuditWriteFailureIsVisible is the 622fb6e finding-2
// regression: when the rejected payload CANNOT be retained, the returned
// failure exposes BOTH causes — the original wire rejection stays primary and
// the retention failure is appended — and the attempt is never represented as
// durably captured.
func TestIntegrationAuditWriteFailureIsVisible(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = minPreservesMiner{}
	app.challengerFn = biasOnlyChallenger{}

	problemID, invID, _ := mineOneCandidate(t, ctx, app, dbPath)
	if resp, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID}); err != nil || resp.Reports[0].StateAfter != "surviving" {
		t.Fatalf("challenge: %v / %+v", err, resp.Reports)
	}

	// From here on, failed-invocation retention is broken.
	inner := app.openStoreFn
	app.openStoreFn = func(ctx context.Context, dbPath string) (string, problemStore, error) {
		path, repo, err := inner(ctx, dbPath)
		if err != nil {
			return "", nil, err
		}
		return path, auditFailingStore{problemStore: repo}, nil
	}

	smuggled := `{"schema_version": "proposal-wire/v1", "proposals": [{
    "mechanism": {"preserves": ["residue locality"], "canonical_id": "core.x", "locality": "global", "construction_mode": "constructive", "uncertainty_mode": "deterministic"},
    "structural_violation_claim": "x", "novelty_argument": "y", "cheapest_falsification_path": "z"}]}`
	path := filepath.Join(t.TempDir(), "smuggled.json")
	if err := os.WriteFile(path, []byte(smuggled), 0o644); err != nil {
		t.Fatalf("write wire: %v", err)
	}
	_, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID, ProposalsFile: path})
	if err == nil {
		t.Fatal("generation must fail")
	}
	if !errors.Is(err, provider.ErrProposalWireViolation) {
		t.Fatalf("the ORIGINAL rejection must stay the primary cause: %v", err)
	}
	if !strings.Contains(err.Error(), "could NOT be retained") || !strings.Contains(err.Error(), "simulated audit storage failure") {
		t.Fatalf("the audit-write failure must be visible alongside the rejection: %v", err)
	}

	// The attempt is genuinely not captured: no invocation envelope exists.
	app.openStoreFn = inner
	repo := openTestStore(t, ctx, dbPath)
	defer repo.Close()
	run, found, ferr := repo.LatestRunForOperation(ctx, problemID, "frontier generate")
	if ferr != nil || !found {
		t.Fatalf("failed run: %v %v", ferr, found)
	}
	invs, ierr := repo.ListProviderInvocationsForRun(ctx, run.ID)
	if ierr != nil {
		t.Fatalf("list invocations: %v", ierr)
	}
	if len(invs) != 0 {
		t.Fatalf("the attempt must not be represented as captured: %+v", invs)
	}
}

// TestIntegrationExternalProposalArms is the pilot's minimum genuine
// comparison (package 3): the SAME external proposer route supplies B0
// (captured output whose permitted context had NO invariant targets) and B3
// (surviving invariants supplied), under equal predeclared budgets. Both
// arms' proposals enter through the production admission boundary; the
// harness retains the generation request, raw wire, admission audit, and
// assessed membership. Here B3's captured proposal is mechanism-near the
// withheld target and B0's is distinct — the harness must report exactly
// that, from genuinely imported (non-fixture-generator) content.
func TestIntegrationExternalProposalArms(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = minPreservesMiner{}
	app.challengerFn = biasOnlyChallenger{}

	trainProblem, targetProblem, invID := seedPositiveControl(t, ctx, app, dbPath, "residue locality")
	if _, err := app.DefineExperiment(ctx, ExperimentDefineInput{DBPath: dbPath, ProblemID: trainProblem, TargetProblemID: targetProblem}); err != nil {
		t.Fatalf("define: %v", err)
	}

	wire := func(preserves, claim string) string {
		return `{"schema_version": "proposal-wire/v1", "proposals": [{
  "mechanism": {"preserves": ["` + preserves + `"], "locality": "global", "construction_mode": "constructive", "uncertainty_mode": "deterministic"},
  "structural_violation_claim": "` + claim + `",
  "novelty_argument": "captured external output",
  "cheapest_falsification_path": "compare against the withheld family"}]}`
	}
	dir := t.TempDir()
	b0Path := filepath.Join(dir, "b0-captured.json")
	b3Path := filepath.Join(dir, "b3-captured.json")
	// B0 (no invariant context): proposes the mean-growth direction — distinct
	// from the withheld residue-locality target.
	if err := os.WriteFile(b0Path, []byte(wire("mean growth rate", "unguided direction")), 0o644); err != nil {
		t.Fatalf("write b0: %v", err)
	}
	// B3 (invariants in context): proposes the residue-locality structure —
	// mechanism-near the withheld target.
	if err := os.WriteFile(b3Path, []byte(wire("residue locality", "guided break of the shared property")), 0o644); err != nil {
		t.Fatalf("write b3: %v", err)
	}

	res, err := app.RunExperiment(ctx, ExperimentRunInput{
		DBPath: dbPath, ProblemID: trainProblem,
		ArmProposalFiles: map[string]string{"b0_undirected": b0Path, "b3_invariant_guided": b3Path},
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	exp := res.Experiment
	if !exp.LeakageCheck.Passed {
		t.Fatalf("leakage: %+v", exp.LeakageCheck)
	}
	b0, b3 := armByName(t, exp, "b0_undirected"), armByName(t, exp, "b3_invariant_guided")

	// B0 is no longer honest-empty: it carries the captured unguided proposal,
	// decisively assessed as non-recovering.
	if b0.ProposalCount != 1 || b0.Recovered || b0.DecisiveCount != 1 {
		t.Fatalf("B0 must carry the captured proposal, decisively non-recovering: %+v", b0)
	}
	// B3 carries the captured guided proposal and recovers the target.
	if b3.ProposalCount != 1 || !b3.Recovered {
		t.Fatalf("B3 must recover with the captured guided proposal: %+v", b3)
	}
	if exp.Conclusion != "structural_recovery" {
		t.Fatalf("conclusion = %q", exp.Conclusion)
	}

	// Both arm generations went through admission (label-only wire claims are
	// corrections) and the audit is persisted per arm.
	repo := openTestStore(t, ctx, dbPath)
	defer repo.Close()
	for _, arm := range []ExperimentArmView{b0, b3} {
		rec, gerr := repo.GetFrontierGeneration(ctx, arm.FrontierGenerationRun)
		if gerr != nil {
			t.Fatalf("generation %s: %v", arm.Arm, gerr)
		}
		if rec.AdmissionCorrected < 1 {
			t.Fatalf("%s must show the admission audit (label re-resolution): %+v", arm.Arm, rec.AdmissionCorrected)
		}
	}
	// B3's proposal genuinely targets the surviving invariant.
	genB3, err := repo.GetFrontierGeneration(ctx, b3.FrontierGenerationRun)
	if err != nil {
		t.Fatalf("b3 gen: %v", err)
	}
	if len(genB3.Proposals) != 1 || len(genB3.Proposals[0].Targets) != 1 || genB3.Proposals[0].Targets[0].InvariantID != invID {
		t.Fatalf("B3's captured proposal must target the surviving invariant: %+v", genB3.Proposals)
	}

	// Idempotent replay with the same captured files returns the SAME experiment.
	again, err := app.RunExperiment(ctx, ExperimentRunInput{
		DBPath: dbPath, ProblemID: trainProblem,
		ArmProposalFiles: map[string]string{"b0_undirected": b0Path, "b3_invariant_guided": b3Path},
	})
	if err != nil {
		t.Fatalf("replay: %v", err)
	}
	if again.Created || again.Experiment.ID != exp.ID {
		t.Fatalf("replay must be idempotent: created=%v id=%s want %s", again.Created, again.Experiment.ID, exp.ID)
	}

	// A file for a scripted control arm is refused.
	if _, err := app.RunExperiment(ctx, ExperimentRunInput{
		DBPath: dbPath, ProblemID: trainProblem,
		ArmProposalFiles: map[string]string{"b1_semantic_summary": b0Path},
	}); err == nil {
		t.Fatal("external proposals for a scripted control arm must be refused")
	}
}

// TestIntegrationExternalArmPreflight is the c860720 finding-1 regression:
// external-file options are validated BEFORE any run record exists. A file
// for an unselected arm is rejected (an explicitly supplied input either
// participates or is rejected, never silently ignored), a file for a
// scripted control arm is rejected, and neither rejection leaves a run row.
func TestIntegrationExternalArmPreflight(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = minPreservesMiner{}
	app.challengerFn = biasOnlyChallenger{}

	trainProblem, targetProblem, _ := seedPositiveControl(t, ctx, app, dbPath, "residue locality")
	if _, err := app.DefineExperiment(ctx, ExperimentDefineInput{DBPath: dbPath, ProblemID: trainProblem, TargetProblemID: targetProblem}); err != nil {
		t.Fatalf("define: %v", err)
	}
	path := filepath.Join(t.TempDir(), "b3.json")
	if err := os.WriteFile(path, []byte(`{"schema_version":"proposal-wire/v1","proposals":[]}`), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	// (a) File for an arm that is NOT selected.
	_, err := app.RunExperiment(ctx, ExperimentRunInput{
		DBPath: dbPath, ProblemID: trainProblem,
		Arms:             []string{"b0_undirected"},
		ArmProposalFiles: map[string]string{"b3_invariant_guided": path},
	})
	if err == nil || !strings.Contains(err.Error(), "not in the selected arm set") {
		t.Fatalf("a file for an unselected arm must be rejected explicitly, got %v", err)
	}

	// (b) File for a scripted control arm.
	_, err = app.RunExperiment(ctx, ExperimentRunInput{
		DBPath: dbPath, ProblemID: trainProblem,
		ArmProposalFiles: map[string]string{"b1_semantic_summary": path},
	})
	if err == nil || !strings.Contains(err.Error(), "scripted machinery controls") {
		t.Fatalf("a file for a scripted arm must be rejected, got %v", err)
	}

	// Neither rejection created (or abandoned) a run record.
	repo := openTestStore(t, ctx, dbPath)
	defer repo.Close()
	if _, found, ferr := repo.LatestRunForOperation(ctx, trainProblem, "experiment run"); ferr != nil || found {
		t.Fatalf("preflight rejection must not leave a run record: found=%v err=%v", found, ferr)
	}
}

// TestIntegrationExecutionAttributionSurvivesReuse is the c860720 finding-2
// regression: experiment identity keys on ASSESSED STRUCTURE, so a changed
// capture whose assessed structure is unchanged (only prose that
// ProposalHash excludes) reuses the experiment artifact — but each execution
// still records WHICH capture it assessed: its own run, generation, and file
// hash. Reuse never obscures the attribution.
func TestIntegrationExecutionAttributionSurvivesReuse(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = minPreservesMiner{}
	app.challengerFn = biasOnlyChallenger{}

	trainProblem, targetProblem, _ := seedPositiveControl(t, ctx, app, dbPath, "residue locality")
	if _, err := app.DefineExperiment(ctx, ExperimentDefineInput{DBPath: dbPath, ProblemID: trainProblem, TargetProblemID: targetProblem}); err != nil {
		t.Fatalf("define: %v", err)
	}
	wire := func(novelty string) string {
		return `{"schema_version": "proposal-wire/v1", "proposals": [{
  "mechanism": {"preserves": ["residue locality"], "locality": "global", "construction_mode": "constructive", "uncertainty_mode": "deterministic"},
  "structural_violation_claim": "guided break",
  "novelty_argument": "` + novelty + `",
  "cheapest_falsification_path": "compare"}]}`
	}
	dir := t.TempDir()
	fileA := filepath.Join(dir, "capture-a.json")
	fileB := filepath.Join(dir, "capture-b.json")
	if err := os.WriteFile(fileA, []byte(wire("first phrasing")), 0o644); err != nil {
		t.Fatalf("write A: %v", err)
	}
	// Only the novelty prose differs: ProposalHash excludes it, so the
	// assessed structure — and therefore the experiment identity — is unchanged.
	if err := os.WriteFile(fileB, []byte(wire("different phrasing, same mechanism")), 0o644); err != nil {
		t.Fatalf("write B: %v", err)
	}

	first, err := app.RunExperiment(ctx, ExperimentRunInput{
		DBPath: dbPath, ProblemID: trainProblem,
		ArmProposalFiles: map[string]string{"b3_invariant_guided": fileA},
	})
	if err != nil {
		t.Fatalf("run A: %v", err)
	}
	second, err := app.RunExperiment(ctx, ExperimentRunInput{
		DBPath: dbPath, ProblemID: trainProblem,
		ArmProposalFiles: map[string]string{"b3_invariant_guided": fileB},
	})
	if err != nil {
		t.Fatalf("run B: %v", err)
	}
	if second.Created || second.Experiment.ID != first.Experiment.ID {
		t.Fatalf("unchanged assessed structure must reuse the experiment artifact: %v %s vs %s", second.Created, second.Experiment.ID, first.Experiment.ID)
	}

	// The attribution record distinguishes the two executions: different runs,
	// different generations, different capture hashes — same experiment.
	repo := openTestStore(t, ctx, dbPath)
	defer repo.Close()
	execs, err := repo.ListExperimentExecutions(ctx, trainProblem)
	if err != nil {
		t.Fatalf("list executions: %v", err)
	}
	var b3execs []store.ExperimentExecutionRow
	for _, e := range execs {
		if e.Arm == "b3_invariant_guided" {
			b3execs = append(b3execs, e)
		}
	}
	if len(b3execs) != 2 {
		t.Fatalf("both executions must be attributed, got %d", len(b3execs))
	}
	a1, a2 := b3execs[0], b3execs[1]
	if a1.ExperimentID != first.Experiment.ID || a2.ExperimentID != first.Experiment.ID {
		t.Fatalf("both executions selected the same (reused) experiment: %+v %+v", a1, a2)
	}
	if a1.RunID == a2.RunID || a1.FrontierGenerationRun == a2.FrontierGenerationRun {
		t.Fatalf("executions must keep their own run + generation: %+v %+v", a1, a2)
	}
	if a1.ProposalsFileSHA256 == "" || a1.ProposalsFileSHA256 == a2.ProposalsFileSHA256 {
		t.Fatalf("the capture hashes must distinguish the two files: %q vs %q", a1.ProposalsFileSHA256, a2.ProposalsFileSHA256)
	}
}

// TestIntegrationExperimentReadiness pins the mechanical half of the pilot
// readiness decision: the full positive-control substrate reports READY with
// the expected per-check facts, and a bare problem reports NOT READY with
// each blocker NAMED — the report never manufactures readiness, and the
// non-mechanical protocol items surface as operator attestations.
func TestIntegrationExperimentReadiness(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = minPreservesMiner{}
	app.challengerFn = biasOnlyChallenger{}

	trainProblem, targetProblem, _ := seedPositiveControl(t, ctx, app, dbPath, "residue locality")
	if _, err := app.DefineExperiment(ctx, ExperimentDefineInput{DBPath: dbPath, ProblemID: trainProblem, TargetProblemID: targetProblem}); err != nil {
		t.Fatalf("define: %v", err)
	}

	ready, err := app.ExperimentReadiness(ctx, ExperimentReadinessInput{DBPath: dbPath, ProblemID: trainProblem})
	if err != nil {
		t.Fatalf("readiness: %v", err)
	}
	if !ready.Ready {
		t.Fatalf("full substrate must be mechanically ready: %+v", ready.Checks)
	}
	status := map[string]string{}
	for _, c := range ready.Checks {
		status[c.Check] = c.Status
	}
	for _, check := range []string{"failure_cohort", "decisive_axis_resolution", "surviving_invariants", "withheld_target"} {
		if status[check] != "ready" {
			t.Fatalf("check %s must be ready: %+v", check, ready.Checks)
		}
	}
	if status["completeness_admissions"] != "info" || status["recovery_criterion"] != "info" {
		t.Fatalf("informational checks must not gate: %+v", ready.Checks)
	}
	if len(ready.OperatorAttestations) == 0 {
		t.Fatal("the non-mechanical protocol items must surface as attestations")
	}

	// Bare problem: every substrate blocker is NAMED.
	bare, err := app.InitProblem(ctx, InitProblemInput{DBPath: dbPath, Statement: "bare readiness probe", Slug: "readiness-bare", ForceNew: true})
	if err != nil {
		t.Fatalf("init bare: %v", err)
	}
	blocked, err := app.ExperimentReadiness(ctx, ExperimentReadinessInput{DBPath: dbPath, ProblemID: bare.ProblemID})
	if err != nil {
		t.Fatalf("readiness bare: %v", err)
	}
	if blocked.Ready {
		t.Fatal("a bare problem must not be ready")
	}
	bstatus := map[string]string{}
	for _, c := range blocked.Checks {
		bstatus[c.Check] = c.Status
	}
	for _, check := range []string{"failure_cohort", "decisive_axis_resolution", "surviving_invariants", "withheld_target"} {
		if bstatus[check] != "blocked" {
			t.Fatalf("bare problem: check %s must be blocked with a named reason: %+v", check, blocked.Checks)
		}
	}
}

// TestIntegrationExperimentReadinessVocabParameterized pins the population
// selector fix: a corpus signed under a successor vocabulary revision reports
// its signatures when the matching --vocab-version is supplied, and reports
// the honest signature-less blocker under a version it was NOT signed with.
// Check semantics are identical either way.
func TestIntegrationExperimentReadinessVocabParameterized(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)

	problemID, runID, snapshotID := seedProblemAndSnapshot(t, ctx, dbPath, now)
	seed, err := app.SeedMechanismFixture(ctx, MechanismFixtureSeedInput{
		DBPath: dbPath, ProblemID: problemID, RunID: runID, SnapshotID: snapshotID,
		Fixture: interpFixture("es-v2-only", "V2-only corpus", "residue locality"),
	})
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	if _, err := app.SignatureMechanism(ctx, SignatureInput{DBPath: dbPath, MechanismID: seed.MechanismIDs[0], VocabVersion: canon.VocabularyMechanismV2}); err != nil {
		t.Fatalf("signature: %v", err)
	}

	statusOf := func(vocab string) string {
		r, rerr := app.ExperimentReadiness(ctx, ExperimentReadinessInput{DBPath: dbPath, ProblemID: problemID, VocabVersion: vocab})
		if rerr != nil {
			t.Fatalf("readiness (%q): %v", vocab, rerr)
		}
		for _, c := range r.Checks {
			if c.Check == "decisive_axis_resolution" {
				return c.Status + " " + c.Detail
			}
		}
		t.Fatal("missing decisive_axis_resolution check")
		return ""
	}

	underV2 := statusOf(canon.VocabularyMechanismV2)
	if !strings.HasPrefix(underV2, "ready") {
		t.Fatalf("v2-signed corpus must report its signatures under v2: %s", underV2)
	}
	underDefault := statusOf("")
	if !strings.Contains(underDefault, "no persisted signatures") {
		t.Fatalf("default (v1) must honestly report the corpus is not signed under it: %s", underDefault)
	}
}
