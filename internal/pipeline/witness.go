// Package pipeline: witness-backed evaluations (issue #23 slice 2, decision
// D2-C recorded on #23).
//
// `newf witness check` is the carrier that brings a domain witness verdict
// into the evaluation record: an operator- or tool-produced concrete outcome
// tuple for a proposal's attempted mechanism is checked by the exact-integer
// witness checker, and the verdict is persisted as a reproducible-computation
// EVALUATION (subject domain-goal, strength reproducible) through the same
// transactional choke point every other evaluation uses. Downstream, that
// evaluation is what rule-admits a decisive witness-invalid outcome as a
// domain-checked failure (admission) and what discharges a domain-realization
// obligation with typed strength (projection discharge --evaluation).
//
// The proposal wire is deliberately NOT extended: witnesses are checkable
// domain claims authored outside the untrusted provider channel (D2 rejected
// option A). Models own nothing in this path.
package pipeline

import (
	"context"
	"fmt"

	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/store"
	"github.com/instagrim-dev/newf/internal/verify"
	"github.com/instagrim-dev/newf/internal/witness"
)

// WitnessCheckInput drives `newf witness check`.
type WitnessCheckInput struct {
	DBPath string
	// ProposalID is the frontier proposal whose attempted mechanism produced
	// the tuple. The verdict lands on this proposal's evaluation history.
	ProposalID string
	// Tuple is the concrete produced outcome "n,x,y,z" (decimal, arbitrary
	// precision), checked exactly.
	Tuple string
	// Note records where the tuple came from (required: a witness claim
	// without provenance is just a number).
	Note       string
	JSONOutput bool
}

// WitnessCheckResponse reports the persisted witness-backed evaluation.
type WitnessCheckResponse struct {
	OK      bool   `json:"ok"`
	Command string `json:"command"`
	Store   string `json:"store"`
	// WitnessVerdict is the checker's own verdict: witness-valid | witness-invalid.
	WitnessVerdict string `json:"witness_verdict"`
	// CanonicalClaim is the deterministic serialization sufficient to re-run
	// the check.
	CanonicalClaim  string         `json:"canonical_claim"`
	Detail          string         `json:"detail,omitempty"`
	EvaluationRunID string         `json:"evaluation_run_id"`
	Evaluation      EvaluationView `json:"evaluation"`
}

// WitnessCheck parses and exactly checks a witness tuple for a proposal, then
// persists the verdict as a one-evaluation run:
//
//	witness-valid   -> verdict success
//	witness-invalid -> verdict failure (re-enters as an evaluated_failures
//	                   marker in the SAME transaction, admissible by rule as
//	                   a domain-checked failure)
//
// A malformed tuple is an input error, never a domain verdict: nothing is
// persisted.
func (a *App) WitnessCheck(ctx context.Context, input WitnessCheckInput) (WitnessCheckResponse, error) {
	if input.ProposalID == "" {
		return WitnessCheckResponse{}, fmt.Errorf("proposal id is required")
	}
	if input.Note == "" {
		return WitnessCheckResponse{}, fmt.Errorf("a note recording the tuple's provenance is required: a witness claim without provenance is just a number")
	}
	claim, err := witness.ParseErdosStraus(input.Tuple)
	if err != nil {
		return WitnessCheckResponse{}, err
	}

	// The check itself: pure, exact, before any store write.
	checkErr := claim.Check()
	witnessVerdict := "witness-valid"
	evalVerdict := string(verify.VerdictSuccess)
	detail := ""
	if checkErr != nil {
		witnessVerdict = "witness-invalid"
		evalVerdict = string(verify.VerdictFailure)
		detail = checkErr.Error()
	}

	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return WitnessCheckResponse{}, err
	}
	defer repoStore.Close()

	problemID, err := repoStore.GetProposalProblem(ctx, input.ProposalID)
	if err != nil {
		return WitnessCheckResponse{}, err
	}

	now := a.now()
	run, err := repoStore.CreateRun(ctx, domain.NewRun{
		ID:          domain.NewRunID(now),
		ProblemID:   problemID,
		Operation:   "witness check",
		Status:      domain.RunStatusRunning,
		InputRef:    "frontier_proposal:" + input.ProposalID,
		ToolName:    "newf",
		ToolVersion: a.version,
		StartedAt:   now,
		CompletedAt: now,
	})
	if err != nil {
		return WitnessCheckResponse{}, err
	}

	// Bind the verdict to the proposal's newest persisted signature revision
	// at assessment time — the same revision-pinning rule the evaluation
	// stage follows. Admission materializes exactly these bytes.
	contentHash := ""
	if content, found, gerr := repoStore.LatestProposalSignatureContent(ctx, input.ProposalID); gerr != nil {
		a.failRun(ctx, repoStore, run.ID, gerr)
		return WitnessCheckResponse{}, gerr
	} else if found {
		contentHash = content.ContentHash
	}

	notes := "witness " + witnessVerdict + " " + claim.Canonical()
	if detail != "" {
		notes += ": " + detail
	}
	notes += " — provenance: " + input.Note

	record := store.EvaluationRunRecord{
		ID:        domain.NewEvaluationRunID(now),
		ProblemID: problemID,
		RunID:     run.ID,
		Mode:      "proposal",
		CreatedAt: now.Format(timeLayout),
		Evaluations: []store.EvaluationRow{{
			ID:                   domain.NewEvaluationID(now),
			ProposalID:           input.ProposalID,
			Verdict:              evalVerdict,
			VerifierKind:         string(verify.KindReproducibleComputation),
			VerificationStrength: string(verify.StrengthReproducible),
			VerificationSubject:  string(verify.SubjectDomainGoal),
			ToolName:             witness.CheckerName,
			ToolVersion:          witness.CheckerVersion,
			Notes:                notes,
			SignatureContentHash: contentHash,
		}},
	}
	persisted, err := repoStore.PersistEvaluationRun(ctx, record)
	if err != nil {
		a.failRun(ctx, repoStore, run.ID, err)
		return WitnessCheckResponse{}, err
	}
	if err := a.finalizeRun(ctx, repoStore, run.ID, nil); err != nil {
		return WitnessCheckResponse{}, err
	}

	ev := persisted.Evaluations[0]
	return WitnessCheckResponse{
		OK:              true,
		Command:         "witness check",
		Store:           dbPath,
		WitnessVerdict:  witnessVerdict,
		CanonicalClaim:  claim.Canonical(),
		Detail:          detail,
		EvaluationRunID: persisted.ID,
		Evaluation: EvaluationView{
			ID:                   ev.ID,
			ProposalID:           ev.ProposalID,
			Verdict:              ev.Verdict,
			VerifierKind:         ev.VerifierKind,
			VerificationStrength: ev.VerificationStrength,
			VerificationSubject:  ev.VerificationSubject,
			ToolName:             ev.ToolName,
			ToolVersion:          ev.ToolVersion,
			Notes:                ev.Notes,
		},
	}, nil
}
