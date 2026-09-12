package pipeline

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/provider"
	"github.com/instagrim-dev/newf/internal/verify"
)

const v2ComposingArtifact = `{
  "schema": "projection/v2",
  "givens": ["congruence-lattice"],
  "target": ["crossing-bound"],
  "steps": [
    {"name": "lift", "statement": "lift the system to its covering congruence classes",
     "requires": ["congruence-lattice"], "provides": ["cover"]},
    {"name": "bound", "statement": "bound crossings uniformly over the cover",
     "requires": ["cover"], "provides": ["crossing-bound"]}
  ],
  "source_domain": "unit-fraction identities over Z",
  "target_domain": "congruence covers of the moduli",
  "preserves": ["solution existence per residue class"],
  "loses": ["constructive witness values"],
  "correspondence": "one-way-implication",
  "grounding_plan": "each residue-class bound maps back to concrete n with a checkable witness obligation (#23)"
}`

// TestIntegrationProjectionV2SemanticPreservation is the cemented D3
// acceptance path (issue #22): a projection/v2 artifact persists with its
// semantic contract, owes a THIRD obligation of kind semantic-preservation
// (external checker, v42 vocabulary), and that obligation refuses manual
// discharge without backing evidence — then discharges against a witness-
// backed domain-goal evaluation (the D2/D3 seam: grounding instruments are
// witness claims persisted per the #23 carrier decision).
func TestIntegrationProjectionV2SemanticPreservation(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = dataDrivenMiner{}
	app.challengerFn = biasOnlyChallenger{}
	app.modelVerifierFn = provider.NewFixtureModelVerifier(verify.VerdictFailure, "high")

	problemID, invID, _ := mineOneCandidate(t, ctx, app, dbPath)
	if _, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID}); err != nil {
		t.Fatalf("challenge: %v", err)
	}
	gen, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil || len(gen.Generation.Proposals) == 0 {
		t.Fatalf("frontier generate: %v (proposals=%d)", err, len(gen.Generation.Proposals))
	}
	proposalID := gen.Generation.Proposals[0].ID

	proj, err := app.ProjectProposal(ctx, ProjectProposalInput{
		DBPath: dbPath, ProposalID: proposalID,
		Path: writeArtifactFile(t, "plan-v2.json", v2ComposingArtifact),
	})
	if err != nil {
		t.Fatalf("projection propose v2: %v", err)
	}
	if !proj.Composes {
		t.Fatalf("v2 composition must be v1-identical: %+v", proj)
	}
	if proj.Projection.SchemaVersion != "projection/v2" {
		t.Fatalf("artifact must persist its declared schema: %+v", proj.Projection)
	}

	// Three obligations: compose (code-discharged), domain-realization (open),
	// semantic-preservation (open, external, v42 vocabulary).
	if len(proj.Projection.Obligations) != 3 {
		t.Fatalf("v2 artifact owes 3 obligations, got %+v", proj.Projection.Obligations)
	}
	sem := proj.Projection.Obligations[2]
	if sem.Kind != "semantic-preservation" || sem.Status != "open" || sem.Checker != "external" {
		t.Fatalf("semantic-preservation must be a recorded OPEN external obligation: %+v", sem)
	}
	if !strings.Contains(sem.Statement, "solution existence per residue class") ||
		!strings.Contains(sem.Statement, "one-way-implication") {
		t.Fatalf("obligation statement must carry the claimed contract: %q", sem.Statement)
	}

	// Acceptance: refuses manual discharge without backing evidence.
	if _, err := app.DischargeObligation(ctx, DischargeObligationInput{
		DBPath: dbPath, ObligationID: sem.ID, Status: "discharged", Note: "it just does",
	}); err == nil || !strings.Contains(err.Error(), "--evaluation is required") {
		t.Fatalf("manual discharge without evidence must be refused, got %v", err)
	}

	// The backing observation: a witness-valid check on the same proposal
	// (subject domain-goal, strength reproducible) — the #23 carrier.
	wit, err := app.WitnessCheck(ctx, WitnessCheckInput{
		DBPath: dbPath, ProposalID: proposalID, Tuple: "2,1,2,2",
		Note: "grounding-plan instrument (test fixture)",
	})
	if err != nil || wit.WitnessVerdict != "witness-valid" {
		t.Fatalf("witness check: %v %+v", err, wit)
	}

	dis, err := app.DischargeObligation(ctx, DischargeObligationInput{
		DBPath: dbPath, ObligationID: sem.ID, Status: "discharged", EvaluationID: wit.Evaluation.ID,
		Note: "witness-backed grounding: the residue-class bound produced a checkable valid witness",
	})
	if err != nil {
		t.Fatalf("semantic-preservation discharge: %v", err)
	}
	ob := dis.Obligation
	if ob.Status != "discharged" || ob.EvidenceKind != "evaluation" || ob.EvidenceRef != wit.Evaluation.ID ||
		ob.EvaluationStrength != string(verify.StrengthReproducible) {
		t.Fatalf("discharge must link the witness observation with typed strength: %+v", ob)
	}

	// A v1 artifact on the same problem still owes only two obligations.
	gen2 := gen.Generation.Proposals
	_ = gen2
	projV1, err := app.ProjectProposal(ctx, ProjectProposalInput{
		DBPath: dbPath, ProposalID: proposalID,
		Path: writeArtifactFile(t, "plan-v1.json", composingArtifact),
	})
	if err != nil {
		t.Fatalf("projection propose v1: %v", err)
	}
	if len(projV1.Projection.Obligations) != 2 || projV1.Projection.SchemaVersion != "projection/v1" {
		t.Fatalf("v1 artifact must keep its two-obligation shape: %+v", projV1.Projection)
	}
}
