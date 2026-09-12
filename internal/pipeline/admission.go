// Package pipeline: evidence admission (v35, structural review finding S2).
//
// `evaluated_failures` is a re-entry MARKER, not an admitted observation: no
// code path carried an evaluated failure back into the atlas population that
// BuildClustering and invariant mining consume. AdmitEvidence closes that gap
// with an explicit, typed decision per evaluated failure — never by
// indiscriminately inserting every model-evaluated failure into the historical
// atlas. A description that fails its own structural claim, a domain-checked
// failed attempt, and a model-judged failure are DIFFERENT observations with
// different admission rules (reviewer S2).
package pipeline

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/instagrim-dev/newf/internal/canon"
	"github.com/instagrim-dev/newf/internal/domain"
	"github.com/instagrim-dev/newf/internal/store"
	"github.com/instagrim-dev/newf/internal/verify"
	"github.com/instagrim-dev/newf/internal/witness"
)

// Observation kinds recorded on every admission decision.
const (
	// ObservationStructuralClaimFailure: a deterministic-check failure — the
	// proposed DESCRIPTION failed its own claimed break ("still satisfies a
	// targeted invariant"). It narrows the description space, not the
	// observed-mechanism space, and is never atlas-admissible.
	ObservationStructuralClaimFailure = "structural-claim-failure"
	// ObservationDomainCheckedFailure: a deterministic / reproducible /
	// independent-evidence strength failure of an actual attempt.
	ObservationDomainCheckedFailure = "domain-checked-failure"
	// ObservationModelJudgedFailure: a model-judgment or independent-critic
	// strength failure. Admissible only by explicit operator attestation and
	// permanently labeled model-judged (ModelJudgment != Verification).
	ObservationModelJudgedFailure = "model-judged-failure"
)

// admissionClassification is the typed admission rule derived from what the
// verifier hierarchy actually recorded about one evaluated failure.
type admissionClassification struct {
	// ObservationKind is one of the Observation* constants.
	ObservationKind string
	// RuleAdmissible: the rule may admit without operator attestation.
	RuleAdmissible bool
	// Attestable: an operator may admit with --attest. Structural-claim
	// failures are NOT attestable: the described mechanism never existed as
	// attempted work, so admitting it would populate the mechanism atlas with
	// descriptions instead of observations.
	Attestable bool
	// Basis is the rule's stated reason (used verbatim for rule decisions).
	Basis string
}

// classifyEvaluatedFailure maps an evaluation row (verifier kind + strength)
// onto the observation taxonomy and its admission rule. It is pure and total
// over valid verifier vocabulary; unknown vocabulary is an error, never a
// silent default (no epistemic promotion by fallthrough).
func classifyEvaluatedFailure(ev store.EvaluationRow) (admissionClassification, error) {
	if verify.VerifierKind(ev.VerifierKind) == verify.KindDeterministicCheck {
		return admissionClassification{
			ObservationKind: ObservationStructuralClaimFailure,
			RuleAdmissible:  false,
			Attestable:      false,
			Basis:           "deterministic-check failure: the description failed its own claimed break; it narrows the description space, not the observed-mechanism space, and is not an atlas observation",
		}, nil
	}
	if verify.VerifierKind(ev.VerifierKind) == verify.KindCounterexampleSearch {
		// A counterexample search over persisted artifacts sits in the
		// reproducible band, but its decisive negative is a predicate bit
		// ("a matching break was previously observed"), not a mechanism-level
		// refutation carrying a checkable witness. Until an admission can
		// reference the concrete witness it rests on, the strength band alone
		// must not rule-admit it (2026-09-12 review F5: rule admission is
		// witness-based, not band-based).
		return admissionClassification{
			ObservationKind: ObservationDomainCheckedFailure,
			RuleAdmissible:  false,
			Attestable:      true,
			Basis:           "counterexample-search failure: reproducible-band predicate verdict without a persisted witness reference; admission requires operator attestation naming the witness",
		}, nil
	}
	switch verify.VerificationStrength(ev.VerificationStrength) {
	case verify.StrengthDeterministic, verify.StrengthReproducible:
		basis := "rule: " + ev.VerificationStrength + "-strength failure of the assessed content admits as a domain-checked failed attempt"
		// D2-C (issue #23): a tool-attributed verdict names its instrument and
		// recorded claim in the admission basis, so a witness-backed rule
		// admission is auditable to the canonical witness it rests on.
		if ev.ToolName != "" {
			basis += " (" + ev.ToolName
			if ev.ToolVersion != "" {
				basis += "/" + ev.ToolVersion
			}
			if ev.Notes != "" {
				basis += ": " + ev.Notes
			}
			basis += ")"
		}
		return admissionClassification{
			ObservationKind: ObservationDomainCheckedFailure,
			RuleAdmissible:  true,
			Attestable:      true,
			Basis:           basis,
		}, nil
	case verify.StrengthIndependentEvidence:
		return admissionClassification{
			ObservationKind: ObservationDomainCheckedFailure,
			RuleAdmissible:  false,
			Attestable:      true,
			Basis:           "independent-evidence strength failure: the external-evidence linkage is not machine-checked here; admission requires operator attestation",
		}, nil
	case verify.StrengthIndependentCritic, verify.StrengthSingleModelJudgment:
		return admissionClassification{
			ObservationKind: ObservationModelJudgedFailure,
			RuleAdmissible:  false,
			Attestable:      true,
			Basis:           "model-judged failure (" + ev.VerificationStrength + "): a model verdict is a judgment, not a verified domain observation; admission requires operator attestation and stays labeled model-judged",
		}, nil
	default:
		return admissionClassification{}, fmt.Errorf("evaluation %s has unrecognized verification strength %q", ev.ID, ev.VerificationStrength)
	}
}

