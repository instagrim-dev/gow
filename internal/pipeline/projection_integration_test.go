package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/provider"
	"github.com/instagrim-dev/newf/internal/store"
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
	// F6 (v40): the backing observation's verdict and strength are TYPED on
	// the decision, not only embedded in prose — a policy consumer can weigh
	// the discharge without parsing the basis.
	if ob.EvaluationVerdict == "" || ob.EvaluationStrength != "single-model-judgment" {
		t.Fatalf("decision must carry typed evaluation verdict/strength: %+v", ob)
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

// TestIntegrationWitnessBackedProjectionDischarge closes issue #23's last
// acceptance path: a domain-realization projection obligation discharged on
// the strength of an EXACT witness check — an external-checker-backed domain
// observation — with two refusal gates proven on the way:
//
//  1. verdict–status coherence: `discharged` (realization holds) over a
//     failure-verdict observation is refused — recording the stronger status
//     over a weaker verdict is a silent epistemic promotion;
//  2. subject gate (v38): an annotation-subject evaluation cannot back a
//     domain-realization decision — a certificate about the DESCRIPTION is
//     not an observation of the DOMAIN, which is the exact wrong-object
//     confusion the subject axis exists to prevent.
func TestIntegrationWitnessBackedProjectionDischarge(t *testing.T) {
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
		t.Fatalf("frontier generate: %v", err)
	}
	proposalID := gen.Generation.Proposals[0].ID

	proj, err := app.ProjectProposal(ctx, ProjectProposalInput{
		DBPath: dbPath, ProposalID: proposalID,
		Path: writeArtifactFile(t, "plan.json", composingArtifact),
	})
	if err != nil || !proj.Composes {
		t.Fatalf("projection propose: %v (composes=%v)", err, proj.Composes)
	}
	var realizeID string
	for _, ob := range proj.Projection.Obligations {
		if ob.Kind == "domain-realization" {
			realizeID = ob.ID
		}
	}
	if realizeID == "" {
		t.Fatal("composing plan must open a domain-realization obligation")
	}

	// The domain observation: the plan's step-3 output tuple is checked
	// exactly and refuted (the issue's near-miss vector). This is a
	// reproducible-computation, domain-goal evaluation of the SAME proposal.
	wres, err := app.WitnessCheck(ctx, WitnessCheckInput{
		DBPath: dbPath, ProposalID: proposalID, Tuple: "5,2,4,21",
		Note: "tuple produced by the projected plan's construction (test fixture)",
	})
	if err != nil {
		t.Fatalf("witness check: %v", err)
	}
	if wres.WitnessVerdict != "witness-invalid" {
		t.Fatalf("near-miss must be refuted: %+v", wres)
	}
	witnessEvalID := wres.Evaluation.ID

	// Gate 1: the observation is a domain FAILURE; recording the obligation
	// as discharged (realization holds) anyway must be refused.
	if _, err := app.DischargeObligation(ctx, DischargeObligationInput{
		DBPath: dbPath, ObligationID: realizeID, Status: "discharged",
		EvaluationID: witnessEvalID, Note: "wishful",
	}); err == nil || !strings.Contains(err.Error(), "unambiguous domain success") {
		t.Fatalf("discharged-over-failure must be refused as an epistemic promotion: %v", err)
	}

	// Gate 2: an annotation-subject evaluation (a certificate about the
	// signature, not the domain) cannot back a domain-realization decision.
	repo := openTestStore(t, ctx, dbPath)
	annRun, err := repo.CreateRun(ctx, domain.NewRun{
		ID: domain.NewRunID(now), ProblemID: problemID, Operation: "test",
		Status: domain.RunStatusRunning, InputRef: "test", ToolName: "t", ToolVersion: "v",
		StartedAt: now, CompletedAt: now,
	})
	if err != nil {
		t.Fatalf("create run: %v", err)
	}
	annEvalID := domain.NewEvaluationID(now.Add(time.Minute))
	if _, err := repo.PersistEvaluationRun(ctx, store.EvaluationRunRecord{
		ID: domain.NewEvaluationRunID(now.Add(time.Minute)), ProblemID: problemID,
		RunID: annRun.ID, Mode: "proposal", CreatedAt: now.Format("2006-01-02T15:04:05Z"),
		Evaluations: []store.EvaluationRow{{
			ID: annEvalID, ProposalID: proposalID, Verdict: "failure",
			VerifierKind: "deterministic-check", VerificationStrength: "deterministic",
			VerificationSubject: "annotation", ToolName: "t", ToolVersion: "v",
			Notes: "annotation-subject certificate (test)",
		}},
	}); err != nil {
		t.Fatalf("persist annotation evaluation: %v", err)
	}
	repo.Close()
	if _, err := app.DischargeObligation(ctx, DischargeObligationInput{
		DBPath: dbPath, ObligationID: realizeID, Status: "failed",
		EvaluationID: annEvalID, Note: "wrong object",
	}); err == nil || !strings.Contains(err.Error(), "domain-goal observation") {
		t.Fatalf("annotation-subject backing must be refused: %v", err)
	}

	// The coherent decision: failed, backed by the witness observation.
	dis, err := app.DischargeObligation(ctx, DischargeObligationInput{
		DBPath: dbPath, ObligationID: realizeID, Status: "failed",
		EvaluationID: witnessEvalID,
		Note:         "the plan's construction produced a tuple the exact checker refutes",
	})
	if err != nil {
		t.Fatalf("witness-backed discharge: %v", err)
	}
	ob := dis.Obligation
	if ob.Status != "failed" || ob.EvidenceRef != witnessEvalID {
		t.Fatalf("decision must link the witness observation: %+v", ob)
	}
	if ob.EvaluationVerdict != "failure" || ob.EvaluationStrength != "reproducible" {
		t.Fatalf("typed epistemic weight must record the reproducible witness check: %+v", ob)
	}
	if !strings.Contains(ob.Basis, "reproducible") {
		t.Fatalf("basis must carry the observation's strength: %q", ob.Basis)
	}
}
