package pipeline

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/provider"
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