// AdmitEvidenceInput drives `newf evidence admit`.
type AdmitEvidenceInput struct {
	DBPath    string
	ProblemID string
	// EvaluationID targets one recorded evaluated failure; empty means the
	// batch rule pass over every undecided evaluated failure of the problem.
	EvaluationID string
	// Attest admits an attestable-but-not-rule-admissible failure by explicit
	// operator decision. Requires EvaluationID and a non-empty Note.
	Attest bool
	// Note is the operator's attestation basis (recorded verbatim).
	Note       string
	JSONOutput bool
}

// EvidenceAdmissionView is one persisted admission decision.
type EvidenceAdmissionView struct {
	ID                 string `json:"id"`
	ProposalID         string `json:"proposal_id"`
	EvaluationID       string `json:"evaluation_id"`
	Decision           string `json:"decision"`
	ObservationKind    string `json:"observation_kind"`
	AdmittedBy         string `json:"admitted_by"`
	Basis              string `json:"basis"`
	ContentHash        string `json:"content_hash,omitempty"`
	ApproachID         string `json:"approach_id,omitempty"`
	ApproachRevisionID string `json:"approach_revision_id,omitempty"`
	MechanismID        string `json:"mechanism_id,omitempty"`
	SignatureID        string `json:"signature_id,omitempty"`
	CreatedAt          string `json:"created_at"`
}

// AdmitEvidenceResponse reports one admit pass.
type AdmitEvidenceResponse struct {
	OK        bool                    `json:"ok"`
	Command   string                  `json:"command"`
	Store     string                  `json:"store"`
	ProblemID string                  `json:"problem_id"`
	RunID     string                  `json:"run_id"`
	Admitted  []EvidenceAdmissionView `json:"admitted"`
	Withheld  []EvidenceAdmissionView `json:"withheld"`
	// Skipped lists evaluated failures that already carry a decision this
	// pass would repeat (idempotency: re-running admit changes nothing).
	Skipped []EvidenceAdmissionView `json:"skipped"`
}

// EvidenceAdmissionListResponse reports the full admission ledger of a problem.
type EvidenceAdmissionListResponse struct {
	OK         bool                    `json:"ok"`
	Command    string                  `json:"command"`
	Store      string                  `json:"store"`
	ProblemID  string                  `json:"problem_id"`
	Admissions []EvidenceAdmissionView `json:"admissions"`
}

func admissionView(r store.EvidenceAdmissionRow) EvidenceAdmissionView {
	return EvidenceAdmissionView{
		ID:                 r.ID,
		ProposalID:         r.ProposalID,
		EvaluationID:       r.EvaluationID,
		Decision:           r.Decision,
		ObservationKind:    r.ObservationKind,
		AdmittedBy:         r.AdmittedBy,
		Basis:              r.Basis,
		ContentHash:        r.ContentHash,
		ApproachID:         r.ApproachID,
		ApproachRevisionID: r.ApproachRevisionID,
		MechanismID:        r.MechanismID,
		SignatureID:        r.SignatureID,
		CreatedAt:          r.CreatedAt,
	}
}

