package pipeline

import (
	"context"
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
}
