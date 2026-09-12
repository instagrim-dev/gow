package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/provider"
	"github.com/instagrim-dev/newf/internal/verify"
)

func writeArtifactFile(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}
	return path
}

const composingArtifact = `{
  "schema": "projection/v1",
  "givens": ["congruence-lattice"],
  "target": ["crossing-bound"],
  "steps": [
    {"name": "lift", "statement": "lift the system to its covering congruence classes",
     "requires": ["congruence-lattice"], "provides": ["cover"]},
    {"name": "bound", "statement": "bound crossings uniformly over the cover",
     "requires": ["cover"], "provides": ["crossing-bound"]}
  ]
}`

// Step "glue" requires a uniformity token NOTHING provides, and a descent
// datum only a LATER step provides: the abstract path cannot compose.
const nonComposingArtifact = `{
  "schema": "projection/v1",
  "target": ["final-bound"],
  "steps": [
    {"name": "lift", "statement": "lift", "provides": ["cover"]},
    {"name": "glue", "statement": "glue local bounds",
     "requires": ["cover", "uniformity", "descent-datum"]},
    {"name": "descend", "statement": "descend to the base", "provides": ["descent-datum"]}
  ]
}`

// TestIntegrationProjectionChainFourRecords is the S5 part-B regression
// (issue #20): one path through all four separated records, ending in a
// domain observation —
//
//	frontier proposal (proposed structural change)
//	-> projection artifact (authored concrete plan; steps-compose DISCHARGED
//	   by the deterministic checker)
//	-> open domain-realization obligation (the missing domain checker is a
//	   recorded OPEN obligation, never a silent assumption)
//	-> operator discharge backed by an evaluation of the SAME proposal (the
//	   domain observation), here an honest FAILED verdict.
//
// Guards: code-owned obligations refuse manual discharge; decisions are
// append-once; the backing evaluation must assess the projected proposal.
func TestIntegrationProjectionChainFourRecords(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = dataDrivenMiner{}
	app.challengerFn = biasOnlyChallenger{}
	app.modelVerifierFn = provider.NewFixtureModelVerifier(verify.VerdictFailure, "high")

	// Record 1: the proposed structural change.
	problemID, invID, _ := mineOneCandidate(t, ctx, app, dbPath)
	if _, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID}); err != nil {
		t.Fatalf("challenge: %v", err)
	}
	gen, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil || len(gen.Generation.Proposals) == 0 {
		t.Fatalf("frontier generate: %v (proposals=%d)", err, len(gen.Generation.Proposals))
	}
	proposalID := gen.Generation.Proposals[0].ID

	// Record 2: the authored concrete plan. Composition is decided by code.
	proj, err := app.ProjectProposal(ctx, ProjectProposalInput{
		DBPath: dbPath, ProposalID: proposalID,
		Path: writeArtifactFile(t, "plan.json", composingArtifact),
	})
	if err != nil {
		t.Fatalf("projection propose: %v", err)
	}
	if !proj.Composes || len(proj.Gaps) != 0 {
		t.Fatalf("composing plan must compose: %+v", proj)
	}
	if proj.Projection.ProposalID != proposalID || proj.Projection.Revision != 1 {
		t.Fatalf("artifact must link the proposal: %+v", proj.Projection)
	}

	// Record 3: the typed obligations — compose discharged by code, domain
	// realization OPEN.
	if len(proj.Projection.Obligations) != 2 {
		t.Fatalf("want 2 obligations, got %+v", proj.Projection.Obligations)
	}
	compose, realize := proj.Projection.Obligations[0], proj.Projection.Obligations[1]
	if compose.Kind != "steps-compose" || compose.Status != "discharged" || compose.DecidedBy != "code" ||
		compose.EvidenceKind != "code-check" || compose.EvidenceRef != proj.Projection.ContentHash {
		t.Fatalf("compose obligation must be code-discharged against the exact content: %+v", compose)
	}
	if realize.Kind != "domain-realization" || realize.Status != "open" || realize.Checker != "external" {
		t.Fatalf("domain realization must be a recorded OPEN obligation: %+v", realize)
	}

	// A code-owned obligation refuses manual discharge.
	if _, err := app.DischargeObligation(ctx, DischargeObligationInput{
		DBPath: dbPath, ObligationID: compose.ID, Status: "failed", EvaluationID: "evl_x", Note: "n",
	}); err == nil || !strings.Contains(err.Error(), "code-owned") {
		t.Fatalf("manual discharge of steps-compose must be refused, got %v", err)
	}

	// Record 4: the domain observation — an evaluation of the SAME proposal.
	res, err := app.Evaluate(ctx, EvaluateInput{DBPath: dbPath, ProblemID: problemID, ProposalID: proposalID})
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	evalID := res.Run.Evaluations[0].ID

	dis, err := app.DischargeObligation(ctx, DischargeObligationInput{
		DBPath: dbPath, ObligationID: realize.ID, Status: "failed", EvaluationID: evalID,
		Note: "single-model verdict: the plan's construction was judged not to hold",
	})
	if err != nil {
		t.Fatalf("projection discharge: %v", err)
	}
	ob := dis.Obligation
	if ob.Status != "failed" || ob.DecidedBy != "operator" ||
		ob.EvidenceKind != "evaluation" || ob.EvidenceRef != evalID {
		t.Fatalf("decision must link the domain observation: %+v", ob)
	}
	if !strings.Contains(ob.Basis, "single-model-judgment") {
		t.Fatalf("basis must carry the observation's strength label: %q", ob.Basis)
	}

	// Append-once: the decided obligation refuses a second verdict.
	if _, err := app.DischargeObligation(ctx, DischargeObligationInput{
		DBPath: dbPath, ObligationID: realize.ID, Status: "discharged", EvaluationID: evalID, Note: "again",
	}); err == nil || !strings.Contains(err.Error(), "already decided") {
		t.Fatalf("re-deciding must be refused, got %v", err)
	}

	// The whole chain is inspectable.
	list, err := app.ListProjections(ctx, ProjectionListInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil || len(list.Projections) != 1 {
		t.Fatalf("list projections: %v (%d)", err, len(list.Projections))
	}
	if list.Projections[0].Obligations[1].Status != "failed" {
		t.Fatalf("listed chain must reflect the decision: %+v", list.Projections[0].Obligations)
	}
}

