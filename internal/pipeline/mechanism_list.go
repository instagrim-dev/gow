package pipeline

import (
	"context"

	"github.com/instagrim-dev/newf/internal/store"
)

// --- mechanism list (E2: the provenance chain is walkable without raw SQL) ---

// MechanismListInput selects a problem's mechanisms.
type MechanismListInput struct {
	DBPath     string
	ProblemID  string
	JSONOutput bool
}

// MechanismListItemView is one mechanism summary row.
type MechanismListItemView struct {
	MechanismID        string `json:"mechanism_id"`
	ApproachID         string `json:"approach_id"`
	ApproachRevisionID string `json:"approach_revision_id"`
	Label              string `json:"label"`
	LogicalIdentity    string `json:"logical_identity"`
	OutcomeClass       string `json:"outcome_class"`
	SignatureCount     int    `json:"signature_count"`
}

// MechanismListResponse is returned by `newf mechanism list`.
type MechanismListResponse struct {
	OK         bool                    `json:"ok"`
	Command    string                  `json:"command"`
	Store      string                  `json:"store"`
	ProblemID  string                  `json:"problem_id"`
	Mechanisms []MechanismListItemView `json:"mechanisms"`
}

// ListMechanisms lists every persisted mechanism for a problem with its
// owning approach identity and signature count.
func (a *App) ListMechanisms(ctx context.Context, input MechanismListInput) (MechanismListResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return MechanismListResponse{}, err
	}
	defer repoStore.Close()

	items, err := repoStore.ListMechanismsForProblem(ctx, input.ProblemID)
	if err != nil {
		return MechanismListResponse{}, err
	}
	resp := MechanismListResponse{OK: true, Command: "mechanism list", Store: dbPath, ProblemID: input.ProblemID, Mechanisms: []MechanismListItemView{}}
	for _, item := range items {
		resp.Mechanisms = append(resp.Mechanisms, MechanismListItemView{
			MechanismID:        item.MechanismID,
			ApproachID:         item.ApproachID,
			ApproachRevisionID: item.ApproachRevisionID,
			Label:              item.Label,
			LogicalIdentity:    item.LogicalIdentity,
			OutcomeClass:       item.OutcomeClass,
			SignatureCount:     item.SignatureCount,
		})
	}
	return resp, nil
}

// --- batch signature (E2: no raw-SQL loop to canonicalize a problem) ---

// SignatureBatchInput selects every mechanism of a problem for signing.
type SignatureBatchInput struct {
	DBPath        string
	ProblemID     string
	VocabVersion  string
	SchemaVersion string
	JSONOutput    bool
}

// SignatureBatchItemView is one per-mechanism batch outcome.
type SignatureBatchItemView struct {
	MechanismID string `json:"mechanism_id"`
	SignatureID string `json:"signature_id"`
	Fingerprint string `json:"fingerprint"`
	Status      string `json:"status"`
}

// SignatureBatchResponse is returned by `newf mechanism signature --problem`.
type SignatureBatchResponse struct {
	OK         bool                     `json:"ok"`
	Command    string                   `json:"command"`
	Store      string                   `json:"store"`
	ProblemID  string                   `json:"problem_id"`
	Signatures []SignatureBatchItemView `json:"signatures"`
}

// SignatureAllMechanisms computes the canonical signature for every mechanism
// of a problem through the SAME per-mechanism path (identical idempotency:
// an already-signed mechanism reports created=false, never a duplicate).
func (a *App) SignatureAllMechanisms(ctx context.Context, input SignatureBatchInput) (SignatureBatchResponse, error) {
	dbPath, repoStore, err := a.openStoreFn(ctx, input.DBPath)
	if err != nil {
		return SignatureBatchResponse{}, err
	}
	items, err := repoStore.ListMechanismsForProblem(ctx, input.ProblemID)
	repoStore.Close()
	if err != nil {
		return SignatureBatchResponse{}, err
	}

	resp := SignatureBatchResponse{OK: true, Command: "mechanism signature", Store: dbPath, ProblemID: input.ProblemID, Signatures: []SignatureBatchItemView{}}
	for _, item := range items {
		one, serr := a.SignatureMechanism(ctx, SignatureInput{
			DBPath:        input.DBPath,
			MechanismID:   item.MechanismID,
			VocabVersion:  input.VocabVersion,
			SchemaVersion: input.SchemaVersion,
		})
		if serr != nil {
			return SignatureBatchResponse{}, serr
		}
		resp.Signatures = append(resp.Signatures, SignatureBatchItemView{
			MechanismID: item.MechanismID,
			SignatureID: one.Signature.ID,
			Fingerprint: one.Signature.Fingerprint,
			Status:      one.Status,
		})
	}
	return resp, nil
}

// mechanismStoreReader is satisfied by *store.Store; declared for doc clarity.
var _ = store.MechanismListItem{}
