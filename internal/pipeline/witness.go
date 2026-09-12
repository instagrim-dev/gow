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
// The verdict is an assessment of ONE OCCURRENCE: the generation whose emitted
// interpretation produced the tuple, that generation's population, and the
// exact content revision it bound. That context is selected and pinned before
// any write, so the contextual current-result reader can attribute the result
// to the occurrence (F2, 2026-09-12 review-flow run). An unattributable request
// is refused rather than persisted as an orphan verdict.
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
	// GenerationID pins the OCCURRENCE the witness verdict is about (F2 of the
	// 2026-09-12 review-flow run). Empty selects the proposal's latest
	// occurrence generation — the same default the evaluation stage uses — so
	// a historical occurrence is replayable by pinning it explicitly.
	GenerationID string
	// Tuple is the concrete produced outcome "n,x,y,z" (decimal, arbitrary
	// precision), checked exactly. Mutually exclusive with Procedure: a
	// supplied tuple is an operator-provenance claim with a recorded
	// attribution gap.
	Tuple string
	// Procedure names a registered deterministic bounded-attempt procedure
	// (witness.AttemptProcedures). When set, the tuple is not supplied — it is
	// COMPUTED here from Params, and the attempt→output binding (procedure,
	// executor version, canonical params, canonical tuple) is persisted in the
	// same transaction as the evaluation, giving admission a checkable link
	// instead of a declared fixture relationship (attribution slice,
	// 2026-09-12 review handoff 3). A procedure that abstains for its params
	// is an input-level refusal: no tuple, no claim, nothing persisted.
	Procedure string
	// Params are the declared parameters of the bounded attempt (decimal
	// integers; e.g. n, x0_offset). Recorded canonically and re-parsed
	// verbatim on recheck.
	Params map[string]string
	// Note records where the tuple came from (required for supplied tuples: a
	// witness claim without provenance is just a number; optional for executed
	// attempts, whose provenance IS the recorded binding).
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
	CanonicalClaim string `json:"canonical_claim"`
	Detail         string `json:"detail,omitempty"`
	// FrontierGenerationRunID / ClusterRunID are the pinned OCCURRENCE context
	// the verdict is about, written onto the evaluation run so the current-result
	// reader can attribute this assessment to that occurrence (F2).
	FrontierGenerationRunID string `json:"frontier_generation_run_id"`
	ClusterRunID            string `json:"cluster_run_id,omitempty"`
	// OccurrenceContentHash is the exact content revision the pinned occurrence
	// bound — the bytes this verdict is about.
	OccurrenceContentHash string `json:"occurrence_content_hash,omitempty"`
	// OccurrencePinned reports whether an actual occurrence binding backed the
	// selected content. False means pre-v24 history without a binding: the
	// verdict is retained, but it cannot be attributed to an occurrence and the
	// contextual current-result reader will not surface it. Never faked.
	OccurrencePinned bool           `json:"occurrence_pinned"`
	EvaluationRunID  string         `json:"evaluation_run_id"`
	Evaluation       EvaluationView `json:"evaluation"`
	// AttemptBinding is the persisted attempt→output binding when the tuple
	// was computed by a registered procedure (nil for supplied tuples: the
	// attribution gap is recorded, not faked).
	AttemptBinding *WitnessAttemptBindingView `json:"attempt_binding,omitempty"`
}