// ListEvidenceAdmissions returns the persisted admission ledger for a problem.
func (a *App) ListEvidenceAdmissions(ctx context.Context, input AdmitEvidenceInput) (EvidenceAdmissionListResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return EvidenceAdmissionListResponse{}, err
	}
	defer repoStore.Close()
	if _, err := repoStore.GetProblem(ctx, input.ProblemID); err != nil {
		return EvidenceAdmissionListResponse{}, err
	}
	rows, err := repoStore.ListEvidenceAdmissions(ctx, input.ProblemID)
	if err != nil {
		return EvidenceAdmissionListResponse{}, err
	}
	resp := EvidenceAdmissionListResponse{OK: true, Command: "evidence list", Store: dbPath, ProblemID: input.ProblemID}
	for _, r := range rows {
		resp.Admissions = append(resp.Admissions, admissionView(r))
	}
	return resp, nil
}

// AdmitEvidence runs one admission pass over the problem's evaluated failures
// (or one targeted evaluation). Rule-admissible failures are admitted by rule;
// everything else is withheld with the refusing rule persisted. With
// Attest+Note, one attestable withheld/undecided failure is admitted by
// operator decision — the earlier withheld row persists, showing supersession.
// Admission materializes the EXACT assessed signature content into the atlas,
// so the next BuildClustering consumes it as a population member.
func (a *App) AdmitEvidence(ctx context.Context, input AdmitEvidenceInput) (AdmitEvidenceResponse, error) {
	if input.Attest {
		if input.EvaluationID == "" {
			return AdmitEvidenceResponse{}, fmt.Errorf("--attest requires --evaluation: operator attestation is a decision about one identified evaluated failure")
		}
		if strings.TrimSpace(input.Note) == "" {
			return AdmitEvidenceResponse{}, fmt.Errorf("--attest requires a non-empty --note recording the operator's basis")
		}
	}

	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return AdmitEvidenceResponse{}, err
	}
	defer repoStore.Close()

	if _, err := repoStore.GetProblem(ctx, input.ProblemID); err != nil {
		return AdmitEvidenceResponse{}, err
	}

	failures, err := repoStore.ListEvaluatedFailures(ctx, input.ProblemID)
	if err != nil {
		return AdmitEvidenceResponse{}, err
	}
	if input.EvaluationID != "" {
		selected := failures[:0:0]
		for _, f := range failures {
			if f.EvaluationID == input.EvaluationID {
				selected = append(selected, f)
			}
		}
		if len(selected) == 0 {
			return AdmitEvidenceResponse{}, fmt.Errorf("evaluation %s is not a recorded evaluated failure for problem %s", input.EvaluationID, input.ProblemID)
		}
		failures = selected
	}

	prior, err := repoStore.ListEvidenceAdmissions(ctx, input.ProblemID)
	if err != nil {
		return AdmitEvidenceResponse{}, err
	}
	decided := make(map[string]map[string]store.EvidenceAdmissionRow) // evaluation id -> decision -> row
	for _, r := range prior {
		if decided[r.EvaluationID] == nil {
			decided[r.EvaluationID] = make(map[string]store.EvidenceAdmissionRow)
		}
		decided[r.EvaluationID][r.Decision] = r
	}

	now := a.now()
	run, err := repoStore.CreateRun(ctx, domain.NewRun{
		ID:          domain.NewRunID(now),
		ProblemID:   input.ProblemID,
		Operation:   "evidence admit",
		Status:      domain.RunStatusRunning,
		InputRef:    "problem:" + input.ProblemID,
		ToolName:    "newf",
		ToolVersion: a.version,
		StartedAt:   now,
		CompletedAt: now,
	})
	if err != nil {
		return AdmitEvidenceResponse{}, err
	}

	resp := AdmitEvidenceResponse{OK: true, Command: "evidence admit", Store: dbPath, ProblemID: input.ProblemID, RunID: run.ID}
	for _, f := range failures {
		// Idempotency: an admitted evaluation is settled forever; a withheld
		// one is settled for RULE passes and reopens only under attestation.
		if row, ok := decided[f.EvaluationID]["admitted"]; ok {
			resp.Skipped = append(resp.Skipped, admissionView(row))
			continue
		}
		withheldRow, alreadyWithheld := decided[f.EvaluationID]["withheld"]
		if alreadyWithheld && !input.Attest {
			resp.Skipped = append(resp.Skipped, admissionView(withheldRow))
			continue
		}

		ev, gerr := repoStore.GetEvaluation(ctx, f.EvaluationID)
		if gerr != nil {
			a.failRun(ctx, repoStore, run.ID, gerr)
			return AdmitEvidenceResponse{}, gerr
		}
		class, cerr := classifyEvaluatedFailure(ev)
		if cerr != nil {
			a.failRun(ctx, repoStore, run.ID, cerr)
			return AdmitEvidenceResponse{}, cerr
		}

		// The observation IS the assessed content: without the exact bytes the
		// verdict was computed over there is nothing admissible to materialize.
		content, found, gerr := repoStore.GetProposalSignatureContentByHash(ctx, f.ProposalID, ev.SignatureContentHash)
		if gerr != nil {
			a.failRun(ctx, repoStore, run.ID, gerr)
			return AdmitEvidenceResponse{}, gerr
		}

		admit := class.RuleAdmissible
		admittedBy := "rule"
		basis := class.Basis
		// Attribution slice (v46, 2026-09-12 review handoff 3): a recorded
		// attempt→output binding is RECHECKED here by recomputation — the
		// admission does not trust the write-time claim that the procedure
		// produced the tuple; it re-runs the procedure. A verified binding is
		// named in the rule basis (checkable attribution); a binding that
		// fails or refuses recomputation degrades rule admission to
		// withholding — the attribution the rule relied on cannot be
		// machine-checked, and operator attestation is the recorded escape.
		// Absence of a binding changes nothing: the supplied-tuple path keeps
		// its explicitly weaker provenance (recorded in the evaluation notes).
		attemptBinding, attemptBound, berr := repoStore.GetWitnessAttemptBinding(ctx, f.EvaluationID)
		if berr != nil {
			a.failRun(ctx, repoStore, run.ID, berr)
			return AdmitEvidenceResponse{}, berr
		}
		if attemptBound && admit {
			if verr := witness.VerifyAttemptBinding(attemptBinding.Procedure, attemptBinding.ExecutorVersion, attemptBinding.ParamsCanonical, attemptBinding.TupleCanonical); verr != nil {
				admit = false
				basis = "attempt→output binding failed recomputation recheck (" + verr.Error() + "); the rule cannot admit an attribution it cannot re-verify — operator attestation may"
			} else {
				basis += " — attempt→output binding verified by recomputation: " + attemptBinding.Procedure + "@" + attemptBinding.ExecutorVersion + "(" + attemptBinding.ParamsCanonical + ")"
			}
		}
		if input.Attest {
			if !class.Attestable {
				err := fmt.Errorf("evaluation %s is a %s and cannot be admitted even by attestation: %s", ev.ID, class.ObservationKind, class.Basis)
				a.failRun(ctx, repoStore, run.ID, err)
				return AdmitEvidenceResponse{}, err
			}
			admit = true
			admittedBy = "operator"
			basis = "operator attestation: " + strings.TrimSpace(input.Note)
		}
		if admit && !found {
			admit = false
			admittedBy = "rule"
			basis = "no persisted assessed signature content matches this evaluation's content hash; the verdict has no materializable observation"
		}

		row := store.EvidenceAdmissionRow{
			ID:              domain.NewEvidenceAdmissionID(now),
			ProblemID:       input.ProblemID,
			RunID:           run.ID,
			ProposalID:      f.ProposalID,
			EvaluationID:    f.EvaluationID,
			ObservationKind: class.ObservationKind,
			AdmittedBy:      admittedBy,
			Basis:           basis,
			ContentHash:     ev.SignatureContentHash,
			CreatedAt:       now.Format(timeLayout),
		}
		if !admit {
			if alreadyWithheld {
				// Attest pass refused above already; a rule pass on an already-
				// withheld row was skipped above. Reaching here means an attest
				// pass degraded to withholding (missing content) that duplicates
				// the earlier decision: skip instead of violating uniqueness.
				resp.Skipped = append(resp.Skipped, admissionView(withheldRow))
				continue
			}
			row.Decision = "withheld"
			persisted, perr := repoStore.PersistEvidenceAdmission(ctx, row)
			if perr != nil {
				a.failRun(ctx, repoStore, run.ID, perr)
				return AdmitEvidenceResponse{}, perr
			}
			resp.Withheld = append(resp.Withheld, admissionView(persisted))
			continue
		}

		row.Decision = "admitted"
		matInput, merr := admittedFailureInput(input.ProblemID, run.ID, f, ev, content, row, now)
		if merr != nil {
			a.failRun(ctx, repoStore, run.ID, merr)
			return AdmitEvidenceResponse{}, merr
		}
		// One transaction: snapshot + normalization + signature + admission
		// decision commit together or not at all — a failure mid-materialization
		// must not leave population rows without their justifying decision.
		mat, merr := repoStore.PersistAdmittedFailure(ctx, matInput)
		if merr != nil {
			a.failRun(ctx, repoStore, run.ID, merr)
			return AdmitEvidenceResponse{}, merr
		}
		resp.Admitted = append(resp.Admitted, admissionView(mat.Admission))
	}

	if err := a.finalizeRun(ctx, repoStore, run.ID, nil); err != nil {
		return AdmitEvidenceResponse{}, err
	}
	return resp, nil
}