// TestIntegrationProjectionCompositionFailure is the reviewer's demanded
// failure case: a proposed abstract path fails BECAUSE its concrete steps
// cannot compose. The refutation is deterministic, names the exact missing
// tokens, happens before any domain work, and leaves NO domain-realization
// obligation (there is no composed plan to realize — a fixed plan is a new
// revision).
func TestIntegrationProjectionCompositionFailure(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)
	app, dbPath := newRealStoreApp(t, now)
	app.invariantMinerFn = dataDrivenMiner{}
	app.challengerFn = biasOnlyChallenger{}

	problemID, invID, _ := mineOneCandidate(t, ctx, app, dbPath)
	if _, err := app.ChallengeInvariants(ctx, ChallengeInput{DBPath: dbPath, InvariantID: invID}); err != nil {
		t.Fatalf("challenge: %v", err)
	}
	gen, err := app.GenerateFrontier(ctx, FrontierGenerateInput{DBPath: dbPath, ProblemID: problemID})
	if err != nil || len(gen.Generation.Proposals) == 0 {
		t.Fatalf("frontier generate: %v", err)
	}
	proposalID := gen.Generation.Proposals[0].ID

	proj, err := app.ProjectProposal(ctx, ProjectProposalInput{
		DBPath: dbPath, ProposalID: proposalID,
		Path: writeArtifactFile(t, "bad-plan.json", nonComposingArtifact),
	})
	if err != nil {
		t.Fatalf("projection propose: %v", err)
	}
	if proj.Composes {
		t.Fatalf("plan must not compose: %+v", proj)
	}
	// The exact gaps are typed: step "glue" is missing its uniformity token AND
	// the descent datum only a LATER step provides (order matters); the target
	// is never provided.
	if len(proj.Gaps) != 2 {
		t.Fatalf("want 2 gaps, got %+v", proj.Gaps)
	}
	g := proj.Gaps[0]
	if g.StepName != "glue" || len(g.Missing) != 2 || g.Missing[0] != "descent-datum" || g.Missing[1] != "uniformity" {
		t.Fatalf("gap must name the failing step and sorted missing tokens: %+v", g)
	}
	if proj.Gaps[1].StepName != "target" || proj.Gaps[1].Missing[0] != "final-bound" {
		t.Fatalf("unmet target must be reported: %+v", proj.Gaps[1])
	}

	// The refutation is a persisted, typed record: steps-compose FAILED by
	// code, and no domain-realization obligation exists.
	if len(proj.Projection.Obligations) != 1 {
		t.Fatalf("non-composing plan must carry only the failed compose obligation: %+v", proj.Projection.Obligations)
	}
	compose := proj.Projection.Obligations[0]
	if compose.Kind != "steps-compose" || compose.Status != "failed" || compose.DecidedBy != "code" ||
		!strings.Contains(compose.Basis, "glue missing [descent-datum, uniformity]") {
		t.Fatalf("failed compose obligation must persist the exact gaps: %+v", compose)
	}

	// Identical re-submission is refused; a FIXED plan is a new revision.
	if _, err := app.ProjectProposal(ctx, ProjectProposalInput{
		DBPath: dbPath, ProposalID: proposalID,
		Path: writeArtifactFile(t, "bad-plan-again.json", nonComposingArtifact),
	}); err == nil || !strings.Contains(err.Error(), "identical projection artifact") {
		t.Fatalf("duplicate content must be refused, got %v", err)
	}
	fixed, err := app.ProjectProposal(ctx, ProjectProposalInput{
		DBPath: dbPath, ProposalID: proposalID,
		Path: writeArtifactFile(t, "fixed-plan.json", composingArtifact),
	})
	if err != nil {
		t.Fatalf("fixed plan: %v", err)
	}
	if !fixed.Composes || fixed.Projection.Revision != 2 {
		t.Fatalf("fixed plan must compose as revision 2: %+v", fixed.Projection)
	}
}