// WitnessAttemptBindingView is the machine-readable attempt→output binding.
type WitnessAttemptBindingView struct {
	Procedure       string `json:"procedure"`
	ExecutorVersion string `json:"executor_version"`
	ParamsCanonical string `json:"params_canonical"`
	TupleCanonical  string `json:"tuple_canonical"`
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
	var claim witness.ErdosStrausClaim
	var binding *store.WitnessAttemptBindingRow
	switch {
	case input.Procedure != "" && input.Tuple != "":
		return WitnessCheckResponse{}, fmt.Errorf("--procedure and --tuple are mutually exclusive: an executed attempt computes its own tuple; a supplied tuple has no executed attempt to bind")
	case input.Procedure != "":
		// Attribution slice: the tuple IS the procedure's output. Abstention
		// and malformed params are input errors — nothing persists.
		executed, err := witness.ExecuteAttempt(input.Procedure, input.Params)
		if err != nil {
			return WitnessCheckResponse{}, err
		}
		claim = executed
		binding = &store.WitnessAttemptBindingRow{
			ProposalID:      input.ProposalID,
			Procedure:       input.Procedure,
			ExecutorVersion: witness.AttemptExecutorVersion,
			ParamsCanonical: witness.CanonicalAttemptParams(input.Params),
			TupleCanonical:  executed.Canonical(),
		}
	default:
		if input.Note == "" {
			return WitnessCheckResponse{}, fmt.Errorf("a note recording the tuple's provenance is required: a witness claim without provenance is just a number")
		}
		parsed, err := witness.ParseErdosStraus(input.Tuple)
		if err != nil {
			return WitnessCheckResponse{}, err
		}
		claim = parsed
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

	// F2 (2026-09-12 review-flow run): SELECT AND PIN THE OCCURRENCE BEFORE ANY
	// WRITE. A witness verdict is an assessment of one occurrence of a proposal
	// — a specific generation's emitted interpretation — not of "whatever bytes
	// are newest". Resolving it here means an unattributable request fails
	// before a run row exists, and the persisted evaluation run carries the
	// generation and cluster context the contextual current-result reader
	// requires.
	occ, err := resolveWitnessOccurrence(ctx, repoStore, problemID, input)
	if err != nil {
		return WitnessCheckResponse{}, err
	}

	now := a.now()
	run, err := repoStore.CreateRun(ctx, domain.NewRun{
		ID:          domain.NewRunID(now),
		ProblemID:   problemID,
		Operation:   "witness check",
		Status:      domain.RunStatusRunning,
		InputRef:    "frontier_generation_run:" + occ.GenerationRunID + " frontier_proposal:" + input.ProposalID,
		ToolName:    "newf",
		ToolVersion: a.version,
		StartedAt:   now,
		CompletedAt: now,
	})
	if err != nil {
		return WitnessCheckResponse{}, err
	}

	notes := "witness " + witnessVerdict + " " + claim.Canonical()
	if detail != "" {
		notes += ": " + detail
	}
	if binding != nil {
		notes += " — attempt-bound: " + binding.Procedure + "@" + binding.ExecutorVersion + "(" + binding.ParamsCanonical + ") produced the tuple"
		if input.Note != "" {
			notes += " — note: " + input.Note
		}
	} else {
		notes += " — provenance: " + input.Note + " (supplied tuple; no executed-attempt binding)"
	}
	notes += " — occurrence: " + occ.GenerationRunID
	if !occ.OccurrencePinned {
		// Pre-v24 history: the generation recorded no occurrence binding, so
		// the assessed bytes cannot be attributed to it. Recorded, never faked.
		notes += " (no occurrence content binding; attribution gap)"
	}

	record := store.EvaluationRunRecord{
		ID:                      domain.NewEvaluationRunID(now),
		ProblemID:               problemID,
		RunID:                   run.ID,
		FrontierGenerationRunID: occ.GenerationRunID,
		ClusterRunID:            occ.ClusterRunID,
		Mode:                    "proposal",
		CreatedAt:               now.Format(timeLayout),
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
			SignatureContentHash: occ.ContentHash,
		}},
	}
	if binding != nil {
		binding.EvaluationID = record.Evaluations[0].ID
		record.Evaluations[0].AttemptBinding = binding
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
	resp := WitnessCheckResponse{
		OK:                      true,
		Command:                 "witness check",
		Store:                   dbPath,
		WitnessVerdict:          witnessVerdict,
		CanonicalClaim:          claim.Canonical(),
		Detail:                  detail,
		FrontierGenerationRunID: occ.GenerationRunID,
		ClusterRunID:            occ.ClusterRunID,
		OccurrenceContentHash:   occ.ContentHash,
		OccurrencePinned:        occ.OccurrencePinned,
		EvaluationRunID:         persisted.ID,
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
	}
	if binding != nil {
		resp.AttemptBinding = &WitnessAttemptBindingView{
			Procedure:       binding.Procedure,
			ExecutorVersion: binding.ExecutorVersion,
			ParamsCanonical: binding.ParamsCanonical,
			TupleCanonical:  binding.TupleCanonical,
		}
	}
	return resp, nil
}

// witnessOccurrence is the pinned assessment context of one witness check: the
// generation whose emitted interpretation the verdict is about, that
// generation's population (cluster run), and the exact content revision it
// bound.
type witnessOccurrence struct {
	GenerationRunID string
	ClusterRunID    string
	ContentHash     string
	// OccurrencePinned is false only for pre-v24 generations that recorded no
	// occurrence content binding: the content hash then comes from the
	// proposal's latest revision and is NOT an occurrence attribution.
	OccurrencePinned bool
}

// resolveWitnessOccurrence selects the occurrence a witness verdict is about,
// following the same context rules the evaluation stage uses (an explicit
// generation pins historical replay; otherwise the proposal's latest occurrence
// generation). Membership — what a generation EMITTED — is authoritative, not
// artifact ownership: a fully-deduped later generation owns no row yet binds
// the revised interpretation the verdict should attach to.
//
// Refusing an unattributable request here is deliberate. Attaching a checked
// tuple to a proposal without its occurrence produced an evaluation that no
// contextual current-result reader could ever surface (F2), which silently
// downgraded a reproducible domain observation to an orphan row.
func resolveWitnessOccurrence(ctx context.Context, repoStore problemStore, problemID string, input WitnessCheckInput) (witnessOccurrence, error) {
	genID := input.GenerationID
	if genID == "" {
		resolved, found, err := repoStore.LatestProposalOccurrenceGeneration(ctx, problemID, input.ProposalID)
		if err != nil {
			return witnessOccurrence{}, err
		}
		if !found {
			return witnessOccurrence{}, fmt.Errorf("proposal %s has no frontier occurrence: a witness verdict must be attributed to the generation whose interpretation it assesses", input.ProposalID)
		}
		genID = resolved
	}
	gen, err := repoStore.GetFrontierGeneration(ctx, genID)
	if err != nil {
		return witnessOccurrence{}, err
	}
	if gen.ProblemID != problemID {
		return witnessOccurrence{}, fmt.Errorf("generation %s belongs to problem %s, not proposal %s's problem %s", gen.ID, gen.ProblemID, input.ProposalID, problemID)
	}

	membership, err := repoStore.ListOccurrenceProposalRows(ctx, gen.ID)
	if err != nil {
		return witnessOccurrence{}, err
	}
	if len(membership) == 0 {
		membership = gen.Proposals
	}
	member := false
	for _, p := range membership {
		if p.ID == input.ProposalID {
			member = true
			break
		}
	}
	if !member {
		return witnessOccurrence{}, fmt.Errorf("proposal %s is not part of generation %s's occurrence membership", input.ProposalID, gen.ID)
	}

	contents, err := repoStore.ListGenerationOccurrenceContents(ctx, gen.ID)
	if err != nil {
		return witnessOccurrence{}, err
	}
	out := witnessOccurrence{GenerationRunID: gen.ID, ClusterRunID: gen.ClusterRunID}
	if oc, ok := contents[input.ProposalID]; ok {
		out.ContentHash = oc.ContentHash
		out.OccurrencePinned = !oc.FallbackLatest
	}
	return out, nil
}