// admittedFailureInput builds the single-transaction materialization input
// carrying one admitted evaluated failure into the atlas population:
//
//  1. a source snapshot whose bytes ARE the assessed signature JSON (real
//     content-addressed provenance, not a fabricated source);
//  2. a normalization revision with one approach under the stable logical
//     identity "frontier-proposal:<id>" — re-admitting a re-assessed proposal
//     therefore produces a NEW REVISION of the SAME approach, and the
//     current-heads population (S4 fix) keeps exactly one interpretation
//     current instead of double-counting;
//  3. the assessed canon.MechanismSignature persisted VERBATIM against the new
//     mechanism (claims, resolution states, and claim statuses unchanged) —
//     only the outcome is set from the evaluation verdict, with `inferred`
//     provenance because it is tool-derived, not source-explicit;
//  4. the admission decision row itself.
//
// All four commit atomically in store.PersistAdmittedFailure (the store
// resolves the snapshot id, mechanism id, and materialization ids mid-write).
// The next BuildClustering then consumes the signature through the ordinary
// current-heads population read; no clustering special case exists for
// admitted evidence.
func admittedFailureInput(problemID, runID string, f store.EvaluatedFailureRow, ev store.EvaluationRow, content store.ProposalSignatureContentRow, row store.EvidenceAdmissionRow, nowTime time.Time) (store.AdmittedFailureInput, error) {
	var sig canon.MechanismSignature
	if err := json.Unmarshal([]byte(content.SignatureJSON), &sig); err != nil {
		return store.AdmittedFailureInput{}, fmt.Errorf("proposal %s: corrupt persisted signature content: %w", f.ProposalID, err)
	}

	// Snapshot: the assessed bytes, content-addressed.
	raw := []byte(content.SignatureJSON)
	sum := sha256.Sum256(raw)
	digest := hex.EncodeToString(sum[:])
	snapshot := store.SnapshotAdmission{
		ProblemID:   problemID,
		Kind:        domain.SourceKindLocalPath,
		LogicalName: "evidence-admission-" + ev.ID + ".json",
		Origin:      "admission://evaluation/" + ev.ID,
		SHA256:      digest,
		ByteLength:  int64(len(raw)),
		MediaType:   "application/json",
		ObjectPath:  filepath.Join("sha256", digest[:2], digest),
		IngestRunID: runID,
		ObservedAt:  nowTime,
	}

	// Outcome from the evaluation verdict; the assessed signature proposed a
	// mechanism whose outcome was unknown at generation time.
	outcomeClass := domain.OutcomeClass(f.Verdict)

	// The snapshot id is resolved inside the transaction (dedup may return an
	// existing snapshot); the store injects it into the revision before write.
	normInput, err := admissionNormalizationInput(problemID, runID, "", f, ev, sig, outcomeClass, nowTime)
	if err != nil {
		return store.AdmittedFailureInput{}, err
	}

	// The assessed signature verbatim; the mechanism id is resolved inside the
	// transaction from the normalization's approach ref. The outcome fields are
	// the only mutation, and their provenance is honest: tool-derived
	// (`inferred`), never source-explicit.
	sig.MechanismID = ""
	sig.SignatureID = ""
	sig.OutcomeClass = outcomeClass
	sig.OutcomeProvenance = domain.ClaimInferred
	rec := signatureRecord(sig, "", runID, nowTime)

	return store.AdmittedFailureInput{
		Snapshot:      snapshot,
		Normalization: normInput,
		Signature:     rec,
		Admission:     row,
	}, nil
}

