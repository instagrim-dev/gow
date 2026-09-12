// Package pipeline: the typed projection chain (v37, structural review S5
// part B). A frontier proposal is a PROPOSED STRUCTURAL CHANGE; a structural
// description plus directed-generation prose is not a concrete construction.
// ProjectProposal records the authored CONCRETE PLAN (projection artifact),
// derives its VERIFICATION OBLIGATIONS — deciding composition deterministically
// and leaving domain realization OPEN for an external checker — and
// DischargeObligation ties the open obligation to a DOMAIN OBSERVATION (an
// evaluation of the same proposal). A plan whose steps cannot compose is
// refuted at projection time, before any domain work.
package pipeline

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/projection"
	"github.com/instagrim-dev/newf/internal/store"
)

// ProjectProposalInput drives `newf projection propose`.
type ProjectProposalInput struct {
	DBPath     string
	ProposalID string
	// Path is the authored projection artifact (projection/v1 JSON).
	Path string
	// AuthorKind records who authored the plan: operator (default) or tool.
	AuthorKind string
	JSONOutput bool
}

// DischargeObligationInput drives `newf projection discharge`.
type DischargeObligationInput struct {
	DBPath       string
	ObligationID string
	// Status is the operator's verdict: discharged or failed.
	Status string
	// EvaluationID is the domain observation backing the verdict: an
	// evaluation of the SAME proposal the artifact projects. Required.
	EvaluationID string
	// Note is the operator's basis (recorded verbatim). Required.
	Note       string
	JSONOutput bool
}

// ProjectionListInput drives `newf projection list`.
type ProjectionListInput struct {
	DBPath     string
	ProblemID  string
	JSONOutput bool
}

// ProjectionObligationView is one verification obligation with any decision.
type ProjectionObligationView struct {
	ID        string `json:"id"`
	Ordinal   int    `json:"ordinal"`
	Kind      string `json:"kind"`
	Statement string `json:"statement"`
	Checker   string `json:"checker_kind"`
	// Status is open|discharged|failed (derived: open when no decision).
	Status       string `json:"status"`
	DecidedBy    string `json:"decided_by,omitempty"`
	Basis        string `json:"basis,omitempty"`
	EvidenceKind string `json:"evidence_kind,omitempty"`
	EvidenceRef  string `json:"evidence_ref,omitempty"`
	// EvaluationVerdict/EvaluationStrength surface the typed epistemic weight
	// of the backing evaluation for operator decisions (v40, review F6).
	EvaluationVerdict  string `json:"evaluation_verdict,omitempty"`
	EvaluationStrength string `json:"evaluation_strength,omitempty"`
}

// ProjectionView is one artifact revision with its obligations.
type ProjectionView struct {
	ID          string                     `json:"id"`
	ProposalID  string                     `json:"proposal_id"`
	Revision    int                        `json:"revision"`
	AuthorKind  string                     `json:"author_kind"`
	ContentHash string                     `json:"content_hash"`
	CreatedAt   string                     `json:"created_at"`
	Obligations []ProjectionObligationView `json:"obligations"`
}

// ProjectProposalResponse reports one projection pass.
type ProjectProposalResponse struct {
	OK        bool   `json:"ok"`
	Command   string `json:"command"`
	Store     string `json:"store"`
	ProblemID string `json:"problem_id"`
	RunID     string `json:"run_id"`
	// Composes reports the deterministic steps-compose verdict; Gaps carries
	// the exact missing tokens when it fails.
	Composes   bool             `json:"composes"`
	Gaps       []projection.Gap `json:"gaps,omitempty"`
	Projection ProjectionView   `json:"projection"`
}

// DischargeObligationResponse reports one recorded decision.
type DischargeObligationResponse struct {
	OK         bool                     `json:"ok"`
	Command    string                   `json:"command"`
	Store      string                   `json:"store"`
	RunID      string                   `json:"run_id"`
	Obligation ProjectionObligationView `json:"obligation"`
}

// ProjectionListResponse reports a problem's projection chains.
type ProjectionListResponse struct {
	OK          bool             `json:"ok"`
	Command     string           `json:"command"`
	Store       string           `json:"store"`
	ProblemID   string           `json:"problem_id"`
	Projections []ProjectionView `json:"projections"`
}

func obligationView(ob store.ProjectionObligationRow) ProjectionObligationView {
	v := ProjectionObligationView{
		ID: ob.ID, Ordinal: ob.Ordinal, Kind: ob.Kind, Statement: ob.Statement,
		Checker: ob.Checker, Status: "open",
	}
	if ob.Decision != nil {
		v.Status = ob.Decision.Status
		v.DecidedBy = ob.Decision.DecidedBy
		v.Basis = ob.Decision.Basis
		v.EvidenceKind = ob.Decision.EvidenceKind
		v.EvidenceRef = ob.Decision.EvidenceRef
		v.EvaluationVerdict = ob.Decision.EvaluationVerdict
		v.EvaluationStrength = ob.Decision.EvaluationStrength
	}
	return v
}

func projectionView(rec store.ProjectionRecord) ProjectionView {
	v := ProjectionView{
		ID:          rec.Artifact.ID,
		ProposalID:  rec.Artifact.ProposalID,
		Revision:    rec.Artifact.Revision,
		AuthorKind:  rec.Artifact.AuthorKind,
		ContentHash: rec.Artifact.ContentHash,
		CreatedAt:   rec.Artifact.CreatedAt,
	}
	for _, ob := range rec.Obligations {
		v.Obligations = append(v.Obligations, obligationView(ob))
	}
	return v
}

// ProjectProposal records an authored concrete plan for one frontier proposal,
// runs the deterministic composition check, and persists the artifact with its
// typed obligations in one transaction. A composing plan carries a discharged
// steps-compose obligation AND an open domain-realization obligation (the
// honest record that the domain checker is external); a non-composing plan is
// refuted here — its steps-compose obligation is failed with the exact gaps
// and NO domain-realization obligation exists (there is no composed plan to
// realize; a fixed plan is a new revision).
func (a *App) ProjectProposal(ctx context.Context, input ProjectProposalInput) (ProjectProposalResponse, error) {
	author := input.AuthorKind
	if author == "" {
		author = "operator"
	}
	if author != "operator" && author != "tool" {
		return ProjectProposalResponse{}, fmt.Errorf("author kind must be operator or tool, got %q", author)
	}
	raw, err := os.ReadFile(input.Path)
	if err != nil {
		return ProjectProposalResponse{}, fmt.Errorf("read projection artifact: %w", err)
	}
	artifact, err := projection.Parse(raw)
	if err != nil {
		return ProjectProposalResponse{}, err
	}
	comp := projection.CheckComposition(artifact)

	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return ProjectProposalResponse{}, err
	}
	defer repoStore.Close()

	problemID, err := repoStore.GetProposalProblem(ctx, input.ProposalID)
	if err != nil {
		return ProjectProposalResponse{}, err
	}

	now := a.now()
	run, err := repoStore.CreateRun(ctx, domain.NewRun{
		ID:          domain.NewRunID(now),
		ProblemID:   problemID,
		Operation:   "projection propose",
		Status:      domain.RunStatusRunning,
		InputRef:    "frontier_proposal:" + input.ProposalID,
		ToolName:    "newf",
		ToolVersion: a.version,
		StartedAt:   now,
		CompletedAt: now,
	})
	if err != nil {
		return ProjectProposalResponse{}, err
	}

	sum := sha256.Sum256(raw)
	contentHash := hex.EncodeToString(sum[:])
	created := now.Format(timeLayout)
	rec := store.ProjectionRecord{
		Artifact: store.ProjectionArtifactRow{
			ID:            domain.NewProjectionArtifactID(now),
			ProblemID:     problemID,
			ProposalID:    input.ProposalID,
			RunID:         run.ID,
			AuthorKind:    author,
			SchemaVersion: projection.SchemaProjectionV1,
			ContentJSON:   string(raw),
			ContentHash:   contentHash,
			CreatedAt:     created,
		},
	}

	composeStatus := "discharged"
	if !comp.OK {
		composeStatus = "failed"
	}
	rec.Obligations = append(rec.Obligations, store.ProjectionObligationRow{
		ID:        domain.NewProjectionObligationID(now),
		Ordinal:   0,
		Kind:      "steps-compose",
		Statement: "every step's requirements are met by givens or strictly earlier provides, and every target token is provided",
		Checker:   "deterministic-check",
		CreatedAt: created,
		Decision: &store.ProjectionObligationDecisionRow{
			RunID:        run.ID,
			Status:       composeStatus,
			DecidedBy:    "code",
			Basis:        comp.Detail(),
			EvidenceKind: "code-check",
			EvidenceRef:  contentHash,
			CreatedAt:    created,
		},
	})
	if comp.OK {
		rec.Obligations = append(rec.Obligations, store.ProjectionObligationRow{
			ID:        domain.NewProjectionObligationID(now),
			Ordinal:   1,
			Kind:      "domain-realization",
			Statement: "the composed plan is realizable in the domain and its outcome is observed by an external checker",
			Checker:   "external",
			CreatedAt: created,
		})
	}

	persisted, err := repoStore.PersistProjection(ctx, rec)
	if err != nil {
		a.failRun(ctx, repoStore, run.ID, err)
		return ProjectProposalResponse{}, err
	}
	if err := a.finalizeRun(ctx, repoStore, run.ID, nil); err != nil {
		return ProjectProposalResponse{}, err
	}

	return ProjectProposalResponse{
		OK:         true,
		Command:    "projection propose",
		Store:      dbPath,
		ProblemID:  problemID,
		RunID:      run.ID,
		Composes:   comp.OK,
		Gaps:       comp.Gaps,
		Projection: projectionView(persisted),
	}, nil
}