// admissionNormalizationInput builds the store.NormalizationInput for one
// admitted failure. The approach's logical identity is stable per proposal
// ("frontier-proposal:<id>"); the mechanism/attribute rows are derived from
// the assessed signature's claims (surface label, falling back to the
// canonical id when the generated claim carries no surface form). The
// invocation records the deterministic in-repo transformation — no provider
// concept enters the domain records.
func admissionNormalizationInput(problemID, runID, snapshotID string, f store.EvaluatedFailureRow, ev store.EvaluationRow, sig canon.MechanismSignature, outcomeClass domain.OutcomeClass, now time.Time) (store.NormalizationInput, error) {
	invocationID := domain.NewProviderInvocationID(now)
	revisionID := domain.NewNormalizationRevisionID(now)
	approachRevisionID := domain.NewApproachRevisionID(now)
	mechanismID := domain.NewMechanismID(now)

	attributes := make([]domain.MechanismAttribute, 0)
	appendClaims := func(kind domain.MechanismAttributeKind, claims []canon.FieldClaim) {
		for _, c := range claims {
			value := strings.TrimSpace(c.SurfaceLabel)
			if value == "" {
				value = string(c.CanonicalID)
			}
			if value == "" {
				continue
			}
			attributes = append(attributes, domain.MechanismAttribute{
				MechanismID: mechanismID,
				Kind:        kind,
				Value:       value,
			})
		}
	}
	appendClaims(domain.AttrRepresentation, sig.Representations)
	appendClaims(domain.AttrOperator, sig.Operators)
	appendClaims(domain.AttrAssumption, sig.Assumptions)
	appendClaims(domain.AttrPreserves, sig.Preserves)
	appendClaims(domain.AttrBreaks, sig.Breaks)
	appendClaims(domain.AttrAuxiliaryObject, sig.AuxiliaryObjects)

	boundaries := make([]domain.FailureBoundary, 0, len(sig.Boundaries))
	for _, b := range sig.Boundaries {
		condition := strings.TrimSpace(b.SurfaceLabel)
		if condition == "" {
			condition = string(b.CanonicalID)
		}
		if condition == "" {
			continue
		}
		boundaries = append(boundaries, domain.FailureBoundary{
			ID:                 domain.NewFailureBoundaryID(now),
			ApproachRevisionID: approachRevisionID,
			Condition:          condition,
		})
	}

	return store.NormalizationInput{
		Invocation: domain.ProviderInvocation{
			ID:            invocationID,
			RunID:         runID,
			Role:          domain.RoleNormalize,
			ProviderName:  "evidence-admission",
			SchemaVersion: "normalize/v1",
			RequestHash:   "admission:" + ev.SignatureContentHash,
			CreatedAt:     now,
		},
		Revision: domain.NormalizationRevision{
			ID:                   revisionID,
			ProblemID:            problemID,
			RunID:                runID,
			SnapshotID:           snapshotID,
			ProviderInvocationID: invocationID,
			SchemaVersion:        "normalize/v1",
			ConfigHash:           "evidence-admission/v1",
			Status:               domain.NormalizationStatusSucceeded,
			CreatedAt:            now,
		},
		Approaches: []store.ApproachInput{{
			LogicalIdentity: "frontier-proposal:" + f.ProposalID,
			Revision: domain.ApproachRevision{
				ID:                      approachRevisionID,
				ApproachID:              domain.NewApproachID(now),
				NormalizationRevisionID: revisionID,
				Label:                   "admitted evaluated failure of proposal " + f.ProposalID,
				Description:             "materialized from evaluation " + ev.ID + " (" + ev.VerifierKind + ", " + ev.VerificationStrength + ") over assessed content " + ev.SignatureContentHash,
				CreatedAt:               now,
			},
			Mechanism: domain.Mechanism{
				ID:                 mechanismID,
				ApproachRevisionID: approachRevisionID,
				Locality:           sig.Posture.Locality,
				ConstructionMode:   sig.Posture.Construction,
				UncertaintyMode:    sig.Posture.Uncertainty,
				Notes:              "admitted evidence from evaluation " + ev.ID,
			},
			Attributes: attributes,
			Outcome: domain.Outcome{
				ID:                 domain.NewOutcomeID(now),
				ApproachRevisionID: approachRevisionID,
				Class:              outcomeClass,
				Notes:              "evaluated verdict (" + ev.VerificationStrength + " / " + ev.VerifierKind + ")",
			},
			Boundaries: boundaries,
		}},
	}, nil
}