// DischargeObligation records the operator's verdict on one OPEN external
// obligation, backed by a domain observation: an evaluation of the SAME
// proposal the artifact projects. Code-owned obligations (steps-compose) are
// refused — their verdicts come from the deterministic checker at projection
// time. Decisions are append-once; re-deciding is a loud refusal.
func (a *App) DischargeObligation(ctx context.Context, input DischargeObligationInput) (DischargeObligationResponse, error) {
	if input.Status != "discharged" && input.Status != "failed" {
		return DischargeObligationResponse{}, fmt.Errorf("status must be discharged or failed, got %q", input.Status)
	}
	if strings.TrimSpace(input.Note) == "" {
		return DischargeObligationResponse{}, fmt.Errorf("--note is required: an obligation decision records its basis")
	}
	if input.EvaluationID == "" {
		return DischargeObligationResponse{}, fmt.Errorf("--evaluation is required: a domain-realization verdict is backed by a domain observation, not an assertion")
	}

	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return DischargeObligationResponse{}, err
	}
	defer repoStore.Close()

	ob, art, err := repoStore.GetProjectionObligation(ctx, input.ObligationID)
	if err != nil {
		return DischargeObligationResponse{}, err
	}
	if ob.Kind == "steps-compose" {
		return DischargeObligationResponse{}, fmt.Errorf("obligation %s is code-owned (steps-compose): its verdict comes from the deterministic checker at projection time", ob.ID)
	}
	if ob.Decision != nil {
		return DischargeObligationResponse{}, fmt.Errorf("obligation %s is already decided (%s); decisions are append-once", ob.ID, ob.Decision.Status)
	}
	ev, err := repoStore.GetEvaluation(ctx, input.EvaluationID)
	if err != nil {
		return DischargeObligationResponse{}, err
	}
	if ev.ProposalID != art.ProposalID {
		return DischargeObligationResponse{}, fmt.Errorf("evaluation %s assesses proposal %s, not the projected proposal %s: a domain observation must be about the same proposed change", ev.ID, ev.ProposalID, art.ProposalID)
	}

	now := a.now()
	run, err := repoStore.CreateRun(ctx, domain.NewRun{
		ID:          domain.NewRunID(now),
		ProblemID:   art.ProblemID,
		Operation:   "projection discharge",
		Status:      domain.RunStatusRunning,
		InputRef:    "projection_obligation:" + ob.ID,
		ToolName:    "newf",
		ToolVersion: a.version,
		StartedAt:   now,
		CompletedAt: now,
	})
	if err != nil {
		return DischargeObligationResponse{}, err
	}

	decision := store.ProjectionObligationDecisionRow{
		ObligationID: ob.ID,
		RunID:        run.ID,
		Status:       input.Status,
		DecidedBy:    "operator",
		Basis:        "operator: " + strings.TrimSpace(input.Note) + " (evaluation verdict " + ev.Verdict + ", " + ev.VerificationStrength + ")",
		EvidenceKind: "evaluation",
		EvidenceRef:  ev.ID,
		// Typed copy of the backing observation's epistemic weight (F6): a
		// policy consumer must not have to parse the prose basis to learn
		// whether the discharge rests on a reproducible computation or a
		// single model judgment.
		EvaluationVerdict:  ev.Verdict,
		EvaluationStrength: ev.VerificationStrength,
		CreatedAt:          now.Format(timeLayout),
	}
	if err := repoStore.PersistObligationDecision(ctx, decision); err != nil {
		a.failRun(ctx, repoStore, run.ID, err)
		return DischargeObligationResponse{}, err
	}
	if err := a.finalizeRun(ctx, repoStore, run.ID, nil); err != nil {
		return DischargeObligationResponse{}, err
	}
	ob.Decision = &decision
	return DischargeObligationResponse{
		OK: true, Command: "projection discharge", Store: dbPath, RunID: run.ID,
		Obligation: obligationView(ob),
	}, nil
}

// ListProjections returns a problem's projection chains.
func (a *App) ListProjections(ctx context.Context, input ProjectionListInput) (ProjectionListResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return ProjectionListResponse{}, err
	}
	defer repoStore.Close()
	if _, err := repoStore.GetProblem(ctx, input.ProblemID); err != nil {
		return ProjectionListResponse{}, err
	}
	recs, err := repoStore.ListProjectionsForProblem(ctx, input.ProblemID)
	if err != nil {
		return ProjectionListResponse{}, err
	}
	resp := ProjectionListResponse{OK: true, Command: "projection list", Store: dbPath, ProblemID: input.ProblemID}
	for _, rec := range recs {
		resp.Projections = append(resp.Projections, projectionView(rec))
	}
	return resp, nil
}
